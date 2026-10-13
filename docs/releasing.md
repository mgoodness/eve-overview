# Releasing

Releases are cut by [release-please](https://github.com/googleapis/release-please),
not by hand. It reads Conventional Commits on `main`, keeps a running release
PR, and — when that PR merges — creates the `vX.Y.Z` tag and the GitHub Release.

Two workflows do the work:

| Workflow | Trigger | Job |
| --- | --- | --- |
| `release-please.yml` | push to `main` | opens/updates the release PR; on merge, tags and releases |
| `release.yml` | a release is published | generates the profile, attaches `overview.yaml` + `SHA256SUMS`, appends attribution and pins to the notes |

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

## Immutable releases

If the repository ever enables immutable releases, switch
`release-please-config.json` to hold the release as a draft
(`"draft": true, "force-tag-creation": true`) and have `release.yml` attach the
assets and publish it — GitHub blocks adding assets to an already-published
immutable release, and a draft defers tag creation unless force-tag-creation is
set.
