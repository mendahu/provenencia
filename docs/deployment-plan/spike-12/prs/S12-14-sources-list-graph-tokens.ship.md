# Ship log — S12-14 Graph counts in the Sources list and on the Source page

**Plan:** [`S12-14-sources-list-graph-tokens.md`](S12-14-sources-list-graph-tokens.md) · **Branch:** `s12-14-sources-list-graph-tokens` · **PR:** #576
**State:** ready (revision 2)

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

## Round 1 — fix
Head `7f711dbb`. Gate PASS (orchestrator run). F-1 to F-8 done. Deviations: F-5/F-6/F-7 tests written first but not run red separately (F-5's could not compile before the fix); `SourcePageModel.swift:110` (`warmFromSession` wait loop, not a fold) also matches F-1's grep and stays; F-2/F-3/F-5–F-8 share one commit (same files); `PVIcon.swift:81` "graph zone" comment reworded. Minors F-2, F-3, F-4 read by the orchestrator: Resolved.

## Round 2 — review
Reviewers: MNT, UX (each had a Fix-now major). Scope `9646520c...7f711dbb`. Not re-run: CON, SEC+PRF (clean in round 1).

MNT: MNT-1 Resolved (copy sheet, Source page model and view read through `QueryHandle.read`; `evidenceGraphCountsError` uses the same fold). No new blockers.
UX: UX-1 Resolved (`.unavailable(message:)` drawn in the token and count line; a stale value still wins; every `SourcesGraphCounts(` call site passes `countsError`). UX-2, UX-3, F-8 read as resolved. No new blockers.
No Fix plan 2.

## Result
Gate PASS (`7f711dbb`, `origin/main` unmoved at `f22b19a7`) · CI green · open: none · follow-ups: none · to user: UX-1, UX-2, UX-3, MNT-2 (answered 2026-10-26, fixed) · logged: SEC-out (min-width overflow, manual check), nits CON-2, UX-6, MNT-6, MNT-7 · CON miss: none

## Revision 2 — board redraw (2026-10-27)

After the user's manual check of revision 1 (a flicker on the Source link hover, fixed in `37571892`), the user redrew the list on the board: four columns, a status badge and a 52 pt button per graph. The plan was revised with plan-pr and passed review-plan in two rounds (Fits — Ready). The ship loop restarts at round 0 on the same branch; revision 1's rounds above stay as history.

### Round 0 — implement
Head `62d34c8e`. Gate PASS (orchestrator run). Commits `d809031a` (STEP-6), `91588926` (STEP-7), `27bdada5` (STEP-8), `62d34c8e` (STEP-9). Report deviations: the plan's §4 Token row renamed to `SourcesGraphButton.swift` (the old path failed `check-plan-docs`); the map row status reads `In review #576`; the docs describe revision 1 as "a row of small graph pills" to pass the STEP-9 grep; `EvidenceStatus.badge` is `@MainActor`; an earlier gate run had 8 Omnibar / CatalogSearch test failures outside this PR that did not recur (flaky); intermediate commits not gated one by one; app not launched (manual check open).

### Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. UX by `macos/App/**` and `Localizable.xcstrings`; SEC+PRF by `macos/App/Features/**` models (per-row reads, `onGeometryChange`). Scope: revision 2, `c7ebefaf...62d34c8e`, on the PR base `f22b19a7`.

CON: clean (1 minor, 2 nits). SEC+PRF: clean. MNT: 2 majors, 2 minors, 3 nits. UX: 4 minors, 1 nit (no blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Badge look and a SwiftUI view on the model enum; `SourcesModel` imports SwiftUI | Fix now → F-1 | verified: `SourcesModel.swift:3` `import SwiftUI`, `:46-69`; no other `Features/*/*Model*.swift` imports SwiftUI; precedent `ReconciledValueRow.badge`. An extension of `EvidenceStatus` in a view file keeps D-22's one owner |
| UX-1 | UX | minor | VoiceOver reads each status twice (badge and button) | To user → Fix now → F-1 (user 2026-10-27: hide the badge from VoiceOver) | D-22 and the states table asked for both; the user chose the button's label as the one reading |
| MNT-2 | MNT | major | No-Artifact and failed blocks use the tooltip key as their VoiceOver label | Fix now → F-2 | verified: `SourcesGraphButton.swift:44`, `:75` call `graphButtonHelp`; the states table says comma join. Dedupes UX-4 |
| UX-4 | UX | minor | Same as MNT-2 | Fix now → F-2 | dup of MNT-2 |
| MNT-3 | MNT | minor | Caption and row each restate the column frame chains | Fix now → F-3 | verified: `SourcesListView.swift:339-343` vs `SourcesListRow.swift:54-58`; new code in this PR |
| MNT-4 | MNT | minor | Style → `NotStartedReading` mapping written twice | Fix now → F-4 | verified: `SourcesGraphCountLine.swift:18-19`, `:64` |
| UX-2 | UX | minor | Compact layout: VoiceOver never says the type | Fix now → F-5 | verified: the link's value omits `typeLabel` |
| UX-3 | UX | minor | Truncated type and title have no help | Fix now → F-6 | the row this PR adds; S10-42 |
| CON-1 | CON | minor | Manual check not run | To user (manual check) | the user's §10 check before merge |
| DOC | orchestrator | minor | Docs call revision 1 "a row of small graph pills", a coined term | Fix now → F-7 | `deployment-plan.md:104`, `:141`, `:339`; memory rule "no coined terms" |

Nits (logged): `CON-2 · plan §4 Token row lists the old type names (history)` · `CON-3 / MNT-7 · PVLastUpdated.swift:3-6 · rewrap doc comment` (→ F-7) · `MNT-5 · L10n.swift:5912-5916 · zeroCounts rebuilds the counts join; reuse .counts` (→ F-4) · `MNT-6 · SourcesListRow.swift:95,191 · empty typeLabel means none; String?` · `UX-5 · SourcesGraphButton.swift:160-166 · italicDetail lacks truncationMode(.tail)` (→ F-2)

User answer (2026-10-27, in chat): UX-1 — hide the badge from VoiceOver; the button's label carries the status. Supersedes D-22's badge accessibility label and the `statusBadgeLabel` string; record in the plan's Landed note at merge.

### Fix plan 1 (revision 2)

**F-1 — Badge look in the view layer; badge hidden from VoiceOver** (from MNT-1, UX-1)
Files: `SourcesModel.swift`, `SourcesListRow.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`.
Do: move `badgeTone`, `badgeIcon`, `badgeSubtle` and `badge` into `extension SourcesModel.EvidenceStatus` at the bottom of `SourcesListRow.swift` (beside `statusSlot`), and drop `import SwiftUI` (and `@MainActor` on `badge` if it no longer needs it) from `SourcesModel.swift`. `badge` becomes `PVBadge(label, tone:, icon:, subtle:)` with `.accessibilityHidden(true)`; delete `L10n.Sources.statusBadgeLabel` and the `sources.list.statusBadge` catalog entry. T-14 drops its `statusBadgeLabel` assertion and keeps the rest.
Check: `rg -n 'import SwiftUI' macos/App/Features/Sources/SourcesModel.swift` and `rg -n 'statusBadge' macos` print nothing; T-14 passes.

**F-2 — A block label key for the No-Artifact and failed blocks** (from MNT-2, UX-4, UX-5)
Files: `SourcesGraphButton.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`.
Do: add `L10n.Sources.graphBlockLabel(title:detail:)`, key `sources.list.graphBlockLabel`, "%1$@, %2$@", comment "VoiceOver label of a Sources list graph block (no Artifact, or the graph list failed): its title and its second line". Use it for both blocks' `.accessibilityLabel`; `graphButtonHelp` stays for `.help` only. Add `.truncationMode(.tail)` to `italicDetail`. T-12 asserts `graphBlockLabel(title: "No evidence graph yet", detail: "Needs an artifact")` == "No evidence graph yet, Needs an artifact".
Check: `rg -n 'accessibilityLabel.*graphButtonHelp' macos/App` prints nothing; T-12 passes.

**F-3 — One frame chain per column for caption and row** (from MNT-3)
Files: `SourcesListColumns.swift`, `SourcesListView.swift`, `SourcesListRow.swift`.
Do: add `View` extensions in `SourcesListColumns.swift`: `sourcesSourceCell()`, `sourcesTypeCell(_ columns:)`, `sourcesStatusCell(_ columns:)`, `sourcesGraphCell(_ columns:)`, each applying that column's existing padding and frame in today's order (Source: `.padding(.trailing, sourceTrailing).frame(maxWidth: .infinity, alignment: .leading)`; Type: `.padding(.trailing, cellTrailing).frame(width: type, alignment: .leading)`; Status: `.padding(.trailing, cellTrailing).frame(width: status, alignment: .leading)`; Graph: `.frame(width: graphs).padding(.leading, graphLeading)`). The caption and the row call them; extra modifiers that differ (the row's `minHeight: firstSlot`, the caption's vertical padding) stay at the call sites. No visual change.
Check: `rg -n 'cellTrailing|graphLeading|sourceTrailing' macos/App/Features/Sources/SourcesListView.swift macos/App/Features/Sources/SourcesListRow.swift` prints nothing; scripts/check-pr.sh PASS.

**F-4 — One style → reading mapping; reuse the counts join** (from MNT-4, MNT-5)
Files: `SourcesGraphCountLine.swift`, `L10n.swift`.
Do: `SourcesGraphCountLine.Style.notStartedReading` (computed: `.button` → `.zeroCounts`, `.card` → `.words`), read by the label and `countsText`; delete the private `notStarted` property and the inline ternary. In `graphCounts`, `.notStarted` with `.zeroCounts` returns `graphCounts(.counts(subjects: 0, observations: 0, filed: 0), style: style)`.
Check: T-11 passes unchanged.

**F-5 — The type in VoiceOver when compact** (from UX-2)
Files: `SourcesListRow.swift`.
Do: when `columns.isCompact`, the link's accessibility value base is `"\(source.title), \(typeLabel), \(source.ref)"`; wide stays `"\(source.title), \(source.ref)"` (the Type cell is read as its own text).
Check: scripts/check-pr.sh PASS.

**F-6 — Help on a truncated title and type** (from UX-3)
Files: `SourcesListRow.swift`.
Do: `.help(title)` on the link's title `Text`; `.help(typeLabel)` on the Type cell and on the compact second line's type `Text`.
Check: scripts/check-pr.sh PASS.

**F-7 — Plain words for revision 1 in the docs; doc comment rewrap** (from DOC, CON-3, MNT-7)
Files: `docs/deployment-plan/spike-12/deployment-plan.md` (lines 104, 141, 339), `macos/App/DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift`.
Do: replace "the row of small graph pills" / "a row of small graph pills" with "the tokens under each Source's title" (the plan's own word for revision 1; the STEP-9 grep `graph tokens?\b|token line|a token` does not match it). Rewrap `PVLastUpdated`'s doc comment to the paragraph width.
Check: `rg -n -i 'pill' docs/deployment-plan/spike-12` prints nothing; the STEP-9 grep still prints nothing.

### Round 1 — fix (revision 2)
Head `01b15c5e`. Gate PASS (orchestrator run). F-1 – F-7 done in `15ea8fe8` (F-1 – F-6, shared files) and `01b15c5e` (F-7). Deviations: `badge` keeps `@MainActor` in the view extension (swift-warnings); F-7's `pill` grep still matches history in the plan, the archived brief and the logs, outside F-7's files (`deployment-plan.md` is clean); the plan's `statusBadgeLabel` mentions stay until the Landed note at merge. UX minors read by the orchestrator: UX-2 (compact value "title, type, ref"), UX-3 (`.help` on title and type), UX-4 / F-2 (`graphBlockLabel`) — Resolved. UX-1 / F-1 badge `accessibilityHidden` — Resolved.

### Round 2 — review (revision 2)
Reviewers: MNT (had Fix-now majors). Scope `62d34c8e...01b15c5e`. Not re-run: CON, SEC+PRF (clean), UX (minors only, read above).
MNT: MNT-1 and MNT-2 Resolved; no new blockers. Nit (logged): `MNT-8 · SourcesListRow.swift:25,30,102 · typeLabel computed three times`.
No Fix plan 2.

### Result (revision 2)
`origin/main` moved to `2243b362` (docs only); merged as `98c8e159`, gate PASS. CI pending · open: none · follow-ups: none · to user: UX-1 (answered 2026-10-27, fixed), CON-1 manual check · logged: CON-2, MNT-6, MNT-8 · CON miss: none
