package main

import (
	"fmt"
	"os"

	"github.com/garysze77/stxai-agent/internal/commands"
	"github.com/garysze77/stxai-agent/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "price":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: stx price <TICKER>")
			os.Exit(1)
		}
		if err := commands.PriceCommand(cfg, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: stx analyze <TICKER> [--fast]")
			os.Exit(1)
		}
		ticker := os.Args[2]
		fast := false
		for _, a := range os.Args[3:] {
			if a == "--fast" {
				fast = true
			}
		}
		if cfg.APIKey == "" {
			fmt.Fprintln(os.Stderr, "Error: STX_API_KEY not set. Run 'stx configure' first.")
			os.Exit(1)
		}
		if err := commands.AnalyzeCommand(cfg, ticker, fast); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "configure":
		if err := commands.ConfigureCommand(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`STX AI — Financial AI Agent CLI

Usage:
  stx price <TICKER>          Quick price lookup
  stx analyze <TICKER>         Streaming deep analysis
  stx analyze <TICKER> --fast  Fast mode (45s timeout)
  stx configure                Set API URL and API key

Examples:
  stx price AAPL
  stx analyze TSLA
  stx analyze 0700 --fast

Config:
  STX_API_URL   API base URL (default: http://localhost:8000)
  STX_API_KEY   Your API key (get from https://stxai.com/dashboard)
`)
}
