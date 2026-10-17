# S10-D15 — Citation composer: likeness

**Kind:** Claude Design board (**enhancement** of the S10-D3 citation composer)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Citation composer — the observation dialog, the viewer while a face is drawn, and a saved likeness line
**Implements later as:** PR **S10-34**
**Depends on:** the agreed S10-D3 board, shipped in S10-16 ([#384](https://github.com/mendahu/provenencia/pull/384)); the engine and FFI for a depiction value (S10-32, S10-54, landed)
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-34; working rules "A depiction's value is its own row: region plus image", "The client renders the image; Go validates it and stores its own copy", "`likeness` is new catalogs only"). [`../depictions-and-likenesses.md`](../depictions-and-likenesses.md). The Person gallery that shows these crops is S10-D12 (S10-35); it is not this board. The S10-D3 brief is [`archive/S10-D3-citation-composer.md`](archive/S10-D3-citation-composer.md).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board. This file is the whole prompt.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
- **Rethink** (this brief says replace): throw away the old frames. Do not keep a before/after to ship.
- **Enhancement**: add to the existing frames. Do not start a second composer / graph / page.

1. **Clear this board’s local design-system cache.** Claude Design keeps a stale pack; drawing against it invents local copies of kit controls.
2. **Delete this board’s reference** to the design-system bundle.
3. **Pull a fresh copy** of the Provenencia design system from the main project. Do not continue until the fetched kit lists current components. If the kit looks stale or empty, delete the cache and refetch. Do **not** draw a replacement kit locally.
4. **Compose from that kit.** Instance existing components. Reach for a **bespoke / local** control only when the use is truly this domain. One call site is not a new design-system primitive.

**Reach for (kit).** Instance these first. The **UI building-block inventory** later in this brief names the snowflakes and which kit piece each situation should use.

| Situation | Use |
| --- | --- |
| Labeled value, textarea, or trailing control | Field + TextArea / Input |
| Primary / secondary / ghost action | Button; icon-only → IconButton |
| Choose one from a short list | Select |
| Searchable pick | ComboBox |
| Warning, error, or inline hint | Callout |
| Page- or pane-level empty | EmptyState |
| Confirm replace or destroy | Confirm (`item:` snapshot, not a Bool) |
| Resource delete with inbound check | DeleteImpact recipe (S8-D9 / **S8-13**) — confirm if allowed, notice if blocked |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### This brief is an ENHANCEMENT of the S10-D3 composer

Add to the composer frames S10-D3 agreed and S10-16 shipped. Keep the two columns, the stepped artifact and citation over the file, the one observation dialog (a `FormDialog`) over the subject groups, and the saved lines grouped by subject. Do not draw a second composer, a second viewer, or a likeness-only page. A likeness is one more kind of property observation: it is authored, saved, edited, duplicated, negated, and deleted the way an occupation is. What is new is that its value is drawn on the photo instead of typed.

---

## 1. Objective

A photograph should be evidence whose value is the picture. Today a researcher can draw a region on a Citation's photo, but every Observation still takes a typed value, so a group photo becomes a sentence per face. This board lets the researcher draw each face on the photo, save it as a **Likeness** of the person, and see that face on the saved line.

**What changes.** The observation dialog offers Likeness on a person when the Citation's file is a photo or a PDF. Choosing it asks for a region on the file instead of a typed value. While the researcher draws, they can see the photo. The dialog shows the face it will save. A saved likeness line shows the face. Editing a likeness can redraw it. Duplicate copies its region.

**What stays.** The composer layout, the Citation chooser, Auto Transcribe and Paste, the bridge flow, the subject groups, Duplicate, Edit, Negate, Delete, and the leave confirm. The region tools (rectangle, four L shapes, circle, freeform) are the ones the viewer already has. The Citation's own region stays the Citation's: drawing a face does not move or replace it.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A likeness is a property observation | Subject (a person) + property **Likeness** + a value. The value is a region on this Citation's file and the image the app cuts from it. Like a date or a name, it is the value, not a caption. |
| The Citation is the photo; each face is an Observation | A group photo of five people is one Citation with five likeness Observations, one per face, each on its person. The Citation's transcription can be empty (*Wedding party, 1902*, or nothing). Nothing asks for one. |
| The region tools exist | The viewer's tool strip already has rectangle, four L shapes, circle, and freeform, and draws one region at a time with handles. Today that region is the Citation's locator. The board decides how the researcher draws a face with those same tools without moving the Citation's region. |
| The viewer shows one region | It cannot show five faces at once today. Showing every saved face on the photo is not in this PR (§2.1). While a face is drawn or edited, that face is the region shown. |
| The dialog is modal | The observation dialog is a `FormDialog` sheet over the composer. Today nothing in it needs the file. A likeness does: the researcher must see and draw on the photo or the page. The board decides how the file and the dialog take turns — draw first and then open the dialog, a step in the dialog that steps aside to the viewer, or another arrangement that keeps one authoring home. |
| What is saved is the bounding box | The image is the rectangle around the drawn region, as a picture, at most 2048 pixels on the long edge. A circle saves the square around it. The polygon itself is kept with the value. |
| PDF pages | On a PDF, the face is on one page, and it must be the Citation's page. A new Citation takes the page the face is drawn on. An existing Citation that has a page only takes a face on that page. An existing Citation with no page cannot hold a face until its page is set; the dialog says so and offers nothing to draw (answered 2026-10-17). A photo has no pages. |
| Which files | A photo (any image the viewer opens) or a PDF. Audio, video, and files the viewer cannot open do not offer Likeness. |
| Which catalogs | A catalog made after S10-32 has Likeness on Person. An older catalog does not, and the dialog never offers it there. Researchers cannot create a likeness property. |
| Saving can fail | The image the app cut may be refused (an image the engine could not read), or the region may be refused. Both already have copy: *Draw the region again, or try a smaller one.* and *Invalid depiction value.* |
| Next face | A group photo is several faces in a row. After a likeness saves, the next face on the same photo is one action away and does not ask for Likeness again. |
| Edit and duplicate | Edit changes the person and negation, and can redraw the face; a redraw saves a new image and the old one goes (answered 2026-10-17). Duplicate opens the dialog with the same person, Likeness, negation, and region, unsaved; the researcher changes the person or redraws, and saving cuts its own image (answered 2026-10-17). |
| Negation | A negative likeness ("this is not Ada") is evidence, like any negative observation. The saved line shows the Negated badge as it does today. |

### 2.1 What this board is not

- The Person gallery and the Persons list thumb (S10-D12, S10-D9; S10-35).
- Choosing a preferred portrait (S10-39).
- Promote's side-by-side likenesses (Spike 11, S11-20).
- Every saved face drawn on the photo at once. One region is shown at a time.
- A crop that follows the polygon with transparency. The saved image is the bounding box.
- Place, event, or signature keys. Likeness on a person is the only key.
- A new image or PDF viewer, or new region tools.
- A redraw of the Citation chooser, the bridge flow, or Auto Transcribe.

---

## 3. Implementation gate (S10-34)

| Ships in **S10-34** | Stays out |
| --- | --- |
| Likeness in the property pick on a photo or PDF Citation; drawing the face with the existing tools without moving the Citation's region; the face shown in the dialog before save; save with an empty transcription; the next face one action away; the saved line showing the face; edit with redraw; duplicate copying the region; the PDF page rule; error and leave states. | Everything in §2.1. Showing saved faces on the photo. A likeness bridge or a likeness on an event or place. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CDR-1 | On a person, the property pick lists **Likeness** beside the other person properties when the Citation's file is a photo or a PDF. It is absent on audio, video, and unreadable files, and in a catalog that has no Likeness. |
| CDR-2 | Choosing Likeness replaces the typed value with a region on the file. The researcher draws it with the existing region tools, and can see the photo or the page while drawing. |
| CDR-3 | Drawing a face does not change the Citation's own region or the locator line above the file. After the face is saved or discarded, the viewer shows the Citation's region again. |
| CDR-4 | Before Save, the dialog shows the face it will save (the bounding box of the region) and lets the researcher draw it again. Save stays disabled until a region exists. |
| CDR-5 | Saving a likeness does not need a transcription. An empty Citation transcription saves. |
| CDR-6 | After a likeness saves, the next face on the same photo is one action away: Likeness stays chosen, the person and the region are empty, and keyboard focus is in the person pick. |
| CDR-7 | A saved likeness line shows the face as a thumbnail, with the person's group and the Likeness label like any line. A negative likeness shows the Negated badge. While the image loads, and when it cannot be read, the thumbnail shows the kit's loading and placeholder states, not an error. |
| CDR-8 | Editing a saved likeness shows its face on the file. The researcher can change the person and negation, and can redraw the region; saving a redraw replaces the image. Saving without a redraw keeps the image. |
| CDR-9 | Duplicate on a likeness line opens the dialog with the same person, Likeness, negation, and region, unsaved, with keyboard focus in the person pick. Saving it cuts its own image. The original is unchanged. |
| CDR-10 | On a PDF, the face is drawn on the Citation's page. A new Citation takes the page the face is drawn on. An existing Citation with a page shows that page while the face is drawn, and a face on another page cannot be committed. An existing Citation with no page shows a Callout in the dialog (*Set this citation's page to draw a likeness.*) and nothing to draw. |
| CDR-11 | A refused save shows the existing error copy in the dialog's Callout and keeps the drawn region, so the researcher can draw again. |
| CDR-12 | A drawn, unsaved face holds navigation the same way a typed draft does: the existing leave confirm asks, and discarding it discards the face. |
| CDR-13 | VoiceOver reads a saved likeness line as "Likeness, {person}" and its thumbnail as "Likeness of {person}". The dialog's face preview reads "Likeness to save". Drawing stays a pointer action, as the Citation region is today. |

