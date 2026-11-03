# S14-D7 — Promote page: notices in view and Try again

**Kind:** Claude Design board (**enhancement** of the Promote page board, S9-D16 / S11-D1 / S11-D3 / S13-D11: states are added to the existing frames)
**Spike:** Provenencia Spike 14 (odds and ends), batch 5
**View:** the Promote page: its notice area (the error callouts above the rows) and the footer
**Implements later as:** PR **S14-20** (The Promote page recovers and speaks)
**Board:** not drawn yet. Draw on [Claude Design — Promote flow](https://claude.ai/design/p/bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd?via=share), beside S13-D11's frames 27a and 27W, as a new turn "S14-D7 · Notices"
**Depends on:** the agreed S13-D11 board (shipped in S13-23, [#633](https://github.com/mendahu/provenencia/pull/633)); S14-14 (landed [#675](https://github.com/mendahu/provenencia/pull/675)), which gave `PVCallout` its opt-in announcement
**Related:** audit 2 B-12 and B-17 ([`../app-health-audit-2.md`](../app-health-audit-2.md)); working rule "Kit components announce themselves" in [`../deployment-plan.md`](../deployment-plan.md); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S14-20-promote-recovers-and-speaks.md`](../prs/S14-20-promote-recovers-and-speaks.md)
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
| Record or detail page at Regular and Wide | Page columns |
| Choose one of a few options inline | Segmented control (ChipGroup segmented) |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### Two widths — draw both

The window opens at 1366 × 900pt but is often stretched across a large monitor.
Every view has two width classes, measured on the view's own pane (the window
minus the sidebar), not the screen. A view with a side pane (an inspector, a detail
pane) plans that pane inside its Wide frame:

- **Regular** — pane below 1600pt. Draw the pane at 1102 × 900 (the default 1366pt window minus the open sidebar).
- **Wide** — pane 1600pt and up. Draw the pane at 2296 × 1440 (a 2560pt window minus the sidebar).

Draw only the view's own pane. Do not draw the app's sidebar, its navigation items or the window's title bar: they are not part of the board. A frame that needs the toolbar (breadcrumbs, Back and Forward) says so in the brief, and draws only the toolbar above the pane.

Rules for Wide:

1. **The extra space does a job.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. If the view truly has nothing to add, center its content. Content jammed to one side with empty space beside it is never the Wide design.
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Draw on the *Promote flow* board as a new turn "S14-D7 · Notices", starting from S13-D11's frames 27a (Regular) and 27W (Wide). Add states to the page that exists. Do not redraw the header, the rows, the target menu, the record search, the row notes, the connection subrows, the Already promoted block or the footer's layout.

---

## 1. Objective

The Promote page shows its errors in a notice area at the top of the scrolling row list. On a long mapped page the researcher presses Done in the fixed footer at the bottom, and a failure appears at the top, out of sight, so nothing seems to happen. A page that fails to load has no way to try again except leaving and coming back.

**What changes.**
- An error notice that appears after the researcher acts is in view when it appears: the stale notice and the save error after Done, and a proposal error after a row decision. The board decides how (PN-1).
- A failed load offers **Try again** in its callout.
- When Done loses the race and the page proposes again, but that proposal also fails, the stale notice goes and the proposal error shows on its own. Today both show, and the stale notice claims "Provenencia proposed the rows again", which is false.

**What stays.** The notice area's place and order (load error, stale, proposal error, save error), every callout's tone (danger) and copy except the Try again label, the header, the rows and their notes, the footer bar and its Done summary, Cancel and Done, and the Wide arrangement S13-D11 agreed (PA-W1).

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A proposal | Each row decision re-runs Promote's matching over the whole page, keeping every decided row as it is. Done is disabled while one runs. |
| Stale | Done files every row in one write. If anything in the catalog changed since the page's last proposal, Done writes nothing; the page shows the stale notice ("Nothing was promoted — the catalog changed while this page was open" / "Provenencia proposed the rows again. Rows you decided are kept. Review the updated rows and press Done.") and proposes again. |
| A proposal error | A proposal can fail. The rows stay as they were, and the error shows in the notice area. Done stays enabled. |
| A save error | Done failed for a reason other than stale. Nothing was promoted. |
| A load error | The page's first proposal failed, or the Evidence graph could not be read. The page has no rows. |
| Long pages | A mapped census page can hold 30 or more rows plus their connection subrows, several screens tall. |

### 2.1 What this board is not

- A decided row whose record or Subject was deleted or merged in another window. The page keeps today's error for it (user, 2026-11-03).
- Toasts, the leave confirm, the record search, or the rows' Updated caption.
- A new notice kind, a dismiss button on a callout, or a banner pinned over the page.
- VoiceOver. The callouts announce themselves in code (the kit's `announces`), which draws nothing.

---

## 3. Implementation gate

| S14-20 ships | Not S14-20 |
| --- | --- |
| PN-1 – PN-4, PN-W1: the notice in view, Try again on a failed load, the stale notice cleared when its proposal fails | Recovering a decided row whose record or Subject went away (descoped 2026-11-03). The proposal's own cost ([`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md)) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PN-1 | When the stale notice, a save error or a proposal error appears, it is in view without the researcher scrolling. **The board decides** between two options, draws both in frame 2, and picks one: **(a)** the row list scrolls to its top, where the notice area is today; **(b)** the stale notice and the save error show as a Callout directly above the footer bar, outside the scroll area, while the load error and the proposal error stay at the top. Whichever is picked, the footer bar and Done do not move or resize, and the rows keep their order. |
| PN-2 | The load error callout has one secondary Button, **Try again**, in the Callout's actions row. While the page loads again, the Button shows its loading state and the callout stays. When the load succeeds, the callout goes and the rows show. When it fails again, the callout stays, with the new message. |
| PN-3 | When Done's stale re-proposal fails, the stale notice is gone and only the proposal error shows. |
| PN-4 | A proposal error clears when the next proposal succeeds, as today. Try again is not offered on a proposal error: the researcher's next decision proposes again. |
| PN-W1 | At Wide the notices keep the row list's maximum width and center with it, as the rows do (S13-D11 PA-W1). Option (b)'s callout above the footer has the same width and centering as the row list, not the footer bar's full width. Nothing else changes between widths. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Promote page | Header and footer bars span the pane; notices and rows in one column, centered, at most 1200 pt | Header and footer bars span the pane; notices and rows keep their column, at most 1600 pt, centered | Nothing moves into a second column: the page has no supporting content to add beside its rows (Wide rules 1 and 6) |

---

## 5. Suggested frames

Frames 1 and 2 are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others at Regular.

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | PN-2, PN-W1 | A page that failed to load: the header, the load error callout ("Provenencia couldn't read this Evidence graph, so it can't list the rows to promote. Go back to the graph and open Promote again.") with **Try again**, no rows, the footer with Done disabled |
| 1c | PN-2 | Frame 1a with Try again in its loading state |
| 2a / 2b | PN-1, PN-W1 | A mapped 30-row page, scrolled to the bottom; the researcher pressed Done and lost the race. Draw option (a) (the list scrolled to its top, the stale notice above the first section) and option (b) (the stale notice above the footer, the list still at the bottom) side by side, and mark the pick |
| 3 | PN-1 | The same page after a save error ("Provenencia couldn't promote these rows." as a placeholder for any save message), as the picked option shows it |
| 4 | PN-3 | Frame 2 after the stale re-proposal failed: no stale notice; the proposal error ("This subject can't be promoted.") where the picked option puts a proposal error |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Notice area | Snowflake | Extend | `macos/App/Features/Promote/PromoteView.swift` `notices` | Load, stale, proposal and save callouts, in that order. PN-1 decides how a new one comes into view |
| Error callout | Component | Ship | `macos/App/DesignSystem/Components/Callout/PVCallout.swift` | `tone: .danger`, as today. Every notice is a Callout; no banner, no toast |
| Try again | Component | Ship | `PVButton` in `PVCallout`'s actions row | Secondary variant, with its `loading` state while the load runs. Not an IconButton, not a link |
| Footer bar | Snowflake | Ship | `PromoteView.swift` `footer` | Unchanged: Done summary, Cancel, Done |
| Row list | Snowflake | Ship | `PromoteView.swift` | Unchanged. Its scroll position moves only in option (a) |

**Non-goals.** No new kit primitive. No dismiss control on a callout. No second notice area: option (b) moves two of the existing callouts, it does not add a kind.

---

## 7. Handoff

When the board is agreed: record the frames and the pick on the **Board** line here, in [`README.md`](README.md) and on the plan's **Board** line; the plan copies PN-1 – PN-4 and PN-W1 into its Scope and moves from `Blocked — on board S14-D7` to `Planned`. When S14-20 lands, the brief moves to `design/archive/`.
