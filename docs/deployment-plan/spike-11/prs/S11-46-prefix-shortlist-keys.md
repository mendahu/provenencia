# S11-46 — Prefix keys in the shortlist

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the shortlist feeds the target menu that already lists a row's candidates; nothing new is drawn)
**Board:** —
**Depends on:** —
**Blocks:** S11-45
**Branch:** `s11-46-prefix-shortlist-keys`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [name candidates § Prefix rule, in detail](../name-candidates.md#prefix-rule-in-detail), [§ Decided](../name-candidates.md#decided-user-2026-10-24), [comparison layers § Layer 6 procedure](../comparison-layers.md#layer-6-procedure), [`matching-model.md` § Knob inventory](../../../matching-model.md#knob-inventory); working rules: "A prefix is an initial generalised", "Registry numbers: carry over, then estimate", "One match registry", "High tolerance for churn"
**Planned:** 2026-10-28 against `main` @ `35c1bb3`
**Review:** [`S11-46-prefix-shortlist-keys.review.md`](S11-46-prefix-shortlist-keys.review.md) — Fits with changes / Ready with changes, 2 rounds (every finding answered)

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

The shortlist hands graph alignment every record whose given name starts another, in both directions: a Source person "Des" reaches the record "Desiree", and a Source person "Desiree" reaches the record "Des". Today the two share no shortlist key, so no later rule can ever see the pair. This PR splits `valuecompare.WordKeys` into the keys a record is filed under (index) and the keys a Subject is looked up by (lookup), and adds prefix keys for the words of given, nickname and untyped name parts from `PrefixMinLength` (2) letters up. Surnames and text are never keyed by prefix.

**Why now.** Dogfood, 2026-10-24 ([case 2](../name-candidates.md#case-2--desiree-against-des)). The map review split the shortlist half out of S11-45 (K-2, K-7), and S11-45's comparer rule depends on it.

**What stays.** Every key the shortlist files and looks up today, on both sides: whole keys (`=`), words (`w:`) and alias sets (`a:`). The comparer: until S11-45 lands, "Des" against "Desiree" is still `Different`, so the new shortlist pair is dropped at the candidate gate and no row changes for that case. A word pair the comparer already relates (a one-letter spelling variant such as "Ann" / "Anne", 0.8 ≥ `FuzzyFloor`) that the shortlist never offered is now offered: the shortlist's guarantee gets closer to "every pair value compare relates shares a key".

## 2. User stories

**US-1.** As the engine, when a Subject's given name is the start of a record's given name, or the reverse, I shortlist the record, so that S11-45's prefix rule can compare them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Index: person 1 given `Desiree`, person 2 given `Desmond`, person 3 given `Mary`. Lookup: given `Des` | `[1 2]` |
| AC-1.2 | Index: person 1 given `Des`, person 2 given `Desmond`, person 3 given `De`. Lookup: given `Desiree` | `[1 3]`: `Desmond` is not a prefix of `Desiree`, nor the reverse |
| AC-1.3 | Index: person 1 given `Desiree`. Lookup: given `Desmond` | Nothing |
| AC-1.4 | Index: person 1 `John MacDonald` (Western: surname `MacDonald`). Lookup: `Mary Mac` (surname `Mac`), and given-only `Mac` | Nothing for either: a surname is never filed or looked up by prefix |
| AC-1.5 | `WordKeys` of given `J.` and of `J. Robins` | No `p:` or `g:` key: an initial is never a prefix key (initials stay `IsInitial`'s) |
| AC-1.6 | `WordKeys` of given `Desiree` with `PrefixMinLength` 0 | Index equals lookup and holds no `p:` or `g:` key: today's keys exactly |
| AC-1.7 | `WordKeys` of a given word of 33 folded runes (one over `MaxSpellingLength` 32) | No `p:` or `g:` key (D-4) |
| AC-1.8 | Text `New York`, `Desiree`, and toponyms through `shortlist.Keys` | Index equals lookup, no `p:` or `g:` key (D-9) |
| AC-1.9 | Every pair of the existing share corpus that `Compare` calls `Same` or `Equivalent` | One's index shares a key with the other's lookup, in both directions |
| AC-1.10 | The whole Go suite | Every existing test and golden passes. Existing tests change only as T-1 and T-2 say (the two-list signature), T-4 adds rows and a helper, and STEP-1 renames the `aliasPartRoles()` calls; `TestLargeSourceBudgets` stays inside its budgets |

## 3. Scope

| In | Covered by |
| --- | --- |
| `valuecompare.WordKeys` returns `(index, lookup []string)`; `wordKeyer.wordKeys` likewise; text returns one list on both sides (D-1) | STEP-2, AC-1.8, AC-1.9 |
| `shortlist.Keys` files the index list and looks up the lookup list (D-1) | STEP-2, AC-1.10 |
| Prefix keys for words of the alias roles: index files `g:<word>` and `p:<prefix>`, lookup asks `p:<word>` and `g:<prefix>` (D-2, D-4) | STEP-3, AC-1.1 – AC-1.7 |
| `valuecompare.AliasRoles()` *(new)*: the one list of the alias tables' roles, read by the product seed and the prefix keys; `namealiases.aliasPartRoles` deleted (D-3) | STEP-1 |
| `value.Names.PrefixMinLength` 2 in the registry with its inventory row (D-5) | STEP-3 |
| Extra keys and shortlist hits measured on S11-21's large fixture, no cap (D-6) | STEP-4, AC-1.10 |
| Docs and the matching skill say a value has index and lookup keys (D-8) | STEP-5 |

| Out | Home |
| --- | --- |
| The prefix comparer rule (`Contains` at `PrefixCredit`, `CausePrefix`), its wording on the sheet, the reconciler fold | S11-45 |
| `PrefixCredit` | S11-45 |
| Surname prefixes ("Mac" / "MacDonald") | Not planned ([what stays later](../name-candidates.md#what-stays-later)) |
| A spelling key for toponyms and other spelling-only pairs ("Provenance" / "Provenanced", `TestSpellingOnlyToponymIsNotShortlisted`) | Dogfood log: the test's own `dogfood:` comment |
| A cap on prefix keys or prefix hits | Not planned: the measurement does not call for one (D-6); revisited if dogfood shows a slow proposal |
| Raising any `TestLargeSourceBudgets` budget | Not planned: a stop condition |
| Counting probes (key and hit counters) in `core/shortlist` | Not planned: the planning measurement used a throwaway copy (D-6) |

## 4. Current state

Verified at `35c1bb3`.

| Area | Where | Today |
| --- | --- | --- |
| Word keys | `core/valuecompare/shortlist.go` `WordKeys`, `wordKeyer`, `keyList`, `name.wordKeys`, `text.wordKeys` | One sorted, duplicate-free list per value: `=`+the whole key at `Equivalent`, `w:`+each folded word that is not `IsInitial`, and for names `a:`+each alias set key. The doc comment says the list is what "shortlisting looks a value up by". No prefix keys |
| Name words | `core/valuecompare/name.go` `nameWords`, `nameWord` | Reads a name's words under its pattern with `text`, `fold`, `role` (`RoleFamily`, `RoleGiven`, `RoleNick`, else `RoleUntyped`); ignored and generation parts are not words |
| Shortlist keys | `core/shortlist/shortlist.go` `Keys` | Word leads append each `WordKeys` key, prefixed `lead|`, to both `index` and `lookup`. Dated kinds already key asymmetrically (`wide` / `any`). Doc comment: "Word leads key every word key of every positive value, the same on both sides" |
| Index | `core/shortlist/index.go` `Index.Add`, `Index.Lookup` | `Add` files under `Keys`' index list; `Lookup` reads its lookup list. Unchanged by this PR |
| Index loader | `core/database/promotealign/index.go` `loadShortlist`, `Warm` | Builds one index per catalog graph from every unmerged handle of the keyed kinds; dropped on a handle commit; built behind the open spinner (`Warm`). Unchanged |
| Alias roles | `core/database/namealiases/registry.go` `aliasPartRoles` | Returns given, nick, untyped; used by `productAliasTables` (two tables) and the in-package tests `registry_test.go` line 35 and `namealiases_test.go` line 152. A catalog's own tables store their roles in catalog rows (`namealiases.go` `Load`) |
| Name rules | `core/valuecompare/rules.go` `NameRules`; `core/matchregistry/value.go` `defaultValue` | No prefix field. `WordRules.MaxSpellingLength` (32) bounds spelling pairs; `words.go` `wordSimilarity` checks it with `utf8.RuneCountInString` |
| Inventory | `docs/matching-model.md` knob inventory; `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | `value.Names.*` rows end with `value.Names.Generation` (line 140). The guard fails on a registry field with no row |
| Word key tests | `core/valuecompare/shortlist_test.go` `TestWordKeys`, `TestWordKeysShareOnSameOrEquivalent`, helpers `billRules`, `slName`, `slText`, `shares` | One list per value; `TestWordKeys` has `want` (exact), `has`, `not`; the share test checks `shares(WordKeys(a), WordKeys(b))` over a corpus for names and text |
| Shortlist tests | `core/shortlist/shortlist_test.go` `TestLookup`, `TestKeys`, helpers `person` (Western), `build`, `ids`, `billRules` | Name rows use `person`, which reads a one-word form as a surname |
| Given-only fixtures | `core/namevalue/namevaluetest/namevaluetest.go` `Given`; `core/propertycompare/propertycomparetest/propertycomparetest.go` `Given` | Every word a given part (S11-43). `slName` and `person` read Western |
| Budgets | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `largeShape`, `zipfNames` (`bench_test.go`) | 4,000 persons plus a 12-generation tree, names `Giv<letters> Sur<letters>`. At the base, run alone here: warm at open 1.10 s (2.5 s), cold 1.08 s (2.5 s), warm 255 ms (350 ms), after a handle commit 1.18 s (1.5 s). Under the full `./core/... ./api/...` run: warm 318 – 340 ms, after a commit 1.27 – 1.42 s |

**Precedent.** `e8046e37` ("Add the shortlist keys and index a Subject is looked up by") gave dated kinds different index and lookup keys (`wide` on one side, `any` on the other) inside `shortlist.Keys`, with `TestLookup` rows that prove each direction. Copy that shape: a tag per side, a `TestLookup` row per direction. `db987d96` ("Key a date wider than the bucket cap as open") bounded the keys one hostile value makes; D-4's length bound copies it. Do not copy its new registry knob: D-4 reuses `MaxSpellingLength`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **`func WordKeys(v Value, r Rules) (index, lookup []string)`**, and `wordKeyer.wordKeys(v Value, r Rules) (index, lookup []string)`. Both lists sorted and duplicate-free. A type with no word keys, or a value with no word, returns `nil, nil`. `text.wordKeys` returns `keyList(…)` and a `slices.Clone` of it. `name.wordKeys` returns `keyList(…)` plus D-2's keys on each side. `shortlist.Keys` appends each index key, prefixed `lead|`, to `index`, and each lookup key to `lookup`. Doc comment of `WordKeys`: "WordKeys are the keys a value is filed under (index) and looked up by (lookup). Two values Compare calls Same or Equivalent meet: an index key of either is a lookup key of the other. Values that share a word that is not an initial meet too, and so do two given-like words where one starts the other (from Names.PrefixMinLength letters). Initials and generation words are never keys alone." | The map's scope entry; one function keeps the two lists made from one read of the words. `Keys` already returns two lists | Two functions `IndexKeys` / `LookupKeys` — each would re-read the words and duplicate `keyList`; keeping one list and adding prefix keys to both sides — "Des" would be filed under `p:de` and a lookup of "Desiree" under `p:de` would meet every name starting "De" |
| D-2 | **Two new tags.** For each word of a role in `AliasRoles()` that passes D-4, with `f` its fold (`nameWord.fold`) and `n` its rune count: index gets `g:`+`f` and `p:`+the first `k` runes of `f` for each `k` from `PrefixMinLength` to `n − 1`; lookup gets `p:`+`f` and `g:`+the same prefixes. So a whole word meets the other side's strict prefixes, and its strict prefixes meet the other side's whole words. Equal words keep meeting through `w:`. "Desiree": index `g:desiree p:de p:des p:desi p:desir p:desire`, lookup `p:desiree g:de g:des g:desi g:desir g:desire`. A helper `prefixKeys(words []nameWord, r Rules) (index, lookup []string)` *(new)* in `core/valuecompare/shortlist.go` makes them; `name.wordKeys` appends them and sorts and compacts each list. Every tag is a named constant in one block at the top of `core/valuecompare/shortlist.go` *(new)*: `keyWhole = "="`, `keyWord = "w:"`, `keyAlias = "a:"`, `keyGiven = "g:"`, `keyPrefix = "p:"`, each with a one-line comment saying which side files it and which side asks for it (`keyGiven`: "a given-like whole word: filed by the index, asked by a lookup's prefixes"; `keyPrefix`: "a strict prefix: filed by the index, asked by a lookup's whole word"; the other three: "both sides"). `keyList` and `name.wordKeys` use the constants in place of their literals | The design in [prefix rule § Shortlist](../name-candidates.md#prefix-rule-in-detail): separate tags so a filed prefix never meets a filed prefix. `g:` and not `w:` for the whole word, so a given word's prefix never meets a surname filed under `w:` (AC-1.4) | Looking up prefixes against `w:` — "Desiree" would reach a record surnamed "Des"; prefixes as long as the word — duplicates `w:` |
| D-3 | **`func AliasRoles() []NameRole`** *(new)* in `core/valuecompare/aliases.go`, after `AliasTable`: returns a fresh `[]NameRole{RoleGiven, RoleNick, RoleUntyped}`. Doc: "AliasRoles are the part roles the product alias tables and the prefix keys read: given, nickname and untyped words, never a surname." `namealiases.aliasPartRoles` is deleted; its four call sites (`registry.go` twice, `registry_test.go`, `namealiases_test.go`) call `valuecompare.AliasRoles()` | One list for the working rule's "the alias tables' roles", in the pure package S11-45's comparer also reads. A fresh slice, like `matchregistry.Default()`, because `AliasTable.Roles` holds it | A second literal list in `valuecompare` — two copies of one rule; reading roles off `Rules.Names.Aliases` — catalog tables carry their own roles, and with no tables (tests, `Default()`) there would be no prefix keys |
| D-4 | **Which words key prefixes: one predicate.** `func prefixWord(w nameWord, r Rules) bool` *(new)* in `core/valuecompare/words.go`, beside `IsInitial`, is true when `r.Names.PrefixMinLength >= 1`, `w.role` is in `AliasRoles()`, `IsInitial(w.text)` is false, and the rune count of `w.fold` is at least `r.Names.PrefixMinLength` and at most `r.Words.MaxSpellingLength`. Doc: "prefixWord says whether a name word takes part in the prefix rule: the shortlist's prefix keys and, from S11-45, the comparer's prefix hit read it, so the two never disagree." `func prefixesOf(fold string, r Rules) []string` *(new)* beside it returns the strict prefixes from `PrefixMinLength` to one rune short of the word, cut on runes (`[]rune`), never bytes. `prefixKeys` calls both and nothing else decides. A word of exactly `PrefixMinLength` runes ("De" at 2) files `g:de` and looks up `p:de`, with no prefixes of its own. `PrefixMinLength` 0 (or less) turns prefix keys off: index and lookup are then today's one list | Bounds the keys one hostile value makes (a 4,096-letter given name would make 4,094 keys), as `db987d96` bounded dates. `MaxSpellingLength` already says "a word this long is a run-on, not a spelling variant", the same judgment; no new knob. Off at 0 is the convention of `YearBucket` 0 and gives AC-1.6 a switch | A new `PrefixMaxLength` knob — a second number for the same run-on judgment; no bound — unbounded keys per value |
| D-5 | **The knob is `NameRules.PrefixMinLength int`** *(new)* in `core/valuecompare/rules.go`, after `Generation`. Doc: "PrefixMinLength is the fewest letters a given, nickname or untyped word's prefix has to be a shortlist key (Des for Desiree). 0 turns prefix keys off." In `defaultValue`, after `Generation: 0.5,`: the comment line `// decided: user, 2026-10-24: "Al", "Ed" and "Jo" count (S11-46 measured the wider shortlist on the large fixture: +13% hits, no cap)` then `PrefixMinLength: 2,`. Inventory row after `value.Names.Generation`: `` | `value.Names.PrefixMinLength` | 2 | decided: user, 2026-10-24 | Fewest letters of a given-name prefix the shortlist files and looks up; 0 turns prefix keys off | `` | A name rule, not a word rule: text never reads it (working rule), and S11-45's name comparer passes it into its word options. `WordRules` changes bump `autoreconciler.CacheVersion` and `searchindex.ProjectionVersion` (`add-match-rule`); this knob changes neither the reconciler nor search, and `NameRules` keeps that rule from firing for nothing | `WordRules.PrefixMinLength` — reads as a text rule too and trips the bump rule |
| D-6 | **No cap.** Measured at the base on a throwaway copy of this design (2026-10-28), `TestLargeSourceBudgets` run alone: index keys summed over the run's index builds 1,065,454 → 1,554,488 (+46%), shortlist hits over 820 lookups 477,412 → 538,578 (+13%), handle rows 62 → 70 and strong rows 46 → 51 (the fixture's `Givb` / `Givba` names are one-letter spelling variants, now shortlisted). Timings: warm at open 1.27 → 1.14 – 1.42 s, cold 1.23 → 1.27 – 1.41 s, warm 255 → 274 – 281 ms, after a handle commit 1.26 → 1.39 – 1.44 s; all inside budget. Every other test passed. STEP-4 repeats the measurement with timings only | The map asked for a cap only if the measurement needs one; it does not. A cap is a product trade (which "Jo" lookups to drop) better made from dogfood | A per-key posting cap or a hit cap — a new knob and a silent miss with no measured need |
| D-7 | **Tests.** `core/valuecompare/shortlist_test.go`: `TestWordKeys` and `TestWordKeysShareOnSameOrEquivalent` are edited for two lists (T-1, T-2); `TestWordKeysReachPrefixes` *(new)* (T-3). `core/shortlist/shortlist_test.go`: four `TestLookup` rows (T-4) and a `given(id byte, form string) entry` helper *(new)* beside `person`, building `subjectcompare.Values{vocabkey.PropertyName: {pt.Given(form)}}`. No `promotealign` test | The map's Check names a shortlist property test. The catalog path is unchanged code; `TestLargeSourceBudgets` covers it at scale | A catalog-level Des / Desiree test — the gate drops the pair until S11-45, so it would assert a non-result |
| D-8 | **Docs.** `docs/deployment-plan/spike-11/name-candidates.md` § Prefix rule: the **Roles** bullet names `valuecompare.AliasRoles` in place of `namealiases.aliasPartRoles`; the **Shortlist** bullet's last two sentences become "S11-46 split `valuecompare.WordKeys` into index and lookup lists and keys a word of the alias roles under `g:` (whole) and `p:` (prefix) tags, up to `MaxSpellingLength` runes; `prefixWord` is the one word test. On S11-21's large fixture it added 46% index keys and 13% shortlist hits, inside every budget, so there is no cap." (the numbers are D-6's planning measurement, copied, not re-measured) `docs/deployment-plan/spike-11/comparison-layers.md` layer 6 step 1: after "Guarantee: shortlisting never misses a pair value compare would call `Same` or `Equivalent` (aliases share a key)." add "A value has index keys and lookup keys; a given name that starts another meets it through prefix keys (S11-46)." `docs/matching-model.md` glossary row **shortlist key**: "made by layer 1" → "made by layer 1 as index and lookup lists". `.agents/skills/add-match-rule/SKILL.md`: the "Shortlist keys come from layer 1" bullet gains "`WordKeys` returns the keys a value is filed under and the keys it is looked up by; a pair meets when one's index shares a key with the other's lookup. Prefix keys (`g:`, `p:`) are names' only asymmetric keys."; "Add a comparer" step 6 becomes "6. A type that shortlists by words implements `wordKeys`, returning its index and lookup lists; add forms to `TestWordKeysShareOnSameOrEquivalent`." and "Add a shortlist key" step 4 becomes "4. A value type the lead uses implements `wordKeys` (index and lookup lists) or has a `YearReach`; extend `TestWordKeysShareOnSameOrEquivalent`'s or `TestYearReachCoversEveryRelatedPair`'s corpus." | Working rule: each PR updates the docs and skill it makes false | — |
| D-9 | **Text never keys prefixes**, so toponyms and every text lead are unchanged | Working rule "text compare never gets it" | — |
| D-10 | **No `CacheVersion` or `ProjectionVersion` bump** | Neither the reconciler nor the search index reads `WordKeys`; only `core/shortlist` does (`rg -n "WordKeys" core api` at the base) | — |
| D-11 | **Order.** STEP-1 moves the role list (no behavior change). STEP-2 splits the signature with identical lists (no behavior change; the suite proves it). STEP-3 adds the prefix keys, the knob and their tests | Each step leaves the tree green and isolates the one behavior change | One commit for all — a failure would not say which part broke |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: `rg -n "WordKeys|wordKeys" core api`, `rg -n "aliasPartRoles|RoleGiven, .*RoleNick" core`, `rg -n "prefix" core/valuecompare core/shortlist`, `rg -n "MaxSpellingLength|RuneCountInString" core/valuecompare`, `rg -n "func given|func person" core/shortlist core/valuecompare`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuecompare.AliasRoles` *(new)* | func | `namealiases.aliasPartRoles` | Moves it (deleted there) so a pure package can read it (D-3) |
| `valuecompare.prefixKeys` *(new)* | func, `core/valuecompare/shortlist.go` | `keyList` | `keyList` makes one symmetric list for names and text; prefix keys are names-only and two-sided. Kept beside it in the same file |
| `valuecompare.prefixWord`, `valuecompare.prefixesOf` *(new)* | func, `core/valuecompare/words.go` | `IsInitial` | The one statement of the prefix rule's word test, for the shortlist now and the comparer in S11-45 (D-4) |
| `keyWhole`, `keyWord`, `keyAlias`, `keyGiven`, `keyPrefix` *(new)* | const, `core/valuecompare/shortlist.go` | literals in `keyList`; `datePrefix` in `core/shortlist/shortlist.go` | Names each tag and its pairing once (D-2) |
| `slGiven`, `slNick` *(new)* | test helper, `core/valuecompare/shortlist_test.go` | `slName` | `slName` reads Western |
| `NameRules.PrefixMinLength` *(new)* | registry field | `WordRules.InitialCredit`, `WordRules.ShortVariantMinLength` | A new tunable; placed in `NameRules` (D-5) |
| `g:` and `p:` key tags *(new)* | shortlist key | `=`, `w:`, `a:`; dated `wide` / `any` | The existing word tags are symmetric; D-2 needs a side per tag, as the dated keys have |
| `TestWordKeysReachPrefixes` *(new)* | test | `TestWordKeys` | `TestWordKeys` checks one value's keys; T-3 checks pairs meeting across sides |
| `given` *(new)* | test helper, `core/shortlist/shortlist_test.go` | `person` | `person` reads Western, so a one-word form is a surname |

**Builds on.** `add-match-rule` "Add a knob"; the dated keys' index / lookup split in `shortlist.Keys` (`e8046e37`); `nameWords`' roles and folds; `IsInitial`; `MaxSpellingLength`; `namevaluetest.Given` / `propertycomparetest.Given` (S11-43).
**Deletes.** `namealiases.aliasPartRoles` (D-3). The "the same on both sides" sentence of `shortlist.Keys`' doc comment.
**Next adds.** S11-45's comparer decides a prefix hit through `prefixWord` (both words) and the shorter fold being the start of the longer, and adds only `PrefixCredit` and `CausePrefix`; it restates no role, minimum or length bound. S11-45 also widens T-2: every corpus pair `Compare` gives `CausePrefix` meets across index and lookup. Another asymmetric word key (a coarse spelling key for toponyms, the dogfood note in `TestSpellingOnlyToponymIsNotShortlisted`) is one more tag in `WordKeys`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Subject shortlists records it shares only a given-name prefix with | `core/graphalign/candidates.go` `Candidates` | The gate scores each shortlisted handle with `subjectcompare.Compare` and drops it under `MinSimilarity`: "Des" / "Desiree" is dropped until S11-45. A one-letter spelling variant ("Ann" / "Anne") that clears the gate becomes a candidate and can take a row, as any shortlisted Near pair can. `TopK` and `TieCap` bound the candidates per Subject. No new state reaches Swift |

### Scale

At S11-21's large fixture (4,000 persons, a 12-generation tree, 8,000 events): D-6's numbers. Per given-like word of `n` folded runes, `n − PrefixMinLength + 1` extra index keys and as many lookup keys ("Desiree": 6 and 6). The index is built once per catalog graph (at open, `Warm`) and after each handle commit, as today; a lookup costs one map read per key.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `WordKeys` and `wordKeyer` return `(index, lookup []string)` (D-1); key-tag constants and `prefixKeys` *(new)* (D-2); `prefixWord`, `prefixesOf` *(new)* (D-4); `AliasRoles` *(new)* (D-3); `NameRules.PrefixMinLength` *(new)* (D-5) |
| `core/shortlist` | `Keys` files `index` and looks up `lookup` from `WordKeys`; doc comment: "Word leads file every index word key of every positive value and look up every lookup word key (valuecompare.WordKeys)." |
| `core/matchregistry` | `defaultValue`: `PrefixMinLength: 2` (D-5) |
| `core/database/namealiases` | `aliasPartRoles` deleted; callers use `valuecompare.AliasRoles()` (D-3) |

Errors: none.

### FFI

None — the wire is unchanged.

### macOS

None — no chrome.

### Strings

None — no copy.

## 7. Steps

Every Go test command is run from the repo root with `CGO_ENABLED=1` and `-tags fts5`, as `scripts/check-pr.sh` runs them. `scripts/check-pr.sh` runs on macOS: off macOS it skips the Xcode checks and prints `check-pr: FAIL` for that reason alone, so a Linux run reports "every other check PASS, Xcode skipped" and the PASS line comes from a macOS run before the PR is marked ready.

**STEP-1 — One list of alias roles**
Files: `core/valuecompare/aliases.go`, `core/database/namealiases/registry.go`, `core/database/namealiases/registry_test.go`, `core/database/namealiases/namealiases_test.go`. Skill: `add-match-rule`.
Do: D-3.
Check: `rg -n "aliasPartRoles" core` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/... ./core/database/namealiases/...` passes.

**STEP-2 — Index and lookup word keys**
Files: `core/valuecompare/shortlist.go`, `core/shortlist/shortlist.go`, `core/valuecompare/shortlist_test.go`. Skill: `add-match-rule`, rule `go-tests.mdc`.
Do: D-1 with no prefix keys yet, and the key-tag constants of D-2 for the three existing tags. Rewrite `shortlist.Keys`' doc comment as §6 Go core says. Edit T-1 and T-2 as §8 says.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/... ./core/shortlist/... ./core/graphalign/... ./core/matchregistry/... ./core/database/promotealign/...` passes.

**STEP-3 — Prefix keys and the knob**
Files: `core/valuecompare/rules.go`, `core/valuecompare/shortlist.go`, `core/matchregistry/value.go`, `docs/matching-model.md`, `core/valuecompare/shortlist_test.go`, `core/shortlist/shortlist_test.go`. Skill: `add-match-rule` ("Add a knob").
Do: D-5 (field, registry value, inventory row), then D-4's `prefixWord` and `prefixesOf`, then D-2's two new constants and `prefixKeys`, appended in `name.wordKeys`. Add T-3 and T-4.
Check: the `add-match-rule` Check command passes: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/segmentcompare/... ./core/kinship/... ./core/shortlist/... ./core/graphalign/... ./core/claimevidence/... ./core/matchregistry/...`, with `TestKnobInventoryMatchesRegistry` and `TestLayerPackages` green.

**STEP-4 — Measure**
Files: none committed.
Do: run `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v -count=1` three times. Then set `PrefixMinLength: 0` in `core/matchregistry/value.go` (local edit, not committed), run it three times again, and restore the value.
Check: every run passes. The PR description lists the `perf_test.go` timing lines of each run, prefixes on and off. `git diff -- core/matchregistry/value.go` against the STEP-3 commit is empty.

**STEP-5 — Docs that describe word keys**
Files: `docs/deployment-plan/spike-11/name-candidates.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `.agents/skills/add-match-rule/SKILL.md`.
Do: the edits of D-8, word for word.
Check: `rg -n "aliasPartRoles" docs/deployment-plan/spike-11/name-candidates.md docs/deployment-plan/spike-11/comparison-layers.md docs/matching-model.md .agents core` prints nothing; `rg -n "index and lookup" .agents/skills/add-match-rule/SKILL.md docs/matching-model.md docs/deployment-plan/spike-11/comparison-layers.md` prints a line in each file.

**STEP-6 — Whole suite**
Files: none.
Do: run `scripts/check-pr.sh` (on macOS for the PASS line; see the note above STEP-1). Then push the branch, made from `main` at STEP-1, and open the PR with §13's title and body.
Check: it prints `check-pr: PASS`. `git diff main --stat -- '*_test.go'` lists only `core/valuecompare/shortlist_test.go`, `core/shortlist/shortlist_test.go`, `core/database/namealiases/registry_test.go` and `core/database/namealiases/namealiases_test.go`.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done, once the PR number is known.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-46-prefix-shortlist-keys.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence ending in a period, like the rest of `git log`. STEP-1: "Keep the alias tables' part roles in one list in valuecompare." STEP-2: "Split word keys into the keys a value is filed under and the keys it is looked up by." STEP-3, committed before STEP-4: "Shortlist a given name that starts another, both ways, so Des reaches Desiree." STEP-5: "Say in the matching docs and skill that a value has index and lookup keys." STEP-7 is a follow-up docs commit on the open PR. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeys` *(edited)* | Each existing row runs on both lists: `index, lookup := valuecompare.WordKeys(tt.v, rules)`. A local `plain(keys)` drops keys starting `p:` or `g:`; `want` (exact), `has` and `not` apply to `plain(index)` and `plain(lookup)`, and `plain(index)` equals `plain(lookup)`. Both raw lists are sorted and unique. The nil row expects `nil, nil`. The text alias check reads the index list. The existing row "initial is no key" (`slName("J. Robins")`) also asserts that neither raw list holds a `p:` or `g:` key. New row `{name: "lone initial keys only its whole key", v: slGiven("J."), want: []string{whole(slGiven("J."))}}`: index equals lookup, exactly that one `=` key, no `p:` or `g:` key | AC-1.5, AC-1.8, D-1 |
| T-2 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeysShareOnSameOrEquivalent` *(edited)* | For each `Same` or `Equivalent` pair: `ia, la := WordKeys(a)`, `ib, lb := WordKeys(b)`; fail unless `shares(ia, lb) && shares(ib, la)`, printing all four lists. Corpus unchanged | AC-1.9, D-1 |
| T-3 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeysReachPrefixes` *(new)* | Table of `{name, filed, probe valuecompare.Value; min int; meets bool}`; `meets` is `shares(index(filed), lookup(probe))` under `matchregistry.Default().Value` with `Names.PrefixMinLength = min`. Given-only values via `slGiven(form)` *(new, beside `slName`)*: `valuecompare.Value{Type: propertymeta.ValueTypeName, Name: namevaluetest.Given(form), NamePattern: valuecompare.WesternPattern}`; a nick value via `slNick(form)` *(new)*: the same with `Name: &namevalue.Value{Form: form, Parts: []namevalue.Part{{Idx: 0, Value: form, Type: namevalue.PartTypeNick}}}`. Rows (min 2 unless named): Desiree / Des → true; Des / Desiree → true; Desiree / Desmond → false; Albert / Al → true; De / Desiree → true; Desiree / De with min 3 → false; Desiree / Dés (accent) → true; filed Western `Mary MacDonald` / probe Western `John Mac` → false; filed Western `Mary MacDonald` / probe given `Mac` → false; filed `slNick("Des")` / probe given `Desiree` → true; Desiree / Des with min 0 → false. Then two checks: given `Desiree` at min 0 has `index` equal to `lookup` and no `p:` / `g:` key; a given word of 33 `x` runes at min 2 has no `p:` / `g:` key in either list | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, D-2, D-4 |
| T-4 | Go | `core/shortlist/shortlist_test.go` `TestLookup` *(four new rows)* | `{"prefix reaches the longer name", "S11-46 AC-1.1", plain, Person, [given(1,"Desiree"), given(2,"Desmond"), given(3,"Mary")], given(0,"Des"), []byte{1, 2}}`; `{"prefix reaches the shorter name", "S11-46 AC-1.2", plain, Person, [given(1,"Des"), given(2,"Desmond"), given(3,"De")], given(0,"Desiree"), []byte{1, 3}}`; `{"no shared prefix", "S11-46 AC-1.3", plain, Person, [given(1,"Desiree")], given(0,"Desmond"), nil}`; `{"surname prefix", "S11-46 AC-1.4", plain, Person, [person(1,"John MacDonald")], person(0,"Mary Mac"), nil}`. Existing rows unchanged; the `S11-46` label keeps these apart from the S11-15 ACs the existing rows cite | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.10 |

Must stay green: everything `scripts/check-pr.sh` runs (AC-1.10). Most likely to break: `core/database/promotealign` (`TestLargeSourceBudgets` warm and after-commit timings, goldens), `core/matchregistry` (`TestKnobInventoryMatchesRegistry`, `TestLayerPackages`'s numbers guard on `prefixKeys`), `core/graphalign`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: AC-1.1 – AC-1.4 → T-3, T-4; AC-1.5 → T-1, T-3; AC-1.6, AC-1.7 → T-3; AC-1.8 → T-1; AC-1.9 → T-2; AC-1.10 → STEP-4, STEP-6
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS on macOS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — n/a, no copy
- [ ] Sweep: `rg -n "aliasPartRoles"` across `.agents/`, `.cursor/rules/`, `docs/*.md`, `docs/deployment-plan/spike-11/*.md` and `core` has no hit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — n/a, none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no chrome
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status Landed #N; completed.md: one short paragraph of what shipped, newest first
- [ ] Brief moved to design/archive/ and design/README.md updated — n/a, no brief
- [ ] Domain docs updated: docs/matching-model.md (inventory row, glossary), comparison-layers.md, name-candidates.md, add-match-rule skill (D-8)
- [ ] PR description: what shipped, AC evidence, STEP-4's timings with prefixes on and off, deviations, manual check
```

## 10. Manual check

1. None in the app: until S11-45, "Des" against "Desiree" is `Different` and the gate drops the shortlisted pair, so no row changes for the slice 10 case. T-3 and T-4 stand in for it, as the map's Check says ("a shortlist property test").
2. STEP-4's timings show Promote's budgets hold with prefix keys on.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The wider shortlist slows a warm proposal past 350 ms | Measured at +20 – 25 ms run alone (D-6); STEP-4 repeats it. The base already reaches 340 ms under the full parallel suite here, so a single slow run under load is noise; three failing isolated runs are a stop |
| A short given name ("Jo", "Al") shortlists many records | `TopK` and `TieCap` bound the candidates; the gate drops them until S11-45. A cap is a dogfood call (Out) |
| A spelling variant now shortlisted takes a row it did not take before | Intended: value compare already relates the pair; the suite and goldens passed with the change at planning |
| `prefixKeys` trips the numbers guard | It uses only 0 and 1 and the registry's fields |

**Stop and report if:**

- `TestLargeSourceBudgets` fails in all three isolated STEP-4 runs with prefixes on, or passes with prefixes off and fails with them on in two of three: report the timing lines; do not raise a budget or add a cap.
- Any test or golden other than T-1 and T-2 fails after STEP-2 or STEP-3: name it and its output; do not edit it.
- `rg -n "WordKeys\(" core api` finds a call other than `core/shortlist/shortlist.go` and `core/valuecompare` (D-10 would be wrong).
- `nameWords` no longer gives each word a `role` and `fold`, or `aliasPartRoles` returns a list other than given, nick, untyped.
- `TestWordKeysShareOnSameOrEquivalent` fails after STEP-2 on a pair: the split dropped a key from one side.

## 12. Open questions

None.

## 13. PR

**Title:** S11-46 — Prefix keys in the shortlist
**Body:** objective in one sentence; what shipped against US-1; AC evidence; STEP-4's timings with prefixes on and off; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-46-prefix-shortlist-keys.md`.
