package deleteimpact

import (
	"database/sql"
	"fmt"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/mendahu/provenencia/core/apperr"
)

// Facet release: research rows an official Delete removes explicitly — and
// audits — before its parent DELETE. Each facetRelease is the single
// definition of one such relationship: the honesty test checks it against the
// FK, Impact reads it for Report.Cascades when Named, and ReleaseFacets
// mutates through it. The schema's ON DELETE CASCADE is only a backstop:
// ReleaseFacets fails the delete if any row it owns is still there afterwards.

// ErrReleaseIncomplete means a facet release left rows its backstop would
// silently delete. It is a programming error, never a researcher outcome.
var ErrReleaseIncomplete = apperr.New(apperr.CodeDeleteImpactReleaseIncomplete, apperr.KindInternal)

// released is what releaseFacets removed for one parent.
type released struct {
	// Changes are audit rows for every removed facet, ready to append ahead of
	// the parent's own delete change in the same revision. Handles come from
	// effects.Handles on these changes, including through ghosts when a claim
	// row is already gone.
	Changes []rowchange.Change
	// Objects are the stored bytes of Files the release deleted. The exported
	// release schedules their purge with files.PurgeAfterCommit.
	Objects []files.Object
}

func (r *released) merge(o released) {
	r.Changes = append(r.Changes, o.Changes...)
	r.Objects = append(r.Objects, o.Objects...)
}

type facetRelease struct {
	Parent catalogmodel.Kind
	// Via is the FK this release covers ("child_table.fk_col"). A CASCADE FK
	// registered Audited must have exactly one release.
	Via string
	// Named facets appear in Report.Cascades: Child is the listed kind, and
	// Count / List must select exactly what Release removes.
	Named bool
	Child catalogmodel.Kind
	Count func(tx *sql.Tx, parentID []byte) (int, error)
	List  func(tx *sql.Tx, parentID []byte, limit int) ([]probeRow, error)
	// Remaining counts rows still pointing at the parent after Release; it
	// must be zero. Empty only for non-CASCADE releases (connection facets),
	// whose leftovers Impact has already refused.
	Remaining string
	Release   func(tx *sql.Tx, parentID []byte) (released, error)
}

// ReleaseFacets removes and audits every registered facet of one parent and
// schedules the purge of any File it released, then returns the audit changes
// to prepend to the delete's own. It takes the write's Tx so no caller holds
// the objects. See releaseFacets.
func ReleaseFacets(tx *database.Tx, kind catalogmodel.Kind, id []byte) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	r, err := releaseFacets(tx.Tx, kind, id)
	if err != nil {
		return nil, err
	}
	files.PurgeAfterCommit(tx, r.Objects)
	return r.Changes, nil
}

// releaseFacets removes and audits every registered facet of one parent, in
// registry order. Call it inside the delete transaction, after Refuse and
// before the parent DELETE. Releases compose: a facet row that is itself a
// parent (a claim, a connection Observation) has its own facets released first.
func releaseFacets(tx *sql.Tx, kind catalogmodel.Kind, id []byte) (released, error) {
	if tx == nil || len(id) != 16 {
		return released{}, ErrInvalid
	}
	var out released
	for _, f := range facetReleasesFor(kind) {
		r, err := f.Release(tx, id)
		if err != nil {
			return released{}, err
		}
		out.merge(r)
		if f.Remaining == "" {
			continue
		}
		var n int
		if err := tx.QueryRow(f.Remaining, id).Scan(&n); err != nil {
			return released{}, err
		}
		if n != 0 {
			return released{}, fmt.Errorf("%w: %s left %d", ErrReleaseIncomplete, f.Via, n)
		}
	}
	return out, nil
}

func facetReleasesFor(kind catalogmodel.Kind) []facetRelease {
	var out []facetRelease
	for _, f := range facetReleases() {
		if f.Parent == kind {
			out = append(out, f)
		}
	}
	return out
}

