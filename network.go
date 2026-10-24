package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

const backupNetworkURL = "https://ipwho.is/"
const geoNetworkURL = "https://get.geojs.io/v1/ip/geo.json"
const infoNetworkURL = "https://ipinfo.io/json"

var networkSources = []string{geoNetworkURL, backupNetworkURL, infoNetworkURL, networkURL}

// Owned by Service.network. Results and provider cooldowns never leave memory.
type networkState struct {
	blocked  map[string]time.Time
	last     NetworkReport
	evidence NetworkReport
	next     time.Time
}

func (s *networkState) detect(ctx context.Context, client *http.Client, now time.Time) NetworkReport {
	if now.Before(s.next) {
		r := s.last
		r.Cached = true
		return r
	}
	if s.blocked == nil {
		s.blocked = map[string]time.Time{}
	}

	ctx, cancelAll := context.WithCancel(ctx)
	defer cancelAll()
	var successes []NetworkReport
	var notes []string
	type answer struct {
		report NetworkReport
		retry  time.Time
	}
	answers := make(chan answer, len(networkSources))
	pending, index := 0, 0
	start := func() {
		for index < len(networkSources) && pending < 2 {
			source := networkSources[index]
			index++
			if until := s.blocked[source]; now.Before(until) {
				notes = append(notes, source+" 限流冷却中，至 "+until.Local().Format("15:04:05"))
				continue
			}
			pending++
			go func() {
				attempt, cancel := context.WithTimeout(ctx, 4*time.Second)
				defer cancel()
				r, retry := queryProvider(attempt, client, source, now)
				answers <- answer{r, retry}
			}()
			return
		}
	}
	start()
	stagger := time.NewTicker(800 * time.Millisecond)
	defer stagger.Stop()
probing:
	for pending > 0 {
		select {
		case a := <-answers:
			pending--
			if !a.retry.IsZero() {
				s.blocked[a.report.Source] = a.retry
			}
			if a.report.Status == "ok" {
				successes = append(successes, a.report)
			} else {
				notes = append(notes, a.report.Source+"："+a.report.Message)
			}
			if len(successes) >= 2 {
				break probing
			}
			start()
		case <-stagger.C:
			start()
		case <-ctx.Done():
			notes = append(notes, "整体检测超时")
			break probing
		}
	}
	r := NetworkReport{Time: now.UTC().Format(time.RFC3339), Status: "unknown", Source: strings.Join(networkSources, " / "), Message: "检测服务暂不可用，请稍后重试；仍可手动选择模板"}
	cooldown := 30 * time.Second
	if len(successes) > 0 {
		if len(successes) == 2 && successes[0].Timezone == "" && successes[1].Timezone != "" {
			successes[0], successes[1] = successes[1], successes[0]
		}
		r = successes[0]
		cooldown = 60 * time.Second
		if len(successes) == 1 {
			notes = append(notes, "仅一个来源可用，位置与推荐为该数据库的估算，尚未交叉确认")
		}
		if len(successes) == 2 {
			other := successes[1]
			notes = append(notes, "交叉查询："+locationEvidence(other))
			if locationsDisagree(r, other) {
				r.LocationConflict = true
				r.RecommendedPreset = ""
				r.RecommendationReason = "检测源的出口 IP 或地区判断存在分歧，暂不推荐模板；请按代理节点标注选择或自定义。"
			}
		}
		// A previously observed disagreement must not disappear merely because
		// one source is temporarily rate limited. Do not keep it across IP changes.
		if len(successes) == 1 && s.evidence.Status == "ok" && s.evidence.IP == r.IP && (s.evidence.LocationConflict || locationsDisagree(s.evidence, r)) {
			r.LocationConflict = true
			r.RecommendedPreset = ""
			r.RecommendationReason = "同一 IP 在本次会话中出现过地区判断分歧，目前只有一个来源可用，暂不推荐模板。"
			notes = append(notes, "此前结果："+locationEvidence(s.evidence))
		}
	}
	if r.Status == "ok" {
		s.evidence = r
	}
	r.Notes = strings.Join(notes, "；")
	s.next = now.Add(cooldown)
	r.RetryAt = s.next.UTC().Format(time.RFC3339)
	s.last = r
	return r
}

func locationEvidence(r NetworkReport) string {
	return r.Source + " → " + r.IP + " / " + r.Country + " / " + r.Region + " / " + r.City + " / " + r.Timezone
}

func locationsDisagree(a, b NetworkReport) bool {
	if a.IP != b.IP || a.CountryCode != b.CountryCode {
		return true
	}
	// Compare normalized preset classification to allow equivalent IANA aliases.
	if a.Timezone != "" && b.Timezone != "" {
		ac, bc := a, b
		recommendTemplate(&ac)
		recommendTemplate(&bc)
		return ac.RecommendedPreset != bc.RecommendedPreset
	}
	return false
}

func retryAfter(value string, now time.Time) time.Time {
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds > 0 {
		return now.Add(time.Duration(seconds) * time.Second)
	}
	if date, err := http.ParseTime(value); err == nil && date.After(now) {
		return date
	}
	return now.Add(5 * time.Minute)
}

func queryNetwork(ctx context.Context, client *http.Client) NetworkReport {
	r, _ := queryProvider(ctx, client, networkURL, time.Now())
	return r
}

