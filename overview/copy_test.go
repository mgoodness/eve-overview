package overview_test

import (
	"strings"
	"testing"

	"github.com/mgoodness/eve-overview/overview"
)

// copySource authors one Preset that copies a loaded library Preset and
// appends its own Group set. The library stand-in's "✥ Target: Frig/Dest"
// carries one Group identifier and a non-empty State veto, so the copy's
// inheritance of both is observable.
func copySource() overview.Source {
	return overview.Source{
		Groups: []byte("groupSets:\n  extras:\n    - Sun\n"),
		Presets: []byte(`
presets:
  - name: "<color=0xFFFF6666>Exploring (+Hazards)</color>"
    copyOf: "✥ Target: Frig/Dest"
    groups:
      - extras
`),
		Tabs: []byte(`
tabs:
  - name: "<b> Exploring </b>"
    list: "Exploring (+Hazards)"
    bracket: "✥ Target: Frig/Dest"
`),
		Appearance: []byte(skeletonAppearanceYAML),
		Columns:    []byte(skeletonColumnsYAML),
	}
}

// TestGenerateCopiesLoadedPreset proves an authored Preset may copy a loaded
// library Preset by markup-free name: the copy inherits that Preset's Group
// whitelist and its two State lists, then appends the source's own Group sets
// on top. The loaded Groups pass through as raw identifiers — the library's
// data is never re-resolved.
func TestGenerateCopiesLoadedPreset(t *testing.T) {
	doc := decodeGeneric(t, generateWith(t, copySource(), overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: []byte(libraryPackYAML),
	}))

	entries := presetEntries(t, doc)
	const copied = "<color=0xFFFF6666>Exploring (+Hazards)</color>"
	body, ok := entries[copied]
	if !ok {
		t.Fatalf("emitted document has no copied Preset %q", copied)
	}
	pairs := body.([]any)

	// The base's group 25 comes first, then the source's own Set (Sun = 6).
	if got, want := groupsOf(t, body), []int{25, 6}; !equalInts(got, want) {
		t.Errorf("copied Preset groups = %v, want base groups plus the source's own %v", got, want)
	}
	if got, want := intList(t, lookup(t, pairs, "filteredStates")), []int{11}; !equalInts(got, want) {
		t.Errorf("copied Preset filteredStates = %v, want the base's %v", got, want)
	}
	if got := intList(t, lookup(t, pairs, "alwaysShownStates")); len(got) != 0 {
		t.Errorf("copied Preset alwaysShownStates = %v, want the base's empty list", got)
	}
}

// TestGenerateRejectsCopyOfMissingLibraryPreset proves a copyOf naming a
// Preset the library does not carry fails loudly and names what was missing,
// rather than emitting a copy with no base Groups.
func TestGenerateRejectsCopyOfMissingLibraryPreset(t *testing.T) {
	src := copySource()
	src.Presets = []byte(strings.Replace(string(src.Presets), "✥ Target: Frig/Dest", "✥ Target: Nonexistent", 1))

	_, err := overview.Generate(src, overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: []byte(libraryPackYAML),
	})
	if err == nil {
		t.Fatal("Generate succeeded with a copyOf naming a missing library Preset, want error")
	}
	if !strings.Contains(err.Error(), "Nonexistent") {
		t.Errorf("error %q does not name the missing library Preset", err)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
