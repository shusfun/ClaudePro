//go:build darwin

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
