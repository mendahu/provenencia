# S14-21 — Search finds every spelling

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the omnibar row draws as before. A nickname hit has no match line, which the row already does for a match reason it does not name)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-21-search-finds-every-spelling`
**Version bump:** patch — two search regressions (B-21, and S11-28's nickname fold hiding the nickname) and a performance fix; the schema step is additive and old apps still open the file ([`docs/versioning.md`](../../../versioning.md)). The next patch after `VERSION` at merge (0.2.8 → 0.2.9 at planning)
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (STEP-4, STEP-5: shared word helpers in `core/valuecompare`, `AliasIndex` as the one alias lookup; no comparer, relation, cause or knob changes), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md). `add-catalog-model` does not apply: `catalog_search_docs` is already registered (`BucketSkip`), the step adds no table and no foreign key, and `catalog_search_fts*` is outside `TestPragmaHonesty`
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/searchable-kinds.mdc`, `.cursor/rules/versioning.mdc`
**Context:** audit 2 [B-7, B-10, B-21](../app-health-audit-2.md) (map review K-45); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §6 "Eliminated means not displayed, never dropped" (search reads every cached value); Spike 11 dogfood bar 6 and S11-07 D-6 (retrieval assumed to fold like rescoring); the map's Spike 11 leftover (2026-10-29: "Tom" no longer finds a person shown as Thomas); working rules: "FakeStore is a seeded fake, not a second engine", "A nickname folds into its set's formal name, as today", "Audit items keep their IDs"
**Planned:** 2026-11-02 against `main` @ `7b9b6e7`
**Review:** [`S14-21-search-finds-every-spelling.review.md`](S14-21-search-finds-every-spelling.review.md) — round 1 answered

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The omnibar finds a record from any spelling the app already treats as the same. "Soby" finds Søby, "Aeroskobing" finds Ærøskøbing, "Lask" finds Łask, "Oeuilly" finds Œuilly and "Strasse" finds Straße, in every kind of search document; "Søby" finds a place written Soby; and "Tom" finds a person the reconciler shows as Thomas. Joining a Subject onto a record whose shown values do not change (a second "City" toponym onto the 6,000-event city) no longer reprojects every dependent's search document. And FakeStore's search stops ranking: it returns seeded hits in seeded order for the requested kinds, so ranking is tested once, in Go.

**Why now.** Bar outcome 20. B-21 is a regression from S11-07 (`distinctAlternates` began dropping alternates that SQLite's tokenizer cannot fold). The nickname gap is a regression from S11-28 (the fold hides the nickname record's spelling from the document). B-10 measured 6.06 s for one join on the large fixture. B-7 leaves a second ranking engine in the fake that nothing pins to Go.

**What stays.** The ranking formulas (`scoreFields`, kind priority, context and ref boosts); match reasons and snippets for every hit that matched before; the trigram typo shortlist; which documents exist; the graph cache's display drops on commit (`SummaryDeps` keeps the full walk); FakeStore's error hook (`searchCatalogError`) and catalog-session hold; every Swift view.

## 2. User stories

**US-1.** As a researcher, when I type a spelling without æ, ø, ł, œ, ß, đ, þ or ı, I find the record written with them, and the reverse, so that the keyboard I have does not decide what I find.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Place records titled Ærøskøbing, Søby, Łask, Œuilly and Straße; queries "Aeroskobing", "Soby", "Lask", "Oeuilly", "Strasse" (Places kind) | Each query returns its place as the first hit, match reason `title` |
| AC-1.2 | A place whose toponyms are Søby (title) and Soby (deduped alternate); queries "Soby" and "Søby" | Both return the place |
| AC-1.3 | A place titled Soby; query "Søby" | Returns the place |
| AC-1.4 | A Source titled "Søby kirkebog", an Evidence graph named "Ærø household", a Source type labelled "Fødested", a metadata field labelled "Łódź parish"; queries "Soby", "Aero", "Fodested", "Lodz" (all kinds) | Each returns its record (kinds `source`, `evidence_graph`, `source_type`, `metadata_field`) |
| AC-1.5 | A place with toponyms Ville-Marie (title) and Villemarie; query "Villemarie" | Returns the place: a spelling that differs only by a word break stays in the document |
| AC-1.6 | A Source titled "Montréal parish register"; queries "Montreal" and "Montréal" | Same hit and same score as today |

