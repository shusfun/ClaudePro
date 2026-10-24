//go:build darwin

package platform

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClaudeProcessFilter(t *testing.T) {
	ps := `  101 /Applications/Claude.app/Contents/MacOS/Claude
  102 /Applications/Claude.app/Contents/Frameworks/Claude Helper.app/Contents/MacOS/Claude Helper --type=renderer
  103 /opt/homebrew/bin/node /opt/homebrew/lib/node_modules/@anthropic-ai/claude-code/cli.js
  104 /Users/测试/.local/bin/claude --resume
  105 /usr/bin/vim claude.md
  106 /Applications/ClaudePro.app/Contents/MacOS/claudepro`
	got := claudeProcesses(ps, 106)
	if len(got) != 3 || !strings.Contains(got[0], "101") || !strings.Contains(got[1], "103") || !strings.Contains(got[2], "104") {
		t.Fatal(got)
	}
}

func TestProxyFlagsOnly(t *testing.T) {
	manual, pac, _ := parseProxy("<dictionary> {\n  HTTPEnable : 0\n  HTTPSEnable : 1\n  HTTPSProxy : 127.0.0.1\n  ProxyAutoConfigEnable : 0\n}")
	if !manual || pac {
		t.Fatal(manual, pac)
	}
}

func TestCLICommandQuotesPaths(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "中文 目录 & ' $x")
	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, "claude'; rm -rf ~")
	os.WriteFile(path, []byte("#!/bin/sh\n"), 0700)
	line, err := CLICommand(path, dir)
	if err != nil {
		t.Fatal(err)
	}
	if line != "cd "+shellQuote(dir)+" && "+shellQuote(path) || !strings.Contains(line, `'\''`) {
		t.Fatal(line)
	}
	os.Chmod(path, 0600)
	if _, err = CLICommand(path, dir); err == nil {
		t.Fatal("non-executable accepted")
	}
	if ValidateClientPath("desktop", dir) == nil {
		t.Fatal("non-.app accepted")
	}
}

func TestRejectsUnsafeTargets(t *testing.T) {
	d := NewHost()
	for _, v := range []string{"../../etc/passwd", "Asia/Tokyo; rm", ""} {
		if d.Apply("timezone", v) == nil {
			t.Fatal("accepted", v)
		}
	}
	if d.Apply("culture", "en_US\n") == nil || d.Apply("region", "us") == nil {
		t.Fatal("bad locale accepted")
	}
}

// 不调用 systemsetup/defaults 写入真实系统；模拟链接重建窗口及异步发布。
func TestReadTimezoneLinkTransientMissing(t *testing.T) {
	calls := 0
	got, err := readTimezoneLink(func() (string, error) {
		calls++
		if calls < 4 {
			return "", &os.PathError{Op: "readlink", Path: "/etc/localtime", Err: os.ErrNotExist}
		}
		return "/var/db/timezone/zoneinfo/America/Los_Angeles", nil
	}, func(time.Duration) {})
	if err != nil || got != "America/Los_Angeles" || calls != 4 {
		t.Fatalf("%q %v calls=%d", got, err, calls)
	}
}

func TestReadTimezoneLinkFailures(t *testing.T) {
	for _, tc := range []struct {
		name, link string
		err        error
		calls      int
	}{
		{"missing", "", os.ErrNotExist, 21},
		{"permission", "", os.ErrPermission, 1},
		{"invalid", "/unexpected/path", nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			_, err := readTimezoneLink(func() (string, error) { calls++; return tc.link, tc.err }, func(time.Duration) {})
			if err == nil || calls != tc.calls {
				t.Fatalf("err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestWaitTimezone(t *testing.T) {
	for _, tc := range []struct {
		name, auto, active string
		stale, autoStale   bool
		readErr            error
		wantErr            bool
		calls              int
	}{
		{name: "switch", auto: "false", active: "0", stale: true, autoStale: true, calls: 3},
		{name: "restore enabled", auto: "true", active: "1", calls: 1},
		{name: "restore absent", auto: "delete", active: "", calls: 1},
		{name: "wrong timezone", auto: "false", active: "0", stale: true, wantErr: true, calls: 31},
		{name: "wrong auto", auto: "false", active: "1", wantErr: true, calls: 31},
		{name: "read error", auto: "false", readErr: os.ErrPermission, wantErr: true, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			err := waitTimezone("America/Los_Angeles", tc.auto, func() (string, string, error) {
				calls++
				tz, active := "America/Los_Angeles", tc.active
				if tc.stale && (tc.wantErr || calls == 1) {
					tz = "Asia/Shanghai"
				}
				if tc.autoStale && calls < 3 {
					active = "1"
				}
				return tz, active, tc.readErr
			}, func(time.Duration) {})
			if (err != nil) != tc.wantErr || calls != tc.calls {
				t.Fatalf("err=%v calls=%d", err, calls)
			}
			if tc.readErr != nil && !errors.Is(err, tc.readErr) {
				t.Fatalf("lost cause: %v", err)
			}
		})
	}
}
