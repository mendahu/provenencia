# Ship log — S13-36 The provenance sheet joins the design system

**Plan:** [`S13-36-provenance-sheet-in-the-kit.md`](S13-36-provenance-sheet-in-the-kit.md) · **Branch:** `s13-36-provenance-sheet-in-the-kit` · **PR:** #702
**State:** ready

## Round 0 — implement
Head `7dd0f904`. Gate PASS (orchestrator re-ran). Report: deviations: one row of S13-15's plan edited, because `check-plan-docs.py` fails on the renamed `ProvenanceSheet.swift`; S13-15's map row reworded; `Landed in #702` per the lint; README row moved to STEP-6, which the kit check needs; `PVProvenancePreviewData` shared by the previews. Manual check, previews, VoiceOver not run (no GUI).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. SEC+PRF routed by `macos/App/Features/**` models and list code built from wire records.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC, PRF | major | The kit table draws every record eagerly, from a wire count with no bound | Measure → F-11, then decide | Verified: `VStack`, no `LIMIT` in `sqlOutcomes`. D-7 and § 6 chose eager, and lazy rows put AC-1.6's Tab order at risk with no app to check it. A deferral needs a measurement at scale, so F-11 measures; the result decides Follow-up or To user |
| CON-1 | CON | major | S13-D5 links the archived brief | Fix now → F-1 | Verified: full `check-plan-docs.py` errors. The gate lints only files the branch changes, so it passed |
| MNT-1 | MNT | major | The kit's guidance says "thin map"; the assembler pattern is unwritten | Fix now → F-2 | Verified: `add-ui-component/SKILL.md:142`, `design-system-layers.mdc:21`. The user's working rule of 2026-11-03 already sets the shape |
| MNT-2 | MNT | major | The lone row needs a `face` S13-15 has no value for | Fix now → F-3 | Verified: `face` is read only under `showsReadAs`; the field-to-face rule is inside `Shown.init` |
| MNT-3 | MNT | major | `Rectangle` hairlines beside `PVDivider` | Fix now → F-4 | Verified, same-pass cleanup: seven copies by body, three on this PR's files. `PVDivider` is that body |
| PRF-1 + MNT-9 | PRF, MNT | minor | An open Why rebuilds a whole content in the row's `body` for two fields | Fix now → F-5 | Verified: `ReconciledValueRow.swift:46-54` |
| UX-1 | UX | minor | With no Read as, VoiceOver reaches the record's line after its links | Fix now → F-6 | Verified by reading: Outcome follows Citation. Code this PR adds, first used by S13-15. Supersedes D-15's "moves to its Outcome cell" |
| MNT-4 | MNT | minor | The recipe reads `L10n.Conclusions.sourceCount` | Fix now → F-7 | Verified: three lines; D-12 says through `Platform/` |
| MNT-5, MNT-6, MNT-7, MNT-8 | MNT | minor | "shown" names three things; four reserved columns; three hand-built `Shown`; `shown: Shown?` unused | Fix now → F-8 | Verified at the lines cited |
| CON-2, MNT-10, CON-3 | CON, MNT | minor | Docs left false | Fix now → F-9 | Verified: `provenance-sheet.md`, S13-15 lines 3 and 7, Landed note (4) |
| CON-4 (from UX) | CON | minor | No PSK-7 or PSK-9 preview | Fix now → F-10 | Verified: STEP-5 lists them, four previews lack them |
| CON-5 (from MNT) | CON | minor | Parts' home differs from Codebase fit | Fix now → F-9 | Verified; add to the Landed note |
| CON-6 (from UX) | CON | minor | `PVProvenanceSheet` cannot omit its value line (PSK-1, PSK-4) | Rejected | Only a likeness needed that, and tiles have no sheet (user, 2026-11-03); D-10 builds no unused path |
| UX-2 | UX | nit | Comments on the two column keys say "Why table" | Fix now → F-7 | A stale comment this PR carried over |

Nits (logged): `MNT-11 · PVProvenanceSheetContent.swift:331 · valueGroups is API only for a test` · `MNT-12 · :155 · content.records has no app caller` · `MNT-13 · PVProvenanceColumns.swift:19 · instance spacing forwards the static` · `MNT-14 · ProvenanceSheetItem.swift:1 · unused import SwiftUI` · `MNT-15 · CitationLineDisplay.swift:24 · one very long comment line` · `MNT-16 · PVProvenanceSheetContent.swift:194-196 · inner let listed shadows the outer`

### Fix plan 1

