# Spike 12 — Source structure

**Done** (closed 2026-10-29). A big Source (a census, a book, a directory) holds several isolated Evidence graphs, each with its own Artifacts and promoted on its own, and Copy Source starts a separately citable item inside a publication. The open audit fixes moved to Spike 14 when it closed. Live models: [`source-layer-data-model.md`](../../../source-layer-data-model.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`catalog-deletes.md`](../../../catalog-deletes.md), [`audit-revision-history.md`](../../../audit-revision-history.md). Design reasoning: [`source-structure.md`](source-structure.md).

## Decisions

- **Two tools, one rule for choosing.** A part that can be cited on its own (its own author, title or date) is its own Source, started with Copy Source; a region of one work is another Evidence graph.
- **Copy Source composes in `core/database/sourcecopy`**, from a draft of the original: title, type, description, metadata and layout, credibility, and the Artifacts the researcher checks. Graphs, Subjects, Citations and notes are not copied.
- **An Evidence graph is its own table** (`evidence_graphs`, ref prefix `GPH`), with `evidence_graph_artifacts` links. A Subject belongs to one graph forever; `subjects.source_id` stays for document-level code. Every Source has at least one graph, `Default`, and the last one can't be deleted.
- **Graphs are fully isolated.** An Observation's Subject, subject value and Citation Artifact must all be in, or linked to, one graph (`requireSameGraph`). No moving a Subject and no splitting a graph.
- **An Artifact is the parent and may stand in no graph.** A graph needs at least one Artifact; Add artifact and card edit list the graphs as checkboxes; an Artifact in no graph says so on its card. One unlink check and one link writer; an Artifact card's Save is one write.
- **A Citation belongs to one Evidence graph** (decided 2026-10-29; built in Spike 14 as S14-10).
- **Promote is per graph**, but agreement and credibility stay per Source: two graphs of one census agreeing count as one Source.
- **A graph is a place.** History, restore and breadcrumbs carry the graph; the omnibar finds a graph by name, description or `GPH-` ref; a destination declares the id its view resets on.
- **Status is per graph** (not started, in progress, promoted), and the Sources list shows each graph as a button with its counts.
- **A write names only the graphs whose loaded rows it moves**, and partial reads ride one keyed effect (`WriteEffects.reads`). Each cache key has one typed descriptor.
- **FakeStore is a seeded fake, not a second engine.** It returns seeded rows and scripted effects and errors.
- **Go decides what the app may open.** The `files` read returns a path only for an allowed type, and Swift keeps no second allowlist.

## Refused

Nested Sources and derivation links; per-graph credibility; moving Subjects between graphs; a graph name in the canvas header.

## Leftovers

| Item | Home |
| --- | --- |
| Open PRs at close: a Citation in one graph, reloading only what moved, file work before the catalog session, crafted-project crashes, copy and VoiceOver, synchronized folders, effects on every write, name folds, a stale graph-cache link after a merge, Promote recovery, search spellings, Promote's wire, oversized JPEGs | Spike 14, S14-10 – S14-23; the two audits moved with them |
| Graph alignment conflicts after a decision, Promote's performance at scale, names compared however written | [`promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Crafted name keys, alias tables and kinship fan-out | [`share-packages.md`](../../../ideas/share-packages.md) "Harden before sharing" |
| Search kind columns, canvas state enums | [`structural-debt.md`](../../../ideas/structural-debt.md) |
| The evidence sheet's New-row opener, the sheet's pins, the typed band | Removed with Identity Claim pins ([Spike 13 slice 0](../../spike-13/deployment-plan.md)) |
| Nested Sources, derivation links, templates, per-graph credibility, layered citations | [`source-structure-later.md`](../../../ideas/source-structure-later.md) |
| `source` Subjects, `mentions` and `remark` | [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md) |
| Ship-log follow-ups: composer load error placement, a silent Open failure, a card edit dropped without asking; a thumbnails test fixture; recount order | Spike 14 S14-14, S14-12, S14-16 |
| Ship-log follow-ups: a shared tiny-PNG test writer, the history menu on shared rows; a `SourceSurface` guard row; type-select on a disabled menu row | [`shared-list-helpers.md`](../../../ideas/shared-list-helpers.md), [`structural-debt.md`](../../../ideas/structural-debt.md), [`docs/dogfood/ux.md`](../../../dogfood/ux.md) |
| Dropped: a doubled `hasFile` check, one unreproduced flaky canvas test, logged nits | Not worth a PR (2026-10-29) |
