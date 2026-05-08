//go:build linux

package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const unitName = ServiceName + ".service"

func unitPath() string {
	home, _ := os.UserHomeDir()
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "systemd", "user", unitName)
}

type linuxManager struct{ binPath string }

func NewManager(binPath string) Manager {
	return &linuxManager{binPath: binPath}
}

func isAdmin() (bool, string) { return true, "systemd --user" }

func (m *linuxManager) unit() string {
	return fmt.Sprintf(`[Unit]
Description=%s
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s start
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment=PATH=/usr/local/bin:/usr/bin:/bin:%s/.local/bin

[Install]
WantedBy=default.target
`, Description, m.binPath, os.Getenv("HOME"))
}

func (m *linuxManager) runSystemctl(args ...string) error {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	cmd.Stdout = os.Stdout
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s: %s — %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (m *linuxManager) enableLinger() error {
	// Allow user services to start at boot (requires root or system-wide config)
	cmd := exec.Command("loginctl", "enable-linger")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("enable-linger: %s — %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (m *linuxManager) Install(binaryPath, _ string) error {
	if binaryPath != "" {
		m.binPath = binaryPath
	}
	if m.binPath == "" {
		return fmt.Errorf("could not determine binary path")
	}

	// Write unit file
	dir := filepath.Dir(unitPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create systemd user dir: %w", err)
	}
	if err := os.WriteFile(unitPath(), []byte(m.unit()), 0644); err != nil {
		return fmt.Errorf("write unit file: %w", err)
	}

	// Reload, enable, start
	m.runSystemctl("daemon-reload")
	if err := m.runSystemctl("enable", unitName); err != nil {
		return err
	}
	if err := m.runSystemctl("start", unitName); err != nil {
		return err
	}

	// Try to enable linger (non-fatal)
	m.enableLinger()

	return nil
}

func (m *linuxManager) Uninstall() error {
	m.runSystemctl("stop", unitName)
	m.runSystemctl("disable", unitName)
	os.Remove(unitPath())
	m.runSystemctl("daemon-reload")
	return nil
}

func (m *linuxManager) Status() (*Status, error) {
	s := &Status{Label: "systemd --user", Path: unitPath()}
	if _, err := os.Stat(unitPath()); err == nil {
		s.Installed = true
	}
	cmd := exec.Command("systemctl", "--user", "is-active", unitName)
	out, _ := cmd.CombinedOutput()
	s.Running = strings.TrimSpace(string(out)) == "active"
	return s, nil
}
