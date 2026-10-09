# Matching

How Provenencia scores how much one thing resembles each canonical handle (a Person, Event or Place), from the thing's own Properties. Promote's target suggestions (Spike 9 R7, **S9-10**) were the first consumer; the one-page Promote (S9-44) replaced them with graph alignment. In Promote graph alignment ([`promote-graph-alignment.md`](promote-graph-alignment.md) §4.1) matching is the **smaller, composable unit**: pairwise evaluation (and `Rank` over many handles) judges sameness from Properties; graph alignment **walks** the Evidence graph and calls that judgment per candidate, adding edge/structure. Matching is also the **property-only fallback** for Subjects no anchor reaches. Merge hints are another consumer: `ForEntity` already exists, but no UI uses it yet.

Code: [`core/match`](../core/match) (the algorithm) and [`core/database/matching`](../core/database/matching) (catalog adapter). Consumer: [`core/database/promotealign`](../core/database/promotealign), which reads `CandidatesOfType` once per kind and runs `Rank` in memory to pick each Subject's candidates.

`core/match` has one file per concern, each with its own test file:

| File | Holds |
| --- | --- |
| `match.go` | Values, candidates, `Score`, `Rank` |
| `registry.go` | **Every tunable number, the built-in name patterns, and the default profiles.** The one place to configure. |
| `profiles.go` | The `Feature` and `Profile` types |
| `compare.go` | The `Comparer` interface, `ComparerFor`, `Set`, and the simple comparers (text, term, integer) |
| `names.go` | `NameComparer`, roles, `NamePattern` and the `NamePatterns` source. No culture-specific logic. |
| `dates.go` | `DateComparer`: spans, precision, tolerance |
| `words.go` | Word handling shared by names and text: splitting, similarity, edit distance, best pairing |

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
  - `ForEntity` reads a handle's auto-reconciled values from the cache, at every rank.
- **Candidates:** the unmerged handles of the probe's Subject type, with their cached auto-reconciled values at every rank. The probe's own handle is never a candidate.
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
| `NameComparer` | name | **Parts only.** `form` is never read. A name with no parts is not comparable. |
| `TextComparer` | text | Same as names, but with no initials and `Partial` 0.7. |
| `TermComparer` | term (by key) | The same term scores 1, a different one 0. `Neutral` terms are not comparable. |
| `DateComparer` | date | Dates are spans of years. **Identical dates score 1** at any precision. **Two points** (exact or ABT): same month with a day missing on one side 0.9; same year with a month missing on one side 0.8; same month with different known days 0.7; different known months 0.6. Up to `Tolerance` years apart (default 2) falls off linearly from 0.8; beyond that scores 0. ABT doubles the tolerance. **A range (FROM/TO/BET) or a bound (BEF, AFT) on either side:** overlapping spans score by the wider span's width: 0.8 for one year, 0.05 less per extra year, never below 0.2. Open spans (BEF, AFT, FROM with no TO) score 0.2. Spans apart fall off from that score within the tolerance, and beyond it score 0, so BEF 1820 vs 1823 is a disagreement. An inverted range reads as its years. |
| `IntegerComparer` | integer | Equal scores 1. Within `Tolerance` falls off linearly. Beyond it scores 0. |

### Names: part types are data, not gates

`form` is the name as written, closer to a transcription. Typed parts are what can be computed on ([`structured-name-model.md`](structured-name-model.md)). `NameComparer` compares names **word by word**, and uses each word's part type as data:

