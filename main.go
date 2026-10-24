package main

import (
	"claudepro/internal/platform"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
)

//go:embed all:frontend/dist
var assets embed.FS

func message(text string) {
	t, _ := windows.UTF16PtrFromString(text)
	c, _ := windows.UTF16PtrFromString("ClaudePro")
	windows.MessageBox(0, t, c, windows.MB_OK|windows.MB_ICONINFORMATION)
}
func main() {
	if len(os.Args) > 1 && os.Args[1] == "--timezone-helper" {
		if len(os.Args) != 4 {
			os.Exit(2)
		}
		pid, err := strconv.ParseUint(os.Args[3], 10, 32)
		if err != nil {
			os.Exit(2)
		}
		if err = platform.TimezoneHelper(os.Args[2], uint32(pid)); err != nil {
			os.Exit(1)
		}
		return
	}
	host := &platform.Windows{}
	if len(os.Args) > 1 && os.Args[1] == "--diagnose" {
		env, e1 := host.Read()
		clients, e2 := host.Discover()
		catalog, e3 := host.Catalog()
		running, e4 := host.Running()
		b, _ := json.MarshalIndent(map[string]any{"environment": env, "clients": clients, "catalog": catalog, "running": running, "errors": []any{errText(e1), errText(e2), errText(e3), errText(e4)}}, "", "  ")
		fmt.Println(string(b))
		return
	}
	demo := len(os.Args) > 1 && os.Args[1] == "--demo"
	dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "ClaudePro")
	var adapter Host = host
	if demo {
		dir += "-Demo"
		adapter = platform.NewDemo(dir)
	}
	name, _ := windows.UTF16PtrFromString("Local\\ClaudePro.EnvironmentSwitcher" + strconv.FormatBool(demo))
	mutex, err := windows.CreateMutex(nil, false, name)
	if err == windows.ERROR_ALREADY_EXISTS {
		if mutex != 0 {
			windows.CloseHandle(mutex)
		}
		message("ClaudePro 已在运行，请切换到已有窗口。")
		return
	}
	if err != nil {
		message(err.Error())
		return
	}
	defer windows.CloseHandle(mutex)
	if err = os.MkdirAll(dir, 0700); err != nil {
		message("无法创建备份目录：" + err.Error())
		return
	}
	var app *application.App
	winOptions := application.WindowsOptions{WebviewUserDataPath: filepath.Join(dir, "WebView2")}
	// Opt-in local debugging is available only with the simulated adapter.
	if demo {
		if port, err := strconv.Atoi(os.Getenv("CLAUDEPRO_DEMO_CDP_PORT")); err == nil && port >= 1024 && port <= 65535 {
			winOptions.AdditionalBrowserArgs = []string{fmt.Sprintf("--remote-debugging-port=%d", port), "--remote-debugging-address=127.0.0.1"}
		}
	}
	svc := newService(adapter, dir, demo, func(msg string) {
		if app != nil {
			app.Event.Emit("operation-progress", msg)
		}
	})
	app = application.New(application.Options{
		Name: "ClaudePro", Description: "Windows 环境切换与精确还原",
		Windows:      winOptions,
		Services:     []application.Service{application.NewService(svc)},
		Assets:       application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		ErrorHandler: func(err error) { message("应用发生错误：" + err.Error()) },
	})
	if version, _ := app.Env.Info().PlatformInfo["WebView2"].(string); version == "" {
		message("未检测到可用的 Microsoft Edge WebView2 Runtime。\n请安装 Evergreen Runtime 后重新打开 ClaudePro：\nhttps://developer.microsoft.com/microsoft-edge/webview2/\n\n本次未修改任何环境设置，已有恢复记录仍保留。")
		return
	}
	var allowExit atomic.Bool
	svc.quit = func() { allowExit.Store(true); app.Quit() }
	title := "ClaudePro · 环境切换"
	if demo {
		title += " [模拟模式]"
	}
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: title, Width: 1160, Height: 820, MinWidth: 850, MinHeight: 640,
		BackgroundColour: application.NewRGB(246, 247, 249), URL: "/",
	})
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if allowExit.Load() {
			return
		}
		switch svc.closeStatus() {
		case "busy":
			e.Cancel()
			app.Event.Emit("close-blocked", "正在操作，请完成后退出")
		case "confirm":
			e.Cancel()
			app.Event.Emit("close-requested", true)
		}
	})
	if err = app.Run(); err != nil {
		message("无法启动 ClaudePro。请检查 Microsoft Edge WebView2 Runtime 是否已安装。\n" + err.Error())
	}
}
func errText(e error) any {
	if e != nil {
		return e.Error()
	}
	return nil
}
