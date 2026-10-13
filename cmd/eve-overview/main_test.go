package main

import (
	"bytes"
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
	os.Exit(m.Run())
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

// cliReferencePack is a minimal stand-in for the cached Z-S reference pack.
// The CLI seeds it into the cache so tests never reach the network, and the
// Presets mirror the ones the real roster binds: the leading Collidables Preset
// proves the pack rides into the emitted profile, "✜ PvX: Exploring" is the
// authored exploration Preset's copyOf base, and the rest satisfy the Tabs'
// list and bracket bindings. A harness that seeded only some of them would make
// the CLI smoke tests fail on a missing binding rather than on the behaviour
// under test.
const cliReferencePack = `presets:
  - - "<color=0xFFFF6666>✜ = CovOps: Collidables</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - []
      - - groups
        - - 6
          - 1211
  - - "<color=0xFFFF6666>✜ PvX: Exploring</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - - 11
      - - groups
        - - 7
  - - "<color=0xFFFFFFFF>✈ Travel (-Ship +Sta)</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - - 11
      - - groups
        - - 10
  - - "<color=0xFF66FF66>➲ Extra: Podsaver (!)</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - []
      - - groups
        - - 29
  - - "<color=0xFF66CCFF>✪ = Friendly: Fleet</color>"
    - - - alwaysShownStates
        - - 11
      - - filteredStates
        - - 9
      - - groups
        - - 25
  - - "<color=0xFFFF6666>✜ PvX: Mission Filter (+LCO +LCS)</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - - 11
      - - groups
        - - 226
  - - "<color=0xFFFF6666>✜ PvX: PI</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - - 11
      - - groups
        - - 8
  - - "<color=0xFFFF6666>✜ PvX: Salvage/Loot</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - - 11
      - - groups
        - - 186
  - - "⌘ Brackets: Combat/Targets/Friendly (+Dro)"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - []
      - - groups
        - - 25
  - - "⌘ Brackets: Warpout (!)"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - []
      - - groups
        - - 29
`

// seedReferencePack writes a reference pack into the checksum-addressed cache
// path, exactly as a warm cache would hold it after a verified fetch.
func seedReferencePack(t *testing.T, cacheDir, pack string) {
	t.Helper()
	cachePath := reference.ReferencePackCachePath(cacheDir)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatalf("create cache dir: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte(pack), 0o644); err != nil {
		t.Fatalf("seed reference pack cache: %v", err)
	}
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
	seedReferencePack(t, cacheDir, cliReferencePack)

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
	seedReferencePack(t, cacheDir, cliReferencePack)

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

func TestGenerateFailsLoudlyOnUnresolvableGroup(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	// The cache holds the real reference data: "Dyson Sphere" is not in it.
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))

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
