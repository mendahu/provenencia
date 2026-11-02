# S13-34 — Display pieces move down a layer

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 4 — Timeline (preparatory)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a move of display code between layers; the one visible change is the wording of a relabelled built-in term in Vitals and the gender badge, which the user approved on 2026-11-02 and which draws nothing new)
**Board:** —
**Depends on:** —
**Blocks:** S13-35 (the card shell and the person card read the moved pieces), S13-13
**Branch:** `s13-34-display-pieces-move-down`
**Runs on:** macOS: every Step and `scripts/check-pr.sh` need Xcode. No Go or proto change
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-12 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** S13-35 plan [`S13-35-person-card-joins-the-design-system.md`](S13-35-person-card-joins-the-design-system.md) (reads every name this PR lands); working rule "The person card lives in the design system; display is decoupled from content" ([`../deployment-plan.md`](../deployment-plan.md#client)); [`docs/design-system-layers.md`](../../../design-system-layers.md) (lower layers do not import higher ones; recipes may read catalog types); the `PersonTile` evaluation of 2026-11-02 (its plan items 1 and 2); the S13-13 plan [`S13-13-timeline-section.md`](S13-13-timeline-section.md), whose D-5, D-18 and D-20 move here (D-4's `years`, `LifeLine.text` and `helpText` move to S13-35); precedent S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)), which moved the artifact card's hover into the kit as `pvLiftHover`
**Planned:** 2026-11-02 against `main` @ `04a2c99`
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

The pieces that word and style a person, a date, a place or a term move out of the Person, Names and Dates feature folders, so the design system can build a person card (S13-35) that a family tree in another feature will reuse, without the design system reaching into feature code. Title styling (name plain, label italic, ref mono) and the date · place line become design-system components; the formatting of names, dates, places, a person's title and life span moves to `Platform/`; and one rule names a term value everywhere.

**Why now.** The user chose to hoist the person card into the design system (2026-11-02); every piece it is built from lives in `Features/` today, and S13-13 needs the card's compact size. This PR is the first of two preparatory PRs (S13-34, then S13-35) before S13-13.

**What changes for the researcher.** One thing: a built-in term the researcher relabelled (for example the gender `woman` relabelled "Female (self-described)") now reads with the app's name ("Woman") in the Vitals row and the header's gender badge, as pickers already show it (user, 2026-11-02). Everything else looks and reads as before.

**What stays.** Every list, page, tile and line looks the same; every type keeps its behavior; editors keep their draft-only functions in their features.

## 2. User stories

**US-1.** As a researcher, a term reads the same everywhere the app names it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The built-in gender term `woman` relabelled "Female (self-described)" on a person | The Vitals gender row and the header's gender badge both read "Woman" (D-2) |
| AC-1.2 | The researcher's own gender term "Two-spirit" | Vitals and the badge read "Two-spirit", unchanged (D-2) |
| AC-1.3 | A role term value `witness` (product origin) passed to `PropertyTermDisplay.name(_:propertyKey: SeededPropertyKey.role)` | "Witness"; a `.text` value → nil (D-1) |
| AC-1.4 | A product term with an empty catalog label | Its product name; a researcher's term with an empty label → its key (D-1, today's fallback) |

**US-2.** As a contributor building the person card or a family tree, every display piece the card needs is outside `Features/`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The files this PR moves or adds under `macos/App/DesignSystem/` and `macos/App/Platform/` | None names a type declared under `macos/App/Features/` (the layer check of STEP-7) |
| AC-2.2 | The Persons, Events and Places lists, the Person header, Family and the Event page | They look and read as before (existing tests; previews) |
| AC-2.3 | A date 3 May 1790 through `DateRowDisplay.line(date:precision: .year)` | "1790"; "abt. 3 May 1790" → "abt. 1790"; a range 1790 – 1795 → "bet 1790 and 1795"; `.full` unchanged (D-7) |
| AC-2.4 | Anna (birth 3 May 1790, death 1851) through `PersonLifeDisplay`'s span at `.year` precision | "1790 – 1851"; at `.full` the Persons list's line is unchanged ("3 May 1790 – 1851") (D-7) |

## 3. Scope

| In | Covered by |
| --- | --- |
| One term-name rule: `PropertyTermDisplay.name(_:propertyKey:)`; Vitals rows and the gender badge read it (S13-13 D-5, D-18) | STEP-1, AC-1.1 – AC-1.4 |
| `PVTitleSource` and `PVTitleText` in the design system | STEP-2, AC-2.1, AC-2.2 |
| `PVDatePlaceLine` and its `Missing` in the design system; the line's copy in `L10n.DesignSystem` | STEP-3, AC-2.1, AC-2.2 |
| Name formatting to `Platform/` | STEP-4, AC-2.1 |
| Date formatting to `Platform/`, with a year precision (S13-13 D-20) | STEP-5, AC-2.1, AC-2.3 |
| Place, person-title and life formatting to `Platform/`; the life span takes the precision | STEP-6, AC-2.1, AC-2.4 |
| The layer check | STEP-7, AC-2.1 |

| Out | Home |
| --- | --- |
| The card shell, the Return modifier, the focus-ring button style, the gender treatment and the person card in the design system; `PersonLifeDisplay.years`, `LifeLine.text`, the card's `helpText` | S13-35 |
| A permanent script that fails CI when the design system names a feature type | S13-35 (its Check is that nothing under `DesignSystem/` names a feature type) |
| `EventTitleDisplay`, `EventSecondaryDisplay`, `PlaceTitleDisplay`, `ConclusionSummaryDisplay` to `Platform/`; event and place cards | when each card has a second caller (map working rule "The person card lives in the design system; display is decoupled from content") |
| Folding the Events list's own `EventSecondaryLine` into `PVDatePlaceLine` | a follow-up; logged in [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) by STEP-8 |
| `PVOmnibarHitRow` to `macos/App/DesignSystem/Components/` with `PVTitleText` | a follow-up; logged by STEP-8 |
| Splitting `DateValueDraft`'s three editor flags (`showStartTime`, `showEndTime`, `showAdvanced`) into the editor's own state | logged by STEP-8 |
| Moving `CachedThumbnail` out of `Platform/` | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Term names | `macos/App/Platform/PropertyTermDisplay.swift` `name(key:origin:propertyKey:catalogLabel:)`, `name(term:propertyKey:)` | Product name for a product term, else the trimmed label, else the key. No overload for a `CatalogConclusionValue` |
| Old term rule | `macos/App/Features/Conclusions/ReconciledValueDisplay.swift` `string(for:locale:orders:)` (line 84 – 105) | `.term(_, key, label, _)` → `label.isEmpty ? key : label`: no product names. Callers: `ReconciledValueRowModel.swift:60, 110` (Vitals), `PlaceTitleDisplay.swift:25` (a place's toponym values), `PersonGenderTreatment.swift:73` (the badge); tests `ReconciledValueDisplayTests.swift:84, 151 – 159` |
| Gender treatment | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `read(_:)`, `treatment(_:basis:)` | Words the term through `ReconciledValueDisplay.string(for:)`; `basis` is `.gender` or `.sexAtBirth` |
| Title styling | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionTitleSource` (line 6), `ConclusionListRow.title(_:extraCount:lineLimit:emphasized:)` (line 45), private `titleText(_:)`, `italicizingEmphasis(_:font:)` (line 84) | Content-agnostic (tokens, `PVBadge`, `PVListText.titleFont`) but in a feature; reads `L10n.Conclusions.moreCount`. `ConclusionTitleSource` is named in 13 files (about 22 references); `title` in `PersonTile.swift:99` and `ConclusionListPage.swift:183`; `italicizingEmphasis` also in `ConclusionDetailPage.swift:773` |
| Date · place line | `macos/App/Features/Conclusions/ConclusionDatePlaceLine.swift` `ConclusionDatePlaceLine` | Tokens only, but typed on `PersonLifeDisplay.Missing` and reading `L10n.Conclusions.dateUnknown` / `placeUnknown`. Callers: `PersonTile.swift:108`, `PersonDetailView.swift:81`, `EventDetailView.swift:31` |
| Name formatting | `macos/App/Features/Names/NameValueDisplay.swift` (`NameFormatPart`, `NameFormatOrders`, `NameDisplayStyle`, `NameValueDisplay`, `NameDisplay`); `NamePartType.swift` (`NamePartDecoration`, `NamePartType`) | `NameValueDisplay.string(for draft: NameValueDraft)` (line 33) and `NamePartType.selectOptions` (`PVSelectOption`, line 52) are editor-only. `NameFormatOrders` is named in 32 files, three of them in `Platform/` already |
| Date formatting | `macos/App/Features/Dates/DateRowDisplay.swift` `line(date:start:end:includesPhrase:locale:)`, private `rowDate`, `civil`; `DateValueDisplay.swift`; `DateValueDraft.swift` | `rowDate` reads the date through `DateValueDraft(from:)`; `DateValueDisplay.string(for draft:)` is the full wording. `DateValueDraft` is Foundation and `L10n` only, with three editor flags; 8 files use it (editors in `Features/Dates` and `Features/CitationComposer`) |
| Places | `macos/App/Features/Conclusions/PlaceChainDisplay.swift`, `DerivedPlace.swift` | Formatting only (`L10n.Conclusions`) |
| Person title and life | `macos/App/Features/Conclusions/PersonTitleDisplay.swift`; `PersonLifeDisplay.swift` (`Line`, `LifeLine`, `Missing`, `line(_:locale:)`, `lifeLines(_:locale:missing:)`, private `lifeLine`, `dateText`, private `dateSpan`, `placeSpan`) | `dateSpan` is the Persons list's life span ("1790 – 1851", "1820 –", "– 1851", keys `conclusions.life.span*`); no year-only form; `LifeLine` has no single visible string |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Conclusions.moreCount`, `dateUnknown`, `placeUnknown`; `macos/App/Resources/Localizable.xcstrings` | `moreCount` is also read by `PersonsListView.swift:71` and `EventsListView.swift:116`; `dateUnknown` / `placeUnknown` by `EventDetailView.swift:35 – 36` and `PersonLifeDisplay.swift:79 – 80` |
| Kit README | `macos/App/DesignSystem/README.md` built table; `scripts/check-guidance-names.py` (`KIT_LAYERS`) | One row per kit component folder |
| Tests | `macos/ProvenenciaTests/PropertyTermDisplayTests.swift`, `ReconciledValueDisplayTests.swift`, `PersonGenderTreatmentTests.swift`, `DateRowDisplayTests.swift`, `DateValueDisplayTests.swift`, `DateValueDraftTests.swift`, `DerivedPlaceTests.swift`, `PersonTitleDisplayTests.swift`, `PersonDetailTests.swift`, `PersonsListTests.swift`, `PlaceTitleDisplayTests.swift`, `EventTitleDisplayTests.swift`, `NameValueDraftTests.swift`, `NameFormatOrdersFixtures.swift` | — |

The app is one Swift module, so a move between folders changes no compile-time import; the layers are a rule of the repo ([`docs/design-system-layers.md`](../../../design-system-layers.md)), checked here by STEP-7. The Xcode target uses synchronized folders (`objectVersion 77`): a moved file needs no project edit.

**Precedent.** S13-12 ([#676](https://github.com/mendahu/provenencia/pull/676)) moved `ArtifactCardHover` from `Features/Sources` into `DesignSystem/Components/HoverEffect/PVLiftHover.swift` with a README row and no visible change. Copy its shape: `git mv`, rename, README row, callers updated in the same commit.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PropertyTermDisplay.name(_ value: CatalogConclusionValue, propertyKey: String) -> String?` *(new)* in `macos/App/Platform/PropertyTermDisplay.swift`: `.term(_, key, label, origin)` → `name(key: key, origin: origin, propertyKey: propertyKey, catalogLabel: label)`; any other case → nil (S13-13 D-5, moved here) | One rule names a term value; S13-13's role badges and S13-35's gender treatment read it | A timeline-only role type |
| D-2 | One rule for Vitals and the badge (user, 2026-11-02; S13-13 D-18, moved here): `ReconciledValueDisplay.string(for:propertyKey:locale:orders:)` gains a required `propertyKey: String`; its `.term` case returns `PropertyTermDisplay.name(key:origin:propertyKey:catalogLabel:)`; every other case unchanged. Callers pass their key: `ReconciledValueRowModel` `field.propertyKey` (both calls), `PlaceTitleDisplay` `toponym?.propertyKey ?? ""`, the tests their own. `PersonGenderTreatment.treatment` calls `PropertyTermDisplay.name(value, propertyKey:)` directly, with `SeededPropertyKey.gender` or `.sexAtBirth` by `basis`; the badge and the row agree because the row routes terms through the same function | Two rules drift; the badge must match its row; S13-35 moves the treatment into the design system, where it may not call `ReconciledValueDisplay` (a feature type) | Logging it as debt (a second mechanism); a default `propertyKey` (a caller could forget it and get no product name) |
| D-3 | `ConclusionTitleSource` → `PVTitleSource` (same cases: `.name(String, styled:spoken:)`, `.label`, `.ref`; same `text`, `spokenText`); `ConclusionListRow.title(_:extraCount:lineLimit:emphasized:)` and its private `titleText` → `struct PVTitleText: View` *(moved)*, `init(_ source: PVTitleSource, extraCount: Int = 0, lineLimit: Int = 1, emphasized: Bool = false)`, same body; `italicizingEmphasis(_:font:)` → `PVTitleText.italicizingEmphasis(_:font:)`. Home `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)*. `ConclusionListRow` keeps `thumbnail`, `updatedColumn` and `accessibilityLabel` (now on `PVTitleSource`). Every reference is renamed in the same commit | Content-agnostic: three text styles and a count badge; a Component | Keeping the case named `.name` under a new `.value` (a churn of every call site for no gain); a recipe (it reads no catalog type) |
| D-4 | `ConclusionDatePlaceLine` → `PVDatePlaceLine` *(moved)* in `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)*, same parameters and body; `PersonLifeDisplay.Missing` → `PVDatePlaceLine.Missing` (`.stated`, `.omitted`), the one enum for the line and the life lines; `PersonLifeDisplay.lifeLines(_:locale:missing:)` takes it | One enum, owned by the view that draws the halves; a `Platform` formatter may name a design-system type (precedent: `Platform/CachedThumbnail.swift` draws `PVThumbnail`) | Two enums with the same cases |
| D-5 | Copy the components read moves to `L10n.DesignSystem`: `moreCount(_:)` (key `designSystem.moreCount`), `dateUnknown` (`designSystem.dateUnknown`), `placeUnknown` (`designSystem.placeUnknown`), same English copy and comments; the old `L10n.Conclusions` keys and their String Catalog entries go; every reader (`PersonsListView`, `EventsListView`, `EventDetailView`, `PersonLifeDisplay`) moves to the new keys | A design-system component reads no feature copy (evaluation flag 5) | Two keys with one copy |
| D-6 | Name formatting moves to `macos/App/Platform/`: `NameValueDisplay.swift` (with `NameFormatPart`, `NameFormatOrders`, `NameDisplayStyle`, `NameDisplay`) and `NamePartType.swift` (with `NamePartDecoration`). The editor-only members stay in `Features/Names` as extensions: `NameValueDisplay.string(for draft: NameValueDraft)` in `Features/Names/NameValueDraft.swift` and `NamePartType.selectOptions` in `Features/Names/NameValueEditorForm.swift` | Names are read by Conclusions, Evidence graph, Promote, search and the composer: formatting, not a feature | Moving `NameValueDraft` (an editor model) |
| D-7 | Date formatting moves to `macos/App/Platform/`: `DateRowDisplay.swift`, `DateValueDisplay.swift` and `DateValueDraft.swift` (`DateRowDisplay.rowDate` and `DateValueDisplay` read the date through it). `DateRowDisplay.line` gains `precision: Precision = .full` *(new)*, `enum Precision { case full, year }`: `.year` passes only the year to `civil` (no month or day) and keeps the qualifier, the range and the phrase rules ("abt. 1790", "bet 1790 and 1795"). (S13-13 D-20, moved here.) In `PersonLifeDisplay`, `dateText(_:locale:)` and the private `dateSpan(birth:death:locale:)` gain `precision: DateRowDisplay.Precision = .full` and pass it on; `line(_:locale:)` passes `.full`, so the Persons list is unchanged. S13-35 adds `years(_:locale:)` as `dateSpan(…, precision: .year)`, `LifeLine.text` and the card's `helpText` on top | One rule for a life span at two precisions (user, 2026-11-02: reuse the Persons list's span); `DateValueDraft` is the one Swift model of a date value's fields and their validation, read by display and editors alike | A raw `startYear` read (drops qualifiers and ranges); keeping the draft in `Features/Dates` and rewriting both display types on `CatalogDateValueInput` (a rewrite of tested code, not a move); adding `years` here (S13-35 owns the card's content) |
| D-8 | Places and the person's title and life move to `macos/App/Platform/`: `PlaceChainDisplay.swift`, `DerivedPlace.swift`, `PersonTitleDisplay.swift`, `PersonLifeDisplay.swift`. Their `L10n.Conclusions` keys stay where they are (`Platform/` may read any namespace) | They are formatting the card, the lists, the header and a tree all need | Moving their keys (churn, no layer gain) |
| D-9 | README rows in `macos/App/DesignSystem/README.md`'s built table for `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)* and `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)*, in the table's alphabetical place | `scripts/check-guidance-names.py` requires one per folder | — |
| D-10 | The layer check (STEP-7) is a one-off command recorded in the PR description: list every type declared under `macos/App/Features/` (`rg -o --no-filename '^\s*(?:(?:public|private|fileprivate|internal) )?(?:final )?(?:enum|struct|class|protocol|typealias|actor) (\w+)' -r '$1' macos/App/Features --type swift | sort -u`), then `rg -w` those names over the files this PR moved or added; the result is empty but for SwiftUI's own `Text` (a feature declares a nested `Text`) | Evidence for AC-2.1 without a new script now | A permanent script (S13-35's, Out) |

