package graphalign

import (
	"testing"

	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestBridgeFrom(t *testing.T) {
	a, b, c := []byte{1}, []byte{2}, []byte{3}
	ka, kb := vocabkey.BridgeKey{Term: "ka"}, vocabkey.BridgeKey{Term: "kb"}
	br := Bridge{A: a, B: b, KeyA: ka, KeyB: kb}
	if o, k, ok := br.From(a); !ok || string(o) != string(b) || k != ka {
		t.Errorf("From(A) = %v, %v, %v", o, k, ok)
	}
	if o, k, ok := br.From(b); !ok || string(o) != string(a) || k != kb {
		t.Errorf("From(B) = %v, %v, %v", o, k, ok)
	}
	if _, _, ok := br.From(c); ok {
		t.Errorf("From(neither) ok")
	}
}
