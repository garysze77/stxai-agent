package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	lang    string
	http    *http.Client
}

type PriceResponse struct {
	Ticker   string  `json:"ticker"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

func New(baseURL, apiKey, lang string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		lang:    lang,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetPrice(ticker string) (*PriceResponse, error) {
	url := fmt.Sprintf("%s/api/v1/price/%s?lang=%s", c.baseURL, ticker, c.lang)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var pr PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return &pr, nil
}
