# S13-34 — Display pieces move down a layer

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 4 — Timeline (preparatory)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a move of display code between layers. Two visible changes the user approved on 2026-11-02 and that draw nothing new: a relabelled built-in term reads with the app's name wherever the app names it, and the Events list's place takes the Person header's small secondary style)
**Board:** —
**Depends on:** —
**Blocks:** S13-35 (the card shell and the person card read the moved pieces), S13-13
**Branch:** `s13-34-display-pieces-move-down`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. One Go test constant changes (a moved Swift path, STEP-4); no proto change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** S13-35 plan [`S13-35-person-card-joins-the-design-system.md`](S13-35-person-card-joins-the-design-system.md) (reads every name this PR lands); working rule "The person card lives in the design system; display is decoupled from content" ([`../deployment-plan.md`](../deployment-plan.md#client)); [`docs/design-system-layers.md`](../../../design-system-layers.md) (lower layers do not import higher ones; recipes may read catalog types); the `PersonTile` evaluation of 2026-11-02 (its plan items 1 and 2); the S13-13 plan [`S13-13-timeline-section.md`](S13-13-timeline-section.md), whose term rule and year precision move here; precedent S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)), which moved the artifact card's hover into the kit as `pvLiftHover`
**Planned:** 2026-11-02 against `main` @ `04a2c99`
**Review:** [`S13-34-display-pieces-move-down.review.md`](S13-34-display-pieces-move-down.review.md) — round 1 Fits with changes / Not ready; every finding Fixed but M-5 (Declined with evidence) and M-14 (Logged); the user's three answers of 2026-11-02 applied; none open

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

The pieces that word and style a person, a date, a place or a term move out of the Person, Names and Dates feature folders, so the design system can build a person card (S13-35) that a family tree in another feature will reuse, without the design system reaching into feature code. Title styling (name plain, label italic, ref mono; list and display styles) and the date · place line become design-system components; the formatting of names, dates, places, a person's title and life span moves to `Platform/`; one rule names a term value everywhere; and a permanent check keeps the design system free of feature types.

**Why now.** The user chose to hoist the person card into the design system (2026-11-02); every piece it is built from lives in `Features/` today. This PR is the first of two preparatory PRs (S13-34, then S13-35) before S13-13.

**What changes for the researcher** (user, 2026-11-02). (1) A built-in term the researcher relabelled (the gender `woman` relabelled "Female (self-described)") reads with the app's name ("Woman") everywhere the app names a term value: the Vitals rows, the header's gender badge, the provenance sheet's "Read as" line, Evidence graph cards (and Promote's Subject cards, which draw them), and the bridge sentences. Pickers already show the app's name. (2) The Events list's second line uses the Person header's date · place line: the place is small, italic and in secondary ink (it was body size in the row's ink), the date and the "+N" badge as before.

**What stays.** Every other list, page, tile and line looks the same; every type keeps its behavior; editors keep their draft-only functions in their features.

## 2. User stories

**US-1.** As a researcher, a term reads the same everywhere the app names it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The built-in gender term `woman` relabelled "Female (self-described)" on a person | The Vitals gender row, the header's gender badge, the gender records' "Read as" lines in the provenance sheet, and the gender Observation on the Evidence graph card all read "Woman" (D-2, D-3) |
| AC-1.2 | The researcher's own gender term "Two-spirit" | All four read "Two-spirit", unchanged |
| AC-1.3 | A role term value `witness` (product origin) through `PropertyTermDisplay.name(_:propertyKey: SeededPropertyKey.role)` | "Witness"; a `.text` value → nil (D-1) |
| AC-1.4 | A product term with an empty catalog label | Its product name; a researcher's term with an empty label → its key; a term with an empty key and label → nil (D-1) |
| AC-1.5 | A participation bridge whose role is the built-in `witness` relabelled "Witness at the ceremony" | The bridge sentence reads "Witness" (D-3) |

