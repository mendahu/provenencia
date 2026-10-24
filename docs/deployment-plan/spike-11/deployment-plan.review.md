# Review — Spike 11 PR map

**Verdict:** Fits with changes. **Rounds:** 1. **Reviewers:** maintainability (fresh agent, map questions).

## To user

- **Two preparatory PRs added: S11-22 (pure value types) and S11-23 (one home for primary kinds).** Recommendation: keep both. Without S11-22, every layer would either import `core/database` or define its own date and name structs. Without S11-23, the registry would add the fifth to seventh hand-written kind lists.
- **Two splits added: S11-24 (delete frequency weighting, out of S11-13) and S11-25 (claim evidence RPC and pin policy, out of S11-19).** Recommendation: keep both. S11-13 was well over 10 steps, and S11-19 would have broken the shipped sheet between PRs.
- **A matching skill (`add-match-rule`)** is added in S11-02 and grown by S11-10, S11-14 and S11-16. Recommendation: keep it. The name is a working name.

## Writer's own calls

- New PR IDs are S11-22 – S11-25, numbered in the order they were cut. Their place in the order comes from the slices and dependencies, not the number.
- The pin-table migration is conditional: S11-19's plan checks the tables, and cuts a migration PR ahead of S11-19 only if they can't hold one incoming Observation against many member Observations.

## After the review (2026-10-17)

The user accepted the preparatory PRs and splits. An ideas-folder pass then added slice 6 (connection subrows, S11-26 and S11-27, sharing brief S11-D1) and moved Spike 10's S10-36 (likeness crops) into S11-20 / S11-D2. These were added after the reviewer ran and were not reviewed.

## Round 1

### K-1 · blocker · Pure value types need a preparatory move, not layer-1 copies
`datevalues.Value`, `namevalues.Value` and `properties`' value-type and cardinality constants live in packages that import `core/database`. About 65 files use the structs.
**Answer:** Fixed. S11-22 is added ahead of S11-02, with the skills updated. Working rule "Pure value types and kinds come first".

### K-2 · major · Alias seeds and name-pattern types have no home once `core/match` goes
**Answer:** Fixed in S11-03. The types move to `valuecompare`. The seeds, `Register` and `third_party/nicknames/` move to `namealiases`, with `add-seeded-vocabulary` and `seeded-vocabulary.md` updated.

### K-3 · major · `core/database/matching` is not dead; S11-13 too big
`promotealign/layer.go:48` calls `matching.SubjectsValues`.
**Answer:** Fixed. The working rule now says `SubjectsValues` moves to the loader in S11-13. Frequency removal is split out as S11-24.

### K-4 · major · A registry is due for primary kinds and property keys
**Answer:** Fixed. S11-23 is added. S11-10 ports `TestDefaultProfileCardinality` as "every registry `PropertyKey` is a seeded Property".

### K-5 · major · One guard, not one per layer package
**Answer:** Fixed. Working rule "One guard, not one per package". S11-02 ships it, modeled on `core/writes/begin_test.go`.

