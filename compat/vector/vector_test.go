package vector

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the committed .bin vectors from this package")

// The committed .bin files are what other-language decoders read. This test
// keeps them byte-equal to the hand assembly here.
func TestCommittedVectorsMatch(t *testing.T) {
	for name, want := range map[string][]byte{
		"native_batch.envelope.bin":       NativeBatchEnvelope(),
		"header_native_unit.envelope.bin": HeaderEnvelope(),
	} {
		path := filepath.Join("..", "..", "testdata", "telemetry", "v2", name)
		if *update {
			if err := os.WriteFile(path, want, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s: committed bytes differ from the hand assembly; rerun with -update if the change is deliberate", name)
		}
	}
}
