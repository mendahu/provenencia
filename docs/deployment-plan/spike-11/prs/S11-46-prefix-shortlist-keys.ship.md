# Ship log — S11-46 Prefix keys in the shortlist

**Plan:** [`S11-46-prefix-shortlist-keys.md`](S11-46-prefix-shortlist-keys.md) · **Branch:** `s11-46-prefix-shortlist-keys` · **PR:** #602
**State:** ready

## Round 0 — implement
Head `3d2db4fe`. Gate PASS (orchestrator run). Report deviations: `main` had moved to `b1bf4c54` (Current state rows rechecked, still true); D-8's comparison-layers sentence reads "index keys and lookup keys", so STEP-5's grep misses that file (wording kept); T-1's no-prefix-key check runs on the text rows and two initial rows. STEP-4 timings: warm proposal 86–90 ms on vs 87–90 off (budget 350 ms); cold proposal ~485 vs ~410 ms (budget 2.5 s).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (matching layers `core/valuecompare`, `core/database/namealiases`; shortlist code). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | Prefix keys have no per-value word bound | Fix now → F-1 | verified: `prefixKeys` (`core/valuecompare/shortlist.go:86`) loops every word; `name` Compare returns `CauseTooLong` past `Names.MaxWords` (`name.go:150`), so prefix keys of such a name can never score. Not D-6's "cap" (that is a posting or hit cap that drops "Jo" lookups); this is D-4's hostile-value bound, reusing an existing knob, losing no recall |
| CON-1 (+ MNT outside-lens) | CON | major | Plan has no Landed note | Fix now → F-2 | verified: plan diff is the Status line only |
| PRF-1 | PRF | major | Short given name ("Jo") shortlists every name it starts; the fixture cannot see it | To user (after F-5 measures it) → user 2026-10-28: keep no cap (D-6 stands) | verified the gap: `bench_test.go` given names are all `Giv…`, so no probe asks `p:` for a short common stem. A bound contradicts D-6 and the Out line, a user decision; F-5 takes a throwaway measurement (D-6's own method) so the user decides from numbers |
| MNT-1 (+ SEC outside-lens) | MNT | minor | `sortedUnique` copied; `keyList` inlines the same two lines | Fix now → F-3 (valuecompare half); cross-package half Rejected | in `valuecompare`, one spelling. Across packages: a two-line standard-library idiom with no home short of an exported grab-bag helper; `core/shortlist` keeps its own |
| MNT-2 (+ CON-3) | MNT | minor | Prefix-key check written three ways; rows picked by name string | Fix now → F-4 | file this PR adds to |
| MNT-3 | MNT | minor | `name.wordKeys` doc comment stale | Fix now → F-3 | comment unchanged while body changed |
| PRF-2 | PRF | minor | Rebuild after a handle commit files ~46% more keys | Follow-up | growth class predates this PR; `docs/ideas/name-index-on-commit.md` holds the fix; note this PR's +85 ms there if the user agrees |

Nits (logged): `CON-2 · core/shortlist/shortlist.go:62 · rewrap doc comment` (rides on F-3) · `MNT-4 · same line as CON-2` · `MNT-5 · core/valuecompare/words.go:73 · AliasRoles() slice per word` (3-element literal, left)

### Fix plan 1

**F-1 — Bound prefix keys by the name's word count** (from SEC-1)
Files: `core/valuecompare/shortlist.go`, `core/valuecompare/name.go`, `core/valuecompare/shortlist_test.go`. Skill: `add-match-rule`.
Failing test first: `TestWordKeysReachPrefixes` new check — a given-only name of 33 words (each `Desiree`-length, distinct, e.g. `Desiree`, `Desireea`, …) has no `p:`/`g:` key in either list, while it still has its `w:` keys.
Do: add one small helper in `core/valuecompare/name.go`, e.g. `func tooManyWords(words []nameWord, generation []string, r Rules) bool { return len(words)+len(generation) > r.Names.MaxWords }`, and call it from both `name` Compare's guard (`name.go:150`, for each side) and `name.wordKeys`, which then skips `prefixKeys` (index and lookup stay today's one list) when it is true. Keep `=`, `w:` and `a:` keys unchanged (AC-1.9). `wordKeys` must keep the generation slice `nameWords` returns to pass it.
Check: the test passes; `go test -tags fts5 ./core/valuecompare/... ./core/shortlist/...`; scripts/check-pr.sh PASS.

