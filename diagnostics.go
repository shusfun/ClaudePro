package main

import (
	"claudepro/internal/core"
	"claudepro/internal/platform"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Check struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
	Action string `json:"action"`
}
type LocalReport struct {
	Time   string  `json:"time"`
	Checks []Check `json:"checks"`
}
type NetworkReport struct {
	RetryAt              string `json:"retryAt"`
	Notes                string `json:"notes"`
	Cached               bool   `json:"cached"`
	LocationConflict     bool   `json:"locationConflict"`
	Time                 string `json:"time"`
	Source               string `json:"source"`
	Status               string `json:"status"`
	Message              string `json:"message"`
	IP                   string `json:"ip"`
	Country              string `json:"country"`
	CountryCode          string `json:"countryCode"`
	ASN                  string `json:"asn"`
	Organization         string `json:"organization"`
	ProxyMode            string `json:"proxyMode"`
	City                 string `json:"city"`
	Region               string `json:"region"`
	Timezone             string `json:"timezone"`
	RecommendedPreset    string `json:"recommendedPreset"`
	RecommendationReason string `json:"recommendationReason"`
}

func (s *Service) GetPresets() ([]core.Preset, error) {
	c, err := s.host.Catalog()
	if err != nil {
		return nil, err
	}
	return core.Presets(c, platform.Kind), nil
}

func (s *Service) GetDiagnostics(target core.Target) (LocalReport, error) {
	if !s.op.TryLock() {
		return LocalReport{}, errors.New("操作正在进行，请稍后检查")
	}
	defer s.op.Unlock()
	r := LocalReport{Time: time.Now().UTC().Format(time.RFC3339), Checks: []Check{}}
	add := func(id, name, status, detail, action string) {
		r.Checks = append(r.Checks, Check{id, name, status, detail, action})
	}
	env, err := s.host.Read()
	if err != nil {
		add("environment", "系统设置", "unknown", "无法读取当前设置，请重新检查", "refresh")
	} else {
		for _, f := range core.Fields {
			status, detail := "info", "当前："+env[f].Label+"；尚未选择目标"
			if v := target.Value(f); v != "" {
				detail = "当前：" + env[f].Label + "；目标：" + v
				if env[f].Label == v {
					status = "ok"
					detail += "（一致）"
				} else {
					detail += "（待切换，不代表故障）"
				}
			}
			name := map[string]string{"timezone": "系统时区", "culture": "区域格式", "region": "国家 / 地区"}[f]
			add(f, name, status, detail, "settings")
		}
	}
	v, err := s.engine.Inspect()
	if err != nil {
		add("recovery", "恢复记录", "error", "无法读取恢复记录，修改功能已停用；请查看恢复指南", "history")
	} else if core.Active(v.Journal) {
		status, detail := "info", "目标环境生效中；更换地区前请恢复原环境"
		if v.Journal.Status != "active" || len(v.Conflicts) > 0 {
			status = "warn"
			detail = "存在未完成操作或配置冲突，请继续恢复"
		}
		add("recovery", "恢复记录", status, detail, "restore")
	} else {
		add("recovery", "恢复记录", "ok", "没有未恢复操作", "history")
	}
	c, configErr := s.config()
	if configErr != nil {
		add("config", "应用配置", "error", "配置文件无法读取，请查看恢复指南", "history")
	}
	clients, discoverErr := s.host.Discover()
	for _, kind := range []string{"cli", "desktop"} {
		path, manual, name := clients.CLI, c.CLIPath, "Claude Code CLI"
		if kind == "desktop" {
			path, manual, name = clients.Desktop, c.DesktopPath, "Claude Desktop"
		}
		status, detail := "ok", "已自动发现客户端"
		if manual != "" {
			if platform.ValidateClientPath(kind, manual) != nil {
				status = "warn"
				detail = "手动指定的客户端路径不可用"
			} else {
				detail = "手动指定的客户端路径有效"
			}
		} else if discoverErr != nil {
			status = "unknown"
			detail = "客户端自动发现失败"
		} else if path == "" {
			status = "warn"
			detail = "未找到客户端，请配置路径"
		}
		add(kind, name, status, detail, "settings")
	}
	processes, err := s.host.Running()
	if err != nil {
		add("sessions", "Claude 会话", "unknown", "无法检查运行状态，请重试", "refresh")
	} else if len(processes) > 0 {
		add("sessions", "Claude 会话", "warn", fmt.Sprintf("检测到 %d 个进程；请保存并关闭后重新检查", len(processes)), "refresh")
	} else {
		add("sessions", "Claude 会话", "ok", "未检测到运行中的 Claude 进程", "refresh")
	}
	if s.demo {
		add("proxy", "系统代理", "info", "模拟数据：未配置手动代理或 PAC", "refresh")
		add("proxy-env", "代理环境变量", "info", "模拟数据：未配置", "refresh")
	} else {
		manual, pac, err := platform.SystemProxy()
		if err != nil {
			add("proxy", "系统代理", "unknown", "无法读取当前用户代理设置", "refresh")
		} else {
			add("proxy", "系统代理", "info", fmt.Sprintf("当前用户系统代理：手动代理启用=%t；PAC 已配置=%t；不代表客户端实际路由", manual, pac), "refresh")
		}
		var names []string
		for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "no_proxy", "all_proxy"} {
			if os.Getenv(name) != "" {
				names = append(names, name)
			}
		}
		detail := "未配置代理环境变量"
		if len(names) > 0 {
			detail = "已配置：" + strings.Join(names, "、") + "（仅显示变量名）"
		}
		add("proxy-env", "代理环境变量", "info", detail, "refresh")
	}
	return r, nil
}

