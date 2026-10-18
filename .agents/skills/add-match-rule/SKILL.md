---
name: add-match-rule
description: >-
  Adds or changes a Provenencia matching layer rule: a registry knob, a value
  comparer, a relation, cause or unit code, or a new layer package
  (core/valuecompare, core/matchregistry). Use when adding a number the
  matching algorithm tunes, a comparer for a value type, a Relation or Cause
  code, a layer package, or a row in the matching-model.md knob inventory.
---

# Add or change a matching rule

The matching stack is layered ([`docs/deployment-plan/spike-11/comparison-layers.md`](../../../docs/deployment-plan/spike-11/comparison-layers.md), vocabulary in [`docs/matching-model.md`](../../../docs/matching-model.md)). Layer 1, `core/valuecompare`, relates two values of one type and returns a relation, a similarity and a detail. `core/matchregistry` holds every number. This skill grows with each layer; later layer PRs extend it.

## When it applies

You add or change a tunable, a comparer, a relation, cause or unit code, a layer package, or a row of the knob inventory.

## Rules that hold everywhere

- **Layers are pure.** No catalog, no I/O, no build tags. A layer package never imports `core/database`, directly or through another package, and never imports `core/matchregistry` or `core/autoreconcile`.
- **Types live in the layer, values in the registry.** The layer declares its `Rules` groups (field types and doc comments, no numbers). `core/matchregistry` declares `Registry` and the numbers, with a comment on each saying where it came from. A layer takes its section as an argument.
- **`matchregistry.Default()` builds fresh slices and maps** on each call. A test overrides a section by editing its own copy.
- **Numbers only in `matchregistry`.** A layer package marked `numbers` in `layerPackages` may hold no numeric literal other than 0 and 1. A fact that is not a tunable (a calendar fact) goes in its pure value package.
- **Pure helpers are shared, not copied.** The word helpers live in `core/valuecompare/words.go`. `wordSimilarity` is the one word table (names add initials and alias tables through `wordOptions`), `BestAssignment` (`pairing.go`) the one pairing search, `AliasIndex` the one alias lookup, and the weighted `meanSimilarity` (`text.go`) the one rule for combining part similarities.
- **Catalog data comes in through `Rules`.** A name value names its pattern by key (`Value.NamePattern`) and the layer reads it from `Rules.Names.Patterns`; alias tables come as `Rules.Names.Aliases`. Catalog data is not a knob: list its registry path in `catalogData` in `core/matchregistry/inventory_test.go` and name it under "Not knobs" in `docs/matching-model.md`, with no inventory row. A caller gets a catalog's rules from `valuerules.Load` (`core/database/valuerules`): the stored default name pattern under its key and the alias index, on a copy of `matchregistry.Default().Value`. New catalog data for `Rules` is loaded there.
- **A doubled weight is written `w + w`.** The numbers guard allows no literal 2 in a layer package.
- **Dates compute similarity through `falloff`.** Calendar facts (days in a month, day numbers, the year line) live in `core/datevalue/calendar.go`, not in the layer. The window algebra is `valuecompare.Window` (`core/valuecompare/datewindow.go`): `Overlaps`, `Contains`, `Intersect` as methods, and `AboutEnds` with `EndsWindow` to read a date as compared.
- **The reconciler composes value compare under its own policy.** `autoreconcile.Reconcile` and `Compatible` take `valuecompare.Rules`; values group on `Key` at `Same` and fold on `Contains`. The cache passes `matchregistry.Default().Value`, never catalog data, so a rule change is a `CacheVersion` bump. `valuecompare.DateWindow` reads a date as written (sort keys, stored bounds, place periods).

## Add a knob

1. Add the field and a doc comment saying what it changes to the layer's `Rules` group (`core/valuecompare/rules.go`).
2. Set the value in `core/matchregistry/<section>.go` with a comment naming its source.
3. Add a row to the knob inventory in [`docs/matching-model.md`](../../../docs/matching-model.md) (key, start, source, changes). The key is the dotted path of the field.
4. When the group's layer has landed, delete its prefix from `notYetBuilt` in `core/matchregistry/inventory_test.go`. `TestKnobInventoryMatchesRegistry` fails on a field with no row, a row with no field, and a prefix that still matches.

## Add a comparer

1. Add a `comparer` (`compare` and `key`) and an entry in `comparers` in `core/valuecompare/valuecompare.go`, keyed by the `propertymeta.ValueType*` constant.
2. Add a sample row and a generator to `core/valuecompare/contract_test.go`; `TestComparerContract` fails without them.
3. Add the type to the expected `ComparerTypes()` list in `TestCompareUnsupported`.
4. A type with a distance computes its similarity through `falloff`, or the plan records why not.
5. `Key` must agree with `Compare`: equal at `Same` exactly when the relation is `Same`, equal at `Equivalent` exactly when `Same` or `Equivalent` (`TestKeyContract`). Agreement holds under rules with no alias tables; name keys ignore aliases (S11-03 D-10).

## Add a relation, cause or unit code

Add the constant in `core/valuecompare/valuecompare.go` and, for a cause, to the cause list in `comparison-layers.md` §Detail. Codes are strings; they reach the wire and UI copy later.

## Add a layer package

Add a row to `layerPackages` in `core/matchregistry/guard_test.go`, with `numbers` true for a layer and false for a pure value package.

## Check

```
go test ./core/valuecompare/... ./core/matchregistry/...
```

No build tags. Then `scripts/check-pr.sh`.

## Do not

- Keep a second copy of a word helper (`NormalizeForm`, `SplitWords`, `SpellingSimilarity`).
- Write a number in a layer package.
- Read the catalog in a layer.
- Import `core/autoreconcile` from a layer package.
- Weaken a guard or add an allow-list to make it pass. The one list the inventory guard keeps, `catalogData`, names fields that are catalog data, never tunables; each is named under Not knobs.
