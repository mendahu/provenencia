# Ship log — S12-21 The canvas keeps its graph and shows its failures

**Plan:** [`S12-21-canvas-keeps-its-graph.md`](S12-21-canvas-keeps-its-graph.md) · **Branch:** `s12-21-canvas-keeps-its-graph` · **PR:** #596
**State:** fixing 2

## Round 0 — implement
Head `1fee85d4`. Gate PASS (orchestrator run, xcode-test 60s). Report deviations: T-4 case (d) is its own test (needs a different store); hosting tests T-1 and T-8 wait up to 15 s instead of the plan's 2 s after one load timeout in a full-suite run (passes alone well within 2 s, so `.task` does run; not treated as the Stop condition, left for review); commits grouped by behavior, not per STEP; T-8 attaches history with project uuid `"project-1"`. Manual check (§10) not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: performance paths (`macos/App/Features/**` models, graph and cache code — `QueryHandle`, `WorkspaceSession+EvidenceGraphs`); no security path, but the combined reviewer runs on either.

Verdicts: CON Clean · MNT Changes needed · UX Clean · SEC+PRF Clean (no findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `SourcePageContent.sourceID` is dead after D-4 | Fix now → F-1 | verified: `SourcePageView.swift:78` declared, no reader below line 100 |
| CON-1a | CON | minor | T-1 red run not recorded | Rejected | implementer round-0 report: T-1 failed (1 issue) with `.id(place?.viewIdentity)` commented out, commit `9ee2a5a1`; recorded here |
| CON-1b | CON | minor | AC-3.5 (graph drops out of an open canvas) has no test | Fix now → F-2 | verified: T-8 mounts on an already-gone graph only |
| CON-out (from MNT) | CON | minor | `goneGraphFallsBackToTheSourcesList` leaves its temp history dir | Fix now → F-2 | minor in a test this PR adds |
| MNT-2 | MNT | minor | Raw `handle.error` reads beside the new `QueryHandle.failure` | Fix now → F-3 (same-pass cleanup, all three readers) | verified: `SourcesModel.swift:206`, `SourceTypesModel.swift:103`, `SourceArtifactsSection.swift:413` |
| MNT-3 | MNT | minor | "Invalidate, then query" hand-written in three places | Fix now → F-4 (same-pass cleanup; supersedes D-6's "copy its two lines") | verified: `EvidenceGraphModel.readGraphAgain`, `SourceCopyModel.reloadWorkspace`, `SourcePageModel.warmFromSession` |
| MNT-4 | MNT | minor | S12-37 map entry does not carry the notice file move | Later PR | written into S12-37's **Ships** in `deployment-plan.md`, commit `17aad8e6` |
| UX-1 | UX | minor | Read again gives no sign of a running read beside stale rows | Fix now → F-5 | verified: `beginLoading(stale: true)` leaves `status == .error`, so the callout is unchanged during the read |
| UX-2a | UX | minor | A composer load error stays under Transcription after a later refresh succeeds | Fix now → F-6 | verified: `refreshCitationLines` success branch does not clear `fields.error` |
| UX-2b | UX | minor | Load error shows in the Transcription field slot and does not name the Citation | Follow-up | copy and placement change beyond D-8/D-10; the slot predates this PR (`loadCitationPreservingAuthoring`) |
| UX-3 | UX | minor | `.readError` outranks a newer delete error, missing file, and placing hint | To user → Fix now → F-7 | contradicts D-7's binding order. User (2026-10-27): "put the delete error and the missing-file notice ahead of the read error on this PR, and update D-7 to match." D-7 and T-7 updated; the placing hint stays behind the read error |
| CON-2 | CON | nit | Hosting tests wait 15 s, plan said 2 s | Logged | the Stop condition guards `.task` not running; it runs (passes alone in < 2 s); the longer wait only absorbs full-suite load |

Nits (logged): `CON-2 · WorkspaceDestinationHostTests T-1/T-8 · 15 s timeout vs plan's 2 s` · `MNT-5 · CitationComposerModelTests.swift:1929 · name the error TestFailure` · `MNT-6 · ResolvedPlace.swift:8 · doc comment duplicates PlaceRegistry.Spec.viewIdentity's` · `UX-4 · EvidenceGraphView.swift:219-221 · silent fallback to Sources list; for S12-30's announcement work`

### Fix plan 1

**F-1 — Delete the dead `SourcePageContent.sourceID`** (from MNT-1)
Files: `macos/App/Features/Sources/SourcePageView.swift`.
Do: delete `let sourceID: String` from `SourcePageContent` and the `sourceID: sourceID,` argument where `SourcePageView.body` builds it.
Check: `rg -n 'sourceID' macos/App/Features/Sources/SourcePageView.swift` shows no hit at or after the `SourcePageContent` declaration; scripts/check-pr.sh PASS.

**F-2 — Test AC-3.5 and clean up the hosting test's temp dir** (from CON-1b, MNT outside-lens)
Files: `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift`. Skill: `add-swift-test`.
Failing test first: `openGraphThatGoesFallsBackToTheSourcesList` *(new)* — mount the host on `.evidenceGraph(sourceId: "src-1", graphId: "g-a", …)` as T-8 does, wait until the `g-a` handle settles, then `session.setQueryValue(session.evidenceGraphsKey, value: [CatalogEvidenceGraph]())` (or the equivalent seeding API the suite already uses) → wait until navigation's location is the Sources list root, as T-8 asserts. Run it once with the `.onChange(of: model.graphIsGone, initial: true)` changed to fire only on mount (or removed) to see it fail, then restore.
Do: also remove the temporary history directory `goneGraphFallsBackToTheSourcesList` creates (`defer { try? FileManager.default.removeItem(at:) }`), and do the same in the new test. Share the setup with T-8 through a private helper rather than copying it.
Check: both tests pass; scripts/check-pr.sh PASS.

**F-3 — Read handle failures only through `QueryHandle.failure`** (from MNT-2, same-pass cleanup)
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`.
Do: `SourcesModel.loadError` → `sourcesHandle?.failure ?? typesHandle?.failure`; `SourceTypesModel.loadError` → `typesHandle?.failure ?? fieldsHandle?.failure`; `SourceArtifactsSection` `reloadError` → `context.session.evidenceGraphsHandle?.failure`. The one difference is during a retry with no value (status `.loading`): the old error no longer shows, and the views' existing loading state shows instead. Read each view that reads these `loadError`s and confirm it shows a loading state in that case; if one would show nothing at all, stop and report it in the fix report rather than adding UI.
Check: `rg -n 'Handle\??\.error\b|handle\.error\b' macos/App` prints only `QueryHandle.swift`'s own lines; existing Sources, Source types and artifacts tests pass; scripts/check-pr.sh PASS.

**F-4 — One `WorkspaceSession.reread` for "invalidate, then query"** (from MNT-3, same-pass cleanup; supersedes D-6's "copy its two lines")
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `EvidenceGraphModel.swift`, `SourceCopyModel.swift`, `SourcePageModel.swift`.
Do: add `@discardableResult func reread<Value>(_ key: CatalogQueryKey) -> QueryHandle<Value>` next to `query`, doc-commented, that calls `invalidate(key)` then returns `query(key)`. Move `readGraphAgain()`, `reloadWorkspace()` and `warmFromSession()`'s workspace key onto it.
Check: `rg -n -A2 '\.invalidate\(' macos/App | rg 'query\('` prints nothing (only `reread` itself pairs them); existing tests pass; scripts/check-pr.sh PASS.

**F-5 — Read again shows the read it started** (from UX-1)
Files: `EvidenceGraphModel.swift`, `EvidenceGraphView.swift` (the `.readError` arm), `EvidenceGraphModelTests.swift`.
Failing test first: extend T-3 `graphReadErrorShowsAndReadsAgain` (or a new `readAgainShowsItIsReading`) — with a stale value and a failed refetch, call `readGraphAgain()` and before the load settles assert `model.isRereadingGraph == true` *(new computed: the graph handle's `isFetching`)*; after a second failure it is `false` and `graphReadError` is set.
Do: keep `graphReadError` as D-6 defines it. In the `.readError` arm, while `model.isRereadingGraph`, disable the **Read again** `PVButton` and show the canvas's existing small `ProgressView` beside it (no new component, same callout).
Check: the test passes; scripts/check-pr.sh PASS.

**F-6 — A later successful load clears the composer's load error** (from UX-2a)
Files: `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `CitationComposerModelTests.swift`.
Failing test first: extend T-6 `lineRefreshShowsAFailedCitationLoad` — after the failing refresh sets `fields.error`, make the store succeed and call the refresh again → `fields.error == nil`.
Do: in `showingLoadFailure`, remember the message it set (a private `loadFailureMessage: String?`); on a later successful `work()`, clear `fields.error` only when it still equals that message, then clear `loadFailureMessage`. Do not clear other errors (save or validation).
Check: the test passes, and T-5 and existing composer tests still pass; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `d03e5ba7`. Gate PASS (orchestrator run, xcode-test 58s). F-1 to F-6 done, one commit each. F-2 red run: with `.onChange(initial: true)` replaced by `.onAppear`, the new test failed ("Expectation failed: fellBack"). F-5's red state was a compile failure (the new computed did not exist). F-6 red run: `fields.error` stayed set with the clear disabled. F-3: with no value, a retry now shows the existing spinner (`SourcesListView`, `SourceTypesListPane`) instead of the old error.

Minors checked by the orchestrator (no round-2 reviewer):
- CON-1b, CON-out (F-2): **Resolved** — `openGraphThatGoesFallsBackToTheSourcesList` added; both hosting tests `defer`-remove their temp dir (`WorkspaceDestinationHostTests.swift:106,120`).
- MNT-2 (F-3), MNT-3 (F-4): **Resolved** — the fix plan's two greps print nothing. MNT rechecks these in round 2 anyway, since it has MNT-1.
- UX-1 (F-5): **Resolved** — `isRereadingGraph` disables Read again and shows a small `ProgressView` in the callout.
- UX-2a (F-6): **Resolved** — `showingLoadFailure` clears only the message it set.

## Round 2 — review
Reviewers: MNT (had the round-1 major MNT-1). Not sent: CON, UX, SEC+PRF (no Fix-now blocker or major).
Verdict: MNT Clean. MNT-1 **Resolved** (commit `bd57ad33`; `SourcePageContent` has no `sourceID`). MNT-2 and MNT-3 rechecked by MNT: Resolved. No new findings. No Fix plan 2.

Nits (logged): `MNT-7 · docs/macos-client-patterns.md · mention WorkspaceSession.reread as the way to read a cached key again`

## Finish
Merged `origin/main` @ `331bb3c3` (docs only: S11-47 plan). Gate PASS after the merge. Plan Landed note updated with the review-round changes (commit `3dae2dbe`).

## Result
Gate PASS · CI green · open: none · follow-ups: UX-2b (composer load errors shown in the Transcription field slot, without the Citation's ref), MNT-7 (patterns doc names `reread`) · to user: UX-3 answered (reorder; F-7) · CON miss: none

## User answer — UX-3
The user chose to reorder (quoted above). The orchestrator updated D-7 and T-7 in the plan and added a line to its Landed note. This fix round changes the order of two existing notices and nothing else, so it goes to no reviewer; the orchestrator reads the fix.

### Fix plan 2
**F-7 — A failed delete and a missing file show ahead of the read error** (from UX-3, user decision)
Files: `macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift` (`canvasNotice`), `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Failing test first: `canvasNoticeOrder` case (f) per the plan's T-7: as (a) (failed graph read), with a failed Subject delete (`deleteError` set, `pendingImpact == nil`) → `.deleteError(_)`; as (a), with `fileOpenError` for the selected file → `.fileNotFound(_)`. Build the states the way the suite's existing delete-error and file-not-found tests do; do not add FakeStore behavior.
Do: in `canvasNotice`, move the `graphReadError` → `.readError` check below the `.deleteError` and `.fileNotFound` checks and above `!canCite` → `.noArtifact`. Update the comment on case (a) ("outranks the rest") and any doc comment on `canvasNotice` or `EvidenceGraphCanvasNotice` that states the order.
Check: `canvasNoticeOrder` passes, and the rest of `EvidenceGraphModelTests` passes; scripts/check-pr.sh PASS.
