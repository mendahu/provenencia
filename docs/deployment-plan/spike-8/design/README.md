# Spike 8 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Dogfood background: [`docs/dogfood/ux.md`](../../../dogfood/ux.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S8-D1 | [`S8-D1-auto-transcribe.md`](S8-D1-auto-transcribe.md) | PR **S8-01** | Auto Transcribe on citation transcription; warn + proceed on large page |

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the shipped visual language. Prefer existing `PV*` components; no new design-system primitives unless raised as a finding.
4. Every open brief must include a **UI building-block inventory** (components / recipes / snowflakes per [`docs/design-system-layers.md`](../../../design-system-layers.md)).
5. When the board is done, archive the brief under `archive/` and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project. Deployment target **macOS 14**.
- **Citation composer** is a navigable place (Option B): viewer \| form. Shipped in Spike 7. Do **not** redesign the shell, locators, or Observation list unless a brief says so.
- **Transcription** is the Citation reading. Observations are separate. Auto Transcribe must not write Observations.
- Media in composer: **image** and **PDF** viewers. **Auto Transcribe (S8-D1) is image-only.** PDF: paste (later text-layer). Audio/video: honest disable.
- Locator: default `artifact`, optional `page` (PDF), optional `region` polygon. One region max. Geometry: [`ArtifactRegionGeometry`](../../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift).
- Extend existing kit (`PVField`, `PVButton`, `PVTextArea`, `PVCallout`, `.pvConfirm`). Do not invent a second dialog stack.
