# Releasing

Releases are cut by [release-please](https://github.com/googleapis/release-please),
not by hand. It reads Conventional Commits on `main`, keeps a running release
PR, and — when that PR merges — creates the `vX.Y.Z` tag and the GitHub Release.

Two workflows do the work:

| Workflow | Trigger | Job |
| --- | --- | --- |
| `release-please.yml` | push to `main` | opens/updates the release PR; on merge, creates the `vX.Y.Z` tag and a **draft** release |
| `release.yml` | the `vX.Y.Z` tag is pushed | generates the profile, attaches `overview.yaml` + `SHA256SUMS` to the draft, appends attribution and pins to the notes, then publishes |

So cutting a release is: **merge the release-please PR.** Nothing is tagged by
hand, and no tag is pushed from a laptop.

## Why a GitHub App

release-please must push the release PR and the tag as the shared
`mgoodness-release-please` App, not as `GITHUB_TOKEN`: a push or PR authored by
`GITHUB_TOKEN` cannot trigger another workflow, so the tag would be created and
`release.yml` would never fire — silently. The App token is minted per run by
`actions/create-github-app-token` and scoped to `contents: write` and
`pull-requests: write`.

`release.yml` itself uses the default `GITHUB_TOKEN`; it only attaches assets to
a release that already exists.

## One-time setup

The App must be installed on this repository and its credentials stored as the
repository variable `RELEASE_PLEASE_APP_CLIENT_ID` and the repository secret
`RELEASE_PLEASE_APP_PRIVATE_KEY`. That is a human step (it installs an App and
writes a secret). The thin wrapper calls the shared dotfiles tool:

```sh
scripts/setup-release-please-app.sh \
  --key-file op://Private/release-please/mgoodness-release-please.2026-09-01.private-key.pem
```

The tool opens the App's install page, then sets both values. It is idempotent.

## Tag shape

`release-please-config.json` leaves `package-name` unset and sets
`include-component-in-tag: false`, so tags are `v1.2.3`, not `eve-overview-v1.2.3`.
A component in the tag would still work, but the shorter form is what the
release workflow and `docs/install.md` assume.

## Draft-first and immutable releases

`release-please-config.json` sets `"draft": true` and
`"force-tag-creation": true`. release-please creates the tag and a draft
release; the tag (App-pushed, so the trigger fires) starts `release.yml`, which
attaches every asset and only then publishes the release.

Publishing last is required because **immutable releases are enabled**
(`gh api -X PUT repos/mgoodness/eve-overview/immutable-releases`). Once
published, a release's tag is locked to its commit and its assets can no longer
be added or changed, and GitHub mints a release attestation for it. Attaching
assets after publish — the default release-please flow — would be rejected.

`force-tag-creation` matters as much as `draft`: without it GitHub defers tag
creation for a draft release until it is published, so the tag-triggered
workflow would never fire, and the draft would sit unpublished forever.
