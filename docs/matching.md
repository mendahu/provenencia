# Matching

How Provenencia scores how much one thing resembles each canonical handle (a Person, Event or Place). Promote's target suggestions (Spike 9 R7, **S9-10**) are the first consumer. Merge hints are the second: `ForEntity` already exists, but no UI uses it yet.

Code: [`core/match`](../core/match) (the algorithm) and [`core/database/matching`](../core/database/matching) (catalog adapter). Consumers: [`core/database/promotetargets`](../core/database/promotetargets).

## Shape

```text
probe ──┐                    ┌── Profile (per Subject type): Features + MinScore
        ├─ core/match.Rank ──┤
handles ┘                    └── Feature: Property · Comparer · Weight · Contradiction
                       │
                       ▼
            []Match { entity, ref, score, reasons[] }
```

- **Probe:** the thing being matched, as `match.Values` (Property → values).
  - `ForSubject` reads an Interpretation Subject's positive Observations.
  - `ForEntity` reads a handle's resolved values from the cache, at every rank.
- **Candidates:** the unmerged handles of the probe's Subject type, with their cached resolved values at every rank. The probe's own handle is never a candidate.
- **Pure core:** `core/match` has no catalog access. It can be tested with literal values, and can run anywhere a probe and candidates can be built.

## Scoring

Scores are **additive points**, so the weights read directly: "a matching name is worth 10, a different sex at birth costs 8".

For each Feature in the Profile:

1. Its Comparer judges every pair of probe and candidate values for that Property. The Feature takes the best similarity, from 0 to 1.
2. If no pair is comparable, the Feature is skipped: it neither helps nor hurts. A pair is not comparable when either side has no value, or the value is neutral (sex `unknown`, or a date with only a phrase).
3. If the similarity is above 0, the Feature adds `Weight × similarity`.
4. If the similarity is exactly 0 (a clear disagreement), the Feature subtracts `Contradiction`.

A candidate is returned when its total reaches `MinScore`. Results are sorted by score, then by ref. Every match carries its **reasons**: each Feature's similarity and contribution, for explaining it in the UI ("same name, born the same year") and for tuning.

## Comparers

| Comparer | Value type | Similarity |
| --- | --- | --- |
| `NameComparer` | name | Same normalized form (the cache's name `sort_key`) scores 1. Otherwise `Partial` (0.8) × a Dice overlap of best-paired words. An equal word counts 1. An initial matching the word it begins counts 0.5. A near spelling (edit ratio ≥ `FuzzyFloor`, 0.8) counts its ratio. |
| `TextComparer` | text | Same as names, but with no initials and `Partial` 0.7. |
| `TermComparer` | term (by key) | The same term scores 1, a different one 0. `Neutral` terms are not comparable. |
| `DateComparer` | date | Same day 1. Same month 0.9. Same year 0.8, or 0.6 if both months are known and differ. Up to `Tolerance` years apart (default 2) falls off linearly. Beyond that it scores 0. ABT/BEF/AFT on either side doubles the tolerance. |
| `IntegerComparer` | integer | Equal scores 1. Within `Tolerance` falls off linearly. Beyond it scores 0. |

`match.ComparerFor(valueType)` gives the default comparer for any Property, so a profile can also weigh researcher-defined Properties.

## Default profiles

These live in [`core/match/profiles.go`](../core/match/profiles.go). They are the one place to tune.

| Kind | Features (Weight / Contradiction) | MinScore |
| --- | --- | --- |
| person | name 10 / 0 · sex_at_birth 1 / 8 (unknown, indeterminate neutral) | 3 |
| event | event_type 4 / 6 · date 6 / 4 · start_date 3 / 2 · end_date 3 / 2 | 5 |
| place | toponym 10 / 0 | 3 |

Worked examples:
- **Person:** "James Robins" against "Mary Robins" scores 10 × 0.4 = 4, which is shown but low. The same name with a different sex at birth scores 10 − 8 = 2, which is hidden.
- **Event:** the same event type alone (every birth) scores 4, which is hidden. The same type plus the same year scores 4 + 4.8.

## Configuring

- **Per call:** `matching.Options{Profile: &p}` replaces the default profile. `Profile.With(feature)` copies a default and replaces or adds one Feature.
- **Shipped defaults:** edit `DefaultProfile`. Every consumer reads it unless it passes its own Profile.
- **Not yet:** researcher-facing tuning, or profiles persisted per project. When that lands, `matching.start` is where a stored profile replaces the default.

## Extending

- **A new kind** needs a `DefaultProfile` case. Its consumer then needs a header for the kind (Promote suggestions carry `PersonHeader` for Persons, and the handle alone for Events and Places until S9-22 / S9-25).
- **A new signal** (birth year for a Person, through its events; family context during a walk) needs:
  - a Comparer, if no existing one fits;
  - a Feature in the profile;
  - the adapter putting those values into the probe and the candidates.
  Edge-derived values arrive with S9-28 / S9-29. Related-first ordering during a walk (S9-29) is a Feature or a boost, not a separate path.
- **Blocking:** candidates are currently every handle of the type with a cached profile value, read in one query (constant query count, but linear rows). When catalogs outgrow that, narrow candidates in `loadCandidates` using the R8 search index (S9-34). Scoring does not change.
- **Merge hints:** `matching.ForEntity(handle)` ranks the other handles of the same type. A merge-hint surface needs only to shape that result.
