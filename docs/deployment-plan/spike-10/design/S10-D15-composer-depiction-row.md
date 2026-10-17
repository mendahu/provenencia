# S10-D15 — Citation composer: depiction row

**Kind:** Claude Design board (**enhancement** of the S10-D3 citation composer)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Citation composer — the observation dialog's depiction value area, drawing mode over the viewer, and a saved depiction line
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

Add to the composer frames S10-D3 agreed and S10-16 shipped. Keep the two columns, the stepped artifact and citation over the file, the one observation dialog (a `FormDialog`) over the subject groups, and the saved lines grouped by subject. Do not draw a second composer, a second viewer, or a likeness-only page. A likeness is one more property observation: it is authored, saved, edited, duplicated, negated, and deleted the way an occupation is. What is new is that its value is drawn on the file instead of typed.

**How drawing works (decided — draw the chrome, not the flow).** The value area of a depiction property has a **Draw on the file** button. Pressing it steps the observation dialog aside — the same way the composer already steps it aside to add a custom term or a new subject — and the composer is in drawing mode: the rectangle tool is armed, an info Callout above the canvas says what to do, with a ghost **Cancel** button beside it. When the researcher finishes a shape (mouse-up, as the region tools already commit), the dialog comes back with that region as its value. Cancel or Escape brings the dialog back unchanged. The board draws the drawing-mode chrome, the value area, and the saved line. It does not choose another flow.

---

## 1. Objective

A photograph should be evidence whose value is the picture. Today a researcher can draw a region on a Citation's photo, but every Observation still takes a typed value, so a group photo becomes a sentence per face. This board lets the researcher draw each face on the photo, save it as a **Likeness** of the person, and see that face on the saved line.

**What changes.** The observation dialog offers depiction properties (today only **Likeness**, on a person) when the Citation's file is a photo or a PDF. Choosing one replaces the typed value with a drawn region and a preview of the image it will save. Drawing mode, above. A saved line shows the image. Editing can redraw. Duplicate copies the region.

**What stays.** The composer layout, the Citation chooser, Auto Transcribe and Paste, the bridge flow, the subject groups, Duplicate, Edit, Negate, Delete, and the leave confirm. The region tools (rectangle, four L shapes, circle, freeform) are the ones the viewer already has. The Citation's own region stays the Citation's.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A depiction is a property observation | Subject + a property whose value type is *depiction* + a value. The value is a region on this Citation's file and the image the app cuts from it. The seeded depiction property is **Likeness**, on a person; later keys (a place, an event, a signature) use the same controls. Copy names the property by its label, so "Likeness" appears where the label is shown. |
| The Citation is the photo; each face is an Observation | A group photo of five people is one Citation with five likeness Observations, one per face. The Citation's transcription can be empty. Nothing asks for one. |
| The region tools exist | The viewer's tool strip has the region radios, Set page, and a Clear menu, and draws one region at a time with handles. Outside drawing mode, the region shown is the Citation's. In drawing mode, and while a depiction dialog is open, the region shown is the depiction's. |
| The dialog is modal | The observation dialog is a sheet over the composer. Drawing mode steps it aside (see the top of this brief). While drawing, the controls that change the Citation — the Citation chooser, the artifact switch, Auto Transcribe, Paste, Clear, and Set page — are disabled. Back still asks before leaving, and Keep editing returns to drawing. |
| What is saved is the bounding box | The image is the rectangle around the drawn region, at most 2048 pixels on the long edge. A circle saves the square around it. The polygon itself is kept with the value. |
| PDF pages | On a PDF, the region is on one page, and it must be the Citation's page (answered 2026-10-17). A **new** Citation takes the page the region is drawn on: its locator line then shows that page; its region is unchanged. Discarding the face does not undo the page. An **existing** Citation with a page: Draw moves the viewer to that page, and a shape finished on another page is not taken — the drawing Callout turns to a warning naming the right page. An existing Citation with **no** page: the value area shows a Callout asking to set the page, and Draw is disabled. A photo has no pages. |
| Which files | A photo (any image the viewer opens) or a PDF. Audio, video, and files the viewer cannot open do not offer depiction properties. |
| Which catalogs | A catalog made after S10-32 has Likeness on Person. An older catalog does not, and the dialog never offers it there. Researchers cannot create a depiction property. |
| Changing the draft | Choosing another Citation in the dialog clears a drawn region. Changing the property clears the value, as for any property. Changing the person keeps the region when the new person's type also has the property. Switching the dialog to Bridge works as today. |
| Saving can fail | The image may be refused (*Draw the region again, or try a smaller one.*) or the region may be refused (*Invalid depiction value.*). The app may fail to cut the image (*The region could not be cut from the file. Draw it again.*). The error shows in the dialog; the region stays. |
| Next face | After a likeness saves, the dialog stays open, as it does today, with Likeness kept, the person and the region empty, and focus in the person pick. Closing the dialog ends the run. |
| Edit and duplicate | Edit shows the face on the file, changes the person and negation, and can redraw; a redraw saves a new image and the old one goes (answered 2026-10-17). The edit dialog's dirty state includes the region. Duplicate opens the dialog with the same person, property, negation, and region, unsaved, focus in the person pick; saving cuts its own image (answered 2026-10-17). |
| A region the engine hides | A saved region that no longer matches its Citation's page reads back empty. Editing it shows a Callout asking to draw it again, and Save waits for the redraw. Its saved line still shows the image if it has one. |
| Negation | A negative likeness ("this is not Ada") is evidence. The saved line keeps the italic *not* before the image and the **Negative** badge, as every negative line does today. |