func queryProvider(ctx context.Context, client *http.Client, source string, now time.Time) (NetworkReport, time.Time) {
	r := NetworkReport{Time: now.UTC().Format(time.RFC3339), Source: source, Status: "unknown", ProxyMode: "直连（Go 环境变量规则）"}
	none := time.Time{}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	proxy, err := http.ProxyFromEnvironment(req)
	if err != nil {
		r.Message = "代理配置无效"
		return r, none
	}
	if proxy != nil {
		r.ProxyMode = "环境变量代理（不展示地址及凭据）"
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		r.Message = "连接失败或请求超时"
		return r, none
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		r.Message = fmt.Sprintf("服务返回 HTTP %d", resp.StatusCode)
		if resp.StatusCode == http.StatusTooManyRequests {
			return r, retryAfter(resp.Header.Get("Retry-After"), now)
		}
		return r, none
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 65537))
	if err != nil || len(body) > 65536 {
		r.Message = "服务数据缺失或格式异常"
		return r, none
	}
	var parsed NetworkReport
	if source == networkURL {
		var data struct {
			IP       string `json:"ip"`
			Country  string `json:"country_name"`
			Code     string `json:"country_code"`
			ASN      string `json:"asn"`
			Org      string `json:"org"`
			Error    bool   `json:"error"`
			City     string `json:"city"`
			Region   string `json:"region"`
			Timezone string `json:"timezone"`
		}
		if json.Unmarshal(body, &data) != nil || data.Error {
			r.Message = "服务数据缺失或格式异常"
			return r, none
		}
		parsed = NetworkReport{IP: data.IP, Country: data.Country, CountryCode: data.Code, ASN: data.ASN, Organization: data.Org, City: data.City, Region: data.Region, Timezone: data.Timezone}

	} else if source == geoNetworkURL {
		var data struct {
			IP       string `json:"ip"`
			Country  string `json:"country"`
			Code     string `json:"country_code"`
			City     string `json:"city"`
			Region   string `json:"region"`
			Timezone string `json:"timezone"`
			ASN      int64  `json:"asn"`
			Org      string `json:"organization_name"`
		}
		if json.Unmarshal(body, &data) != nil {
			r.Message = "服务数据缺失或格式异常"
			return r, none
		}
		parsed = NetworkReport{IP: data.IP, Country: data.Country, CountryCode: data.Code, City: data.City, Region: data.Region, Timezone: data.Timezone, Organization: data.Org}
		if data.ASN > 0 {
			parsed.ASN = fmt.Sprintf("AS%d", data.ASN)
		}
	} else if source == infoNetworkURL {
		var data struct {
			IP       string `json:"ip"`
			Country  string `json:"country"`
			City     string `json:"city"`
			Region   string `json:"region"`
			Timezone string `json:"timezone"`
			Org      string `json:"org"`
			Bogon    bool   `json:"bogon"`
		}
		if json.Unmarshal(body, &data) != nil || data.Bogon {
			r.Message = "服务数据缺失或格式异常"
			return r, none
		}
		parsed = NetworkReport{IP: data.IP, Country: data.Country, CountryCode: data.Country, City: data.City, Region: data.Region, Timezone: data.Timezone, Organization: data.Org}
		if parts := strings.Fields(data.Org); len(parts) > 0 && strings.HasPrefix(parts[0], "AS") {
			parsed.ASN = parts[0]
		}
	} else if source == backupNetworkURL {
		var data struct {
			Success  bool   `json:"success"`
			IP       string `json:"ip"`
			Country  string `json:"country"`
			Code     string `json:"country_code"`
			City     string `json:"city"`
			Region   string `json:"region"`
			Timezone struct {
				ID string `json:"id"`
			} `json:"timezone"`
			Connection struct {
				ASN int64  `json:"asn"`
				Org string `json:"org"`
			} `json:"connection"`
		}
		if json.Unmarshal(body, &data) != nil || !data.Success {
			r.Message = "服务数据缺失或格式异常"
			return r, none
		}
		parsed = NetworkReport{IP: data.IP, Country: data.Country, CountryCode: data.Code, Organization: data.Connection.Org, City: data.City, Region: data.Region, Timezone: data.Timezone.ID}
		if data.Connection.ASN > 0 {
			parsed.ASN = fmt.Sprintf("AS%d", data.Connection.ASN)
		}
	}
	code := strings.ToUpper(parsed.CountryCode)
	ip := net.ParseIP(parsed.IP)
	addr, _ := netip.ParseAddr(parsed.IP)
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() || netip.MustParsePrefix("100.64.0.0/10").Contains(addr.Unmap()) || len(code) != 2 || code[0] < 'A' || code[0] > 'Z' || code[1] < 'A' || code[1] > 'Z' || parsed.Country == "" {
		r.Message = "服务数据缺失或格式异常"
		return r, none
	}
	parsed.Status = "ok"
	parsed.CountryCode = code
	parsed.Source = source
	parsed.Time = r.Time
	parsed.ProxyMode = r.ProxyMode
	// ASN/organization can be unavailable without invalidating valid location data.
	if parsed.ASN == "" {
		parsed.ASN = "未知"
	}
	if parsed.Organization == "" {
		parsed.Organization = "未知"
	}
	recommendTemplate(&parsed)
	return parsed, none
}