**US-2.** As a contributor building the person card or a family tree, every display piece the card needs is outside `Features/`, and CI says so.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `scripts/check-guidance-names.py` after this PR | Passes; its kit check finds no `Features/` type named in code under `macos/App/DesignSystem/` but the allowed `WorkspaceLocation` (D-12) |
| AC-2.2 | The files this PR moves to `macos/App/Platform/` | None names a type declared under `macos/App/Features/` (STEP-7's one-off `rg`, D-12) |
| AC-2.3 | The Persons and Places lists, the Person header, Family, the Event page and the detail page title | They look and read as before (existing tests; previews) |
| AC-2.4 | `DateRowDisplay.line(date:precision: .year)` | 3 May 1790 → "1790"; "abt. 3 May 1790" → "abt. 1790"; "bef. 1790" → "bef. 1790"; a range 1790 – 1795 → "bet 1790 and 1795"; a range 1790 – 1790 → "1790"; "May 1790" → "1790"; a date with a month and no year → ""; a phrase-only date → its phrase; `.full` unchanged (D-9) |
| AC-2.5 | Anna (birth 3 May 1790, death 1851) through `PersonLifeDisplay.dateText` at `.year` | "1790"; the Persons list's line at `.full` unchanged ("3 May 1790 – 1851") (D-9) |

**US-3.** As a researcher, an event's date and place read the same in the Events list as on a Person's header.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | An Events row for a marriage on 4 Jun 1844 at Bremen with two other kept places | The date in mono caption, a faint "·", "Bremen" in small italic secondary ink, a "+2" subtle Badge; VoiceOver reads the row as before (D-5) |
| AC-3.2 | An Events row with a date and no place; with neither | The date alone; no second line (D-5) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One term-name rule: `PropertyTermDisplay.name(_:propertyKey:)`; Vitals rows, "Read as" lines, the gender badge, Evidence graph cards and bridge sentences read it (S13-13's term rule, moved here) | STEP-1, AC-1.1 – AC-1.5 |
| `PVTitleSource` and `PVTitleText` (list and display styles) in the design system; the detail page title uses it; `PVBadge.moreCount` | STEP-2, AC-2.3 |
| `PVDatePlaceLine` with its `Missing`, a date size, an extra count and the spoken halves; the Events list's line moves onto it; the copy in `L10n.DesignSystem` | STEP-3, AC-2.3, AC-3.1, AC-3.2 |
| Name formatting to `Platform/`; the Go parity test's path | STEP-4, AC-2.2 |
| Date formatting to `Platform/`, with a year precision (S13-13's, moved here) | STEP-5, AC-2.2, AC-2.4 |
| Place, person-title and life formatting to `Platform/`; the life span takes the precision | STEP-6, AC-2.2, AC-2.5 |
| A permanent kit check in `scripts/check-guidance-names.py` | STEP-7, AC-2.1 |

| Out | Home |
| --- | --- |
| The card shell, the Return modifier, the focus-ring button style, the gender treatment and the person card in the design system; `PersonLifeDisplay.years`, `LifeLine.text`, the card's tooltip | S13-35 |
| `EventTitleDisplay`, `EventSecondaryDisplay`, `PlaceTitleDisplay`, `ConclusionSummaryDisplay` to `Platform/`; event and place cards | when each card has a second caller (map working rule "The person card lives in the design system; display is decoupled from content") |
| `PVOmnibarHitRow` to `macos/App/DesignSystem/Components/` with `PVTitleText` | a follow-up; logged by STEP-8 |
| Splitting `DateValueDraft`'s three editor flags (`showStartTime`, `showEndTime`, `showAdvanced`) into the editor's own state | logged by STEP-8 (review M-5, Declined with evidence) |
| Unifying `DateRowDisplay` and `DateValueDisplay` behind one style and precision | logged by STEP-8 |
| `L10n.Properties.boundOverflow` ("+%lld") onto `PVBadge.moreCount` | logged by STEP-8 |
| `WorkspaceLocation` out of `PVDeleteImpact` | logged by STEP-8; allowed by STEP-7's check meanwhile |
| Moving `CachedThumbnail` out of `Platform/` | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Term names | `macos/App/Platform/PropertyTermDisplay.swift` `name(key:origin:propertyKey:catalogLabel:)`, `name(term:propertyKey:)` | Product name for a product term, else the trimmed label, else the key. No overload for a `CatalogConclusionValue` |
| Old term rule | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` (line 84 – 105), `readAs(_:locale:orders:)` (line 113), `readAsIsPhrase(_:)` (line 125) | `.term(_, key, label, _)` → `label.isEmpty ? key : label`: no product names. Callers of `string(for:)`: `ReconciledValueRowModel.swift:60` (in `init(field:)`) and `:110` (in `static func otherValue(_:locale:orders:)`, whose callers are lines 73, 138, 172 and 205: `init`, `valueGroups` (has `field`), `pooledValueGroups` (has `member.field`), `listedValues(of: field)`), `PlaceTitleDisplay.swift:25`, `PersonGenderTreatment.swift:73`, and `readAs` / `readAsIsPhrase`, whose caller is `ReconciliationRecord.swift:38 – 39` (`init(outcome:in field:…)`). Tests `ReconciledValueDisplayTests.swift:84, 151 – 159, 170, 198, 199, 229, 230` |
| Observation values | `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift` `string(for:locale:orders:)` | A term returns `valueText`, the catalog label (`core/database/observations/observations.go:754`); no product names. `CatalogObservation` (`Platform/GenealogyStore.swift:868 – 874`) carries `propertyKey`, `valueTermKey`, `valueTermOrigin`. Callers: `EvidenceCardView.swift:480, 661`, `EvidenceGraphSnapshot.swift:82`, `EvidenceBridgeEdgeSummary.swift:197`. Tests `ObservationValueDisplayTests.swift` |
| Bridge term names | `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` private `termLabel(of:propertyKey:)` (line 196) | Re-derives the product name over `ObservationValueDisplay` with `ConnectionSentenceRegistry.termProperty(for: kind)` (line 193): a second copy of the rule. Tests `EvidenceBridgeEdgeSummaryTests.swift`, `ConnectionSentenceRegistryTests.swift` |
| Gender treatment | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `read(_:)`, `treatment(_:basis:)` (line 71 – 75) | Words the term through `ReconciledValueDisplay.string(for:)`, `guard !text.isEmpty`; `basis` is `.gender` or `.sexAtBirth`. `PersonGenderTreatmentTests.swift:70` expects nil for a term with empty key and label |
| Title styling | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionTitleSource` (line 6), `ConclusionListRow.title(_:extraCount:lineLimit:emphasized:)` (line 45), private `titleText(_:)`, `italicizingEmphasis(_:font:)` (line 84) | Content-agnostic but in a feature; draws the count with `L10n.Conclusions.moreCount` (line 57). `ConclusionTitleSource` is named in 13 files (23 lines), all under `macos/App/`; `title` in `PersonTile.swift:99` and `ConclusionListPage.swift:183`; `italicizingEmphasis` in `PersonTitleDisplayTests.swift:122` |
| Detail title | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` private `title`, `nameText(_:styled:)`, `titleFont`, `titleSize` (line 771 – 802) | The same three-style switch at a display size (`PVFont.display`, `h3` compact / `display3`, tracking `PVTracking.display`, `PVColor.textDisplay`, ref mono `h4` / `h1`, `.isHeader`), calling `ConclusionListRow.italicizingEmphasis` |
| Date · place line | `macos/App/Features/Conclusions/ConclusionDatePlaceLine.swift` `ConclusionDatePlaceLine` | Date mono caption; place `bodySmall` italic `textSecondary`; typed on `PersonLifeDisplay.Missing`; reads `L10n.Conclusions.dateUnknown` / `placeUnknown`. Callers: `PersonTile.swift:108`, `PersonDetailView.swift:81`, `EventDetailView.swift:31` |
| Events list line | `macos/App/Features/Conclusions/EventsListView.swift` `EventSecondaryLine` (line 92 – 121) | A copy of the line: date mono caption with `fixedSize`, faint "·", place italic at the row's size and ink, a `+N` `PVBadge` with `L10n.Conclusions.moreCount`. Its spoken text is the row's (`EventsList.accessibilityLabel`) |
| Count badge | `PVBadge(text: L10n.Conclusions.moreCount(n), tone: .neutral, subtle: true).accessibilityHidden(true)` | Three copies: `ConclusionListRow.swift:57`, `PersonsListView.swift:71`, `EventsListView.swift:116` |
| Name formatting | `macos/App/Features/Names/NameValueDisplay.swift` (`NameFormatPart`, `NameFormatOrders`, `NameDisplayStyle`, `NameValueDisplay`, `NameDisplay`); `NamePartType.swift` (`NamePartDecoration`, `NamePartType`) | `NameValueDisplay.string(for draft: NameValueDraft)` (line 33) and `NamePartType.selectOptions` (`PVSelectOption`, line 52) are editor-only. `api/ffi/swift_vocabulary_test.go:51` reads `NamePartType.swift` by path for the name-part parity check |
| Date formatting | `macos/App/Features/Dates/DateRowDisplay.swift` `line(date:start:end:includesPhrase:locale:)`, private `rowDate` (reads through `DateValueDraft(from:)`), `applyQualifier`, `civil`, `appendPhrase`; `DateValueDisplay.swift` (`string(for draft:)` holds the wording, `string(for input:)` builds a draft, line 58 – 59; qualifier literals `"ABT"`, `"BEF"`, `"AFT"`); `DateValueDraft.swift` (Foundation and `L10n` only; `showStartTime`, `showEndTime`, `showAdvanced` set from the input at line 330 – 332 and read by `DateValueEditorForm.swift`) | `DateQualifier` (`Platform/ConclusionVocabulary.swift:248`) names the qualifiers. `.agents/skills/add-date-value/SKILL.md:72` links `macos/App/Features/Dates/DateValueDisplay.swift` |
| Places | `macos/App/Features/Conclusions/PlaceChainDisplay.swift`, `DerivedPlace.swift` | Formatting only (`L10n.Conclusions`) |
| Person title and life | `macos/App/Features/Conclusions/PersonTitleDisplay.swift` (returns `ConclusionTitleSource`); `PersonLifeDisplay.swift` (`Line`, `LifeLine`, `Missing`, `line(_:locale:)`, `lifeLines(_:locale:missing:)`, private `lifeLine` (reads `dateUnknown` / `placeUnknown`, line 79 – 80), `dateText`, private `dateSpan`, `placeSpan`) | `dateSpan` is the Persons list's life span (keys `conclusions.life.span*`); no year-only form |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Conclusions.moreCount` (key `conclusions.list.more`, line 8966), `dateUnknown` (`conclusions.detail.dateUnknown`, line 9538), `placeUnknown` (`conclusions.detail.placeUnknown`, line 9544); `macos/App/Resources/Localizable.xcstrings` (lines 17061, 17073) | `dateUnknown` / `placeUnknown` also read by `EventDetailView.swift:35 – 36` |
| Kit README and checks | `macos/App/DesignSystem/README.md` built table; `scripts/check-guidance-names.py` (`KIT_LAYERS`, `kit_rows_missing`, `allowed()`), `scripts/check-guidance-names.allow` | One row per kit folder; no check that kit code names no feature type. Today `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` takes `WorkspaceLocation` (declared in `Features/Workspace/WorkspaceLocation.swift:83`) in code (lines 152, 174, 247, 355, 398, 454); doc comments and the README name other feature types |
| Layer docs | `docs/design-system-layers.md`; `docs/macos-client-patterns.md:14` (`Platform/`: "store protocol, FFI, fakes, `L10n`"); `.cursor/rules/design-system-layers.mdc`; `.agents/skills/add-ui-component/SKILL.md` checklist | None says where value formatting lives or that kit code names no feature type |
| Tests | `macos/ProvenenciaTests/PropertyTermDisplayTests.swift`, `ReconciledValueDisplayTests.swift`, `ObservationValueDisplayTests.swift`, `EvidenceBridgeEdgeSummaryTests.swift`, `PersonGenderTreatmentTests.swift`, `DateRowDisplayTests.swift`, `DateValueDisplayTests.swift`, `DateValueDraftTests.swift`, `DerivedPlaceTests.swift`, `PersonTitleDisplayTests.swift`, `PersonDetailTests.swift`, `PersonsListTests.swift` (the Events list's label, line 166), `PlaceTitleDisplayTests.swift`, `EventTitleDisplayTests.swift`, `NameValueDraftTests.swift`, `NameFormatOrdersFixtures.swift` | — |

The app is one Swift module, so a move between folders changes no compile-time import; the layers are a rule of the repo ([`docs/design-system-layers.md`](../../../design-system-layers.md)), enforced from STEP-7 on by `scripts/check-guidance-names.py`. The Xcode target uses synchronized folders (`objectVersion 77`): a moved file needs no project edit.

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) moved `ArtifactCardHover` from `Features/Sources` into `DesignSystem/Components/HoverEffect/PVLiftHover.swift` with a README row and no visible change. Copy its shape: `git mv`, rename, README row, callers updated in the same commit.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PropertyTermDisplay.name(_ value: CatalogConclusionValue, propertyKey: String) -> String?` *(new)* in `macos/App/Platform/PropertyTermDisplay.swift`: `.term(_, key, label, origin)` → `name(key:origin:propertyKey:catalogLabel: label)`, nil when that is empty; any other case → nil | One rule names a term value (S13-13's role badges, S13-35's gender treatment, D-2, D-3) | A timeline-only role type |
| D-2 | One rule for conclusion values (user, 2026-11-02). `ReconciledValueDisplay.string(for:propertyKey:locale:orders:)` gains a required `propertyKey: String`; its `.term` case returns `PropertyTermDisplay.name(value, propertyKey:) ?? ""`; every other case unchanged. `readAs(_:propertyKey:locale:orders:)` and `readAsIsPhrase(_:propertyKey:)` gain it too; `ReconciliationRecord.init` passes `field.propertyKey`. `ReconciledValueRowModel`: line 60 passes `field.propertyKey`; `otherValue(_:propertyKey:locale:orders:)` gains the key, passed by its four callers: line 73 and 138 `field.propertyKey`, line 172 `member.field.propertyKey`, line 205 `field.propertyKey`. `PlaceTitleDisplay` passes the unwrapped `toponym.propertyKey` inside its map. `PersonGenderTreatment.treatment` calls `PropertyTermDisplay.name(value, propertyKey:)` with `SeededPropertyKey.gender` or `.sexAtBirth` by `basis`; `guard let text, !text.isEmpty` keeps `PersonGenderTreatmentTests.swift:70` nil. Tests pass a key per fixture: `ReconciledValueDisplayTests.swift:154 – 155` pass `SeededPropertyKey.sexAtBirth`, so line 154 stays "Male" and line 155's expectation changes from "male" to "Male" (a product term with an empty label now reads its product name); every other fixture line passes `"occupation"` (no product terms, wording unchanged) | Two rules drift; the badge must match its row; S13-35 moves the treatment into the design system, where it may not call `ReconciledValueDisplay` | A default `propertyKey` (a caller could forget it); `string(for:in field:)` (the gender treatment has no field) |
| D-3 | Evidence graph values (user, 2026-11-02: term names everywhere). `ObservationValueDisplay.string(for:locale:orders:)`, after the name and date branches: when `observation.valueTermKey` is not empty, return `PropertyTermDisplay.name(key: observation.valueTermKey, origin: observation.valueTermOrigin, propertyKey: observation.propertyKey, catalogLabel: valueText)`; otherwise `valueText` as today. `EvidenceBridgeEdgeSummary.termLabel(of:propertyKey:)` is deleted; `termDisplay` returns `ObservationValueDisplay.string(for: observation)` trimmed, nil when empty. Promote draws its Subject cards with `EvidenceCardView`, so it follows | One rule (review M-2); the observation carries all four inputs; the bridge's private copy goes | Patching each caller; a Vitals-only change (the user chose everywhere) |
| D-4 | `ConclusionTitleSource` → `PVTitleSource` (same cases `.name(String, styled:spoken:)`, `.label`, `.ref`; same `text`, `spokenText`); `ConclusionListRow.title` and `titleText` → `struct PVTitleText: View` *(moved)*: `init(_ source: PVTitleSource, style: Style = .row, extraCount: Int = 0, lineLimit: Int = 1, emphasized: Bool = false)`, `enum Style { case row, display(compact: Bool) }`. `.row` is today's list title. `.display(compact:)` is the detail page title: `PVFont.display(size: compact ? h3 : display3, weight: .medium)`, tracking `size · PVTracking.display`, `PVColor.textDisplay`, the label italic, the ref `PVFont.mono(size: compact ? h4 : h1, weight: .medium)`, `.accessibilityAddTraits(.isHeader)`, the spoken text as today. `italicizingEmphasis` becomes private to `PVTitleText`. `ConclusionDetailPage`'s `title`, `nameText`, `titleFont` and `titleSize` go; it draws `PVTitleText(content.title, style: .display(compact: compact))`. Home `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)*. `ConclusionListRow` keeps `thumbnail`, `updatedColumn` and `accessibilityLabel` | Content-agnostic, so a Component; one owner of "name plain, label italic, ref mono" at both sizes (review M-7) | A public `italicizingEmphasis` the page reaches into; renaming `.name` |
| D-5 | `ConclusionDatePlaceLine` → `struct PVDatePlaceLine: View` *(moved)* in `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)*: `init(leading: String? = nil, date: String?, place: String?, accessibilityLabel: String, missing: Missing = .stated, dateSize: CGFloat = PVTypeScale.caption, extraCount: Int = 0)`, today's body, the date at `dateSize`, then `PVBadge.moreCount(extraCount)` when `extraCount > 0`. `enum Missing { case stated, omitted }` (a policy value, not view state) replaces `PersonLifeDisplay.Missing`. `static func spokenDate(_ date: String?) -> String` and `spokenPlace(_:)` *(new)* return the half or the unknown copy; `PersonLifeDisplay.lifeLine` and `EventDetailView.EventDetailSecondary` build their labels with them. `EventSecondaryLine` becomes a call: `PVDatePlaceLine(date: date.isEmpty ? nil : date, place: place.isEmpty ? nil : place, accessibilityLabel: EventSecondaryDisplay.line(…)'s text, missing: .omitted, extraCount: extraPlaces)`, `.accessibilityHidden(true)` (the row's label already speaks it); with neither half it draws nothing, as today (user, 2026-11-02: the list's place takes the header's style) | One date · place line (review M-1); S13-13 reads `dateSize: PVTypeScale.body`; one owner of the missing-half wording, drawn and spoken (review M-8) | A follow-up (an Out row cannot defer a second mechanism); keeping the list's style (the user chose the header's) |
| D-6 | `extension PVBadge { static func moreCount(_ count: Int) -> some View }` *(new)* in `macos/App/DesignSystem/Components/Badge/PVBadge.swift`: `PVBadge(text: L10n.DesignSystem.badgeMoreCount(count), tone: .neutral, subtle: true).accessibilityHidden(true)`. `PVTitleText`, `PVDatePlaceLine` and `PersonsListView.PersonLifeLine` use it | Three copies today (review M-6); one owner of tone, copy and accessibility | Three hand-built Badges |
| D-7 | Copy keys follow `designSystem.<component>.<name>` (review M-12): `L10n.DesignSystem.badgeMoreCount(_:)` (key `designSystem.badge.moreCount`, from `conclusions.list.more`), `datePlaceLineDateUnknown` (`designSystem.datePlaceLine.dateUnknown`, from `conclusions.detail.dateUnknown`), `datePlaceLinePlaceUnknown` (`designSystem.datePlaceLine.placeUnknown`, from `conclusions.detail.placeUnknown`), same English copy; the three old keys and catalog entries go. `badgeMoreCount` moves in STEP-2 (its first kit reader), the other two in STEP-3 | A kit component reads no feature copy; the namespace convention (108 of 109 keys) | Two keys with one copy |
| D-8 | Name formatting moves to `macos/App/Platform/`: `NameValueDisplay.swift` (with `NameFormatPart`, `NameFormatOrders`, `NameDisplayStyle`, `NameDisplay`) and `NamePartType.swift` (with `NamePartDecoration`). Editor-only members stay in `Features/Names` as extensions: `NameValueDisplay.string(for draft: NameValueDraft)` in `NameValueDraft.swift`, `NamePartType.selectOptions` in `NameValueEditorForm.swift`; `NameValueDisplay.string(form:)` moves (formatting). `api/ffi/swift_vocabulary_test.go:51` `swiftNamePartType` becomes `"../../macos/App/Platform/NamePartType.swift"` | Names are read across features: formatting, not a feature; the Go parity test reads the file by path (review M-4) | Moving `NameValueDraft` |
| D-9 | Date formatting moves to `macos/App/Platform/`: `DateRowDisplay.swift`, `DateValueDisplay.swift` and `DateValueDraft.swift`. The qualifier literals in the moved files become `DateQualifier.about` / `.before` / `.after`. `DateRowDisplay.line` gains `precision: Precision = .full` *(new)*, `enum Precision { case full, year }`. At `.year`: a point passes only its year to `civil` with its qualifier ("abt. 1790", "bef. 1790"); a month or day with no year → "" (no year to show; not the full wording); a range shows each side's year ("bet 1790 and 1795"), the same year on both sides collapses to that year ("1790"), a range with no years → "" as today; a phrase-only date → its phrase, through `appendPhrase` with an empty structured part; `includesPhrase` as at `.full`. In `PersonLifeDisplay`, `dateText(_:locale:precision:)` and the private `dateSpan(birth:death:locale:precision:)` gain `precision: DateRowDisplay.Precision = .full`; `line(_:locale:)` passes `.full`. `.agents/skills/add-date-value/SKILL.md:72` links the new path. S13-35 adds `years`, `LifeLine.text` and the tooltip on top | One rule for a life span at two precisions (user, 2026-11-02); `DateValueDisplay`'s wording runs on the draft (`DateValueDisplay.swift:6, 58 – 59`), so the draft moves with it (review M-5, Declined: its editor flags are set from the input in `DateValueDraft(from:)` and toggled inside the draft, so splitting them is an editor refactor, logged) | A raw `startYear` read; rewriting both display types on `CatalogDateValueInput` |
| D-10 | Places and the person's title and life move to `macos/App/Platform/`: `PlaceChainDisplay.swift`, `DerivedPlace.swift`, `PersonTitleDisplay.swift`, `PersonLifeDisplay.swift`; their `L10n.Conclusions` keys stay. `Platform/` formatters return or take two kit value types, `PVTitleSource` and `PVDatePlaceLine.Missing` (as `Platform/CachedThumbnail.swift` draws `PVThumbnail`) | Formatting the card, the lists, the header and a tree all need | Moving their keys |
| D-11 | README rows in `macos/App/DesignSystem/README.md`'s built table for `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)* and `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)*, in alphabetical place, in the table's format | `scripts/check-guidance-names.py` requires one per folder | — |
| D-12 | A permanent kit check in `scripts/check-guidance-names.py` (review M-9; S13-35 drops its own): `kit_names_feature_types()` *(new)*, run by `main()`. It harvests, at run time from the current tree, every type declared under `macos/App/Features/` with `^\s*(?:(?:public|private|fileprivate|internal) )?(?:final )?(?:enum|struct|class|protocol|typealias|actor) (\w+)`, minus every name also declared under `macos/App/DesignSystem/` or `macos/App/Platform/` and minus SwiftUI names (`Text`); then scans the Swift code under `macos/App/DesignSystem/` with comments stripped (a line from `//` on) and the README not read, and fails on any harvested name matched as a word, unless it is in `scripts/check-guidance-names.allow`. The allow file gains `WorkspaceLocation   # PVDeleteImpact navigates to a delete's referrers; logged as structural debt (S13-34)`. If the check finds another hit at the base, the implementer stops and reports it. STEP-7 also runs a one-off `rg` of the harvested names over the files this PR moved to `Platform/` (AC-2.2), with the same exclusions | The rule "the design system imports no feature code" enforced on every PR; one regex in one place | A one-off command per PR (two regexes that disagree); scanning comments |
| D-13 | Docs: one sentence in `docs/design-system-layers.md` (S13-35 adds none): value formatting (names, dates, terms, places, a person's title and life span) lives in `Platform/`, content-agnostic display in `DesignSystem/Components/`, so a recipe composes both and never names a `Features/` type; the kit check enforces it. `docs/macos-client-patterns.md:14`'s `Platform/` row adds "value formatting". One line each in `.cursor/rules/design-system-layers.mdc` and the `add-ui-component` checklist: kit code names no `Features/` type; recipes may read `Platform/` formatters and `Catalog*` types | One home for the rule (review M-10, S13-35 M-5) | Two sentences in two PRs |

**Rules this plan bends.** None. The term-name and Events-line changes are the user's answers (2026-11-02), recorded in the map's working rule and here.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PropertyTermDisplay.name(_:propertyKey:)` *(new overload)* | func | `name(term:propertyKey:)` | Same rule, one more input type |
| `PVTitleSource`, `PVTitleText` (`Style`) *(moved, extended)* | component | `ConclusionTitleSource`, `ConclusionListRow.title`, the detail page's `title` | Two copies of one switch become one |
| `PVDatePlaceLine` (`Missing`, `dateSize`, `extraCount`, spoken halves) *(moved, extended)* | component | `ConclusionDatePlaceLine`, `EventSecondaryLine` | Two lines become one |
| `PVBadge.moreCount` *(new)* | component | three hand-built count Badges | One owner |
| `L10n.DesignSystem.badgeMoreCount`, `datePlaceLineDateUnknown`, `datePlaceLinePlaceUnknown` *(moved keys)* | keys | `L10n.Conclusions.*` | The component's copy in the kit's namespace |
| `DateRowDisplay.Precision` *(new)*; `precision:` on `DateRowDisplay.line`, `PersonLifeDisplay.dateText` and `dateSpan` | enum, parameter | `DateRowDisplay.line` | Extended |
| `kit_names_feature_types()` *(new)* | check | `kit_rows_missing()` | Same script, same `KIT_LAYERS` |
| Name, date, place and person formatting in `Platform/` *(moved)* | files | — | Moves |

**Builds on.** `PropertyTermDisplay`, the kit's tokens, `PVBadge`, `PVListText`, `L10n.DesignSystem`, `DateQualifier`, `scripts/check-guidance-names.py`, the S13-12 move precedent.
**Deletes.** `ConclusionTitleSource`, `ConclusionListRow.title`, `ConclusionListRow.italicizingEmphasis`, the detail page's `title` / `nameText` / `titleFont` / `titleSize`, `ConclusionDatePlaceLine`, `EventSecondaryLine`'s body, `PersonLifeDisplay.Missing`; `EvidenceBridgeEdgeSummary.termLabel`; `ReconciledValueDisplay`'s own term rule; the three hand-built count Badges; `L10n.Conclusions.moreCount`, `dateUnknown`, `placeUnknown` and their catalog entries.
**Next adds.** S13-35 builds `PVTileCard` and the person card on `PVTitleText`, `PVDatePlaceLine` and the `Platform/` formatting, and adds `years`, `LifeLine.text` and the tooltip on the precision; S13-13 reads `PropertyTermDisplay.name(_:propertyKey:)` for roles and `PVDatePlaceLine(…, dateSize: PVTypeScale.body)` for its event rows; event and place cards later move their own formatting the same way.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A relabelled product term on any term surface | Vitals, the badge, "Read as", Evidence graph cards, bridge sentences | Reads the product name (AC-1.1, AC-1.5; T-2, T-3) |
| A term observation with an empty key | `ObservationValueDisplay` | `valueText` as today (T-3) |
| An Events row with no date or place | `PVDatePlaceLine` | Draws nothing, as `EventSecondaryLine` did (AC-3.2) |

### Scale

None — no write, list or loop changes; the moves are code locations, D-9's precision adds one branch per date, and the kit check reads the Swift sources once per `check-pr` run.

### Edges

- **Assumptions.** `PropertyTermDisplay` returns the label, else the key; D-1 maps an empty result to nil, so the gender guard and `PersonGenderTreatmentTests.swift:70` hold. `PlaceTitleDisplay`'s toponym values are names, not terms (`PlaceTitleDisplayTests` unchanged). A bridge's term observation carries the same `propertyKey` as `ConnectionSentenceRegistry.termProperty(for:)` (T-3 asserts it for every bridge kind; a mismatch is a stop condition).
- **Neighbors.** A product term relabelled → product name (changes); a researcher's term → its label (unchanged); a product term with no product resource → its label (unchanged); a non-term value → unchanged.
- **Removed inputs.** None.
- **Real data.** The product term resources (`L10n.PropertyTerm.resource`) cover gender, sex at birth, event type, role and relationship type; any other seeded term keeps its catalog label.
- **Untrusted size.** None.

### Schema

None — no catalog change.

### Go core

None. One Go test constant changes path (D-8).

### FFI

None.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Title styling | `PVTitleSource`, `PVTitleText` | `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)* | D-4 |
| Date · place line | `PVDatePlaceLine` | `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)* | D-5 |
| Count badge | `PVBadge.moreCount` | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` | D-6 |
| Formatting | moved files | `macos/App/Platform/` | D-8 – D-10 |

No view looks different but the Events list's second line (D-5) and the term names (D-2, D-3); previews change only in type names.

### Strings

Moved (D-7): `conclusions.list.more` → `designSystem.badge.moreCount` (`L10n.DesignSystem.badgeMoreCount`), `conclusions.detail.dateUnknown` → `designSystem.datePlaceLine.dateUnknown`, `conclusions.detail.placeUnknown` → `designSystem.datePlaceLine.placeUnknown`; same copy; their `Localizable.xcstrings` entries move. No new copy. The wording of a relabelled product term changes on every term surface (D-2, D-3): no key changes for it.

## 7. Steps

**STEP-1 — One term-name rule everywhere**
Files: `macos/App/Platform/PropertyTermDisplay.swift`, `macos/App/Features/Conclusions/ReconciledValueDisplay.swift`, `ReconciledValueRowModel.swift`, `ReconciliationRecord.swift`, `PlaceTitleDisplay.swift`, `PersonGenderTreatment.swift`, `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/ProvenenciaTests/PropertyTermDisplayTests.swift`, `ReconciledValueDisplayTests.swift`, `ObservationValueDisplayTests.swift`, `EvidenceBridgeEdgeSummaryTests.swift`, `PersonGenderTreatmentTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-3.
Check: T-1, T-2, T-3 pass; `PlaceTitleDisplayTests`, `PersonDetailTests`, `ConnectionSentenceRegistryTests` pass.

**STEP-2 — Title styling and the count badge into the kit**
Files: `macos/App/Features/Conclusions/ConclusionListRow.swift` → `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)* (the title part), `macos/App/DesignSystem/Components/Badge/PVBadge.swift`, the 13 files that name `ConclusionTitleSource`, `ConclusionDetailPage.swift`, `ConclusionListPage.swift`, `PersonTile.swift`, `PersonsListView.swift`, `macos/ProvenenciaTests/PersonTitleDisplayTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-4, D-6, D-7's `badgeMoreCount`, D-11's first row.
Check: the app builds; `PersonsListTests`, `EventTitleDisplayTests`, `PlaceTitleDisplayTests`, `PersonTitleDisplayTests` pass; `rg -w ConclusionTitleSource macos` finds nothing; the detail page title previews look as before; `scripts/check-guidance-names.py` passes.

**STEP-3 — The date · place line into the kit; the Events list onto it**
Files: `macos/App/Features/Conclusions/ConclusionDatePlaceLine.swift` → `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)*, `EventsListView.swift`, `PersonLifeDisplay.swift`, `PersonTile.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PersonDetailTests.swift`, `PersonsListTests.swift`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-5, D-7's two line keys, D-11's second row.
Check: T-4 passes; `PersonDetailTests`, `PersonsListTests` pass; `python3 scripts/check-localizable-xcstrings.py` passes; the Person header, Family's tiles, the Event page's line look as before and the Events list row reads AC-3.1 (previews).

