# Comparison Layers

**Status:** requirements, scoped 2026-10-16 – 17. Not a PR map yet. The matching design for [Spike 11](README.md): six layers from value comparison to graph alignment, bands, the match registry, and the reuse inventory. Background: [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md) ("First: make the algorithm graspable"). The lasting one-page model and knob inventory: [`matching-model.md`](../../matching-model.md). Today's code: [`matching.md`](../../matching.md), [`promote-graph-alignment.md`](../../promote-graph-alignment.md).

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
| 3 | **Subject comparison** | How alike are these two Subjects on their own values? | Decided in shape (§ Layer 3) |
| 4 | **Entity comparison** | Does this Subject belong with this entity? | Built in S11-11 (§ Entities): subject comparison against the entity's reconciled conclusion |
| 5 | **Segment comparison** | Does this neighbor, through this bridge, line up? | Decided in shape (§ Segments) |
| 6 | **Graph alignment** | Which handle is each Subject on the layer? | Decided in shape (§ Layer 6 procedure) |

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

Relation per value type (confirmed 2026-10-17; worked examples below the table):

| Type | Same | Equivalent | Contains | Overlaps | Near | Different |
|---|---|---|---|---|---|---|
| name | all shared parts same | fold, alias, abbreviation | parts on one side only (`Robins` / `James Robins`), initial / word | — | spelling variant; different middle or untyped words | a surname or first given name on each side that the other lacks (`Mary Robins` / `James Robins`), generation conflict, under the Near minimum |
| date | same span | — | one span inside the other | spans overlap | apart, within tolerance | beyond tolerance |
| text | same letters, word breaks ignored | folded letters | all of one side's words in the other | some words shared | one-word spelling variant | nothing shared |
| term | same key | — | — | — | — | different key (birth / baptism, son / child) |
| integer | equal | — | — | — | within tolerance | beyond tolerance |

**Decided (2026-10-16):**
- **Minimum similarities start from today's floors:** name 0.75, spelling variant 0.8 (`FuzzyFloor`). Types with no floor today get one when their comparer moves.
- **`Contains` scores below 1.** `Robins` vs `James Robins` is no longer a perfect match. How far below 1 is a tuning question.
- **`Near` has a minimum similarity per type.** It lives in layer 1's `Rules`, because "near or different" is layer 1's call. Layer 2 then needs no floors.
- **One `Contains`, two causes.** Less precise (`1817` vs `3 Mar 1817`, `J.` vs `James`) and less complete (`Robins` vs `James Robins`) are both `Contains`. `Detail.Cause` tells them apart (`precision`, `missing_part`), so callers that care can treat them differently.

**Worked cases decided 2026-10-17:**

| Pair | Relation |
|---|---|
| surname "James", given "Robins" / James Robins (part types swapped) | `Different` (under the Near minimum; decided 2026-10-17, S11-03) |
| Mary Smith / Mary Robins (surname change at marriage) | `Different`: marriage evidence comes from links, not from the name |
| ABT 1817 / 1818 | Read `ABT` as a range (below): 1816–1818 contains 1818, so `Contains` |
| BEF 1820 / BEF 1830 | `Contains` |
| Upper Canada, York / York, Upper Canada | `Different`: text is compared as written, and a toponym should name one place |
| birth / baptism; son / child | `Different`. A term-affinity table is not planned |

**`ABT` is a range of one unit of its most precise part,** on each side: `ABT 1817` is 1816–1818; `ABT Mar 1817` is Feb–Apr 1817; `ABT 3 Mar 1817` is 2–4 Mar 1817. It then follows the range rules (`Same`, `Contains`, `Overlaps`, `Near`, `Different`). This replaces "ABT doubles the tolerance" in today's `DateComparer`.

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

