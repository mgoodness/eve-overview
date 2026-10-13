# Install

The generated profile is a single EVE Online overview `.yaml`. Getting it into
the client is **manual and in-client**: there is no installer, no script, and
no automation in this repository.

> **Prerequisites: none.** The profile embeds the entire Z-S preset library —
> all 58 Z-S presets, and the bracket presets the tabs bind — so you do **not**
> need the Z-S Overview pack, or any other pack, installed first. Importing this
> one file is the whole installation.

- Generator and source of truth: <https://github.com/mgoodness/eve-overview>
- Third-party attribution and the licence the profile carries:
  [`THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md)

## 1. Get the profile

Download `overview.yaml` from a
[release](https://github.com/mgoodness/eve-overview/releases), or generate it
from source:

```sh
go run ./cmd/eve-overview generate -out overview.yaml
```

Either way the file's attribution header records the exact reference pack and
SDE build it was generated from. A release asset can be verified against the
release's `SHA256SUMS`:

```sh
shasum -a 256 -c SHA256SUMS
```

## 2. Put the file where the client reads profiles

EVE reads and writes overview YAML in an `EVE/Overview` folder inside the
account's Documents directory. The location depends on the platform:

| Platform | Folder |
| --- | --- |
| Windows | `%USERPROFILE%\Documents\EVE\Overview` |
| macOS | `~/Documents/EVE/Overview` |
| Linux (Steam/Proton) | the EVE prefix under Steam's `compatdata`, e.g. `~/.steam/steam/steamapps/compatdata/8500/pfx/drive_c/users/steamuser/Documents/EVE/Overview` |

The Linux path varies with launcher and prefix — Lutris, Heroic and custom
Proton prefixes each keep their own `compatdata`-style tree, so locate the one
EVE actually runs from.

If the folder does not exist, create it — or, better, **export your current
overview from the game first** (Overview Settings → **Misc** → *Export Overview
Settings*, *Check All*). That both backs up the settings this import is about to
replace and creates the folder for you.

## 3. Import it in-game

1. Open **Overview Settings** — the `≡` button at the top left of the Overview
   window.
2. Open the **Misc** tab and click **Import Overview Settings**.
3. Select `overview.yaml` in the left-hand list, click **Check All**, then
   **Import**.
4. Some elements may not refresh until the UI reloads — **dock and undock once**
   to force it.

The ten tabs arrive configured; use the overview's tab strip to select them. The
d-scan preset dropdown also lists all 62 presets, including the Z-S d-scan ones.

## What you get

- Ten tabs: **Travel, Engage, Exit!, Holes, Fleet, PvE, PI, Exploring, Loot,
  Occupancy** — the hunter's pressure set first.
- Z-S's full 58-preset library untouched, plus four authored presets: `Engage`,
  `Holes`, `Occupancy`, and `✜ PvX: Exploring (+Hazards)`.
- An appearance layer tuned so a highlighted row means something: backgrounds
  only on relationship and threat States, blinks only for war target, criminal
  and low-security status, and one-line bracket labels.

## Notes

- **Account-wide.** An overview profile is per account, not per character, so
  one import covers every character on the account. This profile is built as a
  union for exactly that reason.
- **Importing replaces the account's settings.** The import overwrites the
  account's presets, tabs, appearance and columns. Do the step-2 export first if
  you want to keep what you had.
- **Manual.** There is no install automation and none is planned: the client
  exposes no import-on-start hook, and this repository writes a file, not an
  installed program.
- **If the import does nothing**, the file is usually in the wrong folder or was
  saved with a `.txt` suffix; the client only reads `.yaml` from the folder
  above.
- **Verify before flying.** Z-S ships an online Customiser that previews a
  profile before it is imported; previewing the generated file there is the one
  check this repository's tests cannot perform.
