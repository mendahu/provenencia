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
- **Pure helpers are shared, not copied.** The word helpers live in `core/valuecompare/words.go`.

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
5. `Key` must agree with `Compare`: equal at `Same` exactly when the relation is `Same`, equal at `Equivalent` exactly when `Same` or `Equivalent` (`TestKeyContract`).

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
- Weaken a guard or add an allow-list to make it pass.
