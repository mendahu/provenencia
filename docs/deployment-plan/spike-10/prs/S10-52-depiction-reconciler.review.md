# Review — S10-52 Depiction auto-reconciler module

**Verdict:** Fits with changes — Ready with changes (all applied)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- The key is the image File id, not the checksum or the `depiction_values` id (D-1, D-4).
- The loader reads images through `depictionvalues.Readable`; a refused image is `no_evidence` (D-2, D-5, after M-1).
- The cache column has no FK, so it never keeps an image File alive (D-4).
- A shared test-fixture package `observationstest`, with `observations_test.go` moved onto it (D-10, after M-2).
- `depictionvalues.Joined`, with `observations.go` moved onto it (D-5, after M-9). It edits two files S10-54's branch also edits; §11 says how to rebase.
- Planned against S10-54's plan, not its branch; no code dependency between the two.

## Round 1

### X-1 · blocker · STEP-2 adds a T-3 check that needs STEP-3's field
**Answer:** Fixed — STEP-2 adds T-3 without the node-values comparison and the "Verify sees a stale image" subtest; STEP-3 adds both.

### X-2 · major · T-6 copies raw rows from a test STEP-2 deletes
**Answer:** Fixed — T-6 saves through `observationstest.SaveCitation` (M-2); no raw rows.

### X-3 · major · New test helpers leave choices unmade
**Answer:** Fixed — the helpers live in `observationstest` (D-10) with file names, signatures, and imports given; `ingest.File` reuses a File by checksum (`core/ingest/file.go` `File`), stated in STEP-2.

### X-4 · minor · T-1 rows do not match `TestModules`' columns
**Answer:** Fixed — T-1 gives each row in the struct's column order.

### X-5 · minor · D-5 scan placement
**Answer:** Fixed — D-5 names the four variables, their types, and their place in the `Scan` list.

### X-6 · minor · D-11 has no check
**Answer:** Fixed — `ls core/propertymeta` on the `main` you branch from.

### X-7 · minor · STEP-5 doc edits ambiguous
**Answer:** Fixed — "replace the status after the link"; `completed.md` copies the S10-32 section.

### X-8 · minor · Behavior the dry run could not confirm
**Answer:** Declined — no change needed; the stop conditions cover AC-1.3, and `TestLoaderQueryCountIsConstant` is a STEP-2 check.

### M-1 · major · A refused image is dropped instead of being `no_evidence`
**Answer:** Fixed — `Value.HasDepiction` (D-1); `Carries` reads it, the key reads `ImageFileID` (D-2); AC-1.7 expects `no_evidence`; T-1 and T-2 gain unreadable-image rows.

### M-2 · major · Likeness test fixtures copied a second time
**Answer:** Fixed — `core/database/observations/observationstest` (D-10), in the `namevaluestest` shape; `observations_test.go`, T-3, T-4, and T-6 use it.

### M-3 · major · Spike 11 docs name `hasValue` as open work
**Answer:** Fixed — STEP-5 edits the S11-06 row and entry and the `comparison-layers.md` duplicates row; a DoD line checks it.

### M-4 · minor · `imageKey` copies the subject closure
**Answer:** Fixed — `idKey(id)` shared by subject and depiction (D-1).

### M-5 · minor · "Next adds" undercounts a new value type's sites
**Answer:** Fixed — Next adds lists every site. The `graphcache.Value` / `TruthRow` mirror is logged here and in Out as not planned: no PR in either map touches it next.

### M-6 · minor · D-9's skip has no owner
**Answer:** Fixed — Next adds: S10-35 replaces the skip with the Person detail's gallery list.

### M-7 · nit · `knownValueType` becomes a forwarding wrapper
**Answer:** Fixed — deleted; `Reconcile` and `Compatible` call `moduleFor` once (D-2).

## Round 2

### M-8 · blocker · `autoreconcile.Value` holds a depiction as two loose fields
**Answer:** Fixed — `Value.Depiction *depictionvalues.Value`, the shape of `Date`, `Name`, and `observations.Listed.Depiction` (D-1); `Carries` is `Depiction != nil`, the key reads `Depiction.ImageFileID` (D-2); `graphcache` keeps a plain `[]byte` (D-8). `HasDepiction` is gone.

### M-9 · blocker · D-5 copies the joined-depiction assembly from `observations.go`
**Answer:** Fixed — `depictionvalues.Joined` (D-5) with `TestJoined` (T-7); `observations.go` and the loader both call it; the loader selects `o.value_depiction_id` for its id.

### X-9 · minor · AC-4.2's `no_evidence` outcome untested
**Answer:** Fixed — T-1 asserts `Outcomes[0].Reason == ReasonNoEvidence`.

### X-10 · minor · T-4 region and write shape
**Answer:** Fixed — T-4 gives the region literal and the `writes.Call` / `observations.Listed` shape.

### X-11 · minor · "no other change" vs imports
**Answer:** Fixed — D-10 allows the import list and D-5's move; the stop condition matches.

### X-12 · minor · `TestLoaderQueryCountIsConstant` is not evidence for D-5
**Answer:** Fixed — removed from the stop conditions; §8 says why it stays only as a must-stay-green suite.

### X-13 · nit · §7 doc placement
**Answer:** Fixed — after the paragraph that starts "The same *same value* test".

### X-14 · nit · AC-1.1 side assertions
**Answer:** Fixed — T-3 names the nil fields and the `kept` outcomes.
