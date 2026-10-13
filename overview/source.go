package overview

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// The decoded shape of the source tree. These types are unexported: the source
// is authored as YAML and only ever reaches callers of Generate as bytes.

type decodedSource struct {
	groupSets  map[string][]groupRef
	presets    []presetSpec
	tabs       []tabSpec
	appearance appearanceSpec
	columns    columnsSpec
}

// groupRef is one entry in a named Group set: either a Group name to resolve
// through the SDE matrix, or — as an escape hatch for names the matrix cannot
// disambiguate — a raw Group identifier used verbatim.
type groupRef struct {
	name string
	id   int
	raw  bool
}

func (r *groupRef) UnmarshalYAML(node *yaml.Node) error {
	if node.Tag == "!!int" {
		var id int
		if err := node.Decode(&id); err != nil {
			return err
		}
		r.id, r.raw = id, true
		return nil
	}
	var name string
	if err := node.Decode(&name); err != nil {
		return err
	}
	r.name = name
	return nil
}

type presetSpec struct {
	Name              string   `yaml:"name"`
	CopyOf            string   `yaml:"copyOf"`
	Groups            []string `yaml:"groups"`
	FilteredStates    []int    `yaml:"filteredStates"`
	AlwaysShownStates []int    `yaml:"alwaysShownStates"`
}

type tabSpec struct {
	Name    string `yaml:"name"`
	List    string `yaml:"list"`
	Bracket string `yaml:"bracket"`
}

type appearanceSpec struct {
	FlagOrder           []int           `yaml:"flagOrder"`
	FlagStates          []int           `yaml:"flagStates"`
	BackgroundOrder     []int           `yaml:"backgroundOrder"`
	BackgroundStates    []int           `yaml:"backgroundStates"`
	StateBlinks         orderedPairs    `yaml:"stateBlinks"`
	StateColorsNameList orderedPairs    `yaml:"stateColorsNameList"`
	ShipLabelOrder      []any           `yaml:"shipLabelOrder"`
	ShipLabels          []shipLabelSpec `yaml:"shipLabels"`
}

type columnsSpec struct {
	ColumnOrder     []string `yaml:"columnOrder"`
	OverviewColumns []string `yaml:"overviewColumns"`
	// TabColumns overrides the profile-wide columns for the Tabs it names, by
	// markup-free Tab name. A Tab it does not name inherits, and the client
	// keys are then absent from that Tab's entry entirely.
	TabColumns map[string][]string `yaml:"tabColumns"`
}

// shipLabelSpec is one bracket-label segment. Attribute pointers distinguish
// an explicit null (which the client writes) from an absent value.
type shipLabelSpec struct {
	Key       *string   `yaml:"key"`
	Type      *string   `yaml:"type"`
	Pre       *string   `yaml:"pre"`
	Post      *string   `yaml:"post"`
	State     *int      `yaml:"state"`
	Bold      bool      `yaml:"bold"`
	Italic    bool      `yaml:"italic"`
	Underline bool      `yaml:"underline"`
	FontSize  *int      `yaml:"fontsize"`
	Color     []float64 `yaml:"color"`
}

// orderedPairs decodes a YAML mapping while preserving the author's key order,
// which the client's exports depend on.
type orderedPairs struct {
	keys   []string
	values []any
}

func (p *orderedPairs) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping, got kind %d", node.Kind)
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		var key string
		if err := node.Content[i].Decode(&key); err != nil {
			return err
		}
		var value any
		if err := node.Content[i+1].Decode(&value); err != nil {
			return err
		}
		p.keys = append(p.keys, key)
		p.values = append(p.values, value)
	}
	return nil
}

// build turns the ordered mapping into the client's list-of-pairs form.
func (p orderedPairs) build() []pair {
	out := make([]pair, 0, len(p.keys))
	for i := range p.keys {
		out = append(out, pair{p.keys[i], p.values[i]})
	}
	return out
}

type groupsSource struct {
	GroupSets map[string][]groupRef `yaml:"groupSets"`
}

type presetsSource struct {
	Presets []presetSpec `yaml:"presets"`
}

type tabsSource struct {
	Tabs []tabSpec `yaml:"tabs"`
}

func decodeSource(src Source) (*decodedSource, error) {
	var groups groupsSource
	if err := yaml.Unmarshal(src.Groups, &groups); err != nil {
		return nil, fmt.Errorf("decode groups.yaml: %w", err)
	}
	var presets presetsSource
	if err := yaml.Unmarshal(src.Presets, &presets); err != nil {
		return nil, fmt.Errorf("decode presets.yaml: %w", err)
	}
	var tabs tabsSource
	if err := yaml.Unmarshal(src.Tabs, &tabs); err != nil {
		return nil, fmt.Errorf("decode tabs.yaml: %w", err)
	}
	var appearance appearanceSpec
	if err := yaml.Unmarshal(src.Appearance, &appearance); err != nil {
		return nil, fmt.Errorf("decode appearance.yaml: %w", err)
	}
	var columns columnsSpec
	if err := yaml.Unmarshal(src.Columns, &columns); err != nil {
		return nil, fmt.Errorf("decode columns.yaml: %w", err)
	}
	return &decodedSource{
		groupSets:  groups.GroupSets,
		presets:    presets.Presets,
		tabs:       tabs.Tabs,
		appearance: appearance,
		columns:    columns,
	}, nil
}
