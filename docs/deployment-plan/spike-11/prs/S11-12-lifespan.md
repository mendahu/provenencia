# S11-12 — Lifespan for people

**Status:** Landed in [#480](https://github.com/mendahu/provenencia/pull/480) (2026-10-18)
**Landed:** `TestAddYears`'s leap-day row uses an offset of 110 (the plan's step text said 10, which cannot give 2014). STEP-4 and STEP-5 are one commit, because the new `Build` and `ForSubject` signatures touch the entity compare and invariant tests. `TestLifespanFromEvents` AC-2.11 is one row holding both undated events. Otherwise none.
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: pure Go layers with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-11 (landed [#476](https://github.com/mendahu/provenencia/pull/476); [plan](S11-11-subject-profile.md) and its **Landed** note: STEP-2 and STEP-3 were one commit; expected strings use `Labels`' `key={a, b}` form). Branch from `main`
**Blocks:** S11-13
**Branch:** `s11-12-lifespan`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a knob", "Add a Subject property or pair"), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md) (calendar arithmetic in `core/datevalue`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Lifespan](../comparison-layers.md#lifespan-a-derived-property-for-people) (the user's calls of 2026-10-18 are already written there and in the map's working rule "Lifespan is for people only" and S11-12 scope entry); [Layer 3](../comparison-layers.md#layer-3-subject-comparison) ("Built in S11-10"); [Entities](../comparison-layers.md#entities-compare-against-the-conclusion) ("Built in S11-11"); [`docs/matching-model.md`](../../../matching-model.md) worked example step 3, glossary **lifespan**, knob inventory lifespan rows; S11-10 [ship log](S11-10-subject-compare.ship.md) follow-ups MNT-2 and MNT-3; working rules "Sparse data first", "Registry numbers: carry over, then estimate", "One match registry", "Pure means the package builds and tests with no build tags and no `core/database` import", "One matching skill, grown with the layers"
**Planned:** 2026-10-18 against `main` @ `b799b67` (S11-11 merged)
**Review:** [`S11-12-lifespan.review.md`](S11-12-lifespan.review.md) — Fits / Ready, 2 rounds (findings fixed)

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

A person gets a lifespan, and Subject compare uses it to rule out impossible matches. The profile builder (`core/subjectprofile`) lays every dated event the person takes part in on one timeline. A birth starts the lifespan and a death or burial ends it; an end with no such event is padded by `MaxYears` (110) from the first or last dated event. Each side carries two derived values: its **window** (where its life can lie) and its **span** (its first to last dated event). Subject compare (`core/subjectcompare`) checks each side's span against the other side's window: a span `Different` from the window (outside it, beyond the 2-year date tolerance) is a conflict (−20); otherwise the lifespans agree (+0.1).

**Why now.** S11-13 moves Promote's candidate scoring onto entity compare. With the lifespan in place first, "a person who died before the Subject was born is never proposed" (goal 2) holds the day Promote moves.

**What stays.** `core/match`, `core/database/matching`, `core/autoreconcile` and every caller are untouched; no caller passes life events yet (S11-13's loader does). No catalog, wire, Swift or researcher-visible change. Events and places get no lifespan.

## 2. User stories

Conventions. `L(w1–w2 | s1–s2)` is a person's two derived values at year precision: `lifespan` = `datevaluetest.Between(w1, 0, 0, w2, 0, 0)` and `lifespan_span` = `datevaluetest.Between(s1, 0, 0, s2, 0, 0)`, each a positive `propertycompare.Value` with only `Date` set. A life event is written `type·role year`: `subjectprofile.LifeEvent{EventType: vocabkey.EventType(type), Role: vocabkey.Role(role), Dates: []datevalue.Value{*datevaluetest.YMD(year, 0, 0)}}`. "sim" and "ev" are `subjectcompare.Result.Similarity` and `.Evidence` (within 0.0001) under `matchregistry.Default()`, with metas `subjectcomparetest.Metas()`. Expected labels are `subjectcomparetest.Labels` strings; a year-precision range labels as `Between 1850 and 1960`. Every relation below was read from `propertycompare.Compare` at `b799b67` with a date meta and the default value rules: window 1730–1840 against span 1850 is `different`; window 1850–1960 against span 1990 is `different`; window 1850–1960 against span 1920, and span 1850 against window 1810–1920, are `contains` (partial); span 1950 against window 1790–1900 is `different`; span 1900 against window 1840–2060 is `contains`; window 1850–1960 against span 1962 is `near` (partial, 0.0667).

**US-1.** As the engine, when I compare two people, a lifespan that cannot hold together is a large conflict, and any other pair of lifespans agrees for almost nothing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person: a `L(1730–1840 \| 1840–1840)` (died 1840) against b `L(1850–1960 \| 1850–1850)` (born 1850) | `lifespan` conflict, cause `relation`, points −20; sim 0, ev −20 |
| AC-1.2 | a `L(1850–1960 \| 1850–1850)` (born 1850) against b `L(1880–1990 \| 1990–1990)` (died 1990) | `lifespan` conflict; sim 0, ev −20 |
| AC-1.3 | a `L(1850–1960 \| 1850–1850)` against b `L(1810–1920 \| 1920–1920)` (died 1920) | `lifespan` agree, points 0.1; sim 1, ev 0.1 |
| AC-1.4 | a `L(1840–2060 \| 1950–1950)` (resided 1950) against b `L(1790–1900 \| 1900–1900)` (died 1900) | `lifespan` conflict; sim 0, ev −20 |
| AC-1.5 | a and b both `L(1850–1960 \| 1850–1850)` | agree; sim 1, ev 0.1 |
| AC-1.6 | a `L(1850–1960 \| 1850–1850)`, b with no lifespan values | `lifespan` unknown, cause `no_values`, points 0; sim 0, ev 0 |
| AC-1.7 | a `L(1850–1960 \| 1850–1850)` against b `L(1852–1962 \| 1962–1962)` (two years past the window: `near`) | agree; sim 1, ev 0.1 |
| AC-1.8 | AC-1.1's values plus `name` `{James Robins}` on both sides | sim 0.1111, ev −17.5 |
| AC-1.9 | Event kind: AC-1.1's lifespan values plus `event_type` `{birth}` on both sides | Result keys `[date event_type]` (no `lifespan`); sim 1, ev 0.5 |
| AC-1.10 | Person, both sides empty (`TestCompare`'s existing "both sides empty" row) | Result keys `[lifespan name sex_at_birth]`; `lifespan` unknown, cause `no_values`; sim 0, ev 0 as before |
| AC-1.11 | a `L(1850–1960 \| 1850–1850)` against b with only `lifespan_span` `{Between 1850 and 1850}` (no window) | `lifespan` unknown (a's span has no window to meet); sim 0, ev 0 |

**US-2.** As the engine (S11-13's loader), when I build a person's side from its life events, I get its window and its span, or nothing when the side contradicts itself.

Each row calls `subjectprofile.ForSubject(subjectprofile.Subject{Life: events}, matchregistry.Default().Subject.Lifespans[vocabkey.Person])` and compares `subjectcomparetest.Labels` of the result.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `birth·subject 1850` | `lifespan={Between 1850 and 1960}; lifespan_span={Between 1850 and 1850}` |
| AC-2.2 | `death·subject 1840` | `lifespan={Between 1730 and 1840}; lifespan_span={Between 1840 and 1840}` |
| AC-2.3 | `birth·subject 1850`, `death·subject 1920` | `lifespan={Between 1850 and 1920}; lifespan_span={Between 1850 and 1920}` |
| AC-2.4 | `residence·subject 1950` | `lifespan={Between 1840 and 2060}; lifespan_span={Between 1950 and 1950}` |
| AC-2.5 | `baptism·subject 1830`, `burial·subject 1901` | `lifespan={Between 1720 and 1901}; lifespan_span={Between 1830 and 1901}` |
| AC-2.6 | `birth·father 1880` (a child's birth) | `lifespan={Between 1770 and 1990}; lifespan_span={Between 1880 and 1880}` |
| AC-2.7 | `residence·subject 1900`, `birth·subject 1850`, `marriage·spouse 1875` (that order) | `lifespan={Between 1850 and 2010}; lifespan_span={Between 1850 and 1900}` |
| AC-2.8 | `birth·subject 1852`, `birth·subject 1850` | `lifespan={Between 1850 and 1962}; lifespan_span={Between 1850 and 1852}` |
| AC-2.9 | `death·subject` with date `BEF 1840` (`datevalue.Value{Kind: datevalue.KindPoint, Qualifier: datevalue.QualifierBEF, StartYear: datevaluetest.Part(1840)}`) | Same labels as AC-2.2 |
| AC-2.10 | `birth·subject` on `*datevaluetest.YMD(1904, 2, 29)` | One `lifespan` value; `valuecompare.DateEnds` of its date gives first end `(1904, 2, 29)` and last end `(2014, 2, 28)` |
| AC-2.11 | One `residence·subject` with `Dates` nil, and one whose date is `*datevaluetest.Phrase("spring")` | Non-nil, length 0 |
| AC-2.12 | `birth·subject 1900`, `death·subject 1850` | Non-nil, length 0 (the window would end before it starts) |
| AC-2.13 | AC-2.1's events with a nil lifespan map | Non-nil, length 0 |
| AC-2.14 | `Build` of S11-11's `maleFemale(t)` entity with `Life` = `[birth·subject 1850]` | `lifespan={Between 1850 and 1960}; lifespan_span={Between 1850 and 1850}; sex_at_birth={male}` |
| AC-2.19 | AC-2.14's entity with `Options{Excluding: [][]byte{mem(1)}}` | `lifespan={Between 1850 and 1960}; lifespan_span={Between 1850 and 1850}; sex_at_birth={male, female}` (the lifespan is unchanged) |
| AC-2.15 | `Subject{Values: sex_at_birth male, Life: AC-2.1's events}` | `lifespan={Between 1850 and 1960}; lifespan_span={Between 1850 and 1850}; sex_at_birth={male}` |
| AC-2.16 | `birth·subject 1900`, `residence·subject 1850` | Non-nil, length 0 (the residence lies wholly before the window) |
| AC-2.17 | `death·subject 1850`, `residence·subject 1900` | Non-nil, length 0 (the residence lies wholly after the window) |
| AC-2.18 | `birth·subject` on `*datevaluetest.YMD(1850, 5, 1)`, `baptism·subject 1850` (year only) | Built: `DateEnds` of the window's date gives first end `(1850, 5, 1)` and last end `(1960, 0, 0)`; the span's first end is the year 1850 (a same-year event at a coarser precision is not a contradiction) |

**US-3.** As the engine, the map's Check holds for entity compare: `Build` then `subjectcompare.Compare` rules out a person who cannot be the entity. In each row side A is `Build` of an entity whose `name` is `property(t, vocabkey.PropertyName, cand(1, 1, 1, name("James Robins")))` and whose `Life` is given, and side B is `ForSubject(Subject{Values: {name: [vcand(name("James Robins"), false)]}, Life: given})`; both with `reg.Subject.Lifespans[vocabkey.Person]`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.10 | Entity `death·subject 1840`; Subject `birth·subject 1850` | `lifespan` conflict; sim 0.1111, ev −17.5 |
| AC-3.11 | Entity `birth·subject 1850`; Subject `death·subject 1990` | `lifespan` conflict; sim 0.1111, ev −17.5 |
| AC-3.12 | Entity `birth·subject 1850`; Subject `death·subject 1920` | `lifespan` agree; sim 1, ev 2.6 |
| AC-3.13 | Entity `death·subject 1900`; Subject `residence·subject 1950` | `lifespan` conflict; sim 0.1111, ev −17.5 |
| AC-3.14 | Both `birth·subject` on `*datevaluetest.YMD(1901, 5, 1)` (the worked example's Gracie) | `lifespan` agree; sim 1, ev 2.6 |

(US-3's ACs start at 3.10 because `TestEntityCompare` already holds S11-11's nine US-3 rows.)

**US-4.** As a contributor, the invariants cover every lead in the registry, the lifespan included, so a lead added later cannot be skipped silently (S11-10 MNT-2).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.5 | `core/subjectcompare` `TestInvariants` | Its lead groups come from `matchregistry.Default().Subject`: the leads of `Weights[kind]`, each lead's `Pools` `With`, its `Periods` `Start` / `End`, and its `Lifespans` `Span`. A lead with no `propertyPools` entry fails the test by name |
| AC-4.6 | Its six existing invariant subtests, for person | Run with the lifespan group and pass |
| AC-4.7 | `core/subjectprofile` `TestInvariants`, new subtest: generated life events | When a lifespan is built, its window (`valuecompare.DateWindow` of the `lifespan` date) overlaps `valuecompare.DateWindow` of every dated event's date; the labels are equal after shuffling the events |

(US-4's ACs start at 4.5 because `core/subjectprofile` `TestInvariants` already holds S11-11's four subtests.)

## 3. Scope

| In | Covered by |
| --- | --- |
| `datevalue.AddYears`; `valuecompare.ShiftYears`, `Earliest`, `Latest` (D-6) | STEP-1 |
| `subjectcompare`'s `period` picks its ends with `Earliest` / `Latest` (D-7, MNT-3) | STEP-1 |
| `vocabkey` term key types, term and derived key constants; their seed check (D-3) | STEP-2 |
| `subjectcompare.Lifespan`, `EventRole`, `Rules.Lifespans`; the lifespan special case in `Compare` (D-4, D-5) | STEP-2, US-1 |
| Registry: `subject.Lifespans` and `subject.Weights.person.lifespan`; inventory rows; `notYetBuilt`; registry and seed tests (D-8, D-9) | STEP-2 |
| Lead groups built from the registry (D-10, MNT-2) | STEP-3, US-4 |
| `subjectprofile.LifeEvent`, `Subject`, `Entity.Life`, the lifespan argument of `Build` and `ForSubject`, the builder (D-1, D-11, D-12) | STEP-4, US-2 |
| Entity compare and builder invariant tests (D-13) | STEP-5, US-3, US-4 |
| Docs the build makes true: `comparison-layers.md` "Built in" paragraph, `matching-model.md`, the `add-match-rule` skill, code docs (D-14) | STEP-2, STEP-6 |

| Out | Home |
| --- | --- |
| Loading a person's life events from the catalog (participations, event types, roles, the event's displayed dates or period) and passing them to `Build` / `ForSubject` | S11-13 (the loader that feeds `propertycompare` and `subjectprofile`) |
| Lifespans for events and places | Not planned — working rule "Lifespan is for people only" |
| Age rules (marriage at 14 or older, parents 12–60 at a child's birth, mothers narrower) | Not planned — dropped by the user's call of 2026-10-18 (D-1); their four knobs leave the inventory |
| Two-hop rules (a child's birth read through a parent Relationship) | Not planned in Spike 11 — spike README "Not this spike" names two-hop lifespan rules |
| `Options.Excluding` changing the lifespan (a left-out member's participations) | [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) — the weak-claim review is the first caller of `Excluding` (D-12) |
| Moving `"subject"` / `"birth"` / `"death"` literals in `core/hops/hops.go`, `core/database/conclusionheaders/compose.go` and `core/graphalign/types.go` `EdgeSignature` onto the typed `vocabkey` constants | S11-14 (typed bridge keys) for `EdgeSignature`; the others when next edited — not matching layers |
| A conflict for one side's two different death dates | S11-14 — a death is a one-cardinality segment ([segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer 3 | `core/subjectcompare/subjectcompare.go`: `Weights`, `Pool`, `Period`, `Rules{Weights, Pools, Periods}`, `Values` (doc "Product properties only."), `Metas`, `Result`, `PropertyResult`, `Compare` | `Compare` runs `propertycompare.Compare(side(kind, key, a, rules), side(kind, key, b, rules), metas[key], value)` for each lead of `rules.Weights[kind]` in key order. No lifespan |
| Period building | `core/subjectcompare/pairs.go` `side` (reads `rules.Periods[kind][lead]`), `period` | `period` has two near-identical loops: the earliest start (lowest `Lo`, then lowest `Hi`) and the latest end (highest `Hi`, then highest `Lo`), skipping non-positive values, nil dates, failed `DateEnds` and open ends (S11-10 MNT-3) |
| Layer 3 tests | `core/subjectcompare/subjectcompare_test.go` `TestCompare` row "AC-1.12 both sides empty" whose check reads `slices.Equal(keys(r), []vocabkey.PropertyKey{name, sex})`; `core/subjectcompare/invariants_test.go` `group{lead, props, same}`, `kindGroups`, `propertyPools`, `properties`, `draw`, `empty`, `TestInvariants` (six subtests) | `kindGroups` repeats the registry by hand (S11-10 MNT-2); each group's `same` value equals `propertyPools[lead][0]`; AC-3.5 sets `g.lead` to `g.same` on both sides |
| Layer 3 fixtures | `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` `Metas()`, `Labels` | `Metas()` holds seeded properties only; `TestSubjectSectionMatchesSeed` fails on a fixture key that is not seeded |
| Registry | `core/matchregistry/subject.go` `defaultSubject()`; `core/matchregistry/registry.go` `Registry{Value, Subject}` | No lifespan entry |
| Registry guards | `core/matchregistry/inventory_test.go` `notYetBuilt` (`"subject.Weights.person.lifespan."`, `"subject.person.lifespan."`, `"segment."`, `"alignment."`, `"evidence."`), `registryKeys` (walks maps whose element is a struct or map by sorted key; a slice field is one leaf); `core/matchregistry/subject_test.go` `TestSubjectSection`; `core/matchregistry/registry_test.go` `TestDefaultIsFresh`; `core/matchregistry/guard_test.go` `layerPackages` | The guard already lists `core/datevalue`, `core/vocabkey`, `core/valuecompare`, `core/subjectcompare` and `core/subjectprofile` |
| Seed checks | `core/database/subjectvocab/vocabkey_test.go` `TestVocabKeysMatchSeed` (vocabkey kinds and Property keys against the seed); `core/database/subjectvocab/subject_section_test.go` `TestSubjectSectionMatchesSeed` (every `Weights` key bound by the seed); `core/database/subjectvocab/registry.go` `seedTerms` (`event_type` `birth`, `death`, `burial`, `baptism`, `marriage`, `residence`; `role` `subject`, `father`, `mother`, `spouse`, `child`) | A `Weights` entry for `lifespan` fails `TestSubjectSectionMatchesSeed` today; no term key is checked |
| Calendar | `core/datevalue/calendar.go` `DaysFromCivil`, `DaysInMonth`, `BoundDays`, `AddMonths`, `AddDays`; `core/datevalue/calendar_test.go` | No year move |
| Date ends | `core/valuecompare/datewindow.go` (header "Pure calendar arithmetic lives in core/datevalue") `DateEnd{Year, Month, Day, Lo, Hi, From, To, Mid, Open}`, `DateEnds`, `EndsDate`, `boundEnd` (unexported), `Window` (`Overlaps`, `Contains`), `EndsWindow`, `DateWindow`; `core/valuecompare/datewindow_test.go` | No way to move a `DateEnd` by years or pick the earliest / latest of several |
| Date compare | `core/valuecompare/date.go` `date.compare` | Spans apart by more than `Dates.Tolerance` (2) civil years are `Different`; within it `Near` |
| Profile builder | `core/subjectprofile/subjectprofile.go` `Entity{Properties, NamePattern}`, `Property`, `Candidate`, `Options`, `Build(e, metas, reconcile, opts)`, `without`, `ForSubject(values, namePattern)`, `value` | No life events. Package doc: "Pure: no I/O, no catalog, no numbers" |
| Profile tests | `core/subjectprofile/helpers_test.go` (`obs`, `src`, `mem`, `term`, `text`, `year` (an `autoreconcile.Value`), `name`, `cand`, `weak`, `provisional`, `negative`, `reconcile`, `property`, `vcand`, `maleFemale`, `denied`); `subjectprofile_test.go` `TestBuild`, `TestForSubject`; `entity_test.go` `TestEntityCompare` (rows up to S11-11's last US-3 AC; fields `kind`, `entity func(t *testing.T) subjectprofile.Entity`, `opts`, `subject`, `sim`, `ev`, `check`); `invariants_test.go` `TestInvariants` (four subtests) | Six `Build` / `ForSubject` calls (`rg -n 'subjectprofile\.(Build\|ForSubject)\(' core/subjectprofile`) |
| Keys | `core/vocabkey/vocabkey.go` package doc (origins, kinds, seeded Property keys), `SubjectType`, `PropertyKey`, `Person`, `PropertyName` … `PropertyLikeness` | No derived key, no term key type |
| Participation keys today | `core/graphalign/types.go` `EdgeSignature{BridgeType, RoleOrType, NeighborKind, NeighborTypeTerm}` and `Key()` (pipe-joined); `core/hops/hops.go` line 60 (`TermKey: "subject"`) | Untyped strings; S11-14 replaces `EdgeSignature` with a typed `BridgeKey` |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §Lifespan (already says the 2026-10-18 rule); `docs/matching-model.md` §Worked example step 3, §Glossary **lifespan**, §Knob inventory rows `subject.Weights.person.lifespan.Agree` / `.Conflict` and five `subject.person.lifespan.*` rows; `.agents/skills/add-match-rule/SKILL.md` Layer 3 bullet ("Value type and cardinality come from the catalog as `Metas`, never from the registry"), profile builder bullet ("The builder has no registry section."), "Add a Subject property or pair" | Inventory and skill written before the 2026-10-18 call |

**Precedent.** S11-10 (`core/subjectcompare`, [#471](https://github.com/mendahu/provenencia/pull/471)) added `Periods`, a registry-declared special case keyed `map[SubjectType]map[PropertyKey]Period`: copy how it is declared in `Rules`, read by `side`, keyed in the registry, checked in `TestSubjectSection` and `TestSubjectSectionMatchesSeed`, and given inventory rows `subject.Periods.event.date.*`. S11-11 ([#476](https://github.com/mendahu/provenencia/pull/476)): copy its table tests and fixed-seed `TestInvariants`. `valuecompare.EndsDate` (S11-10 D-8) is the one place date ends become a date again.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The lifespan is a timeline (user, 2026-10-18).** Every dated event a person takes part in counts, any event type, any role. The **span** is the first to the last dated event. The **window** starts at the earliest dated `Starts` event (birth, the person's own) or, with none, `MaxYears` before the first dated event; it ends at the latest dated `Ends` event (death or burial, the person's own) or, with none, `MaxYears` after the last dated event | The user asked for "a really straightforward" guide, not an authority: one math function over a timeline. It replaces per-bridge age rules and their four knobs | Per-bridge age rules; a birth window only — misses resided-after-death |
| D-2 | **Span in window (user, 2026-10-18).** Two lifespans conflict when either side's span is `Different` from the other side's window; otherwise they agree. `Near` (within `Dates.Tolerance`, 2 years) and `Overlaps` are not conflicts | Catches died-before-born, the 140-year span and resided-after-death (AC-1.1, AC-1.2, AC-1.4); window against window catches only the first | Window against window; a merged span over `MaxYears` |
| D-3 | **Keys.** `vocabkey` gains `type EventType string` ("an `event_type` term key, typed for registry rules") and `type Role string` ("a participation `role` term key, typed for registry rules"); in a block "Seeded term keys matching reads by name": `EventBirth EventType = "birth"`, `EventDeath EventType = "death"`, `EventBurial EventType = "burial"`, `RoleSubject Role = "subject"`; in a block "Derived Property keys the profile builder adds (never seeded)": `PropertyLifespan = "lifespan"`, `PropertyLifespanSpan = "lifespan_span"`. The package doc's first sentence gains ", the seeded term keys it reads, and the derived Property keys matching builds". `TestVocabKeysMatchSeed` checks each term constant against `seedTerms` under `event_type` (the three `EventType`s) or `role` (`RoleSubject`); derived keys are not in it | Typed keys make a swapped `{RoleSubject, EventBirth}` a compile error, and one key home serves S11-14's `BridgeKey` (Next adds). `TestVocabKeysMatchSeed` is where a vocabkey constant is checked against the seed | Untyped strings — the registry would accept a role as an event type; `subjectcompare`-local key types — the skill: "no other package declares a key type" |
| D-4 | **Shape in layer 3, like `Periods`.** `subjectcompare.EventRole` *(new)*: `{EventType vocabkey.EventType; Role vocabkey.Role}`. `subjectcompare.Lifespan` *(new)*: `{Span vocabkey.PropertyKey; MaxYears int; Starts, Ends []EventRole}`. `Rules.Lifespans map[vocabkey.SubjectType]map[vocabkey.PropertyKey]Lifespan` *(new field, after `Periods`)*, doc "Lifespans lists, per Subject type and lead property, the derived lifespan whose window is the lead and whose span is Span; the profile builder builds both from life events." The lead (the window) is in `Weights[kind]`; `Span` is not a lead. `MaxYears`, `Starts` and `Ends` are read only by the builder; the type lives in layer 3 because the subject section holds "lifespan rules and constants" | One pattern for a lead with its own rule: `Pools`, `Periods`, `Lifespans`. The inventory keeps its documented `section.Group.<kind>.<lead>.Field` path, with no rows for non-tunables | One struct with `Kind` and `Window` fields — a third shape and two inventory rows that are not knobs; a `profile` registry section — the design lists five sections |
| D-5 | **The special case in `Compare`.** In the loop, `if l, ok := rules.Lifespans[kind][key]; ok`, the lead's result is `lifespan(key, a, b, l, value)` *(new, unexported, in `core/subjectcompare/lifespan.go` (new))*; otherwise it is today's `propertycompare.Compare` call. Scoring is unchanged. `lifespan` runs `ab := propertycompare.Compare(a[window], b[l.Span], dateMeta, value)` and `ba := propertycompare.Compare(a[l.Span], b[window], dateMeta, value)`, with package-level `var dateMeta = propertymeta.Meta{ValueType: propertymeta.ValueTypeDate, Cardinality: propertymeta.CardinalitySingle}`, and returns: `ab` when `ab.Outcome` is conflict; else `ba` when it is conflict; else `ab` when `ab.Outcome` is unknown; else `ba` when it is unknown; else `ab` with `Outcome` set to `propertycompare.OutcomeAgree`. Its doc comment says: "A lifespan's agree means consistent: its speaking pair keeps the relation that decided it, which may be partial (`contains`, `near`)." | Layer 2 already compares date sets with negatives and the tolerance; only the pairing and "consistent is agree" are lifespan's own. A derived key has no catalog meta, so the fixed date meta sits beside the rule that reads it. Agree keeps the worked example's 0.1 | `metas[lead]` from the caller — S11-13 would forge a catalog row; a partial scaled by the window's span similarity (0.2) — the worked example changes for a near-zero weight; a new `Cause` — one reader (S11-19) and the doc comment serves it |
| D-6 | **Date helpers.** `datevalue.AddYears(year, month, day, n int) (int, int, int)` *(new, `core/datevalue/calendar.go`)*, doc "AddYears moves a date by n civil years. A month or day of 0 (not given) stays 0; 29 February in a common year becomes 28 February.": returns `year+n`, `month`, and `day` clamped to `DaysInMonth(year+n, month)` when both are non-zero. In `core/valuecompare/datewindow.go`: `func ShiftYears(e DateEnd, years int) (DateEnd, bool)` *(new)*, doc "ShiftYears moves a closed end by whole civil years (datevalue.AddYears). An open end is returned as is. ok is false when the moved end cannot be placed.", which calls `AddYears` then `boundEnd(&y, endPart(m), endPart(d))` (`endPart` gives nil for 0); `func Earliest(ends []DateEnd) (DateEnd, bool)` *(new)*, doc "Earliest is the closed end that starts first: lowest Lo, then lowest Hi. Open ends are skipped; ok is false when none is closed."; `func Latest(ends []DateEnd) (DateEnd, bool)` *(new)*, "the closed end that finishes last: highest Hi, then highest Lo" | Calendar moves live in `core/datevalue` (the file header says so); `boundEnd` is unexported; one earliest / latest pick for `period` and the builder, typed on `DateEnd` | A day clamp in `valuecompare` — calendar arithmetic outside its home; a `pickEnd` over `propertycompare.Value` — the builder cannot use it; exported order methods plus a loop per caller — two loops again |
| D-7 | **MNT-3.** `period` collects, from the positive values with a `Date` whose `DateEnds` is ok, the first ends of `starts` and the last ends of `ends`, then takes `Earliest` and `Latest` of them; the rest of `period` is unchanged | The map's scope entry; same tie rules and skips | — |
| D-8 | **Registry.** `defaultSubject()` adds `Weights[vocabkey.Person][vocabkey.PropertyLifespan] = {Agree: 0.1, Conflict: 20}` (comment "estimate: near zero, the dates count through their segments; impossible together never recovers") and `Lifespans: {vocabkey.Person: {vocabkey.PropertyLifespan: {Span: vocabkey.PropertyLifespanSpan, MaxYears: 110, Starts: [{EventType: vocabkey.EventBirth, Role: vocabkey.RoleSubject}], Ends: [{EventType: vocabkey.EventDeath, Role: vocabkey.RoleSubject}, {EventType: vocabkey.EventBurial, Role: vocabkey.RoleSubject}]}}}` (comment on `MaxYears`: "estimate: longest plausible human life"). Every `EventRole` literal, in the registry and in tests, is keyed (`go vet`'s composites check flags unkeyed fields of an imported struct). Inventory: the two `subject.Weights.person.lifespan.*` rows stay; the five `subject.person.lifespan.*` rows go; four rows *(new)* `subject.Lifespans.person.lifespan.Span`, `.MaxYears`, `.Starts`, `.Ends` (STEP-2). `notYetBuilt` loses both lifespan prefixes | Numbers carried from the inventory's estimates (working rule); `registryKeys` walks the maps and treats a slice as one leaf | — |
| D-9 | **Registry tests.** `TestSubjectSection` gains a loop over `s.Lifespans`: the lead has a `Weights[kind]` entry and no pool or period; `claim(kind, lead, "lifespan lead")`; `Span` has no `Weights` entry and `claim(kind, l.Span, "lifespan span")`; `MaxYears > 0`; `Starts` and `Ends` are not empty. `TestDefaultIsFresh` also sets `first.Subject.Lifespans[vocabkey.Person][vocabkey.PropertyLifespan].Starts[0] = subjectcompare.EventRole{}`. `TestSubjectSectionMatchesSeed`: its `Weights` loop skips a key that is a `s.Lifespans[kind]` lead (derived, never bound), and a new loop over `s.Lifespans` calls `checkKind(kind)` | Same checks `Periods` gets; the seed check of the term keys is D-3's | — |
| D-10 | **MNT-2: lead groups from the registry.** In `invariants_test.go`, `kindGroups` and `group.same` go. `groups(t *testing.T, kind vocabkey.SubjectType) []group` *(new)* builds, from `matchregistry.Default().Subject`, one `group{lead, props}` per `Weights[kind]` lead in key order, with `props` the lead, then its pool's `With`, its period's `Start` and `End`, and its lifespan's `Span`. Each group's test value is `propertyPools[lead][0]`; a lead or prop with no `propertyPools` entry fails with `t.Fatalf("%s.%s has no propertyPools entry", kind, p)`. `TestInvariants` builds `byKind := map[vocabkey.SubjectType][]group{}` once, with `groups(t, kind)` for each kind; the `invariant` type becomes `func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (a, b Vals, msg string, ok bool)`, `draw(r, gs)` and `properties(gs)` take the groups, and the AC-3.4 / AC-3.5 closures read `gs` instead of `kindGroups[kind]`. AC-3.5 sets, on both sides, `g.lead` to `propertyPools[g.lead][0]` and, for a lifespan group, its span to `propertyPools[span][0]`. `propertyPools` gains `lifespan` (`Between 1850 and 1960`, `Between 1730 and 1840`, `Between 1790 and 1900`) and `lifespan_span` (`Between 1850 and 1850`, `Between 1840 and 1840`, `Between 1950 and 1950`), each built with `pt.Date(datevaluetest.Between(…))` | Closes MNT-2 with one hand-written table, not two: today every `same` value is `propertyPools[lead][0]` | A hand list checked against the registry — still two lists |
| D-11 | **Builder input.** `subjectprofile.LifeEvent` *(new)*: `{EventType vocabkey.EventType; Role vocabkey.Role; Dates []datevalue.Value}`: one event the person takes part in, its `event_type` term key, the person's participation role, and every date the event records. `Entity.Life []LifeEvent` *(new field)*. `subjectprofile.Subject` *(new)*: `{Values map[vocabkey.PropertyKey][]autoreconcile.Candidate; NamePattern string; Life []LifeEvent}`, the plain Subject's side, shaped like `Entity`. `Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan, opts Options)` and `ForSubject(s Subject, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan)`. `lifespans` is the registry's `Subject.Lifespans` for the side's kind (`reg.Subject.Lifespans[vocabkey.Person]` for a person, nil otherwise) | Both entries take "what the side is" plus the rule; the next input is one field on both. The map passes the lead key with its rule, as `Periods` does | Positional `events` beside `namePattern`; a whole `subjectcompare.Rules` argument — the builder needs one map |
| D-12 | **Building the lifespan.** Unexported `lifespan(events []LifeEvent, l subjectcompare.Lifespan) (window, span propertycompare.Value, ok bool)`. For each event and date: `first, last, ok := valuecompare.DateEnds(d)`, skipped when not ok; the date's early end is `first`, or `last` when `first.Open`; its late end is `last`, or `first` when `last.Open`. `spanFirst` is `Earliest` of all early ends, `spanLast` is `Latest` of all late ends; none → not ok. `start` is `Earliest` of the early ends of events whose `{EventType, Role}` is in `l.Starts`, or else `ShiftYears(spanFirst, -l.MaxYears)`; `end` is `Latest` of the late ends of events in `l.Ends`, or else `ShiftYears(spanLast, l.MaxYears)`. Not ok when a `ShiftYears` fails, when `EndsDate(start, end)` or `EndsDate(spanFirst, spanLast)` refuses (AC-2.12), or when a date's late end finishes before `start` begins (`late.Hi < start.Lo`) or its early end begins after `end` finishes (`early.Lo > end.Hi`) (AC-2.16, AC-2.17). Otherwise both values are `propertycompare.Value{Value: valuecompare.Value{Date: &d}}` of the two `EndsDate` results. `Build` and `ForSubject` walk `lifespans` in key order and set `out[lead]` and `out[l.Span]` to one value each when ok. `Options.Excluding` does not change `Entity.Life`. `Build`'s and `ForSubject`'s doc comments each gain: "lifespans is the registry's Subject.Lifespans for the side's kind; for each, a window and a span are added when the life events give them." | The user's "straightforward math function once we have all the dates in a timeline". A side whose own events fall wholly outside its own window contradicts itself; the guide is not an authority, so that side has no lifespan, as a birth after a death has none. Comparing whole-date bounds keeps a same-year event at a coarser precision (AC-2.18) | Clamping the window to the span — hides the side's own contradiction; refusing on any day outside — a year-only baptism beside a dated birth would drop the lifespan; an error — the guide is not an authority |
| D-13 | **Tests and names.** Every new test row and subtest name starts with its AC ID. Rows added to an existing function continue its numbers (US-3 from AC-3.10, US-4 from AC-4.5, AC-2.19), so no row name repeats within one function; the new functions `TestLifespan` and `TestLifespanFromEvents` may reuse numbers other functions hold, since a test is named by function and row. AC-1.10's evidence is `TestCompare`'s row "AC-1.12 both sides empty", which keeps its name. `core/subjectcompare/lifespan_test.go` *(new)*: `TestLifespan` (US-1 except AC-1.10), the `TestCompare` row shape (`name`, `kind`, `a`, `b`, `sim`, `ev`, `check`), helper `life(w1, w2, s1, s2 int) Vals`. `core/subjectprofile/lifespan_test.go` *(new)*: `TestLifespanFromEvents` (US-2 except AC-2.14 and AC-2.19), fields `name`, `events []subjectprofile.LifeEvent`, `noRules bool` (true passes a nil map, AC-2.13; otherwise `matchregistry.Default().Subject.Lifespans[vocabkey.Person]`), `values` (nil except AC-2.15), `want string`, `check`; helper `ev(eventType vocabkey.EventType, role vocabkey.Role, d *datevalue.Value) subjectprofile.LifeEvent` in `helpers_test.go`, called with `datevaluetest.YMD(y, 0, 0)`. AC-2.14 and AC-2.19 are two `TestBuild` rows: the row struct gains `lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan`, passed to `Build` (nil when unset), and both rows set it to `matchregistry.Default().Subject.Lifespans[vocabkey.Person]`. Each row's `entity` func returns `maleFemale(t)` with `Life` set. `TestEntityCompare` gains fields `life, subjectLife []subjectprofile.LifeEvent`: after `tc.entity(t)` the test sets `e.Life = tc.life`; side B is `ForSubject(subjectprofile.Subject{Values: tc.subject, Life: tc.subjectLife}, reg.Subject.Lifespans[vocabkey.Person])` for every row. AC-4.7 is a new subtest of `core/subjectprofile` `TestInvariants`: 500 cases, `rand.New(rand.NewSource(1))`; each case draws 1 + `r.Intn(4)` events, each with, in this order, a type from `birth`, `death`, `burial`, `residence`, `marriage`, a role from `subject`, `spouse`, `father`, and one date `datevaluetest.YMD(1800 + r.Intn(200), 0, 0)`; it calls `ForSubject(Subject{Life: events}, matchregistry.Default().Subject.Lifespans[vocabkey.Person])`, then again on a copy shuffled with `r.Shuffle`. A case with no lifespan counts toward the 500 and checks only the shuffle; the subtest fails when fewer than 100 cases build a lifespan | No invented names; the repo's table shape and the S11-10 / S11-11 precedents | — |
| D-14 | **Docs.** The design doc's lifespan rule, the map's working rule and scope entry already say the 2026-10-18 calls (written with this plan). The PR adds: in code, `subjectcompare.Values`' doc and `vocabkey`'s package doc (STEP-2); in docs, `comparison-layers.md`'s "Built in S11-12" paragraph, `matching-model.md` glossary, worked example and inventory, the skill's Layer 3, profile builder and "Add a Subject property or pair" text (STEP-2, STEP-6). Exact text in the steps | Docs follow the layer as it lands | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `subjectcompare.Lifespan`, `EventRole`, `Rules.Lifespans` *(new)* | type, field | `subjectcompare.Period`, `Rules.Periods` | Same shape (D-4) |
| `lifespan` in `core/subjectcompare/lifespan.go` *(new)* | func | `side` / `period` (`pairs.go`) | Those build one side's values; the lifespan pairs one side's span with the other's window (D-5) |
| `datevalue.AddYears` *(new)* | func | `datevalue.AddMonths`, `AddDays` | `AddMonths` has no year-only form; same file, same style (D-6) |
| `valuecompare.ShiftYears`, `Earliest`, `Latest` *(new)* | func | `AboutEnds`, `period`'s loops | `AboutEnds` widens an `ABT` only; `Earliest` / `Latest` replace `period`'s loops (D-6, D-7) |
| `vocabkey.EventType`, `Role` and six constants *(new)* | type, const | `vocabkey.SubjectType`, `PropertyKey`; `core/graphalign/types.go` `EdgeSignature`; literals in `core/hops/hops.go`, `conclusionheaders/compose.go`; `core/eventtitle` `RuleSubject` | The key home; `EdgeSignature` is untyped and retired by S11-14, `RuleSubject` is an event-title rule name (D-3) |
| `subjectprofile.LifeEvent`, `Subject`, `Entity.Life` *(new)* | type, field | `subjectprofile.Entity`, `Candidate`; `conclusionheaders` `lifePick` | `lifePick` is a header's database read; `Subject` mirrors `Entity` (D-11) |
| `lifespan` in `core/subjectprofile` *(new, unexported)* | func | `subjectcompare.period` | Builds a window and a span from events; shares `DateEnds`, `Earliest`, `Latest`, `ShiftYears`, `EndsDate` (D-12) |
| `groups` *(new, test)* | helper | `kindGroups` | Replaces it (D-10) |
| `TestLifespan`, `TestLifespanFromEvents`, `TestAddYears`, `TestShiftYears`, `TestEarliestLatest` *(new)*; rows in `TestCompare`, `TestBuild`, `TestEntityCompare`, `TestInvariants` | test | S11-10 `TestCompare`, S11-11 `TestBuild` | Same shapes (D-13) |

**Builds on.** `valuecompare.DateEnds`, `EndsDate` (S11-10), `Window`; `propertycompare.Compare`; `subjectcompare.Compare` and its `Periods` pattern (S11-10); `subjectprofile.Build` / `ForSubject` / `value` (S11-11); the inventory guard, `TestVocabKeysMatchSeed` and the seed check; the `add-match-rule` skill.
**Deletes.** `kindGroups` and `group.same`; `period`'s two selection loops; the inventory rows `subject.person.lifespan.MaxYears`, `ParentMinYears`, `ParentMaxYears`, `MotherMaxYears`, `MarriageMinYears`; `notYetBuilt`'s two lifespan prefixes. (The design doc's per-bridge example list went with this plan's commit.)
**Next adds.** S11-13's loader reads each person's participations into `[]LifeEvent` (event type and role term keys, the event's displayed date values or its period) and passes `reg.Subject.Lifespans[vocabkey.Person]` to `Build` and `ForSubject`. Another life-fixing event is one `Starts` / `Ends` entry. S11-14's typed `BridgeKey` starts from these typed term keys (its `Role` field also holds relationship terms, so it needs a role-or-term key built beside `vocabkey.Role`), and its one-cardinality death segment covers a second death date. S11-19 reads the lifespan `PropertyResult` (window and span as the speaking pair) for an evidence line.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/datevalue` | `calendar.go`: `AddYears` (D-6) |
| `core/valuecompare` | `datewindow.go`: `ShiftYears`, `Earliest`, `Latest` (D-6) |
| `core/vocabkey` | `EventType`, `Role`, six constants, package doc (D-3) |
| `core/subjectcompare` | `subjectcompare.go`: `Lifespan`, `EventRole`, `Rules.Lifespans`, the branch in `Compare`, `Values` doc (D-4, D-5); `lifespan.go` *(new)*; `pairs.go`: `period` on `Earliest` / `Latest` (D-7) |
| `core/matchregistry` | `subject.go` defaults (D-8); tests (D-9) |
| `core/database/subjectvocab` | tests only (D-3, D-9) |
| `core/subjectprofile` | `LifeEvent`, `Subject`, `Entity.Life`, the new arguments, `lifespan` (D-11, D-12) |

```go
// vocabkey

// EventType is an event_type term key, typed for registry rules.
type EventType string

// Role is a participation role term key, typed for registry rules.
type Role string

// subjectcompare

// EventRole names an event a person takes part in by its type and the
// person's role in it.
type EventRole struct {
	EventType vocabkey.EventType
	Role      vocabkey.Role
}

// Lifespan is a derived lifespan: the lead holds its window, where a life can
// lie, and Span its first to last dated event. Compare checks each side's span
// against the other side's window; the profile builder builds both.
type Lifespan struct {
	Span     vocabkey.PropertyKey // not a lead
	MaxYears int                  // pads an end no Starts or Ends event fixes
	Starts   []EventRole          // events that fix the start
	Ends     []EventRole          // events that fix the end
}

// subjectprofile

// LifeEvent is one event a person takes part in: its type, the person's role
// in it, and every date the event records.
type LifeEvent struct {
	EventType vocabkey.EventType
	Role      vocabkey.Role
	Dates     []datevalue.Value
}

// Subject is a plain Subject's own Observations by property and its life
// events.
type Subject struct {
	Values      map[vocabkey.PropertyKey][]autoreconcile.Candidate
	NamePattern string // the catalog's name format key, set on every name value
	Life        []LifeEvent
}

func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan, opts Options) (subjectcompare.Values, error)
func ForSubject(s Subject, lifespans map[vocabkey.PropertyKey]subjectcompare.Lifespan) subjectcompare.Values
```

`subjectcompare.Values`' doc becomes "Values are one Subject's values by property: product properties, and the derived properties the profile builder adds (a person's lifespan)." The `subjectprofile` package doc's "no numbers" stays true: `MaxYears` comes in as an argument.

### FFI

None — no caller yet.

### macOS

None — no app change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Date helpers and the period fold**
Files: `core/datevalue/calendar.go`, `core/datevalue/calendar_test.go`, `core/valuecompare/datewindow.go`, `core/valuecompare/datewindow_test.go`, `core/subjectcompare/pairs.go`. Skill: `add-date-value`.
Do: first run `go test ./core/subjectcompare/... -v | grep -c -- '--- PASS'` and write the count down. D-6 with `TestAddYears` (T-1: `(1850, 0, 0, -110)` → `(1740, 0, 0)`; `(1904, 2, 29, 10)` → `(2014, 2, 28)`; `(1904, 2, 29, 4)` → `(1908, 2, 29)`; `(1901, 5, 1, 110)` → `(2011, 5, 1)`; `(1850, 3, 0, 1)` → `(1851, 3, 0)`), `TestShiftYears` (T-1: year-precision 1850 by +110 → `Year` 1960, `Lo` = `DaysFromCivil(1960, 1, 1)`; day 1901-05-01 by +110 → `Lo` = `Hi` = `DaysFromCivil(2011, 5, 1)`; an open end → unchanged, ok) and `TestEarliestLatest` (T-1: lower `Lo` wins `Earliest`; equal `Lo`, lower `Hi` wins; higher `Hi` wins `Latest`; equal `Hi`, higher `Lo` wins; open ends skipped; all open or empty → not ok). D-7.
Check: `go test ./core/datevalue/... ./core/valuecompare/... ./core/subjectcompare/... ./core/matchregistry/...` passes with no tags; the `subjectcompare` PASS count is unchanged.

**STEP-2 — The lifespan lead in layer 3 and the registry**
Files: `core/vocabkey/vocabkey.go`, `core/database/subjectvocab/vocabkey_test.go`, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/lifespan.go` *(new)*, `core/subjectcompare/lifespan_test.go` *(new)*, `core/subjectcompare/subjectcompare_test.go`, `core/matchregistry/subject.go`, `core/matchregistry/subject_test.go`, `core/matchregistry/registry_test.go`, `core/matchregistry/inventory_test.go`, `core/database/subjectvocab/subject_section_test.go`, `docs/matching-model.md` (inventory only). Skill: `add-match-rule` ("Add a knob", "Add a Subject property or pair").
Do: D-3, D-4, D-5, D-8, D-9 and the `Values` doc. `TestLifespan` with US-1 rows except AC-1.10 (T-2). In `TestCompare`'s existing "AC-1.12 both sides empty" row, the expected keys become `{vocabkey.PropertyLifespan, name, sex}` (AC-1.10); no other existing row changes. In `docs/matching-model.md` §Knob inventory, delete the five `subject.person.lifespan.*` rows and add these four table rows after the `subject.Weights.person.lifespan.Conflict` row:

```text
| `subject.Lifespans.person.lifespan.Span` | `lifespan_span` | derived (decided 2026-10-18) | The value holding a life's first to last dated event; the lead `lifespan` holds its window |
| `subject.Lifespans.person.lifespan.MaxYears` | 110 | estimate: longest plausible human life | Years an unfixed start or end is padded from the first or last dated event |
| `subject.Lifespans.person.lifespan.Starts` | `birth` · `subject` | decided 2026-10-18 | Events that fix the start of a life |
| `subject.Lifespans.person.lifespan.Ends` | `death` · `subject`, `burial` · `subject` | decided 2026-10-18 | Events that fix the end of a life |
```

In the `subject.Weights.person.lifespan.Agree` row, the Changes cell "Points for overlapping lifespans" becomes "Points for consistent lifespans". In the §Knob inventory intro, "a pool is keyed by its lead in `subject.Pools`, a period by its lead in `subject.Periods`." becomes "a pool is keyed by its lead in `subject.Pools`, a period by its lead in `subject.Periods`, a lifespan by its lead (the window) in `subject.Lifespans`." `notYetBuilt` loses `"subject.Weights.person.lifespan."` and `"subject.person.lifespan."`.
Check: `go test ./core/vocabkey/... ./core/subjectcompare/... ./core/matchregistry/...` passes with no tags; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` passes; `rg -n 'subject\.person\.lifespan' docs/matching-model.md` prints nothing.

**STEP-3 — Lead groups from the registry**
Files: `core/subjectcompare/invariants_test.go`.
Do: D-10 (T-3). Temporarily delete the `lifespan` entry from `propertyPools`, run the test, and confirm it fails with "person.lifespan has no propertyPools entry"; restore it.
Check: `go test ./core/subjectcompare/... -run TestInvariants -v` passes and lists the six person subtests; `rg -n 'kindGroups|same V' core/subjectcompare` prints nothing.

**STEP-4 — The builder's lifespan**
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/helpers_test.go`, `core/subjectprofile/lifespan_test.go` *(new)*, `core/subjectprofile/subjectprofile_test.go`, `core/subjectprofile/entity_test.go`, `core/subjectprofile/invariants_test.go`.
Do: D-11, D-12. The existing `Build` calls pass `nil` lifespans before `opts`; the existing `ForSubject` calls become `ForSubject(subjectprofile.Subject{Values: <their values>, NamePattern: <their name pattern>}, nil)`, so no existing case changes. `TestLifespanFromEvents` (T-4) with US-2 rows except AC-2.14 and AC-2.19; those two as `TestBuild` rows (T-4).
Check: `go test ./core/subjectprofile/...` passes with no tags; `go vet ./core/subjectprofile/...` is clean; `go test ./core/matchregistry/ -run TestLayerPackages` passes.

**STEP-5 — Entity compare and invariants**
Files: `core/subjectprofile/entity_test.go`, `core/subjectprofile/invariants_test.go`.
Do: D-13's `TestEntityCompare` fields; every row's `Build` now passes `reg.Subject.Lifespans[vocabkey.Person]` and side B is built from `Subject{Values: tc.subject, Life: tc.subjectLife}` with the same map (rows with no life events are unchanged). US-3 rows AC-3.10 – AC-3.14 (T-5); each row's `check` is `lifespanResult(outcome)` *(new closure beside `sexResult` in `entity_test.go`)*, which asserts the `lifespan` `PropertyResult`'s outcome only. AC-4.7's subtest (T-6).
Check: `go test ./core/subjectprofile/... -v -run 'TestEntityCompare|TestInvariants'` passes; then `go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/datevalue/... ./core/vocabkey/... ./core/matchregistry/...` passes with no tags.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan (status).
Do:
- `comparison-layers.md` §Lifespan, after its last bullet (the one that starts "**Bounded by construction:**"), a paragraph *(new)*: "**Built in S11-12.** `subjectprofile.Build` and `ForSubject` take the person's life events (event type, role, dates) and the registry's `subject.Lifespans` for the kind, and add two derived values, `lifespan` (the window) and `lifespan_span`. `subjectcompare.Compare` scores the `lifespan` lead from both directions: a conflict loses `subject.Weights.person.lifespan.Conflict`, consistent lifespans earn its `Agree`."
- `matching-model.md` §Glossary, the **lifespan** row becomes "| **lifespan** | A person's derived window and span, built from the dates of every event they take part in; each side's span is checked against the other's window |". §Worked example step 3: "Both sides also derive a lifespan, "born 1 May 1901", from their birth; the windows are consistent" becomes "Both sides also derive a lifespan from their birth: window 1 May 1901 to 1 May 2011, span 1 May 1901; each span lies in the other's window".
- `add-match-rule` SKILL.md: the Layer 3 bullet's sentence "Value type and cardinality come from the catalog as `Metas`, never from the registry." becomes "Value type and cardinality come from the catalog as `Metas`, never from the registry; a derived property's meta is fixed beside the rule that reads it (the lifespan's date meta)."; at the end of the Layer 3 bullet, append "`subject.Lifespans` declares a lifespan lead (the window) and its `Span`; `Compare` checks each side's span against the other's window, and a lifespan's agree means consistent, whatever partial relation spoke." In the profile builder bullet, "The builder has no registry section." becomes "The builder reads `subject.Lifespans` (passed in) to build a person's window and span from its life events." In "Add a Subject property or pair" step 1, append "A derived key (built by the profile builder, never seeded) is not added there."
- `completed.md`, newest first: "## S11-12 — Lifespan for people" / "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). People now have a lifespan. The profile builder lays every dated event a person takes part in on one timeline; birth starts it, death or burial ends it, and an open end is padded by 110 years. Subject compare checks each side's dated events against the other side's window, so a person who died before a Subject was born, a 140-year life, or a residence after a death is a large conflict. The four age-rule knobs were dropped. No caller yet; nothing changes for researchers." with the date and PR number filled in.
- Map row status `Landed #N`; this plan's Status `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)` and a **Landed** note under the header listing each difference from the plan (or "None."). The map's S11-12 scope entry stays as it is, its "Also builds" paragraph included.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-12-lifespan.md` reports no errors; `rg -n 'ParentMinYears|MotherMaxYears|MarriageMinYears' docs/matching-model.md docs/deployment-plan/spike-11/comparison-layers.md core` prints nothing; `go test ./core/matchregistry/...` passes.

Commits: one per step, in step order. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Check a person's lifespan when two people are compared."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/datevalue/calendar_test.go` `TestAddYears` *(new)*; `core/valuecompare/datewindow_test.go` `TestShiftYears`, `TestEarliestLatest` *(new)* | STEP-1's cases | D-6, D-7 |
| T-2 | Go | `core/subjectcompare/lifespan_test.go` `TestLifespan` *(new)*; `subjectcompare_test.go` `TestCompare` (both-sides-empty row) | Rows AC-1.1 – AC-1.11 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, D-2, D-5 |
| T-3 | Go | `core/subjectcompare/invariants_test.go` `TestInvariants` | Groups from the registry; person subtests with the lifespan group | AC-4.5, AC-4.6, D-10 |
| T-4 | Go | `core/subjectprofile/lifespan_test.go` `TestLifespanFromEvents` *(new)*; `subjectprofile_test.go` `TestBuild` rows AC-2.14, AC-2.19 | Rows AC-2.1 – AC-2.19 | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, AC-2.11, AC-2.12, AC-2.13, AC-2.14, AC-2.15, AC-2.16, AC-2.17, AC-2.18, AC-2.19, D-1, D-11, D-12 |
| T-5 | Go | `core/subjectprofile/entity_test.go` `TestEntityCompare` | Rows AC-3.10 – AC-3.14 | AC-3.10, AC-3.11, AC-3.12, AC-3.13, AC-3.14 |
| T-6 | Go | `core/subjectprofile/invariants_test.go` `TestInvariants`, subtest AC-4.7 *(new)* | 500 generated event sets → the window overlaps every event date; shuffle-stable | AC-4.7, D-13 |
| T-7 | Go | `core/matchregistry` `TestKnobInventoryMatchesRegistry`, `TestSubjectSection`, `TestDefaultIsFresh`, `TestLayerPackages`; `core/database/subjectvocab` `TestVocabKeysMatchSeed`, `TestSubjectSectionMatchesSeed` | New fields have rows; structure, freshness, purity, the term keys and the seed hold | D-3, D-8, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestCompare`'s person rows that list keys, `TestSubjectSectionMatchesSeed` (the derived lead), `TestKnobInventoryMatchesRegistry` (a row typo), `core/subjectprofile`'s existing tests after the signature change.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/datevalue/... ./core/vocabkey/... ./core/matchregistry/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/autoreconcile, core/match and core/database's non-test files are unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md (Built in), matching-model.md, add-match-rule skill, subjectcompare.Values and vocabkey package docs (STEP-2, STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/subjectprofile/... -v -run TestEntityCompare` with no build tags and reads subtests `AC-3.10` to `AC-3.13`: these are the map's Check (died 1840 / born 1850, born 1850 / died 1990 and died 1900 / resided 1950 conflict; born 1850 / died 1920 is consistent). Then `go test ./core/matchregistry/... -v -run TestKnobInventoryMatchesRegistry` to see the lifespan rows match the registry.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A date compare change after the planning base moves a relation §2 reads | §2's relations were read at `b799b67`; a drift is a stop condition, not a test edit |
| The lifespan lead changes existing person results | It is unknown (no points) whenever a side has no lifespan values, so only the key-list row changes (AC-1.10); any other changed row is a stop condition |
| A generated case with two windows on one side breaks symmetry | D-5 returns a conflict from either direction; `same` compares outcomes and numbers only. A failure is a stop condition |

**Stop and report if:**

- `main` differs from §4: `subjectcompare.Rules`, `Compare`'s signature or loop, `period`, `subjectprofile.Build` / `ForSubject` signatures, `valuecompare.DateEnds` / `EndsDate`, `datevalue.DaysInMonth`, `notYetBuilt`, `TestVocabKeysMatchSeed`, or the inventory's lifespan rows. Report each difference.
- `propertycompare.Compare` gives a relation other than §2's conventions paragraph for one of its pairs: report the pair and the result; do not change the AC.
- An existing `TestCompare`, `TestBuild`, `TestForSubject` or `TestEntityCompare` row other than AC-1.10's changes its sim, ev or labels.
- `TestKnobInventoryMatchesRegistry` reports a key other than D-8's four `subject.Lifespans.person.lifespan.*` keys and the two weight rows.
- A layer package fails `TestLayerPackages` (a numeric literal other than 0 and 1, or a `core/database` import).
- An invariant of US-4 fails: report the seed, case index and the values.
- `vocabkey` already declares any of D-3's names, or `subjectcompare.Rules` already has a `Lifespans` field.

## 12. Open questions

None.

## 13. PR

**Title:** S11-12 — Lifespan for people
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-12-lifespan.md`.
