# Keyboard reorder

**Status:** parked 2026-11-12 from the S15-09 review (UX-4, a nit; the user chose an idea file). Nothing on the map yet.

## Problem

A metadata board's saved rows reorder only by dragging the handle. A researcher on the keyboard, or using VoiceOver, cannot change their order at all. That holds on every board: the Source's Metadata section and, since S15-09, each Artifact's Copy details, up to 60 boards on one Source page. S15-10 and S15-11 reuse the same board, so they inherit the gap.

## Rough shape

- **Fix it in the kit, not the board.** `PVReorderableList` (`macos/App/DesignSystem/Components/ReorderableList/`) is the one reorderable list in the app, and its only caller today is `MetadataBoardView`. It supplies `List` + `onMove` and nothing else. Give it keyboard and accessibility moves, and every board gets them.
- **One move path.** `MetadataBoardModel.moveSaved(from:to:)` already takes an `IndexSet` and a destination and handles the optimistic move, the write and the revert. Keyboard and VoiceOver moves call the same `onMove`, so the model does not change.
- **VoiceOver.** Each row gets the accessibility actions "Move up" and "Move down", hidden at the ends, and they announce the new position ("Provider, 1 of 3").
- **Keyboard.** A focused row moves with a modifier and the arrow keys (⌥⌘↑ / ⌥⌘↓ is the macOS convention in Finder and Mail lists). Check that the shortcut does not clash with the workspace's Back/Forward or the omnibar.

## Open questions

- Which shortcut, and whether the row's handle takes focus or the whole row does.
- How it shows on screen: the moved row keeps focus, and a short highlight confirms the move.
- Whether one write per keypress is acceptable, or held arrows should batch into one reorder write.

## Check

With VoiceOver on, "Move down" on a board's first saved row makes it second, and the order survives a reload. With the keyboard alone, a focused row moves up and down, and focus stays on it.
