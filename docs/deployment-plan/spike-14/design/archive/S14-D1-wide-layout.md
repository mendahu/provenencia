# S14-D1 — Wide layout in the design system

**Kind:** Claude Design board (**enhancement** of the Provenencia design-system project: tokens, kit README, one new layout component)
**Spike:** Provenencia Spike 14 (odds and ends), Batch 6 — Wide layout
**View:** The design system itself: `tokens/spacing.css`, the Claude Design kit README, and a Page columns specimen. No product screen.
**Implements later as:** no PR of its own. The first page redesign that uses Page columns builds it: Spike 13's **S13-10** (Person page frame, brief S13-D1). S14-24 added the Swift token and width class (landed #628). S14-25 was descoped on 2026-10-29.
**Board:** agreed 2026-10-29 in the Provenencia design-system project ([Claude Design](https://claude.ai/design/p/0f6c1f29-8408-4735-b114-5ae4c4c18445)); archived the same day
**Depends on:** —
**Related:** [`docs/macos-client-patterns.md` § 9](../../../../macos-client-patterns.md#9-window-widths-regular-and-wide) (the breakpoint and the rules), [`../deployment-plan.md`](../../deployment-plan.md) working rule "(Batch 6) One breakpoint, 1600pt of pane width", [S14-24](../../prs/S14-24-wide-width-class.md)
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** [`S14-D1-wide-layout.review.md`](S14-D1-wide-layout.review.md)

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
Every view has two width classes, measured on the view's own pane (the window
minus the sidebar), not the screen. A view with a side pane (an inspector, a detail
pane) plans that pane inside its Wide frame:

- **Regular** — pane below 1600pt. Draw at the default window: 1366 × 900, sidebar open.
- **Wide** — pane 1600pt and up. Draw at 2560 × 1440, sidebar open.

Rules for Wide:

1. **The extra space does a job.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. If the view truly has nothing to add, center its content. Content jammed to one side with empty space beside it is never the Wide design.
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Work in the Provenencia design-system project. Add two tokens, one README section and one layout component with its specimen. Do not redraw existing components, tokens or the Onboarding Flow board. Product screens get their Wide frames in their own briefs, not here.

---

## 1. Objective

Provenencia opens at a laptop-sized window, 1366 × 900pt. Researchers also stretch it across a large monitor, and today every page keeps its 1180pt content column, so on a 2560pt-wide window a page sits in one column with about 1100pt of empty space around it. The design system gets a **Wide** width class and a layout piece that uses the extra space, so every later brief draws its Wide frame from the kit instead of inventing one.

**What changes.** A breakpoint token and a secondary-column maximum token; a "Width classes" section in the kit README with the rules; a **Page columns** component: a primary column and an optional secondary column that stack at Regular and sit side by side at Wide.

**What stays.** Every existing token value, including `--width-content-max` (1180), `--width-inspector` (340), `--gutter-page` (32) and `--space-9` (32). Every existing component. A page with only a primary column looks the same at both widths.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| One breakpoint | Regular is a pane below 1600pt; Wide is 1600pt and up. There is no third class. |
| The pane, not the screen | The pane is what the page gets: the window minus the sidebar (264pt, 78pt collapsed). In the app the workspace measures the pane once and every page reads its width class. A page with a side pane (an inspector, a detail pane) plans that pane in its own Wide frame. The CSS container query (WL-10) measures the same pane. |
| Why 1600 | Every MacBook full screen is Regular (a 16" MacBook Pro gives a 1464pt pane). A 1920pt-wide monitor gives 1656 and a 2560pt one 2296: Wide. 1600 is about the smallest pane that holds the 1180 primary column, a 32 gap, a 340 secondary column and two 32 gutters (1616); between 1600 and 1616 the primary gives up the difference. |
| Pages are one column at page level today | The Person, Event and Place pages and their lists center one column of up to 1180. The Source page left-aligns its column and pins its title header above the scroll; inside, its overview already has two columns of its own. All leave the extra width empty at Wide. |
| The app is native macOS | The kit renders in a browser for review. The CSS mirrors the app with a container query on the page element, not a viewport media query. |
| Same content at both widths | Width changes arrangement only. Nothing exists only at Wide. |

### 2.1 What this board is not

- A Wide frame for any product page (Source page, Person page, lists, Promote, the composer). Each gets one in its own brief.
- A list-beside-detail layout, a docked inspector or a two-column form. They are in the pattern table in the README section, drawn when a view brief needs them.
- A third breakpoint, a compact class, a Wide-only gutter or a change to the window's default size.

---

## 3. Implementation gate (the first redesign that uses it)

| Ships in **S13-10**, with the Person page frame | Stays out |
| --- | --- |
| `PVPageColumns` in the Swift kit with the primary and secondary slots, matching the agreed component. `ConclusionDetailPage.page` moves onto it, so the Person, Event and Place pages share it; what goes in each slot at Wide is the S13-D1 board's call. `PVSpacing` gains `widthSecondaryMax` (480), transcribed from WL-1, and S14-24's `widthWideBreakpoint` is checked against it. The Swift kit README gets a Page columns row and a link to § 9 of `docs/macos-client-patterns.md`, not a copy of the rules. | The header slot and pinned mode, built with their first caller (the S13-D1 board decides whether the Person header uses it; otherwise the Source page's redesign, [`ideas/wide-layout.md`](../../../../ideas/wide-layout.md)). `ConclusionListPage.page` (its own list redesign). A list-beside-detail layout for the Conclusion lists (their own brief). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| WL-1 | `tokens/spacing.css` gains two layout widths: `--width-wide-breakpoint: 1600px` (comment: compared with the page's pane, not the viewport) and `--width-secondary-max: 480px` (comment: the secondary column of Page columns grows from `--width-inspector` to this). |
| WL-2 | The Claude Design kit README gains a **Width classes** section: the two classes and the breakpoint, the six Wide rules exactly as in this brief's "Two widths — draw both" block, the Page columns sizing rule (WL-5), and the pattern table (§ 4.1). |
| WL-3 | A **Page columns** component with three slots: an optional **header** that spans the column group, a **primary** slot, and an optional **secondary** slot. It owns the scroll; the page owns its background. The header either scrolls with the page or is pinned above the scroll, chosen by the page. Loading, error and empty states go in the primary slot. A slot holds the page's view as it is: Page columns adds no spacing inside a slot, so a page keeps its own stack and its own header spacing. Vertical padding is `--space-8` top and bottom. |
| WL-4 | At **Regular**: one column, as wide as the pane allows inside `--gutter-page` on each side, up to `--width-content-max`, **centered** in the pane. The secondary slot's content stacks below the primary at the same width, `--space-9` below it. |
| WL-5 | At **Wide** with a secondary slot, with `available = pane − 2 × --gutter-page − --space-9`: primary = min(`--width-content-max`, available − `--width-inspector`); secondary = available − primary, between `--width-inspector` and `--width-secondary-max`. The primary grows first (full at a 1616 pane), then the secondary (full at a 1756 pane). Primary on the leading side, secondary on the trailing side, `--space-9` between, both top-aligned, scrolling together. |
| WL-6 | At **Wide** with no secondary slot: the same centered column as Regular. It is never left-aligned with empty space beside it. |
| WL-7 | Past both maximums the column group (1692 wide) is centered in the pane, and the gutters take what is left. Gutters are never narrower than `--gutter-page`. |
| WL-8 | The header spans the column group's width and aligns with it at every width, pinned or not. |
| WL-9 | The secondary column holds supporting content (metadata, linked records, activity), never the page's only copy of an action. At Regular the same content is below the primary, in the same order, and keyboard focus moves through primary then secondary at both widths. |
| WL-10 | The CSS switches classes with an `@container` query on the page element (`container-type: inline-size`, the full pane including gutters) at a literal `1600px`, with a comment naming `--width-wide-breakpoint`: a custom property cannot be used in a container condition. |
| WL-W1 | Wide, growing: the specimen at a 1920 × 1080 window, sidebar open (pane 1656): primary 1180, secondary 380 (WL-5). |
| WL-W2 | Wide, capped: the specimen at a 2560 × 1440 window (pane 2296): primary 1180, secondary 480, the group centered (WL-7). |
| WL-W3 | Wide, primary only: the specimen with no secondary at 2560 × 1440, centered, beside its Regular frame at 1366 × 900 (WL-6). |
| WL-W4 | Wide, pinned header: the specimen with a pinned header at 2560 × 1440; the header spans and aligns with the centered group (WL-8). |

Crossing 1600 keeps the page's scroll position and focus (rule 5). Show it as a note on the specimen; a static frame cannot show it.

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

1. **Tokens and README.** The spacing token sheet with the two new widths beside the layout widths (WL-1); the Width classes README section (WL-2).
2. **Regular** at a 1366 × 900 window, sidebar open (pane 1102): primary then secondary stacked, a scrolling header above (WL-3, WL-4).
3. **Regular, wide laptop** at a 1728 × 1117 window (pane 1464): the same page, the 1180 column centered (WL-4).
4. **Wide, growing** at 1920 × 1080 (WL-W1).
5. **Wide, capped** at 2560 × 1440 (WL-W2).
6. **Primary only**, Regular 1366 × 900 and Wide 2560 × 1440 side by side (WL-W3).
7. **Pinned header** at 2560 × 1440 (WL-W4).
8. **At the line**: windows 1863 and 1864 wide (panes 1599 and 1600), Regular and Wide, with the rule-5 note.

Use neutral placeholder content: Cards with SectionHeaders and a few Field rows. Not a real product page.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Breakpoint token | Token | New | `DesignSystem/Tokens/PVSpacing.swift` | `--width-wide-breakpoint`. S14-24 adds `widthWideBreakpoint` there first; this board adds the CSS. Read by the width class and by WL-10's comment. |
| Secondary-column maximum | Token | New | `DesignSystem/Tokens/PVSpacing.swift` | `--width-secondary-max` (480). Read by Page columns; the pattern table's list column is the next reader when its brief comes. |
| Width class | Component | New | `DesignSystem/Components/WidthClass/` (new, S14-24) | Not drawn. Regular or Wide, measured by the workspace, read by Page columns. Ship once S14-24 lands. |
| Page columns | Component | New | `DesignSystem/Components/PageColumns/` (new, S13-10) | First caller in S13-10: `Features/Conclusions/ConclusionDetailPage.swift` `page` (Person, Event and Place pages). Later: `Features/Conclusions/ConclusionListPage.swift` `page` in the lists' redesign. Next: the Source page content column and its pinned `Features/Sources/SourcePageIdentityHeader.swift` header, in the Source page's own brief. The Conclusion lists stay primary-only until their own list-beside-detail brief. |
| Placeholder blocks | Component | Ship | `DesignSystem/Components/Card/PVCard.swift`, `DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `DesignSystem/Components/Field/PVField.swift` | Specimen content only. |
| Spacing tokens | Token | Ship | `DesignSystem/Tokens/PVSpacing.swift` | `--width-content-max`, `--width-inspector`, `--gutter-page`, `--space-8`, `--space-9`. Values unchanged. `--gutter-page-wide` is not used. |

**Non-goals.** No list-beside-detail, docked-inspector or two-column-form component on this board. No new Card or SectionHeader variant for the specimen. No viewport media query. No Wide-only gutter.

---

## 7. Handoff

When the board is agreed: move this brief to `archive/`, update [`README.md`](../README.md), write the board's paragraph in `docs/deployment-plan/spike-14/completed.md`, and tell Spike 13: S13-D1 pulls the fresh kit and draws the Person frame on Page columns, and S13-10 builds it. That PR checks `--width-wide-breakpoint` against the 1600 S14-24 shipped; a different value on the board is a question for the user, not a silent change.
