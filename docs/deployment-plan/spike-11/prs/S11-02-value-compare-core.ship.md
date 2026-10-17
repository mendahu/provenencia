# Ship log — S11-02 Value compare core and match registry

**Plan:** [`S11-02-value-compare-core.md`](S11-02-value-compare-core.md) · **Branch:** `s11-02-value-compare-core` · **PR:** #441
**State:** ready

Worktree: `/Users/jakerobins/Repos/provenencia-s11-02`, off `origin/main` @ `500396f7`.

## Round 0 — implement
Head `b1015940`. Gate PASS (orchestrator run). Report deviations: contract tests use 4 seeds × 500 pairs with integer tolerance 0–3; `TestCompareEchoesRefs` split out; `OnlyA`/`OnlyB` filled for `no_match` and `order`; registry comments paraphrase deleted constant names; `daitch-mokotoff.md` L5 paraphrases the old path; STEP-6 deliberate failures need `-tags fts5`; backticks removed around `core/autoreconcile/spelling.go` in the plan for `check-plan-docs`; `positionalParts` scores byte-equal 1 else `FoldCredit`; 9 concern commits + 3 docs fixes.

## Round 1 — review
Base `500396f7` · head `b1015940`. UX: clean. CON: clean (minors). MNT, SEC, PRF: changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | Text compare has no word-count or word-length bound; quadratic memory and time | To user → Fix now → F-4 | Verified: `alignWords` builds full n×m `sims` + `grid`, `EditDistance` uncapped; reviewer measured 4,000 words → 10.8s / 7.6 GB. Fix needs a bound knob, which contradicts D-7 ("no search bound") and D-11 (no pairing-limit knob). **User (2026-10-17): "Bound it now"** — amends D-7 (no search bound) and D-11 (no limit knob) by the user's call |
| PRF-1 | PRF | major | Text compare folds each word once per cell and recomputes in `crossesOrder`; ~2× slower than `core/match` | Fix now → F-2 | Verified: `wordSimilarity` (words.go:203) calls `FoldWord` on both words per call; `text.go:35` folds then discards; `crossesOrder` re-calls `wordSimilarity`. Results unchanged |
| MNT-1 | MNT | major | `docs/matching.md` and `core/match/registry.go` comments still say to tune folds/particles/numbers in `match` | Fix now → F-3 | Verified: `docs/matching.md:95,98,171,186`; `registry.go:10-11,32,38-39` "Add … here" above registry aliases |
| CON-1 | CON (+SEC outside-lens) | minor → major | Integer gap in float64: distinct values > 2^53 come back `Near` 1 | Fix now → F-1 | Verified `integer.go:13`. Re-rated: breaks AC-2.1 (`Same` ⇔ 1) in code this PR adds; "real counts are small" is not a reason to rate down. Failing test supplied |
| CON-2 | CON | minor | `TestTextCompare` counts `Parts` but never checks their content or order | Fix now → F-2 | Rides with F-2: the `text.go` refactor needs the `Parts` content pinned first |
| PRF-2 | PRF (+MNT outside-lens) | minor | `particleSet` map built on every `Compare`/`Key` | Fix now → F-2 | `text.go:11-20,75`; F-2 already touches `text.go` |
| MNT-3 | MNT | minor | `meanSimilarity` adds each pair twice with no reason given | Fix now → F-2 | `text.go` touched by F-2; two-line change |
| MNT-2 | MNT | minor | `comparison-layers.md:185` still says the word helpers live in `core/autoreconcile` | Fix now → F-3 | Verified; file already edited by this PR |
| CON-4 | CON | minor | S11-03 map row still says the word helpers move in from `autoreconcile` | Fix now → F-3 | Verified `deployment-plan.md:92`; file already edited by this PR |
| CON-3 | CON | minor | No test exercises the inventory guard's AC-3.4 / AC-3.5 failure branches | Follow-up | `inventory_test.go` not otherwise touched; code read correct by CON |
| MNT-4 | MNT | nit | `wordPair.why` should be `cause` | Logged | — |
| MNT-5 | MNT | nit | Contract-test label prints nothing for integer pairs | Logged | — |

### Fix plan 1

**User decision recorded:** SEC-1 → bound it now. This amends D-7 and D-11 only as far as F-4 says; every other part of D-7's algorithm stands. The plan file is not edited. The Landed note gets one line naming the amendment and pointing here.

