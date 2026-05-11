package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/garysze77/stxai-agent/internal/config"
	"github.com/garysze77/stxai-agent/internal/i18n"
)

func ConfigureCommand(lang string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(i18n.T("cli.configure.api_url_prompt", lang))
	apiURL, _ := reader.ReadString('\n')
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" {
		apiURL = "http://localhost:8000"
	}

	fmt.Print(i18n.T("cli.configure.api_key_prompt", lang))
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf(i18n.T("cli.configure.api_key_required", lang))
	}

	if err := config.SaveWithLang(apiURL, apiKey, lang); err != nil {
		return fmt.Errorf(i18n.T("cli.configure.save_failed", lang, err))
	}

	fmt.Println(i18n.T("cli.configure.saved", lang))
	return nil
}
