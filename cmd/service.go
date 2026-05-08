package cmd

import (
	"fmt"

	"github.com/garysze77/stxai-agent/internal/service"

	"github.com/spf13/cobra"
)

func ServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage background service (macOS/launchd, Linux/systemd, Windows/SCM)",
		Long:  "Install, uninstall, or check the status of the STX AI background service.",
	}

	cmd.AddCommand(
		svcInstallCmd(),
		svcUninstallCmd(),
		svcStatusCmd(),
	)
	return cmd
}

func svcInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install and start the background service",
		Long: `Install STX AI as a background service that starts on boot.

macOS  → launchd plist at ~/Library/LaunchAgents/
Linux  → systemd user service at ~/.config/systemd/user/
Windows → Windows Service via SCM (requires Administrator)`,
		RunE: func(_ *cobra.Command, _ []string) error {
			bin, err := service.BinaryPath()
			if err != nil {
				return fmt.Errorf("get binary path: %w", err)
			}

			mgr := service.NewManager(bin)
			if err := mgr.Install("", ""); err != nil {
				return err
			}

			s, _ := mgr.Status()
			if s != nil && s.Running {
				fmt.Printf("✅ %s installed and running.\n", s.Label)
				if s.Path != "" {
					fmt.Printf("   Config: %s\n", s.Path)
				}
			} else {
				fmt.Println("⚠️  Service installed but may not be running. Check 'stxai service status'.")
			}
			return nil
		},
	}
}

func svcUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop and remove the background service",
		RunE: func(_ *cobra.Command, _ []string) error {
			bin, _ := service.BinaryPath()
			mgr := service.NewManager(bin)
			if err := mgr.Uninstall(); err != nil {
				return err
			}
			fmt.Println("✅ Service uninstalled.")
			return nil
		},
	}
}

func svcStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check if the service is installed and running",
		RunE: func(_ *cobra.Command, _ []string) error {
			bin, _ := service.BinaryPath()
			mgr := service.NewManager(bin)
			s, err := mgr.Status()
			if err != nil {
				return err
			}

			fmt.Printf("Service:        %s\n", service.Name)
			fmt.Printf("Platform:       %s\n", s.Label)

			if s.Installed {
				fmt.Println("Installed:      ✅ Yes")
			} else {
				fmt.Println("Installed:      ❌ No")
			}

			if s.Running {
				fmt.Println("Running:        ✅ Yes")
			} else {
				fmt.Println("Running:        ❌ No")
			}

			if s.Path != "" {
				fmt.Printf("Config path:    %s\n", s.Path)
			}

			if !s.Installed {
				fmt.Println("\nRun 'stxai service install' to install the background service.")
			}

			return nil
		},
	}
}