**STEP-4 — Name formatting to Platform**
Files: `macos/App/Features/Names/NameValueDisplay.swift` → `macos/App/Platform/NameValueDisplay.swift` *(new)*, `NamePartType.swift` → `macos/App/Platform/NamePartType.swift` *(new)*, `macos/App/Features/Names/NameValueDraft.swift`, `NameValueEditorForm.swift`, `api/ffi/swift_vocabulary_test.go`.
Do: D-8.
Check: `NameValueDraftTests`, `PersonTitleDisplayTests`, `ReconciledValueDisplayTests` pass; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `rg -w 'NameValueDraft|PVSelectOption' macos/App/Platform/NameValueDisplay.swift macos/App/Platform/NamePartType.swift` finds nothing.

**STEP-5 — Date formatting to Platform, with a year precision**
Files: `macos/App/Features/Dates/DateRowDisplay.swift`, `DateValueDisplay.swift`, `DateValueDraft.swift` → `macos/App/Platform/`, `macos/ProvenenciaTests/DateRowDisplayTests.swift`, `.agents/skills/add-date-value/SKILL.md`. Skill: `add-date-value`.
Do: D-9's moves, `DateQualifier` constants and `Precision`; the skill link.
Check: T-5 passes; `DateValueDisplayTests`, `DateValueDraftTests` pass.

