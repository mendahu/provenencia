package promotealign

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// pageBridges lists every unfiled bridge Subject on the graph, of a type
// connectrules registers, with the Subjects its endpoint Observations name and
// the terms its disambiguation Observations carry. It reads what the filer
// reads: provenencia-origin Observations of any polarity, in graph order. The
// pick among several terms is connectrules.Term's, not made here.
func pageBridges(eg graphcache.EvidenceGraph, vocab graphcache.Vocabulary) []graphalign.PageBridge {
	rules := map[string]connectrules.Bridge{}
	for _, key := range connectrules.BridgeTypeKeys() {
		if b, ok := connectrules.LookupBridge(key); ok {
			rules[key] = b
		}
	}
	filed := make(map[string]bool, len(eg.Members))
	for _, m := range eg.Members {
		filed[string(m.SubjectID)] = true
	}
	var out []graphalign.PageBridge
	at := map[string]int{}
	for _, s := range eg.Subjects {
		rule, ok := rules[s.TypeKey]
		if !ok || s.TypeOrigin != vocabkey.OriginProvenencia || filed[string(s.ID)] {
			continue
		}
		at[string(s.ID)] = len(out)
		out = append(out, graphalign.PageBridge{
			ID: s.ID, TypeKey: vocabkey.SubjectType(s.TypeKey), Ends: make([][][]byte, len(rule.Endpoints)),
		})
	}
	if len(out) == 0 {
		return nil
	}
	props := vocab.PropertiesByID()
	terms := vocab.TermsByID()
	for _, o := range eg.Observations {
		i, ok := at[string(o.SubjectID)]
		if !ok {
			continue
		}
		p, ok := props[string(o.PropertyID)]
		if !ok || p.Origin != vocabkey.OriginProvenencia {
			continue
		}
		b := &out[i]
		rule := rules[string(b.TypeKey)]
		for e, ep := range rule.Endpoints {
			if p.Key == ep.PropertyKey && len(o.ValueSubjectID) == 16 {
				b.Ends[e] = append(b.Ends[e], o.ValueSubjectID)
			}
		}
		if connectrules.HasDisambiguation(rule.Disambiguation) && p.Key == rule.Disambiguation && len(o.TermID) == 16 {
			if t, ok := terms[string(o.TermID)]; ok {
				b.Terms = append(b.Terms, vocabkey.BridgeTerm(t.Key))
			}
		}
	}
	return out
}

// pageEnds is where the page puts each Subject a bridge can name: what Done
// sends and nothing else. An already-promoted Subject (promoted, the handles
// mergeFixed anchored) is its handle whatever the scope. With only set, every
// other Subject is off the page. A caller's fixed row is its handle, New or
// Skip; any other row starts as the proposal starts it (a handle), or has no
// target yet. A Subject left out is not on the page.
func pageEnds(promoted map[string][]byte, layer evidencelayer.Layer, caller []graphalign.Fixed, rows []graphalign.Row, only []byte) map[string]graphalign.EndTarget {
	ends := map[string]graphalign.EndTarget{}
	decided := make(map[string]graphalign.Fixed, len(caller))
	for _, f := range caller {
		decided[string(f.SubjectID)] = f
	}
	for sk, handle := range promoted {
		ends[sk] = endTarget(graphalign.TargetHandle, handle)
	}
	if len(only) > 0 {
		for _, s := range layer.Subjects {
			sk := string(s.ID)
			if _, anchored := ends[sk]; !anchored && sk != string(only) {
				ends[sk] = graphalign.EndTarget{Waits: graphalign.ConnectionNotOnPage}
			}
		}
	}
	started := make(map[string]graphalign.Row, len(rows))
	for _, r := range rows {
		started[string(r.SubjectID)] = r
	}
	for _, s := range layer.Subjects {
		sk := string(s.ID)
		if _, ok := ends[sk]; ok {
			continue
		}
		if f, ok := decided[sk]; ok {
			if f.Target == "" {
				f.Target = graphalign.TargetHandle
			}
			ends[sk] = endTarget(f.Target, f.HandleID)
			continue
		}
		start := started[sk].Start()
		ends[sk] = endTarget(start.Target, start.HandleID)
	}
	return ends
}

// endTarget is where one Subject lands: its handle, New, a wait for Skip, or
// a wait for a target not set yet.
func endTarget(t graphalign.Target, handleID []byte) graphalign.EndTarget {
	switch t {
	case graphalign.TargetHandle:
		return graphalign.EndTarget{Target: graphalign.TargetHandle, HandleID: handleID}
	case graphalign.TargetNew:
		return graphalign.EndTarget{Target: graphalign.TargetNew}
	case graphalign.TargetSkip:
		return graphalign.EndTarget{Waits: graphalign.ConnectionSkipped}
	}
	return graphalign.EndTarget{Waits: graphalign.ConnectionUnset}
}
