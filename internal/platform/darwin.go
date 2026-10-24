//go:build darwin

package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"claudepro/internal/core"
)

// Kind 标识宿主平台，供模板与界面选择对应的系统标识。
const Kind = "darwin"

// Darwin 管理 macOS 的系统时区、当前用户的区域格式（AppleLocale）和国家 / 地区（Country）。
type Darwin struct{}

func NewHost() *Darwin { return &Darwin{} }

const (
	localtimeLink = "/etc/localtime"
	autoTimezone  = "/Library/Preferences/com.apple.timezone.auto"
)

var (
	timezonePattern = regexp.MustCompile(`^[A-Za-z0-9_+\-]+(/[A-Za-z0-9_+\-]+)*$`)
	localePattern   = regexp.MustCompile(`^[A-Za-z]{2,3}(_[A-Za-z0-9]{2,8})*$`)
	countryPattern  = regexp.MustCompile(`^[A-Z]{2}$`)
)

func command(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.Background(), func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("%v %s", err, strings.TrimSpace(stderr.String()))
	}
	return out.String(), nil
}

// readDefault 读取偏好设置；键不存在时返回 present=false 而不是错误。
func readDefault(domain, key string) (value string, present bool, err error) {
	out, err := command(10*time.Second, "/usr/bin/defaults", "read", domain, key)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return "", false, nil
		}
		return "", false, fmt.Errorf("无法读取 %s %s：%w", domain, key, err)
	}
	return strings.TrimSpace(out), true, nil
}

func readTimezone() (string, error) {
	return readTimezoneLink(func() (string, error) { return os.Readlink(localtimeLink) }, time.Sleep)
}

// macOS 更新时区时可能短暂移除链接；只重试不存在，不掩盖权限或格式错误。
func readTimezoneLink(readlink func() (string, error), sleep func(time.Duration)) (string, error) {
	var link string
	var err error
	for attempt := 0; attempt < 21; attempt++ {
		link, err = readlink()
		if !errors.Is(err, os.ErrNotExist) || attempt == 20 {
			break
		}
		sleep(100 * time.Millisecond)
	}
	if err != nil {
		return "", fmt.Errorf("无法读取系统时区：%w", err)
	}
	if i := strings.Index(link, "zoneinfo/"); i >= 0 {
		return link[i+len("zoneinfo/"):], nil
	}
	return "", errors.New("无法识别系统时区：" + link)
}

func boolText(present bool) string {
	if present {
		return "1"
	}
	return "0"
}

func (d *Darwin) Read() (core.Environment, error) {
	tz, err := readTimezone()
	if err != nil {
		return nil, err
	}
	auto, autoPresent, err := readDefault(autoTimezone, "Active")
	if err != nil {
		return nil, err
	}
	if !autoPresent {
		auto = ""
	}
	locale, localePresent, err := readDefault("-g", "AppleLocale")
	if err != nil {
		return nil, err
	}
	country, countryPresent, err := readDefault("-g", "Country")
	if err != nil {
		return nil, err
	}
	// 标签只取各自键的值：区域格式的修饰符（如 @rg=）不参与比较，国家 / 地区不从区域格式推导，
	// 否则切换区域格式会被误判为国家 / 地区遭到外部修改。
	return core.Environment{
		"timezone": {Label: tz, Data: map[string]string{"id": tz, "auto": auto}},
		"culture":  {Label: strings.SplitN(locale, "@", 2)[0], Data: map[string]string{"value": locale, "present": boolText(localePresent)}},
		"region":   {Label: country, Data: map[string]string{"value": country, "present": boolText(countryPresent)}},
	}, nil
}

// 目录通过 JavaScript for Automation 调用 Foundation，系统自带、无需额外依赖。
const catalogScript = `
ObjC.import('Foundation');
const zh = $.NSLocale.alloc.initWithLocaleIdentifier('zh_CN');
const str = v => ObjC.unwrap(v) || '';
const now = $.NSDate.date;
const offset = s => {const sign = s < 0 ? '-' : '+'; s = Math.abs(s); return 'UTC' + sign + String(Math.floor(s / 3600)).padStart(2, '0') + ':' + String(Math.floor(s % 3600 / 60)).padStart(2, '0');};
const timezones = ObjC.deepUnwrap($.NSTimeZone.knownTimeZoneNames).filter(id => id.includes('/') || id === 'GMT' || id === 'UTC')
  .map(id => {const tz = $.NSTimeZone.timeZoneWithName(id); const s = tz.secondsFromGMTForDate(now); return {s, id, name: '(' + offset(s) + ') ' + id};})
  .sort((a, b) => a.s - b.s || a.id.localeCompare(b.id)).map(({id, name}) => ({id, name}));
const cultures = ObjC.deepUnwrap($.NSLocale.availableLocaleIdentifiers).filter(id => /^[a-z]{2,3}(_[A-Z][a-z]{3})?_[A-Z]{2}$/.test(id)).sort()
  .map(id => ({id, name: (str(zh.localizedStringForLocaleIdentifier(id)) || id) + ' / ' + id}));
const regions = ObjC.deepUnwrap($.NSLocale.ISOCountryCodes)
  .map(id => ({id, name: (str(zh.localizedStringForCountryCode(id)) || id) + ' (' + id + ')'})).sort((a, b) => a.name.localeCompare(b.name, 'zh'));
JSON.stringify({timezones, cultures, regions});
`

