// This file guards the Dependabot contract: a config that keeps dependencies
// moving, and an auto-merge workflow that is safe to run on
// pull_request_target and cannot merge a major bump unattended. Both are
// declarative artefacts, so the tests observe them the way GitHub does — by
// parsing them.
package ci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var (
	dependabotConfigPath   = filepath.Join("..", ".github", "dependabot.yml")
	dependabotWorkflowPath = filepath.Join("..", ".github", "workflows", "dependabot-auto-merge.yml")
)

// TestDependabotConfigCoversGoModulesAndActions pins the ecosystems and the
// cooldown: weekly, and no update proposed until a release has been out a
// week (zizmor's dependabot-cooldown audit wants >= 7 days).
func TestDependabotConfigCoversGoModulesAndActions(t *testing.T) {
	var config struct {
		Version int `yaml:"version"`
		Updates []struct {
			PackageEcosystem string `yaml:"package-ecosystem"`
			Schedule         struct {
				Interval string `yaml:"interval"`
			} `yaml:"schedule"`
			Cooldown struct {
				DefaultDays int `yaml:"default-days"`
			} `yaml:"cooldown"`
		} `yaml:"updates"`
	}
	body, err := os.ReadFile(dependabotConfigPath)
	if err != nil {
		t.Fatalf("read %s: %v", dependabotConfigPath, err)
	}
	if err := yaml.Unmarshal(body, &config); err != nil {
		t.Fatalf("parse %s: %v", dependabotConfigPath, err)
	}
	if config.Version != 2 {
		t.Errorf("%s: version = %d, want 2", dependabotConfigPath, config.Version)
	}

	seen := map[string]bool{}
	for _, update := range config.Updates {
		seen[update.PackageEcosystem] = true
		if update.Schedule.Interval != "weekly" {
			t.Errorf("%s: %s schedule = %q, want weekly", dependabotConfigPath, update.PackageEcosystem, update.Schedule.Interval)
		}
		if update.Cooldown.DefaultDays < 7 {
			t.Errorf("%s: %s cooldown.default-days = %d, want >= 7", dependabotConfigPath, update.PackageEcosystem, update.Cooldown.DefaultDays)
		}
	}
	for _, ecosystem := range []string{"gomod", "github-actions"} {
		if !seen[ecosystem] {
			t.Errorf("%s does not configure the %s ecosystem", dependabotConfigPath, ecosystem)
		}
	}
}

// TestDependabotAutoMergeDoesNotCheckOutPRCode is the safety half of
// pull_request_target: the trigger is only safe because the workflow never
// runs the PR's code.
func TestDependabotAutoMergeDoesNotCheckOutPRCode(t *testing.T) {
	root := loadWorkflowAt(t, dependabotWorkflowPath)
	on := nodeValue(root, "on")
	if on == nil {
		t.Fatal(`dependabot auto-merge workflow has no "on" trigger block`)
	}
	// `on: pull_request_target` parses as a scalar; a mapping is equally valid.
	switch on.Kind {
	case yaml.ScalarNode:
		if on.Value != "pull_request_target" {
			t.Errorf("dependabot auto-merge trigger = %q, want pull_request_target", on.Value)
		}
	default:
		if nodeValue(on, "pull_request_target") == nil {
			t.Errorf("dependabot auto-merge workflow does not run on pull_request_target")
		}
	}

	walk(root, func(node *yaml.Node) {
		if node.Kind != yaml.MappingNode {
			return
		}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == "uses" && strings.HasPrefix(node.Content[i+1].Value, "actions/checkout@") {
				t.Errorf("dependabot auto-merge workflow checks out code under pull_request_target")
			}
		}
	})
}

// TestDependabotAutoMergeIsScopedToDependabotMinorPatch is the other half:
// only Dependabot PRs, only minor and patch bumps, and the PR URL reaches the
// shell through env: rather than interpolation.
func TestDependabotAutoMergeIsScopedToDependabotMinorPatch(t *testing.T) {
	jobs := jobNodes(t, loadWorkflowAt(t, dependabotWorkflowPath))
	job := nodeValue(jobs, "auto-merge")
	if job == nil {
		t.Fatal(`dependabot auto-merge workflow has no job with the id "auto-merge"`)
	}
	condition := nodeValue(job, "if")
	if condition == nil || !strings.Contains(condition.Value, "github.event.pull_request.user.login") {
		t.Errorf("auto-merge job condition = %v, want a check on github.event.pull_request.user.login", condition)
	}
	if condition == nil || !strings.Contains(condition.Value, "dependabot[bot]") {
		t.Errorf("auto-merge job condition = %v, want it scoped to dependabot[bot]", condition)
	}

	steps := nodeValue(job, "steps")
	var mergeStep *yaml.Node
	for _, step := range steps.Content {
		if ref := nodeValue(step, "uses"); ref != nil && strings.HasPrefix(ref.Value, "dependabot/fetch-metadata@") {
			continue
		}
		if stepIf := nodeValue(step, "if"); stepIf != nil {
			mergeStep = step
		}
	}
	if mergeStep == nil {
		t.Fatal("auto-merge workflow has no gated merge step")
	}
	mergeIf := nodeValue(mergeStep, "if").Value
	for _, want := range []string{"steps.metadata.outputs.update-type", "version-update:semver-minor", "version-update:semver-patch"} {
		if !strings.Contains(mergeIf, want) {
			t.Errorf("merge step condition %q does not mention %q", mergeIf, want)
		}
	}
	if strings.Contains(mergeIf, "semver-major") {
		t.Errorf("merge step condition %q auto-merges major updates, which need a human", mergeIf)
	}

	env := nodeValue(mergeStep, "env")
	if nodeValue(env, "PR_URL") == nil {
		t.Errorf("merge step does not pass the PR URL through env:")
	}
	if got := nodeValue(mergeStep, "run"); got == nil || strings.Contains(got.Value, "${{") {
		t.Errorf("merge step run script %v interpolates a PR-derived value; pass it through env: instead", got)
	}
}

// TestDependabotAutoMergeActionsArePinnedToCommitSHAs mirrors the other pin
// guards: a moving tag must not change what runs on pull_request_target.
func TestDependabotAutoMergeActionsArePinnedToCommitSHAs(t *testing.T) {
	walk(loadWorkflowAt(t, dependabotWorkflowPath), func(node *yaml.Node) {
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
