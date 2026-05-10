package commands

import (
	"fmt"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/display"
)

func PriceCommand(cfg *config.Config, ticker string) error {
	cl := client.New(cfg.APIURL, cfg.APIKey)
	price, err := cl.GetPrice(ticker)
	if err != nil {
		return fmt.Errorf("price lookup failed: %w", err)
	}
	display.ShowPrice(price.Ticker, price.Price, price.Currency)
	return nil
}
