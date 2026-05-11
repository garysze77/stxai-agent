package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/web"

	"github.com/spf13/cobra"
)

func SetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Open web UI to configure API key and settings",
		Long: `Setup opens a browser window where you can configure your STX AI Agent.
You'll need an API key from https://stxai.app/dashboard`,
		RunE: runSetup,
	}
}

func runSetup(_ *cobra.Command, _ []string) error {
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{
			APIURL:     "https://api.stxai.app/api/v1",
			Model:      "stxai-agent",
			MaxHistory: 20,
		}
	}

	c := client.New(cfg.APIURL, cfg.APIKey, cfg.Lang)

	ws := web.New(c, cfg, "8080")
	go func() {
		if err := ws.Start(); err != nil {
			log.Printf("Web UI: %v", err)
		}
	}()
	web.OpenBrowser(ws.Addr())

	fmt.Println()
	fmt.Println("🌐 Web setup opened in your browser.")
	fmt.Println("   Fill in your API key and optional Telegram bot token.")
	fmt.Println("   Press Ctrl+C when done.")
	fmt.Println()
	fmt.Printf("   Get your API key at: https://stxai.app/dashboard\n")
	fmt.Println()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println()
	fmt.Println("✅ Setup complete!")
	fmt.Println("   Run 'stxai start' to launch the full agent (Telegram bot + Web UI)")
	fmt.Println("   Run 'stxai chat' for interactive CLI mode")
	return nil
}
