package database

import (
	"errors"

	"github.com/mendahu/provenencia/core/database/graphcache"
)

// ErrWholeOnWrite is returned by ReadWhole on a write's graph: a graph-lifetime
// whole entry is built once per open catalog, never once per write.
var ErrWholeOnWrite = errors.New("database: a graph-lifetime cache is not read on a write's graph")

// ReadWhole returns a graph-lifetime whole entry's value for an open catalog's
// graph, building it on the first read. A write's graph is ErrWholeOnWrite; a
// nil graph is ErrClosed.
func ReadWhole[T any](w graphcache.Whole[T], g *graphcache.Graph) (T, error) {
	var zero T
	if g == nil {
		return zero, ErrClosed
	}
	v, ok, err := w.Get(g)
	if err != nil {
		return zero, err
	}
	if !ok {
		return zero, ErrWholeOnWrite
	}
	return v, nil
}
