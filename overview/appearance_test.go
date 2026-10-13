package overview_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mgoodness/eve-overview/overview"
)

// informationFreeStates never draw a Background: no standing, neutral
// standing, the reserved identifier and agent-interactable. Every stranger in
// J-space carries one of them, so a Background for them is decoration rather
// than signal.
var informationFreeStates = []int{17, 20, 21, 48}

// withoutStates returns values with every occurrence of the removed States
// dropped, preserving the order of what is left.
func withoutStates(values []int, removed ...int) []int {
	drop := make(map[int]bool, len(removed))
	for _, state := range removed {
		drop[state] = true
	}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if !drop[value] {
			out = append(out, value)
		}
	}
	return out
}

// TestBackgroundsAreRestrictedToRelationshipAndThreatStates asserts the
// emitted Background whitelist and priority order are the reference pack's
// with the four information-free States removed: relationship and threat
// States keep exactly their pack treatment, and no stranger glows.
func TestBackgroundsAreRestrictedToRelationshipAndThreatStates(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping appearance conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))
	reference := decodeGeneric(t, pack)

	wantStates := withoutStates(intList(t, reference["backgroundStates"]), informationFreeStates...)
	wantOrder := withoutStates(intList(t, reference["backgroundOrder"]), informationFreeStates...)

	if got := intList(t, emitted["backgroundStates"]); !equalInts(got, wantStates) {
		t.Errorf("backgroundStates = %v, want the reference pack's without %v = %v", got, informationFreeStates, wantStates)
	}
	if got := intList(t, emitted["backgroundOrder"]); !equalInts(got, wantOrder) {
		t.Errorf("backgroundOrder = %v, want the reference pack's without %v = %v", got, informationFreeStates, wantOrder)
	}
}

// TestColortagsAndMilitiaTreatmentAreInheritedUnchanged asserts the emitted
// Colortag whitelist, priority order and colour names are the reference pack's
// exactly — broad coverage on the pack's order, so no information is lost and
// the pack's existing ordering collisions stay resolved — and that the militia
// States keep both their Colortag and their Background, so lowsec behaviour is
// unchanged.
func TestColortagsAndMilitiaTreatmentAreInheritedUnchanged(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping colortag conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))
	reference := decodeGeneric(t, pack)

	for _, key := range []string{"flagStates", "flagOrder"} {
		if got, want := intList(t, emitted[key]), intList(t, reference[key]); !equalInts(got, want) {
			t.Errorf("%s = %v, want the reference pack's %v unchanged", key, got, want)
		}
	}
	if got, want := emitted["stateColorsNameList"], reference["stateColorsNameList"]; !reflect.DeepEqual(got, want) {
		t.Errorf("stateColorsNameList = %#v, want the reference pack's %#v unchanged", got, want)
	}

	// The militia States keep the pack's treatment verbatim: both remain in
	// the Colortag and Background whitelists, and the colour list above is the
	// pack's unchanged.
	flags := intList(t, emitted["flagStates"])
	backgrounds := intList(t, emitted["backgroundStates"])
	for _, state := range []int{44, 45} {
		if !contains(flags, state) {
			t.Errorf("militia State %d lost its Colortag", state)
		}
		if !contains(backgrounds, state) {
			t.Errorf("militia State %d lost its Background; the pack's militia treatment must be unchanged", state)
		}
	}
}

// TestStateBlinksAreOnlyWarCriminalAndLowSecurity asserts the emitted Blink
// set is exactly the three dangerous States — war target (13), criminal (51)
// and low security status (9) — a reduction of the reference pack's seven so
// that a Blink draws the eye.
func TestStateBlinksAreOnlyWarCriminalAndLowSecurity(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping blink conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))

	blinks := map[string]bool{}
	for _, raw := range emitted["stateBlinks"].([]any) {
		pair, ok := raw.([]any)
		if !ok || len(pair) != 2 {
			t.Fatalf("stateBlinks entry = %#v, want [key, value]", raw)
		}
		key, ok := pair[0].(string)
		if !ok {
			t.Fatalf("stateBlinks key = %#v, want a string", pair[0])
		}
		blink, ok := pair[1].(bool)
		if !ok {
			t.Fatalf("stateBlinks[%q] = %#v, want a bool", key, pair[1])
		}
		blinks[key] = blink
	}

	want := map[string]bool{"flag_9": true, "flag_13": true, "flag_51": true}
	if !reflect.DeepEqual(blinks, want) {
		t.Errorf("stateBlinks = %v, want exactly the war target, criminal and low-security Blinks %v", blinks, want)
	}
}

