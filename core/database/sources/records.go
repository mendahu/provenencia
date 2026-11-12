package sources

import (
	"database/sql"
	"errors"
	"reflect"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

// ErrRecordsOnWrite is returned by RecordsOf on a write's graph: the records
// are built once per open catalog, never once per write.
var ErrRecordsOnWrite = errors.New("sources: records are not read on a write's graph")

// Record is a Source's words, held once: every answer that names a Source
// composes them from here.
type Record struct {
	ID           []byte
	Ref          string
	Title        string
	SourceTypeID []byte
	// CredibilityKey is the key of the Source's credibility grade, "" with no
	// assessment. Credibility is that grade's order relative to the default.
	CredibilityKey string
	Credibility    int
	// ArtifactLabels maps each Artifact id to its label as stored. Never nil.
	ArtifactLabels map[string]string
}

// RecordIndex holds every Source's Record and finds the Source of an Artifact.
// It has no order; a reader treats it as read-only.
type RecordIndex struct {
	bySource         map[string]Record
	sourceOfArtifact map[string]string
}

// Source returns the Record of the Source with this id.
func (x *RecordIndex) Source(id []byte) (Record, bool) {
	r, ok := x.bySource[string(id)]
	return r, ok
}

// Artifact returns the Record of the Artifact's Source and the Artifact's label.
func (x *RecordIndex) Artifact(id []byte) (Record, string, bool) {
	sid, ok := x.sourceOfArtifact[string(id)]
	if !ok {
		return Record{}, "", false
	}
	r, ok := x.bySource[sid]
	if !ok {
		return Record{}, "", false
	}
	return r, r.ArtifactLabels[string(id)], true
}

var sqlRecordSources = `SELECT s.id, s.ref, s.title, s.source_type_id, COALESCE(sg.key, ''), ` +
	sourcecredibilitygrades.RelativeOrderColumn + `
		FROM sources s
		` + sourcecredibilitygrades.RelativeOrderJoins("s.id")

const sqlRecordArtifacts = `SELECT id, source_id, label FROM artifacts`

// Records is the Source records entry: one Record per Source, built on the
// first read of an open catalog and patched from effects.Set.SourceScopes.
var Records graphcache.Whole[*RecordIndex]

func init() {
	Records = graphcache.RegisterWhole(graphcache.WholeEntry[*RecordIndex]{
		Name:     "sources.records",
		Lifetime: graphcache.LifeGraph,
		Build: func(g *graphcache.Graph) (*RecordIndex, error) {
			recs, err := loadRecords(g, nil)
			if err != nil {
				return nil, err
			}
			return newRecordIndex(recs), nil
		},
		Apply: func(g *graphcache.Graph, w graphcache.Whole[*RecordIndex], _ [][]byte, set effects.Set) error {
			if len(set.SourceScopes) == 0 {
				return nil
			}
			idx, ok, err := w.Get(g)
			if err != nil || !ok {
				return err
			}
			fresh, err := loadRecords(g, set.SourceScopes)
			if err != nil {
				return err
			}
			// Every named Source leaves before any returns, so an Artifact that
			// moved between two named Sources keeps its new mapping.
			for _, id := range set.SourceScopes {
				if old, ok := idx.bySource[string(id)]; ok {
					for aid := range old.ArtifactLabels {
						delete(idx.sourceOfArtifact, aid)
					}
					delete(idx.bySource, string(id))
				}
			}
			for _, id := range set.SourceScopes {
				if rec, ok := fresh[string(id)]; ok {
					idx.bySource[string(id)] = rec
					for aid := range rec.ArtifactLabels {
						idx.sourceOfArtifact[aid] = string(id)
					}
				}
			}
			w.Set(g, idx)
			return nil
		},
		Check: func(g *graphcache.Graph, stored *RecordIndex) error {
			recs, err := loadRecords(g, nil)
			if err != nil {
				return err
			}
			want := newRecordIndex(recs)
			if !reflect.DeepEqual(stored.bySource, want.bySource) ||
				!reflect.DeepEqual(stored.sourceOfArtifact, want.sourceOfArtifact) {
				return errors.New("sources: the stored Source records do not match the sources table")
			}
			return nil
		},
	})
}

// RecordsOf returns the Source records of an open catalog's graph, building
// them on the first read. A write's graph is ErrRecordsOnWrite; a nil graph is
// database.ErrClosed.
func RecordsOf(g *graphcache.Graph) (*RecordIndex, error) {
	if g == nil {
		return nil, database.ErrClosed
	}
	idx, ok, err := Records.Get(g)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRecordsOnWrite
	}
	return idx, nil
}

func newRecordIndex(recs map[string]Record) *RecordIndex {
	x := &RecordIndex{bySource: recs, sourceOfArtifact: map[string]string{}}
	for sid, r := range recs {
		for aid := range r.ArtifactLabels {
			x.sourceOfArtifact[aid] = sid
		}
	}
	return x
}

// loadRecords reads the Records of the Sources with these ids, or of every
// Source when ids is nil. A Source with no row is absent from the result.
func loadRecords(q graphcache.Querier, ids [][]byte) (map[string]Record, error) {
	out := map[string]Record{}
	scanSource := func(rows *sql.Rows) error {
		var r Record
		if err := rows.Scan(&r.ID, &r.Ref, &r.Title, &r.SourceTypeID, &r.CredibilityKey, &r.Credibility); err != nil {
			return err
		}
		r.ArtifactLabels = map[string]string{}
		out[string(r.ID)] = r
		return nil
	}
	scanArtifact := func(rows *sql.Rows) error {
		var aid, sid []byte
		var label string
		if err := rows.Scan(&aid, &sid, &label); err != nil {
			return err
		}
		if r, ok := out[string(sid)]; ok {
			r.ArtifactLabels[string(aid)] = label
		}
		return nil
	}
	if ids == nil {
		if err := queryEach(q, sqlRecordSources, scanSource); err != nil {
			return nil, err
		}
		if err := queryEach(q, sqlRecordArtifacts, scanArtifact); err != nil {
			return nil, err
		}
		return out, nil
	}
	if err := sqlbatch.Query(q, sqlRecordSources+` WHERE s.id IN (`, `)`, nil, ids, scanSource); err != nil {
		return nil, err
	}
	if err := sqlbatch.Query(q, sqlRecordArtifacts+` WHERE source_id IN (`, `)`, nil, ids, scanArtifact); err != nil {
		return nil, err
	}
	return out, nil
}

func queryEach(q graphcache.Querier, query string, scan func(*sql.Rows) error) error {
	rows, err := q.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
