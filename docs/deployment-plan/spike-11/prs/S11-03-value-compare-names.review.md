# Review — S11-03 Value compare: names

**Verdict:** round 1 — Fits with changes / Not ready → every finding fixed or declined; round 2 pending
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- Answered before review (2026-10-17): swapped part types stay under the 0.75 Near minimum and are `Different`; only a surname or first given name left over on both sides is `Different`, other leftovers on both sides are `Near` (D-11, map scope entry, comparison-layers worked case).

## Writer's own calls

- One-sided words count 0 in the similarity, so `Contains` scores below 1 (`James` / `James Robins` ≈ 0.571); the Near minimum gates the paired mean, so a one-sided surname never turns a spelling variant into `Different` (D-11 steps 9 – 10).
- `GivenOnlyFactor` and `SuffixConflict` leave the registry and the inventory: the relation rule makes them `Different` (D-4, D-6).
- A new knob `value.Names.Generation` 0.5 (estimate), because one-sided generation words now count (D-4).
- Name keys ignore alias tables, since alias sets do not chain (D-10).
- Two new causes, `role` and `below_minimum`, and five name part kinds (D-12).
- `match` reads every name knob, the pairing and the alias index from value compare, so no number or helper has two copies (D-8, D-13).

## Round 1

### M-1 · blocker · Two alias indexes side by side
**Answer:** Fixed. D-8 exports `Hit`; `match` builds an `AliasIndex` in `resolve` and calls `Hit`; `aliasSet`, `indexAliases`, `aliasHit` are deleted in STEP-3. Out row and Deletes changed.

### M-2 · major · Reuse text.go's part helpers
**Answer:** Fixed. D-11 step 9: `meanSimilarity` takes per-word weights, `positionalParts` and `wordParts` take part kinds; text passes weight 1 and `PartWord`. Codebase fit row added.

### M-3 · major · Catalog data by two routes; allow-list against the skill
**Answer:** Fixed, option (a). D-1: `Value.NamePattern` is a key and patterns come from `Rules.Names.Patterns`. D-5: `catalogData` lists `value.Names.Aliases` and `value.Names.Patterns`, read from the registry's type; the skill's "Do not" line is rewritten, not contradicted.

### M-4 · major · Western coverage guard moves with the value
**Answer:** Fixed. D-14 and T-10: `TestWesternPatternCoversEveryPartType` moves to `core/matchregistry`, without the `Key` loop; `TestDefaultIsFresh` changes `PartRoles` too.

### M-5 · major · Docs this PR makes false
**Answer:** Fixed. D-15 and STEP-7 add `matching.md` L19, L171, L186 – L189 and the `core/match/registry.go` header comment.

### M-6 · major · One name generator
**Answer:** Fixed. D-14: `randomName` / `variantName` in `contract_test.go` serve the contract and `TestNameInvariants`.

### M-7 · major · Generation weight outside the pattern; hand-written weight slots
**Answer:** Declined for this PR, logged. Stored formats would need a new slot and a seed row, and `LoadID`'s fallback would change what a stored format reads; the first named consumer is `ideas/international-names.md` (dual surnames), which is now the Out row's home. `Generation` stays a `Names` knob (D-4).

### M-8 · minor · Seed file named `registry.go`
**Answer:** Fixed. `core/database/namealiases/registry.go` and `registry_test.go`.

### M-9 · minor · Western written twice
**Answer:** Fixed. T-11 `TestSeedMatchesRegistryWestern`.

### X-1 · blocker · Empty `AliasIndex` breaks the `catalogData` check in STEP-2
**Answer:** Fixed. `AliasIndex` is declared in full in STEP-2, and D-5 reads `catalogData` from the type, so an empty value still counts.

### X-2 · blocker · Files STEP-1 breaks
**Answer:** Fixed. STEP-1 lists `core/match/registry_test.go` and `core/matchregistry/registry_test.go`; D-14 names the local `testAliasRoles` and the `valuecompare.` prefixes.

### X-3 · major · Suffix-only names and folded generation tokens in `Key`
**Answer:** Fixed. D-10: no key without a D-9 word; generation tokens are never folded.

### X-4 · major · AC-3.3 `rg` inside a table
**Answer:** Fixed. The command is in STEP-2's check with one `-e` per knob; AC-3.3 points at it.

### X-5 · major · Open choices in D-11
**Answer:** Fixed. Part types written out in AC-1.1, AC-1.12 and T-6; the role rule wins over the initial rule; `Parts` re-sorted into A's order after the mirror and the cause picked after it; `BestAssignment` transposes internally and `rowTo` is in `earned`'s orientation; AC-1.3's table roles; generation `PartMatch`es follow the word pairs.

### X-6 · major · Literal 2 in D-11
**Answer:** Fixed. D-11 step 4: every doubled weight is `w + w`; the skill says so (D-15).

### X-7 · minor · Order and size
**Answer:** Fixed: `export_test.go` and `core/match/words.go` in STEP-3; STEP-7's check catches the skill's L32 form; the map row links the plan. STEP-1's size stays one commit: a type move compiles only with every caller moved.

### X-8 · minor · Checks that need the Mac
**Answer:** Declined: STEP-8 already says the gate runs on the user's Mac, as `ship-pr` requires.
