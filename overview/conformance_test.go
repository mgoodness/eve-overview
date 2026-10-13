package overview_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	assertMarkupPresetNames(t, ours)
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

// assertMarkupPresetNames checks every authored Preset name carries the
// client's colour markup in its exact 0xAARRGGBB form.
func assertMarkupPresetNames(t *testing.T, doc map[string]any) {
	t.Helper()
	presets := doc["presets"].([]any)
	if len(presets) == 0 {
		t.Fatal("emitted document has no presets")
	}
	for i, raw := range presets {
		name, _ := raw.([]any)[0].(string)
		if !colorMarkup.MatchString(name) {
			t.Errorf("presets[%d] name %q is not <color=0xAARRGGBB>…</color>", i, name)
		}
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
