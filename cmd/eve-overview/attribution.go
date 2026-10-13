package main

import (
	"fmt"

	"github.com/mgoodness/eve-overview/internal/reference"
)

// generatorURL identifies the tool that produced the profile, so a profile
// found in the wild can be traced back to its source tree.
const generatorURL = "https://github.com/mgoodness/eve-overview"

// zsPackURL is the Z-S Overview pack's public repository, which carries the
// pack's README credits and its GNU GPLv3 LICENSE.
const zsPackURL = "https://github.com/Arziel1992/Z-S-Overview-Pack"

// attributionHeader is the YAML comment block the CLI prepends to every written
// profile.
//
// The emitted profile is a derivative of the GPLv3 Z-S pack, so the pack's
// authorship, lineage and licence have to travel with the artefact rather than
// only living in this repository. The reference pins ride along too, so the
// exact bytes the profile was generated from are recoverable from the profile
// alone.
//
// The header is composed in the CLI rather than in overview.Generate so the
// package's seam stays a pure profile: Generate's output is the document and
// nothing else. The block is comments only, so the document it leads still
// decodes unchanged.
func attributionHeader() string {
	return fmt.Sprintf(`# eve-overview — generated overview profile
#
# Generator: %s
# Import:    manual and in-client; see docs/install.md
#
# Derived from the Z-S Overview pack, licensed GNU GPLv3:
#   Z-S Overview — %s
#   Zirio — YAML coding, pack maintenance, long-term updates
#   Deuce Syundai — design and text formatting
#   A continuation of the SaraShawa Overview, itself continuing
#   Sarah Schneider's Sarah's Overview Pack.
#
# EVE Online Static Data Export (SDE) © Fenris hf., used under the EVE Online
# third-party developer licence.
#
# Reproducibility — regenerate from these pinned inputs, verified by SHA-256:
#   reference pack: %s
#     sha256 %s
#   SDE build %s: %s
#     sha256 %s
#
`, generatorURL, zsPackURL,
		reference.ReferencePackURL, reference.ReferencePackSHA256,
		reference.SDEBuild, reference.SDEURL, reference.SDESHA256)
}
