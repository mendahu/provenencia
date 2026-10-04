package deleteimpact

import (
	"database/sql"
	"sync"
)

type tableSpec struct {
	Name   string
	PK     string
	Kind   Kind
	Bucket Bucket
	Exists string
}

type fkSpec struct {
	FromTable string
	FromCol   string   // leading column; the honesty key
	FromCols  []string // full column list for a composite FK (nil = FromCol only)
	// Audited CASCADE facets are research rows: official deletes remove and
	// audit them through a facetRelease, and the CASCADE is only a backstop.
	// Unaudited CASCADEs (layout, vocab joins, value parts) stay silent.
	Audited  bool
	ToTable  string
	OnDelete string
	Bucket   Bucket
}

type inboundEdge struct {
	Parent Kind
	Via    string
	Child  Kind
	Bucket Bucket
	Count  func(tx *sql.Tx, parentID []byte) (int, error)
	List   func(tx *sql.Tx, parentID []byte, limit int) ([]probeRow, error)
}

type ownedRelease struct {
	Parent Kind
	Column string
	Child  string
	Mode   string // always | ifUnused
	Then   []ownedRelease
}

type originRule struct {
	Kind         Kind
	OriginSQL    string
	PluginNever  bool
	SeededLocked bool
}

type probeRow struct {
	ID  []byte
	Ref string
}

var tables = []tableSpec{
	{Name: "sources", PK: "id", Kind: KindSource, Bucket: BucketResource, Exists: `SELECT 1 FROM sources WHERE id = ?`},
	{Name: "artifacts", PK: "id", Kind: KindArtifact, Bucket: BucketResource, Exists: `SELECT 1 FROM artifacts WHERE id = ?`},
	{Name: "citations", PK: "id", Kind: KindCitation, Bucket: BucketResource, Exists: `SELECT 1 FROM citations WHERE id = ?`},
	{Name: "observations", PK: "id", Kind: KindObservation, Bucket: BucketResource, Exists: `SELECT 1 FROM observations WHERE id = ?`},
	{Name: "subjects", PK: "id", Kind: KindSubject, Bucket: BucketResource, Exists: `SELECT 1 FROM subjects WHERE id = ?`},
	{Name: "source_types", PK: "id", Kind: KindSourceType, Bucket: BucketVocab, Exists: `SELECT 1 FROM source_types WHERE id = ?`},
	{Name: "source_metadata_fields", PK: "id", Kind: KindMetadataField, Bucket: BucketVocab, Exists: `SELECT 1 FROM source_metadata_fields WHERE id = ?`},
	{Name: "source_credibility_grades", PK: "id", Kind: KindCredibilityGrade, Bucket: BucketVocab, Exists: `SELECT 1 FROM source_credibility_grades WHERE id = ?`},
	{Name: "subject_types", PK: "id", Kind: KindSubjectType, Bucket: BucketVocab, Exists: `SELECT 1 FROM subject_types WHERE id = ?`},
	{Name: "properties", PK: "id", Kind: KindProperty, Bucket: BucketVocab, Exists: `SELECT 1 FROM properties WHERE id = ?`},
	{Name: "property_terms", PK: "id", Kind: KindPropertyTerm, Bucket: BucketVocab, Exists: `SELECT 1 FROM property_terms WHERE id = ?`},
	{Name: "canonical_entities", Kind: KindCanonicalEntity, Bucket: BucketResource},
	{Name: "claim_confidence_grades", Bucket: BucketVocab},
	{Name: "files", PK: "id", Kind: KindFile, Bucket: BucketPool, Exists: `SELECT 1 FROM files WHERE id = ?`},
	{Name: "users", PK: "id", Kind: KindUser, Bucket: BucketInfra, Exists: `SELECT 1 FROM users WHERE id = ?`},
	{Name: "project", PK: "id", Kind: KindProject, Bucket: BucketSkip, Exists: `SELECT 1 FROM project WHERE id = ?`},
	{Name: "source_notes", Bucket: BucketFacet},
	{Name: "source_metadata", Bucket: BucketFacet},
	{Name: "source_credibility_assessments", Bucket: BucketFacet},
	{Name: "source_metadata_layout", Bucket: BucketFacet},
	{Name: "source_type_metadata_fields", Bucket: BucketFacet},
	{Name: "citation_notes", Bucket: BucketFacet},
	{Name: "observation_notes", Bucket: BucketFacet},
	{Name: "subject_positions", Bucket: BucketFacet},
	{Name: "subject_type_properties", Bucket: BucketFacet},
	{Name: "name_value_parts", Bucket: BucketFacet},
	{Name: "identity_claims", Kind: KindIdentityClaim, Bucket: BucketFacet},
	{Name: "identity_claim_evidence", Bucket: BucketFacet},
	{Name: "date_values", Bucket: BucketOwnedOutbound},
	{Name: "name_values", Bucket: BucketOwnedOutbound},
	{Name: "file_derivatives", Bucket: BucketOwnedOutbound},
	{Name: "audit_transactions", Bucket: BucketSkip},
	{Name: "audit_changes", Bucket: BucketSkip},
	{Name: "audit_transaction_scopes", Bucket: BucketSkip},
	{Name: "catalog_search_docs", Bucket: BucketSkip},
	{Name: "catalog_search_meta", Bucket: BucketSkip},
	{Name: "conclusion_resolved_values", Bucket: BucketSkip},
	{Name: "conclusion_resolved_meta", Bucket: BucketSkip},
}

