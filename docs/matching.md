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
| `NameComparer` | name | **By typed parts** when both names have a surname or given part (see below). Otherwise it falls back to `form`: the same normalized form scores 1, else `Partial` (0.8) × the overlap of best-paired words. |
| `TextComparer` | text | Same as names, but with no initials and `Partial` 0.7. |
| `TermComparer` | term (by key) | The same term scores 1, a different one 0. `Neutral` terms are not comparable. |
| `DateComparer` | date | Same day 1. Same month 0.9. Same year 0.8, or 0.6 if both months are known and differ. Up to `Tolerance` years apart (default 2) falls off linearly. Beyond that it scores 0. ABT/BEF/AFT on either side doubles the tolerance. |
| `IntegerComparer` | integer | Equal scores 1. Within `Tolerance` falls off linearly. Beyond it scores 0. |

### Names are compared by parts, not by `form`

`form` is the name as written, closer to a transcription. Typed parts are what can be computed on ([`structured-name-model.md`](structured-name-model.md)). When both names have typed parts, `NameComparer` compares them by role:

| Role | Parts | Rule |
| --- | --- | --- |
| Surname | `surname` | Best-paired word overlap, so dual surnames partly match. `surname_prefix` ("van") is not compared. |
| Given | `given`, `initial`, plus `nick` as an alternative first name | The first given name (or a nickname, on either side) counts 70%; the whole given set, initials included, counts 30%. |
| Suffix | `suffix` | Both present and different (Jr. vs Sr.) multiplies the score by `SuffixConflict` (0.3). |

- **Blending:** the surname is `SurnameShare` (0.6) of the score and the given name the rest. A role only one side has gets half credit. A role neither side has is left out, so "James" vs "James" scores 1.
- **Surnames that share nothing:** the given-name match counts at `GivenOnlyFactor` (0.5). A surname can change at marriage, but a shared "James" alone is weak.
- **Words** match when equal (1), as an initial and the word it begins (0.5), or as a near spelling at or above `FuzzyFloor` (0.8, edit-distance ratio).
- **Initials** are lone cased letters ("J"). A lone character in an uncased script (蒋, 王) is a whole word, so single-character names match only themselves.
- **Pairing:** word lists are paired by the exact best one-to-one pairing, so scores are symmetric.
- **Titles and untyped parts** (`prefix`, `undetermined`, no type) carry no role.
- **When it falls back to `form`:** a name with no surname or given part on either side.

Worked examples, all scored by parts:

| Pair | Score |
| --- | --- |
| James Robins vs James Robins (forms "ROBINS, Jas." and "James Robins") | 1 |
| Robins (surname only) vs James Robins | 0.8 |
| J. Robins vs James Robins | 0.8 |
| Mary Robins vs James Robins | 0.6 |
| James Robins Jr. vs James Robins Sr. | 0.3 |
| James Smith vs James Robins | 0.2 |

Tests: [`core/match/names_test.go`](../core/match/names_test.go) has three layers:
- **Exact scores per rule:** identity and normalization, surname, given, a surname conflict, suffix, role-less parts, the form fallback, and the tuning knobs.
- **Ladders:** which name must outrank which, so they hold when weights are retuned.
- **Generated permutations from fixed seeds**, checked for:
  - symmetry;
  - a score between 0 and 1;
  - every name matching itself at 1;
  - the form not mattering once both names have parts;
  - the order of parts across roles not mattering;
  - a suffix conflict never raising a score;
  - titles and untyped parts never moving a score.

When a seed finds a bug, shrink it into an exact-score case.

Not yet handled:
- Nickname equivalence (Jim ↔ James) and phonetic matching.
- Accent folding: "José" vs "Jose" is not a match. Normalization keeps diacritics, and the two words are too short for the near-spelling floor.
- Patronymics, and name changes as their own signals.
- Culture-specific roles from name format profiles.

**Where `form` still rules.** Resolver clustering and the cache name `sort_key` (so list order) key on normalized `form` until **S9-13**. List and card text shows `form` until the name-format work. Both are tracked there; matching does not depend on them.

`match.ComparerFor(valueType)` gives the default comparer for any Property, so a profile can also weigh researcher-defined Properties.

## Default profiles

These live in [`core/match/profiles.go`](../core/match/profiles.go). They are the one place to tune.

| Kind | Features (Weight / Contradiction) | MinScore |
| --- | --- | --- |
| person | name 10 / 0 · sex_at_birth 1 / 8 (unknown, indeterminate neutral) | 3 |
| event | event_type 4 / 6 · date 6 / 4 · start_date 3 / 2 · end_date 3 / 2 | 5 |
| place | toponym 10 / 0 | 3 |

Worked examples:
- **Person:** "James Robins" against "Mary Robins" scores 10 × 0.6 = 6 with typed parts (10 × 0.4 = 4 by form alone), which is shown but below the same name. The same name with a different sex at birth scores 10 − 8 = 2, which is hidden.
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
