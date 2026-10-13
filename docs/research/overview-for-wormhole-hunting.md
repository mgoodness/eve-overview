# Overview design for a mostly-solo, cloaky, mostly-wormhole PvP hunter

Research note for `eve-overview`. Written 2026-10-12.

**Destination:** there is no existing research-note convention in this repo (only
`docs/agents/` exists, per `AGENTS.md`), so this file creates one:
`docs/research/<slug>.md`.

**Trust tiers used below**

| Tag | Meaning |
| --- | --- |
| `[MECH]` | Game/client behaviour, traceable to CCP documentation, client-exported data, or the SDE. Verifiable. |
| `[PRAC]` | Established community practice or pack design. Defensible but a choice, not a rule. Marked as opinion. |
| `[ANEC]` | Forum/Reddit anecdote. Lower trust; used only where no primary source exists, and always labelled. |
| `[GAP]` | Not found in any primary source. Stated as unresolved rather than guessed. |

**Sources pinned in play**

| Source | Version / commit |
| --- | --- |
| `github.com/Arziel1992/Z-S-Overview-Pack` | `9c7dd45` (2019-08-27), README says "Current Version: v9.00.0347, Updated 2019.05.06"; 66 commits; authors by commit count: Arziel 47, waffle911 14, Justin Burnham 5 |
| `gitlab.com/Arziel/Z-S-Overview-Pack` | mirror; 1 branch (`master`), 21 tags, newest tag `v9.00.0347` |
| `github.com/Arziel1992/Z-S-Overview-Customizer` | `40a4bc3` (2026-10-12). Bundles `public/defaults/zs_full_v10.06.09.yaml`, `zs_full_v9.00.0347.yaml`, `fenris_default_v24.01.yaml`, and `public/data/matrix_latest.json` (`metadata.version: "SDE-Latest-Official"`, compiled ~2026-10) |
| `support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings` | CCP, updated 2026-06-01 |
| `support.eveonline.com/hc/en-us/articles/203209902-Scanning` | CCP, updated 2021-05-24 |
| `eveonline.com/news/view/sharing-is-caring` (Hyperion overview dev blog) | CCP, 2014-08-20 |
| `eveonline.com/news/view/citadels-sieges-and-you` | CCP Ytterbium, 2015-08-13 |
| `eveonline.com/news/view/nullsec-local-blackout-in-effect` | CCP Falcon, 2019-07-12 |
| EVE University Wiki: Overview, Overview settings, Overview manipulation, Directional scanning, Grid, Wormhole space, Wormholes, Wormhole scouting, Cloaking, Scanning, Bookmarks, Upwell structure | as served 2026-10-12 |

Numbers I computed myself (preset group counts, preset inventories, v9↔v10 diff)
are labelled **derived** and name the exact file they came from.

---

## 1. The overview's role in wormhole space

### 1.1 There is no useful Local in J-space

