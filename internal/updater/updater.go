package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UpdateChecker checks for new releases on GitHub
type UpdateChecker struct {
	httpClient *http.Client
	repoOwner  string
	repoName   string
}

// GitHubRelease represents the JSON response from GitHub's releases API
type GitHubRelease struct {
	TagName string `json:"tag_name"` // e.g., "v1.10.0"
	Name    string `json:"name"`
}

// NewUpdateChecker creates a new UpdateChecker for the specified GitHub repository
func NewUpdateChecker(owner, repo string) *UpdateChecker {
	return &UpdateChecker{
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // Shorter timeout for non-critical operation
		},
		repoOwner: owner,
		repoName:  repo,
	}
}

// CheckForUpdate queries GitHub for the latest release and compares it with the current version
// Returns: latestVersion, updateAvailable, error
func (uc *UpdateChecker) CheckForUpdate(currentVersion string) (string, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		uc.repoOwner, uc.repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", false, fmt.Errorf("failed to create request: %w", err)
	}

	// GitHub API requires User-Agent header
	req.Header.Set("User-Agent", "timesheetz-app")

	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == 403 {
		return "", false, fmt.Errorf("GitHub API rate limit exceeded")
	}

	if resp.StatusCode != 200 {
		return "", false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", false, fmt.Errorf("failed to parse release data: %w", err)
	}

	// Compare versions
	updateAvailable := compareVersions(currentVersion, release.TagName)

	return release.TagName, updateAvailable, nil
}

// compareVersions returns true if latest > current
func compareVersions(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Handle "dev" version (local builds)
	if current == "dev" {
		return true // Always show update for dev versions
	}

	// Parse semantic versions
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare major, minor, patch
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen && i < 3; i++ {
		currentNum := 0
		latestNum := 0

		if i < len(currentParts) {
			// Parse current version part, ignore errors (default to 0)
			currentNum, _ = strconv.Atoi(currentParts[i])
		}

		if i < len(latestParts) {
			// Parse latest version part, ignore errors (default to 0)
			latestNum, _ = strconv.Atoi(latestParts[i])
		}

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}

	return false // Versions are equal
}
