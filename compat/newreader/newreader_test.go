// Package newreader decodes the same hand-assembled vectors with the
// telemetry.v2 code generated from this branch (`buf generate`, replaced into
// this module by scripts/compat-test.sh).
package newreader

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/proto"

	telemetryv2 "buf.build/gen/go/echotools/nevr-api/protocolbuffers/go/telemetry/v2"

	"github.com/EchoTools/nevr-proto/compat/vector"
)

func expectedBatch() *telemetryv2.NativeBatch {
	b := &telemetryv2.NativeBatch{}
	for _, r := range vector.Records {
		b.Records = append(b.Records, &telemetryv2.NativeRecord{
			Sequence:        r.Sequence,
			Elapsed:         r.Elapsed,
			Stage:           telemetryv2.NativeStage(r.Stage),
			MsgType:         r.MsgType,
			Mailbox:         r.Mailbox,
			ConnectionIndex: r.ConnIndex,
			ConnectionEpoch: r.ConnEpoch,
			Payload:         r.Payload,
			Causes:          r.Causes,
		})
	}
	for _, d := range vector.Dispositions {
		b.Dispositions = append(b.Dispositions, &telemetryv2.NativeSequenceDisposition{
			FirstSequence: d.First,
			Count:         d.Count,
			Reason:        telemetryv2.NativeDispositionReason(d.Reason),
			Stage:         telemetryv2.NativeStage(d.Stage),
		})
	}
	return b
}

func TestNewReaderDecodesNativeBatch(t *testing.T) {
	env := &telemetryv2.Envelope{}
	if err := proto.Unmarshal(vector.NativeBatchEnvelope(), env); err != nil {
		t.Fatal(err)
	}
	got := env.GetNativeBatch()
	if got == nil {
		t.Fatalf("new reader did not decode Envelope.native_batch; case = %T", env.GetMessage())
	}
	if u := env.ProtoReflect().GetUnknown(); len(u) != 0 {
		t.Fatalf("new reader left %d unknown bytes", len(u))
	}
	if want := expectedBatch(); !proto.Equal(got, want) {
		t.Fatalf("decoded batch differs:\n got  %v\n want %v", got, want)
	}
	t.Logf("new reader: %d records, %d dispositions, every field equal", len(got.GetRecords()), len(got.GetDispositions()))
	if s := got.GetRecords()[0].GetStage(); s != telemetryv2.NativeStage_NATIVE_STAGE_INPUT {
		t.Fatalf("record 1 stage = %v", s)
	}
	if rec := got.GetRecords()[1]; !rec.HasConnectionIndex() || rec.GetConnectionIndex() != 0 || !rec.HasConnectionEpoch() {
		t.Fatalf("record 2: connection 0 at epoch 0 must decode as PRESENT zeros, got %v", rec)
	}
	if r := got.GetDispositions()[0].GetReason(); r != telemetryv2.NativeDispositionReason_NATIVE_DISPOSITION_REASON_LOST {
		t.Fatalf("disposition reason = %v", r)
	}
}

// The generated encoder must agree with the hand assembly byte for byte, which
// checks the vector's field numbers and wire types against the schema.
func TestNewEncoderMatchesHandVector(t *testing.T) {
	env := &telemetryv2.Envelope{Message: &telemetryv2.Envelope_NativeBatch{NativeBatch: expectedBatch()}}
	out, err := proto.MarshalOptions{Deterministic: true}.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if want := vector.NativeBatchEnvelope(); !bytes.Equal(out, want) {
		t.Fatalf("generated encoding differs from hand vector:\n gen  %x\n hand %x", out, want)
	}

	h := &telemetryv2.Envelope{Message: &telemetryv2.Envelope_Header{Header: &telemetryv2.CaptureHeader{
		FormatVersion: 2, NativeTimeUnitNs: vector.NativeTimeUnitNs,
	}}}
	out, err = proto.MarshalOptions{Deterministic: true}.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	if want := vector.HeaderEnvelope(); !bytes.Equal(out, want) {
		t.Fatalf("generated header differs from hand vector:\n gen  %x\n hand %x", out, want)
	}
}
