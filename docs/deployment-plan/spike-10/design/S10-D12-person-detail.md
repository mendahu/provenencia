# S10-D12 — Person detail

**Kind:** Claude Design board (**enhancement** of the Person page)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Person detail — the header and a new likeness section
**Implements later as:** PR **S10-23** (the sex / gender treatment). The same board draws the likeness gallery, which ships in **S10-35**, and the preferred-portrait control, which ships in **S10-39**.
**Board:** not drawn yet. When it is agreed, record here the Claude Design link the implementer opens (answered 2026-10-15: the board reaches the implementer as a Claude Design link).
**Depends on:** the agreed S10-D9 Persons board (the likeness thumb and the preferred portrait as the list thumb).
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-23, S10-35, S10-39; working rules "Sex / gender styling", "A preferred portrait is a display pointer, not a claim", "A depiction's value is its Citation"; gotchas "Gender and sex at birth are different Properties", "Gallery and gender share a page"). [`../depictions-and-likenesses.md`](../depictions-and-likenesses.md). [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.7 and §3.9. The page itself is the S9-D5 board (`ConclusionDetailPage`). The preferred portrait as the Persons list thumb is S10-D9 (PSL-8); it has no control there.
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

Add to the Person page frames. Do not draw a second Person page. Keep the header (80pt thumb, display name, the b. and d. lines, the ref and member count at the right), the Details section with its rows, their badges and Why disclosures, the loading and error states, and the refreshing spinner in the Details header.

S10-23 builds the sex / gender treatment. Draw the likeness gallery and the preferred-portrait control on this same page anyway. S10-35 and S10-39 fill those frames. They do not get another board.

---

## 1. Objective

The researcher can tell a person's gender (or, failing that, their recorded sex at birth) at a glance on the Person page, without reading the Details rows. Later, the same page shows every face found for that person, and lets the researcher pick the one that leads.

**What changes.** A small labeled mark under the name says the gender or the sex at birth, and the header thumb takes a soft tint for it. Later, a Likenesses section lists every crop, and one crop can be marked as the portrait.

**What stays.** The name, the b. and d. lines, the ref and member count, every Details row (including the Sex at birth row and, when recorded, the Gender row, each with its badge and Why), the order of the rows, the loading and error states, and the Event and Place pages, which get none of this.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Gender and sex at birth are two Properties | `gender` is gender identity. `sex_at_birth` is sex recorded or assigned at birth. The page never relabels one as the other. |
| Gender wins | When the person's reconciled gender has one value, the treatment follows it, even if sex at birth disagrees. Only when there is no gender value does sex at birth drive it. |
| Three gender terms are seeded | Man, Woman, Non-binary (keys `man`, `woman`, `non_binary`). Each has its own tint. |
| Researchers add gender terms | A project can add its own gender terms (for example *Two-spirit*). Every such term shares one **other** tint. There is no tint per added term, and there is no product “Other” term. |
| Sex at birth borrows a tint | Female uses the Woman tint. Male uses the Man tint. Intersex and any researcher-added sex-at-birth term use the other tint. The label always says it is sex at birth: “Female · at birth”. |
| Neutral is the page as it is today | No gender and no sex at birth; a mixed value (two sources disagree and neither wins); Unknown or Indeterminate sex at birth. Neutral shows no mark and no tint. The Details row already says Mixed or Nothing recorded. |
| Mixed gender does not fall back | A mixed gender is neutral. It does not drop to sex at birth, and it is not a blend of two tints. |
| The mark carries the meaning | Colour is never the only signal. The mark's text names the value, and VoiceOver reads it the way the Details row does: the field's label, then the value (“Gender, Woman”, “Sex at birth, Female”). |
| Tints come from the kit | Each of the four tints is three colours, named like the kind colours: tint (the fill), line (the outline), and ink (the Badge text). Pick each from `PVColor` tokens the design system already has, and label every swatch with its token name. Do not add a palette colour. Do not use a status tone (success, warning, danger) or the Event or Place kind ink, which already mean something else. |
| A likeness is a crop of a photo (S10-35) | Each likeness is a cropped region of a Citation's image, drawn by the researcher on one Source. It is a record of the person's `likeness` Property. A person can have many. They do not compete like two birth dates; every one is listed in the gallery, and `likeness` is not also a Details row. |
| The earliest leads (S10-35) | With no portrait chosen, the header thumb shows the earliest likeness. No likeness: the person mark, as today. |
| One likeness can be the portrait (S10-39) | The researcher marks one likeness as the portrait. It then leads the header thumb and the Persons list thumb. The others stay in the gallery. Going back to the earliest is on the header thumb's menu. It is a display choice, not evidence. This is the same interaction as choosing a Source's cover on the Source page (S10-D5): a Badge on the chosen card, the button hidden there, a toast after choosing, and the revert on the header thumb. |
| A crop may load or fail | While a crop loads, the tile shows the kit thumb's loading state. A crop that cannot be read shows the person mark, not an error. |

### 2.1 What this board is not

- A gender or sex field on Event or Place pages.
- Editing gender or sex at birth from the header. Values come from Observations.
- A colour per researcher-added term.
- A new colour in the palette.
- A pronoun line, a title (Mr, Mrs), or any text inferred from the value.
- A portrait picker in Promote or on the Persons list.
- A likeness gallery for Places or Events.
- A full-size image viewer. A tile opens the Citation that holds the crop.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S10-23** | PDP-1…PDP-6 and PDP-12. The mark, the thumb tint, the precedence, and the neutral cases. |
| Ships in **S10-35** | PDP-7…PDP-9. The Likenesses section and the earliest likeness in the header thumb. |
| Ships in **S10-39** | PDP-10…PDP-11. Choosing the portrait, and going back to the earliest. |
| Stays out | Event and Place pages. Editing values. A colour per added term. Promote. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PDP-1 | A person whose gender has one value shows a Badge directly under the name, above the b. and d. lines, reading the term's label (“Woman”, “Man”, “Non-binary”, or the researcher's term). |
| PDP-2 | A person with no gender value and one sex-at-birth value shows the same Badge reading “{label} · at birth” (“Female · at birth”). |
| PDP-3 | The header thumb takes the value's tint: the fill behind the person mark (tint) and its outline (line). Four tints: Man, Woman, Non-binary, and other. Female and Male borrow Woman and Man. Intersex and any added term use other. The Badge text uses the same tint's ink, on the kit's neutral Badge fill, with no icon. |
| PDP-4 | Neutral shows no Badge and the thumb exactly as today: no gender and no sex at birth, a mixed gender, a mixed sex at birth with no gender, and Unknown or Indeterminate sex at birth. |
| PDP-5 | The Badge reads to VoiceOver as the field's label, then the value: “Gender, Woman”, “Sex at birth, Female”, the same as its Details row. The thumb's spoken label is unchanged. The Badge has no hover help and is not a button. |
| PDP-6 | Every Details row stays as it is, including Sex at birth and Gender. The Badge does not replace them. |
| PDP-12 | Every frame that shows a tint labels its tint, line, and ink with the `PVColor` token names, so the implementer can copy them. |
| PDP-7 | (S10-35) A Likenesses section under Details: a SectionHeader titled “Likenesses” with the count as its aside, then a wrapping grid of 96pt square crop tiles. Each tile is one record of the `likeness` field. Under it, the record's Source title, one line, truncated, with the full title as help. Order: earliest first. No likeness: the section is absent, and `likeness` is never a Details row. |
| PDP-8 | (S10-35) Clicking a tile opens the Citation that holds that crop, the way a Why record does today. The tile is a plain button over the thumb with the kit focus ring, the same shell as the artifact card's preview on the Source page. |
| PDP-9 | (S10-35) The header thumb shows the earliest likeness, filled to the square. With an image, the tint from PDP-3 is the thumb's outline only. Loading and unreadable crops follow the kit Thumbnail states. |
| PDP-10 | (S10-39) Each tile but the chosen one has a ghost “Use as portrait” button under its title. The chosen tile shows a “Portrait” Badge in the accent tone with the photo icon, like the Source page's Cover Badge, and no button. Choosing shows a success toast. One portrait at most. |
| PDP-11 | (S10-39) The portrait leads the header thumb. The header thumb has a menu with “Use earliest likeness”, shown only while a portrait is chosen, like the Source page's revert on its cover. The gallery order does not change when the portrait changes. |

