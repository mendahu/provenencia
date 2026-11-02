# S13-35 — The person card joins the design system

**Status:** Blocked — on S13-34
**Spike / slice:** Spike 13 · Slice 4 — Timeline (preparatory)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: Family looks, reads and behaves as before; the compact size is first drawn by S13-13, from board S13-D4 turn 12)
**Board:** — (the compact size's measures come from board S13-D4 turn 12, frame 12f, as the S13-13 plan records them)
**Depends on:** S13-34 (scoped, planned beside this one: the text pieces and value formatting move down a layer first)
**Blocks:** S13-13
**Branch:** `s13-35-person-card`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rule "The person card lives in the design system; display is decoupled from content" ([`../deployment-plan.md`](../deployment-plan.md#client)); [`docs/design-system-layers.md`](../../../design-system-layers.md); the evaluation of `PersonTile` (2026-11-02, plan items 3 and 4); the S13-12 plan [`S13-12-family-section.md`](S13-12-family-section.md) and its **Landed** note; the S13-13 plan [`S13-13-timeline-section.md`](S13-13-timeline-section.md) D-4, D-10, D-19 (taken over here); the S13-34 plan (beside this one)
**Planned:** 2026-11-02 against `main` @ `51dccd4`; STEP-0 re-checks S13-34's landed names
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

The person card becomes a design-system piece that any feature can place: a content-agnostic card shell (`PVTileCard`) with slots, sizes and its interaction, and on it a person card recipe (`PVPersonCard`) in a regular size (Family) and a compact size (the Timeline, next). A family tree in another feature will reuse it (user, 2026-11-02). The gender colour treatment becomes a recipe too, so the card reads it without reaching into the Person feature.

**Why now.** S13-13 places the compact card; building it on the final card avoids moving it twice (user, 2026-11-02: two preparatory PRs ahead of S13-13).

**What stays.** Family looks, reads and behaves exactly as today: tile size, focus tile, hover lift, focus ring, Return, tooltips and spoken labels. The Vitals header's gender badge. Every view outside Family.

## 2. User stories

**US-1.** As a contributor building a new view (a family tree), I place a person card from the design system without importing the Person feature.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | After this PR, `rg` over `macos/App/DesignSystem/` for the Features-only types this PR touched (`PersonTile`, `PersonGenderTreatment`, `ConclusionDetailBadge`, `PersonFamily`, `L10n.Conclusions.family`) | No hit (D-8) |
| AC-1.2 | A `PVPersonCard(content:projectDir:size: .compact)` preview with Anna (woman, likeness, 3 May 1790 Bremen – 1851) | 52pt card, 40pt likeness with the woman tint and line, "Anna Weber" on one line, "1790 – 1851" in mono caption muted; tooltip "Anna Weber\nb. 3 May 1790 · Bremen\nd. 1851" (D-5, D-6) |
| AC-1.3 | The same card's spoken label with role "Mother" | "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851" (unchanged rule) |

**US-2.** As a researcher, Family is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Johann's Family (S13-12 frame 11a) at Regular and Wide | Pixel-for-pixel as before: 92pt tiles, 70pt likeness, two-line names, *b.* and *d.* lines, the raised focus tile with "This person" |
| AC-2.2 | Hover a relative's tile; Tab to it; press Return; press Space | Grows 2.5% with the medium shadow; the focus ring; Return and Space open the relative's Family (D-2, D-3) |
| AC-2.3 | Tab to a "Show N more" row and press Return | It unfolds (D-2) |
| AC-2.4 | VoiceOver on Anna's and Johann's tiles | "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851", a button; "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person", static text |

## 3. Scope

| In | Covered by |
| --- | --- |
| `pvActivatesOnReturn` and `PVFocusRingButtonStyle` in the kit; the tile and Family's fold row move onto them (D-2, D-3) | STEP-1, AC-2.2, AC-2.3 |
| `PVTileCard` shell with regular, compact and row sizes (D-1, D-4) | STEP-2 |
| `PVGenderTreatment` recipe; the detail badge stays in the feature (D-7) | STEP-3 |
| `PVPersonCard` recipe at the regular size; Family moves onto it; `PersonTile.swift` deleted; its own L10n namespace (D-5, D-9) | STEP-4, AC-2.1, AC-2.4 |
| The compact size: life span at year precision and the tooltip (D-6) | STEP-5, AC-1.2, AC-1.3 |
| Nothing under `DesignSystem/` imports the Person feature; docs (D-8) | STEP-6, AC-1.1 |

| Out | Home |
| --- | --- |
| Moving the title styling, date · place line and value formatting down a layer | S13-34 |
| The Timeline, its event row and the participant grid | S13-13 (it places `PVPersonCard(size: .compact)` and builds the event row on `PVTileCard(size: .row)`) |
| Event and place cards (`PVEventCard`, `PVPlaceCard`) | when each has a second caller (evaluation follow-up 6); not planned |
| Moving `CachedThumbnail` out of `Platform/` | not planned (a `DesignSystem` → `Platform` dependency already exists: `PVArtifactChecklist`) |
| `PVOmnibarHitRow` onto `PVTileCard` or `PVTitleText` | evaluation follow-up 7; not planned |
| Family's own "This person" sibling group label (`L10n.Conclusions.familyThisPerson`) | stays Family's; only the card's badge and spoken suffix move (D-9) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Person tile | `macos/App/Features/Conclusions/PersonTile.swift` `PersonTileContent` (`init(summary:role:isFocus:locale:orders:)`, `id`, `title`, `lines`, `imageRelPath`, `colors`, `role`, `isFocus`, `spokenLabel`), `PersonTile` (`content`, `projectDir`, `open`; `nonisolated static let height: CGFloat = 92`), private `PersonTileButtonStyle`, `PersonTileButtonBody` | A feature snowflake: `PVCard` (raised and elevated for the focus), a 70pt `CachedThumbnail`, `ConclusionListRow.title(_, lineLimit: 2, emphasized:)`, `ConclusionDatePlaceLine(missing: .omitted)` per line, a `PVBadge` with `L10n.Conclusions.familyThisPerson` on the focus. A relative is a `Button` with a hand-written `@FocusState` + `.onKeyPress(.return)`, `.pvLiftHover(scale: 1.025)`, `.help(title)` |
| Family callers | `macos/App/Features/Conclusions/PersonFamilyContent.swift` `Tile.content: PersonTileContent` (lines 41, 110, 148); `PersonFamilySection.swift` `PersonFamilyDiagram` (line 133); `PersonFamilyLayout.swift` `PersonTile.height` (line 86) and the literal minimum 120 (line 89) | Four call sites |
| Fold row | `macos/App/Features/Conclusions/PersonFamilySection.swift` private `PersonFamilyMoreRow` | A ghost `PVButton` with the same hand-written Return handling |
| Gender treatment | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `PersonGenderTreatment` (`Family`, `Basis`, `text`, `fieldLabel`, `badge: ConclusionDetailBadge`, `thumbnailColors`, `read(_:)`); callers `PersonDetailContent.swift:15,36`, `PersonTile.swift:31`; tests `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` | Its colours are kit tokens; `badge` returns the feature's `ConclusionDetailBadge` (`ConclusionDetailPage.swift`) and words with `L10n.Conclusions.sexAtBirthBadge`; `text` reads `ReconciledValueDisplay.string(for:)` (S13-34 routes it through `PropertyTermDisplay`) |
| Display pieces (after S13-34) | S13-34 moves `ConclusionTitleSource` → `PVTitleSource` and `ConclusionListRow.title` → `PVTitleText` (`macos/App/DesignSystem/Components/TitleText/` *(new)* (S13-34)), `ConclusionDatePlaceLine` → `PVDatePlaceLine` (`macos/App/DesignSystem/Components/DatePlaceLine/` *(new)* (S13-34), its own missing-half option), and `PersonTitleDisplay`, `PersonLifeDisplay`, `DateRowDisplay` (gaining `Precision.year`), `NameValueDisplay` to `Platform/` | Not landed; STEP-0 checks the names |
| Kit pieces | `macos/App/DesignSystem/Components/Card/PVCard.swift` `PVCard` (`tone`, `cornerRadius`, `elevated`, `padding`); `macos/App/DesignSystem/Components/HoverEffect/PVLiftHover.swift` `pvLiftHover(scale:resting:)`; `macos/App/DesignSystem/Tokens/PVElevation.swift` `pvFocusRing(_:cornerRadius:)`; `Platform/CachedThumbnail.swift` `CachedThumbnail` | — |
| Recipes | `macos/App/DesignSystem/Recipes/` (`ArtifactChecklist` reads `CatalogArtifact` and `CachedThumbnail`; `DeleteImpact`; `Marks`; `OmnibarHitRow`; `LastUpdated`); `macos/App/DesignSystem/README.md` built table, one row per folder | Recipes may read `Catalog*` types and `Platform/`; no recipe reads `Features/` |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyThisPerson` (the card's badge and Family's sibling group label, `PersonFamilyContent.swift:159`), `familyThisPersonSpoken` (the card only) | — |
| Tests | `macos/ProvenenciaTests/PersonTileTests.swift` (`tileContentFromSummary`, `spokenLabels`), `PersonGenderTreatmentTests.swift`, `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift` | — |

**Precedent.** S13-12 D-16 moved the artifact card's private hover into the kit as `pvLiftHover` with a README row and both callers unchanged; copy that shape for the Return modifier and the focus-ring style. `PVArtifactChecklist` is the precedent for a recipe that reads `Catalog*` data and `CachedThumbnail`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The shell.** `struct PVTileCard<Thumbnail: View, Title: View, Lines: View, Trailing: View>: View` *(new)* in `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)*. `init(size: PVTileCardSize, isRaised: Bool = false, accessibilityLabel: String, help: String, action: (() -> Void)? = nil, @ViewBuilder thumbnail:, @ViewBuilder title:, @ViewBuilder lines:, @ViewBuilder trailing:)`, and a second init in `extension PVTileCard where Trailing == EmptyView` without `trailing`. Layout for `.regular` and `.compact`: a `PVCard(tone: isRaised ? .raised : .card, cornerRadius: PVRadius.sm, elevated: isRaised, padding: size.padding)` holding an `HStack(spacing: 10)` of the thumbnail at `size.thumbnail` and a `VStack(alignment: .leading, spacing: 3)` of an `HStack(spacing: PVSpacing.space4)` (title, then trailing with `.layoutPriority(1)`, the `HStack` drawn only when `Trailing != EmptyView`) and the lines; `.frame(height: size.height)`, `.clipped()`. `.row`: no `PVCard` surface, an `HStack(alignment: .top, spacing: 16)` of the thumbnail, the title and lines, and the trailing at the row's trailing edge; no fixed height. With `action` nil the card is static: `.accessibilityElement(children: .ignore)` with the label, no hover, no focus stop. With an action: `Button(action:)` with `PVFocusRingButtonStyle()`, `.pvActivatesOnReturn(action)`, `.pvLiftHover(scale: 1.025)` (not for `.row`), `.accessibilityLabel`. `.help(help)` on both | Content-agnostic: slots and tokens only, so it is a Component (`docs/design-system-layers.md`); one shell for person, event and place cards | A generic type with statics (Swift cannot read a static on an unspecialised generic: S13-32 review X-1); a recipe (it has no Provenencia meaning) |
| D-2 | **Return activation.** `extension View { func pvActivatesOnReturn(_ action: @escaping () -> Void) -> some View }` *(new)* in `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*: a private `ViewModifier` owning `@FocusState private var focused`, applying `.focused($focused)` and `.onKeyPress(.return) { guard focused else { return .ignored }; action(); return .handled }`, exactly today's `PersonTile` body. Space stays the `Button` default. `PVTileCard` and `PersonFamilyMoreRow` use it; S13-13 adds the event title and its fold row | Two callers now, a third in S13-13 (S13-13 round 1 M-4); content-agnostic | A third hand copy |
| D-3 | **Focus-ring style.** `struct PVFocusRingButtonStyle: ButtonStyle` *(new)* in `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` *(new)*, with a private body view reading `@Environment(\.isFocused)`: the label, `.contentShape(Rectangle())`, `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)`, no other chrome; today's `PersonTileButtonStyle` / `PersonTileButtonBody` moved. It replaces S13-13 D-19's `ConclusionFocusRingButtonStyle` | Content-agnostic, so the kit owns it (evaluation flag 7); one style for the tile and the event title | A feature style (S13-13 D-19 as written) |
| D-4 | **Sizes.** `enum PVTileCardSize: Equatable { case regular, compact, row }` *(new)*, beside `PVTileCard`, non-generic, with `nonisolated` properties: `height: CGFloat?` (92, 52, nil), `thumbnail: CGFloat` (70, 40, 72), `titleLineLimit: Int` (2, 1, 2), `minimumWidth: CGFloat?` (120, 232, nil), `padding: CGFloat` (10, 6, 0). `PersonFamilyLayout` reads `PVTileCardSize.regular.height!` and `.minimumWidth!` in place of `PersonTile.height` and its literal 120 | One place for every card measure (S13-13 round 2 M-21); compact from board S13-D4 frame 12f; the row for S13-13's event row (72pt picture, two-line title) | Measures on each recipe (the event row would repeat them) |
| D-5 | **The person card.** `Recipes/PersonCard/PVPersonCard.swift` *(new)*: `struct PVPersonCardContent: Equatable` (today's `PersonTileContent`, renamed and moved, same `init(summary:role:isFocus:locale:orders:)` and fields, reading S13-34's `Platform/` `PersonTitleDisplay` and `PersonLifeDisplay` and D-7's `PVGenderTreatment`), and `struct PVPersonCard: View` with `init(content: PVPersonCardContent, projectDir: String, size: Size = .regular, open: @escaping () -> Void = {})` and `enum Size { case regular, compact }` mapping to `PVTileCardSize`. It fills `PVTileCard`: thumbnail `CachedThumbnail(projectDir:relPath:size: size.thumbnail, cornerRadius: PVRadius.sm, fallbackMark: .subjectPerson, fallbackLabel: L10n.PersonCard.noLikeness, colors:)`; title `PVTitleText(content.title, lineLimit: size.titleLineLimit, emphasized: content.isFocus)` in `PVListText.titleFont`, `PVColor.textPrimary`; trailing the focus's `PVBadge(L10n.PersonCard.thisPerson)`; lines at regular one `PVDatePlaceLine(…, missing: .omitted)` per life line, at compact D-6's span; `isRaised: content.isFocus`; `action` nil for the focus, `open` otherwise; `accessibilityLabel: content.spokenLabel`; `help: content.helpText`. A compact card is never the focus | The owner's direction: a semantic card in the design system, reused by a family tree in another feature; a recipe reads `CatalogPersonSummary` as `PVArtifactChecklist` reads `CatalogArtifact` | Keeping the tile in the feature (user, 2026-11-02); one content type per size |
| D-6 | **Compact content.** `PersonLifeDisplay.years(_ summary: CatalogPersonSummary?, locale: Locale) -> String` *(new, in S13-34's `Platform/PersonLifeDisplay.swift`)*: the existing life span (`dateSpan` with `L10n.Conclusions.lifeSpan`, `lifeSpanOpen`, `lifeSpanDeathOnly`) at S13-34's `DateRowDisplay` `.year` precision: "1790 – 1851", "1820 –", "– 1851", "abt. 1790 – 1851", "" when neither date has a year (the line is then not drawn), in `PVFont.mono(size: PVTypeScale.caption)`, `PVColor.textMuted`. `PersonLifeDisplay.LifeLine.text` *(new)*: the abbreviation, a space, then date and place joined with `L10n.Conclusions.lineJoin`, a missing half left out ("b. 3 May 1790 · Bremen", "d. 1851"). `PVPersonCardContent.helpText` *(new)*: the title text, then each line's `text`, joined by newlines; both sizes use it as the tooltip (regular today shows the title only; the full lines are added, the tile itself unchanged). Taken over from S13-13 D-4 | The life span has one rule (S13-13 round 2 M-11; user, 2026-11-02); the board's tooltip | A raw `startYear` read; new keys |
| D-7 | **Gender treatment recipe.** `PersonGenderTreatment` moves to `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` *(new)* as `PVGenderTreatment`, with `Family`, `Basis`, `family`, `basis`, `text`, `fieldLabel`, `thumbnailColors`, `read(_ summary: CatalogPersonSummary?)`; its text reads S13-34's term rule (`PropertyTermDisplay` in `Platform/`). `badge` stays in the feature as `extension PVGenderTreatment { var badge: ConclusionDetailBadge }` in `macos/App/Features/Conclusions/PersonGenderTreatmentBadge.swift` *(new)*. `PersonDetailContent` and the card read `PVGenderTreatment`; `PersonGenderTreatmentTests` take the new name | Working rule "Colors reuse `PersonGenderTreatment`" (one treatment); a recipe because it reads `CatalogPersonSummary`; the badge is detail-page chrome (evaluation flag 4) | Leaving it in the feature (the card could not read it); moving the badge (it is `ConclusionDetailBadge`, a feature type) |
| D-8 | **Layering.** After STEP-4, no file under `macos/App/DesignSystem/` names a type declared under `macos/App/Features/`. STEP-6's check: for each `struct`/`enum`/`class` declared under `macos/App/Features/` (`rg -o '^(struct|enum|final class|class) \w+' macos/App/Features -r '$2'`), `rg -w` it under `macos/App/DesignSystem/` and expect no hit; `L10n.Conclusions.family` and `L10n.Conclusions.personNoLikeness` have no hit there either | The working rule: the design system imports no feature code | A manual look |
| D-9 | **Copy.** The card's strings move to `L10n.PersonCard` *(new namespace)*: `thisPerson` = "This person" (key `personCard.thisPerson`), `thisPersonSpoken` = "this person" (`personCard.thisPersonSpoken`), `noLikeness` (from `L10n.Conclusions.personNoLikeness`, `personCard.noLikeness`, same copy). `L10n.Conclusions.familyThisPersonSpoken` is deleted (the card was its only reader); `familyThisPerson` stays for Family's sibling group label; `personNoLikeness` stays if `rg` finds another caller, else is deleted. String Catalog entries move with the keys | A recipe owns its copy (`PVDeleteImpact` → `L10n.DeleteImpact`); a design-system piece names no feature (S13-13 round 1 M-8) | Keeping Family-named keys in the kit |

**Rules this plan bends.** None. It carries out the working rule "The person card lives in the design system; display is decoupled from content" (user, 2026-11-02), which replaced "one person tile, built once" in the planning commit.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVTileCard`, `PVTileCardSize` *(new)* | component | `PVCard`, `PVOmnibarHitRow`, `PersonTile` | `PVCard` is the surface it uses; `PVOmnibarHitRow` is a list row without a card, sizes or an action; the tile's layout becomes the shell |
| `pvActivatesOnReturn` *(new)* | component | `PersonTile`'s and `PersonFamilyMoreRow`'s hand-written Return | Extracted (D-2) |
| `PVFocusRingButtonStyle` *(new, moved)* | component | `PersonTileButtonStyle` | Moved (D-3) |
| `PVGenderTreatment` *(moved)* | recipe | `PersonGenderTreatment` | Moved (D-7) |
| `PVPersonCard`, `PVPersonCardContent` *(moved)* | recipe | `PersonTile`, `PersonTileContent` | Moved onto the shell (D-5) |
| `PersonLifeDisplay.years`, `LifeLine.text`, `PVPersonCardContent.helpText` *(new)* | func | `PersonLifeDisplay.dateSpan`, `lifeLines` | Extended (D-6) |
| `L10n.PersonCard` *(new)* | keys | `L10n.Conclusions.family*` | Moved (D-9) |

**Builds on.** S13-34's `PVTitleText`, `PVTitleSource`, `PVDatePlaceLine`, `DateRowDisplay` precision and `Platform/` formatters; `PVCard`, `pvLiftHover`, `pvFocusRing`, `PVBadge`, `CachedThumbnail`.
**Deletes.** `macos/App/Features/Conclusions/PersonTile.swift` (`PersonTile`, `PersonTileContent`, `PersonTileButtonStyle`, `PersonTileButtonBody`); `macos/App/Features/Conclusions/PersonGenderTreatment.swift` (moved); the hand-written Return in `PersonFamilyMoreRow`; `L10n.Conclusions.familyThisPersonSpoken`; `PersonTile.height` and `PersonFamilyLayout`'s literal 120.
**Next adds.** S13-13 places `PVPersonCard(size: .compact)` and builds its event row on `PVTileCard(size: .row)`; a family tree places `PVPersonCard` from its own feature; `PVEventCard` and `PVPlaceCard` fill the shell when each has a second caller.

### Reach

None — no new path into existing code: Family's tiles keep their destinations, and the compact size has no caller until S13-13.

### Scale

None — no write, list or loop changes; one card per tile as today.

### Edges

None — no rule, comparison or algorithm changes. The year-precision span is S13-34's rule; `years` only calls it.

### Schema

None — no catalog change.

### Go core

None.

### FFI

None.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Card shell | `PVTileCard`, `PVTileCardSize` | `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)* | D-1, D-4; `#Preview` with a static, a raised and an actionable card at each size, text slots only (no Person words in kit previews) |
| Return | `pvActivatesOnReturn` | `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)* | D-2 |
| Focus ring | `PVFocusRingButtonStyle` | `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` *(new)* | D-3 |
| Gender treatment | `PVGenderTreatment` | `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` *(new)* | D-7 |
| Person card | `PVPersonCard`, `PVPersonCardContent` | `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)* | D-5, D-6; `#Preview` at both sizes |

