# The overview-pack landscape: Iridium and Signal Cartel, against our Z-S base

This repository generates a wormhole-hunter overview profile from the **Z-S
Overview v10.06.09** library. Two other actively maintained packs are worth
knowing: **Iridium Overview** (source-built, permissively licensed) and **Signal
Cartel's "Explorer's Overview"** (exploration- and wormhole-focused). This note
synthesises both from primary sources and records what they confirm, where they
disagree with our design, and what they suggest for this project.

It complements [`overview-for-wormhole-hunting.md`](overview-for-wormhole-hunting.md),
which covers the game mechanics. Trust tags follow that note: `[MECH]` game
mechanic, `[PRAC]` established practice, `[ANEC]` anecdote, `[GAP]` unresolved.

## Iridium Overview

**Provenance.** `[MECH]` Repository <https://github.com/iridiumops/overview>.
README: *"Simple yet powerful EVE Online overview pack for everyone… up to 13
preconfigured tabs, 38 overview filters, 3 bracket filters, 3 dscan filters."*
Credited to *Thomas Iridium* and *Feer Truelight*; the changelog runs to
`v3.12.0 (2026/09/22)`.

**Licence.** `[MECH]` `license.md`: *"SPDX-License-Identifier: MIT OR
BSD-3-Clause"*, *"Copyright © 2013-2026, IridiumOps"*. Permissive, unlike Z-S's
GPLv3.

**Architecture.** `[MECH]` Modular rather than monolithic: `parts/*.yaml` are
semantic sub-parts, composed by `src/build.py` into `releases/*.yaml`.
`workflow.md` describes the loop: monitor ESI for new group IDs → edit the parts →
build → import-test in the client → tag and release. This is the same
source → compile → release shape as this repository's `source/` → `Generate`.

**Comments are stripped.** `[PRAC]` (and see the `[GAP]` below) `parts/readme.md`:
*"During the build process, part files get updated with data from SDE in form of
comment labels, however **all YAML comments are stripped from final YAML file as
they are not supported by the game client**."* Comments are used in the *source*
for human legibility and removed from the shipped artefact.

**Filter naming.** `[PRAC]` Filters carry a category symbol as their first
character; bracket filters are prefixed `/` and d-scan filters `?`, *"selected for
its low ASCII numeric value in order to allow their selection via shortcuts in the
Dscan window, which sorts filters lexically in ascending order."* `+`/`-` in
parentheses mark variations of a base filter.

**Upstream monitoring.** `[PRAC]` `src/monitor.py` diffs ESI group IDs against the
last release, mapping each ID to its name (`src/api.py`: `api_base_url =
"https://esi.evetech.net"`). It documents its own blind spots: *"eve uses single
shared id range for all types whether they can show up on the overview or not"*,
and *"new group ids that belong to new categories previously not represented in the
overview will be missed by this script."*

**Install.** `[MECH]` `readme.md` / `workflow.md`: place the YAML in
`%USERPROFILE%/Documents/EVE/Overview`, then Overview Settings → **Misc** → **Import
Settings** → select the file → **Check All** → **Import**, then *"log out and log
back in (recommended)"*. In-game shareable links are limited to 16 filters each, so
a pack of this size needs several links.

**Design stances.** `[PRAC]` `help.md`: the overview is per account; beacons cannot
be filtered by site (*"all beacons have the same group ID"*) nor Upwell structures
by size (*"grouped by their function not their size"*); columns advice (Icon,
Distance and Type essential; angular vs transversal velocity for tracking).

## Signal Cartel "Explorer's Overview"

**Provenance.** `[MECH]`
<https://wiki.signalcartel.space/Public:Signal_Cartel_Overview_Pack>: *"tailored
towards exploration… a lightweight, user-friendly, no-nonsense alternative."*
Created by *Bernard Lachapelle* in 2021; maintained by *Echerie Saissore* "as of
December 2025".

**Licence.** `[MECH]` The page states no licence for the **pack artefact**. The
wiki itself (`Public:Copyright`) is *"available under the Creative Commons
Attribution-ShareAlike License unless otherwise noted"*; EVE materials are Fenris
hf.'s. Unlike Iridium and Z-S, there is no explicit content licence to build on.

