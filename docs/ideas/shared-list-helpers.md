# Shared list helpers

**Status:** idea only — not roadmapped. "Go reads every Swift mirror" and "Fake store keys" closed by Spike 12 S12-35 (2026-10-28). Small duplication found while planning S10-41 (the Sources list filters). Each item is a cleanup, not a feature. "In-list text matching" (closed by S11-08: `TextMatch`) and "Primary Subject kinds" (S11-23) moved into Spike 11.

## Properties empty search state

Properties draws its own "no results" state instead of `PVEmptyState`.

## One catalog querier

S12-01 adds `database.Querier` (`Query` + `QueryRow`, in `core/database/tx.go`) for reads that run over a `*sql.DB` or a write's `*sql.Tx`. Nine packages still declare their own copy: `sqlbatch`, `conclusiondetails`, `graphcache`, `matching`, `namealiases`, `autoreconciler` (adds `Exec`) and `effects` (`querier`), plus inline interfaces in `datevalues`, `namevalues` and `canonicalentities`. Move each onto `database.Querier` (or a `database` type that adds `Exec`) when its package is next touched. `graphcache` and `effects` are blocked: `core/database` imports them, so they need a leaf package for the interface first.

## Trimmed empty-string binder

`datevalues.nullIfEmpty` and `audit.nullIfEmpty` are identical: each trims whitespace and binds an empty result as NULL. `database.NullString` does not trim, so they stay local. Add `database.NullTrimmed` and move both onto it when a third caller needs the trim.

## One popup dismissal

Sheets share one dismissal now (`.pvSheet`: Escape, a click outside, the discard question; S11-33). Menus and popups each dismiss their own way: `pvContextMenu`, `HistoryJumpMenuHost`, `OmnibarResultsHost` and `PVComboBox` each carry their own Escape and outside-click handling. One kit presenter for a popup, as `.pvSheet` is for a sheet, would give them the same rules; move each onto it when its host is next touched.

## One popup list under the combo box and the omnibar

`PVComboBox` and the omnibar results (`OmnibarResultsHost`) are the same popup list built twice. Both have a field that drives a popup of rows, ↑/↓ that stop at the ends, Return on the highlighted row, a highlight shared by the pointer and the keyboard, closing on Escape or a click outside, and the card behind the rows. Each draws them its own way:
- **Keys.** The omnibar watches raw key codes through an `NSEvent` monitor. The combo box uses `onKeyPress` and also has Home/End, ⌘↑/⌘↓ and Page Up/Down.
- **Highlight movement.** The omnibar moves through `PVFloatingMenuSelection.moveIndex`. The combo box moves through `PVComboBoxMatch.move`, which also skips disabled rows.
- **Popup.** The combo box opens its own window, sized to the field and placed against the screen edges by `PVComboBoxPlacement`. The omnibar draws a 640 pt overlay in the content column.
- **Card.** The two cards use different border and corner tokens.
- **Scrolling** to the highlighted row is already shared (`.pvScrollsRowIntoView`).

The fields are not the same, so the omnibar cannot simply become a combo box. The combo box picks a value and keeps it: `@Binding selection`, the field shows the picked label, and only typed text filters. Its Escape has two steps, and a click outside also takes focus off the field. The omnibar has no picked value: Return opens the record and clears the field, Escape keeps the query, and it has an error state of its own. The shared piece is the layer underneath: a kit popup list that owns the window and its placement, the keys, the highlight, the scrolling, closing, and the loading, empty and error lines. `PVComboBox` would add the picked-value field on top, and the omnibar the command field. It folds in the popup dismissal above.

`CatalogSearchModel.selectedIndex`, `moveSelection` and `selectedHit` are the omnibar's highlight; they move onto `OmnibarResultsModel` with that list, since the combo box keeps its own.

## One tiny-PNG test writer

`writeTinyPNG` is copied in `core/database/sourcecopy/sourcecopy_test.go` and `core/database/sources/sources_test.go`; one shared catalog test fixture would serve both (S12-01 ship log).

## The history menu on the shared rows

The toolbar's Back / Forward history menu builds its own rows and presenter (`HistoryJumpMenuModel`, `PVContextMenuPanel`) instead of `PVActionMenuRows` (S12-17 ship log). Pairs with "One popup dismissal" above.
