# Add an artifact to a source without selecting an evidence graph

- **Date:** 2026-10-26
- **Type:** Feature request / workflow
- **Where:** Source page / Add artifact dialog / Evidence graphs
- **Annoyance (verbatim):** "We need a way to be able to add an artifact to a source without selecting an evidence graph."
- **Consequence:** On a Source with two or more Evidence graphs, the Add artifact dialog shows an "Evidence graphs" checklist and refuses to save until at least one box is checked ("Check at least one evidence graph."). Filing a scan onto a Source becomes a graph decision the researcher may not be ready to make, and on a one-graph Source the question never appears, so the behaviour feels inconsistent.
- **Wanted:** Adding an Artifact to a Source is a Source-level act. The dialog should let the researcher attach the file without choosing graphs — either by making the checklist optional (unchecked = link to no graph yet, or to every graph as the engine already does by default) or by moving graph linking to a later step on the artifact card or the graph itself.

## How it works today (observed, not implemented)

- **Schema.** Artifacts belong to a Source, not a graph: `artifacts.source_id` (`000008.sql`, reshaped `000030.sql`). Graph membership is a separate many-to-many join `evidence_graph_artifacts (graph_id, artifact_id)` from Spike 12 (`000055.sql`). Nothing in the schema requires an Artifact to have a graph row; the join can be empty for an Artifact.
- **Go core.** `core/database/artifacts/artifacts.go` `CreateInput.GraphIDs` — "Empty = every graph of the Source"; `Create` calls `evidencegraphs.LinkArtifact(tx, …, in.GraphIDs)`, which, when given no ids, lists the Source's graphs and links all of them (`core/database/evidencegraphs/evidencegraphs.go` ~L648). So the engine already accepts a graph-less request and resolves it itself. There is no "link to zero graphs" path: an Artifact linked to no graph cannot be cited from any graph (`docs/deployment-plan/spike-12/deployment-plan.md` rule 3, "Citing an Artifact not linked to the graph is refused"), but it can exist.
- **FFI.** `CreateArtifactRequest.graph_ids` (`api/proto/engine.proto`, "empty = every graph of the Source"); `api/ffi/handlers/artifacts.go` parses and forwards it unchanged.
- **UI — where the requirement actually lives.** `macos/App/Features/Sources/SourceArtifactsSection.swift`:
  - `visibleGraphs` returns the Source's graphs only when there are **two or more**; with one graph the checklist is hidden and `draftGraphIDs` is empty, so the engine's every-graph default applies.
  - `draftNeedsGraph` is true when the checklist is visible and every box is unchecked; `canSubmitDraft` and `submitDraft` (~L534) both refuse while it is true.
  - `macos/App/Features/Sources/SourcePageArtifactsView.swift` renders the `SourceArtifactGraphLinks` checklist (`SourceArtifactGraphLinks.swift`, a `PVChecklistState`-backed group) and the `L10n.Sources.artifactGraphsNeedsOne` hint ("Check at least one evidence graph.").
  - The boxes start all checked (`draft.excludedGraphIDs` is the unchecked set), so the block only bites when the researcher deliberately unchecks everything, or when the Source's graph list has loaded two graphs and they want "none".
- **Constraint to weigh.** Spike 12's link rules (`deployment-plan.md` rule 8): a new graph needs an Artifact; unlinking a graph's last Artifact, or one its Citations use, is refused. Those rules are about graphs losing Artifacts, not about an Artifact having none, so "add with no graph" does not collide with them — but an Artifact linked to no graph would be invisible to every composer until linked, which is the UX trade-off to decide. The edit-card path (`SourceArtifactsSection.LinkEdit`, with `SourceArtifactGraphLinks` in edit mode) already supports changing links after the fact, so a later link step exists.
- **Related design docs.** `docs/source-layer-data-model.md` § Evidence graphs (every Source has at least one graph; artifacts ↔ graphs many-to-many); `docs/deployment-plan/spike-12/deployment-plan.md` S12-03 / S12-05 / S12-10 (link rules, `GraphIDs` empty = every graph).
