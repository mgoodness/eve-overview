package overview_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/mgoodness/eve-overview/internal/reference"
	"github.com/mgoodness/eve-overview/overview"
)

// conformanceSkipped records that the dialect conformance test could not read
// the reference pack. TestMain turns a skipped conformance run red, so a
// full-skip run never passes silently.
var conformanceSkipped bool

func TestMain(m *testing.M) {
	code := m.Run()
	if code == 0 && conformanceSkipped && !testing.Short() {
		fmt.Fprintln(os.Stderr, "dialect conformance did not run and this was not a -short run; refusing to report success")
		code = 1
	}
	os.Exit(code)
}

// TestDialectConformance decodes the emitted profile and the reference pack
// with the same generic decoder and compares the shape of every shared
// top-level key. It also pins the things a shape comparison alone would miss:
// the ordered-pair collections, the integer State-order lists, the exact
// telemetry column tokens and the colour-markup form.
//
// The reference pack is a real Z-S export fetched at test time and pinned by
// checksum. Offline, the test skips loudly — and TestMain makes that a failing
// run, because a test that cannot check the dialect must not look like one
// that did.
func TestDialectConformance(t *testing.T) {
	pack, ok := referencePack(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference pack unavailable (offline?); skipping dialect conformance — run without -short to require it")
	}

	// The reference pack travels through the same Options struct as the SDE
	// matrix: Generate's inputs are all supplied bytes, never a global cache
	// path. Ticket #2's walking skeleton does not yet overlay the pack, but the
	// seam accepts it so the library ticket can consume it without moving.
	emitted := generateWith(t, skeletonSource(), overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: pack,
	})
	ours := decodeGeneric(t, emitted)
	theirs := decodeGeneric(t, pack)

	if got, want := topLevelKeys(t, emitted), topLevelKeys(t, pack); !equalStrings(got, want) {
		t.Errorf("top-level key order = %v, want %v", got, want)
	}

	for key, referenceValue := range theirs {
		ourValue, ok := ours[key]
		if !ok {
			t.Errorf("emitted document is missing top-level key %q", key)
			continue
		}
		if got, want := shapeOf(ourValue, 2), shapeOf(referenceValue, 2); got != want {
			t.Errorf("shape of %q = %s, want %s", key, got, want)
		}
	}

	assertIntegerList(t, ours, "flagOrder")
	assertIntegerList(t, ours, "backgroundOrder")
	assertIntegerList(t, ours, "flagStates")
	assertIntegerList(t, ours, "backgroundStates")

	if got, want := stringList(t, ours["overviewColumns"]), stringList(t, theirs["overviewColumns"]); !equalStrings(got, want) {
		t.Errorf("overviewColumns = %v, want the reference pack's %v", got, want)
	}
	if got, want := stringList(t, ours["columnOrder"]), stringList(t, theirs["columnOrder"]); !equalStrings(got, want) {
		t.Errorf("columnOrder = %v, want the reference pack's %v", got, want)
	}

	assertOrderedPairs(t, ours, "presets")
	assertOrderedPairs(t, ours, "tabSetup")
	assertAuthoredPresetNamesHaveMarkup(t, ours, theirs)
	assertMarkupTabNames(t, ours)
}

// referencePack reads the pinned reference pack through the cache layer. It
// returns ok=false when the pack is neither cached nor fetchable, which is the
// offline case the test skips on.
func referencePack(t *testing.T) ([]byte, bool) {
	t.Helper()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, false
	}
	pack, err := reference.New(filepath.Join(cacheDir, "eve-overview")).ReferencePack(context.Background())
	if err != nil {
		t.Logf("reference pack unavailable: %v", err)
		return nil, false
	}
	return pack, true
}

