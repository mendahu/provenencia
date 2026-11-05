# S14-32 — Place relationship rows read from the far end

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: rows, cards and the authoring box keep their layout; only words change — the far-end row label, and "is" → "was" in the `succeeded_by` card sentence)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-32-place-relationship-inverse-labels`
**Version bump:** patch — a researcher-visible labelling bug fix in the Mac app ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Place relationship rows use the inverse label on the far end" (reworded by this plan, see §5); bar outcome 27; dogfood note [`ux/2026-10-15-place-relationship-inverse-label.md`](../ux/2026-10-15-place-relationship-inverse-label.md) (deleted on land); [`docs/ideas/organizations.md`](../../../ideas/organizations.md) ("Structure like places": an `organization_relationship` bridge copies `part_of` / `succeeded_by`)
**Planned:** 2026-11-05 against `main` @ `c77f7392` (re-checked after S14-31 landed in #718; the files in §4 are unchanged since `b6cac73d`)
**Review:** [`S14-32-place-relationship-inverse-labels.review.md`](S14-32-place-relationship-inverse-labels.review.md) — Fits with changes — Ready with changes, 2 rounds; M-1 and M-7 answered by the user 2026-11-05; no open blocker

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

The Citation composer lists a saved place-relationship bridge under both of its places. Today both rows use the forward term: with *Calgary is part of Alberta*, Alberta's row reads "Part of Calgary", which is wrong. After this PR the far end reads the relationship from its own side: Alberta's row reads "Contains Calgary", and for *Old Town was succeeded by New Town* New Town's row reads "Succeeded Old Town" — the Place conclusion page's wording for those ends.

The composer words the two product place terms in three separate switches today, and this bug is what a missed case looks like. This PR gives each wording one home (user, 2026-11-05, review M-1 and M-7):

- **The sentence** ("Calgary is part of Alberta") lives in `ConnectionSentenceRegistry`, which Evidence graph cards already read. The composer's draft reading reads it too, so the card and the draft say the same thing. For `succeeded_by` that is "was succeeded by": the card changes from "Old Town is succeeded by New Town".
- **Composer chrome** (end titles, the term combo word, the far-end row label) lives in one wording switch keyed by property and term, which a Go test holds to every seeded place term.

**Why now.** Batch 7 dogfood fix (note 2026-10-15, bar outcome 27).

**What stays.** Every other string stays the same: the near-end row ("Part of Alberta", "Succeeded by New Town"), the box's end titles and combo words, the draft reading, the `part_of` card sentence, and person relationship rows. Place relationship terms stay product-only with no vocabulary `Inverse`.

## 2. User stories

**US-1.** As a researcher, when I look at the saved lines under a place in the Citation composer, each place-relationship row reads true from that place's side, so I can check the reading without working out the direction.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Saved bridge `part_of`, from Calgary to Alberta; Calgary's group | Label "Part of", value "Alberta" (unchanged in English) |
| AC-1.2 | Same bridge; Alberta's group | Label "Contains", value "Calgary" |
| AC-1.3 | Saved bridge `succeeded_by`, from Old Town to New Town; Old Town's group | Label "Succeeded by", value "New Town" (unchanged in English) |
| AC-1.4 | Same bridge; New Town's group | Label "Succeeded", value "Old Town" |
| AC-1.5 | A researcher-origin term keyed `part_of` labelled "Inside of", from a to b | Both rows labelled "Inside of"; the box shows no reading (unchanged) |
| AC-1.6 | A place-relationship bridge with no term (`roleTermID` empty) | Both rows labelled "Place relationship" (unchanged) |

**US-2.** As a researcher, the composer's draft reading and the Evidence graph card word one place relationship the same way.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Product `succeeded_by`, Old Town → New Town: draft reading in the box, and the graph card sentence | Both "Old Town was succeeded by New Town" (the card read "is succeeded by" before) |
| AC-2.2 | Product `part_of`, Leeds → Yorkshire: draft reading and card | Both "Leeds is part of Yorkshire" (unchanged) |
| AC-2.3 | A researcher's place term "within" (no product key): card sentence | "A is within B" from the kind's generic sentence (unchanged) |

**US-3.** As a contributor, when a product place term is added to the Go seed, a test names the composer wording it lacks.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `SeededPlaceRelationshipTypeTerm` gains a constant with no case in `CitationComposerAuthoring.termWording(propertyKey:termKey:)` | `TestSwiftVocabularyMatchesGo/every_composer-worded_term_has_composer_wording` fails naming it |
| AC-3.2 | The box for a product `part_of` / `succeeded_by` bridge | End titles "Part" / "Whole" and "Earlier place" / "Later place"; combo words "is part of" / "was succeeded by" — the same strings as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| Two `.placeRelationship` `Sentence` rows in `ConnectionSentenceRegistry.sentences`, with two `L10n.EvidenceGraph` keys | STEP-2, AC-2.1, AC-2.2, AC-2.3, D-4 |
| `readingSentence`'s place case reads the registry; `L10n.CitationComposer.isPartOf` / `wasSucceededBy` go | STEP-3, AC-2.1, AC-2.2, D-5 |
| `CitationBridgeTermWording` *(new)* and `CitationComposerAuthoring.termWording(propertyKey:termKey:)`, keyed by (property key, term key); `endTitles`, `listedLine` and `CitationBridgeSentence.displayLabel` read it | STEP-3, AC-3.2, D-6, D-7 |
| `listedLine`: far-end label when `viewing == to.id`; near end through `PropertyTermDisplay` | STEP-3, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, D-8 |
| Two `L10n.CitationComposer` far-end keys, with catalog entries | STEP-1, AC-1.2, AC-1.4, D-3 |
| Go subtest over a list of composer-worded term enums | STEP-4, AC-3.1, D-9 |
| Swift tests; `termVocabulary` gains `endLabels`; `savedRow` helper | STEP-5, T-1 – T-4 |
| Map: working rule, bar outcome 27, S14-32 scope entry say "Succeeded" (user, 2026-11-05) | STEP-7 (done by the planning commit; check) |

| Out | Home |
| --- | --- |
| Inverse terms in the seeded `place_relationship_type` vocabulary | not planned (working rule: copy from L10n; Go `TestAcyclicTermsAreSeededDirectedWithNoInverse` forbids them) |
| The registry's place card phrase reads "Is part of of" / "Is succeeded by of" for a product term (`bridgeSummaryPlaceRelationshipTypeOnly` "Is %@ of") | dogfood log — a copy fix the user has not decided; the new `Sentence` rows leave `phrase` nil |
| The `relationship` case of `readingSentence` (`areSpouses` / `areSiblings` / `isATermOf`) beside the registry's relationship sentences | dogfood log — same drift as M-7, person terms (review round 2) |
| Routing the `relationship` and `participation` cases of `listedLine`, and `endTitles`' relationship case, through `PropertyTermDisplay` | dogfood log — raw-`term.label` bypass, older than this PR (review M-2) |
| Moving older `ConnectionRow(...)` literals onto the `savedRow` helper | not planned (review M-4: new tests only) |
| Organization relationship wording | [`docs/ideas/organizations.md`](../../../ideas/organizations.md) — gets one row in D-9's list and cases in `termWording` |
| Changing the Place page's wording | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Saved-line grouping | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationComposerAuthoring.groups(observations:connections:vocabulary:nameOrders:)` | One `CitationSavedLine` per end of each saved connection, labelled by `listedLine(…, viewing: subjectID)` |
| Per-end label | same file, `private static func listedLine(from:to:rule:termID:termProperty:vocabulary:viewing:)` | `relationship` flips through `inverseTerm`; `location` picks `atWord` / `siteOfWord`; **`placeRelationship` returns `term?.label ?? bridgeKindPlace` on both ends** — the bug, bypassing `PropertyTermDisplay` |
| Place switch 1 | same file, `private static func endTitles(subjectA:subjectB:rule:term:vocabulary:property:)`, place case | `switch term?.productKey`: `partOf` → `partEnd` / `wholeEnd`; `succeededBy` → `earlierPlace` / `laterPlace`; `default` → term label both ends, or `firstPlace` / `secondPlace` |
| Place switch 2 | same file, `private static func readingSentence(from:to:rule:term:)`, place case | `partOf` → `L10n.CitationComposer.isPartOf(from, to)` ("%1$@ is part of %2$@"); `succeededBy` → `wasSucceededBy` ("%1$@ was succeeded by %2$@"); else nil. Only callers of those two keys |
| Place switch 3 | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationBridgeSentence.displayLabel(_:)`, called from `termOptions` (`ready` in scope) | `partOf` → `placeTermPartOf`; `succeededBy` → `placeTermSucceededBy`; default `term.label.localizedLowercase` |
| Sentence registry | `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` `ConnectionSentenceRegistry` (`entries`, `sentences`, `lookup(kind:term:eventType:in:)`, `sentence(kind:endpointA:endpointB:term:termKey:eventType:in:)`) | `.placeRelationship` has only its `KindEntry`: `bridgeSummaryPlaceRelationship` "%1$@ is %2$@ %3$@" with the lowercased term → "Old Town is succeeded by New Town". No place `Sentence` rows; the only rows are participation `subjectSentence`s |
| Card caller | `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` `sentence(for:in:)` | Passes `termKey` = the term's product key (nil for a researcher's term), so a kind+term `Sentence` row is picked up with no change here |
| Term display | `macos/App/Platform/PropertyTermDisplay.swift` `PropertyTermDisplay.name(term:propertyKey:)` | Product → `L10n.PropertyTerm.resource` (`placeRelationshipPartOf` "Part of", `placeRelationshipSucceededBy` "Succeeded by"); other → catalog label |
| Term keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPlaceRelationshipTypeTerm.partOf` / `.succeededBy`; `SeededPropertyKey.placeRelationshipType`. Keys repeat across properties (`SeededRoleTerm.child`, `SeededRelationshipTypeTerm.child`) | A term key alone does not name a term |
| Direction | `core/database/subjectvocab/registry.go` seed rows | `part_of` "From is part of to"; `succeeded_by` "From was succeeded by to"; no `Inverse` |
| Line-label copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.atWord`, `.siteOfWord`, `.bridgeKindPlace` | Precedent for the new keys |
| Place page | `macos/App/Features/Conclusions/PlaceDetailContent.swift` `relationshipSections(summary:)`; `L10n.Conclusions.placeContains` (`conclusions.place.contains`), `.placeSucceeded` (`conclusions.place.succeeded`) | Far-end wording this PR matches |
| Same-place link | `CitationComposerAuthoring.resolvePair` | Refuses `endA == endB` |
| Go guard | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`: `seededTermEnums`, subtest "every seeded term has wording" (reads `swiftFuncBody(t, swiftL10n, "static func resource(propertyKey")` with `wordingCase`); helpers `swiftEnumBody`, `swiftNamedKey`; path consts `swiftVocabulary`, `swiftL10n` in the `const (…)` block | Nothing checks the composer's place switches |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `savedBridgeIsListedUnderBothEnds`, `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` (calls `CitationBridgeSentence.displayLabel(theirs)`), helpers `termVocabulary(propertyKey:terms:endTypeKey:)`, `readyBridge(_:termID:)` (ends `"a"`, `"b"`); `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift` `aTermReadsLowercaseInsideEverySentenceAndPhrase` ("Leeds is part of Yorkshire" with no term key) | No test covers a place saved line or a place `Sentence` row |

