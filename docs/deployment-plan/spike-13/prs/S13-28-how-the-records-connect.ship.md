# Ship log — S13-28 How the records connect

**Plan:** [`S13-28-how-the-records-connect.md`](S13-28-how-the-records-connect.md) · **Branch:** `s13-28-how-the-records-connect` (stacked on `s13-27-plan`, #743) · **PR:** #744
**State:** review 1

Stacked PR: the base branch is `s13-27-plan`, not `main` (the user, 2026-11-08). The gate runs `scripts/check-pr.sh --base origin/s13-27-plan`, and reviewers diff against the `s13-27-plan` head.

## Round 0 — implement
Head `d788a00e`. Gate PASS (`--base origin/s13-27-plan`; xcode-test 70s). Report deviations (all in the plan's Landed note):
- `previewSession(entityId:recordLinks:failure:)` instead of a `delay:`; nil links answer an hour late so existing previews show the spinner.
- New file `PersonRecordLinksColumn.swift` (title, caption, width and fitting logic).
- `ConclusionPageSectionHost.emptyTitle` is `LocalizedStringResource?` (nil default) so no empty literal becomes a catalog key.
- Nodes placed with `.position`; the first frame before measuring is off by the dot's inset.
- `PVPageColumnsLayout.widths` is `@Sendable`; `PVPageColumns.columns` captures its inputs as locals (Swift warnings).
- Unguarded hover ordering: a late `onHover(nil)` leave after the next node's enter clears the highlight.
- §10 manual check not run (pin while scrolling, resize focus, short-window shrink, click-through, VoiceOver, Tab).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none (UX: `macos/App/**` and strings; SEC+PRF: `macos/App/Features/**` models loop over the record-links read's nodes and arcs).
