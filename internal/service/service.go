package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	Name        = "STX AI Agent"
	ServiceName = "com.stxai.agent"
	Description = "Autonomous Financial AI Agent — Telegram Bot + Local Web UI"
)

type Status struct {
	Installed bool
	Running   bool
	Label     string // e.g., "launchd", "systemd", "Windows Service"
	Path      string // config file path
}

type Manager interface {
	Install(binaryPath, configDir string) error
	Uninstall() error
	Status() (*Status, error)
}

func BinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// StartCmd returns the command to start the agent (used by service definitions).
func StartCmd() string {
	bin, _ := BinaryPath()
	if bin == "" {
		bin = "stxai"
	}
	return fmt.Sprintf("%s start", bin)
}

// IsAdmin checks if we have permission to manage services.
func IsAdmin() (bool, string) {
	return isAdmin()
}

// FindBinary looks for stxai in PATH or next to current binary.
func FindBinary() string {
	exe, _ := BinaryPath()
	if exe != "" && filepath.Base(exe) == "stxai" {
		if fi, err := os.Stat(exe); err == nil && !fi.IsDir() {
			return exe
		}
	}
	// fall back to PATH lookup
	p, err := exec.LookPath("stxai")
	if err == nil {
		return p
	}
	return exe
}
