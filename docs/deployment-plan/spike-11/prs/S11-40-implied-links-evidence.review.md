# Review — S11-40 Implied links in claim evidence

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- One bundle per Source recording of an implied link, the member's bridge first and keyed by it; a Source with no member of the row's entity on the link's bridges gives no bundle (D-4, D-6).
- A chain step implied through a birth adds no participations to the chain's bundle (D-7).
- A side that records a fact both directly and through a pattern reads as direct (D-9, `Recording.Shown`).
- The tree pivot travels as its existing `EventHeader` / `PersonHeader`; the record pivot as a Subject id the page already has (D-11).

## Round 1

### M-1 · blocker · The canon pivot's date or name is a new way to show a handle, beside `conclusionheaders`
`readPivots`, a `pivotShown` kind → property table, `EvidenceValues.Pivot` and `pivot_values` would be a second display rule; `proposalProto` already sends handle headers through `conclusionheaders`.
**Answer:** Fixed — D-11: `PromoteClaimEvidenceShape` carries `pattern`, `pivot_id` and, on the canon side, `oneof pivot_header { PersonHeader; EventHeader }` read in one batch per kind as `proposalProto` does; `Recorded`, `PivotValues`, `pivotShown`, `readPivots`, `EvidenceValues.Pivot` and the pivot profiling are gone. AC-1.7 became a sections AC; AC-3.1 – AC-3.3 test the headers.

### M-2 · major · Choosing "the one shape to word" belongs in `core/kinship`
**Answer:** Fixed — D-9: `kinship.Recording.Shown()`; T-1 `TestRecordingShown`.

### M-3 · major · The add-match-rule skill's "Add a kinship pattern" steps become false
**Answer:** Fixed — STEP-8 updates the kinship bullet and "Add a kinship pattern"; DoD lists the skill.

### M-4 · minor · The vocabulary test hand-lists the patterns
**Answer:** Fixed — D-12 `kinship.Patterns()`; T-8 derives its `want` from it.

### M-5 · minor · One concept carries three names, and "Recording" means two things
**Answer:** Fixed — `Relationship.Layer/Canon` are `kinship.Shape`; wire `PromoteClaimEvidenceShape`, Swift `CatalogClaimEvidenceShape`, `evidenceShapeProto`, `mapEvidenceShape`. "Recording" is only `kinship.Recording`.

### M-6 · minor · `bridgesOf` reads ends from a union of every bridge's endpoint properties
**Answer:** Fixed — D-2 reads each bridge Subject's spec from its `EvidenceSubject.TypeKey`; a stop condition covers a Subject with no `EvidenceSubject`.

### M-7 · nit · A second `canonlinks.NewReader` per request
**Answer:** Logged — an Out line; sharing one reader through `inputs` changes `loadCanon`, which this PR does not touch.

### X-1 · blocker · STEP-3's check cannot pass: the module stops compiling until STEP-4
**Answer:** Fixed — STEP-1 renames `Bundle.Bridge` and fixes `BundlePinID` and `titles` in the same step; its check runs `go vet` over `./core/... ./api/...`.

### X-2 · major · The canon-side plumbing of D-1 to D-3 leaves who builds the sets and their types open
**Answer:** Fixed — D-1 names one signature (`canonLinks(from, to, rels, row, nbr map[string]bool)`), the set construction and the `accepted`-first order; D-3 says `canonLinks` builds the pivot set; pivot profiling is gone with M-1.

### X-3 · major · D-4 leaves `named`, Sections and recorded-vs-group order open
**Answer:** Fixed — D-4 says a set group `Neighbor` is named and a nil one names nothing; D-5 orders bundles by first bridge, then the stable sort. AC-1.7 and T-2 (b), (c) test it.

### X-4 · major · Pivot values are underspecified
**Answer:** Fixed by M-1 — no pivot values; the header picks what is shown. The canon pivot's kind comes from `Pattern.PivotKind` (D-12); a pivot Subject's kind for defining refs from its `EvidenceSubject.TypeKey` (D-8).

### X-5 · minor · The `marriageEvidence` change and its callers
**Answer:** Fixed — D-16: `marriageScene`, the fixture returns bridge Subjects, `TestPinOffersOfAnImpliedLayerLink` in STEP-4's files; §2 states the `parent` fixture direction; AC-1.6 states its third-Source setup; AC-1.8 the placement.

### X-6 · major · STEP-3 hides about six changes
**Answer:** Fixed — STEP-3 swaps the bridge reader alone for filed links and must pass every test unchanged; STEP-4 adds `Relationships`, pivots and `Placed`.

### X-7 · major · Checks that cannot run here
**Answer:** Fixed — the **Environment** header line: STEP-1 – STEP-5 on Go with cgo; STEP-6, STEP-7, `check-pr.sh` and the manual check on a Mac.

### X-8 · minor · Smaller choices left open
**Answer:** Fixed — `export_test.go` needs no change (§6); D-10 writes `paired`; AC-2.1 does not assert Ann's row; STEP-7 says to import `core/kinship`; D-13 says `init(wire:)` is written as `EvidenceOutcome`'s.

## Round 2

Round-1 answers: maintainability accepted M-1 – M-7; executability accepted all but X-5 (handler fixture, now X-10) and X-6 (moved into STEP-4, now X-11).

### M-9 · blocker · The canon pivot's headers would be the third hand-copied "read headers by kind, re-key by id" block
`proposalProto` and `search_catalog.go` `fillHitHeaders` each hold one copy; S13-04 and S13-05 would add more.
**Answer:** Fixed — D-17 adds `conclusionheaders.ByIDs` / `Headers`; `proposalProto` and `fillHitHeaders` move onto it (STEP-7, T-9); D-11 reads pivots through it; STEP-10 updates Spike 13's "One person header for every tile".

Below the bar, taken anyway: the `add-match-rule` layer-5 line now names claim evidence (STEP-10). Not taken: `canonLinks`' `BridgeTypeKey != relationship` filter beside `canonlinks`' private `relationshipEdges` — a one-line complement, and exporting it changes `canonlinks` (Out).

### X-10 · major · T-7 has no way to build the catalog
**Answer:** Fixed — D-18: T-7 hand-builds `claimevidence.Evidence` over `citedEvent`'s dated event handle and `newEvidenceFixture`'s person handle; AC-3.1 – AC-3.3 are worded on `claimEvidenceProto`; the catalog-built cases are T-4.

### X-11 · major · STEP-4 still hides about six changes
**Answer:** Fixed — STEP-4 (relationship links with recordings, every test unchanged), STEP-5 (pivots, T-4, T-6), STEP-6 (count-once, T-5).

### X-12 · minor · STEP-3 and D-16 leave small choices open
**Answer:** Fixed — STEP-3 calls `canonLinks` with a nil `rels` and reads every `from.Links` entry; D-1 says relationship links are skipped from STEP-4; D-16 says `marriageEvidence` stops calling `f.marriage`.

### X-13 · nit · Subtest renames are not in any step
**Answer:** Fixed — STEP-5 renames and flips them.

## Verdict
Two rounds done; no open blocker; To user empty. Plan marked Planned.
