# S11-45 — Prefix names

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the evidence sheet's existing name line gains one sentence; nothing new is drawn)
**Board:** —
**Depends on:** S11-46 (landed [#602](https://github.com/mendahu/provenencia/pull/602)), S11-28 (landed [#589](https://github.com/mendahu/provenencia/pull/589))
**Blocks:** —
**Branch:** `s11-45-prefix-names`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [name candidates § Prefix rule, in detail](../name-candidates.md#prefix-rule-in-detail), [§ Decided](../name-candidates.md#decided-user-2026-10-24), [comparison layers § Relations](../comparison-layers.md#relations), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md), [`matching-model.md` § Knob inventory](../../../matching-model.md#knob-inventory); working rules: "A prefix is an initial generalised", "The reconciler keeps its own decisions", "One wire vocabulary", "Registry numbers: carry over, then estimate", "High tolerance for churn"
**Planned:** 2026-10-28 against `main` @ `784c7ae`
**Review:** [`S11-45-prefix-names.review.md`](S11-45-prefix-names.review.md) — round 1 Fits with changes / Ready with changes; round 2 pending

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

A given name that is the exact start of another given name now counts as a weaker match, the way an initial does: "Des" against "Desiree" is `Contains` at `PrefixCredit` (0.7) with the new cause `prefix`, so Promote offers one for the other even with no nickname set, the evidence sheet says "Matches by a shortened name", and the reconciler shows "Des Long" and "Desiree Long" as one name, "Desiree Long". Surnames never match by prefix, and a nickname table hit, an accent fold and a spelling variant still win over a prefix.

**Why now.** Dogfood, 2026-10-24 ([case 2](../name-candidates.md#case-2--desiree-against-des)). S11-46 made the shortlist offer these pairs both ways; until this PR the comparer calls them `Different` and the candidate gate drops them.

**What stays.** The shortlist keys (S11-46). Every relation the word table gives today: exact, initial, alias, abbreviation, fold and spelling are each checked before the prefix rule, so "Ann" / "Anne" stays a spelling `Near` and "Jim" / "James" with a nickname table stays an alias. Text compare. Bands, the candidate gate and every other knob. A prefix pair earns what its similarity earns, like an initial: no new band rule.

## 2. User stories

**US-1.** As a researcher, when a Source person's given name is the start of a record's given name, or the reverse, and no nickname set joins them, Promote offers the record, so that a shortened name reaches the person it stands for.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Default rules, no alias tables. Given-only `Desiree` against given-only `Des`; then the sides swapped | `Contains`, similarity 0.7, cause `prefix`, one part pair `desiree` / `des` at 0.7 `prefix`; `Broader` is `b`, and `a` when swapped |
| AC-1.2 | `Des Long` (given, surname) against `Desiree Long` | `Contains`, similarity (0.7·2·G1 + 2·F) / (2·G1 + 2·F) = 0.88, cause `prefix`, `Broader` `a` |
| AC-1.3 | Given `Desmond` against given `Desiree` | `Different`, cause `no_match` |
| AC-1.4 | Surname `Mac` against surname `MacDonald`; `Mary Mac` against `Mary MacDonald` | `Different`, cause `no_match`, for both |
| AC-1.5 | The rules ahead of the prefix: `Ann` / `Anne`; `Des` / `Desi`; `D.` / `Desiree`; `Jim` / `James` with a nickname table holding both; `Dés` / `Desiree` | `Near` 0.75 `spelling`; `Near` 0.75 `spelling`; `Contains` 0.5 `initial`; `Equivalent` 0.98 `alias`; `Contains` 0.7 `prefix` (the fold is applied, then the prefix) |
| AC-1.6 | Given `De` against given `Desiree` at `PrefixMinLength` 2, then 3; given `Des` against `Desiree` at 0 | `Contains` 0.7 `prefix`; `Different`; `Different` (0 turns the rule off) |
| AC-1.7 | Nickname part `Des` against given `Desiree`; `James Des` (two given) against `J. Desiree` (two given) | `Different` `below_minimum` for both: a role mismatch is `Near` (`role`), and two words each broader on its own side are one `Near` overlap, as for initials today |
| AC-1.8 | Promote (graph alignment). Subject given-only `Desiree`; records given-only `Des` (PER-1) and `Desmond` (PER-2) | The row is unplaced with no band; its alternatives are exactly PER-1, with no band |
| AC-1.9 | Promote (graph alignment). Subject `Desiree Long`; record `Des Long` | Target the record, band weak, reason agrees, score 2.5 × 0.88 = 2.2 |
| AC-1.10 | Promote on a catalog. Promoted `Rosalind Long` and `Mary Long`; a new Source's person `Ros Long`. Then the reverse: promoted `Ros Long` and `Mary Long`; a new Source's `Rosalind Long` | Each row targets the `Ros…` / `Rosalind…` record at the weak band (no nickname set holds `ros`: the shortlist's prefix keys and the comparer's prefix rule carry it) |
| AC-1.11 | Ladder, probe `Desiree Long` | `Desiree Long` > `Désirée Long` > `Des Long` > `D. Long` in similarity; `Desmond Long` and `Desiree Short` are `Different` |

**US-2.** As a researcher, when an evidence line's names match by a shortened name, the sheet says so, so that I know why the match is weaker than a nickname.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `L10n.Promote.lineSentence(relation: .contains, cause: .prefix, …)` for a name, a date and a term value | `L10n.Promote.lineContainsPrefix`: "Matches by a shortened name" |
| AC-2.2 | The wire cause `prefix` | `ValueCause(wire: "prefix")` is `.prefix`; `TestSwiftVocabularyMatchesGo` lists `valuecompare.CausePrefix` |

**US-3.** As a researcher, when a person's Sources record "Des Long" and "Desiree Long", the Persons list and Person page show one given name, "Desiree", so that a shortened name reads as the name it stands for.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Reconcile `given=Des|surname=Long` (1) and `given=Desiree|surname=Long` (2); then in the other order | One value, members 1 and 2, `given=Desiree|surname=Long`, state merged, reason kept, both orders |
| AC-3.2 | Reconcile `given=Albert|surname=Long` (1), `given=Alice|surname=Long` (2), `given=Al|surname=Long` (3) | One value, members 1, 2, 3, `given=Albert|given=Alice|surname=Long`: Al folds into the best ranked of equal support, as an ambiguous initial does |
| AC-3.3 | Reconcile `given=Mary|surname=Mac` with `given=Mary|surname=MacDonald`; `nick=Des|surname=Long` with `given=Desiree|surname=Long`; `given=Desmond|surname=Long` with `given=Desiree|surname=Long` | Each result equals the result under the same rules with `PrefixMinLength` 0: a surname never prefix-folds, a fold never crosses part types, and a word that is not a prefix does not fold |
| AC-3.4 | `autoreconciler.CacheVersion` | 20, with its changelog line; `productAliasPin.CacheVersion` 20 and its hash unchanged |

**US-4.** As a contributor, the comparer's contract and the shortlist's guarantee cover prefix pairs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `TestComparerContract` and `TestNameInvariants` with prefix samples and pool words | Pass: a prefix pair is symmetric (mirrored detail), and `Contains` scores below 1 |
| AC-4.2 | `TestWordKeysMeetOnPrefixPairs` corpus with `Des Long` / `Desiree Long` | Every pair `Compare` gives cause `prefix` meets across index and lookup, both directions, and the corpus meets at least one such pair |

## 3. Scope

| In | Covered by |
| --- | --- |
| `NameRules.PrefixCredit` 0.7 in the registry with its inventory row (D-1) | STEP-1, AC-1.1 |
| `valuecompare.CausePrefix` (D-2) | STEP-1, AC-1.1 |
| The prefix rule in the word table after spelling, behind word options names set and text never does (D-3, D-4) | STEP-2, STEP-3, AC-1.1 – AC-1.7 |
| The broader side of a `Contains` word pair from the initial or the shorter fold (D-5) | STEP-3, AC-1.1, AC-1.2 |
| Name tests, ladder, contract samples and the shortlist share corpus (D-11) | STEP-3, AC-1.11, AC-4.1, AC-4.2 |
| Promote tests: graph alignment and catalog (D-11) | STEP-4, AC-1.8 – AC-1.10 |
| `CausePrefix` joins the reconciler's fold allow-list for alias-role part types; `CacheVersion` 20 (D-6, D-7) | STEP-5, AC-3.1 – AC-3.4 |
| `ValueCause.prefix`, its sentence and its L10n key (D-9) | STEP-6, AC-2.1, AC-2.2 |
| Docs, the matching skill and code comments the rule makes false (D-12) | STEP-7 |

| Out | Home |
| --- | --- |
| Surname prefixes ("Mac" / "MacDonald") | Not planned ([what stays later](../name-candidates.md#what-stays-later)) |
| A sentence that names the two words ("Des may be short for Desiree") | Not planned: D-9 keeps the sheet's one-sentence-per-cause table |
| Tuning `PrefixCredit` or any band so a prefix places differently | Dogfood log: retuned from dogfood, as the working rule "Registry numbers: carry over, then estimate" says |
| Relaxing the name curve | Held (working rule "Gate only; the curve is held") |
| A spelling-variant prefix ("Desy" for "Desiree") | Not planned: the rule is an exact start after folding (working rule) |
| A `searchindex.ProjectionVersion` bump | Not planned: search never reads name compare (D-8) |
| Reconciler fold causes on the Person page's provenance sheet (Spike 13 R-9) | Spike 13 ([`person-detail.md`](../../spike-13/person-detail.md)): it words the reconciler's records, which this PR does not change |

## 4. Current state

Verified at `784c7ae`.

| Area | Where | Today |
| --- | --- | --- |
| Word table | `core/valuecompare/words.go` `wordSimilarity`, `wordOptions` | Order: byte-equal → `Same`; an initial → `Contains` `initial` (with `initials`) or no pair; alias set → `Equivalent` `alias` / `abbreviation`; folds equal → `Equivalent` `fold`; a word over `MaxSpellingLength` runes → no pair; spelling → `Near` `spelling`; else `Different` `no_match`. `wordOptions` holds `initials`, `aliases`, `roleX`, `roleY`. No prefix rule |
| Prefix word test | `core/valuecompare/words.go` `prefixWord`, `prefixesOf` | `prefixWord(w nameWord, r Rules)` is true for a word of a role in `AliasRoles()`, not `IsInitial`, of `PrefixMinLength` to `MaxSpellingLength` folded runes, and false when `PrefixMinLength < 1`. Read by the shortlist only (`shortlist.go` `prefixKeys`); its doc says "from S11-45, the comparer's prefix hit read it" |
| Name comparer | `core/valuecompare/name.go` `name.word`, `name.pair` | `word` passes `wordOptions{initials: true, aliases: r.Names.Aliases, roleX: x.role, roleY: y.role}`. `pair`: a role mismatch is `Near` (`role`, or `spelling`); a `Contains` pair sets `broader` to `SideB`, `SideA` when `IsInitial(x.text)` |
| Text comparer | `core/valuecompare/words.go` `alignWords` | Passes `wordOptions{}`: no initials, aliases or prefixes. Unchanged |
| Causes | `core/valuecompare/valuecompare.go` `Cause` constants | `CauseInitial = "initial"`; no prefix cause |
| Name rules | `core/valuecompare/rules.go` `NameRules`; `core/matchregistry/value.go` `defaultValue` | `PrefixMinLength int` (2), doc "the fewest letters a given, nickname or untyped word's prefix has to be a shortlist key". `WordRules.InitialCredit` 0.5, `AliasCredit` 0.98. No `PrefixCredit` |
| Inventory | `docs/matching-model.md` knob inventory; `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | `value.Names.PrefixMinLength` row (line 141) is the last `value.Names.*` field row before the Western weights |
| Reconciler fold | `core/autoreconcile/names.go` `nameModule.fold`, `nameUnit.compareValue` | `compareValue` strips part types, so every word compares untyped. `fold` takes `Contains` with `Broader == SideA` or `Equivalent`; each part pair must be `Same`, `CauseInitial`, or an alias / abbreviation that `Aliases.Expands` toward the into word under the type's role (`m.rules.Names.Western.Role(typ)`); else no fold. A prefix pair today is no pair at all |
| Reconciler cache | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` = 19; `core/database/autoreconciler/autoreconciler_test.go` `productAliasPin` `{CacheVersion: 19, Hash: "0dee1d25…"}` | The cache reads `valuerules.Product()` (`matchregistry.Default().Value` + the product alias seed), so a rule change that moves the table is a bump |
| Wire | `api/proto/engine.proto` claim evidence line `string cause = 5; // valuecompare.Cause` | A string: no proto change for a new code |
| Swift causes | `macos/App/Platform/ConclusionVocabulary.swift` `ValueCause`, `init(wire:)`; `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "value causes" | One case per Go cause plus `none` and `unrecognized`. The Go row lists each cause by hand; no `prefix` |
| Sheet sentence | `macos/App/Platform/L10n.swift` `L10n.Promote.lineSentence`, `lineContainsInitial`; `macos/App/Resources/Localizable.xcstrings` `promote.sheet.line.containsInitial` | `.contains` switches on cause: `precision`, `missingPart`, `initial`, else `lineContainsOther` ("Matches, with less detail"). Each sentence is a fixed string; the line shows both values beside it |
| Sheet badge and check | `core/propertycompare/propertycompare.go` `outcomeOf`; `core/matchregistry/evidence.go` `Checked` | `Contains` → `OutcomePartial` (badge "Similar"); only `Same` / `Equivalent` lines start checked. Unchanged |
| Tests | `core/valuecompare/words_test.go` `TestWordSimilarityNames`; `core/valuecompare/export_test.go` `WordOptions`; `core/valuecompare/name_test.go` `nameCases`, `TestNameCompare`, `TestNameLadders`, `withTables`, `nicknameTable`; `core/valuecompare/contract_test.go` `samples`, `namePools`; `core/valuecompare/shortlist_test.go` `TestWordKeysShareOnSameOrEquivalent`; `core/graphalign/align_test.go` `TestOneGivenNameIsOfferedAFullerRecord`, `givenProp`, `nameProp`; `core/database/promotealign/promotealign_test.go` `TestProposeExtraGivenNamesAreAWeakMatch`, `newFixture`, `fixture.person`; `core/autoreconcile/names_test.go` `TestReconcileNames`, `TestReconcileNamesAliases`, `checkNames`, `testRules`; `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` `lineSentencesFollowTheTable` | No prefix case anywhere |

**Precedent.** `c798a2fe` ("Add initials and alias tables to value compare's word similarity") added a name-only rung to `wordSimilarity` through `wordOptions` and `name.word`, with rows in `TestWordSimilarityNames`: copy that shape for the prefix rung. `ef6f29a4` ("Let an alias fold toward the name it stands for in the reconciler's name rule", S11-28) added a cause to `nameModule.fold`'s allow-list under the type's role, with `TestReconcileNamesAliases`' fold and unchanged tables: copy its role check and its "unchanged" comparison. `ace01fd7` added a value cause (`unjudged`) to `ValueCause`, `init(wire:)` and the "value causes" row of `TestSwiftVocabularyMatchesGo`: copy those three edits. Do not copy `ace01fd7`'s proto change: this code needs none.

**Planning probe.** A throwaway copy of D-1 – D-6 at the base (2026-10-28) passed the whole Go suite (`./core/... ./api/...`) unchanged, with the widened test pools and corpus of D-11; it gave every Go result the ACs above cite, and showed that without D-6's role check the reconciler folds `surname=Mac` into `surname=MacDonald`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **`NameRules.PrefixCredit float64`** *(new)* in `core/valuecompare/rules.go`, right after `PrefixMinLength`. Doc: "PrefixCredit is what a given, nickname or untyped word earns when it is the start of the other's word (Des for Desiree), below an alias and above an initial." In `defaultValue`, after `PrefixMinLength: 2,`: the comment `// An estimate (name-candidates.md, user 2026-10-24): above an initial (InitialCredit 0.5), below a nickname table hit (AliasCredit 0.98).` then `PrefixCredit: 0.7,`. Inventory row after `value.Names.PrefixMinLength`: `` | `value.Names.PrefixCredit` | 0.7 | estimate | A given, nickname or untyped word that is the start of the other's word | ``. The `PrefixMinLength` field doc becomes "PrefixMinLength is the fewest letters a given, nickname or untyped word needs to take part in the prefix rule: a shortlist prefix key and a comparer prefix hit (Des for Desiree). 0 turns both off." and its inventory row's Changes cell becomes "Fewest letters of a given-name prefix the shortlist keys and the comparer counts; 0 turns the prefix rule off" | A name rule beside its sibling knob, as S11-46 D-5 placed `PrefixMinLength`: text never reads it, and a `WordRules` change would trip `add-match-rule`'s `ProjectionVersion` bump rule for nothing (D-8) | `WordRules.PrefixCredit` beside `InitialCredit` — reads as a text rule and trips the bump rule |
| D-2 | **`CausePrefix Cause = "prefix"`** *(new)* in `core/valuecompare/valuecompare.go`, right after `CauseInitial`, with the comment `// CausePrefix: a given, nickname or untyped word that is the start of the other's (Des / Desiree).` | One cause per rule, as `initial`; callers (sheet, reconciler) tell it from an initial | Reusing `CauseInitial` — the sheet would say "Matches by initial" for Des, and the reconciler could not guard roles for prefixes alone |
| D-3 | **The word table's prefix rung, and it names the broader side.** `wordOptions` gains one field after `roleX, roleY`: `prefixCredit float64 // what a prefix hit earns (Names.PrefixCredit); 0 when either word is not a prefix word, so there is no prefix rung`. `wordSimilarity` returns a fourth result, `broader Side`: the side whose word the other contains on a `Contains` hit, else `SideNone`. The initial branch returns `SideA` when `IsInitial(x)`, else `SideB`. After the spelling check and before the last `return`: `if o.prefixCredit > 0 { if strings.HasPrefix(fy, fx) { return o.prefixCredit, Contains, CausePrefix, SideA }; if strings.HasPrefix(fx, fy) { return o.prefixCredit, Contains, CausePrefix, SideB } }`. Every other return gives `SideNone`. In the doc comment, "a spelling variant is Near at its ratio (spelling); else no pair, 0." becomes "a spelling variant is Near at its ratio (spelling); a word that is the start of the other after folding, when the options carry a prefix credit, is a pair at that credit, Contains (prefix); else no pair, 0. broader is the side of the initial or the prefix on a Contains pair, else SideNone." The `wordOptions` doc becomes "wordOptions are what a caller adds to the word table: names read initials, alias tables and prefixes, text none." `alignWords` takes `sim, rel, why, _ :=` and passes no prefix credit. `export_test.go` `WordSimilarity` returns the four results | The working rule's order: after alias, fold and spelling. Equal folds never reach the rung (fold returns first), so a hit is always a strict prefix; `strings.HasPrefix` on two valid UTF-8 folds is a rune prefix. Words over `MaxSpellingLength` returned earlier, and `prefixWord` bounds the same length. One word table for every rung (`add-match-rule`: "`wordSimilarity` is the one word table"), and it decides the broader side where it makes the match (review M-2), so the next `Contains` rung is one branch. One credit field, not a flag and a credit that could disagree (review M-4 b) | A prefix check in `name.word` around `wordSimilarity` — a second word table; before spelling — "Des" / "Desi" and "Ann" / "Anne" would flip from spelling to prefix and the ladders move; `name.pair` guessing the broader side from word length — a proxy for what the table already knows |
| D-4 | **`name.word` turns it on.** `name.word` returns the four results of `wordSimilarity`, called with `wordOptions{initials: true, aliases: r.Names.Aliases, roleX: x.role, roleY: y.role, prefixCredit: credit}`, where `credit` is `r.Names.PrefixCredit` when `prefixWord(x, r) && prefixWord(y, r)`, else 0. Its doc: "word is the word table for two name words: initials count, the catalog's aliases apply, each under its own role, and a word that starts the other counts when both take part in the prefix rule." `name.compare` takes `sim, _, _, _ :=`. **One role test** *(review M-1)*: `func PrefixRole(role NameRole) bool` *(new)* in `core/valuecompare/words.go`, before `prefixWord`, returns `slices.Contains(AliasRoles(), role)`, doc "PrefixRole says whether a name part of this role takes part in the prefix rule: given, nickname and untyped words, never a surname. The shortlist, the comparer and the reconciler's fold read it." `prefixWord` calls `PrefixRole(w.role)` in place of its `slices.Contains`, and its doc becomes "prefixWord says whether a name word takes part in the prefix rule: the shortlist's prefix keys and the comparer's prefix hit read it, so the two never disagree." `AliasRoles`' doc becomes "AliasRoles are the part roles the product alias tables read: given, nickname and untyped words, never a surname. PrefixRole reads them too." | S11-46 D-4 made `prefixWord` the one word test (roles, initials, minimum, length bound); the comparer restates none of it. `PrefixMinLength` 0 turns off keys and hits together (AC-1.6). `PrefixRole` names the prefix half of `AliasRoles` once, so the reconciler (D-6) asks it instead of restating a role list | Checking roles in `wordSimilarity` — it does not see `nameWord`; the reconciler calling `AliasRoles` — two call sites for one rule, which part ways silently if alias and prefix roles ever differ |
| D-5 | **`name.pair` copies the broader side.** The `rel == Contains` arm sets `c.broader` to the `broader` that `n.word` returned, in place of today's `SideB` / `IsInitial(x.text)` rule. `name.go` gains no import | The word table decides it (D-3). Today's rule says `SideB` for "Des" on side A, so the reconciler would never fold Des into Desiree (it needs `Broader == SideA`) | Fold length in `pair` — a proxy with an edge case (a word whose fold drops a combining mark ties with an initial's length) |
| D-6 | **The reconciler folds a prefix, never on a surname.** In `nameModule.fold`, after the `CauseInitial` case: `case p.Cause == valuecompare.CausePrefix && valuecompare.PrefixRole(role): changed = true`. The `fold` doc comment becomes "fold: value compare finds into contains from, or the two equivalent, and every word they pair is the same word, an initial of it, a prefix of it (for a part type whose role takes prefixes), or a nickname or abbreviation the alias list pairs toward the into word. The type only says which part role an alias table or a prefix must allow (compareValue strips it)." | The user's call (2026-10-24): a prefix folds like an initial. `compareValue` compares every word untyped (S11-06 D-4), so `prefixWord` is true for surname words there: the planning probe folded `Mac` into `MacDonald` without the check. `role` is already read for the alias check; `PrefixRole` is the one statement of the prefix roles (D-4) | No role check — a surname prefix folds, against the working rule; typing the reconciler's compare values — reverses S11-06 D-4's weighting decision |
| D-7 | **`CacheVersion = 20`**, changelog line `//	20: a given name that is the start of another folds into it, like an initial (S11-45).` after line 19; `productAliasPin` becomes `{CacheVersion: 20, Hash: <unchanged>}` | The rule changes what the reconciler table holds (`add-match-rule`: "a rule or seed change is a `CacheVersion` bump"); `TestProductAliasTablesPinCacheVersion` fails until the pin moves with it | — |
| D-8 | **No `searchindex.ProjectionVersion` bump** | Search keys text through `valuecompare.TextKey` only (`core/database/searchindex/handles.go`), and `WordRules` is unchanged | — |
| D-9 | **The sheet's sentence.** `ValueCause` gains `case prefix` right after `case initial`, and `init(wire:)` gains `case "prefix": self = .prefix` after `case "initial"`. `L10n.Promote` gains, after `lineContainsInitial`: `static let lineContainsPrefix = LocalizedStringResource("promote.sheet.line.containsPrefix", defaultValue: "Matches by a shortened name", comment: "Evidence sentence: a name that matches by a shortened name (Des for Desiree)")`, with the same key, comment and English value in `Localizable.xcstrings` (`extractionState` `manual`, after `promote.sheet.line.containsInitial`). `lineSentence`'s `.contains` switch gains `case .prefix: return lineContainsPrefix` after `.initial`. `TestSwiftVocabularyMatchesGo`'s "value causes" row gains `string(valuecompare.CausePrefix)` after `CauseInitial` | The working rule "One wire vocabulary". One fixed sentence per cause, the sheet's table (`lineSentence` takes no values; the line already shows both names). "Shortened" covers Al for Albert and Jo for John | "Des may be short for Desiree" (the feature doc's example) — needs the two words in `lineSentence`, a new signature for one cause; "Matches, with less detail" (today's fallback) — says nothing about why |
| D-10 | **Bands are not touched.** A prefix pair earns its similarity like any partial name: `Desiree Long` / `Des Long` is 2.5 × 0.88 = 2.2, weak, and proposed (AC-1.9), as `J. Robins` / `James Robins` is today; a given-only prefix pair is 1.75, under weak, and offered as an alternative (AC-1.8) | Working rule "Bands come from evidence"; the credit is the user's estimate, retuned from dogfood (Out) | A prefix-specific cap or band rule — a second mechanism beside similarity |
| D-11 | **Tests** as §8 lists them: word table rows (T-1), name cases (T-2), a ladder (T-3), contract samples and pools (T-4), the shortlist share corpus (T-5), graph alignment (T-6), a catalog proposal (T-7), reconciler folds (T-8, T-9), the Swift vocabulary row (T-10) and the sheet sentence (T-11). Test helpers *(new)*: `valuecompare.WithPrefixes(o wordOptions, credit float64) wordOptions` in `core/valuecompare/export_test.go` (returns `o` with `prefixCredit: credit`); `sameResult(t *testing.T, label string, with, without valuecompare.Rules, in []Candidate)` in `core/autoreconcile/names_test.go`, the body of `TestReconcileNamesAliases`' unchanged loop moved out (both `Reconcile` calls, the `specs` / `State` / `reasonsOf` comparison, the failure message naming `label`), which that loop and T-9 call (review M-3); `withPrefixMin(n int) func(*valuecompare.Rules)` in `core/valuecompare/name_test.go` beside `withTables` (sets `r.Names.PrefixMinLength = n`); the constant `prefixCredit = 0.7 // des ~ desiree: Names.PrefixCredit` in `name_test.go`'s const block beside `robbins` | Each layer the rule reaches gets its own test; the catalog test proves the shortlist and the comparer compose (S11-46 left that to this PR) | A catalog test with Des / Desiree — the product nickname seed holds `{desiree, des}` since S11-44, so it would test the alias, not the prefix: `ros` is in no set |
| D-12 | **Docs and comments.** `docs/deployment-plan/spike-11/name-candidates.md` § Prefix rule: in **Words**, "(today `name.go` picks it with `IsInitial`, which is wrong for a prefix on side A)" → "(`name.go` takes the initial, or else the shorter word, S11-45)"; the whole **Evidence wording** bullet is replaced by "`CausePrefix` joins the cause codes the evidence sheet words, in `ConclusionVocabulary.swift` under `swift_vocabulary_test.go`: \"Matches by a shortened name\" (`promote.sheet.line.containsPrefix`), one fixed sentence like the initial's (S11-45)."; the whole **The reconciler** bullet is replaced by "Its name fold (`core/autoreconcile/names.go`) takes `CausePrefix` pairs like `CauseInitial` ones, for a part type of a given, nickname or untyped role only: it compares a type's words untyped, so the role comes from the type (S11-45; [decided](#decided-user-2026-10-24) item 2)." (the link to the open questions and the sentence about S11-28 go with the old bullet) `docs/deployment-plan/spike-11/comparison-layers.md`: the relations table's name row, Contains cell → "parts on one side only (`Robins` / `James Robins`), initial / word, prefix / word (`Des` / `Desiree`)"; the `Cause` vocabulary list gains `` `prefix` (a given-like word that starts the other's) `` after `` `initial` ``. `docs/conclusion-reconciliation.md` **Fold** bullet: after "an initial of it," insert "a given name it starts (*Des* → *Desiree*, S11-45; never a surname)," and in the `Fold:` bullet of `core/autoreconcile/names.go`'s file comment, "every word they pair is the same word or an initial of it. [J] → [James], [James] → [James, Kenneth]." becomes "every word they pair is the same word, an initial of it, or a given name it starts (never a surname). [J] → [James], [Des] → [Desiree], [James] → [James, Kenneth].", rewrapped at the comment's width. `.agents/skills/add-match-rule/SKILL.md`: "names add initials and alias tables through `wordOptions`" → "names add initials, alias tables and prefixes through `wordOptions`" | Working rule: each PR updates the docs and skill it makes false | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: `rg -n "CauseInitial|\"initial\"" core api macos/App`, `rg -n "wordOptions\{" core`, `rg -n "PrefixMinLength|prefixWord" core docs`, `rg -n "\.Broader" core api`, `rg -n "valuecompare\." core/database/searchindex`, `rg -n "lineContains" macos`, `rg -n '"ros"|"des"' core/database/namealiases/nicknames_seed.go`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `NameRules.PrefixCredit` *(new)* | registry field | `WordRules.InitialCredit`; `NameRules.PrefixMinLength` | A new tunable; beside its sibling knob (D-1) |
| `CausePrefix` *(new)* | cause code | `CauseInitial` | A rule of its own that callers word and guard apart (D-2) |
| `wordOptions.prefixCredit` *(new)* | struct field | `wordOptions.initials`, `wordOptions.aliases` | Extends the one word table's options (D-3) |
| `wordSimilarity`'s `broader` result *(new)* | return value | `name.pair`'s `IsInitial` rule | The table decides the side where it makes the match (D-3, D-5) |
| `valuecompare.PrefixRole` *(new)* | func, `core/valuecompare/words.go` | `AliasRoles`; `prefixWord`'s role check | One statement of the prefix roles for the shortlist, comparer and reconciler (D-4) |
| `sameResult` *(new)* | test helper, `core/autoreconcile/names_test.go` | `TestReconcileNamesAliases`' unchanged loop | Moved out so T-9 does not copy it (D-11) |
| `TestWordKeysMeetOnPrefixPairs` *(new)* | test, `core/valuecompare/shortlist_test.go` | `TestWordKeysShareOnSameOrEquivalent` | Keeps that test's name true; the prefix contract has its own name (T-5) |
| `ValueCause.prefix` *(new)* | Swift enum case | `ValueCause.initial` | The wire code's mirror (D-9) |
| `L10n.Promote.lineContainsPrefix`, key `promote.sheet.line.containsPrefix` *(new)* | L10n key | `lineContainsInitial` | One sentence per cause (D-9) |
| `WithPrefixes` *(new)* | test helper, `core/valuecompare/export_test.go` | `WordOptions` | Leaves `WordOptions`' six call sites alone |
| `withPrefixMin` *(new)* | test helper, `core/valuecompare/name_test.go` | `withTables` | Same shape, another field |
| `TestAPrefixNameIsOffered` *(new)* | test, `core/graphalign/align_test.go` | `TestOneGivenNameIsOfferedAFullerRecord` | Same shape for case 2 |
| `TestProposeAPrefixNameIsAWeakMatch` *(new)* | test, `core/database/promotealign/promotealign_test.go` | `TestProposeExtraGivenNamesAreAWeakMatch` | Same shape; a pair no nickname set holds |
| `TestPrefixNeverFoldsASurnameOrAcrossTypes` *(new)* | test, `core/autoreconcile/names_test.go` | `TestReconcileNamesAliases` "unchanged" table | The same with-and-without comparison, keyed on `PrefixMinLength` |

**Builds on.** `add-match-rule` "Add a knob" and "Add a relation, cause or unit code"; S11-46's `prefixWord` and `AliasRoles`; the initial rung's `wordOptions` shape (`c798a2fe`); S11-28's role-guarded fold allow-list (`ef6f29a4`); the sheet's `lineSentence` table.
**Deletes.** `name.pair`'s `IsInitial`-only broader rule (D-5); `prefixWord`'s inline role check (D-4); the copy of the unchanged-result loop T-9 would have made (D-11). `prefixWord`'s "from S11-45" wording (D-4). The name-candidates sentences D-12 replaces. No code becomes dead.
**Next adds.** Retuning the credit is one registry number. Another name-only word rung is one `wordOptions` field and one `wordSimilarity` branch. A new value cause on the sheet is one `ValueCause` case, one `lineSentence` arm and one key.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A shortlisted prefix pair now clears the candidate gate (0.7 ≥ `MinSimilarity` 0.4) | `core/graphalign/candidates.go` `Candidates`; Promote's target menu | Offered as an alternative like any candidate; `TopK` and `TieCap` bound a short stem ("Jo") that starts many names. No new state reaches Swift |
| A prefix pair with an agreeing surname reaches the weak band and is proposed | `core/graphalign` rows; `PromoteFlow.draftTarget` | Same path as an initial at 2.0 today (D-10); the row shows its band and evidence as any weak row |
| An evidence line with cause `prefix` | `L10n.Promote.lineSentence` | Worded by D-9; before this PR an unknown code would read "Matches, with less detail" |
| The reconciler folds a prefix into the given name it starts | Persons list and Person page titles (`autoreconciler` cache) | Rebuilt on open by the `CacheVersion` bump (D-7); an ambiguous prefix folds as an ambiguous initial does (AC-3.2) |

### Scale

At the map's Scale (S11-21's large fixture, 4,000 persons and a 12-generation tree): one more `strings.HasPrefix` pair per word pair that reached the end of the word table, and one `prefixWord` call per side per word pair in `name.word`. No new key, cache or loop. The shortlist cost was measured by S11-46; the candidate gate now keeps prefix pairs it used to drop, bounded by `TopK` and `TieCap`. `TestLargeSourceBudgets` stays inside its budgets (the planning probe passed it).

### Schema

None — no catalog change. The reconciler table is rebuilt on open by D-7.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `NameRules.PrefixCredit` *(new)* (D-1); `CausePrefix` *(new)* (D-2); `wordOptions.prefixCredit`, the prefix rung and the `broader` result in `wordSimilarity` (D-3); `PrefixRole` *(new)*, `name.word` sets the credit (D-4); `name.pair` copies the broader side (D-5) |
| `core/matchregistry` | `defaultValue`: `PrefixCredit: 0.7` (D-1) |
| `core/autoreconcile` | `nameModule.fold` allow-list case and doc comments (D-6, D-12) |
| `core/database/autoreconciler` | `CacheVersion` 20 (D-7) |

Errors: none.

### FFI

None — the wire carries the cause as a string (`engine.proto` `cause = 5`). `api/ffi/swift_vocabulary_test.go` gains `CausePrefix` in its "value causes" row (D-9).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Evidence line sentence | existing line (no component change) | `macos/App/Platform/L10n.swift` `L10n.Promote.lineSentence` | `.prefix` → `lineContainsPrefix` (D-9) |
| Cause code | `ValueCause.prefix` *(new)* | `macos/App/Platform/ConclusionVocabulary.swift` | Parsed from the wire (D-9) |

States: unchanged; the badge stays "Similar" (`Contains` is `OutcomePartial`) and the line starts unchecked. Accessibility: the sentence reads through the existing `spokenLine`.

### Strings

L10n keys *(new)*: `L10n.Promote.lineContainsPrefix` = "Matches by a shortened name" (`promote.sheet.line.containsPrefix`). Removed: none.

## 7. Steps

Every Go test command is run from the repo root with `CGO_ENABLED=1` and `-tags fts5`, as `scripts/check-pr.sh` runs them. `scripts/check-pr.sh` runs on macOS: off macOS it skips the Xcode checks and prints `check-pr: FAIL` for that reason alone, so a Linux run reports "every other check PASS, Xcode skipped" and the PASS line comes from a macOS run before the PR is marked ready. On Linux, open the PR as a draft and report STEP-6's `xcodebuild test`, T-11 and DoD box 3 as "not run: needs macOS"; they are run on macOS (by the user or CI on the non-draft PR) before the PR leaves draft.

**STEP-1 — The knob and the cause**
Files: `core/valuecompare/rules.go`, `core/matchregistry/value.go`, `core/valuecompare/valuecompare.go`, `docs/matching-model.md`. Skill: `add-match-rule` ("Add a knob", "Add a relation, cause or unit code").
Do: create the branch `s11-45-prefix-names` from an up-to-date `main`. Then D-1 (field, both doc comments, registry value, inventory rows) and D-2. No reader yet.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/matchregistry/... ./core/valuecompare/...` passes, with `TestKnobInventoryMatchesRegistry` green.

**STEP-2 — The word table's prefix rung**
Files: `core/valuecompare/words.go`, `core/valuecompare/export_test.go`, `core/valuecompare/words_test.go`. Skill: `add-match-rule`; rule `go-tests.mdc`.
Do: D-3, and `WithPrefixes` (D-11). Add T-1. `name.word` takes the fourth result as `_` and sets no credit yet (STEP-3 wires it), so no name changes.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/...` passes.

**STEP-3 — Names compare by prefix**
Files: `core/valuecompare/name.go`, `core/valuecompare/words.go` (`PrefixRole`, `prefixWord`), `core/valuecompare/aliases.go` (`AliasRoles` doc), `core/valuecompare/name_test.go`, `core/valuecompare/contract_test.go`, `core/valuecompare/shortlist_test.go`. Skill: `add-match-rule`.
Do: D-4, D-5. Add T-2, T-3, T-4, T-5 and `withPrefixMin`, `prefixCredit` (D-11).
Check: the `add-match-rule` Check command passes: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/segmentcompare/... ./core/kinship/... ./core/shortlist/... ./core/graphalign/... ./core/claimevidence/... ./core/matchregistry/...`, with `TestLayerPackages` green.

**STEP-4 — Promote offers a prefix name**
Files: `core/graphalign/align_test.go`, `core/database/promotealign/promotealign_test.go`.
Do: add T-6 and T-7. No production code.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/... ./core/database/promotealign/...` passes, `TestLargeSourceBudgets` included.

**STEP-5 — The reconciler folds a prefix**
Files: `core/autoreconcile/names.go`, `core/autoreconcile/names_test.go` (`sameResult`, T-8, T-9), `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`. Skill: `add-match-rule`.
Do: D-6 (case and `fold` doc; the file comment's `Fold:` bullet per D-12), D-7. Add T-8 and T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/... ./core/database/autoreconciler/... ./core/subjectprofile/...` passes.

**STEP-6 — The sheet says why**
Files: `macos/App/Platform/ConclusionVocabulary.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-9. Add T-10 and T-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `python3 scripts/check-localizable-xcstrings.py` passes; on macOS, `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes with no Swift warning.

**STEP-7 — Docs that describe the rule**
Files: `docs/deployment-plan/spike-11/name-candidates.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/conclusion-reconciliation.md`, `.agents/skills/add-match-rule/SKILL.md`.
Do: D-12, word for word.
Check: `rg -n "wrong for a prefix on side A|Des may be short for Desiree" docs .agents --glob '!docs/deployment-plan/spike-11/prs/**'` prints nothing; `rg -n "prefix" docs/conclusion-reconciliation.md docs/deployment-plan/spike-11/comparison-layers.md .agents/skills/add-match-rule/SKILL.md` prints a line in each file.

**STEP-8 — Whole suite**
Files: none.
Do: run `scripts/check-pr.sh` (on macOS for the PASS line; see the note above STEP-1). Then push the branch, made from `main` at STEP-1, and open the PR with §13's title and body.
Check: it prints `check-pr: PASS`. `git diff main --stat -- '*_test.go' '*Tests.swift'` lists only the test files §8 names and `core/valuecompare/export_test.go` (`WithPrefixes`, D-11).

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done, once the PR number is known.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-45-prefix-names.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence ending in a period, like the rest of `git log`. STEP-1: "Add the prefix credit and the prefix cause to value compare's registry and codes." STEP-2 and STEP-3 together: "Match a given name that starts another as a weaker hit, so Desiree finds Des." STEP-4: "Test that Promote offers a record whose given name starts the Subject's." STEP-5: "Fold a given name into the name it starts in the reconciler, never a surname." STEP-6: "Say on the evidence sheet when two names match by a shortened name." STEP-7: "Say in the matching and reconciler docs that a prefix matches and folds like an initial." STEP-9 is a follow-up docs commit on the open PR. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/words_test.go` `TestWordSimilarityNames` *(a `broader` column, six new rows)* | The row struct gains `broader valuecompare.Side`; the loop reads `sim, rel, why, broader := valuecompare.WordSimilarity(…)` and checks it. Existing rows: "initial with initials" expects `SideA`, every other existing row `SideNone` (the zero value). New rows, with `p := valuecompare.WithPrefixes(valuecompare.WordOptions(true, valuecompare.AliasIndex{}, given, given), 0.7)`: `"prefix", "des", "desiree", p` → 0.7, `Contains`, `CausePrefix`, `SideA`; `"prefix, longer first", "desiree", "des", p` → 0.7, `Contains`, `CausePrefix`, `SideB`; `"prefix off", "des", "desiree", valuecompare.WordOptions(true, valuecompare.AliasIndex{}, given, given)` → 0, `Different`, `CauseNoMatch`; `"spelling before prefix", "ann", "anne", p` → 0.75, `Near`, `CauseSpelling`; `"alias before prefix", "jim", "jimmy"`, with `valuecompare.WithPrefixes(valuecompare.WordOptions(true, aliasIndexOf(valuecompare.AliasNickname, []string{"jim", "jimmy"}), given, given), 0.7)` → `words.AliasCredit`, `Equivalent`, `CauseAlias`; `"not a prefix", "desmond", "desiree", p` → 0, `Different`, `CauseNoMatch` | AC-1.5, D-3, D-5 |
| T-2 | Go | `core/valuecompare/name_test.go` `nameCases` *(new rows, run by `TestNameCompare`)* | Names labelled `S11-45 AC-…`: `{a: nv("g=Desiree"), b: nv("g=Des"), rel: Contains, sim: prefixCredit, why: CausePrefix, broader: SideB, parts: [pair(PartGiven, "desiree", "des", Contains, prefixCredit, CausePrefix)]}`; the same swapped with `broader: SideA`; `nv("g=Des|s=Long")` / `nv("g=Desiree|s=Long")` → `Contains`, `(prefixCredit*2*G1 + 2*F) / (2*G1 + 2*F)`, `CausePrefix`, `SideA`; `differentWhy(…, nv("g=Desmond"), nv("g=Desiree"), CauseNoMatch)`; `differentWhy(…, nv("s=Mac"), nv("s=MacDonald"), CauseNoMatch)`; `differentWhy(…, nv("g=Mary|s=Mac"), nv("g=Mary|s=MacDonald"), CauseNoMatch)`; `nv("g=Des")` / `nv("g=Desi")` → `Near`, `anne`, `CauseSpelling`; `nv("g=D.")` / `nv("g=Desiree")` → `Contains`, 0.5, `CauseInitial`, `SideA`; `nv("g=Dés")` / `nv("g=Desiree")` → `Contains`, `prefixCredit`, `CausePrefix`, `SideA`; `nv("g=De")` / `nv("g=Desiree")` → `Contains`, `prefixCredit`, `CausePrefix`, `SideA`; `rules: withPrefixMin(3)`, `nv("g=De")` / `nv("g=Desiree")` → `Different`, `CauseNoMatch`; `rules: withPrefixMin(0)`, `nv("g=Des")` / `nv("g=Desiree")` → `Different`, `CauseNoMatch`; `nv("n=Des")` / `nv("g=Desiree")` → `Different`, `CauseBelowMinimum`; `nv("g=James|g=Des")` / `nv("g=J.|g=Desiree")` → `Different`, `CauseBelowMinimum`. The existing row "AC-1.3 nickname" (`withTables(nicknameTable)`, `Jim` / `James`) stays `Equivalent` `alias` unchanged | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, D-4, D-5 |
| T-3 | Go | `core/valuecompare/name_test.go` `TestNameLadders` *(one ladder, one score row)* | Ladder `{"g=Desiree|s=Long", []string{"g=Desiree|s=Long", "g=Désirée|s=Long", "g=Des|s=Long", "g=D.|s=Long"}, []string{"g=Desmond|s=Long", "g=Desiree|s=Short"}}`; score row `{"g=Desiree|s=Long", "g=Des|s=Long", (prefixCredit*2*G1 + 2*F) / (2*G1 + 2*F)}` | AC-1.11 |
| T-4 | Go | `core/valuecompare/contract_test.go` `samples`, `namePools` *(edited)* | Name samples, before the two non-`ofName` entries: `ofName("g=Des|s=Long", "g=Desiree|s=Long")`, `ofName("g=Desiree", "g=Des")`, `ofName("n=Des", "g=Desiree")`, `ofName("g=Mary|s=Mac", "g=Mary|s=MacDonald")`, `ofName("g=James|g=Des", "g=J.|g=Desiree")`. Pools: `"g"` gains `"Des", "Desiree", "Al", "Albert"`; `"s"` gains `"Mac", "MacDonald"`. `TestComparerContract`, `TestKeyContract` and `TestNameInvariants` pass unchanged | AC-4.1 |
| T-5 | Go | `core/valuecompare/shortlist_test.go` `TestWordKeysMeetOnPrefixPairs` *(new, after `TestWordKeysReachPrefixes`)* | Under `billRules()`, every ordered pair of the name corpus `"Des Long", "Desiree Long", "Al Robins", "Albert Robins", "Mary Mac", "Mary MacDonald", "Desmond Long"` (each through `slName`) that `Compare` gives cause `prefix`: `ia, la := WordKeys(a)`, `ib, lb := WordKeys(b)`, fail unless `shares(ia, lb) && shares(ib, la)`, printing all four lists. The test fails with "the corpus met no prefix pair" when no pair has cause `prefix`. `TestWordKeysShareOnSameOrEquivalent` is unchanged (review M-5) | AC-4.2 |
| T-6 | Go | `core/graphalign/align_test.go` `TestAPrefixNameIsOffered` *(new, after `TestOneGivenNameIsOfferedAFullerRecord`)* | Subtest "a given name reaches the record it starts": Subject `{ID: id("s-desiree"), Ref: "PER-S", Kind: "person", Values: vals(givenProp("Desiree"))}`, `Metas: subjectcomparetest.Metas()`; canon handles `h-des` (PER-1, `vals(givenProp("Des"))`) and `h-desmond` (PER-2, `vals(givenProp("Desmond"))`); `alignWith(…, matchregistry.Default().AlignConfig())` → row target not `TargetHandle`, `AssessNone`, alternatives exactly `[h-des]` with `AssessNone`. Subtest "a prefix with the same surname is a weak match": Subject `vals(nameProp("Desiree Long"))`, handle `h-des` (PER-1, `vals(nameProp("Des Long"))`) → `TargetHandle`, `h-des`, `AssessWeak`, `ReasonAgrees`, score within 1e-9 of 2.2 (name agree 2.5 × similarity 0.88) | AC-1.8, AC-1.9, D-10 |
| T-7 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeAPrefixNameIsAWeakMatch` *(new, after `TestProposeExtraGivenNamesAreAWeakMatch`)* | Table `{filed, other, probe string}`: `{"Rosalind Long", "Mary Long", "Ros Long"}` and `{"Ros Long", "Mary Long", "Rosalind Long"}`. Per row, as `TestProposeExtraGivenNamesAreAWeakMatch`: `f := newFixture(t)`; `known := f.promote(f.person(f.source, f.artifact, filed))`; `f.promote(f.person(f.source, f.artifact, other))`; `srcB, artB := f.newSource("Census")`; `again := f.person(srcB, artB, probe)`; `row := rowFor(f.propose(f.graph(srcB), nil), again.ID)` → `TargetHandle`, `known.Entity.ID`, `AssessWeak`, score in `[Bands.Weak, Bands.Medium)` | AC-1.10 |
| T-8 | Go | `core/autoreconcile/names_test.go` `TestReconcileNames` *(three new "fold" rows, after "an initial never folds across types")* | `{"fold", "a prefix expands: [Des] into [Desiree]", [nm(1, "given=Des|surname=Long"), nm(2, "given=Desiree|surname=Long")], [{[1 2], "given=Desiree|surname=Long"}], StateMerged, "kept"}`; `{"fold", "the prefix expansion does not depend on rank", [nm(1, "given=Desiree|surname=Long"), nm(2, "given=Des|surname=Long")], [{[1 2], "given=Desiree|surname=Long"}], StateMerged, "kept"}`; `{"fold", "an ambiguous prefix folds into the best ranked of equal support", [nm(1, "given=Albert|surname=Long"), nm(2, "given=Alice|surname=Long"), nm(3, "given=Al|surname=Long")], [{[1 2 3], "given=Albert|given=Alice|surname=Long"}], StateMerged, "kept"}` | AC-3.1, AC-3.2, D-6 |
| T-9 | Go | `core/autoreconcile/names_test.go` `TestPrefixNeverFoldsASurnameOrAcrossTypes` *(new, after `TestReconcileNamesAliases`)*; `core/database/autoreconciler/autoreconciler_test.go` `productAliasPin` *(edited)* | Inputs `[nm(1, "given=Mary|surname=Mac"), nm(2, "given=Mary|surname=MacDonald")]`, `[nm(1, "nick=Des|surname=Long"), nm(2, "given=Desiree|surname=Long")]`, `[nm(1, "given=Desmond|surname=Long"), nm(2, "given=Desiree|surname=Long")]`: `Reconcile(testRules, …)` equals `Reconcile(off, …)` where `off := testRules; off.Names.PrefixMinLength = 0`, through `sameResult(t, …, testRules, off, in)` (D-11), which `TestReconcileNamesAliases`' unchanged loop now calls too. `productAliasPin.CacheVersion` 20; `TestProductAliasTablesPinCacheVersion` passes | AC-3.3, AC-3.4, D-6, D-7 |
| T-10 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "value causes" *(edited)* | `want` gains `string(valuecompare.CausePrefix)`; passes once `ValueCause` has `case "prefix"` | AC-2.2 |
| T-11 | Swift | `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift` `lineSentencesFollowTheTable` *(one line in the value loop)* | `#expect(sentence(.contains, cause: .prefix, value) == key(L10n.Promote.lineContainsPrefix))` after the `.initial` line | AC-2.1, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/valuecompare` (`TestNameInvariants` with the new pool words, `TestNameLadders`), `core/autoreconcile`, `core/database/autoreconciler` (`TestProductAliasTablesPinCacheVersion`), `core/database/promotealign` (`TestLargeSourceBudgets`, goldens), `api/ffi` (`TestSwiftVocabularyMatchesGo`), `core/matchregistry` (`TestKnobInventoryMatchesRegistry`, `TestLayerPackages`'s numbers guard).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: AC-1.1 – AC-1.7 → T-2 (AC-1.5 also T-1); AC-1.8, AC-1.9 → T-6; AC-1.10 → T-7; AC-1.11 → T-3; AC-2.1 → T-11; AC-2.2 → T-10; AC-3.1, AC-3.2 → T-8; AC-3.3, AC-3.4 → T-9; AC-4.1 → T-4; AC-4.2 → T-5
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS on macOS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings (promote.sheet.line.containsPrefix); no stale strings left
- [ ] Sweep: `rg -n "wrong for a prefix on side A|Des may be short for Desiree|from S11-45, the comparer" .agents .cursor/rules docs core --glob '!docs/deployment-plan/spike-11/prs/**'` has no hit (plan files quote the old text on purpose); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — n/a, none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no chrome
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status Landed #N; completed.md: one short paragraph of what shipped, newest first
- [ ] Brief moved to design/archive/ and design/README.md updated — n/a, no brief
- [ ] Domain docs updated: docs/matching-model.md (inventory rows), docs/conclusion-reconciliation.md (fold), comparison-layers.md, name-candidates.md, add-match-rule skill (D-1, D-12)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project made after S11-44 (or the dogfood project after the one-off command). File a person "Desmond Long" from one Source. In a new Source, add a person "Des Long" and open Promote: the row targets "Desmond Long" at the weak band (no nickname set holds Desmond). Open the evidence sheet: the name line reads "Matches by a shortened name", badged Similar and unchecked.
2. In the same project, file "Rosalind Long", then promote a new Source's person whose name has only the given name "Ros": the row opens unset, and its target menu lists "Rosalind Long".
3. Add a person "Mary Mac" in a new Source with "Mary MacDonald" filed: Promote does not offer "Mary MacDonald".
4. File the Source person "Des Long" onto "Desmond Long" (step 1). The Persons list shows one given name, "Desmond", for that person.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A short stem ("Al", "Jo") becomes a candidate for many records | `TopK` and `TieCap` bound the candidates; a prefix pair alone stays under the weak band (1.75), so it is offered, not proposed |
| A prefix pair with the same surname is proposed weak on the wrong person ("Al Long" on "Alice Long") | The same as an initial today ("J. Long"); the band is weak and the researcher decides. Retuning is a dogfood call (Out) |
| The reconciler folds a surname prefix | D-6's role check; T-9 pins it |
| The `CacheVersion` bump is missed | `TestProductAliasTablesPinCacheVersion` fails until the pin and the constant agree (D-7) |

**Stop and report if:**

- Any existing test or golden fails after STEP-2, STEP-3 or STEP-5 (the planning probe passed the whole suite): name it and its output; do not edit it.
- `TestNameInvariants` or `TestComparerContract` fails on a generated pair after T-4: report the seed and pair; do not drop a pool word.
- `TestLargeSourceBudgets` fails in two of three isolated runs (`CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v -count=1`): report the timing lines; do not raise a budget.
- `rg -n '"ros"|"rosalind"|"desmond"' core/database/namealiases/nicknames_seed.go` finds a set: T-7 and the manual check would test an alias, not the prefix.
- `prefixWord` no longer exists with S11-46's meaning, or `nameUnit.compareValue` no longer strips part types (D-6's reason would change).
- T-8's expected string differs from the result only in the order of given parts: report the result; do not change the reconciler.

## 12. Open questions

None.

## 13. PR

**Title:** S11-45 — Prefix names
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-45-prefix-names.md`.
