# S11-D4 — Design system: ComboBox and omnibar parity with the shipped app

**Kind:** Claude Design **design-system project** (the main Provenencia kit, not a screen board) — **enhancement**: update the kit's ComboBox, omnibar field, omnibar results panel and omnibar hit row in place so they match what the macOS app ships
**Spike:** Provenencia Spike 11 (Promote matching layers)
**View:** Design-system components — ComboBox (remote results), Omnibar field, Omnibar results panel, Omnibar hit row
**Implements later as:** — (no PR: the Swift code below is already shipped; this brings the kit up to it, so boards drawn after it, S11-D3 first, instance the real behavior)
**Depends on:** #554 (combo box searching line, landed 2026-10-25); S11-37 (landed [#551](https://github.com/mendahu/provenencia/pull/551): the omnibar's search state and hit row moved to `macos/App/Features/Catalog/Search/` for every find-a-record surface)
**Related:** [`S11-D3-promote-target-menu.md`](archive/S11-D3-promote-target-menu.md) (the first board that composes a remote-results ComboBox with omnibar hit rows), [`macos/App/DesignSystem/README.md`](../../../../macos/App/DesignSystem/README.md) § "The combo box subset", [`docs/deployment-plan/archive/spike-3/omnibar-search.md`](../../archive/spike-3/omnibar-search.md), [`docs/design-system-layers.md`](../../../design-system-layers.md) § "Claude Design / web kit (parity)"
**Board:** the Provenencia design-system project in Claude Design (the "main project" every board pulls its kit from)
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

### This brief is an enhancement of the kit itself

This brief is pasted into the **design-system project**, not a screen board. Steps 1 – 3 above are about boards that pull this kit; here they mean: make sure you are editing the project's own current components, not a cached copy. Step 4 holds: change the kit's existing ComboBox component and the omnibar pieces in place. Do not add a second ComboBox, a "SearchComboBox", a second results panel or a second hit row. Where the kit already does what a requirement says, leave it and note it in the handoff.

The macOS app is the source of truth for every behavior here: each requirement names the Swift file it comes from.

---

## 1. Objective

Boards drawn from the kit should show the search controls the app actually ships, so a board like S11-D3 (a ComboBox that searches the catalog and lists omnibar hit rows) can be drawn by instancing kit pieces instead of drawing local copies.

**What changes in the kit.**

1. **ComboBox** gains, or is checked against, its **remote results** mode: the caller searches, the box shows the results unfiltered, a **searching line** while the first results load, results already in that stay while a new search runs, and a caller-drawn empty line (which also carries a failed search). Its row slot takes any caller content, the omnibar hit row included.
2. **Omnibar field**, **results panel** and **hit row** match the shipped sizes, tokens, states and copy.
3. The kit's specimens show the states, so boards copy them instead of guessing.

**What stays.** The ComboBox's single-select contract (seeded vs typed text, the Spotlight rule, clamping arrows, two-stage Escape, click-away blurs, one highlight for pointer and keys). Select, ContextMenu, Input, Badge, Thumbnail, Callout, as they are. The kit adds no new component.

---

## 2. Domain facts

