// Package oldreader decodes new wire bytes with the telemetry.v2 code generated
// from the schema BEFORE native records existed (scripts/compat-test.sh
// generates it from the pinned pre-native commit into internal/oldgen, under its own import path, so
// this binary never links the new code).
//
// The claim under test: an old reader sees a native batch as PRESENT BUT
// UNKNOWN. It must not decode as a fully understood envelope, and it must not
// vanish.
package oldreader

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	oldv2 "github.com/EchoTools/nevr-proto/compat/internal/oldgen/telemetry/v2"
	"github.com/EchoTools/nevr-proto/compat/vector"
)

// onlyField parses b as exactly one field and returns its number, type and
// raw value.
func onlyField(t *testing.T, b []byte) (protowire.Number, protowire.Type, []byte) {
	t.Helper()
	num, typ, n := protowire.ConsumeTag(b)
	if n < 0 {
		t.Fatalf("unknown bytes do not start with a tag: %v", protowire.ParseError(n))
	}
	m := protowire.ConsumeFieldValue(num, typ, b[n:])
	if m < 0 {
		t.Fatalf("field %d value malformed: %v", num, protowire.ParseError(m))
	}
	if n+m != len(b) {
		t.Fatalf("unknown bytes hold more than one field: %d of %d bytes consumed", n+m, len(b))
	}
	return num, typ, b[n : n+m]
}

func TestOldReaderSeesNativeBatchAsUnknown(t *testing.T) {
	wire := vector.NativeBatchEnvelope()

	env := &oldv2.Envelope{}
	if err := proto.Unmarshal(wire, env); err != nil {
		t.Fatalf("old reader rejected the envelope: %v", err)
	}

	// Not understood: no known oneof case is set.
	if env.GetMessage() != nil {
		t.Fatalf("old reader decoded a known case %T; the native batch was misread as something it knows", env.GetMessage())
	}

	// But present: the whole batch is in the unknown fields, as field 5.
	unknown := env.ProtoReflect().GetUnknown()
	t.Logf("old reader: Message=%v, unknown=%d bytes", env.GetMessage(), len(unknown))
	num, typ, val := onlyField(t, unknown)
	if num != 5 || typ != protowire.BytesType {
		t.Fatalf("unknown field is %d/%v, want 5/bytes", num, typ)
	}
	inner, n := protowire.ConsumeBytes(val)
	if n < 0 || !bytes.Equal(inner, vector.NativeBatch()) {
		t.Fatalf("unknown field 5 does not carry the batch bytes exactly")
	}
	t.Logf("old reader: unknown field %d (%v), %d-byte value, equal to the hand-assembled NativeBatch", num, typ, len(inner))

	// And preserved: re-encoding through the old code loses nothing.
	out, err := proto.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, wire) {
		t.Fatalf("old reader re-encode changed the bytes: %d -> %d", len(wire), len(out))
	}
	t.Logf("old reader: re-encode byte-identical (%d bytes)", len(out))
}

func TestOldReaderKeepsHeaderNativeTimeUnit(t *testing.T) {
	env := &oldv2.Envelope{}
	if err := proto.Unmarshal(vector.HeaderEnvelope(), env); err != nil {
		t.Fatal(err)
	}
	h := env.GetHeader()
	if h == nil || h.GetFormatVersion() != 2 {
		t.Fatalf("old reader lost the known header: %v", h)
	}
	unknown := h.ProtoReflect().GetUnknown()
	if !bytes.Equal(unknown, vector.HeaderNativeTimeUnit()) {
		t.Fatalf("header unknown fields = %x, want field 103 = %d (%x)", unknown, vector.NativeTimeUnitNs, vector.HeaderNativeTimeUnit())
	}
	t.Logf("old reader: header format_version=%d known; native_time_unit_ns kept as unknown %x", h.GetFormatVersion(), unknown)
}
