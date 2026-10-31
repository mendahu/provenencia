# Ship log — S14-09 Quote nicknames and add former surnames

**Plan:** [`S14-09-name-display.md`](S14-09-name-display.md) · **Branch:** `s14-09-name-display` · **PR:** #649
**State:** fixing 1

## Round 0 — implement
Base `5c5cdd35`. Head `fa80e134`. Gate PASS (orchestrator run). Report deviations:
- T-6b asserts relation and cause only (a `Different` result carries no `OnlyA` / `OnlyB`).
- Attributes set and read through typed `AttributeScopes` subscripts (as `PVComboBox` does); the dynamic members raised Swift 6 Sendable key-path warnings.
- `NameValueDisplay.display(for:style:orders:)` is the public entry; the order-only overload is private.
- `NameValueDraftTests.formerSurnameIsAllowed` reads `PVSelectOption.id` (the member's name).
- The ship log stub was committed in `2dd45256` with the docs; harmless, it lives on the branch.
- Implementer saw intermittent full-suite flakes under load (unrelated async Swift tests; `promotealign` `TestLargeSourceBudgets/warm_at_open` once); passed on rerun, none changed.
- Manual check §10 not run: italics unverified, left to the user.

## Round 1 — review
Reviewers: CON (sonnet), MNT, UX, SEC+PRF. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: `core/matchregistry` (matching layer); `NameValueDisplay` runs per row in the Persons list.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Plain text read from `StyledFormat` is a path `add-localized-string` and `macos-l10n.mdc` do not describe | Fix now → F-1 | verified: rule line 21 and skill step 4 name only `L10n.format`; `NameValueDisplay.swift` is the only `StyledFormat` caller. Adds guidance for D-7's shape, removes nothing |
| UX-2 | UX | minor | Why heading nests quotes: `Why “James “Jake” Robins”` | Fix now → F-2 | verified: `conclusions.why.title` = `Why “%@”`, `ReconciledValueRowModel.init(field:…)` passes the name lead. User (2026-10-31): Name values drop the outer quotes |
| UX-1 | UX | minor | VoiceOver cannot hear the nickname / former-surname markers | Fix now → F-3 | verified: `ConclusionListRow.accessibilityLabel` reads `source.text`; the detail title has no label. User (2026-10-31): do it in S14-09; supersedes §6 "VoiceOver reads `text`" |
| MNT-2 | MNT | minor | List row repeats `PVList`'s title font as a literal | Fix now → F-4 | verified: `PVList.swift:143` and `ConclusionListRow.titleText` |
| MNT-3 | MNT | minor | Detail header builds its title font three times | Fix now → F-4 | verified: `ConclusionDetailPage` `nameText`, `.name` and `.label` arms |
| SEC-note | SEC+PRF | — | Older catalogs do not get the new seed rows | Rejected | D-12, user 2026-10-19: no migration; AC-2.6 pins the old-catalog display |

Nits (logged): `CON-1 · core/database/conclusionheaders/order.go:38 · rewrap the personSortKey comment` · `CON-2 · ship log uncommitted in working tree (orchestrator's file)` · `MNT-4 · order.go:105 · rewrap` · `MNT-5 · ConclusionListRow.swift:4-5 · doc comment wrap` · `MNT-6 · NamePartType.swift:12 · name the guard row` · `MNT-7 · italicizingEmphasis lives on ConclusionListRow (D-11 placed it)` · `UX-3 · NameValueDraft.swift:138 · "Stored as" shows raw key former_surname:Smith, as surname_prefix does`

### Fix plan 1

**F-1 — Say when plain text comes from `StyledFormat`** (from MNT-1)
Files: `.agents/skills/add-localized-string/SKILL.md` ("Styled arguments"), `.cursor/rules/macos-l10n.mdc` (the line that names `L10n.format`). Skill: `add-localized-string`.
Do: add one line to each: "A format that some surfaces draw styled and others plain (`NamePartDecoration.format`) reads its plain text as `String(StyledFormat.attributed(…).characters)`, so the format has one parser. Copy that is only ever plain goes through `L10n.format`." Replace the skill's stale `L10n.Promote.titleHandle` example with `NameValueDisplay`'s `NameDisplayRun.styledText`. Then `rg -n "L10n.format" .agents .cursor/rules` and confirm no line now says the opposite.
Check: `python3 scripts/check-guidance-names.py` passes; scripts/check-pr.sh PASS.

**F-2 — Why heading without outer quotes for a Name value** (from UX-2)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Conclusions/ReconciledValueRowModel.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-localized-string`.
Failing test first: `PersonDetailTests.aMergedNameRowCarriesItsStateAndWhy` expects `row.whyTitle == "Why James Robins"`; add a case with a nick part expecting `Why James “Jake” Robins`.
Do: add `L10n.Conclusions.whyTitleName(_ value: String) -> String` *(new)*, key `conclusions.why.title.name` = "Why %@", comment "Heading of the records behind a name value; the argument is the displayed name, which may carry its own quotes". In `ReconciledValueRowModel.init(field:…)`, the single-value branch uses `whyTitleName` when `field.valueType == .name`, else `whyTitle` as today. `spanning` and the derived rows (dates, places) keep `whyTitle`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — Spoken markers for nicknames and former surnames on Person titles** (from UX-1)
Files: `macos/App/Features/Names/NamePartType.swift`, `macos/App/Features/Names/NameValueDisplay.swift`, `macos/App/Features/Conclusions/ConclusionListRow.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Conclusions/PersonHeaderDisplay.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonHeaderDisplayTests.swift`, `macos/ProvenenciaTests/NameValueDraftTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Failing test first: `PersonHeaderDisplayTests.spokenNameMarksDecoratedParts` *(new)* — `NameValueDisplay.display` for James/Jake/Robins (natural, `western`) has `spoken == "James Jake, nickname Robins"`; Mary/Smith/Jones has `spoken == "Mary former surname Smith Jones"`; James Robins has `spoken == nil`. `ConclusionListRow.accessibilityLabel(PersonHeaderDisplay.titleSource(name: mary…), ref:)` contains "former surname Smith".
Do:
- `NamePartDecoration` gains `let spoken: LocalizedStringResource`: `.nick` → `L10n.NameValue.a11yNick` (key `nameValue.a11y.nick` = "%@, nickname", comment "VoiceOver: a nickname inside a spoken name; the argument is the nickname. Keep exactly one %@"); `.formerSurname` → `L10n.NameValue.a11yFormerSurname` (key `nameValue.a11y.former_surname` = "former surname %@", comment "VoiceOver: a former surname inside a spoken name; the argument is the surname. Keep exactly one %@").
- `NameDisplay` gains `var spoken: String?`: built from the same runs only when a run has a decoration (each decorated run as `String(StyledFormat.attributed(decoration.spoken, [.init(text: word)]).characters)`, undecorated runs as their text), else nil. Same one-parser rule as F-1.
- `ConclusionTitleSource.name` gains a third payload `spoken: String? = nil`; add `var spokenText: String` (the spoken copy when set, else `text`). `text` and the two view switches bind the new payload. `PersonHeaderDisplay.titleSource` passes `d.spoken`.
- `ConclusionListRow.accessibilityLabel` reads `source.spokenText`. The detail title's `.name` arm adds `.accessibilityLabel(Text(verbatim: spoken))` when `spoken` is set.
- Every other surface keeps its plain text (the user's answer covers the two titles).
Check: the tests pass, existing `PersonsListTests`, `PersonDetailTests`, `PlaceTitleDisplayTests` and `EventTitleDisplayTests` unchanged; scripts/check-pr.sh PASS.

**F-4 — One title font per view** (from MNT-2, MNT-3)
Files: `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/Features/Conclusions/ConclusionListRow.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`. Skill: `add-ui-component`.
Do: `PVList` gains `static let titleFont = PVFont.body(size: PVTypeScale.body)` *(new)*, used at its own `.font(...)` for the primary text and in `ConclusionListRow.titleText`. `ConclusionDetailPage` header gains `private var titleFont: Font { PVFont.display(size: titleSize, weight: PVFontWeight.medium) }`, used in `nameText`, the `.name` modifier and the `.label` arm (`titleFont.italic()`).
Check: no visual change; scripts/check-pr.sh PASS.

Plan updates the fix round owes (the plan's Landed note): F-2 adds one string; F-3 supersedes §6 "VoiceOver reads `text`" (user, 2026-10-31) and adds two strings; F-1 adds a guidance line for D-7.
