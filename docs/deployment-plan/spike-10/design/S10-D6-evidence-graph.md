# S10-D6 — Evidence graph

**Kind:** Claude Design board (**enhancement** of the Evidence graph)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Evidence graph — header, palette, and the subject-card footer
**Implements later as:** PR **S10-20** (header file select, Open, and box-select). The same board specifies Unpromote, which ships in **S10-44**.
**Depends on:** —
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-20, S10-44). Dropping a file on the Source page is **[S10-D5](archive/S10-D5-source-artifacts.md)** and ships in S10-18. Hit boxes, the tiled grid, canvas growth, and the camera already shipped (S10-01…S10-04) and are not this board.
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

### This brief is an enhancement

Add three controls to the graph that exists: a file select and Open in the header, Select on the palette, and Unpromote on a promoted card. Do not draw a second graph. Do not throw away the header, the palette, the cards, the edges, pan, or zoom.

S10-20 builds the header file select, Open, and the palette Select. Draw Unpromote on this same graph anyway. S10-44 fills that frame. It does not get another board.

---

## 1. Objective

The files on this Source can be opened from the graph, without a trip to the Source page. A group of subject cards can be boxed and moved together. A promoted card can take that filing back.

**What changes.** A file select and an Open button in the header, under EG-1. A Select tool and a Shift-drag rectangle that select person, event, and place cards. Unpromote on the promoted card’s footer, with a confirm that names the bridges on this Source.

**What stays.** Promote all, the jump back to the Source page, Place, Connect, the cards, the edges, empty-canvas pan, and zoom. Card delete stays the impact confirm. The full card stays at every zoom. The Source page Artifacts section is S10-D5.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The header is a title row | “Evidence graph”, the Source title, a count of subjects and bridges, then Promote all and the button that jumps to the Source page. Both trailing buttons are secondary and small. The file select and Open join that cluster, in front of Promote all. |
| Every file, then Open | A Source can have many artifacts, and an artifact can have no file. The select lists each artifact that has a file, in Source order. Each menu row shows a 48pt thumbnail, the label (or the ref, in mono, when the label is empty), a Cover mark when that artifact is the cover, then the filename, kind, and size, then how many citations point at it. The closed select is 208pt, shows that label or ref, truncates inside the width, and its help is the filename. The initial selection is the cover artifact when that artifact has a file, otherwise the first artifact in the list that has a file. Choosing an option does not open it. Open, beside the select, is secondary and small, with an external-link icon, the label "Open", and help "Open in its app". It opens the selected file in the system app, the same action as Open on the Source page. Fileless artifacts are not in the list. No files: the select and Open are both absent, not disabled. |
| Open is the system app | The file leaves Provenencia. There is no viewer window in the app. A missing file uses the danger callout the graph already shows for a failed action. It does not open a new dialog. |
| The canvas already has modes | Idle: dragging a card moves it, dragging empty canvas pans. Place and Connect are armed from the floating palette and take the pointer. Select is a third armed mode on that palette, after Connect, split off by a hairline. The armed look matches Place and Connect. The tool uses the rectangle icon. Its help is "Hold Shift to select". |
| Select does not need a file | Place and Connect are dimmed when the Source has nothing to cite. Select moves cards that are already there, so it stays available. |
| The rectangle is document space | Hold Shift and drag on empty canvas, or drag on empty canvas while Select is armed. The rectangle is an accent stroke and a translucent accent fill, and it scales with zoom. That drag does not pan. Releasing selects every person, event, and place card whose frame intersects the rectangle. A bridge inside the rectangle is not selected. The previous selection is replaced. A rectangle that hits nothing clears the selection. |
| Pan stays | With Select disarmed and Shift up, dragging empty canvas still pans. Two-finger pan and pinch zoom stay in every mode. A drag that starts on a card never starts the rectangle. |
| One card’s move, applied to the set | Today one drag snaps a card to the grid and stores that cell. A drag on a selected subject card does that for every selected subject card, by the same delta. Arrow keys do the same, one cell at a time. Bridge cards stay put. Edges follow the ends that moved. Dragging a subject card that is not selected selects only that card and moves only it. |
| The selection chrome already exists | A selected card has an accent border and a soft ring. Each selected subject card uses that. After the pointer goes up, the rectangle is gone. There is no box left around the group. |
| Leaving Select | Escape, or clicking Select again, disarms it, the way Connect disarms. The selection remains. Arming Place or Connect disarms Select, and arming Select disarms those. A click on one subject card selects only that card. |
| The footer is one height | An unpromoted person, event, or place shows Promote on the kind band. A promoted one shows the handle: a ref badge, the name, and a chevron, and the row opens that person, event, or place. Both footers are the same height, so filing never moves an edge. Unpromote fits in that height. |
| Unpromote is the filing, not the card | It sits on a promoted person, event, or place, trailing the handle row, past a hairline, as its own hit. The handle row still opens the page. A labeled button does not fit the 264pt card, so Unpromote is a ghost icon button, small, with the U-turn-down arrow (`arrow.uturn.down`). Promote's footer icon is its opposite, the turn-up arrow (`arrow.turn.right.up`); answered 2026-10-16. The word Unpromote is the accessible label and the help. It is not a trash can. The name truncates; the button does not. Bridge cards do not show it. Unpromoted cards do not show it. |
| The confirm names the bridges | It always asks. The item is this subject plus each bridge on this Source that has an accepted claim with this subject as an end (a participation, relationship, location, or place relationship). The tone is irreversible, so the button is not a trash can. The title names the subject (“Unpromote James Robins?”). The message says this filing is removed. The detail lists each bridge by the sentence on its card. It says the cards stay on the graph, the handle stays, and the other person on a bridge stays filed. When there is no such bridge, the detail is empty and the message does not invent one. Confirm is “Unpromote”. Cancel is “Keep filing”. |
| After confirm | This card shows Promote. Each named bridge card shows Promote. The other person, event, or place on that bridge still shows their handle. The cards and the edges are still on the graph. Cancel leaves every footer as it was. |
| Card delete is a different control | The header trash can still runs the impact check. Unpromote does not. |

