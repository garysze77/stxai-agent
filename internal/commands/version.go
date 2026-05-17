package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/garysze77/stxai-agent/internal/i18n"
)

// Version is set at build time via -ldflags.
var Version = "dev"

const (
	installScript = "curl -fsSL https://raw.githubusercontent.com/garysze77/stxai-agent/main/install.sh | sh"
	tagsAPI       = "https://api.github.com/repos/garysze77/stxai-agent/git/refs/tags"
)

func VersionCommand() {
	fmt.Printf("stxai %s\n", Version)
}

func UpdateCommand(lang string) error {
	current := strings.TrimPrefix(Version, "v")
	if current == "dev" {
		fmt.Println(i18n.T("cli.update.dev_version", lang))
		return nil
	}

	fmt.Printf("stxai %s\n", Version)
	fmt.Println(i18n.T("cli.update.checking", lang))

	latest, err := fetchLatestVersion()
	if err != nil {
		fmt.Println(i18n.T("cli.update.check_failed", lang, err))
		return nil
	}

	latestVer := strings.TrimPrefix(latest, "v")
	if latest == "" || compareSemver(latestVer, current) <= 0 {
		fmt.Println(i18n.T("cli.update.up_to_date", lang))
		return nil
	}

	fmt.Println(i18n.T("cli.update.new_version", lang, latest))
	fmt.Println()
	fmt.Printf("  %s\n", installScript)
	fmt.Println()
	return nil
}

func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", tagsAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "stxai-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	// Parse all tags and find the latest semver
	var tags []struct {
		Ref string `json:"ref"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", err
	}

	latest := ""
	for _, t := range tags {
		tag := strings.TrimPrefix(t.Ref, "refs/tags/")
		tag = strings.TrimPrefix(tag, "v")
		if !isSemver(tag) {
			continue
		}
		if latest == "" || compareSemver(tag, latest) > 0 {
			latest = tag
		}
	}
	if latest != "" {
		return "v" + latest, nil
	}
	return "", nil
}

func isSemver(v string) bool {
	parts := strings.Split(v, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

func compareSemver(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		var an, bn int
		if i < len(ap) {
			fmt.Sscanf(ap[i], "%d", &an)
		}
		if i < len(bp) {
			fmt.Sscanf(bp[i], "%d", &bn)
		}
		if an > bn {
			return 1
		}
		if an < bn {
			return -1
		}
	}
	return 0
}
