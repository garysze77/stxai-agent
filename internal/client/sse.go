package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SSEEvent struct {
	Event string
	Data  json.RawMessage
}

func (c *Client) StreamAnalyze(ticker string, fast bool) (<-chan SSEEvent, <-chan error) {
	events := make(chan SSEEvent, 8)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		url := fmt.Sprintf("%s/api/v1/analyze/%s/stream?lang=%s", c.baseURL, ticker, c.lang)
		if fast {
			url += "&fast=true"
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")

		resp, err := c.http.Do(req)
		if err != nil {
			errs <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
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
