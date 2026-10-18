# S11-07 — One normalization key across the core

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: search ranking, the search index, the term duplicate check and alias storage change in Go; the one visible change is the wording of an existing error)
**Board:** —
**Depends on:** S11-05 (landed [#451](https://github.com/mendahu/provenencia/pull/451)); planned after S11-06 (landed [#456](https://github.com/mendahu/provenencia/pull/456)), whose reconciler reads the text key this PR changes
**Blocks:** —
**Branch:** `s11-07-normalization-key`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md) (projection version and field scoring only), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md) (alias Install), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [reuse across the app](../comparison-layers.md#reuse-across-the-app) (text-equality table, "Found in passing"), [relations](../comparison-layers.md#relations); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5 text row; [`docs/ideas/term-management.md`](../../../ideas/term-management.md) item 3; [S11-03 plan](S11-03-value-compare-names.md) D-8 and Out ("normalizing alias tokens at storage"); [S11-06 plan](S11-06-migrate-reconciler.md) D-2; working rules "A normalized `Key` per value type", "High tolerance for churn", "One match registry"
**Planned:** 2026-10-17 against `main` @ `aa60b63`
**Review:** [`S11-07-normalization-key.review.md`](S11-07-normalization-key.review.md) — Fits / Ready with changes → fixed, 2 rounds

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

Search, the search index, the term duplicate check and the alias tables decide "the same text" with value compare's `Key` instead of their own lowercase and slug rules. A researcher who types `Montreal` gets `Montréal` ranked as a title hit, not as a typo. A Person or Place search document stops repeating a name that differs from its title only by case, accents or punctuation. Creating or renaming a term to `god-father` when the Property already has `Godfather`, seeded or the researcher's own, is refused with the existing term's label. Alias tokens are stored and indexed as the comparison tokens names produce, so `Jim` and `Mary-Ann` hit whatever their case or punctuation.

To catch `god-father` / `godfather`, text value compare stops caring where words break: two texts whose words, joined with no separator, are equal are `Same` (user call 2026-10-17). The reconciler groups on `Same`, so `York Town` and `Yorktown` become one value, as `Sainte-Anne` and `Sainte Anne` already are.

**Why now.** S11-02 built `Key` and S11-06 moved the reconciler onto it; search, the index, terms and aliases still keep four other text rules, and the map's slice Check ("`Montréal` / `Montreal` … compare the same way in Promote, search and the reconciler") needs search on the same rule.

**What stays.** Search's retrieval: FTS `MATCH` text, the trigram shortlist and the Jaro–Winkler gate (`core/search/fuzzy.go`) are unchanged. Name keys (`valuecompare` `name.key`) are unchanged. Term keys (`slug.Kebab`) and the `property_terms` unique key are unchanged. `properties`, `metadatafields` and `sourcetypes` keep their slug-only duplicate checks. No schema, wire or Swift logic change.

## 2. User stories

**US-1.** As a researcher, text that differs only by case, punctuation, spacing or where its words break is the same text everywhere.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `valuecompare.Compare` text `god-father` / `godfather` | `Same`, 1, cause `case`; `Parts` one pair `god father` / `godfather`, `Same`, 1, cause `case` |
| AC-1.2 | Text `York Town` / `Yorktown` | `Same`, 1, cause `case` |
| AC-1.3 | Text `Mont-réal` / `Montreal` | `Equivalent`, 0.95 (`value.Words.FoldCredit`), cause `fold`; `Parts` one pair `mont réal` / `montreal`, `Equivalent`, 0.95, cause `fold` |
| AC-1.4 | `Key` at `Same` of `god-father`, `God Father`, `godfather` | All `godfather`; at `Equivalent`, `Mont-réal` and `Montreal` are both `montreal` |
| AC-1.7 | Text `Mont-réal` / `montréal`; text `ab c` / `a bc` | Each `Same`, 1, cause `case`, `Parts` one joined pair (`mont réal` / `montréal`; `ab c` / `a bc`), `Same`, 1, cause `case` |
| AC-1.5 | Text `Upper Canada, York` / `York, Upper Canada` | Still `Different`, cause `order`; keys at both levels differ |
| AC-1.6 | Reconciler, text candidates `Yorktown` (obs 1), `York Town` (obs 2), single cardinality | One value, members 1 and 2, state `merged`, displayed `Yorktown` |

**US-2.** As a researcher, a search ranks a hit that matches once accents and punctuation are folded like a hit typed exactly.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Source titled `Montréal parish register`; search `Montreal`, then `Montréal` | The Source is the first hit both times, `MatchReason` `title`, and the two scores are equal |
| AC-2.2 | Place handle with toponyms `Montréal` and `Montreal` (title `Montréal`); search `Montreal`, kind place | One hit, title `Montréal`, `MatchReason` `title`, `MatchSnippet` empty |
| AC-2.3 | Place handle with toponyms `Montréal` and `Ville-Marie`; search `Ville Marie`, kind place | One hit, title `Montréal`, `MatchReason` `place`, `MatchSnippet` `Ville-Marie` |
| AC-2.4 | `scoreFields` for a Source spec with title `Sainte-Anne register` and query tokens of `sainteanne` (retrieval is not involved) | Score above 0, field `title` |

**US-3.** As the engine, a handle's search document lists each alternate name once and never repeats its title.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `matchText` for a place: header `Names` `[Montréal, Ville-Marie]`, `Parents` `[Québec]`; cached toponyms `[Montréal, Montreal, Ville-Marie, Ville Marie]` | Title `Montréal`; secondary exactly `["toponym:\tVille-Marie", "Québec"]` |
| AC-3.2 | `matchText` for a person: header name form `José Robins`; cached names `[José Robins, Jose Robins, JOSÉ ROBINS, Joe Robins]`; no life facts | Title `José Robins`; secondary exactly `["name:\tJoe Robins"]` |
| AC-3.3 | Same place as AC-3.1 with handle label `montreal` | The label adds no line |

**US-4.** As a researcher, creating or renaming a term to one the Property already has is refused and names the term I already have.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Property has user term `Godfather`; create `god-father` | `ErrDuplicateKey`, params `["Godfather"]`; no row written |
| AC-4.2 | Property has provenencia term `Godparent` (key `godparent`); create `God Parent` | `ErrDuplicateKey`, params `["Godparent"]` |
| AC-4.3 | Property has user term `Café`; create `Cafe` | `ErrDuplicateKey`, params `["Café"]` |
| AC-4.4 | Property has user term `Godfather`; create `Godmother` | Created |
| AC-4.5 | Property has user terms `Godfather` and `Godmother`; update `Godmother`'s label to `GODFATHER` | `ErrDuplicateKey`, params `["Godfather"]`; label unchanged |
| AC-4.6 | Update `Godfather`'s label to `God-Father` | Updated (a term is never its own duplicate) |
| AC-4.7 | Property has user term `A+B` (key `a-b`, text key `a+b`); create `A B` (key `a-b`, text key `ab`) | `ErrDuplicateKey`, params `["A+B"]` (the slug collision is still caught) |
| AC-4.8 | The macOS term editor shows `propertyterms.duplicate_key` with param `Godfather` | "This property already has the term “Godfather”. Use that one, or give this one a different label." |

**US-5.** As the engine, alias tokens are stored and looked up as the comparison tokens a name produces.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `Register` a nickname table with set `{"Jim", "JAMES"}`, then `Install` | `name_alias_members` holds `jim` and `james` for that set; given `Jim` against given `James` scores `value.Words.AliasCredit` |
| AC-5.2 | `Register` a set `{"Jim", "jim", "James"}`, then `Install` | No error; the set's stored tokens are exactly `james`, `jim` |
| AC-5.3 | `NewAliasIndex` over a set `{"Mary-Ann", "Polly"}` with the given role | Given `Maryann` against given `Polly` scores `AliasCredit` |
| AC-5.4 | `NewAliasIndex` over a set `{"Mc Donald", "MacDonald"}` with the untyped role | Untyped `McDonald` against untyped `MacDonald` scores `AliasCredit` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Text compare and text `Key`: words joined with no separator decide `Same` and `Equivalent` (D-1); `valuecompare.TextKey` (D-2) | STEP-1, AC-1.1 – AC-1.5, AC-1.7 |
| Reconciler cache rebuilt for the new text `Same` (D-3) | STEP-2, AC-1.6 |
| `core/search` Go rescoring (`scoreFields`, `alternateMatch`) on `TextKey` at `Equivalent` (D-4, D-5) | STEP-3, AC-2.1 – AC-2.4 |
| `searchindex` alternate-value dedupe on `TextKey` at `Equivalent`; `ProjectionVersion` 8 (D-6) | STEP-4, AC-3.1 – AC-3.3 |
| `propertyterms.Create` and `Update` duplicate check on `TextKey` at `Equivalent`, any origin; error names the existing label (D-7, D-8) | STEP-5, AC-4.1 – AC-4.8 |
| `namealiases` writes and `NewAliasIndex` key tokens with `TextKey` at `Same` (D-9) | STEP-6, AC-5.1 – AC-5.4 |
| Docs that the new text rule makes false (D-10) | STEP-7 |

| Out | Home |
| --- | --- |
| Search's fuzzy ranker (`jaroWinkler`, trigram shortlist), FTS tokenizers | Not planned — it ranks partial queries as typed ([reuse inventory](../comparison-layers.md#reuse-across-the-app) "Lighter touch") |
| Swift filters, `FakeStore` name grouping and term duplicate check | S11-08 (filters, name grouping); the fake's term duplicate check is not planned |
| Centring a search snippet on a folded hit (a note that matches only after folding shows its head, as a note with no raw hit does today) | Not planned |
| Duplicate checks in `properties`, `metadatafields`, `sourcetypes` (slug only) | Not planned — the map scopes term dedupe only |
| Merging existing duplicate terms | [`docs/ideas/term-management.md`](../../../ideas/term-management.md) item 3 |
| An alias editor or user alias tables | [`docs/ideas/international-names.md`](../../../ideas/international-names.md) (not planned in this spike) |
| `autoreconcile.SortKey` (`strings.ToLower` for text, `NormalizeForm` for names) | Not planned — it orders values, it does not decide equality |
| Name keys ignoring word breaks between parts | Not planned — names compare by part type, and particles already rejoin (`Mc Donald` = `McDonald`) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Text comparer | `core/valuecompare/text.go` `text.compare` | `Same` only when the word lists are equal (`slices.Equal(wa, wb)`); `Equivalent` when the folded lists are equal. A dash splits words, so `god-father` (`[god father]`) and `godfather` are `Different` (`no_match`) |
| Text key | `core/valuecompare/text.go` `text.key` | Words joined by one space (`god father`); folded at `Equivalent` |
| Key entry | `core/valuecompare/valuecompare.go` `Key(v Value, level Relation, rules Rules)` | Callers build a `Value{Type: propertymeta.ValueTypeText, Text: s}` by hand; no text helper |
| Word helpers | `core/valuecompare/words.go` `SplitWords`, `FoldWord`, `NormalizeForm` | `SplitWords` is the comparison tokenizer (dashes and slashes split, particles rejoin) |
| Alias index | `core/valuecompare/aliases.go` `NewAliasIndex(tables []AliasTable) AliasIndex` | Normalizes each token with `NormalizeForm`, so `Mc Donald` indexes as `mc donald` and never meets the comparison token `mcdonald`; `Mary-Ann` indexes as `mary ann` and never meets `maryann` |
| Reconciler text grouping | `core/autoreconcile/modules.go` `moduleFor` (text and integer: `valuecompare.Key(…, Same, rules)`) | Groups on the text key; picks up D-1 with no code change |
| Reconciler cache | `core/database/autoreconciler/autoreconciler.go` `CacheVersion = 16` | A grouping rule change needs a bump (skill `add-match-rule` "The reconciler composes value compare") |
| Search rescoring | `core/search/registry.go` `scoreFields`, `alternateMatch`, `snippetAround`; `core/search/fts.go` `Search` (`scoreTokens`, L111–L116) | `strings.Contains(strings.ToLower(field), tok)`: `montreal` is not in `montréal`, so the Source falls to the 0.1 floor or the fuzzy path |
| Query tokens | `core/search/registry.go` `tokenize` | Lowercased, edge punctuation trimmed; feed FTS `MATCH`, trigram shortlist and rescoring |
| Index documents | `core/database/searchindex/handles.go` `matchText`, `extraValues`, `tagged` | `extraValues` drops a value byte-equal to the title only. A place lists `Names[1:]` toponyms, then `Parents`, then every cached toponym but the title, so `Montreal` and `Ville-Marie` repeat; a person lists every cached name form but the title |
| Projection version | `core/database/searchindex/searchindex.go` `ProjectionVersion = 7` | Bumped when the projector's output changes, so `EnsureCatalog` rebuilds |
| Term create | `core/database/propertyterms/propertyterms.go` `Create` | Refuses only when `slug.Kebab(label)` equals a **user** term's key under the Property: `god-father` / `godfather` and `Café` / `Cafe` pass, and a user `Godparent` beside the seeded one passes |
| Term update | `core/database/propertyterms/propertyterms.go` `Update` | No duplicate check |
| Term error | `ErrDuplicateKey` (`apperr.CodePropertyTermsDuplicateKey` = `propertyterms.duplicate_key`), param: the minted key | Swift: `macos/App/Platform/L10n.swift` `L10n.Errors.propertyTermsDuplicateKey(key:)` (L9390) called at L9660; `macos/App/Resources/Localizable.xcstrings` `error.propertyterms.duplicate_key` "You already have a term with the key %@. Give this one a different label." |
| Term handler | `api/ffi/handlers/property_terms.go` L56 (`Create`), L87 (`Update`) | Passes the error through; no change |
| Alias storage | `core/database/namealiases/namealiases.go` `Install`; `core/database/namealiases/add.go` `insertSet`, `insertMissing`, `tokenSet`, `addProductNicknames` | Store tokens as given; `name_alias_members` has primary key `(table_id, set_idx, token)` and `CHECK (token <> '')` (`core/database/migrations/000049.sql`) |
| Alias rules | `core/database/namealiases/namealiases.go` `withAliases` | The one non-test caller of `NewAliasIndex` |
| Tests | `core/valuecompare/text_test.go` `TestTextCompare`, `TestTextKey`; `core/valuecompare/contract_test.go` `TestComparerContract`, `TestKeyContract` (`samples`, `generators`); `core/valuecompare/words_test.go` L74, L123; `core/valuecompare/name_test.go` L85, L342; `core/autoreconcile/modules_test.go` L205; `core/database/namealiases/registry_test.go` L23, `namealiases_test.go` `TestUserAndPluginRowsScore`; `core/search/search_test.go` `TestSearchAccentedPlaceTypo`; `core/search/handles_test.go` `TestAlternateToponymIsMatchContext`; `core/database/searchindex/searchindex_test.go` (package `searchindex`); `core/database/propertyterms/propertyterms_test.go` `TestPropertyTerms` "duplicate user key" | `TestAlternateToponymIsMatchContext` asserts `Montreal` matches `Montréal` through the alternate line (reason `place`); D-4 and D-6 change that to a title hit (AC-2.2) |

**Precedent.** [#456](https://github.com/mendahu/provenencia/pull/456) (S11-06) moved the reconciler's own text rule onto `valuecompare.Key` and bumped `CacheVersion`: copy how `moduleFor` calls `Key` with the registry's rules and how the cache comment records the rule. [#447](https://github.com/mendahu/provenencia/pull/447) (S11-03, D-8) built `AliasIndex` and moved `match` onto it: copy its test style for alias hits (`name_test.go` L85 option). Do not copy S11-06's `Rules` parameter threading: search, the index and terms read text keys only, which use no catalog data (D-5).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Text ignores word breaks (user call 2026-10-17).** In `text.compare`, after the empty check: `ja, jb := strings.Join(wa, ""), strings.Join(wb, "")`. If `ja == jb` → `Same`, 1; cause `exact` when `a.Text == b.Text`, else `case`. Else `fa, fb` folded; if `strings.Join(fa, "") == strings.Join(fb, "")` → `Equivalent`, cause `fold`. **Parts:** for `Same` when `slices.Equal(wa, wb)`, and for `Equivalent` when `slices.Equal(fa, fb)`, today's `positionalParts` and, for `Equivalent`, today's `meanSimilarity` similarity; otherwise (the words break in different places, even at equal counts: `ab c` / `a bc`), one `PartMatch` whose `A.Text` and `B.Text` are the words joined by one space, kind `PartWord`, with the value's relation and cause, similarity 1 for `Same` and `r.Words.FoldCredit` for `Equivalent`, and the value's similarity is that part's. **Key:** `text.key` joins the words (folded at `Equivalent`) with `""`. The `MaxWords` bound stays after these checks | `god-father` differs from `godfather` only by punctuation, which S11-06 decided is one value; the key and compare change together, so `TestKeyContract` still holds; `case` already names "same after normalizing" (`Sainte-Anne` / `Sainte Anne`), so no new cause code reaches the wire vocabulary | `Equivalent` for word breaks — the user chose `Same` so the reconciler groups `York Town` / `Yorktown`; a separate compact key for terms only — two text rules again |
| D-2 | **`valuecompare.TextKey(s string, level Relation, w WordRules) (string, bool)`** *(new)* in `text.go`: `return Key(Value{Type: propertymeta.ValueTypeText, Text: s}, level, Rules{Words: w})`; doc comment: a text key reads only the word rules. Every caller outside `valuecompare` in this PR uses it | Four callers (search, index, terms, alias tokens) would each build the same `Value`; the signature says a text key depends on the word rules alone | A local helper per package — four near-copies |
| D-3 | **`CacheVersion` 17** with the comment line `//	17: text ignores word breaks: god-father is godfather (S11-07).` No reconciler code changes | The cache groups on text `Same` (`moduleFor`); a rule change rebuilds the cache on open | Leaving the version — an open catalog would keep `York Town` / `Yorktown` split until another rebuild |
| D-4 | **Search rescoring keys at `Equivalent`.** `Search` keeps `tokens` (from `tokenize`) for FTS `MATCH`, the trigram shortlist and the ref checks. Rescoring takes `queryTokens{raw, keys []string}` *(new, `registry.go`)*, built once in `Search` by `newQueryTokens(scoreTokens)` *(new)*: `raw` is `scoreTokens` as today, `keys` each token's `TextKey(tok, Equivalent, valueWords)`, dropping tokens with no key. `scoreFields(spec, values, q queryTokens)` keys each non-empty field value once per call with `TextKey(…, Equivalent, valueWords)` and tests `strings.Contains(fieldKey, key)` for each of `q.keys`; coverage divides by `len(q.keys)`. `alternateMatch(secondary, q)` keys the text after the `name:\t` / `toponym:\t` tag and tests `q.keys` the same way. `matchSnippetForField` and `snippetAround` keep today's code and receive `q.raw`, so an exact query centres its snippet as today; a hit that matches only after folding shows the value's head. **Accepted:** a field key has no spaces (D-1), so a token can match across a word break (`tea` in `sainteanne`); substring matching already crossed letters inside a word (`ann` in `joanna`), and FTS still decides retrieval | FTS already folds accents (`unicode61 remove_diacritics 2`), so rescoring that does not fold ranks a found hit below an exact one; `Equivalent` is the level FTS retrieves at | `Same` level — `Montreal` would still miss `Montréal`; changing FTS or fuzzy tokens — Out |
| D-5 | **Search, the index, terms and alias storage use the shipped word rules**, `matchregistry.Default().Value.Words`, held in each package as `var valueWords = matchregistry.Default().Value.Words` *(new)* with the doc comment "valueWords are the shipped registry's word rules, never catalog data: a text key reads nothing else" (`core/search`, `core/database/searchindex`, `core/database/propertyterms`, `core/database/namealiases`) | A text key reads only the word rules (particles, folds), which the catalog never supplies; the reconciler cache holds the shipped registry the same way (`autoreconciler` `valueRules`) | `valuerules.Load` — it reads name formats and alias tables these callers never use, per query |
| D-6 | **Index dedupe at `Equivalent`.** `distinctAlternates(title string, lists ...[]string) []string` *(new, `handles.go`)* walks the lists in order and keeps a value whose `TextKey(…, Equivalent, valueWords)` is neither the title's key nor an earlier kept value's; a value with no key is dropped. `extraValues` is deleted. Person: secondary = `personLifeMatch(p)`, then `tagged(distinctAlternates(title, v[name]), "name")`. Place: secondary = `tagged(distinctAlternates(title, p.Names[1:], v[toponym]), "toponym")`, then `p.Parents`. The handle label stays last, as today, and is added only when its key differs from the title's and from every kept alternate's. A title with no text key (empty, or only punctuation) excludes nothing, so every alternate with a key is kept. `ProjectionVersion` 8 | Once rescoring folds (D-4), an alternate equivalent to the title can neither retrieve nor be the match reason; the place list repeated `Names[1:]` through the cached toponyms | `Same` level — keeps `Montreal` beside `Montréal` as a line no query can reach; leaving the order — toponyms split around the parents |
| D-7 | **Term duplicates (user call 2026-10-17).** `conflictingTerm(tx *database.Tx, propertyID, selfID []byte, label, slugKey string) (Term, bool, error)` *(new, `propertyterms.go`)* runs `sqlListByProperty` and scans with `scanTermRows`, in one pass, in that query's order (`label COLLATE NOCASE, origin, key`), skipping `selfID`, and returns the first row that either has `TextKey(row.Label, Equivalent, valueWords)` equal to the label's key (any origin), or, when `slugKey` is non-empty, is user-origin with `Key == slugKey`. `Create` calls it with `selfID` nil and `slugKey` = `slug.Kebab(label)` after the lock check and the empty-slug check, replacing the `sqlLookup` duplicate check; `Update` calls it after the lock check, only when `existing.Label != label`, with the trimmed label, `existing.PropertyID`, the term's id and `slugKey` empty; a keyless label is `ErrInvalid` at that point. Both return `ErrDuplicateKey.WithParams(row.Label)`. A label with no text key (`!!!`) is `ErrInvalid` in both, as an empty slug is today | A seeded and a researcher term that read the same are one term to the researcher; a rename is the other way in; the slug check still guards the `ON CONFLICT … DO UPDATE` upsert, which would overwrite a user term whose label keys differ but slugs match (`A+B` / `A B`) | User terms only — the user chose any origin; Create only — the user chose rename too; a SQL `LIKE` — cannot fold or join words |
| D-8 | **Copy (user call 2026-10-17).** `L10n.Errors.propertyTermsDuplicateKey(label:)` (renamed parameter); `error.propertyterms.duplicate_key` = "This property already has the term “%@”. Use that one, or give this one a different label."; comment "FFI error propertyterms.duplicate_key; argument is the existing term's label". The error code and `ErrDuplicateKey` name stay; its doc comment says "a term of the Property already reads the same as the label" | Naming the term the researcher already has tells them which to use | A new error code — the handler, Swift mapping and FFI vocabulary would change for the same failure |
| D-9 | **Alias tokens are comparison tokens.** `valuecompare.AliasToken(token string, w WordRules) (string, bool)` *(new, `aliases.go`, beside `Hit`)* is `TextKey(token, Same, w)`: the one rule for what an alias token is. `NewAliasIndex(tables []AliasTable, w WordRules) AliasIndex` keys each token with it (replacing `NormalizeForm`); empty keys and empty sets are dropped as today. `withAliases` passes `base.Words`; the six test callers pass `matchregistry.Default().Value.Words`. In `namealiases`, `aliasTokens(tokens []string) []string` *(new, `add.go`)* returns the distinct `AliasToken(…, valueWords)` keys in first-seen order, dropping empty ones; `Install`, `insertSet`, `insertMissing` and `tokenSet` write or compare those keys only | `wordSimilarity` calls `Hit` with `SplitWords` tokens, so the index must key with the same tokenizer (particles rejoin, word breaks vanish); storing keys keeps `addProductNicknames`' subset test and the primary key on one spelling (`Jim` / `jim` in one set would violate it) | Normalizing at storage only — rows written by hand or a future editor would still bypass it; keeping `NormalizeForm` in the index — `Mc Donald` and `Mary-Ann` never hit |
| D-10 | **Tests:** valuecompare cases in the existing tables (`TestTextCompare`, `TestTextKey`, contract `samples`); search and index cases as Go integration tests on a real catalog (`-tags fts5`); terms as rows in `TestPropertyTerms`; no Swift test (copy only, checked by hand) | Matches each package's existing test layer | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuecompare.TextKey` *(new)* | func | `valuecompare.Key` | It **is** `Key` for text; the helper saves five callers from building a `Value` by hand (D-2) |
| `valueWords` in `search`, `searchindex`, `propertyterms`, `namealiases` *(new)* | var | `autoreconciler` `valueRules` | Same concept and doc comment (the shipped registry, never catalog data); each package keys text in a loop (D-5) |
| `search.queryTokens`, `newQueryTokens` *(new)* | type, func | `search.tokenize` | `tokenize` feeds FTS raw tokens; rescoring needs raw tokens for snippets and keys for matching, each with one job (D-4) |
| `searchindex.distinctAlternates` *(new)* | func | `searchindex.extraValues` | Replaces it (deleted): `extraValues` compared bytes against the title only |
| `propertyterms.conflictingTerm` *(new)* | func | `propertyterms.Create` slug lookup (`sqlLookup`), `ListByProperty` | Reuses `sqlListByProperty` and `scanTermRows`; one pass replaces the slug lookup as a duplicate check, and keeps the slug rule for the upsert's sake (D-7) |
| `valuecompare.AliasToken` *(new)* | func | `NewAliasIndex`'s `NormalizeForm` call | The index and storage share one token rule beside `Hit` (D-9) |
| `namealiases.aliasTokens` *(new)* | func | `namealiases.tokenSet` | `tokenSet` now builds on it; distinct `AliasToken` keys for writes and the subset test (D-9) |

**Builds on.** `valuecompare.Key` and `SplitWords` (S11-02), `AliasIndex` (S11-03 D-8), the reconciler's `Key` grouping (S11-06), the search registry's field scoring (`add-searchable-kind`), `ProjectionVersion` and `CacheVersion` rebuilds.
**Deletes.** `searchindex.extraValues`; the `NormalizeForm` call in `NewAliasIndex`; the `strings.ToLower` + `strings.Contains` pairs in `scoreFields` and `alternateMatch`; `Create`'s `sqlLookup` duplicate check; the old duplicate-key copy.
**Next adds.** S11-08 mirrors D-1 in the shared Swift normalize helper and `FakeStore`. S11-15's shortlist index keys text with `TextKey`. A duplicate check for Properties, metadata fields or Source types (Out) is one `conflictingTerm`-shaped call.

### Schema

None — no catalog change. `name_alias_members` rows of an existing catalog are not rewritten: the product seed is already lowercase single words, and D-9's index keys whatever is stored.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `text.compare` and `text.key` per D-1; `func TextKey(s string, level Relation, w WordRules) (string, bool)` *(new)*; `func AliasToken(token string, w WordRules) (string, bool)` *(new)*; `func NewAliasIndex(tables []AliasTable, w WordRules) AliasIndex` (new parameter, D-9) |
| `core/database/autoreconciler` | `CacheVersion = 17` and its comment line (D-3) |
| `core/search` | `valueWords` *(new)*; `queryTokens`, `newQueryTokens` *(new)*; `scoreFields`, `alternateMatch` per D-4; `Search` builds `newQueryTokens(scoreTokens)` once and passes it to rescoring |
| `core/database/searchindex` | `valueWords` *(new)*; `distinctAlternates` *(new)*; `matchText` per D-6; `extraValues` deleted; `ProjectionVersion = 8` |
| `core/database/propertyterms` | `valueWords` *(new)*; `conflictingTerm` *(new)*; `Create`, `Update` per D-7; `ErrDuplicateKey` doc comment |
| `core/database/namealiases` | `valueWords` *(new)*; `aliasTokens` *(new)*; `Install`, `insertSet`, `insertMissing`, `tokenSet`, `withAliases` per D-9 |

Errors: `propertyterms.duplicate_key` → a label whose text key equals another term's under the Property (create or rename), or whose slug equals a user term's key (create); param: that term's label.

### FFI

None — the handler passes the error through unchanged.

### macOS

None — no view changes. The term editor already shows `L10n.Errors.propertyTermsDuplicateKey`; only its wording and parameter name change (D-8).

### Strings

`error.propertyterms.duplicate_key` value and comment change per D-8; `L10n.Errors.propertyTermsDuplicateKey(key:)` becomes `propertyTermsDuplicateKey(label:)`, and its caller at `L10n.swift` L9660 passes `label: params.first ?? "?"`. No key added or removed.

## 7. Steps

**STEP-1 — Text ignores word breaks; `TextKey`**
Files: `core/valuecompare/text.go`, `core/valuecompare/text_test.go`, `core/valuecompare/contract_test.go`. Skill: `add-match-rule`.
Do: D-1, D-2. Add T-1 – T-3.
Check: `go test ./core/valuecompare/...` passes with no tags; `CGO_ENABLED=1 go test -tags fts5 ./core/match/... ./core/graphalign/... ./core/database/promotealign/...` passes with no expected value changed.

**STEP-2 — Reconciler cache rebuild**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/autoreconcile/modules_test.go`. Skill: `add-match-rule`.
Do: D-3. Add T-4 beside the `Sainte-Anne` rows.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/... ./core/database/autoreconciler/...` passes.

**STEP-3 — Search rescoring on the key**
Files: `core/search/registry.go`, `core/search/fts.go`, `core/search/search_test.go`, `core/search/handles_test.go`. Skill: `add-searchable-kind`.
Do: D-4, D-5. Add T-5, T-6; keep T-8; rewrite `TestAlternateToponymIsMatchContext` as T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/search/...` passes; `TestSearchAccentedPlaceTypo`, `TestNoteMatchReasonUsesSnippet` and `TestNoteBodyRollsIntoSourceHit` pass unchanged.

**STEP-4 — Index dedupe on the key**
Files: `core/database/searchindex/handles.go`, `core/database/searchindex/searchindex.go`, `core/database/searchindex/searchindex_test.go`. Skill: `add-searchable-kind`.
Do: D-6. Add T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/... ./core/search/...` passes; `rg -n 'func extraValues' core` prints nothing.

**STEP-5 — Term duplicates on the key**
Files: `core/database/propertyterms/propertyterms.go`, `core/database/propertyterms/propertyterms_test.go`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-7, D-8. Add T-10 rows.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/propertyterms/... ./api/ffi/...` passes; `rg -n 'propertyTermsDuplicateKey\(key:' macos` prints nothing.

**STEP-6 — Alias tokens are comparison tokens**
Files: `core/valuecompare/aliases.go`, `core/valuecompare/words_test.go`, `core/valuecompare/name_test.go`, `core/autoreconcile/modules_test.go`, `core/database/namealiases/namealiases.go`, `core/database/namealiases/add.go`, `core/database/namealiases/registry_test.go`, `core/database/namealiases/namealiases_test.go`. Skills: `add-match-rule`, `add-seeded-vocabulary`.
Do: D-9 (`AliasToken` in `aliases.go`). Add T-11, T-12.
Check: `go test ./core/valuecompare/...` passes with no tags; `CGO_ENABLED=1 go test -tags fts5 ./core/database/namealiases/... ./core/database/valuerules/... ./core/autoreconcile/...` passes; `rg -n 'NormalizeForm' core/valuecompare/aliases.go` prints nothing.

**STEP-7 — Docs**
Files: `docs/matching-model.md` (glossary `Same`: "case, whitespace, punctuation, and where words break, for text"), `docs/deployment-plan/spike-11/comparison-layers.md` (relations table text row: `Same` "same letters, word breaks ignored"; reuse inventory: the "Found in passing" text-equality line says reconciler, search scoring and term dedupe share `Key` after S11-07, `core/match` goes in S11-13, and FTS and the header sort (`SortKey`, ordering only) keep their own rules on purpose), `.agents/skills/add-match-rule/SKILL.md` (two lines: callers outside matching that need a catalog-independent text key call `valuecompare.TextKey` with the shipped word rules, not `valuerules.Load`; a change to `text.key` or `WordRules` bumps `autoreconciler.CacheVersion` and `searchindex.ProjectionVersion`), `deployment-plan.md` S11-08 scope entry (add "text ignores word breaks (S11-07 D-1)" to its Ships line), `docs/conclusion-reconciliation.md` §5 text row (add *god-father* = *godfather*), `docs/ideas/term-management.md` item 3 (the check runs on create and rename, against every term of the Property), `completed.md`, `deployment-plan.md` (map row status), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-07-normalization-key.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/text_test.go` `TestTextCompare` (rows added) | `god-father` / `godfather`; `York Town` / `Yorktown`; `Mont-réal` / `Montreal`; `Mont-réal` / `montréal`; `ab c` / `a bc`; `Upper Canada, York` / `York, Upper Canada` unchanged → as AC-1.1, 1.2, 1.3, 1.7, 1.5 | AC-1.1, AC-1.2, AC-1.3, AC-1.5, AC-1.7, D-1 |
| T-2 | Go | `core/valuecompare/text_test.go` `TestTextKey` (cases added; uses `TextKey`) | `god-father`, `God Father`, `godfather` at `Same` → `godfather`; `Mont-réal`, `Montreal` at `Equivalent` → `montreal`; `Montréal` at `Equivalent` → `montreal`; reordered toponym keys still differ | AC-1.4, AC-1.5, D-2 |
| T-3 | Go | `core/valuecompare/contract_test.go` `samples` text row | Pairs `god-father` / `godfather`, `ab c` / `a bc` and `Mont-réal` / `Montreal` added; `TestComparerContract` and `TestKeyContract` pass | D-1 |
| T-4 | Go | `core/autoreconcile/modules_test.go` `TestModules` (row added) and `TestCompatible` (row added) | `Yorktown` (1), `York Town` (2) → `[[1 2]]`, `StateMerged`, displayed `Yorktown`; `Compatible` → `true` | AC-1.6, D-3 |
| T-5 | Go | `core/search/search_test.go` `TestFoldedQueryScoresLikeExact` *(new)* | Source `Montréal parish register`; `Search` `Montreal` and `Montréal` → same first hit, reason `title`, equal `Score` | AC-2.1, D-4 |
| T-6 | Go | `core/search/search_test.go` `TestScoreFieldsJoinsWordBreaks` *(new, package `search`)* | `spec, _ := kindSpec(KindSource)`; `scoreFields(spec, {title: "Sainte-Anne register"}, newQueryTokens(["sainteanne"]))` → score > 0, field `title` | AC-2.4, D-1, D-4 |
| T-7 | Go | `core/search/handles_test.go` `TestAlternateToponymIsMatchContext` (rewritten) | Place `Montréal` + `Montreal`, search `Montreal` → reason `title`, snippet empty; place `Montréal` + `Ville-Marie`, search `Ville Marie` → reason `place`, snippet `Ville-Marie` | AC-2.2, AC-2.3, D-4, D-6 |
| T-8 | Go | `core/search/search_test.go` `TestSearchAccentedPlaceTypo`, `TestNoteMatchReasonUsesSnippet` (unchanged) | `Munchen` finds `München` first; a note hit keeps its centred snippet | D-4 |
| T-9 | Go | `core/database/searchindex/searchindex_test.go` `TestMatchTextListsEachAlternateOnce` *(new, table-driven)* | AC-3.1 place, AC-3.2 person, AC-3.3 label → exact title and secondary slices | AC-3.1, AC-3.2, AC-3.3, D-6 |
| T-10 | Go | `core/database/propertyterms/propertyterms_test.go` `TestPropertyTerms` (rows added on the fixture's `propID`; the provenencia `Godparent` of AC-4.2 is written with `Upsert` and origin `vocabkey.OriginProvenencia`; the "duplicate user key" row now also asserts params `["Dup"]`, the label, where today's param was the key `dup`) | AC-4.1 – AC-4.7 each as a row: error is `ErrDuplicateKey` with the listed params (`(*apperr.Error).Params()`), or success | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, AC-4.6, AC-4.7, D-7 |
| T-11 | Go | `core/valuecompare/words_test.go` `TestAliasIndexKeysTokens` *(new)* | AC-5.3 and AC-5.4 sets → `AliasCredit` | AC-5.3, AC-5.4, D-9 |
| T-12 | Go | `core/database/namealiases/namealiases_test.go` `TestInstallStoresComparisonTokens` *(new)* | `Register` a table `{Key: "test-nicks", Origin: vocabkey.OriginUser, Label: "Test", Kind: valuecompare.AliasNickname, Roles: aliasPartRoles()}` with sets `{"Jim","JAMES"}` and `{"Jim","jim","James"}` (`t.Cleanup(ResetAliasTablesForTest)`), `Install`; read back with `Load` as `TestInstallAndLoad` does and score with `Rules` as `TestUserAndPluginRowsScore` does → stored tokens per AC-5.1, AC-5.2; `Jim` / `James` scores `AliasCredit` | AC-5.1, AC-5.2, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/search` ranking tests (scores rise for folded hits), `TestUserAndPluginRowsScore`, `TestAddProductNicknames*`, the obituary golden and sparse cases in `core/database/promotealign` (D-1 should not change them).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no chrome
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — n/a, not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md, conclusion-reconciliation.md §5, ideas/term-management.md, comparison-layers.md, add-match-rule skill, S11-08 scope entry (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source titled `Montréal parish register`. Type `Montreal` in the omnibar. The Source is the first hit and reads as a title match.
2. Open Properties, pick the Role Property (a term Property researchers may extend), add the term `Godfather`, then add `god-father`. The editor refuses with "This property already has the term “Godfather”. Use that one, or give this one a different label." (AC-4.8). Add `Godmother`, then rename it to `GODFATHER`: the same message. Rename `Godfather` to `God-Father`: it saves.
3. Open a project whose Place has the toponyms `Montréal` and `Montreal` on two Sources. Search `Montreal`: one Place hit, titled `Montréal`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| D-1 changes a Promote golden or sparse case (a text pair that differs only in word breaks) | STEP-1's check runs those suites with no expected value changed; a change is a Stop condition, not an update |
| Rescoring keys every field of every candidate | One `TextKey` per field per candidate (D-4), bounded by FTS candidates; no new query |
| A plugin-registered alias set has two tokens that key alike | `aliasTokens` dedupes before insert (D-9, AC-5.2) |
| `conflictingTerm` reads every term of the Property per create or rename | Term lists are small (tens); it runs inside the write transaction already open |

**Stop and report if:**

- `text.compare` or `text.key` no longer has the shape §4 describes (S11-06 or a later PR changed it).
- STEP-1's check needs an expected value changed in `core/match`, `core/graphalign` or `core/database/promotealign` tests.
- `TestKeyContract` fails on a generated text pair after D-1 (Key and Compare disagree).
- A search test other than `TestAlternateToponymIsMatchContext` needs its expected order or reason changed.
- `core/search`, `core/database/searchindex` or `core/database/namealiases` cannot import `core/matchregistry` without an import cycle (`go build ./...` reports one).
- `TestModules` in `core/autoreconcile/modules_test.go` no longer holds the `Sainte-Anne` / `Sainte Anne` row.

## 12. Open questions

None.

## 13. PR

**Title:** S11-07 — One normalization key across the core
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-07-normalization-key.md`.
