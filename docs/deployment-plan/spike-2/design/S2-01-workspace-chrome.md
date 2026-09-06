# S2-01 — App workspace chrome (sidebar + content)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 2 (Source layer validation)  
**Implements later as:** PR S2-14  
**Depends on:** nothing (start here)  
**Blocks:** Source list/detail boards (S2-02+) should assume this chrome exists

Paste this entire document into Claude Design as the requirements for one board/flow.

---

## 1. Objective

Design the **post-onboarding application shell**: a durable workspace with a **side navigation** and a **single main content host**. Spike 1’s centered “You’re signed in” home must not remain the permanent product chrome. This board is **chrome only** — do not design the Source catalog list or Source editor here (that is S2-02).

---

## 2. Product context (must respect)

1. Provenencia is an offline-first macOS genealogy app. After onboarding, the researcher has an install identity (`display name` + `USR-…` ref) and an open project (catalog label + folder basename such as `robins-family.provenencia`).
2. The Mac client owns windows, navigation, menus, and presentation. Genealogy rules stay out of the UI layer.
3. Visual language: “modern but archival” — warm parchment surfaces, serif display type, Spectral body, IBM Plex Mono for refs/ids, iron-gall accent. Match the shipped onboarding restyle / design-system tokens; do not introduce a purple dashboard aesthetic or a second logo system.
4. Composition rule for this shell: **one clear layout** = leading sidebar (or nav rail) + **one** primary content column. Not a multi-panel IDE, not a card dashboard, not floating widgets on a hero.
5. Spike 2 sidebar destinations that are in scope: **Sources**, **Source types**, and **Source fields**. Do **not** add disabled / “Coming soon” items for Interpretation, People, Settings, or other future areas — add those nav entries only when those features ship.

---

## 3. Requirements

### 3.1 Shell structure

| ID | Requirement |
| --- | --- |
| W-1 | Provide a full-window **app workspace** used after onboarding succeeds (create or open project). |
| W-2 | Include a **leading sidebar or nav rail** listing primary destinations. |
| W-3 | Include a **content host** region that swaps when the selected destination changes. Exactly one primary content surface is visible at a time. |
| W-4 | Show **selected vs idle** states for nav items clearly (do not rely on color alone — use weight, indicator, or surface). |
| W-5 | Provide a **user-controlled toggle** that expands or collapses sidebar destination **labels**. Expanded: icons + text labels. Collapsed: **icon-only rail** (labels hidden). Do **not** auto-collapse based on window width — the researcher chooses. Persist the preference for the install if practical (board need not specify storage; engineering can use UserDefaults). Collapsed icons keep the same selected/idle/disabled treatments; tooltips or accessibility labels must still expose the destination name. |
| W-5b | **Short window (sidebar content taller than the viewport):** the **sidebar scrolls vertically**. Do not clip nav items or session/footer chrome without a way to reach them; prefer a single scrolling sidebar column (logo + destinations + session block) over shrinking hit targets. Content host scrolling is independent and not specified here. |
| W-6 | Preserve room for macOS window chrome (traffic lights); do not draw a fake title bar that fights the system. |

Place the label toggle somewhere obvious in chrome (sidebar edge, rail footer, or near the logo). Project label / folder (W-9…W-10) still need a collapsed strategy if they live in the rail (e.g. slim content-area header). **Sign Out** is menu-bar only (W-11) — do not reserve sidebar space for it.

### 3.2 Session and project identity in chrome

Today’s onboarding home surfaces: logo mark, “You’re signed in”, contributor **display name** + parenthesized **`USR-…`**, project **label**, folder basename, and **Sign out**. Relocate identity and project cues into chrome; **move Sign Out out of the window** into the macOS app menu (see W-11). Do **not** bring created/updated/updated-by project bookkeeping into this chrome.

| ID | Requirement |
| --- | --- |
| W-7 | Surface the **contributor display name** in chrome (sidebar footer is typical). **Expanded sidebar:** show the name as text. **Collapsed sidebar:** replace the name with a **user account icon**; do not show the name inline in the rail. |
| W-8 | Surface the contributor **`USR-…` ref** in monospace. **Expanded sidebar:** show it near the display name. **Collapsed sidebar:** do not show the ref inline; include it (with the display name) in the **hover tooltip** on the account icon from W-7 (e.g. `Jane Smith` / `USR-A1B2C`). Accessibility labels must expose the same data when collapsed. |
| W-9 | Always surface the **project label** (human title from the catalog `project` table), not only the folder path. |
| W-10 | Surface the **project folder basename** (e.g. `robins-family.provenencia`) in a secondary/muted style — useful for disk orientation. |
| W-11 | Provide **Sign Out** only in the macOS **app menu** (`Provenencia` → Sign Out), not as a button or link in window chrome / sidebar. Annotate this on the board (menu-bar callout is enough; Claude Design need not pixel-perfect the system menu). Same action as today’s onboarding Sign Out (clears session / returns to onboarding). Do **not** duplicate Sign Out in the sidebar to “save a click.” |

