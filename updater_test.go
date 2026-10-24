package main

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func TestMatchReleaseAsset(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "SHA256SUMS"},
		{Name: "claudepro.exe.sha256"},
		{Name: "claudepro.exe"},
		{Name: "ClaudePro-macos-universal.dmg"},
		{Name: "claudepro-macos-universal.zip"},
		{Name: "notes.md"},
	}
	cases := []struct {
		plat string
		want string
	}{
		{"windows", "claudepro.exe"},
		{"darwin", "claudepro-macos-universal.zip"},
		{"linux", ""},
	}
	for _, c := range cases {
		idx := matchReleaseAsset(updater.CheckRequest{Platform: c.plat, Arch: "amd64"}, assets)
		if c.want == "" {
			if idx != -1 {
				t.Fatalf("%s matched %s", c.plat, assets[idx].Name)
			}
			continue
		}
		if idx < 0 || assets[idx].Name != c.want {
			t.Fatalf("%s: idx=%d", c.plat, idx)
		}
	}
	idx := matchReleaseAsset(updater.CheckRequest{Platform: "windows", Arch: "amd64"}, []github.ReleaseAsset{{Name: "claudepro-windows-amd64.exe"}})
	if idx != 0 {
		t.Fatal("windows-amd64 alias")
	}
	idx = matchReleaseAsset(updater.CheckRequest{Platform: "darwin", Arch: "arm64"}, []github.ReleaseAsset{{Name: "ClaudePro-macos-universal.dmg"}})
	if idx != -1 {
		t.Fatal("dmg is not an in-app update payload")
	}
}

func TestUpdateInfoFromNil(t *testing.T) {
	if updateInfoFrom(nil, true) != nil {
		t.Fatal("nil release")
	}
	info := updateInfoFrom(&updater.Release{Version: "0.2.0", Name: "r", Notes: "n", Artifact: updater.Artifact{Filename: "claudepro.exe", Size: 12}}, true)
	if info == nil || info.Version != "0.2.0" || !info.AllowInstall || info.Size != 12 {
		t.Fatal(info)
	}
}
