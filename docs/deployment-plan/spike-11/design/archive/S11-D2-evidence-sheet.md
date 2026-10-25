# S11-D2 — Promote evidence sheet: why it was suggested, and the evidence to pin

**Kind:** Claude Design board (**rethink** of the evidence sheet frames on the Promote flow board, S9-D16; the rows' frames from S11-D1 stay)
**Spike:** Provenencia Spike 11 (Promote matching layers)
**View:** Promote page — the evidence sheet a row's assessment opens
**Implements later as:** PR **S11-20** (the sheet). The lines it draws come from S11-19, S11-25 and S11-32
**Depends on:** S11-25 (the evidence RPC: every member's line, pin ids, relation and cause codes); S11-32 (the lifespan and likeness lines) before S11-20 ships
**Related:** [`../claim-evidence.md`](../../claim-evidence.md) (the requirements and draft wording), [`S11-D1-promote-rows.md`](S11-D1-promote-rows.md) (the rows, the band and Contested badges, the move captions), [`../../../promote-graph-alignment.md`](../../../../promote-graph-alignment.md) §3, Spike 10's likeness gallery ([decisions](../../../archive/spike-10/README.md))
**Board:** [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share) (the board S11-D1 extended), agreed 2026-10-22. Frames E01 – E14 gate S11-20; where the board differs from this brief, §9 records it and the board wins
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
| Lifespan (people only) | Worked out from every dated event the person takes part in, on each side, from the earliest to the latest. It is a check, not evidence: never pinned. It fits, it conflicts (the ranges cannot both be true), or one side has none. A side whose own dates contradict each other has no lifespan, and the sheet says which side. |
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
| ES-1 | The panel keeps its title ("John Smith → PER-7QH3D John Smith") and width. Its subtitle is the band and the reason as the row speaks them: "Medium, contested · via John Smith → PER-7QH3D (subject)". The old second sentence ("Agreeing comparisons are pinned…") goes. A row with no record has no subtitle. |
| ES-2 | **Why this was suggested.** The first block of the body: a sunken Card titled "Why this was suggested" holding the band Badge, the Contested Badge when contested, and the reason sentence on one line. It also shows the row's Updated caption whenever the row shows one ("Updated · after Birth of John Smith → EVT-4RJ9T was matched", or "Updated · via …"), worded as on the row. When contested it adds one line: "Something below disagrees with this record." Nothing in this Card has a checkbox. |
| ES-3 | **The evidence heading.** Under the summary, a SectionHeader: "What's the evidence that this is PER-7QH3D John Smith?" (the record's ref and title), with meta "N checked" counting checked pins. |
| ES-4 | The conflict Callout (warning, compact) sits under the heading when any line conflicts: title "Some evidence disagrees", message "Disagreeing lines can't be pinned. The claim can still be filed." |

### 4.2 Property sections and lines

| ID | Requirement |
| --- | --- |
| ES-5 | The Subject's own evidence comes first, one Card per property, in the order Go gives (§2 Order). The Card's header is the property's name ("Name", "Sex at birth", "Date"). A paired section uses the lead's name ("Sex at birth" holds the gender lines). |
| ES-6 | **A line** is one row: a checkbox; a sentence (ES-8) followed by "· N records" when collapsed; the outcome Badge (subtle) at the right: Agrees (success), Similar (info), Conflicts (danger), Judge by eye and Not compared (neutral). Under the sentence, muted: "This source: John Smith" and, for one member, "1871 England Census — Leeds, Briggate: J. Smith". |
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
| ES-13 | **Relationship sentences.** Every row kind and every link kind reads the same way: a sentence from the outcome, then the link on this Source as the page already names it (S11-D1's "·" form, the same words a connection subrow shows). Agrees: "They share a matching connection" — "John Smith · subject of · Birth of John Smith". Similar (the neighbor's own facts disagree): "They share a similar connection", unchecked. There is no disagreeing relationship (§5.1). A person row, an event row (the same words, "John Smith · subject of · Marriage of John Smith and Mary Smith") and a place row ("Leeds · part of · Yorkshire") all use it. When the page has no name for the link, the second line names the neighbor: "with Mary Smith". When more than one member link stands behind it, "· N records" follows the link. |
| ES-14 | A relationship with no lines under it shows its header only, plus a muted caption: "Nothing else to compare." Checking it still pins the link. |

### 4.4 Lifespan and likeness

| ID | Requirement |
| --- | --- |
| ES-15 | **Likeness.** A "Likeness" Card on a person, after Name (§2 Order). Each line shows two crops side by side, 96 pt square, this Source's on the left and the member's on the right, each captioned with its Source title. Same image: "Same likeness", Agrees, checked. Different images: "Compare these likenesses", outcome Badge "Judge by eye" (neutral), pinnable, unchecked. A crop that cannot be shown draws the person mark in the Thumbnail. Likeness lines never collapse unless the image is the same. |
| ES-16 | **Lifespan.** A "Lifespan" Card on a person, last among the own Cards. One line, no checkbox, with a muted caption under it: "Worked out from dated events. A guide for matching, never pinned." Each side's range runs from its earliest to its latest dated event and is written as the Person page writes a span ("14 May 1817 – 2 Jun 1890"; one dated event prints one date: "1840"). Fits: "The lifespans fit: 14 May 1817 – 2 Jun 1890 here, 14 May 1817 – 2 Apr 1871 on the record", Badge Agrees. Conflict: "The lifespans don't fit: 14 May 1817 – 2 Jun 1890 here, 3 Mar 1741 – 9 Jan 1810 on the record", Badge Conflicts. One side has none: "No lifespan to compare: the record has no dated events", Badge Not compared. When neither side has a dated event, there is no Lifespan Card. Contradiction: "No lifespan for this source: its own dates contradict each other" (or "for the record"), Badge Not compared. A lifespan conflict counts as a disagreement for ES-2 and ES-4. |

### 4.5 States

| ID | Requirement |
| --- | --- |
| ES-17 | **Loading.** The summary and heading show at once; the Cards area shows three skeleton bars until the evidence arrives. |
| ES-18 | **Failed.** A danger Callout (compact) in the Cards area: "The evidence could not be loaded" with the error. The claim fields stay usable. |
| ES-19 | **No record.** For a row set to New, Skip or unset: the summary Card is replaced by an EmptyState: "Pick a record to see its evidence" — "Choose a record in this row's menu. New and Skip file no evidence." The claim fields still show for New. |
| ES-20 | **Nothing to compare.** No line, no relationship and no Lifespan Card: EmptyState in the Cards area: "Nothing to compare yet" — "The record's other sources say nothing about these properties." |

### 4.6 Claim fields and accessibility

| ID | Requirement |
| --- | --- |
| ES-21 | The Claim block stays below the evidence. The Argument hint becomes "N pieces of evidence checked · say why these are the same" (N counts checked pins). |
| ES-22 | VoiceOver: a line reads as one element, its sentence followed by the checkbox's own state, which the kit speaks (checked, not checked, or partly checked): "Name. Same name, 5 records. This source: John Smith." A member: "Census 1851: John Smith." A relationship: "They share a matching connection. John Smith · subject of · Birth of John Smith." A likeness line names both Sources: "Likeness from this source and Census 1851. Compare these likenesses." The lifespan line has no checkbox and reads its sentence. The disclosure reads "Show 5 records, collapsed". |

---

## 5. Example data and frames

**Draw only this data.** Every value, ref, title, outcome, sentence and count below follows from the app's rules for this one research project. Do not invent other people, change a date, swap a sentence or pick a different outcome: a value that looks harmless can be a case the app never produces (§5.1). When a frame needs something §5 does not give, leave it out and add a note on the board.

### 5.1 Rules every example obeys

| Rule | So on the board |
| --- | --- |
| A line needs a value on **both** sides | A member that says nothing about a property (no sex recorded, no photo) has no line for it. Never draw "unknown" or "—" lines. |
| A member is on another Source | The Source being promoted is never a member's Source. Every member caption names a different Source. |
| The record's members are Subjects filed on it earlier | A record with three members shows at most three records on a line. "· N records" never exceeds the record's members. |
| Collapsing needs the same value | Only members whose values are the same ("John Smith" and "JOHN SMITH") share a line. "J. Smith" is its own line. |
| Names are compared word by word | "J. Smith" against "John Smith" matches by initial; "John Henry Smith" has a missing part; "Jack" is a nickname of "John"; "Wm" abbreviates "William"; "Smyth" is a similar spelling of "Smith"; "Mary Ames" and "Mary Smith" are different names. |
| Dates are compared by the days they can cover | "1817" contains "14 May 1817" (less precise); "abt 1817" is the year 1817; one or two years apart is close; three or more years apart is different. |
| Terms are the same or different | Sex at birth "Male" vs "Male" is the same; a gender "Man" counts as "Male". "Unknown" sex is no evidence and has no line. |
| Some properties allow many values | Toponym, interests and likeness can hold several values, so two different values are not a disagreement and have **no line**. A place called "Leeds" on one Source and "Kirkgate" on another shows nothing for that pair. |
| A disagreement exists only on single-value properties, or against a "not" value | Name, sex at birth and dates can disagree. "This source: not male" against a male member disagrees. |
| Relationships need both rows placed | A relationship appears only when the neighbor row (the birth, the spouse) is set to a record, and the record's members have the same kind of link to that record. A neighbor set to New or Skip gives no relationship. |
| A relationship never disagrees | It matches, or it is **similar** when the neighbor's own facts disagree (a birth placed on the record's birth but dated eight years apart). There is no conflicting relationship and no disabled relationship box. |
| Lifespans are people only | Events and places have no Lifespan Card. A lifespan is the range of the person's own dated events: from the earliest to the latest. |
| Likenesses are people only | The same likeness means the same image file: the identical photo saved on two Sources with the same crop. Two scans of one photograph are two images. |
| Refs | Records: `PER-`, `EVT-`, `PLC-` plus five characters (digits and capitals, no I, L, O or U). Source Subjects: `CPR-`, `CEV-`, `CPL-`. Sources: `SRC-`. |
| Dates and spans as the app prints them | "14 May 1817", "Oct 1843", "1840", "abt 1817"; spans "14 May 1817 – 2 Jun 1890" with spaces around the dash. |
| Event titles | "Birth of John Smith", "Marriage of John Smith and Mary Smith". |
| Connection words | "John Smith · subject of · Birth of John Smith", "John Smith · spouse of · Mary Smith", "Death of John Smith · took place in · Leeds", "Leeds · part of · Yorkshire" (S11-D1's subrow form). |

### 5.2 The project

**Sources.**

| Ref | Title |
| --- | --- |
| SRC-8F2KQ | Leeds Mercury — obituary of John Smith, 4 Jun 1890 *(the Source being promoted)* |
| SRC-1A7TN | 1851 England Census — Leeds, Kirkgate |
| SRC-3D9PW | 1861 England Census — Leeds, Kirkgate |
| SRC-5H4XM | 1871 England Census — Leeds, Briggate |
| SRC-6R2VB | St Peter's, Leeds — births and baptisms, 1817 |
| SRC-2T6GA | St Peter's, Leeds — marriages, 1843 |
| SRC-7C3ZE | Smith family photograph album |
| SRC-9K5JY | Funeral card — John Smith, 1890 |
| SRC-4Q8MR | St John's, Wakefield — baptisms 1741 and burials 1810 |

**Records already in the project.** Each member is listed with its Source and what that Source says.

| Record | Members |
| --- | --- |
| PER-7QH3D John Smith | 1851 Census: "John Smith", Male · 1861 Census: "John Smith", Male · 1871 Census: "J. Smith", Male · births and baptisms 1817: "John Henry Smith" · marriages 1843: "John Smith" · photograph album: "Jack Smith", likeness photo B · funeral card: "John Smith", likeness photo A |
| PER-9B2XW Mary Smith | 1851 Census: "Mary Smith", Female · 1861 Census: "Mary Smith", Female · marriages 1843: "Mary Ames". The record concludes "Mary Smith"; "Mary Ames" is outvoted |
| PER-3X9QV Jack Smith | a second page of the photograph album: "Jack Smith", likeness photo C. No events, no links |
| PER-5J2NB John Smith | St John's, Wakefield: "John Smith", Male; subject of a baptism on 3 Mar 1741 and a burial on 9 Jan 1810 |
| EVT-4RJ9T Birth of John Smith | births and baptisms 1817: Birth, 14 May 1817; John Henry Smith is its subject |
| EVT-6D3HM Baptism of John Smith | births and baptisms 1817: Baptism, 1 Jun 1817 |
| EVT-2W8NC Marriage of John Smith and Mary Ames | marriages 1843: Marriage, 12 Oct 1843; John Smith and Mary Ames are its subjects |
| EVT-8G7PA, EVT-1N4RC, EVT-3H5TB | the 1851, 1861 and 1871 Census Enumerations (30 Mar 1851, 7 Apr 1861, 2 Apr 1871); John is the subject of each |
| EVT-0P7XR Event at Leeds | a parish note that names an event with no type, no date and no place |
| PLC-6M3FD Leeds | 1851 Census: "Leeds" · 1861 Census: "Leeds" · 1871 Census: "Leedes" · births and baptisms 1817: "Leeds" |
| PLC-8Y4HP Yorkshire | 1861 Census: "Yorkshire". On the 1861 Census, Leeds is part of Yorkshire |

**Links between records.** John is spouse of Mary on the 1851 and 1861 Censuses. John and Mary are the subjects of the marriage. John is the subject of the birth, the baptism and the three censuses.

**The obituary's Subjects (the rows).** Photo A is the obituary's portrait. The funeral card printed the same photograph, and the researcher saved the same image file on both Sources.

| Row | Ref | What the obituary says | Proposed |
| --- | --- | --- | --- |
| John Smith | CPR-2M8KD | "John Smith", Male, likeness photo A; subject of the birth, the death and the marriage; spouse of Mary | PER-7QH3D, **Strong**, "via Birth of John Smith → EVT-4RJ9T (subject)" |
| Mary Smith | CPR-4N7QS | "Mary Smith"; subject of the marriage; spouse of John | PER-9B2XW, **Strong**, "via John Smith → PER-7QH3D (spouse)" |
| Birth of John Smith | CEV-3B5RT | Birth, 14 May 1817 | EVT-4RJ9T, **Strong**, "via John Smith → PER-7QH3D (subject)" |
| Death of John Smith | CEV-7F1WD | Death, 2 Jun 1890; took place in Leeds | **New Event** (no record holds his death) |
| Marriage of John Smith and Mary Smith | CEV-9P2KX | Marriage, 1843 | EVT-2W8NC, **Strong**, "via John Smith → PER-7QH3D (subject)" |
| Leeds | CPL-5G8NB | "Leeds"; part of Yorkshire | PLC-6M3FD, **Strong**, "via Yorkshire → PLC-8Y4HP (part of)" |
| Yorkshire | CPL-1V6JC | "Yorkshire" | PLC-8Y4HP, **Medium**, "Agrees on Toponym" |

### 5.3 Frames

Each frame lists the sheet from top to bottom. The count is one pin per member: a line naming four records is four pins.

**Frame 1 — Person, Strong, with relationships (ES-1 – ES-16).** Row: John Smith.
- Title "John Smith → PER-7QH3D John Smith". Subtitle "Strong · via Birth of John Smith → EVT-4RJ9T (subject)".
- Summary: the Strong Badge and "via Birth of John Smith → EVT-4RJ9T (subject)". The row is neither contested nor updated.
- Heading "What's the evidence that this is PER-7QH3D John Smith?" · "16 checked". No Callout: nothing disagrees.
- **Name** (This source: John Smith):
  - "Same name · 4 records", Agrees, checked, "Show 4 records" (1851 Census, 1861 Census, marriages 1843, funeral card).
  - "Same name (nickname)", Agrees, checked — "Smith family photograph album: Jack Smith".
  - "Matches by initial", Similar, unchecked — "1871 England Census — Leeds, Briggate: J. Smith".
  - "Matches; part of the name is missing", Similar, unchecked — "St Peter's, Leeds — births and baptisms, 1817: John Henry Smith".
- **Likeness**:
  - Photo A beside photo A, captioned "This source" and "Funeral card — John Smith, 1890": "Same likeness", Agrees, checked.
  - Photo A beside photo B, captioned "This source" and "Smith family photograph album": "Compare these likenesses", Judge by eye, unchecked.
- **Sex at birth** (This source: Male): "The same · 3 records", Agrees, checked (the three censuses).
- **Lifespan**: "The lifespans fit: 14 May 1817 – 2 Jun 1890 here, 14 May 1817 – 2 Apr 1871 on the record", Agrees, no checkbox, with the caption.
- Relationship "They share a matching connection" — "John Smith · subject of · Birth of John Smith", EVT-4RJ9T, Agrees, checked.
  - Under it, **Date**: "Same date", checked — "St Peter's, Leeds — births and baptisms, 1817: 14 May 1817".
- Relationship "They share a matching connection" — "John Smith · subject of · Marriage of John Smith and Mary Smith", EVT-2W8NC, Agrees, checked.
  - Under it, **Date**: "Same date, recorded less precisely", Similar, unchecked — "This source: 1843", "St Peter's, Leeds — marriages, 1843: 12 Oct 1843".
- Relationship "They share a matching connection" — "John Smith · spouse of · Mary Smith", PER-9B2XW, Agrees, checked, "· 2 records" (the 1851 and 1861 links).
  - Under it, **Name**: "Same name · 2 records", checked (1851 and 1861 Censuses).
- There is no death relationship, because the death row is New.
- Order: own Cards Name, Likeness, Sex at birth, Lifespan; relationships Birth, Marriage, Spouse.
- Count: name 5 + likeness 1 + sex 3 + birth 1 + birth date 1 + marriage 1 + spouse 2 + Mary's name 2 = 16.

**Frame 2 — Frame 1 with the name line expanded (ES-9, ES-10).**
- The member rows, in this order: "1851 England Census — Leeds, Kirkgate: John Smith", "1861 England Census — Leeds, Kirkgate: John Smith", "St Peter's, Leeds — marriages, 1843: John Smith", "Funeral card — John Smith, 1890: John Smith".
- The 1861 row is unchecked, so the line's box is partly checked.
- The disclosure reads "Hide records". The heading reads "15 checked".

**Frame 3 — Event, the researcher's choice, contested (ES-2, ES-4, ES-7).** A variant of the project: the obituary says the couple married "in 1840". The engine does not propose EVT-2W8NC, because the dates are three years apart. The researcher picks it in the row's menu. Row: Marriage of John Smith and Mary Smith.
- Title "Marriage of John Smith and Mary Smith → EVT-2W8NC Marriage of John Smith and Mary Ames". Subtitle "No match, contested · Your choice".
- Summary: the No match and Contested Badges, "Your choice", and "Something below disagrees with this record."
- Heading · "4 checked". Callout.
- **Event type** (This source: Marriage): "The same", checked — marriages 1843: Marriage.
- **Date** (This source: 1840): "Different dates", Conflicts, disabled — marriages 1843: 12 Oct 1843.
- Relationship "They share a matching connection" — "John Smith · subject of · Marriage of John Smith and Mary Smith", PER-7QH3D, checked.
  - Under it, **Name**: "Same name", checked — marriages 1843: John Smith.
- Relationship "They share a matching connection" — "Mary Smith · subject of · Marriage of John Smith and Mary Smith", PER-9B2XW, checked.
  - Under it, **Name**: "Different names", Conflicts, disabled — "This source: Mary Smith", "St Peter's, Leeds — marriages, 1843: Mary Ames". The relationship still matches, because the record concludes "Mary Smith".
- No Lifespan or Likeness Card (this is an event).

**Frame 4 — Frame 1 in the 1840 variant: a similar connection (ES-13).** Only the marriage relationship changes:
- It reads "They share a similar connection", Similar, unchecked.
- Under it, **Date** reads "Different dates", Conflicts, disabled — 1840 against 12 Oct 1843.
- The Callout shows. The heading reads "15 checked" (Frame 1's 16 less the marriage pin). The row is still Strong and not contested.

**Frame 5 — Person: a member disagrees, but the row is not contested.** Row: Mary Smith.
- Title "Mary Smith → PER-9B2XW Mary Smith". Subtitle "Strong · via John Smith → PER-7QH3D (spouse)".
- Heading · "9 checked". Callout.
- **Name**:
  - "Same name · 2 records", checked (1851 and 1861 Censuses).
  - "Different names", Conflicts, disabled — "St Peter's, Leeds — marriages, 1843: Mary Ames".
- No Sex at birth Card (the obituary records none for Mary) and no Likeness Card.
- **Lifespan**: "The lifespans fit: 1843 here, 12 Oct 1843 – 7 Apr 1861 on the record". One dated event prints one date.
- Relationship "They share a matching connection" — "Mary Smith · subject of · Marriage of John Smith and Mary Smith", EVT-2W8NC, checked.
  - Under it, **Date**: "Same date, recorded less precisely", unchecked — 1843 against 12 Oct 1843.
- Relationship "They share a matching connection" — "John Smith · spouse of · Mary Smith", PER-7QH3D, checked, "· 2 records".
  - Under it, **Name**: "Same name · 2 records", checked.
  - Under it, **Sex at birth**: "The same · 2 records", checked (the obituary's John is Male, and so are the census Johns).
- Count: name 2 + marriage 1 + spouse 2 + John's name 2 + John's sex 2 = 9.

**Frame 6 — Event, Strong.** Row: Birth of John Smith.
- Title "Birth of John Smith → EVT-4RJ9T Birth of John Smith". Subtitle "Strong · via John Smith → PER-7QH3D (subject)".
- **Event type**: "The same", checked — births and baptisms 1817: Birth.
- **Date**: "Same date", checked — 14 May 1817.
- Relationship "They share a matching connection" — "John Smith · subject of · Birth of John Smith", PER-7QH3D, checked.
  - Under it, **Name**: "Matches; part of the name is missing", Similar, unchecked — births and baptisms 1817: John Henry Smith.
- Heading · "3 checked". No Callout.

**Frame 7 — Place, Strong.** Row: Leeds.
- Title "Leeds → PLC-6M3FD Leeds". Subtitle "Strong · via Yorkshire → PLC-8Y4HP (part of)".
- **Toponym**:
  - "The same · 3 records", checked (1851 Census, 1861 Census, births and baptisms 1817).
  - "Similar spelling", Similar, unchecked — 1871 Census: Leedes.
- Relationship "They share a matching connection" — "Leeds · part of · Yorkshire", PLC-8Y4HP, checked.
  - Under it, **Toponym**: "The same", checked — 1861 Census: Yorkshire.
- There is no relationship for the death's place, because the death row is New.
- Heading · "5 checked". No Lifespan or Likeness Card.

**Frame 8 — A relationship with nothing under it (ES-14).** A variant: the obituary says only "born in Leeds", with no birth date. Draw only John's birth relationship card: "They share a matching connection" — "John Smith · subject of · Birth of John Smith", EVT-4RJ9T, checked, and "Nothing else to compare."

**Frame 9 — Person, Weak: a likeness to judge, and no lifespan on the record.** In John Smith's row, the researcher picks the menu alternative PER-3X9QV.
- Title "John Smith → PER-3X9QV Jack Smith". Subtitle "Weak · Agrees on Name".
- **Name**: "Same name (nickname)", checked — "Smith family photograph album: Jack Smith".
- **Likeness**: photo A beside photo C: "Compare these likenesses", Judge by eye, unchecked.
- **Lifespan**: "No lifespan to compare: the record has no dated events", Not compared.
- No relationships (the record has no links). Heading · "1 checked".

**Frame 10 — Person, No match, contested: a lifespan conflict.** In John Smith's row, the researcher picks PER-5J2NB.
- Title "John Smith → PER-5J2NB John Smith". Subtitle "No match, contested · Your choice".
- Summary with "Something below disagrees with this record."; Callout.
- **Name**: "Same name", checked — "St John's, Wakefield — baptisms 1741 and burials 1810: John Smith".
- **Sex at birth**: "The same", checked — Male.
- **Lifespan**: "The lifespans don't fit: 14 May 1817 – 2 Jun 1890 here, 3 Mar 1741 – 9 Jan 1810 on the record", Conflicts.
- No relationships. Heading · "2 checked".

**Frame 11 — A source that contradicts itself (ES-16).** As Frame 1, except that the obituary's death was transcribed "2 Jun 1809". The lifespan line reads "No lifespan for this source: its own dates contradict each other", Not compared. Everything else is as in Frame 1.

**Frame 12 — The Updated caption (ES-2).** Mary's row before John was decided: there were two "Mary Smith" records and the row started unset. The researcher set John's row, and Mary moved onto PER-9B2XW. Frame 5's summary adds "Updated · after John Smith → PER-7QH3D was matched".

**Frame 13 — States (ES-17 – ES-20).**
- **Loading**: Frame 1's title, subtitle, summary and heading, with no "N checked"; three skeleton bars.
- **Failed**: the same, with the danger Callout "The evidence could not be loaded" with the read's error under it, here "Catalog closed."; the claim fields stay usable.
- **New**: row Death of John Smith set to New Event. Title "Death of John Smith → New Event"; no subtitle; the EmptyState "Pick a record to see its evidence"; the claim fields shown.
- **Skip**: the same row set to Skip. Title "Death of John Smith"; the EmptyState; no claim fields.
- **Nothing to compare**: the researcher sets Death of John Smith to EVT-0P7XR. Its member has no type, no date and no place, and events have no lifespan. The title carries the No match Badge with "Your choice" (the researcher's pick, no conflict). The EmptyState reads "Nothing to compare yet".

### 5.4 Specimen lines

Some copy-table rows are not reached by the project above. Draw these in one small specimen frame, with these values (this source → member). They are rare; do not put them into the project's frames.

| Sentence | Values |
| --- | --- |
| "Same, written with different accents" | Zoë Brontë → Zoe Bronte |
| "Same name (abbreviated)" | William Ames → Wm Ames |
| "Partly the same" | Toponym: Kirkgate, Leeds → Leeds, Briggate |
| "The dates overlap" | bet 1815 and 1818 → bet 1817 and 1820 |
| "Similar spelling" (a name) | John Smith → John Smyth |
| "Close dates" | 1817 → 1818 |
| "Similar" | J. Henry Smith → John H. Smith (each side has the fuller word the other lacks) |
| "Disagrees · 3 records" with "This source: not Male" (ES-7) | The obituary's sex at birth recorded as a negative "not Male", against three Male members. The row is then contested |

The copy table's "The same, written differently" and "Matches, with less detail" rows are fallbacks: no value the app compares today reaches them. Do not draw them.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sheet shell | Component | **Extend** | `DesignSystem/Components/Panel/PVPanel.swift` | Title, 880 pt, sunken footer, no subtitle. Gains a trailing title accessory: the band and Contested Badges sit at the right end of the title row (§9). |
| Evidence sheet body | Snowflake | **Rethink** | `Features/Promote/PromoteView.swift` `PromoteEvidenceSheet` | Columns, `ExhibitGroup` and the "Through" cards go; summary, heading, Cards and lines come in. One call site. |
| Why-suggested marks | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift`, SwiftUI `.help` (the kit's tooltip on macOS, `DesignSystem/README.md`) | The row's band and Contested Badges, drawn by one shared Promote view the row and the title both instance; the reason is their tooltip. The Updated and contested lines under the title are plain text. Not a Card, not a Callout. |
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
| Claim fields | Component | **Extend** | `DesignSystem/Components/Field/PVField.swift`, `DesignSystem/Components/Select/PVSelect.swift`, `DesignSystem/Components/TextArea/PVTextArea.swift` | Above the evidence (§9). `PVField` gains a muted "(optional)" marker beside `required`'s, for the argument label. No hints. |
| Column headers, Weight column | Snowflake | **Remove** | `PromoteEvidenceSheet` `sheetColumns`, `weightText` | Go with the table. |

**Non-goals.** No new kit primitive (the two kit rows are Extend). No local Badge, Checkbox, Card, Callout or image view. No table component for lines. No numbers. No select-all. No hover actions on a line.

---

## 7. Handoff

The board was agreed on 2026-10-22: its link is on S11-20's **Board** line and its differences are in §9. Move S11-20 from `Blocked — …` to `Planned` with `plan-pr` once S11-32 has landed. Move this brief to the spike's design archive folder when S11-20 lands, and update [`README.md`](../README.md).

---

## 8. Review log

[`../prs/S11-20-claim-evidence-sheet.review.md`](../../prs/S11-20-claim-evidence-sheet.review.md) (round 1, 2026-10-22) covers this brief: inventory rows for the section header, checkbox, likeness tile and readout badges, Go-owned order, and connection wording were changed from it.

---

## 9. Board notes (agreed 2026-10-22)

Checked against `Promote flow.dc.html` and its `EvidenceCard` component (frames E01 – E14, drawn from §5's project). Where the board differs from §4, the board wins and S11-20's plan copies it. Frame values match §5.3 except where a row below says so.

| Frame | Requirement | The board |
| --- | --- | --- |
| E01 – E13 | ES-1, ES-2 | **No subtitle and no summary Card.** The band Badge and, when contested, the Contested Badge sit at the right end of the panel title, top-aligned. The row's reason is their tooltip; the pair is one focusable element that reads "Why this was suggested: Strong, contested, via …". Under the title, only when they apply, a group named "Why this was suggested" holds the Updated caption (italic, info ink) and "Something below disagrees with this record." (primary text). A row with no record (New, Skip) has no Badges. |
| all | ES-21 | **Claim fields come before the evidence.** A 200 pt column with Status over Confidence, beside the argument field. The argument label is "Additional comments or arguments about this claim" followed by a muted "(optional)". No hints under Status or the argument. New shows the claim fields; Skip hides them. |
| all | ES-3 | Under the heading, one sentence in secondary text: "Check every line that supports this match. Checked lines are pinned to the claim when you press Done." The conflict Callout follows it. |
| E01 | ES-5, ES-6 | Cards have the small radius. A property Card's header is its name on the sunken fill. A line is a grid: checkbox, text, disclosure, then the Badge in a 116 pt right column. Under the sentence, one muted line joins its parts with " · ": "This source: John Smith · 1871 England Census — Leeds, Briggate: J. Smith". A collapsed line shows only "This source: …". |
| E02 | ES-9 | The disclosure is a small ghost Button with a chevron (down when collapsed, up when open), spoken "Show 4 records, collapsed" / "Hide records, expanded". Members list under the line, indented to its text, each row with a hairline above: checkbox, value, and the Source title under it in muted text. |
| E01, E03, E05 | ES-12 | A relationship's header: checkbox, the sentence (medium weight) over the link in secondary text with "· N records" muted, the ref in mono, the Badge. Under it, each of the neighbor's properties has a small-caps label ("Date", "Name", "Sex at birth") indented 28 pt, with a divider between properties. |
| E08 | ES-14 | "Nothing else to compare." is italic muted text under the header, indented to the sentence. |
| E01, E05, E09 – E11 | ES-16 | The lifespan line has a faint info icon in the checkbox column. Its caption sits directly under the sentence, in italics. It has no "This source" line. |
| E01, E09 | ES-15 | Crops are 96 pt with an italic caption under each ("This source", then the member's Source title), below the sentence. No "This source" text line. No image files were supplied, so the board's crops draw the person mark. |
| E13a | ES-17 | Three bars at 100%, 86% and 62% width, 12 pt high. The board drew sunken boxes because the web kit has no skeleton bar; the app uses `PVSkeletonBar`. |
| E13b | ES-18 | The board left the Callout's message empty; the app shows the read's error ("Catalog closed." in §5.3). |
| E13c – E13e | ES-19, ES-20 | Both EmptyStates use the file-search icon. E13e left out the title Badges; the app shows No match with "Your choice" (§5.3). |
| E04 | ES-3 | "15 checked" (§5.3 Frame 4 corrected from 14). |
| E14 | ES-8 | Specimen members show the value alone, with no Source. |
| E10 | ES-16 | The board's record was buried in 1838, which overlaps the obituary's 1817 – 1890 life, so the engine finds no conflict (S11-32 Landed note). §5.2 and §5.3 now date the record's baptism 3 Mar 1741 and burial 9 Jan 1810, so the lives do not overlap: "The lifespans don't fit: 14 May 1817 – 2 Jun 1890 here, 3 Mar 1741 – 9 Jan 1810 on the record". The frame's layout stands. |
