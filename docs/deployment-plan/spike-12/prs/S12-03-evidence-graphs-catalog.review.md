# Review — S12-03 Evidence graphs in the catalog

**Verdict:** Fits with changes — Ready with changes. Every finding fixed or logged; no blocker open.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None open. Three map working rules changed with this plan (D-11, D-12, D-14); they are listed under Writer's own calls for veto.

## Writer's own calls
- Backfilled Default graphs reuse the Source id; graphs made from now on get fresh ids, so a Source-for-graph mix-up fails in tests before S12-08 (D-2).
- `artifacts.CreateInput.GraphIDs` *empty* (not nil) means every graph, because proto3 cannot carry the difference. The map rule now says "empty" (D-11).
- The "Citations use this Artifact" refusal is a plain code with no params. The list of Citations comes from S12-05's preflight over `CitationsUsingArtifactTx`, shown with the DeleteImpact recipe. The map rule "Refusals are of two kinds" now says so (D-12, from M-1).
- The fixture rule is narrowed: only the `subjects` direct insert moves (`subjectpositions_test.go` becomes an external test package). The `sources` / `artifacts` inserts in `deleteimpact/release_test.go` and `derivatives/thumbnail_test.go` stay, because those internal packages cannot import the create functions (D-14).
- `catalogsession.ReconcileOnOpen` becomes the one list of open-time repairs that both open paths and `createCatalog` call (D-15, from M-3 and M-3a).
- Graph tables are audited facets; `subjects (graph_id, source_id) → evidence_graphs` is a resource edge with its probe added now; `KindEvidenceGraph` is not yet deletable (D-3, D-4).
- Error codes are user kind, per the working rule, though `*.in_use` codes elsewhere are conflict kind (D-17).
- No version bump.

## Round 1

### X-1 · blocker · STEP-2's check cannot pass, and the plan's caveat names the wrong packages
**Plan:** STEP-2 Check ran `catalogmodel`, `deleteimpact` and `effects`; it said other packages fail until STEP-3.
**Codebase:** the external tests in `effects` and `deleteimpact` call `subjects.Create`, which fails once `graph_id` is `NOT NULL`.
**Answer:** Fixed. Old STEP-2 and STEP-3 are now one step (STEP-2: schema, registries, Default graph, Subjects join it). Its check is the whole `./core/...` suite plus `go vet`. §4 gains a row naming those tests.

### X-2 · blocker · T-17 cannot live in `registry_completeness_test.go` as written
**Answer:** Fixed. It is split into T-18 (registry assertions, internal file) and T-19 (`Resolve` over real rows, `core/database/effects/evidence_graphs_test.go`, package `effects_test`). STEP-6 lists both files.

### X-3 · major · The STEP-3 check and the single-commit rule contradict each other
**Answer:** Fixed. A rule at the top of §7 says each step updates the count and list assertions it breaks, in that step. Every step ends green and commits. STEP-7 is the whole-suite pass.

### X-4 · major · Tests need graph ids, but `SoleGraphTx` takes a Tx and there is no read API
**Answer:** Fixed. D-20 says tests read graph and link rows with SQL through `c.DB()`, take ids inside a write from `SoleGraphTx(tx.Tx, …)`, and use a per-file `defaultGraphID` helper.

### X-5 · major · The rebuild precedent is mislabelled
**Answer:** Fixed. The §4 row and the Precedent bullet now say `000030.sql` rebuilds `artifacts` / `citations` / `property_terms` with children, that `000052.sql` has no `foreign_key_check`, and that `subjects` was never rebuilt. The `identity_claims` child (`000032.sql`) is named.

### X-6 · minor · D-8's retry-loop description
**Answer:** Fixed. D-8 says the loop is a copy of `subjects.InsertTx`'s, with its own `maxRefRetries` const and `database.IsUniqueConflict`.

### X-7 · minor · Open choices
**Answer:** Fixed.
- D-12 gives `Link`'s single existence query.
- D-11 says a Source with no graph gets no link and no error.
- D-14 says `mustPerson` keeps its signature, uses the enclosing `userID`, and lets `subjects.Create` mint its ref; no test reads the old `CPR-<label>` ref.
- D-6 says JSON null.
- D-20 gives the bad-input return values.

### X-8 · minor · Checks that cannot run in a Linux container
**Answer:** Declined, no change needed. §10 now says the manual check runs on the macOS dev machine, and `check-pr.sh` targets macOS by design. The migration SQL was run during planning (§4).

