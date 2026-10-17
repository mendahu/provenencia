# S11-02 — Value compare core and match registry

**Status:** Landed in [#441](https://github.com/mendahu/provenencia/pull/441) (2026-10-17)
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: new Go packages, a guard test and a skill)
**Board:** —
**Depends on:** S11-01 (landed [#432](https://github.com/mendahu/provenencia/pull/432)), S11-22 (landed [#434](https://github.com/mendahu/provenencia/pull/434)), S11-23 (landed [#438](https://github.com/mendahu/provenencia/pull/438))
**Blocks:** S11-03, S11-04
**Branch:** `s11-02-value-compare-core`
**Version bump:** none — no caller uses the new comparers yet and the moved word helpers behave exactly as before ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`plan-pr`](../../../../.agents/skills/plan-pr/SKILL.md) (only its "Touches" table, which STEP-9 extends); this PR writes `add-match-rule` (new)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [layer 1](../comparison-layers.md#layer-1-value-comparison), [relations](../comparison-layers.md#relations), [detail](../comparison-layers.md#detail), [purity](../comparison-layers.md#purity-requirements), [layer 1 tests](../comparison-layers.md#tests), [match registry](../comparison-layers.md#match-registry), [reuse: `Key`](../comparison-layers.md#reuse-across-the-app); [`docs/matching-model.md`](../../../matching-model.md) knob inventory; working rules "Build, then migrate", "High tolerance for churn", "Pure means the package builds and tests with no build tags…", "Value compare returns a relation, a similarity and a detail", "A normalized `Key` per value type", "One match registry", "One guard, not one per package", "One matching skill, grown with the layers", "Registry numbers: carry over, then estimate"
**Planned:** 2026-10-17 against `main` @ `cf1f051` (S11-23 landed)
**Review:** [`S11-02-value-compare-core.review.md`](S11-02-value-compare-core.review.md) — Fits with changes / Ready with changes → fixed, 2 rounds

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

Layer 1 of the matching stack gets its package and its contract: `core/valuecompare` relates two values and returns a relation, a similarity and a structured detail, and gives each value a normalized `Key`. This PR builds the contract and the text, term and integer comparers; names (S11-03) and dates (S11-04) plug into the same contract. `core/matchregistry` opens with `Default()` and the value section's text, term and integer knobs. One repo-walking guard keeps the layer packages pure and free of numbers, and a second guard keeps the knob inventory in [`matching-model.md`](../../../matching-model.md) in step with the registry. The first version of the matching skill, `add-match-rule`, writes down the steps every later layer PR repeats.

**Why now.** Slice 1 landed: the model, the pure value types (S11-22) and the primary kinds (S11-23). Every later layer PR builds on this contract and this registry.

**What stays.** Nothing changes for researchers. No caller moves onto `valuecompare.Compare` here (S11-05, S11-06, S11-07 do that). The four word helpers that move out of `core/autoreconcile` keep their behavior and signatures; `core/match` and the reconciler call them in their new home.

## 2. User stories

**US-1.** As a contributor building a matching layer, I compare two text, term or integer values through one entry point and read what kind of match it is, how close, and why, with no database.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `go test ./core/valuecompare/... ./core/matchregistry/...` with no build tags | Passes |
| AC-1.2 | Text `York` / `york` | `Same`, similarity 1, cause `case` |
| AC-1.3 | Text `Montréal` / `Montreal` | `Equivalent`, similarity 0.95, cause `fold` |
| AC-1.4 | Text `York` / `York, Upper Canada` | `Contains`, broader side A, similarity 0.7 × (1 + 1) / (1 + 3) = 0.35, cause `missing_part`, `OnlyB` = `upper`, `canada` |
| AC-1.5 | Text `York, Ontario` / `Toronto, York` | `Overlaps`, similarity 0.7 × (1 + 1) / (2 + 2) = 0.35, cause `overlap` |
| AC-1.6 | Text `Upper Canada, York` / `York, Upper Canada` | `Different`, similarity 0, cause `order` (the worked case in [relations](../comparison-layers.md#relations)) |
| AC-1.7 | Text `Provenance` / `Provenanced`; `Upper Canada` / `Upper Kanada`; `Ann` / `Anne` | `Near`, 1 − 1/11 ≈ 0.909; `Near`, (1 + 1 + 5/6 + 5/6) / 4 ≈ 0.917; `Near`, 0.75; each cause `spelling` |
| AC-1.8 | Text `Mary` / `Mark`; text `!!!` / `York` | `Different`, 0, cause `no_match`; `Unknown`, 0, cause `empty` |
| AC-1.9 | Term `male` / `male`; `male` / `female`; `` / `male` | `Same` 1 `exact`; `Different` 0 `no_match`; `Unknown` 0 `empty` |
| AC-1.10 | Integer 3 / 3; 3 / 4 at tolerance 0; 3 / 4 and 3 / 6 at tolerance 2 | `Same` 1; `Different` 0 `beyond_tolerance` distance 1 unit `count`; `Near` 1 − 1/3 ≈ 0.667 `gap` distance 1; `Different` 0 `beyond_tolerance` distance 3 |
| AC-1.11 | A name, date, subject or depiction value, or two values of different types | `Unknown`, cause `unsupported` |
| AC-1.12 | Values carrying refs `r1` / `r2` | `Detail.RefA` is `r1`, `Detail.RefB` is `r2`, for every relation |

**US-2.** As a contributor, I can trust every comparer to keep the same contract, and a key to agree with `Compare`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Every value type with a comparer, over its sample table and 500 generated pairs from a fixed seed | Similarity in 0…1; `Same` ⇔ 1; for relations other than `Unknown`, `Different` ⇔ 0; `Equivalent` ≥ 0.9; `Contains`, `Overlaps`, `Near` strictly between 0 and 1; `Unknown` has similarity 0 |
| AC-2.2 | The same pairs with sides swapped | Same relation and similarity; `Contains`' broader side flips |
| AC-2.3 | Any value compared with itself | `Same` (or `Unknown` when it is empty) |
| AC-2.4 | The same pairs with `Particles` shuffled | Same results |
| AC-2.5 | The same pairs through `Key` | `Key(a, Same) == Key(b, Same)` exactly when the relation is `Same`; `Key(a, Equivalent) == Key(b, Equivalent)` exactly when it is `Same` or `Equivalent`; `Key` of an empty or unsupported value, or at any other level, is `"", false` |
| AC-2.6 | A value type gains a comparer with no sample row in the contract test | `TestComparerContract` fails |

**US-3.** As a contributor, I find every matching number in one place, and a guard stops me from writing one anywhere else.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A numeric literal other than 0 or 1 is added to a non-test file in `core/valuecompare` | `TestLayerPackages/core/valuecompare/numbers` fails and names the file and line |
| AC-3.2 | A listed package imports anything under `core/database`, directly or through another package, or has a file that builds only with a tag | `TestLayerPackages/<package>/pure` fails and names the import chain or file |
| AC-3.3 | A field is added to `valuecompare.Rules` without a row in the knob inventory | `TestKnobInventoryMatchesRegistry` fails naming the dotted key |
| AC-3.4 | An inventory row names no registry field and is not under a not-yet-built prefix | `TestKnobInventoryMatchesRegistry` fails naming the row |
| AC-3.5 | A not-yet-built prefix matches a registry field, or matches no inventory row | `TestKnobInventoryMatchesRegistry` fails naming the prefix |
| AC-3.6 | Two calls to `matchregistry.Default()`; the first result's `Particles` and `Folds` are changed | The second result is unchanged |

**US-4.** As a researcher, nothing behaves differently.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The whole Go and Xcode suite | `scripts/check-pr.sh` prints `check-pr: PASS`; the only test changes outside the two new packages are the three moved tests (D-9) and import paths |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/valuecompare` (new): `Relation`, `Cause`, `Side`, `PartKind`, `Unit`, `Result`, `Detail`, `Part`, `PartMatch`, `Value`, `Rules` and its groups, `Compare`, `Key` | STEP-1, AC-1.11, AC-1.12 |
| `core/matchregistry` (new): `Registry`, `Default()`, the value section's text, term and integer knobs | STEP-2, AC-3.6 |
| `NormalizeForm`, `IsInitial`, `EditDistance`, `SpellingSimilarity` move from `core/autoreconcile` to `core/valuecompare` (pulled forward from S11-03, D-8); `SpellingFloor`, `SpellingShortMin` deleted; `core/match` reads the knobs both packages hold from the registry (D-10) | STEP-2, AC-4.1 |
| Term and integer comparers with `Key` | STEP-3, AC-1.9, AC-1.10 |
| Text comparer with `Key` | STEP-4, AC-1.2 – AC-1.8 |
| Contract tests over every comparer, with generated pairs | STEP-5, AC-2.1 – AC-2.6 |
| One repo-walking guard: purity and numeric literals over a list of layer packages | STEP-6, AC-3.1, AC-3.2 |
| The knob-inventory drift guard | STEP-7, AC-3.3 – AC-3.5 |
| The matching skill `add-match-rule`, first version, with its Cursor rule and a row in `plan-pr`'s "Touches" table | STEP-8 |
| Docs: map row; the inventory's lead paragraph; the cause list; the three docs that name the moved helpers | STEP-10 |

| Out | Home |
| --- | --- |
| The name comparer, `Names` and the `Words` knobs only names use (`InitialCredit`, `AliasCredit`, `MaxExactPairing`), alias tables, name patterns | S11-03 |
| The date comparer and `Dates` knobs | S11-04 |
| Moving any caller onto `valuecompare.Compare` or `Key`; deleting `match.TextComparer`, `TermComparer`, `IntegerComparer` and `core/match/words.go` | S11-05 (comparers), S11-06 (reconciler), S11-07 (`Key` callers) |
| A term's neutral flag, polarity, cardinality | S11-09 (layer 2) |
| Typed registry keys (`SubjectType`, `PropertyKey`, `BridgeKey`) | S11-10, the first keyed section (D-14; the S11-10 scope entry says so) |
| The subject, segment, alignment and evidence sections | S11-10, S11-14, S11-16, S11-19 |
| Swift parsing of relation and cause codes | S11-17, S11-25 (working rule "One wire vocabulary") |
| A minimum similarity for text or integer `Near` | not planned: the spelling rule and the tolerance already decide `Near` (D-6) |
| Folding in `NormalizeForm` (José / Jose as one reconciler group) | S11-06 decides the reconciler's grouping policy |

## 4. Current state

Verified at `cf1f051`.

| Area | Where | Today |
| --- | --- | --- |
| Word helpers in the reconciler | `core/autoreconcile/autoreconcile.go` `NormalizeForm` (L298); `core/autoreconcile/names.go` `IsInitial` (L41); core/autoreconcile/spelling.go `SpellingFloor` (0.8), `SpellingShortMin` (3), `SpellingSimilarity`, `EditDistance` | Value comparison living in the reconciler. `core/autoreconcile` imports `core/database/depictionvalues`, so a pure layer cannot import it |
| Reconciler callers | `autoreconcile.go` L268 `SortKey` (`NormalizeForm`); `names.go` L62 `readName` (`NormalizeForm`), L119 `partFits` (`IsInitial`), L133 `outvotes` (`SpellingSimilarity(…, SpellingFloor, SpellingShortMin)`) | Same package today |
| Matching callers | `core/match/words.go` L110, L116, L151 (`NormalizeForm`), L232 (`IsInitial`), L245 (`SpellingSimilarity`); `core/match/frequency.go` L36, L59 (`IsInitial`); `core/match/registry.go` L52, L57 (`SpellingFloor`, `SpellingShortMin`) | The only callers outside the package |
| Word tests | `core/autoreconcile/autoreconcile_test.go` `TestNormalizeForm` (L204); `core/autoreconcile/names_test.go` `TestIsInitial` (L526); `core/autoreconcile/modules_test.go` `TestSpellingSimilarity` (L148); `names_test.go` L470, L475 call `NormalizeForm` inside `TestReconcileNamesInvariants` | Test the helpers that move |
| Today's text comparer | `core/match/names.go` L162 and `core/match/frequency.go` L35, L39, L63 call `splitWords` / `foldWord`; `core/match/compare.go` `TextComparer` (L69), `TermComparer` (L84), `IntegerComparer` (L97), `ComparerFor` (L37); `core/match/words.go` `splitWords`, `rawTokens`, `spacedParticle`, `periodInitial`, `foldWord`, `compareForms`, `wordSimilarity`, `wordDice`, `bestAssignment`, `foldOnly` (L14, 0.95) | Returns a bare 0…1 and a comparable flag; text pairs words in any order, so `Upper Canada, York` / `York, Upper Canada` is 1; a neutral term is not comparable (a layer 2 decision) |
| Today's knobs | `core/match/registry.go` `SpacedParticles` (L31), `LetterFolds` (L36), `DefaultWordRules` (L48), `DefaultText` (L238, `Partial` 0.7), `DefaultIntegers` (L244, `Tolerance` 0) | The values the value section carries over |
| Registry guard in `match` | `core/match/registry_test.go` `TestRegistrySetsEverySetting` (L15) | Checks every pointer setting has a value; unaffected by D-10, since the values stay set |
| Knob inventory | `docs/matching-model.md` "Knob inventory": 77 rows keyed `section.Group.Field`, lead paragraph says "Keys are dotted paths of the planned registry" | No test reads it |
| Value-type names | `core/propertymeta/propertymeta.go` `ValueTypeText`, `ValueTypeInteger`, `ValueTypeDate`, `ValueTypeName`, `ValueTypeSubject`, `ValueTypeTerm`, `ValueTypeDepiction` | The value-type vocabulary `valuecompare` dispatches on |
| Pure packages | `core/datevalue`, `core/namevalue` (and `core/namevalue/namevaluetest`), `core/propertymeta`, `core/vocabkey` | Pure today (`go list -deps` has nothing under `core/database`); no guard keeps them so |
| Repo-walking precedent | `core/writes/begin_test.go` `TestBeginOnlyInDatabaseAndWrites` | Walks the repo with `go/parser` and fails with a list of hits |
| Repo-file precedent | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | A Go test reading a non-Go file in the repo by relative path |
| Skills | `.agents/skills/plan-pr/SKILL.md` "Touches" table (L127); `.cursor/rules/name-values.mdc` (shape of a skill's Cursor rule) | No matching skill |
| Docs naming the moved helpers | `docs/conclusion-reconciliation.md` L166 (`SpellingSimilarity`); `docs/ideas/daitch-mokotoff.md` L5 (says the code moves in S11-03 / S11-05) and L19 (links core/autoreconcile/spelling.go); `docs/deployment-plan/spike-11/comparison-layers.md` L410 (names core/autoreconcile/spelling.go) | Will point at a deleted file or the wrong PR |

**Precedent.** S11-22 ([plan](S11-22-pure-value-types.md), [#434](https://github.com/mendahu/provenencia/pull/434)) moved code into pure packages with no aliases and repointed every caller in the same commit; copy that for STEP-3. `TestBeginOnlyInDatabaseAndWrites` is the shape of STEP-7's walk. `TestRegistrySetsEverySetting` walks a registry struct with `reflect`; STEP-8 walks `Default()` the same way. `TestSwiftVocabularyMatchesGo` reads a repo file the way STEP-8 reads `matching-model.md`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Section types live in the layer; values live in the registry.** `valuecompare` declares `Rules` and its groups (field types and doc comments, no numbers). `core/matchregistry` declares `type Registry struct { Value valuecompare.Rules }` and `func Default() Registry`, with the numbers in `value.go`. `matchregistry` imports layer packages; no layer imports `matchregistry`. Layer tests that need the registry are external (`package valuecompare_test`) | Each layer takes its section as an argument (rule 5). Later sections need layer types (the evidence section names `valuecompare.Relation`, S11-19); with the import this way round there is no cycle | Section types declared in `matchregistry` and imported by layers — because `matchregistry` could then never name a layer type, such as a relation, without a cycle |
| D-2 | `Default()` builds a new `Registry` on each call, with new slices and maps, so a caller that changes its copy changes no one else's | Tests override a section by editing their copy (rule 5) | A package `var` — because one test's edit would leak into the next |
| D-3 | **API.** `func Compare(a, b Value, rules Rules) (Result, Detail)` and `func Key(v Value, level Relation, rules Rules) (string, bool)`; nothing else is exported for dispatch. Behind them an unexported `comparer` interface (`compare(a, b Value, r Rules) (Result, Detail)`, `key(v Value, level Relation, r Rules) (string, bool)`) and an unexported `comparers map[string]comparer` keyed by `propertymeta.ValueType*`, holding `text`, `term`, `integer`. `Compare` returns `Unknown` / `unsupported` when `a.Type != b.Type` or the type has no comparer, and copies `a.Ref` / `b.Ref` into `Detail.RefA` / `RefB` for every result. `core/valuecompare/export_test.go` (new) exposes `var ComparerTypes = func() []string` (the sorted keys of `comparers`) to the external tests | One entry point per value type behind one interface; S11-03 and S11-04 each add one map entry | Exported per-type comparers — because callers would pick one by hand, as `match.ComparerFor` makes them do; an exported `Supports(valueType)` — because `Compare`'s `unsupported` cause already says it and no named caller needs it |
| D-4 | **`Value`** is `struct { Type string; Text string; Term string; Integer int64; Ref []byte }`. `Type` is a `propertymeta.ValueType*` constant. Text is present when `NormalizeForm` leaves a word; a term when its key is non-empty; an integer value is always present. No neutral flag, polarity or label | Layer 1 never decides evidence (§ "What layer 1 may take"); `Ref` is `[]byte` because Observation ids are `[]byte` throughout (`match.Value.ObservationID`) | `Ref any` — because every caller would type-assert it back |
| D-5 | **Relation and Cause are string codes.** `type Relation string` with `Same = "same"`, `Equivalent = "equivalent"`, `Contains = "contains"`, `Overlaps = "overlaps"`, `Near = "near"`, `Different = "different"`, `Unknown = "unknown"`. `type Cause string` with the design's codes (`exact`, `case`, `fold`, `alias`, `abbreviation`, `initial`, `spelling`, `precision`, `missing_part`, `overlap`, `gap`, `beyond_tolerance`, `generation_conflict`, `empty`, `phrase_only`) plus three new ones: `no_match` (nothing in common), `order` (text whose shared words are out of order, D-7), `unsupported` (no comparer, or two value types). All codes are declared now, used or not. `type Side string` with `SideA = "a"`, `SideB = "b"`, `""` for none. `type PartKind string` with `PartWord = "word"` (names add theirs in S11-03). `type Unit string` with `UnitCount = "count"` (dates add `UnitDays` in S11-04) | Codes reach the wire and UI copy later (rule "One wire vocabulary"), so strings, not iota; declaring the full cause set once keeps S11-03 and S11-04 from inventing spellings | `iota` ints — because the wire and the Swift guard read strings |
| D-6 | **Text and integer have no `Near` minimum knob.** Text `Near` comes only from the spelling rule (`FuzzyFloor`, `ShortVariantMinLength`), integer `Near` only from `Tolerance` | The inventory has no such rows; a second floor would contradict the spelling rule (`Ann` / `Anne` is 0.75, below 0.8, and is a variant) | `value.Text.NearMinimum` = 0.8 — because it would turn `Ann` / `Anne` into `Different` |
| D-7 | **Text compares words in order.** Normalize with `NormalizeForm`, split with particles rejoined (today's `splitWords` rule). In order: (1) no words on a side → `Unknown`, `empty`; (2) the word sequences are equal → `Same`, cause `exact` when the raw strings are byte-equal, else `case`; (3) the folded sequences are equal → `Equivalent`, cause `fold`, similarity the mean of the positional scores, 1 for a byte-equal token pair and `FoldCredit` for any other (the word table below is not used in steps 2 and 3, so `É` / `E` is `Equivalent` at `FoldCredit`; the initial rule applies only in step 4); (4) otherwise align the words in order: the order-preserving pairing with the highest total word similarity, a word pair counting only above 0, computed by dynamic programming over A's words (rows) and B's words (columns). **Tie-break:** among equal totals, the most pairs; then, walking back from the last cell, take the diagonal before skipping an A word, and skipping an A word before skipping a B word. Each pair gets a `PartMatch` by the word table below; `Detail.Parts` lists the pairs in A's order, `OnlyA` / `OnlyB` the unpaired words. Then: no pair → `Different`, `no_match`; **order:** any unpaired word of A and any unpaired word of B (every combination is checked) whose word similarity is above 0 → `Different`, `order`; unpaired words on both sides → `Overlaps`, `overlap`; unpaired words on one side → `Contains`, `missing_part`, `Broader` the side with no unpaired word; none unpaired → `Near`, `spelling`. **Similarity:** `Different` and `Unknown` are 0. Otherwise every word on either side scores its pair's similarity, an unpaired word scores 0, and the value's similarity is the mean over all words on both sides; for `Contains` and `Overlaps` that mean is multiplied by `Text.Partial`. **Word table** (`wordSimilarity`, in this order): byte-equal tokens → 1, `Same`, `exact`; either token `IsInitial` → 0, no pair (text has no initials); `FoldWord` of both equal → `FoldCredit`, `Equivalent`, `fold`; `SpellingSimilarity` of the folded tokens above 0 (the 0.8 ratio or the short-word rule, so `ann` / `anne` 0.75 counts) → that ratio, `Near`, `spelling`; else 0, no pair. `Parts` is filled for every relation in which a pair exists, including `Same` (every positional pair) and `Different` / `order`; it is empty for `no_match` and `Unknown` | Satisfies every text row in [relations](../comparison-layers.md#relations), including the confirmed worked case `Upper Canada, York` / `York, Upper Canada` → `Different`; keeps today's numbers for `York` / `York, Upper Canada` (0.35) and one-word variants; the mean over both sides is the design's "weighted mean of the part similarities" with one-sided parts at 0, written without a numeric literal | Today's any-order pairing — because it calls the reordered toponym `Same`; dropping step (4)'s order check — because the reordered toponym would read `Overlaps` |
| D-8 | **The word helpers move in this PR, not S11-03.** `NormalizeForm`, `IsInitial`, `EditDistance`, `SpellingSimilarity` move from `core/autoreconcile` to `core/valuecompare/words.go` (new) with the same signatures and doc comments. From `core/match/words.go`, `splitWords`, `rawTokens`, `spacedParticle`, `periodInitial` and `foldWord` move there too: `splitWords` becomes exported `SplitWords(s string, particles map[string]bool) []string` (the set of particle tokens; the text comparer builds it once per call from `WordRules.Particles`), `foldWord` becomes exported `FoldWord(word string, folds map[rune]string) string`, the other three stay unexported; `core/match/names.go`, `frequency.go` and `words.go` call `valuecompare.SplitWords` / `FoldWord` and the `match` copies are deleted. `match` keeps only its pairing, alias and word-similarity code until S11-05. `SpellingFloor` and `SpellingShortMin` are deleted; their values are `value.Words.FuzzyFloor` and `value.Words.ShortVariantMinLength`. `core/autoreconcile` declares `var spelling = matchregistry.Default().Value.Words` in `names.go` (its only `matchregistry` import); because `matchregistry` imports every layer, no layer package may import `core/autoreconcile` (the skill's "do not" list says so, and S11-06 decides whether the reconciler's caller passes the rules in instead) and `outvotes` calls `valuecompare.SpellingSimilarity(w[i], o[i], spelling.FuzzyFloor, spelling.ShortVariantMinLength)`. `EditDistance` is rewritten without the literal 2: inside the loops declare `pi, pj := i-1, j-1`, and the adjacent-swap branch reads `if pi > 0 && pj > 0 && a[pi] == b[pj-1] && a[pi-1] == b[pj] { d[i][j] = min(d[i][j], d[pi-1][pj-1]+1) }`; `TestSpellingSimilarity`'s `robnis` / `robins` case covers the swap | The text comparer needs them and cannot import `core/autoreconcile` (impure, §4); a copy would leave two spelling rules; the reconciler "imports the spelling floor from value compare" (map rule) | Copying them until S11-03 or S11-05 — because S11-03 and S11-04 run in parallel and a tokenizer or fold fix would land in one copy only; deferring the text comparer — because the map gives text to this PR |
| D-9 | `TestNormalizeForm`, `TestIsInitial` and `TestSpellingSimilarity` move to `core/valuecompare/words_test.go` (new), in external `package valuecompare_test`, with the same cases; calls are qualified `valuecompare.` and the floor arguments read `matchregistry.Default().Value.Words`. `TestReconcileNamesInvariants` calls `valuecompare.NormalizeForm` | A test lives with the code it tests | — |
| D-10 | **`core/match` keeps no second copy of a registry number.** `core/match/registry.go` declares one `var value = matchregistry.Default().Value`, and every knob below reads it: `SpacedParticles`, `LetterFolds` read `value.Words.Particles` / `value.Words.Folds`; `DefaultWordRules.FuzzyFloor` / `ShortVariantMinLength` read `value.Words.FuzzyFloor` / `value.Words.ShortVariantMinLength`; `DefaultText.Partial` reads `value.Text.Partial`; `DefaultIntegers.Tolerance` reads `value.Integers.Tolerance`. `SpacedParticles` and `LetterFolds` share their backing slice and map with `value`, which only `match` reads. `core/match/words.go` `foldOnly` becomes a `var` reading `value.Words.FoldCredit`. `TestLetterFoldsComeFromTheMap` still mutates `LetterFolds`, which is now `match`'s own snapshot, so no other package sees it. Every value is today's, so no score changes | "Between the two, the old code may call the new layer"; two places for one number would drift before S11-05 deletes `match`'s comparers | Leaving `match`'s literals — because a knob tuned in the registry would not reach today's Promote |
| D-11 | **Value-section knobs in this PR** (inventory keys, start values from the inventory): `value.Words.FuzzyFloor` 0.8, `value.Words.ShortVariantMinLength` 3, `value.Words.FoldCredit` 0.95, `value.Words.Particles` `o`, `mc`, `mac`, `st`, `d`, `fitz`, `value.Words.Folds` today's `LetterFolds` map, `value.Text.Partial` 0.7, `value.Integers.Tolerance` 0. Go shape: `Rules{Words WordRules; Text TextRules; Integers IntegerRules}`, `WordRules{FuzzyFloor float64; ShortVariantMinLength int; FoldCredit float64; Particles []string; Folds map[rune]string}`, `TextRules{Partial float64}`, `IntegerRules{Tolerance int64}` | Only knobs a comparer in this PR reads (text reads no initial, alias or pairing-limit knob: D-7 aligns in order with no search bound) | Adding every `value.Words` row now — because `InitialCredit`, `AliasCredit` and `MaxExactPairing` would be fields nothing reads until S11-03 |
| D-12 | **Term and integer.** Term: the key as given, compared byte-equal, with no normalization (keys are catalog slugs); equal → `Same` 1 `exact`; different → `Different` 0 `no_match`; an empty key → `Unknown` `empty`. Integer: equal → `Same` 1 `exact`; gap ≤ `Tolerance` → `Near`, similarity `falloff(gap, Tolerance)` = 1 − gap / (`Tolerance` + 1), cause `gap`, where `falloff(gap, tolerance float64) float64` (new, unexported, in `valuecompare.go`) is the shared fall-off curve every type with a distance uses (design option C; S11-04 reuses it or records why not); beyond → `Different` 0 `beyond_tolerance`. Integer results set `Detail.Distance` to the gap and `Detail.Unit` to `UnitCount`. `Key`: term key; integer `fmt.Sprint` (no base literal); both the same at `Same` and `Equivalent` | The relation table's term and integer rows; the integer formula is today's `IntegerComparer` | — |
| D-13 | **Text `Key`.** `Same` level: the normalized words joined by one space. `Equivalent` level: the folded words joined by one space. An empty value has no key | D-7's steps (2) and (3) compare exactly these sequences, so key equality and the relation agree by construction (AC-2.5) | Sorted words — because order matters (D-7) |
| D-14 | Typed keys (`SubjectType`, `PropertyKey`) are not declared in this PR. The S11-02 scope entry's sentence about them moves to S11-10 | No section here has keyed entries; an unused type is dead code | Declaring them now — because nothing would use them for eight PRs |
| D-15 | **The guard** lives in `core/matchregistry/guard_test.go` (new): `TestLayerPackages` with one table `layerPackages` of `{dir string; numbers bool}` rows: `core/valuecompare` (numbers true), `core/matchregistry`, `core/datevalue`, `core/namevalue`, `core/namevalue/namevaluetest`, `core/propertymeta`, `core/vocabkey` (numbers false). Per row, subtest `<dir>/pure`: walk the package's `Imports`, `TestImports` and `XTestImports` with a copy of `build.Default` whose `BuildTags` is empty. The repo root is `../..` from the test's directory. An import under `github.com/mendahu/provenencia/` maps to a directory by trimming that prefix and calling `ctx.ImportDir(filepath.Join(root, rest), 0)` (never `ctx.Import`, which shells out in module mode); recurse into it through a visited set (the walk meets cycles through test imports, such as `valuecompare`'s tests importing `matchregistry`); other imports are not followed; fail on any path under `github.com/mendahu/provenencia/core/database`, printing the chain, and fail when the package has `IgnoredGoFiles` or no `GoFiles`. Subtest `<dir>/numbers` (rows with numbers true): parse every non-test `.go` file with `go/parser`; fail on each `*ast.BasicLit` of kind `IMAG`, and of kind `INT` or `FLOAT` whose value (`strconv.ParseFloat` after removing `_`; an `INT` in hex or octal through `strconv.ParseInt(lit, 0, 64)`) is not 0 or 1, so `0.0` and `1.0` pass; print `file:line literal`. No allow-list. `IgnoredGoFiles` is any file the context leaves out, by tag or by a GOOS or GOARCH file-name suffix; none is allowed | Working rule "One guard, not one per package"; the registry owns the numbers, so it is the guard's home; `go/build` needs no subprocess. A fact that is not a tunable (a calendar fact) belongs in its pure value package, which the numbers check does not scan | Shelling out to `go list` — because it is slower and needs the toolchain on `PATH`; an allow-list comment — because each exception would erode the rule |
| D-16 | **The inventory guard** is `TestKnobInventoryMatchesRegistry` in `core/matchregistry/inventory_test.go` (new). It reads `../../docs/matching-model.md`, takes the rows between `## Knob inventory` and the next `## ` heading, and reads each row's first cell inside backticks. It walks `Default()` with `reflect`: the top-level field name lowercased (`Value` → `value`), then struct field names as written; a map whose element is a struct or a map is walked by its keys in sorted order; any other field is a leaf whose path is a key. A package-level `notYetBuilt` list holds prefixes of rows whose fields land later: `value.Words.InitialCredit`, `value.Words.AliasCredit`, `value.Words.MaxExactPairing`, `value.Names.`, `value.Dates.`, `subject.`, `segment.`, `alignment.`, `evidence.`. It fails when a field has no row, when a row is neither a field nor under a prefix, when a prefix matches a field, and when a prefix matches no row | The inventory's own rule "every registry field has a row, and every row is a field", made true one section at a time; each layer PR deletes its prefixes | Checking only the fields that exist — because a row could then name a field that never comes; parsing Go struct tags for keys — because the dotted path already follows from the field names |
| D-17 | **The skill** is `.agents/skills/add-match-rule/SKILL.md` (new) with `.cursor/rules/match-rules.mdc` (new, globs `core/valuecompare/**,core/matchregistry/**,docs/matching-model.md`). `plan-pr`'s "Touches" table gains the row "Matching layer, registry knob, relation or cause code → `add-match-rule`". Its sections: when it applies; adding a knob (field and doc comment in the layer's `Rules` group, value and source comment in `matchregistry/<section>.go`, a row in the inventory, the `notYetBuilt` prefix deleted when its group lands); adding a comparer (an entry in `comparers`, a sample row and a generator in the contract test, the new type in T-3's expected `ComparerTypes()` list); adding a relation, cause or unit code (the constant, the cause list in `comparison-layers.md` §Detail); a type with a distance (similarity through `falloff`, or a written reason why not); adding a layer package (a `layerPackages` row); numbers (only in `matchregistry`; a fact goes in its pure value package); the check command; "do not" (a second copy of a word helper, a literal in a layer, a catalog read in a layer, a layer package importing `core/autoreconcile`) | Working rule "One matching skill, grown with the layers"; S11-10, S11-14 and S11-16 extend it | One skill per layer — because the rule says one |
| D-18 | Commits, each passing the full suite: (1) `valuecompare` types and `matchregistry` skeleton; (2) the word helpers move and `match` reads the registry; (3) term and integer; (4) text; (5) contract tests; (6) the guard; (7) the inventory guard; (8) the skill; (9) docs | One concern per commit | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/valuecompare` *(new)* | package | `core/match` `Comparer`, `ComparerFor` | `core/match` imports the impure `core/autoreconcile` and mixes neutral terms and frequency into comparison; the layer design replaces it, and S11-05 deletes its comparers |
| `valuecompare.Relation`, `Cause`, `Side`, `PartKind`, `Result`, `Detail`, `Part`, `PartMatch` *(new)* | types | `graphalign.Reason`, `match` `(float64, bool)` returns | No relation vocabulary exists; `graphalign.Reason` is a workflow reason code, not a value relation |
| `valuecompare.Value` *(new)* | type | `core/match/match.go` `match.Value` | `match.Value` carries reconciler fields (`Support`, `Reason`, `Denied`) and `Neutral`, which layer 1 must not read |
| `valuecompare.Rules`, `WordRules`, `TextRules`, `IntegerRules` *(new)* | types | `match.WordRules`, `TextComparer`, `IntegerComparer` settings | Pointer settings with nil-means-default; the registry passes whole values instead (rule 5) |
| `valuecompare.Compare`, `Key` *(new)*; `falloff` *(new, unexported)* | funcs | `match.ComparerFor` | See above |
| `core/valuecompare/words.go` *(new)* | file | `core/autoreconcile/spelling.go`, `core/match/words.go` | The four `autoreconcile` helpers and `match`'s tokenizer and folder **move** here (D-8); `match/words.go` keeps only its any-order pairing, aliases and word similarity until S11-05 deletes the file; text's in-order alignment is new (D-7) |
| `core/valuecompare/text.go`, `term.go`, `integer.go` *(new)* | files | `core/match/compare.go` | Build, then migrate: S11-05 deletes the old comparers |
| `core/matchregistry` *(new)*: `registry.go`, `value.go` | package | `core/match/registry.go` | The spike's one registry; `match/registry.go` reads it (D-10) and is deleted in S11-13 |
| `matchregistry.Registry`, `Default` *(new)* | type, func | `match.DefaultWordRules` and siblings | One `Default()` per the design |
| `TestLayerPackages` *(new)* | test | `core/writes/begin_test.go` `TestBeginOnlyInDatabaseAndWrites` | Same walk shape; different question |
| `TestKnobInventoryMatchesRegistry` *(new)* | test | `core/match/registry_test.go` `TestRegistrySetsEverySetting` | Walks a struct the same way, but checks a doc table |
| `TestComparerContract` *(new)* | test | `core/match/names_test.go` generated invariants | Names-only today; the contract covers every comparer |
| `add-match-rule` skill, `match-rules.mdc` *(new)* | skill, rule | `add-seeded-vocabulary`, `name-values.mdc` | No matching skill exists |

Searches run for this table: `rg -n 'type Relation|Relation string'` (none outside proto), `rg -n 'func Key\(' core` (none), `rg -n 'Cause' core --type go` (`graphalign` has none), `rg -n 'Default\(\)' core/match core/graphalign` (none).

**Builds on.** S11-22's pure value types and `propertymeta`'s value-type constants (the dispatch keys); S11-23's `vocabkey` (listed in the guard); the inventory S11-01 wrote; the repo-walk test shape of `core/writes`.
**Deletes.** core/autoreconcile/spelling.go (whole file); `NormalizeForm` from `core/autoreconcile/autoreconcile.go`; `IsInitial` from `core/autoreconcile/names.go`; the three moved tests from their `autoreconcile` files; `match`'s literal values for the seven knobs D-10 names; `splitWords`, `rawTokens`, `spacedParticle`, `periodInitial`, `foldWord` from `core/match/words.go`. `core/match/compare.go`'s comparers and `core/match/words.go` become dead once S11-05 moves their callers, and S11-05 deletes them.
**Next adds.** S11-03 adds `comparers[propertymeta.ValueTypeName]`, a `Names` group and the three name-only `Words` fields, extends `words.go`'s word similarity with initials and aliases, and deletes its `notYetBuilt` prefixes; catalog data (alias tables, stored patterns) must not become `Rules` fields the inventory guard would demand rows for, and S11-03's plan decides how it is passed. S11-04 adds the date comparer and `value.Dates.`. S11-05 moves `match`'s callers and deletes `compare.go` and `words.go`. S11-06, S11-07 call `Key`. S11-09 adds `core/propertycompare` as a `layerPackages` row with numbers true. S11-10 adds `Registry.Subject` and the typed keys.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` *(new)* | `valuecompare.go`: package doc ("Layer 1 of matching: how two values of one value type relate. Pure: no I/O, no catalog; rules come in as an argument from `core/matchregistry`."), the types of D-3 – D-5, `Compare`, `Key`, `comparer`, `comparers`, `falloff` (D-12). `export_test.go` (new): `ComparerTypes` (D-3). `rules.go`: `Rules` and its groups (D-11), one doc comment per field saying what it changes, no values. `words.go`: the moved helpers (D-8): `NormalizeForm`, `IsInitial`, `EditDistance`, `SpellingSimilarity`, `SplitWords`, `FoldWord`, unexported `rawTokens`, `spacedParticle`, `periodInitial`; new unexported `wordSimilarity(x, y string, w WordRules) (float64, Relation, Cause)` (D-7's word table) and `alignWords` (D-7's in-order alignment). `text.go`, `term.go`, `integer.go`: one `comparer` each (D-7, D-12, D-13) |
| `core/matchregistry` *(new)* | `registry.go`: package doc ("Every matching tunable, one section per layer group. Layers take their section as an argument; `Default()` is the shipped set."), `Registry`, `Default`. `value.go`: `func defaultValue() valuecompare.Rules`, each value with a comment naming where it came from (`autoreconcile.SpellingFloor`, `autoreconcile.SpellingShortMin`, `match` `foldOnly`, `match.SpacedParticles`, `match.LetterFolds`, `DefaultText.Partial`, `DefaultIntegers.Tolerance`) |
| `core/autoreconcile` | Delete `spelling.go`, `NormalizeForm`, `IsInitial`; call `valuecompare.*`; `var spelling` in `names.go` (D-8) |
| `core/match` | `words.go`, `frequency.go` call `valuecompare.NormalizeForm`, `IsInitial`, `SpellingSimilarity`; `registry.go` and `foldOnly` read the registry (D-10) |

```go
type Result struct {
	Relation   Relation
	Similarity float64 // 0…1; 0 for Unknown
}

type Detail struct {
	Parts      []PartMatch // what paired with what, in A's order
	OnlyA      []Part      // parts on one side only
	OnlyB      []Part
	Broader    Side        // for Contains: the less specific side
	Distance   float64     // for curve types: the raw gap …
	Unit       Unit        // … and its unit (UnitCount)
	Cause      Cause       // the rule that decided the relation
	RefA, RefB []byte      // the callers' value refs, echoed
}

type Part struct {
	Kind PartKind // word
	Text string   // the comparison token
}

type PartMatch struct {
	A, B       Part
	Relation   Relation
	Similarity float64
	Cause      Cause
}
```

Errors: none. Every input has a result; nothing returns `error`.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Layer 1 types and the registry skeleton**
Files: `core/valuecompare/valuecompare.go` (new), `core/valuecompare/rules.go` (new), `core/matchregistry/registry.go` (new), `core/matchregistry/value.go` (new), `core/matchregistry/registry_test.go` (new).
Do: D-1 – D-5, D-11. `comparers` is an empty map in this step. Write `TestDefaultIsFresh` (T-11).
Check: `go test ./core/valuecompare/... ./core/matchregistry/...` passes with no tags; `go vet ./core/valuecompare/... ./core/matchregistry/...` is clean.

**STEP-2 — Move the word helpers**
Files: `core/valuecompare/words.go` (new), `core/valuecompare/words_test.go` (new), `core/autoreconcile/spelling.go` (deleted), `core/autoreconcile/autoreconcile.go`, `core/autoreconcile/names.go`, `core/autoreconcile/autoreconcile_test.go`, `core/autoreconcile/names_test.go`, `core/autoreconcile/modules_test.go`, `core/match/words.go`, `core/match/names.go`, `core/match/frequency.go`, `core/match/registry.go`.
Do: D-8, D-9, D-10. Move, do not copy. Rewrite `EditDistance` without the literal 2 and keep `TestSpellingSimilarity`'s cases green.
Check: `rg -n 'autoreconcile\.(NormalizeForm|IsInitial|SpellingSimilarity|EditDistance|SpellingFloor|SpellingShortMin)' core api` prints nothing; `rg -n 'func (splitWords|foldWord|rawTokens|spacedParticle|periodInitial)' core/match` prints nothing; `rg -n 'SpellingFloor|SpellingShortMin' core` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/... ./core/match/... ./core/database/autoreconciler/... ./core/database/promotealign/...` passes with no expected value changed.

**STEP-3 — Term and integer comparers**
Files: `core/valuecompare/term.go` (new), `core/valuecompare/integer.go` (new), `core/valuecompare/term_test.go` (new), `core/valuecompare/integer_test.go` (new), `core/valuecompare/valuecompare_test.go` (new).
Do: D-12; register both in `comparers`; T-1, T-2, T-3.
Check: `go test ./core/valuecompare/...` passes with no tags.

**STEP-4 — Text comparer**
Files: `core/valuecompare/text.go` (new), `core/valuecompare/words.go`, `core/valuecompare/text_test.go` (new).
Do: D-7, D-13; register it; T-4, T-5, T-6.
Check: `go test ./core/valuecompare/...` passes; AC-1.2 – AC-1.8 rows are cases in `TestTextCompare`.

**STEP-5 — Contract tests**
Files: `core/valuecompare/contract_test.go` (new).
Do: T-7, T-8. The sample table is keyed by value type; the test loops over all seven `propertymeta.ValueType*` constants and fails when a type in `valuecompare.ComparerTypes()` has no row (AC-2.6). Generated pairs use `math/rand` with a fixed seed per type, 500 pairs each: text from a word pool with case, accent, punctuation, particle and one-letter variants; terms from five keys; integers from 0 to 9 with tolerance from 0 to 3.
Check: `go test -run 'TestComparerContract|TestKeyContract' -count=1 ./core/valuecompare/` passes.

**STEP-6 — The layer-package guard**
Files: `core/matchregistry/guard_test.go` (new).
Do: D-15; T-9.
Check: `go test -run TestLayerPackages ./core/matchregistry/` passes; temporarily add `var x = 2` to `core/valuecompare/text.go` (new) and see `TestLayerPackages/core/valuecompare/numbers` fail with the file and line; temporarily add `_ "github.com/mendahu/provenencia/core/database/datevalues"` to `core/propertymeta/propertymeta.go` and see the `pure` subtest fail; revert both.

**STEP-7 — The inventory guard**
Files: `core/matchregistry/inventory_test.go` (new).
Do: D-16; T-10.
Check: `go test -run TestKnobInventoryMatchesRegistry ./core/matchregistry/` passes; temporarily delete the `value.Text.Partial` row from `docs/matching-model.md` and see it fail naming that key; revert.

**STEP-8 — The matching skill**
Files: `.agents/skills/add-match-rule/SKILL.md` (new), `.cursor/rules/match-rules.mdc` (new), `.agents/skills/plan-pr/SKILL.md`, `.agents/skills/add-name-value/SKILL.md`, `.agents/skills/add-date-value/SKILL.md`.
Do: D-17. In `add-name-value` and `add-date-value`, beside the line that calls the package pure, add (the guard file is (new) in STEP-6): "`TestLayerPackages` in `core/matchregistry/guard_test.go` keeps this package pure; a new pure value package adds a row." The skill's check command is `go test ./core/valuecompare/... ./core/matchregistry/...` (no tags).
Check: every repo path the skill names in backticks exists (`ls` each one); `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-02-value-compare-core.md` still reports no errors.

**STEP-9 — Full suite**
Files: none.
Do: run the gate on the user's Mac, as `ship-pr` requires (`xcode-test` cannot pass elsewhere).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-11/deployment-plan.md`, `docs/deployment-plan/spike-11/completed.md`, this plan, `docs/matching-model.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/conclusion-reconciliation.md`, `docs/ideas/daitch-mokotoff.md`.
Do: per the Definition of done (the map's scope entries for S11-02, S11-03, S11-06 and S11-10 were updated with this plan), and: in `matching-model.md`'s inventory lead paragraph, replace "of the planned registry" with "of the registry, `core/matchregistry`" and add "`TestKnobInventoryMatchesRegistry` enforces it; rows whose layer has not landed are listed in its `notYetBuilt`." In `comparison-layers.md` §Detail's cause list, add `no_match`, `order`, `unsupported`. In `conclusion-reconciliation.md` L166, name `valuecompare.SpellingSimilarity`. In `daitch-mokotoff.md` L5, say `core/autoreconcile/spelling.go`'s helpers moved to value compare in S11-02 and `wordSimilarity` / `DefaultWordRules` follow in S11-03 / S11-05; in L19, link `core/valuecompare/words.go` (new). In `comparison-layers.md` L410, add "(moved to `core/valuecompare` and `core/matchregistry` in S11-02)" after the `autoreconcile/spelling.go` mention.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-02-value-compare-core.md` reports no errors; `rg -n 'autoreconcile/spelling.go' docs --glob '!docs/deployment-plan/archive/**' --glob '!docs/deployment-plan/spike-10/**'` prints only `docs/deployment-plan/spike-11/comparison-layers.md` L410 (now marked as moved) and files under `docs/deployment-plan/spike-11/prs/`.

Commits: one per behavior (D-18). The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Move the word helpers into value compare."). No WIP commits.

## 8. Tests

All in `package valuecompare_test` or `package matchregistry` with no build tags, table-driven (`.cursor/rules/go-tests.mdc`). Expected values are written as the arithmetic.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/term_test.go` `TestTermCompare` *(new)* | The three AC-1.9 pairs → their relation, similarity, cause; `Key` of `male` at both levels is `male` | AC-1.9, D-12 |
| T-2 | Go | `core/valuecompare/integer_test.go` `TestIntegerCompare` *(new)* | The four AC-1.10 pairs, tolerance set on a copy of `Default().Value` → relation, similarity, cause, `Distance`, `Unit` | AC-1.10, D-12 |
| T-3 | Go | `core/valuecompare/valuecompare_test.go` `TestCompareUnsupported` *(new)* | Name, date, subject, depiction values; a text against a term → `Unknown`, `unsupported`; `ComparerTypes()` is exactly `integer`, `term`, `text`; refs `r1` / `r2` echoed on a `Same`, a `Different` and an `Unknown` | AC-1.11, AC-1.12, D-3 |
| T-4 | Go | `core/valuecompare/text_test.go` `TestTextCompare` *(new)* | Every pair in AC-1.2 – AC-1.8, plus `O Brien` / `O'Brien` → `Same` `case`; `Sainte-Anne` / `Sainte Anne` → `Same` `case`; `York` / `York` → `Same` `exact` → relation, similarity (±1e-9), cause, `Broader`, `OnlyA`, `OnlyB`, `Parts` | AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, D-7 |
| T-5 | Go | `core/valuecompare/text_test.go` `TestTextLadder` *(new)* | `York` / `York` > `Montréal` / `Montreal` > `Provenance` / `Provenanced` > `York` / `York, Upper Canada` > `York` / `York, Upper Canada, North America` (0.7 × 2 / 6) by similarity | D-7 |
| T-6 | Go | `core/valuecompare/text_test.go` `TestTextKey` *(new)* | `Montréal` / `Montreal`: `Same` keys differ, `Equivalent` keys equal (`montreal`); `Upper Canada, York` / `York, Upper Canada`: keys differ at both levels; `!!!` → `"", false`; level `Near` → `"", false` | AC-2.5, D-13 |
| T-7 | Go | `core/valuecompare/contract_test.go` `TestComparerContract` *(new)* | Per supported type, sample table plus 500 seeded pairs → the AC-2.1 bounds, AC-2.2 swap, AC-2.3 self, AC-2.4 shuffled `Particles`; a supported type with no sample row fails | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.6 |
| T-8 | Go | `core/valuecompare/contract_test.go` `TestKeyContract` *(new)* | The same pairs → AC-2.5 both directions at both levels | AC-2.5 |
| T-9 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` *(new)* | The seven `layerPackages` rows → `pure` passes for each; `numbers` passes for `core/valuecompare` | AC-3.1, AC-3.2, D-15 |
| T-10 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` *(new)* | `Default()` against `docs/matching-model.md` → the seven fields of D-11 each have a row; every other row is under a `notYetBuilt` prefix; every prefix matches a row and no field | AC-3.3, AC-3.4, AC-3.5, D-16 |
| T-11 | Go | `core/matchregistry/registry_test.go` `TestDefaultIsFresh` *(new)* | Change the first `Default()`'s `Particles[0]` and add a `Folds` entry → a second `Default()` is unchanged | AC-3.6, D-2 |
| T-12 | Go | `core/valuecompare/words_test.go` *(new)* `TestNormalizeForm`, `TestIsInitial`, `TestSpellingSimilarity` (moved) | Today's cases, unchanged → today's results | AC-4.1, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/autoreconcile` (STEP-2 moves its helpers), `core/match` `TestRegistrySetsEverySetting` and the names ladders (STEP-2 repoints its knobs), `core/database/autoreconciler` (calls the reconciler's name module).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md inventory paragraph, conclusion-reconciliation.md L166, comparison-layers.md cause list, ideas/daitch-mokotoff.md link (STEP-10)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No app change. The reviewer runs, from the repo root:

1. `go test ./core/valuecompare/... ./core/matchregistry/...` with no tags: passes (AC-1.1).
2. `go list -deps ./core/valuecompare ./core/matchregistry | grep 'provenencia/core/database'`: no output.
3. `rg -n 'SpellingFloor|SpellingShortMin|autoreconcile\.(NormalizeForm|IsInitial|SpellingSimilarity|EditDistance)' core api`: no output.
4. Add `var x = 2` to `core/valuecompare/term.go` (new) and run `go test -run TestLayerPackages ./core/matchregistry/`: fails naming `term.go` and the line. Revert.
5. Open a project with a Source and use Promote: suggestions and bands read as before (AC-4.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Rewriting `EditDistance` without the literal 2 changes a distance | `TestSpellingSimilarity` and the `match` names ladders run unchanged in STEP-2 |
| Repointing `match`'s knobs (D-10) changes a score | Every value is today's; the `match`, `promotealign` and `autoreconciler` suites run unchanged in STEP-2 |
| `go/build` resolves an import differently from `go list` | STEP-6's two deliberate failures prove both subtests catch what they claim |
| The in-order alignment picks a different pairing on a tie, so swapping sides changes the relation | T-7's swap invariant over 500 seeded pairs per type |

**Stop and report if:**

- A row in §4 is no longer true at your base (for example core/autoreconcile/spelling.go is gone or `match` no longer calls `autoreconcile.SpellingSimilarity`).
- STEP-2 changes any expected value in an existing test, or needs `core/autoreconcile` to import anything new besides `core/valuecompare` and `core/matchregistry`.
- `matchregistry` importing `valuecompare` (D-1) makes an import cycle.
- A text case in AC-1.2 – AC-1.8 cannot hold under D-7's rule, or T-7's swap invariant fails on a generated pair that D-7's rule cannot settle without a tie-break the plan does not give.
- `TestLayerPackages/<pkg>/pure` fails for `core/datevalue`, `core/namevalue`, `core/propertymeta` or `core/vocabkey` at the base: that is a regression from S11-22 or S11-23, outside this PR.
- Any `numbers` failure in `core/valuecompare` you would fix by an exception rather than a registry field.

## 12. Open questions

None.

## 13. PR

**Title:** S11-02 — Value compare core and match registry
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (test names, the deliberate guard failures from STEP-6 and STEP-7); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-02-value-compare-core.md`.

## Landed

Differences from the plan:

- `TestComparerContract` and `TestKeyContract` run the sample table and 500 generated pairs under four seeds (the integer tolerance is the seed index, 0 to 3), not one seed.
- `ComparerTypes()` is asserted in `TestCompareUnsupported` as T-3 says; `TestCompareEchoesRefs` (new) holds the AC-1.12 ref echo, which T-3 lists under the same test.
- For `no_match` and `order`, `Detail.OnlyA` / `OnlyB` list the unpaired words (D-7 says only that `Parts` is empty for `no_match`).
- `core/matchregistry/value.go` names the old sources as "the reconciler's old spelling floor" and "old short-word minimum" instead of the deleted constant names, so STEP-2's `rg SpellingFloor|SpellingShortMin core` check prints nothing.
- `daitch-mokotoff.md` L5 says "the spelling helpers that lived in `core/autoreconcile`" instead of naming `spelling.go`, so STEP-10's `rg` check prints only the allowed files.
- The STEP-6 and STEP-7 deliberate failures needed `-tags fts5`: with a `core/database` import in `core/propertymeta`, the test binary panics at init without it, so the `pure` failure shows with the tag.
- Four current-state and stop-condition lines drop the backticks around the deleted `core/autoreconcile/spelling.go` so `check-plan-docs` passes, as the S11-22 plan did.
