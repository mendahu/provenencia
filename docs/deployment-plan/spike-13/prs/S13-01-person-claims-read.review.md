# Review — S13-01 Person claims read

**Verdict:** Fits with changes — Ready with changes. Every finding Fixed; no blocker open.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- D-4: return the Source type as an id? Asked before review. Answered 2026-10-31: return the id.
- M-10: move the older `.evidenceGraph`, list and `.conclusionDetail` fan-outs onto `refreshes` in S13-01, as a preparatory PR, or not at all? Recommended: in S13-01. Answered 2026-10-31: in S13-01.

## Writer's own calls

- D-6: claims inside a group are ordered by graph name, then graph ref, then Subject ref.
- D-7: an unknown or merged entity reuses `conclusiondetails.not_found`; there is no kind check.
- D-8: a member whose graph or Subject is missing from the loaded graph is skipped, not an error.
- D-12: one `seen` set per `apply` covers invalidated keys and every revalidating refresh.
- D-12: refreshes that revalidate run synchronously inside `apply`, so moved keys keep today's timing. `.run` refreshes (progress) stay async.
- D-15: FakeStore treats any unseeded entity as not found.

## Round 1

### M-1 · blocker · `perHandle` is a second "revalidate the key this write names" mechanism beside `CatalogQueryRefresh.applyHiddenPolicy`
`CatalogLastUpdated.swift` `CatalogQueryRefresh.applyHiddenPolicy` already revalidates a whole warmed key whose id a write names, through `WorkspaceSession.applyHiddenPolicy(to:)`. The plan's reason for rejecting it ("refresh reloads part of a value") was false.
**Answer:** Fixed. D-12 widens `applyHiddenPolicy` to follow any `CatalogQueryFollows` and adds `.touchedHandles`. `.conclusionDetail` and `.personClaims` declare entries in `refreshes`. The `perHandle` struct is gone. `apply` dedupes refreshes against invalidated keys (AC-2.6, T-8). Checked: `WorkspaceSession.apply` L187 – 203 runs each warmed key's refresh with `follows.ids(in:mutation:)`.

### M-2 · blocker · `allOnEvidenceGraphs: Bool` is a per-type flag the next add must add a sibling to
**Answer:** Fixed. `refresh` becomes `refreshes: [CatalogQueryRefresh]`. `applyHiddenPolicy` without an `id` revalidates any warmed key of the kind when the slot names any id, and `CatalogQueryFollows.evidenceGraphs` is added. Moving the hand-written persons, events and places list fan-out onto the same form is logged as an Out line.

### M-3 · blocker · D-10 mapped Observations per claim, so up to one query and two map rebuilds per claim
**Answer:** Fixed. D-10 splits `listedProtosFromGraph` into `newListedContext` and `listedContext.rows`. `ListPersonClaims` calls `listedProtos` once per response. `Claim` carries its loaded graph, and `GraphSubjects` is dropped (D-11). T-6 asserts at most one query on a second call (AC-1.8).

### M-4 · minor · `graphsOfSource` walked twice in D-5
**Answer:** Fixed. D-5 puts the `union` inside `graphsOfSource`.

### M-5 · minor · Three names for the group
**Answer:** Fixed. One stem: `ClaimGroup` / `PersonClaimGroup` / `CatalogPersonClaimGroup` (D-9, D-14).

### M-6 · minor · `loadGraphSource` would grow to eight positional returns
**Answer:** Fixed. D-3 replaces it with `loadGraphHead`, which returns a struct.

### M-7 · minor · Give S13-27 the member→graph walk
**Answer:** Fixed. D-2 adds `memberGraphs`.

### M-8 · minor · Exported title order lives in a Conclusion-headers package
**Answer:** Fixed for now and logged. STEP-3's doc comment says it orders any titled row. Moving it is an Out line for when a third non-header caller appears.

### M-9 · minor · `invalidationsComeFromRegistrySpecs` does not stay unchanged
**Answer:** Fixed with X-1.

### X-1 · blocker · `invalidationsComeFromRegistrySpecs` and `handleScopesInvalidateNothing` change
**Answer:** Fixed. STEP-8 removes the `.conclusionDetail` per-handle entries, which now come from `refreshes`. STEP-9 adds `.allCached(.personClaims)` after `.allCached(.conclusionDetail)` in the `.deletedSource` and `.promotedSubject` lists. The stop condition now names only edits outside these.

### X-2 · major · `everyKindHasOneDescriptor` asserts 18 kinds
**Answer:** Fixed. STEP-9 changes it to 19.

