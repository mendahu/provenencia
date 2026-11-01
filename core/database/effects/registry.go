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
	DocEvidenceGraph = string(catalogmodel.KindEvidenceGraph)
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

// Read names a partial read the app keeps per key: a write that moves its
// rows reports the key ids (the rows of KeyTable), and the app refreshes only
// those. The values are the wire keys (WriteEffects.reads) and Swift's CatalogRead raw
// values.
type Read string

const (
	// ReadCitations: the citation counts and the citations by Artifact of a
	// graph.
	ReadCitations Read = "citations"
	// ReadEvidenceGraphProgress: the Subject, filed and Observation counts of
	// a graph.
	ReadEvidenceGraphProgress Read = "evidence_graph_progress"
	// ReadEvidenceGraphs: the graphs of a Source, with their linked Artifacts
	// (the graph list).
	ReadEvidenceGraphs Read = "evidence_graphs"
)

// KeyTable is the table whose ids key the read: Evidence graphs for
// citations and evidence_graph_progress, Sources for evidence_graphs. "" for
// an unknown read.
func (r Read) KeyTable() string {
	switch r {
	case ReadCitations, ReadEvidenceGraphProgress:
		return "evidence_graphs"
	case ReadEvidenceGraphs:
		return "sources"
	}
	return ""
}

// Reads is every keyed read, in the order Set.Reads reports them.
var Reads = []Read{ReadCitations, ReadEvidenceGraphProgress, ReadEvidenceGraphs}

// ReadPath is one keyed read: Read says which, Path says whose key ids (it
// ends at Read.KeyTable()).
type ReadPath struct {
	Read Read
	Path Path
}

// ReadIDs is a resolved keyed read: one read and its key ids.
type ReadIDs struct {
	Read Read
	IDs  [][]byte
}

// List names a catalog-wide list a write changed: one the app caches whole
// and a row change cannot be patched into. The values are the wire keys
// (WriteEffects.lists) and Swift's CatalogList raw values.
type List string

const (
	// ListVocabulary: a vocabulary label (source type, metadata field,
	// subject type, Property, Property term) or a subject type's property
	// binding changed.
	ListVocabulary List = "vocabulary"
	// ListPropertyUsage: an Observation was created, deleted, or moved to
	// another Property, so the Properties workspace's usedBy counts changed.
	ListPropertyUsage List = "property_usage"
)

// Lists is every catalog-wide list, in the order Set.Lists reports them.
var Lists = []List{ListVocabulary, ListPropertyUsage}

