<!--
Release notes template for eve-overview.

The release workflow copies this file to release-notes.md and appends the
generated SHA256SUMS, so the published release carries the profile's own
checksum as well as the attribution and the reproducibility pins below. Nothing
in this file is release-specific; the tag supplies the version.
-->

# eve-overview

The generated wormhole-hunter overview profile, published as an importable EVE
Online overview `.yaml`.

**Import steps:** [docs/install.md](docs/install.md) — manual and in-client;
there is no installer.

## Assets

- `overview.yaml` — the generated profile, with its attribution header.
- `SHA256SUMS` — its SHA-256, for verifying the download.

## Attribution and licence

**This profile is a derived artefact of the Z-S Overview pack, licensed
GNU GPLv3**, and is distributed under those same terms.

- **Z-S Overview** — <https://github.com/Arziel1992/Z-S-Overview-Pack>
  - **Zirio** — YAML coding, pack maintenance, long-term updates
  - **Deuce Syundai** — design and text formatting
  - A continuation of the **SaraShawa Overview**, itself continuing
    **Sarah Schneider's** Sarah's Overview Pack.

EVE Online Static Data Export (SDE) © Fenris hf., used under the EVE Online
third-party developer licence.

## Reproducibility

Regenerate this profile from the tagged source tree plus the two pinned inputs
below. Both are fetched at generation time and verified against their SHA-256,
so a released profile can only have been built from these exact bytes.

| Input | Pin |
| --- | --- |
| Z-S reference pack | `412541b10959c186dc7958a66fd4362484b72dabce0e3bf12728a57c1aa5f346` |
| Z-S reference pack URL | <https://raw.githubusercontent.com/Arziel1992/Z-S-Overview-Customizer/40a4bc3ed3bf619638f129e6b8a524867733d802/public/defaults/zs_full_v10.06.09.yaml> |
| SDE build 3607705 | `905a199f0114be2e251f98d17ff0adc36bbb38a94febc828276285bf6d2e6315` |
| SDE build 3607705 URL | <https://developers.eveonline.com/static-data/tranquility/eve-online-static-data-3607705-jsonl.zip> |

The profile's own SHA-256, taken over the released asset:

<!-- The workflow appends the SHA256SUMS block here. -->
