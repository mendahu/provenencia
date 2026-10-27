# Review — S11-44 Des as a nickname of Desiree and Desirae

**Verdict:** Round 1: Fits with changes — Ready with changes. Descoped by the user's answer to M-1. Round 2: Fits — Ready
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet); round 2 one reviewer on both lenses (sonnet)

## To user

- **M-1 (answered 2026-10-27).** Asked whether the session rule could carry an exception for an out-of-process command. The user's answer: "I will remember to close the app before I run it. We don't need to build anything for this. It's a one-off command." The plan was descoped to the seed edit: no command, FFI or menu ships, and existing catalogs get the sets and S11-48's Properties from a one-off the user runs by hand. No rule is bent.

## Writer's own calls

- The one-off command is given to the user in the planning session, not committed. It was tested on a stripped catalog in the session: the first run added 2 sets (4 names), 2 Properties, 2 terms and 2 bindings; the second run added nothing; a relabelled seeded Property kept its label.
- S11-48 is no longer a dependency of S11-44 in the map (only the one-off needs it).

## Round 1

The round-1 fixes below were applied, then made moot by the descope: the command, `AddMissing`, the `UpsertTx` variants and their tests left the plan. X-9 (DoD off macOS) still applies.


### M-1 · blocker · "Rules this plan bends: None" leaves out two guidance files the plan contradicts
`use-catalog-session` ("`database.Open` — Migrate-only"; "already_open … a bug") and `.cursor/rules/catalog-session.mdc`; `.cursor/rules/seeded-vocabulary.mdc` ("Install only in createCatalog … no vocab heal").
**Answer:** To user → answered: descoped (see To user). The seeded-vocabulary rule is answered by the user's decision 3 (2026-10-24) and the terminal command (2026-10-27). Both rows are now in **Rules this plan bends**; STEP-6 updates the skill and both rule files.

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

## Round 2

No blockers or majors. The reviewer checked the seed neighbours, the counts (1227 → 1229; 108, 35, 143 unchanged), `CacheVersion` 18, `productAliasPin`, the fold table's row shape, the T-1 helpers, and that `core` is unchanged since `5312337`. Verdict: Fits — Ready.
