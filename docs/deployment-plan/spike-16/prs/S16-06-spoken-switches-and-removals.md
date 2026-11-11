# S16-06 — Section switches and removals are spoken

**Status:** Draft
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
**Review:** — (not reviewed yet)

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

A researcher on VoiceOver or Full Keyboard Access keeps their place on the Person page. Choosing a section with ⌘1–⌘5 (or a click, or Space on a segment) says the section's name ("Timeline shown.") and, when the old section held the keyboard, leaves it on the selected segment instead of dropping it to the window. Removing one of a person's records that keeps the person says so ("J. Weber unpromoted."). The leave-or-stay choice after a Remove moves out of the view into a function the tests reach. The page uses one word for a person's records: the header counts "7 records" on every Conclusion page, and the Unpromote confirm ends "…, and the person stays".

**Why now.** Slice 1c of Spike 16's Batch 1; the Person page audit (P-11, P-12, P-16, P-17) and the working rule "Every outcome on the page is spoken". S16-03 landed the one open section this PR reads.

**What stays.** No Toast on a Remove (S13-17 RC-6). The delete branch's leave and its "PER-8JD2L deleted." announcement. Event and Place pages offer no sections, so only their header count changes. The other three `PVSegmentedControl` callers pass no focus handle and look and behave as today.

## 2. User stories

**US-1.** As a researcher using the keyboard and VoiceOver, when I choose a Person page section, I hear which section opened and the keyboard stays on the section switcher, so that I do not lose my place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person page on Family; a Family tile has keyboard focus; press ⌘3 | VoiceOver says "Timeline shown."; keyboard focus is on the Timeline segment; the next Tab goes into the Timeline section's first stop |
| AC-1.2 | Person page on Vitals; keyboard focus on the Vitals segment; press Space on the Gallery segment after Tab to it | "Gallery shown." is spoken; focus is on the Gallery segment |
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
| AC-2.4 | The write deleted another handle (h-otto, an Event the record took part in) but not h-maria | The page stays; "J. Weber unpromoted." is spoken |

