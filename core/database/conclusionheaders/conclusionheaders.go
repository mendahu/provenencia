// Package conclusionheaders composes the row-shaped header of each canonical
// Person, Event, and Place from the auto-reconciler cache (Spike 9
// R4). One header composer per kind serves every surface that shows a handle
// as a row: lists, Promote's target picker, omnibar hits, later tree nodes.
//
// A header's stored row holds ids and keys; term and Property labels are
// filled on read, from the cached vocabulary. The row is stored on the catalog
// graph. A list read fills any row that is missing, in a fixed number of
// queries whatever its length.
// Go returns structures, including which naming-matrix rule titles an Event
// (eventtitle); the app formats text.
package conclusionheaders

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/namevalue"
)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// PersonHeader is one Person as a row.
type PersonHeader struct {
	Entity canonicalentities.Entity
	// Name is the displayed auto-reconciled name; nil when no member names
	// the Person.
	Name *namevalue.Value
	// NameValueCount is the number of displayed name values: names are one
	// structure, so at most 1. More would read as mixed, the extras as +N.
	NameValueCount int
	// Birth and Death are read off subject-role participations in a birth or
	// death event. Places are folded location chains at the event's date.
	Birth LifeFacts
	Death LifeFacts
	// Gender and SexAtBirth are the displayed gender and sex at birth terms;
	// the app chooses which colors a person.
	Gender     TermFacts
	SexAtBirth TermFacts
	// PictureImageFileID is the image of the picture the Person leads with
	// on its page and in the Persons list; nil when none. The preferred
	// portrait when its picture is in the gallery, else the earliest
	// (gallery.Lead).
	PictureImageFileID []byte
	// Updated is the handle's latest canonical_entity revision and its time.
	// Filled by List only; zero means none.
	Updated effects.Activity
}

// TermFacts is one term-valued Property of a header. Term is the rank-1 kept
// term, nil when none; Count is the number of kept values, more than one is a
// disagreement; Term label and origin and PropertyLabel are filled at read time.
type TermFacts struct {
	Term          *Term
	Count         int
	PropertyLabel string
}

// LifeFacts is a birth or a death composed from the canonical graph.
type LifeFacts struct {
	// Event is the birth or death event read; nil when none is linked.
	// A Person page reads its date's records through
	// conclusiondetails.PersonDerivedRows. EventCount is how many such events
	// survive; more than one is a disagreement.
	Event      *canonicalentities.Entity
	EventCount int
	// Date is the event's date, else its start date.
	Date      *datevalue.Value
	DateCount int
	Places    []HeaderPlace
}

// HeaderPlace is one Place a walk reached (or a folded location chain).
// Names are kept toponyms in rank order. Parents is the hierarchical chain
// at the walk's date (nearest first); ParentsAreCandidates when several
// parents hold at once. Entity is the Place (the leaf when folded), whose
// link records a Person page reads through conclusiondetails.PersonDerivedRows.
type HeaderPlace struct {
	Entity               canonicalentities.Entity
	Names                []string
	Parents              []string
	ParentsAreCandidates bool
}

// FieldOwner is one entity and the Property of it that a Person's life row
// reads. A Place row's owner is the event-place link's association and the
// Property of its place end; Place is the Place that end holds, with its
// own toponyms and no parents. Place is nil for a date owner.
type FieldOwner struct {
	Entity     canonicalentities.Entity
	PropertyID []byte
	Place      *HeaderPlace
}

// LifeOwners are the owners behind one life row pair: the date owners (one
// per surviving event, lead first) and the Place owners (each event–place link once).
type LifeOwners struct {
	Dates, Places []FieldOwner
}
