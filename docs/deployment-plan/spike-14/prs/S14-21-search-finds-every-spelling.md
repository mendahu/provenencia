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
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md). `add-catalog-model` does not apply: `catalog_search_docs` is already registered (`BucketSkip`) and the step adds no table and no foreign key, and `catalog_search_fts*` is outside `TestPragmaHonesty`. `add-match-rule` does not apply: no comparer, relation, cause or knob changes; `AliasIndex.FoldsInto` is a read of the existing fold rule (`Expands`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/searchable-kinds.mdc`, `.cursor/rules/versioning.mdc`
**Context:** audit 2 [B-7, B-10, B-21](../app-health-audit-2.md) (map review K-45); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §6 "Eliminated means not displayed, never dropped" (search reads every cached value); Spike 11 dogfood bar 6 and S11-07 D-6 (retrieval assumed to fold like rescoring); the map's Spike 11 leftover (2026-10-29: "Tom" no longer finds a person shown as Thomas); working rules: "FakeStore is a seeded fake, not a second engine", "A nickname folds into its set's formal name, as today", "Audit items keep their IDs"
**Planned:** 2026-11-02 against `main` @ `7b9b6e7`
**Review:** — (not reviewed yet)

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

The omnibar finds a record from any spelling the app already treats as the same. "Soby" finds Søby, "Aeroskobing" finds Ærøskøbing, "Lask" finds Łask, "Oeuilly" finds Œuilly and "Strasse" finds Straße, in every kind of search document; "Søby" finds a place written Soby; and "Tom" finds a person the reconciler shows as Thomas because a Tom record folded into that name. Joining a Subject onto a record whose shown values do not change (a second "City" toponym onto the 6,000-event city) no longer reprojects every dependent's search document. And FakeStore's search stops ranking: it returns seeded hits in seeded order for the requested kinds, so ranking is tested once, in Go.

**Why now.** Bar outcome 20. B-21 is a regression from S11-07 (`distinctAlternates` began dropping alternates that SQLite's tokenizer cannot fold). The nickname gap is a regression from S11-28 (the fold hides the nickname record's spelling from the document). B-10 measured 6.06 s for one join on the large fixture. B-7 leaves a second ranking engine in the fake that nothing pins to Go.

**What stays.** The ranking formulas (`scoreFields`, kind priority, context and ref boosts); match reasons and snippets for every hit that matched before; the trigram typo shortlist; which documents exist; the graph cache's display drops on commit (`SummaryDeps` keeps the full walk); FakeStore's error hook (`searchCatalogError`) and catalog-session hold; every Swift view.

## 2. User stories

**US-1.** As a researcher, when I type a spelling without æ, ø, ł, œ, ß, đ, þ or ı, I find the record written with them, and the reverse, so that the keyboard I have does not decide what I find.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Place records titled Ærøskøbing, Søby, Łask, Œuilly and Straße; queries "Aeroskobing", "Soby", "Lask", "Oeuilly", "Strasse" (Places kind) | Each query returns its place as the first hit, match reason `title` |
| AC-1.2 | A place whose toponyms are Søby (title) and Soby (deduped alternate); queries "Soby" and "Søby" | Both return the place |
| AC-1.3 | A place titled Soby; query "Søby" | Returns the place |
| AC-1.4 | A Source titled "Søby kirkebog", an Evidence graph named "Ærø household", a Source type labelled "Fødested", a metadata field labelled "Łódź parish"; queries "Soby", "Aero", "Fodested", "Lodz" (all kinds) | Each returns its record |
| AC-1.5 | A place with toponyms Ville-Marie (title) and Villemarie; query "Villemarie" | Returns the place: a spelling that differs only by a word break stays in the document |
| AC-1.6 | A Source titled "Montréal parish register"; queries "Montreal" and "Montréal" | Same hit and same score as today (`TestFoldedQueryScoresLikeExact` unchanged) |

**US-2.** As a researcher, when a person's Tom record folded into Thomas, typing "Tom" finds Thomas, so that the reconciler's fold does not hide the person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A person recorded as "Tom Smith" and "Thomas Smith", shown as "Thomas Smith"; query "Tom" (Persons) | One hit titled "Thomas Smith", match reason `other`, empty snippet (no match line) |
| AC-2.2 | A person recorded only as "Thomas Smith"; query "Tommy" | The person is found: the document carries every name that folds into Thomas (thom, thos, tom, tommy) |
| AC-2.3 | A person shown as "William Smith"; query "Robert" | No hit: robert does not fold into william |
| AC-2.4 | A person shown as "Tom Smith" | Its document has the line `alias:\ttommy` and no `alias:\tthomas`: thomas does not fold into tom |

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
| `searchindex.Upsert` / `Delete` write the folded words of every document kind; `ProjectionVersion` 10 | STEP-2, AC-1.1, AC-1.4 |
| `core/search` retrieval: a query word with a fold letter also matches its folded spelling; bm25 weighs the new column | STEP-3, AC-1.3, AC-1.6 |
| `distinctAlternates` dedupes at a key that keeps word breaks, so it drops no spelling FTS cannot match | STEP-4, AC-1.2, AC-1.5 |
| `AliasIndex.FoldsInto` and the product seed's fold counts pinned | STEP-5, AC-2.2, AC-2.3 |
| A Person document carries `alias:` lines for the names that fold into its shown given names | STEP-6, AC-2.1 – AC-2.4 |
| `RecomputeReported` reprojects a root's dependents only when the root's displayed values changed (B-10) | STEP-7, AC-3.1, AC-3.2 |
| Budget subtest for the join onto the city | STEP-8, AC-3.3 |
| FakeStore search returns seeded hits; its ranking code goes (B-7) | STEP-9, AC-4.1 – AC-4.5 |
| Swift tests that tested the fake's ranking go or seed hits | STEP-10 |
| `add-searchable-kind` drops its `searchPriority` line; domain doc | STEP-11 |

| Out | Home |
| --- | --- |
| The search registry's columns and projector table | Descoped with S12-36, not planned |
| Ranking changes (weights, kind priority, boosts) | Not planned |
| Composing reprojected dependents from the warm graph (the map's B-10 second half) | Not planned: D-9 |
| Narrowing `SummaryDeps` (the graph cache's display drops) to moved roots | Not planned: D-10 |
| Nicknames of a Person's other recorded names, or of its alternates | Not planned: D-7 |
| The catalog's own alias rows (user or plugin tables) in search documents | Not planned: D-6 |
| A match line for a nickname hit | Not planned (user, 2026-11-02) |
| Typo matching of folded spellings in the trigram shortlist | Not planned: D-3 |
| Accent folding before initials and nicknames in matching (B-19); NFC and compatibility forms (B-20) | Not scheduled in the audit; [`app-health-audit-2.md`](../app-health-audit-2.md) |
| Length caps on catalog text (audit L-10) | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| FTS schema | `core/database/migrations/000018.sql` | `catalog_search_docs` (title, ref, secondary, body) and `catalog_search_fts` over them, `tokenize='unicode61 remove_diacritics 2'`, `content='catalog_search_docs'`. `remove_diacritics` strips combining marks only, so æ, ø, ł, œ, ß, đ, þ, ı stay as written |
| Trigram schema | `core/database/migrations/000019.sql` | `catalog_search_fts_trigram` over the same four columns. Untouched by this PR |
| Latest migration | `core/database/migrations/000059.sql` | The next step is `000060.sql`. `000020.sql` and `000056.sql` added docs columns with a `ProjectionVersion` bump |
| Write path | `core/database/searchindex/searchindex.go` `Upsert`, `Delete` | `Delete` reads title, ref, secondary, body back and issues the FTS `'delete'` rows; `Upsert` inserts docs, then both FTS tables. Every document kind goes through `Upsert` |
| Version | `searchindex.go` `ProjectionVersion` = 9 | `EnsureCatalog` → `NeedsRebuild` → `RebuildAll` when the stored version lags |
| Handle docs | `core/database/searchindex/handles.go` `matchText` | Person: title = summary `Name.Form`, secondary = life lines + `name:` alternates. Place: title = first toponym, `toponym:` alternates, parents. Event: list title and date / place lines |
| Alternate dedupe | `handles.go` `distinctAlternates` | Keeps a value whose `TextKey(…, Equivalent)` is new. The key folds the `Folds` letters and joins words with no separator, so "Soby" is dropped beside "Søby" and "Villemarie" beside "Ville-Marie", though FTS matches neither from the other (B-21) |
| Word rules | `handles.go` `valueWords` = `matchregistry.Default().Value.Words` | `Folds` (`core/matchregistry/value.go`): ß→ss, æ→ae, œ→oe, ø→o, ł→l, đ→d, þ→th, ı→i |
| Fold helpers | `core/valuecompare/words.go` `NormalizeForm`, `FoldWord`, `SplitWords` | `NormalizeForm` lowercases, drops punctuation, turns dash / slash / space into one space. `FoldWord(w, nil)` strips accents only (what `remove_diacritics 2` does); `FoldWord(w, Folds)` also applies the table |
| Retrieval | `core/search/fts.go` `buildMatchQuery`, `mergeFTS` | `"tok"` per token, last with `*`, joined by AND or OR; `mergeFTS` ranks with `bm25(catalog_search_fts, title, ref, secondary, body)` weights from `FTSBM25Weights` (`core/search/registry.go`: 10, 12, 4, body) |
| Rescoring | `core/search/registry.go` `newQueryTokens`, `scoreFields`, `alternateMatch` | Keys fields and tokens at `Equivalent`, so it already folds both sides; it never sees a document FTS did not retrieve. `alternateMatch` reads only `name:` and `toponym:` lines; other lines leave reason `other` |
| Swift match line | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `showsMatchContext` | Draws a line only for notes, metadata, filename, description, name, place. `other` draws nothing |
| Aliases | `core/valuecompare/aliases.go` `AliasIndex.Expands`, `SetKeys`, `AliasToken` | `Expands(from, into, role)` is the reconciler's fold rule. Nothing lists the names that fold into a name |
| Reconciler fold | `core/autoreconcile/names.go` `nameModule.fold` | Reads `m.rules.Names.Western.Role(typ)` and `Aliases.Expands(p.A.Text, p.B.Text, role)` |
| Product rules | `core/database/valuerules/valuerules.go` `Product` | Registry value rules + `namealiases.ProductIndex()`: the reconciler cache's rules. No import of `searchindex` (no cycle) |
| Product seed | `core/database/namealiases` `ProductTables` | Two tables (abbreviations, 8 sets; nicknames, 1,229 sets), `FormalFirst`, roles given / nick / untyped. Over every token: 2,627 `Expands` pairs into 1,157 names, at most 15 into one name (margaret, marietta) |
| Summary name | `core/database/conclusionsummaries/conclusionsummaries.go` `PersonSummary.Name` | `*namevalue.Value` with typed `Parts` from the reconciled value |
| Reprojection | `core/database/autoreconciler/autoreconciler.go` `RecomputeReported` | Recomputes `ids`, walks `conclusionsummaries.Dependents` from ids + association endpoints, reprojects all of them, returns the dependents for `SummaryDeps`. A join that changes no shown value still reprojects the city and its 6,000 events (B-10: 6.06 s, `RowsForSearch` 4.41 s through a one-shot `graphcache.Over`) |
| Callers | `core/writes/run.go` `Run`; `core/database/subjects/subjects.go` (`RecomputeTx`); `graphcachetest` | `Run` stores the returned dependents as `set.SummaryDeps`, `SummaryDepsWalked = true`; `graphcache` `dropCommittedDisplays` drops roots + those |
| Dependents' text | `core/database/searchindex/handles.go` `personLifeMatch`, `eventMatchSecondary`, `eventMatchTitle`; `conclusionsummaries` `toponyms`, `FirstPlaceName` | Read only cached value rows (kept toponyms, names, dates, types) of the records they name; never a label, ref or support count |
| Cache rows | `auto_reconciler_values` (`000040.sql`, `000053.sql`) | Value columns, `rank`, `reason`, plus `support`, `against` (vote counts summaries never read; only `core/database/graphcache/verify.go` compares them) |
| Large fixture | `core/database/promotealign/perf_test.go` `newLargeFixture`, `TestLargeSourceBudgets` | The city (6,000 events) is `cityID`, a local; `f.place`, `f.join` in `promotealign_test.go` |
| FakeStore search | `macos/App/Platform/FakeStore.swift` `searchCatalog`, `handleSearchKinds`, `searchPriority`, `orderedByMixedScore`, `handleSearchHits`, `handleSearchSummary`, `handleSearchMatchText`, `scoreRefRecord`, `vocabularySearchHits`, `classifyRefQuery`, `tokenizeSearch`, `scoreSearchFields` | Scores Sources, graphs, vocabulary and handles in memory, mirrors Go's kind mix and ref boosts, caps at 50. Each helper's only caller is in this list |
| Search guidance | `.agents/skills/add-searchable-kind/SKILL.md` line 32 | "until S12-50 removes it, for FakeStore's `searchPriority`". The rule (`.cursor/rules/searchable-kinds.mdc`) and line 33 already say the fake returns seeded hits |
| Go tests | `core/search/handles_test.go` `TestAlternateToponymIsMatchContext`, `TestSummarySearchFindsTheListTitle`; `core/search/search_test.go` `TestFoldedQueryScoresLikeExact`, `TestSearchAccentedPlaceTypo`; `core/database/searchindex/searchindex_test.go` `TestMatchTextListsEachAlternateOnce` | Montréal-only folding; no ø / æ / ł / œ / ß case; no nickname case |
| Swift tests | `macos/ProvenenciaTests/SearchCatalogStoreTests.swift` (all, plus `robinsSearchStore`); `CatalogSearchTests.swift` `exactRefHitIsTop`, `kindsRestrictHits`, `noLocationGivesNoSectionNudge`; `OmnibarResultsModelTests.swift` `debouncedSearchReturnsHits`, `activateFieldHitFromSearch`, `activateEvidenceGraphHitOpensTheGraph`; `PromoteModelTests.swift` `aFoundRecordIsSearchedPickedAndHeld` | Each seeds Sources, vocabulary or memberships and relies on the fake matching the query text |

The map's Context names `conclusionheaders.RowsForSearch`; it is `conclusionsummaries.RowsForSearch` since `1d475bde`.

**Precedent.** `a26a944c` (S12-15, #583): a `catalog_search_docs` column in `000056.sql`, `ProjectionVersion` 8 → 9 with a comment line, and `Upsert` writing it. Copy the migration header comment style and the version comment. `4ccde944` (S11-07): `distinctAlternates` and `TestMatchTextListsEachAlternateOnce`; extend that test's table. #617 (S12-35): FakeStore turned from rules into seeded rows and scripted errors; copy its shape (a `seeded…` property, the `…Error` hook, the session hold).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Folded spellings go in a new FTS column, `folded`, on `catalog_search_docs` and `catalog_search_fts`, written by `Upsert` for every document kind | The user chose every kind (2026-11-02). Every kind's text columns are read by `scoreFields` as named fields (description, notes, key), so folded text in them would score and snippet as that field. One column written in `Upsert` serves every projector, present and future | A tagged `fold:` line in `secondary` — handle-only, and a Source's secondary is its description field; dedupe at a non-folding key only — leaves "Soby" unable to find a lone "Søby" |
| D-2 | `folded` holds, space-joined, de-duplicated in first-seen order, `FoldWord(w, valueWords.Folds)` for each word `w` of the document's title, secondary and body (split by `NormalizeForm` + `strings.Fields`, not `SplitWords`) when that word differs from `FoldWord(w, nil)`, or when its raw token held a rune `NormalizeForm` dropped (an apostrophe: "O'Brien" → `obrien`). Ref is skipped (Crockford ASCII). Empty when no word qualifies | Exactly the words `remove_diacritics 2` and unicode61's separators index differently from rescoring's key. Montréal adds nothing, so its bm25 and score stay as today (AC-1.6). `NormalizeForm` keeps particles apart ("St. Søby" → `soby`), which `SplitWords` would join | Every word folded — doubles every document and moves every bm25 rank; NFD + Mn-strip via `x/text` in a second helper — `FoldWord(w, nil)` already is that |
| D-3 | The trigram table keeps its four columns | It is the typo shortlist; the JW gate reads identity fields. Folded typos are not in the bar | Adding `folded` there too — a second table rebuilt for no AC |
| D-4 | `buildMatchQuery` writes a token whose `FoldWord(tok, Folds)` differs from `FoldWord(tok, nil)` as `("tok" OR "folded")`, with `*` on both when it is the last token | Query "Søby" must reach a doc written "Soby" (AC-1.3), which has no `folded` word. Tokens without a fold letter are written exactly as today | Folding the query only — loses a doc whose own spelling is the unfolded one in a column FTS reads raw |
| D-5 | `FTSBM25Weights` gains `Folded` = 4 (the Secondary weight); `mergeFTS` passes five weights in table column order | The folded words come from title, secondary and body mixed; the lower weight keeps a folded-only match from out-ranking the same doc's direct title match. `scoreFields` still decides the field and most of the score | Title weight 10 — lets body words ride a title weight |
| D-6 | Nickname expansions read the product alias seed (`valuerules.Product()`), not the catalog's alias rows | It is the index the reconciler folded with (`autoreconciler` `ValueRules`), so search restores exactly what the fold hid. The cache, and so the document, changes only with `CacheVersion` / `ProjectionVersion`, not with a catalog alias edit (which would otherwise need every Person reprojected) | Catalog rows via `valuerules.Load` — a second invalidation path for no AC |
| D-7 | A Person document carries, after its `name:` alternates, one `alias:\t<token>` line per name that folds into a word of the shown name's parts whose role (`Names.Western.Role(part.Type)`) is in `AliasRoles()`: `AliasIndex.FoldsInto(AliasToken(word), role)`, de-duplicated, sorted, excluding the name's own words. Only the shown name, not alternates | The user chose fold-in only (2026-11-02): the reconciler's own rule, so "Robert" never finds William and "Rose" never finds Mary. Alternates are already in the document as written | Every set-mate (user declined); alternates' nicknames — not hidden by any fold |
| D-8 | `alternateMatch` does not read `alias:` lines, so a nickname hit keeps reason `other` and an empty snippet; Swift draws no match line for `other` | The user chose no match line (2026-11-02). No Swift or copy change | A `nickname` reason with a new L10n line (user declined) |
| D-9 | B-10 is fixed by not reprojecting dependents of a root whose displayed values did not move. Reprojected dependents still compose through `graphcache.Over(q)` | The warm graph holds committed state; inside the write it would compose a dependent of a moved root from the root's old names, which is the wrong document. With the filter, the join onto the city reprojects the city alone | Composing from `c.Graph()` inside the transaction — stale reads |
| D-10 | "Displayed values" of a handle = its `auto_reconciler_values` rows ordered by `(property_id, rank)`, each `property_id, rank, reason` and the value columns (`value_text, value_integer, value_term_id, value_entity_id, value_date, value_name, value_image_file_id`), without `support, against, date_lo, date_hi, sort_key`. A root moved when its digest after `recomputeBatch` differs from before (a handle with no rows has the empty digest). The Dependents walk for reprojection starts from the moved ids plus the association endpoints (before and after) as today; `SummaryDeps` keeps the full walk from every root | Dependents' search text reads only those rows (§4 "Dependents' text"). `support`/`against` are the vote counts a join moves; `date_lo/hi` and `sort_key` derive from the value. Keeping the full walk for `SummaryDeps` leaves the graph cache's display drops exactly as today, since a display row can show a root's label | Comparing composed summaries before and after — the cost B-10 is about; narrowing `SummaryDeps` too — display rows read the entity row (label), which the digest does not cover |
| D-11 | The digest is `sha256` over the rows' bytes, read with one batched query (`sqlbatch.Query`) before and one after the recompute | Constant query count per batch, as `load` keeps (`TestLoaderQueryCountIsConstant`) | A per-handle query |
| D-12 | FakeStore: `var seededSearchHits: [CatalogSearchHit]` *(new)*; `searchCatalog` throws `searchCatalogError` if set, marks the session held, returns [] for a whitespace-only query, else the seeded hits whose `searchKind` is in the requested kinds (all kinds when the list is empty; an unknown kind matches nothing), in seeded order, uncapped | The working rule and `.cursor/rules/searchable-kinds.mdc` line 16 ("Empty / whitespace queries return no hits; FakeStore returns seeded hits in seeded order for the requested kinds"). Go caps at the request limit; the fake returns what a test seeded | Filtering seeds by query text — a matching engine again |
| D-13 | Swift tests that tested the fake's ranking or matching are deleted; tests of a model's behavior seed `seededSearchHits` with the hits the engine would return | `.cursor/rules/swift-tests.mdc`: engine rules are tested in Go (`core/search` has the ref, prefix, kind-mix and section-nudge tests) | Keeping them against seeds — they would test nothing |
| D-14 | One PR, over plan-pr's size limit (13 steps, Go and Swift) | The user, 2026-11-02, chose to keep one PR | Splitting into three PRs (recommended) |

**Rules this plan bends.** plan-pr **Size** ("Steps go past about 10": propose a split) → the user answered 2026-11-02: keep S14-21 as one PR (D-14). No rule text changes; the map row keeps one PR.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000060.sql` *(new)* | migration | `000056.sql` (docs columns) | Same shape plus the FTS table rebuild an FTS column needs |
| `catalog_search_docs.folded`, `catalog_search_fts.folded` *(new)* | column | `secondary` | `secondary` is a scored field for every kind (D-1) |
| `searchindex.foldedWords(texts ...string) string` *(new)* | func | `valuecompare.TextKey`, `FoldWord`, `NormalizeForm` | Composes them; the rule "what FTS indexes differently" belongs to the search projector, not to value compare |
| `searchindex.alternateKey(s string) string` *(new)* | func | `valuecompare.TextKey(…, Equivalent)` | That key joins words with no separator (it must, for compare); this one keeps the breaks FTS keeps |
| `core/search` `foldedVariant(tok string) (string, bool)` *(new)* | func | `newQueryTokens` (keys at Equivalent) | Retrieval needs the word folded with breaks kept, not the joined key |
| `FTSBM25Weights.Folded` *(new)* | field | `FTSBM25Weights` | Extended in place |
| `valuecompare.AliasIndex.FoldsInto(into string, role NameRole) []string` *(new)* | method | `AliasIndex.Expands`, `SetKeys` | `Expands` answers one pair; nothing lists the set. `FoldsInto` filters set-mates through `Expands`, so it cannot disagree with it |
| `searchindex.nicknamesOf(name *namevalue.Value) []string` *(new)*, `productRules` *(new)* package var | func, var | `valueWords`; `autoreconcile` `nameModule.fold` role read | Same role read (`Names.Western.Role`) as the fold |
| `autoreconciler.displayedDigests(q, ids) (map[string][32]byte, error)` *(new)* | func | `conclusionsummaries.AssociationEndpoints` (the before/after read in the same function) | Same before/after pattern on the value rows |
| `largeFixture.city` *(new)* field, `budgetJoinOntoCity` *(new)* | test | `largeFixture.onePlace`, `budgetOnePlaceWarm` | Extends the fixture and budget block |
| `(f *handleFixture) locatedBirth(…)` *(new)* | test helper | `TestSummarySearchFindsTheListTitle` body | Extracted from it; both tests use it |
| `FakeStore.seededSearchHits` *(new)* | Swift var | `seededEventSummaries`, `seededPlaceSummaries` | Same seeded-row pattern |

**Builds on.** `searchindex.Upsert` as the one write of every document; `ProjectionVersion` heal-on-open; `matchregistry` `Folds` as the one fold table; `AliasIndex.Expands` as the one fold rule; S12-35's seeded FakeStore.
**Deletes.** FakeStore: `handleSearchKinds`, `searchPriority`, `orderedByMixedScore`, `handleSearchHits`, `handleSearchSummary`, `handleSearchMatchText`, `scoreRefRecord`, `vocabularySearchHits`, `classifyRefQuery`, `tokenizeSearch`, `scoreSearchFields` and the `prefix(50)` cap. Swift tests: `SearchCatalogStoreTests` `fakeSearchFoldsLikeGo`, `findsSourceByTitleAndMapsLocation`, `findsEvidenceGraphByNameRefAndDescription` (with `seedCensusGraphs`), `findsTypeAndFieldByLabel`, `findsSourceByRef`, `prefixRefPromotesSource`, `kindPriorityIsATenthAndSectionNudgeStillMatters`, `searchWithoutLocationHasNoSectionNudge`, and the file-level `robinsSearchStore`; `CatalogSearchTests` `noLocationGivesNoSectionNudge`. Guidance: `add-searchable-kind` line 32's `searchPriority` clause. Go: `distinctAlternates`'s `TextKey` call (replaced by `alternateKey`).
**Next adds.** A new search kind gets folded spellings with no code (its projector calls `Upsert`). A catalog-alias-aware expansion (D-6 rejected) would swap `productRules` for `valuerules.Load` plus a reproject on alias edits. S14-26 (a folded nickname shows as a `nick` part) changes the shown name's parts; `nicknamesOf` reads a `nick` part's role (`RoleNick`, in `AliasRoles`) with no change.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A document found only through `folded` or an `alias:` line | `core/search` `Search` scoring branch | `scoreFields` keys every field at Equivalent, so a folded hit scores on its real field (title for AC-1.1); an alias hit scores `other` and falls through `alternateMatch` with `field == ""`, keeping reason `other` (no new branch) |
| A root that moved nothing | `RecomputeReported` | The root itself is always reprojected; only its dependents are skipped |
| A whitespace query or an unknown kind on FakeStore | `searchCatalog` | Returns [] (D-12) |

### Scale

At the map's Scale (6,000 events on one place; 37,000 Subjects): a join onto the city that moves no displayed value reprojects the city and any association endpoints only, not its 6,000 events and their people. `SummaryDeps` is unchanged, so `graphcache` drops the same displays as today (B-10 measured 84–98 ms for that commit). The two digest reads add one batched `SELECT` each over the recomputed ids' rows. A join that renames the city still reprojects all dependents, as today (5.6 s on the fixture); that case is a rename, measured in B-10, and not in the bar. `foldedWords` is linear in a document's text and runs once per `Upsert`. `RebuildAll` on first open after the update rewrites every document once (the heal every `ProjectionVersion` bump costs). Budget: `budgetJoinOntoCity` = 1 s, against 6.06 s today; B-10's profile puts everything outside `ReprojectHandles` at 0.13 s on a Mac, and `budgetAfterCommit` notes CI runs about 3× slower.

### Edges

- **Assumptions.** (1) `remove_diacritics 2` folds exactly what `FoldWord(w, nil)` folds: both strip combining marks after canonical decomposition. Held by T-1 rows (Montréal → no folded word; Søby → `soby`). (2) unicode61 splits at every rune `NormalizeForm` drops or breaks at; T-1 row "O'Brien" → `obrien`. (3) `distinctAlternates` still keeps the first of each key in order and the title first; `TestMatchTextListsEachAlternateOnce` rows stay. (4) Dependents' search text reads only the roots' value rows (§4); a label change on a root with no value change does not move a dependent's document. T-7 holds the join; the label edge is covered by the reprojection of the root itself. (5) `FoldsInto` agrees with `Expands` for every pair: T-5 checks every product token.
- **Neighbors.** Folding: Montréal (no change), "Ville Marie" vs "Ville-Marie" (still one alternate), "Villemarie" (now kept, AC-1.5), "Straße"/"Strasse" (`strasse`), uppercase Ø (lowercased first), a digit word (never folded). Query: "Søby" OR-expanded; "soby" unchanged; a ref-shaped query skips FTS as today. Nicknames: Tom → Thomas found; Thomas → Tom not through aliases (AC-2.4); Robert → William not found (AC-2.3); a person with no parts or only a surname gets no lines; a `nick` part (S14-09) expands as a given. Reprojection: a support-only change skips dependents (AC-3.1); a new toponym, an outvoted value turning kept (reason changes), a removed value, a merge that empties the handle's rows: each moves the digest and walks.
- **Removed inputs.** A deleted or merged root loses its rows; the empty digest differs from its old one, so its dependents are reprojected as today. A root that had no rows before and has none after moves nothing its dependents read. Alias lines come from the shown name only: when the Tom record is excluded or deleted, Thomas keeps its expansions (they come from the seed, not from the Tom record), which is D-7's rule.
- **Real data.** The product seed: 2,627 fold pairs into 1,157 names; most into one name: 15 (margaret, marietta); thomas → {thom, thos, tom, tommy}; william → {bill, billy, wil, will, willie, willy, wm}; john → {jack, jno, jock, johnny, jonnie, jonny}; robert → {bill, billy, bob, bobby, dob, dobbin, hob, hobkin, rob, robby, robt, rupert}. Worth a look: "bill" folds into both william and robert, so "Bill" finds both, which the reconciler would also fold. T-5 pins the 2,627 and 15.
- **Untrusted size.** A shown name's words are capped: `nicknamesOf` returns nil when the name has more than `Names.MaxWords` (32) words, before any lookup; at most 15 tokens per word, so ≤ 480 lines. `foldedWords` is linear in the document text, de-duplicated by a map. The digest reads rows the engine wrote.

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

`catalog_search_docs` stays `BucketSkip` in `catalogmodel`; no FK. No audit. Schema hash: computed at init (`add-catalog-migration` § Schema hash); no constant to edit.

### Go core

| Package | Change |
| --- | --- |
| `core/database/searchindex` | `Document` gains no field: `Upsert` computes `folded := foldedWords(doc.Title, doc.Secondary, doc.Body)` and writes it to docs and `catalog_search_fts`; the trigram insert is unchanged. `Delete` selects and passes `folded` in the `catalog_search_fts` `'delete'` row. `ProjectionVersion` = 10, comment `10: a folded column for spellings FTS indexes differently, and a Person's nicknames that fold into its shown name (S14-21).` |
| `core/database/searchindex` (`handles.go`) | `var productRules = valuerules.Product()` *(new)*. `distinctAlternates` keys with `alternateKey(s)` *(new)*: `strings.Join` of `FoldWord(w, valueWords.Folds)` over `strings.Fields(valuecompare.NormalizeForm(s))`, one space; empty → dropped as today. `matchText` Person case appends `tagged(nicknamesOf(p.Name), "alias")` after the `name:` lines. `nicknamesOf` *(new)*: nil for nil name or > `productRules.Names.MaxWords` words; for each part, `role := productRules.Names.Western.Role(part.Type)`, skip unless `slices.Contains(valuecompare.AliasRoles(), role)`; for each word of `valuecompare.SplitWords(part.Value, productRules.Words)`, `t, ok := valuecompare.AliasToken(word, productRules.Words)`, append `productRules.Names.Aliases.FoldsInto(t, role)`; drop the name's own tokens; sort; compact |
| `core/valuecompare` (`aliases.go`) | `func (a AliasIndex) FoldsInto(into string, role NameRole) []string` *(new)*: every token of a set in `byToken[into]` that allows `role`, other than `into`, for which `a.Expands(token, into, role)`; sorted, no duplicates; nil for the zero index |
| `core/search` | `FTSBM25Weights.Folded` = 4 *(new)*; `mergeFTS` passes `w.Folded` as the fifth bm25 weight. `buildMatchQuery` writes `("tok" OR "fold")` (with `*` on both for the last token) when `foldedVariant(tok)` reports a different spelling *(new)*: `FoldWord(tok, valueWords.Folds)` vs `FoldWord(tok, nil)`, after `sanitizeFTSToken` |
| `core/database/autoreconciler` | `displayedDigests` *(new)*: `SELECT entity_id, property_id, rank, reason, value_text, value_integer, value_term_id, value_entity_id, value_date, value_name, value_image_file_id FROM auto_reconciler_values WHERE entity_id IN (…) ORDER BY entity_id, property_id, rank`, batched; per entity a `sha256` of the scanned columns with length-prefixed bytes (a NULL distinct from empty). `RecomputeReported`: `before := displayedDigests(ids)` ahead of the recompute, `after` behind it; `moved` = ids whose digests differ; `deps := Dependents(roots)` as today for the return value; `reproject := Dependents(moved ∪ before-endpoints ∪ after-endpoints)` (reuse `deps` when `moved` has every id); `ReprojectHandles(roots ∪ reproject)`; return `deps` |

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

**STEP-1 — Catalog step 000060**
Files: `core/database/migrations/000060.sql` *(new)*. Skill: `add-catalog-migration`.
Do: the DDL in §6 Schema, with its comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` fails only where `Upsert`/`Delete` still pass four columns (the `'delete'` and insert column lists); `catalog.schema_mismatch` does not appear.

**STEP-2 — Write the folded words**
Files: `core/database/searchindex/searchindex.go`, `core/database/searchindex/searchindex_test.go`.
Do: `foldedWords` (D-2), `Upsert`/`Delete` carry `folded` (§6 Go core), `ProjectionVersion` 10 with its comment. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**STEP-3 — Retrieve folded spellings**
Files: `core/search/fts.go`, `core/search/registry.go`, `core/search/search_test.go`, `core/search/handles_test.go`.
Do: `FTSBM25Weights.Folded`, five-weight `bm25`, `foldedVariant` and the OR form in `buildMatchQuery` (D-4, D-5). Add T-2, T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/search/...` passes, `TestFoldedQueryScoresLikeExact` unchanged.

**STEP-4 — Keep every spelling FTS cannot match**
Files: `core/database/searchindex/handles.go`, `core/database/searchindex/searchindex_test.go`.
Do: `alternateKey` and `distinctAlternates` on it (§6). Add T-6's rows.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/ ./core/search/` passes.

**STEP-5 — List the names that fold into a name**
Files: `core/valuecompare/aliases.go`, `core/valuecompare/name_test.go`, `core/database/namealiases/namealiases_test.go`.
Do: `AliasIndex.FoldsInto` (§6). Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/ ./core/database/namealiases/` passes.

**STEP-6 — Carry a Person's folded-in nicknames**
Files: `core/database/searchindex/handles.go`, `core/database/searchindex/searchindex_test.go`, `core/search/handles_test.go`.
Do: `productRules`, `nicknamesOf`, the `alias:` lines in `matchText` (D-6, D-7, D-8). Add T-6's nickname row and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/ ./core/search/` passes; `go list -deps ./core/database/searchindex | grep -c searchindex` prints 1 (no cycle).

**STEP-7 — Reproject dependents only when a root moved**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/search/handles_test.go`.
Do: `displayedDigests` and the narrowed reprojection (D-9 – D-11). Extract `locatedBirth` from `TestSummarySearchFindsTheListTitle` and use it there unchanged in behavior; add T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/ ./core/search/ ./core/writes/ ./core/database/graphcache/...` passes, including `TestRebuildEqualsUpkeep_Scenarios`, `TestRebuildEqualsUpkeep_SeededSequences` and `TestLoaderQueryCountIsConstant`.

**STEP-8 — Budget the join onto the city**
Files: `core/database/promotealign/perf_test.go`.
Do: T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets -v ./core/database/promotealign/` passes and logs the join time.

**STEP-9 — FakeStore returns seeded hits**
Files: `macos/App/Platform/FakeStore.swift`. Skill: `add-swift-test` (for STEP-10).
Do: D-12; delete the eleven helpers and the cap (§6 Deletes); keep the doc comment on `searchCatalog` to one line: "Seeded hits in seeded order for the requested kinds; ranks nothing (`.cursor/rules/swift-tests.mdc`)."
Check: the app target builds (`xcodebuild build` per `scripts/check-pr.sh`); `rg -n 'searchPriority|orderedByMixedScore|scoreRefRecord|scoreSearchFields|classifyRefQuery|tokenizeSearch|vocabularySearchHits|handleSearchHits' macos` prints nothing.

**STEP-10 — Swift tests seed hits**
Files: `macos/ProvenenciaTests/SearchCatalogStoreTests.swift`, `CatalogSearchTests.swift`, `OmnibarResultsModelTests.swift`, `PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: delete the tests in §6 Deletes; add T-10 – T-13; change the other tests named in §4 Swift tests to seed `seededSearchHits` (T-14).
Check: `xcodebuild test` as `scripts/check-pr.sh` runs it passes with no Swift warning.

**STEP-11 — Guidance and domain doc**
Files: `.agents/skills/add-searchable-kind/SKILL.md`, `docs/conclusion-reconciliation.md`.
Do: line 32 ends at "`leadThumbnail`" (drop the `searchPriority` clause). In `conclusion-reconciliation.md` §6, after "search reads them all, so an outvoted spelling still finds its Person": "A nickname the reconciler folded away finds it too: a Person's search document carries the names that fold into its shown given names under the product alias seed (S14-21)."
Check: `python3 scripts/check-guidance-names.py` passes.

**STEP-12 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`, both lines).
Do: the next patch after `VERSION` at merge, in its own commit.
Check: the three agree (`rg -n '0\.2\.9' VERSION core/version.go macos/Provenencia.xcodeproj/project.pbxproj` at planning's numbers).

**STEP-13 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-21-search-finds-every-spelling.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/searchindex/searchindex_test.go` `TestFoldedWords` *(new)* | "Søby" → "soby"; "Montréal" → ""; "Ærøskøbing Havn" → "aeroskobing"; "Straße Straße" → "strasse"; "O'Brien" → "obrien"; "St. Søby" → "soby"; ("Łask", "Søby", "Łask") → "lask soby"; "" → "" | D-2 |
| T-2 | Go | `core/search/handles_test.go` `TestFoldedSpellingsFindPlaces` *(new)* | Places Ærøskøbing, Søby, Łask, Œuilly, Straße (one fixture each), queries "Aeroskobing", "Soby", "Lask", "Oeuilly", "Strasse" → that place first, reason `title`; place Søby + toponym Soby → "Soby" and "Søby" both find it; place Soby → "Søby" finds it | AC-1.1, AC-1.2, AC-1.3 |
| T-3 | Go | `core/search/search_test.go` `TestFoldedSpellingsFindEveryKind` *(new)* | Source "Søby kirkebog" ← "Soby"; Evidence graph "Ærø household" (the Source's sole graph renamed with `evidencegraphs.Update` inside `writes.Run`) ← "Aero"; Source type "Fødested" (`runSourceTypeCreate`) ← "Fodested"; metadata field "Łódź parish" (`runFieldCreate`) ← "Lodz"; Source "Montréal parish register" ← "Montreal" and "Montréal" give equal scores (as `TestFoldedQueryScoresLikeExact`, which stays) | AC-1.4, AC-1.6 |
| T-4 | Go | `core/search/search_test.go` `TestBuildMatchQueryAddsFoldedSpelling` *(new)* | ["søby"] → `("søby"* OR "soby"*)`; ["soby"] → `"soby"*`; ["søby","kirke"] AND → `("søby" OR "soby") AND "kirke"*`; ["montréal"] → `"montréal"*` | D-4 |
| T-5 | Go | `core/valuecompare/name_test.go` `TestAliasIndexFoldsInto` *(new)*; `core/database/namealiases/namealiases_test.go` `TestProductFoldsInto` *(new)* | Tiny index {thomas, tom, tommy} FormalFirst: FoldsInto(thomas, given) = [tom tommy]; FoldsInto(tom, given) = []; role surname → []; zero index → nil. Product: for every token, each FoldsInto result `Expands` into it; totals 2,627 pairs, max 15; thomas = [thom thos tom tommy]; william has no robert | D-7, AC-2.2, AC-2.3 |
| T-6 | Go | `core/database/searchindex/searchindex_test.go` `TestMatchTextListsEachAlternateOnce` (extended) | Place toponyms Ville-Marie, Villemarie → secondary has `toponym:\tVillemarie`; place Søby, Soby → no Soby line (folded column covers it); person "Thomas Smith" with typed parts (`namevaluetest.Western`) → `alias:\tthom`, `alias:\tthos`, `alias:\ttom`, `alias:\ttommy` after the `name:` lines; person "Tom Smith" → `alias:\ttommy` only; existing rows unchanged | AC-1.5, AC-2.4, D-7 |
| T-7 | Go | `core/search/handles_test.go` `TestJoinReprojectsDependentsOnlyWhenTheRootMoves` *(new)* | `locatedBirth` (person "James Robins", birth at "York"); overwrite the Event's and Person's documents with title "stalemarker" through `searchindex.Upsert`; promote a place Subject "York" onto the place → "stalemarker" still finds both; promote a place Subject "Eboracum" onto it → "stalemarker" finds nothing, "Eboracum" finds the Event and the Person | AC-3.1, AC-3.2, D-10 |
| T-8 | Go | `core/search/handles_test.go` `TestNicknameFindsTheFormalName` *(new)* | Person "Tom Smith" + "Thomas Smith" → "Tom" → one hit "Thomas Smith", reason `other`, snippet ""; person "Thomas Smith" alone → "Tommy" finds it; person "William Smith" → "Robert" finds nothing | AC-2.1 – AC-2.3 |
| T-9 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` subtest "join a place onto the city" *(new)* | Three place Subjects "City" on `lf.other`; fastest of three `since(func() { f.join(s, lf.city) })` < `budgetJoinOntoCity` (1 s) | AC-3.3 |
| T-10 | Swift | `SearchCatalogStoreTests.swift` `returnsSeededHitsInSeededOrder` *(new)* | seed [s1 source, p1 person, l1 place], kinds [] → [s1, p1, l1] | AC-4.1 |
| T-11 | Swift | `SearchCatalogStoreTests.swift` `filtersSeededHitsByRequestedKinds` *(new)* | kinds ["person","place"] → [p1, l1]; ["nonsense"] → [] | AC-4.2, AC-4.3 |
| T-12 | Swift | `SearchCatalogStoreTests.swift` `emptyQueryReturnsNoHits` (seeds hits instead of Sources) | "   " → [], session held | AC-4.4 |
| T-13 | Swift | `SearchCatalogStoreTests.swift` `searchErrorIsThrown` *(new)* | `searchCatalogError` set → throws it | AC-4.5 |
| T-14 | Swift | `CatalogSearchTests.swift` `exactRefHitIsTop` → `modelKeepsEngineOrder` (seed [s1 ref hit with reason "ref", s2]; first is s1, `refAccent` true); `kindsRestrictHits` (seed a person and a source; kinds ["person"] → the person); `OmnibarResultsModelTests.swift` `debouncedSearchReturnsHits`, `activateFieldHitFromSearch`, `activateEvidenceGraphHitOpensTheGraph` (seed the hit each expects, with the location Go returns for it); `PromoteModelTests.swift` `aFoundRecordIsSearchedPickedAndHeld` (seed person `e-far` and place `p-1`; kinds ["person"] → ["e-far"]) | Same assertions as today on the model's behavior | D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/search` (bm25 arity, `TestFoldedQueryScoresLikeExact`), `core/database/autoreconciler` `TestRebuildEqualsUpkeep_*`, `core/database/searchindex` `TestMigrationCreatesSearchTables`, `TestMatchTextListsEachAlternateOnce`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: each deleted name (`searchPriority`, `orderedByMixedScore`, `handleSearchHits`, `handleSearchSummary`, `handleSearchMatchText`, `scoreRefRecord`, `vocabularySearchHits`, `FakeStore` `classifyRefQuery` / `tokenizeSearch` / `scoreSearchFields`, `robinsSearchStore`, `fakeSearchFoldsLikeGo`) `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/` with no hit left outside archived docs and the audit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — D-14 needs no rule text change
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI change
- [ ] Version bump: the next patch in VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/conclusion-reconciliation.md §6 (STEP-11)
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
| The bm25 fifth weight changes the order of existing hits | `folded` is empty for any document without a fold letter or split punctuation, so those ranks are unchanged; `core/search` ranking tests stay green |
| A dependent's search text reads something the digest misses | §4 lists what dependents read; T-7 holds the join; the root itself always reprojects |
| `RebuildAll` on first open is slow on a large catalog | It is the cost of every `ProjectionVersion` bump (S11-07, S12-15) |

**Stop and report if:**

- `RecomputeReported` is not the only path that reprojects a handle's dependents' search documents on a join (T-7 finds the marker overwritten after the "York" join).
- A `TestRebuildEqualsUpkeep_*` case fails after STEP-7: upkeep and rebuild then disagree on a document, which means a dependent read something outside the digest.
- `TestFoldedQueryScoresLikeExact` or another existing `core/search` ranking test changes order or score after STEP-3.
- T-5's product totals are not 2,627 and 15 at the planning base.
- `FoldsInto` cannot be written without changing `Expands` or `NewAliasIndex`.
- T-9 measures over 1 s on a Mac with the narrowed reprojection.
- A Swift test outside §4 Swift tests fails because it relied on the fake matching query text.

## 12. Open questions

None.

## 13. PR

**Title:** S14-21 — Search finds every spelling
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-21-search-finds-every-spelling.md`.
