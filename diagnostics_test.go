package main

import (
	"claudepro/internal/core"
	"claudepro/internal/platform"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNetworkResponses(t *testing.T) {
	valid := `{"ip":"203.0.113.8","country_name":"Japan","country_code":"JP","asn":"AS64496","org":"Example","city":"Tokyo","region":"Tokyo","timezone":"Asia/Tokyo"}`
	for _, tt := range []struct {
		name, body string
		code       int
		ok         bool
	}{
		{"success", valid, 200, true}, {"limited", "", 429, false}, {"unavailable", "", 503, false},
		{"invalid-json", "<html>", 200, false}, {"missing-fields", `{"ip":"203.0.113.8"}`, 200, false},
		{"invalid-ip", strings.Replace(valid, "203.0.113.8", "secret", 1), 200, false},
		{"provider-error", `{"error":true}`, 200, false}, {"oversize", strings.Repeat("x", 65537), 200, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != networkURL || req.Header.Get("Accept") != "application/json" {
					t.Fatal("unexpected request")
				}
				return &http.Response{StatusCode: tt.code, Body: io.NopCloser(strings.NewReader(tt.body))}, nil
			})}
			r := queryNetwork(context.Background(), client)
			if (r.Status == "ok") != tt.ok {
				t.Fatal(r)
			}
			if tt.ok && (r.City != "Tokyo" || r.Timezone != "Asia/Tokyo" || r.RecommendedPreset != "japan") {
				t.Fatal("location/recommendation missing", r)
			}
			if !tt.ok && (r.IP != "" || r.Message == "") {
				t.Fatal("failed result exposed data or no explanation", r)
			}
		})
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	r := queryNetwork(ctx, &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, errors.New("secret-proxy-password")
	})})
	if r.Status != "unknown" || strings.Contains(r.Message, "secret") {
		t.Fatal(r)
	}
}

func TestPresetsAndDemoDiagnostics(t *testing.T) {
	dir := t.TempDir()
	s := newService(platform.NewDemo(dir), dir, true, nil)
	presets, err := s.GetPresets()
	if err != nil || len(presets) != 6 {
		t.Fatal(presets, err)
	}
	for _, p := range presets {
		if p.Unavailable != "" || s.host.Validate(p.Target) != nil {
			t.Fatal(p)
		}
	}
	for _, p := range core.Presets(core.Catalog{}, platform.Kind) {
		if p.Unavailable == "" {
			t.Fatal("missing catalog enabled preset")
		}
	}
	t.Setenv("HTTPS_PROXY", "https://user:secret@example.invalid")
	os.WriteFile(filepath.Join(dir, "journal.json"), []byte("{broken"), 0600)
	r, err := s.GetDiagnostics(presets[0].Target)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range r.Checks {
		if c.ID == "recovery" && c.Status == "error" {
			found = true
		}
		if strings.Contains(c.Detail, "secret") {
			t.Fatal(c)
		}
	}
	if !found || len(r.Checks) < 8 {
		t.Fatal(r)
	}
	s.op.Lock()
	network, err := s.DetectNetwork()
	s.op.Unlock()
	if err != nil || network.IP != "203.0.113.10" {
		t.Fatal(network, err)
	}
	s.network.Lock()
	if _, err = s.DetectNetwork(); err == nil {
		t.Fatal("duplicate network call accepted")
	}
	s.network.Unlock()
}

type failingHost struct {
	*platform.Demo
	fail    string
	entered chan struct{}
	release chan struct{}
}

func (h *failingHost) Apply(field, value string) error {
	if field == h.fail {
		return errors.New("simulated UAC cancel")
	}
	if h.entered != nil {
		close(h.entered)
		<-h.release
		h.entered = nil
	}
	return h.Demo.Apply(field, value)
}