README rows in the built table, in the existing format: `| TileCard | \`Components/TileCard/PVTileCard.swift\` (…) |`, `ReturnActivation`, `FocusRingButton` among the Components; `GenderTreatment`, `PersonCard` among the Recipes.

States: unchanged. Accessibility: unchanged (D-5 keeps the spoken label and the static focus).

### Strings

D-9: `L10n.PersonCard.thisPerson`, `thisPersonSpoken`, `noLikeness` *(new namespace)*; removed `L10n.Conclusions.familyThisPersonSpoken` (and `personNoLikeness` if no other caller). Same copy; `scripts/check-localizable-xcstrings.py` holds the catalog.

## 7. Steps

**STEP-0 — Check S13-34's names**
Files: this plan.
Do: read S13-34's **Landed** note; if `PVTitleText`, `PVTitleSource`, `PVDatePlaceLine` (its missing option), `DateRowDisplay`'s precision, or the `Platform/` homes of `PersonTitleDisplay` and `PersonLifeDisplay` differ from § 4, update § 4, D-5 and D-6 before code.
Check: the names match the code at the base.

**STEP-1 — Return and the focus ring in the kit**
Files: `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*, `macos/App/DesignSystem/Components/FocusRingButton/PVFocusRingButtonStyle.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/App/Features/Conclusions/PersonTile.swift`, `PersonFamilySection.swift` (`PersonFamilyMoreRow`). Skill: `add-ui-component`.
Do: D-2, D-3; the tile and the fold row use them.
Check: the app builds; `PersonTileTests` pass; `scripts/check-guidance-names.py` passes.

**STEP-2 — The shell**
Files: `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVTileCardTests.swift` *(new)*. Skill: `add-ui-component`, `add-swift-test`.
Do: D-1, D-4.
Check: T-1 passes; the shell's previews render.

**STEP-3 — The gender treatment recipe**
Files: `macos/App/DesignSystem/Recipes/GenderTreatment/PVGenderTreatment.swift` *(new)*, `macos/App/Features/Conclusions/PersonGenderTreatmentBadge.swift` *(new)*, `PersonGenderTreatment.swift` (deleted), `PersonDetailContent.swift`, `PersonTile.swift`, `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift`, README.
Do: D-7.
Check: `PersonGenderTreatmentTests` and `PersonDetailTests` pass unchanged but for the name.

**STEP-4 — The person card at the regular size**
Files: `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)*, `PersonTile.swift` (deleted), `PersonFamilyContent.swift`, `PersonFamilySection.swift`, `PersonFamilyLayout.swift`, `PersonFamilyPreviewData.swift` if it names the tile, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonTileTests.swift` (renamed types), README. Skill: `add-localized-string`.
Do: D-5 (regular), D-9; Family's call sites take `PVPersonCard`, `PVPersonCardContent`, `PVTileCardSize.regular`.
Check: `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass; Family's previews (11a, 11c, 11d, 11h) look as before.