**F-1 — Integer gap in integer arithmetic** (from CON-1, SEC outside-lens)
Files: `core/valuecompare/integer.go`, `core/valuecompare/integer_test.go`, `core/valuecompare/contract_test.go`.
Failing test first: `TestIntegerCompare` new rows (or `TestIntegerCompareLargeGap`): `1<<53+1` vs `1<<53` at tolerance 0 → `Different`, `beyond_tolerance`; `math.MaxInt64` vs `math.MinInt64` → `Different` with no overflow panic.
Do: order the two values, take the difference as a `uint64` (`uint64(hi) - uint64(lo)` is exact for any int64 pair), compare that against `uint64(Tolerance)` (Tolerance ≥ 0). Convert to float64 only for `falloff` and `Detail.Distance`. Add samples near `math.MaxInt64` / `math.MinInt64` to the contract test's integer sample table. No numeric literal in `valuecompare` (the layer guard): use `math` constants only in tests.
Check: tests pass; `TestLayerPackages` passes; scripts/check-pr.sh PASS.

**F-2 — Fold once per side, reuse the alignment's similarities, pin `Parts`** (from PRF-1, PRF-2, CON-2, MNT-3)
Files: `core/valuecompare/text.go`, `core/valuecompare/words.go`, `core/valuecompare/text_test.go`, `core/matchregistry/value.go` only if the particle set is kept on the rules.
Failing test first (CON-2, pins behaviour before the refactor; it should pass on the current code, then stay green): in `TestTextCompare`, replace `parts int` with an expected `[]valuecompare.PartMatch` (A text, B text, Relation, Similarity, Cause) for at least the `contains`, `order`, `spelling two words`, `fold` and `same` rows, in A's order.
Do:
- In `text.compare`, compute `fa, fb := foldWords(wa, …), foldWords(wb, …)` once and use them for the Equivalent check and pass them into `alignWords`.
- Split `wordSimilarity` so the folded tokens are passed in (e.g. `wordSimilarity(x, y, fx, fy string, w WordRules)`); `FoldWord` is not called per cell.
- `alignWords` returns (or exposes) its `sims` matrix; `crossesOrder` reads it instead of calling `wordSimilarity` again. Keep only what the DP needs per cell, not a full `wordPair` per cell, if that is a straightforward change.
- PRF-2: stop building a map per call. Either use `slices.Contains` on `w.Particles` inside `SplitWords`' particle check, or keep the set built once. Pick the smaller change that keeps `SplitWords`' exported signature working for `core/match` and `core/autoreconcile` callers.
- MNT-3: in `meanSimilarity`, write the double count as one statement with the reason (`total += p.Similarity + p.Similarity // one word on each side`) or say it in the doc comment.
- Results must not change: `TestTextCompare`, `TestTextLadder`, `TestComparerContract`, `TestKeyContract` pass with no expected value edited (except the CON-2 `Parts` expansion).
Check: tests pass with no tags; `go test -tags fts5 ./core/match/... ./core/autoreconcile/...` unchanged; scripts/check-pr.sh PASS.