**Precedent.** `4745bcb7` added `atWord` / `siteOfWord` and the `location` case of `listedLine` (a per-end label in `L10n.CitationComposer`). `L10n.PropertyTerm.resource(propertyKey:termKey:)` is the shape for `termWording`: a `switch (propertyKey, termKey)` over `Seeded…` constants, which the Go test reads with `wordingCase`. `subjectSentence(…)` rows are the shape for the new registry rows.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Far end of `succeeded_by` reads **"Succeeded"** | User, 2026-11-05: match the Place page | "Preceded by" |
| D-2 | Far end of `part_of` reads **"Contains"** | Place page and working rule | — |
| D-3 | `L10n.CitationComposer.containsWord` (`citationComposer.containsWord`, "Contains") and `succeededWord` (`citationComposer.succeededWord`, "Succeeded"), right after `siteOfWord`; comments name the Place page key they match | Composer line labels live there; cross-reference keeps them in step (M-5) | Reusing `L10n.Conclusions` keys — they describe a page heading and role |
| D-4 | `ConnectionSentenceRegistry.sentences` gains two rows: `Sentence(kind: .placeRelationship, term: SeededPlaceRelationshipTypeTerm.partOf, eventType: nil, sentence: { Text.bridgeSentencePlacePartOf(part: $0.a, whole: $0.b) })` and the same for `succeededBy` → `Text.bridgeSentencePlaceSucceededBy(earlier: $0.a, later: $0.b)`; `phrase` nil. New `L10n.EvidenceGraph` funcs *(new)*: `bridgeSentencePlacePartOf` (`evidenceGraph.bridge.sentence.placePartOf`, "%1$@ is part of %2$@") and `bridgeSentencePlaceSucceededBy` (`evidenceGraph.bridge.sentence.placeSucceededBy`, "%1$@ was succeeded by %2$@") | User, 2026-11-05 (M-7): one sentence per bridge term; "was" | A `reading` closure in the composer table — a second sentence registry; keeping "is" — user chose "was" |
| D-5 | `readingSentence`'s place case: `guard let term, let key = term.productKey, ConnectionSentenceRegistry.lookup(kind: .placeRelationship, term: key, eventType: nil) != nil else { return nil }`, then `ConnectionSentenceRegistry.sentence(kind: .placeRelationship, endpointA: from.label, endpointB: to.label, term: term.label, termKey: key)`. `L10n.CitationComposer.isPartOf` / `wasSucceededBy` and their catalog keys are deleted | The draft shows a reading only where a product sentence exists, as today (a researcher's term still shows none, AC-1.5) | Calling `sentence` unconditionally — a researcher's term would gain "a is inside of b" in the box, a behavior change nobody asked for |
| D-6 | `struct CitationBridgeTermWording: Sendable` *(new)* `{ fromEnd, toEnd, comboWord, farEnd: LocalizedStringResource }`, declared directly above `enum CitationComposerAuthoring`, with two `static let` values `placePartOf` (`partEnd`, `wholeEnd`, `placeTermPartOf`, `containsWord`) and `placeSucceededBy` (`earlierPlace`, `laterPlace`, `placeTermSucceededBy`, `succeededWord`). On `CitationComposerAuthoring`, as its last members after `inverseTerm`: `static func termWording(propertyKey: String, termKey: String) -> CitationBridgeTermWording?` — a `switch (propertyKey, termKey)` with one line per term, written `case (SeededPropertyKey.placeRelationshipType, SeededPlaceRelationshipTypeTerm.partOf): return .placePartOf`, default nil — and `static func termWording(_ term: CatalogPropertyTerm?, property: CatalogProperty?) -> CitationBridgeTermWording?` returning nil unless both `term?.productKey` and `property?.key` are present | User (M-1): one home for the composer's chrome; keyed by property and term like `L10n.PropertyTerm.resource` (M-6), so an organization `part_of` cannot match a place's wording; neutral name (M-6) | A dictionary keyed by bare term key — collides across properties (M-6); `[String: …]` keyed by a pair — Swift tuples are not `Hashable` |
| D-7 | Readers: `endTitles` place case → `if let w = termWording(term, property: property) { return (L10n.string(w.fromEnd), L10n.string(w.toEnd), true) }`, else today's `default` branch. `CitationBridgeSentence.displayLabel(_ term:, property: CatalogProperty?)` → `CitationComposerAuthoring.termWording(term, property: property).map { L10n.string($0.comboWord) } ?? term.label.localizedLowercase`; `termOptions` passes `ready.termProperty` | One lookup for every reader | — |
| D-8 | `listedLine` place case: when `viewing == to.id` and `termWording(term, property: termProperty)` is non-nil → `L10n.string(w.farEnd)`; otherwise `term.map { PropertyTermDisplay.name(term: $0, propertyKey: SeededPropertyKey.placeRelationshipType) } ?? L10n.string(L10n.CitationComposer.bridgeKindPlace)`. Value stays `other.label` | `from` is the part / earlier place; both ends localize (M-2); a researcher's term keeps its label | Raw `term.label` on the near end |
| D-9 | Go: add `swiftCitationComposerAuthoring = "../../macos/App/Features/CitationComposer/CitationComposerAuthoring.swift"` to the path `const (…)` block beside `swiftL10n`, then `gofmt -w api/ffi/swift_vocabulary_test.go`. New subtest `every composer-worded term has composer wording`, after "every seeded term has wording": `composerWordedTermEnums := []struct{ enum, property string }{{"SeededPlaceRelationshipTypeTerm", "place_relationship_type"}}`; read `swiftFuncBody(t, swiftCitationComposerAuthoring, "static func termWording(propertyKey")`; collect `wordingCase` matches as `enum.name`; `t.Errorf` for each `swiftNamedKey` name in each listed enum without a case | Same shape as its sibling; organizations become one row (M-6); runs off a Mac | A Swift test over a hand-kept list |
| D-10 | Swift tests call `CitationComposerAuthoring.groups` / `resolvePair` with hand-built vocabularies. `termVocabulary` gains a last parameter `endLabels: (a: String, b: String) = ("a", "b")` (existing calls compile unchanged); `savedRow(termProperty:termID:bridgeTypeKey:)` *(new)* builds the saved row. The researcher-term case extends `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` | `groups` is pure (M-3, M-4) | A model test through FakeStore |