---

## 5. Suggested frames

One Person page. The later frames add to it. Do not draw a second page.

1. **Woman.** Gender Woman, sex at birth Female. Badge “Woman”, Woman tint on the thumb. Details shows both rows.
2. **Gender wins.** Gender Man, sex at birth Female. Badge “Man”, Man tint.
3. **Non-binary.** Gender Non-binary, no sex at birth. Badge “Non-binary”, Non-binary tint.
4. **Added term.** Gender *Two-spirit* (researcher term). Badge “Two-spirit”, other tint.
5. **Sex at birth only.** No gender, sex at birth Male. Badge “Male · at birth”, Man tint.
6. **Intersex at birth.** No gender, sex at birth Intersex. Badge “Intersex · at birth”, other tint.
7. **Mixed gender.** Two Sources disagree on gender; sex at birth Female. No Badge, plain thumb. Details shows Gender with the Mixed badge.
8. **Neither.** No gender, no sex at birth. The page as it is today.
8b. **Unknown at birth.** No gender, sex at birth recorded only as Unknown. The page as it is today; the Sex at birth row reads its empty text.
9. **Gallery (S10-35).** Woman from frame 1 with three likenesses. The header thumb shows the earliest crop with the Woman outline. The Likenesses section shows three tiles under Details.
10. **Portrait (S10-39).** Frame 9 with the second tile chosen: its “Portrait” Badge and no button, the toast, and the header thumb now shows that crop with its menu open on “Use earliest likeness”. The tiles keep their order.
11. **Dark mode.** Frames 1, 4, and 9 in dark mode, so every tint reads on both surfaces.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page | Snowflake | **Extend** | `Features/Conclusions/PersonDetailView.swift` | The Person summary slot gains the Badge (PDP-1, PDP-2). |
| Shared Conclusion page | Snowflake | **Extend** | `Features/Conclusions/ConclusionDetailPage.swift` | The header reads one thumb slot from the kind (colours now; the crop in S10-35). Event and Place pass none. The Likenesses section needs a slot below Details that today's relationship sections cannot hold; S10-35 adds it. |
| Header thumb | Component | **Extend** | `DesignSystem/Components/Thumbnail/PVThumbnail.swift` | PDP-3, PDP-9. Optional colours (tint and line), framing like its size. Content-agnostic: the thumb does not know about gender. |
| Gender / sex mark | Component | Ship | `PVBadge` (`DesignSystem/Components/Badge/PVBadge.swift`) | PDP-1, PDP-2. The text init with its `foreground` ink, the way a graph card shows its kind. Not a Chip: `PVChip` is a selectable control, and this mark is not pressed. |
| Value → tint | Snowflake | New | beside `PersonDetailContent` in `Features/Conclusions/` | PDP-3, PDP-4. Reads the reconciled gender and sex-at-birth fields; picks Man, Woman, Non-binary, other, or neutral. Tints are existing `PVColor` tokens. Second call site: none; it is a feature snowflake on one page, not a kit primitive. |
| Details rows | Snowflake | Ship | `ReconciledValueRow` | PDP-6. Unchanged. |
| Section header | Component | Ship | `PVSectionHeader` | PDP-7. Title “Likenesses” with the count. |
| Likeness tile | Component | Ship | `CachedThumbnail` (`Platform/CachedThumbnail.swift`) over `PVThumbnail` | PDP-7, PDP-9. Loads the crop; image, loading, and mark fallback states. 96pt. |
| Tile button and caption | Snowflake | Ship | the artifact card preview shell in `Features/Sources/SourcePageArtifactsView.swift` | PDP-7, PDP-8. Plain button, kit focus ring, help; caption is muted one-line text. Not a new card component. |
| Tile data | Snowflake | Ship | `Features/Conclusions/ReconciliationRecord.swift` | PDP-7, PDP-8. Source title and Citation location per record, as the Why uses them. |
| Use as portrait | Component | Ship | `PVButton` ghost, small | PDP-10. Same shape as the Source page's “Use as thumbnail” on artifact cards (S10-D5), hidden on the chosen tile. |
| Portrait mark | Component | Ship | `PVBadge` accent, photo icon | PDP-10. Same as the Source page's Cover Badge. |
| Chosen toast | Component | Ship | `PVToast` | PDP-10. Like the Source page's thumbnail-updated toast. |
| Header thumb menu | Component | Ship | `PVContextMenu` | PDP-11. “Use earliest likeness”, like the Source page's cover revert. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A gender kit component, chip, or token family | One Person page reads two fields. The Badge and existing colours cover it. |
| A new palette colour, or a colour per added term | Extra terms share the other tint. |
| A Chip for the mark | Chips are selectable. This mark is display only. |
| An image viewer or lightbox | A tile opens the Citation. |
| A portrait control on the Persons list or in Promote | The Person page owns the choice. |

---

## 7. Handoff

1. When the board is agreed, record the Claude Design link and the four tints (tint, line, and ink token for Man, Woman, Non-binary, other) on the **Board** line.
2. Implement **S10-23** against PDP-1…PDP-6 and PDP-12. **S10-35** fills PDP-7…PDP-9 and **S10-39** fills PDP-10…PDP-11.
3. The brief stays in `design/` while S10-35 and S10-39 build on it. **S10-39** archives it under `design/archive/` and records it in [`../completed.md`](../completed.md).
