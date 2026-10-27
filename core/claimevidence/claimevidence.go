// Package claimevidence is the claim evidence view's policy: which lines a
// Promote row shows for one target entity, which lines start checked (every
// line can be pinned), and how a relationship that corresponds through segment
// compare bundles the Observations behind it. Pure: no I/O, no catalog, no numbers;
// the evidence section, layers 1, 3 and 5's rules and every input (the row's
// and its neighbors' values, the entity's accepted members) come in as
// arguments. It composes layer 2 (propertycompare), layer 3's Side and Points
// and layer 5's Compare; it never uses matching's scores to decide a line.
package claimevidence

import (
	"github.com/mendahu/provenencia/core/kinship"
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
	// ByEye lists the causes of an unknown line the researcher judges: the
	// line is kept, and never starts checked.
	ByEye []valuecompare.Cause
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
	Bridges                  [][]byte // every bridge Subject behind the link, the member's first
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
	Checked  bool
	Incoming [][]byte // bundled incoming Observations
	Members  []Bundle
	Sections []Section // the neighbor's other properties

	// Layer and Canon say how the row's Source and the entity record the fact
	// the line stands for. The zero Shape is direct. A layer pivot is a layer
	// Subject; a canon pivot is a handle.
	Layer, Canon kinship.Shape

	// NeighborEnd, NeighborTerm and NeighborTermNamesNeighbor read Key from the
	// neighbor's side (segmentcompare.ReadFrom): the endpoint the neighbor is
	// at, the term it carries, and whether that term names the neighbor.
	// NeighborTermProperty is the property the term belongs to (the bridge
	// type's disambiguation). The loader fills them after Build, from the
	// catalog's terms and the connect rules.
	NeighborEnd               vocabkey.PropertyKey
	NeighborTerm              vocabkey.BridgeTerm
	NeighborTermProperty      vocabkey.PropertyKey
	NeighborTermNamesNeighbor bool
}

// Inputs is what the loader reads beside the values: the catalog's property
// order for each kind.
type Inputs struct {
	Order map[vocabkey.SubjectType][]vocabkey.PropertyKey // each kind's bound properties, in binding order
	// Placed maps each layer Subject the request places on an entity to that
	// entity's handle: how kinship.CountedOnce pairs the two sides' pivots.
	Placed map[string][]byte
}

// Pivot is the event or parent a kinship pattern passes through: the Subject,
// its kind, and its values, whose defining properties a link through it bundles.
type Pivot struct {
	Subject []byte
	Kind    vocabkey.SubjectType
	Values  subjectcompare.Values
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
	Recording    kinship.Recording  // how the row's Source records the link
	Pivots       []Pivot            // the layer Subjects its shapes pass through
}

// Bridge is one bridge Subject behind a link between the entity and the
// neighbor's entity: a filed one joins a member and a member of the neighbor's
// entity; one behind an implied link joins a member or a neighbor member to the
// Pivot, the accepted member of the event or parent the link passes through.
// A recorded bridge has no Pivot.
type Bridge struct {
	Subject, Source, Member, Neighbor, Pivot []byte
	Observations                             [][]byte
}

// CanonLink is a link between the entity and the neighbor's entity.
type CanonLink struct {
	Key       vocabkey.BridgeKey // from the handle's end
	Bridges   []Bridge           // the accepted bridge Subjects behind the link
	Recording kinship.Recording  // how the entity records the link
	Pivots    []Pivot            // the pivot members its bridges reach, each once
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
