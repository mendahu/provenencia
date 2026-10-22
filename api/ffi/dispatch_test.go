package ffi

import "testing"

func TestCallRouter(t *testing.T) {
	tests := []struct {
		name   string
		method int32
	}{
		{name: "unspecified method", method: MethodUnspecified},
		{name: "unknown method", method: 100},
		{name: "reserved update citation with observations", method: 71},
		{name: "claim evidence with no request", method: MethodGetPromoteClaimEvidence},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Call(tt.method, nil); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