// hunterTabs are the Tabs the hunter flies; they drop alliance and corporation
// from their columns so the list is not padded with telemetry the reader has no
// use for mid-fight.
var hunterTabs = []string{
	"<b> Travel </b>", "<b> Engage </b>", "<b> Exit! </b>",
	"<b> Holes </b>", "<b> Fleet </b>", "<b> Occupancy </b>",
}

// inheritedTabs are the alts' Tabs and Loot, which take the profile-wide
// columns untouched.
var inheritedTabs = []string{
	"<b> PvE </b>", "<b> PI </b>", "<b> Exploring </b>", "<b> Loot </b>",
}

// hunterColumns is the per-Tab column set the hunter's Tabs override with: the
// profile-wide nine without alliance or corporation, all three velocity
// columns intact.
var hunterColumns = []string{"ANGULARVELOCITY", "DISTANCE", "ICON", "NAME", "TRANSVERSALVELOCITY", "TYPE", "VELOCITY"}

// hunterColumnOrder is the reference pack's master column order reduced to
// hunterColumns; the client derives a Tab's order from the profile-wide one, so
// this is the per-Tab order a hunter Tab must carry.
var hunterColumnOrder = []string{"ICON", "DISTANCE", "NAME", "TYPE", "VELOCITY", "TRANSVERSALVELOCITY", "ANGULARVELOCITY"}

// findPair returns a decoded ordered-pair collection's value for a key, or
// false when the key is absent — which is itself meaningful: an absent per-Tab
// column key means the Tab inherits the profile-wide columns.
func findPair(pairs []any, key string) (any, bool) {
	for _, raw := range pairs {
		pair, ok := raw.([]any)
		if !ok || len(pair) != 2 {
			continue
		}
		if pair[0] == key {
			return pair[1], true
		}
	}
	return nil, false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// TestColumnsAreProfileWideWithPerTabOverrides asserts, through the emitted
// document, that the profile shows the reference pack's nine columns and its
// master order, that the hunter's Tabs override to drop alliance and
// corporation, and that the alts' Tabs and Loot instead inherit. It reads the
// effective columns back and checks all three velocity columns survive
// everywhere.
func TestColumnsAreProfileWideWithPerTabOverrides(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping column conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))
	reference := decodeGeneric(t, pack)

	if got, want := stringList(t, emitted["overviewColumns"]), stringList(t, reference["overviewColumns"]); !equalStrings(got, want) {
		t.Errorf("overviewColumns = %v, want the reference pack's %v", got, want)
	}
	if got, want := stringList(t, emitted["columnOrder"]), stringList(t, reference["columnOrder"]); !equalStrings(got, want) {
		t.Errorf("columnOrder = %v, want the reference pack's %v", got, want)
	}

	profileColumns := stringList(t, emitted["overviewColumns"])

	tabs := map[string][]any{}
	for _, raw := range emitted["tabSetup"].([]any) {
		pairs := raw.([]any)[1].([]any)
		name, _ := findPair(pairs, "name")
		label, _ := name.(string)
		tabs[label] = pairs
	}

	for _, label := range hunterTabs {
		pairs, present := tabs[label]
		if !present {
			t.Errorf("the roster has no hunter Tab %q", label)
			continue
		}
		shown, present := findPair(pairs, "tabColumns")
		if !present {
			t.Errorf("hunter Tab %q declares no tabColumns override", label)
			continue
		}
		if got := stringList(t, shown); !equalStrings(got, hunterColumns) {
			t.Errorf("hunter Tab %q tabColumns = %v, want %v", label, got, hunterColumns)
		}
		orderValue, present := findPair(pairs, "tabColumnOrder")
		if !present {
			t.Errorf("hunter Tab %q declares no tabColumnOrder", label)
			continue
		}
		if got := stringList(t, orderValue); !equalStrings(got, hunterColumnOrder) {
			t.Errorf("hunter Tab %q tabColumnOrder = %v, want the profile order filtered to its columns %v", label, got, hunterColumnOrder)
		}
	}

	for _, label := range inheritedTabs {
		pairs, present := tabs[label]
		if !present {
			t.Errorf("the roster has no inheriting Tab %q", label)
			continue
		}
		for _, key := range []string{"tabColumns", "tabColumnOrder"} {
			if _, present := findPair(pairs, key); present {
				t.Errorf("Tab %q inherits the profile-wide columns, so %s must be absent", label, key)
			}
		}
	}

	// Every Tab keeps all three velocity columns, whichever way it gets them.
	for label, pairs := range tabs {
		effective := profileColumns
		if shown, present := findPair(pairs, "tabColumns"); present {
			effective = stringList(t, shown)
		}
		for _, column := range []string{"VELOCITY", "ANGULARVELOCITY", "TRANSVERSALVELOCITY"} {
			if !containsString(effective, column) {
				t.Errorf("Tab %q does not show %s; all three velocity columns must survive", label, column)
			}
		}
	}
}

