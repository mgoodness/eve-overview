// This file guards the release contract: the pair of workflows that cut a
// version and publish a tagged profile, the release-please config that decides
// the tag's shape, and the three checked-in artefacts a release consumer
// receives beside the profile — the third-party notices, the import steps, and
// the notes template. Like the CI workflow, each is a declarative artefact, so
// the tests observe it the way its consumer does: by parsing the YAML or JSON,
// or reading the Markdown.
package ci

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mgoodness/eve-overview/internal/reference"
)

var (
	releasePleaseWorkflowPath = filepath.Join("..", ".github", "workflows", "release-please.yml")
	releaseWorkflowPath       = filepath.Join("..", ".github", "workflows", "release.yml")
	releasePleaseConfigPath   = filepath.Join("..", "release-please-config.json")
	releasePleaseManifestPath = filepath.Join("..", ".release-please-manifest.json")
	releaseNotesPath          = filepath.Join("..", "docs", "release-notes-template.md")
	noticesPath               = filepath.Join("..", "THIRD_PARTY_NOTICES.md")
	installDocPath            = filepath.Join("..", "docs", "install.md")
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

// stepWithUses returns the first step in a job whose `uses:` starts with
// prefix, or nil. It names a step the way GitHub Actions resolves it.
func stepWithUses(t *testing.T, job *yaml.Node, prefix string) *yaml.Node {
	t.Helper()
	steps := nodeValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		t.Fatalf("job has no steps sequence")
	}
	for _, step := range steps.Content {
		if ref := nodeValue(step, "uses"); ref != nil && strings.HasPrefix(ref.Value, prefix) {
			return step
		}
	}
	return nil
}

// TestReleasePleaseWorkflowRunsOnMain pins where releases are cut: pushes to
// the default branch, which is the only place Conventional Commits accumulate.
func TestReleasePleaseWorkflowRunsOnMain(t *testing.T) {
	on := nodeValue(loadWorkflowAt(t, releasePleaseWorkflowPath), "on")
	if on == nil {
		t.Fatal(`release-please workflow has no "on" trigger block`)
	}
	push := nodeValue(on, "push")
	if push == nil {
		t.Fatal("release-please workflow does not run on push")
	}
	branches := nodeValue(push, "branches")
	if branches == nil {
		t.Fatal("release-please workflow's push trigger does not filter branches")
	}
	for _, branch := range branches.Content {
		if branch.Value == "main" {
			return
		}
	}
	t.Errorf("release-please workflow does not run on main")
}

// TestReleasePleaseWorkflowMintsAnAppToken is the loop-trap guard: the release
// PR and tag must be pushed by a GitHub App, because a GITHUB_TOKEN-authored
// push cannot trigger the release workflow that attaches the profile. The
// token must also be scoped down to just contents and pull requests.
func TestReleasePleaseWorkflowMintsAnAppToken(t *testing.T) {
	jobs := jobNodes(t, loadWorkflowAt(t, releasePleaseWorkflowPath))
	job := nodeValue(jobs, "release-please")
	if job == nil {
		t.Fatal(`release-please workflow has no job with the id "release-please"`)
	}

	tokenStep := stepWithUses(t, job, "actions/create-github-app-token@")
	if tokenStep == nil {
		t.Fatal("release-please workflow does not mint a GitHub App token")
	}
	if id := nodeValue(tokenStep, "id"); id == nil || id.Value != "app-token" {
		t.Errorf("App-token step id = %v, want app-token so later steps can reference it", id)
	}
	with := nodeValue(tokenStep, "with")
	if got := nodeValue(with, "client-id"); got == nil || !strings.Contains(got.Value, "RELEASE_PLEASE_APP_CLIENT_ID") {
		t.Errorf("App-token client-id = %v, want the RELEASE_PLEASE_APP_CLIENT_ID variable", got)
	}
	if got := nodeValue(with, "private-key"); got == nil || !strings.Contains(got.Value, "RELEASE_PLEASE_APP_PRIVATE_KEY") {
		t.Errorf("App-token private-key = %v, want the RELEASE_PLEASE_APP_PRIVATE_KEY secret", got)
	}
	if got := nodeValue(with, "permission-contents"); got == nil || got.Value != "write" {
		t.Errorf("App-token permission-contents = %v, want write (release-please pushes the tag)", got)
	}
	if got := nodeValue(with, "permission-pull-requests"); got == nil || got.Value != "write" {
		t.Errorf("App-token permission-pull-requests = %v, want write (release-please opens the release PR)", got)
	}

	releasePlease := stepWithUses(t, job, "googleapis/release-please-action@")
	if releasePlease == nil {
		t.Fatal("release-please workflow does not run release-please-action")
	}
	rpWith := nodeValue(releasePlease, "with")
	if got := nodeValue(rpWith, "token"); got == nil || !strings.Contains(got.Value, "steps.app-token.outputs.token") {
		t.Errorf("release-please token = %v, want the minted App token, or the loop-trap reopens", got)
	}
}

// TestReleasePleaseWorkflowKeepsTheDefaultTokenReadOnly is the least-privilege
// half of the App-token design: the job's own GITHUB_TOKEN needs nothing.
func TestReleasePleaseWorkflowKeepsTheDefaultTokenReadOnly(t *testing.T) {
	root := loadWorkflowAt(t, releasePleaseWorkflowPath)
	permissions := nodeValue(root, "permissions")
	if permissions == nil {
		t.Fatal("release-please workflow declares no permissions, so the job inherits the repository default token scope")
	}
	if len(permissions.Content) != 0 {
		t.Errorf("release-please workflow default permissions have %d entries, want none — the App token carries them", len(permissions.Content))
	}
}

