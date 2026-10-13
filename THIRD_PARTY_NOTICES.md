# Third-party notices

The generated overview profile (`overview.yaml`, published as a release asset)
is a **derived artefact**. This file records what it derives from and the terms
those sources carry.

## Z-S Overview

The profile is built **alongside and on top of** the Z-S Overview pack: the
pack's 58 presets ride into the emitted profile untouched, and four presets are
authored or modified beside them.

- **Z-S Overview** — <https://github.com/Arziel1992/Z-S-Overview-Pack>
  - **Zirio** — YAML coding, pack maintenance, long-term updates
  - **Deuce Syundai** — design and text formatting
  - A continuation of the **SaraShawa Overview**, itself continuing
    **Sarah Schneider's** Sarah's Overview Pack.

### Licence

Z-S Overview is licensed **GNU GPLv3**. The emitted profile is a derivative
work and is distributed under the same terms. The pack's licence text is
`LICENSE.md` in the pack repository above.

## EVE Online Static Data Export

Group names in the profile's source are resolved to Group identifiers through
CCP's Static Data Export.

- **Static Data Export** © Fenris hf., used under the EVE Online
  third-party developer licence.

## Reproducibility pins

The profile is reproducible from this repository plus these two pinned inputs.
Both are fetched at generation time and verified against the SHA-256 below; a
mismatch aborts generation.

| Input | Pin |
| --- | --- |
| Z-S reference pack | `412541b10959c186dc7958a66fd4362484b72dabce0e3bf12728a57c1aa5f346` |
| Z-S reference pack URL | <https://raw.githubusercontent.com/Arziel1992/Z-S-Overview-Customizer/40a4bc3ed3bf619638f129e6b8a524867733d802/public/defaults/zs_full_v10.06.09.yaml> |
| SDE build 3607705 | `905a199f0114be2e251f98d17ff0adc36bbb38a94febc828276285bf6d2e6315` |
| SDE build 3607705 URL | <https://developers.eveonline.com/static-data/tranquility/eve-online-static-data-3607705-jsonl.zip> |

The same pins are recorded in the attribution header of every generated
profile, so the exact input bytes are recoverable from the artefact alone.
