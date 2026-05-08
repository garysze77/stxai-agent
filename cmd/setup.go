package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/garysze77/stxai-agent/internal/config"

	"github.com/spf13/cobra"
)

func SetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure STX AI Agent with API key and settings",
		Long: `Setup guides you through configuring your STX AI Agent.
You'll need an API key from https://stxai.vercel.app/dashboard`,
		RunE: runSetup,
	}
}

func runSetup(_ *cobra.Command, _ []string) error {
	reader := bufio.NewReader(os.Stdin)

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{
			APIURL:     "https://api.stxai.app/api/v1",
			Model:      "stxai-agent",
			MaxHistory: 20,
		}
	}

	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║     STX AI Agent Setup           ║")
	fmt.Println("╚══════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Get your API key at: https://stxai.vercel.app/dashboard")
	fmt.Println()

	// API Key
	fmt.Printf("API Key [%s]: ", maskKey(cfg.APIKey))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		cfg.APIKey = input
	}

	// API URL
	fmt.Printf("API URL [%s]: ", cfg.APIURL)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		cfg.APIURL = input
	}

	// Telegram Bot Token (optional)
	fmt.Printf("Telegram Bot Token (optional, from @BotFather) [%s]: ", maskToken(cfg.Telegram))
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		cfg.Telegram = input
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Configuration saved!")
	fmt.Println("   Run 'stxai start' to launch your Telegram bot")
	fmt.Println("   Run 'stxai chat' for interactive CLI mode")
	return nil
}

func maskKey(k string) string {
	if len(k) <= 12 {
		return k
	}
	return k[:12] + "..."
}

func maskToken(t string) string {
	if t == "" {
		return "(not set)"
	}
	if len(t) <= 10 {
		return t
	}
	return t[:10] + "..."
}
