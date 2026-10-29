# Ship log — S12-25 Writes name only the graphs they move

**Plan:** [`S12-25-writes-name-graphs-they-move.md`](S12-25-writes-name-graphs-they-move.md) · **Branch:** `s12-25-writes-name-graphs-they-move` · **PR:** [#621](https://github.com/mendahu/provenencia/pull/621)
**State:** ready

## Round 0 — implement
Base `4145ccdd` (origin/main at branch cut; docs-only since plan base). Head `6a8f6573`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report deviations: `TestCitationReadsPreferArtifactGraphs` lives in `reads_resolve_test.go` (package `effects_test`) since `TestReadsDeclare` needs package `effects`; `SourcePageModelTests.aGraphWriteMovesOnlyItsRow` and `SourcesModelTests.listReadsEachGraphsProgress` now apply a `reads` progress entry; `apply(invalidations:)` dedupes keys before `applyHiddenPolicy`; a graph delete now names the deleted graph in the `citations` read; `TestLinkWritesMoveCitationReads` asserts the read; commits 2–3 do not build alone; manual check not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`, `core/database/**`, Workspace session cache). Skipped: —.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, CON-2 | CON | blocker | First / last Observation moves a Citation in or out of every Artifact-linked graph's citation reads; `observations` Reads name only `observationGraph` | To user → Later PR | verified: `citations.go:55` `sqlInGraph` includes Citations with no Observations in every linked graph; D-3 fixes observations Reads to `observationGraph`. User (2026-10-29): "do the absolute minimum change … leave it and then use the plan PR skill to plot out a schema change" — a Citation belongs to one Evidence graph ("If two graphs need the same citation, they can just make it again"). Known edge case on this PR; recorded in the Landed note (F-6); S12-40 (planned on this branch) stores the graph and removes the fallback |
| MNT-1 | MNT | major | `Reads` hard-codes Evidence-graph keys; S12-26's read is keyed by Source | Fix now → F-1 | verified: map S12-26 Ships "re-reads only the named Sources' graphs"; `reads_test.go` requires `evidence_graphs` end |
| MNT-2 | MNT | major | `fallbackToSectionRoot` copies `replaceCurrent(with:)` | Fix now → F-2 | verified: `WorkspaceNavigation.swift:159-179` same body |
| MNT-3 | MNT | minor | Two identical citation `refresh` closures | Fix now → F-3 | same-pass cleanup; sync path stays S13-01 (Out) |
| MNT-4 | MNT | minor | Progress Reads repeat EvidenceGraph path literals | Fix now → F-4 | verified in registry diff |
| PRF-1 | PRF | minor | Citation change walks Observations twice with no per-step dedupe | Fix now → F-5 | rides along: fix plan already in `registry.go` / `path.go` |
| UX-1 | UX | minor | Artifact switch in an edit composer keeps `citationId` beside a new `artifactId` | To user → Fix now → F-6 | user (2026-10-29): "Clear them now" — amends D-8 |

Nits (logged): `MNT-5 · WorkspaceSessionTests.swift · sleep-based "nothing reloaded" checks; use a call counter` · `MNT-6 · CatalogLastUpdated.swift:43 · name S12-32 on the Promote special case` · `UX-2 · CatalogQueryRegistry.swift:505 · debug log in the silent progress catch`

### Fix plan 1

**F-1 — Each read declares its key table** (from MNT-1)
Files: `core/database/effects/registry.go`, `reads_test.go`, `README.md`; `macos/App/Features/Workspace/Session/CatalogMutation.swift` (doc comments only).
Failing test first: `TestReadsDeclare` — check each `ReadPath`'s end against `r.Read.KeyTable()` instead of the literal `"evidence_graphs"`, and assert every entry of `Reads` has a non-empty `KeyTable()` (fails to compile until F-1 adds it).
Do: add `func (r Read) KeyTable() string` beside `Reads` — one switch, both current reads return `"evidence_graphs"`, unknown returns `""`. If `validate` checks read paths, check the end against `KeyTable()` there too. Rewrite the `Read`, `ReadPath`, `ReadIDs`, `Effect.Reads` comments, the README line, and the Swift `CatalogRead` / `CatalogWriteEffectsRead` comments to "ids of the read's key table (Evidence graphs for `citations` and `evidence_graph_progress`)". S12-26 then adds one constant and one `KeyTable` case.
Check: `rg -n "per Evidence graph|ends at .?evidence_graphs" core/database/effects macos/App/Features/Workspace/Session` prints no line about Reads; `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...`; scripts/check-pr.sh PASS.

**F-2 — `fallbackToSectionRoot` calls `replaceCurrent(with:)`** (from MNT-2)
Files: `macos/App/Features/Workspace/WorkspaceNavigation.swift`.
Do: body becomes `replaceCurrent(with: .sectionRoot(currentLocation.section))`. Keep its doc comment.
Check: existing `WorkspaceNavigationTests` pass; scripts/check-pr.sh PASS.

**F-3 — One citation refresh factory** (from MNT-3)
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` (and `CatalogQueryKey.swift` if the accessor goes there).
Do: add one helper, e.g. `CatalogQueryRefresh.applyHiddenPolicy(follows read: CatalogRead, graphId: @escaping @Sendable (CatalogQueryKey) -> String?)` that returns the `CatalogQueryRefresh` whose `run` guards `graphId(key)` ∈ `ids` then calls `session.applyHiddenPolicy(to: key)`. `.citationCounts` and `.citationsByArtifact` use it with their case pattern. Keep the `Task` hop (sync path is S13-01, plan Out). Ride-along nit MNT-6: the `.evidenceGraphProgress` Promote special case in `CatalogLastUpdated.swift` names S12-32 in its comment.
Check: `rg -n "session.applyHiddenPolicy\(to: key\)" macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` prints exactly one line; `WorkspaceSessionTests` citation refresh tests pass; scripts/check-pr.sh PASS.

**F-4 — Name the progress graph paths once** (from MNT-4)
Files: `core/database/effects/registry.go`.
Do: beside `observationGraph`, add named paths (e.g. `claimGraph = from("subject_id", field("graph_id"))`, `claimEvidenceGraph = up("identity_claim_id", "subject_id", "graph_id")`) and use each in both `EvidenceGraph` and the `ReadEvidenceGraphProgress` Read of its row. Same for `subjects` if its two expressions are literal copies.
Check: `rg -n 'from\("subject_id", field\("graph_id"\)\)|up\("identity_claim_id", "subject_id", "graph_id"\)' core/database/effects/registry.go` prints only the named definitions; effects tests pass.

**F-5 — Dedupe ids per path step** (from PRF-1)
Files: `core/database/effects/path.go`.
Failing test first: a `path_test.go` (or existing path test file) case: 3 Observations on 1 Subject → `observationGraph` issues one `subjects` lookup (count via a query-counting walk or by asserting `readRowsColumn` input dedupe on a helper). If counting queries needs new test plumbing, assert the deduped output and skip the count.
Do: in `readRowsColumn`, skip an input id already seen in this call (a `seen` set keyed by `string(id)`), so a step looks each distinct id up once. Output order follows first sighting.
Check: effects tests pass; scripts/check-pr.sh PASS.

**F-6 — Clear the dropped Citation from the composer location; record CON-1** (from UX-1, CON-1)
Files: `macos/App/Features/CitationComposer/CitationComposerModel.swift` (`followArtifactInLocation`), `CitationComposerModelTests.swift`, the plan's Landed note.
Failing test first: `CitationComposerModelTests.switchingArtifactsAwayFromAnEditedCitationClearsItsIds` — edit composer on Citation X (location has `citationId`, `observationId`), select another Artifact → `navigation.currentLocation` has the new `artifactId` and nil `citationId` / `observationId`; `title`/`ref`/`sourceTitle` kept.
Do: when the switch drops the active Citation (`activeCitationID` becomes nil), the same `replaceCurrent` copy also clears `citationId` and `observationId`. A switch that keeps the active Citation leaves them. Landed note: D-8 amended by user — an Artifact switch that drops the active Citation clears its ids; and CON-1 known gap — the first/last Observation of a Citation does not refresh other Artifact-linked graphs' citation reads; closed by the "Citation belongs to one Evidence graph" map PR(s).
Check: the test passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `9775a92e`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). F-1–F-6 done (F-1, F-4, F-5 share `989ad12a`). Deviations: `selectingAnArtifactRewritesTheLocationInPlace` now opens Citation `cit-1` first, so it pins the keep-ids case; `completed.md` notes the Artifact-switch clearing and the CON-1 gap. Minors read by the orchestrator: PRF-1 Resolved (`readRowsColumn` `seen` set; `TestReadRowsColumnLooksUpEachIDOnce`); UX-1 Resolved (`followArtifactInLocation` clears `citationId` / `observationId` when `activeCitationID == nil`); MNT-3 / MNT-4 go back with MNT.

## Round 2 — review
Reviewers: MNT only (had Fix-now majors MNT-1, MNT-2). Scope: `6a8f6573...9775a92e`. CON-1 went Later PR, so CON does not return.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `Reads` hard-codes Evidence-graph keys | Resolved | `Read.KeyTable()`; `validate` and `TestReadsDeclare` check the end against it |
| MNT-2 | MNT | major | `fallbackToSectionRoot` copies `replaceCurrent(with:)` | Resolved | `WorkspaceNavigation.swift:159-161` calls `replaceCurrent` |
| — | MNT | — | new blockers | — | none |

Nits (logged): `MNT-7 · CatalogLastUpdated.swift:30 · ".reads" comment still says "graph ids"; say "key ids"` · `MNT-8 · effects/registry.go validate · end check repeats TestReadsDeclare's`

## Result
Gate PASS (after merging main @ `0fe0b280`) · CI green (Go + macOS) · open: none · later PR: CON-1 → S12-40 ([plan](S12-40-citation-in-one-graph.md), planned on this branch, user 2026-10-29) · follow-ups: none · to user: CON-1 (answered: leave on this PR, plan the schema change), UX-1 (answered: clear the ids now) · CON miss: none
