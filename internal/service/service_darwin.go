//go:build darwin

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const plistName = ServiceName + ".plist"

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", plistName)
}

type darwinManager struct{ binPath string }

func NewManager(binPath string) Manager {
	return &darwinManager{binPath: binPath}
}

func isAdmin() (bool, string) { return true, "launchd" }

func (m *darwinManager) plist() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin</string>
    </dict>
</dict>
</plist>`, ServiceName, m.binPath,
		filepath.Join(os.TempDir(), "stxai-out.log"),
		filepath.Join(os.TempDir(), "stxai-err.log"))
}

func (m *darwinManager) Install(binaryPath, _ string) error {
	if binaryPath != "" {
		m.binPath = binaryPath
	}
	if m.binPath == "" {
		return fmt.Errorf("could not determine binary path")
	}

	// Write plist
	if err := os.MkdirAll(filepath.Dir(plistPath()), 0755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}
	if err := os.WriteFile(plistPath(), []byte(m.plist()), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	// Load
	cmd := exec.Command("launchctl", "bootout", "gui/"+os.Getenv("UID"), plistPath())
	cmd.Run() // ignore errors (may not exist)
	cmd = exec.Command("launchctl", "bootstrap", "gui/"+os.Getenv("UID"), plistPath())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %s — %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (m *darwinManager) Uninstall() error {
	cmd := exec.Command("launchctl", "bootout", "gui/"+os.Getenv("UID"), plistPath())
	cmd.Run() // ignore errors
	os.Remove(plistPath())
	return nil
}

func (m *darwinManager) Status() (*Status, error) {
	s := &Status{Label: "launchd", Path: plistPath()}
	if _, err := os.Stat(plistPath()); err == nil {
		s.Installed = true
	}
	cmd := exec.Command("launchctl", "print", "gui/"+os.Getenv("UID")+"/"+ServiceName)
	out, _ := cmd.CombinedOutput()
	s.Running = strings.Contains(string(out), "state = running")
	return s, nil
}
