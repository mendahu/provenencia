package effects

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"
)

// Path walks foreign keys declared in catalogmodel. A zero Path resolves nothing.
type Path struct {
	run   func(w *walk, in [][]byte, fromChange bool) ([][]byte, error)
	check func(start string) (string, error)
}

func (p Path) zero() bool { return p.run == nil }

type walk struct {
	q      querier
	ch     rowchange.Change
	ghosts ghostMap
	table  string
}

type querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func self() Path {
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			if fromChange {
				if len(w.ch.EntityID) != 16 {
					return nil, nil
				}
				return [][]byte{w.ch.EntityID}, nil
			}
			return in, nil
		},
		check: func(start string) (string, error) { return start, nil },
	}
}

func up(cols ...string) Path {
	if len(cols) == 0 {
		panic("effects: up() needs a column")
	}
	for _, c := range cols {
		ident(c)
	}
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			table := w.table
			ids := in
			rest := cols
			if fromChange {
				got, next, err := readChangeColumn(w, cols[0])
				if err != nil {
					return nil, err
				}
				ids = got
				table = next
				rest = cols[1:]
			}
			for _, col := range rest {
				next, err := fkTarget(table, col)
				if err != nil {
					return nil, err
				}
				ids, err = readRowsColumn(w, table, ids, col)
				if err != nil {
					return nil, err
				}
				table = next
			}
			return ids, nil
		},
		check: func(start string) (string, error) {
			table := start
			for _, col := range cols {
				next, err := fkTarget(table, col)
				if err != nil {
					return "", fmt.Errorf("up %s.%s: %w", table, col, err)
				}
				table = next
			}
			return table, nil
		},
	}
}

func field(col string) Path {
	ident(col)
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			if fromChange {
				ids, _, err := readChangeColumn(w, col)
				return ids, err
			}
			return readRowsColumn(w, w.table, in, col)
		},
		check: func(start string) (string, error) {
			next, err := fkTarget(start, col)
			if err != nil {
				return "", fmt.Errorf("field %s.%s: %w", start, col, err)
			}
			return next, nil
		},
	}
}

// from reads col and feeds those ids to next. The column is on the changed row
// when this step starts a path, and on the piped rows when it does not.
func from(col string, next Path) Path {
	ident(col)
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			var ids [][]byte
			var err error
			saved := w.table
			if fromChange {
				ids, _, err = readChangeColumn(w, col)
			} else {
				ids, err = readRowsColumn(w, w.table, in, col)
			}
			if err != nil {
				return nil, err
			}
			nextTable, err := fkTarget(saved, col)
			if err != nil {
				return nil, err
			}
			w.table = nextTable
			defer func() { w.table = saved }()
			return next.run(w, ids, false)
		},
		check: func(start string) (string, error) {
			nextTable, err := fkTarget(start, col)
			if err != nil {
				return "", fmt.Errorf("from %s.%s: %w", start, col, err)
			}
			if next.run == nil {
				return "", fmt.Errorf("from %s.%s: empty path", start, col)
			}
			return next.check(nextTable)
		},
	}
}

func inbound(table, col string) Path {
	ident(table)
	ident(col)
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			targets := in
			if fromChange {
				targets = [][]byte{w.ch.EntityID}
			}
			pk := pkOf(table)
			q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, pk, table, col)
			var out [][]byte
			for _, id := range targets {
				if len(id) != 16 {
					continue
				}
				got, err := queryIDs(w.q, q, id)
				if err != nil {
					return nil, err
				}
				out = append(out, got...)
			}
			return out, nil
		},
		check: func(start string) (string, error) {
			to, err := fkTarget(table, col)
			if err != nil {
				return "", fmt.Errorf("inbound %s.%s: %w", table, col, err)
			}
			if to != start {
				return "", fmt.Errorf("inbound %s.%s points at %s, want %s", table, col, to, start)
			}
			return table, nil
		},
	}
}

func across(table, fromCol, toCol string, statuses ...string) Path {
	ident(table)
	ident(fromCol)
	ident(toCol)
	for _, s := range statuses {
		if strings.ContainsAny(s, "'\\") {
			panic("effects: status " + s)
		}
	}
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			targets := in
			if fromChange {
				targets = [][]byte{w.ch.EntityID}
			}
			q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, toCol, table, fromCol)
			if len(statuses) > 0 {
				quoted := make([]string, len(statuses))
				for i, s := range statuses {
					quoted[i] = "'" + s + "'"
				}
				q += ` AND status IN (` + strings.Join(quoted, ", ") + `)`
			}
			var out [][]byte
			for _, id := range targets {
				if len(id) != 16 {
					continue
				}
				got, err := queryIDs(w.q, q, id)
				if err != nil {
					return nil, err
				}
				out = append(out, got...)
			}
			return out, nil
		},
		check: func(start string) (string, error) {
			to, err := fkTarget(table, fromCol)
			if err != nil {
				return "", fmt.Errorf("across %s.%s: %w", table, fromCol, err)
			}
			if to != start {
				return "", fmt.Errorf("across %s.%s points at %s, want %s", table, fromCol, to, start)
			}
			dest, err := fkTarget(table, toCol)
			if err != nil {
				return "", fmt.Errorf("across %s.%s: %w", table, toCol, err)
			}
			return dest, nil
		},
	}
}

