# Review — S11-07 One normalization key across the core

**Verdict:** Fits — Ready with changes (all fixed)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. The three product calls (word breaks are `Same`; term duplicates against any origin on create and rename; the error names the existing label) were asked and answered on 2026-10-17 before review.

## Writer's own calls
- D-4: search rescoring keys at `Equivalent`, matching what FTS retrieves; cross-word substring hits (`tea` in `sainteanne`) are accepted.
- D-5: search, the index, terms and aliases key text with the shipped word rules, not `valuerules.Load`.
- D-6: index alternates are deduped at `Equivalent`, so an accent variant of the title no longer appears as a "matched via place" line (it now matches as the title); `ProjectionVersion` 8.
- D-7: `Create` keeps a slug-collision check against user terms, because the `ON CONFLICT … DO UPDATE` upsert would otherwise overwrite a user term (`A+B` / `A B`).
- D-9: alias tokens are keyed at `Same` (not `Equivalent`), because `Hit` is called with unfolded comparison tokens.
- `AC-1.6`: `CacheVersion` 17 so open catalogs regroup `York Town` / `Yorktown`.

## Round 1

### X-1 · blocker · D-1 Parts rule is wrong when word counts match but splits differ
`ab c` / `a bc` is `Same` but `positionalParts` would mark its pairs `Equivalent` / `fold`.
**Answer:** Fixed — D-1 now uses positional parts only when the word lists (or the folded lists, for `Equivalent`) are equal; otherwise one joined pair. AC-1.7 and T-1/T-3 rows cover `ab c` / `a bc` and `Mont-réal` / `montréal`.

### M-2 / X-2 · major · Keyed tokens break snippet centring for exact hits
**Answer:** Fixed — D-4 adds `queryTokens{raw, keys}`; scoring uses keys, snippets use raw tokens as today. STEP-3's check names `TestNoteMatchReasonUsesSnippet` and `TestNoteBodyRollsIntoSourceHit`.

### X-3 · major · Joined field keys give cross-word substring hits
**Answer:** Fixed — D-4 records it as accepted, with the reason (substring matching already crosses letters within a word; FTS decides retrieval).

### X-4 · major · AC-2.4 depends on an unstated retrieval path
**Answer:** Fixed — AC-2.4 and T-6 now test `scoreFields` directly (package `search`), independent of retrieval.

### M-1 · major · `duplicateOf` adds a second term query beside `sqlListByProperty`
**Answer:** Fixed — D-7 is now `conflictingTerm(tx, propertyID, selfID, label, slugKey)`: one pass over `sqlListByProperty` / `scanTermRows` covering both the text-key rule and the user slug rule; `Create`'s `sqlLookup` duplicate check is deleted.

### M-3 · major · Three names for the shipped rules; hand-built `Rules{Words: w}`
**Answer:** Fixed — `TextKey(s, level, w WordRules)`; each package holds `valueWords` with the `autoreconciler` `valueRules` doc comment (D-2, D-5).

### M-4 · major · `add-match-rule` not updated for the new callers and rebuild obligation
**Answer:** Fixed — STEP-7 adds two skill lines: `TextKey` with shipped words for catalog-independent callers; a text-key or `WordRules` change bumps `CacheVersion` and `ProjectionVersion`.

### M-5 · major · Alias token rule decided in two places
**Answer:** Fixed — D-9 adds `valuecompare.AliasToken` beside `Hit`; the index and `namealiases.aliasTokens` both use it.

### M-6 · minor · Header sort neither moved nor Out
**Answer:** Fixed — Out row for `autoreconcile.SortKey` (ordering, not equality); STEP-7 records which of the six text rules remain and why.

### M-7 · minor · "Next adds" commits S11-08 to work its map row does not name
**Answer:** Fixed — STEP-7 adds "text ignores word breaks (S11-07 D-1)" to the S11-08 scope entry.

### X-5 · minor · Mixed-case worked row for D-1
**Answer:** Fixed — AC-1.7.

### X-6 · minor · D-6 label order and keyless title
**Answer:** Fixed — label stays last; a keyless title excludes nothing.

### X-7 · minor · T-10 setup
**Answer:** Fixed — T-10 names the fixture `propID`, `Upsert` with provenencia origin for AC-4.2, and the "duplicate user key" param change from key to label.

### X-8 · minor · Seed tokens assumption
**Answer:** Declined — verified: every token in `core/database/namealiases/nicknames_seed.go` and `registry.go` is a lowercase single word (`rg` for non-`[a-z]` strings finds only license text), so stored and keyed tokens agree. `TextKey` reads only `Words` (D-2).

### X-9 · nit · Init-time `Default()`, map link
Logged. `matchregistry.Default()` builds fresh values with no dependencies (the `autoreconciler` `valueRules` var already does this at init). The map link is added in wiring.

### M-8 · nit · Rename `ErrDuplicateKey`
Logged; kept — the wire code stays `propertyterms.duplicate_key`, and the name matches its siblings in `properties`, `metadatafields`, `sourcetypes`.

## Round 2

Maintainability: no new blockers; every round-1 answer holds. Verdict Fits. Logged: `valueWords` is now held in five packages; if a sixth Go caller appears, a `matchregistry` accessor is the place for it. Alias storage keys with `valueWords` while the index keys with `base.Words`; both come from `matchregistry.Default()` today.

### X-10 · minor · T-12 fixture and read-back
**Answer:** Fixed — T-12 names the table fixture and the `TestInstallAndLoad` / `TestUserAndPluginRowsScore` read-back.

### X-11 · minor · `Update` order and inputs for `conflictingTerm`
**Answer:** Fixed — D-7: after the lock check, only when the label changes, trimmed label, `existing.PropertyID`.

### X-12 · nit · `kindSpec` returns two values
**Answer:** Fixed — T-6 writes `spec, _ := kindSpec(KindSource)`.
