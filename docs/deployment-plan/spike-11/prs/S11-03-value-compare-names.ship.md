# Ship log — S11-03 Value compare: names

**Plan:** [`S11-03-value-compare-names.md`](S11-03-value-compare-names.md) · **Branch:** `s11-03-value-compare-names` · **PR:** #447
**State:** ready

Base `79f7f042` (main).

## Round 0 — implement
Head `8a55fe8b`. Gate PASS (orchestrator run; go-test and xcode-test cached from the implementer's full run at `221bc9cd`, docs-only commit since). PR #447 (draft).
Report deviations: D-16 commits (4)–(6) folded into one (`25d970db`; contract test needs name rows); `match` `TestWesternPatternCoversEveryPartType` reduced and renamed `TestDefaultPatternIsBuiltIn`; `match.bestAssignment` removed in favour of `valuecompare.BestAssignment` (per D-13); backticks dropped around moved paths in the plan for check-plan-docs; D-11 step 5 "no pair" read as no word pair (a shared generation word alone is not a pair); §10 item 5 (Promote) not run in the app.

## Round 1 — review
Reviewers at `8a55fe8b`: CON changes needed, MNT changes needed, UX clean, SEC changes needed, PRF changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | blocker | Step-8 cause depends on which side is A; AC-2.1 unmet, contract swap check omits cause and mirrored `Parts` | To user → Fix now → F-4 | verified against D-11: step 4 "cause of step 8 is picked after the mirror" + step 8 "pairs in A's sorted order" is side-dependent by construction; probe `s=Robins\|g=Jon` / `g=John\|_=Robins` → role vs spelling. Plan Decision conflicts with its AC |
| MNT-1, CON-2 | MNT, CON | blocker | Names compute steps 9–10 by hand instead of the weighted `meanSimilarity` | Fix now → F-1 | verified: `rg meanSimilarity core/valuecompare` → text.go only; D-11 step 9 says names call it |
| MNT-2 | MNT | major | `wordParts` `kinds` parameter only ever receives nil | Fix now → F-1 | verified text.go:47; D-11 step 9 binds the fix direction: names use `wordParts` with kinds (not removing the parameter) |
| PRF-1 | PRF | blocker | `match.resolve` rebuilds and re-normalizes the alias index per compare on the live Promote/matching path (~2x per compare) | To user → Fix now → F-5 | verified core/match/words.go:60; D-8 says "built in `resolve`" while its reason says "once per load, not per compare": the fix changes D-8's wiring |
| PRF-2, SEC outside-lens | PRF | major | `AliasIndex.Hit` linearly scans ~1,240 sets per word pair | To user → Fix now → F-5 | verified aliases.go:61; fix adds a token→sets map to the struct D-8 spells out |
| SEC-1 | SEC | major | No word-count bound on names (exact route `rows × 2^cols` with unbounded rows) | To user → Fix now → F-6 | verified: text has `Text.MaxWords` 64 + `CauseTooLong`, names none; fix adds a `NameRules` field D-4 does not list |
| SEC-2 | SEC | major | Stored name-format weights unchecked; Similarity can exceed 1 or be +Inf | Fix now → F-2 | verified nameformats.go:216-238 copies weights as stored; new comparer divides by them |
| CON-3 | CON | minor | Landed note omits the D-11 step-5 reading | Fix now → F-3 | rides along; plan file already edited by this PR |
| MNT-3 | MNT | minor | `name.compare` is one ~200-line function | Fix now → rides with F-1 | file touched by F-1; pull out leftovers and weakest-contributor |
| MNT-4 | MNT | nit | Nested `switch role`; unknown role → untyped deserves a comment | Logged | |
| UX outside-lens | UX | — | `LoadID` no longer sets `NamePattern.Key` | Rejected | D-2: "`nameformats.LoadID` builds `valuecompare.NamePattern{PartRoles: …}` with no `Key`" |

**User calls, 2026-10-17** (amend the plan's Decisions; record each in the Landed note as a user-approved change):
- CON-1 → D-11 step 4/8: the step-8 cause is picked in the **canonical side's** sorted order, **before** the mirror.
- PRF-1/PRF-2 → D-8: the alias index is built **once per load**, not in `match.resolve`; `AliasIndex` gains a token → sets map so `Hit` checks only the sets holding `x`.
- SEC-1 → D-4: `NameRules.MaxWords`, **32** (estimate), names over it on either side → `Unknown`, `too_long`.

### Fix plan 1
Order: blockers (F-1, F-4, F-5), then majors (F-2, F-6), then the minor (F-3). MNT-3 rides with F-1.

**F-1 — Names combine through the weighted `meanSimilarity` and leftovers go through `wordParts`** (from MNT-1, CON-2, MNT-2, MNT-3)
Files: `core/valuecompare/name.go`, `core/valuecompare/text.go`. Skill: `add-match-rule`.
Failing test first: none (refactor; numbers must not change). Every `TestNameCompare`, `TestNameLadders`, `TestNameInvariants`, contract case keeps its expected value.
Do: as `det.Parts` is built, build `pairWeight` beside it (`wa[p.a].weight + wb[p.b].weight` per word pair, `Generation + Generation` per shared generation pair). Step 9 = `meanSimilarity(det.Parts, pairWeight, totalWeight)`; step 10's paired mean = `meanSimilarity(det.Parts, pairWeight, pairedWeight)`. Drop `earnedTotal`/`wordEarned` from the combine; compute each side's total weight once. Build `OnlyA`/`OnlyB` word leftovers through `wordParts(tokens, kinds, idx)` with the words' part kinds (D-11 step 9), not `nameWord.part()`. Pull the leftover collection and the weakest-contributor step (step 8) into small functions in `name.go`.
Check: `go test ./core/valuecompare/...` passes with no expected value changed; `rg -n meanSimilarity core/valuecompare` shows a call in `name.go`; scripts/check-pr.sh PASS.

**F-4 — The step-8 cause does not depend on which side is A** (from CON-1; user call)
Files: `core/valuecompare/name.go`, `core/valuecompare/name_test.go`, `core/valuecompare/contract_test.go`.
Failing test first: add to `nameCases` in `name_test.go`:
`{name: "AC-2.1 cause does not depend on the side", a: nv("s=Robins|g=Jon"), b: nv("g=John|_=Robins"), rel: valuecompare.Near, sim: (0.8*(F+U) + 0.75*2*G1) / (F + G1 + G1 + U), why: valuecompare.CauseRole}` — fails today on swap with `spelling` (write the literal `2` as the file's existing doubling idiom if the numbers guard objects). Then in `TestComparerContract`'s swap block, also assert `swappedDet.Cause == det.Cause` and that `Parts` are mirrored (A/B pairs equal as a set, `OnlyA`↔`OnlyB`). Add an untyped surname to the name generator pools so the generator reaches this mix.
Do: pick the step-8 cause by walking contributors in the canonical (`lo`) side's sorted order before the mirror, then mirror. `Parts` ordering after the mirror is unchanged (A's sorted order).
Check: the new case and the tightened contract pass; scripts/check-pr.sh PASS.

**F-5 — Build the alias index once per load and index it by token** (from PRF-1, PRF-2; user call)
Files: `core/valuecompare/aliases.go` (+ `words_test.go` `TestAliasIndex`), `core/match/words.go`, `core/match/registry.go`, `core/database/namealiases/namealiases.go`, `core/match` tests that set `Words.Aliases`.
Failing test first: `TestAliasIndex` case that `Hit` finds a set for a token that appears in several sets, and misses for tokens in different sets (guards the new map; behaviour unchanged).
Do: `AliasIndex` gains a `byToken map[string][]int` built in `NewAliasIndex`; `Hit` checks only `byToken[x]`'s sets for `y` and the roles. `match.WordRules.Aliases` becomes `*valuecompare.AliasIndex` (zero index for `noAliasTables`); `namealiases.withAliases` builds it once with `valuecompare.NewAliasIndex(tables)`; `match.resolve` copies it, never calls `NewAliasIndex`. Delete `aliasTablesOf` if it has no caller left. No `match` expected value changes.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match/... ./core/database/{namealiases,nameformats,matching,promotealign}/...` pass unchanged; `rg -n NewAliasIndex core/match` prints nothing; scripts/check-pr.sh PASS.

**F-2 — Stored name-format weights are checked where they are read; the comparer never divides by ≤ 0** (from SEC-2)
Files: `core/database/nameformats/nameformats.go`, `nameformats_test.go`, `core/valuecompare/name.go`, `name_test.go`.
Failing test first: `nameformats_test.go` `TestLoadIDRefusesBadWeight` — a profile row with weight `-1` (and one with `+Inf` if SQLite accepts it via a bound float) → `LoadID` returns an error. `name_test.go` case: a pattern in `Rules.Names.Patterns` whose weights make the divisor 0 → `Unknown`.
Do: in `LoadID`, refuse a weight that is NaN, ±Inf, or < 0 with an error (refuse, not repair). In `name.compare`, when `totalWeight <= 0` or `pairedWeight <= 0` return `Unknown` (reuse an existing cause; do not add a new code).
Check: both tests pass; scripts/check-pr.sh PASS.

**F-6 — Names have a word bound** (from SEC-1; user call)
Files: `core/valuecompare/rules.go`, `core/valuecompare/name.go`, `core/matchregistry/value.go`, `docs/matching-model.md` (knob inventory row), `name_test.go`.
Failing test first: `name_test.go` case: a name of 33 given words vs `g=John` → `Unknown`, `CauseTooLong`; 32 words → not `too_long`.
Do: `NameRules.MaxWords int`; `matchregistry/value.go` sets 32 with the comment "estimate (S11-03 user call 2026-10-17): far above any real name; bounds BestAssignment's exact route". In `name.compare`, after the `Same`-key step, either side's word count plus generation tokens over `MaxWords` → `Unknown`, `CauseTooLong`, as `text.compare` does. Add the `value.Names.MaxWords` inventory row so `TestKnobInventoryMatchesRegistry` passes.
Check: test passes; `go test ./core/matchregistry/...` passes; scripts/check-pr.sh PASS.

**F-3 — Landed note lists every deviation** (from CON-3)
Files: `docs/deployment-plan/spike-11/prs/S11-03-value-compare-names.md` § Landed.
Do: add the D-11 step-5 reading (a shared generation token alone is not a pair), and the three user-approved amendments above (D-11 cause order, D-8 build-once + token map, D-4 `Names.MaxWords` 32). If `comparison-layers.md` or `add-match-rule` text names the old D-8 wiring or the knob list, update that line too.
Check: `scripts/check-plan-docs.py` 0 errors.

## Round 1 — fix
Head `cb2a3416`. Gate PASS. F-1 `7238ade9`, F-4 `7b0a5c0a`, F-5 `df30e246`, F-2 `ad93b667`, F-6 `53d21d58`, F-3 `cb2a3416` — all done.
Report notes: F-5 `rg NewAliasIndex core/match` matches only `names_test.go` (tests build an index), no product caller; F-5 test is a guard (passed before the change, behaviour unchanged by design); F-2 reuses `CauseEmpty` for the zero-weight `Unknown`; F-4 mirrored-`Parts` assertion is names-only (text alignment is not mirrored, outside AC-2.1's name scope); no change needed in `comparison-layers.md` or `add-match-rule`.

## Round 2 — review
Lenses: CON, MNT, SEC, PRF (UX had no Fix-now). Scope `8a55fe8b...cb2a3416`.
Verdicts at `cb2a3416`: CON clean (CON-1/2/3 resolved), MNT clean (MNT-1/2/3 resolved), PRF clean (PRF-1/2 resolved), SEC changes needed (SEC-1 resolved; SEC-2 partly).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-3 (SEC-2 cont.) | SEC | blocker (re-rated from major) | A large finite stored weight (`1e308`) passes `LoadID`, overflows `x.weight + y.weight` and the totals to +Inf; Similarity NaN, plain names `Different` | Fix now → F-7 | reviewer probe: `1e308` → `John Robins / Robins` contains NaN; `1e307` correct. Re-rated: unvalidated untrusted catalog value reached by S11-05 (rubric blocker row); defect in code this PR adds |
| CON-4 | CON | nit | Test case labelled `AC-1.17`, which the plan does not have | Logged | rides along with F-7 (same file, a rename) |
| MNT r2 note | MNT | nit | `CauseTooLong` comment still says text only | Logged | rides along with F-7 (one comment line) |
| MNT r2 note | MNT | nit | `sortedBy` generic helper with one caller | Logged | |

### Fix plan 2
**F-7 — Stored name weights have a ceiling; the comparer never divides by a non-finite sum** (from SEC-3)
Files: `core/database/nameformats/nameformats.go`, `nameformats_test.go`, `core/valuecompare/name.go`, `name_test.go`, `core/valuecompare/valuecompare.go` (comment only).
Failing test first: `TestLoadIDRefusesBadWeight` gains a `1e308` row → `LoadID` returns an error (fails today). `name_test.go` case: a pattern in `Rules.Names.Patterns` with all weights `1e308`, `John Robins` / `Robins` → `Unknown`, `CauseEmpty` (today: contains NaN).
Do: in `nameformats`, a named constant `maxWeight = 1000` with a one-line comment ("far above the built-in 1.5; refuses weights whose sums overflow") and `LoadID` refuses `weight > maxWeight` alongside the NaN/Inf/negative check. In `name.compare`, extend the `totalWeight <= 0 || pairedWeight <= 0` guard to also return `Unknown`, `CauseEmpty` when either is NaN or Inf (`math.IsInf`, `math.IsNaN`). Ride-alongs: rename the `"AC-1.17 …"` case to drop the AC prefix (e.g. `"weights that sum to nothing are unknown"`); update the `CauseTooLong` comment to cover names ("text or a name with more words than may be compared").
Check: both tests pass; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `3fd1a822`. Gate PASS; orchestrator also ran the changed Go packages uncached (`-count=1`): valuecompare, matchregistry, match, namealiases, nameformats, matching, promotealign all ok. F-7 done (both new tests red first). CON-4 and the `CauseTooLong` comment rode along.

## Result
Gate PASS · CI see PR #447 · open blockers/majors: none · follow-ups: none · logged nits: MNT-4 (nested `switch role`), `sortedBy` single-caller helper · to user: CON-1, PRF-1/PRF-2, SEC-1 — answered 2026-10-17 (canonical-side cause order; build-once alias index + token map; `Names.MaxWords` 32).
