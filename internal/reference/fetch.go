// Package reference fetches the two pinned reference artefacts the generator
// reads — the Z-S reference pack and the SDE Group matrix — and caches them in
// a user cache directory. The reference pack is verified against a pinned
// SHA-256 on download and again on every cache read, so neither a bad fetch
// nor a corrupted or tampered cache is ever fed to the generator: a mismatch
// fails generation. The SDE cache holds the extracted groups.jsonl, whose
// bytes the archive's pin does not describe, so that read is not re-checked.
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

// ReferencePack returns the Z-S reference pack's bytes. A cached pack is
// verified against the pin on every read, so a corrupted or tampered cache
// fails generation rather than feeding the generator unverified data.
func (f *Fetcher) ReferencePack(ctx context.Context) ([]byte, error) {
	return f.cached(ctx, ReferencePackCachePath(f.cacheDir), f.referencePackURL, f.referencePackSHA, nil)
}

// SDEGroups returns the SDE's groups.jsonl bytes, extracting them from the
// pinned archive on a cache miss.
func (f *Fetcher) SDEGroups(ctx context.Context) ([]byte, error) {
	return f.cached(ctx, SDEGroupsCachePath(f.cacheDir), f.sdeURL, f.sdeSHA, extractGroupsJSONL)
}

// cached is the read-cache → download → write-cache spine both artefacts
// share; extract is their only difference. It is nil when the cache holds the
// pinned artefact itself (the reference pack), and a function when the cache
// holds an artefact derived from the pinned archive (the SDE's extracted
// groups.jsonl).
//
// A cache holding the pinned artefact is re-verified on every read: a mismatch
// fails loudly rather than feeding the generator tampered data, and is never
// silently re-downloaded over — the checksum-addressed path means a mismatch
// there is local corruption or tampering the caller has to see. A derived
// cache carries no pin of its own and is not checked on read; it can only have
// been written by a verified download plus a deterministic extraction.
func (f *Fetcher) cached(ctx context.Context, cachePath, url, sha string, extract func([]byte) ([]byte, error)) ([]byte, error) {
	if body, ok, err := readCache(cachePath); err != nil {
		return nil, err
	} else if ok {
		if extract == nil {
			if err := verifyChecksum(cachePath, body, sha); err != nil {
				return nil, err
			}
		}
		return body, nil
	}

	body, err := f.download(ctx, url, sha)
	if err != nil {
		return nil, err
	}
	if extract != nil {
		body, err = extract(body)
		if err != nil {
			return nil, err
		}
	}
	if err := writeCache(cachePath, body); err != nil {
		return nil, err
	}
	return body, nil
}

// verifyChecksum returns an error unless body hashes to want. source names
// where the bytes came from — a URL or a cache path — so the failure says
// which artefact disagreed with its pin.
func verifyChecksum(source string, body []byte, want string) error {
	sum := sha256.Sum256(body)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", source, got, want)
	}
	return nil
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
	if err := verifyChecksum(url, body, sha); err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
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