### X-3 · blocker · `evidence_graphs.ref` is nullable
**Answer:** Fixed. D-3 uses `COALESCE(g.ref, '')`. AC-1.7 and T-5 cover a `SeedRefless` graph. D-6 says how an empty ref sorts.

### X-4 · blocker · No named way to build a provisional claim or a shared fixture
**Answer:** Fixed. D-16 names the public helpers. A provisional claim is `identityclaims.Create` with `StatusProvisional` inside `writes.Call`, as `identityclaims_test.go` `mustClaim` does. A merged handle is set through `merged_into_id`, as in `conclusiondetails_test.go` L444.

### X-5 · major · `SortByTitle` cannot express the claim order
**Answer:** Fixed. D-6 does a stable sort by Subject ref first, then `SortByTitle` on graph name and ref.

### X-6 · major · No import check for `persondetail`
**Answer:** Fixed. STEP-4's check runs `go list -deps`.

### X-7 · major · How claims are mapped
**Answer:** Fixed with M-3 (D-10, D-11). The `subjects.Subject` conversion follows `ListSubjects`, and an empty Observations list is sent as an empty list.

### X-8 · major · Where the Swift code goes
**Answer:** Fixed with M-1 and M-2. D-12 covers the `make` parameter, the follows cases and their sets, and the existing call sites. The persons, events and places list block stays hand-written (Out line).

### X-9 · minor · Smaller choices
**Answer:** Fixed:
- FakeStore unknown entity: D-15.
- `mapSubject` access: D-14.
- Generated files named: STEP-6.
- Reserved method numbers corrected in Current state.
- The Swift comment on `ConclusionListPage.swift` is added to STEP-3 and the sweep.
- `sortByTitle`'s doc comment: STEP-3 and Go core.
- How T-5 changes a Source's type: through `sources.Update`.
- T-4 covers only the Go read; T-6 covers the handler's query count.

## Round 2

The round 1 answers were checked against the code. All hold except M-2's Out line, which M-10 replaces.

### M-10 · blocker · Hand-written effect fan-outs stay beside `refreshes`
After round 1, `invalidations(for:project:)` still hand-wrote `evidenceGraphs` → `.evidenceGraph` and `touchedHandles` → the three lists, beside the new `refreshes`. The Floors do not allow logging a second mechanism.
**Answer:** To user, answered 2026-10-31: move them in S13-01. Fixed. D-13 declares them, STEP-9 deletes the blocks and moves their tests, and AC-2.7 and T-12 cover it. Revalidating refreshes are synchronous (D-12), so timing does not move.

### M-11 · blocker · `.personClaims` revalidated on any graph write although its value lists its graphs
**Answer:** Fixed. D-12's `.revalidate` action returns the ids a key depends on, read from the key or its cached value. `.personClaims` returns its claims' graph ids for `.evidenceGraphs`. `applyHiddenPolicyOnAny` stays only for the lists. T-9 adds the case of a graph the person has no claim on.

### M-12 · minor · Doc comments on `invalidations(for:project:)` and `WorkspaceSession.apply` name the old split
**Answer:** Fixed. STEP-9 rewrites both.

### X-10 · blocker · `ThrowingStore` in `OnboardingModelTests.swift` conforms to `GenealogyStore`
**Answer:** Fixed. It is in §4, D-14 and STEP-7.

### X-11 · major · Two refreshes of `.personClaims` could revalidate one key twice
**Answer:** Fixed. D-12 uses one `seen` set across invalidations and every revalidating refresh. T-8 now promotes with both slots set.

### X-12 · major · `descriptor.refresh?.follows` assertions at `CatalogQueryRegistryTests` L941 – 943
**Answer:** Fixed. STEP-8 changes them to `refreshes.map(\.follows)`.

### X-13 · major · `SeedRefless` closes the catalog
**Answer:** Fixed. T-5 sets the ref to NULL through `c.DB()` and calls `Graph().Drop()` instead.

### X-14 · major · The Go fixtures are under-specified
**Answer:** Fixed. D-16 names the precedents for each Observation kind, the rename and the promotes. `mustClaim`'s body is repeated in the test.

### X-15 · minor · Small choices
**Answer:** Fixed:
- The `ids` closure shape is in D-13.
- A bad id maps to not found, as `GetConclusionDetail` does (D-7).
- The Subject literal is repeated, not extracted (D-11).
- An empty result is a nil slice (D-7).
- STEP-4 adds `go build ./...`.

### X-16 · minor · The Mac-only checks
**Answer:** Fixed. §10 says which steps need macOS and leaves them to `scripts/check-pr.sh` on a Mac.
