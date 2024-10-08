package version

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/mod/semver"
)

const LatestReleaseURL = "https://github.com/massdriver-cloud/airlock/releases/latest"

var (
	version = "unknown"
	gitSHA  = "unknown"
)

func AirlockVersion() string {
	return version
}

func AirlockGitSHA() string {
	return gitSHA
}

func SetVersion(setVersion string) {
	version = setVersion
}

func GetLatestVersion() (string, error) {
	ctx := context.Background()
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, LatestReleaseURL, nil)
	if reqErr != nil {
		return "", reqErr
	}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		return "", respErr
	}
	defer resp.Body.Close()

	redirectURL := resp.Request.URL.String()
	parts := strings.Split(redirectURL, "/")
	latestVersion := parts[len(parts)-1]
	return latestVersion, nil
}

func CheckForNewerVersionAvailable(latestVersion string) (bool, string) {
	currentVersion := version

	// semver requires "v" for version (e.g., v1.0.0 not 1.0.0). Adds "v" if missing
	if !strings.HasPrefix(currentVersion, "v") {
		currentVersion = "v" + currentVersion
	}

	if !strings.HasPrefix(latestVersion, "v") {
		latestVersion = "v" + latestVersion
	}

	if semver.Compare(currentVersion, latestVersion) < 0 {
		return true, latestVersion
	}

	return false, latestVersion
}
