# S11-50 — Matching warms while the project opens

**Status:** Landed in [#570](https://github.com/mendahu/provenencia/pull/570) (2026-10-26)
**Spike / slice:** Spike 11 · Slice 9 — Kinship shapes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the open spinner already covers `OpenProject`; the work moves under it)
**Board:** —
**Depends on:** —
**Blocks:** S11-39
**Branch:** `s11-50-warm-matching-at-open`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rule "Matching warms behind the open spinner"; [Kinship shapes § Where implied links live](../kinship-shapes.md#where-implied-links-live) ("Warm at open"); S11-21 D-1 (the budgets); [#567's ship log](https://github.com/mendahu/provenencia/blob/s11-39-matching-reads-kinship/docs/deployment-plan/spike-11/prs/S11-39-matching-reads-kinship.ship.md) (cold profile)
**Planned:** 2026-10-26 against `main` @ `13dab35f`
**Review:** [`S11-50-warm-matching-at-open.review.md`](S11-50-warm-matching-at-open.review.md) — Fits / Ready, 2 rounds (every finding answered)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Opening a project loads every person, event and place into the graph cache and builds Promote's name index, under the spinner the open already shows. The first Promote after open then pays only for its own Source: ranking candidates, the walk and alignment. An operation that throws the warmed state away (a vocabulary commit, a structure commit, a rebuild, an alias or name-format change) warms it again before it returns.

**Why now.** On S11-39's 12-generation fixture the first proposal spent about 0.65 s of 1.05 s locally (three times that on CI) loading nodes and building the index, and CI's cold budget failed. The user's call (2026-10-26): "Move everything behind the spinner so that it's honest. Then we will address that problem later if it grows too long."

**What stays.** Ranking, the walk and alignment run on the first proposal as today. A handle commit still drops the name index and the next proposal rebuilds it ([`docs/ideas/name-index-on-commit.md`](../../../ideas/name-index-on-commit.md)). No UI change.

## 2. User stories

**US-1.** As a researcher, when I open a project, the open spinner covers the matching warm-up, so the first Promote does not stall on it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A project made by `onboarding.Complete` (which opens it through `catalogsession.Do`), then one more `Do` | The graph holds a match index, `WarmNeeded()` is false, and each shortlist kind is loaded: `g.Kind(typeID)` for person, event and place issues no SELECT. A catalog without those seeded types has no kind to load; the AC covers onboarding-made projects |
| AC-1.2 | AC-1.1, then `promotealign.Propose` on a Source | It issues no SELECT for a shortlist kind load and builds no index (the match index generation is unchanged) |
| AC-1.3 | An empty new project opened through `Do` | Opens; the match index is present |

**US-2.** As the engine, a call that throws the warm state away warms it again before it returns, and a handle commit does not.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | AC-1.1, then a `Do` whose `fn` calls `c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}})` | After it returns, the match index exists (rebuilt: the commit had dropped it) |
| AC-2.2 | AC-1.1, then a `Do` whose `fn` runs `c.Graph().Drop()` | After it returns, the kinds are loaded and the match index exists |
| AC-2.3 | AC-1.1, then a `Do` whose `fn` runs `namealiases.Warm(c)` | After it returns, the match index exists |
| AC-2.4 | AC-1.1, then a `Do` whose `fn` calls `c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Handles: [][]byte{id}}})` for a 16-byte `id` (a handle commit, no list or structure effect) | After it returns, the match index is absent and `WarmNeeded()` is false (the next proposal rebuilds the index, as today) |
| AC-2.5 | AC-1.1, then a `Do` whose `fn` runs `c.Graph().Drop()` and returns `errors.New("boom")` | `Do` returns that error; no warm runs: the match index is absent and `WarmNeeded()` is true |

**US-3.** As a contributor, the budgets say what the researcher waits for: the warm at open, then the first proposal.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `TestLargeSourceBudgets` subtest "warm at open": `Drop`, then `catalogsession.WarmOnOpen` | Within `budgetWarmOnOpen` (2.5 s) |
| AC-3.2 | Subtest "cold": `Drop`, `WarmOnOpen` (untimed), then one proposal | Within `budgetCold` (2.5 s); likewise "births at one place cold" within `budgetOnePlaceCold` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` warm flag: set by `dropDerived` (so `New`, `clear` and `Drop` too); read by `WarmNeeded`, cleared by `MarkWarm`; one `dropMatchIndex` | STEP-1, AC-2.4 |
| `promotealign.loadConfig` *(new)*, shared by `loadInputs`, `Warm` and `LayerBridges`; `promotealign.Warm(c)` *(new)*: loads the shortlist kinds and builds the index | STEP-2, AC-1.1, AC-1.2, AC-1.3 |
| `WarmOnOpen` warms matching last and marks the graph warm; `Do` runs `WarmOnOpen` after `fn` when `WarmNeeded` | STEP-3, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| Budgets: "warm at open"; cold is the first proposal after it | STEP-4, AC-3.1, AC-3.2 |
| Docs | STEP-5 |

| Out | Home |
| --- | --- |
| Keeping the index current on a handle commit | [`docs/ideas/name-index-on-commit.md`](../../../ideas/name-index-on-commit.md) |
| Ranking candidates before a Source is opened | not planned — it depends on the Source |
| Warming kinship pivots | not planned — derived on first read (S11-49) |
| A background warm, a progress message on the spinner | not planned — only if opens grow too long (user, 2026-10-26) |
| The 12-generation tree in the fixture | S11-39 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Open | `core/catalogsession/session.go` `openResearcher` | `database.Open`, `ReconcileOnOpen`, then `WarmOnOpen`; reached from `OpenProject` → `onboarding.Open` → `catalogsession.Do` while the onboarding model shows its spinner (`OnboardingModel.open`, `isBusy`) |
| Warm | `session.go` `WarmOnOpen(c)` | `namealiases.Warm`, then `nameformats.Warm`. Each sets its value on the graph, which calls `dropDerived` |
| Do | `session.go` `Do(projectDir, fn)` | Locks the session, runs `fn`, unlocks |
| Drops | `core/database/graphcache/graph.go` `clear` (from `Drop` and a structure commit), `dropDerived` (vocabulary commit in `commit.go` `OnCommit`, `SetAliasIndex`, `SetNameFormat`) | Empty the match index; `clear` also empties nodes and kinds. A handle commit sets `matchIndex = nil` directly in `OnCommit` |
| Index | `core/database/promotealign/index.go` `loadShortlist(g, prof, cfg)` | Builds on first proposal: `g.Kind` for each shortlist kind, `keyValues`, `shortlist.Index.Add`; stores with a new generation |
| Profiler | `core/database/promotealign/profiles.go` `newProfiler(g, metas, rules, namePattern)`, `loadMetas(g)` | Catalog-wide inputs |
| Config | `core/database/promotealign/promotealign.go` `loadInputs` (lines 101–106) and `core/database/promotealign/export_test.go` `LayerBridges` (lines 19–24) | Both build the config the same way: `valuerules.Load`, `matchregistry.Default().AlignConfig()`, `cfg.Value = rules` |
| Index drops | `core/database/graphcache/graph.go` `dropDerived`, `ResetMatchIndexForTest`; `core/database/graphcache/commit.go` `OnCommit` handle branch | Each clears `matchIndexOK` and `matchIndex` inline |
| Budgets | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`; constants `budgetCold` … `budgetOnePlaceWarm` | "cold" and "births at one place cold" time `Drop` + one proposal; the fixture calls `catalogsession.WarmOnOpen(f.c)` |
| Tests | `core/catalogsession/warm_test.go` `TestDoWarmsNameFormat` | Precedent: open through `Do`, assert a warm value and zero SELECTs with `databasetest.CountSelects` |

**Precedent.** The name format warm (`nameformats.Warm`, `TestDoWarmsNameFormat`): a value built once at open, stored on the graph, asserted with a SELECT count. Copy its test shape.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `func loadConfig(g *graphcache.Graph, q database.Querier) (graphalign.Config, string, error)` *(new)* in `promotealign.go` returns the config with `Value` set and the name pattern; `loadInputs` and `export_test.go` `LayerBridges` call it. `promotealign.Warm(c *database.Catalog) error` *(new)*: `loadConfig`, `loadMetas`, `newProfiler(g, metas, cfg.Subject, namePattern)`, then `loadShortlist` | The config the index is built with and the one proposals rank with come from one place (round 1 M-1); one build path | A third copy of the config lines in `Warm`; building the index in `catalogsession` — it would copy `promotealign` internals |
| D-2 | `WarmOnOpen` is the one list: `namealiases.Warm`, `nameformats.Warm`, `promotealign.Warm`, then `c.Graph().MarkWarm()` | Setting the alias index or name format calls `dropDerived`, which would drop an index built before them; one list means a value added later is warmed at open and after a drop alike (round 1 M-2) | Two lists, one at open and one in `Do` |
| D-3 | `graphcache` gains `warmNeeded bool` *(new)*, `WarmNeeded() bool` *(new)* and `MarkWarm()` *(new)*, and one `dropMatchIndex()` *(new)* that clears `matchIndexOK` and `matchIndex`. `dropDerived` calls `dropMatchIndex`, drops the Evidence graph memos and sets `warmNeeded` (`New` → `clear` → `dropDerived` covers a new graph and `Drop`). `OnCommit`'s handle branch and `ResetMatchIndexForTest` call `dropMatchIndex` only | One flag every warm-state drop already passes through; a handle commit's index drop stays the idea file's job (rebuilding on every write would put about 70 ms on each edit). Two named drops make that rule visible (round 1 M-3) | Checking `MatchIndex()` in `Do` — a handle commit would rebuild on every write |
| D-4 | `Do`: after `fn` returns nil, when `c.Graph().WarmNeeded()`, run `WarmOnOpen(c)` inside the lock and return its error. When `fn` returns an error, return it and skip the warm | The call that threw the state away pays for it, behind whatever the caller already shows; an error leaves the flag set for the next call | Warming lazily on the next proposal (today) — the cost lands on Promote; a warm hook inside `graphcache` — the graph cannot import `promotealign` and `SetAliasIndex` would re-enter it |
| D-5 | `budgetWarmOnOpen = 2500 * time.Millisecond` *(new)* | The open absorbs what the first proposal paid; about 0.45 s locally on today's fixture, three times that on CI. 2.5 s keeps S11-21's cold figure as the open's ceiling | Timing the full `catalogsession` open — reconcile and migration checks are not matching's |
| D-6 | A subtest "warm at open" before "cold", shaped like "cold": `fastest(2, func() time.Duration { f.c.Graph().Drop(); return since(func() { must(t, catalogsession.WarmOnOpen(f.c)) }) })`. "cold" and "births at one place cold": each run `Drop` and `must(t, catalogsession.WarmOnOpen(f.c))` outside `since`, then time one proposal. Their budgets are unchanged. "memo matches a fresh run" keeps `Drop` + proposal (no warm), so the lazy path stays tested | Cold now means the first proposal after open | A new constant for cold-after-open — the old figure still fits |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `loadConfig` *(new)* | func | `loadInputs`, `LayerBridges` (the two copies it replaces) | Extracted from them |
| `dropMatchIndex` *(new)* | func | three inline clears | Extracted from them |
| `MatchIndexGenForTest` *(new)* | test export | `ProfilesKeptForTest` in `core/database/promotealign/export_test.go` | Same file, same shape |
| `Graph.warmNeeded`, `WarmNeeded`, `MarkWarm` *(new)* | field, funcs | `Graph.matchIndexOK`, `aliasIndexOK` | Those say one value is present; this says the open-time warm must run again, set only by the drops that lose it |
| `promotealign.Warm` *(new)* | func | `loadShortlist` (called by it) | A public entry over the existing build |
| `budgetWarmOnOpen` *(new)* | const | `budgetCold` | A separate figure for the open |
| `TestDoWarmsMatching`, `TestDoRewarms` *(new)* | tests | `TestDoWarmsNameFormat` | Same shape |

**Builds on.** `WarmOnOpen`'s one list; `loadShortlist`; `databasetest.CountSelects`.
**Deletes.** The config lines in `loadInputs` and `LayerBridges` (now `loadConfig`); the three inline match-index clears (now `dropMatchIndex`).
**Next adds.** Another open-time value adds one call in `WarmOnOpen`. The name-index-on-commit idea makes AC-2.4 rebuild nothing.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `warmNeeded`; `WarmNeeded() bool`, `MarkWarm()`, `dropMatchIndex()` *(new)* per D-3 |
| `core/database/promotealign` | `loadConfig` *(new)* in `promotealign.go`; `func Warm(c *database.Catalog) error` *(new)* in `index.go`; `func MatchIndexGenForTest(g *graphcache.Graph) uint64` *(new)* in `export_test.go` (the stored index's generation, 0 when none) |
| `core/catalogsession` | `WarmOnOpen` per D-2; `Do` per D-4; package doc and `WarmOnOpen` comment say what is warmed and when it runs again |

### FFI

None — `OpenProject` is unchanged; its existing call reaches `WarmOnOpen`.

### macOS

None — the spinner in `OnboardingModel.open` already covers the call.

### Strings

None.

## 7. Steps

**STEP-1 — The warm flag.**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/commit.go`, `core/database/graphcache/graph_test.go`.
Failing test first: `TestWarmNeeded` (T-1).
Do: D-3.
Check: T-1 passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/` passes.

**STEP-2 — `promotealign.Warm`.**
Files: `core/database/promotealign/promotealign.go`, `core/database/promotealign/index.go`, `core/database/promotealign/export_test.go`, `core/database/promotealign/memo_test.go`.
Failing test first: `TestWarmBuildsTheIndex` (T-2).
Do: D-1. Extract `loadConfig` first and run the package's tests unchanged, then add `Warm` and `MatchIndexGenForTest`.
Check: T-2 passes; `go list -deps ./core/database/promotealign | grep -c core/catalogsession` prints 0 (no cycle for STEP-3).

**STEP-3 — Warm at open and after a drop.**
Files: `core/catalogsession/session.go`, `core/catalogsession/warm_test.go`. Skill: `use-catalog-session`.
Failing test first: `TestDoWarmsMatching` (T-3), `TestDoRewarms` (T-4).
Do: D-2, D-4.
Check: T-3, T-4 pass; `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes.

**STEP-4 — Budgets.**
Files: `core/database/promotealign/perf_test.go`.
Do: D-5, D-6; add subtest "warm at open" before "cold".
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` passes; report the three cold-side figures in the PR body.

**STEP-5 — Docs.**
Files: `docs/promote-graph-alignment.md` (one paragraph: the shortlist kinds and the name index are built at open and after a warm-state drop; a handle commit drops the index until the next proposal), `.agents/skills/use-catalog-session/SKILL.md` (`WarmOnOpen` warms matching; `Do` runs it again after a drop), `docs/deployment-plan/spike-11/completed.md` (newest first), `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports 0 errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/graph_test.go` `TestWarmNeeded` *(new)* | on `openCatalog(t)`'s graph (`graphcachetest.OpenCatalog` once S11-49 has landed): new graph → true; `MarkWarm` → false; `SetMatchIndex(struct{}{})`, then `OnCommit(effects.Commit{Effects: effects.Set{Handles: [][]byte{id}}})` → false, match index absent; `Drop` → true; `MarkWarm`, then `OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}})` → true; `MarkWarm`, `SetNameFormat` → true | D-3, AC-2.4 |
| T-2 | Go | `core/database/promotealign/memo_test.go` `TestWarmBuildsTheIndex` *(new)* | `newWarmFixture`, then `promotealign.Warm(f.c)` → `MatchIndexGenForTest` is non-zero; `g.Kind` for person and event issues 0 SELECTs (`databasetest.CountSelects`); `Propose` on Source B leaves the generation unchanged | AC-1.2, D-1 |
| T-3 | Go | `core/catalogsession/warm_test.go` `TestDoWarmsMatching` *(new)* | project from `onboarding.Complete`, then `Do` → `g.MatchIndex()` present, `WarmNeeded()` false, `g.Kind` for person, event, place issues 0 SELECTs (the project is empty, so this also covers AC-1.3) | AC-1.1, AC-1.3 |
| T-4 | Go | `core/catalogsession/warm_test.go` `TestDoRewarms` *(new)* | one subtest per AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, each a `Do` whose `fn` does what the AC says; assertions read `c.Graph()` in a second `Do` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, D-4 |
| T-5 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "warm at open", "cold", "births at one place cold" | within 2.5 s, 2.5 s, 5 s | AC-3.1, AC-3.2, D-5, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/catalogsession` tests (open now builds an index), `promotealign` memo tests that count queries after `WarmOnOpen`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS
- [ ] CI's Go tests workflow passes TestLargeSourceBudgets; the PR body lists its warm-at-open and cold figures
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Domain docs updated: promote-graph-alignment.md, use-catalog-session SKILL.md
- [ ] PR description: what shipped, AC evidence, budget figures, deviations, manual check
```

## 10. Manual check

1. Build and run the app; open a project with a few hundred people. The open spinner shows for its usual time or a little longer.
2. Open a Source and Promote. The first proposal appears without the pause it had before on the same project.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Opens of very large catalogs grow long | Accepted by the user; revisited if it grows too long (background warm) |
| A test that counts queries after `Do` now sees the warm | T-3 and T-4 fix the expectation; Stop condition below for others |
| A drop path that bypasses `clear` and `dropDerived` | Stop condition |

**Stop and report if:**

- `promotealign` imports `catalogsession` directly or through a dependency (STEP-2's check prints more than 0).
- A path empties the match index or the kinds other than `OnCommit`'s handle branch, `clear`, `dropDerived` or `ResetMatchIndexForTest`.
- An existing test outside `catalogsession` and `promotealign` fails because opening now builds the index.
- "warm at open" exceeds 2.5 s on CI.

## 12. Open questions

None.

## 13. PR

**Title:** S11-50 — Matching warms while the project opens
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; CI budget figures; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-50-warm-matching-at-open.md`.

## Landed

Differences from the plan:

- D-4 was revised by the user on 2026-10-26 after review: "Change D-4: a matching warm failure never fails the open or a committed call (leave the index unbuilt; the next proposal reports the error); after a mid-session drop, re-warm only the alias index and name format, and leave the shortlist index to the next proposal." `WarmOnOpen` runs `warmGraph` (the alias index, the name format, then `MarkWarm`) and returns its error only, then `promotealign.Warm` with its error ignored. `Do` runs only `warmGraph` after `fn` returns nil and the graph needs a warm, and its error never replaces `fn`'s result. The shortlist index is built at open only. `TestDoRewarms` asserts this, and `TestOpenSurvivesABadValue` covers a value that does not decode.
- The tests were written with the code, not seen failing first.
- The Status reads `Landed in [#570](...) (date)`, because `check-plan-docs.py` rejects a bare `Landed #570`.
- The CI budget figures go in the PR body when CI runs.