### 2.1 What this board is not

- An in-app window that shows the file. Open uses the system app.
- A context menu of files. The choice is a select.
- Dropping a file on the graph or on a card. The drop target is the Source page content pane (S10-D5), and it opens Add Artifact.
- Selecting or moving bridge cards with the rectangle. A bridge can still be clicked on its own, as today.
- Shift-click to add one card to the selection. An undo stack.
- A group box that stays after the drag.
- Replacing pan. Idle empty-canvas drag still pans.
- Deleting the card, deleting the handle, or rejecting the claim (“not this entity”). Other members of the handle stay filed. That review is not this button.
- Unpromote on the Person, Event, or Place page.
- A new card face, a new zoom, or a loading cover. Those are already decided.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S10-20** | EG-1…EG-7. The header file select and Open, the missing-file callout, Select, the rectangle, and moving the group. S10-01 and S10-03 have already landed. |
| Ships in **S10-44** | EG-8, EG-9, EG-10. Unpromote, the confirm, and the footers after it. |
| Stays out | An in-app file window. Dropping on the graph. Bridge cards in the marquee. Shift-click add. Undo. Deleting the card or the handle. Rejecting a claim. Unpromote from a handle page. A second graph board for the later PRs. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EG-1 | When at least one artifact has a file, the header’s trailing cluster leads with a select of those artifacts and an Open button (secondary, small, external-link icon, label “Open”, help “Open in its app”). The select’s rows, 208pt width, truncation, filename help, and initial selection are the rule in §2. Changing the selection does not open the file. Open opens the selected file in the system app, the same path as Open on the Source page. Fileless artifacts are omitted. No files: both controls are absent. |
| EG-2 | Promote all, the Source-page jump, the title, the count, Place, Connect, the no-artifact callout, the cards, the edges, pan, and zoom stay. The file select does not replace the callout that sends you to the Source page when nothing can be cited. |
| EG-3 | When the chosen file is missing, the graph’s existing danger callout says so. The title is “File not found”. The message names the artifact and the filename: “{label or ref} points at {filename}, which is no longer where Provenencia stored it.” The select still shows that artifact. No new dialog, and no window inside the app. |
| EG-4 | Select is a palette tool after Connect, past a hairline, using the rectangle icon. Armed state matches Place and Connect. Help is "Hold Shift to select". It stays available when Place and Connect are dimmed. Arming it disarms those tools, and arming those disarms it. |
| EG-5 | The rectangle is drawn on empty canvas, in document space, by holding Shift or by dragging while Select is armed. Accent stroke, translucent accent fill. That drag does not pan. On release, every person, event, and place card whose frame intersects the rectangle is selected, and the previous selection is cleared. A bridge is not selected. A miss clears the selection. Each selected card uses the existing accent border and ring. The rectangle does not remain. |
| EG-6 | With Select disarmed and Shift up, dragging empty canvas pans. Two-finger pan and pinch zoom work in every mode. A drag that starts on a card does not draw the rectangle. |
| EG-7 | Dragging a selected subject card moves every selected subject card by the same delta. Each card snaps and is stored the way one card is today. Arrow keys move the set one cell. Bridge cards stay where they are. Edges follow the ends that moved. Dragging a subject card outside the set selects only that card and moves only it. Escape, or clicking Select again, disarms the tool and leaves the selection. A click on one subject card selects only that card. |
| EG-8 | A promoted person, event, or place shows Unpromote in the footer, trailing the handle row, past a hairline, as its own hit. The handle row still opens the page. The two hits do not overlap. The footer height stays 36pt. The control is a ghost icon button, small, U-turn-down icon (`arrow.uturn.down`). The Promote footer and Promote all use the opposite turn-up arrow (`arrow.turn.right.up`). The accessible label and the help say Unpromote. There is no trash icon. The name truncates. Bridge cards and unpromoted cards do not show it. |
| EG-9 | Unpromote always opens Confirm with an item: this subject, and each bridge on this Source whose accepted claim has this subject as an end. Tone irreversible. The title names the subject. The message says the filing is removed. The detail lists those bridges by the card sentence, and says the cards stay, the handle stays, and the other end of a bridge stays filed. No such bridge: the detail is empty, and the copy does not invent one. Confirm is “Unpromote”. Cancel is “Keep filing”. Cancel changes nothing. |
| EG-10 | Confirm removes this subject’s identity claim and those bridge claims. This card and each named bridge card show Promote. The other primary cards still open their handles. The handle still exists. The cards and the edges stay. Card delete still uses the impact confirm. |