**Rules this plan bends.** None. The term-name change is the user's answer (2026-11-02), recorded in the map's working rule and here.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PropertyTermDisplay.name(_:propertyKey:)` *(new overload)* | func | `name(term:propertyKey:)` | Same rule, one more input type |
| `PVTitleSource`, `PVTitleText` *(moved, renamed)* | component | `ConclusionTitleSource`, `ConclusionListRow.title` | The same code, in the layer that owns content-agnostic display |
| `PVDatePlaceLine`, `PVDatePlaceLine.Missing` *(moved, renamed)* | component | `ConclusionDatePlaceLine`, `PersonLifeDisplay.Missing` | Same; one enum |
| `L10n.DesignSystem.moreCount`, `dateUnknown`, `placeUnknown` *(moved keys)* | keys | `L10n.Conclusions.*` | The component's copy in the kit's namespace |
| `DateRowDisplay.Precision` *(new)*; `precision:` on `DateRowDisplay.line`, `PersonLifeDisplay.dateText` and `dateSpan` *(new parameter)* | enum, parameter | `DateRowDisplay.line`, `dateSpan` | Extended |
| Name, date, place and person formatting in `Platform/` *(moved)* | files | — | Moves |

**Builds on.** `PropertyTermDisplay`, the kit's tokens, `PVBadge`, `PVListText`, `L10n.DesignSystem`, the S13-12 move precedent.
**Deletes.** `ConclusionTitleSource`, `ConclusionListRow.title`, `ConclusionListRow.italicizingEmphasis`, `ConclusionDatePlaceLine`, `PersonLifeDisplay.Missing` (renamed or moved); `L10n.Conclusions.moreCount`, `dateUnknown`, `placeUnknown` and their catalog entries; `ReconciledValueDisplay`'s own term rule (`label.isEmpty ? key : label`).
**Next adds.** S13-35 builds `PVTileCard` on `PVTitleText` and `PVDatePlaceLine` and the person card on the `Platform/` formatting; S13-35 adds `years`, `LifeLine.text` and `helpText` on the precision; S13-13 reads `PropertyTermDisplay.name(_:propertyKey:)` for roles; event and place cards later move their own formatting the same way.

### Reach

None — no new path into existing code: every move keeps its callers and behavior. The one new state is a relabelled product term reading its product name (AC-1.1), shown by T-2.

### Scale

None — no write, list or loop changes; the moves are code locations, and D-7's precision adds one branch per date.

### Edges

- **Assumptions.** `PropertyTermDisplay` returns the key when the label is empty, never "" for a non-empty key: `ReconciledValueDisplay`'s old fallback holds (T-1). `PlaceTitleDisplay`'s toponym values are names, not terms, so its wording does not move (`PlaceTitleDisplayTests` unchanged).
- **Neighbors.** A product term relabelled → product name (changes, AC-1.1); a researcher's term → its label (unchanged, AC-1.2); a product term with no product L10n resource (a seeded key the app does not name) → its label (unchanged); a non-term value → unchanged wording.
- **Removed inputs.** None — no input is decided or deleted.
- **Real data.** The product L10n resources for terms (`L10n.PropertyTerm.resource`) cover gender, sex at birth, event type, role and relationship type; any other seeded term keeps its catalog label, as pickers show it today.
- **Untrusted size.** None — no list grows.

### Schema

None — no catalog change.

### Go core

None.

### FFI

None.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Title styling | `PVTitleSource`, `PVTitleText` | `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)* | D-3 |
| Date · place line | `PVDatePlaceLine` | `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)* | D-4, D-5 |
| Formatting | moved files | `macos/App/Platform/` | D-6 – D-8 |

No view looks different; no preview changes but for type names.

### Strings

Moved (D-5): `L10n.Conclusions.moreCount` → `L10n.DesignSystem.moreCount`, `dateUnknown` → `L10n.DesignSystem.dateUnknown`, `placeUnknown` → `L10n.DesignSystem.placeUnknown`, same copy; their entries in `macos/App/Resources/Localizable.xcstrings` move to the new keys. No new copy. The wording of a relabelled product term changes (D-2): no key changes for it.

## 7. Steps

**STEP-1 — One term-name rule**
Files: `macos/App/Platform/PropertyTermDisplay.swift`, `macos/App/Features/Conclusions/ReconciledValueDisplay.swift`, `ReconciledValueRowModel.swift`, `PlaceTitleDisplay.swift`, `PersonGenderTreatment.swift`, `macos/ProvenenciaTests/PropertyTermDisplayTests.swift`, `ReconciledValueDisplayTests.swift`, `PersonGenderTreatmentTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2.
Check: T-1, T-2 pass; `PlaceTitleDisplayTests` and `PersonDetailTests` pass.

