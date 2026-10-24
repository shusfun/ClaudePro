//go:build windows

package platform

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"claudepro/internal/core"
	"golang.org/x/sys/windows"
)

type Windows struct{}

var PowerShell = filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")

func EncodeScript(s string) string {
	chars := utf16.Encode([]rune(s))
	b := make([]byte, len(chars)*2)
	for i, c := range chars {
		binary.LittleEndian.PutUint16(b[i*2:], c)
	}
	return base64.StdEncoding.EncodeToString(b)
}
func payloadEnv(payload any) ([]string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := []string{}
	for _, s := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(s), "CLAUDEPRO_PAYLOAD=") {
			env = append(env, s)
		}
	}
	return append(env, "CLAUDEPRO_PAYLOAD="+base64.StdEncoding.EncodeToString(b)), nil
}

const prelude = `$ErrorActionPreference='Stop'
[Console]::OutputEncoding=New-Object System.Text.UTF8Encoding($false)
$p=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($env:CLAUDEPRO_PAYLOAD)) | ConvertFrom-Json
`

func run(script string, payload any, result any, timeout time.Duration) error {
	ctx := context.Background()
	cancel := func() {}
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()
	cmd := exec.CommandContext(ctx, PowerShell, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-EncodedCommand", EncodeScript(prelude+script))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	env, err := payloadEnv(payload)
	if err != nil {
		return err
	}
	cmd.Env = env
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Windows 操作失败：%v %s", err, strings.TrimSpace(stderr.String()))
	}
	if result != nil {
		if err = json.Unmarshal(bytes.TrimPrefix(bytes.TrimSpace(out.Bytes()), []byte{239, 187, 191}), result); err != nil {
			return fmt.Errorf("Windows 返回数据无效：%w", err)
		}
	}
	return nil
}

const readScript = `
function Read-Values($path) {
 $key=Get-Item -LiteralPath $path; $values=@{}
 foreach($name in $key.GetValueNames()) {
  if($key.GetValueKind($name) -ne [Microsoft.Win32.RegistryValueKind]::String) {throw "不支持的区域设置类型: $name"}
  $values[$name]=[string]$key.GetValue($name)
 }
 return $values
}
$format=Read-Values 'HKCU:\Control Panel\International'
$geo=Read-Values 'HKCU:\Control Panel\International\Geo'
$tz=(& "$env:SystemRoot\System32\tzutil.exe" /g | Out-String).Trim()
if($LASTEXITCODE -ne 0){throw '无法读取时区'}
@{timezone=@{label=$tz;data=@{id=$tz}};culture=@{label=$format.LocaleName;data=$format};region=@{label=[string]$geo.Nation;data=$geo}} | ConvertTo-Json -Depth 6 -Compress
`

