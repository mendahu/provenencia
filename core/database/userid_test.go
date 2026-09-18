package database

import "testing"

func TestRequireUserID(t *testing.T) {
	sentinel := ErrClosed
	tests := []struct {
		name    string
		userID  []byte
		wantErr bool
	}{
		{name: "exact 16 bytes", userID: make([]byte, 16)},
		{name: "empty", userID: nil, wantErr: true},
		{name: "empty slice", userID: []byte{}, wantErr: true},
		{name: "too short", userID: make([]byte, 8), wantErr: true},
		{name: "too long", userID: make([]byte, 17), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireUserID(tt.userID, sentinel)
			if tt.wantErr {
				if err != sentinel {
					t.Fatalf("got %v want sentinel", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
