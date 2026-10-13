package overview_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mgoodness/eve-overview/internal/reference"
	"github.com/mgoodness/eve-overview/overview"
)

// realReferenceData reads the pinned SDE Group matrix and Z-S reference pack
// through the same cache-and-checksum layer the CLI uses. It returns ok=false
// when either is unavailable — offline, say — which the roster tests skip on,
// loudly, while setting the conformance guard so a fully-skipped non-short run
// still fails. No name→identifier fixture is hand-written here: the matrix
// under test is the real one.
func realReferenceData(t *testing.T) (sde, pack []byte, ok bool) {
	t.Helper()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Logf("no user cache directory: %v", err)
		return nil, nil, false
	}
	fetcher := reference.New(filepath.Join(cacheDir, "eve-overview"))
	sde, err = fetcher.SDEGroups(context.Background())
	if err != nil {
		t.Logf("SDE Group matrix unavailable: %v", err)
		return nil, nil, false
	}
	pack, err = fetcher.ReferencePack(context.Background())
	if err != nil {
		t.Logf("reference pack unavailable: %v", err)
		return nil, nil, false
	}
	return sde, pack, true
}

// realSource reads the checked-in source-of-truth tree. Unlike the reference
// data, an unreadable source file is a repository fault, not an offline skip.
func realSource(t *testing.T) overview.Source {
	t.Helper()
	read := func(name string) []byte {
		body, err := os.ReadFile(filepath.Join("..", "source", name))
		if err != nil {
			t.Fatalf("read source/%s: %v", name, err)
		}
		return body
	}
	return overview.Source{
		Groups:     read("groups.yaml"),
		Presets:    read("presets.yaml"),
		Tabs:       read("tabs.yaml"),
		Appearance: read("appearance.yaml"),
		Columns:    read("columns.yaml"),
	}
}

// expectedRoster is the ten-Tab roster in slot order: the Tab's markup label,
// its markup-free list Preset, and its markup-free Bracket Preset. The hunter's
// pressure set leads; the podsaver Tab is the only one on the warpout bracket.
var expectedRoster = []struct {
	label   string
	list    string
	bracket string
}{
	{"<b> Travel </b>", "✈ Travel (-Ship +Sta)", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> Engage </b>", "Engage", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> Exit! </b>", "➲ Extra: Podsaver (!)", "⌘ Brackets: Warpout (!)"},
	{"<b> Holes </b>", "Holes", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> Fleet </b>", "✪ = Friendly: Fleet", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> PvE </b>", "✜ PvX: Mission Filter (+LCO +LCS)", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> PI </b>", "✜ PvX: PI", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> Exploring </b>", "✜ PvX: Exploring (+Hazards)", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> Loot </b>", "✜ PvX: Salvage/Loot", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
	{"<b> Occupancy </b>", "Occupancy", "⌘ Brackets: Combat/Targets/Friendly (+Dro)"},
}

// TestProfileCarriesTheTenTabRoster asserts the whole roster through the
// emitted document: ten slots in order, each binding a list and a bracket
// Preset that both resolve in the emitted library, the podsaver Tab alone on
// the warpout bracket, and the collidables Preset present but bound to nothing.
func TestProfileCarriesTheTenTabRoster(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping roster conformance — run without -short to require it")
	}

	emitted := generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack})
	doc := decodeGeneric(t, emitted)
	names := presetNamesByMarkupFreeName(t, doc)

	if _, present := names["✜ = CovOps: Collidables"]; !present {
		t.Fatal("the collidables Preset is missing from the emitted library")
	}

	tabs, ok := doc["tabSetup"].([]any)
	if !ok {
		t.Fatalf("tabSetup = %#v, want a sequence", doc["tabSetup"])
	}
	if len(tabs) != len(expectedRoster) {
		t.Fatalf("emitted %d tabs, want the %d-Tab roster", len(tabs), len(expectedRoster))
	}

	for i, want := range expectedRoster {
		tab, ok := tabs[i].([]any)
		if !ok || len(tab) != 2 {
			t.Fatalf("tabSetup[%d] = %#v, want [index, pairs]", i, tabs[i])
		}
		if tab[0] != i {
			t.Errorf("tabSetup[%d] index = %v, want %d", i, tab[0], i)
		}
		pairs := tab[1].([]any)

		if got := lookup(t, pairs, "name"); got != want.label {
			t.Errorf("tabSetup[%d] name = %v, want %q", i, got, want.label)
		}

		overviewName, _ := lookup(t, pairs, "overview").(string)
		if got := withoutMarkup(overviewName); got != want.list {
			t.Errorf("tabSetup[%d] list preset = %q, want %q", i, got, want.list)
		}
		if _, resolves := names[withoutMarkup(overviewName)]; !resolves {
			t.Errorf("tabSetup[%d] binds list preset %q, which is not in the emitted library", i, overviewName)
		}
		if withoutMarkup(overviewName) == "✜ = CovOps: Collidables" {
			t.Errorf("tabSetup[%d] binds the collidables Preset, which must stay unbound", i)
		}

		bracket, _ := lookup(t, pairs, "bracket").(string)
		if got := withoutMarkup(bracket); got != want.bracket {
			t.Errorf("tabSetup[%d] bracket preset = %q, want %q", i, got, want.bracket)
		}
		if _, resolves := names[withoutMarkup(bracket)]; !resolves {
			t.Errorf("tabSetup[%d] binds bracket preset %q, which is not in the emitted library", i, bracket)
		}
	}
}

