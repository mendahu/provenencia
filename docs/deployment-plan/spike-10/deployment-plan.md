# Deployment Plan — Spike 10

Usability fixes from [`docs/dogfood/ux.md`](../../dogfood/ux.md) after Spike 9, plus three ideas pulled in from the parking lot: [depictions and likenesses](depictions-and-likenesses.md), [international names](international-names.md), and [name-matching enhancements](name-matching-enhancements.md). Each dogfood note or idea slice is one PR, unless two of them edit the same part of a view or the same comparer — those are combined below.

**A brief covers one view.** A view may have more than one brief. A PR that changes two views waits on a brief for each. Briefs are written just before the PR they gate ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md)), indexed in [`design/README.md`](design/README.md). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Slice 1 is done (2026-10-11). S10-01…S10-04 landed in [#361](https://github.com/mendahu/provenencia/pull/361)–[#364](https://github.com/mendahu/provenencia/pull/364). S10-05, S10-07, S10-08, S10-12, S10-13, S10-15, S10-42, and S10-43 landed in [#368](https://github.com/mendahu/provenencia/pull/368), [#369](https://github.com/mendahu/provenencia/pull/369), [#370](https://github.com/mendahu/provenencia/pull/370), [#371](https://github.com/mendahu/provenencia/pull/371), [#372](https://github.com/mendahu/provenencia/pull/372), [#375](https://github.com/mendahu/provenencia/pull/375), [#373](https://github.com/mendahu/provenencia/pull/373), and [#374](https://github.com/mendahu/provenencia/pull/374) (2026-10-12). S10-46, S10-47, and S10-48 were descoped in [#367](https://github.com/mendahu/provenencia/pull/367). S10-06 was descoped in [#368](https://github.com/mendahu/provenencia/pull/368): searching `marriage` already lists a title that contains the word ahead of a fuzzy name. S10-10 was descoped in [#370](https://github.com/mendahu/provenencia/pull/370): a common cluster belongs in [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md), which covers more than a default role. S10-11 was already done in [#356](https://github.com/mendahu/provenencia/pull/356): a citation with 70 Observations opens quickly. S10-14 and S10-D1 were descoped in [#370](https://github.com/mendahu/provenencia/pull/370): an adaptive choice field is not a priority for this spike. S10-26 was descoped 2026-10-12: Daitch–Mokotoff is [`ideas/daitch-mokotoff.md`](../../ideas/daitch-mokotoff.md). Later landings go in [`completed.md`](completed.md).

> **Goal of this spike:** the annoyances logged while entering real research are gone, and the app still behaves the way Spike 8 and Spike 9 decided.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Composer stays usable at 20 Observations.** Saved rows scan. Adding an Observation is the same for a bridge and a plain Observation, and you are already looking at it, ready to type. A bridge between two subjects on this Source can be started here, and another after it, without a trip back to the Evidence graph. A saved row can be duplicated into a new unsaved form. Choosing a Citation does not start Auto Transcribe.
2. **The Evidence graph holds a real Source.** The canvas grows before cards hit the edge. Pan and zoom survive a trip to the composer. The grid is a tile, so a zoomed-out or grown document does not turn the background into one giant bitmap. Zoomed-out edges stay legible. The card stays the full card at every zoom. Button hits match the buttons. A hotkey or toolbar control box-selects cards and moves them together.
3. **Files.** A transcription that overflows shows a scrollbar. A file dropped on the Source page opens Add Artifact with that file chosen; it is stored when the researcher names it and confirms. Artifacts are cards. The file opens from a card and from the Evidence graph.
4. **Lists.** Sources, People, Events, and Places show when the row last changed. Coming back through history lands on the row that was at the top. The Sources list remembers sort, type filter, evidence-status filter, and text search.
5. **Entry defaults.** A date modal opens on Year and, when the transcription has a date, that date is already filled. A full name line can be split into parts. The name-part type control does not jump width and cycles same-letter matches.
6. **Small chrome.** Create Source asks for the name first. Enter saves a Property title. Clicking outside a sheet closes it, and asks first when the sheet is dirty. Opening an existing project selects the oldest user, not "create a new user." Searching `marriage` ranks title-word hits above fuzzy name prefixes. People, Places, and Events use one mark each, and the Event mark is not an hourglass. A Person page shows sex or gender, with gender winning when both exist.
7. **Names match the way records are written.** José matches Jose in suggestions and stays a separate grouping on the Person page. Jas. matches James. O Brien matches O'Brien. A rare surname outweighs Smith. Name format profiles live in the catalog; the Western profile is the one this spike seeds. A negative observation counts as a contradiction.
8. **A face is evidence.** A polygon on a photo is a depiction Observation. The Person page shows those crops, and one can be marked to lead that page and the People list. Promote compares two likenesses as pictures. Empty transcription is valid.
9. **A filing can be taken back.** Unpromote on a promoted card deletes that Subject's Identity Claim and the bridge claims on this Source that used it, after a confirm that names those bridges. The card shows Promote again. The handle stays. Other members' claims are not rejected.
10. **Long titles stay in their column.** The Promote target menu keeps its width when a title is long. The full title is available when the label is cut off. Typing `par` in a combo box lists *parent* above *grandparent*.
11. **Interests are a Person property.** A new catalog has an unlocked, multiple, text `interests` Property on Person. An existing catalog is unchanged.

## Where the notes went

| Dogfood note | PR | Brief |
| --- | --- | --- |
| Evidence graph button hit boxes | S10-01 | — |
| Evidence graph misrenders when zoomed far out | S10-02 | — |
| Evidence graph canvas is too small | S10-03 | — |
| Evidence graph viewport resets after the composer | S10-04 | — |
| Multi-line text areas need a scrollbar | S10-05 | — |
| Omnibar ranks fuzzy name matches above exact title words | S10-06, descoped 2026-10-11 | — |
| Onboarding defaults to a new user | S10-07 | — |
| Enter should save a Property title edit | S10-08 | — |
| Sources list sort and filter don't persist | S10-09, folded into S10-41 | S10-D8 |
| Prefill sensible defaults for common Observation kinds | S10-10, descoped 2026-10-12 | — |
| Citation composer slows down with many Observations | S10-11, done in #356 | — |
| Click outside a modal sheet to close it | S10-12 | — |
| Prefill the structured date modal + open focused on Year | S10-13 | — |
| Adaptive choice field | S10-14, descoped 2026-10-12 | S10-D1, descoped |
| Create Source modal: Name first, then Type | S10-15 | — |
| Rethink the Observation list + Add Observation focuses search | S10-16 | S10-D3 |
| Auto-fill name parts + name part type dropdown | S10-17 | S10-D4 |
| Drag a file onto the Source page + artifact cards | S10-18 | S10-D5 |
| Rethink the Event icon + sidebar icons match Subject cards | S10-19 | S10-D16, S10-D7 |
| File select and Open on the graph + box-select | S10-20 | S10-D6 |
| Last updated on list views | S10-21, S10-22, S10-41 | S10-D8…D11 |
| Show sex / gender on the Person detail page | S10-23 | S10-D12 |
| Accent folding + spaced particles | S10-24 | — |
| Abbreviations + nicknames | S10-25 | — |
| Phonetic tier (Daitch–Mokotoff) | S10-26, descoped 2026-10-12 | — |
| Name frequency in this project | S10-27 | — |
| Support weighting + negative evidence | S10-28 | — |
| Name patterns: schema, Western seed, loaders | S10-29 | — |
| Date windows in matching | S10-30 | — |
| Related event types + place chains | S10-31 | — |
| Depiction value type + `likeness` | S10-32 | — |
| Crop cache | S10-33 | — |
| Composer depiction row | S10-34 | S10-D15 |
| Person gallery + list thumb | S10-35 | S10-D12, S10-D9 |
| Preferred portrait | S10-39 | S10-D12, S10-D9 (same boards as S10-35) |
| Promote compares two crops | S10-36 | S10-D13 |
| Project collation locale (no picker) | S10-37 | — |
| Seed event type `engagement`; couple title like marriage | S10-38 | — |
| List scroll on the history entry | S10-40 | — |
| Sources list filters and text search | S10-41 | S10-D8 |
| Promote target menu truncates a long title | S10-42 | — |
| Combo box ranks a prefix above a mid-word match | S10-43 | — |
| Unpromote on a promoted Subject card | S10-44 | S10-D6 (same board as S10-20) |
| Seed Person property `interests` | S10-45 | — |

**Left in the dogfood log**

- **Dense Evidence graphs (census-scale).** Descoped 2026-10-08. Layers and filters stay in the dogfood log. Canvas growth landed (S10-03). Box-select is still S10-20. Neither hides part of a Source.
- **Automatic structure from the Artifact**, beyond the two slices in S10-13 and S10-17. No catalog writes from a detector. No OCR-on-PDF, Live Text, or Foundation Models.

**Back in the ideas folder** (not PRs)

- [Depictions leftovers](../../ideas/depictions-and-likenesses.md) — other keys, signature, transparent polygon, Windows PDF draw.
- [International-name leftovers](../../ideas/international-names.md) — other name patterns, transliteration, Beider–Morse, calendars and double dating, per-person `name_format`, a collation picker.
- [Daitch–Mokotoff](../../ideas/daitch-mokotoff.md) — phonetic tier under near-spelling. Replaces S10-26. Lee and Leigh do not share a code.
- [Name-matching leftovers](../../ideas/name-matching-enhancements.md) — population frequency, known-as, concluded values, calibration, researcher-defined weights, age-to-birth-year, rejected-claim memory, blocking.
- [Evidence graph structure shortcuts](../../ideas/evidence-graph-structure-shortcuts.md) — place a wedding, birth, or residence in one action, or paste a cluster already on the canvas. Replaces S10-10.

## Working rules

These are the calls for this spike. The questions at the bottom are answered; nothing in them still blocks a PR.

- **The observation list is S10-D3.** The board reconsiders the whole citation composer against the requirements. It is not a list of features added to the current form. Outcomes, not a layout: at 20 observations the rows scan; a bridge and a plain observation are added the same way, and the new one is where you are looking; keyboard focus enters the first field; choosing a citation does not start Auto Transcribe or Paste. A bridge between two subjects on this Source can be started in the composer, saved, and followed by another, without returning to the Evidence graph. The pair of kinds picks the bridge type. The role or type starts empty. If the saved list is grouped by subject, a bridge is shown under each end and both lines are that one bridge. A saved row can be duplicated into the new-observation form with the same values; that copy is not saved until the researcher saves it. Connect on the graph still opens the composer with the ends filled, and the server still places the card at the midpoint. Today `CitationConnections` inserts a bridge at index 0 and `CitationObservationRows.addDraft` appends, and `focusedID` does not move keyboard focus. The Citation `PVSelect` and the Auto Transcribe / Paste button share the trailing edge of the form column. Confirm-before-overwrite stays. The brief owns the layout, including its seeds and the two whole-composer directions. Insert-at-top is not required.
- **An adaptive choice field is not in this spike.** S10-14 and S10-D1 are descoped (2026-10-12). Term fields keep the control they have. The composer rethink does not swap a short list to radios or a long list to a combo box.
- **A default role does not cover a repeated cluster.** S10-10 is descoped (2026-10-12). Prefilling `subject` or `part_of` on a generic new connection stays out. Placing a wedding, birth, or residence in one action, and pasting a cluster already on the canvas, are [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md).
- **Date prefill is silent.** First match in the transcription, current locale, Gregorian patterns a genealogist actually types (day month year, month day year, year alone, ISO). 80–90% is the bar. No "suggested" chrome, so the date modal does not need a brief. Year is focused even when the prefill is empty. Landed in S10-13.
- **Name split does not clobber edits.** Blur of the full-form line fills parts only when the part list is empty. The Split button refills from the line; if any part was edited by hand, confirm before replacing. A comma means the text before it is one surname. A later comma is not kept on the following word. Otherwise a compiled title list peels prefixes and suffixes when another word remains, one leftover word is a given name, and the last of two or more leftovers is the surname.
- **Outside click closes a sheet.** If the sheet is dirty, use the existing confirm (discard or keep editing) — the same idea as leaving the composer with unsaved work. A clean sheet closes. Read-only sheets close. This is behavior on the shared sheet presentation, not a new dialog style. Landed in S10-12.
- **Graph camera is session state, not catalog state.** `GraphCanvasCamera` is explicitly not research data. Coming back from the composer, Back, or breadcrumbs restores magnification and content offset for that Source for the rest of the session. A persisted per-user table is out.
- **Canvas grows.** `EvidenceGraphPlacement.contentSize` is a fixed 4000×4000 and placement clamps to `maxCell`. When a card's frame nears the document edge, grow the document (and the clamp) by a step. The grid is a tile, so growth does not allocate a larger bitmap. No new chrome.
- **Zoomed-out grid and edges are a restore, not a new look.** Floor is `GraphCanvasCamera.minMagnification` (0.25). The grid is one repeated cell and must not turn solid black; it may fade out as a whole. Edges must fade or hold together, not disappear one at a time. No brief.
- **The full card stays at every zoom.** S10-46 and S10-D14 are descoped (2026-10-11). The camera stays `0.25`…`1.25`. Drawing the card in AppKit made a zoomed-out graph cheap enough that it does not switch to an icon-and-label face.
- **Every card is mounted.** S10-47 is descoped (2026-10-11). Opening a graph builds a view for each card on the Source. Positions, hit targets, and the VoiceOver rotors still cover every card. Culling views to the window is not the path.
- **No canvas cover.** S10-48 is descoped (2026-10-11). Drawing the card in AppKit made the first layout fast enough that a `ProgressView` over the graph is not needed, on open or while panning.
- **Hit boxes are the painted controls.** Toolbar place/connect buttons and the bridge card's open-citation control. The card icon hit helper (`EvidenceCardLayout`) is the pattern: hits centered on what is drawn.
- **Oldest user** means the earliest UUIDv7 `users.id`, not the first row of `ORDER BY display_name`. If the signed-in install identity is already in the project, keep selecting that person. If the project has no users, stay on create-new.
- **Enter follows the buttons on a single-line inline edit.** Save when that control is enabled. When it is not, Enter leaves the edit, the same as Cancel. Escape cancels. A multiline field still takes Enter as a newline. Landed in S10-08.
- **The Sources list is one page.** Sort, type filter, evidence status, and text search persist per project and user and restore together when that place is opened again, including after relaunch. People, Events, and Places have none of those controls. Last updated on a Sources row uses the revision Sources already store. That cell, the filters, the search field, and the persisted settings are S10-41, from S10-D8. They are not a persistence PR ahead of the brief. List scroll is the history-entry anchor in S10-40.
- **Last updated** is an audit scope, not a timestamp column. Sources already store `UpdatedRevision` (`scope_type = source`) and can show it once the row is designed. People, Events, and Places get a `canonical_entity` scope from S10-21, from `Effect.HandleScope` in `effects`. It moves on the handle's own row, its Identity Claims and evidence pins, and Observations (with their date and name values and notes) on a member Subject, so editing evidence moves the handle's revision. A property, term, citation, credibility, or Subject-label edit does not. The value is the revision's time. It is its own fixed-width column between the row text and the ref: “Today, 3:15 PM”, “Yesterday”, then a short date, in the viewer's locale, behind a history icon, 136pt wide, with “Last updated” and the full note-row stamp as help. No recorded revision reads a faint “—”. S10-D9 draws it; S10-D10, S10-D11, and S10-D8 copy it (answered 2026-10-14).
- **One Event mark.** The Event mark is the S10-D16 calendar. Cards, lists, omnibar hits, Promote, Properties, the sidebar, the history menus, and the three list empty states draw `subject_person`, `subject_event`, and `subject_place` through `PVMark`. The shell reads `WorkspaceSection.glyph`. Landed in S10-19.
- **Sex / gender styling** uses a seeded `gender` Property when the Person has one, otherwise `sex_at_birth`. The seed is `man`, `woman`, `non_binary` (labels Man, Woman, Non-binary). Researchers add further terms (`origin=user`), the same way they extend `role`. It is not product-locked, and there is no product `other` term. The page treatment recognizes the three seeded keys. A researcher-added term still counts as gender and wins over sex at birth; the brief draws that term without inventing a new color per key. Unknown, empty, and mixed (the reconciler already has a mixed state) stay neutral. `sex_at_birth` already appears as an ordinary field row when a record speaks to it (`PersonDetailContent`); this PR adds the page treatment and leaves that row in place. The treatment is a Badge naming the value above the name plus a tint on the header thumb (120pt on every Conclusion page, answered 2026-10-16). Man, Woman, and Non-binary each get a tint from colours the kit already has (non-binary is lapis, not copper, which is the Event ink; the woman pair is two `PVColor` aliases of existing palette steps; answered 2026-10-16); every other gender term shares one other tint. Sex at birth borrows: Female → Woman, Male → Man, Intersex and added terms → other, and its Badge says “· at birth” so it never reads as gender identity (answered 2026-10-15).
- **Composer performance** is already done. [#356](https://github.com/mendahu/provenencia/pull/356) builds a combo box's option list only while its menu is open, so a form of closed pickers does not lay out every choice. Checked 2026-10-12: a citation with 70 Observations opens quickly. No separate profile pass, and no split that shows the form while the file is still loading. S10-16 still has to stay comfortable at 20 Observations; that bar is the layout.
- **List scroll is a field on the history entry.** Sources, People, Events, and Places store the id of the top row on that entry (`listScrollAnchor` in [`navigation-history.md`](../archive/spike-3/navigation-history.md)). It is ignored for identity, the same way `title` and `ref` are. The list reports the row as the researcher scrolls; the next Back, Forward, jump, or `go(to:)` writes it onto the entry being left, once. Showing that entry restores the row, including after relaunch, because history already restores the entry. A visit that does not pass an anchor (sidebar, omnibar) starts at the top and, when the place coalesces, clears the anchor on the current entry. Changing Sources sort or type filter on this visit clears the anchor and scrolls to the top. A row that is gone opens at the top. S10-41 does not add scroll. Status and search, when they land, clear the anchor the same way sort and type do.
- **Sources evidence filters count primaries in one read.** A primary is a person, event, or place. Empty: no primaries. In progress: at least one primary has no accepted Identity Claim. Fully promoted (the filter's label; answered 2026-10-15): at least one primary, and every primary has an accepted claim. Bridges are not a filter axis. The counts are one grouped query beside `sourcegraphprogress`, not a query per row. Text search filters the list already loaded, by title and ref, as you type. It is not the omnibar and not FTS. Status and query persist in the same per-project, per-user store as sort and type filter.
- **Unpromote deletes the filing. It does not reject it.** The button is on a promoted primary card, in the footer beside the control that opens the handle. The agreed S10-D6 board draws it as a ghost `PVIconButton`, small, unlink icon, with “Unpromote” on the accessible label and the help. A labeled button does not fit the 36pt footer. Confirm always, tone irreversible, so the button is not a trash can. The confirm names each bridge Subject on this Source (participation, relationship, location, place relationship) that has this Subject as an end and an accepted claim. One transaction deletes those claims and this Subject's claim, audited, through the write path. The Evidence cards stay; their footers show Promote again. The handle stays, even if it has no members left. Other members' claims are not rejected and not stripped. That review stays [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md). A rejection ("not this entity") is that review's action, not this button.
- **Box-select is a marquee of subject cards.** Hold Shift and drag empty canvas, or arm Select on the palette (its help is “Hold Shift to select”). Person, event, and place cards that intersect the rectangle are selected. Empty-canvas drag still pans when that tool is disarmed and Shift is up. Bridge cards are not selected and do not move with the group. S10-D6 draws it; S10-20 ships it, together with a header select of the Source's files and an Open button that uses the system app. The select is 208pt. Its menu row is a 48pt thumbnail, the label or the ref, Cover when that artifact is the cover, then filename, kind, and size, then the citation count. The closed label is the label or the ref. Its help is the filename.
- **A long select label truncates inside the width it was given.** `PVSelect`'s closed label ellipsizes rather than growing the control. The Promote target column (`PromoteLayout.targetColumn`) stays put; `PromoteTargetMenuRow` already limits the menu title to one line. When the closed label is cut, help shows the full string. S10-17 still gives the name-part type control a stable width sized to its longest label. Those labels are short, so truncation does not replace that fix. Landed in S10-42.
- **Combo box rank is three bands, then the original order.** In `PVComboBoxMatch`, a label that starts with the query ranks above a later word that starts with it (split on spaces and hyphens), which ranks above any other contains match, including a match only in the subtext. `par` lists *parent* above *grandparent*. This is the kit filter, not `core/search` (S10-06). Landed in S10-43.
- **`interests` is prose, not a matcher feature.** Key `interests`, label Interests, `value_type = text`, cardinality `multiple`, bound to `person`, unlocked (the Properties page may unbind it). Same install rule as `engagement`: new catalogs only, no migration. It is not in a match profile and not on the Person header.
- **A drop on the Source page opens Add Artifact.** The drop target is the content pane (identity header and the scrolling column). One allowed file is staged in the dialog's file row; the label and description stay empty. The file is stored when the researcher types a label and confirms. A rejected drop opens that same dialog on the existing callout and stores nothing. The artifact accordion becomes wrapping rows of cards: large preview, label, description, and Open / Use as thumbnail / Delete. S10-D5, in S10-18. The graph header's file select and Open button are S10-D6, in S10-20: the select lists every artifact that has a file, and Open sends the chosen file to the system app. No files: both controls are absent. No in-app viewer window.
- **Name part types stay a select.** The compiled list is unchanged by S10-29. It is still a fixed-width `PVSelect`. User-minted part types stay out. The Western split rules in S10-17 do not become profile-aware.
- **Accent fold is matching only.** `wordSimilarity` scores a fold-only match at 0.95. `autoreconcile.NormalizeForm` and `sort_key` stay as they are, so José and Jose remain separate groupings on the Person page. Display never folds.
- **Word tiers, top to bottom:** exact, abbreviation or nickname (just under exact), accent fold (0.95), then the existing near-spelling score. Daitch–Mokotoff, which would sit under near-spelling, is descoped ([`ideas/daitch-mokotoff.md`](../../ideas/daitch-mokotoff.md)). Spaced particles (`O`, `Mc`, `Mac`, `St`, `D`, `Fitz`, and the rest of that short list) rejoin with the next word before scoring. A lone initial (`O.`) does not rejoin.
- **Abbreviation and nickname tables are seeded vocabulary** a project can extend, same as `role`. Not a compiled-only list, and not a population file.
- **Name frequency is this project's inverse frequency** over name words. No seeded census of names.
- **Support weighting and negatives land together.** A value counts in proportion to its support, and a negative Observation is a contradiction. Reasons name the two values that scored. A concluded value does not get a special weight: Reconciliation Claims are not in this spike.
- **Patterns come from the catalog.** `name_format_profiles` is real. Seed `western` only. Matching and display read those rows. Every name uses the project default until a per-person `name_format` claim exists. Spanish dual surnames and a `patronymic` part are unscheduled ([`ideas/international-names.md`](../../ideas/international-names.md)). No `maternal_surname` part type.
- **Collation is a project setting with no picker.** Default locale `en`. Name ordering uses `golang.org/x/text/collate`. Changing the locale in the UI waits for a settings place.
- **Dates in matching use the windows S9-21 already stores.** Spans and gaps are days. ABT doubles tolerance only when both dates are points. An event `date` can overlap `start_date` / `end_date`. Calendars and double dating stay out.
- **Birth and baptism partly match, and so do death and burial.** A small term-affinity table, same shape as name-role affinity. Place comparison walks `part_of` (already shipped) so York and Toronto can relate. Flat toponym text stays the strong match.
- **A depiction's value is its Citation.** New Observation `value_type = depiction`. No image column and no copied bytes. `likeness` is the seeded person Property of that type, cardinality many (a gallery). Place, event, and signature keys wait. Empty transcription is valid. The crop is a bounding-box JPEG, longest edge 512, cached by File plus locator, minted when the Observation is saved. The composer draws a PDF crop with PDFKit. A missing cache is filled the next time a client can open that Artifact. Two Citations with the same pixels may share bytes; the derivative rows stay separate. Promote's evidence sheet shows the two crops and pins the two Observations.
- **A preferred portrait is a display pointer, not a claim.** One likeness Observation on that person. The gallery still lists every crop. The Person page lead and the Persons list thumb use the chosen crop. With no choice, both use the earliest likeness. Deleting that likeness clears the pointer and the thumb falls back. The choice is made on the Person page. Promote does not pick a portrait.
- **`engagement` titles like any other event.** Key `engagement`, label Engagement, next to `marriage` in `subjectvocab` `seedTerms`. New catalogs get it from Install. Existing catalogs are not migrated; the term is added there by hand. Titles follow the subject count for every event type: one subject is *{Type} of {A}*, two are *{Type} of {A} and {B}*, and three or more are *{Type} of {A} et al.* A marriage of two still reads *Marriage of A and B*. An engagement of two reads *Engagement of A and B*.

## Design track

**Every UI PR is gated by Claude Design briefs. A brief covers one view; a view may have more than one.** A PR that touches two views waits on a brief for each. Briefs follow [`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md) and live in [`design/`](design/). **Each brief is designed alongside its feature**, just before the PR it points at.

| Brief | View | Gates | Later on the same view |
| --- | --- | --- | --- |
| **S10-D1** | Adaptive choice field (kit board) | **S10-14**, descoped 2026-10-12 | — |
| **S10-D3** | Citation composer | **S10-16**, done 2026-10-12 | — |
| **[S10-D4](design/archive/S10-D4-name-parts-modal.md)** | Name parts modal | **S10-17** | Split, fixed-width type control. Does not wait on S10-29 |
| **[S10-D5](design/archive/S10-D5-source-artifacts.md)** | Source page — Artifacts | **S10-18** | Page drop opens Add Artifact. The list becomes cards |
| **[S10-D6](design/S10-D6-evidence-graph.md)** | Evidence graph | **S10-20** file select, Open, and box-select | S10-44 Unpromote on the promoted card |
| **[S10-D16](design/archive/S10-D16-event-mark.md)** | Design system — Event mark | **S10-19** | The `subject_event` glyph. Sidebar placement is S10-D7 |
| **[S10-D7](design/archive/S10-D7-sidebar-marks.md)** | Workspace shell | **S10-19** | Sidebar and toolbar. Depends on S10-D16. Empty states use the same marks |
| [**S10-D8**](design/archive/S10-D8-sources-list.md) | Sources list | **S10-41** | Last updated, evidence-status filter, text search, persisted sort and filters |
| [**S10-D9**](design/archive/S10-D9-persons-list.md) | Persons list | **S10-22** | S10-35 fills the thumb with a likeness. S10-39 makes that thumb the preferred portrait |
| [**S10-D10**](design/archive/S10-D10-events-list.md) | Events list | **S10-22** | — |
| [**S10-D11**](design/archive/S10-D11-places-list.md) | Places list | **S10-22** | — |
| **[S10-D12](design/S10-D12-person-detail.md)** | Person detail | **S10-23** | S10-35 gallery and S10-39 preferred portrait. Design both with the gender treatment, before S10-23 |
| **S10-D13** | Promote page | **S10-36** | Evidence sheet shows two crops |
| **S10-D14** | Evidence graph — zoomed-out card | **S10-46**, descoped 2026-10-11 | — |
| **S10-D15** | Citation composer — depiction row | **S10-34** | — |

S10-D6 is one board for the graph's new controls (open file, box-select, Unpromote) and it gates S10-20. S10-18 does not wait on it. S10-01…S10-04 do not wait on it. S10-D14, the zoomed-out face, is descoped with S10-46 and is not drawn. S10-D1, the adaptive choice field, is descoped with S10-14 and is not drawn. Term fields keep the control they have. S10-D2 is not drawn: name-then-type on Create Source is a field swap, so S10-15 stands on its own. S10-D9…D11 are three briefs for the shared conclusion lists. Events and Places specify the same last-updated cell as Persons so `ConclusionListPage` does not grow three treatments. S10-D8 is the Sources list, a different page, and it gates S10-41. That board draws the same last-updated cell, the evidence-status filter, and text search. Sort, type, status, and query ship from that one board. List scroll is the history-entry anchor in S10-40, including Sources. D9…D11 do not draw those controls. S10-22 changes no Sources-list chrome; it moves that list's cache and engine read onto the shared last-activity patch. Only the Persons list thumb becomes a likeness. S10-D6, S10-D9, and S10-D12 each cover a later PR on that same view, so those briefs are written with the later frame in them. The depiction row is its own brief, S10-D15, written just before S10-34. S10-D3 does not draw it. D12 includes the preferred-portrait control S10-39 fills, and D9 includes that crop as the list thumb. Whichever of D8 and D9 is designed first draws the last-updated cell; the other copies it.

## PR sequence

**✎ = design brief**, just before the PR it points at. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Evidence graph  (done 2026-10-11)
  S10-01  Button hit boxes — done, #361
  S10-02  Tiled grid; edges share one zoom rule — done, #362
  S10-03  Canvas grows before a card hits the edge — done, #363
  S10-04  Camera restored after the composer / Back / Forward — done, #364
  S10-46  Simplified cards — descoped. The full card stays at every zoom.
  S10-47  Mount near the window — descoped. Every card on the Source is drawn.
  S10-48  Canvas cover — descoped. AppKit cards draw fast enough; no spinner.
  Check: pan, zoom, open the composer, come back — same view. A toolbar click
         lands on the padding, not only the icon. Zoomed out, the grid is a
         tile and not black, and the full card is what you see. Opening a
         large graph shows every card, with no loading cover.

SLICE 2 — Behavior, no new chrome
  S10-05  PVTextArea shows a scrollbar when the text overflows — done, #368
  S10-06  Omnibar whole-word rank — descoped. A title word already beats a fuzzy name.
  S10-07  Existing project selects the oldest user — done, #369
  S10-08  Enter saves a Property title — done, #370
  S10-10  Observation defaults — descoped. A repeated cluster is an idea, not a default role.
  S10-11  Composer stays responsive — done, #356. 70 Observations already open quickly.
  S10-12  Outside click closes a sheet; dirty asks — done, #371
  S10-13  Date modal: focus Year, prefill the first transcription match — done, #372
  S10-42  Promote target menu truncates a long title — done, #373
  Check: a long transcription scrolls. A long Promote title stays inside
         the target column.

SLICE 3 — Combo box
  S10-43  Combo box: a prefix outranks a mid-word match — done, #374
  S10-14  Adaptive choice field — descoped. S10-D1 is not drawn.
  Check: typing `par` lists parent above grandparent.

SLICE 4 — Composer and the dialogs it opens
  S10-15  Create Source: name, then type — done, #375
  ✎ S10-D3 ──▶ S10-16  Observation list rethink.
  S10-29  Name patterns (schema, Western seed, loaders) — done, #377
  ✎ S10-D4 ──▶ S10-17  Name parts: split, fixed width, letter cycle
  Check: twenty observations scan. Add a bridge and a plain Observation —
         both land where the board put authoring, and you can type.
         Start a relationship between two people without leaving, then
         another. Duplicate a saved row, change one value, save: the
         original is still there. Choosing a citation does not start Auto Transcribe. "Smith, John"
         splits to Surname, Given. Pressing S twice reaches Surname.

SLICE 5 — Artifacts and multi-select
  ✎ S10-D5 ──▶ S10-18  Page drop opens Add Artifact; artifact cards
  ✎ S10-D16 ──▶
  ✎ S10-D7 ──▶ S10-19  Event mark, then sidebar and toolbar. Does not wait on S10-18 — done, #393
  S10-49  Glyph consolidation: one subject-mark rule, PVGlyph in thumbnails, omnibar, Properties (after S10-19) — done, #394
  ✎ S10-D6 ──▶ S10-20  File select, Open, and box-select (after S10-01 and S10-03) — done, #396
  Check: drop a file on the page and Add Artifact opens with that file
         chosen. Name it and confirm: the card is there. Cancel: it is not.
         Open the file from the card. On the graph, pick a file and Open
         sends it to the system app. A source with no file shows neither
         control. Draw a box, drag once, the group moves. Sidebar
         Event matches the card.

SLICE 6 — Lists and the Person page
  S10-40  People, Events, Places, and Sources store the list's top row on the history entry — done, #392
  S10-21  canonical_entity audit scope + list revision reads — done, #398
  S10-22  Last updated on People, Events, and Places — done, #402
  S10-41  Sources list: last updated, filters, search, persisted sort and filters — done, #405
  S10-23  Person page sex / gender — done, #407. The board also draws the gallery S10-35 fills and the preferred portrait S10-39 fills.
  Check: edit a Source, leave, come back — the row's time moved, and sort,
         type, status, and search are still set. Scroll is the history entry's
         top row (S10-40): Back, Forward, a jump, and relaunch restore it; a
         sidebar visit starts at the top. Changing sort or type returns the
         list to the top. Edit a member's Observation — the Person row's time
         moved. Filter to graphs that still have an unpromoted person: a filed
         Source drops out. A Person with gender shows gender; one with only
         sex at birth shows that; unknown is neutral.

SLICE 7 — How names compare (no new chrome)
  S10-24  Accent fold (0.95) and spaced particles — done, #378
  S10-25  Abbreviation and nickname tables — done, #379
  S10-26  Daitch–Mokotoff below near-spelling — descoped. See ideas/daitch-mokotoff.md.
  S10-27  Project name-frequency weights — done, #409
  S10-28  Support-weighted scores, negative evidence, reasons name the pair
  S10-30  Date windows, ABT-on-spans, date vs start/end
  S10-31  Birth~baptism, death~burial, place part_of
  S10-37  Collation locale on name ordering (after S10-29)
  Check: José suggests Jose and the Person page still shows both spellings.
         Jas. Smith suggests James Smith. One female among many males does not
         cancel the contradiction. December 1817 vs January 1818 is days apart,
         not a year.

SLICE 8 — Likeness
  S10-32  depiction value type + likeness on person
  S10-33  Crop cache (File + locator), minted on save
  ✎ S10-D15 ──▶ S10-34  Composer depiction row (after S10-16)
  ✎ S10-D13 ──▶ S10-36  Promote evidence sheet shows the two crops
  S10-35  Person gallery + Persons list thumb (S10-D12 and S10-D9 already designed)
  S10-39  Preferred portrait leads the page and the list (same boards)
  Check: a group photo yields one crop per face. The Person page shows the
         gallery. Mark one face: it leads the page and the People list, and
         the other faces stay. Promote shows the two pictures. Clearing the
         cache does not lose the Citation.

SLICE 9 — Seeded vocabulary
  S10-38  Seed engagement; two subjects title "Engagement of A and B"
  S10-45  Seed Person property interests (text, multiple, unlocked)
  Check: a new catalog lists Engagement and Interests. Two people title as a
         couple. An existing catalog gains neither.

SLICE 10 — Take a filing back
  ✎ (S10-D6 already done for S10-20) ──▶ S10-44  Unpromote on the promoted card
  Check: confirm names the bridges. The card shows Promote again. The other
         person on that relationship stays promoted. The handle is still there.

S10-99 — Dogfood close / docs
```

S10-05…S10-13, S10-15, S10-19, S10-42, and S10-43 have no dependency on the graph slice. S10-24, S10-29, S10-30, S10-31, S10-32, S10-38, and S10-45 have no dependency on the dogfood slices. They are ordered here so review stays one stream. Parallel PRs are fine when **Depends on** allows it. S10-D4 does not wait on S10-29.

### Dependencies at a glance

```text
S10-01 hit boxes (done) ────────┐
S10-02 zoom, tiled grid (done) ─┤
S10-03 canvas grow (done) ──────┼──▶ S10-20 open file + box-select
S10-04 camera (done)            │         ▲
                                │         │
S10-10 defaults — descoped      │    S10-D6 (gates S10-20, not S10-18)
S10-11 perf — done, #356
S10-14 choice — descoped. S10-D1 is not drawn.
S10-D3 ──▶ S10-16 composer rethink

S10-46 simplified cards — descoped 2026-10-11
S10-47 mount cards near the window — descoped 2026-10-11
S10-48 canvas cover — descoped 2026-10-11

S10-40 list scroll — done, #392. People, Events, Places, and Sources, on the history entry.
S10-21 scope ──▶ S10-22 last updated (D9…D11)
S10-22, S10-D8 ──▶ S10-41 Sources list (last updated, filters, persist)
S10-D5 ──▶ S10-18 source page drop and artifact cards
S10-D6 ──▶ S10-20 graph file select + Open, and box-select
S10-D6 ──▶ S10-44 unpromote (button drawn on the S10-20 board)
S10-D12 (gender + gallery + preferred portrait) ──▶ S10-23
S10-D4 ──▶ S10-17 name parts
                 └──▶ S10-37 collation

S10-24 fold ──▶ S10-25 tables
S10-26 phonetic — descoped 2026-10-12. See ideas/daitch-mokotoff.md.
S10-28 support + negatives (after S10-24, so fixtures settle once)

S10-32 depiction type ──▶ S10-33 crop cache ──▶ S10-34 composer row (after S10-16; ✎ S10-D15)
                      └──▶ S10-36 Promote crops
S10-33 ──▶ S10-35 gallery (after S10-23; briefs D12 and D9 already drawn) ──▶ S10-39 preferred
```

---

## PRs

### Slice 1 — Evidence graph correctness

#### S10-01 — Button hit boxes

**Landed 2026-10-11** in [#361](https://github.com/mendahu/provenencia/pull/361). See [`completed.md`](completed.md).

#### S10-02 — Zoomed-out grid and edges

**Landed 2026-10-11** in [#362](https://github.com/mendahu/provenencia/pull/362). See [`completed.md`](completed.md).

#### S10-03 — Canvas grows

**Landed 2026-10-11** in [#363](https://github.com/mendahu/provenencia/pull/363). See [`completed.md`](completed.md).

#### S10-04 — Camera survives leaving the graph

**Landed 2026-10-11** in [#364](https://github.com/mendahu/provenencia/pull/364). See [`completed.md`](completed.md).

#### S10-46 — Simplified cards when zoomed out

**Descoped 2026-10-11**, in [#367](https://github.com/mendahu/provenencia/pull/367). Drawing the card in AppKit made the zoomed-out graph fast enough. The full card stays at every magnification in `0.25`…`1.25`. S10-D14 is not drawn.

#### S10-47 — Mount hosting views only for cards near the window

**Descoped 2026-10-11**, in [#367](https://github.com/mendahu/provenencia/pull/367). The same AppKit draw builds a view for every card on the Source. Cards outside the window stay mounted. Hit targets and both VoiceOver rotors still include every card.

#### S10-48 — Cover the canvas until the visible cards have drawn

**Descoped 2026-10-11**, in [#367](https://github.com/mendahu/provenencia/pull/367). Drawing subject and bridge cards in AppKit made the first layout fast enough. The graph does not lay a `ProgressView` over the canvas on open, and panning or zooming does not show one either. The cover that already replaces a destination while its query handles are missing stays as it is.

### Slice 2 — Behavior, no new chrome

#### S10-05 — Text area scrollbar

**Landed 2026-10-12** in [#368](https://github.com/mendahu/provenencia/pull/368). See [`completed.md`](completed.md).

#### S10-06 — Omnibar whole-word rank

**Descoped 2026-10-11**, in [#368](https://github.com/mendahu/provenencia/pull/368). Searching `marriage` already lists a marriage certificate and a newspaper announcement ahead of obituaries for Marion Margaret. A fuzzy name prefix does not outrank a whole word in a title, so the scorer stays as it is. A shorter prefix title (`Ilminsterish notes`) still ranks above a longer whole-word title (`Ilminster parish register`); that is the existing test, not this note.

#### S10-07 — Oldest user on an existing project

**Landed 2026-10-12** in [#369](https://github.com/mendahu/provenencia/pull/369). See [`completed.md`](completed.md).

#### S10-08 — Enter saves a Property title

**Landed 2026-10-12** in [#370](https://github.com/mendahu/provenencia/pull/370). See [`completed.md`](completed.md).

#### S10-09 — Sources sort and filter persist

**Folded into S10-41** (2026-10-12). Sort, type filter, evidence status, and text search are one Sources list, designed once on S10-D8. Persisting today's sort and type filter before that board exists would store chrome the filters PR then replaces. The testable behavior moves with S10-41.

#### S10-10 — Observation defaults

**Descoped 2026-10-12**, in [#370](https://github.com/mendahu/provenencia/pull/370). Prefilling `subject` on a new person–event, or `part_of` on a new place–place, fills one field on a connection the researcher is still building by hand. A wedding, a birth, or a residence is the whole cluster, and an obituary's six grandsons are a paste of a cluster already on the canvas. Both are [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md). The generic connection stays as it is.

#### S10-11 — Composer performance

**Done** in [#356](https://github.com/mendahu/provenencia/pull/356). See [`completed.md`](completed.md).

#### S10-12 — Outside click closes a sheet

**Landed 2026-10-12** in [#371](https://github.com/mendahu/provenencia/pull/371). See [`completed.md`](completed.md).

#### S10-13 — Date modal focus and transcription prefill

**Landed 2026-10-12** in [#372](https://github.com/mendahu/provenencia/pull/372). See [`completed.md`](completed.md).

#### S10-42 — Long titles in a select

**Landed 2026-10-12** in [#373](https://github.com/mendahu/provenencia/pull/373). See [`completed.md`](completed.md).

### Slice 3 — Choice field

#### S10-43 — Combo box prefix rank

**Landed 2026-10-12** in [#374](https://github.com/mendahu/provenencia/pull/374). See [`completed.md`](completed.md).

#### S10-14 — Adaptive choice field

**Descoped 2026-10-12**, in [#370](https://github.com/mendahu/provenencia/pull/370). A kit field that swaps `PVRadio`, `PVSelect`, and `PVComboBox` by option count is not a priority for this spike. S10-D1 is not drawn. Term fields keep the control they have. The composer rethink (S10-16) does not wait on it.

### Slice 4 — Composer and the dialogs it opens

#### S10-15 — Create Source field order

**Landed 2026-10-12** in [#375](https://github.com/mendahu/provenencia/pull/375). See [`completed.md`](completed.md).

#### S10-16 — Observation list rethink

**Landed 2026-10-12** in [#384](https://github.com/mendahu/provenencia/pull/384). See [`completed.md`](completed.md).

**Brief:** [S10-D3](design/archive/S10-D3-citation-composer.md).

| | |
| --- | --- |
| **In** | Per S10-D3. Saved observations scan at 20. Adding a bridge and a plain observation is one behavior, and the new one is where you are looking, with keyboard focus in the first field. A bridge between two subjects on this Source can be started here, saved, and followed by another, without returning to the Evidence graph. A saved row can be duplicated into that form with the same values, unsaved until the researcher saves. Choosing a citation does not start Auto Transcribe or Paste. A citation with no transcription is still distinguishable. Term fields keep their control (S10-14 is descoped). The layout stays comfortable at 20 observations (S10-11's bar). |
| **Out** | Name parts modal (S10-17). Date modal (S10-13). The depiction row (S10-34, brief S10-D15). An adaptive choice field. A new bridge kind. Prefilling a role. Replacing Connect on the graph. |
| **Testable** | Twenty rows: you can read each one and see where it ends. Add a bridge and a plain observation: both land in the place the board chose, and you can type without clicking first. Start a relationship between two people on this Source, save it, and start another without leaving. Back on the graph, the card is between them. If the list is grouped by subject, that relationship is under each person, and editing one line changes the other. Duplicate a saved participation, change the person, and save: the original is unchanged and the new row has the other person. A person and a place do not make a bridge. Choosing a citation does not start Auto Transcribe. A citation with an empty transcription is distinguishable from the others. |
| **Depends on** | S10-D3 |

#### S10-17 — Name parts split and type control

**Landed 2026-10-13** in [#385](https://github.com/mendahu/provenencia/pull/385). See [`completed.md`](completed.md).

**Brief:** [S10-D4](design/archive/S10-D4-name-parts-modal.md).

| | |
| --- | --- |
| **In** | Per S10-D4 and the working rules: blur splits an empty part list; Split refills, confirming when a part was hand-edited. Type control has a stable width (sized for the longest label, not the selection) and stays a `PVSelect`. Same-letter keypress cycles matches. That cycle is fixed in `PVSelect` (`handleTriggerKey` / session), so every select gains it; the width fix stays on this field. Split rules stay the Western ones. |
| **Out** | Profile-aware splitting. A combo box for part type. User-minted part types. |
| **Testable** | `Smith, John Henry` → Surname, Given, Given. `John Henry Smith` → Given, Given, Surname. `John` → Given. `Dr. John Smith Jr.` → Prefix, Given, Surname, Suffix. The type control does not resize. Blur with existing hand-edited parts does nothing. Split then asks. Pressing S moves from Surname prefix to Surname. |
| **Depends on** | S10-D4 |

### Slice 5 — Artifacts, marks, multi-select

#### S10-18 — Artifact drop and cards

**Landed 2026-10-14** in [#387](https://github.com/mendahu/provenencia/pull/387). See [`completed.md`](completed.md).

**Brief:** [S10-D5](design/archive/S10-D5-source-artifacts.md).

| | |
| --- | --- |
| **In** | Per S10-D5. The Source page content pane highlights and accepts a drop. The drop opens Add Artifact with that file in the file row and the label empty (`IngestMediaPolicy` stages it; create runs on confirm). A rejected drop, including more than one file, opens the dialog on the existing callout and adds nothing. The accordion becomes wrapping rows of cards: large preview, label, description, ref, and Open / Use as thumbnail / Delete, with edit of the label and description on the card. |
| **Out** | The Evidence graph, including header Open file, box-select, and Unpromote (S10-20 and S10-44, brief S10-D6). Storing the file on drop. Prefilling the label from the filename. Dropping on a card to attach. A new Artifact model. Keeping the expanding row. |
| **Testable** | Drop a PNG on the metadata column: the pane highlights, Add Artifact opens, the file row shows the PNG, and the label is empty. Cancel: no new artifact. Enter a label and confirm: a card appears and the toast says it was created and ingested. Drop a zip, or two files: the dialog shows the ingest callout and the cards are unchanged. A card opens its file without expanding. A fileless card has no Open button. |
| **Depends on** | S10-D5 |

#### S10-19 — Event mark and sidebar icons

**Landed 2026-10-14** in [#393](https://github.com/mendahu/provenencia/pull/393). See [`completed.md`](completed.md).

**Briefs:** [S10-D16](design/archive/S10-D16-event-mark.md) (the `subject_event` glyph) and [S10-D7](design/archive/S10-D7-sidebar-marks.md) (sidebar, toolbar, empty states).

| | |
| --- | --- |
| **In** | Per S10-D16 and S10-D7. A new `subject_event` mark (calendar / datebook — the kit board draws it). The key stays. Person and Place marks stay. Sidebar People, Events, and Places use those subject marks. The toolbar section icon uses that same source. List empty states for those three kinds use the mark. Cards, omnibar, Promote, and Properties already use `PVMark` and update with the asset. |
| **Out** | A redraw of Person or Place marks. New sidebar rows. A local Event glyph on the sidebar board. A new empty-state component. |
| **Testable** | Sidebar, toolbar, an Event card, the Events list empty state, and an omnibar Event hit show the same Event mark. People and Places match between sidebar and card. Sources in the sidebar still uses its symbol. |
| **Depends on** | S10-D16, S10-D7 |

#### S10-49 — Glyph consolidation

**Landed 2026-10-14** in [#394](https://github.com/mendahu/provenencia/pull/394). See [`completed.md`](completed.md).

| | |
| --- | --- |
| **In** | One `PVMarkKey(subjectTypeKey:)` rule replaces the per-feature subject-type → mark switches and string keys, pinned to the registry's `IconSymbol` by a Go test. `PVThumbnail.Content`, the omnibar lead, and the Properties type card take a `PVGlyph`. The Source types and Metadata empty states use their sidebar glyph. The S10-19 section-glyph test is trimmed to literal keys. |
| **Out** | New marks. `ConclusionListKind.mark` and detail-view literals. Any visual change beyond the two empty states. |
| **Testable** | An empty Source types list shows the tag symbol; an empty Metadata list shows the list symbol. Every other thumbnail, omnibar lead, and Properties card looks as before. |
| **Depends on** | S10-19 |

#### S10-20 — Open file and box-select

**Landed 2026-10-14** in [#396](https://github.com/mendahu/provenencia/pull/396). See [`completed.md`](completed.md).

**Plan:** [`prs/S10-20-open-file-box-select.md`](prs/S10-20-open-file-box-select.md) — Landed
**Brief:** [S10-D6](design/S10-D6-evidence-graph.md) (EG-1…EG-7). Unpromote is drawn on that board and ships in S10-44.

| | |
| --- | --- |
| **In** | Per S10-D6. The graph header shows a select of every artifact that has a file, and an Open button. The initial selection is the cover's file when that artifact has one, otherwise the first file. Open sends the selected file to the system app, the same path as the Source page. No files: both controls are absent. Hold Shift and drag on empty canvas, or arm Select on the palette (its help shows Shift) and drag. Person, event, and place cards whose frames intersect the rectangle are selected. Dragging one selected subject moves the group; each position persists the way a single drag does. Empty-canvas drag still pans when Select is disarmed and Shift is up. Escape or a second click of Select leaves the mode. |
| **Out** | Unpromote (S10-44). An in-app viewer window. Selecting or moving bridge cards. An undo stack. Shift-click to add one card. |
| **Testable** | Two files: the select lists both, and Open sends the chosen one to the system app. Changing the selection does not open it. A fileless artifact is not in the list. A Source with no file shows neither control. Box three subject cards with a bridge inside the rectangle: the bridge is not selected. Drag one subject, all three land together and survive reload. Without Select armed, dragging empty canvas still pans. Select works without holding Shift, and its help shows Shift. |
| **Depends on** | S10-01, S10-03, S10-D6 |

### Slice 6 — Lists and the Person page

#### S10-40 — List scroll position

**Landed 2026-10-14** in [#392](https://github.com/mendahu/provenencia/pull/392). See [`completed.md`](completed.md).

| | |
| --- | --- |
| **In** | People, Events, Places, and Sources store the id of the top row on the history entry. Showing that entry scrolls to the row, including Back, Forward, a jump, and relaunch. A visit that does not pass an anchor starts at the top. Changing Sources sort or type filter on this visit clears the anchor. A missing row stays at the top. |
| **Out** | S10-41's persisted status filter and search. A separate session store. A new history entry for each scroll. |
| **Testable** | Scroll to the bottom of People, open a Person, press Back: the list is at that row. Forward and a jump return to it. Quit and relaunch: the entry still has it. Open People from the sidebar: the list starts at the top. The same on Events, Places, and Sources. Change Sources sort or type: the list returns to the top. |
| **Depends on** | — |

#### S10-21 — Handle last-activity scope

**Landed 2026-10-14** in [#398](https://github.com/mendahu/provenencia/pull/398). See [`completed.md`](completed.md).

**Plan:** [`prs/S10-21-handle-last-activity.md`](prs/S10-21-handle-last-activity.md) — Landed

| | |
| --- | --- |
| **In** | New audit scope `canonical_entity` ([`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md)). Identity Claim writes scope to the handle. An Observation write on a member Subject also scopes to that handle, and still scopes to the Source. List reads for Persons, Events, and Places return the latest revision time the way Sources already return `UpdatedRevision`. Backfill existing history in the next migration. Resolver tests; no migration test. |
| **Out** | The row UI (S10-22). A `updated_at` column on `canonical_entities`. |
| **Testable** | Promote a Person: the handle has a revision. Edit a member Observation: that handle's revision moves, and the Source's revision still moves. A handle with no scope rows is listed with revision 0. Rebuild-style check: scope rows match the writes, including a delete of a member. |
| **Depends on** | — |

#### S10-22 — Last updated on People, Events, and Places

**Plan:** [`prs/S10-22-list-last-updated.md`](prs/S10-22-list-last-updated.md) — Landed #402

**Briefs:** S10-D9 Persons, S10-D10 Events, S10-D11 Places. Design the three before the PR. Events and Places copy the Persons cell. The Sources list is a different page; its last-updated cell ships in S10-41 from S10-D8, using the same treatment.

| | |
| --- | --- |
| **In** | Each of these three list rows shows when it last changed, from the revision S10-21 returns. Empty or never-audited reads as the brief's empty. This PR does not add sort to these lists. A note-only Observation save reports no handles (`observation_notes` has no `Handles`), so `CatalogQueryRegistry` does not reload the lists. This PR reports `effects.HandleScopes` ids and the write's own revision and time on `WriteEffects`, and the app patches the cached rows (People, Events, Places, and Sources) in place; it does not invalidate or reload them. Lists not on screen go stale and reload when shown. Source-page and graph writes (notes, metadata, artifacts, credibility, Subject create, update, and delete) return `WriteEffects` too, so a Source edit no longer reloads the lists. |
| **Out** | The Sources list cell, filters, and search (S10-41). A new list layout beyond the cell the briefs add. |
| **Testable** | People, Events, and Places show a time. Edit a member Observation: the Person row moves. Events and Places match the Persons cell. |
| **Depends on** | S10-21, S10-D9, S10-D10, S10-D11 |

#### S10-41 — Sources list

**Plan:** [`prs/S10-41-sources-list.md`](prs/S10-41-sources-list.md) — Landed #405

**Brief:** [S10-D8](design/archive/S10-D8-sources-list.md), board agreed 2026-10-15. One board for the page. It does not add a cell, a filter, or a search field to Events, Places, or Persons.

| | |
| --- | --- |
| **In** | Per the working rules and S10-D8. The row shows when the Source last changed. The time is `CatalogSource.updated` (revision and time), already read from the engine's last-activity cache and patched by S10-22; this PR draws it and moves the cell composition (`updatedDetail`, `labelWithUpdated`, `updatedColumnWidth`) from `ConclusionListRow` to one feature-neutral home that both the conclusion rows and `SourcesSplitRow` call, rather than copying it. Filter the list to empty, in progress, or filed, from one grouped count of primary subjects and their accepted claims, next to `sourcegraphprogress`. A text field filters the loaded rows by title and ref as you type. Sort, type filter, status, and query persist per project and user, and restore when the Sources place is shown again, including after relaunch (this is S10-09). Scroll is the history-entry anchor from S10-40 ([#392](https://github.com/mendahu/provenencia/pull/392)). This PR does not restore scroll. Changing status or search on that visit clears that anchor and scrolls to the top, the way sort and type already do. |
| **Out** | The same filters, search, or last-updated work on People, Events, or Places (scroll there is S10-40; last updated there is S10-22). Omnibar search. A per-row count query. Hiding bridge subjects as their own status. |
| **Testable** | A Source with no people, events, or places matches Empty and not the other two. A Source with one unpromoted person matches In progress. After that person is promoted, it matches Filed. Typing part of a title hides the other rows. Leave and come back: status, query, sort, and type filter are still set, and the row's time moved if the Source changed. Relaunch: those four settings are still set. A long list does not issue one count query per row. Scroll follows the history entry (S10-40). |
| **Depends on** | S10-22 (landed [#402](https://github.com/mendahu/provenencia/pull/402)), S10-D8 (board agreed 2026-10-15) |

#### S10-23 — Sex / gender on the Person page

**Brief:** [S10-D12](design/S10-D12-person-detail.md). **Plan:** [`prs/S10-23-person-gender.md`](prs/S10-23-person-gender.md). **Status:** Landed [#407](https://github.com/mendahu/provenencia/pull/407).

| | |
| --- | --- |
| **In** | The page treatment, per S10-D12 (PDP-1…PDP-6, board agreed 2026-10-16): a Badge above the name and a tint on the header thumb, which grows to 120pt on every Conclusion page. It follows gender when the Person has one value, else sex at birth. Man, Woman, and Non-binary (lapis) each have a tint; any other gender term uses the other tint. Female and Male borrow the Woman and Man tints and read “· at birth”; Intersex and other sex-at-birth terms use the other tint. Unknown, Indeterminate, empty, and mixed are neutral. The existing field rows stay. The `gender` Property and its three terms already ship from Install (`3d70139`, with S10-38 and S10-45); this PR adds no seed and the new § in `seeded-vocabulary.md` is already §3.9. |
| **Out** | The gallery (S10-35) and the preferred portrait (S10-39). This brief already draws both frames. A gender field on Event or Place. Editing claims. Using sex at birth as gender identity. |
| **Testable** | Person with only sex at birth: that treatment. Person with gender: gender wins even if sex at birth disagrees. Mixed gender: neutral, not a blend of two colors. A Person with neither: unchanged page. |
| **Depends on** | S10-D12 |

### Slice 7 — How names compare

#### S10-24 — Accent fold and spaced particles

**Landed 2026-10-12** in [#378](https://github.com/mendahu/provenencia/pull/378). See [`completed.md`](completed.md).

#### S10-25 — Abbreviations and nicknames

**Landed 2026-10-12** in [#379](https://github.com/mendahu/provenencia/pull/379). See [`completed.md`](completed.md).

#### S10-26 — Phonetic tier

**Descoped 2026-10-12.** Daitch–Mokotoff as a tier under near-spelling is [`ideas/daitch-mokotoff.md`](../../ideas/daitch-mokotoff.md). Lee codes as `800000` and Leigh as `850000`, so the pair this slice named does not share a code. Beider–Morse stays later ([`ideas/international-names.md`](../../ideas/international-names.md)). Smyth/Smith still meets by edit distance.

#### S10-27 — Name frequency

**Plan:** [`prs/S10-27-name-frequency.md`](prs/S10-27-name-frequency.md) — Landed #409

| | |
| --- | --- |
| **In** | Weight a name word by its inverse frequency among name words in this project, so a shared rare surname outranks a shared Smith. Computed at score time from the catalog (or a small derived count rebuilt with the search index — the PR picks one and tests it). A project with one household does not zero out every name. |
| **Out** | A seeded population table. |
| **Testable** | In a fixture with many Smiths and one Quirk, Quirk vs Quirk outranks Smith vs Smith when the rest of the name matches. |
| **Depends on** | S10-24 |

#### S10-28 — Support, negatives, and which values matched

| | |
| --- | --- |
| **In** | A candidate value counts in proportion to its share of support, so one stray name or one dissenting sex term cannot score as the consensus. Negative Observations load and score as contradictions. Each reason names the two values (or their ids) that produced it. |
| **Out** | Concluded values first. Remembering a rejected claim. A new Promote layout for the reason text (S10-36 covers crops only). |
| **Testable** | One female among several males still contradicts. One James Robins on a handle that is otherwise Mary Smith does not score as a full name match. A reason identifies the pair. |
| **Depends on** | S10-24 |

#### S10-29 — Name patterns from the catalog

**Landed 2026-10-12** in [#377](https://github.com/mendahu/provenencia/pull/377). See [`completed.md`](completed.md).

#### S10-30 — Date windows in matching

| | |
| --- | --- |
| **In** | `DateComparer` measures spans and gaps from the date windows in days. December 1817 vs January 1818 outranks a same-year pair only when the days say so — it stops counting as "a year apart" while May vs September stays a same-year partial. A range uses its months and days. ABT doubles tolerance only for two points. `date` may overlap `start_date` / `end_date` by window. |
| **Out** | Julian vs Gregorian. Double dating (1717/18). |
| **Testable** | The December/January pair scores above the old year-apart figure and below an exact day. `ABT 1817` vs a range uses the normal tolerance. A point date overlaps a start/end span that contains it. |
| **Depends on** | — |

#### S10-31 — Related events and place chains

| | |
| --- | --- |
| **In** | Term affinity: birth~baptism and death~burial score a partial, not a contradiction and not a full match. Place comparison can credit a `part_of` chain (York contained in Toronto, or the reverse) as a partial. An equal toponym stays the full score. |
| **Out** | A general gazetteer. Place renames beyond the chain the catalog already has (`succeeded_by` may be read the same way if the chain composer already exposes it; the PR says which). |
| **Testable** | Birth vs baptism scores between 0 and 1. Birth vs death still contradicts. A place and its parent score a partial; two unrelated places do not. |
| **Depends on** | — |

#### S10-37 — Collation locale

| | |
| --- | --- |
| **In** | Project setting for the collation locale, default `en`. Name sort order (list `sort_key` consumers that order names, and Sources A–Z if it sorts text) uses `golang.org/x/text/collate` for that locale. Comparison fold is not the sort. |
| **Out** | A control to change the locale. Swedish or Spanish tailoring beyond what CLDR does for the default locale. |
| **Testable** | With the default locale, accented and unaccented names sort in that locale's order, and the match score is unchanged. |
| **Depends on** | S10-29 |

### Slice 8 — Likeness

#### S10-32 — Depiction value type

| | |
| --- | --- |
| **In** | `value_type = depiction`. The value is the Observation's own Citation (artifact, page, region). No image column. Seed `likeness` on `person`, cardinality many. Empty transcription stays valid. Install path only ([`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md)). |
| **Out** | Place, event, and signature keys. The preferred-portrait pointer (S10-39). The crop bytes (S10-33). |
| **Testable** | A likeness Observation stores no raster. Two likenesses on one person both remain (they do not collapse the way two birth dates can). A Citation with an empty transcription and a region saves. |
| **Depends on** | — |

#### S10-33 — Crop cache

| | |
| --- | --- |
| **In** | Cache key is File plus locator, not the one-thumbnail-per-File unique key. Bounding-box JPEG, longest edge 512. Mint when the depiction Observation is saved. PDF: the composer asks PDFKit to draw the locator. A JPEG is clipped the same way. Disposable, not audited, not an Artifact. A changed polygon invalidates that entry. A missing cache is filled the next time a client that can open the Artifact draws it. Same pixels may share the stored file; derivative rows stay separate. |
| **Out** | A transparent polygon mask. A PDF engine in Go. |
| **Testable** | Saving a likeness writes a crop. Deleting the crop row leaves the Citation. Editing the polygon produces a new crop. Two Citations on the same pixels do not duplicate the blob if the PR shares it. |
| **Depends on** | S10-32 |

#### S10-34 — Composer depiction row

**Brief:** S10-D15, written just before this PR. Not drawn on S10-D3.

| | |
| --- | --- |
| **In** | Per S10-D15: pick the region, leave transcription empty, save a likeness (or another depiction property, if the board shows only likeness). The row uses the authoring place and focus rules from S10-16. |
| **Out** | The observation-list rethink (S10-16). Gallery (S10-35). |
| **Testable** | Draw a face on a group photo, add a likeness, save. The row shows the crop. Transcription can be empty. |
| **Depends on** | S10-16, S10-33, S10-D15 |

#### S10-35 — Person gallery and list thumb

**Briefs:** S10-D12 and S10-D9, already designed before S10-23 and S10-22.

| | |
| --- | --- |
| **In** | Person detail shows every likeness crop for the handle's member Subjects (the join in the idea note). Persons list lead thumb shows the earliest likeness when a crop exists, and the placeholder when it does not. Events and Places lists stay on the placeholder. |
| **Out** | Marking a preferred portrait (S10-39 fills that frame on the same boards). Place and event galleries. |
| **Testable** | Two faces of one person both appear on the page. The Persons list shows the earliest crop. A person with no likeness keeps the placeholder. Purging the cache and reopening regenerates it. |
| **Depends on** | S10-23, S10-33, S10-D12, S10-D9 |

#### S10-39 — Preferred portrait

**Briefs:** S10-D12 and S10-D9, already designed before S10-23 and S10-22. No second brief.

| | |
| --- | --- |
| **In** | One display pointer on the person to a single likeness Observation. Set and cleared from the Person page gallery the board drew. That crop leads the Person page and is the Persons list thumb. The other likenesses stay in the gallery. No pointer: both surfaces keep the earliest likeness from S10-35. Deleting the chosen likeness clears the pointer. Not a Reconciliation Claim, and not a rank among competing values. |
| **Out** | A preferred portrait for a place or an event. Picking one from Promote. A second Person-detail or Persons-list brief. |
| **Testable** | Two faces: mark one, and it leads the page and the People list while the other stays in the gallery. Clear the mark: the earliest leads again. Delete the marked likeness: the remaining one leads, and the pointer is gone. |
| **Depends on** | S10-35, S10-D12, S10-D9 |

#### S10-36 — Promote compares crops

**Brief:** S10-D13.

| | |
| --- | --- |
| **In** | On the Promote evidence sheet, a likeness comparison shows the two crops. Pinning pins the two Observations. It does not create a picture and does not write a transcription. |
| **Out** | A new Promote flow. Picking a preferred portrait. |
| **Testable** | Two people with a likeness: the sheet shows both crops. Pinning writes the same kind of pin a name pin writes. |
| **Depends on** | S10-33, S10-D13 |

### Slice 9 — Seeded vocabulary

#### S10-38 — Seed event type `engagement`

| | |
| --- | --- |
| **In** | Add `engagement` / Engagement to `subjectvocab` `seedTerms` for `event_type`, beside `marriage`. Install on catalog create only ([`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md)). [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4 lists the key. `eventtitle.Choose` titles every event type from the subject count: one subject is *{Type} of {A}*, two are *{Type} of {A} and {B}*, and three or more are *{Type} of {A} et al.* The two-subject template is that shape in `EventTitleDisplay` and in the search-index title, so a marriage still reads *Marriage of A and B* and an engagement reads *Engagement of A and B*. Follow [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md) for the template. No new chrome, so no brief. |
| **Out** | A migration that inserts the term into existing catalogs. Those projects get the term by hand. Matching affinity with marriage. A new role. Three or more subjects (still *et al.*). |
| **Testable** | A catalog created after this PR offers Engagement. Two subjects title *Engagement of A and B* on the list, the page, and in search. A marriage of two still titles *Marriage of A and B*. Three subjects of an engagement title *Engagement of A et al.* Opening an older catalog does not add the term. |
| **Depends on** | — |

#### S10-45 — Seed `interests` on Person

| | |
| --- | --- |
| **In** | Add `interests` / Interests to the person property seed: `value_type = text`, cardinality `multiple`, bound to `person`, unlocked ([`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md)). Install on catalog create only. [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2 and §3.3 list the key. No new chrome, so no brief. The composer and the Person page show it the way they show any other text property. |
| **Out** | A migration that inserts the property into existing catalogs. A match feature. A header line. A term list. Locking the binding. |
| **Testable** | A catalog created after this PR offers Interests on a Person, accepts more than one value, and lets the Properties page unbind it. Opening an older catalog does not add the property. |
| **Depends on** | — |

### Slice 10 — Take a filing back

#### S10-44 — Unpromote

**Brief:** [S10-D6](design/S10-D6-evidence-graph.md) (EG-8…EG-10), designed before S10-20. This PR fills the button that board drew on the promoted card.

| | |
| --- | --- |
| **In** | Per the working rules and S10-D6: a promoted primary card offers Unpromote in the footer, beside the control that opens the handle. The control is the ghost unlink icon button the board drew (accessible label and help “Unpromote”), not a labeled text button. Confirm always, tone irreversible. The body names each bridge Subject on this Source whose accepted claim has this Subject as an end. Confirm deletes this Subject's Identity Claim and those bridge claims in one `writes.Run` transaction, audited. Cards stay on the graph and show Promote. The handle is not deleted. Other members' claims are left for the parked review. |
| **Out** | Rejecting the claim ("not this entity"). The weak-claim queue, re-pin, and accept-as-is ([`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md)). Deleting the Evidence cards. Deleting the handle. Unpromote from the Person, Event, or Place page. |
| **Testable** | Unpromote a person who is in a filed relationship: the confirm names that relationship, both cards show Promote, the other person stays on their handle, and that handle still exists. Unpromote a person with no filed bridges: the confirm does not invent a bridge, and the card shows Promote. Cancel leaves the claim. |
| **Depends on** | S10-D6 |

### S10-99 — Dogfood close / docs

| | |
| --- | --- |
| **In** | Pulled notes already left [`docs/dogfood/ux.md`](../../dogfood/ux.md). Do not copy them back. Distill this spike with [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) when the bar above is true. The three idea notes live in this folder; the close distill keeps their decisions and drops the PR sequence. |
| **Out** | Pulling the parked density, extract, transliteration, or Reconciliation-Claim items in at the close. |
| **Depends on** | S10-01…S10-45 |

## Scope boundary

| In | Out |
| --- | --- |
| The dogfood notes in the table above | Census-scale layers / filters |
| Date prefill and name split as the cheap extract | Detector suggestions, guided extract, catalog writes from a parse |
| Accent fold, abbreviations, nicknames, project name frequency | Daitch–Mokotoff (S10-26, descoped), transliteration, Beider–Morse, a population name table |
| Support weighting and negative evidence | Concluded-value priority, memory of a rejected claim |
| Western profile, project-default pattern, catalog roles and weights | Per-person `name_format` claim, Spanish dual surnames, `patronymic`, `maternal_surname` |
| Date windows in the comparer, birth~baptism, place chains | Calendars and double dating |
| `depiction` + `likeness` gallery, bounding-box crop, one preferred portrait | Place/event keys, transparent polygon, Go PDF |
| Session camera | A saved camera table |
| Seeded `part_of` stays as it is | A new place-relationship term. Prefilling `subject` or `part_of` on a generic new connection (S10-10, descoped) |
| `gender` seeded as `man`, `woman`, `non_binary`, plus researcher terms | Treating `sex_at_birth` as gender. A locked term list. |
| `engagement` and `interests` on new catalogs | A migration that inserts either into catalogs that already exist |
| Unpromote: delete this claim and the bridge claims on this Source | Rejecting a claim, the weak-claim queue, re-pin, accept-as-is — still [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) |
| Sources evidence-status filter and text search; list scroll on the history entry | The same filters on People, Events, and Places. A separate session store for scroll |
| Select truncation and combo-box prefix rank | Omnibar ranking (that is S10-06). Sizing every select to its longest label |

## Gotchas

- **Bridge rows and plain rows are different types.** The order S10-D3 chooses has to stay true after reload, not only on the draft. Connections are grouped by bridge in `CitationConnections`; plain rows append in `CitationObservationRows`. One rule for both lists, and the saved order the composer shows has to match.
- **Focus is keyboard focus.** `focusedID` is set on add today and does not move keyboard focus into the property combo box. S10-D3 requires the first field you type to actually take focus.
- **`PVSelect` letter jump is one step.** `handleTriggerKey` sends `.character` once. Cycling is a session change for every select, not a name-modal special case. Width is the opposite: do not make every select as wide as its longest option.
- **Growing the canvas moves clamps.** Landed in S10-03. A small graph still opens on 4000×4000, and `maxCell` for that board stays 96×98. A card near the edge extends the document. Stored card coordinates stay valid.
- **Camera vs document size.** Landed in S10-04. Restoring an offset after the document grows clamps, and growth itself does not jump the view back to 1×.
- **Drop on the page vs drop in the dialog.** Both stage a file on the Add Artifact draft through `IngestMediaPolicy`. The page drop opens that dialog with the label empty. Confirm is what creates and ingests. A page drop must not create the artifact on release.
- **Open file on the graph.** A select lists every artifact that has a file. Open sends the selected file to the system app, the same path as the Source page. No files: both controls are absent. No in-app viewer window. S10-D6, in S10-20.
- **Audit scope fan-out.** An Observation on a shared member can touch one Source and one or more handles. Delete changes still need the parent FKs. Backfill is `INSERT OR IGNORE` in the migration, pattern `000036.sql`. `TestResolversCoverEveryEntityType` must stay green.
- **Last updated is not create time.** UUIDv7 id is create order. The Sources "Updated" sort already uses the scope. The new cell must use that revision, not the id.
- **Gender and sex at birth are different Properties.** Seeding gender does not relabel sex at birth. The page treatment is a reading of reconciled values, not a new truth write.
- **Outside click and macOS sheets.** AppKit sheets do not dismiss on an outside click. This is a product behavior on our presenters. Do not bolt a scrim onto `PVConfirm` (design-system rule: no dimming scrim; the window owns chrome).
- **Folding is not reconciliation.** A test that José and Jose become one auto-reconciled name is a bug. Matching suggestions may still rank them together.
- **`name_format_profiles` is in the catalog** (S10-29). Matching and display read the default profile's rows. A catalog with no default still uses the built-in Western pattern.
- **Depiction is not a file.** Delete and Impact stay on the Citation and the Artifact. The crop cache is disposable. A polygon edit must not leave the old crop as if it were the current face.
- **Gallery and gender share a page.** S10-23 ships the gender treatment inside frames that already reserve the gallery and the preferred-portrait control. S10-35 fills the gallery. S10-39 fills the control. Designing the gender page without those frames means a second Person-detail brief, which this plan does not allow.
- **Frequency on a tiny project.** Inverse frequency with one of each name is a flat weight. The scorer needs a floor so the first Smith is not treated as rare forever, and so an empty project still matches.
- **Do not backfill `engagement` or `interests`.** Seeded terms and properties are Install data. A migration `INSERT` would put them in catalogs the researcher already has, and this spike does not do that. Add them on those projects by hand.
- **List scroll is a field on the history entry.** `listScrollAnchor` is the id of the top row on Sources, People, Events, and Places. It is ignored for identity. It is written when you leave the list, not on each scroll. Showing the entry restores it, including relaunch. A visit with no anchor starts at the top. Sources sort and type filter clear it. S10-41's status and search clear it the same way when they land.
- **Status counts are primaries.** `sourcegraphprogress` counts every canvas subject, including bridges. The filter must not treat an unfiled bridge as "in progress" when every person, event, and place is filed. One extra grouped read, not a loop.
- **Unpromote is a delete of claims, not of subjects.** The bridge cards stay on the Evidence graph. Deleting the handle because it has no members left is a different product, and this spike does not have a handle delete. A rejected claim is not "unpromoted": the footer looks for an accepted claim, but a rejection means "not this entity" and stays parked with the review.
- **Truncation and the name-part width are different fixes.** S10-42 stops a long label from growing `PVSelect`. S10-17 stops the name-part control from changing width when the selection changes. Doing the second for every select makes short menus as wide as their longest option, which the earlier gotcha already forbids.

## Questions

| # | Question | Why it blocks |
| --- | --- | --- |
| Q1 | **Gender terms.** Seed `man`, `woman`, `non_binary`. Researchers extend the vocabulary. Answered 2026-10-08. | — |
| Q2 | **Dense Evidence graphs.** Descoped 2026-10-08. Stays in the dogfood log. Extract beyond date prefill and name split stays there too. | — |
| Q3 | **Sources status filters.** Empty / in progress / filed, counted on primary subjects in one grouped read. Text search filters the loaded list by title and ref. Sort, type, status, and query persist together on the Sources list (S10-41). Scroll is the history entry (S10-40). Answered 2026-10-11. Folded with S10-09 on 2026-10-12. | — |
| Q4 | **Unpromote.** Always confirm, and name the bridge claims on this Source that will be deleted with it. Delete the claims. Do not reject them, and do not delete the handle or the cards. Other members stay for the parked review. Answered 2026-10-11. | — |
| Q5 | **Long titles.** Truncate inside the select's given width. Help shows the full string. The name-part width fix stays S10-17. Answered 2026-10-11. | — |
| Q6 | **Combo rank.** Label prefix, then a later word, then any other contains match. Answered 2026-10-11. | — |
| Q7 | **List scroll.** The top row's id is a field on the history entry, ignored for identity, written when you leave the list. Back, Forward, a jump, and relaunch show that entry, so they restore it. A visit that does not pass an anchor starts at the top. Sources, People, Events, and Places all use it (S10-40). Changing sort or type filter clears it. Answered 2026-10-11. Revised 2026-10-14. | — |
| Q8 | **`interests`.** Unlocked text, cardinality multiple, person only, new catalogs only. Answered 2026-10-11. | — |
| Q9 | **Composer viewer.** No split. [#356](https://github.com/mendahu/provenencia/pull/356) already keeps a closed combo box from laying out every choice. A citation with 70 Observations opens quickly (2026-10-12). | — |

Everything else in **Working rules** is decided for planning. Correct any of them before the PR that implements it.

## Docs to update as work lands

| When | Doc |
| --- | --- |
| S10-21 | [`docs/audit-revision-history.md`](../../audit-revision-history.md) §4.1: `canonical_entity` is a scope. The catalog-audit rule's "conclusion layer carries no Source scope" stays true; the handle scope is additional. |
| S10-23 | None. The gender section is already [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.9 (landed with S10-38 and S10-45). Do not rewrite §3.7. |
| S10-25, S10-32 | Seeded vocabulary: abbreviation and nickname tables, `likeness` / `depiction`. |
| S10-29 | [`docs/matching.md`](../../matching.md) and [`docs/structured-name-model.md`](../../structured-name-model.md): the Western profile is installed; roles and weights are catalog rows. |
| S10-38 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4: `engagement`. Two subjects of any event type title as *{Type} of {A} and {B}*. No migration. |
| S10-40 | [`navigation-history.md`](../archive/spike-3/navigation-history.md): list scroll is a field on the entry, ignored for identity, written when you leave the list. |
| S10-44 | [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5.2 already allows deleting a claim. Note that Unpromote also deletes the bridge claims on that Source. Do not specify the review queue. |
| S10-45 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2 and §3.3: `interests`. No migration. |
| S10-99 | Pulled notes already left the dogfood log. Archive this folder, including the three idea notes, as decisions. |

## Definition of done

The dogfood bar at the top is true in the app, including the name and likeness lines. A person can mark one likeness to lead the page and the People list. A new catalog's event types include Engagement, and two people on one title as *Engagement of A and B*. A new catalog's Person properties include Interests. Each gated PR matches its brief. Parked notes and the ideas-folder leftovers are still parked. The weak-claim review is still an idea. Unpromote deletes a filing; it does not become that review. Existing catalogs were not given `engagement` or `interests` by a migration.
