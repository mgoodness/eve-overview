// This file guards the release contract: the workflow that publishes a tagged
// profile, and the three checked-in artefacts a release consumer receives
// beside it — the third-party notices, the import steps, and the notes
// template. Like the CI workflow, each is a declarative artefact, so the tests
// observe it the way its consumer does: by parsing the YAML or reading the
// Markdown.
package ci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mgoodness/eve-overview/internal/reference"
)

var (
	releaseWorkflowPath = filepath.Join("..", ".github", "workflows", "release.yml")
	releaseNotesPath    = filepath.Join("..", "docs", "release-notes-template.md")
	noticesPath         = filepath.Join("..", "THIRD_PARTY_NOTICES.md")
	installDocPath      = filepath.Join("..", "docs", "install.md")
)

// releaseNotesRepoPath is the notes template's path as the workflow names it:
// relative to the repository root, which is the working directory a workflow
// step runs in.
const releaseNotesRepoPath = "docs/release-notes-template.md"

func readArtefact(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

// attributionPhrases is the attribution a release consumer must receive: the
// pack, its two maintainers, its lineage, its licence, the SDE's terms, and the
// pins that make the profile reproducible. The CLI's test asserts these facts
// against the profile header too, so the same facts travel in every artefact.
func attributionPhrases() []string {
	return []string{
		"Z-S Overview",
		"Zirio",
		"Deuce Syundai",
		"SaraShawa",
		"Sarah Schneider",
		"GNU GPLv3",
		"https://github.com/Arziel1992/Z-S-Overview-Pack",
		"Fenris",
		reference.ReferencePackURL,
		reference.ReferencePackSHA256,
		reference.SDEURL,
		reference.SDESHA256,
		reference.SDEBuild,
	}
}

func assertMentions(t *testing.T, path, body string, wants []string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Errorf("%s does not mention %q", path, want)
		}
	}
}

// releaseJob returns the release workflow's single job. The test names the
// job id so a rename is a visible, deliberate change rather than a silent one.
func releaseJob(t *testing.T) *yaml.Node {
	t.Helper()
	jobs := jobNodes(t, loadWorkflowAt(t, releaseWorkflowPath))
	job := nodeValue(jobs, "release")
	if job == nil {
		t.Fatal(`release workflow has no job with the id "release"`)
	}
	return job
}

func TestReleaseWorkflowTriggersOnVersionTags(t *testing.T) {
	on := nodeValue(loadWorkflowAt(t, releaseWorkflowPath), "on")
	if on == nil {
		t.Fatal(`release workflow has no "on" trigger block`)
	}
	push := nodeValue(on, "push")
	if push == nil {
		t.Fatal("release workflow does not run on push")
	}
	tags := nodeValue(push, "tags")
	if tags == nil {
		t.Fatal("release workflow's push trigger does not filter tags")
	}
	var patterns []string
	for _, tag := range tags.Content {
		patterns = append(patterns, tag.Value)
	}
	if !strings.Contains(strings.Join(patterns, ","), "v*") {
		t.Errorf("release workflow's tag filter = %v, want a v* pattern", patterns)
	}
}

func TestReleaseWorkflowPublishesTheProfileAndItsChecksums(t *testing.T) {
	commands := runCommands(t, releaseJob(t))

	var generated, checksummed, composed, published bool
	for _, command := range commands {
		switch {
		case strings.Contains(command, "generate") && strings.Contains(command, "overview.yaml"):
			generated = true
		case strings.Contains(command, "sha256sum") && strings.Contains(command, "SHA256SUMS"):
			checksummed = true
		case strings.Contains(command, releaseNotesRepoPath):
			composed = true
		case strings.Contains(command, "gh release create"):
			published = true
			for _, want := range []string{"overview.yaml", "SHA256SUMS", "--notes-file"} {
				if !strings.Contains(command, want) {
					t.Errorf("release command %q does not upload or reference %q", command, want)
				}
			}
		}
	}
	if !generated {
		t.Errorf("no step runs the generator to overview.yaml; commands: %q", commands)
	}
	if !checksummed {
		t.Errorf("no step writes SHA256SUMS; commands: %q", commands)
	}
	if !composed {
		t.Errorf("no step composes the notes from %s; commands: %q", releaseNotesRepoPath, commands)
	}
	if !published {
		t.Errorf("no step runs `gh release create`; commands: %q", commands)
	}
}

// TestReleaseWorkflowCanWriteReleases pins the token scope: creating a release
// needs contents: write, which is the whole reason this workflow exists.
func TestReleaseWorkflowCanWriteReleases(t *testing.T) {
	permissions := nodeValue(loadWorkflowAt(t, releaseWorkflowPath), "permissions")
	if permissions == nil {
		t.Fatal("release workflow declares no permissions, so the job inherits the repository default token scope")
	}
	if contents := nodeValue(permissions, "contents"); contents == nil || contents.Value != "write" {
		t.Errorf("permissions.contents = %v, want write", contents)
	}
}

// TestReleaseWorkflowActionsArePinnedToCommitSHAs mirrors the CI pin guard:
// a moving tag must not change what a release runs.
func TestReleaseWorkflowActionsArePinnedToCommitSHAs(t *testing.T) {
	walk(loadWorkflowAt(t, releaseWorkflowPath), func(node *yaml.Node) {
		if node.Kind != yaml.MappingNode {
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value != "uses" {
				continue
			}
			ref := node.Content[i+1]
			_, sha, found := strings.Cut(ref.Value, "@")
			if !found || !commitSHA.MatchString(sha) {
				t.Errorf("action %q is not pinned to a full commit SHA", ref.Value)
				continue
			}
			if !strings.Contains(ref.LineComment, "# v") {
				t.Errorf("action %q has no trailing version comment, so its pin is unreadable", ref.Value)
			}
		}
	})
}

// TestReleaseNotesTemplateCarriesAttributionAndInstallPointer is the acceptance
// guard that the published notes — not just the profile — attribute the pack
// and point at the import steps.
func TestReleaseNotesTemplateCarriesAttributionAndInstallPointer(t *testing.T) {
	body := readArtefact(t, releaseNotesPath)
	assertMentions(t, releaseNotesPath, body, attributionPhrases())
	if !strings.Contains(body, "docs/install.md") {
		t.Errorf("%s does not point at the import steps (docs/install.md)", releaseNotesPath)
	}
}

// TestThirdPartyNoticesCarryAttributionAndPins is the repository-side copy of
// the same facts: a reader who never downloads a release still finds the pack
// credited and its licence named.
func TestThirdPartyNoticesCarryAttributionAndPins(t *testing.T) {
	body := readArtefact(t, noticesPath)
	assertMentions(t, noticesPath, body, attributionPhrases())
}

// TestInstallDocDocumentsTheManualInClientImport covers the import steps:
// where the YAML goes on each platform, the in-client action that loads it, and
// the plain statement that nothing installs itself.
func TestInstallDocDocumentsTheManualInClientImport(t *testing.T) {
	body := readArtefact(t, installDocPath)
	for _, want := range []string{
		`%USERPROFILE%\Documents\EVE\Overview`,
		"~/Documents/EVE/Overview",
		"compatdata",
		"Overview Settings",
		"Import",
		"manual",
		"no installer",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("%s does not document %q", installDocPath, want)
		}
	}
}
