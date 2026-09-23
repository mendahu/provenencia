# S7-D7 — Card component (design-system reference)

**Kind:** Claude Design board / design-system component  
**Spike:** Provenencia Spike 7  
**Implements later as:** PR **S7-13**  
**Depends on:** Shipped macOS [`PVCard`](../../../../../macos/App/DesignSystem/Components/Card/PVCard.swift) (already in the app kit — this brief **documents and references** it in Claude Design, not a greenfield invent)  
**Related:** Evidence graph subject/bridge cards stay **snowflakes** (**S7-D3** / **S7-09**) — out of scope here  
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)

> **Shipped delta (S7-13):** Four planned cousins plus locked Connect observation rows and Source types empty-suggestions are `PVCard`. Web kit header/footer/`hoverable` slots were **not** ported — titles stay composed outside.

Paste this document into Claude Design as the requirements for a **Card** kit board on the **main** design system only. Child view boards cannot see this file — remount them with [`S7-D7B-card-view-remount.md`](S7-D7B-card-view-remount.md), one board at a time.

---

## 1. Objective

Publish a **Card** component in the Provenencia Claude Design system whose visual contract matches what macOS already ships as `PVCard`:

- Surface tones: card / raised / sunken (`PVColor.surfaceCard` / `surfaceRaised` / `surfaceSunken`)
- Border: solid hairline (`borderSubtle`) or dashed (`borderDefault`, provisional rows)
- Corner radius default `md` (overrideable `sm`)
- Optional shallow elevation (`PVElevation.sm`)
- Optional content padding
- Content-agnostic container — **no** header/footer slots required in this pass (call sites compose titles outside, same as today’s Swift)

Use **implemented** Artifacts / Subject fields / metadata call sites as visual reference — not a speculative redesign.

**After Card is on the main design system, this brief’s scope includes going into every view in §2, pointing those call sites at the shared Card, and deleting local card implementations.** Do not leave a parallel rectangle, fill, radius, or dashed stroke on those boards. Claude Design stale-bundle habit: **clear cache and rebundle the main system first**; each child board **deletes its local cache and refetches** before remounting.

---

## 2. Views that reference Card

These are the Claude Design **views** that must remount on Card once the kit page is ready. Evidence graph subject / bridge / palette chrome is **not** on this list — those stay S7-D3 snowflakes (CD-2).

