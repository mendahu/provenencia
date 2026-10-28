package valuecompare

import (
	"slices"
	"sort"
)

// The key tags. Each key is its tag then the folded text it is about.
const (
	keyWhole  = "="  // the whole key at Equivalent: both sides
	keyWord   = "w:" // a folded word that is not an initial: both sides
	keyAlias  = "a:" // an alias set a word belongs to: both sides
	keyGiven  = "g:" // a given-like whole word: filed by the index, asked by a lookup's prefixes
	keyPrefix = "p:" // a strict prefix: filed by the index, asked by a lookup's whole word
)

// wordKeyer is a comparer that also says which keys shortlisting files a value
// under and looks it up by (comparison-layers.md, layer 6 procedure, step 1).
type wordKeyer interface {
	wordKeys(v Value, r Rules) (index, lookup []string)
}

// WordKeys are the keys a value is filed under (index) and looked up by
// (lookup). Two values Compare calls Same or Equivalent meet: an index key of
// either is a lookup key of the other. Values that share a word that is not an
// initial meet too, and so do two given-like words where one starts the other
// (from Names.PrefixMinLength letters). Initials and generation words are never
// keys alone. A value type with no word keys, and a value with no word, has
// none. Both lists are sorted and duplicate-free.
func WordKeys(v Value, r Rules) (index, lookup []string) {
	c, ok := comparers[v.Type]
	if !ok {
		return nil, nil
	}
	k, ok := c.(wordKeyer)
	if !ok {
		return nil, nil
	}
	return k.wordKeys(v, r)
}

// keyList is the sorted, duplicate-free keys: the whole key at Equivalent when
// the value has one, then the folded words that are not initials.
func keyList(v Value, r Rules, words []string, extra []string) []string {
	var out []string
	if k, ok := Key(v, Equivalent, r); ok {
		out = append(out, keyWhole+k)
	}
	for _, word := range words {
		if !IsInitial(word) {
			out = append(out, keyWord+FoldWord(word, r.Words.Folds))
		}
	}
	return sortedUnique(append(out, extra...))
}

// wordKeys: the name's words as nameWords reads them, so ignored parts and
// generation words are never keys, plus the alias sets that hold a word, and
// the given-like words' prefix keys (prefixKeys), the only keys index and
// lookup do not share.
func (name) wordKeys(v Value, r Rules) (index, lookup []string) {
	words, generation := nameWords(v, r)
	if len(words) == 0 {
		return nil, nil
	}
	texts := make([]string, len(words))
	var sets []string
	for i, w := range words {
		texts[i] = w.text
		if t, ok := AliasToken(w.text, r.Words); ok {
			for _, s := range r.Names.Aliases.SetKeys(t) {
				sets = append(sets, keyAlias+s)
			}
		}
	}
	keys := keyList(v, r, texts, sets)
	if tooManyWords(words, generation, r) {
		return keys, slices.Clone(keys)
	}
	pi, pl := prefixKeys(words, r)
	return sortedUnique(append(slices.Clone(keys), pi...)), sortedUnique(append(keys, pl...))
}

// prefixKeys are the given-like words' prefix keys. A word files itself whole
// (g:) and its strict prefixes (p:); a lookup asks for the same word as a
// prefix (p:) and its strict prefixes as whole words (g:). So a word meets the
// words it starts and the words that start it, and a filed prefix never meets
// a filed prefix.
func prefixKeys(words []nameWord, r Rules) (index, lookup []string) {
	for _, w := range words {
		if !prefixWord(w, r) {
			continue
		}
		index = append(index, keyGiven+w.fold)
		lookup = append(lookup, keyPrefix+w.fold)
		for _, p := range prefixesOf(w.fold, r) {
			index = append(index, keyPrefix+p)
			lookup = append(lookup, keyGiven+p)
		}
	}
	return index, lookup
}

// sortedUnique sorts keys in place and drops the duplicates.
func sortedUnique(keys []string) []string {
	sort.Strings(keys)
	return slices.Compact(keys)
}

// wordKeys: text has no alias keys, because text compare never reads aliases.
func (text) wordKeys(v Value, r Rules) (index, lookup []string) {
	words := SplitWords(v.Text, r.Words)
	if len(words) == 0 {
		return nil, nil
	}
	keys := keyList(v, r, words, nil)
	return keys, slices.Clone(keys)
}
