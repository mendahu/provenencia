# Deployment Plan — Spike 10

Usability fixes from [`docs/dogfood/ux.md`](../../dogfood/ux.md) after Spike 9, plus three ideas pulled in from the parking lot: [depictions and likenesses](depictions-and-likenesses.md), [international names](international-names.md), and [name-matching enhancements](name-matching-enhancements.md). Each dogfood note or idea slice is one PR, unless two of them edit the same part of a view or the same comparer — those are combined below.

**A brief covers one view.** A view may have more than one brief. A PR that changes two views waits on a brief for each. Briefs are written just before the PR they gate ([`add-design-brief`](../../../.cursor/skills/add-design-brief/SKILL.md)), indexed in [`design/README.md`](design/README.md). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Slice 1 is done (2026-10-11). S10-01…S10-04 landed in [#361](https://github.com/mendahu/provenencia/pull/361)–[#364](https://github.com/mendahu/provenencia/pull/364). S10-05, S10-07, S10-08, S10-12, S10-13, S10-15, S10-42, and S10-43 landed in [#368](https://github.com/mendahu/provenencia/pull/368), [#369](https://github.com/mendahu/provenencia/pull/369), [#370](https://github.com/mendahu/provenencia/pull/370), [#371](https://github.com/mendahu/provenencia/pull/371), [#372](https://github.com/mendahu/provenencia/pull/372), [#375](https://github.com/mendahu/provenencia/pull/375), [#373](https://github.com/mendahu/provenencia/pull/373), and [#374](https://github.com/mendahu/provenencia/pull/374) (2026-10-12). S10-46, S10-47, and S10-48 were descoped in [#367](https://github.com/mendahu/provenencia/pull/367). S10-06 was descoped in [#368](https://github.com/mendahu/provenencia/pull/368): searching `marriage` already lists a title that contains the word ahead of a fuzzy name. S10-10 was descoped in [#370](https://github.com/mendahu/provenencia/pull/370): a common cluster belongs in [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md), which covers more than a default role. S10-11 was already done in [#356](https://github.com/mendahu/provenencia/pull/356): a citation with 70 Observations opens quickly. S10-14 and S10-D1 were descoped in [#370](https://github.com/mendahu/provenencia/pull/370): an adaptive choice field is not a priority for this spike. Later landings go in [`completed.md`](completed.md).

> **Goal of this spike:** the annoyances logged while entering real research are gone, and the app still behaves the way Spike 8 and Spike 9 decided.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Composer stays usable at 20 Observations.** Rows are easy to scan, a new Observation appears at the top and is focused, and Auto Transcribe is not sitting on the Citation menu.
2. **The Evidence graph holds a real Source.** The canvas grows before cards hit the edge. Pan and zoom survive a trip to the composer. The grid is a tile, so a zoomed-out or grown document does not turn the background into one giant bitmap. Zoomed-out edges stay legible. The card stays the full card at every zoom. Button hits match the buttons. A hotkey or toolbar control box-selects cards and moves them together.
3. **Files.** A transcription that overflows shows a scrollbar. A file dropped on the Artifact section uploads. The file opens from the collapsed row and from the Evidence graph.
4. **Lists.** Sources, People, Events, and Places show when the row last changed. Leaving a list and coming back lands at the same scroll position. The Sources list remembers sort, type filter, evidence-status filter, and text search.
5. **Entry defaults.** A date modal opens on Year and, when the transcription has a date, that date is already filled. A full name line can be split into parts. The name-part type control does not jump width and cycles same-letter matches.
6. **Small chrome.** Create Source asks for the name first. Enter saves a Property title. Clicking outside a sheet closes it, and asks first when the sheet is dirty. Opening an existing project selects the oldest user, not "create a new user." Searching `marriage` ranks title-word hits above fuzzy name prefixes. People, Places, and Events use one mark each, and the Event mark is not an hourglass. A Person page shows sex or gender, with gender winning when both exist.
7. **Names match the way records are written.** José matches Jose in suggestions and stays a separate grouping on the Person page. Jas. matches James. O Brien matches O'Brien. A rare surname outweighs Smith. A patronymic is a part type. Spanish dual surnames weigh the first surname above the second. A negative observation counts as a contradiction.
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
| Drag files onto Artifacts + open the file without expanding | S10-18 | S10-D5, S10-D6 |
| Rethink the Event icon + sidebar icons match Subject cards | S10-19 | S10-D7 |
| Box-select and move multiple Subject cards | S10-20 | S10-D6 (already designed for S10-18) |
| Last updated on list views | S10-21, S10-22, S10-41 | S10-D8…D11 |
| Show sex / gender on the Person detail page | S10-23 | S10-D12 |
| Accent folding + spaced particles | S10-24 | — |
| Abbreviations + nicknames | S10-25 | — |
| Phonetic tier (Daitch–Mokotoff) | S10-26 | — |
| Name frequency in this project | S10-27 | — |
| Support weighting + negative evidence | S10-28 | — |
| Name patterns: patronymic, profiles, Spanish | S10-29 | S10-D4 (part list, before S10-17) |
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
| Restore list scroll position on Back | S10-40, S10-41 | — |
| Sources list filters and text search | S10-41 | S10-D8 |
| Promote target menu truncates a long title | S10-42 | — |
| Combo box ranks a prefix above a mid-word match | S10-43 | — |
| Unpromote on a promoted Subject card | S10-44 | S10-D6 (same board as S10-18) |
| Seed Person property `interests` | S10-45 | — |

**Left in the dogfood log**

- **Dense Evidence graphs (census-scale).** Descoped 2026-10-08. Layers and filters stay in the dogfood log. Canvas growth landed (S10-03). Box-select is still S10-20. Neither hides part of a Source.
- **Automatic structure from the Artifact**, beyond the two slices in S10-13 and S10-17. No catalog writes from a detector. No OCR-on-PDF, Live Text, or Foundation Models.

**Back in the ideas folder** (not PRs)

- [Depictions leftovers](../../ideas/depictions-and-likenesses.md) — other keys, signature, transparent polygon, Windows PDF draw.
- [International-name leftovers](../../ideas/international-names.md) — other name patterns, transliteration, Beider–Morse, calendars and double dating, per-person `name_format`, a collation picker.
- [Name-matching leftovers](../../ideas/name-matching-enhancements.md) — population frequency, known-as, concluded values, calibration, researcher-defined weights, age-to-birth-year, rejected-claim memory, blocking.
- [Evidence graph structure shortcuts](../../ideas/evidence-graph-structure-shortcuts.md) — place a wedding, birth, or residence in one action, or paste a cluster already on the canvas. Replaces S10-10.

## Working rules

These are the calls for this spike. The questions at the bottom are answered; nothing in them still blocks a PR.

- **Add lands at the top for every Observation**, bridge or not. Today `CitationConnections` inserts a bridge at index 0 and `CitationObservationRows.addDraft` appends. Focus follows the new row's property search (the composer already stores `focusedID`; Add does not move keyboard focus into that combo box).
- **Auto Transcribe moves off the Citation menu's corner.** In the composer form, the Citation `PVSelect` and the Auto Transcribe button both sit on the trailing edge of the left column (`CitationComposerFormPane`), which is why the misclick happens. The confirm-before-overwrite stays.
- **An adaptive choice field is not in this spike.** S10-14 and S10-D1 are descoped (2026-10-12). Term fields keep the control they have. The composer rethink does not swap a short list to radios or a long list to a combo box.
- **A default role does not cover a repeated cluster.** S10-10 is descoped (2026-10-12). Prefilling `subject` or `part_of` on a generic new connection stays out. Placing a wedding, birth, or residence in one action, and pasting a cluster already on the canvas, are [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md).
- **Date prefill is silent.** First match in the transcription, current locale, Gregorian patterns a genealogist actually types (day month year, month day year, year alone, ISO). 80–90% is the bar. No "suggested" chrome, so the date modal does not need a brief. Year is focused even when the prefill is empty. Landed in S10-13.
- **Name split does not clobber edits.** Blur of the full-form line fills parts only when the part list is empty. The Split button refills from the line; if any part was edited by hand, confirm before replacing. Guessing rules are the ones in the dogfood note (comma → surname first; otherwise last word is surname; leftovers are given).
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
- **The Sources list is one page.** Sort, type filter, evidence status, and text search persist per project and user and restore together when that place is opened again, including after relaunch. People, Events, and Places have none of those controls. Last updated on a Sources row uses the revision Sources already store. That cell, the filters, the search field, the persisted settings, and the session scroll are S10-41, from S10-D8. They are not a persistence PR ahead of the brief.
- **Last updated** is an audit scope, not a timestamp column. Sources already store `UpdatedRevision` (`scope_type = source`) and can show it once the row is designed. People, Events, and Places have no scope today (`canonical_entity` is `noScope` in `core/database/audit/scopes.go`). S10-21 adds a `canonical_entity` scope: Identity Claim writes, and Observation writes on a Subject that is a member, so editing evidence moves the handle's revision. Display is relative or absolute in the briefs; the value is the revision's time.
- **One Event mark.** Sidebar People / Events / Places use SF Symbols (`person`, `calendar`, `mapPin` on `WorkspaceSection`). Cards, lists, omnibar hits, Promote, and Properties use `PVMark` (`subject_person`, `subject_event`, `subject_place`). The Event asset is the hourglass. S10-D7 draws the new Event mark and switches the sidebar (and the toolbar, which reads `WorkspaceSection.icon`) to the subject marks. Every existing `PVMark` call site picks up the new asset with no extra brief. List empty states that still use `PVSymbol` for these three kinds switch in the same PR.
- **Sex / gender styling** uses a seeded `gender` Property when the Person has one, otherwise `sex_at_birth`. The seed is `man`, `woman`, `non_binary` (labels Man, Woman, Non-binary). Researchers add further terms (`origin=user`), the same way they extend `role`. It is not product-locked, and there is no product `other` term. The page treatment recognizes the three seeded keys. A researcher-added term still counts as gender and wins over sex at birth; the brief draws that term without inventing a new color per key. Unknown, empty, and mixed (the reconciler already has a mixed state) stay neutral. `sex_at_birth` already appears as an ordinary field row when a record speaks to it (`PersonDetailContent`); this PR adds the page treatment and leaves that row in place.
- **Composer performance** is already done. [#356](https://github.com/mendahu/provenencia/pull/356) builds a combo box's option list only while its menu is open, so a form of closed pickers does not lay out every choice. Checked 2026-10-12: a citation with 70 Observations opens quickly. No separate profile pass, and no split that shows the form while the file is still loading. S10-16 still has to stay comfortable at 20 Observations; that bar is the layout.
- **List scroll is session state, like the graph camera.** Sources, People, Events, and Places remember content offset on the workspace session and restore it when that list place appears again (Back, Forward, breadcrumbs). It is not a navigation-history field ([`navigation-history.md`](../archive/spike-3/navigation-history.md) still omits scroll from the entry) and it does not survive relaunch. Sources scroll is restored in S10-41 after that PR has applied the persisted sort and filters, so the list is not put back on a differently ordered page. Changing sort or filter on that visit resets the offset to the top. People, Events, and Places have no sort; their scroll is S10-40.
- **Sources evidence filters count primaries in one read.** A primary is a person, event, or place. Empty: no primaries. In progress: at least one primary has no accepted Identity Claim. Filed: at least one primary, and every primary has an accepted claim. Bridges are not a filter axis. The counts are one grouped query beside `sourcegraphprogress`, not a query per row. Text search filters the list already loaded, by title and ref, as you type. It is not the omnibar and not FTS. Status and query persist in the same per-project, per-user store as sort and type filter.
- **Unpromote deletes the filing. It does not reject it.** The button is on a promoted primary card. Confirm always. The confirm names each bridge Subject on this Source (participation, relationship, location, place relationship) that has this Subject as an end and an accepted claim. One transaction deletes those claims and this Subject's claim, audited, through the write path. The Evidence cards stay; their footers show Promote again. The handle stays, even if it has no members left. Other members' claims are not rejected and not stripped. That review stays [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md). A rejection ("not this entity") is that review's action, not this button.
- **A long select label truncates inside the width it was given.** `PVSelect`'s closed label ellipsizes rather than growing the control. The Promote target column (`PromoteLayout.targetColumn`) stays put; `PromoteTargetMenuRow` already limits the menu title to one line. When the closed label is cut, help shows the full string. S10-17 still gives the name-part type control a stable width sized to its longest label. Those labels are short, so truncation does not replace that fix. Landed in S10-42.
- **Combo box rank is three bands, then the original order.** In `PVComboBoxMatch`, a label that starts with the query ranks above a later word that starts with it (split on spaces and hyphens), which ranks above any other contains match, including a match only in the subtext. `par` lists *parent* above *grandparent*. This is the kit filter, not `core/search` (S10-06). Landed in S10-43.
- **`interests` is prose, not a matcher feature.** Key `interests`, label Interests, `value_type = text`, cardinality `multiple`, bound to `person`, unlocked (the Properties page may unbind it). Same install rule as `engagement`: new catalogs only, no migration. It is not in a match profile and not on the Person header.
- **Artifact drop already exists inside the Add dialogs** (`IngestFileDropRow` on create and attach). It does not exist on the Artifact section of the Source page. S10-18 adds the section target. Open File today is only in the expanded detail (`SourcePageArtifactsView.artifactPrimaryFileColumn`). The Evidence graph header has no open-file control.
- **Name part types stay a select.** `patronymic` joins the compiled registry (S10-29), so the list can pass eight. It is still a fixed-width `PVSelect`. User-minted part types stay out. The Western split rules in S10-17 do not become profile-aware.
- **Accent fold is matching only.** `wordSimilarity` scores a fold-only match at 0.95. `autoreconcile.NormalizeForm` and `sort_key` stay as they are, so José and Jose remain separate groupings on the Person page. Display never folds.
- **Word tiers, top to bottom:** exact, abbreviation or nickname (just under exact), accent fold (0.95), the existing near-spelling score, then Daitch–Mokotoff. Spaced particles (`O`, `Mc`, `Mac`, `St`, `D`, `Fitz`, and the rest of that short list) rejoin with the next word before scoring. A lone initial (`O.`) does not rejoin.
- **Abbreviation and nickname tables are seeded vocabulary** a project can extend, same as `role`. Not a compiled-only list, and not a population file.
- **Name frequency is this project's inverse frequency** over name words. No seeded census of names.
- **Support weighting and negatives land together.** A value counts in proportion to its support, and a negative Observation is a contradiction. Reasons name the two values that scored. A concluded value does not get a special weight: Reconciliation Claims are not in this spike.
- **Patterns come from the catalog.** `name_format_profiles` is real (it is design-only today). Seed `western` and a Spanish dual-surname profile. `NamePatterns` reads those rows. Every name uses the project default profile until a per-person `name_format` claim exists. Spanish: both surnames are `surname`; the first weighs more than the second. No `maternal_surname` part type. `patronymic` is a part type and a role, compared as its own word, not walked to a father.
- **Collation is a project setting with no picker.** Default locale `en`. Name ordering uses `golang.org/x/text/collate`. Changing the locale in the UI waits for a settings place.
- **Dates in matching use the windows S9-21 already stores.** Spans and gaps are days. ABT doubles tolerance only when both dates are points. An event `date` can overlap `start_date` / `end_date`. Calendars and double dating stay out.
- **Birth and baptism partly match, and so do death and burial.** A small term-affinity table, same shape as name-role affinity. Place comparison walks `part_of` (already shipped) so York and Toronto can relate. Flat toponym text stays the strong match.
- **A depiction's value is its Citation.** New Observation `value_type = depiction`. No image column and no copied bytes. `likeness` is the seeded person Property of that type, cardinality many (a gallery). Place, event, and signature keys wait. Empty transcription is valid. The crop is a bounding-box JPEG, longest edge 512, cached by File plus locator, minted when the Observation is saved. The composer draws a PDF crop with PDFKit. A missing cache is filled the next time a client can open that Artifact. Two Citations with the same pixels may share bytes; the derivative rows stay separate. Promote's evidence sheet shows the two crops and pins the two Observations.
- **A preferred portrait is a display pointer, not a claim.** One likeness Observation on that person. The gallery still lists every crop. The Person page lead and the Persons list thumb use the chosen crop. With no choice, both use the earliest likeness. Deleting that likeness clears the pointer and the thumb falls back. The choice is made on the Person page. Promote does not pick a portrait.
- **`engagement` titles like a marriage.** Key `engagement`, label Engagement, next to `marriage` in `subjectvocab` `seedTerms`. New catalogs get it from Install. Existing catalogs are not migrated; the term is added there by hand. Two subjects use the couple rule: *Engagement of A and B*. Three or more stay *et al.*, the same as a marriage. The couple template takes the type word, so a marriage still reads *Marriage of A and B*.

## Design track

**Every UI PR is gated by Claude Design briefs. A brief covers one view; a view may have more than one.** A PR that touches two views waits on a brief for each. Briefs follow [`add-design-brief`](../../../.cursor/skills/add-design-brief/SKILL.md) and live in [`design/`](design/). **Each brief is designed alongside its feature**, just before the PR it points at.

| Brief | View | Gates | Later on the same view |
| --- | --- | --- | --- |
| **S10-D1** | Adaptive choice field (kit board) | **S10-14**, descoped 2026-10-12 | — |
| **S10-D3** | Citation composer | **S10-16** | — |
| **S10-D4** | Name parts modal | **S10-17** | Design after S10-29, so the type list includes `patronymic` |
| **S10-D5** | Source page — Artifacts | **S10-18** | — |
| **S10-D6** | Evidence graph | **S10-18** header open | S10-20 box-select. S10-44 Unpromote on the promoted card |
| **S10-D7** | Sidebar | **S10-19** | Toolbar section icon ships in the same PR |
| **S10-D8** | Sources list | **S10-41** | Last updated, evidence-status filter, text search, persisted sort and filters, session scroll |
| **S10-D9** | Persons list | **S10-22** | S10-35 fills the thumb with a likeness. S10-39 makes that thumb the preferred portrait |
| **S10-D10** | Events list | **S10-22** | — |
| **S10-D11** | Places list | **S10-22** | — |
| **S10-D12** | Person detail | **S10-23** | S10-35 gallery and S10-39 preferred portrait. Design both with the gender treatment, before S10-23 |
| **S10-D13** | Promote page | **S10-36** | Evidence sheet shows two crops |
| **S10-D14** | Evidence graph — zoomed-out card | **S10-46**, descoped 2026-10-11 | — |
| **S10-D15** | Citation composer — depiction row | **S10-34** | — |

S10-D6 is one board for the graph's new controls (open file, box-select, Unpromote). S10-01…S10-04 do not wait on it. S10-D14, the zoomed-out face, is descoped with S10-46 and is not drawn. S10-D1, the adaptive choice field, is descoped with S10-14 and is not drawn. Term fields keep the control they have. S10-D2 is not drawn: name-then-type on Create Source is a field swap, so S10-15 stands on its own. S10-D9…D11 are three briefs for the shared conclusion lists. Events and Places specify the same last-updated cell as Persons so `ConclusionListPage` does not grow three treatments. S10-D8 is the Sources list, a different page, and it gates S10-41. That board draws the same last-updated cell, the evidence-status filter, and text search. Sort, type, status, query, and session scroll ship from that one board. D9…D11 do not draw those controls, and S10-22 does not change the Sources list. Only the Persons list thumb becomes a likeness. S10-D6, S10-D9, and S10-D12 each cover a later PR on that same view, so those briefs are written with the later frame in them. The depiction row is its own brief, S10-D15, written just before S10-34. S10-D3 does not draw it. D12 includes the preferred-portrait control S10-39 fills, and D9 includes that crop as the list thumb. Whichever of D8 and D9 is designed first draws the last-updated cell; the other copies it.

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
  S10-29  Name patterns (patronymic, western + Spanish profiles) — before D4
  ✎ S10-D4 ──▶ S10-17  Name parts: split, fixed width, letter cycle
  Check: add an Observation — it is at the top and focused. Auto Transcribe
         is not beside the Citation menu. "Smith, John"
         splits to Surname, Given. Patronymic is in the type list. Pressing S
         twice reaches Surname.

SLICE 5 — Artifacts and multi-select
  ✎ S10-D5 ──▶
  ✎ S10-D6 ──▶ S10-18  Drop on the Artifact section; open from the row and the graph
  S10-19  Sidebar and Event mark (✎ S10-D7). Can start as soon as D7 is designed;
          it does not wait on S10-18.
  ✎ (S10-D6 already done) ──▶ S10-20  Box-select (after S10-01 and S10-03)
  Check: drop a file on the section and it uploads. Open the file without
         expanding. Draw a box, drag once, the group moves. Sidebar Event
         matches the card.

SLICE 6 — Lists and the Person page
  S10-40  People, Events, and Places scroll restored when the list place appears
  S10-21  canonical_entity audit scope + list revision reads
  ✎ S10-D9…D11 ──▶ S10-22  Last updated on People, Events, and Places
  ✎ S10-D8 ──▶ S10-41  Sources list: last updated, filters, search,
                       persisted sort and filters, session scroll
  ✎ S10-D12 ──▶ S10-23  Person page sex / gender. The board also draws the gallery S10-35 fills and the preferred portrait S10-39 fills.
  Check: edit a Source, leave, come back — the row's time moved, the list is
         scrolled where you left it, and sort, type, status, and search are
         still set. Quit and relaunch: Sources scroll starts at the top, and
         those settings are still set. Edit a member's Observation — the Person
         row's time moved. Filter to graphs that still have an unpromoted person:
         a filed Source drops out. A Person with gender shows gender; one with
         only sex at birth shows that; unknown is neutral.

SLICE 7 — How names compare (no new chrome)
  S10-24  Accent fold (0.95) and spaced particles
  S10-25  Abbreviation and nickname tables
  S10-26  Daitch–Mokotoff below near-spelling
  S10-27  Project name-frequency weights
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
  ✎ (S10-D6 already done) ──▶ S10-44  Unpromote on the promoted card
  Check: confirm names the bridges. The card shows Promote again. The other
         person on that relationship stays promoted. The handle is still there.

S10-99 — Dogfood close / docs
```

S10-05…S10-13, S10-15, S10-19, S10-42, and S10-43 have no dependency on the graph slice. S10-24, S10-29, S10-30, S10-31, S10-32, S10-38, and S10-45 have no dependency on the dogfood slices. They are ordered here so review stays one stream. Parallel PRs are fine when **Depends on** allows it. S10-29 is numbered with the name work and listed in slice 4 because the name-parts brief cannot be drawn until `patronymic` exists.

### Dependencies at a glance

```text
S10-01 hit boxes (done) ────────┐
S10-02 zoom, tiled grid (done) ─┤
S10-03 canvas grow (done) ──────┼──▶ S10-20 box-select
S10-04 camera (done)            │         ▲
                                │         │
S10-10 defaults — descoped      │    S10-D6 (also gates S10-18)
S10-11 perf — done, #356
S10-14 choice — descoped. S10-D1 is not drawn.
S10-D3 ──▶ S10-16 composer rethink

S10-46 simplified cards — descoped 2026-10-11
S10-47 mount cards near the window — descoped 2026-10-11
S10-48 canvas cover — descoped 2026-10-11

S10-40 list scroll — People, Events, and Places. Sources scroll is S10-41.
S10-21 scope ──▶ S10-22 last updated (D9…D11)
S10-D8 ──▶ S10-41 Sources list (last updated, filters, persist, scroll)
S10-D6 ──▶ S10-44 unpromote (button drawn with the open-file / box-select board)
S10-D12 (gender + gallery + preferred portrait) ──▶ S10-23
S10-29 patterns ──▶ S10-D4 ──▶ S10-17 name parts
                 └──▶ S10-37 collation

S10-24 fold ──▶ S10-25 tables ──▶ S10-26 phonetic
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

**Brief:** S10-D3.

| | |
| --- | --- |
| **In** | Per S10-D3: Observations are visually separate, and the fields inside a row are easy to tell apart. Add (bridge or not) inserts at the top and moves keyboard focus to that row's property search. Auto Transcribe (and the PDF paste control, which sits in the same slot) moves off the Citation menu's corner. Term fields keep the control they have (S10-14 is descoped). Still comfortable at 20 Observations (S10-11's bar). |
| **Out** | Name parts modal (S10-17). Date modal (S10-13). The depiction row (S10-34, brief S10-D15). An adaptive choice field. |
| **Testable** | Twenty rows: you can see where one ends. Add a bridge and a plain Observation: both appear at the top, search focused. Reaching for the Citation menu does not start Auto Transcribe. |
| **Depends on** | S10-D3 |

#### S10-17 — Name parts split and type control

**Brief:** S10-D4, designed after S10-29 so the type list includes `patronymic`.

| | |
| --- | --- |
| **In** | Per S10-D4 and the working rules: blur splits an empty part list; Split refills, confirming when a part was hand-edited. Type control has a stable width (sized for the longest label, not the selection) and stays a `PVSelect`. Same-letter keypress cycles matches. That cycle is fixed in `PVSelect` (`handleTriggerKey` / session), so every select gains it; the width fix stays on this field. Split rules stay the Western ones. |
| **Out** | Profile-aware splitting. A combo box for part type. User-minted part types. |
| **Testable** | `Smith, John Henry` → Surname, Given, Given. `John Henry Smith` → Given, Given, Surname. Patronymic is a choice and does not resize the control. Blur with existing hand-edited parts does nothing. Split then asks. Pressing S moves from Surname prefix to Surname. |
| **Depends on** | S10-29, S10-D4 |

### Slice 5 — Artifacts, marks, multi-select

#### S10-18 — Artifact drop and open

**Briefs:** S10-D5 (Source page Artifacts) and S10-D6 (Evidence graph — header open, and box-select specified for S10-20).

| | |
| --- | --- |
| **In** | Drop a file on the Source page Artifact section to upload, with a drop highlight. The collapsed artifact row opens the file (the expanded column keeps its button or the brief says it moves). The Evidence graph header offers the same open for the Source's file (cover, else the primary file the brief names). Reuse the existing ingest path (`applyDroppedURLs` / `artifacts.open`), not a second uploader. |
| **Out** | Box-select (S10-20). Drop onto cards. A new Artifact model. |
| **Testable** | Drop a PNG on the section: highlight, then the file is on the Source. Collapsed row opens it in one click. Graph header opens the same file. A fileless Source does not show a dead open button. |
| **Depends on** | S10-D5, S10-D6 |

#### S10-19 — Event mark and sidebar icons

**Brief:** S10-D7.

| | |
| --- | --- |
| **In** | A new `subject_event` mark (calendar / datebook — the board draws it). Sidebar People, Events, and Places use `PVMark` subject marks, not `person` / `calendar` / `mapPin`. The toolbar section icon uses that same source. List empty states for those three kinds use the mark. Person and Place marks stay. Cards, omnibar, Promote, and Properties already use `PVMark` and update with the asset. |
| **Out** | A redraw of Person or Place marks. New sidebar rows. |
| **Testable** | Sidebar, toolbar, an Event card, the Events list empty state, and an omnibar Event hit show the same Event mark. People and Places match between sidebar and card. |
| **Depends on** | S10-D7 |

#### S10-20 — Box-select

**Brief:** S10-D6, already designed before S10-18.

| | |
| --- | --- |
| **In** | Hold the hotkey the brief names, then drag a rectangle. Every Subject card inside is selected. Dragging one moves the group; positions persist the way a single drag does today. A toolbar button arms the same mode and shows the hotkey. Escape or a second click of the button leaves the mode. |
| **Out** | Selecting bridge cards, unless the brief includes them (default: Subject cards only, matching the note). An undo stack. |
| **Testable** | Box three cards, drag one, all three land together and survive reload. The toolbar button works without the hotkey and its label or tooltip shows the key. |
| **Depends on** | S10-01, S10-03, S10-D6 |

### Slice 6 — Lists and the Person page

#### S10-40 — List scroll position

| | |
| --- | --- |
| **In** | Remember the scroll offset of the People, Events, and Places lists on the workspace session. Restore it when that list place appears again, including Back and Forward. They share `ConclusionListPage`; each list keeps its own offset. |
| **Out** | The Sources list (S10-41 restores that offset after sort and filters). A navigation-history field. Restoring across relaunch. |
| **Testable** | Scroll to the bottom of People, open a Person, press Back: the list is at the bottom. Quit and relaunch: the offset starts at the top. The same round-trip on Events and Places. |
| **Depends on** | — |

#### S10-21 — Handle last-activity scope

| | |
| --- | --- |
| **In** | New audit scope `canonical_entity` ([`add-audit-scope`](../../../.cursor/skills/add-audit-scope/SKILL.md)). Identity Claim writes scope to the handle. An Observation write on a member Subject also scopes to that handle, and still scopes to the Source. List reads for Persons, Events, and Places return the latest revision time the way Sources already return `UpdatedRevision`. Backfill existing history in the next migration. Resolver tests; no migration test. |
| **Out** | The row UI (S10-22). A `updated_at` column on `canonical_entities`. |
| **Testable** | Promote a Person: the handle has a revision. Edit a member Observation: that handle's revision moves, and the Source's revision still moves. A handle with no claims does not appear. Rebuild-style check: scope rows match the writes, including a delete of a member. |
| **Depends on** | — |

#### S10-22 — Last updated on People, Events, and Places

**Briefs:** S10-D9 Persons, S10-D10 Events, S10-D11 Places. Design the three before the PR. Events and Places copy the Persons cell. The Sources list is a different page; its last-updated cell ships in S10-41 from S10-D8, using the same treatment.

| | |
| --- | --- |
| **In** | Each of these three list rows shows when it last changed, from the revision S10-21 returns. Empty or never-audited reads as the brief's empty. This PR does not add sort to these lists. |
| **Out** | The Sources list (S10-41). A new list layout beyond the cell the briefs add. |
| **Testable** | People, Events, and Places show a time. Edit a member Observation: the Person row moves. Events and Places match the Persons cell. |
| **Depends on** | S10-21, S10-D9, S10-D10, S10-D11 |

#### S10-41 — Sources list

**Brief:** S10-D8, just before this PR. One board for the page. It does not add a cell, a filter, or a search field to Events, Places, or Persons.

| | |
| --- | --- |
| **In** | Per the working rules and S10-D8. The row shows when the Source last changed, from the existing `UpdatedRevision`. Filter the list to empty, in progress, or filed, from one grouped count of primary subjects and their accepted claims, next to `sourcegraphprogress`. A text field filters the loaded rows by title and ref as you type. Sort, type filter, status, and query persist per project and user, and restore when the Sources place is shown again, including after relaunch (this is S10-09). Remember the list's scroll offset on the workspace session and restore it on the next appearance, after those settings are applied, including Back and Forward. Changing sort or filter on that visit resets the offset to the top. Scroll does not survive relaunch and is not a navigation-history field. |
| **Out** | The same filters, search, or last-updated work on People, Events, or Places (scroll there is S10-40; last updated there is S10-22). Omnibar search. A per-row count query. Hiding bridge subjects as their own status. |
| **Testable** | A Source with no people, events, or places matches Empty and not the other two. A Source with one unpromoted person matches In progress. After that person is promoted, it matches Filed. Typing part of a title hides the other rows. Leave and come back: status, query, sort, and type filter are still set, the list is scrolled where you left it, and the row's time moved if the Source changed. Relaunch: those four settings are still set, and the offset starts at the top. A long list does not issue one count query per row. |
| **Depends on** | S10-D8 |

#### S10-23 — Sex / gender on the Person page

**Brief:** S10-D12.

| | |
| --- | --- |
| **In** | Seed a `gender` Property on `person`, separate from `sex_at_birth` ([`add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md), new § in `seeded-vocabulary.md`). Product terms: `man`, `woman`, `non_binary`. Researchers may add `origin=user` terms; Create is allowed (not the `place_relationship_type` lock). Per S10-D12: the page treatment follows gender when present, else sex at birth. The three seeded keys are the ones the treatment recognizes. Any other gender term still wins over sex at birth and uses the brief's extra-term treatment. Unknown, empty, and mixed are the neutral treatment. The existing field row for sex at birth stays. |
| **Out** | The gallery (S10-35) and the preferred portrait (S10-39). This brief already draws both frames. A gender field on Event or Place. Editing claims. Using sex at birth as gender identity. |
| **Testable** | Person with only sex at birth: that treatment. Person with gender: gender wins even if sex at birth disagrees. Mixed gender: neutral, not a blend of two colors. A Person with neither: unchanged page. |
| **Depends on** | S10-D12 |

### Slice 7 — How names compare

#### S10-24 — Accent fold and spaced particles

| | |
| --- | --- |
| **In** | In `core/match` word comparison: NFD, drop combining marks, plus the small table (ß, æ, œ, ø, ł, đ, þ, ı). A fold-only match scores 0.95, under an exact match. Rejoin a spaced particle in the working-rules list with the following word. An initial with a period does not rejoin. Display, `NormalizeForm`, and `sort_key` stay unchanged. |
| **Out** | Auto-reconciler merging of accent variants. Phonetics (S10-26). Tables (S10-25). |
| **Testable** | José vs Jose scores 0.95 and José vs José scores 1. The Person page still shows them as separate groupings. `O Brien` matches `O'Brien`. `O.` does not swallow the next word. |
| **Depends on** | — |

#### S10-25 — Abbreviations and nicknames

| | |
| --- | --- |
| **In** | Seeded, project-extensible tables for historical abbreviations (Jas., Wm., Chas., Thos., Geo., Eliz., and the rest of a short product set) and nicknames (Jim/James, and a short product set). A table hit scores just under an exact match, above the 0.95 fold. Edit distance is not how these are learned. |
| **Out** | A UI for editing the tables beyond the vocabulary pattern the other term lists already use. Phonetics. |
| **Testable** | Jas. vs James scores just under exact. Jim vs James does too. A researcher-added nickname participates. An unknown token still falls through to fold and edit distance. |
| **Depends on** | S10-24 |

#### S10-26 — Phonetic tier

| | |
| --- | --- |
| **In** | Daitch–Mokotoff as a `wordSimilarity` tier below near-spelling. Leigh/Lee can meet there. Smyth/Smith still meets by edit distance and does not need this tier to score. |
| **Out** | Beider–Morse. Transliteration across scripts. |
| **Testable** | A phonetic-only pair scores above 0 and below a one-edit near-spelling. An exact match still outranks it. |
| **Depends on** | S10-25 |

#### S10-27 — Name frequency

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

| | |
| --- | --- |
| **In** | `patronymic` in the compiled part-type registry ([`add-name-value`](../../../.cursor/skills/add-name-value/SKILL.md)). Migration for `name_format_profiles`, `name_format_profile_parts` (natural and sorted), and `project_settings` if that table is not already there. Seed `western` and a Spanish dual-surname profile (first surname weighs more than the second; both are `surname`). `NamePatterns` reads the catalog. Loaders set `Value.NamePattern` from the project default. Role for `patronymic` is its own, not family. |
| **Out** | A per-person `name_format` Reconciliation Claim. `maternal_surname`. Walking a patronymic to the father's handle. Profile-driven split in the name modal (S10-17 keeps the Western rules). |
| **Testable** | García Márquez vs García weighs the shared first surname above a maternal mismatch. A patronymic pairs with a patronymic. Two names with no stored pattern still use western. The registry test that every product part type has a role includes `patronymic`. |
| **Depends on** | — |

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
| **In** | `value_type = depiction`. The value is the Observation's own Citation (artifact, page, region). No image column. Seed `likeness` on `person`, cardinality many. Empty transcription stays valid. Install path only ([`add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md)). |
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
| **In** | Per S10-D15: pick the region, leave transcription empty, save a likeness (or another depiction property, if the board shows only likeness). The row uses the same add-at-top and focus rules as S10-16. |
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
| **In** | Add `engagement` / Engagement to `subjectvocab` `seedTerms` for `event_type`, beside `marriage`. Install on catalog create only ([`add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md)). [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4 lists the key. `eventtitle.Choose` uses the couple rule for `engagement` as well as `marriage` when there are exactly two subjects. The couple template is *{Type} of {A} and {B}* in `EventTitleDisplay` and in the search-index title (`handles.go` currently hardcodes "Marriage of"), so a marriage still reads *Marriage of A and B* and an engagement reads *Engagement of A and B*. Follow [`add-localized-string`](../../../.cursor/skills/add-localized-string/SKILL.md) for the template. No new chrome, so no brief. |
| **Out** | A migration that inserts the term into existing catalogs. Those projects get the term by hand. Matching affinity with marriage. A new role. Three or more subjects (still *et al.*). |
| **Testable** | A catalog created after this PR offers Engagement. Two subjects title *Engagement of A and B* on the list, the page, and in search. A marriage of two still titles *Marriage of A and B*. Three subjects of an engagement title *Engagement of A et al.* Opening an older catalog does not add the term. |
| **Depends on** | — |

#### S10-45 — Seed `interests` on Person

| | |
| --- | --- |
| **In** | Add `interests` / Interests to the person property seed: `value_type = text`, cardinality `multiple`, bound to `person`, unlocked ([`add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md)). Install on catalog create only. [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2 and §3.3 list the key. No new chrome, so no brief. The composer and the Person page show it the way they show any other text property. |
| **Out** | A migration that inserts the property into existing catalogs. A match feature. A header line. A term list. Locking the binding. |
| **Testable** | A catalog created after this PR offers Interests on a Person, accepts more than one value, and lets the Properties page unbind it. Opening an older catalog does not add the property. |
| **Depends on** | — |

### Slice 10 — Take a filing back

#### S10-44 — Unpromote

**Brief:** S10-D6, already designed before S10-18. This PR fills the button that board drew on the promoted card.

| | |
| --- | --- |
| **In** | Per the working rules and S10-D6: a promoted primary card offers Unpromote beside the control that opens the handle. Confirm always, with the existing confirm. The body names each bridge Subject on this Source whose accepted claim has this Subject as an end. Confirm deletes this Subject's Identity Claim and those bridge claims in one `writes.Run` transaction, audited. Cards stay on the graph and show Promote. The handle is not deleted. Other members' claims are left for the parked review. |
| **Out** | Rejecting the claim ("not this entity"). The weak-claim queue, re-pin, and accept-as-is ([`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md)). Deleting the Evidence cards. Deleting the handle. Unpromote from the Person, Event, or Place page. |
| **Testable** | Unpromote a person who is in a filed relationship: the confirm names that relationship, both cards show Promote, the other person stays on their handle, and that handle still exists. Unpromote a person with no filed bridges: the confirm does not invent a bridge, and the card shows Promote. Cancel leaves the claim. |
| **Depends on** | S10-D6 |

### S10-99 — Dogfood close / docs

| | |
| --- | --- |
| **In** | Pulled notes already left [`docs/dogfood/ux.md`](../../dogfood/ux.md). Do not copy them back. Distill this spike with [`archive-docs`](../../../.cursor/skills/archive-docs/SKILL.md) when the bar above is true. The three idea notes live in this folder; the close distill keeps their decisions and drops the PR sequence. |
| **Out** | Pulling the parked density, extract, transliteration, or Reconciliation-Claim items in at the close. |
| **Depends on** | S10-01…S10-45 |

## Scope boundary

| In | Out |
| --- | --- |
| The dogfood notes in the table above | Census-scale layers / filters |
| Date prefill and name split as the cheap extract | Detector suggestions, guided extract, catalog writes from a parse |
| Accent fold, abbreviations, nicknames, Daitch–Mokotoff, project name frequency | Transliteration, Beider–Morse, a population name table |
| Support weighting and negative evidence | Concluded-value priority, memory of a rejected claim |
| `patronymic`, western + Spanish profiles, project-default pattern | Per-person `name_format` claim, `maternal_surname` |
| Date windows in the comparer, birth~baptism, place chains | Calendars and double dating |
| `depiction` + `likeness` gallery, bounding-box crop, one preferred portrait | Place/event keys, transparent polygon, Go PDF |
| Session camera | A saved camera table |
| Seeded `part_of` stays as it is | A new place-relationship term. Prefilling `subject` or `part_of` on a generic new connection (S10-10, descoped) |
| `gender` seeded as `man`, `woman`, `non_binary`, plus researcher terms | Treating `sex_at_birth` as gender. A locked term list. |
| `engagement` and `interests` on new catalogs | A migration that inserts either into catalogs that already exist |
| Unpromote: delete this claim and the bridge claims on this Source | Rejecting a claim, the weak-claim queue, re-pin, accept-as-is — still [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) |
| Sources evidence-status filter and text search; list scroll for the session | The same filters on People, Events, and Places. Scroll stored in navigation history or across relaunch |
| Select truncation and combo-box prefix rank | Omnibar ranking (that is S10-06). Sizing every select to its longest label |

## Gotchas

- **Bridge rows and plain rows are different types.** Add-at-top has to stay true after reload, not only on the draft. Connections are grouped by bridge in `CitationConnections`; plain rows append in `CitationObservationRows`. One insert rule, both lists, and the saved order the composer shows has to match.
- **Focus and the new row.** `focusedID` is set on add today and does not move keyboard focus into the property combo box. The rethink wants both: row at top, field focused.
- **`PVSelect` letter jump is one step.** `handleTriggerKey` sends `.character` once. Cycling is a session change for every select, not a name-modal special case. Width is the opposite: do not make every select as wide as its longest option.
- **Growing the canvas moves clamps.** Landed in S10-03. A small graph still opens on 4000×4000, and `maxCell` for that board stays 96×98. A card near the edge extends the document. Stored card coordinates stay valid.
- **Camera vs document size.** Landed in S10-04. Restoring an offset after the document grows clamps, and growth itself does not jump the view back to 1×.
- **Drop on the section vs drop in the dialog.** The dialog row already highlights and ingests. The section target should call that ingest, not a parallel copy, including the reject callout for a bad type.
- **Open file on the graph.** A Source can have several Artifacts. The brief picks which file the header opens (cover if set, otherwise the single file, otherwise no button or a menu — D6 decides; do not invent a menu in the PR if the brief says one file).
- **Audit scope fan-out.** An Observation on a shared member can touch one Source and one or more handles. Delete changes still need the parent FKs. Backfill is `INSERT OR IGNORE` in the migration, pattern `000036.sql`. `TestResolversCoverEveryEntityType` must stay green.
- **Last updated is not create time.** UUIDv7 id is create order. The Sources "Updated" sort already uses the scope. The new cell must use that revision, not the id.
- **Gender and sex at birth are different Properties.** Seeding gender does not relabel sex at birth. The page treatment is a reading of reconciled values, not a new truth write.
- **Outside click and macOS sheets.** AppKit sheets do not dismiss on an outside click. This is a product behavior on our presenters. Do not bolt a scrim onto `PVConfirm` (design-system rule: no dimming scrim; the window owns chrome).
- **Folding is not reconciliation.** A test that José and Jose become one auto-reconciled name is a bug. Matching suggestions may still rank them together.
- **`name_format_profiles` is not in the catalog yet.** S10-29 creates it. Do not point `NamePatterns` at a table that only exists in [`structured-name-model.md`](../../structured-name-model.md) §4.
- **Depiction is not a file.** Delete and Impact stay on the Citation and the Artifact. The crop cache is disposable. A polygon edit must not leave the old crop as if it were the current face.
- **Gallery and gender share a page.** S10-23 ships the gender treatment inside frames that already reserve the gallery and the preferred-portrait control. S10-35 fills the gallery. S10-39 fills the control. Designing the gender page without those frames means a second Person-detail brief, which this plan does not allow.
- **Frequency on a tiny project.** Inverse frequency with one of each name is a flat weight. The scorer needs a floor so the first Smith is not treated as rare forever, and so an empty project still matches.
- **Do not backfill `engagement` or `interests`.** Seeded terms and properties are Install data. A migration `INSERT` would put them in catalogs the researcher already has, and this spike does not do that. Add them on those projects by hand.
- **Scroll is not a history entry.** Spike 3 omits scroll from `WorkspaceLocation`. S10-40 and S10-41 keep it that way and store the offset on the session, the way S10-04 stores the camera. Sources scroll is restored in the same PR that restores sort and filters, so the list is not put back at the bottom of a differently ordered list.
- **Status counts are primaries.** `sourcegraphprogress` counts every canvas subject, including bridges. The filter must not treat an unfiled bridge as "in progress" when every person, event, and place is filed. One extra grouped read, not a loop.
- **Unpromote is a delete of claims, not of subjects.** The bridge cards stay on the Evidence graph. Deleting the handle because it has no members left is a different product, and this spike does not have a handle delete. A rejected claim is not "unpromoted": the footer looks for an accepted claim, but a rejection means "not this entity" and stays parked with the review.
- **Truncation and the name-part width are different fixes.** S10-42 stops a long label from growing `PVSelect`. S10-17 stops the name-part control from changing width when the selection changes. Doing the second for every select makes short menus as wide as their longest option, which the earlier gotcha already forbids.

## Questions

| # | Question | Why it blocks |
| --- | --- | --- |
| Q1 | **Gender terms.** Seed `man`, `woman`, `non_binary`. Researchers extend the vocabulary. Answered 2026-10-08. | — |
| Q2 | **Dense Evidence graphs.** Descoped 2026-10-08. Stays in the dogfood log. Extract beyond date prefill and name split stays there too. | — |
| Q3 | **Sources status filters.** Empty / in progress / filed, counted on primary subjects in one grouped read. Text search filters the loaded list by title and ref. Sort, type, status, and query persist together on the Sources list (S10-41), with that list's scroll. Answered 2026-10-11. Folded with S10-09 on 2026-10-12. | — |
| Q4 | **Unpromote.** Always confirm, and name the bridge claims on this Source that will be deleted with it. Delete the claims. Do not reject them, and do not delete the handle or the cards. Other members stay for the parked review. Answered 2026-10-11. | — |
| Q5 | **Long titles.** Truncate inside the select's given width. Help shows the full string. The name-part width fix stays S10-17. Answered 2026-10-11. | — |
| Q6 | **Combo rank.** Label prefix, then a later word, then any other contains match. Answered 2026-10-11. | — |
| Q7 | **List scroll.** Session offset, restored when the list place appears. Not a history field. Not across relaunch. Sources scroll ships with the Sources list (S10-41). People, Events, and Places are S10-40. Answered 2026-10-11. | — |
| Q8 | **`interests`.** Unlocked text, cardinality multiple, person only, new catalogs only. Answered 2026-10-11. | — |
| Q9 | **Composer viewer.** No split. [#356](https://github.com/mendahu/provenencia/pull/356) already keeps a closed combo box from laying out every choice. A citation with 70 Observations opens quickly (2026-10-12). | — |

Everything else in **Working rules** is decided for planning. Correct any of them before the PR that implements it.

## Docs to update as work lands

| When | Doc |
| --- | --- |
| S10-21 | [`docs/audit-revision-history.md`](../../audit-revision-history.md) §4.1: `canonical_entity` is a scope. The catalog-audit rule's "conclusion layer carries no Source scope" stays true; the handle scope is additional. |
| S10-23 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) new gender section. Do not rewrite §3.7. |
| S10-25, S10-29, S10-32 | Seeded vocabulary: abbreviation and nickname tables, `patronymic` and the Spanish profile, `likeness` / `depiction`. [`docs/matching.md`](../../matching.md) and [`docs/structured-name-model.md`](../../structured-name-model.md) follow the code. |
| S10-38 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4: `engagement`. Couple rule covers `engagement`. No migration. |
| S10-40, S10-41 | [`navigation-history.md`](../archive/spike-3/navigation-history.md): list scroll is session state restored when the list place appears. It is still not a history field. |
| S10-44 | [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5.2 already allows deleting a claim. Note that Unpromote also deletes the bridge claims on that Source. Do not specify the review queue. |
| S10-45 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.2 and §3.3: `interests`. No migration. |
| S10-99 | Pulled notes already left the dogfood log. Archive this folder, including the three idea notes, as decisions. |

## Definition of done

The dogfood bar at the top is true in the app, including the name and likeness lines. A person can mark one likeness to lead the page and the People list. A new catalog's event types include Engagement, and two people on one title as *Engagement of A and B*. A new catalog's Person properties include Interests. Each gated PR matches its brief. Parked notes and the ideas-folder leftovers are still parked. The weak-claim review is still an idea. Unpromote deletes a filing; it does not become that review. Existing catalogs were not given `engagement` or `interests` by a migration.
