# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- `telemetry/v2` native engine records (`native.proto`): `Envelope.native_batch = 5` (`NativeBatch`, made of `NativeRecord` and `NativeSequenceDisposition`) and `CaptureHeader.native_time_unit_ns = 103`. Each record keeps its native type (`msg_type`), mailbox and exact payload bytes. It also carries a session-global `uint64` capture sequence (starting at 1, never wrapping), a `uint64` elapsed capture time in the header's unit (1000 = microseconds), its stage (`INPUT` or `EMITTED`), its connection index and epoch (optional, so connection 0 is distinct from no peer), and optional causal references. Dispositions account for lost or omitted sequence ranges without allocating new sequences. Additive: legacy millisecond fields keep their meaning
- `compat/` and `scripts/compat-test.sh`: an old-reader test. Hand-assembled wire vectors (`testdata/telemetry/v2/`) are decoded by Go code generated from the last pre-native schema (8be0628) and by the branch's own `buf generate`. The test shows that an old reader sees a native batch as an unknown field 5, preserved byte for byte, and never as a known envelope case
- `telemetry/v2` Echo Combat tapes: `CaptureHeader.echo_combat = 11` (`EchoCombatHeader`) and `Frame.echo_combat = 11` (`EchoCombatFrame`), in `echo_combat.proto`. Combat carries its own messages instead of the arena ones. The event messages `WeaponFired`, `DamageDealt`, `PlayerKilled`, `PayloadProgress` and `CombatRoundEnded` are wrapped in `CombatEvent`. Only `DamageDealt.shield_delta` has a field so far. The rest are empty until their wire layouts are read
- `telemetry/v2` `EchoCombatFrame` gains the combat /session fields the real files carry: `contested = 2`, `timers = 3` (`CombatTimers`), `payload = 4` (`CombatPayload`), and `players = 5` (`CombatPlayer`, with `playerid` and `arm`)
- `telemetry/v2` `VoiceChunk` (`voice.proto`) and `Envelope.voice = 4`: per-player voice, one speaker's consecutive Opus packets per chunk, with the speaker's slot and its generation, the first packet's sequence and a millisecond offset on the capture's time base
- `telemetry/v2` `PlayerInfo.generation = 7` and `PlayerJoined.generation = 7`: the slot's generation, so a `VoiceChunk` is matched to its player by (slot, generation) across a slot hand-over
- `telemetry/v2` `MatchSettingsChanged` event (`EchoEvent.match_settings_changed = 16`, `EVENT_TYPE_MATCH_SETTINGS_CHANGED = 7`): `rules_changed_by`, `rules_changed_at` and `total_round_count` change mid-capture, and the header alone held only their first value
- `telemetry/v2` layered capture format with game-agnostic envelope, `EchoArenaHeader`/`EchoArenaFrame` payloads, `CaptureFooter` seek indexes, and `EventType` enum
- `spatial/v1` package with `Vec3`, `Quat`, `Pose` primitives
- CI: `buf breaking` check on pull requests
- CI: `buf push` for BSR publishing on merge to main
- CI: `buf format` check

### Changed

- Renamed `apigame/v1` → `engine/v1` (domain: EchoVR engine HTTP API)
- Renamed `rtapi/v1` → `gameservice/v1` (domain: EchoVR ↔ game service protocol)
- Removed version suffixes from filenames per AIP-191 (`engine_http_v1.proto` → `engine_http.proto`, etc.)
- Normalized `csharp_namespace` across all packages to `Nevr.<Package>.<Version>` pattern
- Normalized `java_package` across all packages to `com.echotools.nevr.<package>.<version>` pattern
- Fixed `SNSLobbyFindSessionRequestv11Message` → `SNSLobbyFindSessionRequestV11Message` (casing consistency)

### Fixed

- Applied `buf format` canonical formatting
- Removed stale excludes from `buf.yaml`

## [1.0.0] - 2026-03-15

### Changed

- Restructured repository from mixed proto+generated-code monorepo to proto-only repository
- Moved proto definitions from `proto/` subdirectory to repository root
- Updated `go_package` options to BSR paths (`buf.build/gen/go/echotools/nevr-api/...`)
- Moved `buf.yaml` from `proto/buf.yaml` to repository root
- Replaced `buf-generate.yml` CI workflow with `buf-ci.yml` (lint + build)

### Removed

- Generated code directory (`gen/`) -- consumers now use BSR Generated SDKs or `buf generate`
- `EngineHttpService` gRPC service and associated request/response messages from `apigame/v1`
- `google/api/annotations.proto` import (no longer needed without gRPC service)
- Vendored `proto/google/` googleapis protos
- `buf.gen.yaml` code generation config (consumers generate in their own repos)
- `go.mod` and `go.sum` (proto-only repo needs no Go module)
- `examples/` and `scripts/` directories

### Added

- Apache 2.0 LICENSE file (Copyright 2026 EchoTools)
- `archive/` directory for deprecated protos (`api/v1`, `apigrpc/v1`, `types/v1`)

### Deprecated

- `api/v1/`, `apigrpc/v1/`, `types/v1/` packages moved to `archive/` (excluded from buf module)