`[MECH]` CCP, describing the 2019 nullsec blackout, states the wormhole baseline
directly: delayed-mode local "**will behave as local chat in wormhole space, with
pilots only appearing in the local population listing should they choose to post
messages**"
([eveonline.com/news/view/nullsec-local-blackout-in-effect](https://www.eveonline.com/news/view/nullsec-local-blackout-in-effect)).

`[MECH]` EVE University: "Local chat in W-Space is different from K-Space. You or
any other players will only show up in channel if you're actually chatting…
Therefore to see who's in system you need to use scouts and D-Scan. **Even then a
cloaky ship won't be noticeable until it decloaks or logs off.**"
([wiki.eveuniversity.org/Wormhole_space](https://wiki.eveuniversity.org/Wormhole_space)).

Consequence for a hunter: neither intel surface is complete, and both are
*channels the hunter must poll manually*. Nothing in J-space pushes a "someone
arrived" event.

### 1.2 What each surface actually is

**Overview `[MECH]`**

- It is a **grid-scoped, continuously live** list. "Anything uncloaked within this
  cube will be visible in the overview to all pilots within that same cube. An
  object located outside of the grid will not be visible in the overview."
  Default grid is a cube ~8,000 km each way, but variable — the Jita 4-4 grid
  runs ~20,000 km
  ([Grid](https://wiki.eveuniversity.org/Grid)).
- It is a **filtered view of the client's object list**, not a scan: CCP's
  Overview Settings article describes Filters (types) + Exceptions (states) →
  Colortags/Backgrounds → Columns → per-Tab bracket filter
  ([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)).
- Per-row telemetry includes **pilot name**, corporation, alliance, faction,
  militia, type, size, distance, velocity, radial/transversal/angular velocity,
  and fleet tag ([Overview settings](https://wiki.eveuniversity.org/Overview_settings)).
  The pilot name is the decisive detail: the overview tells you *who* is home.
- Rows are interactive: right-click, radial menu, and shortcut-key actions
  (ctrl-click to lock, double-click to approach/align). While a shortcut key is
  held "the overview will be frozen to prevent items from shuffling around based
  on the sorting order" ([Overview](https://wiki.eveuniversity.org/Overview)).
- Brackets in space are a **separate filter** from the list: "Filters control what
  shows up in your Overview window. Bracket Filters control what shows up on your
  actual in-game screen." Same source.

**Directional scan `[MECH]`**

- A cone or sphere from your ship, up to **14.3 AU**, angle from **5° to 360°**,
  refreshed manually; "After a scan it takes two seconds for the scanner to
  reset" ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
- It returns **icon, name and type only**: "It does not show distance or
  coordinates"; and "the scanner only returns ship names and types; **never the
  pilot's name**". Same source.
- It sees: all ships *except* cloaked ships and combat recons; celestials;
  structures including ones not on your overview; wrecks; drones; core and combat
  probes; and NPCs whose names carry a **♦** ("mining operations, Blood Raider
  shipyard fleets and NPC ships in the AEGIS… facilities"). Same source.
- It cannot see: object state (empty/damaged/reinforced/pilot info), cosmic
  signatures, cosmic anomalies, normal NPCs, beacons or bookmarks. Same source.
- Force fields are their own d-scan object: "The force field of an active control
  tower is separate object in directional scan and can tell you if anyone is
  living in the system." Same source.
- Upwell structures: "While undockable structures aren't visible in overview they
  are visible in directional scan." Same source — consistent with §3.2 below.

### 1.3 The complementarity, stated precisely

| Question | Overview | D-scan |
| --- | --- | --- |
| Where? | Objects **on your grid**; real-time, automatic | Anything within **14.3 AU**, direction only (cone/bearing) |
| Who? | **Pilot name**, corp, alliance, standings-derived state | Ship *name* and *type* only, never pilot name |
| How far / how fast? | Distance, velocity, transversal, angular | Nothing — no distance, no coordinates |
| Off-grid presence? | No | Yes (ships, wrecks, drones, probes, force fields, structures) |
| On-grid-only objects? | Yes — including **Sleepers**, which the d-scan cannot see (no ♦) | No |
| State/relationship? | Yes — colortags/backgrounds driven by ~24 states | No — state is explicitly invisible to the scanner |
| Signatures / anomalies / bookmarks? | No | No (probe scanner / Locations window, not the overview) |

The practical division of labour for a solo cloaky hunter `[PRAC]`:

1. **D-scan answers "is anyone out there, and roughly where"** before you commit.
   EVE Uni's wormhole-scouting procedure opens with "Ping d-scan, 360°, max range,
   using Wormhole Overview Preset"
   ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)).
2. **The overview answers "who exactly is on this grid, at what range, doing what"**
   once you land — and it is the only surface that shows pilot identity and speed.
3. **The two are deliberately different filters.** A tab that is right for one is
   wrong for the other, which is exactly why the d-scan has its own filter
   selector and why packs ship separate `D-Scan:` presets (§2.3, §4.2).

The J-space twist that makes this matter more than in K-space: the off-grid tells
are *entirely* d-scan (probes, wrecks, force fields, mobile depots, bubbles), and
the on-grid tells are *entirely* overview. There is no Local to backstop either.

---

## 2. Flags/states and presets a wormhole hunter actually uses

### 2.1 "States" (CCP's "Exceptions") vs "flags" (colortags) vs preset filters

These are three different things that get conflated in casual use.

**Preset filter — the outer gate.** `[MECH]` A preset is a **type whitelist**
plus **state logic**. In the exported YAML a preset is three arrays:
`groups`, `filteredStates`, `alwaysShownStates`
([Z-S Customiser README, "Preset variables"](https://github.com/Arziel1992/Z-S-Overview-Customizer)).
CCP words the types half as: "Within the 'Filters' sub tab, the types of items in
space that the overview will show can be selected"
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)).

Critical mechanic: **the type whitelist and the state logic are independent gates,
and the whitelist wins.** The Customiser states it as: "A hull whose group is not
listed never renders, and **no state overrides this**… This follows the client's
own wording for the always-shown column… so the override is scoped to states, not
to the type filter." A wrongly-absent ship group means a war target you simply
cannot see. EVE Uni's blunt version: "A misconfigured overview can result in the
loss of a ship or worse, a diplomatic incident of friendly fire"
([Overview](https://wiki.eveuniversity.org/Overview)).

**States — the built-in classification of an object.** `[MECH]` CCP calls them
**Exceptions**: "An Exception describes the relationship between the current
character and the entity in question, **usually another pilot but can also be an
Agent or a wreck**. The Exceptions are hierarchical with the highest priority
state at the top of the list, as some entities can match multiple criteria…"
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)).
EVE Uni describes the three radio columns: **Always show / Filter out / Show by
default**, with the precedence rules spelled out
([Overview settings](https://wiki.eveuniversity.org/Overview_settings)).

The state inventory is small, fixed, integer-ID'd, and not extensible. The
Customiser publishes the table (attributing it to Tomas Iridium's
`states_all.yaml`), and I reproduce the ones that matter to a hunter from that
file (verbatim meanings, abridged):

| ID | Meaning | Hunter relevance |
| --- | --- | --- |
| 9 | sec status below −5 | outlaw; treat as hostile on sight |
| 10 | sec status below 0 | |
| 11 | in your fleet | filter out to de-clutter a combat tab |
| 12 | in your capsuleer corporation | |
| 13 | at war with your corp/alliance | **always-show** on any PvP preset |
| 14 | in your alliance | |
| 15/16/17/18/19/48 | Excellent / Good / Neutral / Bad / Terrible / No standing | in J-space, standings are *your* bookkeeping — useful for marking known locals |
| 21 | Pilot (agent) is interactable | |
| 36 / 37 | Wreck already viewed / Wreck is empty | **filter-only** — never a colortag or background |
| 44 / 45 | at war with your militia / in your militia or allied | meaningless in J-space; can steal colortag priority (§3.6) |
| 49 | ally in one or more of your wars | |
| 50 / 51 | suspect / criminal | |
| 52 | limited engagement with you | |
| 53 | killright you can activate | |
| 66 | in your Non Capsuleer corporation | |
| 68 | retribution timer | |

Source: [Customiser README, "The state ID table"](https://github.com/Arziel1992/Z-S-Overview-Customizer).
The client honours 36/37 in `filteredStates`/`alwaysShownStates` only, never in
`flagStates`/`backgroundStates`.

**Flags — how a matched state is *drawn*.** `[MECH]` CCP: colortags are "the tiny
icons displayed in the corner of the entity icon… Just like States, they are
hierarchical"; backgrounds "are applied to the background color of the whole row…
much more useful for indicating danger than the Colortags"; both can blink. "The
order of tags in the list dictates which tag is shown if multiple are valid. The
tag that is higher on the list overrides tags that are below it."
([CCP](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings),
[EVE Uni](https://wiki.eveuniversity.org/Overview_settings)).

So in the YAML the two families are:

- **Appearance (drawing)**: `flagStates` + `flagOrder`, `backgroundStates` +
  `backgroundOrder`, `stateBlinks`, `stateColorsNameList`.
- **Filtering (showing)**: `filteredStates` + `alwaysShownStates` per preset.

That is the exact distinction the brief asked for: **states/Exceptions decide
whether an object renders and what rank it has; flags (colortags/backgrounds)
decide only how a rendered object looks, walked top-to-bottom, first match wins.**

### 2.2 The presets a wormhole hunter runs `[PRAC]`

There is no canonical published "wormhole hunter overview" from CCP. The
following is assembled from (a) what the packs ship, (b) EVE Uni's scouting
procedure, and (c) hunter practice. Marked as practice.

| Role | Filter shape | Why |
| --- | --- | --- |
| **Engage / combat** | All player ship groups + capsules + drones/fighters + probes; **filter out** 11/12/14/15/16/45/49 (fleet, corp, alliance, friendly standings); **always show** 13/44 (war) where relevant | The list you actually shoot from. Friendly-state filtering prevents the classic misclick; see Z-S `✜ --- PvX: Basic` presets, all of which filter `[11,12,14,15,16,45,49]` (derived from `zs_full_v10.06.09.yaml`). |
| **D-scan (offensive)** | Player-presence objects only, **no state logic at all** | See §2.3. EVE Uni's scouting flow uses it at 360°/max range as the first act in a new system. |
| **Travel / warp-to** | Celestials + gates/stations/structures you can warp to | Warping and aligning, and finding the hole. Z-S ships `✈ --- Travel: All`, `✈ Travel (+Ship -Sta)`, `✈ Travel (-Ship +Sta)`, `➲ Extra: Align Points`, `➲ Extra: Podsaver (!)`. |
| **Warp-to specifically (hole/pod out)** | The shortest possible list: sun, planet, moon, belt, gate, wormhole, disrupted gate | Z-S `➲ Extra: Podsaver (!)` is 9 groups: `Sun, Planet, Moon, Asteroid Belt, Stargate, Wormhole, Mercenary Bases, Capsuleer Bases, Disrupted Gate` (derived). It is deliberately free of player-owned entities, which is why EVE Uni also recommends it for flushing stale ship-label formatting out of the client: "The podsaver tab is excellent for this, as it shouldn't contain any player entities" ([Overview manipulation](https://wiki.eveuniversity.org/Overview_manipulation)). |
| **Loot / salvage** | Wrecks, cargo/secure/audit containers, freight containers, salvage drones, mobile tractor units + the ships that loot | Z-S `✜ PvX: Salvage/Loot` is 529 groups including `Wreck`, `Cargo Container`, `Secure Cargo Container`, `Audit Log Secure Container`, `Freight Container`, `Scatter Container`, `Salvage Drone`, `Mobile Tractor Unit` (derived). The wreck-only states 36/37 let you hide already-viewed/empty wrecks on this tab. |
| **CovOps collidables** | Every large collidable object, no state filtering | Z-S `✜ = CovOps: Collidables` is 675 groups: `Large Collidable Object`, `Large Collidable Structure`, `Control Tower`, `Force Field`, cans, drones, corpses (`Biomass`), asteroids, even scenery (`Billboard`, `Landmark`) (derived). Exists specifically because of the 2000 m decloak rule — see §5.1. |
| **POS / occupancy finder** | Moons + force field + mobile warp disruptor | EVE Uni's scouting guide gives an explicit recipe: "Open the Celestial folder and check Moon. Open the Deployable folder and ensure that Mobile Warp Disruptor is selected. Save the settings as 6a - POS finder." with the warning that it "does add clutter… should not be used while in combat situations" ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)). |

Tab budget `[MECH]`: the client supports **up to 20 tabs** — "up to 20 Tabs can
be configured to be shown within the overview"
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings),
updated 2026-06-01). EVE Uni's "up to eight different tabs" is stale
([Overview settings](https://wiki.eveuniversity.org/Overview_settings)); CCP's
2014 Hyperion dev blog said five tabs / up to 15 tab presets, which was also a
then-current figure
([Hyperion](https://www.eveonline.com/news/view/sharing-is-caring)). Where sources
conflict, the CCP support article wins — it is the most recent first-party
statement. Having 20 available is not a reason to use them `[PRAC]`: a hunter tab
set that cannot be traversed in two seconds under fire is a liability.

### 2.3 D-scan filters are a *different kind* of filter `[MECH]`

"The d-scan can use any of your overview presets to filter the results… note that
the scanner does not see any information about the objects other than name and
type. This means that **the filtering function will not take into account any of
the state settings the overview may have**"
([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).

Two consequences:

1. A d-scan preset's `filteredStates`/`alwaysShownStates` are **inert**. Z-S's four
   D-Scan presets in `zs_full_v10.06.09.yaml` all carry `filteredStates: []` and
   `alwaysShownStates: []` (derived) — Z-S designs them as pure type lists, which
   matches the stated mechanic.
2. You cannot de-clutter d-scan with "hide blues". You must remove friendly *ship
   groups* instead, which is why Z-S ships paired variants rather than state logic.

---

## 3. Wormhole-specific pitfalls

### 3.1 What a J-space hunter must be able to see

| Object class | Overview | D-scan | Verdict for a WH hunter |
| --- | --- | --- | --- |
| **POS / Starbase (Control Tower, modules, Force Field)** | Yes, group `Control Tower` (365), `Force Field` (411) still present in the current SDE | Yes — force field is its own d-scan line | **Force field: keep, always.** It is the cheapest single occupancy tell in J-space. Full POS module bloat: keep only on the POS-finder tab. |
| **Upwell structures** | Only with docking access, or when on grid | **Always**, and warpable without probes | **Keep in d-scan presets, exclude from the combat tab.** See §3.2. |
| **Sleeper NPCs** | Yes, when on grid (`Entity` category; group `Sleeper Entities` = 4036, plus the ♦-prefixed special-NPC groups) | **No** — ordinary NPCs are invisible to d-scan | **Exclude from the combat/d-scan tabs.** They cannot be seen off-grid, so they are useless as a "site is active" tell; the usable tell is the *wrecks*, which are on d-scan. |
| **Sleeper wrecks** | Yes | Yes | **Keep on a looting tab and in the d-scan Hunter preset** — "if any are visible there has been someone fighting recently" ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)). |
| **Cynosural fields** (`Cynosural Fields` 2020, `Covert Cynosural Field` 897, `Upwell Cyno Beacon`, `Mobile Cynosural Beacon`) | Yes | Yes | **Keep.** In J-space a cyno can be lit but jump drives cannot lock it — "Cynosural fields can be opened in wormhole space, but ships with jump drives cannot lock on to these fields, nor can they lock on to cynosural fields outside of wormhole space to jump out" ([Wormholes](https://wiki.eveuniversity.org/Wormholes)). So a cyno you see is a *presence/practice* signal, not an inbound capital. |
| **Mobile depots and other deployables** (`Mobile Depot` 1246, `Mobile Cyno Inhibitor`, `Mobile Tractor Unit`, `Mobile Siphon Unit`, `Encounter Surveillance System`, `Mobile Scan Inhibitor`) | Yes | Yes | **Keep on both.** Presence tell *and* collidable-that-can-decloak-you. Z-S's `D-Scan: Hunter (+Dro)` includes `Mobile Depot, Mobile Siphon Unit, Mobile Cyno Inhibitor, Mobile Tractor Unit, Encounter Surveillance System, Mobile Scan Inhibitor, Mobile Micro Jump Unit` (derived). |
| **Warp disruption fields / bubbles** (`Mobile Warp Disruptor` 361, `Mobile Jump Disruptor`, `Interdiction Burst Probes`) | Yes | Yes | **Keep.** EVE Uni's scouting procedure treats them as a trap alarm: "D-scan the moon one last time, if there are any warp disruptors in d-scan that implies a trap has been set" ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)). |
| **Wrecks** | Yes | Yes | Split: keep in d-scan Hunter + looting tab, exclude from the combat tab. Wreck states 36/37 are the only per-wreck filtering available. |
| **Faction warfare anything** | Groups exist (`FW *` ship groups, `FW Infrastructure Hub`, `FW Propaganda Broadcast Structure`, `Control Point`) | n/a | **Drop from J-space profiles.** There are no FW complexes in J-space, so these groups buy nothing; and the *militia states* (44/45) can collide with friendly priority — see §3.6. |

### 3.2 Upwell structures: the counter-intuitive one `[MECH]`

CCP's own pre-release design statement: "Structure brackets are going to be
visible in space and in the overview **if: The user can interact with the
structure in question**… On top of the points above, **all structures, no matter
their size or role, will have warpable signatures like cosmic combat anomalies.
None of them will need to be probed to be warped to, even if the user doesn't have
direct access to them.**" ([Citadels, sieges and you](https://www.eveonline.com/news/view/citadels-sieges-and-you)).

EVE Uni's current operational wording: "While you are in space your overview will
show all citadels that you have docking rights for. Citadels that are on grid are
shown even if no docking rights are available. Citadels to which you don't have
docking rights are visible in solar map in brackets."
([Upwell structure](https://wiki.eveuniversity.org/Upwell_structure)).

Net effect: **an enemy's home citadel is invisible in your overview until you are
on its grid, but it always appears on your d-scan.** For a hunter this inverts the
usual rule — d-scan, not the overview, is your structure inventory. `[ANEC]`
Players report the on-grid threshold as roughly 2,000–10,000 km in practice
([forum thread](https://forums.eveonline.com/t/any-high-security-systems-where-player-owned-structures-not-allowed/73155));
the exact figure is not in any primary source I found `[GAP]`.

### 3.3 POS/starbase drift in the packs, and whether POSes still exist

`[MECH]` The legacy `Starbase` category (SDE category 23) and the `Control Tower`
group (365) are **still present in the current official SDE** as compiled into the
Customiser's `matrix_latest.json` (26 groups, e.g. `Control Tower` 44 types,
`Force Field` is `Celestial` group 411). I found **no CCP patch note removing
Player Owned Starbases** `[GAP]`; a 2021 patch note still references a
"Minimum distance from Control Towers: 50 km" for Mobile Cynosural Beacons
([Patch Notes 19.06](https://www.eveonline.com/news/view/patch-notes-version-19-06)).

`[MECH]` Z-S nonetheless dropped its POS-specific d-scan presets: v9.00.0346
(April 2019) is titled "**No more POS's!**" and records removing
`※ D-Scan: POS (+Cel +Mod)`, `※ D-Scan: POS (+Cel -Mod)`, `※ D-Scan: POS (-Cel +Mod)`
([Z-S CHANGELOG.md](https://github.com/Arziel1992/Z-S-Overview-Pack/blob/master/CHANGELOG.md)).
The changelog does not state a reason, and I did not find the corresponding CCP
notice `[GAP]`.

Practical upshot for a hunter `[PRAC]`: keep `Force Field` (occupancy) and keep
`Control Tower` only on the POS-finder tab; a dedicated "POS d-scan" preset is
redundant with a decent occupancy tab.

### 3.4 Sleepers: the d-scan blind spot that shapes hunting

`[MECH]` D-scan sees NPCs only when they carry **♦** in the name, and the examples
given are mining ops, Blood Raider shipyard fleets and AEGIS/SCC facility ships
([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
Sleepers are ordinary Entity NPCs and do not carry ♦.

`[ANEC]` This is corroborated by long-standing player reports: "You do not see
NPCs on d-scan. You can see their wrecks however, which can be a good indication
there is someone else in there with you" (2015, [EVE forums archive](https://forums-archive.eveonline.com/topic/418141)).
A 2025 thread contains a broad claim that d-scan "can pick up NPCs too", but the
same thread immediately narrows it: "This only shows certain NPCs FYI.
Trigs/EDENCOM show up but stuff like belt rats don't"
([2025 forum thread](https://forums.eveonline.com/t/combat-scanner/480896)). The
EVE Uni ♦ rule and the 2015/2025 narrowings agree; treat "all NPCs show on
d-scan" as wrong.

Design consequence: **you cannot use Sleeper activity to find an active site.**
You use the *wrecks* (on d-scan) and the overview once on grid. This is a real
limitation, not a preference: the object class a hunter would most like to see
off-grid is invisible.

### 3.5 What "breaks" a hunter's overview

1. **A missing type group is unrecoverable per-tab.** States cannot rescue a
   group that is not whitelisted (§2.1). New ship/entity groups are added by CCP
   regularly and land in new groups; the E-Uni overview page keeps a dated list of
   "objects missing if your overview is older than…" (Hybrid Drones, Command
   Carriers, Yachts, Odysseus, Lancer Dreadnoughts…) — "[Overview](https://wiki.eveuniversity.org/Overview).
   This is the single highest-cost failure mode.
2. **Colortag order collisions.** Order is first-match-wins. EVE Uni's worked
   example: "if you have Pilot has a bounty on him set higher than Pilot is at war
   with your corporation/alliance, you will only see the black skull bounty
   colortag rather than the red star war target colortag"
   ([Overview settings](https://wiki.eveuniversity.org/Overview_settings)).
   In J-space the usual collisions are standings vs sec-status vs war/militia.
3. **Militia states stealing priority.** EVE Uni has a whole section on the
   friendly-vs-militia collision and requires friendly designations be ranked
   above "at war with your militia" `([Installing the EVE University Overview](https://wiki.eveuniversity.org/Installing_the_EVE_University_Overview))`.
   In J-space there is no militia, so the state is pure downside: it can outrank a
   friendly tag and produce the exact misclick the pack authors worry about.
4. **Backgrounds and blink as noise.** "Generally, backgrounds are quite
   distracting, many players keep them unchecked except for outlaws and war
   targets… In a situation where you want to be able to see neutrals or hostiles
   very clearly, such as in nullsec or wormhole space, it can be handy to keep all
   backgrounds checked." Same source. In a WH, everything is a potential hostile,
   so a full background palette means everything is highlighted, which is the same
   as nothing being highlighted `[PRAC]`.
5. **Tab-count and modal confusion.** Overview settings tabs (Filters /
   Appearance / Columns / Ships / Misc / History) are not overview tabs; the
   in-game Tab Name field shows ~17 characters and ship-label fields ~3
   ([Overview manipulation](https://wiki.eveuniversity.org/Overview_manipulation)).
   Editing 20 tabs in-client is painful; YAML editing is the practical route —
   "almost all of the things that can be changed is much easier and more reliable
   to change in-game, with the exception of changing colors, font sizes, styles
   etc." Same source.
6. **Ship labels only restyle player-owned entities.** "You can only change the
   appearance of players and player items (like mobile depots) this way, celestials,
   anomalies, NPC ships etc will have their default appearance no matter what you
   do" (same source). A hunter cannot colour-code Ratting Site Key or Sleeper
   spawns via `shipLabels`.

---

## 4. What Z-S actually ships

### 4.1 Provenance, maintenance, and where it lives now

- `[MECH]` **Origin.** "What started as a stop-gap measure back in early 2016 to
  bring the **SaraShawa Overview** up-to-date for 118.6, snowballed into a complete
  re-write of the .YAML file and total overhaul of the visuals in July 2016 for the
  first installment." Credits: **Zirio** (YAML coding, pack maintenance, long-term
  updates) and **Deuce Syundai** (design & text formatting). Licence GNU GPLv3.
  ([README](https://github.com/Arziel1992/Z-S-Overview-Pack)).
- `[MECH]` **Lineage.** Z-S is a fork/continuation of **SaraShawa**, which itself
  continued **Sarah Schneider's** "Sarah's Overview Pack" — "The Overview Pack that
  started it all, **Sarah Schneider** pioneered the Overview customization". Same
  source.
- `[MECH]` **Two homes, both stale.** GitHub `Arziel1992/Z-S-Overview-Pack` and
  GitLab `Arziel/Z-S-Overview-Pack` are both last at tag **v9.00.0347 (April 2019)**;
  GitHub `master` HEAD is `9c7dd45` (2019-08-27) and contains the commit message
  "**Migration to GitLab**" (2018-06-22) — i.e. GitLab is the nominated successor
  but did not receive a v10.
- `[MECH]` **Where v10 lives.** The Customiser repo explicitly distinguishes the
  bundled bases: "**Z-S Full v10.06.09** (the default, **an in-game-only release**),
  **Z-S Full v9.00.0347** (the last git-tracked pack, kept for compatibility)".
  ([Customiser README](https://github.com/Arziel1992/Z-S-Overview-Customizer)).
- `[MECH]` **Who maintains it now.** "**Kismeteer**, the volunteer who currently
  keeps the Z-S Overview pack alive" — encountered by Zirio in June 2026. Zirio
  then rebuilt the customiser as a "Z-S 10th-anniversary gift to the EVE
  community", citing his first Z-S commit on 2016-06-09. Same source. (EVE Uni's
  overview page independently attributes *Kisover* to "Kismeteer in 2010",
  [Overview](https://wiki.eveuniversity.org/Overview) — same person, different pack.)
- `[MECH]` **EVE Uni's index** lists "Z-S Overview (GitLab) (In Game version Up to
  Date September 2026 – Hybrid drones added)… Based off of SaraShawa Overview Pack.
  Available in-game in 'Z-S Overview' Channel and 'Z-S Overview 1' Mailing List"
  ([Overview](https://wiki.eveuniversity.org/Overview)). Note the mailing-list name
  conflicts with the pack README's "**Z-S Overview**" and "**Z-S Overview 2**"
  `[GAP]`.

**Answer to "is the GitHub repo the live pack?" — no.** `github.com/Arziel1992/Z-S-Overview-Pack`
is the last git-tracked release line (v9.00.0347, April 2019). The current pack is
an in-game-only release (v10.06.09 as of the 2026 customiser) distributed through
the in-game channel/mailing list, maintained by Kismeteer with Zirio's tooling.

### 4.2 Preset inventory

`[MECH]` The README's own pack/preset inventory (v9 line) — verbatim structure,
abridged:

- **Z-S Core** (the only *required* pack): brackets `Combat/Targets/Friendly (±Dro)`,
  `Incursion`, `Misc`, `Travel`; `➲ Extra: Align Points`; `✈ --- Travel: All`;
  `✈ Travel (+Ship -Sta)`; `✪ = Friendly: Fleet`; `※ --- D-Scan: All`.
- **Z-S D-Scan Extended**: `※ D-Scan: Hunter (+Dro) (Offensive)`,
  `※ D-Scan: Hunter (-Dro) (Offensive)`, `※ D-Scan: Ships (Defensive)`.
- **Z-S PvX 1**: `PvX: Basic (±Neut ±NPC)` ×4, `✜ -- Ships only`,
  `✜ = CovOps: Collidables`, `✜ = Deployables`, `✜ = NPSI`, `✜ = Structures`,
  `✜ = Wartargets`.
- **Z-S PvX 2**: `Anoms`, `Exploring`, `FW`, `Incursion`, `Mining - All`,
  `Mining - Moon`, `Mission Filter (+LCO +LCS)`, `PI`, `Ratting`.
- **Z-S Targets Caps / Subcaps**: 20 target presets (Fighters … Titans; Antibombers …
  Pods).
- **Z-S Friendly Extended**: 10 presets (Caps, Carriers, Command Dest, Dictors,
  Dreads, Flag, Logistics, Recons, SC/T).
- **All 4 layouts** (Standard/Compact × 1BL/2BL), each carrying 8 presets:
  brackets Combat/Targets/Friendly (+Dro), brackets Warpout (!), `➲ Extra: Podsaver (!)`,
  `✈ Travel (-Ship +Sta)`, `PvX: Basic (+Neut +NPC)`, `✜ PvX: Salvage/Loot`,
  `✥ Target: Drones`, `✪ --- Friendly: All`.

`[MECH]` **Derived** from the two full styled bases bundled in the customiser
(`public/defaults/zs_full_v9.00.0347.yaml` and `zs_full_v10.06.09.yaml`):

- v9 full base: **68** presets. v10 full base: **58** presets. The ten absent from
  v10 are exactly the `✪ Friendly: *` extended set (Caps/Carriers/Command Dest/
  Dictors/Dreads/Flag/Logistics/Recons/SC-T/All Caps) — i.e. those now ship as the
  separate optional `Z-S Friendly Extended` pack. No v10 preset is new relative to
  v9 by name.
- The customiser README's Compare section independently states "between Z-S v9 and
  v10 only 6 of 68 names match byte for byte, while 58 match once the colours are
  ignored" — my 68/58 counts agree with theirs.

**No wormhole-named preset exists in Z-S.** Searching the v10 base for
`Wormhole`/`Sleeper`/`Cyno`/`Wreck`/`Depot` in preset *names* returns nothing
(derived). Z-S's WH-relevant capability is spread across `D-Scan: *`,
`CovOps: Collidables`, `PvX: Exploring`, `PvX: Salvage/Loot`, `Structures`,
`Deployables`, and the layout tabs. By contrast EVE University's own overview
ships a preset literally named **"6 - Wormhole"**, referenced by its scouting
guide ("Ping d-scan, 360°, max range, using Wormhole Overview Preset"; "Load 6 -
Wormhole") — [Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting).

### 4.3 Default tab layout

`[MECH]` **Derived** from `tabSetup` in `zs_full_v10.06.09.yaml` (indices 0–6, i.e.
seven configured tabs of a possible 20):

| # | Tab name | Overview preset | Bracket preset | Tab colour |
| --- | --- | --- | --- | --- |
| 0 | `<b> ✈ </b>` | `✈ Travel (-Ship +Sta)` | Brackets: Combat/Targets/Friendly (+Dro) | none |
| 1 | `<b> Main </b>` | `✜ --- PvX: Basic (+Neut +NPC)` | Combat/Targets/Friendly (+Dro) | red |
| 2 | `<b> 3 </b>` | `✥ Target: Drones` | Combat/Targets/Friendly (+Dro) | yellow |
| 3 | `<b> 4 </b>` | `✜ PvX: Salvage/Loot` | Combat/Targets/Friendly (+Dro) | violet |
| 4 | `<b> 5 </b>` | `✪ --- Friendly: All` | Combat/Targets/Friendly (+Dro) | blue |
| 5 | `<b> Exit! </b>` | `➲ Extra: Podsaver (!)` | Brackets: Warpout (!) | green |
| 6 | `Add more ->` | `✥ Target: All Caps` | Combat/Targets/Friendly (+Dro) | red |

Observations `[PRAC]`: Z-S ships **"Add more ->" as a placeholder tab** rather than a
finished set, and the README says "For best results, add additional tabs you need
and call them something you remember". EVE Uni's own install instructions make the
same point: after importing a pack, "you will still need to select what overview
presets are used"
([Installing the EVE University Overview](https://wiki.eveuniversity.org/Installing_the_EVE_University_Overview)).

Only tab 5 (`Exit!`) and the bracket slots deviate from the house bracket preset;
**Z-S does not ship a wormhole/POS-finder tab, a probes tab, or a collidables tab
in the default layout** — you are expected to build them (derived).

### 4.4 The D-Scan presets in detail (the most WH-relevant thing Z-S ships)

`[MECH]` **Derived** from `zs_full_v10.06.09.yaml`. All four have
`filteredStates: []` and `alwaysShownStates: []`.

| Preset | Groups | Contains | Deliberately excludes |
| --- | --- | --- | --- |
| `※ --- D-Scan: All` | 137 | Everything including `Wormhole`, `Sleeper Entities`, `Planet`, `Moon`, `Force Field`, all structures/deployables, wreck, biomass | — |
| `※ D-Scan: Hunter (+Dro) (Offensive)` | 103 | All player ship groups + `Capsule` + drones (`Combat/Mining/Energy Neutralizer/EWAR/Logistic/Stasis Webifying/Salvage Drone`) + probes + `Wreck` + `Biomass` + `Force Field` + all deployables/mobiles + `Citadel`/`Engineering Complex`/`Refinery`/`Upwell Cyno Beacon`/`Upwell Cyno Jammer` + `Cynosural Fields` + `Covert Cynosural Field` | **`Wormhole`, `Planet`, `Moon`, `Sun`, `Asteroid Belt`, `Sleeper Entities`, all NPC groups** |
| `※ D-Scan: Hunter (-Dro) (Offensive)` | 96 | As above minus drones | as above |
| `※ D-Scan: Ships (Defensive)` | 80 | Ships + capsule + probes + cynos + Upwell structures | wrecks, biomass, drones, celestials, NPCs |

The design intent is legible: **"Hunter" = objects that imply a player is present
or has recently been present, with every celestial and NPC removed.** That is the
answer to the clutter problem the 2014 dev-blog commenters complained about ("As a
guy who lived in wormspace most of his EVE-life I'm waiting for this feature…
remove planets and wrecks from d-scan when doing anomalies in wormspace to reduce
clutter and focus on detecting incoming ships and probes"
— [Hyperion dev blog thread](https://forums-archive.eveonline.com/message/4934236/)).

`[MECH]` **v9 → v10 drift in the D-Scan presets** (derived diff): the Hunter
presets gained `Cynosural Fields`, `Covert Cynosural Field`, `Upwell Cyno
Beacon/Jammer`, `Mobile Cynosural Beacon`, `Mobile Observatory`,
`Mobile Analysis Beacon`, `Interdiction Burst Probes`, `Filament Trace`,
`Phased Asteroid Field`, `Mercenary Den`, `Local Beacon`, plus new hulls
(`Lancer Dreadnought`, `Command Carrier`, `Expedition Command Ship`,
`Special Edition Yachts`). Nothing was removed. That the presets *grow* and never
shrink is the pack's mitigation for the §3.5 failure mode 1.

---

## 5. Cloaky specifics

### 5.1 What cloaking does to the two intel surfaces `[MECH]`

- A cloaked ship is "completely invisible to almost all forms of detection"; other
  players "can not see the ship at all" ([Cloaking](https://wiki.eveuniversity.org/Cloaking)).
- D-scan "excludes combat recon and cloaked ships"
  ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
- **But the cloak animation is visible on the overview.** "Once the cloaking module
  is activated the ship is immediately considered to be cloaked. **The cloaking
  animation itself will take a few seconds to play, during which time the cloaking
  ship is visible in space and on the overview, and can be selected, but cannot be
  targeted.** Once the cloaking module is deactivated, the ship will immediately
  appear on the overview and be targetable, even though the decloaking animation
  also takes a few seconds to play." ([Cloaking](https://wiki.eveuniversity.org/Cloaking)).
  This is the mechanic behind the two-way race a WH hunter operates in: you can
  catch someone mid-cloak, and you can be caught mid-cloak.
- A cloaking ship still has to break jump cloak first: "If you are already cloaked
  you cannot cloak… this will force you to break the gate cloak before you can
  activate your cloaking module, and in the process **briefly reveal yourself**."
  Same source. EVE Uni's scouting guide treats this as the one guaranteed
  exposure: "the only time your ship has appeared on D-scan is the brief time when
  swapping from jump cloak to cloaking device"
  ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)).
- Jump/gate cloak gives **60 seconds** of invisibility and invulnerability, and is
  explicitly presented as a *reading* window: "It is often a good idea to take full
  advantage of the 60 seconds of protection to look at the other ships on grid,
  player names in Local chat, or objects visible on directional scan"
  ([Cloaking](https://wiki.eveuniversity.org/Cloaking)) — i.e. precisely the
  overview + d-scan pair.
- **2000 m decloak rule.** "Being within 2000 m from any object will prevent ships
  from cloaking and cause cloaked ships to decloak. This includes often ignored
  objects such as jettisoned cans, corpses, collidable structures, drones, and gas
  clouds." Same source. This is the entire reason `✜ = CovOps: Collidables` exists
  in Z-S (§4.2) — and note the disruption *field* does not decloak you, the object
  does: "The disruption field itself will not de-cloak you, being within 2km of any
  object will" ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)).
- Cloak stabilisation / Mobile Observatory decloak pulses are **not a J-space
  concern**: observatories cannot be deployed in highsec or wormhole space, and
  "In highsec and wormhole space, this timer is effectively irrelevant and can be
  ignored" ([Cloaking](https://wiki.eveuniversity.org/Cloaking)).

### 5.2 What a hunter needs mid-hunt that the overview must not obstruct `[PRAC]`

1. **A one-key, always-available d-scan.** Rebind "Refresh Directional Scan" — off
   by default, set it in ESC → Shortcuts → Combat — and hold the key while clicking
   targets in space to aim the cone ("Simply hold the action shortcut key down and
   click the object… you can also right-click on a bracket and choose 'look at'").
   In a wormhole the wiki's own guidance is brutal about cadence: "If you are in a
   wormhole, you probably want to be spamming the scan every few seconds or so …
   every 8-10 seconds is often recommended… you may only have seconds to respond"
   ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
   Practical consequence for the overview design: the tab you are *in* must
   contain the warp-to targets and the hostile ships simultaneously, because you
   will be aiming the beam by clicking overview rows.
2. **Bookmarking without leaving the overview.** Wormhole bookmarks are made *from
   the overview* on grid, not from probe results: "you should bookmark the wormhole
   itself from your overview while on grid", because "Bookmarks created at a
   signature location from your scan probe results are often more than 5km from the
   Wormhole's actual location" ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)).
   So a hunter's travel tab must show wormholes as rows you can right-click → Save
   Location ([Bookmarks](https://wiki.eveuniversity.org/Bookmarks)). Bookmarks
   themselves are never overview/d-scan objects; they live in the Locations window.
3. **Probe launch, without leaving the overview readable.** Launching probes is the
   single most exposing act: "You and your probes will be visible while launching,
   so you want a nice spot out of d-scan range of any active ships", and the guide
   recommends setting probe range to 16 AU as a visual radius for d-scan exposure,
   then moving probes out of range before scanning ([Wormhole scouting](https://wiki.eveuniversity.org/Wormhole_Scouting)).
   Nothing about the overview helps here; the requirement it imposes is that the
   overview not be the thing you are staring at while probes are out.
4. **A probe-inclusive filter.** "you should always use a d-scan filter that include
   probes… While 'core' scanner probes cannot locate ships, their presence may
   indicate a cloaked ship" ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning));
   Z-S's Hunter presets include `Scanner Probe` and `Interdiction Probe` (derived).
   This is the only reliable way a *cloaked hunter* is detected, because the ship
   itself is not.
5. **Speed/range telemetry on the target for the tackle maths** — `VELOCITY`,
   `RADIALVELOCITY`, `TRANSVERSALVELOCITY`, `ANGULARVELOCITY` are per-tab columns
   in current clients ("You now select what columns are visible on a per-tab basis",
   [Patch Notes 21.02](https://www.eveonline.com/news/view/patch-notes-version-21-02);
   YAML `tabColumns` / `tabColumnOrder` per the Customiser). A minimal hunter tab
   should carry at least `ICON, NAME, TYPE, DISTANCE, VELOCITY` (Z-S's profile-wide
   columns are set from `overviewColumns` with `columnOrder`).
6. **A pods-out tab.** `➲ Extra: Podsaver (!)` in Z-S is 9 celestial groups, no
   player entities — the tab you switch to when the fight has gone wrong and you
   need a warp target fast, and (per EVE Uni) the tab used to flush stale ship-label
   formatting after a YAML edit.

### 5.3 The overview ↔ d-scan filter ↔ cloak interaction, stated as mechanics

- **Cloak does not change what the overview *would* show; it changes what exists to
  be shown.** Nothing in the profile is cloak-aware. There is no state, column, or
  preset variable for "cloaked". `[MECH]` by schema enumeration (the Customiser's
  complete root-key list contains no such key).
- **The d-scan filter and the overview tab are independent selections.** The
  d-scan filter selector offers "all" and "active overview", or any saved preset;
  "With the filter set to 'active overview' only objects permitted in the overview
  will be shown" ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
  Practical trap: switching overview tabs silently changes the d-scan result set
  when the d-scan is on "active overview", and the state settings you carefully
  configured on that tab are ignored (§2.3).
- **Cloaking is not a UI state either.** You can d-scan freely while cloaked; you
  cannot activate modules (other than the first 5 seconds), cannot target, and
  cannot warp without a covert-ops cloak ([Cloaking](https://wiki.eveuniversity.org/Cloaking)).
  So the overview is *read-only intel* while cloaked — which is exactly what a
  cloaky hunter uses it for, and why its copy/paste and name columns matter more
  than its action menu.

---

## 6. What is **not** expressible in an overview profile

The hard boundary. An overview profile is a **client-side display configuration**,
account-wide, with no execution surface.

`[MECH]` **Scope and storage.** "All Overview settings are saved account wide, and
a change will affect all characters on the same account"
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)).
Profiles are exported as YAML into `Documents/EVE/Overview` and imported through
Misc → Import (same source; also [Overview settings](https://wiki.eveuniversity.org/Overview_settings)).
Writes are manual and in-client.

`[MECH]` **The complete expressive surface** is the client's YAML schema:

| Root key | What it can express |
| --- | --- |
| `flagStates`, `flagOrder` | which states may draw a colortag, and their priority |
| `backgroundStates`, `backgroundOrder` | which states may highlight a row, and priority |
| `stateBlinks`, `stateColorsNameList` | flash on/off per state; colour per state |
| `overviewColumns`, `columnOrder`, per-tab `tabColumns`/`tabColumnOrder` | which of the 13 telemetry columns show, and their order |
| `presets` | per-preset `groups` whitelist + `filteredStates` + `alwaysShownStates` |
| `tabSetup` | tab index (0–19) → `name`, `overview` preset, `bracket` preset, `color` |
| `shipLabelOrder`, `shipLabels` | order and styling of bracket-label segments |
| `userSettings` | miscellaneous client-side flags |

Source: [Customiser README, "YAML config key legend"](https://github.com/Arziel1992/Z-S-Overview-Customizer).
That table is the whole vocabulary. Anything not in it is not expressible.

Therefore, definitively **not** possible:

1. **Cannot alert.** There is no sound, notification, or popup hook in the schema.
   `stateBlinks` is a per-frame visual flash and nothing else. (CCP's "Warning
   Sounds" are tied to Crimewatch events, not to the overview — [Patch Notes
   21.02](https://www.eveonline.com/news/view/patch-notes-version-21-02).)
2. **Cannot log or record.** No log output, no history file, no telemetry. The only
   outputs are a YAML export and an in-game share link; the History tab stores only
   the last 15 *shared profile links you clicked*, not events
   ([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)).
   A hunter who wants a record of "who was in this system at 03:14" must use an
   external tool or write it down.
3. **Cannot see beyond the client's own detection rules.** Two hard walls:
   - **Grid.** Off-grid objects are not on the overview at all, by construction
     ([Grid](https://wiki.eveuniversity.org/Grid)).
   - **Cloak.** Cloaked ships and combat recons do not exist on either surface
     ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
     Scan-immune hulls are invisible until they decloak.
4. **Cannot filter by distance or bearing.** The schema has no range/angle
   predicate. Distance is a *column* (drawable) and an in-client sort, never a
   filter. Range/angle filtering exists only on the **d-scan** (1 / 5 / 10 / 14.3 AU;
   5°–360°) ([Directional scanning](https://wiki.eveuniversity.org/Directional_scanning)).
   A hunter cannot say "hide everything past 50 km".
5. **Cannot see cosmic signatures, cosmic anomalies, or bookmarks.** Anomalies and
   signatures live in the Probe Scanner window; bookmarks live in the Locations
   window ([CCP Scanning](https://support.eveonline.com/hc/en-us/articles/203209902-Scanning),
   [Bookmarks](https://wiki.eveuniversity.org/Bookmarks)). Both are explicitly
   invisible to d-scan as well. For a WH hunter this is the biggest structural gap:
   the wormhole you just probed down is only clickable via the probe window or,
   once identified and in space, as an object you can bookmark from the overview.
6. **Cannot invent states.** The state list is fixed and integer-keyed (§2.1). You
   cannot define "pilot who killed me before", "known local", or "prober" as a
   state, though you can approximate reputation bookkeeping with standings
   (15/16/18/19/48) because standings drive states.
7. **Cannot automate or script.** No triggers, no conditions, no macros, no API.
   The file is inert data. CCP's own design position is the mirror image of this —
   structure vulnerability info is deliberately kept out of the API "since we want
   people to actively scout, or infiltrate spies… instead of relying on external
   tools to do the job for them"
   ([Citadels, sieges and you](https://www.eveonline.com/news/view/citadels-sieges-and-you)).
8. **Cannot restyle non-player objects.** "You can only change the appearance of
   players and player items (like mobile depots) this way, celestials, anomalies,
   NPC ships etc will have their default appearance no matter what you do"
   ([Overview manipulation](https://wiki.eveuniversity.org/Overview_manipulation)).
9. **Cannot exceed 20 tabs** ([CCP](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)),
   and cannot sidestep the fixed column menu (13 tokens: `ICON, NAME, TYPE, TAG,
   DISTANCE, CORPORATION, ALLIANCE, FACTION, MILITIA, SIZE, VELOCITY,
   RADIALVELOCITY, TRANSVERSALVELOCITY, ANGULARVELOCITY`).
10. **Cannot be shared selectively.** An in-game share link "will only include the
    tab presets that were actually in use in the overview" of the sharer — a
    long-standing limitation CCP flagged at Hyperion
    ([Hyperion dev blog](https://www.eveonline.com/news/view/sharing-is-caring)).
    YAML export/import is the only complete route.

---

## 7. Dead ends and unresolved questions

| # | Question | Status |
| --- | --- | --- |
| 1 | Can the client whitelist a single **type** (e.g. only "Rifter") rather than a group? | `[GAP]` The YAML stores bare integer IDs in `groups`; groupIDs and typeIDs share the same number space, so the file cannot disambiguate. In Z-S v10's presets, 648 of 660 distinct IDs resolve as groupIDs in the Customiser's SDE matrix and 12 do not resolve at all (`549, 1023, 1071, 1073, 1165, 1211, 1297, 1405, 1407, 1409, 1410, 4748`) — derived. The Customiser asserts "Overview filters operate almost exclusively at the group level", while EVE Uni describes the in-game Types tree as letting you select "specific things". Unresolved; do not build on type-level filtering. |
| 2 | How many overview tabs? | **Resolved in favour of 20** — CCP support, updated 2026-06-01. EVE Uni's "eight" is stale; the 2014 Hyperion dev blog said five/15. |
| 3 | Sensor-recalibration delay after decloaking | Sources disagree: EVE Uni Cloaking says "up to 6 seconds" for covert-ops; EVE Uni Wormhole scouting's ship table says a Stealth Bomber has "No sensor recalibration delay". Probably different hulls/timings, but unresolved. |
| 4 | The exact on-grid distance at which a no-access Upwell structure appears on the overview | `[GAP]` Players report ~2,000 km and ~10,000 km in different threads; CCP's 2015 dev blog only says "if the user can interact with the structure". |
| 5 | Whether POSes were removed from the game | `[GAP]` No CCP removal patch note found. Legacy `Starbase` groups (incl. `Control Tower` = 365) remain in the current SDE. Z-S's "No more POS's!" v9.00.0346 entry gives no reason. |
| 6 | Whether a *current* (2026) first-party statement exists about overview-truncated structure visibility | `[GAP]` Relying on the 2015 Citadel dev blog plus EVE Uni plus forum corroboration. No current patch note found. |
| 7 | Z-S mailing-list names | Conflict: pack README says `Z-S Overview` / `Z-S Overview 2`; E-Uni wiki says `Z-S Overview` / `Z-S Overview 1`. |
| 8 | Whether Z-S ships a wormhole-specific preset in the v10 in-game-only release | `[GAP]` The v10.06.09 base bundled in the customiser has no WH-named preset, but it is one snapshot of an in-game-only release; the pack's in-game channel may hold more (or fewer) presets than the bundled base. |
| 9 | Dead end: `wiki.eveuniversity.org/Wormhole_PvP` and `.../Local` | Both return 404/crawl-not-found. The wormhole PvP material that exists is `Wormhole_scouting` (a step-by-step scouting/hunting guide) and `Wormhole_space`. |
| 10 | Dead end: CCP support article on the directional scanner specifically | Not found; the `Scanning` support article covers probe scanning only. The description of d-scan therefore rests on EVE Uni plus patch-note fragments (e.g. "Light, Support and Heavy Fighters now show up on directional scan" and "Force Auxiliary ships will now appear on directional-scan", both [Patch notes for EVE Online: Citadel](https://www.eveonline.com/news/view/patch-notes-for-eve-online-citadel)). |
| 11 | Lower-trust practice claims used in this note | D-scan cadence ("every 3–4 seconds" in W-space, [2012 forum](https://forums-archive.eveonline.com/topic/61918/); "8-10 seconds" per EVE Uni); "add a custom filter in overview settings that includes all player ships, bubbles and probes" for a defensive d-scan ([r/Eve 2025](https://www.reddit.com/r/Eve/comments/1lwr6o3/guide_for_better_overview_and_dscan_filters/)); the general hunter framing "The Second you see a saber on Dscan, Warp away" ([2021 forum](https://forums.eveonline.com/t/how-to-run-relic-and-data-sites-in-wormholes/339456)). Cited as practice, not mechanic. |

## 8. One-paragraph answer

For a mostly-solo, cloaky, mostly-wormhole hunter, the overview and the d-scan
are two non-overlapping sensors rather than two views of one thing: the overview
is a **live, grid-scoped, identity-bearing** list (pilot names, standings flags,
exact range and velocity, plus more than 100 type groups you can whitelist), while
the d-scan is a **manual, 14.3 AU, direction-only, identity-blind** sweep that is
the *only* source for probes, wrecks, force fields, bubbles, mobile depots, and
structures you have no docking access to — and which silently cannot see Sleepers
or cloaked ships at all. The killer mechanic is that a preset is a **type
whitelist first and a state filter second**, and d-scan filters ignore states
entirely, so a hunter needs at least: a clean combat tab (player ships in, NPCs
and celestials out, friendlies state-filtered), a "Hunter" d-scan preset (presence
objects only — which is exactly what Z-S's `D-Scan: Hunter (±Dro)` do, with 103 and
96 groups and empty state arrays), a warp-to/align tab (with a near-empty podsaver
variant), a loot/salvage tab (wrecks and containers, using the wreck-only states
36/37), and a collidables tab driven by the 2000 m decloak rule. Z-S Overview ships
all of that machinery but no wormhole-specific preset: it is a GPLv3 fork of
SaraShawa (itself the continuation of Sarah's pack), with the last git-tracked
release being v9.00.0347 in April 2019 — the current pack is an in-game-only
release (v10.06.09 as bundled in the 2026 Z-S Customiser), kept alive by Kismeteer
with tooling rebuilt by Zirio. And finally, know the wall: the profile cannot
alert, cannot log, cannot automate, cannot see off-grid or through a cloak, cannot
filter by range, and cannot see signatures, anomalies or bookmarks — it is inert
account-wide display data and nothing more.

---

## 9. Account-wide vs per-character granularity

Appended 2026-10-12. This section answers the narrow question the profile scope
raises for a hunter who shares one account across four characters: **how far does
"account wide" actually reach, and where does per-character state survive?** It
does not re-open the profile's expressive limits (§6) or the pack inventory (§4);
it pins the *scope boundary*.

**§9 sources pinned in play**

| Source | Version / date |
| --- | --- |
| `support.eveonline.com/.../203273831-Overview-Settings` | CCP, updated 2026-06-01 ("account wide" statement; 20-tab cap) |
| `support.eveonline.com/.../5885278408860-Profile-Manager` | CCP, updated 2023-12-11 (launcher profiles; settings-folder layout) |
| `support.eveonline.com/.../11058434966428-EVE-Online-launcher` | CCP (launch groups) |
| `support.eveonline.com/.../8413329735580-EVE-Online-End-User-License-Agreement` | CCP, updated 2024-07-30 (multi-account rule) |
| `eveonline.com/eve-academy/alt-accounts` | CCP first-party (switch vs simultaneous) |
| `eveonline.com/news/view/new-default-overview-in-testing` | CCP dev blog, 2022-11-18 (overview stored locally, per account) |
| `eveonline.com/news/view/improving-photon-ui` | CCP dev blog, 2022-11-24 (columns per tab; multiple overviews) |
| `github.com/Arziel1992/Z-S-Overview-Customizer` `public/defaults/zs_full_v10.06.09.yaml` | schema enumeration of root keys (same pin as §1) |
| `github.com/eve-wrench/eve-wrench-app` `crates/eve-wrench-core/src/{copy.rs,scan.rs}` | `core_user`→account / `core_char`→character section lists; `prefs.ini` bracket key |
| `developers.eveonline.com/docs/community/eve-wrench/` | CCP-hosted community page for the decoder above |
| `github.com/iridiumops/overview/blob/main/help.md` | player-made pack FAQ ("per account and shared between all 3 characters") |
| `wiki.eveuniversity.org/Alternate_characters` | updated 2024-01-25 (one character per account) |
| `wiki.eveuniversity.org/Overview_settings` | still says eight tabs — stale, see §2.2 and §7 |
| Forum: `overview-settings-for-multiple-characters` (494561), 2025-07-30 | `[ANEC]` launcher-profile practice, what breaks |
| Forum: `ui-main-and-secondary-overviews` (483980), 2025-04-22 | `[ANEC]` d-scan "active overview" resolves to last active tab |
| Forum: `multiple-overview-problems-happening` (388004), 2022-12 | `[ANEC]` d-scan/bracket window-priority bugs |
| Forum: `photon-ui-feature-preview-multiple-overviews-feedback-thread` (385591), 2022-11/12 | `[ANEC]` separated overviews and active-tab scope |
| Forum: `request-for-per-character-overview-settings` (225487), 2020-03-13 | `[ANEC]` the "I want per-character" request |
| Reddit `r/Eve` `1rnr485` (2026-03) and `11er225` (2023) | `[ANEC]` bracket-visibility persistence |

`[MECH]` **The account-wide statement, verbatim.** "All Overview settings are
saved account wide, and a change will affect all characters on the same account."
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings),
updated 2026-06-01). CCP's 2022 default-overview dev blog sharpens *where* it
lives: "The overview settings are stored on the machine locally and are
associated with a specific account on that local EVE client instance only."
([New default overview in testing](https://www.eveonline.com/news/view/new-default-overview-in-testing)).
So the scope is **one record per account, per local client instance** — not per
character, and also not synced across machines.

### 9.1 What the exported profile carries — and it is all account-wide

An overview profile *is* the exported YAML. `[MECH]` **Derived** by enumerating
the top-level keys of `public/defaults/zs_full_v10.06.09.yaml` (the Z-S v10 base
pinned in §1), the document contains exactly these thirteen root keys — no more,
and none of them is character-keyed:

| Root key | What it configures | Per-character? |
| --- | --- | --- |
| `presets` | each named preset's `groups` (type whitelist) + `filteredStates` + `alwaysShownStates` (58 presets in the v10 base) | **No** — one shared preset list |
| `tabSetup` | tab slots 0–19 → `name`, `overview` preset, `bracket` preset, `color` | **No** — one shared tab set |
| `overviewColumns`, `columnOrder` | the profile-wide column menu and order (9 columns in the v10 base) | **No** |
| `flagStates`, `flagOrder` | which states may draw a colortag, and their priority | **No** |
| `backgroundStates`, `backgroundOrder` | which states highlight a row, and priority | **No** |
| `stateBlinks` | per-state on/off blink (`flag_*`, `background_*`) | **No** |
| `stateColorsNameList` | per-state colour name | **No** |
| `shipLabels`, `shipLabelOrder` | bracket/label pre/post formatting and order | **No** |
| `userSettings` | misc client-side flags (empty `[]` in the v10 base) | **No** |

CCP confirms the shape of the shared parts: tabs are configured "through the
overview settings, and up to 20 Tabs can be configured to be shown within the
overview"; a tab "can be configured in the saved Tab Filters … and by selecting a
saved Tab Filter within the 'Bracket Filter'"; and an export "includes all
currently configured Tabs"
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings)).

`[MECH]` **Columns are per *tab*, not per *character*.** CCP's Photon UI dev blog:
"Column visibility, order, and width are now configured and persisted per tab,
instead of globally before"
([Improving Photon UI](https://www.eveonline.com/news/view/improving-photon-ui)).
That adds a third axis (tab) inside the one account-wide document; the v10 base
predates/omits per-tab column keys and carries only the profile-wide
`overviewColumns`/`columnOrder` (**derived**). Either way it never becomes
per-character.

**Answer to §9.1: nothing in the exported profile is per-character.** By schema
enumeration there is no per-character field to find; the whole document is one
account-scoped record. Everything the brief lists — presets, tab bindings/names/
colours, per-tab columns, colortag and background config, `stateBlinks`,
`stateColorsNameList`, `shipLabels`/`shipLabelOrder`, `userSettings`, profile-wide
`overviewColumns`/`columnOrder` — is shared with every character on the account.

### 9.2 What is *not* in the profile, and where its scope actually lands

The items below are absent from the YAML by enumeration, so they are *not* part
of the account-wide profile. Where each lands is the whole question; some are
per-character, some are per-client-profile, and two are unresolved.

`[MECH]` **The client splits its two settings files.** CCP documents the profile
folder (`settings_xyz` under `%LOCALAPPDATA%\CCP\EVE\`) but not the internal filenames
([Profile Manager](https://support.eveonline.com/hc/en-us/articles/5885278408860-Profile-Manager);
[Clearing client settings](https://support.eveonline.com/hc/en-us/articles/6918765073436-Clearing-client-settings-and-the-client-side-cache-data)).
Community tooling and CCP's own older forum posts name them: `core_user_<accountid>.dat`
= account settings, `core_char_<charid>.dat` = character settings. The `blue.Marshal`
decoder EVE Wrench is catalogued on CCP's developer portal
([EVE Wrench](https://developers.eveonline.com/docs/community/eve-wrench/)), and its
source assigns settings groups to one file or the other
(`crates/eve-wrench-core/src/copy.rs`,
[source](https://github.com/eve-wrench/eve-wrench-app/blob/main/crates/eve-wrench-core/src/copy.rs)):

- **`core_user_*` (account)** — `overview`, `defaultoverview`, `probescanning.`
  (probe formations), `suppress`, `audio`, `tabgroups`, camera/graphics, market,
  module-slot order.
- **`core_char_*` (character)** — `windows` (positions, sizes, open/stacked/
  minimised state), `neocomButtonRawData`, `chatchannels`, `InfoPanelModes_`,
  `dockPanels`.

Note this is decoded *client* data verified by a third-party library, not a CCP
support statement — treat it as `[MECH]`/verifiable-locally rather than
CCP-authored.

| Item | In the profile YAML? | Where it lives | Scope | Tag |
| --- | --- | --- | --- | --- |
| **Currently selected overview tab** | No (no key) | Client UI session state | Tab→preset *binding* is account-wide; the *selection* is remembered client state, and a 2022 feedback request explicitly asks for "which tab was active … stored per character" — i.e. it was not. Exact per-character persistence **not established** | `[GAP]` |
| **d-scan filter dropdown selection** | No | Client UI state; eligible list = the account-wide `presets` | Same shared preset menu for every character; which preset is *selected* is client memory. EVE Wrench's README places "dscan" in `core_char_*` (character), but no d-scan copy group exists in its source, so the key is unenumerated | `[PRAC]` (per-character) / `[GAP]` (exact) |
| **d-scan range/angle** | No | Client state, character-scoped per the same `core_char` note | Likely per character; not exported, not re-synced by the profile | `[PRAC]` / `[GAP]` |
| **Bracket display mode** ("show all" / "stop showing all" / "hide all") | No | `bracketsAlwaysShowShipText=` in **`prefs.ini`** | **Per client profile**, not per character and not in the account overview record — `prefs.ini` sits beside `core_user_*`/`core_char_*` and is read by whichever profile the client launched with (`scan.rs`, [source](https://github.com/eve-wrench/eve-wrench-app/blob/main/crates/eve-wrench-core/src/scan.rs)). Players report it resetting on restart (2023), and the 2026 answer "you will have to select it on each tab" suggests it is also tab-sensitive | `[MECH]` (third-party decode) / `[ANEC]` (persistence) |
| **Window positions, sizes, open/minimised state** | No | `windows` section of `core_char_*` | **Per character** | `[MECH]` |
| **Which windows are stacked/tabbed together** | No | `tabgroups` section of `core_user_*` | **Account-wide** | `[MECH]` |
| **Column visibility/order/width** | Yes (per tab) | Inside the account-wide profile | Shared (per tab, not per character) | `[MECH]` |
| **Overview list sort order** | No | Not documented anywhere found | Unresolved | `[GAP]` |

`[ANEC]` **What is *not* in the profile and behaves account-wide anyway** is the
*separated-overview structure*. When multiple overviews arrived (13 Dec 2022), a
player put it precisely: "Right now moving a tab to a new separate overview window
on one character also creates a separate overview for other characters on the same
account … ideal[ly] which filter is loaded in which tab, and which tab was active
should also be stored per character"
([Multiple Overviews feedback thread](https://forums.eveonline.com/t/photon-ui-feature-preview-multiple-overviews-feedback-thread/385591)).
The same thread documents that window *positions* are already per character
("You seem to already store their position per character"), matching the
`windows`/`core_char` split above.

**The consequence the brief is really after:** one shared profile does **not**
force all four characters onto the same d-scan *selection* — the selection is
client state outside the profile. But it does force them onto the same d-scan
*menu*: the preset list is account-wide, so a `D-Scan: Hunter` preset your cloaky
hunter edits or adds is in the other three characters' dropdowns too, and there is
no per-character way to hide a preset. And if the d-scan is set to **"active
overview"**, it follows the last active tab, whose preset binding *is*
account-wide; to get genuinely per-character d-scan behaviour every character must
select a specific d-scan preset by name rather than "active overview"
([2025 report](https://forums.eveonline.com/t/ui-main-and-secondary-overviews/483980)).

### 9.3 Can two characters from one account be logged in at once? (2026 rule)

**No — one character per account at a time.** `[MECH]` The EULA restricts play at
the *account* level: "You are however not allowed to play EVE by using more than
one Account at the same time, unless you pay a subscription fee for each of the
Accounts you intend to use for that purpose"
([EULA](https://support.eveonline.com/hc/en-us/articles/8413329735580-EVE-Online-End-User-License-Agreement),
updated 2024-07-30). `[MECH]` CCP's own EVE Academy says the only way to be online
with two characters at once is two accounts, both Omega, and that on a single
account "you also have to log out and log back in to select which character to
use" ([Alt Accounts](https://www.eveonline.com/eve-academy/alt-accounts)).
`[MECH]` EVE University states it plainly: "only one character per account can be
logged in at a time
([Alternate characters](https://wiki.eveuniversity.org/Alternate_characters)).

**Why this matters for profile swapping:** you cannot be simultaneously logged in
with two characters of the *same* account, so re-importing a profile "for the
hunter while the hauler is online" is not a race you can lose — the hauler cannot
be online. The real hazard is different: a **launcher profile is chosen per
account**, and if two *different* accounts launch under the same profile they
share one settings folder. A 2024 launcher report records exactly that failure:
when launching through a group, "all accounts in the group end up using the same
settings instead of their individual profiles"
([EVE Launcher forum](https://forums.eveonline.com/t/launching-multiple-accounts-launcher-forgets-some-profiles/439373)).
So per-character profile swapping is safe *within* one account and unsafe *across*
accounts that reuse the same profile.

### 9.4 Is profile-swapping a real, used workflow?

**Yes — it is the CCP-sanctioned workaround, and it is used.**

- `[MECH]` **CCP provides it.** "Settings are stored in profiles… The profile
  manager allows users to create custom profiles for each account or character…
  After selecting an account and going into Account Settings in the launcher, you
  can select which profile should be loaded with the account or create a new one."
  ([Profile Manager](https://support.eveonline.com/hc/en-us/articles/5885278408860-Profile-Manager)).
  Each profile is a separate `settings_xyz` folder holding its own `core_user_*`
  and `core_char_*` files. `[MECH]` The launcher also offers **launch groups**,
  whose documented purpose is starting a set of characters together
  ([EVE Online launcher](https://support.eveonline.com/hc/en-us/articles/11058434966428-EVE-Online-launcher)).
- `[ANEC]` **The community reaches for profiles first.** In the 2025 thread where
  one account's highsec and nullsec characters kept overwriting each other, the
  recommended fix is "switch between launcher profiles when you log in either your
  NS or your HS character… each launcher profile will have it's own folder with
  it's own separate isolated settings files", with the caveat that swapping
  profiles requires quitting the client and relaunching, and that launch groups
  let you start a whole role-set with one click
  ([Overview settings for multiple characters](https://forums.eveonline.com/t/overview-settings-for-multiple-characters/494561)).
- `[ANEC]` **Re-importing exported YAMLs is the lighter-weight variant.** The
  Iridium pack's FAQ tells players the facts and offers only the blunt answer —
  "No, overview settings is per account and shared between all 3 characters"
  ([help.md](https://github.com/iridiumops/overview/blob/main/help.md)) — while a
  2020 feature request thread suggests keeping "all your overview settings linked
  in the MOTD" of a private chat channel so a character can re-import quickly
  ([Request for Per Character Overview Settings](https://forums.eveonline.com/t/request-for-per-character-overview-settings/225487)).
  A Minmatar Fleet guide repeats the share-link approach
  ([Alesis Overview Guide](https://wiki.minmatar.org/guides/Alesis-Overview-Guide)).
- `[PRAC]` **Tools automate the file copy, not the YAML swap.** A long line of
  third-party settings managers copies `core_user_*`/`core_char_*` between
  characters and accounts: EANM, EANM's successor `mintnick/eve-settings-manager`,
  `eve-wrench/eve-wrench-app`, `PapaChop83/EVEProfileSync`,
  `kshannoninnes/CopyEveLayoutTool`, `folklorelabs/eve-cloner`,
  `walaydenobsidia/eve_profile_copier`, `jpbriend/eve-settings-manager`. Their
  READMEs uniformly warn that the client must be closed first and that account
  settings only copy to account files and character settings only to character
  files.
- `[ANEC]` **What breaks, in practice:** the account-wide bleed itself — "I save a
  Null Sec setting and High Sec alts Overview breaks"; separated overviews
  appearing on characters that never created them; d-scan attaching to the wrong
  overview window after undock/login
  ([2025 thread](https://forums.eveonline.com/t/overview-settings-for-multiple-characters/494561);
  [2022 thread](https://forums.eveonline.com/t/multiple-overview-problems-happening/388004));
  and launcher groups collapsing several accounts onto one profile (2024, above).
  None of these are caused by the profile format — they are the direct cost of the
  account-wide scope.

### 9.5 A shared 20-tab budget: can a character hide tabs it does not need?

`[MECH]` The client caps the overview at 20 tabs
([CCP Overview Settings](https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings);
EVE Uni's "up to eight different tabs" is stale — see §2.2 and §7). Because the
tab set lives entirely in `tabSetup` inside the account-wide profile (§9.1),
**every configured tab is visible to every character on the account.** There is
no per-character tab-visibility key in the YAML by schema enumeration, and none is
documented by CCP.

`[ANEC]` Can a character at least *hide* a tab from its own UI? The available
lever is the separated-overview window, and it does not free the budget: one
player describing multi-overview on the same account writes that moving a tab out
"also creates a separate overview for other characters on the same account … I
have to hide separated overviews (and move their filters around) if I do not want
them taking up the extra screen space on a specific character, **but that leaves me
with less tabs to work with. I cant remove it because then it's gone from the
other character where I do want it separated**"
([Multiple Overviews feedback thread](https://forums.eveonline.com/t/photon-ui-feature-preview-multiple-overviews-feedback-thread/385591)).
So: not per-character visibility, and hiding the window on one character still
spends the shared tab. The only genuine separation is a **separate launcher
profile** (§9.4) — which forces a client restart to swap and reintroduces the
cross-account bleed if another account shares the profile.

**Practical reading for four characters sharing 20 tabs:** design the profile as a
union both roles can live with, not as four role-specific layouts. The cloaky
hunter's Wormhole/POS-finder/Collidables tabs will be present on the three pack
defaults characters, and the pack's PvE/Mining tabs will be present on the hunter.
Per-tab column overrides and window *positions* can differ per character; the tab
*n *set*, names, colours, bindings, presets, colortags and backgrounds cannot.

### 9.6 One-paragraph answer

The account-wide statement is exact and literal, but its boundary is a *file*
boundary, not a "everything in the client" boundary. Everything in the exported
overview YAML — all presets, all 20 tab slots and their name/preset/bracket/colour
bindings, per-tab column layout, colortag and background priority and blink,
`stateColorsNameList`, ship labels, `userSettings` — is one record per account and
is identical for all four characters [MECH]. Outside that record the client keeps
character-scoped state in `core_char_*.dat` (window positions and sizes, and per
EVE Wrench the d-scan) and profile-scoped state in `prefs.ini` (the "show all
brackets" toggle), while the *selected* overview tab and the *selected* d-scan
filter are unexported UI state whose per-character persistence is not established
[GAP]. The practical bite is not that the profile forces one d-scan *filter* on
everyone — it does not; each character can pick its own preset or fall back to
"active overview" — it is that the preset *menu* is shared, so no character can
have a private preset or a private tab, and the 20-tab budget is a single account
budget. When roles genuinely conflict, the only CCP-sanctioned separation is a
separate launcher profile per account (safe for same-account characters, since
only one of them can be online at a time anyway, but requiring a client restart
and dangerous across accounts that reuse the profile); re-importing exported YAMLs
or pasting share links from a chat-channel MOTD is the fast manual alternative,
and the third-party file-copiers automate only the character/account `.dat` copy,
not the YAML swap.
