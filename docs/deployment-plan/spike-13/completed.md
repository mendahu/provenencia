# Spike 13 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S13-11 — Provenance sheet

**Landed 2026-10-31** in [#650](https://github.com/mendahu/provenencia/pull/650). Every Vitals row with records ends with a provenance control that opens a sheet: the shown value and its state, a summary, and each record's citation line (Source, Artifact, Citation, page), outcome and refs, with links to its Source, Citation and Observation. Records are grouped under their owning Event or Place (the shown one marked and spoken "shown value") and, for a mixed value, under the value they count under, with the rest last. The sheet follows its row across a refresh and closes when the row goes. The Person page no longer draws the Why; the Event and Place pages keep it, on the same `ReconciliationRecordRows`. `GetConclusionDetail`'s outcomes gain the Artifact label, Citation label, page and `group_rank` (Go groups the records; `locator.Page` reads the page). The sheet lists every kept value of a multi-valued row in its header, its summary reads "from 1 Source" for a single-Source value, a page above the 32-bit range shows no page, and each record is one VoiceOver element with its links as separate controls.

## S13-25 — The pin table goes

**Landed 2026-10-30** in [#644](https://github.com/mendahu/provenencia/pull/644). An Identity Claim is membership only: subject, entity, status. One format step (`000058.sql`) drops the Observation pin table, the claim's confidence grade and argument columns, and the `claim_confidence_grades` table, keeping every claim and every old audit row. The pin delete facet and its confirm line, `identityclaims` pins, the effects entries, the grades package and its create-time seed, and the Mac app's "removed from the evidence for" sentence are deleted. The data-model, delete and vocabulary docs say the table and columns were dropped.

## S13-06 — Derived rows carry their records

**Landed 2026-10-30** in [#640](https://github.com/mendahu/provenencia/pull/640). The Person page's birth date, birth place, death date and death place rows now show a Why like every other Vitals row. `GetConclusionDetail` on a Person returns `derived_rows`: for each life row, the reconciler records of the field it reads, keyed by the Event or Place that owns it (every surviving birth or death event's date, lead first; each of those events' Places' names once). With several owners the Why groups the records under a line titled as the Events and Places lists title them, the shown owner's in primary ink and the rest secondary; the place row's open arrow now opens the shown Place. `fieldsOf` composes any number of handles' fields in constant queries, and `ForEntity` is its one-handle caller.

## S13-24 — The pin engine goes

**Landed 2026-10-30** in [#639](https://github.com/mendahu/provenencia/pull/639). Done files accepted claims and their connections and nothing else: the claim evidence request, the evidence query, `core/claimevidence`, Apply's pin resolution and claim-field inputs, the confidence grade request and the registry's `evidence` section are deleted, and their wire is reserved. The reconciler no longer weakens or ranks a value by its claim's confidence grade (cache version 21), and one function now reads a record's provenance for the reconciler and the Person page. Go says where each Promote row starts (`graphalign.Row.Start`), so the proposal drops its score, band, Contested flag and reason and the app's draft rule shrinks to reading the target.

## S13-10 — Header and sections

**Landed 2026-10-29** in [#634](https://github.com/mendahu/provenencia/pull/634). The Person page keeps its header and gains a Vitals / Family / Timeline / Gallery / Sources switcher under it. Each section is a history place (`WorkspaceLocation.pageSection`, kind-neutral, nil is the first), so Back, Forward and a reopened project return to it, and the crumbs and Back menu name it. Vitals holds the rows with a reserved slot for S13-11's provenance control, Gallery holds the likenesses, and the other three are empty states until their PRs. The header pins and compacts on scroll. `ConclusionDetailPage` now sits on the kit's new `PVPageColumns` (centered at Wide, shared by Event and Place), and the date-kind, Observation-kind and value-type pickers moved onto the new `PVSegmentedControl`; `PVChip` gained a size, a glyph, a focus ring and the selected trait. "Vitals" now names only the section: the header's b. and d. lines are life lines.

## S13-23 — Promote stops asking for pins

**Landed 2026-10-29** in [#633](https://github.com/mendahu/provenencia/pull/633). Promote is a quick action: a researcher picks a target per row and presses Done. The evidence sheet, the assessment column (band, Contested mark, reason, chevron) and the confidence grade read are gone; the rows, headings and Already promoted rows share one grid (`PromoteGridRow`), and the list widens to 1600 pt with a 460 pt target menu at Wide. Done sends a target per row and nothing else, and the app carries no claim confidence or argument. The Updated caption reads the move reason, else the via neighbor. The share wording, five wire enums and their parity rows go, and every Promote string says promote, never file.

## S13-09 — Sections slot on the Conclusion page

**Landed 2026-10-29** in [#624](https://github.com/mendahu/provenencia/pull/624). `ConclusionDetailPage` draws the shared header, then fills everything below it from a `pageSections` slot. A convenience initializer supplies `ConclusionDetailDefaultPageSections` (Details rows, likenesses, Place relationship rows), so Person, Event and Place look the same and S13-10 can pass a custom Person fill. `docs/macos-client-patterns.md` names the slot apart from the Place hierarchy `sections`.