// The 55 player-ship Groups, in the order the source lists them. Reading them
// from the decoded document and comparing to these literal identifiers keeps
// the expectation independent of the resolution path: a name that resolved to
// the wrong Group, or a Group that went missing, changes the list.
var engageShipGroups = []int{
	25, 26, 27, 28, 29, 30, 31, 90, 237, 324,
	358, 380, 381, 419, 420, 463, 485, 513, 540, 541,
	543, 547, 659, 830, 831, 832, 833, 834, 863, 864,
	883, 893, 894, 898, 900, 902, 906, 941, 963, 1022,
	1201, 1202, 1283, 1305, 1527, 1534, 1538, 1548, 1704, 1972,
	2001, 4594, 4902, 5087, 5120,
}

// TestEngagePresetCarriesShipsCelestialsAndHazards asserts the mixed working
// Preset's whitelist is exactly the 55 player-ship Groups, the six warp-to
// celestials (Sun, Planet, Moon, the raw-identifier Asteroid Belt 9, Stargate,
// Wormhole) and the six decloak hazards, with every Friendly State vetoed and
// the war target rescued.
func TestEngagePresetCarriesShipsCelestialsAndHazards(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping Engage conformance — run without -short to require it")
	}

	doc := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))
	body, present := presetEntriesFrom(t, doc)["<color=0xFFFF6666>Engage</color>"]
	if !present {
		t.Fatal("the emitted library does not carry the authored Engage Preset")
	}

	want := append([]int{}, engageShipGroups...)
	want = append(want,
		6, 7, 8, 9, 10, 988, // Sun, Planet, Moon, Asteroid Belt (raw 9), Stargate, Wormhole
		226, 319, 365, 411, // Large Collidable Object/Structure, Control Tower, Force Field
		361, 1246, // Mobile Warp Disruptor, Mobile Depot
	)
	if got := groupsOf(t, body); !equalInts(got, want) {
		t.Errorf("Engage groups = %v, want %v", got, want)
	}

	pairs := body.([]any)
	if got, want := intList(t, lookup(t, pairs, "filteredStates")), []int{11, 12, 14, 15, 16, 45, 49}; !equalInts(got, want) {
		t.Errorf("Engage filteredStates = %v, want the Friendly States %v", got, want)
	}
	if got, want := intList(t, lookup(t, pairs, "alwaysShownStates")), []int{13}; !equalInts(got, want) {
		t.Errorf("Engage alwaysShownStates = %v, want the war target %v", got, want)
	}
}