**STEP-2 — Title styling into the kit**
Files: `macos/App/Features/Conclusions/ConclusionListRow.swift` → `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` *(new)* (the title part), the 13 files that name `ConclusionTitleSource`, `ConclusionDetailPage.swift`, `ConclusionListPage.swift`, `PersonTile.swift`, the tests that name it, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`.
Do: D-3, D-9's first row.
Check: the app builds; `PersonsListTests`, `EventTitleDisplayTests`, `PlaceTitleDisplayTests`, `PersonTitleDisplayTests` pass; `rg -w ConclusionTitleSource macos` finds nothing; `scripts/check-guidance-names.py` passes.

**STEP-3 — The date · place line into the kit**
Files: `macos/App/Features/Conclusions/ConclusionDatePlaceLine.swift` → `macos/App/DesignSystem/Components/DatePlaceLine/PVDatePlaceLine.swift` *(new)*, `PersonLifeDisplay.swift`, `PersonTile.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `PersonsListView.swift`, `EventsListView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-4, D-5, D-9's second row.
Check: `PersonDetailTests` pass; `python3 scripts/check-localizable-xcstrings.py` passes; the Person header, Family's tiles and the Event page's date line look as before (previews).

**STEP-4 — Name formatting to Platform**
Files: `macos/App/Features/Names/NameValueDisplay.swift` → `macos/App/Platform/NameValueDisplay.swift` *(new)*, `NamePartType.swift` → `macos/App/Platform/NamePartType.swift` *(new)*, `macos/App/Features/Names/NameValueDraft.swift`, `NameValueEditorForm.swift`.
Do: D-6.
Check: `NameValueDraftTests`, `PersonTitleDisplayTests`, `ReconciledValueDisplayTests` pass; `rg -w 'NameValueDraft|PVSelectOption' macos/App/Platform/NameValueDisplay.swift macos/App/Platform/NamePartType.swift` finds nothing.

**STEP-5 — Date formatting to Platform, with a year precision**
Files: `macos/App/Features/Dates/DateRowDisplay.swift`, `DateValueDisplay.swift`, `DateValueDraft.swift` → `macos/App/Platform/`, `macos/ProvenenciaTests/DateRowDisplayTests.swift`.
Do: D-7's moves and `Precision`.
Check: `DateRowDisplayTests` (T-3), `DateValueDisplayTests`, `DateValueDraftTests` pass.

**STEP-6 — Places, the person's title and life to Platform**
Files: `macos/App/Features/Conclusions/PlaceChainDisplay.swift`, `DerivedPlace.swift`, `PersonTitleDisplay.swift`, `PersonLifeDisplay.swift` → `macos/App/Platform/`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-8, and D-7's precision through `dateText` and `dateSpan`.
Check: T-4 passes; `PersonsListTests`, `DerivedPlaceTests`, `PersonTitleDisplayTests` pass.

**STEP-7 — The layer check**
Files: none changed.
Do: D-10.
Check: the command prints no feature type (bar `Text`); paste its output in the PR description.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (S13-34's row), this plan (status, Landed note), `docs/design-system-layers.md` (a sentence: value formatting lives in `Platform/`, display components in the kit, so a recipe composes both without a feature), `docs/ideas/structural-debt.md` (rows, next free A-numbers: the Events list's `EventSecondaryLine` beside `PVDatePlaceLine`; `PVOmnibarHitRow` filed as a recipe though content-agnostic and titled with a `String`; `DateValueDraft`'s three editor flags).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-34-display-pieces-move-down.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PropertyTermDisplayTests.swift` `nameOfAValue` *(new)* | `.term` `witness` (product) with `SeededPropertyKey.role` → "Witness"; researcher "Godparent" → "Godparent"; researcher, empty label, key `godparent` → "godparent"; product `woman`, empty label, `gender` → "Woman"; `.text("x")` → nil | AC-1.3, AC-1.4, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift` `termNamesFollowOneRule` *(new)*; `PersonGenderTreatmentTests.swift` (case added) | Product `woman` labelled "Female (self-described)", `propertyKey: gender` → "Woman"; researcher "Two-spirit" → "Two-spirit"; a name value → unchanged; the gender treatment's `text` for the same summary equals the row's string; the existing cases (lines 151 – 159) pass with their `propertyKey` | AC-1.1, AC-1.2, D-2 |
| T-3 | Swift | `macos/ProvenenciaTests/DateRowDisplayTests.swift` `yearPrecision` *(new)* | 3 May 1790 at `.year` → "1790"; "abt. 3 May 1790" → "abt. 1790"; a range 1790 – 1795 → "bet 1790 and 1795"; a phrase-only date → its phrase; `.full` unchanged | AC-2.3, D-7 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeDateTextPrecision` *(new)* | `PersonLifeDisplay.dateText(3 May 1790, precision: .year)` → "1790"; "abt. 1790" kept; `.full` → "3 May 1790"; `PersonLifeDisplay.line` for Anna unchanged ("3 May 1790 – 1851") | AC-2.4, D-7 |

