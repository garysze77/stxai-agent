package web

import (
	"encoding/json"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/garysze77/stxai-agent/internal/client"
	"github.com/garysze77/stxai-agent/internal/config"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	client *client.Client
	cfg    *config.Config
	http   *http.Server
	port   string
}

type configPayload struct {
	APIKey        string `json:"api_key,omitempty"`
	TelegramToken string `json:"telegram_token,omitempty"`
	Lang          string `json:"lang,omitempty"`
	Configured    bool   `json:"configured"`
}

func New(c *client.Client, cfg *config.Config, port string) *Server {
	return &Server{client: c, cfg: cfg, port: port}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Static file server from embedded FS
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("static fs: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// Config endpoints
	mux.HandleFunc("/api/config", s.handleConfig)

	// API proxy endpoints
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/analyze/", s.handleAnalyze)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/news/", s.handleNews)
	mux.HandleFunc("/api/clear", s.handleClear)
	mux.HandleFunc("/api/pair", s.handlePairGenerate)

	// Portfolio endpoints
	mux.HandleFunc("/api/portfolio", s.handlePortfolio)
	mux.HandleFunc("/api/portfolio/", s.handlePortfolioRemove)

	// Earnings endpoint
	mux.HandleFunc("/api/earnings/", s.handleEarnings)

	// Compare endpoint
	mux.HandleFunc("/api/compare", s.handleCompare)

	// Alerts endpoints
	mux.HandleFunc("/api/alerts", s.handleAlerts)
	mux.HandleFunc("/api/alerts/", s.handleAlertDelete)

	// Autorules endpoints
	mux.HandleFunc("/api/autorules", s.handleAutorules)
	mux.HandleFunc("/api/autorules/", s.handleAutoruleDelete)

	// Settings endpoint
	mux.HandleFunc("/api/settings", s.handleSettings)

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"configured": s.cfg.APIKey != "",
		})
	})

	s.http = &http.Server{
		Addr:         ":" + s.port,
		Handler:      withCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 310 * time.Second, // longer than client timeout
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🌐 Web UI → http://localhost:%s", s.port)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown() error {
	if s.http != nil {
		return s.http.Close()
	}
	return nil
}

func (s *Server) Addr() string {
	return "http://localhost:" + s.port
}

// ── Config handlers ──

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		masked := "••••••••"
		key := ""
		if s.cfg.APIKey != "" {
			key = masked
		}
		token := ""
		if s.cfg.Telegram != "" {
			token = "••••••••"
		}
		lang := s.cfg.Lang
		if lang == "" {
			lang = "en"
		}
		json.NewEncoder(w).Encode(configPayload{
			APIKey:        key,
			TelegramToken: token,
			Lang:          lang,
			Configured:    s.cfg.APIKey != "",
		})

	case http.MethodPost:
		var p configPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		if p.APIKey != "" && p.APIKey != "••••••••" {
			s.cfg.APIKey = p.APIKey
		}
		if p.TelegramToken != "" && p.TelegramToken != "••••••••" {
			s.cfg.Telegram = p.TelegramToken
		}
		if p.Lang != "" {
			s.cfg.Lang = p.Lang
		}
		if err := config.Save(s.cfg); err != nil {
			writeError(w, 500, "save config: "+err.Error())
			return
		}
		// Update client with new API key
		s.client.APIKey = s.cfg.APIKey
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		writeError(w, 405, "method not allowed")
	}
}

