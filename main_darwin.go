package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const alertScript = `on run argv
display alert "ClaudePro" message (item 1 of argv)
end run`

func message(text string) {
	_ = exec.Command("/usr/bin/osascript", "-e", alertScript, text).Run()
}

func platformCommand([]string) bool { return false }

func dataDir(demo bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	name := "ClaudePro"
	if demo {
		name += "-Demo"
	}
	return filepath.Join(home, "Library", "Application Support", name), nil
}

func windowsOptions(string, bool) application.WindowsOptions { return application.WindowsOptions{} }

func checkRuntime(*application.App) bool { return true }
