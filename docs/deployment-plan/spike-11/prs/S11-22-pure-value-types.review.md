# Review — S11-22 Pure value types

**Verdict:** Fits with changes — Ready with changes, then fixed. **Rounds:** 1. **Reviewers:** maintainability (fresh agent), executability (fresh agent).

## To user

None. (One map edit made on the reviewer's evidence: S11-23 now also gives `OriginProvenencia` / `OriginUser` one pure home, see M-1.)

## Writer's own calls

- Package names `core/datevalue`, `core/namevalue`, `core/propertymeta`: singular, so they never clash with the database packages (D-1).
- The database packages keep their names and queries (D-2). `Validate` moves with the types (D-3).
- Constants stay untyped strings (D-5). S11-02 decides whether to type them.

## Round 1

### M-1 · major · Origin constants are not property metadata
`OriginProvenencia` / `OriginUser` are declared in nine packages. The only users of `properties.Origin*` are database packages.
**Answer:** Fixed. They stay in `core/database/properties`, so `propertymeta` holds the eight value-type and cardinality constants. S11-23's map row and scope now give Origin one pure home.

### M-2 · major · `add-catalog-query` would contradict D-4
**Answer:** Fixed. STEP-8 rewrites L40 for the pure-type / table-query split and adds a "Do not" line.

### M-3 · minor · One home for `PropertyMeta`
**Answer:** Fixed. Next adds says S11-09 adds `propertymeta.Meta`. STEP-9 updates `comparison-layers.md` §Layer 2.

### M-4 · minor · Two `Validate` shapes
**Answer:** Fixed. Both are `Validate(v Value) error` (D-3).

### M-5 · minor · Purity guard handoff
**Answer:** Fixed. STEP-9 adds the three packages to S11-02's guard list in the map.

### X-1 · major · T-1 cites a case that doesn't exist
**Answer:** Fixed. AC-2.2 and T-1 cite `About 1882-04` and `Between 1882 and 1882-04-03` from `TestCompactDisplay`.

### X-2 · major · T-4 would be an import cycle
**Answer:** Fixed. D-7 and T-4 set package clauses: `namevalue/validate_test.go` is `package namevalue_test`, `display_test.go` stays external, `registry_test.go` stays internal.

### X-3 · minor · Constant count and lines
**Answer:** Fixed. The const block is L37–51, and eight constants move.

### X-4 · minor · Searches don't return the counted sets
**Answer:** Fixed. §4 and STEP-2/4/6 count by import path, exclude the proto and `apperr` code strings, and name the four files whose `properties` import survives. `api/ffi/handlers/subject_defs.go` doesn't change.

### X-5 · minor · Stale references not covered
**Answer:** Fixed. STEP-8 adds both `.cursor/rules/*.mdc` files, the three Swift doc comments (In) and the `core/match/profiles.go` comment.

### X-6 · minor · Ambiguous base
**Answer:** Fixed. Planned against `origin/main` @ `af50bd6`.

### X-7 · minor · Checks that can't run on Linux
**Answer:** Fixed. On Linux, `check-pr: PASS` with `SKIP xcode-test` is accepted, because only Swift comments change. The app manual check is dropped, since there is no behavior change.

### X-8 · nit · Lint warnings
**Answer:** Fixed. Full paths are used, and the map row links the plan.

## Amendment — 2026-10-17 (ship loop, base check)

The implementer stopped on the base check at `origin/main` @ `233e87ad`: `core/database/properties/properties.go` gained `ValueTypeDepiction` and `researcherCreatable` after the planning base `af50bd6`. **User decision:** `ValueTypeDepiction` moves to `core/propertymeta` with the other value types (nine constants in all); `researcherCreatable` stays in `core/database/properties` and reads `propertymeta`. §3, §4, D-5, §6 Go core, STEP-1 and STEP-2 updated; line numbers refreshed to `233e87ad`. No other row changed (importer counts re-checked by the implementer).
