package search

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/mendahu/provenencia/core/ref"
)

var refPrefixForm = regexp.MustCompile(`^[A-Z]{3}-[0-9A-HJKMNP-TV-Z]{0,4}$`)

// classifyRefQuery returns an uppercased lookup key when the raw query looks
// like a catalog ref or ref prefix (e.g. SRC-ZZ9K2 or SRC-ZZ).
func classifyRefQuery(text string) (key string, exact bool) {
	s := strings.ToUpper(strings.TrimSpace(text))
	if s == "" {
		return "", false
	}
	if ref.Valid(s) {
		return s, true
	}
	if refPrefixForm.MatchString(s) {
		parts := strings.SplitN(s, "-", 2)
		if len(parts) == 2 && len(parts[1]) >= 1 {
			return s, false
		}
	}
	return "", false
}

func lookUpRefDocs(ctx context.Context, db *sql.DB, key string, exact bool, limit int) ([]docRow, error) {
	if err := errIfCancelled(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = DefaultHitLimit
	}
	var (
		rows *sql.Rows
		err  error
	)
	if exact {
		rows, err = db.QueryContext(ctx, `
			SELECT kind, entity_id, display_ref, display_title, display_subtitle,
				display_icon_key, display_thumbnail_rel_path,
				title, ref, secondary, body
			FROM catalog_search_docs
			WHERE upper(ref) = ?
			LIMIT ?
		`, key, limit)
	} else {
		like := escapeLike(key) + "%"
		rows, err = db.QueryContext(ctx, `
			SELECT kind, entity_id, display_ref, display_title, display_subtitle,
				display_icon_key, display_thumbnail_rel_path,
				title, ref, secondary, body
			FROM catalog_search_docs
			WHERE upper(ref) LIKE ? ESCAPE '\'
			LIMIT ?
		`, like, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []docRow
	for rows.Next() {
		if err := errIfCancelled(ctx); err != nil {
			return nil, err
		}
		var d docRow
		if err := rows.Scan(
			&d.kind, &d.entityID, &d.displayRef, &d.displayTitle, &d.displaySubtitle,
			&d.displayIconKey, &d.displayThumbnailRelPath,
			&d.title, &d.ref, &d.secondary, &d.body,
		); err != nil {
			return nil, err
		}
		if exact {
			d.refMatch = refMatchExact
		} else {
			d.refMatch = refMatchPrefix
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func escapeLike(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

type refMatchTier int

const (
	refMatchNone refMatchTier = iota
	refMatchPrefix
	refMatchExact
)

func refBoostFor(tier refMatchTier) float64 {
	switch tier {
	case refMatchExact:
		return RefBoostWeights.Exact
	case refMatchPrefix:
		return RefBoostWeights.Prefix
	default:
		return RefBoostWeights.None
	}
}

func betterRefMatch(a, b refMatchTier) refMatchTier {
	if a > b {
		return a
	}
	return b
}