// TestExploringPresetCopyAddsHazardsAndEscapeCelestials asserts the authored
// exploration Preset is the loaded Z-S Preset's whitelist plus the four large
// collidables and the escape celestials, with the loaded Preset's State veto
// left intact.
func TestExploringPresetCopyAddsHazardsAndEscapeCelestials(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping exploration-copy conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))
	referencePackDoc := decodeGeneric(t, pack)

	base, present := presetEntriesFrom(t, referencePackDoc)["<color=0xFFFF6666>✜ PvX: Exploring</color>"]
	if !present {
		t.Fatal("the reference pack does not carry ✜ PvX: Exploring; fixture assumption broken")
	}
	copied, present := presetEntriesFrom(t, emitted)["<color=0xFFFF6666>✜ PvX: Exploring (+Hazards)</color>"]
	if !present {
		t.Fatal("the emitted library does not carry the authored ✜ PvX: Exploring (+Hazards) Preset")
	}

	baseGroups := groupsOf(t, base)
	copiedGroups := groupsOf(t, copied)
	if len(copiedGroups) <= len(baseGroups) {
		t.Fatalf("exploration copy carries %d groups, want strictly more than the base's %d", len(copiedGroups), len(baseGroups))
	}
	have := map[int]bool{}
	for _, id := range copiedGroups {
		have[id] = true
	}
	for _, id := range baseGroups {
		if !have[id] {
			t.Fatalf("exploration copy dropped base Group %d; it must be a strict superset", id)
		}
	}
	for _, id := range []int{7, 8, 9, 226, 319, 365, 411} {
		if !have[id] {
			t.Errorf("exploration copy is missing Group %d (Planet/Moon/Asteroid Belt and the four large collidables)", id)
		}
	}

	if !equalInts(intList(t, lookup(t, copied.([]any), "filteredStates")), intList(t, lookup(t, base.([]any), "filteredStates"))) {
		t.Errorf("exploration copy filteredStates = %v, want the loaded base's %v",
			lookup(t, copied.([]any), "filteredStates"), lookup(t, base.([]any), "filteredStates"))
	}
}

// TestProfileHasAHoleAndAnOccupancyTab checks the two hunter Tabs the roster
// promises by their Preset contents: a wormhole-only Tab and an occupancy Tab
// of moons and warp disruption fields. The assertions read the emitted Presets
// those Tabs bind, so the Tab-to-Preset binding is exercised end to end.
func TestProfileHasAHoleAndAnOccupancyTab(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping hole/occupancy conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))
	entries := presetEntriesFrom(t, emitted)

	holes, present := entries["<color=0xFFFF6666>Holes</color>"]
	if !present {
		t.Fatal("the emitted library does not carry the authored Holes Preset")
	}
	if got, want := groupsOf(t, holes), []int{988}; !equalInts(got, want) {
		t.Errorf("Holes groups = %v, want the Wormhole Group %v", got, want)
	}
	if got := intList(t, lookup(t, holes.([]any), "filteredStates")); len(got) != 0 {
		t.Errorf("Holes filteredStates = %v, want empty: the scanner ignores States", got)
	}

	occupancy, present := entries["<color=0xFFFF6666>Occupancy</color>"]
	if !present {
		t.Fatal("the emitted library does not carry the authored Occupancy Preset")
	}
	if got, want := groupsOf(t, occupancy), []int{8, 361}; !equalInts(got, want) {
		t.Errorf("Occupancy groups = %v, want Moon and Mobile Warp Disruptor %v", got, want)
	}

	// The Tabs themselves bind those Presets. Find the overview references by
	// name so a roster edit that unbinds them fails here.
	bound := map[string]bool{}
	for _, raw := range emitted["tabSetup"].([]any) {
		pairs := raw.([]any)[1].([]any)
		name, _ := lookup(t, pairs, "overview").(string)
		bound[withoutMarkup(name)] = true
	}
	for _, want := range []string{"Holes", "Occupancy"} {
		if !bound[want] {
			t.Errorf("no Tab binds the %q Preset", want)
		}
	}
}