- **`Cause` is a shared vocabulary of codes:** `exact`, `case`, `fold`, `alias`, `abbreviation`, `initial`, `spelling`, `precision`, `missing_part`, `overlap`, `gap`, `beyond_tolerance`, `generation_conflict`, `role` (words paired across part types), `below_minimum` (a pair close enough to be `Near` but under its type's minimum, so `Different`), `empty`, `phrase_only`, `no_match` (nothing in common), `too_long` (text with more words than the alignment may compare), `order` (text whose shared words are out of order), `unsupported` (no comparer, or two value types). UI copy, reason codes and labeler features are all keyed on these.
- **`Parts` explains the match.** Which words or ends paired is what the UI shows and what the labeler records as features.
- **No strings to parse.** The display text is for showing; logic reads the codes.
- **Echoes the callers' value refs.** Each value can carry an opaque ref (an Observation id) that the layers never read and `Detail` hands back, so the claim evidence view knows exactly which Observations spoke without the layers knowing about the catalog. Segment comparison does the same for a whole segment: bridge ends, role or term, and the neighbor's agreeing values.
- **Rich enough to word evidence lines.** The claim evidence view ([`claim-evidence.md`](claim-evidence.md)) builds each sentence ("same name, different spelling", "same date, recorded less precisely") from the relation and `Cause` alone.

### Purity requirements

- **Its own package,** importing only the standard library and other pure packages. Today `core/match` imports `core/database/properties`, `datevalues` and `namevalues`, which pull in `core/database`, `graphcache` and a migration at start-up. That's why `go test ./core/match` fails without `-tags fts5`.
- **Its own value types.** The date and name values it compares must come from a package with no database dependency. Either move the structs out of `core/database/datevalues` and `namevalues`, or define layer-1 types that loaders fill.
- **An import guard test** fails if the package's dependencies grow to include anything under `core/database`.
- **The word primitives move in.** `SpellingSimilarity`, `NormalizeForm` and `IsInitial` live in `core/autoreconcile` today (moved to `core/valuecompare` in S11-02). They are value comparison, and both matching and reconciliation should import them from layer 1.

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
func Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result // propertycompare.Compare
```

- **Each side** is one Subject's values for the property (`[]propertycompare.Value`). Each value carries only what a Subject can say about itself:
  - polarity (positive or negative);
  - for a term, its neutral flag.
- **No reconciler context.** Support counts and standing (kept, outvoted, denied…) describe an entity's group, not a Subject. They never reach layer 2. Layer 4 translates them before comparing (§ Entities).
- **`propertymeta.Meta`:** value type and cardinality. The property sets each value's type.
- **No points, no catalog statistics.** Layer 2 says *whether* the sides agree and *which values spoke*, not how much it's worth.

### What it does, in order

1. **Drop non-evidence.** Remove neutral values, and record each one with its reason.
2. **Apply negatives.** A negative value whose relation to a positive value on the other side is `Same`, `Equivalent` or `Contains` contradicts it: the outcome is conflict, whatever the positive values say.
3. **Compare pairs.** Run layer 1 on every pair of positive values.
4. **Choose the pair that speaks:** by outcome (agree, partial, conflict, unknown), then similarity, then the stronger relation, then each side's `Same` key. Step 2 picks among contradicting pairs the same way.
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
type Result struct {
    Outcome  Outcome        // agree, partial, conflict, unknown
    Cause    Cause          // relation, negative, no_values, cardinality
    Speaking *SpeakingPair  // the pair that spoke; nil when a side has no positive value
    Excluded []Excluded     // values dropped as non-evidence, side A's then side B's
}

type SpeakingPair struct {
    A, B   Value                // side A's and side B's value; one may be negative
    Result valuecompare.Result
    Detail valuecompare.Detail
}

type Excluded struct {
    Side   valuecompare.Side
    Value  Value
    Reason ExcludeReason        // neutral
}
```

Built in S11-09. Outcome causes: `relation`, `negative`, `no_values`, `cardinality` (a `Different` pair on a multiple-value property). Refs are never read.

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
- **Comparisons that span properties are special cases, declared in the registry.** Two known ones:
  - an event's `date` against another's `start_date` + `end_date`, read as one period;
  - a person's **sex at birth and gender, compared as one pair**: both sides' sex and gender values are pooled into one direct comparison, so a Subject recording only gender still meets an entity recording only sex at birth.
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
| Name agrees, sex conflicts | 0.71 | 1.5 |
| Event types agree, nothing else | 1.0 | 0.5 |
| Birth vs death, nothing else | 0 | −20 (effectively never recovers) |

**Decided (2026-10-16): a conflict weighs its penalty in similarity.** A property counts with its agree weight when it agrees and its conflict penalty when it conflicts. Otherwise a birth vs death with matching dates would read as very similar (event type's agree weight is only 0.5) while evidence reads −20. With penalties, that pair reads about 0.09, and the two axes agree.

**Registry shape,** per Subject type and property: agree weight and conflict penalty. They are separate on purpose: event type has a low agree weight and a very large conflict penalty. Comparisons that span properties are entries too, with their own weights. Numbers above are today's where they exist and placeholders otherwise.

**Built in S11-10.** `subjectcompare.Compare(kind, a, b, metas, rules, value)` runs layer 2 once per lead property in `subject.Weights[kind]`, in key order. Agree earns the agree weight, partial the agree weight × the speaking pair's similarity, conflict loses the conflict penalty, unknown counts for nothing. Similarity is the weighted mean over counted properties: agree 1 and partial its similarity, each at the agree weight; conflict 0 at the penalty. `subject.Pools` adds properties' values to the lead's on each side (sex with gender). `subject.Periods` builds, on a side with no positive `date`, one range from its earliest start and latest end, or `AFT` the start or `BEF` the end when it has only one (`valuecompare.EndsDate`). Value types and cardinalities come from the catalog as `Metas`. A derived property (lifespan) is a lead in `subject.Weights` like any other.

## Segments (layer 5) and graph alignment (layer 6)

**Decided (2026-10-16): two layers.** Comparing a segment is a pure, one-hop comparison. Walking the graph and deciding is its own logic layer.

**A segment** is two primary Subjects joined by a bridge: primary – bridge – primary. Segment comparison compares a segment on the layer with one in the research graph:
- **The neighbor** by entity comparison (layer 4).
- **The bridge** by Subject comparison on the bridge's own properties (role, relationship term, dates), plus direction: a parent maps to a parent, and a term and its inverse correspond with the ends swapped. This replaces exact signature equality (`corresponds`).
- **Output:** outcome (match, partial, conflict, unknown), similarity, evidence, detail. No walking, no decisions.

**A segment behaves like a property whose value is a neighbor,** with the same cardinality rule as property comparison:

| Bridge cardinality | Neighbor matches | Neighbor differs | No corresponding link |
|---|---|---|---|
| One: subject of a birth, subject of a death | strong support | **conflict: subtracts evidence** | unknown |
| Not declared: everything else (parents, spouses, children, residences, witnesses…) | support | unknown | unknown |

**Decided:**
- **Cardinality is declared in the registry,** replacing fan-out measured from the catalog (the same kind of statistic as the descoped frequency).
- **Only birth and death are declared, as one (2026-10-17).** A person has at most one birth and one death; that's axiomatic. More than two parents and more than one spouse are reasonable and common, so modelling limits there gets too complex. Every other bridge has no declared cardinality: a matching neighbor supports, and a differing one is unknown. The key is participation · subject · event type, so adding another axiomatic case later is one registry entry.
- **A conflict through a one-cardinality bridge subtracts evidence.**
- **Segment evidence includes how well the neighbor itself matched** (today's credit).

**Graph alignment (layer 6)** holds everything that searches or decides: seeding and candidates (with the property-only fallback), listing segments around candidate pairs, one-to-one, propagation (each accepted pair becomes an anchor), conflicts, and chains. It replaces the three walk-like passes today (walk, fallback re-walk, refinement) with one defined procedure. **Chains come from the walk:** an accepted pair is an anchor whose own segments then count, so a double or triple segment match is several well-supported Subjects, not a multiplied score.

### Layer 6 procedure

**Decided (2026-10-17).** A pair is compared only when something cheap put it forward: a shared shortlist key, or a link to an accepted pair. Nothing is compared just because it exists. A Source with 100 Subjects stays at a few thousand pure comparisons, not hundreds of thousands. Terms are in § Vocabulary.

1. **Shortlist.** For each Subject, look up research handles that share a shortlist key. Keys are declared per kind in the registry and built with value compare's `Key`, so shortlisting and comparison agree on "the same". Examples: person, name words (not initials); event, type + year bucket within tolerance; place, toponym. The index is one key → handles map per kind on the in-memory catalog graph (`graphcache`), built once and dropped with a handle commit, like the stats. Guarantee: shortlisting never misses a pair value compare would call `Same` or `Equivalent` (aliases share a key).
2. **Score and keep the top K.** Entity comparison (layer 4) on shortlisted handles; keep the best K per Subject (registry, about 5). Nominated handles skip shortlisting.
3. **One queue.** Every candidate enters one best-first queue at the start, scored on its properties. Anchors (decided rows, already-promoted Subjects) add segment evidence to their neighbors' entries. There is no separate fallback pass and no second walk: an isolated Subject is a candidate that never gets link support. **The answer does not depend on where the walk starts.**
4. **Accept, then compare segments around the pair only.** Pop the best pair; accept it if it clears the accept bar and keeps one-to-one. On accept, run segment comparison (layer 5) for each layer neighbor of the Subject against each corresponding research neighbor of the handle: degree × degree, a handful. Re-score and re-queue what changed; skip stale queue entries by version.
5. **Settle dirty pairs.** A pair is dirty when a neighbor was placed, moved or released after it was scored. When the queue empties, re-score only dirty pairs. A pair that falls under the bar, or that a rival beats by the settle margin, is released and re-queued, which makes its neighbors dirty. Changes ripple only as far as they reach. A round cap is the safety net. This replaces today's refinement loop (`refineOneHop`), which re-scores everything each round. Each move has a reason ("Alice moved because Gracie was placed"), which feeds the page's "updated because…" mark.
6. **Remember results across re-runs: last, and only if a test shows it's needed.** Layers 1–4 are pure in (Subject values, handle values), and a decision changes placements, not values, so their results could be memoized while the Promote page is open. This is the final step of the plan, gated by a performance test on a large Source. No pre-optimizing: if the test passes without it, it isn't built. If built, an invariant checks that cached and uncached runs give identical proposals.

**Registry knobs this adds:** shortlist keys per kind, K, the accept bar, the settle margin and round cap.

**Invariants** (from [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md)): order doesn't matter, decided rows don't move, stable on re-run, and the same proposal whether or not any row starts decided, given the same decisions.

### Bands

**Decided (2026-10-17).** A band (strong, medium, weak, no match) always describes a whole row: one Subject on the Source against its proposed handle. The per-property detail stays in the row's sheet.

| Step | Uses |
|---|---|
| Minimum to be a candidate; first sort for the top K | **similarity** (entity comparison) |
| Queue order, acceptance, order of alternatives, band | **evidence**: properties + segments |

- **Similarity gates and roughly sorts; it never decides.** Sparse Subjects tie: "John Smith" against 30 John Smiths is 1.0 for all of them. Every candidate tied at the K cutoff is kept, up to a hard cap, so the right one isn't cut before its segments are looked at. Nominated handles skip K entirely.
- **Evidence decides everything after that,** so the dropdown never ranks a name-only look-alike above the handle whose family lines up.
- **The accept bar is its own registry setting,** separate from the weak floor, even if they start equal. "Can the walk place this?" and "what does the researcher see?" are different questions.
- **A contested marker, not a cap.** A row whose comparisons include a conflict shows a visible "contested" mark, whatever its band. The band stays a pure evidence cutoff. Large registry penalties already handle impossible pairs.
- **The strong cutoff must be reachable by the sparse case:** a matching name plus a matching birth link. The sparse-data tier pins that case.

### Lifespan: a derived property for people

**Decided (2026-10-16): lifespan applies to people only.** Person matching is the chief job, and events and places differ enough that a lifespan wouldn't help them. The dates of the events a person takes part in constrain each other. An entity that died before the Subject was born is not that Subject, even when neither side has the other's event.

- **Decided (2026-10-18): a timeline, not per-bridge rules.** The profile builder lays every dated event a person takes part in on one line, any event type and any role. Its **span** runs from the first dated event to the last. Its **window** starts at the person's own birth or, with none, `MaxYears` before the first dated event, and ends at the person's own death or burial or, with none, `MaxYears` after the last. A side whose own events fall outside its own window has no lifespan. Age rules (marriage, parents) are dropped: the lifespan guides matching and is not an authority.
- **Subject comparison checks each side's span against the other side's window** as a registry-declared special case for people, using value compare's date relations. `Different` either way is a conflict (died 1840 against born 1850; born 1850 against died 1990; resided 1950 against died 1900). Anything else, `Near` included, is consistent.
- **Weighted as a conflict detector:** large conflict penalty, near-zero agree weight, because the dates it reads are already counted through their own segments.
- **Bounded by construction:** the builder reads only the person's own events, so nothing recurses across the tree. Both sides' lifespans can be built before alignment, once each.

**Built in S11-12.** `subjectprofile.Build` and `ForSubject` take the person's life events (event type, role, dates) and the registry's `subject.Lifespans` for the kind, and add two derived values, `lifespan` (the window) and `lifespan_span`. `subjectcompare.Compare` scores the `lifespan` lead from both directions: a conflict loses `subject.Weights.person.lifespan.Conflict`, consistent lifespans earn its `Agree`.

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
| A displayed value (several when tied leaders show together, or when the property is multi-valued) | positive |
| A value a stronger negative Observation denied (reason `denied`) | negative |
| Outvoted, weak, provisional | left out |

**The profile is strict (decided 2026-10-16).** Only the conclusion goes in. An outvoted value is, by the group's own judgment, not true. A real alternate, such as a maiden name, belongs in the conclusion: the auto-reconciler flags it, or the researcher records a Reconciliation Claim. Maiden names are not built yet, but that is the intended path. If a true value is missing from the conclusion, that is a reconciliation fix, not a matching rule. A single name mismatch is also weighed alongside sex, dates, relationships and the rest of the graph above this layer, so it doesn't decide a match alone.

**The profile builder is its own package:** `core/subjectprofile`, with `subjectprofile.Build`. It turns reconciler output into a Subject-shaped profile and compares nothing.

**Built in S11-11.** `subjectprofile.Build(entity, metas, rules, opts)` reads each Property's auto-reconciled values: values the reconciler displays become positives, values with reason `denied` negatives, and the rest are left out. A Property with nothing left is absent. The output is `subjectcompare.Values`, so entity compare is `Build` then `subjectcompare.Compare`, and entity against entity is `Compare` on two profiles. `Options.Excluding` runs `autoreconcile.Reconcile` again on each Property those members recorded, without their candidates. `subjectprofile.ForSubject` builds a plain Subject's side from its own Observations through the same conversion. A negative Observation that denies no positive value leaves no reconciled value, so it is not in the profile. Reconciliation Claims, when they ship, are laid over the profile as a replacing list.

**Migrated in S11-13.** Promote scores every candidate with entity compare. `promotealign` builds each Source Subject's side with `ForSubject` from the Source's own Observations and life events, and each handle's with `Build` from the cache and its participations, once per proposal; `graphalign` runs `subjectcompare.Compare` for every pair it scores, the duplicate-New check included. A handle is a candidate when its own properties reach the weak band (`graphalign.Candidates`, user 2026-10-18). The lifespan counts but is not shown until S11-20. `core/match` and `core/database/matching` are deleted.

## Frequency

**Decided:** frequency is not a comparison. Layers 1 and 2 never read catalog statistics.

**Decided (2026-10-16): name frequency is descoped.** "Common names count for less" is a subjective weighting. The one place it could matter is Subject-to-entity matching, and there the other properties and the graph (dates, sex, relationships, neighbors) already tell look-alikes apart. It adds noise and a tunable without earning its place.

- Remove `NameComparer.Frequency`, the discount, `DiscountComparer`, `CompareDiscount` and `Stats.Words` (shipped in S10-27).

**Decided (2026-10-16): all value-frequency weighting is descoped.** Promote also weights event types and toponyms from how common the value is in the catalog (`ln(m/u)` from `Stats.ValueFreq`, lifted to 2). Catalog frequency is the wrong reason for a weight:
- **Event type** carries little weight because it is generic, not because the catalog holds many births. "Both are births" says nothing specific about which birth. That is a fixed registry weight for the property.
- **Toponyms** match often. Their real weight comes from the place's relationships (`part_of`, the events located there), which graph alignment already scores.

So every property weight is a plain registry number. The labeler can fit them later. Remove `Stats.ValueFreq`, the `m` / `u` priors (`MPrior`, `UPrior`), the frequency scales and their lift to the accept bar. The toponym agreement bonus existed to work around the frequency weight; re-set it as a plain weight if still needed.

**Built in S11-24.** Promote's walk scores every property on a plain weight: 2 per value type, and name, sex, gender and toponym on their `subject.Weights`. `Stats.ValueFreq`, the priors and the lift are gone.

## Match registry

**Decided (2026-10-17): one registry package, `core/matchregistry`, for every matching tunable.** Today the numbers are spread over `core/match/registry.go`, `core/graphalign/registry.go`, `promotealign/canon.go` (`seedCandidateLimit`), `match/words.go` (`foldOnly`) and `autoreconcile/spelling.go` (`SpellingFloor`, `SpellingShortMin`; moved to `core/valuecompare` and `core/matchregistry` in S11-02). Two weight sets exist for the same property (`Rank`'s name 10, Promote's 2.5), and some keys are pipe-joined strings.

**One pure package, one `Default()`, one file per section:**

| File | Holds |
|---|---|
| `value.go` | Per value type: `Near` minimum similarity, date tolerance, the `ABT` unit rule, fuzzy floor, initial and alias credit, letter folds, particles, the built-in name pattern |
| `subject.go` | Per Subject type and property: agree weight, conflict penalty; pairs compared together (sex with gender, `date` with `start_date` / `end_date`); lifespan rules and constants |
| `segment.go` | Bridge cardinality (birth and death are one), neighbor credit, link support |
| `alignment.go` | Shortlist keys per kind, K, tie cap, accept bar, band cutoffs, settle margin and round cap |
| `evidence.go` | Claim evidence policy: which relations start checked or are pinnable, each kind's defining properties (event: `event_type`) |

**Rules:**
- **Each layer takes its own section as an argument** (rule 5). Tests override it without editing the defaults.
- **Typed keys,** not pipe-joined strings: `BridgeKey{Kind, Role, NeighborType, End}`, `PropertyKey`, `SubjectType`. A misspelled key fails to compile.
- **Catalog data is passed in at runtime,** never copied: alias tables, stored name patterns. Their install seeds stay where they are.
- **Data model facts stay in the catalog and vocabulary** (property cardinality, neutral / directed / inverse terms, bridge kinds). The registry refers to them by key. Bridge cardinality (a person has one birth) is different: it's a matching rule declared in the registry's segment section.
- **The reconciler keeps its own numbers** (`MinMajoritySupport`); it imports the spelling floor from value compare.
- **One weight set.** `Rank`'s separate scale goes away with the layers.

**Guard tests:**
- **No magic numbers:** a test scans the layer packages for numeric literals other than 0 and 1.
- **Knob inventory:** the first PR's inventory lists every registry field; a test fails when they drift.
- **One home for a fit:** the labeler's fit, later, proposes a diff to this package.

**Retires:** `core/match/registry.go`, `core/match/profiles.go`, `core/graphalign/registry.go`, `seedCandidateLimit`, `foldOnly`, the frequency settings.

## Reuse across the app

A survey of comparison, normalization and dedupe code outside `core/match` (2026-10-16). Each layer has consumers beyond Promote, and several duplicates can retire.

**Decided: all of this is Spike 11.** Dedupe happens as we go: the PR that lands a layer moves that layer's other callers onto it and deletes what it replaces. Nothing below is parked for later.

| Layer | Consumers today, once moved | Future consumers (`docs/ideas/`) |
|---|---|---|
| 1 value compare | Promote, `Rank`, the reconciler's grouping and pins, place-chain date windows (`conclusionheaders`) | Term dedupe ("godfather" vs "god-father", `term-management.md`), gazetteer lookup, Daitch–Mokotoff tier, international names |
| 2 property compare | Promote sheet exhibits (replacing `pairExhibits`), reconciler checks | Whether provisional values count (`possible-values.md`) |
| 3 subject compare | Duplicate New rows (`alikeNew`), merge hints (entity vs entity = two profiles) | Gazetteer place matching, duplicate Subjects while entering a Source |
| 4 entity compare | Promote candidates (S11-13) | Weak-claim review (`identity-claim-review.md`), share-package import (`share-packages.md`) |

**Small changes that widen reach:**
- **A normalized key per value type** beside `Compare`: `valuecompare.Key(v, level)`, where `level` is `Same` or `Equivalent`. Two values have the same key exactly when `Compare` says that relation. One function then serves reconciler grouping, search index dedupe, term dedupe and sort-independent equality, instead of the six text rules there are today.
- **Leave-one-out profiles:** `subjectprofile.Build(entity, metas, rules, Options{Excluding: members})`. "Does this member still fit the rest?" is the weak-claim review, and costs one option.
- **Profiles on both sides:** because a profile is Subject-shaped, entity-vs-entity (merge hints) is layer 3 on two profiles, with no new code.
- **The date window primitives** (`Window` with `Overlaps`, `Contains`, `Intersect`) moved into value compare in S11-04, with `AboutEnds` and `EndsWindow`; `DateBounds` and `WindowOfDate` became `valuecompare.DateWindow`, and `PeriodWindow` and `LinkMembership` moved to `conclusionheaders` on it, in S11-06, so place chains and reconciliation use the same day-precision rules as matching.

**Duplicates to retire:**

| Duplicate | Replaced by |
|---|---|
| `autoreconcile` `textKey`, `termKey`, `dateKey`, `subsumes` / `partFits`, `IsInitial`, `SpellingSimilarity`, `EditDistance`, `NormalizeForm`, `DateEnds` and windows (`DateEnds` and windows moved in S11-04) | Value compare (the reconciler imports it) — done in S11-06 |
| `autoreconcile.Compatible` | Value compare's relation, under the reconciler's own grouping policy (rule 7) — done in S11-06 |
| `promotealign.pairExhibits` (its own pairing; ignores negatives and cardinality) | Property compare's speaking pair |
| `match/pairs.go` leaders and weighting (`weigh`, `pairValues`) | The strict profile; negatives and cardinality move to property compare — done in S11-13 |
| `graphalign` `alikeNew`, `pairNode` (property part) | Subject compare — done in S11-13 |
| `conclusionheaders` `laterStart` / `earlierEnd` (year only, while reconciliation uses days) | Value compare's date windows |
| ~~`autoreconciler.hasValue`~~ — done in S10-52: `autoreconcile.Carries` | One function |
| `Stats.ValueFreq`, `countNameWords` | Removed (frequency descoped) |

**Lighter touch (still in scope):**
- **Search's fuzzy ranker** (`jaroWinkler`, trigram shortlist). It ranks partial queries as they are typed, which is a different job, so it stays. Its Go rescoring moves onto the shared normalization key, so it folds accents like the FTS tokenizer already does.
- **Swift filters** (combo box, list filters). They are UI-local, with different needs. They move onto one shared Swift normalize helper. Done in S11-08: `TextMatch`, checked against `valuecompare.TextKey` by `TestSwiftTextMatchMirrorsTextKey`.
- **`FakeStore`** clusters names by lowercase form, which contradicts Go's parts-only rule. Fix that when the reconciler moves, so Swift tests don't encode the wrong rule. Done in S11-08: names group by parts at `Same`.

**Found in passing (fixed in this spike):**
- `match.indexAliases` doesn't normalize alias tokens. That's latent: only the lowercase product seed writes tables today. It matters once user or plugin tables exist; `Key` fixes it.
- Text equality disagreed across six places (reconciler, match, search scoring, FTS, header sort, frequency stats). After S11-07 the reconciler, search scoring and the term duplicate check share `Key` (`valuecompare.TextKey`), and `core/match` went in S11-13. FTS and the header sort (`SortKey`, ordering only) keep their own rules on purpose.

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

## Vocabulary: finding candidates

Agreed 2026-10-17. Record linkage calls the first step "blocking", which reads as "rejected"; we don't use that word.

The terms are defined in the model's [glossary](../../matching-model.md#glossary): shortlisting, shortlist key, shortlisted handle, nominated handle, candidate.

Flow: shortlist or nominate → score → keep the top K as candidates → accept.

## Decided on 2026-10-17

- **Neighbor credit travels one hop.** Start with one walk and one hop; chains come from the walk. Revisit only if goldens show chains under-rewarded.
- **Misspellings and shortlisting: start simple.** Accept the gap ("Robnis" vs "Robins" shares no exact key); links catch most of it. A coarse key (surname consonant skeleton) is the fallback if the sparse-data tier shows a problem.
- **Reconciler grouping rules** are the reconciler's own decision, designed in its PR. Today's disagreements to resolve there: `Montréal` / `Montreal` and `Jim` / `James` agree in matching but don't group or pin; `J. Robins` / `James Robins` groups but only scores partial. Decided in S11-06: `Same` groups, `Contains` folds, `Equivalent` never; text never folds; ABT is a range.

## Open questions

- None for the design. Starting registry numbers follow the map's working rule "Registry numbers: carry over, then estimate" ([`deployment-plan.md`](deployment-plan.md#working-rules)).
