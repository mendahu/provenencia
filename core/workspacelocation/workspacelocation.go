// Package workspacelocation is the Go form of proto WorkspaceLocation: the
// navigable place a search hit or a Delete impact row opens in the app.
package workspacelocation

import "strings"

// Section values match macOS WorkspaceSection raw values.
const (
	Sources     = "sources"
	SourceTypes = "source-types"
	Metadata    = "metadata"
	Properties  = "properties"
	Files       = "files"
	Persons     = "persons"
	Events      = "events"
	Places      = "places"
)

// Surface values match macOS SourceSurface raw values.
const (
	SurfacePage             = "page"
	SurfaceGraph            = "graph"
	SurfaceCitationComposer = "citationComposer"
)

// Location is the navigable place for go(to:).
type Location struct {
	Section  string
	SourceID string
	// GraphID is the Evidence graph of a graph or citationComposer location.
	GraphID              string
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
	PropertyID string
	// EntityID is the Person, Event or Place of a Conclusion location.
	EntityID      string
	SourceSurface string
	Ref           string
	Title         string
	SourceTitle   string
}

// Graph is an Evidence graph's canvas. A graph location's Ref and Title name
// the Source, as the app's do (the canvas breadcrumb falls back to Title for
// the Source crumb); the graph's own name comes from its id. Title falls back
// to the Source's ref when the Source has no title.
func Graph(sourceID, graphID, sourceRef, sourceTitle string) Location {
	title := strings.TrimSpace(sourceTitle)
	if title == "" {
		title = sourceRef
	}
	return Location{
		Section:       Sources,
		SourceID:      sourceID,
		GraphID:       graphID,
		SourceSurface: SurfaceGraph,
		Ref:           sourceRef,
		Title:         title,
		SourceTitle:   sourceTitle,
	}
}
