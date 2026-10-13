# eve-overview

A generated EVE Online **overview profile** for a mostly-solo, cloaky,
mostly-wormhole PvP hunter — and the generator that produces it.

The profile is a **ten-Tab union** over Z-S Overview v10.06.09's full 58-preset
library, built **additively** (no Z-S preset is pruned), plus four presets
authored for wormhole hunting. It is one account-wide profile serving three
characters: the hunter, and two alts flying Planetary Industry, exploration and
PvE.

The definition lives in this repository as a readable, name-keyed source tree;
one command resolves Group names to SDE identifiers and emits a single
importable `.yaml`. A change is reviewable in a diff instead of buried in a
thousand generated lines.

## Install

Download `overview.yaml` from the
[latest release](https://github.com/mgoodness/eve-overview/releases/latest) and
import it in the client.

> **Nothing must be installed first.** The profile embeds the entire Z-S preset
> library (all 58 presets, and the bracket presets they reference), so the Z-S
> Overview pack is **not** a prerequisite. Full steps:
> [docs/install.md](docs/install.md).

## What you get

| | |
| --- | --- |
| Tabs | 10 — Travel, Engage, Exit!, Holes, Fleet, PvE, PI, Exploring, Loot, Occupancy |
| Presets | 62 — Z-S's 58 untouched, plus `Engage`, `Holes`, `Occupancy`, `✜ PvX: Exploring (+Hazards)` |
| Appearance | Backgrounds only on relationship/threat States; blinks on war target, criminal and low-sec; one-line ship labels |
| Scope | Account-wide — EVE stores one overview profile per account |

The reasoning is recorded in [`docs/adr/`](docs/adr/) and the cited research in
[`docs/research/`](docs/research/).

## Generate it yourself

```sh
go run ./cmd/eve-overview generate -out overview.yaml
```

The generator fetches the pinned reference pack and SDE once, verifies both
against their SHA-256, caches them, and works offline thereafter. A Group name
that does not resolve fails generation **and names itself** — a missing Group is
invisible until a fight, so it can never ship silently.

## How it works

```text
source/                 the reviewable source of truth
├── groups.yaml           named Group sets — names resolve to SDE Group IDs
├── presets.yaml          the authored presets
├── tabs.yaml             the ten-Tab roster
├── appearance.yaml       colortags, backgrounds, blinks, ship labels
└── columns.yaml          profile-wide + per-tab columns
overview/               the single seam: Generate(source, opts) → profile bytes
internal/reference/     pinned fetch/cache for the Z-S pack + SDE
cmd/eve-overview/        the CLI — generate, upstream
```

## Upstream data and dependencies

Two reference artefacts are pinned by checksum and watched for drift:

| Input | Pin | Source |
| --- | --- | --- |
| EVE Static Data Export | build `3607705` | CCP / Fenris hf. |
| Z-S reference pack | `zs_full_v10.06.09.yaml` | the Z-S Customiser's bundled base |

A scheduled [`upstream` workflow](.github/workflows/upstream.yml) checks both
weekly with `go run ./cmd/eve-overview upstream` and opens an issue when either
moves. Bumping a pin is a deliberate regeneration, reviewed and released like
any other change — never an automatic upgrade.

Go modules and the workflows' actions are kept current by Dependabot
([`.github/dependabot.yml`](.github/dependabot.yml)); minor and patch bumps
auto-merge once the required `test` check passes, majors wait for a human.

## Releases

Releases are cut by
[release-please](https://github.com/googleapis/release-please) from Conventional
Commits. Merging its release PR creates the `vX.Y.Z` tag and a draft release; a
tag-triggered workflow attaches `overview.yaml` + `SHA256SUMS` and only then
publishes. Releases are immutable. See [docs/releasing.md](docs/releasing.md).

## Attribution and licence

The generated profile is a **derived artefact** of the Z-S Overview pack,
licensed **GNU GPLv3**, and is distributed under those same terms. Z-S Overview
is by **Zirio** (YAML, maintenance) and **Deuce Syundai** (design), a
continuation of the SaraShawa Overview. EVE Static Data Export © Fenris hf. See
[`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md); the emitted profile carries
the same attribution and the reproducibility pins in its header.

## Development

```sh
go test ./... -count=1     # -short skips the network-backed conformance run
go vet ./...
```

The conformance test fetches the real pinned Z-S pack and fails a fully-skipped
run, so a formatting mistake is caught here rather than in space.
