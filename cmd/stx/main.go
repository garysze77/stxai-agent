package main

import (
	"fmt"
	"os"

	"github.com/garysze77/stxai-agent/internal/commands"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/i18n"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, i18n.T("cli.errors.load_config", "en", err))
		os.Exit(1)
	}

	lang := cfg.Lang
	if lang == "" {
		lang = "en"
	}

	if len(os.Args) < 2 {
		printUsage(lang)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "price":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, i18n.T("cli.errors.usage_price", lang))
			os.Exit(1)
		}
		if err := commands.PriceCommand(cfg, os.Args[2], lang); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
			os.Exit(1)
		}

	case "analyze":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, i18n.T("cli.errors.usage_analyze", lang))
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
			fmt.Fprintln(os.Stderr, i18n.T("cli.errors.missing_key", lang))
			os.Exit(1)
		}
		if err := commands.AnalyzeCommand(cfg, ticker, fast, lang); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
			os.Exit(1)
		}

	case "configure":
		if err := commands.ConfigureCommand(lang); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
			os.Exit(1)
		}

	default:
		printUsage(lang)
		os.Exit(1)
	}
}

func printUsage(lang string) {
	fmt.Println(i18n.T("cli.usage.header", lang))
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.desc", lang))
	fmt.Println("  " + i18n.T("cli.usage.price", lang))
	fmt.Println("  " + i18n.T("cli.usage.analyze", lang))
	fmt.Println("  " + i18n.T("cli.usage.analyze_fast", lang))
	fmt.Println("  " + i18n.T("cli.usage.configure", lang))
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.examples", lang))
	fmt.Println("  stx price AAPL")
	fmt.Println("  stx analyze TSLA")
	fmt.Println("  stx analyze 0700 --fast")
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.config_header", lang))
	fmt.Println("  " + i18n.T("cli.usage.env_url", lang))
	fmt.Println("  " + i18n.T("cli.usage.env_key", lang))
}
