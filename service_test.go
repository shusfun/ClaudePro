package main

import (
	"claudepro/internal/core"
	"claudepro/internal/platform"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

type fakeUpdater struct {
	rel      *updater.Release
	err      error
	skipped  string
	installs int
	restarts int
}

func (f *fakeUpdater) Check(context.Context) (*updater.Release, error) { return f.rel, f.err }
func (f *fakeUpdater) DownloadAndInstall(context.Context) error {
	f.installs++
	return f.err
}
func (f *fakeUpdater) Restart(context.Context) error { f.restarts++; return f.err }
func (f *fakeUpdater) SkipVersion(v string)          { f.skipped = v }
func (f *fakeUpdater) CurrentVersion() string        { return "0.1.0" }


func TestServiceRestartRecoveryAndTargetLock(t *testing.T) {
	dir := t.TempDir()
	host := platform.NewDemo(dir)
	s := newService(host, dir, true, nil)
	before, _ := host.Read()
	c := core.Config{Version: 1, Target: demoPreset("us-east"), WorkDir: dir}
	if err := s.SaveConfig(c); err != nil {
		t.Fatal(err)
	}
	if err := s.Switch(); err != nil {
		t.Fatal(err)
	}
	c.Target = demoPreset("us-west")
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
	c := core.Config{Version: 1, Target: demoPreset("us-east"), WorkDir: dir}
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

func TestDemoSkipsUpdates(t *testing.T) {
	s := newService(platform.NewDemo(t.TempDir()), t.TempDir(), true, nil)
	if _, err := s.CheckUpdate(); err == nil {
		t.Fatal("demo check")
	}
	if s.InstallUpdate() == nil || s.RestartUpdate() == nil {
		t.Fatal("demo install")
	}
	st, err := s.GetState()
	if err != nil || st.AllowUpdateInstall || st.AppVersion == "" {
		t.Fatal(st, err)
	}
}

func TestUpdateBusyAndSkip(t *testing.T) {
	dir := t.TempDir()
	s := newService(platform.NewDemo(dir), dir, false, nil)
	fake := &fakeUpdater{rel: &updater.Release{Version: "0.2.0", Artifact: updater.Artifact{Filename: "claudepro.exe"}}}
	s.updater = fake
	info, err := s.CheckUpdate()
	if err != nil || info == nil || info.Version != "0.2.0" {
		t.Fatal(info, err)
	}
	if err = s.SkipUpdate("0.2.0"); err != nil || fake.skipped != "0.2.0" {
		t.Fatal(err, fake.skipped)
	}
	s.op.Lock()
	if s.InstallUpdate() == nil || s.RestartUpdate() == nil {
		s.op.Unlock()
		t.Fatal("busy update permitted")
	}
	s.op.Unlock()
}