func (d *Darwin) Catalog() (core.Catalog, error) {
	var result core.Catalog
	out, err := command(45*time.Second, "/usr/bin/osascript", "-l", "JavaScript", "-e", catalogScript)
	if err != nil {
		return result, fmt.Errorf("无法读取 macOS 区域选项：%w", err)
	}
	if err = json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		return result, fmt.Errorf("macOS 返回数据无效：%w", err)
	}
	return result, nil
}

func (d *Darwin) Validate(t core.Target) error {
	if !timezonePattern.MatchString(t.Timezone) || !localePattern.MatchString(t.Culture) || !countryPattern.MatchString(t.Region) {
		return errors.New("请选择 macOS 支持的有效设置")
	}
	c, err := d.Catalog()
	if err != nil {
		return err
	}
	for _, item := range []struct {
		v    string
		list []core.Option
	}{{t.Timezone, c.Timezones}, {t.Culture, c.Cultures}, {t.Region, c.Regions}} {
		found := false
		for _, option := range item.list {
			if option.ID == item.v {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("请选择 macOS 支持的有效设置：%s", item.v)
		}
	}
	return nil
}

// 系统时区需要管理员授权。参数经 argv 传入并用 quoted form 转义，不拼接进脚本。
// 同时关闭“自动设置时区”，避免系统按定位把时区改回去；恢复时按原值还原。
const timezoneScript = `on run argv
set tz to item 1 of argv
set auto to item 2 of argv
set cmd to "/usr/sbin/systemsetup -settimezone " & quoted form of tz & " >/dev/null"
if auto is "delete" then
set cmd to cmd & " && (/usr/bin/defaults delete /Library/Preferences/com.apple.timezone.auto Active 2>/dev/null || true)"
else
set cmd to cmd & " && /usr/bin/defaults write /Library/Preferences/com.apple.timezone.auto Active -bool " & auto
end if
do shell script cmd with prompt "ClaudePro 需要修改系统时区。" with administrator privileges
end run`

func setTimezone(id, auto string) error {
	if !timezonePattern.MatchString(id) {
		return errors.New("无效 macOS 时区")
	}
	if _, err := os.Stat(filepath.Join("/var/db/timezone/zoneinfo", id)); err != nil {
		if _, err = os.Stat(filepath.Join("/usr/share/zoneinfo", id)); err != nil {
			return errors.New("本机不存在该时区：" + id)
		}
	}
	if auto != "true" && auto != "false" && auto != "delete" {
		return errors.New("无效的自动时区设置")
	}
	if _, err := command(0, "/usr/bin/osascript", "-e", timezoneScript, id, auto); err != nil {
		if strings.Contains(err.Error(), "-128") {
			return errors.New("时区授权被取消，未修改系统时区")
		}
		return fmt.Errorf("时区修改失败：%w", err)
	}
	return waitTimezone(id, auto, func() (string, string, error) {
		tz, err := readTimezone()
		if err != nil {
			return "", "", err
		}
		active, present, err := readDefault(autoTimezone, "Active")
		if !present {
			active = ""
		}
		return tz, active, err
	}, time.Sleep)
}

// systemsetup 返回不代表系统已发布最终值；切换和恢复均须确认两项设置。
func waitTimezone(id, auto string, read func() (string, string, error), sleep func(time.Duration)) error {
	expectedAuto := map[string]string{"true": "1", "false": "0", "delete": ""}[auto]
	for attempt := 0; ; attempt++ {
		tz, active, err := read()
		if err != nil {
			return fmt.Errorf("时区修改后核验失败：%w", err)
		}
		if tz == id && active == expectedAuto {
			return nil
		}
		if attempt == 30 {
			return fmt.Errorf("时区修改后核验超时：目标 %s / 自动时区 %q，实际 %s / 自动时区 %q", id, expectedAuto, tz, active)
		}
		sleep(100 * time.Millisecond)
	}
}

func writeDefault(key, value string) error {
	_, err := command(10*time.Second, "/usr/bin/defaults", "write", "-g", key, "-string", value)
	return err
}

func (d *Darwin) Apply(field, value string) error {
	switch field {
	case "timezone":
		return setTimezone(value, "false")
	case "culture":
		if !localePattern.MatchString(value) {
			return errors.New("无效区域格式")
		}
		return writeDefault("AppleLocale", value)
	case "region":
		if !countryPattern.MatchString(value) {
			return errors.New("无效国家 / 地区")
		}
		return writeDefault("Country", value)
	default:
		return errors.New("不支持的设置项")
	}
}

func (d *Darwin) Restore(field string, value core.Value) error {
	switch field {
	case "timezone":
		auto := map[string]string{"1": "true", "0": "false", "": "delete"}[value.Data["auto"]]
		if auto == "" {
			return errors.New("原始自动时区数据无效")
		}
		return setTimezone(value.Data["id"], auto)
	case "culture", "region":
		key := map[string]string{"culture": "AppleLocale", "region": "Country"}[field]
		switch value.Data["present"] {
		case "1":
			return writeDefault(key, value.Data["value"])
		case "0":
			_, err := command(10*time.Second, "/usr/bin/defaults", "delete", "-g", key)
			if err != nil && strings.Contains(err.Error(), "does not exist") {
				return nil
			}
			return err
		default:
			return errors.New("缺少原始区域数据")
		}
	default:
		return errors.New("不支持的恢复项")
	}
}

var (
	desktopProcess = regexp.MustCompile(`/Claude\.app/Contents/MacOS/Claude(\s|$)`)
	cliScript      = regexp.MustCompile(`@anthropic-ai/claude-code|/claude-code/`)
)

// claudeProcesses 从 `ps -axww -o pid=,args=` 的输出中筛选 Claude 进程。
func claudeProcesses(ps string, self int) []string {
	result := []string{}
	for _, line := range strings.Split(ps, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil || pid == self {
			continue
		}
		args := strings.Join(fields[1:], " ")
		name := filepath.Base(fields[1])
		switch {
		case desktopProcess.MatchString(args):
			result = append(result, fmt.Sprintf("Claude Desktop (PID %d)", pid))
		case name == "claude" || name == "claude-code":
			result = append(result, fmt.Sprintf("%s (PID %d)", name, pid))
		case (name == "node" || name == "bun") && cliScript.MatchString(args):
			result = append(result, fmt.Sprintf("%s (PID %d)", name, pid))
		}
	}
	return result
}

func (d *Darwin) Running() ([]string, error) {
	out, err := command(15*time.Second, "/bin/ps", "-axww", "-o", "pid=,args=")
	if err != nil {
		return nil, fmt.Errorf("无法检查运行中的进程：%w", err)
	}
	return claudeProcesses(out, os.Getpid()), nil
}

func executable(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir() && st.Mode()&0111 != 0
}

// 图形界面启动的应用 PATH 很短，先查常见安装位置，再询问登录 Shell。
func findCLI() string {
	home, _ := os.UserHomeDir()
	for _, path := range []string{
		filepath.Join(home, ".claude", "local", "claude"),
		filepath.Join(home, ".local", "bin", "claude"),
		"/opt/homebrew/bin/claude",
		"/usr/local/bin/claude",
		filepath.Join(home, ".npm-global", "bin", "claude"),
		filepath.Join(home, ".bun", "bin", "claude"),
	} {
		if executable(path) {
			return path
		}
	}
	shell := os.Getenv("SHELL")
	if !filepath.IsAbs(shell) {
		shell = "/bin/zsh"
	}
	out, _ := command(5*time.Second, shell, "-lic", "command -v claude")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if last := strings.TrimSpace(lines[len(lines)-1]); filepath.IsAbs(last) && executable(last) {
		return last
	}
	return ""
}

func (d *Darwin) Discover() (core.Clients, error) {
	result := core.Clients{CLI: findCLI()}
	home, _ := os.UserHomeDir()
	for _, path := range []string{"/Applications/Claude.app", filepath.Join(home, "Applications", "Claude.app")} {
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			result.Desktop, result.DesktopKind = path, "app"
			break
		}
	}
	return result, nil
}

