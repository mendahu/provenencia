# Ship log — S11-15 Shortlist index

**Plan:** [`S11-15-shortlist-index.md`](S11-15-shortlist-index.md) · **Branch:** `s11-15-shortlist-index` · **PR:** #497
**State:** ready

Base: `origin/main` @ `219a3b83` (S11-14 #493 merged; D-9 applies the "S11-14 landed first" order).

## Round 0 — implement
Head `05be55c4` (local, not pushed; STEP-6 uncommitted in the working tree). Gate not run.

**Stopped on:** a plan contradiction. Objective (line 38) accepts the gap "a toponym that differs only by a misspelling" (2026-10-17 decision), but AC-3.3 / T-11 require every existing `promotealign` test to pass with no expected value changed, and §11 stops on any such change.
- Failing: `TestSpellingVariantToponymIsAWeakMatch` (`core/database/promotealign/promotealign_test.go:664`). Place "Provenance" against promoted "Provenanced": no shared word key (`w:provenance` vs `w:provenanced`), so D-7 never seeds the handle → Target `new`, Reason `no_match` (was weak match to PLC-ZFFSH).
- Everything else in `promotealign` and `api/ffi/handlers` passes unchanged, including the person spelling-variant cases (they share a word).

Done: STEP-1 `88ee5137`, STEP-2 `8c9a985e`, STEP-3/4 `e8046e37` `f9bc38f5`, STEP-5 `05be55c4`. STEP-6 written, uncommitted. Not done: T-9, T-10, T-12, T-13, STEP-7, push, draft PR.
Report deviations: STEP-1/2 separate commits; shortlist package + registry + inventory in one commit; `TestWordKeys` builds the generation word from a Suffix part; alias case uses "Bill Robins"/"William Robins".

Taken to the user.

**User call (2026-10-19):** accept the gap; no misspelling key. `TestSpellingVariantToponymIsAWeakMatch` asserts the new behavior (Target `new`, `no_match`), is renamed, and carries a `// dogfood:` mark to revisit the coarse misspelling key if spelling-only names or toponyms go unmatched in the Spike 11 matching dogfood. Written into the plan under §5 after D-15; AC-3.3 and T-11 name the exception; STEP-7 adds a review note to `comparison-layers.md`'s misspellings bullet. Resumed the round-0 implementer.

**Round 0 resumed:** head `69af2fb6`, PR #497 (draft). Report deviations: S11-14-landed shape of D-9 (stats slot renamed to the match index; holds only the shortlist); user's call applied (`TestSpellingOnlyToponymIsNotShortlisted`, `// dogfood:` mark, comparison-layers note); STEP-3/4 share a commit; `promote-graph-alignment.md` edited in S11-14's rewritten bullet instead of D-15's quoted sentence. §10 manual check not run (user's).
T-13 benchmark (`BenchmarkProposeWarmHub`, 20k persons, reported not gated): main 52.1 ms / 18.9 MB → branch 70.5 ms / 98.9 MB per op. Implementer: the bulk Source holds only a place, so main never loaded persons; the index build reads every indexed node. To check in triage.
Gate (orchestrator) on `69af2fb6`: PASS against `219a3b83`.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes). SEC+PRF routed by the matching layers (`core/valuecompare`, `core/subjectcompare`, `core/propertycompare`, `core/matchregistry`, new `core/shortlist`), `core/database/**`, and graph/cache code (`graphcache`).