**F-1 — S13-D5 links the archived brief** (from CON-1)
Files: `docs/deployment-plan/spike-13/design/S13-D5-gallery-and-viewer.md:8`.
Do: link `archive/S13-D13-provenance-sheet-in-the-kit.md`.
Check: `python3 scripts/check-plan-docs.py` with no file arguments reports 0 errors.

**F-2 — The guides state the content recipe shape** (from MNT-1)
Files: `.agents/skills/add-ui-component/SKILL.md`, `.cursor/rules/design-system-layers.mdc`, `docs/design-system-layers.md`, the plan's Rules this plan bends.
Do: keep "thin map" for chrome, and add one short rule: a recipe may own assembly as pure content built from `Catalog*` types and `Platform/` formatters (`PVPersonCardContent`, `PVProvenanceSheetContent`), built on open, tested as pure functions, and it reports a link for the caller to map. Name the bend in the plan with the working rule "The provenance sheet is a kit recipe" as the user's answer.
Check: `scripts/check-guidance-names.py` passes.

**F-3 — A lone row needs no face** (from MNT-2)
Files: `PVProvenanceRecordRow.swift`, `PVProvenanceSheetContent.swift`, `PVProvenanceRecordTable.swift` (preview), the S13-15 plan's D-9 row.
Failing test first: `PVProvenanceColumnsTests`-style `aFaceComesFromTheField` — a date field gives `.date`, any other `.text`.
Do: `PVProvenanceFace.init(field:)` holds the rule and `Shown.init(field:)` calls it. `PVProvenanceRecordRow.face` defaults to `.text`, documented as read only with Read as on; the no-Read-as preview passes none.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — One hairline** (from MNT-3)
Files: `PVProvenanceRecordTable.swift`, `ReconciliationRecordRows.swift`, `ReconciledValueRow.swift`, `PVSectionHeader.swift`, `SourcePageNotesView.swift`, `SourcePageArtifactsView.swift`, `SourcePageIdentityHeader.swift`.
Do: replace each `Rectangle().fill(color).frame(height: 1)` with `PVDivider(color:)`, or `PVDivider()` for `borderSubtle`, and delete both `rule` helpers. Leave any hairline that sets a width.
Check: `grep -rnE -A2 'Rectangle\(\)' --include='*.swift' macos/App | grep -E 'frame\((width: nil, )?height: 1\)'` prints only `PVDivider.swift`'s line; scripts/check-pr.sh PASS.

**F-5 — The Why builds its own content** (from PRF-1, MNT-9)
Files: `ReconciledValueRow.swift`, `ReconciliationReasoningView.swift`.
Do: `ReconciliationReasoningView` takes `model: ReconciledValueRowModel` in place of `title`, `label`, `face`, `groups`, and calls `provenanceContent()` in its own `body`, so a parent redraw with an equal model does not rebuild it.
Check: scripts/check-pr.sh PASS.

**F-6 — The record's line comes first with no Read as** (from UX-1)
Files: `PVProvenanceRecordRow.swift`, the plan's Landed note.
Do: with `showsReadAs` false, the spoken element sits at the row's leading edge in an `.overlay(alignment: .topLeading)` that takes no layout space, not on the Outcome cell. The Outcome cell stays hidden, as in the sheet. The Landed note says this supersedes D-15's sentence and is unverified with VoiceOver.
Check: scripts/check-pr.sh PASS.

**F-7 — The recipe's words** (from MNT-4, UX-2)
Files: `ReconciledValueDisplay.swift`, `PVProvenanceSheetContent.swift`, `L10n.swift`, `Localizable.xcstrings`.
Do: `ReconciledValueDisplay.sourceCountLine(_ count: Int) -> String` wraps `L10n.Conclusions.sourceCount`; the three recipe lines and `ReconciledValueDisplay`'s own calls use it. The `column.readAs` and `column.outcome` comments say "Provenance table column heading".
Check: `rg 'L10n\.Conclusions\.' macos/App/DesignSystem/Recipes/ProvenanceSheet` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes.

**F-8 — Names and small copies in the assembler** (from MNT-5, MNT-6, MNT-7, MNT-8)
Files: `PVProvenanceSheetContent.swift`, `PVProvenanceParts.swift`, `PVProvenanceRecordRow.swift`, `ReconciliationRecordRows.swift`, `ReconciledValueRowModel.swift`, the tests.
Do: `OwnerLine.isShownOwner`, and the local `displayed`. `PVProvenanceReservedCell` in the parts file replaces all four. `Shown.init(face:lead:emptyText:badge:)` normalizes an empty lead to nil; `spanning`, `derived` and `empty` call it. `shown:` is required on `init(field:…)`.
Check: scripts/check-pr.sh PASS.