func chain(steps ...Path) Path {
	if len(steps) == 0 {
		panic("effects: chain() needs a step")
	}
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			ids := in
			fc := fromChange
			saved := w.table
			defer func() { w.table = saved }()
			for _, step := range steps {
				got, err := step.run(w, ids, fc)
				if err != nil {
					return nil, err
				}
				out, err := step.check(w.table)
				if err != nil {
					return nil, err
				}
				w.table = out
				ids = got
				fc = false
			}
			return ids, nil
		},
		check: func(start string) (string, error) {
			table := start
			for _, step := range steps {
				next, err := step.check(table)
				if err != nil {
					return "", err
				}
				table = next
			}
			return table, nil
		},
	}
}

func union(steps ...Path) Path {
	if len(steps) == 0 {
		panic("effects: union() needs a step")
	}
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			var out [][]byte
			for _, step := range steps {
				got, err := step.run(w, in, fromChange)
				if err != nil {
					return nil, err
				}
				out = append(out, got...)
			}
			return out, nil
		},
		check: func(start string) (string, error) {
			out := ""
			for _, step := range steps {
				next, err := step.check(start)
				if err != nil {
					return "", err
				}
				out = next
			}
			return out, nil
		},
	}
}

func onField(field string, inner Path) Path {
	ident(field)
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			if _, ok := w.ch.Fields[field]; !ok {
				return nil, nil
			}
			return inner.run(w, in, fromChange)
		},
		check: func(start string) (string, error) { return inner.check(start) },
	}
}

// onStatus runs inner when the old or the new status is want. A diff that
// omits status did not change it, so the inner path does not run.
func onStatus(want string, inner Path) Path {
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			diff, ok := w.ch.Fields["status"]
			if !ok {
				return nil, nil
			}
			if asString(diff.Old) == want || asString(diff.New) == want {
				return inner.run(w, in, fromChange)
			}
			return nil, nil
		},
		check: func(start string) (string, error) { return inner.check(start) },
	}
}

func sqlPath(query string) Path {
	return Path{
		run: func(w *walk, in [][]byte, fromChange bool) ([][]byte, error) {
			ids := in
			if fromChange {
				ids = [][]byte{w.ch.EntityID}
			}
			var out [][]byte
			for _, id := range ids {
				if len(id) != 16 {
					continue
				}
				got, err := queryIDs(w.q, query, id)
				if err != nil {
					return nil, err
				}
				out = append(out, got...)
			}
			return out, nil
		},
		check: func(string) (string, error) { return "", nil },
	}
}

func fkTarget(from, col string) (string, error) {
	for _, fk := range catalogmodel.FKs {
		if fk.From == from && fk.Column == col {
			return fk.To, nil
		}
	}
	return "", fmt.Errorf("no foreign key %s.%s", from, col)
}

func pkOf(table string) string {
	for _, t := range catalogmodel.Tables {
		if t.Name == table && t.PK != "" {
			return t.PK
		}
	}
	return "id"
}

func ident(name string) string {
	if name == "" {
		panic("effects: empty identifier")
	}
	for _, r := range name {
		if r != '_' && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			panic("effects: bad identifier " + name)
		}
	}
	return name
}

func readChangeColumn(w *walk, col string) ([][]byte, string, error) {
	next, err := fkTarget(w.table, col)
	if err != nil {
		return nil, "", err
	}
	if diff, ok := w.ch.Fields[col]; ok {
		return uuidPair(diff.Old, diff.New), next, nil
	}
	if w.ch.Action == rowchange.ActionDelete {
		if id := w.ghosts.id(w.ch.EntityType, w.ch.EntityID, col); len(id) == 16 {
			return [][]byte{id}, next, nil
		}
		return nil, next, nil
	}
	ids, err := readRowsColumn(w, w.table, [][]byte{w.ch.EntityID}, col)
	return ids, next, err
}

func readRowsColumn(w *walk, table string, ids [][]byte, col string) ([][]byte, error) {
	pk := pkOf(table)
	q := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, ident(col), ident(table), ident(pk))
	entity := entityTable[table]
	var out [][]byte
	for _, id := range ids {
		if len(id) != 16 {
			continue
		}
		var raw []byte
		err := w.q.QueryRow(q, id).Scan(&raw)
		if err == sql.ErrNoRows {
			raw = nil
			err = nil
		}
		if err != nil {
			return nil, err
		}
		if len(raw) != 16 && entity != "" {
			raw = w.ghosts.id(entity, id, col)
		}
		if len(raw) == 16 {
			out = append(out, append([]byte(nil), raw...))
		}
	}
	return out, nil
}

func queryIDs(q querier, query string, args ...any) ([][]byte, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if len(id) == 16 {
			out = append(out, append([]byte(nil), id...))
		}
	}
	return out, rows.Err()
}

func uuidPair(old, new any) [][]byte {
	var out [][]byte
	for _, v := range []any{new, old} {
		id := uuidValue(v)
		if len(id) != 16 {
			continue
		}
		if len(out) == 1 && string(out[0]) == string(id) {
			continue
		}
		out = append(out, id)
	}
	return out
}

func uuidValue(v any) []byte {
	switch x := v.(type) {
	case string:
		u, err := uuid.Parse(x)
		if err != nil {
			return nil
		}
		return u[:]
	case []byte:
		if len(x) != 16 {
			return nil
		}
		return append([]byte(nil), x...)
	default:
		return nil
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