**Rules this plan bends.** Working rule "(Batch 7) Place relationship rows use the inverse label on the far end" named "preceded by" while also saying "matching the Place conclusion page" ("Succeeded"). User answer 2026-11-05: "Succeeded". The planning commit rewords the rule, bar outcome 27 and the S14-32 scope entry; STEP-7 checks them.

## 6. Design

### Codebase fit

Searched: `inverse`, `SeededPlaceRelationshipTypeTerm`, `productKey`, `PropertyTermDisplay`, `ConnectionSentenceRegistry`, `isPartOf`, `Contains`, `Succeeded`, `atWord`, `viewing` in `macos/App`; `seededTermEnums`, `wordingCase` in `api/ffi`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Two `.placeRelationship` rows in `ConnectionSentenceRegistry.sentences` | registry entries | `subjectSentence` rows | Extends the registry (D-4) |
| `L10n.EvidenceGraph.bridgeSentencePlacePartOf`, `.bridgeSentencePlaceSucceededBy` *(new)* | key | `L10n.CitationComposer.isPartOf`, `.wasSucceededBy` (deleted) | The sentence moves to the registry's namespace |
| `CitationBridgeTermWording` *(new)*, `CitationComposerAuthoring.termWording(propertyKey:termKey:)`, `termWording(_:property:)` *(new)* | type / func | the three switches in §4; `L10n.PropertyTerm.resource` | Replaces the switches; keyed like `resource`; holds composer chrome, not term names |
| `L10n.CitationComposer.containsWord`, `.succeededWord` *(new)* | key | `L10n.Conclusions.placeContains`, `.placeSucceeded` | D-3 |
| `swiftCitationComposerAuthoring`, `composerWordedTermEnums`, subtest *(new)* | test | "every seeded term has wording" | Different table, same shape; reuses `wordingCase` |
| `savedRow(…)` *(new)*, `termVocabulary(… endLabels:)` | test helper | `ConnectionRow(...)` literals; `termVocabulary` | D-10 |
| `placeRelationshipRowsReadFromEachEnd`, `placeTermWordingKeepsTheBoxWording`, `placeTermsHaveTheirOwnSentence` *(new)* | test | `savedBridgeIsListedUnderBothEnds`, `aTermReadsLowercaseInsideEverySentenceAndPhrase` | New behavior per test |

