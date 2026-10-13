package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mgoodness/eve-overview/internal/reference"
)

// binaryPath is the CLI built once for the smoke tests, so the tests exercise
// the real process boundary — exit codes, stderr, written file — rather than
// an in-process entry point.
var binaryPath string

// referenceUnavailable records that a smoke test could not read the pinned
// reference pack. TestMain fails a non-short run that skipped on it, so an
// offline cold cache cannot masquerade as a passing CLI.
var referenceUnavailable bool

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "eve-overview-cli")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create temp dir:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "eve-overview")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "build CLI:", err)
		os.Exit(1)
	}
	code := m.Run()
	if code == 0 && referenceUnavailable && !testing.Short() {
		fmt.Fprintln(os.Stderr, "CLI smoke tests could not read the pinned reference pack and this was not a -short run; refusing to report success")
		code = 1
	}
	os.Exit(code)
}

// seedReferenceData writes a Group matrix into a cache directory so the CLI
// never reaches the network in a test. The cache is keyed by the pinned SDE
// checksum, which the CLI trusts once populated.
func seedReferenceData(t *testing.T, cacheDir string, groupsYAML string) {
	t.Helper()
	var doc struct {
		GroupSets map[string][]any `yaml:"groupSets"`
	}
	if err := yaml.Unmarshal([]byte(groupsYAML), &doc); err != nil {
		t.Fatalf("decode seeded groups.yaml: %v", err)
	}

	var lines []string
	seen := map[string]bool{}
	id := 1000
	for _, refs := range doc.GroupSets {
		for _, ref := range refs {
			name, ok := ref.(string)
			if !ok || seen[name] {
				continue
			}
			seen[name] = true
			lines = append(lines, fmt.Sprintf(`{"_key":%d,"name":{"en":%q}}`, id, name))
			id++
		}
	}

	cachePath := reference.SDEGroupsCachePath(cacheDir)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("create cache dir: %v", err)
	}
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(cachePath, []byte(body), 0o644); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
}

func repoSourceDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "source"))
	if err != nil {
		t.Fatalf("resolve source dir: %v", err)
	}
	return dir
}

func readSource(t *testing.T, dir, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(body)
}

// realReferencePack returns the pinned Z-S reference pack's bytes through the
// same cache-and-checksum layer the CLI uses. The smoke tests seed the
// checksum-addressed cache from it, so the CLI is exercised with the bytes the
// pin actually describes rather than an arbitrary fixture that the read-time
// checksum would rightly reject. A warm cache serves it with no network I/O;
// ok is false when the pack can be neither read nor fetched, which the tests
// skip on — loudly, through TestMain.
func realReferencePack(t *testing.T) ([]byte, bool) {
	t.Helper()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Logf("no user cache directory: %v", err)
		return nil, false
	}
	pack, err := reference.New(filepath.Join(cacheDir, "eve-overview")).ReferencePack(context.Background())
	if err != nil {
		t.Logf("reference pack unavailable: %v", err)
		return nil, false
	}
	return pack, true
}

// seedReferencePack writes the pinned reference pack into the checksum-addressed
// cache path, exactly as a warm cache would hold it after a verified fetch.
func seedReferencePack(t *testing.T, cacheDir string, pack []byte) {
	t.Helper()
	cachePath := reference.ReferencePackCachePath(cacheDir)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("create cache dir: %v", err)
	}
	if err := os.WriteFile(cachePath, pack, 0o644); err != nil {
		t.Fatalf("seed reference pack cache: %v", err)
	}
}

// seedPinnedReferencePack seeds the CLI's cache with the real pinned reference
// pack, so a smoke test measures the CLI rather than a fixture the read-time
// checksum cannot accept. It skips when the pack is unavailable, recording
// that for TestMain so a cold offline cache does not look like a green CLI.
func seedPinnedReferencePack(t *testing.T, cacheDir string) {
	t.Helper()
	pack, ok := realReferencePack(t)
	if !ok {
		referenceUnavailable = true
		t.Skip("reference pack unavailable (offline with a cold cache?); skipping CLI smoke test — run without -short to require it")
	}
	seedReferencePack(t, cacheDir, pack)
}

// recordingTransport answers every request with a canned body (or error) and
// records the URLs, so a test can prove the CLI made no request at all.
type recordingTransport struct {
	body     []byte
	err      error
	requests []string
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.requests = append(t.requests, req.URL.String())
	if t.err != nil {
		return nil, t.err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(t.body)),
		Header:     http.Header{},
	}, nil
}

