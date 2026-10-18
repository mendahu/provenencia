# Review — S10-66 Pure depiction value package

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user
- None.

## Writer's own calls
- D-1: the package is `core/depictionvalue`, one file `depictionvalue.go`; tests `validate_test.go` and `readable_test.go` (after round 1).
- D-2: `Readable` and `ReadableImage` move into the pure package (`core/locator` imports only `core/apperr` and stdlib).
- D-3: `FromJoin` stays in `core/database/depictionvalues`, beside the joins it describes.
- D-4: the old symbols are deleted with no type alias, as S11-22 did.
- D-5: `Validate` checks the shape only and does not trim; `InsertTx` trims first, as the date and name stores do (after round 1).
- D-7: the error code stays `depictionvalues.invalid` (FFI and L10n contract).
- D-8: one guard row per new pure package; no rows for `core/match`, `core/graphalign`, `core/autoreconcile`.
- D-14: a pure fixture package `core/depictionvalue/depictionvaluetest`; `observationstest` aliases its consts (after round 1).
- D-17: the `catalog-queries.mdc` domain-error line is corrected; it has been false since S11-22 (after round 1).
- No version bump; no app manual check, only command checks.

## Round 1

### X-1 · blocker · STEP-2's `gofmt -l core api` check fails at the base
At the base it lists `core/database/sourcecredibility/assessments.go` and `core/database/sources/notes.go`, which the PR does not touch.
**Answer:** Fixed — STEP-2's check is now `gofmt -l $(git diff --name-only --diff-filter=ACMR "$(git merge-base HEAD origin/main)" -- '*.go')`, the same file set `scripts/check-pr.sh` `gofmt_changed` uses; §4 gains a "gofmt at the base" row; Out says the two files are not reformatted here.

### X-2 · major · D-13 says "two code commits and one docs commit"
**Answer:** Fixed — D-13 and the §7 Commits line say one code commit, one docs commit, then "Record S10-66 as landed in pull request N."

### X-3 · major · `catalog-queries.mdc` contradicts D-6 and D-10
"Domain errors stay in the table package." is bound by the header and contradicts moving `ErrInvalid`.
**Answer:** Fixed — new D-17 gives the replacement line; STEP-4 edits the file and its check proves the old line is gone; §4 records that the line has been false for dates and names since S11-22, which left it.

### X-4 · major · The plan is Draft and unreviewed
**Answer:** Declined — Draft is correct until review-plan finishes (`plan-pr` step 8); this log is that review, and round 2 runs because X-1 was a blocker.

### X-5 · minor · §4 importer rows are wrong about test files
**Answer:** Fixed — the non-test row says "plus one test, `core/autoreconcile/modules_test.go`"; `depictionvalues_test.go` moves to its own "Table package test" row (internal package, imports nothing from it).

### X-6 · minor · STEP-2 says "only the type names change" in the table test
`ErrInvalid` is used unqualified at L53, L58.
**Answer:** Fixed — STEP-2 and §6 Go core say `Value` and `ErrInvalid` become `depictionvalue.Value` and `depictionvalue.ErrInvalid`, and the file imports `core/depictionvalue` (and `depictionvaluetest`, D-14).

### X-7 · minor · Import list for `depictionvalues` is incomplete
**Answer:** Fixed — §6 Go core: drop `core/locator` and `core/apperr`, add `core/depictionvalue`; `strings` stays for the trim.

### X-8 · minor · Comment sites not fully enumerated
**Answer:** Fixed — §4 lists all five sites; four are rewritten with the new qualifier; `api/ffi/handlers/observations.go` L397 names no package and stays (STEP-2 says so).

### X-9 · minor · T-1 and T-2 leave small choices open
**Answer:** Fixed — T-1 asserts `errors.Is(err, ErrInvalid)` and `apperr.From(err).Code() == apperr.CodeDepictionValuesInvalid`; every T-2 row asserts both `RegionJSON` and `ImageFileID`, and the page-9 case is built with `strings.Replace` as `TestFromJoin` does.

