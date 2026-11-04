# Review — S13-36 The provenance sheet joins the design system

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- Answered. M-18: D-8 hands a link the `Record`, where the working rule and `provenance-sheet.md` § 5.1 (user, 2026-11-03) give it the outcome. The user chose the `Record` (2026-11-03).

## Writer's own calls
- One PR, not two (D-1): the size rule does not trigger.
- The board is taken as agreed (user, 2026-11-03: "it is drawn"); where it differs from the app, the app wins (D-11), since the map's Check is "looks and behaves as before".
- No `.narrowed`: it has no caller since tiles lost their sheet (D-10).
- The Why shares the recipe's owner line, outcome and read-as parts, narrowing the brief's inventory row (D-9).

## Round 1
### X-1 · blocker · STEP-2 leaves the old sheet and the Why with no way to navigate
**Answer:** Fixed — STEP-3 gives `ReconciliationRecordRows` the callback for both modes.
### X-2 · blocker · The assembler's signatures are not given
**Answer:** Fixed — § 6 **Assembler**.
### M-1 · blocker · The sheet header gets two derivations
**Answer:** Fixed — `Shown` (D-16), held by the row (D-5); `provenanceContent()` passes the row's label and `shown`.
### M-2 · blocker · The kit sheet takes a `close` closure beside `PVSheetClose`
**Answer:** Fixed — `\.pvSheetClose`, § 6 macOS Sheet row; § 4 row added.
### M-3 · blocker · Two existing copies move unchanged
**Answer:** Fixed — D-2's second commit; `runs(of:)` in STEP-3.
### X-3 · major · STEP-2 is five changes in one step
**Answer:** Fixed — STEP-2, STEP-3, STEP-4.
### X-4 · major · The Why's callback
**Answer:** Fixed — D-8: the Why's Source cell reports `.observation`.
### X-5 · major · The sheet's ids
**Answer:** Fixed — D-13.
### X-6 · major · The row's `otherValues` and `listedValues`
**Answer:** Fixed — D-5, D-16, `value(…)`.
### X-7 · major · Key strings
**Answer:** Fixed — D-12's key rule and its stays list.
### M-4 · major · The callback hands back the outcome
**Answer:** To user, answered — the `Record`; see M-18.
### M-5 · major · `CitationLineDisplay.line(_:)` loses its last app caller
**Answer:** Fixed — D-4, Deletes.
### M-6 · major · The Why's constants
**Answer:** Fixed — D-9, Deletes.
### M-7 · major · The link mapping's home
**Answer:** Fixed — D-8.
### M-8 · minor · `recordCount` stored
**Answer:** Fixed — D-5.
### M-9 · minor · The sheet width
**Answer:** Fixed — D-15 `panelWidth`.
### M-10 · minor · `groupRank` on `Record`
**Answer:** Fixed — D-4.
### M-11 · minor · `showsHeadings`
**Answer:** Fixed — D-10.
### M-12 · minor · Docs under-listed
**Answer:** Fixed — STEP-8.
### X-8 · minor · D-2 and the STEP-1 check
**Answer:** Declined — no plan change; the Stop condition covers it.
### X-9 · minor · The row's `==`
**Answer:** Fixed — § 6 Scale.
### X-10 · minor · Rule and padding details
**Answer:** Fixed in part — § 6 macOS Table row; the rest is manual check 1.
### Nits
M-13 Fixed (`open:`). M-14 Fixed (`face`). M-15 Fixed in part (`PVProvenanceFace` beside the content). M-16 Fixed (D-3). M-17 Fixed (D-1).

## Round 2
### M-18 · blocker · D-8 bends a user-stamped working rule
**Answer:** To user, answered 2026-11-03: the `Record`, in § 5 **Rules this plan bends**. The Context line and STEP-8's brief row are Fixed.
### M-19 · blocker · `valueLine` has two callers
**Answer:** Fixed — `PVProvenanceValueLine`, D-16, Codebase fit.
### X-11 · major · `ReconciledValueRow.Provenance`
**Answer:** Fixed — STEP-6: the enum stays.
### X-12 · major · STEP-3 builds content before `Shown` exists
**Answer:** Fixed — STEP-3, STEP-4.
### X-13 · minor · How the Why obtains `open`
**Answer:** Fixed — STEP-6.
### X-14 · minor · Details left to the implementer
**Answer:** Fixed — STEP-4, T-3, § 10's base.
### Outside the lens: STEP-2's check
**Answer:** Fixed — word-bounded `rg`.
### M-15 · nit, recheck
**Answer:** Fixed — Codebase fit row.
