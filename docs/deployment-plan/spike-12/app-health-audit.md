# App-health audit — 2026-10-27

The findings behind slices 5–7 of the [map](deployment-plan.md). An area audit of the whole app at `main` @ 731e423 with [`review-pr`](../../../.agents/skills/review-pr/SKILL.md) ("Area audit"), all five lenses, with Evidence graphs (slices 1–4) read closest. PR order and status are in the map; the decisions taken on these findings are its working rules under "Audit fixes".

**How to use this doc.** A scope entry names its items (`A-n`). `plan-pr` reads those items here for where, evidence and fix, and re-checks each against the code first: line numbers are from 731e423 and move. Go findings marked *probe* were proven by a failing test; Swift findings come from reading the code (no Xcode on the audit machine) and carry manual steps.

**Baseline on 731e423.** `go vet` clean; every Go package passes (92); `gofmt -l core api` clean; `check-localizable-xcstrings.py` OK; `check-plan-docs.py` 0 errors. About 42,000 fuzzed FFI handler calls, no panic.

**Verdict.** The Evidence graph engine held under adversarial probes: isolation (`requireSameGraph` on every Observation write), cross-Source refusals (`checkLink`, Subject Source from its graph), the `last_graph` and in-use delete gates, Promote per graph, the auto-reconciler counting Sources not graphs, and the format-55 migration. The defects are at the edges: the canvas's identity, the Artifact ↔ graph edge, how far a write's effects reach, link saving orchestrated in Swift, and docs and skills that drifted.

## Scheduled

| Item | Title | Priority | Lens | PR |
| --- | --- | --- | --- | --- |
| A-1 | Opening an Artifact from a shared project can launch a program | High | SEC | S12-23 |
| A-2 | Canvas to canvas keeps the first graph's model | High | CON | S12-21 |
| A-3 | An Artifact can strand a Citation in no graph | High | CON | S12-22, S12-38 |
| A-4 | Every Source page load scans the audit history per note | High | PRF | S12-24 |
| A-5 | Saved metadata URLs open with any scheme | High | SEC | S12-23 |
| A-6 | Skills and rules that teach the wrong pattern | High | MNT | S12-20 |
| A-7 | A Citation, Artifact or Source-title write reloads every graph of its Source | Medium | PRF | S12-25 |
| A-8 | An Artifact card's link save is several transactions run from Swift | Medium | MNT, PRF | S12-22, S12-38 |
| A-9 | Link and graph writes reload the whole project's graph list | Medium | PRF | S12-26 |
| A-10 | Hidden graph and Source-page keys reload instead of going stale | Medium | PRF | S12-26 |
| A-11 | The canvas has no failed-read or deleted-graph state | Medium | UX, CON | S12-21 |
| A-12 | Nothing says which Evidence graph you are on | Medium | UX, CON | S12-29 |
| A-13 | A Kept-linked Citation link discards a dirty graph form | Medium | UX | S12-30 |
| A-14 | Copy says "this Source" where the action is one graph | Medium | UX | S12-29 |
| A-15 | FakeStore is a second engine with no parity check | Medium | MNT, CON | S12-35 |
| A-16 | Source-keyed fields on graph-keyed paths | Medium | MNT | S12-25 |
| A-17 | Return confirms destructive deletes | Medium | UX | S12-30 |
| A-18 | VoiceOver is not told graph write results or failures | Medium | UX | S12-30 |
| A-19 | The composer swallows a failed Citation load | Medium | UX | S12-21 |
| A-20 | Counts without plural rules | Medium | UX | S12-29 |
| A-21 | Section headers are not VoiceOver headings | Medium | UX | S12-30 |
| A-22 | Cover and portrait resets are not audited | Medium | CON | S12-24 |
| A-23 | Domain docs that contradict the code | Medium | CON | S12-20 |
| A-24 | Ingest holds the catalog session across copy, hash and decode | Medium | PRF | S12-27 |
| A-25 | The Sources list filters and sorts every Source per body run | Medium | PRF | S12-26 |
| A-26 | Crafted graph positions crash the app | Medium | SEC | S12-28 |
| A-27 | The Artifact viewer decodes originals unbounded | Medium | SEC | S12-28 |
| A-28 | A Go panic kills the app | Medium | SEC | S12-28 |
| A-29 | Twenty-one audited writes drop their effects | Medium | MNT | S12-32, S12-33 |
| A-30 | A cache key is eight hand-kept lists | Medium | MNT | S12-34 |
| A-31 | Search kinds: a registry plus four per-kind switches | Medium | MNT | S12-36 |
| A-32 | Canvas tool and sheet state is a flag pile | Medium | MNT | S12-37 |
| A-33 | The DesignSystem README misses six components | Medium | MNT | S12-20 |
| A-34 | `project.pbxproj` is maintained by hand | Medium | MNT | S12-31 |
| A-35 | Dead FFI surface | Medium | MNT | S12-32, S12-33 |

### A-1 — Opening an Artifact from a shared project can launch a program

