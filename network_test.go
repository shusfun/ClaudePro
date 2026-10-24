package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const geoWest = `{"ip":"203.0.113.8","country":"United States","country_code":"US","city":"Los Angeles","region":"California","timezone":"America/Los_Angeles","asn":701,"organization_name":"Example"}`
const whoWest = `{"success":true,"ip":"203.0.113.8","country":"United States","country_code":"US","city":"Los Angeles","region":"California","timezone":{"id":"America/Los_Angeles"},"connection":{"asn":701,"org":"Example"}}`
const infoWest = `{"ip":"203.0.113.8","country":"US","city":"Los Angeles","region":"California","timezone":"America/Los_Angeles","org":"AS701 Example"}`

func fakeResponse(code int, body string) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}
}

func TestFallbackCacheAndCooldown(t *testing.T) {
	var mu sync.Mutex
	calls := map[string]int{}
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		calls[r.URL.String()]++
		mu.Unlock()
		switch r.URL.String() {
		case geoNetworkURL:
			resp := fakeResponse(429, "")
			resp.Header.Set("Retry-After", "600")
			return resp, nil
		case backupNetworkURL:
			return fakeResponse(200, whoWest), nil
		case infoNetworkURL:
			return fakeResponse(200, infoWest), nil
		default:
			t.Error("quota-limited final source should not be required")
			return fakeResponse(503, ""), nil
		}
	})}
	now := time.Now()
	state := networkState{}
	r := state.detect(context.Background(), client, now)
	if r.Status != "ok" || r.RecommendedPreset != "us-west" || r.LocationConflict || !strings.Contains(r.Notes, "429") {
		t.Fatal(r)
	}
	if !state.blocked[geoNetworkURL].Equal(now.Add(10 * time.Minute)) {
		t.Fatal(state.blocked)
	}
	if !state.detect(context.Background(), client, now.Add(time.Second)).Cached {
		t.Fatal("cache not reused")
	}
	r = state.detect(context.Background(), client, now.Add(61*time.Second))
	if r.Status != "ok" || !strings.Contains(r.Notes, "限流冷却中") {
		t.Fatal(r)
	}
	mu.Lock()
	defer mu.Unlock()
	if calls[geoNetworkURL] != 1 || calls[backupNetworkURL] != 2 {
		t.Fatal(calls)
	}
}

func TestDisagreementAndFailure(t *testing.T) {
	now := time.Now()
	stage := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if stage == 1 {
			return fakeResponse(503, ""), nil
		}
		if r.URL.String() == geoNetworkURL && stage == 0 {
			return fakeResponse(200, geoWest), nil
		}
		if r.URL.String() == backupNetworkURL {
			return fakeResponse(200, strings.ReplaceAll(whoWest, "America/Los_Angeles", "America/New_York")), nil
		}
		return fakeResponse(429, ""), nil
	})}
	s := networkState{}
	r := s.detect(context.Background(), client, now)
	if !r.LocationConflict || r.RecommendedPreset != "" || !strings.Contains(r.Notes, "America/") {
		t.Fatal(r)
	}
	stage = 1
	r = s.detect(context.Background(), client, now.Add(time.Minute))
	if r.Status != "unknown" || r.RecommendedPreset != "" || r.RetryAt == "" {
		t.Fatal(r)
	}
	stage = 2
	r = s.detect(context.Background(), client, now.Add(2*time.Minute))
	if r.Status != "ok" || !r.LocationConflict || r.RecommendedPreset != "" {
		t.Fatal("lost prior same-IP disagreement after outage", r)
	}
}

func TestProviderFormatsAndRetryAfter(t *testing.T) {
	for _, tt := range []struct {
		source, body string
		ok           bool
	}{
		{geoNetworkURL, geoWest, true}, {backupNetworkURL, whoWest, true}, {infoNetworkURL, infoWest, true},
		{backupNetworkURL, `{"success":false}`, false}, {geoNetworkURL, `<html>captcha</html>`, false},
		{geoNetworkURL, strings.Replace(geoWest, "203.0.113.8", "127.0.0.1", 1), false},
		{geoNetworkURL, strings.Replace(geoWest, "203.0.113.8", "100.64.0.1", 1), false},
		{geoNetworkURL, strings.Replace(geoWest, "203.0.113.8", "192.168.1.1", 1), false},
	} {
		client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) { return fakeResponse(200, tt.body), nil })}
		r, _ := queryProvider(context.Background(), client, tt.source, time.Now())
		if (r.Status == "ok") != tt.ok {
			t.Fatal(r)
		}
		if tt.ok && (r.RecommendedPreset != "us-west" || r.Source != tt.source || r.ASN != "AS701") {
			t.Fatal(r)
		}
	}
	now := time.Now().Truncate(time.Second)
	for _, v := range []string{"600", now.Add(10 * time.Minute).UTC().Format(http.TimeFormat)} {
		if !retryAfter(v, now).Equal(now.Add(10 * time.Minute)) {
			t.Fatal(v)
		}
	}
	if !retryAfter("bad", now).Equal(now.Add(5 * time.Minute)) {
		t.Fatal("missing default cooldown")
	}
}

func TestStaggerAndCancellation(t *testing.T) {
	started, finished := make(chan struct{}), make(chan struct{})
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() == geoNetworkURL {
			close(started)
			<-r.Context().Done()
			close(finished)
			return nil, r.Context().Err()
		}
		if r.URL.String() == backupNetworkURL {
			return fakeResponse(200, whoWest), nil
		}
		return fakeResponse(200, infoWest), nil
	})}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s := networkState{}
	r := s.detect(ctx, client, time.Now())
	<-started
	if r.Status != "ok" {
		t.Fatal(r)
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("losing request not cancelled")
	}
}

// Explicitly opt in for a read-only live smoke test. Ordinary tests never query providers.
func TestLiveNetworkProviders(t *testing.T) {
	if os.Getenv("CLAUDEPRO_LIVE_NETWORK_TEST") != "1" {
		t.Skip("offline by default")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s := networkState{}
	r := s.detect(ctx, &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, time.Now())
	t.Logf("status=%s source=%s country=%s city=%s timezone=%s conflict=%t recommendation=%s notes=%s", r.Status, r.Source, r.CountryCode, r.City, r.Timezone, r.LocationConflict, r.RecommendedPreset, r.Notes)
	if r.Status != "ok" {
		t.Fatal(r.Message)
	}
}