**F-3 — Point the docs and comments at `core/matchregistry`** (from MNT-1, MNT-2, CON-4)
Files: `docs/matching.md`, `core/match/registry.go`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/deployment-plan/spike-11/deployment-plan.md`.
Do:
- `docs/matching.md` L95, L98: `LetterFolds` / `SpacedParticles` → `value.Words.Folds` / `value.Words.Particles` in `core/matchregistry/value.go` (what `core/match` reads them from). L171 and L186: the value-section numbers live in `core/matchregistry`; the rest stay in `core/match/registry.go` until their layer lands. L94's `productAliasTables` stays as is.
- `core/match/registry.go`: header comment no longer says "every number … in one file. Tune here."; the "Add a word here" / "Add a letter here" comments say the values are read from `matchregistry.Default().Value.Words` and are added in `core/matchregistry/value.go`.
- `comparison-layers.md` L185: add "(moved to `core/valuecompare` in S11-02)", matching L410.
- `deployment-plan.md` L92 (S11-03 row): drop "word helpers move in from `autoreconcile`;" → "Name comparer on relations; initials and aliases in word similarity; alias seeds to `namealiases`".
Check: `rg -n 'match/registry.go' docs/matching.md` shows only L94 and lines that correctly describe what is still there; check-plan-docs passes; scripts/check-pr.sh PASS.

**F-4 — Bound the text comparer** (from SEC-1; user decision)
Files: `core/matchregistry/value.go`, `core/matchregistry/registry_test.go` if it lists fields, `core/valuecompare/valuecompare.go` (or wherever `Cause` constants live), `core/valuecompare/rules.go`, `core/valuecompare/text.go`, `core/valuecompare/words.go`, `core/valuecompare/text_test.go`, `core/valuecompare/contract_test.go`, `docs/matching-model.md` (knob inventory + cause list), `.agents/skills/add-match-rule/SKILL.md` only if it lists the value knobs.
Failing test first: `TestTextCompare` rows: (a) two texts each longer than `MaxWords` with no shared words → `Unknown`, cause `too_long`, similarity 0, and the test runs fast (build the inputs from `MaxWords+1` generated words); (b) two over-limit texts with the same words → still `Same` / `exact`; (c) over-limit folded-equal → still `Equivalent` / `fold`; (d) one word longer than `MaxSpellingLength` against a one-letter variant of it → no spelling pair (`Different`, `no_match`), while the same pair at `MaxSpellingLength` runes still pairs `Near` / `spelling`.
Do:
- Registry: add `TextRules.MaxWords int` (inventory key `value.Text.MaxWords`, start value **64**, an estimate: toponyms and short remarks stay well under it) and `WordRules.MaxSpellingLength int` (key `value.Words.MaxSpellingLength`, start value **32** runes, an estimate). Add both rows to the knob inventory in `docs/matching-model.md` marked as estimates, so `TestKnobInventoryMatchesRegistry` passes.
- New cause `CauseTooLong` (`too_long`) beside the others; add it to the cause list in `docs/matching-model.md` and to any cause enumeration test.
- `text.compare`: D-7 steps 1–3 (empty, Same, Equivalent) run as today; they are linear. Before step 4, if either side has more than `r.Text.MaxWords` words → `Unknown`, similarity 0, `Detail{Cause: CauseTooLong}` (no Parts, no OnlyA/OnlyB), and `alignWords` is never called.
- `wordSimilarity` (valuecompare only): before the spelling step, if either folded token has more than `w.MaxSpellingLength` runes → 0, no pair. Do **not** put the cap in the exported `SpellingSimilarity`: `core/autoreconcile` calls it and US-4 says the reconciler does not change.
- `Key` is linear and is unchanged.
- The contract tests still hold: `Unknown` ⇒ 0, swap-symmetric, self-compare of an over-limit value is `Same` (step 2 runs first), and `Key` agreement holds.
- No numeric literal in `valuecompare`; numbers live in `matchregistry` only.
- Add one line to the plan's Landed note: "SEC-1 (round 1): user amended D-7/D-11 to add `value.Text.MaxWords` and `value.Words.MaxSpellingLength`; see the ship log."
Check: tests pass with no tags; `TestLayerPackages`, `TestKnobInventoryMatchesRegistry`, `TestDefaultIsFresh` pass; `go test -tags fts5 ./core/autoreconcile/... ./core/match/...` unchanged; scripts/check-pr.sh PASS.

Order: F-4 and F-1 (behavior), then F-2 (refactor, behind the pinned tests), then F-3 (docs).

## Round 1 — fix
Head `64010b70`. Gate PASS (orchestrator run). F-1 `d54a695b`, F-4 `905a0789`, F-2 `a3123271` (Parts pin) + `db9a3a56` (refactor), F-3 `64010b70`. All reproduced red first where a test applied. Deviations: `too_long` also added to `comparison-layers.md`'s cause list; `completed.md` mentions the bounds; bound tests pair N/N−1 and N/N+1 runes. Note: the round-0 implementer committed an early copy of this ship log in `3ed8ac70`; harmless, superseded by the final commit.

## Round 2 — review
Lenses: CON, MNT, SEC, PRF (UX had no Fix-now). Scope `b1015940...64010b70`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, CON-2, CON-4 | CON | — | Round-1 fixes | Resolved | New rows in `TestIntegerCompare`, contract extremes, full `Parts` pins, map row L92 |
| MNT-1, MNT-2, MNT-3, PRF-2 | MNT | — | Round-1 fixes | Resolved | Docs and comments point at `core/matchregistry`; `meanSimilarity` commented; one splitter |
| SEC-1 | SEC | — | Round-1 fix | Resolved | `MaxWords` before `alignWords` (text.go:36); spelling cap before `EditDistance` (words.go:214); registry-only |
| PRF-1, PRF-2 | PRF | — | Round-1 fixes | Resolved | Fold once per side; `crossesOrder` reads `sims`; fixed per-compare cost |
| MNT-6 | MNT | minor | `docs/matching.md:14` still calls `core/match/registry.go` "the one place to configure" | Follow-up | Not a blocker; round 2 adds blockers only |
| CON-5 | CON (outside-lens → SEC/MNT) | minor | `uint64(Tolerance)` wraps on a negative Tolerance; nothing validates `Tolerance ≥ 0` | Follow-up | Only source is `matchregistry.Default()` = 0; rate when the registry becomes editable |

No blockers. No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: CON-3, MNT-6, CON-5 (nits MNT-4, MNT-5 logged) · to user: SEC-1 (answered: bound it now)
