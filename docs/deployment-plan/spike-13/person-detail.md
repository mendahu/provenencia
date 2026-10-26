# Person detail page

**Status:** requirements agreed (2026-10-26). Dictated by the product owner, then refined through a round of scope questions. Not yet cut into PRs.

## Why

The Person page (`macos/App/Features/Conclusions/PersonDetailView.swift`, worded by `PersonDetailContent`, board S9-D5) is a stub: a header with name, born/died vitals and a likeness thumb, one column of reconciled property rows with a Why accordion (`ReconciliationReasoningView`), a member count, and a likeness gallery. Spike 13 makes it the place a researcher goes to understand one canonical person: who they were, who their family was, what happened in their life, what they looked like, and what evidence says so.

## Spike-wide rules

- **R-1. Person page only.** Event and Place pages keep their current layout this spike. The section pattern may be reused for them later.
- **R-2. Canonical people only.** Everything on the page is canonical (promoted) entities. A relative, participant, or anything else that exists only as an unpromoted Subject does not appear. Where nothing is promoted, nothing is drawn.
- **R-3. Reuse the existing gender/sex colors.** Every person treatment on the page (Family, Timeline participants) uses the color logic already shipped (`PersonGenderTreatment`, S10-23), including its treatment for unknown values.

## Design first, then re-cut

- **R-D1. The whole page is designed up front.** Before any chrome PR is planned, Claude Design boards cover every part of the page together: header, section switching, all five sections, the provenance sheet, the gallery viewer, and the claim review and remove-last-claim flows. The product owner approves the page as a whole, not view by view.
- **R-D2. Expect scope to change after the design.** The PR map is cut from these requirements, then re-cut once the boards are approved. Chrome PRs stay `Scoped` until then; requirements here may be amended by what the design finds.
- **R-D3. Engine work may run ahead of the design** where it does not depend on what the boards decide (for example sibling inference, R-18), so the spike is not idle while design iterates.

## Layout

- **R-4. A fixed header.** The current header (name, born/died line, portrait, gender badge, ref) stays at the top whichever section is showing. It is display only: clean and easy to read, with no controls.
- **R-5. Five sections.** Below the header the page is divided into **Vitals**, **Family**, **Timeline**, **Gallery**, **Sources**.
- **R-6. Easy switching.** The user can move between the five sections quickly from anywhere on the page.
- **R-6a. Sections are places.** Switching sections pushes navigation history; Back / Forward and restore return to the section (2026-10-26).
- **R-7. The pattern is a design call.** Stacked sections with jump shortcuts, tab navigation, or something else: the design brief explores the options and picks an elegant one.

## Provenance sheet (new control)

A new, reusable control that sits next to any concluded value on a page. It replaces and extends today's Why accordion.

- **R-8. One control, one sheet.** Clicking the control beside a value opens a sheet describing that value's entire provenance.
- **R-9. What the sheet shows.** Every Source that speaks to the value; the auto-reconciler's reasoning (everything the Why accordion shows today moves here, with more detail); and, later, any manual Reconciliation Claims (R-10).
- **R-10. Reconciliation Claims are out of scope.** They are not built yet. The sheet is designed so they slot in later without a redesign.
- **R-11. Rich links.** From the sheet the user can click through to the Source, the Citation, and the exact Observation behind each record.
- **R-12. Conflicting values live here.** Vitals keeps today's "mixed" marker on a row whose records disagree; the detail behind it is in the provenance sheet.
- **R-13. Reusable.** Built as a shared control, not Person-specific, so Event and Place pages can adopt it later (R-1).
- **R-13a. Wherever a single value is shown, in Vitals.** Every value in Vitals gets the control, including values that belong to another entity: the birth date is a property of the birth event, and it still opens its provenance from the Person page.
- **R-13b. Not in the other sections, for now.** Family, Timeline, Gallery and Sources are dense, and their items usually stand for more than one property, so they do not get the control this spike. Revisit as the design and the page mature.

## 1. Vitals

- **R-14.** Shows the basic facts: name, birth date and place, death date and place. These repeat the header, with the provenance control beside each one (R-13a).
- **R-15.** Shows every Person property of this canonical person: anything that appears on the person's cards today belongs here, each with the provenance control.

## 2. Family

The first quasi-family-tree visualization in the app.

