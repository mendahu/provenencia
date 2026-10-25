# S11-35 — Hold a record Promote did not offer

**Status:** Landed in [#543](https://github.com/mendahu/provenencia/pull/543) (2026-10-25)
**Spike / slice:** Spike 11 · Slice 8 — Find another record
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the call has no surface until S11-36 adds the menu entry and the picker; a held record reads through the existing row readout, which already draws band None with `decided` and Contested)
**Board:** —
**Depends on:** —
**Blocks:** S11-36
**Branch:** `s11-35-hold-found-record`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [Find another record § Shape](../promote-find-another-record.md#shape), [§ What the engine already does](../promote-find-another-record.md#what-the-engine-already-does), [§ Where the Swift side changes](../promote-find-another-record.md#where-the-swift-side-changes); working rules "A found record is a decided row, nothing new", "The picked record joins the row's menu", "Engine first, proven by tests"
**Planned:** 2026-10-25 against `main` @ `22e3878`
**Review:** [`S11-35-hold-found-record.review.md`](S11-35-hold-found-record.review.md) — Fits with changes / Ready with changes, 1 round (findings fixed)

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

The Promote page can hold a row on a record that Promote never offered: a call on `PromoteFlow` takes the record's id, ref and list header, builds its menu item the way the menu builds every other, selects it and marks the row decided, and the page sends it to Go as a fixed pair on every re-proposal. Go already holds, scores, explains and files such a pair; this PR pins that with tests instead of assuming it. It is the engine half of [Find another record](../promote-find-another-record.md); the menu entry and the search that call it are S11-36.

**Why now.** Since S11-05, a record whose first given name or surname differs from the Subject's is never a candidate, so a researcher who knows they are one person cannot file the Subject on it from Promote (the feature doc's "The problem"). S11-36 needs a call to hand its pick to, and S11-35 and S11-37 run in parallel ahead of it.

**What stays.** The target menu, its tokens and `setTarget`'s behavior; the proposal wire (no new field, flag or reason code); every Go package's code; `FakeStore`.

## 2. User stories

**US-1.** As a researcher, when I choose a record for a row that Promote did not offer, the row holds it, decided, and the menu shows it as the selection, so that I can go back to an offered record, New or Skip.

In this table, **Polly** is the header `.person(CatalogPersonHeader(entity: CatalogCanonicalEntity(id: "e7", ref: "PER-7", subjectTypeID: "t", label: ""), name: CatalogNameValue(form: "Polly Jones", parts: [given "Polly", surname "Jones"]), nameValueCount: 1))`, and the flow's `nameOrders` are the natural / sorted orders of `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person row `a` opened unset with reason `noMatch` and an empty menu; `hold(subjectID: "a", id: "e7", ref: "PER-7", header: Polly)` | Returns true. Target `.handle(id: "e7", ref: "PER-7", title: "Jones, Polly")`; the menu's last item is `e7` with title `Jones, Polly`, reading title `Polly Jones`, subtitle `PersonLifeDisplay.line(<Polly's person header>).text`, assessment `.none`, reason `.decided`; `decided` and `manual` are true; the row's `readingTitle` is `Polly Jones`; `selectionReadout` is `.none` / `.decided` |
| AC-1.2 | AC-1.1, then `fixed(for:)` over the decided rows and `batchRows()` (rest mapped) | `fixed` holds `(a, handle e7)`; the batch row for `a` is target `handle`, entity `e7`, `defaultPins` true |
| AC-1.3 | AC-1.1, then a proposal whose row `a` is target `handle`, handle `e7`, ref `PER-7`, assessment `weak`, reason `decided`, contested | The menu still holds `e7`; its assessment is `.weak`; its reason stays `.decided`; the row is contested and its target is `e7` |
| AC-1.4 | AC-1.3, then row `b` is decided with `setTarget(subjectID: "b", token: "handle:e2")` and a proposal arrives whose row `a` is target `new`, assessment `none` | Row `a` keeps target `e7` and `e7` in its menu |
| AC-1.5 | Row `a` with menu `[e1, e2]` on target `e1`; `hold(subjectID: "a", id: "e2", ref: "PER-2", header: nil)` | The menu is still `[e1, e2]`; `e2` keeps the title, assessment and reason it was offered with; the target is `e2` |
| AC-1.6 | Row `a` was holding `e1` with evidence and pins; `hold` with id `e7` | `board` is nil, `pins` empty, `contested` false, `move` nil (the selection reset) |
| AC-1.7 | Row `b` targets `e1` (ref `PER-1`); row `a`'s menu is empty; `hold(subjectID: "a", id: "e1", ref: "PER-1", header: nil)` | Row `a`'s duplicate note is `.sharesHandle(otherName: "b", ref: "PER-1")` |
| AC-1.8 | `hold` on an anchor row, on a subject id that is not a row, or with id `""` | Returns false; `rows` equal before and after (no menu item added) |
| AC-1.9 | AC-1.1 with `header: nil` | The menu item's title and reading title are `PER-7`; its subtitle is empty |