// ── API proxy handlers ──

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "POST required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	var body struct {
		Message   string `json:"message"`
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}

	resp, err := s.client.Chat(body.Message, body.SessionID, s.cfg.Lang, false)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "GET required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	ticker := strings.TrimPrefix(r.URL.Path, "/api/analyze/")
	ticker = strings.TrimSuffix(ticker, "/")
	if ticker == "" {
		writeError(w, 400, "ticker required")
		return
	}

	resp, err := s.client.Analyze(ticker, s.cfg.Lang)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "POST required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	var body struct {
		Market   string `json:"market"`
		Criteria string `json:"criteria"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}

	resp, err := s.client.Scan(body.Market, body.Criteria, s.cfg.Lang)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "GET required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	ticker := strings.TrimPrefix(r.URL.Path, "/api/news/")
	ticker = strings.TrimSuffix(ticker, "/")
	if ticker == "" {
		writeError(w, 400, "ticker required")
		return
	}

	resp, err := s.client.News(ticker, s.cfg.Lang)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleClear(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── Helpers ──

func (s *Server) checkAuth(w http.ResponseWriter) bool {
	if s.cfg.APIKey == "" {
		writeError(w, 401, "API key not configured — run setup first")
		return false
	}
	return true
}

func (s *Server) handlePairGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed")
		return
	}
	code, err := s.client.PairGenerate()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"code": code})
}

// ── Portfolio handlers ──

func (s *Server) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if !s.checkAuth(w) {
			return
		}
		resp, err := s.client.Portfolio()
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		json.NewEncoder(w).Encode(resp)

	case http.MethodPost:
		if !s.checkAuth(w) {
			return
		}
		var body struct {
			Ticker    string  `json:"ticker"`
			Shares    float64 `json:"shares"`
			CostBasis float64 `json:"cost_basis"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		if err := s.client.AddPosition(body.Ticker, body.Shares, body.CostBasis); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) handlePortfolioRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, 405, "DELETE required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	ticker := strings.TrimPrefix(r.URL.Path, "/api/portfolio/")
	ticker = strings.TrimSuffix(ticker, "/")
	if ticker == "" {
		writeError(w, 400, "ticker required")
		return
	}

	if err := s.client.RemovePosition(ticker); err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── Earnings handler ──

func (s *Server) handleEarnings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "GET required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	ticker := strings.TrimPrefix(r.URL.Path, "/api/earnings/")
	ticker = strings.TrimSuffix(ticker, "/")
	if ticker == "" {
		writeError(w, 400, "ticker required")
		return
	}

	resp, err := s.client.Earnings(ticker)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ── Compare handler ──

func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "POST required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	var body struct {
		Tickers string `json:"tickers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}

	resp, err := s.client.Compare(body.Tickers)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ── Alerts handlers ──

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if !s.checkAuth(w) {
			return
		}
		alerts, err := s.client.GetAlerts()
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		json.NewEncoder(w).Encode(alerts)

	case http.MethodPost:
		if !s.checkAuth(w) {
			return
		}
		var body struct {
			Ticker    string  `json:"ticker"`
			Condition string  `json:"condition"`
			Price     float64 `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		if err := s.client.CreateAlert(body.Ticker, body.Condition, body.Price); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) handleAlertDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, 405, "DELETE required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/alerts/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		writeError(w, 400, "id required")
		return
	}

	if err := s.client.DeleteAlert(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── Autorules handlers ──

func (s *Server) handleAutorules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		if !s.checkAuth(w) {
			return
		}
		rules, err := s.client.GetAutorules()
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		json.NewEncoder(w).Encode(rules)

	case http.MethodPost:
		if !s.checkAuth(w) {
			return
		}
		var body struct {
			Ticker    string  `json:"ticker"`
			Condition string  `json:"condition"`
			Threshold float64 `json:"threshold"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "invalid JSON")
			return
		}
		if err := s.client.CreateAutorule(body.Ticker, body.Condition, body.Threshold); err != nil {
			writeError(w, 500, err.Error())
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) handleAutoruleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, 405, "DELETE required")
		return
	}
	if !s.checkAuth(w) {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/autorules/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		writeError(w, 400, "id required")
		return
	}

	if err := s.client.DeleteAutorule(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ── Settings handler ──

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "POST required")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid JSON")
		return
	}

	if err := s.client.UpdateSettings(body); err != nil {
		writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// OpenBrowser tries to open the web UI in the default browser.
func OpenBrowser(addr string) {
	url := addr
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	}

	if runtime.GOOS != "windows" {
		args = []string{url}
	}

	// Best-effort: don't block or error
	go func() {
		time.Sleep(500 * time.Millisecond) // wait for server to be ready
		c := exec.Command(cmd, args...)
		c.Stdout = io.Discard
		c.Stderr = io.Discard
		c.Run()
	}()
}
