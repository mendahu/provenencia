# Comparison Layers

**Status:** requirements, being scoped (2026-10-16). Not a PR map yet. This is the first decision of [Spike 11](README.md): the bottom two layers of matching. Background: [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) ("First: make the algorithm graspable"). Today's code: [`matching.md`](../../matching.md), [`promote-graph-alignment.md`](../../promote-graph-alignment.md).

## Why

Promote's matching is hard to reason about because its smallest pieces mix concerns. A single comparer call today can:

- compare two values;
- decide whether a value counts as evidence at all (neutral terms);
- discount the result by how common the name is in the catalog.

The code that picks which of several values speaks for a property (`pairValues`) is real, tested and subtle, but has no name in the docs. Reconciliation has its own idea of "the same value" (`autoreconcile.Compatible`), with different rules for text, terms and names. So Promote can call a comparison "agree" and still refuse to pin it.

We want a stack of layers. Each layer is pure, has one job, and gives the layer above it everything it needs to explain itself.

## The layers

Packages: layer 1 is **value compare** (`core/valuecompare`), and layer 2 is **property compare** (`core/propertycompare`). Go package names can't have spaces, so the folders are written as one word.

| # | Layer | Question it answers | Status |
|---|---|---|---|
| 1 | **Value comparison** | How do these two values relate? | This doc |
| 2 | **Property comparison** | Do these two Subjects agree on one property? | This doc |
| 3 | **Subject comparison** | How alike are these two Subjects on their own values? | Later. Combines property comparisons and holds comparisons that span properties |
| 4 | **Entity comparison** | Does this Subject belong with this entity? | Decided in shape (§ Entities): subject comparison against the entity's reconciled conclusion |
| 5 | Candidate scoring and graph alignment | Which handle is this Subject? | Later (`Rank`, link points, `Align`) |

Layers 1–3 are **discrete**: one thing against one thing. An entity is a group of Subjects, so comparing against one is one-to-many. Layer 4 turns that back into one-to-one by comparing against the group's conclusion.

## Rules for every layer

1. **Pure.** No I/O, no globals read at compare time, no clock. The same inputs give the same output.
2. **Inputs only from below or as arguments.** A layer may call the layers under it, and may take configuration and data as arguments. It never reaches up or sideways.
3. **If it needs more, it moves up.** When a rule needs something a layer doesn't take (a property's cardinality, a catalog count, a neighbor), the rule belongs to a higher layer. We don't widen the lower layer's inputs.
4. **Output is the explanation.** A layer's result carries enough structure that the layer above can explain itself without redoing the work, and the UI can word it without parsing strings.
5. **One registry per layer.** Every tunable a layer reads lives in that layer's registry. The registry is passed in, so tests can override it.
6. **Tested at its own level.** Each layer has exact-value tables and property-based invariants that need nothing above it. A layer's package builds and tests with no build tags and no database.
7. **Single-purpose and composable.** A comparison layer answers one question and makes no workflow decision. Each workflow (Promote, reconciliation, merge hints, search) keeps its own decision logic and composes the layers it needs. For example, the reconciler can run value compare across five Subjects' values for one property, then apply its own rules to decide what the entity shows. Reconciling uses comparison, but it is not comparison.

## Layer 1: Value comparison

### Contract

```go
// Compare relates two values of one value type under the given rules.
func Compare(a, b Value, rules Rules) (Result, Detail) // valuecompare.Compare

type Result struct {
    Relation   Relation // what kind of match: same, contains, near, different, …
    Similarity float64  // how close, 0…1
}
```

- **One entry point per value type** behind one interface. Name, date, text, term and integer each have a comparer. A value type with no comparer (subject, depiction) says so explicitly, and is never silently compared as 0.
- **`Result` is the normalized part.** Every value type returns the same `Relation` set and a `Similarity` with a shared meaning (§ Normalizing). A caller can use `Result` without knowing the value type.
- **`Detail` is the reasoning.** It is structured, with a shared shape across types and type-specific parts inside it (§ Detail).

