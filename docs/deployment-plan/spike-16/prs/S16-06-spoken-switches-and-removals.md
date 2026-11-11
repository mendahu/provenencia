# S16-06 — Section switches and removals are spoken

**Status:** Landed in [#785](https://github.com/mendahu/provenencia/pull/785) (2026-11-11)
**Spike / slice:** Spike 16 · Slice 1c — Keyboard and VoiceOver
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the copy changes sit in slots that exist today, and the focus ring on the selected segment is `PVChip`'s existing ring; user, 2026-11-11)
**Board:** —
**Depends on:** S16-03 (landed [#769](https://github.com/mendahu/provenencia/pull/769))
**Blocks:** —
**Branch:** `s16-06-spoken-switches-and-removals`
**Version bump:** patch — a fix: a section switch and a Remove that keeps the person are spoken, and the header and the confirm use the page's word; no catalog change: 0.8.1 → 0.8.2, or the next patch on `main` ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (STEP-1, a kit component's behavior), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-11, P-12, P-16, P-17; working rules "(Batch 1) Every outcome on the page is spoken", "(Batch 1) A person's records are records", "(Batch 1) One open section, one list of reads"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) "One sectioned detail page"
**Planned:** 2026-11-11 against `main` @ `0cd3700`. The Swift steps need a Mac with Xcode (`xcodebuild`); the doc checks run anywhere
**Review:** [`S16-06-spoken-switches-and-removals.review.md`](S16-06-spoken-switches-and-removals.review.md) — Fits — Ready with changes, 2 rounds; no open blocker; To user: none

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

## Landed

- The version bump is 0.9.0 → 0.9.1 (main was at 0.9.0).
- Round 1 (UX-2, user decision 2026-11-11): a `PVSegmentedControl` with a `focus:` handle is one Tab stop and ←/→ move the selection, not one Tab stop per segment; the Out line, §6 States and manual check step 2 changed to match (F-1).
- Round 3 (user request 2026-11-11, F-5): the auto-reconciler's provisional outcome reads "provisional record" (was "provisional member"), so the page keeps one word for a person's records.
- Round 4 (user request 2026-11-11, F-6): Vitals gets a section crumb like the other sections. `WorkspaceLocationLabels.labels(for:names:)` reads `openPageSection`, so the crumbs and the Back / Forward menu name the first section too; the record's crumb is inert on it.

## 1. Objective

A researcher on VoiceOver or Full Keyboard Access keeps their place on the Person page. Choosing a section with ⌘1–⌘5 (or a click, or Space on a segment) says the section's name ("Timeline shown.") and, when the keyboard was in the old section or on the switcher, leaves it on the selected segment instead of dropping it to the window. Removing one of a person's records that keeps the person says so ("J. Weber unpromoted."). The leave-or-stay choice after a Remove moves out of the view into the shared Unpromote flow, where the tests reach it. The page uses one word for a person's records: the header counts "7 records" on every Conclusion page, and the Unpromote confirm ends ", and the person stays".

**Why now.** Slice 1c of Spike 16's Batch 1; the Person page audit (P-11, P-12, P-16, P-17) and the working rule "Every outcome on the page is spoken". S16-03 landed the one open section this PR reads.

**What stays.** No Toast on a Remove (S13-17 RC-6). The delete branch's leave and its "PER-8JD2L deleted." announcement (the string moves namespace, the words stay). Event and Place pages offer no sections, so only their header count changes. The other three `PVSegmentedControl` callers pass no focus handle and look and behave as today. The Evidence graph's Unpromote writes through `UnpromoteFlow.confirm(session:userID:)` as today.

## 2. User stories

**US-1.** As a researcher using the keyboard and VoiceOver, when I choose a Person page section, I hear which section opened and the keyboard stays on the section switcher, so that I do not lose my place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person page on Family; a Family tile has keyboard focus; press ⌘3 | VoiceOver says "Timeline shown."; keyboard focus is on the Timeline segment; the next Tab goes into the Timeline section |
| AC-1.2 | Person page on Vitals; keyboard focus on the Vitals segment (Tab into the switcher); press ⌘4 | "Gallery shown." is spoken; focus is on the Gallery segment |
| AC-1.3 | `PVSegmentedControl` with a focus handle, items [A, B (disabled), C], selection C | The handle is on C. Selection B (disabled) → the handle is on A. Every item disabled → no segment carries it |
| AC-1.4 | Person page on Vitals; focus in the omnibar field; press ⌘2 | "Family shown." is spoken; focus stays in the omnibar |
| AC-1.5 | Person page on Timeline; press ⌘3 (the open section) | Nothing is spoken; focus does not move |
| AC-1.6 | Open a Person page from the Persons list, or press Back from Timeline to Family | Nothing is spoken (only a switch chosen on the switcher speaks) |

**US-2.** As a researcher using VoiceOver, when I unpromote one of a person's records and the person keeps another, I hear that the record was unpromoted, so that the Remove does not sound like nothing happened.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Person PER-8JD2L (h-maria) with records J. Weber and M. Weber; Unpromote J. Weber, confirm | The page stays on Sources; "J. Weber unpromoted." is spoken; the session applies the effects without leaving |
| AC-2.2 | Person h-maria with one record J. Weber; preflight lists h-maria (ref PER-8JD2L) as deleted; confirm | The page leaves for the Persons list; "PER-8JD2L deleted." is spoken; the detail key is evicted |
| AC-2.3 | As AC-2.2, but the preflight failed (no listed handles) and the write deleted h-maria | The page leaves for the Persons list; nothing is spoken |
| AC-2.4 | The write deleted another handle (h-otto) but not h-maria | The page stays; "J. Weber unpromoted." is spoken |
| AC-2.5 | The write fails (`promote.stale`) | The page stays; the confirm keeps its error; nothing is spoken |

**US-3.** As a researcher, I read one word for a person's records on the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Header of a Person, Event or Place with 4 records; with 1 | "4 records"; "1 record" |
| AC-3.2 | Unpromote confirm of a Person record whose person keeps another record | Message: "This promotion is removed. The cards stay on the graph, and the person stays." |
| AC-3.3 | The same confirm for an Event record; for a Place record (the Evidence graph's confirm) | "This promotion is removed. The cards stay on the graph, and the event stays."; "This promotion is removed. The cards stay on the graph, and the place stays." |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVSegmentedControl`'s `focus:` handle lands on the selected segment when it is enabled, else on the first enabled one; its doc comment and the kit README row say so (P-11) | STEP-1, AC-1.3 |
| The Conclusion page's switcher passes a handle; a switch chosen on it announces "{Section} shown." and claims the selected segment when the keyboard was in the sections column or on the switcher (P-11) | STEP-2, STEP-3, AC-1.1, AC-1.2, AC-1.4, AC-1.5, AC-1.6 |
| The leave-or-stay decision after a Remove is `UnpromotePageOutcome` *(new)*, applied by `UnpromoteFlow.confirm(session:userID:leavingPage:announce:)` *(new)*; the Sources section calls it; the stay branch announces "{record} unpromoted." (P-12, P-16) | STEP-4, STEP-5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| The public `apply:` hook of `UnpromoteFlow.confirm` becomes private, so a page leaves only through the new method (P-16) | STEP-4 |
| `UnpromoteFlowTests.confirmReportsTheEffects`' inline copy of the decision goes; both branches and the failure are tested through the new method against a `WorkspaceNavigation` (P-16) | STEP-4, T-4, T-5, T-6 |
| `conclusions.memberCount` reads "%lld records" / "%lld record" (P-17) | STEP-6, AC-3.1 |
| The Unpromote confirm's keep message is per kind and ends ", and the person stays" (", and the event stays", ", and the place stays") (P-17) | STEP-6, AC-3.2, AC-3.3 |
| `macos-client-patterns.md` "One sectioned detail page" says what a switch speaks and where focus goes | STEP-7 |

| Out | Home |
| --- | --- |
| Show more focus in the sections | S16-07 (landed #761) |
| Back restoring scroll, folds and focus | P-19, [`docs/ideas/person-page-followups.md`](../../../ideas/person-page-followups.md) |
| Announcing a section reached by Back, Forward or a link (AC-1.6) | not planned: the Back menu names its target (`conclusions.section.leaf`), and a link opens a new page, which VoiceOver reads |
| An announcement for a Remove on the Evidence graph's own Unpromote | not planned: the graph shows the card leave, and P-12 names the Sources section only |
| A Tab order between segments (`next` / `previous` neighbours) | ~~not planned: the control leaves Tab between segments to the key loop, as its doc comment says~~ superseded by the user on 2026-11-11 in round 1 of the ship log (UX-2): one Tab stop, arrows move the selection |
| Focus handles for the other three `PVSegmentedControl` callers | not planned: no audit finding names them |
| Rewording "handle" or "members" in docs, the audit, archives and code comments | not planned: P-17 is about copy the researcher sees |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Segmented control focus | `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` `PVSegmentedControl.segment(_:)` | With a `focus:` handle, applies it to `items.first(where: { !$0.isDisabled })`, not the selected segment; every other enabled segment is `.focusable()` with no binding. No caller passes a handle today (`PropertiesView.swift`, `DateValueEditorForm.swift`, `CitationComposerAuthoringBox.swift`, `ConclusionDetailPage.swift` all omit it) |
| Kit README | `macos/App/DesignSystem/README.md` row "SegmentedControl" | Describes the `focus:` handle's Tab stops, not where the handle lands |
| Switcher | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailHeaderBand.switcher(selection:)` | Builds a `PVSegmentedControl` with ⌘1–⌘5 shortcuts, no disabled item, and no focus handle, so it adds no Tab stop. `select` is passed straight through |
| Section select | `ConclusionDetailPage.swift` `ConclusionDetailLoaded.body` | `select: { navigation.go(to: navigation.currentLocation.withPageSection($0)) }`. Nothing is spoken. The primary column (`pageSections(…)` in a `VStack`) is replaced on a switch, and the focused stop inside it goes with it; the header band is in `PVPageColumns`' header and stays |
| Navigation | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `go(to:)`, `fallbackToSectionRoot()` | `go(to:)` holds only through `leaveGuard`, which only `CitationComposerView.swift` and `PromoteView.swift` set; no guard is mounted beside a Conclusion page, so a switch applies synchronously. With no `attachProject`, `history` is nil and `fallbackToSectionRoot()` applies `.sectionRoot(currentLocation.section)` directly |
| Section label | `macos/App/Features/Workspace/ConclusionPageSection.swift` `ConclusionPageSection.label` | `LocalizedStringResource` per section: Vitals, Family, Timeline, Gallery, Sources |
| Remove confirm | `macos/App/Features/Conclusions/PersonSourcesSection.swift` `PersonSourcesSection.confirm()` | Reads the deleted ref from `unpromote.pending`, then calls `unpromote.confirm(session:userID:apply:)` with a closure: when `effects.deletedHandles` holds `entityId`, `session.apply(effects, leavingPage: navigation.fallbackToSectionRoot)` and posts `L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:)` when a ref was listed; else `session.apply(effects)` and nothing is spoken. `begin(_:)` opens the flow with `handleID: entityId` |
| Unpromote flow | `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` `UnpromoteFlow.confirm(session:userID:apply:)`, `UnpromoteConfirmItem.message`, `deletesOwnHandle` | `confirm` returns false without writing when `pending` is nil or a write runs; on success it calls `apply` (or `session.apply`), then sets `pending = nil` and returns true; on failure it stores the error in `pending` and returns false. `message` reads the deleted ref as `deletedHandles?.listed.first?.ref` and falls through to `L10n.string(L10n.Unpromote.confirmMessage)` |
| `apply:` callers | `PersonSourcesSection.swift` `confirm()`; `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `confirmReportsTheEffects` | The only two. The Evidence graph (`macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `model.unpromote.confirm(session:userID:)`) and `EvidenceGraphModelTests` pass no `apply` |
| Doc comment | `macos/App/Features/Workspace/Session/CatalogMutation.swift` (line ~119) | Says a leaving page applies effects with `apply(_:leavingPage:)` "(`UnpromoteFlow.confirm`'s `apply`)" |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Unpromote.confirmMessage`, `confirmMessageDeletes(kind:ref:)`, `L10n.Conclusions.memberCount(_:)`, `sourcesPersonDeletedAnnouncement(ref:)`; `macos/App/Resources/Localizable.xcstrings` | `unpromote.confirmMessage` = "This promotion is removed. The cards stay on the graph, and the handle stays."; `conclusions.memberCount` plural "%lld member" / "%lld members"; `conclusions.sources.personDeletedAnnouncement` = "%@ deleted.", used only in `PersonSourcesSection.confirm()` |
| Header count | `macos/App/Features/Conclusions/PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift` `members` | Each sets `members = L10n.Conclusions.memberCount(detail.memberCount)` |
| Preview copy | `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` `#Preview("Unpromote confirm")` and `#Preview("Unpromote confirm, 200 connections")` | Both carry the literal "…, and the handle stays." |
| Announce helpers | `macos/App/Platform/AccessibilityAnnouncement.swift` `AccessibilityAnnouncement.post(_:priority:)`; `macos/App/DesignSystem/Components/Announce/PVAnnounce.swift` `pvAnnounce(_:priority:)` | `post` sends `announcementRequested`; `pvAnnounce` speaks on appear and on a change to a different text |
| Focus helpers | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)`; `macos/App/DesignSystem/Components/Input/PVFieldFocus.swift` `PVFieldFocus.init(_:equals:next:previous:)` | Claim writes on the next main turn and again a turn later; the handle applies `.focused(binding, equals:)` |
| Tests | `UnpromoteFlowTests.swift` `confirmReportsTheEffects`, `messageFollowsThePreflight`, `confirmFailureStaysOpen`; helpers `item(handleID:deletes:failed:)` (name "Maria"), `group(total:listed:via:)`, `makeFixture()` | `confirmReportsTheEffects` rewrites the leave decision in its own `apply:` closure (P-16) |
| Tests | `macos/ProvenenciaTests/L10nResolutionTests.swift` `unpromoteCopyResolves`; `PersonDetailTests.swift` (line ~407); `PlaceDetailTests.swift` (line ~131); `EventDetailTests.swift` (a `members: Int = 4` helper argument, no count assertion) | Expect "…, and the handle stays.", "4 members", "1 member" |
| Tests | `macos/ProvenenciaTests/PVFieldFocusTests.swift` | The kit focus handle's tests; no `PVSegmentedControl` test exists |

**Precedent.** S13-17 (#729) added the delete branch's announcement in `PersonSourcesSection.confirm` with `AccessibilityAnnouncement.post`; copy its spoken-only string shape ("%@ deleted."). `UnpromoteConfirmItem.message` / `confirmMessageDeletes(kind:ref:)` is the per-kind copy to mirror. S16-07 (#761) is the focus precedent: decision in plain data that a test reaches, the write through `FocusClaim`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVSegmentedControl` puts the handle on the selected segment when it is enabled, else on the first enabled segment, else on none. The choice is `static func focusTarget(items: [Item], selection: Value) -> String?` *(new)* returning the item id, which `segment(_:)` compares with `item.id` | The audit's fix (P-11); a static lets a test pin the rule without a view. No caller passes a handle today, so no other screen changes | Keep "first enabled" and claim the selected one from the page — because Tab into the control would land on the wrong segment |
| D-2 | The switcher's handle is `PVFieldFocus($switcherFocus, equals: selection)` over `@FocusState private var switcherFocus: ConclusionPageSection?` *(new)* in `ConclusionDetailHeaderBand`. This holds because the Conclusion switcher has no disabled item, so `focusTarget` is always the selected segment | One state, no new stop type | A one-case enum stop type — because the section already names the stop |
| D-3 | Whether the keyboard is in the page's sections column is a Bool focused value. In `ConclusionDetailPage.swift`, beside `ConclusionDetailPageSectionsContext`: `private struct ConclusionSectionsFocusedKey: FocusedValueKey { typealias Value = Bool }` and `extension FocusedValues { var conclusionSectionsFocused: Bool? { get { self[ConclusionSectionsFocusedKey.self] } set { self[ConclusionSectionsFocusedKey.self] = newValue } } }`. `ConclusionDetailLoaded` adds `.focusedValue(\.conclusionSectionsFocused, true)` to the primary column's `VStack`. The band reads `@FocusedValue(\.conclusionSectionsFocused) private var sectionsFocused` and passes `sectionsFocused == true` | A focused value propagates from the focused stop up its ancestors, so the header knows focus sits inside the column that is about to be replaced, without each section reporting its own `FocusState`. `FocusedValueKey` is SwiftUI API on macOS 14 (`@Entry` for `FocusedValues` needs 15) | Always claiming the switcher on a switch — because ⌘2 typed in the omnibar or sidebar would pull focus out of it (AC-1.4); carrying the open section as the value — because the column and the band always read the same open section, so it adds an input that cannot differ (review M-4); reading `window.firstResponder` — because SwiftUI focus inside one hosting view does not move the first responder |
| D-4 | What a switch does beyond navigating is `struct ConclusionPageSectionSwitch: Equatable` *(new)* in `ConclusionPageSection.swift`: `init?(from: ConclusionPageSection, to: ConclusionPageSection, sectionsFocused: Bool, switcherFocused: Bool)`, nil when `from == to`; `let to: ConclusionPageSection`, `let announcement: String` (`L10n.Conclusions.sectionShownAnnouncement(L10n.string(to.label))`), `let claimsSwitcher: Bool` (= `sectionsFocused \|\| switcherFocused`) | Plain data a unit test reaches (rule "Every outcome on the page is spoken"; S16-07's shape) | Logic inline in the view — because the audit's P-16 lesson is that a decision in a view is untested |
| D-5 | The band wraps `select`: on a press it builds `let sectionSwitch = ConclusionPageSectionSwitch(…)` from `openPageSection`, the chosen section, `sectionsFocused == true` and `switcherFocus != nil`, calls `select`, then, when the local `sectionSwitch` is non-nil and `navigation.currentLocation.openPageSection == sectionSwitch.to` (the band reads `@Environment(WorkspaceNavigation.self)`), posts `AccessibilityAnnouncement.post(sectionSwitch.announcement)` and, when `claimsSwitcher`, `FocusClaim.claim($switcherFocus, sectionSwitch.to)`. Click, Space and ⌘ shortcut all call this wrapper. Focus is read before `select`, while the old column still holds it | No guard is mounted beside a Conclusion page, so `go(to:)` applies at once; the check after `select` still refuses to speak a section that did not open if a guard is ever added (review M-3). Back, Forward and links never call the wrapper, so they speak nothing (AC-1.6) | A `@State` pending switch performed in `.onChange(of: openPageSection)` — because it guards a hold that cannot happen here and adds synced state (review M-3) |
| D-6 | The spoken text is "{Section label} shown." (`conclusions.sectionShownAnnouncement`) | User, 2026-11-11; matches "PER-8JD2L deleted." | "{Section}" alone; "{Section} section" |
| D-7 | The leave-or-stay decision is `enum UnpromotePageOutcome: Equatable { case leave(String?), stay(String) }` (unlabelled, so `.leave("PER-8JD2L deleted.")` reads as in T-3) *(new)* with `static func after(_ effects: CatalogWriteEffects, item: UnpromoteConfirmItem) -> UnpromotePageOutcome`, in `UnpromoteFlow.swift`. Leave when `item.handleID` is non-nil and `effects.deletedHandles` contains it; its announcement is `L10n.Unpromote.deletedAnnouncement(ref:)` of `item.ownHandleDeletedRef` *(new accessor)*, nil when that is nil. Else stay with `L10n.Unpromote.unpromotedAnnouncement(name: item.name)`. `ownHandleDeletedRef` = `deletedHandles?.listed.first { $0.id == handleID }?.ref`, and `message` reads it too in place of `listed.first?.ref` | The page's entity is the item's own handle (`begin(_:)` passes `handleID: entityId`), so one field decides; one accessor finds the ref for both the confirm and the announcement (review M-5). The working rule names this function and its two results | A separate `pageEntityId` argument — because it restates `handleID` |
| D-8 | `@discardableResult func confirm(session: WorkspaceSession, userID: String, leavingPage leave: @escaping @MainActor () -> Void, announce: @escaping @MainActor (String) -> Void = { AccessibilityAnnouncement.post($0) }) async -> Bool` *(new)*. Body: `guard let item = pending else { return false }`; then the private write with an apply closure that switches on `UnpromotePageOutcome.after(effects, item: item)`: `.leave(let text)` → `session.apply(effects, leavingPage: leave)`, then `announce(text)` when non-nil; `.stay(let text)` → `session.apply(effects)`, then `announce(text)`. The closure runs before `pending = nil`, as `apply` does today. A failed write calls neither `leave` nor `announce`. `PersonSourcesSection.confirm()` becomes `await unpromote.confirm(session: session, userID: userID, leavingPage: navigation.fallbackToSectionRoot)` | Moves the applying out of the view, so a test drives the real branch against a `WorkspaceNavigation` (P-16) and no test rewrites it; `leavingPage:` mirrors `session.apply(_:leavingPage:)`. `announce` is injected so a test reads what was said | Returning the outcome and announcing in the view (review M-8, nit) — because the view would again hold the post |
| D-9 | `confirm(session:userID:apply:)` becomes `private func write(session:userID:apply:)`; a public `@discardableResult func confirm(session:userID:) async -> Bool` *(new overload, same call shape the Evidence graph uses today through the defaulted `apply:`)* calls `write` with `session.apply(effects)` | After this PR its `apply:` has no caller but the new method; one public way for a page to leave (review M-1) | Keeping the public `apply:` hook — because the next page could rebuild the branch inline, the P-16 bug |
| D-10 | The two spoken Remove strings live in `L10n.Unpromote`: `unpromotedAnnouncement(name:)` *(new)* and `deletedAnnouncement(ref:)` *(new, moved)*. `conclusions.sources.personDeletedAnnouncement` and `L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:)` go | The decision moves into the shared flow under `Features/Catalog`, which speaks only through `L10n.Unpromote` (review M-2) | Keeping them under `L10n.Conclusions` — because the flow would read one page's namespace |
| D-11 | The stay announcement names the record (`item.name`, "J. Weber"), not the person | The audit's string "{name} unpromoted."; the record is what left the list | The person's name — because the person did not change |
| D-12 | `conclusions.memberCount` keeps its key and Swift name; its values become "%lld record" (one) / "%lld records" (other) and its comment says records | Rule "A person's records are records": one string on every Conclusion page | A new key `conclusions.recordCount` — because the old key would linger |
| D-13 | `L10n.Unpromote.confirmMessage` (a `LocalizedStringResource`) becomes `static func confirmMessage(kind: EvidencePrimaryKind) -> String` with keys `unpromote.confirmMessage.person` / `.event` / `.place` *(new)*, shaped like `confirmMessageDeletes(kind:ref:)`; the key `unpromote.confirmMessage` leaves the catalog | Rule: "worded per kind like `confirmMessageDeletes`, the Evidence graph's confirm included" | Interpolating the kind's noun — because the catalog's other per-kind lines are whole sentences |
| D-14 | Tests: Swift Testing in `macos/ProvenenciaTests`; no UI test and no window. Focus landing and VoiceOver speech are checked by hand (§10) | Batch 3 rule "A test run shows no window" | An XCUITest for ⌘3 — because it needs a shown window |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVSegmentedControl.focusTarget(items:selection:)` *(new)* | func | `PVSegmentedControl.segment(_:)`'s inline `items.first(where:)` | Extends it: the inline expression moves into the static and changes rule |
| `ConclusionPageSectionSwitch` *(new)* | type | `ConclusionPageSection` (label), `FocusKeeper` (focus rules as plain data) | `FocusKeeper` keeps a stop across width changes and reloads; a switch is a one-shot decision with an announcement |
| `ConclusionSectionsFocusedKey`, `FocusedValues.conclusionSectionsFocused` *(new)* | key | none: `rg "focusedValue\|FocusedValue" macos/App` finds nothing | First use; one key, set in one place |
| Posting with `AccessibilityAnnouncement.post` in the band | call | `pvAnnounce(_:priority:)` | `pvAnnounce` speaks on appear (a page opened on Timeline would say "Timeline shown.", against AC-1.6) and does not repeat the same text, so Timeline → Family → Timeline would be silent the second time; a switch is an event, not a state (review M-6) |
| `UnpromotePageOutcome` *(new)* | type | `PersonSourcesSection.confirm()`'s inline branch | Replaces it |
| `UnpromoteConfirmItem.ownHandleDeletedRef` *(new)* | accessor | `UnpromoteConfirmItem.message`'s `listed.first?.ref` | `message` moves onto it |
| `UnpromoteFlow.confirm(session:userID:leavingPage:announce:)` *(new)* | func | `UnpromoteFlow.confirm(session:userID:apply:)` | That one becomes the private `write`; the graph keeps `confirm(session:userID:)` |
| `L10n.Conclusions.sectionShownAnnouncement(_:)` *(new)* | key | `L10n.Conclusions.pageSectionLeaf(_:_:)` | Different sentence |
| `L10n.Unpromote.unpromotedAnnouncement(name:)`, `deletedAnnouncement(ref:)` *(new / moved)* | key | `L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:)` | Moved into the flow's namespace (D-10) |
| `L10n.Unpromote.confirmMessage(kind:)` *(new form)* | key | `L10n.Unpromote.confirmMessageDeletes(kind:ref:)` | Mirrors it |
| `macos/ProvenenciaTests/PVSegmentedControlTests.swift` *(new)* | test file | `PVFieldFocusTests.swift` | One file per kit component, as the other `PV…Tests` |
| `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift` *(new)* | test file | `WorkspaceNavigationTests.swift` | Navigation tests cover locations, not the switch's speech and focus |

**Builds on.** The kit's one focus handle (`PVFieldFocus`), `FocusClaim` as the one writer, `AccessibilityAnnouncement.post`, `UnpromoteFlow` as the one Unpromote flow, per-kind copy as in `confirmMessageDeletes`.
**Deletes.** `PersonSourcesSection.confirm()`'s inline leave-or-stay branch; the public `apply:` parameter of `UnpromoteFlow.confirm` (now the private `write`); `UnpromoteFlowTests.confirmReportsTheEffects` (its busy and evict assertions move to T-4); `L10n.Unpromote.confirmMessage` as a `LocalizedStringResource` and the key `unpromote.confirmMessage`; `L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:)` and the key `conclusions.sources.personDeletedAnnouncement`; the "handle stays" literal in both `PVUnpromoteConfirm.swift` previews. `CatalogMutation.swift`'s doc comment names `UnpromoteFlow.confirm(session:userID:leavingPage:announce:)` in place of "`UnpromoteFlow.confirm`'s `apply`".
**Next adds.** A Conclusion kind that gains sections (Event, Place) gets the spoken switch and the focus landing from the shared header with no code; another page that offers Unpromote and can leave calls `confirm(session:userID:leavingPage:)`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The switcher becomes a Tab stop (one segment carries the handle, every enabled segment is focusable) | `ConclusionDetailHeaderBand` | The band is inside the page, which is keyed by entity, so `switcherFocus` starts nil on another person. No section is missing: `offered(by:)` is static |
| Focus written to a segment by `FocusClaim` | `PVSegmentedControl` | The selected segment exists on the next turn (the band is not replaced on a switch); `FocusClaim` writes twice for a late view |
| A Remove whose preflight failed but the write deleted the person | `UnpromotePageOutcome.after` | Leave with no announcement (AC-2.3), as today |

### Scale

None — no write, list or loop changes. The switcher has at most five items.

### Edges

- **Assumptions.** `openPageSection` changes synchronously inside `go(to:)` when no `leaveGuard` is set (§4 Navigation); the band stays mounted across a switch (it sits in `PVPageColumns`' header). T-2 holds the switch rule; the manual check holds the mount and the timing.
- **Neighbors.** Same section chosen again → nil switch, nothing spoken (AC-1.5, T-2). Focus in the header's Try again or lead menu → neither in the column nor on the switcher → no claim (T-2 "focus elsewhere"). Only the selected segment is in the key loop (with or without Full Keyboard Access), so Space, ←/→ or ⌘1–⌘5 on the switcher always start from the bound segment and `switcherFocus != nil` claims the new one (ship log round 1, UX-2 decision; round 2, F-4). A disabled selected segment → handle on the first enabled (T-1); the Conclusion switcher has none.
- **Removed inputs.** A Remove that also deleted another handle but not the page's → stay (AC-2.4, T-3). A Remove whose item has no `handleID` (the Evidence graph opens some with nil) → stay; the graph does not call this method.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — no list read from data the app did not write.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section switcher with a focus handle | `PVSegmentedControl` (kit, changed rule D-1) | `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` | Doc comment and README row: "A `focus:` handle lands on the selected segment, or the first enabled one when the selected one is disabled" |
| Switch speech and focus | `ConclusionPageSectionSwitch` *(new)* | `macos/App/Features/Workspace/ConclusionPageSection.swift` | Plain data |
| Header band wiring | `ConclusionDetailHeaderBand` | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` | `@FocusState switcherFocus`, `@FocusedValue sectionsFocused`, `@Environment(WorkspaceNavigation.self)`, the `select` wrapper (D-5) |
| Sections-focused key | `FocusedValues.conclusionSectionsFocused` *(new)* | `ConclusionDetailPage.swift`, beside `ConclusionDetailPageSectionsContext` | D-3 |
| Remove outcome | `UnpromotePageOutcome`, `UnpromoteFlow.confirm(session:userID:leavingPage:announce:)` *(new)* | `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` | |

States: nothing new is drawn. Accessibility: the switcher gains Tab stops at default macOS settings (~~one per enabled segment, as the control documents~~ superseded by the user on 2026-11-11 in round 1 of the ship log (UX-2): one Tab stop, arrows move the selection); ⌘1–⌘5 keep working with focus anywhere.

### Strings

L10n keys *(new)*:
- `L10n.Conclusions.sectionShownAnnouncement(_ section: String)` → `conclusions.sectionShownAnnouncement` = "%@ shown." — comment: "VoiceOver announcement after a record page's section switcher opens a section; the argument is the section's name".
- `L10n.Unpromote.unpromotedAnnouncement(name:)` → `unpromote.unpromotedAnnouncement` = "%@ unpromoted." — comment: "VoiceOver announcement after a record is unpromoted and its person, event or place stays; the argument is the record's name".
- `L10n.Unpromote.deletedAnnouncement(ref:)` → `unpromote.deletedAnnouncement` = "%@ deleted." — comment: "VoiceOver announcement after a record's Unpromote deleted the person, event or place the page showed; the argument is its ref".
- `L10n.Unpromote.confirmMessage(kind:)` → `unpromote.confirmMessage.person` = "This promotion is removed. The cards stay on the graph, and the person stays."; `unpromote.confirmMessage.event` = "This promotion is removed. The cards stay on the graph, and the event stays."; `unpromote.confirmMessage.place` = "This promotion is removed. The cards stay on the graph, and the place stays." — comment per kind: "Message of the Unpromote confirm when the {person | event | place} keeps another record".

Changed: `conclusions.memberCount` = "%lld record" / "%lld records"; comment "Page header: how many records (Identity Claims) this person, event or place has; argument is the count".

Removed: `unpromote.confirmMessage`, `conclusions.sources.personDeletedAnnouncement`.

Old-scope copy check (app and tests only; docs, the audit, archives and code comments are Out): `rg -i 'and the handle stays|%lld members?"|[0-9] members?"' macos/App macos/ProvenenciaTests` finds nothing after STEP-6.

## 7. Steps

**STEP-1 — The segmented control's handle lands on the selected segment**
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/DesignSystem/README.md` (SegmentedControl row), `macos/ProvenenciaTests/PVSegmentedControlTests.swift` *(new)*. Skill: `add-ui-component`, `add-swift-test`.
Do: add `focusTarget(items:selection:)` (D-1); `segment(_:)` compares `item.id` with it; update the doc comment and the README row. Write T-1.
Check: `xcodebuild test -scheme Provenencia -only-testing:ProvenenciaTests/PVSegmentedControlTests` passes.

**STEP-2 — The switch value**
Files: `macos/App/Features/Workspace/ConclusionPageSection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift` *(new)*. Skill: `add-localized-string`.
Do: add `conclusions.sectionShownAnnouncement` and `ConclusionPageSectionSwitch` (D-4, D-6). Write T-2.
Check: `-only-testing:ProvenenciaTests/ConclusionPageSectionSwitchTests` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The switcher speaks and keeps focus**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`.
Do: add the focused value key and set it on the primary column's `VStack` (D-3). In `ConclusionDetailHeaderBand`: `@FocusState switcherFocus`, `@FocusedValue sectionsFocused`, `@Environment(WorkspaceNavigation.self) navigation`; pass `focus: PVFieldFocus($switcherFocus, equals: selection)` (D-2) and the wrapped `select` (D-5).
Check: the app builds with no Swift warning; manual check steps 1–3 (§10) pass. (Step 4's header copy lands in STEP-6.)

**STEP-4 — The Remove outcome in the flow, and the Sources section calls it**
Files: `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift`, `macos/App/Features/Conclusions/PersonSourcesSection.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift` (doc comment), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift`. Skill: `add-localized-string`.
Do: add `L10n.Unpromote.unpromotedAnnouncement(name:)` and `deletedAnnouncement(ref:)`; add `ownHandleDeletedRef` and move `message` onto it; add `UnpromotePageOutcome` (D-7, D-11) and `confirm(session:userID:leavingPage:announce:)` (D-8); make the `apply:` variant the private `write` (D-9); in the same step, `PersonSourcesSection.confirm()` becomes the one call in D-8 (so the app compiles once `apply:` is private) and its doc comment says the flow decides leave or stay; update the `CatalogMutation.swift` comment. Before deleting `confirmReportsTheEffects`, copy its `CatalogWriteEffects(handles:deleted: [CatalogWriteEffectsDeleted(entity: .handle, ids:)])` shape into T-3 – T-5. Write T-3 – T-6 and delete `confirmReportsTheEffects`.
Check: `-only-testing:ProvenenciaTests/UnpromoteFlowTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests` pass; `rg "func confirm\(" macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` lists only `confirm(session:userID:)` and `confirm(session:userID:leavingPage:announce:)`.

**STEP-5 — The old delete announcement goes**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: remove `L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:)` and its key `conclusions.sources.personDeletedAnnouncement` (D-10). Nothing else changes.
Check: `rg "sourcesPersonDeletedAnnouncement|personDeletedAnnouncement" macos` and `rg "deletedHandles.contains" macos/App/Features/Conclusions` find nothing; the app builds; `python3 scripts/check-localizable-xcstrings.py` passes; manual check steps 5–6 pass.

**STEP-6 — Records, and the person stays**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` (`UnpromoteConfirmItem.message`), `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` (both previews' literal → the person sentence), `macos/ProvenenciaTests/L10nResolutionTests.swift`, `UnpromoteFlowTests.swift`, `PersonDetailTests.swift`, `PlaceDetailTests.swift`, `EventDetailTests.swift`. Skill: `add-localized-string`.
Do: D-12, D-13. `message` returns `L10n.Unpromote.confirmMessage(kind: kind)` on its last line. Update T-7, T-8.
Check: `-only-testing:ProvenenciaTests/L10nResolutionTests -only-testing:ProvenenciaTests/UnpromoteFlowTests -only-testing:ProvenenciaTests/PersonDetailTests -only-testing:ProvenenciaTests/PlaceDetailTests -only-testing:ProvenenciaTests/EventDetailTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests -only-testing:ProvenenciaTests/PVDeleteImpactTests` pass; the §6 Strings `rg` finds nothing.

**STEP-7 — Docs and version**
Files: `docs/macos-client-patterns.md` ("One sectioned detail page"), `VERSION`, `core/version.go`, `MARKETING_VERSION` in the Xcode project, `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status, Landed note).
Do: in the Cache contract's item 2, the sentence that says the Person page's Unpromote leaves "through `UnpromoteFlow.confirm`'s `apply`" names `UnpromoteFlow.confirm(session:userID:leavingPage:)` instead. Add one sentence to "One sectioned detail page": the switcher carries a focus handle on the selected segment; a switch chosen on it says "{Section} shown." and keeps the keyboard on the switcher when it was in the sections column or on the switcher (`ConclusionPageSectionSwitch`). Bump the patch. Per the Definition of done.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVSegmentedControlTests.swift` `focusLandsOnTheSelectedSegment` *(new)* | items with ids a, b, c: sel c → "c"; b disabled, sel b → "a"; a disabled, sel a → "b"; all disabled → nil | AC-1.3, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift` `switchSpeaksAndClaims` *(new)* | family→timeline, sectionsFocused → announcement "Timeline shown.", claims; vitals→gallery, switcherFocused → "Gallery shown.", claims; vitals→family, neither → "Family shown.", no claim; timeline→timeline → nil | AC-1.1, AC-1.2, AC-1.4, AC-1.5, D-4, D-6 |
| T-3 | Swift | `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `pageOutcomeLeavesOnlyWhenTheOwnHandleIsDeleted` *(new)* | Built with `item(deletes:)` and `group(total:listed:)` (name "Maria", handleID "h-maria"): deleted [h-maria], `group(total: 1, listed: [("h-maria", "PER-8JD2L")])` → `.leave("PER-8JD2L deleted.")`; deleted [h-maria], `item()` (no deletes listed) → `.leave(nil)`; deleted [h-otto] → `.stay("Maria unpromoted.")`; nothing deleted → `.stay("Maria unpromoted.")`; `item(handleID: nil)`, deleted [h-maria] → `.stay("Maria unpromoted.")` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-7, D-11 |
| T-4 | Swift | `UnpromoteFlowTests.swift` `confirmLeavesWhenThePersonIsDeleted` *(new, replaces `confirmReportsTheEffects`)* | Setup: `makeFixture()`; `let navigation = WorkspaceNavigation()` with no `attachProject` (nil history, intended); `navigation.go(to: .personDetail(entityId: "h-maria", ref: "PER-8JD2L", title: "Maria"))` and `session.apply(location:)` with the same location; warm the detail key; open with `handleID: "h-maria"` and a preflight listing h-maria / PER-8JD2L; `store.nextWriteEffects` deletes h-maria; hold the write and assert `flow.isBusy`; confirm with `leavingPage: navigation.fallbackToSectionRoot`, `announce: { said.append($0) }` → returns true, `navigation.currentLocation == .sectionRoot(.persons)`, detail key evicted, `said == ["PER-8JD2L deleted."]`, `pending == nil`, not busy | AC-2.2, D-8 |
| T-5 | Swift | `UnpromoteFlowTests.swift` `confirmStaysAndSaysSo` *(new)* | Same setup, open as "J. Weber" with no preflight deletes; effects `handles: ["h-maria"]` → returns true, `navigation.currentLocation` still the person detail, `said == ["J. Weber unpromoted."]` | AC-2.1, D-8 |
| T-6 | Swift | `UnpromoteFlowTests.swift` `confirmFailureNeitherLeavesNorSpeaks` *(new)* | Same setup; `store.unpromoteSubjectError` = `CoreInvokeError.coded(status: 1, code: "promote.stale", kind: .user, params: [])` (as `confirmFailureStaysOpen` builds it) → returns false, location unchanged, `said.isEmpty`, `pending?.error` set; on a fresh flow with nothing opened (`pending == nil`) → returns false, location unchanged, `said.isEmpty` | AC-2.5, D-8 |
| T-7 | Swift | `UnpromoteFlowTests.swift` `messageFollowsThePreflight` (updated); `L10nResolutionTests.swift` `unpromoteCopyResolves` (updated) | none deleted, kind .person → the person sentence; `confirmMessage(kind: .event)` / `.place` resolve to their sentences; `unpromotedAnnouncement(name: "J. Weber")` == "J. Weber unpromoted."; `deletedAnnouncement(ref: "PER-8JD2L")` == "PER-8JD2L deleted." | AC-3.2, AC-3.3, D-10, D-13 |
| T-8 | Swift | `PersonDetailTests.swift` (line ~407 updated), `PlaceDetailTests.swift` (line ~131 updated), `EventDetailTests.swift` `headerCountsRecords` *(new)* | Person: members 4 → "4 records", 1 → "1 record"; Place: 1 → "1 record"; Event: the file's detail helper with `members: 4` → "4 records", `members: 1` → "1 record" | AC-3.1, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceGraphModelTests` (reads `UnpromoteConfirmItem.message`), `PVDeleteImpactTests`, `L10nResolutionTests`, `check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md "One sectioned detail page" (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

Sweep names: `L10n.Unpromote.confirmMessage` (resource form), `unpromote.confirmMessage` (the bare key), `sourcesPersonDeletedAnnouncement`, `conclusions.sources.personDeletedAnnouncement`, `confirm(session:userID:apply:)`, "`UnpromoteFlow.confirm`'s `apply`". Old-scope copy: the §6 Strings `rg`, app and tests only. "UI matches the brief inventory": no brief; no new kit primitive.

## 10. Manual check

Turn on VoiceOver (⌘F5) and Full Keyboard Access (System Settings → Keyboard → Keyboard navigation). Use a project with a person who has two records on different Evidence graphs, and a second person with one record.

1. Open the first person on Family. Tab to a family tile. Press ⌘3. Hear "Timeline shown."; the Timeline segment shows its focus ring; Tab moves into the Timeline (AC-1.1).
2. Shift-Tab to the Timeline segment, press ⌘4. Hear "Gallery shown."; focus is on the Gallery segment (AC-1.2). press → to move to Sources: hear "Sources shown."; focus stays on Sources. Press ← three times to reach Family, hearing each section; Tab into the Family tiles, then click the Vitals segment with the mouse: hear "Vitals shown."; the next Tab does not start from the window's first view.
3. Click in the omnibar field, press ⌘2. Hear "Family shown."; the caret stays in the omnibar (AC-1.4). Press ⌘2 again: nothing is spoken (AC-1.5). Press Back (⌘[): nothing is spoken (AC-1.6).
4. The header reads "2 records" (AC-3.1); an Event page and a Place page read "N records".
5. Open Sources. Unpromote one record: the confirm ends ", and the person stays." (AC-3.2). Confirm: hear "{record} unpromoted."; the page stays (AC-2.1).
6. Open the second person on Sources, Unpromote its one record, confirm: the app leaves for the Persons list and says "{ref} deleted." (AC-2.2).
7. On an Evidence graph, Unpromote an Event card whose event keeps another record: the confirm ends ", and the event stays." (AC-3.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `@FocusedValue` reads nil while a stop inside the column has focus (the band is a sibling of the column) | Manual check step 1; Stop condition below |
| The focus move to the next Unpromote button speaks over "J. Weber unpromoted." | The announcement is posted inside the apply closure, before the reload moves focus; manual check step 5 listens for both. Stop condition below |
| The `announce:` default closure needs an explicit actor annotation to compile | Write it as `{ @MainActor in AccessibilityAnnouncement.post($0) }` if the compiler asks; this is not a design change |

**Stop and report if:**

- With a Family tile focused, `@FocusedValue(\.conclusionSectionsFocused)` reads nil at the press (D-3 cannot work as written).
- After a switch, `navigation.currentLocation.openPageSection` has not changed by the line after `select` (a guard or an async path exists; D-5 assumed none).
- After a mouse click on a segment while focus was in the column, the next Tab starts from the window's first view (the click path did not report `sectionsFocused`).
- After ⌘3 from a Family tile, focus is on the Timeline segment but the next Tab does not enter the Timeline section (AC-1.1's second clause).
- VoiceOver does not speak "{record} unpromoted." in manual check step 5 at the default priority.
- `UnpromoteConfirmItem.name` is not the record's display name on the Sources section (the confirm title shows something else).
- A `PVSegmentedControl` caller other than `ConclusionDetailPage.swift` passes `focus:` at the base, or the Conclusion switcher gains a disabled item.
- A caller of `confirm(session:userID:apply:)` other than the two in §4 exists at the base.

## 12. Open questions

None.

## 13. PR

**Title:** S16-06 — Section switches and removals are spoken
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-06-spoken-switches-and-removals.md`.
