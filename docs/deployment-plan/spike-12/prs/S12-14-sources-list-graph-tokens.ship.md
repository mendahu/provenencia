# Ship log — S12-14 Graph counts in the Sources list and on the Source page

**Plan:** [`S12-14-sources-list-graph-tokens.md`](S12-14-sources-list-graph-tokens.md) · **Branch:** `s12-14-sources-list-graph-tokens` · **PR:** #576
**State:** fixing 1

## Round 0 — implement
Head `9646520c`. Gate PASS (orchestrator run, against `f22b19a7`). PR #576 (draft). Report deviations (recorded in the plan's Landed note):
- `PersonDetailTests.presentationFollowsTheHandle` moved to `QueryRead` values (it called the old `ConclusionDetailPresentation` init).
- `evidenceGraphsRead(sourceID:)` reads `.loading` for a loaded list with no graph for the Source, including an empty list (D-17).
- `SourcesGraphToken` uses a private `SourcesGraphTokenFrame` and an `inert` style flag for the list-failed token.
- `PVStroke` README line sits in the component table next to the `FlowLayout` row.
- `FakeStore` filed count stays hard-coded 0.
- Running-app checks not run (plan §10 manual check, truncation at min width, `pvRightClickMenu` vs link click and VoiceOver, D-2 row repaint, D-10 highlight via environment key, tab order and focus rings). Left for the user's manual check.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. UX routed by `macos/App/**` and `Localizable.xcstrings`. SEC+PRF routed by `api/ffi/**`, `core/database/**`, and `macos/App/Features/**` models.

CON: clean (2 nits). SEC+PRF: clean (1 outside-lens note). UX: 1 major, 2 minors, 3 nits. MNT: 1 major, 4 minors, 2 nits.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | A failed progress map shows every graph as "not started", with no error | To user → Fix now → F-5 (user 2026-10-26: add a failed state) | verified: `SourcesGraphCounts.init(progress: nil, countsLoading: false)` falls to `.notStarted`; `evidenceGraphCountsLoading` is false once the read is `.failed`; `statusCountsError` returns nil without a status filter. The fix adds a fourth state to D-3's three and a token/card look the board does not have |
| MNT-1 | MNT | major | `QueryRead` added, but exact copies of its fold stay hand-written in Sources | Fix now → F-1 | verified: `SourceCopyModel.loadState` matches `read` line for line; `SourcePageModel.sync` `loadError` and `SourcePageView.swift:248` are `read`'s `.failed`. Same-pass cleanup; supersedes the §6 "Next adds" deferral for these. `statusCountsError` deliberately reports an error beside a stale value ("behave as today", D-7), so it gets a comment, not `read`. `status == .loading && value == nil` folds (SourceTypesModel, Metadata, Properties) differ from `read.isLoading` on `.idle` and are not copies; `canCite` reads `value == nil` only |
| MNT-3 | MNT | minor | Comments name deleted things (split row, entry graph, list-only warm) | Fix now → F-2 | verified by grep; dedupes CON-1, UX-4, UX-5, UX outside-lens notes |
| MNT-4 | MNT | minor | Count line hand-draws `PVSkeletonBar`; three-`Text` "not started" line | Fix now → F-3 | verified: `PVSkeletonBar` fills `borderSubtle` at the given height (radius 2, invisible at 1 pt). The three `Text`s mix mono and italic body, so the split stays with a comment |
| MNT-5 | MNT | minor | `PVStroke` README row sits in the component table | Fix now → F-4 | verified: README lines 34-42 vs 180 |
| UX-2 | UX | minor | A truncated graph name cannot be read in full by pointer | To user → Fix now → F-6 (user 2026-10-26: name in help) | help copy is the plan's token states table (board) |
| UX-3 | UX | minor | List-failed token speaks only the error, not what failed | To user → Fix now → F-7 (user 2026-10-26: say what failed) | label/help copy is the plan's token states table (board) |
| MNT-2 | MNT | minor | Token `onOpen`/`blockedReason` and style `blocked`/`inert` allow impossible combinations | To user → Fix now → F-8 (user 2026-10-26: reshape now) | D-11 fixes the `Kind.graph(_, counts:, onOpen:, blockedReason:)` shape |
| SEC-out | SEC+PRF → UX | minor | `PVFlowLayout` places a token at full width; can overflow at the minimum window width | Logged → manual check | needs the running app; plan §10 manual check covers truncation at min width |

Nits (logged): `CON-1 · SourcesGraphCountLine.swift:5 · stale "split row"` (→ F-2) · `CON-2 · PVList.swift:246 · stale "Sources split row" (plan Out, X-28)` · `UX-4 · Localizable.xcstrings sources.list.openSourcePage · comment says "left (filing) zone"` (→ F-2) · `UX-5 · L10n.swift needsArtifact comments · "graph zone"` (→ F-2) · `UX-6 · SourcesGraphToken.swift:84-86 · hard-coded ", " join in VoiceOver label` · `MNT-6 · SourcesModel.swift:264 · statuses(for:) reads the handle, not evidenceGraphProgress(graphID:)` · `MNT-7 · SourcesModel.swift:140 · inline graphsHandle`