### What layer 1 may take, and what moves up

| Input | Layer 1? | Why |
|---|---|---|
| The two values | Yes | |
| Language rules: alias and nickname tables, letter folds, particles, name patterns, date tolerance | Yes, as `Rules` | They change *how alike* two values are, and are the same for anyone holding the same rules |
| A term's neutral flag (sex `unknown`) | **No → layer 2** | "Is this value evidence?" is a property-level decision |
| Negative polarity, denied, provisional, support counts | **No → layer 2** | Context about the value on its side, not about the value |
| Cardinality (single or multiple) | **No → layer 2** | A property fact |
| Word or value frequency in the catalog | **No** (descoped, § Frequency) | Changes *how informative* a match is, not how alike |
| Another property (an event's `start_date` when comparing `date`) | **No → layer 3** | Spans properties, so it is a Subject-level comparison |

**A test for the boundary:** would two researchers with the same rules but different catalogs expect a different answer? If yes, the input belongs higher up.

**Layer 1 never decides whether a value is evidence.** It only decides whether two values can be compared. `Unknown` means "nothing to compare": a name with no parts, a date that is only a phrase, an empty value.

### Normalizing similarity: options

Today every comparer returns 0–1, but the numbers mean different things by type. 0.2 is "a different person" for names and "the best an open span can do" for dates. Promote compensates with a floor per type (name 0.75, term 1, date 0). These are the options for giving the number one meaning.

**A. Anchored bands.** Each relation owns a fixed similarity band (same = 1, equivalent ≥ 0.9, contains 0.7–0.9, …).
- Pro: a floor means the same thing for every type, and the outcome can be read from the number alone.
- Con: it forces one strength order on relations that don't really have one. Is a typo stronger than a containment? It also puts hard edges inside what should be continuous.

**B. Relation plus free similarity.** `Relation` says what kind of match it is, and `Similarity` ranks within it. Callers that need a sameness answer read the relation. Callers that rank read the similarity.
- Pro: no forced order between kinds. The relation is fully type-independent.
- Con: on its own, similarity still drifts by type at the margins.

**C. One fall-off curve.** Where a type has a natural distance (days between dates, a gap between integers, edits between words), similarity is one shared curve over `distance ÷ tolerance`. Each type sets only its tolerance unit.
- Pro: the shape of "how fast closeness drops off" is the same everywhere, with one knob per type.
- Con: it doesn't cover categorical values (terms) or values made of parts (names, text). Those need a rule for combining parts (below).

**D. Calibrated probability.** Similarity is defined as "how likely these two values, alone, record the same fact", fitted from labels.
- Pro: genuinely comparable across types, and it feeds layer 3 directly.
- Con: it needs the labelled data from the matching test suite, which doesn't exist yet.

**Recommendation: B, with C's curve and A's two fixed ends now, and D as the target once labels exist.**
- `Relation` is the cross-type contract.
- Similarity has fixed meanings only at the ends: **1** is `Same`, **0** is `Different`, and `Equivalent` is ≥ 0.9.
- Between the ends, types with a distance use the shared curve.
- Types made of parts combine their parts with one shared rule.
- When the labeler exists, we calibrate similarity against it without changing the contract.

### Relations

| Relation | Means | Symmetric? |
|---|---|---|
| `Same` | Identical after the normalization every comparer does (case, whitespace, punctuation rules) | yes |
| `Equivalent` | The same value written another way: accent fold, alias or nickname, abbreviation | yes |
| `Contains` | One side is more specific and fits inside the other: `J.` vs `James`, `1817` vs `3 Mar 1817`, `Robins` vs `James Robins`, `BEF 1820` vs `1815` | **no**: `Detail` says which side is the broader one |
| `Overlaps` | Each side has something the other lacks, but they share some: overlapping date ranges, text sharing some words | yes |
| `Near` | Separate values that are close: a spelling variant, dates within tolerance, integers within tolerance. Requires at least the type's **minimum similarity** (`Rules`); below it the pair is `Different` | yes |
| `Different` | No match | yes |
| `Unknown` | Nothing to compare | yes |

Proposed relation per value type. Every cell is a proposal to confirm; the name row is chosen to keep today's agree / conflict calls:

| Type | Same | Equivalent | Contains | Overlaps | Near | Different |
|---|---|---|---|---|---|---|
| name | all shared parts same | fold, alias, abbreviation | parts on one side only (`Robins` / `James Robins`), initial / word | — | spelling variant | any shared role disagrees (`Mary Robins` / `James Robins`), generation conflict |
| date | same span | — | one span inside the other | spans overlap | apart, within tolerance | beyond tolerance |
| text | same words | folded words | all of one side's words in the other | some words shared | one-word spelling variant | nothing shared |
| term | same key | — (later: a term-affinity table, e.g. birth / baptism → `Near`) | — | — | — | different key |
| integer | equal | — | — | — | within tolerance | beyond tolerance |

**Decided (2026-10-16):**
- **Minimum similarities start from today's floors:** name 0.75, spelling variant 0.8 (`FuzzyFloor`). Types with no floor today get one when their comparer moves.
- **`Contains` scores below 1.** `Robins` vs `James Robins` is no longer a perfect match. How far below 1 is a tuning question.
- **`Near` has a minimum similarity per type.** It lives in layer 1's `Rules`, because "near or different" is layer 1's call. Layer 2 then needs no floors.
- **One `Contains`, two causes.** Less precise (`1817` vs `3 Mar 1817`, `J.` vs `James`) and less complete (`Robins` vs `James Robins`) are both `Contains`. `Detail.Cause` tells them apart (`precision`, `missing_part`), so callers that care can treat them differently.

**Parts combine by one rule.** Names and text are made of parts. Each part pair gets its own relation and similarity. The value's relation is the weakest relation among the parts both sides carry, and parts on one side only make it `Contains`. Its similarity is the weighted mean of the part similarities. The weights come from `Rules`, e.g. name roles from the name pattern.

**Why the relation matters above layer 1:**
- **Promote floors disappear.** Layer 2 reads agree, partial or conflict straight from the relation (§ Layer 2), not from a per-type cutoff on the number.
- **Reconciliation can build on it.** The reconciler keeps its own grouping logic (rule 7), but bases it on value compare's relations instead of a second comparer. Because `Same` and `Equivalent` are separate relations, the reconciler can decide for itself whether `José` and `Jose` show as one value.
- **The UI gets words for free.** "Same name", "same name, different spelling", "less specific date" map to relations, not numbers.

### Detail

A shared shape, so layers 2 and 3, the UI and the labeler can read any type the same way:

```go
type Detail struct {
    Parts      []PartMatch // what paired with what
    OnlyA      []Part      // parts on one side only
    OnlyB      []Part
    Broader    Side        // for Contains: which side is less specific
    Distance   float64     // for curve types: the raw gap …
    Unit       string      // … and its unit ("days", "edits", "count")
    Cause      Cause       // the rule that decided the relation
}

type PartMatch struct {
    A, B       Part        // the two parts, with display text and a kind (family, given, from, to, word)
    Relation   Relation
    Similarity float64
    Cause      Cause
}
```

- **`Cause` is a shared vocabulary of codes:** `exact`, `case`, `fold`, `alias`, `abbreviation`, `initial`, `spelling`, `precision`, `missing_part`, `overlap`, `gap`, `beyond_tolerance`, `generation_conflict`, `empty`, `phrase_only`. UI copy, reason codes and labeler features are all keyed on these.
- **`Parts` explains the match.** Which words or ends paired is what the UI shows and what the labeler records as features.
- **No strings to parse.** The display text is for showing; logic reads the codes.

### Purity requirements

- **Its own package,** importing only the standard library and other pure packages. Today `core/match` imports `core/database/properties`, `datevalues` and `namevalues`, which pull in `core/database`, `graphcache` and a migration at start-up. That's why `go test ./core/match` fails without `-tags fts5`.
- **Its own value types.** The date and name values it compares must come from a package with no database dependency. Either move the structs out of `core/database/datevalues` and `namevalues`, or define layer-1 types that loaders fill.
- **An import guard test** fails if the package's dependencies grow to include anything under `core/database`.
- **The word primitives move in.** `SpellingSimilarity`, `NormalizeForm` and `IsInitial` live in `core/autoreconcile` today. They are value comparison, and both matching and reconciliation should import them from layer 1.

### Tests

- **Contract tests, run against every comparer:**
  - similarity is in 0…1;
  - `Same` ⇔ 1;
  - `Different` ⇔ 0;
  - `Equivalent` ≥ 0.9;
  - `Unknown` has no similarity;
  - swapping sides gives the same similarity and relation, with `Contains`' broader side flipped;
  - a value compared with itself is `Same`;
  - the result doesn't depend on the order of parts or rules.
- **Exact tables per rule,** with expected values written as the arithmetic (as `names_test.go` does today).
- **Ladders:** which pair must outrank which. They survive retuning.
- **Generated permutations** from fixed seeds, for every type (today only names have them).

## Layer 2: Property comparison

### Contract

```go
// Compare judges whether two Subjects agree on one property,
// from every value each one holds for it.
func Compare(a, b Side, meta PropertyMeta, rules Rules) PropertyResult // propertycompare.Compare
```

- **`Side`** is one Subject's values for the property. Each value carries only what a Subject can say about itself:
  - polarity (positive or negative);
  - for a term, its neutral flag.
- **No reconciler context.** Support counts and standing (kept, outvoted, denied…) describe an entity's group, not a Subject. They never reach layer 2. Layer 4 translates them before comparing (§ Entities).
- **`PropertyMeta`:** value type and cardinality.
- **No points, no catalog statistics.** Layer 2 says *whether* the sides agree and *which values spoke*, not how much it's worth.

### What it does, in order

1. **Drop non-evidence.** Remove neutral values, and record each one with its reason.
2. **Compare pairs.** Run layer 1 on every pair of positive values.
3. **Choose the pair that speaks:** the best pair.
4. **Apply negatives.** A negative value whose relation to a positive value on the other side is `Same`, `Equivalent` or `Contains` contradicts it.
5. **Apply cardinality.** On a multiple-value property, a difference is not a conflict.
6. **Decide the outcome** from the relation (below).

### Outcome from relation

| Relation of the pair that speaks | Single-value property | Multiple-value property |
|---|---|---|
| `Same`, `Equivalent` | agree | agree |
| `Contains`, `Overlaps`, `Near` | partial | partial |
| `Different` | conflict | unknown |
| `Unknown`, or no voting values | unknown | unknown |
| A negative matches a positive value on the other side | conflict, cause `negative` | conflict, cause `negative` |

This replaces Promote's per-type floors (`core/graphalign/registry.go`). A spelling variant too far apart never reaches layer 2 as `Near`: layer 1's minimum similarity has already made it `Different`.

### Output

```go
type PropertyResult struct {
    Outcome  Outcome          // agree, partial, conflict, unknown
    Cause    OutcomeCause     // relation, negative, no_values, …
    Pair     *SpokenPair      // the two values that spoke, their layer-1 Result and Detail
    Excluded []ExcludedValue  // values dropped in step 1, with the reason
}
```

This is the feature vector the matching test suite's fit needs: one outcome, relation and similarity per property. Layer 3 turns it into points.

### Tests

- **Invariants:**
  - adding a non-evidence value changes nothing;
  - a negative never raises the outcome;
  - value order on either side doesn't matter;
  - swapping sides gives the same outcome;
  - on a multiple-value property, adding a different value never makes it conflict.
- **Exact tables** for the Subject-level cases in today's `pairs_test.go` (negatives, cardinality, neutral terms), ported to the new types. The support and standing cases move to layer 4's tests.

## Layer 3: Subject comparison

**Requirements (2026-10-16):**
- **Compare the properties both Subjects have.** If both have a name, compare names. A property on one side only is unknown and neither helps nor hurts.
- **Registry-driven.** One registry defines every Subject type, its properties, and how each one counts in a Subject comparison. Layer 3 has the most levers of any comparison layer, and all of them live there.
- **Comparisons that span properties are special cases, declared in the registry.** The known one: an event's `date` against another's `start_date` + `end_date`, read as one period.
- **Useful whatever the size.** A Subject with three or four properties against an entity with four or five does happen, and the comparison should get full value from it. In practice most Subjects carry one or two properties, so its *effective* job will be mostly agreement plus ruling out the obvious (names nothing alike, different event types). That is an outcome of the weights, not a separate job. Relationships are not compared here; graph alignment does that, and it does most of the positive work.
- **Agreement and amount of evidence are different.** Two people whose names match exactly, and two whose names *and* genders match, both agree perfectly, but the second has more points of connection and is the better match. One 0–1 number can't say both.

**Options for the result:**

| Option | Name agrees | Name + gender agree | Catch |
|---|---|---|---|
| A. Agreement only: weighted mean over the properties compared | 1.0 | 1.0 | Can't tell the two apart |
| B. Evidence points: sum of agree weights, minus conflict penalties | 2.5 | 3.1 | Not 0–1; meaning depends on the weights |
| C. B squashed to 0–1 with a logistic curve | ~0.92 | ~0.96 | Means "confidence", not "similarity"; needs calibration |
| D. Agreement × coverage of the type's properties | lower | higher | Punishes a sparse Subject as if missing were disagreeing |

**Decided (2026-10-16): two axes, A and B.**
- **Similarity (0–1):** are these alike? The weighted mean over the properties compared, with a conflict counting as 0. It is how candidates are chosen to show.
- **Evidence (points):** for things that are alike, how strong is the connection? Agree weights minus conflict penalties. It is the main currency, because graph alignment adds relationship points on top, so both must be points.
- **No verdict and no knockout flag.** A conflict that should rule a pair out is a large conflict penalty in the registry, large enough that nothing recovers from it. It stays one number, on the same scale as everything else.
- **A 0–1 confidence (C), if wanted, is applied at the top of the stack,** to the total of properties + links. The labeler's ordered logistic regression is what fits it ([`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md)).

| Case | Similarity | Evidence |
|---|---|---|
| Name agrees | 1.0 | 2.5 |
| Name + gender agree | 1.0 | 3.1 |
| Name agrees, sex conflicts | 0.81 | 1.5 |
| Event types agree, nothing else | 1.0 | 0.5 |
| Birth vs death, nothing else | 0 | −20 (effectively never recovers) |

**Registry shape,** per Subject type and property: agree weight and conflict penalty. They are separate on purpose: event type has a low agree weight and a very large conflict penalty. Comparisons that span properties are entries too, with their own weights. Numbers above are today's where they exist and placeholders otherwise.

## Entities: compare against the conclusion

**Decided (2026-10-16):** a Subject is compared with an entity's **reconciled values**, not with each member.

**Why.** An Identity Claim says "this Subject belongs in this group". The group's values are its conclusion, reconciled by the auto-reconciler or overridden by the researcher. Both kinds of decision matter. A conclusion the researcher hasn't overridden is taken as the conclusion. Whether a Subject joins the group is judged against what the group concludes.

**What that buys.**
- **Weighting by support is outsourced.** Which value an entity holds, and how strongly its members back it, is the reconciler's job, already done. Matching doesn't re-weigh members.
- **One comparison per candidate,** not one per member.
- **Layer 4 is thin.** It builds a Subject-shaped **profile** from the conclusion, then runs layer 3 against it. No new comparison logic.

**The profile.** The reconciler's standing is translated into concepts layer 2 already has, so layer 2 stays free of standing:

| Reconciled value | In the profile |
|---|---|
| The concluded value (several when tied leaders show together, or when the property is multi-valued) | positive |
| A value the researcher denied | negative |
| Outvoted, weak, provisional | left out |

**The profile is strict (decided 2026-10-16).** Only the conclusion goes in. An outvoted value is, by the group's own judgment, not true. A real alternate, such as a maiden name, belongs in the conclusion: the auto-reconciler flags it, or the researcher records a Reconciliation Claim. Maiden names are not built yet, but that is the intended path. If a true value is missing from the conclusion, that is a reconciliation fix, not a matching rule. A single name mismatch is also weighed alongside sex, dates, relationships and the rest of the graph above this layer, so it doesn't decide a match alone.

**The profile builder is its own package:** `core/subjectprofile`, with `subjectprofile.Build`. It turns reconciler output into a Subject-shaped profile and compares nothing.

## Frequency

**Decided:** frequency is not a comparison. Layers 1 and 2 never read catalog statistics.

**Decided (2026-10-16): name frequency is descoped.** "Common names count for less" is a subjective weighting. The one place it could matter is Subject-to-entity matching, and there the other properties and the graph (dates, sex, relationships, neighbors) already tell look-alikes apart. It adds noise and a tunable without earning its place.

- Remove `NameComparer.Frequency`, the discount, `DiscountComparer`, `CompareDiscount` and `Stats.Words` (shipped in S10-27).

**Decided (2026-10-16): all value-frequency weighting is descoped.** Promote also weights event types and toponyms from how common the value is in the catalog (`ln(m/u)` from `Stats.ValueFreq`, lifted to 2). Catalog frequency is the wrong reason for a weight:
- **Event type** carries little weight because it is generic, not because the catalog holds many births. "Both are births" says nothing specific about which birth. That is a fixed registry weight for the property.
- **Toponyms** match often. Their real weight comes from the place's relationships (`part_of`, the events located there), which graph alignment already scores.

So every property weight is a plain registry number. The labeler can fit them later. Remove `Stats.ValueFreq`, the `m` / `u` priors (`MPrior`, `UPrior`), the frequency scales and their lift to the accept bar. The toponym agreement bonus existed to work around the frequency weight; re-set it as a plain weight if still needed.

## What moves where in today's code

| Today | Goes to |
|---|---|
| `match.Comparer`, `ComparerFor`, the five comparers | Layer 1, returning `Result` + `Detail` |
| `NameComparer.Frequency`, `DiscountComparer`, `CompareDiscount`, `Stats.Words` | Removed (name frequency descoped) |
| `TermComparer`'s neutral check | Layer 2, step 1 |
| `autoreconcile.SpellingSimilarity`, `NormalizeForm`, `IsInitial` | Layer 1 |
| `pairValues`' negatives and multiple cardinality (`pairs.go`) | Layer 2 |
| `pairValues`' leaders and support weights | Gone: the entity profile (layer 4) contains only the conclusion |
| `poolPeriods` (`periods.go`) | Layer 3 (spans properties) |
| Promote's scale floors and outcome logic (`ApplyScale`) | Layer 2's outcome table |
| Promote's scale weights and contradiction (`resolvedScale`) | Layer 3, as plain registry weights. Its frequency weights are removed |
| `autoreconcile.Compatible` and grouping | Stays in the reconciler as its own decision logic, built on value compare (and property compare where it helps) instead of its own key rules |
| `match.Value`'s reconciler fields (`Support`, `Against`, `Reason`, `Denied`) | The profile builder package. `Negative` stays as layer 2 polarity |

## Open questions

- **The reconciler's grouping rules.** Which relations make two values "show once" is the reconciler's call (rule 7), designed with its own PR. Today's disagreements to resolve there: `Montréal` / `Montreal` and `Jim` / `James` agree in matching but don't group or pin; `J. Robins` / `James Robins` groups but only scores partial.
