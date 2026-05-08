package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIKey     string `mapstructure:"api_key"`
	APIURL     string `mapstructure:"api_url"`
	Telegram   string `mapstructure:"telegram_token"`
	Model      string `mapstructure:"model"`
	MaxHistory int    `mapstructure:"max_history"`
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "stxai")
}

func configPath() string {
	return filepath.Join(configDir(), "config.yaml")
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetDefault("api_url", "https://api.stxai.io/api/v1")
	v.SetDefault("model", "stxai-agent")
	v.SetDefault("max_history", 20)

	_ = v.ReadInConfig()
	// File doesn't exist yet — fine, use defaults

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigFile(configPath())
	v.Set("api_key", cfg.APIKey)
	v.Set("api_url", cfg.APIURL)
	v.Set("telegram_token", cfg.Telegram)
	v.Set("model", cfg.Model)
	v.Set("max_history", cfg.MaxHistory)

	return v.WriteConfig()
}
