# Local Go codegen (red-green a proto change before publishing)

`nevr-proto` is proto-only. The canonical Go SDK is published to the Buf Schema
Registry (`buf.build/echotools/nevr-api`) and consumed downstream (e.g. `tape`)
as the module `buf.build/gen/go/echotools/nevr-api/protocolbuffers/go`, pinned in
that repo's `go.mod`.

To make a proto change **test-first** in a consumer *before* any BSR publish,
regenerate that exact SDK locally and point the consumer at it with a temporary
`replace` directive.

`buf.gen.yaml` reproduces the BSR SDK byte-for-byte-compatibly:

- plugin `buf.build/protocolbuffers/go` pinned to **v1.36.11** (the exact BSR gen)
- **`default_api_level=API_HYBRID`** — the BSR SDK is hybrid: each message has
  exported struct fields (`protogen:"hybrid.v1"`) *and* opaque accessors/builders,
  shipped as the dual build-tagged pair (`*.pb.go` + `*_protoopaque.pb.go`).
  Matching this keeps consumer code — including struct literals like
  `&EchoArenaFrame{GameStatus: ...}` — compiling identically. `API_OPAQUE` would
  emit unexported `xxx_hidden_*` fields and break those literals.
- output honors each file's `go_package`, so files land under
  `gen/go/buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/<pkg>/...` — a
  self-contained module rooted at that path.

`gen/` is gitignored; generated code is never committed here.

## Dev loop

Assumes `nevr-proto/` and the consumer (`tape/`) are sibling checkouts under
`~/src/`.

```bash
# 1. Generate the SDK from the current .proto (opaque hybrid Go).
cd ~/src/nevr-proto
buf generate

# 2. buf does not emit a go.mod for the generated module — drop one in so the
#    module root can be a `replace` target. (Matches the BSR module's go.mod.)
GENMOD=gen/go/buf.build/gen/go/echotools/nevr-api/protocolbuffers/go
cat > "$GENMOD/go.mod" <<'EOF'
module buf.build/gen/go/echotools/nevr-api/protocolbuffers/go

go 1.23

require google.golang.org/protobuf v1.36.11
EOF

# 3. In the CONSUMER, add a DEV-ONLY replace. NEVER commit this.
cd ~/src/tape
go mod edit -replace \
  buf.build/gen/go/echotools/nevr-api/protocolbuffers/go=../nevr-proto/gen/go/buf.build/gen/go/echotools/nevr-api/protocolbuffers/go

# 4. Now red-green against the locally-generated types.
go build ./...
go test ./...

# 5. Revert the replace BEFORE committing the consumer, so its committed go.mod
#    stays pinned to the real BSR version.
go mod edit -dropreplace buf.build/gen/go/echotools/nevr-api/protocolbuffers/go
go build ./...   # confirm it still builds against the pinned BSR module
```

## Shipping the proto change for real

Once the change is green locally:

1. Merge the `.proto` change to `nevr-proto` main; CI `buf push` publishes a new
   BSR module version.
2. In the consumer, `go get` that new version (bump the pin in `go.mod`), drop
   the dev `replace` if still present, and commit.

The `replace` is a scaffold for the gap between "proto edited" and "BSR
published" — it must never appear in a committed consumer `go.mod`.
