package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/display"
	"github.com/garysze77/stxai-agent/internal/i18n"
)

func AnalyzeCommand(cfg *config.Config, ticker string, fast bool, lang string) error {
	cl := client.New(cfg.APIURL, cfg.APIKey, lang)

	price, err := cl.GetPrice(ticker)
	if err == nil {
		display.ShowPrice(price.Ticker, price.Price, price.Currency)
	} else {
		fmt.Fprintf(os.Stderr, i18n.T("cli.display.price_unavailable", lang, err))
	}

	events, errs := cl.StreamAnalyze(ticker, fast)

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
				display.ShowNodeHeader(node, lang)
				display.ShowNodeContent(node, evt.Data, lang)

				stopSpinner = display.Spinner(i18n.T("cli.display.generating", lang))

			case "heartbeat":
				// Spinner handles this

			case "error":
				if stopSpinner != nil {
					stopSpinner()
					stopSpinner = nil
				}
				display.ShowError(evt.Data, lang)

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
				return fmt.Errorf(i18n.T("cli.errors.stream_error", lang, err))
			}

		case <-sig:
			if stopSpinner != nil {
				stopSpinner()
			}
			fmt.Println()
			return fmt.Errorf(i18n.T("cli.errors.interrupted", lang))
		}
	}
}
