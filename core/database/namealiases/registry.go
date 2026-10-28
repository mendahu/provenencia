package namealiases

import (
	"sync"

	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// productAliasTables is the create-time seed. Install writes this list.
// Matching reads the catalog rows; the reconciler cache reads the product seed
// through valuerules.Product. Tokens are normalized comparison forms. Each set
// is a head followed by the names that stand for it (FormalFirst).
var productAliasTables = []valuecompare.AliasTable{
	{
		Key: "abbreviations", Origin: vocabkey.OriginProvenencia, Label: "Abbreviations",
		Kind: valuecompare.AliasAbbreviation, Roles: valuecompare.AliasRoles(), FormalFirst: true,
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
		Kind: valuecompare.AliasNickname, Roles: valuecompare.AliasRoles(),
		// Each set is one canonical name and its direct nicknames, from
		// carltonnorthern/nicknames. A nickname shared by two canonical
		// names does not join those sets. License and credit:
		// third_party/nicknames/.
		Sets:        nicknameSets,
		FormalFirst: true,
		Directions:  nicknameDirections,
	},
}

// aliasTables is the product seed plus anything Register appended.
// Install writes it. Matching reads the catalog rows, not this variable.
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

// ProductTables returns a copy of the product seed alone, without anything
// Register appended. The reconciler cache reads it (valuerules.Product), so a
// plugin table cannot change the cache.
func ProductTables() []valuecompare.AliasTable {
	return append([]valuecompare.AliasTable(nil), productAliasTables...)
}

// AliasTables returns the seed Install writes. Matching reads the catalog
// rows; the reconciler cache reads ProductTables.
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
