package shortlist

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Index maps shortlist keys to handle ids, per kind. It keeps the rules it was
// built with, so a lookup never keys differently from the build, and the keys
// each handle is filed under, so a handle can be refiled. It is patched by Put
// and Remove; the caller serializes them with Lookup. Postings are sorted by
// bytes.Compare, so an index patched to a state equals one built to it.
type Index struct {
	rules   Rules
	subject subjectcompare.Rules
	value   valuecompare.Rules
	metas   subjectcompare.Metas
	byKey   map[vocabkey.SubjectType]map[string]posting
	filed   map[string]filing
}

// posting is the ids filed under one key. It holds the key string the byKey map
// is keyed by, so the filings that name the key share its bytes.
type posting struct {
	key string
	ids [][]byte // sorted by bytes.Compare
}

// filing is where one handle is filed.
type filing struct {
	kind vocabkey.SubjectType
	keys []string // the index keys id is filed under, sorted; each is a posting's key
}

// NewIndex is an empty index under the given rules.
func NewIndex(r Rules, subject subjectcompare.Rules, value valuecompare.Rules, metas subjectcompare.Metas) *Index {
	return &Index{rules: r, subject: subject, value: value, metas: metas,
		byKey: map[vocabkey.SubjectType]map[string]posting{}, filed: map[string]filing{}}
}

// Put files a copy of id under each index key of v, in place of wherever it
// was filed before: a key it no longer has loses it, and a handle with no keys
// is not filed. A nil index ignores it.
func (x *Index) Put(kind vocabkey.SubjectType, id []byte, v subjectcompare.Values) {
	if x == nil {
		return
	}
	keys, _ := Keys(kind, v, x.metas, x.rules, x.subject, x.value)
	old, had := x.filed[string(id)]
	if had && old.kind != kind {
		x.Remove(id)
		old, had = filing{}, false
	}
	if had && slices.Equal(old.keys, keys) {
		return
	}
	for _, k := range old.keys {
		if _, found := slices.BinarySearch(keys, k); !found {
			x.unfile(kind, k, id)
		}
	}
	if len(keys) == 0 {
		delete(x.filed, string(id))
		return
	}
	id = bytes.Clone(id)
	if x.byKey[kind] == nil {
		x.byKey[kind] = map[string]posting{}
	}
	shared := make([]string, len(keys))
	for i, k := range keys {
		post, ok := x.byKey[kind][k]
		if !ok {
			post.key = k
		}
		shared[i] = post.key
		if _, found := slices.BinarySearch(old.keys, k); found {
			continue
		}
		at, _ := slices.BinarySearchFunc(post.ids, id, bytes.Compare)
		post.ids = slices.Insert(post.ids, at, id)
		x.byKey[kind][post.key] = post // by post.key: assigning k would replace the map's key string
	}
	x.filed[string(id)] = filing{kind: kind, keys: shared}
}

// Remove takes id out of every posting it is filed in. A nil index ignores it.
func (x *Index) Remove(id []byte) {
	if x == nil {
		return
	}
	old, had := x.filed[string(id)]
	if !had {
		return
	}
	for _, k := range old.keys {
		x.unfile(old.kind, k, id)
	}
	delete(x.filed, string(id))
}

// unfile takes id out of the posting of key, deleting an emptied posting and an
// emptied kind map.
func (x *Index) unfile(kind vocabkey.SubjectType, key string, id []byte) {
	post := x.byKey[kind][key]
	at, found := slices.BinarySearchFunc(post.ids, id, bytes.Compare)
	if !found {
		return
	}
	post.ids = slices.Delete(post.ids, at, at+1)
	if len(post.ids) == 0 {
		delete(x.byKey[kind], key)
		if len(x.byKey[kind]) == 0 {
			delete(x.byKey, kind)
		}
		return
	}
	x.byKey[kind][post.key] = post
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
		for _, id := range x.byKey[kind][k].ids {
			if !seen[string(id)] {
				seen[string(id)] = true
				out = append(out, bytes.Clone(id))
			}
		}
	}
	slices.SortFunc(out, bytes.Compare)
	return out
}
