# Review — S14-21 Search finds every spelling

**Verdict:** Fits with changes → all fixed; Ready with changes → all fixed. Planned.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- M-11 (2026-11-02): direct folds only; the chained-nickname gap (Ann for a Hannah) is written into Edges and pinned by T-6. Answered.
- None open. Answered before review (2026-11-02): keep one PR over the size limit (D-14); nicknames that fold in only (D-7); no match line for a nickname hit (D-8); folded spellings for every document kind (D-1).

## Writer's own calls
- D-2 – D-5: a `folded` FTS column written by `Upsert`, one shared fold rule (`searchindex.FoldedSpelling`) for documents, alternates and queries, query OR-expansion, bm25 weight 4.
- D-6: nickname expansions read the product alias seed, not the catalog's alias rows.
- D-9: the map's "compose unchanged dependents from the warm graph" is dropped (the warm graph would give stale reads inside the write); the scope entry is corrected in STEP-12.
- D-10: `SummaryDeps` keeps the full walk; only the search reprojection is narrowed.
- Patch bump.

## Round 1

### X-1 · blocker · D-2's dropped-rune clause contradicts T-1's "St. Søby" → "soby"
**Answer:** Fixed. D-2 is now exact: a dropped (non-break) rune counts only with a kept rune on both sides. "St." gives nothing; "1,200" → `1200`. T-1 has rows for both.

### X-2 · major · T-3's setup and expected kinds are unstated
**Answer:** Fixed. T-3 names `evidencegraphstest.SoleGraph`, the `evidencegraphs.Update` input, `runSourceTypeCreate` / `runFieldCreate`, and the expected kind per query. AC-1.4 lists the kinds. §6 Edges says "Aero" reaches `folded` through the last token's `*`.

### X-3 · major · `largeFixture.city` is not added by any step
**Answer:** Fixed. STEP-7 adds the field, sets it, and adds the budget const. T-9 creates the three Subjects before timing with `f.place(lf.other, lf.otherA, "City")` and times only `f.join`.

### X-4 · major · The FakeStore target body and leftover state are unstated
**Answer:** Fixed. D-12 gives the body in order inside `withState`. §4 says the state the helpers read has other readers and stays, and names the tests that use `searchCatalogError`, which stay unaffected.

### X-5 · major · The Swift seeded hits are unspecified
**Answer:** Fixed. §8 gives a file-private `hit(kind:id:)` builder for T-10 – T-13, and each T-14 test's seeded hits field by field, with the seeds that go.

### X-6 · major · Swift checks need macOS
**Answer:** Fixed. STEP-9: on a host without macOS, stop after STEP-8's `rg` and report, so the orchestrator runs `check-pr.sh` on macOS. The DoD says "on macOS".

### X-7 · major · AC-2.4 and T-5 depend on seed directions the plan does not show
**Answer:** Fixed. T-5's tiny index is given as an `AliasTable` literal and kept apart from the product rows. The product rows pin tom = [tommy] and thomas = [thom thos tom tommy]. AC-2.4 says it rests on the product seed's direction. A stop condition covers a different count.

### X-8 · minor · The map row does not link the plan
**Answer:** Fixed in this planning commit (map row Plan column).

### X-9 · minor · STEP-1's check cannot pass as worded
**Answer:** Fixed. The migration and the column writes are one step (STEP-1), checked by the package tests. §6 Schema says nothing writes the index between the step and `RebuildAll` (`catalogsession` runs both `EnsureCatalog` calls on open).

### X-10 · minor · Smaller choices
**Answer:** Fixed. D-4 says `FoldedSpelling` runs after `sanitizeFTSToken`, and `*` placement is as today, even when the last token sanitizes away. §4 says a name with no parts gives `nameWords` nothing, so it gets no lines (`Form` is not split). D-10 keys moved ids by `string(entity_id)` per batch. STEP-11 names every `MARKETING_VERSION` line.

### M-1 · blocker · The fold rule is written twice
**Answer:** Fixed. D-2: one exported `searchindex.FoldedSpelling` is used by `foldedWords`, `alternateKey` and `core/search` `buildMatchQuery` (D-4). The raw split is `valuecompare.RawTokens`, exported (D-15), not copied. `core/search`'s own `valueWords` stays for rescoring keys (Out row). It no longer folds anything itself.

### M-2 · blocker · `nicknamesOf` re-derives a name's words and roles
**Answer:** Fixed. D-7: `valuecompare.NameFoldsInto(v, r)` sits beside `WordKeys` and is built on `nameWords`, `tooManyWords`, `AliasToken` and `AliasIndex.FoldsInto`. `searchindex` calls it. `add-match-rule` is loaded for STEP-1 and STEP-4.

### M-3 · blocker · `displayedDigests` adds a third scan and a second equality rule
**Answer:** Fixed. D-10 / D-11:
- `recomputeBatch` reads the batch's stored rows into `graphcache.TruthRow` before its `DELETE` (`storedRows`, one query per batch).
- It builds the new rows with `truthRow` from `res.Values` and compares them with `graphcache.SameDisplay`. There is no after-query and no hash.
- `valueEqual` is rewritten on `SameDisplay` through `rowOf`.
- STEP-6 adds the verify-mode `promotealign` run.

