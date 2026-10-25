# Ship log — S11-16 Rewrite graph alignment

**Plan:** [`S11-16-graph-alignment.md`](S11-16-graph-alignment.md) · **Branch:** `s11-16-graph-alignment` · **PR:** #501
**State:** ready

## Round 0 — implement
Head `63ae80d2` (local, not pushed; no PR). Gate not run.
STEP-1 done (`90d03c6c`; also added `segment.UnplacedAlike` to the `docs/matching-model.md` inventory so `TestKnobInventoryMatchesRegistry` passes). STEP-2 done (`63ae80d2`).
STEP-3 patch kept at the session scratchpad `step3.patch` (not committed).

**Stopped on:** Stop condition "`TestBirthIsNeverProposedOnADeath` needs a different target or band". With D-4's similarity gate, an event_type-only match (similarity 1, evidence 0.5) becomes a candidate. Under D-8 ("New only when it has no candidate"), it comes out Skip `weak` with an alternative instead of New `no_match`:
- `TestBirthIsNeverProposedOnADeath/AC-1.8` (pinned by D-13): birth, no date, vs canon birth, no date → Skip `weak` with EVT-1 as an alternative; it expects New `no_match`.
- Not pinned, but they fail D-13's STEP-3 rule (similarity ≥ 0.5, not a K/tie effect): `TestCandidateLoserDoesNotNominate`, `TestAcceptedPropertyMatchNominatesItsNeighbor` (EVT-9 is now an alternative), `TestRowReasons/a_surname-only_look-alike_is_not_offered` (now `weak`).
Orchestrator checked D-4, D-8, D-13 and AC-2.6: the plan contradicts itself. Taken to the user.

**User (2026-10-19):** option (a): keep D-4 and D-8, so a one-property match is a candidate and the row is Skip `weak`. Plan refreshed through plan-pr in `7e4a7a7b`. A fresh reviewer found the STEP-3 rule still forbade the change, and that Promote never shortlists undated events; both fixed (review log, "Refresh" section). Idea filed: `docs/ideas/coverage-aware-similarity.md`.
Resumed round 0 at STEP-3 with a new implementer.
Head `a24e7837` (local). STEP-3 done (`7b238ae4`; the four D-13 exception expectations changed, nothing else moved). STEP-4's `rows.go` move done (`a24e7837`). STEP-4/5 work in progress kept as scratchpad `step4-wip-spec-as-written.patch`.

**Stopped on:** D-7 "one that is already a candidate with no `via` gets `via = s`" makes `TestPartOfChainIsStrongWithoutADecision`'s city row read `via` instead of `agrees` / toponym. The target and band are right. D-13 pins reasons. Limiting the clause to a fixed neighbor passes the whole suite (graphalign, promotealign, handlers). Taken to the user.

