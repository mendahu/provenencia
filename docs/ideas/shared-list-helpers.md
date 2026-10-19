# Shared list helpers

**Status:** idea only — not roadmapped. Small duplication found while planning S10-41 (the Sources list filters). Each item is a cleanup, not a feature. "In-list text matching" (closed by S11-08: `TextMatch`) and "Primary Subject kinds" (S11-23) moved into Spike 11.

## Source graph mutations

`WorkspaceSession.refreshSourceGraphProgress(for:)` switches over the graph mutations by hand; `CatalogQueryRegistry` lists the same ones for `.sourceGraph`. After S10-41 the lists differ only by `.deletedSource`, which would also be correct to patch (`sourcegraphprogress.Get` returns zeros for a gone Source, and `isZero` removes the entry). A `CatalogMutation.graphSourceId` read by both would keep them in step.

## Go reads every Swift mirror

`FieldSlug.fixtures` and `ProjectSlug.folderNameFixtures` are kept in sync with Go by hand. `swiftFixturePairs` in `api/ffi/swift_vocabulary_test.go` can check them as it checks `TextMatch`; `FieldSlug` has a `\u{2019}` fixture, and `ProjectSlug` has `"Robins\\Family"` / `"Robins\"Family"` rows, which are escapes the reader refuses and would need to decode.

## Fake store keys

`FakeStore`'s property and term creates build keys inline, near-copies of `FieldSlug.kebab`.

## Properties empty search state

Properties draws its own "no results" state instead of `PVEmptyState`.

## One catalog querier

S12-01 adds `database.Querier` (`Query` + `QueryRow`, in `core/database/tx.go`) for reads that run over a `*sql.DB` or a write's `*sql.Tx`. Nine packages still declare their own copy: `sqlbatch`, `sourceeventtitles`, `conclusiondetails`, `graphcache`, `matching`, `namealiases`, `promotealign`, `autoreconciler` (adds `Exec`) and `effects` (`querier`), plus inline interfaces in `datevalues`, `namevalues` and `canonicalentities`. Move each onto `database.Querier` (or a `database` type that adds `Exec`) when its package is next touched.

## Query-wait test helper

S12-02 added `macos/ProvenenciaTests/QueryHandleWaiting.swift` (`waitForQuery`, which also waits out a refetch). `MetadataModelTests`, `PropertiesModelTests` and `SourceTypesModelTests` still keep a private `waitForQuery`; move them onto it.

## One ref mint loop

Seven table packages mint a ref and insert with the same loop: a package `maxRefRetries = 8`, `ref.Mint`, and a retry on `database.IsUniqueConflict` (`sources`, `artifacts`, `citations`, `observations`, `subjects`, `canonicalentities`, and `evidencegraphs` from S12-03). Eight copy `uuidJSON` (the same seven plus `identityclaims` and `deleteimpact`). A `ref`-side helper that takes the prefix and an insert func, and a `rowchange.UUID(id)`, would give a ref rule one home. Raised by the S12-03 plan review (M-5).

## Catalog-wide list flags

`effects.Effect` and `effects.Set` carry one bool per catalog-wide list: `Vocabulary`, `PropertyUsage`, and `EvidenceGraphs` from S12-03. Each is a Set field, an Effect field, a line in `Resolve`, a check in `validate`, a `WriteEffects` field and a Swift invalidation. One set of list keys on `Effect` and `Set`, with one wire field and one `CatalogQueryRegistry` mapping, would make the next list a registry entry. Raised by the S12-03 plan review (M-6).
