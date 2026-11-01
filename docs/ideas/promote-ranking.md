# Promote ranking, rebuilt

**Status:** parked 2026-10-29 when Spike 12 closed. Promote's suggestions are still ranked by graph alignment's score, unseen since Spike 13's S13-23 ([`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) §5.3). This note collects the matching work that waits until that ranking is rebuilt, likely on the Evidence graph scoring of [Spike 13's requirements](../deployment-plan/spike-13/evidence-graph-comparison.md) once it has proven itself.

## The direction

Rank a Source Subject's candidate records with the same scoring the handle pages use to show how records connect, instead of a second number tuned only for proposing. Until then, graph alignment keeps ranking and nothing extends it.

## Work that waits for it

Each item is described in the second app-health audit, now at [`deployment-plan/spike-14/app-health-audit-2.md`](../deployment-plan/spike-14/app-health-audit-2.md).

| Item | Was | Problem |
| --- | --- | --- |
| B-2 | S12-41 | Setting a Source person to New or Skip, or another row taking a record, can make a one-cardinality conflict (their birth) vanish from the neighbors' scores. |
| B-8, B-9 | S12-43 (performance half) | A handle commit throws away the name index and every Source's ranked candidates, and profiles are rebuilt; see [`name-index-on-commit.md`](name-index-on-commit.md). Also: the graph cache's three warm slots (alias index, name format, match index) should be one keyed slot, and its summary display rows should be memo slots. |
| B-8, B-11 | S12-51 | Each proposal walks the whole connected canon and compares would-be-new rows against each other more than once; see [`narrower-canon-walk.md`](narrower-canon-walk.md). |
| B-19, B-20, B-22 | S12-46 | Names compare differently depending on how they are written: initials and aliases on folded tokens, NFC and compatibility forms, format characters. Also mirrored in Swift's `TextMatch`. |

Budgets for these were to be measured at Spike 12's Scale: 37,000 canonical records, 100 member Sources on one record, 100 new people on one Source, 6,000 events on one place.
