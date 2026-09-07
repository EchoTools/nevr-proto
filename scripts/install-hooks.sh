#!/usr/bin/env bash
# A HOOK NOTHING INSTALLS IS A HOOK NOBODY RUNS.
#
# .git/hooks is untracked, so a hook placed there exists only on the machine that
# made it. This repo tracks its hooks in .githooks/ and points git at them, which
# is the same convention as the other repos here.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
git config core.hooksPath .githooks
echo "core.hooksPath = $(git config core.hooksPath)"
echo "installed: $(ls .githooks | tr '\n' ' ')"
