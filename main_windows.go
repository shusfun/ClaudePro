package main

import (
	"claudepro/internal/platform"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
)

func message(text string) {
	t, _ := windows.UTF16PtrFromString(text)
	c, _ := windows.UTF16PtrFromString("ClaudePro")
	windows.MessageBox(0, t, c, windows.MB_OK|windows.MB_ICONINFORMATION)
}

// platformCommand 处理只在 Windows 存在的内部子命令。
func platformCommand(args []string) bool {
	if len(args) < 2 || args[1] != "--timezone-helper" {
		return false
	}
	if len(args) != 4 {
		os.Exit(2)
	}
	pid, err := strconv.ParseUint(args[3], 10, 32)
	if err != nil {
		os.Exit(2)
	}
	if err = platform.TimezoneHelper(args[2], uint32(pid)); err != nil {
		os.Exit(1)
	}
	return true
}

func dataDir(demo bool) (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", errors.New("缺少 LOCALAPPDATA")
	}
	dir := filepath.Join(base, "ClaudePro")
	if demo {
		dir += "-Demo"
	}
	return dir, nil
}

func windowsOptions(dir string, demo bool) application.WindowsOptions {
	opts := application.WindowsOptions{WebviewUserDataPath: filepath.Join(dir, "WebView2")}
	// Opt-in local debugging is available only with the simulated adapter.
	if demo {
		if port, err := strconv.Atoi(os.Getenv("CLAUDEPRO_DEMO_CDP_PORT")); err == nil && port >= 1024 && port <= 65535 {
			opts.AdditionalBrowserArgs = []string{fmt.Sprintf("--remote-debugging-port=%d", port), "--remote-debugging-address=127.0.0.1"}
		}
	}
	return opts
}

func checkRuntime(app *application.App) bool {
	if version, _ := app.Env.Info().PlatformInfo["WebView2"].(string); version == "" {
		message("未检测到可用的 Microsoft Edge WebView2 Runtime。\n请安装 Evergreen Runtime 后重新打开 ClaudePro：\nhttps://developer.microsoft.com/microsoft-edge/webview2/\n\n本次未修改任何环境设置，已有恢复记录仍保留。")
		return false
	}
	return true
}