**US-3.** As a researcher, I read one word for a person's records on the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Person page header of a person with 4 records; with 1 | "4 records"; "1 record". Event and Place headers read the same way |
| AC-3.2 | Unpromote confirm of a Person record whose person keeps another record | Message: "This promotion is removed. The cards stay on the graph, and the person stays." |
| AC-3.3 | The same confirm for an Event record; for a Place record (the Evidence graph's confirm) | "…, and the event stays."; "…, and the place stays." |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVSegmentedControl`'s `focus:` handle lands on the selected segment when it is enabled, else on the first enabled one; its doc comment says so (P-11) | STEP-1, AC-1.3 |
| The Conclusion page's switcher passes a handle; a switch chosen on it announces "{Section} shown." and claims the selected segment when the keyboard was in the old section or on the switcher (P-11) | STEP-2, STEP-3, AC-1.1 – AC-1.6 |
| The leave-or-stay decision after a Remove is `UnpromotePageOutcome` *(new)*, applied by `UnpromoteFlow.confirm(onPageOf:…)` *(new)*; the Sources section calls it; the stay branch announces "{record} unpromoted." (P-12, P-16) | STEP-4, STEP-5, AC-2.1 – AC-2.4 |
| `UnpromoteFlowTests.confirmReportsTheEffects`' inline copy of the decision goes; both branches are tested through the new method against a `WorkspaceNavigation` (P-16) | STEP-5, T-4, T-5 |
| `conclusions.memberCount` reads "%lld records" / "%lld record" (P-17) | STEP-6, AC-3.1 |
| The Unpromote confirm's keep message is per kind and ends ", and the person stays" (", and the event stays", ", and the place stays") (P-17) | STEP-6, AC-3.2, AC-3.3 |
| `macos-client-patterns.md` "One sectioned detail page" says what a switch speaks and where focus goes | STEP-7 |

| Out | Home |
| --- | --- |
| Show more focus in the sections | S16-07 (landed #761) |
| Back restoring scroll, folds and focus | P-19, [`docs/ideas/person-page-followups.md`](../../../ideas/person-page-followups.md) |
| Announcing a section reached by Back, Forward or a link (AC-1.6) | not planned: the Back menu names its target (`conclusions.section.leaf`), and a link opens a new page, which VoiceOver reads |
| An announcement for a Remove on the Evidence graph's own Unpromote | not planned: the graph shows the card leave, and it is not a Person page outcome (P-12 names the Sources section only) |
| A Tab order between segments (`next` / `previous` neighbours) | not planned: the control leaves Tab between segments to the key loop, as its doc comment says |
| Focus handles for the other three `PVSegmentedControl` callers | not planned: no audit finding names them |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Segmented control focus | `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` `PVSegmentedControl.segment(_:)` | With a `focus:` handle, applies it to `items.first(where: { !$0.isDisabled })`, not the selected segment. No caller passes a handle today (`PropertiesView.swift`, `DateValueEditorForm.swift`, `CitationComposerAuthoringBox.swift`, `ConclusionDetailPage.swift` all omit it) |
| Switcher | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailHeaderBand.switcher(selection:)` | Builds a `PVSegmentedControl` with ⌘1–⌘5 shortcuts and no focus handle, so it adds no Tab stop. `select` is passed straight through |
| Section select | `ConclusionDetailPage.swift` `ConclusionDetailLoaded.body` | `select: { navigation.go(to: navigation.currentLocation.withPageSection($0)) }`. Nothing is spoken. The primary column (`pageSections(…)` in a `VStack`) is replaced on a switch, and the focused stop inside it goes with it |
| Navigation gate | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `go(to:)` | A different location may be held by `hold(.location(_:))` (a discard gate) and performed later or never, so a press is not yet a switch |
| Section label | `macos/App/Features/Workspace/ConclusionPageSection.swift` `ConclusionPageSection.label` | `LocalizedStringResource` per section: Vitals, Family, Timeline, Gallery, Sources |
| Remove confirm | `macos/App/Features/Conclusions/PersonSourcesSection.swift` `PersonSourcesSection.confirm()` | Reads the deleted ref from `unpromote.pending`, then calls `unpromote.confirm(session:userID:apply:)` with a closure: when `effects.deletedHandles` holds `entityId`, `session.apply(effects, leavingPage: navigation.fallbackToSectionRoot)` and posts `L10n.Conclusions.sourcesPersonDeletedAnnouncement(ref:)` when a ref was listed; else `session.apply(effects)` and nothing is spoken |
| Unpromote flow | `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` `UnpromoteFlow.confirm(session:userID:apply:)`, `UnpromoteConfirmItem.message` | `confirm` writes, calls `apply` or `session.apply`, clears `pending`. `message` falls through to `L10n.string(L10n.Unpromote.confirmMessage)` when nothing is deleted and the preflight ran |
| Evidence graph Unpromote | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (`model.unpromote.confirm(session:userID:)`) | Uses the same `UnpromoteConfirmItem.message`, so it reads the per-kind keep message after this PR |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Unpromote.confirmMessage`, `confirmMessageDeletes(kind:ref:)`, `L10n.Conclusions.memberCount(_:)`, `sourcesPersonDeletedAnnouncement(ref:)`; `macos/App/Resources/Localizable.xcstrings` | `unpromote.confirmMessage` = "This promotion is removed. The cards stay on the graph, and the handle stays."; `conclusions.memberCount` plural "%lld member" / "%lld members", comment "how many Subjects are members of this handle" |
| Header count | `macos/App/Features/Conclusions/PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift` `members` | Each sets `members = L10n.Conclusions.memberCount(detail.memberCount)` |
| Preview copy | `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` `#Preview("Unpromote confirm")` | Literal "…, and the handle stays." |
| Announce helper | `macos/App/Platform/AccessibilityAnnouncement.swift` `AccessibilityAnnouncement.post(_:priority:)` | Posts `announcementRequested` on `NSApp`; silent when VoiceOver is off |
| Focus helper | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)`; `macos/App/DesignSystem/Components/Input/PVFieldFocus.swift` `PVFieldFocus.init(_:equals:next:previous:)` | Claim writes on the next main turn and again a turn later; the handle applies `.focused(binding, equals:)` |
| Tests | `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `confirmReportsTheEffects`, `messageFollowsThePreflight` | `confirmReportsTheEffects` rewrites the leave decision in its own `apply:` closure (P-16); `messageFollowsThePreflight` expects `L10n.string(L10n.Unpromote.confirmMessage)` |
| Tests | `macos/ProvenenciaTests/L10nResolutionTests.swift` `unpromoteCopyResolves`; `PersonDetailTests.swift` (line ~407); `PlaceDetailTests.swift` (line ~131) | Expect "…, and the handle stays.", "4 members", "1 member" |
| Tests | `macos/ProvenenciaTests/PVFieldFocusTests.swift` | The kit focus handle's tests; no `PVSegmentedControl` test exists |

**Precedent.** S13-17 (#729) added the delete branch's announcement in `PersonSourcesSection.confirm` with `AccessibilityAnnouncement.post`; copy its spoken-only string shape (`conclusions.sources.personDeletedAnnouncement`, "%@ deleted."). `UnpromoteConfirmItem.message` / `confirmMessageDeletes(kind:ref:)` is the per-kind copy to mirror. S16-07 (#761) is the focus precedent: decision in plain data that a test reaches, the write through `FocusClaim`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVSegmentedControl` puts the handle on the selected segment when it is enabled, else on the first enabled segment, else on none. The choice is a static function `PVSegmentedControl.focusTarget(items:selection:) -> String?` *(new)* returning the item id, which `segment(_:)` reads | The audit's fix (P-11); a static lets a test pin the rule without a view. No caller passes a handle today, so no other screen changes | Keep "first enabled" and claim the selected one from the page — because Tab into the control would then land on the wrong segment |
| D-2 | The switcher's handle is `PVFieldFocus($switcherFocus, equals: selection)` over `@FocusState private var switcherFocus: ConclusionPageSection?` *(new)* in `ConclusionDetailHeaderBand`. The focused value is the section of the segment that holds the handle | One state, no new focus type; the selected section is the only segment that carries the handle | A one-case enum stop type — because the section already names the stop |
| D-3 | Which section holds the keyboard is read with `@FocusedValue(\.conclusionOpenPageSection)` *(new key)*. `ConclusionDetailLoaded` puts `.focusedValue(\.conclusionOpenPageSection, openPageSection)` on the primary column's `VStack` | A focused value propagates from the focused stop up its ancestors, so the header knows whether focus sits inside the section that is about to be replaced, without each section reporting its own `FocusState`. It is SwiftUI API on the deployment target (macOS 14) | Always claiming the switcher on a switch — because ⌘2 typed with focus in the omnibar or sidebar would pull focus out of it (AC-1.4); reading `window.firstResponder` — because SwiftUI focus inside one hosting view does not move the first responder |
| D-4 | What a switch does beyond navigating is the value `ConclusionPageSectionSwitch` *(new)* in `ConclusionPageSection.swift`: a failable `init?(from:to:focusedSection:switcherFocused:)` that is nil when `from == to`, with `announcement: String` (`L10n.Conclusions.sectionShownAnnouncement(_:)`) and `claimsSwitcher: Bool` = `switcherFocused \|\| focusedSection == from` | Plain data a unit test reaches (rule "Every outcome on the page is spoken"; S16-07's shape) | Logic inline in the view — because the audit's P-16 lesson is that a decision in a view is untested |
| D-5 | The header band builds the switch value on the press and keeps it in `@State private var pendingSwitch: ConclusionPageSectionSwitch?` *(new)*, then calls `select`. `.onChange(of: openPageSection)` performs it only when the new section equals the switch's `to`: posts the announcement with `AccessibilityAnnouncement.post` (default priority) and, when `claimsSwitcher`, `FocusClaim.claim($switcherFocus, to)`. Any `openPageSection` change clears `pendingSwitch` | A held navigation (`WorkspaceNavigation.go(to:)` → `hold`) is not yet a switch; speaking on the change, not the press, never says a section that did not open. Focus is read on the press, before the old section is gone. Back, Forward and links change the section with no pending switch and speak nothing (AC-1.6) | Speaking in the `select` closure — because a discard gate could cancel it |
| D-6 | The spoken text is "{Section label} shown." (`conclusions.sectionShownAnnouncement`, argument the section's label) | User, 2026-11-11; matches "PER-8JD2L deleted." | "{Section}" alone; "{Section} section" |
| D-7 | The leave-or-stay decision is `enum UnpromotePageOutcome: Equatable { case leave(announcement: String?), stay(announcement: String) }` *(new)* with `static func after(_ effects: CatalogWriteEffects, item: UnpromoteConfirmItem, pageEntityId: String) -> UnpromotePageOutcome`, in `UnpromoteFlow.swift`. Leave when `effects.deletedHandles` contains `pageEntityId`; its announcement is `sourcesPersonDeletedAnnouncement(ref:)` of the ref `item.deletedHandles?.listed` gives that id, nil when not listed. Else stay with `sourcesUnpromotedAnnouncement(name: item.name)` | The working rule names this function and its two results; reading the ref from the item keeps today's behavior | A method on the view — because P-16 |
| D-8 | `UnpromoteFlow.confirm(onPageOf entityId: String, session: WorkspaceSession, userID: String, leave: @escaping @MainActor () -> Void, announce: @escaping @MainActor (String) -> Void = { AccessibilityAnnouncement.post($0) }) async -> Bool` *(new)* captures `pending` before the write, calls `confirm(session:userID:apply:)` and in `apply` switches on D-7: leave → `session.apply(effects, leavingPage: leave)` then announce when non-nil; stay → `session.apply(effects)` then announce. `PersonSourcesSection.confirm()` becomes one call with `leave: navigation.fallbackToSectionRoot` | Moves the applying out of the view too, so a test drives the real branch against a `WorkspaceNavigation` (P-16's fix) and no test rewrites it. `announce` is injected so a test reads what was said | Only the pure outcome, with the view still applying — because breaking the view's switch would again fail no test |
| D-9 | The stay announcement names the record (`item.name`, "J. Weber"), not the person | The audit's string "{name} unpromoted."; the record is what left the list | The person's name — because the person did not change |
| D-10 | `conclusions.memberCount` keeps its key and Swift name; its values become "%lld record" (one) / "%lld records" (other) and its comment says records | Rule "A person's records are records": one string on every Conclusion page; renaming the Swift function is churn the rule does not ask for | A new key `conclusions.recordCount` — because the old key would linger for no reader |
| D-11 | `L10n.Unpromote.confirmMessage` (a `LocalizedStringResource`) becomes `static func confirmMessage(kind: EvidencePrimaryKind) -> String` with keys `unpromote.confirmMessage.person` / `.event` / `.place` *(new)*, shaped like `confirmMessageDeletes(kind:ref:)`; the key `unpromote.confirmMessage` leaves the catalog | Rule: "worded per kind like `confirmMessageDeletes`, the Evidence graph's confirm included" | Interpolating the kind's noun — because the catalog's other per-kind lines are whole sentences per kind |
| D-12 | Tests: Swift Testing in `macos/ProvenenciaTests`; no UI test and no window. Focus landing and VoiceOver speech are checked by hand (§10) | Batch 3 rule "A test run shows no window"; the decisions are covered as plain data | An XCUITest for ⌘3 — because it needs a shown window |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVSegmentedControl.focusTarget(items:selection:)` *(new)* | func | `PVSegmentedControl.segment(_:)`'s inline `items.first(where:)` | Extends it: the inline expression moves into the static and changes rule |
| `ConclusionPageSectionSwitch` *(new)* | type | `ConclusionPageSection` (label), `FocusKeeper` (focus rules as plain data) | `FocusKeeper` keeps a stop across width changes and reloads; a switch is a one-shot decision with an announcement. Searched `rg "announce" macos/App/Features/Conclusions` |
| `FocusedValues.conclusionOpenPageSection` *(new)* | key | none: `rg "focusedValue\|FocusedValue" macos/App` finds nothing | First use; one key, set in one place |
| `UnpromotePageOutcome` *(new)* | type | `PersonSourcesSection.confirm()`'s inline branch | Replaces it |
| `UnpromoteFlow.confirm(onPageOf:session:userID:leave:announce:)` *(new)* | func | `UnpromoteFlow.confirm(session:userID:apply:)` | Built on it; the Evidence graph keeps calling the old one |
| `L10n.Conclusions.sectionShownAnnouncement(_:)` *(new)* | key | `sourcesPersonDeletedAnnouncement(ref:)` | Different sentence |
| `L10n.Conclusions.sourcesUnpromotedAnnouncement(name:)` *(new)* | key | `sourcesPersonDeletedAnnouncement(ref:)` | Different sentence; sits beside it |
| `L10n.Unpromote.confirmMessage(kind:)` *(new form)* | key | `L10n.Unpromote.confirmMessageDeletes(kind:ref:)` | Mirrors it |
| `macos/ProvenenciaTests/PVSegmentedControlTests.swift` *(new)* | test file | `PVFieldFocusTests.swift` | One file per kit component, as the other `PV…Tests` |
| `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift` *(new)* | test file | `WorkspaceNavigationTests.swift` | Navigation tests cover locations, not the switch's speech and focus |

**Builds on.** The kit's one focus handle (`PVFieldFocus`), `FocusClaim` as the one writer, `AccessibilityAnnouncement.post`, `UnpromoteFlow` as the one Unpromote flow, per-kind copy as in `confirmMessageDeletes`.
**Deletes.** `PersonSourcesSection.confirm()`'s inline leave-or-stay branch; `UnpromoteFlowTests.confirmReportsTheEffects`' inline `apply:` copy of it (the busy and evict assertions move to T-4); `L10n.Unpromote.confirmMessage` as a `LocalizedStringResource` and its catalog key `unpromote.confirmMessage`; the "handle stays" literal in `PVUnpromoteConfirm.swift`'s preview.
**Next adds.** A Conclusion kind that gains sections (Event, Place) gets the spoken switch and the focus landing from the shared header with no code; another page that offers Unpromote and can leave calls `confirm(onPageOf:…)`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| The switcher becomes a Tab stop (one segment carries the handle, every enabled segment is focusable) | `ConclusionDetailHeaderBand` | The band is inside the page, which is keyed by entity, so `switcherFocus` and `pendingSwitch` start nil on another person. No section is missing: `offered(by:)` is static |
| Focus written to a segment by `FocusClaim` | `PVSegmentedControl` | The selected segment exists on the next turn (the switcher is not replaced on a switch); `FocusClaim` writes twice for a late view |
| A Remove whose preflight failed but the write deleted the person | `UnpromotePageOutcome.after` | Leave with no announcement (AC-2.3), as today |

### Scale

None — no write, list or loop changes. The switcher has at most five items.

### Edges

- **Assumptions.** `openPageSection` changes only by navigation; the header band stays mounted across a switch (it sits in `PVPageColumns`' header, not the replaced column). T-2's cases hold the switch rule; the manual check holds the mount.
- **Neighbors.** Same section chosen again → nil switch, nothing spoken (AC-1.5, T-2). Focus in the header's Try again or lead menu → not in the old section and not on the switcher → no claim (T-2 "focus elsewhere"). A disabled selected segment → handle on the first enabled (T-1); the Person switcher has none disabled.
- **Removed inputs.** A switch held by the navigation gate and then cancelled leaves `pendingSwitch` set until the next section change, which clears it without speaking unless it matches (D-5). A record whose Remove also deleted another handle but not the page's → stay (AC-2.4, T-3).
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
| Section switcher with a focus handle | `PVSegmentedControl` (kit, changed rule D-1) | `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` | Doc comment: "A `focus:` handle lands on the selected segment, or the first enabled one when the selected one is disabled" |
| Switch speech and focus | `ConclusionPageSectionSwitch` *(new)* | `macos/App/Features/Workspace/ConclusionPageSection.swift` | Plain data |
| Header band wiring | `ConclusionDetailHeaderBand` | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` | `@FocusState switcherFocus`, `@FocusedValue focusedSection`, `@State pendingSwitch`, `.onChange(of: openPageSection)` |
| Focused section key | `FocusedValues.conclusionOpenPageSection` *(new)* | `ConclusionDetailPage.swift`, beside `ConclusionDetailPageSectionsContext` | `@Entry` or an explicit `FocusedValueKey` (macOS 14: use `FocusedValueKey`, since `@Entry` for `FocusedValues` needs macOS 15) |
| Remove outcome | `UnpromotePageOutcome`, `UnpromoteFlow.confirm(onPageOf:…)` *(new)* | `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` | |

States: nothing new is drawn. Accessibility: the switcher gains a Tab stop at default macOS settings (one per enabled segment, as the control documents); ⌘1–⌘5 keep working with focus anywhere.

### Strings

L10n keys *(new)*:
- `L10n.Conclusions.sectionShownAnnouncement(_ section: String)` → `conclusions.sectionShownAnnouncement` = "%@ shown." — comment: "VoiceOver announcement after a record page's section switcher opens a section; the argument is the section's name".
- `L10n.Conclusions.sourcesUnpromotedAnnouncement(name:)` → `conclusions.sources.unpromotedAnnouncement` = "%@ unpromoted." — comment: "VoiceOver announcement after one of a person's records is unpromoted and the person stays; the argument is the record's name".
- `L10n.Unpromote.confirmMessage(kind:)` → `unpromote.confirmMessage.person` = "This promotion is removed. The cards stay on the graph, and the person stays."; `.event` = "This promotion is removed. The cards stay on the graph, and the event stays."; `.place` = "This promotion is removed. The cards stay on the graph, and the place stays." — comment: "Message of the Unpromote confirm when the {kind} keeps another record".

Changed: `conclusions.memberCount` = "%lld record" / "%lld records"; comment "Page header: how many records (Identity Claims) this person, event or place has; argument is the count".

Removed: `unpromote.confirmMessage`.

Old-scope copy: "members" and "handle" leave the page; `rg -i "handle stays|[0-9] members" macos docs .agents` finds nothing after STEP-6.

## 7. Steps

**STEP-1 — The segmented control's handle lands on the selected segment**
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/ProvenenciaTests/PVSegmentedControlTests.swift` *(new)*. Skill: `add-ui-component`, `add-swift-test`.
Do: add `static func focusTarget(items: [Item], selection: Value) -> String?` (D-1); `segment(_:)` compares `item.id` with it; update the doc comment. Write T-1.
Check: `xcodebuild test -scheme Provenencia -only-testing:ProvenenciaTests/PVSegmentedControlTests` passes.

**STEP-2 — The switch value**
Files: `macos/App/Features/Workspace/ConclusionPageSection.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift` *(new)*. Skill: `add-localized-string`.
Do: add `conclusions.sectionShownAnnouncement` and `ConclusionPageSectionSwitch` (D-4, D-6). Write T-2.
Check: `-only-testing:ProvenenciaTests/ConclusionPageSectionSwitchTests` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The switcher speaks and keeps focus**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`.
Do: add the `conclusionOpenPageSection` focused value key and set it on the primary column's `VStack` in `ConclusionDetailLoaded` (D-3). In `ConclusionDetailHeaderBand`: `@FocusState switcherFocus`, pass `focus: PVFieldFocus($switcherFocus, equals: selection)` (D-2), wrap `select` to build `pendingSwitch` from `openPageSection`, the chosen section, `focusedSection` and `switcherFocus != nil`, then call `select`; `.onChange(of: openPageSection)` performs and clears it (D-5).
Check: the app builds with no Swift warning; the manual check steps 1–4 (§10) pass.

**STEP-4 — The Remove outcome**
Files: `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift`. Skill: `add-localized-string`.
Do: add `conclusions.sources.unpromotedAnnouncement`, `UnpromotePageOutcome` (D-7, D-9) and `confirm(onPageOf:…)` (D-8). Write T-3, T-4, T-5; delete `confirmReportsTheEffects`' inline decision (keep its busy assertions in T-4).
Check: `-only-testing:ProvenenciaTests/UnpromoteFlowTests` passes.

**STEP-5 — The Sources section calls it**
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift`.
Do: `confirm()` becomes `await unpromote.confirm(onPageOf: entityId, session: session, userID: userID, leave: navigation.fallbackToSectionRoot)`; the doc comment says the flow decides leave or stay. Nothing else in the file changes.
Check: `rg "deletedHandles.contains" macos/App/Features/Conclusions` finds nothing; the app builds; manual check steps 5–6 pass.

**STEP-6 — Records, and the person stays**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` (`UnpromoteConfirmItem.message`), `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` (preview literal), `macos/ProvenenciaTests/L10nResolutionTests.swift`, `UnpromoteFlowTests.swift`, `PersonDetailTests.swift`, `PlaceDetailTests.swift`, `EventDetailTests.swift`. Skill: `add-localized-string`.
Do: D-10, D-11. `message` returns `L10n.Unpromote.confirmMessage(kind: kind)`. Update T-6, T-7.
Check: `-only-testing:ProvenenciaTests/L10nResolutionTests -only-testing:ProvenenciaTests/UnpromoteFlowTests -only-testing:ProvenenciaTests/PersonDetailTests -only-testing:ProvenenciaTests/PlaceDetailTests -only-testing:ProvenenciaTests/EventDetailTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests` pass; `rg -i "handle stays|[0-9] members" macos docs .agents` finds nothing.

**STEP-7 — Docs and version**
Files: `docs/macos-client-patterns.md` ("One sectioned detail page"), `VERSION`, `core/version.go`, `MARKETING_VERSION` in the Xcode project, `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status, Landed note).
Do: add one sentence to "One sectioned detail page": the switcher carries a focus handle on the selected segment; a switch chosen on it says "{Section} shown." and keeps the keyboard on the switcher when it was in the old section (`ConclusionPageSectionSwitch`). Bump the patch. Per the Definition of done.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVSegmentedControlTests.swift` `focusLandsOnTheSelectedSegment` *(new)* | [A, B, C] sel C → C; [A, B(disabled), C] sel B → A; [A(disabled), B] sel A → B; all disabled → nil | AC-1.3, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/ConclusionPageSectionSwitchTests.swift` `switchSpeaksAndClaims` *(new)* | family→timeline, focusedSection .family → announcement "Timeline shown.", claims; vitals→gallery, switcherFocused → claims; vitals→family, focusedSection nil, not switcher → "Family shown.", no claim; timeline→timeline → nil | AC-1.1, AC-1.2, AC-1.4, AC-1.5, D-4, D-6 |
| T-3 | Swift | `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `pageOutcomeLeavesOnlyWhenThePageEntityIsDeleted` *(new)* | deleted [h-maria], listed PER-8JD2L → `.leave("PER-8JD2L deleted.")`; deleted [h-maria], none listed → `.leave(nil)`; deleted [h-otto] → `.stay("J. Weber unpromoted.")`; nothing deleted → `.stay("J. Weber unpromoted.")` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-7, D-9 |
| T-4 | Swift | `UnpromoteFlowTests.swift` `confirmOnPageLeavesWhenThePersonIsDeleted` *(new, replaces `confirmReportsTheEffects`)* | `WorkspaceNavigation` at `.personDetail(h-maria)`, detail key warmed, effects delete h-maria → `navigation.currentLocation == .sectionRoot(.persons)`, detail evicted, announced ["PER-8JD2L deleted."], `pending == nil`, busy while the write is held | AC-2.2, D-8 |
| T-5 | Swift | `UnpromoteFlowTests.swift` `confirmOnPageStaysAndSaysSo` *(new)* | effects `handles: [h-maria]`, nothing deleted → location unchanged, announced ["J. Weber unpromoted."], returns true | AC-2.1, D-8 |
| T-6 | Swift | `UnpromoteFlowTests.swift` `messageFollowsThePreflight` (updated); `L10nResolutionTests.swift` `unpromoteCopyResolves` (updated) | none deleted, kind .person → "…, and the person stays."; `confirmMessage(kind: .event)` / `.place` resolve to their sentences | AC-3.2, AC-3.3, D-11 |
| T-7 | Swift | `PersonDetailTests.swift`, `PlaceDetailTests.swift` (updated), `EventDetailTests.swift` (one assertion added) | memberCount 4 → "4 records"; 1 → "1 record" | AC-3.1, D-10 |

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

The sweep names: `L10n.Unpromote.confirmMessage` (old resource form), `unpromote.confirmMessage` (old key), "handle stays", "members" in the header copy. "UI matches the brief inventory": no brief; no new kit primitive.

## 10. Manual check

Turn on VoiceOver (⌘F5) and Full Keyboard Access (System Settings → Keyboard → Keyboard navigation). Use a project with a person who has two records on different Evidence graphs, and a second person with one record.

1. Open the first person on Family. Tab to a family tile. Press ⌘3. Hear "Timeline shown."; the Timeline segment shows its focus ring; Tab moves into the Timeline (AC-1.1).
2. Shift-Tab back to the Timeline segment, press ← or Tab to the Gallery segment, press Space. Hear "Gallery shown."; focus is on Gallery (AC-1.2).
3. Click in the omnibar field, press ⌘2. Hear "Family shown."; the caret stays in the omnibar (AC-1.4). Press ⌘2 again: nothing is spoken (AC-1.5).
4. Press Back (⌘[) to Gallery: nothing is spoken (AC-1.6). The header reads "2 records" (AC-3.1); an Event page and a Place page read "N records".
5. Open Sources. Unpromote one record: the confirm ends "…, and the person stays." (AC-3.2). Confirm: hear "{record} unpromoted."; the page stays (AC-2.1).
6. Open the second person on Sources, Unpromote its one record, confirm: the app leaves for the Persons list and says "{ref} deleted." (AC-2.2).
7. On an Evidence graph, Unpromote an Event card whose event keeps another record: the confirm ends "…, and the event stays." (AC-3.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `@FocusedValue` reads nil while a stop inside the section has focus (the header is a sibling of the column) | Manual check step 1; Stop condition below |
| The focus move to the next Unpromote button speaks over "J. Weber unpromoted." | The announcement is posted after `session.apply`, before the reload moves focus; manual check step 5 listens for both. Stop condition below |
| ⌘1–⌘5 reach the header band through `PVSegmentShortcut`, which calls `select` without the wrapper | D-5 wraps `select` at the band, so every path (click, Space, shortcut) goes through it |

**Stop and report if:**

- With a Family tile focused, `@FocusedValue(\.conclusionOpenPageSection)` reads nil at the press (D-3 cannot work as written).
- The header band remounts on a switch (its `@State` resets between the press and `onChange`), so `pendingSwitch` is lost.
- VoiceOver does not speak "{record} unpromoted." in manual check step 5 at the default priority.
- `UnpromoteConfirmItem.name` is not the record's display name on the Sources section (the confirm title shows something else).
- A `PVSegmentedControl` caller other than `ConclusionDetailPage.swift` passes `focus:` at the base.
- Any `.agents/` skill or `docs/*.md` line quotes "members" for the header or "handle stays" (sweep it in STEP-6 and report it).

## 12. Open questions

None.

## 13. PR

**Title:** S16-06 — Section switches and removals are spoken
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-06-spoken-switches-and-removals.md`.
