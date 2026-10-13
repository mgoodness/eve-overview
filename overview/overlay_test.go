package overview_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mgoodness/eve-overview/overview"
)

// libraryPackYAML is a small stand-in for the Z-S reference pack: two Presets
// in the client's ordered-pair shape. The first carries a repeated Group
// identifier and two identifiers no SDE knows, which must ride through
// untouched — the real Collidables Preset has thirty duplicates, and 1211 and
// 4748 do not resolve as Groups.
const libraryPackYAML = `
presets:
  - - "<color=0xFFFF6666>✜ = CovOps: Collidables</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - []
      - - groups
        - - 6
          - 6
          - 1211
          - 4748
  - - "<color=0xFFFFFF66>✥ Target: Frig/Dest</color>"
    - - - alwaysShownStates
        - []
      - - filteredStates
        - - 11
      - - groups
        - - 25
`

// presetEntries decodes a profile's Presets into a map keyed by the Preset's
// exact (markup-bearing) name, so two documents can be compared name by name.
func presetEntries(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	presets, ok := doc["presets"].([]any)
	if !ok {
		t.Fatalf("presets = %#v, want a sequence", doc["presets"])
	}
	entries := make(map[string]any, len(presets))
	for i, raw := range presets {
		entry, ok := raw.([]any)
		if !ok || len(entry) != 2 {
			t.Fatalf("presets[%d] = %#v, want [name, body]", i, raw)
		}
		name, ok := entry[0].(string)
		if !ok {
			t.Fatalf("presets[%d] name = %#v, want a string", i, entry[0])
		}
		entries[name] = entry[1]
	}
	return entries
}

// TestGenerateCarriesLoadedLibraryUntouched proves the overlay is additive:
// every Preset the reference pack carries appears in the emitted document with
// its body unchanged, and the authored Preset is added alongside it rather
// than replacing anything.
func TestGenerateCarriesLoadedLibraryUntouched(t *testing.T) {
	emitted := decodeGeneric(t, generateWith(t, skeletonSource(), overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: []byte(libraryPackYAML),
	}))
	library := decodeGeneric(t, []byte(libraryPackYAML))

	got := presetEntries(t, emitted)
	want := presetEntries(t, library)

	for name, body := range want {
		emittedBody, ok := got[name]
		if !ok {
			t.Errorf("loaded preset %q is missing from the emitted document", name)
			continue
		}
		if !reflect.DeepEqual(emittedBody, body) {
			t.Errorf("loaded preset %q was modified:\n got %#v\nwant %#v", name, emittedBody, body)
		}
	}

	authored := "<color=0xFFFF6666>Engage</color>"
	if _, ok := got[authored]; !ok {
		t.Errorf("authored preset %q is missing from the emitted document", authored)
	}
	if len(got) != len(want)+1 {
		t.Errorf("emitted %d presets, want %d loaded plus one authored", len(got), len(want))
	}
}

// TestGenerateResolvesTabAgainstLoadedLibrary proves a Tab may bind an
// untouched Z-S Preset by markup-free name, and that the emitted reference
// carries the library's exact markup-bearing name.
func TestGenerateResolvesTabAgainstLoadedLibrary(t *testing.T) {
	src := skeletonSource()
	src.Tabs = []byte(`
tabs:
  - name: "<b> Targets </b>"
    list: "✥ Target: Frig/Dest"
    bracket: Engage
`)
	doc := decodeGeneric(t, generateWith(t, src, overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: []byte(libraryPackYAML),
	}))

	tabs := doc["tabSetup"].([]any)
	tab := tabs[0].([]any)
	overviewName := lookup(t, tab[1].([]any), "overview")
	if want := "<color=0xFFFFFF66>✥ Target: Frig/Dest</color>"; overviewName != want {
		t.Errorf("tab overview = %q, want the loaded library's exact name %q", overviewName, want)
	}
	if got := lookup(t, tab[1].([]any), "bracket"); got != "Engage" {
		t.Errorf("tab bracket = %q, want the authored preset's markup-free name %q", got, "Engage")
	}
}

// TestGenerateRejectsAuthoredPresetCollidingWithLoadedLibrary proves a
// collision between an authored Preset name and a loaded one — compared once
// colour markup is stripped — fails loudly rather than silently overriding the
// library entry.
func TestGenerateRejectsAuthoredPresetCollidingWithLoadedLibrary(t *testing.T) {
	pack := strings.Replace(libraryPackYAML,
		"<color=0xFFFFFF66>✥ Target: Frig/Dest</color>",
		"<color=0x00FF00FF>Engage</color>", 1)

	_, err := overview.Generate(skeletonSource(), overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: []byte(pack),
	})
	if err == nil {
		t.Fatal("Generate succeeded with an authored Preset name colliding with a loaded one, want error")
	}
	if !strings.Contains(err.Error(), "Engage") {
		t.Errorf("error %q does not name the colliding Preset", err)
	}
}