### 3.3 Navigation destinations

| ID | Requirement |
| --- | --- |
| W-13 | Include an active **Sources** destination (primary catalog entry). Content for Sources may be a labeled empty placeholder on this board (list/detail is S2-02). |
| W-16 | Include two additional **top-level** sidebar destinations for vocabulary admin: **Source types** and **Source fields** (exact labels designer’s call; must read as distinct). Each selects its own content host pane (empty placeholder OK on this board; full editors are S2-04). Do **not** bury these under Sources as a sheet/subflow for Spike 2 — if the nav grows crowded later, we can reconsider grouping. |

Do **not** add placeholder nav items for unimplemented product areas (Interpretation, People/Conclusion, Settings, etc.).

### 3.4 Entry from onboarding

| ID | Requirement |
| --- | --- |
| W-17 | Show at least one frame for **transition**: onboarding complete → workspace with Sources selected (or empty content host ready for Sources). |
| W-18 | Returning launch (active project already open) should land directly in this workspace, not a separate “welcome back” marketing screen. |

### 3.5 Brand and components

| ID | Requirement |
| --- | --- |
| W-19 | Include the **logo mark** somewhere restrained in chrome (sidebar top is typical) — brand-present, not a full-bleed hero. |
| W-20 | Prefer design-system controls already used in onboarding (`Button`, `Icon`, `LogoMark`). Decide whether to introduce a `SidebarNav`-style component; if yes, specify states (idle / hover / selected / disabled). |
| W-21 | Do not put Interpretation credibility badges, evidence-grade chips, or lineage colors into this chrome — those belong to later research UI. |

---

## 4. Screen / frame inventory (minimum)

Produce frames sufficient to review:

1. Workspace, **Sources** selected, empty content placeholder — sidebar **expanded** (labels visible); show all three destinations: Sources, Source types, Source fields.
2. Workspace, **Source types** selected (empty content placeholder) — proves top-level nav swap.
3. Same workspace with sidebar **collapsed** via the toggle — icon-only rail for all three destinations; account icon + tooltip; show where the toggle lives and how project identity compresses (no Sign Out in chrome).
4. Workspace at a **short** height — sidebar scrolled (show that overflow destinations / session chrome remain reachable).
5. Chrome detail: contributor + `USR-…` + project label in the **expanded** state (no Sign Out button).
6. Callout or note: **Provenencia → Sign Out** in the macOS app menu.
7. Optional: after-onboarding entry state if it differs from (1); optional **Source fields** selected frame if it differs from (2).

---

## 5. Out of scope (do not design here)

- Source list rows, create-Source wizard, metadata forms (S2-02).
- Artifact ingest / file preview (S2-03).
- Custom type/field editors (S2-04).
- Audit history browser, sync, accounts, Windows layouts.
- Multi-window document model or tabbed documents.
- Sign Out (or Log Out) controls inside the window / sidebar — app menu only.
- Project created / updated / updated-by bookkeeping in workspace chrome.
- Disabled / “Coming soon” sidebar items for future layers (Interpretation, People, Settings, …).
- Nesting Source types / Source fields under Sources as the primary IA for Spike 2.

---

## 6. Acceptance checklist

- [ ] Sidebar + one content column reads as a single composition.
- [ ] Top-level nav includes Sources, Source types, and Source fields — and nothing else for future layers.
- [ ] Contributor (`USR-…`) and project identity are always reachable (expanded text or collapsed account tooltip / documented project placement).
- [ ] Sign Out is specified in the app menu only — absent from window chrome.
- [ ] Visual language matches onboarding / design-system tokens.
- [ ] A toggle expands/collapses nav labels; width does not auto-collapse the rail.
- [ ] Collapsed account control is a user icon with hover tooltip for name + `USR-…`.
- [ ] Short windows scroll the sidebar instead of clipping destinations or session identity.
- [ ] Board notes any chrome decisions S2-02+ must inherit (nav width, header presence, where content padding starts).
