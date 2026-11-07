# Ship log — S13-17 Remove a claim

**Plan:** [`S13-17-remove-a-claim.md`](S13-17-remove-a-claim.md) · **Branch:** `s13-17-remove-a-claim` · **PR:** #729
**State:** ready

## Round 0 — implement
Head `69e06cdf`. Gate PASS (orchestrator run). Report: `UnpromoteFlowTests` and T-7 seed through `FakeStore` and let the session load; `unpromoteLastRecordNamesTheDeletedHandle` lists the Subject's own handle id; `PersonSourcesList` keeps `unpromotingClaimID` (D-10 did not say where the removed claim's id comes from); the Source page keeps its `model.sync` `.onChange`; the strings landed in the STEP-2/3 commit. The manual check was not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC+PRF runs for `macos/App/Features/**` models and session code.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC+PRF | major | The connection list has no scroll or height limit | Fix now → F-1 | verified: `PVUnpromoteConfirmDetail` draws a plain `VStack`; `PVDeleteImpactGroupStack` caps at `listMaxHeight` |
| UX-1 | UX | major | A failed write is not announced | Fix now → F-2 | verified: `announcesError` is `impactCheckFailed`; `confirm` never clears `error` |
| UX-2 | UX | major | Keep promotion strands keyboard focus | Fix now → F-3 | verified: nothing restores focus when `pending` closes with no write |
| UX-3 | UX | major | A gone handle leaves silently | Fix now → F-4 | To user, answered 2026-11-06: show a Toast with the existing message |
| UX-4 | UX | major | Deleting a person is silent for VoiceOver | Fix now → F-5 | verified; spoken only, so RC-6's no-Toast answer stands |
| UX-5 | UX | minor | Same-named records are not told apart | Fix now → F-5 (label) | Confirm title kept: to user, answered 2026-11-06 |
| MNT-1 | MNT | major | Two wire codes stay raw strings | Fix now → F-6 | same-pass cleanup; one enum, one parity row |
| MNT-2 | MNT | major | No guidance names the leave modifier | Fix now → F-7 | verified: `rg leavesForSectionRoot .agents .cursor` finds nothing |
| MNT-3 | MNT | major | `unpromotingClaimID` copies `pending.id` | Fix now → F-3 | verified: `identity_claims_one_accepted_per_subject` |
| MNT-4 | MNT | major | A global fixture with one consumer | Fix now → F-8 | verified: only `EvidenceGraphModelTests` calls `graphRows` |
| MNT-5 | MNT | minor | `confirm` delivers the effects twice | Fix now → F-9 | minor in a file this PR adds; supersedes D-1's return type |
| MNT-6 | MNT | minor | Gone sits outside the presentation enum | Fix now → F-4 | minor in a file this PR changes |
| MNT-7 | MNT | minor | False comments and a forward-only wrapper | Fix now → F-8 | |
| MNT-8 | MNT | minor | `bridges` and `bridgesUnlisted` admit a meaningless state | Fix now → F-1 | |
| MNT-9 | MNT | minor | Recipe copy in `L10n.EvidenceGraph` | Fix now → F-10 | `L10n.DeleteImpact` is the precedent |
| PRF-1 | SEC+PRF | minor | Leaving before `apply` can load the Persons list twice | Fix now → F-9 | code this PR adds; supersedes D-7's order inside the session |
| CON-1 | CON | major | No evidence for the in-app paths | Rejected | The manual check is the user's step (Finish §4); F-9 makes T-5 run the real leave path, and the report names §10 as unverified |
| UX-6, UX-8 | UX | nit | Return while disabled; a second confirm | Fix now → F-3 | same lines as F-3 |
| MNT out of lens | PRF | — | `evidenceGraphSnapshot` rebuilds on each open | Rejected | One build per click that names bridges, the build the canvas runs on every rows change; no loop or per-line cost |

