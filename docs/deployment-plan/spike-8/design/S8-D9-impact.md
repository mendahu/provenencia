# S8-D9 — Shared delete confirm / blocked notice

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-13** (DeleteImpact recipe only)  
**Depends on:** Frozen policy + Impact report in [`../deployment-plan.md`](../deployment-plan.md) §S8-09.1–§S8-09.4; **S8-12** `GetDeleteImpact` landed or mocked as the same payload  
**Related:** Every later delete board instances this recipe: Source page (**S8-D12**), Source Types (**S8-D13**), Source Fields (**S8-D14**), Subject Fields (**S8-D15**), composer (**S8-D11**), graph (**S8-D10**). Do not redesign those surfaces here.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is a **rethink of resource-delete confirmation**. Today every trash uses a yes/no Confirm (“Are you sure?”). That sheet cannot show why a delete is blocked. Replace the **entry point** with one recipe that branches on the Impact report. Do not fork a second Confirm component; compose the shipped Confirm for the clear path.

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
| Confirm replace or destroy (inbound empty) | Confirm (`item:` snapshot, not a Bool) |
| Resource delete blocked (inbound resources) | **This recipe’s notice** — not Confirm-as-acknowledge |
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

One **common** delete presentation for every official resource / vocab trash. The caller already has an Impact report (`GetDeleteImpact`). This recipe only **renders** it.

**Two destinations — the caller does not choose:**

1. **`allowed`** → today’s destructive **confirm** (yes / keep). Name the target + ref. Optional quiet line that nothing else references it. Confirm runs `Delete`.
2. **`!allowed`** → a **notice** (not a confirm). Title: you can’t delete this. Body: grouped dependencies with **refs + titles + tappable links** (`location`). One dismiss. No “Delete anyway.”

**Today:** `.pvConfirm` is yes/no only. Graph and composer would each invent a refuse. Vocab screens disable trash from a `used_by` count and never name the holders.

**This board designs** the recipe (panel, grouping, row, overflow, VoiceOver). Use Citation + Subject examples so the payload is concrete. Do **not** draw the Evidence graph or the composer form.

Do **not** invent cascade copy (“also delete these 3 Observations”).

---

## 2. Domain facts (this board)

| Fact | UI implication |
| --- | --- |
| Policy §S8-09.1 / report §S8-09.4 | Engine returns facts (`via`, `kind`, `ref`, `title`, `location`, `total`, `gate`). **This recipe** L10n writes group headings, overflow, extra-gate copy, and unknown-`via` fallback. Screens only name the target. |
| `allowed` iff the row exists, extra gates pass, and every resource-edge count is 0 | Confirm only on that path. Facets / owned outbound are not listed. `not_found` / `edge_locked` / `infra` are notices, not confirms. |
| Groups are per **via** | “Cites this card” (`observations.subject_id`) ≠ “used as an endpoint” (`value_subject_id`). Do not merge them. |
| List cap 20 + uncapped `total` | “And 14 more” when `total > listed.count`. Do not fetch the rest in the sheet. |
| Links are `WorkspaceLocation` | Tapping a row `go(to:)` that place (or the host focuses a local row). Dismiss the notice on navigate. |
| Edge-locked Observations still appear | They are why a Citation / bridge can’t go. The notice does not offer delete on that row. |
| Preview then write | Sheet is fed by `GetDeleteImpact` (a fetch). Confirm calls `Delete` **only** when `allowed`. A blocked write is a race or a client bug — generic `in_use` toast, not a second notice. |
| Second call site is already known | Source page, three vocab pages, composer, graph. **New recipe is allowed.** |

### 2.1 What this board is not

- Not graph card chrome (**S8-D10**).
- Not composer Delete citation placement (**S8-D11**).
- Not Source / Artifact / vocab screen layout (**S8-D12**…**S8-D15**).
- Not the register, probes, or FFI.
- Not undo, Change type, or a cascade.

---

## 3. Implementation gate (S8-13)

