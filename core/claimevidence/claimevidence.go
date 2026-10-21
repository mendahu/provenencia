// Package claimevidence is the claim evidence view's policy: which lines a
// Promote row shows for one target entity, which are pinnable and which start
// checked, and how a relationship that corresponds through segment compare
// bundles the Observations behind it. Pure: no I/O, no catalog, no numbers;
// the evidence section, layers 1, 3 and 5's rules and every input (the row's
// and its neighbors' values, the entity's accepted members) come in as
// arguments. It composes layer 2 (propertycompare), layer 3's Side and Points
// and layer 5's Compare; it never uses matching's scores to decide a line.
package claimevidence

import (
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Rules is the match registry's evidence section.
type Rules struct {
	// Checked lists the relations whose agreeing lines start checked.
	Checked []valuecompare.Relation
	// Pinnable lists the relations whose partial lines can be pinned but
	// start unchecked.
	Pinnable []valuecompare.Relation
	// Defining lists, per Subject type, the properties bundled into a
	// relationship line: the neighbor's own facts that say which one it is.
	Defining map[vocabkey.SubjectType][]vocabkey.PropertyKey
}

// Member is one accepted member of the target entity: a Subject on a Source,
// with its values. Each value's Ref is its Observation id.
type Member struct {
	Subject, Source []byte
	Values          subjectcompare.Values
}

// Pin is one member Observation a line would pin, and where it came from.
// Value is the member's value the Observation holds.
type Pin struct {
	Subject, Observation, Source []byte
	Value                        propertycompare.Value
}

// Line is one incoming Observation against the members that say something
// about the same property.
type Line struct {
	Property      vocabkey.PropertyKey  // the incoming value's property
	Incoming      propertycompare.Value // Ref is the incoming Observation
	Outcome       propertycompare.Outcome
	PropertyCause propertycompare.Cause
	Relation      valuecompare.Relation
	Cause         valuecompare.Cause
	Points        float64
	Pinnable      bool
	Checked       bool
	Members       []Pin // one per member, member order
}

// Section is the lines for one property. Key is the lead property a pooled or
// period property counts under.
type Section struct {
	Key   vocabkey.PropertyKey
	Lines []Line
}

// Bundle is one member's side of a relationship: the Observations behind a
// bridge between a member and a member of the neighbor's entity.
type Bundle struct {
	Member, Neighbor, Source []byte
	Bridge                   []byte // the member's filed bridge Subject
	Observations             [][]byte
}

// Relationship is one link of the row's Subject that corresponds, through
// segment compare, to a link between the entity's members and its neighbor's.
type Relationship struct {
	Key      vocabkey.BridgeKey   // from the row Subject's end
	Neighbor []byte               // the layer neighbor
	Kind     vocabkey.SubjectType // the neighbor's kind
	Outcome  propertycompare.Outcome
	Cause    segmentcompare.Cause
	Points   float64 // segment evidence
	Pinnable bool
	Checked  bool
	Incoming [][]byte // bundled incoming Observations
	Members  []Bundle
	Sections []Section // the neighbor's other properties
}

// Evidence is everything the sheet shows for one row.
type Evidence struct {
	Kind          vocabkey.SubjectType // the row's kind
	Sections      []Section
	Relationships []Relationship
	Sources       map[string]string // Source id → title; the loader fills it
}

// LayerLink is a link between the row's Subject and a layer neighbor.
type LayerLink struct {
	Key          vocabkey.BridgeKey // from the row Subject's end
	Observations [][]byte           // on the layer bridge Subjects
}

// Bridge is one filed bridge Subject between a member and a member of the
// neighbor's entity.
type Bridge struct {
	Subject, Source, Member, Neighbor []byte
	Observations                      [][]byte
}

// CanonLink is a link between the entity and the neighbor's entity.
type CanonLink struct {
	Key     vocabkey.BridgeKey // from the handle's end
	Bridges []Bridge           // the association's accepted bridge Subjects
}

// Neighbor is one layer neighbor the request places on an entity.
type Neighbor struct {
	Subject []byte
	Kind    vocabkey.SubjectType
	Values  subjectcompare.Values // the layer neighbor's
	Handle  subjectcompare.Values // its entity's profile
	Members []Member              // its entity's accepted members
	Layer   []LayerLink
	Canon   []CanonLink
}
