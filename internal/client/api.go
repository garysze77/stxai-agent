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
	Lang    string
	HTTP    *http.Client
}

func New(baseURL, apiKey, lang string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if lang == "" {
		lang = "en"
	}
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Lang:    lang,
		HTTP:    &http.Client{Timeout: 300 * time.Second},
	}
}

func langQ(lang string) string {
	if lang == "" || lang == "en" {
		return ""
	}
	return "?lang=" + lang
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

type PortfolioPosition struct {
	Ticker       string  `json:"ticker"`
	Shares       float64 `json:"shares"`
	CostBasis    float64 `json:"cost_basis"`
	CurrentPrice float64 `json:"current_price"`
	CostTotal    float64 `json:"cost_total"`
	Value        float64 `json:"value"`
	Pnl          float64 `json:"pnl"`
	PnlPct       float64 `json:"pnl_pct"`
}

type PortfolioResponse struct {
	Positions   []PortfolioPosition `json:"positions"`
	TotalValue  float64             `json:"total_value"`
	TotalCost   float64             `json:"total_cost"`
	TotalPnl    float64             `json:"total_pnl"`
	TotalPnlPct float64             `json:"total_pnl_pct"`
}

type Alert struct {
	ID        string  `json:"id"`
	Ticker    string  `json:"ticker"`
	Condition string  `json:"condition"`
	Price     float64 `json:"price"`
	Triggered bool    `json:"triggered"`
}

type Autorule struct {
	ID              string  `json:"id"`
	Ticker          string  `json:"ticker"`
	Condition       string  `json:"condition"`
	Threshold       float64 `json:"threshold"`
	Enabled         bool    `json:"enabled"`
	LastTriggeredAt int64   `json:"last_triggered_at"`
}

type EarningsResponse struct {
	Ticker          string  `json:"ticker"`
	Market          string  `json:"market"`
	EarningsDate    string  `json:"earnings_date"`
	EpsEstimate     float64 `json:"eps_estimate"`
	RevenueEstimate float64 `json:"revenue_estimate"`
	EarningsGrowth  float64 `json:"earnings_growth"`
	RevenueGrowth   float64 `json:"revenue_growth"`
	Source          string  `json:"source"`
}

type CompareRequest struct {
	Tickers string `json:"tickers"`
}

type CompareResponse struct {
	Tickers    string `json:"tickers"`
	Comparison string `json:"comparison"`
}

// ── Helpers ──

type errDetail struct {
	Detail string `json:"detail"`
}

func readError(resp *http.Response) string {
	var e errDetail
	json.NewDecoder(resp.Body).Decode(&e)
	if e.Detail != "" {
		return e.Detail
	}
	return fmt.Sprintf("API error %d", resp.StatusCode)
}

// ── API methods ──

func (c *Client) Chat(message, session, lang string, deep bool) (*ChatResponse, error) {
	body := ChatRequest{Message: message, Session: session, DeepAnalysis: deep}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/chat"+langQ(lang), bytes.NewReader(b))
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
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var cr ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

func (c *Client) ChatStream(message, session, lang string, deep bool, onChunk func(ChatResponse)) error {
	body := ChatRequest{Message: message, Session: session, DeepAnalysis: deep}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/chat"+langQ(lang), bytes.NewReader(b))
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

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}

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
				onChunk(chunk)
			}
		}
	}
	return scanner.Err()
}

func (c *Client) Analyze(ticker, lang string) (*AnalyzeResponse, error) {
	url := c.BaseURL + "/analyze/" + ticker + langQ(lang)
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

	var ar AnalyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

func (c *Client) Scan(market, criteria, lang string) (*ChatResponse, error) {
	body := ScanRequest{Market: market, Criteria: criteria}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/scan"+langQ(lang), bytes.NewReader(b))
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

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
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

func (c *Client) PairGenerate() (string, error) {
	req, err := http.NewRequest("POST", c.BaseURL+"/pairing/generate", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%s", readError(resp))
	}

	var result struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Code, nil
}

func (c *Client) PairClaim(code string, telegramUserID int64, telegramUsername string) error {
	body := map[string]any{
		"code":              code,
		"telegram_user_id":  telegramUserID,
		"telegram_username": telegramUsername,
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/pairing/claim", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) Portfolio() (*PortfolioResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/portfolio", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var pr PortfolioResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) AddPosition(ticker string, shares, costBasis float64) error {
	body := map[string]any{
		"ticker":     ticker,
		"shares":     shares,
		"cost_basis": costBasis,
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/portfolio", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) RemovePosition(ticker string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/portfolio/"+ticker, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) Earnings(ticker string) (*EarningsResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/earnings/"+ticker, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var er EarningsResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		return nil, err
	}
	return &er, nil
}

func (c *Client) Compare(tickers string) (*CompareResponse, error) {
	body := CompareRequest{Tickers: tickers}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/compare", bytes.NewReader(b))
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
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var cr CompareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

func (c *Client) GetAlerts() ([]Alert, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/alerts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var result struct {
		Alerts []Alert `json:"alerts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Alerts, nil
}

func (c *Client) CreateAlert(ticker, condition string, price float64) error {
	body := map[string]any{
		"ticker":    ticker,
		"condition": condition,
		"price":     price,
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/alerts", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) DeleteAlert(alertID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/alerts/"+alertID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) GetAutorules() ([]Autorule, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/autorules", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", readError(resp))
	}

	var result struct {
		Rules []Autorule `json:"rules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Rules, nil
}

func (c *Client) CreateAutorule(ticker, condition string, threshold float64) error {
	body := map[string]any{
		"ticker":    ticker,
		"condition": condition,
		"threshold": threshold,
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", c.BaseURL+"/autorules", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) DeleteAutorule(ruleID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/autorules/"+ruleID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) UpdateSettings(settings map[string]any) error {
	b, _ := json.Marshal(settings)

	req, err := http.NewRequest("POST", c.BaseURL+"/user/settings", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return fmt.Errorf("daily quota exceeded — upgrade at https://stxai.app")
	}
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key — run: stxai setup")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s", readError(resp))
	}
	return nil
}

func (c *Client) News(ticker, lang string) (*NewsResponse, error) {
	url := c.BaseURL + "/news/" + ticker + langQ(lang)
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

	var nr NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&nr); err != nil {
		return nil, err
	}
	return &nr, nil
}

