package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  json.RawMessage
}

// StreamAnalyze opens an SSE connection to the streaming analyze endpoint
// and returns channels for events and errors.
func (c *Client) StreamAnalyze(ticker string, fast bool) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent, 8)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		url := fmt.Sprintf("%s/analyze/%s/stream", c.BaseURL, ticker)
		params := []string{}
		if fast {
			params = append(params, "fast=true")
		}
		if c.Lang != "" && c.Lang != "en" {
			params = append(params, "lang="+c.Lang)
		}
		if len(params) > 0 {
			url += "?" + strings.Join(params, "&")
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			errs <- fmt.Errorf("request failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			errs <- fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
			return
		}
		if resp.StatusCode == 401 {
			errs <- fmt.Errorf("invalid API key — run: stxai setup")
			return
		}
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			errs <- fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		var currentEvent string
		var dataLines []string

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "event: ") {
				currentEvent = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
			} else if line == "" && len(dataLines) > 0 {
				data := strings.Join(dataLines, "\n")
				events <- SSEEvent{
					Event: currentEvent,
					Data:  json.RawMessage(data),
				}
				currentEvent = ""
				dataLines = nil
			}
		}

		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	return events, errs
}

type PriceResponse struct {
	Ticker   string  `json:"ticker"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
}

// GetPrice does a quick price lookup for a ticker.
func (c *Client) GetPrice(ticker string) (*PriceResponse, error) {
	url := c.BaseURL + "/price/" + ticker
	if c.Lang != "" && c.Lang != "en" {
		url += "?lang=" + c.Lang
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var pr PriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return &pr, nil
}