// shapeOf renders a value's structural signature to depth, erasing scalar
// values but keeping their kinds. Nulls are wildcards: an absent value and an
// explicit null are the same thing to the client.
func shapeOf(value any, depth int) string {
	if depth <= 0 {
		return "?"
	}
	switch v := value.(type) {
	case nil:
		return "?"
	case bool:
		return "bool"
	case int, int64, int32, uint64, uint32, float64, float32:
		return "number"
	case string:
		return "string"
	case []any:
		kinds := map[string]struct{}{}
		for _, item := range v {
			if kind := shapeOf(item, depth-1); kind != "?" {
				kinds[kind] = struct{}{}
			}
		}
		return "seq[" + strings.Join(sortedKeys(kinds), "|") + "]"
	case map[string]any:
		keys := sortedKeysOf(v)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, key+":"+shapeOf(v[key], depth-1))
		}
		return "map{" + strings.Join(parts, ",") + "}"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func sortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func equalStrings(a, b []string) bool {
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

func topLevelKeys(t *testing.T, data []byte) []string {
	t.Helper()
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatalf("decode YAML document: %v", err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) == 0 {
		t.Fatal("expected a YAML document")
	}
	mapping := document.Content[0]
	if mapping.Kind != yaml.MappingNode {
		t.Fatalf("top-level value is not a mapping (kind %d)", mapping.Kind)
	}
	keys := make([]string, 0, len(mapping.Content)/2)
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		keys = append(keys, mapping.Content[i].Value)
	}
	return keys
}

func assertIntegerList(t *testing.T, doc map[string]any, key string) {
	t.Helper()
	items, ok := doc[key].([]any)
	if !ok {
		t.Errorf("%s is not a sequence: %#v", key, doc[key])
		return
	}
	for i, item := range items {
		if _, ok := item.(int); !ok {
			t.Errorf("%s[%d] = %#v, want an integer State identifier", key, i, item)
		}
	}
}

func stringList(t *testing.T, value any) []string {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected a sequence, got %#v", value)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("expected a string, got %#v", item)
		}
		out = append(out, text)
	}
	return out
}

// assertOrderedPairs checks the collection is a sequence of [key, body] pairs,
// each body a sequence of [attribute, value] pairs — the Python ordered-map
// shape the client writes and reads.
func assertOrderedPairs(t *testing.T, doc map[string]any, key string) {
	t.Helper()
	collection, ok := doc[key].([]any)
	if !ok {
		t.Fatalf("%s is not a sequence: %#v", key, doc[key])
	}
	for i, raw := range collection {
		entry, ok := raw.([]any)
		if !ok || len(entry) != 2 {
			t.Fatalf("%s[%d] is %#v, want [key, body]", key, i, raw)
		}
		if key == "presets" {
			if _, ok := entry[0].(string); !ok {
				t.Errorf("%s[%d] name is %#v, want a string", key, i, entry[0])
			}
		}
		body, ok := entry[1].([]any)
		if !ok {
			t.Fatalf("%s[%d] body is %#v, want a sequence of pairs", key, i, entry[1])
		}
		for j, rawPair := range body {
			pair, ok := rawPair.([]any)
			if !ok || len(pair) != 2 {
				t.Fatalf("%s[%d] body[%d] is %#v, want [attribute, value]", key, i, j, rawPair)
			}
			if _, ok := pair[0].(string); !ok {
				t.Errorf("%s[%d] body[%d] attribute is %#v, want a string", key, i, j, pair[0])
			}
		}
	}
}

var colorMarkup = regexp.MustCompile(`^<color=0x[0-9A-Fa-f]{8}>.*</color>$`)

// eveMarkup matches the inline markup the client writes into names, the same
// classes the generator strips. The conformance test strips it independently,
// so the two sides of the comparison cannot share a bug.
var eveMarkup = regexp.MustCompile(`</?(?:color(?:=0x[0-9A-Fa-f]+)?|fontsize(?:=\d+)?)>|</?[biu]>`)

// withoutMarkup strips the client's inline markup, leaving the plain name a
// preset is referenced by.
func withoutMarkup(name string) string {
	return strings.TrimSpace(eveMarkup.ReplaceAllString(name, ""))
}

// presetNamesByMarkupFreeName indexes a decoded profile's Presets by the name
// a Tab or another Preset would refer to them by.
func presetNamesByMarkupFreeName(t *testing.T, doc map[string]any) map[string]string {
	t.Helper()
	presets, ok := doc["presets"].([]any)
	if !ok {
		t.Fatalf("presets = %#v, want a sequence", doc["presets"])
	}
	byName := make(map[string]string, len(presets))
	for i, raw := range presets {
		name, ok := raw.([]any)[0].(string)
		if !ok {
			t.Fatalf("presets[%d] name = %#v, want a string", i, raw.([]any)[0])
		}
		byName[withoutMarkup(name)] = name
	}
	return byName
}

// authoredPresetNames are the Presets our source authors rather than inherits.
// They are the names the library-match assertion sets aside.
var authoredPresetNames = map[string]bool{"Engage": true}

