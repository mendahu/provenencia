# S13-32 — The person tile becomes a kit shell

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 4 — Timeline (preparatory)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: Family's tiles and fold row look, read and behave as before)
**Board:** —
**Depends on:** S13-12 (landed [#676](https://github.com/mendahu/provenencia/pull/676))
**Blocks:** S13-13
**Branch:** `s13-32-person-tile-kit-shell`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rule "One person tile, built once; the shell in the kit, the content at the call site" (user, 2026-11-02); [`docs/design-system-layers.md`](../../../design-system-layers.md) (§ 1 Components, § Layering rules: lower layers do not import higher ones); the S13-12 plan [`S13-12-family-section.md`](S13-12-family-section.md) (D-5, D-16) and its **Landed** note; the S13-13 plan [`S13-13-timeline-section.md`](S13-13-timeline-section.md) (D-4, D-5, D-9, D-10, D-11) and its review log [`S13-13-timeline-section.review.md`](S13-13-timeline-section.review.md) (M-1, M-2, M-4, M-8)
**Planned:** 2026-11-02 against `main` @ `51dccd4`
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

The person tile's chrome (a card with a thumbnail beside a title, a trailing mark and lines; the focus ring, hover lift and Return activation of a tile that opens something) becomes a content-agnostic design-system component, and the person content that fills it (which name, which life lines, which colors, "This person") stays in `Features/Conclusions`, at the call site. Family looks and behaves exactly as before. It prepares S13-13, whose Timeline reuses the tile, and answers the S13-13 review: the tile could not move into `DesignSystem/` as it was, because it reads six Conclusions display types (review M-1; user, 2026-11-02: "The components that should be promoted into the design system should be content-agnostic … the content should live at the call sites").

**Why now.** S13-13 is the tile's second caller and is blocked on this PR.

**What stays.** Family's look, layout, ids, spoken labels and tests; `PersonTileContent`'s fields and initializer; `PersonGenderTreatment` and its colors; `ConclusionListRow`, `ConclusionDatePlaceLine`, `PersonTitleDisplay`, `PersonLifeDisplay` and the date and name formatters, which stay where they are.

## 2. User stories

**US-1.** As a researcher, the Family section is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann's Family (S13-12 frame 11a) | Pixel-identical to `main`: tiles 92pt tall, 70pt likeness, title, b. and d. lines, the focus raised with "This person" (D-2, D-3) |
| AC-1.2 | Hover a relative's tile; Tab to it; press Return | It grows 2.5% with the medium shadow; the kit focus ring; Return opens the relative's Family (D-2, D-4) |
| AC-1.3 | Tab to a "Show N more" row, press Return | The list unfolds (D-4) |
| AC-1.4 | VoiceOver on Anna's tile; on the focus | "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851", a button; "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person", static text (D-5) |

**US-2.** As a contributor, I can put a tile on another view by passing its content, and name any term value with one rule.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `rg` over `macos/App/DesignSystem/` | No `Person*`, `Conclusion*`, `Catalog*`, `CachedThumbnail` or `L10n.Conclusions` symbol (D-1, D-2) |
| AC-2.2 | `PropertyTermDisplay.name(.term(id: "t", key: "witness", label: "Witness", origin: CatalogOrigin.provenencia), propertyKey: SeededPropertyKey.role)` | "Witness"; a user term labelled "Godparent" → "Godparent"; a user term with label "  " and key `godparent` → "godparent"; `.text("x")` → nil (D-6) |
| AC-2.3 | `PVTile.height`, `PVTile.minimumWidth`, `PVTile.thumbnailSize` | 92, 240, 70 (D-2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVTile`, a content-agnostic tile shell in `macos/App/DesignSystem/Components/Tile/PVTile.swift` *(new)* with thumbnail, title, trailing and lines slots, its sizes and its interaction (D-1 – D-3) | STEP-2, AC-1.1, AC-1.2, AC-2.1, AC-2.3 |
| `.pvActivatesOnReturn(_:)`, one kit modifier for Return on a focused button; the tile and `PersonFamilyMoreRow` use it (D-4) | STEP-1, AC-1.2, AC-1.3 |
| `PersonTile` becomes the person content placed in `PVTile`, in `Features/Conclusions` (D-5) | STEP-3, AC-1.1, AC-1.4 |
| The tile's "this person" keys lose their Family name (D-7) | STEP-3, AC-1.4 |
| `PropertyTermDisplay.name(_ value:propertyKey:)` (D-6) | STEP-4, AC-2.2 |

| Out | Home |
| --- | --- |
| The Timeline's use of the tile | S13-13 |
| Moving `PersonFamilyMoreRow` out of `PersonFamilySection.swift` and renaming `familyShowMore` | S13-13 (its D-9) |
| `PersonGenderTreatment` reading through `PropertyTermDisplay` instead of `ReconciledValueDisplay.string(for:)` | not this PR: its badge text must match the Vitals row, which reads `ReconciledValueDisplay`; both move together when the drift is fixed, logged by S13-13 STEP-7 in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (D-6) |
| The gender colors (`PersonGenderTreatment.Family` tint, line, ink) as a kit treatment | not planned: the shell takes `PVThumbnail.Colors`, already a kit type; the colors move when a kit view needs the mapping |
| `PersonTitleDisplay`, `PersonLifeDisplay`, `ConclusionListRow`, `ConclusionDatePlaceLine`, `DateRowDisplay`, `NameValueDisplay` into `DesignSystem/` or `Platform/` | not planned (map scope entry): the content builder calls them where they are |
| `CitationComposerSubjectGroups`'s Return handling (`macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift`) | not planned: a focusable row that is not a `Button`, with Space handling and a shared `FocusState` enum; a different shape from D-4 |
| A `PVTile` without a thumbnail, or other card sizes | not planned: one caller shape |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Tile | `macos/App/Features/Conclusions/PersonTile.swift` `PersonTile`, `PersonTileContent`, private `PersonTileButtonStyle`, `PersonTileButtonBody` | `PersonTile.height` = 92. The card: `PVCard(tone: isFocus ? .raised : .card, cornerRadius: PVRadius.sm, elevated: isFocus, padding: 10)` holding an `HStack(spacing: 10)` of `CachedThumbnail(size: 70, cornerRadius: PVRadius.sm, fallbackMark: .subjectPerson, fallbackLabel: L10n.Conclusions.personNoLikeness, colors:)` and a `VStack(alignment: .leading, spacing: 3)`: an `HStack(spacing: PVSpacing.space4)` of `ConclusionListRow.title(_, lineLimit: 2, emphasized: isFocus)` (`.font(PVListText.titleFont)`, `PVColor.textPrimary`) and, for the focus, `PVBadge(L10n.Conclusions.familyThisPerson).layoutPriority(1)`; then a `ConclusionDatePlaceLine(missing: .omitted)` per line. The VStack `.frame(maxWidth: .infinity, alignment: .leading)`, the HStack `.frame(maxWidth: .infinity, minHeight: 70, maxHeight: 70, alignment: .leading)`, the card `.frame(height: 92).clipped()`. The focus: `.accessibilityElement(children: .ignore)`, label, `.help`. A relative: `Button(action: open)` with `PersonTileButtonStyle` (label `.contentShape(Rectangle())` + `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)` from `@Environment(\.isFocused)`), `.focused($focused)`, `.onKeyPress(.return) { guard focused … open() }`, `.pvLiftHover(scale: 1.025)`, label, `.help` |
| Tile callers | `macos/App/Features/Conclusions/PersonFamilySection.swift` `PersonFamilyDiagram` (`PersonTile(content:projectDir:open:)`); `PersonFamilyLayout.swift` `init` (`PersonTile.height`); `PersonFamilyContent.swift` (`PersonTileContent`) | — |
| Fold row | `PersonFamilySection.swift` private `PersonFamilyMoreRow` | `Button` `.buttonStyle(.pv(.ghost, size: .sm))`, `.focused($focused)`, `.onKeyPress(.return) { guard focused … expand() }` |
| Keys | `macos/App/Platform/L10n.swift` `L10n.Conclusions.familyThisPerson` (`conclusions.family.thisPerson`, "This person"), `familyThisPersonSpoken` (`conclusions.family.thisPersonSpoken`, "this person") | Read by `PersonTile.swift` (badge, spoken label) and `PersonFamilyContent.swift` (the siblings group label when the focus stands alone) |
| Term names | `macos/App/Platform/PropertyTermDisplay.swift` `name(key:origin:propertyKey:catalogLabel:)`, `name(term:propertyKey:)` | Product L10n first, then the trimmed catalog label, else the key. No overload for `CatalogConclusionValue` |
| Older term rule | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` `.term` | `label.isEmpty ? key : label`, no product names; read by Vitals rows and `PersonGenderTreatment` |
| Kit pieces | `macos/App/DesignSystem/Components/Card/PVCard.swift` `PVCard`; `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.Colors`; `macos/App/DesignSystem/Components/HoverEffect/PVLiftHover.swift` `.pvLiftHover(scale:resting:)`; `macos/App/DesignSystem/Tokens/PVElevation.swift` `pvFocusRing(_:cornerRadius:)` | Content-agnostic |
| Kit README | `macos/App/DesignSystem/README.md` built table | One row per `Components/<Name>/` folder (`scripts/check-guidance-names.py`) |
| Tests | `macos/ProvenenciaTests/PersonTileTests.swift` (`tileContentFromSummary`, `spokenLabels`), `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`, `PropertyTermDisplayTests.swift` | — |

**Precedent.** S13-12 D-16 ([#676](https://github.com/mendahu/provenencia/pull/676)) moved the artifact card's private hover into the kit as `.pvLiftHover`, content-agnostic, with the artifact card looking as before: copy that shape for D-2 and D-4 (a private feature mechanism → a kit piece with one README row, callers unchanged in look).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The shell is `PVTile` *(new)* in `macos/App/DesignSystem/Components/Tile/PVTile.swift` *(new)*: a component, not a recipe, because it carries no product meaning. It imports only kit types (`PVCard`, `PVRadius`, `PVSpacing`, `PVBadge` is not used inside it, `pvFocusRing`, `pvLiftHover`, `pvActivatesOnReturn`) | User, 2026-11-02: kit pieces are content-agnostic shells with slots; `design-system-layers.md` § 1 | `Recipes/PersonTile/PVPersonTile` (would import Conclusions types: review M-1); keeping all of it in `Features/` (the Timeline and later pages need the same chrome) |
| D-2 | `struct PVTile<Thumbnail: View, Title: View, Trailing: View, Lines: View>: View` with `init(emphasized: Bool = false, accessibilityLabel: String, help: String, action: (() -> Void)? = nil, @ViewBuilder thumbnail:, @ViewBuilder title:, @ViewBuilder trailing: = { EmptyView() }, @ViewBuilder lines:)`. Statics: `height: CGFloat = 92`, `thumbnailSize: CGFloat = 70`, `minimumWidth: CGFloat = 240` (the narrowest width where a two-line title and a date · place line stay readable; Family does not read it, S13-13 D-11 does). Layout, exactly today's: `PVCard(tone: emphasized ? .raised : .card, cornerRadius: PVRadius.sm, elevated: emphasized, padding: 10)` { `HStack(spacing: 10)` { thumbnail; `VStack(alignment: .leading, spacing: 3)` { `HStack(spacing: PVSpacing.space4)` { title; trailing `.layoutPriority(1)` }; lines } `.frame(maxWidth: .infinity, alignment: .leading)` } `.frame(maxWidth: .infinity, minHeight: thumbnailSize, maxHeight: thumbnailSize, alignment: .leading)` } `.frame(height: height).clipped()` | Family pixel-identical (AC-1.1); the sizes have one owner, on the tile (S13-13 review M-5) | Slots for a whole content area (each caller would rebuild the title row); fonts inside the shell (the caller's title slot sets them, as today) |
| D-3 | Interaction in `PVTile`: with `action` nil it is static: `.accessibilityElement(children: .ignore)`, `.accessibilityLabel`, `.help`. With an action it is `Button(action:)` with a private `PVTileButtonStyle` (the label with `.contentShape(Rectangle())` and `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)` from `@Environment(\.isFocused)`, as `PersonTileButtonBody`), `.pvActivatesOnReturn(action)`, `.pvLiftHover(scale: 1.025)`, `.accessibilityLabel`, `.help`. `emphasized` and `action` are independent (Family's focus passes `emphasized: true, action: nil`) | One owner for a tile that opens something (S13-13 review M-4); today's behavior unchanged | A separate `PVTileButton` (two types for one card) |
| D-4 | `View.pvActivatesOnReturn(_ action: @escaping () -> Void) -> some View` *(new)* in `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*: a `ViewModifier` with its own `@FocusState private var focused`, applying `.focused($focused)` and `.onKeyPress(.return) { guard focused else { return .ignored }; action(); return .handled }`. Doc: "A focused macOS Button takes Space, not Return; this adds Return. Put it on the Button, after its style." `PVTile` and `PersonFamilyMoreRow` use it and drop their own `@FocusState` and `onKeyPress` | Two hand copies today, a third in S13-13 (the event title): extract on the third (review M-4) | Leaving the copies; folding it into `PVHoverEffect` (a style body cannot hold the focus binding the key press needs) |
| D-5 | The person content stays in `macos/App/Features/Conclusions/PersonTile.swift`: `PersonTileContent` is unchanged (fields, `init(summary:role:isFocus:locale:orders:)`, `spokenLabel`), and `PersonTile(content:projectDir:open:)` keeps its signature and becomes the composition: `PVTile(emphasized: content.isFocus, accessibilityLabel: content.spokenLabel, help: content.title.text, action: content.isFocus ? nil : open)` with thumbnail `CachedThumbnail(projectDir:relPath: content.imageRelPath, size: PVTile.thumbnailSize, cornerRadius: PVRadius.sm, fallbackMark: .subjectPerson, fallbackLabel: L10n.Conclusions.personNoLikeness, colors: content.colors)`, title `ConclusionListRow.title(content.title, lineLimit: 2, emphasized: content.isFocus).font(PVListText.titleFont).foregroundStyle(PVColor.textPrimary)`, trailing the focus's `PVBadge(L10n.Conclusions.thisPerson)`, lines one `ConclusionDatePlaceLine(leading: line.abbreviation, date: line.date, place: line.place, accessibilityLabel: line.accessibilityLabel, missing: .omitted)` per line, `ForEach(content.lines, id: \.abbreviation)`. `PersonTile.height`, `PersonTileButtonStyle` and `PersonTileButtonBody` are deleted; `PersonFamilyLayout` reads `PVTile.height`. S13-13 calls `PersonTile` and `PersonTileContent` as they are | The content at the call site, in the feature that owns it (user, 2026-11-02); Family's three call sites barely move; S13-13 D-4 uses these names | A `Person*` type in the kit; a new content type (today's is already pure and caller-agnostic: the role is passed in) |
| D-6 | `PropertyTermDisplay.name(_ value: CatalogConclusionValue, propertyKey: String) -> String?` *(new)* beside `name(term:propertyKey:)`: `.term(_, key, label, origin)` → `name(key: key, origin: origin, propertyKey: propertyKey, catalogLabel: label)`; every other case nil. S13-13 D-5 calls it for roles. `PersonGenderTreatment` does not move onto it here: its `text` must equal the Vitals row's value, which `ReconciledValueDisplay.string(for:)` words (label, else key, no product names); moving one without the other makes the badge and the row disagree for a relabelled product term | One term-name rule for term values (S13-13 review M-2), in `Platform/`, the repository-level home the user named for value resolvers | Moving `PersonGenderTreatment` alone (the map's scope entry said so; research showed the mismatch, § 12) |
| D-7 | `L10n.Conclusions.familyThisPerson` → `thisPerson` (key `conclusions.thisPerson`, comment "Badge on a person's own tile, and the label of that tile alone in a group"), `familyThisPersonSpoken` → `thisPersonSpoken` (key `conclusions.thisPersonSpoken`, comment "VoiceOver ending of a person's own tile, mid-sentence"). Same copy; the String Catalog entries move with them; `PersonTile.swift` and `PersonFamilyContent.swift` read the new names | The tile's copy is no longer Family's (S13-13 review M-8) | Keeping the Family names on shared copy |

**Rules this plan bends.** None. The working rule "One person tile, built once; the shell in the kit, the content at the call site" was rewritten with the user on 2026-11-02 before this plan; D-6 departs from the map's scope entry (it says `PersonGenderTreatment` moves off `ReconciledValueDisplay.string(for:)`), listed in § 12 for the map to be corrected.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVTile` *(new)* | component | `PVCard`, `PersonTile`, `PVOmnibarHitRow` | `PVCard` is the surface it composes; `PersonTile` holds person content; the hit row is a list row |
| `pvActivatesOnReturn` *(new)* | modifier | `PVHoverEffect`, the two hand copies | No kit piece handles Return; a style body cannot own the focus binding |
| `PropertyTermDisplay.name(_:propertyKey:)` *(new)* | func | `name(term:propertyKey:)` | An overload for the value type the reads carry |
| `L10n.Conclusions.thisPerson`, `thisPersonSpoken` *(renamed)* | keys | `familyThisPerson`, `familyThisPersonSpoken` | Renamed (D-7) |

**Builds on.** `PVCard`, `PVThumbnail.Colors`, `pvLiftHover`, `pvFocusRing`, `PropertyTermDisplay`, `CachedThumbnail`, `ConclusionListRow.title`, `ConclusionDatePlaceLine`; the S13-12 D-16 move as precedent.
**Deletes.** `PersonTile.height`, `PersonTileButtonStyle`, `PersonTileButtonBody` (into `PVTile`); the `@FocusState` and `.onKeyPress(.return)` in `PersonTile` and `PersonFamilyMoreRow` (into `pvActivatesOnReturn`); keys `conclusions.family.thisPerson`, `conclusions.family.thisPersonSpoken` (renamed).
**Next adds.** S13-13: participant tiles through `PersonTile` / `PersonTileContent` (its D-4), their grid at `PVTile.minimumWidth` (D-11), the event title and the moved fold row on `pvActivatesOnReturn` (D-9, D-10), roles through `PropertyTermDisplay.name(_:propertyKey:)` (D-5). Any page's tile of a Source, Event or Place fills `PVTile` with its own content.

### Reach

None — no new path into existing code, no new state reachable: the same views draw the same content.

### Scale

None — no write, list or loop changes: one `PVTile` per tile as today.

### Edges

- **Assumptions.** `@FocusState` inside a `ViewModifier` binds to the view it modifies, as the inline copies did on the `Button`; T-4 and the manual check hold it (Return on a tile and a fold row).
- **Neighbors.** `action: nil` with `emphasized: false` → a plain static card (no caller today; allowed); a term value with an empty label → its key (D-6, as every term name).
- **Removed inputs.** None — no input is decided or deleted.
- **Real data.** None — no table or seed drives the change.
- **Untrusted size.** None — no new list.

### Schema

None — no catalog change.

### Go core

None.

### FFI

None.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Tile shell | `PVTile` *(new)* | `macos/App/DesignSystem/Components/Tile/PVTile.swift` *(new)* | D-1 – D-3; a `#Preview` with neutral content (a symbol thumbnail, "Title", two text lines), static, emphasized and actionable |
| Return | `pvActivatesOnReturn` *(new)* | `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)* | D-4 |
| Person tile | `PersonTile`, `PersonTileContent` | `macos/App/Features/Conclusions/PersonTile.swift` | D-5 |
| README | rows `Tile`, `ReturnActivation` | `macos/App/DesignSystem/README.md` | "content-agnostic card: a thumbnail beside a title, a trailing mark and lines; 92pt tall, `minimumWidth` 240; static or a button with the focus ring, lift hover and Return"; "Return on a focused Button (macOS Buttons take Space only)" |

States: unchanged. Accessibility: unchanged (D-3 keeps the label, help and element shape).

### Strings

Renamed (D-7): `familyThisPerson` → `thisPerson`, `familyThisPersonSpoken` → `thisPersonSpoken`, copy unchanged, entries moved in `macos/App/Resources/Localizable.xcstrings`. New: none.

## 7. Steps

**STEP-1 — Return activation in the kit**
Files: `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` *(new)*, `macos/App/Features/Conclusions/PersonFamilySection.swift` (`PersonFamilyMoreRow`), `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`.
Do: D-4; `PersonFamilyMoreRow` uses it.
Check: the app builds; `scripts/check-guidance-names.py` passes; Return on a "Show N more" row unfolds (manual).

**STEP-2 — The tile shell**
Files: `macos/App/DesignSystem/Components/Tile/PVTile.swift` *(new)*, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PVTileTests.swift` *(new)*. Skill: `add-ui-component`, `add-swift-test`.
Do: D-1 – D-3, with its `#Preview`.
Check: T-3 passes; `rg -n "Person|Conclusion|Catalog|CachedThumbnail|L10n\.Conclusions" macos/App/DesignSystem/Components/Tile macos/App/DesignSystem/Components/ReturnActivation` prints nothing.

**STEP-3 — Person content in the shell; keys renamed**
Files: `macos/App/Features/Conclusions/PersonTile.swift`, `PersonFamilyLayout.swift`, `PersonFamilyContent.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonTileTests.swift` (only if it names a renamed key). Skill: `add-localized-string`.
Do: D-5, D-7.
Check: `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests` pass unchanged; Family's `#Preview`s "11a Full family", "11c Conflicting parents", "11d Recorded siblings only", "11h After clicking Anna" match `main` side by side; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-4 — One term-name rule for term values**
Files: `macos/App/Platform/PropertyTermDisplay.swift`, `macos/ProvenenciaTests/PropertyTermDisplayTests.swift`. Skill: `add-swift-test`.
Do: D-6.
Check: T-1 passes.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row; the S13-32 scope entry's `PersonGenderTreatment` clause per D-6), this plan (status, Landed note), `docs/design-system-layers.md` (§ 1: a card shell takes its content through slots; the tile is the example).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-32-person-tile-kit-shell.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PropertyTermDisplayTests.swift` `nameOfAConclusionValue` *(new)* | `.term` product `witness` on `role` → "Witness"; user term "Godparent" → "Godparent"; user term label "  ", key `godparent` → "godparent"; product key with a user origin → its own label; `.text("x")`, `.none`, `.integer(1)` → nil | AC-2.2, D-6 |
| T-2 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` `tileContentFromSummary`, `spokenLabels` (unchanged) | As today: the focus's spoken label ends "this person"; a relative's starts with its role | AC-1.4, D-5, D-7 |
| T-3 | Swift | `macos/ProvenenciaTests/PVTileTests.swift` *(new)* `sizes` | `PVTile.height` 92, `thumbnailSize` 70, `minimumWidth` 240 | AC-2.3, D-2 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonFamilyLayoutTests.swift` (unchanged) | Every measure as today, reading `PVTile.height` | AC-1.1, D-5 |

AC-1.1, AC-1.2, AC-1.3 and AC-2.1 are also covered by the manual check (§ 10) and STEP-2's `rg`. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PersonTileTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests` (STEP-3).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonTileButtonStyle`, `PersonTileButtonBody`, `PersonTile.height`, `familyThisPerson`, `familyThisPersonSpoken`, `conclusions.family.thisPerson` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory — no brief; Family pixel-identical to main (manual check 1)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; the S13-32 scope entry's PersonGenderTreatment clause corrected (D-6); completed.md: one short paragraph of what shipped
- [ ] docs/design-system-layers.md: the slot sentence (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-1.1) Build `main` and this branch; open Johann's Family (or Family's `#Preview`s 11a, 11c, 11d, 11h) in light and dark: the two match pixel for pixel.
2. (AC-1.2) Hover a relative: the lift; Tab to it: the focus ring; Return: the relative's Family opens; Space does the same.
3. (AC-1.3) On a family with more than 30 siblings, Tab to "Show N more" and press Return: the list unfolds.
4. (AC-1.4) VoiceOver on Anna's tile and on the focus: as AC-1.4.
5. (AC-2.1) STEP-2's `rg` prints nothing.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `@FocusState` in a modifier does not reach the `Button` the way the inline copy did | Manual checks 2 and 3; if Return stops working, stop (below) |
| Generic slots change layout by a point (an `EmptyView` trailing slot adds no spacing in an `HStack`) | Manual check 1 compares pixels |

**Stop and report if:**

- `PersonTile.swift`, `PersonFamilyMoreRow` or the two keys differ from § 4.
- Return no longer activates a tile or a fold row with `pvActivatesOnReturn` (D-4).
- Family is not pixel-identical after STEP-3 (manual check 1).
- `PVTile` needs a non-kit import to build.

## 12. Open questions

None. `PersonGenderTreatment` stays on `ReconciledValueDisplay.string(for:)` (D-6); the map's scope entry was corrected in the planning commit (2026-11-02).

## 13. PR

**Title:** S13-32 — The person tile becomes a kit shell
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-32-person-tile-kit-shell.md`.
