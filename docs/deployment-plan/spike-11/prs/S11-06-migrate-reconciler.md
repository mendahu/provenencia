# S11-06 — Migrate the reconciler onto value compare

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the reconciler's grouping rules change in Go; pages render the cache as they do today)
**Board:** —
**Depends on:** S11-05 (landed [#451](https://github.com/mendahu/provenencia/pull/451))
**Blocks:** S11-08
**Branch:** `s11-06-migrate-reconciler`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [relations](../comparison-layers.md#relations), [reuse across the app](../comparison-layers.md#reuse-across-the-app), [decided on 2026-10-17](../comparison-layers.md#decided-on-2026-10-17); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5, §7, §7.1, §7.2, §10; [S11-02 plan](S11-02-value-compare-core.md) D-8; [S11-04 plan](S11-04-value-compare-dates.md); [S11-05 plan](S11-05-migrate-matching.md) D-2, D-3, D-6, D-10; working rules "Build, then migrate", "High tolerance for churn", "The reconciler keeps its own decisions", "One match registry"
**Planned:** 2026-10-17 against `main` @ `9c3b6f1` (S11-05 landed; first drafted at `75ed63c`)
**Review:** [`S11-06-migrate-reconciler.review.md`](S11-06-migrate-reconciler.review.md) — Fits with changes / Ready with changes → fixed, 1 round; refreshed after S11-05 landed

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

The auto-reconciler decides "the same value" with value compare instead of its own key rules. Under the grouping policy the user chose on 2026-10-17, values group when value compare says `Same` and fold when it says `Contains`; `Equivalent` (an accent fold, an alias) never groups. Researchers see text that differs only by punctuation or spacing merge (`Sainte-Anne` / `Sainte Anne`), `Mc Donald` and `McDonald` read as one surname, `ABT 1985` read as 1984 – 1986 like matching reads it, and place membership spans pick the later start and earlier end by day, not by year. Promote's pins use the same test, so a pin and a Person page still agree.

**Why now.** S11-04 left `DateBounds`, `WindowOfDate`, `PeriodWindow` and `LinkMembership` in the reconciler, and S11-05 moved matching onto value compare; until the reconciler moves, two definitions of "the same value" disagree (comparison-layers.md "Found in passing"). S11-08 moves the Swift fake onto the rule this PR settles.

**What stays.** The pipeline (`pipeline.go`: rank, admit, deny, group, majority, confidence), its reasons and outcomes, `MinMajoritySupport`, and the name module's per-part-type units, assembly and spelling-only outvoting. Terms, subjects and depictions keep their catalog-id identity. `SortKey` for text and names. No Swift, wire or schema change; the cache rebuilds on open through `CacheVersion`.

## 2. User stories

**US-1.** As a researcher, a Person or Place shows one value for text that differs only by case, punctuation or spacing, and still shows two values for an accent or a nickname.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Text candidates `Sainte-Anne` (obs 1), `Sainte Anne` (obs 2), single cardinality | One value, members 1 and 2, state `merged`, displayed `Sainte-Anne` |
| AC-1.2 | Text `Montréal` (1), `Montreal` (2) | Two values, state `mixed` (`Equivalent` does not group) |
| AC-1.3 | Text `York` (1), `York, Upper Canada` (2) | Two values, state `mixed` (text never folds, D-3) |
| AC-1.4 | Text `Upper Canada` (1), `Upper  Canada` (2), `U.C.` (3) | `[[1 2] [3]]`, state `merged`: 1 and 2 are one value with two Sources, which outvotes `U.C.` |

**US-2.** As a researcher, a name's parts group as words value compare reads, and fold only into a fuller name that repeats them or spells out their initials.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Names `given=J.|surname=Robins` (1), `given=James|surname=Robins` (2) | One value `given=James|surname=Robins`; obs 1 `folded` (unchanged behavior, now through value compare) |
| AC-2.2 | `given=James|surname=Mc Donald` (1), `given=James|surname=McDonald` (2) | One value, members 1 and 2, displayed obs 1's parts (particle `mc` rejoins) |
| AC-2.3 | `given=Jose|surname=Robins` (1), `given=José|given=Kenneth|surname=Robins` (2) | One name `given=Jose|given=José|given=Kenneth|surname=Robins`; obs 1 `kept`, not `folded` (an accent-folded pair never folds) |
| AC-2.4 | `Compatible` under rules whose alias index holds `jim` / `james` for every role: `given=Jim|surname=Robins` against `given=James|given=Kenneth|surname=Robins` | `false`, while `valuecompare.Compare` on the whole names is `Contains` |
| AC-2.6 | `given=James|surname=McDonald` from Sources a and b (1, 2), `given=James|surname=Mc Donnald` from Source c (3) | Obs 3 `outvoted`: the misspelling is outvoted whatever its spacing |
| AC-2.5 | `given=Kenneth James|surname=Robins` (1), `given=James|given=Kenneth|surname=Robins` (2) | One value, members 1 and 2, displayed `given=Kenneth James|surname=Robins` (word order within a part type no longer splits a value) |

**US-3.** As a researcher, an approximate date reads as the range matching reads, and dates on the same day are one value.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Dates `ABT 1985` (1), `1986` (2), single | One value `1986`, obs 1 `folded`, state `merged` |
| AC-3.2 | `fold(ABT MAY 1985, JUN 1985)` | `true` (ABT May is April – June) |
| AC-3.3 | `14 MAY 1985` (1) and `14 MAY 1985 15:00` (2) | One value, state `merged` (clock times are not compared) |
| AC-3.4 | A range with only an end year `– 1900` (1), `1885` (2) | One value `1885`, obs 1 `folded` (today obs 1 is `no_evidence`) |
| AC-3.5 | The cache's `date_lo` / `date_hi` for `ABT 1817` | The days of 1817 as written (not widened): equal to `1817`'s |

**US-4.** As a researcher, a place's membership span picks the later start and the earlier end by day.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Undated `part_of` link; the part starts `1 MAR 1850`, the whole starts `1 SEP 1850` | Membership start `1 SEP 1850`, either way round (today `1 MAR 1850` or `1 SEP 1850` by argument order) |
| AC-4.2 | The part ends `31 OCT 1867`, the whole ends `1 JUL 1867` | Membership end `1 JUL 1867`, either way round |
| AC-4.3 | Today's place-chain cases (`TestPlaceParentsAtDateFromPeriods`, `TestIrelandUKDatedMembershipDropsAfter1922`, the moved `TestPeriodWindowAndLinkMembership`) | Pass unchanged |

**US-5.** As a contributor, the reconciler takes value compare's rules as an argument and holds no comparison rule of its own.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `rg -n 'matchregistry|DateBounds|WindowOfDate|PeriodWindow|LinkMembership|subsumes|partFits|textKey|dateKey|windowOf|moreSpecificThan' core/autoreconcile --glob '!*_test.go'` | Nothing |
| AC-5.2 | `rg -n 'autoreconcile\.(DateBounds|WindowOfDate|PeriodWindow|LinkMembership)' core api` | Nothing |
| AC-5.3 | `scripts/check-pr.sh` | `check-pr: PASS`; every expected value that moved is a row of §8 "Expected changes" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `Reconcile` and `Compatible` take `valuecompare.Rules`; `moduleFor` passes them to the modules (D-1) | STEP-3, AC-5.1 |
| The grouping policy: `Same` groups, `Contains` folds, `Equivalent` never does (D-2) | STEP-3, AC-1.2, AC-2.3, AC-2.4 |
| Text and integer keys are `valuecompare.Key` at `Same`; text never folds (D-3) | STEP-3, AC-1.1 – AC-1.4 |
| Name units key and fold through value compare; `subsumes` / `partFits` deleted (D-4) | STEP-3, AC-2.1 – AC-2.6 |
| Date units key and fold through value compare; the module's own window code deleted (D-5) | STEP-3, AC-3.1 – AC-3.4 |
| `valuecompare.DateWindow` *(new)* replaces `DateBounds` / `WindowOfDate` for the date sort key, `date_lo` / `date_hi`, and place chains (D-6) | STEP-1, STEP-2, STEP-3, AC-3.5 |
| `PeriodWindow` and `LinkMembership` move to `conclusionheaders` as `periodWindow` / `linkMembership`; `laterStart` / `earlierEnd` compare by day (D-7) | STEP-2, AC-4.1 – AC-4.3 |
| The cache passes `matchregistry.Default().Value`; `CacheVersion` 16 (D-8) | STEP-3 |
| Promote's `compatible` (`core/match/pairs.go`) and `exhibitCompatible` pass the rules they hold (D-9) | STEP-3 |
| `conclusion-reconciliation.md`, `comparison-layers.md`, `docs/ideas/international-names.md`, the `words.go` header, `add-match-rule` (D-11) | STEP-4 |

| Out | Home |
| --- | --- |
| `FakeStore` grouping names by parts | S11-08 |
| Search rescoring, search-index dedupe, term dedupe, alias tokens on `Key` | S11-07 |
| `SortKey` for text and names (an ordering, not a sameness test) | not planned |
| Term keys by `property_terms.key` instead of the catalog id | not planned (D-3) |
| Pins from claim evidence instead of `Compatible` | S11-19, S11-25 |
| `pairValues`' negatives on property compare | S11-09 |
| Calendar-aware date comparison (a Julian and a Gregorian `1 MAR 1700` key the same) | not planned: value compare reads civil components as Gregorian (`DateEnds` doc) |
| The name module's spelling-only outvote rewritten on relations | not planned (D-4) |
| Unexporting `valuecompare.IsInitial` once `partFits` is gone | not planned: no caller asks for it |
| Swift copy or UI for "folded" | not planned: reasons are unchanged |

## 4. Current state

Verified at `9c3b6f1`, after S11-05 landed.

| Area | Where | Today |
| --- | --- | --- |
| Entry points | `core/autoreconcile/autoreconcile.go` `Reconcile(valueType, candidates, concluded, cardinality)`, `Carries`, `SortKey`, `dateKey`, `optInt` | `SortKey` for dates reads `DateBounds`; `dateKey` keys a date on every structured field |
| Modules | `core/autoreconcile/modules.go` `module`, `moduleFor(valueType)`, `keyModule`, `textKey` (lower + trim), `termKey`, `idKey`, `Compatible(valueType, a, b)` | Text `Sainte-Anne` ≠ `Sainte Anne`; integers key by `strconv` |
| Names | `core/autoreconcile/names.go` `var spelling = matchregistry.Default().Value.Words` (its only registry import, S11-02 D-8), `nameModule`, `nameUnit`, `readName`, `split` (key = words joined in order), `fold` → `subsumes` / `partFits`, `outvotes`, `assemble`, `containsAll` | Words from `strings.Fields(NormalizeForm(part))`: particles do not rejoin, word order inside a type splits values |
| Dates | `core/autoreconcile/dates.go` `dateModule`, `dateWindow`, `dateUnit`, `windowOf`, `pointWindow`, `rangeWindow`, `spanOf`, `contains`, `moreSpecificThan`, `openEnds`, `DateBounds`, `WindowOfDate`, `PeriodWindow`, `LinkMembership`, `dateSortOffset` | ABT reads as its point; a one-sided range has no window (no evidence) |
| Pipeline | `core/autoreconcile/pipeline.go` `reconcile(m, …)` | Unchanged by this PR |
| Tests | `core/autoreconcile/dates_test.go` `TestDateFold` (`dateUnitOf`, `dateModule{}`), `TestDateBounds`, `TestDateReconcile`, `TestPeriodWindowAndLinkMembership`, `mustWindow`, `ptrDate`; `modules_test.go` `TestModules`, `TestModuleFor`, `TestCompatible`; `names_test.go` `TestReconcileNames` and kin; `autoreconcile_test.go`, `pipeline_test.go` (every `Reconcile(` call) | — |
| Cache | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` 15 with its history comment, `recomputeBatch` (`Reconcile`), `insertRow` (`DateBounds` → `date_lo` / `date_hi`, `SortKey`); `truth.go` `truthRows` (`Reconcile`) | — |
| Place chains | `core/database/conclusionheaders/placechains.go` `linkHolds` (`LinkMembership`), `queryWindow`, `chainQueryDate` (`WindowOfDate`), `AttachPlaceRelationships` (`PeriodWindow`, `LinkMembership`), `membershipDates`, `yearOf`, `laterStart`, `earlierEnd` (year only); `compose.go` three `PeriodWindow` calls, two `SortKey` calls | Two same-year starts keep the first argument |
| Place tests | `core/database/conclusionheaders/placechains_test.go` (package `conclusionheaders_test`, year-only fixture), `export_test.go` | No day-precision case |
| Value compare | `core/valuecompare/datewindow.go` `DateEnds`, `AboutEnds`, `EndsWindow`, `Window`; `valuecompare.go` `Compare`, `Key`; `name.go` `name.compare` (any-order pairing, `Detail.Parts`, `Broader`), `nameWords` (type `""` → `RoleUntyped` under `Western`); `text.go` `text.key`; `datewindow_test.go` `TestDateEnds` | No as-written window helper |
| Promote pair selection | `core/match/pairs.go` `compatible(a, b Value)` → `autoreconcile.Compatible(t, a.Auto(), b.Auto())`, called from `pairValues` (which takes `rules valuecompare.Rules`, S11-05 D-3) and from `refutedBy` | `refutedBy(v, others, info)` has no rules |
| Promote exhibits | `core/database/promotealign/exhibits.go` `exhibitCompatible(valueType, a, b)`, called in `pairExhibits`, which reads `cfg.Value` (S11-05 D-6, D-10) | — |
| Docs | `docs/conclusion-reconciliation.md` §7 table (text, integer, date, name rows), L144 (`Compatible`), §7.1, §7.2, §12; `docs/deployment-plan/spike-11/comparison-layers.md` L454, the duplicates rows for `autoreconcile` keys and `Compatible`, "Decided on 2026-10-17" reconciler bullet; map working rule "The reconciler keeps its own decisions"; `.agents/skills/add-match-rule/SKILL.md` "Dates compute similarity through `falloff`" bullet | Describe the old keys and "ABT as its point" |

**Precedent.** S11-05 (D-2, D-6): a caller converts its own value to `valuecompare.Value` in one adapter and passes the rules it holds; copy that. S10-52 (`447917a`, `CacheVersion` 15): a reconciler rule change bumps `CacheVersion` with one history line, and the rebuild-equals-upkeep tests cover it; copy that. The measured "Expected changes" come from a prototype of D-1 – D-10 (with the review's fixes) on `9c3b6f1`, with S11-05's rules at Promote's two call sites; every other Go test under `./core/... ./api/...` passed.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Rules are an argument.** New signatures: `Reconcile(rules valuecompare.Rules, valueType string, candidates []Candidate, concluded *Value, cardinality string) (Result, error)`, `Compatible(rules valuecompare.Rules, valueType string, a, b Value) bool`, `moduleFor(rules valuecompare.Rules, valueType string) module`. `nameModule` and `dateModule` become structs with a `rules valuecompare.Rules` field; `keyModule`'s `keyOf` closures capture the rules. `names.go` loses `var spelling` and its `matchregistry` import; `outvotes` reads `m.rules.Words.FuzzyFloor` and `m.rules.Words.ShortVariantMinLength`. The `module` interface is unchanged | Answers S11-02 D-8: the caller passes the rules, as every layer takes its section ("One match registry"); `autoreconcile` then imports no registry | Keeping `var spelling` — because Promote already holds the catalog's rules and a test could not vary them |
| D-2 | **The grouping policy (user, 2026-10-17).** Two values are one value when `valuecompare.Key(…, valuecompare.Same, rules)` is equal. A value folds into a fuller one when `valuecompare.Compare(from, into, rules)` is `Contains` with `Detail.Broader == valuecompare.SideA` (and, for names, D-4's pair rule). `Equivalent`, `Near` and `Overlaps` never group or fold. Majority outvoting is the module's existing rule | The user's answer, the recommendation on record ([decided 2026-10-17](../comparison-layers.md#decided-on-2026-10-17)); §8 lists `Montréal` and `Montreal` as two true toponyms | Grouping `Equivalent` — the user declined it |
| D-3 | **Keys per type.** One adapter, as S11-05 D-2 has for `match.Value`: unexported `func (v Value) compareValue(valueType string) valuecompare.Value` in `modules.go` sets `Type` = `valueType` and copies `Text`, `Integer` and `Date` (names go through D-4's unit adapter; `Term` is not copied, see below). Text and integer share one `keyOf`: `Key(v.compareValue(valueType), Same, rules)`. Text: no key (blank or punctuation only) is no evidence; text never folds (user, 2026-10-17). Term, subject, depiction: unchanged catalog-id keys (`termKey`, `idKey`) | `property_terms` is `UNIQUE (property_id, key, origin)`: a user term and a product term may share a key and are two terms; the id is the identity. Text fold was declined for remarks, toponyms and interests | Term keys from `TermKey` through `valuecompare` — because it would merge two terms of one key |
| D-4 | **Name units through value compare.** A unit is still one part type's parts (`nameUnit`). New unexported `func (u nameUnit) compareValue() valuecompare.Value`: `Type` name, `Name` = the unit's raw parts with `Type` `""` and `Idx` renumbered from 0, no `NamePattern`, so every word is `RoleUntyped` under `rules.Names.Western` and no part type weighs more than another (§7.2 "Format-agnostic"). `split` keys each unit by `Key(u.compareValue(), Same, rules)` and skips a unit with no key; a name with no keyed unit is no evidence. `fold(from, into)` is true when `Compare(from.compareValue(), into.compareValue(), rules)` is `Contains`, `Broader == SideA`, and every `Detail.Parts` entry is `Same` or has `Cause == CauseInitial`. `subsumes` and `partFits` are deleted. `readName` becomes a `nameModule` method and builds each part's words with `valuecompare.SplitWords(p.Value, particles)`, where `particles` is the set of `m.rules.Words.Particles` built once in `moduleFor` and held on `nameModule` (field `particles map[string]bool`), so `parts`, `words`, `outvotes` and `assemble`'s `containsAll` read the tokens `Key` and `Compare` read; `assemble`'s receiver is named so it can call `readName` (its loop variable `m` is renamed `member`). `split` builds `u.compareValue()` once and stores it on `nameUnit` (field `cmp valuecompare.Value`), and `fold` compares the stored values. `outvotes`' rule (per-word `SpellingSimilarity`, same word count) and `assemble`'s logic are unchanged | Value compare pairs initials and rejoins particles, which the policy wants; its alias and fold pairs are `Equivalent` and must not fold (AC-2.3, AC-2.4); a real role (family, first given) would make the result depend on the type name and drop `prefix` / `suffix` units, which `Western` ignores or treats as generation | Comparing whole names — because the pipeline votes per part type (§7.2); the catalog's stored pattern — same reason as roles; rewriting `outvotes` on `Near` — because name `Near` also covers different middle words (Kevin / Kenneth), which §7.2 says are never outvoted |
| D-5 | **Date units through value compare.** `dateModule.split` keys a date by `Key(Value{Type: date, Date: v.Date}, Same, rules)` (the `AboutEnds` window: ABT widened, clock times and calendars not compared) and keeps `v.compareValue(propertymeta.ValueTypeDate)` (D-3's adapter) as unit data, so `fold` compares the stored values: `Compare(from, into, rules)` `Contains` with `Broader == SideA`; `assemble` returns `Value{Date: &d}` with `d` a copy of the first settled unit's `Date`. Deleted: `dateWindow`, `dateUnit`, `windowOf`, `pointWindow`, `rangeWindow`, `spanOf`, `contains`, `moreSpecificThan`, `openEnds`, `dateKey`, `optInt` | One ABT reading everywhere (user, 2026-10-17); equal windows are one value by value compare's `Same` | Keeping `moreSpecificThan` — because value compare's `Same` already merges equal windows, and its `Contains` is strict |
| D-6 | **`valuecompare.DateWindow` *(new)*** in `core/valuecompare/datewindow.go`: `func DateWindow(d datevalue.Value) (Window, bool)` returns `EndsWindow(DateEnds(d))`, ok false when `DateEnds` is not ok. Doc: "the days d covers as written: DateEnds as a Window, with an ABT point not widened". `SortKey` for dates reads `w.Lo` (no `Lo` → no key); `insertRow` writes `date_lo` / `date_hi` from it; `conclusionheaders` uses it where it used `WindowOfDate`. `DateBounds` and `WindowOfDate` are deleted | Sorting, the stored date filter and place periods read a date as written; one helper replaces two reconciler functions with the same meaning | Widening ABT in `date_lo` / `date_hi` — because a filter "born in 1817" would then match `ABT 1816`; writing `EndsWindow(DateEnds(…))` at five call sites — because each must handle `ok` |
| D-7 | **Place periods live with place chains.** `PeriodWindow` and `LinkMembership` move to `core/database/conclusionheaders/placechains.go` as unexported `periodWindow` and `linkMembership`, same bodies on `valuecompare.DateWindow`. `laterStart(a, b)`: a nil `a` returns `b` and a nil `b` returns `a`, as today; otherwise each is read with `valuecompare.DateWindow` (the read `periodWindow` uses; a phrase-only date is unreadable, a nil `Lo` is open). It returns `b` when `b` is readable and either `a` is unreadable, or `b`'s `Lo` is set and `a`'s is nil, or both are set and `*b.Lo > *a.Lo`; otherwise `a` (a tie, two unreadable dates, or an unreadable `b` keep `a`). `earlierEnd` mirrors it on `Hi` with `*b.Hi < *a.Hi`. `yearOf` is deleted. `comparison-layers.md` L454 is corrected to say so | They answer "when does this place or link hold", not "how do two values relate"; their only callers are `conclusionheaders` | Moving them into `core/valuecompare` — because layer 1 makes no workflow decision ([rule 7](../comparison-layers.md#rules-for-every-layer)) |
| D-8 | **The cache reads the shipped registry.** `core/database/autoreconciler/autoreconciler.go` declares `var valueRules = matchregistry.Default().Value` with the doc "the value compare rules reconciling reads: the shipped registry, never a catalog's alias tables or name formats, so the cache changes only when CacheVersion does"; `recomputeBatch` and `truth.go` `truthRows` pass it. `CacheVersion` becomes 16 with the history line "16: values group by value compare's Same and fold by its Contains: text ignores punctuation and spacing, ABT is a range, name parts compare as words with particles (S11-06)." | No write to alias tables or name formats recomputes the cache; under D-2 and D-4 aliases and patterns cannot change a grouping, so `valuerules.Load` would add a catalog read for nothing | `valuerules.Load` per batch — because the cache would then depend on data its upkeep does not watch |
| D-9 | **Promote passes the rules it holds.** `core/match/pairs.go` `compatible(rules valuecompare.Rules, a, b Value)` calls `autoreconcile.Compatible(rules, t, a.Auto(), b.Auto())`; `pairValues` passes its `rules`; `refutedBy` gains a first parameter `rules valuecompare.Rules` and its caller in `pairValues` passes it. `exhibitCompatible(rules valuecompare.Rules, valueType string, a, b match.Value)`; `pairExhibits` passes `cfg.Value`. `match.Value.Auto` is unchanged | S11-05 gave both callers their catalog's rules; no second rules source | A package-level default in `match` — because S11-05 D-3 removed `match`'s registry import |
| D-10 | **Tests.** `core/autoreconcile/autoreconcile_test.go` declares `var testRules = matchregistry.Default().Value` beside its other helpers (test files may import the registry; no test mutates it); every `Reconcile(` / `Compatible(` / `moduleFor(` / `dateModule{}` in the package's tests passes it. `TestDateBounds`'s assertions move to `core/valuecompare/datewindow_test.go` `TestDateWindow` *(new)*. `TestPeriodWindowAndLinkMembership`, `mustWindow` and `ptrDate` move to `core/database/conclusionheaders/placechains_internal_test.go` *(new, package `conclusionheaders`)*, calling `periodWindow`, `linkMembership`, `valuecompare.DateWindow`; the date builders it uses (`ymd`, `qualified`, `between`, `ip`) are copied there as unexported test helpers, and `ptrDate` is deleted from `dates_test.go` (only the moved test used it; `TestDateReconcile` uses `datePtr`). Three name subtests keep their cases and change only their titles, because word order inside a part type no longer matters: `TestReconcileNames` "parts of one type compare in idx order, wherever they sit" → "parts of one type compare as words, wherever they sit"; "order matters: [Kenneth, James] does not fold into [James, Kenneth]; each part appears once" → "[Kenneth, James] and [James, Kenneth] are one value; each part appears once"; `TestReconcileNamesWords` "the same words in another order add nothing new to show" stays. The `names.go` header bullets "A unit is the type's words in idx order" and "Fold (subsumption)" are rewritten to D-4's rule | Tests follow the code they test; an internal test reaches the unexported helpers, as `core/graphalign/pick_internal_test.go` does | Exporting the helpers through `export_test.go` — because only this test reads them |
| D-11 | **Docs the PR makes false are corrected in it** (working rule "One matching skill, grown with the layers"). Exact text in STEP-4 | — | — |
| D-12 | **Expected values move only where §8 "Expected changes" says.** A failing assertion that is not a row is a Stop condition | The prototype measured them | Updating whatever fails |
| D-13 | **Commits:** (1) STEP-1; (2) STEP-2, place periods and membership by day; (3) STEP-3 (the reconciler, the cache and Promote's callers: one behavior, since `Reconcile`'s signature change cannot build in halves); (4) STEP-4 docs and skill; (5) STEP-6 plan and map. Each builds and passes `go test -tags fts5 ./core/... ./api/...` | One behavior per commit | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuecompare.DateWindow` *(new)* | func | `DateEnds` + `EndsWindow`; `autoreconcile.WindowOfDate` / `DateBounds` | Replaces both reconciler functions (D-6); `AboutEnds` reads ABT widened, which the sort key and stored bounds must not |
| `nameUnit.compareValue` *(new)* | func | `match.Value.compareValue` (S11-05 D-2) | Same adapter shape for a part-type unit; no shared conversion fits a unit with no pattern |
| `Value.compareValue` *(new, unexported in `modules.go`)* | func | `match.Value.compareValue` (S11-05 D-2) | The same adapter for `autoreconcile.Value`; text, integer and date keys and date folds share it (D-3, D-5) |
| `periodWindow`, `linkMembership` *(moved)* | func | `autoreconcile.PeriodWindow`, `LinkMembership` | Moved, not copied (D-7) |
| `autoreconciler.valueRules` *(new)* | var | `matchregistry.Default().Value` at each call | One value for two call sites (D-8) |
| `core/database/conclusionheaders/placechains_internal_test.go` *(new)* | test | `placechains_test.go` (external, year-only fixture) | Reaches unexported helpers with day dates (D-10) |
| `TestDateWindow` *(new)* | test | `TestDateEnds` | Tests the new helper |

Searches run for this table: `rg -n 'func .*Window\(' core/valuecompare` (only `EndsWindow`, `Always`), `rg -n 'autoreconcile\.(DateBounds|WindowOfDate|PeriodWindow|LinkMembership)' core api` (only `conclusionheaders` and `autoreconciler`), `rg -n 'Compatible\(' core --glob '!*_test.go'` (`core/match/pairs.go`, `core/database/promotealign/exhibits.go`), `rg -n 'func .*compareValue' core` (S11-05's `match` adapter only, after it lands).

**Builds on.** `valuecompare.Compare`, `Key`, `DateEnds`, `EndsWindow` (S11-02 – S11-04); `matchregistry.Default()`; S11-05's rules threading in `match` and `promotealign`.
**Deletes.** `subsumes`, `partFits`, `var spelling`, `textKey`, `dateKey`, `optInt`, `dateWindow`, `dateUnit`, `windowOf`, `pointWindow`, `rangeWindow`, `spanOf`, `contains`, `moreSpecificThan`, `openEnds`, `DateBounds`, `WindowOfDate`, `PeriodWindow`, `LinkMembership` from `core/autoreconcile`; `yearOf` from `conclusionheaders`; `TestDateBounds`; the `matchregistry` import of `core/autoreconcile`.
**Next adds.** S11-07 uses the same `Key` for search and term dedupe. S11-08 mirrors D-2 / D-4 in `FakeStore`. S11-19 replaces `Compatible` in pins with claim-evidence lines; S11-09 replaces it in `pairValues` with property compare, after which `Compatible` is the reconciler's alone.

### Schema

None — no catalog change. The cache rebuilds on open through `CacheVersion` 16 (D-8).

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `datewindow.go` `DateWindow` *(new)* (D-6) |
| `core/autoreconcile` | `autoreconcile.go` `Reconcile` (D-1), `SortKey` dates (D-6), `dateKey` / `optInt` deleted; `modules.go` `moduleFor`, text and integer keys, `Compatible` (D-1, D-3); `names.go` (D-1, D-4); `dates.go` rewritten (D-5) |
| `core/database/autoreconciler` | `autoreconciler.go` `valueRules`, `CacheVersion` 16, `insertRow` (D-6, D-8); `truth.go` (D-8) |
| `core/database/conclusionheaders` | `placechains.go` (D-6, D-7); `compose.go` `periodWindow` (D-7) |
| `core/match` | `pairs.go` `compatible`, `refutedBy` (D-9) |
| `core/database/promotealign` | `exhibits.go` `exhibitCompatible` (D-9) |

```go
func Reconcile(rules valuecompare.Rules, valueType string, candidates []Candidate, concluded *Value, cardinality string) (Result, error)
func Compatible(rules valuecompare.Rules, valueType string, a, b Value) bool
func moduleFor(rules valuecompare.Rules, valueType string) module

type nameModule struct{ rules valuecompare.Rules }
type dateModule struct{ rules valuecompare.Rules }

func DateWindow(d datevalue.Value) (Window, bool) // valuecompare
```

Errors: none new.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — An as-written date window in value compare**
Files: `core/valuecompare/datewindow.go`, `core/valuecompare/datewindow_test.go`. Skill: `add-match-rule`, `add-date-value`.
Do: D-6's `DateWindow`; T-1.
Check: `go test -count=1 ./core/valuecompare/... ./core/matchregistry/...` passes with no tags (`TestLayerPackages` included).

**STEP-2 — Place periods and membership by day**
Files: `core/database/conclusionheaders/placechains.go`, `compose.go`, `core/database/conclusionheaders/placechains_internal_test.go` *(new)*; `core/autoreconcile/dates.go`, `dates_test.go`.
Do: D-7 whole (move `PeriodWindow` / `LinkMembership` as `periodWindow` / `linkMembership` on `valuecompare.DateWindow`, `WindowOfDate` → `valuecompare.DateWindow` in `placechains.go`, `laterStart` / `earlierEnd` by day, `yearOf` deleted); delete `PeriodWindow`, `LinkMembership` and `WindowOfDate` from `core/autoreconcile/dates.go` (no other caller); D-10's move of `TestPeriodWindowAndLinkMembership` with its helpers; T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 -count=1 ./core/autoreconcile/ ./core/database/conclusionheaders/` passes; `TestMembershipDatesByDay` fails when `laterStart` is reverted to year-only (run once, then restore); `rg -n 'autoreconcile\.(PeriodWindow|LinkMembership|WindowOfDate)' core api` prints nothing.

**STEP-3 — Reconcile on value compare**
Files: `core/autoreconcile/*.go` (every file); `core/database/autoreconciler/autoreconciler.go`, `truth.go`; `core/match/pairs.go`; `core/database/promotealign/exhibits.go`. Skill: `add-match-rule`, `add-name-value`.
Do: D-1 – D-6, D-8, D-9, the rest of D-10; T-2 – T-6; then run the suite and update each "Expected changes" row and the three subtest titles D-10 names, nothing else.
Check: `go build ./...`; `go vet -tags fts5 ./core/... ./api/...`; `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; AC-5.1's and AC-5.2's `rg` print nothing.

**STEP-4 — Docs and the matching skill**
Files: `docs/conclusion-reconciliation.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/ideas/international-names.md`, `core/valuecompare/words.go` (header comment only), `.agents/skills/add-match-rule/SKILL.md`.
Do:
- `conclusion-reconciliation.md` §7 table: text "Same value" becomes "the same words, ignoring case, punctuation and spacing (value compare's `Same`; *Sainte-Anne* = *Sainte Anne*); accents count (*Montréal*, *Montreal* are two)"; integer "equal (value compare's `Same`)"; date "the same span of days (§7.1)", fold "span containment"; name fold "containment per part type (§7.2)". Above the table add: "**Grouping policy (decided 2026-10-17, S11-06).** Modules decide sameness with value compare (`matching-model.md` layer 1, as a relative link): values group on `Same` and fold on `Contains`; `Equivalent` (an accent fold, an alias) never groups or folds. Terms, subjects and depictions keep their catalog identity." L144: "Scoring uses matching's comparers" becomes "Scoring uses value compare's similarity".
- §7.1: the `ABT` bullet becomes "`ABT` is a range of one unit of its most precise part, as in matching (`ABT 1985` is 1984 – 1986, so it folds into `1986`). `BEF` is open on the low side through the end of its bound; `AFT` open on the high side from the start of its bound. A range is the window from the start of its first bound to the end of its second; a range with one side is open on the other. Clock times and calendars are not compared."; "Equal structured dates merge" becomes "Dates with the same span merge."; "Qualifiers and ranges fold into the narrowest closed value they contain" stays. Add: "`date_lo` / `date_hi` and the sort key read the date as written (`ABT` not widened)."
- §7.2: the comparable-unit bullet's parenthesis becomes "(value compare's words: case, punctuation and whitespace ignored, particles rejoined so *Mc Donald* = *McDonald*; accents are not folded; word order inside a type does not matter)"; the Fold bullet becomes "**Fold:** a type's words fold into a fuller list when value compare finds the list contains them and every word they pair is the same word or an initial of it: `[J]` → `[James]`, `[James]` → `[James, Kenneth]`, `[J, K]` → `[James, Kenneth]`. An accent fold or an alias never folds (`[Jose]` stays beside `[José, Kenneth]`)."
- §12 dates row: "Dates: value compare's span keys and containment fold, ABT as a range; `date_lo` / `date_hi` and `sort_key` as written".
- `comparison-layers.md` L454: "`DateBounds` and `WindowOfDate` became `valuecompare.DateWindow`, and `PeriodWindow` and `LinkMembership` moved to `conclusionheaders` on it, in S11-06"; the duplicates rows for the `autoreconcile` keys and `Compatible` gain "— done in S11-06"; the "Reconciler grouping rules" bullet under "Decided on 2026-10-17" gains "Decided in S11-06: `Same` groups, `Contains` folds, `Equivalent` never; text never folds; ABT is a range."
- `docs/ideas/international-names.md` L37's bullet becomes: "**Accent merging on the Person page** — decided 2026-10-17 (S11-06): an accent fold is value compare's `Equivalent`, which never groups or folds, so *José* and *Jose* show as two values."
- `core/valuecompare/words.go` header: "Shared by the text and name comparers and by the reconciler's name module (`SplitWords`, `SpellingSimilarity`), so each agrees on what a word and a misspelling is."
- `add-match-rule`, after the "Dates compute similarity through `falloff`" bullet: "**The reconciler composes value compare under its own policy.** `autoreconcile.Reconcile` and `Compatible` take `valuecompare.Rules`; values group on `Key` at `Same` and fold on `Contains`. The cache passes `matchregistry.Default().Value`, never catalog data, so a rule change is a `CacheVersion` bump. `valuecompare.DateWindow` reads a date as written (sort keys, stored bounds, place periods)."
Check: `rg -n 'ABT. uses the same span|Equal structured dates merge|DateBounds|WindowOfDate' docs/conclusion-reconciliation.md` prints nothing; `rg -c 'Grouping policy \(decided 2026-10-17' docs/conclusion-reconciliation.md` prints 1; `rg -n 'valuecompare.DateWindow' .agents/skills/add-match-rule/SKILL.md` prints one line.

**STEP-5 — Full gate**
Files: none.
Do: run the gate on the user's Mac, as `ship-pr` requires.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-06-migrate-reconciler.md` reports no errors.

Commits: one per behavior (D-13). The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Group reconciled values with value compare and read ABT as a range."). No WIP commits.

## 8. Tests

Table-driven where a test holds more than one case (`.cursor/rules/go-tests.mdc`). New cases join the existing tables named here.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/datewindow_test.go` `TestDateWindow` *(new)* | `1 JAN 1970` → 0..0; `1985` → Lo < Hi; `14 MAY 1985` inside 1985's; `BEF 1900` → Lo nil, Hi set; `AFT 1872` → Lo set, Hi nil; `ABT 1817` → equal to `1817`'s window; phrase-only → not ok | D-6, AC-3.5 |
| T-2 | Go | `core/autoreconcile/modules_test.go` `TestModules` (new rows) | AC-1.1, AC-1.2, AC-1.3 as rows of the table | AC-1.1, AC-1.2, AC-1.3, D-3 |
| T-3 | Go | `core/autoreconcile/modules_test.go` `TestCompatible` (new rows) and `TestCompatibleIgnoresAliases` *(new)* | text `Sainte-Anne` / `Sainte Anne` → true; `Montréal` / `Montreal` → false; `York` / `York, Upper Canada` → false; AC-2.4 | AC-2.4, D-2 |
| T-4 | Go | `core/autoreconcile/names_test.go` `TestReconcileNamesWords` (new rows) | AC-2.2, AC-2.3, AC-2.5, AC-2.6; existing `J.` / `James` row stays (AC-2.1) | AC-2.1, AC-2.2, AC-2.3, AC-2.5, AC-2.6, D-4 |
| T-5 | Go | `core/autoreconcile/dates_test.go` `TestDateReconcile` (new rows) | AC-3.1; AC-3.3; AC-3.4 | AC-3.1, AC-3.3, AC-3.4, D-5 |
| T-6 | Go | `core/autoreconcile/dates_test.go` `TestDateFold` (rows per "Expected changes") | AC-3.2 | AC-3.2, D-5 |
| T-7 | Go | `core/database/conclusionheaders/placechains_internal_test.go` `TestMembershipDatesByDay` *(new)*, `TestPeriodWindowAndLinkMembership` (moved) | `membershipDates(placeLink{}, from, to)`: starts `1 MAR 1850` / `1 SEP 1850` → `1 SEP 1850` both ways; ends `31 OCT 1867` / `1 JUL 1867` → `1 JUL 1867` both ways; the moved test unchanged but for the names | AC-4.1, AC-4.2, AC-4.3, D-7 |
| T-8 | Go | The whole Go suite, including `core/database/autoreconciler` rebuild-equals-upkeep tests | Passes; every moved value is a row below | AC-5.3, D-8, D-12 |

**Expected changes.** Measured on the prototype at `9c3b6f1`; every other test passed.

| Test · assertion | Old | New | Rule | Step |
| --- | --- | --- | --- | --- |
| `TestDateFold` "about 1985 does not fold into 1986" | false | true; renamed "about 1985 folds into 1986" | ABT is a range (D-5) | STEP-3 |
| `TestDateFold` "about May does not fold into June" | false | true; renamed "about May folds into June" | ABT is a range | STEP-3 |
| `TestDateFold` "a day folds into its hour" | true | false; renamed "a day and its hour are one value" (equal keys, T-5 checks the grouping) | clock times are not compared | STEP-3 |
| `TestDateFold` "a one-day range folds into that day" | true | false; renamed "a one-day range and that day are one value" | equal windows are `Same` | STEP-3 |
| `TestModules` "inner spacing and punctuation still matter" | `[[1] [2] [3]]`, mixed | `[[1 2] [3]]`, merged; renamed "inner spacing is ignored; U.C. is the word uc" | value compare's text `Same` (AC-1.4) | STEP-3 |
| `TestDateBounds` | — | deleted; its cases are T-1 | `DateBounds` deleted | STEP-1, STEP-3 |
| `TestPeriodWindowAndLinkMembership` | in `core/autoreconcile` | moved to `conclusionheaders` (T-7) | D-7 | STEP-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break if a Decision is not followed: `core/autoreconcile` `TestReconcileNamesInvariants` (shuffled and re-formed inputs), `core/database/autoreconciler` (rebuild equals upkeep), `core/match` `pairs_test.go` and `core/database/promotealign` `exhibits_pin_test.go` (pins on the new sameness), `TestLayerPackages`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: conclusion-reconciliation.md §7, §7.1, §7.2, §12; comparison-layers.md; ideas/international-names.md; words.go header; add-match-rule (STEP-4)
- [ ] PR description: what shipped, AC evidence, deviations, every "Expected changes" row with its rule, manual check
```

## 10. Manual check

1. AC-5.1's and AC-5.2's `rg`: no output.
2. `CGO_ENABLED=1 go test -tags fts5 -count=1 -run 'TestModules|TestCompatible|TestReconcileNames|TestDateReconcile|TestDateFold' -v ./core/autoreconcile/`: passes.
3. In the app, open a project; open it once so the cache rebuilds (version 16). On a Place with toponyms `Sainte-Anne` and `Sainte Anne` from two Sources, the Place page shows one toponym, merged. On a Place with `Montréal` and `Montreal`, it shows both.
4. On a Person with a birth dated `ABT 1985` in one Source and `1986` in another, the birth date shows `1986`, merged, and the `ABT 1985` evidence reads as folded.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A grouping change hides a value a researcher relied on (two clock times on one day, a Julian and a Gregorian date) | Out lines name them; the values stay in the cache with their outcomes; S11-99's dogfood pass looks for it |
| A pin moves in a way no test covers | The prototype at `9c3b6f1` ran Promote's suites on S11-05's rules with no change; D-12 makes any new failing Promote test a Stop condition |
| Name units under `Western`'s untyped weight change a Near minimum outcome | Fold reads only `Contains` with `Same` / initial pairs; `Near` never folds, so the minimum cannot change a grouping |

**Stop and report if:**

- `pairValues`, `refutedBy`, `pairExhibits` or `cfg.Value` no longer has the shape §4 gives.
- An expected value changes that is not a row of "Expected changes", or a row's new value differs.
- `valuecompare.Key` for a name with only untyped parts returns no key for a unit that has words.
- `core/valuecompare` or `core/matchregistry` needs a change beyond `DateWindow`, its test, and the `words.go` header comment.
- Any layer package would import `core/autoreconcile`.

## 12. Open questions

None. The user answered the three product calls on 2026-10-17 (D-2, D-3, D-5).

## 13. PR

**Title:** S11-06 — Migrate the reconciler onto value compare
**Body:** objective in one sentence; what shipped against US-1 – US-5; AC evidence (test names); the "Expected changes" rows with their rules; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-06-migrate-reconciler.md`.
