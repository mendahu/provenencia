# S7-D10 — Select (native-popup contract + remount)

**Kind:** Claude Design board / design-system component (kit page + view remount)  
**Spike:** Provenencia Spike 7  
**Implements later as:** PR **S7-15**  
**Depends on:** Shipped macOS [`PVSelect`](../../../../macos/App/DesignSystem/Components/Select/PVSelect.swift) (already in the app kit — this brief **documents the native-popup contract** and remounts cousins; not a greenfield invent)  
**Related:** [`S7-D7`](archive/S7-D7-card-component.md) remount pattern. Type-to-filter fields stay [`PVComboBox`](../../../../macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift). Action menus stay `PVContextMenu`. Segmented `PVChip` groups are not Select.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md) — **Extend** `PVSelect`; compose it at call sites. Child remount slip: [`S7-D10B-select-view-remount.md`](S7-D10B-select-view-remount.md)

Paste this document into Claude Design as the requirements for a **Select** kit page on the **main** design system only. Child view boards cannot see this file — remount them with **S7-D10B**, one board at a time.

---

## 1. Objective

Publish **Select** on the Provenencia Claude Design system whose **visual** contract matches shipped `PVSelect` and whose **behavior** matches a native macOS popup button (`NSPopUpButton`).

The kit page is not chrome-only. **It must document the full interaction contract** — every closed-field and open-menu behavior in §3 — as first-class kit copy (tables, state frames, and annotations). S7-15 implements that board, then copies the same contract into `DesignSystem/README.md`. Do not leave behavior as an engineering footnote.

Variants (no visual restyle):

- **Field** (default): full-width control matching Input rest border + chevron
- **Chip** (`icon` set, compact): toolbar filter/sort
- Optional **icon-only** chip (table column filter) if the kit page needs a label-hidden trigger

**After Select is on the main design system, this brief’s scope includes going into every view in §2, pointing those call sites at the shared Select, and deleting local `Picker` / `Menu` dropdowns.** Do not leave a parallel system popup on those boards. Claude Design stale-bundle habit: **clear cache and rebundle the main system first**; each child board **deletes its local cache and refetches** before remounting.

---

## 2. Views that reference Select

These are the Claude Design **views** that must remount on Select once the kit page is ready. ComboBox search fields, right-click / overflow `PVContextMenu`s, and segmented chips are **not** on this list.