// TestEmittedPresetsMatchReferenceLibrary is ticket #4's central property: the
// emitted Preset set, minus our authored Presets, is exactly the reference
// pack's Preset set — matched once colour markup is ignored — and each loaded
// Preset rides through with its body unmodified. Both sides are decoded with
// the same generic decoder, so this compares the documents a client would read
// rather than the generator's model.
func TestEmittedPresetsMatchReferenceLibrary(t *testing.T) {
	pack, ok := referencePack(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference pack unavailable (offline?); skipping library overlay conformance — run without -short to require it")
	}

	emitted := generateWith(t, skeletonSource(), overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: pack,
	})
	ours := decodeGeneric(t, emitted)
	theirs := decodeGeneric(t, pack)

	ourNames := presetNamesByMarkupFreeName(t, ours)
	referenceNames := presetNamesByMarkupFreeName(t, theirs)

	for name := range authoredPresetNames {
		delete(ourNames, name)
	}

	if len(ourNames) != len(referenceNames) {
		t.Errorf("emitted %d inherited Presets, want the reference pack's %d", len(ourNames), len(referenceNames))
	}
	for name, referenceExact := range referenceNames {
		ourExact, carried := ourNames[name]
		if !carried {
			t.Errorf("reference Preset %q is missing from the emitted document", referenceExact)
			continue
		}
		if ourExact != referenceExact {
			t.Errorf("emitted %q, want the reference pack's exact name %q", ourExact, referenceExact)
		}
	}

	ourEntries := presetEntriesFrom(t, ours)
	referenceEntries := presetEntriesFrom(t, theirs)
	for name, referenceExact := range referenceNames {
		ourExact, carried := ourNames[name]
		if !carried {
			continue
		}
		if !reflect.DeepEqual(ourEntries[ourExact], referenceEntries[referenceExact]) {
			t.Errorf("loaded Preset %q was modified:\n got %#v\nwant %#v", referenceExact, ourEntries[ourExact], referenceEntries[referenceExact])
		}
	}
}

// TestReferenceLibraryEdgeCaseIdentifiersPassThrough proves the library's own
// data is never repaired: the collidables Preset keeps its repeated
// identifiers, and the identifiers no SDE maps (1211, 4748) ride through
// verbatim. Only an unresolvable Group name in our source is fatal.
func TestReferenceLibraryEdgeCaseIdentifiersPassThrough(t *testing.T) {
	pack, ok := referencePack(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference pack unavailable (offline?); skipping library edge-case conformance — run without -short to require it")
	}

	emitted := generateWith(t, skeletonSource(), overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: pack,
	})
	ours := decodeGeneric(t, emitted)
	theirs := decodeGeneric(t, pack)

	ourEntries := presetEntriesFrom(t, ours)
	referenceEntries := presetEntriesFrom(t, theirs)

	const (
		collidables = "<color=0xFFFF6666>✜ = CovOps: Collidables</color>"
		dscanAll    = "<color=0xFFFD96CF>※ --- D-Scan: All</color>"
	)

	// The library's repeated identifiers survive untouched. The collidables
	// Preset carries thirty of them across 675 entries; if it ever stops, this
	// tolerance is no longer being exercised.
	groups := groupsOf(t, ourEntries[collidables])
	seen := map[int]bool{}
	duplicates := 0
	for _, id := range groups {
		if seen[id] {
			duplicates++
		}
		seen[id] = true
	}
	if duplicates == 0 {
		t.Error("collidables Preset lost its repeated identifiers; the tolerance is untested")
	}
	if !reflect.DeepEqual(groups, groupsOf(t, referenceEntries[collidables])) {
		t.Error("collidables Preset's Group list was not carried through verbatim")
	}

	// Identifiers no SDE maps survive untouched too: 1211 and 4748 sit in this
	// Preset's whitelist and must not be dropped or resolved.
	orphans := groupsOf(t, ourEntries[dscanAll])
	for _, orphan := range []int{1211, 4748} {
		if !contains(orphans, orphan) {
			t.Errorf("D-Scan: All no longer carries unresolvable identifier %d", orphan)
		}
	}
	if !reflect.DeepEqual(orphans, groupsOf(t, referenceEntries[dscanAll])) {
		t.Error("D-Scan: All's Group list was not carried through verbatim")
	}
}