### X-10 · minor · `\|` inside AC commands
**Answer:** Fixed — a note under the US-1 table and under §4 says `\|` is the Markdown escape and the commands run with a plain pipe, as STEP-2, STEP-3 and §10 write them.

### X-11 · minor · Skill and rule specified by outline
**Answer:** Fixed — §6 Skill says the `*(new)*` marks are plan notation, not part of a heading, and the prose stays at the length of `add-date-value`; the Package section no longer carries the mark in its heading.

### M-1 · major · `Validate` trims, unlike its siblings
**Answer:** Fixed — D-5: `Validate` checks `v.RegionJSON == ""` and the id length, with the sibling doc line "Callers trim string fields first"; `InsertTx` stays the one trim site; AC-2.1 and T-1 use `""`; AC-2.1 notes that `"  "` through the store is still refused (`TestDepictionValues` "empty region").

### M-2 · minor · `TestReadable` repeats three `TestFromJoin` rows
**Answer:** Fixed — D-9: `TestReadable` owns the read-rule rows; `TestFromJoin` keeps "no image is no depiction", "off-page region" (S10-61 AC-1.5 cites it) and "does not alias inputs", and drops "both halves" and "not a png"; AC-2.3, T-4, Deletes and the stop condition allow that edit.

### M-3 · minor · A sixth copy of the likeness region JSON
**Answer:** Fixed — D-14 adds `core/depictionvalue/depictionvaluetest` (consts `Locator`, `Region`, no imports) on the `core/namevalue/namevaluetest` precedent, with its own guard row; T-1, T-2 and `TestFromJoin` use it; `observationstest.LikenessLocator` and `LikenessRegion` alias it, so callers are unchanged. The three remaining copies in other packages' tests are an Out line (not planned).

### M-4 · minor · The skill restates rules other skills and docs own
**Answer:** Fixed — D-15: the skill's write, read and gallery sections name the depiction symbols and link `add-ffi-handler`, `add-audit-scope`, and the conclusion layer doc §6 and reconciliation doc; "Do not" drops the restated request rules; `add-ffi-handler` L64 and `add-audit-scope` L44 each gain "See `add-depiction-value`." (STEP-4, with a check).

### M-5 · minor · The pure package doc calls a depiction "a region that shows a face"
**Answer:** Fixed — D-16: both package docs and the `DepictionInput` comment use the `propertymeta.ValueTypeDepiction` wording; "face" stays with `likeness` in the skill's gallery section.

### M-6 · minor · The PNG media type is two independent literals
**Answer:** Declined for this PR, logged — `ingest.StageImage` is Out. A new Out row points at `docs/ideas/depictions-and-likenesses.md`, where STEP-5 adds a "One media type for the stored image" bullet under Still open.

### M-7 · nit · Test file name differs from the siblings
**Answer:** Fixed — D-1 and T-1/T-2 name `validate_test.go` and `readable_test.go`.

## Round 2

Maintainability: every round-1 answer holds (M-1 – M-7 resolved; M-6 decline sound). No new blocker. Verdict: Fits.
Executability: round-1 answers reflected; no false Current state row. Verdict: Ready with changes.

### X-12 · minor · T-4 says three `TestFromJoin` rows remain; D-9 leaves two
**Answer:** Fixed — T-4 says "two rows plus \"does not alias inputs\"".

### X-13 · minor · The STEP-2 gofmt check misses STEP-1's untracked new files
**Answer:** Fixed — the STEP-2 check runs after `git add -A`.

### X-14 · nit · Local `main` in a sandbox can be stale
**Answer:** Logged — the Implementer block is copied verbatim from the template; `implement-pr` fetches `origin/main` before comparing.

### Outside lens (maintainability) · D-10's quoted replacement stops before L40's last sentence
**Answer:** Fixed — D-10 says the "Do **not** add them to `core/database/errors.go` …" sentence stays.

### Outside lens (maintainability) · `graphcache.SourceObservation` stores a depiction as three flat fields
**Answer:** Logged — pre-dates this PR and matches dates and names (`DateID`, `NameID`); this PR only rewrites the L58 comment.

