// Package catalogmodel is the declared model of catalog tables and foreign
// keys. A test checks it against the SQLite schema (sqlite_schema,
// PRAGMA foreign_key_list). Delete policy — probes, releases, Impact —
// lives in deleteimpact and reads this model.
package catalogmodel

// Kind is the stable machine key for a catalog table that delete policy
// and later registries name.
type Kind string

// Bucket classifies a table or a foreign key.
type Bucket string

// Kind identifiers for tables in this model.
const (
	KindSource           Kind = "source"
	KindArtifact         Kind = "artifact"
	KindCitation         Kind = "citation"
	KindObservation      Kind = "observation"
	KindSubject          Kind = "subject"
	KindSourceType       Kind = "source_type"
	KindMetadataField    Kind = "metadata_field"
	KindCredibilityGrade Kind = "source_credibility_grade"
	KindSubjectType      Kind = "subject_type"
	KindProperty         Kind = "property"
	KindPropertyTerm     Kind = "property_term"
	KindFile             Kind = "file"
	KindUser             Kind = "user"
	KindProject          Kind = "project"

	// Conclusion kinds: registered for inbound probes; no delete path yet.
	KindCanonicalEntity      Kind = "canonical_entity"
	KindClaimConfidenceGrade Kind = "claim_confidence_grade"
	// KindIdentityClaim parents facet releases (its pins) only; claims have no
	// ref and no delete of their own yet.
	KindIdentityClaim Kind = "identity_claim"
)

// Bucket tags one table or one live FK.
const (
	BucketResource        Bucket = "resource"
	BucketVocab           Bucket = "vocab"
	BucketFacet           Bucket = "facet"
	BucketOwnedOutbound   Bucket = "ownedOutbound"
	BucketPool            Bucket = "pool"
	BucketInfra           Bucket = "infra"
	BucketSkip            Bucket = "skip"
	BucketOptional        Bucket = "optional"
	BucketConnectionFacet Bucket = "connectionFacet"
)

// Table is one catalog table.
type Table struct {
	Name   string
	PK     string
	Kind   Kind
	Bucket Bucket
}

// FK is one live foreign key, keyed by its leading column. A composite key
// lists the whole tuple in FromCols; PRAGMA foreign_key_list returns one row
// per column, and the honesty test compares that tuple.
type FK struct {
	From, Column string
	FromCols     []string // nil means Column only
	To           string
	OnDelete     string
	Bucket       Bucket
	// Audited CASCADE facets are research rows: official deletes remove and
	// audit them through a facet release, and the CASCADE is only a backstop.
	// Unaudited CASCADEs (layout, vocab joins, value parts) stay silent.
	Audited bool
}

// Tables is every catalog table except sqlite_% and FTS shadow tables.
var Tables = []Table{
	{Name: "sources", PK: "id", Kind: KindSource, Bucket: BucketResource},
	{Name: "artifacts", PK: "id", Kind: KindArtifact, Bucket: BucketResource},
	{Name: "citations", PK: "id", Kind: KindCitation, Bucket: BucketResource},
	{Name: "observations", PK: "id", Kind: KindObservation, Bucket: BucketResource},
	{Name: "subjects", PK: "id", Kind: KindSubject, Bucket: BucketResource},
	{Name: "source_types", PK: "id", Kind: KindSourceType, Bucket: BucketVocab},
	{Name: "source_metadata_fields", PK: "id", Kind: KindMetadataField, Bucket: BucketVocab},
	{Name: "source_credibility_grades", PK: "id", Kind: KindCredibilityGrade, Bucket: BucketVocab},
	{Name: "subject_types", PK: "id", Kind: KindSubjectType, Bucket: BucketVocab},
	{Name: "properties", PK: "id", Kind: KindProperty, Bucket: BucketVocab},
	{Name: "property_terms", PK: "id", Kind: KindPropertyTerm, Bucket: BucketVocab},
	{Name: "canonical_entities", Kind: KindCanonicalEntity, Bucket: BucketResource},
	{Name: "claim_confidence_grades", Bucket: BucketVocab},
	{Name: "files", PK: "id", Kind: KindFile, Bucket: BucketPool},
	{Name: "users", PK: "id", Kind: KindUser, Bucket: BucketInfra},
	{Name: "project", PK: "id", Kind: KindProject, Bucket: BucketSkip},
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
	{Name: "auto_reconciler_values", Bucket: BucketSkip},
	{Name: "auto_reconciler_meta", Bucket: BucketSkip},
	{Name: "auto_reconciler_outcomes", Bucket: BucketSkip},
}