### K-6 · major · S11-19 mixed engine, FFI and the Swift pin flow
**Answer:** Fixed. The work is split into S11-19 (Go query) and S11-25 (RPC, Go default pins at Apply, Swift drafting moved, proto field 9 deleted). Working rule "The pin policy has one owner, in Go". The migration is conditional (see Writer's own calls).

### K-7 · major · Extend `RowFlags` / `Reason`; one wire vocabulary
**Answer:** Fixed in S11-16 and S11-17. Working rule "One wire vocabulary".

### K-8 · major · No skill teaches the new pattern; docs go stale
**Answer:** Fixed. Working rule "One matching skill, grown with the layers". Each migration PR names the docs and idea files it makes false (S11-05, S11-06, S11-16).

### K-9 · minor · S11-08: the Swift helper already exists
**Answer:** Fixed. `PVComboBoxMatch.normalize` is promoted, with a locale parameter. `DateValueTranscriptionMatch` is added, and the `shared-list-helpers.md` section is closed.

### K-10 · minor · Reuse the graphcache stats slot; name who deletes fan-out
**Answer:** Fixed. S11-14 deletes fan-out. S11-15 reuses the stats slot.

### K-11 · minor · S11-21: extend the existing benchmark; make the budget run
**Answer:** Fixed. S11-21 is a budgeted `Test…` on `bench_test.go`'s fixture. S11-24 deletes `BenchmarkLoadStatsNames`.

## Map review — S11-34 – S11-37 (2026-10-24)

**Verdict:** Fits with changes. **Rounds:** 1. **Reviewers:** maintainability (fresh agent, map questions), over S11-34 – S11-36, the "Find another record" rules, bar 10 – 11 and [`promote-find-another-record.md`](promote-find-another-record.md). The engine claims held: `checkFixed`, `loadCanon`, `anchor`, Apply and `ResolvePins` take any handle of the right kind, so S11-35 needs no Go code.

### To user

- **One preparatory PR added: S11-37 (one kind-scoped catalog search).** Recommendation: keep it. Without it S11-36 rebuilds the omnibar's results model and copies its private hit row inside Promote.
- **The menu entry is a sentinel option, decided in the map** (K-6). A kit action item on `PVSelect` is cut as its own PR only if S11-D3 draws what a sentinel cannot do.

### Round 1

#### K-1 · major · A second search state and hit row would grow inside Promote
**Answer:** Fixed. S11-37 moves `OmnibarResultsModel`'s search state and `OmnibarHitRowView` to a shared home with `kinds` and an optional location; S11-36 depends on it. Working rule "No new search RPC, and one search model".

#### K-2 · major · A found record's reason is never refreshed
**Answer:** Fixed. S11-35's call writes reason `decided`; the band is refreshed from the proposal. How a held record with band None reads is S11-D3's. Feature doc "Where the Swift side changes".

#### K-3 · minor · `ConflictWithFixed` cannot fire for a found record
**Answer:** Fixed. Dropped from the engine table and S11-35's tests; only Contested applies.

#### K-4 · minor · S11-34 names no existing hover style
**Answer:** Fixed. The scope names `PVMenuRowButtonStyle`, `PVListRowStyle` and `SourcesSplitZoneButtonStyle`, reuse or extend one, and says to check the board wants `PVHoverEffect`'s press scale.

#### K-5 · minor · A sentinel option passed to `setTarget` marks the row decided
**Answer:** Fixed. Working rule "The menu entry is a sentinel option"; S11-36's Check adds "opening and cancelling the search leaves the row undecided".

#### K-6 · minor · S11-36 too big if the kit route is chosen
**Answer:** Fixed. Sentinel decided in the map; the search refactor is S11-37.

#### K-7 · minor · S11-35's Apply test is in the wrong package
**Answer:** Fixed. Filing is checked in the handler or `promote` tests; the propose tests extend `promote_align_test.go` and `promotealign_test.go`.

#### K-8 · minor · No skill update for a second `SearchCatalog` consumer
**Answer:** Fixed. S11-37 updates `add-searchable-kind`.

## Map review — slice 9, kinship shapes (2026-10-24)

**Verdict:** Fits with changes. **Rounds:** 1. **Reviewers:** maintainability (fresh agent, map questions), over S11-38 – S11-40, the "Kinship shapes" rules, bar 12 and [`kinship-shapes.md`](kinship-shapes.md).

### To user

- **One preparatory PR added: S11-41 (one link reading for bridges).** Recommendation: keep it. `core/kinship` needs the key a direct link would have, and today only `promotealign`'s private `linkOf` builds it.
- **S11-40 split into S11-40 (claim evidence engine and wire) and S11-42 (the sheet's sentences).** Recommendation: keep it. The line needs a new field across the FFI, so it is not copy alone.

### Round 1

#### K-1 · major · The evidence query never reads `Canon.Edges`
**Answer:** Fixed. `canonLinks` is the third caller of `core/kinship`, in S11-40; working rule updated.

#### K-2 · major · A bundle is one bridge to one neighbor member
**Answer:** Fixed in scope. S11-40 carries the event on the implied link and its plan decides the bundle's shape and pin id.

#### K-3 · major · S11-40 is a wire change
**Answer:** Fixed. S11-40 lists the proto field, handler, store mappings and vocabulary; the Swift copy is S11-42.

#### K-4 · major · `linkOf` and `eventEndpoint` are private to `promotealign`
**Answer:** Fixed. S11-41 moves them to `segmentcompare`; `core/kinship` takes `Terms` and keys through `Key`.

#### K-5 · major · A second merge beside the loaders'
**Answer:** Fixed. `core/kinship` returns only implied links; the loaders' existing merge (`edgeSeenKey`) counts a fact once.

#### K-6 · minor · Canon reach already covers grandparent
**Answer:** Fixed. `loadCanon` walks the whole component; S11-39 times the derivation, not extra reach.

#### K-7 · minor · "via" can name a neighbor with no subrow
**Answer:** Fixed. S11-39 pins what the via reason says through an implied link, and notes implied bridges are never filed or shown.

#### K-8 · minor · The terms are already seeded
**Answer:** Fixed. S11-38 adds `vocabkey` constants only, checked by `TestVocabKeysMatchSeed`.

#### K-9 · minor · `add-match-rule` must learn `core/kinship`
**Answer:** Fixed in S11-38.

#### K-10 · minor · Bar 12 overclaimed "strong"
**Answer:** Fixed. Bar 12 says "at the band a spouse relationship would get".

#### K-11 · minor · Event type passed twice
**Answer:** Fixed. It is `BridgeKey.EventType`; dropped from the input.
