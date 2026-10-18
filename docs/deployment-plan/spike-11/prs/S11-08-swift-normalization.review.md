# Review — S11-08 One normalization helper in Swift

**Verdict:** Fits — Ready with changes (all fixed)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. No product call was open: word breaks never matter (S11-07 D-1, user call 2026-10-17) and the reconciler's grouping policy (S11-06) decide the behavior; the map's S11-08 scope entry (as S11-07 rewrote it) names the callers.

## Writer's own calls
- D-1: `TextMatch` lives in `macos/App/Platform/`, beside `FieldSlug`, not in the design system.
- D-2: keys are built per character (each `Character` lowercased alone), to match Go's per-rune lowercase; the locale parameter defaults to `.autoupdatingCurrent`.
- D-3: an explicit eight-letter fold table in Swift, equal to Go's, instead of relying on Foundation's diacritic fold.
- D-5 / D-7: the combo highlight moves onto `TextMatch.range` (from M-1), so a match through punctuation or word breaks is marked.
- D-6: Go reads the Swift fixtures (`swiftFixturePairs`); `FieldSlug` and `ProjectSlug` are not wired in this PR (logged in `docs/ideas/shared-list-helpers.md`).
- D-10: `NameTitleRegistry` joins the rule (from M-3), so `D.r.` and `Jr.,` are titles.
- D-11: a part-less fake name is not a value (Go's rule), so fixtures that seeded form-only names get parts; header support is now counted in Sources, as the conclusion detail already did.
- D-13: the fake's search folds like Go's rescoring (from M-4).

## Round 1

### M-1 · blocker · The combo-box highlight keeps a second text rule beside `TextMatch`
**Answer:** Fixed — D-5 adds `TextMatch.range(of:in:)` on the per-character rule; D-7 moves `PVComboBoxHighlight` onto it; US-2 (AC-2.1, AC-2.2), T-5, T-8. The Out row is gone.

### M-2 · major · `FakeStore` still clusters names in two loops; the plan only shares the key
**Answer:** Fixed — D-11 is now one `nameField(memberIDs:sources:)` used by both `conclusionDetail` and `personHeaders()`; support is distinct Sources for both; the ranking loop skips part-less outcomes (also X-2).

### M-3 · major · `NameTitleRegistry.lookupKey` is another text rule
**Answer:** Fixed — §4 row, D-10, US-5, T-14; `lookupKey` deleted.

### M-4 · major · `FakeStore` search still lowercases while Go search folds
**Answer:** Fixed — D-13, STEP-8, AC-4.4, T-17; the Out row is gone.

### M-5 · major · The Go drift guard is a second sync mechanism; `FieldSlug` / `ProjectSlug` still sync by hand
**Answer:** Fixed in part — D-6 builds a reusable `swiftFixturePairs(t, path, enum)` helper and rejects backslash escapes loudly. Wiring `FieldSlug` (which has a `\u{2019}` fixture) and `ProjectSlug` is Declined for this PR: the map scopes S11-08 to the text rule, and `ProjectSlug` reaches `core/onboarding`. Logged as an item in `docs/ideas/shared-list-helpers.md` (STEP-9), and the `macos-client.mdc` bullet names the pattern.

### M-6 · major · Docs and rules this PR makes stale
**Answer:** Fixed — STEP-9 adds the `macos-client.mdc` bullet, the `swift_vocabulary_test.go` header comment (D-6), and rewrites the S11-08 scope entry's Ships line. The scope entry is rewritten in the implementing PR, not in this plan PR, because S11-07's branch edits the same line.

### M-7 · minor · The empty-query rule lives in `contains` and again in each caller
**Answer:** Fixed — D-4 replaces `contains(_:key:)` with `TextMatch.Query` (`isEmpty`, `matches(_:...)`, `matchesKey`); D-7, D-9, D-13 use it.

### M-8 · minor · "Next adds" invites a duplicate check in Swift
**Answer:** Fixed — Next adds now says duplicate checks stay in Go (`conflictingTerm`).

### M-9 · minor · `TextMatch.words` looks like `SplitWords` but is not
**Answer:** Fixed — D-4's doc comment says particles do not rejoin; the Out row records it.

### M-10 · nit · The combo box keys each label twice
**Answer:** Fixed — D-7 keys the label once per option and passes it to the band.

### X-1 · blocker · S11-07 is not on main
**Answer:** Fixed — the header says to start only after #460 merges and to branch from that `main`; the first Stop condition checks it. The user asked for the plan on the assumption that S11-07 lands first.

### X-2 · blocker · D-10's `.noEvidence` outcome is overwritten by the ranking loop
**Answer:** Fixed — D-11 (c): the ranking loop skips outcomes whose rank was `nil`; AC-4.2 states `voteSupport` / `voteTotal` 0 for it.

### X-3 · major · T-12 / T-13 expected values left to the implementer
**Answer:** Fixed — AC-4.1 and AC-4.2 give every form, part and assertion (including `displayedValues` with parts, outcome order, vote numbers, `state`, `memberCount`); D-12 makes `parts` a required parameter and names the only callers; T-16 names its setup.

### X-4 · major · STEP-8 docs edits point at text that does not exist
**Answer:** Fixed — STEP-9 gives the exact status sentence, the three appended items, the skill bullet and its position, the rule bullet, and the completed.md heading.

### X-5 · major · Checks that cannot run without Xcode
**Answer:** Declined as a plan change — the implementer runs on macOS per `scripts/check-pr.sh`; the Go guard (STEP-2) runs anywhere.

### X-6 · minor · Unmade choices and a stale line range
**Answer:** Fixed — STEP-1 names the four pbxproj entries and four free IDs with a check; `conclusionDetail` is L1764–L1845; D-14 pins tests to `en_US`; D-2 keys per character, so Greek final sigma matches Go (AC-3.3); AC-1.7 states `isFiltering` for `-`; T-10 gives the full `CatalogProperty`; D-11's doc comment says "simplified".

## Round 2

Maintainability: no new blockers; every round-1 answer checked against the code. **Fits.** One taste point (key each fake search field once, not per token) — Fixed in D-13.

Executability: every Current state row true; the fifteen fixtures give the listed keys under Go's `TextKey`; pbxproj IDs free. **Ready with changes.**

### X-7 · major · Locale unspecified for titles and the fake
**Answer:** Fixed — D-2 adds `TextMatch.posix`; D-10 (titles), D-11 and D-13 (`FakeStore`) pass it, so `III` is a suffix and fake tests pass on any machine locale.

### X-8 · minor · Empty form with parts
**Answer:** Fixed — D-11 keeps the `!o.nameForm.isEmpty` filter; Go refuses an empty form.

### X-9 · minor · More fake-backed fixtures may need parts
**Answer:** Fixed — the Stop condition allows adding parts in any test file with assertions unchanged, named in the Landed note; only an assertion change outside `CatalogQueryRegistryTests` stops.

### X-10 · minor · Scope entry's "Not this PR. Go changes." goes false
**Answer:** Fixed — STEP-9 rewrites that line too.

### X-11 · nit · T-10 setup; T-13 already passes today
**Answer:** Fixed — T-10 names `makeModel` and `warm`; T-13 is marked a regression test.

## Refresh — 2026-10-18, after #460 merged

`main` @ `028a190`. S11-07 landed with one change after the plan was read (`40ee8c2`): the text key's per-rune rules moved to `wordRune` / `appendFolded` in `words.go`, shared with `SplitWords`, and Hangul syllables now decompose at `Equivalent`. The rule the plan mirrors is the same (punctuation and spaces dropped, per-rune lowercase, NFD, `Mn` dropped, folds), and Swift's `decomposedStringWithCanonicalMapping` already decomposes Hangul. All fifteen AC-3.1 fixtures give the same keys under `TextKey` on `main`. No Swift file in §4 changed. Plan edits: header (S11-07 landed, start condition removed, refresh base), the §4 Go text key row, the first Stop condition, and the STEP-9 skill edit (S11-07 added a `TextKey` bullet to `add-match-rule`; the Swift-mirror sentence joins it). No Decision changed.
