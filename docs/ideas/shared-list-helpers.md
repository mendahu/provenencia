# Shared list helpers

**Status:** idea only — not roadmapped. Small duplication found while planning S10-41 (the Sources list filters). Each item is a cleanup, not a feature. "In-list text matching" (closed by S11-08: `TextMatch`) and "Primary Subject kinds" (S11-23) moved into Spike 11.

## Go reads every Swift mirror

`FieldSlug.fixtures` and `ProjectSlug.folderNameFixtures` are kept in sync with Go by hand. `swiftFixturePairs` in `api/ffi/swift_vocabulary_test.go` can check them as it checks `TextMatch`; `FieldSlug` has a `\u{2019}` fixture, and `ProjectSlug` has `"Robins\\Family"` / `"Robins\"Family"` rows, which are escapes the reader refuses and would need to decode.

## Fake store keys

`FakeStore`'s property and term creates build keys inline, near-copies of `FieldSlug.kebab`.

## Properties empty search state

Properties draws its own "no results" state instead of `PVEmptyState`.

## One catalog querier

S12-01 adds `database.Querier` (`Query` + `QueryRow`, in `core/database/tx.go`) for reads that run over a `*sql.DB` or a write's `*sql.Tx`. Nine packages still declare their own copy: `sqlbatch`, `conclusiondetails`, `graphcache`, `matching`, `namealiases`, `autoreconciler` (adds `Exec`) and `effects` (`querier`), plus inline interfaces in `datevalues`, `namevalues` and `canonicalentities`. Move each onto `database.Querier` (or a `database` type that adds `Exec`) when its package is next touched. `graphcache` is blocked: `core/database` imports it, so it needs a leaf package for the interface first.

## Query-wait test helper

S12-02 added `macos/ProvenenciaTests/QueryHandleWaiting.swift` (`waitForQuery`, which also waits out a refetch). `MetadataModelTests`, `PropertiesModelTests` and `SourceTypesModelTests` still keep a private `waitForQuery`; move them onto it.

## Trimmed empty-string binder

`datevalues.nullIfEmpty` and `audit.nullIfEmpty` are identical: each trims whitespace and binds an empty result as NULL. `database.NullString` does not trim, so they stay local. Add `database.NullTrimmed` and move both onto it when a third caller needs the trim.

## One popup dismissal

Sheets share one dismissal now (`.pvSheet`: Escape, a click outside, the discard question; S11-33). Menus and popups each dismiss their own way: `pvContextMenu`, `HistoryJumpMenuHost`, `OmnibarResultsHost` and `PVComboBox` each carry their own Escape and outside-click handling. One kit presenter for a popup, as `.pvSheet` is for a sheet, would give them the same rules; move each onto it when its host is next touched.
