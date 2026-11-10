# Ship log — S16-04 The Person page says when a reload failed or the diagram is partial

**Plan:** [`S16-04-person-page-notices.md`](S16-04-person-page-notices.md) · **Branch:** `s16-04-person-page-notices` · **PR:** #776
**State:** ready

Base: `main` @ `5a10ccb5`.

## Round 0 — implement
Head `9762e355`. Gate PASS (orchestrator run: all nine checks). Report deviations: version 0.7.1 (main was 0.7.0, plan said 0.6.1); `previewSession(claims:)` replaces two bools; `PersonSourcesPreviewData.notCounted(_:)`/`linesDropped(_:)` added for previews; `PVRetryCallout` puts `.id(failureRevision)` on itself; section host keeps one `VStack`; three more Go tests rewritten to D-9 (`TestPairsCounted` "AC-2.3 cap one short", crafted "unshared links", `TestPairsInvariants`); proto regenerated (comment-only). §10 manual checks and canvas previews not run.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none (UX: `macos/App/**`, `Localizable.xcstrings`; SEC: `core/recordlinks`; PRF: `core/database/persondetail`, `macos/App/Features/**` models).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-6 (MNT outside-lens) | CON | blocker | `PVRetryCallout`'s `onAppear` sits outside `.id(failureRevision)`, so a repeat failure never re-claims focus on Try again; `retryPressed` stays set | Fix now → F-1 | Verifier: Confirmed. Throwaway `NSHostingView` program: outer `onAppear` fires at rev 0 only; inside the id'd view it fires per bump. D-4 says "on the inner callout's appear". AC-1.3/AC-2.2 unmet for page, host, Evidence graph, Promote |
| MNT-1 | MNT | major | Try again wiring hand-copied at three Conclusions callers; band takes four values from one handle; `readAgain` rebuilds the key | Fix now → F-2 | Same-pass cleanup: all three copies are added by this PR; the skill line teaches the copy. A feature-layer `init` taking a `QueryHandle` leaves the kit's plain `failureRevision:` for recipes, so D-4's reason (a recipe cannot use a feature wrapper) still holds |
| CON-1 | CON | major | PR body cites unseen previews as evidence for AC-1.3 (focus/speech), 1.4, 1.5, 1.6, 2.1, 2.2 (drawn), 2.3, 3.2 | Fix now (orchestrator) | Verified: PR body says "seen in the previews"; Landed note says not looked at. Orchestrator rewrites the body at finish to list these as the user's §10 manual checks |
| UX-1 | UX | minor | With "left out", a node whose lines were all dropped is spoken "shares no one" | Fix now → F-3 | Verified `PersonRecordLinksContent.swift:81-87`: uncounted label only when `notCounted` |
| UX-2 | UX | minor | Several identical "Try again" buttons, nothing names the read | Fix now → F-2 | Kit file this PR changes; hint from the callout's title |
| MNT-2 | MNT | minor | `readPlan` keeps two past-limit shapes nobody uses | Fix now → F-4 | Verified: `Pairs` returns `nil, false` on `work > limit` |
| MNT-3 | MNT | minor | Two near-identical "answer once then fail" preview helpers; preview error ×4 | Fix now → F-5 | Same-pass cleanup, both added by this PR |
| MNT-4 | MNT | minor | Host asks "is this the column?" three ways; wrapping twice; `bandGap` comment names nothing | Fix now → F-6 | Verified `rg bandGap macos` hits only the comment |
| MNT-5 | MNT | minor | Kit README silent on `failureRevision` | Fix now → F-2 | README row at `DesignSystem/README.md:176` |
| MNT-6 | MNT | minor | `graphReadError` folds `QueryReading` by hand | Fix now → F-7 | Code this PR changes; next callers named in the plan |
| MNT-7 | MNT | minor | New tests copy `graph(_:)` and a 12-line setup | Fix now → F-8 | Same-pass cleanup in test files this PR changes |
| CON-2 | CON | minor | `failure` doc "Nil while a retry runs" wrong behind a value | Fix now → F-9 | Verified `beginLoading(keepingValue: true)` leaves `.error` |
| CON-3 | CON | minor | T-1 asserts `read` at one step, plan says every step | Fix now → F-8 | Plan T-1 wording |
| UX-3 | UX | minor | Notices failing together interrupt each other (high priority) | Follow-up | Kit-wide announcement policy; not this PR's alone |
| UX-4 | UX | minor | Focus goes nowhere after a successful Try again | Follow-up | Kit-wide; Promote and Evidence graph already behave so |
| CON-7 (CON outside-lens) | UX | minor | A mouse click that fails again also moves focus to Try again | Rejected | D-4: "the Button's action sets `retryPressed`", no input distinction; the focus lands on the button just clicked |
| CON-8 (CON outside-lens) | UX | minor | Event/Place: header-band notice gap is the columns' gap, not 20 pt | To user (manual check) | Needs the board and eyes; added to the PR's manual check list, does not block the fix round |

