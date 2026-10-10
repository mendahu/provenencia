// Package conclusionheaders composes the row-shaped header of each canonical
// Person, Event, and Place from the auto-reconciler cache (Spike 9
// R4). One header composer per kind serves every surface that shows a handle
// as a row: lists, Promote's target picker, omnibar hits, later tree nodes.
//
// Headers are composed at read time, never stored, and set-based: a whole
// list is a fixed number of queries whatever its length. Go returns
// structures, including which naming-matrix rule titles an Event
// (eventtitle); the app formats text.
package conclusionheaders

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
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
	Name *namevalues.Value
	// NameValueCount is the number of displayed name values: names are one
	// structure, so at most 1. More would read as mixed, the extras as +N.
	NameValueCount int
	// Birth and Death are read off subject-role participations in a birth or
	// death event. Places are folded location chains at the event's date.
	Birth LifeFacts
	Death LifeFacts
}

// LifeFacts is a birth or a death composed from the canonical graph.
type LifeFacts struct {
	// Event is the birth or death event read; nil when none is linked.
	// Its page owns the date's Why. EventCount is how many such events
	// survive; more than one is a disagreement.
	Event      *canonicalentities.Entity
	EventCount int
	// Date is the event's date, else its start date.
	Date      *datevalues.Value
	DateCount int
	Places    []HeaderPlace
}

// HeaderPlace is one Place a walk reached (or a folded location chain).
// Names are kept toponyms in rank order. Parents is the hierarchical chain
// at the walk's date (nearest first); ParentsAreCandidates when several
// parents hold at once. Entity is the Place (the leaf when folded), whose
// page owns the names' Why.
type HeaderPlace struct {
	Entity               canonicalentities.Entity
	Names                []string
	Parents              []string
	ParentsAreCandidates bool
}
