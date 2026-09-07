# CLAUDE.md

## Project

nevr-proto is a proto-only repository for the NEVR platform. Protocol Buffer definitions are distributed via the [Buf Schema Registry](https://buf.build/echotools/nevr-api). There is no generated code in this repository.

## Packages

| Package | Purpose |
|---------|---------|
| `engine/v1` | EchoVR engine HTTP API types (/session, /player_bones) |
| `gameservice/v1` | EchoVR ↔ game service protocol messages (packed-struct representations) |
| `telemetry/v1` | Session capture format v1 (frames, events, header) |
| `telemetry/v2` | Layered capture format v2 (game-agnostic envelope + game payloads) |
| `spatial/v1` | 3D primitives (Vec3, Quat, Pose) |
| `archive/` | Deprecated protos (excluded from buf module) |

## Commands

```bash
buf lint                                    # Lint proto files
buf build                                   # Compile check
buf format --diff --exit-code               # Check formatting
buf format -w                               # Auto-fix formatting

./scripts/install-hooks.sh                  # One-time: point git at .githooks/
./scripts/breaking-gate.sh --target registry            # compat vs the published module
./scripts/breaking-gate.sh --target '.git#branch=main'  # compat vs local main (PR shape)
```

**Run `./scripts/install-hooks.sh` once per clone.** `.git/hooks` is untracked, so
a hook that lives there is a hook nobody else has; the tracked hooks in
`.githooks/` do nothing until `core.hooksPath` points at them.

- `pre-commit` — `buf format --diff --exit-code` and `buf lint`. Prints the fix
  command; never rewrites your working tree.
- `pre-push` — blocks a push to `main` that breaks compatibility with the
  published module. It blocks *here* because a push to `main` publishes to a
  registry other repos fetch, and a follow-up commit cannot un-consume it. Being
  offline is not a violation: an unreachable registry warns loudly and lets the
  push through.

A deliberate break is declared on the commit that breaks it:

```bash
git commit --amend --trailer "Breaking-Approved: <why this is safe now>"
```

The trailer excuses only its own commit — not the rest of the push — and CI and
the hook both call `scripts/breaking-gate.sh`, so they cannot disagree about what
"approved" means.

## Conventions

- Package naming: `<domain>.<version>` (e.g., `telemetry.v2`)
- Version in directory path, never in filename (per AIP-191)
- Options order: `csharp_namespace`, `go_package`, `java_multiple_files`, `java_outer_classname`, `java_package`
- `csharp_namespace`: `Nevr.<Package>.<Version>` (e.g., `Nevr.Telemetry.V2`)
- `java_package`: `com.echotools.nevr.<package>.<version>`
- `go_package`: `buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/<package>/<version>`
- Events use `oneof` envelopes with field ranges by category (10-19, 20-29, etc.)
- `gameservice/v1` message names preserve reverse-engineered `SNS`/`STCP` prefixes from the game binary