| Fact | What it means for the kit |
| --- | --- |
| Remote results | Set when the caller passes a query-change handler. The box then never filters: the options *are* the caller's results, in the caller's order (the catalog search's ranking). `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` `onQueryChange` |
| Loading | `isLoading` is true while the caller's search runs. With **no options in yet**, the list shows the searching line in place of the empty line. With options already in, they stay until the new ones replace them; no spinner over them. `PVComboBoxMatch.listLine(optionCount:isLoading:)` |
| Failure | Not a box state. A failed search is the caller's empty line, worded by the caller. |
| The searching line | A small native spinner, then "Searching…" in muted body-small text, aligned to the option labels' column (after the 22 pt checkmark gutter). Copy key `designSystem.comboBox.searching`. |
| The empty line | Caller content, given the typed query (empty when nothing is typed). The box sets it in body-small *italic*, muted, at the labels' column, wrapping. Default kit form: one plain-text line ("No source types match"). |
| The list | Its own window under the field (never an overlay clipped by a scroll view), exactly the field's width, 4 pt gap, never covering the field; flips above only when it does not fit below and there is more room above; scrolls inside a max height of 288 pt. Card surface, 1 pt default border, medium radius, the window's own shadow. Rows: 22 pt checkmark gutter (accent check on the committed row), caller content, small vertical padding, highlight fill `surface-selected`, disabled rows at 42 % opacity. |
| Highlight | The matched span of a plain row's label and subtext wears the highlighter mark (`mark-background` / `mark-foreground`). Remote rows draw their own content. |
| The omnibar field | The toolbar's search: Input, small size, leading search icon, trailing suffix "⌘K", placeholder "Search people, sources, places and files", spoken "Search everything". Flexible width: 420 pt at a comfortable window, compressing to 200 pt before the breadcrumbs give up more room. `macos/App/Features/Workspace/WorkspaceToolbar.swift` `omnibarField` |
| When the panel shows | Once the query has two or more characters and a search has run or is in flight. Searches wait 180 ms after the last keystroke. Escape or a click outside closes the panel and keeps the query. `macos/App/Features/Catalog/Search/CatalogSearchModel.swift`, `macos/App/Features/Workspace/OmnibarResultsModel.swift` |
| The results panel | 640 pt wide, its trailing edge on the field's trailing edge, 8 pt below the field, over the content column (not a popover). Card surface, 1 pt subtle border, small radius, overlay elevation, 4 pt inner padding. The list grows to its content up to 420 pt, then scrolls. Spoken "Search results". `macos/App/Features/Workspace/OmnibarResultsHost.swift` `OmnibarResultsPanel` |
| Panel states | **Loading** (no hits yet): small spinner + "Searching…", muted body-small. **No matches**: "No matches for “Polly”." in body-small medium, primary, then "Check the spelling, or paste a catalog reference such as SRC-3K9M2" in caption, muted. **Failure**: a danger Callout, "Search couldn’t finish. Try again." (or the search's own message). **Hits**: one hit row per hit. |
| Panel rows | Each hit is a menu-row button: fill `surface-hover` on hover, `surface-selected` for the keyboard row, `surface-active` while pressed; extra-small radius. ↑ / ↓ move the keyboard row and stop at the ends; Return opens the row's record. |
| The hit row | `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift`. A 40 × 40 lead (a Source's cover or type icon; a Thumbnail with the kind's mark for people, events and places; a glyph for types and metadata fields). Then: the **title** in body-small semibold, primary, one line, truncating, with a **kind Badge** (neutral, subtle) at the right of the title line; under it the **secondary** line in caption, muted (a person's life line "1871 – 1950 · Bristol", an event's date and place, a place's parent chain), with the **ref** at its right in mono micro, faint, or accent when the match was on the ref; and an optional **match line** in caption, faint ("Note: …", "Metadata: …", "Filename: …", "Name: …", "Place: …", "Description"). Padding 9 pt horizontal, 7 pt vertical, 12 pt gap after the lead. |
| Hit titles | A person, event or place hit is titled from its list header (a person sorted: "Jones, Polly"); other hits use the indexed title. `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` |
| Hit kinds | Source, Type, Metadata field, Person, Event, Place (the Badge text). A caller can ask for some kinds only (a person picker asks for Person). |
| Spoken hit | One element: title, kind, secondary line, ref, match line, in that order ("Jones, Polly, Person, 1871 – 1950, PER-7QH3D"). |
| Who uses them | The omnibar today (panel + hit rows). S11-D3 next: a ComboBox in remote mode whose rows are hit rows, so the ComboBox draws the highlight and the checkmark gutter and the hit row draws no selection of its own. |

### 2.1 What this brief is not

- A screen board. No Promote, Source or workspace frames.
- New behavior. Anything the app does not ship (multi-select tokens, a "use what you typed" create row, a person row kind in ComboBox) stays out, as in the app's "combo box subset".
- A redesign of the omnibar. Its look is the one Spike 3 agreed; this brief only brings the kit's copy of it up to date.

---

## 3. Implementation gate

| Kit change | Ships in | Stays out |
| --- | --- | --- |
| KP-1 – KP-9 | No PR: the app already ships them. Boards drawn after this brief instance them | Multi-select ComboBox, create row, per-hit band or score, a kind-less hit row variant |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| KP-1 | **ComboBox remote results.** The ComboBox has a remote-results mode: setting a query-change handler turns off local filtering, and the options are shown exactly as passed. Document it on the component with one specimen. |
| KP-2 | **ComboBox searching line.** In remote mode with `isLoading` and no options, the list shows a small spinner and "Searching…" (muted body-small) at the labels' column, in place of the empty line. With options already in, the list keeps them and shows no spinner. Two specimens: first search loading; a new search running over earlier results. |
| KP-3 | **ComboBox empty line is caller content.** The empty line takes caller content and the typed query, styled by the box (body-small italic, muted, wrapping, labels' column). Specimens: a plain "No source types match" line; a caller line that shows a failed search's message. |
| KP-4 | **ComboBox custom rows.** The row slot takes any caller content inside the box's row shell (gutter, highlight, hit area). Specimen: a remote ComboBox at 330 pt wide whose rows are Omnibar hit rows for three people, the first highlighted, one with a match line. The hit rows draw no fill of their own here; the box's `surface-selected` highlight is the only selection. |
| KP-5 | **ComboBox list geometry.** The list matches §2: field width, 4 pt gap, never over the field, flips up only when needed, 288 pt max then scrolls, card surface, default border, medium radius. |
| KP-6 | **Omnibar field.** Input, small, search icon, "⌘K" suffix, the placeholder in §2, 420 pt wide with a 200 pt floor. |
| KP-7 | **Omnibar results panel.** 640 pt, trailing-aligned to the field, 8 pt below, card surface, subtle border, small radius, overlay elevation, 420 pt max list height. States: loading, no matches (title + hint), failure (danger Callout), hits. Frames for each. |
| KP-8 | **Omnibar hit row.** Matches §2 slot for slot: 40 pt lead, title + kind Badge, secondary + ref (faint, accent on a ref match), optional match line, 9 / 7 pt padding. Hover, keyboard and pressed fills come from the panel's menu-row button, not from the row. Specimens: a Source hit with a cover, a Person hit with a life line, an Event hit, a Place hit with a parent chain, a ref-match hit with the accented ref, a hit with a match line. |
| KP-9 | **Copy.** The kit's strings for these pieces read as §2: "Searching…", "Search people, sources, places and files", "No matches for “…”.", "Check the spelling, or paste a catalog reference such as SRC-3K9M2", "Search couldn’t finish. Try again.", and the kind Badges "Source", "Type", "Metadata field", "Person", "Event", "Place". |

---

## 5. Suggested frames

1. **ComboBox · remote results**: the field with "Pol" typed and the list loading (KP-1, KP-2); the same with three plain results in and a new search running (KP-2).
2. **ComboBox · empty lines**: no match, and a failed search (KP-3).
3. **ComboBox · hit rows**: the 330 pt remote ComboBox with three Person hit rows (KP-4).
4. **ComboBox · geometry**: the list below a field, and flipped above a field near the bottom of the screen, scrolling at 288 pt (KP-5).
5. **Omnibar · field and panel**: the toolbar field (KP-6) with the panel open on six mixed hits, the second keyboard-selected and the fourth hovered (KP-7, KP-8).
6. **Omnibar · states**: loading, no matches, failure (KP-7).
7. **Omnibar hit row · specimens**: the six rows of KP-8, at 640 pt.

---

## 6. UI building-block inventory

This table is **binding**. Each row is an existing kit piece updated in place, with the shipped Swift it must match.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| ComboBox | Component | **Extend** | `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift` | Remote results, searching line, caller empty line, custom rows (KP-1 – KP-5). The one searchable pick in the kit. |
| Input (omnibar field) | Component | Ship | `macos/App/DesignSystem/Components/Input/PVInput.swift` | The omnibar field is an Input with icon and suffix (KP-6), not its own control. |
| Omnibar results panel | Snowflake | **Extend** | `macos/App/Features/Workspace/OmnibarResultsHost.swift` `OmnibarResultsPanel` | Geometry and states (KP-7). Stays a workspace snowflake. |
| Menu-row button | Component | Ship | `macos/App/DesignSystem/Components/Button/PVButton.swift` `PVMenuRowButtonStyle` | Hover, keyboard and pressed fills for panel rows. |
| Omnibar hit row | Recipe | **Extend** | `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift`, filled by `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` | Slots and type (KP-8). Used in the panel and inside ComboBox rows (KP-4). |
| Kind Badge | Component | Ship | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` | Neutral, subtle, on the hit row's title line. |
| Lead tile | Component | Ship | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` | 40 pt; a kind mark for people, events and places. |
| Failure message | Component | Ship | `macos/App/DesignSystem/Components/Callout/PVCallout.swift` | Danger tone in the omnibar panel. |

**Non-goals.** No new component, recipe or variant. No second ComboBox for search. No fill or selection drawn by the hit row itself inside a ComboBox. No multi-select, create row or person row kind. No change to the omnibar's agreed look beyond matching the app.

---

## 7. Handoff

When the kit is updated: add a **Kit notes** section here listing what changed and anything the kit already had, record the date in [`README.md`](README.md), and move this brief to `design/archive/`. S11-D3's board then pulls the fresh kit (its step 3) before drawing.
