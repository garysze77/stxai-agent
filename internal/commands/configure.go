package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/garysze77/stxai-agent/internal/config"
)

func ConfigureCommand() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("API URL [http://localhost:8000]: ")
	apiURL, _ := reader.ReadString('\n')
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" {
		apiURL = "http://localhost:8000"
	}

	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	if err := config.Save(apiURL, apiKey); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("  Config saved to ~/.stx/config.json")
	return nil
}
