# S10-D5 — Source page, Artifacts

**Landed 2026-10-14** in [#387](https://github.com/mendahu/provenencia/pull/387). The sections below are the original board prompt.

**Kind:** Claude Design board (**rethink** of the artifact list; the rest of the Source page stays)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Source page — the content pane, and the artifact cards on it
**Implements later as:** PR **S10-18**
**Depends on:** —
**Related:** [`../../deployment-plan.md`](../../deployment-plan.md) (S10-18). The Evidence graph is **[S10-D6](../S10-D6-evidence-graph.md)** and ships in S10-20. Box-select, Unpromote, and the graph’s file select and Open button are that board.
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.agents/skills/add-ui-component/SKILL.md)

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

### This brief replaces the artifact list

Throw away the accordion. Do not keep the chevron, the expanding row, or the primary-file column. Draw artifact cards in wrapping rows on the Source page that exists. Do not draw a second Source page. The identity header, description, credibility, metadata, and notes stay.

The drop target is new on that same page. It is not a second page.

---

## 1. Objective

A scan still has to be named before it is stored, and the way to open it should not be an expand. This board lets a drop anywhere on the Source page open Add Artifact with that file already chosen, and it replaces the stacked rows with cards.

**What changes.** The content pane highlights and accepts a drop. The drop opens Add Artifact; the file row shows the file; the label and description are empty. Add Artifact is what stores it. Artifacts are cards: a large thumbnail or icon, the label, the description, and actions along the bottom for open, use as thumbnail, and delete. How the label and description enter edit mode is this board’s to draw.

**What stays.** Add Artifact is the same dialog: label required, description optional, file optional, the existing drop row inside it, the existing reject callout. Add File still attaches a first file to an artifact that has none. Cover, delete, and the `ART-` ref stay. Description, metadata, notes, and the identity header are not restyled. The Evidence graph is a different board.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The content pane is the Source page | The identity header and the scrolling column under it: description, credibility, metadata, artifacts, notes. That whole pane is the drop target. The sidebar is not. The Evidence graph is not. A sheet that is already open is not; its own controls take the drop. |
| A drop fills the dialog | One file that passes the existing checks opens Add Artifact. The file row shows that filename, the way Choose file or a drop inside the dialog already does. The label is empty. The description is empty. Nothing is stored until Add Artifact. Cancel, or a dirty-leave confirm that discards, leaves the list as it was. Clearing the file in the dialog is still allowed; Add can then create an artifact with no file, as it does today. |
| A bad drop still opens the dialog | More than one file, a folder, an empty file, a type outside the list, or a file over 512 MB uses the reject callout the dialog already shows. Add stays disabled. No artifact is created. The allowed types stay the caption inside the dialog. The page does not grow its own. |
| The highlight is only while the file is over the pane | Soft accent fill and accent border on the content pane, and a short hint that the drop opens Add Artifact. On release the highlight is gone and the dialog is up. There is no standing drop zone and no types line on the page. |
| The list is cards, in rows | Wrapping rows of cards at the content width. A card is a raised card. The top is a large preview: the file thumbnail, or the type icon when there is no thumbnail. Under that, the label (or the ref when the label is empty) and the description. An empty description does not collapse that card into a different shape from its neighbors. |
| The bottom of the card is the actions | Open, Use as thumbnail, and Delete. The board places them. Open is shown when the artifact has a file, and it opens that file in the system app. An artifact with no file has no Open control; Add file, the existing first-attach dialog, covers that case. Use as thumbnail is shown when this artifact can become the cover. Delete is always there and still confirms before it removes the artifact. |
| Cover is a mark, not a second card | The artifact that is the cover wears the Cover badge. Use as thumbnail is absent on that card. Setting the cover does not reorder the cards. |
| The ref stays | The `ART-` ref is on the card. The board chooses where. It is not dropped to make room for the preview. |
| Editing is on the card | Label and description are edited from the card. The board chooses the affordance: a pencil, a click on the text, or another control already in the kit. Label stays required. Description stays optional. Save writes the artifact. Cancel restores the stored text. There is no expanded panel under the card. |
| Add Artifact on the heading stays | The section heading still offers Add Artifact, which opens the same dialog with an empty file row. The empty state stays when there are no artifacts, and it is inside the drop pane. |

### 2.1 What this board is not

- A drop that stores the file by itself. The researcher types a label and confirms Add Artifact.
- A drop that attaches to the card under the pointer. A first file on an existing artifact is still Add file.
- A drop on the sidebar or the Evidence graph.
- The accordion: chevron, one expanded row, primary-file column.
- A kit primitive for an artifact card or a drop zone. Compose Card, Thumbnail, Button, and the dialog that exists.
- A redesign of description, metadata, notes, or the identity header. They are in the drop target, and they look as they do now.
- Box-select, Unpromote, or the graph’s Open button. That is S10-D6.

---

## 3. Implementation gate (S10-18)

