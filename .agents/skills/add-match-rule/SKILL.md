---
name: add-match-rule
description: >-
  Adds or changes a Provenencia matching layer rule: a registry knob, a value
  comparer, a relation, cause or unit code, or a new layer package
  (core/valuecompare, core/propertycompare, core/subjectcompare,
  core/subjectprofile, core/matchregistry). Use when adding a number the
  matching algorithm tunes, a comparer for a value type, a Relation or Cause
  code, a Subject property weight, pool or period, an entity profile rule, a
  layer package, or a row in the matching-model.md knob inventory.
---

# Add or change a matching rule

The matching stack is layered ([`docs/deployment-plan/spike-11/comparison-layers.md`](../../../docs/deployment-plan/spike-11/comparison-layers.md), vocabulary in [`docs/matching-model.md`](../../../docs/matching-model.md)). Layer 1, `core/valuecompare`, relates two values of one type and returns a relation, a similarity and a detail. `core/matchregistry` holds every number. This skill grows with each layer; later layer PRs extend it.

## When it applies

You add or change a tunable, a comparer, a relation, cause or unit code, a layer package, or a row of the knob inventory.

## Rules that hold everywhere

- **Layers are pure.** No catalog, no I/O, no build tags. A layer package never imports `core/database`, directly or through another package, and never imports `core/matchregistry`; only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile.
- **Types live in the layer, values in the registry.** The layer declares its `Rules` groups (field types and doc comments, no numbers). `core/matchregistry` declares `Registry` and the numbers, with a comment on each saying where it came from. A layer takes its section as an argument.
- **`matchregistry.Default()` builds fresh slices and maps** on each call. A test overrides a section by editing its own copy.
- **Numbers only in `matchregistry`.** A layer package marked `numbers` in `layerPackages` may hold no numeric literal other than 0 and 1. A fact that is not a tunable (a calendar fact) goes in its pure value package.
- **Pure helpers are shared, not copied.** The word helpers live in `core/valuecompare/words.go`. `wordSimilarity` is the one word table (names add initials and alias tables through `wordOptions`), `BestAssignment` (`pairing.go`) the one pairing search, `AliasIndex` the one alias lookup, and the weighted `meanSimilarity` (`text.go`) the one rule for combining part similarities. `valuecompare.EndsDate` is the one way back from date ends to a date, and `valuecompare.Earliest` / `Latest` the one earliest / latest pick of date ends.
- **Catalog data comes in through `Rules`.** A name value names its pattern by key (`Value.NamePattern`) and the layer reads it from `Rules.Names.Patterns`; alias tables come as `Rules.Names.Aliases`. Catalog data is not a knob: list its registry path in `catalogData` in `core/matchregistry/inventory_test.go` and name it under "Not knobs" in `docs/matching-model.md`, with no inventory row. A caller gets a catalog's rules from `valuerules.Load` (`core/database/valuerules`): the stored default name pattern under its key and the alias index, on a copy of `matchregistry.Default().Value`. New catalog data for `Rules` is loaded there.
- **Layer 2 (`core/propertycompare`) reads relations, never numbers.** `Compare` takes the `value` section and has no section of its own; the outcome comes from the speaking pair's relation through `outcomeOf`, and relations rank with `valuecompare.Stronger`, the one relation order. A value carries only `Negative` and `Neutral` beside its layer 1 value: support, standing and denied are turned into positives and negatives by the profile builder before they reach it.
- **Layer 3 (`core/subjectcompare`) turns outcomes into points.** It reads only the lead properties in `subject.Weights` for the kind; agree earns the agree weight, partial agree weight × similarity, conflict loses the penalty and weighs it in similarity. Registry maps are keyed by `vocabkey.SubjectType` and `vocabkey.PropertyKey`, written as `vocabkey` constants, never as string literals; no other package declares a key type. Value type and cardinality come from the catalog as `Metas`, never from the registry; a derived property's meta is fixed beside the rule that reads it (the lifespan's date meta). Tests build values with `propertycomparetest` and metas with `subjectcomparetest.Metas`. `subject.Lifespans` declares a lifespan lead (the window) and its `Span`; `Compare` checks each side's span against the other's window, and a lifespan's agree means consistent, whatever partial relation spoke.
- **The profile builder (`core/subjectprofile`) turns the reconciler's output into a Subject.** `Build` keeps a Property's displayed values (`ReconciledValue.Displayed`) as positives and its denied values as negatives, and leaves everything else out; `ForSubject` turns a Subject's own Observations into positives and negatives through the same conversion. Layer 3 compares either like any Subject, so entity compare is `Build` then `subjectcompare.Compare`. `Options.Excluding` reconciles each Property those members recorded again without their candidates. Callers pass `Build` `matchregistry.Default().Value`, the rules the cache was built with. The builder reads `subject.Lifespans` (passed in) to build a person's window and span from its life events.
- **A doubled weight is written `w + w`.** The numbers guard allows no literal 2 in a layer package.
- **Dates compute similarity through `falloff`.** Calendar facts (days in a month, day numbers, the year line) live in `core/datevalue/calendar.go`, not in the layer. The window algebra is `valuecompare.Window` (`core/valuecompare/datewindow.go`): `Overlaps`, `Contains`, `Intersect` as methods, and `AboutEnds` with `EndsWindow` to read a date as compared. `ShiftYears` is the year move on an end (then `boundEnd` places it); do not move years by hand.
- **The reconciler composes value compare under its own policy.** `autoreconcile.Reconcile` and `Compatible` take `valuecompare.Rules`; values group on `Key` at `Same` and fold on `Contains`. The cache passes `matchregistry.Default().Value`, never catalog data, so a rule change is a `CacheVersion` bump. `valuecompare.DateWindow` reads a date as written (sort keys, stored bounds, place periods).
- **Callers outside matching key text with `valuecompare.TextKey` and the shipped word rules** (`matchregistry.Default().Value.Words`), not `valuerules.Load`: a text key reads no catalog data. A change to `text.key` or `WordRules` bumps `autoreconciler.CacheVersion` and `searchindex.ProjectionVersion`. The key has a Swift mirror, `macos/App/Platform/TextMatch.swift` (with `Words.Folds`); a change to either side fails `TestSwiftTextMatchMirrorsTextKey` (`api/ffi`) until both agree.

