# S7-D6 — Curated marks (unified mark pack)

**Kind:** Design-system brief (light Claude Design / handoff notes — not a full product surface board)  
**Spike:** Provenencia Spike 7  
**Implements later as:** PR **S7-12**  
**Depends on:** Shipped [`PVEvidenceIcon`](../../../../macos/App/DesignSystem/Recipes/EvidenceIcon/PVEvidenceIcon.swift) + [`EVIDENCE-ICONS.md`](../../../../macos/App/DesignSystem/Recipes/EvidenceIcon/EVIDENCE-ICONS.md) (migrate into **Marks**); [`PVSubjectIcon`](../../../../macos/App/DesignSystem/Recipes/SubjectIcon/PVSubjectIcon.swift) (S6 Canvas paths + S7-D2 **source** folio); Subject fields type strip ([`SubjectFieldsView`](../../../../../../macos/App/Features/SubjectFields/SubjectFieldsView.swift)); registry presentation tokens from **S7-01**  
**Related briefs:** [`S7-D3`](S7-D3-evidence-graph-updates.md) — graph cards / palette consume the consolidated marks; do not redesign card chrome here. Archived [`S7-D2`](../archive/S7-D2-subject-fields.md) — **source** subject mark SoT (not on the original S6 graph board).  
**Schedule:** After schema/Go (**S7-03**) on the checklist, **before** Add property / graph chrome thickening (**S7-09**). Independent of Citations/Observations code — design-system only.

### Claude Design boards (artwork SoT)

Open these when harvesting glyphs or checking stroke weight — do not redesign the surfaces:

| Board | URL | Pull from it |
| --- | --- | --- |
| **Evidence graph** (S6 cards / palette / bridges) | https://claude.ai/design/p/2239e965-3b09-4c13-b85a-d54316ffd8fb?via=share | person, event, place, relationship, participation, location marks |
| **Subject fields** (S7-D2 type strip) | https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share | **source** folio mark (+ parity check for other type-strip glyphs) |

Paste this document into Claude Design only if you need visual confirmation of stroke weight / size at canvas zoom. Prefer a short implementer handoff: pack membership, key naming, tint API, and call-site migration. Read the shared product facts in [`README.md`](../README.md) first.

---

## 1. Objective

Pull the **subject type marks** (person / event / place / relationship / participation / location / **source**) into one curated **Marks** recipe at the design-system root — same asset-backed, template-tinted pipeline as today’s evidence icons, not a parallel `Canvas`-path implementation.

**Destination (locked):** `DesignSystem/Recipes/Marks/` — broader than “evidence icons” once `subject_*` joins `file_*` and `type_*`. **S7-12** renames/moves `Recipes/EvidenceIcon/` + `Assets.xcassets/EvidenceIcons/` into that home (update imports, README, and docs in the same PR).

**Artwork is currently split across two designs.** Reconcile them in this brief before **S7-12** exports assets:

| Kind | Design SoT today | Notes |
| --- | --- | --- |
| person, event, place | [Evidence graph board](https://claude.ai/design/p/2239e965-3b09-4c13-b85a-d54316ffd8fb?via=share) → `PVSubjectIcon` | On graph cards + palette |
| relationship, participation, location | Same graph board / S6-D2 → `PVSubjectIcon` | On bridge cards; Subject fields may share / collapse chrome |
| **source** | [Subject fields board](https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share) → `SourceMark` in `PVSubjectIcon` | Folio mark for the reification type. **Not** on the original Evidence graph board — Subject fields added it. This brief **must pull that glyph into Marks** so graph and fields are not two art pipelines. |

```text
Today
  Recipes/EvidenceIcon/   ← Assets.xcassets/EvidenceIcons (file_* + type_*)
  Recipes/SubjectIcon/    ← in-file Canvas strokes (S6 graph + S7-D2 source)

After S7-12
  Recipes/Marks/          ← one curated pack + Assets.xcassets/Marks (or equivalent)
    file_*     MIME stand-ins
    type_*     source-type marks (evidence / Source *types*)
    subject_*  subject / bridge / source *kind* marks (incl. subject_source)
  Graph + Subject fields (type strip / bindings) call the shared View API
  Recipes/EvidenceIcon/ and Recipes/SubjectIcon/ retired (shim OK briefly, then delete)
```

**Keep this brief small.** Do not redesign Evidence graph cards, palette tool toggles, Subject fields IA, or Source-type picker chrome. Those stay **S7-D3** / **S7-D2** / existing Source types UI.

---

## 2. Domain facts the UI must reflect

| Fact | Implication |
| --- | --- |
| Evidence marks are template assets | Subject marks must import the same way (`provenencia_<key>` imagesets, `.renderingMode(.template)`) so call sites can tint. |
| Subject marks are fixed product kinds | Seven glyphs only for Spike 7 — not researcher-picked like `source_types.icon_key`. Keys still belong in the curated enum / pack for one API. |
| **Source mark SoT is Subject fields** | Export `SourceMark` / S7-D2 folio art from the [Subject fields board](https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share) into Marks; do not invent a second source metaphor for the graph. |
| Graph zooms | Vector template assets (not SF Symbols / bitmaps). Dogfood stroke weight at min/max magnification after conversion. |
| Registry presentation | **S7-01** `Presentation` icon token should name a **curated mark key**, not an SF Symbol, once this lands — so **S7-09** registry migration does not invent a second icon channel. |
| Semantic families stay distinct | `file_*` / `type_*` / `subject_*` remain separate key namespaces; consolidation is **pipeline + View API**, not one flat picker of unrelated marks. Do not confuse **source** (subject kind) with **`type_*`** (Source *type* evidence marks). |
| Folder name is **Marks** | Not `EvidenceIcon` — the pack is broader than Source/Artifact chrome. |

### 2.1 What this board is not

- Not Evidence graph Add-property / cited rows — **S7-D3**.
- Not Subject fields layout redesign — **S7-D2** (only harvest the **source** mark art and migrate call sites).
- Not Source-type icon picker redesign (already ships for `type_*`).
- Not replacing Lucide / inventing a third icon language.
- Not making subject marks catalog-editable in Spike 7.

### 2.2 Implementation gate (S7-12)

| Ships in S7-12 | Does **not** ship there |
| --- | --- |
| `Recipes/Marks/` with all three families (move evidence + add subject, incl. **source**) | Card chrome / Add property (**S7-09**) |
| Shared View API + tint / size conventions documented | Researcher subject-icon picker |
| **All** `PVSubjectIcon` call sites migrated: Evidence graph (cards, palette, bridges) **and** Subject fields type strip / binding chrome | New mark metaphors beyond today’s seven kinds |
| Registry presentation token aligned to curated keys (or explicit follow-up note if FFI already shipped SF Symbol tokens) | Full S7-D3 / S7-D2 visual redesign |
| `EvidenceIcon` / `SubjectIcon` recipe folders retired | Leaving a permanent dual home |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| CM-0 | **Home = `DesignSystem/Recipes/Marks/`.** Move today’s EvidenceIcon recipe + asset catalog into Marks; update all `PVEvidenceIcon` call sites / Xcode references in the same PR. Do not keep `EvidenceIcon` as a second folder. |
| CM-1 | Export today’s `PVSubjectIcon` path art (24×24 board viewBox) to template SVG/PDF imagesets under the Marks asset catalog. Preserve stroke character; verify at canvas zoom. Harvest graph glyphs from the [Evidence graph board](https://claude.ai/design/p/2239e965-3b09-4c13-b85a-d54316ffd8fb?via=share). |
| CM-1b | **Source mark:** treat Subject fields / [S7-D2 board](https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share) `SourceMark` as authoritative. Pull it into Marks with the S6 graph marks — one pack, no leftover Canvas-only source glyph. |
| CM-2 | Extend the curated key enum under Marks with stable string keys for the seven kinds. Prefer a clear prefix (e.g. `subject_person`, `subject_source`) so families stay greppable. Rename types sensibly (`PVMarkKey` / `PVMark` or keep `PVEvidenceIcon*` as transitional aliases only if needed, then prefer Marks-named API). |
| CM-3 | **Tint API:** document and keep call-site recoloring. Template rendering + ambient / explicit `.foregroundStyle` (or a thin wrapper param if the recipe already centralizes tint). Colors are **not** baked into assets. |
| CM-4 | **Size API:** align with evidence presets and/or raw `CGFloat` as needed for cards (≈12–17pt), palette, and Subject fields strip (≈14–16pt); document recommended sizes. |
| CM-5 | Migrate **every** `PVSubjectIcon` call site onto Marks: Evidence graph cards, palette, bridge chrome, **and Subject fields** type strip / binding rows. Grep for `PVSubjectIcon` / `PVSubjectIconKind` and clear them. Delete or thin the old SubjectIcon recipe once call sites are gone. |
| CM-6 | Accessibility: decorative vs named — match today’s evidence-icon patterns where the mark is the sole affordance; graph cards / fields strips may keep marks decorative when the adjacent label already names the kind. |
| CM-7 | Colocate `MARKS.md` (replace/redirect `EVIDENCE-ICONS.md`) + DesignSystem README so implementers know one pack at `Recipes/Marks/`, three families (`file_*` / `type_*` / `subject_*`), that **`subject_source` came from S7-D2**, and how to tint. |
| CM-8 | UI building-block inventory (§7): **Marks** Recipe; call-site snowflakes **Extend** (swap mark only); **Retire** `EvidenceIcon` folder name + `SubjectIcon`. |

---

## 4. Tint / color API (design SoT for implementers)

Lock the call-site contract in the brief so **S7-12** does not invent per-feature hacks:

1. **Assets are monochrome templates** — ink comes from SwiftUI environment / `.foregroundStyle`, same as today’s `PVEvidenceIcon` and `PVSubjectIcon`.
2. **Kind pigment** (person wash vs place wash, etc.) stays **outside** the mark asset: registry presentation tokens → `PVColor` on the parent (card chip, palette tool, Subject fields strip). The mark inherits that foreground.
3. **Do not** ship multicolor “original” rendering for subject marks unless a future brief explicitly needs it.
4. If the shared View grows an optional `tint:` / `style:` parameter, it must remain a thin pass-through to foreground styling — not a second color system.

Propose the exact Swift surface in the handoff notes under `Recipes/Marks/` (e.g. `PVMark(key:size:)` with family helpers, or thin wrappers). Prefer one public recipe entry point over two divergent Views.

---

## 5. Screen / frame inventory (minimum)

Only if Claude Design is used for visual QA (start from the boards linked above):

1. All **seven** subject marks at 12 / 15 / 17 / 24 / 40 pt — including **source** folio from Subject fields — template tinted with kind pigment.
2. Same marks on a zoomed-out and zoomed-in card mock (stroke still readable).
3. Subject fields type-strip mock: person / event / place / bridges / **source** using Marks (parity with shipped S7-05 chrome).
4. Side-by-side: `type_*` evidence mark vs `subject_source` mark at tile size (families distinct — Source *type* vs subject kind `source`).

Otherwise: implementer checklist in **S7-12** dogfood is enough.

---

## 6. Out of scope

| Out | Why / where |
| --- | --- |
| Card / palette / connect UX redesign | **S7-D3** / **S7-09** |
| Subject fields IA redesign | **S7-D2** / already shipped **S7-05** — migrate icons only |
| Catalog-backed subject `icon_key` like Source types | Product kinds stay registry-fixed in Spike 7 |
| Unifying with SF Symbols `PVIcon` | Different layer — chrome vs curated research marks |
| New subject metaphors | Only migrate the seven existing kinds (S6 + S7-D2 source) |
| Keeping `Recipes/EvidenceIcon/` forever | Destination is **Marks** |

---

## 7. UI building-block inventory (binding for design + implement)

Provenencia UI is layered as **components / recipes / snowflakes** ([`docs/design-system-layers.md`](../../../../../design-system-layers.md)). This table is the repo SoT for *what* **S7-12** may touch. Paths are from `macos/App/` unless noted. Template: [`S7-D3`](S7-D3-evidence-graph-updates.md) §9.

**How to read**

| Column | Meaning |
| --- | --- |
| **Layer** | Frost layer: Component (`DesignSystem/Components/<Name>/`), Recipe (`DesignSystem/Recipes/<Name>/`), or Snowflake (`Features/…`). |
| **Status** | **Ship** = already correct; compose as-is. **Extend** = exists; this brief changes it. **New** = create. **Retire** = delete or thin after migration. |
| **Home** | Intended file / folder after **S7-12**. |

Do **not** invent new design-system **components** for marks. Curated research artwork stays a **Recipe**. Call sites stay snowflakes (swap the View API only).

### 7.1 Curated mark pack (design-system)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| **Marks** recipe | Recipe | **New** home (**Extend** content) | `DesignSystem/Recipes/Marks/` | Move `EvidenceIcon` here; add `subject_*`. One public View API + tint contract (§4). |
| Mark key enum / families | Recipe | **Extend** | `DesignSystem/Recipes/Marks/` | `file_*` / `type_*` / `subject_*`. Document `subject_source` provenance (Subject fields board). |
| Mark assets catalog | Recipe assets | **Extend** / rename | `Resources/Assets.xcassets/Marks/` (move from `EvidenceIcons/`) | `provenencia_subject_*.imageset` template SVGs. Graph board + Subject fields board as art SoT. |
| Marks doc | Recipe docs | **Extend** / rename | `DesignSystem/Recipes/Marks/MARKS.md` | Three families; tint API; size presets; a11y; board links. Retire or redirect `EVIDENCE-ICONS.md`. |
| DesignSystem README recipe row | Docs | **Extend** | `DesignSystem/README.md` | Row for **Marks**; remove `EvidenceIcon` + `SubjectIcon` rows after migrate. |
| **EvidenceIcon** (old home) | Recipe | **Retire** | Was `DesignSystem/Recipes/EvidenceIcon/` | Fold into Marks — do not leave dual folders. |
| **SubjectIcon (Canvas)** | Recipe | **Retire** | Was `DesignSystem/Recipes/SubjectIcon/PVSubjectIcon.swift` | Short-lived shim OK; delete in-PR once call sites use Marks. |
| SF Symbols `PVIcon` | Component | Ship | `DesignSystem/Components/Icon/PVIcon.swift` | Unrelated chrome icons — do not merge into Marks. |

### 7.2 Call sites (migrate to Marks)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| Existing `PVEvidenceIcon` hosts | Snowflake / Recipe consumers | **Extend** | Sources list, Source types, thumbnails, vocabulary chrome, etc. | Update imports to Marks; behavior unchanged (`file_*` / `type_*`). |
| Primary subject card mark | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Swap `PVSubjectIcon` → Marks; keep size (~15) + kind pigment from parent. No card redesign (**S7-D3**). |
| Bridge card mark | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Same swap (~12pt). |
| Graph palette tool marks | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphPalette.swift` | Same swap (~17pt). Palette stays a snowflake tool-toggle row — not `PVButton`. |
| Graph create / a11y mark helpers | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphView.swift` (+ `EvidenceCanvasInputMode` kind→mark maps) | Remap to curated `subject_*` keys. |
| Subject fields type strip / chips | Snowflake | **Extend** | `Features/SubjectFields/SubjectFieldsView.swift` | Swap marks on type strip + binding chrome (~14–16pt). **Source** mark must keep working. |
| Subject fields kind mapping | Snowflake | **Extend** | `Features/SubjectFields/SubjectFieldsSnapshot.swift` (`SubjectFieldsTypeChrome.stripIconKind`) | Map type keys → curated `subject_*` keys (incl. bridge collapse rules if still product intent). |
| Kind → ink / style (pigment only) | Snowflake (debt) | Ship | `Features/EvidenceGraph/EvidenceSubjectKindStyle.swift` (+ registry presentation) | **Do not** move pigment into mark assets. Full registry chrome migration remains **S7-09**. |

### 7.3 Explicit non-goals for this inventory

| Do not add | Why |
| --- | --- |
| New `DesignSystem/Components/*` for “subject mark” or “tool toggle” | Marks are a Recipe; palette toggles stay Evidence-graph snowflakes. |
| Second pack under `Recipes/SubjectIcon/` or keeping `EvidenceIcon/` | One home: **`Recipes/Marks/`**. |
| Card / Add-property / cited-row chrome | **S7-D3** / **S7-09**. |
| Subject fields layout / IA | **S7-D2** / shipped **S7-05** — icon swap only. |
| Source-type picker / `type_*` redesign | Already ships; only document family coexistence under Marks. |
| Lucide / SF Symbol substitution for subject kinds | Curated marks stay. |

### 7.4 Suggested implement order (S7-12)

1. Create `Recipes/Marks/` — move EvidenceIcon + assets; add `subject_*` (incl. **source** from Subject fields board).  
2. Update existing evidence-icon call sites to Marks imports.  
3. **Extend** graph call sites (card, bridge, palette, view maps).  
4. **Extend** Subject fields strip / `stripIconKind`.  
5. **Retire** `Recipes/EvidenceIcon` + `Recipes/SubjectIcon`; update DesignSystem README + `MARKS.md`.  
6. Grep-clean old type/path names.

---

## 8. Handoff

When content is agreed (board optional):

1. Archive this brief under `archive/`.
2. Record in [`../completed.md`](../../completed.md).
3. Implement **S7-12** against this contract (especially §7 — **`Recipes/Marks/`**) before **S7-09** thickens graph chrome.