Orchestrator benchmark check (`BenchmarkProposeWarmHub`, `-benchtime 60x -count 3`, M1; branch's bench file copied onto a `219a3b83` worktree): main 67 / 63 / 60 ms, retained 18.9 MB per catalog; branch 93 / 83 / 79 ms, retained 98.8 MB per catalog. The CPU profile shows `loadShortlist` is the one-time cold build (stays ~2.4 s over 10x and 60x), so the per-op gap is in the warm path itself.

CON verdict: Clean (no findings). Outside its lens: the warm benchmark regression (routed to PRF).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker | Unbounded year span in `appendBuckets` hangs every Promote | Fix now → F-1 | verified: `core/shortlist/shortlist.go:116` loops `floorDiv(lo)..floorDiv(hi)` with no cap; `YearReach` passes any validated range; reviewer timed 1e7-year span at 5 s |
| SEC-2 | SEC | major | Alias-set key holds every token of the set | Fix now → F-2 | verified: `aliases.go` `key: strings.Join(sorted, " ")`, copied per word into `a:` keys; tables come from the catalog with no size bound. Fix keeps D-3's content key (table key + smallest token), so no Decision changes |
| MNT-1 | MNT | major | `promote-graph-alignment.md` still describes the kind scan and stats | Fix now → F-3 | verified: lines 107, 185, 196, 283 at head |
| CON-x (MNT, PRF outside-lens) | CON | major | Vocabulary commit keeps a match index whose keys read term keys and metas | Fix now → F-4 | verified: `graphcache/commit.go` drops labels and vocab on `set.Vocabulary`, not `matchIndex`; D-12's premise is "every input the keys read changes only through a path that drops the slot" — this completes that rule, does not change it |
| PRF-2 | PRF | minor | Index built when nothing needs seeding; T-13 times only that case | Fix now → F-5 | verified: `canon.go` calls `loadShortlist` with empty `todo`; bench Source holds only the fixed Hub. Rides along (F-4/F-5 touch the same package; the Landed note's numbers are wrong without it) |
| MNT-3 | MNT | minor | `TestShortlistSection` re-implements `Rules.Derived` | Fix now → F-6 | one line; F-1 touches the registry section and its test |
| MNT-4 | MNT | minor | Two more test-local BEF/AFT builders | Fix now → F-6 | F-1 adds a case to `core/shortlist/shortlist_test.go`; add `datevaluetest.Before`/`After` and use them in the two new files only |
| PRF-1 | PRF | major | Every handle commit rebuilds the whole index | Follow-up (S11-21) | plan §3 Out and D-12 defer in-place update to S11-21; reviewer's own bench shows head (128 ms) beats base (181 ms) on the promote-then-propose loop, so this PR does not make it worse. Reported to the user |
| PRF-3 | PRF | minor | Common name words and `wide` give shortlists that grow with the catalog | Follow-up (S11-21) | not measured; S11-21 budget needs a named-person fixture |
| MNT-2 | MNT | minor | `matchIndex` is a one-field wrapper | Rejected | D-9's "S11-14 landed first" shape names this struct |

Nits (logged): `MNT-5 · promotealign/index.go:119 · fourth "Ref then ID" comparator` · `MNT-6 · valuecompare/shortlist.go:58 · wordKeys re-folds words nameWords folded` · `MNT-7 · shortlist/shortlist.go:95-107 · "wide"/"any" as constants` · `MNT-8 · add-match-rule/SKILL.md:62 · blank line splits "Add a comparer" list`

### Fix plan 1

**F-1 — Cap the year buckets a date may key** (from SEC-1)
Files: `core/shortlist/shortlist.go`, `core/matchregistry/alignment.go`, `core/matchregistry/registry_test.go`, `core/matchregistry/inventory_test.go` (if the guard needs it), `docs/matching-model.md`. Skill: `add-match-rule` ("Add a knob").
Failing test first: `TestKeys/huge range keys as open` in `core/shortlist/shortlist_test.go` — an event `{event_type: term(birth), date: Between(1, 10_000_000)}` makes fewer than 10 keys and includes the type's `wide` key; a lookup with the same date returns an open-dated birth and a 1850 birth (AC-2.2 still holds).
Do: add `MaxBuckets int` to `shortlist.Rules` (registry `Alignment.Shortlist`), default `100` with `// estimate: two centuries at YearBucket 2; a wider reach keys as open`. In `Keys`, when `floorDiv(hi) - floorDiv(lo) + 1 > MaxBuckets`, treat the date exactly as an open one (index `wide`, lookup `any`). A zero or negative `MaxBuckets` is not "no cap": validate it with the other registry numbers. Add the inventory row `alignment.Shortlist.MaxBuckets`.
Check: the test passes; `TestKnobInventoryMatchesRegistry` passes; scripts/check-pr.sh PASS.

**F-2 — Key an alias set by table and smallest token** (from SEC-2)
Files: `core/valuecompare/aliases.go`, `core/valuecompare/shortlist.go`, `core/valuecompare/shortlist_test.go`.
Failing test first: `TestAliasSetKeys/key does not grow with the set` — a table with one set of 1,000 tokens; every key `SetKeys` returns for one of its words is under 128 bytes; two words of that set still share a key; words of two different sets in the same table whose smallest tokens differ do not.
Do: `aliasSet.key` becomes `table.Key + " " + smallest token` (the sorted slice's first entry). Keep `SetKeys`' contract (two words that `Hit` share a key). Update the D-3 doc comment on `key` to say why it is bounded.
Check: the new and existing `TestAliasSetKeys`, `TestWordKeysShareOnSameOrEquivalent` pass; scripts/check-pr.sh PASS.

**F-3 — Bring `promote-graph-alignment.md` up to date** (from MNT-1)
Files: `docs/promote-graph-alignment.md`.
Do: line 107 drop `Stats`; line 185 seeds come from the handles the shortlist index offers (`core/shortlist`, `loadShortlist`), not `g.Kind`; line 196 "stats" → "the match index"; line 283 a warm proposal reads "the shortlist index if a handle or vocabulary commit dropped it (one read of every person, event and place node)".
Check: `python3 scripts/check-plan-docs.py` clean; scripts/check-pr.sh PASS.

**F-4 — A vocabulary commit drops the match index** (from MNT and PRF outside-lens notes)
Files: `core/database/graphcache/commit.go`, `core/database/graphcache/graph_test.go`.
Failing test first: `TestMatchIndexDrops/vocabulary commit` — set a match index, `OnCommit` with `Effects{Vocabulary: true}` and no handles, `MatchIndex()` reports not set.
Do: in `OnCommit`'s `set.Vocabulary` block, clear `matchIndexOK` and `matchIndex` beside `DropLabels`. Update `MatchIndex`'s doc and the D-12 sentence in the plan's Landed note to name vocabulary commits.
Check: the test passes; scripts/check-pr.sh PASS.

**F-5 — Skip the index when nothing needs seeding; time a real seed** (from PRF-2)
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/bench_test.go`, the plan's Landed note.
Failing test first: `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` — a Source whose only Subject is fixed; after `Propose`, `graphcache`'s `MatchIndex()` is not set (use the existing test seam or `ResetMatchIndexForTest` neighbor).
Do: in `loadCanon`, when `todo` is empty, skip `loadShortlist` and `seedHandles`. Give `BenchmarkProposeWarmHub` (and add `BenchmarkProposeAfterCommit` if cheap) a fixture with named persons (reuse `bulkCanon`, adding name values) and one unfixed named person on the Source, so it times the build and the lookup. Rerun both on base (bench file copied onto a `219a3b83` worktree) and head; replace the Landed note's numbers with these and say what fixture they used.
Check: the test passes; benchmarks run; scripts/check-pr.sh PASS.

**F-6 — Ride-alongs** (from MNT-3, MNT-4)
Files: `core/matchregistry/registry_test.go`, `core/datevalue/datevaluetest/*.go`, `core/shortlist/shortlist_test.go`, `core/valuecompare/shortlist_test.go`.
Do: `TestShortlistSection` uses `reg.Subject.Derived(kind, lead)`. Add `datevaluetest.Before(y, m, d)` and `After(y, m, d)` beside `About`; replace the `bef`/`aft` helpers in the two new test files with them. Do not touch older copies.
Check: tests pass unchanged in meaning; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `f27112b1` (commits `db987d96`..`f27112b1`). Gate (orchestrator) PASS against `219a3b83`. F-1 to F-6 done, each failing test seen red first. Deviations: `TestAliasSetKeys` expectations rewritten for the new key text; `b42fe719` alone fails `valuecompare` (fixed by `6527d987`).
F-5 benchmarks (named 20k persons, 30k events, one unfixed named person; `-benchtime 10x`): `BenchmarkProposeWarmHub` branch ~229 ms / 116 MB vs main ~480 ms / 102 MB; `BenchmarkProposeAfterCommit` (index rebuilt each op) branch ~522 ms vs main ~580 ms. The round-0 regression was the fixture (no seeding).
Minors read by the orchestrator: MNT-3 **Resolved** (`registry_test.go:60` uses `Subject.Derived`); MNT-4 **Resolved** (`datevaluetest.Before`/`After`; no local `bef`/`aft` left in the two new test files). PRF-2 rode with F-5; its guard is reviewed with the SEC+PRF round 2.

## Round 2 — review
Reviewers: SEC+PRF (SEC-1 blocker, SEC-2 major), MNT (MNT-1 major), CON (the vocabulary-commit major routed to it). Scope `69af2fb6...f27112b1`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker | Unbounded year span | **Resolved** | `shortlist.go:95-97` caps at `MaxBuckets`; overflow at the int extremes worked through by hand, not reachable |
| SEC-2 | SEC | major | Alias-set key grows with catalog content | **Not resolved** | F-2 bounds token count, not token or table-key length (`aliases.go:83`; `namealiases.Load` has no length check). Fix: fixed-size digest of `table.Key + "\x00" + sorted[0]`, plus a 4 KB test row. Open; to user (not a round-2 blocker, so no Fix plan 2 without the user) |
| MNT-1 | MNT | major | Stale `promote-graph-alignment.md` | **Not resolved** | F-4 made `docs/promote-graph-alignment.md:197` false ("a vocabulary-only write keeps it"); `graphcache/graph.go:270` `SetMatchIndex` doc omits vocabulary commits. Two lines. Open; to user |
| CON-x | CON | major | Vocabulary commit keeps the match index | **Resolved** | `commit.go:27-28`; `TestMatchIndexDrops/vocabulary commit` |
| PRF-2 | PRF | minor | Index built with nothing to seed | **Resolved** | `canon.go:75`; `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` |
| PRF-4 | PRF | minor | `Keys` multiplies duplicate type tokens by date buckets before dedupe | Follow-up | outside-lens, not a blocker; dedupe `types` before the loop (`shortlist.go:78-90`) |
| SEC-3 | SEC | major | `namealiases.Load` allocates `max(set_idx)+1` from catalog rows | Follow-up | code this PR does not change (`namealiases.go:158-164`); idea or map entry if the user agrees |

No new blockers in round 2 from any lens. CON round 2: Clean. Round cap reached.

**User call (2026-10-19):** run Fix plan 2 for SEC-2 and MNT-1 (round-1 majors left open by round 2); no round-3 review — the orchestrator reads the commits, then finishes.

### Fix plan 2

**F-7 — Alias-set key is a fixed-size digest** (from SEC-2)
Files: `core/valuecompare/aliases.go`, `core/valuecompare/shortlist_test.go`.
Failing test first: a row in `TestAliasSetKeysDoNotGrowWithTheSet` (or a sibling subtest) with a 4 KB smallest token and a 4 KB table key; every key `SetKeys` returns is under 64 bytes; two words of that set still share a key.
Do: `aliasSet.key` becomes the first 16 bytes of `sha256(table.Key + "\x00" + sorted[0])`, hex-encoded (`crypto/sha256`, `encoding/hex`). Keep `SetKeys`' contract; a collision only widens the shortlist — say so in the doc comment and drop "bounded by the table key and one token". Update the tests that assert key text (`TestWordKeys`, `TestAliasSetKeys`) to compare shared/not-shared, not literal key strings, where they now spell the old text.
Check: tests pass; scripts/check-pr.sh PASS.

**F-8 — Finish the match-index doc lines** (from MNT-1)
Files: `docs/promote-graph-alignment.md`, `core/database/graphcache/graph.go`.
Do: line 197 reads "dropped by any commit that names a handle, by a vocabulary commit (its keys read term keys and metas), and by new alias tables" and loses "a vocabulary-only write keeps it". `SetMatchIndex`'s doc (around line 270) names vocabulary commits beside handle commits, alias table loads and drops.
Check: `git grep -n -i "vocabulary-only" -- docs core` finds only test error strings, if any; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `a6712343` (commits `cd170431`, `200062dd`, `a6712343`). Gate (orchestrator) PASS against `219a3b83`. Read by the orchestrator (no round 3, per the user):
- SEC-2 **Resolved**: `aliasSetKey` is hex sha256 of `table.Key + "\x00" + smallest`; fixed 64 bytes (full digest: a `[:16]` literal trips `TestLayerPackages`' number guard). `TestAliasSetKeysDoNotGrowWithTheTokenOrTableKey` (4 KB key and tokens) red first.
- MNT-1 **Resolved**: `promote-graph-alignment.md:197` names vocabulary commits; `SetMatchIndex` doc too. Plan AC-3.4/T-8 rows still say vocabulary-only keeps the index; left as plan history (the Landed note records F-4).

## Result
Gate PASS · CI green · open: none · follow-ups: PRF-1 (S11-21), PRF-3 (S11-21), PRF-4, SEC-3 · to user: toponym-gap call (resolved), Fix plan 2 approval (resolved) · CON miss: CON-x (vocabulary commit kept the match index; raised by MNT and PRF in round 1)

## Round 3 — user-requested follow-ups
**User call (2026-10-19):** implement PRF-4 and SEC-3 on this PR (were Follow-up). No review round; the orchestrator reads the commits.

### Fix plan 3

**F-9 — A table's sets are a list, not slots by `set_idx`** (from SEC-3)
Files: `core/database/namealiases/namealiases.go`, `core/database/namealiases/namealiases_test.go`.
Failing test first: `TestLoadHugeSetIndex` — insert one `name_alias_members` row with `set_idx` 2,000,000,000 (and one set at 0) in a seeded test catalog; `Load` returns without allocating per index: the table has exactly the stored sets, in `set_idx` order, and no empty ones.
Do: in `Load`, replace `make([][]string, max+1)` and `Sets[idx] = tokens` with the sets in ascending `set_idx` order (sort the map's keys, append each). No caller reads a position as a `set_idx`: `valuecompare.NewAliasIndex` ranges the sets and skips empty ones; `add.go` reads `set_idx` from the table itself (`loadMemberSets`). Update `Load`'s doc and the `AliasTable.Sets` doc if either says positions are `set_idx`.
Check: the test passes; existing `namealiases` and `valuecompare` alias tests pass; scripts/check-pr.sh PASS.

**F-10 — Dedupe event types before keying dates** (from PRF-4)
Files: `core/shortlist/shortlist.go`, `core/shortlist/shortlist_test.go`.
Failing test first: `TestKeys/repeated type keys once` — an event with `event_type` term(birth) three times and one date: its index and lookup keys equal those of the same event with birth once. (A count assertion on the pre-`sortedUnique` slice is not reachable; assert equality of output, and keep the change itself to the dedupe.)
Do: after collecting `types` in `Keys`, sort and compact them (`slices.Sort`, `slices.Compact`) before building `indexTypes`/`lookupTypes`.
Check: tests pass; scripts/check-pr.sh PASS.

## Round 3 — fix
Head `b245aee6` (commits `d66b2a22`, `b245aee6`). Gate (orchestrator) PASS against `219a3b83`. Read by the orchestrator:
- SEC-3 **Resolved**: `Load` sorts the `set_idx` keys and appends; no slot per index. `TestLoadHugeSetIndex` red first (killed after 62 s allocating), now passes. `add.go` reads `set_idx` from the table, unaffected.
- PRF-4 **Resolved**: `slices.Sort` + `slices.Compact` on `types` in `Keys`. The F-10 test passed before the change (output was already deduped by `sortedUnique`; the saving is internal work), so it guards behavior only; the change applied as the fix plan named it.

## Result (final)
Gate PASS · CI see PR · open: none · follow-ups: PRF-1 (S11-21 scope), PRF-3 (S11-21 scope) · fixed on user request: PRF-4, SEC-3 · CON miss: CON-x
