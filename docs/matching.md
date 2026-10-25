# Matching

> **Spike 11 rebuilt matching.** The model, the terms and every knob are in [matching-model.md](matching-model.md). This page says where the running code is.

Matching decides whether two things are the same Person, Event or Place. It is a stack of pure layers over one registry, [`core/matchregistry`](../core/matchregistry); the [`add-match-rule`](../.agents/skills/add-match-rule/SKILL.md) skill is how to change one.

| Layer | Package | Says |
| --- | --- | --- |
| 1 Value compare | [`core/valuecompare`](../core/valuecompare) | How two values of one type relate |
| 2 Property compare | [`core/propertycompare`](../core/propertycompare) | Whether two sides agree on one property |
| 3 Subject compare | [`core/subjectcompare`](../core/subjectcompare) | How alike two Subjects of one type are: similarity and evidence |
| 4 Entity compare | [`core/subjectprofile`](../core/subjectprofile) then layer 3 | Whether a Subject belongs with an entity's conclusion |

## Where it runs

**Promote.** [`core/database/promotealign`](../core/database/promotealign) builds each side once per proposal: a Source Subject's with `subjectprofile.ForSubject` from its own Observations and the dated events it takes part in, a handle's with `subjectprofile.Build` from its reconciled values and its participations. [`core/graphalign`](../core/graphalign) scores every pair with `subjectcompare.Compare` and adds link support ([promote-graph-alignment.md](promote-graph-alignment.md)). A handle is a candidate when it is shortlisted and at least half alike (`graphalign.Candidates`), or reached through a link to an accepted pair.

**The reconciler** groups values with value compare under its own policy ([conclusion-reconciliation.md](conclusion-reconciliation.md)).

**Search, the search index and term dedupe** key text with `valuecompare.TextKey`.

## Not built yet

Segment compare (S11-14), the shortlist index (S11-15) and the layer 6 walk (S11-16). Merge hints are entity against entity, `subjectcompare.Compare` on two `Build` profiles; no surface uses them.
