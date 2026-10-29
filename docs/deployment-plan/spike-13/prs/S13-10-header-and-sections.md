# S13-10 — Header and sections

**Status:** Blocked — on board S13-D1
**Spike / slice:** Spike 13 · Slice 1 — Page frame and Vitals
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D1](../design/S13-D1-person-page-frame.md), board not agreed yet; [S14-D1](../../spike-14/design/archive/S14-D1-wide-layout.md) (Page columns), board agreed 2026-10-29
**Board:** S13-D1 — (to record when agreed: the Claude Design link and frame numbers; the implementer reads §6 **macOS**, which describes each frame in words). S14-D1 — [Claude Design](https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445), the Page columns specimen; its requirements WL-3 – WL-9 are copied into §5 D-14
**Depends on:** S13-09 (landed [#624](https://github.com/mendahu/provenencia/pull/624)), S14-24 (landed [#628](https://github.com/mendahu/provenencia/pull/628)), S14-D1 (agreed), S13-D1. Shares code with Spike 14's S14-02 (Blocked — re-plan); either may land first, see §4 and §11
**Blocks:** S13-11, S13-12, S13-14, S13-16
**Branch:** `s13-10-header-and-sections`
**Version bump:** none — Spike 11 – 14 PRs leave `VERSION` alone until a release is cut, as S13-09 and S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-1 – R-7, R-14, R-15; working rules: "Sections are places", "One sectioned Conclusion page, not a Person fork", "S13-D1 draws the whole frame", "Redesigns adopt Page columns", "Each section reads its own query"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-10 ("Reuse first", from S13-09 ship log MNT-1; "Page columns"); [`macos-client-patterns.md`](../../../macos-client-patterns.md) Conclusion keys and [§ 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); Spike 14 working rule "(Batch 6) One breakpoint, 1600pt of pane width"
**Planned:** 2026-10-29 against `main` @ `4ac9305f` (draft; frame-dependent Decisions wait on the S13-D1 board)
**Review:** [`S13-10-header-and-sections.review.md`](S13-10-header-and-sections.review.md) — round 1 Fits with changes / Not ready → every finding Fixed; the Page columns and Wide rework since then is unreviewed, round 2 runs on it with the board's answers

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

The Person page keeps its header and gains a section switcher under it: Vitals, Family, Timeline, Gallery, Sources. Vitals holds today's rows; Gallery holds today's likeness grid; Family, Timeline and Sources say they arrive later. Each section is a place: switching pushes history, Back returns to the previous section, and a restored session reopens on it. The page is laid out on the kit's new Page columns (built here from the S14-D1 board), so on a wide monitor it uses the extra width as S13-D1 draws, and the Event and Place pages share the same scaffold.

**Why now.** Slice 1's chrome PR; every later Person section (S13-11, S13-12, S13-14, S13-16) fills this frame. Working rules "Sections are places", "S13-D1 draws the whole frame" and "Redesigns adopt Page columns" (S14-25 was folded into this PR, user 2026-10-29).

**What stays.** The header's parts (thumb, gender Badge, name, b. and d. lines, ref, member count) and the thumb's right-click "Use earliest likeness" (user, 2026-10-29: stays until the S13-D5 Gallery board); every row with its Badge, Sources and Why; the likeness section as it is drawn today, its own "Likenesses" heading and count included, now inside Gallery (user, 2026-10-29); one `conclusionDetail` read; the Event and Place pages' look at Regular (R-1); every existing link into a Person page, which lands on Vitals.

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

**US-3.** As a researcher on a wide monitor, when I open a Person, Event or Place, the page uses the width as the boards draw it and never sits jammed to one side, so that a big window is worth having (brief PPF-W1 – PPF-W5; S14-D1 WL-4 – WL-8).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A 2560 × 1440 window, sidebar open (pane 2296), a Person on Vitals | The header and switcher span the column group; Vitals in the primary column at most 1180pt; the secondary column as PPF-W2 decides (Q-6); the group centered, gutters at least 32pt |
| AC-3.2 | Same window, Gallery | More tiles per row at the same tile size (PPF-W3, Q-6) |
| AC-3.3 | Same window, Family | The empty state centered in a primary-only column (PPF-W4) |
| AC-3.4 | A 1366 × 900 window (pane 1102), any Person section | One centered column up to 1180pt inside 32pt gutters; anything the Wide secondary column holds is stacked below the primary, in the same order (rule 4) |
| AC-3.5 | On Vitals scrolled halfway with the switcher focused, resize the window across 1600pt of pane | Still Vitals, the scroll position kept, the switcher still focused (PPF-W5, rule 5) |
| AC-3.6 | Open an Event and a Place at 1366 × 900 and at 2560 × 1440 | At 1366, the same page as before this PR; at 2560, the same column centered (WL-6) |

**US-4.** As a contributor, when I build the next Person section or move another page onto Page columns, Vitals and Gallery reuse named parts of the default fill, sections are not Person-only, and the column sizing has one tested owner.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `ConclusionDetailDefaultPageSections` | Built from `ConclusionDetailRowsSection` *(new)*, `ConclusionDetailLikenessSection` and the relationship `ForEach`; Event and Place look the same |
| AC-4.2 | Person's below-header fill | One `switch` over `ConclusionPageSection`; Vitals is `ConclusionDetailRowsSection`, Gallery is `ConclusionDetailLikenessSection`; no row loop copied into `PersonDetailView.swift` |
| AC-4.3 | `ConclusionPageSection.offered(by:)` | `.persons` → the five cases in order; every other `WorkspaceSection` → `[]` |
| AC-4.4 | `PVPageColumns.columnWidths(pane:widthClass:hasSecondary:)` | Pane 1102, Regular, secondary → primary 1038, secondary stacked at 1038. Pane 1600, Wide, secondary → 1164 / 340. Pane 1656 → 1180 / 380. Pane 2296 → 1180 / 480, group 1692. Pane 2296, no secondary → primary 1180 (WL-4 – WL-7) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionPageSection` *(new)* and `WorkspaceLocation.pageSection` *(new)*, kind-neutral: Codable, nil is the kind's first section, unknown or not offered decodes to nil, part of `==`, one normalizer (PPF-5) | STEP-1, AC-2.1 – AC-2.5, AC-4.3, D-1, D-2 |
| Rename the header's "vitals" (the b. and d. lines) to life lines, so "Vitals" names only the section | STEP-2, D-13 |
| `PVSpacing.widthSecondaryMax` = 480 (S14-D1 WL-1); `widthWideBreakpoint` checked against the board's 1600 | STEP-3, D-15 |
| `PVPageColumns` *(new)* in the Swift kit: header slot (scrolling, and pinned per Q-1), primary, optional secondary; owns the scroll and the scroll reset; sizing per WL-4 – WL-7; focus order per WL-9; kit README row (S14-D1 WL-3 – WL-9) | STEP-3, AC-4.4, AC-3.4, AC-3.5, D-12, D-14 |
| `ConclusionDetailPage.page` moves onto `PVPageColumns`: header in the header slot, loading and error in the primary; Event and Place unchanged at Regular, centered at Wide | STEP-4, AC-3.6, D-16 |
| Split the default fill into `ConclusionDetailRowsSection` *(new)* and an internal `ConclusionDetailLikenessSection` (S13-10 scope "Reuse first") | STEP-5, AC-4.1, D-4 |
| `PVChip` owns the selected trait and the kit focus ring (all chip call sites gain them) | STEP-6, AC-1.5, AC-1.6, D-6 |
| The page draws a kind's section switcher in the header slot from `ConclusionDetailSectioning` *(new)* data, and resets scroll on a section change | STEP-7, AC-2.9, D-5, D-6, D-12 |
| Person's sections: Vitals rows, Gallery section or empty, three empty sections (PPF-1, PPF-2, PPF-4, PPF-8 – PPF-10); history through `navigation.go(to:)` (PPF-5) | STEP-8, AC-1.1 – AC-1.6, AC-2.1, AC-2.2, AC-2.6, AC-4.2, D-3, D-5, D-7, D-9 |
| Person's Wide arrangement (PPF-W1 – PPF-W5) | STEP-8, AC-3.1 – AC-3.5, Q-6 |
| One label function for crumbs and jump rows, naming the section (PPF-6) | STEP-9, AC-2.7, AC-2.8, D-8 |
| The pinned or scrolling header the board picks (PPF-3); narrow window (PPF-12) | STEP-3, STEP-8 after the board, Q-1, Q-2 |
| L10n for five section names, the switcher label, the empty states and the jump-row format | STEP-1, STEP-8, STEP-9, D-10 |
| Docs: patterns bullet, `completed.md`, map status, brief archive, `docs/ideas/wide-layout.md` Person row | STEP-10 |

| Out | Home |
| --- | --- |
| Provenance control and sheet on Vitals rows | S13-11 (brief S13-D2) |
| Family, Timeline, Sources content and their queries | S13-12, S13-13, S13-16 (queries: S13-01, S13-04, S13-05) |
| Gallery redesign, viewer, where "Use earliest likeness" lives; a full-width grid mode | S13-14, S13-15 (brief S13-D5) |
| The gallery view-layer rename (`ConclusionDetailGallerySection`, the `gallery` slot and its wording) | S14-02 |
| How the records connect | S13-28 |
| Sections and Wide arrangements on Event and Place pages | not this spike (R-1); D-1 and D-16 leave them a slot |
| `ConclusionListPage.page` on Page columns; list beside detail | the Conclusion lists' redesign ([`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md)) |
| The Source page on Page columns | S13-21 if S13-D8 says so, else [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md) |
| Pinned header mode, if the S13-D1 board does not pin (Q-1) | its first caller (the Source page, [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md)) |
| The kit CSS (`spacing.css`, the `@container` query, WL-10) | the Claude Design design-system project (S14-D1 board, done) |
| `pageSection` on the engine's `WorkspaceLocation` proto (search context) | not planned — D-2 |
| A segmented-picker initializer on `PVChipGroup` replacing the three hand-written copies | To user (review M-2); not this PR unless the user says so |
| A Tab kit component | not planned — brief non-goal; segmented `PVChipGroup` |
| Arrow-key movement between chips | not planned — D-9 |
| Remembering the last section per person | not planned — brief non-goal (PPF-5) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Page and slot | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | Three generics after S13-09; memberwise (full) init takes `pageSections: (Content, ConclusionDetailPageSectionsContext) -> PageSections`; convenience init (where `PageSections == ConclusionDetailDefaultPageSections<Content>`) supplies the default fill. `static func page(identifier:body:)`: a `ScrollView` around a `VStack(spacing: PVSpacing.space9)` framed to `widthContentMax`, `gutterPage` horizontal padding, `space8` vertical, centered, `surfacePage` background |
| Loaded body | same file `ConclusionDetailLoaded` (private) | On `.content`: `ConclusionDetailHeader`, then `pageSections(content, context)`, both inside `page`'s `VStack` |
| Default fill | same file `ConclusionDetailDefaultPageSections` (internal) | One block: Details `PVSectionHeader` (`L10n.Conclusions.detailsSection`, refreshing `ProgressView` aside) + `ForEach(content.rows)` → `ReconciledValueRow`; likeness section when non-empty; `ForEach(content.sections)` → relationship section. Returns siblings that take the page `VStack`'s spacing |
| Likeness grid | same file `ConclusionDetailLikenessSection` (private) | Heading "Likenesses" with count, adaptive grid of 120pt columns, tile opens `tile.location` via `navigation.go(to:)`, Portrait Badge / Use as portrait from `ConclusionPortraitModel`; uses `@FocusState` + `.pvFocusRing` on tiles |
| Header | same file `ConclusionDetailHeader` (private) | Thumb (with "Use earliest likeness" right-click menu when `offersEarliest`), Badge, title, summary, ref, members |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView` | Convenience init; `summary: { PersonDetailVitals(vitals: $0.vitals) }` (private view, the b. and d. lines); `.vocabularyToastOverlay($portrait.toast, …)`; does not read `WorkspaceNavigation` |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent` | `struct Vital`, `var vitals: [Vital]`, `static func vital(…) -> Vital` = the header's b. and d. lines; `rows`, `likenesses`, `thumbnail`, `badge` |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Conclusions.a11yVital(_:date:place:)`; `macos/App/Resources/Localizable.xcstrings` | Spoken b./d. line; no section keys |
| Width class | `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` `PVWidthClass`, `\.pvWidthClass`, `pvMeasuresWidthClass()` (S14-24) | Regular below `PVSpacing.widthWideBreakpoint` (1600), Wide at or above; measured once at `WorkspaceDestinationHost` (`.frame(maxWidth: .infinity, maxHeight: .infinity).pvMeasuresWidthClass()`); no page reads it yet |
| Tokens | `macos/App/DesignSystem/Tokens/PVSpacing.swift` | `widthInspector` 340, `widthContentMax` 1180, `widthWideBreakpoint` 1600, `gutterPage`, `space8`, `space9`; no `widthSecondaryMax` |
| Kit README | `macos/App/DesignSystem/README.md` component table | Has a WidthClass row; no Page columns row |
| Wide rules | `docs/macos-client-patterns.md` § 9; S14-D1 brief (archived) WL-1 – WL-10 | The breakpoint, six rules, pattern table; Page columns sizing |
| Location | `macos/App/Features/Workspace/WorkspaceLocation.swift` `WorkspaceLocation` | `entityId` names the person; memberwise `init(section:…listScrollAnchor:)` with one parameter per field; tolerant `init(from:)` assigns fields directly; `personDetail(entityId:ref:title:)`; `withListScrollAnchor` copies `self`; `SourceSurface` enum and its `normalizedSurface` live in this file; `==` lists deep ids |
| Sections | `macos/App/Features/Workspace/WorkspaceSection.swift` `WorkspaceSection.glyph` | Sources `.symbol(.library)`, Properties `.symbol(.listTree)` |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` spec `.personDetail` | `viewIdentity: { $0.entityId }` (a section change does not remount); `queryKeys: { project, location in … }` already receives the location |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.personDetail` | `PersonDetailView(session:entityId: navigation.currentLocation.entityId ?? "", userID:)` |
| Navigation | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `go(to:)` | Coalesces when `location == currentLocation`; otherwise pushes and persists |
| Crumbs, jump rows | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems(for:names:goTo:)`, private `HistoryJumpRow.leafLabel` | Two derivations of one label: the crumb leaf is `title ?? ref ?? "…"` with id `"leaf-\(section.rawValue)"` and `action: nil`; the jump row is `title ?? ref` |
| Proto | `api/proto/engine.proto` `WorkspaceLocation`; `macos/App/Platform/GoStore.swift` `mapWorkspaceLocationToProto` / `FromProto` | Sent as search context only (`searchCatalog`); `listScrollAnchor` is not on the wire |
| Kit chip | `macos/App/DesignSystem/Components/Chip/PVChip.swift` `PVChip`, `PVChipGroup(style: .segmented)` | `PVChip` is a `Button` with `isSelected`, but adds no `.isSelected` trait and no focus ring. Segmented copies: `Features/Dates/DateValueEditorForm.swift`, `Features/CitationComposer/CitationComposerAuthoringBox.swift`, `Features/Properties/PropertiesView.swift` |
| Kit empty | `macos/App/DesignSystem/Components/EmptyState/` `PVEmptyState` | `init(glyph: PVGlyph, title: LocalizedStringResource?, message: String?, compact:, action:)`; `message` is a `String` |
| Pinned precedent | `Features/Sources/SourcesListView.swift` `LazyVStack(pinnedViews: [.sectionHeaders])`; `Features/Sources/SourcePageIdentityHeader.swift` (pinned above the Source page's scroll) | Two ways the app pins today |
| Spike 14 | `docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` (Blocked — re-plan) | D-1 replaces `likenesses` with a `gallery` slot carrying its wording; D-3 renames `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection` |
| Project file | `macos/Provenencia.xcodeproj/project.pbxproj` | Hand-managed (S14-15 not landed). S14-24 (`dabe3e7e`) added `PVWidthClass.swift` and `PVWidthClassTests.swift`: a build file, a file reference, a group entry and a Sources phase entry each, ids `5C…F24A1/A2` and `AA…F24B1/B2` |
| Shortcuts | `rg keyboardShortcut macos/App` | ⌘1 – ⌘9 unused |
| Tests | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift`, `PersonDetailTests.swift` (lines 268, 269, 318 read `vitals`), `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PVWidthClassTests.swift`, `HostedWindow.swift` | Push/restore, legacy decode, equality, crumbs; Person content; width class by pane width and in a hosted window |

