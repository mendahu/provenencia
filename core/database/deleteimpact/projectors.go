package deleteimpact

import (
	"database/sql"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

const listedSnippetMaxRunes = 80

func projectListed(tx *sql.Tx, kind catalogmodel.Kind, row probeRow) (Listed, error) {
	out := Listed{ID: row.ID, Ref: row.Ref}
	title, loc, err := projectKind(tx, kind, row)
	if err != nil {
		return Listed{}, err
	}
	out.Title = title
	out.Location = loc
	if out.Location.Ref == "" {
		out.Location.Ref = row.Ref
	}
	if out.Location.Title == "" {
		out.Location.Title = title
	}
	return out, nil
}

func projectKind(tx *sql.Tx, kind catalogmodel.Kind, row probeRow) (string, Location, error) {
	switch kind {
	case catalogmodel.KindObservation:
		return projectObservation(tx, row.ID, row.Ref)
	case catalogmodel.KindSubject:
		return projectSubject(tx, row.ID, row.Ref)
	case catalogmodel.KindCitation:
		return projectCitation(tx, row.ID, row.Ref)
	case catalogmodel.KindArtifact:
		return projectArtifact(tx, row.ID, row.Ref)
	case catalogmodel.KindSource:
		return projectSource(tx, row.ID, row.Ref)
	case catalogmodel.KindEvidenceGraph:
		return projectEvidenceGraph(tx, row.ID, row.Ref)
	case catalogmodel.KindSourceType:
		return projectVocab(tx, `SELECT label FROM source_types WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionSourceTypes, TypeID: database.UUIDString(row.ID),
		})
	case catalogmodel.KindMetadataField:
		return projectVocab(tx, `SELECT label FROM source_metadata_fields WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionMetadata, FieldID: database.UUIDString(row.ID),
		})
	case catalogmodel.KindProperty:
		return projectVocab(tx, `SELECT label FROM properties WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionProperties, PropertyID: database.UUIDString(row.ID),
		})
	case catalogmodel.KindPropertyTerm:
		return projectPropertyTerm(tx, row.ID, row.Ref)
	case catalogmodel.KindSubjectType:
		return projectSubjectType(tx, row.ID, row.Ref)
	case catalogmodel.KindCredibilityGrade:
		return projectVocab(tx, `SELECT label FROM source_credibility_grades WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionSources,
		})
	default:
		return row.Ref, Location{Ref: row.Ref, Title: row.Ref}, nil
	}
}

// sqlFirstArtifactGraph is the first graph, by name, an Artifact is linked to.
// The order is the graph list's (evidencegraphs.sqlListOrder): name without
// case, then ref, then id.
const sqlFirstArtifactGraph = `SELECT ega.graph_id FROM evidence_graph_artifacts ega
	JOIN evidence_graphs g ON g.id = ega.graph_id
	WHERE ega.artifact_id = ?
	ORDER BY g.name COLLATE NOCASE, g.ref, g.id
	LIMIT 1`

