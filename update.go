package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// appVersion is the running version. It is overridden at build time via
// `-ldflags "-X main.appVersion=X.Y.Z"` (the release workflow injects the tag);
// the default tracks CHANGELOG.md for local/dev builds.
var appVersion = "0.0.2"

// updateRepo is the GitHub repository releases are pulled from.
const updateRepo = "radaiko/Turnpoint"

// AppVersion returns the running application version.
func (a *App) AppVersion() string { return appVersion }

// UpdateInfo is the result of an update check. It is best-effort: on any network
// failure Available is false and Error carries a short, user-facing reason.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Notes          string `json:"notes"`
	ReleaseURL     string `json:"releaseUrl"`
	DownloadURL    string `json:"downloadUrl"` // Windows installer asset, if any
	Error          string `json:"error"`
}

// CheckForUpdate asks GitHub for the latest release and compares versions. This
// is the application's ONLY outbound network call — it is best-effort and never
// gates any offline, core functionality.
func (a *App) CheckForUpdate() UpdateInfo {
	cur := strings.TrimPrefix(appVersion, "v")
	info := UpdateInfo{CurrentVersion: cur}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/"+updateRepo+"/releases/latest", nil)
	if err != nil {
		info.Error = "could not build request"
		return info
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Turnpoint")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		info.Error = "offline or unreachable"
		return info
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("GitHub returned %d", resp.StatusCode)
		return info
	}

	var rel struct {
		TagName    string `json:"tag_name"`
		HTMLURL    string `json:"html_url"`
		Body       string `json:"body"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		info.Error = "unexpected response"
		return info
	}

	info.LatestVersion = strings.TrimPrefix(rel.TagName, "v")
	info.Notes = rel.Body
	info.ReleaseURL = rel.HTMLURL
	for _, as := range rel.Assets {
		n := strings.ToLower(as.Name)
		if strings.HasSuffix(n, ".exe") && (strings.Contains(n, "setup") || strings.Contains(n, "installer")) {
			info.DownloadURL = as.BrowserDownloadURL
			break
		}
	}
	info.Available = !rel.Prerelease && semverLess(cur, info.LatestVersion)
	return info
}

// OpenReleasePage opens a release URL (or the latest-release page) in the user's
// default browser.
func (a *App) OpenReleasePage(url string) {
	if url == "" {
		url = "https://github.com/" + updateRepo + "/releases/latest"
	}
	wruntime.BrowserOpenURL(a.ctx, url)
}

// DownloadAndRunUpdate downloads the Windows installer, launches it, and quits
// the app so the installer can replace the running files. On non-Windows
// platforms (or when no installer asset is available) it opens the release page
// instead.
func (a *App) DownloadAndRunUpdate(url string) error {
	if runtime.GOOS != "windows" || !strings.HasSuffix(strings.ToLower(url), ".exe") {
		a.OpenReleasePage(url)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("could not start download")
	}
	req.Header.Set("User-Agent", "Turnpoint")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed — check your connection")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed (HTTP %d)", resp.StatusCode)
	}

	tmp := filepath.Join(os.TempDir(), "Turnpoint-update-setup.exe")
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("could not save the installer")
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("download was interrupted")
	}
	f.Close()

	if err := exec.Command(tmp).Start(); err != nil {
		return fmt.Errorf("could not launch the installer")
	}
	// Quit so the installer can overwrite the running executable.
	wruntime.Quit(a.ctx)
	return nil
}

// semverLess reports whether dotted numeric version a is older than b.
func semverLess(a, b string) bool {
	pa, pb := parseVer(a), parseVer(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func parseVer(s string) [3]int {
	var out [3]int
	s = strings.SplitN(s, "-", 2)[0] // drop any pre-release suffix
	for i, p := range strings.SplitN(s, ".", 3) {
		if i > 2 {
			break
		}
		n, _ := strconv.Atoi(strings.TrimSpace(p))
		out[i] = n
	}
	return out
}
