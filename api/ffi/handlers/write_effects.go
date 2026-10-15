package handlers

import (
	"github.com/mendahu/provenencia/api/proto/engine"
	"github.com/mendahu/provenencia/core/writes"
)

func writeEffectsProto(res writes.Result) *engine.WriteEffects {
	set := res.Effects
	return &engine.WriteEffects{
		Handles:       idStrings(set.Handles),
		Sources:       idStrings(set.Sources),
		SourceScopes:  idStrings(set.SourceScopes),
		Vocabulary:    set.Vocabulary,
		Structure:     set.Structure,
		HeaderHandles: idStrings(set.HeaderDeps),
		HandleScopes:  idStrings(set.HandleScopes),
		Revision:      res.Revision,
		RevisionAt:    res.At,
	}
}

func idStrings(ids [][]byte) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if s := uuidString(id); s != "" {
			out = append(out, s)
		}
	}
	return out
}
