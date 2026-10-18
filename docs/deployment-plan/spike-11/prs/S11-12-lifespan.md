# S11-12 — Lifespan for people

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 3 — Property, Subject and entity compare
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: pure Go layers with no caller yet; nothing reaches the app)
**Board:** —
**Depends on:** S11-11 (landed [#476](https://github.com/mendahu/provenencia/pull/476); [plan](S11-11-subject-profile.md) and its **Landed** note: STEP-2 and STEP-3 were one commit; expected strings use `Labels`' `key={a, b}` form). Branch from `main`
**Blocks:** S11-13
**Branch:** `s11-12-lifespan`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) ("Add a knob", "Add a Subject property or pair")
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [Lifespan](../comparison-layers.md#lifespan-a-derived-property-for-people) (superseded in part by D-1, the user's call of 2026-10-18); [Layer 3](../comparison-layers.md#layer-3-subject-comparison) ("Built in S11-10", "A derived property (lifespan) is a lead in `subject.Weights` like any other"); [Entities](../comparison-layers.md#entities-compare-against-the-conclusion) ("Built in S11-11"); [`docs/matching-model.md`](../../../matching-model.md) worked example step 3, glossary **lifespan**, knob inventory `subject.Weights.person.lifespan.*` and `subject.person.lifespan.*` rows; S11-10 [ship log](S11-10-subject-compare.ship.md) follow-ups MNT-2 and MNT-3; working rules "Lifespan is for people only", "Sparse data first", "Registry numbers: carry over, then estimate", "One match registry", "Pure means the package builds and tests with no build tags and no `core/database` import", "One matching skill, grown with the layers"
**Planned:** 2026-10-18 against `main` @ `b799b67` (S11-11 merged)
**Review:** [`S11-12-lifespan.review.md`](S11-12-lifespan.review.md)

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

A person gets a lifespan, and Subject compare uses it to rule out impossible matches. The profile builder (`core/subjectprofile`) lays every dated event the person takes part in on one timeline. A birth starts the lifespan and a death or burial ends it; an end with no such event is padded by `MaxYears` (110) from the first or last dated event. Each side carries two derived values: its **window** (where its life can lie) and its **span** (its first to last dated event). Subject compare (`core/subjectcompare`) checks each side's span against the other side's window: a span outside the window, beyond the date tolerance, is a conflict (−20); otherwise the lifespans agree (+0.1).

**Why now.** S11-13 moves Promote's candidate scoring onto entity compare. With the lifespan in place first, "a person who died before the Subject was born is never proposed" (goal 2) holds the day Promote moves.

**What stays.** `core/match`, `core/database/matching`, `core/autoreconcile` and every caller are untouched; no caller passes life events yet (S11-13's loader does). No catalog, wire, Swift or researcher-visible change. Events and places get no lifespan.

## 2. User stories

Conventions. `L(w1–w2 | s1–s2)` is a person's two derived values at year precision: `lifespan` = `datevaluetest.Between(w1, 0, 0, w2, 0, 0)` and `lifespan_span` = `datevaluetest.Between(s1, 0, 0, s2, 0, 0)`, each a positive `propertycompare.Value` with only `Date` set. A life event is written `type·role year`: `subjectprofile.LifeEvent{EventType: type, Role: role, Dates: []datevalue.Value{*datevaluetest.YMD(year, 0, 0)}}`. "sim" and "ev" are `subjectcompare.Result.Similarity` and `.Evidence` (within 0.0001) under `matchregistry.Default()`, with metas `subjectcomparetest.Metas()`. Expected labels are `subjectcomparetest.Labels` strings. Every relation below was read from `propertycompare.Compare` at `b799b67` with the date meta and the default value rules: window 1730–1840 against span 1850 is `different`; window 1850–1960 against span 1990 is `different`; window 1850–1960 against span 1920, and span 1850 against window 1810–1920, are `contains` (partial); span 1950 against window 1790–1900 is `different`; span 1900 against window 1840–2060 is `contains`; window 1850–1960 against span 1962 is `near` (partial, 0.0667).

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

**US-2.** As the engine (S11-13's loader), when I build a person's side from its life events, I get its window and its span.

Each row calls `subjectprofile.ForSubject(nil, events, "", matchregistry.Default().Subject.Lifespan)` and compares `subjectcomparetest.Labels` of the result.

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
| AC-2.9 | `death·subject` with date `BEF 1840` (`datevalue.Value{Kind: datevalue.KindPoint, Qualifier: datevalue.QualifierBEF, StartYear: datevaluetest.Part(1840)}`) | Same as AC-2.2 |
| AC-2.10 | `birth·subject` on `datevaluetest.YMD(1904, 2, 29)` | One `lifespan` value; `valuecompare.DateEnds` of it gives first `(1904, 2, 29)` and last `(2014, 2, 28)` |
| AC-2.11 | One `residence·subject` with `Dates` nil, and one whose date is `*datevaluetest.Phrase("spring")` | Non-nil, length 0 |
| AC-2.12 | `birth·subject 1900`, `death·subject 1850` | Non-nil, length 0 (the side contradicts itself) |
| AC-2.13 | AC-2.1's events with the zero `subjectcompare.Lifespan` | Non-nil, length 0 |
| AC-2.14 | `Build` with `Entity{Properties: maleFemale's, Life: [birth·subject 1850]}`, then again with `Options{Excluding: [][]byte{mem(1)}}` | First: `lifespan={Between 1850 and 1960}; lifespan_span={Between 1850 and 1850}; sex_at_birth={male}`. Second: the same lifespan values with `sex_at_birth={male, female}` |
| AC-2.15 | `ForSubject` with `sex_at_birth` male and AC-2.1's events | `lifespan=…; lifespan_span=…; sex_at_birth={male}` (AC-2.1's two values) |

**US-3.** As the engine, the map's Check holds for entity compare: `Build` then `subjectcompare.Compare` rules out a person who cannot be the entity. In each row side A is `Build` of an entity whose `name` is James Robins (M1, S1) and whose `Life` is given, and side B is `ForSubject` of `name` James Robins with the given events; both with `matchregistry.Default().Subject.Lifespan`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Entity `death·subject 1840`; Subject `birth·subject 1850` | `lifespan` conflict; sim 0.1111, ev −17.5 |
| AC-3.2 | Entity `birth·subject 1850`; Subject `death·subject 1990` | `lifespan` conflict; sim 0.1111, ev −17.5 |
| AC-3.3 | Entity `birth·subject 1850`; Subject `death·subject 1920` | `lifespan` agree; sim 1, ev 2.6 |
| AC-3.4 | Entity `death·subject 1900`; Subject `residence·subject 1950` | `lifespan` conflict; sim 0.1111, ev −17.5 |
| AC-3.5 | Both `birth·subject` on `datevaluetest.YMD(1901, 5, 1)` (the worked example's Gracie) | `lifespan` agree; sim 1, ev 2.6 |

**US-4.** As a contributor, the invariants cover every lead in the registry, the lifespan included, so a lead added later cannot be skipped silently (S11-10 MNT-2).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `core/subjectcompare` `TestInvariants` | Its lead groups come from `matchregistry.Default().Subject`: the leads of `Weights`, each lead's `Pools` `With` and `Periods` `Start` / `End`, and `Lifespan.Span` for `Lifespan.Window`. A lead with no hand-written test value fails the test by name |
| AC-4.2 | Its six existing invariant subtests for person | Run with the lifespan group and pass |
| AC-4.3 | `core/subjectprofile` `TestInvariants` *(new subtest)*: generated life events | When a lifespan is built, its window contains its span (`valuecompare.Window.Contains` on `valuecompare.DateWindow` of each); labels are equal after shuffling the events |

## 3. Scope

| In | Covered by |
| --- | --- |
| `valuecompare.ShiftYears`, `DateEnd.StartsBefore`, `DateEnd.EndsAfter` (D-6) | STEP-1 |
| `subjectcompare`'s `period` start and end loops folded into one helper (D-7, MNT-3) | STEP-1 |
| `vocabkey` derived property keys and the anchor term keys (D-3) | STEP-2 |
| `subjectcompare.Lifespan`, `EventRole`, `Rules.Lifespan`; the lifespan special case in `Compare` (D-2, D-4, D-5) | STEP-2, US-1 |
| Registry: `subject.Lifespan` and `subject.Weights.person.lifespan`; inventory rows; `notYetBuilt`; registry and seed tests (D-8, D-9) | STEP-2 |
| `kindGroups` built from the registry (D-10, MNT-2) | STEP-3, US-4 |
| `subjectprofile.LifeEvent`, `Entity.Life`, the lifespan argument of `Build` and `ForSubject`, the builder (D-1, D-11, D-12) | STEP-4, US-2 |
| Entity compare and builder invariant tests (D-13) | STEP-5, US-3, US-4 |
| Docs the lifespan makes true: `comparison-layers.md`, `matching-model.md`, the map's working rule and scope entry, the `add-match-rule` skill (D-14) | STEP-6 |

| Out | Home |
| --- | --- |
| Loading a person's life events from the catalog (participations, event types, roles, the event's displayed dates or period) and passing them to `Build` / `ForSubject` | S11-13 (the loader that feeds `propertycompare` and `subjectprofile`) |
| Lifespans for events and places | Not planned — working rule "Lifespan is for people only" |
| Age rules (marriage at 14 or older, parents 12–60 at a child's birth, mothers narrower) | Not planned — dropped by the user's call of 2026-10-18 (D-1); the four knobs leave the inventory |
| Two-hop rules (a child's birth read through a parent Relationship, a grandchild's birth) | Not planned in Spike 11 — spike README "Not this spike" names two-hop lifespan rules |
| `Options.Excluding` changing the lifespan (a left-out member's participations) | [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) — the weak-claim review is the first caller of `Excluding`; D-12 |
| Moving `conclusionheaders`' `"birth"` / `"death"` literals (`core/database/conclusionheaders/compose.go`) onto the new `vocabkey` constants | Not planned — not matching code; moves when that file is next edited |
| A conflict for one side's two different death dates | S11-14 — a death is a one-cardinality segment ([segments](../comparison-layers.md#segments-layer-5-and-graph-alignment-layer-6)) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Layer 3 | `core/subjectcompare/subjectcompare.go`: `Weights`, `Pool`, `Period`, `Rules{Weights, Pools, Periods}`, `Values`, `Metas`, `Result`, `PropertyResult`, `Compare` | `Compare` runs `propertycompare.Compare(side(kind, key, a, rules), side(kind, key, b, rules), metas[key], value)` for each lead of `rules.Weights[kind]` in key order. No lifespan |
| Period building | `core/subjectcompare/pairs.go` `side`, `period` | `period` has two near-identical loops: the earliest start (lowest `Lo`, then lowest `Hi`) and the latest end (highest `Hi`, then highest `Lo`) (S11-10 MNT-3) |
| Layer 3 tests | `core/subjectcompare/subjectcompare_test.go` `TestCompare` row "both sides empty" check `slices.Equal(keys(r), []vocabkey.PropertyKey{name, sex})` (line 132); `core/subjectcompare/invariants_test.go` `group{lead, props, same}`, `kindGroups` (hand-written leads per kind), `propertyPools`, `properties`, `draw`, `empty`, `TestInvariants` AC-3.1 – AC-3.6 | `kindGroups` repeats the registry by hand (S11-10 MNT-2); AC-3.5 sets `g.lead` to `g.same` on both sides |
| Layer 3 fixtures | `core/subjectcompare/subjectcomparetest/subjectcomparetest.go` `Metas()`, `Labels` | `Metas()` holds seeded properties only; `core/database/subjectvocab/subject_section_test.go` fails on a key that is not a seeded Property |
| Registry | `core/matchregistry/subject.go` `defaultSubject()`; `core/matchregistry/registry.go` `Registry{Value, Subject}` | No lifespan entry |
| Registry guards | `core/matchregistry/inventory_test.go` `notYetBuilt` (`"subject.Weights.person.lifespan."`, `"subject.person.lifespan."`, `"segment."`, `"alignment."`, `"evidence."`), `registryKeys` (a slice field is one leaf); `core/matchregistry/subject_test.go` `TestSubjectSection`; `core/matchregistry/registry_test.go` `TestDefaultIsFresh`; `core/matchregistry/guard_test.go` `layerPackages` | The guard already lists `core/subjectcompare`, `core/subjectprofile` and `core/valuecompare` with numbers true |
| Seed check | `core/database/subjectvocab/subject_section_test.go` `TestSubjectSectionMatchesSeed` (every `Weights` key bound by the seed); `core/database/subjectvocab/registry.go` `seedTerms` (`event_type` `birth`, `death`, `burial`, `baptism`, `marriage`, `residence`; `role` `subject`, `father`, `mother`, `spouse`, `child`) | A `Weights` entry for `lifespan` fails it today |
| Date ends | `core/valuecompare/datewindow.go` `DateEnd{Year, Month, Day, Lo, Hi, From, To, Mid, Open}`, `DateEnds`, `EndsDate`, `boundEnd` (unexported), `Window`, `DateWindow`; `core/datevalue/calendar.go` `DaysInMonth`, `BoundDays`, `AddMonths`, `AddDays` | No way to move a `DateEnd` by years; `EndsDate(first, last)` reads only `Year`, `Month`, `Day`, `Lo`, `Hi` |
| Date compare | `core/valuecompare/date.go` `date.compare` | Windows apart by more than `Dates.Tolerance` (2) civil years are `Different`; within it `Near` |
| Profile builder | `core/subjectprofile/subjectprofile.go` `Entity{Properties, NamePattern}`, `Property`, `Candidate`, `Options`, `Build(e, metas, reconcile, opts)`, `without`, `ForSubject(values, namePattern)`, `value` | No life events. Doc: "Pure: no I/O, no catalog, no numbers" |
| Profile tests | `core/subjectprofile/helpers_test.go` (`obs`, `src`, `mem`, `term`, `text`, `year`, `name`, `cand`, `weak`, `provisional`, `negative`, `reconcile`, `property`, `vcand`, `maleFemale`, `denied`); `subjectprofile_test.go` `TestBuild`, `TestForSubject`; `entity_test.go` `TestEntityCompare`; `invariants_test.go` `TestInvariants` | Eight `Build` / `ForSubject` calls (`rg -n 'subjectprofile\.(Build\|ForSubject)\(' core/subjectprofile`) |
| Keys | `core/vocabkey/vocabkey.go` `SubjectType`, `PropertyKey`, `Person`, `PropertyName` … `PropertyLikeness` | No derived key, no term keys |
| Docs | `docs/deployment-plan/spike-11/comparison-layers.md` §Lifespan (five bullets; examples list marriage, parent, residence rules); `docs/matching-model.md` §Worked example step 3, §Glossary **lifespan**, §Knob inventory rows `subject.Weights.person.lifespan.Agree` / `.Conflict` and five `subject.person.lifespan.*` rows; `docs/deployment-plan/spike-11/deployment-plan.md` working rule "**Lifespan is for people only**", §S11-12 scope entry; `.agents/skills/add-match-rule/SKILL.md` Layer 3 bullet, profile builder bullet ("The builder has no registry section.") | Written before the user's call of 2026-10-18 |

**Precedent.** S11-10 (`core/subjectcompare`, [#471](https://github.com/mendahu/provenencia/pull/471)) added `Pools` and `Periods`, the two registry-declared special cases in `Compare`: copy how `Periods` is declared in `Rules`, keyed in the registry, checked in `TestSubjectSection` and `TestSubjectSectionMatchesSeed`, and given inventory rows. S11-11 ([#476](https://github.com/mendahu/provenencia/pull/476)): copy its table tests (`TestBuild`, `TestForSubject`, `TestEntityCompare`) and fixed-seed `TestInvariants` for the new rows. `valuecompare.EndsDate` (S11-10 D-8) is the one place date ends become a date again; the builder uses it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The lifespan is a timeline (user, 2026-10-18).** Every dated event a person takes part in counts, any event type, any role. The **span** is the first to the last dated event. The **window** starts at the earliest dated `Starts` event (birth, the person's own) or, with none, `MaxYears` before the first dated event; it ends at the latest dated `Ends` event (death or burial, the person's own) or, with none, `MaxYears` after the last dated event | The user asked for "a really straightforward" guide, not an authority: one math function over a timeline. It replaces the design's per-bridge rules (marriage age, parent ages) and their four knobs | Per-bridge age rules — four knobs for guidance; a birth window only — misses resided-after-death |
| D-2 | **Span in window (user, 2026-10-18).** Two lifespans conflict when either side's span is `Different` from the other side's window; otherwise they agree. `Near` (within `Dates.Tolerance`, 2 years) and `Overlaps` are not conflicts | Catches died-before-born, the 140-year span and resided-after-death (AC-1.1, AC-1.2, AC-1.4); comparing two windows alone catches only the first. Tolerance absorbs a mis-dated record | Window against window — misses AC-1.2 and AC-1.4; the merged span over `MaxYears` — a second rule for what span-in-window already finds |
| D-3 | **Keys.** `vocabkey` gains, in a new const block "Derived Property keys matching builds", `PropertyLifespan = "lifespan"` and `PropertyLifespanSpan = "lifespan_span"`, and, in a block "Seeded term keys matching reads by name", `EventBirth = "birth"`, `EventDeath = "death"`, `EventBurial = "burial"` (`event_type` terms) and `RoleSubject = "subject"` (`role` term) | Registry maps and keys are `vocabkey` constants, never literals (skill, Layer 3 bullet); a misspelled key fails to compile | `subjectcompare`-local constants — a second key home; term keys as literals in the registry |
| D-4 | **Shape in layer 3.** `subjectcompare.Lifespan` *(new)*: `{Kind vocabkey.SubjectType; Window, Span vocabkey.PropertyKey; MaxYears int; Starts, Ends []EventRole}`; `subjectcompare.EventRole` *(new)*: `{EventType, Role string}`; `Rules.Lifespan Lifespan` *(new field, after `Periods`)*. One struct, not a map: lifespans are for one kind. `Window` is a lead in `Weights[Kind]`; `Span` is not a lead. `MaxYears`, `Starts` and `Ends` are read only by the builder; layer 3 owns the type because the design puts "lifespan rules and constants" in the subject section | S11-10 D-13 made lifespan a derived lead with `subject.Weights.person.lifespan` weights; the builder already imports `subjectcompare`. Inventory keys become `subject.Lifespan.<Field>` | `map[SubjectType]map[PropertyKey]Lifespan` — a map for one entry; a `profile` registry section — the design lists five sections, and the subject section holds lifespan rules |
| D-5 | **The special case in `Compare`.** When `kind == rules.Lifespan.Kind` and `key == rules.Lifespan.Window`, the lead's `propertycompare.Result` comes from `lifespan(a, b, rules.Lifespan, value)` *(new, unexported, in `lifespan.go` (new) of `core/subjectcompare`)* instead of the `propertycompare.Compare` call on the two `side` results; scoring is unchanged. `lifespan` runs `ab := propertycompare.Compare(a[l.Window], b[l.Span], dateMeta, value)` and `ba := propertycompare.Compare(a[l.Span], b[l.Window], dateMeta, value)`, with `dateMeta = propertymeta.Meta{ValueType: propertymeta.ValueTypeDate, Cardinality: propertymeta.CardinalitySingle}`, and returns: `ab` when `ab.Outcome` is conflict; else `ba` when it is conflict; else `ab` when `ab.Outcome` is unknown; else `ba` when it is unknown; else `ab` with `Outcome` set to `propertycompare.OutcomeAgree` | Layer 2 already compares date sets with negatives and the tolerance; only the pairing (span against window) and the outcome (consistent is agree) are lifespan's own. A derived key has no catalog meta, so the meta is fixed here, beside the rule that reads it. Agree keeps the worked example's 0.1 for consistent lifespans | `metas[Window]` from the caller — S11-13 would have to add a fake catalog row; a partial with the window's 0.2 span similarity — a near-zero weight scaled again for nothing, and the worked example changes; a new `propertycompare` entry — a layer 2 rule that only layer 3 needs |
| D-6 | **Date-end helpers in layer 1** (`core/valuecompare/datewindow.go`): `func ShiftYears(e DateEnd, years int) (DateEnd, bool)` *(new)*, doc "ShiftYears moves a closed end by whole civil years, keeping its month and day; a 29 February that lands in a common year becomes 28 February. An open end is returned as is. ok is false when the moved end cannot be placed."; it computes the year, clamps the day with `datevalue.DaysInMonth`, and calls `boundEnd`. `func (e DateEnd) StartsBefore(o DateEnd) bool` *(new)*: `e.Lo < o.Lo \|\| (e.Lo == o.Lo && e.Hi < o.Hi)`. `func (e DateEnd) EndsAfter(o DateEnd) bool` *(new)*: `e.Hi > o.Hi \|\| (e.Hi == o.Hi && e.Lo > o.Lo)`. Tests `TestShiftYears`, `TestDateEndOrder` *(new, `datewindow_test.go`)* | `boundEnd` is unexported, and `EndsDate` needs an end with its `Lo` / `Hi`; the period and the lifespan pick ends by the same two orders | A shifted `DateEnd` built by hand in `subjectprofile` — wrong `Lo` / `Hi`, a second calendar rule; `AddMonths(y, m, 12×n)` per caller — breaks on year precision (month 0) |
| D-7 | **MNT-3 fold.** `period`'s two loops become one unexported helper in `pairs.go`: `pickEnd(values []propertycompare.Value, end func(first, last valuecompare.DateEnd) valuecompare.DateEnd, better func(x, y valuecompare.DateEnd) bool) (valuecompare.DateEnd, bool)`, skipping values that are not positive, have no `Date`, fail `DateEnds`, or whose chosen end is open. `period` calls it with the first end and `DateEnd.StartsBefore`, then with the last end and `DateEnd.EndsAfter` | The map's scope entry; same tie rules | — |
| D-8 | **Registry.** `defaultSubject()` adds `Weights[vocabkey.Person][vocabkey.PropertyLifespan] = {Agree: 0.1, Conflict: 20}` (comment "estimate: near zero, dates count through their segments; impossible together never recovers") and `Lifespan: subjectcompare.Lifespan{Kind: vocabkey.Person, Window: vocabkey.PropertyLifespan, Span: vocabkey.PropertyLifespanSpan, MaxYears: 110, Starts: [{EventBirth, RoleSubject}], Ends: [{EventDeath, RoleSubject}, {EventBurial, RoleSubject}]}` (comment on `MaxYears`: "estimate: longest plausible human life"). Inventory: the two `subject.Weights.person.lifespan.*` rows stay; the five `subject.person.lifespan.*` rows go; six rows *(new)*: `subject.Lifespan.Kind`, `.Window`, `.Span`, `.MaxYears`, `.Starts`, `.Ends` (STEP-2 text). `notYetBuilt` loses both lifespan prefixes | Numbers carried from the inventory's estimates (working rule); `registryKeys` walks a struct field by field and treats a slice as one leaf | — |
| D-9 | **Registry tests.** `TestSubjectSection` gains: `Lifespan.Window` has a `Weights[Lifespan.Kind]` entry; `Lifespan.Span` has none and is in no pool or period; `Lifespan.MaxYears > 0`; `Starts` and `Ends` are not empty. `TestDefaultIsFresh` also changes `first.Subject.Lifespan.Starts[0]` and `first.Subject.Lifespan.Ends[0]`. `TestSubjectSectionMatchesSeed` skips `Lifespan.Window` in its `Weights` loop (derived, not bound) and checks `Lifespan.Kind` is a seeded type, each `Starts` / `Ends` `EventType` is a seeded `event_type` term and each `Role` a seeded `role` term (from `seedTerms`) | The seed check is where term keys meet the catalog; the skip is the "derived key is not a seeded binding" S11-10 named | A `vocabkey` list of derived keys — one more list for two keys the struct already names |
| D-10 | **MNT-2: lead groups from the registry.** In `invariants_test.go`, `kindGroups` becomes a function `groups(kind) []group` built from `matchregistry.Default().Subject`: one group per `Weights[kind]` lead in key order, its `props` the lead, then its pool's `With`, its period's `Start` and `End`, and `Lifespan.Span` when the lead is `Lifespan.Window` for that kind. Test values stay by hand: `sameValues map[vocabkey.PropertyKey]V` (name `James Robins`, sex `male`, event type `birth`, date 1817, toponym `York`, lifespan `Between 1850 and 1960`, lifespan_span `Between 1850 and 1850`); a lead missing from it fails with `t.Fatalf("lead %s.%s has no test value in sameValues", kind, lead)`. AC-3.5 sets every prop of the group that has a `sameValues` entry (today: only the lead, so existing groups behave as before). `propertyPools` gains `lifespan` (`Between 1850 and 1960`, `Between 1730 and 1840`, `Between 1790 and 1900`) and `lifespan_span` (`Between 1850 and 1850`, `Between 1840 and 1840`, `Between 1950 and 1950`) | Closes MNT-2: a lead added to the registry is tested or fails by name | A check that the hand list matches the registry — still two lists |
| D-11 | **Builder input.** `subjectprofile.LifeEvent` *(new)*: `{EventType, Role string; Dates []datevalue.Value}`: one event the person takes part in, its `event_type` term key, the person's participation `role` term key, and every date the event records. `Entity.Life []LifeEvent` *(new field)*. `Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, life subjectcompare.Lifespan, opts Options)` and `ForSubject(values map[vocabkey.PropertyKey][]autoreconcile.Candidate, events []LifeEvent, namePattern string, life subjectcompare.Lifespan)`: the new argument is the registry's `Subject.Lifespan`. Callers pass life events for people only; layer 3 weighs a lifespan only for `Lifespan.Kind` | S11-11's Next adds named this shape; the loader (S11-13) reads term keys and dates, so the builder takes them plainly | A `subjectcompare.Rules` argument — the builder needs one field; a typed term key — `vocabkey` has no term key type, and one is not worth adding for two fields |
| D-12 | **Building the lifespan.** Unexported `lifespan(events []LifeEvent, l subjectcompare.Lifespan) (window, span propertycompare.Value, ok bool)` in `subjectprofile.go`. `ok` is false when `l.Window` is empty (the zero rules). For each event and each date: `first, last, ok := valuecompare.DateEnds(d)`, skipped when not ok; its early end is `first`, or `last` when `first.Open`; its late end is `last`, or `first` when `last.Open`. The span's first end is the earliest early end (`StartsBefore`), its last end the latest late end (`EndsAfter`). The window's start is the earliest early end of the events whose `{EventType, Role}` is in `l.Starts`, or else `ShiftYears(spanFirst, -l.MaxYears)`; its end is the latest late end of those in `l.Ends`, or else `ShiftYears(spanLast, l.MaxYears)`. Both values are `valuecompare.EndsDate(start, end)` (and of the span's ends), each as `propertycompare.Value{Value: valuecompare.Value{Date: &d}}`. No dated event, a failed `ShiftYears`, or a window or span that `EndsDate` refuses (start after end, AC-2.12) → `ok` false. `Build` and `ForSubject` set `out[l.Window]` and `out[l.Span]` to one value each when ok. `Options.Excluding` does not change `Entity.Life` | The user's "straightforward math function once we have all the dates in a timeline"; dates read as written (`DateEnds`, not `AboutEnds`): compare's tolerance absorbs an `ABT` | An error for a self-contradicting side — the guide is not an authority, so the side has no lifespan |
| D-13 | **Tests.** `core/subjectcompare/lifespan_test.go` *(new)*: `TestLifespan` (US-1), rows named by AC, the `TestCompare` row shape (`kind`, `a`, `b`, `sim`, `ev`, `check`), with a helper `life(w1, w2, s1, s2 int) Vals`. `core/subjectprofile/lifespan_test.go` *(new)*: `TestLifespanFromEvents` (US-2 except AC-2.14), rows `name`, `events []subjectprofile.LifeEvent`, `rules` (zero → `matchregistry.Default().Subject.Lifespan`; AC-2.13 sets a flag `zeroRules`), `values` (nil except AC-2.15), `want string`, `check`; helpers `ev(eventType, role string, d *datevalue.Value) subjectprofile.LifeEvent` and `yr(y int) *datevalue.Value` in `helpers_test.go`. AC-2.14 is a `TestBuild` row. US-3 rows join `TestEntityCompare` with a new field `life`, `subjectLife []subjectprofile.LifeEvent`. AC-4.3 is a new `TestInvariants` subtest in `core/subjectprofile/invariants_test.go`: 500 cases, `rand.New(rand.NewSource(1))`, 1 – 4 events each from types `birth`, `death`, `burial`, `residence`, `marriage` and roles `subject`, `spouse`, `father`, with a year in 1800 – 1999 drawn `1800 + r.Intn(200)` | The repo's table shape (`go-tests.mdc`) and the S11-10 / S11-11 precedents | — |
| D-14 | **Docs.** `comparison-layers.md` §Lifespan gets a "Decided (2026-10-18)" paragraph and a "Built in S11-12" paragraph and loses the per-bridge example list; `matching-model.md` glossary **lifespan**, worked example step 3 and the inventory; the map's working rule and S11-12 scope entry and Check; the skill's Layer 3 and profile builder bullets. Exact text in STEP-6 | Docs follow the layer as it lands | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `subjectcompare.Lifespan`, `EventRole`, `Rules.Lifespan` *(new)* | type, field | `subjectcompare.Period`, `Rules.Periods` | Extends `Rules` the same way; one struct because lifespans are for one kind (D-4) |
| `lifespan` in `core/subjectcompare/lifespan.go` *(new)* | func | `side` / `period` (`pairs.go`) | Those build one side's values; the lifespan pairs one side's span with the other's window, which `side` cannot (D-5) |
| `valuecompare.ShiftYears`, `DateEnd.StartsBefore`, `DateEnd.EndsAfter` *(new)* | func, method | `boundEnd`, `AboutEnds` (moves by units of the most precise part), `datevalue.AddMonths` | `AboutEnds` widens an `ABT` only; `AddMonths` has no year-only form (D-6) |
| `pickEnd` *(new, unexported)* | func | `period`'s two loops | Replaces them (D-7) |
| `vocabkey.PropertyLifespan`, `PropertyLifespanSpan`, `EventBirth`, `EventDeath`, `EventBurial`, `RoleSubject` *(new)* | const | `vocabkey.PropertyName` and the other seeded Property keys; `core/eventtitle` `RuleSubject`; literals in `conclusionheaders/compose.go` | The keys' home; `RuleSubject` is an event-title rule name, not the term (D-3) |
| `subjectprofile.LifeEvent`, `Entity.Life` *(new)* | type, field | `subjectprofile.Candidate`; `conclusionheaders` `lifePick` | `lifePick` is a header's database read; the builder takes plain keys and dates (D-11) |
| `lifespan` in `core/subjectprofile` *(new, unexported)* | func | `subjectcompare.period` | Builds a window and a span from events, not one range from two properties; shares `EndsDate`, `ShiftYears` and the end orders (D-12) |
| `TestLifespan`, `TestLifespanFromEvents`, `TestShiftYears`, `TestDateEndOrder` *(new)*; rows in `TestBuild`, `TestEntityCompare`, `TestInvariants` | test | S11-10 `TestCompare`, S11-11 `TestBuild` | Same shapes (D-13) |
| `sameValues`, `groups` *(new, test)* | helper | `kindGroups` | Replaces it (D-10) |

**Builds on.** `valuecompare.DateEnds`, `EndsDate` (S11-10), `Window`; `propertycompare.Compare`; `subjectcompare.Compare` and its registry-declared special cases (S11-10); `subjectprofile.Build` / `ForSubject` / `value` (S11-11); the inventory guard and the seed check; the `add-match-rule` skill.
**Deletes.** `kindGroups` (now built from the registry); `period`'s duplicated loop; the inventory rows `subject.person.lifespan.MaxYears`, `ParentMinYears`, `ParentMaxYears`, `MotherMaxYears`, `MarriageMinYears`; `notYetBuilt`'s two lifespan prefixes; the per-bridge example list in `comparison-layers.md` §Lifespan.
**Next adds.** S11-13's loader reads each person's participations into `[]LifeEvent` (event type term key, role term key, the event's displayed date values or its period) and passes `reg.Subject.Lifespan` to `Build` and `ForSubject`. An anchor (another event type that starts or ends a life) is one `Starts` / `Ends` entry. S11-14's one-cardinality death segment covers a second death date. S11-19 reads the lifespan `PropertyResult` (window and span as the speaking pair) for an evidence line.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `datewindow.go`: `ShiftYears`, `DateEnd.StartsBefore`, `DateEnd.EndsAfter` (D-6) |
| `core/vocabkey` | Six constants (D-3) |
| `core/subjectcompare` | `subjectcompare.go`: `Lifespan`, `EventRole`, `Rules.Lifespan`, the branch in `Compare` (D-4, D-5); `lifespan.go` *(new)*; `pairs.go`: `pickEnd` (D-7) |
| `core/matchregistry` | `subject.go` defaults (D-8); tests (D-9) |
| `core/database/subjectvocab` | `subject_section_test.go` only (D-9) |
| `core/subjectprofile` | `LifeEvent`, `Entity.Life`, the new arguments, `lifespan` (D-11, D-12) |

```go
// subjectcompare

// EventRole names an event a person takes part in: its event_type term key
// and the person's participation role term key.
type EventRole struct{ EventType, Role string }

// Lifespan is the registry's lifespan rule for one Subject kind. The profile
// builder derives two values for that kind: Window, where its life can lie,
// and Span, its first to last dated event. Compare checks each side's span
// against the other side's window under the Window lead's weights.
type Lifespan struct {
	Kind     vocabkey.SubjectType
	Window   vocabkey.PropertyKey // a lead in Weights[Kind]
	Span     vocabkey.PropertyKey // not a lead
	MaxYears int                  // pads an end no Starts or Ends event fixes
	Starts   []EventRole          // events that fix the start
	Ends     []EventRole          // events that fix the end
}

// in Rules, after Periods:
	// Lifespan is the derived lifespan for one kind (people).
	Lifespan Lifespan

// subjectprofile

// LifeEvent is one event a person takes part in: its event_type term key, the
// person's participation role term key, and every date the event records.
type LifeEvent struct {
	EventType, Role string
	Dates           []datevalue.Value
}

type Entity struct {
	Properties  map[vocabkey.PropertyKey]Property
	NamePattern string
	Life        []LifeEvent // the person's events; a lifespan is built from them
}

func Build(e Entity, metas subjectcompare.Metas, reconcile valuecompare.Rules, life subjectcompare.Lifespan, opts Options) (subjectcompare.Values, error)
func ForSubject(values map[vocabkey.PropertyKey][]autoreconcile.Candidate, events []LifeEvent, namePattern string, life subjectcompare.Lifespan) subjectcompare.Values
```

`Build`'s and `ForSubject`'s doc comments each gain one sentence: "life is the registry's Subject.Lifespan; when it has a Window and the events give one, the profile also holds the lifespan's window and span." The package doc's "no numbers" stays true: `MaxYears` comes in as an argument.

### FFI

None — no caller yet.

### macOS

None — no app change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Date-end helpers and the period fold**
Files: `core/valuecompare/datewindow.go`, `core/valuecompare/datewindow_test.go`, `core/subjectcompare/pairs.go`.
Do: first run `go test ./core/subjectcompare/... -v | grep -c -- '--- PASS'` and write the count down. D-6 with `TestShiftYears` (T-1: 1850 year precision −110 → 1740, +110 → 1960; 1904-02-29 +10 → 2014-02-28; 1901-05-01 +110 → 2011-05-01 with `Lo` equal to `datevalue.DaysFromCivil(2011, 5, 1)`; an open end → unchanged, ok) and `TestDateEndOrder` (T-1: lower `Lo` starts before; equal `Lo`, lower `Hi` starts before; higher `Hi` ends after; equal `Hi`, higher `Lo` ends after; equal ends → false both ways). D-7.
Check: `go test ./core/valuecompare/... ./core/subjectcompare/... ./core/matchregistry/...` passes with no tags; the `subjectcompare` PASS count is unchanged.

**STEP-2 — The lifespan lead in layer 3 and the registry**
Files: `core/vocabkey/vocabkey.go`, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/lifespan.go` *(new)*, `core/subjectcompare/lifespan_test.go` *(new)*, `core/subjectcompare/subjectcompare_test.go`, `core/matchregistry/subject.go`, `core/matchregistry/subject_test.go`, `core/matchregistry/registry_test.go`, `core/matchregistry/inventory_test.go`, `core/database/subjectvocab/subject_section_test.go`, `docs/matching-model.md` (inventory rows only). Skill: `add-match-rule` ("Add a knob", "Add a Subject property or pair").
Do: D-3, D-4, D-5, D-8, D-9. `TestLifespan` with US-1 rows AC-1.1 – AC-1.9 and AC-1.11 (T-2). In `TestCompare`'s existing "both sides empty" row, whose check reads `slices.Equal(keys(r), []vocabkey.PropertyKey{name, sex})`, the expected keys become `{vocabkey.PropertyLifespan, name, sex}` (AC-1.10); no other existing row changes. Inventory, in `docs/matching-model.md` §Knob inventory, after the `subject.Weights.person.lifespan.Conflict` row: delete the five `subject.person.lifespan.*` rows and add, in this order:
- `` | `subject.Lifespan.Kind` | `person` | decided: people only (the lifespan design) | The kind that gets a lifespan | ``
- `` | `subject.Lifespan.Window` | `lifespan` | derived | The lead holding where a life can lie | ``
- `` | `subject.Lifespan.Span` | `lifespan_span` | derived | The value holding a life's first to last dated event | ``
- `` | `subject.Lifespan.MaxYears` | 110 | estimate: longest plausible human life | Years an unfixed start or end is padded from the first or last dated event | ``
- `` | `subject.Lifespan.Starts` | `birth` · `subject` | decided 2026-10-18 | Events that fix the start of a life | ``
- `` | `subject.Lifespan.Ends` | `death` · `subject`, `burial` · `subject` | decided 2026-10-18 | Events that fix the end of a life | ``
`notYetBuilt` loses `"subject.Weights.person.lifespan."` and `"subject.person.lifespan."`.
Check: `go test ./core/vocabkey/... ./core/subjectcompare/... ./core/matchregistry/...` passes with no tags; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` passes; `rg -n 'subject\.person\.lifespan' docs/matching-model.md` prints nothing.

**STEP-3 — Lead groups from the registry**
Files: `core/subjectcompare/invariants_test.go`.
Do: D-10 (T-3). Temporarily delete the `lifespan` entry from `sameValues`, run the test, and confirm it fails with "lead person.lifespan has no test value in sameValues"; restore it.
Check: `go test ./core/subjectcompare/... -run TestInvariants -v` passes and lists the six person subtests, from `AC-3.1/person` to the last; `rg -n 'kindGroups' core/subjectcompare` prints nothing.

**STEP-4 — The builder's lifespan**
Files: `core/subjectprofile/subjectprofile.go`, `core/subjectprofile/helpers_test.go`, `core/subjectprofile/lifespan_test.go` *(new)*, `core/subjectprofile/subjectprofile_test.go`, `core/subjectprofile/entity_test.go`, `core/subjectprofile/invariants_test.go`.
Do: D-11, D-12. Every existing `Build` call passes `subjectcompare.Lifespan{}` before `opts`; every existing `ForSubject` call passes `nil` events after `values` and `subjectcompare.Lifespan{}` last, so no existing case changes. `TestLifespanFromEvents` (T-4) with US-2 rows except AC-2.14; AC-2.14 as a `TestBuild` row (T-4).
Check: `go test ./core/subjectprofile/...` passes with no tags; `go vet ./core/subjectprofile/...` is clean; `go test ./core/matchregistry/ -run TestLayerPackages` passes.

**STEP-5 — Entity compare and invariants**
Files: `core/subjectprofile/entity_test.go`, `core/subjectprofile/invariants_test.go`.
Do: US-3 rows in `TestEntityCompare` (T-5): side A `Build(entity, subjectcomparetest.Metas(), reg.Value, reg.Subject.Lifespan, opts)`, side B `ForSubject(subject, subjectLife, "", reg.Subject.Lifespan)`, then `subjectcompare.Compare(vocabkey.Person, a, b, subjectcomparetest.Metas(), reg.Subject, reg.Value)`; each row's `check` asserts the `lifespan` `PropertyResult` outcome. AC-4.3 subtest (T-6).
Check: `go test ./core/subjectprofile/... -v -run 'TestEntityCompare|TestInvariants'` passes; then `go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/vocabkey/... ./core/matchregistry/...` passes with no tags.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `docs/deployment-plan/spike-11/deployment-plan.md`, `.agents/skills/add-match-rule/SKILL.md`, `docs/deployment-plan/spike-11/completed.md`, this plan (status).
Do:
- `comparison-layers.md` §Lifespan: replace the bullet that starts "**The subject profile builder derives a lifespan**" (with its nested examples and "Maximum lifespan, minimum and maximum parent ages are registry numbers.") by: "- **Decided (2026-10-18): a timeline, not per-bridge rules.** The profile builder lays every dated event a person takes part in on one line, any event type and any role. Its **span** runs from the first dated event to the last. Its **window** starts at the person's own birth or, with none, `MaxYears` before the first dated event, and ends at the person's own death or burial or, with none, `MaxYears` after the last. Age rules (marriage, parents) are dropped: the lifespan guides matching and is not an authority." Replace the bullet that starts "**Subject comparison compares the lifespans**" by: "- **Subject comparison checks each side's span against the other side's window** with value compare's date relations. `Different` either way is a conflict (died 1840 against born 1850; born 1850 against died 1990; resided 1950 against died 1900); anything else, `Near` included, is consistent." After the last bullet, a paragraph *(new)*: "**Built in S11-12.** `subjectprofile.Build` and `ForSubject` take the person's life events (event type, role, dates) and the registry's `subject.Lifespan`, and add two derived values, `lifespan` (the window) and `lifespan_span`. `subjectcompare.Compare` scores the `lifespan` lead from both directions: a conflict loses `subject.Weights.person.lifespan.Conflict`, consistent lifespans earn its `Agree`."
- `matching-model.md` §Glossary, the **lifespan** row becomes "| **lifespan** | A person's derived window and span, built from the dates of every event they take part in; each side's span is checked against the other's window |". §Worked example step 3: "Both sides also derive a lifespan, "born 1 May 1901", from their birth; the windows are consistent" becomes "Both sides also derive a lifespan from their birth: window 1 May 1901 to 1 May 2011, span 1 May 1901; each span lies in the other's window".
- `deployment-plan.md`: in the working rule that starts "**Lifespan is for people only**", keep the bold text and its link, and replace the text after the link's closing parenthesis and colon, "derived from direct neighbors' dates by registry rules; an impossible overlap is a large conflict, agreement weighs almost nothing.", by "a window and a span from every dated event the person takes part in (user, 2026-10-18); a span outside the other side's window is a large conflict, agreement weighs almost nothing." In §S11-12, **Ships.** becomes "The profile builder derives a person's lifespan from every dated event they take part in: birth starts it, death or burial ends it, `MaxYears` pads an open end; subject compare checks each side's span against the other's window, with a large conflict penalty and near-zero agree weight (user, 2026-10-18)."; **Not this PR.** becomes "Events and places. Age rules and two-hop rules. Loading life events (S11-13)."; **Check.** becomes "An entity who died in 1840 against a Subject born in 1850 conflicts; born 1850 / died 1990 conflicts; died 1900 / resided 1950 conflicts; born 1850 / died 1920 is consistent."
- `add-match-rule` SKILL.md: at the end of the Layer 3 bullet, append "The lifespan is a registry-declared special case: `subject.Lifespan` names the kind, its `Window` lead and its `Span`; `Compare` checks each side's span against the other's window."; in the profile builder bullet, "The builder has no registry section." becomes "The builder reads `subject.Lifespan` (passed in) to build a person's window and span from its life events."
- `completed.md`, newest first: "## S11-12 — Lifespan for people" / "**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). People now have a lifespan. The profile builder lays every dated event a person takes part in on one timeline; birth starts it, death or burial ends it, and an open end is padded by 110 years. Subject compare checks each side's dated events against the other side's window, so a person who died before a Subject was born, or a 140-year life, is a large conflict. The four age-rule knobs were dropped. No caller yet; nothing changes for researchers." with the date and PR number filled in.
- Map row status `Landed #N`; this plan's Status `Landed in [#N](…) (YYYY-MM-DD)` and a **Landed** note under the header listing each difference from the plan (or "None.").
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-12-lifespan.md` reports no errors; `rg -n 'ParentMinYears|MotherMaxYears|MarriageMinYears' docs/matching-model.md docs/deployment-plan/spike-11/comparison-layers.md core` prints nothing; `go test ./core/matchregistry/...` passes (`TestKnobInventoryMatchesRegistry`).

Commits: one per step, in step order. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Check a person's lifespan when two people are compared."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/datewindow_test.go` `TestShiftYears`, `TestDateEndOrder` *(new)* | STEP-1's cases | D-6 |
| T-2 | Go | `core/subjectcompare/lifespan_test.go` `TestLifespan` *(new)*; `subjectcompare_test.go` `TestCompare` (keys row) | Rows AC-1.1 – AC-1.11 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, D-2, D-5 |
| T-3 | Go | `core/subjectcompare/invariants_test.go` `TestInvariants` | Groups from the registry; person subtests with the lifespan group | AC-4.1, AC-4.2, D-10 |
| T-4 | Go | `core/subjectprofile/lifespan_test.go` `TestLifespanFromEvents` *(new)*; `subjectprofile_test.go` `TestBuild` row AC-2.14 | Rows AC-2.1 – AC-2.15 | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-2.10, AC-2.11, AC-2.12, AC-2.13, AC-2.14, AC-2.15, D-1, D-11, D-12 |
| T-5 | Go | `core/subjectprofile/entity_test.go` `TestEntityCompare` | Rows AC-3.1 – AC-3.5 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5 |
| T-6 | Go | `core/subjectprofile/invariants_test.go` `TestInvariants` subtest `AC-4.3` *(new)* | 500 generated event sets → window contains span; shuffle-stable | AC-4.3, D-13 |
| T-7 | Go | `core/matchregistry` `TestKnobInventoryMatchesRegistry`, `TestSubjectSection`, `TestDefaultIsFresh`, `TestLayerPackages`; `core/database/subjectvocab` `TestSubjectSectionMatchesSeed` | New fields have rows; structure, freshness, purity and the seed hold | D-8, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestCompare`'s person rows that list keys, `TestSubjectSectionMatchesSeed` (the derived lead), `TestKnobInventoryMatchesRegistry` (a row typo), `core/subjectprofile`'s existing tests after the signature change.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] go test ./core/subjectprofile/... ./core/subjectcompare/... ./core/propertycompare/... ./core/valuecompare/... ./core/vocabkey/... ./core/matchregistry/... passes with no build tags
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host without macOS every check but xcode-test passes, and CI runs the Xcode tests on the PR (this PR changes no Swift)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; core/autoreconcile, core/match and core/database's non-test files are unchanged
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none (no copy)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: comparison-layers.md, matching-model.md, deployment-plan.md (working rule, scope entry), add-match-rule skill (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing reaches it until S11-13. The reviewer runs `go test ./core/subjectprofile/... -v -run TestEntityCompare` with no build tags and reads subtests `AC-3.1 …` to `AC-3.4 …`: these are the map's Check (died 1840 / born 1850, born 1850 / died 1990 and died 1900 / resided 1950 conflict; born 1850 / died 1920 is consistent). Then `go test ./core/matchregistry/... -v -run TestKnobInventoryMatchesRegistry` to see the lifespan rows match the registry.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A date compare change after the planning base moves a relation §2 reads | §2's relations were read at `b799b67`; a drift is a stop condition, not a test edit |
| The lifespan lead changes existing person results | It is unknown (no points) whenever a side has no lifespan values, so only the key-list row changes (AC-1.10); any other changed row is a stop condition |
| A generated invariant case with two windows on one side breaks symmetry | D-5 returns a conflict from either direction; `same` compares outcomes and numbers only. A failure is a stop condition |

**Stop and report if:**

- `main` differs from §4: `subjectcompare.Rules`, `Compare`'s signature or loop, `period`, `subjectprofile.Build` / `ForSubject` signatures, `valuecompare.DateEnds` / `EndsDate`, `notYetBuilt`, or the inventory's lifespan rows. Report each difference.
- `propertycompare.Compare` gives a relation other than §2's conventions paragraph for one of its pairs: report the pair and the result; do not change the AC.
- An existing `TestCompare`, `TestBuild`, `TestForSubject` or `TestEntityCompare` row other than AC-1.10's changes its sim, ev or labels.
- `TestKnobInventoryMatchesRegistry` reports a key other than D-8's six `subject.Lifespan.*` keys and the two weight rows for the new fields.
- `core/subjectcompare` or `core/subjectprofile` fails `TestLayerPackages` (a numeric literal other than 0 and 1, or a new import beyond `core/propertymeta`, `core/datevalue` and the layer packages they already import).
- An invariant of US-4 fails: report the seed, case index and the values.
- `vocabkey` already declares any of D-3's names, or `subjectcompare.Rules` already has a `Lifespan` field.

## 12. Open questions

None.

## 13. PR

**Title:** S11-12 — Lifespan for people
**Body:** objective in one sentence; what shipped against each US; AC evidence (test names); deviations from the plan; manual check (the `go test` runs of §10); `Plan: docs/deployment-plan/spike-11/prs/S11-12-lifespan.md`.
