package ffi

import "testing"

func TestCallRouter(t *testing.T) {
	tests := []struct {
		name   string
		method int32
	}{
		{name: "unspecified method", method: MethodUnspecified},
		{name: "unknown method", method: 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Call(tt.method, nil); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
