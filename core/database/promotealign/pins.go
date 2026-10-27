package promotealign

import (
	"slices"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
)

// LinePinID is the id of one member Observation a line offers to pin: the
// incoming Observation and the member's.
func LinePinID(l claimevidence.Line, p claimevidence.Pin) string {
	return "l:" + database.UUIDString(l.Incoming.Ref) + ":" + database.UUIDString(p.Observation)
}

// BundlePinID is the id of one member's side of a relationship: the neighbor,
// the link's key and the member's bridge Subject.
func BundlePinID(rel claimevidence.Relationship, b claimevidence.Bundle) string {
	return "b:" + database.UUIDString(rel.Neighbor) + ":" + rel.Key.String() + ":" + database.UUIDString(b.Bridges[0])
}

// PinOffer is one pin the evidence offers: the sheet shows it, Checked says
// whether it starts pinned, and Set is what Done writes for it.
type PinOffer struct {
	ID      string
	Checked bool
	Set     promote.PinSet
}

// PinOffers are every pin the evidence offers, in the order the sheet shows
// them: the row's own lines, then each relationship's bundles followed by its
// lines. Every line and relationship offers its pins, a conflict included,
// unchecked. This is the one place that decides which pins start checked; the
// RPC sends the offers and Apply writes them, so the sheet and a Done that
// never opened it agree.
func PinOffers(e claimevidence.Evidence) []PinOffer {
	var out []PinOffer
	lines := func(sections []claimevidence.Section) {
		for _, s := range sections {
			for _, l := range s.Lines {
				for _, p := range l.Members {
					out = append(out, PinOffer{
						ID:      LinePinID(l, p),
						Checked: l.Checked,
						Set:     promote.PinSet{Member: p.Subject, Observations: [][]byte{l.Incoming.Ref, p.Observation}},
					})
				}
			}
		}
	}
	lines(e.Sections)
	for _, rel := range e.Relationships {
		for _, b := range rel.Members {
			out = append(out, PinOffer{
				ID:      BundlePinID(rel, b),
				Checked: rel.Checked,
				Set:     promote.PinSet{Member: b.Member, Observations: append(slices.Clone(rel.Incoming), b.Observations...)},
			})
		}
		lines(rel.Sections)
	}
	return out
}

// ApplyRow is one row of a Done with the pins the researcher chose.
type ApplyRow struct {
	promote.BatchRow
	DefaultPins bool     // the sheet was never opened: pin what the evidence checks
	PinIDs      []string // the pins the researcher left checked
}

// ResolvePins turns each handle row's pin choice into the pin sets Done
// writes, from the page's final targets: the other rows' targets are the
// neighbors. A row that asks for DefaultPins gets every checked offer; a row
// with PinIDs gets the offers those ids name, and an id the evidence no longer
// offers is dropped. A row already filed on the graph gets none. A New or Skip
// row with PinIDs, or a row with both DefaultPins and PinIDs, is
// promote.ErrInvalid. Rows keep their order; sets keep the offers' order.
func ResolvePins(c *database.Catalog, graphID []byte, rows []ApplyRow) ([]promote.BatchRow, error) {
	out := make([]promote.BatchRow, len(rows))
	var (
		in    inputs
		r     *evidenceReader
		eg    graphcache.EvidenceGraph
		ready bool
	)
	for i, row := range rows {
		out[i] = row.BatchRow
		out[i].PinSets = nil
		if len(row.PinIDs) > 0 && (row.DefaultPins || row.Target != promote.TargetHandle) {
			return nil, promote.ErrInvalid
		}
		if row.Target != promote.TargetHandle || (!row.DefaultPins && len(row.PinIDs) == 0) {
			continue
		}
		if !ready {
			var err error
			if in, err = loadInputs(c, graphID); err != nil {
				return nil, err
			}
			if r, err = newEvidenceReader(c.Graph(), in); err != nil {
				return nil, err
			}
			if eg, err = c.Graph().EvidenceGraph(in.graphID); err != nil {
				return nil, err
			}
			ready = true
		}
		if slices.ContainsFunc(eg.Members, func(m graphcache.EvidenceMember) bool { return string(m.SubjectID) == string(row.SubjectID) }) {
			continue
		}
		req := EvidenceRequest{GraphID: graphID, SubjectID: row.SubjectID, HandleID: row.EntityID}
		for j, other := range rows {
			if j != i {
				req.Neighbors = append(req.Neighbors, graphalign.Fixed{
					SubjectID: other.SubjectID, HandleID: other.EntityID, Target: graphalign.Target(other.Target),
				})
			}
		}
		e, err := evidence(c.Graph(), in, r, req)
		if err != nil {
			return nil, err
		}
		chosen := map[string]bool{}
		for _, id := range row.PinIDs {
			chosen[id] = true
		}
		for _, o := range PinOffers(e) {
			if row.DefaultPins && o.Checked || chosen[o.ID] {
				out[i].PinSets = append(out[i].PinSets, o.Set)
				delete(chosen, o.ID)
			}
		}
	}
	return out, nil
}
