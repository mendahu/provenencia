# Spike 9 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**. Workflow: [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md).

Design track and gating: [`../deployment-plan.md`](../deployment-plan.md#design-track). **One brief per view.** Spike overview: [`../README.md`](../README.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S9-D1 | [Workspace sidebar](S9-D1-sidebar.md) | **S9-18** | Conclusions group + counts |
| S9-D2 | [Persons list](S9-D2-persons-list.md) | **S9-19** | Row anatomy the other lists extend |
| S9-D3 | [Events list](S9-D3-events-list.md) | **S9-20** | Extends D2 |
| S9-D4 | [Places list](S9-D4-places-list.md) | **S9-21** | Extends D2 |
| S9-D5 | [Person detail](S9-D5-person-detail.md) | **S9-22** | Value states + clusters the other details extend |
| S9-D6 | [Event detail](S9-D6-event-detail.md) | **S9-23** | Extends D5 |
| S9-D7 | [Place detail](S9-D7-place-detail.md) | **S9-24** | Extends D5 |
| S9-D8 | [Evidence graph subject card](S9-D8-graph-subject-card.md) | **S9-25** | Promote control + membership badge (enhancement) |
| S9-D9 | [Promote — choose target](S9-D9-promote-target.md) | **S9-26** | Decides place vs sheet; defines the Promote shell |
| S9-D10 | [Promote — claim fields](S9-D10-promote-claim-fields.md) | **S9-27** | Status dropdown laid out for three |
| S9-D11 | [Promote — compare](S9-D11-promote-compare.md) | **S9-28** | Existing handle only |
| S9-D12 | [Promote — walk](S9-D12-promote-walk.md) | **S9-29** | Connected subjects, bridge confirm, Done |
| S9-D13 | [Omnibar results](S9-D13-omnibar-hits.md) | **S9-30** | Hit rows for three kinds (enhancement) |

Design each brief alongside its feature, just before the PR it gates. Order and dependencies: [PR sequence](../deployment-plan.md#pr-sequence).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project. Deployment target **macOS 14**.
- In the UI a canonical `person` row is a **Person** (`PER-7KD45`), not a "canonical entity." Interpretation Subjects keep candidate refs (`CPR-…`).
- A handle's **members** are the Subjects with an accepted Identity Claim for it. Every Property is **multi-valued** across members. Nothing on these pages is a committed value (no Reconciliation Claims this spike).
- Each field shows one resolved value in a state: **single**, **auto-reconciled** (names, dates), **mixed** (top-ranked by provenance, with an indicator), or empty. Details can list every ranked candidate; list rows show one value plus a *mixed* marker or *+N*. A future **concluded** state (Reconciliation Claim) needs room too.
- Promote only **creates** claims, one subject per saved step, with a Done off-ramp after each. Minting a new handle skips the comparison, so the first claim usually has no pins. Editing or removing claims is a separate workflow (Spike 10).
- Promote's claim fields include a **Status dropdown** with one option (`accepted`) this spike. Lay it out for `provisional` / `rejected` too; they arrive later.
- Promote starts from a subject card on the Evidence graph. Do **not** redesign the graph, cards, or composer except for the Promote entry point and membership chrome.
- No likeness value type exists. Thumbnails are a slot with a per-kind placeholder.
- Extend the existing kit (`PVSidebarNav`, `PVField`, `PVButton`, `PVCallout`, `.pvConfirm`). Layers: [`docs/design-system-layers.md`](../../../design-system-layers.md).
