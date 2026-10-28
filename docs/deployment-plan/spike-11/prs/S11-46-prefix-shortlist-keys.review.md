# Review — S11-46 Prefix keys in the shortlist

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet); round 2 one reviewer on both lenses (sonnet): Fits / Ready with changes

## To user
- None.

## Writer's own calls
- D-2: the tag scheme `g:` (given-like whole word) and `p:` (strict prefix), crossed between index and lookup, so surnames never meet by prefix.
- D-3: the alias roles move to `valuecompare.AliasRoles()`; `namealiases.aliasPartRoles` is deleted.
- D-4: words over `MaxSpellingLength` runes key no prefixes (no new knob).
- D-5: `PrefixMinLength` lives in `NameRules`, not `WordRules`, so no `CacheVersion` / `ProjectionVersion` bump.
- D-6: no cap. The planning measurement on the large fixture: +46% index keys, +13% shortlist hits, every budget held.
- Side effect, accepted: one-letter spelling variants that value compare already relates ("Ann" / "Anne") now shortlist too.
- Map: S11-45's "Not this PR. Surname or two-letter prefixes." became "Surname prefixes.", because two-letter prefixes count (user, 2026-10-24).

## Round 1

### M-1 · blocker · Prefix-word rule decided twice (here and in S11-45's comparer)
**Answer:** Fixed. D-4 now names one predicate, `prefixWord(w nameWord, r Rules) bool`, with `prefixesOf` beside it in `words.go`. `prefixKeys` reads only them. Next adds says S11-45's comparer calls `prefixWord` and widens T-2 to `CausePrefix` pairs.

### M-2 · minor · T-3 builds given-only values inline
**Answer:** Fixed. `slGiven` and `slNick` are added beside `slName` (T-3, Codebase fit).

### M-3 · minor · Five key tags as bare literals with cross-side pairing
**Answer:** Fixed. D-2 adds a const block (`keyWhole`, `keyWord`, `keyAlias`, `keyGiven`, `keyPrefix`), each commented with which side files it and which side asks for it. STEP-2 names the existing three; STEP-3 adds the new two.

### X-1 · blocker · `check-pr.sh` cannot print PASS off macOS
**Answer:** Fixed. The note above STEP-1 says the PASS line comes from a macOS run, and a Linux run reports "every other check PASS, Xcode skipped". The DoD line says "on macOS".

### X-2 · major · AC-1.10 contradicts which tests are edited
**Answer:** Fixed. AC-1.10 now lists T-1, T-2, T-4 and STEP-1's renames.

### X-3 · major · Commit plan conflicts with STEP-4's check
**Answer:** Fixed. STEP-3 is committed on its own before STEP-4. STEP-5 is its own docs commit with its own subject.

### X-4 · major · Word shorter than `PrefixMinLength`
**Answer:** Fixed. `prefixWord` requires at least `PrefixMinLength` runes. A word of exactly that length files `g:` and looks up `p:` with no prefixes of its own (D-4).

### X-5 · minor · Nick-part row has no recipe
**Answer:** Fixed. `slNick` gives the value literal (T-3).

### X-6 · minor · T-1's new row duplicates an existing one
**Answer:** Fixed. The assertion moves onto the existing "initial is no key" row. A new row covers a lone `slGiven("J.")`.

### X-7 · minor · Skill edits not word for word
**Answer:** Fixed. D-8 gives both replacement sentences.

### X-8 · minor · STEP-5's check cannot see the wrapped comment
**Answer:** Fixed. STEP-2 rewrites `shortlist.Keys`' doc comment. STEP-5's check drops that phrase and checks that the new wording is present.

### X-9 · minor · No step opens the PR
**Answer:** Fixed. STEP-6 pushes the branch made from `main` and opens the PR; STEP-7 is the follow-up docs commit.

### X-10 · minor · `-tags fts5` inconsistent; D-8 numbers unverifiable
**Answer:** Fixed. The note above STEP-1 makes `CGO_ENABLED=1 -tags fts5` the rule for every command, and the STEP-2 and STEP-3 commands carry it. D-8 says its numbers are D-6's, copied, not re-measured.

### X-11 · nit · Stop condition trips on a comment
**Answer:** Fixed. The pattern is now `WordKeys\(` and the condition says "a call".

### X-12 · nit · Reused AC labels in `TestLookup`
**Answer:** Fixed. The new rows are labelled `S11-46 AC-1.n`.

## Round 2

### X-13 · blocker · T-1's `slGiven("J.")` row expected `nil`, but a lone initial files its whole key
**Answer:** Fixed. The row now wants exactly `whole(slGiven("J."))` on both lists, with no `p:` or `g:` key. `keyList` adds the `=` key for any name with a word, as the existing "initial is no key" row already shows.

Also noted, not changed: `prefixWord` builds `AliasRoles()` once per word. That is a three-element slice, which is negligible at index build; a package-level list is the fix if it ever shows in a profile.
