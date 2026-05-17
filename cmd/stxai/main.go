package main

import (
	"fmt"
	"os"
	"strconv"

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

	// No args → start web dashboard (if configured) or show setup guide
	if len(os.Args) < 2 {
		if cfg.APIKey != "" {
			port := 8420
			if err := commands.ServeCommand(cfg, port, lang); err != nil {
				fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
				os.Exit(1)
			}
			return
		}
		printSetup(lang)
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

	case "configure", "setup":
		if err := commands.ConfigureCommand(lang); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
			os.Exit(1)
		}

	case "version", "--version", "-v":
		commands.VersionCommand()

	case "update":
		if err := commands.UpdateCommand(lang); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
			os.Exit(1)
		}

	case "serve":
		port := 8420
		for i, a := range os.Args[2:] {
			if a == "--port" && i+1 < len(os.Args)-2 {
				if p, err := strconv.Atoi(os.Args[i+3]); err == nil && p > 0 && p < 65536 {
					port = p
				}
			}
		}
		if cfg.APIKey == "" {
			fmt.Fprintln(os.Stderr, i18n.T("cli.errors.missing_key", lang))
			os.Exit(1)
		}
		if err := commands.ServeCommand(cfg, port, lang); err != nil {
			fmt.Fprintf(os.Stderr, i18n.T("cli.errors.command_error", lang, err))
			os.Exit(1)
		}

	default:
		printUsage(lang)
		os.Exit(1)
	}
}

func printSetup(lang string) {
	fmt.Println(i18n.T("cli.usage.header", lang))
	fmt.Println()
	fmt.Println(i18n.T("cli.setup.welcome", lang))
	fmt.Println()
	fmt.Println("  stxai configure")
	fmt.Println()
	fmt.Println(i18n.T("cli.setup.get_key", lang))
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.config_header", lang))
	fmt.Println("  " + i18n.T("cli.usage.env_url", lang))
	fmt.Println("  " + i18n.T("cli.usage.env_key", lang))
}

func printUsage(lang string) {
	fmt.Println(i18n.T("cli.usage.header", lang))
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.desc", lang))
	fmt.Println("  " + i18n.T("cli.usage.price", lang))
	fmt.Println("  " + i18n.T("cli.usage.analyze", lang))
	fmt.Println("  " + i18n.T("cli.usage.analyze_fast", lang))
	fmt.Println("  " + i18n.T("cli.usage.serve", lang))
	fmt.Println("  " + i18n.T("cli.usage.update", lang))
	fmt.Println("  " + i18n.T("cli.usage.configure", lang))
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.examples", lang))
	fmt.Println("  stxai price AAPL")
	fmt.Println("  stxai analyze TSLA")
	fmt.Println("  stxai analyze 0700 --fast")
	fmt.Println()
	fmt.Println(i18n.T("cli.usage.config_header", lang))
	fmt.Println("  " + i18n.T("cli.usage.env_url", lang))
	fmt.Println("  " + i18n.T("cli.usage.env_key", lang))
}