**STEP-6 — Places, the person's title and life to Platform**
Files: `macos/App/Features/Conclusions/PlaceChainDisplay.swift`, `DerivedPlace.swift`, `PersonTitleDisplay.swift`, `PersonLifeDisplay.swift` → `macos/App/Platform/`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-10, and D-9's precision through `dateText` and `dateSpan`.
Check: T-6 passes; `PersonsListTests`, `DerivedPlaceTests`, `PersonTitleDisplayTests` pass.

**STEP-7 — The kit check**
Files: `scripts/check-guidance-names.py`, `scripts/check-guidance-names.allow`.
Do: D-12; then the one-off `rg` over the moved `Platform/` files.
Check: T-7: `python3 scripts/check-guidance-names.py` passes; adding a scratch reference to `PersonsListView` in a kit file makes it fail (then remove it); the one-off `rg` prints nothing (paste it in the PR description).

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (S13-34's row), this plan (status, Landed note), `docs/design-system-layers.md`, `docs/macos-client-patterns.md`, `.cursor/rules/design-system-layers.mdc`, `.agents/skills/add-ui-component/SKILL.md`, `docs/ideas/structural-debt.md` (rows, next free A-numbers: `PVOmnibarHitRow` filed as a recipe though content-agnostic and titled with a `String`; `DateValueDraft`'s three editor flags; `DateRowDisplay` and `DateValueDisplay` as two formatters; `L10n.Properties.boundOverflow` beside `PVBadge.moreCount`; `WorkspaceLocation` in `PVDeleteImpact`).
Do: D-13; per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-34-display-pieces-move-down.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PropertyTermDisplayTests.swift` `nameOfAValue` *(new)* | `.term` `witness` (product), `SeededPropertyKey.role` → "Witness"; researcher "Godparent" → "Godparent"; researcher, empty label, key `godparent` → "godparent"; product `woman`, empty label, `gender` → "Woman"; empty key and label → nil; `.text("x")` → nil | AC-1.3, AC-1.4, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `termNamesFollowOneRule` *(new)*, the existing lines per D-2; `PersonGenderTreatmentTests.swift` (case added) | Product `woman` labelled "Female (self-described)", `gender` → "Woman" through `string(for:)` and `readAs`; researcher "Two-spirit" → "Two-spirit"; line 155 → "Male"; the gender treatment's `text` for the same summary equals the row's string; line 70 still nil | AC-1.1, AC-1.2, D-2 |
| T-3 | Swift | `macos/ProvenenciaTests/ObservationValueDisplayTests.swift` `termNamesFollowOneRule` *(new)*; `EvidenceBridgeEdgeSummaryTests.swift` (case added) | A gender Observation `woman` (product) with `valueText` "Female (self-described)" → "Woman"; a researcher term → its `valueText`; an empty `valueTermKey` → `valueText`; a participation bridge with role `witness` relabelled → sentence with "Witness"; for every bridge kind, `ConnectionSentenceRegistry.termProperty(for:)` equals the fixture observation's `propertyKey` | AC-1.1, AC-1.5, D-3 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonsListTests.swift` `eventsRowSecondLine` *(new)* | `EventSecondaryDisplay.content` for a marriage at Bremen with two extra places → date "4 Jun 1844", place "Bremen", extra 2; the row's accessibility label unchanged (line 166) | AC-3.1, AC-3.2, D-5 |
| T-5 | Swift | `macos/ProvenenciaTests/DateRowDisplayTests.swift` `yearPrecision` *(new)* | AC-2.4's eight cases; `.full` unchanged | AC-2.4, D-9 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeDateTextPrecision` *(new)* | `PersonLifeDisplay.dateText(3 May 1790, precision: .year)` → "1790"; "abt. 1790" kept; `.full` → "3 May 1790"; `PersonLifeDisplay.line` for Anna unchanged; the `.stated` life line's spoken "date unknown" built by `PVDatePlaceLine.spokenDate(nil)` | AC-2.5, D-5, D-9 |
| T-7 | script | `scripts/check-guidance-names.py` run by `check-pr.sh` | Passes at the end of STEP-7; fails with a feature type added to a kit file | AC-2.1, D-12 |

