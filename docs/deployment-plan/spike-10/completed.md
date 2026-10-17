# Spike 10 — completed

## S10-34 — Composer depiction row

**Landed 2026-10-17** in [#442](https://github.com/mendahu/provenencia/pull/442). A researcher can now choose Likeness on a person in the citation composer, on a photo or a PDF page (not on audio). Draw on the file steps the dialog aside and uses the viewer's region tools, with a hint band and Cancel above the canvas; the Citation's own region is untouched. The dialog shows a 96pt preview of the cropped face and a caption, and Save sends it as a temporary PNG that is deleted after the call. After a save the dialog stays open on Likeness with the person pick focused, for the next face. Saved lines show a 32pt thumbnail with a VoiceOver label ("Likeness, Ben Ford"). Edit can redraw, and sends an image only when it did; a hidden region asks for a redraw. Duplicate copies the region, not the saved image. On a PDF a new Citation takes the page the face was drawn on, and an existing one holds drawing to its saved page. The composer's property-menu cache is gone, `ArtifactRegionCrop` is the one place a region is cut from an Artifact (Auto Transcribe uses it too), and a save-error callout is titled "Not saved" under the value controls.

## S10-52 — Depiction auto-reconciler module

**Landed 2026-10-17** in [#437](https://github.com/mendahu/provenencia/pull/437). A person's likenesses now reconcile into a gallery. `depiction` is a module in the auto-reconciler keyed by the image File, so the same pixels drawn twice are one value and every distinct image stays under `likeness`; negatives, provisional members, majority, and weak evidence apply to faces as they do to names. An image `depictionvalues.Readable` refuses is no evidence. The cache stores each value's image File id in `auto_reconciler_values.value_image_file_id` (migration `000053.sql`, no foreign key, so it never keeps an image alive), the graph cache mirrors it so `Verify` covers it, and `CacheVersion` is 15, so every catalog rebuilds once on open. The Person field list skips depiction fields until S10-35 draws the gallery. `carries` is exported as `Carries`, `knownValueType` and the loader's `hasValue` copy are gone, and the likeness save fixtures are shared in `observationstest`. No FFI or Swift change.

## S10-54 — Depiction values over the FFI

**Landed 2026-10-17** in [#433](https://github.com/mendahu/provenencia/pull/433). The app can now send and read a likeness: a save carries the region and the absolute path of the crop the app rendered, Go stages it before the write (create, append, update, and bridge drafts), and every response and read returns the saved region and the image's `objects/…` path, with a stored half Go would not show read as empty. Create and append now return the same listed rows as update and connect (Property summary, depiction), in draft order. `files.RelPaths` resolves many Files to `objects/…` paths in one query and the source cover thumbnails use it. In Swift, `PropertyValueType.depiction` labels as *Depiction*, the store carries the value and an update's new crop beside the row, and the two new save errors have copy. Nothing in the app offers a likeness yet (S10-34).

## S10-32 — Depiction value

**Landed 2026-10-17** in [#430](https://github.com/mendahu/provenencia/pull/430). A photograph can now be evidence whose value is the picture. The engine has an Observation value type, `depiction`: an owned `depiction_values` row holding the region of the Citation's page that shows the face and an image File Go stored from the app's rendering of it, through `ingest.StageDepictions` and S10-53's checks. Editing a likeness replaces the row and releases the old picture, and deleting one releases the value, its File, and (after commit) its object, unless another row still uses the same PNG. A new catalog seeds `likeness` on Person (multiple, locked); an older catalog gains no Property, and researchers cannot create a depiction Property. Nothing in the app can make a likeness yet (S10-54, S10-34), and the auto-reconciler skips depiction Observations until S10-52. A date or name replaced by an update is now released through the same `ReleaseReplaced` as a depiction. A stored depiction is checked when Observations are listed: the region is hidden unless it is a valid region of the Citation's page, and the image id is hidden unless the stored File is a PNG; readers go through `depictionvalues.Readable`.

## S10-57 — Hold Shift to select

**Landed 2026-10-17** in [#428](https://github.com/mendahu/provenencia/pull/428). Holding Shift over the Evidence graph engages Select and lights its palette button; releasing Shift returns to pan, and Place and Connect still win. Shift-click adds a card (a relationship card too) to the selection or takes it out, without running the card's nested action, and a Shift rectangle adds to the selection while a plain one replaces it. Select armed from the palette disarms when its rectangle is released, even one that caught nothing. A click on empty canvas still clears the selection. Shift counts as up while a text field has focus or the window is not key.

## S10-56 — Object store stays inside objects/

**Landed 2026-10-17** in [#425](https://github.com/mendahu/provenencia/pull/425). A project folder someone else sent cannot make an import, a staged image, or a thumbnail write outside the project through a symlinked `objects/` or `objects/ab`, and a FIFO at an object path cannot block a read. Every object read, write, install, and removal now goes through an `os.Root` on `objects/` in `core/objectstore` (`OpenObject`, `WriteIfAbsent(projectDir, rel, …)`, `CreateTemp` and `Temp`, `RemoveObject`); `ingest.installObject`, `MatchesChecksum`, and `SyncDir` are gone. Deleting the last object in a shard also removes the `objects/ab/cd` and `objects/ab` folders it leaves empty. A folder holding only a Finder `.DS_Store` counts as empty. A project whose `objects/` is a link to another location now gets its own message on import, "This project’s objects folder can’t be used", instead of a generic failure. Shard folders already left empty in existing projects stay until the next removal in each. No schema or FFI change; the only Swift change is that message.

## S10-53 — Stage a client image

**Landed 2026-10-16** in [#424](https://github.com/mendahu/provenencia/pull/424). Go can accept an image the Mac app rendered and store it safely as a File; nothing calls it until S10-32. `ingest.StageImage` gives the app's temporary file every check an import gets, decodes only PNG or JPEG within the pixel budget through the one codec now in `core/raster` (shared with thumbnails), and stores its own PNG of at most 2048 pixels on the long edge, never the app's bytes. The staged object is written inside the write that commits its `files` row, so a failed write leaves no row and no object. Every exported `deleteimpact` release now takes the write's `*database.Tx` and schedules its own purge, so no delete can drop the objects a release freed. `catalogsession.Do` is named, at the purge and in the data model, as what keeps a delete from racing an import of the same bytes. No schema, FFI, or Swift change.

## S10-55 — File release from the declared model

**Landed 2026-10-16** in [#423](https://github.com/mendahu/provenencia/pull/423). Deleting one of two Artifacts whose Files share a deduplicated thumbnail keeps the thumbnail and its other link, and the thumbnail goes with the second Artifact. A File is deleted only when no declared foreign key into `files` points at it, and releasing one removes only its own thumbnail links. Owned releases (date and name values, Files) are read from the `BucketOwnedOutbound` foreign keys in `catalogmodel`, so the hand-kept list is gone and a new owned File pointer is two FK declarations. `ReleaseReplaced` releases a value an update swaps out. Every delete that releases owned rows (`artifacts`, `observations`, `subjects`, `sources`, `citations`, `metadatafields`) removes the freed objects after commit through `files.PurgeAfterCommit`, and the removal cannot leave the project's `objects/` folder through a symlink (`objectstore.RemoveObject`). No schema, FFI, or Swift change.

## S10-37 — Collation locale

**Descoped 2026-10-16.** A project collation locale (default `en`, no picker) and `x/text/collate` sort order for name lists are in [`ideas/international-names.md`](../../ideas/international-names.md). Name lists keep their current `sort_key` order. Matching's accent fold (S10-24) does not change.

## S10-31 — Related events and place chains

**Descoped 2026-10-16.** Birth~baptism and death~burial term affinity, and crediting a `part_of` place chain as a partial, are in [`ideas/archive/name-matching-enhancements.md`](../../ideas/archive/name-matching-enhancements.md).

## S10-45 — Seed `interests` on Person

**Landed 2026-10-12** in [#382](https://github.com/mendahu/provenencia/pull/382), with S10-38. A new catalog's Person properties include Interests: text, more than one value, bound to `person` and unlocked, so the Properties page can unbind it. Install writes it on catalog create only; opening an older catalog does not add it. [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2 and §3.3 list the key.

## S10-38 — Seed event type `engagement`

**Landed 2026-10-12** in [#382](https://github.com/mendahu/provenencia/pull/382). A new catalog's event types include Engagement, beside Marriage. Two subjects title *Engagement of A and B* on the list, the page, and in search, the same couple rule a marriage uses; three or more read *Engagement of A et al.* Install writes the term on catalog create only; no migration adds it to an older catalog. The same PR seeded the `travel` event type and the `gender` Property with its terms, which S10-23 used. [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4 lists the key.

## S10-30 — Date windows in matching

**Landed 2026-10-16** in [#415](https://github.com/mendahu/provenencia/pull/415). Dates compare by the days they cover. December 1817 against January 1818 scores 0.777 instead of reading as a year apart (0.533), while two points in one calendar year keep their tiers and two whole years apart keep their score; a month-level range keeps its months, so a September date no longer sits inside March to June (0.734). The cutoff still counts civil years and the fall-off counts days, from a new `autoreconcile.DateEnds` that gives each end of a date its day span. An Event's `date` now compares with another Event's `start_date` and `end_date` as a period, at the `date` weight, only when exactly one side has a `date` and no start or end. A birth on `14 MAY 1817` against a year-only `1818` now scores 7.0, not 7.2. Julian dates and double dating stay out. No view changed.

## S10-50 — Promote sheet lines show the values they compared

**Landed 2026-10-16** in [#413](https://github.com/mendahu/provenencia/pull/413). A term on a Promote sheet line reads as its label (`Male`, `Two-Spirit`), not its key, and a name with parts and no `Form` shows its parts. A row with no exhibit, such as one fixed to a handle whose only member claim is provisional, now shows each comparison that spoke with its two values and dates, never pinned. `match.Value.TermLabel` and `ValueRef.Date` carry the label and date; `graphcache.Vocabulary.TermsByID` replaces three hand-built term maps; the FFI sends `Row.Exhibits` only and `graphalign.Comparison.Pinned` is gone.

## S10-28 — Support, negatives, and which values matched

**Landed 2026-10-16** in [#411](https://github.com/mendahu/provenencia/pull/411). A handle's cached values count in proportion to their effective support (support minus the negatives against it) over the leading value's, so one stray "female" among three "male" Sources contradicts a female Subject and one "James Robins" on a handle that is mostly "Mary Smith" is not a name match. Equally supported values both lead. Negative Observations now load: one that is compatible with a leading value on the other side is a contradiction, as is a value the reconciler `denied`. Promote reads the same weights from the graph cache, so such a row reads as a conflict. Each `Reason` and `PropertyComparison` names the probe value and the candidate value. A multiple-cardinality Property (a Place's toponyms) gives every displayed value full weight, so a handle's minority toponym still matches. On the Promote sheet a row-level conflict shows as an unpinnable line when no exhibit shows it, and a negative Observation reads "not <value>". No catalog change; Promote's exhibit chooser keeps its own member picker.

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

## S10-44 — Unpromote a filed card

**Landed 2026-10-16** in [#418](https://github.com/mendahu/provenencia/pull/418). A promoted person, event, or place card on the Evidence graph has an Unpromote button (a U-turn arrow) at the right of its footer, after a hairline, beside the handle row that still opens the Person, Event, or Place page. Hovering it shows the tooltip "Unpromote" and a small fill; the handle row does not highlight. The confirm, "Unpromote James Robins?", says the filing is removed and the cards and handle stay, and lists under "Also unfiled" the relationships, participations, and locations on that Source that were filed through the card, with a note that the other end of each stays filed. Confirming deletes the Subject's accepted claim and those bridge claims, with their pins, in one audited revision; every card, Observation, edge, and handle stays, other members' claims and pins stay, and the cards return to Promote. A failed write stays in the confirm as a red callout. VoiceOver lists "Unpromote" after the open-handle action. Promote and Promote all now draw a turn-up arrow. The engine's one claim removal is `deleteimpact.ReleaseClaim`, which Subject delete and Unpromote both call, and the canvas gained a per-action tooltip.

## S10-51 — Draw Evidence cards in one place

**Landed 2026-10-16** in [#421](https://github.com/mendahu/provenencia/pull/421). Nothing looks different to a researcher. Every Evidence card, and the faded placement ghost, is now drawn by one view, `EvidenceCardView` (formerly `EvidenceAppKitCard`); the ghost is that view built with `init(ghostOf:)` at 62% opacity, with no VoiceOver element. The SwiftUI subject and bridge cards, their layout reporting, the hand-written accessibility chain and `EvidenceCitedPropertyRow.swift` are deleted, so the one VoiceOver action list is `EvidenceCanvasCardContent.namedActions(for:)` and Unpromote is listed only on a filed card. `GraphCanvasItem` now takes one `makeView` that returns an `NSView`, and the item layer no longer imports SwiftUI. The layout tests build and render the live card. The Unpromote hover fill and glyph read `PVControlSize.sm` instead of literal 22 and 13. `EvidenceSubjectCard` and `EvidenceBridgeCard` stay as `@MainActor enum`s holding the shared geometry.
