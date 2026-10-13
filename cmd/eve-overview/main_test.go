package main

import (
	"fmt"
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

func TestGenerateWritesProfileFromSourceTree(t *testing.T) {
	sourceDir := repoSourceDir(t)
	cacheDir := t.TempDir()
	seedReferenceData(t, cacheDir, readSource(t, sourceDir, "groups.yaml"))

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
