# Review — S11-44 Des as a nickname of Desiree and Desirae

**Verdict:** Fits with changes — Ready with changes (round 1; round 2 after the To user answer)
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-1.** The terminal command opens the catalog with `database.Open` outside the app, and treats `catalog.already_open` as its normal "close the project first" refusal. `use-catalog-session` and `.cursor/rules/catalog-session.mdc` call that a bypass bug. Recommendation: accept, and add one scoped line to each (STEP-6). The command must open out of process because the user asked for a terminal command (2026-10-27).

## Writer's own calls

- The command is a committed Go binary (`cmd/add-product-vocabulary`), not a throwaway script.
- It runs only on a closed project; a busy or locked error from `Open` is mapped to `already_open`.
- New bindings are appended after the researcher's own, not placed at registry order.
- No audit rows, as `Install` writes none.
- Seeded types and Source vocabulary are left out: they can be deleted, so "missing" is not "new".
- The command is run from the commit the installed app was built from (it migrates).

## Round 1

### M-1 · blocker · "Rules this plan bends: None" leaves out two guidance files the plan contradicts
`use-catalog-session` ("`database.Open` — Migrate-only"; "already_open … a bug") and `.cursor/rules/catalog-session.mdc`; `.cursor/rules/seeded-vocabulary.mdc` ("Install only in createCatalog … no vocab heal").
**Answer:** To user for the session rule. The seeded-vocabulary rule is answered by the user's decision 3 (2026-10-24) and the terminal command (2026-10-27). Both rows are now in **Rules this plan bends**; STEP-6 updates the skill and both rule files.

### M-2 · blocker · `AddMissing` converts registry rows into DB rows a second time, beside `Install`
**Answer:** Fixed — D-6a: `seedProperty.row` and `seedTerm.row` in `registry.go`, called by `Install` and `AddMissing`.

### M-3 · major · D-7 adds a local querier interface
**Answer:** Fixed — D-7: `UpsertTx` is the body and `Upsert` wraps it (Begin, `UpsertTx`, Commit), as `namevalues.Insert` wraps `InsertTx`. No interface.

### M-4 · major · `assertMatchesRegistry` checks binding order `AddMissing` does not keep
**Answer:** Fixed — D-8: the shared helper checks rows, the binding set and `Locked`; `TestInstallMatchesRegistry` keeps its `SortOrder` check.

### M-5 · major · "Next adds" promises a slot the shape does not give
**Answer:** Fixed — D-6a keys bindings on one added set; Next adds and the Out row say deletable kinds need a "new since create" signal, logged in `docs/ideas/organizations.md` (STEP-6).

### M-6 · major · A domain doc line becomes false, and Reach misses its case
**Answer:** Fixed — STEP-6 rewrites the `likeness` row in `seeded-vocabulary.md` §3.2; Reach has a row for the appended locked binding; T-3 (e) pins it.

### M-7 · minor · The binary sits under `api/` although the stack doc names `cmd/`
**Answer:** Fixed — D-3: the binary goes in the top-level cmd folder, `check-pr.sh` gains `./cmd/...`, and `application-stack.md` §26 lists the command.

### M-8 · nit · Parallel count types have unparallel names
**Answer:** Fixed — `subjectvocab.SeedAdd`.

### X-1 · major · The T-5 and T-3 setup SQL is left to the implementer
**Answer:** Fixed — T-3 and T-5 give the exact statements (a Property delete cascades to its terms and bindings; Des sets are deleted whole by `set_idx`).

### X-2 · major · D-7 leaves the shared-helper shape open
**Answer:** Fixed by M-3.

### X-3 · minor · How `SeedAdd.Bindings` is counted
**Answer:** Fixed — D-6a: only when `RowsAffected` is 1.

### X-4 · minor · AC-2.5 may surface an unmapped lock error
**Answer:** Fixed — D-4 maps `IsBusyOrLocked` errors from `Open` to `ErrAlreadyOpen`; a new Stop condition covers §10 step 4.

### X-5 · minor · STEP-2's skill is not in the header
**Answer:** Fixed — `add-catalog-query` and `use-catalog-session` added to Skills to load.

### X-6 · minor · Scope and STEP-7 disagree about the map
**Answer:** Fixed — STEP-7 says the scope entry was set at planning; only the row status changes.

### X-7 · minor · The `Invalidate` sentence in `add.go`
**Answer:** Fixed — STEP-6 says it stays (true for in-process callers).

### X-8 · nit · T-1 harness
**Answer:** Fixed — T-1 builds rules as `TestInstallAndLoad` and scores with `assertScore`.

### X-9 · nit · DoD check-pr off macOS
**Answer:** Fixed — the DoD box says what to report off macOS.
