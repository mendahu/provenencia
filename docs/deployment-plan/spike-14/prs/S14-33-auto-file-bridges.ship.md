# Ship log — S14-33 A bridge between two handles files when it is made

**Plan:** [`S14-33-auto-file-bridges.md`](S14-33-auto-file-bridges.md) · **Branch:** `s14-33-auto-file-bridges` · **PR:** #723
**State:** ready

## Round 0 — implement
Head `b6996e68`. Gate PASS (run by the orchestrator on the pushed head, base `d5cefe54`). Report: Status line reads "Landed in #723 (2026-11-06)" because `check-plan-docs` rejects the shorter form; the deleted dogfood note was unlinked in the plan's Context line and STEP-7 (text only) so the link check passes; manual check not run (needs the Mac app).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (`core/database/**`, bridge filing over catalog data). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes; the only `macos/` file is the version bump in `project.pbxproj`).

The plan's §6 Scale ("No per-graph or per-catalog loop is added", "at most one handle and its observers") is wrong: it left out `findAssociation`, the `RecomputeReported` walk from the ends, and the graph cache. Measurements below use the session's throwaway fixtures: the reviewer's and verifiers' `go test -overlay` runs; nothing from them is committed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | `queryAssociation` scans every bridge of its type on a filed connect | Fix now → F-1 | read `bridgefile.go:488`: starts at `subject_types`, filters by `EXISTS`. Reviewer's EXPLAIN and timing: 27 ms at 400 filed participations, 374 ms at 1,600 (×4 input, ×14 cost); unfiled path flat at about 1 ms. A query change only: `findAssociation`'s rule stays |
| PRF-2 | PRF | blocker | A filed link reprojects every event at its place | Fix now → F-2 | verifier confirmed: 3.23 s and 10,004 `summary_handles` at 6,000 events on one place, 1 ms at base; `promote.Save` pays 3.66 s at base. Row-diff probe: no sibling row moves |
| CON-4 | PRF verifier, routed to CON | major | A filed connect names neither end to the client, so a cached detail page can stay stale | Fix now → F-2 | `Dependents` excludes its roots and `Handles` is the association alone; `conclusionDetail` follows `touchedHandles` by id (`CatalogQueryRegistry.swift:519`). Bar outcome: the link shows at once. Failing test proves it |
| MNT-1 | MNT | major | Mint-and-claim pair copied in `promotetest` and the effects test | Fix now → F-3 | same-pass cleanup: this PR adds a copy; supersedes the plan's Out line and the D-6 note on those copies. Test-only |
| CON-1 | CON | major | Plan has no Landed note | Fix now → F-4 | read: Status set, no note; DoD requires it |
| CON-2 / MNT-2 | CON, MNT | major | Map still links the deleted dogfood note | Fix now → F-4 | `deployment-plan.md:527`; the gate does not scan the map |
| MNT-3 | MNT | minor | `FileBridgesTx` doc says the caller recomputes the returned handles | Fix now → F-4 | stale comment in a file this PR changes; only the sentence, not the return shape |
| MNT-4 | MNT | minor | `promoteAllLocation` comment gives a reason this PR ended | Fix now → F-4 (To user, answered 2026-11-06) | Out line and D-5 said no Swift change; user approved this one comment, no code |
| MNT-5 | MNT | minor | New handler case threads `dir` through a shared variable | Fix now → F-4 | minor in a file this PR changes; `after` receives the request |
| MNT-8, MNT-9 | MNT | nit, raised to minor | Doc comments now part-true | Fix now → F-4 | stale comments in files this PR changes |
| PRF-3 | PRF | major | A filed connect drops Promote's match index | To user → deferred (answered 2026-11-06) | verifier: per-proposal cost unchanged, frequency up. At 37,000 handles the first same-graph proposal is about 1.35 s against 0.95 s, inside `budgetAfterCommit` (2.5 s) on a Mac; that test does not run at 37,000 and CI is about 3× a Mac. The reviewer's fix (skip the drops for bridges) is unsafe: stale Promote matches for participation, stale kin links for relationship. User accepted the cost; recorded in `docs/ideas/promote-ranking.md` by F-4 |
| PRF-4 | PRF | minor | A `part_of` connect reads every filed link of the term | Follow-up | 0.78 ms at 300 links, 2.97 ms at 1,200, measured; the plan named this read; quiet path |
| FU-1 | PRF verifier, routed to CON | major | After Unpromote or a bridge delete, documents and stored rows go stale | Follow-up | predates this PR: 234 stale documents over 451 unpromotes in the probe, 238 with F-2's walk (the wide walk repaired a few by accident). Likely `AssociationEndpoints` joining `canonical_entities`, so a deleted association has no ends. F-2 must not hide this |
| FU-2 | PRF verifier | minor | A located event's write still loads the place's links | Follow-up | 81 ms at 6,000 links, `graphcache.loadIDs`, after F-2 |

