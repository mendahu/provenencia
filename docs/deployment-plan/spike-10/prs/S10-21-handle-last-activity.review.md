# Review — S10-21 Handle last-activity scope

**Verdict:** Fits with changes — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (Opus 5.5), executability (Sonnet 5.5)

## To user
- None. The two product calls were asked and answered before round 1 (2026-10-14): the handle scope is claims, pins, the handle row, and member Observations with their values and notes (D-2); a handle with no scope rows is still listed with a zero revision (AC-1.3).

## Writer's own calls
- D-1: the handle scope is its own `HandleScope` path, not the existing `Handles` recompute set.
- D-5 / D-7: lists return both the revision and its RFC3339 time (`updated_revision`, `updated_at`), not the time alone.
- D-6: the revision is filled on the returned list rows, never stored in the display cache; only `List*` fill it.
- D-8: Swift stops at the regenerated protobuf file; Swift model fields are S10-22.
- D-9: the backfill maps old Observation history through today's membership.
- D-10: `TestSourceScopes`' seed moves to file level and is shared with the new test, so this PR does refactor that test file.
- M-8: the stale Swift list after a note-only save is given to S10-22 (a line added to its map **In**: report `effects.HandleScopes` on `WriteEffects` and invalidate the lists), rather than widening this PR into `writes.Run` and the FFI write effects. This adds engine work to S10-22.

## Round 1

### M-1 · major · T-2 copies `TestSourceScopes`' seed instead of sharing it
`seed` / `mustSeed` are closures inside `TestSourceScopes`; a sibling file would copy ~100 lines. Lift to a file-level seed or add rows to the existing test.
**Answer:** Fixed — D-10; STEP-3 moves the seed to `scopeSeed` / `newScopeSeed` with named methods, and `TestHandleScopes` lives in `scopes_test.go`. Codebase fit rows updated.

### M-2 · major · T-1 adds a fourth effects fixture
Every T-1 case already has a fixture in `TestConclusionHandleSets` or `TestHandlePaths`.
**Answer:** Fixed — T-1 now extends the named subtests and adds three subtests to `TestHandlePaths`, with `mustHandleScopes` beside `mustHandles`. `handle_scope_test.go` is gone.

### M-3 · minor · `HandleScope` repeats the `Handles` paths
**Answer:** Fixed — D-2 names `observationMembers` and `valueMembers(col)`; both `Handles` and `HandleScope` use them. Stop condition added if any existing handle test result changes.

### M-4 · minor · Six scalar fields where one typed field would do
**Answer:** Fixed — D-6: one `Updated audit.Latest` per header, filled through `latestHandles`. Proto stays two scalars (D-7).

### M-5 · minor · Scope types and effect paths are hand lists in several places
**Answer:** Fixed — D-4 `scopeResolvers` table read by `Record` and (via `export_test.go`) `checkNewEffects`; D-11 `Effect.paths()` used by `validate` and `hasWork`.

### M-6 · minor · Handoff to S10-41 is only a "can"
**Answer:** Fixed — STEP-8 adds the `LatestByScope(ScopeSource)` line to the S10-41 map entry and makes the skill's Read section name the list reader. Out line updated.

### M-7 · nit · `LatestByScope` signature and file
**Answer:** Fixed — D-5: named `audit.Querier`, reader in `scopes.go`; no `latest.go`.

### X-1 · blocker · T-2 and T-3 have no seed to run on
**Answer:** Fixed — same as M-1. STEP-3 names every helper, including `promote` and `handleRevision`; T-2 lists its setup per case using the seed's existing Subjects.

### X-2 · blocker · The zero-scope handle cannot be built through `writes`
`canonical_entities` create is audited, so `self()` scopes every new handle.
**Answer:** Fixed — AC-1.3 now says so and names the raw `INSERT` (columns, id, ref prefix) used in T-3 and T-4.

