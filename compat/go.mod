module github.com/EchoTools/nevr-proto/compat

go 1.23

require (
	buf.build/gen/go/echotools/nevr-api/protocolbuffers/go v0.0.0-local
	google.golang.org/protobuf v1.36.11
)

// The branch's own generated SDK, from `buf generate` (gitignored). Never the
// published module: this module tests the schema in this tree.
replace buf.build/gen/go/echotools/nevr-api/protocolbuffers/go => ../gen/go/buf.build/gen/go/echotools/nevr-api/protocolbuffers/go