// withDefaultClient points the process's default HTTP client at transport for
// the duration of the test, so the CLI's fetcher — which captures
// http.DefaultClient — can be exercised without the network.
func withDefaultClient(t *testing.T, transport http.RoundTripper) {
	t.Helper()
	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: transport}
	t.Cleanup(func() { http.DefaultClient = original })
}

// TestGenerateWritesProfileFromSourceTree proves the CLI writes a profile that
// includes the cached reference library, not just the authored source.
func TestGenerateWritesProfileFromSourceTree(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))
	seedPinnedReferencePack(t, cacheDir)

	outPath := filepath.Join(t.TempDir(), "profile.yaml")
	cmd := exec.Command(binaryPath, "generate",
		"-source", sourceDir,
		"-cache", cacheDir,
		"-out", outPath,
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generate failed: %v\nstderr: %s", err, stderr.String())
	}

	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("profile was not written: %v", err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatalf("emitted profile is not valid YAML: %v", err)
	}
	for _, key := range []string{"presets", "tabSetup", "overviewColumns", "userSettings"} {
		if _, ok := doc[key]; !ok {
			t.Errorf("emitted profile is missing %q", key)
		}
	}

	names := map[string]bool{}
	presets, _ := doc["presets"].([]any)
	for _, raw := range presets {
		name, _ := raw.([]any)[0].(string)
		names[name] = true
	}
	if !names["<color=0xFFFF6666>✜ = CovOps: Collidables</color>"] {
		t.Errorf("emitted profile does not carry the cached reference pack's Preset; got %v", names)
	}
	if !names["<color=0xFFFF6666>Engage</color>"] {
		t.Errorf("emitted profile does not carry the authored Preset; got %v", names)
	}
}

// TestGenerateSucceedsOfflineFromCache proves generation is a pure read of the
// warm cache: with both artefacts cached, the CLI makes no network request at
// all — the transport fails every request, so a request would be fatal.
func TestGenerateSucceedsOfflineFromCache(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))
	seedPinnedReferencePack(t, cacheDir)

	transport := &recordingTransport{err: errors.New("network disabled")}
	withDefaultClient(t, transport)

	var stdout, stderr strings.Builder
	code := run([]string{"generate",
		"-source", sourceDir,
		"-cache", cacheDir,
		"-out", filepath.Join(t.TempDir(), "profile.yaml"),
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("generate exited %d, want 0\nstderr: %s", code, stderr.String())
	}
	if len(transport.requests) != 0 {
		t.Errorf("generate made %d network requests from a warm cache, want 0: %v", len(transport.requests), transport.requests)
	}
}

// TestGenerateFailsOnReferencePackChecksumMismatch proves a fetch whose bytes
// do not hash to the pinned checksum aborts generation rather than feeding the
// generator unverified data. The SDE cache is warm, so the reference-pack
// fetch is the only request and the mismatch is its checksum.
func TestGenerateFailsOnReferencePackChecksumMismatch(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))

	// No reference pack is cached, so the CLI must fetch; the transport serves
	// bytes that cannot match the pinned SHA-256.
	transport := &recordingTransport{body: []byte("presets: []\n")}
	withDefaultClient(t, transport)

	var stderr strings.Builder
	code := run([]string{"generate",
		"-source", sourceDir,
		"-cache", cacheDir,
		"-out", filepath.Join(t.TempDir(), "profile.yaml"),
	}, io.Discard, &stderr)
	if code == 0 {
		t.Fatal("generate exited zero despite a reference-pack checksum mismatch")
	}
	if !strings.Contains(stderr.String(), "checksum mismatch") {
		t.Errorf("stderr does not report a checksum mismatch: %q", stderr.String())
	}
}

// TestGenerateFailsOnCorruptedReferencePackCache proves the warm-cache path
// enforces the pin too, not only the download: a cache holding bytes that do
// not hash to it aborts generation rather than feeding the generator tampered
// data. Both artefacts are cached, so a passing run would need no network — and
// the transport fails every request, so a silent re-download would be fatal.
func TestGenerateFailsOnCorruptedReferencePackCache(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))
	seedReferencePack(t, cacheDir, []byte("presets:\n  - - tampered\n"))

	transport := &recordingTransport{err: errors.New("network disabled")}
	withDefaultClient(t, transport)

	var stderr strings.Builder
	code := run([]string{"generate",
		"-source", sourceDir,
		"-cache", cacheDir,
		"-out", filepath.Join(t.TempDir(), "profile.yaml"),
	}, io.Discard, &stderr)
	if code == 0 {
		t.Fatal("generate exited zero from a corrupted reference-pack cache, want non-zero")
	}
	if !strings.Contains(stderr.String(), "checksum mismatch") {
		t.Errorf("stderr does not report a checksum mismatch: %q", stderr.String())
	}
	if len(transport.requests) != 0 {
		t.Errorf("generate made %d network requests after a cache mismatch, want 0: %v", len(transport.requests), transport.requests)
	}
}