// Every live catalog FK. Honesty tests compare this to PRAGMA foreign_key_list.
var foreignKeys = []fkSpec{
	{FromTable: "project", FromCol: "updated_by", ToTable: "users", OnDelete: "NO ACTION", Bucket: BucketSkip},
	{FromTable: "audit_transactions", FromCol: "user_id", ToTable: "users", OnDelete: "NO ACTION", Bucket: BucketSkip},
	{FromTable: "audit_changes", FromCol: "audit_transaction_id", ToTable: "audit_transactions", OnDelete: "NO ACTION", Bucket: BucketSkip},
	{FromTable: "audit_transaction_scopes", FromCol: "audit_transaction_id", ToTable: "audit_transactions", OnDelete: "NO ACTION", Bucket: BucketSkip},

	{FromTable: "sources", FromCol: "source_type_id", ToTable: "source_types", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "sources", FromCol: "primary_artifact_id", ToTable: "artifacts", OnDelete: "SET NULL", Bucket: BucketOptional},
	{FromTable: "source_notes", FromCol: "source_id", ToTable: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "source_metadata", FromCol: "source_id", ToTable: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "source_metadata", FromCol: "field_id", ToTable: "source_metadata_fields", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "source_credibility_assessments", FromCol: "source_id", ToTable: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "source_credibility_assessments", FromCol: "credibility_grade_id", ToTable: "source_credibility_grades", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "source_metadata_layout", FromCol: "source_id", ToTable: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "source_metadata_layout", FromCol: "field_id", ToTable: "source_metadata_fields", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "source_type_metadata_fields", FromCol: "source_type_id", ToTable: "source_types", OnDelete: "CASCADE", Bucket: BucketFacet},
	{FromTable: "source_type_metadata_fields", FromCol: "field_id", ToTable: "source_metadata_fields", OnDelete: "CASCADE", Bucket: BucketFacet},

	{FromTable: "artifacts", FromCol: "source_id", ToTable: "sources", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "artifacts", FromCol: "file_id", ToTable: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},

	{FromTable: "file_derivatives", FromCol: "source_file_id", ToTable: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},
	{FromTable: "file_derivatives", FromCol: "derived_file_id", ToTable: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},

	{FromTable: "citations", FromCol: "artifact_id", ToTable: "artifacts", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "citation_notes", FromCol: "citation_id", ToTable: "citations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},

	{FromTable: "observations", FromCol: "citation_id", ToTable: "citations", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "observations", FromCol: "subject_id", ToTable: "subjects", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "observations", FromCol: "property_id", ToTable: "properties", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "observations", FromCol: "value_date_id", ToTable: "date_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},
	{FromTable: "observations", FromCol: "value_name_id", ToTable: "name_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},
	{FromTable: "observations", FromCol: "value_subject_id", ToTable: "subjects", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "observations", FromCol: "value_term_id", ToTable: "property_terms", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "observation_notes", FromCol: "observation_id", ToTable: "observations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},

	{FromTable: "subjects", FromCol: "source_id", ToTable: "sources", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "subjects", FromCol: "subject_type_id", ToTable: "subject_types", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "subject_positions", FromCol: "subject_id", ToTable: "subjects", OnDelete: "CASCADE", Bucket: BucketFacet},
	{FromTable: "subject_type_properties", FromCol: "subject_type_id", ToTable: "subject_types", OnDelete: "CASCADE", Bucket: BucketFacet},
	{FromTable: "subject_type_properties", FromCol: "property_id", ToTable: "properties", OnDelete: "CASCADE", Bucket: BucketFacet},

	{FromTable: "property_terms", FromCol: "property_id", ToTable: "properties", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "name_value_parts", FromCol: "name_value_id", ToTable: "name_values", OnDelete: "CASCADE", Bucket: BucketFacet},

	{FromTable: "canonical_entities", FromCol: "subject_type_id", ToTable: "subject_types", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "canonical_entities", FromCol: "merged_into_id", ToTable: "canonical_entities", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "identity_claims", FromCol: "subject_id", FromCols: []string{"subject_id", "subject_type_id"}, ToTable: "subjects", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "identity_claims", FromCol: "entity_id", FromCols: []string{"entity_id", "subject_type_id"}, ToTable: "canonical_entities", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "identity_claims", FromCol: "confidence_grade_id", ToTable: "claim_confidence_grades", OnDelete: "NO ACTION", Bucket: BucketResource},
	{FromTable: "identity_claim_evidence", FromCol: "identity_claim_id", ToTable: "identity_claims", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{FromTable: "identity_claim_evidence", FromCol: "observation_id", ToTable: "observations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	// Derived resolved-values cache: silent backstops; upkeep rewrites the rows.
	{FromTable: "conclusion_resolved_values", FromCol: "entity_id", ToTable: "canonical_entities", OnDelete: "CASCADE", Bucket: BucketSkip},
	{FromTable: "conclusion_resolved_values", FromCol: "property_id", ToTable: "properties", OnDelete: "CASCADE", Bucket: BucketSkip},
}

var originRules = []originRule{
	{Kind: KindSourceType, OriginSQL: `SELECT origin FROM source_types WHERE id = ?`, PluginNever: true},
	{Kind: KindMetadataField, OriginSQL: `SELECT origin FROM source_metadata_fields WHERE id = ?`, PluginNever: true},
	{Kind: KindProperty, OriginSQL: `SELECT origin FROM properties WHERE id = ?`, PluginNever: true, SeededLocked: true},
	{Kind: KindPropertyTerm, OriginSQL: `SELECT origin FROM property_terms WHERE id = ?`, PluginNever: true, SeededLocked: true},
	{Kind: KindSubjectType, OriginSQL: `SELECT origin FROM subject_types WHERE id = ?`, PluginNever: true},
	{Kind: KindCredibilityGrade, OriginSQL: `SELECT origin FROM source_credibility_grades WHERE id = ?`, PluginNever: true},
}

var ownedReleases = []ownedRelease{
	{Parent: KindObservation, Column: "observations.value_date_id", Child: "date_values", Mode: "always"},
	{Parent: KindObservation, Column: "observations.value_name_id", Child: "name_values", Mode: "always"},
	{
		Parent: KindArtifact, Column: "artifacts.file_id", Child: "files", Mode: "ifUnused",
		Then: []ownedRelease{
			{Parent: KindFile, Column: "file_derivatives.source_file_id", Child: "file_derivatives", Mode: "always"},
			{Parent: KindFile, Column: "file_derivatives.derived_file_id", Child: "file_derivatives", Mode: "always"},
		},
	},
}

func tableByKind(kind Kind) (tableSpec, bool) {
	for _, t := range tables {
		if t.Kind == kind && t.Exists != "" {
			return t, true
		}
	}
	return tableSpec{}, false
}

func inboundFor(kind Kind) []inboundEdge {
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
		fkEdge(KindSource, "artifacts.source_id", KindArtifact, "artifacts", "source_id", "ref"),
		fkEdge(KindSource, "subjects.source_id", KindSubject, "subjects", "source_id", "ref"),
		fkEdge(KindArtifact, "citations.artifact_id", KindCitation, "citations", "artifact_id", "ref"),
		fkEdge(KindCitation, "observations.citation_id", KindObservation, "observations", "citation_id", "ref"),
		subjectResourceInbound(),
		fkEdge(KindSubject, "observations.value_subject_id", KindObservation, "observations", "value_subject_id", "ref"),
		fkEdge(KindSourceType, "sources.source_type_id", KindSource, "sources", "source_type_id", "ref"),
		joinEdge(KindMetadataField, "source_metadata.field_id", KindSource,
			`SELECT COUNT(DISTINCT source_id) FROM source_metadata WHERE field_id = ?`,
			`SELECT s.id, s.ref FROM source_metadata m
				JOIN sources s ON s.id = m.source_id
				WHERE m.field_id = ?
				ORDER BY s.ref COLLATE NOCASE LIMIT ?`),
		joinEdge(KindCredibilityGrade, "source_credibility_assessments.credibility_grade_id", KindSource,
			`SELECT COUNT(DISTINCT source_id) FROM source_credibility_assessments WHERE credibility_grade_id = ?`,
			`SELECT s.id, s.ref FROM source_credibility_assessments a
				JOIN sources s ON s.id = a.source_id
				WHERE a.credibility_grade_id = ?
				ORDER BY s.ref COLLATE NOCASE LIMIT ?`),
		fkEdge(KindSubjectType, "subjects.subject_type_id", KindSubject, "subjects", "subject_type_id", "ref"),
		fkEdge(KindProperty, "observations.property_id", KindObservation, "observations", "property_id", "ref"),
		joinEdge(KindProperty, "property_terms.property_id", KindPropertyTerm,
			`SELECT COUNT(*) FROM property_terms WHERE property_id = ?`,
			`SELECT id, key FROM property_terms WHERE property_id = ? ORDER BY key COLLATE NOCASE LIMIT ?`),
		fkEdge(KindPropertyTerm, "observations.value_term_id", KindObservation, "observations", "value_term_id", "ref"),
		fkEdge(KindFile, "artifacts.file_id", KindArtifact, "artifacts", "file_id", "ref"),
		fkEdge(KindSubjectType, "canonical_entities.subject_type_id", KindCanonicalEntity, "canonical_entities", "subject_type_id", "ref"),
		fkEdge(KindCanonicalEntity, "canonical_entities.merged_into_id", KindCanonicalEntity, "canonical_entities", "merged_into_id", "ref"),
		// Claims have no ref; name the handle each claim files onto.
		joinEdge(KindClaimConfidenceGrade, "identity_claims.confidence_grade_id", KindCanonicalEntity,
			`SELECT COUNT(DISTINCT entity_id) FROM identity_claims WHERE confidence_grade_id = ?`,
			`SELECT DISTINCT e.id, e.ref FROM identity_claims ic
				JOIN canonical_entities e ON e.id = ic.entity_id
				WHERE ic.confidence_grade_id = ?
				ORDER BY e.ref COLLATE NOCASE LIMIT ?`),
		reservedStub(KindObservation, ViaReconciliationEvidence, KindReconciliationClaim),
		reservedStub(KindObservation, ViaNarrativeTarget, KindNarrative),
	}
})

func originRuleFor(kind Kind) (originRule, bool) {
	for _, r := range originRules {
		if r.Kind == kind {
			return r, true
		}
	}
	return originRule{}, false
}

func ownedFor(kind Kind) []ownedRelease {
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
		if e.Bucket == BucketResource && e.List != nil {
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