**Where.** `core/database/files/files.go` `StorageRelPath` (stored `media_type` used as is); `core/objectpath/objectpath.go` `Rel` (an unknown type gives no extension); `macos/App/Platform/ProjectFiles.swift` `isSafeObject` and `openObject` (`NSWorkspace.shared.open`); callers `SourceArtifactsSection.swift` (preview tile primary click) and `EvidenceGraphModel.swift`.
**Problem.** `files.media_type` is checked against the ingest allowlist on write only. A catalog someone else wrote can hold any type; the object path then has no extension, and `isSafeObject` checks regular file and Finder alias but not the execute bit. Launch Services runs an extensionless executable in Terminal. Gatekeeper helps only while the folder carries quarantine (not after USB, git or most sync clients). The app has no sandbox.
**Evidence.** *Probe:* `UPDATE files SET media_type='application/x-sh'`, then `GetSourceWorkspace` returns `rel_path` `objects/9e/e8/9ee8…f297` with no extension, which `openObject` passes to `NSWorkspace.open`.
**Fix.** Go: the `files` read keeps only types `mediatypes.Allowed` accepts. `objectstore.openRegular` and Swift `isSafeObject` refuse a file with any execute bit. `openObject` refuses a URL with no extension or one off the allowlist, and types conforming to `.executable`, `.application` or `.script`; a refusal shows an error naming the file (working rule "What the app hands to macOS to open is checked first").

### A-2 — Canvas to canvas keeps the first graph's model

**Where.** `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.sourceGraph` arm: `EvidenceGraphView` has no `.id`, while the composer and Promote arms have `.id(entry.identityKey)`. `EvidenceGraphView.init` builds `@State model` from `graphID`; `EvidenceGraphModel.graphID` and `sourceID` are `let`; `.task(id: graphID)` re-runs `prepare()` on the old model.
**Problem.** SwiftUI keeps `@State` across re-renders of the same arm, so going from graph A's canvas straight to graph B's shows B's rows (the view's own `graphKey` follows B) while the model acts on A: Place creates the Subject in A (`createSubject(graphID:)`), Promote all opens A's Promote, Connect hands B's Subjects to A's composer, `canCite` and the Files pane read A, and across Sources the header shows A's title. A Subject never moves between graphs, so a misplaced one must be deleted. Latent since Spike 6; S12-15's graph hits make canvas → canvas a main path.
**Evidence.** Manual: census Source with graphs Jones and Smith, both linked to a page. Open Jones from the Sources list; omnibar "Smith", Return on the graph hit; arm Person, place "Test". No card appears; Jones holds "Test". Back / Forward between two canvases repeats it.
**Fix.** `.id(graphID)` on `EvidenceGraphView` in the host. A hosting test in `WorkspaceDestinationHostTests` (graph A of Source 1, then graph B of Source 2; the header reads Source 2). A rule in `add-workspace-place` (working rule "A view built from a location id resets on that id").

### A-3 — An Artifact can strand a Citation in no graph

