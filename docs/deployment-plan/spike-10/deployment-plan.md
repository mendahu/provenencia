# Deployment Plan — Spike 10

Usability fixes from [`docs/dogfood/ux.md`](../../dogfood/ux.md) after Spike 9, plus three ideas pulled in from the parking lot: [depictions and likenesses](depictions-and-likenesses.md), [international names](international-names.md), and [name-matching enhancements](name-matching-enhancements.md). Each dogfood note or idea slice is one PR, unless two of them edit the same part of a view or the same comparer — those are combined below.

**One brief per view.** A UI PR that changes two views waits on two briefs. Briefs are written just before the PR they gate ([`add-design-brief`](../../../.cursor/skills/add-design-brief/SKILL.md)), indexed in [`design/README.md`](design/README.md). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Nothing landed. Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** the annoyances logged while entering real research are gone, and the app still behaves the way Spike 8 and Spike 9 decided.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Composer stays usable at 20 Observations.** Rows are easy to scan, a new Observation appears at the top and is focused, and Auto Transcribe is not sitting on the Citation menu.
2. **The Evidence graph holds a real Source.** The canvas grows before cards hit the edge. Pan and zoom survive a trip to the composer. Zoomed-out grid and edges stay legible. Button hits match the buttons. A hotkey or toolbar control box-selects cards and moves them together.
3. **Files.** A transcription that overflows shows a scrollbar. A file dropped on the Artifact section uploads. The file opens from the collapsed row and from the Evidence graph.
4. **Lists.** Sources, People, Events, and Places show when the row last changed. The Sources list remembers sort and filter.
5. **Entry defaults.** Person→Event role prefills `subject`. Place→Place relationship prefills `part_of`. A date modal opens on Year and, when the transcription has a date, that date is already filled. A full name line can be split into parts. The name-part type control does not jump width and cycles same-letter matches.
6. **Small chrome.** Create Source asks for the name first. Enter saves a Property title. Clicking outside a sheet closes it, and asks first when the sheet is dirty. Opening an existing project selects the oldest user, not "create a new user." Searching `marriage` ranks title-word hits above fuzzy name prefixes. People, Places, and Events use one mark each, and the Event mark is not an hourglass. A Person page shows sex or gender, with gender winning when both exist.
7. **Names match the way records are written.** José matches Jose in suggestions and stays a separate grouping on the Person page. Jas. matches James. O Brien matches O'Brien. A rare surname outweighs Smith. A patronymic is a part type. Spanish dual surnames weigh the first surname above the second. A negative observation counts as a contradiction.
8. **A face is evidence.** A polygon on a photo is a depiction Observation. The Person page shows those crops, and one can be marked to lead that page and the People list. Promote compares two likenesses as pictures. Empty transcription is valid.

## Where the notes went

| Dogfood note | PR | Brief |
| --- | --- | --- |
| Evidence graph button hit boxes | S10-01 | — |
| Evidence graph misrenders when zoomed far out | S10-02 | — |
| Evidence graph canvas is too small | S10-03 | — |
| Evidence graph viewport resets after the composer | S10-04 | — |
| Multi-line text areas need a scrollbar | S10-05 | — |
| Omnibar ranks fuzzy name matches above exact title words | S10-06 | — |
| Onboarding defaults to a new user | S10-07 | — |
| Enter should save a Property title edit | S10-08 | — |
| Sources list sort and filter don't persist | S10-09 | — |
| Prefill sensible defaults for common Observation kinds | S10-10 | — |
| Citation composer slows down with many Observations | S10-11 | — |
| Click outside a modal sheet to close it | S10-12 | — |
| Prefill the structured date modal + open focused on Year | S10-13 | — |
| Adaptive choice field | S10-14, used by S10-16 | S10-D1, S10-D3 |
| Create Source modal: Name first, then Type | S10-15 | S10-D2 |
| Rethink the Observation list + Add Observation focuses search | S10-16 | S10-D3 |
| Auto-fill name parts + name part type dropdown | S10-17 | S10-D4 |
| Drag files onto Artifacts + open the file without expanding | S10-18 | S10-D5, S10-D6 |
| Rethink the Event icon + sidebar icons match Subject cards | S10-19 | S10-D7 |
| Box-select and move multiple Subject cards | S10-20 | S10-D6 (already designed for S10-18) |
| Last updated on list views | S10-21, S10-22 | S10-D8…D11 |
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
| Composer depiction row | S10-34 | S10-D3 (same board as S10-16) |
| Person gallery + list thumb | S10-35 | S10-D12, S10-D9 |
| Preferred portrait | S10-39 | S10-D12, S10-D9 (same boards as S10-35) |
| Promote compares two crops | S10-36 | S10-D13 |
| Project collation locale (no picker) | S10-37 | — |
| Seed event type `engagement`; couple title like marriage | S10-38 | — |

