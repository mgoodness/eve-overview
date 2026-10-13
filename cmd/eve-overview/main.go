// Command eve-overview generates an EVE Online overview profile from the
// source-of-truth tree in this repository.
//
// The command is deliberately thin: parse arguments, make sure the reference
// data is cached, hand the decoded source to the generator, write the result.
// All of the generator's behaviour is testable through the overview package's
// seam; the CLI has a smoke test and nothing more.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mgoodness/eve-overview/internal/reference"
	"github.com/mgoodness/eve-overview/overview"
)

const usage = `eve-overview generates an EVE Online overview profile.

Usage:
  eve-overview generate [flags]
  eve-overview upstream

Flags for generate:
  -source dir   source-of-truth directory (default "source")
  -out file     profile to write (default "overview.yaml")
  -cache dir    reference cache directory (default: user cache directory)

Run upstream to report whether the pinned SDE build and Z-S reference pack are
still the newest upstream. It is read-only and exits non-zero on drift.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "generate":
		return runGenerate(args[1:], stdout, stderr)
	case "upstream":
		return runUpstream(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n%s", args[0], usage)
		return 2
	}
}

func runGenerate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("generate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sourceDir := flags.String("source", "source", "directory holding the source-of-truth files")
	outPath := flags.String("out", "overview.yaml", "file to write the overview profile to")
	cacheDir := flags.String("cache", "", "reference cache directory (default: user cache directory)")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	cache := *cacheDir
	if cache == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			fmt.Fprintf(stderr, "find user cache directory: %v\n", err)
			return 1
		}
		cache = filepath.Join(userCache, "eve-overview")
	}

	source, err := loadSource(*sourceDir)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	// Fetching is a no-op once cached, so a second run makes no network
	// request. Both artefacts are pinned and checksum-verified by the fetcher;
	// a mismatch aborts generation rather than feeding it unverified data.
	fetcher := reference.New(cache)
	sdeGroups, err := fetcher.SDEGroups(context.Background())
	if err != nil {
		fmt.Fprintf(stderr, "fetch SDE Group matrix: %v\n", err)
		return 1
	}
	referencePack, err := fetcher.ReferencePack(context.Background())
	if err != nil {
		fmt.Fprintf(stderr, "fetch reference pack: %v\n", err)
		return 1
	}
	profile, err := overview.Generate(source, overview.Options{
		SDEGroups:     sdeGroups,
		ReferencePack: referencePack,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	// The written file leads with the attribution and reproducibility
	// header; the profile itself is the comment block's tail. Leading with
	// the comments keeps Generate's output a pure profile while the artefact
	// on disk carries its own provenance.
	document := append([]byte(attributionHeader()), profile...)
	if err := os.WriteFile(*outPath, document, 0o644); err != nil {
		fmt.Fprintf(stderr, "write %s: %v\n", *outPath, err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s\n", *outPath)
	return 0
}

// runUpstream reports whether the pinned reference data is still the newest
// upstream. It exits non-zero on drift so a scheduled workflow can open an
// issue; bumping a pin stays a human decision.
func runUpstream(stdout, stderr io.Writer) int {
	statuses, err := reference.CheckUpstream(context.Background(), nil)
	if err != nil {
		fmt.Fprintf(stderr, "check upstream: %v\n", err)
		return 1
	}

	stale := false
	for _, status := range statuses {
		if !status.Current {
			stale = true
		}
	}
	if !stale {
		fmt.Fprintln(stdout, "All pinned reference data is current.")
		return 0
	}

	fmt.Fprintln(stdout, "The pinned reference data has moved upstream. Bumping a pin is a deliberate")
	fmt.Fprintln(stdout, "regeneration, not an automatic upgrade: update `internal/reference/fetch.go`,")
	fmt.Fprintln(stdout, "run `go test ./...`, and let release-please cut the version.")
	fmt.Fprintln(stdout)
	for _, status := range statuses {
		marker := "up to date"
		if !status.Current {
			marker = "**UPDATE AVAILABLE**"
		}
		fmt.Fprintf(stdout, "- **%s** — pinned `%s`, latest `%s` — %s", status.Name, status.Pinned, status.Latest, marker)
		if status.Detail != "" {
			fmt.Fprintf(stdout, " (%s)", status.Detail)
		}
		fmt.Fprintln(stdout)
	}
	return 1
}

// loadSource reads the five source-of-truth files as raw bytes. Decoding them
// is the generator's job.
func loadSource(dir string) (overview.Source, error) {
	var source overview.Source
	files := []struct {
		name string
		dest *[]byte
	}{
		{"groups.yaml", &source.Groups},
		{"presets.yaml", &source.Presets},
		{"tabs.yaml", &source.Tabs},
		{"appearance.yaml", &source.Appearance},
		{"columns.yaml", &source.Columns},
	}
	for _, file := range files {
		path := filepath.Join(dir, file.name)
		body, err := os.ReadFile(path)
		if err != nil {
			return overview.Source{}, fmt.Errorf("read source %s: %w", path, err)
		}
		*file.dest = body
	}
	return source, nil
}
