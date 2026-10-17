package namealiases

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// NicknameAdd is how many product nickname rows were new.
type NicknameAdd struct {
	Sets   int // equivalence sets that were not in the catalog
	Tokens int // tokens inserted, including ones added onto an existing set
}

// AddProductNicknames copies the product nickname list into this catalog.
// A set already stored in full is left as it is. A stored set that is a
// smaller version of exactly one product set gains the missing tokens.
// A stored set that could belong to more than one product set is left
// alone, and those product sets are added beside it, so two formal names
// that share a nickname stay separate. Other alias tables are not changed.
// Call Invalidate after this if the catalog stays open.
func AddProductNicknames(c *database.Catalog) (NicknameAdd, error) {
	seed, ok := productNicknames()
	if !ok {
		return NicknameAdd{}, nil
	}
	db, err := c.DB()
	if err != nil {
		return NicknameAdd{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return NicknameAdd{}, err
	}
	defer func() { _ = tx.Rollback() }()

	added, err := addProductNicknames(tx, seed)
	if err != nil {
		return NicknameAdd{}, err
	}
	if err := tx.Commit(); err != nil {
		return NicknameAdd{}, err
	}
	return added, nil
}

func productNicknames() (valuecompare.AliasTable, bool) {
	for _, table := range AliasTables() {
		if table.Key == "nicknames" && table.Origin == vocabkey.OriginProvenencia && table.Kind == valuecompare.AliasNickname {
			return table, true
		}
	}
	return valuecompare.AliasTable{}, false
}

func addProductNicknames(tx *sql.Tx, seed valuecompare.AliasTable) (NicknameAdd, error) {
	var id []byte
	err := tx.QueryRow(`SELECT id FROM name_alias_tables WHERE key = ? AND origin = ?`,
		seed.Key, seed.Origin).Scan(&id)
	if err == sql.ErrNoRows {
		return insertNicknameTable(tx, seed)
	}
	if err != nil {
		return NicknameAdd{}, err
	}

	stored, err := loadMemberSets(tx, id)
	if err != nil {
		return NicknameAdd{}, err
	}
	seeds := make([]map[string]bool, len(seed.Sets))
	for i, set := range seed.Sets {
		seeds[i] = tokenSet(set)
	}
	for i := range stored {
		stored[i].supersets = supersetCount(stored[i].tokens, seeds)
	}

	var added NicknameAdd
	next := 0
	for _, set := range stored {
		if set.idx >= next {
			next = set.idx + 1
		}
	}
	for i, want := range seed.Sets {
		if setContains(stored, seeds[i]) {
			continue
		}
		if idx, ok := soleSmallerSet(stored, seeds[i]); ok {
			n, err := insertMissing(tx, id, idx, stored, want)
			if err != nil {
				return NicknameAdd{}, err
			}
			added.Tokens += n
			continue
		}
		if _, err := insertSet(tx, id, next, want); err != nil {
			return NicknameAdd{}, err
		}
		stored = append(stored, memberSet{idx: next, tokens: seeds[i], supersets: 1})
		next++
		added.Sets++
		added.Tokens += len(want)
	}
	return added, nil
}

type memberSet struct {
	idx       int
	tokens    map[string]bool
	supersets int
}

func loadMemberSets(tx *sql.Tx, tableID []byte) ([]memberSet, error) {
	rows, err := tx.Query(`SELECT set_idx, token FROM name_alias_members WHERE table_id = ? ORDER BY set_idx, token`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byIdx := map[int]map[string]bool{}
	var order []int
	for rows.Next() {
		var idx int
		var token string
		if err := rows.Scan(&idx, &token); err != nil {
			return nil, err
		}
		if byIdx[idx] == nil {
			byIdx[idx] = map[string]bool{}
			order = append(order, idx)
		}
		byIdx[idx][token] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]memberSet, len(order))
	for i, idx := range order {
		out[i] = memberSet{idx: idx, tokens: byIdx[idx]}
	}
	return out, nil
}

func tokenSet(tokens []string) map[string]bool {
	out := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if token != "" {
			out[token] = true
		}
	}
	return out
}

func supersetCount(have map[string]bool, seeds []map[string]bool) int {
	n := 0
	for _, seed := range seeds {
		if len(have) > 0 && isSubset(have, seed) {
			n++
		}
	}
	return n
}

func isSubset(small, big map[string]bool) bool {
	for token := range small {
		if !big[token] {
			return false
		}
	}
	return true
}

func setContains(sets []memberSet, want map[string]bool) bool {
	for _, set := range sets {
		if isSubset(want, set.tokens) {
			return true
		}
	}
	return false
}

// soleSmallerSet is the one stored set that is a smaller copy of want and
// of no other product set. Two product sets that share a token must not
// grow into each other.
func soleSmallerSet(sets []memberSet, want map[string]bool) (int, bool) {
	found := -1
	for _, set := range sets {
		if set.supersets != 1 || !isSubset(set.tokens, want) || isSubset(want, set.tokens) {
			continue
		}
		if found >= 0 {
			return 0, false
		}
		found = set.idx
	}
	return found, found >= 0
}

func insertMissing(tx *sql.Tx, tableID []byte, idx int, sets []memberSet, tokens []string) (int, error) {
	var have map[string]bool
	for i := range sets {
		if sets[i].idx == idx {
			have = sets[i].tokens
			break
		}
	}
	n := 0
	for _, token := range tokens {
		if token == "" || have[token] {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO name_alias_members (table_id, set_idx, token) VALUES (?, ?, ?)`,
			tableID, idx, token); err != nil {
			return 0, err
		}
		have[token] = true
		n++
	}
	return n, nil
}

func insertSet(tx *sql.Tx, tableID []byte, idx int, tokens []string) (int, error) {
	n := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO name_alias_members (table_id, set_idx, token) VALUES (?, ?, ?)`,
			tableID, idx, token); err != nil {
			return 0, err
		}
		n++
	}
	return n, nil
}

func insertNicknameTable(tx *sql.Tx, seed valuecompare.AliasTable) (NicknameAdd, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return NicknameAdd{}, err
	}
	if _, err := tx.Exec(`INSERT INTO name_alias_tables (id, key, origin, label, kind) VALUES (?, ?, ?, ?, ?)`,
		id[:], seed.Key, seed.Origin, seed.Label, string(seed.Kind)); err != nil {
		return NicknameAdd{}, err
	}
	var added NicknameAdd
	for i, set := range seed.Sets {
		n, err := insertSet(tx, id[:], i, set)
		if err != nil {
			return NicknameAdd{}, err
		}
		added.Sets++
		added.Tokens += n
	}
	for _, role := range seed.Roles {
		if _, err := tx.Exec(`INSERT INTO name_alias_roles (table_id, role) VALUES (?, ?)`,
			id[:], string(role)); err != nil {
			return NicknameAdd{}, err
		}
	}
	return added, nil
}
