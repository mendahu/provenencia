# Move an Artifact to another Source

**Status:** Spike 15 design input, not yet cut into PRs ([`README.md`](README.md)); decisions recorded 2026-11-09; planned with the Source schema pass ([`citation-formats.md`](citation-formats.md), [`source-provenance-date.md`](source-provenance-date.md)) (user, 2026-11-01). Once Sources carry the fields a citation needs, which Source an Artifact sits under matters more, and researchers will want to re-file Artifacts they already built Evidence graphs on.

## Problem

A researcher files an Artifact under one Source, builds an Evidence graph on it (Subjects, Observations, Citations, identity claims to the canon), and later wants their Sources organized differently: the obituary belongs under the newspaper issue it came from, two census images were filed as one Source and should be two, a scan was attached to the wrong certificate. Today the only way is to rebuild the graph by hand under the other Source and delete the old one, which loses the Citations, the regions drawn on the pages and the claims to canonical people.

## The feature

**Decisions (user, 2026-11-09).**

- **The unit is a bundle.** An Artifact and the Evidence graphs built on it move together: a graph cannot move without its Artifacts, because its Citations point at them. The bundle is the closure: the Artifacts picked, every graph linked to any of them, and every other Artifact linked to those graphs, repeated until nothing new joins.
- **One flow, wherever it starts.** **Move to another Source…** is offered on an Artifact (its row on the Source page, or its menu) and on an Evidence graph. A preflight reads the bundle and the confirm shows it in full: "These Artifacts and graphs move together", each listed, with the target picked. Moving the whole bundle is the default and only option; there is no refuse-and-name path and no way to split a shared graph.
- **Nothing is copied.** Every row keeps its id and ref, so identity claims, Conclusions and the provenance sheet keep pointing at the same records.
- **Source-level values are not looked at.** The move does not compare, copy or warn about the two Sources' metadata. After it, citations simply read the target Source's title, date and metadata and its type's citation formats ([`evidence-explained-citation-map.md`](evidence-explained-citation-map.md) § 6.3). Grouped citations on the provenance sheet and the Person page change accordingly; that is expected.
- **Artifact and graph metadata travel for free.** `artifact_metadata`, `evidence_graph_metadata` and their layout rows hang off `artifact_id` and `graph_id` (map § 6.6), so the copy details and item numbers move with the bundle. If the target's type suggests different fields at those levels, filled values stay as extra fields.
- **No undo** in this spike. Moving the bundle back is the way back.

### Combining Sources

Two Sources that should be one (two pages of a census filed separately) are combined by moving each bundle of one into the other: the move already targets an existing Source. **Merge Sources** would be a convenience over that: move every bundle of Source B into Source A in one confirm, then offer to delete the emptied B. Whether it ships in this spike is open; it needs no engine work beyond the move and the existing Source delete.

**Splitting a Source** (move to a **new** Source, prefilled by Copy Source) is deferred: with the cited item on the Evidence graph, one census page Source with a graph per household already gives each household its own citation (map § 6.9), so splitting is no longer needed for citations.

## What it touches in the schema

Today (S12-03, S14-10, migrations `000055` and `000059`):

- `artifacts.source_id` names the Artifact's Source.
- `evidence_graphs.source_id` names the graph's Source; a Source keeps at least one graph (`ErrLastGraph`).
- `evidence_graph_artifacts` links graphs and Artifacts **many to many**, within one Source.
- `subjects` carry both `graph_id` and `source_id`, held together by the composite key `(graph_id, source_id) → evidence_graphs(id, source_id)`, so a Subject cannot leave its graph's Source.
- `citations` carry `artifact_id` and `graph_id` (never changes); Observations hang off Subjects and Citations by id.
- Files are content-addressed by checksum, so no bytes move.

So a move is an update of `source_id` on three tables in one transaction, with no new tables:

1. `artifacts.source_id` for the moved Artifact.
2. `evidence_graphs.source_id` for each graph that moves with it.
3. `subjects.source_id` for every Subject in those graphs.

Citations, Observations, `evidence_graph_artifacts`, locators, depictions and identity claims do not change. The composite key has no `ON UPDATE CASCADE`, so steps 2 and 3 either run with `PRAGMA defer_foreign_keys = ON` inside the transaction, or a migration adds `ON UPDATE CASCADE` to that key ([`add-catalog-migration`](../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../.agents/skills/add-catalog-model/SKILL.md)). Decided (user, 2026-11-09): whichever makes the move work; prefer `PRAGMA defer_foreign_keys` inside the transaction, which needs no migration, unless planning finds a reason not to.

## Which graphs move

| Case | What happens |
| --- | --- |
| The Artifact is in no graph | Only the Artifact moves |
| Each of its graphs holds only this Artifact (the common case) | The Artifact and those graphs move |
| A graph also holds other Artifacts | They join the bundle, and so do their other graphs; the confirm lists everything |
| The bundle holds all the Source's graphs | It moves; the old Source gets a fresh empty Default graph so it keeps one (`ErrLastGraph`) |
| The bundle holds all the Source's Artifacts | It moves; the old Source is left empty and is not deleted (Merge Sources would offer to) |
| A moved graph's name matches one on the target (two "Default"s) | Allowed: graph names are not unique per Source. The researcher can rename it |

## Side effects to name in the confirm

- **Credibility.** Assessments are per Source ([`source-structure-later.md`](source-structure-later.md) "Per-graph credibility"), so the moved Observations are scored with the target Source's credibility from then on, and Conclusions that rest on them can re-rank. The confirm says so when the two Sources' assessments differ.
- **Empty Source.** If the old Source is left with no Artifacts, say so; do not delete it.

## Engine notes

- One write through the catalog session ([`use-catalog-session`](../../../.agents/skills/use-catalog-session/SKILL.md)) and one FFI handler, for example `MoveArtifactsToSource(artifact_ids, target_source_id)` ([`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md)), plus a preflight read that returns the bundle (Artifacts, graphs, Subject counts) and the credibility difference for the confirm; the same closure function drives both, so the confirm and the write cannot disagree.
- **Audit:** one transaction scoped to **both** Sources, so each Source's history and "Updated" sort show the move ([`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md)). Earlier transactions keep their old scope.
- **Refs:** unchanged; `ART-`, `GPH-`, `CIT-` and Subject refs are project-wide.
- **Search and caches:** re-project the moved rows' search documents if a hit shows its Source title ([`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md)); refresh both Sources' pages, the Sources list counts, and any Person page whose provenance rows name the old Source.
- **Navigation:** if the researcher was on the moved Artifact or graph, land them on it under its new Source ([`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md)).

## Open questions

- ~~Entry point~~: the Artifact and the Evidence graph, one flow (user, 2026-11-09).
- ~~Several Artifacts at once~~: the move is always the whole bundle, shown in the confirm (user, 2026-11-09).
- ~~Move to a new Source~~: deferred; combining is covered by moving into an existing Source (user, 2026-11-09).
- **Merge Sources.** Whether this spike ships it as a convenience over the move.
- ~~Undo~~: not in this spike (user, 2026-11-09).
- ~~Subjects' composite key~~: whatever works; prefer deferred foreign keys (user, 2026-11-09).
- ~~Source-level values~~: not looked at (user, 2026-11-09).
