//go:build windows

package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type windowsManager struct{ binPath string }

func NewManager(binPath string) Manager {
	return &windowsManager{binPath: binPath}
}

func isAdmin() (bool, string) {
	// Check if running with elevated privileges on Windows
	_, err := exec.LookPath("sc.exe")
	return err == nil, "Windows Service (SCM)"
}

func (m *windowsManager) Install(binaryPath, _ string) error {
	if binaryPath != "" {
		m.binPath = binaryPath
	}
	if m.binPath == "" {
		return fmt.Errorf("could not determine binary path")
	}

	// First, stop and remove if exists
	m.Uninstall()

	// Create the service: sc create <name> binPath= <path> start
	binPath := fmt.Sprintf("%s start", m.binPath)
	cmd := exec.Command("sc", "create", ServiceName,
		"binPath=", binPath,
		"start=", "auto",
		"DisplayName=", Name,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sc create: %s — %w", strings.TrimSpace(string(out)), err)
	}

	// Set description
	descCmd := exec.Command("sc", "description", ServiceName, Description)
	descCmd.Run()

	// Set failure action: restart on failure
	failCmd := exec.Command("sc", "failure", ServiceName,
		"actions=", "restart/10000/restart/30000/restart/60000",
		"reset=", "86400",
	)
	failCmd.Run()

	// Start the service
	startCmd := exec.Command("sc", "start", ServiceName)
	if out, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sc start: %s — %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

func (m *windowsManager) Uninstall() error {
	// Stop
	stopCmd := exec.Command("sc", "stop", ServiceName)
	stopCmd.Run()

	// Delete
	delCmd := exec.Command("sc", "delete", ServiceName)
	if out, err := delCmd.CombinedOutput(); err != nil {
		// "service does not exist" is OK
		if !strings.Contains(string(out), "1060") && !strings.Contains(string(out), "does not exist") {
			return fmt.Errorf("sc delete: %s — %w", strings.TrimSpace(string(out)), err)
		}
	}
	return nil
}

func (m *windowsManager) Status() (*Status, error) {
	s := &Status{Label: "Windows Service (SCM)"}

	cmd := exec.Command("sc", "query", ServiceName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return s, nil
	}

	output := string(out)
	s.Installed = strings.Contains(output, "SERVICE_NAME")
	s.Running = strings.Contains(output, "RUNNING")
	s.Path = fmt.Sprintf("sc query %s", ServiceName)

	return s, nil
}

// IsWindowsService returns true if the current process is running as a Windows service.
// This is detected by checking if the parent process is services.exe.
func IsWindowsService() bool {
	// Check if we were started with the special --svc flag
	for _, arg := range os.Args[1:] {
		if arg == "--svc" {
			return true
		}
	}
	return false
}
