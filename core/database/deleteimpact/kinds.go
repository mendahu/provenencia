// Package deleteimpact is the catalog delete-impact register: speakable inbound
// reports, extra gates, and owned-outbound / connection-facet release helpers.
package deleteimpact

import (
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/workspacelocation"
)

// Gate is the Impact extra-gate / inbound outcome.
const (
	GateOK           Gate = "ok"
	GateInbound      Gate = "inbound"
	GateNotFound     Gate = "not_found"
	GateEdgeLocked   Gate = "edge_locked"
	GateInfra        Gate = "infra"
	GateOriginLocked Gate = "origin_locked"
	GateLastGraph    Gate = "last_graph"
)

const listedCap = 20

// Gate is one Impact outcome.
type Gate string

// ErrInvalid is returned for an unknown kind or malformed id.
var ErrInvalid = apperr.New(apperr.CodeDeleteImpactInvalid, apperr.KindUser)

// Report is the speakable delete report (Impact is the function that builds it).
type Report struct {
	Allowed bool
	Gate    Gate
	Groups  []Group
	// Cascades names rows that go with the target on erase without blocking it
	// (facet CASCADEs worth telling the researcher about). Refuse ignores them.
	Cascades []Group
}

// Group is one inbound resource edge.
type Group struct {
	Via    string
	Kind   catalogmodel.Kind
	Total  int
	Listed []Listed
}

// Listed is one named inbound row (list cap 20).
type Listed struct {
	ID       []byte
	Ref      string
	Title    string
	Location workspacelocation.Location
}
