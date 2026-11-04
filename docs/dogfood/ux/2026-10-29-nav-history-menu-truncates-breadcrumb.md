# Back/Forward history menu truncates breadcrumb tails

- **Date:** 2026-10-29
- **Type:** Bug (layout)
- **Where:** Workspace toolbar / Back and Forward navigation history dropdown (long-press or right-click on ⌘[ / ⌘])
- **Annoyance (verbatim):** "The drop-down menu for the navigation bar's back and forward needs to be wider because they're truncating the last part of the breadcrumb and so all the entries look the same."
- **Consequence:** Jumping to a specific place in the Back or Forward stack means guessing from identical-looking rows (same section glyph and a truncated prefix) instead of reading the distinguishing tail of the title, ref, or deep-place label.
- **Wanted:** Each history row stays readable at a glance — especially the part that differs between entries — so Back/Forward jump is trustworthy in long sessions with many Sources, graphs, composers, and entity pages.

## How it works today (observed, not fixed)

- **Where the menu is built.** `macos/App/Features/Workspace/WorkspaceToolbar.swift` — `HistoryJumpMenuHost` gathers `navigation.backJumpItems()` / `forwardJumpItems()`, maps them to `HistoryJumpRow`, and renders `HistoryJumpMenuPanel` anchored under the Back or Forward control (`HistoryNavControl` long-press / `PVRightClickCatcher`). History data comes from `WorkspaceNavigation` → `NavigationHistoryStore.backJumpItems` / `forwardJumpItems` (`macos/App/Features/Workspace/WorkspaceNavigation.swift`, `NavigationHistoryStore.swift`); each row is a `(index, WorkspaceLocation)`.
- **What each row shows (not the full toolbar breadcrumb).** `HistoryJumpRow` (`WorkspaceToolbar.swift`) labels a row as `{section label} › {leaf}` plus an optional mono `ref` on the right. `rootLabel` is `L10n.string(location.section.label)`. `leafLabel` is `location.title` or `location.ref` only — it does **not** walk the same multi-segment path as `WorkspaceToolbar.breadcrumbItems(for:names:goTo:)` (Source title, graph name, composer/promote leaf, entity title, etc.). Deep places therefore collapse to "Sources › …" with a single leaf string, so many stacks look alike even before truncation.
- **Fixed width.** `HistoryJumpMenuPanel` wraps rows in `PVContextMenuPanel(width: 360, …)` (`WorkspaceToolbar.swift` ~L551–552). Default panel width elsewhere is 180 (`macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` `PVContextMenuPanel.width`).
- **Truncation.** Inside each row, `rootLabel` and `leafLabel` are `Text(…).lineLimit(1)`. The leaf uses `.truncationMode(.tail)` and `.layoutPriority(1)` so it wins horizontal space over the muted section prefix but still ellipsizes at the **end** when the 360pt panel (minus glyph, chevron, padding, and optional `refLabel` with `.fixedSize(horizontal: true, vertical: false)`) runs out of room — matching Jake's "last part of the breadcrumb" complaint. There is no `.minimumScaleFactor`, wrapping, or head/middle truncation.
- **Contrast with the toolbar.** The live breadcrumb strip uses the same file's `breadcrumbItems` + `PVBreadcrumbs` with `.lineLimit(1)` on the whole trail (`WorkspaceToolbar.swift` ~L153–167); the jump menu is a separate, narrower presentation of `WorkspaceLocation`, not a reuse of those crumb strings.

## Fix directions (options, not chosen)

1. **Wider or minimum-width panel** — raise `PVContextMenuPanel(width:)` (or add `minWidth` / intrinsic sizing) in `HistoryJumpMenuPanel` so typical deep-place titles fit; watch overlap with the omnibar on narrow windows.
2. **Truncation mode** — keep width but show the distinguishing suffix: `.truncationMode(.head)` or middle truncation on `leafLabel` (and/or the mono `refLabel` when it is the unique part).
3. **Richer or reordered label** — build row titles from `WorkspaceToolbar.breadcrumbItems` (or only the last one or two segments / leaf-first) so graph and Source names appear before truncation; optionally drop the redundant section prefix when the leaf already embeds context.
4. **Multi-line rows** — allow `lineLimit(2…)` or wrapping inside the panel for long composer / entity labels (trade vertical space against scan speed).