User answers (2026-10-26, in chat): UX-1 add a failed count state (`.unavailable`); UX-2 and UX-3 apply; MNT-2 reshape now. These supersede D-3's three states, D-11's `Kind.graph` shape, and the token states table's help/label copy for the cells F-5 to F-8 name; the plan's Landed note records it.

### Fix plan 1

Order: majors (F-1, F-5), then minors. Record every supersession below in the plan's Landed note.

**F-1 — Move exact copies of the value/error/loading fold onto `QueryHandle.read`** (from MNT-1)
Files: `macos/App/Features/Sources/SourceCopyModel.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/Sources/SourcesModel.swift`. Skill: `add-workspace-place`.
Failing test first: none (no behavior change); existing `SourceCopyModel` / `SourcePageModel` tests must stay green.
Do:
- `SourceCopyModel.loadState`: `switch workspaceHandle.read { case .loaded: .ready; case .failed(let e): .failed(L10n.Errors.message(for: e)); case .loading: .loading }`. (`read` only yields `.failed` with an error, so the `?? L10n.Errors.unknown` fallback goes.)
- `SourcePageModel.sync`: `if case .failed(let e) = handle.read { loadError = e } else { loadError = nil }`. Leave its `isLoading` line (`status == .loading && value == nil` is false on `.idle`; `read.isLoading` is not, so they differ).
- `SourcePageView.swift:248`: `if case .failed = workspaceHandle.read { navigation.fallbackToSectionRoot() }`.
- `SourcesModel.statusCountsError`: do not change its logic (D-7 "as today"); add a one-line comment above it: it reports a failure even beside a stale value, because a status filter must not apply stale counts silently.
- Plan §6 "Next adds": drop `SourcePageModel.sync` from the "still fold by hand" list.
Check: `rg -n "status == \.error" macos/App/Features/Sources` prints only the two `SourcesModel.swift` `statusCountsError` lines; `rg -n "value != nil \{ return \.ready" macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-5 — A failed progress map shows a failed count state, not "not started"** (from UX-1)
Files: `macos/App/Features/Sources/SourcesGraphCounts.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/App/Features/Sources/SourcesGraphToken.swift`, `macos/App/Features/Sources/SourcesGraphCountLine.swift`, `macos/App/Features/Sources/SourcesListRow.swift`, `macos/App/Features/Sources/SourceGraphsSection.swift` (and whichever call site builds `SourcesGraphCounts` on the Source page), `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, tests. Skills: `add-localized-string`, `add-swift-test`.
Failing test first: `WorkspaceSessionTests.queryReadAndProgressReads` (or a new `progressFailureReadsUnavailable` beside it) — with the graph list loaded and the progress handle `applyFailure(clearValue: true)`, `SourcesGraphCounts(progress: nil, countsLoading: false, countsError: session.evidenceGraphCountsError)` is `.unavailable(message:)` with `L10n.Errors.message(for:)` of that error, not `.notStarted`. A second case: a stale progress value plus a failure still reads `.counts` (value wins, as `read`).
Do:
- `SourcesGraphCounts` gains `case unavailable(message: String)` (a String keeps it `Equatable`). `init(progress:countsLoading:countsError:)`: progress non-nil → as today; else `countsError` non-nil → `.unavailable(message: L10n.Errors.message(for: countsError))`; else as today.
- `WorkspaceSession` gains `var evidenceGraphCountsError: Error?` beside `evidenceGraphCountsLoading`: `if case .failed(let e) = evidenceGraphProgressHandle?.read { return e }; return nil`. Every `SourcesGraphCounts(...)` call site passes it.
- New string `sources.graph.countsUnavailable`, default "Counts could not load", comment "Evidence graph token help and Source page count line when graph counts failed to load". `L10n.Sources.graphCounts(.unavailable(m), style:)`: `.compact` and `.full` → that string; `.spoken` → "\(that string). \(m)".
- Token `countView`: `.unavailable` → `PVIcon(.danger, size: 11)` in `PVColor.danger`. Its help comes from F-6.
- Count line (`SourcesGraphCountLine`): `.unavailable` → `HStack(spacing: PVSpacing.space2) { PVIcon(.danger, size: 11) in PVColor.danger; Text(verbatim: graphCounts(_, .full)) in PVFont.body(size: PVTypeScale.micro) PVColor.textMuted }` with `.help(m)`.
- `SourcesGraphCounts` doc comment: four states.
Check: the failing test passes; `graphProgressCopyMatchesBoard` (or `graphTokenCopyMatchesBoard`) adds the `.unavailable` row for all three styles; scripts/check-pr.sh PASS.

