package valuecompare

import (
	"slices"
	"sort"
)

// wordKeyer is a comparer that also says which keys shortlisting looks a value
// up by (comparison-layers.md, layer 6 procedure, step 1).
type wordKeyer interface {
	wordKeys(v Value, r Rules) []string
}

// WordKeys are the keys shortlisting looks a value up by. Two values Compare
// calls Same or Equivalent share one: the whole key at Equivalent, or, for an
// alias, the set. Values that share a word that is not an initial share one
// too. Initials and generation words are never keys alone. A value type with no
// word keys, and a value with no word, has none.
func WordKeys(v Value, r Rules) []string {
	c, ok := comparers[v.Type]
	if !ok {
		return nil
	}
	k, ok := c.(wordKeyer)
	if !ok {
		return nil
	}
	return k.wordKeys(v, r)
}

// keyList is the sorted, duplicate-free keys: the whole key at Equivalent when
// the value has one, then the folded words that are not initials.
func keyList(v Value, r Rules, words []string, extra []string) []string {
	var out []string
	if k, ok := Key(v, Equivalent, r); ok {
		out = append(out, "="+k)
	}
	for _, word := range words {
		if !IsInitial(word) {
			out = append(out, "w:"+FoldWord(word, r.Words.Folds))
		}
	}
	out = append(out, extra...)
	sort.Strings(out)
	return slices.Compact(out)
}

// wordKeys: the name's words as nameWords reads them, so ignored parts and
// generation words are never keys, plus the alias sets that hold a word.
func (name) wordKeys(v Value, r Rules) []string {
	words, _ := nameWords(v, r)
	if len(words) == 0 {
		return nil
	}
	texts := make([]string, len(words))
	var sets []string
	for i, w := range words {
		texts[i] = w.text
		if t, ok := AliasToken(w.text, r.Words); ok {
			for _, s := range r.Names.Aliases.SetKeys(t) {
				sets = append(sets, "a:"+s)
			}
		}
	}
	return keyList(v, r, texts, sets)
}

// wordKeys: text has no alias keys, because text compare never reads aliases.
func (text) wordKeys(v Value, r Rules) []string {
	words := SplitWords(v.Text, r.Words)
	if len(words) == 0 {
		return nil
	}
	return keyList(v, r, words, nil)
}
