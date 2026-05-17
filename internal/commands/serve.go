package commands

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/i18n"
	"github.com/garysze77/stxai-agent/internal/web"
)

func ServeCommand(cfg *config.Config, port int, lang string) error {
	if cfg.APIKey == "" {
		return fmt.Errorf(i18n.T("cli.errors.missing_key", lang))
	}

	apiURL, err := url.Parse(cfg.APIURL)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}

	mux := http.NewServeMux()

	// API proxy — injects auth header
	proxy := httputil.NewSingleHostReverseProxy(apiURL)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = apiURL.Host
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Accept-Language", lang)
	}

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	// Static files from embedded FS
	staticFS, err := fs.Sub(web.Static, ".")
	if err != nil {
		return fmt.Errorf("failed to load embedded files: %w", err)
	}
	staticHandler := http.FileServer(http.FS(staticFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// API paths handled above
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			proxy.ServeHTTP(w, r)
			return
		}

		// Specific static files
		switch r.URL.Path {
		case "/style.css", "/app.js":
			staticHandler.ServeHTTP(w, r)
		default:
			// Everything else → index.html (SPA fallback)
			r.URL.Path = "/"
			staticHandler.ServeHTTP(w, r)
		}
	})

	addr := fmt.Sprintf("localhost:%d", port)
	server := &http.Server{Addr: addr, Handler: mux}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println()
		log.Println("Shutting down…")
		server.Close()
	}()

	serverURL := fmt.Sprintf("http://%s", addr)
	fmt.Printf("╭──────────────────────────────────────╮\n")
	fmt.Printf("│  STX AI Dashboard                     │\n")
	fmt.Printf("│  %s                 │\n", serverURL)
	fmt.Printf("│  Press Ctrl+C to stop                 │\n")
	fmt.Printf("╰──────────────────────────────────────╯\n")

	// Auto-open browser
	go openBrowser(serverURL)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	log.Println("Server stopped.")
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}
