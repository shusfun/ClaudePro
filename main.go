package main

import (
	"claudepro/internal/platform"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if handled := platformCommand(os.Args); handled {
		return
	}
	host := platform.NewHost()
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
	dir, err := dataDir(demo)
	if err != nil {
		message("无法定位备份目录：" + err.Error())
		return
	}
	var adapter Host = host
	if demo {
		adapter = platform.NewDemo(dir)
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		message("无法创建备份目录：" + err.Error())
		return
	}
	var app *application.App
	var window *application.WebviewWindow
	svc := newService(adapter, dir, demo, func(msg string) {
		if app != nil {
			app.Event.Emit("operation-progress", msg)
		}
	})
	var allowExit atomic.Bool
	app = application.New(application.Options{
		Name: "ClaudePro", Description: "系统环境切换与精确还原",
		Windows:  windowsOptions(dir, demo),
		Mac:      application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
		Services: []application.Service{application.NewService(svc)},
		Assets:   application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "local.claudepro.environment." + strconv.FormatBool(demo),
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if window != nil {
					window.UnMinimise()
					window.Show()
					window.Focus()
				}
			},
		},
		// macOS 的 ⌘Q 与 Dock 退出走这里；与窗口关闭共用同一套确认。
		ShouldQuit: func() bool {
			if allowExit.Load() {
				return true
			}
			return requestClose(app, svc)
		},
		ErrorHandler: func(err error) { message("应用发生错误：" + err.Error()) },
	})
	if !checkRuntime(app) {
		return
	}
	svc.quit = func() { allowExit.Store(true); app.Quit() }
	if !demo {
		svc.updater = setupUpdater(app)
	}
	title := "ClaudePro"
	if demo {
		title += " [模拟模式]"
	}
	window = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: title, Width: 1120, Height: 780, MinWidth: 820, MinHeight: 600,
		// 自绘标题栏：Windows 去掉原生边框但保留系统阴影与圆角；macOS 隐藏标题并保留红绿灯。
		Frameless:        platform.Kind == "windows",
		BackgroundColour: application.NewRGB(242, 242, 247), URL: "/",
		Mac: application.MacWindow{
			TitleBar:                application.MacTitleBarHiddenInset,
			InvisibleTitleBarHeight: 0,
		},
	})
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if allowExit.Load() {
			return
		}
		if !requestClose(app, svc) {
			e.Cancel()
		}
	})
	if err = app.Run(); err != nil {
		message("无法启动 ClaudePro。\n" + err.Error())
	}
}

// requestClose 返回 true 表示可以直接关闭；否则通知前端处理。
func requestClose(app *application.App, svc *Service) bool {
	switch svc.closeStatus() {
	case "busy":
		app.Event.Emit("close-blocked", "正在操作，请完成后退出")
		return false
	case "confirm":
		app.Event.Emit("close-requested", true)
		return false
	}
	return true
}

func errText(e error) any {
	if e != nil {
		return e.Error()
	}
	return nil
}
