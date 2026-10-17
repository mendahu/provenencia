# Review — S11-02 Value compare core and match registry

**Verdict:** Fits with changes — Ready with changes (round 1)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- The four `autoreconcile` word helpers, and `match`'s tokenizer and folder, move in S11-02 instead of S11-03 (D-8). The S11-02 and S11-03 scope entries change with it.
- Text compares words in order. A shared word out of order makes the pair `Different` with cause `order`, which is what the confirmed worked case `Upper Canada, York` / `York, Upper Canada` asks for (D-7).
- Three new cause codes (`no_match`, `order`, `unsupported`) and one unit code (`count`).
- Section types live in the layer and numbers live in `matchregistry`, which imports the layers (D-1).
- `core/match` reads its seven shared knobs from the registry (D-10).
- Typed keys (`SubjectType`, `PropertyKey`) move to S11-10 (D-14).
- There is no allow-list in the numbers guard (D-15).

## Round 1

### M-1 · major · The plan moves the word helpers but copies the tokenizer and folder
`splitWords`, `rawTokens`, `spacedParticle`, `periodInitial` and `foldWord` would be copied into `valuecompare` while `match` keeps its own until S11-05; the reviewer asked to move them.
**Answer:** Fixed. D-8 moves them. `SplitWords` and `FoldWord` are exported, and `match/names.go`, `frequency.go` and `words.go` call them. The Codebase fit row, Deletes, §4, STEP-2's files and its check all changed with it.

### M-2 · minor · `Detail.Unit` is the only untyped code
**Answer:** Fixed. D-5 adds `type Unit string` with `UnitCount`, and the skill's codes section names units.

### M-3 · minor · `Supports` is public API with only a test caller
**Answer:** Fixed. `Supports` is gone, and `export_test.go` exposes `ComparerTypes` to the tests (D-3, AC-1.11, AC-2.1, T-3, STEP-5).

### M-4 · minor · Integer writes the shared fall-off curve inline
**Answer:** Fixed. D-12 adds an unexported `falloff`, and the skill tells types with a distance to use it or say why not.

### M-5 · minor · The reconciler's dependency arrow points at the hub package
**Answer:** Fixed. D-8 states the rule that no layer package imports `core/autoreconcile`, and the skill's "do not" list repeats it. STEP-10 adds the open choice to the S11-06 scope entry.

### M-6 · minor · The pure packages' skills do not mention the guard
**Answer:** Fixed. STEP-8 adds the guard line to `add-name-value` and `add-date-value`.

### M-7 · nit · `Default()` is called seven times at init
**Answer:** Fixed. D-10 uses one `var value = matchregistry.Default().Value`.

### X-1 · blocker · The text alignment has no tie-break
**Answer:** Fixed. D-7 step (4) names the dynamic programming, the most-pairs rule and the walk-back order. T-7's swap invariant and the existing stop condition still guard it.

### X-2 · blocker · The similarity formula contradicts `Different` / `order` = 0
**Answer:** Fixed. D-7's Similarity sentence opens with "`Different` and `Unknown` are 0".

### X-3 · major · The `order` condition and the per-word relation and cause are ambiguous
**Answer:** Fixed. D-7 says every unpaired A × unpaired B combination is checked, adds the word table (exact, initial, fold, spelling), and says when `Parts` is filled.

### X-4 · major · Term normalization unspecified
**Answer:** Fixed. D-12 compares the key as given and byte-equal; D-4 says when text and terms are present.

### X-5 · major · The guard walk needs unstated choices
**Answer:** Fixed. D-15 covers the repo root, `ImportDir` with a prefix trim (never `Import`), a visited set, literal parsing (`0.0` and `1.0` pass, `IMAG` fails), and what `IgnoredGoFiles` covers.

### X-6 · major · The Xcode check cannot run on Linux
**Answer:** Declined. `.agents/skills/ship-pr/SKILL.md` L50 already requires the user's Mac and says to stop on another machine. STEP-9 now says so in one clause.

### X-7 · major · STEP-10's `rg` check will not pass as written
**Answer:** Fixed. STEP-10 also updates `daitch-mokotoff.md` L5 and `comparison-layers.md` L410, and the check names the allowed hits. §4's docs row lists all three files.

### X-8 · minor · `IgnoredGoFiles` covers more than tags
**Answer:** Fixed in D-15 (see X-5).

### X-9 · minor · Smaller choices
**Answer:** Fixed where a choice was open. `matchregistry` is imported only in `names.go` (D-8). The `LetterFolds` snapshot is noted in D-10. `Parts` contents are stated in D-7. The T-5 arithmetic is right as written: the reviewer confirmed six words. The generator wording stands. D-18 needs no commit for STEP-9.

## Round 2

### M-8 · minor · The skill still lists `Supports` after M-3 removed it
**Answer:** Fixed. D-17's "adding a comparer" now says to add the new type to T-3's expected `ComparerTypes()` list.

Logged, not raised: the `Key` row could also name `autoreconcile` `textKey` / `termKey` (S11-06 moves them); two word-similarity functions exist until S11-05, as "Build, then migrate" allows.
