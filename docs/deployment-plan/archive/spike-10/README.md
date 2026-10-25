# Spike 10 — Dogfood UX

**Done.** The annoyances logged while entering real research are gone, names match across accents and abbreviations, and a face on a photo is evidence. Live models: [`matching.md`](../../../matching.md), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`source-layer-data-model.md`](../../../source-layer-data-model.md), [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md), [`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md), [`audit-revision-history.md`](../../../audit-revision-history.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md), [`structured-name-model.md`](../../../structured-name-model.md), [`navigation-history.md`](../spike-3/navigation-history.md).

## Decisions

- **The Evidence graph draws the full card at every zoom.** Cards draw in AppKit, so the camera stays 0.25 to 1.25 with no simplified face, every card on the Source is mounted, and no cover hides the first layout. The grid is one tile. The document grows by a step before a card reaches its edge. Hit boxes are the painted controls.
- **The graph camera is session state.** Magnification and offset come back after the composer, Back, or a breadcrumb for the rest of the session. No saved camera table.
- **Box-select is a marquee of subject cards.** Hold Shift, or arm Select, and drag. A Shift rectangle adds; Shift-click toggles one card. Select disarms when its rectangle is released.
- **The composer adds a bridge and a plain Observation the same way.** The new row is where you are looking, with keyboard focus in its first field. A bridge between two subjects on the Source starts in the composer. A saved row duplicates into an unsaved form. Choosing a Citation does not start Auto Transcribe.
- **Entry defaults are silent.** The date modal opens on Year, filled from the first date in the transcription. A full name line fills an empty part list on blur; Split asks before replacing hand edits. Name part types stay a fixed-width select.
- **Shared controls, not one-off fixes.** Clicking outside a sheet closes it and asks first when it is dirty. Enter saves a single-line edit when Save is enabled and leaves it otherwise. A long select label truncates, with the full string as help. Same-letter keypresses cycle in every select. A combo box ranks a label prefix, then a later word, then any other match.
- **Opening an existing project selects the oldest user**, the earliest `users.id`.
- **Last updated is an audit scope, not a column.** Handles get a `canonical_entity` scope that moves with their claims, pins, and member evidence, not with a property or term edit. Lists show it in one fixed column.
- **List scroll is a field on the history entry**, written when you leave the list, ignored for identity.
- **The Sources list is one page.** Sort, type, evidence status, and text search persist per project and user. Status counts primaries in one grouped read. Search filters the loaded rows by title and ref; it is not the omnibar.
- **One mark per subject kind.** The Event mark is a calendar, drawn everywhere through `PVMark`.
- **Gender wins over sex at birth on the Person page.** `gender` is seeded `man`, `woman`, `non_binary`, and researchers add terms. Sex at birth borrows a tint and says “· at birth”.
- **Names match the way records are written.** Accent folding is matching only; display, `NormalizeForm`, and `sort_key` never fold, so José and Jose stay separate groupings. Abbreviation and nickname tables are catalog vocabulary. Spaced particles rejoin. A value counts by its support, and a negative Observation is a contradiction. Date windows compare in days. The Western name-format profile is a catalog row.
- **A depiction is an owned value row: a region plus an image File.** The app renders the region; Go checks, decodes, and re-encodes its own PNG and never stores the app's bytes. The image is not an Artifact and not a thumbnail. A File is released when no declared pointer remains, and one purge removes its object.
- **A likeness takes its region from the keyboard too.** Drawing stays a pointer action; **Use the whole photo** (or page) and **Use the citation's region** set the region without drawing, and Draw moves focus to Cancel and announces drawing mode.
- **A gallery is the reconciled likeness values; a preferred portrait is a display pointer.** With no choice, the earliest likeness leads the Person page and the People list. A likeness record is in the gallery by its own outcome (kept or weak), so a negative, provisional, or denied record of the same picture never leads, opens, or becomes the portrait.
- **A depiction value is pure.** `core/depictionvalue` holds the type, its shape check, and the read rule (`Readable`); `core/database/depictionvalues` holds only the queries, so the matching layers carry a likeness without reaching `core/database`. Skill: [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md).
- **Someone else's project folder is untrusted input.** A locator holds at most 256 region points and 64 KiB, in Go and in the Swift mirror; an oversized stored region reads as hidden. The app refuses a symlinked or aliased object, shard folder, or `objects/`, as `core/objectstore` does. `ThumbnailCache` decodes only a PNG or JPEG with a long edge of at most 4096 px, four at a time.
- **The likeness preview is cut off the main thread**, once per region, drawing only the face box of a PDF page with PDFKit so annotations the viewer shows are in the likeness. An edit with no redraw keeps the saved image.
- **Unpromote deletes a filing.** It deletes the Subject's claim and the bridge claims on that Source that use it. It does not reject, and the handle stays.
- **New seeds are for new catalogs.** `engagement`, `interests`, `gender`, and `likeness` come from Install. No migration adds them to an existing catalog.

## Refused

An adaptive choice field. A default role on a new connection. Omnibar whole-word rank (already right). Simplified cards, culling cards to the window, and a canvas cover. A crop cache. Related event types (birth and baptism) in matching. A handle delete. Parked items are under Leftovers.

## Leftovers

| Item | Home |
| --- | --- |
| Promote compares two likenesses as pictures; place chains; concluded values first | Spike 11 |
| Depiction leftovers, including the 2026-10-18 likeness audit items no PR took | [`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |
| International-name leftovers, including the collation locale | [`international-names.md`](../../../ideas/international-names.md) |
| Daitch–Mokotoff | [`daitch-mokotoff.md`](../../../ideas/daitch-mokotoff.md) |
| Matcher calibration and a labelled set | [`matching-test-suite.md`](../../../ideas/matching-test-suite.md) |
| Place a common cluster in one action | [`evidence-graph-structure-shortcuts.md`](../../../ideas/evidence-graph-structure-shortcuts.md) |
| Reject a claim, remove a member, weak-claim review, remember researcher decisions | [`identity-claim-review.md`](../../../ideas/identity-claim-review.md) |
| Dogfood notes still open, including census-scale layers and guided extract | [`docs/dogfood/ux.md`](../../../dogfood/ux.md) |
