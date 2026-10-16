# Ship log — S10-41 Sources list: last updated, status filter, search, saved settings

**Plan:** [`S10-41-sources-list.md`](S10-41-sources-list.md) · **Branch:** `s10-41-sources-list` · **PR:** #405
**State:** ready

Base: `main` @ `daf4e086`. Plan review To user item 1 (move the cell into S10-22) is moot: S10-22 landed and the plan was refreshed so D-10 extracts the cell here.

## Round 0 — implement
Head `ed10f183` (local, not pushed). Gate not run. STEP-1 (`af048c35`) and STEP-2 (`ed10f183`) done with their targeted tests passing. STEP-3 to STEP-6 not started.

**Stopped on:** the plan's stop condition "`WorkspaceNavigation.projectUuid` is empty when the Sources list first mounts after relaunch". The orchestrator checked it against the code:
- `WorkspaceView.swift:27-35`: `init` builds `WorkspaceNavigation()`, and `projectUuid` starts as `""` (`WorkspaceNavigation.swift:43`).
- `WorkspaceView.swift:83-86`: `navigation.attachProject(uuid:)` runs only in `.onAppear`, which comes after the first body evaluation.
- `WorkspaceContent.swift:26` always mounts `WorkspaceDestinationHost`. With the seed `.sectionRoot(.sources)`, `SourcesListView.init` builds `SourcesModel` (D-13) while `projectUuid` is still empty. `SourcesListSettingsStore.init?` returns nil, and `@State` keeps that model, so settings never load or save.

Changing D-13 is a plan change, so it goes back to plan-pr. Possible fixes: (a) attach the project in `WorkspaceView.init`, where `model.project?.uuid` is already available, before the host mounts; (b) resolve the store lazily in `SourcesModel` once `projectUuid` is set, which conflicts with D-13's rule that restored values are the initial values.

**Resolution (user, 2026-10-15):** "attach the project in WorkspaceView.init before the host mounts, keeping restored values as initial values." D-13 was revised with plan-pr in `9d224a74`. A diff-scoped review came back Ready with changes, and all of it is fixed (review log, "D-13 revision"). Resuming round 0 at STEP-3.

Resumed: head `9e741ec1` (STEP-3 `870001d1`, STEP-4 `2be3af9c`, STEP-5 `cb66caae`, STEP-6 docs `9e741ec1`). Draft PR #405. Orchestrator gate PASS (against `daf4e086`). Deviations reported:
- The single `.onChange(of: model.settings)` landed in STEP-3, not STEP-5.
- The STEP-4 `rg` check also matches `L10n.Promote.updatedDetail`, an unrelated Promote string.
- Plan and review-log links now point at the archived brief.
- The pbxproj gets `AA1041…` ids and a `Recipes/LastUpdated` group.

Not run: the §10 manual check and the temporary-`print` check for the stop condition (both need the real app).