**F-2 — Comments that name deleted things** (from MNT-3, CON-1, UX-4, UX-5)
Files: `SourcesGraphCountLine.swift:4-5`, `SourcesModel.swift` (`graphCountsLoading` doc), `WorkspaceSession.swift:259`, `:280`, `L10n.swift` comments on `needsArtifact`, `needsArtifactTooltip`, `graphNotStarted`, `graphSubjectCount`, `graphObservationCount`, `graphCountsLoading` and their `Localizable.xcstrings` comments, and the `sources.list.openSourcePage` comment.
Do: count line → "The Source page's graph cards draw it."; `graphCountsLoading` → "True while the graph list or progress map can still arrive."; session → "Skips when no place warmed the map" and "…or the next place visit that warms it retries"; every "graph zone" / "graph-zone" → "graph token" (or "graph count line" where the Source page reads it); `openSourcePage` → describe the Source link. Comments only; keep L10n `comment:` and the catalog in sync.
Check: `rg -n "split row|entry graph|graph zone|graph-zone|filing\) zone" macos/App` prints only `DesignSystem/Components/List/PVList.swift:246` (plan Out, X-28); scripts/check-pr.sh PASS (localizable).

**F-3 — Count line loading bar uses `PVSkeletonBar`** (from MNT-4)
Files: `SourcesGraphCountLine.swift`.
Do: `.loading` → `PVSkeletonBar(width: nil, height: 1)` keeping both paddings. Above the `.notStarted` `HStack`, one comment: the line is split only so "not started" can be italic body beside mono counts; the words come from the same strings `graphCounts(.notStarted, .full)` joins.
Check: scripts/check-pr.sh PASS.

**F-4 — `PVStroke` README row in the Tokens table** (from MNT-5)
Files: `macos/App/DesignSystem/README.md`.
Do: move the `Stroke | Tokens/PVStroke.swift` row from the component table into the Tokens table (lines ~34-39), source column "—"; add `PVStroke` to the "Reach for" list on line ~42.
Check: `rg -n "PVStroke" macos/App/DesignSystem/README.md` shows the Tokens-table row and the Reach-for line only.

**F-6 — Token help starts with the graph name** (from UX-2)
Files: `SourcesGraphToken.swift`, `L10n.swift`, `Localizable.xcstrings`, the token copy test.
Failing test first: `graphTokenCopyMatchesBoard` — help for "Jones household" with counts 12/40 is "Jones household — 12 subjects · 40 observations"; for a blocked "Unlinked", "Unlinked — Add an artifact to open this evidence graph".
Do: new `L10n.Sources.graphTokenHelp(name:detail:)`, key `sources.list.graphTokenHelp`, default "%1$@ — %2$@", comment "Evidence graph token tooltip; 1 is the graph name, 2 its counts or why it cannot open". Graph token `.help(graphTokenHelp(name: graph.name, detail: graphCounts(counts, .full)))`; blocked token `.help(graphTokenHelp(name: graph.name, detail: reason))`. VoiceOver labels unchanged.
Check: the test passes; scripts/check-pr.sh PASS.

**F-7 — List-failed token says what failed** (from UX-3)
Files: `SourcesGraphToken.swift` (`failedToken`), `L10n.swift`, `Localizable.xcstrings`, the token copy test.
Failing test first: `graphTokenCopyMatchesBoard` — the list-failed label is "Evidence graphs could not load".
Do: new `L10n.Sources.graphsUnavailable`, key `sources.list.graphsUnavailable`, default "Evidence graphs could not load", comment "Sources list token VoiceOver label and tooltip lead when the Source's Evidence graph list failed to load". `failedToken`: `.accessibilityLabel(graphsUnavailable)`, `.accessibilityHint(message)`, `.help(graphTokenHelp(name: L10n.string(graphsUnavailable), detail: message))` (reuses F-6's format).
Check: the test passes; scripts/check-pr.sh PASS.

**F-8 — Token open/blocked and style as single enums** (from MNT-2)
Files: `SourcesGraphToken.swift`, `SourcesListRow.swift`, tests that build `Kind.graph`.
Do: `Kind.graph(CatalogEvidenceGraph, counts: SourcesGraphCounts, open: Open)` with `enum Open { case opens(() -> Void), blocked(reason: String) }`. `SourcesListRow.tokens`: `graphBlockedReason(for:)` non-nil → `.blocked(reason:)`; else `graphLocation(for:graph:)` → `.opens { navigation.go(to:) }`; if both are nil (cannot happen today: `graphLocation` is nil only when blocked), use `.blocked(reason: L10n.string(L10n.Sources.graphNeedsArtifactTooltip))` — check that `graphLocation` returns nil only when `graphBlockedReason` is non-nil, and say so in a comment. `SourcesGraphTokenButtonStyle(look: Look)` with private `enum Look { case live, blocked, inert }`; drop the two Bools and the `live` fold. No visible change.
Check: `rg -n "blocked: (true|false)|inert:" macos/App/Features/Sources` prints nothing; scripts/check-pr.sh PASS.
