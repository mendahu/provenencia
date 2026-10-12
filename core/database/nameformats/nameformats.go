// Package nameformats stores name format profiles: display order, matching
// roles, and role weights. A name value does not point at a profile. Install
// runs once when a catalog is created. Load reads the project default, or
// the built-in Western pattern when that row is missing.
package nameformats

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/match"
)

const (
	OriginProvenencia = "provenencia"

	StyleNatural = "natural"
	StyleSorted  = "sorted"

	settingDefaultNameFormatID = "default_name_format_id"
)

// Querier is *sql.DB or *sql.Tx.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// PartOrder is one part type in a display style, plus the separator that
// follows that group. Empty TrailingSeparator is a normal space.
type PartOrder struct {
	Type              string
	TrailingSeparator string
}

// Profile is one name_format_profiles row and the parts, roles, and weights
// that belong to it.
type Profile struct {
	ID          []byte
	Key         string
	Origin      string
	Label       string
	Description string
	Pattern     match.NamePattern
	Natural     []PartOrder
	Sorted      []PartOrder
}

// Order is the part list for style. An unknown style has no rows.
func (p Profile) Order(style string) []PartOrder {
	switch style {
	case StyleSorted:
		return p.Sorted
	default:
		return p.Natural
	}
}

// Format joins name's parts in this profile's order for style. A part type
// the style does not list follows, in stored order. No parts returns form.
// form is never rewritten.
func (p Profile) Format(name namevalues.Value, style string) string {
	return Format(name, p.Order(style))
}

// Format joins name's parts by order. An empty order keeps stored order,
// spaces between parts. No parts returns form.
func Format(name namevalues.Value, order []PartOrder) string {
	if len(name.Parts) == 0 {
		return strings.TrimSpace(name.Form)
	}
	byType := map[string][]namevalues.Part{}
	var stored []string
	seen := map[string]bool{}
	for _, part := range name.Parts {
		byType[part.Type] = append(byType[part.Type], part)
		if !seen[part.Type] {
			seen[part.Type] = true
			stored = append(stored, part.Type)
		}
	}
	listed := map[string]bool{}
	type group struct {
		text  string
		trail string
	}
	var groups []group
	add := func(typ, trail string) {
		parts := byType[typ]
		text := joinPartValues(parts)
		if text == "" {
			return
		}
		groups = append(groups, group{text: text, trail: trail})
	}
	for _, row := range order {
		listed[row.Type] = true
		add(row.Type, row.TrailingSeparator)
	}
	for _, typ := range stored {
		if listed[typ] {
			continue
		}
		add(typ, "")
	}
	if len(groups) == 0 {
		return ""
	}
	var b strings.Builder
	for i, g := range groups {
		if i > 0 {
			sep := groups[i-1].trail
			if sep == "" {
				sep = " "
			}
			b.WriteString(sep)
		}
		b.WriteString(g.text)
	}
	return b.String()
}

func joinPartValues(parts []namevalues.Part) string {
	var words []string
	for _, part := range parts {
		word := strings.TrimSpace(part.Value)
		if word != "" {
			words = append(words, word)
		}
	}
	return strings.Join(words, " ")
}

// Loaded is the catalog's default name format. Builtin is the in-code
// Western pattern, used when the catalog has no default row.
type Loaded struct {
	Key     string
	Profile Profile
	Set     match.NamePatternSet
	Builtin bool
}

// Comparer scores names with the loaded patterns. A builtin load is
// DefaultNames, so a catalog with no default scores as it does today.
func (l Loaded) Comparer() match.NameComparer {
	if l.Builtin || l.Set == nil {
		return match.DefaultNames
	}
	c := match.DefaultNames
	c.Patterns = l.Set
	c.Pattern = match.Set(l.Key)
	return c
}

// Load reads the project default. A missing or empty setting returns the
// built-in Western pattern.
func Load(q Querier) (Loaded, error) {
	var raw sql.NullString
	err := q.QueryRow(`SELECT value_text FROM project_settings WHERE key = ?`, settingDefaultNameFormatID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) || !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return builtinLoaded(), nil
	}
	if err != nil {
		return Loaded{}, err
	}
	id, err := hex.DecodeString(strings.TrimSpace(raw.String))
	if err != nil || len(id) != 16 {
		return Loaded{}, errors.New("nameformats: default_name_format_id")
	}
	p, err := LoadID(q, id)
	if errors.Is(err, sql.ErrNoRows) {
		return builtinLoaded(), nil
	}
	if err != nil {
		return Loaded{}, err
	}
	return Loaded{
		Key:     p.Key,
		Profile: p,
		Set:     match.NamePatternSet{p.Key: p.Pattern},
	}, nil
}