// FKs is every live catalog foreign key. The honesty test compares this to
// PRAGMA foreign_key_list.
var FKs = []FK{
	{From: "project", Column: "updated_by", To: "users", OnDelete: "NO ACTION", Bucket: BucketSkip},
	{From: "audit_transactions", Column: "user_id", To: "users", OnDelete: "NO ACTION", Bucket: BucketSkip},
	{From: "audit_changes", Column: "audit_transaction_id", To: "audit_transactions", OnDelete: "NO ACTION", Bucket: BucketSkip},
	{From: "audit_transaction_scopes", Column: "audit_transaction_id", To: "audit_transactions", OnDelete: "NO ACTION", Bucket: BucketSkip},

	{From: "sources", Column: "source_type_id", To: "source_types", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "sources", Column: "primary_artifact_id", To: "artifacts", OnDelete: "SET NULL", Bucket: BucketOptional},
	{From: "source_notes", Column: "source_id", To: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "source_metadata", Column: "source_id", To: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "source_metadata", Column: "field_id", To: "source_metadata_fields", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "source_credibility_assessments", Column: "source_id", To: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "source_credibility_assessments", Column: "credibility_grade_id", To: "source_credibility_grades", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "source_metadata_layout", Column: "source_id", To: "sources", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "source_metadata_layout", Column: "field_id", To: "source_metadata_fields", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "source_type_metadata_fields", Column: "source_type_id", To: "source_types", OnDelete: "CASCADE", Bucket: BucketFacet},
	{From: "source_type_metadata_fields", Column: "field_id", To: "source_metadata_fields", OnDelete: "CASCADE", Bucket: BucketFacet},

	{From: "artifacts", Column: "source_id", To: "sources", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "artifacts", Column: "file_id", To: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},

	{From: "file_derivatives", Column: "source_file_id", To: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},
	{From: "file_derivatives", Column: "derived_file_id", To: "files", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},

	{From: "citations", Column: "artifact_id", To: "artifacts", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "citation_notes", Column: "citation_id", To: "citations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},

	{From: "observations", Column: "citation_id", To: "citations", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "observations", Column: "subject_id", To: "subjects", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "observations", Column: "property_id", To: "properties", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "observations", Column: "value_date_id", To: "date_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},
	{From: "observations", Column: "value_name_id", To: "name_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound},
	{From: "observations", Column: "value_subject_id", To: "subjects", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "observations", Column: "value_term_id", To: "property_terms", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "observation_notes", Column: "observation_id", To: "observations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},

	{From: "subjects", Column: "source_id", To: "sources", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "subjects", Column: "subject_type_id", To: "subject_types", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "subject_positions", Column: "subject_id", To: "subjects", OnDelete: "CASCADE", Bucket: BucketFacet},
	{From: "subject_type_properties", Column: "subject_type_id", To: "subject_types", OnDelete: "CASCADE", Bucket: BucketFacet},
	{From: "subject_type_properties", Column: "property_id", To: "properties", OnDelete: "CASCADE", Bucket: BucketFacet},

	{From: "property_terms", Column: "property_id", To: "properties", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "name_value_parts", Column: "name_value_id", To: "name_values", OnDelete: "CASCADE", Bucket: BucketFacet},

	{From: "canonical_entities", Column: "subject_type_id", To: "subject_types", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "canonical_entities", Column: "merged_into_id", To: "canonical_entities", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "identity_claims", Column: "subject_id", FromCols: []string{"subject_id", "subject_type_id"}, To: "subjects", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "identity_claims", Column: "entity_id", FromCols: []string{"entity_id", "subject_type_id"}, To: "canonical_entities", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "identity_claims", Column: "confidence_grade_id", To: "claim_confidence_grades", OnDelete: "NO ACTION", Bucket: BucketResource},
	{From: "identity_claim_evidence", Column: "identity_claim_id", To: "identity_claims", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	{From: "identity_claim_evidence", Column: "observation_id", To: "observations", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true},
	// Derived auto-reconciler cache: silent backstops; upkeep rewrites the rows.
	{From: "auto_reconciler_values", Column: "entity_id", To: "canonical_entities", OnDelete: "CASCADE", Bucket: BucketSkip},
	{From: "auto_reconciler_values", Column: "property_id", To: "properties", OnDelete: "CASCADE", Bucket: BucketSkip},
	{From: "auto_reconciler_outcomes", Column: "entity_id", To: "canonical_entities", OnDelete: "CASCADE", Bucket: BucketSkip},
	{From: "auto_reconciler_outcomes", Column: "property_id", To: "properties", OnDelete: "CASCADE", Bucket: BucketSkip},
}