Nits (logged): `CON-N1 · deployment-plan.md:489 · S13-08 row names confirmUnpromote` · `MNT-10 · EvidenceGraphView.swift:206 · view assembles the confirm call` · `MNT-11 · CoreInvoke.swift:31 · say "code-only read"` · `MNT-12 · PersonSourcesSection.swift:86 · PersonSourcesFocus can be private` · `MNT-13 · UnpromoteFlow.swift:43 · throwaway CatalogDeleteImpact` · `MNT-14 · three .missingDeepId tails stay inline` · `UX-7 · WorkspaceNavigation.swift:159 · Back steps through list entries, accepted by the plan's Out line`

### Fix plan 1

**F-1 — The connection list scrolls, and one prop says listed or unlisted** (from SEC-1, MNT-8)
Files: `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift`. Skill: `add-ui-component`.
Failing test first: `UnpromoteFlowTests/listComesFromThePreflight` asserts the item's `connections` value.
Do: `PVUnpromoteConfirmContent` replaces `bridges` and `bridgesUnlisted` with `connections: PVUnpromoteConnections` *(new)*, `enum { case listed([String]), unlisted }`; `.listed([])` draws nothing. The sentences sit in a `ScrollView` over a `LazyVStack`, `.frame(maxHeight: PVDeleteImpactLayout.listMaxHeight)` as `PVDeleteImpactGroupStack` does; the header and note stay outside it. Add a `#Preview` with 200 sentences. No row cap: the sheet shows every connection.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — A failed write is always heard** (from UX-1)
Files: `PVUnpromoteConfirm.swift`, `UnpromoteFlow.swift`, `UnpromoteFlowTests.swift`.
Failing test first: `confirmFailureStaysOpen` asserts a retry clears `pending.error` before the write, using `unpromoteSubjectHold`.
Do: drop `announcesError`; pass `announces: true`, as `SourcePageView.swift:99` does. `confirm` sets `pending?.error = nil` before the write, as `DeleteImpactFlow.confirm` does.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — Focus survives Keep promotion, and the list stops copying `pending.id`** (from UX-2, MNT-3, UX-6, UX-8)
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`.
Failing test first: `lineAfterRemoving` takes a Subject id.
Do: `PersonSourcesContent.lineAfterRemoving(subjectID:) -> String?` (claim id of the target); `confirmAndTrackFocus` starts with `guard !unpromote.isRunning, let subjectID = unpromote.pending?.id` and stores `pendingFocus: (subjectID, target)`; `applyPendingFocus` tests lines by `subjectID`. Delete `unpromotingClaimID` and the wrapper closure. Add `@State returnFocus: PersonSourcesFocus?`, set in the line's begin only when `focus == .unpromote(claimID: line.id)`, so a mouse click draws no ring. `.onChange(of: unpromote.pending?.id)`: when it becomes nil with no `pendingFocus`, `FocusClaim.claim($focus, returnFocus)` if that line is still drawn, then clear `returnFocus`. `pvActivatesOnReturn(when:)` adds `&& !unpromoteDisabled`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — A gone handle is a presentation state and leaves with a Toast** (from UX-3, MNT-6)
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `ConclusionPageSectionHost.swift`, `macos/App/Features/Workspace/Session/QueryHandle.swift` if `isConclusionNotFound` lived near it, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Failing test first: `presentationFollowsTheHandle` expects `.gone` for `conclusiondetails.not_found`, `.failed` for any other code.
Do: `ConclusionDetailPresentation` gains `case gone`, mapped in its `init`; `QueryRead.isConclusionNotFound` goes, and T-9 moves into that test. `ConclusionDetailLoaded` draws `loadingPage` for `.gone`, and on the transition to `.gone` sets `session.noticeToast` to `L10n.Errors.conclusionDetailsNotFound` (the shape at `PromoteModel.swift:379`), then `leavesForSectionRoot(when: presentation == .gone)`. `ConclusionPageSectionHost` draws its spinner for `.gone`. The Toast covers Person, Event and Place. Supersedes the plan's §11 merged-handle row; AC-2.6's "no Callout" still holds.
Check: the test passes; scripts/check-pr.sh PASS.

**F-5 — Spoken results and a label that says which record** (from UX-4, UX-5)
Files: `PersonSourcesSection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Failing test first: none (view wiring); the label is pure.
Do: when the write deletes the person, `AccessibilityAnnouncement.post(L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:))` *(new key)* "%@ deleted.", comment "VoiceOver announcement after a person's last record is unpromoted; the argument is the person's ref", with the ref from `pending.deletedHandles`' listed row, falling back to no announcement when none is listed. `sourcesUnpromoteAccessibility(graph:subject:ref:)` becomes "Unpromote %1$@, record %2$@, %3$@", the graph link's shape, set on `Line` beside `accessibilityLabel` as a new `unpromoteAccessibilityLabel`.
Check: scripts/check-pr.sh PASS.

