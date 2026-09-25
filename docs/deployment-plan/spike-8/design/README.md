# Spike 8 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Dogfood background: [`docs/dogfood/ux.md`](../../../dogfood/ux.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S8-D2 | [`S8-D2-pdf-text-find.md`](S8-D2-pdf-text-find.md) | PRs **S8-03**, **S8-04**, **S8-05** | PDF Find on the tool strip; I-beam + scroll-to-pan; paste transcription from selection |
| S8-D3 | [`S8-D3-graph-visuals.md`](S8-D3-graph-visuals.md) | PR **S8-06** | Conflict + negated; Source-page jump; richer bridge sentences; Add property on bridges |
| S8-D4 | [`S8-D4-source-page.md`](S8-D4-source-page.md) | PR **S8-07** | Source page enhancements; first item: jump to Evidence graph |
| S8-D5 | [`S8-D5-sources-list.md`](S8-D5-sources-list.md) | PR **S8-08** | Sources list refresh; first item: subject + observation counts |
| S8-D6 | [`S8-D6-delete-paths.md`](S8-D6-delete-paths.md) | PR **S8-09** | Delete matrix; refine allow/refuse before the PR; no Change type |

## Agreed

| Step | Brief | Shipped as | Pick |
| --- | --- | --- | --- |
| S8-D7 | [`archive/S8-D7-composer-rethink.md`](archive/S8-D7-composer-rethink.md) | **S8-10** | **Layout A** (viewer \| 520pt stacked form). At window ≥ **1500pt**, form widens to 760–880pt and splits **reading \| observation stack**. **C** (subject headings) and **D** (four columns) rejected. |
| S8-D8 | [`archive/S8-D8-composer-connect-simplification.md`](archive/S8-D8-composer-connect-simplification.md) | **S8-11** | Row-level Save / Revert / Delete + **Save citation**; unsaved-work guard; one-row connection (endpoints fixed); location connections have no term; computed bridge names; New person / event / place from a row. Copy is citation wording, not "reading". |
| S8-D1 | [`archive/S8-D1-auto-transcribe.md`](archive/S8-D1-auto-transcribe.md) | **S8-01** | Trailing secondary **Auto transcribe** on the transcription label row; combined replace + whole-page confirm; field locked while running; image Artifacts only. |

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. **Start with S8-D2** (PDF Find / select / paste on the S8-11 composer). Paste **one** open brief (it already contains the [working-rules](claude-design-working-rules.md) ritual + kit reach-for table). New briefs: [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md).
3. Claude **clears cache, drops the stale pack, pulls a fresh design system**, and **works in place**. Rethink briefs replace frames; enhancement briefs extend them.
4. **Inventory is binding.** Instance kit components named in the brief; bespoke only when the use is truly domain-specific.
5. When the board is done, archive the brief under `archive/` and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project. Deployment target **macOS 14**.
- **Citation composer** is a navigable place (Option B): viewer \| form. **S8-10** shipped Layout A + the 1500pt two-column form (Citation is the document; subjects are per row). **S8-D8** (agreed) changes how it saves (row-level Save / Revert / Delete, Save citation, Done + unsaved guard) and moves Connect's role choice into it. Later briefs **extend the S8-D8 frames** — do not restyle the Spike 7 subject-locked shell or draw a footer Save. Locators stay unless a brief says so.
- **Evidence graph cards** already list every Observation as its own row. Competing values on one Property are legal. Negative polarity is a denial, not a missing line. **S8-D3** adds notice badges, a Source-page jump, and richer bridge sentences — not a resolve flow. Incomplete-bridge chrome is descoped.
- **Source page** is filing (identity, metadata, artifacts, notes). The Sources list already opens the Evidence graph; **S8-D4** adds that jump on the page. No Artifact → graph stays blocked.
- **Sources list** is the split-row file | graph destination. **S8-D5** adds graph-progress counts (subjects + observations) without folding them into the `sourcesList` payload.
- **Delete** is allowed for uncited subjects today; cited delete fails in the engine. **S8-D6** designs the full matrix. Wrong type is delete + place — no Change type.
- **Transcription** is the Citation reading. Observations are separate. Auto Transcribe must not write Observations.
- Media in composer: **image** and **PDF** viewers. **Auto Transcribe (S8-D1) is image-only.** PDF Find / select / paste is **S8-D2** (text layer, no Vision). Audio/video: honest disable ([`audio-video-sources.md`](../../../ideas/audio-video-sources.md)).
- Locator: default `artifact`, optional `page` (PDF), optional `region` polygon. One region max. Geometry: [`ArtifactRegionGeometry`](../../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift).
- Extend existing kit (`PVField`, `PVButton`, `PVTextArea`, `PVCallout`, `.pvConfirm`). Do not invent a second dialog stack.
