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
// source mentions through the SDE matrix.
func buildProfile(src *decodedSource, matrix *groupMatrix) (*profileDocument, error) {
	sets, err := resolveGroupSets(src.groupSets, matrix)
	if err != nil {
		return nil, err
	}

	presets, presetNames, err := buildPresets(src.presets, sets)
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

// buildPresets emits each authored preset and indexes it by its markup-free
// name, which is how tabs reference it.
func buildPresets(specs []presetSpec, sets map[string][]int) ([]pair, map[string]string, error) {
	presets := make([]pair, 0, len(specs))
	names := make(map[string]string, len(specs))
	for _, spec := range specs {
		key := stripMarkup(spec.Name)
		if _, duplicate := names[key]; duplicate {
			return nil, nil, fmt.Errorf("preset %q collides with an earlier preset once markup is stripped", spec.Name)
		}
		names[key] = spec.Name

		groups := []int{}
		for _, setName := range spec.Groups {
			set, ok := sets[setName]
			if !ok {
				return nil, nil, fmt.Errorf("preset %q references unknown Group set %q", spec.Name, setName)
			}
			groups = append(groups, set...)
		}
		presets = append(presets, pair{spec.Name, []pair{
			{"alwaysShownStates", spec.AlwaysShownStates},
			{"filteredStates", spec.FilteredStates},
			{"groups", groups},
		}})
	}
	return presets, names, nil
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