- **R-16. Simple, not a canvas.** A plain laid-out diagram, not a `GraphCanvas` host or a pannable/zoomable surface.
- **R-17. The nuclear family.** Shows this person's **parents, siblings, spouse(s), and children**: the people they grew up with and the people who grew up with them.
- **R-18. Implied relationships count.** The diagram shows as much as can be concluded automatically: direct links, links implied through events (shared marriage, a birth's parents, already read through `canonlinks.Reader.For`, S11-49), and **siblings from a shared parent** (new). Extending the graph cache's implied links to make this easy is in scope.
- **R-18a. Siblings are for the page, not for matching.** Matching keeps its current patterns (spouse, parent, grandparent); Promote and the evidence sheet do not count a shared parent as sibling evidence. Revisit if matching needs it later ([`kinship-view.md`](../../ideas/kinship-view.md)).
- **R-19. Several families.** A person with more than one spouse shows each family. A person with conflicting parents shows each set. The layout handles both.
- **R-20. Every relative is a link** to that relative's Person page, so the user can bounce from person to person.
- **R-21. Each person on the diagram shows:** a likeness, name, birth date and place, death date and place, colored per R-3. This **person tile** is shared with the Timeline (R-27).
- **R-22. Lines** connect the people appropriately (parent to child, spouse to spouse, siblings through their parents).

## 3. Timeline

- **R-23. Events this person took part in directly,** in chronological order, so the section reads as a visualization of the life from birth to death. Family events the person was not a participant in (a child's birth they are not recorded at) are not listed.
- **R-24. Undated events** go in an "Undated" group at the end.
- **R-25. Each event shows** a name, a date, a location, and the person's role in it.
- **R-26. A placeholder thumbnail slot for an event depiction.** Event depictions are not built; the slot is reserved so the layout does not change when they arrive.
- **R-27. Participants.** Each event lists the other canonical people who took part, each with their role and drawn as the shared person tile (R-21), each a link to their Person page.
- **R-28. Each event links** to its Event page.

## 4. Gallery

- **R-29. Every likeness of this person** lives here.
- **R-30. A built-in viewer.** Click a likeness to expand it, and ideally step through likenesses like a slideshow. The design brief presents options.
- **R-31. Set as portrait** stays available from the gallery (and the viewer).
- **R-32. Open the Citation** a likeness was cropped from.

## 5. Sources

The user thinks of this section as "where the sources are". The data is the person's **identity claims**: every Subject connected (promoted) into this canonical person.

- **R-33. One entry per identity claim.**
- **R-34. Grouped by Source.** Since Spike 12 a Source can hold several Evidence graphs, so one Source can contribute several claims. Claims from the same Source are grouped under it.
- **R-35. Each Source group shows** the Source's title, type, and date (a new first-class Source date, [`source-provenance-date.md`](source-provenance-date.md)), and the claims' Subjects beneath it as Evidence graph subject cards.
- **R-36. Groups are ordered by Source title.**
- **R-37. Each Source links** to its Source page.

### Identity claim controls

On each claim in the Sources section.

- **R-38. Remove.** Un-promote the Subject from this person, the same behavior as the un-promote control in the Evidence graph today.
- **R-39. Removing the last claim** deletes the canonical person, wherever the unpromote happens (the Evidence graph too). A warning says so before it happens; if the user proceeds, the app returns to the Persons list.
- **R-40. Review (view only).** Open the claim's pinned evidence and its argument, so the user can see why the Subject was connected to this person. Editing pins or the argument is not this spike.

## Not this spike

- **Event and Place pages in the new layout:** later; the provenance sheet is built to be reused (R-13). They keep the Why accordion until then; it is deleted when they adopt the sheet.
- **Reconciliation Claims** (manual reconciliation) in the provenance sheet: later; the sheet leaves room (R-10).
- **Editing an identity claim's pins or argument:** later; this spike is view only (R-40).
- **Weak-claim alerts and the claim review queue:** stay parked in [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md).
- **Event depictions:** later; the Timeline reserves a placeholder (R-26).
- **Unpromoted Subjects anywhere on the page:** refused (R-2).
- **Kinship beyond the nuclear family** (grandparents, aunts and uncles, godparents, census households): stays in [`ideas/kinship-view.md`](../../ideas/kinship-view.md).

## Cross-spike dependencies

Spike 12 is expected to close before Spike 13 starts. Spike 11 may still be open, and Spike 13 runs alongside it.

**Spike 12 (assumed closed):** Source grouping (R-34) relies on multiple Evidence graphs per Source (S12-03 onward) and on per-graph unpromote (S12-09) for R-38. If any S12 PR is still open when Spike 13 starts, the PR that needs it names it as a dependency.

**Spike 11 (may be open).** Landed and relied on: kinship patterns (S11-38, #564), one link reading for bridges (S11-41, #561), kinship links in the graph cache with `canonlinks.Reader.For` (S11-49, #569). Still open and touching the same code:

| Spike 11 PR | Status (2026-10-26) | Where it meets Spike 13 |
| --- | --- | --- |
| S11-39 Matching reads kinship shapes | Blocked — re-plan | Siblings are page-only (R-18a), so the sibling inference must not reach matching's pattern set and S11-39's goldens do not move. Both still edit `core/kinship` and the graph cache's kinship results, so the sibling PR names S11-39 as a dependency or rebases on it |
| S11-40 / S11-42 Implied links in claim evidence and on the evidence sheet | Scoped | Claim review (R-40) shows a claim's evidence and argument; S11-40 changes what the claim evidence query returns for relationships. The review view should reuse that query and sheet, so it follows S11-40 or is designed to absorb it |
| S11-29 Quote nicknames and add former surnames | Planned | Changes how Person titles render: the header (R-4), Family tiles and Timeline participants |
| S11-30 Gender meets sex at birth | In review | Feeds the gender/sex color treatment reused on every tile (R-3) |
| S11-28, S11-44, S11-45 Nickname fold, nickname sets, prefix names | Planned / Scoped | Change the reconciler's outcomes and causes, which the provenance sheet explains (R-9). The sheet should word causes from the reconciler's records, not a fixed list |

**Possible preparatory refactor:** `core/kinship` still reads links through `segmentcompare` and so imports the whole compare stack ([`kinship-view.md`](../../ideas/kinship-view.md), S11-38 review M-3; S11-41 unified the reading but did not move it to its own pure package). Before the Family section's engine path imports `core/kinship`, consider moving link reading out. Note that this touches the same files as S11-39.

## Dogfood bar

The product owner has several real people in their project ready and will test them by hand as PRs land. The bar outcomes are written with the PR map.

## Related material already in the repo

- [`ideas/kinship-view.md`](../../ideas/kinship-view.md): family on person pages must read every kinship shape as one fact through `canonlinks.Reader.For` (S11-49). Notes that `core/kinship`'s link reading should move to its own pure package before a person page imports it (S11-38 review M-3), and that siblings from shared parents is not built yet (R-18).
- [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md): weak-claim alerts and re-pinning; parked.
- [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md): likeness leftovers after Spike 10.
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5 (identity claims, pins, promote), §7 (Reconciliation Claims, Participation and Relationship associations).
