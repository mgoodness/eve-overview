package overview_test

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mgoodness/eve-overview/overview"
)

// The walking-skeleton source: one Group set, one Preset carrying a full Group
// whitelist plus a State veto and a State override, and one Tab binding that
// Preset as both its list and its bracket preset.
const (
	skeletonGroupsYAML = `
groupSets:
  celestials:
    - Sun
    - Planet
    - Moon
    - 9 # raw Group identifier escape hatch; the SDE name is ambiguous
`
	skeletonPresetsYAML = `
presets:
  - name: "<color=0xFFFF6666>Engage</color>"
    groups:
      - celestials
    filteredStates: [11, 12, 14, 15, 16, 45, 49]
    alwaysShownStates: [13]
`
	skeletonTabsYAML = `
tabs:
  - name: "<b> Engage </b>"
    list: Engage
    bracket: Engage
`
	skeletonAppearanceYAML = `
flagOrder: [13, 44, 52, 11, 12, 14, 15, 16, 45, 9, 51, 68, 50, 53, 19, 18, 49, 10, 17, 48, 21, 20, 66]
flagStates: [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 44, 45, 48, 49, 50, 51, 52, 53, 68]
backgroundOrder: [13, 44, 52, 11, 12, 14, 15, 16, 45, 49, 19, 18, 9, 51, 68, 50, 53, 10, 20, 21, 17, 48, 66]
backgroundStates: [9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 44, 45, 48, 49, 50, 51, 52, 53, 68]
stateBlinks:
  flag_9: true
  flag_13: true
stateColorsNameList:
  background_9: orange
  flag_48: black
shipLabelOrder: [ship type, pilot name]
shipLabels:
  - key: ship type
    type: ship type
    pre: " "
    post: ""
    state: 1
    bold: true
    fontsize: 12
    color: [1.0, 0.7, 0.0]
  - key: pilot name
    type: pilot name
    pre: " ["
    post: "]"
    state: 1
`
	skeletonColumnsYAML = `
columnOrder: [ICON, TAG, DISTANCE, NAME, TYPE, ALLIANCE, CORPORATION, FACTION, MILITIA, SIZE, VELOCITY, RADIALVELOCITY, TRANSVERSALVELOCITY, ANGULARVELOCITY]
overviewColumns: [ALLIANCE, ANGULARVELOCITY, CORPORATION, DISTANCE, ICON, NAME, TRANSVERSALVELOCITY, TYPE, VELOCITY]
`
)

// skeletonSDEGroups is a minimal stand-in for the SDE Group matrix, holding
// only the Groups the walking-skeleton source names.
const skeletonSDEGroups = `{"_key":6,"name":{"en":"Sun"}}
{"_key":7,"name":{"en":"Planet"}}
{"_key":8,"name":{"en":"Moon"}}
`

func skeletonSource() overview.Source {
	return overview.Source{
		Groups:     []byte(skeletonGroupsYAML),
		Presets:    []byte(skeletonPresetsYAML),
		Tabs:       []byte(skeletonTabsYAML),
		Appearance: []byte(skeletonAppearanceYAML),
		Columns:    []byte(skeletonColumnsYAML),
	}
}

func generate(t *testing.T, src overview.Source) []byte {
	t.Helper()
	return generateWith(t, src, overview.Options{SDEGroups: []byte(skeletonSDEGroups)})
}

func generateWith(t *testing.T, src overview.Source, opts overview.Options) []byte {
	t.Helper()
	out, err := overview.Generate(src, opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}

// decodeGeneric decodes an emitted profile with the same generic decoder used
// for the reference pack, so tests observe the document the way the client
// would rather than the generator's internal model.
func decodeGeneric(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode emitted profile: %v\n%s", err, data)
	}
	return doc
}

// lookup finds a pair by key in a decoded ordered-pair collection.
func lookup(t *testing.T, pairs []any, key string) any {
	t.Helper()
	for _, raw := range pairs {
		pair, ok := raw.([]any)
		if !ok || len(pair) != 2 {
			t.Fatalf("expected ordered pair, got %#v", raw)
		}
		if pair[0] == key {
			return pair[1]
		}
	}
	t.Fatalf("key %q not found in %#v", key, pairs)
	return nil
}

// pairKeys returns the attribute names of a decoded ordered-pair collection,
// in order.
func pairKeys(t *testing.T, pairs []any) []string {
	t.Helper()
	keys := make([]string, 0, len(pairs))
	for _, raw := range pairs {
		pair, ok := raw.([]any)
		if !ok || len(pair) != 2 {
			t.Fatalf("expected an ordered pair, got %#v", raw)
		}
		key, ok := pair[0].(string)
		if !ok {
			t.Fatalf("pair key is %#v, want a string", pair[0])
		}
		keys = append(keys, key)
	}
	return keys
}

