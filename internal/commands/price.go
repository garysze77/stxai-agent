package commands

import (
	"fmt"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/display"
	"github.com/garysze77/stxai-agent/internal/i18n"
)

func PriceCommand(cfg *config.Config, ticker, lang string) error {
	cl := client.New(cfg.APIURL, cfg.APIKey, lang)
	price, err := cl.GetPrice(ticker)
	if err != nil {
		return fmt.Errorf(i18n.T("cli.errors.price_failed", lang, err))
	}
	display.ShowPrice(price.Ticker, price.Price, price.Currency)
	return nil
}