**US-2.** As the engine, when the page sends a fixed pair on a handle that no shortlist key reaches, I load it, score it, mark it contested on a conflict, explain it and file it like any decided handle.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Filed `Robins`; Subject `Robnis` on another Source, unfixed | The row neither targets the handle nor lists it among its alternatives |
| AC-2.2 | AC-2.1 with `Fixed{Robnis, handle}`; sex at birth on neither side / female on both / female filed against male | Target `handle` on that handle, reason `decided`, `ConflictWithFixed` false; a `name` comparison exists; the `sex_at_birth` comparison is `unknown` / `agree` / `conflict`; `Contested` false / false / true |
| AC-2.3 | Over FFI: filed `Mary Smith` (male); `Polly Jones` (male) on another Source, fixed on that handle | Row target `handle`, the handle's id, reason `decided`, assessment `none`, contested true |
| AC-2.4 | AC-2.3's pair through `GetPromoteClaimEvidence` | No error; a `name` section and a `sex_at_birth` section; at least one offer |
| AC-2.5 | AC-2.3's pair through `ApplyPromoteGraphAlignment` with `DefaultPins` | One written row on the handle's entity id, with pins > 0 |

AC-2.1 – AC-2.5 were observed by throwaway tests at the planning base, deleted before this plan was written.

**US-3.** As a contributor, the fake store holds a found record the way Go does, so the Swift tests and S11-36's previews can drive it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `FakeStore` with no scripted proposals; Ada (`a`) unfiled on graph `g-1`; a person entity `e-far` filed through Subject `x` on another graph; `PromoteModel.hold(subjectID: "a", id: "e-far", ref: "PER-FAR", header: nil)` | `store.lastPromoteFixed == [(a, handle e-far)]`; after the proposal the row targets `e-far`, the menu holds it with reason `.decided`; `done()` sends a batch row `a` → `handle`, entity `e-far` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PromoteFlow.hold(subjectID:id:ref:header:)` *(new)*: builds the menu item through the flow's own title and subtitle helpers, adds it with reason `decided` and band None, then selects it | STEP-3, AC-1.1 – AC-1.9 |
| `PromoteFlow.select(_:at:)` *(new, private)*: `setTarget`'s selection body, shared by `setTarget` and `hold` | STEP-3, AC-1.6, existing `PromoteFlowTests` |
| `PromoteModel.hold(subjectID:id:ref:header:)` *(new)*, with `setTarget`'s re-propose follow-up through one private helper | STEP-4, AC-3.1 |
| Go tests: an off-shortlist fixed handle is loaded, banded, contested on a conflict (`promotealign`), and the same over the FFI for propose, evidence and Apply (`handlers`) | STEP-1, STEP-2, AC-2.1 – AC-2.5 |
| `FakeStore` holds a fixed handle outside its candidates: pinned by a model test, no code change (D-7) | STEP-4, AC-3.1 |
| Feature doc and map: the engine claims are pinned and where; the stale `FakeStore` and call-shape sentences; the map row links this plan | STEP-5 |

