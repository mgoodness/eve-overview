package overview

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// library is the Z-S Preset library the emitted profile is built alongside.
// The generator overlays the source's authored Presets on top of it: the
// loaded Presets ride through untouched, and the source only adds.
//
// A Preset's Group whitelist is never resolved here. Z-S's own data contains
// identifiers the SDE no longer maps (1211, 4748) and repeated identifiers
// (its collidables Preset lists thirty duplicates); those are the library's
// data and are not ours to repair. Only a Group name the *source* cannot
// resolve is fatal.
type library struct {
	// order is the loaded Presets' markup-free keys in pack order, so the
	// emitted document keeps the order the client expects.
	order []string
	// byName holds each loaded Preset by its markup-free key — the name a Tab
	// or an authored Preset refers to it by. Each Preset's shipped name and
	// body live together in the one value, so nothing is keyed twice.
	byName map[string]loadedPreset
}

// loadedPreset is one Preset the reference pack ships: its markup-bearing name
// exactly as shipped, and the [name, body] entry that is emitted verbatim.
type loadedPreset struct {
	name  string
	entry pair
}

// body returns the Preset's attribute pairs, read by an authored Preset that
// copies this one.
func (p loadedPreset) body() []any {
	body, _ := p.entry[1].([]any)
	return body
}

// parseLibrary reads the reference pack's Presets. An empty pack is a
// legitimate empty library — the walking skeleton authors its own Presets and
// binds only those — while a pack that cannot be decoded fails loudly rather
// than emitting a profile silently missing the library.
func parseLibrary(pack []byte) (*library, error) {
	lib := &library{byName: map[string]loadedPreset{}}
	if len(bytes.TrimSpace(pack)) == 0 {
		return lib, nil
	}

	var doc struct {
		Presets []pair `yaml:"presets"`
	}
	if err := yaml.Unmarshal(pack, &doc); err != nil {
		return nil, fmt.Errorf("decode reference pack: %w", err)
	}

	for i, preset := range doc.Presets {
		name, ok := preset[0].(string)
		if !ok {
			return nil, fmt.Errorf("reference pack preset %d has no name", i)
		}
		if _, ok := preset[1].([]any); !ok {
			return nil, fmt.Errorf("reference pack preset %q has no attribute list", name)
		}
		key := stripMarkup(name)
		lib.order = append(lib.order, key)
		lib.byName[key] = loadedPreset{name: name, entry: preset}
	}
	return lib, nil
}

// body returns a loaded Preset's attribute pairs by markup-free name. The
// identifiers inside are the library's own and pass through verbatim: Z-S's
// data contains Groups the SDE no longer maps and repeated identifiers, and
// an authored copy inherits them as-is.
func (l *library) body(name string) ([]any, bool) {
	preset, ok := l.byName[name]
	if !ok {
		return nil, false
	}
	return preset.body(), true
}