func TestGenerateEmitsImportableProfile(t *testing.T) {
	out := generate(t, skeletonSource())
	doc := decodeGeneric(t, out)

	wantKeys := []string{
		"backgroundOrder", "backgroundStates", "columnOrder", "flagOrder",
		"flagStates", "overviewColumns", "presets", "shipLabelOrder",
		"shipLabels", "stateBlinks", "stateColorsNameList", "tabSetup",
		"userSettings",
	}
	if len(doc) != len(wantKeys) {
		t.Fatalf("emitted %d top-level keys, want %d: %v", len(doc), len(wantKeys), doc)
	}
	for _, key := range wantKeys {
		if _, ok := doc[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	presets, ok := doc["presets"].([]any)
	if !ok || len(presets) != 1 {
		t.Fatalf("presets = %#v, want one preset", doc["presets"])
	}
	preset, ok := presets[0].([]any)
	if !ok || len(preset) != 2 {
		t.Fatalf("preset = %#v, want [name, pairs]", presets[0])
	}
	if got, want := preset[0], "<color=0xFFFF6666>Engage</color>"; got != want {
		t.Errorf("preset name = %q, want %q", got, want)
	}
	groups := lookup(t, preset[1].([]any), "groups").([]any)
	wantGroups := []any{6, 7, 8, 9}
	if len(groups) != len(wantGroups) {
		t.Fatalf("groups = %#v, want %#v", groups, wantGroups)
	}
	for i := range wantGroups {
		if groups[i] != wantGroups[i] {
			t.Errorf("groups[%d] = %v, want %v", i, groups[i], wantGroups[i])
		}
	}
	// The client always writes a preset's attributes in this order.
	if got := pairKeys(t, preset[1].([]any)); !equalStrings(got, []string{"alwaysShownStates", "filteredStates", "groups"}) {
		t.Errorf("preset attribute order = %v, want [alwaysShownStates filteredStates groups]", got)
	}

	tabs, ok := doc["tabSetup"].([]any)
	if !ok || len(tabs) != 1 {
		t.Fatalf("tabSetup = %#v, want one tab", doc["tabSetup"])
	}
	tab, ok := tabs[0].([]any)
	if !ok || len(tab) != 2 {
		t.Fatalf("tab = %#v, want [index, pairs]", tabs[0])
	}
	if tab[0] != 0 {
		t.Errorf("tab index = %v, want 0", tab[0])
	}
	if got := lookup(t, tab[1].([]any), "overview"); got != "<color=0xFFFF6666>Engage</color>" {
		t.Errorf("tab overview = %v", got)
	}
	if got := lookup(t, tab[1].([]any), "bracket"); got != "Engage" {
		t.Errorf("tab bracket = %v, want markup-free preset name", got)
	}
	if got := lookup(t, tab[1].([]any), "name"); got != "<b> Engage </b>" {
		t.Errorf("tab name = %v", got)
	}
	// A tab binds its list and bracket presets independently, so both slots
	// are always present and the client writes them in this order.
	if got := pairKeys(t, tab[1].([]any)); !equalStrings(got, []string{"bracket", "color", "name", "overview"}) {
		t.Errorf("tab attribute order = %v, want [bracket color name overview]", got)
	}

	columns, ok := doc["overviewColumns"].([]any)
	if !ok || len(columns) == 0 {
		t.Fatalf("overviewColumns = %#v", doc["overviewColumns"])
	}
	if columns[0] != "ALLIANCE" {
		t.Errorf("overviewColumns[0] = %v, want ALLIANCE", columns[0])
	}
	if userSettings, ok := doc["userSettings"].([]any); !ok || len(userSettings) != 0 {
		t.Errorf("userSettings = %#v, want []", doc["userSettings"])
	}
}

func intList(t *testing.T, value any) []int {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected a list, got %#v", value)
	}
	out := make([]int, 0, len(items))
	for _, item := range items {
		n, ok := item.(int)
		if !ok {
			t.Fatalf("expected an integer, got %#v", item)
		}
		out = append(out, n)
	}
	return out
}

func contains(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestEmittedVetoHidesAndOverrideRescues reads the two State lists back out of
// the emitted document — not the internal model — and applies the client's
// documented rule to them: a carried State in filteredStates vetoes the row,
// and a carried State in alwaysShownStates overrides that veto. An entity
// carrying only the friendly State 11 is hidden; the same entity also
// carrying the war-target State 13 is rescued.
func TestEmittedVetoHidesAndOverrideRescues(t *testing.T) {
	doc := decodeGeneric(t, generate(t, skeletonSource()))
	preset := doc["presets"].([]any)[0].([]any)
	pairs := preset[1].([]any)
	filtered := intList(t, lookup(t, pairs, "filteredStates"))
	always := intList(t, lookup(t, pairs, "alwaysShownStates"))

	if len(filtered) == 0 || len(always) == 0 {
		t.Fatalf("emitted preset must carry a veto and an override: filtered=%v always=%v", filtered, always)
	}

	visible := func(states ...int) bool {
		vetoed := false
		for _, state := range states {
			if contains(filtered, state) {
				vetoed = true
				break
			}
		}
		if !vetoed {
			return true
		}
		for _, state := range states {
			if contains(always, state) {
				return true
			}
		}
		return false
	}

	if visible(11) {
		t.Errorf("friendly State 11 is in the emitted veto %v, so a matching entity must be hidden", filtered)
	}
	if !visible(11, 13) {
		t.Errorf("war-target State 13 is in the emitted override %v, so it must rescue an entity the veto hides", always)
	}
}

func TestGenerateRejectsUnresolvableGroup(t *testing.T) {
	src := skeletonSource()
	src.Groups = []byte(strings.Replace(skeletonGroupsYAML, "- Sun", "- Dyson Sphere", 1))
	_, err := overview.Generate(src, overview.Options{SDEGroups: []byte(skeletonSDEGroups)})
	if err == nil {
		t.Fatal("Generate succeeded with an unresolvable Group, want error")
	}
	if !strings.Contains(err.Error(), "Dyson Sphere") {
		t.Errorf("error %q does not name the offending Group", err)
	}
}
