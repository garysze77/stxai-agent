package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/garysze77/stxai-agent/internal/bot"
	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/service"
	"github.com/garysze77/stxai-agent/internal/store"
	"github.com/garysze77/stxai-agent/internal/web"

	"github.com/spf13/cobra"
)

var svcFlag bool

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the STX AI Agent (TG Bot + Web UI)",
		Long:  "Launch the Telegram bot and local Web UI for AI-powered stock analysis via the STX AI Cloud API.",
		RunE:  runStart,
	}
	cmd.Flags().BoolVar(&svcFlag, "svc", false, "Run as a Windows service")
	cmd.Flags().MarkHidden("svc")
	return cmd
}

func runStart(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w (run 'stxai setup' first)", err)
	}

	s, err := store.Open()
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	c := client.New(cfg.APIURL, cfg.APIKey, cfg.Lang)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	// Start web UI server
	ws := web.New(c, cfg, "8080")
	go func() {
		if err := ws.Start(); err != nil {
			log.Printf("Web UI: %v", err)
		}
	}()
	web.OpenBrowser(ws.Addr())

	runAgent := func(ctx context.Context) error {
		hasTG := cfg.Telegram != ""
		if hasTG {
			b, err := bot.New(cfg.Telegram, c, s, cfg.Lang)
			if err != nil {
				return fmt.Errorf("create bot: %w", err)
			}
			fmt.Println("Agent is running. Press Ctrl+C to stop.")
			fmt.Printf("🌐 Web UI → %s\n", ws.Addr())
			return b.Start(ctx)
		}

		// No Telegram token — just keep web UI alive
		fmt.Println("Web UI only mode (no Telegram bot configured). Press Ctrl+C to stop.")
		fmt.Printf("🌐 Web UI → %s\n", ws.Addr())
		<-ctx.Done()
		return nil
	}

	if svcFlag {
		return service.RunWindowsService(ctx, runAgent)
	}
	return runAgent(ctx)
}