Nits (logged): `CON-3 · connect.go:50 · rewrap the doc comment` · `MNT-6 · connect.go:51 · same` · `MNT-7 · connect.go:50 · name the doc, not a bare §9.1`. F-4 does these while editing that comment.
Security: no findings.

### Fix plan 1

Order: blockers, then majors, then minors. The verifier's prototype is at `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/31694556-111e-4659-b976-94f07986645e/scratchpad/dependents_fix3.go`, with its probe at `/private/tmp/claude-501/-Users-jakerobins-Repos-provenencia/31694556-111e-4659-b976-94f07986645e/scratchpad/probe_test.go`. Read them as evidence, not as code to paste.

**F-1 — Find an association by seeking from an end** (from PRF-1)
Files: `core/database/identityclaims/bridgefile.go`, a test beside the `bridgeHeadsSQL` plan test (`export_test.go` already exports it). Skill: none.
Failing test first: an `EXPLAIN QUERY PLAN` test on the `queryAssociation` SQL for two ends and a term. It fails while the plan uses `subjects_subject_type_id_idx` or `observations_property_id_idx`, and passes when the outer loop seeks `identity_claims` by `entity_id` and then `observations` by `value_subject_id`. Extract the SQL builder the way `bridgeHeadsSQL` is.
Do: rewrite the query with the `CROSS JOIN` pinning `bridgeHeadsSQL` uses: from the accepted claims of `ends[0]`'s handle, to Observations naming those Subjects under that end's property, to the bridge Subject of `typeKey` and its accepted claim. Keep the remaining `EXISTS` on `o.subject_id = s.id`, using `+o.property_id` so they keep `observations_subject_id_idx`. Results must not change: same type, origin, property, status and term conditions, still `LIMIT 1`. Seek from the end that is not the hub: check that `ends[0]` is the `from` end (the event of a location, the child of a part_of), and if it is not, seek from the end that is, because a place can hold 6,000 links. If `observations` has no index on `value_subject_id`, stop and report.
Check: new test passes; `TestBridgeFiling`, `TestBridgeFilesWhenDrawnBetweenHandles` pass; scripts/check-pr.sh PASS.

**F-2 — Walk from the end a link changes, and name both ends** (from PRF-2, CON-4)
Files: `core/database/conclusionsummaries/dependents.go`, `core/database/autoreconciler/autoreconciler.go`, tests in `core/database/conclusionsummaries` or `core/search`, `TestLargeSourceBudgets`.
Failing tests first:
- `TestCreateCitedBridge` (handler) case: after a filed participation and a filed relationship, `effects.handles` plus `effects.summary_handles` hold both ends' handle ids.
- One location filed at a place with N sibling events: `summary_handles` holds no sibling event, and does not grow from N to 4N. The same for a `part_of` whose `to` end is that place.
- A positive pin: a `part_of` still reaches the events at `from` and its descendants and their birth or death people.
- A `TestLargeSourceBudgets` row, `budgetLocateAtCity` (500 ms), for one filed location at the 6,000-event city.
Do: make `AssociationEndpoints` also read `st.key` and `p.key`, and return two sets: every end, and the ends to walk. An end to walk is every end except `location.place` and `place_relationship.to`, named from `hops.PlacesOfEvent` and `hops.ParentsOfPlace` (`.Bridge()`, `.To()`) in one map, with the comment that no row reads a link through that end. In `RecomputeReported`, build `roots` and `movedRoots` from the ends to walk, and add every end, `before` and `after`, to the `deps` it returns, so the client hears about both ends without a walk from either. At `dependents.go:50` take `parents` out of the `located` input and keep `parents` in `out`. `Dependents`' doc comment says why.
Check: tests pass; `PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/... ./core/search/...` green; scripts/check-pr.sh PASS.

**F-3 — One in-transaction mint-and-claim step** (from MNT-1)
Files: `core/database/promote/promotetest/promotetest.go`, `core/database/effects/conclusion_handles_test.go`.
Failing test first: none (refactor; `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted`, `TestSaveBatchFilesOnlyItsGraph` and the `effects` tests hold it).
Do: add `promotetest.ClaimTx(tx *sql.Tx, s subjects.Subject)` holding the `canonicalentities.InsertTx` + `identityclaims.InsertTx` pair, returning the entity and the changes. `ClaimWithoutFiling` wraps it in `writes.Call`. Move the mint branch of the `claim` closure (`:198`) onto it, and replace the inline pairs at `:312-330` with it or with the existing `claim`. Supersedes the plan's Out line "An in-transaction mint-and-claim step…"; say so in the Landed note. No production change: `promote.fileSubject` has its own rules.
Check: `grep -rnE "canonicalentities\.InsertTx|identityclaims\.InsertTx" core --include='*_test.go' --include='promotetest.go'` prints no pair outside `ClaimTx` and the `claim` closure's non-mint branch; scripts/check-pr.sh PASS.

