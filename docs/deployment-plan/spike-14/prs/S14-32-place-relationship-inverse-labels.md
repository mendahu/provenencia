# S14-32 — Place relationship rows read from the far end

**Status:** Planned
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the saved-line row keeps its layout, mark and type; only the label word on the far end changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-32-place-relationship-inverse-labels`
**Version bump:** patch — a researcher-visible labelling bug fix in the Mac app ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Place relationship rows use the inverse label on the far end" (reworded by this plan, see §5); bar outcome 27; dogfood note [`ux/2026-10-15-place-relationship-inverse-label.md`](../ux/2026-10-15-place-relationship-inverse-label.md) (deleted on land); [`docs/ideas/properties-and-terms.md`](../../../ideas/properties-and-terms.md) ("`place_relationship_type` takes no new terms at all")
**Planned:** 2026-11-05 against `main` @ `b6cac73d`
**Review:** [`S14-32-place-relationship-inverse-labels.review.md`](S14-32-place-relationship-inverse-labels.review.md) — Fits with changes — Ready with changes, 1 round; M-1 answered by the user 2026-11-05; no open blocker

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

The composer words the two product place terms in three separate switches today, and this bug is what a missed case looks like. This PR puts every composer wording of a place term into one table, read by all four places that word one, and a Go test fails when a seeded place term has no entry (user, 2026-11-05, review M-1).

**Why now.** Batch 7 dogfood fix (note 2026-10-15, bar outcome 27).

**What stays.** Every string the composer shows today for a place relationship stays the same, except the far-end row label and the near-end row label's source (now `PropertyTermDisplay`, same English). Evidence graph cards, the Place page, and person relationship rows are unchanged. Place relationship terms stay product-only with no vocabulary `Inverse`.

## 2. User stories

**US-1.** As a researcher, when I look at the saved lines under a place in the Citation composer, each place-relationship row reads true from that place's side, so I can check the reading without working out the direction.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Saved bridge `part_of`, from Calgary to Alberta; Calgary's group | Label "Part of", value "Alberta" → "Part of Alberta" (unchanged in English) |
| AC-1.2 | Same bridge; Alberta's group | Label "Contains", value "Calgary" → "Contains Calgary" |
| AC-1.3 | Saved bridge `succeeded_by`, from Old Town to New Town; Old Town's group | "Succeeded by New Town" (unchanged in English) |
| AC-1.4 | Same bridge; New Town's group | "Succeeded Old Town" |
| AC-1.5 | A researcher-origin term keyed `part_of` labelled "Inside of", from a to b | Both rows read "Inside of" + the other place |
| AC-1.6 | A place-relationship bridge with no term (`roleTermID` empty) | Both rows read "Place relationship" + the other place (unchanged) |

**US-2.** As a contributor, when a product place term is added to the Go seed, a test names every composer wording it lacks, so the composer cannot fall back to the forward label without failing CI.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `SeededPlaceRelationshipTypeTerm` gains a constant with no entry in `CitationComposerAuthoring.placeTermWordings` | `TestSwiftVocabularyMatchesGo/every_seeded_place_term_has_composer_wording` fails naming the constant |
| AC-2.2 | The authoring box for a product `part_of` / `succeeded_by` bridge | End titles "Part" / "Whole" and "Earlier place" / "Later place", sentence verb "is part of" / "was succeeded by", reading "A is part of B" / "A was succeeded by B" — the same strings as today, now read from the table |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CitationPlaceTermWording` *(new)* and one table `CitationComposerAuthoring.placeTermWordings` keyed by product term key, with a lookup `placeWording(_:)` | STEP-2, D-4 |
| `endTitles`, `readingSentence`, `listedLine` and `CitationBridgeSentence.displayLabel` read the table; their place-term switches go | STEP-2, AC-2.2, D-5 |
| `listedLine` picks the far-end label when `viewing == to.id`; the near end reads `PropertyTermDisplay` | STEP-2, AC-1.1 – AC-1.6, D-6, D-7 |
| Two L10n keys for the far-end labels, with String Catalog entries | STEP-1, AC-1.2, AC-1.4 |
| Go subtest: every seeded place term constant has a table entry | STEP-3, AC-2.1 |
| Swift tests over `CitationComposerAuthoring.groups` and the extended researcher-term test; `termVocabulary` gains `endLabels`; a `savedRow` helper | STEP-4, T-1 – T-4 |
| Map: working rule, bar outcome 27 and S14-32 scope entry say "Succeeded" for the far end of `succeeded_by` (user, 2026-11-05) | STEP-6 (done by the planning commit; check) |

