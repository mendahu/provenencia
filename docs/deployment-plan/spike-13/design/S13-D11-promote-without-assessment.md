# S13-D11 — Promote page without the evidence sheet and the assessment column, with a status per row

**Kind:** Claude Design board (**enhancement** of the Promote page board, S9-D16 / S11-D1 / S11-D3: frames are removed and the row is redrawn in place)
**Spike:** Provenencia Spike 13 (Person page), slice 0 (removing Observation pins)
**View:** Promote page — the rows, the column headings and the evidence sheet
**Implements later as:** PR **S13-23** (Promote stops asking for pins: PA-1 – PA-8) and PR **S13-30** (Provisional claims at Promote: PS-1 – PS-6)
**Depends on:** — (the decision is [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §5, docs PR #623)
**Related:** [`../../../promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3 and §3.1 (the page as it is after this board); working rules "The score ranks, unseen" and "Promote's removal has a brief" in [`../deployment-plan.md`](../deployment-plan.md#slice-0--removing-observation-pins); [`../prs/S13-23-promote-stops-asking-for-pins.md`](../prs/S13-23-promote-stops-asking-for-pins.md)
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share) (the board S11-D1 and S11-D3 extended); not agreed yet
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

### This brief is an enhancement that removes

Work on the Promote page frames that exist: the header, the kind sections, the rows (kind tile, name and ref, arrow, target menu, assessment column, Updated / Decided badges), the Updated caption, the conflict and duplicate callouts, the connection subrows, the record search in the target cell (S11-D3), the Already filed block, the footer, and the evidence sheet frames (S11-D2). **Delete** the evidence sheet frames and every frame or state that only exists to show the sheet. **Redraw** the row and the column headings without the assessment column. Do not redraw the header, the target menu and its items, the record search, the subrows, the callouts or the footer.

---

## 1. Objective

Promote becomes a quick filing action: the researcher picks a target per row and presses Done. Nothing on the page asks them to judge evidence.

**What changes.**

1. The **evidence sheet** is gone. Nothing on a row opens a sheet: no pins, no Status, Confidence or argument fields.
2. The **assessment column** is gone from every row: no band Badge (Strong, Medium, Weak, No match, Little evidence), no **Contested** Badge, no reason text ("Agrees on birth date", "Too little agrees…", "Another row took the best match"), no chevron. The column heading **Assessment** goes.
3. The row, the column headings and the Already filed rows are laid out again without that column (PA-2, PA-3, PA-4: **this board decides how**).
4. The **Updated caption** under a moved row says why it moved only when the proposal names a settling move or a via neighbor. Otherwise the row shows the Updated badge alone.
5. Each filed row gets a **status** control, **Accepted** by default or **Provisional**: the row's only claim field (PS-1 – PS-6). It ships later than the removal, in S13-30, but is drawn on this board so the row is designed once.

**What stays.** The header; the kind sections and their counts; Map the rest; the kind tile, name and Source ref; the target menu, New, Skip and Find another record…, with today's widths and items; the Updated and Decided badges; the conflict and duplicate callouts; the connection subrows; the Already filed block's content; the footer's summary, Cancel and Done; the leave confirm.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The score still ranks | Promote still chooses which record a row starts on and orders each menu's records by its score. The page shows none of the score: no band, no Contested, no reason (working rule "The score ranks, unseen"). |
| A row starts on its best record, or unset | Strong and Medium rows start on their record; a Weak one starts on its best record for review; a row with no suggestion, or whose best record another row took, starts unset with the red "Select where this files" placeholder, as today. Nothing on the row says which of these happened. |
| Done files claims | Each filed row becomes an Identity Claim with its status and nothing else: claims carry no confidence and no argument (decided 2026-10-29). Until S13-30 ships, every row files Accepted. |
| Updated compares proposals | After a decision, the page re-proposes; a suggested row whose target changed shows the Updated badge (info tone, refresh icon). |
| Why it moved | The page knows a neighbor behind the change in two cases. **Settling moved it:** "Updated · after Gracie Smith moved to PER-7QH3D", "… after Gracie Smith → PER-7QH3D was matched", "… after Gracie Smith was unmatched". **It was reached from a neighbor:** "Updated · via Gracie Smith → PER-7QH3D (spouse)", or without the term when the connection has none. Any other change has no reason the page can say. |
| A neighbor not on the page | The caption reads "Updated · Reached from a matched neighbor" or "Updated · after a connected row changed", as today. |
| Already filed rows | Read-only: kind mark, name and ref, arrow, the record's title and ref, and the italic "Filed from this Source earlier". They sit in the same column grid as the working rows today. |
| Decided | A row the researcher changed shows the Decided badge (neutral, subtle, check). A decided row never shows the Updated caption. |
| Status | A filed row's Identity Claim is **Accepted** (default) or **Provisional**. Rejected is not offered here. Accepted files the Subject as a member of the record, and its connections file when both ends are members. Provisional files a provisional claim and nothing else: the Subject still counts as not promoted, its connections wait, and the record's facts do not change. |
| Provisional next time | When Promote opens again on that graph, a provisional row comes back on its provisional record with status Provisional. Setting it to Accepted and pressing Done files it and its waiting connections. Skip leaves the provisional claim as it is. |
| A provisional end | A connection with a provisional end does not file. The subrow shows it in the existing "waits for a named end" state, naming that end. |
| Widths | The page content is at most 1200 pt wide. Today's columns: kind tile 28 pt, Source name (flexible), arrow 14 pt, target 330 pt, assessment (flexible), badges 112 pt, with the same gap between each. |
| Headings | Today the columns are headed **On this Source**, **Files on** and **Assessment**. |
| VoiceOver today | The target menu reads "Files Mary Smith on" and the chosen option ("PER-4KX2M  Smith, Mary"); the assessment button reads "Evidence for Mary Smith: Strong, contested, via …". |

### 2.1 What this board is not

- Confidence or argument. Identity Claims no longer carry either (decided 2026-10-29).
- A Rejected status, or any way to see or edit a claim after Done other than reopening Promote ([identity-claim-review idea](../../../ideas/identity-claim-review.md)).
- Any other signal of match quality on the row or in the target menu (a dot, a percentage, a tint). The score ranks unseen.
- Changes to the target menu items, the record search, the connection subrows or the footer.
- The Person page's comparison of a person's records (S13-D10).

---

## 3. Implementation gate

| Ships in **S13-23** | Ships in **S13-30** | Stays out |
| --- | --- | --- |
| PA-1 – PA-8: the sheet and the assessment column removed; the row, headings and Already filed rows laid out without it; the Updated caption's rule; VoiceOver. Every row files Accepted, with no status control yet. | PS-1 – PS-6: the status control, its default, a provisional row coming back preselected, its connections waiting. | What a row starts on moving into Go (S13-24, no visible change). Rejected at Promote. Any redesign of the target menu or subrows (S14-22 changes the rows' Promote copy after this board). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PA-1 | No frame shows the evidence sheet. Clicking anywhere on a row except the target menu and its controls does nothing. |
| PA-2 | A working row shows, left to right: kind tile, name over Source ref, arrow, target menu, and the Updated / Decided badges. No band, Contested, reason or chevron. **The board decides** what the freed width becomes. Draw both options in frame 1 and pick one: **(a)** the column's space stays empty, so the target menu and badges stay exactly where they are today; **(b)** the column closes and the Source name column takes the freed width; the target menu (330 pt) and the badges (112 pt) move right. Whichever is picked, the target menu is the same width on every row and lines up down the page. |
| PA-3 | The column headings keep today's copy, **On this Source** and **Files on**, and line up with their columns. The **Assessment** heading goes. |
| PA-4 | An Already filed row lines up with the working rows above it under the option PA-2 picks: its record title under the target menu, and "Filed from this Source earlier" placed where the board puts it (today it sits in the assessment column). |
| PA-5 | A suggested row whose target changed shows the Updated badge and, under the row at the note indent, the caption in today's style (italic caption, info ink): the move sentence when settling moved it, else the via sentence when it was reached from a neighbor. When that neighbor is not a row on the page, the caption reads "Updated · after a connected row changed" (move) or "Updated · Reached from a matched neighbor" (via), as today. |
| PA-6 | A row whose target changed for no reason the page can say shows the Updated badge and **no** caption line. The row's height is the same as a row with no caption. |
| PA-7 | A decided row shows Decided and no Updated badge or caption, as today. |
| PA-8 | VoiceOver: the target menu reads as today ("Files Mary Smith on" and the chosen option), and the badges and caption read as today. Nothing reads "Evidence for …", a band or "contested": the assessment button is gone. |

### 4.1 Status (S13-30)

| ID | Requirement |
| --- | --- |
| PS-1 | A working row whose target is a record shows a **status** control with two choices, **Accepted** and **Provisional**; it starts on Accepted. It is a Select at the small size, placed where the board puts it (the freed column of PA-2 option (a) is one candidate). It lines up down the page like the target menu. |
| PS-2 | A row set to Skip, or still unset, shows no status control and keeps its row height. **Whether a New row can be Provisional is open** (S13-30's plan decides; recommendation: no). Draw New both ways: with the control, and without it. |
| PS-3 | Choosing Provisional marks the row Decided, like any change. Every connection subrow with that row as an end shows the waits state, naming that row: "Waits for Mary Smith". Switching back to Accepted restores the subrows' states. |
| PS-4 | A row that comes back provisional (filed Provisional on an earlier Done) opens on its record with status **Provisional** and is not Decided. Nothing else on the row marks it; the status control says it. |
| PS-5 | The footer summary counts a Provisional row as filed ("File 12"). Whether it says how many are provisional is the board's call; if it does, show the count as one more part of the summary, in the same style. |
| PS-6 | VoiceOver reads the control as "Status for Mary Smith, Accepted" (or Provisional), and a waiting subrow as today. |

---

## 5. Suggested frames

1. **Rows, both PA-2 options side by side** (the board then keeps the one picked and deletes the other): a People section with a Strong row on a record, a Weak row drafted onto a record, an unset row with the red placeholder, and a decided row; an Already filed block of two rows under it (PA-2, PA-3, PA-4, PA-7).
2. **After deciding Gracie** on the picked layout: one row with "Updated · after Gracie Smith moved to PER-7QH3D", one with "Updated · via Gracie Smith → PER-7QH3D (spouse)", one with "Updated · after a connected row changed" (its neighbor is not on the page), and one with the Updated badge and no caption (PA-5, PA-6).
3. **A busy row** on the picked layout: a Birth event with its connection subrows and a duplicate callout, so the indent of notes and subrows is checked against the new grid.
4. **Deleted**: the evidence sheet frames and any row state that pointed at them (PA-1). List their frame numbers in the board notes.
5. **Status** on the picked layout: a People section with an Accepted row, a Provisional row with two waiting subrows, a Skip row with no control, and a New row drawn with and without the control (PS-1 – PS-3).
6. **Coming back**: the same graph reopened, the provisional row on its record with status Provisional, and the footer summary (PS-4, PS-5).

---

## 6. UI building-block inventory

This table is **binding**. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Promote row | Snowflake | **Extend** | `Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | Loses the assessment cell; laid out per PA-2. No new control on the row. |
| Column headings | Snowflake | **Extend** | `Features/Promote/PromoteView.swift` `PromoteView.columnHeaders` | On this Source and Files on, as today; the Assessment heading goes (PA-3). |
| Already filed row | Snowflake | **Extend** | `Features/Promote/PromoteView.swift` `PromoteAnchorRow` | Re-aligned per PA-4. |
| Column grid | Snowflake | **Extend** | `Features/Promote/PromoteView.swift` `PromoteLayout`, and one private grid row view shared by the headings, the working row and the Already filed row | Today the headings, working row and Already filed row each spell the grid by hand. After this board they fill the slots of one grid row, and whatever PA-2 picks is drawn once there with the widths in `PromoteLayout`. |
| Band and Contested marks, reason line, chevron | Snowflake | **Remove** | `Features/Promote/PromoteReadout.swift` `PromoteReadoutBadges`, the row's evidence button | Gone; `PromoteReadout.swift` is deleted. `PVBadge` stays in the kit. |
| Updated caption | Snowflake | Ship | `Features/Promote/PromoteReadout.swift` `PromoteUpdatedCaption`, which becomes private in `PromoteView.swift` | Same type, ink and indent; shown per PA-5, PA-6. |
| Updated, Decided badges | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` | As today. |
| Status control | Component | Ship | `DesignSystem/Components/Select/PVSelect.swift` (size `.sm`) | Two options. Not a Radio, Switch or Chip; no new control. Lives in the shared grid row. |
| Target menu, record search | Component | Ship | `DesignSystem/Components/Select/PVSelect.swift`, `DesignSystem/Components/ComboBox/PVComboBox.swift` | Unchanged. |
| Conflict and duplicate notes | Component | Ship | `DesignSystem/Components/Callout/PVCallout.swift` (`compact: true`) | Unchanged. |
| Connection subrows | Snowflake | Ship | `Features/Promote/PromoteView.swift` `PromoteConnectionSubrows` | Unchanged. |
| Evidence sheet | Snowflake | **Remove** | `Features/Promote/PromoteEvidenceSheet.swift`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift` | Gone with every frame that drew it. |

**Non-goals.** No new kit primitive. No replacement signal of match quality. No local Badge, Select or Callout.

---

## 7. Handoff

When the board is agreed: record the board link and the picked PA-2 option on this brief and in the S13-23 plan (**Board** line, Design → macOS, and the Decision that today reads the slot), move S13-23 from `Blocked — on board S13-D11` to `Planned` with `plan-pr`, and when S13-23 lands move this brief to the spike design archive folder and update [`README.md`](README.md).