AC-2.1 is STEP-7's command; AC-2.2 is the existing suites and the previews. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `ReconciledValueDisplayTests` (STEP-1), every suite naming `ConclusionTitleSource` (STEP-2), `PersonDetailTests` (STEP-3, STEP-6).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `ConclusionTitleSource`, `ConclusionListRow.title`, `ConclusionListRow.italicizingEmphasis`, `ConclusionDatePlaceLine`, `PersonLifeDisplay.Missing`, `L10n.Conclusions.moreCount`, `L10n.Conclusions.dateUnknown`, `L10n.Conclusions.placeUnknown`, `conclusions.date.unknown` and `conclusions.place.unknown` (the old catalog keys, as `L10n.swift` names them) `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` and `L10n.swift` with no hit left outside spike-13 plan history; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] The layer check (STEP-7) is empty but for `Text`, its output in the PR description
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI unchanged but the term names (AC-1.1)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped, naming the term-name change
- [ ] docs/design-system-layers.md and docs/ideas/structural-debt.md (STEP-8)
- [ ] PR description: what shipped, AC evidence, the layer check, deviations, manual check
```

## 10. Manual check

1. (AC-1.1, AC-1.2) In a project with a person whose gender is the built-in "Woman" relabelled "Female (self-described)" in Properties: the Vitals gender row and the header badge read "Woman". A person with the researcher's own term "Two-spirit": both read "Two-spirit".
2. (AC-2.2) The Persons, Events and Places lists, a Person's header and Family, an Event page: as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rename of `ConclusionTitleSource` touches about 13 files and many tests | One mechanical commit (STEP-2), checked by the build and an `rg` |
| A test pins today's term wording for a relabelled product term elsewhere (Promote, the Evidence graph) | Those surfaces read `PropertyTermDisplay` already; only `ReconciledValueDisplay`'s callers change (§ 4) |
| `L10n` key moves leave an unused catalog entry | `check-localizable-xcstrings.py` (STEP-3) |

**Stop and report if:**

- A file under § 4 has callers the rows do not list and a move would change their behavior.
- A moved file still names a feature type that cannot be passed in or left behind as an extension (STEP-7 prints it).
- A test outside § 8 expects the old term wording.
- `PVDatePlaceLine` or `PVTitleText` needs a feature type to compile.

## 12. Open questions

None.

## 13. PR

**Title:** S13-34 — Display pieces move down a layer
**Body:** objective in one sentence; what shipped against each US; AC evidence; the layer check's output; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-34-display-pieces-move-down.md`.
