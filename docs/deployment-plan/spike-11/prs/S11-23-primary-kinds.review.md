# Review — S11-23 One home for primary kinds

**Verdict:** Round 1: Fits with changes (maintainability) and Ready with changes (executability). All findings fixed or declined. Round 2: Fits with changes (maintainability) and Ready with changes (executability), no new blockers, all fixed. **Rounds:** 2. **Reviewers:** maintainability (fresh agent) and executability (fresh agent, dry run). Both read S11-22's plan as landed code.

## To user

None.

## Writer's own calls

- The package is `core/vocabkey`, with untyped string constants (D-1, D-2), following S11-22's convention.
- Origins get one home and so does their check: six `originOK` copies are replaced by `vocabkey.ValidOrigin` (D-4). This widens the map row a little, and the map was updated to match.
- The Property-key set is the keys matching reads by name: the locked keys plus `sex_at_birth` and `gender`, minus `likeness` (D-6, D-8). The map scope entry was updated.
- Swift's `EvidencePrimaryKind` stays where it is and becomes the Swift home (D-10).
- Status was `Blocked — on S11-22` until S11-22 landed; see the refresh at the end.

## Round 1

### X-1 · blocker · The D-7 search cannot come back clean
**Answer:** Fixed. The searches are now named S-1 to S-4, in a code block in §2. After M-5, S-1 must print nothing. Test-data lists are excluded with `-g '!*_test.go'`.

### X-2 · major · `\|` escapes break copied searches
**Answer:** Fixed. The searches are in a fenced block with bare `|`. S-2 includes test files and the lower-case `deleteimpact` names.

### X-3 · major · `fts.go` L432 case list undecided
**Answer:** Fixed by M-4. The registry flag replaces the case list. The `handle(...)` calls stay, because each sets one kind's priority (D-7).

### X-4 · minor · `searchindex.Kind*` ambiguous
**Answer:** Fixed. The plan names `KindPerson` / `KindEvent` / `KindPlace` and says `KindSource` stays.

### X-5 · minor · Open choices in STEP-5 and STEP-7
**Answer:** Fixed. T-5 is a table row using `rawCase` (M-2), and the `vocabkey` import is named. STEP-8 places the skill text by appending to rule 3, adding the last "Do not" bullet, and adding a sentence after the §1.1 origins paragraph.

### X-6 · minor · Counts are off
**Answer:** Fixed. §4 now says about 330 uses in 82 files. The stop condition is above 95.

### X-7 · nit · Context points to nothing
**Answer:** Fixed. The pointer now goes to the map's scope entry.

### M-1 · major · Constants added, but edited lines still spell keys by hand
**Answer:** Fixed. New D-13 and STEP-5: in non-test Go, every spelling of a primary kind or one of the eight Property keys reads `vocabkey`. The exceptions are seed data, code that S11-13 or S11-16 deletes, and literals with another meaning (value types, search document tags). AC-2.6 and search S-4 list the expected leftovers.

### M-2 · major · A subtest where the table takes a row
**Answer:** Fixed. D-10 and T-5 add one `TestSwiftVocabularyMatchesGo` row with `rawCase`.

### M-3 · major · Later map entries would add a second seed check or second names
**Answer:** Fixed in the map in this commit. S11-02 uses the `vocabkey` constants directly as its typed keys, with no re-declared constants. S11-10 checks "registry key is a `vocabkey` constant" and relies on T-4 for the seed.

### M-4 · minor · `isHandleKind` should read the search registry
**Answer:** Fixed. D-9: `KindSpec.Handle` is set by the `handle` helper. `isHandleKind` and the field map read it.

### M-5 · minor · The Swift list in `OmnibarHitPresentation` stays
**Answer:** Fixed. D-7: an `EvidencePrimaryKind(rawValue:)` check comes before the switch, following S10-49's `typeRank`.

### M-6 · nit · `OriginPluginPrefix` has no outside caller
**Answer:** Fixed. It is now unexported `originPluginPrefix`.

### M-7 · nit · "kind" and "type" for one concept
**Answer:** Declined, with a doc line. `IsPrimaryType` takes a catalog row's key and origin, and `IsPrimaryKind` takes a key whose origin is already known. The package doc defines *kind* as the key of a primary Subject type (D-3). Renaming would collide with `PrimaryKinds()`, the list the map calls "primary kinds".

## Round 2

All round-1 answers were checked against the code and hold.

### X-8 · major · STEP-3's S-1 check cannot pass
**Answer:** Fixed. STEP-3's check now also expects `core/search/fts.go` L244, which STEP-5 clears.

### X-9 · minor · S-2 finds two constants, not three
**Answer:** Fixed. AC-2.2 now says "the two exported origin constants".

### X-10 · nit · Unnamed `fts.go` function
**Answer:** Fixed. D-9 names `fieldValuesForKind`.

### M-8 · major · The skill rule is false for SQL text
**Answer:** Fixed. A new Out row says SQL text keeps its literals. D-13 now says "Go expressions", and the STEP-8 skill bullet says "as a Go string literal … (SQL text keeps its literals)".

### Below the bar
D-6 now says "Swift's seven `SeededPropertyKey` keys plus `event_type`". The reviewer noted that S11-22's plan counts six value types, while `ValueTypeDepiction` (S10-32) makes seven. That belongs to S11-22's base check, and I passed it on to the user.

## Refresh after S11-22 landed

S11-22 landed in #434, and its Landed note lists no change that touches this plan. At `ac05ae8`, I reran the §4 rows and searches S-1 to S-4. Only line numbers moved: `properties.go` Origin L38, `originOK` L448, `FakeStore.primaryTypeKeys` L3008, and the `observations.go` value-type lines. The plan is updated, and the status is Planned.
