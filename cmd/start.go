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

	"github.com/spf13/cobra"
)

var svcFlag bool

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the STX AI Agent (TG Bot + Web UI)",
		Long:  "Launch the Telegram bot that provides AI-powered stock analysis via the STX AI Cloud API.",
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

	if cfg.Telegram == "" {
		return fmt.Errorf("telegram bot token not set — run 'stxai setup' and add a token from @BotFather")
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("API key not set — run 'stxai setup'")
	}

	s, err := store.Open()
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	c := client.New(cfg.APIURL, cfg.APIKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	runAgent := func(ctx context.Context) error {
		b, err := bot.New(cfg.Telegram, c, s)
		if err != nil {
			return fmt.Errorf("create bot: %w", err)
		}
		fmt.Println("Agent is running. Press Ctrl+C to stop.")
		return b.Start(ctx)
	}

	if svcFlag {
		return service.RunWindowsService(ctx, runAgent)
	}
	return runAgent(ctx)
}