### X-3 · blocker · Migration body is a sketch and the backfill is never checked
**Answer:** Fixed — §6 Schema is the complete SQL, to be written verbatim (D-9). §10 step 2 adds two queries on a copied catalog and a stop condition when the first is not `0`.

### X-4 · major · Query-count check cannot see the new read
**Answer:** Fixed — STEP-5's check now says those tests count graph queries and are unchanged by `LatestByScope`.

### X-5 · major · `hasWork` is not updated
**Answer:** Fixed — D-11; STEP-1 lists `registry_completeness_test.go`.

### X-6 · major · Doc comments left to the implementer
**Answer:** Fixed — §6 Go core gives the `Record`, `Scope`, `ScopeCanonicalEntity`, and `HandleScopes` comment wording; STEP-3 gives the `seenEffectsRevisions` comment.

### X-7 · major · Test setup choices open in T-1, T-4, T-5
**Answer:** Fixed — T-1 names the subtests and fixture variables; T-4 names `f.person`, `f.event`, `f.place`, `f.book`, `f.list` and the note write (`observations.Update` with `Notes`); T-5 extends each test's existing promoted-rows case.

### X-8 · minor · `LatestByScope` nil and error handling
**Answer:** Fixed — D-5: nil `q` or empty type → `ErrInvalid`; query, scan, `rows.Err` returned; D-6: `c.DB()` error returned.

### X-9 · minor · Plan not yet reviewed
**Answer:** Declined — this is the review; status stays Draft until it passes.

### X-10 · minor · Temptations
**Answer:** Fixed — Out lines added for `effects.Set` / `Resolve` and the `000036.sql` comment; the `sources` switch is now S10-41's.

### X-11 · minor · Generator prerequisites
**Answer:** Fixed — stop condition for a missing generator or a diff beyond the new fields; STEP-6 check says the same.

## Round 2

Both reviewers confirmed every round-1 answer against the code (no import cycle; `export_test.go` and per-package `Querier` are house patterns; `000036.sql` already uses `unhex`).

### M-8 · blocker · A note-only save moves the scope but the app's lists never reload
`observation_notes` has no `Handles`, so `WriteEffects.touchedHandles` is empty and `CatalogQueryRegistry.invalidations(for:project:)` keeps the cached lists. The Out line marked this "not planned".
**Answer:** Fixed — the Out row's home is now S10-22, and STEP-8 adds the line to S10-22's map **In** (report `effects.HandleScopes` ids on `WriteEffects`, invalidate the three lists). Listed in Writer's own calls.

### X-12 · blocker · T-1 note subtest names an Observation on alice that does not exist
**Answer:** Fixed — T-1 uses Boston's date Observation `created.Observations[0]`, expects `{bostonRes.Entity.ID}`, and gives the change literal. The subject-move subtest builds on `observationSubject`.

### X-13 · major · T-4 raw insert versus a loaded graph; map Testable line contradicts AC-1.3
**Answer:** Fixed — T-4 inserts the raw handle first, before any list or promote. STEP-8 rewrites the S10-21 Testable line.

### X-14 · major · Seed move leaves `userID` and `validLocator` unresolved
**Answer:** Fixed — STEP-3 names `scopeUserID`, `scopeLocator`, and the `run` signature.

### X-15 · minor · `paths()` does not cover `Search` / `Vocabulary`; "five paths" wording
**Answer:** Fixed — Current state row and D-11 corrected.

### X-16 · minor · Ada's handle name; T-2 inputs; date property
**Answer:** Fixed — AC-3.3 and T-1 name `adaPromoted.Entity.ID`; T-2 says to copy the update inputs and use `s.dateProp`.

### X-17 · minor · T-4 event inputs; T-5 accessors
**Answer:** Fixed — T-4 uses `TestListEvents`' `fire` inputs and `f.book("Gazetteer")`. T-5 accessors are generated; no change.

### X-18 · minor · Map status line, map link, skill "Do not" line
**Answer:** Fixed — STEP-8 and the DoD name the slice 6 sequence line and the skill sentence. The map link is added with this plan.
