#!/usr/bin/env bash
# Old-reader / new-reader compatibility test for telemetry.v2 additions.
#
#   ./scripts/compat-test.sh                 # old = PRE_NATIVE, new = this tree
#   ./scripts/compat-test.sh --old-ref REF   # old = any git ref
#
# Generates the NEW Go SDK from this tree (buf.gen.yaml -> gen/go) and the OLD
# one from --old-ref (compat/buf.gen.old.yaml -> compat/internal/oldgen), both
# gitignored, then runs the compat module's tests: newreader against the new
# code, oldreader against the old code, each in its own test binary.
#
# Needs network for buf's remote plugin and Go's module cache, like
# `buf generate` itself (docs/local-codegen.md).
set -euo pipefail

# THE OLD READER IS PINNED, NOT `main`. It is the last schema before
# Envelope.native_batch existed (8be0628). Once native records merge, main
# knows field 5, so an old reader built from main is no longer old, and the
# test would go red on main for everyone. A pinned ref keeps the claim true
# forever: readers built before native records see them as unknown.
PRE_NATIVE=8be0628e3c8911f29c313f609eab7f3ec3f34119
OLD_REF="$PRE_NATIVE"
while [ $# -gt 0 ]; do
  case "$1" in
    --old-ref) OLD_REF="$2"; shift 2 ;;
    *) echo "compat-test: unknown argument: $1" >&2; exit 2 ;;
  esac
done

root="$(git rev-parse --show-toplevel)"
cd "$root"

GENMOD=gen/go/buf.build/gen/go/echotools/nevr-api/protocolbuffers/go
# gen/go is the dev-loop tree (docs/local-codegen.md) and is regenerated in
# place, not deleted. compat/internal/oldgen belongs to this script alone.
rm -rf -- compat/internal/oldgen
buf generate
cat > "$GENMOD/go.mod" <<'MOD'
module buf.build/gen/go/echotools/nevr-api/protocolbuffers/go

go 1.23

require google.golang.org/protobuf v1.36.11
MOD

buf generate ".git#ref=${OLD_REF}" --template compat/buf.gen.old.yaml
echo "compat-test: old = $(git rev-parse --short "${OLD_REF}"), new = working tree"

cd compat
go test -count=1 -v ./...
