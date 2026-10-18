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
