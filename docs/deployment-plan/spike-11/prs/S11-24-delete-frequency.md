# S11-24 — Delete frequency weighting

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Promote's scores move by a fraction of a point and no view, copy or wire field changes)
**Board:** —
**Depends on:** S11-05 (landed [#451](https://github.com/mendahu/provenencia/pull/451); [plan](S11-05-migrate-matching.md) D-4 deleted the name frequency discount and `Stats.Words`, and left `BenchmarkLoadStatsNames` to this PR). Branch from `main` (S11-12 landed [#480](https://github.com/mendahu/provenencia/pull/480); this PR does not read its code)
**Blocks:** S11-13
**Branch:** `s11-24-delete-frequency`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (rules that hold everywhere; this PR adds no knob to `core/matchregistry`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Frequency](../comparison-layers.md#frequency) (both decisions of 2026-10-16); [`docs/matching-model.md`](../../../matching-model.md) "Removed knobs" (rows for `MPrior`, `UPrior`, frequency `ValueTypeScale` entries, `PropertyAgreement`, `mFor` / `uOr`) and the knob rows `subject.Weights.event.date.Agree` and `subject.Weights.place.toponym.Agree`; [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §4, §4.1, §6, §11 and "Differs from the text above"; [S11-05 plan](S11-05-migrate-matching.md) D-4, D-14 and its Out table; working rules "Frequency weighting is gone", "High tolerance for churn", "Registry numbers: carry over, then estimate", "Sparse data first", "Build, then migrate"
**Planned:** 2026-10-18 against `main` @ `801131a` (S11-12 merged)
**Review:** [`S11-24-delete-frequency.review.md`](S11-24-delete-frequency.review.md) — round 1 fixed, round 2 pending

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

Promote stops weighting a match by how common the value is in the catalog. Today an agreeing event type, toponym, date, integer or text value earns `ln(m / u)` points, where `u` is the share of the catalog's entities that carry that value, lifted to the accept bar when the log-odds come out small. After this PR every property earns a plain registry weight, so the same two records score the same in an empty project and in a large one. This is the spike's decision "Frequency weighting is gone" ([frequency](../comparison-layers.md#frequency)).

**Why now.** S11-13 moves Promote's property scoring onto subject compare, which has no frequency input. Deleting the frequency path first means S11-13 migrates plain weights onto plain weights, and the golden and sparse cases show what the frequency removal moved on its own.

**What stays.** The old walk in `core/graphalign`, its bands, accept bar, name and sex scales, edge support, neighbor credit and `part_of` signature scales (S11-13, S11-14, S11-16 replace them). `graphalign.Stats` keeps `FanOut`, and `promotealign.loadStats` keeps its cached fan-out scan (S11-14 deletes both). `ConflictPenalty` keeps its second role, the conflict-with-decision margin in `proposal()` (S11-16). No Swift, wire, catalog or registry (`core/matchregistry`) change.

## 2. User stories

**US-1.** As a researcher, when I promote a Source into a project of any size, the same pair of records proposes the same band, so that a match does not get weaker because the catalog holds many births or many Yorks.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source place `York` and a filed place `York`, nothing else; `graphalign.Stats{}` | Proposed on the filed place, medium, reason agrees on `toponym`, score exactly 3 (today 3.197) |
| AC-1.2 | AC-1.1 with the `toponym` entry removed from `cfg.PropertyScale` | Weak: the text value-type weight, 2 |
| AC-1.3 | A Source place `Provenance` and a filed place `Provenanced` | Weak, `OutcomePartial` on `toponym` (3 × similarity, under medium) |
| AC-1.4 | AC-1.3 with the `toponym` entry removed from `cfg.PropertyScale` | No match: 2 × similarity is under the accept bar |
| AC-1.5 | A Source place `York` and a filed place `York, Upper Canada` | No match (unchanged) |

**US-2.** As a researcher, when a Source person has only a name and a birth date, Promote proposes the filed person whose name and birth match, and not a look-alike with another birth year (spike goal 1, "Sparse matches work").

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Source person `Ann Ames`, subject of a birth dated 1817. Filed `PER-1` `Ann Ames` with a birth dated 1817, filed `PER-2` `Ann Ames` with a birth dated 1830. Fan-out 1 on the birth signature; no fixed rows | The person row is `PER-1`, strong, score exactly 6 (2.5 name + 1.5 support + 0.5 × (2 event type + 2 date)); the birth row is `EVT-1`, strong, score exactly 6.75 |
| AC-2.2 | The obituary golden (`TestGracieObituaryGolden`) | Passes with no change to its expectations |

**US-3.** As a contributor, no frequency code is left to read or maintain.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The first command below, over code and tests | No output |
| AC-3.2 | The second command, over the comments of the two touched packages | No output |
| AC-3.3 | The third command, over the alignment doc | No output |

The AC-3 commands, copied as written (plain `|` alternation):

```sh
rg -n 'ValueFreq|MPrior|UPrior|mFor|uOr\b|uFor|logOdds|flooredAgreementWeight|PropertyAgreement|agreementWeight|BenchmarkLoadStatsNames|profilePropertyKeys|Frequency: true|Fellegi' core api
rg -n -i 'frequen|log-odds|small-catalog|small catalog|cold-start' core/graphalign core/database/promotealign
rg -n -i 'frequenc|cold-start|log-odds' docs/promote-graph-alignment.md
```

## 3. Scope

| In | Covered by |
| --- | --- |
| Plain weights in `ValueTypeScale`; the walk's property scales read from the registry's subject section (D-1, D-2, D-3) | STEP-1, STEP-3, AC-1.1 – AC-1.5, AC-2.1 |
| Delete `Scale.Frequency`, `MPrior`, `UPrior`, `PropertyAgreement`, `agreementWeight`, `mFor`, `uOr`, `logOdds`, `flooredAgreementWeight`, `uFor`, `valueKey` (D-4) | STEP-1, AC-3.1 |
| Drop the frequency-only parameters from the scoring functions and their callers (D-5) | STEP-1, STEP-2 |
| Delete `Stats.ValueFreq` and the value-frequency scan in `promotealign` (`computeStats`'s first query, `cloneStats`'s copy, `profilePropertyKeys`) (D-6) | STEP-2, AC-3.1 |
| Delete `BenchmarkLoadStatsNames` (D-7) | STEP-2, AC-3.1 |
| Rewrite the tests that set frequencies; add the sparse acceptance test (D-8) | STEP-1, STEP-2, STEP-3 |
| Correct the docs that describe the frequency path (D-9) | STEP-4, AC-3.3 |

| Out | Home |
| --- | --- |
| Promote scoring on subject compare (event type at `subject.Weights.event.event_type`, 0.5 / 20) | S11-13 |
| Fan-out: `Stats.FanOut`, `addFanOut`, `sqlFanOut`, `FanOutLowMax`, `FanOutUnknown`, the stats cache on `graphcache.Graph` | S11-14 |
| `ConflictPenalty`'s conflict-with-decision margin; bands from evidence; `core/graphalign/registry.go` itself | S11-16 |
| `pairExhibits` and `attachExhibits` | S11-19, S11-25 |
| Renaming `TestOneAgreeingPropertyStaysAWeakMatch` (its name says weak, it asserts medium) | not planned: S11-16 rewrites the file |
| `gofmt` drift in `core/database/sources/notes.go` on `main` | not planned in this PR: `check-pr.sh` checks only changed files |

## 4. Current state

Verified at `801131a`. Line numbers are at that commit.

| Area | Where | Today |
| --- | --- | --- |
| Registry | `core/graphalign/registry.go` `DefaultConfig` | `PropertyAgreement` gives toponym +1. `ValueTypeScale` marks `text`, `term`, `integer`, `date`, `name`, `subject` as `{Frequency: true}`. `MPrior` per value type (text 0.9, term 0.95, integer 0.9, date 0.85, name 0.8, subject 0.9), `UPrior` 0.1. `PropertyScale` holds name (2.5 / 3), sex_at_birth and gender (0.6 / 1). Imports `math` for `logOdds` |
| Registry | `core/graphalign/registry.go` `Config`, `Scale` | `Config.PropertyAgreement`, `Config.MPrior`, `Config.UPrior`; `Scale.Frequency` ("derives Weight from catalog frequency … lifts a lone exact agreement to the weak bar"); `Scale.Contradiction`'s comment says zero on a frequency scale means `ConflictPenalty`. Helpers `agreementWeight`, `mFor` (fallback 0.85), `uOr` (fallback 0.1), `logOdds` |
| Scoring | `core/graphalign/score.go` | `ScoreCandidate(ev, probe, edge, provenance, cfg, stats)`, `propertyPoints(ev, probe, cfg, stats)`, `ScoreSimilarity(similarity, comparable, cardinality, property, valueType, vals, cfg, stats)`. `resolvedScale(property, valueType, vals, stats)` fills `Weight` from `flooredAgreementWeight` and `Contradiction` from `ConflictPenalty` when `scale.Frequency`. `flooredAgreementWeight` = max(`logOdds(m, u)`, `AcceptScore`) + bonus. `uFor` and `valueKey` look the probe's values up in `stats.ValueFreq`. `probe` and `stats` are read only to reach `uFor`. Imports `strconv` for `valueKey` |
| Walk | `core/graphalign/align.go` | `scoreCandidate` (L292, L294), `nodeScore` (L363), `proposal` (L656), `alikeNew` (L772 – L773), `neighborsMatch` (L852), `pairNode` (L863) pass `X.Values` and `st.stats` to `propertyPoints` / `ScoreCandidate`. `comparisonsFrom(ev, probe, cfg, stats)` (L879). `alts` (L919) builds `probe` only to pass it to `alternative(sc, probe)` (L945), which passes it only to `comparisonsFrom`. `st.stats.FanOut` (L335) stays |
| Types | `core/graphalign/types.go` `Stats` | `ValueFreq map[string]map[string]float64` and `FanOut map[string]float64`; doc comment "Stats supplies catalog frequencies for Fellegi–Sunter weights." |
| Stats scan | `core/database/promotealign/stats.go` | `loadStats` caches `computeStats` on `graphcache.Graph`. `computeStats` first runs the value-frequency query over `auto_reconciler_values` for `profilePropertyKeys()` and fills `ValueFreq`, then the fan-out scan. `cloneStats` copies both maps. `profilePropertyKeys` reads `match.DefaultProfile` for `vocabkey.PrimaryKinds()`; it has no other caller |
| Exhibits | `core/database/promotealign/exhibits.go` `attachExhibits` (L36), `pairExhibits` (L225) | Both take `stats graphalign.Stats` only to pass it to `graphalign.ScoreSimilarity` (L266), with `[]match.Value{in.value}` as `vals` |
| Caller | `core/database/promotealign/promotealign.go` L66 – L78 | `loadStats` → `graphalign.Align(layer, canon, stats, anchors, &cfg)` → `attachExhibits(…, cfg, stats, namePattern)` |
| Benchmark | `core/database/promotealign/bench_test.go` `BenchmarkLoadStatsNames` (L183 – L218) | Times a cold stats read over 20k named persons. Its imports `core/namevalue/namevaluetest` and `core/valuecodec` have no other user in the file |
| Tests | `core/graphalign/align_test.go` | `TestOneAgreeingPropertyStaysAWeakMatch` (L175) and `TestSpellingVariantIsAWeakerTextMatch` (L219) set `Stats.ValueFreq` and clear `PropertyAgreement`; `TestAcceptedPropertyMatchNominatesItsNeighbor` (L906) computes `math.Log(cfg.MPrior["term"] / cfg.UPrior)`; L652, L664, L678 call `ScoreCandidate` with the old signature; `TestScoreSimilarityWeightedCredit` (L1662) calls `ScoreSimilarity` with `nil, cfg, graphalign.Stats{}`. `TestGracieObituaryGolden` (L369) is the golden |
| Tests | `core/database/promotealign/exhibits_pin_test.go` L22, L47 | Call `pairExhibits(…, graphalign.DefaultConfig(), graphalign.Stats{})` |
| Tests | `core/database/promotealign/promotealign_test.go` `TestSameToponymAloneIsAMediumMatch`, `TestSpellingVariantToponymIsAWeakMatch`, `TestStatsFanOut`, `TestStatsCacheInvalidatesOnWrite` | Pass unchanged on the prototype of D-1 – D-7 |
| Docs | `docs/promote-graph-alignment.md` L83, L116, L173, L200, L272, L299 | Describe value frequencies, `m` / `u` priors and frequency scales as running code |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` "Frequency" (L407 – L419) | Decided text; no "Built in" line |
| Docs | `docs/matching-model.md` "Removed knobs"; knob rows L166, L168 | "Removed knobs" already lists every knob this PR deletes. L166 and L168 explain the date and toponym starts as "today's cold-start" weights, which this PR deletes |
| Registry | `core/matchregistry/subject.go` `defaultSubject` | Person name 2.5 / 3, sex_at_birth 0.6 / 1 (pools gender), place toponym 3 / 3; comments L14 – L15 cite graphalign's scales, L20 and L23 cite the cold-start weights |

**Precedent.** S11-05's `dde125dc` ("Compare Promote's values with value compare and drop the per-type floors and the name frequency discount.") deleted the name half of the same path: `core/match`'s `frequency.go`, `Stats.Words`, `sqlNameWords` / `countNameWords` and their `computeStats` / `cloneStats` lines, and `ScoreSimilarity`'s `discount` parameter. Copy its shape: delete the field, its loader lines and every parameter that only carried it, in one PR, and correct `promote-graph-alignment.md` in the same PR (S11-05 D-4, D-14). Do not copy its size: this PR touches no `core/match` file.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Five value types in `ValueTypeScale` are `{Weight: plain, Contradiction: conflict}`**: `text`, `term`, `integer`, `date`, `name`. The `subject` entry is deleted. `DefaultConfig` declares `accept := 2.0`, `plain := accept` and `conflict := 3.0`, and uses them for `AcceptScore`, `ConflictPenalty` and these entries | Working rule "Registry numbers: carry over": the old starting weights `ln(m / 0.1)` were 2.08 – 2.25, and the knob inventory already rounds the date one to 2 (`subject.Weights.event.date.Agree`). Tying `plain` to `accept` keeps in code what the lift guaranteed: one exact agreement alone reaches the weak bar. Contradiction 3 is what `resolvedScale` filled in from `ConflictPenalty`. `subject` has no reader: no default profile has a subject-typed feature and `pairExhibits` skips `valueType == "subject"` (`exhibits.go` L228), so the lookup falls back to the zero scale as before (M-4) | Unrounded per-type values (2.197, 2.251, 2.14) — false precision from priors that are being deleted; event type at subject compare's 0.5 / 20 — that is S11-13's migration, and moving the walk's numbers twice hides which change moved a golden |
| D-2 | **The walk's property scales come from the registry's subject section.** `DefaultConfig` takes `reg := matchregistry.Default()` once (it already read `Value` from it) and builds name, sex_at_birth, gender and toponym from `reg.Subject.Weights` with a local helper `subject(kind, key) Scale` (`Weight: w.Agree, Contradiction: w.Conflict`): name from `Person`/`PropertyName` (2.5 / 3), sex_at_birth and gender both from `Person`/`PropertySexAtBirth` (0.6 / 1; subject compare pools gender into it), toponym from `Place`/`PropertyToponym` (3 / 3). `PropertyAgreement` is deleted | The inventory names `subject.Weights.place.toponym.Agree` (3) as the knob that replaces the toponym bonus; writing 3 again in `graphalign` would make the "two weight sets for one property" the match registry exists to end (M-1). The name and sex values are already equal in both places, so reading them changes no score; measured on the prototype, every Go test passes. An exact toponym alone is 3 ≥ `MediumScore`: medium, short of strong | A literal `{Weight: 3, Contradiction: conflict}` toponym entry — a second copy kept in step by a comment; event type and date from the subject section — their walk weights are D-1's plain 2 until S11-13 |
| D-3 | **No other number changes.** Edge support, neighbor credit, bands, `WeakScore`, `MediumScore`, `StrongScore`, `AlternativeLimit`, the fan-out knobs and `SignatureScale` keep their literals | Frequency never touched them | — |
| D-4 | **`Scale` keeps only `Weight` and `Contradiction`.** `resolvedScale(property, valueType)` returns the property's scale, else the value type's, else the zero scale; it fills nothing in. Deleted: `Scale.Frequency`, `Config.PropertyAgreement`, `Config.MPrior`, `Config.UPrior`, `agreementWeight`, `mFor`, `uOr`, `logOdds`, `flooredAgreementWeight`, `uFor`, `valueKey`, and the `math` and `strconv` imports they needed. Comments that described the lift or log-odds are rewritten to describe the plain scale | Working rule "High tolerance for churn": no kept-for-now paths. With every entry explicit (D-1, D-2), the implicit `ConflictPenalty` fill-in has no case left | Keeping "zero Contradiction means `ConflictPenalty`" — a rule with no remaining user |
| D-5 | **Parameters that only carried frequency inputs are dropped.** New signatures: `ScoreCandidate(ev match.Evaluation, edge, provenance float64, cfg Config) Scored`; `propertyPoints(ev *match.Evaluation, cfg Config) float64`; `ScoreSimilarity(similarity float64, comparable bool, cardinality string, property match.Property, valueType string, cfg Config) (match.Outcome, float64)`; `comparisonsFrom(ev match.Evaluation, cfg Config) []Comparison`; `alternative(sc scoredCand) Alternative`, and `alts` drops its `probe` lookup; `attachExhibits(q, g, layer, prop, cfg, namePattern)`; `pairExhibits(incoming, neighbor, memberObs, cfg)`. `Align` and `newState` keep `stats Stats` (fan-out) | `probe`, `vals` and `stats` reach nothing but `uFor` once it is gone; an unused parameter is dead code (S11-05 D-4 dropped `discount` the same way). Measured on the prototype: the tree builds and `go vet` passes | Keeping the parameters for S11-13 to drop — S11-13 rewrites these callers anyway, and dead parameters invite a reader to look for a use |
| D-6 | **`Stats` keeps only `FanOut`**, with the doc comment "Stats supplies catalog edge fan-outs." `computeStats` loses the value-frequency query and its loop; `cloneStats` copies only `FanOut`; `profilePropertyKeys` is deleted with the `core/match` and `core/vocabkey` imports it needed in `stats.go`. `loadStats`'s doc comment loses "The property list is the match profiles, not a fixed key list." The cache on `graphcache.Graph` (`Stats`, `SetStats`) stays | Fan-out still reads the cached scan until S11-14 | Deleting `Stats` now — S11-14's scope |
| D-7 | **`BenchmarkLoadStatsNames` is deleted** with the two imports only it used. The other benchmarks in `bench_test.go` stay | It times the name-word count S11-05 removed and the value scan this PR removes; the map scope names it | Re-pointing it at the fan-out scan — no PR asks for that number; S11-21 owns performance |
| D-8 | **Tests follow the new weights, they are not preserved** (working rule "High tolerance for churn"). `TestOneAgreeingPropertyStaysAWeakMatch` and `TestSpellingVariantIsAWeakerTextMatch` drop `Stats.ValueFreq` (they pass `graphalign.Stats{}`), replace `cfg.PropertyAgreement = nil` with `delete(cfg.PropertyScale, match.Property{Key: "toponym", Origin: "provenencia"})`, and their comments describe plain weights; their assertions stay. `TestAcceptedPropertyMatchNominatesItsNeighbor` reads `cfg.ValueTypeScale["term"].Weight`. The `ScoreCandidate`, `ScoreSimilarity` and `pairExhibits` calls take the new signatures. One new test, `TestSparseNameAndBirthDate`, in `core/graphalign/align_test.go` (AC-2.1) | The spike's "Sparse data first" rule makes a name + birth date case each layer PR's acceptance test, and none exists in `graphalign` yet. Measured on the prototype: every Go test passes with only these edits | Adding the sparse case to `promotealign` through a catalog fixture — slower and tests the loader, which this PR does not change |
| D-9 | **Docs that describe the frequency path are corrected here.** `docs/promote-graph-alignment.md`: L83 `stats` row reads "Catalog statistics: edge fan-outs (§6)"; L116 bullet reads "- fallbacks when fan-out is missing;"; L173 bullet is replaced whole by "- **Plain scales** (text, terms, dates, integers, unless a property replaces them): weight 2, contradiction 3. A lone exact agreement is 2, the accept bar and the weak bar. Toponym has its own scale, weight 3, so an exact toponym alone is medium and short of strong until an edge adds support. A partial is the weight times the similarity, so a one-letter toponym typo is weak."; L200 reads "4. **`stats`:** signature fan-outs, computed once and kept on the graph (§8)."; L272's "Thresholds and bands / cold-start priors:" reads "Thresholds and bands:"; L299's last sentence, the one that starts "Value frequencies (`u`) are read for every profile Property", is deleted. `comparison-layers.md` "Frequency" gets a last paragraph: "**Built in S11-24.** Promote's walk scores every property on a plain weight: 2 per value type, and name, sex, gender and toponym on their `subject.Weights`. `Stats.ValueFreq`, the priors and the lift are gone." `docs/matching-model.md` L166's Source reads "estimate: the cold-start frequency weight ln(0.85 / 0.1) ≈ 2.1 that S11-24 removed, rounded" and L168's reads "estimate: the cold-start ln(0.9 / 0.1) + toponym `PropertyAgreement` 1 ≈ 3.2 that S11-24 removed, rounded, so a toponym alone reads medium". `core/matchregistry/subject.go` L20's comment reads `// estimate: the cold-start ln(0.85 / 0.1) S11-24 removed, rounded; ConflictPenalty` and L23's `// estimate: the cold-start + toponym bonus S11-24 removed, rounded; ConflictPenalty`; L14's `// graphalign name PropertyScale` and L15's `// graphalign sex / gender PropertyScale; pools gender` read `// graphalign's name scale; the walk reads it until S11-13` and `// graphalign's sex / gender scale; pools gender; the walk reads it until S11-13` | Working rule "One matching skill, grown with the layers": each PR fixes the docs it makes false; S11-05 D-14 left the L173 bullet to this PR | Rewriting `promote-graph-alignment.md` §6 against the model — S11-16's job |
| D-10 | **No `core/matchregistry` field or number changes, no inventory row is added or removed.** Only comments and inventory wording change (D-9). The walk's own numbers (D-1, D-3) stay in `core/graphalign/registry.go` | `TestKnobInventoryMatchesRegistry` covers `core/matchregistry` only; the inventory already lists these knobs as removed | Moving the walk's remaining knobs into `matchregistry` — S11-13 / S11-16 |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `TestSparseNameAndBirthDate` *(new)* | test | `core/graphalign/align_test.go` `TestExactNameCreditsTheNeighborOnce` (name + birth, one candidate, fixed person) | That test fixes the person and checks the event's credit; the sparse case needs two look-alike candidates and no fixed row. Reuse its helpers `nameVals`, `nameMetas`, `eventMetas`, `termProp`, `vals`, `partSig`, `rowBySubject` |
| `accept`, `plain`, `conflict` locals and the `subject` helper in `DefaultConfig` *(new)* | registry locals | `matchregistry.Default().Value`, already read in `DefaultConfig`; `ConflictPenalty: 3.0` literal | The helper reads numbers the subject section already owns instead of copying them (D-2); the locals keep the five plain entries, `AcceptScore` and `ConflictPenalty` in step until S11-16 deletes the file |

Searched: `rg 'func Test.*Sparse' core` (none), `rg 'datevalue' core/graphalign/align_test.go` (no date helper in this file; the new test builds its dates with `datevaluetest.YMD`, as `core/subjectprofile/entity_test.go` does).

**Builds on.** `graphalign`'s registry file (`DefaultConfig`) and its two-level scale lookup (property, then value type); `core/matchregistry`'s subject section, one weight set for name, sex, gender and toponym. S11-05's deletion of the name half (`dde125dc`).
**Deletes.** `Scale.Frequency`; the `subject` value-type scale; the walk's own copies of the name, sex, gender and toponym numbers; `Config.PropertyAgreement`, `MPrior`, `UPrior`; `agreementWeight`, `mFor`, `uOr`, `logOdds`, `flooredAgreementWeight`, `uFor`, `valueKey`; `Stats.ValueFreq`; `computeStats`'s value-frequency query; `profilePropertyKeys`; `BenchmarkLoadStatsNames`; the `probe` / `vals` / `stats` parameters in D-5; the frequency sentences in `promote-graph-alignment.md`.
**`TestSparseNameAndBirthDate` (T-6), as written:**

```go
func TestSparseNameAndBirthDate(t *testing.T) {
	// A person with only a name and a birth date files on the person whose
	// name and birth match, not on a look-alike born in another year.
	date := match.Property{Key: "date", Origin: "provenencia"}
	year := func(y int) match.Value { return match.Value{Date: datevaluetest.YMD(y, 0, 0)} }
	et, birth := termProp("event_type", "birth")
	metas := append(nameMetas(), eventMetas()...)
	metas = append(metas, match.PropertyMeta{Property: date, ValueType: propertymeta.ValueTypeDate})
	sig := partSig("subject", "event", "birth")
	per, ev := id("s-person"), id("s-event")
	layer := graphalign.Layer{
		Subjects: []graphalign.Subject{
			{ID: per, Ref: "PER-S", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: ev, Ref: "EVT-S", Kind: "event", Values: vals(et, birth, date, year(1817))},
		},
		Bridges: []graphalign.Bridge{{A: per, B: ev, Signature: sig}},
		Metas:   metas,
	}
	p1, e1, p2, e2 := id("h-person-1"), id("h-event-1"), id("h-person-2"), id("h-event-2")
	canon := graphalign.Canon{
		Handles: []graphalign.Handle{
			{ID: p1, Ref: "PER-1", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: e1, Ref: "EVT-1", Kind: "event", Values: vals(et, birth, date, year(1817))},
			{ID: p2, Ref: "PER-2", Kind: "person", Values: nameVals("Ann", "Ames")},
			{ID: e2, Ref: "EVT-2", Kind: "event", Values: vals(et, birth, date, year(1830))},
		},
		Edges: []graphalign.CanonEdge{
			{From: p1, To: e1, Signature: sig},
			{From: p2, To: e2, Signature: sig},
		},
	}
	stats := graphalign.Stats{FanOut: map[string]float64{sig.Key(): 1}}
	p := graphalign.Align(layer, canon, stats, nil, nil)

	cfg := graphalign.DefaultConfig()
	name := cfg.PropertyScale[match.Property{Key: "name", Origin: "provenencia"}].Weight
	event := cfg.ValueTypeScale["term"].Weight + cfg.ValueTypeScale["date"].Weight
	tests := []struct {
		name    string
		subject []byte
		handle  []byte
		score   float64
	}{
		{"person", per, p1, name + cfg.EdgeSupportLow + cfg.NeighborCreditLow*event},
		{"birth", ev, e1, event + cfg.EdgeSupportLow + cfg.NeighborCreditLow*name},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := rowBySubject(p, tt.subject)
			if row.Target != graphalign.TargetHandle || !bytes.Equal(row.HandleID, tt.handle) ||
				row.Assessment != graphalign.AssessStrong || math.Abs(row.Score-tt.score) > 1e-9 {
				t.Fatalf("row %+v, want %s strong at %.2f", row, tt.handle, tt.score)
			}
		})
	}
}
```

`align_test.go` gains the `core/datevalue/datevaluetest` import. With D-1's weights the scores are 6 and 6.75 (measured on the prototype).

**Next adds.** S11-13 replaces `resolvedScale` with subject compare over plain `subject.Weights` and has no frequency case to carry. S11-14 deletes what is left of `Stats` and `loadStats`.

### Schema

None — no catalog change. The value-frequency query only read `auto_reconciler_values`.

### Go core

| Package | Change |
| --- | --- |
| `core/graphalign` `registry.go` | D-1, D-2, D-4: plain `ValueTypeScale`, `PropertyScale` from the subject section, frequency fields and helpers deleted |
| `core/graphalign` `score.go` | D-4, D-5: `ScoreCandidate`, `propertyPoints`, `ScoreSimilarity`, `resolvedScale` on the new signatures; `flooredAgreementWeight`, `uFor`, `valueKey` deleted |
| `core/graphalign` `align.go` | D-5: callers on the new signatures; `comparisonsFrom(ev, cfg)`, `alternative(sc)`, `alts` without `probe` |
| `core/graphalign` `types.go` | D-6: `Stats` holds `FanOut` only |
| `core/database/promotealign` `stats.go`, `exhibits.go`, `promotealign.go`, `bench_test.go` | D-5, D-6, D-7 |

Sketch of the registry after STEP-1, measured on the prototype. The comments in the sketch and below are the text to write (AC-3.2 bans `frequen`, `log-odds`, `small catalog` and `cold-start` in these packages). The existing name / sex comment above `PropertyScale` stays as it is:

```go
func DefaultConfig() Config {
	reg := matchregistry.Default()
	// subject is a property's weights in the registry's subject section, as
	// a scale: the walk scores these properties on the same numbers.
	subject := func(kind vocabkey.SubjectType, key vocabkey.PropertyKey) Scale {
		w := reg.Subject.Weights[kind][key]
		return Scale{Weight: w.Agree, Contradiction: w.Conflict}
	}
	// accept is the least evidence to place a pair. plain, every value
	// type's weight, equals it, so one exact agreement alone reaches the
	// weak bar.
	accept := 2.0
	plain := accept
	// conflict is the points one disagreement costs.
	conflict := 3.0
	return Config{
		Value:           reg.Value,
		AcceptScore:     accept,
		// StrongScore, MediumScore, WeakScore unchanged
		ConflictPenalty: conflict,
		ValueTypeScale: map[string]Scale{
			"text": {Weight: plain, Contradiction: conflict},
			// term, integer, date, name: the same; no subject entry
		},
		PropertyScale: map[match.Property]Scale{
			{Key: "name", Origin: vocabkey.OriginProvenencia}:         subject(vocabkey.Person, vocabkey.PropertyName),
			{Key: "sex_at_birth", Origin: vocabkey.OriginProvenencia}: subject(vocabkey.Person, vocabkey.PropertySexAtBirth),
			{Key: "gender", Origin: vocabkey.OriginProvenencia}:       subject(vocabkey.Person, vocabkey.PropertySexAtBirth),
			{Key: "toponym", Origin: vocabkey.OriginProvenencia}:      subject(vocabkey.Place, vocabkey.PropertyToponym),
		},
		// edge, fan-out, alternatives and SignatureScale unchanged
	}
}
```

Comments to write, word for word:

| Where | Comment |
| --- | --- |
| `ValueTypeScale` (above the map) | `// Value-type defaults. A property entry replaces one of these. Any resemblance value compare reports is partial. A disagreement scores 0 and conflicts. Open bounds that overlap (before 2001, before 1985) resemble; they are not the same date.` |
| Above `PropertyScale`, after the name / sex comment | `// Toponym: an exact toponym alone is 3, medium, and short of strong (4) until an edge adds support.` |
| `Scale.Weight` | `// Weight is the points at similarity 1. The contribution is Weight × similarity.` |
| `Scale.Contradiction` | `// Contradiction is the penalty when value compare finds the values different (similarity 0).` |
| `resolvedScale` | `// resolvedScale is the property's scale, else its value type's, else the zero scale.` |
| `Stats` | `// Stats supplies catalog edge fan-outs.` |

Errors: none added or removed.

### FFI

None — no proto or handler change; `api/` reads nothing this PR deletes (`rg ValueFreq api` is empty at the base).

### macOS

None — no Swift change. Scores reach Swift as bands and reasons, which move only where §2 says.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Plain scales, and scoring without frequency inputs**
Files: `core/graphalign/registry.go`, `core/graphalign/score.go`, `core/graphalign/align.go`, `core/graphalign/align_test.go`. Skill: `add-match-rule` (rules that hold everywhere).
Do: D-1, D-2, D-4 in `registry.go`: `reg`, the `subject` helper and the `accept`, `plain` and `conflict` locals; the five `ValueTypeScale` entries (the `subject` entry deleted); the four `PropertyScale` entries from `subject(…)`; delete `PropertyAgreement`, `MPrior`, `UPrior`, `Scale.Frequency`, `agreementWeight`, `mFor`, `uOr`, `logOdds` and the `math` import; rewrite the `Scale` field comments and the `ValueTypeScale` comment (keep its date sentence: "Any resemblance value compare reports is partial. A disagreement scores 0 and conflicts. Open bounds that overlap (before 2001, before 1985) resemble; they are not the same date."). D-4 and D-5 in `score.go` and `align.go`: `resolvedScale(property, valueType)`; delete `flooredAgreementWeight`, `uFor`, `valueKey` and the `strconv` import; the new `ScoreCandidate`, `propertyPoints`, `ScoreSimilarity`, `comparisonsFrom` and `alternative` signatures and every caller in `align.go`; `alts` without `probe`. D-8 in `align_test.go` for the calls on the old signatures (L652, L664, L678, L1662) and L906; in the two toponym tests, `graphalign.Stats{}` replaces the `ValueFreq` literals (and the `stats.ValueFreq = …` line is deleted) and `delete(cfg.PropertyScale, match.Property{Key: "toponym", Origin: "provenencia"})` replaces each `PropertyAgreement = nil`. Their comments are STEP-3's.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./core/graphalign/` and `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/` pass. `core/database/promotealign` does not build until STEP-2, so STEP-1 and STEP-2 are one commit.

**STEP-2 — The stats scan keeps fan-out only**
Files: `core/graphalign/types.go`, `core/database/promotealign/stats.go`, `core/database/promotealign/exhibits.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/bench_test.go`, `core/database/promotealign/exhibits_pin_test.go`.
Do: D-6 (`Stats`, `computeStats`, `cloneStats`, `profilePropertyKeys`, the `loadStats` comment); D-5 for `attachExhibits` and `pairExhibits` and the `attachExhibits` call in `promotealign.go`; D-7; the two `pairExhibits` calls in `exhibits_pin_test.go`.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./...` passes; `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench . -benchtime 1x ./core/database/promotealign/` runs the remaining benchmarks; AC-3.1's `rg` prints nothing.
Commit (STEP-1 and STEP-2): "Weigh every Promote property with a plain registry weight instead of catalog frequency."

**STEP-3 — Toponym tests and the sparse case**
Files: `core/graphalign/align_test.go`.
Do: D-8: rewrite the comments of `TestOneAgreeingPropertyStaysAWeakMatch` and `TestSpellingVariantIsAWeakerTextMatch` for plain weights (exact toponym 3, medium; without its scale the text weight 2, weak; a spelling variant 3 × similarity, weak; without its scale 2 × similarity, no match); add an exact-score assertion `row.Score == 3` to the first test's first case (AC-1.1). Add `TestSparseNameAndBirthDate` as written in §6 "Go core".
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestOneAgreeingPropertyStaysAWeakMatch|TestSpellingVariantIsAWeakerTextMatch|TestSparseNameAndBirthDate|TestGracieObituaryGolden' -v ./core/graphalign/` passes; AC-3.2's `rg` prints nothing.
Commit: "Test that a sparse name and birth date match the right person on plain weights."

**STEP-4 — Docs**
Files: `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `core/matchregistry/subject.go` (comments only), `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: D-9's edits. Map row S11-24 Status reads `Landed #N` and its Plan column keeps the link. `completed.md` gets, above S11-12: "## S11-24 — Delete frequency weighting" and "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). Promote no longer weighs a value by how common it is in the catalog. Every property scores a plain weight from the walk's registry: 2 for each value type and 3 for a toponym, so an exact toponym alone is still medium and a sparse name and birth date still files strong. `Stats.ValueFreq`, the priors, the lift to the accept bar and the value-frequency scan are gone; the fan-out scan stays until S11-14." This plan: Status `Landed in [#N](…) (date)` and a **Landed** note. Fill N and the date once the PR is open.
Check: AC-3.3's `rg` prints nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-24-delete-frequency.md` reports no errors; links resolve.
Commit: "Document that Promote no longer weighs values by catalog frequency."

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign/align_test.go` `TestOneAgreeingPropertyStaysAWeakMatch` | `York` / `York`, `Stats{}` → handle, medium, agrees on `toponym`, score 3; toponym scale deleted → weak; `York` / `Leeds` → none | AC-1.1, AC-1.2, D-2, D-8 |
| T-2 | Go | `core/graphalign/align_test.go` `TestSpellingVariantIsAWeakerTextMatch` | `Provenance` / `Provenanced` → weak, partial; toponym scale deleted → none; `York` / `York, Upper Canada` → none | AC-1.3, AC-1.4, AC-1.5, D-2 |
| T-3 | Go | `core/graphalign/align_test.go` `TestAcceptedPropertyMatchNominatesItsNeighbor` | The expected score reads the term weight from `ValueTypeScale["term"]` → passes | D-1, D-8 |
| T-4 | Go | `core/graphalign/align_test.go` `TestScoreSimilarityWeightedCredit` | New `ScoreSimilarity` signature, same table → passes | D-5 |
| T-5 | Go | `core/graphalign/align_test.go` `TestGracieObituaryGolden` | Unchanged expectations → passes | AC-2.2 |
| T-6 | Go | `core/graphalign/align_test.go` `TestSparseNameAndBirthDate` *(new)* | Layer: person `PER-S` `nameVals("Ann", "Ames")`; event `EVT-S` `event_type` birth, `date` point 1817; bridge `partSig("subject", "event", "birth")`. Canon: `PER-1` / `EVT-1` (1817) and `PER-2` / `EVT-2` (1830), each person linked to its birth. `Stats{FanOut: {sig.Key(): 1}}`, no fixed rows. Metas: `nameMetas()`, `eventMetas()`, `date` as `ValueTypeDate` → person row `PER-1`, `AssessStrong`, score within 1e-9 of 6; event row `EVT-1`, `AssessStrong`, score within 1e-9 of 6.75 | AC-2.1, D-1 |
| T-7 | Go | `core/database/promotealign/exhibits_pin_test.go` `TestOverlappingBoundsAreSimilarAndUnpinned`, `TestDisjointDatesConflict` | New `pairExhibits` signature → pass unchanged | D-5 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestSameToponymAloneIsAMediumMatch`, `TestSpellingVariantToponymIsAWeakMatch`, `TestStatsFanOut`, `TestStatsCacheInvalidatesOnWrite` | Unchanged → pass | AC-1.1, AC-1.3, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` (exact-score assertions) and `core/database/promotealign` (exhibit weights in proposal tests).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks) — on a non-macOS machine every check but xcode-test passes and xcode-test reports "needs macOS"; report that line, and the CI Xcode run on the non-draft PR is the evidence (no Swift changes)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/promote-graph-alignment.md and comparison-layers.md per D-9
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The reviewer runs, from the repo root:

1. AC-3.1's, AC-3.2's and AC-3.3's `rg`: no output.
2. `CGO_ENABLED=1 go test -tags fts5 -run 'TestOneAgreeingPropertyStaysAWeakMatch|TestSpellingVariantIsAWeakerTextMatch|TestSparseNameAndBirthDate|TestGracieObituaryGolden' -v ./core/graphalign/`: passes (AC-1.1 – AC-1.5, AC-2.1, AC-2.2).
3. (Reviewer, on macOS.) In the app, open a project with a filed place `York` and a Source holding a place `York`; open Promote on that Source. The row proposes the filed `York`, medium (AC-1.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A real catalog had rare values whose frequency weight was above 2 (u < 0.1), so a pair that was strong drops a band | Accepted by the spike decision ([frequency](../comparison-layers.md#frequency)); the golden and sparse cases are the bar, and the prototype passes both |
| An exact toponym at 3 sits exactly on `MediumScore`, so any provenance scale below 1 would drop it to weak | Provenance is 1 for every Subject today (`promote-graph-alignment.md` "Differs from the text above"); S11-16 sets bands from evidence |
| An implementer keeps a dead parameter "for S11-13" | D-5 lists every signature; the prototype shows the tree builds with all of them dropped |

**Stop and report if:**

- A test other than those D-8 names fails after STEP-2, or `TestGracieObituaryGolden` needs an expectation change.
- `TestSparseNameAndBirthDate` does not give `PER-1` strong at 6 with D-1's weights.
- Any package outside `core/graphalign` and `core/database/promotealign` references a name D-4 – D-6 deletes (`rg` at the base shows none).
- `matchregistry.Default().Subject.Weights` no longer holds name 2.5 / 3, sex_at_birth 0.6 / 1 or toponym 3 / 3 at your base (D-2 assumes those values).
- `core/graphalign/registry.go`, `core/graphalign/score.go` or `core/database/promotealign/stats.go` at your base differ from §4 (a later PR moved them).

## 12. Open questions

None.

## 13. PR

**Title:** S11-24 — Delete frequency weighting
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-24-delete-frequency.md`.
