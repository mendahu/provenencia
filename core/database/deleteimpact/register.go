package deleteimpact

import (
	"database/sql"
	"sync"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

type inboundEdge struct {
	Parent catalogmodel.Kind
	Via    string
	Child  catalogmodel.Kind
	Bucket catalogmodel.Bucket
	Count  func(tx *sql.Tx, parentID []byte) (int, error)
	List   func(tx *sql.Tx, parentID []byte, limit int) ([]probeRow, error)
}

type ownedRelease struct {
	Parent catalogmodel.Kind
	Column string
	Child  string
	Mode   string // always | ifUnused
	Then   []ownedRelease
}

type originRule struct {
	Kind         catalogmodel.Kind
	OriginSQL    string
	PluginNever  bool
	SeededLocked bool
}

type probeRow struct {
	ID  []byte
	Ref string
}

// existsSQL is how Impact checks that a deletable row is present. Tables
// without an entry have no official delete lookup.
var existsSQL = map[string]string{
	"sources":                   `SELECT 1 FROM sources WHERE id = ?`,
	"artifacts":                 `SELECT 1 FROM artifacts WHERE id = ?`,
	"citations":                 `SELECT 1 FROM citations WHERE id = ?`,
	"observations":              `SELECT 1 FROM observations WHERE id = ?`,
	"subjects":                  `SELECT 1 FROM subjects WHERE id = ?`,
	"source_types":              `SELECT 1 FROM source_types WHERE id = ?`,
	"source_metadata_fields":    `SELECT 1 FROM source_metadata_fields WHERE id = ?`,
	"source_credibility_grades": `SELECT 1 FROM source_credibility_grades WHERE id = ?`,
	"subject_types":             `SELECT 1 FROM subject_types WHERE id = ?`,
	"properties":                `SELECT 1 FROM properties WHERE id = ?`,
	"property_terms":            `SELECT 1 FROM property_terms WHERE id = ?`,
	"files":                     `SELECT 1 FROM files WHERE id = ?`,
	"users":                     `SELECT 1 FROM users WHERE id = ?`,
	"project":                   `SELECT 1 FROM project WHERE id = ?`,
}

type kindTable struct {
	catalogmodel.Table
	Exists string
}

var originRules = []originRule{
	{Kind: catalogmodel.KindSourceType, OriginSQL: `SELECT origin FROM source_types WHERE id = ?`, PluginNever: true},
	{Kind: catalogmodel.KindMetadataField, OriginSQL: `SELECT origin FROM source_metadata_fields WHERE id = ?`, PluginNever: true},
	{Kind: catalogmodel.KindProperty, OriginSQL: `SELECT origin FROM properties WHERE id = ?`, PluginNever: true, SeededLocked: true},
	{Kind: catalogmodel.KindPropertyTerm, OriginSQL: `SELECT origin FROM property_terms WHERE id = ?`, PluginNever: true, SeededLocked: true},
	{Kind: catalogmodel.KindSubjectType, OriginSQL: `SELECT origin FROM subject_types WHERE id = ?`, PluginNever: true},
	{Kind: catalogmodel.KindCredibilityGrade, OriginSQL: `SELECT origin FROM source_credibility_grades WHERE id = ?`, PluginNever: true},
}

var ownedReleases = []ownedRelease{
	{Parent: catalogmodel.KindObservation, Column: "observations.value_date_id", Child: "date_values", Mode: "always"},
	{Parent: catalogmodel.KindObservation, Column: "observations.value_name_id", Child: "name_values", Mode: "always"},
	{
		Parent: catalogmodel.KindArtifact, Column: "artifacts.file_id", Child: "files", Mode: "ifUnused",
		Then: []ownedRelease{
			{Parent: catalogmodel.KindFile, Column: "file_derivatives.source_file_id", Child: "file_derivatives", Mode: "always"},
			{Parent: catalogmodel.KindFile, Column: "file_derivatives.derived_file_id", Child: "file_derivatives", Mode: "always"},
		},
	},
}

func tableByKind(kind catalogmodel.Kind) (kindTable, bool) {
	for _, t := range catalogmodel.Tables {
		exists, ok := existsSQL[t.Name]
		if t.Kind == kind && ok {
			return kindTable{Table: t, Exists: exists}, true
		}
	}
	return kindTable{}, false
}

func inboundFor(kind catalogmodel.Kind) []inboundEdge {
	var out []inboundEdge
	for _, e := range inboundEdges() {
		if e.Parent == kind {
			out = append(out, e)
		}
	}
	return out
}

func inboundEdges() []inboundEdge {
	return inboundEdgeList()
}

var inboundEdgeList = sync.OnceValue(func() []inboundEdge {
	return []inboundEdge{
		fkEdge(catalogmodel.KindSource, "artifacts.source_id", catalogmodel.KindArtifact, "artifacts", "source_id", "ref"),
		fkEdge(catalogmodel.KindSource, "subjects.source_id", catalogmodel.KindSubject, "subjects", "source_id", "ref"),
		fkEdge(catalogmodel.KindArtifact, "citations.artifact_id", catalogmodel.KindCitation, "citations", "artifact_id", "ref"),
		fkEdge(catalogmodel.KindCitation, "observations.citation_id", catalogmodel.KindObservation, "observations", "citation_id", "ref"),
		subjectResourceInbound(),
		fkEdge(catalogmodel.KindSubject, "observations.value_subject_id", catalogmodel.KindObservation, "observations", "value_subject_id", "ref"),
		fkEdge(catalogmodel.KindSourceType, "sources.source_type_id", catalogmodel.KindSource, "sources", "source_type_id", "ref"),
		joinEdge(catalogmodel.KindMetadataField, "source_metadata.field_id", catalogmodel.KindSource,
			`SELECT COUNT(DISTINCT source_id) FROM source_metadata WHERE field_id = ?`,
			`SELECT s.id, s.ref FROM source_metadata m
				JOIN sources s ON s.id = m.source_id
				WHERE m.field_id = ?
				ORDER BY s.ref COLLATE NOCASE LIMIT ?`),
		joinEdge(catalogmodel.KindCredibilityGrade, "source_credibility_assessments.credibility_grade_id", catalogmodel.KindSource,
			`SELECT COUNT(DISTINCT source_id) FROM source_credibility_assessments WHERE credibility_grade_id = ?`,
			`SELECT s.id, s.ref FROM source_credibility_assessments a
				JOIN sources s ON s.id = a.source_id
				WHERE a.credibility_grade_id = ?
				ORDER BY s.ref COLLATE NOCASE LIMIT ?`),
		fkEdge(catalogmodel.KindSubjectType, "subjects.subject_type_id", catalogmodel.KindSubject, "subjects", "subject_type_id", "ref"),
		fkEdge(catalogmodel.KindProperty, "observations.property_id", catalogmodel.KindObservation, "observations", "property_id", "ref"),
		joinEdge(catalogmodel.KindProperty, "property_terms.property_id", catalogmodel.KindPropertyTerm,
			`SELECT COUNT(*) FROM property_terms WHERE property_id = ?`,
			`SELECT id, key FROM property_terms WHERE property_id = ? ORDER BY key COLLATE NOCASE LIMIT ?`),
		fkEdge(catalogmodel.KindPropertyTerm, "observations.value_term_id", catalogmodel.KindObservation, "observations", "value_term_id", "ref"),
		fkEdge(catalogmodel.KindFile, "artifacts.file_id", catalogmodel.KindArtifact, "artifacts", "file_id", "ref"),
		fkEdge(catalogmodel.KindSubjectType, "canonical_entities.subject_type_id", catalogmodel.KindCanonicalEntity, "canonical_entities", "subject_type_id", "ref"),
		fkEdge(catalogmodel.KindCanonicalEntity, "canonical_entities.merged_into_id", catalogmodel.KindCanonicalEntity, "canonical_entities", "merged_into_id", "ref"),
		// Claims have no ref; name the handle each claim files onto.
		joinEdge(catalogmodel.KindClaimConfidenceGrade, "identity_claims.confidence_grade_id", catalogmodel.KindCanonicalEntity,
			`SELECT COUNT(DISTINCT entity_id) FROM identity_claims WHERE confidence_grade_id = ?`,
			`SELECT DISTINCT e.id, e.ref FROM identity_claims ic
				JOIN canonical_entities e ON e.id = ic.entity_id
				WHERE ic.confidence_grade_id = ?
				ORDER BY e.ref COLLATE NOCASE LIMIT ?`),
	}
})

func originRuleFor(kind catalogmodel.Kind) (originRule, bool) {
	for _, r := range originRules {
		if r.Kind == kind {
			return r, true
		}
	}
	return originRule{}, false
}

func ownedFor(kind catalogmodel.Kind) []ownedRelease {
	var out []ownedRelease
	for _, r := range ownedReleases {
		if r.Parent == kind {
			out = append(out, r)
		}
	}
	return out
}

func resourceInboundVias() map[string]bool {
	out := map[string]bool{}
	for _, e := range inboundEdges() {
		if e.Bucket == catalogmodel.BucketResource && e.List != nil {
			out[e.Via] = true
		}
	}
	return out
}

func ownedColumns() map[string]bool {
	out := map[string]bool{}
	var walk func([]ownedRelease)
	walk = func(rs []ownedRelease) {
		for _, r := range rs {
			out[r.Column] = true
			walk(r.Then)
		}
	}
	walk(ownedReleases)
	return out
}
