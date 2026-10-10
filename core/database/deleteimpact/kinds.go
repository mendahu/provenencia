// Package deleteimpact is the catalog delete-impact register: speakable inbound
// reports, extra gates, and owned-outbound / connection-facet release helpers.
package deleteimpact

import (
	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

// Gate is the Impact extra-gate / inbound outcome.
const (
	GateOK           Gate = "ok"
	GateInbound      Gate = "inbound"
	GateNotFound     Gate = "not_found"
	GateEdgeLocked   Gate = "edge_locked"
	GateInfra        Gate = "infra"
	GateOriginLocked Gate = "origin_locked"
)

const (
	listedCap = 20

	sectionSources     = "sources"
	sectionSourceTypes = "source-types"
	sectionMetadata    = "metadata"
	sectionProperties  = "properties"

	surfacePage             = "page"
	surfaceGraph            = "graph"
	surfaceCitationComposer = "citationComposer"

	originProvenencia = "provenencia"
	originUser        = "user"
	originPluginPref  = "plugin:"
)

// Gate is one Impact outcome.
type Gate string

// ErrInvalid is returned for an unknown kind or malformed id.
var ErrInvalid = apperr.New(apperr.CodeDeleteImpactInvalid, apperr.KindUser)

// Location is the Go form of proto WorkspaceLocation after S8-12.
type Location struct {
	Section              string
	SourceID             string
	FieldID              string
	TypeID               string
	SubjectID            string
	CitationID           string
	ArtifactID           string
	ObservationID        string
	ConnectFromSubjectID string
	ConnectToSubjectID   string
	ConnectBridgeTypeKey string
	// SubjectTypeKey is the Properties type-strip category (empty = All properties).
	SubjectTypeKey string
	// PropertyID is the Properties inspector row.
	PropertyID    string
	SourceSurface string
	Ref           string
	Title         string
	SourceTitle   string
}

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
	Location Location
}

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}
