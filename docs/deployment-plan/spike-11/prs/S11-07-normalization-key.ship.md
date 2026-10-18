# Ship log — S11-07 One normalization key across the core

**Plan:** [`S11-07-normalization-key.md`](S11-07-normalization-key.md) · **Branch:** `s11-07-normalization-key` · **PR:** #460
**State:** fixing 1

## Round 0 — implement
Base `e97cf54d`. Head `789ec019`. Gate PASS (orchestrator run, cold test cache: go-test 128s, xcode-test 15s). Report deviations: `addProductNicknames` counts tokens `insertSet` wrote; the label check reuses `distinctAlternates`; two extra `TestPropertyTerms` rows; alias test scores through word similarity; a follow-up commit fixed gofmt and the Swift `key` reference. The implementer committed this ship log in `0b284fa6` (asked to leave it untracked; harmless).

## Round 1 — review
Lenses: conformance Clean (2 minor), maintainability Changes needed (2 major, 1 minor), UX Clean (1 minor, 1 nit), security Clean, performance Changes needed (1 blocker).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Rescoring keys every field of every candidate per keystroke (~20x per candidate, under the session lock) | Fix now → F-1 | Reviewer microbenchmark (2 KB notes: 8 µs → 193 µs per candidate); `mergeFTS` has no LIMIT. Verified by reading `text.key`: `SplitWords` + per-word `NormalizeForm` ×2 + `norm.NFD` per word. The key joins with `""`, so splitting and particle rejoin cannot change it; a single pass with an ASCII fast path is the same function. Stored keys (schema) held back as the fallback, to the user only if F-1's numbers do not hold |
| MNT-2 | MNT | major | `Install` duplicates `insertSet` | Fix now → F-2 | `namealiases.go` L64–71 vs `add.go` `insertSet`; identical after this PR |
| MNT-1 | MNT | major | `valueWords` declared in four packages | Rejected | D-5 chose it; `matchregistry.Default()` returns fresh copies on purpose (skill `add-match-rule`: "builds fresh slices and maps on each call"), so a shared cached value is the thing it avoids; the plan's round-2 review logged "a sixth Go caller → a `matchregistry` accessor" |
| CON-1 | CON | minor | Install test's alias score passes on the seed alone | Fix now → F-3 (rides with F-2, same package) | `nicknames_seed.go:596` holds `james`/`jim` |
| (UX note) | CON | minor | `Update` with a keyless label (`!!!`) → `ErrInvalid` (D-7) has no test | Fix now → F-3 | `rg '"!!!"' core/database/propertyterms/propertyterms_test.go` finds nothing |
| UX-1, CON-2 | UX, CON | minor | Manual check step 2 names a Properties term editor and rename UI the app lacks | Fix now → F-4 | `createPropertyTerm` has one caller, `CitationComposerModel.swift` `confirmCustomTerm`; no Swift caller of `updatePropertyTerm` |
| MNT-3 | MNT | minor | Label check re-keys via `distinctAlternates` length | Follow-up | Fix plan does not touch `handles.go` |
| UX-2 | UX | nit | Offer "Use “Godfather”" from the duplicate error | Follow-up | Idea for `docs/ideas/term-management.md`, if the user agrees |

### Fix plan 1

**F-1 — Make the text key one pass over the string** (from PRF-1)
Files: `core/valuecompare/text.go`, `core/valuecompare/text_test.go` (or a new `core/valuecompare/text_bench_test.go`), `core/search/search_test.go`. Skill: `add-match-rule`.
Failing test first: none (output must not change). Add `TestTextKeyMatchesWords` *(new)*: for every text value in the contract `samples` and generator draws, plus `Mc Donald`, `O. Brien`, `god-father`, `Mont-réal`, `Sainte-Anne register`, `ß`-containing text, and a 2 KB French paragraph, the new key equals `strings.Join(SplitWords)` (folded per word at `Equivalent`) at both levels, and both report no key for `!!!` and `""`.
Do: add one unexported `joinedKey(s string, fold bool, w WordRules) string` in `text.go`: one pass over the runes of `s`; skip dashes (`unicode.Pd`), `/`, whitespace and punctuation exactly as `NormalizeForm` does; lowercase the rest; at `fold`, an ASCII rune passes through and a non-ASCII rune goes through the same accent strip and `w.Folds` as `FoldWord` (decompose that rune only, drop `unicode.Mn`). `text.key` returns `joinedKey`; `text.compare`'s `Same` / `Equivalent` checks use `joinedKey` on `a.Text` / `b.Text` instead of joining the word lists. No other behavior changes; `TestKeyContract` and `TestComparerContract` must pass unchanged. Add `BenchmarkScoreFieldsSource` *(new, `core/search`)*: one Source candidate with title, ref, a 300-byte description, 2 KB notes, 400 bytes of metadata and a filename, one query token; report its ns/op and allocs/op at base `e97cf54d` (old `ToLower`+`Contains`, by checking out base in a temp worktree) and at the new head, in the report.
Check: `go test ./core/valuecompare/...`; `CGO_ENABLED=1 go test -tags fts5 ./core/search/... ./core/autoreconcile/... ./core/database/promotealign/...` pass; `go test -tags fts5 -run '^$' -bench ScoreFieldsSource -benchmem ./core/search` numbers reported; scripts/check-pr.sh PASS.

**F-2 — `Install` writes members through `insertSet`** (from MNT-2)
Files: `core/database/namealiases/namealiases.go`.
Do: replace `Install`'s inner member loop with `if _, err := insertSet(tx, stored, i, set); err != nil { return err }`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/namealiases/...` passes.

**F-3 — Make two tests prove what they claim** (from CON-1 and the UX outside-lens note)
Files: `core/database/namealiases/namealiases_test.go`, `core/database/propertyterms/propertyterms_test.go`.
Do: in `TestInstallStoresComparisonTokens`, add the set `{"Mary-Ann", "Polly"}` to the registered table and assert given `Maryann` against given `Polly` scores `AliasCredit` through `Rules` (no seed set holds `maryann`). In `TestPropertyTerms`, add rows: create `!!!` → `ErrInvalid`; update an existing user term's label to `!!!` → `ErrInvalid`, label unchanged.
Check: both packages pass under `-tags fts5`; the new alias assertion fails if the registered set is removed.

**F-4 — Manual check goes through the custom-term sheet** (from UX-1, CON-2)
Files: `docs/deployment-plan/spike-11/prs/S11-07-normalization-key.md` (§10 step 2 and the Landed note only; no AC, Decision or Scope line changes).
Do: §10 step 2 becomes: open a Source's citation composer, choose a term Property's value (Role), choose Add custom term, enter `Godfather` and confirm, then Add custom term `god-father`: the sheet shows the D-8 message under the Label field. Renaming has no macOS UI yet; AC-4.5 and AC-4.6 are covered by `TestPropertyTerms`. Add a Landed-note line: AC-4.8's "term editor" is the citation composer's Add custom term sheet; F-1 rewrote `text.key` as one pass; the Risks row's "bounded by FTS candidates" was wrong (`mergeFTS` has no LIMIT).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-07-normalization-key.md` reports no errors.