### M-4 · major · `add-searchable-kind` is not told about `folded`
**Answer:** Fixed. STEP-10 adds the sentence to §3 step 1: `Upsert` writes `folded`, the query OR-expands, a projector never writes FTS rows, and the trigram table keeps four columns.

### M-5 · minor · Two rule variables in `searchindex`
**Answer:** Fixed. D-6: `valueWords = productRules.Words`. Memoizing `namealiases.ProductIndex` is not taken: it is a change outside this PR's files for one extra index built once per process.

### M-6 · minor · Secondary tags are bare literals
**Answer:** Fixed. D-8: `SecondaryTagName`, `SecondaryTagToponym` and `SecondaryTagAlias` sit beside the `BodyTag*` constants and are used in `tagged` and `alternateMatch`.

### M-7 · minor · The map scope entry stays false
**Answer:** Fixed. STEP-12 rewrites the S14-21 Ships line, and its text is in the step.

### M-8 · nit · Say the alias lines' lasting purpose after S14-26
**Answer:** Fixed. One sentence in D-7.

## Round 2

Executability round 2 confirmed that X-1 – X-10 hold.

### X-11 · blocker · `locatedBirth` cannot be extracted as specified
**Answer:** Fixed. STEP-6 now gives the signature `locatedBirth(toponym string, forms ...string)`. The helper covers:
- the person from `forms`
- the birth event: type `birth`, date 1817
- the place, with its toponym
- the three positions
- both bridges

It stops before `promote` and `fileBridges`. `TestSummarySearchFindsTheListTitle` passes "James Robins", "Kenneth Robins" and keeps every assertion.

### X-12 · blocker · Query tokens with a dash or slash break `FoldedSpelling`'s one-token contract
**Answer:** Fixed. D-4 splits each sanitized token with `RawTokens` and folds each part, writing `("tok" OR ("p1" AND "p2"))`. T-4 has a `søby-kirke` row and a "Soby-Kirke" retrieval case. The reviewer's other point was that dots and commas mid-word would add `folded` text to filenames and numbers. In answer, D-2's punctuation clause now covers mid-word apostrophes only, so existing documents' bm25 stays as today. T-1 rows: "1,200", "scan.pdf" → none; "D’Arcy" → `darcy`.

### X-13 · major · T-5's `Rules` and `Value` are unstated
**Answer:** Fixed. T-5 names:
- the package: external `valuecompare_test`
- the rules: `matchregistry.Default().Value` with `NewAliasIndex`
- the value: `namevaluetest.Western`
- the cap: the 33-word case runs against `MaxWords` 32

### X-14 · major · Row order and keying in `recomputeBatch` are implicit
**Answer:** Fixed, then superseded by M-9. Rows are compared per entity and property in rank order (`StoredValues` returns them so, and `res.Values` index + 1 is the rank), so no cross-property sort is needed. An id with no rows on either side is unmoved.

### X-15 · minor · The trigram `'delete'` row
**Answer:** Fixed. §6 says the trigram `'delete'` row keeps four columns.

Maintainability round 2 confirmed M-2 and M-4 – M-8 are carried into the plan. It found M-1 and M-3 only partly carried and raised three blockers.

### M-9 · blocker · `storedRows` is a second read of the stored row, and `truthRow` a second build beside `insertRow`
**Answer:** Fixed. D-10 and §6:
- `loadValues` is exported as `(*Graph).StoredValues`, and `recomputeBatch` reads the before rows through `graphcache.Over(q).StoredValues`. `storedRows` is gone.
- `insertRow` takes the `TruthRow` that `truthRow` built and adds only `date_lo`, `date_hi` and `sort_key`, so each row is built once and the compared row is the written row.
- The comparison is `SameDisplay(rowOf(v), t)`.

### M-10 · blocker · `alternateKey` and `FoldedSpelling` compose the fold themselves
**Answer:** Fixed. `FoldedSpelling`, `alternateKey` and the query's plain parts all use `valuecompare.TextKey(…, Equivalent, …)`, the key rescoring reads. `accentOnly` is `valueWords` with no `Folds`. The apostrophe test is not taken into `valuecompare`: since round-2 X-12, D-2 counts only mid-word apostrophes (U+0027, U+2019, U+02BC). That is a fact about SQLite's tokenizer, not a rule for which runes a word drops, so it does not copy `wordRune`.

### M-11 · blocker · `FoldsInto` lists direct folds; the reconciler follows chains
**Answer:** To user, answered 2026-11-02: direct folds only. I measured chains on the product seed: following them grows 2,627 pairs to 3,478, up to 21 per name, with 286 names gaining, and brings "Fred" to William and Robert. Edges Assumption 5 states the gap, Real data gives hannah's direct folds, and T-6 pins "Hannah Smith" with no `alias:\tann`.

### Below blocker (writer's call)
- The Out row's STEP reference was wrong: Fixed (STEP-12).
- `NameFoldsInto` used `slices.Contains(AliasRoles(), …)` where `PrefixRuleRole` already exists: Fixed.
- A shared Swift `CatalogSearchHit` fixture file: Declined. The builder is one line, private to `SearchCatalogStoreTests`, and the T-14 seeds are per-test values a fixture would only rename. Logged for the next PR that builds hits.