// LoadID reads one profile. sql.ErrNoRows when the id is absent.
func LoadID(q Querier, id []byte) (Profile, error) {
	var p Profile
	var description sql.NullString
	err := q.QueryRow(`SELECT id, key, origin, label, description FROM name_format_profiles WHERE id = ?`, id).
		Scan(&p.ID, &p.Key, &p.Origin, &p.Label, &description)
	if err != nil {
		return Profile{}, err
	}
	p.Description = description.String
	p.Pattern = match.NamePattern{Key: p.Key, PartRoles: map[string]match.NameRole{}}
	rows, err := q.Query(`SELECT part_type, role FROM name_format_profile_roles WHERE profile_id = ?`, id)
	if err != nil {
		return Profile{}, err
	}
	for rows.Next() {
		var partType, role string
		if err := rows.Scan(&partType, &role); err != nil {
			rows.Close()
			return Profile{}, err
		}
		p.Pattern.PartRoles[partType] = match.NameRole(role)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Profile{}, err
	}
	rows.Close()

	rows, err = q.Query(`SELECT slot, weight FROM name_format_profile_weights WHERE profile_id = ?`, id)
	if err != nil {
		return Profile{}, err
	}
	for rows.Next() {
		var slot string
		var weight float64
		if err := rows.Scan(&slot, &weight); err != nil {
			rows.Close()
			return Profile{}, err
		}
		switch slot {
		case "family":
			p.Pattern.Weights.Family = weight
		case "first_given":
			p.Pattern.Weights.FirstGiven = weight
		case "other_given":
			p.Pattern.Weights.OtherGiven = weight
		case "nick":
			p.Pattern.Weights.Nick = weight
		case "untyped":
			p.Pattern.Weights.Untyped = weight
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Profile{}, err
	}
	rows.Close()

	rows, err = q.Query(`SELECT style, part_type, trailing_separator FROM name_format_profile_parts
		WHERE profile_id = ? ORDER BY style, idx`, id)
	if err != nil {
		return Profile{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var style, partType, trail string
		if err := rows.Scan(&style, &partType, &trail); err != nil {
			return Profile{}, err
		}
		row := PartOrder{Type: partType, TrailingSeparator: trail}
		switch style {
		case StyleSorted:
			p.Sorted = append(p.Sorted, row)
		case StyleNatural:
			p.Natural = append(p.Natural, row)
		}
	}
	return p, rows.Err()
}

func builtinLoaded() Loaded {
	return Loaded{Key: match.WesternPattern, Builtin: true}
}

type partSeed struct {
	style, partType, trail string
}

// westernParts is seeded-vocabulary §4.3. The comma is the surname row's
// trailing separator in the sorted style.
var westernParts = []partSeed{
	{StyleNatural, namevalues.PartTypePrefix, ""},
	{StyleNatural, namevalues.PartTypeGiven, ""},
	{StyleNatural, namevalues.PartTypeNick, ""},
	{StyleNatural, namevalues.PartTypeSurnamePrefix, ""},
	{StyleNatural, namevalues.PartTypeSurname, ""},
	{StyleNatural, namevalues.PartTypeSuffix, ""},
	{StyleSorted, namevalues.PartTypeSurnamePrefix, ""},
	{StyleSorted, namevalues.PartTypeSurname, ", "},
	{StyleSorted, namevalues.PartTypePrefix, ""},
	{StyleSorted, namevalues.PartTypeGiven, ""},
	{StyleSorted, namevalues.PartTypeNick, ""},
	{StyleSorted, namevalues.PartTypeSuffix, ""},
}

// Install writes the western profile and points the project default at it.
// Call only from catalog create. A second call inserts nothing new for the
// same key and leaves the existing default in place when one is already set.
func Install(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO name_format_profiles (id, key, origin, label, description)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(key, origin) DO NOTHING`,
		id[:], match.WesternPattern, OriginProvenencia, "Western", "GEDCOM-aligned spoken/display order")
	if err != nil {
		return err
	}
	var stored []byte
	if err := tx.QueryRow(`SELECT id FROM name_format_profiles WHERE key = ? AND origin = ?`,
		match.WesternPattern, OriginProvenencia).Scan(&stored); err != nil {
		return err
	}
	idx := map[string]int{}
	for _, row := range westernParts {
		i := idx[row.style]
		idx[row.style] = i + 1
		if _, err := tx.Exec(`INSERT INTO name_format_profile_parts
			(profile_id, style, idx, part_type, trailing_separator)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(profile_id, style, idx) DO NOTHING`,
			stored, row.style, i, row.partType, row.trail); err != nil {
			return err
		}
	}
	for _, pt := range namevalues.PartTypes() {
		role, ok := match.WesternNamePattern.PartRoles[pt.Key]
		if !ok {
			return errors.New("nameformats: " + pt.Key + " has no western role")
		}
		if _, err := tx.Exec(`INSERT INTO name_format_profile_roles (profile_id, part_type, role)
			VALUES (?, ?, ?)
			ON CONFLICT(profile_id, part_type) DO NOTHING`,
			stored, pt.Key, string(role)); err != nil {
			return err
		}
	}
	weights := match.WesternNamePattern.Weights
	for _, slot := range []struct {
		name   string
		weight float64
	}{
		{"family", weights.Family},
		{"first_given", weights.FirstGiven},
		{"other_given", weights.OtherGiven},
		{"nick", weights.Nick},
		{"untyped", weights.Untyped},
	} {
		if _, err := tx.Exec(`INSERT INTO name_format_profile_weights (profile_id, slot, weight)
			VALUES (?, ?, ?)
			ON CONFLICT(profile_id, slot) DO NOTHING`,
			stored, slot.name, slot.weight); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`INSERT INTO project_settings (key, value_text) VALUES (?, ?)
		ON CONFLICT(key) DO NOTHING`,
		settingDefaultNameFormatID, hex.EncodeToString(stored)); err != nil {
		return err
	}
	return tx.Commit()
}
