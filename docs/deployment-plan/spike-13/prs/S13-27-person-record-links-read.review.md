# Review — S13-27 Person record links read

Plan: [`S13-27-person-record-links-read.md`](S13-27-person-record-links-read.md). Lint: `scripts/check-plan-docs.py` clean before round 1.

## Round 1 (2026-11-07)

Maintainability (opus): **Fits with changes**. Executability (sonnet): **Not ready**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | The key reloads on any touched handle, though every write that can change a line names a graph it holds (`effects/registry.go` `"identity_claims"` `claimGraph`) | **Fixed.** D-16: `.personClaims`' keyed set (own handle, value graphs, property labels); AC-3.1 – AC-3.3 and T-10 rewritten, with a row for a write outside the value. Alias and name-format writes report no effect (`None: true`) for any key: Out line |
| M-2 | major | The response repeats graph and Source labels `.personClaims` holds; send graph ids only | **Declined.** The S13-28 scope entry says "The view reads only S13-27's output" and the map row lists these fields; with M-1 both keys refresh on the same writes, so they cannot show two titles after a write. Noted in D-7's Rejected column; reported to the user, who may change the map instead |
| M-3 | major | `Lines` stays an exported unbounded entry | **Fixed.** D-4: renamed `pairLines`, re-exported for tests in `export_test.go`; `Pairs` the only exported comparison; DoD sweep |
| M-4 | major | `add-match-rule` (five knobs, `Lines`), `add-workspace-place` and the `databasetest` package doc go false | **Fixed.** STEP-10 updates both skills; D-14 widens the package doc; DoD sweep `rg 'five knobs'` |
| M-5 | minor | A second copy of the Person check, `Handle` read twice | **Fixed.** D-6: `personNode` and `claims(c, n)` in `claims.go`; `PersonEvents` uses `personNode` |
| M-6 | minor | Another Subject-to-entity scan over `EvidenceGraph.Members` | **Declined.** `RecordLinks` builds a map once per graph; `mergeFixed` walks members in order and needs every one; no shared access pattern. Out line |
| M-7 | minor | Keep the loader independent of page kind for Event and Place | **Fixed.** D-5: `linksBetween(c, page, graphs)` |
| M-8 | nit | Two hand-kept copies of the household fixture | **Fixed.** D-13: each helper's doc names the other and the shared numbers |
| X-1 | blocker | `Records` cannot put Subjects on a second graph (AC-1.4) | **Fixed.** STEP-3 builds with `claims_test.go`'s `fixture` (`person(graphID, …)`, `provisional`, `evidencegraphstest.Create`), as `burialsAndBaptisms` does; no `Records` change |
| X-2 | blocker | AC-1.1 and D-13 disagree; a parent link already counts a fact, so the weak / strong outcome was a guess | **Fixed.** §2 table of each graph's Subjects and links (no `relationship`, births on the father); expected facts per line; D-13 spells the household (four parent links, three strong lines); stop conditions both ways |
| X-3 | major | `TestPairsAtScale` and `householdGraphs` unspecified | **Fixed.** T-4 row; D-13 names ids, entities, `pt.Name`, `rel` |
| X-4 | major | No procedure for `MaxReadWork` and the benchmark; `PairWork` after the refactor | **Fixed.** D-3 (20 × `ReadWork` at 50 graphs), D-11 (`readWork`, `PairWork` derives the list as the one-pair count does), D-18 (`soloGraph`, m graphs on the names axis) |
| X-5 | major | Pair order under reversed input; capped pair marking; overflow | **Fixed.** T-2 states the mirror (n−1−B, n−1−A); D-2 (4), (5) say how capped lines read; counting stops past the cap, so no overflow |
| X-6 | major | AC-1.5 promote; provisional vs unfiled; `Open`'s close and user | **Fixed.** AC-1.3 splits provisional and never promoted (`fixture.provisional`); stop condition for AC-1.5; D-12: `Open` registers no close, upserts `UserID` beside the onboarding user |
| X-7 | minor | Wire id encoding | **Fixed.** D-7: `database.UUIDString` |
| X-8 | minor | Builder signature; `warmedVocabularySession` value | **Fixed.** D-17 signatures; STEP-9 names the value |
| X-9 | minor | Fuzz command, `since(` sweep false hits, `PairWork` / `Lines` doc | **Fixed.** STEP-1 check notes `go test` runs the seeds; DoD sweep uses `rg -w`; Out line for the docs |

Round 2 runs because round 1 had blockers.

## Round 2 (2026-11-07)

Maintainability (opus): **Fits with changes**, no new blocker. Executability (sonnet): **Ready with changes**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 (r2) | major | The S13-27 scope entry's Ships line still says the key refreshes on any promote or unpromote and that handles come "from the canonical graph" | **Fixed.** STEP-10 rewrites that Ships sentence as well as the Check line |
| M-3 (r2) | major | `Rules` field docs and `matching-model.md`'s `MaxPairWork` row still say "one Lines call" | **Fixed.** STEP-1 rewords them; DoD sweep `rg -w 'Lines'` on both files |
| X-1 (r2) | blocker | `k(k−1)/2` puts a literal 2 in `core/recordlinks`, which `TestLayerPackages` rejects | **Fixed.** D-2 (2): the i-th holder adds i, no halving; STEP-1 notes literals are 0 or 1 only |
| X-2 (r2) | major | `budget.go` build tag and the package doc's home | **Fixed.** D-14: `//go:build cgo` as `selects.go`; the doc stays in `selects.go`. No import cycle (round-2 maintainability checked) |
| X-3 (r2) | major | `BenchmarkPairsAtCap`: `exact`, the tag per graph, how m is found | **Fixed.** D-18 names `cfg.Value.Words.MaxExactPairing`, tag `g<i>`, and m grown from 2 |
| X-4 (r2) | major | Which builder makes which Subject in T-6 | **Fixed.** STEP-3 lists the builder per AC |
| X-5 (r2) | minor | "first by GraphID" ambiguous | **Fixed.** D-5: each `GraphID` once, at its first occurrence |
| X-6 (r2) | minor | `Open` on an onboarding catalog | **Declined.** Already a stop condition; `graphcachetest.NewCanon(t, c)` is the same shape inside `catalogsession.Do` in `person_detail_test.go` |
| X-7 (r2) | minor | Shared list filtering; over-bound entity work; per-pair cost | **Fixed.** D-2 (3) states all three |
| X-8 (r2) | — | Checks the reviewer could not run | Covered by stop conditions; no change |

Nits not raised by the reviewer (logged): `personNode` sits in `claims.go`; `databasetest.Since` reads like `time.Since`; a third caller of `valuerules.Load` + config would justify a helper.

## Verdict

Planned after round 2: Fits with changes / Ready with changes, every finding answered. Open: none. To user: M-2's decline — the user kept the labels (2026-11-07).
