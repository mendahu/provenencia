# S8-D8 — Composer and Connect simplification

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-11**  
**Depends on:** **S8-D7** / **S8-10** (shipped Citation-document composer, Layout A + 1500pt two-column; brief [`archive/S8-D7-composer-rethink.md`](archive/S8-D7-composer-rethink.md)); shipped Evidence graph Connect tool and create dialog  
**Related:** [`S8-D6-delete-paths.md`](S8-D6-delete-paths.md) (this board decides **single Observation delete** from the composer; S8-D6 keeps subject / bridge / Citation cascades); [`docs/audit-revision-history.md`](../../../audit-revision-history.md) (one research action = one revision); [`docs/interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) §6 (bridges are cited subject-valued edges)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); navigation via [`add-workspace-location`](../../../../.cursor/skills/add-workspace-location/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the **S8-D7** composer and the shipped Evidence graph. Keep Layout A, the 1500pt two-column split, the viewer, the identity line, and the observation stack. Change **how work is saved**, **how Connect reaches the composer**, and **a few graph dialogs**. Do not restyle anything this brief does not name.

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
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

---

## 1. Objective

An adversarial review of the shipped composer and graph found that several design choices add a lot of state and edge cases for little researcher value. Some of them already lose or corrupt data. This board redesigns those choices. The product model does not change: the Citation is still the document, Observations are still rows on it, and Connect is still one atomic write.

Five changes, in priority order:

1. **Save Observations one row at a time.** Today one footer **Save** sends the Citation reading plus every row, and the engine **deletes any saved row the payload leaves out**. Removing a row, or changing its subject to one whose type does not allow the property, silently deletes a persisted Observation with no confirm. After this board, each row has its own **Save** and **Revert**, and deleting a saved row asks first. The reading (transcription, uncertainty, note, description, locator) keeps its own **Save reading**.
2. **Guard every exit while work is unsaved.** Today only switching Artifact on a dirty saved Citation asks. Cancel, Back, the sidebar, the omnibar, and switching Citation all throw edits away silently. After this board, any navigation away from a composer with unsaved work asks **Discard / Keep editing**. While work is unsaved, the Artifact and Citation menus are **disabled** (with a reason) instead of offering an abandon confirm.
3. **Connect goes straight to the composer, and the composer finishes it as one unit.** The graph's role / relationship-type sheet is **removed**. The composer shows a **connection row**: a single compact row in the observation stack, not three property rows. It reads as one assertion — the bridge sentence with its two endpoints ("John Robins · Birth 1850") as read-only text, plus the one thing the researcher chooses: the role or relationship type. **Save connection** commits it. The endpoints come from the graph pick and **cannot be changed** in the composer. A wrong endpoint means **Discard connection** before saving, or deleting the bridge (S8-D6 / S8-09) and connecting again after. The connection can attach to a **new or an existing** Citation on the Artifact. Today, picking an existing Citation in Connect mode silently drops the connection.
4. **Bridges have no typed label.** A bridge's name is its sentence ("John Robins was born in 1850 London"), computed from its cited edges every time it is shown. The bridge **Edit** dialog edits **description only**. Pickers that list bridges show the sentence. When an endpoint has no cited name / event type / toponym, the sentence still reads cleanly through a fixed **fallback chain** (§2.4): each endpoint falls back to its working label, then its type + ref. If the edges cannot be read at all, the name falls back to a bare kind phrase plus the bridge's ref. A bridge is never shown with an empty name.
5. **Create a primary subject from the composer row.** The row's subject picker gets **New person… / New event… / New place…**. It opens a short FormDialog (label, optional description), creates the subject on this Source's graph, and fills the row. This removes the trip "cancel draft → graph → place → Add property → find the Citation again". **This reverses an S8-D7 non-goal on purpose.** It is limited to the three primary kinds. Bridges still come only from Connect.

Also, as copy only: the graph's primary **create / edit dialog** explains that the label is a **working label**, and that the card shows the cited name once there is one.

```text
Before                                         After
──────                                         ─────
[ Save ] sends reading + all rows              Reading   … [ Save reading ]
  (omitted saved rows are deleted)             Row  Mary · birth_date · 1850   [✓] [↺] [⋯]
                                               Row  (draft) John · name · …     [✓] [↺] [⋯]
Cancel / Back discard silently                 Any exit while dirty → Discard / Keep editing
Connect → role sheet → composer (3 fixed rows) Connect → composer: one connection row
  existing Citation picked → bridge dropped      ⇄ John Robins · Birth 1850   Role [child ▾]  [Save connection] [Discard]
Bridge label typed / stored                    Bridge title = cited sentence, always computed
Need a new person → leave the composer         Subject ▾ … New person…  (FormDialog, stays in place)
```

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| A Citation must exist before an Observation can | On a **New** Citation, the **first** commit (row Save, Save connection, or Save reading) creates the Citation from the current reading draft. After that the identity line shows the new `CIT-…` and further commits attach to it. |
| Empty Citation is legal | **Save reading** on a New Citation with no rows is a normal save, not an error. |
| A saved Citation cannot change Artifact | Artifact is fixed once the Citation is saved. Switching Artifact means "start another Citation on that Artifact". Allowed only when nothing is unsaved. |
| One row = one Observation = one audited change | Row **Save** writes that row only. Row **Delete** removes that Observation only and never touches the Citation or other rows. Deleting the **last** row leaves an empty Citation (legal). |
| Row commits do not save the reading | On a saved Citation, a dirty reading stays dirty after a row Save. The two are independent. The **only** time a row commit writes the reading is when it creates the Citation (first commit on New). |
| Connect is atomic | The bridge subject, its grid position, its two endpoint Observations, its role / relationship-type Observation, and (if New) the Citation are written together by **Save connection**. Nothing about the connection persists before that. |
| A connection is one unit | Its two endpoint Observations are structural and **immutable**: no subject, Property, value, or polarity change, and no delete. The engine enforces this, not just the UI. The composer never shows them as separate property rows — before and after save, the bridge's endpoints and role appear as **one connection row**. |
| Wrong endpoint = redo the connection | Before save: **Discard connection**, go back to the graph, Connect again. After save: delete the bridge from the graph (**S8-D6** / **S8-09** designs that delete) and Connect again. The composer offers no endpoint editing. |
| Role / relationship type is the only choice | The connection cannot be saved without it. After save, the connection row still lets the researcher change the role (row Save / Revert, like any row). The connection row has **no** Delete. |
| Bridge sentence is derived | Show `EvidenceBridgeEdgeSummary`'s sentence wherever a bridge is named (card title, composer subject picker, connection row, breadcrumb). No stored bridge label is edited. Missing nouns follow §2.4; the name is never blank. |
| Grid position is layout, not evidence | A new bridge lands at the midpoint of its two endpoints (engine-computed). A primary created from the composer lands to the right of the existing graph (§3.6). Neither is a research assertion. |
| Delete scope | This board ships **single Observation delete from the composer** only. Subject / bridge / whole-Citation deletes and their counted cascades stay on **S8-D6**. |

### 2.1 Row states (the core of this board)

Every editable Observation row is in exactly one state. The board must draw each one.

| State | Meaning | Controls |
| --- | --- | --- |
| **Draft** | New row, never saved. Property may be empty. | Save (disabled until valid), Revert = remove the draft (no confirm), ⋯ menu |
| **Saved** | Matches the catalog. | ⋯ menu (Delete…, Negate/Affirm) — Save/Revert hidden |
| **Edited** | Saved row with local changes. | Save, Revert (back to saved values, no confirm) |
| **Saving** | Commit in flight. | Save shows busy; row inputs disabled |
| **Error** | Last commit failed. | Error text **on the row** (Field error / compact Callout), Save + Revert still available |

A **Draft** row with no Property picked does not count as unsaved document work (it is an empty placeholder). Every other Draft or Edited row counts (§2.3).

Polarity toggle on a **Saved** row is an edit: the row becomes **Edited** until saved. Do not auto-commit polarity.

### 2.2 Connection row (Connect entry, and saved bridges)

One compact row type, used in two states. It is **not** three stacked property rows, and it is not a big grouped card. It should read like a single property line.

```text
Pending   ⇄  John Robins · Birth 1850        Role [ — pick — ▾ ]   [Save connection] [Discard]
             Will be saved to CIT-7Q2…   (only when an existing Citation is selected)

Saved     ⇄  John Robins · child · Birth 1850   Role [ child ▾ ]   (Save / Revert appear only when the role is edited)
```

- **Leading mark + sentence.** A connection glyph (reuse the bridge-kind icon the graph uses) and the bridge sentence with both endpoint names as **read-only text**. No endpoint pickers, no Property labels (`person`, `event`, …), no per-edge values.
- **Role / relationship type** is the only control: the existing inline term picker + "Add term…" dialog.
- **Pending** (Connect entry, nothing written yet): sits at the **top** of the observation stack, above the Citation's existing rows. Trailing actions **Save connection** (primary, small) and **Discard** (ghost). Save connection is disabled until a term is picked. On success the row switches to **Saved** in place.
- **Saved** (the connection's Citation is open — after Save connection, or when the composer opens the bridge's connect Citation): the same row, showing the saved role. Changing the role makes it **Edited** with row Save / Revert (§2.1). There is **no** Delete and no polarity toggle in its ⋯ menu (the menu may be omitted entirely).
- The Citation's other rows show below in their normal states and keep their own row Saves.
- **Discard** removes the pending row without writing. The composer becomes a plain Add-property-style composer on the same Citation. If the Citation is still New and has no other work, Done returns to the graph with nothing written.
- Other role Observations on a bridge (for example a second role cited from a different Citation) are ordinary term rows on their own Citation. Only the connect Citation's edges + its role Observation form the connection row.

### 2.3 Unsaved-work guard

Two terms, used exactly this way in the PR:

- **Unsaved document work** = dirty reading **or** any Draft-with-Property / Edited row.
- **Touched pending connection** = the researcher picked a term on the pending connection row. An untouched pending row (endpoints from the graph, no term yet) is **not** touched.

| Exit | Behavior |
| --- | --- |
| **Done** (footer; replaces Cancel) | Confirm **Discard changes** / **Keep editing** when there is unsaved document work **or** a touched pending connection. Otherwise leave. |
| Toolbar Back / Forward / history jump menu | Same rule and confirm; on Discard, the navigation proceeds |
| Sidebar, omnibar hit, breadcrumb, any other in-app place change | Same rule and confirm; on Discard, the navigation proceeds |
| Artifact menu, Citation menu | **Disabled** with a one-line hint ("Save or discard changes to switch reading") when there is **unsaved document work**. A pending connection (touched or not) never disables them — it moves with the switch. No confirm. |
| Save connection / row Save / Save reading | Not an exit. No confirm. |

Not guarded in this story: quitting the app, closing the window, switching project, signing out. Say so on the board (a quiet note, not chrome).

The confirm body counts what is lost, e.g. "Unsaved: the reading, 2 observations, a new connection."

### 2.4 Bridge name fallback

A bridge's name is built from **nouns** (one per endpoint) and the **role / relationship type** term. Each noun resolves independently, top to bottom, and stops at the first non-empty value:

| Tier | Noun source | Example | Who ships it |
| --- | --- | --- | --- |
| 1 | The endpoint's **identity Observation**: `name` (person, NameValue form), `event_type` (event, term label), `toponym` (place). With several, the first in card display order. | "John Robins", "Birth", "London" | **S8-06** (already scoped) |
| 2 | The endpoint's **working label** (`subjects.label`, trimmed). | "Person 3", "Baptism?" | today; kept |
| 3 | The endpoint's **type label + ref**. | "Person CPR-F4N2P" | **S8-11** |

The endpoint always exists, because the engine refuses to delete a subject that an edge points at, so tier 3 always resolves. Tier 3 is also what shows when the working label is blank. The graph dialog requires a label, but other writers (import, older catalogs) may not have one.

The **role** term is optional in the sentence. When it is missing, use the existing role-less template ("John Robins — Birth 1850"). Never print a placeholder like "(no role)".

**Whole-name fallback** (the edges themselves cannot be read: an older bridge with no cited edges on this Source, or the graph snapshot has not loaded the endpoint yet):

| Tier | Show | Example |
| --- | --- | --- |
| A | The bridge's **stored label** if one exists (older bridges only; new bridges store none). Read-only. | "Mary's baptism" |
| B | **Kind phrase** (+ role term if known) **· bridge ref** | "Participation · CPA-7K2QD", "Participation as godparent · CPA-7K2QD" |

The ref is always the picker subtext and the VoiceOver hint, so two bridges that fall back to the same words are still distinguishable.

### 2.5 What this board is not

- Not auto-save / save-on-blur. Every commit is an explicit Save (or Return in an inline field).
- Not undo / ⌘Z.
- Not deleting subjects, bridges, or whole Citations (**S8-D6**).
- Not creating bridges or `source` subjects from the composer. Only primaries (person / event / place).
- Not Auto Transcribe (**S8-D1**) or PDF Find / paste (**S8-D2**). Leave the transcription Field's trailing slot as it is; **S8-D1** designs against this board's reading section.
- Not graph visual badges or bridge sentence wording (**S8-D3**). This board only decides that the sentence replaces the stored label.
- Not changing Connect rules, edge Properties, or which pairs are refused.
- Not a new navigation mode or place. The composer stays the same `WorkspaceLocation` surface.

---

## 3. Flows to draw

### 3.1 Add property → New Citation → first row commit

1. Entry pre-selects Artifact (if one), **New** Citation, one Draft row aimed at the card's subject.
2. Researcher types transcription, picks a Property, fills the value, presses row **Save**.
3. Citation and Observation are created together. Identity line changes from "New" to `CIT-…`. Reading is now saved. Row is **Saved**.
4. Researcher adds a second row (Add observation), saves it. Only that row is written.
5. **Done** returns to the graph (no confirm; nothing unsaved).

### 3.2 Pencil → saved Citation → edit one row, delete another

1. Entry loads the Citation, all rows **Saved**, focus on the pencil's row.
2. Researcher edits the value → row **Edited** → **Save** → **Saved**.
3. Researcher opens another row's ⋯ → **Delete observation…** → Confirm ("Delete OBS-… ? This removes one observation from CIT-…. The reading and other observations stay.") → row disappears.
4. If that was the last row, the stack shows the quiet empty hint (Citation stays).

### 3.3 Change a saved row's subject to an incompatible type

1. Row is **Saved** (Mary · birth_date · 1850). Researcher picks subject "London" (a place).
2. `birth_date` is not a place Property. The row becomes **Edited** with Property **cleared** and the old value kept in memory. The row shows an inline error on Property ("Pick a property for this subject").
3. **Save** is disabled until a valid Property + value is chosen. **Revert** restores Mary · birth_date · 1850. **Nothing is deleted** unless the researcher explicitly chooses Delete.

### 3.4 Connect → composer → Save connection on a New Citation

1. On the graph, Connect: pick Person A, pick Event B. The pair is valid. **No sheet.** Navigate straight to the composer.
2. Composer: New Citation, one pending connection row at the top: "⇄ John Robins · Birth 1850", Role empty, Save connection disabled.
3. Researcher fills transcription, picks role "child", **Save connection**.
4. The same row switches to **Saved** ("John Robins · child · Birth 1850"). Identity shows `CIT-…`. **Done** → graph shows the new bridge at the endpoints' midpoint.

### 3.5 Connect → attach to an existing Citation

1. Same entry. Researcher opens the Citation menu (enabled: there is no unsaved document work; a pending connection never disables it) and picks `CIT-7Q2…`.
2. Transcription and existing rows load; the pending connection row stays on top with the role the researcher already picked (if any).
3. **Save connection** attaches the connection to `CIT-7Q2…`. No new Citation is minted.

The board must make it obvious in step 2 that the connection will be saved onto the chosen reading (the "Will be saved to CIT-7Q2…" caption under the row).

### 3.6 Create a person from the row

1. Draft row → Subject picker → type "Wil" → no match → **New person…** (also New event…, New place… at the bottom of the list, always visible).
2. FormDialog: Label (prefilled with the typed text), Description (optional). Confirm.
3. The subject is created on this Source's graph. It lands in a column just to the right of the existing cards, level with the topmost card, and further creates in the same composer visit stack downward in that column (client placement rule `EvidenceGraphPlacement` in S8-11; the board does not draw the graph for this). The row's subject is set to it. The draft is otherwise unchanged. Nothing else on the composer resets.
4. Failure keeps the dialog open with the error on the dialog.

### 3.7 Wrong endpoint

1. Pending connection row reads "John Robins · Birth 1850", but the researcher meant William. The endpoint names are plain text; nothing on the row edits them.
2. Researcher presses **Discard**, then **Done** (no confirm if nothing else is unsaved), and Connects William → Birth 1850 on the graph.
3. After save, the only path is deleting the bridge on the graph (S8-D6 / S8-09) and connecting again. The board does not draw that delete; it only confirms the saved connection row offers no endpoint edit and no Delete.

### 3.8 Guarded exit

1. Reading is dirty and one row is Edited. Researcher clicks the sidebar "Sources".
2. Confirm: "Discard unsaved changes? Unsaved: the reading, 1 observation." **Discard changes** / **Keep editing**.
3. Keep editing → stays, nothing lost. Discard → navigates to Sources.

---

## 4. Implementation gate (S8-11)

| Ships in **S8-11** | Does **not** ship there |
| --- | --- |
| Row-level commit (Save / Revert / Delete… with confirm) for every editable row | Auto-save, save-on-blur, ⌘Z |
| **Save reading** for reading fields; first commit on New creates the Citation | Moving a saved Citation to another Artifact |
| Unsaved-work guard on Done / Back / Forward / history / sidebar / omnibar / any in-app `go(to:)` | Guarding app quit, window close, project switch, sign-out |
| Artifact / Citation menus disabled while unsaved; abandon confirm **removed** | A Citation-level delete |
| Connection row (pending + saved); **Save connection**; attach to New **or** existing Citation; role editable after save | Graph role / relationship-type sheet (removed) |
| Endpoint Observations immutable (engine-enforced) | Any endpoint edit, retarget, or per-edge row in the composer; deleting a saved connection from the composer |
| Computed bridge sentence everywhere; bridge Edit = description only; no stored label for new bridges | Rewriting existing stored bridge labels in the catalog |
| New person / event / place from the row subject picker | New bridge or `source` subject from the composer |
| Working-label hint copy in the graph create / edit dialog | Card title changes beyond the hint (S8-D3 owns card chrome) |
| Engine + audit + FFI changes listed in the S8-11 plan | Schema migrations (none needed) |

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| CS-1 | Every editable Observation row shows its state (§2.1). Draft, Saved, Edited, Saving, and Error are visually distinct without relying on color alone (icon, label, or text). |
| CS-2 | Row **Save** is a `PVIconButton` (checkmark) with an accessible name "Save observation"; **Revert** is a `PVIconButton` (counter-clockwise arrow) "Revert observation". Both are hidden on Saved rows. Return in an inline text / integer field triggers Save when valid. Esc in an inline field triggers Revert. |
| CS-3 | Row errors render **on the row** (the Field's error slot or a compact `PVCallout` under the row). The footer no longer shows observation errors. |
| CS-4 | **Delete observation…** lives in the row ⋯ `ContextMenu` for Saved and Edited rows, opens `.pvConfirm(item:)` with the Observation ref and Citation ref, and deletes only that row. Draft rows use Revert (no confirm). The connection row has no Delete. |
| CS-5 | The reading section gets **Save reading** (`PVButton` secondary, small) under the description field, enabled only when the reading is dirty (or always on a New Citation with no rows, so an empty reading-only Citation can be saved). The footer primary **Save** is removed. |
| CS-6 | The footer has **Done** (secondary) only, plus the unsaved summary text when there is unsaved work ("Unsaved: reading, 2 observations"). |
| CS-7 | Unsaved-work guard per §2.3, one `.pvConfirm(item:)` for every exit, counting what is lost. |
| CS-8 | While unsaved document work exists (§2.3), the Artifact and Citation `PVSelect`s are disabled and show the hint. A pending connection never disables them. No abandon confirm remains. |
| CS-9 | Connect entry shows one pending **connection row** (§2.2) at the top of the stack. No graph sheet precedes it. When an existing Citation is selected, the row states which Citation it will be saved to. |
| CS-10 | **Save connection** is disabled until the term is set. The role / relationship-type control reuses the existing inline term picker + "Add term…" dialog. |
| CS-11 | The connection row is **one line**: connection glyph, bridge sentence with endpoint names as read-only text, role picker, trailing actions. It never shows endpoint pickers, edge Property labels, or one row per edge. Saved connection rows have no Delete and no polarity toggle; changing the role uses row Save / Revert. |
| CS-12 | Bridges are named by their computed sentence in the card title, the composer subject picker, the connection row, and the composer breadcrumb title. Missing nouns and unreadable edges follow the §2.4 fallback chain; no surface ever shows a blank bridge name. The bridge **Edit** FormDialog has Description only. |
| CS-13 | Row subject `PVComboBox` has three fixed trailing options **New person… / New event… / New place…** (ordinary rows only; the connection row has no subject picker). Each opens a `pvFormDialog` with Label (required, prefilled from the query) and Description. Success fills the row's subject; failure shows the error in the dialog. |
| CS-14 | Graph create / edit dialog for primaries shows a caption under Label: "Working label. Cards show the cited name once one is recorded." (final copy on the board). |
| CS-15 | VoiceOver: row state is part of the row's accessibility value; Save / Revert / Delete are named; the connection row is one element named with its sentence and role; the guard confirm announces what is lost. |
| CS-16 | L10n for every new string. `Text(verbatim:)` for refs and counts. |
| CS-17 | No new DesignSystem component. Everything composes existing kit (see inventory). |

---

## 6. Suggested frames

1. **Add property, New Citation** — Draft row focused, reading empty, Save reading enabled.
2. **After first row Save** — identity shows `CIT-…`, row Saved, second Draft row being filled.
3. **Row states** — one stack showing Saved, Edited, Saving, Error, and Draft together (reference frame).
4. **Delete observation confirm** — shared Citation with rows on two subjects; delete one.
5. **Incompatible subject change** — Property cleared, inline error, Revert available (§3.3).
6. **Pending connection, New Citation** — one connection row on top, role empty, Save connection disabled.
7. **Pending connection, existing Citation** — connection row on top with "Will be saved to CIT-…", existing rows below.
8. **After Save connection** — the same row in Saved state; then the role changed (Edited, Save / Revert visible).
9. **Connection row next to ordinary rows** — reference frame showing it reads like one property line in the stack (compare density with frame 3).
10. **New person from row** — ComboBox with New… options; FormDialog open.
11. **Unsaved guard** — confirm from a sidebar click, counts listed.
12. **Menus disabled while unsaved** — Artifact + Citation selects disabled with hint.
13. **Bridge Edit dialog** — description only.
13b. **Bridge name fallbacks** — one bridge card + its picker row at each tier: working label ("Person 3 — Birth"), type + ref ("Person CPR-F4N2P — Birth"), and whole-name fallback ("Participation · CPA-7K2QD").
14. **Graph create dialog** — working-label caption.
15. **Wide (≥1500pt) variant** of frames 2 and 7 — reading | observation stack split still works with Save reading on the reading side.

---

## 7. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Composer place | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerView.swift` | Hosts the guard confirm and the new-subject FormDialog. Abandon confirm removed. |
| Form pane | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerFormPane.swift` | Save reading under the reading fields; footer = Done + unsaved summary; connection row(s) at the top of the stack. |
| Observation row | Snowflake | **Extend** | `Features/CitationComposer/CitationComposerObservationRow.swift` | State marker, Save / Revert IconButtons, row error, Delete in ⋯. Ordinary rows only — edge Observations never render here. |
| Connection row | Snowflake | **New** (feature-private) | `Features/CitationComposer/CitationComposerConnectionRow.swift` | One `PVCard` row (same height class as an Observation row): bridge-kind icon + sentence `Text` + term `PVComboBox` + trailing `PVButton`s (pending) or Save / Revert `PVIconButton`s (edited). One call site → stays a snowflake. |
| Observation dialog (name / date) | Snowflake | Ship | `Features/CitationComposer/CitationComposerObservationDialogForm.swift` | Dialog confirm now updates the row draft only (row becomes Edited); the row's Save commits. |
| New-subject dialog | Snowflake | **New** (feature-private) | inside `CitationComposerView` via `pvFormDialog` | Label + Description `PVField`s. Do not add a kit control. |
| Graph create / edit dialog | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphView.swift` (existing sheet) | Primaries: working-label caption. Bridges: Description only. |
| Graph disambiguation sheet | Snowflake | **Remove** | `Features/EvidenceGraph/EvidenceConnectDisambiguationForm.swift` | Role / relationship type moves into the composer's connection row. |
| Bridge card title | Snowflake | Ship | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Renders `EvidenceBridgeEdgeSummary.sentence(for:in:)`, which applies the §2.4 fallback chain. |
| IconButton | Component | Ship | `DesignSystem/Components/IconButton/` | Row Save / Revert. |
| Button | Component | Ship | `DesignSystem/Components/Button/` | Save reading, Save connection, Discard connection, Done. |
| ComboBox | Component | Ship | `DesignSystem/Components/ComboBox/PVComboBox.swift` | Row subject (with New… options), Property, connection role term. |
| Select | Component | Ship | `DesignSystem/Components/Select/PVSelect.swift` | Artifact / Citation identity (disabled state + hint). |
| Confirm | Component | Ship | `.pvConfirm(item:)` | Delete observation; unsaved-work guard. |
| FormDialog | Component | Ship | `.pvFormDialog` | New subject; bridge description edit. |
| Callout / Field error | Component | Ship | `PVCallout`, `PVField(error:)` | Row errors; "Will be saved to CIT-…" note. |
| Card | Component | Ship | `PVCard` | Observation rows and the connection row. |
| Badge / Chip | Component | Ship | `PVBadge` / `PVChip` | Row state marker if the board wants a token (e.g. "Unsaved"). |
| ContextMenu | Component | Ship | `pvContextMenu` | Row ⋯: Delete observation…, Negate / Affirm. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit `PVRowCommit` / `PVInlineSaveRow` control | One call site. Compose IconButtons on the existing row. |
| A second "connect mode" layout of the composer | The connection is one row in the same stack. |
| Endpoint pickers or one row per edge | A connection is one unit. Wrong endpoint = discard / delete and reconnect. |
| A modal wizard for Connect | The graph sheet is what this board removes. |
| An autosave indicator / sync spinner | Commits are explicit. |
| A navigation-blocking banner | The guard is a confirm at the moment of leaving. |

---

## 8. Out of scope

- Subject / bridge / Citation delete and counted cascades (**S8-D6** / **S8-09**)
- Auto Transcribe (**S8-D1**), PDF Find / paste (**S8-D2**)
- Graph card badges and bridge sentence wording (**S8-D3**)
- Undo, autosave, app-quit / window-close guards
- Creating bridges or `source` subjects from the composer

---

## 9. Handoff

1. Agree the frames. Record any board-level copy decisions in this brief before archiving.
2. Archive this brief under `archive/` when the board is agreed.
3. Record in [`../completed.md`](../completed.md).
4. Implement **S8-11** against the board and the S8-11 plan in [`../deployment-plan.md`](../deployment-plan.md#s8-11--pr-composer-and-connect-simplification--interpretation-write-integrity).
5. **S8-D1** and **S8-D2** design against this board's reading section and row chrome, not the S8-10 frames.
