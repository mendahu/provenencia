# Ship log — S16-11 Outcome records live on the canonical node

**Plan:** [`S16-11-outcomes-on-node.md`](S16-11-outcomes-on-node.md) · **Branch:** `s16-11-outcomes-on-node` · **PR:** #793
**State:** ready

Base: `main` @ `7d2f3e7a`.

## Round 0 — implement
Head `9e0578fc`. Gate PASS (orchestrator run on `9e0578fc`). Report: all ACs done; Manual check not run (no UI change); base moved to `7d2f3e7a` with no Current-state file changed.

Measurements (STEP-1 before, STEP-8 after, same machine, large fixture):

| Figure | Before | After |
| --- | --- | --- |
| `nodes-B` at 4,000 persons | 45,342,376 | 57,034,832 |
| `nodes-B` at 16,000 persons | 58,249,320 | 76,868,264 (+18.6 MB; line 20 MB) |
| `nodes-ms` at 4,000 / 16,000 | 736 / 802 | 626 / 795 |
| Warm at open (city) | 694 ms | 761 ms |
| Join onto city | 77 ms | 76 ms |
| File one location | 75 ms | 83 ms |
| Unpromote | 39 ms | 46 ms |
| Delete location | 40 ms | 41 ms |
| Second part_of | 39 ms | 41 ms |
| Clear the portrait at the city (new, budget 500 ms) | 0.007 ms | 38 ms |

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF (all Opus). Skipped: UX (no `macos/App/**`, strings, `core/apperr` or error codes in the diff).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker | Gallery and list-row pictures return the stored `value_image_file_id` with no read-time media-type check | To user → Fix now → F-1 (user, 2026-11-12: validate in the load) | Verified: `graphcache/load.go:132,170` scans the column with no `files` join; `gallery.Of` (`gallery.go:59`) clones it as stored; base `sqlGallery` passed it through `depictionvalue.ReadableImage`. Column has no FK (000053) and survives open when `cache_version` matches. Contradicts D-12a and STEP-9's skill line ("returns it as stored", review M-3) |
| MNT-1 | MNT | major | Verify's `sameOutcomes` skips `SubjectID` | Fix now → F-2 | Verified: `verify.go:297` compares Observation, Reason, ValueRank, DeniedBy, votes only. D-4's field list predates review M-1 adding Subject; D-4's stated aim ("the node equals the committed rows, as `sameMembers` does") needs it. `observations/update.go:20` can move `subject_id` |
| MNT-2 | MNT | minor | Detail read builds gallery Property list twice; finds lead rows a second way | Fix now → F-3 | Code this PR adds (`conclusiondetails.go:191-195`, `:246-249`, `loadRecords`) |
| PRF-2 | PRF | minor | `gallery.Of` is values × outcomes per Property | Fix now → F-4 | Code this PR adds (`gallery.go:57-64`); group outcomes by rank once |
| CON-1, PRF-1 | CON, PRF | minor | Portrait-clear budget times a no-op clear; comment claims a commit reload; one-shot load reads every link | CON-1 part: Fix now → F-5. PRF-1 part (load only identity/values/outcomes in a one-shot graph): Follow-up | No-op verified: fixture never sets a portrait, `canonicalentities.go:286` returns early. 38 ms at 6,000 links is D-9's accepted cost; a narrower one-shot load is a graphcache change outside this PR |

Nits (logged): `MNT-3 · conclusiondetails.go loadRecords · hand-rolled cloneID on DeniedBy` · `MNT-4 · gallery.go Of · redundant ObservationIDs sort` · `MNT-5 · graph_test.go likenessObservation · rename` · `MNT-6 · graph_test.go propertyID · read from vocabulary` · `CON-2 · plan :33 · Landed note says figures are not in the ship log` · `CON-3 · selects_test.go:46 · Read case outside the table`

### Fix plan 1

**F-1 — Validate each value's image once, in the graphcache values load** (from SEC-1; user 2026-11-12)
Files: `core/database/graphcache/load.go`, `core/database/gallery/gallery_test.go` or `core/database/conclusiondetails/conclusiondetails_test.go` (wherever `TestForEntityLikenessNeedsAPNG` lives), `.agents/skills/add-depiction-value/SKILL.md`, `docs/conclusion-layer-data-model.md` §6 / `docs/conclusion-reconciliation.md` §7 if they repeat "as stored", the plan's Landed note. Skill: `add-depiction-value`, `add-catalog-query`.
Failing test first: restore `TestForEntityLikenessNeedsAPNG` to assert the refusal **on read**: after its `UPDATE files` (media type not `image/png`), drop the graph (`f.C.Graph().Drop()`) but do **not** run `autoreconciler.Rebuild`; expect no picture. It must fail at head `9e0578fc`.
Do: in the values load, `LEFT JOIN files f ON f.id = r.value_image_file_id` and select `COALESCE(f.media_type,'')`; store `ImageFileID: depictionvalue.ReadableImage(imageID, mediaType)` (cloned). Take the `files` table name from `catalogmodel` the way `valueTable` / `outcomeTable` are. Keep Verify consistent: if Verify compares node values with a fresh load or a truth recompute, both must go through `ReadableImage` so a non-PNG image is not reported as drift (confirm with `PROVENENCIA_GRAPH_VERIFY=1`). Rewrite the skill's "Read path" line: the graphcache values load passes `value_image_file_id` through `depictionvalue.ReadableImage`, so every node reader (gallery, lead pictures, likeness check) gets only readable images. In the plan's Landed note record that F-1 supersedes D-12a's setup edit and STEP-9's "as stored" line (user, 2026-11-12), and reword the note's "figures are in the PR description, not the ship log" (CON-2): the ship log holds them.
Check: the test passes; `TestOf`, `TestOfPreferred`, `TestSetPreferredLikeness` and the warm-open no-SQL test unchanged and passing; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes; `rg "as stored" .agents docs` prints no gallery line; scripts/check-pr.sh PASS.

