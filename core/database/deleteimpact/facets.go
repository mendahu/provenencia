package deleteimpact

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/observationrow"
	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/mendahu/provenencia/core/apperr"
)

// Facet release: research rows an official Delete removes explicitly — and
// audits — before its parent DELETE, plus optional back-pointers cleared with
// an update change before the parent DELETE (see optionalPointer). Each
// facetRelease is the single definition of one such relationship: the honesty
// test checks it against the FK, Impact reads it for Report.Cascades when
// Named, and ReleaseFacets mutates through it. The schema's ON DELETE CASCADE
// and SET NULL are only backstops: ReleaseFacets fails the delete if any row
// it owns is still there afterwards.

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
	// registered Audited and a BucketOptional SET NULL FK whose parent is
	// deletable must each have exactly one release.
	Via string
	// Named facets appear in Report.Cascades: Child is the listed kind, and
	// Count / List must select exactly what Release removes.
	Named bool
	Child catalogmodel.Kind
	Count func(tx *sql.Tx, parentID []byte) (int, error)
	List  func(tx *sql.Tx, parentID []byte, limit int) ([]probeRow, error)
	// Remaining counts rows still pointing at the parent after Release; it
	// must be zero. Empty only for connection-facet releases, whose leftovers
	// Impact has already refused. Required for CASCADE and optional SET NULL.
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

// cascadesFor is the Named facets of kind plus cascade-only edges, as non-blocking Impact edges.
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
	for _, e := range cascadeEdgeList {
		if e.Parent == kind {
			out = append(out, e)
		}
	}
	return out
}

// facetReleaseList is built in init: releases recurse through ReleaseFacets,
// which a package-level initializer cannot reference (initialization cycle).
var facetReleaseList []facetRelease

func init() {
	facetReleaseList = buildFacetReleases()
	cascadeEdgeList = buildCascadeEdges()
}

func facetReleases() []facetRelease { return facetReleaseList }