**US-2.** As a researcher, when a person's Tom record folded into Thomas, typing "Tom" finds Thomas, so that the reconciler's fold does not hide the person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A person recorded as "Tom Smith" and "Thomas Smith", shown as "Thomas Smith"; query "Tom" (Persons) | One hit titled "Thomas Smith", match reason `other`, empty snippet (no match line) |
| AC-2.2 | A person recorded only as "Thomas Smith"; query "Tommy" | The person is found: the document carries every name that folds into Thomas under the product seed (thom, thos, tom, tommy) |
| AC-2.3 | A person shown as "William Smith"; query "Robert" | No hit: robert does not fold into william |
| AC-2.4 | A person shown as "Tom Smith" | Its document has the line `alias:\ttommy` and no `alias:\tthomas`: in the product seed thomas does not fold into tom (the seed's direction keeps thomas) |

**US-3.** As a researcher, joining a Subject onto a big place is quick when the place's shown values do not change.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A birth Event located at place "York", its person; both documents overwritten with a marker; join a second place Subject with toponym "York" onto the place | Both documents still hold the marker: the join reprojects only the place |
| AC-3.2 | Same, then join a place Subject with toponym "Eboracum" | Both documents are rewritten: "Eboracum" finds the Event and the Person; the marker finds nothing |
| AC-3.3 | The large fixture (city with 6,000 events); join a place Subject with toponym "City" onto the city | Under `budgetJoinOntoCity` (1 s); B-10 measured 6.06 s |

**US-4.** As a contributor, FakeStore's search returns what a test seeds, so a Swift test that needs an order scripts it and ranking stays tested in Go.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `seededSearchHits` = [source s1, person p1, place l1]; kinds [] | [s1, p1, l1] |
| AC-4.2 | Same seed; kinds ["person", "place"] | [p1, l1] |
| AC-4.3 | Same seed; kinds ["nonsense"] | [] |
| AC-4.4 | Same seed; query "   " | [] and the catalog session is held |
| AC-4.5 | `searchCatalogError` set | Throws it |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration `000060`: a `folded` column on `catalog_search_docs` and on `catalog_search_fts` (B-21) | STEP-1 |
| One fold rule, `searchindex.FoldedSpelling`, used by documents, alternates and queries | STEP-1, STEP-2, STEP-3 |
| `searchindex.Upsert` / `Delete` write the folded words of every document kind; `ProjectionVersion` 10 | STEP-1, AC-1.1, AC-1.4 |
| `core/search` retrieval: a query word with a fold letter also matches its folded spelling; bm25 weighs the new column | STEP-2, AC-1.3, AC-1.6 |
| `distinctAlternates` dedupes at a key that keeps word breaks, so it drops no spelling FTS cannot match | STEP-3, AC-1.2, AC-1.5 |
| `AliasIndex.FoldsInto` and `valuecompare.NameFoldsInto`; the product seed's fold counts pinned | STEP-4, AC-2.2, AC-2.3 |
| A Person document carries `alias:` lines for the names that fold into its shown name | STEP-5, AC-2.1 – AC-2.4 |
| `RecomputeReported` reprojects a root's dependents only when the root's displayed rows changed (B-10) | STEP-6, AC-3.1, AC-3.2 |
| Budget subtest for the join onto the city | STEP-7, AC-3.3 |
| FakeStore search returns seeded hits; its ranking code goes (B-7) | STEP-8, AC-4.1 – AC-4.5 |
| Swift tests that tested the fake's ranking go or seed hits | STEP-9 |
| `add-searchable-kind` drops its `searchPriority` line and states the `folded` rule; domain doc | STEP-10 |

| Out | Home |
| --- | --- |
| The search registry's columns and projector table | Descoped with S12-36, not planned |
| Ranking changes (weights, kind priority, boosts) | Not planned |
| Composing reprojected dependents from the warm graph (the map's B-10 second half) | Not planned: D-9; the map's scope entry is corrected in STEP-11 |
| Narrowing `SummaryDeps` (the graph cache's display drops) to moved roots | Not planned: D-10 |
| Nicknames of a Person's other recorded names, or of its alternates | Not planned: D-7 |
| The catalog's own alias rows (user or plugin tables) in search documents | Not planned: D-6 |
| A match line for a nickname hit | Not planned (user, 2026-11-02) |
| Typo matching of folded spellings in the trigram shortlist | Not planned: D-3 |
| `core/search`'s own `valueWords` variable | Not planned: it still keys rescoring; folding goes through `searchindex.FoldedSpelling` |
| Accent folding before initials and nicknames in matching (B-19); NFC and compatibility forms (B-20) | Not scheduled in the audit; [`app-health-audit-2.md`](../app-health-audit-2.md) |
| Length caps on catalog text (audit L-10) | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| FTS schema | `core/database/migrations/000018.sql` | `catalog_search_docs` (title, ref, secondary, body) and `catalog_search_fts` over them, `tokenize='unicode61 remove_diacritics 2'`, `content='catalog_search_docs'`. `remove_diacritics` strips combining marks only, so æ, ø, ł, œ, ß, đ, þ, ı stay as written; unicode61 also splits a word at punctuation ("O'Brien" → o, brien) |
| Trigram schema | `core/database/migrations/000019.sql` | `catalog_search_fts_trigram` over the same four columns. Untouched by this PR |
| Latest migration | `core/database/migrations/000059.sql` | The next step is `000060.sql`. `000020.sql` and `000056.sql` added docs columns with a `ProjectionVersion` bump |
| Write path | `core/database/searchindex/searchindex.go` `Upsert`, `Delete` | `Delete` reads title, ref, secondary, body back and issues the FTS `'delete'` rows; `Upsert` inserts docs, then both FTS tables. Every document kind goes through `Upsert` |
| Tags | `searchindex.go` `BodyTagNote`, `BodyTagMetadata`, `BodyTagFilename`; `handles.go` `tagged(…, "name")`, `tagged(…, "toponym")`; `core/search/registry.go` `alternateMatch` literals `"name:\t"`, `"toponym:\t"` | Body tags are constants; secondary tags are literals on both sides |
| Version | `searchindex.go` `ProjectionVersion` = 9 (line 36) | `EnsureCatalog` → `NeedsRebuild` → `RebuildAll` when the stored version lags |
| Handle docs | `core/database/searchindex/handles.go` `matchText` | Person: title = summary `Name.Form`, secondary = life lines + `name:` alternates. Place: title = first toponym, `toponym:` alternates, parents. Event: list title and date / place lines |
| Alternate dedupe | `handles.go` `distinctAlternates` | Keeps a value whose `TextKey(…, Equivalent)` is new. The key folds the `Folds` letters and joins words with no separator, so "Soby" is dropped beside "Søby" and "Villemarie" beside "Ville-Marie", though FTS matches neither from the other (B-21) |
| Word rules | `handles.go` `valueWords` = `matchregistry.Default().Value.Words`; `core/search/registry.go` `valueWords` (same value) | `Folds` (`core/matchregistry/value.go`): ß→ss, æ→ae, œ→oe, ø→o, ł→l, đ→d, þ→th, ı→i |
| Word helpers | `core/valuecompare/words.go` `NormalizeForm`, `FoldWord`, `SplitWords`, `rawTokens` (unexported), `wordRune` (unexported) | `rawTokens` splits at whitespace, dashes and slashes and keeps other punctuation. `NormalizeForm` lowercases and drops what `wordRune` drops (punctuation that is not a break). `FoldWord(w, nil)` strips accents only (what `remove_diacritics 2` does); `FoldWord(w, Folds)` also applies the table |
| Retrieval | `core/search/fts.go` `buildMatchQuery`, `sanitizeFTSToken`, `mergeFTS` | `"tok"` per sanitized token, `*` on the token at index `len(tokens)-1` (even when an earlier token sanitizes away), joined by AND or OR; `mergeFTS` ranks with `bm25(catalog_search_fts, title, ref, secondary, body)` weights from `FTSBM25Weights` (`core/search/registry.go`). `core/search` already imports `searchindex` |
| Rescoring | `core/search/registry.go` `newQueryTokens`, `scoreFields`, `alternateMatch` | Keys fields and tokens at `Equivalent`, so it already folds both sides; it never sees a document FTS did not retrieve. `alternateMatch` reads only `name:` and `toponym:` lines; other lines leave reason `other` |
| Swift match line | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `showsMatchContext` | Draws a line only for notes, metadata, filename, description, name, place. `other` draws nothing |
| Aliases | `core/valuecompare/aliases.go` `AliasIndex.Expands`, `SetKeys`, `AliasToken` | `Expands(from, into, role)` is the reconciler's fold rule. Nothing lists the names that fold into a name |
| Name words | `core/valuecompare/name.go` `nameWords`, `tooManyWords`; `shortlist.go` `name.wordKeys` | `nameWords` is the one read of a name's words and roles (pattern, ignored, generation, untyped); `wordKeys` already walks it with `AliasToken` + an `AliasIndex` lookup and the `tooManyWords` cap |
| Reconciler fold | `core/autoreconcile/names.go` `nameModule.fold` | Reads `m.rules.Names.Western.Role(typ)` and `Aliases.Expands(p.A.Text, p.B.Text, role)` |
| Product rules | `core/database/valuerules/valuerules.go` `Product` | Registry value rules + `namealiases.ProductIndex()` (not memoized): the reconciler cache's rules. No import of `searchindex` (no cycle) |
| Product seed | `core/database/namealiases/nicknames_seed.go`, `ProductTables` | Two tables (abbreviations, 8 sets; nicknames, 1,229 sets), `FormalFirst`, roles given / nick / untyped, with directions for pairs listed both ways (thomas / tom). Over every token: 2,627 `Expands` pairs into 1,157 names, at most 15 into one name (margaret, marietta); `tom` ← {tommy} only |
| Summary name | `core/database/conclusionsummaries/conclusionsummaries.go` `PersonSummary.Name` | `*namevalue.Value` with typed `Parts` from the reconciled value. A name with no parts has no words for `nameWords` |
| Reprojection | `core/database/autoreconciler/autoreconciler.go` `RecomputeReported`, `recomputeBatch` | `recomputeBatch` deletes a batch's rows and inserts `res.Values` per group. `RecomputeReported` walks `conclusionsummaries.Dependents` from ids + association endpoints, reprojects all of them, returns the dependents for `SummaryDeps`. A join that changes no shown value still reprojects the city and its 6,000 events (B-10: 6.06 s, `RowsForSearch` 4.41 s through a one-shot `graphcache.Over`) |
| Row shape | `core/database/graphcache/verify.go` `TruthRow`, `valueEqual`; `core/database/autoreconciler/truth.go` `truthRow` | `TruthRow` is the one shape of a value row; `truthRow` builds it from a `ReconciledValue`; `valueEqual` compares a cached `Value` to it, `Support` and `Against` included |
| Callers | `core/writes/run.go` `Run`; `core/database/subjects/subjects.go` (`RecomputeTx`); `graphcachetest` | `Run` stores the returned dependents as `set.SummaryDeps`, `SummaryDepsWalked = true`; `graphcache` `dropCommittedDisplays` drops roots + those |
| Dependents' text | `core/database/searchindex/handles.go` `personLifeMatch`, `eventMatchSecondary`, `eventMatchTitle`; `conclusionsummaries` `toponyms`, `FirstPlaceName` | Read only cached value rows (kept toponyms, names, dates, types) of the records they name; never a label, ref or support count |
| Large fixture | `core/database/promotealign/perf_test.go` `largeFixture`, `newLargeFixture`, `TestLargeSourceBudgets` | Fields `a`, `onePlace`, `other`, `otherA`; the city id is a local, `cityID` (line 57). `f.place`, `f.join` in `promotealign_test.go` |
| FakeStore search | `macos/App/Platform/FakeStore.swift` `searchCatalog`, `handleSearchKinds`, `searchPriority`, `orderedByMixedScore`, `handleSearchHits`, `handleSearchSummary`, `handleSearchMatchText`, `scoreRefRecord`, `vocabularySearchHits`, `classifyRefQuery`, `tokenizeSearch`, `scoreSearchFields` | Scores Sources, graphs, vocabulary and handles in memory, mirrors Go's kind mix and ref boosts, caps at 50. Each helper's only caller is in this list. The state they read (`sourcesByProject`, `membershipBySubject`, `seededEventSummaries` and the rest) has other readers and stays |
| Search guidance | `.agents/skills/add-searchable-kind/SKILL.md` line 32, §3 step 1 | Line 32: "until S12-50 removes it, for FakeStore's `searchPriority`". §3 step 1 tells a projector to keep both FTS tables in sync; nothing on `folded`. The rule (`.cursor/rules/searchable-kinds.mdc`) and line 33 already say the fake returns seeded hits |
| Go tests | `core/search/handles_test.go` `TestAlternateToponymIsMatchContext`, `TestSummarySearchFindsTheListTitle`; `core/search/search_test.go` `TestFoldedQueryScoresLikeExact`, `TestSearchAccentedPlaceTypo`, `runSourceTypeCreate`, `runFieldCreate`; `core/database/searchindex/searchindex_test.go` `TestMatchTextListsEachAlternateOnce` | Montréal-only folding; no ø / æ / ł / œ / ß case; no nickname case |
| Swift tests | `macos/ProvenenciaTests/SearchCatalogStoreTests.swift` (all, plus `robinsSearchStore`); `CatalogSearchTests.swift` `exactRefHitIsTop`, `kindsRestrictHits`, `noLocationGivesNoSectionNudge`; `OmnibarResultsModelTests.swift` `debouncedSearchReturnsHits`, `activateFieldHitFromSearch`, `activateEvidenceGraphHitOpensTheGraph`; `PromoteModelTests.swift` `aFoundRecordIsSearchedPickedAndHeld` | Each seeds Sources, vocabulary or memberships and relies on the fake matching the query text. Other tests that set `searchCatalogError` (`PromoteModelTests` line 370, `OmnibarResultsModelTests` `searchFailureSurfacesMappedErrorNotEmpty`) are unaffected |

The map's Context names `conclusionheaders.RowsForSearch`; it is `conclusionsummaries.RowsForSearch` since `1d475bde`.

**Precedent.** `a26a944c` (S12-15, #583): a `catalog_search_docs` column in `000056.sql`, `ProjectionVersion` 8 → 9 with a comment line, and `Upsert` writing it. Copy the migration header comment style and the version comment. `4ccde944` (S11-07): `distinctAlternates` and `TestMatchTextListsEachAlternateOnce`; extend that test's table. `shortlist.go` `name.wordKeys`: the shape of `NameFoldsInto`. #617 (S12-35): FakeStore turned from rules into seeded rows and scripted errors; copy its shape (a `seeded…` property, the `…Error` hook, the session hold).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Folded spellings go in a new FTS column, `folded`, on `catalog_search_docs` and `catalog_search_fts`, written by `Upsert` for every document kind | The user chose every kind (2026-11-02). Every kind's text columns are read by `scoreFields` as named fields (description, notes, key), so folded text in them would score and snippet as that field. One column written in `Upsert` serves every projector, present and future | A tagged `fold:` line in `secondary` — handle-only, and a Source's secondary is its description field; dedupe at a non-folding key only — leaves "Soby" unable to find a lone "Søby" |
| D-2 | One rule, `searchindex.FoldedSpelling(raw string) (string, bool)` *(new, exported)*: `raw` is one token as `valuecompare.RawTokens` splits it. Let `w = NormalizeForm(raw)` and `f = FoldWord(w, valueWords.Folds)`. It returns `(f, true)` when `f != FoldWord(w, nil)` (a fold letter), or when `raw` holds a rune that `NormalizeForm` drops (not a break) with a kept rune on both sides (FTS splits there: "O'Brien" → `obrien`, "1,200" → `1200`). Otherwise `("", false)`: "St." (trailing "."), "Montréal", "soby". `folded` holds, space-joined and de-duplicated in first-seen order, every `f` of every `RawTokens` token of the document's title, secondary and body. Ref is skipped (Crockford ASCII). Empty when no token qualifies | Exactly the tokens that unicode61 with `remove_diacritics 2` indexes differently from rescoring's key. Montréal adds nothing, so its bm25 and score stay as today (AC-1.6). One function serves the document, the alternate key and the query, so the two sides cannot disagree | Every word folded — doubles every document and moves every bm25 rank; NFD + Mn-strip via `x/text` in a second helper — `FoldWord(w, nil)` already is that |
| D-3 | The trigram table keeps its four columns | It is the typo shortlist; the JW gate reads identity fields. Folded typos are not in the bar | Adding `folded` there too — a second table rebuilt for no AC |
| D-4 | `buildMatchQuery` calls `searchindex.FoldedSpelling` on each token after `sanitizeFTSToken`. When it reports a spelling, the token is written `("tok" OR "f")`, with `*` on both halves when the token is the one that takes `*` today. Tokens with no folded spelling, and the `*` placement (index `len(tokens)-1`, even when that token sanitizes away), are as today | Query "Søby" must reach a doc written "Soby" (AC-1.3), which has no `folded` word. A sanitized token has lost its apostrophe already ("o'brien" → "obrien"), which matches the document's `folded` word | Folding the query only — loses a doc whose own spelling is the unfolded one in a column FTS reads raw |
| D-5 | `FTSBM25Weights` gains `Folded` = 4 (the Secondary weight); `mergeFTS` passes five weights in table column order | The folded words come from title, secondary and body mixed; the lower weight keeps a folded-only match from out-ranking the same doc's direct title match. `scoreFields` still decides the field and most of the score | Title weight 10 — lets body words ride a title weight |
| D-6 | Nickname expansions read the product alias seed: `searchindex` holds one `productRules = valuerules.Product()`, and `valueWords` becomes `productRules.Words` | It is the index the reconciler folded with (`autoreconciler` `ValueRules`), so search restores exactly what the fold hid. The document changes only with `CacheVersion` / `ProjectionVersion`, not with a catalog alias edit (which would otherwise need every Person reprojected) | Catalog rows via `valuerules.Load` — a second invalidation path for no AC |
| D-7 | A Person document carries, after its `name:` alternates, one `alias:\t<token>` line per token of `valuecompare.NameFoldsInto(Value{Type: name, Name: p.Name}, productRules)` *(new)*: for each word of `nameWords` whose role is in `AliasRoles()`, `AliasIndex.FoldsInto(AliasToken(word), role)`; de-duplicated, sorted, without the name's own words; nil when `tooManyWords`. Only the shown name, not alternates. This expansion stays after S14-26: once the recorded nickname is a `nick` part of the title, the alias lines still find "Tommy" for a Thomas no record calls Tommy (AC-2.2) | The user chose fold-in only (2026-11-02): the reconciler's own rule, so "Robert" never finds William and "Rose" never finds Mary. `nameWords` is the one read of a name's words and roles; `wordKeys` already walks it the same way | Every set-mate (user declined); a word walk in `searchindex` — a second read of a name's roles |
| D-8 | Secondary tags become constants beside the body tags: `SecondaryTagName = "name:\t"`, `SecondaryTagToponym = "toponym:\t"`, `SecondaryTagAlias = "alias:\t"` *(new)*. `tagged` takes the constant; `alternateMatch` parses name and toponym by constant and does not read `SecondaryTagAlias`, so a nickname hit keeps reason `other` and an empty snippet; Swift draws no match line for `other` | The user chose no match line (2026-11-02). A constant makes the one tag that is not read visible at the parse site | A `nickname` reason with a new L10n line (user declined) |
| D-9 | B-10 is fixed by not reprojecting dependents of a root whose displayed rows did not move. Reprojected dependents still compose through `graphcache.Over(q)` | The warm graph holds committed state; inside the write it would compose a dependent of a moved root from the root's old names, which is the wrong document. With the filter, the join onto the city reprojects the city alone | Composing from `c.Graph()` inside the transaction — stale reads |
| D-10 | A root moved when its displayed rows differ. `recomputeBatch` returns the ids of its batch that moved: before its `DELETE` it reads the batch's stored rows into `[]graphcache.TruthRow` per entity (one query per batch, `storedRows` *(new)* in `autoreconciler/truth.go`), and it builds the new rows with `truthRow` from the `res.Values` it already holds. Per entity, rows ordered by `(PropertyID, Rank)` are compared with `graphcache.SameDisplay(a, b TruthRow) bool` *(new)*: every field but `Support` and `Against`. An entity moved when the row counts differ or any pair is not `SameDisplay`. The Dependents walk for reprojection starts from the moved ids plus the association endpoints (before and after) as today; `SummaryDeps` keeps the full walk from every root | Dependents' search text reads only those rows (§4 "Dependents' text"). `Support`/`Against` are the vote counts a join moves. `TruthRow` is the one shape of a value row, so a new value column reaches both Verify and this filter. Keeping the full walk for `SummaryDeps` leaves the graph cache's display drops exactly as today, since a display row can show a root's label | A separate column list and hash per entity — a fifth hand-kept list of the row's columns; comparing composed summaries — the cost B-10 is about; narrowing `SummaryDeps` — display rows read the entity row (label) |
| D-11 | `valueEqual` becomes `SameDisplay(rowOf(v), t) && v.Support == t.Support && v.Against == t.Against`, with `rowOf(v Value) TruthRow` *(new, unexported)* in `verify.go` | One definition of value-row equality | Leaving `valueEqual` beside `SameDisplay` — two lists of the same fields |
| D-12 | FakeStore: `var seededSearchHits: [CatalogSearchHit]` *(new)*; `searchCatalog` is, inside `withState`: throw `searchCatalogError` if set; `markCatalogSessionHeld(projectDir)`; return [] for a whitespace-only query; else the seeded hits whose `searchKind` is in the requested kinds (all kinds when the list is empty; an unknown kind matches nothing), in seeded order, uncapped | The working rule and `.cursor/rules/searchable-kinds.mdc` line 16 ("Empty / whitespace queries return no hits; FakeStore returns seeded hits in seeded order for the requested kinds"). Go caps at the request limit; the fake returns what a test seeded | Filtering seeds by query text — a matching engine again |
| D-13 | Swift tests that tested the fake's ranking or matching are deleted; tests of a model's behavior seed `seededSearchHits` with the hits the engine would return (T-14 lists the fields) | `.cursor/rules/swift-tests.mdc`: engine rules are tested in Go (`core/search` has the ref, prefix, kind-mix and section-nudge tests) | Keeping them against seeds — they would test nothing |
| D-14 | One PR, over plan-pr's size limit (12 steps, Go and Swift) | The user, 2026-11-02, chose to keep one PR | Splitting into three PRs (recommended) |
| D-15 | `valuecompare` exports `RawTokens` (the existing `rawTokens`, renamed; its in-package callers follow) | `FoldedSpelling` and `alternateKey` need the same raw split `SplitWords` uses; `add-match-rule` forbids a second copy of a word helper | A raw split in `searchindex` — a copy |

**Rules this plan bends.** plan-pr **Size** ("Steps go past about 10": propose a split) → the user answered 2026-11-02: keep S14-21 as one PR (D-14). No rule text changes; the map row keeps one PR.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000060.sql` *(new)* | migration | `000056.sql` (docs columns) | Same shape plus the FTS table rebuild an FTS column needs |
| `catalog_search_docs.folded`, `catalog_search_fts.folded` *(new)* | column | `secondary` | `secondary` is a scored field for every kind (D-1) |
| `searchindex.FoldedSpelling` *(new)* | func | `valuecompare.TextKey`, `FoldWord`, `NormalizeForm` | Composes them; "what FTS indexes differently" is the search projector's rule, shared by the document and the query (D-2) |
| `searchindex.foldedWords(texts ...string) string` *(new)* | func | — | Collects `FoldedSpelling` over a document's tokens |
| `searchindex.alternateKey(s string) string` *(new)* | func | `valuecompare.TextKey(…, Equivalent)` | That key joins words with no separator (it must, for compare); this one keeps the breaks FTS keeps |
| `valuecompare.RawTokens` (renamed from `rawTokens`) | func | `rawTokens` | Exported, not copied (D-15) |
| `searchindex.SecondaryTagName`, `SecondaryTagToponym`, `SecondaryTagAlias` *(new)* | const | `BodyTagNote`, `BodyTagMetadata`, `BodyTagFilename` | Same pattern, for secondary lines |
| `FTSBM25Weights.Folded` *(new)* | field | `FTSBM25Weights` | Extended in place |
| `valuecompare.AliasIndex.FoldsInto(into string, role NameRole) []string` *(new)* | method | `AliasIndex.Expands`, `SetKeys` | `Expands` answers one pair; nothing lists the set. `FoldsInto` filters set-mates through `Expands`, so it cannot disagree with it |
| `valuecompare.NameFoldsInto(v Value, r Rules) []string` *(new)* | func | `shortlist.go` `name.wordKeys` | Same walk (`nameWords`, `AliasToken`, alias lookup, `tooManyWords`), different lookup |
| `searchindex.productRules` *(new)* package var | var | `valueWords` | `valueWords` becomes `productRules.Words`: one rules value in the package |
| `graphcache.SameDisplay(a, b TruthRow) bool` *(new)*, `rowOf` *(new)* | func | `valueEqual` | `valueEqual` is rewritten on it (D-11) |
| `autoreconciler.storedRows(q, ids) (map[string][]graphcache.TruthRow, error)` *(new)* | func | `truthRows` (recomputed rows), `graphcache.loadValues` (unexported, into `Value`) | Reads stored rows into the shared shape; `loadValues` is a `Graph` method into a different type |
| `largeFixture.city` *(new)* field, `budgetJoinOntoCity` *(new)* | test | `largeFixture.onePlace`, `budgetOnePlaceWarm` | Extends the fixture and budget block |
| `(f *handleFixture) locatedBirth(…)` *(new)* | test helper | `TestSummarySearchFindsTheListTitle` body | Extracted from it; both tests use it |
| `FakeStore.seededSearchHits` *(new)* | Swift var | `seededEventSummaries`, `seededPlaceSummaries` | Same seeded-row pattern |

**Builds on.** `searchindex.Upsert` as the one write of every document; `ProjectionVersion` heal-on-open; `matchregistry` `Folds` as the one fold table; `AliasIndex.Expands` as the one fold rule; `nameWords` as the one read of a name; `graphcache.TruthRow` as the one row shape; S12-35's seeded FakeStore.
**Deletes.** FakeStore: `handleSearchKinds`, `searchPriority`, `orderedByMixedScore`, `handleSearchHits`, `handleSearchSummary`, `handleSearchMatchText`, `scoreRefRecord`, `vocabularySearchHits`, `classifyRefQuery`, `tokenizeSearch`, `scoreSearchFields` and the `prefix(50)` cap. Swift tests: `SearchCatalogStoreTests` `fakeSearchFoldsLikeGo`, `findsSourceByTitleAndMapsLocation`, `findsEvidenceGraphByNameRefAndDescription` (with `seedCensusGraphs`), `findsTypeAndFieldByLabel`, `findsSourceByRef`, `prefixRefPromotesSource`, `kindPriorityIsATenthAndSectionNudgeStillMatters`, `searchWithoutLocationHasNoSectionNudge`, and the file-level `robinsSearchStore`; `CatalogSearchTests` `noLocationGivesNoSectionNudge`. Guidance: `add-searchable-kind` line 32's `searchPriority` clause. Go: `distinctAlternates`'s `TextKey` call (replaced by `alternateKey`); the `"name:\t"` / `"toponym:\t"` literals (replaced by constants); `valueEqual`'s field list (replaced by `SameDisplay`).
**Next adds.** A new search kind gets folded spellings with no code (its projector calls `Upsert`; `add-searchable-kind` says so). A new fold letter is one `Folds` entry, and documents and queries follow through `FoldedSpelling`. A new value column is one `TruthRow` field, covered by Verify and the reprojection filter. S14-26 (a folded nickname shows as a `nick` part) changes the shown name's parts; `NameFoldsInto` reads a `nick` word's role (`RoleNick`, in `AliasRoles`) with no change.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A document found only through `folded` or an `alias:` line | `core/search` `Search` scoring branch | `scoreFields` keys every field at Equivalent, so a folded hit scores on its real field (title for AC-1.1); an alias hit scores `other` and falls through `alternateMatch` with `field == ""`, keeping reason `other` (no new branch) |
| A root that moved nothing | `RecomputeReported` | The root itself is always reprojected; only its dependents are skipped |
| A whitespace query or an unknown kind on FakeStore | `searchCatalog` | Returns [] (D-12) |

### Scale

At the map's Scale (6,000 events on one place; 37,000 Subjects): a join onto the city that moves no displayed row reprojects the city and any association endpoints only, not its 6,000 events and their people. `SummaryDeps` is unchanged, so `graphcache` drops the same displays as today (B-10 measured 84–98 ms for that commit). The before-read adds one batched `SELECT` per recompute batch; the after rows are already in memory. A join that renames the city still reprojects all dependents, as today (5.6 s on the fixture); that case is a rename, measured in B-10, and not in the bar. `foldedWords` is linear in a document's text and runs once per `Upsert`. `RebuildAll` on first open after the update rewrites every document once (the heal every `ProjectionVersion` bump costs). Budget: `budgetJoinOntoCity` = 1 s, against 6.06 s today; B-10's profile puts everything outside `ReprojectHandles` at 0.13 s on a Mac, and `budgetAfterCommit` notes CI runs about 3× slower.

### Edges

- **Assumptions.** (1) `remove_diacritics 2` folds exactly what `FoldWord(w, nil)` folds: both strip combining marks after canonical decomposition. Held by T-1 rows (Montréal → no folded word; Søby → `soby`). (2) unicode61 splits a token at a mid-word dropped rune; T-1 rows "O'Brien" → `obrien`, "St." → nothing. (3) `distinctAlternates` still keeps the first of each key in order and the title first; `TestMatchTextListsEachAlternateOnce` rows stay. (4) Dependents' search text reads only the roots' value rows (§4); a label change on a root with no row change does not move a dependent's document. T-7 holds the join; `TestRebuildEqualsUpkeep_*` hold upkeep against rebuild. (5) `FoldsInto` agrees with `Expands` for every pair: T-5 checks every product token. (6) `storedRows` and `truthRow` order rows the same way: both sorted by `(PropertyID, Rank)` before comparing.
- **Neighbors.** Folding: Montréal (no change), "Ville Marie" vs "Ville-Marie" (still one alternate), "Villemarie" (now kept, AC-1.5), "Straße"/"Strasse" (`strasse`), uppercase Ø (lowercased first), "St." (nothing), "1,200" (`1200`), a digit-only word (never folded). Query: "Søby" OR-expanded; "soby" unchanged; "Aero" is the last token, so its plain prefix `"aero"*` reaches `folded` `aero`; a ref-shaped query skips FTS as today. Nicknames: Tom → Thomas found; Thomas → Tom not through aliases (AC-2.4); Robert → William not found (AC-2.3); a name with no parts, or only a surname, gets no lines; a `nick` part (S14-09) expands as a given. Reprojection: a support-only change skips dependents (AC-3.1); a new toponym, an outvoted value turning kept (reason changes), a removed value, a merge that empties the handle's rows: each moves the root and walks.
- **Removed inputs.** A deleted or merged root loses its rows; the empty row list differs from its old one, so its dependents are reprojected as today. A root with no rows before and after moves nothing its dependents read. Alias lines come from the shown name only: when the Tom record is excluded or deleted, Thomas keeps its expansions (they come from the seed, not from the Tom record), which is D-7's rule.
- **Real data.** The product seed: 2,627 fold pairs into 1,157 names; most into one name: 15 (margaret, marietta); thomas ← {thom, thos, tom, tommy}; tom ← {tommy}; william ← {bill, billy, wil, will, willie, willy, wm}; john ← {jack, jno, jock, johnny, jonnie, jonny}; robert ← {bill, billy, bob, bobby, dob, dobbin, hob, hobkin, rob, robby, robt, rupert}. Worth a look: "bill" folds into both william and robert, so "Bill" finds both, which the reconciler would also fold. T-5 pins the 2,627, the 15, thomas and tom.
- **Untrusted size.** A shown name's words are capped by `tooManyWords` (`Names.MaxWords` = 32) before any lookup; at most 15 tokens per word, so ≤ 480 lines. `foldedWords` is linear in the document text, de-duplicated by a map. `storedRows` reads rows the engine wrote.

### Schema

Migration `core/database/migrations/000060.sql` *(new)*:

```sql
-- S14-21: the words of a search document that SQLite's tokenizer indexes
-- differently from rescoring's key (letters remove_diacritics cannot fold, and
-- words it splits at punctuation). Written by searchindex.Upsert for every
-- kind; ProjectionVersion 10 fills it on open. An FTS5 table cannot add a
-- column, so catalog_search_fts is rebuilt empty here and RebuildAll refills it.

ALTER TABLE catalog_search_docs ADD COLUMN folded TEXT NOT NULL DEFAULT '';

DROP TABLE catalog_search_fts;
CREATE VIRTUAL TABLE catalog_search_fts USING fts5(
	title,
	ref,
	secondary,
	body,
	folded,
	content='catalog_search_docs',
	content_rowid='rowid',
	tokenize='unicode61 remove_diacritics 2',
	prefix='2 3'
);
```

`catalog_search_docs` stays `BucketSkip` in `catalogmodel`; no FK. No audit. Schema hash: computed at init (`add-catalog-migration` § Schema hash); no constant to edit. Between the step and `RebuildAll`, nothing writes the index: `core/catalogsession/session.go` runs `autoreconciler.EnsureCatalog` then `searchindex.EnsureCatalog` on open, before any researcher write.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` (`words.go`) | `rawTokens` → `RawTokens` (exported; doc comment kept; in-package callers renamed) |
| `core/valuecompare` (`aliases.go`) | `func (a AliasIndex) FoldsInto(into string, role NameRole) []string` *(new)*: every token of a set in `byToken[into]` that allows `role`, other than `into`, for which `a.Expands(token, into, role)`; sorted, no duplicates; nil for the zero index |
| `core/valuecompare` (`shortlist.go`, beside `WordKeys`) | `func NameFoldsInto(v Value, r Rules) []string` *(new)*: `words, generation := nameWords(v, r)`; nil when `tooManyWords`; for each word with `slices.Contains(AliasRoles(), w.role)`, `t, ok := AliasToken(w.text, r.Words)`, append `r.Names.Aliases.FoldsInto(t, w.role)`; drop the name's own tokens; sort; compact |
| `core/database/searchindex` (`searchindex.go`) | `SecondaryTag*` constants (D-8). `FoldedSpelling` (D-2) and `foldedWords`. `Document` gains no field: `Upsert` computes `folded := foldedWords(doc.Title, doc.Secondary, doc.Body)` and writes it to docs and `catalog_search_fts`; the trigram insert is unchanged. `Delete` selects and passes `folded` in the `catalog_search_fts` `'delete'` row. `ProjectionVersion` = 10, comment `10: a folded column for spellings FTS indexes differently, and a Person's nicknames that fold into its shown name (S14-21).` |
| `core/database/searchindex` (`handles.go`) | `var productRules = valuerules.Product()` *(new)*; `var valueWords = productRules.Words`. `tagged(vals, tag)` takes a `SecondaryTag*` constant and writes `tag + v`. `distinctAlternates` keys with `alternateKey(s)` *(new)*: `strings.Join` of `FoldWord(NormalizeForm(raw), valueWords.Folds)` over `valuecompare.RawTokens(s)`, skipping empty words, one space; an empty key is dropped as today. `matchText` Person case appends `tagged(valuecompare.NameFoldsInto(valuecompare.Value{Type: propertymeta.ValueTypeName, Name: p.Name}, productRules), SecondaryTagAlias)` after the `name:` lines, when `p.Name != nil` |
| `core/search` | `FTSBM25Weights.Folded` = 4 *(new)*; `mergeFTS` passes `w.Folded` as the fifth bm25 weight. `buildMatchQuery`: D-4. `alternateMatch` parses `searchindex.SecondaryTagName` and `SecondaryTagToponym` |
| `core/database/graphcache` (`verify.go`) | `SameDisplay(a, b TruthRow) bool` *(new)*: `PropertyID, Rank, Reason, HasText, Text, HasInteger, Integer, TermID, EntityID, Date, Name, ImageFileID` equal (bytes by `bytes.Equal`). `rowOf(v Value) TruthRow` *(new)*. `valueEqual` per D-11 |
| `core/database/autoreconciler` | `storedRows` *(new, `truth.go`)*: `SELECT entity_id, property_id, rank, reason, value_text, value_integer, value_term_id, value_entity_id, value_date, value_name, value_image_file_id, support, against FROM auto_reconciler_values WHERE entity_id IN (…) ORDER BY entity_id, property_id, rank`, into `map[string(entity_id)][]graphcache.TruthRow`. `recomputeBatch(q, ids) (moved [][]byte, err error)`: reads `storedRows` for the batch before the `DELETE`; collects each group's new rows with `truthRow` beside `insertRow`; returns the batch ids whose sorted row lists differ (count or a pair not `SameDisplay`). `RecomputeReported`: gathers `moved` from every batch; `deps := Dependents(roots)` as today for the return value; `reproject := deps` when `moved` holds every id, else `Dependents(moved ∪ before-endpoints ∪ after-endpoints)`; `ReprojectHandles(roots ∪ reproject)`; return `deps` |

Errors: none new.

### FFI

None — no proto or handler change; `MatchReason` keeps its values.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Seeded search | `FakeStore.searchCatalog` | `macos/App/Platform/FakeStore.swift` | D-12; no view changes |

States: the omnibar's existing empty, error and results states. No view, no accessibility change.

### Strings

None — no copy changes. The nickname hit uses the existing no-context row.

## 7. Steps

**STEP-1 — Catalog step 000060 and the folded words**
Files: `core/database/migrations/000060.sql` *(new)*, `core/valuecompare/words.go` (and its in-package callers of `rawTokens`), `core/database/searchindex/searchindex.go`, `core/database/searchindex/searchindex_test.go`. Skills: `add-catalog-migration`, `add-match-rule`.
Do: the DDL in §6 Schema, with its comment; `RawTokens` (D-15); `SecondaryTag*` constants; `FoldedSpelling` (D-2), `foldedWords`, `Upsert`/`Delete` carrying `folded`, `ProjectionVersion` 10 with its comment. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/... ./core/valuecompare/...` passes; no `catalog.schema_mismatch`.

**STEP-2 — Retrieve folded spellings**
Files: `core/search/fts.go`, `core/search/registry.go`, `core/search/search_test.go`, `core/search/handles_test.go`.
Do: `FTSBM25Weights.Folded`, five-weight `bm25`, the OR form in `buildMatchQuery` (D-4, D-5), `alternateMatch` on the tag constants. Add T-2, T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/search/...` passes, `TestFoldedQueryScoresLikeExact` unchanged.

**STEP-3 — Keep every spelling FTS cannot match**
Files: `core/database/searchindex/handles.go`, `core/database/searchindex/searchindex_test.go`.
Do: `productRules`, `valueWords = productRules.Words`, `tagged` on the constants, `alternateKey` and `distinctAlternates` on it (§6). Add T-6's place rows.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/ ./core/search/` passes; `go list -deps ./core/database/valuerules | grep -c core/database/searchindex` prints 0 (no cycle).

**STEP-4 — List the names that fold into a name**
Files: `core/valuecompare/aliases.go`, `core/valuecompare/shortlist.go`, `core/valuecompare/name_test.go`, `core/database/namealiases/namealiases_test.go`. Skill: `add-match-rule`.
Do: `AliasIndex.FoldsInto`, `NameFoldsInto` (§6). Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/ ./core/database/namealiases/` passes.

**STEP-5 — Carry a Person's folded-in nicknames**
Files: `core/database/searchindex/handles.go`, `core/database/searchindex/searchindex_test.go`, `core/search/handles_test.go`.
Do: the `alias:` lines in `matchText` (D-6, D-7, D-8). Add T-6's person rows and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/ ./core/search/` passes.

**STEP-6 — Reproject dependents only when a root moved**
Files: `core/database/graphcache/verify.go`, `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/truth.go`, `core/search/handles_test.go`.
Do: `SameDisplay`, `rowOf`, `valueEqual` (D-11); `storedRows`, `recomputeBatch` returning moved ids, the narrowed reprojection (D-9, D-10). Extract `locatedBirth(personForm, toponym string) (person, event, place subjects.Subject)` from `TestSummarySearchFindsTheListTitle` (the person, event and place Subjects with their participation and location bridges; promotion stays in the test) and call it there; add T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/ ./core/database/graphcache/... ./core/search/ ./core/writes/` passes, including `TestRebuildEqualsUpkeep_Scenarios`, `TestRebuildEqualsUpkeep_SeededSequences` and `TestLoaderQueryCountIsConstant`; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./core/database/promotealign/` passes.

**STEP-7 — Budget the join onto the city**
Files: `core/database/promotealign/perf_test.go`.
Do: add `city []byte` to `largeFixture` and set it from `cityID` in `newLargeFixture`'s return; add `budgetJoinOntoCity = 1 * time.Second` to the budget block with a comment citing B-10 (6.06 s before S14-21); add T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -v ./core/database/promotealign/` passes and logs the join time.

**STEP-8 — FakeStore returns seeded hits**
Files: `macos/App/Platform/FakeStore.swift`. Skill: `add-swift-test` (for STEP-9).
Do: D-12; delete the eleven helpers and the cap (§6 Deletes); the doc comment on `searchCatalog` is one line: "Seeded hits in seeded order for the requested kinds; ranks nothing (`.cursor/rules/swift-tests.mdc`)."
Check: `rg -n 'searchPriority|orderedByMixedScore|scoreRefRecord|scoreSearchFields|classifyRefQuery|tokenizeSearch|vocabularySearchHits|handleSearchHits|handleSearchSummary|handleSearchMatchText|handleSearchKinds' macos` prints nothing; the app target builds on macOS (`scripts/check-pr.sh`).

**STEP-9 — Swift tests seed hits**
Files: `macos/ProvenenciaTests/SearchCatalogStoreTests.swift`, `CatalogSearchTests.swift`, `OmnibarResultsModelTests.swift`, `PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: delete the tests in §6 Deletes; add T-10 – T-13; change the tests in T-14 to seed `seededSearchHits` with the hits listed there.
Check: `xcodebuild test` as `scripts/check-pr.sh` runs it passes with no Swift warning. On a host without macOS, this check and STEP-8's build cannot run: stop after STEP-8's `rg` and report, so the orchestrator runs them on macOS.

**STEP-10 — Guidance and domain doc**
Files: `.agents/skills/add-searchable-kind/SKILL.md`, `docs/conclusion-reconciliation.md`.
Do: line 32 ends at "`leadThumbnail`" (drop the `searchPriority` clause). §3 step 1 gains: "`Upsert` writes `folded` (spellings unicode61 indexes differently, from `searchindex.FoldedSpelling`) for every kind, and `core/search` OR-expands a query word the same way; a projector never writes FTS rows itself, and the trigram table keeps four columns." In `conclusion-reconciliation.md` §6, after "search reads them all, so an outvoted spelling still finds its Person": "A nickname the reconciler folded away finds it too: a Person's search document carries the names that fold into its shown name under the product alias seed (S14-21)."
Check: `python3 scripts/check-guidance-names.py` passes.

**STEP-11 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (every `MARKETING_VERSION` line).
Do: the next patch after `VERSION` at merge, in its own commit.
Check: `rg -n 'MARKETING_VERSION' macos/Provenencia.xcodeproj/project.pbxproj`, `cat VERSION` and `core/version.go` show the same version.

**STEP-12 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md` (new section at the top, under the intro), `docs/deployment-plan/spike-14/deployment-plan.md` (map row status; the S14-21 scope entry's Ships line), this plan (status, Landed note).
Do: per the Definition of done. The scope entry's Ships line says: a `folded` column (migration `000060`) written for every kind and a query that OR-expands fold letters; alternates deduped at a key that keeps word breaks; a Person's names that fold into its shown name; dependents reprojected only when a root's displayed rows moved (no warm-graph composing, D-9); FakeStore seeded search.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-21-search-finds-every-spelling.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/searchindex/searchindex_test.go` `TestFoldedSpelling` *(new)*, `TestFoldedWords` *(new)* | `FoldedSpelling`: "Søby" → ("soby", true); "Montréal" → ("", false); "Straße" → ("strasse", true); "O'Brien" → ("obrien", true); "1,200" → ("1200", true); "St." → ("", false); "soby" → ("", false). `foldedWords`: "Ærøskøbing Havn" → "aeroskobing"; "Straße Straße" → "strasse"; "St. Søby" → "soby"; ("Łask", "Søby", "Łask") → "lask soby"; "" → "" | D-2 |
| T-2 | Go | `core/search/handles_test.go` `TestFoldedSpellingsFindPlaces` *(new)* | Places Ærøskøbing, Søby, Łask, Œuilly, Straße (one fixture each), queries "Aeroskobing", "Soby", "Lask", "Oeuilly", "Strasse" → that place first, reason `title`; place Søby + toponym Soby → "Soby" and "Søby" both find it; place Soby → "Søby" finds it | AC-1.1, AC-1.2, AC-1.3 |
| T-3 | Go | `core/search/search_test.go` `TestFoldedSpellingsFindEveryKind` *(new)* | Source "Søby kirkebog" ← "Soby" (kind `source`); the Source's sole graph (`evidencegraphstest.SoleGraph`) renamed with `evidencegraphs.Update(tx, userID, graph.ID, evidencegraphs.UpdateInput{Name: "Ærø household"})` inside `writes.Run` ← "Aero" (kind `evidence_graph`); Source type "Fødested" (`runSourceTypeCreate`) ← "Fodested" (kind `source_type`); metadata field "Łódź parish" (`runFieldCreate`, data type "text") ← "Lodz" (kind `metadata_field`); Source "Montréal parish register" ← "Montreal" and "Montréal" give equal scores (`TestFoldedQueryScoresLikeExact` also stays) | AC-1.4, AC-1.6 |
| T-4 | Go | `core/search/search_test.go` `TestBuildMatchQueryAddsFoldedSpelling` *(new)* | ["søby"] → `("søby"* OR "soby"*)`; ["soby"] → `"soby"*`; ["søby","kirke"] AND → `("søby" OR "soby") AND "kirke"*`; ["montréal"] → `"montréal"*`; ["o'brien"] → `"obrien"*` | D-4 |
| T-5 | Go | `core/valuecompare/name_test.go` `TestAliasIndexFoldsInto` *(new)*, `TestNameFoldsInto` *(new)*; `core/database/namealiases/namealiases_test.go` `TestProductFoldsInto` *(new)* | Tiny index from `AliasTable{Key: "t", Kind: AliasNickname, Roles: AliasRoles(), FormalFirst: true, Sets: [][]string{{"thomas", "tom", "tommy"}}}`: FoldsInto(thomas, given) = [tom tommy]; FoldsInto(tom, given) = []; role family → []; zero index → nil. `NameFoldsInto` on that index: "Thomas Smith" (given, surname parts) → [tom tommy]; a 33-word name → nil. Product (`ProductIndex()`): every FoldsInto result `Expands` into its name; totals 2,627 pairs, max 15; thomas = [thom thos tom tommy]; tom = [tommy]; william holds no robert | D-7, AC-2.2, AC-2.3, AC-2.4 |
| T-6 | Go | `core/database/searchindex/searchindex_test.go` `TestMatchTextListsEachAlternateOnce` (extended) | Place toponyms Ville-Marie, Villemarie → secondary has `toponym:\tVillemarie`; place Søby, Soby → no Soby line (the folded column covers it); person "Thomas Smith" (`namevaluetest.Western`) → `alias:\tthom`, `alias:\tthos`, `alias:\ttom`, `alias:\ttommy` after the `name:` lines; person "Tom Smith" → `alias:\ttommy` only; existing rows unchanged | AC-1.5, AC-2.4, D-7 |
| T-7 | Go | `core/search/handles_test.go` `TestJoinReprojectsDependentsOnlyWhenTheRootMoves` *(new)* | `locatedBirth("James Robins", "York")`, all three promoted, `fileBridges`; overwrite the Event's and Person's documents through `searchindex.Upsert` with `Title` and `DisplayTitle` "stalemarker" (kind and entity id as stored); promote a new place Subject with toponym "York" onto the place handle → "stalemarker" still finds both; promote one with toponym "Eboracum" → "stalemarker" finds nothing, "Eboracum" finds the Event and the Person | AC-3.1, AC-3.2, D-10 |
| T-8 | Go | `core/search/handles_test.go` `TestNicknameFindsTheFormalName` *(new)* | Person "Tom Smith" + "Thomas Smith" → "Tom" → one Person hit "Thomas Smith", reason `other`, snippet ""; person "Thomas Smith" alone → "Tommy" finds it; person "William Smith" → "Robert" finds nothing | AC-2.1 – AC-2.3 |
| T-9 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` subtest "join a place onto the city" *(new)* | Before timing, `s := []subjects.Subject{f.place(lf.other, lf.otherA, "City")…}` three times; `d := fastest(3, …)` where each run takes the next Subject and returns `since(func() { f.join(s[i], lf.city) })`, so only the join is timed; `over(t, "join a place onto the city", d, budgetJoinOntoCity)` | AC-3.3 |
| T-10 | Swift | `SearchCatalogStoreTests.swift` `returnsSeededHitsInSeededOrder` *(new)* | seed [s1 source, p1 person, l1 place] (built with `hit(kind:id:)` below), kinds [] → ids [s1, p1, l1] | AC-4.1 |
| T-11 | Swift | `SearchCatalogStoreTests.swift` `filtersSeededHitsByRequestedKinds` *(new)* | kinds ["person","place"] → [p1, l1]; ["nonsense"] → [] | AC-4.2, AC-4.3 |
| T-12 | Swift | `SearchCatalogStoreTests.swift` `emptyQueryReturnsNoHits` (seeds hits instead of Sources) | "   " → [], `heldCatalogProjectDir == projectDir` | AC-4.4 |
| T-13 | Swift | `SearchCatalogStoreTests.swift` `searchErrorIsThrown` *(new)* | `searchCatalogError = CoreInvokeError.coded(…)` as in `OmnibarResultsModelTests` → `#expect(throws:)` | AC-4.5 |
| T-14 | Swift | The model tests below seed `seededSearchHits` and keep today's assertions | See the list below | D-13 |

T-10 – T-13 use a file-private helper in `SearchCatalogStoreTests.swift`: `hit(kind: String, id: String) -> CatalogSearchHit` = `CatalogSearchHit(kind: kind, id: id, ref: id.uppercased(), title: id, subtitle: "", matchReason: "title", location: WorkspaceLocation(section: .sources))`. The location is not asserted.

T-14 seeds:
- `CatalogSearchTests.exactRefHitIsTop` → renamed `modelKeepsEngineOrder`: `[CatalogSearchHit(kind: "source", id: "s1", ref: "SRC-ZZ9K2", title: "Quiet", subtitle: "", matchReason: "ref", location: WorkspaceLocation(section: .sources, sourceId: "s1", ref: "SRC-ZZ9K2", title: "Quiet")), the same for s2 with ref "SRC-OTHER", title "SRC-ZZ9K2 in title", reason "title"]`; expects first s1, reason "ref", `refAccent` true.
- `CatalogSearchTests.kindsRestrictHits`: `[person e-1 (ref "PER-1", title "James Robins", location .persons entityId "e-1"), source src-1 (ref "SRC-1", title "Robins register", location .sources sourceId "src-1")]`, kinds ["person"]; expects kinds ["person"], first ref "PER-1".
- `OmnibarResultsModelTests.debouncedSearchReturnsHits`: one source hit s1, ref "SRC-ILMIN", title "Ilminster parish", subtitle "Parish", iconKey "type_certificate", location `.sources` sourceId "s1"; the `sourcesByProject` and `sourceTypesByProject` seeds go.
- `OmnibarResultsModelTests.activateFieldHitFromSearch`: one hit kind "metadata_field", id "f9", ref "", title "Author", subtitle "author", location `WorkspaceLocation(section: .metadata, fieldId: "f9", title: "Author")`; the `fieldsByProject` seed goes.
- `OmnibarResultsModelTests.activateEvidenceGraphHitOpensTheGraph`: keep the `CatalogSource` and `CatalogEvidenceGraph` values as locals; one hit kind "evidence_graph", id "g-jones", ref "GPH-J0NES", title "Jones household", subtitle "1881 Census, Ward 3", location `SourcesListNavigation.canvasLocation(for: source, graph: graph)`; the `sourcesByProject` and `evidenceGraphsByProject` seeds go.
- `PromoteModelTests.aFoundRecordIsSearchedPickedAndHeld`: `storeWithFoundRecords` also sets `seededSearchHits` = [person e-far (ref "PER-FAR", title "Polly Jones", location `.persons` entityId "e-far"), place p-1 (ref "PLC-1", title "Polly Street", location `.places` entityId "p-1")]; the record search's kinds ["person"] gives ["e-far"]. The memberships stay (Promote reads them).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/search` (bm25 arity, `TestFoldedQueryScoresLikeExact`), `core/database/autoreconciler` `TestRebuildEqualsUpkeep_*`, `core/database/graphcache` verify tests, `core/database/searchindex` `TestMigrationCreatesSearchTables`, `TestMatchTextListsEachAlternateOnce`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS on macOS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: each deleted or renamed name (`searchPriority`, `orderedByMixedScore`, `handleSearchHits`, `handleSearchSummary`, `handleSearchMatchText`, `handleSearchKinds`, `scoreRefRecord`, `vocabularySearchHits`, FakeStore `classifyRefQuery` / `tokenizeSearch` / `scoreSearchFields`, `robinsSearchStore`, `fakeSearchFoldsLikeGo`, `rawTokens`) `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/`, `core/` with no hit left outside archived docs and the audit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — D-14 needs no rule text change
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI change
- [ ] Version bump: the next patch in VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the S14-21 scope entry (STEP-12); completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/conclusion-reconciliation.md §6 (STEP-10)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project and add a Place Subject with toponym "Søby", promote it. In the omnibar type "Soby": the place is the first result, with no match line.
2. Add a Source titled "Ærø kirkebog". Type "Aero": the Source is found.
3. On a Person recorded as "Tom Smith" in one Source and "Thomas Smith" in another (shown as Thomas Smith), type "Tom": Thomas Smith is found, with no match line under it.
4. Close and reopen an existing project made before this build: the first open rebuilds the search index (`ProjectionVersion` 10), and step 1's search works on it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The bm25 fifth weight changes the order of existing hits | `folded` is empty for any document without a fold letter or mid-word punctuation, so those ranks are unchanged; `core/search` ranking tests stay green |
| A dependent's search text reads something the row comparison misses | §4 lists what dependents read; T-7 holds the join; the root itself always reprojects; `TestRebuildEqualsUpkeep_*` compare upkeep to rebuild |
| `RebuildAll` on first open is slow on a large catalog | It is the cost of every `ProjectionVersion` bump (S11-07, S12-15) |

**Stop and report if:**

- `RecomputeReported` is not the only path that reprojects a handle's dependents' search documents on a join (T-7 finds the marker overwritten after the "York" join).
- A `TestRebuildEqualsUpkeep_*` case fails after STEP-6: upkeep and rebuild then disagree on a document, which means a dependent read something outside the displayed rows.
- `TestFoldedQueryScoresLikeExact` or another existing `core/search` ranking test changes order or score after STEP-2.
- T-5's product totals are not 2,627 and 15, or tom is not [tommy], at the planning base.
- `FoldsInto` cannot be written without changing `Expands` or `NewAliasIndex`.
- `valuerules` imports `searchindex` (an import cycle at STEP-3).
- T-9 measures over 1 s on a Mac with the narrowed reprojection.
- A Swift test outside §4 Swift tests fails because it relied on the fake matching query text.

## 12. Open questions

None.

## 13. PR

**Title:** S14-21 — Search finds every spelling
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-21-search-finds-every-spelling.md`.
