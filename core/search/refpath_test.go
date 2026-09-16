package search

import "testing"

func TestClassifyRefQuery(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		key   string
		exact bool
	}{
		{"empty", "", "", false},
		{"whitespace", "   ", "", false},
		{"prefix alone", "SRC", "", false},
		{"prefix and dash", "SRC-", "", false},
		{"canonical partial", "SRC-ZZ", "SRC-ZZ", false},
		{"canonical complete", "SRC-ZZ9K2", "SRC-ZZ9K2", true},
		{"candidate marker", "PER-C", "PER-C", false},
		{"candidate marker and dash", "PER-C-", "PER-C-", false},
		{"candidate partial", "PER-C-7KD", "PER-C-7KD", false},
		{"candidate complete", "PER-C-7KD45", "PER-C-7KD45", true},
		{"canonical token starting with C", "PER-C4N2P", "PER-C4N2P", true},
		{"lowercase is normalized", "per-c-7kd45", "PER-C-7KD45", true},
		{"surrounding space is trimmed", "  SRC-ZZ9K2  ", "SRC-ZZ9K2", true},
		{"free text", "Ilminster parish", "", false},
		{"hyphenated free text", "Smith-Jones", "", false},
		{"excluded letter", "SRC-IIIII", "", false},
		{"overlong token", "SRC-ZZ9K2X", "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key, exact := classifyRefQuery(c.in)
			if key != c.key || exact != c.exact {
				t.Fatalf("classifyRefQuery(%q) = (%q, %v), want (%q, %v)",
					c.in, key, exact, c.key, c.exact)
			}
		})
	}
}

// Every keystroke of a candidate ref should stay on the ref fast path. The
// earlier local regex had no room for a second dash, so PER-C- onward fell
// through to FTS.
func TestClassifyRefQueryCandidateTypingSequence(t *testing.T) {
	full := "PER-C-7KD45"
	for i := len("PER-") + 1; i <= len(full); i++ {
		typed := full[:i]
		key, exact := classifyRefQuery(typed)
		if key != typed {
			t.Fatalf("typing %q dropped off the ref path (key %q)", typed, key)
		}
		if want := typed == full; exact != want {
			t.Fatalf("typing %q: exact = %v, want %v", typed, exact, want)
		}
	}
}
