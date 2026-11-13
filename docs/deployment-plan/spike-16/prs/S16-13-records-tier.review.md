# Review — S16-13 Per-graph records in the cache registry, and the Sources section reads them

Plan: [`S16-13-records-tier.md`](S16-13-records-tier.md). Lint: `scripts/check-plan-docs.py` clean before round 1.

## Round 1 (2026-11-13)

Maintainability (fresh, strongest model): **Fits with changes**. Executability (fresh, implementing model): **Not ready**.

### Maintainability

### M-1 · blocker · The loaded-graph store stays a hand-written LRU beside a registry that can now hold it
**Answer:** Fixed — D-2: the store is a keyed entry inside `graphcache` (`graphcache.evidencegraphs`), Cap `MaxEvidenceGraphs`, Apply `DropNamedGraphs`, Check the registered check; its LRU, drop, Verify loop and typed field go; `SetCapForTest` replaces the store's test hook. STEP-3 (new). Within the user's one-PR answer and their 2026-11-13 reading of "simpler code".

### M-2 · blocker · Layers and candidates register with no Check, against `catalog-queries.mdc`
**Answer:** Fixed — D-8 (layers: rebuild with `build` and DeepEqual) and D-9 (candidates: rebuild when the stored generation is current, through the extracted `rankCandidates` and `shortlistInputs(g, true)`); AC-4.8; a Stop if the candidates Check needs a second ranking path.

### M-3 · blocker · The name pattern is hand-checked instead of `DerivedFrom nameformats.Default`
**Answer:** Fixed — D-8, D-9: both entries derive from `Vocabulary{}` and `nameformats.Default`; `evidencelayer.Load` drops its pattern parameter and reads `nameformats.LoadFor(g).Key` (every caller passed exactly that, via `valuerules.Load`); `patternLayer` and the pattern field are gone; AC-4.2 replaces the pattern-mismatch case.

### M-4 · major · `Claim` copies `SubjectRecord`'s fields
**Answer:** Fixed — D-11: `Claim{Graph; graphfill.SubjectRecord (embedded); Records}`; the claim id is `cl.Claim.ClaimID`.

### M-5 · major · Subject label map built in two places
**Answer:** Fixed — D-5: `graphcache.SubjectLabels`, called by `newGraphRecords` and `listedContext.rows`.

