# Review — S14-22 Promote's wire says what Go decided

**Verdict:** Round 1: Fits with changes — Ready with changes. Round 2: pending.
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- None. Two product calls were asked while planning and answered (2026-11-03): the loop warning reads "Not promoted · these places would form a loop" (one string for both terms); a bridge with a missing end, two records on one end, or no type reads "Waits for its other end" (no new state).

## Writer's own calls

- The filing rule extends `core/connectrules` (no new package); `graphalign.Connections` maps its result to wire outcomes.
- The page's scope goes to Go (`only_subject_id`), and "Map the rest" proposes again.
- The subrows keep the last proposal's outcomes until the next one lands; no Swift prediction.
- FakeStore forwards scripted connections and computes none.
- A proposal failure after a decision leaves the last outcomes and Done enabled: Out → S14-20.

## Round 1

### M-1 · blocker · Only two of the filer's refusals are shared; D-5 rewrites the rest of the filer's ladder in graphalign
The draft shared self-link and loop through `connectrules.Refuse` and copied the missing-end, two-records, no-term and first-waiting-end steps into `graphalign.Connections`, already in a different order from `fileBridge` (term before self-link).
**Answer:** Fixed — D-1 now puts the whole ladder in `connectrules.Decide(b, term, ends [][]Value, filed)`, in `fileBridge`'s order, with `Bridge.KeysOnType` moved from `identityclaims.keysOnType`. D-3: the filer builds `Value`s from its claim reads and calls `Decide`. D-5: `Connections` builds `Value`s from page targets and only maps the result. T-2 is one row per rung, including the two order cases the reviewer named.

### M-2 · major · Two lists the next refusal must keep in step
**Answer:** Fixed — D-4: `ConnectionOutcomes()` is files, the three waits, then `connectrules.Refusals()` converted; `ConnectionFiles` derives from `DecisionFiles`. Next adds updated.

### M-3 · major · "One rule, two loaders" read different filed-link sets; parity unstated
Checked: `graphcache` `linkLoadSQL` reads `auto_reconciler_values` rank 1 `kept` on unmerged associations; the filer reads claims and recorded Observations.
**Answer:** Fixed — §6 Edges states the parity assumption and when it holds; T-5 adds a merged place handle case and a type-changed-after-filing case asserting the proposal's outcome equals what `SaveBatch` files; §11 already stops on any T-5 count mismatch, which then goes to the user.

### M-4 · major · `BridgeFact.alreadyPromoted` is a second answer to "is this bridge filed"
**Answer:** Fixed — D-11 deletes it and its assignment; Deletes lists it; T-9 rewrites `promoteAllStartsExpandedAndHidesPromotedBridges` with a scripted proposal that leaves the filed bridge out.

### M-5 · minor · Term literals left elsewhere; `propertyterms.Key*` lose their last non-test use
**Answer:** Fixed in part — D-2: `hops.ParentsOfPlace` / `SuccessorsOfPlace` and `segmentcomparetest/terms.go` use the `vocabkey` constants. Declined: deleting the `propertyterms` aliases — 25 `conclusionsummaries` test call sites use them as `placeRel` arguments and both now derive from `vocabkey`, so there is one spelling.

### M-6 · minor · Plain strings beside typed `vocabkey` keys
**Answer:** Fixed — `AcyclicTerms []vocabkey.BridgeTerm`; `PageBridge.TypeKey vocabkey.SubjectType`, `Term vocabkey.BridgeTerm`; `Decide` takes `vocabkey.BridgeTerm`.

### M-7 · nit · Swift keeps `WaitReason` beside the wire enum
Logged. D-11 keeps `WaitReason` as the page's three wait words.

### X-1 · major · Checks that cannot run without protoc and Xcode
**Answer:** Fixed — §5 "Where it runs": STEP-1 – STEP-4 run anywhere; STEP-5 – STEP-8, T-9 – T-11 and the manual check need the macOS machine; an implementer without them stops after STEP-4.

### X-2 · major · FakeStore drops scripted connections
**Answer:** Fixed — D-14 names both construction sites, where `lastPromoteOnlySubjectID` is set, and the stub; STEP-6 repeats it.

### X-3 · major · `Propose` callers and the fixture
**Answer:** Fixed — D-8: `propose` takes `only`; `f.propose` keeps its signature; T-5 calls `Propose` directly for `only`. STEP-4 lists `memo_test.go` and the `rg` for callers.

### X-4 · minor · Unstated `ends` cases and term order
**Answer:** Fixed — D-8: `Start` New/Skip cannot occur unfixed; an absent Subject is not on the page (D-5); observations in `eg.Observations` order, first term.

### X-5 · minor · The snapshot-missing branch of `repropose`
**Answer:** Fixed — D-11: outcomes change only in `load` and `merge`; that branch keeps the old ones.

### X-6 · minor · Shape of the new L10n entries
**Answer:** Fixed — D-15: `static func connectionCycle() -> String`, a `.closesCycle` case inside `connectionSpoken`, both keys and comments, English-only xcstrings entries.

### X-7 · minor · The "tolerate an unknown value" subtest
**Answer:** Fixed — D-16 adds `PromoteConnectionOutcome` to that list.

### X-8 · minor · AC-2.6 and existing tests not in the rewrite list
**Answer:** Fixed — T-4 row for two anchor ends; T-9 adds `bothAnchorsOnAnOpenLinkAreCounted` and rewrites `promoteAllStartsExpandedAndHidesPromotedBridges`.

### X-9 · nit · "1 connections" and the audit's B-23 text
**Answer:** Fixed — two Out rows: the plural is the existing format string; the audit text stays as the record.

## Round 2

### X-10 · blocker · The AC-3.1 / STEP-7 / manual-check `rg` cannot return "no hit"
`func resolve(` also matches `core/database/effects/resolve.go`, `EvidenceSubjectKindStyle.swift`, `PVField.swift`, `PlaceRegistry.swift`.
**Answer:** Fixed — AC-3.1, STEP-7, manual check 5 and the DoD sweep use `rg "placeRelationshipWouldCycle|keysOnType" core` and `rg "EndResolution|func resolve\(_ subjectID|alreadyPromoted" macos/App/Features/Promote`.

### X-11 · major · STEP-4 misses `export_test.go`; direct `Propose` calls
**Answer:** Fixed — STEP-4 lists `export_test.go` and says which calls gain which `nil`.

### X-12 · minor · T-5's parity cases have no fixture
**Answer:** Fixed — the type-change case uses `observationstest.UpdateObservation`, with a mismatch as §11's stop. The merged-handle case is dropped with evidence: no app write sets `merged_into_id` (`rg "SET merged_into_id" core` finds only `graphcachetest`); §6 Edges says so.

### X-13 · minor · D-3 moves the `property_terms` read ahead of the wait and self-link rungs
**Answer:** Fixed — D-3: a term id with no `property_terms` row reads as no term, not an error.

### X-14 · minor · Waiting-end order differs between filer and proposal
**Answer:** Fixed — one line in §6 Edges: display only; both leave the bridge unfiled.