## Add a knob

1. Add the field and a doc comment saying what it changes to the layer's `Rules` group (`core/valuecompare/rules.go`).
2. Set the value in `core/matchregistry/<section>.go` with a comment naming its source.
3. Add a row to the knob inventory in [`docs/matching-model.md`](../../../docs/matching-model.md) (key, start, source, changes). The key is the dotted path of the field.
4. When the group's layer has landed, delete its prefix from `notYetBuilt` in `core/matchregistry/inventory_test.go`. `TestKnobInventoryMatchesRegistry` fails on a field with no row, a row with no field, and a prefix that still matches.

## Add a Subject property or pair

1. Add the key to `core/vocabkey` if it has none, and to the key list in `TestVocabKeysMatchSeed`. A derived key (built by the profile builder, never seeded) is not added there.
2. Add its `Weights` (or a `Pools` / `Periods` / `Lifespans` entry under an existing weighted lead) in `core/matchregistry/subject.go`, with a source comment.
3. Add the inventory rows: `subject.Weights.<kind>.<property>.Agree` and `.Conflict`, `subject.Pools.<kind>.<lead>.With`, or `subject.Periods.<kind>.<lead>.Start` and `.End`, or `subject.Lifespans.<kind>.<lead>.Span`, `.MaxYears`, `.Starts` and `.Ends`. A new life-fixing event is a keyed `EventRole` in `Starts` or `Ends`.
4. If the property is new to `subjectcomparetest.Metas`, add it there. `TestSubjectSection` and `TestSubjectSectionMatchesSeed` (`core/database/subjectvocab`) fail on a key the seed does not bind to the kind, a pair whose lead has no weights, a pooled property of another value type, and a fixture meta that differs from the seed.
5. Add a `TestCompare` row for it in `core/subjectcompare`.

## Add a comparer

1. Add a `comparer` (`compare` and `key`) and an entry in `comparers` in `core/valuecompare/valuecompare.go`, keyed by the `propertymeta.ValueType*` constant.
2. Add a sample row and a generator to `core/valuecompare/contract_test.go`; `TestComparerContract` fails without them.
3. Add the type to the expected `ComparerTypes()` list in `TestCompareUnsupported`.
4. A type with a distance computes its similarity through `falloff`, or the plan records why not.
5. `Key` must agree with `Compare`: equal at `Same` exactly when the relation is `Same`, equal at `Equivalent` exactly when `Same` or `Equivalent` (`TestKeyContract`). Agreement holds under rules with no alias tables; name keys ignore aliases (S11-03 D-10).

## Add a relation, cause or unit code

A relation, a relation cause or a unit is layer 1's: add the constant in `core/valuecompare/valuecompare.go`, a new relation also in `relationOrder` at its strength, its outcome in `propertycompare.outcomeOf` and whether a negative contradicts it in `contradicts` (`TestEveryRelationHasAnOutcome` fails until both are decided), and a cause to the cause list in `comparison-layers.md` §Detail. An outcome cause or an exclude reason is layer 2's: add it in `core/propertycompare/propertycompare.go` and to `comparison-layers.md` §Layer 2 Output. Codes are strings; they reach the wire and UI copy later.

## Add a layer package

Add a row to `layerPackages` in `core/matchregistry/guard_test.go`, with `numbers` true for a layer and false for a pure value or fixture package. Add the package's glob to `.cursor/rules/match-rules.mdc` and its path to the Check command below. Give it a package doc naming its layer and saying it is pure. Test it from an external `_test` package with an exact table and a fixed-seed invariants test, building fixtures with `namevaluetest` and `datevaluetest` (`core/propertycompare` is the model).

## Check

```
go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/matchregistry/...
```

and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` after a subject section change.

No build tags. Then `scripts/check-pr.sh`.

## Do not

- Keep a second copy of a word helper (`NormalizeForm`, `SplitWords`, `SpellingSimilarity`).
- Write a number in a layer package.
- Read the catalog in a layer.
- Import `core/autoreconcile` from a layer package other than `core/subjectprofile`.
- Weaken a guard or add an allow-list to make it pass. The one list the inventory guard keeps, `catalogData`, names fields that are catalog data, never tunables; each is named under Not knobs.
