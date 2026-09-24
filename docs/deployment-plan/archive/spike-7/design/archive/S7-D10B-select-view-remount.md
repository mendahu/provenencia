# S7-D10B — Remount Select on this view

**Kind:** Claude Design **child-board** slip (not the main design-system project)  
**Spike:** Provenencia Spike 7  
**Parent:** [`S7-D10-pvselect-native-parity.md`](S7-D10-pvselect-native-parity.md) (Select kit page — paste that on the **main** system only)  
**Implements later as:** PR **S7-15** (Swift cousins). This slip is design-board remount only.

Paste **this entire document** into **one** product view board. Do not paste it into the main design system. You cannot see the parent brief or the other view files — everything you need is below.

---

## Who you are

Match **this** board. Remount only your row. Ignore the others.

| If this board is | Remount these treatments onto kit **Select** |
| --- | --- |
| **Onboarding Flow** | Existing-project dropdown (field) |
| **Source fields** | Create-field data-type dropdown (field) |
| **Sources list** | Type filter + sort chips |
| **NameValue** | Part-type dropdown on each structured part (field) |
| **DateValue** / date editor | Calendar dropdown; month dropdown — replace system `Picker` |
| **Citation composer** | Nested DateValue / NameValue popups only — they must be Select, not a local popup |
| **Source detail** | Nested DateValue calendar / month |
| **Subject fields** | Table column filter chevron, if the board shows one — icon-only Select chip. **Not** type-strip tiles |

If you cannot tell which board you are, stop and ask. Do not remount every row.

---

## Do this (in order)

Select already exists on the **main** Provenencia design system. You are not designing Select. You are not copying Select into this project.

1. **Delete this board’s local design-system cache.**
2. **Refetch the latest bundle** from the main Provenencia design system. Do not continue until the fetched kit lists **Select**.
3. Confirm Select supports: field (Input-matched chrome + chevron); chip (icon + label); optional icon-only chip. **Do not re-document keyboard or pointer behavior here** — that contract lives on the main-system Select kit page. If this view’s popup would need different keys, stop and ask.
4. Replace every treatment in **your** row with an **instance of kit Select**.
5. **Delete the local implementation** — no leftover system `Picker`, borderless `Menu`, or hand-drawn chevron field.

If Select is still missing after refetch: delete this board’s cache and refetch again. If it is still missing, stop. Do **not** draw a replacement Select here.

---

## Do not

- Redesign this view, move sections, or change copy.
- Remount ComboBox search fields, right-click / overflow menus, or segmented chips.
- Turn Evidence graph palette tools into Select.
- Copy Select source into this board’s local design system.
