# Matching

> **Spike 11 replaces this page.** The target model is [matching-model.md](matching-model.md); until then this page describes the running code.

How Provenencia scores how much one thing resembles each canonical handle (a Person, Event or Place), from the thing's own Properties. Promote's target suggestions (Spike 9 R7, **S9-10**) were the first consumer; the one-page Promote (S9-44) replaced them with graph alignment. In Promote graph alignment ([`promote-graph-alignment.md`](promote-graph-alignment.md) §4.1) matching is the **smaller, composable unit**: pairwise evaluation (and `Rank` over many handles) judges sameness from Properties; graph alignment **walks** the Evidence graph and calls that judgment per candidate, adding edge/structure. Matching is also the **property-only fallback** for Subjects no anchor reaches. Merge hints are another consumer: `ForEntity` already exists, but no UI uses it yet.

Code: [`core/match`](../core/match) (the algorithm) and [`core/database/matching`](../core/database/matching) (catalog adapter). Consumer: [`core/database/promotealign`](../core/database/promotealign), which reads each kind's handles from the in-memory catalog graph (`graphcache`) once and runs `Rank` over them to pick each Subject's candidates, and reads layer values with `SubjectsValues`.

`core/match` has one file per concern:

| File | Holds |
| --- | --- |
| `match.go` | Values (with the cache evidence a handle value carries), candidates, `Score`, `Rank`, `Reason` |
| `registry.go` | The default profiles and the period map. Value numbers live in `core/matchregistry` and arrive as `Rules`. |
| `profiles.go` | The `Feature` and `Profile` types |
| `pair.go` | Pairwise evaluation: `Evaluate`, `OutcomeOf`, outcomes, `PropertyMeta`, `PropertyComparison` (graph alignment's per-pair unit) |
| `pairs.go` | Pair selection: `pairValues`, support weights, negatives, `ValueRef`, `Value.Label` and `Value.Auto` |
| `compare.go` | `Compare`: one value pair through `core/valuecompare` under the caller's `Rules` |
| `periods.go` | `poolPeriods`: an Event's `date` against another's `start_date` and `end_date` |

## Shape

```text
probe ──┐                    ┌── Profile (per Subject type): Features + MinScore
        ├─ core/match.Rank ──┤
handles ┘                    └── Feature: Property · ValueType · Weight · Contradiction
                       │
                       ▼
            []Match { entity, ref, score, reasons[] }
```

- **Probe:** the thing being matched, as `match.Values` (Property → values).
  - `ForSubject` reads an Interpretation Subject's Observations, negative ones marked; only a positive value earns credit.
  - `ForEntity` reads a handle's auto-reconciled values from the cache, at every rank.
- **Candidates:** the unmerged handles of the probe's Subject type, with their cached auto-reconciled values at every rank. The probe's own handle is never a candidate.
- **Pure core:** `core/match` has no catalog access. It can be tested with literal values, and can run anywhere a probe and candidates can be built.

## Scoring

Scores are **additive points**, so the weights read directly: "a matching name is worth 10, a different sex at birth costs 2".

For each Feature in the Profile:

1. Value compare judges every pair of probe and candidate values for that Property, by the Feature's value type. The Feature takes the best similarity, from 0 to 1.
2. If no pair is comparable, the Feature is skipped: it neither helps nor hurts. A pair is not comparable when either side has no value, or the value is neutral (sex `unknown`, or a date with only a phrase).
3. If the similarity is above 0, the Feature adds `Weight × similarity`.
4. If the similarity is exactly 0 (a clear disagreement), the Feature subtracts `Contradiction`.

A candidate is returned when its total reaches `MinScore`. Results are sorted by score, then by ref. Every match carries its **reasons**: each Feature's similarity and contribution, for explaining it in the UI ("same name, born the same year") and for tuning.

## Comparing values

Values compare with [`core/valuecompare`](../core/valuecompare) under the registry's value section ([`matching-model.md`](matching-model.md) layer 1 and knob inventory): names word by word with roles and weights, text and toponyms in word order, terms by key, integers, and dates by the days they cover (`ABT 1817` reads as 1816 – 1818). `match.Compare(rules, valueType, a, b)` turns a `match.Value` into a layer-1 value and returns the similarity. A `Neutral` term and an `Unknown` relation are not comparable.

The catalog's data arrives in `Rules`: the stored default name pattern under its key (a name value carries that key in `NamePattern`) and the alias tables. `valuerules.Load` ([`core/database/valuerules`](../core/database/valuerules)) builds them; a caller passes the rules to `Rank`, `Score` and `Evaluate`.

Typed parts are what a name compares on ([`structured-name-model.md`](structured-name-model.md)); `form` is never read. A name with words on one side only is a resemblance below a full match: `Robins` against `James Robins` is 0.75. A different first given name or surname on each side is `Different`, similarity 0. A common surname costs nothing extra: there is no frequency weighting.

**Support weights and negatives.** A handle's cached values carry the support the auto-reconciler counted (`support`, `against`, `reason`), and a Subject's Observations carry a polarity. `pairValues` (`pairs.go`) decides which value on each side speaks for a Feature, and both `Score` and `Evaluate` read their pair through it.
- **Weight.** A value's effective support is `support − against` (at least 1 for a displayed `kept` value, at least 0 for any other). The **leaders** are the displayed values with the highest effective support, or when none is displayed, the values that share the highest. A value's weight is its effective support over the leaders', at most 1. A neutral, `denied`, or `provisional` value, or one with no support left, never votes. Two equally supported names both lead, so a maiden and a married name each match in full.
- **Contradictions.** A negative Observation that is `autoreconcile.Compatible` with a leader on the other side contradicts it. A value compatible with a `denied` value that nothing else on that side backs contradicts it. When both sides have leaders and every comparable leader pair scores 0, that is a contradiction, so a minority that agrees does not rescue a mismatch. A negative never raises a score; a probe with only negatives has nothing to compare and the Feature is skipped.
- **Multiple cardinality.** A Feature whose Property is `multiple` (a Place's toponyms) gives every voting value full weight and skips the leader rule, so a handle's minority toponym still matches. A negative or a `denied` value still contradicts.
- **Credit.** Otherwise the best `similarity × weight × weight` among voting pairs, reported as `similarity × weights`. A handle or probe with no cache evidence weighs every value 1 and scores as before.
- **Reasons.** `Reason` and `PropertyComparison` carry `Probe` and `Candidate` (`ValueRef`: label, negative flag, the date of a date value, and the Observation id of a probe value). A term's label is its catalog label (`Value.TermLabel`), and the tie-break between equal values uses the term's key so a rename moves no reason. `attachExhibits` builds every Promote row's sheet lines: a row with no exhibit gets one line per comparison that spoke, and a row with exhibits gets a line for each row-level conflict no exhibit shows. These lines cannot be pinned, and a negative probe reads "not <value>".

| Probe | Handle values (support) | Result |
| --- | --- | --- |
| female | male (3), female (1) | Contradiction: −2; the reason names `female` and `male` |
| male | male (3), female (1) | +1 |
| female | male (2), female (2) | +1: both lead |
| James Robins | Mary Smith (5), James Robins (1) | Name similarity 0, a conflict in Promote; not a full match |
| not male (negative) | male (3), female (1) | Contradiction: −2 |

Promote's exhibit chooser keeps its own member picker, so a row can read conflict while an exhibit line shows an agreeing pair for a minority value.

Not yet handled:
- **Daitch–Mokotoff:** [`ideas/daitch-mokotoff.md`](ideas/daitch-mokotoff.md). Descoped from Spike 10 (S10-26). Spanish dual surnames, a patronymic part, other patterns, transliteration, Beider–Morse, and calendars: [`ideas/international-names.md`](ideas/international-names.md).
- **Abbreviations, event and place affinity:** Spike 10, [`deployment-plan/spike-10/name-matching-enhancements.md`](deployment-plan/spike-10/name-matching-enhancements.md). Known-as, concluded values, calibration, and rejected-claim memory: [`ideas/archive/name-matching-enhancements.md`](ideas/archive/name-matching-enhancements.md).

**Resolution uses the same parts.** The reconciler (`core/autoreconcile/names.go`) and value compare both ignore `form`. A name with no parts is no evidence there and not comparable here. The People list also ignores `form`: it orders by the profile's sorted part order. The cache name `sort_key` is still the normalized form. That column is not a match and not the list order.

A Feature names its Property's value type, so a profile can also weigh researcher-defined Properties.

## Default profiles

These live in [`core/match/registry.go`](../core/match/registry.go). The value-section numbers (words, names, text, integers, dates) live in [`core/matchregistry`](../core/matchregistry).

| Kind | Features (Weight / Contradiction) | MinScore |
| --- | --- | --- |
| person | name 10 / 0 · sex_at_birth 1 / 2 · gender 1 / 2 (neutral terms not compared) | 3 |
| event | event_type 4 / 6 · date 6 / 4 · start_date 3 / 2 · end_date 3 / 2 | 5 |
| place | toponym 10 / 0 | 3 |

Worked examples:
- **Person:** "Mary Robins" against "James Robins" is `Different`: 0, hidden; "Robins" against "James Robins" is 7.5, shown. The same name with a different sex at birth scores 10 − 2 = 8: the mismatch narrows the score and does not hide a name that cleared MinScore.
- **Event:** the same event type alone (every birth) scores 4, which is hidden. The same type plus the same year scores 4 + 4.8.

## Configuring

- **Per call:** `matching.Options{Profile: &p}` replaces the default profile. `Profile.With(feature)` copies a default and replaces or adds one Feature.
- **One registry:** the value numbers (words, names, text, integers, dates) live in [`core/matchregistry`](../core/matchregistry) and reach `match` as `Rules`; the default profiles live in [`core/match/registry.go`](../core/match/registry.go).
- **Neutral terms** are not in the registry: a term flagged `property_terms.neutral` (seeded on *unknown* and *indeterminate* sex at birth, [`seeded-vocabulary.md`](seeded-vocabulary.md) §3.7) reaches `match.Compare` as `Value.Neutral`, and it is not comparable there.
- **Not yet:** researcher-facing tuning and profiles persisted per project. Stored profiles replace `DefaultProfile` in `matching.start`, and in promotealign and graphalign, which read it directly.

## Extending

- **A new kind** needs a `DefaultProfile` case. Its consumer then needs a header for the kind (Promote rows and alternatives carry the Person, Event or Place header).
- **A new signal on the thing itself** (a researcher-added Property, say) needs:
  - a value type `core/valuecompare` compares, if no existing one fits ([`add-match-rule`](../.agents/skills/add-match-rule/SKILL.md));
  - a Feature in the profile;
  - the adapter putting those values into the probe and the candidates.
  Signals reached through edges (a Person's birth year through its birth event, its family) are **not** matching Features: Promote graph alignment compares neighbors directly, over the canonical graph (S9-28), with its own weights (S9-41).
- **Blocking:** candidates are currently every handle of the type with a cached profile value, read in one query (constant query count, but linear rows). When catalogs outgrow that, narrow candidates in `loadCandidates` (and in promotealign's `loadCanon`, which reads the same set from the graph with `g.Kind`) using the handle search index (S9-34a: `catalog_search_docs` kinds `person` / `event` / `place`, kept current by the cache's upkeep).
- **Merge hints:** `matching.ForEntity(handle)` ranks the other handles of the same type. A merge-hint surface needs only to shape that result.