**Distribution.** `[MECH]` In-game channels only — `Explorer's Overview` and
`EvE-Scout` — installed by clicking shareable links; the page describes no YAML
download. Versioning is an incremental build number ("TYPE"): the changelog runs
TYPE 21 (2026-02-28) through TYPE 26 (2026-08-21), and each update asks for a
reset-and-reinstall.

**Tabs and presets.** `[MECH]` Four tabs — **TRAVEL, EXPLORATION, COMBAT,
EVERYTHING ELSE** — with roughly eleven presets. The wormhole-relevant ones are
`>>DANGER CLOSE`, *"meant to be used as a dscan filter while out in wormholes,
contains ships and everything related to interdiction"*, and `>EXPLORATION + POS`.
A `>COLLIDABLES` preset existed but was **removed in v1.7**.

**Design stances.** `[PRAC]` *"Unless overruled by a specific preset (e.g:
'>FRIENDS' or '>FOES'), all players will appear on all tabs, ensuring you can spot
potentially hostile players from any tab you might be on."* `[PRAC]` v1.5: *"Disabled
blinking background for war targets."*

## What this confirms, contradicts, and suggests

1. **YAML comments may not be safe in the shipped profile `[GAP]`.** Iridium strips
   every comment, stating they are *"not supported by the game client"*. This
   repository's emitted profile opens with a `#` comment block carrying attribution
   and the reproducibility pins. No first-party statement was found either way, and
   EVE University's YAML-editing page is ambiguous. **Unresolved: whether the client
   rejects a comment-bearing file or merely ignores the comments.** If it rejects
   them, the artefact is unimportable and the attribution must move to a sidecar.

2. **Source → compile → release is established practice `[PRAC]`.** Iridium's
   `parts/` → `build.py` → `releases/` mirrors this repository's `source/` →
   `Generate`. Related prior art: `rossos/eve-pog` and Evansenko's generator do the
   same. The approach is not novel, which is reassuring.

3. **Upstream type monitoring has a documented blind spot `[PRAC]`.** Iridium diffs
   **ESI group IDs** against its last release; this repository's `upstream` command
   watches the **SDE build** and the **Z-S pack** instead. Neither catches a new
   Group that appears in an existing category under a new SDE build — Iridium's own
   caveat — which is the spec's highest-cost failure mode. A group-ID diff is the
   missing complement to the pack/build check.

4. **Wormhole d-scan practice, independently confirmed `[PRAC]`.** Iridium's
   `? DSCAN: Extra` is *"for reconnaissance in wormholes, hunting specific targets
   (… secondary evidence such as wreck accumulation can be observed)"*, and Signal
   Cartel's `>>DANGER CLOSE` is a d-scan filter *"while out in wormholes… everything
   related to interdiction."* Two mature packs back the practice this repository's
   research note already records.

5. **Group-level-only filtering, confirmed `[PRAC]`.** Iridium's FAQ independently
   states the two limits our spec builds on: beacons share one group ID, and Upwell
   structures are grouped by function, not size. Type-level filtering is not
   expressible.

6. **Where the packs disagree with our design — recorded as prior art `[PRAC]`.**
   - Signal Cartel **disabled war-target background blinking** (v1.5) and **removed
     its collidables preset** (v1.7). This profile keeps a Blink on the war target
     (see [ADR-0003](../adr/0003-militia-blink-reduction.md)) and keeps a collidables
     preset in the library (unbound, its members folded into the working tabs).
   - Signal Cartel's default is the inverse of ours for friendlies: *all players on
     all tabs unless a preset overrides*, versus this profile's friendly-vetoed
     `Engage` plus a dedicated `Fleet` tab.

