#!/usr/bin/env sh
#
# Wire the shared release-please GitHub App into this repository: install it,
# then set the RELEASE_PLEASE_APP_CLIENT_ID variable and the
# RELEASE_PLEASE_APP_PRIVATE_KEY secret that .github/workflows/release-please.yml
# reads.
#
# See the global AGENTS.md "Release automation" section. The App's Client ID is
# read from ~/.config/release-please/apps.conf, so only the private key has to
# be supplied — and as an op:// reference it never has to be pasted by hand:
#
#   scripts/setup-release-please-app.sh \
#     --key-file op://Private/release-please/mgoodness-release-please.2026-09-01.private-key.pem
#
# This tool touches GitHub App installation and repository secrets, so it is
# run by a human, not in CI.
set -eu
exec setup-release-please-app "$@"