**Builds on.** `ConnectionSentenceRegistry` (one sentence per term); `L10n.PropertyTerm.resource`'s keyed switch; `PropertyTermDisplay`; `TestSwiftVocabularyMatchesGo`'s Swift-table checks; `4745bcb7`'s per-end labels.
**Deletes.** The place cases' `switch term?.productKey` in `endTitles`, `readingSentence` and `CitationBridgeSentence.displayLabel`; `L10n.CitationComposer.isPartOf` and `.wasSucceededBy` with catalog keys `citationComposer.isPartOf` and `citationComposer.wasSucceededBy`; the raw `term?.label` read in `listedLine`'s place case.
**Next adds.** The place term set is closed ([`properties-and-terms.md`](../../../ideas/properties-and-terms.md)). Organizations ([`organizations.md`](../../../ideas/organizations.md)) add `Sentence` rows of their kind, `termWording` cases under their property key, and one `composerWordedTermEnums` row.

### Reach

None — no new path into existing code, no new state reachable.

### Scale

None — no write, list or loop changes; a switch replaces a switch, and the registry lookup scans a list of eight rows.

### Edges

- **Assumptions.** `from` is the part / earlier place (Go seed descriptions). Held for both ends by T-1, for the box by T-3, for the sentence by T-4.
- **Neighbors.** Near end of each product term (T-1); a researcher's term keyed `part_of` — no wording, label both ends, no reading (T-2); no term (T-1); a researcher's place term on a card keeps the generic sentence (T-4); `part_of` on a card keeps "is part of" (T-4, and `aTermReadsLowercaseInsideEverySentenceAndPhrase`); a role term's combo word stays lowercased (T-3); a role term key `child` does not match a place wording (keyed by property, T-3).
- **Removed inputs.** None — label and sentence choices.
- **Real data.** `place_relationship_type` seeds exactly `part_of` and `succeeded_by`; both get a `Sentence` row and a `termWording` case; the Go subtest reads the live constants.
- **Untrusted size.** None.

