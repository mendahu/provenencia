# S13-35 — The person card joins the design system

**Status:** Landed in #690 (2026-11-02)
**Spike / slice:** Spike 13 · Slice 4 — Timeline (preparatory)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: Family looks, reads and behaves as before; the compact size is first drawn by S13-13, from board S13-D4 turn 12)
**Board:** — (the compact size's measures come from board S13-D4 turn 12, frame 12f, as the S13-13 plan records them)
**Depends on:** S13-34 ([plan](S13-34-display-pieces-move-down.md), Planned: the text pieces, value formatting and the kit check move down a layer first)
**Blocks:** S13-13
**Branch:** `s13-35-person-card`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** working rule "The person card lives in the design system; display is decoupled from content" ([`../deployment-plan.md`](../deployment-plan.md#client)); [`docs/design-system-layers.md`](../../../design-system-layers.md); the evaluation of `PersonTile` (2026-11-02, plan items 3 and 4); the S13-12 plan [`S13-12-family-section.md`](S13-12-family-section.md) and its **Landed** note; the S13-13 plan [`S13-13-timeline-section.md`](S13-13-timeline-section.md) (its compact card, Return modifier and focus-ring style are taken over here); the S13-34 plan
**Planned:** 2026-11-02 against `main` @ `04a2c99` (S13-34's base); STEP-0 re-checks S13-34's landed names
**Landed:** the code differs from the plan in four places. (1) The compact years use the app's qualifier wording ("abt 1790"), as S13-34 landed it, not "abt." (T-2 asserts the app's wording). (2) `PersonDetailContent.badge(_:)` is a static function taking the treatment, so `PVGenderTreatmentTests` can check the badge without a detail fixture. (3) The shell's `hasTrailing` is read from the `Trailing` type (`EmptyView`) at init. (4) The regular lines slot applies no line limit (as the tile did); `linesLineLimit` is read by the compact years line only.
**Review:** [`S13-35-person-card-joins-the-design-system.review.md`](S13-35-person-card-joins-the-design-system.review.md) — round 1 Fits with changes / Not ready; every finding Fixed; the user's answer of 2026-11-02 (no `.row` size) applied; none open

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

The person card becomes a design-system piece that any feature can place: a content-agnostic card shell (`PVTileCard`) with slots, a regular and a compact size and its interaction, and on it a person card recipe (`PVPersonCard`) in both sizes. A family tree in another feature will reuse it (user, 2026-11-02). The gender colour treatment becomes a recipe too, so the card reads it without reaching into the Person feature. The kit gains one Return modifier and one focus-ring button style, and the existing hand-written copies move onto them.

**Why now.** S13-13 places the compact card; building it on the final card avoids moving it twice (user, 2026-11-02: two preparatory PRs ahead of S13-13).

**What stays.** Family looks, reads and behaves exactly as today: tile size, focus tile, hover lift, focus ring, Return, tooltips (the title only) and spoken labels. The header's gender badge and the "no likeness" wording. The gallery tile, the Source page's artifact preview and the citation composer's rows behave as today.

## 2. User stories

**US-1.** As a contributor building a new view (a family tree), I place a person card from the design system without importing the Person feature.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `scripts/check-guidance-names.py` (S13-34's kit check) after this PR | Passes; the card, the shell, the gender treatment and the two kit pieces name no `Features/` type |
| AC-1.2 | A `PVPersonCard(content:projectDir:size: .compact)` preview with Anna (woman, likeness, 3 May 1790 Bremen – 1851) | 52pt card, 40pt likeness with the woman tint and line, "Anna Weber" on one line, "1790 – 1851" in mono caption muted, one line; tooltip "Anna Weber\nb. 3 May 1790 · Bremen\nd. 1851" (D-5, D-6) |
| AC-1.3 | The same card's spoken label with role "Mother" | "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851" (unchanged rule) |

**US-2.** As a researcher, Family and the screens whose controls move onto the kit are unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Johann's Family (S13-12 frame 11a) at Regular and Wide | As before: 92pt tiles, 70pt likeness, two-line names, *b.* and *d.* lines, the raised focus tile with "This person" |
| AC-2.2 | Hover a relative's tile; Tab to it; press Return; press Space | Grows 2.5% with the medium shadow; the focus ring; Return and Space open the relative's Family (D-2, D-3) |
| AC-2.3 | Tab to a "Show N more" row and press Return | It unfolds (D-2) |
| AC-2.4 | VoiceOver on Anna's and Johann's tiles | "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851", a button; "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person", static text |
| AC-2.5 | Hover a Family tile | The tooltip is the name only, as today (D-5) |
| AC-2.6 | Tab to a gallery tile on a Person page and to an artifact preview on a Source page | The same focus ring as today (radius `sm`, and 0 on the artifact preview) (D-3) |
| AC-2.7 | In the citation composer, Tab to an Observation row and press Return | It opens for editing, as today; Space too (D-2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `pvActivatesOnReturn` (two forms) and `PVFocusRingButtonStyle(cornerRadius:)` in the kit; the tile, Family's fold row, the composer row, the gallery tile and the artifact preview move onto them (D-2, D-3) | STEP-1, AC-2.2, AC-2.3, AC-2.6, AC-2.7 |
| `PVTileCard` shell with regular and compact sizes (D-1, D-4) | STEP-2 |
| `PVGenderTreatment` recipe; the detail badge built in `PersonDetailContent` (D-7) | STEP-3 |
| `PVPersonCard` recipe at the regular size; Family moves onto it; `PersonTile.swift` deleted; its own L10n namespace (D-5, D-9) | STEP-4, AC-2.1, AC-2.4, AC-2.5 |
| The compact size: life span at year precision and its tooltip (D-6) | STEP-5, AC-1.2, AC-1.3 |
| Docs, the working rule's and R-3's type name (D-10) | STEP-6, AC-1.1 |

| Out | Home |
| --- | --- |
| Moving the title styling, date · place line and value formatting down a layer; the kit check | S13-34 |
| The Timeline, its event row and the participant grid | S13-13 (it places `PVPersonCard(size: .compact)`; its event row is a Timeline snowflake, user 2026-11-02) |
| A row size on the shell | not planned (user, 2026-11-02); an event card with a surface uses the shell's sizes and becomes a recipe on its second caller |
| Event and place cards (`PVEventCard`, `PVPlaceCard`) | when each has a second caller; not planned |
| `CitationLineButton`'s own ring (`ReconciliationRecordRows.swift:276`) and `SourcesGraphButtonBody`'s ring with hover | logged in `docs/ideas/structural-debt.md` by STEP-6 |
| Moving `CachedThumbnail` out of `Platform/` | not planned (a `DesignSystem` → `Platform` dependency already exists: `PVArtifactChecklist`) |
| `PVOmnibarHitRow` onto `PVTileCard` or `PVTitleText` | logged by S13-34 |
| Family's own "This person" sibling group label (`L10n.Conclusions.familyThisPerson`) | stays Family's; only the card's badge and spoken suffix move (D-9) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Person tile | `macos/App/Features/Conclusions/PersonTile.swift` `PersonTileContent` (`init(summary:role:isFocus:locale:orders:)`; stored `id`, `title`, `lines`, `imageRelPath`, `colors`, `role`, `isFocus`, `spokenLabel`; no summary kept), `PersonTile` (`content`, `projectDir`, `open`; `nonisolated static let height: CGFloat = 92`), private `PersonTileButtonStyle`, `PersonTileButtonBody` | `PVCard` (raised and elevated for the focus, padding 10); an `HStack(spacing: 10)` of a 70pt `CachedThumbnail` and a `VStack(alignment: .leading, spacing: 3)` (`.frame(maxWidth: .infinity, alignment: .leading)`) of the title row and lines; the `HStack` `.frame(maxWidth: .infinity, minHeight: 70, maxHeight: 70, alignment: .leading)`; `.frame(height: 92)`, `.clipped()`. A relative is a `Button` with a hand-written `@FocusState` + `.onKeyPress(.return)`, `.pvLiftHover(scale: 1.025)`, `.help(title.text)` |
| Family callers | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `Tile.content: PersonTileContent` (lines 41, 110, 148); `PersonFamilySection.swift` `PersonFamilyDiagram` (line 133); `PersonFamilyLayout.swift` `PersonTile.height` (line 86) and the literal minimum 120 (line 89) | Four call sites |
| Fold row | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow` | A ghost `PVButton` (`.pv(.ghost, size: .sm)`) with the same hand-written Return handling; it needs no focus-ring style |
| Composer row | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift:200 – 213` | `.focusable()`, `.focused($lineFocus, equals: .row(line.id))`, `.onKeyPress(.return)` and `.onKeyPress(.space)` guarded by `lineFocus == .row(line.id)`: a third Return copy, its focus an enum binding the caller owns |
| Ring copies | `macos/App/Features/Conclusions/ConclusionDetailPage.swift:630 – 632` (gallery tile); `macos/App/Features/Sources/SourcePageArtifactsView.swift:208 – 210` (artifact preview, radius 0) | `.buttonStyle(.plain)` + `.focused($id, equals:)` + `.pvFocusRing(id == …, cornerRadius:)` |
| Gender treatment | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `PersonGenderTreatment` (`Family`, `Basis`, `text`, `fieldLabel`, `badge: ConclusionDetailBadge`, `thumbnailColors`, `read(_:)`); callers `PersonDetailContent.swift:15, 20, 36`, `PersonTile.swift:31`; doc comment `macos/App/Platform/GenealogyStore.swift:468`; tests `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` | Its colours are kit tokens; `badge` returns the feature's `ConclusionDetailBadge` and words with `L10n.Conclusions.sexAtBirthBadge`; after S13-34 its text calls `PropertyTermDisplay.name(_:propertyKey:)` |
| Display pieces (after S13-34) | `PVTitleSource`, `PVTitleText(_:style:extraCount:lineLimit:emphasized:)` in `macos/App/DesignSystem/Components/TitleText/` (S13-34); `PVDatePlaceLine(leading:date:place:accessibilityLabel:missing:dateSize:extraCount:)` and `PVDatePlaceLine.Missing` (`.stated`, `.omitted`) in `macos/App/DesignSystem/Components/DatePlaceLine/` (S13-34); `PersonLifeDisplay.lifeLines(_:locale:missing:)`, `dateText(_:locale:precision:)`, private `dateSpan(birth:death:locale:precision:)` and `DateRowDisplay.Precision` (`.full`, `.year`) in `macos/App/Platform/`; `PersonTitleDisplay` in `macos/App/Platform/`; `PropertyTermDisplay.name(_:propertyKey:)`; `PVBadge.moreCount`; the kit check in `scripts/check-guidance-names.py` | Not landed; STEP-0 checks the names *(new)* |
| Kit pieces | `macos/App/DesignSystem/Components/Card/PVCard.swift` `PVCard` (`tone`, `cornerRadius`, `elevated`, `padding`); `macos/App/DesignSystem/Components/HoverEffect/PVLiftHover.swift` `pvLiftHover(scale:resting:)`; `macos/App/DesignSystem/Tokens/PVElevation.swift` `pvFocusRing(_:cornerRadius:)`; `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail` | — |
| Recipes | `macos/App/DesignSystem/Recipes/` (`ArtifactChecklist` reads `CatalogArtifact` and `CachedThumbnail`; `DeleteImpact` takes `WorkspaceLocation`, a `Features/` type, allowed by S13-34's check; `Marks`; `OmnibarHitRow`; `LastUpdated`); `macos/App/DesignSystem/README.md` built table, one row per folder, rows written as `` `Components/HoverEffect/PVHoverEffect.swift` `` | Recipes may read `Catalog*` types and `Platform/` |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyThisPerson` (the card's badge and Family's sibling group label, `PersonFamilyContent.swift:159`), `familyThisPersonSpoken` (the card only), `personNoLikeness` (the card and the header, `PersonDetailContent.swift:51`; key `conclusions.person.noLikeness`, `Localizable.xcstrings:16701`) | — |
| Rule and requirement names | `docs/deployment-plan/spike-13/deployment-plan.md` working rule "Colors reuse `PersonGenderTreatment`" (line 111); `docs/deployment-plan/spike-13/person-detail.md` R-3 (line 13) | Name the type this PR renames |
| Tests | `macos/ProvenenciaTests/PersonTileTests.swift` (`tileContentFromSummary`, `spokenLabels`), `PersonGenderTreatmentTests.swift`, `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift` | — |

**Precedent.** S13-12 D-16 moved the artifact card's private hover into the kit as `pvLiftHover` with a README row and both callers unchanged; copy that shape for the Return modifier and the focus-ring style. `PVArtifactChecklist` is the precedent for a recipe that reads `Catalog*` data and `CachedThumbnail`, and `PVArtifactChecklistTests.swift` for a kit test named after its type.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The shell.** `struct PVTileCard<Thumbnail: View, Title: View, Lines: View, Trailing: View>: View` *(new)* in `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)*. `init(size: PVTileCardSize, isRaised: Bool = false, accessibilityLabel: String, help: String, action: (() -> Void)? = nil, @ViewBuilder thumbnail:, @ViewBuilder title:, @ViewBuilder lines:, @ViewBuilder trailing:)`, and a second init in `extension PVTileCard where Trailing == EmptyView` without `trailing`. Layout: `PVCard(tone: isRaised ? .raised : .card, cornerRadius: PVRadius.sm, elevated: isRaised, padding: size.padding)` holding an `HStack(alignment: .center, spacing: 10)` of the thumbnail and a `VStack(alignment: .leading, spacing: 3)` (`.frame(maxWidth: .infinity, alignment: .leading)`) of the title row (an `HStack(spacing: PVSpacing.space4)` of the title and the trailing with `.layoutPriority(1)` when `Trailing != EmptyView`, else the title alone) and the lines; the `HStack` `.frame(maxWidth: .infinity, minHeight: size.thumbnail, maxHeight: size.thumbnail, alignment: .leading)`; then `.frame(height: size.height)`, `.clipped()`. With `action` nil the card is static: `.accessibilityElement(children: .ignore)` with the label, no hover, no focus stop. With an action: `Button(action:)` with `PVFocusRingButtonStyle()`, `.pvActivatesOnReturn(action)`, `.pvLiftHover(scale: 1.025)`, `.accessibilityLabel`. `.help(help)` on both | Content-agnostic: slots and tokens only, so a Component; the same layout as today's tile at 92 / 70 | A row size without a surface (user, 2026-11-02: cut; a different layout behind a size, and its static mode hid a title button); a recipe |
| D-2 | **Return activation.** Two forms in `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*: `func pvActivatesOnReturn(_ action: @escaping () -> Void) -> some View`, a private `ViewModifier` owning `@FocusState private var focused`, applying `.focused($focused)` and the Return handler (today's `PersonTile` body); and `func pvActivatesOnReturn(when isFocused: Bool, _ action: @escaping () -> Void) -> some View`, the handler only, for a caller that owns its focus binding. Space stays the `Button` default. `PVTileCard` and `PersonFamilyMoreRow` use the first; the composer row uses the second (`when: lineFocus == .row(line.id)`) and keeps its own `.focused` and Space handler (it is not a `Button`); S13-13 adds the event title | One Return mechanism for all three copies (review M-3); content-agnostic | A third hand copy; a modifier that cannot serve an enum focus binding |
| D-3 | **Focus-ring style.** `struct PVFocusRingButtonStyle: ButtonStyle` *(new)* in `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` *(new)*, `init(cornerRadius: CGFloat = PVRadius.sm)`, with a private body view reading `@Environment(\.isFocused)`: the label, `.contentShape(Rectangle())`, `.pvFocusRing(isFocused, cornerRadius:)`, no other chrome (today's `PersonTileButtonStyle` / `PersonTileButtonBody`). The gallery tile (`ConclusionDetailPage.swift:630 – 632`) and the artifact preview (`SourcePageArtifactsView.swift:208 – 210`, `cornerRadius: 0`) move onto it: `.buttonStyle(PVFocusRingButtonStyle(cornerRadius:))`, keeping their `.focused(…, equals:)` for programmatic focus and dropping their `.pvFocusRing`. `PersonFamilyMoreRow` keeps its ghost `PVButton` style | One place that draws a bare button's ring (review M-4) | A fixed radius; a feature style |
| D-4 | **Sizes.** `enum PVTileCardSize: Equatable { case regular, compact }` *(new)*, beside `PVTileCard`, non-generic, with `nonisolated` non-optional properties: `height` (92, 52), `thumbnail` (70, 40), `titleLineLimit` (2, 1), `linesLineLimit` (2, 1: the compact card's one line), `minimumWidth` (120, 232), `padding` (10, 6). The shell applies `height`, `thumbnail` and `padding`; the title and lines slots apply `titleLineLimit` and `linesLineLimit` (a doc line says so). `PersonFamilyLayout` reads `PVTileCardSize.regular.height` and `.minimumWidth` in place of `PersonTile.height` and its literal 120 | One place for every card measure; compact from board S13-D4 frame 12f | Optionals for a row size (cut) |
| D-5 | **The person card.** `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)*: `struct PVPersonCardContent: Equatable` (today's `PersonTileContent`, renamed and moved, same `init(summary:role:isFocus:locale:orders:)`, reading S13-34's `Platform/` `PersonTitleDisplay` and `PersonLifeDisplay` and D-7's `PVGenderTreatment`), with two stored fields *(new)*: `years: String` (D-6, filled in the init) and `helpText: String` (D-6). `struct PVPersonCard: View`, `init(content: PVPersonCardContent, projectDir: String, size: PVTileCardSize = .regular, open: @escaping () -> Void = {})`. It fills `PVTileCard`: thumbnail `CachedThumbnail(projectDir:relPath:size: size.thumbnail, cornerRadius: PVRadius.sm, fallbackMark: .subjectPerson, fallbackLabel: L10n.PersonCard.noLikeness, colors:)`; title `PVTitleText(content.title, lineLimit: size.titleLineLimit, emphasized: content.isFocus)` in `PVListText.titleFont`, `PVColor.textPrimary`; trailing at the regular size, the focus's `PVBadge(L10n.PersonCard.thisPerson)`; lines at regular one `PVDatePlaceLine(…, missing: .omitted)` per life line, at compact one `Text(verbatim: content.years)` (`lineLimit(1)`, not drawn when empty); `isRaised: content.isFocus`; `action` nil for the focus, `open` otherwise; `accessibilityLabel: content.spokenLabel`; `help`: the title text at regular (as today), `content.helpText` at compact. A compact focus draws raised with no badge (no caller places one) | The owner's direction: a semantic card in the design system; a recipe reads `CatalogPersonSummary` as `PVArtifactChecklist` reads `CatalogArtifact`; Family's tooltip unchanged (review M-8) | Keeping the tile in the feature (user, 2026-11-02); a mirror `Size` enum (review M-7) |
| D-6 | **Compact content.** `PersonLifeDisplay.years(_ summary: CatalogPersonSummary?, locale: Locale) -> String` *(new, in S13-34's `Platform/PersonLifeDisplay.swift`)*: `dateSpan(birth:death:locale:precision: .year)`: "1790 – 1851", "1820 –", "– 1851", "abt. 1790 – 1851", "bet 1790 and 1795 – 1851", a phrase-only birth's phrase in its place ("during the war – 1851"); "" when neither date has a year or a phrase. `PersonLifeDisplay.LifeLine.text` *(new)*: the abbreviation, a space, then date and place joined with `L10n.Conclusions.lineJoin`, a missing half left out ("b. 3 May 1790 · Bremen", "d. 1851"). `helpText`: the title text, then each life line's `text`, joined by "\n"; with no lines, the title alone. The years line uses `PVFont.mono(size: PVTypeScale.caption)`, `PVColor.textMuted` | One rule for the life span (S13-34 D-9; user, 2026-11-02); the board's tooltip | A raw `startYear` read; new keys |
| D-7 | **Gender treatment recipe.** `PersonGenderTreatment` moves to `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` *(new)* as `PVGenderTreatment`, with `Family`, `Basis`, `family`, `basis`, `text`, `fieldLabel`, `thumbnailColors`, `read(_ summary: CatalogPersonSummary?)`; its text as S13-34 leaves it. `badge` is deleted from it and built in `PersonDetailContent.badge` (`PersonDetailContent.swift:20`, its only reader) from `treatment.text`, `treatment.basis` and `L10n.Conclusions.sexAtBirthBadge`. `PersonDetailContent`, the card and `GenealogyStore.swift:468`'s doc comment name `PVGenderTreatment`. Tests move to `macos/ProvenenciaTests/PVGenderTreatmentTests.swift` (`git mv`) | A recipe because it reads `CatalogPersonSummary`; the badge is detail-page chrome (review M-10: one reader, no new file) | An extension file for one member |
| D-8 | **Layering.** S13-34's kit check in `scripts/check-guidance-names.py` is the evidence (AC-1.1): `check-pr.sh` runs it. This PR adds no check of its own | One check (S13-34 review M-9) | A one-off `rg` loop |
| D-9 | **Copy.** The card's strings move to `L10n.PersonCard` *(new namespace)*: `thisPerson` = "This person" (key `personCard.thisPerson`), `thisPersonSpoken` = "this person" (`personCard.thisPersonSpoken`), `noLikeness` (from `L10n.Conclusions.personNoLikeness`, key `personCard.noLikeness`, same copy). `familyThisPersonSpoken` and `personNoLikeness` are deleted: the header (`PersonDetailContent.swift:51`) reads `L10n.PersonCard.noLikeness` (one key for a person's missing likeness, review M-6). `familyThisPerson` stays for Family's sibling group label. String Catalog entries move with the keys | A recipe owns its copy (`PVDeleteImpact` → `L10n.DeleteImpact`) | Two keys with one copy |
| D-10 | **Names in docs and tests.** The working rule "Colors reuse `PersonGenderTreatment`" and R-3 in `person-detail.md` name `PVGenderTreatment` (the planning commit). Tests: `PersonTileTests.swift` → `macos/ProvenenciaTests/PVPersonCardTests.swift` (`git mv`, suite `PVPersonCardTests`). The layering sentence is S13-34's (no second sentence here) | Guides and tests point at the real types (review M-5) | Exempting the old names in the sweep *(new)* |

**Rules this plan bends.** None. The working rule "Colors reuse `PersonGenderTreatment`" is renamed to the moved type, not changed.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVTileCard`, `PVTileCardSize` *(new)* | component | `PVCard`, `PVList`'s row slots, `PVOmnibarHitRow`, `PersonTile` | `PVCard` is the surface it uses; `PVList` and `PVOmnibarHitRow` are rows without a card or sizes; the tile's layout becomes the shell |
| `pvActivatesOnReturn` (two forms) *(new)* | component | three hand-written Return handlers | Extracted (D-2) |
| `PVFocusRingButtonStyle(cornerRadius:)` *(new, moved)* | component | `PersonTileButtonStyle`; two `.plain` + ring copies | Moved and parameterized (D-3) |
| `PVGenderTreatment` *(moved)* | recipe | `PersonGenderTreatment` | Moved (D-7) |
| `PVPersonCard`, `PVPersonCardContent` *(moved)* | recipe | `PersonTile`, `PersonTileContent` | Moved onto the shell (D-5) |
| `PersonLifeDisplay.years`, `LifeLine.text`, `PVPersonCardContent.years`, `.helpText` *(new)* | func, field | `PersonLifeDisplay.dateSpan`, `lifeLines` | Extended (D-6) |
| `L10n.PersonCard` *(new)* | keys | `L10n.Conclusions.family*`, `personNoLikeness` | Moved (D-9) |

**Builds on.** S13-34's `PVTitleText`, `PVTitleSource`, `PVDatePlaceLine`, `DateRowDisplay` precision, `Platform/` formatters and kit check; `PVCard`, `pvLiftHover`, `pvFocusRing`, `PVBadge`, `CachedThumbnail`.
**Deletes.** `macos/App/Features/Conclusions/PersonTile.swift` (`PersonTile`, `PersonTileContent`, `PersonTileButtonStyle`, `PersonTileButtonBody`); `macos/App/Features/Conclusions/PersonGenderTreatment.swift` (moved) and its `badge`; the hand-written Return in `PersonFamilyMoreRow` and the composer row; the explicit `.pvFocusRing` on the gallery tile and the artifact preview; `L10n.Conclusions.familyThisPersonSpoken`, `personNoLikeness`; `PersonTile.height` and `PersonFamilyLayout`'s literal 120.
**Next adds.** S13-13 places `PVPersonCard(size: .compact)` and uses `pvActivatesOnReturn` and `PVFocusRingButtonStyle` on its event title; a family tree places `PVPersonCard` from its own feature; `PVEventCard` and `PVPlaceCard` fill the shell when each has a second caller.

### Reach

None — no new path into existing code: Family's tiles keep their destinations, the moved controls keep theirs, and the compact size has no caller until S13-13.

### Scale

None — no write, list or loop changes; one card per tile as today.

### Edges

- **Neighbors of the compact span.** A range birth → "bet 1790 and 1795 – 1851"; a phrase-only birth → the phrase; a long phrase or range → cut at the tail on one line (`linesLineLimit` 1), the full line in the tooltip; no dated birth or death → no years line (T-2).
- **Assumptions.** The shell's static mode hides its children from VoiceOver (`.ignore`): a static card's slots hold no control (the person card's focus has none). No other edge: the span's wording is S13-34's rule.

### Schema

None — no catalog change.

### Go core

None.

### FFI

None.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Card shell | `PVTileCard`, `PVTileCardSize` | `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)* | D-1, D-4; `#Preview` with a static, a raised and an actionable card at each size, text slots only |
| Return | `pvActivatesOnReturn` | `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)* | D-2 |
| Focus ring | `PVFocusRingButtonStyle` | `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` *(new)* | D-3 |
| Gender treatment | `PVGenderTreatment` | `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` *(new)* | D-7 |
| Person card | `PVPersonCard`, `PVPersonCardContent` | `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)* | D-5, D-6; `#Preview` at both sizes, its `CatalogPersonSummary` built inline in the file (a recipe reads no feature fixture) |

README rows in the built table, in the existing format and alphabetical place: Components `` `Components/FocusRingButton/PVFocusRingButtonStyle.swift` `` (the kit way to give a bare button a ring), `` `Components/ReturnActivation/PVReturnActivation.swift` `` (Return activates a focused control), `` `Components/TileCard/PVTileCard.swift` `` (a card with thumbnail, title, lines and trailing slots at regular and compact sizes); Recipes `` `Recipes/GenderTreatment/PVGenderTreatment.swift` `` (a person's gender colours and wording), `` `Recipes/PersonCard/PVPersonCard.swift` `` (a person on a `PVTileCard`).

States: unchanged. Accessibility: unchanged (D-5 keeps the spoken label and the static focus). Identifiers stay at the call site (Family sets `persons.detail.family.tile.*`).

### Strings

D-9: `L10n.PersonCard.thisPerson`, `thisPersonSpoken`, `noLikeness` *(new namespace)*; removed `L10n.Conclusions.familyThisPersonSpoken` and `personNoLikeness`. Same copy; `scripts/check-localizable-xcstrings.py` holds the catalog.

## 7. Steps

**STEP-0 — Check S13-34's names**
Files: this plan.
Do: read S13-34's **Landed** note; if the names in § 4's "Display pieces" row differ, update § 4, D-5 and D-6 before code.
Check: the names match the code at the base.

**STEP-1 — Return and the focus ring in the kit**
Files: `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*, `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/App/Features/Conclusions/PersonTile.swift`, `PersonFamilySection.swift` (`PersonFamilyMoreRow`: the Return modifier only), `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (gallery tile), `macos/App/Features/Sources/SourcePageArtifactsView.swift`. Skill: `add-ui-component`.
Do: D-2, D-3.
Check: the app builds; `PersonTileTests` pass; `scripts/check-guidance-names.py` passes; AC-2.6 and AC-2.7 by hand.

**STEP-2 — The shell**
Files: `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVTileCardTests.swift` *(new)*. Skill: `add-ui-component`, `add-swift-test`.
Do: D-1, D-4.
Check: T-1 passes; the shell's previews render.

**STEP-3 — The gender treatment recipe**
Files: `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` *(new)*, `macos/App/Features/Conclusions/PersonGenderTreatment.swift` (deleted), `PersonDetailContent.swift`, `PersonTile.swift`, `macos/App/Platform/GenealogyStore.swift` (doc comment), `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` → `PVGenderTreatmentTests.swift`, README.
Do: D-7.
Check: T-4 and `PersonDetailTests` pass.

**STEP-4 — The person card at the regular size**
Files: `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)*, `macos/App/Features/Conclusions/PersonTile.swift` (deleted), `PersonFamilyContent.swift`, `PersonFamilySection.swift`, `PersonFamilyLayout.swift`, `PersonDetailContent.swift` (`noLikeness`), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonTileTests.swift` → `PVPersonCardTests.swift`, README. Skill: `add-localized-string`.
Do: D-5 (regular), D-9, D-10's test rename; Family's call sites take `PVPersonCard`, `PVPersonCardContent`, `PVTileCardSize.regular`.
Check: T-3, T-5, `PersonFamilyContentTests` pass; Family's previews (11a, 11c, 11d, 11h) look as before; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-5 — The compact size**
Files: `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift`, `macos/App/Platform/PersonLifeDisplay.swift` (moved there by S13-34), `macos/ProvenenciaTests/PVPersonCardTests.swift`. *(new)*
Do: D-6; a compact `#Preview`.
Check: T-2 passes.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `docs/ideas/structural-debt.md` (rows, next free A-numbers: `CitationLineButton`'s own ring; `SourcesGraphButtonBody`'s ring with hover).
Do: per the Definition of done.
Check: `scripts/check-pr.sh` passes (it runs the kit check); `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-35-person-card-joins-the-design-system.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVTileCardTests.swift` *(new)* `sizes` | `PVTileCardSize.regular` → height 92, thumbnail 70, title 2 lines, lines 2, minimum 120, padding 10; `.compact` → 52, 40, 1, 1, 232, 6 | D-1, D-4 |
| T-2 | Swift | `macos/ProvenenciaTests/PVPersonCardTests.swift` `compactYearsAndHelp` *(new)* | Anna 3 May 1790 Bremen – 1851 → `years` "1790 – 1851", `helpText` "Anna Weber\nb. 3 May 1790 · Bremen\nd. 1851"; birth only 1820 → "1820 –"; death only → "– 1851"; "abt. 1790" birth → "abt. 1790 – 1851"; a range birth 1790 – 1795 → "bet 1790 and 1795 – 1851"; a phrase-only birth → its phrase then " – 1851"; neither → "" and `helpText` the name alone | AC-1.2, D-5, D-6 |
| T-3 | Swift | same file `spokenLabels` (existing, renamed types) | Unchanged expectations: "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851"; the focus "…, this person"; plus role "Mother" on Anna → "Mother, Anna Weber, …" | AC-1.3, AC-2.4, D-5, D-9 |
| T-4 | Swift | `macos/ProvenenciaTests/PVGenderTreatmentTests.swift` (renamed) | Unchanged expectations on `PVGenderTreatment`; `PersonDetailContent.badge` unchanged | D-7 *(new)* |
| T-5 | Swift | `PersonFamilyLayoutTests` (existing) | Unchanged geometry with `PVTileCardSize.regular` in place of `PersonTile.height` and 120 | AC-2.1, D-4 |

AC-1.1 is S13-34's kit check, run by `check-pr.sh`. AC-2.1 – AC-2.7 are the manual check (§ 10) and T-3, T-5. Most likely to break: `PVPersonCardTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests`, `PVGenderTreatmentTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog, kit and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonTile`, `PersonTileContent`, `PersonTileButtonStyle`, `PersonTileTests`, `PersonGenderTreatment`, `PersonGenderTreatmentTests`, `familyThisPersonSpoken`, `personNoLikeness`, `conclusions.person.noLikeness` and `ConclusionFocusRingButtonStyle` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/**/*.md` (bar spike plan history), `macos/` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] UI matches today's Family, gallery, artifact preview and composer rows; no new kit primitive but `PVTileCard`, `pvActivatesOnReturn`, `PVFocusRingButtonStyle`
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] docs/ideas/structural-debt.md (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-2.1, AC-2.5) Open Johann's Family at Regular and Wide: frames 11a and 11b as S13-12 shipped them; a tile's tooltip is the name.
2. (AC-2.2, AC-2.3) Hover a relative's tile: it grows with a shadow. Tab to it: the ring. Return and Space open the relative's Family. Tab to a "Show N more" row and press Return: it unfolds.
3. (AC-2.4) VoiceOver on Anna's and Johann's tiles reads as before.
4. (AC-2.6, AC-2.7) Tab to a gallery tile and to a Source page artifact preview: the same ring. In the citation composer, Tab to an Observation row: Return and Space open it.
5. (AC-1.2) The `PVPersonCard` compact preview reads as AC-1.2.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The shell changes Family's layout by a point (spacing around an empty trailing slot) | D-1 draws the title row's `HStack` only with a trailing; T-5 and manual check 1 |
| Moving the gallery tile and artifact preview onto the style changes their ring | Same radius, same focus binding (D-3); manual check 4 |
| S13-34 lands with other names | STEP-0 |

**Stop and report if:**

- S13-34 has not landed, or it left `PersonTitleDisplay`, `PersonLifeDisplay` or the title and line views under `Features/`.
- S13-34's kit check fails on a type this PR moves and the card cannot do without it.
- Family's previews or `PersonFamilyLayoutTests` change beyond names.
- `@Environment(\.isFocused)` in `PVFocusRingButtonStyle` does not track a `.focused(…, equals:)` binding on the gallery tile or the artifact preview (keep their explicit ring then, and report).

## 12. Open questions

None.

## 13. PR

**Title:** S13-35 — The person card joins the design system
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-35-person-card-joins-the-design-system.md`.