// TestShipLabelsAreOneLineOfShipTypeAndPilotName asserts, through the emitted
// document, that the Ship Labels reduce to one line — ship type then pilot
// name — with the reference pack's alliance, corporation and linebreak
// segments dropped, so a Bracket does not occlude a busy grid and carries no
// markup that opened a segment we no longer ship.
func TestShipLabelsAreOneLineOfShipTypeAndPilotName(t *testing.T) {
	sde, pack, ok := realReferenceData(t)
	if !ok {
		conformanceSkipped = true
		t.Skip("reference data unavailable (offline?); skipping ship-label conformance — run without -short to require it")
	}

	emitted := decodeGeneric(t, generateWith(t, realSource(t), overview.Options{SDEGroups: sde, ReferencePack: pack}))

	if got, want := stringList(t, emitted["shipLabelOrder"]), []string{"ship type", "pilot name"}; !equalStrings(got, want) {
		t.Errorf("shipLabelOrder = %v, want the single-line %v", got, want)
	}

	segments, ok := emitted["shipLabels"].([]any)
	if !ok {
		t.Fatalf("shipLabels = %#v, want a sequence", emitted["shipLabels"])
	}

	seen := map[string][]any{}
	for i, raw := range segments {
		entry, ok := raw.([]any)
		if !ok || len(entry) != 2 {
			t.Fatalf("shipLabels[%d] = %#v, want [key, attributes]", i, raw)
		}
		body, ok := entry[1].([]any)
		if !ok {
			t.Fatalf("shipLabels[%d] attributes = %#v, want attribute pairs", i, entry[1])
		}
		kind, _ := findPair(body, "type")
		name, _ := kind.(string)
		seen[name] = body

		switch name {
		case "alliance", "corporation", "linebreak":
			t.Errorf("emitted Ship Label segment has type %q, which must be dropped", name)
		}
	}

	if len(segments) != 2 {
		t.Errorf("shipLabels carries %d segments, want the two of a single line", len(segments))
	}
	for _, want := range []string{"ship type", "pilot name"} {
		if _, present := seen[want]; !present {
			t.Errorf("emitted Ship Labels carry no %q segment", want)
		}
	}

	// The reference pack's pilot-name post text opened the next segment with
	// markup; with that segment dropped, no emitted segment may carry it.
	if body, present := seen["pilot name"]; present {
		post, _ := findPair(body, "post")
		text, _ := post.(string)
		if strings.Contains(text, "<") {
			t.Errorf("pilot name post = %q, want plain text with no markup opening a dropped segment", text)
		}
	}
}
