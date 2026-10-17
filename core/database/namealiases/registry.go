package namealiases

import (
	"sync"

	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func aliasPartRoles() []valuecompare.NameRole {
	return []valuecompare.NameRole{valuecompare.RoleGiven, valuecompare.RoleNick, valuecompare.RoleUntyped}
}

// productAliasTables is the create-time seed. Install writes this list.
// Scoring does not read it. Tokens are normalized comparison forms.
var productAliasTables = []valuecompare.AliasTable{
	{
		Key: "abbreviations", Origin: vocabkey.OriginProvenencia, Label: "Abbreviations",
		Kind: valuecompare.AliasAbbreviation, Roles: aliasPartRoles(),
		Sets: [][]string{
			{"james", "jas"},
			{"william", "wm"},
			{"charles", "chas"},
			{"thomas", "thos"},
			{"george", "geo"},
			{"elizabeth", "eliz"},
			{"john", "jno"},
			{"robert", "robt"},
		},
	},
	{
		Key: "nicknames", Origin: vocabkey.OriginProvenencia, Label: "Nicknames",
		Kind: valuecompare.AliasNickname, Roles: aliasPartRoles(),
		// Each set is one canonical name and its direct nicknames, from
		// carltonnorthern/nicknames. A nickname shared by two canonical
		// names does not join those sets. License and credit:
		// third_party/nicknames/.
		Sets: nicknameSets,
	},
}

// aliasTables is the product seed plus anything Register appended.
// Install writes it. Scoring does not read it.
var (
	aliasMu     sync.Mutex
	aliasTables []valuecompare.AliasTable
)

func init() {
	resetAliasTables()
}

func resetAliasTables() {
	aliasTables = append([]valuecompare.AliasTable(nil), productAliasTables...)
}

// AliasTables returns the seed Install writes. Scoring does not read this.
func AliasTables() []valuecompare.AliasTable {
	aliasMu.Lock()
	defer aliasMu.Unlock()
	out := make([]valuecompare.AliasTable, len(aliasTables))
	copy(out, aliasTables)
	return out
}

// Register appends a table. Future plugins call this before catalog create
// so Install writes the table. Enabling a plugin on an existing catalog is later.
func Register(table valuecompare.AliasTable) {
	aliasMu.Lock()
	defer aliasMu.Unlock()
	aliasTables = append(aliasTables, table)
}

// ResetAliasTablesForTest restores the product seed. Call from tests only.
func ResetAliasTablesForTest() {
	aliasMu.Lock()
	defer aliasMu.Unlock()
	resetAliasTables()
}
