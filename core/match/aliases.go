package match

import "sync"

// aliasTables is the product seed plus anything Register appended.
// Install writes it. Scoring does not read it.
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

// AliasTables returns the seed Install writes. Scoring does not read this.
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

// aliasTablesOf is the list on the comparer. Nil and empty both hit nothing.
func aliasTablesOf(p *[]AliasTable) []AliasTable {
	if p == nil {
		return nil
	}
	out := make([]AliasTable, len(*p))
	copy(out, *p)
	return out
}
