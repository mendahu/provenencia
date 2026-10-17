package nameformats

import (
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/namevalue"
)

// seedPart is one display-order row. Idx is the position within Style.
type seedPart struct {
	Style, PartType, TrailingSeparator string
}

type seedRole struct {
	PartType string
	Role     match.NameRole
}

type seedWeight struct {
	Slot   string
	Weight float64
}

// seedProfile is one create-time name format. Adding a profile is another
// entry here; Install writes the rows and does not branch on Key.
type seedProfile struct {
	Key, Label, Description string
	Default                 bool
	Parts                   []seedPart
	Roles                   []seedRole
	Weights                 []seedWeight
}

// seedProfiles is the provenencia name-format registry (create-time only).
// Part order is seeded-vocabulary §4.3. The comma is the surname row's
// trailing separator in the sorted style.
var seedProfiles = []seedProfile{
	{
		Key:         match.WesternPattern,
		Label:       "Western",
		Description: "GEDCOM-aligned spoken/display order",
		Default:     true,
		Parts: []seedPart{
			{StyleNatural, namevalue.PartTypePrefix, ""},
			{StyleNatural, namevalue.PartTypeGiven, ""},
			{StyleNatural, namevalue.PartTypeNick, ""},
			{StyleNatural, namevalue.PartTypeSurnamePrefix, ""},
			{StyleNatural, namevalue.PartTypeSurname, ""},
			{StyleNatural, namevalue.PartTypeSuffix, ""},
			{StyleSorted, namevalue.PartTypeSurnamePrefix, ""},
			{StyleSorted, namevalue.PartTypeSurname, ", "},
			{StyleSorted, namevalue.PartTypePrefix, ""},
			{StyleSorted, namevalue.PartTypeGiven, ""},
			{StyleSorted, namevalue.PartTypeNick, ""},
			{StyleSorted, namevalue.PartTypeSuffix, ""},
		},
		Roles: []seedRole{
			{namevalue.PartTypeSurname, match.RoleFamily},
			{namevalue.PartTypeGiven, match.RoleGiven},
			{namevalue.PartTypeNick, match.RoleNick},
			{namevalue.PartTypeSuffix, match.RoleGeneration},
			{namevalue.PartTypePrefix, match.RoleIgnored},
			{namevalue.PartTypeSurnamePrefix, match.RoleIgnored},
			{namevalue.PartTypeUndetermined, match.RoleUntyped},
		},
		Weights: []seedWeight{
			{"family", 1.5},
			{"first_given", 1},
			{"other_given", 0.5},
			{"nick", 0.3},
			{"untyped", 1},
		},
	},
}

// Install writes the name-format registry and points the project default at
// the profile marked Default. Call only from catalog create. A second call
// inserts nothing new for the same key and leaves an existing default in place.
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

	var defaultID []byte
	defaults := 0
	for _, profile := range seedProfiles {
		if profile.Default {
			for _, pt := range namevalue.PartTypes() {
				if !profile.hasRole(pt.Key) {
					return errors.New("nameformats: default profile has no role for " + pt.Key)
				}
			}
		}
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO name_format_profiles (id, key, origin, label, description)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(key, origin) DO NOTHING`,
			id[:], profile.Key, OriginProvenencia, profile.Label, profile.Description); err != nil {
			return err
		}
		var stored []byte
		if err := tx.QueryRow(`SELECT id FROM name_format_profiles WHERE key = ? AND origin = ?`,
			profile.Key, OriginProvenencia).Scan(&stored); err != nil {
			return err
		}
		idx := map[string]int{}
		for _, row := range profile.Parts {
			i := idx[row.Style]
			idx[row.Style] = i + 1
			if _, err := tx.Exec(`INSERT INTO name_format_profile_parts
				(profile_id, style, idx, part_type, trailing_separator)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(profile_id, style, idx) DO NOTHING`,
				stored, row.Style, i, row.PartType, row.TrailingSeparator); err != nil {
				return err
			}
		}
		for _, row := range profile.Roles {
			if _, err := tx.Exec(`INSERT INTO name_format_profile_roles (profile_id, part_type, role)
				VALUES (?, ?, ?)
				ON CONFLICT(profile_id, part_type) DO NOTHING`,
				stored, row.PartType, string(row.Role)); err != nil {
				return err
			}
		}
		for _, row := range profile.Weights {
			if _, err := tx.Exec(`INSERT INTO name_format_profile_weights (profile_id, slot, weight)
				VALUES (?, ?, ?)
				ON CONFLICT(profile_id, slot) DO NOTHING`,
				stored, row.Slot, row.Weight); err != nil {
				return err
			}
		}
		if profile.Default {
			defaults++
			defaultID = stored
		}
	}
	if defaults != 1 || len(defaultID) != 16 {
		return errors.New("nameformats: seed needs one default profile")
	}
	if _, err := tx.Exec(`INSERT INTO project_settings (key, value_text) VALUES (?, ?)
		ON CONFLICT(key) DO NOTHING`,
		settingDefaultNameFormatID, hex.EncodeToString(defaultID)); err != nil {
		return err
	}
	return tx.Commit()
}

func (p seedProfile) hasRole(partType string) bool {
	for _, row := range p.Roles {
		if row.PartType == partType {
			return true
		}
	}
	return false
}
