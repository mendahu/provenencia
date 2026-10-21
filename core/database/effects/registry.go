package effects

import (
	"slices"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"
)

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

// List names a catalog-wide list a write changed: one the app caches whole
// and a row change cannot be patched into. The values are the wire keys
// (WriteEffects.lists) and Swift's CatalogList raw values.
type List string

const (
	// ListVocabulary: a vocabulary label (source type, metadata field,
	// subject type, Property, Property term) changed.
	ListVocabulary List = "vocabulary"
	// ListPropertyUsage: an Observation was created, deleted, or moved to
	// another Property, so the Properties workspace's usedBy counts changed.
	ListPropertyUsage List = "property_usage"
	// ListEvidenceGraphs: a graph or link row was written.
	ListEvidenceGraphs List = "evidence_graphs"
)

// Lists is every catalog-wide list, in the order Set.Lists reports them.
var Lists = []List{ListVocabulary, ListPropertyUsage, ListEvidenceGraphs}

// Set is what a batch of changes touches. Empty Handles or Search means that
// job has nothing to do. Lists and Structure are the catalog-wide fields.
type Set struct {
	Handles [][]byte
	// Sources are the Sources whose graph changed (Effect.Source): what the
	// Source graph and graph progress load.
	Sources [][]byte
	// CitationSources are the Sources whose citation reads (citation counts,
	// citations by artifact) a change moves (Effect.CitationSource).
	CitationSources [][]byte
	// SourceScopes are the Sources whose last activity a change moves
	// (Effect.SourceScope), the audit source scope.
	SourceScopes [][]byte
	// HandleScopes are the handles whose canonical_entity audit scope the
	// batch moves (Effect.HandleScope); Handles is the recompute set.
	HandleScopes [][]byte
	Search       []SearchIDs
	// Lists are the catalog-wide lists the batch changed, deduped, in the
	// order of Lists.
	Lists []List
	// Structure is set when the graph cache drops whole; it is not a list.
	Structure bool
	// HeaderDeps are handles whose stored rows read Handles. Run fills this
	// from the walk RecomputeTx already did. HeaderDepsWalked says that walk
	// happened, including when it found nobody, so a commit drops those rows
	// and does not walk again.
	HeaderDeps       [][]byte
	HeaderDepsWalked bool
}

// Effect is what a change to one table touches. An empty field means that job
// is unaffected. None is an explicit empty entry for a table outside the skip
// bucket that has no effects.
type Effect struct {
	Entity string
	// Source is the Source whose graph this change changes: the Source graph
	// and graph progress load the table. A table neither loads leaves it empty.
	Source Path
	// CitationSource is the Source whose citation reads (citation counts,
	// citations by artifact) load the table. A table neither loads leaves it
	// empty.
	CitationSource Path
	// SourceScope is the Source whose last activity (audit source scope) this
	// change moves.
	SourceScope Path
	Handles     Path
	// HandleScope is the handles whose canonical_entity audit scope this
	// change moves; Handles is the recompute set.
	HandleScope Path
	Search      []SearchDoc
	// Lists are the catalog-wide lists a change to the table moves.
	Lists []List
	// ListsOnField, when set, narrows an update: it moves Lists only if it
	// writes this field. Creates and deletes always move them.
	ListsOnField string
	Structure    Path
	None         bool
}

