# S14-20 — The Promote page recovers and speaks

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes (batch 5)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D7](../design/S14-D7-promote-notices.md), board agreed 2026-11-03
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), file *Promote flow.dc.html*, section "S14-D7", frames 1a, 1b, 1c, 2a, 2aW, 3 and 4. PN-1: option (a), the list scrolls to its top; option (b) is deleted from the board. Try again is drawn as a small (`size: .sm`) secondary Button. The board's second load message ("Provenencia still couldn't read this Evidence graph…") is a placeholder for whatever the second failure says: no new copy
**Depends on:** S14-14 (landed [#675](https://github.com/mendahu/provenencia/pull/675)), S13-24 (landed [#639](https://github.com/mendahu/provenencia/pull/639)), S14-22 (landed [#700](https://github.com/mendahu/provenencia/pull/700)), S14-D7 (agreed 2026-11-03)
**Blocks:** —
**Branch:** `s14-20-promote-recovers-and-speaks`
**Version bump:** patch, 0.2.12 → 0.2.13 — responsiveness and accessibility fixes to an existing page, no feature ([`docs/versioning.md`](../../../versioning.md)). If `main` is past 0.2.12 when the PR merges, take the next patch.
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (STEP-1 only), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** audit 2 B-12, B-16, B-17, B-37 ([`../app-health-audit-2.md`](../app-health-audit-2.md)); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3, §3.1, §6; working rules "Kit components announce themselves", "FakeStore is a seeded fake, not a second engine", "Audit items keep their IDs"
**Planned:** 2026-11-03 against `main` @ `b2166b4` (refreshed after S14-22 landed; first drafted @ `0e20544`)
**Review:** [`S14-20-promote-recovers-and-speaks.review.md`](S14-20-promote-recovers-and-speaks.review.md) — Fits with changes, Ready with changes; 2 rounds

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Quick row decisions on the Promote page each start a full proposal, so five picks hold the catalog session for five proposals in a row, and every other catalog call waits behind them (B-12). After this PR, the page runs at most one proposal at a time. Decisions made while one runs are sent together, as one proposal, when it returns. On a long mapped page, Done's failures appear at the top of the scroll area, out of sight, and VoiceOver hears none of them. A failed load has no way to try again (B-17). After this PR, each error notice comes into view and is announced when it appears, and a failed load offers Try again. The stale notice no longer claims "Provenencia proposed the rows again" after that proposal failed. The always-1 `graphalign.Subject.Provenance` multiplier goes (B-37's last item), with scores unchanged.

**Why now.** Batch 5 closes the Spike 12 audit, and the kit's callout announcement (S14-14) and the pin engine's removal (S13-24) have both landed.

**What stays.** What a proposal proposes, every score, Done's write and its revision check, and the rows, notes, subrows, header and footer as S13-D11 agreed them. A decided row whose record or Subject was deleted or merged in another window keeps today's `promote.invalid` error: recovering from it is descoped (user, 2026-11-03). The "start as Skip" header line (B-16) is already gone on `main`: `L10n.Promote.openedFrom` reads "Opened from %@", and this PR changes nothing there.

## 2. User stories

**US-1.** As a researcher, when I make several row decisions quickly, the catalog runs one proposal at a time and the page settles on the proposal for my last decision, so other work in the app is not stuck behind a queue of proposals.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A loaded page; five `setTarget` calls while the first proposal is still running (held on `proposePromoteGraphAlignmentHold`) | The store records exactly 2 proposal calls after load: the first, and one carrying all five decisions |
| AC-1.2 | AC-1.1's second proposal | Its `fixed` holds the five decided rows' targets as they were after the fifth decision |
| AC-1.3 | The first proposal of AC-1.1 returns while the second is pending | The first proposal's rows are not merged: the page merges only the proposal that answers the latest decisions |
| AC-1.4 | A proposal is running or pending | `canFinish` is false; it is true again once the last proposal settles and every row has a target |
| AC-1.5 | One decision, nothing else pending | One proposal call, merged as today (`retargetKeepsTheDecidedRowAndMarksTheSuggestionUpdated` stays green) |

**US-2.** As a researcher, when Promote fails, I see and hear the failure where I am, and I can retry a failed load.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The first proposal fails (store error) | `loadError` is set and the load callout shows **Try again** (PN-2) |
| AC-2.2 | AC-2.1, then Try again with the store now succeeding | `isLoading` is true while the call runs and Try again shows its loading state; afterwards `loadError` is nil and the rows show |
| AC-2.3 | AC-2.1, then Try again with the store failing again with the same error | `loadError` is set again; the callout is remounted (`loadAttempt` bumped), so VoiceOver hears it again |
| AC-2.4 | A mapped page scrolled to the bottom; Done fails stale | The list scrolls to its top, the stale notice is in view above the first section (frame 2a), and it announces |
| AC-2.5 | Done fails with a non-stale error | The list scrolls to its top, the save error is in view (frame 3), and it announces |
| AC-2.6 | A row decision's proposal fails | The list scrolls to its top, the proposal error is in view (frame 4), and it announces |
| AC-2.7 | Done fails stale, and the re-proposal fails | `flow.staleNote` is nil and `proposeError` is set (PN-3) |
| AC-2.8 | Done fails stale, and the re-proposal succeeds | `flow.staleNote` stays set until the next decision or Done, as today (`aStaleDoneReproposes` stays green) |

**US-3.** As a contributor, graph alignment has no scoring input that is always 1.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg -n 'Subject\.Provenance|Provenance:|provenance float64' core/graphalign core/database/promotealign` | No hit |
| AC-3.2 | The `core/graphalign` and `core/database/promotealign` suites | Pass with no expected value changed |

## 3. Scope

| In | Covered by |
| --- | --- |
| One proposal at a time, the latest decisions winning, private to `PromoteModel`; `proposeGeneration` and `nextProposal` go (B-12) | STEP-3, AC-1.1 – AC-1.5 |
| A failed proposal clears the stale notice (B-13's stale half, PN-3) | STEP-3, AC-2.7, AC-2.8 |
| Try again on a failed load, reading a failed graph again; the load message no longer says to go back (B-17, PN-2) | STEP-4, AC-2.1 – AC-2.3 |
| Stale, save, proposal and load errors announce, and a new stale, save or proposal error scrolls the list to its top (B-17, PN-1 option (a), PN-W1) | STEP-5, AC-2.4 – AC-2.6 |
| One kit retry callout, `PVRetryCallout`, used by Promote, the Evidence graph canvas and the Source forms (review M-10, user 2026-11-03) | STEP-4, T-14 |
| FakeStore error hooks for propose and apply, a `FakeHold` for propose; `promoteProposeDelays` goes | STEP-2 |
| `graphalign.Subject.Provenance` and `ScoreCandidate`'s `provenance` parameter go (B-37) | STEP-1, AC-3.1, AC-3.2 |

| Out | Home |
| --- | --- |
| A decided row whose record or Subject went away in another window: Go releasing the pair, a released-reason wire vocabulary, row notes for it (B-13's main fix) | Descoped by the user, 2026-11-03: a far edge case; today's `promote.invalid` error stays. Noted on B-13 in [`../app-health-audit-2.md`](../app-health-audit-2.md) |
| A `promote.invalid` at Done proposing again once | Descoped with B-13 (user, 2026-11-03). Apply already checks the revision before it files a row (§4), so an invalid row after a passing revision check is not a race |
| The "start as Skip" header line (B-16) | Already fixed on `main` (§4); nothing to do |
| Announcing rows marked Updated after a proposal (named in B-17's Problem) | Not planned: no scope entry or working rule asks for it |
| A pending sign on the connection subrows and the footer's connection count while a proposal runs (S14-22 ship log UX-1) | Dropped from S14-20 by the user, 2026-11-03: not drawn on S14-D7. Not planned |
| The sidebar recount (`CatalogCounts.refreshAll`) on the same one-at-a-time queue | Declined by the user, 2026-11-03. Not planned |
| A repeat key on `pvAnnounce` so a retry callout re-announces without a remount | Not planned: one call site (D-8) |
| The proposal's own cost (B-11, B-8) | [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Promote Done's effects | S14-16 (landed) |
| B-37's wire half (`score`, `via_bridge_type`, `via_role`) | Reserved by S13-24 (landed) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Proposal after a decision | `macos/App/Features/Promote/PromoteModel.swift` `proposeAfterDecision`, `nextProposal`, `repropose(generation:)`, `proposeGeneration`, `isProposing` | Each decision (`setTarget`, `hold`) and `mapRest()` (since S14-22) calls `proposeAfterDecision`, which clears `flow.staleNote`, bumps `proposeGeneration`, and starts a `Task` that calls `store.proposePromoteGraphAlignment` with `fixed: flow.fixed(for: flow.rows.filter(\.decided))` and `onlySubjectID: flow.onlySubjectID` (nil once the rest is mapped). Every call runs; only the latest generation's answer is merged. Five picks are five FFI calls, serialized by `catalogsession.Do` in Go (B-12) |
| Stale Done | `PromoteModel.promote()` | On `promote.stale`: sets `flow.staleNote`, then `await repropose(generation: nextProposal())`. When that proposal fails, `proposeError` is set and `staleNote` stays, so both callouts show and the stale one says the rows were proposed again |
| Load | `PromoteModel.load()` | One proposal with `fixed: []`, then the graph snapshot; a failure sets `loadError`. Nothing retries; `canFinish` is false while `loadError != nil` |
| Notices | `macos/App/Features/Promote/PromoteView.swift` `notices` | Four `PVCallout(tone: .danger, …)` (load, stale, proposal, save) at the top of the `ScrollView`, none with `announces:`; Done is in `footer`, outside the scroll area. `rg AccessibilityAnnouncement macos/App/Features/Promote` has no hit |
| Kit announcement | `macos/App/DesignSystem/Components/Callout/PVCallout.swift` `announces`, `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` `PVAnnounceModifier` | `announces: true` speaks title and message on appear and on a change to a different text. The same text again while mounted is not spoken |
| Retry callouts today | `macos/App/Features/Sources/SourceFormLoadArea.swift` `.failed` (callers `SourceCopyDialog`, `SourceGraphFormDialog`, `SourcePageArtifactsView` twice); `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `readErrorCallout(_:)` | Each builds a danger `PVCallout` with a secondary `.sm` `PVButton` ("Read again") by hand. The Source forms show no busy state; the canvas disables the button and shows a `ProgressView` beside it (`evidenceGraph.readError.reading`). Neither announces |
| Callout actions | `PVCallout.init(…, @ViewBuilder actions:)` | An actions row under the message; `SourcePageArtifactsView.swift` and `EvidenceGraphView.swift` pass one |
| Button loading | `macos/App/DesignSystem/Components/Button/PVButton.swift` `loading:` | Shows a spinner and blocks the action |
| Header line (B-16) | `macos/App/Platform/L10n.swift` `L10n.Promote.openedFrom` | "Opened from %@". No "start as Skip" anywhere: `rg -n "start as Skip" macos docs/*.md .agents` has no hit |
| Apply order | `core/database/promote/batch.go` `SaveBatch` | Checks the graph, then `latestRevision` against `SeenRevision` (`ErrStale`), then files rows. `ResolvePins` is gone (S13-24) |
| Fixed pairs | `core/database/promotealign/promotealign.go` `checkFixed` | Unchanged by this PR: a pair naming a Subject not on the graph, or a deleted, merged or wrong-kind handle, is `promote.ErrInvalid` |
| Provenance | `core/graphalign/types.go` `Subject.Provenance`; `core/graphalign/score.go` `ScoreCandidate(r, edge, provenance float64)`; `core/graphalign/align.go` `(*state).score` passes `s.Provenance`; `core/graphalign/rows.go` line 150 passes `1`; `core/database/promotealign/layer.go` line 134 sets `Provenance: 1` | Always 1, so `Score = (r.Evidence + edge) * 1` |
| Provenance tests | `core/graphalign/align_test.go` lines 778, 791, 805, 840 | Call `ScoreCandidate(…, 1)` |
| Provenance docs | `docs/promote-graph-alignment.md` §6 line 137, §12 line 242 | "Graph alignment's `Subject.Provenance` is 1 today (`core/database/promotealign/layer.go`)." and "Every Subject's provenance is 1; …" |
| Load retry | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `readyValue`, `reread`; `EvidenceGraphModel.readGraphAgain()` | `readyValue` waits only for loads in flight and returns nil for a failed key; `reread` invalidates and queries it again |
| FakeStore | `macos/App/Platform/FakeStore.swift` `proposePromoteGraphAlignment`, `promoteProposals`, `promoteProposeDelays`, `applyPromoteGraphAlignment` | Scripted proposals and delays (`promoteProposeDelays`, used only by `anOlderProposalAnsweringLastIsDropped` and `doneWaitsForAPendingProposal`); no error hook or hold for either call; apply throws only `promote.stale`. `FakeHold` exists (`workspaceNavCountsHold` and others) |
| Tests | `macos/ProvenenciaTests/PromoteModelTests.swift` `anOlderProposalAnsweringLastIsDropped`, `doneWaitsForAPendingProposal`, `aStaleDoneReproposes`, `aPageWithoutItsGraphSaysSoInsteadOfLookingEmpty`, and S14-22's `theProposalCarriesThePagesScope`, `mappingTheRestProposesAgain`; `macos/ProvenenciaTests/QueryHandleWaiting.swift` `waitUntil` | Cover generation drop, Done waiting, stale re-proposal, the graph-missing load error, and the scope a proposal carries |

**Precedent.** `CatalogSearchModel.scheduleSearch` (`macos/App/Features/Catalog/Search/CatalogSearchModel.swift`) is the nearest latest-wins shape, using a generation counter and a `Task`. Do not copy its debounce, its cancel or its overlapping calls: a proposal is not debounced, and an FFI call in flight cannot be cancelled, so the queue waits for it. `EvidenceGraphModel.readGraphAgain()` (`session.reread(graphKey)`) is the precedent for reading a failed key again. S14-14 ([#675](https://github.com/mendahu/provenencia/pull/675)) is the precedent for `announces: true` on a danger callout (`SourcePageView.swift` line 99). `FakeHold` (`FakeStore.swift`, used as `workspaceNavCountsHold` since S14-16) is the precedent for pausing a fake call in a test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The proposal queue is private to `PromoteModel`: `private var pendingProposal: PromoteProposalInput?` *(new)*, with `isProposing` as the loop's one running flag (D-3). `submitProposal()` *(new)* stores the input in `pendingProposal` and, when `!isProposing`, sets `isProposing = true` and starts the loop. The loop is one `Task` that takes `pendingProposal`, sets it to nil, awaits the call, merges the answer only when `pendingProposal` is still nil, and repeats until nothing is pending. Each iteration takes a strong reference to the model from `[weak self]` for that iteration only, so after the page goes the loop stops once the call in flight returns | One researcher clicking fast must not queue N proposals (B-12). One caller, so no shared type (user, 2026-11-03: no extra machinery for the Promote flow) | A generic `LatestRunner` in `Platform/` — one caller; moving `CatalogCounts.refreshAll` onto it — declined by the user, 2026-11-03; keeping the generation counter — it is what lets five calls run |
| D-2 | `PromoteProposalInput` *(new, a private struct in `PromoteModel.swift`)* holds `fixed` and `onlySubjectID`, both taken in `submitProposal()` from `flow.fixed(for: flow.rows.filter(\.decided))` and `flow.onlySubjectID` | The last submit carries every decision and the page's scope as they are now, so dropping earlier pending inputs loses nothing | Reading `flow` when the pending call starts — the test could not see what was sent |
| D-3 | `isProposing` stays a stored `private(set) var` and is the only running flag: `submitProposal()` sets it true when it starts the loop; the loop sets it false when it ends with `pendingProposal` nil. `canFinish` keeps reading it | Done waits for the decision that is pending, not only the one in flight | — |
| D-4 | The loop's call body is `repropose`'s today, without the generation: call `store.proposePromoteGraphAlignment`, read `graphSnapshot()`, then merge, or set `graphUnavailableRepropose`, or set `proposeError`. A result or error that arrives while `pendingProposal` is set is dropped | The merge rules stay as they are | — |
| D-5 | A proposal failure the page shows (store error or missing snapshot) also sets `flow.staleNote = nil` | PN-3, frame 4: the stale note claims the rows were proposed again, which is false once that proposal fails | — |
| D-6 | `proposeAfterDecision()` clears `flow.staleNote` and calls `submitProposal()`. The stale path in `promote()` sets `flow.staleNote` and calls `submitProposal()` directly, without awaiting | One path for every proposal; the stale path keeps its note | — |
| D-7 | `load()` stays outside the queue. `retryLoad()` *(new)* returns at once when `isLoading` *(new)*; else it sets `isLoading = true` synchronously, always calls `session.reread(graphKey)` (as `EvidenceGraphModel.readGraphAgain()` does: `readyValue` returns nil for a failed key and for a stale one that still holds a value), and starts `load()` in a `Task`. `load()` sets `isLoading = false` when it ends, and bumps `loadAttempt` *(new)* on each failure | A failed graph read stays failed in the session until it is read again, so retrying only the proposal cannot recover `graphUnavailable`. Bumping on failure, not on start, announces the new failure, not the old one | Running load through the queue — no decision exists before rows do; a load-state enum — three members on one page do not earn one |
| D-8 | The load callout carries `.id(model.loadAttempt)`, so a repeated identical failure remounts and is announced again | `PVAnnounceModifier` does not repeat an unchanged text while mounted (§4) | A repeat key on `pvAnnounce` in the kit — one call site; not this PR |
| D-9 | All four notices pass `announces: true` | Working rule "Kit components announce themselves"; danger tone already gets high priority in `PVCallout` | Posting `AccessibilityAnnouncement` from the model — the rule forbids call-site posts |
| D-10 | PN-1 is option (a), as the board agreed. The `ScrollView` takes `.pvScrollsRowIntoView(PromoteView.noticesAnchor, anchor: .top, whenRequested: model.noticeToken)`, and the notice area carries `.id(PromoteView.noticesAnchor)` *(new)*. `noticeToken` *(new, `private(set) var Int`)* is bumped by one helper, `showingNotices(_ change: () -> Void)` *(new)*: it records which of `flow.staleNote`, `saveError` and `proposeError` are set, runs `change`, and bumps the token once when any of them went from nil to set. Every assignment of those three in `promote()` and in the proposal loop goes through it | One scroll mechanism (the kit README forbids another `ScrollViewReader`), with Reduce Motion and the insert-then-scroll turn handled there; one place decides when a notice is new | A hand-built `ScrollViewReader` — the kit has one; option (b) — deleted on the board |
| D-11 | Try again's copy is a new `L10n.Promote.tryAgain` *(new)* = "Try again", as the board draws it. `L10n.Promote.graphUnavailable` is reworded to "Provenencia couldn't read this Evidence graph, so it can't list the rows to promote." | The old message told the researcher to go back and reopen, which Try again replaces (user, 2026-11-03) | "Read again", the canvas's word — the button re-runs the proposal as well as the read (user, 2026-11-03) |
| D-12 | `ScoreCandidate(r subjectcompare.Result, edge float64) Scored`: the `provenance` parameter and `Subject.Provenance` go. `Score = r.Evidence + edge` | Always 1 today, so every score is unchanged | Keeping the parameter for a future caller — nothing in the map adds one |
| D-13 | FakeStore gets `var proposePromoteGraphAlignmentError: Error?`, `var applyPromoteGraphAlignmentError: Error?` and `let proposePromoteGraphAlignmentHold = FakeHold()` *(new)*. Each method records its call in `recordedCalls` first, then waits on the hold (propose only), then throws the error when it is set. `promoteProposeDelays` and its delay claim go | Working rule "FakeStore is a seeded fake"; `add-swift-test` pauses a fake call with `FakeHold`, not a sleep. Once proposals never overlap, delays have no job | — |
| D-14 | One kit piece for a failed read with a retry: `PVRetryCallout` *(new)*, in `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift`, beside `PVCallout`. `init(title: (any PVCopy)? = nil, message: String, retry: some PVCopy, isRetrying: Bool, compact: Bool = false, retryIdentifier: String, onRetry: @escaping () -> Void)`; its body is `PVCallout(tone: .danger, title:, message:, compact:, announces: true) { PVButton(retry, variant: .secondary, size: .sm, loading: isRetrying, action: onRetry).accessibilityIdentifier(retryIdentifier) }`. Three call sites use it: the Promote load callout (`isRetrying: model.isLoading`), `SourceFormLoadArea`'s `.failed` case (`compact: true`, `isRetrying: false`, since a retry there switches the area to `.loading`), and `EvidenceGraphView.readErrorCallout` (`isRetrying: model.isRereadingGraph`). Each keeps its own copy (`Try again`, `L10n.Sources.formReadAgain`, `L10n.EvidenceGraph.readAgain`) and its identifiers | Promote's Try again would be a third hand-built copy; the two existing ones show busy two ways and announce nothing (review M-10). User, 2026-11-03: extract it in this PR. A colocated helper in the Callout folder is the lowest layer that holds it: content-agnostic, so not a recipe, and one concept beside `PVCallout` rather than a new folder | A new `Components/RetryCallout/` folder — the piece is a Callout with one fixed action; an `init` on `PVCallout` where `Actions ==` a private view — the generic constraint leaks the button type into the kit's API; leaving the two others as they are — declined by the user |

**Rules this plan bends.** `plan-pr` § 3 and the 2026-10-29 rule that every view whose look changes waits on a board: D-14 changes two views S14-D7 does not draw. On the Evidence graph canvas, the read-error callout's spinner moves from beside Read again into the button, and the callout now announces. In the Source forms (Copy Source, the graph form and the Artifact card), the read-error callout now announces and looks the same. The user chose to extract the shared piece now, knowing it changes those looks without a board (2026-11-03). STEP-6 records it in `completed.md`; no rule text changes, since the user's answer is for this PR only.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteModel` `pendingProposal`, `submitProposal()`, `PromoteProposalInput` *(new)* | members, private struct | `CatalogSearchModel` generation + `Task`; `CatalogCounts.refreshAll` generation (searched `generation`, `coalesc`, `latest` under `macos/App`) | Both let calls overlap; this one must not. One caller, so private (D-1) |
| `PromoteModel.isLoading`, `loadAttempt`, `retryLoad()` *(new)* | members | `EvidenceGraphModel.readGraphAgain()`; `SourceFormLoadArea` | Reuses the session's `reread`; the page's load also runs a proposal, so it keeps its own method |
| `PromoteModel.noticeToken`, `showingNotices(_:)` *(new)* | members | `pvScrollsRowIntoView(whenRequested:)` callers pass a token the same way | The token is the page's |
| `PromoteView.noticesAnchor` *(new)* | constant | `PVPageColumns` `restartsAt` ids | One scroll target |
| `L10n.Promote.tryAgain` *(new)* | key | `L10n.EvidenceGraph.readAgain`, `L10n.Sources.formReadAgain` ("Read again") | The board's word; this button re-runs the proposal too (D-11) |
| `PVRetryCallout` *(new)* | component helper | `SourceFormLoadArea` `.failed`; `EvidenceGraphView.readErrorCallout`; `PVCallout` actions row | Those two are the copies it replaces (D-14) |
| FakeStore `proposePromoteGraphAlignmentError`, `applyPromoteGraphAlignmentError`, `proposePromoteGraphAlignmentHold` *(new)* | hooks | `listSubjectsError`; `workspaceNavCountsHold` | Same shapes |

**Builds on.** `PVCallout`, its `announces:` and actions row (S14-14), `PVButton`'s `loading:`, `pvScrollsRowIntoView(whenRequested:)`, `WorkspaceSession.reread`, `FakeHold`, the FakeStore `…Error` hooks.
**Deletes.** The hand-built callout bodies in `SourceFormLoadArea`'s `.failed` case and `EvidenceGraphView.readErrorCallout`, with the latter's `.disabled(model.isRereadingGraph)`, its `ProgressView` and the `evidenceGraph.readError.reading` identifier (no test reads it); `PromoteModel.proposeGeneration`, `nextProposal()`, `repropose(generation:)` (its body moves into the loop), and the generation comment on `isProposing`; FakeStore `promoteProposeDelays` and its delay claim in `proposePromoteGraphAlignment`; the test `anOlderProposalAnsweringLastIsDropped` (replaced by T-2); `graphalign.Subject.Provenance`, `ScoreCandidate`'s `provenance` parameter, its `≤0 treated as 1` branch and its doc lines (`types.go` line 18, `score.go` lines 12–15); `Provenance: 1` in `layer.go`; the `, 1` argument in the four `align_test.go` calls and in `rows.go` line 150.
**Next adds.** None named.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| `mapRest()` while a decision's proposal runs | `submitProposal()` | The scope change is a pending input like a decision; the next proposal carries `onlySubjectID: nil` (T-5) |
| A proposal result dropped because a newer input is pending | `PromoteFlow.merge` | Never reached for the dropped result; the newer result merges as today (T-2) |
| Try again while rows are empty | `PromoteModel.load()` | Same path as the first load; `flow.load` replaces rows; `loadError` nil on success (T-6, T-7) |
| The page leaves while a proposal runs | the loop | D-1: the loop stops after the call in flight; a pending input is not sent |

### Scale

At the map's Scale (100 new people on one Source; B-8 measured 2.8 s per proposal on the audit machine), five quick picks run 2 proposals instead of 5, and N picks during one proposal run at most 2. The queue adds no work per row. No write, list or loop in Go changes.

### Edges

- **Assumptions.** The merge assumes the proposal answers the current decisions; D-4's drop-when-pending keeps that true (T-2). Done assumes no proposal is pending (D-3, T-3).
- **Neighbors.** One decision → one call (T-1). A decision after the last proposal returned → a new call (T-1).
- **Removed inputs.** A pending input replaced by a newer one is dropped; the newer one is a superset of decisions (D-2, T-2).
- **Real data.** None — no table- or seed-driven rule changes.
- **Untrusted size.** None — no data the app did not write is read.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/graphalign` | `Subject.Provenance` goes; `func ScoreCandidate(r subjectcompare.Result, edge float64) Scored` (D-12); `(*state).score` and `rows.go` drop the argument |
| `core/database/promotealign` | `layer.go`: drop `Provenance: 1` |

Errors: none change.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Notice area | snowflake | `PromoteView.swift` `notices` | Four callouts, `announces: true` (D-9); `.id(PromoteView.noticesAnchor)`; the `ScrollView` scrolls to it on `noticeToken` (D-10) |
| Load error | `PVRetryCallout(message: model.loadError, retry: L10n.Promote.tryAgain, isRetrying: model.isLoading, retryIdentifier: "promote.tryAgain") { model.retryLoad() }` | `notices` | `.id(model.loadAttempt)` (D-8). The board's small secondary Button in the Callout's actions row, as D-14 builds it |
| Evidence graph read error | `PVRetryCallout(title: L10n.EvidenceGraph.readErrorTitle, message:, retry: L10n.EvidenceGraph.readAgain, isRetrying: model.isRereadingGraph, retryIdentifier: "evidenceGraph.readError.readAgain") { model.readGraphAgain() }` | `EvidenceGraphView.readErrorCallout` | Keeps `.accessibilityIdentifier("evidenceGraph.readError.callout")` and the 360 pt frame |
| Source form read error | `PVRetryCallout(title: errorTitle, message: errorMessage(message), retry: L10n.Sources.formReadAgain, isRetrying: false, compact: true, retryIdentifier: "\(accessibilityIdentifierPrefix).readAgain", onRetry: onReadAgain)` | `SourceFormLoadArea` `.failed` | Keeps `.accessibilityIdentifier("\(accessibilityIdentifierPrefix).loadError")` |
| Stale, proposal, save | `PVCallout(tone: .danger, …, announces: true)` | `notices` | Copy unchanged |
| Footer | snowflake | `PromoteView.swift` `footer` | Unchanged |

States: load failed (callout + Try again, no rows, Done disabled); retrying (Try again loading, frame 1c); stale (frame 2a); save error (frame 3); stale re-proposal failed (proposal error only, frame 4). Accessibility: each notice announces when it appears; Try again is a labelled button.

### Strings

L10n keys *(new)*: `L10n.Promote.tryAgain` = "Try again" (`promote.error.tryAgain`, comment "Button in the Promote load error callout that loads the page again"). Changed: `L10n.Promote.graphUnavailable` (`promote.error.graphUnavailable`) = "Provenencia couldn't read this Evidence graph, so it can't list the rows to promote." (was "Provenencia couldn't read this Evidence graph, so it can't list the rows to promote. Go back to the graph and open Promote again."). Removed: none.

## 7. Steps

**STEP-1 — Provenance goes**
Files: `core/graphalign/types.go`, `score.go`, `align.go`, `rows.go`, `align_test.go`; `core/database/promotealign/layer.go`; `docs/promote-graph-alignment.md` §6 and §12. Skill: `add-match-rule`.
Do: D-12. In §6, replace "Graph alignment's `Subject.Provenance` is 1 today (`core/database/promotealign/layer.go`)." with "Graph alignment does not scale by provenance yet: a score is the pair's evidence plus its edge support." In §12, replace "Every Subject's provenance is 1; credibility and transcription certainty don't scale comparisons yet." with "A score is the pair's evidence plus its edge support; credibility and transcription certainty don't scale comparisons yet."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/...` passes with no expected value changed; T-9's `rg` has no hit.

**STEP-2 — FakeStore hooks**
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: D-13, except that `promoteProposeDelays` stays until STEP-3. In `proposePromoteGraphAlignment`, the hold's `wait()` comes after the `withState` claim (call recorded, `lastPromoteFixed` set, scripted answer taken) and before the error is thrown. Move `doneWaitsForAPendingProposal` onto the hold, keeping what it asserts about `canFinish`.
Check: every `PromoteModelTests` test passes.

**STEP-3 — One proposal at a time**
Files: `PromoteModel.swift`, `PromoteModelTests.swift`.
Do: D-1 – D-6. Delete what **Deletes** lists for `PromoteModel`, and delete `promoteProposeDelays` with its delay claim. Delete `anOlderProposalAnsweringLastIsDropped`: it tests two overlapping calls, which can no longer happen; T-2 replaces it.
Check: T-1 – T-5 pass; every remaining `PromoteModelTests` test passes; `rg promoteProposeDelays macos` has no hit.

**STEP-4 — Try again, through one retry callout**
Files: `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift` *(new)*, `macos/App/DesignSystem/README.md` (Callout row), `macos/App/Features/Sources/SourceFormLoadArea.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `PromoteModel.swift`, `PromoteView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-14 first, moving the two existing call sites onto it; then D-7, D-8, D-11. Add to the README's Callout row: "`PVRetryCallout` is a danger callout for a failed read, with one small secondary retry Button whose loading state shows the retry, and it always announces".
Check: T-6, T-7, T-10, T-12, T-14 pass; `python3 scripts/check-localizable-xcstrings.py` and `python3 scripts/check-guidance-names.py` pass.

**STEP-5 — Notices announce and come into view**
Files: `PromoteView.swift`, `PromoteModel.swift`. Skill: `add-ui-component`.
Do: D-9, D-10.
Check: T-8 passes; manual check steps 2 and 3.

**STEP-6 — Docs**
Files: `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S14-D7-promote-notices.md` → `../design/archive/`, `../design/README.md`, `docs/promote-graph-alignment.md` §3.1.
Do: per the Definition of done. In §3.1's "On any change" bullet, add: "At most one proposal runs at a time; decisions made while it runs are sent together in the next one, and only the proposal answering the latest decisions is shown."
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-20-promote-recovers-and-speaks.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Fixtures: T-2 opens with `mapAll: true` (the `model(store:subjectID:graphID:mapAll:)` helper) so every row is visible without `mapRest()`. Rows come from the file's `row(_:handleID:)` helper, whose menu offers `handle:other`. "Held" means `store.proposePromoteGraphAlignmentHold.isOn = true` before the action, `await waitUntil { hold.count == 1 }`, then `hold.release()` per held call. Waits use `waitUntil` (`QueryHandleWaiting.swift`) on the named condition; no sleeps.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PromoteModelTests` `oneDecisionProposesOnce` *(new)* | load; one `setTarget`; wait `!isProposing` → 2 `proposePromoteGraphAlignment` calls recorded in total (load + 1), merged | AC-1.5 |
| T-2 | Swift | `PromoteModelTests` `quickDecisionsRunOneProposalAtATime` *(new)* | Subjects a–f, `mapAll: true`; `promoteProposals` = [load: every row `e1`; answer 1: every row `e1` but f `f-old`; answer 2: every row `e1` but f `f-new`]. Load. Hold on; `setTarget` a `handle:other`; wait `hold.count == 1`; `setTarget` b, c, d, e `handle:other`; `release()`; wait `hold.count == 1` → f's target is still `e1` (answer 1 dropped); `release()`; wait `!isProposing` → 3 calls in total; `lastPromoteFixed` holds a–e on `other`; f's target is `f-new` | AC-1.1, AC-1.2, AC-1.3, D-2, D-4 |
| T-3 | Swift | `PromoteModelTests` `doneWaitsForAPendingProposal` (kept, on the hold) | hold on; `setTarget` (held); `setTarget` again → `canFinish` false; release; wait `canFinish` → true; `lastPromoteBatch == nil` | AC-1.4, D-3 |
| T-4 | Swift | `PromoteModelTests` `aStaleDoneWhoseProposalFailsDropsTheStaleNote` *(new)* | Done stale (as `aStaleDoneReproposes`), with `proposePromoteGraphAlignmentError` set before Done; wait `!isProposing` → `staleNote == nil`, `proposeError != nil` | AC-2.7, D-5 |
| T-5 | Swift | `PromoteModelTests` `mappingTheRestWhileProposingSendsTheWholePage` *(new)* | page opened on "a" (not `mapAll`); hold on; `setTarget` on "a" (held); `mapRest()`; release; wait `!isProposing` → 3 calls in total, the last with `lastPromoteOnlySubjectID == nil` and "a" in `lastPromoteFixed` | D-2 |
| T-6 | Swift | `PromoteModelTests` `tryAgainLoadsThePage` *(new)* | `proposePromoteGraphAlignmentError` set → load → `loadError` set, `loadAttempt == 1`; clear the error; `retryLoad()` → `isLoading` true at once; wait `!isLoading` → `loadError == nil`, rows loaded | AC-2.1, AC-2.2, D-7 |
| T-7 | Swift | `PromoteModelTests` `tryAgainThatFailsAgainCountsTheAttempt` *(new)* | error stays set → load (`loadAttempt == 1`); `retryLoad()` twice in a row → one load runs (one more proposal call recorded); wait `!isLoading` → `loadAttempt == 2`, `loadError` set | AC-2.3, D-7, D-8 |
| T-8 | Swift | `PromoteModelTests` `aNewErrorBumpsTheNoticeToken` *(new)* | `noticeToken` bumps once on a stale Done, once on a save error (`applyPromoteGraphAlignmentError`), once on a proposal error; not on a second proposal error while the first still shows | AC-2.4, AC-2.5, AC-2.6, D-10 |
| T-9 | Go | `core/graphalign` and `core/database/promotealign` suites (existing), plus `rg -n 'Subject\.Provenance|Provenance:|provenance float64' core/graphalign core/database/promotealign` | suites pass with no expected value changed; the `rg` has no hit | AC-3.1, AC-3.2, D-12 |
| T-10 | Swift | `PromoteModelTests` `tryAgainReadsAFailedGraphAgain` *(new)* | `aPageWithoutItsGraphSaysSoInsteadOfLookingEmpty`'s setup (graph read fails) → `loadError == graphUnavailable`; seed the graph; `retryLoad()`; wait `!isLoading` → `loadError == nil`, rows loaded | AC-2.2, D-7 |
| T-12 | Swift | `PromoteModelTests` `tryAgainReadsAStaleGraphAgain` *(new)* | `proposePromoteGraphAlignmentError` set → load → `loadError` set; mark the graph handle stale (`(model.session.queryHandle(model.graphKey) as QueryHandle<EvidenceGraphRows>?)?.markStale()`); clear the error; `retryLoad()`; wait `!isLoading` → `loadError == nil`, rows loaded | D-7 |
| T-14 | Repo | `rg -n 'readAgain, variant|formReadAgain, variant|ProgressView\(\)' macos/App/Features/EvidenceGraph/EvidenceGraphView.swift macos/App/Features/Sources/SourceFormLoadArea.swift` | no hit: both sites build their callout through `PVRetryCallout` | D-14 |
| T-11 | Swift | `PromoteModelTests` `aStaleDoneReproposes` (kept) | Done stale; the re-proposal succeeds → `staleNote` set | AC-2.8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteModelTests` (S14-22's `theProposalCarriesThePagesScope` and `mappingTheRestProposesAgain` count proposals and must stay green), `core/graphalign` score tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`Subject.Provenance`, `provenance is 1`, `proposeGeneration`, `repropose(generation`, `promoteProposeDelays`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — 0.2.12
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated: promote-graph-alignment.md §3.1 (one proposal at a time), §6 and §12 (no provenance scaling yet)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source of 30 people. Promote any Subject and Map the rest. Pick targets on five rows as fast as you can. Done stays disabled until the page settles, and the rows show the result of your last pick. Searching from the omnibar while you pick answers without waiting behind five proposals.
2. Scroll to the bottom. In a second window (⌘N), rename any record. With VoiceOver on, back in the first window, press Done. The list scrolls to its top and the stale notice shows above the first section, as board frame 2a draws it, and VoiceOver reads its title and message.
3. Resize the window across 1600 pt of pane width with the notice showing. It keeps the row list's width and centers with it (PN-W1).
4. On an Evidence graph canvas whose read fails (any graph whose read the session reports as failed), Read again shows its spinner inside the button while it reads, not beside it. VoiceOver reads the callout when it appears.
5. The save error, the stale re-proposal failure and Try again cannot be forced from the app without a debug hook. They are checked by T-4, T-6 – T-8.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Test timing: the queue changes how many calls the held-proposal tests see | T-2 states the new count; `FakeHold` replaces the delays |

**Stop and report if:**

- `PromoteModel.repropose(generation:)` or `PromoteView.notices` no longer match §4 at your base.
- `session.reread(graphKey)` followed by `load()` still finds no graph in T-10.
- Any `core/graphalign` or `promotealign` test needs a changed expected value after STEP-1 (D-12 says scores are unchanged).
- The agreed board draws something D-10 does not cover (a new component, a dismiss button, a notice outside the notice area).
- `anOlderProposalAnsweringLastIsDropped` or `doneWaitsForAPendingProposal` cannot move onto the hold without changing what it asserts about the merged rows or `canFinish`.

## 12. Open questions

None. PN-1's pick (option (a)) is copied into D-10 from the agreed board.

## 13. PR

**Title:** S14-20 — The Promote page runs one proposal at a time and its errors come into view
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names, the AC-3.1 `rg`); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-20-promote-recovers-and-speaks.md`.