| Ships in **S10-18** | Stays out |
| --- | --- |
| The Source page content pane is the drop target. A drop opens Add Artifact with that file in the file row and the label empty. Add Artifact stores it. A rejected drop opens the dialog on the existing callout and stores nothing. Artifacts are wrapping rows of cards: large preview, label, description, ref, Cover when it applies, and Open / Use as thumbnail / Delete along the bottom, with the enablement in §2. Label and description are editable from the card, by the affordance the board draws. | Creating an artifact on drop. Prefilling the label with the filename. Attaching to the card under the pointer. Keeping the accordion. A new artifact model. The Evidence graph (S10-D6, S10-20). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| AR-1 | The drop target is the Source page content pane: identity header and the scrolling column (description, credibility, metadata, artifacts, notes). A drop on the sidebar or the Evidence graph does nothing. A sheet that is already open keeps its own drop. |
| AR-2 | While a file is over that pane, the pane shows the accent highlight and a short hint that releasing opens Add Artifact. The highlight is gone once the dialog is open. The page has no permanent drop well and no allowed-types line. |
| AR-3 | One file that passes the existing checks opens Add Artifact. The file row shows the filename. Label and description are empty. The file is stored only when the researcher enters a label and confirms Add Artifact. Cancel leaves the cards unchanged. |
| AR-4 | A file the existing checks reject, including more than one file, opens Add Artifact with the existing danger callout. Add is disabled. No card is added. |
| AR-5 | Artifacts are cards in wrapping rows. Each card leads with a large preview: thumbnail, or the type icon when there is no thumbnail. Under it, the label and the description. Cards in a row share a structure, including when the description is empty. |
| AR-6 | The card’s bottom offers Open, Use as thumbnail, and Delete. Open appears only when a file exists, and opens it in the system app. Use as thumbnail appears only when this artifact can become the cover. Delete always confirms, with the confirm the page already uses. A fileless card offers Add file instead of Open. |
| AR-7 | The cover card shows the Cover badge and does not offer Use as thumbnail. The `ART-` ref is on every card. |
| AR-8 | Label and description are editable on the card. The board chooses how edit mode starts. Label is required. Save and cancel stay on the card. There is no accordion and no primary-file column. |
| AR-9 | The heading’s Add Artifact button and the empty state stay. That button opens the dialog with no file chosen. The dialog’s own file row, reject callout, and Clear stay. |

---

## 5. Suggested frames

Draw the Source page content width. Show enough of the header, the overview, and the notes that the drop clearly covers them. The artifact block is the redesign. Do not restyle the overview.

1. **Cards, idle.** Three artifacts in a wrapping row. One is the cover. Previews are large. Labels and descriptions read without expanding. Open, Use as thumbnail, and Delete sit on the bottom. The cover card has the badge and no Use as thumbnail.
2. **Fileless card.** The preview is the type icon. Add file is where Open would be. Delete is there. No disabled Open.
3. **File over the page.** The pointer is on the metadata column, not on an artifact. The whole content pane is highlighted. The hint is up. The cards are unchanged.
4. **Add Artifact, file chosen.** The dialog is open. The file row shows the dropped filename. Label and description are empty. Add is disabled until the label has text.
5. **Rejected.** A zip, or two files, was dropped. The dialog is open, the danger callout is in it, and no new card exists.
6. **Cancel.** The dialog is gone. The cards match frame 1.
7. **After Add.** A new card sits with the others. The toast is the existing created-with-file toast. The preview is that file.
8. **Editing.** The board’s edit affordance is active on one card. The label and description are fields. Save and cancel are on that card. The other cards are unchanged. There is no row expanded beneath it.
9. **Empty page.** No artifacts. The empty state is inside the highlighted pane while a file is over the page. Releasing opens the same dialog as frame 4.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Source page | Snowflake | **Extend** | `Features/Sources/SourcePageView.swift` | AR-1, AR-2. The content pane is the drop target. Do not restyle the header, overview, or notes. |
| Artifact list | Snowflake | **Rethink** | `Features/Sources/SourcePageArtifactsView.swift` | Replace the accordion with cards. AR-5 through AR-8. |
| Section heading | Component | Ship | `PVSectionHeader` | Title, count, Add Artifact. AR-9. |
| Empty artifacts | Component | Ship | `PVEmptyState` | Stays, inside the drop pane. |
| Artifact card | Component | Ship | `PVCard` | One card per artifact. Not a new card type. |
| Preview | Component | Ship | `PVThumbnail` via `CachedThumbnail` | Large. Type icon when there is no thumbnail. |
| Label and description | Component | Ship | `PVField`, `PVInput`, `PVTextArea` | Shown on the card. The edit affordance is the board’s. Fields while editing. |
| Open | Component | Ship | `PVButton` | AR-6. External-link icon, label “Open”. Only when a file exists. |
| Use as thumbnail | Component | Ship | `PVButton` | AR-6, AR-7. Hidden when it cannot become the cover, including when it already is. |
| Delete | Component | Ship | `PVButton` or `PVIconButton` | Danger. The existing delete confirm. The board chooses button or icon button. |
| Add file | Component | Ship | `PVButton` | Fileless card. Opens the existing Add File dialog. |
| Cover | Component | Ship | `PVBadge` | On the cover card. |
| Ref | Snowflake | Ship | mono text on the card | AR-7. The board places it. |
| Drop highlight | Snowflake | **Extend** | the content pane | AR-2. Accent fill and border. Not a kit drop zone, and not a second `IngestFileDropRow` on the page. |
| Add Artifact dialog | Recipe | Ship | `pvFormDialog` | AR-3, AR-4. Opened by the page drop with the file staged. Not redrawn. |
| Dialog file row | Snowflake | Ship | `Features/Sources/IngestFileDropRow.swift` | Stays inside the dialog, including the reject callout. |
| Toast | Component | Ship | `PVToast` | The existing toast after Add. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit artifact-card, card grid, or page drop zone | One page. Compose Card, Thumbnail, Button, Field. |
| An upload that skips the dialog | AR-3. The label is typed, then Add Artifact. |
| The accordion, or a primary-file column beside the card | AR-8. Edit happens on the card. |
| A types caption on the page | The dialog already has it. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../../completed.md`](../../completed.md).
3. Implement **S10-18** against this board. The Evidence graph is [S10-D6](../S10-D6-evidence-graph.md) and is not this PR.
