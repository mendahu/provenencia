# S13-10 — Header and sections

> **Update 2026-10-29: Wide layout.** Written before Spike 14's Wide layout changes; refresh against `main` before this plan is Planned.
> - **S14-24 landed** ([#628](https://github.com/mendahu/provenencia/pull/628)): every workspace page reads `@Environment(\.pvWidthClass)` (`.regular` below a 1600pt pane, `.wide` at and above), measured once at `WorkspaceDestinationHost`.
> - **S14-25 is descoped; this PR builds Page columns.** Build `PVPageColumns` from the agreed Spike 14 brief S14-D1, archived by [#629](https://github.com/mendahu/provenencia/pull/629) board (WL-3 – WL-10) and move `ConclusionDetailPage.page` onto it, so the Event and Place pages share it with no other change. Build the header slot and pinned mode only if S13-D1 uses them. Add `PVSpacing.widthSecondaryMax` (480) and check the kit's `--width-wide-breakpoint` is 1600.
> - **Owes:** Steps and a Codebase fit row for `PVPageColumns`; tests for the WL-5 sizes at panes 1102, 1600, 1656 and 2296; S13-D1's `PPF-W` requirements in Scope; a manual check at Regular and Wide, plus the Event and Place pages unchanged at Regular; **Depends on** gains S14-D1 (board agreed 2026-10-29) and S14-24.
> - Source: Spike 13 working rule "Redesigns adopt Page columns" and the S13-10 scope entry ([#629](https://github.com/mendahu/provenencia/pull/629)).

**Status:** Blocked — on board S13-D1
**Spike / slice:** Spike 13 · Slice 1 — Page frame and Vitals
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D1](../design/S13-D1-person-page-frame.md), board not agreed yet
**Board:** — (to record when agreed: the Claude Design link and frame numbers; the implementer reads §6 **macOS**, which describes each frame in words)
**Depends on:** S13-09 (landed [#624](https://github.com/mendahu/provenencia/pull/624)), S13-D1. Shares code with Spike 14's S14-02 (Blocked — re-plan); either may land first, see §4 and §11
**Blocks:** S13-11, S13-12, S13-14, S13-16
**Branch:** `s13-10-header-and-sections`
**Version bump:** none — Spike 11 – 14 PRs leave `VERSION` alone until a release is cut, as S13-09 and S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-1 – R-7, R-14, R-15; working rules: "Sections are places", "One sectioned Conclusion page, not a Person fork", "S13-D1 draws the whole frame", "Each section reads its own query"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-10 ("Reuse first", from S13-09 ship log MNT-1); [`macos-client-patterns.md`](../../../macos-client-patterns.md) Conclusion keys
**Planned:** 2026-10-29 against `main` @ `b8ca4cfc` (draft; frame-dependent Decisions wait on the board)
**Review:** [`S13-10-header-and-sections.review.md`](S13-10-header-and-sections.review.md) — Fits with changes / Not ready → every finding Fixed, 1 round; one To user item

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

**What stays.** The header's parts (thumb, gender Badge, name, b. and d. lines, ref, member count) and the thumb's right-click "Use earliest likeness" (user, 2026-10-29: stays until the S13-D5 Gallery board); every row with its Badge, Sources and Why; the likeness section as it is drawn today, its own "Likenesses" heading and count included, now inside Gallery (user, 2026-10-29); one `conclusionDetail` read; the Event and Place pages, which keep the convenience initializer and today's layout (R-1); every existing link into a Person page, which lands on Vitals.

## 2. User stories

**US-1.** As a researcher, when I open a Person, I see the header and a switcher with five sections, Vitals showing, so that I can tell where I am and where else I can go.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Open a Person from the Persons list | The header as today; the switcher lists Vitals, Family, Timeline, Gallery, Sources in that order; Vitals is selected and shows today's rows in today's order |
| AC-1.2 | Same person, choose Gallery | Today's likeness section (heading "Likenesses" with the count, tiles, Portrait Badge on the portrait tile, Use as portrait on the others); the header does not change |
| AC-1.3 | A person with no likeness, choose Gallery | `PVEmptyState` titled "No likenesses yet" (wording per board) |
| AC-1.4 | Choose Family, Timeline or Sources | `PVEmptyState` with that section's glyph and name and the line that it arrives in a later version (wording per board); no rows, no skeleton |
| AC-1.5 | VoiceOver on the switcher with Vitals showing | Reads the group as "Sections", "Vitals, selected", and each other section by name |
| AC-1.6 | Keyboard: Tab to the switcher; press ⌘3 | The focused chip shows the kit focus ring; ⌘3 shows Timeline (keys per board, Q-5) |

**US-2.** As a researcher, when I switch sections, Back and Forward walk them and a reopened project returns to the section I was on, so that a section behaves like any other place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Persons list → Jane Doe (Vitals) → Family → Gallery, then Back | Family; Back again → Vitals; Back again → Persons list |
| AC-2.2 | On Gallery, choose Gallery again | No history entry is added |
| AC-2.3 | On Timeline, quit and reopen the project | Jane Doe on Timeline |
| AC-2.4 | A history file written before this PR (no `pageSection` key) | Decodes; Person entries open on Vitals |
| AC-2.5 | A history entry with `"pageSection": "lineage"` (unknown) | Decodes; opens on Vitals |
| AC-2.6 | From Gallery, open a likeness tile's Citation, then Back | Jane Doe on Gallery |
| AC-2.7 | Breadcrumbs on Jane Doe's Gallery | Persons › Jane Doe › Gallery, Jane Doe going to Vitals; on Vitals, Persons › Jane Doe (wording per board, Q-3) |
| AC-2.8 | Back jump menu after Vitals → Family | The Vitals entry reads "Jane Doe", the Family entry "Jane Doe · Family" (wording per board, Q-3) |
| AC-2.9 | Scroll Vitals to its last row, choose Gallery | Gallery shows from its first line (PPF-7) |

**US-3.** As a contributor, when I build the next Person section, Vitals and Gallery reuse named parts of the default fill, and the section mechanism is not Person-only, so that a section is a case in one switch and Event or Place can adopt sections without a second field.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ConclusionDetailDefaultPageSections` | Built from `ConclusionDetailRowsSection` *(new)*, `ConclusionDetailLikenessSection` and the relationship `ForEach`; Event and Place look the same |
| AC-3.2 | Person's below-header fill | One `switch` over `ConclusionPageSection`; Vitals is `ConclusionDetailRowsSection`, Gallery is `ConclusionDetailLikenessSection`; no row loop copied into `PersonDetailView.swift` |
| AC-3.3 | `ConclusionPageSection.offered(by:)` | `.persons` → the five cases in order; every other `WorkspaceSection` → `[]` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionPageSection` *(new)* and `WorkspaceLocation.pageSection` *(new)*, kind-neutral: Codable, nil is the kind's first section, unknown or not offered decodes to nil, part of `==`, one normalizer (PPF-5) | STEP-1, AC-2.1 – AC-2.5, AC-3.3, D-1, D-2 |
| Rename the header's "vitals" (the b. and d. lines) to life lines, so "Vitals" names only the section | STEP-2, D-13 |
| Split the default fill into `ConclusionDetailRowsSection` *(new)* and an internal `ConclusionDetailLikenessSection` (S13-10 scope "Reuse first") | STEP-3, AC-3.1, D-4 |
| `PVChip` owns the selected trait and the kit focus ring (all chip call sites gain them) | STEP-4, AC-1.5, AC-1.6, D-6 |
| `ConclusionDetailPage` scrolls to the top when its `scrollResetKey` changes | STEP-5, AC-2.9, D-12 |
| Person's fill: switcher + one section at a time; Vitals rows, Gallery section or empty, three empty sections (PPF-1, PPF-2, PPF-4, PPF-8 – PPF-10) | STEP-6, AC-1.1 – AC-1.6, AC-3.2, D-3, D-5, D-7, D-9 |
| Switching pushes history through `navigation.go(to:)` (PPF-5) | STEP-6, AC-2.1, AC-2.2, AC-2.6, D-3 |
| One label function for crumbs and jump rows, naming the section (PPF-6) | STEP-7, AC-2.7, AC-2.8, D-8 |
| The pinned or scrolling layout the board picks (PPF-3); narrow window (PPF-12) | STEP-6 after the board, Q-1, Q-2 |
| L10n for five section names, the switcher label, the empty states and the jump-row format | STEP-1, STEP-6, STEP-7, D-10 |
| Docs: patterns bullet, `completed.md`, map status, brief archive | STEP-8 |

| Out | Home |
| --- | --- |
| Provenance control and sheet on Vitals rows | S13-11 (brief S13-D2) |
| Family, Timeline, Sources content and their queries | S13-12, S13-13, S13-16 (queries: S13-01, S13-04, S13-05) |
| Gallery redesign, viewer, where "Use earliest likeness" lives | S13-14, S13-15 (brief S13-D5) |
| The gallery view-layer rename (`ConclusionDetailGallerySection`, the `gallery` slot and its wording) | S14-02 |
| How the records connect | S13-28 |
| Sections on Event and Place pages | not this spike (R-1); D-1 leaves them a slot |
| `pageSection` on the engine's `WorkspaceLocation` proto (search context) | not planned — D-2 |
| A segmented-picker initializer on `PVChipGroup` replacing the three hand-written copies | To user (review M-2); not this PR unless the user says so |
| A Tab kit component | not planned — brief non-goal; segmented `PVChipGroup` |
| Arrow-key movement between chips | not planned — D-9 |
| Remembering the last section per person | not planned — brief non-goal (PPF-5) |
| New `.swift` file or `project.pbxproj` entry | not planned — D-11 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Page and slot | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | Three generics after S13-09; memberwise (full) init takes `pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections`; convenience init (where `PageSections == ConclusionDetailDefaultPageSections<Content>`) supplies the default fill. `static func page(identifier:body:)` wraps every state in one `ScrollView` (header included) |
| Loaded body | same file `ConclusionDetailLoaded` (private) | On `.content`: `ConclusionDetailHeader`, then `pageSections(content, context)` |
| Default fill | same file `ConclusionDetailDefaultPageSections` (internal) | One block: Details `PVSectionHeader` (`L10n.Conclusions.detailsSection`, refreshing `ProgressView` aside) + `ForEach(content.rows)` → `ReconciledValueRow`; likeness section when non-empty; `ForEach(content.sections)` → relationship section. Returns siblings that take the page `VStack`'s spacing |
| Likeness grid | same file `ConclusionDetailLikenessSection` (private) | Heading "Likenesses" with count, adaptive grid, tile opens `tile.location` via `navigation.go(to:)`, Portrait Badge / Use as portrait from `ConclusionPortraitModel`; uses `@FocusState` + `.pvFocusRing` on tiles |
| Header | same file `ConclusionDetailHeader` (private) | Thumb (with "Use earliest likeness" right-click menu when `offersEarliest`), Badge, title, summary, ref, members |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView` | Convenience init; `summary: { PersonDetailVitals(vitals: $0.vitals) }` (private view, the b. and d. lines); `.vocabularyToastOverlay($portrait.toast, …)`; does not read `WorkspaceNavigation` |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent` | `struct Vital`, `var vitals: [Vital]`, `static func vital(…) -> Vital` = the header's b. and d. lines; `rows`, `likenesses`, `thumbnail`, `badge` |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Conclusions.a11yVital(_:date:place:)`; `macos/App/Resources/Localizable.xcstrings` | Spoken b./d. line; no section keys |
| Location | `macos/App/Features/Workspace/WorkspaceLocation.swift` `WorkspaceLocation` | `entityId` names the person; memberwise `init(section:…listScrollAnchor:)` with one parameter per field; tolerant `init(from:)` assigns fields directly; `personDetail(entityId:ref:title:)`; `withListScrollAnchor` copies `self`; `SourceSurface` enum and its `normalizedSurface` live in this file; `==` lists deep ids |
| Sections | `macos/App/Features/Workspace/WorkspaceSection.swift` `WorkspaceSection.glyph` | Sources `.symbol(.library)`, Properties `.symbol(.listTree)` |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` spec `.personDetail` | `viewIdentity: { $0.entityId }` (a section change does not remount); `queryKeys: { project, location in … }` already receives the location |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.personDetail` | `PersonDetailView(session:entityId: navigation.currentLocation.entityId ?? "", userID:)` |
| Navigation | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `go(to:)` | Coalesces when `location == currentLocation`; otherwise pushes and persists |
| Crumbs, jump rows | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems(for:names:goTo:)`, private `HistoryJumpRow.leafLabel` | Two derivations of one label: the crumb leaf is `title ?? ref ?? "…"` with id `"leaf-\(section.rawValue)"` and `action: nil`; the jump row is `title ?? ref` |
| Proto | `api/proto/engine.proto` `WorkspaceLocation`; `macos/App/Platform/GoStore.swift` `mapWorkspaceLocationToProto` / `FromProto` | Sent as search context only (`searchCatalog`); `listScrollAnchor` is not on the wire |
| Kit chip | `macos/App/DesignSystem/Components/Chip/PVChip.swift` `PVChip`, `PVChipGroup(style: .segmented)` | `PVChip` is a `Button` with `isSelected`, but adds no `.isSelected` trait and no focus ring. Segmented copies: `Features/Dates/DateValueEditorForm.swift`, `Features/CitationComposer/CitationComposerAuthoringBox.swift`, `Features/Properties/PropertiesView.swift` |
| Kit empty | `macos/App/DesignSystem/Components/EmptyState/` `PVEmptyState` | `init(glyph: PVGlyph, title: LocalizedStringResource?, message: String?, compact:, action:)`; `message` is a `String` |
| Scroll precedent | `WorkspaceNavigation.swift` `ListScrollAnchorModifier`; `Features/Sources/SourcesListView.swift` `LazyVStack(pinnedViews: [.sectionHeaders])` | Scroll-to-anchor and pinned headers inside one scroll |
| Spike 14 | `docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` (Blocked — re-plan) | D-1 replaces `likenesses` with a `gallery` slot carrying its wording; D-3 renames `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection` |
| Shortcuts | `rg keyboardShortcut macos/App` | ⌘1 – ⌘9 unused |
| Tests | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift`, `PersonDetailTests.swift` (lines 268, 269, 318 read `vitals`), `EventDetailTests.swift`, `PlaceDetailTests.swift` | Push/restore, legacy decode (`legacyHistoryMissingListScrollAnchorDecodesAsNil`), equality, crumbs (`conclusionDetailCrumbPrefersTheTitle`); Person content |

**Precedent.** `listScrollAnchor` (S10) added an optional, Swift-only, tolerant-decoding field with a copy-with helper and legacy-decode tests: copy its shape for `pageSection`, but the section **is** identity (it is in `==`). `SourceSurface` + `normalizedSurface` is the precedent for a small enum beside `WorkspaceLocation` with one private normalizer every construction path calls. S13-09 ([#624](https://github.com/mendahu/provenencia/pull/624)) is the precedent for the slot: the full init with a custom `pageSections`, returning siblings. `ConclusionDetailLikenessSection`'s `@FocusState` + `.pvFocusRing` is the focus-ring precedent for `PVChip`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `enum ConclusionPageSection: String, Codable, CaseIterable, Sendable` *(new)* in `WorkspaceLocation.swift` beside `SourceSurface`, cases `vitals, family, timeline, gallery, sources` in that order. Raw values are frozen: they are written to history files. `static func offered(by section: WorkspaceSection) -> [ConclusionPageSection]` returns all five for `.persons` and `[]` for every other section. It owns `label: LocalizedStringResource` (D-10) and `glyph: PVGlyph`: vitals `.symbol(.list)`, family `.symbol(.gitBranch)`, timeline `.symbol(.calendar)`, gallery `.symbol(.photo)`, sources `.symbol(.library)` (as `WorkspaceSection.sources`) | Kind-neutral (review M-1): Event or Place adopts sections by returning its cases from `offered(by:)` and adding any case of its own to the enum, with no second field, key, `==` line or label branch. Glyphs are existing `PVSymbol`s (review X-3, M-8) | A Person-only enum and field (a per-type field the map's named next kinds would repeat); new marks |
| D-2 | `WorkspaceLocation.pageSection: ConclusionPageSection?` *(new)*, coding key `pageSection`, memberwise `init` parameter `pageSection: ConclusionPageSection? = nil` (after `entityId`). One private `static func normalizedPageSection(_:section:entityId:) -> ConclusionPageSection?` returns nil when `entityId` is nil, when the value is not in `offered(by: section)`, or when it equals `offered(by: section).first` (the default section); otherwise the value. `init`, `init(from:)` (which decodes a `String?` and maps it with `ConclusionPageSection(rawValue:)`, unknown → nil) and `withPageSection(_:)` *(new)* all call it. `withPageSection` copies `self`, sets the field through the normalizer, and keeps `ref`, `title`, `sourceTitle`. `==` compares `pageSection`. `personDetail(entityId:ref:title:)` gains `section: ConclusionPageSection? = nil`. Not on the proto | Sections are places (working rule); one normalizer for every path (review X-1, `normalizedSurface` precedent); Vitals is stored as nil, so Vitals-by-argument equals Vitals-by-default and `go(to:)` coalesces; existing links land on Vitals; old and unknown history decodes | Storing `.vitals` explicitly (two equal-looking locations unequal); normalizing only in `init`; adding it to `engine.proto` (search context does not need it; `listScrollAnchor` precedent) |
| D-3 | `PersonDetailView` reads `@Environment(WorkspaceNavigation.self)`. The section is `navigation.currentLocation.pageSection ?? .vitals`; choosing one calls `navigation.go(to: navigation.currentLocation.withPageSection(choice))`. No local `@State` for the section. `.vocabularyToastOverlay` and the `portrait` argument stay | `add-workspace-location`: the location is the one source of where we are; `go(to:)` coalesces the same section (AC-2.2); `viewIdentity` stays `entityId`, so `ConclusionPortraitModel` and its toast survive a switch | A `@State` section synced to the location; the host passing the section in |
| D-4 | Split `ConclusionDetailDefaultPageSections` into named internal parts in the same file: `ConclusionDetailRowsSection` *(new)* with `rows: [ReconciledValueRowModel]`, `refreshing: Bool`, `title: LocalizedStringResource?` (nil draws no heading and puts the refreshing spinner where Q-4 says), and `ConclusionDetailLikenessSection` changed from `private` to internal with no other change. The default fill becomes `ConclusionDetailRowsSection(rows: content.rows, refreshing: context.refreshing, title: L10n.Conclusions.detailsSection)`, the likeness `if`, and the relationship `ForEach`, with unchanged output | Map scope "Reuse first" (S13-09 MNT-1); Vitals and Gallery reuse by name; optional title covers every Q-4 answer (review X-7) | Copying the heading, spinner and rows into the Person view; flags on the whole default fill |
| D-5 | Person passes its own fill through `ConclusionDetailPage`'s full initializer with `pageSections: { content, context in PersonDetailSections(content: content, context: context, section: section, select: select) }` and `scrollResetKey: navigation.currentLocation.pageSection` (D-12). `PersonDetailSections` *(new, private in `PersonDetailView.swift`)* returns sibling views: the switcher, then one `switch section`: `.vitals` → `ConclusionDetailRowsSection(rows:refreshing:title:)` (title per Q-4); `.gallery` → `ConclusionDetailLikenessSection(projectDir:likenesses:portrait:)` when `content.likenesses` is non-empty, else `PVEmptyState(glyph: .symbol(.photo), title: L10n.Conclusions.galleryEmptyTitle, message: L10n.string(L10n.Conclusions.galleryEmptyMessage))`; `.family`, `.timeline`, `.sources` → `PVEmptyState(glyph: section.glyph, title: section.label, message: L10n.string(L10n.Conclusions.pageSectionLater))` | Working rule "One sectioned Conclusion page"; S13-09's slot; siblings take the page's spacing (S13-09 F-2); `PVEmptyState.message` is a `String` (review X-9) | A Person-only page type; a wrapper `VStack` around the fill |
| D-6 | `PVChip` gains, for every caller: `.accessibilityAddTraits(isSelected ? .isSelected : [])` on its `Button`, and a `@FocusState private var isFocused: Bool` with `.focused($isFocused)` and `.pvFocusRing(isFocused, cornerRadius: chipRadius)` (the likeness tile precedent). The switcher is `PVChipGroup(style: .segmented)` with one `PVChip(section.label, isSelected: section == current, expands: true, selectionLift: true) { select(section) }` per `ConclusionPageSection.offered(by: .persons)`, each with `.accessibilityIdentifier("persons.detail.section.\(section.rawValue)")`; the group has `.accessibilityElement(children: .contain)` and `.accessibilityLabel(Text(L10n.Conclusions.pageSectionsLabel))`, as `PropertiesView` labels its segmented group | The selected trait and focus ring belong to the kit control (review M-2; `PVRadio`, `PVSidebarNav` own their trait), so the ten existing chip call sites gain them too; brief inventory marks `PVChip` Extend | Traits added per call site; a new Tab component; a native `TabView` or `Picker(.segmented)` |
| D-7 | Each empty section is one `PVEmptyState(glyph:title:message:)`; Family, Timeline and Sources share one message key | User, 2026-10-29: placeholders show an empty state; brief PPF-10 | Hiding unbuilt sections (user declined); skeleton rows |
| D-8 | `WorkspaceToolbar` gains `static func locationLabels(for location: WorkspaceLocation) -> (leaf: String, section: String?)` *(new)*: `leaf` is `title ?? ref ?? "…"`, `section` is `L10n.string(location.pageSection.label)` when `pageSection != nil`. `breadcrumbItems` uses it: with a section, the entity crumb gets id `"entity-\(location.section.rawValue)"` and `action: { goTo(location.withPageSection(nil)) }`, and a third crumb with id `"leaf-\(location.section.rawValue)"`, the section label and `action: nil` follows; without a section the crumbs are as today. `HistoryJumpRow.leafLabel` returns `L10n.Conclusions.pageSectionLeaf(leaf, section)` when `section != nil`, else `leaf` (so the Vitals entry reads the name alone) | One label rule for both (review M-5, X-5); the Vitals entry reads like today; testable without opening `HistoryJumpRow` (review X-6) | Two derivations edited separately; "Jane Doe · Vitals" (Vitals is stored as nil, D-2) |
| D-9 | ⌘1 – ⌘5 are `.keyboardShortcut(KeyEquivalent(Character(String(index + 1))), modifiers: .command)` on each switcher chip, `index` from `offered(by: .persons).enumerated()`, so they act only while the Person page shows. Keyboard focus moves between chips with Tab (the macOS default for buttons); no arrow-key handling | Shortcuts scoped by the view, not `ProvenenciaApp.commands` (the skill warns Scene commands freeze their disabled state). Arrow keys would need custom key handling for one picker (review X-4); the brief's PPF-13 now asks for Tab | App-menu commands; `onKeyPress` arrow handling |
| D-10 | New copy in `L10n.Conclusions`, each with a `comment:` per `add-localized-string`: `pageSectionVitals` "Vitals", `pageSectionFamily` "Family", `pageSectionTimeline` "Timeline", `pageSectionGallery` "Gallery", `pageSectionSources` "Sources" (keys `conclusions.section.<raw>`), `pageSectionsLabel` "Sections" (`conclusions.sections`), `pageSectionLater` "This section arrives in a later version." (`conclusions.section.later`), `galleryEmptyTitle` "No likenesses yet" (`conclusions.gallery.empty.title`), `galleryEmptyMessage` "A likeness is drawn on a Citation's image." (`conclusions.gallery.empty.message`), `pageSectionLeaf(_ title: String, _ section: String) -> String` "%@ · %@" (`conclusions.section.leaf`). The gallery keys are named so S14-02 can move them onto its `gallery` slot unchanged | `add-localized-string`; one namespace with the page's other copy | A new namespace for one page |
| D-11 | No new `.swift` file: `ConclusionPageSection` in `WorkspaceLocation.swift`, `ConclusionDetailRowsSection` in `ConclusionDetailPage.swift`, `PersonDetailSections` in `PersonDetailView.swift`, `locationLabels` in `WorkspaceToolbar.swift` | `project.pbxproj` is hand-managed until S14-15; S13-09 D-4 precedent | New files and pbxproj entries |
| D-12 | `ConclusionDetailPage` gains `var scrollResetKey: ConclusionPageSection? = nil` (memberwise; the convenience init leaves it nil, so Event and Place are unchanged). `page(identifier:scrollResetKey:body:)` wraps its `ScrollView` in a `ScrollViewReader`, gives the content `VStack` `.id(Self.topAnchor)` (`private static let topAnchor = "conclusion.detail.top"`), and on `.onChange(of: scrollResetKey)` calls `proxy.scrollTo(Self.topAnchor, anchor: .top)`. `ConclusionDetailLoaded` and the loading branch pass it through | The page owns the one `ScrollView`, so it owns the reset (review M-3, X-2); kind-neutral; scroll stays out of history (`add-workspace-location`) | `.id(section)` on the fill (does not reset the offset); storing a scroll anchor per section; a second slot |
| D-13 | Rename the header's b. and d. lines from "vital" to "life line": `PersonDetailContent.Vital` → `LifeLine`, `vitals` → `lifeLines`, `static func vital(…)` → `lifeLine(…)`, private `PersonDetailVitals` → `PersonDetailLifeLines` (its `vitals` property → `lines`), `L10n.Conclusions.a11yVital` → `a11yLifeLine` (Swift symbol only; the string key does not change), and the three reads in `PersonDetailTests.swift` | One word, one concept, in the file that now has a Vitals section (review M-6) | Two meanings of "vitals" side by side |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionPageSection` *(new)* | type | `SourceSurface` (same file) | A different axis; extends the same pattern |
| `WorkspaceLocation.pageSection`, `withPageSection(_:)`, `normalizedPageSection` *(new)* | field, funcs | `listScrollAnchor`, `withListScrollAnchor(_:)`, `normalizedSurface` | Same shapes; the section is identity |
| `ConclusionDetailRowsSection` *(new)* | view | the Details block inside `ConclusionDetailDefaultPageSections` | Extracted, not added beside it (D-4) |
| `PersonDetailSections` *(new, private)* | view | `ConclusionDetailDefaultPageSections` | Person's fill differs by section; reuses the default's parts |
| `ConclusionDetailPage.scrollResetKey` *(new)* | parameter | `ListScrollAnchorModifier` | The page owns its `ScrollView`; a one-line reset, not a list anchor |
| `WorkspaceToolbar.locationLabels(for:)` *(new)* | func | `breadcrumbItems` leaf, `HistoryJumpRow.leafLabel` | Replaces two derivations with one |
| `PVChip` trait and focus ring | kit extend | `PVRadio`, `PVSidebarNav` (trait); likeness tile (focus ring) | Extends the one chip; no new control |
| L10n keys (D-10) *(new)* | key | `L10n.Conclusions.detailsSection`, `likenessesSection` | New names; same namespace |
| Tests (§8) *(new)* | test | `legacyHistoryMissingListScrollAnchorDecodesAsNil`, `conclusionDetailCrumbPrefersTheTitle` | Same suites, new cases |

**Builds on.** S13-09's `pageSections` slot and full initializer; `add-workspace-location` (one location, `go(to:)`, tolerant decode, one normalizer); the kit's segmented chip group and `PVEmptyState`.
**Deletes.** The single-block body of `ConclusionDetailDefaultPageSections` (rebuilt from its parts); `private` on `ConclusionDetailLikenessSection`; the second leaf derivation in `HistoryJumpRow.leafLabel` (now calls `locationLabels`); the names `Vital`, `vitals`, `vital(…)`, `PersonDetailVitals`, `a11yVital` (renamed, D-13); `PersonDetailView`'s use of the convenience initializer. No callee loses its last caller.
**Next adds.** S13-11 adds the provenance control inside the Vitals rows (Person only, by caller). S13-12, S13-13 and S13-16 each replace one empty-state case in `PersonDetailSections`; each section's query key is warmed from the `.personDetail` spec's `queryKeys` in `PlaceRegistry`, which already receives the location (working rule "Each section reads its own query"), and the section view reads a `QueryHandle` rather than starting its own load. S13-14 replaces the Gallery case. S13-28 adds a case or a view inside Sources (its board decides). Event and Place adopt sections by returning cases from `ConclusionPageSection.offered(by:)` and passing `scrollResetKey`, with no new field. S14-02 renames `ConclusionDetailLikenessSection` and moves the two gallery keys onto its slot.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A location with a section for a person that was deleted | `ConclusionDetailPage` failed state | Same as today's Vitals: the danger Callout; no switcher |
| Restoring a non-Vitals section from history on launch | `PersonDetailView` | Section read from `currentLocation` on every render; no apply step (D-3) |
| Switching person (a family tile later) while on Gallery | `PlaceRegistry` remounts on `entityId` | A new person's location has no section → Vitals (D-2) |
| Gallery with no likenesses | `PersonDetailSections` | Empty state (AC-1.3) |
| Every `PVChip` caller now carries the selected trait and a focus ring | the ten chip call sites (date kind, credibility, Properties, composer) | VoiceOver gains "selected"; the ring shows only with keyboard focus |

### Scale

None — no write and no new read: switching re-renders from the one cached `conclusionDetail`. A person with 20 rows and 30 likenesses renders the same views as today, one section at a time.

### Edges

- **Assumptions.** `go(to:)` coalesces equal locations, so `==` includes `pageSection` (T-3) and every construction path normalizes the default section to nil (T-2).
- **Neighbors.** Event, Place and list-root locations never carry a section, built or decoded (T-4). `withPageSection(.vitals)` equals `withPageSection(nil)` (T-2).
- **Removed inputs.** None — no data is decided or deleted.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The history file is app-written; an unknown string decodes to nil (T-5).

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — `pageSection` stays off `engine.proto` (D-2).

### macOS

From the brief's inventory. Frame descriptions are added when the board is agreed (Q-1 – Q-5).

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Page | `ConclusionDetailPage` full init, `scrollResetKey` | `ConclusionDetailPage.swift` | Person only; Event and Place keep the convenience init |
| Header | `ConclusionDetailHeader` | same | Unchanged unless Q-2 adds a compact form |
| Switcher | `PVChipGroup(.segmented)` + `PVChip` (extended) | `PersonDetailView.swift` `PersonDetailSections`; `PVChip.swift` | D-6, D-9 |
| Vitals | `ConclusionDetailRowsSection` | `ConclusionDetailPage.swift` | D-4, Q-4 |
| Gallery | `ConclusionDetailLikenessSection` / `PVEmptyState` | same / kit | D-5 |
| Family, Timeline, Sources | `PVEmptyState` | kit | D-7 |
| Crumbs, jump rows | `PVBreadcrumbs`, `HistoryJumpRow` via `locationLabels` | `WorkspaceToolbar.swift` | D-8 |

States: loading and failed as today (no switcher until content); refreshing spinner per Q-4. Accessibility: chips are buttons with names and the selected trait; empty states read title then message.

### Strings

L10n keys *(new)*: as D-10; final wording per board. Renamed Swift symbol (key unchanged): `L10n.Conclusions.a11yVital` → `a11yLifeLine`. Removed: none (`detailsSection` stays for Event, Place and the default fill).

## 7. Steps

**STEP-1 — The section is part of the location**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-workspace-location`, `add-swift-test`, `add-localized-string`.
Do: D-1 (with the five section-name keys of D-10), D-2. Write T-1 – T-5, T-11.
Check: `xcodebuild test` runs `WorkspaceNavigationTests` green; `python3 scripts/check-localizable-xcstrings.py` clean.

**STEP-2 — "Vitals" names only the section**
Files: `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-13.
Check: `rg -w 'Vital|vitals|a11yVital|PersonDetailVitals' macos/App macos/ProvenenciaTests` finds nothing; `PersonDetailTests` green.

**STEP-3 — Split the default fill into named parts**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`. Skill: `add-ui-component`.
Do: D-4.
Check: `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` green; Event and Place unchanged in the app.

**STEP-4 — Chips carry their selected state and focus**
Files: `macos/App/DesignSystem/Components/Chip/PVChip.swift`. Skill: `add-ui-component`.
Do: D-6, the `PVChip` half only.
Check: the app builds; with VoiceOver, the date-kind picker reads its selected chip as selected.

**STEP-5 — The page scrolls to the top when told**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`.
Do: D-12.
Check: Event and Place unchanged (they pass nil).

**STEP-6 — Person sections and the switcher** (after the board; Q-1 – Q-5 answered in this plan)
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-3, D-5, D-6 (switcher half), D-7, D-9, D-10; the layout the board picks (Q-1, Q-2).
Check: T-10; manual check §10 steps 1 – 5 and 7.

**STEP-7 — Crumbs and jump rows name the section**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: D-8. Write T-6, T-7.
Check: `WorkspaceToolbarBreadcrumbTests` green.

**STEP-8 — Docs**
Files: `docs/macos-client-patterns.md` (one sentence under the sectioned detail page bullet: a kind passes its own fill and its section lives on `WorkspaceLocation.pageSection`, offered per kind by `ConclusionPageSection.offered(by:)`), `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S13-D1-person-page-frame.md` → `../design/archive/` *(new)*, `../design/README.md`.
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py` on this plan.

Commits: one per behavior, full imperative sentences ("Remember which section of a Person page is open in history.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `WorkspaceNavigationTests` `pageSectionPushAndRestore` *(new)* | go Vitals → Family → Gallery; back, back; persist, reattach → index and sections restored | AC-2.1, AC-2.3 |
| T-2 | Swift | `WorkspaceNavigationTests` `firstSectionIsStoredAsNil` *(new)* | `.personDetail(…section: .vitals)`, `.personDetail(…)` and `withPageSection(.vitals)` are equal with `pageSection == nil`; going to the same section does not push | AC-2.2, D-2 |
| T-3 | Swift | `WorkspaceNavigationTests` `pageSectionIsIdentity` *(new)* | same person, Family vs Gallery → not equal | D-2 |
| T-4 | Swift | `WorkspaceNavigationTests` `pageSectionOnlyWhereOffered` *(new)* | event detail, place detail and persons list root, each built, decoded and `withPageSection(.family)` → nil | D-1, D-2 |
| T-5 | Swift | `WorkspaceNavigationTests` `legacyAndUnknownPageSectionDecodeAsTheFirst` *(new)* | JSON without the key; JSON with `"lineage"` → nil | AC-2.4, AC-2.5 |
| T-6 | Swift | `WorkspaceToolbarBreadcrumbTests` `pageSectionAddsACrumb` *(new)* | Gallery → three crumbs, the entity crumb's action goes to the Vitals location; Vitals → two crumbs as today | AC-2.7, D-8 |
| T-7 | Swift | `WorkspaceToolbarBreadcrumbTests` `locationLabelsNameTheSection` *(new)* | Family with title "Jane Doe" → ("Jane Doe", "Family"); Vitals → ("Jane Doe", nil); no title, ref "PER-1" → leaf "PER-1" | AC-2.8, D-8 |
| T-8 | Swift | `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` existing (renamed reads per D-13) | unchanged | AC-3.1, D-13 |
| T-9 | Manual | §10 | the five sections, empty states, VoiceOver, keyboard, scroll to top | AC-1.1 – AC-1.6, AC-2.6, AC-2.9 |
| T-10 | Structural | STEP-6 `rg` | In `macos/App/Features/Conclusions/PersonDetailView.swift`: `rg -F 'ForEach(content.rows)'` finds nothing; `rg -F 'ConclusionDetailRowsSection'` and `rg -F 'ConclusionDetailLikenessSection'` each find a hit | AC-3.2, D-5 |
| T-11 | Swift | `WorkspaceNavigationTests` `sectionsOfferedPerKind` *(new)* | `offered(by: .persons)` → the five in order; `.events`, `.places`, `.sources` → `[]` | AC-3.3, D-1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceNavigationTests` (`==` change), the detail suites (STEP-2, STEP-3), any test that compares chip accessibility traits (STEP-4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`PersonDetailContent.Vital`, `PersonDetailVitals`, `a11yVital`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md (a kind's own fill; pageSection)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Person that has a name, both life lines, three likenesses and a Mixed row. Open them from the Persons list: header as before, Vitals selected, rows as before. (AC-1.1)
2. Choose Gallery: today's likeness section with its heading, Portrait Badge, Use as portrait; choose one, see the toast; the header thumb follows. (AC-1.2)
3. Open a Person with no likeness, choose Gallery: the empty state. Choose Family, Timeline, Sources: each empty state with its glyph. (AC-1.3, AC-1.4)
4. Back three times from Gallery after Vitals → Family → Gallery: Family, Vitals, Persons list. Choose Gallery twice: one Back step. (AC-2.1, AC-2.2)
5. On Gallery, click a tile, then Back: Gallery. On Timeline, quit and reopen: Timeline. Scroll Vitals to the bottom, choose Gallery: it shows from the top. (AC-2.6, AC-2.3, AC-2.9)
6. Breadcrumbs on Gallery; click the person crumb → Vitals; open the Back jump menu: "Jane Doe · Family" and "Jane Doe". (AC-2.7, AC-2.8)
7. VoiceOver on the switcher; Tab to it (focus ring), ⌘3 → Timeline. Open the date editor: its kind chips read "selected". (AC-1.5, AC-1.6)
8. Open an Event and a Place: unchanged.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Adding `pageSection` to `==` changes coalescing for Person entries | T-2 normalizes the first section; T-1 covers restore |
| A pinned header (Q-1) changes the page for Event and Place | Q-1 keeps one `ScrollView` and one slot; stop if the board needs a split |
| A ⌘ shortcut on a chip fires while a sheet is open on the page | Manual check; stop if a shortcut changes section behind a modal |
| S14-02 lands first and renames `ConclusionDetailLikenessSection` and `likenesses` | Refresh §4 and D-4, D-5 to its names before STEP-3; the Gallery case calls the gallery section by its new name |
| The `PVChip` focus ring shows in a picker where it looks wrong | The ring follows keyboard focus only; stop and report the call site rather than adding an opt-out |

**Stop and report if:**

- `ConclusionDetailPage` no longer has the full initializer or the `pageSections` slot §4 describes.
- The agreed board needs a Person-only page type, a second slot, a split `ScrollView`, or a change to Event or Place layout.
- `PVChip` cannot carry `.focused`, `.keyboardShortcut` or the selected trait without changing its public initializers.
- The board's frames and an AC disagree.

## 12. Open questions

Answered by the S13-D1 board; the plan is refreshed and reviewed again then.

- **Q-1 (PPF-2, PPF-3).** Pattern A (one section at a time) or B (stacked with shortcuts), and are the header and switcher pinned while the section scrolls (user, 2026-10-29: the board decides)? Recommended: A. If pinned, the pinning stays inside the page's one `ScrollView` and the one `pageSections` slot, the way `SourcesListView` pins section headers (`LazyVStack(pinnedViews: [.sectionHeaders])`). A board that needs a second slot or a split `ScrollView` goes to the user before this plan changes.
- **Q-2 (PPF-3).** If pinned, does the header compact on scroll, and to what?
- **Q-3 (PPF-6).** Crumb and jump-row wording (structure fixed by D-8).
- **Q-4 (PPF-8).** Vitals heading: "Details", "Vitals" or none (D-4's `title`), and where the refreshing spinner sits.
- **Q-5 (PPF-13).** ⌘1 – ⌘5, or other keys.

## 13. PR

**Title:** S13-10 — Header and sections
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md`.
