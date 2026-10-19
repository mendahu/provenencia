package valuecompare

import (
	"slices"
	"sort"
)

// AliasKind is abbreviation or nickname. Both score the alias credit.
type AliasKind string

const (
	AliasAbbreviation AliasKind = "abbreviation"
	AliasNickname     AliasKind = "nickname"
)

// AliasTable is one contributed list. Origin records who contributed it
// (provenencia, user, or plugin:<id>) and does not turn the table off.
// A set is one equivalence group. Roles are the part roles that may use it.
type AliasTable struct {
	Key, Origin, Label string
	Kind               AliasKind
	Roles              []NameRole
	Sets               [][]string
}

// AliasIndex is a catalog's alias tables, indexed once per load. The zero
// index hits nothing.
type AliasIndex struct {
	sets    []aliasSet
	byToken map[string][]int // token -> indexes into sets that hold it
}

// aliasSet is one equivalence set, the roles that may use it, and its kind.
type aliasSet struct {
	tokens map[string]bool
	roles  map[NameRole]bool
	kind   AliasKind
	// key is the table key and the set's smallest token, joined by one space.
	// It is bounded by the table key and one token, not by the set: a key holding
	// every token would be copied for each word of a set with thousands. Tokens
	// of one set share it, so two words that Hit share it; a set sharing a
	// smallest token with another set of its table only makes a wider shortlist.
	key string
}

// AliasToken is the one rule for what an alias token is: the comparison token
// a name part produces, the text key of the token at Same. Storage and the
// index both use it. It is empty (false) for a token with no letters.
func AliasToken(token string, w WordRules) (string, bool) {
	return TextKey(token, Same, w)
}

// NewAliasIndex indexes tables once per load. Every token is keyed with
// AliasToken, and empty keys, and sets left with none, are dropped.
func NewAliasIndex(tables []AliasTable, w WordRules) AliasIndex {
	var idx AliasIndex
	for _, table := range tables {
		roles := make(map[NameRole]bool, len(table.Roles))
		for _, role := range table.Roles {
			roles[role] = true
		}
		for _, raw := range table.Sets {
			tokens := make(map[string]bool, len(raw))
			for _, token := range raw {
				if k, ok := AliasToken(token, w); ok {
					tokens[k] = true
				}
			}
			if len(tokens) == 0 {
				continue
			}
			if idx.byToken == nil {
				idx.byToken = map[string][]int{}
			}
			for token := range tokens {
				idx.byToken[token] = append(idx.byToken[token], len(idx.sets))
			}
			sorted := make([]string, 0, len(tokens))
			for token := range tokens {
				sorted = append(sorted, token)
			}
			sort.Strings(sorted)
			idx.sets = append(idx.sets, aliasSet{tokens: tokens, roles: roles, kind: table.Kind, key: table.Key + " " + sorted[0]})
		}
	}
	return idx
}

// Hit reports whether one set holds both tokens and allows both roles, and
// that set's kind. x and y are comparison tokens.
func (a AliasIndex) Hit(x, y string, rx, ry NameRole) (AliasKind, bool) {
	for _, i := range a.byToken[x] {
		set := a.sets[i]
		if set.tokens[x] && set.tokens[y] && set.roles[rx] && set.roles[ry] {
			return set.kind, true
		}
	}
	return "", false
}

// SetKeys is the key of every set that holds token (an AliasToken): its table
// key and smallest token, joined by one space. Roles are ignored, so a shortlist may
// include more than Hit allows. Sorted, no duplicates; nil for the zero index.
// Sets do not chain, so no word key can join two words of a set, but a set can
// be a key: two words that Hit share it.
func (a AliasIndex) SetKeys(token string) []string {
	var out []string
	for _, i := range a.byToken[token] {
		out = append(out, a.sets[i].key)
	}
	sort.Strings(out)
	return slices.Compact(out)
}
