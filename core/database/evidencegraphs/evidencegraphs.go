// Package evidencegraphs owns Evidence graphs and their Artifact links.
package evidencegraphs

import (
	"bytes"
	"database/sql"
	"errors"
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
	sqlGraphSource  = `SELECT source_id FROM evidence_graphs WHERE id = ?`
	sqlIsLinked     = `SELECT 1 FROM evidence_graph_artifacts WHERE graph_id = ? AND artifact_id = ?`
	sqlLinkSources  = `SELECT g.source_id, a.source_id FROM evidence_graphs g, artifacts a
		WHERE g.id = ? AND a.id = ?`
	sqlLinkInsert = `INSERT INTO evidence_graph_artifacts (graph_id, artifact_id) VALUES (?, ?)`
	sqlLinkDelete = `DELETE FROM evidence_graph_artifacts WHERE graph_id = ? AND artifact_id = ?`
	sqlLinkCount  = `SELECT COUNT(*) FROM evidence_graph_artifacts WHERE graph_id = ?`
	// sqlArtifactGraphs reads one Artifact's links inside a write.
	sqlArtifactGraphs = `SELECT graph_id FROM evidence_graph_artifacts
		WHERE artifact_id = ? ORDER BY graph_id`
	// A graph's Citations on an Artifact, with or without Observations. The ref
	// is selected for the order and dropped.
	sqlCitationsUsing = `SELECT c.id, c.ref FROM citations c
		WHERE c.artifact_id = ? AND c.graph_id = ?
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

// GraphLinks says which graphs an Artifact is in. The zero value is the
// default: every graph of its Source on create, unchanged on update. Given
// with no IDs is none.
type GraphLinks struct {
	Given bool
	IDs   [][]byte
}

// Pair is one (graph, Artifact) link.
type Pair struct{ GraphID, ArtifactID []byte }

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
	g, change, err := insert(tx.Tx, in.SourceID, in.Name, in.Description)
	if err != nil {
		return Graph{}, nil, err
	}
	want := make([]Pair, len(in.ArtifactIDs))
	for i, id := range in.ArtifactIDs {
		want[i] = Pair{GraphID: g.ID, ArtifactID: id}
	}
	changes := []rowchange.Change{change}
	linked, err := applyLinks(tx, userID, nil, want)
	if err != nil {
		return Graph{}, nil, err
	}
	changes = append(changes, linked...)
	for _, id := range in.ArtifactIDs {
		g.ArtifactIDs = append(g.ArtifactIDs, append([]byte(nil), id...))
	}
	return g, changes, nil
}

// Delete deletes a graph that is not its Source's last and has no Subjects,
// releasing its links and its Citations (audited, through deleteimpact).
// Deleteimpact's report decides, so the delete preview and this write agree.
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
	want := make([]Pair, len(in.ArtifactIDs))
	for i, id := range in.ArtifactIDs {
		want[i] = Pair{GraphID: graphID, ArtifactID: id}
	}
	graphs, err := list(tx.Tx, g.SourceID)
	if err != nil {
		return nil, err
	}
	var stored []Pair
	for _, x := range graphs {
		if bytes.Equal(x.ID, graphID) {
			for _, id := range x.ArtifactIDs {
				stored = append(stored, Pair{GraphID: graphID, ArtifactID: id})
			}
		}
	}
	linked, err := applyLinks(tx, userID, stored, want)
	if err != nil {
		return nil, err
	}
	return append(changes, linked...), nil
}

// getTx loads one graph row; sql.ErrNoRows when there is none. It does not fill
// ArtifactIDs.
func getTx(tx *sql.Tx, id []byte) (Graph, error) {
	return scanGraph(tx.QueryRow(sqlGet, id))
}

// requireDistinctPairs refuses a repeated pair.
func requireDistinctPairs(pairs []Pair) error {
	seen := map[string]bool{}
	for _, p := range pairs {
		k := string(p.GraphID) + string(p.ArtifactID)
		if seen[k] {
			return ErrInvalid
		}
		seen[k] = true
	}
	return nil
}

// applyLinks is the one link writer: it links every pair of want that is not
// in stored, in want order, then unlinks every pair of stored that is not in
// want, in stored order. Links come first, so a swap never leaves an Artifact
// or a graph with no link between the two. The first refusal is returned
// unchanged; the caller's write rolls everything back.
func applyLinks(tx *database.Tx, userID []byte, stored, want []Pair) ([]rowchange.Change, error) {
	if err := requireDistinctPairs(want); err != nil {
		return nil, err
	}
	var changes []rowchange.Change
	for _, p := range want {
		if containsPair(stored, p) {
			continue
		}
		ch, err := Link(tx, userID, p.GraphID, p.ArtifactID)
		if err != nil {
			return nil, err
		}
		changes = append(changes, ch...)
	}
	for _, p := range stored {
		if containsPair(want, p) {
			continue
		}
		ch, err := Unlink(tx, userID, p.GraphID, p.ArtifactID)
		if err != nil {
			return nil, err
		}
		changes = append(changes, ch...)
	}
	return changes, nil
}

func containsPair(pairs []Pair, p Pair) bool {
	for _, q := range pairs {
		if bytes.Equal(q.GraphID, p.GraphID) && bytes.Equal(q.ArtifactID, p.ArtifactID) {
			return true
		}
	}
	return false
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

// SourceOf returns the Source a graph belongs to. ErrInvalid for a nil
// querier, a malformed id or no such graph.
func SourceOf(q database.Querier, graphID []byte) ([]byte, error) {
	if q == nil || len(graphID) != 16 {
		return nil, ErrInvalid
	}
	var id []byte
	err := q.QueryRow(sqlGraphSource, graphID).Scan(&id)
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

// CitationsUsingArtifactTx returns the graph's Citations on the Artifact, with
// or without Observations, with their refs, ordered by ref. Only ID and Ref are
// set.
func CitationsUsingArtifactTx(tx *sql.Tx, graphID, artifactID []byte) ([]deleteimpact.Listed, error) {
	if tx == nil || len(graphID) != 16 || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	return queryCitationRows(tx, sqlCitationsUsing, artifactID, graphID)
}

// queryCitationRows runs a query selecting Citation id and ref and returns the
// rows as Listed, in the query's order.
func queryCitationRows(tx *sql.Tx, q string, args ...any) ([]deleteimpact.Listed, error) {
	rows, err := tx.Query(q, args...)
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

// citationRowsExist reports whether a query selecting Citation id and ref
// returns any row.
func citationRowsExist(tx *sql.Tx, q string, args ...any) (bool, error) {
	var found bool
	err := tx.QueryRow("SELECT EXISTS("+q+")", args...).Scan(&found)
	return found, err
}

// UnlinkVia names the preflight group of Citations that keep an Artifact
// linked. A domain-preflight via is named for its refusal code, not an FK edge.
const UnlinkVia = apperr.CodeEvidenceGraphsArtifactInUse

// UnlinkBlock is one pair whose unlink is refused, with the refusal and, for a
// Citation reason, those Citations shaped for the DeleteImpact recipe.
type UnlinkBlock struct {
	GraphID, ArtifactID []byte
	Reason              error // ErrArtifactInUse or ErrLastArtifact
	Citations           deleteimpact.Group
}

// UnlinkImpact reports which of the (graph, Artifact) pairs an edit would
// unlink are refused, in the order given, each with its reason. A pair that is
// not linked, or whose unlink is allowed, has no block. Pairs may span graphs
// and Artifacts.
func UnlinkImpact(c *database.Catalog, pairs []Pair) ([]UnlinkBlock, error) {
	for _, p := range pairs {
		if len(p.GraphID) != 16 || len(p.ArtifactID) != 16 {
			return nil, ErrInvalid
		}
	}
	if err := requireDistinctPairs(pairs); err != nil {
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
	var out []UnlinkBlock
	for _, p := range pairs {
		if _, err := getTx(tx, p.GraphID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrInvalid
			}
			return nil, err
		}
		block, rows, err := checkUnlink(tx, p, true)
		if err != nil {
			return nil, err
		}
		if block == nil {
			continue
		}
		if len(rows) > 0 {
			block.Citations, err = deleteimpact.ListedGroup(tx, apperr.From(block.Reason).Code(), catalogmodel.KindCitation, rows)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, *block)
	}
	return out, nil
}

// checkUnlink is the one unlink check behind Unlink and UnlinkImpact: nil when
// the pair is not linked or may be unlinked. Otherwise the first reason, in
// order: the graph has a Citation on the Artifact (ErrArtifactInUse), or it is
// the graph's only link (ErrLastArtifact). The rows are the Citations behind the
// first reason; they are listed only when list is true, since Unlink needs just
// the reason and UnlinkImpact the rows.
func checkUnlink(tx *sql.Tx, p Pair, list bool) (*UnlinkBlock, []deleteimpact.Listed, error) {
	linked, err := IsLinkedTx(tx, p.GraphID, p.ArtifactID)
	if err != nil {
		return nil, nil, err
	}
	if !linked {
		return nil, nil, nil
	}
	block := func(reason error) *UnlinkBlock {
		return &UnlinkBlock{
			GraphID:    append([]byte(nil), p.GraphID...),
			ArtifactID: append([]byte(nil), p.ArtifactID...),
			Reason:     reason,
		}
	}
	if !list {
		used, err := citationRowsExist(tx, sqlCitationsUsing, p.ArtifactID, p.GraphID)
		if err != nil {
			return nil, nil, err
		}
		if used {
			return block(ErrArtifactInUse), nil, nil
		}
	} else {
		using, err := CitationsUsingArtifactTx(tx, p.GraphID, p.ArtifactID)
		if err != nil {
			return nil, nil, err
		}
		if len(using) > 0 {
			return block(ErrArtifactInUse), using, nil
		}
	}
	var n int
	if err := tx.QueryRow(sqlLinkCount, p.GraphID).Scan(&n); err != nil {
		return nil, nil, err
	}
	if n == 1 {
		return block(ErrLastArtifact), nil, nil
	}
	return nil, nil, nil
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
// changes. checkUnlink decides: the graph's Citations keep the Artifact linked
// (ErrArtifactInUse), and a graph keeps at least one Artifact (ErrLastArtifact).
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
	block, _, err := checkUnlink(tx.Tx, Pair{GraphID: graphID, ArtifactID: artifactID}, false)
	if err != nil {
		return nil, err
	}
	if block != nil {
		return nil, block.Reason
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

// LinkArtifact links a new Artifact to the graphs of its Source that links
// names: every graph when links is not Given, exactly links.IDs when it is,
// and none when it is Given with no ids. A repeated id, or a graph of another
// Source, is ErrInvalid. A Source with no graph (only direct-SQL fixtures)
// gets no link.
func LinkArtifact(tx *database.Tx, userID, sourceID, artifactID []byte, links GraphLinks) ([]rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	ids := links.IDs
	if !links.Given {
		graphs, err := list(tx.Tx, sourceID)
		if err != nil {
			return nil, err
		}
		ids = nil
		for _, g := range graphs {
			ids = append(ids, g.ID)
		}
	}
	want := make([]Pair, len(ids))
	for i, id := range ids {
		want[i] = Pair{GraphID: id, ArtifactID: artifactID}
	}
	return applyLinks(tx, userID, nil, want)
}

// SetArtifactGraphs makes graphIDs the full set of graphs an existing
// Artifact is linked to, in one write: it links first, then unlinks, so a swap
// is allowed, and every refusal is Unlink's. It does not check that the
// Artifact exists; its caller has read the row, and every pair it links goes
// through Link's checks.
func SetArtifactGraphs(tx *database.Tx, userID, artifactID []byte, graphIDs [][]byte) ([]rowchange.Change, error) {
	if tx == nil || len(artifactID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := tx.Query(sqlArtifactGraphs, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stored []Pair
	for rows.Next() {
		var graphID []byte
		if err := rows.Scan(&graphID); err != nil {
			return nil, err
		}
		stored = append(stored, Pair{GraphID: graphID, ArtifactID: artifactID})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	want := make([]Pair, len(graphIDs))
	for i, id := range graphIDs {
		want[i] = Pair{GraphID: id, ArtifactID: artifactID}
	}
	return applyLinks(tx, userID, stored, want)
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
