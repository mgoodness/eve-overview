package reference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// The two pinned reference artefacts move independently of this repository.
// These are the endpoints that say whether the pins are still current:
//
//   - the SDE publishes its current build number in a one-line JSONL manifest;
//   - the Z-S pack's only machine-readable copies are the base packs bundled in
//     the Z-S Customiser, so the Customiser's defaults directory is the pack's
//     upstream.
const (
	sdeLatestURL      = "https://developers.eveonline.com/static-data/tranquility/latest.jsonl"
	zsDefaultsAPI     = "https://api.github.com/repos/Arziel1992/Z-S-Overview-Customizer/contents/public/defaults"
	zsDefaultsRawBase = "https://raw.githubusercontent.com/Arziel1992/Z-S-Overview-Customizer/main/public/defaults"
)

// zsPackName matches a bundled full pack: zs_full_v<major>.<minor>.<patch>.yaml.
var zsPackName = regexp.MustCompile(`^zs_full_v(\d+)\.(\d+)\.(\d+)\.yaml$`)

// UpstreamStatus reports whether one pinned reference artefact is still the
// newest upstream. It is read-only: nothing here changes a pin.
type UpstreamStatus struct {
	Name    string
	Pinned  string
	Latest  string
	Current bool
	Detail  string
}

// CheckUpstream compares the pinned SDE build and Z-S reference pack against
// upstream. The client is injectable for tests; nil means http.DefaultClient. A
// GITHUB_TOKEN in the environment raises the GitHub API rate limit.
//
// It reports drift; it never changes a pin. Bumping one is a deliberate
// regeneration — the emitted profile's whole value is that it is built from
// known bytes.
func CheckUpstream(ctx context.Context, client *http.Client) ([]UpstreamStatus, error) {
	if client == nil {
		client = http.DefaultClient
	}
	sde, err := checkSDEUpstream(ctx, client, SDEBuild)
	if err != nil {
		return nil, err
	}
	pack, err := checkZSPackUpstream(ctx, client, os.Getenv("GITHUB_TOKEN"), ReferencePackSHA256)
	if err != nil {
		return nil, err
	}
	return []UpstreamStatus{sde, pack}, nil
}

// checkSDEUpstream reads the current SDE build number and compares it to the
// pinned one.
func checkSDEUpstream(ctx context.Context, client *http.Client, pinned string) (UpstreamStatus, error) {
	body, err := fetchBytes(ctx, client, sdeLatestURL, "")
	if err != nil {
		return UpstreamStatus{}, fmt.Errorf("check SDE: %w", err)
	}
	line := strings.TrimSpace(strings.SplitN(string(body), "\n", 2)[0])
	var manifest struct {
		BuildNumber int    `json:"buildNumber"`
		ReleaseDate string `json:"releaseDate"`
	}
	if err := json.Unmarshal([]byte(line), &manifest); err != nil {
		return UpstreamStatus{}, fmt.Errorf("check SDE: parse latest manifest: %w", err)
	}
	if manifest.BuildNumber == 0 {
		return UpstreamStatus{}, fmt.Errorf("check SDE: latest manifest has no build number")
	}
	latest := strconv.Itoa(manifest.BuildNumber)
	return UpstreamStatus{
		Name:    "SDE",
		Pinned:  pinned,
		Latest:  latest,
		Current: latest == pinned,
		Detail:  "released " + manifest.ReleaseDate,
	}, nil
}

// checkZSPackUpstream finds the newest full pack bundled in the Z-S Customiser
// and, when it is the pinned one, verifies the bytes still hash to the pin (it
// can be updated in place under the same filename). pin is injectable so the
// content check is testable without the real pack.
func checkZSPackUpstream(ctx context.Context, client *http.Client, token, pin string) (UpstreamStatus, error) {
	pinnedName := path.Base(ReferencePackURL)

	body, err := fetchBytes(ctx, client, zsDefaultsAPI, token)
	if err != nil {
		return UpstreamStatus{}, fmt.Errorf("check Z-S pack: %w", err)
	}
	var files []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &files); err != nil {
		return UpstreamStatus{}, fmt.Errorf("check Z-S pack: parse defaults listing: %w", err)
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		if zsPackName.MatchString(file.Name) {
			names = append(names, file.Name)
		}
	}
	newest := newestZSPack(names)
	if newest == "" {
		return UpstreamStatus{}, fmt.Errorf("check Z-S pack: no zs_full_*.yaml in the Customiser defaults")
	}

	status := UpstreamStatus{Name: "Z-S reference pack", Pinned: pinnedName, Latest: newest, Current: newest == pinnedName}
	if newest != pinnedName {
		status.Detail = "a newer pack is bundled upstream"
		return status, nil
	}

	// Same filename: it may have been updated in place. Compare the bytes at
	// the default branch to the pin.
	content, err := fetchBytes(ctx, client, zsDefaultsRawBase+"/"+pinnedName, "")
	if err != nil {
		return UpstreamStatus{}, fmt.Errorf("check Z-S pack: %w", err)
	}
	sum := sha256.Sum256(content)
	if got := hex.EncodeToString(sum[:]); got != pin {
		status.Current = false
		status.Detail = "the bundled pack changed in place (sha256 " + got[:12] + "…)"
	} else {
		status.Detail = "sha256 matches the pin"
	}
	return status, nil
}

// newestZSPack returns the highest-versioned pack name, or "" when none match.
func newestZSPack(names []string) string {
	var best string
	var bestVersion [3]int
	for _, name := range names {
		match := zsPackName.FindStringSubmatch(name)
		if match == nil {
			continue
		}
		var version [3]int
		for i := range version {
			version[i], _ = strconv.Atoi(match[i+1])
		}
		if best == "" || versionLess(bestVersion, version) {
			best, bestVersion = name, version
		}
	}
	return best
}

// versionLess reports whether a < b for two [major, minor, patch] tuples.
func versionLess(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// fetchBytes GETs url and returns the body. token, when non-empty, is sent as a
// bearer token (for the rate-limited GitHub API).
func fetchBytes(ctx context.Context, client *http.Client, url, token string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "eve-overview-upstream-check")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if strings.Contains(url, "api.github.com") {
		request.Header.Set("Accept", "application/vnd.github+json")
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, response.Status)
	}
	return io.ReadAll(response.Body)
}
