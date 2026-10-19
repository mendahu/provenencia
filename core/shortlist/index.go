package shortlist

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Index maps shortlist keys to handle ids, per kind. It keeps the rules it was
// built with, so a lookup never keys differently from the build. It is
// read-only once built.
type Index struct {
	rules   Rules
	subject subjectcompare.Rules
	value   valuecompare.Rules
	metas   subjectcompare.Metas
	byKey   map[vocabkey.SubjectType]map[string][][]byte
}

// NewIndex is an empty index under the given rules.
func NewIndex(r Rules, subject subjectcompare.Rules, value valuecompare.Rules, metas subjectcompare.Metas) *Index {
	return &Index{rules: r, subject: subject, value: value, metas: metas, byKey: map[vocabkey.SubjectType]map[string][][]byte{}}
}

// Add files a copy of id under each index key of v.
func (x *Index) Add(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values) {
	keys, _ := Keys(kind, v, x.metas, x.rules, x.subject, x.value)
	if len(keys) == 0 {
		return
	}
	id = bytes.Clone(id)
	if x.byKey[kind] == nil {
		x.byKey[kind] = map[string][][]byte{}
	}
	for _, k := range keys {
		x.byKey[kind][k] = append(x.byKey[kind][k], id)
	}
}

// Lookup is the ids filed under any lookup key of v: deduplicated, sorted by
// bytes.Compare, each a copy. It is nil for a kind with no entry and for a nil
// index.
func (x *Index) Lookup(kind vocabkey.SubjectType, v subjectcompare.Values) [][]byte {
	if x == nil || x.byKey[kind] == nil {
		return nil
	}
	_, keys := Keys(kind, v, x.metas, x.rules, x.subject, x.value)
	seen := map[string]bool{}
	var out [][]byte
	for _, k := range keys {
		for _, id := range x.byKey[kind][k] {
			if !seen[string(id)] {
				seen[string(id)] = true
				out = append(out, bytes.Clone(id))
			}
		}
	}
	slices.SortFunc(out, bytes.Compare)
	return out
}
