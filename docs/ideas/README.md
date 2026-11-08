# Ideas parking lot

Scratch space for product ideas that are **not** on the current deployment plan and should not be treated as commitments.

Use this folder when something useful pops up mid-work and would otherwise get lost. Capture enough context that a future reader understands the problem and the rough shape of a solution. Do **not** turn entries into spike tasks, schemas, or UI specs until they are deliberately pulled into a roadmap spike.

## Rules of thumb

- **Not authoritative.** Domain and stack decisions stay in the sibling docs under [`docs/`](../). Spikes live under [`deployment-plan/`](../deployment-plan/).
- **Not scheduled.** Adding a file here does not put it on the current spike (or any spike).
- **Prefer one file per domain cluster.** Keep related leftovers in one file with sections; link out to model docs when the idea depends on an existing layer.
- **Rough is fine.** Bullets, open questions, and “maybe later” notes are enough.
- **Promote out.** When an idea is pulled into a deployment-plan spike, move the file into that spike folder (or cut the section that shipped).
- **Delete when done.** A shipped or abandoned idea does not stay here. Git and the spike / PR that shipped it are the record. Open leftovers move into another current idea, a spike note, a domain doc, or a skill — never an `ideas/archive/` folder.
- **Dogfood first.** Fresh UI friction from using the app on real data starts as one file under [`docs/dogfood/ux/`](../dogfood/ux/). Move it here when it is parked longer-term; delete the dogfood file when this (or a spike) is the copy.

## Current ideas

### Evidence graph and canvas

- [`evidence-graph-followups.md`](evidence-graph-followups.md) — card image memory, upward shadows, P3 color drift, keyboard multi-select
- [`evidence-graph-structure-shortcuts.md`](evidence-graph-structure-shortcuts.md) — place or paste a common cluster; field hints from transcription
- [`graph-canvas-second-host.md`](graph-canvas-second-host.md) — what a family-tree canvas host still needs from GraphCanvas
- [`artifact-outside-graph-marker.md`](artifact-outside-graph-marker.md) — name, in a graph delete's confirm, the Artifacts it leaves in no Evidence graph; the card marker went to Spike 12 (S12-38)

### Citation composer and media

- [`citation-composer-followups.md`](citation-composer-followups.md) — bridge-mode subject filter; cropped citation preview
- [`depictions-and-likenesses.md`](depictions-and-likenesses.md) — Spike 10 / 14 leftovers (signature, page-change confirm, PDF Artifact thumbnails, audit items)
- [`audio-video-sources.md`](audio-video-sources.md)
- [`text-quote-locators.md`](text-quote-locators.md)

### Promote, matching, names

- [`promote-ranking.md`](promote-ranking.md) — rebuild Promote's ranking on Evidence graph scoring; name index on commit, narrower canon walk, coverage-aware similarity
- [`provisional-identity-claims.md`](provisional-identity-claims.md) — Promote Accepted / Provisional status; possible values from provisional members on conclusion pages
- [`matching-test-suite.md`](matching-test-suite.md) — scenario goldens, invariants, simulated world, scoreboard, labeler; remember researcher decisions
- [`international-names.md`](international-names.md) — other name patterns, transliteration, calendars, impossible dates, Daitch–Mokotoff
- [`kinship-view.md`](kinship-view.md) — family trees reading every kinship shape as one fact, and patterns beyond spouse, parent and grandparent
- [`organizations.md`](organizations.md) — Organization as a Subject type; descoped from S11-48
- [`place-gazetteer-service.md`](place-gazetteer-service.md)

### Conclusion pages and provenance

- [`provenance.md`](provenance.md) — cap records drawn in a provenance table; load records when the sheet opens
- [`person-page-performance.md`](person-page-performance.md) — record links memo and a household-sized budget, Timeline participants sent once, section content memoised, lazy Timeline cards, siblings sent once
- [`person-page-followups.md`](person-page-followups.md) — folded record line on hover, Person page wording, a birth and death events plural, Back keeping a section's place, Timeline scale after a date edit
- [`wide-layout.md`](wide-layout.md) — pages still waiting for a Wide / Page columns frame

### Properties, vocabulary, Sources

- [`properties-and-terms.md`](properties-and-terms.md) — binding order, event-type-scoped properties, Occupation seed, term management (direction / inverses)
- [`source-to-source-relationships.md`](source-to-source-relationships.md)
- [`registry-source-import.md`](registry-source-import.md) — paste a record URL from Ancestry, FamilySearch and similar; access unproven
- [`share-packages.md`](share-packages.md) — also holds the crafted-project hardening parked when Spike 12 closed

### Kit and structural debt

- [`shared-list-helpers.md`](shared-list-helpers.md) — empty search state, catalog querier, popup dismissal / list, history menu rows, type-select on disabled menus, Source section counts
- [`structural-debt.md`](structural-debt.md) — audit refactors with no visible change; leftover test artifact and Subject fixtures

Closed spikes (decisions, not PR order): [`docs/deployment-plan/archive/`](../deployment-plan/archive/).

Pulled into [Spike 15](../deployment-plan/spike-15/): nested Sources / derivation / templates ([`source-structure-later.md`](../deployment-plan/spike-15/source-structure-later.md)); Source privacy toggle; credibility field UX; rethink of Source metadata and type suggestions.