func buildFacetReleases() []facetRelease {
	return []facetRelease{
		// --- Conclusion ---------------------------------------------------
		// A Subject releases every claim it has on a handle; a handle whose last
		// accepted member goes is deleted inside releaseClaim.
		{
			Parent: catalogmodel.KindSubject, Via: ViaLeavesHandle,
			Remaining: `SELECT COUNT(*) FROM identity_claims WHERE subject_id = ?`,
			Release:   releaseEach(`SELECT id FROM identity_claims WHERE subject_id = ? ORDER BY id`, releaseClaim),
		},
		{
			Parent: catalogmodel.KindCanonicalEntity, Via: "identity_claims.entity_id",
			Remaining: `SELECT COUNT(*) FROM identity_claims WHERE entity_id = ?`,
			Release:   releaseEach(`SELECT id FROM identity_claims WHERE entity_id = ? ORDER BY id`, releaseClaimRowOnly),
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
		optionalPointer(catalogmodel.KindObservation, "canonical_entities", "preferred_likeness_observation_id", "canonical_entity", "id"),
		rowFacet(catalogmodel.KindCitation, "citation_notes", "citation_id", "citation_note", "id",
			col{"id", colUUID}, col{"citation_id", colUUID}, col{"body", colText}),
		// A link is audited under its graph's id, like the graph's own create.
		rowFacet(catalogmodel.KindArtifact, "evidence_graph_artifacts", "artifact_id", "evidence_graph_artifact", "graph_id",
			col{"graph_id", colUUID}, col{"artifact_id", colUUID}),
		optionalPointer(catalogmodel.KindArtifact, "sources", "primary_artifact_id", "source", "id", col{"cover_mode", colText}),
		// An Artifact's metadata goes with it; layout rows are audited under
		// field_id, as a Source's are.
		rowFacet(catalogmodel.KindArtifact, "artifact_metadata", "artifact_id", "artifact_metadata", "id",
			col{"id", colUUID}, col{"artifact_id", colUUID}, col{"field_id", colUUID}, col{"value_text", colText}),
		rowFacet(catalogmodel.KindArtifact, "artifact_metadata_layout", "artifact_id", "artifact_metadata_layout", "field_id",
			col{"artifact_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),
		// A graph's Citations go with it, each through releaseCitation (a graph
		// with Subjects is refused first, so none has an Observation).
		{
			Parent: catalogmodel.KindEvidenceGraph, Via: ViaDeletesCitations,
			Named: true, Child: catalogmodel.KindCitation,
			Count:     countSQLFn(`SELECT COUNT(*) FROM citations WHERE graph_id = ?`),
			List:      listSQLFn(`SELECT id, ref FROM citations WHERE graph_id = ? ORDER BY ref COLLATE NOCASE LIMIT ?`),
			Remaining: `SELECT COUNT(*) FROM citations WHERE graph_id = ?`,
			Release:   releaseEach(`SELECT id FROM citations WHERE graph_id = ? ORDER BY id`, releaseCitation),
		},
		// A graph's links go with it, audited under its id.
		rowFacet(catalogmodel.KindEvidenceGraph, "evidence_graph_artifacts", "graph_id", "evidence_graph_artifact", "graph_id",
			col{"graph_id", colUUID}, col{"artifact_id", colUUID}),
		rowFacet(catalogmodel.KindEvidenceGraph, "evidence_graph_metadata", "graph_id", "evidence_graph_metadata", "id",
			col{"id", colUUID}, col{"graph_id", colUUID}, col{"field_id", colUUID}, col{"value_text", colText}),
		rowFacet(catalogmodel.KindEvidenceGraph, "evidence_graph_metadata_layout", "graph_id", "evidence_graph_metadata_layout", "field_id",
			col{"graph_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),

		// --- Source ---------------------------------------------------------
		rowFacet(catalogmodel.KindSource, "source_notes", "source_id", "source_note", "id",
			col{"id", colUUID}, col{"source_id", colUUID}, col{"body", colText}),
		// Each graph goes through releaseGraph, in id order: its own facets
		// (a deletable Source's graphs hold only metadata), then its row and change.
		{
			Parent: catalogmodel.KindSource, Via: "evidence_graphs.source_id",
			Remaining: `SELECT COUNT(*) FROM evidence_graphs WHERE source_id = ?`,
			Release:   releaseEach(`SELECT id FROM evidence_graphs WHERE source_id = ? ORDER BY id`, releaseGraph),
		},
		rowFacet(catalogmodel.KindSource, "source_metadata", "source_id", "source_metadata", "id",
			col{"id", colUUID}, col{"source_id", colUUID}, col{"field_id", colUUID}, col{"value_text", colText}),
		rowFacet(catalogmodel.KindSource, "source_credibility_assessments", "source_id", "source_credibility_assessment", "id",
			col{"id", colUUID}, col{"source_id", colUUID}, col{"credibility_grade_id", colUUID}, col{"argument", colText}),
		// Layout rows are audited under field_id (sourcemetadata precedent).
		rowFacet(catalogmodel.KindSource, "source_metadata_layout", "source_id", "source_metadata_layout", "field_id",
			col{"source_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),
		rowFacet(catalogmodel.KindMetadataField, "source_metadata_layout", "field_id", "source_metadata_layout", "field_id",
			col{"source_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),
		rowFacet(catalogmodel.KindMetadataField, "artifact_metadata_layout", "field_id", "artifact_metadata_layout", "field_id",
			col{"artifact_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),
		rowFacet(catalogmodel.KindMetadataField, "evidence_graph_metadata_layout", "field_id", "evidence_graph_metadata_layout", "field_id",
			col{"graph_id", colUUID}, col{"field_id", colUUID}, col{"sort_order", colInt}, col{"dismissed", colBool}),

		// A type's bindings are audited under the property id (subjectvocab
		// precedent), released from either parent. Subject types have no
		// delete command yet; the release is what an audited CASCADE needs.
		rowFacet(catalogmodel.KindProperty, "subject_type_properties", "property_id", "subject_type_property", "property_id",
			col{"subject_type_id", colUUID}, col{"property_id", colUUID}, col{"sort_order", colInt}),
		rowFacet(catalogmodel.KindSubjectType, "subject_type_properties", "subject_type_id", "subject_type_property", "property_id",
			col{"subject_type_id", colUUID}, col{"property_id", colUUID}, col{"sort_order", colInt}),
	}
}

// --- Conclusion releases ----------------------------------------------------

// releaseEach releases every row selectIDs returns for the parent id through
// one, in that order, and merges what they removed. The ids are read and the
// rows closed before the first release, which writes on the same Tx.
func releaseEach(selectIDs string, one func(*sql.Tx, []byte) (released, error)) func(*sql.Tx, []byte) (released, error) {
	return func(tx *sql.Tx, parentID []byte) (released, error) {
		rows, err := tx.Query(selectIDs, parentID)
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
			r, err := one(tx, id)
			if err != nil {
				return released{}, err
			}
			out.merge(r)
		}
		return out, nil
	}
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

// releaseClaim removes one identity claim, audited, and deletes the handle when
// the claim was its last accepted member. Subject delete and Unpromote both call it.
func releaseClaim(tx *sql.Tx, claimID []byte) (released, error) {
	subjectID, entityID, status, out, err := releaseClaimRow(tx, claimID)
	if err != nil {
		return released{}, err
	}
	if status != claimStatusAccepted {
		return out, nil
	}
	extra, err := releaseHandleIfGone(tx, entityID, subjectID)
	if err != nil {
		return released{}, err
	}
	out.merge(extra)
	return out, nil
}

func releaseClaimRowOnly(tx *sql.Tx, claimID []byte) (released, error) {
	_, _, _, out, err := releaseClaimRow(tx, claimID)
	return out, err
}

// releaseClaimRow removes one identity claim row and its delete change without
// deleting the handle (the entity facet uses this to avoid recursion).
func releaseClaimRow(tx *sql.Tx, claimID []byte) (subjectID, entityID []byte, status string, out released, err error) {
	if tx == nil || len(claimID) != 16 {
		return nil, nil, "", released{}, ErrInvalid
	}
	var id, typeID []byte
	if err := tx.QueryRow(`SELECT id, subject_id, entity_id, subject_type_id, status
		FROM identity_claims WHERE id = ?`, claimID).
		Scan(&id, &subjectID, &entityID, &typeID, &status); err != nil {
		return nil, nil, "", released{}, err
	}
	out, err = releaseFacets(tx, catalogmodel.KindIdentityClaim, id)
	if err != nil {
		return nil, nil, "", released{}, err
	}
	if _, err := tx.Exec(`DELETE FROM identity_claims WHERE id = ?`, id); err != nil {
		return nil, nil, "", released{}, err
	}
	out.Changes = append(out.Changes, rowchange.Change{
		EntityType: "identity_claim",
		EntityID:   id,
		Action:     rowchange.ActionDelete,
		Fields: rowchange.DeletedRow(map[string]any{
			"id":              database.UUIDJSON(id),
			"subject_id":      database.UUIDJSON(subjectID),
			"entity_id":       database.UUIDJSON(entityID),
			"subject_type_id": database.UUIDJSON(typeID),
			"status":          status,
		}),
	})
	return subjectID, entityID, status, out, nil
}

// ReleaseCitation is releaseCitation on the write's Tx: it schedules the purge
// of any File released and returns the audit changes.
func ReleaseCitation(tx *database.Tx, citationID []byte) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	r, err := releaseCitation(tx.Tx, citationID)
	if err != nil {
		return nil, err
	}
	files.PurgeAfterCommit(tx, r.Objects)
	return r.Changes, nil
}

// releaseCitation removes one Citation, audited: its notes (through
// releaseFacets), then the row, then the Citation's own delete change over
// every column. It is the one Citation removal; citations.Delete and a graph
// delete both call it. The caller has already refused a Citation that still has
// Observations. Does not commit.
func releaseCitation(tx *sql.Tx, citationID []byte) (released, error) {
	if tx == nil || len(citationID) != 16 {
		return released{}, ErrInvalid
	}
	return releaseRow(tx, catalogmodel.KindCitation, "citations", "citation", citationID,
		col{"id", colUUID}, col{"ref", colText}, col{"artifact_id", colUUID}, col{"graph_id", colUUID},
		col{"locator_json", colText}, col{"transcription", colText}, col{"label", colText},
		col{"transcription_uncertain", colBool}, col{"transcription_note", colText})
}

// ReleaseGraph is releaseGraph on the write's Tx: it schedules the purge of any
// File released and returns the audit changes.
func ReleaseGraph(tx *database.Tx, graphID []byte) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	r, err := releaseGraph(tx.Tx, graphID)
	if err != nil {
		return nil, err
	}
	files.PurgeAfterCommit(tx, r.Objects)
	return r.Changes, nil
}

// releaseGraph removes one Evidence graph, audited: its facets (through
// releaseFacets), then the row, then the graph's own delete change over every
// column. It is the one graph removal; evidencegraphs.Delete and a Source
// delete both call it. The caller has already refused a graph that still has
// Subjects. Does not commit.
func releaseGraph(tx *sql.Tx, graphID []byte) (released, error) {
	if tx == nil || len(graphID) != 16 {
		return released{}, ErrInvalid
	}
	return releaseRow(tx, catalogmodel.KindEvidenceGraph, "evidence_graphs", "evidence_graph", graphID,
		col{"id", colUUID}, col{"ref", colText}, col{"source_id", colUUID}, col{"name", colText},
		col{"description", colText})
}

// releaseRow removes one row by id, audited whole: its facets (through
// releaseFacets), then the row, then a delete change of entityType over cols.
// Citation and graph removal share it. Does not commit.
func releaseRow(tx *sql.Tx, kind catalogmodel.Kind, table, entityType string, id []byte, cols ...col) (released, error) {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.name
	}
	dest, vals := colScanDest(cols)
	if err := tx.QueryRow(fmt.Sprintf(`SELECT %s FROM %s WHERE id = ?`, strings.Join(names, ", "), table), id).
		Scan(dest...); err != nil {
		return released{}, err
	}
	out, err := releaseFacets(tx, kind, id)
	if err != nil {
		return released{}, err
	}
	if _, err := tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table), id); err != nil {
		return released{}, err
	}
	fields := make(map[string]any, len(cols))
	for i, c := range cols {
		fields[c.name] = colJSON(c.kind, vals[i])
	}
	out.Changes = append(out.Changes, rowchange.Change{
		EntityType: entityType,
		EntityID:   id,
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(fields),
	})
	return out, nil
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
	row, err := observationrow.Scan(tx.QueryRow(`SELECT `+observationrow.Columns+` FROM observations WHERE id = ?`, id))
	if err != nil {
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
	out.Changes = append(out.Changes, rowchange.Change{
		EntityType: "observation",
		EntityID:   row.ID,
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(row.AuditFields()),
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

// optionalPointer clears a BucketOptional SET NULL FK on parent delete: it
// nulls the FK, re-reads side-effect columns (for example cover_mode after
// sources_clear_cover_on_primary_null), and emits one update change per row.
func optionalPointer(parent catalogmodel.Kind, table, fkCol, entityType, entityIDCol string, sideEffectCols ...col) facetRelease {
	readCols := append([]col{{entityIDCol, colUUID}, {fkCol, colUUID}}, sideEffectCols...)
	selectNames := readCols[0].name
	for i := 1; i < len(readCols); i++ {
		selectNames += ", " + readCols[i].name
	}
	selectSQL := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ? ORDER BY %s`, selectNames, table, fkCol, entityIDCol)
	updateSQL := fmt.Sprintf(`UPDATE %s SET %s = NULL WHERE %s = ?`, table, fkCol, entityIDCol)
	var reReadSQL string
	if len(sideEffectCols) > 0 {
		reReadNames := sideEffectCols[0].name
		for i := 1; i < len(sideEffectCols); i++ {
			reReadNames += ", " + sideEffectCols[i].name
		}
		reReadSQL = fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`, reReadNames, table, entityIDCol)
	}
	return facetRelease{
		Parent:    parent,
		Via:       table + "." + fkCol,
		Remaining: fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s = ?`, table, fkCol),
		Release: func(tx *sql.Tx, parentID []byte) (released, error) {
			rows, err := tx.Query(selectSQL, parentID)
			if err != nil {
				return released{}, err
			}
			type pointed struct {
				entityID []byte
				fkOld    []byte
				sideOld  []any
			}
			var found []pointed
			for rows.Next() {
				dest, vals := colScanDest(readCols)
				if err := rows.Scan(dest...); err != nil {
					rows.Close()
					return released{}, err
				}
				entityID, _ := vals[0].([]byte)
				fkVal, _ := vals[1].([]byte)
				sideOld := make([]any, len(sideEffectCols))
				for i, c := range sideEffectCols {
					sideOld[i] = colJSON(c.kind, vals[i+2])
				}
				found = append(found, pointed{
					entityID: append([]byte(nil), entityID...),
					fkOld:    append([]byte(nil), fkVal...),
					sideOld:  sideOld,
				})
			}
			if err := rows.Close(); err != nil {
				return released{}, err
			}
			var out released
			for _, r := range found {
				if _, err := tx.Exec(updateSQL, r.entityID); err != nil {
					return released{}, err
				}
				sideNew := make([]any, len(sideEffectCols))
				if len(sideEffectCols) > 0 {
					dest, vals := colScanDest(sideEffectCols)
					if err := tx.QueryRow(reReadSQL, r.entityID).Scan(dest...); err != nil {
						return released{}, err
					}
					for i, c := range sideEffectCols {
						sideNew[i] = colJSON(c.kind, vals[i])
					}
				}
				fields := map[string]rowchange.FieldDiff{
					fkCol: {Old: colJSON(colUUID, r.fkOld), New: nil},
				}
				for i, c := range sideEffectCols {
					fields[c.name] = rowchange.FieldDiff{Old: r.sideOld[i], New: sideNew[i]}
				}
				out.Changes = append(out.Changes, rowchange.Change{
					EntityType: entityType,
					EntityID:   r.entityID,
					Action:     rowchange.ActionUpdate,
					Fields:     fields,
				})
			}
			return out, nil
		},
	}
}

func colScanDest(cols []col) (dest []any, vals []any) {
	vals = make([]any, len(cols))
	dest = make([]any, len(cols))
	for i := range cols {
		dest[i] = &vals[i]
	}
	return dest, vals
}

func colJSON(kind colKind, v any) any {
	if v == nil {
		return nil
	}
	switch kind {
	case colUUID:
		b, _ := v.([]byte)
		return database.UUIDJSON(b)
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

func nullStringJSON(s sql.NullString) any {
	if !s.Valid {
		return nil
	}
	return s.String
}
