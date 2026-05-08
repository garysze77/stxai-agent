package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.stxai.app/api/v1"

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func New(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// ── Request types ──

type ChatRequest struct {
	Message      string `json:"message"`
	Session      string `json:"session_id,omitempty"`
	DeepAnalysis bool   `json:"deep_analysis"`
	Ticker       string `json:"ticker,omitempty"`
}

type ScanRequest struct {
	Market   string `json:"market"`
	Criteria string `json:"criteria"`
}

// ── Response types ──

type SignalData struct {
	DirectionalBias string `json:"directional_bias"`
	ConfidenceScore int    `json:"confidence_score"`
	SignalStrength  string `json:"signal_strength"`
}

type ChatResponse struct {
	Reply      string      `json:"reply"`
	SessionID  string      `json:"session_id"`
	TokensUsed *int        `json:"tokens_used"`
	Signal     *SignalData `json:"signal"`
}

type AnalyzeResponse struct {
	Ticker           string             `json:"ticker"`
	Name             string             `json:"name"`
	Price            float64            `json:"price"`
	Currency         string             `json:"currency"`
	Summary          string             `json:"summary"`
	TechnicalSignals map[string]any      `json:"technical_signals"`
	Signal           *SignalData        `json:"signal"`
}

type NewsResponse struct {
	Ticker   string `json:"ticker"`
	Articles []struct {
		Summary string `json:"summary"`
	} `json:"articles"`
}

// ── API methods ──

func (c *Client) Chat(message, session string, deep bool) (*ChatResponse, error) {
	body := ChatRequest{Message: message, Session: session, DeepAnalysis: deep}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/chat", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.vercel.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode == 400 {
		var errResp struct {
			Detail string `json:"detail"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Detail != "" {
			return nil, fmt.Errorf("%s", errResp.Detail)
		}
		return nil, fmt.Errorf("bad request")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var cr ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

func (c *Client) ChatStream(message, session string, deep bool, onChunk func(string)) error {
	body := ChatRequest{Message: message, Session: session, DeepAnalysis: deep}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/chat", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			var chunk ChatResponse
			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				onChunk(chunk.Reply)
			}
		}
	}
	return scanner.Err()
}

func (c *Client) Analyze(ticker string) (*AnalyzeResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/analyze/"+ticker, nil)
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
		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("invalid API key — run: stxai setup")
		}
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var ar AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

func (c *Client) Scan(market, criteria string) (*ChatResponse, error) {
	body := ScanRequest{Market: market, Criteria: criteria}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/scan", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("upgrade to Pro for market scanning — https://stxai.vercel.app")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	// Scan returns {results, market}, we wrap results in ChatResponse-like format
	var result struct {
		Results string `json:"results"`
		Market  string `json:"market"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &ChatResponse{Reply: result.Results}, nil
}

func (c *Client) News(ticker string) (*NewsResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/news/"+ticker, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("upgrade to Pro for news analysis — https://stxai.vercel.app")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var nr NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&nr); err != nil {
		return nil, err
	}
	return &nr, nil
}
