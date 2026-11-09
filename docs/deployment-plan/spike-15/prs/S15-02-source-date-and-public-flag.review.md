# Review — S15-02 Sources carry a date and a not-public flag

**Verdict:** Round 1: Fits with changes — Not ready. Round 2: pending.
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **M-6 · A second, older date sort key.** `core/autoreconcile/autoreconcile.go` `SortKey` stores a key for dates in `auto_reconciler_values.sort_key`. It is built from `DateWindow` with no before/after tie-break, so it disagrees with the one date order this PR adopts. No query was found that orders by it. Recommendation: keep it out of S15-02 (it is not Source work), and run it as a separate chore. That chore either moves the date case onto `valuecompare.DateKeyOf` with an autoreconciler stamp bump, or deletes the column if it truly has no reader.

## Writer's own calls

- **Two RPCs, `SetSourceDate` and `SetSourcePublic`, not new fields on `UpdateSourceRequest` (D-3).** An update request sends every field of the Source, and a protobuf bool left unset reads as false. Any caller that omitted the new field would mark the Source not public.
- **Batched date read for the list instead of a SQL join (D-6, from M-2).** `LookupManyTx` is the existing batched read with three callers; a join needed a second 20-column scanner. The map's S15-02 scope entry now says "one batched lookup, never one query per row" instead of "joins".
- **Wire `is_public` keeps the column's name (D-5).** Every Source on the wire is built in one place (`sourceProto`), which always sets it.
- **The integer sort key lives in `sources.DateSortKey` (D-13).** `valuecompare` holds the order (`DateKey`, `Compare`) but no integer encoding, because that package is guarded against numeric literals.
- **A date with no year (phrase only) is stored and returned but has no sort key** (AC-3.4). It sorts with the undated Sources.
- **Copy Source carrying the date** is S15-32, gated by a new brief S15-D13 (user, 2026-11-09: a checkbox on the Copy form).

## Round 1

### X-1 · blocker · `DateKey.Int64` breaks `TestLayerPackages` (numbers layer)
`core/matchregistry/guard_test.go` `checkNumbers` refuses the literal `3` in `core/valuecompare`.
**Answer:** Fixed. D-4 drops `Int64`. D-13 puts the encoding `Day*3 + Side + 1` in `sources.DateSortKey`, outside the guarded layer. T-1 tests `Compare`; the new T-10 tests the integer order.

### X-2 · major · `dateKey` / `placedDate` fields after D-4 unspecified
**Answer:** Fixed. D-4 names both structs' fields (`at valuecompare.DateKey`), the literal rewrites in `eventDating` and `dateKeyOf`, and `dateKeyLess`'s order (ok, then `Compare`, then ref).

### X-3 · major · The audited delete when a Source date is cleared is under-specified
**Answer:** Fixed through M-1. D-7 exports `deleteimpact.ReleaseReplacedRow` with its full signature. D-8 names the `rowMap` (`LookupTx`, then `RowMap`).

### X-4 · major · `SetDate` value→value: how equality is decided and what it returns
**Answer:** Fixed. `datevalues.PutTx` (D-7) decides equality with `UpdateChange` and skips the write when nothing changed. D-8 returns the Source re-read with `getTx`, so `Date.ID` is set. Trimming stays where it is today, in `InsertTx` and `UpdateTx`.

### X-5 · major · Edge cases left to the implementer
**Answer:** Fixed. D-3 says a blank `kind` clears the date whatever else the request holds, the same test the Observation handler uses. T-1 names the `datevaluetest` helpers and spells out the two fixtures that have no helper. D-9 says the clear's Source scope comes from the `source` update change.

### X-6 · minor · `completed.md` path ambiguous
**Answer:** Fixed. STEP-8 and the DoD give the full spike paths.

### X-7 · minor · Checks not runnable in the review container
**Answer:** Declined. Not a plan defect: the implementer runs `generate-proto.sh`, Go 1.27 and `check-pr.sh` on macOS, as every PR in this repo does.

### X-8 · minor · Scan changes not spelled out
**Answer:** Fixed. D-6 names the two new columns in `scanSource` / `scanSourceList`, the `is_public` int scan, and where `List` runs its batched lookup.

### M-1 · blocker · `sources.SetDate` would be a second copy of the Observation's owned-date write
**Answer:** Fixed. D-7 moves the whole job: `datevalues.PutTx` (insert, or update in place, with its change) and `deleteimpact.ReleaseReplacedRow` (today's `releaseReplacedValue`, exported). The Observation date case, the Observation release loop and `sources.SetDate` all call them. Checked that neither package imports the other; a Stop condition covers it.

### M-2 · blocker · `datevalues.Joined` adds a second scan beside `scanValue`
**Answer:** Fixed. `Joined` and `JoinColumns` are dropped. `sources.List` uses the existing `LookupManyTx`: 6 queries at 3,000 dated Sources, never one per row. The map scope entry was reworded in the same commit. The Current state row that wrongly listed the `deleteimpact` projector as a scan is removed, and so is the Out line.

### M-3 · major · Docs and skills the PR makes false
**Answer:** Fixed. STEP-8 and the DoD add `structured-date-model.md` § 4 and § 5, `source-layer-data-model.md` § 10, `add-date-value`, `date-values.mdc` and `add-audit-scope`.

### M-4 · major · The owned-date index is plain
**Answer:** Fixed. D-1 and Schema use `CREATE UNIQUE INDEX sources_date_value_id_uidx … WHERE date_value_id IS NOT NULL`, the `000030.sql` shape.

### M-5 · major · `dateCreateChange` re-spells `dateValueMap`; `intPtrAny` goes dead
**Answer:** Fixed. D-7 builds `CreateChange` as `FullRow(RowMap(…))`. Deletes and STEP-2's `rg` check include `intPtrAny`.

### M-6 · major · `autoreconcile.SortKey` is a second date sort key
**Answer:** To user (above). Codebase fit and Current state now list it, and Out records it.

### M-7 · minor · Hand-built `date_values` effect
**Answer:** Fixed. D-9 gives `ownedValue` a variadic `owners ...Path` parameter. The registry line stays one call.

## Round 2

Executability (Sonnet): no blocker; every round-1 answer checked against the code and holds. Verdict Ready with changes.

### X-9 · minor · D-4's `dateKeyOf` literal is not `{at: p.at, ok: true}`
**Answer:** Fixed — D-4 names the two `eventDating` literals and `dateKeyOf`'s own (`ok: ok`, `ref`) separately.

### X-10 · minor · D-8's `date_value_id` change encoding
**Answer:** Fixed — D-8 writes both sides with `database.UUIDJSON`, as `SetCover` does, so the unset side is JSON null.
