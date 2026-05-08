package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/garysze77/stxai/agent/internal/client"
	"github.com/garysze77/stxai/agent/internal/config"

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

	c := client.New(cfg.APIURL, cfg.APIKey)

	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║     STX AI Chat                  ║")
	fmt.Println("║  Type /help for commands         ║")
	fmt.Println("║  Type /quit to exit              ║")
	fmt.Println("╚══════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	sessionID := fmt.Sprintf("cli:%d", os.Getpid())

	for {
		fmt.Print("📊 > ")
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
			fmt.Println("  /analyze <ticker>  — Quick stock analysis")
			fmt.Println("  /quit              — Exit chat")
			fmt.Println("  /help              — Show this help")
			fmt.Println()
			fmt.Println("Or just ask about any US or HK stock:")
			fmt.Println("  AAPL price? | TSLA RSI? | 港股 0700 現價？")
			fmt.Println()
			continue
		case "/clear":
			sessionID = fmt.Sprintf("cli:%d:%d", os.Getpid(), os.Getpid())
			fmt.Println("✅ Session cleared.")
			continue
		}

		if strings.HasPrefix(input, "/analyze ") {
			ticker := strings.TrimSpace(strings.TrimPrefix(input, "/analyze "))
			ar, err := c.Analyze(ticker)
			if err != nil {
				fmt.Printf("❌ %s\n\n", err.Error())
				continue
			}
			fmt.Printf("\n📈 %s  $%.2f (%.2f%%)\n", ar.Ticker, ar.Price, ar.Change)
			fmt.Printf("RSI: %.1f  |  MACD: %.2f\n", ar.RSI, ar.MACD)
			fmt.Printf("MA50: $%.2f  |  MA200: $%.2f\n", ar.MA50, ar.MA200)
			fmt.Printf("Bollinger: $%.2f — $%.2f\n\n", ar.BollLower, ar.BollUpper)
			fmt.Printf("%s\n\n", ar.Summary)
			continue
		}

		fmt.Print("🤖 ")
		resp, err := c.Chat(input, sessionID)
		if err != nil {
			fmt.Printf("❌ %s\n\n", err.Error())
			continue
		}
		fmt.Printf("\n%s\n\n", resp.Reply)
	}
	return nil
}