**Left in the dogfood log**

- **Dense Evidence graphs (census-scale).** Descoped 2026-10-08. Layers and filters stay in the dogfood log. Canvas growth (S10-03) and box-select (S10-20) still land; they do not add a way to hide part of a Source.
- **Automatic structure from the Artifact**, beyond the two slices in S10-13 and S10-17. No catalog writes from a detector. No OCR-on-PDF, Live Text, or Foundation Models.

**Back in the ideas folder** (not PRs)

- [Depictions leftovers](../../ideas/depictions-and-likenesses.md) — other keys, signature, transparent polygon, Windows PDF draw.
- [International-name leftovers](../../ideas/international-names.md) — other name patterns, transliteration, Beider–Morse, calendars and double dating, per-person `name_format`, a collation picker.
- [Name-matching leftovers](../../ideas/name-matching-enhancements.md) — population frequency, known-as, concluded values, calibration, researcher-defined weights, age-to-birth-year, rejected-claim memory, blocking.

## Working rules

These are the calls for this spike. The questions at the bottom are answered; nothing in them still blocks a PR.

- **Add lands at the top for every Observation**, bridge or not. Today `CitationConnections` inserts a bridge at index 0 and `CitationObservationRows.addDraft` appends. Focus follows the new row's property search (the composer already stores `focusedID`; Add does not move keyboard focus into that combo box).
- **Auto Transcribe moves off the Citation menu's corner.** In the composer form, the Citation `PVSelect` and the Auto Transcribe button both sit on the trailing edge of the left column (`CitationComposerFormPane`), which is why the misclick happens. The confirm-before-overwrite stays.
- **Choice field cutoffs:** 2 options → `PVRadio`. 3–8 → `PVSelect`. 9 or more → `PVComboBox`. Count the options the field would show, including an empty/untyped row when that row is real. Vocabulary edits happen on other screens, so the control will not swap mid-edit. Name part types are 8 (`NamePartType` plus untyped), so they stay a select. This field does not replace that dropdown.
- **Observation defaults** write into the existing controls and stay editable. Person→Event role: seeded `subject` ([`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.5). Place→Place: seeded `part_of` (§3.8). That term already exists; the dogfood note could not find it. No new place-relationship term.
- **Date prefill is silent.** First match in the transcription, current locale, Gregorian patterns a genealogist actually types (day month year, month day year, year alone, ISO). 80–90% is the bar. No "suggested" chrome, so the date modal does not need a brief. Year is focused even when the prefill is empty.
- **Name split does not clobber edits.** Blur of the full-form line fills parts only when the part list is empty. The Split button refills from the line; if any part was edited by hand, confirm before replacing. Guessing rules are the ones in the dogfood note (comma → surname first; otherwise last word is surname; leftovers are given).
- **Outside click closes a sheet.** If the sheet is dirty, use the existing confirm (discard or keep editing) — the same idea as leaving the composer with unsaved work. A clean sheet closes. Read-only sheets close. This is behavior on the shared sheet presentation, not a new dialog style.
- **Graph camera is session state, not catalog state.** `GraphCanvasCamera` is explicitly not research data. Coming back from the composer, Back, or breadcrumbs restores magnification and content offset for that Source for the rest of the session. A persisted per-user table is out.
- **Canvas grows.** `EvidenceGraphPlacement.contentSize` is a fixed 4000×4000 and placement clamps to `maxCell`. When a card's frame nears the document edge, grow the document (and the clamp) by a step. No new chrome.
- **Zoom bug is a restore, not a new look.** Floor is `GraphCanvasCamera.minMagnification` (0.25). The grid must not turn solid black. Edges must fade or hold together, not disappear one at a time. No brief.
- **Hit boxes are the painted controls.** Toolbar place/connect buttons and the bridge card's open-citation control. The card icon hit helper (`EvidenceCardLayout`) is the pattern: hits centered on what is drawn.
- **Oldest user** means the earliest UUIDv7 `users.id`, not the first row of `ORDER BY display_name`. If the signed-in install identity is already in the project, keep selecting that person. If the project has no users, stay on create-new.
- **Enter saves a single-line inline edit.** `PVInlineEdit` on the Property title. Do not steal Enter from a multiline description.
- **Sources sort and filter** persist per project and user and restore when the Sources place is opened again. People, Events, and Places lists have no sort or filter control; they are out of this PR.
- **Last updated** is an audit scope, not a timestamp column. Sources already store `UpdatedRevision` (`scope_type = source`) and can show it once the row is designed. People, Events, and Places have no scope today (`canonical_entity` is `noScope` in `core/database/audit/scopes.go`). S10-21 adds a `canonical_entity` scope: Identity Claim writes, and Observation writes on a Subject that is a member, so editing evidence moves the handle's revision. Display is relative or absolute in the briefs; the value is the revision's time.
- **One Event mark.** Sidebar People / Events / Places use SF Symbols (`person`, `calendar`, `mapPin` on `WorkspaceSection`). Cards, lists, omnibar hits, Promote, and Properties use `PVMark` (`subject_person`, `subject_event`, `subject_place`). The Event asset is the hourglass. S10-D7 draws the new Event mark and switches the sidebar (and the toolbar, which reads `WorkspaceSection.icon`) to the subject marks. Every existing `PVMark` call site picks up the new asset with no extra brief. List empty states that still use `PVSymbol` for these three kinds switch in the same PR.
- **Sex / gender styling** uses a seeded `gender` Property when the Person has one, otherwise `sex_at_birth`. The seed is `man`, `woman`, `non_binary` (labels Man, Woman, Non-binary). Researchers add further terms (`origin=user`), the same way they extend `role`. It is not product-locked, and there is no product `other` term. The page treatment recognizes the three seeded keys. A researcher-added term still counts as gender and wins over sex at birth; the brief draws that term without inventing a new color per key. Unknown, empty, and mixed (the reconciler already has a mixed state) stay neutral. `sex_at_birth` already appears as an ordinary field row when a record speaks to it (`PersonDetailContent`); this PR adds the page treatment and leaves that row in place.
- **Composer performance** is profiled before the row rethink. S10-11 fixes fetch or render cost without a new layout. S10-16 must still be comfortable at 20 Observations. If the profile says the row view *is* the cost, S10-11 only removes work that a redesign would not redo (duplicate loads, per-row queries), and the layout fix waits for S10-16.
- **Artifact drop already exists inside the Add dialogs** (`IngestFileDropRow` on create and attach). It does not exist on the Artifact section of the Source page. S10-18 adds the section target. Open File today is only in the expanded detail (`SourcePageArtifactsView.artifactPrimaryFileColumn`). The Evidence graph header has no open-file control.
- **Adaptive field adoption** in this spike is the composer, where term lists grow. Other screens keep their current control until a later PR. A view that must grow to fit a radio row gets its own brief at that time; none of those are in this spike except the composer.
- **Name part types stay a select.** `patronymic` joins the compiled registry (S10-29), so the list can pass eight. It is still a fixed-width `PVSelect`. The adaptive field is for vocabulary that grows without a bound. User-minted part types stay out. The Western split rules in S10-17 do not become profile-aware.
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

**Every UI PR is gated by Claude Design briefs — one brief per view.** A PR that touches two views waits on two briefs. Briefs follow [`add-design-brief`](../../../.cursor/skills/add-design-brief/SKILL.md) and live in [`design/`](design/). **Each brief is designed alongside its feature**, just before the PR it points at.

| Brief | View | Gates | Later on the same view |
| --- | --- | --- | --- |
| **S10-D1** | Adaptive choice field (kit board) | **S10-14** | S10-16 instances it |
| **S10-D2** | Create Source dialog | **S10-15** | — |
| **S10-D3** | Citation composer | **S10-16** | S10-34 depiction row |
| **S10-D4** | Name parts modal | **S10-17** | Design after S10-29, so the type list includes `patronymic` |
| **S10-D5** | Source page — Artifacts | **S10-18** | — |
| **S10-D6** | Evidence graph | **S10-18** header open | S10-20 box-select |
| **S10-D7** | Sidebar | **S10-19** | Toolbar section icon ships in the same PR |
| **S10-D8** | Sources list | **S10-22** | — |
| **S10-D9** | Persons list | **S10-22** | S10-35 fills the thumb with a likeness. S10-39 makes that thumb the preferred portrait |
| **S10-D10** | Events list | **S10-22** | — |
| **S10-D11** | Places list | **S10-22** | — |
| **S10-D12** | Person detail | **S10-23** | S10-35 gallery and S10-39 preferred portrait. Design both with the gender treatment, before S10-23 |
| **S10-D13** | Promote page | **S10-36** | Evidence sheet shows two crops |

S10-D6 is one board for the graph's new chrome (open file, box-select). S10-01…S10-04 do not wait on it. S10-D8…D11 are four briefs because they are four places; Events and Places should specify the same last-updated cell as Persons so the shared `ConclusionListPage` does not grow three treatments. Only the Persons list thumb becomes a likeness. S10-D1 is a kit board: the control is the design, and stamping it into an unchanged form is not a new view. S10-D3, S10-D9, and S10-D12 each cover a later PR on that same view, so those briefs are written with the later frame in them. D12 includes the preferred-portrait control S10-39 fills, and D9 includes that crop as the list thumb.

## PR sequence

**✎ = design brief**, just before the PR it points at. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Evidence graph correctness (no new chrome)
  S10-01  Button hit boxes match the buttons
  S10-02  Zoomed-out grid and edges
  S10-03  Canvas grows before a card hits the edge
  S10-04  Camera restored after the composer / Back / Forward
  Check: pan, zoom, open the composer, come back — same view. A toolbar click
         lands on the padding, not only the icon. Zoomed out, the grid is not black.

SLICE 2 — Behavior, no new chrome
  S10-05  PVTextArea shows a scrollbar when the text overflows
  S10-06  Omnibar: a whole word in a title beats stacked fuzzy prefixes
  S10-07  Existing project selects the oldest user
  S10-08  Enter saves a Property title
  S10-09  Sources sort and filter persist
  S10-10  Observation defaults: role subject, place part_of
  S10-11  Composer stays responsive; profile first
  S10-12  Outside click closes a sheet; dirty asks
  S10-13  Date modal: focus Year, prefill the first transcription match
  Check: `marriage` hits the certificate before Marion Margaret. A long
         transcription scrolls. A new Person→Event row already says subject.

SLICE 3 — Choice field
  ✎ S10-D1 ──▶ S10-14  Adaptive choice field in the kit
  Check: a 2-option field is radios, a 5-option field is a select, a long
         term list is a combo box. Preview or a kit host is enough; the
         composer adopts it in S10-16.

SLICE 4 — Composer and the dialogs it opens
  ✎ S10-D2 ──▶ S10-15  Create Source: name, then type
  ✎ S10-D3 ──▶ S10-16  Observation list rethink (after S10-10, S10-11, S10-14).
                       The board also draws the depiction row S10-34 fills.
  S10-29  Name patterns (patronymic, western + Spanish profiles) — before D4
  ✎ S10-D4 ──▶ S10-17  Name parts: split, fixed width, letter cycle
  Check: add an Observation — it is at the top, focused, and a role is already
         subject. Auto Transcribe is not beside the Citation menu. "Smith, John"
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
  S10-21  canonical_entity audit scope + list revision reads
  ✎ S10-D8…D11 ──▶ S10-22  Last updated on four lists
  ✎ S10-D12 ──▶ S10-23  Person page sex / gender. The board also draws the gallery S10-35 fills and the preferred portrait S10-39 fills.
  Check: edit a Source, leave, come back — the row's time moved. Edit a member's
         Observation — the Person row's time moved. A Person with gender shows
         gender; one with only sex at birth shows that; unknown is neutral.

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
  S10-34  Composer depiction row (S10-D3 already designed)
  ✎ S10-D13 ──▶ S10-36  Promote evidence sheet shows the two crops
  S10-35  Person gallery + Persons list thumb (S10-D12 and S10-D9 already designed)
  S10-39  Preferred portrait leads the page and the list (same boards)
  Check: a group photo yields one crop per face. The Person page shows the
         gallery. Mark one face: it leads the page and the People list, and
         the other faces stay. Promote shows the two pictures. Clearing the
         cache does not lose the Citation.

SLICE 9 — Event vocabulary
  S10-38  Seed engagement; two subjects title "Engagement of A and B"
  Check: a new catalog lists Engagement. Two people title as a couple.
         An existing catalog is unchanged.

S10-99 — Dogfood close / docs
```

S10-05…S10-13, S10-15, and S10-19 have no dependency on the graph slice. S10-24, S10-29, S10-30, S10-31, S10-32, and S10-38 have no dependency on the dogfood slices. They are ordered here so review stays one stream. Parallel PRs are fine when **Depends on** allows it. S10-29 is numbered with the name work and listed in slice 4 because the name-parts brief cannot be drawn until `patronymic` exists.

### Dependencies at a glance

```text
S10-01 hit boxes ───────────────┐
S10-02 zoom                     │
S10-03 canvas grow ─────────────┼──▶ S10-20 box-select
S10-04 camera (after 02, 03)    │         ▲
                                │         │
S10-10 defaults ─┐              │    S10-D6 (also gates S10-18)
S10-11 perf ─────┼──▶ S10-16 composer rethink
S10-14 choice ───┘         ▲
   ▲                       │
S10-D1                     S10-D3

S10-21 scope ──▶ S10-22 last updated (D8…D11)
S10-D12 (gender + gallery + preferred portrait) ──▶ S10-23
S10-29 patterns ──▶ S10-D4 ──▶ S10-17 name parts
                 └──▶ S10-37 collation

S10-24 fold ──▶ S10-25 tables ──▶ S10-26 phonetic
S10-28 support + negatives (after S10-24, so fixtures settle once)

S10-32 depiction type ──▶ S10-33 crop cache ──▶ S10-34 composer row (after S10-16)
                      └──▶ S10-36 Promote crops
S10-33 ──▶ S10-35 gallery (after S10-23; briefs D12 and D9 already drawn) ──▶ S10-39 preferred
```

---

## PRs

### Slice 1 — Evidence graph correctness

#### S10-01 — Button hit boxes

| | |
| --- | --- |
| **In** | Toolbar controls (add place, add person, add event, connect) take a click on the whole button. The bridge card control that opens the citation is hit where it is drawn, not one button-width to the left. Follow the card icon hit helper: document-space frames centered on the painted control. |
| **Out** | New buttons. Box-select (S10-20). |
| **Testable** | Click the padding of a toolbar button and it arms. Click the bridge citation button and the composer opens; the empty space beside it does not. |
| **Depends on** | — |

#### S10-02 — Zoomed-out grid and edges

| | |
| --- | --- |
| **In** | At `minMagnification` (0.25) the background grid does not turn solid black, and connector lines do not vanish unevenly. Grid may fade out as a whole. Lines share one rule at every zoom. |
| **Out** | A new zoom range. Minimap. |
| **Testable** | Zoom to the floor on a graph with several edges: grid is intact or evenly faded; every edge is still there or every edge has faded the same way. |
| **Depends on** | — |

#### S10-03 — Canvas grows

| | |
| --- | --- |
| **In** | The document and `maxCell` grow by a step when a card's frame approaches the edge, so placement and dragging are not clamped at 4000×4000. Existing graphs open as they do today until something nears the edge. |
| **Out** | A visible boundary, minimap, or user-set canvas size. |
| **Testable** | Drag a card to the current edge: the canvas extends and the card is not stuck. A small graph does not open on a huge empty document. |
| **Depends on** | S10-02 (grid drawing has to survive a larger document) |

#### S10-04 — Camera survives leaving the graph

| | |
| --- | --- |
| **In** | Store `GraphCanvasCamera` (magnification + content offset) for the Source on the workspace session. Restore it when the graph place appears again, including composer round-trip, Back, and Forward. |
| **Out** | Persisting the camera across relaunch. A catalog table. |
| **Testable** | Pan and zoom, open the composer, go back: same center and zoom. Quit and relaunch may reset. |
| **Depends on** | S10-02, S10-03 |

### Slice 2 — Behavior, no new chrome

#### S10-05 — Text area scrollbar

| | |
| --- | --- |
| **In** | `PVTextArea` shows a scrollbar when the text overflows its `lineLimit`. Transcription, transcription note, description, and every other caller pick it up. SwiftUI `TextField` axis vertical does not show one today; an `NSTextView` (or equivalent) inside the existing chrome is in bounds. |
| **Out** | A per-screen scrollbar. A new text style. |
| **Testable** | Paste a long transcription: a scrollbar appears and tracks the caret. A short note does not show an empty track. |
| **Depends on** | — |

#### S10-06 — Omnibar whole-word rank

| | |
| --- | --- |
| **In** | In `core/search`, a whole-token match in a title outranks fuzzy and prefix matches. Several partial tokens (Marion + Margaret for the query `marriage`) must not sum past one exact title hit. Kind mix (`ScoreMix`) stays as Spike 9 left it: it may break ties, not overturn a clearly stronger text match. |
| **Out** | New hit chrome (S9-D13 already shipped). A different kind order. |
| **Testable** | Fixture: a marriage certificate and a newspaper announcement whose titles contain `marriage`, plus obituaries for Marion Margaret. Query `marriage` lists the title hits first. Existing rank tests still hold (shorter exact title, exact ref). |
| **Depends on** | — |

#### S10-07 — Oldest user on an existing project

| | |
| --- | --- |
| **In** | After choosing an existing project, if the install identity is not in `catalogUsers`, select the user with the smallest UUIDv7 id. Keep the current selection when that identity is already a contributor. Empty user list still offers create-new. `users` has no `created_at`; id order is the clock. The list query stays alphabetical for display. |
| **Out** | A new onboarding layout. |
| **Testable** | Open a project whose users are not you: the oldest user is selected, not "New user". Open a project you already belong to: you stay selected. |
| **Depends on** | — |

#### S10-08 — Enter saves a Property title

| | |
| --- | --- |
| **In** | Enter in the Property title `PVInlineEdit` saves, same as the save control. Multiline description does not save on Enter. |
| **Out** | A new editor. |
| **Testable** | Edit a title, press Enter: it saves. Edit the description, press Enter: a newline, not a save. |
| **Depends on** | — |

#### S10-09 — Sources sort and filter persist

| | |
| --- | --- |
| **In** | Remember `SourcesModel.sort` and `typeFilterID` per project and user. Restore when the Sources place is shown again, including after leaving to a Source page and after relaunch. |
| **Out** | Sort or filter on People, Events, or Places. New menu items. |
| **Testable** | Choose Updated and a type, open a Source, come back: both are still set. Relaunch: still set. |
| **Depends on** | — |

#### S10-10 — Observation defaults

| | |
| --- | --- |
| **In** | A new Person→Event connection prefills role with the seeded `subject` term. A new Place→Place connection prefills `place_relationship_type` with seeded `part_of`. Both stay editable. Apply on the draft row before the first save. More defaults wait for another dogfood note. |
| **Out** | New vocabulary. Changing `part_of` / `succeeded_by` (they stay product-locked). Overwriting a role the researcher already picked. |
| **Testable** | Connect a person to an event: role is subject, and choosing another role sticks. Connect two places: relationship is part of. |
| **Depends on** | — |

#### S10-11 — Composer performance

| | |
| --- | --- |
| **In** | Profile opening a Citation with about 20 Observations. Fix the time that is data fetching or redundant work. Leave the row layout to S10-16. Record the before/after in the PR (a numbers note is enough; no new ledger file). |
| **Out** | A virtualized list designed from scratch (that is the rethink, if the profile says paint is the cost). |
| **Testable** | A Citation with 20 Observations opens without a visible hitch on the machine we dogfood on. A Citation with a handful is unchanged. |
| **Depends on** | — |

#### S10-12 — Outside click closes a sheet

| | |
| --- | --- |
| **In** | Clicking outside a modal sheet closes it. Dirty → the existing confirm (discard or keep editing). Clean or read-only → close. Apply it on the shared sheet path (`PVFormDialog` and the other `.sheet` presenters), and teach a sheet to say it is dirty when it has edits. |
| **Out** | A new confirm style. Click-outside on menus and popovers (those already dismiss). |
| **Testable** | Open Create Source, type a title, click outside: confirm. Open it and click outside with nothing typed: it closes. A delete confirm is unchanged. |
| **Depends on** | — |

#### S10-13 — Date modal focus and transcription prefill

| | |
| --- | --- |
| **In** | The structured date modal opens with focus on Year. When the Citation transcription contains a date, prefill from the first match (working rules). The researcher can clear or edit it. Patterns live with the date editor, keyed by locale, not as catalog data. |
| **Out** | A suggestion chip. Parsing names. Writing an Observation without the researcher saving. Languages we do not ship a catalog for, beyond the current locale's Gregorian forms. |
| **Testable** | Transcription `14 May 1817`: the modal opens on Year with 1817, month May, day 14. Two dates in one transcription: the first wins. No date: Year focused, fields empty. |
| **Depends on** | — |

### Slice 3 — Choice field

#### S10-14 — Adaptive choice field

**Brief:** S10-D1, just before this PR.

| | |
| --- | --- |
| **In** | One kit field that picks `PVRadio`, `PVSelect`, or `PVComboBox` from the option count (working rules). Hosts pass options; they do not pick the control. Ship it with a preview or host that shows all three bands. Do not restyle every form in this PR. |
| **Out** | Composer adoption (S10-16). A new menu primitive. Switching control while the field is on screen. |
| **Testable** | 2, 5, and 20 options render radio, select, and combo. Keyboard and accessibility labels match the control that appeared. |
| **Depends on** | S10-D1 |

### Slice 4 — Composer and the dialogs it opens

#### S10-15 — Create Source field order

**Brief:** S10-D2.

| | |
| --- | --- |
| **In** | In the Create Source dialog, name (`draft.title`) is the first field and type is the second. Description stays third. Validation and the type combo row are unchanged. |
| **Out** | New fields. The Source page identity editor. |
| **Testable** | Open Create Source: focus and tab order are name, type, description. Create still requires both name and type. |
| **Depends on** | S10-D2 |

#### S10-16 — Observation list rethink

**Brief:** S10-D3. The board instances the S10-14 field for term values; design it after that field is in the kit.

| | |
| --- | --- |
| **In** | Per S10-D3: Observations are visually separate, and the fields inside a row are easy to tell apart. Add (bridge or not) inserts at the top and moves keyboard focus to that row's property search. Auto Transcribe (and the PDF paste control, which sits in the same slot) moves off the Citation menu's corner. Term-valued fields on the row use the adaptive choice field. Still comfortable at 20 Observations (S10-11's bar). |
| **Out** | Name parts modal (S10-17). Date modal (S10-13). The depiction row itself (S10-34 fills the frame this brief already draws). |
| **Testable** | Twenty rows: you can see where one ends. Add a bridge and a plain Observation: both appear at the top, search focused. Reaching for the Citation menu does not start Auto Transcribe. A two-term property shows radios; a long term list shows a combo box. |
| **Depends on** | S10-10, S10-11, S10-14, S10-D3 |

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

#### S10-21 — Handle last-activity scope

| | |
| --- | --- |
| **In** | New audit scope `canonical_entity` ([`add-audit-scope`](../../../.cursor/skills/add-audit-scope/SKILL.md)). Identity Claim writes scope to the handle. An Observation write on a member Subject also scopes to that handle, and still scopes to the Source. List reads for Persons, Events, and Places return the latest revision time the way Sources already return `UpdatedRevision`. Backfill existing history in the next migration. Resolver tests; no migration test. |
| **Out** | The row UI (S10-22). A `updated_at` column on `canonical_entities`. |
| **Testable** | Promote a Person: the handle has a revision. Edit a member Observation: that handle's revision moves, and the Source's revision still moves. A handle with no claims does not appear. Rebuild-style check: scope rows match the writes, including a delete of a member. |
| **Depends on** | — |

#### S10-22 — Last updated on four lists

**Briefs:** S10-D8 Sources, S10-D9 Persons, S10-D10 Events, S10-D11 Places. Design the four before the PR. Events and Places copy the Persons cell.

| | |
| --- | --- |
| **In** | Each list row shows when it last changed, from the revision S10-21 (Sources: the existing `UpdatedRevision`). Empty or never-audited reads as the brief's empty. Sorting Sources by Updated already exists; this PR does not add sort to the other lists. |
| **Out** | A new list layout beyond the cell the briefs add. Sort/filter persistence (S10-09, already landed). |
| **Testable** | All four lists show a time. Edit a Source note: Sources row moves. Edit a member Observation: the Person row moves and the Source row moves. Events and Places match the Persons cell. |
| **Depends on** | S10-21, S10-D8, S10-D9, S10-D10, S10-D11 |

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

**Brief:** S10-D3, already designed for S10-16.

| | |
| --- | --- |
| **In** | The depiction row the board drew: pick the region, leave transcription empty, save a likeness (or another depiction property, if the board shows only likeness). The row uses the same add-at-top and focus rules as S10-16. |
| **Out** | A new composer layout. Gallery (S10-35). |
| **Testable** | Draw a face on a group photo, add a likeness, save. The row shows the crop. Transcription can be empty. |
| **Depends on** | S10-16, S10-33, S10-D3 |

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

### Slice 9 — Event vocabulary

#### S10-38 — Seed event type `engagement`

| | |
| --- | --- |
| **In** | Add `engagement` / Engagement to `subjectvocab` `seedTerms` for `event_type`, beside `marriage`. Install on catalog create only ([`add-seeded-vocabulary`](../../../.cursor/skills/add-seeded-vocabulary/SKILL.md)). [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4 lists the key. `eventtitle.Choose` uses the couple rule for `engagement` as well as `marriage` when there are exactly two subjects. The couple template is *{Type} of {A} and {B}* in `EventTitleDisplay` and in the search-index title (`handles.go` currently hardcodes "Marriage of"), so a marriage still reads *Marriage of A and B* and an engagement reads *Engagement of A and B*. Follow [`add-localized-string`](../../../.cursor/skills/add-localized-string/SKILL.md) for the template. No new chrome, so no brief. |
| **Out** | A migration that inserts the term into existing catalogs. Those projects get the term by hand. Matching affinity with marriage. A new role. Three or more subjects (still *et al.*). |
| **Testable** | A catalog created after this PR offers Engagement. Two subjects title *Engagement of A and B* on the list, the page, and in search. A marriage of two still titles *Marriage of A and B*. Three subjects of an engagement title *Engagement of A et al.* Opening an older catalog does not add the term. |
| **Depends on** | — |

### S10-99 — Dogfood close / docs

| | |
| --- | --- |
| **In** | Move shipped notes to Done in [`docs/dogfood/ux.md`](../../dogfood/ux.md). Distill this spike with [`archive-docs`](../../../.cursor/skills/archive-docs/SKILL.md) when the bar above is true. The three idea notes live in this folder; the close distill keeps their decisions and drops the PR sequence. |
| **Out** | Pulling the parked density, extract, transliteration, or Reconciliation-Claim items in at the close. |
| **Depends on** | S10-01…S10-39 |

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
| `part_of` as the place default (already seeded) | A new place-relationship term |
| `gender` seeded as `man`, `woman`, `non_binary`, plus researcher terms | Treating `sex_at_birth` as gender. A locked term list. |
| `engagement` on new catalogs | A migration that inserts it into catalogs that already exist |
| Claim editing, weak-claim review | Still [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md) |

## Gotchas

- **Bridge rows and plain rows are different types.** Add-at-top has to stay true after reload, not only on the draft. Connections are grouped by bridge in `CitationConnections`; plain rows append in `CitationObservationRows`. One insert rule, both lists, and the saved order the composer shows has to match.
- **Focus and the new row.** `focusedID` is set on add today and does not move keyboard focus into the property combo box. The rethink wants both: row at top, field focused.
- **`PVSelect` letter jump is one step.** `handleTriggerKey` sends `.character` once. Cycling is a session change for every select, not a name-modal special case. Width is the opposite: do not make every select as wide as its longest option.
- **Growing the canvas moves clamps.** `EvidenceGraphPlacement` tests lock `maxCell` at grid 96×98. Growing is a behavior change those tests will need to follow. Card coordinates already stored stay valid.
- **Camera vs document size.** Restoring an offset from a smaller document after S10-03 grows it should clamp, not scroll into emptiness. `clampedContentOrigin` already exists.
- **Drop on the section vs drop in the dialog.** The dialog row already highlights and ingests. The section target should call that ingest, not a parallel copy, including the reject callout for a bad type.
- **Open file on the graph.** A Source can have several Artifacts. The brief picks which file the header opens (cover if set, otherwise the single file, otherwise no button or a menu — D6 decides; do not invent a menu in the PR if the brief says one file).
- **Audit scope fan-out.** An Observation on a shared member can touch one Source and one or more handles. Delete changes still need the parent FKs. Backfill is `INSERT OR IGNORE` in the migration, pattern `000036.sql`. `TestResolversCoverEveryEntityType` must stay green.
- **Last updated is not create time.** UUIDv7 id is create order. The Sources "Updated" sort already uses the scope. The new cell must use that revision, not the id.
- **Gender and sex at birth are different Properties.** Seeding gender does not relabel sex at birth. The page treatment is a reading of reconciled values, not a new truth write.
- **Outside click and macOS sheets.** AppKit sheets do not dismiss on an outside click. This is a product behavior on our presenters. Do not bolt a scrim onto `PVConfirm` (design-system rule: no dimming scrim; the window owns chrome).
- **Defaults must not fight Add focus.** S10-10 fills role on the new row. S10-16 focuses the property search, not the role control.
- **Folding is not reconciliation.** A test that José and Jose become one auto-reconciled name is a bug. Matching suggestions may still rank them together.
- **`name_format_profiles` is not in the catalog yet.** S10-29 creates it. Do not point `NamePatterns` at a table that only exists in [`structured-name-model.md`](../../structured-name-model.md) §4.
- **Depiction is not a file.** Delete and Impact stay on the Citation and the Artifact. The crop cache is disposable. A polygon edit must not leave the old crop as if it were the current face.
- **Gallery and gender share a page.** S10-23 ships the gender treatment inside frames that already reserve the gallery and the preferred-portrait control. S10-35 fills the gallery. S10-39 fills the control. Designing the gender page without those frames means a second Person-detail brief, which this plan does not allow.
- **Frequency on a tiny project.** Inverse frequency with one of each name is a flat weight. The scorer needs a floor so the first Smith is not treated as rare forever, and so an empty project still matches.
- **Do not backfill `engagement`.** Seeded terms are Install data. A migration `INSERT` would put the term in catalogs the researcher already has, and this spike does not do that. Add it on those projects by hand.

## Questions

| # | Question | Why it blocks |
| --- | --- | --- |
| Q1 | **Gender terms.** Seed `man`, `woman`, `non_binary`. Researchers extend the vocabulary. Answered 2026-10-08. | — |
| Q2 | **Dense Evidence graphs.** Descoped 2026-10-08. Stays in the dogfood log. Extract beyond date prefill and name split stays there too. | — |

Everything else in **Working rules** is decided for planning. Correct any of them before the PR that implements it.

## Docs to update as work lands

| When | Doc |
| --- | --- |
| S10-10 | Nothing in seeded vocabulary (terms already exist). Mention the default in the composer only if a domain doc states entry defaults — none does today. |
| S10-21 | [`docs/audit-revision-history.md`](../../audit-revision-history.md) §4.1: `canonical_entity` is a scope. The catalog-audit rule's "conclusion layer carries no Source scope" stays true; the handle scope is additional. |
| S10-23 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) new gender section. Do not rewrite §3.7. |
| S10-25, S10-29, S10-32 | Seeded vocabulary: abbreviation and nickname tables, `patronymic` and the Spanish profile, `likeness` / `depiction`. [`docs/matching.md`](../../matching.md) and [`docs/structured-name-model.md`](../../structured-name-model.md) follow the code. |
| S10-38 | [`docs/seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.4: `engagement`. Couple rule covers `engagement`. No migration. |
| S10-99 | Dogfood log Done. Archive this folder, including the three idea notes, as decisions. |

## Definition of done

The dogfood bar at the top is true in the app, including the name and likeness lines. A person can mark one likeness to lead the page and the People list. A new catalog's event types include Engagement, and two people on one title as *Engagement of A and B*. Each gated PR matches its brief. Parked notes and the ideas-folder leftovers are still parked. Claim editing is still an idea, not a half-finished extra. Existing catalogs were not given `engagement` by a migration.
