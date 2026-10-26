# Person detail page

**Status:** requirements capture (2026-10-26). Dictated by the product owner; not yet scoped into PRs. Open questions at the end are unanswered and are **not** decisions.

## Why

The Person page (`macos/App/Features/Conclusions/PersonDetailView.swift`, worded by `PersonDetailContent`, board S9-D5) is a stub: a header with name, born/died vitals and a likeness thumb, one column of reconciled property rows, a member count, and a likeness gallery. Spike 13 makes it the place a researcher goes to understand one canonical person: who they were, who their family was, what happened in their life, what they looked like, and what evidence says so.

## Layout

- **R-1. Five sections.** The page is divided into five sections: **Vitals**, **Family**, **Timeline**, **Gallery**, **Sources**.
- **R-2. Easy switching.** The user can move between the five sections quickly from anywhere on the page.
- **R-3. The pattern is a design call.** Stacked sections with jump shortcuts, tab navigation, or something else: the spike's design brief explores the options and picks an elegant one. Not decided here.

## 1. Vitals

- **R-4.** Shows the basic facts: name, birth date and place, death date and place.
- **R-5.** Shows every Person property of this canonical person: anything that appears on the person's cards today belongs here.

## 2. Family

The first quasi-family-tree visualization in the app.

- **R-6. Simple, not a canvas.** A plain laid-out diagram, not a `GraphCanvas` host or a pannable/zoomable surface.
- **R-7. The nuclear family.** Shows this person's **parents, siblings, spouse(s), and children**: the people they grew up with and the people who grew up with them.
- **R-8. Every relative is a link** to that relative's Person page, so the user can bounce from person to person.
- **R-9. Each person on the diagram shows:** a likeness, name, birth date and place, death date and place.
- **R-10. Color-coded by gender or sex.**
- **R-11. Lines** connect the people appropriately (parent to child, spouse to spouse, siblings through their parents).

## 3. Timeline

- **R-12. Every event this person is associated with,** in chronological order, so the section reads as a visualization of the life from birth to death.
- **R-13. Each event shows** a name, a date, and a location.
- **R-14. A thumbnail slot for an event depiction.** Event depictions are not built yet; the slot is reserved so the layout does not change when they arrive.
- **R-15. Each event links** to its Event page.
- **R-16. Participants.** Each event lists the other people who took part in it, each a link to their Person page, drawn with the **same person treatment as the Family section** (R-9, R-10).

## 4. Gallery

- **R-17. Every likeness of this person** lives here.
- **R-18. A built-in viewer.** Click a likeness to expand it, and ideally step through likenesses like a slideshow. The design brief should present options.

## 5. Sources

- **R-19. One entry per Subject promoted into this person.** In the engine this is the list of member Subjects (identity claims) of the canonical person.
- **R-20. Presented as Sources.** The user thinks of each member as a source: every Subject lives in an Evidence graph, and every Evidence graph belongs to a Source.
- **R-21. Grouped by Source.** Since Spike 12 a Source can hold several Evidence graphs; members from graphs of the same Source are grouped under that one Source.
- **R-22. Each Source group shows** the Source's name, some information about the Source, and the member Subject entries (as Evidence graph subject cards) beneath it.
- **R-23. Each Source links** to its Source page.

## 6. Identity claim controls

Controls on each member (identity claim) in the Sources section.

- **R-24. Remove.** Un-promote the Subject from this person. Same behavior as the un-promote control in the Evidence graph today.
- **R-25. Review / edit (new).** Open the claim's pinned evidence and its argument, so the user can see why the Subject was connected to this person. Editing the pins or the argument from here is the new capability.

## Related material already in the repo

- [`ideas/kinship-view.md`](../../ideas/kinship-view.md): family on person pages must read every kinship shape (direct links and links implied through events) as one fact, through `canonlinks.Reader.For` (S11-49). Also notes that `core/kinship`'s link reading should move to its own pure package before a person page imports it, and lists siblings-from-shared-parents as a pattern not yet built.
- [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md): weak-claim alerts, re-pin against a remaining member, reject a claim. Overlaps R-24 and R-25.
- [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md): likeness leftovers after Spike 10.
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5 (identity claims, pins, promote) and §7.3 (Participation and Relationship associations).