**F-6 — One typed enum for the engine error codes the client branches on** (from MNT-1)
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/CoreInvoke.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Sources/SourceMetadataSection.swift`, `ConclusionDetailPage.swift`.
Failing test first: the Go row "engine error codes" wants `apperr.CodeConclusionDetailsNotFound`, `apperr.CodePromoteStale`, `apperr.CodeSourceMetadataInvalid`.
Do: `enum EngineErrorCode: Sendable, Equatable { case conclusionNotFound, promoteStale, sourceMetadataInvalid; init?(wire:) }` replaces `ConclusionDetailsError`; `CoreInvokeError.engineCode: EngineErrorCode?`. The three sites match on it.
Check: the test passes; `rg '== "[a-z_]+\.[a-z_]+"' macos/App` finds nothing; scripts/check-pr.sh PASS.

**F-7 — Guidance names the leave** (from MNT-2)
Files: `.agents/skills/add-workspace-location/SKILL.md`, `.agents/skills/add-workspace-place/SKILL.md`, `.cursor/rules/workspace-navigation.mdc`.
Do: a page whose read reports its entity gone applies `.leavesForSectionRoot(when:)`; a write that deletes the shown entity applies its effects with `session.apply(_:leavingPage:)`.
Check: `python3 scripts/check-guidance-names.py` passes.

**F-8 — Remove the fixture file and the forward-only wrapper** (from MNT-4, MNT-7)
Files: `macos/ProvenenciaTests/EvidenceGraphRowsFixture.swift` (delete), `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`.
Do: `graphRows` returns to a private method of `EvidenceGraphModelTests`, without its `types` parameter or arguments. `PromoteModel` calls `session.evidenceGraphSnapshot(graphID: entry.graphID)` at both sites and `graphSnapshot()` goes. The session comment reads "Nil when the rows could not be read."
Check: scripts/check-pr.sh PASS.

**F-9 — Leaving a page applies the effects once, and `confirm` returns a Bool** (from PRF-1, MNT-5)
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `UnpromoteFlow.swift`, `PersonSourcesSection.swift`, `EvidenceGraphView.swift`, `UnpromoteFlowTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `CatalogMutation.swift` and `docs/macos-client-patterns.md` (their sentence).
Failing test first: `WorkspaceSessionTests/applyLeavingPageLoadsTheNextListOnce`: a stale `.personsList`, then `apply(effects, leavingPage:)` that moves to the Persons list → `listPersons` runs once and the person's `.conclusionDetail` key is evicted.
Do: `WorkspaceSession.apply(_ effects:, leavingPage leave: () -> Void)` *(new)*: `visibleKeys = []`, `apply(effects)`, `leave()`. `UnpromoteFlow.confirm(session:userID:apply: (@MainActor (CatalogWriteEffects) -> Void)? = nil) async -> Bool`, `@discardableResult`, as `DeleteImpactFlow.confirm`: it calls `apply` when given, else `session.apply(effects)`. The section passes `{ effects in if effects.deletedHandles.contains(entityId) { session.apply(effects, leavingPage: navigation.fallbackToSectionRoot) } else { session.apply(effects) } }`. `beforeApply` goes. T-5 captures the effects through `apply`. Supersedes D-1's `beforeApply` and D-7's order; note both in the plan's Landed note.
Check: the test passes; scripts/check-pr.sh PASS.