func TestGenerateFailsLoudlyOnUnresolvableGroup(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	// The cache holds the real reference data: "Dyson Sphere" is not in it.
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))
	seedPinnedReferencePack(t, cacheDir)

	broken := t.TempDir()
	for _, name := range []string{"presets.yaml", "tabs.yaml", "appearance.yaml", "columns.yaml"} {
		if err := os.WriteFile(filepath.Join(broken, name), []byte(readSource(t, sourceDir, name)), 0o644); err != nil {
			t.Fatalf("copy %s: %v", name, err)
		}
	}
	groups := readSource(t, sourceDir, "groups.yaml") + "\n    - Dyson Sphere\n"
	if err := os.WriteFile(filepath.Join(broken, "groups.yaml"), []byte(groups), 0o644); err != nil {
		t.Fatalf("write groups.yaml: %v", err)
	}

	cmd := exec.Command(binaryPath, "generate",
		"-source", broken,
		"-cache", cacheDir,
		"-out", filepath.Join(t.TempDir(), "profile.yaml"),
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("generate exited zero with an unresolvable Group, want non-zero")
	}
	if !strings.Contains(stderr.String(), "Dyson Sphere") {
		t.Errorf("stderr does not name the offending Group: %q", stderr.String())
	}
}

// TestGenerateWritesAttributedProfile proves the file the CLI writes leads with
// the attribution and reproducibility header, and that the document still
// decodes. The emitted profile is a derivative of the GPLv3 Z-S pack, so the
// attribution has to travel with the artefact itself — not only live in the
// repository — and it must not stop the profile being a profile.
func TestGenerateWritesAttributedProfile(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))
	seedPinnedReferencePack(t, cacheDir)

	outPath := filepath.Join(t.TempDir(), "profile.yaml")
	cmd := exec.Command(binaryPath, "generate",
		"-source", sourceDir,
		"-cache", cacheDir,
		"-out", outPath,
	)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generate failed: %v\nstderr: %s", err, stderr.String())
	}

	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("profile was not written: %v", err)
	}

	header := attributionHeader()
	if !bytes.HasPrefix(body, []byte(header)) {
		t.Fatalf("written profile does not start with the attribution header:\n%s", body)
	}

	// The header comes from the CLI, not from overview.Generate: strip it and
	// the seam's bytes must be the document alone, leading with its first
	// profile key. A second header, or a document that begins with a comment,
	// would mean attribution had leaked into the seam.
	profile := bytes.TrimPrefix(body, []byte(header))
	if !bytes.HasPrefix(profile, []byte("backgroundOrder:")) {
		t.Errorf("the bytes after the header are not the profile document itself: %q", firstLine(profile))
	}

	// The header must actually say who and what: the pack, its two
	// maintainers, its lineage, its licence, the SDE's terms, and the pins
	// that make the artefact reproducible from the repository plus the
	// cached reference data.
	for _, want := range []string{
		"https://github.com/mgoodness/eve-overview",
		"Z-S Overview",
		"Zirio",
		"Deuce Syundai",
		"SaraShawa",
		"Sarah Schneider",
		"GNU GPLv3",
		"https://github.com/Arziel1992/Z-S-Overview-Pack",
		"Fenris",
		"docs/install.md",
		reference.ReferencePackURL,
		reference.ReferencePackSHA256,
		reference.SDEURL,
		reference.SDESHA256,
		reference.SDEBuild,
	} {
		if !strings.Contains(header, want) {
			t.Errorf("attribution header does not mention %q", want)
		}
	}

	// The header is a YAML comment block, so the document it leads decodes to
	// a valid profile unchanged. A comment that broke decoding would make the
	// attribution cost the artefact.
	var doc map[string]any
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatalf("attributed profile is not valid YAML: %v", err)
	}
	if _, ok := doc["presets"]; !ok {
		t.Errorf("attributed profile does not decode to a profile document; got keys %v", keys(doc))
	}
}

// keys returns a map's keys, for readable failure output.
func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}

// firstLine returns body's first line, for a failure message about the bytes
// that did not look like a profile document.
func firstLine(body []byte) string {
	line, _, _ := bytes.Cut(body, []byte("\n"))
	return string(line)
}
