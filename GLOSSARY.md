# eve-overview

A custom EVE Online **overview profile** tuned for a mostly-solo, cloaky,
mostly-wormhole PvP hunter, plus the decisions and tooling that produce it.

## Language

### The artefact

**Overview profile**:
The account-wide YAML document configuring the in-game overview: presets, tab
bindings, columns, colortag and background priority, bracket labels. The whole
deliverable; inert display data with no execution surface.
_Avoid_: overview, overview settings, overview file, overview pack

**Overview pack**:
A named, distributable profile authored by someone else and installed as a
starting point — Z-S Overview, the EVE University overview.
_Avoid_: overview, profile, filter pack

**Preset**:
A named rule deciding which objects may render: a group whitelist, a state veto,
and a state override that rescues entities from that veto.
_Avoid_: filter, overview filter, d-scan filter

**Tab**:
One of up to twenty slots in the overview window, binding a list preset and a
bracket preset — independently — and optionally overriding the profile's columns.
_Avoid_: overview tab, window, page

**Bracket**:
The floating icon drawn in space for an object, filtered by a tab's bracket
preset rather than its list preset.
_Avoid_: label, marker, icon

**Binding**:
The pairing of a list preset and a bracket preset to one tab slot. The two are
chosen independently, which is what lets a tab hide friendlies from the list it
reads while still drawing their brackets in space.
_Avoid_: assignment, mapping

### Relationship to the client

**Profile scope** (account-wide):
The client stores one overview profile per *account*, not per character — every
character on the account shares its presets, tab bindings, colortags, columns and
bracket labels. The twenty-tab cap is therefore a budget shared by all of them.
_Avoid_: per-character settings, character settings

**Friendly** (colloquially "blue"):
The state set marking pilots you must not shoot: fleet (11), corp (12), alliance
(14), standings (15/16), militia (45), war allies (49). A preset interacts with
them two distinct ways — **vetoed** (hidden from a list preset) or **colortagged**
(visible, drawn in blue) — so unqualified "blue" is ambiguous here; say which.
_Avoid_: blue, blues, friendly (unqualified)

**State**:
The client's classification of an entity's relationship or legal status (pilot,
hostile, criminal, war target, fleet member, …), which a profile colours rather
than filters. CCP's UI calls these "Exceptions".
_Avoid_: exception, standing, relation

**Colortag**:
The small coloured icon drawn beside an entity to communicate its matched state.
_Avoid_: flag, tag, colour

**Background**:
The full-row highlight drawn for an entity whose matched state appears in the
profile's background order.
_Avoid_: row colour, highlight, shading

### Sensors

**D-scan** (directional scanner):
The manual, 14.3 AU, direction-only sweep. Its selector picks a preset, and it
ignores states entirely.
_Avoid_: scanner, directional scan, probe scan

**Probe scan**:
The scanner window's probe-based search for signatures and anomalies. Not part of
an overview profile.
_Avoid_: scan, scanning, d-scan

### Space

**Wormhole space** (J-space):
The systems reachable only through wormholes, where Local chat carries no pilots
and the overview and d-scan are the only intel surfaces.
_Avoid_: WH, w-space, unknown space

**Known space** (K-space):
High-, low- and nullsec.
_Avoid_: normal space, empire space

**Collidable**:
An object with physical volume; being within 2,000 m of one decloaks a cloaked
ship. The rule covers *any* collidable, which is why collidable group lists
balloon into the hundreds.
_Avoid_: obstacle, obstacle object

**Decloak hazard**:
The subset of collidables a hunter actually meets on grid — structures, force
fields, control towers, player deployables. Carried on the working tabs so a cloak
is never lost by bumping something the overview never showed.
_Avoid_: obstacle, hazard

### Data

**SDE** (Static Data Export):
CCP's dump of every game type; the source of the `typeID → groupID → categoryID`
hierarchy that presets whitelist by.
_Avoid_: database, datadump, static data

**Group**:
The SDE level a preset whitelists at — *Frigate*, *Stargate*, *Citadel* — so one
entry admits every type in that group. Type-level whitelisting is not established
as expressible and is not built on.
_Avoid_: type, category, class

### Role

**Hunter**:
A pilot who seeks PvP by locating other players; in this project, one flying
cloaked and mostly alone. Z-S separately uses the word for a *d-scan preset* that
whitelists presence-implying objects — an overload, disambiguated as "Hunter
preset" for the preset and "the hunter" for the pilot.
_Avoid_: roamer, ganker, PvPer