**F-10 — The Unpromote copy gets its own namespace** (from MNT-9)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, the recipe, `UnpromoteFlow.swift`, `PersonSourcesSection.swift`, the card's callers, tests. Skill: `add-localized-string`.
Do: every `L10n.EvidenceGraph.unpromote*` member moves to `L10n.Unpromote`, with catalog keys `unpromote.*`. Copy and comments are unchanged.
Check: `rg 'L10n\.EvidenceGraph\.unpromote|evidenceGraph\.subject\.unpromote' macos` finds nothing; `check-localizable-xcstrings.py` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `66138766`. Gate PASS (orchestrator run). F-1 – F-10 done. `L10n.Unpromote` members take shorter names (`button`, `keep`, `confirmTitle`, …), copy unchanged. Landed S13-23's plan still cites the old keys; it is history, left alone. Manual check not run.

## Round 2 — review
Reviewers: MNT, UX, SEC+PRF (each had a Fix-now major). CON skipped: CON-1 was Rejected, so it has no fix to re-review. Scope `69e06cdf...66138766`.

Resolved: SEC-1, UX-1 – UX-4, MNT-1 – MNT-4 (reviewers). Read by the orchestrator: UX-5, MNT-5 – MNT-9, PRF-1, UX-6, UX-8 Resolved.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-15 | MNT | blocker | Two ways to leave a page whose entity a write deleted | Fix now → F-11 | verified: `SourcePageModel` applies with `.sourceWorkspace` visible, then the view leaves; the guidance states the new order as the rule |
| MNT out of lens (UX) | UX | blocker | `ConclusionPageSectionHost` spins forever on `.gone` | Fix now → F-12 | re-rated: F-4 added a state with no exit where a Callout stood; the page may hold its value while a section read fails |
| PRF note | PRF | minor | `presentation` builds the page's content three times per `body` | Fix now → F-12 | rides F-12, same lines this PR added |

### Fix plan 2

**F-11 — One way to leave a deleted entity's page** (from MNT-15)
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` if the signature needs it, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`.
Failing test first: `SourcePageModelTests/resourceDeleteDoesNotReloadTheDeletedSource`: confirming a Source delete with its page visible → `getSourceWorkspace` is not called for that id again, and `leave` ran once.
Do: `apply(_ effects:, _ mutation: CatalogMutation? = nil, leavingPage leave: () -> Void)`. `SourcePageModel.confirmResourceDelete(leaving:)` takes the leave closure and passes it to that call; on a failed write it does not run. `SourcePageView` passes `navigation.fallbackToSectionRoot` and stops calling it after the `await`. Master–detail pages (`MetadataView`, `SourceTypesView`, `PropertiesView`) stay as they are: their list stays on screen. Say that scope in `add-workspace-location`, `add-workspace-place` and `workspace-navigation.mdc`: the rule is for a page that the deleted entity is the whole of.
Check: the test passes; scripts/check-pr.sh PASS.

**F-12 — A section with a gone handle has an exit, and `presentation` is read once** (from the two notes above)
Files: `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`.
Do: the host's `.gone` branch draws the `.failed` branch's `PVCallout` and id with `L10n.Errors.conclusionDetailsNotFound`, so a section whose read fails under a page holding a value shows what it showed before F-4. `ConclusionDetailLoaded.body` reads `presentation` once into a `let`, as `ConclusionPageSectionHost.loaded` does, and every use in `body` takes it.
Check: scripts/check-pr.sh PASS.

## Round 2 — fix
Head `88ba5bb3`, `main` merged at `285c81ff`. Gate PASS (orchestrator run on the merge). F-11, F-12 done.

## Result
Gate PASS · CI green (Go tests, macOS tests) · open: none · follow-ups: none · to user: answered (UX-3 Toast, confirm title, recipe split) · CON miss: none
