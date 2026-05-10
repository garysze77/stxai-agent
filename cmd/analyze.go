package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/display"

	"github.com/spf13/cobra"
)

func AnalyzeCmd() *cobra.Command {
	var fast bool

	cmd := &cobra.Command{
		Use:   "analyze <TICKER>",
		Short: "Streaming deep analysis with live per-node results",
		Long: `Run a multi-agent debate analysis on a stock ticker with live streaming.

Results stream in layer by layer: price → bull case → bear case →
lead synthesis → signal card. Each section appears as soon as it completes.

Examples:
  stxai analyze AAPL
  stxai analyze TSLA --fast
  stxai analyze 0700`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ticker := args[0]
			return runAnalyze(ticker, fast)
		},
	}

	cmd.Flags().BoolVar(&fast, "fast", false, "Fast mode (45s timeout)")
	return cmd
}

func runAnalyze(ticker string, fast bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w (run 'stxai setup' first)", err)
	}
	if cfg.APIKey == "" {
		return fmt.Errorf("API key not set — run 'stxai setup'")
	}

	c := client.New(cfg.APIURL, cfg.APIKey)

	// Phase 1: quick price
	price, err := c.GetPrice(ticker)
	if err == nil {
		display.ShowPrice(price.Ticker, price.Price, price.Currency)
	} else {
		fmt.Fprintf(os.Stderr, "  price unavailable: %v\n", err)
	}

	// Phase 2: streaming analysis
	events, errs := c.StreamAnalyze(ticker, fast)

	var stopSpinner func()
	firstContent := true

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)

	for {
		select {
		case evt, ok := <-events:
			if !ok {
				if stopSpinner != nil {
					stopSpinner()
				}
				fmt.Println()
				return nil
			}

			switch evt.Event {
			case "price":
				var payload map[string]interface{}
				json.Unmarshal(evt.Data, &payload)
				if p, ok := payload["price"].(float64); ok && p > 0 && (price == nil || price.Price == 0) {
					display.ShowPrice(ticker, p, "")
				}

			case "node_complete":
				if stopSpinner != nil {
					stopSpinner()
					stopSpinner = nil
				}
				var payload map[string]interface{}
				json.Unmarshal(evt.Data, &payload)
				node, _ := payload["node"].(string)
				if firstContent {
					fmt.Println()
					firstContent = false
				}
				display.ShowNodeHeader(node)
				display.ShowNodeContent(node, evt.Data)
				stopSpinner = display.Spinner("generating...")

			case "heartbeat":
				// spinner handles this

			case "error":
				if stopSpinner != nil {
					stopSpinner()
					stopSpinner = nil
				}
				display.ShowError(evt.Data)

			case "done":
				if stopSpinner != nil {
					stopSpinner()
					stopSpinner = nil
				}
				fmt.Println()
				return nil
			}

		case err := <-errs:
			if err != nil {
				return fmt.Errorf("stream error: %w", err)
			}

		case <-sig:
			if stopSpinner != nil {
				stopSpinner()
			}
			fmt.Println()
			return fmt.Errorf("interrupted")
		}
	}
}