// ValidateClientPath 校验手动指定的 macOS 客户端路径：CLI 为可执行文件，Desktop 为 .app 应用包。
func ValidateClientPath(client, path string) error {
	if !filepath.IsAbs(path) {
		return errors.New("客户端路径必须是绝对路径")
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if client == "cli" {
		if st.IsDir() || st.Mode()&0111 == 0 {
			return errors.New("请选择可执行的 Claude CLI 文件")
		}
		return nil
	}
	if !st.IsDir() || !strings.EqualFold(filepath.Ext(path), ".app") {
		return errors.New("请选择 Claude.app 应用")
	}
	return nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// CLICommand 构造在“终端”中执行的命令行；路径一律按 Shell 单引号转义。
func CLICommand(path, workDir string) (string, error) {
	if err := ValidateClientPath("cli", path); err != nil {
		return "", err
	}
	info, err := os.Stat(workDir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("CLI 工作目录不存在")
	}
	return "cd " + shellQuote(workDir) + " && " + shellQuote(path), nil
}

const terminalScript = `on run argv
tell application "Terminal"
activate
do script (item 1 of argv)
end tell
end run`

func (d *Darwin) LaunchCLI(path, dir string) error {
	line, err := CLICommand(path, dir)
	if err != nil {
		return err
	}
	if _, err = command(30*time.Second, "/usr/bin/osascript", "-e", terminalScript, line); err != nil {
		return fmt.Errorf("无法打开终端（请在“系统设置 › 隐私与安全性 › 自动化”中允许 ClaudePro 控制终端）：%w", err)
	}
	return nil
}

// open -a 会激活已在运行的实例，不会重复启动。
func (d *Darwin) LaunchDesktop(path, kind string) error {
	if err := ValidateClientPath("desktop", path); err != nil {
		return err
	}
	_, err := command(30*time.Second, "/usr/bin/open", "-a", path)
	return err
}
