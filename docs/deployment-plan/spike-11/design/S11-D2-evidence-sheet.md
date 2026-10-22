# S11-D2 — Promote evidence sheet: why it was suggested, and the evidence to pin

**Kind:** Claude Design board (**rethink** of the evidence sheet frames on the Promote flow board, S9-D16; the rows' frames from S11-D1 stay)
**Spike:** Provenencia Spike 11 (Promote matching layers)
**View:** Promote page — the evidence sheet a row's assessment opens
**Implements later as:** PR **S11-20** (the sheet). The lines it draws come from S11-19, S11-25 and S11-32
**Depends on:** S11-25 (the evidence RPC: every member's line, pin ids, relation and cause codes); S11-32 (the lifespan and likeness lines) before S11-20 ships
**Related:** [`../claim-evidence.md`](../claim-evidence.md) (the requirements and draft wording), [`S11-D1-promote-rows.md`](S11-D1-promote-rows.md) (the rows, the band and Contested badges, the move captions), [`../../../promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3, Spike 10's likeness gallery ([decisions](../../archive/spike-10/README.md))
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share) (the board S11-D1 extended); not yet drawn
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

### This brief is a rethink

Replace the evidence sheet frames on the Promote flow board. Drop the six-column comparison table (Pin, Compared, This source, target, Result, Weight), the "This subject" card and the "Through <name>" cards. Keep the sheet's shell: the panel at 880 pt, its title, the claim fields (Status, Confidence, Argument) and the "Back to all rows" footer. Do not redraw the Promote rows, the target menu or the page header: S11-D1 owns them.

---

## 1. Objective

The sheet answers two different questions and keeps them apart.

1. **"Why was this suggested?"** A short summary at the top: the band, Contested when it applies, and the reason the row already shows, in one sentence. It explains matching; nothing in it is pinned.
2. **"What's the evidence that this record is <entity>?"** Every piece of evidence for the claim, as sentences the researcher checks or unchecks. Checked evidence is pinned on the claim when Done files the page.

**What changes.**

- Every member of the target record shows. Joining a record with five named members shows five name matches; identical values collapse into one line ("The names match · 5 records") that expands to each record.
- Lines read as sentences built from how the values relate ("Same name, different spelling", "Same date, recorded less precisely"), never from a number. The Weight column goes.
- Relationships say what matched: "They share a matching connection — John Smith · subject of · Birth of John Smith" is its own checkable line, and the birth's date lines sit under it.
- A person's lifespan shows as a line that is never pinned, including when one side's own dates contradict each other.
- Likenesses show as two crops side by side.

**What stays.** The panel, its width and title; the claim fields and their behavior; the footer; the conflict callout (reworded); pins as a set the researcher edits, starting from what Go checks; the rows behind the sheet.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A row proposes an Identity Claim | "This Source Subject is this record." The sheet is that claim's evidence. It opens from the row's assessment button. A row set to New, Skip or unset has no record and no evidence. |
| Band and Contested | Strong, Medium, Weak, None, with today's Badge tones; Contested is a danger Badge beside the band (S11-D1 CM-1). Both describe matching, not the pins. |
| The row's reason | Today's caption, already worded: "via John Smith → PER-7QH3D (son)", "Agrees on Name", "Too little agrees to suggest a match", or a move: "after John Smith moved to PER-7QH3D". |
| Members | The record's accepted members: Subjects on other Sources already filed on it. Each line pairs this Source's value with one member's value and names that member's Source ("Census 1851"). |
| A line | One value on this Source against one member's value of the same property. Its outcome is agrees, similar, conflicts or unknown, and it carries how the two values relate (same, equivalent, contains, overlaps, near, different) and why (spelling, nickname, precision, a missing part, a gap). |
| What starts checked | Same and equivalent lines start checked. Contains, overlaps and near are pinnable and start unchecked. A conflict is shown and cannot be pinned. Go decides; the sheet shows it. |
| Collapsed lines | Members whose values are identical share one line. Checking it pins every member; expanding it lets the researcher uncheck one. A line with some members unchecked is partly checked. |
| Properties pair up | Sex at birth and gender are one section; an event's date with its start and end dates is one section. |
| Negative values | This Source can say "not male". Against a member who is male it is a conflict. |
| Relationships | A link this Subject has (a birth it is the subject of, a spouse, a parent, a place an event happened at) that matches a link the record's members have, to the record the neighbor row is placed on. The relationship line pins the plumbing behind it (the link itself, its role or type, the event's type). The neighbor's other properties are lines under it, with the same rules as the Subject's own lines. |
| Order | Go orders everything. Sections: the identity property (Name on a person), then Likeness, then the rest. Relationships: birth and death first, then the rest. The lifespan comes last among the Subject's own Cards. |
| Neighbors | Each relationship names the neighbor row by its name and the ref it shows on the page: the record ref when placed (PER-7QH3D), else its Source ref. |
| Lifespan (people only) | Worked out from every dated event the person takes part in, on each side, as a range of years. It is a check, not evidence: never pinned. It fits, it conflicts (the ranges cannot both be true), or one side has none. A side whose own dates contradict each other has no lifespan, and the sheet says which side. |
| Likeness | A face crop drawn on a Citation's page (Spike 10). The app cannot judge whether two different photos show the same face. The same image file on both sides is the same likeness and starts checked; two different images are pinnable and start unchecked, for the researcher to judge by eye. A likeness never counts toward the band. |
| Loading | The evidence is read when the sheet opens, one row at a time. It takes well under a second; a large record takes longer. |
| Pins and Done | Checked lines and relationships are pinned on Done. A row whose sheet was never opened pins exactly what starts checked. |
| Argument | Empty until the researcher types. Nothing drafts it. |

### 2.1 What this board is not

- The Promote rows, their badges and captions (S11-D1).
- Opening the sheet from a connection subrow (not planned).
- A "select all" or "pin nothing" control (not planned: matches start checked).
- Points, weights or similarity numbers anywhere on the sheet.
- Editing a value, a member or a link from the sheet. The way to fix a record is its own page.
- A likeness comparison by the app (face matching is not planned).
- Status choices beyond Accepted (a later spike, as today's hint says).

---

## 3. Implementation gate

| Ships in **S11-20** | Stays out |
| --- | --- |
| ES-1 – ES-22: the whole sheet below the title, its copy, states and VoiceOver. | The lines' data (S11-19, S11-25, S11-32). Rows (S11-D1). Opening from a subrow, face matching, select-all (not planned). |

---

## 4. Requirements

### 4.1 Shell and summary

| ID | Requirement |
| --- | --- |
| ES-1 | The panel keeps its title ("Mary Smith → PER-7QH3D Mary Smith") and width. Its subtitle is the band and the reason as the row speaks them: "Strong, contested · via John Smith → PER-7QH3D (son)". The old second sentence ("Agreeing comparisons are pinned…") goes. A row with no record has no subtitle. |
| ES-2 | **Why this was suggested.** The first block of the body: a sunken Card titled "Why this was suggested" holding the band Badge, the Contested Badge when contested, and the reason sentence on one line; a move caption ("Updated · after John Smith moved to PER-7QH3D") under it when the row shows one. It also shows the row's Updated caption whenever the row shows one ("Updated · after John Smith moved to PER-7QH3D", or "Updated · via …"), worded as on the row. When contested it adds one line: "Something below disagrees with this record." Nothing in this Card has a checkbox. |
| ES-3 | **The evidence heading.** Under the summary, a SectionHeader: "What's the evidence that this is PER-7QH3D Mary Smith?" (the record's ref and title), with meta "N checked" counting checked pins. |
| ES-4 | The conflict Callout (warning, compact) sits under the heading when any line conflicts: title "Some evidence disagrees", message "Disagreeing lines can't be pinned. The claim can still be filed." |

### 4.2 Property sections and lines

| ID | Requirement |
| --- | --- |
| ES-5 | The Subject's own evidence comes first, one Card per property, in the order Go gives (§2 Order). The Card's header is the property's name ("Name", "Sex at birth", "Date"). A paired section uses the lead's name ("Sex at birth" holds the gender lines). |
| ES-6 | **A line** is one row: a checkbox; a sentence (ES-8) followed by "· N records" when collapsed; the outcome Badge (subtle) at the right: Agrees (success), Similar (info), Conflicts (danger), Not compared (neutral). Under the sentence, muted: "This source: John Smith" and, for one member, "Census 1851: J. Smith". |
| ES-7 | A conflict line shows the checkbox disabled and unchecked; its sentence is in primary text, not dimmed. A negative value on this Source reads "This source: not male". |
| ES-8 | **Sentences**, from relation and cause. The copy table: |

| Relation · cause | Sentence |
| --- | --- |
| same (any cause) | "The same" — Name section: "Same name"; Date: "Same date"; every other section: "The same" |
| equivalent · fold | "Same, written with different accents" |
| equivalent · alias | "Same name (nickname)" |
| equivalent · abbreviation | "Same name (abbreviated)" |
| equivalent · other | "The same, written differently" |
| contains · precision | "Same date, recorded less precisely" |
| contains · missing_part | "Matches; part of the name is missing" |
| contains · initial | "Matches by initial" |
| contains · other | "Matches, with less detail" |
| overlaps | "The dates overlap" (Date) / "Partly the same" (other) |
| near · spelling | "Similar spelling" |
| near · gap | "Close dates" |
| near · other | "Similar" |
| different, or a negative that contradicts | "Disagrees" — Date: "Different dates"; Name: "Different names" |
| unknown (likeness only, ES-15) | "Compare these likenesses" |

| ID | Requirement |
| --- | --- |
| ES-9 | **Collapsed lines.** A line with more than one member reads "Same name · 5 records" and has a ghost "Show 5 records" disclosure Button at its right, before the Badge. Expanded, it lists each member indented under the line, one per row: its own checkbox, the member's value, its Source title. The disclosure becomes "Hide records". |
| ES-10 | A collapsed line's checkbox is checked when every member is checked, unchecked when none is, and partly checked (the indeterminate dash) otherwise. Pressing a partly checked box checks every member. |
| ES-11 | A Card's lines are separated by a thin divider. A Card with 20 lines grows; the sheet scrolls. A member list of 50 records grows the same way. |

### 4.3 Relationships

| ID | Requirement |
| --- | --- |
| ES-12 | After the Subject's own Cards, one Card per relationship. Its header row is a line: a checkbox, the relationship sentence, the neighbor's ref in mono caption, and the outcome Badge. The neighbor's lines follow in the Card, laid out as ES-6, indented one step. |
| ES-13 | **Relationship sentences.** Every row kind and every link kind reads the same way: a sentence from the outcome, then the link on this Source as the page already names it (S11-D1's "·" form, the same words a connection subrow shows). Agrees: "They share a matching connection" — "John Smith · subject of · Birth of John Smith". Similar: "They share a similar connection" (unchecked). Conflicts: "A connection disagrees" (box disabled) — "Mary Smith · spouse of · Robert Gates". A person row, an event row (its person neighbor: "Birth of John Smith · subject · John Smith") and a place row ("Toronto · part of · Ontario") all use it. When the page has no name for the link, the second line names the neighbor: "with Robert Gates". |
| ES-14 | A relationship with no lines under it shows its header only, plus a muted caption: "Nothing else to compare." Checking it still pins the link. |

### 4.4 Lifespan and likeness

| ID | Requirement |
| --- | --- |
| ES-15 | **Likeness.** A "Likeness" Card on a person, after Name (§2 Order). Each line shows two crops side by side, 96 pt square, this Source's on the left and the member's on the right, each captioned with its Source title. Same image: "Same likeness", Agrees, checked. Different images: "Compare these likenesses", outcome Badge "Judge by eye" (neutral), pinnable, unchecked. A crop that cannot be shown draws the person mark in the Thumbnail. Likeness lines never collapse unless the image is the same. |
| ES-16 | **Lifespan.** A "Lifespan" Card on a person, last among the own Cards. One line, no checkbox, with a muted caption under it: "Worked out from dated events. A guide for matching, never pinned." Each side's range is written as the Person page writes a life span ("1817–1880", "about 1815–1890", "–1840"). Fits: "The lifespans fit: 1817–1880 here, about 1815–1890 on the record", Badge Agrees. Conflict: "The lifespans don't fit: 1850–1890 here, 1790–1840 on the record", Badge Conflicts. One side has none: "No lifespan to compare: the record has no dated events", Badge Not compared. Contradiction: "No lifespan for this source: its own dates contradict each other" (or "for the record"), Badge Not compared. A lifespan conflict counts as a disagreement for ES-2 and ES-4. |

### 4.5 States

| ID | Requirement |
| --- | --- |
| ES-17 | **Loading.** The summary and heading show at once; the Cards area shows three skeleton bars until the evidence arrives. |
| ES-18 | **Failed.** A danger Callout (compact) in the Cards area: "The evidence could not be loaded" with the error. The claim fields stay usable. |
| ES-19 | **No record.** For a row set to New, Skip or unset: the summary Card is replaced by an EmptyState: "Pick a record to see its evidence" — "Choose a record in this row's menu. New and Skip file no evidence." The claim fields still show for New. |
| ES-20 | **Nothing to compare.** A record with no member values to compare and no relationships: EmptyState in the Cards area: "Nothing to compare yet" — "The record's other sources say nothing about these properties." |

### 4.6 Claim fields and accessibility

| ID | Requirement |
| --- | --- |
| ES-21 | The Claim block stays below the evidence. The Argument hint becomes "N pieces of evidence checked · say why these are the same" (N counts checked pins). |
| ES-22 | VoiceOver: a line reads as one element, its sentence followed by the checkbox's own state, which the kit speaks (checked, not checked, or partly checked): "Name. Same name, 5 records. This source: John Smith." A member: "Census 1851: John Smith." A relationship: "They share a matching connection. John Smith · subject of · Birth of John Smith." A likeness line names both Sources: "Likeness from this source and Census 1851. Compare these likenesses." The lifespan line has no checkbox and reads its sentence. The disclosure reads "Show 5 records, collapsed". |

---

## 5. Suggested frames

1. **Strong row, five members.** Summary (Strong, "via John Smith → PER-7QH3D (son)"); heading with "7 checked"; Name Card with "Same name · 5 records" collapsed (ES-6, ES-9); Sex at birth Card with "The same"; Lifespan Card that fits (ES-16).
2. **Frame 1 expanded.** The name line expanded, one member unchecked, the line's box partly checked (ES-10).
3. **Contested Medium row.** Summary with Contested and the "Something below disagrees" line (ES-2); the conflict Callout (ES-4); a "Different dates" conflict line with the disabled box (ES-7); a checked "Same name (nickname)" line (equivalent); an unchecked "Close dates" line (near).
4. **Relationships.** "They share a matching connection — John Smith · subject of · Birth of John Smith" with a "Same date, recorded less precisely" line under it, unchecked (ES-12, ES-13); "They share a matching connection — John Smith · spouse of · Mary Ames" with name and sex lines; a conflicting connection with its box disabled; an event's place connection with no lines (ES-14).
5. **Likeness.** One "Same likeness" line and two "Compare these likenesses" lines, crops side by side (ES-15).
6. **Lifespan edge cases.** A conflict line, and a "No lifespan for this source: its own dates contradict each other" line (ES-16).
7. **States.** Loading (ES-17), failed (ES-18), a New row (ES-19), nothing to compare (ES-20).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sheet shell | Component | Ship | `DesignSystem/Components/Panel/PVPanel.swift` | Title, subtitle, 880 pt, sunken footer; as today. |
| Evidence sheet body | Snowflake | **Rethink** | `Features/Promote/PromoteView.swift` `PromoteEvidenceSheet` | Columns, `ExhibitGroup` and the "Through" cards go; summary, heading, Cards and lines come in. One call site. |
| Why-suggested summary | Component | Ship | `DesignSystem/Components/Card/PVCard.swift` (`tone: .sunken`) + `DesignSystem/Components/Badge/PVBadge.swift` | The same band and Contested Badges the row uses (S11-D1), drawn by one shared Promote view the row and the summary both instance. Not a Callout: it is not a warning. |
| Evidence heading | Component | **Extend** | `DesignSystem/Components/SectionHeader/PVSectionHeader.swift` (`meta:` for "N checked") | Its title takes `PVCopy`, as Button, Callout and DisclosureButton do, because the title names the record. No local heading style. |
| Property and relationship Cards | Component | Ship | `DesignSystem/Components/Card/PVCard.swift` | One per section or relationship; header on `surfaceSunken`, as today's group cards. |
| Evidence line | Snowflake | **Rethink** | `Features/Promote/PromoteView.swift` | Composes Checkbox, text, disclosure Button, Badge. One call site: not a kit primitive. Replaces `comparisonLine`. |
| Line and member checkbox | Component | **Extend** | `DesignSystem/Components/Checkbox/PVCheckbox.swift` (`isIndeterminate:` for a partly checked line, `isDisabled:` for a conflict) | Gains a spoken "partly checked" state for `isIndeterminate`, so every caller speaks it. No local check control. |
| Show / Hide records | Component | Ship | `DesignSystem/Components/Disclosure/PVDisclosureButton.swift` | Not a chevron on the whole row; not a link. |
| Outcome mark | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` (`subtle: true`) | Agrees / Similar / Conflicts / Judge by eye / Not compared. |
| Likeness crop | Snowflake | **Extend** | `Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailLikenessSection` tile (a 96 pt `Platform/CachedThumbnail.swift` crop over a muted Source caption) | Its crop-and-caption becomes one shared view the gallery and the sheet both use (second call site). No local image loader. |
| Conflict and failure notices | Component | Ship | `DesignSystem/Components/Callout/PVCallout.swift` (`compact: true`) | Warning for disagreement, danger for a failed read. |
| Loading | Component | Ship | `DesignSystem/Components/SkeletonBar/PVSkeletonBar.swift` | Three bars, as S11-25 drew. |
| No record, nothing to compare | Component | Ship | `DesignSystem/Components/EmptyState/PVEmptyState.swift` | Pane-level empty inside the sheet. |
| Claim fields | Component | Ship | `DesignSystem/Components/Field/PVField.swift`, `DesignSystem/Components/Select/PVSelect.swift`, `DesignSystem/Components/TextArea/PVTextArea.swift` | Unchanged; the Argument hint is new copy. |
| Column headers, Weight column | Snowflake | **Remove** | `PromoteEvidenceSheet` `sheetColumns`, `weightText` | Go with the table. |

**Non-goals.** No new kit primitive (the two kit rows are Extend). No local Badge, Checkbox, Card, Callout or image view. No table component for lines. No numbers. No select-all. No hover actions on a line.

---

## 7. Handoff

When the board is agreed: record the board link and frame numbers on S11-20's **Board** line and in a §9 "Board notes" table here (as S11-D1 did in its §8), and move S11-20 from `Blocked — …` to `Planned` with `plan-pr` once S11-25 and S11-32 have landed. Move this brief to the spike's design archive folder when S11-20 lands, and update [`README.md`](README.md).

---

## 8. Review log

[`../prs/S11-20-claim-evidence-sheet.review.md`](../prs/S11-20-claim-evidence-sheet.review.md) (round 1, 2026-10-22) covers this brief: inventory rows for the section header, checkbox, likeness tile and readout badges, Go-owned order, and connection wording were changed from it.
