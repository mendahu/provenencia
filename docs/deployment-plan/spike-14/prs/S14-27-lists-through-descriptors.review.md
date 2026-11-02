# Review — S14-27 Every write reaches its keys through their descriptors

**Verdict:** Fits with changes — Ready with changes (round 1, every finding Fixed; no blocker left open, so no round 2)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None. (The order — S14-27 ahead of S14-17 — was the user's answer on the S14-17 review, 2026-11-02.)

## Writer's own calls

- One slot type: `CatalogQueryFollows.list(_:)` and `.structure` report one token, used only through `applyHiddenPolicyOnAny(follows:)`; no second "lists followed" property on descriptors.
- `.deletedSubject` is retired here rather than left for S14-17 (M-1); the guidance that teaches it, one `.cursor/rules` line included, is rewritten to stay true, which bends no rule.
- Version bump none, as S14-24 took none.

## Round 1

### M-1 · blocker · `.deletedSubject` reloads `propertiesWorkspace` a second way; the Out line's home kept it
**Answer:** Fixed — D-6 deletes the mutation, its kind and arm; both callers apply effects alone; AC-2.5 is effects-only; AC-2.6 and T-7 pin that a Subject delete reports `property_usage`; D-5 rewrites the Subject-erase lines in `add-workspace-place`, `add-catalog-delete`, `.cursor/rules/catalog-deletes.mdc` and `docs/catalog-deletes.md`; the Out row is gone; S14-17's D-6, D-13 and Out list updated to match.

### M-2 · major · The catalog-wide rule is enforced at one of four constructors
**Answer:** Fixed — D-1 puts the `isCatalogWide` precondition in `applyHiddenPolicy(follows:ids:)`, `applyHiddenPolicy(followsValue:ids:)`, `merged` and `CatalogQueryPatch.lastUpdated`; `applyHiddenPolicyOnAny`'s doc names it the constructor for catalog-wide slots; the type doc is reworded; T-4 also checks `patches`.

### X-1 · major · Manual check 3 cannot show a store call
**Answer:** Fixed — the debug-log claim is gone; the load count is T-3 / T-5's `getPropertiesWorkspaceCalls`; a Subject-delete step is added.

### X-2 · major · When T-2 and T-3 are written
**Answer:** Fixed — STEP-1 writes T-1 to T-4 and checks all four; STEP-2 deletes `listInvalidations` and `listCases`.

### X-3 · major · How the tests make the key visible and when they count
**Answer:** Fixed — T-3 follows `eachReadFollowsOnlyWhatFeedsIt`'s setup (`apply(location:)` on the graph surface), baseline after `waitForQuery`, assert after `settleWrites()`; hidden cases warm by `setQueryValue`.

### X-4 · minor · The 899–910 conversion has no session
**Answer:** Fixed — D-4 moves it into T-2 with only `propertiesWorkspace` warmed and `==`.

### X-5 · minor · Doc wording not given
**Answer:** Fixed — D-3 and D-5 give the `apply` doc, `CatalogWriteEffects.lists` doc, rule 2 sentence and descriptor comment verbatim; the "Docs in code" row is corrected (`CatalogList.counts` stays true).

### X-6 · minor · Small inaccuracies
**Answer:** Fixed — "Must stay green" names the right suites; T-6 is labelled an `rg` source check with the empty output as its evidence; table order fixed.