7. **Two install caveats worth adopting `[PRAC]`.** Iridium recommends **logging out
   and back in** after import (this repository's docs only said dock/undock), and
   warns that **filters sharing a name are overwritten** — so an in-game edit to an
   imported preset does not survive a re-import (the spec's "regenerate, don't
   re-edit" assumption, stated from the client side). Both are now in
   [`docs/install.md`](../install.md).

8. **The base library is a live design question, not a settled one.** Z-S is stale
   (the machine-readable copy is v10.06.09, in-game-only; the public repository stops
   at v9.00.0347, 2019) and **GPLv3**, so the emitted profile carries GPLv3. Iridium
   is actively maintained and **MIT OR BSD-3-Clause**, which would make the emitted
   profile permissive. Switching bases is an [ADR-0001](../adr/0001-account-wide-union-profile.md)
   / [ADR-0002](../adr/0002-ten-tab-roster.md)-level decision — the roster and four
   authored presets would be re-derived — not a pin bump.

## Sources

Primary and near-primary pages on overview and UI settings, grouped. Pages cited
inline above are omitted; this is the wider reading list, including the
layout material a "how to fly it" guide would draw on.

**First-party (CCP / Fenris)**

- *Overview Settings* — <https://support.eveonline.com/hc/en-us/articles/203273831-Overview-Settings>. The canonical mechanics: 20 tabs, filters vs bracket filters, the Exception/State hierarchy, colortag/background priority, columns, ship labels, import/export, account-wide scope.
- *User Interface – Tips and Tricks* — <https://support.eveonline.com/hc/en-us/articles/7264100568220-User-Interface-Tips-and-Tricks>. Window layout: compact mode, margin size, locking a window's size and position, and moving an overview tab into its own window.
- *Scanning* — <https://support.eveonline.com/hc/en-us/articles/203209902-Scanning>. Probe scanner and Sensor Overlay.
- *Improving Photon UI* — <https://www.eveonline.com/news/view/improving-photon-ui>. Where multiple overview windows, per-tab columns and the presets→filters rename landed.
- *February Update – Photon UI Full Rollout* — <https://www.eveonline.com/news/view/february-update-photon-ui-full-rollout>. Multiple overviews standard, tab cap raised, filter grouping, global compact mode.
- *The future of probe and directional scanning UI* — <https://www.eveonline.com/news/view/the-future-of-probe-and-directional-scanning-ui>. D-scan window states, Show Scan Cone, Align with Camera.

**EVE University**

- *Overview* — <https://wiki.eveuniversity.org/Overview>; *Overview settings* — <https://wiki.eveuniversity.org/Overview_settings>; *Overview manipulation* — <https://wiki.eveuniversity.org/Overview_manipulation>. The how-to triad; *manipulation* is the page that would help settle the comment-header question (#12).
- *Directional Scanner Guide* — <https://wiki.eveuniversity.org/Directional_Scanner_Guide>. Range/angle, the filter menu, "active overview", and the statement that d-scan ignores state settings.
- *Cloaking* — <https://wiki.eveuniversity.org/Cloaking>; *Wormhole scouting* — <https://wiki.eveuniversity.org/Wormhole_Scouting>. The 2,000 m decloak rule and J-space scan discipline.
- *Varius Arcturus' "New Overview"* — <https://wiki.eveuniversity.org/User:Paul_knuth/New_Overview> (PDF: <https://eve-files.com/media/1010/varius_arcturus_overview_guide_1.8.pdf>). A step-by-step walkthrough with a Wormhole preset and a screen-layout section.

**Layout and UI (community)**

- *Kismeteer's EVE UI guide* — <https://www.wckg.net/Newbie/eve-ui>. Practical layout and d-scan setup, by the current Z-S maintainer.
- *EVE Online 2026 Guide — UI & HUD* — <https://hardlygospel.github.io/eve-online-2026-guide/skills-navigation/>. A concrete recommended window arrangement.
- *[Photon UI] Multiple Overviews feedback* — <https://forums.eveonline.com/t/photon-ui-feature-preview-multiple-overviews-feedback-thread/385591>; *Photon UI Setup for all Accounts* — <https://forums.eveonline.com/t/photon-ui-setup-for-all-accounts/395746>. Multi-window behaviour and the layout-copy pain.

**Prior art — source→compile generators**

- *rossos/eve-pog* — <https://github.com/rossos/eve-pog>. Named groups → presets → tabs → compiled EVE-compatible YAML; the closest architectural twin to this repository's generator.
- *Evansenko's Overview Generator* — <https://io.evansosenko.com/eve-overview/>. Another parts → generate → import project.

**Notable discrepancy.** The all-brackets toggle hotkey is documented inconsistently: EVE University's *Overview* page says `Alt/Option+Z`, while the 2026 community guide says `Alt+Shift+B`. Verify in-client rather than trusting either.