### Schema

None — no catalog change.

### Go core

None — the one Go change is a test (STEP-4).

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Place sentences | `ConnectionSentenceRegistry.sentences` | `ConnectionSentenceRegistry.swift` | D-4; after the `subjectSentence` rows |
| Composer wording | `CitationBridgeTermWording`, `termWording` | `CitationComposerAuthoring.swift` | D-6 |
| Readers | `endTitles`, `readingSentence`, `listedLine`, `CitationBridgeSentence.displayLabel` | `CitationComposerAuthoring.swift`, `CitationComposerAuthoringBox.swift` | D-5, D-7, D-8 |

States: unchanged. Accessibility: `CitationSavedLine.body` (VoiceOver) uses the same label; the card's VoiceOver label is the sentence.

### Strings

Keys *(new)*, each with a catalog entry in the shape of its neighbors (`comment`, `extractionState: manual`, `en` `stringUnit` `state: translated`):

- `L10n.CitationComposer.containsWord` → `citationComposer.containsWord` = "Contains"; comment "Place-relationship line label on the whole place, read from its side of part of. Matches conclusions.place.contains". In `L10n.swift` and the catalog immediately after `siteOfWord`.
- `L10n.CitationComposer.succeededWord` → `citationComposer.succeededWord` = "Succeeded"; comment "Place-relationship line label on the later place, read from its side of succeeded by. Matches conclusions.place.succeeded". After `containsWord`.
- `L10n.EvidenceGraph.bridgeSentencePlacePartOf(part:whole:)` → `evidenceGraph.bridge.sentence.placePartOf` = "%1$@ is part of %2$@"; comment "Connection sentence for part of; arguments are the part, then the whole". In `L10n.swift` and the catalog immediately after `bridgeSentenceBurialSubject` / `evidenceGraph.bridge.sentence.burialSubject`.
- `L10n.EvidenceGraph.bridgeSentencePlaceSucceededBy(earlier:later:)` → `evidenceGraph.bridge.sentence.placeSucceededBy` = "%1$@ was succeeded by %2$@"; comment "Connection sentence for succeeded by; arguments are the earlier place, then the later". After `placePartOf`.

