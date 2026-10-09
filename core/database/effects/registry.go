package effects

import "github.com/mendahu/provenencia/core/database/catalogmodel"

// Search document kinds. Nothing reads these until the write orchestrator does.
const (
	DocSource        = "source"
	DocSourceType    = "source_type"
	DocMetadataField = "metadata_field"
)

// SearchDoc is one reprojection: Kind says which document, Path says whose ids.
type SearchDoc struct {
	Kind string
	Path Path
}

// SearchIDs is a resolved search reprojection: one document kind and its ids.
type SearchIDs struct {
	Kind string
	IDs  [][]byte
}

// Set is what a batch of changes touches. Empty Handles or Search means that
// job has nothing to do. Vocabulary and Structure are catalog-wide.
type Set struct {
	Handles    [][]byte
	Sources    [][]byte
	Search     []SearchIDs
	Vocabulary bool
	Structure  bool
}

// Effect is what a change to one table touches. An empty field means that job
// is unaffected. None is an explicit empty entry for a table outside the skip
// bucket that has no effects.
type Effect struct {
	Entity     string
	Source     Path
	Handles    Path
	Search     []SearchDoc
	Vocabulary bool
	Structure  Path
	None       bool
}

// Handle queries mirror autoreconciler. They stay here so the registry can
// name them before anything calls Handles.
const (
	sqlHandlesForCitation = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		WHERE ic.status IN ('accepted', 'provisional') AND o.citation_id = ?`

	sqlHandlesForSource = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		WHERE ic.status IN ('accepted', 'provisional') AND a.source_id = ?`

	sqlHandlesForProperty = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		WHERE ic.status IN ('accepted', 'provisional') AND o.property_id = ?`
)

// membersOf is a subject to its accepted and provisional handles.
var membersOf = across("identity_claims", "subject_id", "entity_id", "accepted", "provisional")

// observers is the handles whose members point at a subject.
var observers = chain(
	inbound("observations", "value_subject_id"),
	field("subject_id"),
	membersOf,
)

var registry = map[string]Effect{
	"sources": {
		Entity: "source",
		Source: self(),
		Search: []SearchDoc{{Kind: DocSource, Path: self()}},
	},
	"source_notes": {
		Entity: "source_note",
		Source: field("source_id"),
		Search: []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"source_metadata": {
		Entity: "source_metadata",
		Source: field("source_id"),
		Search: []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"source_credibility_assessments": {
		Entity:  "source_credibility_assessment",
		Source:  field("source_id"),
		Handles: onField("credibility_grade_id", from("source_id", sqlPath(sqlHandlesForSource))),
		Search:  []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"source_metadata_layout": {
		Entity: "source_metadata_layout",
		Source: field("source_id"),
		Search: []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"artifacts": {
		Entity: "artifact",
		Source: field("source_id"),
		Search: []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"citations": {
		Entity:  "citation",
		Source:  up("artifact_id", "source_id"),
		Handles: onField("transcription_uncertain", sqlPath(sqlHandlesForCitation)),
	},
	"citation_notes": {Entity: "citation_note", Source: up("citation_id", "artifact_id", "source_id")},
	"observations": {
		Entity:  "observation",
		Source:  up("citation_id", "artifact_id", "source_id"),
		Handles: from("subject_id", membersOf),
	},
	"observation_notes": {Entity: "observation_note", Source: up("observation_id", "citation_id", "artifact_id", "source_id")},
	"subjects":          {Entity: "subject", Source: field("source_id")},
	"files":             {Entity: "file", Source: chain(inbound("artifacts", "file_id"), field("source_id"))},
	"date_values": {
		Entity: "date_value",
		Source: chain(
			inbound("observations", "value_date_id"),
			up("citation_id", "artifact_id", "source_id"),
		),
		Handles: chain(
			inbound("observations", "value_date_id"),
			field("subject_id"),
			membersOf,
		),
	},
	"name_values": {
		Entity: "name_value",
		Source: chain(
			inbound("observations", "value_name_id"),
			up("citation_id", "artifact_id", "source_id"),
		),
		Handles: chain(
			inbound("observations", "value_name_id"),
			field("subject_id"),
			membersOf,
		),
	},
	"source_types": {
		Entity:     "source_type",
		Vocabulary: true,
		Search: []SearchDoc{
			{Kind: DocSourceType, Path: self()},
			{Kind: DocSource, Path: inbound("sources", "source_type_id")},
		},
	},
	"source_metadata_fields": {
		Entity:     "metadata_field",
		Vocabulary: true,
		Search:     []SearchDoc{{Kind: DocMetadataField, Path: self()}},
	},
	"subject_types": {Vocabulary: true},
	"properties": {
		Entity:     "property",
		Vocabulary: true,
		Handles:    onField("cardinality", sqlPath(sqlHandlesForProperty)),
	},
	"property_terms": {
		Entity:     "property_term",
		Vocabulary: true,
		Structure: union(
			onField("directed", self()),
			onField("inverse_key", self()),
		),
	},
	"canonical_entities": {Entity: "canonical_entity"},
	"identity_claims": {
		Entity: "identity_claim",
		Handles: union(
			field("entity_id"),
			onStatus("accepted", from("subject_id", observers)),
		),
	},
	"identity_claim_evidence": {
		Entity:  "identity_claim_evidence",
		Handles: up("identity_claim_id", "entity_id"),
	},

	"source_credibility_grades":   {None: true},
	"claim_confidence_grades":     {None: true},
	"users":                       {None: true},
	"subject_positions":           {None: true},
	"subject_type_properties":     {None: true},
	"source_type_metadata_fields": {None: true},
	"name_value_parts":            {None: true},
	"file_derivatives":            {None: true},
}

// entityTable maps a catalog table to the entity type stored on a change.
var entityTable map[string]string

func init() {
	seen := map[string]struct{}{}
	byEntity := map[string]string{}
	for _, t := range catalogmodel.Tables {
		if t.Bucket == catalogmodel.BucketSkip {
			if _, ok := registry[t.Name]; ok {
				panic("effects: skip table has an entry: " + t.Name)
			}
			continue
		}
		eff, ok := registry[t.Name]
		if !ok {
			panic("effects: no entry for " + t.Name)
		}
		if err := eff.validate(t.Name); err != nil {
			panic(err)
		}
		if eff.Entity != "" {
			if prev, ok := byEntity[eff.Entity]; ok {
				panic("effects: entity " + eff.Entity + " on " + prev + " and " + t.Name)
			}
			byEntity[eff.Entity] = t.Name
		}
		seen[t.Name] = struct{}{}
	}
	for name := range registry {
		if _, ok := seen[name]; !ok {
			panic("effects: entry for unknown table " + name)
		}
	}
	entityTable = byEntity
}

func (e Effect) validate(table string) error {
	if e.None {
		if e.Entity != "" || !e.Source.zero() || !e.Handles.zero() || !e.Structure.zero() || len(e.Search) > 0 || e.Vocabulary {
			return errf("%s: None entry still has effects", table)
		}
		return nil
	}
	for _, p := range []Path{e.Source, e.Handles, e.Structure} {
		if p.zero() {
			continue
		}
		if _, err := p.check(table); err != nil {
			return errf("%s: %v", table, err)
		}
	}
	for _, doc := range e.Search {
		if doc.Kind == "" || doc.Path.zero() {
			return errf("%s: search doc is incomplete", table)
		}
		if _, err := doc.Path.check(table); err != nil {
			return errf("%s search: %v", table, err)
		}
	}
	return nil
}
