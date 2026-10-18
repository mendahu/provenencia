# S11-08 — One normalization helper in Swift

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: list filters, type-select and the fake store change which rows match or group; no view, layout or copy changes)
**Board:** —
**Depends on:** S11-06 (landed [#456](https://github.com/mendahu/provenencia/pull/456)); S11-07 ([#460](https://github.com/mendahu/provenencia/pull/460), in progress at planning), whose text key (D-1, `joinedKey`) this PR mirrors in Swift
**Blocks:** —
**Branch:** `s11-08-swift-normalization`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the combo box and type-select are kit components; behavior only), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (what a text key is; read only)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [reuse across the app](../comparison-layers.md#reuse-across-the-app) "Lighter touch" (Swift filters, `FakeStore`); [S11-07 plan](S11-07-normalization-key.md) D-1, D-2 and its **Landed** note (`joinedKey`); [S11-06 plan](S11-06-migrate-reconciler.md) (grouping policy); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5; working rules "A normalized `Key` per value type", "The reconciler keeps its own decisions", "High tolerance for churn"
**Planned:** 2026-10-18 against `main` @ `db59a0a`, with S11-07 read at `s11-07-normalization-key` @ `74b3fa9`
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

Every filter-as-you-type list in the app decides "this row matches what I typed" with one Swift rule, the same rule Go's text key uses since S11-07: case, accents, punctuation, spacing and where words break do not matter. Typing `jose` finds `José` in the Sources list, the Properties search, a combo box, a table's type-select and the graph canvas rotor; typing `godfather` finds `God-father`. Today the app has four rules: the combo box folds case and accents, the Sources list folds case and accents through Foundation, and Properties, the canvas rotor and type-select only lowercase.

`FakeStore`, which stands in for Go in Swift tests, groups a Person's names by their parts under the reconciler's policy instead of by lowercase form, so tests built on it stop encoding a rule Go does not have.

**Why now.** S11-07 put Go's text key on one rule (words joined, folded at `Equivalent`). The slice Check ("`Montréal` / `Montreal` … compare the same way") reaches the app's own lists only when Swift follows. The reuse inventory scopes the Swift side to this PR.

**What stays.** The combo box's three ranking bands and its highlight, the Sources list's sort, `FakeStore`'s header sort and its search stand-in, every Go file outside one test, and the FFI. No copy changes.

## 2. User stories

**US-1.** As a researcher, when I type in a list's filter, rows match whatever their case, accents, punctuation or word breaks, the same way search does.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Sources `Smith family Bible` (`SRC-1AAAA`), `1851 census` (`SRC-4F2NP`), `José Ruiz letter` (`SRC-3CCCC`); query `jose`, then `smith-family`, then `1851census`, then `src-4f` | `José Ruiz letter`; `Smith family Bible`; `1851 census`; `1851 census` |
| AC-1.2 | Properties `Name` (`name`), `Event type` (`event_type`), `Lieu de décès` (`death_place`); search `deces`, then `eventtype`, then `EVENT-T`, then `birth` | `Lieu de décès`; `Event type`; `Event type`; no rows |
| AC-1.3 | Combo box options `Godfather`, `Godmother`; query `god-f` | `Godfather` only |
| AC-1.4 | Combo box options `Ó Riain`, `Smith`; query `o ri`, then `SMITH` | `Ó Riain`; `Smith` (today's behavior holds) |
| AC-1.5 | Table type-select over `José`, `O'Brien`, `Publication date`; prefix `jose`, then `ob`, then `date` | Row 0; row 1; no row (`-1`, still prefix-only) |
| AC-1.6 | Canvas rotor `Subjects` over items labeled `José label`, `Smith label`; VoiceOver filter `jose` | The `José label` item only |
| AC-1.7 | Any of the above with a query of spaces or punctuation only (`  `, `-`) | Every row (a list filter), or no row (type-select), as an empty query does today |

**US-2.** As the app, the Swift text key equals Go's `valuecompare.TextKey` at `Equivalent` for every shared fixture, and a drift on either side fails a test.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `TextMatch.key` of each fixture in `TextMatch.fixtures` | The fixture's key: `god-father` → `godfather`, `Mont-réal` → `montreal`, `Ó Riain` → `oriain`, `O'Brien` → `obrien`, `A+B` → `a+b`, `Straße` → `strasse`, `Søren Ærø` → `sorenaero`, `Łódź` → `lodz`, `Þórr` → `thorr`, `event_type` → `eventtype`, `SRC-4F2NP` → `src4f2np`, `München / Munich` → `munchenmunich`, `İstanbul` → `istanbul`, `!!!` → `` |
| AC-2.2 | Go reads `TextMatch.fixtures` and `TextMatch.letterFolds` from the Swift source | Each fixture's key equals `valuecompare.TextKey(input, Equivalent, matchregistry.Default().Value.Words)` (empty when that returns false); the fold table equals `matchregistry.Default().Value.Words.Folds` |
| AC-2.3 | `TextMatch.key("José", at: .same)`; `TextMatch.key("İstanbul", at: .same, locale: Locale(identifier: "tr_TR"))` | `josé`; `istanbul` |
| AC-2.4 | `TextMatch.words("Mary-Ann O'Brien/Smith", at: .same)` | `["mary", "ann", "obrien", "smith"]` |

**US-3.** As a Swift test, `FakeStore` groups a Person's names the way Go's reconciler does: by parts, case and punctuation ignored, accents kept, and a name with no parts is not a value.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | One promoted Person; name Observations on one Source: o1 `James Robins` (given `James`, surname `Robins`), o2 `Jim Robins` (given `Jim`, surname `Robins`), o3 `ROBINS, James` (surname `ROBINS`, given `James`) | Persons list header: name `James Robins`, `nameValueCount` 2 |
| AC-3.2 | Same data; the conclusion detail's `name` field | Values `James Robins` (rank 1, `kept`) and `Jim Robins` (rank 2, `outvoted`); outcomes `OBS-o1`, `OBS-o3` on rank 1, `OBS-o2` on rank 2 |
| AC-3.3 | Add o4 `J. Robins` with no parts | o4's outcome has reason `noEvidence` and no `valueRank`; the value count stays 2 |
| AC-3.4 | o1 `José Robins` (given `José`, surname `Robins`), o2 `Jose Robins` (given `Jose`, surname `Robins`) | Two values (`Equivalent` never groups: S11-06 policy); header `nameValueCount` 2 |

## 3. Scope

| In | Covered by |
| --- | --- |
| `TextMatch` *(new)*: `key`, `words`, `contains`, `letterFolds`, `fixtures`, mirroring Go's text key (D-1 – D-4) | STEP-1, AC-2.1, AC-2.3, AC-2.4 |
| Go drift guard over `TextMatch.swift` (D-5) | STEP-2, AC-2.2 |
| `PVComboBoxMatch` filter and bands on `TextMatch`; `normalize` deleted (D-6) | STEP-3, AC-1.3, AC-1.4, AC-1.7 |
| `PVTypeSelectMatcher.index` on `TextMatch` (D-7) | STEP-4, AC-1.5, AC-1.7 |
| `PropertiesModel.visibleProperties`, `SourcesModel.visibleSources`, `GraphCanvasRotorSearch.result(for:)` on `TextMatch.contains` (D-8) | STEP-5, AC-1.1, AC-1.2, AC-1.6, AC-1.7 |
| `DateValueTranscriptionMatch.monthNumber` on `TextMatch.key` (D-9) | STEP-6 |
| `FakeStore` name grouping by parts (D-10) | STEP-7, AC-3.1 – AC-3.4 |
| Docs the PR makes false (STEP-8) | STEP-8 |

| Out | Home |
| --- | --- |
| `PVComboBoxHighlight` finding a span that matches only once word breaks or punctuation are ignored (`god-f` over `Godfather` shows no highlight) | Not planned — the highlight marks the typed text; it has no span to mark when the match crosses a removed character |
| The month-name regex in `DateValueTranscriptionMatch` (`monthAlternation`, `.caseInsensitive` only), so `fevrier` is not found in a transcription | Not planned — a pattern search, not an equality rule |
| `FakeStore` search (`handleSearchHits`, `tokenizeSearch`) | Not planned — a stand-in for Go's FTS ranking; no test reads accents through it |
| `FakeStore` header sort (`x.form.lowercased()`) | Not planned — it orders, it does not decide equality (Go's `SortKey` is Out in S11-07 too) |
| `FakeStore` folding a name into another on `Contains` (`J. Robins` into `James Robins`) and outvoting | Not planned — the fake groups on `Same` only; tests that need folds use `seededConclusionDetails` |
| Particles rejoining in `TextMatch.words` (`Mc Donald` as one word) | Not planned — only the fake reads words, and its fixtures have no particles |
| Properties drawing its own "no results" state instead of `PVEmptyState` | [`ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (STEP-8 adds the item) |
| Any Go behavior change | Not this PR (map scope); STEP-2 adds a Go test only |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Combo box fold | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` `PVComboBoxMatch.normalize` | `folding(options: [.diacriticInsensitive, .caseInsensitive], locale: nil)`; keeps spaces and punctuation, so `god-f` does not find `Godfather`. Doc comment cites `ComboBox.jsx` |
| Combo box filter | same file, `PVComboBoxMatch.matches(_:normalizedQuery:)`, `filter(_:query:)`, `band(for:normalizedQuery:)` | `filter` trims and normalizes the query, keeps rows whose normalized label or subtext contains it, and sorts by band: 0 label prefix, 1 a later word prefix (label split on `" "` and `"-"`), 2 any other |
| Combo box highlight | same file, `PVComboBoxHighlight.attributed(_:query:)` | Own `range(of:options:)` search; stays (Out) |
| Type-select | `macos/App/DesignSystem/Components/FloatingMenuSelection/PVFloatingMenuSelection.swift` `PVTypeSelectMatcher.index(in:prefix:fromIndex:)` | `lowercased().hasPrefix`; `jose` misses `José`. Called by `PVSelectSession` (L200), `PVTable` (L389, through `typealias PVTableTypeSelectMatcher`), `PVContextMenu` (L52) |
| Properties search | `macos/App/Features/Properties/PropertiesModel.swift` `visibleProperties` (L99–L115) | `(label + " " + key).lowercased().contains(q)`; misses accents, and a query can match across the label/key seam |
| Sources search | `macos/App/Features/Sources/SourcesModel.swift` `visibleSources` (L212–L228) | `localizedStandardContains` on title and ref: case and accents fold, punctuation and spaces do not |
| Canvas rotor | `macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift` `GraphCanvasRotorSearch.result(for:)` (L181–L185) | `label.lowercased().contains(filterString.lowercased())` |
| Month names | `macos/App/Features/Dates/DateValueTranscriptionMatch.swift` `monthNumber(_:)` (L161–L176) | Two `folding(options: [.caseInsensitive, .diacriticInsensitive], locale: locale)` calls, each followed by trimming `.` |
| Fake name grouping | `macos/App/Platform/FakeStore.swift` `personHeaders()` (L1711–L1742), `conclusionDetail(projectDir:entityID:)` (L1764–L1800) | Each keys a name Observation by `o.nameForm.lowercased().trimmingCharacters(in: .whitespaces)`, inline, twice. `ROBINS, James` and `James Robins` are two values; a form-only name is a value |
| Go text key (after S11-07) | `core/valuecompare/text.go` `joinedKey`, `TextKey` (branch `s11-07-normalization-key`) | Drops `unicode.IsPunct` and `unicode.IsSpace` runes, lowercases each rune; at `Equivalent` also decomposes (NFD), drops `Mn` marks, and applies `Words.Folds`. Symbols (`+`) stay |
| Go word split | `core/valuecompare/words.go` `rawTokens`, `NormalizeForm`, `SplitWords` | Words break on whitespace, `Pd` dashes and `/`; other punctuation is dropped inside a word; particles rejoin |
| Go fold table | `core/matchregistry/value.go` `Default()` `Value.Words.Folds` | `ß`→`ss`, `æ`→`ae`, `œ`→`oe`, `ø`→`o`, `ł`→`l`, `đ`→`d`, `þ`→`th`, `ı`→`i` |
| Go name grouping | `core/autoreconcile/names.go` `readName`, `split`; `core/autoreconcile/pipeline.go` L101–L106 | A name's parts are grouped by trimmed part type; each type's words are keyed at `Same` as untyped words (sorted). A name with no word has no units: outcome `ReasonNoEvidence`, value `-1` |
| Swift ↔ Go guard | `api/ffi/swift_vocabulary_test.go` `swiftEnumBody`, consts `swiftVocabulary` … | Reads Swift sources so a rename on either side fails in Go |
| Mirror precedent | `macos/App/Platform/FieldSlug.swift` `FieldSlug.fixtures`; `macos/ProvenenciaTests/FieldSlugTests.swift` `fixturesMatchGo`; `core/slug/slug_test.go` | Swift mirrors Go `slug.Kebab` with a fixture table kept in sync by hand |
| Tests | `macos/ProvenenciaTests/PVComboBoxTests.swift`; `PVTableTests.swift` `PVTableTypeSelectMatcherTests`; `SourcesModelTests.swift` `searchMatchesTitleAndRefIgnoringCaseAndAccents`; `PropertiesModelTests.swift` (no search test); `GraphCanvasItemLayerTests.swift` `rotorsListItemsInHostOrderAndSelectOnMove`; `DateValueTranscriptionMatchTests.swift`; `CatalogQueryRegistryTests.swift` `personsListLoadsHeadersOfPromotedPersons`, `conclusionDetailLoadsAPersonsNameValuesAndOutcomes`, helper `nameObservation(id:subjectID:form:)`; `SearchCatalogStoreTests.swift` (person hit, form-only `James Robins`); `WorkspaceSessionTests.swift` (form-only `Jim Robins`, ~L490) | The two `CatalogQueryRegistryTests` tests assert today's lowercase-form grouping with form-only names; the other two seed form-only names that D-10 stops reconciling |

**Precedent.** `FieldSlug` (`macos/App/Platform/FieldSlug.swift`) is a Swift mirror of a Go rule with a shared fixture table: copy its shape (an enum in `Platform/`, `static let fixtures`, a doc comment naming the Go function) and its test (`fixturesMatchGo`). Do not copy its by-hand sync: D-5 makes Go read the Swift fixtures. `api/ffi/swift_vocabulary_test.go` is the precedent for Go reading a Swift file: copy `swiftEnumBody` use and its failure style. For the pbxproj, copy `ProjectSlug.swift`'s three entries (`AA00000000000000000000F0`/`F1`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **`TextMatch` *(new)*, `macos/App/Platform/TextMatch.swift`**, an `enum` with `enum Level { case same, equivalent }`. Doc comment: "The app's one rule for "the same text" in filters and the fake store. Mirrors Go `valuecompare.TextKey` (S11-07 D-1): case, spaces, punctuation and word breaks never matter; at `equivalent` accents and the registry's letter folds do not either." It lives in `Platform/` beside `FieldSlug` because design-system components, features and `FakeStore` all call it | One owner for a rule four places copy today; `Platform` already holds Go mirrors and is visible to every layer | Keeping it on `PVComboBoxMatch` and calling that from features — a feature would reach into a kit component's private rules; a `String` extension — hides which rule is applied at the call site |
| D-2 | **`static func key(_ text: String, at level: Level = .equivalent, locale: Locale = .autoupdatingCurrent) -> String`.** Steps: `text.lowercased(with: locale)`; at `.equivalent`, take `.decomposedStringWithCanonicalMapping`; walk `unicodeScalars`, skip a scalar whose `properties.isWhitespace` is true or whose `properties.generalCategory` is one of the seven punctuation categories (`connectorPunctuation`, `dashPunctuation`, `openPunctuation`, `closePunctuation`, `initialPunctuation`, `finalPunctuation`, `otherPunctuation`); at `.equivalent` also skip `.nonspacingMark` and replace a scalar found in `letterFolds` with its fold; append the rest. Returns `""` when nothing survives | Go's `joinedKey` step for step (punctuation is category P, space is White_Space, marks are Mn); the locale parameter is the map's scope line and lets Turkish `İ` lowercase to `i` at `.same` | Foundation `folding(options:)` — it keeps punctuation and spaces and its letter folds are not Go's; `CharacterSet.punctuationCharacters` — the same set, but per-scalar category reads like the Go code it mirrors |
| D-3 | **`static let letterFolds: [Unicode.Scalar: String]`** with exactly Go's eight entries (`"ß": "ss"`, `"æ": "ae"`, `"œ": "oe"`, `"ø": "o"`, `"ł": "l"`, `"đ": "d"`, `"þ": "th"`, `"ı": "i"`), one per line, comment "Go `matchregistry` `Value.Words.Folds`; `TestSwiftTextMatchMirrorsTextKey` keeps them equal". **`static let fixtures: [(text: String, key: String)]`** holds AC-2.1's fourteen pairs, one per line, as plain literals with no escapes | An explicit table makes the result the same on every macOS version; Go reads both (D-5) | Relying on Foundation's diacritic fold for `ø` / `ł` — undocumented, and no test could tie it to Go |
| D-4 | **`static func words(_ text: String, at level: Level = .equivalent, locale: Locale = .autoupdatingCurrent) -> [String]`**: split `text` on whitespace, `dashPunctuation` scalars and `/` (Go `rawTokens`), `key(_:at:locale:)` each piece, drop empty ones. **`static func contains(_ text: String, key queryKey: String, locale: Locale = .autoupdatingCurrent) -> Bool`**: `queryKey.isEmpty || key(text, locale: locale).contains(queryKey)`. Callers key the query once with `key(_:)` and pass it | `words` serves the combo box's word band (D-6) and the fake's name parts (D-10) with Go's word breaks; `contains` names the empty-query rule once (AC-1.7) and keys the query once per filter pass | A `contains(_:query:)` that keys the query per row — rekeys the same query for every row |
| D-5 | **Go drift guard `TestSwiftTextMatchMirrorsTextKey` *(new)*** in `api/ffi/swift_vocabulary_test.go`, with const `swiftTextMatch = "../../macos/App/Platform/TextMatch.swift"`. It takes `swiftEnumBody(t, swiftTextMatch, "TextMatch")`, reads fixtures with ``regexp.MustCompile(`\("([^"\\]*)", "([^"\\]*)"\)`)`` and folds with ``regexp.MustCompile(`"(.)": "([^"\\]+)"`)``, fails when either list is empty, and asserts: each fixture's key equals `valuecompare.TextKey(text, valuecompare.Equivalent, matchregistry.Default().Value.Words)` (`""` when that returns false); the fold map equals `matchregistry.Default().Value.Words.Folds` (rune → string). The Swift test `fixturesMatchGo` asserts `TextMatch.key` on the same table | A change to Go's text key or fold table fails in Go CI and a change to the Swift table fails too, with no hand sync; the file already reads Swift sources for this reason | A by-hand "keep in sync" comment as `FieldSlug` has — the rule this PR exists to unify would drift silently; passing the fold table over FFI — a new RPC for eight constants |
| D-6 | **Combo box.** `PVComboBoxMatch.normalize` is deleted. `filter(_:query:)` computes `let q = TextMatch.key(query)` and returns `options` unchanged when `q` is empty. `matches(_:normalizedQuery:)` becomes `matches(_:queryKey:)`: `TextMatch.contains(option.label, key: queryKey) || TextMatch.contains(option.subtext, key: queryKey)`. `band(for:queryKey:)`: 0 when `TextMatch.key(option.label).hasPrefix(queryKey)`; 1 when `TextMatch.words(option.label).dropFirst()` has a word with that prefix; else 2. The `matches` doc comment keeps "Substring, not prefix"; the `normalize` doc comment goes with it. `PVComboBoxHighlight` is unchanged | The combo box is the map's named source of the helper; bands keep their meaning with Go's word breaks (spaces, dashes, slashes) | Keeping `normalize` as a forwarding alias — the churn rule forbids aliases |
| D-7 | **Type-select.** `PVTypeSelectMatcher.index(in:prefix:fromIndex:)`: `let query = TextMatch.key(prefix)`; `-1` when it is empty or `values` is empty; a row matches when `TextMatch.key(values[idx]).hasPrefix(query)`. Still prefix-only, still wrapping | Type-select is a filter on typed text; `PVSelectSession`, `PVTable` and `PVContextMenu` gain it with no change of their own | Leaving type-select on `lowercased()` — the map lists `PVFloatingMenuSelection` as a filter to move |
| D-8 | **List filters.** `PropertiesModel.visibleProperties`: `let q = TextMatch.key(searchQuery)`; a property passes when `TextMatch.contains(property.label, key: q) || TextMatch.contains(property.key, key: q)` (label and key tested apart). `SourcesModel.visibleSources`: `let q = TextMatch.key(trimmedQuery)`; filter only when `q` is non-empty, with `TextMatch.contains($0.title, key: q) || TextMatch.contains($0.ref, key: q)`; `isFiltering` and `trimmedQuery` are unchanged. `GraphCanvasRotorSearch.result(for:)`: `let q = TextMatch.key(searchParameters.filterString)`; `TextMatch.contains($0.label, key: q)` | One rule per the map; testing label and key apart stops a query matching across the seam between the label and the key | `localizedStandardContains` everywhere — locale-aware but keeps punctuation, so Swift and Go would still disagree on `god-father` |
| D-9 | **Month names.** In `monthNumber(_:)`, `folded` is `TextMatch.key(name, locale: locale)` and each symbol is compared as `TextMatch.key(symbol, locale: locale) == folded`; both `.trimmingCharacters(in: CharacterSet(charactersIn: "."))` calls are deleted (the key drops `.`). `monthAlternation` is unchanged | The map lists `DateValueTranscriptionMatch` as a caller; the key does the fold and the trim in one rule | Leaving its own `folding` call — a fifth copy |
| D-10 | **Fake name grouping.** `private static func nameKey(_ parts: [CatalogNameValuePart]) -> String?` *(new, `FakeStore.swift`, beside `personHeaders`)*: group parts by `type.trimmingCharacters(in: .whitespaces)`; each group's words are `TextMatch.words(value, at: .same)` over its parts, sorted, joined with `" "`; the key is the non-empty groups as `"\(type)=\(words)"` sorted by type, joined with `"\|"`; `nil` when no group has a word. Doc comment: "Mirrors Go `autoreconcile` `nameModule.split`: by part type, untyped words at `Same`; case and punctuation never matter, accents do (S11-06 policy); a name with no part words is not a value." `personHeaders()` and `conclusionDetail` call it in place of both inline `lowercased()` keys. In `personHeaders()`, an Observation with no key is skipped (it is not a value). In `conclusionDetail`, an Observation with no key gets an outcome with reason `.noEvidence` and `valueRank` `nil`, and adds no value. The `!o.nameForm.isEmpty` filters stay. The doc comments of both functions say "by parts" instead of "case-folded form" | The reuse inventory says the fake contradicts Go's parts-only rule; Go's pipeline gives a part-less name `ReasonNoEvidence` and no value; one helper replaces two inline copies | Falling back to the form when parts are empty — encodes a rule Go does not have, which is what the map asks to remove; porting folds and outvoting — Out |
| D-11 | **Fixtures that D-10 changes get parts.** `CatalogQueryRegistryTests.nameObservation(id:subjectID:form:)` gains `parts: [CatalogNameValuePart]` and sets `nameParts`; the two tests in §4 use AC-3.1 – AC-3.3's data. The `SearchCatalogStoreTests` and `WorkspaceSessionTests` name Observations get `nameParts: [CatalogNameValuePart(value: "James", type: "given"), CatalogNameValuePart(value: "Robins", type: "surname")]` (and `Jim` for `Jim Robins`). Their assertions are unchanged | Those tests seed form-only names to get a header name; under D-10 a header name needs parts, as in Go | Changing their assertions — they test search and session behavior, not grouping |
| D-12 | **Tests:** Swift Testing per `add-swift-test`; one new file `TextMatchTests.swift`; the other cases are rows or tests in the existing files of §4; one Go test (D-5). No UI test | Matches each layer's existing tests | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `macos/App/Platform/TextMatch.swift` `TextMatch` *(new)* | file, enum | `PVComboBoxMatch.normalize` | Replaces it (deleted); the rule moves to `Platform/` so features and `FakeStore` do not import combo-box internals (D-1) |
| `TextMatch.Level` *(new)* | enum | Go `valuecompare.Relation` levels `Same` / `Equivalent` | Swift has no value-compare relation type; two cases mirror the two key levels |
| `TextMatch.key` *(new)* | func | `PVComboBoxMatch.normalize`, `localizedStandardContains`, `lowercased()` in four files | One rule for all of them (D-2) |
| `TextMatch.words` *(new)* | func | `PVComboBoxMatch.band`'s `split { $0 == " " \|\| $0 == "-" }` | Go's word breaks for both the band and the fake (D-4) |
| `TextMatch.contains(_:key:locale:)` *(new)* | func | Each filter's own `contains` | Names the empty-query rule once (D-4) |
| `TextMatch.letterFolds`, `TextMatch.fixtures` *(new)* | constants | `FieldSlug.fixtures` | Same shape; Go reads them (D-3, D-5) |
| `FakeStore.nameKey` *(new)* | func | The two inline `o.nameForm.lowercased()` keys in `FakeStore` | Replaces both (D-10) |
| `macos/ProvenenciaTests/TextMatchTests.swift` *(new)* | test file | `FieldSlugTests.swift` | A new type gets its own suite (`add-swift-test`) |
| `TestSwiftTextMatchMirrorsTextKey`, const `swiftTextMatch` *(new)* | Go test | `api/ffi/swift_vocabulary_test.go` tests | Added to that file, reusing `swiftEnumBody` (D-5) |

**Builds on.** S11-07's text key (`joinedKey`, `TextKey`) and the registry fold table; the `FieldSlug` mirror pattern; `swift_vocabulary_test.go`'s Swift-reading guard; the kit's `PVComboBoxMatch` and `PVTypeSelectMatcher` split of pure rules from views.
**Deletes.** `PVComboBoxMatch.normalize` and its doc comment; the `lowercased()` matching in `PropertiesModel`, `PVTypeSelectMatcher` and `GraphCanvasRotorSearch`; `localizedStandardContains` in `SourcesModel`; the two `folding` calls and `.` trims in `monthNumber`; the two inline name keys in `FakeStore`. Closes "In-list text matching" in `docs/ideas/shared-list-helpers.md`.
**Next adds.** A new list filter calls `TextMatch.key` once and `TextMatch.contains` per row. A Swift duplicate check (a term or field label typed in a sheet) compares `TextMatch.key` values. A change to Go's text key fails `TestSwiftTextMatchMirrorsTextKey` until Swift follows.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `api/ffi` (test only) | `swift_vocabulary_test.go`: const `swiftTextMatch` *(new)*; `TestSwiftTextMatchMirrorsTextKey` *(new)* per D-5; imports `core/matchregistry` and `core/valuecompare` |

No production Go code changes.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Text rule | `TextMatch` *(new)* | `macos/App/Platform/TextMatch.swift` | D-1 – D-4; registered in the app target (STEP-1) |
| Combo box filter | `PVComboBoxMatch` | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` | D-6; highlight unchanged |
| Type-select | `PVTypeSelectMatcher` | `macos/App/DesignSystem/Components/FloatingMenuSelection/PVFloatingMenuSelection.swift` | D-7 |
| Properties search | `PropertiesModel.visibleProperties` | `macos/App/Features/Properties/PropertiesModel.swift` | D-8 |
| Sources search | `SourcesModel.visibleSources` | `macos/App/Features/Sources/SourcesModel.swift` | D-8 |
| Canvas rotor | `GraphCanvasRotorSearch.result(for:)` | `macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift` | D-8 |
| Month names | `DateValueTranscriptionMatch.monthNumber(_:)` | `macos/App/Features/Dates/DateValueTranscriptionMatch.swift` | D-9 |
| Fake grouping | `FakeStore.nameKey`, `personHeaders()`, `conclusionDetail` | `macos/App/Platform/FakeStore.swift` | D-10 |

States: an empty or punctuation-only query shows every row in a list filter and selects nothing in type-select (AC-1.7). Large: one key per row per keystroke, as `normalize` does today. Accessibility: the rotor filter gains the same rule; no label changes.

### Strings

None — no copy change.

## 7. Steps

**STEP-1 — `TextMatch`**
Files: `macos/App/Platform/TextMatch.swift` *(new)*, `macos/ProvenenciaTests/TextMatchTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-1 – D-4. Register `TextMatch.swift` in the app target by copying `ProjectSlug.swift`'s `PBXBuildFile`, `PBXFileReference` and Platform group child with unused `AA…` IDs; register the test file per `add-swift-test`. Add T-1 – T-4.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/TextMatchTests` passes.

**STEP-2 — Go drift guard**
Files: `api/ffi/swift_vocabulary_test.go`.
Do: D-5. Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftTextMatchMirrorsTextKey` passes; changing one fixture key in `TextMatch.swift` by hand makes it fail (revert after).

**STEP-3 — Combo box**
Files: `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift`, `macos/ProvenenciaTests/PVComboBoxTests.swift`. Skill: `add-ui-component` (behavior only).
Do: D-6. Add T-6.
Check: `-only-testing:ProvenenciaTests/PVComboBoxTests` passes with every existing test unchanged; `rg -n 'PVComboBoxMatch.normalize|func normalize' macos/App/DesignSystem/Components/ComboBox` prints nothing.

**STEP-4 — Type-select**
Files: `macos/App/DesignSystem/Components/FloatingMenuSelection/PVFloatingMenuSelection.swift`, `macos/ProvenenciaTests/PVTableTests.swift`.
Do: D-7. Add T-7.
Check: `-only-testing:ProvenenciaTests/PVTableTypeSelectMatcherTests -only-testing:ProvenenciaTests/PVSelectSessionTests -only-testing:ProvenenciaTests/PVContextMenuTests` passes.

**STEP-5 — List filters**
Files: `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift`, `macos/ProvenenciaTests/PropertiesModelTests.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`, `macos/ProvenenciaTests/GraphCanvasItemLayerTests.swift`.
Do: D-8. Add T-8 – T-10.
Check: `-only-testing:ProvenenciaTests/PropertiesModelTests -only-testing:ProvenenciaTests/SourcesModelTests -only-testing:ProvenenciaTests/GraphCanvasItemLayerTests` passes; `rg -n 'localizedStandardContains|lowercased\(\)\.contains' macos/App/Features/Properties macos/App/Features/Sources macos/App/Features/GraphCanvas` prints nothing.

**STEP-6 — Month names**
Files: `macos/App/Features/Dates/DateValueTranscriptionMatch.swift`, `macos/ProvenenciaTests/DateValueTranscriptionMatchTests.swift`.
Do: D-9. Add T-11.
Check: `-only-testing:ProvenenciaTests/DateValueTranscriptionMatchTests` passes with every existing test unchanged; `rg -n 'folding\(options' macos/App/Features/Dates` prints nothing.

**STEP-7 — Fake name grouping**
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/SearchCatalogStoreTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`.
Do: D-10, D-11. Add T-12, T-13.
Check: the full `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes; `rg -n 'nameForm\.lowercased' macos/App/Platform/FakeStore.swift` prints nothing.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md` (reuse inventory "Lighter touch": the Swift filters and `FakeStore` lines say done in S11-08, on `TextMatch`, which mirrors `valuecompare.TextKey` under `TestSwiftTextMatchMirrorsTextKey`), `docs/ideas/shared-list-helpers.md` (status line: "In-list text matching" closed by S11-08; new item "Properties empty search state": Properties draws its own "no results" state instead of `PVEmptyState`), `.agents/skills/add-match-rule/SKILL.md` (one line: a change to the text key or `Words.Folds` also updates `macos/App/Platform/TextMatch.swift` *(new)*, which `TestSwiftTextMatchMirrorsTextKey` checks), `completed.md`, `deployment-plan.md` (map row status), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-08-swift-normalization.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/TextMatchTests.swift` `fixturesMatchGo` *(new)* | Each `TextMatch.fixtures` pair → `TextMatch.key(text) == key` | AC-2.1, D-2, D-3 |
| T-2 | Swift | `TextMatchTests.swift` `sameLevelKeepsAccentsAndHonorsLocale` *(new)* | `key("José", at: .same)` → `josé`; `key("İstanbul", at: .same, locale: Locale(identifier: "tr_TR"))` → `istanbul` | AC-2.3, D-2 |
| T-3 | Swift | `TextMatchTests.swift` `wordsBreakOnSpacesDashesAndSlashes` *(new)* | `words("Mary-Ann O'Brien/Smith", at: .same)` → `["mary", "ann", "obrien", "smith"]`; `words("  ")` → `[]` | AC-2.4, D-4 |
| T-4 | Swift | `TextMatchTests.swift` `containsIgnoresCaseAccentsAndWordBreaks` *(new)* | `contains("José Ruiz letter", key: key("jose"))` → true; `contains("Godfather", key: key("god-father"))` → true; `contains("Smith", key: key("jones"))` → false; `contains("Smith", key: key("-"))` → true | AC-1.7, D-4 |
| T-5 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftTextMatchMirrorsTextKey` *(new)* | Parsed fixtures → `valuecompare.TextKey` at `Equivalent` with the shipped words; parsed folds → registry `Folds` | AC-2.2, D-5 |
| T-6 | Swift | `macos/ProvenenciaTests/PVComboBoxTests.swift` `filterIgnoresWordBreaksAndPunctuation` *(new)*; existing tests unchanged | `Godfather`, `Godmother`, query `god-f` → `[Godfather]`; query `-` → both | AC-1.3, AC-1.4, AC-1.7, D-6 |
| T-7 | Swift | `macos/ProvenenciaTests/PVTableTests.swift` `PVTableTypeSelectMatcherTests.matchesThroughAccentsAndPunctuation` *(new)* | `["José", "O'Brien", "Publication date"]`: `jose` → 0, `ob` → 1, `date` → -1, `-` → -1 | AC-1.5, AC-1.7, D-7 |
| T-8 | Swift | `macos/ProvenenciaTests/PropertiesModelTests.swift` `searchMatchesLabelOrKeyIgnoringAccentsAndWordBreaks` *(new)* | `nameProperty()`, `eventTypeProperty()`, and a `CatalogProperty` built like them (id `prop-death-place`, key `death_place`, label `Lieu de décès`, value type `subject`); `searchQuery` `deces`, `eventtype`, `EVENT-T`, `birth` → as AC-1.2 | AC-1.2, D-8 |
| T-9 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `searchMatchesTitleAndRefIgnoringCaseAndAccents` (rows added) | `smith-family` → `["s1"]`; `1851census` → `["s2"]`; existing rows unchanged | AC-1.1, D-8 |
| T-10 | Swift | `macos/ProvenenciaTests/GraphCanvasItemLayerTests.swift` `rotorFilterIgnoresCaseAndAccents` *(new)* | Items `item("José", at: .zero, rotor: "Subjects")`, `item("Smith", at: .zero, rotor: "Subjects")`; `SearchParameters` with `searchDirection = .next`, `filterString = "jose"` → the `José` view; `"x"` → `nil` | AC-1.6, D-8 |
| T-11 | Swift | `macos/ProvenenciaTests/DateValueTranscriptionMatchTests.swift` `accentedMonthNameReadsAtAnyCase` *(new)* | `DateValueTranscriptionMatch.first(in: "3 FÉVRIER 1851", locale: Locale(identifier: "fr_FR"))` → `DateValueTranscriptionMatch(year: 1851, month: 2, day: 3)` | D-9 |
| T-12 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personsListLoadsHeadersOfPromotedPersons`, `conclusionDetailLoadsAPersonsNameValuesAndOutcomes` (data and expectations per AC-3.1 – AC-3.3) | As AC-3.1, AC-3.2, AC-3.3 | AC-3.1, AC-3.2, AC-3.3, D-10, D-11 |
| T-13 | Swift | `CatalogQueryRegistryTests.swift` `personsListKeepsAccentVariantsApart` *(new)* | As AC-3.4 | AC-3.4, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVComboBoxTests` (band order), `PVSelectSessionTests` and `PVContextMenuTests` (type-select), `SearchCatalogStoreTests` and `WorkspaceSessionTests` until D-11's fixtures carry parts, `PersonsListTests` and `PromoteFlowTests` if they read fake headers from form-only names.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — n/a, no copy
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no chrome
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — n/a, not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md reuse inventory, docs/ideas/shared-list-helpers.md, add-match-rule skill (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source titled `José Ruiz letter`. In the Sources list filter, type `jose`: the Source is listed. Type `ruizletter`: it is still listed.
2. Open Properties. Type `eventtype` in the search field: `Event type` is listed. Type `EVENT-T`: still listed.
3. In Properties, select the Person type and open the "Add a property to Person" combo box. Type `event-t`: `Event type` is offered.
4. In the Sources table, click a row and type `jo` quickly: the selection lands on `José Ruiz letter`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Joined keys match across word breaks (`nr` finds `John Robins`) | Accepted, as S11-07 D-4 accepted it for search: substring filters already match inside words, and the user chose that word breaks never matter |
| Foundation's `lowercased(with:)` and Go's `unicode.ToLower` differ for a few letters at `.same` (`İ` under a non-Turkish locale) | Fixtures run at `.equivalent`, where the dot is a dropped mark; the fake's `.same` keys only compare Swift with Swift |
| D-10 drops header names from more fake-backed tests than D-11 lists | STEP-7's full-suite check finds them; each gets parts the way D-11 does, with assertions unchanged |
| Keying every row per keystroke in a long list | One key per row per keystroke, as `PVComboBoxMatch.normalize` already does for combo options |

**Stop and report if:**

- S11-07 has not landed, or its **Landed** note or `core/valuecompare/text.go` shows a text key other than §4's row (word breaks not ignored, folds applied at `Same`, symbols dropped).
- `TestSwiftTextMatchMirrorsTextKey` fails on a fixture after `TextMatch.key` follows D-2 (Swift and Go disagree on a character class).
- A fake-backed test outside `CatalogQueryRegistryTests`, `SearchCatalogStoreTests`, `WorkspaceSessionTests`, `PersonsListTests` and `PromoteFlowTests` needs an assertion changed (not only parts added) after STEP-7.
- An existing `PVComboBoxTests`, `PVTableTypeSelectMatcherTests` or `DateValueTranscriptionMatchTests` case needs its expected value changed.
- `swiftEnumBody` cannot read `enum TextMatch` because the file's layout differs from its assumption (closing brace at column 0).

## 12. Open questions

None.

## 13. PR

**Title:** S11-08 — One normalization helper in Swift
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-08-swift-normalization.md`.