**User (2026-10-19):** option (a): narrow D-7 so only a fixed row's existing candidate gets `via`. Plan refreshed in `8ef4984a` (one clause; the implementer had already run this variant green, so no separate plan review). Resumed round 0 at STEP-4 with a new implementer.
Head `7612e683`. Gate PASS (orchestrator run; go-test cached from the implementer's uncached PASS on the same tree). Report: STEP-4 and STEP-5 are one commit (`92c1ce3b`, the patch interleaves them); D-7 clause per `8ef4984a`; benchmark main 1.623 s/op vs branch 1.550 s/op.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes; the wire is unchanged).

CON: no findings (clean).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Placing on a place with many filed events nominates every filed neighbor per layer link (L × D candidates, ~3 scorings each) | To user → Follow-up (S11-21) | Measured (overlay bench): 30 births at a 6k-event city 2.88 s head vs 9.34 s base, so 3× faster than main, not a regression. The fix bounds nominations, which contradicts D-7 "no K"; S11-21 owns the large-Source budget |
| MNT-1 | MNT | blocker | `checkPure` has two import walkers; the second ignores `ImportDir` errors | Fix now → F-1 | verified: `guard_test.go:77-127`, `walk` and `walkBuild`; D-11 says one walker |
| MNT-2 | MNT | major | `promote-graph-alignment.md` §4 signature and §7 `seedCandidateLimit` are stale | Fix now → F-2 | verified by grep at lines 74 and 153 |
| MNT-3 | MNT | major | `add-match-rule` skill names the deleted `WeakScore` | Fix now → F-2 | verified: SKILL.md:34 |
| MNT-4 | MNT | minor | Dead `sc == nil` branch in `proposal`; dead `adjLink.back` | Fix now → F-3 | the fix plan touches these files; a few lines |
| MNT-5 | MNT | minor | Dirty-mark "first reason wins" rule written twice | Fix now → F-3 | a few lines in `align.go` |
| MNT-6 | MNT | minor | Contested rule inline in `proposal` | Fix now → F-3 | a few lines; S11-18 per-alternative flag reuses it |
| MNT-7 | MNT | minor | Corresponding-canon-link filter copied four times | Fix now → F-3 | `align.go` is rewritten in this PR; same-pass dedupe of touched copies |

Nits (logged): `MNT-8 · rows.go:169 · rewrap newPairSegments doc` · `MNT-9 · rows.go:250 · bestNonAssigned → bestExcept` · `MNT-10 · rows.go:71 · Bands.Weak guard repeats band` · `MNT-11 · align.go:31 · state.layer unused` · `MNT-12 · align.go:428 · local release bool name`

**User (2026-10-19):** defer PRF-1 to S11-21; the shape (a Source whose births are all located at one place with many filed events) and its remedy are added to S11-21's scope in the map (`70f26d69`).

### Fix plan 1

**F-1 — One import walker in `checkPure`** (from MNT-1)
Files: `core/matchregistry/guard_test.go`. Skill: `add-match-rule` ("One guard, not one per package").
Failing test first: none (a test-helper refactor); `TestLayerPackages` must still pass and still report each rule.
Do: replace `walk` and `walkBuild` with one `walk(pkg, chain, prodOnly bool)`. Walk `pkg.Imports` with the caller's `prodOnly`, and `TestImports` / `XTestImports` with `prodOnly=false`. Key `visited` by import path plus `prodOnly`, so a package first reached through a test import is walked again when it is reached through a production import. The `core/database` rule applies on every hop. The `core/matchregistry` rule (when `numbers`) and the `core/autoreconcile` rule (dir not `core/subjectprofile`) apply only when `prodOnly`. Keep the one `ImportDir` error path that reports `t.Errorf`. The doc comment says it is one walker with three rules (D-11).
Check: `go test ./core/matchregistry/...` passes. Temporarily add a production import of `core/matchregistry` in a numbers layer and confirm the walker reports it; revert that before you commit.

**F-2 — Stale docs and skill lines** (from MNT-2, MNT-3)
Files: `docs/promote-graph-alignment.md`, `.agents/skills/add-match-rule/SKILL.md`.
Do: in §4 of `promote-graph-alignment.md`, change the `Align` signature to `cfg Config` and drop the `// nil cfg: DefaultConfig()` comment. In §7 step 2, replace "the top 5 (`seedCandidateLimit`) property-only candidates" with "the candidates `graphalign.Candidates` keeps (`alignment.MinSimilarity`, the best `TopK`, ties up to `TieCap`), passed to `Align` in `Canon.Candidates`", and link the knob inventory in `docs/matching-model.md`. In SKILL.md:34, change `WeakScore` to `alignment.Bands.Weak`.
Check: `git grep -n 'DefaultConfig\|seedCandidateLimit\|WeakScore' -- docs/promote-graph-alignment.md .agents/` prints nothing.

**F-3 — Tidy the rewritten walk** (from MNT-4, MNT-5, MNT-6, MNT-7)
Files: `core/graphalign/align.go`, `core/graphalign/rows.go`.
Failing test first: none (no behavior change); the full `core/graphalign`, `core/database/promotealign` and `api/ffi/handlers` suites must pass unchanged.
Do:
- MNT-4: in `proposal` (`rows.go:23-33`), read `sc := st.cands[sk][string(hid)]` directly and drop the `sc == nil` branch, plus the `h != nil` guards that exist only for it. Drop the `back` field from `adjLink`, its four writes, and the mention in its doc.
- MNT-5: add `func (st *state) markDirty(nk string, m mark)`, which skips a fixed Subject and keeps the first mark (D-7). Call it from `place` and `release`.
- MNT-6: add `func (c *cand) contested() bool`, with the `RowFlags.Contested` doc sentence, and call it from `proposal`.
- MNT-7: add one `func (st *state) corresponding(link adjLink, hid []byte) []adjLink` returning the canon links of `hid` that correspond to `link`. Use it in `place`, both `segments` loops and `supports`, and keep each site's own extra filters (kind, owned) at the call site.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/ ./api/ffi/handlers/` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `6e64af81`. Gate PASS (orchestrator run). F-1 `52acc280`, F-2 `a1a14d5e`, F-3 `6e64af81` done. Deviation: kept the `h != nil` guard around `row.HandleRef` in `proposal` (it guards a different case).

## Round 2 — review
Reviewers: MNT (had a Fix-now blocker and majors). Not sent: CON (no findings), SEC+PRF (PRF-1 deferred to S11-21, nothing fixed). Scope `7612e683...6e64af81`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | (round 1) | Two import walkers | Resolved | one `walk(pkg, chain, prodOnly)`, `guard_test.go:78-123` |
| MNT-2 | MNT | (round 1) | Stale §4 / §7 | Resolved | `promote-graph-alignment.md:74`, `:153` |
| MNT-3 | MNT | (round 1) | Skill names `WeakScore` | Resolved | SKILL.md:34 |
| MNT-4–7 | MNT | (round 1 minors) | Dead branch/field, markDirty, contested, corresponding | Resolved (orchestrator read F-3 and the round-2 reviewer confirmed) | `markDirty`, `corresponding`, `(*cand).contested` each once |
| MNT-13 | MNT | minor | Kept `h != nil` guard in `proposal` is also dead | Fixed in round 3 (user asked) → F-4 | a placed handle always comes from `st.handles`; reviewer's minor check |
| PRF-2 | PRF (outside MNT) | minor | `corresponding` allocates a slice per call in `segments` | Follow-up (S11-21) | orchestrator overlay bench, births=30: fix head 3.09–3.22 s, 3.30 GB, 28.2M allocs vs round-1 head 3.68–3.94 s, 3.07 GB, 27.6M allocs; +7% bytes, no time regression |

Round 2 verdict: Clean. No round-2 blockers, so no Fix plan 2.

## Result
Gate PASS · CI green · open: none · follow-ups: PRF-1 (S11-21 scope, `70f26d69`), PRF-2 (S11-21), MNT-8–12 nits, coverage-aware similarity idea · to user: STEP-3 one-property candidates (a), STEP-4 D-7 `via` fixed-only (a), PRF-1 deferred · CON miss: none

## Round 3 — user request (2026-10-19)
The user asked to resolve MNT-13 on this PR and update from main. Merged `origin/main` (S12-03, #498) cleanly into the branch.

### Fix plan 3
**F-4 — Drop the dead `h != nil` guard in `proposal`** (from MNT-13)
Files: `core/graphalign/rows.go`.
Failing test first: none (no behavior change).
Do: in `proposal`, set `row.HandleRef = st.handles[string(hid)].Ref` directly and drop the `h` local and its `if h != nil` guard. A placed handle always comes from `st.handles`: `newState` places a fixed pair only when its handle exists, and `walk` places only popped candidates of known handles. If `h` is read anywhere else in that branch, index it directly the same way.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/ ./api/ffi/handlers/` passes; scripts/check-pr.sh PASS.

## Round 3 — fix
Head `775ea038`. Gate PASS against the new base `3bd08621` (orchestrator run). F-4 done: one line, no behavior change. Not re-reviewed: it removes a dead guard the round-2 reviewer had already traced.
