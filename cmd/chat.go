package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"

	"github.com/spf13/cobra"
)

func ChatCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chat",
		Short: "Interactive CLI chat with STX AI",
		Long:  "Open an interactive chat session with the STX AI financial agent.",
		RunE:  runChat,
	}
}

func runChat(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w (run 'stxai setup' first)", err)
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("API key not set — run 'stxai setup'")
	}

	c := client.New(cfg.APIURL, cfg.APIKey, cfg.Lang)

	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║     STX AI Chat                  ║")
	fmt.Println("║  Type /help for commands         ║")
	fmt.Println("║  Type /quit to exit              ║")
	fmt.Println("╚══════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	sessionID := fmt.Sprintf("cli:%d", os.Getpid())
	deepMode := false

	for {
		if deepMode {
			fmt.Print("🔬 > ")
		} else {
			fmt.Print("📊 > ")
		}
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "/quit", "/q", "exit":
			fmt.Println("👋 Goodbye!")
			return nil
		case "/help":
			fmt.Println("Commands:")
			fmt.Println("  /analyze <ticker>  — Full multi-agent deep analysis")
			fmt.Println("  /deep              — Toggle deep analysis mode (Bull vs Bear debate)")
			fmt.Println("  /clear             — Clear session history")
			fmt.Println("  /quit              — Exit chat")
			fmt.Println("  /help              — Show this help")
			fmt.Println()
			fmt.Println("Or just ask about any US or HK stock:")
			fmt.Println("  AAPL price? | TSLA RSI? | 港股 0700 現價？")
			if deepMode {
				fmt.Println("🔬 Deep analysis mode ON — multi-agent debate + full report")
			}
			fmt.Println()
			continue
		case "/clear":
			sessionID = fmt.Sprintf("cli:%d:%d", os.Getpid(), os.Getpid())
			fmt.Println("✅ Session cleared.")
			continue
		case "/deep":
			deepMode = !deepMode
			if deepMode {
				fmt.Println("🔬 Deep analysis mode ON — multi-agent debate + full report")
			} else {
				fmt.Println("📊 Deep analysis mode OFF — quick responses")
			}
			continue
		}

		if strings.HasPrefix(input, "/analyze ") {
			ticker := strings.TrimSpace(strings.TrimPrefix(input, "/analyze "))
			fmt.Print("🔬 Running multi-agent deep analysis... ")
			ar, err := c.Analyze(ticker)
			if err != nil {
				fmt.Printf("❌ %s\n\n", err.Error())
				continue
			}
			fmt.Println()
			fmt.Printf("\n%s\n\n", ar.Summary)
			continue
		}

		if deepMode {
			fmt.Print("🔬 ")
		} else {
			fmt.Print("🤖 ")
		}
		resp, err := c.Chat(input, sessionID, deepMode)
		if err != nil {
			fmt.Printf("❌ %s\n\n", err.Error())
			continue
		}
		fmt.Printf("\n%s\n\n", resp.Reply)
	}
	return nil
}
