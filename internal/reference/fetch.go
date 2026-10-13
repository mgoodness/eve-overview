// Package reference fetches the two pinned reference artefacts the generator
// reads — the Z-S reference pack and the SDE Group matrix — and caches them in
// a user cache directory. Every download is verified against a pinned SHA-256:
// a mismatch fails generation rather than feeding the generator stale or
// tampered data.
package reference

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// The pinned reference artefacts. URLs point at immutable revisions and every
// fetched byte is checked against the SHA-256 beside it.
const (
	ReferencePackURL = "https://raw.githubusercontent.com/Arziel1992/Z-S-Overview-Customizer/40a4bc3ed3bf619638f129e6b8a524867733d802/public/defaults/zs_full_v10.06.09.yaml"

	ReferencePackSHA256 = "412541b10959c186dc7958a66fd4362484b72dabce0e3bf12728a57c1aa5f346"

	SDEURL = "https://developers.eveonline.com/static-data/tranquility/eve-online-static-data-3607705-jsonl.zip"

	SDESHA256 = "905a199f0114be2e251f98d17ff0adc36bbb38a94febc828276285bf6d2e6315"

	// SDEBuild is the Static Data Export build the pinned archive was cut from,
	// the build CCP names in the archive's own URL. It is recorded beside the
	// URL and checksum so a release consumer can identify the data the profile
	// was generated against even if the URL moves.
	SDEBuild = "3607705"
)

// Fetcher fetches pinned artefacts and caches them under a directory. The
// first call for an artefact downloads it; every later call reads the cache,
// so generation works offline once warm.
type Fetcher struct {
	cacheDir string
	client   *http.Client

	referencePackURL string
	referencePackSHA string
	sdeURL           string
	sdeSHA           string
}

// New returns a Fetcher caching under cacheDir.
func New(cacheDir string) *Fetcher {
	return newWithClient(cacheDir, http.DefaultClient)
}

func newWithClient(cacheDir string, client *http.Client) *Fetcher {
	return &Fetcher{
		cacheDir:         cacheDir,
		client:           client,
		referencePackURL: ReferencePackURL,
		referencePackSHA: ReferencePackSHA256,
		sdeURL:           SDEURL,
		sdeSHA:           SDESHA256,
	}
}

// ReferencePackCachePath is where the reference pack is cached under dir. The
// pinned checksum in the name means a changed pin cannot read a stale cache.
func ReferencePackCachePath(dir string) string {
	return filepath.Join(dir, "reference-pack-"+ReferencePackSHA256+".yaml")
}

// SDEGroupsCachePath is where the extracted SDE Group matrix is cached under
// dir.
func SDEGroupsCachePath(dir string) string {
	return filepath.Join(dir, "sde-groups-"+SDESHA256+".jsonl")
}

func (f *Fetcher) referencePackCachePath() string { return ReferencePackCachePath(f.cacheDir) }
func (f *Fetcher) sdeGroupsCachePath() string     { return SDEGroupsCachePath(f.cacheDir) }

// ReferencePack returns the Z-S reference pack's bytes.
func (f *Fetcher) ReferencePack(ctx context.Context) ([]byte, error) {
	return f.cached(ctx, f.referencePackCachePath(), f.referencePackURL, f.referencePackSHA)
}

// SDEGroups returns the SDE's groups.jsonl bytes, extracting them from the
// pinned archive on a cache miss.
func (f *Fetcher) SDEGroups(ctx context.Context) ([]byte, error) {
	if body, ok, err := readCache(f.sdeGroupsCachePath()); err != nil {
		return nil, err
	} else if ok {
		return body, nil
	}

	archive, err := f.download(ctx, f.sdeURL, f.sdeSHA)
	if err != nil {
		return nil, err
	}
	groups, err := extractGroupsJSONL(archive)
	if err != nil {
		return nil, err
	}
	if err := writeCache(f.sdeGroupsCachePath(), groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (f *Fetcher) cached(ctx context.Context, cachePath, url, sha string) ([]byte, error) {
	if body, ok, err := readCache(cachePath); err != nil {
		return nil, err
	} else if ok {
		return body, nil
	}

	body, err := f.download(ctx, url, sha)
	if err != nil {
		return nil, err
	}
	if err := writeCache(cachePath, body); err != nil {
		return nil, err
	}
	return body, nil
}

// download fetches url and returns its body only if it hashes to sha.
func (f *Fetcher) download(ctx context.Context, url, sha string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	response, err := f.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", url, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	sum := sha256.Sum256(body)
	if got := hex.EncodeToString(sum[:]); got != sha {
		return nil, fmt.Errorf("fetch %s: checksum mismatch: got %s, want %s", url, got, sha)
	}
	return body, nil
}

// extractGroupsJSONL pulls groups.jsonl out of the SDE archive. The archive's
// member paths vary across SDE builds, so the file is found by base name.
func extractGroupsJSONL(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("read SDE archive: %w", err)
	}
	for _, file := range reader.File {
		if path.Base(file.Name) != "groups.jsonl" {
			continue
		}
		contents, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s in SDE archive: %w", file.Name, err)
		}
		defer contents.Close()
		body, err := io.ReadAll(contents)
		if err != nil {
			return nil, fmt.Errorf("read %s in SDE archive: %w", file.Name, err)
		}
		return body, nil
	}
	return nil, errors.New("SDE archive contains no groups.jsonl")
}

func readCache(path string) ([]byte, bool, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read cache %s: %w", path, err)
	}
	return body, true, nil
}

// writeCache writes atomically, so a failed or interrupted fetch cannot leave
// a half-written artefact for the next run to trust.
func writeCache(path string, body []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create cache directory %s: %w", dir, err)
	}
	temp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("write cache %s: %w", path, err)
	}
	tempName := temp.Name()
	if _, err := temp.Write(body); err != nil {
		temp.Close()
		os.Remove(tempName)
		return fmt.Errorf("write cache %s: %w", path, err)
	}
	if err := temp.Close(); err != nil {
		os.Remove(tempName)
		return fmt.Errorf("write cache %s: %w", path, err)
	}
	if err := os.Rename(tempName, path); err != nil {
		os.Remove(tempName)
		return fmt.Errorf("write cache %s: %w", path, err)
	}
	return nil
}