func (w *Windows) Read() (core.Environment, error) {
	var env core.Environment
	err := run(readScript, nil, &env, 30*time.Second)
	return env, err
}
func (w *Windows) Catalog() (core.Catalog, error) {
	var result core.Catalog
	err := run(`
$cultures=@([Globalization.CultureInfo]::GetCultures([Globalization.CultureTypes]::SpecificCultures) | Sort-Object Name)
$regions=@{}
foreach($c in $cultures) {try {$r=New-Object Globalization.RegionInfo($c.Name); if($r.GeoId -gt 0){$regions[[string]$r.GeoId]=@{id=[string]$r.GeoId;name=$r.DisplayName+' ('+$r.TwoLetterISORegionName+')'}}} catch {}}
@{timezones=@(Get-TimeZone -ListAvailable | ForEach-Object {@{id=$_.Id;name=$_.DisplayName}});cultures=@($cultures | ForEach-Object {@{id=$_.Name;name=$_.DisplayName+' / '+$_.Name}});regions=@($regions.Values | Sort-Object name)} | ConvertTo-Json -Depth 5 -Compress
`, nil, &result, 45*time.Second)
	return result, err
}
func (w *Windows) Validate(t core.Target) error {
	c, err := w.Catalog()
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
			return fmt.Errorf("请选择 Windows 支持的有效设置：%s", item.v)
		}
	}
	return nil
}
func (w *Windows) Apply(field, value string) error {
	switch field {
	case "timezone":
		return SetTimezone(value)
	case "culture":
		return run(`Set-Culture -CultureInfo $p.value`, map[string]string{"value": value}, nil, 30*time.Second)
	case "region":
		return run(`Set-WinHomeLocation -GeoId ([int]$p.value)`, map[string]string{"value": value}, nil, 30*time.Second)
	default:
		return errors.New("不支持的设置项")
	}
}
func (w *Windows) Restore(field string, value core.Value) error {
	if field == "timezone" {
		return SetTimezone(value.Label)
	}
	if field != "culture" && field != "region" {
		return errors.New("不支持的恢复项")
	}
	if len(value.Data) == 0 {
		return errors.New("缺少原始区域数据")
	}
	return run(`
$path='HKCU:\Control Panel\International'
if($p.field -eq 'region'){$path+='\Geo'}
$subkey='Control Panel\International'
if($p.field -eq 'region'){$subkey+='\Geo'}
$key=[Microsoft.Win32.Registry]::CurrentUser.OpenSubKey($subkey,$true)
if(-not $key){throw '无法打开区域设置注册表键'}
$names=@($p.data.PSObject.Properties.Name)
foreach($name in $key.GetValueNames()){if($name -notin $names){$key.DeleteValue($name)}}
foreach($prop in $p.data.PSObject.Properties){$key.SetValue($prop.Name,[string]$prop.Value,[Microsoft.Win32.RegistryValueKind]::String)}
$key.Flush();$key.Close()
Add-Type -TypeDefinition 'using System;using System.Runtime.InteropServices;public class CPNotify{[DllImport("user32.dll",CharSet=CharSet.Unicode,SetLastError=true)]public static extern IntPtr SendMessageTimeout(IntPtr h,uint m,UIntPtr w,string l,uint f,uint t,out UIntPtr r);}'
$result=[UIntPtr]::Zero
[void][CPNotify]::SendMessageTimeout([IntPtr]0xffff,0x1a,[UIntPtr]::Zero,'intl',2,1000,[ref]$result)
`, map[string]any{"field": field, "data": value.Data}, nil, 30*time.Second)
}
func SetTimezone(id string) error {
	// This child has a fixed operation and validates the timezone again after UAC.
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return run(`
$args=@('--timezone-helper',$p.id,[string]$p.parent)
try {$proc=Start-Process -FilePath $p.exe -ArgumentList $args -Verb RunAs -WindowStyle Hidden -PassThru -Wait; if($proc.ExitCode -ne 0){throw "时区辅助进程退出码 $($proc.ExitCode)"}} catch {throw "时区授权被取消或修改失败：$($_.Exception.Message)"}
`, map[string]any{"exe": exe, "id": base64.RawURLEncoding.EncodeToString([]byte(id)), "parent": os.Getpid()}, nil, 0)
}
func TimezoneHelper(encoded string, parentPID uint32) error {
	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(b) > 256 {
		return errors.New("无效时区参数")
	}
	// Retain a handle to the original parent, preventing PID reuse from making a
	// delayed UAC approval act on behalf of a different process.
	h, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, parentPID)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	status, err := windows.WaitForSingleObject(h, 0)
	if err != nil || status != uint32(windows.WAIT_TIMEOUT) {
		return errors.New("主程序已退出，取消时区操作")
	}
	return run(`
$id=[string]$p.id;$base=$id -replace '_dstoff$',''
if(-not (Get-TimeZone -ListAvailable | Where-Object {$_.Id -ceq $base})){throw '无效 Windows 时区'}
if(-not (Get-Process -Id $p.parent -ErrorAction SilentlyContinue)){throw '主程序已退出'}
& "$env:SystemRoot\System32\tzutil.exe" /s $id
if($LASTEXITCODE -ne 0){throw '时区修改失败'}
`, map[string]any{"id": string(b), "parent": parentPID}, nil, 30*time.Second)
}
func (w *Windows) Running() ([]string, error) {
	result := []string{}
	err := run(`
$found=@(Get-CimInstance Win32_Process | Where-Object {
 $_.ProcessId -ne $p.pid -and (
  $_.Name -match '^(claude|claude-code)(\.exe)?$' -or
  ($_.Name -match '^(node|bun)(\.exe)?$' -and $_.CommandLine -match '@anthropic-ai[\\/]claude-code|[\\/]claude-code[\\/]') -or
  ($_.Name -ieq $p.self -and $_.CommandLine -match '--timezone-helper')
 )
} | ForEach-Object { $_.Name+' (PID '+$_.ProcessId+')' })
ConvertTo-Json -InputObject $found -Compress
`, map[string]any{"pid": os.Getpid(), "self": filepath.Base(os.Args[0])}, &result, 30*time.Second)
	return result, err
}
func (w *Windows) Discover() (core.Clients, error) {
	var result core.Clients
	err := run(discoverScript, nil, &result, 30*time.Second)
	return result, err
}

const discoverScript = `
$cli=Get-Command claude.exe,claude.cmd,claude.ps1 -ErrorAction SilentlyContinue | Select-Object -First 1
$desktop='';$kind=''
$pkg=Get-AppxPackage -Name 'Claude' | Select-Object -First 1
if($pkg){$manifest=Get-AppxPackageManifest -Package $pkg.PackageFullName;$a=$manifest.Package.Applications.Application | Select-Object -First 1;if($a){$desktop=$pkg.PackageFamilyName+'!'+$a.Id;$kind='msix'}}
if(-not $desktop){foreach($path in @("$env:LOCALAPPDATA\AnthropicClaude\Claude.exe","$env:LOCALAPPDATA\AnthropicClaude\claude.exe","$env:LOCALAPPDATA\Programs\Claude\Claude.exe")){if(Test-Path -LiteralPath $path){$desktop=$path;$kind='exe';break}}}
@{cli=$(if($cli){$cli.Source}else{''});desktop=$desktop;desktopKind=$kind} | ConvertTo-Json -Compress
`