| Out | Home |
| --- | --- |
| The "Find another record" entry in the target menu, the inline picker, passing a `SearchHit`'s `id`, `ref` and `header` to `hold` | S11-36 |
| The shared search model and hit row | S11-37 |
| How a held record with band None reads (copy, styling) | S11-D3, S11-36 |
| Any new proposal field, flag or reason code | not planned (working rule "A found record is a decided row, nothing new") |
| Go code changes | not planned unless a test in STEP-1 or STEP-2 fails (Stop conditions) |
| A kind check in `hold` or in `FakeStore`'s held row | not planned: S11-36 searches the row's kind only, and Go's `checkFixed` refuses another kind with `promote.invalid` |
| `FakeStore` scripted proposals (`promoteProposals`) honouring `fixed` | not planned: tests script the row they expect |
| Replacing the `"handle:"` token spelled in `PromoteView.targetOptions` and `PromoteTargetMenuRow` | not planned: S11-36 touches that menu and decides its sentinel token |
| Searching from the evidence sheet; records of another kind; merging records | [what stays later](../promote-find-another-record.md#what-stays-later) |

## 4. Current state

Verified at `22e3878`. The Go rows were also checked by running throwaway tests against the planning base (the results are in AC-2.1 – AC-2.5).

| Area | Where | Today |
| --- | --- | --- |
| Menu pick | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.setTarget(subjectID:token:)` (line 432) | Guards on the row and `!anchor` (line 433), resolves the token through `target(for:on:)`; on a change clears `board`, `pins`, `contested`, `move`; then sets `decided`, clears `updated`, sets `manual`, runs `noteDuplicates()`, returns true (lines 434 – 449) |
| Token resolution | same file, `target(for:on:)` (line 610) | `handle:<id>` resolves only when `row.menu` holds `<id>`; any other token leaves the target as it was. A record outside the menu has no way in. This is the gap |
| Re-proposal of a decided row | same file, `makeRow` decided branch (about lines 639 – 665) | `kept = previous`, so `menu` and `target` survive. When the proposal's row is target `handle` on the held id, the menu item's `assessment` is refreshed, and its `reason` only when the new one is not `.decided` |
| Menu item | same file, `PromoteFlow.Alternative` (line 140) | `id`, `ref`, `title` (sorted), `readingTitle`, `subtitle`, `assessment` (default `.none`), `reason` (default `.empty`) |
| Menu titles | same file, `menu(for:kind:orders:)` (line 707), `conclusionTitle(kind:person:event:place:fallback:orders:style:)` (line 759), `conclusionSubtitle(kind:person:event:place:)` (line 781), all `private static` | Turn a person / event / place header into the sorted and natural title and the subtitle, under `nameOrders` (an instance property the model keeps in step through `syncNameOrders`). The fallback title is the ref |
| Search hit header | `macos/App/Platform/GenealogyStore.swift` `CatalogSearchHit.header` (line 207), `enum CatalogConclusionHeader` (line 786) | `.person` / `.event` / `.place`: the inputs `conclusionTitle` and `conclusionSubtitle` take. What S11-36 will hold for a pick |
| Readout | `PromoteFlow.swift` `Row.selectionReadout` (line 192); `PromoteReadout.swift` preview "Readout badges" (line 42) | Reads the selected menu item's band and reason; the preview already draws `.none` + `.decided` + contested |
| Model | `macos/App/Features/Promote/PromoteModel.swift` `setTarget(subjectID:token:)` (line 173) | Calls the flow; on true clears `staleNote`, takes `nextProposal()` and starts `repropose(generation:)`, which sends `flow.fixed(for: flow.rows.filter(\.decided))` |
| Fake propose | `macos/App/Platform/FakeStore.swift` `proposePromoteGraphAlignment` (line 1640), `buildPromoteProposal` (line 1662) | Scripted `promoteProposals` win and ignore `fixed`. Unscripted, a fixed handle resolves among `membershipBySubject` entities of any graph and comes back target `handle`, `strong`, reason `decided`; a Subject with no membership and no fixed handle comes back `skip` / `none`, which the flow opens unset. The fake has no candidates and every fake entity is a membership's, so any record the fake's search (`handleSearchHits`, also over memberships) can find, it can hold. No code gap |
| Fake Apply | same file, `applyPromoteGraphAlignment` (line 1746) | A `handle` row files onto any membership entity of the Subject's type |
| Fixed check | `core/database/promotealign/promotealign.go` `checkFixed` (line 163) | A handle must be 16 bytes, a known unmerged node, of the Subject's kind. It never checks the handle was a candidate |
| Canon | `core/database/promotealign/canon.go` `loadCanon` (line 36) | Adds every fixed handle before the candidates; fixed Subjects get no shortlist |
| Scoring | `core/graphalign/align.go` `(*state).anchor` (line 184); `core/graphalign/rows.go` `proposal` (line 21) | Scores each fixed handle as the Subject's first candidate; the row reports `ReasonDecided`, `st.band(score)`, comparisons and `Contested` |
| Evidence | `core/database/promotealign/evidence.go` `Evidence` (line 40) | Validates the pair with `checkFixed`; any handle of the right kind |
| Apply | `api/ffi/handlers/promote_batch.go` `ApplyPromoteGraphAlignment`; `promotealign.ResolvePins`; `core/database/promote/batch.go` `SaveBatch` | A `handle` row files on any 16-byte entity of the Subject's type |
| Go tests | `core/database/promotealign/promotealign_test.go` `TestProposeDoesNotCompareAnUnshortlistedHandle` (line 1252), `TestProposeMarksAConflictingRowContested` (line 1222), `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` (line 279), helper `comparisonOutcome` (line 836) | Robins / Robnis is not shortlisted; nothing fixes such a handle |
| FFI tests | `api/ffi/handlers/promote_align_test.go` `personOnNewSource` (line 30, name fixed to `Mary Smith`), `TestProposeSendsContested` (line 86); `promote_evidence_test.go` `evidenceFixture` (line 24), `newEvidenceFixture`, `request()`, `evidence(t)`, `sectionOf`; `promote_batch_test.go` `TestApplyPromoteGraphAlignment` | Every fixed handle in these tests is a look-alike the shortlist also finds |
| Swift tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` helpers `proposal`, `row` (always `handleRef` `PER-1`), `fact`; `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu` (line 100), `anUnmatchedRowStartsUnselected` (line 242), `duplicateNotesNameTheOtherRowAndItsRef` (line 853); `macos/ProvenenciaTests/PromoteModelTests.swift` helpers `store`, `model`, `subject`; `newAndSkipDecisionsAreHeldOnTheNextProposal` (line 399), `doneSendsDefaultOrChosenPins` (line 310) | Menu picks only |

**Precedent.** `22600f2a` ("promote: hold New and Skip decisions, and warn when rows collide") made New and Skip fixed pairs: copy the Go test shape of `TestProposeHoldsNewAndSkipDecisions` (one table, a fixture per case) and the model test `newAndSkipDecisionsAreHeldOnTheNextProposal` (read `store.lastPromoteFixed` after `waitUntil { !model.isProposing }`). `TestProposeDoesNotCompareAnUnshortlistedHandle` gives the off-shortlist pair (`Robins` / `Robnis`). Do not copy that commit's engine changes; nothing in Go changes here.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The flow call is `mutating func hold(subjectID: String, id: String, ref: String, header: CatalogConclusionHeader?) -> Bool`, `@discardableResult`, on `PromoteFlow`, beside `setTarget`. The flow builds the menu item | Menu titles have one owner, the flow's header helpers under the flow's `nameOrders`; S11-36 passes `hit.id`, `hit.ref`, `hit.header` through and formats nothing (review M-1). "Held, not moved" is the feature doc's word | Taking a built `Alternative` — because the title helpers are `private static` and S11-36 would format titles a second way; a new `FoundRecord` type — because no new type is needed |
| D-2 | `hold` guards first: it returns false and changes nothing when `id` is empty, the row is missing, or the row is an anchor. Only then does it touch the menu | Same refusals as `setTarget`, plus an empty id; guarding first keeps a refused call from leaving an item in a menu (review X-1) | Asserting — because a stale pick after a reload must not crash the page |
| D-3 | When `row.menu` has no item with `id`, `hold` appends `Alternative(id:, ref:, title: <sorted>, readingTitle: <natural>, subtitle:, assessment: .none, reason: .decided)`, the titles and subtitle from the row's `kind` and the header (D-5); when it has one, the menu is unchanged and the existing item keeps its title, band and reason | The map's reason is `decided` (K-2): a found record was never offered. Band None until Go answers, then `makeRow` refreshes it. An offered record found by search is the same pick as from the menu (feature doc decided 3) | Inserting first — because the menu's order is Promote's offer, best first; replacing the offered item — because it would lose the sentence it was offered with |
| D-4 | `setTarget`'s selection body (lines 434 – 449) moves into `private mutating func select(_ next: Target, at index: Int)` *(new)*. `setTarget` keeps its guard, then returns `select(target(for: token, on: rows[index]), at: index)`. `hold` ends with `select(.handle(id: item.id, ref: item.ref, title: item.title), at: index)`, `item` being the menu item found or added | One selection path and one guard per entry point, with no `"handle:"` string built only to be parsed back (review M-2) | `hold` calling `setTarget` with a token — because it spells the token format a fifth time and repeats the guard; copying the body — because two copies drift |
| D-5 | The header-to-titles step reuses `conclusionTitle` and `conclusionSubtitle`. A `private static func headerParts(_ header: CatalogConclusionHeader?) -> (person: CatalogPersonHeader?, event: CatalogEventHeader?, place: CatalogPlaceHeader?)` *(new)* splits the enum; `fallback` is `ref`; `kind` is the row's `kind` | The existing helpers already take the three optionals; one small split is the whole adapter | A third `headerTitle` overload — because it would duplicate the sorted / natural pair the call site needs; making the helpers internal — because the flow stays their only user |
| D-6 | `PromoteModel.hold(subjectID:id:ref:header:)` calls the flow and, on true, runs the same follow-up as `setTarget`. That follow-up moves into one private helper, `proposeAfterDecision()` *(new)*, which both call | The map: "with `setTarget`'s re-propose follow-up" | A second inline copy of three lines — because the next decision call would make a third |
| D-7 | `FakeStore` code does not change. AC-3.1 pins that its unscripted proposal holds a membership entity from another graph and its Apply files it | §4: the fake has no candidates, so every held handle is "outside its candidates", and its search hits come from the same memberships. The map scope entry and the feature doc are corrected in STEP-5 | Adding a standalone entity store to the fake — because nothing can search or show such an entity |
| D-8 | No Go code. The engine claims are pinned by four Go tests: one in `promotealign` (no FFI) and three over the FFI handlers for propose, evidence and Apply | Working rule "Engine first, proven by tests"; map review K-7 put filing in the handler tests. The planning-base run passed all four claims | A `promote` package test for filing — because `SaveBatch` never sees whether a handle was a candidate; the FFI test covers `ResolvePins` too |
| D-9 | The off-shortlist proof in AC-2.1 is "the unfixed row neither targets the handle nor lists it", not "New with `no_match`" | S11-47 removes New and `no_match` from Go's proposals; the test must outlive it | Asserting `TargetNew` / `ReasonNoMatch` — because S11-47 would have to rewrite it |
| D-10 | Go tests pin target, handle, reason, `Contested`, `ConflictWithFixed` and comparison outcomes; only the FFI test pins a band (`none` for a different first name and surname), which the feature doc states | Bands move with registry knobs (S11-43, S11-45) | Pinning `weak` for Robnis — because a knob change would break a test that is about holding |
| D-11 | The FFI tests get `namedPersonOnNewSource(t, dir, userID, typeID, name, sexKey, polarity)` *(new)*; `personOnNewSource` keeps its signature and calls it with `"Mary Smith"`. `newFoundRecordFixture(t) evidenceFixture` *(new)* fills the existing fixture type: one promoted `Mary Smith` member (`members`, `claims` of length 1), `incoming` `Polly Jones`; T-2 – T-4 read its fields and T-3 calls `f.evidence(t)` | One fixture shape and request builder for the evidence and Apply tests (review M-3); the existing person helper hard-codes the name | A second fixture struct — because `evidenceFixture` has every field; changing every caller's arguments — because the Implementer block forbids touching unrelated callers |
| D-12 | T-1 is table-driven over the three sex cases. T-2, T-3 and T-4 are single-case tests: each checks one pair, so `.cursor/rules/go-tests.mdc`'s table rule has nothing to tabulate | The rule tabulates cases; these tests have one each (review X-5) | Folding T-2 – T-4 into one table — because they call three different RPCs |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PromoteFlow.hold(subjectID:id:ref:header:)` *(new)* | func | `PromoteFlow.setTarget(subjectID:token:)` | Shares `select(_:at:)` (D-4). A token cannot carry the record's ref and header |
| `PromoteFlow.select(_:at:)` *(new, private)* | func | `setTarget`'s body | Extracted, not copied |
| `PromoteFlow.headerParts(_:)` *(new, private static)* | func | `conclusionTitle`, `conclusionSubtitle` | Adapter in front of them, reused as they are (D-5); searched `rg "CatalogConclusionHeader" macos/App`: no shared split exists outside the views that switch on it inline |
| `PromoteModel.hold(subjectID:id:ref:header:)` *(new)* | func | `PromoteModel.setTarget(subjectID:token:)` | Same follow-up, shared through `proposeAfterDecision()` (D-6) |
| `PromoteModel.proposeAfterDecision()` *(new, private)* | func | the three lines in `setTarget` after the flow call | Extracted, not copied |
| `TestProposeHoldsAnUnshortlistedHandle` *(new)* | test | `TestProposeDoesNotCompareAnUnshortlistedHandle`, `TestProposeMarksAConflictingRowContested` | Uses their fixture and pair; searched `rg "Fixed\{" core/database/promotealign/*_test.go`: no test fixes an unshortlisted handle |
| `namedPersonOnNewSource` *(new)*, `newFoundRecordFixture` *(new)* | test helpers | `personOnNewSource`, `newEvidenceFixture` | `personOnNewSource` becomes a call to the named helper; the new fixture returns `evidenceFixture` (D-11) |
| `TestProposeHoldsAFoundRecord`, `TestGetPromoteClaimEvidenceForAFoundRecord` *(new)*; subtest "files a found record" | tests | `TestProposeSendsContested`, `TestGetPromoteClaimEvidence`, `TestApplyPromoteGraphAlignment` "pins the evidence's offers" | New cases; the existing ones fix only look-alikes |
| Six `PromoteFlowTests` and one `PromoteModelTests` case *(new)* | tests | `aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu`, `anUnmatchedRowStartsUnselected`, `newAndSkipDecisionsAreHeldOnTheNextProposal` | Existing helpers; no new helper |

**Builds on.** The fixed-pair path from `22600f2a`; the flow's header title helpers; `makeRow`'s decided branch, unchanged.
**Deletes.** None — nothing becomes dead. `setTarget`'s body moves into `select(_:at:)`, and the three follow-up lines in `PromoteModel.setTarget` into `proposeAfterDecision()`.
**Next adds.** S11-36 calls `PromoteModel.hold(subjectID:id:ref:header:)` with a `SearchHit`'s fields and formats nothing; a later decision call uses `select(_:at:)`. S11-47 changes what an unfixed row carries; D-9 keeps these tests out of its way.

### Schema

None — no catalog change.

### Go core

None — no production code. Tests only, in `core/database/promotealign` (D-8).

### FFI

None — no proto or handler change. Tests only, in `api/ffi/handlers`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Hold a found record | `PromoteFlow.hold(subjectID:id:ref:header:)` *(new)*, `select(_:at:)` *(new, private)*, `headerParts(_:)` *(new, private static)* | `macos/App/Features/Promote/PromoteFlow.swift`: `hold` after `setTarget`, `select` after `hold`, `headerParts` beside `conclusionTitle` | Sketch below |
| Model entry | `PromoteModel.hold(subjectID:id:ref:header:)` *(new)*, `proposeAfterDecision()` *(new, private)* | `macos/App/Features/Promote/PromoteModel.swift`, after `setTarget` | — |

```swift
/// Returns true when suggested rows should be proposed again.
@discardableResult
mutating func setTarget(subjectID: String, token: String) -> Bool {
    guard let index = rows.firstIndex(where: { $0.subjectID == subjectID }), !rows[index].anchor else { return false }
    return select(target(for: token, on: rows[index]), at: index)
}

/// Holds a record Promote did not offer: it joins the row's menu as the
/// researcher's choice, titled as the menu titles every record, then is
/// selected like a menu pick. Returns true when suggested rows should be
/// proposed again.
@discardableResult
mutating func hold(subjectID: String, id: String, ref: String, header: CatalogConclusionHeader?) -> Bool

/// Selects a target on a working row: the old target's evidence goes, the
/// row is decided. Returns true.
private mutating func select(_ next: Target, at index: Int) -> Bool

private static func headerParts(_ header: CatalogConclusionHeader?)
    -> (person: CatalogPersonHeader?, event: CatalogEventHeader?, place: CatalogPlaceHeader?)

// PromoteModel
func setTarget(subjectID: String, token: String) {
    guard flow.setTarget(subjectID: subjectID, token: token) else { return }
    proposeAfterDecision()
}

func hold(subjectID: String, id: String, ref: String, header: CatalogConclusionHeader?) {
    guard flow.hold(subjectID: subjectID, id: id, ref: ref, header: header) else { return }
    proposeAfterDecision()
}

/// A decision changed the fixed pairs: propose again.
private func proposeAfterDecision()
```

No view changes. States: the row reads its band and reason through `selectionReadout` as today; band None with `.decided` is already drawn (PromoteReadout preview "Readout badges").

### Strings

None — no copy. `.decided` is worded by the existing `L10n.Promote.reasonDecided`.

## 7. Steps

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`).

**STEP-1 — Pin the engine in `promotealign`**
Files: `core/database/promotealign/promotealign_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: add `TestProposeHoldsAnUnshortlistedHandle` (T-1) after `TestProposeDoesNotCompareAnUnshortlistedHandle`. One table over three cases: `"no sex"` (no sex Observation on either side), `"same sex"` (`female` on both), `"conflicting sex"` (`female` on Robins, `male` on Robnis); each case builds its own `newFixture(t)`. Filed `f.person(f.source, f.artifact, "Robins")`; cite its sex Observation, if any, **before** `f.promote`, as `TestProposeMarksAConflictingRowContested` does. Subject `f.person(src, art, "Robnis")` on `src, art := f.newSource("Register")`, with its sex Observation, if any. Propose unfixed, then with `[]graphalign.Fixed{{SubjectID: robnis.ID, HandleID: h}}`, and assert AC-2.1 and AC-2.2 with `rowFor` and `comparisonOutcome`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestProposeHoldsAnUnshortlistedHandle -v` passes all three cases.

**STEP-2 — Pin propose, evidence and Apply over the FFI**
Files: `api/ffi/handlers/promote_align_test.go`, `promote_evidence_test.go`, `promote_batch_test.go`. Rule: `.cursor/rules/go-tests.mdc` (D-12).
Do: in `promote_align_test.go`, move `personOnNewSource`'s body into `namedPersonOnNewSource` *(new)* with a `name string` parameter after `typeID`, passed to `namevaluetest.Western(name)`, and the Source title still `"Register " + polarity`; `personOnNewSource` returns `namedPersonOnNewSource(t, dir, userID, typeID, "Mary Smith", sexKey, polarity)`. In `promote_evidence_test.go`, add `newFoundRecordFixture(t) evidenceFixture` *(new)* beside `newEvidenceFixture`: the same `t.Cleanup` and `subjectFixture`; one `Mary Smith` (male, positive) promoted with `promoteOK`, its Subject in `members` and claim in `claims`, its entity in `handle`; then `sourceID, incoming` from `namedPersonOnNewSource(…, "Polly Jones", "male", observations.PolarityPositive)`, and `graphID` = `soleGraphID` of that Source. Add `TestProposeHoldsAFoundRecord` (T-2) in `promote_align_test.go`, `TestGetPromoteClaimEvidenceForAFoundRecord` (T-3) in `promote_evidence_test.go`, and the subtest `files a found record` (T-4) at the end of `TestApplyPromoteGraphAlignment`. T-4 writes its calls inline (the `apply` closure belongs to another subtest): `mustRPC` a propose with `Fixed: [{incoming, handle}]` to read the revision, then `rpc` the Apply.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run 'TestProposeHoldsAFoundRecord|TestGetPromoteClaimEvidence|TestApplyPromoteGraphAlignment|TestProposeSendsContested|TestProposePromoteGraphAlignment' -v` passes.

**STEP-3 — `PromoteFlow.hold` and `select`**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: extract `select(_:at:)` from `setTarget` (D-4), then add `headerParts(_:)` (D-5) and `hold` (D-1 – D-3) per the sketch. Add T-5 – T-9.
Check: XT with `PromoteFlowTests` passes, old and new.

**STEP-4 — `PromoteModel.hold`, and the fake as it is**
Files: `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: add `proposeAfterDecision()` and `hold(subjectID:id:ref:header:)` per D-6; `setTarget` calls the helper. Add T-10. Do not edit `FakeStore.swift` (D-7).
Check: XT with `PromoteModelTests` passes, old and new; `git diff --stat -- macos/App/Platform/FakeStore.swift` is empty.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-11/promote-find-another-record.md`, `deployment-plan.md` (S11-35 map row and scope entry), `completed.md`, this plan.
Do, in the feature doc:
- After the "What the engine already does" table, replace the paragraph starting "S11-35 pins this with tests" with one that names T-1 – T-4 by test name and says no Go code changed.
- Under "Where the Swift side changes", first bullet: replace "a flow call that takes the record (id, ref, sorted and reading titles, subtitle)" with "a flow call, `hold(subjectID:id:ref:header:)`, that takes the record's id, ref and list header and titles it as the menu titles every record".
- Same section, third bullet: replace "`FakeStore`'s propose must hold a fixed handle that is not in its candidates, for the Swift tests and previews." with "`FakeStore`'s unscripted propose already holds a fixed handle of any filed record (it has no candidates); `PromoteModelTests` pins it."

Do, in `deployment-plan.md`: in the S11-35 scope entry, replace "`FakeStore`'s propose holds a fixed handle outside its candidates (today it finds entities only among memberships)" with "`FakeStore`'s unscripted propose already holds a fixed handle of any filed record (it has no candidates); a model test pins it". Then the Definition of done (the map row's Plan link is already set by planning).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-35-hold-found-record.md` reports no errors and no map-link warning; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Hold a record Promote did not offer as a decided row."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeHoldsAnUnshortlistedHandle` *(new)* | Robins filed, Robnis incoming; no sex / same sex / conflicting sex → unfixed row neither targets nor lists the handle; fixed row `TargetHandle` on it, `ReasonDecided`, `ConflictWithFixed` false, a `name` comparison, `sex_at_birth` `OutcomeUnknown` / `OutcomeAgree` / `OutcomeConflict`, `Contested` false / false / true | AC-2.1, AC-2.2, D-9, D-10 |
| T-2 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeHoldsAFoundRecord` *(new)* | `newFoundRecordFixture`; unfixed → the incoming row's `handle_id` is not `f.handle` and no alternative is; fixed → `target` `handle`, `handle_id` `f.handle`, `reason` `decided`, `assessment` `none`, `contested` true | AC-2.1, AC-2.3 |
| T-3 | Go | `api/ffi/handlers/promote_evidence_test.go` `TestGetPromoteClaimEvidenceForAFoundRecord` *(new)* | `f.evidence(t)` → `sectionOf(resp, "name")` and `sectionOf(resp, "sex_at_birth")` non-nil; `len(resp.GetOffers()) > 0` | AC-2.4, D-11 |
| T-4 | Go | `api/ffi/handlers/promote_batch_test.go` `TestApplyPromoteGraphAlignment` / `files a found record` *(new)* | `newFoundRecordFixture`; propose fixed, Apply `{f.incoming, "handle", f.handle, DefaultPins: true}` at its revision → one written row, entity id `f.handle`, `GetPins() > 0` | AC-2.5 |
| T-5 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aFoundRecordIsHeldAndJoinsTheMenu` *(new)* | Flow with Polly's `nameOrders`; `flow.load(entryID: "a", proposal: proposal(rows: [a]), subjects: [fact("a")], bridges: [])` where `a = row(id: "a", target: "new", handleID: "", assessment: "none")` with `reason = .noMatch`, `alternatives = []`; `mapRest()`; `hold(subjectID: "a", id: "e7", ref: "PER-7", header: Polly)` → AC-1.1, AC-1.2. A second flow, same start, `header: nil` → AC-1.9 | AC-1.1, AC-1.2, AC-1.9, D-3, D-5 |
| T-6 | Swift | same file, `aFoundRecordSurvivesAReproposalWithItsBandRefreshed` *(new)* | T-5's start with rows `a` and `row(id: "b")`, subjects `[fact("a"), fact("b")]`; after `hold`, `flow.merge(proposal(rows: [heldA, row(id: "b")], revision: 4), subjects:)` with `heldA = row(id: "a", handleID: "e7", assessment: "weak")`, `handleRef = "PER-7"`, `reason = .decided`, `contested = true` → AC-1.3; `setTarget(subjectID: "b", token: "handle:e2")`, `merge(proposal(rows: [newA, row(id: "b")], revision: 5), subjects:)` with `newA = row(id: "a", target: "new", handleID: "", assessment: "none")` → AC-1.4 | AC-1.3, AC-1.4 |
| T-7 | Swift | same file, `aFoundRecordAlreadyInTheMenuIsPickedNotAddedAgain` *(new)* | `load` with `row(id: "a")` (menu `e1`, `e2`); `hold(subjectID: "a", id: "e2", ref: "PER-2", header: nil)` → AC-1.5 | AC-1.5, D-3 |
| T-8 | Swift | same file, `holdingAFoundRecordResetsTheOldTargetsEvidence` and `aFoundRecordSharingAHandleNamesTheOtherRow` *(new)* | `load` with `row(id: "a")` on `e1`, `setEvidence(subjectID: "a", handleID: "e1", board:)`, then `hold` `e7` → AC-1.6. `load` with the T-5 row `a` and `row(id: "b")` (on `e1`, ref `PER-1`), `mapRest()`; `hold(subjectID: "a", id: "e1", ref: "PER-1", header: nil)` → AC-1.7 | AC-1.6, AC-1.7, D-4 |
| T-9 | Swift | same file, `anAnchorOrMissingRowOrEmptyIDHoldsNothing` *(new)* | Rows `a` (anchor, via `fact("a", anchor: true)`) and `b`; `hold` on `a`, on `"zz"`, and on `b` with id `""` → false each; `flow.rows` equal before and after | AC-1.8, D-2 |
| T-10 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `aFoundRecordIsProposedHeldAndFiled` *(new)* | `store(subjects: [subject("a", ref: "CPR-A", label: "Ada")])`, no `promoteProposals` (the first proposal gives Ada `skip` / `none`, so the row opens unset); `store.membershipBySubject["x"] = CatalogSubjectMembership(subjectID: "x", claimID: "claim-x", entity: CatalogCanonicalEntity(id: "e-far", ref: "PER-FAR", subjectTypeID: "type-person", label: "Polly Jones"), kind: "person")`; `load`, `mapRest`, `hold(subjectID: "a", id: "e-far", ref: "PER-FAR", header: nil)`, `waitUntil { !model.isProposing }` → AC-3.1's fixed pair, row and menu; `done()`, `waitUntil { store.lastPromoteBatch != nil }` → `a` / `handle` / `e-far` | AC-3.1, D-6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteFlowTests` (the selection body moved), `PromoteModelTests` (the `setTarget` follow-up moved), `TestProposeSendsContested` and every user of `personOnNewSource` (its body moved).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; no Go production file and no FakeStore.swift line changed
- [ ] No new user-facing copy
- [ ] Version bump — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the S11-35 scope entry's FakeStore sentence (STEP-5); completed.md: one short paragraph of what shipped
- [ ] promote-find-another-record.md: the engine tests named, the call-shape and FakeStore bullets corrected (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None — nothing on the page calls `hold` until S11-36 adds the menu entry. The slice Check for this PR is T-5 – T-10 and the Go tests T-1 – T-4.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S11-47 or S11-34 lands first and moves `makeRow`, `draftTarget`, `setTarget` or the model's `setTarget` | Step 2 of the Implementer block: compare the §4 rows, stop on a change |
| A later registry change moves Robnis's band | D-10: T-1 pins no band |
| `personOnNewSource`'s move changes a fixture's Source title or name | STEP-2's check reruns its callers |

**Stop and report if:**

- T-1 – T-4 fail for a reason other than a mistake in the test: a fixed off-shortlist handle is refused, not loaded (no `name` comparison), not contested on a sex conflict, has no evidence, or does not file. That is the engine gap the working rule says S11-35 fixes; report the failing assertion and the function before writing Go code.
- The unfixed Robnis or Polly Jones row targets or lists the filed handle (the pair is no longer off-shortlist).
- `makeRow`'s decided branch no longer keeps `previous.menu`, or no longer refreshes the held item's band.
- `conclusionTitle` or `conclusionSubtitle` no longer take person / event / place headers with a fallback.
- T-10 needs a `FakeStore.swift` change to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S11-35 — Hold a record Promote did not offer
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names); deviations from the plan; manual check (none, per §10); `Plan: docs/deployment-plan/spike-11/prs/S11-35-hold-found-record.md`.

## Landed

The code matches the plan. One difference: the Swift tests share one small private helper, `unmatched(_:)`, for the New / no-band / no-alternatives row that T-5, T-6 and T-8 start from, where the plan listed no new helper.
