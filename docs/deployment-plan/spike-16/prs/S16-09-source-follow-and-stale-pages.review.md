# Review — S16-09 Conclusion pages refresh from the Sources they cite and show their last value while they refresh

**Verdict:** (round 2 pending)
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- The Batch 1 rule "A cached detail follows the Sources its records cite" says a hidden detail "is evicted, as today"; the Batch 2 rule sets `.markStale` for the same key. The plan treats your merge of S16-23 into S16-09 (2026-11-08) as the answer and rewrites the Batch 1 clause (D-12, **Rules this plan bends**). Confirm.

## Writer's own calls
- D-4 / D-6 (from M-1, M-5): the accessor is `ownId` over every single-id key, not `entityId` over the five per-entity keys, so the deleted eviction is declared per kind (`deletedWith`) and four more arms lose their `guard case`. A slightly larger diff in `CatalogQueryRegistry.swift` than the map row names; the closures it removes are the ones S16-25 would remove.
- D-5 (from M-3): option (a), `ThumbnailCache`'s one-array shape copied, not a shared `RecencyOrder` type; extracting it would touch `ThumbnailCache` outside this PR for a five-line idea.
- D-4: a visible key of a deleted entity is not evicted (the page would stay on a spinner); it reloads and reads gone.

## Round 1

### M-1 · blocker · The deleted-entity eviction only works for handles, and the next deleted entity has no slot
**Answer:** Fixed. D-4: `CatalogQueryDescriptor.deletedWith: CatalogDeletedEntity?` declared on `.conclusionDetail`, `personSection`, `personGraphsSection`; `apply` walks every `effects.deleted` entry through `evict(_:ids:)`. D-5's cap counts keys whose `deletedWith` is `.handle`. Next adds names the deleted-Sources entry. T-7 (a `.sourceWorkspace` key survives a handle deletion), T-14.

### M-2 · major · D-9's comment and a Stop condition are untrue for the Source page's `leavingPage` caller
**Answer:** Fixed. D-9's comment states what holds for both callers; §4 row "Source page leave"; the Stop condition is narrowed to `PersonSourcesSection`'s leave.

### M-3 · major · D-5 adds a second recency mechanism beside `ThumbnailCache`
**Answer:** Fixed with option (a). D-5: `shownOrder: [String]` in `ThumbnailCache`'s shape; counter stamps and the sort go; an id not in the order is oldest. §4 row "Recency precedent", Codebase fit row. Option (b) declined: see Writer's own calls.

### M-4 · major · STEP-7 and the sweep miss text this PR makes false
**Answer:** Fixed. STEP-7 adds rule 2's `deleted` clause, the `applyHiddenPolicy(to:)` comment and the skill's own-id sentence; STEP-3 rewrites `apply(_:_:)`'s comment; the DoD sweep adds "`apply` ignores it", "evict, mark stale", "evicts or marks stale" and covers `macos/App`.

### M-5 · minor · `entityId` replaces 3 of 7 own-id closures
**Answer:** Fixed. D-6: `ownId` and `applyHiddenPolicy(followsOwnId:)` replace six; `.citationsByArtifact` (two ids) keeps its closure until S16-25 (Next adds).

### X-1 · blocker · `showCount` start value and tie-break
**Answer:** Fixed, then superseded by M-3: there is no counter; an id not in `shownOrder` is oldest.

### X-2 · major · T-7's second half and AC-3.2 need `handles` beside `deleted`; shorthand `apply(…:)` calls do not exist
**Answer:** Fixed. AC-3.2 and T-7 pass `handles` and `deleted`; every test row spells `CatalogWriteEffects(…)`.

### X-3 · major · The Swift checks need a Mac
**Answer:** Fixed. The **Planned** line says so, as S16-04's does.

### X-4 · minor · "22 `stalePolicy` lines"
**Answer:** Fixed. 20: 18 arms plus the two helpers.

### X-5 · minor · Header says Draft with a Planned line
**Answer:** Declined. The template puts the planning base on **Planned:** for a Draft too; Status moves to Planned after this review.

### X-6 · minor · T-1's fixtures
**Answer:** Fixed. T-1 builds records with `ReconciledValueDisplayTests.outcome(_:_:source:)` and `field(_:outcomes:)`.

### X-7 · minor · On-screen entity vs on-screen key
**Answer:** Fixed. D-5: every key of an on-screen entity is excluded, a section key the place did not warm included; AC-4.4 and T-9 test it.

### X-8 · nit · Stale "evicted" prose in test comments
**Answer:** Fixed. STEP-4 names each comment.

## Round 2

Maintainability: every round-1 answer checked against the code; no new blocker. Verdict: Fits with changes.

### M-6 · minor · `deletedWith` also marks the cap's per-entity keys
**Answer:** Fixed. D-4: `deletedWith`'s doc comment names both readers (the deleted eviction and the hidden-entity cap).

### M-7 · minor · The sketch's `ownId` doc comment contradicts D-6
**Answer:** Fixed. The sketch comment is D-6's wording.

### M-8 · nit · `maxArcs` does not exist
**Answer:** Fixed. Codebase fit cites `PersonRecordLinksContent.maxNodes`.
