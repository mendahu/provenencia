# S14-D1 — Wide layout in the design system

**Kind:** Claude Design board (**enhancement** of the Provenencia design-system project: tokens, kit README, one new layout component)
**Spike:** Provenencia Spike 14 (odds and ends), Batch 6 — Wide layout
**View:** The design system itself: `tokens/spacing.css`, the kit README, and a Page columns specimen. No product screen.
**Implements later as:** PR **S14-25** (Page columns in the kit). S14-24 adds the Swift token and width class first and is not gated by this board.
**Depends on:** —
**Related:** [`docs/macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide) (the breakpoint and the rules), [`../deployment-plan.md`](../deployment-plan.md) working rule "(Batch 6) One breakpoint, 1600pt of pane width"
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

### Two widths — draw both

The window opens at 1366 × 900pt but is often stretched across a large monitor.
Every view has two width classes, measured on the view's own pane (window minus
sidebar and any inspector), not the screen:

- **Regular** — pane below 1600pt. Draw at the default window: 1366 × 900, sidebar open.
- **Wide** — pane 1600pt and up. Draw at 2560 × 1440, sidebar open.

Rules for Wide:

1. **Use the extra space.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. Content jammed left with empty space beside it is never the Wide design. If the view truly has nothing to add, center its content group and say so on the frame.
2. **Keep readable things readable.** Prose, forms and the primary column keep their Regular maximum widths. Add columns; do not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Past every column's maximum, center the group** and let the gutters grow.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Work in the Provenencia design-system project. Add one token, one README section and one layout component with its specimen. Do not redraw existing components, tokens or the Onboarding Flow board. Product screens get their Wide frames in their own briefs, not here.

---

## 1. Objective

Provenencia opens at a laptop-sized window, 1366 × 900pt. Researchers also stretch it across a large monitor, and today every page keeps its 1180pt content column, so on a 2560pt-wide window the page sits in one column with about 1100pt of empty space beside it. The design system gets a **Wide** width class and a layout piece that uses the extra space, so every later brief draws a Wide frame from the kit instead of inventing one.

**What changes.** A breakpoint token; a "Width classes" section in the kit README with the rules; a **Page columns** component: a primary column and an optional secondary column that stack at Regular and sit side by side at Wide.

**What stays.** Every existing token value, including `--width-content-max` (1180), `--width-inspector` (340), `--gutter-page` (32) and `--gutter-page-wide` (56). Every existing component.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| One breakpoint | Regular is a pane below 1600pt; Wide is 1600pt and up. There is no third class. |
| The pane, not the screen | The width is what the page itself gets: the window minus the sidebar (264pt, 78pt collapsed) and minus any inspector beside the page. Collapsing the sidebar can move a page from Regular to Wide. |
| Why 1600 | Every MacBook full screen is Regular (a 16" MacBook Pro gives a 1464pt pane). A 1920pt-wide monitor gives 1656 and a 2560pt one 2296: Wide. 1600 is the smallest pane that holds the 1180 primary column, a 32 gap, a 340 secondary column and two 32 gutters. |
| Pages are one column today | The Source page left-aligns its 1180 column; the Person, Event and Place pages and the lists center theirs. Both leave the extra width empty at Wide. |
| The app is native macOS | The kit renders in a browser for review. In the app, the page measures its own width; the CSS here mirrors that with a container query on the page, not a viewport media query. |
| Same content at both widths | Width changes arrangement only. Nothing exists only at Wide. |

### 2.1 What this board is not

- A Wide frame for any product page (Source page, Person page, lists, Promote, the composer). Each gets one in its own brief.
- A list-beside-detail layout, a docked inspector or a two-column form. They are in the pattern table in the README section, drawn when a view brief needs them.
- A third breakpoint, a compact class or a change to the window's default size.

---

## 3. Implementation gate (S14-25)

| Ships in **S14-25** | Stays out |
| --- | --- |
| `PVPageColumns` in the Swift kit matching the agreed component; the kit README row and the "Width classes" section transcribed from this board. | Adopting it on any page (each page's own brief and PR). The token and the width class (S14-24, already planned). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| WL-1 | `tokens/spacing.css` gains `--width-wide-breakpoint: 1600px` beside the other layout widths, with a comment: measured on the page's pane, not the viewport. |
| WL-2 | The kit README gains a **Width classes** section: the two classes and the breakpoint, the five Wide rules below, and the pattern table (§ 4.1). |
| WL-3 | A **Page columns** component: a primary slot and an optional secondary slot, in a scrolling page with the page gutters. |
| WL-4 | At **Regular**: one column. The primary slot is as wide as the pane allows up to `--width-content-max`; the secondary slot's content stacks below the primary, at the same width. Gutters are `--gutter-page`. |
| WL-5 | At **Wide** with a secondary slot: primary on the leading side up to `--width-content-max`, a `--space-9` gap, secondary on the trailing side at `--width-inspector` (340) growing to 480 as the pane grows. Both columns are top-aligned and scroll together with the page. Gutters are `--gutter-page-wide`. |
| WL-6 | At **Wide** with no secondary slot: the primary column is centered in the pane. It is never left-aligned with empty space beside it. |
| WL-7 | When both columns are at their maximum, the column group is centered in the pane and the gutters take the rest. |
| WL-8 | The secondary column holds supporting content (metadata, linked records, activity), never the page's only copy of an action. At Regular the same content is below the primary, in the same order. |
| WL-9 | Crossing 1600 keeps the page's scroll position and focus. Show a note on the specimen; there is nothing to draw. |
| WL-W1 | This board's own Wide frame: the Page columns specimen at a 2560 × 1440 window, sidebar open, with a secondary column (WL-5). |

**The five Wide rules** (WL-2 copies these into the README):

1. Use the extra space: a second column, the detail beside its list, more grid or table columns, or an inspector docked open. If a view has nothing to add, center it.
2. Prose, forms and the primary column keep their Regular maximum widths. Add columns; do not stretch text or fields.
3. Lists, tables, grids and canvases fill the width. Cards keep their size and add columns.
4. Same content, same actions at both widths.
5. Past every column's maximum, center the group and let the gutters grow.

### 4.1 Pattern table (README, WL-2)

| View shape | Regular | Wide |
| --- | --- | --- |
| List → detail page | List; selecting navigates to the detail | List column (420–520) and the selected detail side by side |
| Record / detail page | One column; supporting sections below | Page columns: primary + secondary column of supporting content |
| Grid | Columns that fit at the card minimum | More columns at the same card size |
| Table | Essential columns | The columns hidden at Regular |
| Form / editor | One column at the form measure | Field groups in two columns, or the form beside a preview or reference panel |
| Canvas | Canvas; inspector on demand | Canvas with the inspector docked open |

---

## 5. Suggested frames

1. **Token and README.** The spacing token sheet with `--width-wide-breakpoint` beside the layout widths (WL-1), and the Width classes README section (WL-2).
2. **Page columns, Regular** at a 1366 × 900 window, sidebar open (pane ~1102): primary then secondary stacked, placeholder Cards in each (WL-3, WL-4).
3. **Page columns, Wide** at a 2560 × 1440 window, sidebar open (pane ~2296): primary and secondary side by side (WL-5, WL-W1).
4. **Page columns, Wide, primary only** at the same window: the centered column (WL-6).
5. **Page columns, very wide** at a 3008 × 1692 window: both columns at maximum, the group centered (WL-7).
6. **At the line**: the same page at pane 1599 (Regular) and 1600 (Wide), with the WL-9 note.

Use neutral placeholder content: Cards with SectionHeaders and a few Field rows. Not a real product page.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Breakpoint token | Token | New | `DesignSystem/Tokens/PVSpacing.swift` | `--width-wide-breakpoint`. S14-24 adds `widthWideBreakpoint` there; this board adds the CSS it is transcribed from. |
| Width class | Token | New | `DesignSystem/Components/WidthClass/` (new, S14-24) | Not drawn. Regular or Wide, read by Page columns. |
| Page columns | Component | New | `DesignSystem/Components/PageColumns/` (new, S14-25) | Replaces three hand-written page scaffolds, each a known call site: `Features/Conclusions/ConclusionDetailPage.swift` `page`, `Features/Conclusions/ConclusionListPage.swift` `page`, and the content column in `Features/Sources/SourcePageView.swift`. |
| Placeholder blocks | Component | Ship | `DesignSystem/Components/Card/PVCard.swift`, `DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `DesignSystem/Components/Field/PVField.swift` | Specimen content only. |
| Spacing tokens | Token | Ship | `DesignSystem/Tokens/PVSpacing.swift` | `--width-content-max`, `--width-inspector`, `--gutter-page`, `--gutter-page-wide`, `--space-9`. Values unchanged. |

**Non-goals.** No list-beside-detail, docked-inspector or two-column-form component on this board. No new Card or SectionHeader variant for the specimen. No viewport media query: the CSS uses a container query on the page.

---

## 7. Handoff

When the board is agreed: move this brief to [`archive/`](archive/), update [`README.md`](README.md), and plan S14-25 with `plan-pr`. S14-25 transcribes `--width-wide-breakpoint` against the value S14-24 shipped; a different value on the board is a question for the user, not a silent change.
