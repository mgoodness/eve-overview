# Install

The generated profile is a single EVE Online overview `.yaml`. Getting it into
the client is **manual and in-client**: there is no installer, no script, and
no automation in this repository. The client loads a profile only when you tell
it to, in the Overview Settings window.

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
release's `SHA256SUMS`.

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
EVE actually runs from. If the folder does not exist, create it; it is the same
folder the client exports to.

## 3. Import it in-game

1. Open **Overview Settings** (the overview's hamburger menu → Overview
   Settings).
2. Open the tab settings menu — labelled **Misc** in the client's own wording —
   and choose **Import**.
3. Select `overview.yaml` (or paste its contents, if you imported from a
   clipboard).
4. The profile is applied account-wide. Importing replaces the account's
   current overview settings, so export a backup first if you want to keep them.

If the import does nothing, the file is usually in the wrong folder or was
saved with a `.txt` suffix; the client only reads `.yaml` from the folder above.

## Notes

- **Account-wide.** An overview profile is per account, not per character, so
  one import covers every character on the account.
- **Manual.** There is no install automation and none is planned: the client
  exposes no import-on-start hook, and this repository writes a file, not an
  installed program.
- **Verify before flying.** Z-S ships an online Customiser that previews a
  profile; previewing the generated file there is the one check this
  repository's tests cannot perform.
