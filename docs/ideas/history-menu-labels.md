# History menu labels

**Status:** idea only — not roadmapped. Moved out of `shared-list-helpers.md` on 2026-11-17 (that file is retired; its other items went to Spike 16 Batch 6) so it can be thought through on its own; more changes to the menu are to be added here. Spike 16's S16-53 moved the menu onto the kit menu item (with a feature-private label) and `.pvPopup`, and left this out.

## Where it started

Dogfood 2026-10-29:

> "The drop-down menu for the navigation bar's back and forward needs to be wider because they're truncating the last part of the breadcrumb and so all the entries look the same."

## What is already fixed

- **Width.** S14-35 widened the panel to 600 pt (`HistoryJumpMenuHost.panelWidth`), narrowed only to stay inside the content column (`PVPopupListPlacement.fittedWidth`).
- **The open section is named.** A record page's row reads "Ada Robins · Timeline": the row builder joins the leaf from `WorkspaceLocationLabels.labels(for:names:)` to its section with `L10n.Conclusions.pageSectionLeaf` (`WorkspaceToolbar.swift`, `HistoryJumpMenuHost.rows(_:names:)`).

## What is still wrong (checked in the code, 2026-11-17)

- **The end of a label is what gets cut.** The leaf and its " · Timeline" suffix are one `Text` with `.truncationMode(.tail)` in `HistoryJumpRowLabel`. A long name loses its section first, so walking one long-named person's sections shows rows that read the same.
- **Graph rows do not name their Source.** For an Evidence graph, its composer or its Promote page, `labels(for:names:)` returns the graph's name as the leaf and the surface as the section. The root is "Sources", and the Source title is never in the row. Two Sources whose graphs are both "Default" give two rows reading "Sources › Default". The toolbar's breadcrumb (`WorkspaceToolbar.breadcrumbItems`) has the Source title; the menu does not use it.
- **The ref is not a tie-breaker between sections.** It is the location's carried ref, so a person's sections share one.

## Directions

- **Keep the section whole.** Draw the section as its own text that never truncates, and truncate the name in the middle (or at the head), so the part that differs survives. Apply the same to the mono ref when it is the unique part.
- **Name the Source on graph rows.** Build graph, composer and Promote rows from the breadcrumb (Source title › graph name › surface), or have `labels(for:names:)` return the Source title for them. Either way the toolbar's crumbs and the menu keep one label rule (`WorkspaceLocationLabels` is the "one home for all three").
- **Leaf first.** Put the record's name before the sidebar section, and drop the section prefix when the leaf already carries it.
- **Two-line rows.** Let long composer and entity labels wrap to two lines, trading height for scan speed. Only if the first two are not enough.

Any of these changes how rows look, so it needs a Claude Design brief. It is easiest after S16-53, which gives the menu a kit row to change.

## Open questions

- Which of the directions above, and whether together.
- Whether the toolbar's breadcrumb and the menu should read exactly the same, or the menu keeps its shorter form.