| Out | Home |
| --- | --- |
| Inverse terms in the seeded `place_relationship_type` vocabulary | not planned (working rule: copy from L10n; Go `TestAcyclicTermsAreSeededDirectedWithNoInverse` forbids them) |
| Person relationship inverses | not planned — `inverseTerm` and its Go guard already handle them |
| Routing the `relationship` and `participation` cases of `listedLine`, and `endTitles`' relationship case, through `PropertyTermDisplay` | dogfood log — same raw-`term.label` bypass, older than this PR (review M-2) |
| Moving older `ConnectionRow(...)` literals in `CitationComposerModelTests` onto the new `savedRow` helper | not planned (review M-4: new tests only) |
| Evidence graph bridge cards, `ConnectionSentenceRegistry` | not planned — they word the whole sentence, not one end |
| Changing the Place page's "Succeeded" / "Succeeded by" / "Contains" wording | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Saved-line grouping | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationComposerAuthoring.groups(observations:connections:vocabulary:nameOrders:)` | One `CitationSavedLine` per end of each saved connection, labelled by `listedLine(…, viewing: subjectID)`; `body` joins label and value |
| Per-end label | same file, `private static func listedLine(from:to:rule:termID:termProperty:vocabulary:viewing:)` | `relationship` flips through `inverseTerm` when `viewing == to.id`; `location` picks `atWord` / `siteOfWord`; **`placeRelationship` returns `term?.label ?? bridgeKindPlace` on both ends** — the bug, and it bypasses `PropertyTermDisplay` |
| Place switch 1 | same file, `private static func endTitles(subjectA:subjectB:rule:term:vocabulary:property:)`, `placeRelationship` case | `switch term?.productKey`: `partOf` → `partEnd` / `wholeEnd`; `succeededBy` → `earlierPlace` / `laterPlace`; `default` → term label both ends, or `firstPlace` / `secondPlace` |
| Place switch 2 | same file, `private static func readingSentence(from:to:rule:term:)`, `placeRelationship` case | `partOf` → `isPartOf(from, to)`; `succeededBy` → `wasSucceededBy(from, to)`; else nil |
| Place switch 3 | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationBridgeSentence.displayLabel(_:)` | `partOf` → `placeTermPartOf`; `succeededBy` → `placeTermSucceededBy`; default `term.label.localizedLowercase` (used for every bridge kind's term combo) |
| Person inverse | `CitationComposerAuthoring.inverseTerm(of:property:vocabulary:)` | `SeededRelationshipTypeTerm.inverse`; person terms only. Not touched |
| Term display | `macos/App/Platform/PropertyTermDisplay.swift` `PropertyTermDisplay.name(term:propertyKey:)` | Product term → `L10n.PropertyTerm.resource` (`placeRelationshipPartOf` "Part of", `placeRelationshipSucceededBy` "Succeeded by"); other → catalog label |
| Place term keys | `macos/App/Platform/ConclusionVocabulary.swift` `SeededPlaceRelationshipTypeTerm.partOf`, `.succeededBy`; `SeededPropertyKey.placeRelationshipType` | Constants |
| Direction | `core/database/subjectvocab/registry.go` seed rows | `part_of` "From is part of to"; `succeeded_by` "From was succeeded by to"; no `Inverse` on either |
| Product key | `macos/App/Platform/GenealogyStore.swift` `CatalogPropertyTerm.productKey` | `key` for a product term, nil otherwise |
| Line-label copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.atWord`, `.siteOfWord`, `.bridgeKindPlace` | Precedent for the new keys |
| Place page | `macos/App/Features/Conclusions/PlaceDetailContent.swift` `relationshipSections(summary:)`; `L10n.Conclusions.placeContains` (`conclusions.place.contains`), `.placeSucceeded` (`conclusions.place.succeeded`) | The far-end wording this PR matches |
| Same-place link | `CitationComposerAuthoring.resolvePair` | Refuses `endA == endB`, so `from.id == to.id` never reaches `groups` |
| Go guard | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`: `seededTermEnums` (includes `SeededPlaceRelationshipTypeTerm`), subtests "every seeded term is in Swift", "every seeded term has wording" (`wordingCase` over `L10n.PropertyTerm.resource`); path vars `swiftVocabulary`, `swiftL10n` | Every seeded place term has a Swift constant and a `PropertyTerm` resource; nothing checks the composer's place switches |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `savedBridgeIsListedUnderBothEnds`, `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel`, helpers `termVocabulary(propertyKey:terms:endTypeKey:)` (labels each end with its id), `readyBridge(_:termID:)` | No test covers a place-relationship saved line |

**Precedent.** `4745bcb7` added `atWord` / `siteOfWord` and the `location` case of `listedLine`: a per-end label chosen by `viewing`, copy in `L10n.CitationComposer`. The "every seeded term has wording" subtest is the shape for the new Go guard: read a Swift table with a regexp and check it against the constants in `SeededPlaceRelationshipTypeTerm`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The far end of `succeeded_by` reads **"Succeeded"** | User, 2026-11-05: match the Place page's `placeSucceeded` | "Preceded by" (the map's first wording, the dogfood note) |
| D-2 | The far end of `part_of` reads **"Contains"** | Matches the Place page's `placeContains` and the working rule | — |
| D-3 | Two keys in `L10n.CitationComposer`, right after `siteOfWord`: `containsWord` (`citationComposer.containsWord`, "Contains") and `succeededWord` (`citationComposer.succeededWord`, "Succeeded"). Each comment names the Place page key it must match | Composer line labels live in its namespace; a translator sees a row-label comment; the cross-reference keeps the two in step (review M-5) | Reusing `L10n.Conclusions.placeContains` / `placeSucceeded` — their comments describe a page heading and role |
| D-4 | One table *(new)*, in `CitationComposerAuthoring.swift`: `struct CitationPlaceTermWording { fromEnd, toEnd, sentenceTerm, farEnd: LocalizedStringResource; reading: @Sendable (String, String) -> String }`, and `static let placeTermWordings: [String: CitationPlaceTermWording]` on `CitationComposerAuthoring`, keyed by `SeededPlaceRelationshipTypeTerm` constants, each entry written `SeededPlaceRelationshipTypeTerm.<name>: CitationPlaceTermWording(` on one line. Lookup `static func placeWording(_ term: CatalogPropertyTerm?) -> CitationPlaceTermWording?` returns `term?.productKey.flatMap { placeTermWordings[$0] }` | User, 2026-11-05 (review M-1): one table replaces three switches and the proposed fourth; "from is the part / earlier place" is written once; a researcher's term (`productKey` nil) gets no entry | A fourth switch (`placeFarEndLabel`) — the set is closed, but four unguarded switches is how this bug happened; a map in `ConclusionVocabulary.swift` — it holds keys, not copy |
| D-5 | `placeTermWordings` is internal (not private) so `CitationBridgeSentence.displayLabel` in `CitationComposerAuthoringBox.swift` reads it; `displayLabel` keeps its default for every other term | Both files are the composer feature; one table | A copy of the two cases in the box |
| D-6 | `listedLine`'s `placeRelationship` case: when `viewing == to.id` and `placeWording(term)` is non-nil, label = `L10n.string(wording.farEnd)`; otherwise label = `term.map { PropertyTermDisplay.name(term: $0, propertyKey: SeededPropertyKey.placeRelationshipType) } ?? L10n.string(L10n.CitationComposer.bridgeKindPlace)`. Value stays `other.label` | `from` is the part / earlier place (§4 Direction); the near end now localizes like the far end, so one row never mixes languages (review M-2) | Raw `term.label` on the near end — catalog English beside L10n copy |
| D-7 | No entry for a researcher's term: both ends read its own label | No known inverse; matches `endTitles`' default | Guessing an inverse from the key |
| D-8 | The Go guard is a new subtest of `TestSwiftVocabularyMatchesGo`, `every seeded place term has composer wording`, with a path var `swiftCitationComposerAuthoring` *(new)* = `"../../macos/App/Features/CitationComposer/CitationComposerAuthoring.swift"` and a regexp `placeWordingEntry` *(new)* = `` `SeededPlaceRelationshipTypeTerm\.(\w+): CitationPlaceTermWording\(` ``. It fails for each `swiftNamedKey` name in `SeededPlaceRelationshipTypeTerm` with no entry | Same shape as "every seeded term has wording"; runs off a Mac; already reads the constants Go pins to the seed | A Swift test over a hand-kept `all` list — a new constant can miss the list as easily as the table |
| D-9 | Swift tests call `CitationComposerAuthoring.groups` with a hand-built vocabulary. `termVocabulary` gains `endLabels: (a: String, b: String) = ("a", "b")`; a private `savedRow(termProperty:termID:bridgeTypeKey:)` *(new)* builds the saved `ConnectionRow` (review M-4). The researcher-term case extends `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` (review M-3) | `groups` is pure; one test per behavior | A model test through FakeStore |

**Rules this plan bends.** Working rule "(Batch 7) Place relationship rows use the inverse label on the far end" named "preceded by" for `succeeded_by` while also saying "matching the Place conclusion page" ("Succeeded"). User answer 2026-11-05: "Succeeded". The planning commit rewords the rule, bar outcome 27 and the S14-32 scope entry; STEP-6 checks them.

## 6. Design

### Codebase fit

Searched: `inverse`, `SeededPlaceRelationshipTypeTerm`, `productKey`, `PropertyTermDisplay`, `Contains`, `Succeeded`, `atWord`, `siteOfWord`, `viewing` in `macos/App`; `seededTermEnums`, `wordingCase` in `api/ffi`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CitationPlaceTermWording` *(new)*, `CitationComposerAuthoring.placeTermWordings` *(new)*, `.placeWording(_:)` *(new)* | type / table / func | the three switches in §4; `SeededRelationshipTypeTerm.inverse` | Replaces the switches (D-4); `inverse` returns a key to look up in vocabulary, and place terms have no inverse row |
| `L10n.CitationComposer.containsWord`, `.succeededWord` *(new)* | key | `L10n.Conclusions.placeContains`, `.placeSucceeded` | D-3 |
| `swiftCitationComposerAuthoring`, `placeWordingEntry` *(new)*, subtest `every seeded place term has composer wording` *(new)* | test | subtest "every seeded term has wording" | Different table; same shape |
| `savedRow(termProperty:termID:bridgeTypeKey:)` *(new)*, `termVocabulary(… endLabels:)` | test helper | five `ConnectionRow(...)` literals in `CitationComposerModelTests`; `termVocabulary` | Extends `termVocabulary`; `savedRow` serves the new tests (D-9) |
| `placeRelationshipRowsReadFromEachEnd` *(new)* | test | `savedBridgeIsListedUnderBothEnds` | That one drives the model for a participation bridge |
| `placeTermWordingKeepsTheBoxWording` *(new)* | test | `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` | That one covers a researcher's term; this pins AC-2.2 for both product terms |

**Builds on.** `listedLine`'s per-end switch (`4745bcb7`); `PropertyTermDisplay`; `SeededPlaceRelationshipTypeTerm`; `TestSwiftVocabularyMatchesGo`'s Swift-table checks.
**Deletes.** The `switch term?.productKey` place cases in `endTitles`, `readingSentence` and `CitationBridgeSentence.displayLabel` (replaced by `placeWording`); the raw `term?.label` read in `listedLine`'s place case.
**Next adds.** The term set is closed (`docs/ideas/properties-and-terms.md`). If it ever opens, a product place term is one constant (Go guard: "every seeded term is in Swift"), one `PropertyTerm` resource (guard: "every seeded term has wording"), and one table entry with its keys (guard: this PR's subtest).

### Reach

None — no new path into existing code, no new state reachable.

### Scale

None — no write, list or loop changes; a dictionary lookup replaces a switch.

### Edges

- **Assumptions.** `from` is the part / earlier place: the Go seed descriptions (§4). Held for both ends of both terms by T-1, and for the box by T-3.
- **Neighbors.** Near end of each product term (T-1); a researcher's term keyed like a product term — no entry, label on both ends, box titles unchanged (T-2); no term (T-1); person relationship and participation rows unchanged (existing tests); `displayLabel` for non-place terms keeps `localizedLowercase` (T-3 checks a role term).
- **Removed inputs.** None — a label choice.
- **Real data.** `place_relationship_type` seeds exactly `part_of` and `succeeded_by`; both have entries; the Go subtest runs over the live Swift constants.
- **Untrusted size.** None.

### Schema

None — no catalog change.

### Go core

None — Swift only; the one Go change is a test (STEP-3).

### FFI

None — no wire change. `api/ffi/swift_vocabulary_test.go` gains the subtest in D-8.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Place-term wording table | `CitationPlaceTermWording`, `placeTermWordings`, `placeWording(_:)` | `CitationComposerAuthoring.swift`, beside `inverseTerm` | Entries: `partOf` → `fromEnd: partEnd`, `toEnd: wholeEnd`, `sentenceTerm: placeTermPartOf`, `farEnd: containsWord`, `reading: L10n.CitationComposer.isPartOf`; `succeededBy` → `earlierPlace`, `laterPlace`, `placeTermSucceededBy`, `succeededWord`, `wasSucceededBy` |
| End titles | `endTitles` place case | same file | `if let w = placeWording(term)` → `(L10n.string(w.fromEnd), L10n.string(w.toEnd), true)`; else today's `default` branch unchanged |
| Reading | `readingSentence` place case | same file | `placeWording(term).map { $0.reading(from.label, to.label) }` |
| Saved line | `listedLine` place case | same file | D-6 |
| Term combo label | `CitationBridgeSentence.displayLabel` | `CitationComposerAuthoringBox.swift` | `CitationComposerAuthoring.placeWording(term).map { L10n.string($0.sentenceTerm) } ?? term.label.localizedLowercase` |

States: unchanged. Accessibility: `CitationSavedLine.body` (VoiceOver) is built from the same label.

### Strings

L10n keys *(new)*, in `L10n.CitationComposer` after `siteOfWord`; catalog entries right after `citationComposer.siteOfWord`, same shape (`comment`, `extractionState: manual`, `en` `stringUnit` `state: translated`):

- `containsWord` → `citationComposer.containsWord` = "Contains"; comment "Place-relationship line label on the whole place, read from its side of part of. Matches conclusions.place.contains".
- `succeededWord` → `citationComposer.succeededWord` = "Succeeded"; comment "Place-relationship line label on the later place, read from its side of succeeded by. Matches conclusions.place.succeeded".

Removed: none. Every existing place-term key stays in use through the table.

## 7. Steps

**STEP-1 — Add the two line-label keys**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: add `containsWord` then `succeededWord` immediately after `siteOfWord`, and the two catalog entries immediately after `citationComposer.siteOfWord`, as §6 Strings says.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — One wording table, four readers**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`.
Do: add `CitationPlaceTermWording` (file scope, internal) and `placeTermWordings` / `placeWording(_:)` on `CitationComposerAuthoring` (D-4, D-5), with the two entries in §6 macOS, each entry starting `SeededPlaceRelationshipTypeTerm.<name>: CitationPlaceTermWording(` on one line. Rewrite the place cases of `endTitles`, `readingSentence`, `listedLine` (D-6) and `CitationBridgeSentence.displayLabel` to read it, as §6 macOS says. No other case changes.
Check: the app target builds with no new warning (`scripts/check-pr.sh` on macOS); `rg -n "SeededPlaceRelationshipTypeTerm\.(partOf|succeededBy)" macos/App/Features/CitationComposer` prints only the two table lines.

**STEP-3 — Go guard**
Files: `api/ffi/swift_vocabulary_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: add the path var, the regexp and the subtest in D-8 after "every seeded term has wording". Read the file with `os.ReadFile`; collect `placeWordingEntry` names; for each `swiftNamedKey` match in `swiftEnumBody(t, swiftVocabulary, "SeededPlaceRelationshipTypeTerm")`, `t.Errorf` when the name has no entry; `t.Fatal` when no entry is read.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run 'TestSwiftVocabularyMatchesGo'` passes; deleting the `succeededBy` entry locally makes it fail naming `succeededBy` (revert after).

**STEP-4 — Swift tests**
Files: `macos/ProvenenciaTests/CitationComposerModelTests.swift`. Skill: `add-swift-test`.
Do: in the "term keys honour the term's origin" MARK section, give `termVocabulary` the `endLabels` parameter (subjects keep ids `"a"` / `"b"`, labels from `endLabels`) and add `savedRow(termProperty:termID:bridgeTypeKey:)`, which returns `ConnectionRow(id: UUID(), isPending: false, fromSubjectID: "a", toSubjectID: "b", fromLabel: "", toLabel: "", bridgeTypeKey:, bridgeID: "bridge-1", sentence: "", termProperty:, roleTermID: termID, termTouched: false, isSaving: false)` (declaration order). Term ids from the helper are `term-<origin>-<key>`; the property is `vocabulary.propertiesByID["p-place_relationship_type"]`. Add T-1 and T-3; extend `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` for T-2. Find groups by `subjectID`, never by index. Compare with `L10n.string(…)` of the keys named in §8, not English literals.
Check: `CitationComposerModelTests` passes in `xcodebuild test` (`scripts/check-pr.sh`). On a host without Xcode (`check-pr.sh` skips `xcode-test` off macOS), report the Swift build and T-1 – T-3 as **not run** and name them in the PR description to be run on a Mac before merge.

**STEP-5 — Version bump**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`. Rule: `.cursor/rules/versioning.mdc`.
Do: raise PATCH by one from what `main` holds when you branch (0.3.3 → 0.3.4 at the planning base). If `main` moves past your number before merge (S14-31 is also a patch), rebase and take the next PATCH; all three places always agree.
Check: `rg -n "0\.3\.3" VERSION core/version.go macos/Provenencia.xcodeproj/project.pbxproj` prints nothing (with the old number substituted if `main` moved).

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-14/deployment-plan.md` (map row status, "Where the notes went" row), `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/ux/2026-10-15-place-relationship-inverse-label.md` (delete), this plan.
Do: per the Definition of done. Check the working rule, bar outcome 27 and the S14-32 scope entry say "Succeeded".
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-32-place-relationship-inverse-labels.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Read a place relationship from the far place's side in the composer's saved lines.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `placeRelationshipRowsReadFromEachEnd` *(new)* | (a) product `part_of` "Part of", ends Calgary → Alberta: `a` label `L10n.string(L10n.PropertyTerm.placeRelationshipPartOf)`, value "Alberta"; `b` label `L10n.string(L10n.CitationComposer.containsWord)`, value "Calgary". (b) product `succeeded_by`, Old Town → New Town: `a` label `PropertyTerm.placeRelationshipSucceededBy`, value "New Town"; `b` label `succeededWord`, value "Old Town". (c) `roleTermID: ""`: both labels `bridgeKindPlace` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.6, D-1, D-2, D-6 |
| T-2 | Swift | `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` (extended) | user-origin `part_of` "Inside of": `groups` lines for `a` and `b` both labelled "Inside of" | AC-1.5, D-7 |
| T-3 | Swift | `placeTermWordingKeepsTheBoxWording` *(new)* | via `readyBridge`: product `part_of` → `endATitle` `partEnd`, `endBTitle` `wholeEnd`, `reading` `L10n.CitationComposer.isPartOf("a", "b")`; product `succeeded_by` → `earlierPlace` / `laterPlace`, `wasSucceededBy("a", "b")`; `CitationBridgeSentence.displayLabel` → `placeTermPartOf` / `placeTermSucceededBy`, and a role term "Witness" → "witness" | AC-2.2, D-4, D-5 |
| T-4 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo/every_seeded_place_term_has_composer_wording` *(new)* | live Swift constants → every one has a table entry | AC-2.1, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` (`savedBridgeIsListedUnderBothEnds`, `aProductTermsInverseIsNeverAResearchersTerm`, `aResearchersSpouseTermGetsNoSpouseSentence`), `TestSwiftVocabularyMatchesGo`, `scripts/check-localizable-xcstrings.py`.

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
- [ ] deployment-plan.md: map row status; "Where the notes went" row says the note was deleted on land; completed.md: one short paragraph of what shipped
- [ ] Dogfood note ux/2026-10-15-place-relationship-inverse-label.md deleted
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source and an Evidence graph holding two places, Calgary and Alberta.
2. In the Citation composer, connect Calgary → Alberta with Part of. The box shows Part / Whole and "Calgary is part of Alberta" as before. Save. Calgary's group reads "Part of Alberta"; Alberta's group reads "Contains Calgary".
3. Add Old Town and New Town; connect Old Town → New Town with Succeeded by and save. Old Town reads "Succeeded by New Town"; New Town reads "Succeeded Old Town".
4. Open Alberta's and New Town's Place pages after Promote: their "Contains" and "Succeeded" rows use the same words.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving three switches changes a string the box shows today | T-3 pins each one against its existing key |
| `CitationPlaceTermWording` holds a closure, so it is not `Equatable` | Nothing compares it; tests compare the strings it produces |

**Stop and report if:**

- `listedLine`'s `placeRelationship` case no longer returns `term?.label` on both ends (someone fixed it first).
- The Go seed descriptions for `part_of` / `succeeded_by` no longer say "From is part of to" / "From was succeeded by to".
- `place_relationship_type` gains a product term beyond `part_of` and `succeeded_by` at your base.
- `CitationComposerAuthoring.groups` or `resolvePair` is no longer callable from tests.
- Swift 6 strict concurrency refuses a `@Sendable` closure in a `static let` dictionary (then report the compiler error; do not switch to a `switch`).

## 12. Open questions

None.

## 13. PR

**Title:** S14-32 — Place relationship rows read from the far end
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-32-place-relationship-inverse-labels.md`.
