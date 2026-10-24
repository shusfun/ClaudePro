package main

import (
	"context"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

const updateRepo = "shusfun/ClaudePro"

type appUpdater interface {
	Check(ctx context.Context) (*updater.Release, error)
	DownloadAndInstall(ctx context.Context) error
	Restart(ctx context.Context) error
	SkipVersion(v string)
	CurrentVersion() string
}

func setupUpdater(app *application.App) appUpdater {
	gh, err := github.New(github.Config{
		Repository:    updateRepo,
		Prerelease:    true,
		ChecksumAsset: "SHA256SUMS",
		AssetMatcher:  matchReleaseAsset,
	})
	if err != nil {
		return nil
	}
	if err = app.Updater.Init(updater.Config{
		CurrentVersion: appVersion,
		Providers:      []updater.Provider{gh},
		Window:         updater.WindowNone,
	}); err != nil {
		return nil
	}
	return app.Updater
}

func matchReleaseAsset(req updater.CheckRequest, assets []github.ReleaseAsset) int {
	plat := strings.ToLower(req.Platform)
	for i, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.HasSuffix(name, ".sig") || strings.HasSuffix(name, ".asc") || strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".sha512") {
			continue
		}
		if name == "sha256sums" || strings.HasPrefix(name, "sha256sums.") {
			continue
		}
		switch plat {
		case "windows":
			if name == "claudepro.exe" {
				return i
			}
			if strings.Contains(name, "windows") && (strings.Contains(name, "amd64") || strings.Contains(name, "x64") || strings.Contains(name, "x86_64")) {
				return i
			}
		case "darwin":
			if name == "claudepro-macos-universal.zip" {
				return i
			}
			if strings.Contains(name, "macos") || strings.Contains(name, "darwin") {
				return i
			}
		}
	}
	return -1
}

func updateInfoFrom(rel *updater.Release, install bool) *UpdateInfo {
	if rel == nil {
		return nil
	}
	return &UpdateInfo{
		Version:      rel.Version,
		Name:         rel.Name,
		Notes:        rel.Notes,
		Filename:     rel.Artifact.Filename,
		Size:         rel.Artifact.Size,
		AllowInstall: install,
	}
}

func checkRelease(u appUpdater, timeout time.Duration) (*updater.Release, error) {
	if u == nil {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return u.Check(ctx)
}
