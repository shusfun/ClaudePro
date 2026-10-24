package main

import (
	"claudepro/internal/core"
	"claudepro/internal/platform"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Host interface {
	core.Adapter
	Catalog() (core.Catalog, error)
	Discover() (core.Clients, error)
	LaunchCLI(string, string) error
	LaunchDesktop(string, string) error
}
type State struct {
	View               core.View   `json:"view"`
	Config             core.Config `json:"config"`
	DataDir            string      `json:"dataDir"`
	Demo               bool        `json:"demo"`
	Warning            string      `json:"warning"`
	Platform           string      `json:"platform"`
	AppVersion         string      `json:"appVersion"`
	AllowUpdateInstall bool        `json:"allowUpdateInstall"`
}
type UpdateInfo struct {
	Version      string `json:"version"`
	Name         string `json:"name"`
	Notes        string `json:"notes"`
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	AllowInstall bool   `json:"allowInstall"`
}
type Service struct {
	op           sync.Mutex
	network      sync.Mutex
	networkState networkState
	engine       *core.Engine
	host         Host
	dir          string
	demo         bool
	quit         func()
	updater      appUpdater
}

func newService(host Host, dir string, demo bool, notify func(string)) *Service {
	return &Service{host: host, dir: dir, demo: demo, engine: core.New(host, core.FileStore{Dir: dir}, notify)}
}
func (s *Service) config() (core.Config, error) {
	var c core.Config
	err := core.ReadJSON(filepath.Join(s.dir, "config.json"), &c)
	if errors.Is(err, os.ErrNotExist) {
		home, _ := os.UserHomeDir()
		return core.Config{Version: 1, WorkDir: home}, nil
	}
	if err != nil {
		return c, err
	}
	if c.Version != 1 {
		return c, errors.New("配置版本不受支持")
	}
	return c, nil
}
func (s *Service) GetState() (State, error) {
	if !s.op.TryLock() {
		return State{}, errors.New("操作正在进行，请稍后刷新")
	}
	defer s.op.Unlock()
	v, err := s.engine.Inspect()
	if err != nil {
		return State{}, err
	}
	c, err := s.config()
	if err != nil && v.Journal != nil {
		// A damaged preferences file must never hide a valid recovery journal.
		c = core.Config{Version: 1, Target: v.Journal.Target}
		return State{View: v, Config: c, DataDir: s.dir, Demo: s.demo, Platform: platform.Kind, AppVersion: appVersion, AllowUpdateInstall: s.canInstallUpdate(), Warning: "设置文件损坏，但恢复记录仍可读取。若有未完成操作，请先恢复原环境，再修复 config.json：" + err.Error()}, nil
	}
	return State{View: v, Config: c, DataDir: s.dir, Demo: s.demo, Platform: platform.Kind, AppVersion: appVersion, AllowUpdateInstall: s.canInstallUpdate()}, err
}
func (s *Service) canInstallUpdate() bool {
	return allowUpdateInstall && !s.demo && s.updater != nil
}
func (s *Service) CheckUpdate() (*UpdateInfo, error) {
	if s.demo {
		return nil, errors.New("模拟模式不检查更新")
	}
	if s.updater == nil {
		return nil, errors.New("更新服务未启用")
	}
	rel, err := checkRelease(s.updater, 8*time.Second)
	if err != nil {
		return nil, err
	}
	return updateInfoFrom(rel, s.canInstallUpdate()), nil
}
func (s *Service) InstallUpdate() error {
	if err := s.updateGuard("安装更新"); err != nil {
		return err
	}
	if !allowUpdateInstall {
		return errors.New("开发构建不能替换当前程序，请到 GitHub Releases 下载新版本")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return s.updater.DownloadAndInstall(ctx)
}
func (s *Service) RestartUpdate() error {
	if err := s.updateGuard("重启"); err != nil {
		return err
	}
	if !allowUpdateInstall {
		return errors.New("开发构建不能替换当前程序，请到 GitHub Releases 下载新版本")
	}
	return s.updater.Restart(context.Background())
}
func (s *Service) updateGuard(action string) error {
	if s.demo {
		return errors.New("模拟模式不能安装更新")
	}
	if s.updater == nil {
		return errors.New("更新服务未启用")
	}
	if !s.op.TryLock() {
		return errors.New("正在操作环境，请完成后再" + action)
	}
	s.op.Unlock()
	return nil
}
func (s *Service) SkipUpdate(version string) error {
	if s.updater == nil {
		return errors.New("更新服务未启用")
	}
	if version == "" {
		return errors.New("缺少版本号")
	}
	s.updater.SkipVersion(version)
	return nil
}
func (s *Service) GetCatalog() (core.Catalog, error) { return s.host.Catalog() }
func (s *Service) GetClients() (core.Clients, error) { return s.host.Discover() }
func (s *Service) SaveConfig(c core.Config) error {
	if !s.op.TryLock() {
		return errors.New("操作正在进行")
	}
	defer s.op.Unlock()
	if !s.engine.Lock() {
		return errors.New("正在读取或修改环境，请稍后保存")
	}
	defer s.engine.Unlock()
	j, err := (core.FileStore{Dir: s.dir}).Load()
	if err != nil {
		return err
	}
	old, err := s.config()
	if err != nil {
		return err
	}
	if core.Active(j) && c.Target != old.Target {
		return errors.New("恢复原环境后才能修改目标配置")
	}
	// 设置窗口只修改客户端路径并沿用已保存的目标；目标未变时无需再读取系统目录。
	if c.Target != old.Target {
		if err = s.host.Validate(c.Target); err != nil {
			return err
		}
	}
	st, err := os.Stat(c.WorkDir)
	if err != nil || !st.IsDir() {
		return errors.New("CLI 工作目录必须是存在的文件夹")
	}
	if c.CLIPath != "" {
		if err = platform.ValidateClientPath("cli", c.CLIPath); err != nil {
			return err
		}
	}
	if c.DesktopPath != "" {
		if err = platform.ValidateClientPath("desktop", c.DesktopPath); err != nil {
			return err
		}
	}
	c.Version = 1
	return core.AtomicJSON(filepath.Join(s.dir, "config.json"), c)
}
func (s *Service) Switch() error {
	if !s.op.TryLock() {
		return errors.New("操作正在进行")
	}
	defer s.op.Unlock()
	c, err := s.config()
	if err != nil {
		return err
	}
	return s.engine.Switch(c.Target)
}

// SwitchTo keeps target persistence and the transaction under one service lock.
// Client preferences are deliberately not taken from an unsaved frontend draft.
func (s *Service) SwitchTo(target core.Target) error {
	if !s.op.TryLock() {
		return errors.New("操作正在进行")
	}
	defer s.op.Unlock()
	j, err := (core.FileStore{Dir: s.dir}).Load()
	if err != nil {
		return err
	}
	if core.Active(j) {
		return errors.New("请先恢复原环境，再选择新的目标")
	}
	if err = s.host.Validate(target); err != nil {
		return err
	}
	c, err := s.config()
	if err != nil {
		return err
	}
	c.Target = target
	if err = core.AtomicJSON(filepath.Join(s.dir, "config.json"), c); err != nil {
		return err
	}
	if err = s.engine.Switch(target); err != nil {
		return errors.New("目标已保存，但环境未切换成功：" + err.Error())
	}
	return nil
}
func (s *Service) Restore(choices map[string]core.Choice) error {
	if !s.op.TryLock() {
		return errors.New("操作正在进行")
	}
	defer s.op.Unlock()
	return s.engine.Restore(choices)
}
func (s *Service) Launch(client string) error {
	if !s.op.TryLock() {
		return errors.New("操作正在进行")
	}
	defer s.op.Unlock()
	j, err := (core.FileStore{Dir: s.dir}).Load()
	if err != nil {
		return err
	}
	if core.Active(j) && j.Status != "active" {
		return errors.New("请先完成恢复，再启动 Claude")
	}
	c, err := s.config()
	if err != nil {
		return err
	}
	clients, err := s.host.Discover()
	if err != nil {
		return err
	}
	switch client {
	case "cli":
		path := c.CLIPath
		if path == "" {
			path = clients.CLI
		}
		if path == "" {
			return errors.New("未找到 Claude CLI，请在设置中指定路径")
		}
		return s.host.LaunchCLI(path, c.WorkDir)
	case "desktop":
		path, kind := clients.Desktop, clients.DesktopKind
		if c.DesktopPath != "" {
			path = c.DesktopPath
			kind = map[string]string{"windows": "exe", "darwin": "app"}[platform.Kind]
		}
		if path == "" {
			return errors.New("未找到 Claude Desktop，请在设置中指定路径")
		}
		return s.host.LaunchDesktop(path, kind)
	default:
		return errors.New("未知客户端")
	}
}
func (s *Service) ConfirmExit() error {
	if !s.op.TryLock() {
		return errors.New("正在操作，请完成后退出")
	}
	defer s.op.Unlock()
	if !s.engine.Lock() {
		return errors.New("正在检查环境，请稍候")
	}
	s.engine.Unlock()
	if s.quit != nil {
		s.quit()
	}
	return nil
}
func (s *Service) closeStatus() string {
	if !s.op.TryLock() {
		return "busy"
	}
	defer s.op.Unlock()
	active, err := s.engine.HasActive()
	if err != nil || active {
		return "confirm"
	}
	return "ok"
}