| Ships in **S8-13** | Does **not** ship there |
| --- | --- |
| DeleteImpact recipe + previews + L10n + VoiceOver | Wiring any screen (**S8-14**…**S8-19**) |
| Confirm branch composes `.pvConfirm` / `PVConfirm` | Wiring screens; metadata-clear `.pvConfirm` (facet) |
| Notice branch: groups, ref rows, overflow, link action; recipe L10n for `via` / `kind` / extra gates / unknown `via` | Engine / `GetDeleteImpact` (**S8-12**); target-noun copy |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| DI-1 | One call site API: present the recipe with a target (kind + ref + title) and an Impact report. Caller does not pick confirm vs notice. |
| DI-2 | `allowed` → Confirm (`item:` snapshot). Destructive confirm, keep/cancel. Copy names the target. No blocker list. |
| DI-3 | `!allowed` → notice sheet / pop-up. Not a confirm. No destructive button. Dismiss only (and row links). |
| DI-4 | Notice lists groups in engine order. Each group heading is speakable from `via` + `kind` + `total` (**recipe** L10n). Each listed row shows **ref** + title; the row is a link. Unknown `via` still lists refs. |
| DI-5 | Overflow when `total > listed.count`. Honest remainder. No “load more” in this spike. |
| DI-6 | Tapping a blocker runs the provided `go(to:)` / focus action and dismisses the notice. |
| DI-7 | VoiceOver: confirm path says erase; notice path says blocked and names the first blockers + totals. |
| DI-8 | Compose `PVPanel` / Confirm. No scrim/dim. No second dialog stack. Width stays in the Confirm / notice family (not a form). |
| DI-9 | Recipe lives under `DesignSystem/Recipes/DeleteImpact/` (`PV*` public type). No catalog models inside Components/. |
| DI-10 | Extra-gate reports (`not_found`, `edge_locked`, `infra`, `origin_locked`) use the notice path: no destructive button, recipe copy, no invented inbound list. |

---

## 5. Suggested frames

1. **Clear Citation** — `CIT-7KD45`, allowed. Confirm yes/no. Baseline vs today’s Confirm.
2. **Blocked Citation** — three Observations (`OBS-…` + titles). Notice, not confirm. Rows look tappable.
3. **Blocked Subject (G2)** — dashed / uncited-looking person. Two groups: none as `subject_id`, one+ as endpoint. Copy must not say “uncited.”
4. **Overflow** — `total` 27, listed 20, remainder 7.
5. **VoiceOver / reduced width** — groups still parse; refs remain readable.
6. **Unknown via / extra gate** — a future `via` still lists refs; `edge_locked` / `not_found` / `origin_locked` are notices.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Confirm (clear path) | Component | Ship | `.pvConfirm` / `PVConfirm` | Compose; do not fork. |
| DeleteImpact recipe | Recipe | **New** | `DesignSystem/Recipes/DeleteImpact/` | Branches on `allowed`. ≥2 call sites (graph, composer). |
| Blocked notice panel | Recipe (part of above) | **New** | same | Panel + dismiss; not Confirm tone / trash button. |
| Blocker group | Recipe (part of above) | **New** | same | Heading from `via` + `kind` + `total`. |
| Blocker row | Recipe (part of above) | **New** | same | Ref + title; tap → location. Chip/Badge only if it helps scan. |
| Callout | Component | Ship | `PVCallout` | Optional lead-in on the notice (“Remove these first”). |
| Button | Component | Ship | `PVButton` | Dismiss on notice; danger only on the confirm branch. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Confirm-as-acknowledge for the blocked path | That is still a yes/no trap. |
| “Delete anyway” / cascade | Policy. |
| Load-more / infinite list in the sheet | Cap + overflow is enough. |
| Rewriting metadata-clear Confirm | Facet write, not a resource Delete. Observation row **is** an inbound check (empty today) — composer **S8-18** puts it on this recipe. |
| New Component-layer Confirm | Extend / compose the shipped one. |

---

## 7. Out of scope

- Graph / composer placement of the trash control
- Source / Artifact / vocab page redesign (recipe only)
- Engine probes
- ⌘Z

---

## 8. Handoff

1. Archive this brief under [`archive/`](archive/) and write [`../completed.md`](../completed.md).
2. Implement **S8-13** against these frames. Later screen PRs instance the recipe; they do not restyle it.