## Round 1 — review
Base `daf4e086`, head `9e741ec1`. Verdicts: CON changes needed · MNT clean · UX clean · SEC clean (no findings) · PRF changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | `visibleSources` filters and sorts the whole list 4× per render; each search keystroke is a render | Fix now → F-1 | Verified: `visibleSources` is computed (`SourcesModel.swift:217`). It is read by `countLine` (`:274`) and at `SourcesListView.swift:292`, `:312`, `:343`. Reviewer benchmark: ~57–71 ms at 2,000 rows |
| MNT-3 | MNT | minor | Progress handle looked up three ways beside the new `typesHandle` | Fix now → F-1 (ride-along) | F-1 touches the same accessors |
| PRF-x (MNT outside-lens) | PRF | minor | `graphProgress(for:)` repeats `queryHandle` per row inside the status filter | Fix now → F-1 (ride-along) | Same function as F-1 |
| CON-3 | CON | minor | Archived S10-D8 brief's relative links broken after the move | Fix now → F-2 | STEP-6 Check "links resolve" fails; six targets broken (reviewer loop) |
| CON-1 | CON | major | §10 manual check and the D-13 stop-condition `print` check not run; AC-4.2, AC-4.5 lack evidence | To user (manual check) | Verified: no test reaches `WorkspaceView.init`. Ship-pr gives the manual check to the user; the report puts the relaunch check (§10 step 7, the `sources-list/<hex>.json` entry) first. Not code-fixable by the implementer |
| PRF-2 | PRF | minor | Every keystroke rewrites the settings JSON on the main actor; the model reloads on host re-init | Follow-up | Small fixed-size file; debounce changes D-14's save timing. Also raised by MNT and UX outside-lens |
| PRF-3 | PRF | minor | `WorkspaceView.init` reads the history file on every parent re-evaluation | Follow-up | Accepted by D-13's rationale (one file read, no store calls) |
| PRF (CON outside-lens) | CON | — | Claim-status writes might not refresh `filed_count` | Rejected | Claim writes return `CatalogWriteEffects` with the Source (`effects/registry.go:254-256` `Source: from("subject_id", field("source_id"))`), which `apply(_ effects:)` patches; plan Current state row "Progress refresh" |
| CON-2 | CON | minor | T-4 is a non-table test, contrary to `go-tests.mdc` | Follow-up | Fix plan does not touch the Go test file; coverage is complete |
| CON-4 | CON | nit | T-11 fixture reads "2 of 8" rather than the plan's "3 of 8" | Logged | Rule is proven |
| MNT-1 | MNT | minor | `FakeStore.primaryTypeKeys` duplicates `EvidencePrimaryKind` | Follow-up | Add to `ideas/shared-list-helpers.md` with the four older sites |
| MNT-2 | MNT | minor | `scanSubjects` copies `scanGrouped`; "source" reification rule spelled twice | Follow-up | Go file not in fix plan |
| MNT-4 | MNT | minor | `add-workspace-place` skill does not point at `PVLastUpdated` | Follow-up | One-line skill edit |
| MNT-5 | MNT | minor | Progress patch triggers live in a session switch, not the registry Spec | Follow-up | Idea; precedent from S8-08 |
| UX-1 | UX | minor | Clear filters drops keyboard focus | Follow-up | With UX-2: one search focus binding |
| UX-2 | UX | minor | No ⌘F for the Sources search (Properties has it) | Follow-up | Same as UX-1 |
| UX-3 | UX | minor | The counts-error callout does not say that the status filter caused it | Follow-up | Copy change; product wording |
| UX-4 | UX | minor | Truncated titles have no `.help` | Follow-up | `SourcesSplitRow` not in fix plan |
| UX-5 | UX | nit | Search field label relies on prompt | Logged | Follows plan §6 and Properties precedent |

### Fix plan 1

**F-1 — Compute the visible rows once per render** (from PRF-1, MNT-3, PRF outside-lens)
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/Sources/SourcesListView.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`. Skill: `add-swift-test`.
Failing test first: none. This is a performance fix; behavior is unchanged and the existing `SourcesModelTests` (`filtersCombineAndClear`, `statusFilterWaitsForCounts`, `restoredTypeMissingFromTypesShowsEveryRow`) must stay green.
Do:
- In `SourcesModel`, add `private var progressHandle: QueryHandle<[String: SourceGraphProgress]>?` beside `typesHandle`. Use it in `graphProgress(for:)`, `graphCountsLoading` and `statusCountsError`.
- In `visibleSources`, read the progress map once before the status filter, not per row through `graphProgress(for:)`.
- Change `countLine` to `func countLine(visibleCount: Int) -> String`, keeping the same rules, and update its test callers.
- In `SourcesListContent`, compute `let rows = model.visibleSources` once per `body`. Pass `rows` (or `rows.count` and `rows.map(\.id)`) to the header's count line and to `listBody`, so `visibleSources` is evaluated exactly once per render.
- Do not add caching, debouncing or a stored rows array.
Check: `rg -n "visibleSources" macos/App` shows one read in `SourcesListView.swift`. `SourcesModelTests` pass. `scripts/check-pr.sh` PASS.

**F-2 — Fix the archived S10-D8 brief's relative links** (from CON-3)
Files: `docs/deployment-plan/spike-10/design/archive/S10-D8-sources-list.md`.
Do: add one `../` to each relative link that no longer resolves from `design/archive/`: `../prs/…`, `../deployment-plan.md`, `../completed.md`, `../../../design-system-layers.md`, `../../../../.agents/skills/…`. Leave the other archived briefs alone (follow-up).
Check: every relative `](…)` target in the file resolves (`test -e` loop). `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `e2314481` (F-1 `e2ec3a6c`, F-2 `e2314481`). Gate PASS (orchestrator, against `daf4e086`). F-1 done: `visibleSources` read once in `SourcesListContent.body`. F-2 done: all six links resolve. Deviations: none.

## Round 2 — review
Scope `9e741ec1...e2314481`, lenses PRF, MNT and CON. All clean: PRF-1, MNT-3, the PRF ride-along and CON-3 are resolved. No new blocker. No Fix plan 2.

## Result
Gate PASS · CI see PR · open: none (blocker/major) · to user: CON-1, the §10 manual check, including the D-13 relaunch and open-path check · follow-ups: PRF-2, PRF-3, CON-2, MNT-1, MNT-2, MNT-4, MNT-5, UX-1+UX-2, UX-3, UX-4
