# Spike 13 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S13-23 — Promote stops asking for pins

**Landed 2026-10-29** in [#633](https://github.com/mendahu/provenencia/pull/633). Promote is a quick action: a researcher picks a target per row and presses Done. The evidence sheet, the assessment column (band, Contested mark, reason, chevron) and the confidence grade read are gone; the rows, headings and Already promoted rows share one grid (`PromoteGridRow`), and the list widens to 1600 pt with a 460 pt target menu at Wide. Done sends a target per row and nothing else, and the app carries no claim confidence or argument. The Updated caption reads the move reason, else the via neighbor. The share wording, five wire enums and their parity rows go, and every Promote string says promote, never file.

## S13-09 — Sections slot on the Conclusion page

**Landed 2026-10-29** in [#624](https://github.com/mendahu/provenencia/pull/624). `ConclusionDetailPage` draws the shared header, then fills everything below it from a `pageSections` slot. A convenience initializer supplies `ConclusionDetailDefaultPageSections` (Details rows, likenesses, Place relationship rows), so Person, Event and Place look the same and S13-10 can pass a custom Person fill. `docs/macos-client-patterns.md` names the slot apart from the Place hierarchy `sections`.
