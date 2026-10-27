# Matching

> **Spike 11 rebuilt matching.** The model, the terms and every knob are in [matching-model.md](matching-model.md). This page says where the running code is.

Matching decides whether two things are the same Person, Event or Place. It is a stack of pure layers over one registry, [`core/matchregistry`](../core/matchregistry); the [`add-match-rule`](../.agents/skills/add-match-rule/SKILL.md) skill is how to change one.

The six layers, their packages and what each says are in the stack ([`matching-model.md#the-stack`](matching-model.md#the-stack)).

## Where it runs

**Promote.** [`core/database/promotealign`](../core/database/promotealign) builds each side once per proposal: a Source Subject's with `subjectprofile.ForSubject` from its own Observations and the dated events it takes part in, a handle's with `subjectprofile.Build` from its reconciled values and its participations. [`core/graphalign`](../core/graphalign) scores every pair with `subjectcompare.Compare` and adds link support ([promote-graph-alignment.md](promote-graph-alignment.md)). A handle is a candidate when it is shortlisted with similarity at least `alignment.MinSimilarity` (`graphalign.Candidates`), or reached through a link to an accepted pair.

**The reconciler** groups values with value compare under its own policy ([conclusion-reconciliation.md](conclusion-reconciliation.md)).

**Search, the search index and term dedupe** key text with `valuecompare.TextKey`.

## Not built yet

Merge hints are entity against entity, `subjectcompare.Compare` on two `Build` profiles; no surface uses them.
