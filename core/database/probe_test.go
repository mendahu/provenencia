package database

import "testing"

// TEMPORARY (Spike 1 PR3): delete with probe.go when PR4 lands.
func TestProbe(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "temp db round trip", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Probe()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
