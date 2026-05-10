package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/display"
)

func AnalyzeCommand(cfg *config.Config, ticker string, fast bool) error {
	cl := client.New(cfg.APIURL, cfg.APIKey)

	// Phase 1: quick price
	price, err := cl.GetPrice(ticker)
	if err == nil {
		display.ShowPrice(price.Ticker, price.Price, price.Currency)
	} else {
		fmt.Fprintf(os.Stderr, "  price unavailable: %v\n", err)
	}

	// Phase 2: streaming analysis
	events, errs := cl.StreamAnalyze(ticker, fast)

	var stopSpinner func()
	firstContent := true

	// Handle Ctrl-C gracefully
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
				// Already shown in Phase 1; update if different
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

				// Extract node name from payload
				var payload map[string]interface{}
				json.Unmarshal(evt.Data, &payload)
				node, _ := payload["node"].(string)

				if firstContent {
					fmt.Println()
					firstContent = false
				}
				display.ShowNodeHeader(node)
				display.ShowNodeContent(node, evt.Data)

				// Start a new spinner while waiting for next node
				stopSpinner = display.Spinner("generating...")

			case "heartbeat":
				// Spinner handles this — nothing extra needed

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