// TestTabBindsLoadedLibraryPreset proves a Tab can bind an untouched Z-S Preset
// from the real pack by markup-free name, and that the emitted reference is the
// library's exact markup-bearing name.
func TestTabBindsLoadedLibraryPreset(t *testing.T) {
	pack, ok := referencePack(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference pack unavailable (offline?); skipping loaded-library tab binding — run without -short to require it")
	}

	const loadedPreset = "<color=0xFFFFFFFF>✈ Travel (-Ship +Sta)</color>"
	names := presetNamesByMarkupFreeName(t, decodeGeneric(t, pack))
	if names["✈ Travel (-Ship +Sta)"] != loadedPreset {
		t.Fatalf("fixture assumption broken: the reference pack's name is %q", names["✈ Travel (-Ship +Sta)"])
	}

	src := skeletonSource()
	src.Tabs = []byte("tabs:\n  - name: \"<b> Travel </b>\"\n    list: \"✈ Travel (-Ship +Sta)\"\n    bracket: Engage\n")
	emitted := generateWith(t, src, overview.Options{
		SDEGroups:     []byte(skeletonSDEGroups),
		ReferencePack: pack,
	})
	doc := decodeGeneric(t, emitted)
	tabs := doc["tabSetup"].([]any)
	tab := tabs[0].([]any)
	if got := lookup(t, tab[1].([]any), "overview"); got != loadedPreset {
		t.Errorf("tab overview = %q, want the reference pack's exact name %q", got, loadedPreset)
	}
}

// presetEntriesFrom decodes a profile's Presets into a map keyed by exact
// name. The generator tests have their own copy; this one keeps the
// conformance assertions independent of that package's helpers.
func presetEntriesFrom(t *testing.T, doc map[string]any) map[string]any {
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

// groupsOf reads a preset body's Group identifier list.
func groupsOf(t *testing.T, body any) []int {
	t.Helper()
	pairs, ok := body.([]any)
	if !ok {
		t.Fatalf("preset body = %#v, want pairs", body)
	}
	groups, ok := lookup(t, pairs, "groups").([]any)
	if !ok {
		t.Fatalf("preset groups = %#v, want a sequence", lookup(t, pairs, "groups"))
	}
	ids := make([]int, 0, len(groups))
	for _, raw := range groups {
		id, ok := raw.(int)
		if !ok {
			t.Fatalf("Group identifier = %#v, want an integer", raw)
		}
		ids = append(ids, id)
	}
	return ids
}

// assertAuthoredPresetNamesHaveMarkup checks the Presets we author carry the
// client's colour markup in its exact 0xAARRGGBB form. The reference pack's
// own bracket Presets are legitimately plain text, so names the pack already
// carries are exempt — the library rides through exactly as it was shipped.
func assertAuthoredPresetNamesHaveMarkup(t *testing.T, ours, theirs map[string]any) {
	t.Helper()
	reference := map[string]bool{}
	for _, raw := range theirs["presets"].([]any) {
		name, _ := raw.([]any)[0].(string)
		reference[name] = true
	}

	presets := ours["presets"].([]any)
	if len(presets) == 0 {
		t.Fatal("emitted document has no presets")
	}
	authored := 0
	for i, raw := range presets {
		name, _ := raw.([]any)[0].(string)
		if reference[name] {
			continue
		}
		authored++
		if !colorMarkup.MatchString(name) {
			t.Errorf("authored presets[%d] name %q is not <color=0xAARRGGBB>…</color>", i, name)
		}
	}
	if authored == 0 {
		t.Error("emitted document carries no authored Preset; the overlay assertion would be vacuous")
	}
}

// assertMarkupTabNames checks a tab's list preset reference keeps its markup
// while its bracket reference does not, which is the client's convention.
func assertMarkupTabNames(t *testing.T, doc map[string]any) {
	t.Helper()
	tabs := doc["tabSetup"].([]any)
	if len(tabs) == 0 {
		t.Fatal("emitted document has no tabs")
	}
	for i, raw := range tabs {
		pairs := raw.([]any)[1].([]any)
		overviewName, _ := lookup(t, pairs, "overview").(string)
		if !colorMarkup.MatchString(overviewName) {
			t.Errorf("tabSetup[%d] overview %q is not <color=0xAARRGGBB>…</color>", i, overviewName)
		}
		bracketName, _ := lookup(t, pairs, "bracket").(string)
		if bracketName == "" || strings.Contains(bracketName, "<") {
			t.Errorf("tabSetup[%d] bracket %q must be a markup-free preset name", i, bracketName)
		}
	}
}
