# S14-20 — The Promote page recovers and speaks

**Status:** Blocked — on board S14-D7
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
**Review:** — (not reviewed yet; review-plan runs once the board is agreed and PN-1's pick is copied in)

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
| AC-1.1 | A loaded page; five `setTarget` calls while the first proposal is still running (first proposal delayed 300 ms) | The store records exactly 2 proposal calls after load: the first, and one carrying all five decisions |
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
| AC-3.1 | `rg -n "Provenance" core/graphalign core/database/promotealign --type go` | No hit outside test strings (`textProp(…, "Provenance")` place names) |
| AC-3.2 | The `core/graphalign` and `core/database/promotealign` suites | Pass with no expected value changed |

## 3. Scope

| In | Covered by |
| --- | --- |
| One serial, latest-wins runner for the page's proposals; `proposeGeneration` and `nextProposal` go (B-12) | STEP-2, STEP-3, AC-1.1 – AC-1.5 |
| A failed proposal clears the stale notice (B-13's stale half, PN-3) | STEP-3, AC-2.7, AC-2.8 |
| Try again on a failed load (B-17, PN-2) | STEP-4, AC-2.1 – AC-2.3 |
| Stale, save, proposal and load errors announce, and a new stale, save or proposal error scrolls the list to its top (B-17, PN-1 option (a), PN-W1) | STEP-5, AC-2.4 – AC-2.6 |
| FakeStore error hooks for propose and apply, in the existing `…Error` shape | STEP-3 |
| `graphalign.Subject.Provenance` and `ScoreCandidate`'s `provenance` parameter go (B-37) | STEP-1, AC-3.1, AC-3.2 |

| Out | Home |
| --- | --- |
| A decided row whose record or Subject went away in another window: Go releasing the pair, a released-reason wire vocabulary, row notes for it (B-13's main fix) | Descoped by the user, 2026-11-03: a far edge case; today's `promote.invalid` error stays. Noted on B-13 in [`../app-health-audit-2.md`](../app-health-audit-2.md) |
| A `promote.invalid` at Done proposing again once | Descoped with B-13 (user, 2026-11-03). Apply already checks the revision before it files a row (§4), so an invalid row after a passing revision check is not a race |
| The "start as Skip" header line (B-16) | Already fixed on `main` (§4); nothing to do |
| Announcing rows marked Updated after a proposal (named in B-17's Problem) | Not planned: no scope entry or working rule asks for it |
| A pending sign on the connection subrows and the footer's connection count while a proposal runs (S14-22 ship log UX-1) | Dropped from S14-20 by the user, 2026-11-03: not drawn on S14-D7. Not planned |
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
| Callout actions | `PVCallout.init(…, @ViewBuilder actions:)` | An actions row under the message; `SourcePageArtifactsView.swift` and `EvidenceGraphView.swift` pass one |
| Button loading | `macos/App/DesignSystem/Components/Button/PVButton.swift` `loading:` | Shows a spinner and blocks the action |
| Header line (B-16) | `macos/App/Platform/L10n.swift` `L10n.Promote.openedFrom` | "Opened from %@". No "start as Skip" anywhere: `rg -n "start as Skip" macos docs/*.md .agents` has no hit |
| Apply order | `core/database/promote/batch.go` `SaveBatch` | Checks the graph, then `latestRevision` against `SeenRevision` (`ErrStale`), then files rows. `ResolvePins` is gone (S13-24) |
| Fixed pairs | `core/database/promotealign/promotealign.go` `checkFixed` | Unchanged by this PR: a pair naming a Subject not on the graph, or a deleted, merged or wrong-kind handle, is `promote.ErrInvalid` |
| Provenance | `core/graphalign/types.go` `Subject.Provenance`; `core/graphalign/score.go` `ScoreCandidate(r, edge, provenance float64)`; `core/graphalign/align.go` `(*state).score` passes `s.Provenance`; `core/graphalign/rows.go` line 150 passes `1`; `core/database/promotealign/layer.go` line 134 sets `Provenance: 1` | Always 1, so `Score = (r.Evidence + edge) * 1` |
| Provenance tests | `core/graphalign/align_test.go` lines 778, 791, 805, 840 | Call `ScoreCandidate(…, 1)` |
| Provenance docs | `docs/promote-graph-alignment.md` §6, line 137 | "Graph alignment's `Subject.Provenance` is 1 today (`core/database/promotealign/layer.go`)." |
| FakeStore | `macos/App/Platform/FakeStore.swift` `proposePromoteGraphAlignment`, `promoteProposals`, `promoteProposeDelays`, `applyPromoteGraphAlignment` | Scripted proposals and delays; no error hook for either call; apply throws only `promote.stale` |
| Tests | `macos/ProvenenciaTests/PromoteModelTests.swift` `anOlderProposalAnsweringLastIsDropped`, `doneWaitsForAPendingProposal`, `aStaleDoneReproposes`, `aPageWithoutItsGraphSaysSoInsteadOfLookingEmpty`, and S14-22's `theProposalCarriesThePagesScope`, `mappingTheRestProposesAgain`; `macos/ProvenenciaTests/QueryHandleWaiting.swift` `waitUntil` | Cover generation drop, Done waiting, stale re-proposal, the graph-missing load error, and the scope a proposal carries |

**Precedent.** `CatalogSearchModel.scheduleSearch` (`macos/App/Features/Catalog/Search/CatalogSearchModel.swift`) is the nearest latest-wins shape: a generation counter, a `Task`, and `settle()` for tests. Copy its `settle()` idea for tests. Do not copy its debounce or its cancel: a proposal is not debounced, and an FFI call in flight cannot be cancelled, so the runner waits for it. S14-14 ([#675](https://github.com/mendahu/provenencia/pull/675)) is the precedent for `announces: true` on a danger callout (`SourcePageView.swift` line 99).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A generic `LatestRunner<Input: Sendable, Output: Sendable>` *(new)*, `@MainActor @Observable final class`, in `macos/App/Platform/LatestRunner.swift` *(new)*. It runs `work(input)`; a `submit` while it runs replaces the one pending input; when `work` returns, it calls `deliver(output)` only if nothing is pending, then runs the pending input if there is one | One owner for "one in flight, latest wins", tested on its own without a store | Keeping the generation counter and adding a queue inside `PromoteModel` — because the counter is exactly what lets five calls run; a Go-side cancel — because `catalogsession.Do` cannot cancel a running op |
| D-2 | The pending input is `PromoteProposalInput` *(new, a private struct in `PromoteModel.swift`)*: `fixed` and `onlySubjectID`, both taken at `submit` time from `flow.fixed(for: flow.rows.filter(\.decided))` and `flow.onlySubjectID` | The last submit carries every decision and the page's scope as they are now, so dropping earlier pending inputs loses nothing | Reading `flow` when the pending call starts — because it hides the input from the runner's tests |
| D-3 | `PromoteModel.isProposing` becomes `proposals.isBusy` *(new: running or pending)*. `canFinish` keeps reading `isProposing` | Done must wait for the decision that is pending, not only the one in flight | — |
| D-4 | The runner's output is `Result<(CatalogPromoteGraphAlignmentProposal, EvidenceGraphSnapshot?), Error>`; `deliver` does what `repropose` does after the call today (merge, or `graphUnavailableRepropose`, or `proposeError`) | Moves only the call into the runner; the merge rules stay as they are | — |
| D-5 | A delivered failure (store error or missing snapshot) sets `flow.staleNote = nil` as well as `proposeError` | The stale note claims the rows were proposed again; that is false once a proposal fails. `proposeAfterDecision` already clears it on the next decision | Clearing it only for the stale re-proposal — because every failed proposal makes the claim false |
| D-6 | The stale path in `promote()` calls `proposals.submit(…)` and returns; it no longer awaits | One path for every proposal | — |
| D-7 | `load()` stays outside the runner. It sets `isLoading` *(new)* true for its duration. `retryLoad()` *(new)* returns at once when `isLoading`, else bumps `loadAttempt` *(new)* and starts `load()` in a `Task` | The first load also reads the graph snapshot and builds bridges; it runs once per page, plus each Try again, and Try again is disabled while it runs | Running load through the runner — because a decision cannot happen before rows exist |
| D-8 | The load callout carries `.id(model.loadAttempt)`, so a repeated identical failure remounts and is announced again | `PVAnnounceModifier` does not repeat an unchanged text while mounted (§4) | Clearing `loadError` before retrying — because PN-2 keeps the callout during the retry |
| D-9 | All four notices pass `announces: true` | Working rule "Kit components announce themselves"; danger tone already gets high priority in `PVCallout` | Posting `AccessibilityAnnouncement` from the model — because the rule forbids call-site posts |
| D-10 | PN-1 is option (a), as the board agreed. The `ScrollView` gets a `ScrollViewReader`; the notice area carries `.id(PromoteView.noticesAnchor)` *(new)*, and the view scrolls to it with `anchor: .top`, animated, on a change of `model.noticeToken` *(new)*: a counter the model bumps whenever `flow.staleNote`, `saveError` or `proposeError` goes from nil to set. The notice area keeps its place and order | One notice area in one place; the stale notice's "review the updated rows" starts where the rows start (board, frame 2a) | Option (b), callouts above the footer — deleted on the board; a toast — ruled out by the brief |
| D-11 | Try again's copy is a new `L10n.Promote.tryAgain` *(new)* = "Try again" | No shared "Try again" key exists in `L10n.swift`; the Promote namespace owns its button copy | A shared `L10n.Common` key — because none exists, and adding a namespace for one key is out of scope |
| D-12 | `ScoreCandidate(r subjectcompare.Result, edge float64) Scored`: the `provenance` parameter and `Subject.Provenance` go. `Score = r.Evidence + edge` | Always 1 today, so every score is unchanged; the doc's "Provenance scales each comparison" stays as the rule for the future comparison, worded as not built | Keeping the parameter for a future caller — because nothing in the map adds one |
| D-13 | FakeStore gets `var proposePromoteGraphAlignmentError: Error?` and `var applyPromoteGraphAlignmentError: Error?` *(new)*, checked first in each method, in the shape of `listSubjectsError` | Working rule "FakeStore is a seeded fake": scripted errors only | — |

**Rules this plan bends.** `plan-pr` § 3 (a PR that changes chrome waits on a brief): not bent. The map row said "no new chrome"; the user chose a brief on 2026-11-03, and the plan waits on S14-D7.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `LatestRunner` *(new)* | type | `CatalogSearchModel` generation + `Task` (searched `generation`, `coalesc`, `latest` under `macos/App`) | That model debounces and lets calls overlap; this one must never overlap. `ThumbnailCache` coalesces equal keys, not a newest input |
| `LatestRunnerTests` *(new)* | test file | `PromoteModelTests` | The runner's rules are tested without a store |
| `PromoteModel.isLoading`, `loadAttempt`, `retryLoad()`, `noticeToken` *(new)* | members | `CatalogSearchModel.isLoading` | Page-own state |
| `PromoteView.noticesAnchor` *(new)* | constant | — | One scroll target |
| `L10n.Promote.tryAgain` *(new)* | key | Messages that say "Try again." in prose (`L10n.swift` line 1288) | Those are sentences, not a button label |
| FakeStore `proposePromoteGraphAlignmentError`, `applyPromoteGraphAlignmentError` *(new)* | hooks | `listSubjectsError` and the other `…Error` hooks | Same shape |

**Builds on.** `PVCallout`'s `announces:` and actions row (S14-14), `PVButton`'s `loading:`, the FakeStore `…Error` hook pattern.
**Deletes.** `PromoteModel.proposeGeneration`, `nextProposal()`, `repropose(generation:)` (its body moves into the runner's `work` and `deliver`), the generation comment on `isProposing`; `graphalign.Subject.Provenance`, `ScoreCandidate`'s `provenance` parameter and its `≤0 treated as 1` branch and doc line; `Provenance: 1` in `layer.go`; the `, 1` argument in the four `align_test.go` calls and `rows.go` line 150.
**Next adds.** None named.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| `mapRest()` while a decision's proposal runs | `LatestRunner.submit` | The scope change is a pending input like a decision; the next proposal carries `onlySubjectID: nil` (T-13) |
| A proposal result dropped because a newer input is pending | `PromoteFlow.merge` | Never reached for the dropped result; the newer result merges as today (AC-1.3) |
| Try again while rows are empty | `PromoteModel.load()` | Same path as the first load; `flow.load` replaces rows; `loadError` nil on success (AC-2.2) |
| The page leaves while a proposal runs | `LatestRunner` | `deliver` captures the model weakly; a delivered result on a gone model does nothing. The pending input is dropped with the runner |

### Scale

At the map's Scale (100 new people on one Source; B-8 measured 2.8 s per proposal on the audit machine): five quick picks run 2 proposals instead of 5, and N picks during one proposal run at most 2. The runner adds no work per row. No write, list or loop in Go changes.

### Edges

- **Assumptions.** The merge assumes the proposal answers the current decisions; D-1's "deliver only when nothing is pending" keeps that true (T-1, T-5). Done assumes no proposal is pending (D-3, T-6).
- **Neighbors.** One decision → one call (AC-1.5, T-4). A decision made after the last proposal delivered → a new call (T-2).
- **Removed inputs.** A pending input replaced by a newer one is dropped; the newer one is a superset of decisions (D-2, T-1).
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

`LatestRunner` sketch:

```swift
@MainActor @Observable
final class LatestRunner<Input: Sendable, Output: Sendable> {
    init(work: @escaping @MainActor (Input) async -> Output,
         deliver: @escaping @MainActor (Output) -> Void)
    private(set) var isBusy: Bool          // running, or an input pending
    func submit(_ input: Input)            // starts now if idle, else replaces the pending input
    func settle() async                    // returns once idle; tests await it
}
```

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Notice area | snowflake | `PromoteView.swift` `notices` | Four callouts, `announces: true` (D-9); PN-1 per D-10 |
| Load error | `PVCallout(tone: .danger, message:, announces: true) { … }` | `notices` | Actions: `PVButton(L10n.Promote.tryAgain, variant: .secondary, size: .sm, loading: model.isLoading) { model.retryLoad() }`, `.accessibilityIdentifier("promote.tryAgain")`; `.id(model.loadAttempt)` (D-8) |
| Stale, proposal, save | `PVCallout(tone: .danger, …, announces: true)` | `notices` | Copy unchanged; a new one scrolls the list to its top (D-10) |
| Footer | snowflake | `PromoteView.swift` `footer` | Unchanged |

States: load failed (callout + Try again, no rows, Done disabled); retrying (Try again loading); stale (PN-1); stale re-proposal failed (proposal error only, PN-3). Accessibility: each notice announces when it appears; Try again is a labelled button.

### Strings

L10n keys *(new)*: `L10n.Promote.tryAgain` = "Try again" (`promote.error.tryAgain`, comment "Button in the Promote load error callout that loads the page again"). Removed: none. Old-scope copy: none, since no action changes what it covers.

## 7. Steps

**STEP-1 — Provenance goes**
Files: `core/graphalign/types.go`, `score.go`, `align.go`, `rows.go`, `align_test.go`; `core/database/promotealign/layer.go`; `docs/promote-graph-alignment.md` §6. Skill: `add-match-rule`.
Do: D-12. In §6, replace the sentence "Graph alignment's `Subject.Provenance` is 1 today (`core/database/promotealign/layer.go`)." with "Graph alignment does not scale by provenance yet: a score is the pair's evidence plus its edge support."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/...` passes with no expected value changed; AC-3.1's `rg` has no hit.

**STEP-2 — The runner**
Files: `macos/App/Platform/LatestRunner.swift` *(new)*, `macos/ProvenenciaTests/LatestRunnerTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-1 as the sketch.
Check: T-1 – T-3 pass.

**STEP-3 — Proposals through the runner**
Files: `PromoteModel.swift`, `macos/App/Platform/FakeStore.swift`, `PromoteModelTests.swift`.
Do: D-2 – D-6, D-13. `proposeAfterDecision` clears `flow.staleNote` and submits. Delete what **Deletes** lists for Swift.
Check: T-4 – T-8 pass; every existing `PromoteModelTests` test passes. `anOlderProposalAnsweringLastIsDropped` may change only its call count expectations, as T-5 states.

**STEP-4 — Try again**
Files: `PromoteModel.swift`, `PromoteView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-localized-string`, `add-ui-component`.
Do: D-7, D-8, D-11.
Check: T-9, T-10 pass; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-5 — Notices announce and come into view**
Files: `PromoteView.swift`, `PromoteModel.swift`. Skill: `add-ui-component`.
Do: D-9, D-10.
Check: T-11 passes; the manual check's steps 2 – 4.

**STEP-6 — Docs**
Files: `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S14-D7-promote-notices.md` → `../design/archive/`, `../design/README.md`, `docs/promote-graph-alignment.md` §3.1.
Do: per the Definition of done. In §3.1's "On any change" bullet, add: "At most one proposal runs at a time; decisions made while it runs are sent together in the next one, and only the proposal answering the latest decisions is shown."
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-20-promote-recovers-and-speaks.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `LatestRunnerTests` `submitsWhileRunningCollapseToTheLatest` *(new)* | `work` waits on a gate; submit 1, 2, 3, 4; open the gate → `work` saw [1, 4]; `deliver` saw only 4's output | D-1, D-2, AC-1.1 |
| T-2 | Swift | `LatestRunnerTests` `anIdleRunnerRunsAtOnce` *(new)* | submit 1, settle, submit 2, settle → `work` saw [1, 2]; `deliver` saw both | D-1, AC-1.5 |
| T-3 | Swift | `LatestRunnerTests` `busyCoversRunningAndPending` *(new)* | `isBusy` true from the first submit until `settle` returns after the pending input ran | D-3 |
| T-4 | Swift | `PromoteModelTests` `oneDecisionProposesOnce` *(new)* | load; one `setTarget` → 2 `proposePromoteGraphAlignment` calls recorded in total (load + 1) | AC-1.5 |
| T-5 | Swift | `PromoteModelTests` `quickDecisionsRunOneProposalAtATime` *(new)*, and `anOlderProposalAnsweringLastIsDropped` updated | first post-load proposal delayed 300 ms; five `setTarget` on five rows → 3 calls in total (load + 2); `store.lastPromoteFixed` holds all five; the rows show the last proposal's targets | AC-1.1, AC-1.2, AC-1.3 |
| T-6 | Swift | `PromoteModelTests` `doneWaitsForAPendingProposal` (kept) | as today, plus `canFinish` false while a second decision is pending | AC-1.4 |
| T-7 | Swift | `PromoteModelTests` `aStaleDoneWhoseProposalFailsDropsTheStaleNote` *(new)* | Done stale; `proposePromoteGraphAlignmentError` set before the re-proposal → `staleNote == nil`, `proposeError != nil` | AC-2.7, D-5 |
| T-8 | Swift | `PromoteModelTests` `aStaleDoneReproposes` (kept) | `staleNote` set after a successful re-proposal | AC-2.8 |
| T-9 | Swift | `PromoteModelTests` `tryAgainLoadsThePage` *(new)* | `proposePromoteGraphAlignmentError` set → load → `loadError` set; clear the error; `retryLoad()`; wait → `loadError == nil`, rows loaded, `loadAttempt == 1` | AC-2.1, AC-2.2 |
| T-10 | Swift | `PromoteModelTests` `tryAgainThatFailsAgainCountsTheAttempt` *(new)* | error stays set → `retryLoad()` twice, waiting between → `loadAttempt == 2`, `loadError` set; a `retryLoad()` while `isLoading` does not bump it | AC-2.3, D-7 |
| T-11 | Swift | `PromoteModelTests` `aNewErrorBumpsTheNoticeToken` *(new)* | `noticeToken` bumps on a stale Done, on a save error (`applyPromoteGraphAlignmentError`), and on a proposal error; not on a second identical proposal error while one shows | AC-2.4, AC-2.5, AC-2.6, D-10 |
| T-12 | Go | `core/graphalign` and `core/database/promotealign` suites (existing), plus AC-3.1's `rg` in the PR description | `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/...` → pass, no expected value changed; the `rg` → no hit | AC-3.1, AC-3.2, D-12 |
| T-13 | Swift | `PromoteModelTests` `mappingTheRestWhileProposingSendsTheWholePage` *(new)* | first post-load proposal delayed 300 ms after a `setTarget`; `mapRest()` while it runs → 3 proposal calls in total, the last with `lastPromoteOnlySubjectID == nil` and the decided row in `lastPromoteFixed` | D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteModelTests` (timing; S14-22's `theProposalCarriesThePagesScope` and `mappingTheRestProposesAgain` count proposals and must stay green), `core/graphalign` score tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`Subject.Provenance`, `proposeGeneration`, `repropose(generation`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — 0.2.12
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated: promote-graph-alignment.md §3.1 (one proposal at a time) and §6 (no provenance scaling yet)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source of 30 people. Promote any Subject and Map the rest. Pick targets on five rows as fast as you can. Done stays disabled until the page settles, and the rows show the result of your last pick. Searching from the omnibar while you pick answers without waiting behind five proposals.
2. Scroll to the bottom. In a second window (⌘N), rename any record. With VoiceOver on, back in the first window, press Done. The list scrolls to its top and the stale notice shows above the first section, as board frame 2a draws it, and VoiceOver reads its title and message.
3. Resize the window across 1600 pt of pane width with the notice showing. It keeps the row list's width and centers with it (PN-W1).
4. The save error, the stale re-proposal failure and Try again cannot be forced from the app without a debug hook. They are checked by T-7, T-9 – T-11.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Test timing: the runner changes how many calls the delayed-proposal tests see | T-5 states the new count; `settle()` replaces sleeps where a test waits for the runner |
| `@Observable` on a generic class | Supported by the Observation macro; if the build refuses it, stop (below) |

**Stop and report if:**

- `PromoteModel.repropose(generation:)` or `PromoteView.notices` no longer match §4 at your base.
- The `@Observable` macro does not compile on `LatestRunner<Input, Output>`.
- Any `core/graphalign` or `promotealign` test needs a changed expected value after STEP-1 (D-12 says scores are unchanged).
- The agreed board draws something D-10 does not cover (a new component, a dismiss button, a notice outside the notice area).
- `anOlderProposalAnsweringLastIsDropped` cannot be updated to T-5's counts without changing what it asserts about the merged rows.

## 12. Open questions

None. PN-1's pick (option (a)) is copied into D-10 from the agreed board.

## 13. PR

**Title:** S14-20 — The Promote page runs one proposal at a time and its errors come into view
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names, the AC-3.1 `rg`); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-20-promote-recovers-and-speaks.md`.
