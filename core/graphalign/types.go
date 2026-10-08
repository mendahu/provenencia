// Package graphalign proposes a handle, New, or Skip for every primary
// Subject on one Evidence layer (S9-41). Pure: no catalog I/O. S9-42 loads
// Layer / Canon / Stats and calls Align.
package graphalign

import "github.com/mendahu/provenencia/core/match"

// EdgeSignature identifies corresponding bridges on the layer and edges on
// the canonical graph (design §6).
type EdgeSignature struct {
	BridgeType       string // participation, relationship, location, place_relationship
	RoleOrType       string // role or relationship / place-relationship term key
	NeighborKind     string // person, event, place
	NeighborTypeTerm string // e.g. birth on an event
}

// Key is a stable map key for the signature.
func (s EdgeSignature) Key() string {
	return s.BridgeType + "|" + s.RoleOrType + "|" + s.NeighborKind + "|" + s.NeighborTypeTerm
}

// Subject is one primary Subject on the Evidence layer.
type Subject struct {
	ID         []byte
	Ref        string
	Kind       string // person, event, place
	Values     match.Values
	Provenance float64 // scales node scores; 1 is default, below 1 weakens
}

// Bridge links two primary Subjects on the layer (walked undirected).
type Bridge struct {
	A, B      []byte
	Signature EdgeSignature
}

// Handle is one canonical entity in the prefetched canon neighborhood.
type Handle struct {
	ID     []byte
	Ref    string
	Kind   string
	Values match.Values
}

// CanonEdge is one filed association hop between handles.
type CanonEdge struct {
	From, To  []byte
	Signature EdgeSignature
}

// Layer is one Source's primary Subjects and bridges.
type Layer struct {
	Subjects []Subject
	Bridges  []Bridge
	Metas    []match.PropertyMeta
}

// Canon is a bounded piece of the canonical graph.
type Canon struct {
	Handles []Handle
	Edges   []CanonEdge
}

// Stats supplies catalog frequencies for Fellegi–Sunter weights.
type Stats struct {
	// ValueFreq[propertyKey][valueKey] ≈ how often unrelated handles agree.
	ValueFreq map[string]map[string]float64
	// FanOut[signature.Key()] is the typical neighbor count for that edge.
	FanOut map[string]float64
}

// Fixed is a decided or already-promoted Subject→handle anchor.
type Fixed struct {
	SubjectID []byte
	HandleID  []byte
}

// Target is what Align proposes for one Subject.
type Target string

const (
	TargetHandle Target = "handle"
	TargetNew    Target = "new"
	TargetSkip   Target = "skip"
)

// Assessment is the strong / weak / no-match band.
type Assessment string

const (
	AssessStrong Assessment = "strong"
	AssessWeak   Assessment = "weak"
	AssessNone   Assessment = "none"
)

// Comparison is one Property outcome on the chosen (or top) candidate,
// suitable for drafting pins later (S9-43).
type Comparison struct {
	Property  match.Property
	Outcome   match.Outcome
	ValueType string
	Pinned    bool // drafted true when OutcomeAgree
	// Weight is the log-odds this outcome added (negative for a conflict).
	Weight float64
}

// Via is the layer edge that seeded this row's handle, when the walk
// reached it from an already-decided neighbor.
type Via struct {
	NeighborSubjectID []byte
	Signature         EdgeSignature
}

// Exhibit is one pin-able observation pair the loader attaches after Align.
// GroupLabel is empty for the Subject's own records and a neighbor label
// ("Birth · date") for a one-hop record.
type Exhibit struct {
	Property              match.Property
	Outcome               match.Outcome
	ValueType             string
	Pinned                bool
	Weight                float64
	GroupLabel            string
	IncomingObservationID []byte
	IncomingDisplay       string
	IncomingSource        string
	MemberObservationID   []byte
	MemberDisplay         string
	MemberSource          string
}

// Alternative is a runner-up handle suggestion.
type Alternative struct {
	HandleID []byte
	Ref      string
	Score    float64
}

// RowFlags call out conflicts and duplicates for the page.
type RowFlags struct {
	ConflictWithFixed bool
	PossibleDuplicate bool // another row already took this handle
}

// Row is one Subject's proposal.
type Row struct {
	SubjectID    []byte
	Kind         string
	Target       Target
	HandleID     []byte
	HandleRef    string
	Score        float64
	Assessment   Assessment
	Alternatives []Alternative
	Comparisons  []Comparison
	Exhibits     []Exhibit
	Reasons      []string
	Via          *Via
	Flags        RowFlags
}

// Proposal is Align's full output: one row per primary Subject.
type Proposal struct {
	Rows []Row
}
