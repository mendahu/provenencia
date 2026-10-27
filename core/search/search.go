// Package search is the catalog omnibar engine: kind registry, scoring, and
// FTS5 retrieval over projected search documents.
package search

import (
	"context"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/workspacelocation"
)

// Kind identifiers for SearchHit.kind (stable across retrieval backends).
const (
	KindSource        = "source"
	KindSourceType    = "source_type"
	KindMetadataField = "metadata_field"
	// Conclusion handles (S9-34a). Not in the omnibar's default set until
	// S9-35; callers ask for them with Query.Kinds.
	KindPerson = vocabkey.Person
	KindEvent  = vocabkey.Event
	KindPlace  = vocabkey.Place
	// KindEvidenceGraph is declared by catalogmodel (S12-15).
	KindEvidenceGraph = string(catalogmodel.KindEvidenceGraph)
)

// DefaultHitLimit caps SearchCatalog responses.
const DefaultHitLimit = 50

// Query is one SearchCatalog request (after proto mapping).
type Query struct {
	Text     string
	Location workspacelocation.Location
	Limit    int
	// Kinds restricts hits to these kinds. Empty means the omnibar's default
	// set (KindSpec.DefaultInEverything). Unknown kinds are ignored.
	Kinds []string
}

// Hit is one navigable search result.
type Hit struct {
	Kind             string
	ID               string
	Ref              string
	Title            string
	Subtitle         string
	MatchReason      string // stable field code; not localized UI copy
	MatchSnippet     string // optional raw snippet for body/rollup matches
	Location         workspacelocation.Location
	ThumbnailRelPath string  // Source cover when a derivative already exists
	IconKey          string  // Type icon (source type, or source's type)
	MemberCount      int     // accepted members, for handle kinds
	Score            float64 // internal; not exposed on the wire
}

// Searcher retrieves and ranks hits.
type Searcher interface {
	Search(ctx context.Context, c *database.Catalog, q Query) ([]Hit, error)
}

// Engine is the FFI entry point: registry + Searcher.
type Engine struct {
	Searcher Searcher
}

// DefaultEngine returns an Engine with the FTS5 Searcher.
func DefaultEngine() *Engine {
	return &Engine{Searcher: NewFTSSearcher()}
}

// Search runs the configured Searcher (empty/short query → no hits).
func (e *Engine) Search(ctx context.Context, c *database.Catalog, q Query) ([]Hit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if e == nil || e.Searcher == nil {
		e = DefaultEngine()
	}
	return e.Searcher.Search(ctx, c, q)
}
