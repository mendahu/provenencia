# S13-10 — Header and sections

**Status:** Landed in #634 (2026-10-29)
**Spike / slice:** Spike 13 · Slice 1 — Page frame and Vitals
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D1](../design/archive/S13-D1-person-page-frame.md), board agreed 2026-10-29; [S14-D1](../../spike-14/design/archive/S14-D1-wide-layout.md) (Page columns), board agreed 2026-10-29
**Board:** S13-D1 — [Claude Design, *Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, turn "S13-D1 · Person page on Page columns — Regular and Wide", frames 4a – 4y (F1 – F19 of the brief). §6 **macOS** describes each frame in words; the board's leftover pattern-B and docked-Why branches are set aside and are not built. S14-D1 — [Claude Design design system](https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445), the Page columns specimen and the kit's segmented `ChipGroup`; WL-3 – WL-9 are copied into D-14
**Depends on:** S13-09 (landed [#624](https://github.com/mendahu/provenencia/pull/624)), S14-24 (landed [#628](https://github.com/mendahu/provenencia/pull/628)), S14-D1 (agreed), S13-D1 (agreed). Shares code with Spike 14's S14-02 (Blocked — re-plan); either may land first, see §4 and §11
**Blocks:** S13-11, S13-12, S13-14, S13-16
**Branch:** `s13-10-header-and-sections`
**Version bump:** none — Spike 11 – 14 PRs leave `VERSION` alone until a release is cut, as S13-09 and S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-1 – R-7, R-14, R-15; working rules: "Sections are places", "One sectioned Conclusion page, not a Person fork", "S13-D1 draws the whole frame", "Redesigns adopt Page columns", "Each section reads its own query", "The provenance sheet replaces the Why on the Person page only"; [`../deployment-plan.md`](../deployment-plan.md) scope S13-10 ("Reuse first", "Page columns", "One segmented control"); [`macos-client-patterns.md`](../../../macos-client-patterns.md) Conclusion keys and [§ 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); Spike 14 working rule "(Batch 6) One breakpoint, 1600pt of pane width"
**Planned:** 2026-10-29 against `main` @ `cb828f82`
**Review:** [`S13-10-header-and-sections.review.md`](S13-10-header-and-sections.review.md) — round 1 Fits with changes / Not ready, round 2 Fits with changes / Ready with changes; every finding Fixed; the user answered the one To user item (consolidate the segmented picker here)

**Landed:** the code differs from the plan in six places. (1) The pure sizing (`columnWidths`), `isScrolled`, `compactMinOverflow` and `topAnchor` are static members of a non-generic `PVPageColumnsMetrics`, not of the generic `PVPageColumns`, so callers and tests do not name its three view types; AC-5.4 and T-12, T-15 read `PVPageColumnsMetrics.columnWidths` and `.isScrolled`. (2) `pageSectionLeaf` uses the numbered format `%1$@ · %2$@`, which `scripts/check-localizable-xcstrings.py` requires for two arguments. (3) `PersonDetailLifeLines` switches an `AnyLayout` between an HStack and a VStack instead of two branches, so the lines keep their identity when the header compacts. (4) A loading or failed page draws a header-less `PVPageColumns` (the primary-only initializer) instead of a header builder that returns nothing, so no gap sits above the spinner. `WorkspaceLocation.pageSection` is `private(set)`; it changes only through `withPageSection(_:)`. After review (round 1, user 2026-10-29): (5) the page section is called `pageSection` everywhere it is a page section, which supersedes D-2 and D-5: `WorkspaceLocation.personDetail(entityId:ref:title:pageSection:)`, `ConclusionDetailPageSectionsContext.pageSection`, and `pageSection` and `offeredPageSections` in `ConclusionDetailLoaded` and `ConclusionDetailHeaderBand`. (6) `PVPageColumnsMetrics.isScrolled` enters compact only when the overflow is at least `max(compactMinOverflow, bandShrink + compactOffset)` and leaves compact when `overflow <= 0`, which supersedes D-17's literal "false when overflow < 120": compacting hands the viewport back the height the header loses, about 116pt for a one-line header and more when the name wraps, so a fixed 120 guard flipped the header back. `PVPageColumns` measures `bandShrink` from both forms of the header, drawn hidden at the band's width.

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

The Person page keeps its header and gains a section switcher under it: Vitals, Family, Timeline, Gallery, Sources. Vitals holds today's rows; Gallery holds today's likeness grid; Family, Timeline and Sources say they arrive later. Each section is a place: switching pushes history, Back returns to the previous section, and a restored session reopens on it. The header and switcher stay pinned and the header compacts as the section scrolls. The page is laid out on the kit's new Page columns (built here from the S14-D1 board), so on a wide monitor it is centered, and the Event and Place pages share the same scaffold. The switcher is the kit's segmented control, which this PR also gives the three pickers that hand-build it today.

**Why now.** Slice 1's chrome PR; every later Person section (S13-11, S13-12, S13-14, S13-16) fills this frame. Working rules "Sections are places", "S13-D1 draws the whole frame" and "Redesigns adopt Page columns" (S14-25 was folded into this PR, user 2026-10-29). The segmented control: user, 2026-10-29 ("fix it on this plan … consolidating this and moving all the callers over to it").

**What stays.** The header's parts (thumb, gender Badge, name, b. and d. lines, ref, member count) and the thumb's right-click "Use earliest likeness" (user, 2026-10-29: stays until the S13-D5 Gallery board); every row with its Badge, Sources and Why; the likeness section as it is drawn today, its own "Likenesses" heading and count included, now inside Gallery (user, 2026-10-29); one `conclusionDetail` read; the Event and Place pages' look (R-1); the date-kind, Observation-kind and Property value-type pickers' look and behavior; every existing link into a Person page, which lands on Vitals.

## 2. User stories

**US-1.** As a researcher, when I open a Person, I see the header and a switcher with five sections, Vitals showing, so that I can tell where I am and where else I can go.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Open a Person from the Persons list | The header as today; the switcher lists Vitals, Family, Timeline, Gallery, Sources in that order, each with its glyph; Vitals is selected and shows today's rows in today's order, with no heading |
| AC-1.2 | Same person, choose Gallery | Today's likeness section (heading "Likenesses" with the count, tiles, Portrait Badge on the portrait tile, Use as portrait on the others); the header does not change |
| AC-1.3 | A person with no likeness, choose Gallery | `PVEmptyState` titled "No likenesses yet" with "A likeness is drawn on a Citation’s image — draw one there and it appears here." |
| AC-1.4 | Choose Family, Timeline or Sources | `PVEmptyState` with that section's glyph, its name, and "Family arrives in a later version of Provenencia." (the section's name in the line); no rows, no skeleton |
| AC-1.5 | VoiceOver on the switcher with Vitals showing | Reads the group as "Sections", "Vitals, selected", and each other section by name |
| AC-1.6 | With Full Keyboard Access on (System Settings › Keyboard › Keyboard navigation), Tab to the switcher; then press ⌘3 with it on or off | The focused chip shows the kit focus ring; ⌘3 shows Timeline |
| AC-1.7 | On a person with 20 Vitals rows, scroll down past 8pt, then back to the top | The header compacts (40pt thumb, name at h3, b. and d. on one line, ref and members at the right, no Badge) and the switcher stays; at the top the full header returns |
| AC-1.8 | A stale Person page refreshes | A small spinner and "Refreshing" sit at the end of the switcher row, in any section and at any scroll |
| AC-1.9 | Any Vitals row | An empty 28pt slot after the row's disclosures, reserved for S13-11's provenance control; VoiceOver skips it; Event and Place rows have no slot |
| AC-1.10 | A person whose Vitals overflow the window by less than 120pt; scroll to the bottom | The header stays full; it never flickers between full and compact |

**US-2.** As a researcher, when I switch sections, Back and Forward walk them and a reopened project returns to the section I was on, so that a section behaves like any other place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Persons list → Jane Doe (Vitals) → Family → Gallery, then Back | Family; Back again → Vitals; Back again → Persons list |
| AC-2.2 | On Gallery, choose Gallery again | No history entry is added |
| AC-2.3 | On Timeline, quit and reopen the project | Jane Doe on Timeline |
| AC-2.4 | A history file written before this PR (no `pageSection` key) | Decodes; Person entries open on Vitals |
| AC-2.5 | A history entry with `"pageSection": "lineage"` (unknown) | Decodes; opens on Vitals |
| AC-2.6 | From Gallery, open a likeness tile's Citation, then Back | Jane Doe on Gallery |
| AC-2.7 | Breadcrumbs on Jane Doe's Gallery | Persons › Jane Doe › Gallery, Jane Doe going to Vitals; on Vitals, Persons › Jane Doe (board 4q) |
| AC-2.8 | Back jump menu after Vitals → Family | The Vitals entry reads "Jane Doe", the Family entry "Jane Doe · Family" (board 4q); Source, Metadata and Properties entries read as today |
| AC-2.9 | Scroll Vitals to its last row (header compact), choose Gallery | Gallery shows from its first line with the full header |

**US-3.** As a researcher on a wide monitor, when I open a Person, Event or Place, the page is centered and never sits jammed to one side, so that a big window is worth having (brief PPF-W1 – PPF-W5; S14-D1 WL-4 – WL-8).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A 2560 × 1440 window, sidebar open (pane 2296), a Person on Vitals | The header and switcher span the 1180pt column, centered; Vitals primary-only (no secondary column); the pinned bar's bottom rule spans the pane (board 4b) |
| AC-3.2 | Same window, Gallery | Primary-only; the grid fits more 96pt tiles per row in the 1180pt column (board 4j) |
| AC-3.3 | Same window, Family | The empty state centered in a primary-only column (board 4o) |
| AC-3.4 | A 1366 × 900 window (pane 1102), any Person section | One centered column up to 1180pt inside 32pt gutters |
| AC-3.5 | On Vitals scrolled halfway with the switcher focused, resize the window across 1600pt of pane | Still Vitals, the scroll position kept, the switcher still focused, the header still compact (PPF-W5, rule 5) |
| AC-3.6 | Open an Event and a Place at 1366 × 900 and at 2560 × 1440 | At 1366, the same page as before this PR, its header scrolling with the page; at 2560, the same column centered (WL-6) |

**US-4.** As a researcher, when I pick a date's kind, an Observation's kind or a new Property's value type, the picker looks and works as before and now reads its selection to VoiceOver and shows keyboard focus.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The date editor, then the Citation composer's Property / Connection picker, then New Property's value type | Each looks as before (28pt chips, caption type, same order, same identifiers); the composer's locked kind still disables the other chip |
| AC-4.2 | VoiceOver on each | The group reads its label ("Kind", "Observation kind", "Value type") and the selected chip as selected |

**US-5.** As a contributor, when I build the next Person section, move another page onto Page columns or add another segmented picker, there is one place to change for each.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `ConclusionDetailDefaultPageSections` | Built from `ConclusionDetailRowsSection` *(new)*, `ConclusionDetailLikenessSection` and the relationship `ForEach`; Event and Place look the same |
| AC-5.2 | Person's below-header fill | One `switch` over `context.section`; Vitals is `ConclusionDetailRowsSection`, Gallery is `ConclusionDetailLikenessSection`; no row loop copied into `PersonDetailView.swift`; `PersonDetailView` does not read `WorkspaceNavigation` |
| AC-5.3 | `ConclusionPageSection.offered(by:)` | `.persons` → the five cases in order; every other `WorkspaceSection` → `[]` |
| AC-5.4 | `PVPageColumns.columnWidths(pane:widthClass:hasSecondary:)` | Pane 1102, Regular, secondary → primary 1038, secondary stacked at 1038. Pane 1600, Wide, secondary → 1164 / 340. Pane 1656 → 1180 / 380. Pane 2296 → 1180 / 480, group 1692. Pane 2296, no secondary → primary 1180 (WL-4 – WL-7) |
| AC-5.5 | `rg -F 'style: .segmented' macos/App/Features` | No hit: every segmented picker uses `PVSegmentedControl` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionPageSection` *(new)* and `WorkspaceLocation.pageSection` *(new)*, kind-neutral: Codable, nil is the kind's first section, unknown or not offered decodes to nil, part of `==`, one normalizer (PPF-5) | STEP-1, AC-2.1 – AC-2.5, AC-5.3, D-1, D-2 |
| Rename the header's "vitals" (the b. and d. lines) to life lines, so "Vitals" names only the section | STEP-2, D-13 |
| `PVChip` sizes from `PVControlSize`, takes a glyph, draws focus through a `ButtonStyle` and carries the selected trait | STEP-3, AC-1.5, AC-1.6, AC-4.2, D-6 |
| `PVSegmentedControl` *(new)* in the kit; the date-kind, Observation-kind and value-type pickers move onto it (user, 2026-10-29) | STEP-3, AC-4.1, AC-4.2, AC-5.5, D-20 |
| `PVSpacing.widthSecondaryMax` = 480 (S14-D1 WL-1); `widthWideBreakpoint` checked against the board's 1600 | STEP-4, D-15 |
| `pvScrollsRowIntoView` gains a restart-at-top trigger | STEP-4, AC-2.9, D-12 |
| `PVPageColumns` *(new)* in the Swift kit: header slot (scrolling or pinned, `isScrolled` for compacting), primary, optional secondary; owns the scroll and its reset; sizing per WL-4 – WL-7; focus order per WL-9; kit README row (S14-D1 WL-3 – WL-9) | STEP-4, AC-5.4, AC-3.4, AC-3.5, AC-1.10, D-12, D-14, D-17 |
| Split the default fill into `ConclusionDetailRowsSection` *(new)* and an internal `ConclusionDetailLikenessSection`; `ReconciledValueRow.Provenance` *(new)* with the reserved slot (S13-10 scope "Reuse first"; PPF-8) | STEP-5, AC-5.1, AC-1.9, D-4, D-19 |
| `ConclusionDetailPage` moves onto `PVPageColumns`, draws the switcher, pins and compacts the header for a kind with sections, shows refreshing in the switcher row, and passes the section to the fill (PPF-1 – PPF-4, PPF-8, PPF-11) | STEP-6, AC-1.5 – AC-1.8, AC-2.9, AC-3.6, D-5, D-16, D-17, D-18, D-21 |
| Person's sections: Vitals rows, Gallery section or empty, three empty sections (PPF-8 – PPF-10); Wide primary-only (PPF-W1 – PPF-W5) | STEP-7, AC-1.1 – AC-1.4, AC-2.1, AC-2.2, AC-2.6, AC-3.1 – AC-3.3, AC-5.2, D-3, D-7 |
| Crumbs and jump rows name the section; one leaf rule (PPF-6) | STEP-8, AC-2.7, AC-2.8, D-8 |
| L10n for five section names, the switcher label, the empty states, "Refreshing" and the jump-row format | STEP-1, STEP-6, STEP-7, STEP-8, D-10 |
| Docs: patterns bullets, `docs/ideas/wide-layout.md`, `add-design-brief` (Page columns and the segmented control), `completed.md`, map status, brief archive | STEP-9 |

| Out | Home |
| --- | --- |
| Provenance control and sheet on Vitals rows; the Why leaving the Person page | S13-11 (brief S13-D2) |
| Family, Timeline, Sources content and their queries | S13-12, S13-13, S13-16 (queries: S13-01, S13-04, S13-05) |
| Gallery redesign, viewer, where "Use earliest likeness" lives; a full-width grid mode | S13-14, S13-15 (brief S13-D5) |
| The gallery view-layer rename (`ConclusionDetailGallerySection`, the `gallery` slot and its wording) | S14-02 |
| How the records connect | S13-28 |
| Sections and Wide arrangements on Event and Place pages | not this spike (R-1); D-1 and D-5 leave them a slot |
| A secondary column for any Person section | S13-D2 / S13-11, S13-D6 / S13-16 (board PPF-W2: the column stays free for them) |
| `ConclusionListPage.page` on Page columns; list beside detail | the Conclusion lists' redesign ([`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md)) |
| The Source page on Page columns | S13-21 if S13-D8 says so, else [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md) |
| A narrow-window frame or fallback for the switcher (PPF-12) | not planned — the board set it aside at the user's request; D-6's chips drop their glyph before a label truncates |
| The kit CSS (`spacing.css`, the `@container` query, WL-10) and a size or glyph prop on the web `PVChip` | the Claude Design design-system project |
| `PVTabs` (the web kit's underline tabs) in the Swift kit | not planned — the board uses the segmented `ChipGroup` |
| `pageSection` on the engine's `WorkspaceLocation` proto (search context) | not planned — D-2 |
| The loose (non-segmented) `PVChipGroup` pickers (credibility, date qualifier) | not planned — they are a different control; they gain only D-6's trait and focus ring |
| Arrow-key movement between chips | not planned — D-9 |
| Remembering the last section per person | not planned — brief non-goal (PPF-5) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Page and slot | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | Three generics after S13-09; memberwise (full) init: `session`, `entityId`, `pageIdentifier`, `errorIdentifier`, `make`, `summary`, `portrait` (default nil), `pageSections`; convenience init (where `PageSections == ConclusionDetailDefaultPageSections<Content>`) supplies the default fill. `body`: a `QueryHandle` → `ConclusionDetailLoaded`, else `Self.page(identifier:) { Self.loading }`. `static func page(identifier:body:)`: a `ScrollView` around a `VStack(spacing: PVSpacing.space9)` framed to `widthContentMax`, `gutterPage` horizontal padding, `space8` vertical, centered, `surfacePage` background |
| Loaded body | same file `ConclusionDetailLoaded` (private) | Fields `handle`, `projectDir`, `pageIdentifier`, `errorIdentifier`, `make`, `summary`, `portrait`, `pageSections`; `body` calls `page { switch presentation … }`: `.loading` → spinner, `.failed` → `PVCallout`, `.content` → `ConclusionDetailHeader` then `pageSections(content, context)` |
| Context | same file `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing`, `portrait`; documented as "What the `pageSections` slot receives beside the content" |
| Default fill | same file `ConclusionDetailDefaultPageSections` (internal) | One block: Details `PVSectionHeader` (`L10n.Conclusions.detailsSection`, refreshing `ProgressView` aside) + `ForEach(content.rows)` → `ReconciledValueRow`; likeness section when non-empty; `ForEach(content.sections)` → relationship section. Returns siblings that take the page `VStack`'s spacing |
| Likeness grid | same file `ConclusionDetailLikenessSection` (private) | Heading "Likenesses" with count, adaptive grid of 120pt columns, tile opens `tile.location` via `navigation.go(to:)` (it reads `@Environment(WorkspaceNavigation.self)`), Portrait Badge / Use as portrait from `ConclusionPortraitModel` |
| Header | same file `ConclusionDetailHeader` (private) | 120pt thumb (with "Use earliest likeness" right-click menu when `offersEarliest`), Badge, title (display 3; a ref title mono h1), `summary(content)`, ref and members at the right |
| Row | `macos/App/Features/Conclusions/ReconciledValueRow.swift` `ReconciledValueRow` | `let model`; `body`: rule, then `HStack(spacing: space6) { summary; disclosures }`, min height 52, then the other-values and Why disclosures; no per-caller mode |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView` | Convenience init; `summary: { PersonDetailVitals(vitals: $0.vitals) }`; private `PersonDetailVitals` is a `VStack` of two `ConclusionDatePlaceLine`s (`ForEach(vitals, id: \.abbreviation) { vital in … }`); `.vocabularyToastOverlay($portrait.toast, …)` |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` `PersonDetailContent` | `struct Vital`, `var vitals: [Vital]`, `private static func vital(…) -> Vital` = the header's b. and d. lines |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Conclusions.a11yVital(_:date:place:)`; `macos/App/Resources/Localizable.xcstrings` | Spoken b./d. line; no section keys |
| Kit chip | `macos/App/DesignSystem/Components/Chip/PVChip.swift` `PVChip`, `PVChipGroup` | `PVChip` has two inits (`_ titleKey: LocalizedStringResource` and `text: String`), each `isSelected`, `tone`, `isDashed`, `expands`, `selectionLift`, `action`; body is a `Button` with `.buttonStyle(.plain)`, fixed `frame(height: 28)`, caption type; no `.isSelected` trait, no focus indication. `PVChipGroup(style: .loose | .segmented)` lays them out |
| Segmented copies | `Features/Dates/DateValueEditorForm.swift` (kind, ~line 39), `Features/CitationComposer/CitationComposerAuthoringBox.swift` (~line 577), `Features/Properties/PropertiesView.swift` (~line 831) | Each hand-writes `PVChipGroup(style: .segmented) { PVChip(…, expands: true, selectionLift: true, action:) … }` with an identifier per chip; the composer disables the other chip when `lockKind` and labels the group `L10n.CitationComposer.observationKind`; Properties labels it `L10n.Properties.createValueType`; the date editor has a visible `L10n.Sources.dateKindLabel` and no group label |
| Web kit | the S14-D1 design system (its PVChip.jsx) | `PVChipGroup variant="segmented" label=…` with `PVChip expands selectionLift`: the kit's segmented control; no size or glyph props (the S13-D1 board overrode them) |
| Control sizes | `macos/App/DesignSystem/Components/Button/PVButton.swift` `PVControlSize` | `.sm`, `.md`, `.lg` with `height` 22/28/36 (`PVSpacing.controlHeight*`), `horizontalPadding`, `font`, `iconGlyphSize` 13/15/17; `PVButtonBody` (a `ButtonStyle`) reads `@Environment(\.isFocused)` and applies `.pvFocusRing` |
| Scroll reset | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `pvScrollsRowIntoView(_:anchor:animated:scrollsOnAppear:startsAtTopWhenRowsChange:)` | The kit's one `ScrollViewReader`: scrolls a row into view; `startsAtTopWhenRowsChange` scrolls to the first row when the rows change. README: "do not add another `ScrollViewReader` for this" |
| Width class | `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` `PVWidthClass`, `\.pvWidthClass` (S14-24) | Regular below `PVSpacing.widthWideBreakpoint` (1600), Wide at or above; measured once at `WorkspaceDestinationHost`; no page reads it yet |
| Tokens | `macos/App/DesignSystem/Tokens/PVSpacing.swift` | `space3` 6, `space5` 12, `space6` 16, `space7` 20, `space8` 24, `space9` 32, `gutterPage` 32, `widthInspector` 340, `widthContentMax` 1180, `widthWideBreakpoint` 1600; no `widthSecondaryMax` |
| Kit README | `macos/App/DesignSystem/README.md` component table | Has Chip, ScrollRowIntoView and WidthClass rows; no Page columns or segmented-control row |
| Brief skill | `.agents/skills/add-design-brief/SKILL.md` | Paste block "Reach for (kit)" has no Page columns or segmented row; § "Wide layout (binding)" says a two-column scaffold is a question "raised once a second view needs it" |
| Location | `macos/App/Features/Workspace/WorkspaceLocation.swift` `WorkspaceLocation` | `entityId` names the person; memberwise `init(section:…listScrollAnchor:)`; tolerant `init(from:)` assigns fields directly; `personDetail(entityId:ref:title:)`; `withListScrollAnchor` copies `self`; `SourceSurface` and `normalizedSurface` live here; `==` lists deep ids |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` spec `.personDetail` | `viewIdentity: { $0.entityId }` (a section change does not remount); `queryKeys: { project, location in … }` already receives the location |
| Navigation | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `go(to:)` | Coalesces when `location == currentLocation`; otherwise pushes and persists |
| Crumbs, jump rows | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems(for:names:goTo:)`, private `HistoryJumpRow.leafLabel` | The crumb leaf is `title ?? ref ?? "…"` for an entity location (`ref ?? title ?? "…"` otherwise) with id `"leaf-\(section.rawValue)"` and `action: nil`; the jump row's `leafLabel: String?` is `title ?? ref` for every location (nil shows no "›") |
| Proto | `api/proto/engine.proto` `WorkspaceLocation`; `GoStore.mapWorkspaceLocationToProto` | Search context only; `listScrollAnchor` is not on the wire |
| Kit empty | `DesignSystem/Components/EmptyState/` `PVEmptyState` | `init(glyph: PVGlyph, title: LocalizedStringResource?, message: String?, compact:, action:)`; `message` is a `String` |
| Glyphs | `DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol`; `DesignSystem/Recipes/Marks/PVMark.swift` `PVGlyph`, `PVGlyphView` | `.account`, `.gitBranch`, `.calendar`, `.photo`, `.library` exist |
| Deployment | `macos/Provenencia.xcodeproj/project.pbxproj` | `MACOSX_DEPLOYMENT_TARGET = 14.0` (`onScrollGeometryChange` is macOS 15); `onGeometryChange` is used in `PVWidthClass.swift` and `SourcesListView.swift` |
| Project file | same | Hand-managed (S14-15 not landed). S14-24 (`dabe3e7e`) added `PVWidthClass.swift` and `PVWidthClassTests.swift` with a build file, a file reference, a group entry and a Sources phase entry each (ids `5C…F24A1/A2`, `AA…F24B1/B2`) |
| Spike 14 | `docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` (Blocked — re-plan) | Replaces `likenesses` with a `gallery` slot and renames `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection` |
| Shortcuts | `rg keyboardShortcut macos/App` | ⌘1 – ⌘9 unused |
| Tests | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift`, `PersonDetailTests.swift` (lines 268, 269, 318 – 321 read `vitals`; 318 is a local `let vitals`), `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PVWidthClassTests.swift`, `HostedWindow.swift` (`hostInBorderlessWindow`, `waitUntil`) | Push/restore, legacy decode, equality, crumbs; Person content; width class by pane width and in a hosted window |

**Precedent.** `listScrollAnchor` (S10) added an optional, Swift-only, tolerant-decoding field with a copy-with helper and legacy-decode tests: copy its shape for `pageSection`, but the section **is** identity. `SourceSurface` + `normalizedSurface` is the precedent for a small enum beside `WorkspaceLocation` with one private normalizer every construction path calls. S13-09 ([#624](https://github.com/mendahu/provenencia/pull/624)) is the precedent for the slot. S14-24 ([#628](https://github.com/mendahu/provenencia/pull/628), `dabe3e7e`) is the precedent for a new kit component folder: the file, its test suite, four pbxproj entries per file, a README row, table-driven tests plus a `HostedWindow` test. `PVButtonBody` is the precedent for a kit button drawing focus through its `ButtonStyle`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `enum ConclusionPageSection: String, Codable, CaseIterable, Sendable` *(new)* in `WorkspaceLocation.swift` beside `SourceSurface`, cases `vitals, family, timeline, gallery, sources` in that order. Raw values are frozen: they are written to history files. `static func offered(by section: WorkspaceSection) -> [ConclusionPageSection]` returns all five for `.persons` and `[]` for every other section. It owns `label: LocalizedStringResource` (D-10) and `glyph: PVGlyph`: vitals `.symbol(.account)` (the board's `user-round`), family `.symbol(.gitBranch)`, timeline `.symbol(.calendar)`, gallery `.symbol(.photo)`, sources `.symbol(.library)` | Kind-neutral (review M-1): Event or Place adopts sections by returning its cases from `offered(by:)` and adding any case of its own to the enum, with no second field, key, `==` line or label branch. Existing `PVSymbol`s | A Person-only enum and field; new marks |
| D-2 | `WorkspaceLocation.pageSection: ConclusionPageSection?` *(new)*, coding key `pageSection`, memberwise `init` parameter `pageSection: ConclusionPageSection? = nil` (after `entityId`). One private `static func normalizedPageSection(_:section:entityId:) -> ConclusionPageSection?` returns nil when `entityId` is nil, when the value is not in `offered(by: section)`, or when it equals `offered(by: section).first`; otherwise the value. `init`, `init(from:)` (decodes a `String?`, maps it with `ConclusionPageSection(rawValue:)`, unknown → nil) and `withPageSection(_:)` *(new)* all call it. `withPageSection` copies `self`, sets the field through the normalizer, and keeps `ref`, `title`, `sourceTitle`. `==` compares `pageSection`. `personDetail(entityId:ref:title:)` gains `section: ConclusionPageSection? = nil`. Not on the proto | Sections are places; one normalizer for every path (review X-1); the first section is stored as nil, so it equals a link without a section and `go(to:)` coalesces | Storing `.vitals`; adding it to `engine.proto` (`listScrollAnchor` precedent) |
| D-3 | `PersonDetailView` does not read navigation: it passes its fill through the full initializer, and `PersonDetailSections` *(new, private in `PersonDetailView.swift`)* switches on `context.section` (D-5). `.vocabularyToastOverlay` and the `portrait` argument stay | The page owns the section for every kind (review M2-2); `viewIdentity` stays `entityId`, so `ConclusionPortraitModel` and its toast survive a switch | Each kind reading navigation and re-deriving the section |
| D-4 | Split `ConclusionDetailDefaultPageSections` into named internal parts in the same file: `ConclusionDetailRowsSection` *(new)* with `rows: [ReconciledValueRowModel]`, `refreshing: Bool`, `title: LocalizedStringResource?` (nil draws no heading and no spinner) and `provenance: ReconciledValueRow.Provenance = .why` (D-19), and `ConclusionDetailLikenessSection` changed from `private` to internal with no other change. The default fill becomes `ConclusionDetailRowsSection(rows: content.rows, refreshing: context.refreshing, title: L10n.Conclusions.detailsSection)`, the likeness `if`, and the relationship `ForEach`, with unchanged output | Map scope "Reuse first" (S13-09 MNT-1); Vitals and Gallery reuse by name | Copying the heading, spinner and rows into the Person view |
| D-5 | **The page owns sections.** `ConclusionDetailPageSectionsContext` gains `section: ConclusionPageSection?`. `ConclusionDetailLoaded` reads `@Environment(WorkspaceNavigation.self)` (as `ConclusionDetailLikenessSection` does in this file) and computes `let offered = ConclusionPageSection.offered(by: navigation.currentLocation.section)` and `let section = offered.isEmpty ? nil : (navigation.currentLocation.pageSection ?? offered[0])`. When `offered` is non-empty the page draws the switcher (D-20) under the header, pins the header (D-17), shows refreshing in the switcher row (D-18), and passes `section` in the context and as the scroll reset key (D-12); choosing calls `navigation.go(to: navigation.currentLocation.withPageSection(choice))`. `PersonDetailSections(content:context:)` returns sibling views from one `switch context.section ?? .vitals`: `.vitals` → `ConclusionDetailRowsSection(rows: content.rows, refreshing: context.refreshing, title: nil, provenance: .sheet)`; `.gallery` → `ConclusionDetailLikenessSection(projectDir: context.projectDir, likenesses: content.likenesses, portrait: context.portrait)` when `content.likenesses` is non-empty, else `PVEmptyState(glyph: ConclusionPageSection.gallery.glyph, title: L10n.Conclusions.galleryEmptyTitle, message: L10n.string(L10n.Conclusions.galleryEmptyMessage))`; `.family`, `.timeline`, `.sources` → `PVEmptyState(glyph: section.glyph, title: section.label, message: L10n.Conclusions.pageSectionLater(L10n.string(section.label)))` | One type fewer and one owner of the selected section and its action (review M2-2); Event and Place adopt sections with one `offered(by:)` line and a `switch` in their fill; `PVEmptyState.message` is a `String` | A `ConclusionDetailSectioning` value each kind builds; a second view slot for the switcher |
| D-6 | **`PVChip` (kit, every caller).** Both initializers gain `size: PVControlSize = .md` and `glyph: PVGlyph? = nil` (after `tone`). Height is `size.height`; the font is `size == .lg ? PVFont.body(size: PVTypeScale.bodySmall, weight:) : PVFont.body(size: PVTypeScale.caption, weight:)` (`.md` keeps today's caption; `.sm` is caption too); horizontal padding stays `expands ? space4 : space5`. A private `extension PVControlSize { var chipGlyphSize: CGFloat }` in `PVChip.swift` gives 12 / 13 / 14 for `.sm / .md / .lg` (the board's 14 at `.lg`). The glyph is a `PVGlyphView` at that size, `space3` before the label, in the label's color; with a glyph and `expands`, the content is `ViewThatFits(in: .horizontal) { glyph + label; label }`, so a narrow column drops the glyph before the label truncates. The `Button` renders through a private `PVChipStyle: ButtonStyle` that reads `@Environment(\.isFocused)` and applies `.pvFocusRing(isFocused, cornerRadius: chipRadius)` (as `PVButtonBody` does), and carries `.accessibilityAddTraits(isSelected ? .isSelected : [])`. No `.focused` binding inside the kit | One size vocabulary for kit controls (review M2-1: `PVControlSize` already has 28 and 36); kit buttons show focus through their style (review M2-6); defaults keep the ten existing call sites' look | A `PVChip.Size` enum; a private `@FocusState` in the kit control; hard-coded 36 and 14 |
| D-7 | Each empty section is one `PVEmptyState(glyph:title:message:)`; Family, Timeline and Sources share one format key | User, 2026-10-29: placeholders show an empty state; board 4l – 4n | Hiding unbuilt sections (user declined); skeleton rows |
| D-8 | `WorkspaceToolbar` gains `static func locationLabels(for location: WorkspaceLocation) -> (leaf: String?, section: String?)` *(new)*: `leaf` is `title ?? ref` (today's jump-row rule, which is also the entity crumb's rule before its "…" fallback), `section` is `L10n.string(location.pageSection.label)` when `pageSection != nil`. `HistoryJumpRow.leafLabel` (still `String?`) returns `L10n.Conclusions.pageSectionLeaf(leaf, section)` when both are set, else `leaf`, so every non-Person row reads as today. `breadcrumbItems`' entity branch uses `leaf ?? "…"`; with a section, the entity crumb gets id `"entity-\(location.section.rawValue)"` and `action: { goTo(location.withPageSection(nil)) }`, and a third crumb with id `"leaf-\(location.section.rawValue)"`, the section label and `action: nil` follows; without a section the crumbs are as today. The non-entity branch (`ref ?? title ?? "…"`) is not touched | One rule where the two agreed, and only one place adds the section (review M-5); nothing changes for Sources, Metadata or Properties (review X2-6) | One function for every crumb branch (it would change Source crumbs) |
| D-9 | ⌘1 – ⌘5 are `.keyboardShortcut(KeyEquivalent(Character(String(index + 1))), modifiers: .command)` per switcher item (D-20's `Item.shortcut`), `index` from `offered.enumerated()`, so they act only while a sectioned page shows. Keyboard focus moves between chips with Tab when Full Keyboard Access is on, as every kit button; no arrow-key handling | Shortcuts scoped by the view, not `ProvenenciaApp.commands`; arrow keys would need custom key handling (review X-4) | App-menu commands; `onKeyPress` |
| D-10 | New copy in `L10n.Conclusions`, each with a `comment:` per `add-localized-string`: `pageSectionVitals` "Vitals", `pageSectionFamily` "Family", `pageSectionTimeline` "Timeline", `pageSectionGallery` "Gallery", `pageSectionSources` "Sources" (keys `conclusions.section.<raw>`), `pageSectionsLabel` "Sections" (`conclusions.sections`), `pageSectionLater(_ section: String) -> String` "%@ arrives in a later version of Provenencia." (`conclusions.section.later`), `galleryEmptyTitle` "No likenesses yet" (`conclusions.gallery.empty.title`), `galleryEmptyMessage` "A likeness is drawn on a Citation’s image — draw one there and it appears here." (`conclusions.gallery.empty.message`), `detailRefreshing` "Refreshing" (`conclusions.detail.refreshing`), `pageSectionLeaf(_ title: String, _ section: String) -> String` "%@ · %@" (`conclusions.section.leaf`). The gallery keys are named so S14-02 can move them onto its `gallery` slot | `add-localized-string`; one namespace with the page's other copy | A new namespace for one page |
| D-11 | New files only for kit components: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` *(new)*, `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` *(new)*, `macos/ProvenenciaTests/PVPageColumnsTests.swift` *(new)*, `macos/ProvenenciaTests/PVSegmentedControlTests.swift` *(new)*, each added to `project.pbxproj` with a build file, a file reference, a group entry (a new `PageColumns` group under `Components`; the existing `Chip` group) and a Sources phase entry, copying S14-24's entries with fresh ids in the same scheme. Everything else stays in existing files | A kit component has its own file (S14-D1 inventory, `add-ui-component`); feature types stay beside their callers (S13-09 D-4) | Page columns inside `ConclusionDetailPage.swift` |
| D-12 | **One kit scroll reset.** `pvScrollsRowIntoView` gains `restartsAt top: ID? = nil, whenChanged key: AnyHashable? = nil`: when `key` changes and `top` is set, it calls `proxy.scrollTo(top, anchor: .top)` with no animation, beside its existing `startsAtTopWhenRowsChange`. `PVPageColumns` applies `.pvScrollsRowIntoView(Optional<String>.none, restartsAt: PVPageColumns.topAnchor, whenChanged: scrollResetKey)` to its `ScrollView`, where `topAnchor` (`"pv.pageColumns.top"`) is the id of the zero-height sentinel that is the first child of the scroll content, before the `space8` top padding, so a reset lands at offset 0 and the header is full (review X2-2). A width-class change is not a reset (rule 5) | The README's one-`ScrollViewReader` rule (review M2-5); the sentinel is the true top | A second `ScrollViewReader` in Page columns; anchoring on the padded column group (lands at 24pt, compact) |
| D-13 | Rename the header's b. and d. lines from "vital" to "life line": `PersonDetailContent.Vital` → `LifeLine`, `vitals` → `lifeLines`, `private static func vital(…)` → `lifeLine(…)`, private `PersonDetailVitals` → `PersonDetailLifeLines` (its `vitals` property → `lines`, its `ForEach` closure variable `vital` → `line`), `L10n.Conclusions.a11yVital` → `a11yLifeLine` (Swift symbol only; the string key does not change), and in `PersonDetailTests.swift` the reads at 268 – 269 and the local `let vitals` at 318 → `let lines` with its uses | One word, one concept (review M-6) | Two meanings of "vitals" side by side |
| D-14 | `struct PVPageColumns<Header: View, Primary: View, Secondary: View>: View` *(new)*, from S14-D1 WL-3 – WL-9. Inits: `init(headerPinned: Bool = false, scrollResetKey: AnyHashable? = nil, @ViewBuilder header: @escaping (_ isScrolled: Bool) -> Header, @ViewBuilder primary: () -> Primary, @ViewBuilder secondary: () -> Secondary)`; `extension PVPageColumns where Secondary == EmptyView { init(headerPinned:scrollResetKey:header:primary:) }`; `extension PVPageColumns where Header == EmptyView, Secondary == EmptyView { init(scrollResetKey:primary:) }`. `hasSecondary` is `Secondary.self != EmptyView.self`. It reads `@Environment(\.pvWidthClass)` and does not re-measure the class. Widths come from `static func columnWidths(pane: CGFloat, widthClass: PVWidthClass, hasSecondary: Bool) -> (primary: CGFloat, secondary: CGFloat?, group: CGFloat)`: Regular, or Wide without a secondary → primary = max(0, min(`widthContentMax`, pane − 2 × `gutterPage`)), secondary (if any) the same width, group = primary (WL-4, WL-6); Wide with a secondary → `available = pane − 2 × gutterPage − space9`, primary = max(0, min(`widthContentMax`, available − `widthInspector`)), secondary = clamp(available − primary, `widthInspector`, `widthSecondaryMax`), group = primary + `space9` + secondary (WL-5). Two private `Layout`s use that function from their proposed width (the pane): `PVPageColumnsLayout` places the columns (side by side at Wide, top-aligned, primary leading; stacked `space9` apart at Regular), and `PVPageColumnsBandLayout` sizes one child to `group` width; both center the group (WL-7). Unpinned: the scroll content is the sentinel, `space8` top padding, the header in a band, `space9`, the columns, `space8` bottom padding; `isScrolled` is always false. Pinned: a `VStack(spacing: 0)` of the header band (no padding of its own; the header supplies it), a 1pt `PVColor.borderDefault` rule across the pane, then the `ScrollView` (sentinel, `space8`, columns, `space8`). The component sets no background (the page owns it, WL-3; review M2-7). Primary precedes secondary in the view tree, so focus moves primary then secondary (WL-9) | S14-D1 is the agreed spec (working rule "Redesigns adopt Page columns"); one pure, tested sizing function; `Layout`s size from the proposal without state, so crossing 1600 keeps scroll and focus (rule 5) | `GeometryReader` with state; a kit fill on the bar; a full-width mode (Out) |
| D-15 | `PVSpacing.widthSecondaryMax: CGFloat = 480` *(new)* beside `widthInspector`, with the S14-D1 comment. `widthWideBreakpoint` stays 1600: the S14-D1 board's WL-1 says `1600px`; a test pins both values | S14-D1 handoff: the PR that builds Page columns checks the breakpoint | Changing either value here |
| D-16 | **The page on Page columns.** `ConclusionDetailPage` keeps its memberwise order (`…, portrait, pageSections`) and gains no property: sections come from navigation (D-5). The `static func page(identifier:body:)` is deleted. `ConclusionDetailPage.body` without a handle draws `PVPageColumns(primary: { Self.loading })`. `ConclusionDetailLoaded.body` draws one `PVPageColumns(headerPinned: !offered.isEmpty, scrollResetKey: section, header: { isScrolled in if case .content(let content, let refreshing) = presentation { ConclusionDetailHeaderBand(…) } }, primary: { switch presentation { case .loading: …; case .failed: …; case .content: VStack(alignment: .leading, spacing: PVSpacing.space9) { pageSections(content, context) } } })`, so loading and failure sit in the primary with no header (WL-3) and the fill's siblings keep today's spacing. `ConclusionDetailHeaderBand` *(new, private)* is the header plus, when `offered` is non-empty, the switcher row; it supplies the band's padding: unpinned, none (the page's `space8` top padding as today); pinned and full, `space9` top, `space7` between header and switcher, `space7` bottom; pinned and compact, `space5` for all three (board 4a, 4g). `.background(PVColor.surfacePage)` and `.accessibilityIdentifier(pageIdentifier)` stay on the outermost view | S14-D1 gate: "`ConclusionDetailPage.page` moves onto it"; at Regular a primary-only, unpinned Page columns has today's width, gutters and padding, so Event and Place look the same (AC-3.6) | Keeping `page` for Event and Place (two scaffolds) |
| D-17 | **Compact header (board PPF-3, 4g, 4h, 4w).** Pinned `PVPageColumns` reads the sentinel's `minY` in the `ScrollView`'s coordinate space `"pv.pageColumns.scroll"` with `onGeometryChange` (macOS 14) and the content and viewport heights with `onGeometryChange` on the content and the `ScrollView`; `offset = −minY` (positive when scrolled down), `overflow = contentHeight − viewportHeight`. `static func isScrolled(offset: CGFloat, overflow: CGFloat, wasScrolled: Bool) -> Bool` *(new, pure)*: false when `overflow < PVPageColumns.compactMinOverflow` (120pt, more than the full header's height over the compact one, so compacting cannot shrink the bar enough to undo itself, review X2-3); otherwise true past 8pt, false at 0 or less, `wasScrolled` in between. The switch is immediate (no animation), as on the board. The header builder receives it; `ConclusionDetailHeaderBand` draws `ConclusionDetailHeader(compact: isScrolled)` *(new parameter on the private view)*: compact is a 40pt `CachedThumbnail`, no Badge, the title at `PVTypeScale.h3` (a ref title at `h4` mono), the summary, ref and members at the right, vertically centered; the thumb's right-click menu stays in both forms. The kind's summary learns it is compact from `\.conclusionDetailHeaderCompact` *(new `EnvironmentValues` key in `ConclusionDetailPage.swift`, default false)*, which the header sets; `PersonDetailLifeLines` lays its two lines in an `HStack(spacing: PVSpacing.space6)` when it is true and a `VStack` otherwise (D-21). Event and Place never compact (they are not pinned) | The board pins and compacts at both widths; one header view in two forms; the kind owns its summary's layout (review X2-1) | `onScrollGeometryChange` (macOS 15); a second compact header type; a second summary closure |
| D-18 | **Refresh indicator (board PPF-8, 4t).** On a page with sections the switcher row is an `HStack(spacing: PVSpacing.space6)` of the segmented control (flexible) and, while `refreshing`, a small `ProgressView` and `L10n.Conclusions.detailRefreshing` in italic caption `textMuted`. Pages without sections keep the spinner in the Details heading as today | Pinned, so it shows at any scroll and in any section; Vitals has no heading | A spinner per section |
| D-19 | **Provenance mode on the row (board PPF-8).** `ReconciledValueRow` gains `enum Provenance { case why, sheet }` *(new)* and `var provenance: Provenance = .why`. `.why` is today's row. `.sheet` is the Person page's mode: in S13-10 it draws today's row and ends the main `HStack` with `Color.clear.frame(width: 28, height: 28).accessibilityHidden(true)`; S13-11 changes only the `.sheet` body (the control in that slot, the Why accordion out). `ConclusionDetailRowsSection` passes `provenance` through; only Person's Vitals sets `.sheet` | One caller-mode now instead of a Bool that S13-11 would pile a second Bool beside (review M2-3); names the Event and Place debt of working rule "The provenance sheet replaces the Why on the Person page only" | `reservesProvenanceSlot: Bool` |
| D-20 | **`PVSegmentedControl<Value: Hashable>` *(new, kit)*** in `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` *(new)*, the Swift form of the kit's segmented `ChipGroup` (user, 2026-10-29). `struct Item: Identifiable { enum Title { case key(LocalizedStringResource), text(String) }; var value: Value; var title: Title; var glyph: PVGlyph? = nil; var isDisabled = false; var shortcut: KeyEquivalent? = nil; var accessibilityIdentifier: String; var id: String { accessibilityIdentifier } }` with `init(_ title: LocalizedStringResource, value:glyph:isDisabled:shortcut:accessibilityIdentifier:)` and `init(text: String, value:…)` setting `.key` or `.text`; the control calls `PVChip`'s matching initializer for each case. `init(_ label: LocalizedStringResource, items: [Item], selection: Value, size: PVControlSize = .md, select: @escaping (Value) -> Void)`. Body: `PVChipGroup(style: .segmented) { ForEach(items) { PVChip(title, isSelected: $0.value == selection, size: size, glyph: $0.glyph, expands: true, selectionLift: true) { select($0.value) } .disabled(…) .keyboardShortcut(…, modifiers: .command) when set .accessibilityIdentifier(…) } }` with `.accessibilityElement(children: .contain)` and `.accessibilityLabel(Text(label))`. Callers move onto it unchanged in look: the date editor (label `L10n.Sources.dateKindLabel`, items point / range, `select` calls `draft.setKind`); the Citation composer (label `L10n.CitationComposer.observationKind`, items property / bridge with `isDisabled` from `lockKind`, `select` calls `onProperty()` or `onBridge()`); Properties (label `L10n.Properties.createValueType`, `items` from `SubjectPropertyValueType.researcherCreatable` with `init(text:)`, `select` sets `draft?.valueType`); the Person switcher (label `pageSectionsLabel`, `size: .lg`, glyphs, shortcuts). Every existing accessibility identifier is kept. `PVChipGroup(style: .segmented)` remains the layout it composes | The fourth hand-written copy is the trigger (rubric Q5; review M-2, M2-4); the user chose to consolidate here; selection by value plus a `select` action fits all four callers (navigation, a model setter, two actions with a lock, a binding) | A `Binding` selection (the composer and Person do not set state directly); leaving the three copies |
| D-21 | `PersonDetailLifeLines` reads `@Environment(\.conclusionDetailHeaderCompact)` and switches its stack; nothing else in the summary changes | The summary is the kind's view; an environment value reaches it without changing `summary`'s signature | A `compact` parameter on `summary` (all three kinds change) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVSegmentedControl` *(new)* | component | the three hand-written segmented groups; the web kit's segmented `ChipGroup` | Replaces the copies (D-20) |
| `PVChip` `size`, `glyph`, `PVChipStyle`, selected trait | kit extend | `PVControlSize`; `PVButtonBody` (focus); `PVRadio`, `PVSidebarNav` (trait, glyph slot) | Extends the one chip with the kit's own vocabulary |
| `PVPageColumns`, its two `Layout`s, `columnWidths`, `isScrolled` *(new)* | component | `ConclusionDetailPage.page`; `SourcePageIdentityHeader`'s copy of the column frame | Agreed on S14-D1 as the one scaffold; replaces `page`, and is the Source page's next home |
| `pvScrollsRowIntoView(…restartsAt:whenChanged:)` | kit extend | its `startsAtTopWhenRowsChange` | One `ScrollViewReader` (D-12) |
| `PVSpacing.widthSecondaryMax` *(new)* | token | `widthInspector` | S14-D1 WL-1 |
| `ConclusionPageSection`, `WorkspaceLocation.pageSection`, `withPageSection`, `normalizedPageSection` *(new)* | type, field, funcs | `SourceSurface`, `listScrollAnchor`, `withListScrollAnchor`, `normalizedSurface` | Same shapes; the section is identity |
| `ConclusionDetailPageSectionsContext.section` | field | the context | The fill learns its section where it learns the rest |
| `ConclusionDetailHeaderBand`, `ConclusionDetailHeader(compact:)`, `\.conclusionDetailHeaderCompact` *(new)* | view, param, env key | `ConclusionDetailHeader` | One header in two forms |
| `ConclusionDetailRowsSection` *(new)* | view | the Details block inside `ConclusionDetailDefaultPageSections` | Extracted, not added beside it |
| `ReconciledValueRow.Provenance` *(new)* | enum | — | The by-caller mode the working rule names |
| `PersonDetailSections` *(new, private)* | view | `ConclusionDetailDefaultPageSections` | Person's fill differs by section; reuses the default's parts |
| `WorkspaceToolbar.locationLabels(for:)` *(new)* | func | `HistoryJumpRow.leafLabel`, the entity crumb leaf | One rule where they agree |
| L10n keys (D-10) *(new)* | key | `L10n.Conclusions.detailsSection` | Same namespace |
| Tests (§8) *(new)* | test | `PVWidthClassTests`, `HostedWindow`, `WorkspaceNavigationTests`, `WorkspaceToolbarBreadcrumbTests` | Same patterns |

**Builds on.** S14-24's width class and the S14-D1 board; S13-09's `pageSections` slot; `add-workspace-location`; `PVControlSize`, `PVButtonBody`'s focus pattern, `pvScrollsRowIntoView`, `PVEmptyState`.
**Deletes.** `ConclusionDetailPage.page(identifier:body:)` and its `ScrollView` frame; the three hand-written segmented groups in `DateValueEditorForm`, `CitationComposerAuthoringBox` and `PropertiesView`; `PVChip`'s fixed `frame(height: 28)`; the single-block body of `ConclusionDetailDefaultPageSections`; `private` on `ConclusionDetailLikenessSection`; the names `Vital`, `vitals`, `vital(…)`, `PersonDetailVitals`, `a11yVital`; `PersonDetailView`'s use of the convenience initializer. No callee loses its last caller.
**Next adds.** S13-11 changes `ReconciledValueRow`'s `.sheet` body. S13-12, S13-13 and S13-16 each replace one empty-state case in `PersonDetailSections`; each section's query key is warmed from the `.personDetail` spec's `queryKeys` in `PlaceRegistry`, which already receives the location, and the section view reads a `QueryHandle`. A section that wants a secondary column at Wide adds a secondary fill to the page, drawn on its board. S13-14 replaces the Gallery case. Event and Place adopt sections by returning cases from `offered(by:)` and switching on `context.section`. The Conclusion lists and the Source page move onto `PVPageColumns` (pinned mode already built). The next segmented picker is one `PVSegmentedControl`. S14-02 renames `ConclusionDetailLikenessSection` and moves the two gallery keys.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A location with a section for a person that was deleted | `ConclusionDetailLoaded` failed state | The danger Callout in the primary; no header, no switcher |
| Restoring a non-Vitals section from history on launch | `ConclusionDetailLoaded` | Section read from `currentLocation` on every render |
| Switching person (a family tile later) while on Gallery | `PlaceRegistry` remounts on `entityId` | A new person's location has no section → Vitals |
| Gallery with no likenesses | `PersonDetailSections` | Empty state |
| Every `PVChip` caller now carries the selected trait and a focus ring | the chip call sites, loose and segmented | VoiceOver gains "selected"; the ring shows only with keyboard focus |
| Three pickers now go through `PVSegmentedControl` | date editor, composer, Properties | Same look at `.md`, same identifiers, composer lock kept (AC-4.1) |
| Event and Place on `PVPageColumns`, every page can be Wide | `EventDetailView`, `PlaceDetailView` | Regular: today's frame; Wide: centered; header unpinned |

### Scale

None — no write and no new read: switching re-renders from the one cached `conclusionDetail`. A person with 20 rows and 30 likenesses renders one section at a time. The `Layout`s and `isScrolled` are constant work per pass.

### Edges

- **Assumptions.** `go(to:)` coalesces equal locations, so `==` includes `pageSection` (T-3) and every path normalizes the first section to nil (T-2). `columnWidths` never returns a negative width (T-12). `isScrolled` cannot oscillate: compacting frees less height than `compactMinOverflow` (T-15).
- **Neighbors.** Event, Place and list-root locations never carry a section (T-4). Panes 1599 and 1600 give Regular and Wide (S14-24 tests); at 1600 with a secondary 1164 / 340, at 1616 1180 / 340, at 1756 1180 / 480 (T-12). Offsets 8 and 9, overflow 119 and 120 (T-15). Jump rows for Sources, Metadata and Properties are unchanged (T-7).
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

From the S13-D1 inventory and both boards. The frames, in words (board ids):

- **4a Regular Vitals.** Pinned band: `space9` top padding; the full header (120pt thumb with the gender tint outline, the Badge eyebrow, the name at display 3, b. and d. lines, ref and members at the right); `space7`; the switcher across the 1038pt column: five 36pt chips (`.lg`), each glyph + label, Vitals selected (raised, lifted, semibold); `space7`; then a 1pt rule across the pane. Below, `space8`, then the Vitals rows with no heading, each ending in the reserved slot (drawn tinted only here, to show it; empty in the app).
- **4b Wide Vitals.** The same at a 2296 pane: band and rows in the centered 1180pt column; no secondary column; the rule spans the pane.
- **4g, 4h Scrolled.** The compact band: `space5` padding, a 40pt thumb, the name at h3 with the b. and d. lines on one line beneath it, ref and members at the right; `space5`; the switcher; `space5`; the rule. Rows scroll beneath it.
- **4i, 4j Gallery.** Today's Likenesses section unchanged (heading and count, 96pt tiles, Portrait Badge, Use as portrait), primary-only at Wide.
- **4k Gallery, none.** The EmptyState with the photo glyph, "No likenesses yet" and its message; the header shows the person mark with the Man tint.
- **4l – 4n, 4o Family, Timeline, Sources.** The EmptyState with the section glyph, the section name, and "<Section> arrives in a later version of Provenencia."; centered primary-only at Wide.
- **4p Sparse.** A ref-only person: the ref as a mono title, the four life rows stated empty.
- **4q History.** The toolbar above the pane: Persons › Margaret Ellen Robins › Gallery, and the Back jump menu listing "Margaret Ellen Robins" (the Vitals entry) and "Persons".
- **4r, 4s Loading, failure.** One small spinner in the primary; the danger Callout in the primary; no header or switcher.
- **4t Refreshing.** D-18's spinner and "Refreshing" at the end of the switcher row; the chips narrow to make room.
- **4u Focus.** The kit focus ring on the Timeline chip.
- **4v – 4y Dark.** 4a, 4g, 4l and 4b in dark mode.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Scaffold | `PVPageColumns` *(new)* | `DesignSystem/Components/PageColumns/PVPageColumns.swift` | D-14, D-17; README row "PageColumns" linking § 9 |
| Segmented control | `PVSegmentedControl` *(new)* | `DesignSystem/Components/Chip/PVSegmentedControl.swift` *(new)* | D-20; README row "SegmentedControl" |
| Chip | `PVChip` (extended) | `DesignSystem/Components/Chip/PVChip.swift` | D-6 |
| Page | `ConclusionDetailPage` on `PVPageColumns` | `Features/Conclusions/ConclusionDetailPage.swift` | D-5, D-16; Event and Place keep the convenience init |
| Header band | `ConclusionDetailHeaderBand`, `ConclusionDetailHeader(compact:)` | same | D-16, D-17, D-18 |
| Vitals | `ConclusionDetailRowsSection`, `ReconciledValueRow(provenance: .sheet)` | same; `ReconciledValueRow.swift` | D-4, D-19 |
| Gallery | `ConclusionDetailLikenessSection` / `PVEmptyState` | same / kit | D-5 |
| Family, Timeline, Sources | `PVEmptyState` | kit | D-7 |
| Crumbs, jump rows | `PVBreadcrumbs`, `HistoryJumpRow` via `locationLabels` | `Features/Workspace/WorkspaceToolbar.swift` | D-8 |

States: loading and failed in the primary, no header (WL-3); refreshing at the end of the switcher row (D-18). Widths: Regular 1366 × 900 and Wide 2560 × 1440, sidebar open. Accessibility: chips are buttons with names and the selected trait; each segmented group has a label; empty states read title then message; focus primary then secondary.

### Strings

L10n keys *(new)*: as D-10. Renamed Swift symbol (key unchanged): `L10n.Conclusions.a11yVital` → `a11yLifeLine`. Removed: none.

## 7. Steps

**STEP-1 — The section is part of the location**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-workspace-location`, `add-swift-test`, `add-localized-string`.
Do: D-1 (with the five section-name keys of D-10), D-2. Write T-1 – T-5, T-11.
Check: `xcodebuild test` runs `WorkspaceNavigationTests` green; `python3 scripts/check-localizable-xcstrings.py` clean.

**STEP-2 — "Vitals" names only the section**
Files: `macos/App/Features/Conclusions/PersonDetailContent.swift`, `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-13.
Check: `rg -w 'Vital|vitals|vital|a11yVital|PersonDetailVitals' macos/App/Features macos/App/Platform macos/ProvenenciaTests` finds nothing (`macos/App/DesignSystem`'s "Vital records" preview string is not in scope); `PersonDetailTests` green.

**STEP-3 — One chip, one segmented control**
Files: `macos/App/DesignSystem/Components/Chip/PVChip.swift`, `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift` *(new)*, `macos/ProvenenciaTests/PVSegmentedControlTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/Features/Dates/DateValueEditorForm.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`, `add-swift-test`.
Do: D-6, D-20 (the three existing callers), D-11's entries for the two files; README rows for SegmentedControl and the Chip row's new `size` and `glyph`.
Check: T-16 green; `rg -F 'style: .segmented' macos/App/Features` finds nothing; the three pickers look as before (§10 step 11).

**STEP-4 — Page columns in the kit**
Files: `macos/App/DesignSystem/Tokens/PVSpacing.swift`, `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift`, `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` *(new)*, `macos/ProvenenciaTests/PVPageColumnsTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`, `add-swift-test`.
Do: D-15, D-12, D-14, D-17's `isScrolled`, sentinel and measurements, D-11's entries; a README row "PageColumns" after "WidthClass" naming the slots, the pinned mode, the sizing owner and a link to § 9 of `docs/macos-client-patterns.md`.
Check: T-12 – T-15 green.

**STEP-5 — Split the default fill and give the row its provenance mode**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Conclusions/ReconciledValueRow.swift`. Skill: `add-ui-component`.
Do: D-4, D-19.
Check: the detail suites green; Event and Place unchanged.

**STEP-6 — The Conclusion page sits on Page columns and owns sections**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: D-16, D-5 (the page half: `offered`, `section`, the context field, `select`), D-17 (the header band, compact header, environment key), D-18, D-9 (items' shortcuts), the `pageSectionsLabel` and `detailRefreshing` keys.
Check: the detail suites green; §10 step 9 (Event and Place unchanged at 1366, centered at 2560).

**STEP-7 — Person sections**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-3, D-5 (`PersonDetailSections` and the full init), D-7, D-21, D-10's remaining keys.
Check: T-10; §10 steps 1 – 8 and 10.

**STEP-8 — Crumbs and jump rows name the section**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: D-8. Write T-6, T-7.
Check: `WorkspaceToolbarBreadcrumbTests` green.

**STEP-9 — Docs and skills**
Files: `docs/macos-client-patterns.md` (under the sectioned detail page bullet: a kind with sections is offered them by `ConclusionPageSection.offered(by:)`, its section lives on `WorkspaceLocation.pageSection`, and the page reads it, pins the header and passes it in the fill's context; in § 9, one sentence naming `PVPageColumns` as the record-page scaffold), `docs/ideas/wide-layout.md` (the Person, Event and Place row: on Page columns), `.agents/skills/add-design-brief/SKILL.md` (Paste block "Reach for (kit)": rows "Record or detail page at Regular and Wide → Page columns" and "Choose one of a few options inline → Segmented control (ChipGroup segmented)"; in "Wide layout (binding)", replace the "raised once a second view needs it" sentence with "Reach for Page columns (`PVPageColumns`) for a record page's primary and secondary columns"), `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/archive/S13-D1-person-page-frame.md` *(moved)*, `../design/README.md`.
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
| T-7 | Swift | `WorkspaceToolbarBreadcrumbTests` `locationLabelsNameTheSection` *(new)* | Family with title "Jane Doe" → ("Jane Doe", "Family"); Vitals → ("Jane Doe", nil); no title, ref "PER-1" → leaf "PER-1"; a Source page location with ref and title → leaf is the title (today's jump-row rule), section nil; a metadata root → (nil, nil) | AC-2.8, D-8 |
| T-8 | Swift | `PersonDetailTests`, `EventDetailTests`, `PlaceDetailTests` existing (renamed reads per D-13) | unchanged | AC-5.1, D-13 |
| T-9 | Manual | §10 | sections, empty states, VoiceOver, keyboard, scroll to top, compact header, refreshing, reserved slot, Regular and Wide, the three pickers | AC-1.1 – AC-1.10, AC-2.6, AC-2.9, AC-3.1 – AC-3.6, AC-4.1, AC-4.2 |
| T-10 | Structural | STEP-7 `rg` | In `macos/App/Features/Conclusions/PersonDetailView.swift`: `rg -F 'ForEach(content.rows)'` and `rg -F 'WorkspaceNavigation'` find nothing; `rg -F 'ConclusionDetailRowsSection'`, `rg -F 'ConclusionDetailLikenessSection'` and `rg -F 'context.section'` each find a hit | AC-5.2, D-3, D-5 |
| T-11 | Swift | `WorkspaceNavigationTests` `sectionsOfferedPerKind` *(new)* | `offered(by: .persons)` → the five in order; `.events`, `.places`, `.sources` → `[]` | AC-5.3, D-1 |
| T-12 | Swift | `PVPageColumnsTests` `columnWidthsFollowTheBoard` *(new, table-driven)* | (pane, class, secondary) → (primary, secondary, group): (1102, R, yes) → (1038, 1038, 1038); (1102, R, no) → (1038, nil, 1038); (1600, W, yes) → (1164, 340, 1536); (1616, W, yes) → (1180, 340, 1552); (1656, W, yes) → (1180, 380, 1592); (1756, W, yes) → (1180, 480, 1692); (2296, W, yes) → (1180, 480, 1692); (2296, W, no) → (1180, nil, 1180); (40, R, no) → (0, nil, 0) | AC-5.4, D-14 |
| T-13 | Swift | `PVPageColumnsTests` `tokensMatchTheBoard` *(new)* | `widthWideBreakpoint == 1600`, `widthSecondaryMax == 480`, `widthInspector == 340`, `widthContentMax == 1180` | D-15 |
| T-14 | Swift | `PVPageColumnsTests` `wideHostPlacesTheSecondaryBeside` *(new)* | `hostInBorderlessWindow` a `PVPageColumns` with probe views reporting their frames (via `onGeometryChange`) at pane widths 2296 and 1102 with `\.pvWidthClass` set to match; `waitUntil` both frames are known → secondary beside the primary at Wide, below it at Regular | AC-3.1, AC-3.4, D-14 |
| T-15 | Swift | `PVPageColumnsTests` `scrolledFollowsTheBoard` *(new, table-driven)* | (offset, overflow, was) → isScrolled: (0, 500, false) → false; (8, 500, false) → false; (9, 500, false) → true; (4, 500, true) → true; (0, 500, true) → false; (−20, 500, true) → false; (50, 119, false) → false; (50, 120, false) → true | AC-1.7, AC-1.10, D-17 |
| T-17 | Structural | STEP-3 `rg` | `rg -F 'style: .segmented' macos/App/Features` finds nothing | AC-5.5, D-20 |
| T-16 | Swift | `PVSegmentedControlTests` *(new)* `selectSendsTheItemValue`, `disabledItemDoesNotSelect` | A control over three items with one disabled, hosted with `hostInBorderlessWindow`; invoking each enabled chip's action (through its accessibility identifier) calls `select` with that item's value; the disabled one does not | AC-4.1, D-20 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceNavigationTests` (`==` change), the detail suites (STEP-2, STEP-5, STEP-6), the date editor, composer and Properties suites (STEP-3), `PVWidthClassTests` and `WorkspaceDestinationHostTests` (STEP-4, STEP-6).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`PersonDetailContent.Vital`, `PersonDetailVitals`, `a11yVital`, `ConclusionDetailPage.page`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the S13-D1 inventory and the S14-D1 board; no new kit primitive beyond `PVPageColumns` and `PVSegmentedControl`
- [ ] Each view checked at Regular (1366 × 900) and Wide (2560 × 1440), sidebar open (§10)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc and skill updated: macos-client-patterns.md (sections owned by the page; pageSection; PVPageColumns in § 9); ideas/wide-layout.md Person row; add-design-brief (Page columns and Segmented control rows, the Wide layout sentence)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

At 1366 × 900, sidebar open, unless a step says Wide (2560 × 1440, sidebar open). Turn on Full Keyboard Access for step 7.

1. Open a project with a Person that has a name, both life lines, three likenesses, a Mixed row and 20 rows. Open them from the Persons list: header as before, Vitals selected, rows as before with no heading. (AC-1.1)
2. Choose Gallery: today's likeness section with its heading, Portrait Badge, Use as portrait; choose one, see the toast; the header thumb follows. (AC-1.2)
3. Open a Person with no likeness, choose Gallery: the empty state. Choose Family, Timeline, Sources: each empty state with its glyph and line. (AC-1.3, AC-1.4)
4. Back three times from Gallery after Vitals → Family → Gallery: Family, Vitals, Persons list. Choose Gallery twice: one Back step. (AC-2.1, AC-2.2)
5. On Gallery, click a tile, then Back: Gallery. On Timeline, quit and reopen: Timeline. Scroll Vitals to the bottom, choose Gallery: it shows from the top with the full header. (AC-2.6, AC-2.3, AC-2.9)
6. Breadcrumbs on Gallery; click the person crumb → Vitals; open the Back jump menu: "Jane Doe · Family" and "Jane Doe"; a Source page entry reads as before. (AC-2.7, AC-2.8)
7. VoiceOver on the switcher; Tab to it (focus ring), ⌘3 → Timeline. (AC-1.5, AC-1.6)
8. Wide: the same Person on Vitals, Gallery and Family match board frames 4b, 4j and 4o (AC-3.1, AC-3.2, AC-3.3). Scroll Vitals halfway, focus the switcher, then drag the window narrower across the line: still Vitals, same scroll, still focused, still compact (AC-3.5).
9. Open an Event and a Place at 1366: the same page as before, header scrolling with the page. At Wide: the same column, centered. (AC-3.6)
10. On Vitals scroll down: the header compacts, the switcher stays; back at the top it expands (AC-1.7). On a person whose Vitals barely overflow, scroll to the bottom: no flicker (AC-1.10). Edit a record in another window so the page refreshes: the spinner and "Refreshing" at the end of the switcher row (AC-1.8). Each Vitals row ends in an empty 28pt slot; VoiceOver does not stop on it; Event rows have none (AC-1.9).
11. Open the date editor, the Citation composer (with the kind locked, then unlocked) and New Property: each picker looks and works as before; VoiceOver reads its label and the selected chip as selected. (AC-4.1, AC-4.2)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Adding `pageSection` to `==` changes coalescing for Person entries | T-2 normalizes the first section; T-1 covers restore |
| Moving the page onto `PVPageColumns` shifts Event and Place at Regular | D-16: same max width, gutters, padding and inter-section spacing; §10 step 9; stop if they differ |
| Moving three pickers changes their look or identifiers | D-20 keeps `.md`, the identifiers and the composer lock; §10 step 11; their suites |
| Pinning changes the page for Event and Place | Only a kind with offered sections pins |
| This is a large PR | Page columns (S14-25) and the picker consolidation were folded in by the user; each step builds and checks on its own |
| A ⌘ shortcut on a chip fires while a sheet is open on the page | Manual check; stop if a shortcut changes section behind a modal |
| S14-02 lands first and renames `ConclusionDetailLikenessSection` and `likenesses` | Refresh §4 and D-4, D-5 to its names before STEP-5 |

**Stop and report if:**

- `ConclusionDetailPage` no longer has the full initializer or the `pageSections` slot §4 describes, or `PVWidthClass` no longer reads as §4 says.
- Building a frame needs a Person-only page type, a second view slot, a third column, a Page columns full-width mode, or a change to Event or Place layout.
- The S14-D1 board's breakpoint or token values differ from D-15.
- The sentinel cannot report the scroll offset on macOS 14 without a `GeometryReader` in state, or `isScrolled` still oscillates on short content.
- `PVChip` cannot take a `ButtonStyle` focus ring or the `.lg` size without changing the look of an existing caller at `.md`.
- A migrated picker cannot keep its look, its identifiers or the composer's lock through `PVSegmentedControl`.
- Five `.lg` chips with their glyphs do not fit across a 1038pt column at `bodySmall`.
- A `Layout` cannot place the band and columns per D-14 without measuring width in state.

## 12. Open questions

None. The S13-D1 board (agreed 2026-10-29) answered the frame questions: pattern A, the header pinned and compacting on scroll, crumbs and jump rows as D-8, no Vitals heading with refreshing in the switcher row, ⌘1 – ⌘5 and Tab, Vitals and Gallery primary-only at Wide. The user answered the segmented-picker question on 2026-10-29: consolidate onto the kit's segmented control here (D-20).

## 13. PR

**Title:** S13-10 — Header and sections
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check at Regular and Wide; `Plan: docs/deployment-plan/spike-13/prs/S13-10-header-and-sections.md`.