| View | Board | Status | What to remount |
| --- | --- | --- | --- |
| **Onboarding** (Choose file) | [Onboarding Flow](https://claude.ai/design/p/b51790c9-7f65-4e91-a7ef-a900095c5870?via=share) | Already `PVSelect` | Existing-project dropdown |
| **Source fields** (S2-02) | Locate **Source fields** in the Claude Design project | Already `PVSelect` | Create-field data-type dropdown |
| **Sources list** (S2-04 / S5-D2) | Locate **Sources list** in the Claude Design project | Already `PVSelect` (chip) | Type filter + sort chips |
| **NameValue** (S7-D5) | Locate the **NameValue** board | Already `PVSelect` | Part-type dropdown on each structured part |
| **DateValue** (S2 template) | [`date-value-editor.dc.html`](../../../archive/spike-2/design/templates/date-value-editor.dc.html) / locate the DateValue board | **S7-15** migrate | Calendar + month — today SwiftUI `Picker` |
| **Citation composer** (S7-D4) | [Citation composer](https://claude.ai/design/p/8b120853-4ba4-4c0d-8492-75e76b9f9b7a?via=share) | Hosts DateValue / NameValue | Confirm those nested editors use Select, not a local popup |
| **Source page** (S2-23) | [Source detail](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share) | Hosts DateValue | Same — DateValue month/calendar become Select |
| **Subject fields / Source fields tables** | [Subject fields](https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share) + Source fields | Kit-internal | `PVTable` column filter is `Menu`+`Picker` — remount on Select (icon-only chip) even if no live filter is wired yet |

### Remount rule (in scope)

Once Select exists on the **main** Provenencia design system:

1. **Clear the main design-system cache and rebundle** before any view remount.
2. On **each** §2 child board, **delete the local design-system cache and refetch** from the main system. Do not remount until that board’s bundle shows Select.
3. Open each board in the table.
4. Replace every local popup / `Picker` listed under **What to remount** with kit Select (field or chip as that site already uses).
5. **Remove the local implementation** — no leftover system popup, borderless `Menu`, or hand-drawn chevron field that duplicates Select.
6. Leave ComboBox, context menus, and segmented chips alone.

If a child board still cannot see Select after refetch, treat that as a stale-bundle failure. Do not copy Select into the child project.

### Call sites (Swift → view)

| View | File | Chrome today |
| --- | --- | --- |
| Onboarding | [`OnboardingOpenPicker`](../../../../macos/App/Features/Onboarding/OnboardingOpenPicker.swift) | `PVSelect` field — existing project |
| Source fields | [`SourceFieldsDetailPane`](../../../../macos/App/Features/SourceFields/SourceFieldsDetailPane.swift) | `PVSelect` field — data type (create only) |
| Sources list | [`SourcesListView`](../../../../macos/App/Features/Sources/SourcesListView.swift) | `PVSelect` chip — type filter + sort |
| NameValue | [`NameValueEditorForm`](../../../../macos/App/Features/Names/NameValueEditorForm.swift) | `PVSelect` field — part type |
| DateValue | [`DateValueEditorForm`](../../../../macos/App/Features/Dates/DateValueEditorForm.swift) | SwiftUI `Picker` — calendar (~255) and month (~307) |
| Table kit | [`PVTable`](../../../../macos/App/DesignSystem/Components/Table/PVTable.swift) `filterMenu` | `Menu` + inline `Picker` — icon-only chevron/filter |

**Not Select** (do not remount):

| Interaction | Examples |
| --- | --- |
| Type-to-filter combo | Source type assign, metadata suggestions, composer term picker, connect disambiguation, Subject fields bind search — all `PVComboBox` |
| Action / overflow menu | Composer Clear locator, Source cover menu — `PVContextMenu` |
| Segmented exclusive chips | Date point/range, Observation polarity — `PVChipGroup` |

---

## 3. Behavior contract (must appear on the kit page)

Claude Design: put this on the **Select** kit page as documented product behavior, not as a hidden comment. Mirror a native macOS popup button. If a row would differ from `NSPopUpButton`, stop and ask — do not invent a custom policy.

**Required on the board:**

1. A **closed vs open** interaction table (copy the table below onto the page).
2. **State frames** with labels: rest; keyboard-focused (focus ring); menu open below; menu flipped above; long list scrolling; committed checkmark vs hover/keyboard highlight (two rows must not look equally selected).
3. **Key annotations** on those frames (↑/↓, a–z, Space, Return, Esc, Home/End) so an engineer can implement without this brief.
4. **Pointer annotations:** click to open; press–drag–release tracking; click away dismisses without commit.
5. **Accessibility notes** on the page: VoiceOver role is popup button; announce current value; announce expanded/collapsed. One tab stop on the trigger — the menu is not a second tab stop.
6. Field, chip, and (if present) icon-only variants share **one** contract. Do not write a second keyboard story per variant.

### Closed (focused trigger) vs open (menu showing)

| Input | Closed | Open |
| --- | --- | --- |
| Click | Open | Choose the row under the pointer |
| Press–drag–release | Open on press; highlight follows the pointer; release on a row commits | Same tracking loop |
| Space / Return | Open | Commit the highlighted row |
| ↑ / ↓ | **Change the value** (do not open). Clamp at both ends — do not wrap. | Move highlight. Clamp. Enter commits. |
| a–z, 0–9 | Type-to-select: commit the matching option **without opening**. Buffer resets after ~800ms; a single keystroke cycles the next prefix match. | Jump highlight to the match; Enter commits. Escape does **not** keep a type-select that was not committed. |
| Escape | No-op | Dismiss; restore the value from when the menu opened |
| Home / End | First / last option (commit, stay closed) | First / last highlight |
| Click away | — | Dismiss without commit (same as Escape) |

### Also document (not only the table)

- **Selected vs highlight** — committed row keeps a distinct mark (checkmark / selected trait). Keyboard and hover highlight are a separate fill.
- **Placement** — open below the trigger when there is room; flip above when the list would clip the screen. Width follows the trigger (minimum width allowed). Long lists scroll; they do not grow off-screen.
- **Empty / disabled** — empty options cannot open; a disabled Select does not open and shows the disabled opacity already used by other kit controls.
- **What this is not** — on the same page, one line each: ComboBox (type-to-filter), context/overflow menu (actions, not exclusive choice), segmented chips (exclusive but not a popup).

---

## 4. Implementation gate (S7-15)

| Ships in S7-15 | Does **not** ship there |
| --- | --- |
| Native-popup **behavior** on `PVSelect` as documented on the S7-D10 kit page | `PVComboBox` rewrite |
| Migrate DateValue `Picker`s and `PVTable.filterMenu` onto `PVSelect` | Option sections / separators / icons / disabled rows |
| Icon-only chip slot **if** the table filter needs it | Visual restyle of field/chip |
| DesignSystem README interaction contract (like `PVTable`) | New product call sites |
| Existing `PVSelect` hosts keep compiling | Replacing action menus or segmented chips |

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| SL-1 | Kit page shows field + chip (+ icon-only if needed). Chrome matches shipped `PVSelect`. |
| SL-2 | Kit page **documents the entire §3 contract** on-canvas: closed/open table, state frames, key and pointer annotations, placement, a11y, and “what this is not.” A chrome-only Select page is incomplete. |
| SL-3 | Selected row is distinct from hover/keyboard highlight (checkmark vs fill) — shown as a labeled pair of rows, not only described. |
| SL-4 | No new kit primitive besides Select; menus stay `PVContextMenu` under the hood. |
| SL-5 | UI inventory: Select = Component **Extend**; DateValue + `PVTable` filter = Snowflake / kit **Extend** onto Select. ComboBox / context menu / chip = **Ship** as-is. |
| SL-6 | After Select is on the main design system, update **every** §2 board: point remount sites at kit Select and **delete local popups**. Child boards do not re-spec behavior — they inherit the kit page. |
| SL-7 | After Select lands: clear cache and **rebundle** the main design system; on each §2 board delete the local cache and **refetch** from main. |

---

## 6. Out of scope

- ComboBox type-ahead lists
- Right-click / overflow action menus
- Segmented chip groups
- Evidence graph palette tools (those are armed toggles, not a select)
- Redesigning DateValue or NameValue layout beyond swapping the control

---

## 7. Handoff

1. Publish Select on the main design system **with the §3 contract on the kit page**. Clear cache and rebundle that system (SL-7). A page that only shows field/chip chrome is not done.
2. Paste [`S7-D10B`](S7-D10B-select-view-remount.md) into **one** §2 board at a time. Each board clears its cache, refetches, remounts Select, and deletes local popups (SL-6).
3. Archive this brief under `archive/` when the board is agreed.
4. Record in [`../completed.md`](../completed.md).
5. Implement **S7-15** against the **kit-page contract** (then keep `DesignSystem/README.md` in sync).
