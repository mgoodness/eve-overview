// Package ci guards the CI workflow's contract: the properties that decide
// whether `go test ./...` in CI can be trusted as a required check. The
// workflow is a declarative artefact, so the test observes it the way GitHub
// Actions does — by parsing the YAML at .github/workflows/ci.yml.
//
// The package holds only tests; `go test ./...` exercises it wherever the
// workflow runs.
package ci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// workflowPath is relative to this package's directory, which is the working
// directory `go test` gives each test binary.
var workflowPath = filepath.Join("..", ".github", "workflows", "ci.yml")

func loadWorkflow(t *testing.T) *yaml.Node {
	t.Helper()
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read %s: %v", workflowPath, err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatalf("parse %s: %v", workflowPath, err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) == 0 {
		t.Fatalf("%s is not a YAML document", workflowPath)
	}
	return document.Content[0]
}

// nodeValue returns the value node for key in a mapping node, or nil when the
// mapping is absent or the key is missing. It compares raw key text, so it is
// immune to YAML 1.1 resolving `on` and friends as booleans.
func nodeValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// jobNodes returns the jobs mapping under a workflow root.
func jobNodes(t *testing.T, root *yaml.Node) *yaml.Node {
	t.Helper()
	jobs := nodeValue(root, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		t.Fatal("workflow has no jobs mapping")
	}
	return jobs
}

// runCommands returns every `run:` script in a job, in step order.
func runCommands(t *testing.T, job *yaml.Node) []string {
	t.Helper()
	steps := nodeValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		t.Fatalf("job has no steps sequence")
	}
	var commands []string
	for _, step := range steps.Content {
		if script := nodeValue(step, "run"); script != nil {
			commands = append(commands, script.Value)
		}
	}
	return commands
}

func TestWorkflowRunsOnPushAndPullRequest(t *testing.T) {
	root := loadWorkflow(t)
	on := nodeValue(root, "on")
	if on == nil {
		t.Fatal(`workflow has no "on" trigger block`)
	}
	for _, event := range []string{"push", "pull_request"} {
		if nodeValue(on, event) == nil {
			t.Errorf("workflow does not run on %s", event)
		}
	}
}

func TestTestJobIsAStableRequiredCheck(t *testing.T) {
	jobs := jobNodes(t, loadWorkflow(t))
	job := nodeValue(jobs, "test")
	if job == nil {
		t.Fatal(`workflow has no job with the id "test"`)
	}
	if got := nodeValue(job, "name"); got == nil || got.Value != "test" {
		t.Errorf(`job "test" name = %v, want the literal "test" so the check name cannot drift`, got)
	}
}

// TestWorkflowRunsTheFullTestSuiteWithoutShort is the acceptance guard for the
// conformance test's offline skip. The mode must be a full run (-count=1, no
// -short) so overview.TestMain turns an all-skipped run red, and verbose so the
// skip line is visible in the job log.
func TestWorkflowRunsTheFullTestSuiteWithoutShort(t *testing.T) {
	jobs := jobNodes(t, loadWorkflow(t))
	commands := runCommands(t, nodeValue(jobs, "test"))

	var testCommand string
	for _, command := range commands {
		if strings.Contains(command, "go test") {
			testCommand = command
		}
	}
	if testCommand == "" {
		t.Fatalf("job \"test\" has no step running `go test`; commands: %q", commands)
	}
	if !strings.Contains(testCommand, "./...") {
		t.Errorf("test command %q does not cover ./...", testCommand)
	}
	if !strings.Contains(testCommand, "-count=1") {
		t.Errorf("test command %q does not pass -count=1, so a cached result could stand in for a run", testCommand)
	}
	if strings.Contains(testCommand, "-short") {
		t.Errorf("test command %q passes -short, which disables the conformance guard", testCommand)
	}
	if !strings.Contains(testCommand, "-v") {
		t.Errorf("test command %q is not verbose, so a conformance skip would not appear in the log", testCommand)
	}
}

// TestTestStepCannotSwallowFailure pins the other half of "a failing test fails
// the workflow": the step must not opt out of GitHub's non-zero-exit failure.
func TestTestStepCannotSwallowFailure(t *testing.T) {
	jobs := jobNodes(t, loadWorkflow(t))
	job := nodeValue(jobs, "test")
	steps := nodeValue(job, "steps")
	for _, step := range steps.Content {
		script := nodeValue(step, "run")
		if script == nil || !strings.Contains(script.Value, "go test") {
			continue
		}
		if cont := nodeValue(step, "continue-on-error"); cont != nil {
			t.Errorf("test step sets continue-on-error: %q, so a failing test would not fail the workflow", cont.Value)
		}
	}
}

// TestWorkflowDefaultPermissionsAreReadOnly keeps the token scope minimal: the
// test job only needs to read the repository it checks out.
func TestWorkflowDefaultPermissionsAreReadOnly(t *testing.T) {
	root := loadWorkflow(t)
	permissions := nodeValue(root, "permissions")
	if permissions == nil {
		t.Fatal("workflow declares no permissions, so the job inherits the repository default token scope")
	}
	if contents := nodeValue(permissions, "contents"); contents == nil || contents.Value != "read" {
		t.Errorf("permissions.contents = %v, want read", contents)
	}
}

// commitSHA matches a full 40-hex commit id.
var commitSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

// TestThirdPartyActionsArePinnedToCommitSHAs stops a moving tag from changing
// what runs in CI: every `uses:` must reference a full commit, with the version
// kept as a trailing comment for humans.
func TestThirdPartyActionsArePinnedToCommitSHAs(t *testing.T) {
	walk(loadWorkflow(t), func(node *yaml.Node) {
		if node.Kind != yaml.MappingNode {
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value != "uses" {
				continue
			}
			ref := node.Content[i+1]
			repository, sha, found := strings.Cut(ref.Value, "@")
			if !found || !commitSHA.MatchString(sha) {
				t.Errorf("action %q is not pinned to a full commit SHA", ref.Value)
				continue
			}
			if !strings.Contains(ref.LineComment, "# v") {
				t.Errorf("action %q has no trailing version comment, so its pin is unreadable: %q", ref.Value, repository)
			}
		}
	})
}

// TestSetupGoReadsTheModuleVersion stops the workflow's Go version from being a
// second source of truth beside go.mod.
func TestSetupGoReadsTheModuleVersion(t *testing.T) {
	jobs := jobNodes(t, loadWorkflow(t))
	steps := nodeValue(nodeValue(jobs, "test"), "steps")
	var setupGo *yaml.Node
	for _, step := range steps.Content {
		if ref := nodeValue(step, "uses"); ref != nil && strings.HasPrefix(ref.Value, "actions/setup-go@") {
			setupGo = step
		}
	}
	if setupGo == nil {
		t.Fatal("workflow does not use actions/setup-go")
	}
	with := nodeValue(setupGo, "with")
	if got := nodeValue(with, "go-version-file"); got == nil || got.Value != "go.mod" {
		t.Errorf("setup-go go-version-file = %v, want go.mod", got)
	}
}

// walk visits every node in the tree rooted at node.
func walk(node *yaml.Node, visit func(*yaml.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for _, child := range node.Content {
		walk(child, visit)
	}
}