**STEP-5 — The compact size**
Files: `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` *(new)* (STEP-4), `macos/App/Platform/PersonLifeDisplay.swift` *(new)* (moved there by S13-34), `macos/ProvenenciaTests/PersonTileTests.swift`.
Do: D-6; a compact `#Preview`.
Check: T-2, T-3 pass.

**STEP-6 — Layering check and docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `docs/design-system-layers.md` (one line: a recipe may read `Platform/` value formatters and `Catalog*` types; `PVPersonCard` as the example).
Do: D-8's check; per the Definition of done.
Check: D-8's `rg` loop prints nothing; `scripts/check-pr.sh` passes; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-35-person-card-joins-the-design-system.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVTileCardTests.swift` *(new)* `sizes` | `PVTileCardSize.regular` → height 92, thumbnail 70, 2 lines, minimum 120, padding 10; `.compact` → 52, 40, 1, 232, 6; `.row` → nil, 72, 2, nil, 0 | D-1, D-4 |
| T-2 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` `compactYearsAndHelp` *(new)* | Anna 3 May 1790 Bremen – 1851 → `years` "1790 – 1851", `helpText` "Anna Weber\nb. 3 May 1790 · Bremen\nd. 1851"; birth only 1820 → "1820 –"; death only → "– 1851"; "abt. 1790" birth → "abt. 1790 – 1851"; neither → "" | AC-1.2, D-6 |
| T-3 | Swift | same file `spokenLabels` (existing, renamed types) | Unchanged expectations: "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851"; the focus "…, this person"; plus role "Mother" on Anna → "Mother, Anna Weber, …" | AC-1.3, AC-2.4, D-5, D-9 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonGenderTreatmentTests.swift` (renamed type) | Unchanged expectations on `PVGenderTreatment`; the badge from the feature extension unchanged | D-7 |
| T-5 | Swift | `PersonFamilyLayoutTests` (existing) | Unchanged geometry with `PVTileCardSize.regular` in place of `PersonTile.height` and 120 | AC-2.1, D-4 |