Nits (logged): `UX-5/MNT-10 · PersonRecordLinksColumn.swift:4-5 · rewrap doc comment` · `MNT-8 · PersonSourcesSection.swift:855 · two Bools → enum` · `MNT-9 · PersonRecordLinksContent.swift:762 · read partial == .notCounted` · `CON-4 · PersonRecordLinksContentTests · compare against L10n` · `CON-5 · recordlinks.go linesOf · rename skip param`

### Fix plan 1

**F-1 — Re-claim Try again on every remount** (from CON-6)
Files: `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift`.
Failing test first: none possible without a UI target (plan §4); if a hosted `NSHostingView` test can observe `claimFocusAfterPressedRetry` firing on a `failureRevision` bump without a test-only hook in the kit, add it; otherwise say so and rely on §10 check 3.
Do: move `.onAppear(perform: claimFocusAfterPressedRetry)` before `.id(failureRevision)` so it sits on the id'd view (D-4 "on the inner callout's appear"). `@State retryPressed` stays on `PVRetryCallout`. Fix the code comment to say the appear is the inner callout's.
Check: `rg -n "\.id\(failureRevision\)" -B2 macos/App/DesignSystem` shows the `onAppear` above it; gate PASS.

**F-2 — One handle-taking Try again** (from MNT-1, MNT-5, UX-2)
Files: new `macos/App/Features/Workspace/Session/QueryHandle+RetryCallout.swift` (or beside `QueryHandle`), `ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `macos/App/DesignSystem/README.md`, `.agents/skills/add-workspace-place/SKILL.md`, `docs/macos-client-patterns.md` if it repeats the pair. Skill: `add-ui-component`, `add-workspace-place`.
Do: add `extension PVRetryCallout { init<V>(title:message:handle: QueryHandle<V>, compact:id:onRetry:) }` that fills `retry: L10n.Conclusions.readNoticeTryAgain`, `isRetrying: handle.isReading`, `failureRevision: handle.failureRevision`, `retryIdentifier: "\(id).tryAgain"`. Move the three Conclusions callers onto it. `ConclusionDetailHeaderBand` takes the handle (or the built callout) in place of `refreshFailure`/`isRetrying`/`failureRevision`/`retry`; `ConclusionDetailLoaded.readAgain` uses the page's `key` passed down instead of rebuilding it. Evidence graph and Promote keep the kit init (their labels differ). In the kit, when `title` is set, add `.accessibilityHint(title)` on the Try again button (visible label unchanged). Add one clause to the README Callout row: pass `failureRevision` to speak again for each failure and keep the keyboard after a pressed Try again that fails. Update the add-workspace-place line to name the convenience instead of the hand-passed pair.
Check: `rg -n "readNoticeTryAgain" macos/App` prints only the convenience (and L10n); `rg -n "failureRevision:" macos/App/Features/Conclusions` prints nothing; gate PASS.

**F-3 — "Left out" never says "shares no one"** (from UX-1, MNT-9)
Files: `macos/App/Features/Conclusions/PersonRecordLinksContent.swift`, `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`.
Failing test first: `droppedLinesAreLeftOut` (or a new case) — a node whose only line is dropped is spoken with `L10n.Conclusions.recordLinksNodeUncountedAccessibility(…)`, not `recordLinksNodeApartAccessibility`.
Do: the uncounted label applies when `partial == .notCounted`, or when `partial == .leftOut` and the node's `connected == 0`. Replace the local `notCounted` with the `partial` reading.
Check: test passes; gate PASS.

**F-4 — `readPlan` has one past-limit shape** (from MNT-2, CON-5)
Files: `core/recordlinks/recordlinks.go` (+ tests if they assert the partial shape).
Do: `readPlan` returns `nil, n` whenever it passes `limit`; `Pairs` tests only `work > limit`; doc comment says "past limit: no pairs". Rename `linesOf`'s last parameter if still `skip`.
Check: `go test ./core/recordlinks/...` and the persondetail record links tests pass; gate PASS.

**F-5 — One "fail after the first answer" preview helper** (from MNT-3)
Files: `PersonDetailNoticePreviews.swift`, `PersonSourcesSection.swift`, `PersonSourcesPreviewData.swift` (or `FakeStore.swift` beside `sourceNotFound`).
Do: one `previewFailure` error value and one `#if DEBUG` `failAfterFirstAnswer<V>(_ key:, session:, fail: (FakeStore) -> Void)`; both previews call it.
Check: `rg -n "CoreInvokeError.coded\(" macos/App --glob '!FakeStore.swift'` prints at most the one shared definition; gate PASS.

