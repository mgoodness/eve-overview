package overview

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// pair is one entry of the client's Python-style ordered maps: presets,
// tabSetup, shipLabels, stateBlinks and stateColorsNameList are sequences of
// [key, value] pairs rather than YAML mappings.
type pair [2]any

// profileDocument is the emitted profile. Field order is the order the client
// exports its top-level keys in, and yaml.v3 marshals struct fields in
// declaration order.
type profileDocument struct {
	BackgroundOrder     []int    `yaml:"backgroundOrder"`
	BackgroundStates    []int    `yaml:"backgroundStates"`
	ColumnOrder         []string `yaml:"columnOrder"`
	FlagOrder           []int    `yaml:"flagOrder"`
	FlagStates          []int    `yaml:"flagStates"`
	OverviewColumns     []string `yaml:"overviewColumns"`
	Presets             []pair   `yaml:"presets"`
	ShipLabelOrder      []any    `yaml:"shipLabelOrder"`
	ShipLabels          []pair   `yaml:"shipLabels"`
	StateBlinks         []pair   `yaml:"stateBlinks"`
	StateColorsNameList []pair   `yaml:"stateColorsNameList"`
	TabSetup            []pair   `yaml:"tabSetup"`
	UserSettings        []any    `yaml:"userSettings"`
}

// buildProfile assembles the profile document, resolving every Group name the
// source mentions through the SDE matrix and overlaying the authored Presets
// on the loaded Z-S library.
func buildProfile(src *decodedSource, matrix *groupMatrix, lib *library) (*profileDocument, error) {
	sets, err := resolveGroupSets(src.groupSets, matrix)
	if err != nil {
		return nil, err
	}

	presets, presetNames, err := buildPresets(src.presets, sets, lib)
	if err != nil {
		return nil, err
	}

	tabs, err := buildTabs(src.tabs, presetNames)
	if err != nil {
		return nil, err
	}

	return &profileDocument{
		BackgroundOrder:     src.appearance.BackgroundOrder,
		BackgroundStates:    src.appearance.BackgroundStates,
		ColumnOrder:         src.columns.ColumnOrder,
		FlagOrder:           src.appearance.FlagOrder,
		FlagStates:          src.appearance.FlagStates,
		OverviewColumns:     src.columns.OverviewColumns,
		Presets:             presets,
		ShipLabelOrder:      src.appearance.ShipLabelOrder,
		ShipLabels:          buildShipLabels(src.appearance.ShipLabels),
		StateBlinks:         src.appearance.StateBlinks.build(),
		StateColorsNameList: src.appearance.StateColorsNameList.build(),
		TabSetup:            tabs,
		UserSettings:        []any{},
	}, nil
}

// resolveGroupSets turns each named set of Group names and raw identifiers
// into the identifier list a preset's whitelist is built from.
func resolveGroupSets(sets map[string][]groupRef, matrix *groupMatrix) (map[string][]int, error) {
	resolved := make(map[string][]int, len(sets))
	for name, refs := range sets {
		ids := make([]int, 0, len(refs))
		for _, ref := range refs {
			if ref.raw {
				ids = append(ids, ref.id)
				continue
			}
			id, err := matrix.resolve(ref.name)
			if err != nil {
				return nil, fmt.Errorf("Group set %q: %w", name, err)
			}
			ids = append(ids, id)
		}
		resolved[name] = ids
	}
	return resolved, nil
}

