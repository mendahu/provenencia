---
name: add-match-rule
description: >-
  Adds or changes a Provenencia matching layer rule: a registry knob, a value
  comparer, a relation, cause or unit code, or a new layer package
  (core/valuecompare, core/propertycompare, core/subjectcompare,
  core/subjectprofile, core/segmentcompare, core/kinship, core/shortlist, core/graphalign, core/claimevidence, core/matchregistry). Use when adding a number the
  matching algorithm tunes, a comparer for a value type, a Relation or Cause
  code, a Subject property weight, pool or period, an entity profile rule, a
  kinship pattern, a layer package, or a row in the matching-model.md knob inventory.
---

# Add or change a matching rule

The matching stack is layered ([`docs/deployment-plan/spike-11/comparison-layers.md`](../../../docs/deployment-plan/spike-11/comparison-layers.md), vocabulary in [`docs/matching-model.md`](../../../docs/matching-model.md)). Layer 1, `core/valuecompare`, relates two values of one type and returns a relation, a similarity and a detail. `core/matchregistry` holds every number. Promote is the layers' caller: `core/database/promotealign` builds each side with the profile builder and `core/graphalign` scores pairs with `subjectcompare.Compare`; a new knob reaches Promote through the registry, with no Promote change. This skill grows with each layer; later layer PRs extend it.

## When it applies

You add or change a tunable, a comparer, a relation, cause or unit code, a layer package, or a row of the knob inventory.

## Rules that hold everywhere

