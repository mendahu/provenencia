# Ideas parking lot

Scratch space for product ideas that are **not** on the current deployment plan and should not be treated as commitments.

Use this folder when something useful pops up mid-work and would otherwise get lost. Capture enough context that a future reader understands the problem and the rough shape of a solution. Do **not** turn entries into spike tasks, schemas, or UI specs until they are deliberately pulled into a roadmap spike.

## Rules of thumb

- **Not authoritative.** Domain and stack decisions stay in the sibling docs under [`docs/`](../). Spikes live under [`deployment-plan/`](../deployment-plan/).
- **Not scheduled.** Adding a file here does not put it on the current spike (or any spike).
- **Prefer one file per idea.** Keep the entry short; link out to model docs when the idea depends on an existing layer.
- **Rough is fine.** Bullets, open questions, and “maybe later” notes are enough.
- **Promote out.** When an idea is pulled into a deployment-plan spike, move the file into that spike folder.
- **Archive when shipped.** Move finished ideas into [`archive/`](archive/).
- **Dogfood nits stay out.** UI friction from using the app on real data goes in [`docs/dogfood/ux.md`](../dogfood/ux.md), not here.

## Current ideas

- [`source-to-source-relationships.md`](source-to-source-relationships.md)
- [`text-quote-locators.md`](text-quote-locators.md)
- [`audio-video-sources.md`](audio-video-sources.md)
- [`share-packages.md`](share-packages.md)

## Archived

Shipped ideas keep **decisions**, not PR order. Closed spikes: [`docs/deployment-plan/archive/`](../deployment-plan/archive/).

- [`archive/interpretation-graph-ui.md`](archive/interpretation-graph-ui.md) — Evidence graph decisions; leftovers re-homed
- [`archive/page-navigation-performance.md`](archive/page-navigation-performance.md) — session cache / place registry
- [`archive/catalog-access-serialization.md`](archive/catalog-access-serialization.md) — held catalog session
- [`archive/design-system-hardening.md`](archive/design-system-hardening.md) — graduated to [`docs/design-system-layers.md`](../design-system-layers.md)
- Also archived: nav-count RPC, ingest MIME, thumbnails, image cache (same folder)