**F-4 — Comments, docs and the Landed note** (from CON-1, CON-2, MNT-2, MNT-3, MNT-4, MNT-5, MNT-8, MNT-9, PRF-3; nits CON-3, MNT-6, MNT-7)
Files: `api/ffi/handlers/connect_test.go`, `core/database/connect/connect.go`, `core/database/identityclaims/bridgefile.go`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `docs/deployment-plan/spike-14/deployment-plan.md`, `docs/ideas/promote-ranking.md`, the plan.
Failing test first: none.
Do:
- `connect_test.go`: delete `var dir string`; read `req.(*engine.CreateCitedBridgeRequest).GetProjectDir()` in `after`, as `promote_test.go:130` does.
- `connect.go` doc comment: rewrap, cite `promote-graph-alignment.md §9.1`, and say that a filing database error also rolls the write back.
- `bridgefile.go`: `FileBridgesTx`'s doc says `writes.Run` recomputes handles from the changes. "The claim that triggered filing still stands" becomes "the write that triggered filing".
- `EvidenceGraphModel.swift:962`: comment only. Promote stays open after the cards are filed, so a bridge left unfiled by an older catalog or a refused filing can be filed from this page. No code change.
- `deployment-plan.md:527`: replace the link with "dogfood note 2026-11-02 (deleted on land)".
- `docs/ideas/promote-ranking.md`, "Name index on commit": a filed connect (S14-33) is now a handle commit too. At 18,233 handles the first same-graph proposal after one is about 0.78 s against 0.57 s after an unfiled connect; at 36,977 handles about 1.35 s against 0.95 s. Skipping the drops is not safe: participation changes a Person's profile and relationship changes `canonlinks`. Location and place_relationship could skip both. Accepted by the user 2026-11-06.
- Plan: a Landed note listing every difference: T-5 split into two subtests; T-2 promotes Upper Canada too; the dogfood link removed from Context; the Status wording; §6 Scale was wrong (findAssociation, the dependents walk, the match index), with F-1 and F-2 as the fixes; `queryAssociation` and `RecomputeReported` changed; effects now name both ends; the Swift comment; PRF-3 accepted; F-3 supersedes the Out line. Decisions and Scope text stay as written.
Check: `rg relationship-auto-promote docs` prints only the plan's backticked mentions; `git diff` on the Swift file shows comment lines only; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `c64a2a5f`. Gate PASS (run by the orchestrator on the pushed head). F-1 to F-4 done. Deviations reported: `TestAssociationQueryPlan` was written after the rewrite, so it was never run red against the old query; `+o.value_term_id` added because the term subquery otherwise took `observations_value_term_id_idx`; `RecomputeReported` also reprojects every end, so an end's own document is still rewritten; `ClaimTx` takes a status for the effects test's provisional case; `requireEndsNamed` replaces the shared `ends` variable too. Budget row: 2.73 s before F-2, 96 ms after, at the 6,000-event city.

## Round 2 — review
Reviewers: CON (Fix-now majors CON-1, CON-2), MNT (MNT-1, MNT-2), SEC+PRF (PRF-1, PRF-2). Scope `b6996e68...c64a2a5f`.

CON: CON-1, CON-2 Resolved; CON-4 checked and Resolved. MNT: MNT-1, MNT-2 Resolved. PRF: PRF-1 Resolved (3.26 ms at 300 filed participations, 3.40 ms at 1,200); PRF-2 Resolved for a location and a link's `to` end (72 ms against 3.23 s at 6,000 events). Security: no findings. Round-1 minors, read by the orchestrator: MNT-4, MNT-5, MNT-8, MNT-9 Resolved; MNT-3 Not resolved: the new sentence reads oddly and overruns its line (CON-5, CON-6), finished in F-5.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-5 | PRF | blocker | A connect that joins an existing part_of reprojects every event at the `from` place though nothing moved | Fix now → F-5 | read `autoreconciler.go:233`: `movedRoots` adds the walk ends of associations outside `moved`. Reviewer: 2.75 s at 6,000 events, 21.5 s at 24,000; 0.14 s and 0.58 s with `movedRoots` gated. Connect is a new caller |
| PRF-6 | PRF | blocker | A filed location loads every link of its place, twice per write | To user → deferred (answered 2026-11-06) | supersedes FU-2. 72 ms at 6,000 links, the map's Scale; 306 ms at 24,000. The fix is non-trivial graph cache work. User accepted the cost; noted as a possible fix in `docs/ideas/promote-ranking.md` by F-6 |

