// Package evidencegraphs owns Evidence graphs and their Artifact links.
package evidencegraphs

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/ref"
)

// DefaultName is the name of the graph every Source starts with.
const DefaultName = "Default"

var (
	ErrInvalid       = apperr.New(apperr.CodeEvidenceGraphsInvalid, apperr.KindUser)
	ErrArtifactInUse = apperr.New(apperr.CodeEvidenceGraphsArtifactInUse, apperr.KindUser)
	ErrLastArtifact  = apperr.New(apperr.CodeEvidenceGraphsLastArtifact, apperr.KindUser)
	// ErrInUse is returned by Delete while Subjects belong to the graph.
	ErrInUse = apperr.New(apperr.CodeEvidenceGraphsInUse, apperr.KindConflict)
	// ErrLastGraph: a Source keeps at least one graph.
	ErrLastGraph = apperr.New(apperr.CodeEvidenceGraphsLastGraph, apperr.KindUser)
)

const (
	sqlInsert = `INSERT INTO evidence_graphs (id, ref, source_id, name, description)
		VALUES (?, ?, ?, ?, ?)`
	// The per-Source and project-wide reads share their columns and order; the
	// per-Source form filters on source_id directly so it seeks
	// evidence_graphs_source_id_idx.
	sqlListSelect = `SELECT id, COALESCE(ref, ''), source_id, name, COALESCE(description, '')
		FROM evidence_graphs`
	sqlListOrder    = ` ORDER BY name COLLATE NOCASE, ref, id`
	sqlList         = sqlListSelect + sqlListOrder
	sqlListBySource = sqlListSelect + ` WHERE source_id = ?` + sqlListOrder
	sqlLinksSelect  = `SELECT l.graph_id, l.artifact_id FROM evidence_graphs g
		JOIN evidence_graph_artifacts l ON l.graph_id = g.id
		JOIN artifacts a ON a.id = l.artifact_id`
	sqlLinksOrder        = ` ORDER BY a.ref COLLATE NOCASE, a.id`
	sqlListLinks         = sqlLinksSelect + sqlLinksOrder
	sqlListLinksBySource = sqlLinksSelect + ` WHERE g.source_id = ?` + sqlLinksOrder

	sqlGet = `SELECT id, COALESCE(ref, ''), source_id, name, COALESCE(description, '')
		FROM evidence_graphs WHERE id = ?`
	sqlDelete = `DELETE FROM evidence_graphs WHERE id = ?`
	sqlUpdate = `UPDATE evidence_graphs SET name = ?, description = ? WHERE id = ?`

	sqlSubjectGraph = `SELECT graph_id FROM subjects WHERE id = ?`
	sqlIsLinked     = `SELECT 1 FROM evidence_graph_artifacts WHERE graph_id = ? AND artifact_id = ?`
	sqlLinkSources  = `SELECT g.source_id, a.source_id FROM evidence_graphs g, artifacts a
		WHERE g.id = ? AND a.id = ?`
	sqlLinkInsert = `INSERT INTO evidence_graph_artifacts (graph_id, artifact_id) VALUES (?, ?)`
	sqlLinkDelete = `DELETE FROM evidence_graph_artifacts WHERE graph_id = ? AND artifact_id = ?`
	sqlLinkCount  = `SELECT COUNT(*) FROM evidence_graph_artifacts WHERE graph_id = ?`
	// The ref is selected for the order and dropped.
	sqlCitationsUsing = `SELECT DISTINCT c.id, c.ref FROM citations c
		JOIN observations o ON o.citation_id = c.id
		JOIN subjects s ON s.id = o.subject_id
		WHERE c.artifact_id = ? AND s.graph_id = ?
		ORDER BY c.ref COLLATE NOCASE`
	sqlMissingRef = `SELECT id FROM evidence_graphs WHERE ref IS NULL`
	sqlSetRef     = `UPDATE evidence_graphs SET ref = ? WHERE id = ?`
)

// Graph is one evidence_graphs row.
type Graph struct {
	ID          []byte
	Ref         string
	SourceID    []byte
	Name        string
	Description string
	ArtifactIDs [][]byte // linked Artifacts: by Artifact ref from List, in input order from Create
}