**F-2 — Verify compares each outcome row's Subject** (from MNT-1)
Files: `core/database/graphcache/verify.go` (`sameOutcomes`), `core/database/graphcache/graph_test.go`.
Failing test first: a `TestVerifyChecksOutcomesAndLikeness` case that changes the joined Observation's `subject_id` with raw SQL (no commit), then `Verify` must error naming "outcomes". Fails at head.
Do: add `!sameID(o.SubjectID, w.SubjectID)` to `sameOutcomes`. Add a second case through the production path: move an Observation of a handle to another Subject with `observations.Update` inside a committed write, then Verify passes (the node reloaded). If that second case fails, stop and report it — it means the node goes stale on a real write.
Check: both cases pass; scripts/check-pr.sh PASS.

**F-3 — The detail read builds the gallery Property list once and passes the lead rows** (from MNT-2)
Files: `core/database/conclusiondetails/conclusiondetails.go`.
Do: `ForEntity` computes `gallery.PropertyIDs(vocab)` once and passes it (and the vocabulary) down; `fieldsOf` stops calling `Vocabulary()` / `PropertyIDs` again. `loadRecords` takes the lead outcome rows (the node's row for each tile's `ObservationIDs[0]`) instead of `galleryProps` plus `leads`, so the lead record is found one way. Ride along nit MNT-3: assign node `DeniedBy` directly (the node rows are cloned at load) or use one clone helper, not a hand-rolled copy.
Check: existing `conclusiondetails` and handler tests unchanged and passing; scripts/check-pr.sh PASS.

**F-4 — `gallery.Of` groups outcomes by rank once** (from PRF-2)
Files: `core/database/gallery/gallery.go`.
Do: per Property, one pass over `n.Outcomes[prop]` building `map[int][]obsID` of `InGallery` rows by `ValueRank` (rows already ascend by Observation id, so order is kept), then look up each value's rank. O(V+O). Ride along nit MNT-4: drop the per-picture `ObservationIDs` sort if the order is already ascending, or comment that it is defensive.
Check: `TestOf`, `TestOfPreferred` unchanged and passing; scripts/check-pr.sh PASS.

**F-5 — The portrait-clear budget says what it measures** (from CON-1)
Files: `core/database/promotealign/perf_test.go` (comment on `budgetClearPortraitAtCity`, ~:64-68).
Do: reword: the city has no portrait, so the clear is a no-op that commits nothing; the budget holds the one-shot graph's load of the city node (its 6,000 links). Comment only; no fixture change.
Check: scripts/check-pr.sh PASS.

Order: F-1, F-2, F-3, F-4, F-5. One commit per F.

## Round 1 — fix
Head `e53eee0ec`. Gate PASS (orchestrator run). F-1 `2d3fdb501`, F-2 `6c29e2958`, F-3 `82062dd4a`, F-4 `c7df44ea1`, F-5 `e53eee0ec`; all done. Deviations: F-1's `StoredValues` is also the autoreconciler's stored-vs-written read, which now sees the PNG-filtered image; F-2's raw-SQL case adds a second member to the handle so only outcomes differ; F-3 changed `fieldsOf`'s signature (two `export_test.go` call sites). F-2's production-path test `TestOutcomeSubjectFollowsAWrite` passes: the node reloads on `observations.Update`.

CON-1 (minor, F-5): **Resolved** — comment now says the clear is a no-op and the budget holds the one-shot node load (orchestrator read).

## Round 2 — review
Reviewers: SEC+PRF (SEC-1 blocker), MNT (MNT-1 major). Not sent: CON (minor only; resolved above). Scope: `9e0578fc...e53eee0ec`.

| ID | Lens | Round-1 severity | Status | Evidence |
| --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker | Resolved | `StoredValues` (`load.go:130-175`) joins `files` and stores `ReadableImage(id, media_type)`; every image reader goes through the node; autoreconciler stored-vs-written and Verify filter both sides; `TestForEntityLikenessNeedsAPNG` refuses on read |
| MNT-1 | MNT | major | Resolved | `verify.go:308` compares `SubjectID`; raw-SQL case and `TestOutcomeSubjectFollowsAWrite` |
| MNT-2 | MNT | minor | Resolved | Gallery Property list built once in `ForEntity`; `loadRecords` takes `leadRows` (both reviewers read it) |
| PRF-2 | PRF | minor | Resolved | `gallery.Of` builds `byRank` once per Property, O(V+O) |

New findings: none at blocker. Both reviewers Clean.

Nits (logged): `MNT-7 · conclusiondetails.go:422-425 · loadRecords doc comment still describes the lead search F-3 moved to ForEntity`

## Finish
Merged `origin/main` @ `b53c464d7` (S15-26 plan docs only); gate PASS on the merge.

## Result
Gate PASS · CI pending · open: none · follow-ups: PRF-1 (a write's one-shot graph loads the whole node, including every link, for a portrait check: 38 ms at the 6,000-link city; load only identity, values and outcomes) · to user: SEC-1 (answered 2026-11-12: validate in the load)