- **Roles come from a name pattern.** A `NamePattern` holds everything culture-specific: which part type plays which role, and each role's word weight. The comparer holds none of it.
  - Each name takes its pattern from `Value.NamePattern` (its Person's name format, once loaders supply it); otherwise it uses the comparer's pattern.
  - Patterns come through a `NamePatterns` source. Today that's `BuiltinNamePatterns` in the registry, with only `western`; later it becomes a read of `name_format_profiles`.
  - Because each name uses its own pattern, names under different patterns still compare.

  | Role | Default part types | Word weight |
  | --- | --- | --- |
  | family | `surname` | 1.5 |
  | given | `given`, `initial` | 1 for the first, 0.5 for the rest |
  | nick | `nick` | 0.3 |
  | untyped | `undetermined`, or a part with no type | 1 |
  | generation | `suffix` | compared separately (below) |
  | ignored | `prefix` (titles), `surname_prefix` ("van") | — |

- **Any word can pair with any word** of the other name, one-to-one, using the exact best pairing (so scores are symmetric).
- **A pair earns** word similarity × role affinity × the two words' weights. The score is the earned share of both names' total weight, so an unmatched word costs its weight.
- **Role affinity:**
  - the same role: 1;
  - a nickname against a given name: 0.9;
  - a typed word against an untyped one: 0.8;
  - any other mismatch, such as a surname against a given name: 0.5.

  **Names entered in different formats still connect,** just lower. Types never block a pairing.
- **Surnames that share nothing:** when both names have family words and none resembles any word of the other name, the score is scaled by `GivenOnlyFactor` (0.5). A surname can change at marriage, but a shared "James" alone is weak. Swapped types are not a conflict.
- **Generations:** when both names carry suffixes and share none (Jr. vs Sr.), the score is scaled by `SuffixConflict` (0.3).
- **Words** match when equal (1), as an initial and the word it begins (0.5), or as a near spelling at or above `FuzzyFloor` (0.8, edit-distance ratio).
  - A swap of two adjacent letters counts as one edit (Robnis ~ Robins).
  - Words of three or more letters may also differ by one added or dropped letter (Ann ~ Anne, Jon ~ John: 0.75). A substituted letter in a short word is a different name (Mary ≠ Mark).
- **Normalization:** dashes and slashes separate words (Smith-Jones ~ Smith Jones). Apostrophes join (O'Brien ~ OBrien).
- **Surname conflict:** an initial doesn't count as resembling a surname, so "S." doesn't hide a Smith/Robins conflict.
  - An initial is a lone cased letter ("J").
  - A lone character in an uncased script (蒋, 王) is a whole word.
- **No parts:** nothing to compare. `form` is not a match input.

Every weight and affinity is a `NameComparer` field.

Worked examples:

| Pair | Score |
| --- | --- |
| James Robins vs James Robins (forms "ROBINS, Jas." and "James Robins") | 1 |
| J. Robins vs James Robins | 0.8 |
| James Robins vs James Kenneth Robins (parts; the forms can disagree) | 0.91 |
| Robins (surname only) vs James Robins | 0.75 |
| Mary Robins vs James Robins | 0.6 |
| surname "James", given "Robins" (types swapped) vs James Robins | 0.5 |
| James Robins Jr. vs James Robins Sr. | 0.3 |
| James Smith vs James Robins | 0.2 |

Tests: [`core/match/names_test.go`](../core/match/names_test.go) has three layers:
- **Exact scores per rule:** identity and normalization, surname, given, cross-format pairs (swapped types, typed vs untyped, profile-supplied roles), a surname conflict, suffix, ignored parts, and the tuning knobs. Expected values are written as the weighted arithmetic. A name with only a form is not comparable.
- **Ladders:** which name must outrank which, so they hold when weights are retuned.
- **Generated permutations from fixed seeds**, checked for:
  - symmetry;
  - a score between 0 and 1;
  - every name matching itself at 1;
  - the form not mattering once both names have comparable parts;
  - the order of parts across roles not mattering;
  - a suffix conflict never raising a score;
  - titles and surname particles never moving a score;
  - the same words with their types cleared still connecting (above 0, never above 1).

When a seed finds a bug, shrink it into an exact-score case.

Not yet handled:
- **Accents, Spanish dual surnames, patronymic, nicknames, Daitch–Mokotoff:** Spike 10, [`deployment-plan/spike-10/international-names.md`](deployment-plan/spike-10/international-names.md). Other patterns, transliteration, and calendars: [`ideas/international-names.md`](ideas/international-names.md).
- **Abbreviations, project name frequency, spaced particles, support weighting, date windows, event and place affinity:** Spike 10, [`deployment-plan/spike-10/name-matching-enhancements.md`](deployment-plan/spike-10/name-matching-enhancements.md). Known-as, concluded values, calibration, and rejected-claim memory: [`ideas/name-matching-enhancements.md`](ideas/name-matching-enhancements.md).

**Resolution uses the same parts.** The reconciler (`core/autoreconcile/names.go`) and `NameComparer` both ignore `form`. A name with no parts is no evidence there and not comparable here. The cache name `sort_key` (list order) is still the normalized form, and list and card text shows it. That sort key is not a match.

`match.ComparerFor(valueType)` gives the default comparer for any Property, so a profile can also weigh researcher-defined Properties.

## Default profiles

These live in [`core/match/registry.go`](../core/match/registry.go), with every other tunable number.

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
- **One registry:** [`core/match/registry.go`](../core/match/registry.go) holds everything:
  - word rules (`DefaultWordRules`)
  - name affinities and conflict factors (`DefaultNames`)
  - name patterns (`WesternNamePattern`, `BuiltinNamePatterns`)
  - date scores and span scoring (`DefaultDates`)
  - text and integer settings
  - neutral terms
  - default profiles
  - the suggestion limit

  Comparers fall back to these values, so a bare `NameComparer{}` scores exactly like `DefaultNames`, and no number is written twice.
- **Guard tests** (`registry_test.go`):
  - Every comparer setting must have a registry value.
  - Bare comparers must score like the registry's.
  - Default profiles must use the registry's comparers.
  - The Western pattern must give every product part type a role.
- **Comparer settings** are pointers set with `match.Set`. Nil takes the registry value, and zero is a real value: `NameComparer{CrossRole: match.Set(0.0)}` makes types strict, and `DateComparer{Tolerance: match.Set(0)}` requires the same year.
- **Not yet:** researcher-facing tuning, profiles persisted per project, and data-driven name patterns.
  - **Stored profiles** replace `DefaultProfile` in `matching.start`.
  - **A stored-pattern source** replaces `BuiltinNamePatterns` as `NameComparer.Patterns`.

  Both use the same shapes, so scoring code doesn't change.

## Extending

- **A new kind** needs a `DefaultProfile` case. Its consumer then needs a header for the kind (Promote suggestions carry `PersonHeader` for Persons, and the handle alone for Events and Places until S9-22 / S9-25).
- **A new signal on the thing itself** (a researcher-added Property, say) needs:
  - a Comparer, if no existing one fits;
  - a Feature in the profile;
  - the adapter putting those values into the probe and the candidates.
  Signals reached through edges (a Person's birth year through its birth event, its family) are **not** matching Features: Promote graph alignment compares neighbors directly, over the canonical graph (S9-28), with its own weights (S9-41).
- **Blocking:** candidates are currently every handle of the type with a cached profile value, read in one query (constant query count, but linear rows). When catalogs outgrow that, narrow candidates in `loadCandidates` using the handle search index (S9-34a: `catalog_search_docs` kinds `person` / `event` / `place`, kept current by the cache's upkeep). Scoring does not change.
- **Merge hints:** `matching.ForEntity(handle)` ranks the other handles of the same type. A merge-hint surface needs only to shape that result.
