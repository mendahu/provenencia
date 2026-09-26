# S8-D11 — Citation composer delete chrome

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-18** (Citation Delete + Observation row on DeleteImpact)  
**Depends on:** Frozen policy, path matrix, and table register in [`../deployment-plan.md`](../deployment-plan.md) §S8-09.1–§S8-09.4; **S8-09**, **S8-12**, and the **S8-D9** recipe (**S8-13**) landed or mocked as the same contract; shipped composer ([**S8-D8**](archive/S8-D8-composer-connect-simplification.md) / **S8-11**) — row Delete… already ships  
**Related:** Shared confirm / blocked notice is [`S8-D9-impact.md`](S8-D9-impact.md) — **instance it, do not restyle.** Graph delete chrome is a **different board** ([`S8-D10-graph-delete.md`](S8-D10-graph-delete.md) → **S8-19**). Do not redraw the canvas.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped S8-D8 composer. Do not start a second form.

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
| Resource delete with inbound check | DeleteImpact recipe — confirm if allowed, notice if blocked |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

---

## 1. Objective

The composer already deletes **one ordinary Observation** (⋯ → Delete… + confirm). This board adds **Citation erase** when nothing points at it, and honest **refuse** when Observations (including a saved connection’s edges) still do.

Policy is frozen in the plan. This board does not invent a cascade that deletes Observations with the Citation.

**This board designs**

1. **Delete citation** on a saved Citation with **zero Observations** (including no connection edges). DeleteImpact **confirm**, then the Citation is gone. Identity falls back to New Citation / another Citation on the Artifact.
2. **Blocked notice** when any Observation remains. Instance DeleteImpact (named `OBS-` rows; edges count). Shared / connect Citations are the same notice — not a special “also drop 2 other subjects” cascade.
3. **Saved connection row** stays without a Delete (S8-11). If the board shows why, it is a short refuse / hint (“this connection is one unit; remove it from the graph” is **wrong** until graph erase exists — say it is locked, or omit Delete). Do not add a second Observation delete for edges.
4. Observation **row** Delete… is the same inbound-resource check (C1–C4, C6). Today `Impact` is empty, so it is always the confirm branch. Put it on DeleteImpact anyway — later claim pins must not need a new composer PR. Empty Citation after the last row is legal; that is when Delete citation becomes available.

Do **not** restyle Layout A / the 1500pt split. Do **not** draw graph cards.

---

## 2. Domain facts (this board)

| Fact | UI implication |
| --- | --- |
| Policy §S8-09.1 | Erase iff inbound *resources* are empty. Facets (notes) CASCADE. App explains; SQLite is the gate. |
| Path matrix §S8-09.2 **C\*** | C1–C4 / C6 already ship. C10 / C12 are this board. C11 refuse. C7–C9 no edge/connection delete. |
| Empty Citation is legal | Save citation with zero rows, or last row deleted. Those Citations need an erase. |
| One Citation → many Observations, many subjects | Deleting the Citation is refused while any row remains. Do not silently empty other subjects. |
| Edge rows are `observations.edge_locked` | They never appear as Observation rows. They still count as inbound on the Citation. |
| Role on a connection is an ordinary Observation | S8-11 does not offer Delete on the connection row. Do not add it here unless the board finds a quiet ⋯; engine would allow the role, not the edges. Prefer leave it. |
| Citation notes | Facets: they CASCADE with the Citation. No composer chrome. Do not ask the researcher to delete notes first. |
| Identity `PVSelect` | Switches Citations; it is not delete. |
| After erase | Stay in the composer on this Artifact. Unsaved-work guard still applies to leftover drafts. |
| Artifact / Citation deleted elsewhere | This composer may be open on a Source-page delete (**S8-14**). Treat as `missingDeepId` / identity fallback — no crash, no leftover Save on a ghost id. |

### 2.1 What this board is not

- Not the Evidence graph (**S8-D10**).
- Not restyling DeleteImpact (**S8-D9**).
- Not a second Observation delete (reuse S8-11).
- Not Source / Artifact delete.
- Not undo, Change type, or Auto Transcribe / PDF chrome.

---

## 3. Implementation gate (S8-18)

| Ships in **S8-18** | Does **not** ship there |
| --- | --- |
| Delete citation + Observation row on DeleteImpact; `citations.Delete` (new) and `observations.Delete` through the register | Other screens; the recipe itself (**S8-13**) |
| Calling `GetDeleteImpact` / `DeleteCitation` | Schema / registry (**S8-09**, **S8-12**) |
| L10n + VoiceOver | A second Observation delete; connection-row Delete |
| Fallback identity after erase | `DeleteObservation` rewrite |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CD-1 | A saved Citation with zero Observations offers **Delete citation**. DeleteImpact **confirm** names `CIT-…`. |
| CD-2 | A Citation with any Observation (ordinary, role, or edges) does not confirm-then-fail. DeleteImpact **notice** lists `OBS-` refs. Rows already on this Citation can focus locally; a shared-Citation row may `go(to:)` its location. |
| CD-3 | Observation row Delete… presents DeleteImpact (`GetDeleteImpact` for that Observation). Today always confirm. Last row may empty the Citation; then CD-1 applies. |
| CD-4 | Saved connection row still has no Delete. Pending **Discard** is unchanged (not a catalog delete). |
| CD-5 | After a successful Citation erase, the composer does not keep a stale `citationID`. Switch to New Citation or another listed Citation. Same recovery if the Artifact or Citation was deleted from the Source page. |
| CD-6 | VoiceOver: delete control + confirm / refuse name the Citation. |
| CD-7 | Citation erase **and** Observation row both present DeleteImpact. Do not fork a composer-only sheet. |

---

## 5. Suggested frames

1. Saved empty Citation — Delete citation in the citation-fields block (near Save citation). Confirm.
2. Citation with two ordinary rows — Delete citation refused with those `OBS-` refs named. Row ⋯ Delete… still works.
3. Citation that is only a saved connection — same refuse (edges count).
4. After last Observation delete — the empty Citation, then frame 1.
5. After Citation erase — New Citation / next identity, no crash, no leftover Save citation dirty on a ghost id.
6. Composer open; Artifact deleted on the Source page — recover, no crash.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Observation row Delete… | Snowflake | **Extend** | `CitationComposerObservationRow` | Same DeleteImpact recipe; do not add a second trash. |
| Citation fields actions | Snowflake | **Extend** | `CitationFieldsDraft` / form pane | Home for Delete citation + Save citation. |
| Connection row | Snowflake | Ship | `CitationComposerConnectionRow` | No Delete. Discard is pending-only. |
| DeleteImpact | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/` | **S8-13**. Citation erase **and** Observation row (inbound empty today). |
| Button | Component | Ship | `PVButton` / `PVIconButton` | Danger only on the confirm branch, not a red extra footer. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| “Delete citation and 7 Observations” | Policy is refuse, not cascade. |
| Connection-row Delete that erases the bridge | Graph **S8-19** trash on the bridge releases connection facets. Do not add a second delete here. |
| Composer-only refuse sheet | Shared recipe is **S8-D9**. |

---

## 7. Out of scope

- Graph cards
- Source / Artifact delete
- ⌘Z
- Redesigning the viewer, locators, or transcription assist

---

## 8. Handoff

1. Archive this brief under [`archive/`](archive/) and write [`../completed.md`](../completed.md).
2. Implement **S8-18** against these frames + §S8-09.2 C-rows. Add `citations.Delete` and cut over `observations.Delete` in this PR.