Removed: `L10n.CitationComposer.isPartOf` / `citationComposer.isPartOf`, `L10n.CitationComposer.wasSucceededBy` / `citationComposer.wasSucceededBy`.

Copy that changes for an existing action: the `succeeded_by` card sentence, "%1$@ is succeeded by %2$@" (via `evidenceGraph.bridge.summary.placeRelationship`) → "%1$@ was succeeded by %2$@" (user, 2026-11-05).

## 7. Steps

**STEP-1 — Add the keys**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: add the four keys in §6 Strings, where it says.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — Place sentences in the registry**
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`.
Do: append the two rows in D-4 to `sentences`, after the `subjectSentence` rows.
Check: `rg -n "placeRelationship, term: SeededPlaceRelationshipTypeTerm" macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift` prints two lines.

**STEP-3 — Composer wording, one lookup, four readers**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: add `CitationBridgeTermWording` and `termWording` (D-6). Rewrite the place cases of `endTitles`, `readingSentence`, `listedLine` and `CitationBridgeSentence.displayLabel` (D-5, D-7, D-8); update `termOptions`' call. Delete `isPartOf` / `wasSucceededBy` and their catalog entries. No other case changes.
Check: `rg -n "SeededPlaceRelationshipTypeTerm\.(partOf|succeededBy)" macos/App/Features` prints only the two `termWording` case lines and the two registry rows; `rg -n "CitationComposer\.(isPartOf|wasSucceededBy)|citationComposer\.(isPartOf|wasSucceededBy)" macos` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes; on macOS the app builds with no new warning. Off macOS, report the build as **not run**.

**STEP-4 — Go guard**
Files: `api/ffi/swift_vocabulary_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: D-9; `t.Fatal` when the body has no `wordingCase` match.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run 'TestSwiftVocabularyMatchesGo'` passes; removing the `succeededBy` case locally makes it fail naming `SeededPlaceRelationshipTypeTerm.succeededBy` (revert after); `gofmt -l api/ffi` prints nothing.

**STEP-5 — Swift tests**
Files: `macos/ProvenenciaTests/CitationComposerModelTests.swift`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift`. Skill: `add-swift-test`.
Do: in `CitationComposerModelTests`'s "term keys honour the term's origin" section (the helpers are private instance methods of that struct), add `endLabels` to `termVocabulary` (subjects keep ids `"a"` / `"b"`; labels from `endLabels`) and add `savedRow(termProperty:termID:bridgeTypeKey:)` returning `ConnectionRow(id: UUID(), isPending: false, fromSubjectID: "a", toSubjectID: "b", fromLabel: "", toLabel: "", bridgeTypeKey: bridgeTypeKey, bridgeID: "bridge-1", sentence: "", termProperty: termProperty, roleTermID: termID, termTouched: false, isSaving: false)`. Callers pass `EvidenceBridgeKind.placeRelationship.rawValue` and `vocabulary.propertiesByID["p-place_relationship_type"]`; term ids are `term-<origin>-<key>`. Add T-1 and T-3; extend `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` for T-2 (its `displayLabel(theirs)` call gains `property: nil`, and the product half passes the vocabulary's property). Build role terms directly with `CatalogPropertyTerm(id:propertyID:key:origin:label:description:)` as that test does, and a role property with `CatalogProperty(id:key:origin:label:description:valueType:)`. Add T-4 to `ConnectionSentenceRegistryTests`. Find groups by `subjectID`, never by index. Compare with `L10n.string(…)` or the L10n format funcs named in §8, not English literals.
Check: `CitationComposerModelTests` and `ConnectionSentenceRegistryTests` pass in `xcodebuild test` (`scripts/check-pr.sh`). Off macOS (`check-pr.sh` skips `xcode-test`), report T-1 – T-4 and the Swift build as **not run** and name them in the PR description to be run on a Mac before merge.

