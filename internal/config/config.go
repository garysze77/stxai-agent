package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIURL string `json:"api_url"`
	APIKey string `json:"api_key"`
}

func Load() (*Config, error) {
	cfg := &Config{
		APIURL: getEnv("STX_API_URL", "http://localhost:8000"),
		APIKey: getEnv("STX_API_KEY", ""),
	}

	// Also try config file
	if cfg.APIKey == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			data, err := os.ReadFile(filepath.Join(home, ".stx", "config.json"))
			if err == nil {
				var fileCfg Config
				if json.Unmarshal(data, &fileCfg) == nil {
					if cfg.APIURL == "http://localhost:8000" && fileCfg.APIURL != "" {
						cfg.APIURL = fileCfg.APIURL
					}
					if cfg.APIKey == "" {
						cfg.APIKey = fileCfg.APIKey
					}
				}
			}
		}
	}

	return cfg, nil
}

func Save(apiURL, apiKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".stx")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cfg := Config{APIURL: apiURL, APIKey: apiKey}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0600)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
