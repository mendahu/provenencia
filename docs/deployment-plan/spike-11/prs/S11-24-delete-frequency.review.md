# Review — S11-24 Delete frequency weighting

**Verdict:** Fits — Ready (round 2 findings fixed)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None.

## User's calls (2026-10-18, during planning)

- Plan S11-24 as normal now that S11-12 is on `main`. No product question came up: the spike's frequency decisions (2026-10-16) and the knob inventory already fix what goes and which numbers replace it.

## Writer's own calls

- Plain weight 2 for every value type, tied to the accept bar in code (D-1). The old starting weights were 2.08 – 2.25.
- Event type keeps the walk's plain 2 rather than subject compare's 0.5 / 20 until S11-13 (D-1).
- Parameters that only carried frequency inputs are dropped now, not left for S11-13 (D-5).
- One new sparse test in `graphalign`, not a `promotealign` fixture (D-8).
- All of D-1 – D-8 were measured on a throwaway prototype: the full Go suite passes, and the golden needs no expectation change.

## Round 1

### X-1 · blocker · The D-1 registry sketch fails AC-3.2
The sketch comment said "cold-start", which AC-3.2's `rg` bans in `core/graphalign`.
**Answer:** Fixed. §6 gives the sketch's comments as the text to write, with no banned word, and says which words AC-3.2 bans.

### X-2 · major · AC-3 `rg` commands use `\|`
Copied out of the table, `\|` is a literal pipe in ripgrep, so the checks prove nothing.
**Answer:** Fixed. The three commands sit in a `sh` block under §2 with a plain `|`; the table rows point at them.

### X-3 · major · `TestSparseNameAndBirthDate` leaves several choices open
**Answer:** Fixed. §6 "Go core" gives the whole test as written, table-driven, with keys, kinds, refs, edge and bridge directions and the expected scores from registry fields. It was compiled and run on the prototype: both subtests pass.

### X-4 · major · Checks that cannot run here; which gate is authoritative
**Answer:** Fixed. The DoD's `check-pr.sh` line says that off macOS every check but `xcode-test` must pass, `xcode-test` reports "needs macOS" (`scripts/check-pr.sh` L96 – L100), and the CI Xcode run on the non-draft PR is the evidence. Manual check 3 is marked for the reviewer on macOS.

### X-5 · minor · No text for the map row or `completed.md`
**Answer:** Fixed. STEP-4 gives the map status, the `completed.md` paragraph, and when N and the date are filled in.

### X-6 · minor · Comment wording left open
**Answer:** Fixed. §6 has a table of each comment to write, word for word.

### X-7 · minor · Temptations not closed
**Answer:** Declined. The reviewer found nothing that needs a decision beyond X-3. `alts`' `except` and `alternative`'s receiver are untouched by D-5 as written, and `pick_internal_test.go`'s `newState` call keeps working because `newState` keeps `stats`.

### X-8 · minor · AC-3.1 includes tests, so it passes only after the test edits
**Answer:** Declined. STEP-1 and STEP-2 are one commit and AC-3.1 is STEP-2's check, after both. The "L437" `Stats{` literal sets only `FanOut` (`align_test.go` L437 – L439), so nothing more needs editing there.

### M-1 · major · Toponym's new weight copies a number `matchregistry` already owns
**Answer:** Fixed. D-2: `DefaultConfig` reads name, sex_at_birth, gender and toponym from `matchregistry.Default().Subject.Weights` through a local `subject(kind, key)` helper. No number moves (2.5 / 3, 0.6 / 1, 3 / 3 are equal in both places). The full Go suite passes on the prototype. D-10 is reworded and §4 gains the `subject.go` row. A Stop condition covers the values drifting.

### M-2 · major · The sparse test hand-builds a date `datevaluetest.YMD` already makes
**Answer:** Fixed. The test uses `datevaluetest.YMD(y, 0, 0)`; the Codebase fit note cites `core/subjectprofile/entity_test.go`.

### M-3 · minor · Comments and docs outside AC-3.2 still call the weights "cold-start" as current
**Answer:** Fixed in part. D-9 rewords `docs/matching-model.md` L166 and L168 and `core/matchregistry/subject.go` L14, L15, L20 and L23 as history ("… that S11-24 removed"). Declined for the `.Floor` mentions in `core/matchregistry/value.go` L44 and L89 and the inventory's Source column: that column records where a starting value came from, so the citations are history and still true. AC-3.2 is not widened to `core/matchregistry`, because the history wording names the frequency weight on purpose.

### M-4 · minor · `ValueTypeScale["subject"]` has no reader
**Answer:** Fixed. D-1 deletes the entry. No default profile has a subject-typed feature, and `pairExhibits` skips `valueType == "subject"` (`exhibits.go` L228). The lookup falls back to the zero scale, as before.

### M-5 · nit · `plain` matching `AcceptScore` is only a comment
**Answer:** Fixed. D-1 writes `accept := 2.0`, `plain := accept`, `AcceptScore: accept`.

## Round 2

Maintainability: Fits, no findings. It checked every round-1 answer against the code. Its one note, below the bar: D-2 maps `gender` onto the sex_at_birth weights by hand, a second copy of `subject.Pools`' pairing until S11-13.
**Answer:** Fixed. §6's comment table adds an end-of-line comment on the `gender` entry that names `matchregistry subject.Pools`.

### X-9 · major · The map row has no Plan link for STEP-4 to keep
**Answer:** Fixed. The map row links the plan and reads `Planned` (wired at planning). STEP-4 says the link is already there.

### X-10 · minor · No text for the test comments; a third stale comment at L265 – L267
**Answer:** Fixed. §6's comment table gives all four test comments word for word, the one at L265 – L267 included. STEP-3 points at it.

### X-11 · minor · The Landed fill-in has no commit
**Answer:** Fixed. STEP-4 writes literal `#N` / `YYYY-MM-DD` placeholders and names the follow-up commit that fills them in once the PR is open.

### X-12 · nit · Header still says Draft
**Answer:** Fixed. Status is `Planned`.

### X-13 · nit · Which assertion style for the exact score
**Answer:** Fixed. STEP-3 uses `math.Abs(row.Score-3) > 1e-9`, the file's style.
