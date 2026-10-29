# S13-10 — Header and sections

**Status:** Blocked — on board S13-D1
**Spike / slice:** Spike 13 · Slice 1 — Page frame and Vitals
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D1](../design/S13-D1-person-page-frame.md), board not agreed yet
**Board:** — (to record when agreed: the Claude Design link and frame numbers; the implementer reads §6 **macOS**, which describes each frame in words)
**Depends on:** S13-09 (landed [#624](https://github.com/mendahu/provenencia/pull/624)), S13-D1
**Blocks:** S13-11, S13-12, S13-14, S13-16
**Branch:** `s13-10-header-and-sections`
**Version bump:** minor — the Person page gains sections a researcher uses, a new screen behavior ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-1 – R-7, R-14, R-15; working rules: "Sections are places", "One sectioned Conclusion page, not a Person fork", "S13-D1 draws the whole frame"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-10 ("Reuse first", from S13-09 ship log MNT-1); [`macos-client-patterns.md`](../../../macos-client-patterns.md) Conclusion keys
**Planned:** 2026-10-29 against `main` @ `b8ca4cfc` (draft; frame-dependent Decisions wait on the board)
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

The Person page keeps its header and gains a section switcher under it: Vitals, Family, Timeline, Gallery, Sources. Vitals holds today's rows; Gallery holds today's likeness grid; Family, Timeline and Sources say they arrive later. Each section is a place: switching pushes history, Back returns to the previous section, and a restored session reopens on it.

**Why now.** Slice 1's chrome PR; every later Person section (S13-11, S13-12, S13-14, S13-16) fills this frame. Working rules "Sections are places" and "S13-D1 draws the whole frame".

**What stays.** The header's parts (thumb, gender Badge, name, b. and d. lines, ref, member count) and the thumb's right-click "Use earliest likeness" (user, 2026-10-29: stays until the S13-D5 Gallery board); every row with its Badge, Sources and Why; the likeness tiles with Use as portrait, the Portrait Badge and open Citation; one `conclusionDetail` read; the Event and Place pages, which keep the convenience initializer and today's layout (R-1); every existing link into a Person page, which lands on Vitals.

## 2. User stories

**US-1.** As a researcher, when I open a Person, I see the header and a switcher with five sections, Vitals showing, so that I can tell where I am and where else I can go.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Open a Person from the Persons list | The header as today; the switcher lists Vitals, Family, Timeline, Gallery, Sources in that order; Vitals is selected and shows today's rows in today's order |
| AC-1.2 | Same person, choose Gallery | Today's likeness grid, with the Portrait Badge on the portrait tile and Use as portrait on the others; the header does not change |
| AC-1.3 | A person with no likeness, choose Gallery | `PVEmptyState` titled "No likenesses yet" (wording per board) |
| AC-1.4 | Choose Family, Timeline or Sources | `PVEmptyState` with that section's name and the line that it arrives in a later version (wording per board); no rows, no skeleton |
| AC-1.5 | VoiceOver on the switcher with Vitals showing | Reads "Vitals, selected" and each other section by name |

**US-2.** As a researcher, when I switch sections, Back and Forward walk them and a reopened project returns to the section I was on, so that a section behaves like any other place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Persons list → Jane Doe (Vitals) → Family → Gallery, then Back | Family; Back again → Vitals; Back again → Persons list |
| AC-2.2 | On Gallery, choose Gallery again | No history entry is added |
| AC-2.3 | On Timeline, quit and reopen the project | Jane Doe on Timeline |
| AC-2.4 | A history file written before this PR (no `personSection` key) | Decodes; Person entries open on Vitals |
| AC-2.5 | A history entry with `"personSection": "lineage"` (unknown) | Decodes; opens on Vitals |
| AC-2.6 | From Gallery, open a likeness tile's Citation, then Back | Jane Doe on Gallery |
| AC-2.7 | Breadcrumbs on Jane Doe's Gallery | Persons › Jane Doe › Gallery, with Jane Doe going to Vitals; on Vitals, Persons › Jane Doe (wording per board, Q-3) |
| AC-2.8 | Back jump menu after Vitals → Family | The Vitals row and the Family row are distinguishable (Q-3) |

**US-3.** As a contributor, when I build the next Person section, Vitals and Gallery reuse named parts of the default fill, so that a section is a case in one switch and not a copy of the Details stack.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ConclusionDetailDefaultPageSections` | Built from `ConclusionDetailRowsSection` *(new)*, `ConclusionDetailLikenessSection` and the relationship `ForEach`; Event and Place look the same |
| AC-3.2 | Person's below-header fill | One `switch` over `PersonPageSection`; Vitals is `ConclusionDetailRowsSection`, Gallery is `ConclusionDetailLikenessSection`; no Details header, spinner or row loop copied into `PersonDetailView.swift` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PersonPageSection` *(new)* and `WorkspaceLocation.personSection` *(new)*: Codable, nil is Vitals, unknown decodes to nil, part of `==` (PPF-5) | STEP-1, AC-2.1 – AC-2.5, D-1, D-2 |
| Split the default fill into `ConclusionDetailRowsSection` *(new)* and an internal `ConclusionDetailLikenessSection` (S13-10 scope "Reuse first") | STEP-2, AC-3.1, D-4 |
| Person's fill: switcher + one section at a time; Vitals rows, Gallery grid or empty, three empty sections (PPF-1, PPF-2, PPF-4, PPF-8 – PPF-10) | STEP-3, STEP-4, AC-1.1 – AC-1.5, AC-3.2, D-3, D-5 – D-7 |
| Switching pushes history through `navigation.go(to:)`; each section opens at its top (PPF-5, PPF-7) | STEP-4, AC-2.1, AC-2.2, AC-2.6, D-3 |
| Breadcrumbs and jump rows name the section (PPF-6) | STEP-5, AC-2.7, AC-2.8, D-8 |
| Keyboard: focus, arrows, ⌘1 – ⌘5 (PPF-13) | STEP-4, D-9 |
| The pinned or scrolling header the board picks (PPF-3); narrow window (PPF-12) | STEP-3, Q-1, Q-2 |
| L10n for five section names, the empty states and the crumb/jump-row format | STEP-3, STEP-5, D-10 |
| Docs: patterns bullet, `completed.md`, map status, brief archive | STEP-6 |

| Out | Home |
| --- | --- |
| Provenance control and sheet on Vitals rows | S13-11 (brief S13-D2) |
| Family, Timeline, Sources content | S13-12, S13-13, S13-16 |
| Gallery redesign, viewer, where "Use earliest likeness" lives | S13-14, S13-15 (brief S13-D5) |
| How the records connect | S13-28 |
| Sections on Event and Place pages | not this spike (R-1) |
| `personSection` on the engine's `WorkspaceLocation` proto (search context) | not planned — D-2 |
| Renaming `PersonDetailContent.vitals` / `Vital` (the header's b. and d. lines) | not planned — §6 Codebase fit names the two meanings |
| A Tab kit component | not planned — brief non-goal; segmented `PVChipGroup` |
| Remembering the last section per person | not planned — brief non-goal (PPF-5) |
| New `.swift` file or `project.pbxproj` entry | not planned — D-11 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Page and slot | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | Three generics after S13-09; full init takes `pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections`; convenience init (where `PageSections == ConclusionDetailDefaultPageSections<Content>`) supplies the default fill. `static func page` wraps every state in one `ScrollView` (header included) |
| Loaded body | same file `ConclusionDetailLoaded` (private) | On `.content`: `ConclusionDetailHeader`, then `pageSections(content, context)` |
| Default fill | same file `ConclusionDetailDefaultPageSections` (internal) | One block: Details `PVSectionHeader` (`L10n.Conclusions.detailsSection`, refreshing `ProgressView` aside) + `ForEach(content.rows)` → `ReconciledValueRow`; likeness section when non-empty; `ForEach(content.sections)` → relationship section. Doc comment: siblings take spacing from the page `VStack`, so a custom fill returns siblings too |
| Likeness grid | same file `ConclusionDetailLikenessSection` (private) | Header with count, adaptive grid of 120pt columns, tile opens `tile.location` via `navigation.go(to:)`, Portrait Badge / Use as portrait from `ConclusionPortraitModel` |
| Header | same file `ConclusionDetailHeader` (private) | Thumb (with the "Use earliest likeness" right-click menu when `offersEarliest`), Badge, title, summary, ref, members |
| Context | same file `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing`, `portrait` |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView` | Convenience init; `summary` is private `PersonDetailVitals` (the b. and d. lines); `.vocabularyToastOverlay($portrait.toast, …)`; does not read `WorkspaceNavigation` |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent` | `vitals: [Vital]` = the header's b. and d. lines (not the Vitals section); `rows`, `likenesses`, `thumbnail`, `badge` |
| Location | `macos/App/Features/Workspace/WorkspaceLocation.swift` `WorkspaceLocation` | `entityId` names the person; no section; `personDetail(entityId:ref:title:)`; `withListScrollAnchor` is the copy-with precedent; `SourceSurface` enum lives in this file; `==` lists deep ids; tolerant `init(from:)` with `decodeIfPresent` per key |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` spec `.personDetail` | `viewIdentity: { $0.entityId }`, so a section change does not remount the page; query keys `[.conclusionDetail]` |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.personDetail` | `PersonDetailView(session:entityId: navigation.currentLocation.entityId ?? "", userID:)` |
| Navigation | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `go(to:)` | Coalesces when `location == currentLocation`; otherwise pushes and persists |
| Breadcrumbs | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems(for:names:goTo:)` | Entity location → [section crumb, leaf = `title` ?? `ref`] |
| Jump rows | same file `HistoryJumpRow.leafLabel` (private) | `title` ?? `ref` |
| Proto | `api/proto/engine.proto` `WorkspaceLocation`; `macos/App/Platform/GoStore.swift` `mapWorkspaceLocationToProto` / `FromProto` | Sent as search context only (`searchCatalog`); `listScrollAnchor` is not on the wire |
| Links in | `PersonsListView.swift` (`.personDetail(`), `EvidenceGraphModel.swift` (`.personDetail(`), omnibar hits via `mapWorkspaceLocationFromProto` | All build a location without a section → Vitals |
| Kit | `DesignSystem/Components/Chip/PVChip.swift` `PVChip`, `PVChipGroup(style: .segmented)`; `DesignSystem/Components/EmptyState/` `PVEmptyState` | `PVChip` is a `Button`; the segmented group is used by the date-kind and Properties value-type pickers |
| Shortcuts | `rg keyboardShortcut macos/App` | ⌘1 – ⌘9 unused |
| Tests | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift`, `PersonDetailTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift` | Push/restore, legacy decode (`legacyHistoryMissingListScrollAnchorDecodesAsNil`), equality, crumbs; Person content |

**Precedent.** `listScrollAnchor` (S10) added an optional, Swift-only, tolerant-decoding field to `WorkspaceLocation` with a copy-with helper and legacy-decode tests: copy its shape for `personSection`, but unlike the anchor the section **is** identity (it is in `==`). `SourceSurface` is the precedent for a small Codable enum beside `WorkspaceLocation`. S13-09 ([#624](https://github.com/mendahu/provenencia/pull/624)) is the precedent for the page's slot: use the full init with a custom `pageSections`, and return sibling views.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `enum PersonPageSection: String, Codable, CaseIterable, Sendable` *(new)* with cases `vitals, family, timeline, gallery, sources` in that order, in `WorkspaceLocation.swift` beside `SourceSurface`. It owns `label: LocalizedStringResource` and `glyph` for its empty state | One ordered list drives the switcher, shortcuts, labels and decode; `SourceSurface` precedent | Raw `String` on the location (untyped); a Conclusions-folder file (pbxproj churn, D-11) |
| D-2 | `WorkspaceLocation.personSection: PersonPageSection?` *(new)*, coding key `personSection`. Nil means Vitals: `init` stores nil for `.vitals`, and nil unless `section == .persons` and `entityId != nil`. `init(from:)` decodes a `String?` and maps it with `PersonPageSection(rawValue:)`, unknown → nil. `==` compares it. `personDetail(entityId:ref:title:section:)` gains `section: PersonPageSection = .vitals`. `func withPersonSection(_:) -> WorkspaceLocation` *(new)* copies `self` with the section, like `withListScrollAnchor`. Not on the proto | Sections are places (working rule); existing links land on Vitals with no change; old and unknown history decodes; one Person location for Vitals whether built with or without the argument | `.vitals` stored explicitly (two equal-looking locations unequal); adding it to `engine.proto` (search context does not need it; `listScrollAnchor` precedent) |
| D-3 | `PersonDetailView` reads `@Environment(WorkspaceNavigation.self)`. The section is `navigation.currentLocation.personSection ?? .vitals`; choosing one calls `navigation.go(to: navigation.currentLocation.withPersonSection(choice))`. No local `@State` for the section | `add-workspace-location`: the location is the one source of where we are; `go(to:)` already coalesces the same section (AC-2.2); `viewIdentity` stays `entityId`, so `ConclusionPortraitModel` and its toast survive a switch | A `@State` section synced to the location (a parallel "where we are"); `WorkspaceDestinationHost` passing the section in (the view still needs `navigation` to switch) |
| D-4 | Split `ConclusionDetailDefaultPageSections` into named internal parts in the same file: `ConclusionDetailRowsSection` *(new)* taking `rows: [ReconciledValueRowModel]`, `refreshing: Bool` and `title: LocalizedStringResource` (the Details header with its spinner, then the rows), and `ConclusionDetailLikenessSection` changed from `private` to internal. The default fill becomes `ConclusionDetailRowsSection(rows:refreshing:title: L10n.Conclusions.detailsSection)`, the likeness `if`, and the relationship `ForEach`, unchanged in output | Map scope "Reuse first" (S13-09 MNT-1); Vitals and Gallery reuse by name | Copying the header, spinner and rows into the Person view; making the whole default fill configurable with flags |
| D-5 | Person passes its own fill through `ConclusionDetailPage`'s full initializer: `pageSections: { content, context in PersonDetailSections(content:context:section:onSelect:) }` where `PersonDetailSections` *(new, private in `PersonDetailView.swift`)* returns sibling views: the switcher (unless Q-1 moves it to a pinned region), then one `switch section`: `.vitals` → `ConclusionDetailRowsSection(rows:refreshing:title:)` (title per Q-4); `.gallery` → `ConclusionDetailLikenessSection` when `content.likenesses` is non-empty, else the Gallery empty state; `.family`, `.timeline`, `.sources` → the section empty state | Working rule "One sectioned Conclusion page"; S13-09's slot; siblings take the page's spacing (S13-09 F-2) | A Person-only page type; a wrapper `VStack` around the fill (joins the siblings) |
| D-6 | The switcher is `PVChipGroup(style: .segmented)` with one `PVChip(section.label, isSelected:, expands: true, selectionLift: true)` per `PersonPageSection.allCases`, if the board picks pattern A (Q-1). Each chip has `.accessibilityIdentifier("persons.detail.section.<raw>")` and `.accessibilityAddTraits(.isSelected)` when selected; the group has `.accessibilityElement(children: .contain)` and the label "Sections" | Brief inventory; `PVChip` is a `Button`, so focus and VoiceOver come from the kit | A new Tab component; a native `TabView` or `Picker(.segmented)` (brief non-goal) |
| D-7 | Each empty section is `PVEmptyState(glyph:title:message:)` with the section's glyph, its title and one message; Family, Timeline and Sources share one message key | User, 2026-10-29: placeholders show an empty state; brief PPF-10 | Hiding unbuilt sections (user declined); skeleton rows |
| D-8 | `WorkspaceToolbar.breadcrumbItems` appends a third crumb with the section's label when `location.personSection != nil`; the person crumb then navigates to `location.withPersonSection(.vitals)`. `HistoryJumpRow.leafLabel` returns `L10n.Conclusions.personSectionLeaf(title, section)` *(new)* when a section is set. Wording waits on the board (Q-3) | PPF-6: two history entries of one person must be told apart | Section only in the jump row (crumbs then disagree with history) |
| D-9 | ⌘1 – ⌘5 are `.keyboardShortcut(KeyEquivalent(Character("\(index + 1)")), modifiers: .command)` on each switcher chip, so they act only while the Person page shows. Arrow keys between chips are the focus system's default for buttons in a row; no custom key handling. Final keys per board (Q-5) | Shortcuts scoped by the view, not by `ProvenenciaApp.commands` (the skill warns Scene commands freeze their disabled state) | App-menu commands for sections |
| D-10 | New copy lives in `L10n.Conclusions`: `personSectionVitals`, `personSectionFamily`, `personSectionTimeline`, `personSectionGallery`, `personSectionSources`, `personSectionsLabel`, `personSectionLater`, `galleryEmptyTitle`, `galleryEmptyMessage`, `personSectionLeaf(_:_:)` *(all new)*, keys `conclusions.person.section.*` and `conclusions.person.gallery.empty.*` | `add-localized-string`; one namespace with the page's other copy | A new `L10n.PersonPage` namespace for one page |
| D-11 | No new `.swift` file: `PersonPageSection` in `WorkspaceLocation.swift`, `ConclusionDetailRowsSection` in `ConclusionDetailPage.swift`, `PersonDetailSections` in `PersonDetailView.swift` | `project.pbxproj` is hand-managed until S14-15; S13-09 D-4 precedent | New files and pbxproj entries |
| D-12 | Each section opens at its top (PPF-7): the section content is keyed by `section` (`.id(section)`) so SwiftUI treats a switch as new content, and the page scrolls to the top on a section change. The mechanism depends on Q-1 | Brief PPF-7; scroll is UI ephemera, not stored in history (`add-workspace-location`) | Storing a per-section scroll anchor in the location |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PersonPageSection` *(new)* | type | `SourceSurface` (same file) | A different axis (which part of a Person page); extends the same pattern |
| `WorkspaceLocation.personSection`, `withPersonSection(_:)` *(new)* | field, func | `listScrollAnchor`, `withListScrollAnchor(_:)` | Same shape; differs only in being identity |
| `ConclusionDetailRowsSection` *(new)* | view | the Details block inside `ConclusionDetailDefaultPageSections` | Extracted, not added beside it (D-4) |
| `PersonDetailSections` *(new, private)* | view | `ConclusionDetailDefaultPageSections` | Person's fill differs by section; it reuses the default's parts |
| Switcher | composition | `PVChipGroup(style: .segmented)` + `PVChip` in `PropertiesView` and the date-kind picker | Reused as is |
| L10n keys (D-10) *(new)* | key | `L10n.Conclusions.detailsSection`, `likenessesSection` | New names; same namespace |
| Tests (§8) *(new)* | test | `legacyHistoryMissingListScrollAnchorDecodesAsNil`, `conclusionDetailCrumbPrefersTheTitle` | Same suites, new cases |

Two meanings of "vitals": `PersonDetailContent.vitals` and the private `PersonDetailVitals` are the header's b. and d. lines; `PersonPageSection.vitals` is the section. The plan does not rename the header's names (Out); the section type's name keeps them apart.

**Builds on.** S13-09's `pageSections` slot and full initializer; `add-workspace-location` (one location, `go(to:)`, tolerant decode); the kit's segmented chip group and `PVEmptyState`.
**Deletes.** The single-block body of `ConclusionDetailDefaultPageSections` (rebuilt from its parts); `private` on `ConclusionDetailLikenessSection`. No callee loses its last caller. `PersonDetailView`'s use of the convenience initializer (it moves to the full one).
**Next adds.** S13-11 adds the provenance control inside `ConclusionDetailRowsSection`'s rows (Person only, by caller). S13-12, S13-13 and S13-16 each replace one empty-state case in `PersonDetailSections` with their view. S13-14 replaces the Gallery case. S13-28 adds a case or a view inside Sources (its board decides). Event and Place adopt sections by adding their own section enum and location field when R-1 lifts.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A location with a section for a person that was deleted | `ConclusionDetailPage` failed state | Same as today's Vitals: the danger Callout; the section changes nothing |
| Restoring a non-Vitals section from history on launch | `PersonDetailView` | Section read from `currentLocation` on every render; no `onAppear` apply step needed (D-3) |
| Switching person (family tile later) while on Gallery | `PlaceRegistry` remounts on `entityId` | A new person's location has no section → Vitals (D-2) |
| Gallery with no likenesses | `PersonDetailSections` | Empty state (AC-1.3) |

### Scale

None — no write and no new read: switching re-renders from the one cached `conclusionDetail`. A person with 20 rows and 30 likenesses renders the same views as today, one section at a time.

### Edges

- **Assumptions.** `go(to:)` coalesces equal locations, so `==` must include `personSection` (T-3) and `init` must normalize `.vitals` to nil, or Vitals-by-argument and Vitals-by-default push twice (T-2).
- **Neighbors.** Event and Place locations never carry a section, even if decoded with one (T-4). A Persons list root with a section key decodes to no section (T-4).
- **Removed inputs.** None — no data is decided or deleted.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The history file is app-written; an unknown section string decodes to nil (T-5).

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — `personSection` stays off `engine.proto` (D-2).

### macOS

From the brief's inventory. Frame descriptions are filled in when the board is agreed (Q-1 – Q-5).

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Page | `ConclusionDetailPage` full init | `ConclusionDetailPage.swift` | Person only; Event and Place keep the convenience init |
| Header | `ConclusionDetailHeader` | same | Unchanged unless Q-2 adds a compact form |
| Switcher | `PVChipGroup(.segmented)` + `PVChip` | `PersonDetailView.swift` `PersonDetailSections` | D-6, D-9 |
| Vitals | `ConclusionDetailRowsSection` | `ConclusionDetailPage.swift` | D-4, Q-4 |
| Gallery | `ConclusionDetailLikenessSection` / `PVEmptyState` | same / kit | D-5 |
| Family, Timeline, Sources | `PVEmptyState` | kit | D-7 |
| Crumbs, jump rows | `PVBreadcrumbs`, `HistoryJumpRow` | `WorkspaceToolbar.swift` | D-8 |

States: loading and failed as today (no switcher drawn until content); refreshing spinner in the Vitals heading (Q-4 may move it). Accessibility: chips are buttons with names and the selected trait; empty states read title then message.

### Strings

L10n keys *(new)*, final wording per board: `L10n.Conclusions.personSectionVitals` = "Vitals", `personSectionFamily` = "Family", `personSectionTimeline` = "Timeline", `personSectionGallery` = "Gallery", `personSectionSources` = "Sources", `personSectionsLabel` = "Sections", `personSectionLater` = "This section arrives in a later version.", `galleryEmptyTitle` = "No likenesses yet", `galleryEmptyMessage` = "A likeness is drawn on a Citation's image.", `personSectionLeaf(title, section)` = "%@ · %@". Removed: none (`detailsSection` stays for Event, Place and the default fill).

## 7. Steps

**STEP-1 — The section is part of the location**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`. Skill: `add-workspace-location`, `add-swift-test`.
Do: D-1, D-2. Add the coding key, the tolerant decode, the normalization in `init`, `==`, `personDetail(…section:)`, `withPersonSection(_:)`. Write T-1 – T-5.
Check: `xcodebuild test` runs `WorkspaceNavigationTests` green.

**STEP-2 — Split the default fill into named parts**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`. Skill: `add-ui-component`.
Do: D-4. Event and Place still use the convenience init.
Check: `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` green; Event and Place pages unchanged in the app.

**STEP-3 — Blocked on the board: page structure for the pinned or scrolling header**
Files: `ConclusionDetailPage.swift`. Do: per Q-1 and Q-2 once answered. Check: Event and Place look as today.

**STEP-4 — Person sections and the switcher**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-3, D-5 – D-7, D-9, D-10, D-12. Switch `PersonDetailView` to the full init.
Check: `python3 scripts/check-localizable-xcstrings.py` clean; manual check §10 steps 1 – 5.

**STEP-5 — Crumbs and jump rows name the section**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `L10n.swift`, `Localizable.xcstrings`.
Do: D-8. Write T-6, T-7.
Check: `WorkspaceToolbarBreadcrumbTests` green.

**STEP-6 — Docs**
Files: `docs/macos-client-patterns.md` (one sentence under the sectioned detail page bullet: Person passes its own fill and its section lives on `WorkspaceLocation.personSection`), `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S13-D1-person-page-frame.md` → `../design/archive/` *(new)*, `../design/README.md`.
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py` on this plan.

Commits: one per behavior, full imperative sentences ("Remember which section of a Person page is open in history.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `WorkspaceNavigationTests` `personSectionPushAndRestore` *(new)* | go Vitals → Family → Gallery; back, back; persist, reattach → index and sections restored | AC-2.1, AC-2.3 |
| T-2 | Swift | `WorkspaceNavigationTests` `vitalsIsTheDefaultPersonSection` *(new)* | `.personDetail(…section: .vitals) == .personDetail(…)`; `personSection == nil`; going to the same section does not push | AC-2.2, D-2 |
| T-3 | Swift | `WorkspaceNavigationTests` `personSectionIsIdentity` *(new)* | same person, Family vs Gallery → not equal | D-2 |
| T-4 | Swift | `WorkspaceNavigationTests` `personSectionOnlyOnAPersonPage` *(new)* | event detail, persons list root built or decoded with a section → nil | D-2 |
| T-5 | Swift | `WorkspaceNavigationTests` `legacyAndUnknownPersonSectionDecodeAsVitals` *(new)* | JSON without the key; JSON with `"lineage"` → nil | AC-2.4, AC-2.5 |
| T-6 | Swift | `WorkspaceToolbarBreadcrumbTests` `personSectionAddsACrumb` *(new)* | Gallery → three crumbs, the person crumb goes to Vitals; Vitals → two crumbs | AC-2.7, D-8 |
| T-7 | Swift | `WorkspaceToolbarBreadcrumbTests` (jump row leaf, if `HistoryJumpRow` is made testable) or manual §10 step 6 | Family entry → "Jane Doe · Family" | AC-2.8 |
| T-8 | Swift | `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` existing | unchanged | AC-3.1 |
| T-9 | Manual | §10 | the five sections, empty states, VoiceOver, keyboard | AC-1.1 – AC-1.5, AC-2.6 |
| T-10 | Structural | STEP-4 `rg` | In `macos/App/Features/Conclusions/PersonDetailView.swift`: `rg -F 'ForEach(content.rows)'` and `rg -F 'detailsSection'` find nothing; `rg -F 'ConclusionDetailRowsSection'` and `rg -F 'ConclusionDetailLikenessSection'` each find a hit | AC-3.2, D-5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceNavigationTests` (`==` change), detail suites (STEP-2).

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
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — minor
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md (Person's fill and personSection)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Person that has a name, both life lines, three likenesses and a Mixed row. Open them from the Persons list: header as before, Vitals selected, rows as before. (AC-1.1)
2. Choose Gallery: today's grid, Portrait Badge, Use as portrait; choose one, see the toast; the header thumb follows. (AC-1.2)
3. Open a Person with no likeness, choose Gallery: the empty state. Choose Family, Timeline, Sources: each empty state. (AC-1.3, AC-1.4)
4. Back three times from Gallery after Vitals → Family → Gallery: Family, Vitals, Persons list. Choose Gallery twice: one Back step. (AC-2.1, AC-2.2)
5. On Gallery, click a tile, then Back: Gallery. On Timeline, quit and reopen: Timeline. (AC-2.6, AC-2.3)
6. Breadcrumbs on Gallery; open the Back jump menu: the section is named. (AC-2.7, AC-2.8)
7. VoiceOver on the switcher; Tab to it, arrow between chips, ⌘3 → Timeline. (AC-1.5, D-9)
8. Open an Event and a Place: unchanged.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Adding `personSection` to `==` changes history coalescing for Person entries | T-2 normalizes Vitals; T-1 covers restore |
| A pinned header (Q-1) changes `ConclusionDetailPage.page` for Event and Place | Q-1's answer must keep Event and Place unchanged (R-1); stop if it cannot |
| Chips' ⌘ shortcuts fire while a sheet is open on the page | Manual check; stop if a shortcut changes section behind a modal |

**Stop and report if:**

- `ConclusionDetailPage` no longer has the full initializer or the `pageSections` slot §4 describes.
- The agreed board needs a Person-only page type, or a change to Event or Place layout.
- `PVChip` cannot carry `.keyboardShortcut` or the selected trait without a kit change.
- The board's frames and an AC disagree.

## 12. Open questions

Answered by the S13-D1 board; the plan is refreshed then.

- **Q-1 (PPF-2, PPF-3).** Pattern A (one section at a time) or B (stacked with shortcuts), and are the header and switcher pinned while the section scrolls (user, 2026-10-29: the board decides)? Recommended: A, pinned. If pinned, STEP-3 splits `ConclusionDetailPage.page` into a fixed top (header plus an optional accessory slot, default empty) and the `ScrollView` for `pageSections`, with Event and Place passing no accessory and keeping today's look.
- **Q-2 (PPF-3).** If pinned, does the header compact on scroll, and to what?
- **Q-3 (PPF-6).** Crumb and jump-row wording.
- **Q-4 (PPF-8).** Vitals heading: "Details", "Vitals" or none, and where the refreshing spinner sits.
- **Q-5 (PPF-13).** ⌘1 – ⌘5, or other keys.

## 13. PR

**Title:** S13-10 — Header and sections
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md`.