// List returns the graphs of one Source, or of the whole project when sourceID
// is empty, by name (case-insensitive), each with its linked Artifact ids by
// Artifact ref. An unknown Source has no graphs.
func List(c *database.Catalog, sourceID []byte) ([]Graph, error) {
	if len(sourceID) != 0 && len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return list(tx, sourceID)
}

// list is the one read of graph rows, shared by List and the package's writes.
func list(tx *sql.Tx, sourceID []byte) ([]Graph, error) {
	if len(sourceID) != 0 && len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	graphQuery, linkQuery := sqlList, sqlListLinks
	var args []any
	if len(sourceID) != 0 {
		graphQuery, linkQuery = sqlListBySource, sqlListLinksBySource
		args = []any{sourceID}
	}
	rows, err := tx.Query(graphQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Graph
	for rows.Next() {
		g, err := scanGraph(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	index := make(map[string]int, len(out))
	for i, g := range out {
		index[string(g.ID)] = i
	}
	links, err := tx.Query(linkQuery, args...)
	if err != nil {
		return nil, err
	}
	defer links.Close()
	for links.Next() {
		var graphID, artifactID []byte
		if err := links.Scan(&graphID, &artifactID); err != nil {
			return nil, err
		}
		if i, ok := index[string(graphID)]; ok {
			out[i].ArtifactIDs = append(out[i].ArtifactIDs, artifactID)
		}
	}
	return out, links.Err()
}

// scanGraph scans the columns sqlList and sqlGet select.
func scanGraph(row interface{ Scan(dest ...any) error }) (Graph, error) {
	var g Graph
	err := row.Scan(&g.ID, &g.Ref, &g.SourceID, &g.Name, &g.Description)
	return g, err
}

// CreateInput is the input of Create.
type CreateInput struct {
	SourceID    []byte
	Name        string
	Description string
	ArtifactIDs [][]byte // at least one, each an Artifact of SourceID
}

// Create inserts a graph of a Source and links the given Artifacts to it. Names
// need not be unique. Any refusal leaves nothing behind once the write rolls
// back.
func Create(tx *database.Tx, userID []byte, in CreateInput) (Graph, []rowchange.Change, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if tx == nil || len(in.SourceID) != 16 || in.Name == "" || len(in.ArtifactIDs) == 0 {
		return Graph{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Graph{}, nil, err
	}
	if err := requireDistinct(in.ArtifactIDs); err != nil {
		return Graph{}, nil, err
	}
	g, change, err := insert(tx.Tx, in.SourceID, in.Name, in.Description)
	if err != nil {
		return Graph{}, nil, err
	}
	changes := []rowchange.Change{change}
	for _, id := range in.ArtifactIDs {
		ch, err := Link(tx, userID, g.ID, id)
		if err != nil {
			return Graph{}, nil, err
		}
		changes = append(changes, ch...)
	}
	for _, id := range in.ArtifactIDs {
		g.ArtifactIDs = append(g.ArtifactIDs, append([]byte(nil), id...))
	}
	return g, changes, nil
}

// Delete deletes a graph that is not its Source's last and has no Subjects,
// releasing its links. Deleteimpact's report decides, so the delete preview and
// this write agree.
func Delete(tx *database.Tx, userID, graphID []byte) ([]rowchange.Change, error) {
	if tx == nil || len(graphID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	g, err := getTx(tx.Tx, graphID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	report, err := deleteimpact.Impact(tx.Tx, catalogmodel.KindEvidenceGraph, graphID)
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.Refuse(report, deleteimpact.Codes{
		InUse: ErrInUse, LastGraph: ErrLastGraph, NotFound: ErrInvalid,
	}); err != nil {
		return nil, err
	}
	released, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindEvidenceGraph, graphID)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDelete, graphID); err != nil {
		return nil, err
	}
	return append(released, rowchange.Change{
		EntityType: "evidence_graph",
		EntityID:   append([]byte(nil), graphID...),
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(graphFields(g)),
	}), nil
}

// UpdateInput is the input of Update.
type UpdateInput struct {
	Name        string
	Description string
	ArtifactIDs [][]byte // empty = links unchanged; otherwise the full set of linked Artifacts
}

// Update saves a graph's name, description and linked Artifacts in one write.
// Artifacts are linked before others are unlinked, so a swap never leaves the
// graph with no link. Any refusal returns its error unchanged, and the write
// rolls back the whole edit.
func Update(tx *database.Tx, userID, graphID []byte, in UpdateInput) ([]rowchange.Change, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if tx == nil || len(graphID) != 16 || in.Name == "" {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if err := requireDistinct(in.ArtifactIDs); err != nil {
		return nil, err
	}
	g, err := getTx(tx.Tx, graphID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}

	var changes []rowchange.Change
	fields := map[string]rowchange.FieldDiff{}
	if g.Name != in.Name {
		fields["name"] = rowchange.FieldDiff{Old: g.Name, New: in.Name}
	}
	if g.Description != in.Description {
		fields["description"] = rowchange.FieldDiff{Old: database.NullString(g.Description), New: database.NullString(in.Description)}
	}
	if len(fields) > 0 {
		if _, err := tx.Exec(sqlUpdate, in.Name, database.NullString(in.Description), graphID); err != nil {
			return nil, mapConstraint(err)
		}
		changes = append(changes, rowchange.Change{
			EntityType: "evidence_graph",
			EntityID:   append([]byte(nil), graphID...),
			Action:     rowchange.ActionUpdate,
			Fields:     fields,
		})
	}

	if len(in.ArtifactIDs) == 0 {
		return changes, nil
	}
	graphs, err := list(tx.Tx, g.SourceID)
	if err != nil {
		return nil, err
	}
	var stored [][]byte
	for _, x := range graphs {
		if bytes.Equal(x.ID, graphID) {
			stored = x.ArtifactIDs
		}
	}
	for _, id := range in.ArtifactIDs {
		if database.ContainsBlobID(stored, id) {
			continue
		}
		ch, err := Link(tx, userID, graphID, id)
		if err != nil {
			return nil, err
		}
		changes = append(changes, ch...)
	}
	for _, id := range stored {
		if database.ContainsBlobID(in.ArtifactIDs, id) {
			continue
		}
		ch, err := Unlink(tx, userID, graphID, id)
		if err != nil {
			return nil, err
		}
		changes = append(changes, ch...)
	}
	return changes, nil
}

// getTx loads one graph row; sql.ErrNoRows when there is none. It does not fill
// ArtifactIDs.
func getTx(tx *sql.Tx, id []byte) (Graph, error) {
	return scanGraph(tx.QueryRow(sqlGet, id))
}

// requireDistinct refuses a repeated id.
func requireDistinct(ids [][]byte) error {
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[string(id)] {
			return ErrInvalid
		}
		seen[string(id)] = true
	}
	return nil
}

// CreateDefault inserts the Default graph of a new Source. Its only caller is
// sources.Create, which has already checked the user and the Source.
func CreateDefault(tx *database.Tx, sourceID []byte) (Graph, []rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 {
		return Graph{}, nil, ErrInvalid
	}
	g, change, err := insert(tx.Tx, sourceID, DefaultName, "")
	if err != nil {
		return Graph{}, nil, err
	}
	return g, []rowchange.Change{change}, nil
}

// mintRef is the ref source for new graphs; tests replace it to force a collision.
var mintRef = database.Minter(ref.PrefixEvidenceGraph)

func insert(tx *sql.Tx, sourceID []byte, name, description string) (Graph, rowchange.Change, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return Graph{}, rowchange.Change{}, err
	}
	idBytes := id[:]

	graphRef, err := database.InsertWithRef(mintRef, func(storedRef string) error {
		_, err := tx.Exec(sqlInsert, idBytes, storedRef, sourceID, name, database.NullString(description))
		return err
	})
	if errors.Is(err, database.ErrRefsExhausted) {
		return Graph{}, rowchange.Change{}, ErrInvalid
	}
	if err != nil {
		return Graph{}, rowchange.Change{}, mapConstraint(err)
	}

	g := Graph{
		ID:          append([]byte(nil), idBytes...),
		Ref:         graphRef,
		SourceID:    append([]byte(nil), sourceID...),
		Name:        name,
		Description: description,
	}
	change := rowchange.Change{
		EntityType: "evidence_graph",
		EntityID:   append([]byte(nil), idBytes...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(graphFields(g)),
	}
	return g, change, nil
}

// graphFields is the audit field map of a graph row, for its create and delete.
func graphFields(g Graph) map[string]any {
	return map[string]any{
		"id":          database.UUIDJSON(g.ID),
		"ref":         database.NullString(g.Ref),
		"source_id":   database.UUIDJSON(g.SourceID),
		"name":        g.Name,
		"description": database.NullString(g.Description),
	}
}

// SoleGraphTx returns the id of the one graph of a Source.
//
// The one bridge between S12-03 and S12-08. S12-08 deletes it.
func SoleGraphTx(tx *sql.Tx, sourceID []byte) ([]byte, error) {
	if tx == nil || len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	graphs, err := list(tx, sourceID)
	if err != nil {
		return nil, err
	}
	switch len(graphs) {
	case 0:
		return nil, ErrInvalid
	case 1:
		return graphs[0].ID, nil
	default:
		return nil, fmt.Errorf("evidencegraphs: source %x has more than one graph", sourceID)
	}
}

// SoleGraph returns the id of the one graph of a Source, in its own read
// transaction, the way List wraps list. S12-08 deletes it with SoleGraphTx.
func SoleGraph(c *database.Catalog, sourceID []byte) ([]byte, error) {
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return SoleGraphTx(tx, sourceID)
}

// SubjectGraphTx returns the graph a Subject belongs to.
func SubjectGraphTx(tx *sql.Tx, subjectID []byte) ([]byte, error) {
	if tx == nil || len(subjectID) != 16 {
		return nil, ErrInvalid
	}
	var id []byte
	err := tx.QueryRow(sqlSubjectGraph, subjectID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	return id, nil
}

// IsLinkedTx reports whether the graph and the Artifact are linked.
func IsLinkedTx(tx *sql.Tx, graphID, artifactID []byte) (bool, error) {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return false, ErrInvalid
	}
	var one int
	err := tx.QueryRow(sqlIsLinked, graphID, artifactID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CitationsUsingArtifactTx returns the Citations on the Artifact that have at
// least one Observation on a Subject of the graph, with their refs, ordered by
// ref. Only ID and Ref are set.
func CitationsUsingArtifactTx(tx *sql.Tx, graphID, artifactID []byte) ([]deleteimpact.Listed, error) {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := tx.Query(sqlCitationsUsing, artifactID, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []deleteimpact.Listed
	for rows.Next() {
		var row deleteimpact.Listed
		if err := rows.Scan(&row.ID, &row.Ref); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// UnlinkVia names the preflight group of Citations that keep an Artifact
// linked. A domain-preflight via is named for its refusal code, not an FK edge.
const UnlinkVia = "evidencegraphs.artifact_in_use"

// UnlinkBlock is one Artifact an edit would remove that the graph's Citations
// still use, with those Citations shaped for the DeleteImpact recipe.
type UnlinkBlock struct {
	ArtifactID []byte
	Citations  deleteimpact.Group
}

// UnlinkImpact reports which of the Artifacts an edit would unlink from the
// graph are used by the graph's Citations, in the order given. An Artifact that
// is not linked, or that no Citation uses, has no block.
func UnlinkImpact(c *database.Catalog, graphID []byte, artifactIDs [][]byte) ([]UnlinkBlock, error) {
	if len(graphID) != 16 {
		return nil, ErrInvalid
	}
	for _, id := range artifactIDs {
		if len(id) != 16 {
			return nil, ErrInvalid
		}
	}
	if err := requireDistinct(artifactIDs); err != nil {
		return nil, err
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := getTx(tx, graphID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalid
		}
		return nil, err
	}
	var out []UnlinkBlock
	for _, artifactID := range artifactIDs {
		linked, err := IsLinkedTx(tx, graphID, artifactID)
		if err != nil {
			return nil, err
		}
		if !linked {
			continue
		}
		rows, err := CitationsUsingArtifactTx(tx, graphID, artifactID)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			continue
		}
		group, err := deleteimpact.ListedGroup(tx, UnlinkVia, catalogmodel.KindCitation, rows)
		if err != nil {
			return nil, err
		}
		out = append(out, UnlinkBlock{ArtifactID: append([]byte(nil), artifactID...), Citations: group})
	}
	return out, nil
}

// checkLink validates the shared input of Link and Unlink: the graph and the
// Artifact exist and belong to the same Source.
func checkLink(tx *database.Tx, userID, graphID, artifactID []byte) error {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}
	var graphSource, artifactSource []byte
	err := tx.QueryRow(sqlLinkSources, graphID, artifactID).Scan(&graphSource, &artifactSource)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	if !bytes.Equal(graphSource, artifactSource) {
		return ErrInvalid
	}
	return nil
}

func linkFields(graphID, artifactID []byte) map[string]any {
	return map[string]any{
		"graph_id":    database.UUIDJSON(graphID),
		"artifact_id": database.UUIDJSON(artifactID),
	}
}

// Link links an Artifact to a graph of its Source. An existing link returns no
// changes.
func Link(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error) {
	if err := checkLink(tx, userID, graphID, artifactID); err != nil {
		return nil, err
	}
	linked, err := IsLinkedTx(tx.Tx, graphID, artifactID)
	if err != nil {
		return nil, err
	}
	if linked {
		return nil, nil
	}
	if _, err := tx.Exec(sqlLinkInsert, graphID, artifactID); err != nil {
		return nil, mapConstraint(err)
	}
	return []rowchange.Change{{
		EntityType: "evidence_graph_artifact",
		EntityID:   append([]byte(nil), graphID...),
		Action:     rowchange.ActionCreate,
		Fields:     rowchange.FullRow(linkFields(graphID, artifactID)),
	}}, nil
}

// Unlink removes an Artifact from a graph. A link that is not there returns no
// changes. The graph's Citations keep the Artifact linked (ErrArtifactInUse),
// and a graph keeps at least one Artifact (ErrLastArtifact).
func Unlink(tx *database.Tx, userID, graphID, artifactID []byte) ([]rowchange.Change, error) {
	if err := checkLink(tx, userID, graphID, artifactID); err != nil {
		return nil, err
	}
	linked, err := IsLinkedTx(tx.Tx, graphID, artifactID)
	if err != nil {
		return nil, err
	}
	if !linked {
		return nil, nil
	}
	using, err := CitationsUsingArtifactTx(tx.Tx, graphID, artifactID)
	if err != nil {
		return nil, err
	}
	if len(using) > 0 {
		return nil, ErrArtifactInUse
	}
	var n int
	if err := tx.QueryRow(sqlLinkCount, graphID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 1 {
		return nil, ErrLastArtifact
	}
	if _, err := tx.Exec(sqlLinkDelete, graphID, artifactID); err != nil {
		return nil, err
	}
	return []rowchange.Change{{
		EntityType: "evidence_graph_artifact",
		EntityID:   append([]byte(nil), graphID...),
		Action:     rowchange.ActionDelete,
		Fields:     rowchange.DeletedRow(linkFields(graphID, artifactID)),
	}}, nil
}

// LinkArtifact links a new Artifact to the given graphs of its Source; no ids
// means every graph of the Source. A Source with no graph (only direct-SQL
// fixtures) gets no link.
func LinkArtifact(tx *database.Tx, userID, sourceID, artifactID []byte, graphIDs [][]byte) ([]rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	if len(graphIDs) == 0 {
		graphs, err := list(tx.Tx, sourceID)
		if err != nil {
			return nil, err
		}
		for _, g := range graphs {
			graphIDs = append(graphIDs, g.ID)
		}
	} else {
		if err := requireDistinct(graphIDs); err != nil {
			return nil, err
		}
	}
	var changes []rowchange.Change
	for _, id := range graphIDs {
		ch, err := Link(tx, userID, id, artifactID)
		if err != nil {
			return nil, err
		}
		changes = append(changes, ch...)
	}
	return changes, nil
}

// EnsureRefs mints GPH-… refs for any evidence_graphs row missing one: the
// graphs the format migration backfilled. Unaudited, like users.EnsureRefs, but
// in one transaction with insert's retry on a ref collision.
func EnsureRefs(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	rows, err := db.Query(sqlMissingRef)
	if err != nil {
		return err
	}
	defer rows.Close()
	var need [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return err
		}
		need = append(need, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range need {
		_, err := database.InsertWithRef(mintRef, func(r string) error {
			_, err := tx.Exec(sqlSetRef, r, id)
			return err
		})
		if errors.Is(err, database.ErrRefsExhausted) {
			return ErrInvalid
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
