// Package vector hand-assembles telemetry.v2 wire vectors with protowire.
//
// It imports no generated code on purpose. A vector built by the generated
// encoder can only show that the generator agrees with itself. These bytes are
// written from the .proto field numbers directly, so the old reader, the new
// reader and any other-language decoder are all tested against the same
// independent bytes.
package vector

import "google.golang.org/protobuf/encoding/protowire"

// Field numbers, copied from the .proto by hand. If one of these disagrees with
// the schema, newreader fails.
const (
	envelopeHeader      = 1 // Envelope.header
	envelopeNativeBatch = 5 // Envelope.native_batch

	headerFormatVersion    = 3   // CaptureHeader.format_version
	headerNativeTimeUnitNs = 103 // CaptureHeader.native_time_unit_ns

	batchRecords      = 1 // NativeBatch.records
	batchDispositions = 2 // NativeBatch.dispositions

	recSequence  = 1 // NativeRecord.sequence
	recElapsed   = 2 // NativeRecord.elapsed
	recStage     = 3 // NativeRecord.stage
	recMsgType   = 4 // NativeRecord.msg_type (fixed64)
	recMailbox   = 5 // NativeRecord.mailbox (fixed64)
	recConnIndex = 6 // NativeRecord.connection_index (optional)
	recConnEpoch = 7 // NativeRecord.connection_epoch (optional)
	recPayload   = 8 // NativeRecord.payload
	recCauses    = 9 // NativeRecord.causes (packed)

	dispFirst  = 1 // NativeSequenceDisposition.first_sequence
	dispCount  = 2 // NativeSequenceDisposition.count
	dispReason = 3 // NativeSequenceDisposition.reason
	dispStage  = 4 // NativeSequenceDisposition.stage
)

// Enum values, copied from the .proto by hand.
const (
	StageInput   = 1 // NATIVE_STAGE_INPUT
	StageEmitted = 2 // NATIVE_STAGE_EMITTED
	ReasonLost   = 2 // NATIVE_DISPOSITION_REASON_LOST
)

// Record is the logical content of one NativeRecord in the vector.
type Record struct {
	Sequence, Elapsed uint64
	Stage             uint64
	MsgType, Mailbox  uint64
	// Optional connection identity: nil is absent, a pointer to 0 is
	// connection 0, which must still reach the wire.
	ConnIndex, ConnEpoch *uint64
	Payload              []byte
	Causes               []uint64
}

// Disposition is the logical content of one NativeSequenceDisposition.
type Disposition struct {
	First, Count, Reason, Stage uint64
}

// ReliableMailbox is the server's reliable lane, mailbox == u64::MAX.
const ReliableMailbox = ^uint64(0)

func u(v uint64) *uint64 { return &v }

// Records are the vector's native records. Sequences 1 and 2 share an elapsed
// time, so only sequence orders them. Sequence 2 is an empty EMITTED reply
// caused by 1, sent to connection 0 at epoch 0 (present zeros). Sequences 3-4 are a LOST gap, and 5 follows it.
var Records = []Record{
	{Sequence: 1, Elapsed: 1500, Stage: StageInput, MsgType: 0x8f3a_51c2_0d6e_b947, Mailbox: ReliableMailbox, ConnIndex: u(7), ConnEpoch: u(3), Payload: []byte{0x01, 0x02, 0x03, 0x00, 0xff}},
	{Sequence: 2, Elapsed: 1500, Stage: StageEmitted, MsgType: 0x1c44_09ab_e7f0_2d31, Mailbox: 0x0000_0000_0000_0010, ConnIndex: u(0), ConnEpoch: u(0), Causes: []uint64{1}},
	{Sequence: 5, Elapsed: 1501, Stage: StageEmitted, MsgType: 0x1c44_09ab_e7f0_2d31, Mailbox: 0x0000_0000_0000_0010, ConnIndex: u(9), ConnEpoch: u(2), Payload: []byte("native")},
}