**F-6 — The host asks "column?" one way** (from MNT-4)
Files: `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift`.
Do: use `isColumn` in `idRoot` and the empty-state check; one `@ViewBuilder func placed(_:)` applies `.pvKeepsPageBand()` in the band; `retryNotice` and `goneNotice` use it; drop the `bandGap` sentence.
Check: `rg -n "columnLoadingHeight == nil|bandGap" macos/App/Features/Conclusions` prints nothing; gate PASS.

**F-7 — `QueryReading` names its notice error** (from MNT-6)
Files: `QueryHandle.swift`, `EvidenceGraphModel.swift`, and any Conclusions caller that does the same fold.
Do: add `QueryReading.noticeError: Error?` (`.failed(e)` → e, `.loaded(_, f)` → f, `.loading` → nil); `graphReadError` maps it.
Check: gate PASS.

**F-8 — Tests share their setup and T-1 checks every step** (from MNT-7, CON-3, CON-4)
Files: `PersonRecordLinksContentTests.swift`, `WorkspaceSessionTests.swift`.
Do: drop the local `graph(_:)` in `allLinesDrawAndNodesStopAtTheCap`; extract a private `sourcesWithAFailedReload()` used by T-5 and T-6; in `queryReadingCarriesARefreshFailure` assert `handle.read`'s case and value after every step; compare spoken labels against the `L10n` functions.
Check: tests pass; gate PASS.

**F-9 — `failure`'s doc says when it is nil** (from CON-2)
Files: `QueryHandle.swift:58`.
Do: "Nil while a retry after a first-load failure runs (a retry behind a value keeps it); a notice with Try again reads `reading`."
Check: gate PASS.

## Round 1 — fix
Head `ad5830ad`. Gate PASS (orchestrator run). F-1 – F-9 done. F-1 has no failing test (no UI target; relies on §10 check 3). F-3's test `aNodeWhoseLinesWereAllDroppedIsNotSpokenAsSharingNoOne` failed first ("c, c, , shares no one"). F-4: `linesOf`'s parameter was already `capped` (CON-5 moot). F-2: callers still add `.accessibilityIdentifier(id)` themselves. Docs touched: plan Landed note, `completed.md`.

UX minors (no round 2 for UX): UX-1 **Resolved** — `PersonRecordLinksContent.swift` uses the uncounted label when `reading == .leftOut && connected == 0`. UX-2 **Resolved** — `PVRetryCallout.retryButton` adds `.accessibilityHint(title?.pvString ?? "")`.

## Round 2 — review
Reviewers: CON (CON-6 blocker, CON-1 major), MNT (MNT-1 major). Scope `9762e355...ad5830ad`. Not sent: UX, SEC+PRF (no Fix-now blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-6 | CON | blocker | (round 1) focus claim outside `.id` | Resolved | `PVRetryCallout.swift:52-53`: `onAppear` before `.id(failureRevision)` |
| MNT-1 | MNT | major | (round 1) Try again wiring ×3 | Resolved | `QueryHandle+RetryCallout.swift`; three callers use it; README and skill name it |
| CON-1 | CON | major | (round 1) PR body cites unseen previews | Resolved (orchestrator) | PR body rewritten: those ACs listed as "Not yet verified — the user's manual checks" |

No new findings. Logged (CON outside-lens): `PVRetryCallout.swift:61` sets an empty hint when `title` is nil — harmless.

## Finish
Merged `origin/main` @ `cbe6b57d` (docs-only: S15-32 plan). Gate PASS after merge.

## Result
Gate PASS · CI pending · open: none · follow-ups: UX-3, UX-4 (home: user to choose) · to user: CON-8 (Event/Place notice gap, in the manual check)