AC-1.1 is STEP-6's D-8 check. AC-2.1 – AC-2.4 are the manual check (§ 10) and T-3, T-5. Most likely to break: `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests`, `PersonGenderTreatmentTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonTile`, `PersonTileContent`, `PersonTileButtonStyle`, `PersonGenderTreatment` (the type, not the working rule's name), `familyThisPersonSpoken` and `ConclusionFocusRingButtonStyle` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] D-8's layering check prints nothing
- [ ] UI matches today's Family; no new kit primitive but `PVTileCard`, `pvActivatesOnReturn`, `PVFocusRingButtonStyle`
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] docs/design-system-layers.md: the recipe line (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-2.1) Open Johann's Family at Regular and Wide: frames 11a and 11b as S13-12 shipped them.
2. (AC-2.2, AC-2.3) Hover a relative's tile: it grows with a shadow. Tab to it: the ring. Return and Space open the relative's Family. Tab to a "Show N more" row and press Return: it unfolds.
3. (AC-2.4) VoiceOver on Anna's and Johann's tiles reads as before.
4. (AC-1.2) The `PVPersonCard` compact preview reads as AC-1.2.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The shell changes Family's layout by a point (spacing around an empty trailing slot) | D-1 draws the title `HStack` only with a trailing; T-5 and manual check 1 |
| S13-34 lands with other names | STEP-0 |
| A recipe reading `Platform/` reads as a layer break | `PVArtifactChecklist` already does; STEP-6 records it in `docs/design-system-layers.md` |

**Stop and report if:**

- S13-34 has not landed, or it left `PersonTitleDisplay`, `PersonLifeDisplay` or the title and line views under `Features/`.
- D-8's check finds a `Features/` type the card cannot do without.
- Family's previews or `PersonFamilyLayoutTests` change beyond names.

## 12. Open questions

None.

## 13. PR

**Title:** S13-35 — The person card joins the design system
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-35-person-card-joins-the-design-system.md`.
