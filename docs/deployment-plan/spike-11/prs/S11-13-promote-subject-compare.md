# S11-13 — Migrate Promote scoring onto subject compare

**Status:** Landed in [#487](https://github.com/mendahu/provenencia/pull/487) (2026-10-18)
**Landed:** Ported tests whose fixtures leaned on the old plain scale were reshaped to keep their intent, each a listed change of §1: `TestSeedFromNewlyFixedIntoUnpromoted` passes fan-out stats (an undated event type now earns 0.5); `TestNewEventsDisagreeWhenTheirParticipantsDo` dates its events (two undated events are under the weak band); `TestPartialNamesAreWeak` fixes the sex-mismatch pair so the row is scored (seeding no longer offers a name agreement with a sex conflict); `TestRowReasons` takes its `weak` case from a neighbor-reached event under the accept bar, and the surname-only look-alike case now reads `no_match`; `TestRankLoserDoesNotNominate` is renamed `TestCandidateLoserDoesNotNominate`; `TestProposeOverlappingBoundShowsTheDateAndResembles` uses a residence instead of a birth (two births "before" different dates rule the person out through the lifespan); `TestProposeJudgesAgainstTheMajority` caches the stray name as `outvoted` instead of `kept` (a handle holds only its conclusion). AC-4.6 is a new case of `TestProposePromoteGraphAlignment` with two look-alike handles. `layerValues` skips Observations that carry a value Subject. STEP-2, STEP-3 and STEP-4 are one commit (D-17). Otherwise none.
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the wire, the page and its copy are unchanged; the derived lifespan stays off the sheet until S11-20, D-6)
**Board:** —
**Depends on:** S11-12 (landed [#480](https://github.com/mendahu/provenencia/pull/480); [plan](S11-12-lifespan.md) and its **Landed** note: STEP-4 and STEP-5 were one commit because a signature change crossed both), S11-24 (landed [#483](https://github.com/mendahu/provenencia/pull/483)). Branch from `main`
**Blocks:** S11-14, S11-15
**Branch:** `s11-13-promote-subject-compare`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the layer rules; this PR adds two layer functions), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) is not touched (no FFI method changes)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Entities](../comparison-layers.md#entities-compare-against-the-conclusion) ("Built in S11-11"), [Lifespan](../comparison-layers.md#lifespan-a-derived-property-for-people) ("Built in S11-12"), [Layer 3](../comparison-layers.md#layer-3-subject-comparison), [Reuse across the app](../comparison-layers.md#reuse-across-the-app), [What moves where](../comparison-layers.md#what-moves-where-in-todays-code); [`docs/matching-model.md`](../../../matching-model.md) knob inventory and "Removed knobs"; working rules "High tolerance for churn", "Build, then migrate", "Sparse data first", "`core/database/matching` is deleted", "One matching skill, grown with the layers", "One wire vocabulary"; user calls of 2026-10-18 written as D-4 and D-6
**Planned:** 2026-10-18 against `main` @ `97eea0b3` (S11-24 merged)
**Review:** [`S11-13-promote-subject-compare.review.md`](S11-13-promote-subject-compare.review.md) — Fits / Ready, 2 rounds (findings fixed)

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

Promote judges a Source Subject against a handle with the layers Spike 11 built, and the old matching code goes. The loader builds each side once per proposal: a Source Subject's side with `subjectprofile.ForSubject` from its own Observations and life events, a handle's with `subjectprofile.Build` from its reconciled values and its participations. `graphalign` scores every pair with `subjectcompare.Compare`, so a birth is never proposed on a death (event type conflict −20) and a person who died before the Subject was born is never proposed for them (lifespan conflict −20). `match.Rank` is replaced by one candidate rule, `graphalign.Candidates`. `core/match` and `core/database/matching` are deleted.

**Why now.** S11-09 to S11-12 built layers 2–4 with no caller (working rule "Build, then migrate"). Slice 3's Check ("a birth is never proposed on a death; a person who died before the Subject's birth is not proposed; `core/match`, `core/database/matching` and the frequency stats are gone") is true only once Promote runs on them. S11-14 and S11-15 build on these profiles.

**What changes for the researcher** (each is the design, not a regression; tests are updated to it, D-16):
- An event's type earns 0.5 when it agrees and costs 20 when it differs (it was 2 and 3). An undated birth is no longer a candidate for another undated birth (0.5 is under the weak band). A dated birth that agrees on type and date scores 2.5, weak (it was 4).
- A person's lifespan counts: −20 when one side's dated events lie outside the other's window, +0.1 when they are consistent. It is not shown on the sheet yet (D-6).
- Sex at birth and gender are one comparison under `sex_at_birth`: a Subject recording only gender meets a handle recording only sex. There is no separate `gender` comparison.
- An event with no `date` reads its `start_date` and `end_date` as one range against the other side's `date` (`subject.Periods`); start and end dates are no longer compared on their own.
- A handle side holds only its conclusion: displayed values as positives, denied values as negatives (strict profile). An outvoted value no longer earns partial credit.
- Seeding and the unreachable fallback offer a handle only when its own properties reach the weak band, 2 points (D-4). A surname-only look-alike ("Robins" against "James Robins", 1.875) and a name agreement with a sex conflict (1.5) are no longer offered by them; an unreachable row with no such handle proposes New. A handle the walk nominates through a mapped neighbor is still recorded, and listed as an alternative, at any score (S11-16 decides).
- An evidence sheet line's weight is what its property counts in the score; a property the subject section does not weigh (an event's `event_name`) shows 0 (D-11).

**What stays.** The walk, the fallback pass, `refineOneHop`, edge support and credit, fan-out, signatures, `ConflictWithFixed`, the bands and the accept bar (S11-14 – S11-16). The wire (`engine.proto`), Swift, copy, and the evidence sheet's layout. `pairExhibits` stays, moved off `match` (S11-19 / S11-25 replace it). Provenance stays 1.

## 2. User stories

Conventions. Bare term names are `vocabkey` constants where one exists (`life(death, …)` is `vocabkey.EventDeath`, `life(birth, …)` is `vocabkey.EventBirth`) and term keys otherwise (`term(male)` is `propertycomparetest.Term("male")`). `name(f)` is `propertycomparetest.Name(f)`, `term(k)` is `propertycomparetest.Term(k)`, `date(y)` is `propertycomparetest.Date(datevaluetest.YMD(y, 0, 0))`. `life(t, y)` is `subjectprofile.LifeEvent{EventType: t, Role: vocabkey.RoleSubject, Dates: []datevalue.Value{*datevaluetest.YMD(y, 0, 0)}}`. `withLife(v, events…)` copies `v` and sets on the copy the keys of `subjectprofile.ForSubject(subjectprofile.Subject{Life: events}, matchregistry.Default().Subject.Lifespans[vocabkey.Person])` (`lifespan` and `lifespan_span`, helper in `align_test.go`, T-4). Metas are `subjectcomparetest.Metas()` unless a row says otherwise. Scores are `Row.Score` under `graphalign.DefaultConfig()`, within 1e-9. "Unreachable" means no anchor reaches the Subject, so only `fallbackUnreachable` can place it.

**US-1.** As the engine, I score a candidate's properties with entity compare, so impossible matches never file and sparse matches still do.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `TestSparseNameAndBirthDate` ported: layer person "Ann Ames" linked (subject · event · birth) to event {birth, 1817}; canon PER-1 "Ann Ames" → EVT-1 {birth, 1817}, PER-2 "Ann Ames" → EVT-2 {birth, 1830}; fan-out 1 | Person → PER-1 strong, score `name + EdgeSupportLow + NeighborCreditLow × event` = 2.5 + 1.5 + 0.5 × 2.5 = 5.25; birth → EVT-1 strong, 2.5 + 1.5 + 0.5 × 2.5 = 5.25, where `name` = `Subject.Weights.person.name.Agree` and `event` = `Weights.event.event_type.Agree + Weights.event.date.Agree` |
| AC-1.2 | Unreachable event {`term(birth)`, `date(1850)`}; one canon event {`term(death)`, `date(1850)`} | Target New, reason `no_match`, no alternatives (evidence −18) |
| AC-1.3 | Layer person P {name "Ann Ames"} fixed to handle H; P linked (subject · event · birth) to event E {birth, 1850}; canon H → D {death, 1850} under the same signature | E's row: Target Skip, Assessment none, reason `weak`; `Alternatives` holds D with Assessment none (the walk nominates D and records its negative score; only `Candidates` is gated, D-4) |
| AC-1.4 | Unreachable person `withLife({name: name("James Robins")}, life(death, 1840))`; one canon person `withLife({name: name("James Robins")}, life(birth, 1850))` | Target New, reason `no_match` (evidence 2.5 − 20) |
| AC-1.5 | AC-1.4 with the Subject's event `life(death, 1920)` | Target handle, Assessment weak, score 2.6; `Comparisons` holds one entry with `Key` `lifespan`, `Result.Outcome` agree, `Points` 0.1, `Derived` true; `ReasonProperty` is `name` |
| AC-1.6 | Unreachable person {name "Mary Robins", gender `term(male)`}; canon person {name "Mary Robins", sex_at_birth `term(male)`} | Target handle, Assessment medium, score 3.1; `Comparisons` keys are `[lifespan name sex_at_birth]`; `sex_at_birth` agrees |
| AC-1.7 | Event S {`term(residence)`, start_date 1850, end_date 1860} fixed to handle {`term(residence)`, `date(1855)`} | `Comparisons` keys are `[date event_type]`; `date` is partial |
| AC-1.8 | Unreachable event {`term(birth)`}; canon event {`term(birth)`} | Target New, reason `no_match` (0.5 is under the weak band) |
| AC-1.9 | Unreachable place with toponym "York" against a handle with the same toponym (`TestOneAgreeingPropertyStaysAWeakMatch`) | Medium, score 3; with `cfg.Subject.Weights[vocabkey.Place][vocabkey.PropertyToponym] = subjectcompare.Weights{Agree: 2, Conflict: 3}` passed to `Align`, weak, score 2 |
| AC-1.10 | Every other `core/graphalign` test, ported (D-16) | Passes; a changed expected number is one of §1's listed changes |

