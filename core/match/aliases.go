package match

import "sync"

// aliasTables is the product seed plus anything Register appended.
// DefaultWordRules.Aliases points here, so a nil setting reads this list.
var (
	aliasMu     sync.Mutex
	aliasTables []AliasTable
)

func init() {
	resetAliasTables()
}

func resetAliasTables() {
	aliasTables = append([]AliasTable(nil), productAliasTables...)
}

// AliasTables returns the tables comparison uses when the caller does not
// pass its own list. The catalog replaces this when it has any row.
func AliasTables() []AliasTable {
	aliasMu.Lock()
	defer aliasMu.Unlock()
	out := make([]AliasTable, len(aliasTables))
	copy(out, aliasTables)
	return out
}

// Register appends a table. Future plugins call this before catalog create
// so Install writes the table. Enabling a plugin on an existing catalog is later.
func Register(table AliasTable) {
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

// aliasTablesOf resolves a comparer setting. Nil, and the registry pointer,
// read AliasTables. An empty slice hits nothing.
func aliasTablesOf(p *[]AliasTable) []AliasTable {
	if p != nil && p != &aliasTables {
		out := make([]AliasTable, len(*p))
		copy(out, *p)
		return out
	}
	return AliasTables()
}