---

## 5. Suggested frames

Add these to the S10-D3 composer frames, at the narrow and the wide width where the layout differs. Use a group photo of five people (*Wedding party, 1902*), a two-page PDF (*Parish newsletter*, page 2 has a photo), and the people Ada Lovelace, Ben Ford, and Clara Ford.

1. **Likeness in the pick.** The dialog on a person over the wedding photo: the property pick open with Name, Occupation, Interests, and Likeness. The value area for Likeness before any region is drawn.
2. **Drawing a face.** The researcher draws a rectangle around Ada's face on the photo. Show how the photo is reachable while the dialog's draft exists, and where the tools are. The Citation's own region (the whole party) stays the Citation's.
3. **Ready to save.** The dialog with Ada, Likeness, and her face preview; Save enabled; a control to draw again.
4. **Next face.** Ada's likeness just saved. The dialog is ready for the next face on the same photo: Likeness kept, person empty and focused, no region.
5. **Saved lines.** Ada's and Ben's groups, each with a likeness line showing the face, plus Ada's occupation line. One negative likeness (*not Clara*) with the Negated badge. One thumbnail loading, one unreadable.
6. **Edit and redraw.** Editing Ben's likeness: his face shown on the photo, the person pick, and the redraw control. A second frame after a redraw, ready to save.
7. **Duplicate.** Duplicate on Ada's likeness: the dialog with her region and Likeness, unsaved, the person pick focused; the researcher changes the person to Clara.
8. **PDF.** A new Citation on the newsletter, viewer on page 2: a face drawn there. Then an existing Citation on page 1 whose viewer was on page 2: the viewer shows page 1 while the face is drawn. Then an existing Citation with no page: the Callout and nothing to draw.
9. **Refused.** A save refused with *Draw the region again, or try a smaller one.* in the dialog's Callout; the region still drawn.
10. **Leave.** A face drawn and not saved, the researcher clicks Back: the existing leave confirm.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Observation dialog | Recipe | Extend | `Features/CitationComposer/CitationComposerAuthoringBox.swift` (`pvFormDialog`) | The one authoring home (CC-3). Likeness is a property kind inside it, not a second dialog. How it steps aside for the viewer while a face is drawn is this board's call. |
| Property value area | Snowflake | Extend | `Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields` | Shared by the new-observation and edit dialogs. A Likeness case beside text, integer, term, name, and date: the face preview and a draw / draw-again Button. |
| Property and person pick | Component | Ship | `PVComboBox` | Likeness is one more option (CDR-1). |
| Face preview | Component | Ship | `PVThumbnail` | CDR-4. The bounding-box image the save will send. Not a new image control. |
| Draw / Draw again | Component | Ship | `PVButton` | Starts or restarts drawing the face on the file. |
| Region tools | Snowflake | Ship | `Features/ArtifactViewer/ArtifactViewer.swift` `ArtifactViewerToolChrome` | The same radios the Citation region uses. No new tool. |
| Viewer | Snowflake | Ship | `Features/CitationComposer/CitationComposerViewerPane.swift` | Shows the face being drawn instead of the Citation's region while a likeness draft is drawn or edited (CDR-3). No second viewer. |
| Locator line | Snowflake | Ship | `Features/CitationComposer/CitationComposerViewerPane.swift` | Still the Citation's locator. A face does not appear in it. |
| Saved likeness line | Snowflake | Extend | `Features/CitationComposer/CitationComposerSubjectGroups.swift` | The existing line with a thumbnail in place of the value text (CDR-7). Same actions. |
| Saved face thumbnail | Component | Ship | `Platform/CachedThumbnail.swift` → `PVThumbnail` | Loads the saved image from the project; loading and unreadable states come from the kit. |
| Negated mark | Component | Ship | `PVBadge` | As on every line. |
| Line actions | Component | Ship | `PVIconButton`, `PVContextMenu` | Edit, Duplicate, Negate, Delete, as today. |
| No page / refused | Component | Ship | `PVCallout` | CDR-10, CDR-11. |
| Leave | Component | Ship | `PVConfirm` | The existing leave confirm (CDR-12). |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit image-crop, region-editor, or likeness-row primitive | One call site. Compose Thumbnail, Button, and the existing viewer. |
| A second viewer, a viewer inside the dialog, or a zoomed face editor | The composer's viewer is the one file view (S10-D3). |
| New region tools or a polygon-masked preview | The saved image is the bounding box. |
| A likeness page, a likeness list, or a gallery here | The gallery is S10-D12 on the Person page. |
| Drawing every saved face on the photo | One region at a time today. |
| A caption or transcription field on the likeness | The Citation's transcription is the reading; a likeness has none. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md) with S10-34.
3. Implement **S10-34** against the board and this inventory. The plan is [`../prs/S10-34-composer-likeness.md`](../prs/S10-34-composer-likeness.md).
