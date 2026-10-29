package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/writes"
)

func writeEffectsProto(res writes.Result) *engine.WriteEffects {
	set := res.Effects
	return &engine.WriteEffects{
		Handles:        idStrings(set.Handles),
		EvidenceGraphs: idStrings(set.EvidenceGraphs),
		SourceScopes:   idStrings(set.SourceScopes),
		Reads:          readProtos(set.Reads),
		Lists:          listStrings(set.Lists),
		Structure:      set.Structure,
		HeaderHandles:  idStrings(set.HeaderDeps),
		HandleScopes:   idStrings(set.HandleScopes),
		Revision:       res.Revision,
		RevisionAt:     res.At,
	}
}

func readProtos(reads []effects.ReadIDs) []*engine.WriteEffectsRead {
	if len(reads) == 0 {
		return nil
	}
	out := make([]*engine.WriteEffectsRead, len(reads))
	for i, r := range reads {
		out[i] = &engine.WriteEffectsRead{Read: string(r.Read), Ids: idStrings(r.IDs)}
	}
	return out
}

func idStrings(ids [][]byte) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if s := database.UUIDString(id); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func listStrings(lists []effects.List) []string {
	if len(lists) == 0 {
		return nil
	}
	out := make([]string, len(lists))
	for i, l := range lists {
		out[i] = string(l)
	}
	return out
}
