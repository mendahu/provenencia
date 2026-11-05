# Ship log — S14-32 Place relationship rows read from the far end

**Plan:** [`S14-32-place-relationship-inverse-labels.md`](S14-32-place-relationship-inverse-labels.md) · **Branch:** `s14-32-place-relationship-inverse-labels` · **PR:** #720
**State:** ready

## Round 0 — implement
Base `c8a8bdb1`. Head `5067b8c6`. Gate PASS (orchestrator run, all nine checks, xcode-test 49s). Report deviations: `wordingCase` takes an optional leading dot, because D-6 writes `return .placePartOf` (in the Landed note); code and tests are one behavior commit, because the registry rows, keys and composer changes break the build apart; manual check not run. The report said the ship log was never staged, but `48166fc1` committed it: no harm, the log belongs on this branch.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`, `Localizable.xcstrings`), SEC+PRF (`api/ffi/**`, `macos/App/Features/**` model). Skipped: none. CON and SEC+PRF: clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Place `Sentence` rows have no guard | Fix now → F-1 | verified: the subtest reads only `termWording(propertyKey`; `rg ConnectionSentenceRegistry api` is empty; a missing row silently drops the draft reading and falls back to the card's generic "is" sentence, the drift this PR removes. §6 Next adds names rows as part of the next add. Extends D-9, contradicts nothing |
| MNT-2 | MNT | major | `add-seeded-vocabulary` omits the new steps | Fix now → F-2 | verified: `.agents/skills/add-seeded-vocabulary/SKILL.md:79` lists two steps and names neither `termWording` nor the registry |
| MNT-5, UX-1 | MNT, UX | minor | Three Out lines homed in the dogfood log were never logged | Fix now → F-3 | verified: nothing under `docs/dogfood` or `spike-14/ux`. The plan's Out table promises those homes and the plan is now Landed. Copy stays the user's call, so only the notes land |
| UX outside lens, MNT-8 | CON, MNT | minor | Catalog entries are not where §6 Strings puts them | Fix now → F-4 | verified: `containsWord` and `succeededWord` at lines 4 and 16, `siteOfWord` at 21566; `placePartOf` at 4099, `burialSubject` at 4291. Minor in a file this PR changes |
| MNT-3 | MNT | minor | `readingSentence` looks the row up twice | Rejected | contradicted by D-5, which spells out this shape. No behavior differs and the registry has eight rows. Named in the report so the user can overturn it |
| MNT-4 | MNT | minor | `succeededWord` duplicates `L10n.Conclusions.placeSucceeded` | Rejected | D-3 rejects reusing `L10n.Conclusions` keys by name, and its comments are the cross-reference it chose. Named in the report so the user can overturn it |
| MNT-7 | MNT | nit | `wordingCase` comment shows only the undotted form | Fix now → F-1 | stale comment in a file this PR changes |

Nits (logged): `CON-1 · plan DoD boxes still [ ]` (no action: S14-30's landed plan keeps 15 unticked) · `MNT-6 · CitationComposerAuthoring.swift:735 · make termWording(propertyKey:termKey:) private` · `MNT-9 · CitationComposerModelTests.swift:1609 · placeLines could take an origin parameter`

### Fix plan 1

**F-1 — Require a registry sentence for every composer-worded term** (from MNT-1, MNT-7)
Files: `api/ffi/swift_vocabulary_test.go`.
Failing test first: remove the `SeededPlaceRelationshipTypeTerm.succeededBy` row from `ConnectionSentenceRegistry.sentences`, run `go test -tags fts5 ./api/ffi/ -run 'TestSwiftVocabularyMatchesGo'`, and see it pass (the gap). Add the check, see it fail naming that term, then restore the row.
Do: add `swiftConnectionSentenceRegistry = "../../macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift"` to the path `const (…)` block. Add a `sentenceRow` regexp beside `wordingCase` that reads `term: SeededPlaceRelationshipTypeTerm.partOf` from a `Sentence(` row, with enum and name groups. Add a sibling subtest `every composer-worded term has a connection sentence` after `every composer-worded term has composer wording`, over the same `composerWordedTermEnums` (the one list a new property joins); `t.Fatal` when no rows are read, as its sibling does. Update the `wordingCase` comment to show the `return .placePartOf` form as well.
Check: both subtests pass with the row back; scripts/check-pr.sh PASS.

**F-2 — Say what a composer-worded term needs in the vocabulary skill** (from MNT-2)
Files: `.agents/skills/add-seeded-vocabulary/SKILL.md`.
Do: in the paragraph "A seeded term the Mac app names has a Swift step too", add one sentence: a term of a property listed in `composerWordedTermEnums` (`TestSwiftVocabularyMatchesGo`, today `place_relationship_type`) also needs a `CitationComposerAuthoring.termWording` case with a `CitationBridgeTermWording` value and a `ConnectionSentenceRegistry.sentences` row, and that test fails, naming the term, until both exist; a new composer-worded property adds a row to that list. Touch nothing else in the skill.
Check: `scripts/check-guidance-names.py` passes; scripts/check-pr.sh PASS.

**F-3 — Log the plan's three dogfood Out lines** (from MNT-5, UX-1)
Files: three new notes under `docs/dogfood/ux/`, one per finding as `docs/dogfood/ux/README.md` asks, dated 2026-11-05, in its template with a `## Code notes` section where it helps. Take the words from the plan's Out table.
1. `2026-11-05-place-card-fallback-phrase.md`: with a missing endpoint and no stored label, a place bridge reads "Is part of of · <ref>" or "Is succeeded by of · <ref>" (`bridgeSummaryPlaceRelationshipTypeOnly`, `wholeNameFallback`). It is ungrammatical and now says "is" where the sentence says "was". A copy fix the user has not decided.
2. `2026-11-05-relationship-reading-beside-registry.md`: the `relationship` case of `readingSentence` (`areSpouses`, `areSiblings`, `isATermOf`) sits beside the registry's relationship sentences, the same drift this PR closed for places.
3. `2026-11-05-raw-term-label-bypass.md`: the `relationship` and `participation` cases of `listedLine`, and `endTitles`' relationship case, read raw `term.label` rather than `PropertyTermDisplay`.
Do not change any copy or Swift for these.
Check: the three files exist; scripts/check-pr.sh PASS.

**F-4 — Put the new catalog entries where the plan says** (from the UX outside-lens note, MNT-8)
Files: `macos/App/Resources/Localizable.xcstrings`.
Do: move `citationComposer.containsWord` and `citationComposer.succeededWord` to immediately after `citationComposer.siteOfWord`, and `evidenceGraph.bridge.sentence.placePartOf` and `.placeSucceededBy` to immediately after `evidenceGraph.bridge.sentence.burialSubject`. A move only: entries unchanged, no other key reordered, so `git diff` on the file shows only those four blocks moving against the round-0 head.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `git diff --stat 5067b8c6 -- macos/App/Resources/Localizable.xcstrings` shows equal insertions and deletions; scripts/check-pr.sh PASS.

Also add F-1's guard to the plan's Landed note as a difference from D-9. Change no Decision, Scope line or AC.

## Round 1 — fix
Head `2212ebf6`. Gate PASS (orchestrator run, all nine checks). F-1 to F-4 done. F-1's "see it pass first" step was vacuous, so the gap was confirmed by the new check failing with the `succeededBy` row removed, naming the term; row restored. `sentenceRow` matches any `Sentence(kind:` row, not only place rows. Orchestrator read F-3 and F-4: Resolved (three notes in the README's template; the four catalog entries follow `siteOfWord` and `burialSubject`, 48 insertions, 48 deletions).

## Round 2 — review
Reviewers: MNT only (the only lens with a Fix-now major). Scope `5067b8c6...2212ebf6`. MNT-1 Resolved. MNT-2 Resolved. No new blocker.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-10 | MNT | below the round-2 bar | F-1 made a third copy of the enum check loop | Fix now → F-5 | verified: `swift_vocabulary_test.go` lines 651, 673, 694 repeat the loop, and lines 646, 667, 688 repeat the `named` collection; this PR added two of the three. Same-pass cleanup, whatever its severity; test-only, so the same-pass rule wins (S12-18). The "Fix plan 2 only for blockers" cap is stretched for this one test helper, with no third review round |

### Fix plan 2

**F-5 — Fold the term-coverage loops into helpers** (from MNT-10)
Files: `api/ffi/swift_vocabulary_test.go`.
Failing test first: none, a refactor. Prove the helper still fails: with the `succeededBy` row removed from `ConnectionSentenceRegistry.sentences`, the subtest fails naming that term, then restore the row.
Do: add two file-level helpers beside `swiftFuncBody`, one that collects `enum.name` from a regexp's first two groups and one that takes the term enum list, the `named` set and the "has no …" text and reports each `swiftNamedKey` name missing. Move the subtests `every seeded term has wording`, `every composer-worded term has composer wording` and `every composer-worded term has a connection sentence` onto them. Keep each error message's words and each subtest name unchanged, and keep the two `t.Fatal` guards. Touch no other subtest.
Check: `grep -n 'swiftNamedKey.FindAllStringSubmatch' api/ffi/swift_vocabulary_test.go` prints one line in the helper, plus any line outside those three subtests that was there before; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `d3f0e885`. Gate PASS (orchestrator run, all nine checks). F-5 done: `namedTerms` and `reportUnnamedTerms` in `api/ffi/swift_vocabulary_test.go`; the three subtests use them, with names, message words and both `t.Fatal` guards unchanged. Proof: with the `succeededBy` row's term renamed, the sentence subtest failed naming the term; restored. Orchestrator read the commit: Resolved. No third review round.

`main` still at `c8a8bdb1` by the GitHub API, because `git fetch` failed on DNS: no merge needed.

## CI — round 1
Go tests red on `9ac1ebbb`: `TestLargeSourceBudgets/warm` 350.57 ms against the 350 ms budget, in `core/database/promotealign`, which this PR does not change. macOS tests and localizable green. The same budget flaked in S14-16 (373 ms) and S14-20 (360 ms). The next push reruns it.

## User direction after ready (2026-11-05)
`docs/dogfood/` is only for feedback from people using the app, so F-3's three notes are removed and their words stay in Fix plan 1 above. `ship-pr`, `plan-pr` (skill and template) and `archive-docs` no longer offer it as a home, the two dogfood READMEs say so, and follow-ups go to the user for a home. The plan's three Out homes are decided by the user.

## Scope added by the user (2026-11-05)
Asked where the follow-ups go, the user put all of them on this PR and set the warm budget to 375 ms. This supersedes the plan's three "dogfood log" Out lines and its "What stays" line for person relationship wording. The S14-29 vertex items are held for the user's design answers (Fix plan 4).

### Fix plan 3

**F-6 — Raise the warm budget to 375 ms** (user; CI round 1)
Files: `core/database/promotealign/perf_test.go`.
Do: `budgetWarm = 375 * time.Millisecond`. In the comment above the const block add: raised from 350 ms by the user on 2026-11-05, when the CI runner measured 350.57 ms (S14-32).
Check: `go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets` passes.

**F-7 — Place cards with a missing endpoint read the term, not "Is … of"** (from UX-1)
Files: `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift`.
Failing test first: `placeTermsHaveTheirOwnPhrase`: `ConnectionSentenceRegistry.sentence(kind: .placeRelationship, endpointA: nil, endpointB: nil, term: "Part of", termKey: SeededPlaceRelationshipTypeTerm.partOf)` is "Part of"; `succeededBy` gives "Succeeded by"; a researcher term "capital", with no term key, still gives "Is capital of".
Do: give the two place `Sentence` rows a `phrase` that returns the term's own `L10n.PropertyTerm` wording through `L10n.PropertyTerm.resource(propertyKey:termKey:)`. The card fallback then reads "Part of · <ref>" and "Succeeded by · <ref>". Keep `bridgeSummaryPlaceRelationshipTypeOnly` for researcher terms. Add no new key.
Check: the test passes.

**F-8 — The relationship reading reads the registry** (from the plan's Out line, same shape as D-5)
Files: `ConnectionSentenceRegistry.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `ConnectionSentenceRegistryTests.swift`, `macos/ProvenenciaTests/CitationComposerModelTests.swift`. Skill: `add-localized-string`.
Failing test first: `spouseAndSiblingHaveTheirOwnSentence` in the registry tests: with term keys, `sentence(kind: .relationship, endpointA: "Ann", endpointB: "Bo", term: "Spouse", termKey: SeededRelationshipTypeTerm.spouse)` is "Ann and Bo are spouses", and `sibling` is "Ann and Bo are siblings". `relationshipReadingMatchesTheCard` in the model tests: the draft reading equals the registry sentence for spouse, for parent, and for a researcher's term.
Do: add `L10n.EvidenceGraph.bridgeSentenceSpouse(a:b:)` (`evidenceGraph.bridge.sentence.spouse`, "%1$@ and %2$@ are spouses") and `bridgeSentenceSibling(a:b:)` (`evidenceGraph.bridge.sentence.sibling`, "%1$@ and %2$@ are siblings"), each after `bridgeSentencePlaceSucceededBy` in `L10n.swift` and the catalog. Add two `.relationship` `Sentence` rows with `phrase` nil. `readingSentence`'s relationship case: `guard let term else { return nil }`, then `ConnectionSentenceRegistry.sentence(kind: .relationship, endpointA: from.label, endpointB: to.label, term: <the F-9 display name>, termKey: term.productKey)`. Delete `L10n.CitationComposer.areSpouses`, `areSiblings`, `isATermOf` and their catalog keys. Do not add `SeededRelationshipTypeTerm` to `composerWordedTermEnums`: nine of its terms use the kind's sentence.
Researcher-visible: the draft reading moves to the card's generic wording for other terms, and the spouse and sibling cards move to the draft's wording.
Check: `rg -n 'areSpouses|areSiblings|isATermOf' macos` prints nothing; tests pass.

**F-9 — Composer rows read a term's display name** (from the plan's Out line)
Files: `CitationComposerAuthoring.swift`, `CitationComposerModelTests.swift`.
Failing test first: `productTermsReadTheirLocalizedNameOnEveryRow`: a product term with catalog label "Catalog words", for a relationship and a participation, reads the `L10n.PropertyTerm` name on the `listedLine` rows and the `endTitles` titles; a researcher's term keeps its label.
Do: in `listedLine`'s relationship and participation cases and `endTitles`' relationship case, replace raw `.label` with `PropertyTermDisplay.name(term:propertyKey:)`, using `SeededPropertyKey.relationshipType` and `SeededPropertyKey.role`, including the inverse term. Add one small private helper only if the three sites would otherwise repeat it.
Check: `rg -n '\$0\.label|term\?\.label|term\.label|inverse\?\.label' macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` prints no term reader in those two functions; scripts/check-pr.sh PASS.

**Docs for Fix plan 3.** In the plan: change the three Out lines' homes to "this PR (user, 2026-11-05)", and extend the Landed note with F-6 to F-9, naming the card and draft copy changes. Add the same copy changes to the S14-32 paragraph in `spike-14/completed.md`. Change no AC or Decision.

## Result
Gate PASS · CI pending · open: homes for the three F-3 follow-ups · to user: MNT-3 and MNT-4, rejected on D-5 and D-3, named so they can be overturned; Fix plan 2 stretched its blockers-only cap for one test-only same-pass cleanup · CON miss: none
