# Ideas parking lot

Scratch space for product ideas that are **not** on the current deployment plan and should not be treated as commitments.

Use this folder when something useful pops up mid-work and would otherwise get lost. Capture enough context that a future reader understands the problem and the rough shape of a solution. Do **not** turn entries into spike tasks, schemas, or UI specs until they are deliberately pulled into a roadmap spike.

## Rules of thumb

- **Not authoritative.** Domain and stack decisions stay in the sibling docs under [`docs/`](../). Spikes live under [`deployment-plan/`](../deployment-plan/).
- **Not scheduled.** Adding a file here does not put it on the current spike (or any spike).
- **Prefer one file per idea.** Keep the entry short; link out to model docs when the idea depends on an existing layer.
- **Rough is fine.** Bullets, open questions, and “maybe later” notes are enough.
- **Promote out.** When an idea is pulled into a deployment-plan spike, move the file into that spike folder.
- **Archive when shipped.** Move finished ideas into [`archive/`](archive/) and distill with [`archive-docs`](../../.agents/skills/archive-docs/SKILL.md).
- **Dogfood nits stay out.** UI friction from using the app on real data goes in [`docs/dogfood/ux.md`](../dogfood/ux.md), not here.

## Current ideas

- [`source-to-source-relationships.md`](source-to-source-relationships.md)
- [`source-structure-later.md`](source-structure-later.md) — what Spike 12 parked: nested Sources, derivation links, named templates, per-graph credibility, layered citations
- [`graph-focus-views.md`](graph-focus-views.md) — focus views inside one large Evidence graph
- [`evidence-graph-structure-shortcuts.md`](evidence-graph-structure-shortcuts.md) — place a common cluster (wedding, birth, residence) in one action, or paste a cluster already on the canvas
- [`registry-source-import.md`](registry-source-import.md) — paste a record URL from Ancestry, FamilySearch and similar to create a Source; access (logins, OAuth) is unproven
- [`source-provenance-date.md`](source-provenance-date.md)
- [`text-quote-locators.md`](text-quote-locators.md)
- [`audio-video-sources.md`](audio-video-sources.md)
- [`shared-list-helpers.md`](shared-list-helpers.md) — one list of graph mutations (the other two items are in Spike 11)
- [`test-subject-fixtures.md`](test-subject-fixtures.md) — test Subject helpers that place a Subject do it in a second write; let `subjectstest` take a placement
- [`settled-list-count-publish.md`](settled-list-count-publish.md) — Properties badge is off by one after create or delete; one publish-on-settled rule for Sources, Properties and Metadata
- [`depictions-and-likenesses.md`](depictions-and-likenesses.md) — leftovers after Spike 10 (other keys, signature, transparent crop, Windows PDF draw)
- [`coverage-aware-similarity.md`](coverage-aware-similarity.md) — a match on one shared property (a surname-only record) scores similarity 1 and shows as an unbanded alternative
- [`international-names.md`](international-names.md) — leftovers after Spike 10 (other name patterns, transliteration, calendars)
- [`daitch-mokotoff.md`](daitch-mokotoff.md) — phonetic tier descoped from Spike 10 (S10-26)
- [`matching-test-suite.md`](matching-test-suite.md) — scenario goldens, invariants, a simulated world, a scoreboard, and a labeler that fits the registry numbers. Its "make the algorithm graspable" part is Spike 11 (S11-01); the rest waits on the re-implementation
- [`share-packages.md`](share-packages.md)
- [`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md)
- [`large-jpeg-thumbnails.md`](large-jpeg-thumbnails.md) — JPEGs over the 64-megapixel decode budget save and keep the missing-image glyph
- [`identity-claim-review.md`](identity-claim-review.md)
- [`binding-order-editing.md`](binding-order-editing.md) — a per-type ordered list of properties on the Properties page, calling `ReorderSubjectTypeProperties` (engine side shipped in S11-32)
- [`place-gazetteer-service.md`](place-gazetteer-service.md)
- [`possible-values.md`](possible-values.md)
- [`term-management.md`](term-management.md) — one place to manage terms, including relationship direction and inverses
- [`kinship-view.md`](kinship-view.md) — family trees and person pages reading every kinship shape as one fact, and patterns beyond spouse, parent and grandparent (matching's part is Spike 11)

## Archived

Shipped ideas keep **decisions**, not PR order. Closed spikes: [`docs/deployment-plan/archive/`](../deployment-plan/archive/).

- [`archive/interpretation-graph-ui.md`](archive/interpretation-graph-ui.md) — Evidence graph decisions; leftovers re-homed
- [`archive/evidence-graph-drag-performance.md`](archive/evidence-graph-drag-performance.md) — AppKit cards; a drag moves the view
- [`archive/page-navigation-performance.md`](archive/page-navigation-performance.md) — session cache / place registry
- [`archive/catalog-access-serialization.md`](archive/catalog-access-serialization.md) — held catalog session
- [`archive/catalog-graph-cache.md`](archive/catalog-graph-cache.md) — one write path and the in-memory catalog graph
- [`archive/design-system-hardening.md`](archive/design-system-hardening.md) — graduated to [`docs/design-system-layers.md`](../design-system-layers.md)
- [`archive/promote-matching.md`](archive/promote-matching.md) — what proves a join; answered by graph alignment
- [`archive/name-matching-enhancements.md`](archive/name-matching-enhancements.md) — Spike 10 matcher leftovers, each closed or re-homed
- Also archived: nav-count RPC, ingest MIME, thumbnails, image cache (same folder)