func ValidateLaunchPath(path string, extensions ...string) error {
	if !filepath.IsAbs(path) {
		return errors.New("客户端路径必须是绝对路径")
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return errors.New("请选择客户端文件")
	}
	for _, ext := range extensions {
		if strings.EqualFold(filepath.Ext(path), ext) {
			return nil
		}
	}
	return errors.New("不支持的客户端文件类型")
}
func CLICommand(path, workDir string) (*exec.Cmd, error) {
	if err := ValidateLaunchPath(path, ".exe", ".cmd", ".ps1"); err != nil {
		return nil, err
	}
	info, err := os.Stat(workDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("CLI 工作目录不存在")
	}
	// Both paths are transported as JSON in a private environment variable.
	// PowerShell's invocation operator handles spaces, Unicode and shell syntax.
	script := prelude + `Set-Location -LiteralPath $p.workDir; $launch=$p.path;Remove-Item Env:CLAUDEPRO_PAYLOAD; & $launch`
	cmd := exec.Command(PowerShell, "-NoLogo", "-NoProfile", "-NoExit", "-ExecutionPolicy", "Bypass", "-EncodedCommand", EncodeScript(script))
	cmd.Dir = workDir
	cmd.Env, err = payloadEnv(map[string]string{"path": path, "workDir": workDir})
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE}
	return cmd, err
}
func (w *Windows) LaunchCLI(path, dir string) error {
	cmd, err := CLICommand(path, dir)
	if err != nil {
		return err
	}
	process, err := startConsoleProcess(cmd, windows.SW_SHOWNORMAL)
	if err == nil {
		windows.CloseHandle(process)
	}
	return err
}

// os/exec connects nil standard streams to NUL and sets STARTF_USESTDHANDLES.
// A GUI launching an interactive terminal must instead let Windows attach the
// child's standard streams to its newly allocated console.
func startConsoleProcess(cmd *exec.Cmd, show uint16) (windows.Handle, error) {
	app, err := windows.UTF16PtrFromString(cmd.Path)
	if err != nil {
		return 0, err
	}
	line, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(cmd.Args))
	if err != nil {
		return 0, err
	}
	dir, err := windows.UTF16PtrFromString(cmd.Dir)
	if err != nil {
		return 0, err
	}
	env := append([]string{}, cmd.Env...)
	sort.Slice(env, func(i, j int) bool { return strings.ToUpper(env[i]) < strings.ToUpper(env[j]) })
	block := utf16.Encode([]rune(strings.Join(env, "\x00") + "\x00\x00"))
	si := windows.StartupInfo{Flags: windows.STARTF_USESHOWWINDOW, ShowWindow: show}
	si.Cb = uint32(unsafe.Sizeof(si))
	var pi windows.ProcessInformation
	err = windows.CreateProcess(app, line, nil, nil, false, windows.CREATE_NEW_CONSOLE|windows.CREATE_UNICODE_ENVIRONMENT, &block[0], dir, &si, &pi)
	if err != nil {
		return 0, err
	}
	windows.CloseHandle(pi.Thread)
	return pi.Process, nil
}
func (w *Windows) LaunchDesktop(path, kind string) error {
	if kind == "msix" {
		// Use IApplicationActivationManager via Shell's AppsFolder activation.
		if strings.ContainsAny(path, "\\/:\"\r\n ") || !strings.Contains(path, "!") {
			return errors.New("无效 MSIX 应用标识")
		}
		verb, _ := windows.UTF16PtrFromString("open")
		target, _ := windows.UTF16PtrFromString("shell:AppsFolder\\" + path)
		return windows.ShellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL)
	}
	if err := ValidateLaunchPath(path, ".exe"); err != nil {
		return err
	}
	var running bool
	if err := run(activateDesktopScript, map[string]string{"path": path}, &running, 30*time.Second); err != nil {
		return err
	}
	if running {
		return nil
	}
	cmd := exec.Command(path)
	cmd.Dir = filepath.Dir(path)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}

const activateDesktopScript = `
$existing=@(Get-Process | Where-Object { try {$_.Path -ieq $p.path} catch {$false} })
if($existing.Count -eq 0){ConvertTo-Json $false;return}
$window=$existing | Where-Object {$_.MainWindowHandle -ne 0} | Select-Object -First 1
if(-not $window){throw 'Claude Desktop 已在运行但窗口不可见，请通过系统托盘打开。'}
Add-Type -TypeDefinition 'using System;using System.Runtime.InteropServices;public class CPActivate{[DllImport("user32.dll")]public static extern bool ShowWindowAsync(IntPtr h,int n);[DllImport("user32.dll")]public static extern bool SetForegroundWindow(IntPtr h);}'
[void][CPActivate]::ShowWindowAsync($window.MainWindowHandle,9)
[void][CPActivate]::SetForegroundWindow($window.MainWindowHandle)
ConvertTo-Json $true
`
