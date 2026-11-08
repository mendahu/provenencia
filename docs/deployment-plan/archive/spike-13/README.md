# Spike 13 — Person detail page

**Done** (closed 2026-11-08). The Person page is where a researcher understands one canonical person: a fixed header over Vitals, Family, Timeline, Gallery and Sources, a provenance sheet beside every Vitals value, the claims that make the person up, and how the person's records connect. Before it, Identity Claims lost their Observation pins and Promote lost its evidence sheet. The user checked the dogfood bar on their own people and accepted what shipped as built (2026-11-08). Live models: [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`matching-model.md`](../../../matching-model.md), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`design-system-layers.md`](../../../design-system-layers.md), [`catalog-deletes.md`](../../../catalog-deletes.md). Design reasoning for record links: [`evidence-graph-comparison.md`](evidence-graph-comparison.md).

## Decisions

- **An Identity Claim is membership only:** subject, entity, status. No Observation pins, no confidence grade, no argument; the tables and columns are dropped and audit history keeps its old rows. Promote is a quick action that files accepted claims and their connections, and its score ranks suggestions unseen.
- **One sectioned Conclusion page, not a Person fork.** `ConclusionDetailPage` on Page columns serves Person, Event and Place; the Person page passes five sections. A section is a history place (`WorkspaceLocation.pageSection`), so Back, Forward and restore return to it.
- **Each section reads its own query** (claims, family, timeline, record links, beside the detail read), served from the graph cache and refreshed by one per-person invalidation spec.
- **Canonical only, filtered in Go.** Nothing unpromoted appears on the page.
- **The provenance sheet is a kit recipe** (a record row, a record table, the sheet), filled by a pure assembler from the wire records, and it reports a link for the caller to map to a place. It replaces the Why on the Person page only; Event and Place keep the Why accordion, drawing the recipe's parts, until they adopt sections. The sheet leaves room for manual Reconciliation Claims ([conclusion §13](../../../conclusion-layer-data-model.md) item 5), so they slot in without a redesign.
- **A derived row carries its owner's records:** birth and death date and place show every surviving event's or Place's records, grouped by owner.
- **Family is grouped in Go** and shaped as a family tree's first step: units over a list of ids, each link saying what it rests on. Siblings from a shared parent are page-only; matching and Promote are unchanged.
- **The Timeline is direct participation, any role,** on a lifeline drawn to scale: birth and death always have a marker (a placeholder when unrecorded), undated events sit before death, and a date known only by its end ("bef. 1880") is dated and ordered at that end everywhere.
- **Display is decoupled from content.** The person card (`PVTileCard`, `PVPersonCard`) and gender treatment are design-system pieces; name, date, place and title formatting live in `Platform/`; a built-in term reads with the app's name everywhere.
- **One Unpromote flow** for the Evidence graph and the Person page. Removing a handle's last member deletes the handle in the same revision through `deleteimpact`, and a write's effects name the ids it deleted.
- **How a person's records connect is computed, never stored.** The first version counts what two Evidence graphs share about each other person they both name, weak or strong at two facts (the match registry's `RecordLinks.StrongAt`), drawn as a circular link diagram beside the Sources rows. It is not Promote's alignment score.

## Refused

Provisional claims at Promote ([`provisional-identity-claims.md`](../../../ideas/provisional-identity-claims.md)); a claim review sheet (replaced by record links); Subject cards in the Sources section (Source rows with a line per record instead); a compact header on scroll; stored, researcher-chosen links between records and a research log; unpromoted Subjects anywhere on the page.

## Leftovers

| Item | Home |
| --- | --- |
| The Source date and standard citations | Spike 15: [`source-provenance-date.md`](../../spike-15/source-provenance-date.md), [`citation-formats.md`](../../spike-15/citation-formats.md) |
| Record links after the first version: conflicts, reasons, a richer weight, Source quality and independence, events and places as lines | [conclusion §13](../../../conclusion-layer-data-model.md) item 4 |
| Promote ranking with the record-link scoring | [`promote-ranking.md`](../../../ideas/promote-ranking.md) |
| Kinship beyond the nuclear family; sibling evidence for matching | [`kinship-view.md`](../../../ideas/kinship-view.md) |
| Remember researcher decisions | [`matching-test-suite.md`](../../../ideas/matching-test-suite.md) |
| Ship-log and review follow-ups not already in an idea file or `structural-debt.md`: flaky macOS suites (`SourceGraphFormModelTests`, `SourcesModelTests`, `SourcePageModelTests`, `PVWidthClassTests`, `PVSegmentedControlTests`, the graph-progress `waitUntil` tests); gallery and likeness viewer focus and VoiceOver details; Family's "Show N more" focus; the Timeline card id namespace; `edgeKey`'s missing separator; the gallery's unbounded tile count; `DateRowDisplay` letting a date with no year win; a test seam for kit Return and focus; a wider `observationstest.Cite`; `productKey(of:)`, the summary-state rule and hand-rolled dividers as copies; the via term derived in Swift; the "A stronger match is %@" copy; and every logged nit | Dropped (user, 2026-11-08) |