**STEP-6 — Version bump**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`. Rule: `.cursor/rules/versioning.mdc`.
Do: raise PATCH by one from what `main` holds when you branch (0.3.4 → 0.3.5 at the planning base). If `main` moves past your number before merge, rebase and take the next PATCH; all three places always agree.
Check: `rg -n "0\.3\.4" VERSION core/version.go macos/Provenencia.xcodeproj/project.pbxproj` prints nothing (old number substituted if `main` moved).

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-14/deployment-plan.md` (map row status, "Where the notes went" row), `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/ux/2026-10-15-place-relationship-inverse-label.md` (delete), this plan.
Do: per the Definition of done. Check the working rule, bar outcome 27 and the scope entry say "Succeeded".
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-32-place-relationship-inverse-labels.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Read a place relationship from the far place's side in the composer's saved lines.", "Word a place succession as 'was succeeded by' on graph cards, as the composer does.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CitationComposerModelTests.swift` `placeRelationshipRowsReadFromEachEnd` *(new)* | (a) product `part_of` "Part of", Calgary → Alberta: `a` label `L10n.string(L10n.PropertyTerm.placeRelationshipPartOf)`, value "Alberta"; `b` label `L10n.string(L10n.CitationComposer.containsWord)`, value "Calgary". (b) product `succeeded_by` "Succeeded by", Old Town → New Town: `a` label `PropertyTerm.placeRelationshipSucceededBy`, value "New Town"; `b` label `succeededWord`, value "Old Town". (c) `roleTermID: ""`: both labels `bridgeKindPlace` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.6, D-1, D-2, D-8 |
| T-2 | Swift | `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` (extended) | user-origin `part_of` "Inside of": `groups` lines for `a` and `b` both "Inside of"; `ready.reading == nil` (kept) | AC-1.5, D-5, D-8 |
| T-3 | Swift | `placeTermWordingKeepsTheBoxWording` *(new)* | via `readyBridge`: product `part_of` → `endATitle` `partEnd`, `endBTitle` `wholeEnd`, `reading == L10n.EvidenceGraph.bridgeSentencePlacePartOf(part: "a", whole: "b")`; product `succeeded_by` → `earlierPlace` / `laterPlace`, `bridgeSentencePlaceSucceededBy(earlier: "a", later: "b")`; `displayLabel(_:property:)` → `placeTermPartOf` / `placeTermSucceededBy`; a product role term key `child` "Child" with a `role` property → "child" | AC-2.1, AC-2.2, AC-3.2, D-5, D-6, D-7 |
| T-4 | Swift | `ConnectionSentenceRegistryTests.swift` `placeTermsHaveTheirOwnSentence` *(new)* | `sentence(kind: .placeRelationship, endpointA: "Old Town", endpointB: "New Town", term: "Succeeded by", termKey: "succeeded_by")` → `bridgeSentencePlaceSucceededBy(earlier: "Old Town", later: "New Town")`; same with `part_of` → `bridgeSentencePlacePartOf`; `term: "within", termKey: nil` → `L10n.EvidenceGraph.bridgeSummaryPlaceRelationship(from: "A", type: "within", to: "B")` | AC-2.1, AC-2.2, AC-2.3, D-4 |
| T-5 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo/every_composer-worded_term_has_composer_wording` *(new)* | live constants → each has a `termWording` case | AC-3.1, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` (`aProductTermsInverseIsNeverAResearchersTerm`, `aResearchersSpouseTermGetsNoSpouseSentence`), `ConnectionSentenceRegistryTests`, `EvidenceBridgeEdgeSummaryTests`, `TestSwiftVocabularyMatchesGo`, `scripts/check-localizable-xcstrings.py`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no new component)
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; "Where the notes went" row says the note was deleted on land; completed.md: one short paragraph of what shipped, naming the card copy change
- [ ] Dogfood note ux/2026-10-15-place-relationship-inverse-label.md deleted
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check, the tests not run off a Mac
```

