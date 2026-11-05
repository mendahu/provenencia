# Ship log — S14-32 Place relationship rows read from the far end

**Plan:** [`S14-32-place-relationship-inverse-labels.md`](S14-32-place-relationship-inverse-labels.md) · **Branch:** `s14-32-place-relationship-inverse-labels` · **PR:** #720
**State:** fixing 1

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
