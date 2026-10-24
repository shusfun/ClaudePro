//go:build windows

package platform

import (
	"claudepro/internal/core"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIPathsAreDataNotExecutableScript(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "中文 工作目录 & ' 引号")
	os.MkdirAll(dir, 0700)
	for _, ext := range []string{".exe", ".cmd", ".ps1"} {
		path := filepath.Join(dir, "claude $x; test"+ext)
		os.WriteFile(path, []byte{}, 0600)
		cmd, err := CLICommand(path, dir)
		if err != nil {
			t.Fatal(err)
		}
		if cmd.Dir != dir {
			t.Fatal("wrong directory")
		}
		if strings.Contains(strings.Join(cmd.Args, " "), path) {
			t.Fatal("path interpolated into script")
		}
		found := false
		for _, v := range cmd.Env {
			if strings.HasPrefix(v, "CLAUDEPRO_PAYLOAD=") {
				b, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(v, "CLAUDEPRO_PAYLOAD="))
				var data map[string]string
				json.Unmarshal(b, &data)
				if data["path"] != path || data["workDir"] != dir {
					t.Fatal(data)
				}
				found = true
			}
		}
		if !found {
			t.Fatal("missing payload")
		}
	}
}

func TestPowerShellTerminalHasInteractiveStreams(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("GitHub-hosted Windows runners do not keep a hidden PowerShell working directory with & in the path")
	}
	dir := filepath.Join(t.TempDir(), "中文 工作目录 & 引号")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "模拟 claude.ps1")
	// A harmless CLI stand-in checks the real terminal handles and working dir.
	const script = `
Add-Type -TypeDefinition 'using System;using System.Runtime.InteropServices;public class CPConsole{[DllImport("kernel32.dll")]static extern IntPtr GetStdHandle(int n);[DllImport("kernel32.dll")]static extern bool GetConsoleMode(IntPtr h,out uint m);public static bool IsConsole(int n){uint m;return GetConsoleMode(GetStdHandle(n),out m);}}'
@{input=[CPConsole]::IsConsole(-10);output=[CPConsole]::IsConsole(-11);directory=(Get-Location).Path} | ConvertTo-Json | Set-Content -LiteralPath 'console-result.json' -Encoding UTF8
`
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	cmd, err := CLICommand(path, dir)
	if err != nil {
		t.Fatal(err)
	}
	args := []string{}
	for _, arg := range cmd.Args {
		if arg != "-NoExit" {
			args = append(args, arg)
		}
	}
	cmd.Args = args
	h, err := startConsoleProcess(cmd, windows.SW_HIDE)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	status, err := windows.WaitForSingleObject(h, 15000)
	if err != nil || status != windows.WAIT_OBJECT_0 {
		windows.TerminateProcess(h, 1)
		t.Fatal("test terminal did not exit", status, err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "console-result.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Input     bool
		Output    bool
		Directory string
	}
	if err = json.Unmarshal([]byte(strings.TrimPrefix(string(b), "\ufeff")), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Input || !result.Output || result.Directory != dir {
		t.Fatal("terminal is not interactive", result)
	}
}

func TestDesktopDiscoveryWithWindowsCommandFixtures(t *testing.T) {
	const fixtures = `
function Get-Command { if($p.cli){[pscustomobject]@{Source=$p.cli}} }
function Get-AppxPackage { if($p.msix){[pscustomobject]@{PackageFullName='Fixture';PackageFamilyName='Claude_fixture'}} }
function Get-AppxPackageManifest { [xml]'<Package><Applications><Application Id="DesktopEntry" /></Applications></Package>' }
function Test-Path { param($LiteralPath) $LiteralPath -eq $p.desktop }
$env:LOCALAPPDATA=$p.root
`
	for _, mode := range []string{"msix", "exe", "missing"} {
		t.Run(mode, func(t *testing.T) {
			root := `C:\Users\中文 用户\AppData\Local`
			payload := map[string]any{"cli": `C:\中文 npm\claude.cmd`, "root": root, "msix": mode == "msix", "desktop": ""}
			if mode != "missing" {
				payload["desktop"] = root + `\AnthropicClaude\Claude.exe`
			}
			var got core.Clients
			if err := run(fixtures+discoverScript, payload, &got, 30*time.Second); err != nil {
				t.Fatal(err)
			}
			if got.CLI != payload["cli"] {
				t.Fatal(got)
			}
			switch mode {
			case "msix":
				if got.DesktopKind != "msix" || got.Desktop != "Claude_fixture!DesktopEntry" {
					t.Fatal(got)
				}
			case "exe":
				if got.DesktopKind != "exe" || got.Desktop != payload["desktop"] {
					t.Fatal(got)
				}
			case "missing":
				if got.Desktop != "" || got.DesktopKind != "" {
					t.Fatal(got)
				}
			}
		})
	}
}

func TestDesktopActivationDoesNotStartAnotherProcess(t *testing.T) {
	const fixtures = `
Add-Type -TypeDefinition 'using System;public class CPActivate{public static bool ShowWindowAsync(IntPtr h,int n){return true;}public static bool SetForegroundWindow(IntPtr h){return true;}}'
function Add-Type {}
function Get-Process {if($p.exists){[pscustomobject]@{Path=$p.path;MainWindowHandle=[IntPtr]$p.window}}}
function Start-Process {throw 'unexpected process launch'}
`
	for _, tc := range []struct {
		exists  bool
		window  int
		wantErr bool
	}{{false, 0, false}, {true, 123, false}, {true, 0, true}} {
		var running bool
		err := run(fixtures+activateDesktopScript, map[string]any{"path": `C:\中文 目录\Claude.exe`, "exists": tc.exists, "window": tc.window}, &running, 30*time.Second)
		if (err != nil) != tc.wantErr || (err == nil && running != tc.exists) {
			t.Fatal(tc, running, err)
		}
	}
}
func TestMissingAndUnsupportedClients(t *testing.T) {
	if _, err := CLICommand(filepath.Join(t.TempDir(), "absent.exe"), t.TempDir()); err == nil {
		t.Fatal("missing accepted")
	}
	path := filepath.Join(t.TempDir(), "bad.txt")
	os.WriteFile(path, []byte{}, 0600)
	if _, err := CLICommand(path, t.TempDir()); err == nil {
		t.Fatal("txt accepted")
	}
	if err := (&Windows{}).LaunchDesktop("bad/!injection", "msix"); err == nil {
		t.Fatal("bad identity accepted")
	}
}
func TestDemoDoesNotNeedWindows(t *testing.T) {
	d := NewDemo(t.TempDir())
	before, err := d.Read()
	if err != nil {
		t.Fatal(err)
	}
	d.Apply("culture", "en-US")
	d.Restore("culture", before["culture"])
	after, _ := d.Read()
	if after["culture"].Data["sShortDate"] != "yyyy/MM/dd" {
		t.Fatal(after)
	}
}