**US-2.** As the engine, one rule says which handles are candidates (user, 2026-10-18).

`graphalign.Candidates(kind, values, handles, metas, cfg, limit)` with probe `{name: name("James Robins")}` and `subjectcomparetest.Metas()`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Handles PER-3 "James Robins", PER-1 "James Robins", PER-2 "Robins", PER-4 "Mary Robins", all person; limit 0 | `[PER-1, PER-3]` (equal evidence and similarity: ref order). The test first asserts PER-2's evidence is under 2 and PER-4's is negative |
| AC-2.2 | AC-2.1, limit 1 | `[PER-1]` |
| AC-2.3 | AC-2.1 plus a place handle PLC-1 whose values are AC-2.1's probe values | PLC-1 is not returned |
| AC-2.4 | Empty probe values | Nothing |

**US-3.** As the engine (the Promote loader), I build both sides from the catalog, life events included, without a read that grows with the layer.

Catalog tests in `core/database/promotealign` with the existing `fixture`. "Dated event" is `f.bare(src, "event")` cited with `event_type` and `date` as in `TestProposeOneHopDateExhibit`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Source A: person "James Robins", subject of a dated death 1840; both promoted. Source B: person "James Robins", subject of a dated birth 1850. Propose B, nothing fixed | B's person: Target New, reason `no_match` |
| AC-3.2 | AC-3.1 with A's death in 1920 | B's person: Target handle (A's person handle), Assessment weak, score 2.6 |
| AC-3.3 | `TestProposeQueryCountDoesNotGrowWithTheLayer` (every person is the subject of a birth on both Sources) | 2 and 6 people take the same number of queries |

**US-4.** As a researcher on the Promote page, the sheet shows only what it can word, and a line's weight is what its property counts.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | AC-3.1 with B's person fixed to A's person handle | The row's `Comparisons` holds one entry with `Key` `lifespan`, `Result.Outcome` conflict, `Points` −20, `Derived` true; no `Exhibit` has `PropertyKey` `lifespan` |
| AC-4.2 | `pairExhibits`, kind event: `start_date` `BEF 2001-03-31` against `BEF 1985-02-01` | One line, partial, unpinned, weight above 0 and at most 2 (`date`'s agree weight) |
| AC-4.3 | `pairExhibits`, kind event: `start_date` `BEF 1985-02-01` against `2001` | One line, conflict, unpinned, weight −3 |
| AC-4.4 | `pairExhibits`, kind event: `event_name` (product origin, text) "Great Fire" against "Great Fire"; kind person: `gender` term female against female | Event name: agree, pinned, weight 0. Gender: agree, pinned, weight 0.6 |
| AC-4.5 | `TestProposeRowWithoutExhibitsShowsValues` | Passes unchanged in its assertions: two ref lines, name "Mary Smith" / "Mary Smith", sex_at_birth "Female" / "Female", no gender line |
| AC-4.6 | `api/ffi/handlers` `TestProposePromoteGraphAlignment` | Every row and alternative with a non-empty `reason_property_key` has `reason_property_origin` `provenencia`; an empty key has an empty origin |

**US-5.** As a contributor, the old matching code is gone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `go build ./...`; `ls core/match core/database/matching` | Builds; both directories are absent |
| AC-5.2 | STEP-5's first grep (in its fenced block) | Prints nothing |
| AC-5.3 | STEP-5's second grep (in its fenced block) | Prints nothing |

## 3. Scope

| In | Covered by |
| --- | --- |
| `subjectcompare.Points` and `Rules.Lead`; `Compare` uses `Points`; `subjectprofile.Value` exported (D-12) | STEP-1 |
| `graphalign` on subject compare: types, `Config.Subject`, scoring, `Candidates`, the fallback, comparisons, the duplicate-New checks; scale code deleted (D-1, D-3, D-4, D-5, D-14) | STEP-2, US-1, US-2 |
| The derived lifespan off the row reason and the sheet lines (D-6) | STEP-2, STEP-4, AC-1.5, AC-4.1 |
| The Promote loader: layer values from the Source graph, life events, handle profiles, metas, seeding through `Candidates` (D-2, D-8, D-9, D-10) | STEP-3, US-3 |
| Exhibits and the FFI handler off `match` (D-7, D-11, D-13) | STEP-4, US-4 |
| Delete `core/match`, `core/database/matching`, `core/database/subjectvocab/profile_test.go` (D-15) | STEP-5, US-5 |
| Slice Check tests in the catalog (US-3, AC-4.1) | STEP-6 |
| Docs the migration makes false (D-18) | STEP-7 |

| Out | Home |
| --- | --- |
| Shortlist keys and an index instead of scanning a kind's handles | S11-15 |
| Segment compare, typed bridge keys, declared birth / death cardinality, fan-out removal | S11-14 |
| The layer 6 walk, top K with ties, similarity as the candidate gate, the alignment registry section, `seedCandidateLimit` and `core/graphalign/registry.go` deleted, `ConflictWithFixed` | S11-16 |
| Showing the lifespan on the sheet, and saying why a person has none when their own dates contradict each other (S11-12 D-12) | S11-20 (scope entry updated with this plan) |
| Replacing `pairExhibits` with property compare's speaking pair (and its second SQL read of the layer's Observations in `loadExhibitObservations`); the registry evidence section | S11-19, S11-25 |
| A "find another record" entry for a record Promote did not offer | [`promote-find-another-record.md`](../promote-find-another-record.md) (moved from `docs/ideas/` on 2026-10-24) |
| Merge hints (entity against entity: `subjectcompare.Compare` on two `Build` profiles) | Not planned in Spike 11 — no UI; the reuse inventory names the shape |
| Provenance scaling comparisons | Not planned in Spike 11 ([`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §12) |
| Benchmark-driven memoization | S11-21 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Old matching | `core/match/` (11 files): `match.go` `Value`, `Values`, `Property`, `Candidate`, `Rank`, `Score`; `compare.go` `Compare`; `pair.go` `Outcome*`, `PropertyMeta`, `Evaluate`, `OutcomeOf`; `pairs.go` `pairValues`, `weigh`, `ValueRef`, `Value.Label`, `Value.Auto`, `SetEvidence`, `HasPositive`; `periods.go` `poolPeriods`; `profiles.go` `Feature`, `Profile`; `registry.go` `DefaultProfile`, `DatePeriods`, `DefaultSuggestionLimit` | Support-weighted pair selection and `Rank`'s 10-point profiles. Imported by `graphalign`, `promotealign`, `core/database/matching`, `core/database/subjectvocab/profile_test.go` |
| Catalog adapter | `core/database/matching/matching.go` `ForSubject`, `ForEntity`, `CandidatesOfType`, `EntitiesValues`, `SubjectsValues`, `sqlSubjectsValues`; `matching_test.go`, `export_test.go` | Only `SubjectsValues` has a caller (`core/database/promotealign/layer.go`); it reads the layer's Observations with SQL |
| Walk types | `core/graphalign/types.go` `Subject.Values`, `Handle.Values` (`match.Values`), `Layer.Metas` (`[]match.PropertyMeta`), `Comparison{Property match.Property; Outcome match.Outcome; ValueType; Weight; Probe, Candidate match.ValueRef}`, `Exhibit.Property match.Property`, `Exhibit.Outcome match.Outcome`, `Alternative.ReasonProperty`, `Row.ReasonProperty` (`match.Property`) | `match` types throughout |
| Walk scoring | `core/graphalign/score.go` `Scored{Score, Edge, Eval match.Evaluation}`, `ScoreCandidate(ev, edge, provenance, cfg)`, `propertyPoints`, `EdgeFact`, `EdgePoints`, `ScoreSimilarity`, `ApplyScale`, `Config.resolvedScale` | Points from a per-property or per-value-type scale over `match.Evaluate` |
| Walk registry | `core/graphalign/registry.go` `DefaultConfig`, `Config{AcceptScore, StrongScore, MediumScore, WeakScore, ConflictPenalty, PropertyScale, ValueTypeScale, EdgeSupportLow, EdgeSupportHigh, NeighborCreditLow, NeighborCreditHigh, FanOutLowMax, FanOutUnknown, AlternativeLimit, Value, SignatureScale}`, `Scale`, `SignatureScale` | `PropertyScale` reads name, sex, gender and toponym from `matchregistry.Default().Subject.Weights`; every other property scores the plain 2 / 3 |
| Walk | `core/graphalign/align.go` `state.metas`, `state.kindCands`, `evaluate`, `scoreCandidate`, `nodeScore`, `fallbackUnreachable` (calls `match.DefaultProfile` and `match.Rank` with `AlternativeLimit`), `candidatesOfKind`, `proposal`, `alikeNew`, `neighborsMatch`, `pairNode`, `strongestAgreement`, `comparisonsFrom`, `alternative`, `subjectHasValues` | Every pair is `match.Evaluate` then `propertyPoints` |
| Walk tests | `core/graphalign/align_test.go` (helpers `textProp`, `termProp`, `vals`, `personMetas`, `placeMetas`, `eventMetas`, `nameVals`, `surnameVals`, `nameMetas`; tests listed by `rg -n '^func Test' core/graphalign/align_test.go`), `core/graphalign/pick_internal_test.go` `TestPickScoreScoresAMissingPick` (scores `occupation` on the plain text scale) | 90 `match.` references in `align_test.go` |
| Layer loader | `core/database/promotealign/layer.go` `loadLayer` (filters on `match.DefaultProfile`, calls `matching.SubjectsValues` per kind), `loadMetas` (profile properties only), `loadLayerBridges` (participation loop reads `roles` and `eventTypes`), `neighborTypeProperty`, `hopPersonToEvent` | Layer values from SQL; no life events |
| Canon loader | `core/database/promotealign/canon.go` `seedCandidateLimit` (5), `loadCanon(g, q, layer, fixed, rules, namePattern)` (seeds with `match.Rank` over `g.Kind` nodes; BFS builds each `Handle.Values` as it meets the node), `canonSignature`, `keptTerm`, `handleValues` (`match.Value` with `SetEvidence`), `propertyByID` (`map[string]match.Property`), `typeByKey`, `typeKey` | Handle values keep support weighting; no life events |
| Entry | `core/database/promotealign/promotealign.go` `Propose`: `loadLayer` → `mergeFixed` → `valuerules.Load` → `loadStats` → `loadCanon` → `DefaultConfig` with `cfg.Value = rules` → `Align` → `attachExhibits` | Rules load after the layer |
| Exhibits | `core/database/promotealign/exhibits.go` `exhibitObs{id, subjectID, prop match.Property, valueType, cardinality, source, value match.Value}`, `attachExhibits`, `withRowLines`, `refLine`, `pairExhibits(incoming, neighbor, memberObs, cfg)` (`match.Compare`, `exhibitCompatible`, `graphalign.ScoreSimilarity`), `exhibitCompatible`, `loadExhibitObservations` (SQL, every property, positive only, `t.label`) | Line weights from the plain scale |
| Exhibit tests | `core/database/promotealign/exhibits_pin_test.go` `TestOverlappingBoundsAreSimilarAndUnpinned`, `TestDisjointDatesConflict` | Call `pairExhibits` with `match.Value` |
| Catalog tests | `core/database/promotealign/promotealign_test.go` (fixture helpers `newFixture`, `person`, `bare`, `cite`, `birth`, `participation`, `promote`, `propose`, `rowFor`, `comparisonOutcome`, `conflictExhibit`, `birthEvent`, `dateLine`; 10 `match.` references) | |
| FFI | `api/ffi/handlers/promote_align.go`: `ReasonPropertyKey: r.ReasonProperty.Key`, `ReasonPropertyOrigin: r.ReasonProperty.Origin` (row and alternative), `PropertyKey: ex.Property.Key`, `PropertyOrigin: ex.Property.Origin`, `Outcome: string(ex.Outcome)`; `promote_align_test.go` `TestProposePromoteGraphAlignment` | |
| Layers | `core/subjectcompare/subjectcompare.go` `Weights`, `Pool`, `Period`, `Lifespan`, `Rules`, `Values`, `Metas`, `Result`, `PropertyResult`, `Compare` (points in a `switch`); `pairs.go` `side`, `period`; `core/subjectprofile/subjectprofile.go` `Entity{Properties, NamePattern, Life}`, `LifeEvent`, `Subject`, `Property`, `Build(e, metas, reconcile, lifespans, opts)`, `ForSubject(s, lifespans)`, `value` (unexported) | No caller outside tests |
| Registry | `core/matchregistry/subject.go` `defaultSubject` (comments "the walk reads it until S11-13" on name and sex) | |
| Source graph | `core/database/graphcache/source.go` `SourceBridge`, `SourceGraph.Bridges` (no writer, no reader: `sourcefill` never sets it); `SourceGraph.Observations` (`SourceObservation`, among its fields `ID`, `SubjectID`, `PropertyID`, `Polarity`, `Text`, `HasText`, `Integer`, `HasInteger`, `TermID`, `Date`, `Name`; filled by `core/database/sourcefill/load.go` `loadSourceObservations`: every Observation on the Source, both polarities, date and name blobs) | Already loaded by `loadLayer` (`g.Source`) |
| Handle graph | `core/database/graphcache/graph.go` `Node{Values map[string][]Value, Links}`, `Value{Rank, Reason, Text, HasText, Integer, HasInteger, TermID, TermKey, Date, Name, Support, Against}`, `Link{Neighbor, TermKey, BridgeTypeKey, FromEnd, …}`, `Graph.Nodes(ids)` (batch load), `Graph.Kind(typeID)`; `follow.go` `Graph.Follow(id, hop)`; `vocab.go` `Vocabulary{Properties, Terms, Types}`, `VocabProperty{ID, Key, Origin, ValueType, Cardinality}`, `VocabTerm{ID, PropertyKey, Key, Label, Neutral}`, `TermsByID` | |
| Seed check | `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` (reads `match.DefaultProfile`); `subject_section_test.go` `TestSubjectSectionMatchesSeed` (S11-10's port of it) | |
| Docs | `docs/matching.md` (describes `core/match`); `docs/promote-graph-alignment.md` lines 105–106, 111, 138, 166–176, 288, 299, 302; `docs/deployment-plan/spike-11/comparison-layers.md` §Entities, §Reuse across the app (rows "match/pairs.go leaders and weighting" and "graphalign alikeNew, pairNode", consumer row 4, "Found in passing" last bullet); `docs/matching-model.md` §Removed knobs; `docs/ideas/README.md` line 36; `docs/ideas/daitch-mokotoff.md` line 127; `.agents/skills/add-match-rule/SKILL.md` intro | |

**Precedent.** S11-05 ([#451](https://github.com/mendahu/provenencia/pull/451), [plan](S11-05-migrate-matching.md)) moved `match`, `graphalign` and `promotealign` onto value compare in one PR and deleted what it replaced: copy its shape (every caller moved, tests ported in place, docs corrected in the same PR). S11-24 ([#483](https://github.com/mendahu/provenencia/pull/483)) changed `graphalign`'s scoring and updated tests to the new numbers by reading them from the config rather than writing literals: copy that. S11-12's **Landed** note: a signature change that crosses packages is one commit (D-17).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Scoring is entity compare.** `graphalign` scores every pair (Subject against handle, and Subject against Subject in `alikeNew`, `neighborsMatch`, `pairNode`) with `subjectcompare.Compare(vocabkey.SubjectType(kind), a, b, st.metas, st.cfg.Subject, st.cfg.Value)`. A pair's property points are `Result.Evidence`. Edge support and credit, the walk, the fallback, refinement, bands and the accept bar are unchanged | The map's scope entry; the walk stays until S11-16 calling the new scoring | Keeping `match.Evaluate` beside it — two scorers; the working rule allows none |
| D-2 | **Profiles are built once, by the loader.** `graphalign.Subject.Values` and `Handle.Values` become `subjectcompare.Values`; `Layer.Metas` becomes `subjectcompare.Metas`. `promotealign` fills them: a Source Subject with `subjectprofile.ForSubject`, a handle with `subjectprofile.Build`, each with `cfg.Subject.Lifespans[vocabkey.SubjectType(kind)]` (nil for event and place). Metas are every product-origin property's catalog value type and cardinality | "Both sides' lifespans can be built before alignment, once each" ([lifespan](../comparison-layers.md#lifespan-a-derived-property-for-people)); `graphalign` stays pure | Building profiles inside `graphalign` — it would need reconciler rows and life events, both catalog shapes |
| D-3 | **Config.** `graphalign.Config` gains `Subject subjectcompare.Rules` (after `Value`, doc "Subject is the registry's subject section: what each property counts. DefaultConfig holds the registry's."), set from `matchregistry.Default().Subject`. `PropertyScale`, `ValueTypeScale`, `Scale`, `ScoreSimilarity`, `ApplyScale`, `resolvedScale` and `propertyPoints` are deleted. `ScoreCandidate(r subjectcompare.Result, edge, provenance float64) Scored`; `Scored{Score, Edge float64; Result subjectcompare.Result}`. In `DefaultConfig` the `plain` and `conflict` locals go; `ConflictPenalty: 3.0` gets the comment `// margin for ConflictWithFixed (S11-16)`; `accept`'s comment becomes `// accept is the least evidence to place a pair.` In `align.go`, `scoredCand.eval` becomes `result subjectcompare.Result`, and `evaluate(kind string, a, b subjectcompare.Values) subjectcompare.Result` | Layer 3 owns property points ([What moves where](../comparison-layers.md#what-moves-where-in-todays-code): `resolvedScale` → layer 3, `ApplyScale` → layer 2's outcome table) | Keeping a scale map keyed by property — a second weight set (registry rule "One weight set") |
| D-4 | **Candidates (user, 2026-10-18).** `graphalign.Candidates(kind string, values subjectcompare.Values, handles []Handle, metas subjectcompare.Metas, cfg Config, limit int) []Handle` in `core/graphalign/candidates.go` *(new)*: the handles of `kind` whose `subjectcompare.Compare` evidence against `values` is at least `cfg.WeakScore`, ordered by evidence (high first), then similarity (high first), then `Ref`, then `ID` bytes; at most `limit`, all when `limit <= 0`. Doc: "Candidates are the handles a Subject is offered on its own properties: entity compare evidence at least the weak band. Shortlisting replaces the scan (S11-15) and top K replaces the limit (S11-16)." `promotealign` seeds with it (`seedCandidateLimit`); `fallbackUnreachable` uses it (`AlternativeLimit`) and then `scoreCandidate` as today | The user's call: every alternative can show a band. One rule for both callers. `WeakScore` is already the walk's band floor | Any positive evidence — every undated same-type event becomes an alternative; similarity alone — S11-16's gate, which needs top K with ties first |
| D-5 | **Comparison shape.** `graphalign.Comparison{subjectcompare.PropertyResult; ValueType string; Derived bool}`: it embeds layer 3's result, so `Key`, `Points` and `Result` (`Outcome`, `Cause`, `Speaking`, `Excluded`) are read from it. `comparisonsFrom(r subjectcompare.Result, kind string, metas subjectcompare.Metas, rules subjectcompare.Rules)` makes one per `r.Properties` entry with `ValueType` = `metas[pr.Key].ValueType` and `Derived` = `rules.Derived(vocabkey.SubjectType(kind), pr.Key)`. Doc on `Derived`: "the lead is a property the profile builder derives (a person's lifespan), with no catalog row". `Result.Speaking.A` is the Subject's side (Compare's `a`). `proposal` fills an empty list when `len(sc.result.Properties) == 0` | One property-result shape; S11-19 reads `Cause` and `Speaking` with no new field | Copying `Outcome`, `Points` and `Speaking` into fields — a second shape that must track the layer; `ValueRef` — a value shape beside `propertycompare.Value` |
| D-6 | **The lifespan stays off the page (user, 2026-10-18).** It counts in `Score` and stays in `Row.Comparisons`. `strongestAgreement` skips `Derived` comparisons, so `ReasonProperty` is never `lifespan`. `withRowLines` skips `Derived` comparisons, so no ref line carries it. S11-12's "say why a lifespan is missing" line moved to S11-20's scope entry (written with this plan) | No copy for it exists; S11-20 words the whole sheet | Showing the raw key `lifespan` — copy without a brief |
| D-7 | **Keys on the wire.** `Row.ReasonProperty` and `Alternative.ReasonProperty` become `vocabkey.PropertyKey` (subject section leads are product properties). `Exhibit.Property` becomes `Exhibit.PropertyKey string` and `Exhibit.PropertyOrigin string` (exhibits include non-product properties). `Exhibit.Outcome` becomes `propertycompare.Outcome`. The FFI handler sends `ReasonPropertyKey: string(key)` and `ReasonPropertyOrigin: vocabkey.OriginProvenencia` when the key is non-empty, `""` otherwise, through `productOrigin(key vocabkey.PropertyKey) string` *(new, unexported, in `promote_align.go`)*; exhibits send their own key and origin. No proto change | The outcome strings and the proto fields are unchanged, so Swift is untouched ("One wire vocabulary") | A `graphalign.Property{Key, Origin}` — a copy of `match.Property` for one caller |
| D-8 | **Layer values come from the Source graph.** `layerValues` in `core/database/promotealign/profiles.go` *(new)* reads `sg.Observations` of each primary Subject whose property has origin `provenencia`, both polarities, grouped by property key and sorted by Observation id within a key (today's `ORDER BY o.id`). Each becomes `autoreconcile.Candidate{ObservationID: o.ID, Value: decode(…), Negative: o.Polarity == observations.PolarityNegative}`. `decode(text string, hasText bool, integer int64, hasInteger bool, termID, date, name []byte, terms map[string]graphcache.VocabTerm) (autoreconcile.Value, error)` *(new, `profiles.go`)* is the one row decoder for both sides: `TermID`, and `TermKey` and `Neutral` from the term; `Date` from `valuecodec.UnmarshalDate` and `Name` from `valuecodec.UnmarshalName` when the blob is non-empty. `NamePattern` is `valuerules.Load`'s. Property keys and origins come from a `map[string]graphcache.VocabProperty` by id that `loadLayer` builds from `g.Vocabulary()` (the same map the `profiler` holds) | The Source is already in memory; one decoder for Source and cache rows | Moving `SubjectsValues` as is — a second read of rows the graph holds; a decoder per row shape — two copies of one rule |
| D-9 | **Life events from event profiles.** Events have no lifespan, so their profiles are built first and a person's life is read from them. `lifeEvent(event subjectcompare.Values, role vocabkey.Role) subjectprofile.LifeEvent` *(new, unexported, `profiles.go`)*: `EventType` = the `Term` of the first positive `event_type` value, `Role` = `role`, `Dates` = the `Date` of every positive `date`, `start_date` and `end_date` value, in any order. *Layer:* `loadLayerBridges` also returns `[]participation` *(new type in `layer.go`: `{person, event []byte; role string}`)*, one per participation edge: `person` = `e.From`, `event` = `e.To`, `role` = `roles[string(e.Association)]`; its signatures keep reading `eventTypes` as today. `loadLayer` builds every event Subject's values with `ForSubject` first, then each person's `Life` with `lifeEvent(eventValues, vocabkey.Role(p.role))` per participation. *Handle:* for each link of `g.Follow(n.ID, hopPersonToEvent)` whose neighbor node is present and unmerged, `lifeEvent(p.values(l.Neighbor), vocabkey.Role(l.TermKey))`, after the profiler has built that event (D-10). Life is read only for a kind with a `cfg.Subject.Lifespans` entry | One life rule for both sides, on the one "displayed" rule `Build` and `ForSubject` apply; no second read of event types; start and end dates on the timeline give the same earliest and latest ends as their range | Two extractions per row shape — `keptTerm` (rank-1 kept) and a `"kept"` filter would be second copies of `Reason.Displayed`; reading life through `hops.EventsOfSubject` — subject role only, while the lifespan counts every role (S11-12 D-1) |
| D-10 | **Handle profiles through one memo.** `profiler` *(new, unexported, `profiles.go`)* holds the graph, properties by id (`map[string]graphcache.VocabProperty`), terms by id, the metas, `types` (from `typeByKey`), the subject section, the name pattern and `done map[string]subjectcompare.Values`. `profile(nodes []*graphcache.Node) error`: for nodes not in `done`, the kind is `typeKey(types, n.SubjectTypeID)`; one `g.Nodes` call loads every participation neighbor of those nodes whose kind has a lifespan; those neighbor events are built first (no life), then the remaining nodes, each person with its life (D-9); every `Build` result goes into `done`, so an event met as a neighbor and as a handle is built once. `values(id []byte) subjectcompare.Values` reads `done`. `loadCanon` keeps nodes, not `Handle`s, while it walks: `handles map[string]*graphcache.Node` and `kinds map[string]string`; `add` records the node; before seeding a kind it calls `profile` on that kind's `g.Kind` nodes; after the BFS it calls `profile` on every collected node (fixed, seeded and reached handles alike) and builds `Canon.Handles` with `values`. `reconciled(row graphcache.Value, terms) (autoreconcile.ReconciledValue, error)` *(new)* is `decode` plus `Support`, `Against` and `Reason: autoreconcile.Reason(row.Reason)`. `Build` gets `matchregistry.Default().Value` and `subjectprofile.Options{}` | Two neighbor loads per person would grow with the layer (AC-3.3); a handle met by seeding and by the BFS is built once | Building each handle as the BFS meets it — one event load per person |
| D-11 | **Exhibits off `match`.** `exhibitObs{id, subjectID []byte; key, origin, valueType, cardinality, source, termLabel string; auto autoreconcile.Value; value propertycompare.Value}`. `loadExhibitObservations` also selects `o.value_term_id` into `auto.TermID` and scans the property into `key` and `origin`; `auto.TermKey` is the term key; the key-as-id rule of `match.Value.Auto` goes. `value` is `subjectprofile.Value(auto, false, namePattern)` with `Type` = `valueType`. The skip of a line whose value type is empty or `subject` stays. `pairExhibits(incoming []exhibitObs, neighbor []byte, kind vocabkey.SubjectType, memberObs []exhibitObs, cfg graphalign.Config)`: a pair's result is `propertycompare.Compare([]propertycompare.Value{in.value}, []propertycompare.Value{mem.value}, propertymeta.Meta{ValueType: in.valueType, Cardinality: in.cardinality}, cfg.Value)`; it is comparable when the outcome is not unknown; its similarity is `Speaking.Result.Similarity`. The first `autoreconcile.Compatible(cfg.Value, in.valueType, in.auto, mem.auto)` member is chosen and pinned; otherwise the comparable member with the highest similarity. Today's "pinned but not comparable reads as similarity 1" rule goes (a Compatible pair is `Same` or `Contains`, which layer 2 always judges). A line whose outcome is unknown is dropped. Weight: when `in.origin` is `provenencia` and `lead, ok := cfg.Subject.Lead(kind, vocabkey.PropertyKey(in.key))`, `subjectcompare.Points(outcome, similarity, cfg.Subject.Weights[kind][lead])`; otherwise 0. `kind` is the row's kind for its own records and the neighbor Subject's kind for a neighbor group (from `layer.Subjects`). `Pinned` = chosen by Compatibility and outcome agree | Layer 2 already owns neutral terms, cardinality and the outcome; the line weight now matches what the property counts; the real term id is one column away | Keeping `match.Compare` — the package goes; per-line weights from the row's comparison — one number for many lines |
| D-12 | **Three layer helpers and one export.** `subjectcompare.Points(outcome propertycompare.Outcome, similarity float64, w Weights) float64` *(new)*: agree → `w.Agree`; partial → `w.Agree × similarity`; conflict → `−w.Conflict`; unknown → 0. `Compare` sets `pr.Points` with it. `func (r Rules) Derived(kind vocabkey.SubjectType, key vocabkey.PropertyKey) bool` *(new)*: true when `key` is a `r.Lifespans[kind]` lead or a lifespan's `Span`. `func (r Rules) Lead(kind vocabkey.SubjectType, key vocabkey.PropertyKey) (vocabkey.PropertyKey, bool)` *(new)*: `key` when it is in `r.Weights[kind]`; else the first lead, in ascending key order, whose `r.Pools[kind]` `With` holds `key` or whose `r.Periods[kind]` `Start` or `End` is `key`; else `"", false` (a `Span` is in neither, so it is not ok). `subjectprofile.value` is renamed `Value`, doc "Value is the one conversion from the reconciler's value to layer 2's. Type and Ref stay empty: layer 2 sets the type from the meta. Name and Date are shared with v." | One points rule for the score and the sheet; the registry, not a missing catalog row, says what is derived; one place that knows which lead a property joins | A points switch in `promotealign` — a second copy of layer 3's rule; "no catalog meta" as derived — a product property missing from a catalog would read as derived |
| D-13 | **Labels.** `valueLabel(v propertycompare.Value, termLabel string) string` *(new, `exhibits.go`)*, by `v.Type`: name → `v.Name.Form`, or its parts' values joined with one space when `Form` is empty; date → `datevalue.CompactDisplay(*v.Date)`; term → `termLabel`, or `v.Term` when empty; integer → `strconv.FormatInt(v.Integer, 10)`; otherwise `v.Text`. An exhibit line labels with the Observation's `termLabel`. A ref line labels each speaking value with `termLabel` = the vocabulary label of the comparison's lead property and the value's term key, else of each `cfg.Subject.Pools[kind][lead].With` property in order, else empty | `match.Value.Label` moves with its only caller; `Type` is set by layer 2 on the speaking pair | Term keys on the sheet — "Female" would read "female" (AC-4.5) |
| D-14 | **Has values.** The package function `subjectHasValues(s *Subject)` becomes the method `state.subjectHasValues(s *Subject)` (one call site, in `proposal`); it is true when `s.Values` holds a positive value (`propertycompare.Value.Positive`) under a key `k` with `cfg.Subject.Lead(kind, k)` ok and `cfg.Subject.Derived(kind, k)` false. A Subject with only a derived lifespan, or only properties the subject section does not read (a remark), still proposes Skip with reason `empty`. The `loadCanon` seed loop has no pre-check: it calls `Candidates` for every unfixed Subject of a kind with a `cfg.Subject.Weights` entry (a Subject with nothing to match on reaches no evidence, since negatives only conflict and a lifespan agree is 0.1) | "Empty" means nothing recorded that matching reads, as today's profile properties did | Counting every product property — a remark-only person would turn New; a pre-check in `promotealign` — a copy of this rule |
| D-15 | **Deletions.** `core/match/` and `core/database/matching/` whole; `core/database/subjectvocab/profile_test.go` (`TestSubjectSectionMatchesSeed` is its port); `graphalign`'s scale code (D-3); `TestScoreSimilarityWeightedCredit` (support weights are gone; partial names stay pinned by `TestPartialNamesAreWeak`); `promotealign`'s `handleValues`, `propertyByID` and `match.Rank` seeding | Working rule "High tolerance for churn": every caller moves, the old code goes | — |
| D-16 | **Ported tests.** In `align_test.go`: `textProp(key, text string) (vocabkey.PropertyKey, propertycompare.Value)` with `propertycomparetest.Text`; `termProp` with `propertycomparetest.Term`; `vals` builds `subjectcompare.Values`; `personMetas`, `placeMetas`, `eventMetas`, `nameMetas` return `subjectcompare.Metas` with today's value types (names stay text in `personMetas`); `nameVals` and `surnameVals` build `propertycompare.Value{Value: valuecompare.Value{Name: …}}`. New tests use `subjectcomparetest.Metas()`. A test that reads `cfg.PropertyScale[name]` reads `cfg.Subject.Weights[vocabkey.Person][vocabkey.PropertyName]`; `cfg.ValueTypeScale["term"]` for an event type reads `Weights[vocabkey.Event][vocabkey.PropertyEventType]`; `["date"]` reads `Weights[vocabkey.Event][vocabkey.PropertyDate]`. A sub-case that deletes the toponym scale sets AC-1.9's override instead. A sub-case that calls `match.Evaluate` then `ScoreCandidate` calls `subjectcompare.Compare(vocabkey.Person, a, b, metas, cfg.Subject, cfg.Value)` then `ScoreCandidate(r, 0, 1)`. `pick_internal_test.go` scores `name` (text meta) instead of `occupation`. **Expected changes, and only these:** (a) `Comparisons` lists every lead of the kind, unknown ones included (`lifespan` on a person), has no `gender`, `start_date` or `end_date` entry, and a test that counts comparisons or reads one by key is updated to that; (b) a score is recomputed from `cfg.Subject.Weights`, which changes events (type 0.5 / 20) and anything credited from an event node; (c) a handle that seeding or the fallback offered under 2 points is no longer recorded (D-4); (d) an exhibit line's weight (D-11). Any other change is a stop condition. A test whose name states the behavior (`TestSparseNameAndBirthDate`, `TestGracieObituaryGolden`, `TestPartOfChainIsStrongWithoutADecision`, `TestPartialNamesAreWeak`, `TestPossibleDuplicates`, `TestUnknownSexIsNoEvidence`, `TestProposeJudgesAgainstTheMajority`, `TestPartOfChainIsAStrongMatch`) keeps every Target and handle it asserts | The S11-24 precedent; the goldens are the acceptance tests ("Sparse data first"). The skill's "metas with `subjectcomparetest.Metas`" sits in its Layer 3 bullet and binds layer tests; the walk goldens compare names as text on purpose | Rewriting the ported suite on `subjectcomparetest.Metas()` — every text name becomes a name value, so the goldens' numbers move for no behavior |
| D-17 | **Commits.** STEP-2, STEP-3 and STEP-4 are one commit: `graphalign`'s types change breaks `promotealign` and the FFI handler until all three move. Every other step is its own commit | S11-12's Landed note | Three commits that do not build |
| D-18 | **Docs.** Exact text in STEP-7. `matching.md` becomes a short page on where the layers run; the rest corrects named code | Working rule "One matching skill, grown with the layers": each migration PR fixes the docs it makes false | Deleting `matching.md` — eight live docs link to it |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphalign.Candidates` *(new)*, `core/graphalign/candidates.go` *(new)* | func, file | `match.Rank` | Rank scores on its own 10-point profiles; deleted (D-4) |
| `graphalign.Config.Subject` *(new field)* | field | `Config.PropertyScale`, `ValueTypeScale` | Replaces them with the registry section (D-3) |
| `graphalign.Comparison` embeds `subjectcompare.PropertyResult`; `.Derived` *(new field)* | field | `Comparison.Property`, `.Outcome`, `.Weight`, `.Probe`, `.Candidate` | The layer's result already holds them (D-5) |
| `graphalign.Exhibit.PropertyKey`, `.PropertyOrigin` *(new fields)* | field | `Exhibit.Property` (`match.Property`) | `match` goes; no other {key, origin} type exists in `graphalign` (D-7) |
| `subjectcompare.Points` *(new)* | func | `Compare`'s `switch`; `graphalign.ApplyScale` | One points rule for the score and the sheet (D-12) |
| `subjectcompare.Rules.Derived` *(new)* | method | `Rules.Lifespans` | Reads the declaration instead of guessing from a missing catalog row (D-12) |
| `subjectcompare.Rules.Lead` *(new)* | method | `side` (`pairs.go`), `graphalign.Config.resolvedScale` | `side` builds values for a known lead; `Lead` answers the reverse question for one property (D-12) |
| `subjectprofile.Value` *(renamed export)* | func | `subjectprofile.value` | Same function; exhibits need the one conversion (D-12) |
| `core/database/promotealign/profiles.go` *(new)*: `layerValues`, `decode`, `profiler`, `profiler.profile`, `profiler.values`, `reconciled`, `loadMetas` (moved from `layer.go`) | file, func, type | `matching.SubjectsValues`; `canon.go` `handleValues`, `propertyByID`; `layer.go` `loadMetas` | The loader that feeds the profile builder (working rule "`core/database/matching` is deleted"); reads the Source graph instead of SQL (D-8, D-10) |
| `lifeEvent` *(new, `profiles.go`)* | func | S11-12's `subjectprofile.LifeEvent`; `canon.go` `keptTerm` | Reads an already-built event profile, so both sides share one rule (D-9); `keptTerm` stays for `canonSignature` only |
| `participation` *(new type, `layer.go`)* | type | `graphcache.SourceBridge` | Dead: no writer or reader, so this PR deletes it with `SourceGraph.Bridges` (STEP-3); `loadLayerBridges` already reads the role and event type (D-9) |
| `valueLabel` *(new, `exhibits.go`)* | func | `match.Value.Label`; `propertycomparetest.Label` (test-only) | Moves with its caller onto layer 2's value (D-13) |
| `productOrigin` *(new, `promote_align.go`)* | func | — (searched `OriginProvenencia` uses in `api/ffi/handlers`) | Two wire fields need it (D-7) |
| `TestCandidates`, `TestBirthIsNeverProposedOnADeath`, `TestLifespanRulesOutAPerson`, `TestGenderPoolsWithSex`, `TestEventPeriodStandsForTheDate`, `TestLead`, `TestPoints`, `TestDerived`, `TestProposeLifespanFromLifeEvents`, `TestProposeKeepsTheLifespanOffTheSheet`, `TestExhibitWeightIsTheLeadWeight` *(new)*; helper `withLife` *(new, `align_test.go`)* | test, helper | `TestSparseNameAndBirthDate`, `TestUnreachableFallbackScoresOnTheWalkScale`; `subjectprofile` `TestEntityCompare` | Same table shapes (D-16) |

**Builds on.** `subjectcompare.Compare` and the registry's subject section (S11-10), `subjectprofile.Build` / `ForSubject` (S11-11), lifespans (S11-12), `propertycompare.Compare` (S11-09), `valuerules.Load`, the cached Source graph and canon graph, the `add-match-rule` skill.
**Deletes.** `core/match/` (11 files) and `core/database/matching/` (3 files); `core/database/subjectvocab/profile_test.go`; `graphalign` `Scale`, `PropertyScale`, `ValueTypeScale`, `ScoreSimilarity`, `ApplyScale`, `resolvedScale`, `propertyPoints`, `kindCands`'s `match.Candidate`, `TestScoreSimilarityWeightedCredit`; `promotealign` `handleValues`, `propertyByID`, the `match.Rank` seed loop, `exhibitCompatible`; `graphcache.SourceBridge` and `SourceGraph.Bridges` (already dead); `docs/matching.md`'s description of `core/match`; the registry's "until S11-13" comments.
**Next adds.** S11-14's segment compare reads the neighbor through these handle profiles and the bridge through `subjectcompare`. `Rules.Lead`'s only caller is `pairExhibits`; S11-19 / S11-25 decide whether it outlives it. A new derived property is one `Rules.Lifespans`-style entry that `Rules.Derived` reads. S11-15's shortlist index replaces the kind scan inside `Candidates`. S11-16 replaces `Candidates`' limit with top K and moves the gate to similarity. S11-19 builds evidence lines from `Comparison.Speaking` and words `Derived` leads; S11-20 shows the lifespan.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/subjectcompare` | `Points`, `Rules.Derived`, `Rules.Lead`; `Compare` uses `Points` (D-12) |
| `core/subjectprofile` | `value` → `Value` (D-12) |
| `core/matchregistry` | `subject.go` comments only (STEP-1) |
| `core/graphalign` | `types.go`, `registry.go`, `score.go`, `align.go`, `candidates.go` *(new)*; tests (D-1, D-3 – D-6, D-14, D-16) |
| `core/database/promotealign` | `profiles.go` *(new)*, `layer.go`, `canon.go`, `promotealign.go`, `exhibits.go`; tests (D-2, D-8 – D-11, D-13) |
| `core/database/graphcache` | `source.go`: `SourceBridge` and `SourceGraph.Bridges` deleted (dead) |
| `core/match`, `core/database/matching` | deleted (D-15) |

```go
// graphalign

type Subject struct {
	ID         []byte
	Ref        string
	Kind       string // person, event, place
	Values     subjectcompare.Values
	Provenance float64
}

type Handle struct {
	ID     []byte
	Ref    string
	Kind   string
	Values subjectcompare.Values
}

type Layer struct {
	Subjects []Subject
	Bridges  []Bridge
	Metas    subjectcompare.Metas
}

type Comparison struct {
	subjectcompare.PropertyResult        // Key, Result (Outcome, Cause, Speaking: A the Subject's value, B the handle's), Points
	ValueType string                     // the lead's catalog value type; empty when Derived
	Derived   bool                       // the lead is a property the profile builder derives (a person's lifespan), with no catalog row
}

func Candidates(kind string, values subjectcompare.Values, handles []Handle, metas subjectcompare.Metas, cfg Config, limit int) []Handle
func ScoreCandidate(r subjectcompare.Result, edge, provenance float64) Scored

// subjectcompare

func Points(outcome propertycompare.Outcome, similarity float64, w Weights) float64
func (r Rules) Derived(kind vocabkey.SubjectType, key vocabkey.PropertyKey) bool
func (r Rules) Lead(kind vocabkey.SubjectType, key vocabkey.PropertyKey) (vocabkey.PropertyKey, bool)

// promotealign (unexported)

type participation struct {
	person, event []byte
	role          string
}

func loadLayer(g *graphcache.Graph, q Querier, sourceID []byte, cfg graphalign.Config, namePattern string) (graphalign.Layer, map[string]primarySubject, error)
func loadLayerBridges(g *graphcache.Graph, q Querier, sourceID []byte, primary map[string]primarySubject, byKind map[string][][]byte) ([]graphalign.Bridge, []participation, error)
func loadCanon(g *graphcache.Graph, layer graphalign.Layer, fixed []graphalign.Fixed, cfg graphalign.Config, namePattern string) (graphalign.Canon, error)
func pairExhibits(incoming []exhibitObs, neighbor []byte, kind vocabkey.SubjectType, memberObs []exhibitObs, cfg graphalign.Config) []graphalign.Exhibit
```

`loadCanon` loses its `q` argument (no SQL left in it; check with `rg -n 'q\.' core/database/promotealign/canon.go` before removing — any hit is a stop condition). `Propose` becomes: revision → source exists → `valuerules.Load` → `cfg := graphalign.DefaultConfig(); cfg.Value = rules` → `loadLayer(…, cfg, namePattern)` → empty layer returns → `mergeFixed` → `loadStats` → `loadCanon(…, cfg, namePattern)` → `Align` → `attachExhibits`. `loadLayer` keeps a primary Subject when `vocabkey.IsPrimaryType` holds and `cfg.Subject.Weights[vocabkey.SubjectType(kind)]` exists (replacing `match.DefaultProfile`).

### FFI

None in the proto. `api/ffi/handlers/promote_align.go` reads the new fields (D-7).

### macOS

None — no Swift change; the wire is unchanged.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Layer helpers**
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go`, `core/subjectprofile/subjectprofile.go`, `core/matchregistry/subject.go`, `core/matchregistry/value.go`. Skill: `add-match-rule`.
Do: D-12. `TestPoints`, `TestDerived` and `TestLead` (T-1). In `defaultSubject`, the name comment becomes `// carried from graphalign's name scale` and the sex comment `// carried from graphalign's sex / gender scale; pools gender`. In `core/matchregistry/value.go`, the comment `// graphalign's name PropertyScale.Floor.` becomes `// carried from graphalign's name scale floor.` and `// graphalign ValueTypeScale["date"].Floor: no floor.` becomes `// carried from graphalign's date scale: no floor.` In `Compare`, the `switch` keeps accumulating `counted` and `score`; after it, `pr.Points = Points(res.Outcome, sim, w)` where `sim` is `res.Speaking.Result.Similarity` for a partial and 0 otherwise.
Check: `go test ./core/subjectcompare/... ./core/subjectprofile/... ./core/matchregistry/...` passes with no tags; the existing `TestCompare` rows are unchanged.

**STEP-2 — Graph alignment on subject compare**
Files: `core/graphalign/types.go`, `registry.go`, `score.go`, `align.go`, `candidates.go` *(new)*, `align_test.go`, `pick_internal_test.go`.
Do: D-1, D-3, D-4, D-5, D-6 (`strongestAgreement`), D-14. `state.kindCands` holds `[]Handle` sorted as `candidatesOfKind` sorts today (renamed `handlesOfKind`). `fallbackUnreachable` loops over `Candidates(s.Kind, s.Values, st.handlesOfKind(s.Kind), st.metas, st.cfg, st.cfg.AlternativeLimit)`, skipping a kind with no `st.cfg.Subject.Weights` entry. `proposal` fills an empty comparison list with `comparisonsFrom(st.evaluate(s.Kind, s.Values, h.Values), s.Kind, st.metas, st.cfg.Subject)`; `alternative` calls `comparisonsFrom` the same way. Port the tests (D-16). New tests T-2 – T-6.
Check: `go test ./core/graphalign/...` passes with no tags. `go build ./...` fails only in `core/database/promotealign` and `api/ffi/handlers` (D-17).

**STEP-3 — The Promote loader builds profiles**
Files: `core/database/promotealign/profiles.go` *(new)*, `layer.go`, `canon.go`, `promotealign.go`.
Do: D-2, D-8, D-9, D-10 and the `Propose` order in §6. `loadLayer`: primary Subjects → `loadLayerBridges` (bridges and participations) → `layerValues` → `ForSubject` for each event Subject → each person's `Life` (D-9) → `ForSubject` for the other Subjects. `loadCanon`: the seed loop builds, per kind, `[]graphalign.Handle` from `g.Kind` nodes after `profiler.profile`, and adds `graphalign.Candidates(s.Kind, s.Values, handles, layer.Metas, cfg, seedCandidateLimit)`; the BFS collects nodes and fills `Values` from the profiler after it ends. `canonSignature`'s and `keptTerm`'s `props` argument becomes `map[string]graphcache.VocabProperty` (every property by id; the field they read, `Key`, is the same). Delete `graphcache.SourceBridge` and the `SourceGraph.Bridges` field (`core/database/graphcache/source.go`); `go build ./...` at STEP-4 proves nothing used them, and `rg -n 'SourceBridge' --type go` then prints nothing.
Check: `go build ./core/database/promotealign/` reports errors only in `exhibits.go` (STEP-4 moves it); `go test ./core/database/graphcache/ ./core/database/sourcefill/` builds (`-tags fts5`, `CGO_ENABLED=1`). The loader's behavior is checked by STEP-4's catalog suites and STEP-6.

**STEP-4 — Exhibits and the handler off `match`**
Files: `core/database/promotealign/exhibits.go`, `exhibits_pin_test.go`, `promotealign_test.go`, `api/ffi/handlers/promote_align.go`, `api/ffi/handlers/promote_align_test.go`.
Do: D-6 (`withRowLines`), D-7, D-11, D-13. `attachExhibits` builds the term labels from `g.Vocabulary()` and the Subject kinds from `layer.Subjects`. `withRowLines(exhibits, comparisons, label func(graphalign.Comparison, propertycompare.Value) string)` and `refLine(c, label)`; `attachExhibits` builds `label` once per row as a closure over the row's kind (D-13). With `sp := c.Result.Speaking`: `IncomingDisplay` = `label(c, sp.A)`, `MemberDisplay` = `label(c, sp.B)`, dates from `sp.A.Date` / `sp.B.Date`, `IncomingNegative` = `sp.A.Negative`, `Weight` = `c.Points`, `Outcome` = `c.Result.Outcome`, `PropertyKey` = `string(c.Key)`, `PropertyOrigin` = `vocabkey.OriginProvenencia`; a comparison with a nil `sp` gets no line. The "already shown" check compares `ex.PropertyOrigin == vocabkey.OriginProvenencia && ex.PropertyKey == string(c.Key)`. Port `exhibits_pin_test.go` (kind `vocabkey.Event`; AC-4.2, AC-4.3) and add `TestExhibitWeightIsTheLeadWeight` (T-9). In `promotealign_test.go`, `match.Outcome*` become `propertycompare.Outcome*`, `ex.Property.Key` becomes `ex.PropertyKey`, `c.Property.Key` becomes `string(c.Key)` and `c.Outcome` becomes `c.Result.Outcome`; `comparisonOutcome` returns `propertycompare.Outcome`. AC-4.6's assertions join `TestProposePromoteGraphAlignment` (T-10). Then run the catalog suites and update expected values per D-16.
Check: `go build ./...` passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/ffi/handlers/` passes; `go test ./core/graphalign/...` passes. Commit STEP-2 – STEP-4 as one commit.

**STEP-5 — Delete the old matching code**
Files: `core/match/` (delete), `core/database/matching/` (delete), `core/database/subjectvocab/profile_test.go` (delete).
Do: D-15.
Check: `go build ./...` passes; `ls core/match core/database/matching` reports both missing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` passes; each grep below prints nothing (AC-5.2, AC-5.3; the pipes are plain here, unlike in a table):

```bash
rg -n 'provenencia/core/match"|core/database/matching"|match\.(Rank|Evaluate|DefaultProfile|Value|Property)\b' --type go
rg -n 'PropertyScale|ValueTypeScale|ScoreSimilarity|ApplyScale|resolvedScale|propertyPoints|SubjectsValues' --type go
```

**STEP-6 — The slice Check in the catalog**
Files: `core/database/promotealign/promotealign_test.go`.
Do: `TestProposeLifespanFromLifeEvents` (AC-3.1, AC-3.2) and `TestProposeKeepsTheLifespanOffTheSheet` (AC-4.1) (T-7, T-8). `birthEvent(f, src, art, year, polarity)` becomes `datedEvent(f *fixture, src sources.Source, art artifacts.Artifact, eventType string, year int, polarity string) subjects.Subject` (one helper, not two); existing callers pass `"birth"`. Participations in these tests use role `"subject"`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run 'Lifespan|QueryCount' -v` passes and lists both new tests and AC-3.3's.

**STEP-7 — Docs**
Files: `docs/matching.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `docs/ideas/README.md`, `docs/ideas/daitch-mokotoff.md`, `docs/ideas/promote-find-another-record.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan (status).
Do: apply each edit below. Text in a block is copied exactly; its links are relative to the file it goes in.
- `docs/matching.md`: the whole file becomes block DOC-1.
- `docs/promote-graph-alignment.md`:
  - Line 105 (the "Pairwise evaluation" row): its middle cell becomes block DOC-2.
  - Line 106 (the "Property-only matching" row): the whole row becomes block DOC-3.
  - Line 111: the paragraph becomes block DOC-4.
  - Line 138: in the paragraph, the words core/match `Rank` picks their candidates become `graphalign.Candidates` picks their candidates; a neighbor that never cleared `Rank` becomes a neighbor that was never a candidate; A candidate `Rank` recorded but did not accept becomes A candidate recorded but not accepted.
  - Lines 166–174, from "**Node comparisons.**" through the end of the "**Fixed scales**" bullet: replaced by block DOC-5. The `autoreconcile.Compatible` paragraph after it stays.
  - Line 288: the words over core/match.Evaluate (in backticks there) become over entity compare (`core/subjectcompare`).
  - Line 299: the bullet becomes block DOC-6.
  - Line 302: A `Rank` loser does not. becomes A candidate that was not accepted does not.; the walk or `Rank` already recorded becomes the walk or the fallback already recorded.
- `docs/deployment-plan/spike-11/comparison-layers.md`:
  - §Entities, after the "**Built in S11-11.**" paragraph: block DOC-7 as a new paragraph.
  - §Reuse across the app, consumers table, layer 4 row: Promote candidates, Subject suggestions (`matching.ForSubject`) becomes Promote candidates (S11-13).
  - Duplicates table: the row "match/pairs.go leaders and weighting" and the row "graphalign alikeNew, pairNode (property part)" each get " — done in S11-13" at the end of their second cell.
  - "Found in passing", last bullet: and `core/match` goes in S11-13 becomes and `core/match` went in S11-13.
- `docs/matching-model.md` §Removed knobs: block DOC-8 as a new row after the `graphalign` `PropertyAgreement` row.
- `docs/ideas/README.md` line 36: for a record `Rank` did not offer becomes for a record Promote did not offer.
- `docs/ideas/daitch-mokotoff.md` line 127: its first sentence becomes block DOC-9; "No UI, no migration, no product-version bump." stays after it.
- `docs/ideas/promote-find-another-record.md`: line 7, the records `Rank` offers becomes the records Promote offers, and so `Rank` offers no candidate becomes so Promote offers no candidate; line 11, one `Rank` did not offer becomes one Promote did not offer.
- `.agents/skills/add-match-rule/SKILL.md`, intro paragraph: block DOC-10 is inserted after the sentence that ends "holds every number.".
- `deployment-plan.md`: the S11-13 scope entry's "**Also (from S11-12 review).**" paragraph is already replaced by a "**Lifespan on the page.**" line, and the S11-20 scope entry already carries the lifespan line (both written with this plan). On landing: the map row status `Landed #N`.
- `completed.md`, newest first: block DOC-11 with the date and PR number filled in.
- This plan's Status `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)` and a **Landed** note under the header listing each difference from the plan (or "None.").

  DOC-1:

  ```text
  # Matching

  > **Spike 11 rebuilt matching.** The model, the terms and every knob are in [matching-model.md](matching-model.md). This page says where the running code is.

  Matching decides whether two things are the same Person, Event or Place. It is a stack of pure layers over one registry, [`core/matchregistry`](../core/matchregistry); the [`add-match-rule`](../.agents/skills/add-match-rule/SKILL.md) skill is how to change one.

  | Layer | Package | Says |
  | --- | --- | --- |
  | 1 Value compare | [`core/valuecompare`](../core/valuecompare) | How two values of one type relate |
  | 2 Property compare | [`core/propertycompare`](../core/propertycompare) | Whether two sides agree on one property |
  | 3 Subject compare | [`core/subjectcompare`](../core/subjectcompare) | How alike two Subjects of one type are: similarity and evidence |
  | 4 Entity compare | [`core/subjectprofile`](../core/subjectprofile) then layer 3 | Whether a Subject belongs with an entity's conclusion |

  ## Where it runs

  **Promote.** [`core/database/promotealign`](../core/database/promotealign) builds each side once per proposal: a Source Subject's with `subjectprofile.ForSubject` from its own Observations and the dated events it takes part in, a handle's with `subjectprofile.Build` from its reconciled values and its participations. [`core/graphalign`](../core/graphalign) scores every pair with `subjectcompare.Compare` and adds link support ([promote-graph-alignment.md](promote-graph-alignment.md)). A handle is a candidate when its own properties reach the weak band (`graphalign.Candidates`).

  **The reconciler** groups values with value compare under its own policy ([conclusion-reconciliation.md](conclusion-reconciliation.md)).

  **Search, the search index and term dedupe** key text with `valuecompare.TextKey`.

  ## Not built yet

  Segment compare (S11-14), the shortlist index (S11-15) and the layer 6 walk (S11-16). Merge hints are entity against entity, `subjectcompare.Compare` on two `Build` profiles; no surface uses them.
  ```

  DOC-2:

  ```text
  Given two sides, return entity compare's evidence and one outcome per lead property (`subjectcompare.Compare`). `Compatible` is sameness for reconciliation and pins, not the score
  ```

  DOC-3:

  ```text
  | **Candidates** ([`matching.md`](matching.md)) | List the handles a Subject is offered on its own properties: entity compare evidence at least the weak band (`graphalign.Candidates`) | Know Evidence bridges, fixed anchors, or propagation |
  ```

  DOC-4:

  ```text
  Alignment **orchestrates**; entity compare **judges a pair**. Alignment calls it for every pair it scores, adding edge terms when neighbors are already mapped. Candidates and the walk read the same evidence.
  ```

  DOC-5:

  ```text
  **Node comparisons.** The score does not name a property. Entity compare (`subjectcompare.Compare` on the two sides' profiles) runs layer 2 on each lead property in the registry's subject section and returns evidence in points: agree earns the agree weight, partial the agree weight × similarity, conflict loses the conflict penalty, unknown counts nothing. Weights are in the [knob inventory](matching-model.md#knob-inventory): name 2.5 / 3, sex at birth (with gender) 0.6 / 1, toponym 3 / 3, event type 0.5 / 20, date 2 / 3, a person's lifespan 0.1 / 20. A property the subject section does not list scores nothing. A neutral term is no evidence.
  ```

  DOC-6:

  ```text
  - **Properties compared (§6):** only the lead properties of the registry's subject section, with their pools and periods, not every Property both sides carry.
  ```

  DOC-7:

  ```text
  **Migrated in S11-13.** Promote scores every candidate with entity compare. `promotealign` builds each Source Subject's side with `ForSubject` from the Source's own Observations and life events, and each handle's with `Build` from the cache and its participations, once per proposal; `graphalign` runs `subjectcompare.Compare` for every pair it scores, the duplicate-New check included. A handle is a candidate when its own properties reach the weak band (`graphalign.Candidates`, user 2026-10-18). The lifespan counts but is not shown until S11-20. `core/match` and `core/database/matching` are deleted.
  ```

  DOC-8:

  ```text
  | `graphalign` `PropertyScale`, `ValueTypeScale` (weight and contradiction per property and per value type) | Promote scores properties with `subject.Weights` (S11-13); a property with no weight there scores nothing |
  ```

  DOC-9:

  ```text
  [`docs/matching-model.md`](../matching-model.md) then lists the tier's knob in the inventory.
  ```

  DOC-10:

  ```text
   Promote is the layers' caller: `core/database/promotealign` builds each side with the profile builder and `core/graphalign` scores pairs with `subjectcompare.Compare`; a new knob reaches Promote through the registry, with no Promote change.
  ```

  DOC-11:

  ```text
  ## S11-13 — Migrate Promote scoring onto subject compare

  **Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). Promote now judges a Source Subject against a handle with entity compare: each side is built once per proposal (the handle from its conclusion, displayed values positive and denied values negative; a person with the dated events they take part in), and every property scores its subject section weight. A birth is never filed on a death, a person who died before the Subject was born is never proposed, and sex at birth and gender compare as one. A handle is offered only when its own properties reach the weak band. Event types now earn 0.5 and cost 20. The lifespan counts but is not shown until S11-20. `core/match` and `core/database/matching` are gone.
  ```

Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-13-promote-subject-compare.md` reports no errors; `rg -n 'core/database/matching|match\.Rank|match\.Evaluate|core/match/|`Rank`' docs/matching.md docs/promote-graph-alignment.md docs/ideas/README.md docs/ideas/promote-find-another-record.md docs/ideas/daitch-mokotoff.md .agents/skills` prints nothing (`matching-model.md`'s glossary and Removed knobs, and `docs/ideas/matching-test-suite.md`, keep their historical mentions); every link in `docs/matching.md` resolves (`python3 scripts/check-plan-docs.py` does not cover it: open each target path with `ls`).

Commits: STEP-1, then STEP-2 – STEP-4 as one commit, then STEP-5, STEP-6 and STEP-7 each. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Score Promote candidates with entity compare."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/subjectcompare/subjectcompare_test.go` `TestPoints`, `TestDerived`, `TestLead` *(new)* | `Points` with `Weights{Agree: 2.5, Conflict: 3}`: agree 0.4 → 2.5; partial 0.75 → 1.875; conflict → −3; unknown → 0. `Derived` under `matchregistry.Default().Subject`: (person, lifespan), (person, lifespan_span) → true; (person, name), (event, lifespan) → false. `Lead` under `matchregistry.Default().Subject`: (person, name) → name; (person, gender) → sex_at_birth; (event, start_date) → date; (event, end_date) → date; (person, lifespan) → lifespan; (person, lifespan_span), (person, occupation), (event, name), (`"source"`, name) → not ok | D-12 |
| T-2 | Go | `core/graphalign/align_test.go`, every existing test ported; `TestSparseNameAndBirthDate`; `TestOneAgreeingPropertyStaysAWeakMatch` | AC-1.1, AC-1.9, AC-1.10 | AC-1.1, AC-1.9, AC-1.10, D-1, D-3, D-16 |
| T-3 | Go | `align_test.go` `TestBirthIsNeverProposedOnADeath` *(new)*, subtests `AC-1.2`, `AC-1.3`, `AC-1.8` | As in §2 | AC-1.2, AC-1.3, AC-1.8 |
| T-4 | Go | `align_test.go` `TestLifespanRulesOutAPerson` *(new)*, subtests `AC-1.4`, `AC-1.5`; helper `withLife` *(new)* | As in §2 | AC-1.4, AC-1.5, D-6, D-14 |
| T-5 | Go | `align_test.go` `TestGenderPoolsWithSex`, `TestEventPeriodStandsForTheDate` *(new)* | AC-1.6; AC-1.7 | AC-1.6, AC-1.7 |
| T-6 | Go | `core/graphalign/candidates_test.go` *(new)* `TestCandidates`, external package `graphalign_test` | AC-2.1 – AC-2.4 | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-4 |
| T-7 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeLifespanFromLifeEvents` *(new)*, subtests `AC-3.1`, `AC-3.2` | As in §2 | AC-3.1, AC-3.2, D-8, D-9, D-10 |
| T-8 | Go | `promotealign_test.go` `TestProposeKeepsTheLifespanOffTheSheet` *(new)*; `TestProposeQueryCountDoesNotGrowWithTheLayer`; `TestProposeRowWithoutExhibitsShowsValues` | AC-4.1; AC-3.3; AC-4.5 | AC-3.3, AC-4.1, AC-4.5, D-6, D-13 |
| T-9 | Go | `exhibits_pin_test.go` `TestOverlappingBoundsAreSimilarAndUnpinned`, `TestDisjointDatesConflict` (ported, weight −3 added), `TestExhibitWeightIsTheLeadWeight` *(new)* | AC-4.2 – AC-4.4 | AC-4.2, AC-4.3, AC-4.4, D-11 |
| T-10 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposePromoteGraphAlignment` | AC-4.6 | AC-4.6, D-7 |
| T-11 | Go | `go build ./...` and the greps of AC-5.1 – AC-5.3; `core/matchregistry` `TestLayerPackages`, `TestKnobInventoryMatchesRegistry`; `subjectvocab` `TestSubjectSectionMatchesSeed` | Builds; greps print nothing; guards pass | AC-5.1, AC-5.2, AC-5.3, D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` tests that compute an event's score from the plain scale, `core/database/promotealign` tests that count comparisons or read a `gender` comparison, `TestProposeQueryCountDoesNotGrowWithTheLayer` (handle life events), `api/ffi/handlers` promote tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching.md, promote-graph-alignment.md, comparison-layers.md, matching-model.md, ideas/README.md, ideas/daitch-mokotoff.md, add-match-rule skill (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has a person "James Robins" who is the subject of a death event dated 1840, promoted (person and event filed).
2. On a new Source, add a person "James Robins" and a birth event dated 1850 with him as subject. Open Promote for that Source. His row proposes New and the target menu offers no James Robins handle.
3. Change the first Source's death to 1920 (re-promote if the handle does not update) and reopen Promote on the second Source. His row proposes the James Robins handle as a weak match, and the evidence sheet shows a name line and no lifespan line.
4. On the second Source, add an event of type death dated 1850 with no participants, and on the first Source a filed event of type birth dated 1850. The second Source's death row never proposes the birth handle.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Handle life events add a read per person | `profiler.profile` prefetches every participation neighbor with one `g.Nodes` call (D-10); AC-3.3 fails if a read grows |
| A golden changes its target because of a cause §1 does not list | D-16 names the tests that keep their targets; any other change is a stop condition |
| The candidate gate drops a handle a catalog test expects as an alternative | D-4 is the user's call; the test is updated when that alternative came from seeding or the fallback and scores under 2 on its own properties, and the update is listed in the Landed note |
| The Source graph lacks an Observation the SQL read had | `loadSourceObservations` reads every Observation of the Source's Subjects, both polarities; a catalog test that loses a value is a stop condition |

**Stop and report if:**

- `main` differs from §4: `subjectcompare.Compare`'s signature, `subjectprofile.Build` / `ForSubject` signatures, `graphalign.Config` fields, `loadCanon` / `loadLayer` / `pairExhibits` signatures, `graphcache.SourceObservation` fields, or `Graph.Follow` / `Graph.Nodes`. Report each difference.
- `rg -n 'q\.' core/database/promotealign/canon.go` prints a hit after the seed loop moves (the canon loader still reads SQL).
- A test D-16 names changes a Target or handle it asserts, or AC-1.1's scores differ from 5.25.
- `TestProposeQueryCountDoesNotGrowWithTheLayer` fails after STEP-3: report both counts and the reads that grew.
- `TestProposeRowWithoutExhibitsShowsValues` shows a term key where it expects a label.
- `graphalign` or `promotealign` needs a `match` symbol after STEP-4, or a non-test file outside §6's packages imports `core/match` or `core/database/matching`.
- A layer package fails `TestLayerPackages` (a numeric literal other than 0 and 1, or a `core/database` import).
- AC-2.1's precondition fails (PER-2's evidence is 2 or more, or PER-4's is not negative): report both numbers.

## 12. Open questions

None.

## 13. PR

**Title:** S11-13 — Migrate Promote scoring onto subject compare
**Body:** objective in one sentence; what shipped against each US; the researcher-visible changes of §1; AC evidence (test names); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-13-promote-subject-compare.md`.