const networkURL = "https://ipapi.co/json/"

func (s *Service) DetectNetwork() (NetworkReport, error) {
	if !s.network.TryLock() {
		return NetworkReport{}, errors.New("网络检测正在进行")
	}
	defer s.network.Unlock()
	if s.demo {
		r := NetworkReport{Time: time.Now().UTC().Format(time.RFC3339), Source: "模拟数据（未联网）", Status: "ok", IP: "203.0.113.10", Country: "United States", CountryCode: "US", City: "Los Angeles", Region: "California", Timezone: "America/Los_Angeles", ASN: "AS64496", Organization: "演示网络", ProxyMode: "模拟"}
		recommendTemplate(&r)
		return r, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.networkState.detect(ctx, &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, time.Now()), nil
}

// Match the reported IANA timezone, never guess a US coast from longitude,
// current UTC offset (DST), or a state spanning multiple timezones.
func recommendTemplate(r *NetworkReport) {
	r.RecommendedPreset = ""
	r.RecommendationReason = ""
	if r.Status != "ok" {
		return
	}
	if r.CountryCode == "US" {
		switch r.Timezone {
		case "America/Los_Angeles", "US/Pacific", "US/Pacific-New":
			r.RecommendedPreset = "us-west"
			r.RecommendationReason = "出口时区为 " + r.Timezone + "（美国太平洋时间），推荐美国西部模板。"
		case "America/New_York", "America/Detroit", "America/Indiana/Indianapolis", "America/Indiana/Marengo", "America/Indiana/Petersburg", "America/Indiana/Vevay", "America/Indiana/Vincennes", "America/Indiana/Winamac", "America/Kentucky/Louisville", "America/Kentucky/Monticello", "US/Eastern", "US/Michigan", "US/East-Indiana":
			r.RecommendedPreset = "us-east"
			r.RecommendationReason = "出口时区为 " + r.Timezone + "（美国东部时间），推荐美国东部模板。"
		case "":
			r.RecommendationReason = "检测到美国出口，但服务未提供时区，暂不能推荐美东或美西；请重试或自定义。"
		default:
			r.RecommendationReason = "出口时区为 " + r.Timezone + "，未匹配现有美东 / 美西模板；美国中部、山地、阿拉斯加和夏威夷等时区请使用自定义。"
		}
		return
	}
	for _, p := range core.Presets(core.Catalog{}, platform.Kind) {
		if p.Country == r.CountryCode {
			r.RecommendedPreset = p.ID
			r.RecommendationReason = "出口国家为 " + r.Country + "，推荐" + p.Name + "模板。"
			return
		}
	}
	r.RecommendationReason = "当前出口国家没有内置模板，请使用自定义区域设置。"
}
