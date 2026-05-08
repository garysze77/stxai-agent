package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev"

const (
	releaseAPI = "https://api.github.com/repos/garysze77/stxai-agent/releases/latest"
	downloadFmt = "https://github.com/garysze77/stxai-agent/releases/latest/download/stxai_%s_%s.%s"
)

type releaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func UpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for updates and install the latest version",
		Long:  "Checks GitHub Releases for a newer version and auto-updates the binary.",
		RunE:  runUpdate,
	}
}

func runUpdate(_ *cobra.Command, _ []string) error {
	fmt.Printf("Current version: %s\n", version)
	fmt.Println("Checking for updates...")

	latest, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("check release: %w", err)
	}

	latestVer := strings.TrimPrefix(latest.TagName, "v")
	currentVer := strings.TrimPrefix(version, "v")

	if latestVer == currentVer || latestVer == "" {
		fmt.Printf("✅ You're on the latest version (%s).\n", version)
		return nil
	}

	// Compare versions — simple string comparison
	if !isNewer(latestVer, currentVer) {
		fmt.Printf("✅ You're on the latest version (%s).\n", version)
		return nil
	}

	fmt.Printf("📦 New version available: %s (current: %s)\n", latest.TagName, version)
	fmt.Println("Downloading...")

	// Figure out the asset name
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	assetName := fmt.Sprintf("stxai_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	// Download
	tmpDir, err := os.MkdirTemp("", "stxai-update")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	downloadPath := filepath.Join(tmpDir, assetName)

	// Try downloading from the known URL pattern
	downloadURL := fmt.Sprintf(downloadFmt, runtime.GOOS, runtime.GOARCH, ext)
	if err := downloadFile(downloadPath, downloadURL); err != nil {
		// Fall back: search assets
		found := false
		for _, a := range latest.Assets {
			if strings.Contains(a.Name, runtime.GOOS) && strings.Contains(a.Name, runtime.GOARCH) {
				if err := downloadFile(downloadPath, a.URL); err != nil {
					return fmt.Errorf("download: %w", err)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
	}

	// Extract binary
	newBin, err := extractBinary(downloadPath, tmpDir)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	// Verify
	if fi, err := os.Stat(newBin); err != nil || fi.Size() < 1000 {
		return fmt.Errorf("downloaded binary appears corrupted")
	}

	// Get current binary path
	currentBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get current binary path: %w", err)
	}
	currentBin, _ = filepath.EvalSymlinks(currentBin)

	// Replace: copy new over old
	if err := os.Rename(newBin, currentBin); err != nil {
		// If rename fails (cross-device), copy instead
		if err := copyFile(newBin, currentBin); err != nil {
			return fmt.Errorf("replace binary: %w", err)
		}
	}

	// Make executable (non-Windows)
	if runtime.GOOS != "windows" {
		os.Chmod(currentBin, 0755)
	}

	fmt.Printf("✅ Updated to %s!\n", latest.TagName)
	fmt.Println("Restart the service to use the new version:")
	fmt.Println("  stxai service uninstall && sleep 2 && stxai service install")
	return nil
}

func fetchLatestRelease() (*releaseInfo, error) {
	req, _ := http.NewRequest("GET", releaseAPI, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "stxai-agent")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var info releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func downloadFile(dst, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("release asset not found at %s", url)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractBinary(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destDir)
	}
	return extractTarGz(archivePath, destDir)
}

func extractTarGz(path, dest string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if strings.HasSuffix(hdr.Name, "stxai") || hdr.Name == "stxai" || hdr.Name == "stxai.exe" {
			outPath := filepath.Join(dest, filepath.Base(hdr.Name))
			out, err := os.Create(outPath)
			if err != nil {
				return "", err
			}
			defer out.Close()
			os.Chmod(outPath, 0755)
			if _, err := io.Copy(out, tr); err != nil {
				return "", err
			}
			return outPath, nil
		}
	}
	return "", fmt.Errorf("stxai binary not found in archive")
}

func extractZip(path, dest string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "stxai.exe") || f.Name == "stxai.exe" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			outPath := filepath.Join(dest, "stxai.exe")
			out, err := os.Create(outPath)
			if err != nil {
				return "", err
			}
			defer out.Close()
			if _, err := io.Copy(out, rc); err != nil {
				return "", err
			}
			return outPath, nil
		}
	}
	return "", fmt.Errorf("stxai.exe not found in archive")
}

func isNewer(a, b string) bool {
	// Strip 'v' prefix if present
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	// Parse X.Y.Z
	var aMajor, aMinor, aPatch, bMajor, bMinor, bPatch int
	fmt.Sscanf(a, "%d.%d.%d", &aMajor, &aMinor, &aPatch)
	fmt.Sscanf(b, "%d.%d.%d", &bMajor, &bMinor, &bPatch)

	if aMajor != bMajor {
		return aMajor > bMajor
	}
	if aMinor != bMinor {
		return aMinor > bMinor
	}
	return aPatch > bPatch
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
