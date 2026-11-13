package graphfill

import (
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

// SubjectRecord is one Subject of an Evidence graph: the Subject, its accepted
// claim (zero when it has none; the catalog allows one per Subject) and its
// Observations in ref order (nil when it has none). A reader treats it as
// read-only.
type SubjectRecord struct {
	Subject      graphcache.EvidenceSubject
	Accepted     graphcache.EvidenceMember
	Observations []graphcache.EvidenceObservation
}

// GraphRecords is an Evidence graph's Subject records, in the loader's Subject
// order, and the graph's Subject labels. Everything it returns is shared and
// read-only.
type GraphRecords struct {
	GraphID  []byte
	subjects []SubjectRecord
	byID     map[string]int
	labels   map[string]string
}

// newGraphRecords is the one constructor of a graph's records, from the rows
// the loaders return.
func newGraphRecords(graphID []byte, subjects []graphcache.EvidenceSubject, observations []graphcache.EvidenceObservation, members []graphcache.EvidenceMember) *GraphRecords {
	r := &GraphRecords{
		GraphID:  append([]byte(nil), graphID...),
		subjects: make([]SubjectRecord, len(subjects)),
		byID:     make(map[string]int, len(subjects)),
		labels:   graphcache.SubjectLabels(subjects),
	}
	for i, s := range subjects {
		r.subjects[i].Subject = s
		r.byID[string(s.ID)] = i
	}
	for _, m := range members {
		if i, ok := r.byID[string(m.SubjectID)]; ok && r.subjects[i].Accepted.ClaimID == nil {
			r.subjects[i].Accepted = m
		}
	}
	for _, o := range observations {
		if i, ok := r.byID[string(o.SubjectID)]; ok {
			r.subjects[i].Observations = append(r.subjects[i].Observations, o)
		}
	}
	return r
}

// Subject is the record of the Subject with this id, and false when the graph
// has none.
func (r *GraphRecords) Subject(id []byte) (SubjectRecord, bool) {
	if r == nil {
		return SubjectRecord{}, false
	}
	i, ok := r.byID[string(id)]
	if !ok {
		return SubjectRecord{}, false
	}
	return r.subjects[i], true
}

// Labels maps each Subject's id (as a string) to its label. The map is shared.
func (r *GraphRecords) Labels() map[string]string {
	if r == nil {
		return nil
	}
	return r.labels
}

// Entity is the canonical entity the Subject's accepted claim names, nil when
// the Subject has no accepted claim or is not on the graph.
func (r *GraphRecords) Entity(subjectID []byte) []byte {
	rec, ok := r.Subject(subjectID)
	if !ok {
		return nil
	}
	return rec.Accepted.EntityID
}

// maxGraphRecords is how many graphs' records stay held: 40% of the 10,000
// graphs the batch's Scale files, so a catalog of that size keeps the graphs a
// researcher is working in, and a larger one evicts the least recently used.
const maxGraphRecords = 4_096 // decided: user, 2026-11-13

// Records is each Evidence graph's Subject records, keyed by graph id. A graph
// that does not exist holds empty records until a commit that names it
// (Set.EvidenceGraphs) drops them. They are filled from the loaded graph when
// it is held, else from the Subject, Observation and member rows, and checked
// against a fresh read of those rows.
var Records graphcache.Keyed[*GraphRecords]

func init() {
	Records = graphcache.RegisterKeyed(graphcache.KeyedEntry[*GraphRecords]{
		Name:  "graphfill.records",
		Apply: graphcache.DropNamedGraphs[*GraphRecords],
		Check: checkRecords,
		Cap:   maxGraphRecords,
	})
}

// RecordsOf returns the records of each graph in graphIDs, keyed by graph id,
// filling those not held in one fill. A nil graph is database.ErrClosed.
func RecordsOf(g *graphcache.Graph, graphIDs [][]byte) (map[string]*GraphRecords, error) {
	if g == nil {
		return nil, database.ErrClosed
	}
	return Records.GetMany(g, graphIDs, func(missing [][]byte) (map[string]*GraphRecords, error) {
		out := make(map[string]*GraphRecords, len(missing))
		for _, id := range missing {
			if eg, ok := PeekEvidenceGraph(g, id); ok {
				if !eg.Found {
					out[string(id)] = newGraphRecords(id, nil, nil, nil)
					continue
				}
				out[string(id)] = newGraphRecords(id, eg.Subjects, eg.Observations, eg.Members)
				continue
			}
			subjects, observations, members, err := loadGraphRows(g, id)
			if err != nil {
				return nil, err
			}
			out[string(id)] = newGraphRecords(id, subjects, observations, members)
		}
		return out, nil
	})
}

// checkRecords builds the graph's records again from the rows, never from the
// loaded graph or the entry, and compares.
func checkRecords(g *graphcache.Graph, graphID []byte, stored *GraphRecords) error {
	subjects, observations, members, err := loadGraphRows(g, graphID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(stored, newGraphRecords(graphID, subjects, observations, members)) {
		return fmt.Errorf("graphfill: the stored records of %x do not match a fresh read", graphID)
	}
	return nil
}
