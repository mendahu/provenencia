# Spike 9 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**. Workflow: [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md).

Design track and gating: [`../deployment-plan.md`](../deployment-plan.md#design-track). **One brief per view.** Spike overview: [`../README.md`](../README.md).

## Open

| Step | Brief | Feeds (later PRs on the view) | Notes |
| --- | --- | --- | --- |
| S9-D15 | [Custom term — category](S9-D15-custom-term-category.md) | **S9-38b** | Enhancement to the composer's custom term dialog |
| S9-D16 | [Promote — one page](S9-D16-promote-page.md) | **S9-44** | **Rethink:** throws out the D9 / D10 frames on the Promote board; one page aligning the whole Evidence graph |
| S9-D13 | [Omnibar results](S9-D13-omnibar-hits.md) | **S9-35** | Hit rows for three kinds (enhancement) |

## Done

| Step | Brief | Shipped in | Notes |
| --- | --- | --- | --- |
| S9-D1 | [Workspace sidebar](archive/S9-D1-sidebar.md) | **S9-08** (S9-23 / S9-26 go live) | Titled Source / Conclude / Configure groups; Configure bottom-aligned behind space + hairline, no counts |
| S9-D2 | [Persons list](archive/S9-D2-persons-list.md) | **S9-09** (S9-32) | Kit `PVList` + `ConclusionListRow`; title name → italic label → mono ref; trailing ref; no mixed marker in rows |
| S9-D3 | [Events list](archive/S9-D3-events-list.md) | **S9-23** (S9-32) | Same row; date on the secondary line; subject titles and places later |
| S9-D4 | [Places list](archive/S9-D4-places-list.md) | **S9-26** (S9-40) | Same row; first name and +N; chain cell empty until S9-40 |
| S9-D8 | [Evidence graph subject card](archive/S9-D8-graph-subject-card.md) | **S9-04** (S9-09 name, S9-11 flow) | One 36pt kind-chip footer (rev 1): Promote, then the membership link |
| S9-D9 | [Promote — choose target](archive/S9-D9-promote-target.md) | **S9-11** (replaced by S9-D16 / S9-44) | A workspace place, not a sheet; subject header in the kind's wash, composed step row, New / Existing radios, search + Suggested radio rows, Done + leave guard |
| S9-D10 | [Promote — claim fields](archive/S9-D10-promote-claim-fields.md) | **S9-12** (replaced by S9-D16 / S9-44) | Summary card of the write, Status Select (one option) beside Confidence, Argument; Done discards through the guard; footer Back keeps the draft |
| S9-D5 | [Person detail](archive/S9-D5-person-detail.md) | **S9-16** (S9-32 fills the life rows) | Field rows of label · value · state badge · Sources · disclosures; Why as Read as · Source · Outcome with a mark and a phrase per outcome; member list slot deferred |
| S9-D6 | [Event detail](archive/S9-D6-event-detail.md) | **S9-24** (S9-32) | Same page as Person detail; Date row, empty Place; subject titles and places later |
| S9-D7 | [Place detail](archive/S9-D7-place-detail.md) | **S9-27** (S9-40) | Same page; every kept name and its Why; period and relationship sections empty until S9-40 |
| S9-D14 | [Properties — cardinality](archive/S9-D14-properties-cardinality.md) | **S9-37** | Holds on the inspector and the create form; seeded Properties read-only |

## Superseded

| Step | Brief | Why |
| --- | --- | --- |
| S9-D11 | [Promote — compare](archive/S9-D11-promote-compare.md) | Never built. Replaced by S9-D16's evidence sheet (2026-10-06). |
| S9-D12 | [Promote — walk](archive/S9-D12-promote-walk.md) | Never built. Replaced by S9-D16's whole-graph page (2026-10-06). |

Design each brief alongside its feature, just before the PR it gates. Order and dependencies: [PR sequence](../deployment-plan.md#pr-sequence).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project. Deployment target **macOS 14**.
- In the UI a canonical `person` row is a **Person** (`PER-7KD45`), not a "canonical entity." Interpretation Subjects keep candidate refs (`CPR-…`).
- A handle's **members** are the Subjects with an accepted Identity Claim for it. Each Property collects values from every member and the engine **reconciles** them ([`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md)). Nothing on these pages is a committed value (no Reconciliation Claims this spike).
- Each field shows its reconciled value in a state: **single**, **merged**, **mixed** (every surviving value shown), or empty. Details explain every value: each record considered and its outcome (kept, folded, outvoted, weak, denied, no usable value); support counts Sources. List rows show one value plus *+N*, never a mixed marker. A few Properties hold several true values (a Place's names). A future **concluded** state (Reconciliation Claim) needs room too.
- **Places form a hierarchy.** Each Place may have a **nature** (administrative, informal, ecclesiastical). Relationships between places are *part of* or *succeeded by* (York → Toronto). Places have periods, and a part-of link has its own span. An undated link holds for the overlap of the places' periods; a dated link can end while both places continue. A place's parents depend on the date; display prefers administrative parents.
- Promote only **creates** claims. It is **one page** that proposes a handle, New or Skip for every person, event and place on an Evidence graph, with the researcher fine-tuning and one **Done** filing everything; bridges file automatically ([`promote-alignment.md`](../../../promote-alignment.md), brief S9-D16). Editing or removing claims is a separate workflow (Spike 10).
- Promote's claim fields include a **Status dropdown** with one option (`accepted`) this spike. Lay it out for `provisional` / `rejected` too; they arrive later.
- Promote starts from a subject card on the Evidence graph. Do **not** redesign the graph, cards, or composer except for the Promote entry point and membership chrome.
- No likeness value type exists. Thumbnails are a slot with a per-kind placeholder.
- Extend the existing kit (`PVSidebarNav`, `PVField`, `PVButton`, `PVCallout`, `.pvConfirm`). Layers: [`docs/design-system-layers.md`](../../../design-system-layers.md).