// TestReleasePleaseConfigKeepsTheTagShape is the tag-trap guard: a component
// name would make the tag `<component>-vX.Y.Z`, which no downstream trigger
// watches for.
func TestReleasePleaseConfigKeepsTheTagShape(t *testing.T) {
	var config struct {
		IncludeComponentInTag *bool `json:"include-component-in-tag"`
		Packages              map[string]struct {
			PackageName *string `json:"package-name"`
		} `json:"packages"`
	}
	if err := json.Unmarshal([]byte(readArtefact(t, releasePleaseConfigPath)), &config); err != nil {
		t.Fatalf("parse %s: %v", releasePleaseConfigPath, err)
	}
	pkg, ok := config.Packages["."]
	if !ok {
		t.Fatalf("%s does not configure the root package", releasePleaseConfigPath)
	}
	if pkg.PackageName != nil {
		t.Errorf("%s sets package-name, which prefixes the tag with a component", releasePleaseConfigPath)
	}
	if config.IncludeComponentInTag == nil || *config.IncludeComponentInTag {
		t.Errorf("%s must set include-component-in-tag to false so tags are v<version>", releasePleaseConfigPath)
	}

	var manifest map[string]string
	if err := json.Unmarshal([]byte(readArtefact(t, releasePleaseManifestPath)), &manifest); err != nil {
		t.Fatalf("parse %s: %v", releasePleaseManifestPath, err)
	}
	if manifest["."] == "" {
		t.Errorf("%s does not pin the root package's current version", releasePleaseManifestPath)
	}
}

// TestReleaseWorkflowTriggersOnPublishedReleases pins the downstream trigger:
// release-please publishes the release, and this workflow reacts to it rather
// than to a hand-pushed tag.
func TestReleaseWorkflowTriggersOnPublishedReleases(t *testing.T) {
	on := nodeValue(loadWorkflowAt(t, releaseWorkflowPath), "on")
	if on == nil {
		t.Fatal(`release workflow has no "on" trigger block`)
	}
	release := nodeValue(on, "release")
	if release == nil {
		t.Fatal("release workflow does not run on release events")
	}
	types := nodeValue(release, "types")
	if types == nil {
		t.Fatal("release workflow's release trigger does not filter types")
	}
	for _, action := range types.Content {
		if action.Value == "published" {
			return
		}
	}
	t.Errorf("release workflow does not run on published releases")
}

// TestReleaseWorkflowPublishesTheProfileAndItsChecksums is the acceptance guard
// for the asset half: generate the profile, checksum it, attach both to the
// release release-please created, and append the attribution and the pins to
// its notes.
func TestReleaseWorkflowPublishesTheProfileAndItsChecksums(t *testing.T) {
	jobs := jobNodes(t, loadWorkflowAt(t, releaseWorkflowPath))
	job := nodeValue(jobs, "attach-profile")
	if job == nil {
		t.Fatal(`release workflow has no job with the id "attach-profile"`)
	}
	commands := runCommands(t, job)

	var generated, checksummed, uploaded, attributed bool
	for _, command := range commands {
		switch {
		case strings.Contains(command, "generate") && strings.Contains(command, "overview.yaml"):
			generated = true
		case strings.Contains(command, "sha256sum") && strings.Contains(command, "SHA256SUMS"):
			checksummed = true
		case strings.Contains(command, "gh release upload"):
			uploaded = true
			for _, want := range []string{"overview.yaml", "SHA256SUMS"} {
				if !strings.Contains(command, want) {
					t.Errorf("release upload command %q does not attach %q", command, want)
				}
			}
		case strings.Contains(command, "gh release edit") && strings.Contains(command, releaseNotesRepoPath):
			attributed = true
		}
	}
	if !generated {
		t.Errorf("no step runs the generator to overview.yaml; commands: %q", commands)
	}
	if !checksummed {
		t.Errorf("no step writes SHA256SUMS; commands: %q", commands)
	}
	if !uploaded {
		t.Errorf("no step runs `gh release upload`; commands: %q", commands)
	}
	if !attributed {
		t.Errorf("no step edits the notes from %s; commands: %q", releaseNotesRepoPath, commands)
	}
}

// TestReleaseWorkflowCanWriteReleases pins the token scope: attaching assets
// and editing notes needs contents: write, which is the whole reason this
// workflow exists.
func TestReleaseWorkflowCanWriteReleases(t *testing.T) {
	permissions := nodeValue(loadWorkflowAt(t, releaseWorkflowPath), "permissions")
	if permissions == nil {
		t.Fatal("release workflow declares no permissions, so the job inherits the repository default token scope")
	}
	if contents := nodeValue(permissions, "contents"); contents == nil || contents.Value != "write" {
		t.Errorf("permissions.contents = %v, want write", contents)
	}
}

// TestReleaseWorkflowsActionsArePinnedToCommitSHAs mirrors the CI pin guard: a
// moving tag must not change what cuts or publishes a release.
func TestReleaseWorkflowsActionsArePinnedToCommitSHAs(t *testing.T) {
	for _, path := range []string{releaseWorkflowPath, releasePleaseWorkflowPath} {
		walk(loadWorkflowAt(t, path), func(node *yaml.Node) {
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
					t.Errorf("%s: action %q is not pinned to a full commit SHA", path, ref.Value)
					continue
				}
				if !strings.Contains(ref.LineComment, "# v") {
					t.Errorf("%s: action %q has no trailing version comment, so its pin is unreadable", path, ref.Value)
				}
			}
		})
	}
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