---

## 5. Suggested frames

One graph. The later frames add Select and Unpromote to that graph. Do not draw a second page.

1. **One file.** The select shows that artifact. Open sits beside it, then Promote all, then the Source page.
2. **Several files.** The select lists each artifact that has a file. The cover’s file is the initial selection when the cover has a file. A second frame shows another artifact chosen. The file has not opened yet.
3. **Open.** Open sends the selected file to the system app. The graph is unchanged.
4. **No file.** The select and Open are absent. The canvas callout that points at the Source page is still there when the Source cannot be cited.
5. **Missing file.** Open was used, the file is gone, and the graph’s danger callout is up. The select still shows the artifact.
6. **Select armed.** The palette shows Person, Event, Place, Connect, a hairline, and Select. Select is armed the way Connect arms. Help says Shift. Place and Connect are not armed.
7. **The rectangle.** Shift-drag on empty canvas covers three people and one relationship. On release the three people have the selection ring. The relationship does not. The rectangle is gone.
8. **The move.** Those three people have been dragged. They moved together. The relationship card is where it was. The edges meet the people in their new cells.
9. **Pan still pans.** Select is disarmed. A drag on empty canvas has panned. No rectangle, no new selection.
10. **Promoted footer.** A filed person shows the ref, the name, the chevron, and Unpromote. The footer is the same height as the Promote footer on the unfiled card beside it. A relationship card has no Unpromote.
11. **Confirm, one bridge.** “Unpromote James Robins?” The detail names the relationship by its card sentence. The copy says the cards stay, the handle stays, and the other person stays filed. The button is Unpromote, not a trash can.
12. **Confirm, no bridge.** The same sheet with an empty detail. It does not name a relationship.
13. **After confirm.** James’s card and the relationship card show Promote. The other person still shows their handle. Both cards are still on the canvas.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Evidence graph | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphView.swift` | The only graph. The file select, Open, and the palette Select ship in S10-20. |
| Header cluster | Snowflake | **Extend** | `header` in `EvidenceGraphView` | EG-1. File select, Open, then the two buttons that exist. |
| File select | Component | Ship | `PVSelect` | EG-1. 208pt, size small. Menu width 360, row height 64. Artifacts that have a file. Closed label truncates. Help is the filename. Not the palette Select tool. |
| Open | Component | Ship | `PVButton` | Secondary, small, external-link icon, label “Open”, help “Open in its app”. Opens the selected file in the system app. |
| Promote all | Component | Ship | `PVButton` | Stays. |
| Source page jump | Component | Ship | `PVButton` | Stays. |
| No-artifact callout | Component | Ship | `PVCallout` | EG-2. Stays. It is not an Open button. |
| Missing-file callout | Component | Ship | `PVCallout` (danger) | EG-3. The slot the graph already uses for a failed action. |
| Palette | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphPalette.swift` | EG-4. Select after Connect. |
| Select | Component | Ship | palette button, rectangle icon (`PVSymbol.regionRectangle`) | Armed state matches Place and Connect. Help is “Hold Shift to select”. Not a new kit tool. |
| Rectangle | Snowflake | **Extend** | the canvas document | EG-5. Accent stroke and translucent fill, in document space. Not a kit primitive. |
| Selection ring | Snowflake | **Extend** | `EvidenceSubjectCard` selection chrome | The existing accent border and ring, on each selected subject. |
| Subject card | Snowflake | Ship | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Unchanged except the footer in EG-8 and the shared selection ring. |
| Unpromote | Component | Ship | `PVIconButton` ghost, small, `arrow.uturn.down` (Promote uses `arrow.turn.right.up`) | EG-8. On the promoted footer, past a hairline. Accessible label and help “Unpromote”. No trash icon. Its hit is not the handle hit. S10-44. |
| Handle row | Snowflake | Ship | membership footer in `EvidenceSubjectCard` | Stays the open-handle control. |
| Promote footer | Snowflake | Ship | `promoteFooter` in `EvidenceSubjectCard` | Stays on unpromoted cards, and returns after EG-10. |
| Unpromote confirm | Component | Ship | `PVConfirm` (`item:`, tone irreversible) | EG-9. Not DeleteImpact. Not a Bool. |
| Bridge card | Snowflake | Ship | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Not in the marquee. No Unpromote. Stays put when the group moves. |
| Card delete | Recipe | Ship | `PVDeleteImpact` | Stays the trash can. Unpromote does not use it. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit marquee tool, a group-box component, or an Unpromote button primitive | One graph. Compose Button, Select, Confirm, and the card that exists. |
| An in-app viewer window, or a context menu of files | EG-1. The choice is `PVSelect`. Open leaves the app. |
| Bridge cards in the rectangle, or a drag that moves them | EG-5 and EG-7. Edges follow the ends. |
| Trash-can confirm, or DeleteImpact, for Unpromote | The card and the handle stay. Tone is irreversible. |
| A second graph for S10-20 or S10-44 | Those PRs fill frames 6–13. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-20** against EG-1…EG-7. Implement **S10-44** against EG-8…EG-10. The Source page drop is [S10-D5](archive/S10-D5-source-artifacts.md).