Nits (logged): `CON-5 · bridgefile.go:22-25 · rewrap` · `CON-6 · bridgefile.go:23 · say which set writes.Run recomputes`. Both finish MNT-3 in F-5.
Follow-up homes (user, 2026-11-06): FU-1 → a new row on this spike's map, in F-6. PRF-4; FU-3, the first `part_of` filed from a city (1.84 s at 6,000 events, 17.0 s at 24,000; rows do move; the reprojection cost S14-21 deferred); FU-4, `subjects.Delete` still walks from the place end (not measured, MNT round 2) → `docs/ideas/structural-debt.md`, in F-6.

### Fix plan 2

**F-5 — Walk only from the ends of associations that moved** (from PRF-5; finishes MNT-3)
Files: `core/database/conclusionsummaries/dependents.go`, `core/database/autoreconciler/autoreconciler.go`, `core/database/promotealign/perf_test.go`, `core/database/conclusionsummaries/filedlinks_test.go`, `core/database/identityclaims/bridgefile.go`.
Failing test first: a `TestLargeSourceBudgets` row, `budgetSecondPartOfAtCity` (500 ms): a second record's `part_of` from the 6,000-event city, both Subjects joined to their handles, so the bridge joins the existing association. About 2.75 s today.
Do: `AssociationEndpoints` also selects `r.entity_id` and returns the walk ends by association id. In `RecomputeReported`, `roots` stays every association's walk ends and `deps` stays the full walk, as the existing comment says. `movedRoots` becomes `moved` plus the before and after walk ends of the associations in `moved` only. `ReprojectHandles` still gets `roots` and `ends`: a root's own document is always rewritten.
Add a positive pin beside `TestFiledPartOfWalksFromThePart`: a part_of that does move a row (the first one filed from a place) still rewrites the search document of an event at `from`.
`bridgefile.go`: rewrap `FileBridgesTx`'s doc. Say the ids are the handles filing touched, and that the caller discards them because `writes.Run` recomputes every handle named by the changes.
Check: tests pass; `PROVENENCIA_GRAPH_VERIFY=1 go test -tags fts5 ./core/database/... ./core/search/...` green; scripts/check-pr.sh PASS.

**F-6 — Record what the user placed** (from PRF-6, FU-1, PRF-4, FU-3, FU-4)
Files: `docs/ideas/promote-ranking.md`, `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, the plan's Landed note, `docs/deployment-plan/spike-14/completed.md` if it names these costs.
Failing test first: none.
Do:
- `promote-ranking.md`: a possible fix. A filed location loads the place's whole link list twice in one write (`graphcache.loadIDs`, from `OnCommit`'s reload and `composeEvents`): about 72 ms at 6,000 links, 306 ms at 24,000. Patching the loaded node with the one new link would remove it. Accepted by the user 2026-11-06.
- `structural-debt.md`, three rows, `S14-33 ship log` in the Was column: a `part_of` connect reads every filed link of the term (`filedLinks`, 2.97 ms at 1,200 links; walking up from the new link's `to` end would follow only the reachable chain); the first `part_of` filed from a city reprojects its events (1.84 s at 6,000, 17.0 s at 24,000; the rows do move, and the cost is faster than linear; the same reprojection cost as S14-21's follow-up); `subjects.Delete` calls `Dependents` directly, so deleting a location bridge still walks from the place end (not measured).
- `deployment-plan.md`: take the next free S14 number and add a row in Slice 7c's table plus a short scope entry, status Scoped, no plan yet: after Unpromote or a bridge delete, search documents and stored summary rows go stale. Probe: 234 stale documents over 451 unpromotes, 45 stale display rows. Likely cause: `AssociationEndpoints` joins `canonical_entities`, so a deleted association has no ends. Bug, predates S14-33.
- Plan's Landed note: add `movedRoots` gated by F-5, PRF-6 accepted, and the new map row.
Check: `scripts/check-plan-docs.py` clean; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `10e34e8f`, then `main` merged (`1f5b38ab`; only S13-37's plan docs had landed). Gate PASS on the merged head. F-5 done: `TestLargeSourceBudgets/second_part_of_from_the_city` 2.87 s red, 144 ms green; `TestFiledPartOfReprojectsEventsOnlyWhenARowMoved` pins that a part_of that moves a row still rewrites its event's document, and ran red on its second half. F-6 done: the new map row is S14-38. PRF-5 Resolved by the orchestrator's read of the diff; there is no round 3, so no reviewer has seen F-5. MNT-3 Resolved.

## Result
Gate PASS · CI pending · open: none · follow-ups: S14-38 (FU-1), `promote-ranking.md` (PRF-3, PRF-6), `structural-debt.md` (PRF-4, FU-3, FU-4) · to user: PRF-3, PRF-6, MNT-4, all answered 2026-11-06 · CON miss: CON-4
