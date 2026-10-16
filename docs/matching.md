# Matching

How Provenencia scores how much one thing resembles each canonical handle (a Person, Event or Place), from the thing's own Properties. Promote's target suggestions (Spike 9 R7, **S9-10**) were the first consumer; the one-page Promote (S9-44) replaced them with graph alignment. In Promote graph alignment ([`promote-graph-alignment.md`](promote-graph-alignment.md) §4.1) matching is the **smaller, composable unit**: pairwise evaluation (and `Rank` over many handles) judges sameness from Properties; graph alignment **walks** the Evidence graph and calls that judgment per candidate, adding edge/structure. Matching is also the **property-only fallback** for Subjects no anchor reaches. Merge hints are another consumer: `ForEntity` already exists, but no UI uses it yet.

Code: [`core/match`](../core/match) (the algorithm) and [`core/database/matching`](../core/database/matching) (catalog adapter). Consumer: [`core/database/promotealign`](../core/database/promotealign), which reads each kind's handles from the in-memory catalog graph (`graphcache`) once and runs `Rank` over them to pick each Subject's candidates, and reads layer values with `SubjectsValues`.

`core/match` has one file per concern:

| File | Holds |
| --- | --- |
| `match.go` | Values (with the cache evidence a handle value carries), candidates, `Score`, `Rank`, `Reason` |
| `registry.go` | **Every tunable number, the built-in name patterns, and the default profiles.** The one place to configure. |
| `profiles.go` | The `Feature` and `Profile` types |
| `pair.go` | Pairwise evaluation: `Evaluate`, outcomes, `PropertyMeta`, `PropertyComparison` (graph alignment's per-pair unit) |
| `pairs.go` | Pair selection: `pairValues`, support weights, negatives, `ValueRef`, `Value.Label` and `Value.Auto` |
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
  - `ForSubject` reads an Interpretation Subject's Observations, negative ones marked; only a positive value earns credit.
  - `ForEntity` reads a handle's auto-reconciled values from the cache, at every rank.
- **Candidates:** the unmerged handles of the probe's Subject type, with their cached auto-reconciled values at every rank. The probe's own handle is never a candidate.
- **Pure core:** `core/match` has no catalog access. It can be tested with literal values, and can run anywhere a probe and candidates can be built.

## Scoring

Scores are **additive points**, so the weights read directly: "a matching name is worth 10, a different sex at birth costs 2".

For each Feature in the Profile:

1. Its Comparer judges every pair of probe and candidate values for that Property. The Feature takes the best similarity, from 0 to 1.
2. If no pair is comparable, the Feature is skipped: it neither helps nor hurts. A pair is not comparable when either side has no value, or the value is neutral (sex `unknown`, or a date with only a phrase).
3. If the similarity is above 0, the Feature adds `Weight × similarity × (1 − discount)`. The discount is 0 for every comparer except names, which discount a word common in the project's own people (below).
4. If the similarity is exactly 0 (a clear disagreement), the Feature subtracts `Contradiction`.

A candidate is returned when its total reaches `MinScore`, after the discount: a name that only a very common surname supports can fall under the bar. Results are sorted by score, then by ref. Every match carries its **reasons**: each Feature's similarity and contribution, for explaining it in the UI ("same name, born the same year") and for tuning.

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
  - Each name takes its pattern from `Value.NamePattern`. Loaders copy the project default's key onto it. An empty key uses the comparer's pattern.
  - Patterns come through a `NamePatterns` source. A catalog with a default profile supplies that profile (`nameformats.Load`): its role rows and weight rows, and that key. A catalog with no default, and `ComparerFor` for a caller with no catalog, stay on `BuiltinNamePatterns` (`western`).
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
- **A pair earns** word similarity × role affinity × the two words' weights. The score is that earned total over the weight of the words that bear on it. A word left over when the other name is fully paired exists only on one side: it adds nothing and takes nothing away. Words left over on both sides did not meet, and that disagreement costs their weight.
- **Role affinity:**
  - the same role: 1;
  - a nickname against a given name: 0.9;
  - a typed word against an untyped one: 0.8;
  - any other mismatch, such as a surname against a given name: 0.5.

  **Names entered in different formats still connect,** just lower. Types never block a pairing.
- **Surnames that share nothing:** when both names have family words and none resembles any word of the other name, the score is scaled by `GivenOnlyFactor` (0.5). A surname can change at marriage, but a shared "James" alone is weak. Swapped types are not a conflict.
- **Generations:** when both names carry suffixes and share none (Jr. vs Sr.), the score is scaled by `SuffixConflict` (0.3).
- **Words** match when equal (1), as an abbreviation or nickname (0.98), as a fold-only pair (0.95), as an initial and the word it begins (0.5), or as a near spelling at or above `FuzzyFloor` (0.8, edit-distance ratio). Near spelling compares the folded words, so an accent plus a typo does not pay for the accent twice.
  - Abbreviation and nickname tables are catalog rows. Install writes the seed from `productAliasTables` in `core/match/registry.go`. Scoring reads the catalog and no other list. Opening a project loads those rows once onto the catalog graph and keeps them until the project closes. `namealiases.Invalidate` reloads them after a table changes. An empty catalog makes no abbreviation or nickname hits. Every installed table is consulted. Origin (`provenencia`, `user`, `plugin:<id>`) records who contributed the table and does not turn it off. A pair hits when both tokens share one set and both words' roles are allowed. Product tables allow given, nick, and untyped, so two surnames do not. Free text does not use these tables. `Jas.` vs `James` and `Jim` vs `James` are 0.98. `Jas.` vs `Jim` is not one set, so it falls through to fold and edit distance. Adding a set extends comparison; the scorer does not name the words. The nickname sets are one canonical name and its direct nicknames, from [carltonnorthern/nicknames](https://github.com/carltonnorthern/nicknames) (Apache 2.0; license and credit in `core/match/third_party/nicknames/`). A shared nickname does not merge two canonical names: Chris matches Christine, and Chris matches Christopher, and Tina does not match Christopher.
  - Folding is NFD, drop combining marks, then the replacements in `LetterFolds` (`core/match/registry.go`). Adding a letter extends comparison; the fold does not name the letters. An empty map still drops combining marks, so `José` vs `Jose` is 0.95. `José` vs `José` stays 1. Display, `NormalizeForm`, and `sort_key` do not fold, so the Person page still keeps both spellings.
  - A swap of two adjacent letters counts as one edit (Robnis ~ Robins).
  - Words of three or more letters may also differ by one added or dropped letter (Ann ~ Anne, Jon ~ John: 0.75). A substituted letter in a short word is a different name (Mary ≠ Mark).
- **Normalization:** dashes and slashes separate words (Smith-Jones ~ Smith Jones). Apostrophes join (O'Brien ~ OBrien). A word in `SpacedParticles` (`core/match/registry.go`) rejoins with the next word (`O Brien` ~ `O'Brien`). A lone initial with a period (`O.`) does not rejoin. Adding a word to that list extends comparison; the splitter does not name the words.
- **Surname conflict:** an initial doesn't count as resembling a surname, so "S." doesn't hide a Smith/Robins conflict.
  - An initial is a lone cased letter ("J").
  - A lone character in an uncased script (蒋, 王) is a whole word.
- **No parts:** nothing to compare. `form` is not a match input.

**Name frequency.** A shared surname that is rare among the researcher's filed people counts for more than a shared common one. The comparer scales **points**, never similarity: an identical name is still 1, `agree`, and pinned in Promote.
- **Counts.** A word's count is how many unmerged Person handles carry it in a reconciled name value, once per handle however many values spell it. The key is the comparison token after the comparer's particles and folds (`José` and `Jose` are one key). Initials are not counted, every part word is, whatever its role. Promote reads the counts once into `graphalign.Stats.Words` and drops them with the rest of the stats on a handle commit. Merge hints and Subject suggestions (`core/database/matching`) count the candidates they loaded. A bare `NameComparer{}` or `DefaultNames` has no counts, so it scores as before.
- **Discount.** A word costs `FrequencyDiscount × ln(df) / ln(pop)` (0.2 at most). `df` is its count, capped at `pop`; `pop` is the larger of the people counted and `FrequencyMinPeople` (100), so a small project's first shared surname stays cheap. A word on one handle, an initial, and an empty project cost nothing. A pair of names pays the role-weighted mean of the discounts of the words the best pairing matched, on both sides.
- **Where it applies.** `Score` multiplies a positive contribution by `1 − discount`. A contradiction is not scaled. Promote's fixed name scale multiplies its points the same way and keeps the outcome and the pin.

| Pair (100 people; `smith` on 40, `quirk` on 1, `john` on 10) | Similarity | Discount | Name points (weight 10) |
| --- | --- | --- | --- |
| John Quirk vs John Quirk | 1 | 0.040 | 9.60 |
| John Smith vs John Smith | 1 | 0.136 | 8.64 |

Every weight and affinity is a `NameComparer` field.

**Support weights and negatives.** A handle's cached values carry the support the auto-reconciler counted (`support`, `against`, `reason`), and a Subject's Observations carry a polarity. `pairValues` (`pairs.go`) decides which value on each side speaks for a Feature, and both `Score` and `Evaluate` read their pair through it.
- **Weight.** A value's effective support is `support − against` (at least 1 for a displayed `kept` value, at least 0 for any other). The **leaders** are the displayed values with the highest effective support, or when none is displayed, the values that share the highest. A value's weight is its effective support over the leaders', at most 1. A neutral, `denied`, or `provisional` value, or one with no support left, never votes. Two equally supported names both lead, so a maiden and a married name each match in full.
- **Contradictions.** A negative Observation that is `autoreconcile.Compatible` with a leader on the other side contradicts it. A value compatible with a `denied` value that nothing else on that side backs contradicts it. When both sides have leaders and every comparable leader pair scores 0, that is a contradiction, so a minority that agrees does not rescue a mismatch. A negative never raises a score; a probe with only negatives has nothing to compare and the Feature is skipped.
- **Multiple cardinality.** A Feature whose Property is `multiple` (a Place's toponyms) gives every voting value full weight and skips the leader rule, so a handle's minority toponym still matches. A negative or a `denied` value still contradicts.
- **Credit.** Otherwise the best `similarity × weight × weight × (1 − discount)` among voting pairs, reported as `similarity × weights`. A handle or probe with no cache evidence weighs every value 1 and scores as before.
- **Reasons.** `Reason` and `PropertyComparison` carry `Probe` and `Candidate` (`ValueRef`: label, negative flag, the date of a date value, and the Observation id of a probe value). A term's label is its catalog label (`Value.TermLabel`), and the tie-break between equal values uses the term's key so a rename moves no reason. `attachExhibits` builds every Promote row's sheet lines: a row with no exhibit gets one line per comparison that spoke, and a row with exhibits gets a line for each row-level conflict no exhibit shows. These lines cannot be pinned, and a negative probe reads "not <value>".

| Probe | Handle values (support) | Result |
| --- | --- | --- |
| female | male (3), female (1) | Contradiction: −2; the reason names `female` and `male` |
| male | male (3), female (1) | +1 |
| female | male (2), female (2) | +1: both lead |
| James Robins | Mary Smith (5), James Robins (1) | Name similarity 0, a conflict in Promote; not a full match |
| not male (negative) | male (3), female (1) | Contradiction: −2 |

Promote's exhibit chooser keeps its own member picker, so a row can read conflict while an exhibit line shows an agreeing pair for a minority value.

Worked examples:

| Pair | Score |
| --- | --- |
| James Robins vs James Robins (forms "ROBINS, Jas." and "James Robins") | 1 |
| J. Robins vs James Robins | 0.8 |
| James Robins vs James Kenneth Robins (the middle name is only on one side) | 1 |
| Robins (surname only) vs James Robins | 1 |
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
- **Daitch–Mokotoff:** [`ideas/daitch-mokotoff.md`](ideas/daitch-mokotoff.md). Descoped from Spike 10 (S10-26). Spanish dual surnames, a patronymic part, other patterns, transliteration, Beider–Morse, and calendars: [`ideas/international-names.md`](ideas/international-names.md).
- **Abbreviations, date windows, event and place affinity:** Spike 10, [`deployment-plan/spike-10/name-matching-enhancements.md`](deployment-plan/spike-10/name-matching-enhancements.md). Known-as, concluded values, calibration, and rejected-claim memory: [`ideas/name-matching-enhancements.md`](ideas/name-matching-enhancements.md).

**Resolution uses the same parts.** The reconciler (`core/autoreconcile/names.go`) and `NameComparer` both ignore `form`. A name with no parts is no evidence there and not comparable here. The People list also ignores `form`: it orders by the profile's sorted part order. The cache name `sort_key` is still the normalized form. That column is not a match and not the list order.

`match.ComparerFor(valueType)` gives the default comparer for any Property, so a profile can also weigh researcher-defined Properties.

## Default profiles

These live in [`core/match/registry.go`](../core/match/registry.go), with every other tunable number.

| Kind | Features (Weight / Contradiction) | MinScore |
| --- | --- | --- |
| person | name 10 / 0 · sex_at_birth 1 / 2 · gender 1 / 2 (neutral terms not compared) | 3 |
| event | event_type 4 / 6 · date 6 / 4 · start_date 3 / 2 · end_date 3 / 2 | 5 |
| place | toponym 10 / 0 | 3 |

Worked examples:
- **Person:** "James Robins" against "Mary Robins" scores 10 × 0.6 = 6, which is shown but below the same name. The same name with a different sex at birth scores 10 − 2 = 8: the mismatch narrows the score and does not hide a name that cleared MinScore.
- **Event:** the same event type alone (every birth) scores 4, which is hidden. The same type plus the same year scores 4 + 4.8.

## Configuring

- **Per call:** `matching.Options{Profile: &p}` replaces the default profile. `Profile.With(feature)` copies a default and replaces or adds one Feature.
- **One registry:** [`core/match/registry.go`](../core/match/registry.go) holds everything:
  - word rules (`DefaultWordRules`)
  - name affinities and conflict factors (`DefaultNames`)
  - name patterns (`WesternNamePattern`, `BuiltinNamePatterns`)
  - date scores and span scoring (`DefaultDates`)
  - text and integer settings
  - default profiles

  Comparers fall back to these values, so a bare `NameComparer{}` scores exactly like `DefaultNames`, and no number is written twice.
- **Neutral terms** are not in the registry: a term flagged `property_terms.neutral` (seeded on *unknown* and *indeterminate* sex at birth, [`seeded-vocabulary.md`](seeded-vocabulary.md) §3.7) reaches the comparer as `Value.Neutral`, and `TermComparer` treats it as not comparable.
- **Guard tests** (`registry_test.go`):
  - Every comparer setting must have a registry value.
  - Bare comparers must score like the registry's.
  - Default profiles must use the registry's comparers.
  - The Western pattern must give every product part type a role.
- **Comparer settings** are pointers set with `match.Set`. Nil takes the registry value, and zero is a real value: `NameComparer{CrossRole: match.Set(0.0)}` makes types strict, and `DateComparer{Tolerance: match.Set(0)}` requires the same year.
- **Not yet:** researcher-facing tuning, profiles persisted per project, and data-driven name patterns.
  - **Stored profiles** replace `DefaultProfile` in `matching.start`, and in promotealign and graphalign, which read it directly.
  - **A stored-pattern source** replaces `BuiltinNamePatterns` as `NameComparer.Patterns`.

  Both use the same shapes, so scoring code doesn't change.

## Extending

- **A new kind** needs a `DefaultProfile` case. Its consumer then needs a header for the kind (Promote rows and alternatives carry the Person, Event or Place header).
- **A new signal on the thing itself** (a researcher-added Property, say) needs:
  - a Comparer, if no existing one fits;
  - a Feature in the profile;
  - the adapter putting those values into the probe and the candidates.
  Signals reached through edges (a Person's birth year through its birth event, its family) are **not** matching Features: Promote graph alignment compares neighbors directly, over the canonical graph (S9-28), with its own weights (S9-41).
- **Blocking:** candidates are currently every handle of the type with a cached profile value, read in one query (constant query count, but linear rows). When catalogs outgrow that, narrow candidates in `loadCandidates` (and in promotealign's `loadCanon`, which reads the same set from the graph with `g.Kind`) using the handle search index (S9-34a: `catalog_search_docs` kinds `person` / `event` / `place`, kept current by the cache's upkeep). `rank` takes its name-word counts from the set `loadCandidates` returns, so blocking must count a separate, unnarrowed population of the type's handles. Scoring does not change only if blocking does that.
- **Merge hints:** `matching.ForEntity(handle)` ranks the other handles of the same type. A merge-hint surface needs only to shape that result.
