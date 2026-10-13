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
	presets []pair
	// names indexes each loaded Preset by its markup-free name, which is how
	// a Tab or an authored Preset refers to it.
	names map[string]string
	// bodies indexes each loaded Preset's attribute pairs by markup-free
	// name, so an authored Preset can copy a loaded one's Group whitelist and
	// State lists without re-resolving the library's raw identifiers.
	bodies map[string][]any
}

// parseLibrary reads the reference pack's Presets. An empty pack is a
// legitimate empty library — the walking skeleton authors its own Presets and
// binds only those — while a pack that cannot be decoded fails loudly rather
// than emitting a profile silently missing the library.
func parseLibrary(pack []byte) (*library, error) {
	lib := &library{
		names:  map[string]string{},
		bodies: map[string][]any{},
	}
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
		lib.presets = append(lib.presets, preset)
		lib.names[stripMarkup(name)] = name
		body, ok := preset[1].([]any)
		if !ok {
			return nil, fmt.Errorf("reference pack preset %q has no attribute list", name)
		}
		lib.bodies[stripMarkup(name)] = body
	}
	return lib, nil
}

// body returns a loaded Preset's attribute pairs by markup-free name. The
// identifiers inside are the library's own and pass through verbatim: Z-S's
// data contains Groups the SDE no longer maps and repeated identifiers, and
// an authored copy inherits them as-is.
func (l *library) body(name string) ([]any, bool) {
	body, ok := l.bodies[name]
	return body, ok
}
