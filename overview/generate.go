// Package overview renders an EVE Online overview profile from the
// name-keyed source-of-truth tree.
//
// Generate is the package's single seam: a source tree in, the profile's bytes
// out. Group resolution, the internal document model and source decoding all
// sit behind it as unexported implementation, so tests observe only the
// emitted document and the returned error.
package overview

import (
	"regexp"
	"strings"
)

// Source is the raw, undecoded contents of the source-of-truth tree, one entry
// per checked-in file. Passing the bytes rather than a path keeps decoding
// behind Generate, where a test can exercise it through the seam.
type Source struct {
	Groups     []byte
	Presets    []byte
	Tabs       []byte
	Appearance []byte
	Columns    []byte
}

// Options supplies the reference data the generator reads. Tests supply their
// own bytes, so generation never depends on the machine's cache directories.
type Options struct {
	// SDEGroups is the SDE Group matrix: the JSON Lines export of CCP's
	// groups.jsonl, mapping Group names to Group identifiers.
	SDEGroups []byte

	// ReferencePack is the Z-S reference pack the emitted profile builds
	// alongside. Ticket #2's walking skeleton authors its own Presets and does
	// not yet overlay the pack; the overlay lands with the library ticket.
	ReferencePack []byte
}

// Generate renders the overview profile described by src, resolving Group
// names through the SDE matrix in opts.
//
// A Group name that does not resolve — because the SDE has no such Group, or
// because the name is ambiguous — is an error naming the offending Group. That
// failure is invisible until a fight, so it is never swallowed.
func Generate(src Source, opts Options) ([]byte, error) {
	matrix, err := parseGroupMatrix(opts.SDEGroups)
	if err != nil {
		return nil, err
	}
	decoded, err := decodeSource(src)
	if err != nil {
		return nil, err
	}
	doc, err := buildProfile(decoded, matrix)
	if err != nil {
		return nil, err
	}
	return marshalProfile(doc)
}

// markupTag matches EVE's inline markup: <color=0xAARRGGBB>, <fontsize=NN>,
// and the <b>/<i>/<u> style tags, opening or closing.
var markupTag = regexp.MustCompile(`</?(?:color(?:=0x[0-9a-fA-F]+)?|fontsize(?:=\d+)?)>|</?[biu]>`)

// stripMarkup removes EVE's inline markup, leaving the plain text a Preset or
// Tab is referenced by — the same rule the Z-S Customiser calls
// stripEveMarkup.
func stripMarkup(s string) string {
	return strings.TrimSpace(markupTag.ReplaceAllString(s, ""))
}