**Where.** `core/database/evidencegraphs/evidencegraphs.go` `Unlink` (checks only the graph's own in-use Citations and its last link) and `Delete`; `core/database/deleteimpact/projectors.go` `projectCitation` (a Citation whose Artifact is in no graph gets `GraphID ""`); `SourceArtifactsSection.swift` `saveFields` (card edit accepts zero graphs).
**Problem.** A Citation with no Observation belongs to no graph and shows in any composer whose graph links its Artifact. When the Artifact leaves its last graph, no composer offers it: the Citation is unreachable, and the Artifact cannot be deleted while it exists (the delete preview's link opens a composer with no graph). Two paths: uncheck every graph on the card (two-graph Source), or delete a Default graph after making Jones from one page (the other page is left in no graph, and with one graph left the card's boxes are hidden).
**Evidence.** *Probe* (`evidencegraphs_test`): a Source with P1 and P2 auto-linked to Default; Create Jones with P2; Delete Default succeeds and Jones, the only graph, links 1 of 2 Artifacts. *Probe:* a Citation with no Observation on X, X in Default and Jones; unlink from both succeeds, `deleteimpact.Impact(KindArtifact, X)` reports gate `inbound` and its Citation opens `citationComposer` with no graph.
**Fix.** Decided with the user 2026-10-27: an Artifact may be in no graph (working rule "An Artifact is the parent"), so only the Citation is the defect. Working rule "An Artifact with Citations keeps at least one graph": `Unlink` refuses that Artifact's last graph link with its own user-kind code; graph `Delete` is refused through a `deleteimpact` gate when it would leave such an Artifact in no graph; both name the Citations. Swift shows the refusal as the card's locked box (S12-38).

### A-4 — Every Source page load scans the audit history per note

**Where.** `core/database/sources/notes.go` `sqlListNotes` and `sqlGetNoteAttributed`; called by `GetSourceWorkspace`. The only index on `audit_changes` is `audit_changes_transaction_id_idx` (`migrations/000031.sql`).
**Problem.** Each note's author and date are two correlated subqueries on `audit_changes WHERE entity_type = 'source_note' AND entity_id = n.id`: a full scan each, so notes × all history, under the held session, on every Source page load and every save that reloads the workspace.
**Evidence.** *Probe:* plan `SCAN c`; `sources.ListNotes` with 5 notes 8.3 ms at 20k `audit_changes` rows, 140.7 ms at 200k; with an index on `audit_changes(entity_id)` 0.13 ms and 0.09 ms.
**Fix.** Migration `000057` adds `audit_changes_entity_id_idx` ([`add-catalog-migration`](../../../.agents/skills/add-catalog-migration/SKILL.md)); a `dbtest` query-plan test fails on `SCAN c`. Optionally fold the two subqueries into one.

### A-5 — Saved metadata URLs open with any scheme

**Where.** `macos/App/Features/Sources/SourcePageMetadataView.swift` `openSavedURL`; `core/database/sourcemetadata` validates with `urlshape.Canonical` on write only; `core/urlshape/urlshape.go` `OpenHref` has no caller.
**Problem.** Swift prefixes `https://` only when the text has no `://`, then opens any scheme: `file:`, `smb:`, `vnc:`, another app's scheme.
**Evidence.** *Probe:* `value_text='file:///System/Applications/Calculator.app'` on a url field is returned unchanged by `GetSourceWorkspace`.
**Fix.** `MetadataWorkspaceEntry` gains `open_href` from `urlshape.OpenHref` (empty when invalid); Swift opens only that and `openSavedURL` goes.

### A-6 — Skills and rules that teach the wrong pattern

**Where and problem.**
- `.agents/skills/add-catalog-model/SKILL.md` step 6 lists `Effect.Source`; the field is `EvidenceGraph` (`core/database/effects/registry.go`) since S12-08.
- `.cursor/rules/swift-tests.mdc` names `eventTitlesBySource` (now `FakeStore.eventTitlesByGraph`) and says FakeStore does not re-implement ordering or graph walks; it does (`graphListOrder`, `citationIsInGraph`, `graphEffects`). The name is fixed here; the rule's rewrite is S12-35's.
- `.agents/skills/add-searchable-kind/SKILL.md` names `locationFor`, `Reproject` and `RebuildAll` but not `core/search/fts.go` `fieldValuesForKind`; a kind missing there falls back to title/ref/body, and a declared field such as `description` scores zero with no error.
- `.cursor/rules/catalog-refs.mdc` reserved prefixes omit `GPH`.
- `.cursor/rules/design-system-layers.mdc` uses `PVEvidenceIcon` as its example; no such component exists.
**Fix.** Correct each. Docs only.

### A-7 — A Citation, Artifact or Source-title write reloads every graph of its Source

**Where.** `core/database/effects/registry.go` `"artifacts"` and `"citations"` (`EvidenceGraph: graphsOfSource(…)`), `"sources"` on title; Swift `WorkspaceSession.apply` loops `effects.evidenceGraphs` and `refreshEvidenceGraphProgress` makes one progress call per graph.
**Problem.** On a Source split into G graphs, every Citation save (the composer's main write), Artifact create, ingest or label edit names all G: Go drops every loaded graph and its Promote memos; Swift makes G progress calls and reloads every graph opened this session. Progress cannot change from a Citation or Artifact field write; the graph loader reads Citations only through Observations (depiction locators in `graphfill/load.go`) and no Artifact column.
**Evidence.** *Probe* on a 60-graph Source: a new Citation with one Observation on graph 3 reports `evidence_graphs=60`; `create_artifact` and an Artifact label update likewise. Go cost per write: `graphprogress.Get` ×60 4.1 ms, reloading 60 small graphs 26.9 ms, before FFI.
**Fix.** Working rule "A write names only the graphs whose loaded rows it moves": `citations.EvidenceGraph` becomes `chain(inbound("observations","citation_id"), …graph)`; `artifacts` names none; `sources` title keeps its graphs but the app refetches progress only for writes that change counts. An effects test pins "a Citation names only its Observations' graphs".

### A-8 — An Artifact card's link save is several transactions run from Swift

**Where.** `SourceArtifactsSection.swift` `saveFields` and `loadLinkLocks`; `GenealogyStore.swift` `isOnlyArtifact`.
**Problem.** One card Save sends `updateArtifact`, then one `linkEvidenceGraphArtifact` per added graph, then one `unlinkEvidenceGraphArtifact` per removed graph, each its own `writes.Run`. The links-before-unlinks order, partial-save bookkeeping (`attemptedLink`, `wroteAnything`, `refusedGraphID`) and three error strings live in Swift, and `isOnlyArtifact` copies Go's last-link rule. The lock preflight is one `GetEvidenceGraphUnlinkImpact` per graph, in sequence (60 on a 60-graph census). A refusal or a quit mid-save leaves links half applied. `artifacts.Create` and `evidencegraphs.Update` already link in one write; Artifact update is the odd one out.
**Fix.** Working rule "An Artifact card's Save is one write": `UpdateArtifactRequest` carries the links; one preflight per Artifact returns every reason (`last_artifact`, `artifact_in_use` with Citations, A-3's new code). Delete `isOnlyArtifact`, the per-graph loop and the partial-save strings.

### A-9 — Link and graph writes reload the whole project's graph list

**Where.** `CatalogQueryRegistry.swift` (`.evidenceGraphs` list → `.evidenceGraphsList`, a visible key on every Sources place); `evidencegraphs.List(c, nil)`; `.createdSource` → `.sourcesList`.
**Problem.** Artifact create, Source create and Copy, graph create, edit and delete, link and unlink all report the `evidence_graphs` list, so each reloads every graph with its links (two scans and temp B-tree sorts under the session) and rebuilds `EvidenceGraphIndex` on the main actor; Add Source also reloads the Sources list. The write already reports which Source moved, and `List(c, sourceID)` seeks `evidence_graphs_source_id_idx`.
**Evidence.** *Probe:* project-wide `List` 20.9 ms at 3,060 graphs; `sources.List` warm 14 ms at 3,061 Sources.
**Fix.** `ListEvidenceGraphsRequest` gains a Source filter; a write re-reads only its Source's graphs and splices them into the cached list through the cache layer. Add and Copy Source insert the row `CreateSource` returns. Known: S12-05 chose one cached project list; S12-13 D-8 declined a patch.

### A-10 — Hidden graph and Source-page keys reload instead of going stale

**Where.** `CatalogQueryRegistry.swift` specs for `.evidenceGraph` and `.sourceWorkspace` have no `whenHidden`, so `.revalidate`.
**Problem.** Every off-screen graph or Source page visited this session reloads when a write names it, multiplying A-4 and A-7. The Sources, People, Events and Places lists already use `.markStale`.
**Fix.** `whenHidden: .markStale` on both; check the canvas, Source page and composer read through `readyValue` or `query` so a stale key reloads on the next visit. Known in part: S12-08 PRF-2.

### A-11 — The canvas has no failed-read or deleted-graph state

**Where.** `EvidenceGraphView.swift` (`graphHandle.status == .error` is never read; the empty overlay shows only on `.ready`); `EvidenceGraphModel.prepare` and `canCite`.
**Problem.** A failed graph read draws an empty grid under "Evidence graph (0 subjects)" with no message. Back or a relaunch onto a deleted graph shows "No artifacts in this evidence graph…", with Place live; Go then refuses every write with a generic `subjects.invalid`. `SourcePageView` has `fallbackToSectionRoot`; the canvas has none.
**Fix.** An error callout with Read again that re-queries the key. When the loaded graph list lacks the id: the graph no longer exists, with Go to Source page. Reuses the canvas's existing callout; no new chrome. Known: S12-11 Out row "Pruning history entries of a deleted graph".

### A-12 — Nothing says which Evidence graph you are on

**Where.** `EvidenceGraphView.swift` (VoiceOver container label "Evidence graph"; header reads the Source title); `WorkspaceToolbar.swift` `HistoryJumpRow` (reads `location.title` and `location.ref`, the Source's); `SourcesListNavigation.canvasLocation`.
**Problem.** Only the toolbar breadcrumb names the graph, and it truncates first. The Source page, the Jones canvas and the Smith canvas read the same in the history menu.
**Fix.** The canvas's VoiceOver label names the graph ("Evidence graph, %@"); graph, composer and Promote history rows read the graph name from `session.evidenceGraph(id:)` as `BreadcrumbNames` does, falling back to "Evidence graph". No header change (working rule "No graph name in the canvas header").

### A-13 — A Kept-linked Citation link discards a dirty graph form

**Where.** `SourceGraphFormDialog.swift` (`dismiss: { model = nil }` handed to the Kept-linked `PVDeleteImpactGroupList`); `PVDeleteImpact.swift` `activate`.
**Problem.** `dismiss` sets the binding directly, skipping `PVSheetChrome.dismiss()` and its `isDirty` check, so a typed name or description is lost; the rows stay live while saving.
**Fix.** When the form is dirty, ask the kit's "Discard changes?" first (`activate(…, then:)` exists); disable the rows while saving. Overrides S12-D3 GF-8 (working rule "A dirty form asks before a link inside it navigates away").

### A-14 — Copy says "this Source" where the action is one graph

**Where.** `Platform/L10n.swift`: `promote.page.mapRestHint` ("…on this Source"), `promote.error.graphUnavailable` and `…Repropose` ("this Source's Evidence graph"), `evidenceGraph.subject.addPropertyUnavailable` ("this Source has no Artifacts", spoken as a card action when the graph links none), `citationComposer.pairPickBoth`.
**Problem.** Map the rest adds one graph's Subjects (`PromoteFlow.swift`); VoiceOver says the Source has no Artifacts when it does.
**Fix.** Reword each to name the Evidence graph; `addPropertyUnavailable` matches the canvas callout ("…this evidence graph has no artifacts").

### A-15 — FakeStore is a second engine with no parity check

**Where.** `macos/App/Platform/FakeStore.swift`: `graphEffects` (create, update, delete, link, unlink report only `citationSources` and `lists`); `insertSource` (no `lists: [.evidenceGraphs]`, an extra `citationSources`); `graphListOrder` (`localizedCaseInsensitiveCompare` against Go's `name COLLATE NOCASE, ref, id`); `deleteEvidenceGraph` and `getDeleteImpact("evidence_graph")` never refuse; `copySource` accepts a blank title and unknown or repeated Artifact ids that Go refuses.
**Problem.** Swift tests pass on behavior the app does not have; nothing pins the fake to Go.
**Fix.** Route graph writes through `bumpSource` as the engine does; refuse the last graph and a graph with Subjects; match Go's order; refuse `copySource` input with `sources.invalid`. Longer term, a Go test writes the expected effects per action as a golden file a Swift test reads, beside `swift_vocabulary_test.go`. Rewrite `.cursor/rules/swift-tests.mdc` to say which engine rules the fake mirrors and how parity is held.

### A-16 — Source-keyed fields on graph-keyed paths

**Where.** `effects.Set.CitationSources` / `Effect.CitationSource` and `WriteEffects.citation_sources` (`engine.proto`); `CatalogQueryRegistry.swift` (`.allCached(.citationCounts)`, `.allCached(.citationsByArtifact)`); `CitationComposerContext.swift` (calls `session.query` after `apply`, citing an orphaned "D-28"); `EvidenceGraphProgress.source_id` (filled by an extra `SourceOf` query in `graphprogress.Get`, read by no Swift code); `graphprogress.ErrInvalid` uses `sources.invalid`.
**Problem.** Against the spike rule "replaces Source-keyed fields … instead of adding graph ones beside them". Every citation-moving write drops every cached citation read in the project.
**Fix.** Key citation effects by graph on the wire and invalidate `.citationCounts(graphId:)` per graph (or make it a list key if a flag is all the app needs); make the composer's Artifact key a visible key through `PlaceRegistry`; drop `EvidenceGraphProgress.source_id` (reserve it); give `graphprogress` an `evidencegraphs` code. Known: S12-10 D-6 kept `citation_sources`.

### A-17 — Return confirms destructive deletes

**Where.** `DesignSystem/Components/Confirm/PVConfirm.swift` (confirm button `.keyboardShortcut(.defaultAction)`; `.onAppear { cancelFocused = true }` with the comment "Return must not complete a destructive action"); `DesignSystem/README.md`.
**Problem.** Return activates the default button whatever has focus, so Return deletes the graph, Artifact or Source.
**Fix.** For the danger tone, Cancel is the default action and the destructive button has no Return shortcut (working rule "A danger confirm's default action is Cancel"). Update the README line.

### A-18 — VoiceOver is not told graph write results or failures

**Where.** `PVToast.swift`, `VocabularyChrome.swift` (`VocabularyToastOverlay`); `SourceGraphFormDialog.swift` and `SourcePageArtifactsView.swift` error callouts; `Platform/AccessibilityAnnouncement.swift` (three callers, none here).
**Problem.** "Evidence graph added" and "Evidence graph deleted" toasts vanish after 5 s unannounced; "Couldn't save the evidence graph" and a refused card save appear with focus still on the button.
**Fix.** Post `AccessibilityAnnouncement.post` once in the toast overlay when the toast changes (every toast gains it); announce a form's error callout when it appears.

### A-19 — The composer swallows a failed Citation load

**Where.** `CitationComposerModel.swift` `selectArtifactAndLoad` (`try? await loadCitation(…)` after resetting the fields) and `refreshCitationLines` (`try?`).
**Problem.** A failed load shows a blank new Citation on that Artifact; the researcher may type a duplicate. The sibling `loadCitationPreservingAuthoring` sets `fields.error`.
**Fix.** The sibling's do/catch with `L10n.Errors.message(for:)`; the refresh keeps its lines and shows the error the same way.

### A-20 — Counts without plural rules

**Where.** `L10n.swift` and `Localizable.xcstrings`: `promote.toast.filedTitle` and `.filedBody`, `promote.page.promoting`, `.rowSummary`, `.mapRest`, `citationComposer.artifactCitationCount`, `.citationStepperCount`, `.pageCount`; `DeleteImpact.noun` picks singular or plural in Swift (`count == 1 ? …`).
**Problem.** "Filed 1 subjects and 1 connections", "1 citations", "1 pages". `L10nPluralTests` only checks keys that already have variations; `check-localizable-xcstrings.py` has no such check.
**Fix.** `one` / `other` variations (substitutions for two counts); plural keys replace the Swift noun pick; the script fails on `%lld <word>` with no variations, with an allow-list ("Page %lld").

### A-21 — Section headers are not VoiceOver headings

**Where.** `DesignSystem/Components/SectionHeader/PVSectionHeader.swift` (no `.isHeader`); precedent `ConclusionDetailPage.swift`.
**Fix.** `.accessibilityAddTraits(.isHeader)` on the header title, in the kit.

### A-22 — Cover and portrait resets are not audited

**Where.** `catalogmodel/model.go` (`sources.primary_artifact_id` and `canonical_entities.preferred_likeness_observation_id`, `SET NULL`, `BucketOptional`); `migrations/000015.sql` trigger `sources_clear_cover_on_primary_null`; `artifacts.Delete`, `observations.Delete`.
**Problem.** Deleting the cover Artifact resets the Source's `cover_mode` and `primary_artifact_id` in SQLite; deleting the portrait's Observation clears the pointer. The revision records only the delete, so history cannot say when the cover or portrait went, and replay rebuilds a pointer to a deleted row. `SetCover` and `SetPreferredLikeness` audit the same columns.
**Evidence.** *Probe* (`sources`): set cover, delete the Artifact; the `delete_artifact` revision has no `source` change. *Probe* (`likenesses`): set preferred, delete the Observation; no `canonical_entity` change.
**Fix.** `deleteimpact` releases `BucketOptional` FKs explicitly before the parent delete, with an update change on the pointing row; an honesty test (beside `TestFacetReleaseHonesty`) that every `BucketOptional` FK has a release.

### A-23 — Domain docs that contradict the code

**Where and problem.**
- `docs/conclusion-reconciliation.md` §6 says search and matching read outvoted values; matching reads the profile (displayed positive, denied negative: `subjectprofile.go`, `promotealign/index.go`), as `matching-model.md` says. S11-11 D-4 updated one doc, not the other.
- `docs/matching.md` lists Segment compare (S11-14), the shortlist (S11-15) and the layer 6 walk (S11-16) as not built; they landed in #493, #497, #501. The layer table lacks `core/segmentcompare`, `core/shortlist`, `core/kinship`.
- This map's working rule "Sources list" and `completed.md` S12-14 described four columns with a Status column; the code has three, the badge inside the graph button (`SourcesListRow.swift`). Corrected with the map, 2026-10-27 (map review K-20).
- `docs/catalog-deletes.md` gate enum omits `last_graph`; `docs/macos-client-patterns.md` still says "Evidence graph for one source"; `core/database/effects/README.md` does not say a Source title edit names its graphs and reprojects their search documents; `api/proto/README.md` "There are no Source tables"; `docs/source-layer-data-model.md` §4.2 does not say a copy gets its own Default graph linking the copied Artifacts; `docs/promote-graph-alignment.md` "the Source's Evidence graph memo".
**Fix.** Correct each. Docs only.

### A-24 — Ingest holds the catalog session across copy, hash and decode

**Where.** `api/ffi/handlers/artifacts.go` ingest inside `withProjectCatalog`: `ingest.File` streams and hashes up to `MaxBytes` (512 MiB), then `artifactProto` → `derivatives.EnsureThumbnail` decodes; `catalogsession.Do` holds `s.mu` for all of it.
**Problem.** Every other catalog call (Source page, graph loads, counts, search) waits until a large scan is copied, hashed and thumbnailed.
**Fix.** Split `ingest.File`: open, stream to an `objectstore` temp file, hash and sniff outside `Do`; inside `Do` only `LookupByChecksum`, `Install` and `writes.Run`. Thumbnail bytes outside the lock, the derivative row inside. Keep the purge/import exclusion `source-layer-data-model.md` relies on.

### A-25 — The Sources list filters and sorts every Source per body run

**Where.** `SourcesListView.swift` (`let rows = model.visibleSources`; the Add dialog's `isDirty` and `confirmDisabled` read `model.draft` in the same body); `SourcesModel.visibleSources` (A–Z with `localizedCaseInsensitiveCompare`; `statuses(for:)` per Source under a status filter).
**Problem.** Each keystroke in Add Source or the list search re-runs a filter and full sort of N Sources, `rows.map(\.id)` and the `ForEach` diff on the main actor.
**Fix.** Memoize `visibleSources` against the Sources handle revision, settings and, under a status filter, the graph list and progress revisions (`searchKeyCache` is the precedent); optionally move the Add form into its own view.

### A-26 — Crafted graph positions crash the app

**Where.** `core/database/subjectpositions` (no bound on write or read; any int64 over FFI); `Features/Catalog/Connect/EvidenceGraphPlacement.swift` (`max()! + horizontalStep`, `Int64(floor(raw))`); `GraphCanvasGridMapping.swift`.
**Evidence.** *Probe:* `(9223372036854775807, -9223372036854775808)` is returned unchanged by `ListSubjectPositions`; Swift `+` and `Int64(_:)` trap on it.
**Fix.** One cell range (for example ±1,000,000): `SetSubjectPosition` refuses outside it and the list read clamps; Swift placement uses checked or saturating arithmetic.

### A-27 — The Artifact viewer decodes originals unbounded

**Where.** `Features/ArtifactViewer/ArtifactViewerModel.swift` `classify` (trusts the catalog's `media_type`; `NSImage(contentsOf:)`); `ArtifactRegionCrop.swift`.
**Problem.** Thumbnails are bounded (`ThumbnailDecodePolicy`); the viewer decodes the full original through every ImageIO decoder, so a small file declaring a huge image exhausts memory.
**Fix.** Check the declared pixel size and the sniffed type first with a viewer budget and the ingest allowlist; display through `CGImageSourceCreateThumbnailAtIndex` at screen size; crop from a bounded decode.

### A-28 — A Go panic kills the app

**Where.** `api/ffi/dispatch.go` `Call` and `api/libprovenencia/main.go` (no `recover`); `core/catalogsession/session.go` `Do` (`s.mu.Unlock()` not deferred).
**Problem.** The core is a c-shared library in the app process. The fuzz found no panic, but a decoder bug on a crafted file would crash, possibly on every launch.
**Fix.** One `recover` in `ffi.Call` returning `internal.unknown`; `defer s.mu.Unlock()` in `Do`, so a recovered panic cannot leave the session locked.

### A-29 — Twenty-one audited writes drop their effects

**Where.** Handlers that call `writes.Run` and return no `WriteEffects`: `PromoteSubject`, `ApplyPromoteGraphAlignment`, `UpdateSource`, `DeleteSource`, `SetSourceCover`, and the Source type, metadata field, Property, binding and term writes. Swift: `CatalogMutation.swift` (20 cases, `CatalogMutationKind`, `invalidationKind`).
**Problem.** Go's `effects.Resolve` computes graph, handle and search effects for these writes and the handler discards them; Swift restates some by hand (`.promotedSubject(graphId:)` after Promote). Every new key is wired twice. `UpdateSource` returning nothing also leaves Promote's cached Source titles stale.
**Fix.** Return `writeEffectsProto(wrote)` and move each caller to `apply(effects[, mutation])`; keep mutations only for facts effects cannot express (`.createdSource`); delete each case as it dies. Promote and Source writes in S12-32, vocabulary writes in S12-33.

### A-30 — A cache key is eight hand-kept lists

**Where.** `CatalogQueryKey.swift` (cases, `Kind`, `kind`, `project`); `CatalogQueryRegistry.swift` (`specs`, `load(key:store:) -> Any`, `Kind.invalidation`); `WorkspaceSession.swift` `warmQuery`; and for mutations `CatalogMutation`, `CatalogMutationKind`, `invalidationKind`.
**Problem.** A missing spec silently falls back to `.sessionFresh` with no invalidation; `load` returns `Any` and `warmQuery` restates the type, caught only at runtime; `Kind` is not `CaseIterable`. S12-05 and S12-14 each paid it.
**Fix.** One descriptor per kind (typed value, loader, spec, invalidation), looked up from `Kind`; at minimum `CaseIterable` and a test that every kind has one spec and `warmQuery` and `load` agree.

### A-31 — Search kinds: a registry plus four per-kind switches

**Where.** `core/search/registry.go` `KindSpec.Fields`; `core/search/fts.go` `fieldValuesForKind`, `locationFor`; `core/database/searchindex/searchindex.go` `Reproject`, `RebuildAll`.
**Problem.** Field names and weights are declared, the name → column mapping and the location builder are separate switches; a kind missing from `fieldValuesForKind` silently scores declared fields zero.
**Fix.** Each `KindSpec` field names its column, and the spec carries its location builder; the switches read the registry. Update `add-searchable-kind` (A-6 adds the interim line).

### A-32 — Canvas tool and sheet state is a flag pile

**Where.** `EvidenceGraphModel.swift` (`armedKind`, `armedConnect`, `armedSelect`, `shiftSelectHeld`, `isCreating`, `editingSubjectID`, `editingPrimaryKind`, `editingBridgeKind`); `EvidenceCanvasInputMode.swift`.
**Problem.** "Mutually exclusive" by 22 hand-written assignments; `inputMode` resolves impossible combinations by precedence.
**Fix.** `var tool: EvidenceCanvasInputMode` (shift stays a modifier) and `var sheet: … ?` (create or edit with its kind); the old reads derive from them. Precedent `CitationComposerModel.Phase`.

### A-33 — The DesignSystem README misses six components

**Where.** `macos/App/DesignSystem/README.md` has no row for `Components/Chip`, `Copy`, `FloatingMenuSelection`, `HoverEffect`, `InlineEdit` and `Recipes/LastUpdated`.
**Fix.** Add the rows; `add-ui-component` gains a "README row" checklist line; optionally `check-pr.sh` fails on a component folder with no row.

### A-34 — `project.pbxproj` is maintained by hand

**Where.** `macos/Provenencia.xcodeproj/project.pbxproj` (826 Swift build entries, no `PBXFileSystemSynchronizedRootGroup`, 95 commits since 2026-10-01); `add-swift-test` ("copy an existing test file's four objects and give them unused `AA…xx` IDs"), `add-ui-component`, `.cursor/rules/swift-tests.mdc`.
**Problem.** Every new Swift file is four hand-written objects edited by agents with no Xcode to check them; a conflict hotspot across parallel spikes.
**Fix.** Convert the App and test groups to synchronized folders on a Mac; delete the pbxproj steps from the skills and rule.

### A-35 — Dead FFI surface

**Where.** RPCs with no Swift caller and no idea file: `EnsureFileThumbnail`, `GetSubjectTypePresentation`, `ListPlaceableSubjectTypes`, `ListProperties`, `ListSubjectTypeProperties`, `RemoveInstallIdentity`. Store methods only tests call: `promoteSubject` (with its handler and the `GenealogyStore` overload), `listSubjectTypes`, `clearSubjectPosition`. `CatalogMutation.swift` comment "The S9-04 card dispatches it after `promoteSubject`". Kept: `UpdatePropertyTerm`, `DeletePropertyTerm` ([`term-management.md`](../../ideas/term-management.md)), `ReorderSubjectTypeProperties` ([`binding-order-editing.md`](../../ideas/binding-order-editing.md)).
**Fix.** Delete them with their handlers, tests and FakeStore code; reserve the proto method numbers. Confirm `RemoveInstallIdentity` is not planned first. The Promote path goes in S12-32, the rest in S12-33.

## Not scheduled

Low items from the audit, not pulled into the spike (user, 2026-10-27). Each names the scheduled PR that touches the same code, should one be pulled in at planning.

- **L-1.** Graph names with a leading accent sort after Z: SQLite `NOCASE` folds ASCII only (`evidencegraphs.sqlListOrder`, `deleteimpact` `sqlFirstArtifactGraph`). Sort in Go with the root collator `conclusionheaders` uses. Near S12-22.
- **L-2.** No test pins Go `apperr` codes to `L10n.Errors.message(code:)`; `claimconfidence.invalid` and `deleteimpact.release_incomplete` fall to "Unknown error". Near S12-35.
- **L-3.** `VERSION` has read 0.2.0 since catalog format 27 (now 56); plans cite `versioning.md` for a "until a release is cut" rule it does not contain.
- **L-4.** Catalogs made before #499 keep the 31 removed nickname pairs (`namealiases.Install` runs at create only); S11-44 only adds. Known: S11-28 review.
- **L-5.** A Sources list row draws one 52 pt button per graph: a 60-household census row is about 3,600 pt (board S12-D5).
- **L-6.** `PDFView` opens link annotations with any scheme (`ArtifactPDFViewport.swift`, no delegate). Near S12-23.
- **L-7.** Project open migrates a foreign catalog before `verifySchema`, so a refused catalog is left upgraded or half-migrated; open with `trusted_schema=OFF` and defensive mode.
- **L-8.** Derived caches in a foreign catalog (search documents, `display_parent_id`, thumbnail paths, reconciled values) are trusted when the version matches.
- **L-9.** A copied project folder shares navigation history and Sources list settings with its original (`InstallPaths.swift`, keyed by `project.uuid` only).
- **L-10.** No length caps on text read from the catalog (graph names, labels, notes).
- **L-11.** `GetSourceWorkspace` runs four lookups per Artifact (`artifactProto`); `SourcePageContext.refreshCoverFromStore` reloads the whole workspace after an Artifact add. Near S12-24.
- **L-12.** The first Sources list visit runs `graphprogress.List`, the activity load and the graph list, about 240 ms at 37k Subjects; group progress by `graph_id` alone. Near S12-25.
- **L-13.** `L10n.swift` is 11,205 lines; six members are unused (`armedEscHint`, `evidenceGraphStubBody`, `gateUnknownTitle`, `gateUnknownBody`, `ingestUnreadableTitle`, `ingestUnreadableHelp`).
- **L-14.** `WorkspaceLocation` is twenty optional fields keyed by surface; graph locations are built five ways.
- **L-15.** `PVFlowLayout` has one caller; `WorkspaceSession.evidenceGraphsBySource` has none. Near S12-26.
- **L-16.** Test net: `TestCallRouter`'s unknown-method row calls method 100 (`METHOD_UNPROMOTE_SUBJECT`); `PROVENENCIA_GRAPH_VERIFY=1` is not suite-clean (`promotealign` `TestMemoFollowsCommit/a_handle_merged`, `TestLargeSourceBudgets`); a `c.DB()` read inside `writes.Run` hangs on the single connection instead of failing. Near S12-33.
- **L-17.** Breadcrumbs, the canvas header title and DeleteImpact Citation rows truncate with no `.help`; a refused card save reads "…an artifact you removed" for an unchecked graph. Near S12-29.

**Nits.** `SourceGraph*` Swift type names against the `EvidenceGraph` rule; the `"filed"` filter shim in `SourcesListSettings.swift`; error codes compared as string literals in Swift; `FakeStore` (Platform) calling `SourcesListNavigation` (Features); `CoreInvoke.swift` and `dispatch.go` restating the proto `Method` enum; `graphfill/load.go` repeating `evidencegraphs.SourceOf`; the graph status rule decided in Swift from Go's counts; two glyphs for an Evidence graph; graph-form Save enabled with no change; a blocked graph button is a focusable no-op; `graphprogress` drops a user type keyed `source` without checking origin.