// Set is what a batch of changes touches. Empty Handles or Search means that
// job has nothing to do. Lists and Structure are the catalog-wide fields.
type Set struct {
	Handles [][]byte
	// EvidenceGraphs are the Evidence graphs whose loaded rows changed
	// (Effect.EvidenceGraph): graphcache drops each.
	EvidenceGraphs [][]byte
	// Reads are the keyed partial reads a batch moved, each with the key ids
	// it moved them for (Effect.Reads), in the order of Reads.
	Reads []ReadIDs
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
	// EvidenceGraph is the Evidence graph whose loaded rows this change moves
	// (a Path that ends at evidence_graphs): graphcache drops it. A table the
	// graph cache does not load leaves it empty.
	EvidenceGraph Path
	// Reads are the keyed partial reads this change moves, each with a Path
	// that ends at the read's KeyTable.
	Reads []ReadPath
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

// observationGraph is an Observation to the Evidence graph of its Subject.
var observationGraph = up("subject_id", "graph_id")

// subjectGraph is a Subject to its Evidence graph.
var subjectGraph = field("graph_id")

// claimGraph is an Identity Claim to the Evidence graph of its Subject.
var claimGraph = from("subject_id", field("graph_id"))

// graphsOfSource is a path ending at a Source, to every Evidence graph of it.
func graphsOfSource(p Path) Path {
	return chain(p, inbound("evidence_graphs", "source_id"))
}

// valueGraph is a value row to the Evidence graphs of the Observations that
// hold it in col.
func valueGraph(col string) Path {
	return chain(inbound("observations", col), observationGraph)
}

// citationObservationGraphs is a Citation to the Evidence graphs of its
// Observations.
var citationObservationGraphs = chain(inbound("observations", "citation_id"), observationGraph)

// valueSource is a value row to the Source of the Observations that hold it
// in col, through each Observation's Citation. SourceScope walks it;
// EvidenceGraph walks valueGraph, through the Observation's Subject.
func valueSource(col string) Path {
	return chain(inbound("observations", col), up("citation_id", "artifact_id", "source_id"))
}

// ownedValue is the effect of a value row an Observation owns in col: the
// change carries the Evidence graphs, the Source and the handles of the
// Observations that hold it.
func ownedValue(entity, col string) Effect {
	return Effect{
		Entity:        entity,
		EvidenceGraph: valueGraph(col),
		Reads:         []ReadPath{{ReadCitations, valueGraph(col)}},
		SourceScope:   valueSource(col),
		Handles:       valueMembers(col),
		HandleScope:   valueMembers(col),
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
		Entity:        "source",
		EvidenceGraph: graphsOfSource(union(onField("title", self()), onField("source_type_id", self()))),
		SourceScope:   self(),
		Search: []SearchDoc{
			{Kind: DocSource, Path: self()},
			{Kind: DocEvidenceGraph, Path: onField("title", graphsOfSource(self()))},
		},
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
	// An Artifact's own fields move no loaded graph row and no citation
	// membership: its Citations and Observations carry their own graphs.
	"artifacts": {
		Entity:      "artifact",
		SourceScope: field("source_id"),
		Search:      []SearchDoc{{Kind: DocSource, Path: field("source_id")}},
	},
	"citations": {
		Entity:        "citation",
		EvidenceGraph: citationObservationGraphs,
		// A Citation belongs to the one graph it was made in.
		Reads:       []ReadPath{{ReadCitations, field("graph_id")}},
		SourceScope: up("artifact_id", "source_id"),
		Handles:     onField("transcription_uncertain", sqlPath(sqlHandlesForCitation)),
	},
	"citation_notes": {Entity: "citation_note", SourceScope: up("citation_id", "artifact_id", "source_id")},
	"observations": {
		Entity:        "observation",
		EvidenceGraph: observationGraph,
		Reads: []ReadPath{
			{ReadCitations, observationGraph},
			{ReadEvidenceGraphProgress, observationGraph},
		},
		Lists:        []List{ListPropertyUsage},
		ListsOnField: "property_id",
		SourceScope:  up("citation_id", "artifact_id", "source_id"),
		Handles:      observationMembers,
		HandleScope:  observationMembers,
	},
	"observation_notes": {
		Entity:      "observation_note",
		SourceScope: up("observation_id", "citation_id", "artifact_id", "source_id"),
		HandleScope: chain(up("observation_id", "subject_id"), membersOf),
	},
	"subjects": {
		Entity:        "subject",
		EvidenceGraph: subjectGraph,
		Reads:         []ReadPath{{ReadEvidenceGraphProgress, subjectGraph}},
		SourceScope:   field("source_id"),
	},
	"evidence_graphs": {
		Entity:        "evidence_graph",
		EvidenceGraph: self(),
		Reads:         []ReadPath{{ReadEvidenceGraphs, field("source_id")}},
		SourceScope:   field("source_id"),
		Search:        []SearchDoc{{Kind: DocEvidenceGraph, Path: self()}},
	},
	"evidence_graph_artifacts": {
		Entity:      "evidence_graph_artifact",
		Reads:       []ReadPath{{ReadEvidenceGraphs, up("graph_id", "source_id")}},
		SourceScope: up("graph_id", "source_id"),
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
		Entity:        "identity_claim",
		EvidenceGraph: claimGraph,
		Reads:         []ReadPath{{ReadEvidenceGraphProgress, claimGraph}},
		Handles: union(
			field("entity_id"),
			onStatus("accepted", from("subject_id", observers)),
		),
		HandleScope: field("entity_id"),
	},

	"source_credibility_grades":   {None: true},
	"users":                       {None: true},
	"subject_positions":           {Entity: "subject_position", None: true},
	"subject_type_properties":     {Entity: "subject_type_property", Lists: []List{ListVocabulary}},
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
	out := []Path{e.EvidenceGraph, e.SourceScope, e.Handles, e.HandleScope, e.Structure}
	for _, r := range e.Reads {
		out = append(out, r.Path)
	}
	return out
}

func (e Effect) validate(table string) error {
	if e.None {
		for _, p := range e.paths() {
			if !p.zero() {
				return errf("%s: None entry still has effects", table)
			}
		}
		if len(e.Search) > 0 || len(e.Lists) > 0 || len(e.Reads) > 0 {
			return errf("%s: None entry still has effects", table)
		}
		return nil
	}
	for _, r := range e.Reads {
		if !slices.Contains(Reads, r.Read) {
			return errf("%s: undeclared read %q", table, r.Read)
		}
		if r.Path.zero() {
			return errf("%s: read %q has no path", table, r.Read)
		}
		if end, err := r.Path.check(table); err == nil && end != "" && end != r.Read.KeyTable() {
			return errf("%s: read %q ends at %s, want %s", table, r.Read, end, r.Read.KeyTable())
		}
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
