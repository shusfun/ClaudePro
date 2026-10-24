package main

import (
	"claudepro/internal/core"
	"claudepro/internal/platform"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceRestartRecoveryAndTargetLock(t *testing.T) {
	dir := t.TempDir()
	host := platform.NewDemo(dir)
	s := newService(host, dir, true, nil)
	before, _ := host.Read()
	c := core.Config{Version: 1, Target: core.Target{Timezone: "Eastern Standard Time", Culture: "en-US", Region: "244"}, WorkDir: dir}
	if err := s.SaveConfig(c); err != nil {
		t.Fatal(err)
	}
	if err := s.Switch(); err != nil {
		t.Fatal(err)
	}
	c.Target.Timezone = "Pacific Standard Time"
	if s.SaveConfig(c) == nil {
		t.Fatal("target changed while active")
	}
	if s.closeStatus() != "confirm" {
		t.Fatal("missing exit warning")
	}
	s = newService(platform.NewDemo(dir), dir, true, nil)
	state, err := s.GetState()
	if err != nil || state.View.Journal.Status != "active" {
		t.Fatal(state, err)
	}
	if err = s.Restore(nil); err != nil {
		t.Fatal(err)
	}
	after, _ := s.host.Read()
	for k, v := range before {
		if !core.Equal(v, after[k]) {
			t.Fatal("restore failed", k)
		}
	}
	if s.closeStatus() != "ok" {
		t.Fatal("exit incorrectly blocked")
	}
	if err = s.SaveConfig(c); err != nil {
		t.Fatal(err)
	}
}
func TestBusyExitAndConfigurationCorruption(t *testing.T) {
	dir := t.TempDir()
	s := newService(platform.NewDemo(dir), dir, true, nil)
	called := false
	s.quit = func() { called = true }
	s.op.Lock()
	if s.ConfirmExit() == nil || s.closeStatus() != "busy" || called {
		t.Fatal("busy exit permitted")
	}
	s.op.Unlock()
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("{bad"), 0600)
	if _, err := s.GetState(); err == nil {
		t.Fatal("bad config accepted")
	}
}

func TestDamagedConfigDoesNotHideRecovery(t *testing.T) {
	dir := t.TempDir()
	s := newService(platform.NewDemo(dir), dir, true, nil)
	c := core.Config{Version: 1, Target: core.Target{Timezone: "Eastern Standard Time", Culture: "en-US", Region: "244"}, WorkDir: dir}
	if err := s.SaveConfig(c); err != nil {
		t.Fatal(err)
	}
	if err := s.Switch(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetState()
	if err != nil || v.Warning == "" || !core.Active(v.View.Journal) {
		t.Fatal(v, err)
	}
	if err = s.Restore(nil); err != nil {
		t.Fatal(err)
	}
	state, err := s.GetState()
	if err != nil || state.View.Journal.Status != "restored" {
		t.Fatal(state, err)
	}
}
