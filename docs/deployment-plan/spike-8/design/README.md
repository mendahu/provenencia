# Spike 8 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Dogfood background: [`docs/dogfood/ux.md`](../../../dogfood/ux.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S8-D1 | [`S8-D1-auto-transcribe.md`](S8-D1-auto-transcribe.md) | PR **S8-01** | Auto Transcribe on citation transcription; warn + proceed on large page |
| S8-D2 | [`S8-D2-pdf-text-find.md`](S8-D2-pdf-text-find.md) | PRs **S8-03**, **S8-04**, **S8-05** | PDF Find on the tool strip; I-beam vs pan; paste transcription from selection |
| S8-D3 | [`S8-D3-graph-visuals.md`](S8-D3-graph-visuals.md) | PR **S8-06** | Conflict + negated; Source-page jump; richer bridge sentences; Add property on bridges |
| S8-D4 | [`S8-D4-source-page.md`](S8-D4-source-page.md) | PR **S8-07** | Source page enhancements; first item: jump to Evidence graph |
| S8-D5 | [`S8-D5-sources-list.md`](S8-D5-sources-list.md) | PR **S8-08** | Sources list refresh; first item: subject + observation counts |

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the shipped visual language. Prefer existing `PV*` components; no new design-system primitives unless raised as a finding.
4. Every open brief must include a **UI building-block inventory** (components / recipes / snowflakes per [`docs/design-system-layers.md`](../../../design-system-layers.md)).
5. When the board is done, archive the brief under `archive/` and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project. Deployment target **macOS 14**.
- **Citation composer** is a navigable place (Option B): viewer \| form. Shipped in Spike 7. Do **not** redesign the shell, locators, or Observation list unless a brief says so.
- **Evidence graph cards** already list every Observation as its own row. Competing values on one Property are legal. Negative polarity is a denial, not a missing line. **S8-D3** adds notice badges, a Source-page jump, and richer bridge sentences — not a resolve flow. Incomplete-bridge chrome is descoped.
- **Source page** is filing (identity, metadata, artifacts, notes). The Sources list already opens the Evidence graph; **S8-D4** adds that jump on the page. No Artifact → graph stays blocked.
- **Sources list** is the split-row file | graph destination. **S8-D5** adds graph-progress counts (subjects + observations) without folding them into the `sourcesList` payload.
- **Transcription** is the Citation reading. Observations are separate. Auto Transcribe must not write Observations.
- Media in composer: **image** and **PDF** viewers. **Auto Transcribe (S8-D1) is image-only.** PDF Find / select / paste is **S8-D2** (text layer, no Vision). Audio/video: honest disable ([`audio-video-sources.md`](../../../ideas/audio-video-sources.md)).
- Locator: default `artifact`, optional `page` (PDF), optional `region` polygon. One region max. Geometry: [`ArtifactRegionGeometry`](../../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift).
- Extend existing kit (`PVField`, `PVButton`, `PVTextArea`, `PVCallout`, `.pvConfirm`). Do not invent a second dialog stack.