func TestSwitchToTransactions(t *testing.T) {
	for _, failure := range []string{"", "culture", "timezone"} {
		t.Run("failure-"+failure, func(t *testing.T) {
			dir := t.TempDir()
			host := &failingHost{Demo: platform.NewDemo(dir), fail: failure}
			s := newService(host, dir, true, nil)
			before, _ := host.Read()
			target := core.Target{Timezone: "Tokyo Standard Time", Culture: "ja-JP", Region: "122"}
			err := s.SwitchTo(target)
			if failure != "" {
				if err == nil || !strings.Contains(err.Error(), "目标已保存") {
					t.Fatal(err)
				}
				after, _ := host.Read()
				for k, v := range before {
					if !core.Equal(v, after[k]) {
						t.Fatal("rollback failed", k)
					}
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				journal, _ := os.ReadFile(filepath.Join(dir, "journal.json"))
				if s.SwitchTo(core.Target{Timezone: "Eastern Standard Time", Culture: "en-US", Region: "244"}) == nil {
					t.Fatal("active switch allowed")
				}
				after, _ := os.ReadFile(filepath.Join(dir, "journal.json"))
				if string(journal) != string(after) {
					t.Fatal("backup overwritten")
				}
				if err = s.Restore(nil); err != nil {
					t.Fatal(err)
				}
			}
			c, err := s.config()
			if err != nil || c.Version != 1 || c.Target != target {
				t.Fatal(c, err)
			}
		})
	}
}

func TestSwitchToPersistenceAndConcurrency(t *testing.T) {
	target := core.Target{Timezone: "Tokyo Standard Time", Culture: "ja-JP", Region: "122"}
	t.Run("save-failure", func(t *testing.T) {
		dir := t.TempDir()
		host := platform.NewDemo(dir)
		s := newService(host, dir, true, nil)
		// A read-only preferences file remains readable but cannot be replaced on Windows.
		core.AtomicJSON(filepath.Join(dir, "config.json"), core.Config{Version: 1, WorkDir: dir})
		os.Chmod(filepath.Join(dir, "config.json"), 0400)
		defer os.Chmod(filepath.Join(dir, "config.json"), 0600)
		if s.SwitchTo(target) == nil {
			t.Fatal("save failure ignored")
		}
		if _, err := os.Stat(filepath.Join(dir, "environment.json")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("settings modified before persistence", err)
		}
	})
	t.Run("concurrent", func(t *testing.T) {
		dir := t.TempDir()
		entered, release := make(chan struct{}), make(chan struct{})
		host := &failingHost{Demo: platform.NewDemo(dir), entered: entered, release: release}
		s := newService(host, dir, true, nil)
		done := make(chan error, 1)
		go func() { done <- s.SwitchTo(target) }()
		<-entered
		err := s.SwitchTo(target)
		close(release)
		if err == nil {
			t.Fatal("concurrent switch accepted")
		}
		if err = <-done; err != nil {
			t.Fatal(err)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		dir := t.TempDir()
		s := newService(platform.NewDemo(dir), dir, true, nil)
		if s.SwitchTo(core.Target{Timezone: "invalid"}) == nil {
			t.Fatal("invalid target accepted")
		}
		if _, err := os.Stat(filepath.Join(dir, "config.json")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("invalid target persisted")
		}
	})
}

func TestTemplateRecommendation(t *testing.T) {
	for _, tt := range []struct{ country, timezone, want string }{
		{"US", "America/Los_Angeles", "us-west"}, {"US", "US/Pacific", "us-west"},
		{"US", "America/New_York", "us-east"}, {"US", "America/Detroit", "us-east"},
		{"US", "America/Indiana/Indianapolis", "us-east"}, {"US", "America/Chicago", ""},
		{"US", "America/Denver", ""}, {"US", "America/Phoenix", ""}, {"US", "America/Anchorage", ""},
		{"US", "Pacific/Honolulu", ""}, {"US", "", ""}, {"US", "Europe/London", ""},
		{"GB", "Europe/London", "uk"}, {"DE", "Europe/Berlin", "germany"},
		{"JP", "Asia/Tokyo", "japan"}, {"SG", "Asia/Singapore", "singapore"},
		{"CA", "America/Los_Angeles", ""},
	} {
		t.Run(tt.country+tt.timezone, func(t *testing.T) {
			r := NetworkReport{Status: "ok", CountryCode: tt.country, Country: tt.country, Timezone: tt.timezone}
			recommendTemplate(&r)
			if r.RecommendedPreset != tt.want || r.RecommendationReason == "" {
				t.Fatal(r)
			}
			r.Status = "unknown"
			recommendTemplate(&r)
			if r.RecommendedPreset != "" || r.RecommendationReason != "" {
				t.Fatal("stale recommendation", r)
			}
		})
	}
}