**Precedent.** `listScrollAnchor` (S10) added an optional, Swift-only, tolerant-decoding field with a copy-with helper and legacy-decode tests: copy its shape for `pageSection`, but the section **is** identity (it is in `==`). `SourceSurface` + `normalizedSurface` is the precedent for a small enum beside `WorkspaceLocation` with one private normalizer every construction path calls. S13-09 ([#624](https://github.com/mendahu/provenencia/pull/624)) is the precedent for the slot. S14-24 ([#628](https://github.com/mendahu/provenencia/pull/628), `dabe3e7e`) is the precedent for a new kit component folder: the file, its test suite, the four pbxproj entries per file, a README row, and table-driven width tests plus a `HostedWindow` test. `ConclusionDetailLikenessSection`'s `@FocusState` + `.pvFocusRing` is the focus-ring precedent for `PVChip`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `enum ConclusionPageSection: String, Codable, CaseIterable, Sendable` *(new)* in `WorkspaceLocation.swift` beside `SourceSurface`, cases `vitals, family, timeline, gallery, sources` in that order. Raw values are frozen: they are written to history files. `static func offered(by section: WorkspaceSection) -> [ConclusionPageSection]` returns all five for `.persons` and `[]` for every other section. It owns `label: LocalizedStringResource` (D-10) and `glyph: PVGlyph`: vitals `.symbol(.list)`, family `.symbol(.gitBranch)`, timeline `.symbol(.calendar)`, gallery `.symbol(.photo)`, sources `.symbol(.library)` (as `WorkspaceSection.sources`) | Kind-neutral (review M-1): Event or Place adopts sections by returning its cases from `offered(by:)` and adding any case of its own to the enum, with no second field, key, `==` line or label branch. Glyphs are existing `PVSymbol`s (review X-3, M-8) | A Person-only enum and field; new marks |
| D-2 | `WorkspaceLocation.pageSection: ConclusionPageSection?` *(new)*, coding key `pageSection`, memberwise `init` parameter `pageSection: ConclusionPageSection? = nil` (after `entityId`). One private `static func normalizedPageSection(_:section:entityId:) -> ConclusionPageSection?` returns nil when `entityId` is nil, when the value is not in `offered(by: section)`, or when it equals `offered(by: section).first` (the default section); otherwise the value. `init`, `init(from:)` (which decodes a `String?` and maps it with `ConclusionPageSection(rawValue:)`, unknown → nil) and `withPageSection(_:)` *(new)* all call it. `withPageSection` copies `self`, sets the field through the normalizer, and keeps `ref`, `title`, `sourceTitle`. `==` compares `pageSection`. `personDetail(entityId:ref:title:)` gains `section: ConclusionPageSection? = nil`. Not on the proto | Sections are places (working rule); one normalizer for every path (review X-1); Vitals is stored as nil, so Vitals-by-argument equals Vitals-by-default and `go(to:)` coalesces; existing links land on Vitals; old and unknown history decodes | Storing `.vitals` explicitly; normalizing only in `init`; adding it to `engine.proto` (`listScrollAnchor` precedent) |
| D-3 | `PersonDetailView` reads `@Environment(WorkspaceNavigation.self)`. The section is `navigation.currentLocation.pageSection ?? .vitals`; choosing one calls `navigation.go(to: navigation.currentLocation.withPageSection(choice))`. No local `@State` for the section. `.vocabularyToastOverlay` and the `portrait` argument stay | `add-workspace-location`: the location is the one source of where we are; `go(to:)` coalesces the same section (AC-2.2); `viewIdentity` stays `entityId`, so `ConclusionPortraitModel` and its toast survive a switch | A `@State` section synced to the location; the host passing the section in |
| D-4 | Split `ConclusionDetailDefaultPageSections` into named internal parts in the same file: `ConclusionDetailRowsSection` *(new)* with `rows: [ReconciledValueRowModel]`, `refreshing: Bool`, `title: LocalizedStringResource?` (nil draws no heading and puts the refreshing spinner where Q-4 says), and `ConclusionDetailLikenessSection` changed from `private` to internal with no other change. The default fill becomes `ConclusionDetailRowsSection(rows: content.rows, refreshing: context.refreshing, title: L10n.Conclusions.detailsSection)`, the likeness `if`, and the relationship `ForEach`, with unchanged output | Map scope "Reuse first" (S13-09 MNT-1); Vitals and Gallery reuse by name; optional title covers every Q-4 answer (review X-7) | Copying the heading, spinner and rows into the Person view; flags on the whole default fill |
| D-5 | `struct ConclusionDetailSectioning` *(new, in `ConclusionDetailPage.swift`)* holds `sections: [ConclusionPageSection]`, `selected: ConclusionPageSection`, `select: (ConclusionPageSection) -> Void`. `ConclusionDetailPage` gains `var sectioning: ConclusionDetailSectioning? = nil` (memberwise; the convenience init leaves it nil). When it is set, the page draws the switcher (D-6) under the header in Page columns' header slot, and passes `sectioning.selected` to `PVPageColumns` as its `scrollResetKey` (D-12). Person passes `sectioning` and its own fill through the full initializer: `PersonDetailSections` *(new, private in `PersonDetailView.swift`)* takes `content`, `context` and `section` and returns sibling views from one `switch section`: `.vitals` → `ConclusionDetailRowsSection(rows:refreshing:title:)` (title per Q-4); `.gallery` → `ConclusionDetailLikenessSection(projectDir:likenesses:portrait:)` when `content.likenesses` is non-empty, else `PVEmptyState(glyph: .symbol(.photo), title: L10n.Conclusions.galleryEmptyTitle, message: L10n.string(L10n.Conclusions.galleryEmptyMessage))`; `.family`, `.timeline`, `.sources` → `PVEmptyState(glyph: section.glyph, title: section.label, message: L10n.string(L10n.Conclusions.pageSectionLater))` | The switcher belongs in the header slot whether the board pins it or not (brief PPF-3, PPF-W1); handing the page data, not a view, keeps one view slot and three generics (review M-3), and Event and Place adopt sections by passing data; working rule "One sectioned Conclusion page" | A fourth generic or second view slot for a header accessory; the switcher as the fill's first sibling (it could then not be pinned or span the column group) |
| D-6 | `PVChip` gains, for every caller: `.accessibilityAddTraits(isSelected ? .isSelected : [])` on its `Button`, and a `@FocusState private var isFocused: Bool` with `.focused($isFocused)` and `.pvFocusRing(isFocused, cornerRadius: chipRadius)` (the likeness tile precedent). The switcher is a private `ConclusionDetailSwitcher` *(new)* in `ConclusionDetailPage.swift`: `PVChipGroup(style: .segmented)` with one `PVChip(section.label, isSelected: section == sectioning.selected, expands: true, selectionLift: true) { sectioning.select(section) }` per `sectioning.sections`, each with `.accessibilityIdentifier("conclusion.detail.section.\(section.rawValue)")`; the group has `.accessibilityElement(children: .contain)` and `.accessibilityLabel(Text(L10n.Conclusions.pageSectionsLabel))`, as `PropertiesView` labels its segmented group | The selected trait and focus ring belong to the kit control (review M-2; `PVRadio`, `PVSidebarNav` own their trait); the switcher is kind-neutral chrome of the shared page | Traits added per call site; a new Tab component; a native `TabView` or `Picker(.segmented)` |
| D-7 | Each empty section is one `PVEmptyState(glyph:title:message:)`; Family, Timeline and Sources share one message key | User, 2026-10-29: placeholders show an empty state; brief PPF-10 | Hiding unbuilt sections (user declined); skeleton rows |
| D-8 | `WorkspaceToolbar` gains `static func locationLabels(for location: WorkspaceLocation) -> (leaf: String, section: String?)` *(new)*: `leaf` is `title ?? ref ?? "…"`, `section` is `L10n.string(location.pageSection.label)` when `pageSection != nil`. `breadcrumbItems` uses it: with a section, the entity crumb gets id `"entity-\(location.section.rawValue)"` and `action: { goTo(location.withPageSection(nil)) }`, and a third crumb with id `"leaf-\(location.section.rawValue)"`, the section label and `action: nil` follows; without a section the crumbs are as today. `HistoryJumpRow.leafLabel` returns `L10n.Conclusions.pageSectionLeaf(leaf, section)` when `section != nil`, else `leaf` | One label rule for both (review M-5, X-5); the Vitals entry reads like today; testable without opening `HistoryJumpRow` (review X-6) | Two derivations edited separately; "Jane Doe · Vitals" |
| D-9 | ⌘1 – ⌘5 are `.keyboardShortcut(KeyEquivalent(Character(String(index + 1))), modifiers: .command)` on each switcher chip, `index` from `sectioning.sections.enumerated()`, so they act only while a sectioned page shows. Keyboard focus moves between chips with Tab (the macOS default for buttons); no arrow-key handling | Shortcuts scoped by the view, not `ProvenenciaApp.commands`; arrow keys would need custom key handling for one picker (review X-4) | App-menu commands; `onKeyPress` arrow handling |
| D-10 | New copy in `L10n.Conclusions`, each with a `comment:` per `add-localized-string`: `pageSectionVitals` "Vitals", `pageSectionFamily` "Family", `pageSectionTimeline` "Timeline", `pageSectionGallery` "Gallery", `pageSectionSources` "Sources" (keys `conclusions.section.<raw>`), `pageSectionsLabel` "Sections" (`conclusions.sections`), `pageSectionLater` "This section arrives in a later version." (`conclusions.section.later`), `galleryEmptyTitle` "No likenesses yet" (`conclusions.gallery.empty.title`), `galleryEmptyMessage` "A likeness is drawn on a Citation's image." (`conclusions.gallery.empty.message`), `pageSectionLeaf(_ title: String, _ section: String) -> String` "%@ · %@" (`conclusions.section.leaf`). The gallery keys are named so S14-02 can move them onto its `gallery` slot unchanged | `add-localized-string`; one namespace with the page's other copy | A new namespace for one page |
| D-11 | New files only for the kit component: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` *(new)* and `macos/ProvenenciaTests/PVPageColumnsTests.swift` *(new)*, each added to `project.pbxproj` with a build file, a file reference, a group entry (a new `PageColumns` group under `Components`) and a Sources phase entry, copying S14-24's `PVWidthClass` entries with fresh ids in the same scheme. Everything else stays in existing files: `ConclusionPageSection` in `WorkspaceLocation.swift`; `ConclusionDetailRowsSection`, `ConclusionDetailSectioning`, `ConclusionDetailSwitcher` in `ConclusionDetailPage.swift`; `PersonDetailSections` in `PersonDetailView.swift`; `locationLabels` in `WorkspaceToolbar.swift` | A kit component has its own folder (S14-D1 inventory, `add-ui-component`); feature types stay beside their callers (S13-09 D-4) | Page columns inside `ConclusionDetailPage.swift` (it is kit, with later callers) |
| D-12 | `PVPageColumns` owns the scroll reset: `scrollResetKey: AnyHashable? = nil`; it wraps its `ScrollView` in a `ScrollViewReader`, gives the column group `.id(PVPageColumns.topAnchor)` (`static let topAnchor = "pv.pageColumns.top"`), and on `.onChange(of: scrollResetKey)` calls `proxy.scrollTo(topAnchor, anchor: .top)`. A width-class change does not reset scroll (rule 5) | The component that owns the scroll owns its reset (review M-3, X-2; S14-D1 WL-3 "it owns the scroll"); kind-neutral | A reset in `ConclusionDetailPage` (it no longer owns the `ScrollView`); `.id(section)` on the fill (does not reset the offset) |
| D-13 | Rename the header's b. and d. lines from "vital" to "life line": `PersonDetailContent.Vital` → `LifeLine`, `vitals` → `lifeLines`, `static func vital(…)` → `lifeLine(…)`, private `PersonDetailVitals` → `PersonDetailLifeLines` (its `vitals` property → `lines`), `L10n.Conclusions.a11yVital` → `a11yLifeLine` (Swift symbol only; the string key does not change), and the three reads in `PersonDetailTests.swift` | One word, one concept, in the file that now has a Vitals section (review M-6) | Two meanings of "vitals" side by side |
| D-14 | `struct PVPageColumns<Header: View, Primary: View, Secondary: View>: View` *(new)*, from S14-D1 WL-3 – WL-9: `init(headerPinned: Bool = false, scrollResetKey: AnyHashable? = nil, @ViewBuilder header:, @ViewBuilder primary:, @ViewBuilder secondary:)`, plus constrained inits with `Header == EmptyView` and/or `Secondary == EmptyView`. It reads `@Environment(\.pvWidthClass)` and does not re-measure the class. Column widths come from one pure function `static func columnWidths(pane: CGFloat, widthClass: PVWidthClass, hasSecondary: Bool) -> (primary: CGFloat, secondary: CGFloat?, group: CGFloat)`: Regular, or Wide without a secondary → primary = min(`widthContentMax`, pane − 2 × `gutterPage`), secondary (if any) the same width, stacked `space9` below, group = primary (WL-4, WL-6); Wide with a secondary → `available = pane − 2 × gutterPage − space9`, primary = min(`widthContentMax`, available − `widthInspector`), secondary = clamp(available − primary, `widthInspector`, `widthSecondaryMax`), group = primary + `space9` + secondary (WL-5). A private `Layout` (`PVPageColumnsLayout`, given the class) places the columns from the proposed width with that function: side by side at Wide, top-aligned, primary leading; the group centered (WL-7). The header spans `group` and aligns with it (WL-8); it sits `space9` above the columns; when `headerPinned` it is outside the `ScrollView`, above it, otherwise the first thing inside it. The component adds `space8` top and bottom padding and no spacing inside a slot (WL-3); it does not set a background. Primary comes before secondary in the view tree at both widths, so focus moves primary then secondary (WL-9). Build `headerPinned: true` only if the board pins (Q-1); otherwise the parameter is not added | S14-D1 is the agreed spec (working rule "Redesigns adopt Page columns"); one pure, tested sizing function; a `Layout` sizes from the proposal without state, so crossing 1600 keeps scroll and focus (rule 5) | Measuring width with `GeometryReader` and state (a second measurement beside S14-24's); a stretched primary; a full-width mode (Out) |
| D-15 | `PVSpacing.widthSecondaryMax: CGFloat = 480` *(new)* beside `widthInspector`, with the S14-D1 comment (the secondary column of Page columns grows from `widthInspector` to this). `widthWideBreakpoint` stays 1600: the S14-D1 board's WL-1 says `1600px`, so they agree; a test pins both values | S14-D1 handoff: "That PR checks `--width-wide-breakpoint` against the 1600 S14-24 shipped" | Changing either value here |
| D-16 | `ConclusionDetailPage.page` builds `PVPageColumns` instead of its own `ScrollView`/`VStack`: on `.content` the header slot holds `ConclusionDetailHeader` (and the switcher when `sectioning` is set) and the primary holds `pageSections(content, context)` inside a `VStack(alignment: .leading, spacing: PVSpacing.space9)`, so the fill's siblings keep today's spacing; `.loading` and `.failed` go in the primary with no header (WL-3); no secondary for any kind in this PR unless Q-6 gives Vitals one; `.background(PVColor.surfacePage)` and the page's `accessibilityIdentifier` stay on the page | S14-D1 implementation gate: "`ConclusionDetailPage.page` moves onto it, so the Person, Event and Place pages share it"; at Regular, primary-only Page columns has today's width, gutters and padding, so Event and Place look the same (AC-3.6) | Keeping `page`'s own scroll for Event and Place (two scaffolds) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVPageColumns`, `PVPageColumnsLayout`, `columnWidths` *(new)* | component | `ConclusionDetailPage.page`; `SourcePageIdentityHeader`'s copy of the column frame | Agreed on S14-D1 as the one scaffold; replaces `page`'s own frame, and is the Source page's next home |
| `PVSpacing.widthSecondaryMax` *(new)* | token | `widthInspector` | The secondary column's upper bound (S14-D1 WL-1) |
| `ConclusionPageSection` *(new)* | type | `SourceSurface` (same file) | A different axis; extends the same pattern |
| `WorkspaceLocation.pageSection`, `withPageSection(_:)`, `normalizedPageSection` *(new)* | field, funcs | `listScrollAnchor`, `withListScrollAnchor(_:)`, `normalizedSurface` | Same shapes; the section is identity |
| `ConclusionDetailSectioning`, `ConclusionDetailSwitcher` *(new)* | type, view | `ConclusionDetailPageSectionsContext`, the segmented groups in `PropertiesView` | Data the page draws from; kind-neutral |
| `ConclusionDetailRowsSection` *(new)* | view | the Details block inside `ConclusionDetailDefaultPageSections` | Extracted, not added beside it (D-4) |
| `PersonDetailSections` *(new, private)* | view | `ConclusionDetailDefaultPageSections` | Person's fill differs by section; reuses the default's parts |
| `WorkspaceToolbar.locationLabels(for:)` *(new)* | func | `breadcrumbItems` leaf, `HistoryJumpRow.leafLabel` | Replaces two derivations with one |
| `PVChip` trait and focus ring | kit extend | `PVRadio`, `PVSidebarNav` (trait); likeness tile (focus ring) | Extends the one chip; no new control |
| L10n keys (D-10) *(new)* | key | `L10n.Conclusions.detailsSection`, `likenessesSection` | New names; same namespace |
| `PVPageColumnsTests` *(new)* and other tests (§8) | test | `PVWidthClassTests`, `HostedWindow`, `legacyHistoryMissingListScrollAnchorDecodesAsNil`, `conclusionDetailCrumbPrefersTheTitle` | Same patterns |

**Builds on.** S14-24's width class and the S14-D1 board; S13-09's `pageSections` slot and full initializer; `add-workspace-location` (one location, `go(to:)`, tolerant decode, one normalizer); the kit's segmented chip group and `PVEmptyState`.
**Deletes.** `ConclusionDetailPage.page`'s own `ScrollView`, `VStack` frame, gutters and padding (now `PVPageColumns`); the single-block body of `ConclusionDetailDefaultPageSections`; `private` on `ConclusionDetailLikenessSection`; the second leaf derivation in `HistoryJumpRow.leafLabel`; the names `Vital`, `vitals`, `vital(…)`, `PersonDetailVitals`, `a11yVital` (renamed, D-13); `PersonDetailView`'s use of the convenience initializer. No callee loses its last caller.
**Next adds.** S13-11 adds the provenance control inside the Vitals rows (Person only, by caller). S13-12, S13-13 and S13-16 each replace one empty-state case in `PersonDetailSections`; each section's query key is warmed from the `.personDetail` spec's `queryKeys` in `PlaceRegistry`, which already receives the location (working rule "Each section reads its own query"), and the section view reads a `QueryHandle` rather than starting its own load; a section that wants supporting content at Wide fills a secondary slot that `pageSections` then gains, drawn on its own board. S13-14 replaces the Gallery case. S13-28 adds a case or a view inside Sources. Event and Place adopt sections by returning cases from `ConclusionPageSection.offered(by:)` and passing `sectioning`. The Conclusion lists and the Source page move onto `PVPageColumns` in their redesigns (`docs/ideas/wide-layout.md`). S14-02 renames `ConclusionDetailLikenessSection` and moves the two gallery keys onto its slot.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A location with a section for a person that was deleted | `ConclusionDetailPage` failed state | Same as today's Vitals: the danger Callout in the primary; no header, no switcher |
| Restoring a non-Vitals section from history on launch | `PersonDetailView` | Section read from `currentLocation` on every render; no apply step (D-3) |
| Switching person (a family tile later) while on Gallery | `PlaceRegistry` remounts on `entityId` | A new person's location has no section → Vitals (D-2) |
| Gallery with no likenesses | `PersonDetailSections` | Empty state (AC-1.3) |
| Every `PVChip` caller now carries the selected trait and a focus ring | the ten chip call sites (date kind, credibility, Properties, composer) | VoiceOver gains "selected"; the ring shows only with keyboard focus |
| The Event and Place pages now sit on `PVPageColumns`, and every page can be Wide | `EventDetailView`, `PlaceDetailView` | Regular: today's frame (AC-3.6); Wide: centered (WL-6); loading and error in the primary |

### Scale

None — no write and no new read: switching re-renders from the one cached `conclusionDetail`. A person with 20 rows and 30 likenesses renders the same views as today, one section at a time. `PVPageColumnsLayout` is constant work per layout pass.

### Edges

- **Assumptions.** `go(to:)` coalesces equal locations, so `==` includes `pageSection` (T-3) and every construction path normalizes the default section to nil (T-2). `columnWidths` assumes `pane ≥ widthWindowMin − widthSidebar`; below `2 × gutterPage` it returns 0, never negative (T-12).
- **Neighbors.** Event, Place and list-root locations never carry a section (T-4). Panes 1599 and 1600 give Regular and Wide (S14-24 tests); at 1600 with a secondary, primary 1164 and secondary 340; at 1616 primary 1180 and secondary 340; at 1756 secondary 480 (T-12).
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

From the S13-D1 inventory and the S14-D1 board. Frame descriptions are added when the S13-D1 board is agreed (Q-1 – Q-6).

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Scaffold | `PVPageColumns` *(new)* | `DesignSystem/Components/PageColumns/PVPageColumns.swift` | D-14; kit README row "PageColumns" linking § 9 |
| Page | `ConclusionDetailPage` on `PVPageColumns`; `sectioning` | `ConclusionDetailPage.swift` | D-5, D-16; Event and Place keep the convenience init |
| Header | `ConclusionDetailHeader` in the header slot | same | Unchanged unless Q-2 adds a compact form |
| Switcher | `ConclusionDetailSwitcher`: `PVChipGroup(.segmented)` + `PVChip` (extended) | same; `PVChip.swift` | D-6, D-9 |
| Vitals | `ConclusionDetailRowsSection` | `ConclusionDetailPage.swift` | D-4, Q-4; secondary per Q-6 |
| Gallery | `ConclusionDetailLikenessSection` / `PVEmptyState` | same / kit | D-5; Wide per Q-6 |
| Family, Timeline, Sources | `PVEmptyState` | kit | D-7; primary-only at Wide |
| Crumbs, jump rows | `PVBreadcrumbs`, `HistoryJumpRow` via `locationLabels` | `WorkspaceToolbar.swift` | D-8 |

States: loading and failed in the primary, no header (WL-3); refreshing spinner per Q-4. Widths: Regular 1366 × 900 and Wide 2560 × 1440, sidebar open. Accessibility: chips are buttons with names and the selected trait; empty states read title then message; focus primary then secondary.

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

**STEP-3 — Page columns in the kit**
Files: `macos/App/DesignSystem/Tokens/PVSpacing.swift`, `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` *(new)*, `macos/ProvenenciaTests/PVPageColumnsTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`, `add-swift-test`.
Do: D-15, D-14 (pinned mode only if Q-1 pins), D-12, D-11's pbxproj entries. README: a "PageColumns" row after "WidthClass" naming the slots, the sizing owner and a link to § 9 of `docs/macos-client-patterns.md`, not a copy of the rules.
Check: T-12 – T-14 green.

**STEP-4 — The Conclusion page sits on Page columns**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`.
Do: D-16.
Check: `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` green; §10 step 9 (Event and Place unchanged at 1366, centered at 2560).

**STEP-5 — Split the default fill into named parts**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`. Skill: `add-ui-component`.
Do: D-4.
Check: the detail suites green; Event and Place unchanged.

**STEP-6 — Chips carry their selected state and focus**
Files: `macos/App/DesignSystem/Components/Chip/PVChip.swift`. Skill: `add-ui-component`.
Do: D-6, the `PVChip` half.
Check: the app builds; with VoiceOver, the date-kind picker reads its selected chip as selected.

**STEP-7 — The page draws a kind's sections**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: D-5 (`ConclusionDetailSectioning`, `sectioning`, the reset key), D-6 (`ConclusionDetailSwitcher`), D-9, the `pageSectionsLabel` key.
Check: Event and Place unchanged (no `sectioning`).

**STEP-8 — Person sections** (after the S13-D1 board; Q-1 – Q-6 answered in this plan)
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-3, D-5 (`PersonDetailSections`, the full init with `sectioning`), D-7, D-10's remaining keys; the board's header mode and Wide arrangement (Q-1, Q-2, Q-6).
Check: T-10; §10 steps 1 – 5, 7 and 8.

**STEP-9 — Crumbs and jump rows name the section**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: D-8. Write T-6, T-7.
Check: `WorkspaceToolbarBreadcrumbTests` green.

**STEP-10 — Docs**
Files: `docs/macos-client-patterns.md` (one sentence under the sectioned detail page bullet: a kind passes its own fill and its `sectioning`, its section lives on `WorkspaceLocation.pageSection`, and the page sits on `PVPageColumns`; one sentence in § 9 naming `PVPageColumns` as the record-page scaffold), `docs/ideas/wide-layout.md` (the Person, Event and Place row: Person shipped on Page columns, Event and Place on the shared scaffold), `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S13-D1-person-page-frame.md` → `../design/archive/` *(new)*, `../design/README.md`.
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
| T-8 | Swift | `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` existing (renamed reads per D-13) | unchanged | AC-4.1, D-13 |
| T-9 | Manual | §10 | sections, empty states, VoiceOver, keyboard, scroll to top, Regular and Wide | AC-1.1 – AC-1.6, AC-2.6, AC-2.9, AC-3.1 – AC-3.6 |
| T-10 | Structural | STEP-8 `rg` | In `macos/App/Features/Conclusions/PersonDetailView.swift`: `rg -F 'ForEach(content.rows)'` finds nothing; `rg -F 'ConclusionDetailRowsSection'` and `rg -F 'ConclusionDetailLikenessSection'` each find a hit | AC-4.2, D-5 |
| T-11 | Swift | `WorkspaceNavigationTests` `sectionsOfferedPerKind` *(new)* | `offered(by: .persons)` → the five in order; `.events`, `.places`, `.sources` → `[]` | AC-4.3, D-1 |
| T-12 | Swift | `PVPageColumnsTests` `columnWidthsFollowTheBoard` *(new, table-driven)* | (pane, class, secondary) → (primary, secondary, group): (1102, R, yes) → (1038, 1038, 1038); (1102, R, no) → (1038, nil, 1038); (1600, W, yes) → (1164, 340, 1536); (1616, W, yes) → (1180, 340, 1552); (1656, W, yes) → (1180, 380, 1592); (1756, W, yes) → (1180, 480, 1692); (2296, W, yes) → (1180, 480, 1692); (2296, W, no) → (1180, nil, 1180); (40, R, no) → (0, nil, 0) | AC-4.4, D-14 |
| T-13 | Swift | `PVPageColumnsTests` `tokensMatchTheBoard` *(new)* | `widthWideBreakpoint == 1600`, `widthSecondaryMax == 480`, `widthInspector == 340`, `widthContentMax == 1180` | D-15 |
| T-14 | Swift | `PVPageColumnsTests` `wideHostPlacesTheSecondaryBeside` *(new, `HostedWindow`)* | Host a `PVPageColumns` with probe views at a 2296 and a 1102 pane width with `\.pvWidthClass` set → secondary's frame is beside the primary (Wide) and below it (Regular) | AC-3.1, AC-3.4, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceNavigationTests` (`==` change), the detail suites (STEP-2, STEP-4, STEP-5), `PVWidthClassTests` and `WorkspaceDestinationHostTests` (STEP-3, STEP-4), any test that compares chip accessibility traits (STEP-6).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`PersonDetailContent.Vital`, `PersonDetailVitals`, `a11yVital`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the S13-D1 inventory and the S14-D1 board; no new kit primitive beyond `PVPageColumns`
- [ ] Each view checked at Regular (1366 × 900) and Wide (2560 × 1440), sidebar open (§10)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md (a kind's own fill and sections; pageSection; PVPageColumns in § 9); ideas/wide-layout.md Person row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

At 1366 × 900, sidebar open, unless a step says Wide (2560 × 1440, sidebar open).

1. Open a project with a Person that has a name, both life lines, three likenesses and a Mixed row. Open them from the Persons list: header as before, Vitals selected, rows as before. (AC-1.1)
2. Choose Gallery: today's likeness section with its heading, Portrait Badge, Use as portrait; choose one, see the toast; the header thumb follows. (AC-1.2)
3. Open a Person with no likeness, choose Gallery: the empty state. Choose Family, Timeline, Sources: each empty state with its glyph. (AC-1.3, AC-1.4)
4. Back three times from Gallery after Vitals → Family → Gallery: Family, Vitals, Persons list. Choose Gallery twice: one Back step. (AC-2.1, AC-2.2)
5. On Gallery, click a tile, then Back: Gallery. On Timeline, quit and reopen: Timeline. Scroll Vitals to the bottom, choose Gallery: it shows from the top. (AC-2.6, AC-2.3, AC-2.9)
6. Breadcrumbs on Gallery; click the person crumb → Vitals; open the Back jump menu: "Jane Doe · Family" and "Jane Doe". (AC-2.7, AC-2.8)
7. VoiceOver on the switcher; Tab to it (focus ring), ⌘3 → Timeline. Open the date editor: its kind chips read "selected". (AC-1.5, AC-1.6)
8. Wide: the same Person on Vitals, Gallery and Family match the board's Wide frames (AC-3.1, AC-3.2, AC-3.3). Scroll Vitals halfway, focus the switcher, then drag the window narrower across the line: still Vitals, same scroll, still focused (AC-3.5).
9. Open an Event and a Place at 1366: the same page as before. At Wide: the same column, centered. (AC-3.6)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Adding `pageSection` to `==` changes coalescing for Person entries | T-2 normalizes the first section; T-1 covers restore |
| Moving `page` onto `PVPageColumns` shifts Event and Place by a few points at Regular | D-16: same max width, gutters, padding and inter-section spacing; §10 step 9; stop if they differ |
| Ten steps make this a large PR | Page columns was folded in by the user (map, S14-25 descoped); each step builds and checks on its own |
| A ⌘ shortcut on a chip fires while a sheet is open on the page | Manual check; stop if a shortcut changes section behind a modal |
| S14-02 lands first and renames `ConclusionDetailLikenessSection` and `likenesses` | Refresh §4 and D-4, D-5 to its names before STEP-5 |
| The `PVChip` focus ring shows in a picker where it looks wrong | The ring follows keyboard focus only; stop and report the call site rather than adding an opt-out |

**Stop and report if:**

- `ConclusionDetailPage` no longer has the full initializer or the `pageSections` slot §4 describes, or `PVWidthClass` no longer reads as §4 says.
- The agreed S13-D1 board needs a Person-only page type, a second view slot, a third column, a Page columns full-width mode, or a change to Event or Place layout.
- The S14-D1 board's breakpoint or token values differ from D-15.
- `PVChip` cannot carry `.focused`, `.keyboardShortcut` or the selected trait without changing its public initializers.
- A `Layout` cannot place the header and columns per D-14 without measuring width in state.
- The board's frames and an AC disagree.

## 12. Open questions

Answered by the S13-D1 board; the plan is refreshed and reviewed again then.

- **Q-1 (PPF-2, PPF-3).** Pattern A (one section at a time) or B (stacked with shortcuts), and is the header slot pinned (user, 2026-10-29: the board decides)? Recommended: A. Pinning uses Page columns' pinned header mode (S14-D1 WL-3, WL-8), built in STEP-3 only if the board pins.
- **Q-2 (PPF-3).** If pinned, does the header compact on scroll, and to what?
- **Q-3 (PPF-6).** Crumb and jump-row wording (structure fixed by D-8).
- **Q-4 (PPF-8).** Vitals heading: "Details", "Vitals" or none (D-4's `title`), and where the refreshing spinner sits.
- **Q-5 (PPF-13).** ⌘1 – ⌘5, or other keys.
- **Q-6 (PPF-W2, PPF-W3).** What Vitals' secondary column holds at Wide, if anything, and whether Gallery stays primary-only. Recommended: both primary-only and centered in S13-10, leaving the secondary column for S13-D2 and S13-D6. If the board fills it, `pageSections` gains a secondary fill in STEP-8 the same way it supplies the primary, and the plan is re-reviewed.

## 13. PR

**Title:** S13-10 — Header and sections
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check at Regular and Wide; `Plan: docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md`.