## 10. Manual check

1. Open a project with a Source and an Evidence graph holding Calgary and Alberta.
2. In the Citation composer, connect Calgary → Alberta with Part of. The box shows Part / Whole and "Calgary is part of Alberta" as before. Save. Calgary's group reads "Part of Alberta"; Alberta's reads "Contains Calgary". The graph card reads "Calgary is part of Alberta".
3. Add Old Town and New Town; connect Old Town → New Town with Succeeded by. The box reads "Old Town was succeeded by New Town". Save. Old Town reads "Succeeded by New Town"; New Town reads "Succeeded Old Town". The graph card reads "Old Town was succeeded by New Town".
4. After Promote, Alberta's and New Town's Place pages use the same "Contains" and "Succeeded".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving three switches changes a string the box shows today | T-3 pins each against its existing key |
| A card or test elsewhere asserts "is succeeded by" | `rg -n "is succeeded by" macos` before STEP-2; a hit outside L10n is updated to the new sentence in STEP-5 and named in the PR |

**Stop and report if:**

- `listedLine`'s `placeRelationship` case no longer returns `term?.label` on both ends.
- The Go seed descriptions for `part_of` / `succeeded_by` no longer say "From is part of to" / "From was succeeded by to".
- `place_relationship_type` gains a product term beyond `part_of` and `succeeded_by` at your base.
- `EvidenceBridgeEdgeSummary.sentence(for:in:)` no longer passes the product term key as `termKey` (then the cards do not pick up D-4).
- `ConnectionSentenceRegistry.lookup` or `.sentence` changes signature.

## 12. Open questions

None.

## 13. PR

**Title:** S14-32 — Place relationship rows read from the far end
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; the card copy change; deviations from the plan; manual check; tests not run off a Mac; `Plan: docs/deployment-plan/spike-14/prs/S14-32-place-relationship-inverse-labels.md`.
