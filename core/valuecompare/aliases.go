package valuecompare

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
			idx.sets = append(idx.sets, aliasSet{tokens: tokens, roles: roles, kind: table.Kind})
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
