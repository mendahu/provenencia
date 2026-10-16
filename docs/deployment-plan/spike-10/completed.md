# Spike 10 — completed

## S10-27 — Name frequency

**Landed 2026-10-16** in [#409](https://github.com/mendahu/provenencia/pull/409). A name word common among the project's filed people earns up to 20% fewer points. A word costs `0.2 × ln(df) / ln(pop)`, with `pop` at least 100 people, so `John Smith` vs `John Smith` scores 8.64 and `John Quirk` vs `John Quirk` 9.60 at weight 10. Similarity is untouched: an identical name is still `agree` and pins in Promote. Promote counts Person handles' names once into `Stats.Words` and drops them on a handle commit; merge hints and Subject suggestions count the candidates they loaded. A candidate that only a common surname supported can now fall under `MinScore`. A bare `NameComparer{}` has no counts and scores as before.

## S10-23 — Sex / gender on the Person page

**Landed 2026-10-16** in [#407](https://github.com/mendahu/provenencia/pull/407). The Person page shows a Badge above the name: the gender when the Person has one value, else the sex at birth worded “Male · at birth”. The header thumb takes the family's tint and line: man, woman, non-binary (lapis), or other for any researcher term and for Intersex. Mixed, unknown, indeterminate, and empty values show no Badge and a plain thumb, and a disagreeing gender does not fall back to sex at birth. VoiceOver reads the Badge as “Gender, Woman” or “Sex at birth, Male”. The header thumb is 120pt on Person, Event, and Place pages. The Details rows are unchanged.

## S10-20 — Open file and box-select

**Landed 2026-10-14** in [#396](https://github.com/mendahu/provenencia/pull/396). The Evidence graph header has a 208pt select of the Source's files and an Open button that sends the chosen one to the system app. The cover is chosen first when it has a file, and a missing file shows a "File not found" callout until you pick another file or arm a tool. A file with no thumbnail shows its type glyph in the menu. A Source with no file shows neither control. Hold Shift and drag on empty canvas, or arm Select on the palette (help: Hold Shift to select), to draw a rectangle. The person, event, and place cards it touches are selected and the relationship cards are not. Dragging one selected card, or pressing an arrow key, moves the whole group. Empty-canvas drag still pans with Select disarmed and Shift up.

## S10-49 — Glyph consolidation

**Landed 2026-10-14** in [#394](https://github.com/mendahu/provenencia/pull/394). One `PVMarkKey(subjectTypeKey:)` rule maps a subject type to its mark, and Go and Swift tests pin it to the registry's `IconSymbol`. Thumbnails, the omnibar lead, and the Properties type cards take one `PVGlyph` instead of a mark and a symbol. An empty Source types list now shows the tag symbol and an empty Metadata list the list symbol, matching their sidebar rows. Every other thumbnail looks as before. Omnibar Source-type and Metadata leads match their sidebar icons, and Properties palette order and bridge labels read the registry presentation.

## S10-19 — Event mark and sidebar icons

**Landed 2026-10-14** in [#393](https://github.com/mendahu/provenencia/pull/393). The Event mark is a calendar, not an hourglass, on cards, omnibar hits, Promote, and Properties. The sidebar rows for Persons, Events, and Places draw the same marks as the cards, at 15pt and 16pt in the collapsed rail, in the row's ink. The Back and Forward history rows show the same marks beside their section, and an empty Persons, Events, or Places list shows its kind's mark. One `PVGlyph` value names a symbol or a mark, and the sidebar item, the history row, and the empty state carry it.

## S10-40 — List scroll position

**Landed 2026-10-14** in [#392](https://github.com/mendahu/provenencia/pull/392). People, Events, Places, and Sources remember the row at the top of the list on that history entry. Back, Forward, a jump, and relaunch return to it. Opening the list from the sidebar starts at the top. Changing Sources sort or type filter returns the list to the top. A row that is gone stays at the top.

## S10-18 — Artifact drop and cards

**Landed 2026-10-14** in [#387](https://github.com/mendahu/provenencia/pull/387). Dropping a file on the Source page content pane opens Add Artifact with that file staged and the label empty. A rejected drop opens the same dialog on the existing callout and stores nothing. Artifacts are wrapping cards: a large preview, the ref, Cover when it applies, and Open or Add file, Use as thumbnail, and Delete. The pencil edits the label and description on the card.

## S10-17 — Name parts split

**Landed 2026-10-13** in [#385](https://github.com/mendahu/provenencia/pull/385). Split into parts sits beside the full form. Leaving that field fills parts only when the list is empty. Split refills, and asks before it replaces a hand edit. A comma makes the text before it one surname, and a later comma is not kept on the following word. Otherwise a compiled title list peels prefixes and suffixes when another word remains: one leftover word is a given name, and the last of two or more is the surname. The part-type column stays the width of Surname prefix. Up and down move a part; the row has no drag grip.

## S10-16 — Observation list rethink

**Landed 2026-10-12** in [#384](https://github.com/mendahu/provenencia/pull/384). The citation composer is two columns: stepped artifact and citation over the file, and one authoring box over subject groups. A bridge is one write and is listed under both ends. Connect still opens that box with the ends filled and an empty role or type.

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

## S10-26 — Phonetic tier

**Descoped 2026-10-12.** Daitch–Mokotoff under near-spelling is [`ideas/daitch-mokotoff.md`](../../ideas/daitch-mokotoff.md). Lee is `800000` and Leigh is `850000`, so they do not meet. Smyth/Smith stays on edit distance.

## S10-25 — Abbreviations and nicknames

**Landed 2026-10-12** in [#379](https://github.com/mendahu/provenencia/pull/379). An abbreviation or nickname table hit scores 0.98. `Jas.` vs `James` and `Jim` vs `James` are that score. `José` vs `Jose` stays 0.95. Every installed table is consulted. Origin records whether the app, a user, or a plugin contributed it. Product tables apply to given names, nicknames, and untyped words. New catalogs install the product lists. Open does not heal them. There is no editor UI.

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

## S10-21 — Handle last-activity scope

**Landed 2026-10-14** in [#398](https://github.com/mendahu/provenencia/pull/398). Every Person, Event, and Place handle has a last-activity revision (`canonical_entity` audit scope). It moves on the handle's own row, its claims and evidence pins, and Observations (with their date and name values and notes) on a member Subject; the Source's revision still moves too, and a promote still moves no Source. The People, Events, and Places list reads return `updated_revision` and `updated_at` on each row, and a migration backfills existing history. The row cell is S10-22.

## S10-22 — Last updated on People, Events, and Places

**Landed 2026-10-15** in [#402](https://github.com/mendahu/provenencia/pull/402). Each row on People, Events, and Places shows when the researcher last worked on it, in its own 136pt column before the ref: "Today, 3:15 PM", "Yesterday", or a short date, with the full stamp as hover help and a faint "—" when nothing is recorded. The spoken row label ends with "updated …" or "last updated unknown". Every write reports what it moved and when: the handles and Sources whose last activity moved (`handle_scopes`, `source_scopes`), which reads it changed (`sources` for the Source graph, `citation_sources`, `property_usage`), and its own revision and time. Source notes, metadata, artifacts, credibility, and Subject writes now return these effects too, and the blanket triggers that reloaded every list after a Source-page edit are gone. The app patches the moved rows in place (People, Events, Places, and Sources) through one declared patch in the session cache, reloads only the reads a write changed, and marks hidden lists stale instead of reloading them; the composer reloads the Citation list it shows. The engine keeps each handle's and Source's last activity in its display cache, so no list load scans the audit history. Every catalog timestamp is formatted by one function, `TimestampFormat.string`, with named styles; the row composes its cell from it. `PVList` gained an optional detail column; S10-41 reuses the cell composition for the Sources row.

## S10-41 — Sources list: last updated, status filter, search, saved settings

**Landed 2026-10-15** in [#405](https://github.com/mendahu/provenencia/pull/405). Each Sources row shows when the researcher last worked on it, in the same 136pt column People, Events, and Places use, and speaks it. A search field finds a Source by part of its title or ref (case- and accent-insensitive), and a status select narrows the list to Empty, In progress, or Fully promoted graphs. The status comes from one grouped read beside the Evidence-graph counts that also returns each Source's primary and filed subjects; filing a person on the Promote page refreshes that Source's count. When nothing matches, the empty state offers Clear filters, and the count line reads "N of M sources". Sort, type, status, and search are saved per project and install user and come back after Back or relaunch (S10-09). The last-updated cell composition moved out of the Conclusion lists into a shared `PVLastUpdated` recipe and a `PVListDetailCell` view that `PVList` and the Sources row both draw.