// Has reports whether the batch changed the list.
func (s Set) Has(l List) bool {
	for _, got := range s.Lists {
		if got == l {
			return true
		}
	}
	return false
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
	sqlHandlesForTerm = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		WHERE ic.status IN ('accepted', 'provisional') AND o.value_term_id = ?`
)

// membersOf is a subject to its accepted and provisional handles.
var membersOf = across("identity_claims", "subject_id", "entity_id", "accepted", "provisional")

// observationMembers is an Observation to the handles of its Subject.
var observationMembers = from("subject_id", membersOf)

// valueMembers is a value row to the handles of the Subjects whose
// Observations hold it in col.
func valueMembers(col string) Path {
	return chain(inbound("observations", col), field("subject_id"), membersOf)
}

// valueSource is a value row to the Source of the Observations that hold it
// in col. Both Source and SourceScope walk it.
func valueSource(col string) Path {
	return chain(inbound("observations", col), up("citation_id", "artifact_id", "source_id"))
}

// ownedValue is the effect of a value row an Observation owns in col: the
// change carries the Source and handles of the Observations that hold it.
func ownedValue(entity, col string) Effect {
	return Effect{
		Entity:         entity,
		Source:         valueSource(col),
		CitationSource: valueSource(col),
		SourceScope:    valueSource(col),
		Handles:        valueMembers(col),
		HandleScope:    valueMembers(col),
	}
}

// movesLists reports whether the change moves the entry's Lists: always for
// a create or delete, and for an update only when it writes ListsOnField.
func (e Effect) movesLists(ch rowchange.Change) bool {
	if len(e.Lists) == 0 {
		return false
	}
	if ch.Action != rowchange.ActionUpdate || e.ListsOnField == "" {
		return true
	}
	_, ok := ch.Fields[e.ListsOnField]
	return ok
}

// observers is the handles whose members point at a subject.
var observers = chain(
	inbound("observations", "value_subject_id"),
	field("subject_id"),
	membersOf,
)

var registry = map[string]Effect{
	"sources": {
		Entity:      "source",
		Source:      self(),
		SourceScope: self(),
		Search:      []SearchDoc{{Kind: DocSource, Path: self()}},
	},
	"source_notes": {
		Entity:      "source_note",
		SourceScope: field("source_id"),
		Search:      []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"source_metadata": {
		Entity:      "source_metadata",
		SourceScope: field("source_id"),
		Search:      []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"source_credibility_assessments": {
		Entity:      "source_credibility_assessment",
		SourceScope: field("source_id"),
		Handles:     onField("credibility_grade_id", from("source_id", sqlPath(sqlHandlesForSource))),
		Search:      []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"source_metadata_layout": {
		Entity:      "source_metadata_layout",
		SourceScope: field("source_id"),
		Search:      []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"artifacts": {
		Entity:         "artifact",
		Source:         field("source_id"),
		CitationSource: field("source_id"),
		SourceScope:    field("source_id"),
		Search:         []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"citations": {
		Entity:         "citation",
		Source:         up("artifact_id", "source_id"),
		CitationSource: up("artifact_id", "source_id"),
		SourceScope:    up("artifact_id", "source_id"),
		Handles:        onField("transcription_uncertain", sqlPath(sqlHandlesForCitation)),
	},
	"citation_notes": {Entity: "citation_note", SourceScope: up("citation_id", "artifact_id", "source_id")},
	"observations": {
		Entity:         "observation",
		Source:         up("citation_id", "artifact_id", "source_id"),
		CitationSource: up("citation_id", "artifact_id", "source_id"),
		Lists:          []List{ListPropertyUsage},
		ListsOnField:   "property_id",
		SourceScope:    up("citation_id", "artifact_id", "source_id"),
		Handles:        observationMembers,
		HandleScope:    observationMembers,
	},
	"observation_notes": {
		Entity:      "observation_note",
		SourceScope: up("observation_id", "citation_id", "artifact_id", "source_id"),
		HandleScope: chain(up("observation_id", "subject_id"), membersOf),
	},
	"subjects": {Entity: "subject", Source: field("source_id"), SourceScope: field("source_id")},
	"evidence_graphs": {
		Entity:      "evidence_graph",
		SourceScope: field("source_id"),
		Lists:       []List{ListEvidenceGraphs},
	},
	"evidence_graph_artifacts": {
		Entity:      "evidence_graph_artifact",
		SourceScope: up("graph_id", "source_id"),
		Lists:       []List{ListEvidenceGraphs},
	},
	"files": {
		Entity:      "file",
		SourceScope: chain(inbound("artifacts", "file_id"), field("source_id")),
		Search:      []SearchDoc{{Kind: DocSource, Path: chain(inbound("artifacts", "file_id"), field("source_id"))}},
	},
	"date_values":      ownedValue("date_value", "value_date_id"),
	"name_values":      ownedValue("name_value", "value_name_id"),
	"depiction_values": ownedValue("depiction_value", "value_depiction_id"),
	"source_types": {
		Entity: "source_type",
		Lists:  []List{ListVocabulary},
		Search: []SearchDoc{
			{Kind: DocSourceType, Path: self()},
			{Kind: DocSource, Path: inbound("sources", "source_type_id")},
		},
	},
	"source_metadata_fields": {
		Entity: "metadata_field",
		Lists:  []List{ListVocabulary},
		Search: []SearchDoc{{Kind: DocMetadataField, Path: self()}},
	},
	"subject_types": {Lists: []List{ListVocabulary}},
	"properties": {
		Entity:  "property",
		Lists:   []List{ListVocabulary},
		Handles: onField("cardinality", sqlPath(sqlHandlesForProperty)),
	},
	"property_terms": {
		Entity: "property_term",
		Lists:  []List{ListVocabulary},
		// A neutral term reconciles differently, so the handles that observe
		// it are recomputed.
		Handles: onField("neutral", sqlPath(sqlHandlesForTerm)),
		Structure: union(
			onField("directed", self()),
			onField("inverse_key", self()),
			onField("neutral", self()),
		),
	},
	"canonical_entities": {Entity: "canonical_entity", Handles: self(), HandleScope: self()},
	"identity_claims": {
		Entity: "identity_claim",
		Source: from("subject_id", field("source_id")),
		Handles: union(
			field("entity_id"),
			onStatus("accepted", from("subject_id", observers)),
		),
		HandleScope: field("entity_id"),
	},
	"identity_claim_evidence": {
		Entity:      "identity_claim_evidence",
		Source:      up("identity_claim_id", "subject_id", "source_id"),
		Handles:     up("identity_claim_id", "entity_id"),
		HandleScope: up("identity_claim_id", "entity_id"),
	},

	"source_credibility_grades":   {None: true},
	"claim_confidence_grades":     {None: true},
	"users":                       {None: true},
	"subject_positions":           {Entity: "subject_position", None: true},
	"subject_type_properties":     {None: true},
	"source_type_metadata_fields": {None: true},
	"name_value_parts":            {None: true},
	"name_format_profiles":        {None: true},
	"name_format_profile_parts":   {None: true},
	"name_format_profile_roles":   {None: true},
	"name_format_profile_weights": {None: true},
	"name_alias_tables":           {None: true},
	"name_alias_members":          {None: true},
	"name_alias_roles":            {None: true},
	"project_settings":            {None: true},
	"file_derivatives":            {Entity: "file_derivative", None: true},
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

// paths lists every path field of the effect.
func (e Effect) paths() []Path {
	return []Path{e.Source, e.CitationSource, e.SourceScope, e.Handles, e.HandleScope, e.Structure}
}

func (e Effect) validate(table string) error {
	if e.None {
		for _, p := range e.paths() {
			if !p.zero() {
				return errf("%s: None entry still has effects", table)
			}
		}
		if len(e.Search) > 0 || len(e.Lists) > 0 {
			return errf("%s: None entry still has effects", table)
		}
		return nil
	}
	for _, l := range e.Lists {
		if !slices.Contains(Lists, l) {
			return errf("%s: undeclared list %q", table, l)
		}
	}
	if e.ListsOnField != "" && len(e.Lists) == 0 {
		return errf("%s: ListsOnField without Lists", table)
	}
	for _, p := range e.paths() {
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
