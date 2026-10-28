# Review — S11-45 Prefix names

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- **The sheet's sentence is "Matches by a shortened name"** (D-9), one fixed sentence like "Matches by initial". The feature doc's example, "Des may be short for Desiree", would need the two words passed into `lineSentence` for one cause.
- **A prefix pair with the same surname is proposed at the weak band** (D-10, AC-1.9): "Desiree Long" / "Des Long" scores 2.2, as "J. Robins" / "James Robins" scores 2.0 today. A given-only prefix pair scores 1.75 and is offered as an alternative. This follows from the user's 0.7 estimate; no band rule is added.
- **The catalog test and the manual check use Ros / Rosalind and Des / Desmond**, because the product nickname seed has held `{desiree, des}` since S11-44, so Des / Desiree in a new catalog is an alias, not a prefix.
- **`PrefixCredit` lives in `NameRules`**, beside `PrefixMinLength` (D-1), not in `WordRules` beside `InitialCredit` (M-4 option a declined below).

## Round 1

### M-1 · blocker · The prefix role test is stated a second time in the reconciler
D-6 called `slices.Contains(valuecompare.AliasRoles(), role)` in `nameModule.fold`, while `prefixWord` already holds the same test; `AliasRoles` then stood for two ideas.
**Answer:** Fixed. D-4 adds `valuecompare.PrefixRole(role NameRole) bool` (renamed `PrefixRuleRole` in round 2, M-10); `prefixWord` and D-6's fold case both call it, and the `AliasRoles` doc no longer claims the prefix rule. Codebase fit and Deletes list it.

### M-2 · major · `name.pair` guesses the broader side from word length
**Answer:** Fixed. D-3: `wordSimilarity` returns a fourth result, `broader Side`, decided where it makes the initial or prefix match; D-5: `name.pair` copies it. The length rule, its edge case and the `unicode/utf8` import are gone. T-1 checks the side.

### M-3 · major · T-9 copies the unchanged-result loop
**Answer:** Fixed. D-11 adds `sameResult`, the body of `TestReconcileNamesAliases`' unchanged loop moved out; that loop and T-9 call it.

### M-4 · minor · `PrefixCredit` apart from its sibling credits, read through options
**Answer:** Fixed, option (b): `wordOptions` has one field, `prefixCredit`, 0 when either word is not a prefix word. Option (a), moving the credit to `WordRules`, declined: it needs a change to `add-match-rule`'s bump rule (a skill line) for no behavior gain, and `NameRules` is where S11-46 D-5 put the rule's other knob.

### M-5 · minor · T-5 makes `TestWordKeysShareOnSameOrEquivalent`'s name untrue
**Answer:** Fixed. T-5 is a new `TestWordKeysMeetOnPrefixPairs` beside `TestWordKeysReachPrefixes`; the existing test is unchanged, so the skill lines that name it stay true.

### M-6 · minor · Two comments missing from D-12
**Answer:** Fixed. D-4 rewrites `name.word`'s doc and `AliasRoles`' doc.

### M-7 · minor · T-6 near-copies two graphalign tests
**Answer:** Declined. `align_test.go` keeps one named test per dogfood case (`TestOneGivenNameIsOfferedAFullerRecord` for case 1); T-6 is case 2 beside it, and folding the weak-band subtest into `TestIsolatedSubjectScoresOnTheWalkScale` would need a score column that table does not have.

### X-1 · blocker · STEP-7 check and the DoD sweep hit the plan files
**Answer:** Fixed. Both `rg` commands exclude `docs/deployment-plan/spike-11/prs/**`.

### X-2 · major · T-8 says four rows, lists three
**Answer:** Fixed: "three new fold rows".

### X-3 · minor · D-5's `unicode/utf8` parenthetical
**Answer:** Fixed, then made moot by M-2: `name.go` gains no import.

### X-4 · major · D-12 doc edits leave the old text ambiguous
**Answer:** Fixed. D-12 replaces the whole **Evidence wording** and **The reconciler** bullets (the open-question link and the S11-28 sentence go with the old text; the new one links § Decided item 2), and gives the `names.go` file comment's old and new sentence word for word.

### X-5 · major · Checks that cannot run on Linux
**Answer:** Fixed. The note above STEP-1 says: on Linux, open the PR as a draft and report STEP-6's `xcodebuild test`, T-11 and DoD box 3 as "not run: needs macOS"; they run on macOS before the PR leaves draft.

### X-6 · minor · STEP-8's test-file list misses `export_test.go`
**Answer:** Fixed: STEP-8's check names it.

### X-7 · nit · No step creates the branch
**Answer:** Fixed: STEP-1 creates `s11-45-prefix-names` from an up-to-date `main`.

## Round 2

Both reviewers found every round-1 answer held and raised no new blocker.

### X-8 · blocker · STEP-1 branches from `main`, which does not hold the plan
**Answer:** Fixed. STEP-1 says `main` holds this plan at status Planned before the build (it lands on `main` first, as every Spike 11 plan has), and the PR carries no plan commit; a Stop condition covers a `main` without it.

### X-9 · major · STEP-2 cannot compile as written
**Answer:** Fixed. STEP-2 lists `name.go`: `name.word` returns four results there, `compare` and `pair` discard the fourth until STEP-3.

### X-10 · major · A second `WordSimilarity` caller
**Answer:** Fixed. STEP-2 names the alias-index loop's `sim, _, _, _ :=` edit.

### X-11 · major · STEP-7's check does not discriminate
**Answer:** Fixed. The check greps D-12's new phrases, one per file, and that the old "decides only whether an alias table applies" is gone.

### X-12 · major · Two "role decides only the alias table" statements
**Answer:** Fixed (with M-9). D-12 rewrites both Format-agnostic statements, and the sweep covers the old phrase.

### X-13 · minor · T-6 helper and metas
**Answer:** Fixed. T-6 names `rowBySubject` and `nameMetas()` for the second subtest.

### X-14 · minor · Pool words and corpus pairs
**Answer:** Declined. T-5 needs at least one prefix pair, which the plan says; the planning probe ran the widened pools green, and the Stop condition covers a generated failure.

### X-15 · nit · Rewrap layout
**Answer:** Logged.

### M-8 · major · D-12's name-candidates text false after M-2
**Answer:** Fixed. D-12 says the word table names the broader side, and rewrites the "`prefixes` flag" sentence to `prefixCredit`.

### M-9 · major · Statements that the role decides only the alias table
**Answer:** Fixed. D-12 covers the Format-agnostic bullet in `names.go`, the same sentence in `conclusion-reconciliation.md`, and `compareValue`'s doc; STEP-5 edits the code comments.

### M-10 · minor · `PrefixRole` reads as the `prefix` part type's role
**Answer:** Fixed: renamed `PrefixRuleRole`.

### M-11 · minor · `daitch-mokotoff.md` word-table order
**Answer:** Fixed: D-12 adds the prefix rung to the idea's order, and STEP-7 lists the file.
