# Ship log — S11-23 One home for primary kinds

**Plan:** [`S11-23-primary-kinds.md`](S11-23-primary-kinds.md) · **Branch:** `s11-23-primary-kinds` · **PR:** #438
**State:** ready

Worktree: `../provenencia-s11-23`, branched from `origin/main` @ `ca2d971d` (plan's base `ac05ae8`).

## Round 0 — implement
Base check: `ca2d971d` vs `ac05ae8` differs in docs only; Current state held (line numbers off by 1–5).

Head `6eeecbaf`. Gate PASS (orchestrator run, cold test cache). Report deviations: reworded the leading comment in `core/database/searchindex/handles.go` (it documented the deleted `Kind*` constants); manual drift check (§10 step 4) not run, the two guard tests pass; pre-existing gofmt drift in two untouched files left alone.

## Round 1 — review
Lenses: CON (1 major, 2 minor), MNT (1 major, 1 minor, 2 nit, 2 logged notes), UX clean, SEC clean, PERF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Plan has no Landed note listing the differences from the plan | Fix now → F-1 | verified: plan headings stop at `## 13. PR`; DoD line "a Landed note lists every difference" |
| MNT-1, CON-2a | MNT, CON | major | `docs/matching-model.md:197` names the deleted `ProfiledKinds` and `match.OriginProvenencia` | Fix now → F-2 | verified: line 197 still lists both; `rg ProfiledKinds core` empty |
| CON-2b | CON | minor | `add-searchable-kind` `KindSpec` table has no `Handle` row | Fix now → F-3 | verified: SKILL.md L47–48 list only `Priority`, `DefaultInEverything`. Rides along (fix plan already edits docs) |
| MNT-2 | MNT | minor → major | No test ties `KindSpec.Handle` kinds to `vocabkey.PrimaryKinds()` | Fix now → F-4 | verified: `registry.go:129-145`, no test reads `Handle`. Re-rated: `Handle` is new in this PR and is a second declaration of the primary kinds that no guard checks, which goes against the plan's "Next adds" claim and this PR's purpose. A test leaves D-9's runtime independence intact |
| CON-3, MNT-4 | CON, MNT | minor | `TestVocabKeysMatchSeed` hand-lists the eight Property constants | Follow-up (S11-02) | T-4 as written ("each of the eight") is met; D-8 exports no key list. S11-02's typed `PropertyKey` registry can drive the loop. Named in the Landed note |
| MNT-3 | MNT | nit | `IsPrimaryKind` switch and `PrimaryKinds` slice are two lists; T-2 checks one direction | Logged | nit |
| MNT-log-1 | MNT | — | `observations.go` value-type literals (S11-22 leftover) | Follow-up | D-13 leaves it out; outside this PR |
| MNT-log-2 | MNT | — | Per-kind header switches in `search_catalog.go` / `promote_align.go` | Follow-up | old debt; this PR only re-spells case labels |

### Fix plan 1

**F-4 — Guard the search registry's handle kinds against `vocabkey.PrimaryKinds()`** (from MNT-2)
Files: `core/search/registry_test.go` (or the package's existing registry test file).
Failing test first: none (a guard; it passes against the head). Show it is a real guard: temporarily drop `Handle: true` (or the `handle(KindPlace, 2)` call) locally, check the test fails, then revert. Record the result in the report.
Do: one test, e.g. `TestHandleKindsArePrimaryKinds`, that collects the `Kind` of every `Registry` entry with `Handle` set, sorts it, and compares it with sorted `vocabkey.PrimaryKinds()`. Runtime code is unchanged (D-9).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/search/` passes.

**F-2 — Fix `docs/matching-model.md` "Not knobs"** (from MNT-1, CON-2a)
Files: `docs/matching-model.md:197`.
Do: drop "`ProfiledKinds` (S11-23)". Replace `OriginProvenencia` with "the vocabulary origins and primary kinds (`core/vocabkey`)". Leave the landed S11-01 plan alone.
Check: `rg -n 'ProfiledKinds' docs/matching-model.md` is empty.

**F-3 — Add the `Handle` row to `add-searchable-kind`** (from CON-2b)
Files: `.agents/skills/add-searchable-kind/SKILL.md` (the `KindSpec` field table, about L47).
Do: one row: `Handle` — true for Conclusion handle kinds (person, event, place); set only by the `handle` helper in `init`; drives member counts and the handle field map; `TestHandleKindsArePrimaryKinds` (F-4) ties it to `vocabkey.PrimaryKinds()`.
Check: `scripts/check-plan-docs.py` passes.

**F-1 — Add the plan's Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-11/prs/S11-23-primary-kinds.md` (new `## Landed` section at the end, in the shape of `S11-22-pure-value-types.md`).
Do: first run §10 step 4's two mutation checks locally (add `case source` to `EvidencePrimaryKind` → `TestSwiftVocabularyMatchesGo` fails; add a `RoleRoot` type to a local `seedTypes` copy → `TestVocabKeysMatchSeed` fails; revert both) and record each result. Then list under "Differences from the plan:" the reworded leading comment in `core/database/searchindex/handles.go`; that `search.isHandleKind` is kept and reads the registry per D-7 / D-9, although the §3 In row says "deleted"; the AC-2.6 line numbers that moved; the §10 step 4 results; the added guard (F-4) and the two doc fixes (F-2, F-3); and that `TestVocabKeysMatchSeed` hand-lists the eight keys, with a typed key list left for S11-02.
Check: scripts/check-pr.sh PASS.

Order: F-4, then F-2, F-3, F-1 (F-1's Landed note describes the other three).

## Round 1 — fix
Head `48bfef56`. Gate PASS (orchestrator run). F-1 to F-4 done. F-4 guard confirmed: it fails with `handle(KindPlace, 2)` removed. §10 step 4's mutation checks were run and both fail as expected (recorded in the plan's Landed note).

## Round 2 — review
Lenses: CON, MNT (the lenses with Fix-now findings). Scope `6eeecbaf...48bfef56`.
CON: CON-1, CON-2a, CON-2b resolved; no new findings. MNT: MNT-1, MNT-2 resolved; no new findings. No blockers, so no Fix plan 2.

## Main merge
Merged `origin/main` @ `b578a47e` (S10-52, #437, which adds migration `000053`) so the app opens a catalog at the current format. Clean merge. The gate then failed `go-vet`: S10-52's `conclusiondetails_test.go:238` used `properties.OriginProvenencia`, which this PR deletes (D-5). One-item gate fix `cdf9e205` reads `vocabkey.OriginProvenencia`. Gate PASS (orchestrator run).

## Result
Gate PASS · CI pending · open: none · follow-ups: CON-3/MNT-4 (typed Property key list, S11-02), MNT-log-1 (`observations.go` value-type literals), MNT-log-2 (per-kind header switches in FFI handlers) · to user: —
