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

const defaultBaseURL = "https://api.stxai.io/api/v1"

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

type ChatRequest struct {
	Message      string `json:"message"`
	Session      string `json:"session_id,omitempty"`
	DeepAnalysis bool   `json:"deep_analysis"`
}

type ChatResponse struct {
	Reply      string `json:"reply"`
	SessionID  string `json:"session_id"`
	TokensUsed *int   `json:"tokens_used"`
}

type AnalyzeResponse struct {
	Ticker    string             `json:"ticker"`
	Price     float64            `json:"price"`
	Change    float64            `json:"change_pct"`
	RSI       float64            `json:"rsi"`
	MACD      float64            `json:"macd"`
	MA50      float64            `json:"ma50"`
	MA200     float64            `json:"ma200"`
	BollUpper float64            `json:"bollinger_upper"`
	BollLower float64            `json:"bollinger_lower"`
	News      []string           `json:"news"`
	Summary   string             `json:"summary"`
}

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
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var ar AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}
	return &ar, nil
}
