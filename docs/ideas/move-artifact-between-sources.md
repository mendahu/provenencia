# Move an Artifact to another Source

**Status:** idea, planned for Spike 15 with the Source schema pass ([`citation-formats.md`](citation-formats.md), [`source-provenance-date.md`](source-provenance-date.md)) (user, 2026-11-01). Once Sources carry the fields a citation needs, which Source an Artifact sits under matters more, and researchers will want to re-file Artifacts they already built Evidence graphs on.

## Problem

A researcher files an Artifact under one Source, builds an Evidence graph on it (Subjects, Observations, Citations, identity claims to the canon), and later wants their Sources organized differently: the obituary belongs under the newspaper issue it came from, two census images were filed as one Source and should be two, a scan was attached to the wrong certificate. Today the only way is to rebuild the graph by hand under the other Source and delete the old one, which loses the Citations, the regions drawn on the pages and the claims to canonical people.

## The feature

On an Artifact (its row on the Source page, or its menu), **Move to another Source…**: pick the target Source, confirm, and the Artifact moves together with the Evidence graph or graphs built on it. Nothing is copied: every row keeps its id and ref, so identity claims, Conclusions and the provenance sheet keep pointing at the same records. The Artifact then reads its citation fields (title, date, jurisdiction, repository) from the new Source.

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

Citations, Observations, `evidence_graph_artifacts`, locators, depictions and identity claims do not change. The composite key has no `ON UPDATE CASCADE`, so steps 2 and 3 either run with `PRAGMA defer_foreign_keys = ON` inside the transaction, or a migration adds `ON UPDATE CASCADE` to that key ([`add-catalog-migration`](../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../.agents/skills/add-catalog-model/SKILL.md)). Pick one when planning.

## Which graphs move

Because the link is many to many, "the Artifact and its graph" is not always a clean pair:

| Case | First answer |
| --- | --- |
| The Artifact is in no graph | Only the Artifact moves |
| Each of its graphs holds only this Artifact (the common case) | The Artifact and those graphs move |
| A graph also holds Artifacts that stay | Refuse, and name the other Artifacts; offer to move them together (the move takes a set of Artifacts, closed over their shared graphs) |
| The graphs moving are all the Source's graphs | The Artifact moves; the old Source gets a fresh empty Default graph so it keeps one (`ErrLastGraph`) |
| A moved graph's name matches one on the target (two "Default"s) | Allowed: graph names are not unique per Source. The researcher can rename it |

Splitting a shared graph (moving only the part built on one Artifact) is out of scope: Subjects can draw on Citations from several Artifacts, so there is no clean cut.

## Side effects to name in the confirm

- **Credibility.** Assessments are per Source ([`source-structure-later.md`](source-structure-later.md) "Per-graph credibility"), so the moved Observations are scored with the target Source's credibility from then on, and Conclusions that rest on them can re-rank. The confirm says so when the two Sources' assessments differ.
- **Citation line.** The provenance sheet's citation line, and later the formatter, read the new Source's title and fields.
- **Empty Source.** If the old Source is left with no Artifacts, say so; do not delete it.

## Engine notes

- One write through the catalog session ([`use-catalog-session`](../../.agents/skills/use-catalog-session/SKILL.md)) and one FFI handler, for example `MoveArtifactsToSource(artifact_ids, target_source_id)` ([`add-ffi-handler`](../../.agents/skills/add-ffi-handler/SKILL.md)), plus a read that previews the closure (graphs, Subjects, Artifacts that would have to come along) for the confirm.
- **Audit:** one transaction scoped to **both** Sources, so each Source's history and "Updated" sort show the move ([`add-audit-scope`](../../.agents/skills/add-audit-scope/SKILL.md)). Earlier transactions keep their old scope.
- **Refs:** unchanged; `ART-`, `GPH-`, `CIT-` and Subject refs are project-wide.
- **Search and caches:** re-project the moved rows' search documents if a hit shows its Source title ([`add-searchable-kind`](../../.agents/skills/add-searchable-kind/SKILL.md)); refresh both Sources' pages, the Sources list counts, and any Person page whose provenance rows name the old Source.
- **Navigation:** if the researcher was on the moved Artifact or graph, land them on it under its new Source ([`add-workspace-location`](../../.agents/skills/add-workspace-location/SKILL.md)).

## Open questions

- Whether the entry point is only the Artifact, or also a graph ("move this graph and its Artifacts").
- Whether moving several Artifacts at once is in the first cut, or only the refuse-and-name path for shared graphs.
- Whether "move to a **new** Source" (split a Source in two, using Copy Source to prefill it) belongs here or follows.
- Undo: rely on moving it back, or offer an undo right after the move.
