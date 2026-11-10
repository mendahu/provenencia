# Review — S15-32 Copy Source can carry the date

Plan: [`S15-32-copy-source-date.md`](S15-32-copy-source-date.md). Brief: [S15-D13](../design/archive/S15-D13-copy-source-dialog.md), reviewed with it.

**Verdict:** Fits with changes — Ready with changes (round 1's only open blocker, X-1, closed when the S15-D13 board was agreed on 2026-11-10)
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet). No round 2: round 1 raised no maintainability blocker, and the executability blockers were the board (expected) and the build host.

## To user
- None. The S15-D13 board (2026-11-10) was read against the brief: CS-1 – CS-6 and CS-W1 are drawn as written, the date tick first, so the plan moved to Planned.

## Writer's own calls
- The tick is named "Published date" (S15-04 D-14: the date's one name), with the date itself as its hint; the brief recommends it first among the parts.
- Availability and the hint read `SourcesRowDate.line` on the loaded workspace (D-4), so a date with no visible line is not carried.
- The copy's date goes through `sources.SetDate` after `sources.Create`, not a date on `sources.CreateInput` (D-2).
- The value-row owner check stays parked; its idea file stops naming S15-32 (D-6).
- One PR, not an engine PR split from the tick (X-1): the PR is small and the board is a one-frame enhancement.

## Round 1

### M-1 · major · The fourth Copy-form part is a fourth hand-copied tick; extract it now
`SourceCopyForm.partCheckboxes` repeats one block three times; D-5 added a fourth with a different initializer.
**Answer:** Fixed — D-5 now adds a private `partTick` that every part goes through (`verbatim:` initializer); Codebase fit, Deletes, §6 macOS, STEP-5 and the brief's inventory row say so. The model's `include && available` lines stay (four one-liners; folding them adds a type for no rule).

### M-2 · minor · The plan edits `SourceCopyModel` but leaves its logged missing-Source debt unanswered
`docs/ideas/structural-debt.md` row "S15-03 plan" names `SourceCopyModel`.
**Answer:** Fixed — an Out row says why it stays (the Sources list host has no `SourcePageContext`) and STEP-6 adds that to the debt row.

### M-3 · minor · "Next adds" gives S15-08 the wrong slot
**Answer:** Fixed — Next adds now names `DraftArtifact` for S15-08 and the S15-07 overlap in `ReadDraft` / `CreateFromDraft`.

### M-4 · minor · `SourcesRowDate` gets a third caller that is not a row
**Answer:** Fixed as a log — an Out row and a new `structural-debt.md` row (STEP-6); the rename touches files this PR does not otherwise edit.

### M-5 · nit · Each new Copy part widens `copySource` in six places
**Answer:** Logged — an Out row, not planned: no further Copy part is named in the map.

### X-1 · blocker · The plan cannot start: Blocked on a board that does not exist
**Answer:** Declined as a plan defect — this is the gate working (`plan-pr` § 3: status `Blocked — on board`). Splitting the engine into its own PR was considered and not taken (Writer's own calls).

### X-2 · blocker · The checks for STEP-2 – STEP-5 cannot run on this host
**Answer:** Fixed — a **Runs on** header line: macOS with `protoc` plugins and Xcode, per `README.md`; without them, stop, never hand-edit generated files.

### X-3 · major · STEP-1 tests: how to read the revision's changes and build the missing-row fixture
**Answer:** Fixed — STEP-1 names a test helper `revisionChanges` modeled on `audit/scopes_effects_test.go` `loadRevisionChanges`, keeps `runCopy`, cites the `PRAGMA foreign_keys=OFF` precedent in `sources_test.go`, and states no `TestCopy` case asserts the full change list (the risk row is gone).

### X-4 · major · STEP-4 Swift tests leave locale and fixture open
**Answer:** Fixed — T-6 compares with `SourcesRowDate.line(date)`, not a literal; STEP-4 names the `CatalogDateValueInput` fixture and that `FakeStore.getSourceWorkspace` keeps `date`. T-9 drives `includeDate` (already in §8).

### X-5 · major · Version bump is ambiguous
**Answer:** Fixed — the header names 0.8.0, the rule if `main` moved, and all four `MARKETING_VERSION` lines.

### X-6 · major · STEP-6 doc moves under-specified
**Answer:** Fixed — STEP-6 gives the archive move as S15-D2's (`git mv`, one more `../` per link, the README row text, the inbound links to repoint), the `completed.md` shape and the map status. The map's working rule now says "Published date tick".

### X-7 · minor · Smaller choices
**Answer:** Fixed — D-1 copies the struct before clearing `ID`; D-5 notes the identifier and the L10n key share a string in separate namespaces; the risk row on change-order assertions is removed. AC-1.2's list half, AC-1.4 and AC-1.8 keep their manual-check lines and Go tests (T-1, T-4).