// cascadesFor is the Named facets of kind as non-blocking Impact edges.
func cascadesFor(kind catalogmodel.Kind) []inboundEdge {
	var out []inboundEdge
	for _, f := range facetReleasesFor(kind) {
		if !f.Named {
			continue
		}
		out = append(out, inboundEdge{
			Parent: f.Parent, Via: f.Via, Child: f.Child, Bucket: catalogmodel.BucketFacet,
			Count: f.Count, List: f.List,
		})
	}
	return out
}

// facetReleaseList is built in init: releases recurse through ReleaseFacets,
// which a package-level initializer cannot reference (initialization cycle).
var facetReleaseList []facetRelease

func init() { facetReleaseList = buildFacetReleases() }

func facetReleases() []facetRelease { return facetReleaseList }

func buildFacetReleases() []facetRelease {
	return []facetRelease{
		// --- Conclusion ---------------------------------------------------
		// A Subject leaves every handle it has a claim on (any status); the
		// handles stay. Claims' own pins go first.
		{
			Parent: catalogmodel.KindSubject, Via: "identity_claims.subject_id",
			Named: true, Child: catalogmodel.KindCanonicalEntity,
			Count: countSQLFn(`SELECT COUNT(DISTINCT entity_id) FROM identity_claims WHERE subject_id = ?`),
			List: listSQLFn(`SELECT DISTINCT e.id, e.ref FROM identity_claims ic
				JOIN canonical_entities e ON e.id = ic.entity_id
				WHERE ic.subject_id = ?
				ORDER BY e.ref COLLATE NOCASE LIMIT ?`),
			Remaining: `SELECT COUNT(*) FROM identity_claims WHERE subject_id = ?`,
			Release:   releaseSubjectClaims,
		},
		// A pinned Observation leaves the exhibit of every claim that pinned
		// it; those claims stay, weaker (model §5.2 review).
		{
			Parent: catalogmodel.KindObservation, Via: "identity_claim_evidence.observation_id",
			Named: true, Child: catalogmodel.KindCanonicalEntity,
			Count: countSQLFn(`SELECT COUNT(DISTINCT ic.entity_id) FROM identity_claim_evidence ev
				JOIN identity_claims ic ON ic.id = ev.identity_claim_id
				WHERE ev.observation_id = ?`),
			List: listSQLFn(`SELECT DISTINCT e.id, e.ref FROM identity_claim_evidence ev
				JOIN identity_claims ic ON ic.id = ev.identity_claim_id
				JOIN canonical_entities e ON e.id = ic.entity_id
				WHERE ev.observation_id = ?
				ORDER BY e.ref COLLATE NOCASE LIMIT ?`),
			Remaining: `SELECT COUNT(*) FROM identity_claim_evidence WHERE observation_id = ?`,
			Release: pinRelease(`SELECT ev.identity_claim_id, ev.observation_id
				FROM identity_claim_evidence ev
				JOIN identity_claims ic ON ic.id = ev.identity_claim_id
				WHERE ev.observation_id = ? ORDER BY ev.identity_claim_id`),
		},
		{
			Parent: catalogmodel.KindIdentityClaim, Via: "identity_claim_evidence.identity_claim_id",
			Remaining: `SELECT COUNT(*) FROM identity_claim_evidence WHERE identity_claim_id = ?`,
			Release: pinRelease(`SELECT ev.identity_claim_id, ev.observation_id
				FROM identity_claim_evidence ev
				JOIN identity_claims ic ON ic.id = ev.identity_claim_id
				WHERE ev.identity_claim_id = ? ORDER BY ev.observation_id`),
		},

		// --- Interpretation -------------------------------------------------
		// Edge + disambiguation Observations on a bridge Subject (not a CASCADE:
		// Impact refuses any other Observation first). Each is erased whole.
		{
			Parent: catalogmodel.KindSubject, Via: "observations.subject_id",
			Release: releaseConnectionFacets,
		},
		rowFacet(catalogmodel.KindObservation, "observation_notes", "observation_id", "observation_note", "id",
			col{"id", colUUID}, col{"observation_id", colUUID}, col{"body", colText}),
		rowFacet(catalogmodel.KindCitation, "citation_notes", "citation_id", "citation_note", "id",
			col{"id", colUUID}, col{"citation_id", colUUID}, col{"body", colText}),

		// --- Source ---------------------------------------------------------
		rowFacet(catalogmodel.KindSource, "source_notes", "source_id", "source_note", "id",
			col{"id", colUUID}, col{"source_id", colUUID}, col{"body", colText}),
		rowFacet(catalogmodel.KindSource, "source_metadata", "source_id", "source_metadata", "id",
			col{"id", colUUID}, col{"source_id", colUUID}, col{"field_id", colUUID}, col{"value_text", colText}),
		rowFacet(catalogmodel.KindSource, "source_credibility_assessments", "source_id", "source_credibility_assessment", "id",
			col{"id", colUUID}, col{"source_id", colUUID}, col{"credibility_grade_id", colUUID}, col{"argument", colText}),
		// Layout rows are audited under field_id (sourcemetadata precedent).
		rowFacet(catalogmodel.KindSource, "source_metadata_layout", "source_id", "source_metadata_layout", "field_id",
			col{"source_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),
		rowFacet(catalogmodel.KindMetadataField, "source_metadata_layout", "field_id", "source_metadata_layout", "field_id",
			col{"source_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),
	}
}

// --- Conclusion releases ----------------------------------------------------

func releaseSubjectClaims(tx *sql.Tx, subjectID []byte) (released, error) {
	rows, err := tx.Query(`SELECT id FROM identity_claims WHERE subject_id = ? ORDER BY id`, subjectID)
	if err != nil {
		return released{}, err
	}
	var ids [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return released{}, err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return released{}, err
	}
	var out released
	for _, id := range ids {
		r, err := releaseClaim(tx, id)
		if err != nil {
			return released{}, err
		}
		out.merge(r)
	}
	return out, nil
}

// ReleaseClaim is releaseClaim on the write's Tx: it schedules the purge of any
// File released and returns the audit changes.
func ReleaseClaim(tx *database.Tx, claimID []byte) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	r, err := releaseClaim(tx.Tx, claimID)
	if err != nil {
		return nil, err
	}
	files.PurgeAfterCommit(tx, r.Objects)
	return r.Changes, nil
}

// releaseClaim removes one identity claim, audited: its pins first (through
// ReleaseFacets), then the row, then the claim's own delete change carrying the
// fields the effects registry resolves a ghost row from. It is the one claim
// removal; Subject delete and Unpromote both call it. Does not commit.
func releaseClaim(tx *sql.Tx, claimID []byte) (released, error) {
	if tx == nil || len(claimID) != 16 {
		return released{}, ErrInvalid
	}
	var (
		id, subjectID, entityID, typeID, gradeID []byte
		status                                   string
		argument                                 sql.NullString
	)
	if err := tx.QueryRow(`SELECT id, subject_id, entity_id, subject_type_id, status,
		confidence_grade_id, argument FROM identity_claims WHERE id = ?`, claimID).
		Scan(&id, &subjectID, &entityID, &typeID, &status, &gradeID, &argument); err != nil {
		return released{}, err
	}
	out, err := releaseFacets(tx, catalogmodel.KindIdentityClaim, id)
	if err != nil {
		return released{}, err
	}
	if _, err := tx.Exec(`DELETE FROM identity_claims WHERE id = ?`, id); err != nil {
		return released{}, err
	}
	out.Changes = append(out.Changes, rowchange.Change{
		EntityType: "identity_claim",
		EntityID:   id,
		Action:     rowchange.ActionDelete,
		Fields: rowchange.DeletedRow(map[string]any{
			"id":                  uuidJSON(id),
			"subject_id":          uuidJSON(subjectID),
			"entity_id":           uuidJSON(entityID),
			"subject_type_id":     uuidJSON(typeID),
			"status":              status,
			"confidence_grade_id": uuidJSON(gradeID),
			"argument":            nullStringJSON(argument),
		}),
	})
	return out, nil
}

// pinRelease deletes the pins q selects (claim id, observation id) and audits
// each under its claim's id, so replaying (identity_claim_evidence, claim id)
// yields that claim's exhibit history. The claim join keeps a pin whose claim
// is already gone out of this release; effects.Handles walks identity_claim_id
// to the handle, including through that claim's delete when the row is gone.
func pinRelease(q string) func(*sql.Tx, []byte) (released, error) {
	return func(tx *sql.Tx, parentID []byte) (released, error) {
		rows, err := tx.Query(q, parentID)
		if err != nil {
			return released{}, err
		}
		type pin struct{ claimID, observationID []byte }
		var pins []pin
		for rows.Next() {
			var p pin
			if err := rows.Scan(&p.claimID, &p.observationID); err != nil {
				rows.Close()
				return released{}, err
			}
			pins = append(pins, p)
		}
		if err := rows.Close(); err != nil {
			return released{}, err
		}
		var out released
		for _, p := range pins {
			if _, err := tx.Exec(`DELETE FROM identity_claim_evidence
				WHERE identity_claim_id = ? AND observation_id = ?`, p.claimID, p.observationID); err != nil {
				return released{}, err
			}
			out.Changes = append(out.Changes, rowchange.Change{
				EntityType: "identity_claim_evidence",
				EntityID:   p.claimID,
				Action:     rowchange.ActionDelete,
				Fields: rowchange.DeletedRow(map[string]any{
					"identity_claim_id": uuidJSON(p.claimID),
					"observation_id":    uuidJSON(p.observationID),
				}),
			})
		}
		return out, nil
	}
}

// --- Interpretation releases ------------------------------------------------

func releaseConnectionFacets(tx *sql.Tx, subjectID []byte) (released, error) {
	rows, err := listSubjectObs(tx, subjectID)
	if err != nil {
		return released{}, err
	}
	var out released
	for _, r := range rows {
		if !isConnectionFacet(r) {
			continue
		}
		erased, err := eraseObservation(tx, r.id)
		if err != nil {
			return released{}, err
		}
		out.merge(erased)
	}
	return out, nil
}

// eraseObservation removes one Observation whole: its facets, the row, and its
// owned date / name values, auditing the full row. Only connection-facet
// release uses it; ordinary Observations go through observations.Delete.
func eraseObservation(tx *sql.Tx, id []byte) (released, error) {
	var (
		obsID, citationID, subjectID, propertyID, dateID, nameID, valueSubjectID, termID, depictionID []byte
		ref, polarity                                                                                 string
		text                                                                                          sql.NullString
		integer                                                                                       sql.NullInt64
	)
	if err := tx.QueryRow(`SELECT id, ref, citation_id, subject_id, property_id, polarity,
		value_text, value_integer, value_date_id, value_name_id, value_subject_id, value_term_id, value_depiction_id
		FROM observations WHERE id = ?`, id).Scan(&obsID, &ref, &citationID, &subjectID, &propertyID,
		&polarity, &text, &integer, &dateID, &nameID, &valueSubjectID, &termID, &depictionID); err != nil {
		return released{}, err
	}
	out, err := releaseFacets(tx, catalogmodel.KindObservation, id)
	if err != nil {
		return released{}, err
	}
	snap, err := SnapshotOwned(tx, catalogmodel.KindObservation, id)
	if err != nil {
		return released{}, err
	}
	if _, err := tx.Exec(`DELETE FROM observations WHERE id = ?`, id); err != nil {
		return released{}, err
	}
	objects, err := releaseSnapshot(tx, snap)
	if err != nil {
		return released{}, err
	}
	out.Objects = append(out.Objects, objects...)
	var integerJSON any
	if integer.Valid {
		integerJSON = integer.Int64
	}
	out.Changes = append(out.Changes, rowchange.Change{
		EntityType: "observation",
		EntityID:   obsID,
		Action:     rowchange.ActionDelete,
		Fields: rowchange.DeletedRow(map[string]any{
			"id":                 uuidJSON(obsID),
			"ref":                ref,
			"citation_id":        uuidJSON(citationID),
			"subject_id":         uuidJSON(subjectID),
			"property_id":        uuidJSON(propertyID),
			"polarity":           polarity,
			"value_text":         nullStringJSON(text),
			"value_integer":      integerJSON,
			"value_date_id":      uuidJSON(dateID),
			"value_name_id":      uuidJSON(nameID),
			"value_subject_id":   uuidJSON(valueSubjectID),
			"value_term_id":      uuidJSON(termID),
			"value_depiction_id": uuidJSON(depictionID),
		}),
	})
	return out, nil
}

// --- Plain row facets -------------------------------------------------------

type colKind int

const (
	colUUID colKind = iota
	colText
	colInt
	colBool
)

type col struct {
	name string
	kind colKind
}

// rowFacet is a facet table whose rows are audited whole as DeletedRow, keyed
// by entityIDCol. Rows are removed in rowid order.
func rowFacet(parent catalogmodel.Kind, table, fkCol, entityType, entityIDCol string, cols ...col) facetRelease {
	names := ""
	idIdx := -1
	for i, c := range cols {
		if i > 0 {
			names += ", "
		}
		names += c.name
		if c.name == entityIDCol {
			idIdx = i
		}
	}
	if idIdx < 0 {
		panic("deleteimpact: rowFacet " + table + " lacks entity id column " + entityIDCol)
	}
	selectSQL := fmt.Sprintf(`SELECT rowid, %s FROM %s WHERE %s = ? ORDER BY rowid`, names, table, fkCol)
	deleteSQL := fmt.Sprintf(`DELETE FROM %s WHERE rowid = ?`, table)
	return facetRelease{
		Parent:    parent,
		Via:       table + "." + fkCol,
		Remaining: fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ?`, table, fkCol),
		Release: func(tx *sql.Tx, parentID []byte) (released, error) {
			rows, err := tx.Query(selectSQL, parentID)
			if err != nil {
				return released{}, err
			}
			type row struct {
				rowid    int64
				entityID []byte
				fields   map[string]any
			}
			var found []row
			for rows.Next() {
				var rowid int64
				vals := make([]any, len(cols))
				dest := make([]any, len(cols)+1)
				dest[0] = &rowid
				for i := range vals {
					dest[i+1] = &vals[i]
				}
				if err := rows.Scan(dest...); err != nil {
					rows.Close()
					return released{}, err
				}
				fields := make(map[string]any, len(cols))
				for i, c := range cols {
					fields[c.name] = colJSON(c.kind, vals[i])
				}
				id, _ := vals[idIdx].([]byte)
				found = append(found, row{rowid: rowid, entityID: append([]byte(nil), id...), fields: fields})
			}
			if err := rows.Close(); err != nil {
				return released{}, err
			}
			var out released
			for _, r := range found {
				if _, err := tx.Exec(deleteSQL, r.rowid); err != nil {
					return released{}, err
				}
				out.Changes = append(out.Changes, rowchange.Change{
					EntityType: entityType,
					EntityID:   r.entityID,
					Action:     rowchange.ActionDelete,
					Fields:     rowchange.DeletedRow(r.fields),
				})
			}
			return out, nil
		},
	}
}

func colJSON(kind colKind, v any) any {
	if v == nil {
		return nil
	}
	switch kind {
	case colUUID:
		b, _ := v.([]byte)
		return uuidJSON(b)
	case colText:
		switch s := v.(type) {
		case string:
			return s
		case []byte:
			return string(s)
		}
	case colInt:
		if n, ok := v.(int64); ok {
			return n
		}
	case colBool:
		if n, ok := v.(int64); ok {
			return n != 0
		}
	}
	return v
}

func uuidJSON(id []byte) any {
	if s := uuidString(id); s != "" {
		return s
	}
	return nil
}

func nullStringJSON(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}
