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