| View | Board | Status | What to remount |
| --- | --- | --- | --- |
| **Source page** (S2-23) | [Source detail](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share) | Already `PVCard` | Artifacts list; expanded primary-file tile; dashed metadata suggestion row |
| **Subject fields** (S7-D2) | [Subject fields](https://claude.ai/design/p/6dceb4b9-d08a-40ad-a46f-430651ea9b3c?via=share) | Already `PVCard` | Property list card; inspector card. Type-strip tiles stay selectable chips, not Card |
| **Citation composer** (S7-D4) | [Citation composer](https://claude.ai/design/p/8b120853-4ba4-4c0d-8492-75e76b9f9b7a?via=share) | Already `PVCard` | Observation summary rows (raised). Locked Connect rows are still a sunken cousin — compose Card later if the board wants one treatment |
| **Source types** (S2-03) | [Source types](https://claude.ai/design/p/bff7a2e6-4ea1-4032-9800-eecd4ad81284?via=share) | **S7-13** migrate | Suggested-fields stack (small radius) |
| **Onboarding** (Identify + Choose file) | [Onboarding Flow](https://claude.ai/design/p/b51790c9-7f65-4e91-a7ef-a900095c5870?via=share) | **S7-13** migrate | Identify contributor list; Choose-file empty-folder dashed sunken block; selected-project meta card |

### Remount rule (in scope)

Once Card exists on the **main** Provenencia design system:

1. **Clear the main design-system cache and rebundle** before any view remount. Claude Design often keeps a stale bundle; remounting against the old pack will invent or keep local Cards.
2. On **each** §2 child board, **delete the local design-system cache and refetch** from the main system. Do not remount until that board’s bundle shows the new Card.
3. Open each board in the table.
4. Replace every local card-shaped treatment listed under **What to remount** with an instance of the kit Card (tone / border / radius / elevation / padding as that site already uses).
5. **Remove the local implementation** — no leftover custom fill, clip, hairline, or dashed overlay that duplicates Card.
6. Leave out-of-scope chrome alone (graph snowflakes, type-strip tiles, file-choice / icon-picker cells, pane fills).

If a child board still cannot see Card after refetch, treat that as a stale-bundle failure: clear/rebundle the main system again, then delete that board’s cache and refetch once more. Do not copy Card into the child project.

### Stale bundles (Claude Design)

Claude Design frequently serves an old design-system pack. After Card lands on the main system, **cache clear + rebundle on the main system is required**, and **every child view board must drop its local cache and refetch**. Skipping that step is why remounts fail or recreate a second Card.

### Call sites (Swift → view)

| View | File | Chrome today |
| --- | --- | --- |
| Source page | [`SourcePageArtifactsView`](../../../../../macos/App/Features/Sources/SourcePageArtifactsView.swift) | `PVCard(elevated: true)` — Artifacts list |
| Source page | same file, primary-file column | `PVCard(cornerRadius: .sm)` — File tile in accordion expand |
| Source page | [`SourcePageMetadataView`](../../../../../macos/App/Features/Sources/SourcePageMetadataView.swift) | `PVCard(border: .dashed, cornerRadius: .sm)` — type-suggested metadata row |
| Subject fields | [`SubjectFieldsView`](../../../../../macos/App/Features/SubjectFields/SubjectFieldsView.swift) `listCard` / `inspectorCard` | Default `PVCard` |
| Citation composer | [`CitationComposerObservationRow`](../../../../../macos/App/Features/CitationComposer/CitationComposerObservationRow.swift) | `PVCard(tone: .raised)` — editable Observation rows |
| Source types | [`SourceTypesDetailPane`](../../../../../macos/App/Features/SourceTypes/SourceTypesDetailPane.swift) (~suggested-fields stack) | Hand-built `surfaceCard` + `sm` + `borderSubtle` |
| Onboarding | [`OnboardingIdentifyView`](../../../../../macos/App/Features/Onboarding/OnboardingIdentifyView.swift) | Hand-built `surfaceCard` + `md` + `borderSubtle` — contributor list |
| Onboarding | [`OnboardingOpenPicker`](../../../../../macos/App/Features/Onboarding/OnboardingOpenPicker.swift) | Sunken + dashed — empty Documents folder |
| Onboarding | [`OnboardingProjectMetaLines`](../../../../../macos/App/Features/Onboarding/OnboardingProjectMetaLines.swift) | Hand-built `surfaceCard` + `md` — project bookkeeping under the open picker |

---

## 3. Implementation gate (S7-13)

| Ships in S7-13 | Does **not** ship there |
| --- | --- |
| Migrate the four manual “good candidates” to `PVCard` (see deployment plan) | Evidence graph subject/bridge/palette chrome (**S7-D3** snowflakes) |
| Align Swift `PVCard` with any board-clarified props if needed | Selectable strip tiles / file-choice / icon-picker cells |
| DesignSystem README already lists Card — keep in sync | Full-bleed pane fills (sidebar, list bands) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CD-1 | Board shows Card at default, elevated, dashed, sunken, and `sm` radius. |
| CD-2 | Document that Evidence graph cards are **not** this component (kind tint / selection halo). |
| CD-3 | No new kit primitives beyond Card; tokens already exist. |
| CD-4 | UI inventory: Card = Component **Ship** / light **Extend** if board adds documented slots; S7-13 call sites = Snowflake **Extend**. |
| CD-5 | After Card is on the main design system, update **every** §2 board: point remount sites at kit Card and **delete local card implementations**. Do not remount Evidence graph cards, type-strip tiles, file-choice / icon-picker cells, or pane fills. |
| CD-6 | After Card lands: clear cache and **rebundle** the main design system; on each §2 board delete the local cache and **refetch** from main. Do not remount against a stale bundle. |

---

## 5. Out of scope

- Graph card growth / cited rows (**S7-D3**)
- Selectable “chip card” API for type strips
- Omnibar overlay elevation (`PVElevation.overlay`) unless the board explicitly unifies it

---

## 6. Handoff

1. Publish Card on the main design system. Clear cache and rebundle that system (CD-6).
2. Paste [`S7-D7B`](S7-D7B-card-view-remount.md) into **one** §2 board at a time (child boards cannot read this file). Each board clears its cache, refetches, remounts Card, and deletes local card chrome (CD-5).
3. Archive this brief under `archive/` when the board is agreed.
4. Record in [`../../completed.md`](../../completed.md).
5. Implement **S7-13** against the board + existing `PVCard` API.
