# Review — S16-15 An Evidence-graph catalog in the graph cache

**Verdict:** Fits — Ready with changes (round 2 fixes applied)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user (answered 2026-11-12: both as recommended)
- M-1: D-5 drops `List`'s read transaction against `add-catalog-query` "Reads inside a write". Recommendation: accept and rewrite the skill paragraph (registry loaders on the session graph run under `catalogsession.Do`'s lock; a direct `c.DB()` multi-query read still opens a transaction).
- M-2: S16-13's graph order. Recommendation: the Person page keeps `conclusionsummaries.SortByTitle`; S16-13 reads only graph shells from the index; the map and design text for S16-13 are corrected.

## Writer's own calls
- `graphprogress` folds into `evidencegraphs` and the package is deleted (D-1); the map said "`graphprogress` reads it".
- The index holds the whole graph row (Description, ArtifactIDs) so `List` reads it instead of keeping its SQL (D-3).
- The project-wide graph list is ordered by Source id, then each Source's list order, instead of by name across Sources (D-4). No client reads the cross-Source order.
- `SourceOf` stays on its two write paths; it has no read-path caller (Out).
- Budgets: Build ≤ 150 ms and 8 MB at the Scale; recounting a 10,000-Observation graph ≤ 20 ms; re-listing a 10,000-graph Source ≤ 50 ms; no cap.

## Round 1

### M-1 · blocker · D-5 removes the read transaction `add-catalog-query` uses as its example
**Answer:** To user → accepted 2026-11-12. Rules this plan bends has the row; STEP-7 rewrites the skill paragraph.

### M-2 · major · S16-13 told to order a person's graphs from `OfSource`, but the Person page uses `SortByTitle`
**Answer:** Fixed in the plan text (Blocks, Next adds, D-3 and D-4 Rejected now cite `SortByTitle`); To user → 2026-11-12: S16-13 keeps `SortByTitle`; the map (S16-13 row and scope entry) and the design doc say so.

### M-3 · major · effects README documents no Go reader of `Set.Reads`
**Answer:** Fixed — STEP-7 adds the README sentence, corrects "Evidence graphs" to "key ids", and corrects the map's Spike 15 note (both reads named).

### M-4 · minor · AC-3.1's sweep hits three review-lens files
**Answer:** Fixed — STEP-7 names the three files and the replacement wording; Scope's Docs row covers them.

### M-5 · minor · `IndexOf`/`ErrIndexOnWrite` copy `RecordsOf`/`ErrRecordsOnWrite`
**Answer:** Declined — two callers; a shared `graphcache` guard is due with the third whole `LifeGraph` entry. D-2's Rejected column says so.

### M-6 · minor · "Catalog" collides with `database.Catalog` and `LifeCatalog`
**Answer:** Fixed — `evidencegraphs.Index`, `GraphIndex`, `IndexOf`, `ErrIndexOnWrite`, entry `evidencegraphs.index`, files `index.go`, `index_test.go`, `index_bench_test.go`.

### M-7 · minor · The precedent claims one loader over an id list
**Answer:** Fixed — the Precedent names both loaders; D-5 says why `list` stays per Source (one Source per commit today, two for S15-24).

### M-8 · nit · `AllProgress`' non-zero filter is dead
**Answer:** Fixed — D-3 states the invariant instead.

### M-9 · nit · `database.Querier` vs `graphcache.Querier`
**Answer:** Fixed — D-5 says `database.Querier`, the one this file already uses.

### X-1 · blocker · The AC-3.1 sweep cannot pass
**Answer:** Fixed — same as M-4.

### X-2 · major · T-1's fixture path
**Answer:** Fixed — T-1 names `evidencelayertest.New(t)` at `core/database/evidencelayer/evidencelayertest`, `r.NewSource`, `evidencegraphstest.Create`.

### X-3 · major · T-2's write helpers and probe
**Answer:** Fixed — T-2 names the function for each row and copies `derivedFromRecords` and `check` from `TestRecordsApply`.

### X-4 · major · Source id order
**Answer:** Fixed — AC-2.4, T-1 and T-6 derive the expected order from `bytes.Compare` of the ids.

### X-5 · minor · `protoc` for a comment change
**Answer:** Fixed — the proto is not touched; its comment stays true.

### X-6 · minor · Gaps in the steps
**Answer:** Fixed — STEP-1 gives the `BenchmarkCacheSizes` command; STEP-1 states the moved tests use no unexported names (checked); STEP-6 lists each D-11 file in the report. The transaction question is M-1.

### X-7 · minor · Apply edges
**Answer:** Fixed — D-7 (5): a named graph that is not in `graphs` gets no row; an unknown Source leaves no `bySource` key (already stated).

## Round 2

### X-8 · blocker · `r.Person` reads `SoleGraph`, so it fails once a Source has several graphs
**Answer:** Fixed — T-1 builds Ada before the extra graphs; T-2 adds its Observation with `r.Cite` on Ada and its Subject with `subjectstest.On` on Smith.

### X-9 · major · Test package, second Artifact, `r.Run`, the `check` closure
**Answer:** Fixed — T-1 names package `evidencegraphs_test` and the second Artifact (`artifactstest.Create`, "Page 2"); T-2 runs every `tx` call in `r.Run`, and inlines `check` and a package-level `derivedFromIndex` probe.

### Round 2 maintainability
No new blockers; every round-1 answer holds. Verdict: Fits.