AC-2.2 is STEP-7's one-off `rg`; AC-2.3 is the existing suites and the previews. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `ReconciledValueDisplayTests`, `ObservationValueDisplayTests`, `EvidenceBridgeEdgeSummaryTests` (STEP-1), every suite naming `ConclusionTitleSource` (STEP-2), `PersonDetailTests` (STEP-3, STEP-6), `go test ./api/ffi/...` (STEP-4).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `ConclusionTitleSource`, `ConclusionListRow.title`, `italicizingEmphasis`, `ConclusionDatePlaceLine`, `PersonLifeDisplay.Missing`, `EvidenceBridgeEdgeSummary.termLabel`, `L10n.Conclusions.moreCount`, `L10n.Conclusions.dateUnknown`, `L10n.Conclusions.placeUnknown`, `conclusions.list.more`, `conclusions.detail.dateUnknown`, `conclusions.detail.placeUnknown` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/**/*.md` (bar spike plan history), `macos/` and `L10n.swift` with no hit left; every moved file's old path (`Features/Names/NameValueDisplay.swift`, `Features/Names/NamePartType.swift`, `Features/Dates/DateRowDisplay.swift`, `Features/Dates/DateValueDisplay.swift`, `Features/Dates/DateValueDraft.swift`, `Features/Conclusions/PlaceChainDisplay.swift`, `DerivedPlace.swift`, `PersonTitleDisplay.swift`, `PersonLifeDisplay.swift`, `ConclusionDatePlaceLine.swift`) `rg --hidden`'d across `.agents/`, `.cursor/`, `docs/`, `api/`, `scripts/` with no hit outside plan history; `scripts/check-guidance-names.py` passes
- [ ] The kit check is in check-guidance-names.py and passes; the one-off Platform `rg` output is in the PR description
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI unchanged but the term names (AC-1.1, AC-1.5) and the Events list's line (AC-3.1)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped, naming both visible changes
- [ ] docs/design-system-layers.md, docs/macos-client-patterns.md, .cursor/rules/design-system-layers.mdc, add-ui-component, docs/ideas/structural-debt.md (STEP-8)
- [ ] PR description: what shipped, AC evidence, the one-off rg, deviations, manual check
```

## 10. Manual check

1. (AC-1.1, AC-1.2) In a project with a person whose gender is the built-in "Woman" relabelled "Female (self-described)" in Properties: the Vitals gender row, the header badge, the provenance sheet's "Read as" lines for the gender records and the gender Observation on the Evidence graph card read "Woman". A person with the researcher's own term "Two-spirit": all read "Two-spirit".
2. (AC-1.5) Relabel the built-in role `witness`; a participation bridge sentence on an Evidence graph reads "Witness".
3. (AC-3.1, AC-3.2) The Events list: each row's second line has the date in mono, the place small, italic and secondary, a "+N" where events have more places; VoiceOver reads each row once, as before.
4. (AC-2.3) The Persons and Places lists, a Person's header and Family, an Event page, a detail page title compact and full: as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rename of `ConclusionTitleSource` touches 13 files | One mechanical commit (STEP-2), checked by the build and an `rg` |
| The detail page title moves into `PVTitleText` and drifts by a point | `.display(compact:)` copies today's font, size, tracking and colour (D-4); the detail page previews in STEP-2 |
| The Events list row is read twice by VoiceOver | `PVDatePlaceLine` is hidden there (D-5); manual check 3 |
| The kit check finds hits beyond `WorkspaceLocation` at the base | Stop and report (D-12) |

**Stop and report if:**

- A file under § 4 has callers the rows do not list and a move would change their behavior.
- A moved file still names a feature type that cannot be passed in or left behind as an extension.
- A test outside § 8 expects the old term wording.
- `ConnectionSentenceRegistry.termProperty(for:)` differs from a bridge observation's `propertyKey` for some kind (T-3).
- `PVDatePlaceLine` or `PVTitleText` needs a feature type to compile.
- The kit check needs an allow entry other than `WorkspaceLocation`.

## 12. Open questions

None.

## 13. PR

**Title:** S13-34 — Display pieces move down a layer
**Body:** objective in one sentence; what shipped against each US; AC evidence; the one-off `rg`'s output; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-34-display-pieces-move-down.md`.
