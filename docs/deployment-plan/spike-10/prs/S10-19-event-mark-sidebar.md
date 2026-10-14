# S10-19 — Event mark and sidebar icons

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 5 — Artifacts and multi-select
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-19--event-mark-and-sidebar-icons)
**Brief:** [S10-D16](../design/S10-D16-event-mark.md) (EM-1 to EM-4) and [S10-D7](../design/S10-D7-sidebar-marks.md) (WS-1 to WS-5), both boards agreed 2026-10-14
**Board:** S10-D16 is the Claude Design project *Provenencia Design System*, id 0f6c1f29-8408-4735-b114-5ae4c4c18445 (https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445). STEP-1 reads the Event asset from it. S10-D7 is the project *Main Application Layout*, id fb7b3e33-1683-4961-aec1-90e6e19f8214 (https://claude.ai/design/p/fb7b3e33-1683-4961-aec1-90e6e19f8214), file *Workspace Chrome.dc.html*, Part 3, Frames 12–16. §6 macOS describes every frame in words. The implementer needs the boards only for STEP-1.
**Depends on:** S10-D16, S10-D7
**Blocks:** —
**Branch:** `s10-19-event-mark-sidebar`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rules: "One Event mark"; [`macos/App/DesignSystem/Recipes/Marks/MARKS.md`](../../../../macos/App/DesignSystem/Recipes/Marks/MARKS.md) (templates, tint at the call site, decorative vs named); [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Planned:** 2026-10-14 against `main` @ `8b742aec`
**Review:** [`S10-19-event-mark-sidebar.review.md`](S10-19-event-mark-sidebar.review.md) — Fits, Ready with changes (applied); 2 rounds

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

The Event mark is an hourglass, so an event reads as a timer. This PR replaces that artwork with the calendar the design-system board drew (S10-D16). The same PR puts the subject marks in the workspace shell (S10-D7). Persons, Events, and Places use them in the sidebar, in the Back and Forward history menus, and in their empty list states. Then the shell and the cards agree. The dogfood note is "Rethink the Event icon + sidebar icons match Subject cards" (bar item 6).

To do that without a fourth hand-rolled "symbol or mark" branch, the PR adds one small value, `PVGlyph`. It names either an SF Symbol or a mark, and one view draws it. The section icon, the sidebar row, the history row, and the empty state carry that value.

**Why now.** Both boards are agreed. Slice 5's Check ends with "Sidebar Event matches the card."

**What stays.** The key `subject_event`. The Person, Place, bridge, and source artwork. `PVMark` stays the one mark view, and marks stay monochrome templates. Sources, Source types, Metadata, and Properties keep their SF Symbols everywhere. Row labels, order, counts, and selection stay as they are. Empty-state titles and messages stay as they are. Every other `PVEmptyState` caller keeps its `icon:` call unchanged.

## 2. User stories

**US-1.** As a researcher, when I see an event anywhere in the app, I see a calendar and not an hourglass, so an event does not read as a timer.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The repo's `provenencia_subject_event.svg` after STEP-1 | It draws the calendar: one `rect x="4.5" y="5.5" width="15" height="15" rx="2"`, the band `M4.5 10h15`, the rings `M8.5 3.5v3.5M15.5 3.5v3.5`, and five `h.01` dots. The hourglass path `M8.2 3.6v2.9` is gone. |
| AC-1.2 | An Evidence graph with an event card, at 100% zoom | The card's kind mark is the calendar, in Event copper |
| AC-1.3 | Omnibar: type the title of a promoted Event | The Event hit's lead mark is the calendar |

**US-2.** As a researcher, when I scan the sidebar, Persons, Events, and Places show the same marks as their cards, so I can tell which list I am in at a glance.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Sidebar expanded, Events selected | Persons, Events, and Places rows show `subject_person`, `subject_event`, and `subject_place` at 15pt. Sources shows the library symbol. Source types, Metadata, and Properties show tag, list, and list-tree. Labels, order, and counts are unchanged. |
| AC-2.2 | The same, at rest vs selected | A mark takes the row's text ink: secondary at rest, primary on the selected row, the same as a symbol row |
| AC-2.3 | Sidebar collapsed to the rail | The three marks draw at 16pt in the 32pt rail buttons. Hovering a mark row shows its label as a tooltip. VoiceOver reads the row label (Persons, Events, Places), not the mark name. |

**US-3.** As a researcher, when I hold Back or Forward, each history row shows the same mark as its sidebar row, so the jump menu and the sidebar agree.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Visit Sources, then Events, then a Person page; long-press Back | The Events row shows the calendar mark at 15pt. The Sources row shows the library symbol. Both use the menu's secondary ink. |
| AC-3.2 | From that state press Back twice, then long-press Forward | The Events and Persons rows show their marks. The Person page row shows the person mark beside *Persons › {name}* and its ref. |

**US-4.** As a researcher on a fresh project, an empty Persons, Events, or Places list shows that kind's mark, so the empty page matches the sidebar row I clicked.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Fresh project, open Events | The empty state shows the calendar mark at 26pt in faint ink, the title "No events yet", and the existing message. The dashed sunken box is unchanged. |
| AC-4.2 | Fresh project, open Persons, then Places | "No persons yet" with the person mark; "No places yet" with the place mark |
| AC-4.3 | Any list kind | The empty-state mark and the sidebar row mark are the same key, because both read `Kind.mark` (D-4, D-9) |

## 3. Scope

| In | Covered by |
| --- | --- |
| EM-1 to EM-4: the `subject_event` asset is replaced with the kit's calendar export. The key is unchanged, and no other mark asset changes. | STEP-1, AC-1.1, AC-1.2, AC-1.3 |
| `PVGlyph` and `PVGlyphView` *(new)*: one value for "an SF Symbol or a mark", and one view that draws it (D-3) | STEP-2 |
| WS-1, WS-2: `WorkspaceSection.glyph` replaces `icon`. The sidebar's Persons, Events, and Places rows draw their mark at the row's icon size, expanded and collapsed. The other rows keep their symbols. | STEP-3, AC-2.1, AC-2.2, AC-2.3 |
| WS-3: the toolbar's section icon, which is the Back and Forward history jump-menu rows (D-2), draws the section's glyph | STEP-3, AC-3.1, AC-3.2 |
| WS-4: `PVEmptyState` takes a `PVGlyph`. The Persons, Events, and Places empty states pass their kind's mark. | STEP-4, AC-4.1, AC-4.2, AC-4.3 |
| WS-5: cards, omnibar hits, Promote, and Properties pick up the new asset with no code change | STEP-1, AC-1.2, AC-1.3 |
| Dead code this PR makes: `ConclusionListKind.emptyIcon` and its three implementations, `WorkspaceSection.icon`, `PVSymbol.person`, `PVSymbol.mapPin` | STEP-3, STEP-4, STEP-5 |
| Docs: `MARKS.md`, the DesignSystem README, `docs/design-system-layers.md`, the map, `completed.md`, the two briefs archived | STEP-7, STEP-8 |

| Out | Home |
| --- | --- |
| A redraw of the Person, Place, bridge, or source marks | not planned (S10-D16 §2.1) |
| New sidebar rows, a new order, or the Narrate row | not planned; Narrate waits for the Narrative layer (comment in `WorkspaceSidebarSections.make`) |
| A section icon in the breadcrumb, or anywhere in the toolbar outside the history rows | not planned (D-2) |
| Marks on the Sources, Source types, Metadata, or Properties rows | not planned (WS-1) |
| Moving `PVThumbnail.Content`, `PropertiesView.typeCard`, and `OmnibarResultsHost` to `PVGlyph` | [glyph consolidation](../../../ideas/glyph-consolidation.md) |
| One `PVMarkKey(subjectTypeKey:)` built on the registry's `IconSymbol` | [glyph consolidation](../../../ideas/glyph-consolidation.md) |
| `ConclusionListKind` deriving `title` and `identifierRoot` from a `WorkspaceSection` | not planned; raised in review M-1, no PR in the map needs it |
| The `emptyIcon` field on the vocabulary list strings (`VocabularyListPane`, `MetadataListPane`, `SourceTypesListPane`) | not planned; a different field from `ConclusionListKind.emptyIcon`, and it stays |
| Last updated, filters, or a likeness thumb on the lists | S10-22 (D9 to D11), S10-41 (D8), S10-35 / S10-39 |
| A `verbatimTitle:` glyph init on `PVEmptyState` | not planned; no caller (D-8) |
| Stale comments this PR does not need to touch: the DesignSystem README's EmptyState "action slot isn't ported" clause, the `PVSidebarNav` header's S5-D3 note, the `WorkspaceSection.persons` "stub pages until S9-23 / S9-26" comment | not planned; leave them |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Event artwork | `macos/App/Resources/Assets.xcassets/Marks/provenencia_subject_event.imageset/provenencia_subject_event.svg` | The hourglass (`M6.6 3.6h10.8M6.6 20.4h10.8` plus two bulbs that start `M8.2 3.6v2.9` and `M15.8 3.6v2.9`). It is template-rendered per `Contents.json` and has no trailing newline. Wrong for EM-1. |
| Mark recipe | `macos/App/DesignSystem/Recipes/Marks/PVMark.swift` `PVMark`, `PVMarkKey`, `PVFileTypeGlyph` | `PVMark.body` is `Image(key.assetName).renderingMode(.template)` with no foreground of its own. `decorative:` hides it from VoiceOver. The file already holds a second helper type (`PVFileTypeGlyph`). |
| Symbol icon | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol`, `PVIcon` | `PVIcon` is always `accessibilityHidden(true)` and has no rest dim. `PVSymbol.person` (`"person"`) and `PVSymbol.mapPin` (`"mappin"`) are read only by `WorkspaceSection.icon`, the two `emptyIcon`s, and the `PVSidebarNav` preview. `.calendar` has other readers (`PVBadge`, `MetadataFieldBadges`, `PVTable`). |
| Card mark | `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift` `mark(_:in:_:)` | Loads `NSImage(named: key.assetName)` and tints it. Picks up the new asset with no change. |
| Omnibar mark | `macos/App/Features/Workspace/OmnibarHitPresentation.swift` `leadMark(for:)` | `"event"` → `.subjectEvent`. Picks up the new asset. |
| Section icon source | `macos/App/Features/Workspace/WorkspaceSection.swift` `WorkspaceSection.icon` | A `PVSymbol` per section: `.library`, `.tag`, `.list`, `.listTree`, then `.person`, `.calendar`, `.mapPin` for the three Conclude sections. There is no mark. Wrong for WS-1 / WS-3. |
| Section icon readers | `macos/App/Features/Workspace/WorkspaceSidebarSections.swift:54`, `macos/App/Features/Workspace/WorkspaceToolbar.swift:427` | The only two readers of `section.icon` (`rg -n "section\.icon" macos`) |
| Sidebar rows | `macos/App/Features/Workspace/WorkspaceSidebarSections.swift` `item(_:count:)` | Builds `PVSidebarNavItem(id:label:icon: section.icon, accessibilityIdentifier:count:)` |
| Sidebar component | `macos/App/DesignSystem/Components/SidebarNav/PVSidebarNav.swift` `PVSidebarNavItem`, `PVSidebarNavButton` | Item has `let icon: PVSymbol`. Row draws `PVIcon(item.icon, size: iconSize)`. `iconSize` is 15 expanded, 16 collapsed, and 14 for a child. The row's foreground is `textSecondary` at rest and `textPrimary` when selected (`PVSidebarNavRowBody`). The `#Preview` builds seven items with `icon:`. These are the only `PVSidebarNavItem(` constructors in the repo besides `WorkspaceSidebarSections`. |
| History jump rows | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `HistoryJumpRow.icon` (line 427), `HistoryJumpMenuPanel` (line 535) | `icon` is `location.section.icon`. The panel draws `PVIcon(item.icon, size: 15)`, then `.foregroundStyle(PVColor.textSecondary)` and `.frame(width: 15)`. This is the only place `WorkspaceToolbar` draws a section icon. Breadcrumbs are text. |
| Empty state | `macos/App/DesignSystem/Components/EmptyState/PVEmptyState.swift` `PVEmptyState` | Stores `private let icon: PVSymbol`. Draws `PVIcon(icon, size: compact ? 20 : 26)` in `PVColor.textFaint`. Two inits (`icon:title:` and `icon:verbatimTitle:`) plus two `EmptyView` conveniences, all symbol-only. `rg "PVEmptyState\(" macos/App | wc -l` is 15 at the planning base. Wrong for WS-4. |
| Conclude list page | `macos/App/Features/Conclusions/ConclusionListPage.swift` `ConclusionListKind.emptyIcon` (line 46), `ConclusionListBody` `.empty` case (line 144) | `PVEmptyState(icon: Kind.emptyIcon, …)`. `Kind.mark` (line 51) already exists and feeds the row thumbnails. |
| Conclude kinds | `macos/App/Features/Conclusions/PersonsListView.swift`, `EventsListView.swift`, `PlacesListView.swift` | Each declares `emptyIcon` (`.person` / `.calendar` / `.mapPin`) and `mark` (`.subjectPerson` / `.subjectEvent` / `.subjectPlace`). They are the only `ConclusionListKind` conformers. `mark` is a nonisolated `static var`. |
| Tests | `macos/ProvenenciaTests/WorkspaceSidebarSectionsTests.swift` `WorkspaceSidebarSectionsTests` (`@MainActor`) | Covers sections, order, identifiers, and counts. Nothing on icons or marks. The file is already in the `ProvenenciaTests` target. `ConclusionListKindTests` (`macos/ProvenenciaTests/PersonsListTests.swift`) does not read `emptyIcon`. |
| Docs | `macos/App/DesignSystem/Recipes/Marks/MARKS.md`, `macos/App/DesignSystem/README.md`, `docs/design-system-layers.md` | `MARKS.md` lists `subject_event` as "S6 Evidence graph board". Its `subject_*` table has 7 rows for 8 keys, with `subject_place_relationship` missing. The layers doc says lower layers must not import higher ones. `PVThumbnail` (a Component) already takes a `PVMarkKey` (a Recipe type), and no doc records that as allowed. |

**Precedent.** `6516bc3b` ("Replace place-relationship mark stand-in with designed glyph") replaced one mark SVG with the design-system export, verbatim, and changed nothing else. Copy that for STEP-1. For the glyph, three sites already choose between a mark and a symbol by hand: `PVThumbnail.Content` (`macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift`), `PropertiesView.typeCard` (`macos/App/Features/Properties/PropertiesView.swift`), and `OmnibarResultsHost` (`macos/App/Features/Workspace/OmnibarResultsHost.swift`). `PVGlyph` is the shared value they should have had. This PR does not migrate them (Out).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The new asset is the kit's export, copied byte for byte with its c2pa manifest. STEP-1 downloads the exact bytes and checks them against the size and SHA-256 the planner recorded. `Contents.json` is unchanged. The export has no `color="#27495B"` attribute and no `<title>`; that is expected, as with `subject_place_relationship` in `6516bc3b`. | Every mark in `Assets.xcassets/Marks` is a verbatim kit export with the manifest. The user chose this path (2026-10-14). | Hand-writing the five elements without the manifest, because it diverges from every sibling. Redrawing from the frame, because the board is the source of truth. |
| D-2 | WS-3's "toolbar section icon" is the Back and Forward history jump-menu rows (`HistoryJumpRow`). Nothing else in the toolbar changes. | `HistoryJumpRow.icon` is the only toolbar code that reads the section icon (`rg -n "section\.icon" macos`). Board Frame 14's note says the same. | Adding a section glyph to the breadcrumb or the toolbar, because that is new chrome no frame draws. |
| D-3 | Add `enum PVGlyph: Equatable { case symbol(PVSymbol); case mark(PVMarkKey) }` *(new)* and `struct PVGlyphView: View` *(new)*, `init(_ glyph: PVGlyph, size: CGFloat)`. Its body is `PVIcon(symbol, size: size)` for `.symbol` and `PVMark(key, size: size, decorative: true)` for `.mark`. Both go at the end of `macos/App/DesignSystem/Recipes/Marks/PVMark.swift`, after `PVFileTypeGlyph` and before the `#Preview`. Each gets a doc comment. `PVGlyph`'s comment says it is unrelated to `PVFileTypeGlyph`, which maps a file to a mark key. The view is always decorative, as `PVIcon` already is. A label beside it names the thing. | One value for one concept. The section, the sidebar row, the history row, and the empty state each hold one fact instead of an icon plus an optional mark. The user chose this over the optional-mark shape (2026-10-14, review M-3). The recipe layer is where `PVMarkKey` lives. The same file avoids `project.pbxproj` edits (the app target lists files by hand). | An optional `mark` beside a required `icon`: two facts per section, and dead fallback symbols. A private enum inside `PVEmptyState` alone: a second shape for the same idea in this PR. A new file: two small types do not justify hand-edited `pbxproj` entries. Migrating the three older sites now: Out, logged in [glyph consolidation](../../../ideas/glyph-consolidation.md). |
| D-4 | `WorkspaceSection.glyph: PVGlyph` *(new)* replaces `icon`. It maps `.sources` → `.symbol(.library)`, `.sourceTypes` → `.symbol(.tag)`, `.metadata` → `.symbol(.list)`, `.properties` → `.symbol(.listTree)`, `.persons` → `.mark(PersonsList.mark)`, `.events` → `.mark(EventsList.mark)`, `.places` → `.mark(PlacesList.mark)`. `icon` is deleted. | One switch per section. The Conclude cases read the list kind's mark, so the sidebar, the history row, the empty state, and the list thumbnails share one source and cannot drift (review M-1). | Repeating the mark keys in `WorkspaceSection` and keeping them equal with a test. Mapping through `PropertiesTypeChrome.stripMarkKey(typeKey:)`, which couples the workspace to Properties chrome by subject-type string. |
| D-5 | `PVSidebarNavItem.icon: PVSymbol` becomes `let glyph: PVGlyph` *(new)*, in the same position. `PVSidebarNavButton` draws `PVGlyphView(item.glyph, size: iconSize)`. `iconSize` (15 / 16 / 14) is unchanged. The accessibility label stays `item.label`. | Brief inventory: SidebarNav **Extend**. The label names the row (`MARKS.md` Accessibility). | Keeping `icon` and adding `mark`, which D-3 rejects. |
| D-6 | A mark takes the row's foreground exactly as `PVIcon` does. No extra opacity. | The board says the selected-row treatment is "identical for marks and symbols". Its 0.7 rest dim copies the React `PVIcon` light weight. The Swift `PVIcon` has no rest dim, and `PVMark` has no foreground of its own, so with no added dim the two match. | `.opacity(0.7)` on marks at rest, because it would make them fainter than the Swift symbols beside them. |
| D-7 | `HistoryJumpRow.icon` becomes `var glyph: PVGlyph { location.section.glyph }` *(new)*. `HistoryJumpMenuPanel` draws `PVGlyphView(item.glyph, size: 15)` with the existing `.foregroundStyle(PVColor.textSecondary)` and `.frame(width: 15)` chained on it. | Board Frame 14: rows show the same glyph as the sidebar row, at 15pt, in the menu's secondary ink. One view, so the modifiers chain directly. | Moving the choice into `PVContextMenuPanel`, which does not know about sections. |
| D-8 | `PVEmptyState` stores `private let glyph: PVGlyph`. The two existing inits and their two `EmptyView` conveniences keep their `icon: PVSymbol` signatures and store `.symbol(icon)`. Add `init(glyph: PVGlyph, title: LocalizedStringResource? = nil, message: String? = nil, compact: Bool = false, @ViewBuilder action: () -> Action)` *(new)* and the `EmptyView` convenience `init(glyph:title:message:compact:)` *(new)*. The body draws `PVGlyphView(glyph, size: compact ? 20 : 26).foregroundStyle(PVColor.textFaint)`. | Brief: Extend `PVEmptyState`, not a second empty state. Board Frame 16 uses the 26pt slot and the faint ink. Existing callers do not change. | Changing every caller to pass a `PVGlyph`: churn with no gain. A `verbatimTitle:` glyph init, which has no caller. |
| D-9 | `ConclusionListBody` calls `PVEmptyState(glyph: .mark(Kind.mark), title: Kind.emptyTitle, message: L10n.string(Kind.emptyMessage))`. `ConclusionListKind.emptyIcon` and the `emptyIcon` of `PersonsList`, `EventsList`, and `PlacesList` are deleted. The vocabulary panes' `emptyIcon` is a different field and stays. | `Kind.mark` is the key the rows use. That makes `emptyIcon` dead. | Keeping `emptyIcon` as a fallback that no caller takes. |
| D-10 | Delete `PVSymbol.person` and `PVSymbol.mapPin`. Keep `.calendar`. | After D-4, D-5, and D-9 they have no reader (§4 Symbol icon row). | Keeping unused cases "in case", which is dead code. |
| D-11 | Record the layering rule this PR relies on. A Component may take recipe-layer artwork (a `PVMarkKey` or a `PVGlyph`) in its icon slot, and draws it with `PVMark` or `PVGlyphView`. `PVThumbnail`, `PVSidebarNav`, and `PVEmptyState` do this. It goes in `docs/design-system-layers.md` and `MARKS.md` (STEP-7). | The layers doc forbids a lower layer importing a higher one. `PVThumbnail` already breaks that silently, and this PR adds two more cases. Writing the exception down keeps the next review from flagging it as debt (review M-2). | Moving `PVMark` down to Components: it carries `PVFileTypeGlyph` and product L10n names, so it would be a preparatory PR. |
| D-12 | Tests are Swift Testing cases added to `WorkspaceSidebarSectionsTests`, which is already in the target, so there is no `project.pbxproj` edit. The asset is checked with `grep`. Views are checked by hand. | The repo has no snapshot tests (`add-swift-test`), and SwiftUI view trees are not unit-tested here. | A new test file. Pixel tests, because there is no harness. |
| D-13 | No new user-facing strings. Glyphs are decorative here, and `PVMarkKey.accessibilityName` already exists. | Labels, titles, and messages are unchanged. | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVGlyph` *(new)* | enum | `PVThumbnail.Content` (`mark` + `icon`), `PropertiesView.typeCard`, `OmnibarResultsHost` | Those choose by hand at each site, with an optional mark beside an optional icon. A two-case enum holds one fact, and later sites can adopt it ([glyph consolidation](../../../ideas/glyph-consolidation.md)). Searched `rg "case symbol\|case mark\("`: no existing type. |
| `PVGlyphView` *(new)* | view | `PVIcon`, `PVMark` | It composes both and adds no drawing of its own. |
| `WorkspaceSection.glyph` *(new)* | property | `WorkspaceSection.icon` (deleted), `ConclusionListKind.mark` | It replaces `icon` and reads `Kind.mark` for the Conclude sections (D-4). |
| `PVSidebarNavItem.glyph` *(new)* | property | `PVSidebarNavItem.icon` (replaced) | It is the same slot with a wider type. |
| `PVEmptyState.init(glyph:…)` *(new)* | init | `PVEmptyState.init(icon:…)` | It extends the one empty state. The `icon:` inits store `.symbol(icon)` and keep their signatures. |
| `HistoryJumpRow.glyph` *(new)* | property | `HistoryJumpRow.icon` (replaced) | It is the same slot. |
| Tests `sidebarRowsCarryTheSectionGlyph`, `sectionGlyphsMatchTheBoard` *(new)* | test | `WorkspaceSidebarSectionsTests` | They are added to that suite. |

**Builds on.** The `PVMark` recipe and its asset catalog, `PVIcon`, and `ConclusionListKind.mark`. Design-system layering: `PVSidebarNav` and `PVEmptyState` are components extended in place (`add-ui-component`).
**Deletes.** The hourglass artwork. `WorkspaceSection.icon`. `PVSidebarNavItem.icon`. `HistoryJumpRow.icon`. `ConclusionListKind.emptyIcon` and `PersonsList.emptyIcon`, `EventsList.emptyIcon`, `PlacesList.emptyIcon`. `PVSymbol.person` and `PVSymbol.mapPin`.
**Next adds.** A Narrate sidebar row or a new Conclude kind sets one `WorkspaceSection.glyph` case and gets the sidebar, the jump menu, and its empty state. S10-22 (D9 to D11) and later list briefs reuse the empty state as it is. `PVThumbnail.Content`, `PropertiesView.typeCard`, and `OmnibarResultsHost` can each move to `PVGlyph` in one change ([glyph consolidation](../../../ideas/glyph-consolidation.md)).

### Schema

None — no catalog change.

### Go core

None — macOS only.

### FFI

None — no RPC change.

### macOS

Sketch of the new types (the implementer writes the real code and doc comments):

```swift
/// An SF Symbol or a curated mark, for an icon slot that takes either.
enum PVGlyph: Equatable {
    case symbol(PVSymbol)
    case mark(PVMarkKey)
}

/// Draws a `PVGlyph` at one size, tinted by the ambient foreground. Decorative.
struct PVGlyphView: View {
    init(_ glyph: PVGlyph, size: CGFloat)
    // .symbol → PVIcon(symbol, size: size); .mark → PVMark(key, size: size, decorative: true)
}
```

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Event mark | `PVMark` asset `subject_event` | `macos/App/Resources/Assets.xcassets/Marks/provenencia_subject_event.imageset/provenencia_subject_event.svg` | D-1. A calendar: a rounded 15×15 square, a band under the top, two rings above it, five dots. Stroke 1.5 on a 24 grid, the same as Person and Place. |
| Glyph | `PVGlyph`, `PVGlyphView` | `macos/App/DesignSystem/Recipes/Marks/PVMark.swift` | D-3 |
| Section glyph | `WorkspaceSection.glyph` | `macos/App/Features/Workspace/WorkspaceSection.swift` | D-4 |
| Sidebar | `PVSidebarNav` (Extend) | `macos/App/DesignSystem/Components/SidebarNav/PVSidebarNav.swift` | D-5, D-6. The `#Preview` items pass `glyph:`, with marks for persons, events, and places. |
| Sidebar rows | `WorkspaceSidebarSections.item` | `macos/App/Features/Workspace/WorkspaceSidebarSections.swift` | Passes `glyph: section.glyph` |
| History jump rows | `HistoryJumpRow`, `HistoryJumpMenuPanel` | `macos/App/Features/Workspace/WorkspaceToolbar.swift` | D-2, D-7 |
| Empty state | `PVEmptyState` (Extend) | `macos/App/DesignSystem/Components/EmptyState/PVEmptyState.swift` | D-8. The `#Preview` adds `PVEmptyState(glyph: .mark(.subjectEvent), title: "No events yet", compact: true)`. |
| Conclude empty | `ConclusionListBody` | `macos/App/Features/Conclusions/ConclusionListPage.swift` | D-9 |

The agreed frames, in words:

- **S10-D16, the family.** Person, Event, and Place at one size and one stroke. Event is a calendar page: a rounded square, a horizontal band a third of the way down, two short rings rising above the top edge, and two rows of dots (three, then two) below the band. It is still readable as a calendar at 15–16pt, and at 24pt and 40pt. The ink comes from the caller: the same glyph in Event copper and in a neutral ink.
- **Frame 12, sidebar expanded, Events selected.** The Source group has Sources with the library symbol. The Conclude group has Persons (person mark), Events (calendar mark, selected-row fill and semibold label), and Places (place mark), each with its count. The Configure group sits at the bottom, with Source types, Metadata, and Properties on their symbols and no counts. Marks sit in the 15pt icon slot.
- **Frame 13, sidebar collapsed.** A rail of 32pt buttons with icons only. The three Conclude buttons show the marks at 16pt, and the other buttons keep their symbols. Each button's tooltip and accessibility label is its row label.
- **Frame 14, toolbar on Events, Back menu open.** Each row is a 15pt glyph column, then *Section › title* crumbs, then a mono ref. Rows for Events, Places, and Persons show their marks. Rows for Sources and the Configure sections show their symbols. All glyphs use the secondary ink.
- **Frame 15, toolbar on Sources, Back menu open.** The same panel. Marks and symbols sit side by side at one weight.
- **Frame 16, empty lists.** Three empty states, each the existing dashed sunken box. The icon slot holds the kind's mark at 26pt in faint ink, over the existing title and message.

States: there are no new states. Selection, hover, and disabled are unchanged on the sidebar and the menu rows. Empty, loading, rows, and failed are unchanged on the lists; only the empty glyph changes. Accessibility: `PVGlyphView` is decorative. VoiceOver still reads "Events" on the sidebar row and the crumb text on a jump row.

### Strings

None — no copy changes (D-13).

## 7. Steps

**STEP-1 — Replace the Event artwork**
Files: `macos/App/Resources/Assets.xcassets/Marks/provenencia_subject_event.imageset/provenencia_subject_event.svg`. Skill: `add-ui-component`.
Do:
1. Call the Claude Design `render_preview` tool with project 0f6c1f29-8408-4735-b114-5ae4c4c18445 and path swift/Assets.xcassets/Marks/provenencia_subject_event.imageset/provenencia_subject_event.svg. It returns a short-lived `serve_url`. That URL carries a token, so never write it into a file, a commit, a log, or the PR.
2. Run `curl -sS -o macos/App/Resources/Assets.xcassets/Marks/provenencia_subject_event.imageset/provenencia_subject_event.svg "<serve_url>"`. This writes the export's exact bytes over the repo file. Do not open or edit the result.
3. Leave `Contents.json` alone. Change no other file under `Assets.xcassets/Marks`.

The planner fetched the same export the same way on 2026-10-14. The checks below are its size and hash.
Check: with `F` set to the repo file, each line below holds.
- `wc -c < "$F"` prints `8218`.
- `shasum -a 256 "$F"` prints `477f7565973129798a1c9be1e3b00256ceb85d62320f3462cc178b8cb167b2b7`.
- `grep -c 'rect x="4.5" y="5.5" width="15" height="15" rx="2"' "$F"` prints `1`. `grep -o 'h\.01' "$F" | wc -l` prints `5`. `grep -c 'M8.2 3.6v2.9' "$F"` prints `0`.
- `git status --short macos/App/Resources/Assets.xcassets` lists only this file.

A different hash means the board's asset changed after planning. That is a stop condition.

Commit: "Draw the Event mark as a calendar instead of an hourglass."

**STEP-2 — Add the glyph value**
Files: `macos/App/DesignSystem/Recipes/Marks/PVMark.swift`. Skill: `add-ui-component`.
Do: add `PVGlyph` and `PVGlyphView` per D-3 and the §6 sketch, after `PVFileTypeGlyph` and before `#Preview("Marks")`.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO` succeeds.
No commit of its own; it is committed with STEP-3.

**STEP-3 — Section glyphs in the sidebar and the history menus**
Files: `macos/App/Features/Workspace/WorkspaceSection.swift`, `macos/App/DesignSystem/Components/SidebarNav/PVSidebarNav.swift`, `macos/App/Features/Workspace/WorkspaceSidebarSections.swift`, `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/ProvenenciaTests/WorkspaceSidebarSectionsTests.swift`. Skills: `add-ui-component`, `add-swift-test`.
Do:
- D-4: replace `var icon: PVSymbol` with `var glyph: PVGlyph`.
- D-5: rename `PVSidebarNavItem.icon` to `glyph: PVGlyph`, and draw `PVGlyphView(item.glyph, size: iconSize)` in `PVSidebarNavButton`. In the `#Preview`, pass `.symbol(.library)`, `.symbol(.tag)`, `.symbol(.list)`, `.symbol(.listTree)`, `.mark(.subjectPerson)`, `.mark(.subjectEvent)`, `.mark(.subjectPlace)`.
- `WorkspaceSidebarSections.item` passes `glyph: section.glyph`.
- D-7: `HistoryJumpRow.glyph` and `PVGlyphView` in `HistoryJumpMenuPanel`.
- Add T-1 and T-2.

These change together because the renamed members do not compile alone.
Check: `rg -n "section\.icon|item\.icon" macos/App/Features/Workspace macos/App/DesignSystem/Components/SidebarNav` prints nothing. `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSidebarSectionsTests CODE_SIGNING_ALLOWED=NO` passes.
Commit (with STEP-2): "Show the subject marks for Persons, Events, and Places in the sidebar and the history menus."

**STEP-4 — Empty states take a glyph**
Files: `macos/App/DesignSystem/Components/EmptyState/PVEmptyState.swift`, `macos/App/Features/Conclusions/ConclusionListPage.swift`, `macos/App/Features/Conclusions/PersonsListView.swift`, `macos/App/Features/Conclusions/EventsListView.swift`, `macos/App/Features/Conclusions/PlacesListView.swift`. Skill: `add-ui-component`.
Do:
- D-8: replace `private let icon: PVSymbol` with `private let glyph: PVGlyph`. Add the `glyph:` init and its `EmptyView` convenience. Have the two `icon:` inits and their conveniences store `.symbol(icon)`. Draw with `PVGlyphView(glyph, size: compact ? 20 : 26).foregroundStyle(PVColor.textFaint)`. Change the header comment to say the icon slot takes a symbol or a mark. Add the preview case from §6.
- D-9: switch `ConclusionListBody`'s `.empty` case to `PVEmptyState(glyph: .mark(Kind.mark), …)`. Remove `static var emptyIcon: PVSymbol { get }` from `ConclusionListKind` and the three implementations.

Check:
- `rg "PVEmptyState\(" macos/App | wc -l` prints `16`: the 15 at the planning base plus the new preview line.
- The step's own changes (`git diff --stat` against the STEP-3 commit) touch only the five files above.
- `rg -n "emptyIcon" macos/App/Features/Conclusions macos/ProvenenciaTests` prints nothing.
- The STEP-3 test command passes.

Commit: "Show each kind's mark on the empty Persons, Events, and Places lists."

**STEP-5 — Delete the dead symbols**
Files: `macos/App/DesignSystem/Components/Icon/PVIcon.swift`.
Do: D-10. Delete `case person = "person"` and `case mapPin = "mappin"` from `PVSymbol`.
Check: the STEP-2 build command succeeds. `rg -n "\.mapPin\b" macos` prints nothing.
Commit: "Remove the person and map-pin symbols the shell no longer draws."

**STEP-6 — Full gate**
Files: none.
Do: run the repo gate.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Design-system docs**
Files: `macos/App/DesignSystem/Recipes/Marks/MARKS.md`, `macos/App/DesignSystem/README.md`, `docs/design-system-layers.md`.
Do:
- `MARKS.md`:
  - Heading `### subject_* (7)` becomes `(8)`. Add the missing row `| subject_place_relationship | bridge | S9-38 Place model |` after `subject_location`.
  - The `subject_event` row keeps its first two columns. Its provenance becomes "S10-D16 design-system board (calendar; replaced the S6 hourglass)".
  - In **Size**, add the row `| raw 15–16 / 20–26 | — | Shell chrome: sidebar row 15 (rail 16), history menu 15, empty state 26 (compact 20) |`.
  - Under **Tint / color**, add: "Beside SF Symbols in one column (sidebar, menus), a mark takes the same foreground as `PVIcon`, with no extra dim."
  - Add a section **Glyph** before **Board links**: "`PVGlyph` names either a `PVSymbol` or a `PVMarkKey`; `PVGlyphView` draws it, decorative. A component whose icon slot can hold either takes a `PVGlyph` (`PVSidebarNav`, `PVEmptyState`)."
  - Under **Board links**, add "Design system (subject_event, S10-D16): https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445".
- DesignSystem `README.md` (each append goes after the row's closing parenthesis):
  - On the SidebarNav row, append "; an item's icon is a `PVGlyph`, a symbol or a mark (S10-19)".
  - On the EmptyState row, append "; the icon slot takes a `PVSymbol` or a `PVGlyph` (S10-19)".
  - On the Marks row, append "; `PVGlyph` / `PVGlyphView` for a symbol-or-mark slot".
- `docs/design-system-layers.md`: after the paragraph that says lower layers must not import higher ones, add D-11's rule as one paragraph. Say that recipe-layer artwork (`PVMarkKey`, `PVGlyph`) may fill a component's icon slot, and name `PVThumbnail`, `PVSidebarNav`, and `PVEmptyState`.

Check: `rg -n "PVGlyph" macos/App/DesignSystem/Recipes/Marks/MARKS.md macos/App/DesignSystem/README.md docs/design-system-layers.md` finds each file.
Commit: "Document the glyph slot and the new Event mark in the design-system notes."

**STEP-8 — Plan bookkeeping**
Do this after the draft PR is open, so its number is known. Use the date the PR is opened as the landing date; the user corrects it at merge if it differs.
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/design/README.md`, `docs/deployment-plan/spike-10/design/S10-D16-event-mark.md`, `docs/deployment-plan/spike-10/design/S10-D7-sidebar-marks.md`, this plan.
Do:
- `git mv` both briefs into `docs/deployment-plan/spike-10/design/archive/`. In each moved brief, add one `../` to every relative link that leaves `design/` (`../deployment-plan.md` becomes `../../deployment-plan.md`, and so on). The link between the two briefs stays as it is. Under each title, add the line `archive/S10-D5-source-artifacts.md` has: "**Landed YYYY-MM-DD** in [#N](…). The sections below are the original board prompt." Leave the rest of each brief unchanged.
- Map, Design track table: link both rows to `design/archive/…`.
- Map, `#### S10-19`: copy the `#### S10-18` shape. Insert the landed line under the heading ("**Landed YYYY-MM-DD** in [#N](…). See" plus the `completed.md` link, as S10-18 has it). Point the **Briefs:** line at `design/archive/…`. Delete the **Plan:** line. Keep the In / Out / Testable table.
- Map, PR sequence: append " — done, #N" to the line `✎ S10-D7 ──▶ S10-19  …`. Keep the `✎ S10-D16 ──▶` line above it.
- Map, working rule "One Event mark": replace the whole bullet with: "**One Event mark.** The Event mark is the S10-D16 calendar. Cards, lists, omnibar hits, Promote, Properties, the sidebar, the history menus, and the three list empty states draw `subject_person`, `subject_event`, and `subject_place` through `PVMark`. The shell reads `WorkspaceSection.glyph`. Landed in S10-19."
- `design/README.md`: delete the S10-D16 and S10-D7 rows from Scheduled. Add them to Done in the S10-D5 row's form. The Notes column reads "Done YYYY-MM-DD." followed by a link to `archive/<file>`.
- `completed.md`: add a `## S10-19 — Event mark and sidebar icons` heading at the top, then one **Landed** paragraph in the voice of the S10-18 entry.
- This plan: Status becomes `Landed in [#N](…) (date)`. Add a **Landed** note that lists every difference from this plan. The **Brief:** line links `../design/archive/…`.

Check:
- `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.md` reports no errors.
- `rg -n "design/S10-D16-event-mark|design/S10-D7-sidebar-marks" docs --glob '!docs/deployment-plan/spike-10/prs/**'` prints only `design/archive/` paths.

Commit: "Record S10-19 as landed in pull request N."

Commits: one per behavior, as listed in each step. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/WorkspaceSidebarSectionsTests.swift` `sidebarRowsCarryTheSectionGlyph` *(new)* | `WorkspaceSidebarSections.make(counts:)` → item `persons` has glyph `.mark(.subjectPerson)`, `events` has `.mark(.subjectEvent)`, `places` has `.mark(.subjectPlace)`. `sources` has `.symbol(.library)`, `source-types` has `.symbol(.tag)`, `metadata` has `.symbol(.list)`, `properties` has `.symbol(.listTree)`. | AC-2.1, D-4, D-5 |
| T-2 | Swift | same file, `sectionGlyphsMatchTheBoard` *(new)* | For every `WorkspaceSection.allCases`, `glyph` equals the D-4 table, and the three Conclude cases equal `.mark(PersonsList.mark)`, `.mark(EventsList.mark)`, `.mark(PlacesList.mark)` | AC-3.1, AC-3.2, AC-4.3, D-4, D-7 |
| T-3 | Shell | STEP-1 Check (size, SHA-256, and `grep` on the asset) | 8218 bytes, hash `477f7565…b7b2b7`, calendar `rect` once, five dots, no hourglass path | AC-1.1, D-1 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break are `WorkspaceSidebarSectionsTests`, `ConclusionListKindTests`, `WorkspaceNavigationTests`, and `WorkspaceToolbarBreadcrumbTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected: D-13)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (PVGlyph is the user's 2026-10-14 call, D-3)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none (none)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-19 entry, PR sequence line, and the "One Event mark" working rule; completed.md: one short paragraph of what shipped
- [ ] S10-D16 and S10-D7 moved to design/archive/ with their relative links fixed; design/README.md moves both to Done
- [ ] MARKS.md, DesignSystem/README.md, and docs/design-system-layers.md updated per STEP-7
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The implementer runs these if it can drive the macOS app. If it cannot, it lists each line in the PR body as "not run — for the reviewer", and the user runs them at review. Build and run the scheme `Provenencia` in Xcode. Use two projects: the researcher's dogfood project, which has a Source whose Evidence graph holds an event card plus at least one promoted Event and Person, and a new project created through onboarding.

1. **AC-1.2.** Open that Source's Evidence graph at 100%. The event card's kind mark is a calendar in Event copper, not an hourglass.
2. **AC-1.3.** Press ⌘K and type the promoted Event's title. The Event hit's lead mark is the calendar.
3. **AC-2.1, AC-2.2.** Click Events in the expanded sidebar. Persons, Events, and Places show the person, calendar, and pin marks. Sources shows the library symbol. Source types, Metadata, and Properties keep their symbols. Counts and labels are unchanged. The selected Events mark is as dark as its label, and the unselected marks match the unselected symbols' ink.
4. **AC-2.3.** Collapse the sidebar. The three marks are readable in the rail. Hovering Events shows "Events". With VoiceOver on, the Events button reads "Events".
5. **AC-3.1, AC-3.2.** Click Sources, then Events, then open a Person. Long-press Back. The Events row shows the calendar, and the Sources row shows the library symbol. Press Back twice, then long-press Forward. The Events and Persons rows show their marks, and the Person row reads *Persons › {name}* with its ref.
6. **AC-4.1, AC-4.2, AC-4.3.** In the new project, open Events. The dashed empty box shows the calendar mark over "No events yet" and the existing message, the same mark as the sidebar row. Open Persons and Places: the person and pin marks over "No persons yet" and "No places yet".
7. Slice 5 Check: "Sidebar Event matches the card." Steps 1 and 3 show the same calendar.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Xcode's SVG importer drops part of the new drawing (the dots are `h.01` round-capped lines) | STEP-1's check covers the file. Manual checks 1 and 3, at 15–16pt, confirm the dots render. |
| A mark looks heavier or lighter than the SF Symbols beside it | D-6 keeps the same ink. The board accepted the weights side by side (Frame 15). |
| Renaming `icon` to `glyph` misses a reader | STEP-3's `rg` check and the build. §4 lists every reader at the planning base. |

**Stop and report if:**

- The Claude Design `render_preview` tool is not available to you, or the downloaded file fails a STEP-1 check (including the hash). Do not draw or retype the SVG. The user can export the file into the repo to unblock you.
- After STEP-1, the built app shows no Event mark anywhere (a blank glyph on the card or in the sidebar). That means the asset failed to compile.
- `rg -n "section\.icon" macos` at the start finds a reader other than `WorkspaceSidebarSections.swift:54` and `WorkspaceToolbar.swift:427`. D-2 assumes those are the only two.
- `rg -n "ConclusionListKind" macos` finds a conformer other than `PersonsList`, `EventsList`, and `PlacesList`.
- The STEP-5 build fails because a file outside §4 still reads `PVSymbol.person` or `PVSymbol.mapPin`.
- A mark draws in its own color rather than the row's ink, which would contradict D-6.
- Any frame described in §6 disagrees with an AC in §2.

## 12. Open questions

None.

## 13. PR

**Title:** S10-19 — Event mark and sidebar icons
**Body:** objective in one sentence; what shipped against US-1 to US-4; AC evidence (T-1 to T-3 and the manual check lines, each marked run or "not run — for the reviewer"); deviations from the plan; `Plan: docs/deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.md`.