**F-9 — Docs** (from CON-2, CON-3, CON-5, MNT-10)
Files: `provenance-sheet.md`, `S13-15-likeness-viewer.md` lines 3 and 7, the plan's Landed note.
Do: §§ 3, 4, 5: say what S13-36 moved or deleted; the sheet takes `columns` and `idPrefix`; the content row names `Shown`. S13-15 says S13-36 landed in #702. Correct Landed note (4), and record that the owner line, value heading and column headings live in `PVProvenanceParts.swift`.
Check: `python3 scripts/check-plan-docs.py` with no file arguments reports 0 errors.

**F-10 — The two missing previews** (from CON-4)
Files: `PVProvenanceRecordTable.swift` or `PVProvenanceSheet.swift`.
Do: a 30-record sheet (PSK-7), and PSK-2 and PSK-4 in dark (PSK-9), on `PVProvenancePreviewData`.
Check: the app builds.

**F-11 — Measure the table** (from SEC-1). No commit.
Do: in a temporary test, not committed, host `PVProvenanceRecordTable` in an `NSHostingView` at 880pt and time `layoutSubtreeIfNeeded` at 30, 300 and 1,200 records. Report three runs of each in your report.
Check: the numbers are in the report; `git status` shows no test added.

## Round 1 — fix
Head `74def9fe`. Gate PASS (orchestrator re-ran); full `check-plan-docs.py` 0 errors. F-1 to F-10 done; F-3 ran red first. Deviations: commits grouped by file, not by `F-#`; S13-15's Status, Depends and closing sentence edited. Minors read by the orchestrator: UX-1 (F-6), UX-2 and MNT-4 (F-7) **Resolved**; F-6 is unverified with VoiceOver. The rest go to round 2's reviewers.

F-11, `PVProvenanceRecordTable` in an `NSHostingView` at 880pt, first `layoutSubtreeIfNeeded`, Debug build, three runs: 30 records 172, 157, 155 ms; 300 records 1595, 1713, 1825 ms; 1,200 records 8704, 10693, 12155 ms. Linear, about 5 to 10 ms a row. The old `Grid` was not timed.

## Round 2 — review
Reviewers: CON (Sonnet, for CON-1), MNT, SEC+PRF. Skipped: UX (no Fix-now blocker or major). Scope: `7dd0f904...74def9fe`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Round 1 | Resolved | Full `check-plan-docs.py` 0 errors |
| MNT-1, MNT-2, MNT-3 | MNT | major | Round 1 | Resolved | Guides updated; `face` defaults; seven hairlines on `PVDivider` |
| PRF-1 + MNT-9 | PRF, MNT | minor | Round 1 | Resolved | `provenanceContent()` is called in `ReconciliationReasoningView.body` |
| MNT-5, MNT-6, MNT-7, MNT-8, MNT-10, CON-5 | MNT | minor | Round 1 | Resolved | MNT-7 has a residual, below |
| SEC-1 | SEC, PRF | major | Eager table, unbounded count | **Not resolved → To user** | F-11 does not clear it: 5 to 9 ms a row on the main thread on first layout, no cap, Debug only, no `Grid` baseline. Content assembly is O(n²) for negative records and unmeasured; the Why has the same input. Reverses D-7 either way, and a cap adds chrome to a "no visible change" PR |

No new blockers from CON, MNT or SEC+PRF. UX not re-run.

Open minors, no fix plan without the user: `MNT-7 residual · ReconciledValueRowModel.swift:102,134 · the empty-lead rule still lives in four places` · `guides · "built when the recipe opens" does not describe PVPersonCardContent` · `Landed note · omits shown: required and Shown.init(face:…); (8) says zero-size, the code is 1×1; § 6 Assembler still shows shown: Shown? = nil` · `ReconciledValueRow.swift:48-49 · padding one indent shallow` · logged: `PVPageColumns.swift:179 · hairline without Rectangle` · `PromoteView.swift:393, PVSidebarNav.swift:102 · hand-rolled vertical dividers`

**To user, answered:** SEC-1: the user accepted the risk and asked for an idea file (2026-11-03): [`provenance-record-cap.md`](../../../ideas/provenance-record-cap.md). The table stays as D-7 has it. The stated ceiling is about 300 records on one value; no Release figure was taken, and the idea file says so. The open minors above were not applied: they go to the PR body's follow-ups.

## Result
Gate PASS · CI pending · open: none · follow-ups: SEC-1, the open minors above · to user: SEC-1 answered · CON miss: none
