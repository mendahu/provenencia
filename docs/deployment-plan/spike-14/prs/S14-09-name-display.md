# S14-09 — Quote nicknames and add former surnames

**Status:** Landed in #649 (2026-10-31)
**Spike / slice:** Spike 14 · Slice 3 — Name display (batch 3; was Spike 11 slice 2, S11-29)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no board: the titles gain punctuation and one italic run, with no new component, layout or state. The user chose the treatment on 2026-10-19, and on 2026-10-31 exempted this PR from the board rule; the design board's examples are updated afterwards from a prompt, outside this PR)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-09-name-display`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/name-values.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc` (wire enums have a row in `api/ffi/swift_vocabulary_test.go`)
**Context:** [`docs/structured-name-model.md`](../../../structured-name-model.md) §4.2a roles, §4.5 display styles; [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §4.1, §4.3; [`docs/ideas/international-names.md`](../../../ideas/international-names.md) "Name changes as a signal"; [audit 2](../app-health-audit-2.md) B-24 (part-type half, moved here from S14-22 on 2026-10-31); S9-D2 title anatomy in `macos/App/Features/Conclusions/ConclusionListRow.swift`; working rule "(Batch 3) S14-09 is Spike 11's S11-29, refreshed"
**Planned:** 2026-10-19 against `main` @ `84a9b31`; refreshed 2026-10-31 against `main` @ `b6a8f3e2`
**Review:** [`S14-09-name-display.review.md`](S14-09-name-display.review.md) — 2026-10-19: Ready after 2 rounds; 2026-10-31 refresh: Fits with changes / Ready with changes → fixed, 2 rounds
**Landed:** T-6b asserts relation and cause only (a `Different` result carries no `OnlyA` / `OnlyB`). `NameValueDisplay` and `ConclusionListRow.italicizingEmphasis` read and set attributes through typed `AttributeScopes` subscripts, as `PVComboBox` does, because the `.inlinePresentationIntent` / `.font` dynamic members warn under Swift 6. `NameValueDisplay.display(for:style:orders:)` is the public entry; the order-only overload is private. Review round 1 (user, 2026-10-31): the Why heading of a Name value drops the outer quotes (`conclusions.why.title.name`, "Why %@"); VoiceOver reads a spoken copy of a decorated name on the Persons list row and the Person page title (`NameDisplay.spoken`, `ConclusionTitleSource.name(_, styled:, spoken:)`, strings `nameValue.a11y.nick` and `nameValue.a11y.former_surname`), which supersedes §6 "VoiceOver reads `text`"; the list title font is `PVListText.titleFont` (a generic `PVList` cannot hold a stored static) and the detail header builds its font once; `add-localized-string` and `macos-l10n.mdc` say when plain text is read from `StyledFormat`. Manual check (§10) is the user's: italic `Smith` in the display font was not seen in a running app.

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

A Person's name shows which parts are a nickname and which a former surname. A nickname shows in curly double quotes (*James “Jake” Robins*), the genealogy convention. A new product part type, **Former surname**, records a maiden or other earlier surname. It shows in parentheses (*Mary (Smith) Jones*), and on the Persons list and Person page titles the surname inside the parentheses is italic. Today a nickname reads as one more given name (*James Jake Robins*) and a maiden name has no part type of its own. The Swift part-type enum gains the Go guard row every wire enum carries, so the next part type cannot drift.

**Why now.** User request (2026-10-19). [`international-names.md`](../../../ideas/international-names.md) names maiden names in the conclusion as the intended path for name changes. Audit 2 B-24 names this PR's `former_surname` as the case an unguarded `NamePartType` would drift on.

**What stays.** Stored parts and `form` are never rewritten; the decoration is display only. The Go list sort key (`nameformats.Format`, read by `conclusionheaders`) stays undecorated. Existing catalogs are not migrated (user, 2026-10-19). S11-28's nickname fold (the reconciler reads *Jake* in a given part as *James*) is untouched: it compares given-name words across records and never reads or writes `nick` parts.

## 2. User stories

**US-1.** As a researcher, when a name has a nickname part, I see it in quotes, so that I can tell it from a middle name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Parts `given=James`, `nick=Jake`, `surname=Robins`, the `NameFormatOrders.western` fixture, natural | `James “Jake” Robins` |
| AC-1.2 | Same, sorted | `Robins, James “Jake”` |
| AC-1.3 | Two nick parts `Jake`, `Jim` | `James “Jake” “Jim” Robins` (each part quoted) |
| AC-1.4 | The omnibar hit (with a person header) and the Promote row for that Person | Show the quotes; both reach `NameValueDisplay` through `ConclusionHeaderDisplay.title` |

**US-2.** As a researcher, I can type a name part as **Former surname**, and see it in parentheses and italics on the Person's title, so that a maiden name is recorded and reads as earlier.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The part-type picker in the name editor | Lists *Former surname* right after *Surname* |
| AC-2.2 | Parts `given=Mary`, `former_surname=Smith`, `surname=Jones`, natural | Text `Mary (Smith) Jones`; the styled title has one emphasized run, `Smith` |
| AC-2.3 | Same, sorted (Persons list row) | Text `Jones, Mary (Smith)`; one emphasized run, `Smith` |
| AC-2.4 | A name with no former surname | `ConclusionTitleSource` equals `.name("James Robins")` as today (`styled` nil) |
| AC-2.5 | Go `namevalue.Validate` on a part typed `former_surname` | Accepted |
| AC-2.6 | Orders that do not list `former_surname` (a catalog created before this PR), natural | `Mary Jones (Smith)`: an unlisted type follows, as `NameValueDisplay` does today |
| AC-2.7 | `valuecompare.Compare` of *Mary Smith* (`given=Mary`, `surname=Smith`) with `given=Mary`, `former_surname=Smith`, `surname=Jones`, shipped rules | `Contains`; `Parts` holds a `family` pair `smith`/`smith` with relation `Same` (beside the `given` pair `mary`/`mary`; part text is folded lowercase); `OnlyB` is the `family` part `jones` |
| AC-2.8 | `valuecompare.Compare` of *Mary Brown* (`given=Mary`, `surname=Brown`) with `given=Mary`, `former_surname=Smith`, `surname=Jones` | `Different`, cause `CauseNoMatch` — both sides keep an unpaired family word, as *Mary Brown* / *Mary Jones* does today |

**US-3.** As the next developer adding a part type, a Go test fails when the Swift enum and the Go registry disagree, and the skill lists every place a part type lives.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `TestSwiftVocabularyMatchesGo` after this PR | Has a "name part types" row; passes with `NamePartType`'s raw values equal to `namevalue.PartTypes()` keys |
| AC-3.2 | A Go part type with no Swift case (or the reverse) | That row fails, naming both lists |
| AC-3.3 | `add-name-value` step 6 | Lists every place a part type lives (STEP-5) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `former_surname` in the Go part-type registry, Western seed parts and roles, match registry Western roles | STEP-1, AC-2.5, AC-2.7, AC-2.8 |
| `NamePartType.formerSurname`, explicit raw values on every case, its label, picker entry, and `NamePartType.decoration` | STEP-2, AC-2.1 |
| The "name part types" row in `api/ffi/swift_vocabulary_test.go` | STEP-2, AC-3.1, AC-3.2 |
| `NameValueDisplay` builds runs; nick parts quoted, former surnames parenthesized, for every caller; a styled copy through `StyledFormat` when a run is emphasized | STEP-3, AC-1.1 – AC-1.4, AC-2.6 |
| A styled Person title with the former surname italic, on the Persons list row and Person detail title (regular and compact header) | STEP-4, AC-2.2, AC-2.3, AC-2.4 |
| Docs and guidance: structured-name-model §4.2a and §4.5, seeded-vocabulary §4.1 / §4.3, international-names, `add-name-value` step 6; comments on `nameformats.Format`, `personSortKey` and `NameValueDisplay` | STEP-5, AC-3.3 |

| Out | Home |
| --- | --- |
| Migrating existing catalogs' name-format rows | not planned (user, 2026-10-19: new catalogs only) |
| Italics anywhere but the Persons list row and Person detail title (value rows, Event titles, omnibar, Promote, Evidence) | not planned; those surfaces are plain text and keep the parentheses |
| Quotes in Go `nameformats.Format` (list sort key) and the Go-indexed omnibar title used when a hit has no person header | not planned; a sort or index key carries no punctuation (§4.5 rewritten in STEP-5) |
| `NameFormSplit` recognising quoted nicknames and parenthesized former surnames when splitting a typed form | not planned; logged in [`docs/ideas/international-names.md`](../../../ideas/international-names.md) in STEP-5, pointing at `NamePartType.decoration` |
| Moving `NamePartType` to `Platform/` | not planned; `.cursor/rules/macos-client.mdc` sends *new* wire enums to `Platform/`; an existing one keeps its home and gains its row, as `PropertyValueType` and `PVMarkKey` do |
| Dropping the unused `Sort` field | not planned; logged in `docs/ideas/international-names.md` in STEP-5 |
| Deriving the seed's Western roles from `matchregistry` and retiring `TestSeedMatchesRegistryWestern` | not planned; logged in `docs/ideas/international-names.md` in STEP-5 |
| A Go↔Swift fixture pin of the undecorated join (`nameformats.Format` ↔ `NameValueDisplay`) | not planned; logged in `docs/ideas/international-names.md` in STEP-5 |
| The reconciler or Promote reading a former and a current surname as the same person across records (*Mary Smith* in one record, *Mary Jones* in another) | not planned; "Name changes as a signal" in international-names stays open |
| Moving the hand-written orders in `ObservationValueDisplayTests` and `CatalogSearchTests` onto the shared fixtures | not planned; those tests do not touch the new type |
| Updating the Claude Design board's examples to the new formats | outside this PR: the user asks for a board prompt after planning (2026-10-31) |
| GEDCOM import / export of former surnames | not planned; no GEDCOM path exists |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Part-type registry (Go) | `core/namevalue/namevalue.go` consts; `core/namevalue/registry.go` `partTypeRegistry` | Seven types: prefix 0, given 1, nick 2, surname_prefix 3, surname 4, suffix 5, undetermined 6 |
| Western seed | `core/database/nameformats/registry.go` `seedProfiles` | Natural: prefix, given, nick, surname_prefix, surname, suffix. Sorted: surname_prefix, surname (`", "`), prefix, given, nick, suffix. Roles for all seven types. `Install` fails when the default profile lacks a role for a registry type (`hasRole` loop over `namevalue.PartTypes()`) |
| Match registry | `core/matchregistry/value.go` `Names.Western.PartRoles` | Roles for the seven types |
| Name compare | `core/valuecompare/name.go` `nameWords`, `leftoverOf`; step (6) "Leftovers" | Words keep their role; a name may hold several `family` words. Leftover words make the compare `Different` (`CauseNoMatch`) only when both sides keep a leftover family word (or both a leftover first given) |
| Name compare tests | `core/valuecompare/name_test.go` (package `valuecompare_test`, imports `matchregistry`): `nv` builds a Value from `partCodes` (`g n s sp x p u _`), `nameCases` rows run by `TestNameCompare` under `matchregistry.Default().Value`; `pair(kind, a, b, rel, sim, why)` | The one home for name-compare cases |
| Reconciler rules | `core/database/autoreconciler/autoreconciler.go` `ValueRules` = `valuerules.Product()` (`matchregistry.Default().Value` + product alias seed); `CacheVersion = 21` | The cache changes only when `CacheVersion` does |
| Go display join | `core/database/nameformats/nameformats.go` `Format` | The same order-and-separator join as Swift, used as the Persons list sort key (`core/database/conclusionheaders/order.go` `personSortKey`) |
| Swift vocabulary guard | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` table (`path`, `enum`, `re`, `want`); `listCase` regex reads `case x = "raw"`; `matches` sorts | No row for `NamePartType` |
| Swift part types | `macos/App/Features/Names/NamePartType.swift` `NamePartType` | Seven cases, only `surnamePrefix` with an explicit raw value; `label` switch; `selectOptions` builds the picker from `allCases` |
| Styled formats | `macos/App/Platform/StyledFormat.swift` `StyledFormat.attributed(_:_:literal:)`, `StyledFormat.Argument(text:attributes:)`; tests in `macos/ProvenenciaTests/StyledFormatTests.swift`; `add-localized-string` "Styled arguments" | Fills a localized format with styled arguments; literal text plain. No app caller today |
| Strings | `macos/App/Platform/L10n.swift` `L10n.NameValue.partType*` (static `LocalizedStringResource` constants); `macos/App/Resources/Localizable.xcstrings`; `L10n.format(_:_:)` | `nameValue.part.type.<key>` labels |
| Display | `macos/App/Features/Names/NameValueDisplay.swift` `NameValueDisplay.join` (private) | Groups parts by type, joins words with spaces, groups in the profile's order with trailing separators; unlisted types follow in stored order. Returns `String`; no per-type decoration |
| Display callers | `PersonHeaderDisplay`, `ReconciledValueDisplay`, `EventTitleDisplay.subjectName`, `ObservationValueDisplay`, `EvidenceSubjectCard`, `CitationObservationValue` | All call `NameValueDisplay.string`; they gain the quotes and parentheses with no change |
| Header titles | `macos/App/Features/Conclusions/ConclusionHeaderDisplay.swift` `ConclusionHeaderDisplay.title` | The one switch over person, event and place headers; read by `PromoteFlow.swift` (Promote rows) and `CatalogSearchHitPresentation.swift` (omnibar hits with a header); a person calls `PersonHeaderDisplay.title` |
| Title source | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionTitleSource` | `.name(String)`, `.label(String)` (rendered italic), `.ref(String)`; `text` matches `case .name(let text), .label(let text), .ref(let text)` in one arm. Built with `.name(…)` in `PersonHeaderDisplay`, `EventTitleDisplay`, `PlaceTitleDisplay` and tests |
| Person title | `macos/App/Features/Conclusions/PersonHeaderDisplay.swift` `titleSource(name:entity:style:orders:)` | The one place a Person's title is made; calls `NameValueDisplay.string`; used by `PersonsList.titleSource` (sorted, `PersonsListView.swift`), `PersonDetailContent.init` (natural, the lead displayed name) and `title(...)` |
| Title rendering | `ConclusionListRow.titleText(_:)`; `ConclusionDetailPage.swift` header `title` (`titleSize` is `PVTypeScale.h3` when `compact`, else `PVTypeScale.display3`) | `Text(verbatim: text)` for `.name`; the list row inherits `PVList`'s title font `PVFont.body(size: PVTypeScale.body)` |
| Tests | Go: `core/namevalue/registry_test.go` `TestPartTypesRegistry`; `core/matchregistry/registry_test.go` `TestWesternPatternCoversEveryPartType`; `core/database/nameformats/nameformats_test.go` `TestInstallLoadsWesternAndFormats` (`len(roles) == len(PartTypes())`), `TestSeedMatchesRegistryWestern`; `api/ffi/handlers/nameformat_test.go` `TestGetNameFormat` (exact seeded Natural and Sorted orders of a new catalog). Swift: `macos/ProvenenciaTests/NameFormatOrdersFixtures.swift` (`NameFormatOrders.givenSurname`), `PersonHeaderDisplayTests.swift` (`sortedOrderUsesTheCatalogSeparator` hand-writes a given / surname_prefix / surname order), `PersonsListTests.swift`, `PersonDetailTests.swift`, `NameValueDraftTests.swift`, `ObservationValueDisplayTests.swift`, `CitationComposerModelTests.swift` | Parity tests force a role for every registry type; `TestGetNameFormat` pins the seeded order. No Swift test shows a name with a `nick` part |
| Skill | `.agents/skills/add-name-value/SKILL.md` "Extending helpers" step 6 | "append to `core/namevalue/registry.go` + consts + seeded-vocabulary §4.1 + L10n" — omits the seed parts and roles, the match registry role, the Swift case and the guard row |
| Map | `docs/deployment-plan/spike-14/deployment-plan.md` S14-09 and S14-22 rows and scope entries | Re-cut on 2026-10-31 with this refresh: B-24's part-type half is S14-09's; S14-22 depends only on S13-23 |

**Precedent.** `nick` is the nearest part type: copy every place it appears (Go const and registry row, seed part rows in both styles, seed role, match registry role, Swift case, L10n key, xcstrings entry). The "value types" row (`PropertyValueType`) is the guard-row shape to copy. `StyledFormat.attributed` with `L10n.Promote.titleHandle` in `add-localized-string` shows a styled argument. The `.label` title case shows how a title goes italic in both the list row and the detail page. `NameFormatOrders.givenSurname` shows where a shared Swift orders fixture lives.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Nickname display: each `nick` part wrapped by `L10n.NameValue.displayNick` (key `nameValue.display.nick`, value `“%@”`) | User call (2026-10-19); GEDCOM tools, Gramps and RootsMagic quote nicknames; localizable for « » or „ “ | Parentheses — conventionally a maiden name |
| D-2 | Former surname display: each `former_surname` part wrapped by `L10n.NameValue.displayFormerSurname` (key `nameValue.display.former_surname`, value `(%@)`); the word inside is emphasized, the parentheses are not; italics show only on the Persons list row and Person detail title | User call (2026-10-19): parentheses + italic. Plain-text surfaces keep the parentheses | Italic only — lost on plain-text surfaces and close to the italic working-label title (S9-D2); *née* — gendered and language-bound |
| D-3 | Key `former_surname`; label "Former surname"; Sort 5 (after surname); suffix 6, undetermined 7 | Generic for maiden, earlier married or changed names (user) | `maiden` — narrower |
| D-4 | Western order: natural prefix, given, nick, **former_surname**, surname_prefix, surname, suffix; sorted surname_prefix, surname (`", "`), prefix, given, nick, **former_surname**, suffix. The seed role row `{PartTypeFormerSurname, RoleFamily}` goes right after the `surname` role | *Mary (Smith) Jones* / *Jones, Mary (Smith)* | After the surname — reads as a second surname |
| D-5 | Role `family` in the seed roles and the match registry (both, so `TestSeedMatchesRegistryWestern` holds) | A former surname is a family name for matching; the record *Mary Smith* should pair it (AC-2.7) | `ignored` — throws away the maiden-name signal; a new role — needs a comparer change |
| D-6 | Decoration lives in Swift: `struct NamePartDecoration { let format: LocalizedStringResource; let emphasized: Bool }` (not `Equatable`; nothing compares it) *(new)* and `NamePartType.decoration: NamePartDecoration?` *(new)*: `.nick` → `(displayNick, false)`, `.formerSurname` → `(displayFormerSurname, true)`, every other case `nil`. Go has no display decoration | One property per decorated type; display is Swift's job; Go `Format` is a sort key | Two parallel properties (`displayFormat`, `isEmphasized`) — two switches per type; a catalog column — schema for two constants |
| D-7 | A decorated word is rendered by `StyledFormat` only: styled text is `StyledFormat.attributed(decoration.format, [StyledFormat.Argument(text: word, attributes: decoration.emphasized ? emphasis : AttributeContainer())])`, where `emphasis` sets `inlinePresentationIntent = .emphasized`; plain text is `String(StyledFormat.attributed(decoration.format, [StyledFormat.Argument(text: word)]).characters)`. No `L10n.format` call outside `L10n.swift`, no sentinel, no splitting | `StyledFormat` is the one styled-format mechanism (`add-localized-string` "Styled arguments") and the one parser of a decoration format; `L10n.format` stays inside `L10n.swift` (`add-localized-string` step 4). The cost is per decorated word, not per name | A private splitter around a placeholder character — a second mechanism; `L10n.format` in `NameValueDisplay` — its first call outside `L10n.swift`, against step 4 |
| D-8 | `private struct NameDisplayRun { var text: String; var decoration: NamePartDecoration? }` *(new)*. `NameValueDisplay.runs` *(new, private)* replaces `join`: one run per non-blank part word (with its type's `decoration`, looked up by `NamePartType(rawValue:)`), and one undecorated run for each space between words and each trailing separator between groups (an empty separator is a run of one space, as `join` writes today); runs are never merged. A part whose value already holds quotes or parentheses is shown as typed, inside the decoration (accepted). `NameValueDisplay.display(for:style:orders:) -> NameDisplay` *(new)* returns `struct NameDisplay: Equatable { var text: String; var styled: AttributedString? }` *(new)*: `text` is each run rendered plain (D-7) and concatenated; `styled` is built only when a run's decoration is emphasized, each run rendered styled (D-7; an undecorated run is `AttributedString(text)`) and concatenated, else nil. No runs (no parts, or all blank) → `text` is `string(form:)`, `styled` nil. `string(for:order:)` and `string(for:style:orders:)` return `display(...).text` | One join for both outputs; the common name (no former surname) builds no `AttributedString`, so list cost is unchanged | Always building `AttributedString` and stripping it — per-name cost on every caller; a second join for styled text — a near-copy |
| D-9 | `ConclusionTitleSource.name(String, styled: AttributedString? = nil)`. `text` splits its arm: `case .name(let text, _): text` and `case .label(let text), .ref(let text): text`. No new constructor: callers write `.name(text)` or `.name(d.text, styled: d.styled)`. The enum's doc comment adds: "a name may carry a styled copy whose emphasized runs draw italic" | `.name("James Robins")` stays the String case with `styled` nil (AC-2.4); no overload on `AttributedString`, which is a string literal type | A static `.name(_ styled: AttributedString)` — ambiguous with `.name("…")` literals; a new `.styledName` case — every switch grows a case |
| D-10 | `PersonHeaderDisplay.titleSource(name:…)` calls `NameValueDisplay.display` once and returns `.name(d.text, styled: d.styled)` when `d.text` is not empty; `title(...)` still returns `titleSource(...).text` | One join per title | Calling `string` and a styled builder — two joins |
| D-11 | Both title views render `Text(ConclusionListRow.italicizingEmphasis(styled, font: <the view's title font>))` when `styled` is set, else `Text(verbatim: text)`, keeping today's modifiers (the detail page's font at `titleSize`, tracking, colour and header trait). `ConclusionListRow.italicizingEmphasis(_ styled: AttributedString, font: Font) -> AttributedString` *(new)* sets `.font` to `font.italic()` on each run whose `inlinePresentationIntent` is `.emphasized`; the detail page passes `PVFont.display(size: titleSize, weight: PVFontWeight.medium)` and the list row `PVFont.body(size: PVTypeScale.body)` (the font `PVList` gives titles) | The italic is set explicitly, so it does not hang on whether SwiftUI synthesizes italic from `.emphasized` in the custom display font, and no headless run has to judge a render; one helper keeps the mapping in one place; the display font already has an italic (the `.label` title uses it) | Relying on `.emphasized` alone with a fallback chosen after looking — a decision an agent cannot make; `.italic()` on the whole title — the label treatment; the range-font code inline in two views — a near-copy |
| D-12 | No migration; old catalogs show the type after the listed ones (AC-2.6) and match it untyped | User call (2026-10-19) | Backfill migration |
| D-13 | The omnibar hit without a person header keeps the Go-indexed title, unquoted | That title is a search index projection (Out) | Re-formatting index titles in Swift |
| D-14 | No `autoreconciler.CacheVersion` bump and no match-index rebuild | No stored name can carry `former_surname` before this PR (`Validate` rejects unknown types), so no cached result changes; `nick` keeps its role | A bump — recomputes every catalog's cache for no change |
| D-15 | The Swift fixture is `NameFormatOrders.western` *(new)* in `NameFormatOrdersFixtures.swift`, beside `givenSurname`: D-4's two orders, `", "` after `surname` in sorted. `sortedOrderUsesTheCatalogSeparator` uses it in place of its hand-written orders (same output for its parts) | The shared fixtures file is where test orders live | A private fixture in `PersonHeaderDisplayTests` — a second home for orders |
| D-16 | The seeded order is pinned by updating `TestGetNameFormat`'s exact want lists (insert `{PartType: "former_surname"}` after `nick` in both), not by a new Go test | That test already asserts the full seeded order of a new catalog | A new `TestWesternSeedOrder` — a near-copy |
| D-17 | The compare cases go in `core/valuecompare/name_test.go`: `partCodes` gains `"f": namevalue.PartTypeFormerSurname` (and `f former_surname` in `nv`'s doc comment); AC-2.8 is a `nameCases` row; AC-2.7 is `TestFormerSurnamePairsAsFamily` *(new)* in the same file, asserting relation, the `Smith` pair and `OnlyB` (a `nameCases` row would pin a similarity number this PR does not decide) | One home for name-compare cases; `TestNameCompare`'s swap pass covers the new type | A test in `core/matchregistry/registry_test.go` — that file checks the registry's shape, not compare outcomes |
| D-18 | Guard row in `TestSwiftVocabularyMatchesGo`: `{name: "name part types", path: swiftNamePartType, enum: "NamePartType", re: listCase, want: sorted(partTypeKeys()...)}` with `swiftNamePartType = "../../macos/App/Features/Names/NamePartType.swift"` *(new const)* and `partTypeKeys()` *(new helper, beside `listStrings`)* returning each `namevalue.PartTypes()` key. Every `NamePartType` case spells its raw value (`case given = "given"`) so `listCase` reads all of them | `macos-client.mdc`: every wire enum is a row; B-24 (user, 2026-10-31: moved here from S14-22) | `rawCase` plus `listCase` — two regexes for one enum |

**Rules this plan bends.** The design-board rule (`plan-pr` §3, user 2026-10-29: a PR that changes how a view looks waits on a board) → the titles change only by punctuation and one italic run, with a treatment chosen on 2026-10-19 → the user exempted this PR on 2026-10-31 and will update the board's examples from a prompt → no rule text changes; the exemption is this PR's alone.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `namevalue.PartTypeFormerSurname` *(new)* | const | `PartTypeNick` | Extends the registry |
| `NamePartType.formerSurname`, `NamePartType.decoration` *(new)* | enum case, property | `NamePartType.label` | Extends the one Swift part-type registry |
| `NamePartDecoration` *(new)*, in `NamePartType.swift` | type | `StyledFormat.Argument` | Says which localized format wraps a part type and whether its word is emphasized; rendering reuses `L10n.format` and `StyledFormat` (D-7) |
| `NameDisplayRun`, `NameValueDisplay.runs` *(new, private)* | type, func | `NameValueDisplay.join` (replaced) | `join` becomes `runs` |
| `NameDisplay`, `NameValueDisplay.display(for:style:orders:)` *(new)* | type, func | `NameValueDisplay.string(for:style:orders:)` | `string` now returns `display(...).text`; the title needs the styled copy too |
| `ConclusionTitleSource.name`'s `styled` payload | enum payload | `.label` italic | D-9 |
| `ConclusionListRow.italicizingEmphasis(_:font:)` *(new)* | func | `ConclusionListRow.titleText` `.label` italic | D-11; one mapping for both title views |
| `L10n.NameValue.partTypeFormerSurname`, `displayNick`, `displayFormerSurname` *(new)* | keys | `L10n.NameValue.partTypeNick` | New copy |
| `NameFormatOrders.western` *(new)* | test fixture | `NameFormatOrders.givenSurname` | Extends the same fixtures file (D-15) |
| `TestFormerSurnamePairsAsFamily` *(new)*, `partCodes` `"f"` | Go test, test alphabet | `core/valuecompare/name_test.go` `nameCases` | Same file (D-17) |
| `swiftNamePartType`, `partTypeKeys()` *(new)* and the guard row | test const, helper, table row | `swiftValueKinds`, `listStrings()`, the "value types" row | Extends the guard table (D-18) |

**Builds on.** `namevalue` registry, `nameformats` seed, `matchregistry` Western pattern, `NamePartType`, `NameValueDisplay`, `StyledFormat`, `PersonHeaderDisplay`, `L10n.format`, `NameFormatOrdersFixtures.swift`, `swift_vocabulary_test.go`.
**Deletes.** `NameValueDisplay.join` (replaced by `runs`). The hand-written orders in `sortedOrderUsesTheCatalogSeparator`. The sentence in structured-name-model §4.5 requirement 2 that the sort key is "the same string the row shows".
**Next adds.** `patronymic` or `maternal_surname` ([international-names](../../../ideas/international-names.md)) is one enum case and, if decorated, one `decoration` arm; the guard row fails until Go and Swift agree, and `add-name-value` step 6 lists the rest. Event titles that want the italic would first need `StyledFormat.Argument` to carry an `AttributedString` (it holds one `AttributeContainer` today); not planned.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A part typed `former_surname` | The Go readers of part types: `valuecompare.nameWords` (through `NamePattern.Role`) and `nameformats.Format` (by order). The search index reads a name's `form`, not its parts | An unlisted type follows (`Format`) or is untyped (`Role`), so old profiles degrade, never fail |
| `ConclusionTitleSource.name` with `styled` set | `ConclusionListRow.titleText`, `ConclusionDetailPage` `title` | The only two views that switch on `.name`; every other reader uses `text` |
| A name with only a former surname | Both titles | `(Smith)`, italic inside |

### Scale

None — no write, query, list or loop changes. The join stays one pass per name; an `AttributedString` is built only for a name holding an emphasized run (D-8), so the Persons list does no extra work for names without a former surname.

### Edges

- **Assumptions.** `valuecompare` allows several `family` words per name and sorts by role then text, so pairing is order-free; T-6 holds *Mary Smith* ↔ *Mary (Smith) Jones* and `TestNameCompare`'s swap pass holds AC-2.8 both ways. `Install` requires a role for every registry type; T-2 holds it. `StyledFormat` drops an argument a translation omits; the catalog is English only and the xcstrings comment says "Keep exactly one %@".
- **Neighbors.** *Mary Jones* ↔ *Mary (Smith) Jones* → `Contains` (Jones pairs, Smith is one side's leftover). *Mary Brown* ↔ *Mary (Smith) Jones* → `Different`, `CauseNoMatch` (AC-2.8, T-6b), as two different surnames are today.
- **Removed inputs.** None — no input is decided, excluded or deleted by this PR.
- **Real data.** The Western seed is the only table; T-2 pins its whole order and roles.
- **Untrusted size.** None — no new list read from data the app did not write.

### Schema

None — part types are a compiled registry; `name_format_profile_parts.part_type` and `name_format_profile_roles.part_type` are open TEXT. New catalogs get the rows from the seed.

### Go core

| Package | Change |
| --- | --- |
| `core/namevalue` | `PartTypeFormerSurname = "former_surname"` *(new)*; registry row `{Key: PartTypeFormerSurname, L10nKey: "nameValue.part.type.former_surname", Sort: 5}`; suffix Sort 6, undetermined 7 |
| `core/database/nameformats` (`registry.go`) | Part rows and role per D-4. `nameformats.go` `Format` doc comment adds: "The same order and separators as Swift `NameValueDisplay`; only Swift decorates (quotes, parentheses). This string is a sort key." |
| `core/database/conclusionheaders` (`order.go`) | `personSortKey` comment: "the profile's sorted part order, undecorated; the row may add quotes and parentheses" |
| `core/matchregistry` (`value.go`) | `namevalue.PartTypeFormerSurname: valuecompare.RoleFamily` in `Western.PartRoles` |

### FFI

None — part types and name-format rows already cross as strings. `TestGetNameFormat`'s want lists change (D-16); `swift_vocabulary_test.go` gains a row (D-18).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Part type | `NamePartType.formerSurname = "former_surname"` | `macos/App/Features/Names/NamePartType.swift` | Case right after `surname`; every case spells its raw value (D-18); `label` → `L10n.NameValue.partTypeFormerSurname`; `decoration` per D-6 |
| Decoration | `NamePartDecoration` | same file | D-6 |
| Runs and display | `NameDisplayRun`, `NameValueDisplay.runs`, `NameDisplay`, `NameValueDisplay.display` | `macos/App/Features/Names/NameValueDisplay.swift` | D-7, D-8. A comment on `runs`: "The same order and separators as Go `nameformats.Format`; only this side decorates." |
| Title payload | `ConclusionTitleSource` | `ConclusionListRow.swift` | D-9; the three matches on `.name` are `text`, `titleText` (this file) and `ConclusionDetailPage` `title` |
| Person title | `PersonHeaderDisplay.titleSource(name:…)` | `PersonHeaderDisplay.swift` | D-10 |
| List row | `ConclusionListRow.titleText` | `ConclusionListRow.swift` | `.name(let text, let styled)`: `if let styled { Text(italicizingEmphasis(styled, font: PVFont.body(size: PVTypeScale.body))) } else { Text(verbatim: text) }` |
| Detail title | `ConclusionDetailPage` `title` | `ConclusionDetailPage.swift` | Same, with today's font at `titleSize`, tracking, colour and header trait |

States: a name with only a former surname → `(Smith)`, italic inside. Accessibility: VoiceOver reads `text` (with quotes and parentheses); `ConclusionListRow.accessibilityLabel` is unchanged.

### Strings

Static `LocalizedStringResource` constants *(new)* in `L10n.NameValue`, each with an entry in `Localizable.xcstrings`:

- `partTypeFormerSurname`: `nameValue.part.type.former_surname` = "Former surname" — comment "NameValue part type: former_surname"
- `displayNick`: `nameValue.display.nick` = "“%@”" — comment "A nickname inside a displayed name; the argument is the nickname. Keep exactly one %@"
- `displayFormerSurname`: `nameValue.display.former_surname` = "(%@)" — comment "A former surname inside a displayed name; the argument is the surname. Keep exactly one %@"

No existing copy changes what it covers.

## 7. Steps

Steps 2–4 need macOS with Xcode for their checks (`scripts/check-pr.sh` runs `xcodebuild test`). On a host without Xcode, finish STEP-1 and the code of STEP-2 – STEP-4, then stop and report that the Swift checks and the manual check need a macOS runner.

**STEP-1 — Go: the former surname part type**
Files: `core/namevalue/namevalue.go`, `core/namevalue/registry.go`, `core/namevalue/registry_test.go`, `core/database/nameformats/registry.go`, `core/database/nameformats/nameformats.go` (comment), `core/database/conclusionheaders/order.go` (comment), `core/matchregistry/value.go`, `core/valuecompare/name_test.go`, `api/ffi/handlers/nameformat_test.go`. Skills: `add-name-value`, `add-match-rule`.
Do: D-3, D-4, D-5, D-16, D-17. Add T-1, T-2, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/namevalue/ ./core/database/nameformats/ ./core/matchregistry/ ./core/valuecompare/ ./api/ffi/handlers/` passes.

**STEP-2 — Swift: the part type, its label, decoration and guard row**
Files: `macos/App/Features/Names/NamePartType.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/NameValueDraftTests.swift`, `api/ffi/swift_vocabulary_test.go`. Skill: `add-localized-string`.
Do: the case, raw values, label, `NamePartDecoration`, `decoration`, the three keys (D-6), and the guard row (D-18). Add T-3, T-7, T-8.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; T-3 and T-7 pass.

**STEP-3 — Swift: quotes and parentheses in every name display**
Files: `macos/App/Features/Names/NameValueDisplay.swift`, `macos/ProvenenciaTests/NameFormatOrdersFixtures.swift`, `macos/ProvenenciaTests/PersonHeaderDisplayTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-7, D-8, D-15. Add T-4.
Check: Xcode tests pass T-4 and the existing `PersonHeaderDisplayTests`, `ObservationValueDisplayTests`, `CitationComposerModelTests`, `CatalogSearchTests`, `StyledFormatTests`.

**STEP-4 — Swift: italic former surname on the Person titles**
Files: `macos/App/Features/Conclusions/ConclusionListRow.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Conclusions/PersonHeaderDisplay.swift`, `macos/ProvenenciaTests/PersonHeaderDisplayTests.swift`. Skill: `add-ui-component` (no new component; the title stays a snowflake of `ConclusionListRow`).
Do: D-9, D-10, D-11. Add T-5.
Check: Xcode tests pass, including `PersonsListTests`, `PersonDetailTests`, `PlaceTitleDisplayTests` and the Event title tests unchanged; manual check steps 2–3.

**STEP-5 — Docs and guidance**
Files:
- `docs/seeded-vocabulary.md` §4.1 (inside the code fence, after `surname`: `former_surname      -- no GEDCOM part tag (GEDCOM 7: a NAME with TYPE MAIDEN or BIRTH)`) and §4.3 (both orders per D-4).
- `docs/structured-name-model.md` §4.2a roles sentence (`former_surname` → `family`) and §4.5: requirement 2 becomes "The key is that part order, undecorated: the row may add quotes and parentheses, the key never does"; the table gains a row "**Decoration** — nickname quotes, former-surname parentheses | The **part type**, in the client, localized by UI language; never in the sort key or the profile | *James “Jake” Robins*".
- `docs/ideas/international-names.md` under "Name changes as a signal": "The `former_surname` part type (S14-09) records a maiden or earlier surname, and a record carrying the former surname pairs it as a family name; reading a former and a current surname as one person across records is still open." Plus four parked lines (the Out rows): `NameFormSplit` wrappers; the unused `Sort` field; seed roles derived from `matchregistry`; a fixture pin of the undecorated join.
- `.agents/skills/add-name-value/SKILL.md` step 6 becomes: "Adding a product part type (not a migration): the Go const and `partTypeRegistry` row (`core/namevalue`); the Western seed parts in both styles and its role (`core/database/nameformats/registry.go`); the role in `core/matchregistry/value.go` `Western.PartRoles`; the Swift `NamePartType` case with its raw value, `label`, and a `decoration` arm when the type is displayed with punctuation (display only — never in `nameformats.Format`); the L10n key and xcstrings entry; seeded-vocabulary §4.1 and §4.3. The guard row in `api/ffi/swift_vocabulary_test.go` fails until Go and Swift agree."
- `docs/deployment-plan/spike-14/deployment-plan.md`: this PR's map row status. (The S14-22 re-cut was made in the map when this plan was refreshed.)
- `docs/deployment-plan/spike-14/completed.md`; this plan.

Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-09-name-display.md` reports no errors; `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per STEP, each passing its Check: STEP-1 "Add a former surname name part type in Go."; STEP-2 "Add the former surname to Swift's part types and guard them against Go's registry."; STEP-3 "Show nicknames in quotes and former surnames in parentheses."; STEP-4 "Italicize the former surname on Person titles."; STEP-5 "Document the former surname part type."

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/namevalue/registry_test.go` `TestPartTypesRegistry` | Replace the `len(types) != 7` check: the keys in order equal `[prefix given nick surname_prefix surname former_surname suffix undetermined]`; `Validate` accepts a `former_surname` part | AC-2.5, D-3 |
| T-2 | Go | `api/ffi/handlers/nameformat_test.go` `TestGetNameFormat` (want lists updated); `core/database/nameformats` `TestSeedMatchesRegistryWestern`, `TestInstallLoadsWesternAndFormats` and `core/matchregistry` `TestWesternPatternCoversEveryPartType` (existing) | A new catalog's `Natural` and `Sorted` orders equal D-4's; the existing tests pass with role `family` | D-4, D-5, D-16 |
| T-6 | Go | `core/valuecompare/name_test.go` `TestFormerSurnamePairsAsFamily` *(new)* | `Compare(nv("g=Mary\|s=Smith"), nv("g=Mary\|f=Smith\|s=Jones"), matchregistry.Default().Value)` → relation `Contains`; `det.Parts` contains a `PartMatch` whose `A` and `B` are `part(valuecompare.PartFamily, "smith")` with relation `Same`; `slices.Equal(det.OnlyB, []valuecompare.Part{part(valuecompare.PartFamily, "jones")})` | AC-2.7, D-5, D-17 |
| T-6b | Go | `core/valuecompare/name_test.go` `nameCases` row *(new)* "S14-09 former surname beside a different surname" | `a: nv("g=Mary\|s=Brown")`, `b: nv("g=Mary\|f=Smith\|s=Jones")` → `rel: Different`, `why: CauseNoMatch` | AC-2.8, D-17 |
| T-8 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "name part types" row *(new)* | Swift raw values equal the eight Go keys; removing a Swift case locally makes it fail (checked once, not committed) | AC-3.1, AC-3.2, D-18 |
| T-3 | Swift | `NameValueDraftTests.swift` `formerSurnameIsAllowed` *(new)* | `NamePartType.isAllowed("former_surname")`; in `selectOptions`, `former_surname` comes right after `surname` | AC-2.1 |
| T-7 | Swift | `NameValueDraftTests.swift` `decorations` *(new)* | `NamePartType.formerSurname.decoration?.emphasized == true`; `.nick.decoration?.emphasized == false`; `.given.decoration == nil`; `String(StyledFormat.attributed(NamePartType.nick.decoration!.format, [.init(text: "Jake")]).characters) == "“Jake”"` | D-6, D-7 |
| T-4 | Swift | `PersonHeaderDisplayTests.swift` `nicknamesAreQuoted`, `formerSurnameIsParenthesized`, `unlistedFormerSurnameFollows` *(new)*; `sortedOrderUsesTheCatalogSeparator` on `western` | With `NameFormatOrders.western`: AC-1.1, AC-1.2, AC-1.3 strings through `PersonHeaderDisplay.title` and `ConclusionHeaderDisplay.title`; `Mary (Smith) Jones` natural and `Jones, Mary (Smith)` sorted; with `western` minus `former_surname` → `Mary Jones (Smith)`; Vincent van Gogh unchanged | AC-1.1 – AC-1.4, AC-2.6, D-1, D-2, D-15 |
| T-5 | Swift | `PersonHeaderDisplayTests.swift` `formerSurnameRunIsEmphasized` *(new)* | `NameValueDisplay.display` for Mary, natural and sorted: `styled` is set, `String(styled.characters) == text`, and the runs with `inlinePresentationIntent == .emphasized` are exactly one, text `Smith`; for James “Jake” Robins `styled == nil`; `titleSource` for Mary is `.name("Mary (Smith) Jones", styled: d.styled)`; for James Robins `== .name("James Robins")`; `ConclusionListRow.italicizingEmphasis(d.styled!, font: .body)` sets `.font` on the `Smith` run only | AC-2.2, AC-2.3, AC-2.4, D-8, D-9, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestPartTypesRegistry`, `TestGetNameFormat`, `nameformats_test.go`, and any Go or Swift golden that shows a name with a `nick` part (it gains quotes). Such a golden is updated to the quoted string and listed in the PR.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `NameValueDisplay.join` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer (2026-10-31); no rule text changes
- [ ] UI: no new component or kit primitive — not gated (user, 2026-10-31)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status `Landed in #N`; docs/deployment-plan/spike-14/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: seeded-vocabulary §4.1 / §4.3, structured-name-model §4.2a / §4.5, international-names; skill: add-name-value step 6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Create a new project (the seed carries the new rows). Add a Source with a person named given *James*, nick *Jake*, surname *Robins*; promote it.
2. Persons list: *Robins, James “Jake”*. Person page title: *James “Jake” Robins*. The Promote row and an omnibar search for *Robins* show the quotes (AC-1.4).
3. Add a person given *Mary*, Former surname *Smith*, surname *Jones*; the picker lists Former surname after Surname. Persons list: *Jones, Mary (Smith)* with *Smith* italic. Person page title: *Mary (Smith) Jones* with *Smith* italic in the display font, at Regular and Wide window widths, and again after scrolling until the header compacts.
4. Open an older project with a former-surname part (or set one): the title reads *Mary Jones (Smith)* (AC-2.6).
5. Read `add-name-value` step 6: it names each place this PR touched for `former_surname` (AC-3.3).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The emphasized run renders upright in the display font | D-11 sets the italic font explicitly; manual check step 3 confirms it |
| `family` role makes a former surname cost points against records that carry only the married surname | Accepted (D-5): such a record pairs the married surname and the former surname is one side's leftover, so the compare is `Contains`, not a conflict (Edges) |
| The list row shows quotes the sort key does not have | §4.5 says so (STEP-5); the key sorts by the same words |

**Stop and report if:**

- Manual check step 3 shows *Smith* upright with D-11's explicit font. An agent that cannot see the running app reports "italics unverified" and leaves step 3 to the user; it does not guess.
- A test outside the files in §7 fails for a reason other than a name with a `nick` part now showing quotes.
- `nameformats.Install` or `TestWesternPatternCoversEveryPartType` fails after STEP-1 for a reason other than the new type.
- T-6 returns anything but `Contains` with the `smith`/`smith` `Same` pair and `jones` in `OnlyB`, or T-6b anything but `Different` (D-5's premise is wrong).
- `StyledFormat.attributed` puts the argument anywhere but between the parentheses for `(%@)` (D-7's premise is wrong).
- The host has no Xcode (see §7): stop after STEP-4's code and report.

## 12. Open questions

None.

## 13. PR

**Title:** S14-09 — Quote nicknames and add former surnames
**Body:** objective in one sentence; what shipped against US-1, US-2 and US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-09-name-display.md`.