### 2.1 What this board is not

- The Person gallery and the Persons list thumb (S10-D12, S10-D9; S10-35).
- Choosing a preferred portrait (S10-39).
- Promote's side-by-side likenesses (Spike 11, S11-20).
- Every saved face drawn on the photo at once. One region is shown at a time.
- A crop that follows the polygon with transparency. The saved image is the bounding box.
- Another flow for drawing (draw first, a viewer inside the dialog).
- A new image or PDF viewer, or new region tools.
- A redraw of the Citation chooser, the bridge flow, or Auto Transcribe.

---

## 3. Implementation gate (S10-34)

| Ships in **S10-34** | Stays out |
| --- | --- |
| Depiction properties in the property pick on a photo or PDF Citation; the value area with preview and Draw; drawing mode with its Callout and Cancel; the PDF page rule; save with an empty transcription; the next face; the saved line with its image; edit with redraw; duplicate copying the region; error, hidden-region, and leave states. | Everything in §2.1. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CDR-1 | The property pick lists each depiction property bound to the subject's type (on a person today: **Likeness**) beside the other properties, when the Citation's file is a photo or a PDF. It is absent on audio, video, and unreadable files, and in a catalog that has none. |
| CDR-2 | Choosing a depiction property replaces the typed value with a value area: the preview (CDR-4) and **Draw on the file** (**Draw again** once a region exists). Draw enters drawing mode with the rectangle tool armed; the researcher can pick another region tool. |
| CDR-3 | Drawing mode does not change the Citation's region, and disables the Citation chooser, the artifact switch, Auto Transcribe, Paste, Clear, and Set page. Above the canvas an info Callout reads *Draw the region with the tools above. Escape goes back to the observation.* with a ghost **Cancel**. Finishing a shape returns to the dialog with that region; Cancel or Escape returns with the region unchanged. |
| CDR-4 | The value area shows the image the save will send (the bounding box of the region) as a 96pt square thumbnail, cover-fit, read by VoiceOver as *Image to save*. Save stays disabled until a region exists. |
| CDR-5 | Saving a depiction does not need a transcription. An empty Citation transcription saves. |
| CDR-6 | After a depiction saves, the dialog stays open with the property kept, the person and the region empty, and keyboard focus in the person pick. |
| CDR-7 | A saved depiction line shows the image as a 32pt square thumbnail, cover-fit, in place of the value text, so the line keeps its height. The person's group and the property label are as on any line. Tapping the line or the thumbnail opens Edit, as any line does. While the image loads, and when it cannot be read, the thumbnail shows the kit's loading state and its image placeholder glyph, not an error. With no image, the line shows "—" as today. |
| CDR-8 | Editing a saved depiction shows its region on the file. The researcher can change the person and negation, and can redraw; saving a redraw replaces the image, and saving without one keeps it. |
| CDR-9 | Duplicate on a depiction line opens the dialog with the same person, property, negation, and region, unsaved, with keyboard focus in the person pick. Saving it cuts its own image. The original is unchanged. |
| CDR-10 | On a PDF, the region is drawn on the Citation's page, per §2 "PDF pages". The wrong-page warning reads *Draw on page {n}, this citation's page.* The no-page Callout reads *Set this citation's page to draw a region.* and sits in the value area under the property pick. |
| CDR-11 | A refused or failed save shows its copy in the dialog's error Callout and keeps the region, so the researcher can draw again. |
| CDR-12 | A drawn, unsaved region holds navigation like a typed draft: the leave confirm asks, discarding discards the region, and Keep editing returns to where the researcher was (the dialog or drawing mode). |
| CDR-13 | VoiceOver reads a saved depiction line as "{property}, {person}" (*Likeness, Ada*) and its thumbnail as "{property} of {person}" (*Likeness of Ada*). Drawing stays a pointer action, as the Citation region is today. |
| CDR-14 | Editing a depiction whose region the engine hid shows the Callout *This region no longer matches the citation's page. Draw it again.* in the value area; Save waits for the redraw. |