// Dispositions are the vector's sequence dispositions.
var Dispositions = []Disposition{
	{First: 3, Count: 2, Reason: ReasonLost, Stage: StageEmitted},
}

// NativeTimeUnitNs is the header unit in HeaderEnvelope: microseconds.
const NativeTimeUnitNs = 1000

func appendVarintField(b []byte, num protowire.Number, v uint64) []byte {
	if v == 0 {
		return b // proto3 omits a zero scalar
	}
	b = protowire.AppendTag(b, num, protowire.VarintType)
	return protowire.AppendVarint(b, v)
}

func appendOptionalVarintField(b []byte, num protowire.Number, v *uint64) []byte {
	if v == nil {
		return b
	}
	b = protowire.AppendTag(b, num, protowire.VarintType)
	return protowire.AppendVarint(b, *v) // written even when 0: presence
}

func appendFixed64Field(b []byte, num protowire.Number, v uint64) []byte {
	if v == 0 {
		return b
	}
	b = protowire.AppendTag(b, num, protowire.Fixed64Type)
	return protowire.AppendFixed64(b, v)
}

func appendBytesField(b []byte, num protowire.Number, v []byte) []byte {
	if len(v) == 0 {
		return b
	}
	b = protowire.AppendTag(b, num, protowire.BytesType)
	return protowire.AppendBytes(b, v)
}

func record(r Record) []byte {
	var b []byte
	b = appendVarintField(b, recSequence, r.Sequence)
	b = appendVarintField(b, recElapsed, r.Elapsed)
	b = appendVarintField(b, recStage, r.Stage)
	b = appendFixed64Field(b, recMsgType, r.MsgType)
	b = appendFixed64Field(b, recMailbox, r.Mailbox)
	b = appendOptionalVarintField(b, recConnIndex, r.ConnIndex)
	b = appendOptionalVarintField(b, recConnEpoch, r.ConnEpoch)
	b = appendBytesField(b, recPayload, r.Payload)
	if len(r.Causes) > 0 {
		var packed []byte
		for _, c := range r.Causes {
			packed = protowire.AppendVarint(packed, c)
		}
		b = appendBytesField(b, recCauses, packed)
	}
	return b
}

func disposition(d Disposition) []byte {
	var b []byte
	b = appendVarintField(b, dispFirst, d.First)
	b = appendVarintField(b, dispCount, d.Count)
	b = appendVarintField(b, dispReason, d.Reason)
	b = appendVarintField(b, dispStage, d.Stage)
	return b
}

// NativeBatch is the NativeBatch message bytes: the value of Envelope field 5.
func NativeBatch() []byte {
	var b []byte
	for _, r := range Records {
		b = protowire.AppendTag(b, batchRecords, protowire.BytesType)
		b = protowire.AppendBytes(b, record(r))
	}
	for _, d := range Dispositions {
		b = protowire.AppendTag(b, batchDispositions, protowire.BytesType)
		b = protowire.AppendBytes(b, disposition(d))
	}
	return b
}

// NativeBatchEnvelope is Envelope{native_batch = NativeBatch()}.
func NativeBatchEnvelope() []byte {
	b := protowire.AppendTag(nil, envelopeNativeBatch, protowire.BytesType)
	return protowire.AppendBytes(b, NativeBatch())
}

// HeaderNativeTimeUnit is the CaptureHeader field 103 bytes alone.
func HeaderNativeTimeUnit() []byte {
	return appendVarintField(nil, headerNativeTimeUnitNs, NativeTimeUnitNs)
}

// HeaderEnvelope is Envelope{header{format_version: 2, native_time_unit_ns: 1000}}.
func HeaderEnvelope() []byte {
	h := appendVarintField(nil, headerFormatVersion, 2)
	h = append(h, HeaderNativeTimeUnit()...)
	b := protowire.AppendTag(nil, envelopeHeader, protowire.BytesType)
	return protowire.AppendBytes(b, h)
}
