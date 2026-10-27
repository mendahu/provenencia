package valuecompare

import (
	"crypto/sha256"
	"encoding/hex"
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
//
// FormalFirst says each set's first token is its head, the name the others
// stand for. Directions say which name a fold keeps for a pair whose sets each
// hold the other (bill and william). Both are for the reconciler's fold
// (AliasIndex.Expands); matching ignores them, and a catalog load sets neither.
type AliasTable struct {
	Key, Origin, Label string
	Kind               AliasKind
	Roles              []NameRole
	Sets               [][]string
	FormalFirst        bool
	Directions         []AliasDirection
}

// AliasDirection is the name a fold keeps for a pair listed both ways: Formal
// is kept, Nickname folds into it.
type AliasDirection struct{ Formal, Nickname string }

// AliasIndex is a catalog's alias tables, indexed once per load. The zero
// index hits nothing.
type AliasIndex struct {
	sets    []aliasSet
	byToken map[string][]int // token -> indexes into sets that hold it
	// prefer holds {Formal, Nickname} (AliasTokens) for each table Direction.
	prefer map[AliasDirection]bool
}

// aliasSet is one equivalence set, the roles that may use it, and its kind.
type aliasSet struct {
	tokens map[string]bool
	roles  map[NameRole]bool
	kind   AliasKind
	// head is the AliasToken of the set's first token, only in a FormalFirst
	// table.
	head string
	// key is sha256(table key, 0x00, smallest token), hex-encoded. It is a fixed
	// size however long the table key or a token is, and not the set's tokens: a
	// key holding every token would be copied for each word of a set with
	// thousands. Tokens of one set share it, so two words that Hit share it. A
	// set sharing a smallest token with another set of its table only makes a
	// wider shortlist.
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
		for _, d := range table.Directions {
			formal, ok1 := AliasToken(d.Formal, w)
			nick, ok2 := AliasToken(d.Nickname, w)
			if !ok1 || !ok2 {
				continue
			}
			if idx.prefer == nil {
				idx.prefer = map[AliasDirection]bool{}
			}
			idx.prefer[AliasDirection{formal, nick}] = true
		}
		for _, raw := range table.Sets {
			head := ""
			if table.FormalFirst {
				for _, token := range raw {
					if k, ok := AliasToken(token, w); ok {
						head = k
						break
					}
				}
			}
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
			idx.sets = append(idx.sets, aliasSet{tokens: tokens, roles: roles, kind: table.Kind, head: head, key: aliasSetKey(table.Key, sorted[0])})
		}
	}
	return idx
}

// aliasSetKey is the fixed-size key of a set: see aliasSet.key.
func aliasSetKey(tableKey, smallest string) string {
	sum := sha256.Sum256([]byte(tableKey + "\x00" + smallest))
	return hex.EncodeToString(sum[:])
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

// Expands reports whether a fold keeps into over from: a set that allows role,
// headed by into, holds from. When a set headed by from also holds into, a
// Direction in any table must name into as the one kept. from and into are
// comparison tokens. Without FormalFirst nothing expands.
func (a AliasIndex) Expands(from, into string, role NameRole) bool {
	forward, reverse := false, false
	for _, i := range a.byToken[from] {
		set := a.sets[i]
		if !set.roles[role] || !set.tokens[into] {
			continue
		}
		forward = forward || set.head == into
		reverse = reverse || set.head == from
	}
	return forward && (!reverse || a.prefer[AliasDirection{into, from}])
}

// SetKeys is the key of every set that holds token (an AliasToken): a
// fixed-size digest of its table key and smallest token. Roles are ignored, so a shortlist may
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
