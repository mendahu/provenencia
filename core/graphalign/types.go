// Package graphalign proposes a handle, New, or Skip for every primary
// Subject on one Evidence layer (S9-41). Pure: no catalog I/O. S9-42 loads
// Layer / Canon and calls Align.
package graphalign

import (
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Subject is one primary Subject on the Evidence layer.
type Subject struct {
	ID         []byte
	Ref        string
	Kind       string // person, event, place
	Values     subjectcompare.Values
	Provenance float64 // scales node scores; 1 is default, below 1 weakens
}

// Bridge links two primary Subjects on the layer (walked undirected).
type Bridge struct {
	// Edge's Associations are the bridge Subjects behind the edge. The layers
	// never read them; the loader sets them for a later reader of the bridge
	// Subjects (S13-29's shared loader).
	kinship.Edge
	// Recording is how the kinship patterns imply the link; the zero value is a
	// link only recorded.
	kinship.Recording
}

// Handle is one canonical entity in the prefetched canon neighborhood.
type Handle struct {
	ID     []byte
	Ref    string
	Kind   string
	Values subjectcompare.Values
}

// CanonEdge is one filed association hop, or a link the kinship patterns
// imply, between handles.
type CanonEdge struct {
	From, To       []byte
	KeyFrom, KeyTo vocabkey.BridgeKey // the edge read from From's end and from To's
	// Recording is as on Bridge.
	kinship.Recording
}

// Layer is one Source's primary Subjects and bridges, recorded or implied by
// the kinship patterns.
type Layer struct {
	Subjects []Subject
	Bridges  []Bridge
	Metas    subjectcompare.Metas
}

// Canon is the part of the canonical graph reachable from the seed handles.
type Canon struct {
	Handles []Handle
	Edges   []CanonEdge
	// Candidates are each Subject's shortlisted candidates, by Subject id, in
	// Candidates order; the loader fills them.
	Candidates map[string][][]byte
}

// Fixed is a decided or already-promoted row Align holds as given. Target
// empty or TargetHandle anchors SubjectID on HandleID; TargetNew or
// TargetSkip keeps the Subject off every handle, so the walk neither maps it
// nor walks through it.
type Fixed struct {
	SubjectID []byte
	HandleID  []byte
	Target    Target
}

// Target is what Align proposes for one Subject.
type Target string

const (
	TargetHandle Target = "handle"
	TargetNew    Target = "new"
	TargetSkip   Target = "skip"
)

// Assessment is the strong / medium / weak / no-match band.
type Assessment string

const (
	AssessStrong Assessment = "strong"
	AssessMedium Assessment = "medium"
	AssessWeak   Assessment = "weak"
	AssessNone   Assessment = "none"
)

// Comparison is one lead property's outcome on the chosen (or top) candidate:
// an input to the score. Key, Result (Outcome, Cause, Speaking: A the
// Subject's value, B the handle's) and Points come from subject compare.
type Comparison struct {
	subjectcompare.PropertyResult
	// Derived: the lead is a property the profile builder derives (a person's
	// lifespan), with no catalog row.
	Derived bool
}

// Via is the layer edge that seeded this row's handle, when the walk
// reached it from an already-decided neighbor.
type Via struct {
	NeighborSubjectID []byte
	Key               vocabkey.BridgeKey
}

// Reason is why a row reads as it does. Codes, not copy: the page words them.
type Reason string

const (
	ReasonVia     Reason = "via"     // reached from a mapped neighbor (Row.Via)
	ReasonDecided Reason = "decided" // held as given (a decision or an existing claim)
	ReasonAgrees  Reason = "agrees"  // property-only match; ReasonProperty agrees most
	ReasonWeak    Reason = "weak"    // an alternative below the weak bar: too little agrees to name
	ReasonTaken   Reason = "taken"   // another row took the handle (Flags.DuplicateOf)
	ReasonEmpty   Reason = "empty"   // nothing to match on
	// ReasonUnplaced is the zero value: Align placed nothing and nothing else
	// applies; the page says no record is suggested.
	ReasonUnplaced Reason = ""
)

// Move reasons: why settling released a row's earlier pair. They extend Reason
// and appear only in Move.Reason.
const (
	ReasonNeighborPlaced   Reason = "neighbor_placed"   // a neighbor was placed
	ReasonNeighborMoved    Reason = "neighbor_moved"    // a neighbor was placed on another handle than before
	ReasonNeighborReleased Reason = "neighbor_released" // a neighbor was released
)

// Move is why settling released this row's earlier pair: a neighbor was
// placed, moved or released.
type Move struct {
	Reason            Reason
	NeighborSubjectID []byte
}

// Alternative is a handle the researcher can choose. Assessment and Reason
// describe this record on its own, so the page can show them for the
// selected record and show nothing for Skip, New, or an empty menu.
type Alternative struct {
	HandleID       []byte
	Ref            string
	Score          float64
	Assessment     Assessment
	Reason         Reason // via, agrees, or weak
	ReasonProperty vocabkey.PropertyKey
	ViaNeighbor    []byte
}

// RowFlags call out conflicts and duplicates for the page.
type RowFlags struct {
	// ConflictWithFixed: a held handle row whose neighbors now point at a
	// clearly stronger handle (Alternatives[0]).
	ConflictWithFixed bool
	// PossibleDuplicate: DuplicateOf may be the same entity as this row. It
	// took the handle this row would have matched, holds the same handle,
	// or both rows would be new records (unplaced or decided New) and score
	// alike.
	PossibleDuplicate bool
	DuplicateOf       []byte
	// Contested: a comparison behind the pair this row reports conflicts,
	// whatever its band.
	Contested bool
}

// Row is one Subject's proposal.
type Row struct {
	SubjectID []byte
	Kind      string
	// Target is TargetHandle when Align places the Subject; a held decision
	// keeps its TargetNew or TargetSkip; otherwise empty: Align suggests nothing.
	Target       Target
	HandleID     []byte
	HandleRef    string
	Score        float64
	Assessment   Assessment
	Alternatives []Alternative
	Comparisons  []Comparison
	Reason       Reason
	// ReasonProperty is the agreeing Property behind ReasonAgrees.
	ReasonProperty vocabkey.PropertyKey
	Via            *Via
	Flags          RowFlags
	// Move is set when settling released an earlier pair of this row and the
	// row's final target is not that handle.
	Move *Move
}

// Proposal is Align's full output: one row per primary Subject.
type Proposal struct {
	Rows []Row
}