---

## 5. Suggested frames

Add these to the S10-D3 composer frames, at the narrow and the wide width where the layout differs. Use a group photo of five people (*Wedding party, 1902*), a two-page PDF (*Parish newsletter*, page 2 has a photo), and the people Ada Lovelace, Ben Ford, and Clara Ford.

1. **Likeness in the pick.** The dialog on Ada over the wedding photo: the property pick open with Interests, Likeness, Name, and Occupation. Then Likeness chosen: the value area with no preview and **Draw on the file**.
2. **Drawing mode.** The dialog stepped aside; the info Callout with Cancel above the canvas; the rectangle tool armed; the disabled Citation controls; a rectangle being drawn round Ada's face. The locator line still shows the Citation's own region.
3. **Ready to save.** The dialog back with Ada, Likeness, her 96pt preview, **Draw again**, and Save enabled.
4. **Next face.** Ada's likeness just saved: the dialog with Likeness kept, the person empty and focused, no preview.
5. **Saved lines.** Ada's and Ben's groups, each with a likeness line and its 32pt thumbnail, plus Ada's occupation line. A negative likeness on Clara (*not* + thumbnail + Negative badge). One thumbnail loading, one showing the placeholder glyph.
6. **Edit and redraw.** Editing Ben's likeness: his region on the photo behind the dialog, the person pick, and Draw again. Then drawing mode over his face, then the dialog ready to save.
7. **Duplicate.** Duplicate on Ada's likeness: the dialog with her region and Likeness, unsaved, the person pick focused, the person being changed to Clara.
8. **PDF.** A new Citation on the newsletter, viewer on page 2, a face drawn: the locator line now shows page 2. An existing Citation on page 1 with the researcher paged to 2 in drawing mode: the warning Callout. An existing Citation with no page: the no-page Callout and a disabled Draw.
9. **Refused and hidden.** A save refused with *Draw the region again, or try a smaller one.*, region kept. Editing a likeness whose region was hidden: the redraw Callout, Save disabled.
10. **Leave.** Drawing mode with a region drawn and not saved, the researcher clicks Back: the existing leave confirm.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Observation dialog | Recipe | Extend | `Features/CitationComposer/CitationComposerAuthoringBox.swift` (`pvFormDialog`) | The one authoring home (CC-3). Steps aside for drawing mode the way it does for a custom term or a new subject. Not a second dialog. |
| `depiction` value area | Snowflake | Extend | `Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields` `valueControls` | A case keyed on the value type, beside text, integer, term, name, and date. Shared by the new and edit dialogs. Holds the preview, Draw / Draw again, and the no-page or hidden-region Callout. |
| Property and person pick | Component | Ship | `PVComboBox` | A depiction property is one more option (CDR-1). |
| Preview | Component | Ship | `PVThumbnail` | CDR-4. 96pt, cover. Not a new image control. |
| Draw / Draw again | Component | Ship | `PVButton` (secondary) | Enters drawing mode. |
| Drawing-mode hint | Component | Ship | `PVCallout` (info; warning for the wrong page) | Above the canvas in `Features/CitationComposer/CitationComposerViewerPane.swift`. Not a local banner or overlay. |
| Cancel drawing | Component | Ship | `PVButton` (ghost) | Beside the hint; Escape does the same. |
| Region tools | Snowflake | Ship | `Features/ArtifactViewer/ArtifactViewer.swift` `ArtifactViewerToolChrome` | The same radios. No new tool. Rectangle is armed on Draw. |
| Clear menu and Set page | Component | Ship | `PVContextMenu`, `PVButton` in the tool strip | Act on the Citation only; disabled in drawing mode. |
| Viewer | Snowflake | Ship | `Features/ArtifactViewer/ArtifactViewer.swift` | API unchanged: it shows the region its caller hands it. |
| Region target | Snowflake | Extend | `Features/CitationComposer/CitationComposerModel.swift` | Hands the viewer the Citation's region, or the depiction's while drawing or while a depiction dialog is open. |
| Locator line | Snowflake | Ship | `Features/CitationComposer/CitationComposerViewerPane.swift` | Still the Citation's locator. A depiction never appears in it. |
| Saved depiction line | Snowflake | Extend | `Features/CitationComposer/CitationComposerAuthoring.swift` `CitationSavedLine` + `Features/CitationComposer/CitationComposerSubjectGroups.swift` `lineRow` | The existing line with a thumbnail in place of the value text (CDR-7). Same actions. |
| Saved image thumbnail | Component | Ship | `Platform/CachedThumbnail.swift` → `PVThumbnail` | 32pt, cover. Loads the saved image; loading and placeholder states come from the kit. |
| Negative mark | Component | Ship | `PVBadge` | The existing Negative badge, with the italic *not*. |
| Line actions | Component | Ship | `PVIconButton`, `PVContextMenu` | Edit, Duplicate, Negate, Delete, as today. |
| Errors and notices | Component | Ship | `PVCallout` | CDR-10, CDR-11, CDR-14. |
| Leave | Component | Ship | `PVConfirm` | The existing leave confirm (CDR-12). |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit image-crop, region-editor, or likeness-row primitive | One call site. Compose Thumbnail, Button, Callout, and the existing viewer. |
| A second viewer, a viewer inside the dialog, or a zoomed face editor | The composer's viewer is the one file view (S10-D3). |
| A local drawing banner, toolbar, or overlay | The hint is a Callout with a Cancel Button. |
| New region tools or a polygon-masked preview | The saved image is the bounding box. |
| A likeness page, a likeness list, or a gallery here | The gallery is S10-D12 on the Person page. |
| Drawing every saved face on the photo | One region at a time today. |
| A caption or transcription field on the depiction | The Citation's transcription is the reading. |
| Copy that hard-codes "likeness" in the controls | Controls name the property by its label; later depiction keys reuse them. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md) with S10-34.
3. Implement **S10-34** against the board and this inventory. The plan is [`../prs/S10-34-composer-depiction-row.md`](../prs/S10-34-composer-depiction-row.md).
