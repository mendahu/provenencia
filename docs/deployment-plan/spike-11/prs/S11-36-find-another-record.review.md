# Review — S11-36 Find another record in the target menu

**Verdict:** round 1: Fits with changes — Not ready. Round 2: pending
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **Kit fixes in this PR or a preparatory one?** (M-3, M-6, X-1, X-2.) Recommended: in S11-36. **Answered** (user, 2026-10-25): in S11-36.
- **Board.** **Answered** (user, 2026-10-25): the Promote flow board, frames 20 – 26b, checked against the brief in its Board notes.

## Writer's own calls

- The search stays open when the field loses focus; it closes on a pick, Cancel, Escape, or Done (D-8, D-6). The board's frame 26 agrees.
- "Little evidence" applies to every record the researcher chose at band None, menu picks included (D-9). The board changed this from the brief's first draft (S11-D2 E03).
- The model's names use one noun, "record search" (D-4 – D-6, D-8).
- The combo box reports its own query resets only to remote callers (D-16), so no existing caller changes.

## Round 1

### M-1 · major · Two public ways to word a row's band
**Answer:** Fixed — D-9: `L10n.Promote.assessment(_:)` takes the readout; the band-only mapping is `private`; tests move to it (STEP-3).

### M-2 · major · The picker re-derives the search list's state
**Answer:** Fixed — D-13: `CatalogSearchModel.isBelowMinimumQuery` *(new)*, read by `showsResults`, `scheduleSearch` and the picker; the picker's line comes from `showsErrorState` and it. T-8.

### M-3 · major · The combo box's lost environment is patched at the call site
**Answer:** Fixed — D-17 forwards `context.environment` in `PVComboBoxPopupWindow`; D-12 no longer applies `.environment`. User chose this PR.

### M-4 · major · The flow special-cases a UI token
**Answer:** Fixed — D-2: `target(for:on:)` returns `Target?`; `setTarget` refuses any unresolved token. T-1 covers three kinds.

### M-5 · major · D-1 consolidates only the handle token
**Answer:** Fixed — D-1 spells all four tokens from `Target` and adds `Row.menuItem(forToken:)` as the one decoder; STEP-1's check covers all four. The proposal wire strings stay (Out).

### M-6 · major · `PVOmnibarHitRow.selected` becomes dead
**Answer:** Fixed — D-11 removes it with its fill and preview. User chose this PR. S11-D3's non-goal now reads "no visual change to the hit row recipe".

### M-7 · minor · The skill that teaches the search pieces is not updated
**Answer:** Fixed — STEP-7 edits `add-searchable-kind`.

### M-8 · minor · `CatalogSearchModel` keeps omnibar-only state
**Answer:** Fixed — logged in `docs/ideas/shared-list-helpers.md` by STEP-7; Out line added.

### M-9 · minor · Brief: FR-7's home names the badge view
**Answer:** Fixed — S11-D3 inventory row "Row readout" now names the one L10n wording.

### M-10 · nit · Several stems for the search
Logged; taken: "record search" across D-4 – D-8.

### X-1 · blocker · The `empty:` init is not reachable
**Answer:** Fixed — D-15 adds a public `PVComboBox` init with `empty:`; STEP-4.

### X-2 · blocker · Escape and click-away reset the field without telling the model
**Answer:** Fixed — D-16; AC-3.3, AC-3.4; T-9. The board's frame 26 has the typed text dropped on click-away and the search kept.

### X-3 · blocker · The plan is not Planned
**Answer:** Fixed — the board is agreed and copied (header **Board**, §6 frames in words, brief Board notes); Open questions is empty. Status goes to Planned after round 2.

### X-4 · major · Choices left in STEP-6
**Answer:** Fixed — D-12 names the hit lookup, `projectDir` and the empty row; D-8 names the view's parameters and the list height; D-14 says where each id goes; D-6 closes the search in `done()` and `leave()` (AC-3.5, T-4). The step is split: the kit work is STEP-4 and STEP-5, the Promote view STEP-6.

### X-5 · major · Checks not runnable off macOS
**Answer:** Declined — the plan's commands are the repo's (`add-swift-test`, `scripts/check-pr.sh`); the implementer runs on macOS (`implement-pr`). Manual checks are the board's behaviors and stay manual.

### X-6 · major · STEP-5's check is wrong
**Answer:** Fixed — the entry and the search are one step (STEP-6), so the sentinel never reaches a mounted `PVSelect` between steps.

### X-7 · minor · Plan and brief disagree on FR-7 and FR-8
**Answer:** Fixed — the board settles FR-7 (D-9, AC-4.3, AC-4.4; the brief's last sentence corrected). AC-5.1 uses the fake store's title. `accessibilityLabel(for:orders:)` takes the session's name order (D-10, T-10).

### X-8 · minor · Tests
**Answer:** Fixed — T-5 uses `model.recordSearch!.search`; AC titles use the fake's labels; existing `"handle:…"` test literals stay (Out).