// buildPresets emits the loaded library unchanged, then appends each authored
// preset and indexes the combined library by markup-free name, which is how
// tabs and authored presets reference a preset. An authored name that collides
// with a loaded one fails loudly: silently overriding would drop a Z-S Preset
// the profile promises to carry.
func buildPresets(specs []presetSpec, sets map[string][]int, lib *library) ([]pair, map[string]string, error) {
	presets := make([]pair, 0, len(lib.presets)+len(specs))
	names := make(map[string]string, len(lib.names)+len(specs))
	for name, emitted := range lib.names {
		names[name] = emitted
	}
	presets = append(presets, lib.presets...)

	for _, spec := range specs {
		key := stripMarkup(spec.Name)
		if existing, collision := names[key]; collision {
			return nil, nil, fmt.Errorf("preset %q collides with preset %q once markup is stripped", spec.Name, existing)
		}
		names[key] = spec.Name

		groups := []int{}
		filtered := spec.FilteredStates
		always := spec.AlwaysShownStates
		if spec.CopyOf != "" {
			base, ok := lib.body(spec.CopyOf)
			if !ok {
				return nil, nil, fmt.Errorf("preset %q copies unknown library preset %q", spec.Name, spec.CopyOf)
			}
			// The base is a loaded Z-S Preset: its Groups ride through as raw
			// identifiers and its State lists are inherited, so a copy is the
			// exploration Preset plus hazards rather than a rebuild.
			groups = append(groups, presetGroups(base)...)
			if filtered == nil {
				filtered = presetStates(base, "filteredStates")
			}
			if always == nil {
				always = presetStates(base, "alwaysShownStates")
			}
		}
		for _, setName := range spec.Groups {
			set, ok := sets[setName]
			if !ok {
				return nil, nil, fmt.Errorf("preset %q references unknown Group set %q", spec.Name, setName)
			}
			groups = append(groups, set...)
		}
		presets = append(presets, pair{spec.Name, []pair{
			{"alwaysShownStates", always},
			{"filteredStates", filtered},
			{"groups", groups},
		}})
	}
	return presets, names, nil
}

// presetGroups reads a loaded Preset body's Group identifier list. Its
// identifiers are returned exactly as the library carried them.
func presetGroups(body []any) []int {
	groups, ok := presetAttribute(body, "groups")
	if !ok {
		return nil
	}
	return asIntList(groups)
}

// presetStates reads one of a loaded Preset body's State lists.
func presetStates(body []any, key string) []int {
	value, ok := presetAttribute(body, key)
	if !ok {
		return nil
	}
	return asIntList(value)
}

func presetAttribute(body []any, key string) (any, bool) {
	for _, raw := range body {
		entry, ok := raw.([]any)
		if !ok || len(entry) != 2 {
			continue
		}
		if entry[0] == key {
			return entry[1], true
		}
	}
	return nil, false
}

// asIntList converts a decoded YAML sequence to the integer list the profile
// document carries, dropping anything that is not an integer.
func asIntList(value any) []int {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]int, 0, len(items))
	for _, item := range items {
		n, ok := item.(int)
		if !ok {
			continue
		}
		out = append(out, n)
	}
	return out
}

// buildTabs emits one tabSetup slot per authored tab. The list preset is
// written with its colour markup; the bracket preset is referenced markup-free,
// which is the client's convention.
func buildTabs(specs []tabSpec, presetNames map[string]string) ([]pair, error) {
	tabs := make([]pair, 0, len(specs))
	for index, spec := range specs {
		listName, ok := presetNames[spec.List]
		if !ok {
			return nil, fmt.Errorf("tab %q binds unknown list preset %q", spec.Name, spec.List)
		}
		bracketName, ok := presetNames[spec.Bracket]
		if !ok {
			return nil, fmt.Errorf("tab %q binds unknown bracket preset %q", spec.Name, spec.Bracket)
		}
		tabs = append(tabs, pair{index, []pair{
			{"bracket", stripMarkup(bracketName)},
			{"color", nil},
			{"name", spec.Name},
			{"overview", listName},
		}})
	}
	return tabs, nil
}

// buildShipLabels emits each segment's attributes in the fixed order the
// client's own exports use, with bold/italic/underline written as 1/0.
func buildShipLabels(specs []shipLabelSpec) []pair {
	labels := make([]pair, 0, len(specs))
	for _, spec := range specs {
		labels = append(labels, pair{optString(spec.Key), []pair{
			{"type", optString(spec.Type)},
			{"pre", optString(spec.Pre)},
			{"post", optString(spec.Post)},
			{"state", optInt(spec.State)},
			{"bold", boolInt(spec.Bold)},
			{"italic", boolInt(spec.Italic)},
			{"underline", boolInt(spec.Underline)},
			{"fontsize", optInt(spec.FontSize)},
			{"color", optColor(spec.Color)},
		}})
	}
	return labels
}

func marshalProfile(doc *profileDocument) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("encode profile: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("encode profile: %w", err)
	}
	return buf.Bytes(), nil
}

func optString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func optInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func optColor(value []float64) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