### M-1 · blocker · The "Citations use this Artifact" refusal builds a second way to report blocking rows
**Answer:** Fixed, in a different shape than suggested. `deleteimpact.Group` with projected titles needs `deleteimpact`'s unexported `projectListed`, and that projection belongs to S12-05's preflight. So in this PR:
- `Unlink` returns a plain `ErrArtifactInUse` with no params, like every delete refusal (`catalog-deletes.md` § FFI).
- `CitationsUsingArtifactTx` returns ids only, and `CitationRef` is gone.
- S12-05 shapes the list for the DeleteImpact recipe (Out row).

The map working rule is updated to match (D-12, D-13).

### M-2 · blocker · `artifacts.Create` resolves and validates graphs outside `evidencegraphs`
**Codebase:** `writes.Run` rolls back on error, so the stated reason for validating up front did not hold.
**Answer:** Fixed. `evidencegraphs.LinkArtifact` (D-11) owns the "empty means every graph" rule and the repeat check, and calls `Link`, which owns the same-Source rule. `artifacts` holds no graph SQL. AC-2.5 now expects `evidencegraphs.ErrInvalid` after rollback.

### M-3 · major · The open sequence is copied in two places
**Answer:** Fixed. `catalogsession.ReconcileOnOpen` (D-15) is called by `openResearcher` and `onboarding.OpenCatalog`; `onboarding` already imports `catalogsession`. T-16 covers the session path.

### M-4 · major · Skills the PR makes false are missing from STEP-9
**Answer:** Fixed. STEP-8 and the DoD now list `use-catalog-session`, `add-seeded-vocabulary` and `add-catalog-model` step 6 (the full `Effect` field list).

### M-5 · major · The new package copies the mint loop and `uuidJSON`
**Answer:** Logged. The two new sections are in `docs/ideas/shared-list-helpers.md` ("One ref mint loop"), and the plan has an Out row. The plan's wording now matches the code: `EnsureRefs` has `users.EnsureRefs`'s shape with no retry, and `insert` copies `subjects.InsertTx`'s loop.

### M-6 · minor · `EvidenceGraphs` is the third catalog-wide bool
**Answer:** Logged. It is in `docs/ideas/shared-list-helpers.md` ("Catalog-wide list flags") with an Out row. The working rule names the field, so this PR follows it.

## Round 2

Both reviewers confirmed the round-1 answers and raised no new blocker.

### X-9 · major · STEP-2 runs only `./core/...`, but `api` tests can break in that step
**Answer:** Fixed. The STEP-2 check runs `./api/... ./core/...`, and its Files line covers any `api/...` test that the §7 rule lets it update.

### X-10 · minor · STEP-3 and STEP-5 mark the STEP-2 files as new
**Answer:** Fixed. Both steps now say the files were created in STEP-2 and this step adds to them. They keep the `(new)` tag that the plan linter needs.

### X-11 · minor · T-17's second Artifact
**Answer:** Fixed. The case creates it with `runArtifactCreate` and takes the graph id from `SoleGraphTx` inside each write. `newScopeSeed` is unchanged.

### X-12 · minor · Fixtures for T-4, T-15 and T-16; evidence for AC-1.1
**Answer:** Fixed. T-4 gives the insert columns. T-15 and T-16 name their builders, their own dirs, and `catalogsession.Close`. The DoD says AC-1.1 is checked by hand on macOS, and an implementer without macOS reports it as not run.

### X-13 · minor · The map's link bullet still said "the error lists them"
**Answer:** Fixed. Map line 87 now reads "the refusal is a plain code; S12-05's preflight lists them".

### M-3a · major · `createCatalog` keeps a third copy of the repair list
**Answer:** Fixed. In D-15, `createCatalog` runs `database.Create`, then its six `Install` calls, then `ReconcileOnOpen`. None of the installers reads `users` or `project`. The §4 row, Codebase fit and Deletes are updated.

### K-1 · minor · The map scope entry and `source-structure.md` contradict the changed rules
**Answer:** Fixed. The S12-03 scope entry now says "empty = every graph", names the narrowed fixture move and `ReconcileOnOpen`. `source-structure.md` § Rules no longer says the message lists the Citations. `deployment-plan.review.md` is an old log and is left as it is.

### Below the bar, taken anyway
- One per-Source graph-id read: D-9 adds an unexported `graphIDsTx`, which `SoleGraphTx` and `LinkArtifact` share.
- `Link` and `Unlink` check membership through `IsLinkedTx` (D-12).
