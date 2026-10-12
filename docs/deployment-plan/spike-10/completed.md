# Spike 10 — completed

## S10-01 — Button hit boxes

**Landed 2026-10-11** in [#361](https://github.com/mendahu/provenencia/pull/361). Toolbar place, person, event, and connect buttons take a click on their padding. The bridge citation hit sits on the painted pencil, not one button-width to the left.

## S10-02 — Zoomed-out grid and edges

**Landed 2026-10-11** in [#362](https://github.com/mendahu/provenencia/pull/362). The background grid is one repeated cell. It fades as a whole when a cell would be under about a screen point. Connector strokes stay one screen point wide at every zoom. A grown document uses the same tile.

## S10-03 — Canvas grows

**Landed 2026-10-11** in [#363](https://github.com/mendahu/provenencia/pull/363). The document stays 4000×4000 while every card has a card-width of margin. A card inside that margin extends that edge by 800 points, and `maxCell` follows. Dragging and a new composer card use the grown size. The grid tile does not change. A small graph still opens on 4000×4000.

## S10-04 — Camera survives leaving the graph

**Landed 2026-10-11** in [#364](https://github.com/mendahu/provenencia/pull/364). The workspace session stores magnification and content offset per Source. The composer, Back, and Forward restore it. Growing the document does not reset it. Quit drops it. A Source with no stored camera opens centered at 1×.

## S10-05 — Text area scrollbar

**Landed 2026-10-12** in [#368](https://github.com/mendahu/provenencia/pull/368). `PVTextArea` inserts a newline at the caret when Return is pressed. Overflow shows a scrollbar that tracks the caret. A short note does not show an empty track.

## S10-06 — Omnibar whole-word rank

**Descoped 2026-10-11** in [#368](https://github.com/mendahu/provenencia/pull/368). A title that contains `marriage` already ranks above obituaries for Marion Margaret. No change to `core/search`.

## S10-07 — Oldest user on an existing project

**Landed 2026-10-12** in [#369](https://github.com/mendahu/provenencia/pull/369). Opening a project you do not belong to selects the user with the smallest UUIDv7 id. If you are already a contributor, you stay selected. An empty user list still offers create-new. The list stays alphabetical.

## S10-08 — Enter saves a Property title

**Landed 2026-10-12** in [#370](https://github.com/mendahu/provenencia/pull/370). Enter in a single-line inline edit saves when Save is enabled. When Save is disabled, Enter leaves the edit the same way Cancel does. Escape cancels. A multiline field still takes Enter as a newline.

## S10-10 — Observation defaults

**Descoped 2026-10-12** in [#370](https://github.com/mendahu/provenencia/pull/370). A repeated cluster is [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md): place a wedding, birth, or residence in one action, or paste a cluster already on the canvas. No default role on a generic new connection.

## S10-11 — Composer performance

**Done** in [#356](https://github.com/mendahu/provenencia/pull/356). A combo box builds its option list only while the menu is open, so a form of closed pickers does not lay out every choice. Checked 2026-10-12: a citation with 70 Observations opens quickly. No further profile pass, and the form still appears after the file viewer loads.

## S10-12 — Outside click closes a sheet

**Landed 2026-10-12** in [#371](https://github.com/mendahu/provenencia/pull/371). A click outside a form sheet closes it. A sheet with edits asks to discard or keep editing. A clean or read-only sheet closes. Delete confirms are unchanged.

## S10-13 — Date modal focus and transcription prefill

**Landed 2026-10-12** in [#372](https://github.com/mendahu/provenencia/pull/372). The structured date modal opens with focus on Year. When the draft is empty and the transcription contains a date, the first match fills year, month, and day. A date already on the Observation stays. `14 May 1817` opens as 1817, May, 14.

## S10-14 — Adaptive choice field

**Descoped 2026-10-12** in [#370](https://github.com/mendahu/provenencia/pull/370). A kit field that swaps radio, select, and combo box by option count is not a priority for this spike. S10-D1 is not drawn. Term fields keep the control they have.

## S10-15 — Create Source field order

**Landed 2026-10-12** in [#375](https://github.com/mendahu/provenencia/pull/375). Create Source asks for the name, then the type, then the description. Focus and tab follow that order. Create still requires a name and a type.

## S10-24 — Accent fold and spaced particles

**Landed 2026-10-12** in [#378](https://github.com/mendahu/provenencia/pull/378). A fold-only word match scores 0.95. `José` vs `Jose` is that score; `José` vs `José` stays 1. Display, `NormalizeForm`, and `sort_key` do not fold, so the Person page still keeps both spellings. Spaced `o`, `mc`, `mac`, `st`, `d`, and `fitz` rejoin with the next word. `O Brien` matches `O'Brien`. `O.` does not.

## S10-29 — Name patterns from the catalog

**Landed 2026-10-12** in [#377](https://github.com/mendahu/provenencia/pull/377). New catalogs install a Western name-format profile: part order, roles, weights, and the project default. Matching and display read those rows. A catalog with no default still compares as the built-in Western pattern. Sorted order of Vincent van Gogh is `van Gogh, Vincent`. Spanish dual surnames and a patronymic part stay unscheduled.

## S10-42 — Long titles in a select

**Landed 2026-10-12** in [#373](https://github.com/mendahu/provenencia/pull/373). A select's closed label ellipsizes inside the width the caller gave it. The chevron stays put. When the label is cut, help shows the full string. A short label is unchanged.

## S10-43 — Combo box prefix rank

**Landed 2026-10-12** in [#374](https://github.com/mendahu/provenencia/pull/374). A combo box keeps every contains match and orders it in three bands: the label starts with the query, a later word starts with it, then any other contains match. Order inside a band stays the caller's order. `par` lists parent above grandparent.

## S10-46 — Simplified cards when zoomed out

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Drawing the card in AppKit made a zoomed-out graph fast enough. The full card stays at every zoom. S10-D14 is not drawn.

## S10-47 — Mount cards near the window

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). The same draw builds a view for every card on the Source. Cards outside the window stay mounted.

## S10-48 — Canvas cover

**Descoped 2026-10-11** in [#367](https://github.com/mendahu/provenencia/pull/367). Subject and bridge cards are drawn in AppKit, and the first paint is fast enough that a loading cover is not needed. No `ProgressView` over the canvas.
