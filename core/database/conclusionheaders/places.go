package conclusionheaders

import (
	"time"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
)

// PlaceRelationshipKind is how a related Place appears on a Place detail.
const (
	RelPartOf      = "part_of"
	RelContains    = "contains"
	RelPredecessor = "predecessor"
	RelSuccessor   = "successor"
)

// PlaceRelationship is one related Place for a detail section (S9-40).
type PlaceRelationship struct {
	Entity    canonicalentities.Entity
	Title     string
	Kind      string
	StartDate *datevalues.Value // membership span, or the other place's period
	EndDate   *datevalues.Value
}

// PlaceHeader is one Place as a row. Names are the kept toponyms in rank
// order (a multi-valued Property: Montréal and Montreal are both names).
// StartDate / EndDate are the Place's period. Parents is today's hierarchical
// chain (nearest first); ParentsAreCandidates is true when several parents
// hold at once (join with "or"). Kind stays empty (place_nature deferred).
// PartOf / Contains / Predecessors / Successors are filled only for detail
// (AttachPlaceRelationships); list headers leave them empty.
type PlaceHeader struct {
	Entity               canonicalentities.Entity
	Names                []string
	StartDate            *datevalues.Value
	EndDate              *datevalues.Value
	Kind                 string
	Parents              []string
	ParentsAreCandidates bool
	PartOf               []PlaceRelationship
	Contains             []PlaceRelationship
	Predecessors         []PlaceRelationship
	Successors           []PlaceRelationship
}

// TodayDate is a Gregorian point for "today's chain" on Place rows.
func TodayDate() datevalues.Value {
	now := time.Now()
	y, m, d := now.Date()
	month, day := int(m), d
	return datevalues.Value{
		Kind: datevalues.KindPoint, Calendar: "gregorian",
		StartYear: &y, StartMonth: &month, StartDay: &day,
	}
}