**F-2 — Landed note on the plan** (from CON-1)
Files: `docs/deployment-plan/spike-11/prs/S11-46-prefix-shortlist-keys.md`.
Do: under the Status line add a **Landed** note listing every difference from the plan: `main` moved to `b1bf4c54` before the build; D-8's comparison-layers wording vs STEP-5's grep; T-1's no-prefix check scope; F-1's word-count bound (a D-4 bound past `Names.MaxWords`, reusing the knob, added from review SEC-1); the MNT-1 decision (`sortedUnique` stays per package); STEP-4's three-run timings on and off (the table from round 0) and F-5's numbers.
Check: `python3 scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.

**F-3 — One sort-and-dedupe spelling in valuecompare; doc comments** (from MNT-1, MNT-3, CON-2)
Files: `core/valuecompare/shortlist.go`, `core/shortlist/shortlist.go`.
Do: `keyList` ends `return sortedUnique(out)`; give `sortedUnique` a one-line doc comment. Extend `name.wordKeys`' doc: "...plus the alias sets that hold a word, and the given-like words' prefix keys (prefixKeys), the only keys index and lookup do not share." Rewrap the `Keys` doc paragraph in `core/shortlist/shortlist.go` to the file's comment width. No behaviour change.
Check: `rg -n "sort.Strings" core/valuecompare/shortlist.go` prints only the line inside `sortedUnique`; tests pass; scripts/check-pr.sh PASS.

**F-4 — One prefix-key predicate in the test file; rows flagged by field** (from MNT-2, CON-3)
Files: `core/valuecompare/shortlist_test.go`.
Do: add `func prefixKey(k string) bool` (true for `p:`/`g:` via the package's `keyPrefix`/`keyGiven` constants if the test is in-package, else the literals once); use it in `plain`, `TestWordKeys` and `TestWordKeysReachPrefixes` (drop the `hasPrefix` closure). Add a `noPrefix bool` field to `TestWordKeys` rows, set on "initial is no key" and "lone initial keys only its whole key", replacing the `tt.name ==` match.
Check: `rg -n 'tt.name ==|HasPrefix\(k, "p:"\)' core/valuecompare/shortlist_test.go` prints at most the one line inside `prefixKey`; tests pass.

**F-5 — Measure the short-name fan-out (throwaway, not committed)** (for PRF-1)
Files: none committed.
Do: on a scratch copy of `core/database/promotealign/bench_test.go`'s large fixture (or `BenchmarkProposeColdCity`'s 20k scale, whichever is quicker to adapt), give persons given names from a realistic pool that shares short stems (John, Joseph, Joan, Josephine, Jo, Albert, Alfred, Al, Ann, Anna, Anne, Edward, Edmund, Ed, Elizabeth, Eliza, …) and put short-name probes ("Jo", "Al", "Ed", "Ann") in the Source. Time a cold proposal, a warm proposal and a proposal after a handle commit, three runs each, with `PrefixMinLength` 2 and 0. If you can count shortlisted handles per short-name Subject cheaply, report that too. Restore every scratch edit; `git status` clean of it.
Check: report the table in the fix report. No commit.

## Round 1 — fix
Head `ed247b59`. Gate PASS (orchestrator run). F-1 to F-4 done; F-5 measured, nothing committed. Implementer note: one commit attempt was denied by the permission classifier and retried as a plain `git commit` (same content).

F-5 (20k persons, 40-name given pool with short stems, probes Jo / Al / Ed / Ann, 3 runs each):

| Proposal | PrefixMinLength 2 | PrefixMinLength 0 | Budget |
| --- | --- | --- | --- |
| Cold | 1.92 / 1.95 / 1.93 s | 1.83 / 1.89 / 1.84 s | 2.5 s |
| Warm | 231 / 241 / 245 ms | 229 / 253 / 233 ms | 350 ms |
| After handle commit | 1.93 / 1.96 / 1.93 s | 1.79 / 1.83 / 1.83 s | 1.5 s (the 20k scale is over it with prefixes off too) |

MNT-1, MNT-2, MNT-3 (minors, no round 2 for MNT): **Resolved** — read `3d2db4fe..ed247b59`: `keyList` returns `sortedUnique(...)`, `wordKeys` doc names prefix keys, `Keys` doc rewrapped, `prefixKey` is the one predicate (`rg` prints only line 61), no `tt.name ==`.

## Round 2 — review
Reviewers: CON (Sonnet; CON-1 major), SEC+PRF (SEC-1 major). Not sent: MNT (minors only, resolved above).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Landed note | Resolved | Landed paragraph lists every difference; `check-plan-docs` 0 errors |
| SEC-1 | SEC | major | Word bound | Resolved | `tooManyWords` shared by Compare's guard and `wordKeys`; `TestWordKeysReachPrefixes/too_many_words` |

No new findings. Nits (logged): `CON-4 · plan:6 · no blank line before **Spike / slice:**` (orchestrator added the blank line in the ship-log commit).

## Result
Gate PASS · CI green · open: none · follow-ups: PRF-2 (noted in `docs/ideas/name-index-on-commit.md`) · to user: PRF-1 → keep no cap · CON miss: none
