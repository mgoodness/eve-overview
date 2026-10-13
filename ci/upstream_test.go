// This file guards the upstream-drift workflow: it must run on a schedule,
// check the pinned reference data without being able to change it, and raise an
// issue only when that check reports drift. The workflow is a declarative
// artefact, so the test parses it.
package ci

import (
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var upstreamWorkflowPath = filepath.Join("..", ".github", "workflows", "upstream.yml")

func TestUpstreamWorkflowRunsOnASchedule(t *testing.T) {
	root := loadWorkflowAt(t, upstreamWorkflowPath)
	on := nodeValue(root, "on")
	if on == nil {
		t.Fatal(`upstream workflow has no "on" trigger block`)
	}
	schedule := nodeValue(on, "schedule")
	if schedule == nil || schedule.Kind != yaml.SequenceNode || len(schedule.Content) == 0 {
		t.Fatal("upstream workflow has no schedule, so drift is only noticed if someone remembers")
	}
	if nodeValue(on, "workflow_dispatch") == nil {
		t.Errorf("upstream workflow cannot be run on demand")
	}
}

// TestUpstreamWorkflowChecksAndRaisesAnIssue pins the two halves: the check
// runs the read-only `upstream` command, and the issue step fires only when
// that command reported drift.
func TestUpstreamWorkflowChecksAndRaisesAnIssue(t *testing.T) {
	jobs := jobNodes(t, loadWorkflowAt(t, upstreamWorkflowPath))
	job := nodeValue(jobs, "check-upstream")
	if job == nil {
		t.Fatal(`upstream workflow has no job with the id "check-upstream"`)
	}
	commands := runCommands(t, job)

	var checked, raised bool
	for _, command := range commands {
		if strings.Contains(command, "eve-overview upstream") {
			checked = true
		}
		if strings.Contains(command, "gh issue") && (strings.Contains(command, "create") || strings.Contains(command, "comment")) {
			raised = true
		}
	}
	if !checked {
		t.Errorf("no step runs the upstream check; commands: %q", commands)
	}
	if !raised {
		t.Errorf("no step opens or updates a drift issue; commands: %q", commands)
	}

	// The issue step must be gated on the check's outcome, so a green check
	// never spams the tracker.
	steps := nodeValue(job, "steps")
	for _, step := range steps.Content {
		script := nodeValue(step, "run")
		if script == nil || !strings.Contains(script.Value, "gh issue") {
			continue
		}
		condition := nodeValue(step, "if")
		if condition == nil || !strings.Contains(condition.Value, "steps.check.outcome") {
			t.Errorf("issue step condition = %v, want a gate on steps.check.outcome", condition)
		}
	}
}

// TestUpstreamWorkflowCannotWriteContents keeps the drift check read-only on
// the repository itself: it may open issues, never push.
func TestUpstreamWorkflowCannotWriteContents(t *testing.T) {
	permissions := nodeValue(loadWorkflowAt(t, upstreamWorkflowPath), "permissions")
	if permissions == nil {
		t.Fatal("upstream workflow declares no permissions")
	}
	if contents := nodeValue(permissions, "contents"); contents == nil || contents.Value != "read" {
		t.Errorf("permissions.contents = %v, want read", contents)
	}
	if issues := nodeValue(permissions, "issues"); issues == nil || issues.Value != "write" {
		t.Errorf("permissions.issues = %v, want write so the drift issue can be filed", issues)
	}
}

func TestUpstreamWorkflowActionsArePinnedToCommitSHAs(t *testing.T) {
	walk(loadWorkflowAt(t, upstreamWorkflowPath), func(node *yaml.Node) {
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