- **Layers are pure.** No catalog, no I/O, no build tags. A layer package never imports `core/database`, directly or through another package, and never imports `core/matchregistry` (`TestLayerPackages` walks every non-test import chain); only `core/subjectprofile` imports `core/autoreconcile`, to turn its output into a profile.
- **Types live in the layer, values in the registry.** The layer declares its `Rules` groups (field types and doc comments, no numbers). `core/matchregistry` declares `Registry` and the numbers, with a comment on each saying where it came from. A layer takes its section as an argument.
- **`matchregistry.Default()` builds fresh slices and maps** on each call. A test overrides a section by editing its own copy.
- **Numbers only in `matchregistry`.** A layer package marked `numbers` in `layerPackages` may hold no numeric literal other than 0 and 1. A fact that is not a tunable (a calendar fact) goes in its pure value package.
- **Pure helpers are shared, not copied.** The word helpers live in `core/valuecompare/words.go`. `wordSimilarity` is the one word table (names add initials and alias tables through `wordOptions`), `BestAssignment` (`pairing.go`) the one pairing search, `AliasIndex` the one alias lookup, and the weighted `meanSimilarity` (`text.go`) the one rule for combining part similarities. `valuecompare.EndsDate` is the one way back from date ends to a date, and `valuecompare.Earliest` / `Latest` the one earliest / latest pick of date ends.
- **Catalog data comes in through `Rules`.** A name value names its pattern by key (`Value.NamePattern`) and the layer reads it from `Rules.Names.Patterns`; alias tables come as `Rules.Names.Aliases`. Catalog data is not a knob: list its registry path in `catalogData` in `core/matchregistry/inventory_test.go` and name it under "Not knobs" in `docs/matching-model.md`, with no inventory row. A caller gets a catalog's rules from `valuerules.Load` (`core/database/valuerules`): the stored default name pattern under its key and the alias index, on a copy of `matchregistry.Default().Value`. New catalog data for `Rules` is loaded there.
- **Layer 2 (`core/propertycompare`) reads relations, never numbers.** `Compare` takes the `value` section and has no section of its own; the outcome comes from the speaking pair's relation through `outcomeOf`, and relations rank with `valuecompare.Stronger`, the one relation order. A value carries only `Negative` and `Neutral` beside its layer 1 value: support, standing and denied are turned into positives and negatives by the profile builder before they reach it.
- **Layer 3 (`core/subjectcompare`) turns outcomes into points.** It reads only the lead properties in `subject.Weights` for the kind; agree earns the agree weight, partial agree weight × similarity, conflict loses the penalty and weighs it in similarity. Registry maps are keyed by `vocabkey.SubjectType` and `vocabkey.PropertyKey`, written as `vocabkey` constants, never as string literals; no other package declares a key type. Value type and cardinality come from the catalog as `Metas`, never from the registry; a derived property's meta is fixed beside the rule that reads it (the lifespan's date meta). `subjectcompare.Points` is what one property adds to evidence, for the score and the sheet alike. `Rules.Lead` maps a pooled or period property to its lead. `Rules.Derived` names a lead with no catalog row (the lifespan); a new derived property is declared in the registry, not detected from a missing meta. Tests build values with `propertycomparetest` and metas with `subjectcomparetest.Metas`. `subject.Lifespans` declares a lifespan lead (the window) and its `Span`; `Compare` checks each side's span against the other's window, and a lifespan's agree means consistent, whatever partial relation spoke.
- **The profile builder (`core/subjectprofile`) turns the reconciler's output into a Subject.** `Build` keeps a Property's displayed values (`ReconciledValue.Displayed`) as positives and its denied values as negatives, and leaves everything else out; `ForSubject` turns a Subject's own Observations into positives and negatives through the same conversion. Layer 3 compares either like any Subject, so entity compare is `Build` then `subjectcompare.Compare`. `Options.Excluding` reconciles each Property those members recorded again without their candidates. Callers pass `Build` `matchregistry.Default().Value`, the rules the cache was built with. The builder reads `subject.Lifespans` (passed in) to build a person's window and span from its life events.
- **Layer 5 (`core/segmentcompare`) scores one link.** Keys come from `Key` with the catalog's `Terms`; outside tests, a bridge reads as its keys through `Keys` (or as a `Link` through `LinkOf`), with its event end from `EventEndpoint`; correspondence is `Corresponds`, except a pair `kinship.CountedOnce` skips at layer 6. `One` (keys with an empty `End`, each holding from both ends) and `Gates` (per end) are keyed by `vocabkey.BridgeKey`. Placement (`Together`, `Apart` or `Unplaced`) comes from layer 6; the neighbor's evidence comes through layer 3. Tests build terms with `segmentcomparetest.Terms`.
- **Kinship patterns (`core/kinship`) add implied links before alignment.** `Derive(edges, terms, rules)` reads one side's links (`Edge`: ends, keys from each end, bridge Subjects as `Associations`, the fields of `graphalign.Bridge`) through `eventPatterns` and one parent-of-parent chain rule and returns only the implied relationship links, keyed through `segmentcompare.Keys`; it recognizes a link by its keys alone. A pivot with more than `Rules.MaxPivotEnd` participants at either end implies nothing (the registry's `kinship` section). The canon's links are read through `core/database/canonlinks` (`EdgeOf`, `LoadTerms`, `Reader.For`, which answers one person's implied links from one cached result per pivot on the graph cache, and `Reader.Relationships`, which returns one person's recorded and implied links merged through `kinship.LinkSet`). Implied links enter a side only through `kinship.LinkSet`'s `AddImplied` (or `PlaceImplied` in `loadCanon`, each recorded by `Implied.Recording`), the one same-ends, same-key merge for recorded and implied links, and carry a `kinship.Recording` (their shapes, and whether only implied) into `graphalign.Bridge` and `CanonEdge`. Layer 6's `corresponding` skips a pair `kinship.CountedOnce` reports. An implied link weighs as the direct link it stands for: no knob.
- **Layer 6 (`core/graphalign`) walks and decides.** Candidates come from the shortlist through `Candidates` or from a link to a placed pair; scores are replaced, never kept higher; placement is `Together`, `Apart` or `Unplaced`; bands, contested and moves are row output. Its section is `graphalign.Rules`.
- **The evidence section (`core/claimevidence`) is the claim evidence view's policy, not a matching layer.** `Rules` lists the relations whose lines start checked (`Checked`), and the causes of an unknown line the researcher judges by eye, kept but never checked (`ByEye`; two different likenesses are `unjudged`), and per kind the properties a relationship line bundles (`Defining`); the inventory rows are `evidence.Checked`, `evidence.ByEye` and `evidence.Defining`. Sections are in the kind's binding order (`Inputs.Order`, from `Vocabulary.BindingOrder`, ranked by `vocabkey.CompareByOrder`), never a second order in the registry; the lifespan is a matching signal only, so `claimevidence` never shows it. `Sections` compares each incoming value that names an Observation with each accepted member (layer 2, never Promote's scores) and collapses members with an identical value into one line; `Build` adds a relationship for each link that corresponds through `segmentcompare`. A line's points come from `subjectcompare.Points`; no caller reads the speaking similarity itself (`propertycompare.Result.Similarity`). The loader, `promotealign.Evidence`, only reads. A new defining property is one `Defining` entry.
- **Shortlist keys come from layer 1.** `WordKeys` and `YearReach` make them; `core/shortlist` only buckets and indexes. A lead whose agree weight alone can reach `alignment.Bands.Weak` needs a shortlist entry (`TestUnkeyedEvidenceStaysUnderTheWeakBand`).
- **A doubled weight is written `w + w`.** The numbers guard allows no literal 2 in a layer package.
- **Dates compute similarity through `falloff`.** Calendar facts (days in a month, day numbers, the year line) live in `core/datevalue/calendar.go`, not in the layer. The window algebra is `valuecompare.Window` (`core/valuecompare/datewindow.go`): `Overlaps`, `Contains`, `Intersect` as methods, and `AboutEnds` with `EndsWindow` to read a date as compared. `ShiftYears` is the year move on an end (then `boundEnd` places it); do not move years by hand.
- **The reconciler composes value compare under its own policy.** `autoreconcile.Reconcile` takes `valuecompare.Rules`; values group on `Key` at `Same` and fold on `Contains`. The cache passes `matchregistry.Default().Value`, never catalog data, so a rule change is a `CacheVersion` bump. `valuecompare.DateWindow` reads a date as written (sort keys, stored bounds, place periods).
- **Callers outside matching key text with `valuecompare.TextKey` and the shipped word rules** (`matchregistry.Default().Value.Words`), not `valuerules.Load`: a text key reads no catalog data. A change to `text.key` or `WordRules` bumps `autoreconciler.CacheVersion` and `searchindex.ProjectionVersion`. The key has a Swift mirror, `macos/App/Platform/TextMatch.swift` (with `Words.Folds`); a change to either side fails `TestSwiftTextMatchMirrorsTextKey` (`api/ffi`) until both agree.

## Add a knob

1. Add the field and a doc comment saying what it changes to the layer's `Rules` group (`core/valuecompare/rules.go`; `core/graphalign/rules.go` for a walk knob).
2. Set the value in `core/matchregistry/<section>.go` with a comment naming its source.
3. Add a row to the knob inventory in [`docs/matching-model.md`](../../../docs/matching-model.md) (key, start, source, changes). The key is the dotted path of the field.
4. When the group's layer has landed, delete its prefix from `notYetBuilt` in `core/matchregistry/inventory_test.go`. `TestKnobInventoryMatchesRegistry` fails on a field with no row, a row with no field, and a prefix that still matches.

## Add a Subject property or pair

1. Add the key to `core/vocabkey` if it has none, and to the key list in `TestVocabKeysMatchSeed`. A derived key (built by the profile builder, never seeded) is not added there.
2. Add its `Weights` (or a `Pools` / `Periods` / `Lifespans` entry under an existing weighted lead) in `core/matchregistry/subject.go`, with a source comment.
3. Add the inventory rows: `subject.Weights.<kind>.<property>.Agree` and `.Conflict`, `subject.Pools.<kind>.<lead>.With` (and `.Terms.<pooled>` when the pooled property's terms differ from the lead's), or `subject.Periods.<kind>.<lead>.Start` and `.End`, or `subject.Lifespans.<kind>.<lead>.Span`, `.MaxYears`, `.Starts` and `.Ends`. A new life-fixing event is a keyed `EventRole` in `Starts` or `Ends`.
4. If the property is new to `subjectcomparetest.Metas`, add it there. `TestSubjectSection` and `TestSubjectSectionMatchesSeed` (`core/database/subjectvocab`) fail on a key the seed does not bind to the kind, a pair whose lead has no weights, a pooled property of another value type, a pool `Terms` entry that is not a seeded term of its property, and a fixture meta that differs from the seed. A pooled term property whose keys differ from the lead's (gender `man`, sex at birth `male`) needs a `Terms` crosswalk, or every pair of its values conflicts.
5. Add a `TestCompare` row for it in `core/subjectcompare`.

## Add a comparer

1. Add a `comparer` (`compare` and `key`) and an entry in `comparers` in `core/valuecompare/valuecompare.go`, keyed by the `propertymeta.ValueType*` constant.
2. Add a sample row and a generator to `core/valuecompare/contract_test.go`; `TestComparerContract` fails without them.
3. Add the type to the expected `ComparerTypes()` list in `TestCompareUnsupported`.
4. A type with a distance computes its similarity through `falloff`, or the plan records why not.
5. `Key` must agree with `Compare`: equal at `Same` exactly when the relation is `Same`, equal at `Equivalent` exactly when `Same` or `Equivalent` (`TestKeyContract`). Agreement holds under rules with no alias tables; name keys ignore aliases (S11-03 D-10).

6. A type that shortlists by words implements `wordKeys`; add forms to `TestWordKeysShareOnSameOrEquivalent`.

## Add a shortlist key

1. Add the lead to `Words` or a kind to `Dated` in `core/matchregistry/alignment.go`.
2. The lead is a catalog property, never a derived one (`Lifespans`): `keyValues` builds no life (`TestShortlistSection`).
3. Add its inventory row (`alignment.Shortlist.Words`, `alignment.Shortlist.Dated.<kind>.Type` / `.Date`).
4. A value type the lead uses implements `wordKeys` or has a `YearReach`; extend `TestWordKeysShareOnSameOrEquivalent`'s or `TestYearReachCoversEveryRelatedPair`'s corpus.

## Add a relation, cause or unit code

A relation, a relation cause or a unit is layer 1's: add the constant in `core/valuecompare/valuecompare.go`, a new relation also in `relationOrder` at its strength, its outcome in `propertycompare.outcomeOf` and whether a negative contradicts it in `contradicts` (`TestEveryRelationHasAnOutcome` fails until both are decided), and a cause to the cause list in `comparison-layers.md` §Detail. An outcome cause or an exclude reason is layer 2's: add it in `core/propertycompare/propertycompare.go` and to `comparison-layers.md` §Layer 2 Output. A segment cause is layer 5's: add it in `core/segmentcompare` and to the layer 5 output list in `comparison-layers.md` §Segments. A segment cause stays in Go: the evidence sheet does not read it, so it has no Swift enum. Codes are strings. A code that reaches the wire (a relation, a value cause, an outcome or an outcome cause) also gets a case in its enum in `macos/App/Platform/ConclusionVocabulary.swift` (`ValueRelation`, `ValueCause`, `EvidenceOutcome`, `EvidencePropertyCause`), and stays in its row of `TestSwiftVocabularyMatchesGo` (`api/ffi/swift_vocabulary_test.go`), which fails when the two sides drift. UI copy built from it follows.

A move reason is layer 6's: a `Reason` constant in `core/graphalign/types.go`.

## Add a link rule

A one-cardinality link is a `One` key, and a gate a `Gates` entry, in `core/matchregistry/segment.go`, each with a source comment. Add its inventory row named by the key's `String()` (`segment.Gates.place_relationship·part_of·from.MinOwnEvidence`), make sure `TestSegmentSectionMatchesSeed` passes, and add a `TestCompare` row in `core/segmentcompare`.

## Add a kinship pattern

1. Add a `vocabkey` constant for each term the pattern names, and its row in `TestVocabKeysMatchSeed`. Product terms only: a catalog's own terms imply nothing.
2. An event pattern is a row in `eventPatterns` (an event type, the roles at each end, the implied relationship term) in `core/kinship/patterns.go`, with a new `Pattern` code. The chain is one rule in `Derive`, parent of parent only; a pattern of another shape (siblings, great-grandparents) changes `Derive` and is planned as its own PR.
3. Add `TestDerive` rows: the pattern holds, a near miss (another role or event type) does not, and the implied link's direction.
4. Add no knob: the implied link weighs as the direct link it stands for (the segment registry's Other pair).
5. Say how `canonlinks` reads it: which pivot the pattern is (an event result, or a chain result for the person between), what `Reader.For` reads to find it, and how far `Deps` reaches. A pattern over an event is one more row and needs nothing there; a two-hop pattern changes `canonlinks`' pivots, `For` and its `Deps` only. Add a `TestFor` subtest and an oracle row in `TestForMatchesDerive`.

## Add a layer package

Add a row to `layerPackages` in `core/matchregistry/guard_test.go`, with `numbers` true for a layer and false for a pure value or fixture package. Add the package's glob to `.cursor/rules/match-rules.mdc` and its path to the Check command below. Give it a package doc naming its layer and saying it is pure. Test it from an external `_test` package with an exact table and a fixed-seed invariants test, building fixtures with `namevaluetest` and `datevaluetest` (`core/propertycompare` is the model).

## Check

```
go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/segmentcompare/... ./core/kinship/... ./core/shortlist/... ./core/graphalign/... ./core/claimevidence/... ./core/matchregistry/...
```

and `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/` after a subject section change.

No build tags. Then `scripts/check-pr.sh`.

## Do not

- Keep a second copy of a word helper (`NormalizeForm`, `SplitWords`, `SpellingSimilarity`).
- Write a number in a layer package.
- Read the catalog in a layer.
- Import `core/autoreconcile` from a layer package other than `core/subjectprofile`.
- Write a points switch or a lead lookup outside `core/subjectcompare`.
- Weaken a guard or add an allow-list to make it pass. The one list the inventory guard keeps, `catalogData`, names fields that are catalog data, never tunables; each is named under Not knobs.
