package deleteimpact

import (
	"database/sql"
	"strings"
)

func projectListed(tx *sql.Tx, kind Kind, row probeRow) (Listed, error) {
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

func projectKind(tx *sql.Tx, kind Kind, row probeRow) (string, Location, error) {
	switch kind {
	case KindObservation:
		return projectObservation(tx, row.ID, row.Ref)
	case KindSubject:
		return projectSubject(tx, row.ID, row.Ref)
	case KindCitation:
		return projectCitation(tx, row.ID, row.Ref)
	case KindArtifact:
		return projectArtifact(tx, row.ID, row.Ref)
	case KindSource:
		return projectSource(tx, row.ID, row.Ref)
	case KindSourceType:
		return projectVocab(tx, `SELECT label FROM source_types WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionSourceTypes, TypeID: uuidString(row.ID),
		})
	case KindSourceField:
		return projectVocab(tx, `SELECT label FROM source_metadata_fields WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionSourceFields, FieldID: uuidString(row.ID),
		})
	case KindProperty, KindPropertyTerm, KindSubjectType:
		sqlLabel := `SELECT label FROM properties WHERE id = ?`
		if kind == KindPropertyTerm {
			sqlLabel = `SELECT label FROM property_terms WHERE id = ?`
		} else if kind == KindSubjectType {
			sqlLabel = `SELECT label FROM subject_types WHERE id = ?`
		}
		return projectVocab(tx, sqlLabel, row.ID, row.Ref, Location{Section: sectionSubjectFields})
	case KindCredibilityGrade:
		return projectVocab(tx, `SELECT label FROM source_credibility_grades WHERE id = ?`, row.ID, row.Ref, Location{
			Section: sectionSources,
		})
	default:
		return row.Ref, Location{Ref: row.Ref, Title: row.Ref}, nil
	}
}

func projectObservation(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var (
		propLabel, value, subLabel, srcTitle, srcID, citID, subID string
		propLabelN, valueN, subLabelN, srcTitleN                  sql.NullString
		srcIDB, citIDB, subIDB                                    []byte
	)
	err := tx.QueryRow(`
		SELECT p.label,
			COALESCE(NULLIF(pt.label, ''), NULLIF(o.value_text, ''), nv.form, vs.label,
				CASE WHEN o.value_integer IS NOT NULL THEN CAST(o.value_integer AS TEXT) END,
				NULLIF(TRIM(dv.phrase), ''),
				CASE WHEN dv.start_year IS NOT NULL THEN CAST(dv.start_year AS TEXT) END,
				''),
			sub.label, src.title, src.id, o.citation_id, o.subject_id
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
	).Scan(&propLabelN, &valueN, &subLabelN, &srcTitleN, &srcIDB, &citIDB, &subIDB)
	if err == sql.ErrNoRows {
		return ref, Location{
			Section: sectionSources, SourceSurface: surfaceCitationComposer,
			ObservationID: uuidString(id), Ref: ref, Title: ref,
		}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	propLabel = propLabelN.String
	value = strings.TrimSpace(valueN.String)
	subLabel = subLabelN.String
	srcTitle = srcTitleN.String
	srcID = uuidString(srcIDB)
	citID = uuidString(citIDB)
	subID = uuidString(subIDB)
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
		SubjectID:     subID,
		CitationID:    citID,
		ObservationID: uuidString(id),
		SourceSurface: surfaceCitationComposer,
		Ref:           ref,
		Title:         subLabel,
		SourceTitle:   srcTitle,
	}, nil
}

func projectSubject(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var label string
	var srcID []byte
	err := tx.QueryRow(`SELECT COALESCE(label, ''), source_id FROM subjects WHERE id = ?`, id).
		Scan(&label, &srcID)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfaceGraph, SubjectID: uuidString(id), Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	title := strings.TrimSpace(label)
	if title == "" {
		title = ref
	}
	return title, Location{
		Section:       sectionSources,
		SourceID:      uuidString(srcID),
		SubjectID:     uuidString(id),
		SourceSurface: surfaceGraph,
		Ref:           ref,
		Title:         title,
	}, nil
}

func projectCitation(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var srcID, artID []byte
	var srcTitle sql.NullString
	err := tx.QueryRow(`
		SELECT a.source_id, c.artifact_id, src.title
		FROM citations c
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		WHERE c.id = ?`, id,
	).Scan(&srcID, &artID, &srcTitle)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfaceCitationComposer, CitationID: uuidString(id), Ref: ref, Title: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	return ref, Location{
		Section:       sectionSources,
		SourceID:      uuidString(srcID),
		CitationID:    uuidString(id),
		ArtifactID:    uuidString(artID),
		SourceSurface: surfaceCitationComposer,
		Ref:           ref,
		Title:         ref,
		SourceTitle:   srcTitle.String,
	}, nil
}

func projectArtifact(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var label sql.NullString
	var srcID []byte
	err := tx.QueryRow(`SELECT label, source_id FROM artifacts WHERE id = ?`, id).Scan(&label, &srcID)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfacePage, ArtifactID: uuidString(id), Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	title := strings.TrimSpace(label.String)
	if title == "" {
		title = ref
	}
	return title, Location{
		Section:       sectionSources,
		SourceID:      uuidString(srcID),
		ArtifactID:    uuidString(id),
		SourceSurface: surfacePage,
		Ref:           ref,
		Title:         title,
	}, nil
}

func projectSource(tx *sql.Tx, id []byte, ref string) (string, Location, error) {
	var title sql.NullString
	err := tx.QueryRow(`SELECT title FROM sources WHERE id = ?`, id).Scan(&title)
	if err == sql.ErrNoRows {
		return ref, Location{Section: sectionSources, SourceSurface: surfacePage, SourceID: uuidString(id), Ref: ref}, nil
	}
	if err != nil {
		return "", Location{}, err
	}
	label := strings.TrimSpace(title.String)
	if label == "" {
		label = ref
	}
	return label, Location{
		Section:       sectionSources,
		SourceID:      uuidString(id),
		SourceSurface: surfacePage,
		Ref:           ref,
		Title:         label,
	}, nil
}

func projectVocab(tx *sql.Tx, q string, id []byte, ref string, loc Location) (string, Location, error) {
	var label sql.NullString
	err := tx.QueryRow(q, id).Scan(&label)
	if err != nil && err != sql.ErrNoRows {
		return "", Location{}, err
	}
	title := strings.TrimSpace(label.String)
	if title == "" {
		title = ref
	}
	loc.Ref = ref
	loc.Title = title
	return title, loc, nil
}
