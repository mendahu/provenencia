# Review — S10-19 Event mark and sidebar icons

**Verdict:** Fits — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **M-3, a shared symbol-or-mark value.** Answered 2026-10-14: add `PVGlyph` in this PR. D-3 was rewritten.
- **X-14, commit before dispatch.** The plan, this log, `docs/ideas/glyph-consolidation.md`, and the map and ideas-index edits are uncommitted. Recommendation: commit them to `main` as one docs commit before `ship-pr` runs.

## Writer's own calls

- The planned PR is S10-19. The request named S10-29, which already landed (#377) and is about name patterns. The user confirmed S10-19.
- D-2: the toolbar "section icon" in WS-3 is the Back and Forward history jump-menu rows. They are the only toolbar code that reads the section icon, and board Frame 14 says the same.
- D-6: no 0.7 rest dim on marks. The Swift `PVIcon` has no rest dim, and the board asks for parity between marks and symbols.
- D-3: `PVGlyph` lives at the end of `PVMark.swift`, not in a new file, so the PR makes no hand edits to `project.pbxproj`.
- D-10: `PVSymbol.person` and `PVSymbol.mapPin` are deleted, because nothing reads them after this PR.
- STEP-8: the landing date is the day the PR opens. The user corrects it at merge if it differs.

## Round 1

### M-1 · major · Two copies of the section-to-mark mapping, kept equal only by T-3
**Answer:** Fixed. D-4: `WorkspaceSection.glyph` reads `PersonsList.mark`, `EventsList.mark`, and `PlacesList.mark` for the Conclude sections. The old T-3 is dropped. T-2 pins the keys. Deriving `title` and `identifierRoot` from a section is Out, not planned.

### M-2 · major · Two more components depend on a Recipe type, and no doc says this is allowed
**Answer:** Fixed. D-11 and STEP-7 add the rule to `docs/design-system-layers.md` and `MARKS.md`.

### M-3 · major · "Symbol or mark" ends up written several ways, and some fallback symbols are never drawn
**Answer:** To user, then Fixed. The user chose `PVGlyph` (2026-10-14). D-3, D-4, D-5, D-7, D-8, and D-10 were rewritten. `WorkspaceSection.icon`, `PVSidebarNavItem.icon`, `HistoryJumpRow.icon`, `PVSymbol.person`, and `PVSymbol.mapPin` are deleted. The three older sites are logged in `docs/ideas/glyph-consolidation.md`.

### M-4 · minor · The Go registry already owns subject kind → mark, and Swift repeats it in many places
**Answer:** Fixed by logging it in `docs/ideas/glyph-consolidation.md`, and Out in §3. With M-1, this PR adds no new copy.

### M-5 · minor · `MARKS.md` doesn't learn the new placements, and its subject table is already short a key
**Answer:** Fixed. STEP-7 adds the shell-chrome size row, the ink rule, the Glyph section, and the missing `subject_place_relationship` row.

### X-1 · blocker · STEP-6 check `rg -n "emptyIcon" macos` can never print nothing
**Answer:** Fixed. STEP-4's check is scoped to `macos/App/Features/Conclusions macos/ProvenenciaTests`. D-9 and §3 Out say the vocabulary panes' `emptyIcon` is a different field and stays.

### X-2 · blocker · STEP-5 check contradicts the step
**Answer:** Fixed. STEP-4 expects `16`: the 15 at the base plus the preview line. It also checks `git diff --stat`.

### X-3 · blocker · The `section.icon` stop-condition command cannot fire
**Answer:** Fixed. Every `rg` uses `section\.icon` with Rust-regex alternation, and the expected hits are listed with line numbers.

### X-4 · major · STEP-1 leaves the trailing newline and the `color` attribute unspecified
**Answer:** Fixed. STEP-1 drops the wrapper's newline, decodes with `html.unescape`, and writes no trailing newline. D-1 says the missing `color` and `<title>` are expected.

### X-5 · major · STEP-1 check does not verify the drawing; the tool may be unavailable
**Answer:** Fixed. The check now covers the band, the five dots, and the file ends. The stop condition names the user as the unblocker.

### X-6 · minor · Board path warnings; plan still Draft
**Answer:** Fixed. The board ids are plain text in the header. Status goes to Planned after this review.

### X-7 · minor · SwiftUI shape of the branches left open
**Answer:** Fixed by D-3. `PVGlyphView` is one view, so the modifiers chain directly. D-6 and the §4 Mark recipe row confirm that `PVMark` has no foreground of its own. A stop condition covers a mark drawn in its own color.

### X-8 · minor · STEP-8 docs edits leave choices open
**Answer:** Fixed. STEP-8 now specifies:
- the S10-18 entry shape, with the table and the Plan line deleted and the Briefs line kept;
- " — done, #N" on the `✎ S10-D7 ──▶ S10-19` line;
- the full text of the new working rule;
- the `../` link fix in the archived briefs;
- the timing of the date and PR number.

### X-9 · minor · T-3 does not test what its name claims
**Answer:** Fixed by M-1. T-3 is gone. AC-4.3 holds by construction, and T-2 and manual check 6 cover it.

### X-10 · minor · Manual check is not runnable by an agent as written
**Answer:** Fixed. §10 says the implementer marks a line "not run — for the reviewer" when it cannot drive the app, and it names the two projects to use.

### X-11 · nit · Step size, commit mapping, temptations
**Answer:** Fixed. Docs are split into STEP-7 (design-system docs) and STEP-8 (bookkeeping). Each step names its commit. The stale comments are listed in §3 Out.

## Round 2

Maintainability: **Fits**, with no new blockers. Every round-1 answer was verified against the code. These notes fell below the round-2 bar and were applied anyway:
- D-3: `PVGlyph`'s doc comment separates it from `PVFileTypeGlyph`.
- D-11 and STEP-7: the layering exception names only recipe-layer types (`PVMarkKey`, `PVGlyph`), because `PVSymbol` is a Component type.
- The vocabulary panes' empty-icon drift is logged in `docs/ideas/glyph-consolidation.md`.

Executability: **Ready with changes**.

### X-12 · blocker · STEP-8's last check cannot pass as written
**Answer:** Fixed. The grep excludes `docs/deployment-plan/spike-10/prs/**`.

### X-13 · major · STEP-1 is a retype of about 5 KB of base64, with no integrity check
**Answer:** Fixed. STEP-1 now runs `render_preview`, then `curl` of the `serve_url` straight into the repo file, so there is no decode and no retype. The check pins the size (8218 bytes) and the SHA-256 (`477f7565…b7b2b7`), which the planner computed from the same export on 2026-10-14. A hash mismatch is a stop condition.

### X-14 · major · The plan and its inputs are uncommitted
**Answer:** To user. The plan, this log, the idea file, and the map and ideas-index edits need to be on `main` before `ship-pr` dispatches. Status is now Planned.

### X-15 · minor · STEP-8's map entry contradicts itself
**Answer:** Fixed. The S10-18 shape keeps the table. STEP-8 now inserts the landed line, points **Briefs:** at the archive, deletes **Plan:**, and keeps the table.

### X-16 · minor · STEP-8 leaves small doc choices open
**Answer:** Fixed. STEP-8 now specifies each of these:
- the archived-brief landed line;
- the `completed.md` heading;
- the `design/README.md` Done-row form;
- where the README appends go (moved into STEP-7).

The PR-sequence suffix stays " — done, #N", as already decided.

### X-17 · nit · Commit and check details
**Answer:** Fixed. STEP-2 says it has no commit of its own. Codebase fit now matches D-8 ("store `.symbol(icon)`"). The `git diff --stat` check in STEP-4 is scoped to the step's own changes.
