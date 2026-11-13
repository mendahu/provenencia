# S15-35 — The Artifact list shows each Artifact's copy details

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 2 — Metadata at three levels
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D14](../design/S15-D14-source-page-artifacts.md) SA-1 – SA-3, SA-5 (without Move), SA-10, SA-11, SA-W1, board agreed 2026-11-13; plus the user's 2026-11-13 additions on the same board (the header's type pill as the kit Tag, the meta row that never clips, the description at caption size)
**Board:** [Claude Design — Source Detail](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK), page *Source page*: row "Artifacts" (`art-list`, `art-list-wide`, `art-sixty`, `art-menu`, `art-empty`) and row "The page" (`page`, `page-wide`, `header-wrap`); decisions panel entry "S15-D14". The implementer reads it with the Artifact tool's `read` on that link, `path` "project/Source.dc.html" (the list at `id="artifacts"`, its `artVals()` with `rowStyle`; the header's `PVTag`; the resting description `descResting`) and "project/_ds/provenencia-design-system-0f6c1f29-8408-4735-b114-5ae4c4c18445/_ds_bundle.js" (`components/PVTag.jsx`). § 6 macOS describes each frame in words, so the plan stands without the board
**Depends on:** S15-34 (landed [#819](https://github.com/mendahu/provenencia/pull/819)), S15-D14 (board agreed 2026-11-13)
**Blocks:** S15-25, S15-40
**Branch:** `s15-35-artifact-list-rows`
**Version bump:** minor, 0.12.0 → **0.13.0** (or the next minor after `main`'s `VERSION` when you rebase): `VERSION`, `core/version.go`, and every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` — a backward-compatible feature on a screen (the Artifacts section becomes a list of rows with a menu; a new kit component); no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (STEP-1's `PVTag`, STEP-2's `PVFlowLayout`; the row stays a Source page snowflake), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "An Artifact is added and edited in one sheet" (a square thumbnail of at most 80pt; no derivative size change) and "One brief, one PR" (S15-D14 predates it and is exempt) ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); the map's S15-35 scope entry; [S15-D14](../design/S15-D14-source-page-artifacts.md) § 2, § 4, § 4.1, § 6 and its **Board** line; S15-34's [plan](S15-34-artifact-edit-sheet.md) D-16 and **Next adds** ("S15-35: the list redraws each entry, places `ArtifactCopyDetailsView` and calls `openSheet(…, at:)`") and its **Landed** note; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide)
**Planned:** 2026-11-13 against `main` @ `64a91ae`
**Review:** — (not reviewed yet)
**Runs on:** macOS. Every Swift check and `scripts/check-pr.sh` need Xcode ([`README.md`](../../../../README.md)). No proto or catalog change: never run `./scripts/generate-proto.sh` in this PR

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

On a Source page, the Artifacts section reads as a list: one row per Artifact with a sharp 64pt square thumbnail, its label, ref, a two-line description, the graphs it is in, and its saved copy details in a column of their own, then its actions; a right-click on a row offers Edit…, Set cover and Delete…. At 60 Artifacts the section stays scannable, and at Wide the copy-details column gets more room. The page header's type shows as the kit's Tag (a rounded pill with a colour dot), on the Person page's Sources rows too, and the header's meta row never draws past its edge. The Source's description reads at caption size, as the board draws it.

**Why now.** S15-34 moved editing into the Artifact sheet and put read-only copy details on today's tall cards; the board agreed on 2026-11-13 replaces the cards with rows (SA-1, SA-2). The user scoped three header fixes the board also draws into this PR (map scope entry, 2026-11-13). S15-25 (Move on the row menu) and S15-40 (a missing-field mark on the row) build on the row.

**What stays.** The section's header (title, count, Add artifact), its empty state, its order (the workspace's), the Add artifact and Add file dialogs, the Artifact sheet and every way into it, the file open and Add file on the thumbnail, Set cover and the Cover badge, Delete and its DeleteImpact confirm, the page file drop. `PVBadge` stays exactly as it is. The rest of the Source page is unchanged.

## 2. User stories

**US-1.** As a researcher, I scan a Source's Artifacts as a list of rows and see each one's copy details without opening it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Census Source, 12 Artifacts, Regular (pane 1102). Look at the section | One row per Artifact in workspace order, separated by hairlines, with a hairline above the first. Each row, left to right: a 64pt square thumbnail; label (semibold) with ref (mono, muted) and, on the cover, the "Cover" badge; the description (at most two lines); the graph line; then the copy-details column; then the actions. A row's height follows its content; nothing is editable in place; the row has no hover |
| AC-1.2 | `ART-7QK2M` has Provider "Ancestry", URL, Accessed; "Sheet 17" has eight values; "Sheet 22" has none | `ART-7QK2M`: three lines, field name (96pt column, muted) then the mono value, the URL a link. Sheet 17: four lines and "+4 more". Sheet 22: an empty column of the same width, so the next row's details start at the same x |
| AC-1.3 | An Artifact in Robins and Hale; one in Robins, Hale and Ostrander; one in none (graphs loaded) | "Hale household, dwelling 113 · Robins household, dwelling 112" after a branch glyph (graphs in the Graphs section's order, by name); the second adds "+1"; the third shows the "Not in an evidence graph" badge and "Link to a graph". While the graph list is not loaded, neither line shows |
| AC-1.4 | An Artifact with no description | No description line (the "No description" placeholder is gone) |
| AC-1.5 | A JPEG Artifact | The thumbnail is the engine's 256px thumbnail cropped to fill 64pt, sharp at 2x. A PDF shows its file mark, an Artifact with no file the Source type's mark, in the kit tile |
| AC-1.6 | Click a thumbnail; click one with no file | Opens the file; opens Add file (as today) |
| AC-1.7 | 60 Artifacts, scrolled to the middle | Every row keeps its layout; no row grows with a form |
| AC-1.8 | No Artifacts | Today's empty state with Add artifact (SA-10) |

**US-2.** As a researcher, a row's actions are where I expect them, on the row and in its right-click menu.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | An image Artifact that is not the cover | The actions read Set cover, pencil, trash. On the cover, or on a PDF or fileless Artifact, Set cover is hidden but keeps its space, so every row's copy-details column lines up |
| AC-2.2 | Right-click a non-cover image row | A menu titled with the ref: "Edit…", "Set cover", "Delete…". Edit… opens the Artifact sheet; Set cover sets the cover (today's toast); Delete… opens today's DeleteImpact confirm |
| AC-2.3 | Right-click the cover row | "Edit…", "Delete…" (no Set cover) |
| AC-2.4 | "+4 more"; "Link to a graph"; the pencil | Open the sheet at Copy details; at Evidence graphs; at the top (as S15-34) |

**US-3.** As a researcher, the row column for copy details is wider at Wide.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Regular | The copy-details column is 320pt |
| AC-3.2 | Wide (pane 2296) | The rows stay in the main column beside the Metadata side column; the copy-details column is 440pt; a long URL shows more of itself |
| AC-3.3 | A pane too narrow for 320pt beside the label | The label column and the copy-details column both shrink; nothing overlaps the actions |

**US-4.** As a researcher, VoiceOver reads a row as one group.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | VoiceOver on `ART-7QK2M`'s row | The group is named "Sheet 14 (image 1)"; inside it reads the label with "ART-7QK2M", the description, the graphs ("Evidence graphs: Hale household, dwelling 113, Robins household, dwelling 112"), then "Provider, Ancestry" and each detail; the pencil reads "Edit ART-7QK2M"; the menu's items are in the actions rotor |

**US-5.** As a researcher, the Source page header shows the type as a Tag and never cuts anything off, and the description reads quieter.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | The header of a Census Source | The type is a rounded pill: raised surface, subtle border, a 7pt plum dot, "Census" in secondary caption text, no type icon. The Cover and "Not in an evidence graph" badges look as before |
| AC-5.2 | The Person page's Sources section | Each row's type is the same Tag (user, 2026-11-13) |
| AC-5.3 | Regular, a type "Church register (parish copy)" and a range date "between 1 Mar 1880 and 30 Jun 1880", then with the type picker open | The meta row wraps after the type; the date group sits on the next line; no item draws past the row's right edge; the date truncates with its full text as the tooltip only when its line alone is too narrow |
| AC-5.4 | The Source's description | Caption size, muted, normal line height, at most `PVSpacing.measureProse` wide; editing it is unchanged |

## 3. Scope

| In | Covered by |
| --- | --- |
| Kit `PVTag` ported from `PVTag.jsx`; `CatalogSourceTypePill` composes it with the plum dot and no icon, on the header and the Person page (user, 2026-11-13) | STEP-1, AC-5.1, AC-5.2 |
| `PVFlowLayout` never places an item past the row's edge; the header's ref, Tag and Private switch keep their width | STEP-2, AC-5.3 |
| The resting Source description at caption size (user, 2026-11-13) | STEP-3, AC-5.4 |
| The row's pure rules: the graph line, the copy-details column width, the menu items (SA-3, SA-5, SA-W1) | STEP-4, AC-1.3, AC-2.2, AC-2.3, AC-3.1, AC-3.2 |
| The list of rows replaces the card grid; 64pt thumbnail; right-click menu; read-only copy details in a fixed column (SA-1 – SA-5 without Move, SA-10, SA-11) | STEP-5, AC-1.1 – AC-1.8, AC-2.1, AC-2.4, AC-3.3, AC-4.1 |
| Strings, version, docs | STEP-6 – STEP-8 |

| Out | Home |
| --- | --- |
| Move to another Source… on the row menu | S15-25 |
| A mark on a row with a needed field empty | S15-40 |
| Add mode in the sheet; the Add artifact dialog removed | S15-10 |
| `PVTag`'s remove button and interactive hover (the web kit's `onRemove`, `interactive`): no caller | not planned: add with the first caller (`add-ui-component`) |
| A larger thumbnail derivative or a thumbnail that keeps the image's shape | not planned (brief § 6 Non-goals; working rule) |
| A Swift test of the row's SwiftUI layout | not planned: `swift-tests.mdc` tests models and pure rules; the manual check covers layout |
| Per-type colours for the Tag's dot | not planned: `CatalogSourceTypePill`'s comment already parks it ("per-type categorical coloring can land when type keys are wired to record tones") |

## 4. Current state

Verified at `64a91ae`.

| Area | Where | Today |
| --- | --- | --- |
| Section | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `SourcePageArtifactsView` (`cardMinWidth` 264, `cardColumns`, `body`'s `LazyVGrid`, `artifactCard`, `previewWell`, `metaRow`, `restingText`, `noGraphRow`, `descriptionLine`, `activatePreview`, `previewLabel`, `sourceTypeIconKey`, `focusedPreviewID`; `addForm`, `addFileForm`) | A grid of tall `PVCard`s with `pvLiftHover`: a square preview well, a meta row (ref, Cover badge, Set cover, pencil, trash), the label, the description at caption size with a 3-line minimum height and a "No description" placeholder, the copy details, the no-graph row. Graph names are not shown. No right-click menu. The card half goes; `addForm` and `addFileForm` stay (S15-10 owns the first) |
| Square well | same file, `private struct ArtifactSquarePreview` | Only caller is `previewWell`. Goes |
| Thumbnail | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail` (`enum Layout { tile, well }`, `layout`, `wellBody`, `wellGlyph`) | `.tile` draws `PVThumbnail` at `size`; `.well` fills a card. `SourcePageArtifactsView.previewWell` is the only `.well` caller; no caller passes `layout: .tile` |
| Kit tile | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` | Image `scaledToFill` and clipped to `size`; a mark at `size * 0.82`; `PVRadius.sm` corners, a 1pt line |
| Copy details | `macos/App/Features/Sources/ArtifactCopyDetailsView.swift` `ArtifactCopyDetailsView` (`shownCount`, `shown(_:)`, `body`: a `Grid` with `horizontalSpacing: PVSpacing.space4`, the field name at its own width) | S15-34. The label column is as wide as its widest label |
| Section model | `macos/App/Features/Sources/SourceArtifactsSection.swift` `items`, `visibleGraphs` (`context.evidenceGraphs`, nil until loaded), `inNoGraph`, `copyDetails(for:)`, `openSheet(artifactID:at:)`, `isCover`, `canUseAsThumbnail`, `useAsThumbnail`, `open`, `openAttach(to:)` | Everything the row needs except the Artifact's graph names |
| Graphs order | `macos/App/Features/Sources/SourceGraphsSection.swift` `rows` ("in list order (sorted by name)") | The same array as `visibleGraphs` |
| Delete | `macos/App/Features/Sources/SourcePageModel.swift` `askDeleteArtifact(id:)` | Today's trash action |
| Menu kit | `macos/App/DesignSystem/Components/ContextMenu/PVActionMenu.swift` `.pvRightClickMenu(title:accessibilityIdentifier:items:)`, `PVActionMenuItem` | Used by `SourcesListRow`, `SourcePageIdentityHeader` (cover), `ConclusionDetailPage` |
| Header | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `metaRow` (`PVFlowLayout` of `refAndType` and `dateAndPrivate`), `refAndType` (`CatalogSourceTypePill(label:iconKey:)`), `sourceDate` (the date text has `lineLimit(1)`, `.truncationMode(.tail)`, `maxWidth: PVSpacing.widthInspector`, `.help(line)`), `privateSwitch` | The type pill is cut off (user, 2026-11-13) |
| Flow layout | `macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift` `PVFlowLayout` (`rows(in:subviews:)` measures each item at `.unspecified` and places it at that size) | An item wider than the row is placed at its ideal width and draws past the row's right edge. Callers: the header, `CitationComposerAuthoringBox.swift` (lines 710 and 897), which cap their children |
| Type pill | `macos/App/Features/Catalog/SourceTypeMarkers.swift` `CatalogSourceTypePill(label:iconKey:)` | A type mark plus caption text in `PVColor.recordMarriage` on a `recordMarriageSoft` rounded rect (`PVRadius.sm`). Callers: the header and `macos/App/Features/Conclusions/PersonSourcesSection.swift` `typeCell` (in a 176pt column, `PersonSourcesColumns.typeWidth`) |
| Person rows | `macos/App/Features/Conclusions/PersonSourcesContent.swift` `Row.typeIconKey` (set from `sourceTypes.iconKey(forTypeID:)`) | Read only by the pill |
| Tag | `macos/App/DesignSystem/README.md` "no files yet" table: "Tag · Core · Removable/interactive pill with a color dot" | Not ported. Web spec (`components/PVTag.jsx` in the board's bundle): inline row, gap 6; padding 3 top/bottom, 8 leading, 9 trailing; `--radius-pill`; `--surface-raised`; 1px `--border-subtle`; `--text-secondary`; `--font-body` at `--text-caption`; a 7×7 circle in `color` (default `--paper-500`); optional `onRemove` (12pt x in `--text-faint`) and `interactive` (shadow-sm on hover) |
| Description | `macos/App/Features/Sources/SourcePageDescriptionView.swift` resting `Text(model.identity.description)` | Body size, `textSecondary`, `lineSpacing` from `PVLineHeight.normal` at body, full width |
| Tokens | `PVRadius.pill`, `PVColor.surfaceRaised`, `PVColor.borderSubtle`, `PVColor.textSecondary`, `PVColor.recordMarriage`, `PVColor.textMuted`, `PVSpacing.measureProse` (480), `PVSymbol.gitBranch` | All exist |
| Width class | `@Environment(\.pvWidthClass)` (`SourcePageView.swift` `SourcePageOverview`) | The page reads it; Artifacts sit in the main column at both widths (`SourcePageOverviewLayout`) |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Sources.artifactNoDescription`, `coverBadge`, `useAsThumbnail` ("Set cover"), `editArtifactLabel(ref:)` ("Edit %@"), `deleteArtifact`, `artifactNotInGraph`, `artifactLinkToGraph`, `artifactLinkToGraphLabel(ref:)`, `artifactOpenFileTip`, `artifactAddFileTip`, `openArtifactFile(name:)`, `addFileToArtifact(ref:)`; `L10n.DesignSystem.badgeMoreCount(_:)` ("+%lld") | No "Edit…" or "Delete…" menu titles for an Artifact |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` `copyDetailsShowFourThenMore`, `restingCardMarksAnArtifactInNoGraph` (helpers `makeCensus`, `graph`, `makeStore`, `makeModel`); `macos/ProvenenciaTests/PersonSourcesContentTests.swift` (asserts `typeIconKey` at lines 74 and 90); `macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift` (the pure `arrange` precedent) | Kept, updated as § 8 lists |

**Precedent.** `SourcesListRow` (`macos/App/Features/Sources/SourcesListRow.swift`): a list row with a `CachedThumbnail` tile at a fixed size and `.pvRightClickMenu` with `PVActionMenuItem`s; copy its menu wiring. `SourcePageOverviewLayout.arrange(width:widthClass:fit:)` with `SourcePageOverviewLayoutTests`: a `Layout` whose placement is a static pure function over a `fit` closure, tested without SwiftUI; copy that shape for `PVFlowLayout` (D-3). `ArtifactCopyDetailsView.shown(_:)` and its test: a row's pure rule as a `static func` on the view; copy it for D-6. `PVBadge.swift` is the shape for a kit pill (header comment naming the `.jsx`, tokens only, `#Preview`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **`PVTag`** *(new)*, `macos/App/DesignSystem/Components/Tag/PVTag.swift`, a Component. `init(text: String, dot: Color = PVColor.textMuted)` (data text; `textMuted` is `paper500` in light, the web default). Body: `HStack(spacing: 6)` of a 7×7 `Circle().fill(dot)` and `Text(text)` in `PVFont.body(size: PVTypeScale.caption)`, `PVColor.textSecondary`, `lineLimit(1)`; padding top/bottom 3, leading 8, trailing 9; background `PVColor.surfaceRaised` in a `Capsule()`; `Capsule().strokeBorder(PVColor.borderSubtle, lineWidth: 1)`; `.fixedSize()`; `.accessibilityElement(children: .combine)`. The header comment names `components/PVTag.jsx` and the deliberate deviation: no `onRemove` and no `interactive` until a caller needs them. `#Preview` with three tags. The README's built table gains a Tag row and the "no files yet" table loses it | Map scope: port the Tag, keep `PVBadge`; a content-agnostic pill is a Component (`add-ui-component` step 2); the padding numbers are the web kit's (README: "don't guess at colors/spacing/sizes") | Restyling `PVBadge` (the map forbids merging them); porting `onRemove` with no caller |
| D-2 | **The type pill composes the Tag.** `CatalogSourceTypePill(label: String)` draws `PVTag(text: label, dot: PVColor.recordMarriage)`; the `iconKey` parameter and the `PVMark` go; the comment says the dot is the record colour. Both callers drop `iconKey:`. `PersonSourcesContent.Row.typeIconKey` loses its only reader and goes, with its assignment and the two test assertions | User, 2026-11-13: one pill on both pages; the board draws the dot `--rec-marriage` and no icon | Two pills (the user chose one) |
| D-3 | **`PVFlowLayout` never draws past its edge.** `static func arrange(width: CGFloat, spacing: CGFloat, lineSpacing: CGFloat, count: Int, fit: (Int, CGFloat?) -> CGSize) -> (frames: [CGRect], height: CGFloat)` *(new)*: each item is measured at its ideal size (`fit(i, nil)`); when that width is more than `width`, it is measured again at `fit(i, width)` and placed at that size. Wrapping is today's rule. `sizeThatFits` and `placeSubviews` call it with `fit = { subviews[$0].sizeThatFits(ProposedViewSize(width: $1, height: nil)) }`, placing each at `bounds.origin` + its frame, vertically centred in its line as today. The doc comment says: "Children are measured at their ideal width; one wider than the row gets the row's width." | The map: "no item is drawn past the row's edge"; a pure rule is testable (precedent `SourcePageOverviewLayout`); composer callers already cap children, so their layout is unchanged when everything fits | Fixing only the header with `ViewThatFits` (the kit layout would still overflow for the next caller) |
| D-4 | **The header's items keep their width; only the date gives.** In `SourcePageIdentityHeader`, the ref `Text` and the Private switch gain `.fixedSize()`; `PVTag` is fixed already (D-1). The date `Text` keeps its `lineLimit(1)`, tail truncation and `.help(line)` and gains `.layoutPriority(-1)` so the `dateAndPrivate` group shrinks there when D-3 proposes the row's width | AC-5.3: the type is never cut; the date already truncates with a tooltip (S15-03) | Wrapping inside a group (the board wraps between the two groups only) |
| D-5 | **The resting description.** The resting `Text` uses `PVFont.body(size: PVTypeScale.caption)`, `PVColor.textMuted`, `.lineSpacing((PVLineHeight.normal - 1) * PVTypeScale.caption)`, `.frame(maxWidth: PVSpacing.measureProse, alignment: .leading)`. The editor is unchanged | User 2026-11-13 and the board (`descResting`: caption, muted, `--lh-normal`, `--measure-prose`) | — |
| D-6 | **The row's pure rules**, as statics on `SourcePageArtifactRow` *(new)*: `static let thumbnailSize: CGFloat = 64`; `static func copyDetailsWidth(_ widthClass: PVWidthClass) -> CGFloat` (320 Regular, 440 Wide); `static func graphLine(_ names: [String]) -> (shown: [String], more: Int)` (the first two names, and the count past them); `static func menuItems(idPrefix: String, canSetCover: Bool, onEdit: @escaping () -> Void, onSetCover: @escaping () -> Void, onDelete: @escaping () -> Void) -> [PVActionMenuItem]` (Edit…, then Set cover only when `canSetCover`, then Delete…; ids `<idPrefix>.edit`, `<idPrefix>.setCover`, `<idPrefix>.delete`; the row passes `idPrefix: "sources.page.artifact.<id>.menu"`). `SourceArtifactsSection.graphNames(for artifactID: String) -> [String]?` *(new)*: nil when `visibleGraphs` is nil, else the names of the graphs whose `artifactIDs` contain the id, in `visibleGraphs` order | Board (`graphLine`: first two joined " · ", then "+n"; widths 320 / 440; menu with Set cover only when it can show); one order with the Graphs section; precedent `ArtifactCopyDetailsView.shown` | Reading the session's graph index (it holds ids per Source, not per Artifact); a separate content struct (three statics are enough) |
| D-7 | **The row view.** `SourcePageArtifactRow` *(new, `macos/App/Features/Sources/SourcePageArtifactRow.swift`)*, a Source page snowflake: `let art`, `@Bindable var model: SourcePageModel`, `let inNoGraph: Bool`, `let projectDir: String`, `let sourceTypeIconKey: String?`. `HStack(alignment: .top, spacing: PVSpacing.space7)`: (1) the thumbnail: a `Button` (`PVFocusRingButtonStyle(cornerRadius: PVRadius.sm)`) around `CachedThumbnail(…, size: 64)` with today's `activatePreview`, `.help`, accessibility label and ids (`…open` / `…addFile`), moved from the card; (2) the text column, `.frame(maxWidth: .infinity, alignment: .leading)`, `VStack(alignment: .leading, spacing: PVSpacing.space2)`: a `PVFlowLayout(spacing: PVSpacing.space4, lineSpacing: PVSpacing.space1)` of the label (`PVTypeScale.body`, semibold, `textPrimary`, `lineLimit(2)`), the ref (mono caption, muted) and the Cover badge (today's); the description when non-blank (`PVTypeScale.bodySmall`, `textSecondary`, `lineLimit(2)`, `.frame(maxWidth: PVSpacing.measureProse, alignment: .leading)`); then either the graph line — `HStack(spacing: PVSpacing.space3)` of `PVIcon(.gitBranch, size: 13)` (muted) and one `Text(verbatim:)` of `shown.joined(separator: " · ")` plus, when `more > 0`, two spaces and `L10n.DesignSystem.badgeMoreCount(more)`, caption, muted, `lineLimit(1)`, with `.accessibilityLabel(L10n.Sources.artifactGraphsSpoken(names:))` *(new)* listing every name joined by ", ", id `sources.page.artifact.<id>.graphs` — when `graphNames` is non-empty, or the no-graph line when `inNoGraph`: `PVFlowLayout(spacing: PVSpacing.space3, lineSpacing: PVSpacing.space2)` of today's badge and "Link to a graph" button (ids and action unchanged); (3) the copy details: `ArtifactCopyDetailsView` (unchanged call) inside `.frame(maxWidth: Self.copyDetailsWidth(widthClass), alignment: .topLeading)` read from `@Environment(\.pvWidthClass)`, so an Artifact with none still takes the column; (4) the actions `HStack(spacing: PVSpacing.space2)`, `.fixedSize()`, `.frame(minHeight: 28)`: Set cover (today's `PVButton`), drawn always and, when `!canUseAsThumbnail`, `.opacity(0)`, `.disabled(true)`, `.accessibilityHidden(true)`; the pencil and trash (today's). The row: `.padding(.vertical, PVSpacing.space6)`, `.overlay(alignment: .bottom) { PVDivider() }`, `.contentShape(Rectangle())`, `.pvRightClickMenu(title: LocalizedStringResource(stringLiteral: art.ref), accessibilityIdentifier: "sources.page.artifact.<id>.menu", items: menuItems(…))` (the ref is data, stored the way S15-34's `PVFormDialogCopy.verbatimTitle` stores a title; it has no catalog entry, so it draws as itself); `.accessibilityElement(children: .contain)`, `.accessibilityLabel(Text(verbatim: art.displayTitle))`, id `sources.page.artifact.<id>` | Board `art-list` row: grid `64px minmax(0,1fr) minmax(0,320|440px) auto`, column gap `--space-7`, padding `--space-6` 0, hairline under each row; label row wraps; description body-sm two lines; graph line caption with a 13pt branch glyph; Set cover keeps its space; no hover (inventory "A row has no hover of its own"). Two flexible frames (`maxWidth: .infinity` and `maxWidth: 320`) share a narrow row as the board's `minmax(0, …)` does (AC-3.3) | A `Grid` across rows (one row's content would size every row); `ViewThatFits` for the no-graph line (the fixed `PVFlowLayout` wraps it) |
| D-8 | **The section.** `SourcePageArtifactsView.body`'s list branch becomes `LazyVStack(alignment: .leading, spacing: 0)` of `SourcePageArtifactRow`, with a `PVDivider()` above the first row, id `sources.page.artifacts.list`. `cardMinWidth`, `cardColumns`, `artifactCard`, `previewWell`, `metaRow`, `restingText`, `noGraphRow`, `descriptionLine`, `activatePreview`, `previewLabel`, `focusedPreviewID` move into the row or go; `ArtifactSquarePreview` goes. The header, empty state, `addForm`, `addFileForm` and `sourceTypeIconKey` stay | The grid's replacement keeps the lazy build at 60 rows | — |
| D-9 | **Copy details column.** `ArtifactCopyDetailsView` gains `static let labelWidth: CGFloat = 96` *(new)*; the field name `Text` gets `.frame(width: Self.labelWidth, alignment: .leading)` (one line, tail-truncated), and the grid's `horizontalSpacing` becomes `PVSpacing.space5`. Everything else is unchanged | Board: `grid-template-columns: 96px minmax(0,1fr)`, `column-gap: var(--space-5)` | — |
| D-10 | **The well goes.** `CachedThumbnail.Layout`, `layout`, `wellBody`, `wellGlyph` and the `switch` in `body` go; `body` is `tileBody.task(id:)`. The type doc comment drops "a frame-filling well" | Its only caller (`previewWell`) goes; nothing passes `layout:` | Keeping `.well` for a later card (no PR in the map draws one) |
| D-11 | **Thumbnail fallback.** A PDF or fileless Artifact draws `PVThumbnail`'s mark at its own size (0.82 of the tile), not the board's 30pt glyph | Inventory "Artifact thumbnail · Component · Ship · `CachedThumbnail` `.tile`", as `SourcesListRow` draws it; no kit change for one call site | A glyph-size parameter on `PVThumbnail` |
| D-12 | **Strings.** New: `L10n.Sources.artifactMenuEdit` ("Edit…", comment: "Artifact row menu item that opens the Artifact sheet"), `artifactMenuDelete` ("Delete…"), `artifactGraphsSpoken(names:)` ("Evidence graphs: %@"). Reused: `useAsThumbnail` ("Set cover", comment widened to "Button and row menu item that pins an Artifact as the Source cover"), `badgeMoreCount`. Removed: `artifactNoDescription`, from `L10n.swift` and `Localizable.xcstrings`. Comments that say "artifact card" on kept keys (`editArtifactLabel`, `coverBadge`, `artifactOpenFileTip`, `artifactAddFileTip`, `artifactNotInGraph`, `artifactLinkToGraph`, `artifactCopyDetailsMore*`) say "artifact row" | `add-localized-string`; the menu copy is the brief's § 2.1 | — |

**Rules this plan bends.** None. The Person page pill change has no board of its own; the user chose it on 2026-11-13 (D-2), and the map's scope entry already named both callers.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVTag` | component | `PVBadge`, `PVChip`, `CatalogSourceTypePill` | The map keeps Badge as it is; a Chip is a selectable token; the pill becomes a thin wrapper over the Tag. Searched `DesignSystem/Components` for `Capsule()` pills with a dot: none |
| `PVFlowLayout.arrange(…)` | static func | `SourcePageOverviewLayout.arrange` | Same shape for testability; extends the existing layout |
| `SourcePageArtifactRow` (+ `thumbnailSize`, `copyDetailsWidth`, `graphLine`, `menuItems`) | view / statics | `SourcePageArtifactsView.artifactCard`, `SourcesListRow` | Replaces the card; one row view keeps the section's file short; `SourcesListRow` is the Sources list's own row |
| `SourceArtifactsSection.graphNames(for:)` | func | `inNoGraph`, `EvidenceGraphIndex.linkedArtifactIDsBySource` | `inNoGraph` answers yes/no for all; names need the graph rows, which `visibleGraphs` holds in list order |
| `ArtifactCopyDetailsView.labelWidth` | constant | — | Board measure |
| L10n keys (D-12) | key | `L10n.Sources` | — |

**Builds on.** S15-34's `ArtifactCopyDetailsView`, `openSheet(artifactID:at:)` and the sheet; `CachedThumbnail` `.tile`; `.pvRightClickMenu`; `PVFlowLayout`; `PVWidthClass`.

**Deletes.** The card (`artifactCard`, `previewWell`, `metaRow`, `restingText`, `descriptionLine`, `noGraphRow` as written, `cardMinWidth`, `cardColumns`, `focusedPreviewID`, `ArtifactSquarePreview`); `CachedThumbnail`'s `.well` (`Layout`, `layout`, `wellBody`, `wellGlyph`); `CatalogSourceTypePill`'s `iconKey` and `PVMark`; `PersonSourcesContent.Row.typeIconKey`; `L10n.Sources.artifactNoDescription`; the README's "Tag · no files yet" row. `PVColor.recordMarriageSoft` loses its last caller and stays: tokens mirror `colors.css` one to one (README § Token file map).

**Next adds.** S15-25: a "Move to another Source…" item between Edit… and Delete… in `SourcePageArtifactRow.menuItems`. S15-40: a mark in the row's label line. A future removable chip: `PVTag` gains `onRemove`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Right-click → Edit… / Set cover / Delete… | `openSheet`, `useAsThumbnail`, `askDeleteArtifact` | The same calls the row's buttons make today; `openSheet` ignores an id no longer in `items`; `useAsThumbnail` guards `canUseAsThumbnail` |
| A graph deleted or relinked elsewhere | `graphNames(for:)` | Read from `visibleGraphs` each render; the graph list's effects refresh it, as `inNoGraph` today |

### Scale

At the map's Scale (60 Artifacts, up to 60 graphs): each row reads `copyDetails(for:)` (S15-34's cost) and `graphNames(for:)`, one pass over the Source's graphs and their `artifactIDs` (60 × 60 comparisons per row, about 216,000 per full render at the extreme, in memory, no store call). `LazyVStack` builds rows as they scroll in; each thumbnail loads through `ThumbnailCache` as today. No write changes.

### Edges

- **Assumptions.** `PVFlowLayout` callers: an item no wider than the row is placed exactly as today (T-1's neighbor case). `visibleGraphs` is sorted by name (`SourceGraphsSection.rows`' comment); the graph line follows it.
- **Neighbors.** An item exactly the row's width stays on its own line at its ideal width; one 1pt wider is re-measured at the row's width (T-1). Two graphs show no "+n"; three show "+1" (T-2). An Artifact with zero details still takes its column (manual AC-1.2).
- **Removed inputs.** An Artifact deleted while its menu is open: the action's callee guards a missing id (Reach). A graph list not loaded: no graph line and no no-graph line (T-3).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** A long label wraps to two lines; a long URL truncates in its column (one line); a type label longer than the header row truncates inside `PVTag`'s one line only when D-3 proposes the row's width.

### Schema

None — no catalog change.

### Go core

None — no engine change.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Artifacts section | `PVSectionHeader`, `LazyVStack`, `PVEmptyState` | `SourcePageArtifactsView.swift` | D-8 |
| Artifact row | snowflake of kit pieces | `SourcePageArtifactRow.swift` *(new)* | D-7 |
| Thumbnail | `CachedThumbnail` `.tile` (`PVThumbnail`), 64pt | `Platform/CachedThumbnail.swift` | D-10, D-11 |
| Copy details | `ArtifactCopyDetailsView` | `ArtifactCopyDetailsView.swift` | D-9 |
| Cover / no graph | `PVBadge` | — | Unchanged |
| Set cover, Link to a graph, +n more | `PVButton` | — | Unchanged |
| Edit / Delete | `PVIconButton` | — | Unchanged |
| Row menu | `.pvRightClickMenu` + `PVActionMenuItem` | — | D-6 |
| Type pill | `PVTag` via `CatalogSourceTypePill` | `Components/Tag/PVTag.swift` *(new)*; `Features/Catalog/SourceTypeMarkers.swift` | D-1, D-2 |
| Meta row | `PVFlowLayout` | — | D-3, D-4 |

**Frames in words.** *Twelve Artifacts* (`art-list`, Regular): under the "Artifacts · 12" header, a hairline, then rows 24pt apart top and bottom with a hairline under each. Row: a 64pt rounded-square image; "Sheet 14 (image 1)" semibold, `ART-7QK2M` mono muted and an accent "COVER" badge on the first; one or two lines of description in secondary body-small; a muted branch glyph and "Hale household, dwelling 113 · Robins household, dwelling 112"; a 320pt column of up to four lines "Provider   Ancestry" (name 96pt muted, value mono); then a ghost "Set cover" (blank space on the cover), a pencil and a red trash. "Sheet 20" shows the grey "NOT IN AN EVIDENCE GRAPH" badge and "Link to a graph" in its graph line's place. "Sheet 19" (PDF) shows the PDF mark; "Sheet 24" the type mark. *Wide* (`art-list-wide`): the same rows in the main column beside the 480pt Metadata column; the details column is 440pt. *Sixty* (`art-sixty`): the same rows, scrolled. *Menu* (`art-menu`): a panel titled `ART-7QK2M` with "Edit…", "Set cover", "Move to another Source…" (S15-25, not this PR), "Delete…". *Empty* (`art-empty`): today's empty state. *Header* (`page`, `header-wrap`): `SRC-4KD2P · (● Census) ✎`, then `· 📅 date ✎ · Private`; the Tag is a white rounded pill with a thin border and a plum dot; with a range date and the picker open, the date group wraps to the second line. *Description*: small muted prose, about 480pt wide.

States: empty (today's), graphs not loaded (no graph line), large (60 rows, eight details → four and "+4 more"). Accessibility: SA-11 per D-7; the menu items are VoiceOver actions (the kit adds them); the hidden Set cover is hidden from VoiceOver.

### Strings

D-12. The change of what an action covers: none — Set cover and Delete keep their scope; the menu repeats them.

## 7. Steps

**STEP-1 — Kit: the Tag; the type pill composes it.**
Files: `macos/App/DesignSystem/Components/Tag/PVTag.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/App/Features/Catalog/SourceTypeMarkers.swift`, `macos/App/Features/Sources/SourcePageIdentityHeader.swift` (the call), `macos/App/Features/Conclusions/PersonSourcesSection.swift` (the call), `macos/App/Features/Conclusions/PersonSourcesContent.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`. Skill: `add-ui-component`.
Do: D-1, D-2. README: a Tag row in the built table (`Components/Tag/PVTag.swift` (S15-35: a pill with a colour dot, ported from `PVTag.jsx` without remove or hover; the Source type pill composes it)); the "no files yet" Tag row goes.
Check: `xcodebuild test` passes (`PersonSourcesContentTests` without the `typeIconKey` assertions); `rg -n "typeIconKey" macos/App/Features/Conclusions` finds nothing; `python3 scripts/check-guidance-names.py` passes.

**STEP-2 — The flow layout keeps every item inside its row; the header keeps the type whole.**
Files: `macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift`, `macos/App/DesignSystem/README.md` (the FlowLayout row: "an item wider than the row gets the row's width"), `macos/App/Features/Sources/SourcePageIdentityHeader.swift`, `macos/ProvenenciaTests/PVFlowLayoutTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-3, D-4.
Check: T-1 passes; manual AC-5.3.

**STEP-3 — The Source description reads at caption size.**
Files: `macos/App/Features/Sources/SourcePageDescriptionView.swift`.
Do: D-5.
Check: build passes; manual AC-5.4.

**STEP-4 — The row's rules.**
Files: `macos/App/Features/Sources/SourcePageArtifactRow.swift` *(new)* (the statics only), `macos/App/Features/Sources/SourceArtifactsSection.swift` (`graphNames(for:)`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings` (the menu keys), `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-6 and D-12's menu keys.
Check: T-2, T-3, T-4 pass.

**STEP-5 — The Artifacts section is a list of rows.**
Files: `SourcePageArtifactRow.swift` (the view), `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Features/Sources/ArtifactCopyDetailsView.swift`, `macos/App/Platform/CachedThumbnail.swift`, `L10n.swift`, `Localizable.xcstrings` (`artifactGraphsSpoken`).
Do: D-7, D-8, D-9, D-10, D-11. A `#Preview` of three rows from fixtures (cover image with three details, a PDF in no graph, an Artifact with eight details) at Regular and at Wide (`.environment(\.pvWidthClass, .wide)`).
Check: build passes; `rg -n "ArtifactSquarePreview|cardMinWidth|wellBody|wellGlyph|layout: \.well|artifactCard" macos` finds nothing; the preview matches § 6 "Frames in words"; manual AC-1.1 – AC-4.1.

**STEP-6 — Strings sweep.**
Files: `L10n.swift`, `Localizable.xcstrings`.
Do: D-12's removed key and the "card" → "row" comments.
Check: `rg -n "artifactNoDescription" macos .agents .cursor docs/*.md` finds nothing.

**STEP-7 — Version bump.**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (every `MARKETING_VERSION`). Skill: `.cursor/rules/versioning.mdc`.
Do: 0.13.0 (or the next minor after `main`).
Check: the three agree.

**STEP-8 — Docs.**
Files: `docs/deployment-plan/spike-15/completed.md`; `docs/deployment-plan/spike-15/deployment-plan.md` (map row); this plan (status, Landed note); `docs/deployment-plan/spike-15/design/README.md` (S15-D14 stays until S15-25 lands; note S15-35 landed); `docs/deployment-plan/spike-15/design/S15-D14-source-page-artifacts.md` § 6 inventory "Artifact entry" home → `macos/App/Features/Sources/SourcePageArtifactRow.swift` *(new)*.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-35-artifact-list-rows.md` reports no errors.

Commits: one per step. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Show a Source's Artifacts as rows with their copy details in a column.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVFlowLayoutTests.swift` *(new)* `wrapsAsBefore`, `widerThanTheRowGetsTheRowWidth`, `exactlyTheRowWidthStaysIdeal` | Widths 100, 120 in a 250 row, spacing 10 → one line, frames x 0 and 110; 100, 200 in 250 → two lines; one item ideal 300 (re-measured at 250 → 250 × 40) in 250 → frame width 250, height 40, `fit` called with 250; one item of exactly 250 → width 250, `fit` called only with nil | D-3, AC-5.3 |
| T-2 | Swift | `SourcePageModelTests.swift` `artifactRowGraphLine` *(new)* | `graphLine([])` → ([], 0); two names → both, 0; three → first two, 1 | D-6, AC-1.3 |
| T-3 | Swift | same, `artifactGraphNamesFollowTheGraphList` *(new)*, using `makeCensus(graphs:)` and `graph(…)` | Graphs "Smith household" [P3, P9], "Jones household" [P3, P4] as the session lists them → `graphNames(for: "P3")` is the names in `visibleGraphs` order; `"P6"` → []; before the graph list loads → nil | D-6, AC-1.3 |
| T-4 | Swift | same, `artifactRowMenuAndWidths` *(new)* | `menuItems(canSetCover: true)` titles → "Edit…", "Set cover", "Delete…"; `false` → "Edit…", "Delete…"; invoking each item calls its closure; `copyDetailsWidth(.regular)` → 320, `(.wide)` → 440 | D-6, AC-2.2, AC-2.3, AC-3.1, AC-3.2 |
| T-5 | Swift | `PersonSourcesContentTests.swift` *(updated)* | The `typeIconKey` assertions go; `typeLabel` assertions stay | D-2 |

Kept unchanged: `copyDetailsShowFourThenMore`, `restingCardMarksAnArtifactInNoGraph` (renamed nothing; it pins `inNoGraph`), every S15-34 sheet test.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests`, `PersonSourcesContentTests`, and the Citation composer's chip rows (they use `PVFlowLayout`; manual check step 6).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — PVTag is the map's scoped port (user, 2026-11-13)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief: S15-D14 stays in design/ until S15-25 lands (its handoff); design/README.md notes S15-35 landed; the brief's inventory names the row's home
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with a census Source (type Census), three Evidence graphs ("Robins household, dwelling 112", "Hale household, dwelling 113", "Ostrander household, dwelling 114"), twelve Artifacts (ten JPEGs, one PDF, one with no file; the first is the cover; one in no graph; one in all three graphs; one with three copy details, one with eight, one with none), a range date, and a description of three sentences. Also a Source with 60 Artifacts and one with none.

1. Regular (window 1366 × 900), the census Source: rows as AC-1.1; the details columns line up (AC-1.2, AC-2.1); graph lines and "+1" (AC-1.3); no "No description" (AC-1.4); thumbnails sharp on a 2x display, the PDF and fileless marks in the tile (AC-1.5); click a thumbnail, then a fileless one (AC-1.6).
2. Right-click a non-cover image row, then the cover row: the menus of AC-2.2 and AC-2.3; use each item once. "+4 more", "Link to a graph" and the pencil open the sheet where AC-2.4 says.
3. Stretch the window to Wide (pane ≥ 1600pt): the details column widens to 440pt beside the Metadata column (AC-3.2). Narrow the window to its minimum: nothing overlaps the actions (AC-3.3).
4. The 60-Artifact Source: scroll to the middle (AC-1.7). The empty Source: the empty state (AC-1.8).
5. The header: the Tag (AC-5.1). Change the type to a long one ("Church register (parish copy)"), keep the range date, open the type picker, and narrow the window: the row wraps and nothing draws past its edge (AC-5.3). The description reads small and muted (AC-5.4).
6. A Person page with Sources: each row's type is the Tag and fits its column (AC-5.2). The Citation composer's chip rows wrap as before.
7. VoiceOver on a row (AC-4.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The header's clip has a cause besides `PVFlowLayout` (a frame or the page band) | D-3 and D-4 fix the layout's overflow; manual step 5 checks; a Stop condition covers the rest |
| Two flexible frames in the row share space unlike the board | Manual step 3 at minimum width; AC-3.3 |
| `LazyVStack` inside `SourcePageOverviewLayout` sizes differently from the old `LazyVGrid` | The grid sat in the same layout; the manual check scrolls 60 rows |
| `PVTag` in the Person page's 176pt type column is wider than the old pill | It truncates its one line inside the column; manual step 6 |

**Stop and report if:**

- After STEP-2, the header's type is still cut off at the manual check's step 5 (the clip comes from somewhere D-3 does not reach): report the window width and a screenshot.
- The row menu's title draws anything but the ref (D-7).
- Re-measuring a `PVFlowLayout` item at the row's width changes a Citation composer row that fit before.
- `visibleGraphs` is not in the Graphs section's order at the planning base (D-6).
- The board's frames and an AC disagree.

## 12. Open questions

None.

## 13. PR

**Title:** S15-35 — The Artifact list shows each Artifact's copy details
**Body:** objective in one sentence; what shipped against US-1 – US-5; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-35-artifact-list-rows.md`.