func projectObservation(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var (
		propLabel, value, subLabel, srcTitle, srcID, citID, subID, graphID string
		propLabelN, valueN, subLabelN, srcTitleN                           sql.NullString
		subRef                                                             string
		srcIDB, citIDB, subIDB, graphIDB                                   []byte
	)
	err := tx.QueryRow(`
		SELECT p.label,
			COALESCE(NULLIF(pt.label, ''), NULLIF(o.value_text, ''), nv.form, vs.label,
				CASE WHEN o.value_integer IS NOT NULL THEN CAST(o.value_integer AS TEXT) END,
				NULLIF(TRIM(dv.phrase), ''),
				CASE WHEN dv.start_year IS NOT NULL THEN CAST(dv.start_year AS TEXT) END,
				''),
			sub.label, sub.ref, src.title, src.id, o.citation_id, o.subject_id, sub.graph_id
		FROM observations o
		JOIN properties p ON p.id = o.property_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		JOIN subjects sub ON sub.id = o.subject_id
		LEFT JOIN property_terms pt ON pt.id = o.value_term_id
		LEFT JOIN name_values nv ON nv.id = o.value_name_id
		LEFT JOIN subjects vs ON vs.id = o.value_subject_id
		LEFT JOIN date_values dv ON dv.id = o.value_date_id
		WHERE o.id = ?`, id,
	).Scan(&propLabelN, &valueN, &subLabelN, &subRef, &srcTitleN, &srcIDB, &citIDB, &subIDB, &graphIDB)
	if err == sql.ErrNoRows {
		return ref, Location{
			Section: sectionSources, SourceSurface: surfaceCitationComposer,
			ObservationID: database.UUIDString(id), Ref: ref, Title: ref,
		}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	propLabel = propLabelN.String
	value = strings.TrimSpace(valueN.String)
	subLabel = subLabelN.String
	srcTitle = srcTitleN.String
	srcID = database.UUIDString(srcIDB)
	citID = database.UUIDString(citIDB)
	subID = database.UUIDString(subIDB)
	graphID = database.UUIDString(graphIDB)
	title := propLabel
	if value != "" {
		title = propLabel + ": " + value
	}
	if title == "" {
		title = ref
	}
	return title, Location{
		Section:       sectionSources,
		SourceID:      srcID,
		GraphID:       graphID,
		SubjectID:     subID,
		CitationID:    citID,
		ObservationID: database.UUIDString(id),
		SourceSurface: surfaceCitationComposer,
		Ref:           subRef,
		Title:         titleOrRef(subLabel, subRef),
		SourceTitle:   srcTitle,
	}, nil
}

// projectSubject lists the Subject by its label and opens its graph's canvas
// with the Subject selected. The location is a graph location, so it names
// the Source (graphLocation).
func projectSubject(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var label, srcRef, srcTitle sql.NullString
	var srcID, graphID []byte
	err := tx.QueryRow(`
		SELECT sub.label, sub.source_id, sub.graph_id, src.ref, src.title
		FROM subjects sub
		JOIN sources src ON src.id = sub.source_id
		WHERE sub.id = ?`, id,
	).Scan(&label, &srcID, &graphID, &srcRef, &srcTitle)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfaceGraph, SubjectID: database.UUIDString(id), Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	loc := graphLocation(srcID, graphID, srcRef.String, srcTitle.String)
	loc.SubjectID = database.UUIDString(id)
	return titleOrRef(label.String, ref), loc, nil
}

func projectCitation(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var srcID, artID []byte
	var srcTitle, transcription sql.NullString
	err := tx.QueryRow(`
		SELECT a.source_id, c.artifact_id, src.title, c.transcription
		FROM citations c
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		WHERE c.id = ?`, id,
	).Scan(&srcID, &artID, &srcTitle, &transcription)
	if err == sql.ErrNoRows {
		return "", Location{Section: sectionSources, SourceSurface: surfaceCitationComposer, CitationID: database.UUIDString(id), Ref: ref, Title: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	locRef, locTitle := ref, ref
	var subID, graphID []byte
	var subLabel, subRef sql.NullString
	obsErr := tx.QueryRow(`
		SELECT o.subject_id, sub.label, sub.ref, sub.graph_id
		FROM observations o
		JOIN subjects sub ON sub.id = o.subject_id
		WHERE o.citation_id = ?
		ORDER BY o.ref COLLATE NOCASE
		LIMIT 1`, id,
	).Scan(&subID, &subLabel, &subRef, &graphID)
	if obsErr != nil && obsErr != sql.ErrNoRows {
		return "", Location{}, obsErr
	}
	if obsErr == sql.ErrNoRows {
		// No Observation: the first graph the Artifact is linked to, so the
		// link still opens the composer.
		err := tx.QueryRow(sqlFirstArtifactGraph, artID).Scan(&graphID)
		if err != nil && err != sql.ErrNoRows {
			return "", Location{}, err
		}
	}
	if subRef.Valid {
		locRef, locTitle = subRef.String, titleOrRef(subLabel.String, subRef.String)
	}
	title := listedSnippet(transcription.String, listedSnippetMaxRunes)
	if title == "" {
		title = titleOrRef(subLabel.String, ref)
	}
	return title, Location{
		Section:       sectionSources,
		SourceID:      database.UUIDString(srcID),
		GraphID:       database.UUIDString(graphID),
		SubjectID:     database.UUIDString(subID),
		CitationID:    database.UUIDString(id),
		ArtifactID:    database.UUIDString(artID),
		SourceSurface: surfaceCitationComposer,
		Ref:           locRef,
		Title:         locTitle,
		SourceTitle:   srcTitle.String,
	}, nil
}

func listedSnippet(s string, maxRunes int) string {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" || maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}

func projectArtifact(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var label sql.NullString
	var srcID []byte
	err := tx.QueryRow(`SELECT label, source_id FROM artifacts WHERE id = ?`, id).Scan(&label, &srcID)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfacePage, ArtifactID: database.UUIDString(id), Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	title := titleOrRef(label.String, ref)
	return title, Location{
		Section:       sectionSources,
		SourceID:      database.UUIDString(srcID),
		ArtifactID:    database.UUIDString(id),
		SourceSurface: surfacePage,
		Ref:           ref,
		Title:         title,
	}, nil
}

func projectSource(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var title sql.NullString
	err := tx.QueryRow(`SELECT title FROM sources WHERE id = ?`, id).Scan(&title)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfacePage, SourceID: database.UUIDString(id), Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	label := titleOrRef(title.String, ref)
	return label, Location{
		Section:       sectionSources,
		SourceID:      database.UUIDString(id),
		SourceSurface: surfacePage,
		Ref:           ref,
		Title:         label,
	}, nil
}

// projectEvidenceGraph lists the graph by its name and opens its canvas. The
// location names the Source (graphLocation).
func projectEvidenceGraph(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var name string
	var srcRef, srcTitle sql.NullString
	var sourceID []byte
	err := tx.QueryRow(`
		SELECT g.name, g.source_id, src.ref, src.title
		FROM evidence_graphs g
		JOIN sources src ON src.id = g.source_id
		WHERE g.id = ?`, id,
	).Scan(&name, &sourceID, &srcRef, &srcTitle)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfacePage, Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	return titleOrRef(name, ref), graphLocation(sourceID, id, srcRef.String, srcTitle.String), nil
}

// graphLocation is a graph's canvas. A graph location's Ref and Title name
// the Source, as the app's do (the canvas breadcrumb falls back to Title for
// the Source crumb); the graph's own name comes from its id. A
// citationComposer location instead names its scope Subject in Ref and Title
// and carries the Source's title in SourceTitle.
func graphLocation(sourceID, graphID []byte, sourceRef, sourceTitle string) Location {
	return Location{
		Section:       sectionSources,
		SourceID:      database.UUIDString(sourceID),
		GraphID:       database.UUIDString(graphID),
		SourceSurface: surfaceGraph,
		Ref:           sourceRef,
		Title:         titleOrRef(sourceTitle, sourceRef),
		SourceTitle:   sourceTitle,
	}
}

// projectPropertyTerm deep-links a term to its parent property's Properties inspector row.
func projectPropertyTerm(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var label sql.NullString
	var propID []byte
	err := tx.QueryRow(`SELECT label, property_id FROM property_terms WHERE id = ?`, id).Scan(&label, &propID)
	if err != nil && err != sql.ErrNoRows {
		return "", Location{}, err
	}
	return titleOrRef(label.String, ref), Location{
		Section:    sectionProperties,
		PropertyID: database.UUIDString(propID),
		Ref:        ref,
		Title:      titleOrRef(label.String, ref),
	}, nil
}

// projectSubjectType deep-links a type to its Properties type-strip category.
func projectSubjectType(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var label, key sql.NullString
	err := tx.QueryRow(`SELECT label, key FROM subject_types WHERE id = ?`, id).Scan(&label, &key)
	if err != nil && err != sql.ErrNoRows {
		return "", Location{}, err
	}
	return titleOrRef(label.String, ref), Location{
		Section:        sectionProperties,
		SubjectTypeKey: key.String,
		Ref:            ref,
		Title:          titleOrRef(label.String, ref),
	}, nil
}

func titleOrRef(label, ref string) string {
	if title := strings.TrimSpace(label); title != "" {
		return title
	}
	return ref
}

func projectVocab(tx *sql.Tx, q string, id []byte, ref string, loc Location) (string, Location, error) {
	var label sql.NullString
	err := tx.QueryRow(q, id).Scan(&label)
	if err != nil && err != sql.ErrNoRows {
		return "", Location{}, err
	}
	title := titleOrRef(label.String, ref)
	loc.Ref = ref
	loc.Title = title
	return title, loc, nil
}
