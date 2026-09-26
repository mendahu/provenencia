// Package deleteimpact is the catalog delete-impact register: speakable inbound
// reports, extra gates, and owned-outbound / connection-facet release helpers.
package deleteimpact

import (
	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
)

// Kind identifiers for Impact / GetDeleteImpact (stable machine keys).
const (
	KindSource           Kind = "source"
	KindArtifact         Kind = "artifact"
	KindCitation         Kind = "citation"
	KindObservation      Kind = "observation"
	KindSubject          Kind = "subject"
	KindSourceType       Kind = "source_type"
	KindSourceField      Kind = "source_field"
	KindCredibilityGrade Kind = "source_credibility_grade"
	KindSubjectType      Kind = "subject_type"
	KindProperty         Kind = "property"
	KindPropertyTerm     Kind = "property_term"
	KindFile             Kind = "file"
	KindUser             Kind = "user"
	KindProject          Kind = "project"

	// Reserved child kinds for stub inbound probes (tables not in the catalog yet).
	KindSamenessClaim       Kind = "sameness_claim"
	KindReconciliationClaim Kind = "reconciliation_claim"
	KindNarrative           Kind = "narrative"
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

// Bucket tags one live FK or a logical predicate on an FK.
const (
	BucketResource        Bucket = "resource"
	BucketVocab           Bucket = "vocab"
	BucketFacet           Bucket = "facet"
	BucketOwnedOutbound   Bucket = "ownedOutbound"
	BucketPool            Bucket = "pool"
	BucketInfra           Bucket = "infra"
	BucketSkip            Bucket = "skip"
	BucketOptional        Bucket = "optional"
	BucketConnectionFacet Bucket = "connectionFacet"
)

const (
	listedCap = 20

	sectionSources       = "sources"
	sectionSourceTypes   = "source-types"
	sectionSourceFields  = "source-fields"
	sectionSubjectFields = "subject-fields"

	surfacePage             = "page"
	surfaceGraph            = "graph"
	surfaceCitationComposer = "citationComposer"

	originProvenencia = "provenencia"
	originUser        = "user"
	originPluginPref  = "plugin:"
)

// Kind is one registered catalog delete kind.
type Kind string

// Gate is one Impact outcome.
type Gate string

// Bucket is one FK / table classification.
type Bucket string

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
	SourceSurface        string
	Ref                  string
	Title                string
	SourceTitle          string
}

// Report is the speakable delete report (Impact is the function that builds it).
type Report struct {
	Allowed bool
	Gate    Gate
	Groups  []Group
}

// Group is one inbound resource edge.
type Group struct {
	Via    string
	Kind   Kind
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