### M-6 · major · Layer cap 64 is below the 80-graph Scale and record links read the same section
**Answer:** Fixed — US-1 and the Objective narrowed to the claims read (the map already keeps record links' whole-graph loads out of S16-13); AC-1.7 and T-10 add `TestRecordLinksBudgets` at 80 graphs under `budgetLinksPast` with the planning-base figure. Raising the caps stays Out (not planned), as the map and design already say.

### M-7 · minor · 64 written three times
**Answer:** Fixed — `graphcache.MaxEvidenceGraphs` is the Cap of the store, layers and candidates.

### M-8 · minor · Two rules for a read that does not fill
**Answer:** Fixed — D-6: `PeekEvidenceGraph` is the store entry's `Peek`, which counts a use like every Peek.

### M-9 · nit · A third copy of the Scale seed in `BenchmarkSourceRecords`
**Answer:** Logged — Out row (not planned).

### Executability

### X-1 · blocker · GetMany's eviction order over a map
**Answer:** Fixed — D-1: held ids touched in `ids` order, then fills stored in `missing` order; unasked keys neither stored nor returned; AC-3.3 spells the result.

### X-2 · blocker · `Layers.Get` cannot express a pattern mismatch
**Answer:** Fixed — with M-3 there is no pattern in the value; D-8 spells `Load` as `Layers.Get` with its fill and `ErrNoGraph`.

### X-3 · major · T-8 has no seam
**Answer:** Fixed — T-8 crafts records with the exported `graphfill.Records.Set`; the missing-shell and missing-Source branches are named unreachable in Reach (a deleted graph leaves the node and index on the same commit; a Source with graphs cannot be deleted).

### X-4 · major · A missing Source record
**Answer:** Fixed — D-11: an error, as today.

### X-5 · major · SeedScale's contract
**Answer:** Fixed — D-14: `ScaleSeed` with its constructor, methods, constants and id kinds; the crafted half stays in `BenchmarkGraphIndex`; records bench needs no claims.

### X-6 · major · Test tables as intents
**Answer:** Fixed — §8 names the fixtures (`evidencelayertest.Records`, `promotetest.Save`, `households`), packages, budget constants; T-6 rewrites the existing memo tests instead of adding one; T-7 extends `TestClaimsWarmReadIssuesNoQuery`.

### X-7 · major · Label map key type and sharing
**Answer:** Fixed — D-4, D-5, D-13: keyed `string(id)`, shared and read-only; `rows` builds it once per call.

### X-8 · major · Smaller open decisions
**Answer:** Fixed — held not-Found graph → empty records (D-6); records asked only for ids with a shell (D-11); nil Observations for none (D-4); read-only returns (D-4); a stale-generation Peek's touch intended (D-9); the effects README sentence given in full (STEP-8); the worktree path and the before/after comparison spelled (STEP-7). LRU growth past +20% stays a Stop.

### X-9 · minor · `family.go` comment names `memberGraphs`
**Answer:** Fixed — STEP-5 files; the DoD sweep covers `core/` comments.

### X-10 · minor · checks not runnable on Linux
**Answer:** Fixed — §8 says `check-pr.sh`'s Xcode half runs on macOS and the Go half is reported on Linux.

## Verdict

Round 1 raised blockers that were all Fixed and one reshape (M-1). Round 2 is due.

## Round 2 (2026-11-13)

Fresh reviewers given the plan and this log. Maintainability: **Fits with changes** (round-1 answers hold; two new blockers). Executability: **Not ready** (two blockers, five lesser).

### M-10 · blocker · The store entry registered in `graphcache`, fed by `graphfill`'s setter hooks
**Answer:** Fixed — D-2: `graphfill` owns `graphfill.Graphs` with `EvidenceGraph`, `PeekEvidenceGraph`, `LoadEvidenceGraph` and `MaxEvidenceGraphs`; the ten callers switch; the setters, hook vars and `*ForTest` hooks leave `graphcache`. STEP-3.

### M-11 · blocker · Layers and candidates Checks fill other entries during Verify
**Answer:** Fixed — D-1: a Check reads no keyed entry through `Get`, `GetMany`, `Peek` or `Set`; D-8 and D-9 rebuild from `graphfill.LoadEvidenceGraph` and `evidencelayer.Build`; AC-4.9 and T-14 hold Verify read-only.

### Below the bar: `cl.Claim.ClaimID` names two concepts
**Answer:** Fixed — `SubjectRecord.Accepted`; the claim id is `cl.Accepted.ClaimID`.

### X-11 · blocker · STEP-2 and STEP-3 order
**Answer:** Fixed — STEP-2 removes memos only (store stays, without memos; `graphfill.MaxEvidenceGraphs` introduced; the cap test's memo cases go there) with its own `rg`; STEP-3 moves the store and runs AC-4.6's full `rg`; Scope rows and T-13 relabelled.

### X-12 · blocker · Home of the test cap override
**Answer:** Fixed — D-1: `Graph.capsForTest`, by entry id, survives `Drop` and `Clear`, applies to later containers; T-1 and T-1b test it across `Drop`.

### X-13 · major · Candidates Check recipe
**Answer:** Fixed — D-9 spells the inputs (`matchIndexOf`, `shortlistInputs(g, true)`, `LoadEvidenceGraph`, `Build`, `valuerules.Metas`), Subject selection, missing ids as mismatch, and `[][]byte{}` lists.

### X-14 · major · `ScaleSeed` sources and signatures
**Answer:** Fixed — D-14: `Base` seeds `ScaleSources+2`; method signatures; a second `NewScaleSeed` does not seed; imports checked.

### X-15 · major · 80-graph record links budget unknown
**Answer:** Fixed — STEP-1 measures the base first; D-15 adds `budgetLinks80` from it (× 1.25, not below `budgetLinksPast`).

### X-16 · minor · Drop order and Peek during Verify
**Answer:** Fixed — D-3 states the drop now runs in `applyEntries` and why it is safe; D-1's Check rule covers Peek.

### X-17 · minor · T-8 fixture
**Answer:** Fixed — T-8 names `f.source` and `evidencegraphstest.SoleGraph`.

## Verdict

Two rounds done. Every blocker Fixed; nothing To user. Plan **Planned**.

## Amendment — 2026-11-12, ship round 0

**D-1:** the LRU's use-order list is held in the items of `keyedValues[T]` (prev/next links in the item), not `container/list`. The implementer stopped on §11's kinship +20% bound: `container/list` measured +20.4% / +34.6% (4,000 / 16,000 persons); the in-item list measured −6.6% / +3.2%. Eviction semantics unchanged. User approved the amendment ("Amend D-1, resume"). The 20% stop stays.

## Amendment — 2026-11-12, ship round 1

**D-15:** `budgetLinks80` is 200 ms (planning base 113 ms x 1.25, rounded up); "no lower than `budgetLinksPast`" is dropped, since a 1.5 s floor could not catch growth. User: "Set 200 ms".
**CON-2:** `EmptyShortlistForTest` keeps minting an index generation (a §11 test-helper deviation). User: "Keep it". Recorded in the Landed note.

## Amendment — 2026-11-13, CI fix

**D-15:** `budgetLinks80` is 600 ms (CI 414 ms x 1.25, rounded up). `promotealign` `budgetWarm` is 500 ms (CI 383 ms x 1.25) and `budgetComposeAll` 900 ms (CI 679 ms x 1.25; still fails a per-row place graph at about four times the cost). CI run 43364250762. Raised by the user on 2026-11-13 (ship log F-7).
