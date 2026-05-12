package cmd

import (
	"fmt"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/spf13/cobra"
)

func PairCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pair",
		Short: "Generate a Telegram pairing code",
		Long:  "Generate a one-time 6-digit pairing code. Send this code to your STX AI Telegram bot with /pair <code> to link your Telegram account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'stxai setup' first)", err)
			}
			if cfg.APIKey == "" {
				return fmt.Errorf("no API key configured — run: stxai setup")
			}

			c := client.New(cfg.APIURL, cfg.APIKey, cfg.Lang)
			code, err := c.PairGenerate()
			if err != nil {
				return fmt.Errorf("generate pairing code failed: %w", err)
			}

			fmt.Printf("🔗 Your pairing code: %s\n", code)
			fmt.Println()
			fmt.Println("Send this to your STX AI bot on Telegram:")
			fmt.Printf("  /pair %s\n", code)
			fmt.Println()
			fmt.Println("Code expires in 10 minutes.")
			return nil
		},
	}
}
