# Review — S11-42 Implied links on the evidence sheet

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet). No blocker, no Reshape, so no round 2.

## To user
- None. The three product calls (placement, both sides named, event by date and person by name) were asked and answered before the plan was written (user, 2026-10-27).

## Writer's own calls
- A direct side is worded from the link's term (`neighborTerm`: "recorded as spouses / parent and child / grandparent and grandchild"), and "recorded directly" otherwise (D-4, after M-2).
- An undated event is named by its ref; no pivot at all reads "married in a marriage", "parent and child in a birth", "grandparent and grandchild through a parent", "through another record" (D-3).
- The Source pivot's date comes from a new `Row.date`, read from the Source's own date Observations by the Events list's rule (D-10, D-11).
- VoiceOver gets its own whole format without the "·" (D-8).
- `ConclusionHeaderDisplay` extracts the two existing header switches in this PR (D-12, after M-1); it changes no title.
- A Mac-less session reports "needs a Mac" and does not start (Environment, after X-1); §10 also runs S11-40's unrun manual check, with a stop if the Card is missing (after X-2).

## Round 1

### M-1 · major · `Pivot(header:orders:)` is the third switch that titles a `CatalogConclusionHeader` by kind
**Plan:** D-5 / Codebase fit row `EvidenceBoard.Pivot`.
**Codebase:** `macos/App/Features/Promote/PromoteFlow.swift` private `conclusionTitle` / `conclusionSubtitle`; `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `title(for:orders:)` / `secondary(for:)`.
**Better shape:** One `ConclusionHeaderDisplay` in `Features/Conclusions/`; both callers and `Pivot` use it.
**Answer:** Fixed — D-12 *(new)*, STEP-2 *(new)*, T-5; D-6 builds `Pivot` through it; Codebase fit and Deletes list the two removed switches.

### M-2 · major · Direct phrases keyed by the other side's pattern; the dated/ref/none ladder repeated per pattern
**Plan:** D-3, §6 Strings, Next adds.
**Codebase:** `CatalogClaimEvidenceRelationship.neighborTerm`; `SeededRelationshipTypeTerm`; `docs/ideas/kinship-view.md` "Other patterns".
**Better shape:** Direct words from the link's term; one fallback ladder over a per-pattern wording entry, keeping the exhaustive switch.
**Answer:** Fixed — D-3 (`PatternWording`, `PivotShape`, one `implied` ladder, `recordedOn`), D-4 (`direct(term:)` keyed by `neighborTerm.productKey`), D-5 takes `term`; Strings drop from 16 to 15 keys; Next adds is one arm and two keys per pattern.

### M-3 · minor · The caption string is assembled in the view, not by the board
**Plan:** D-7, D-8.
**Codebase:** `EvidenceBoard` doc comment; `Line.thisSource` formatted in `build`.
**Answer:** Fixed — D-8: `RelationshipCard.recorded` is a formatted `String?`; the `Recorded` type is gone (D-5 returns a tuple); D-9 draws the string.

### M-4 · minor · "Next adds" sends a future Person page into Promote's internals
**Plan:** §6 Next adds.
**Codebase:** `docs/deployment-plan/spike-13/deployment-plan.md` S13-04 is a Go read.
**Answer:** Fixed — the sentence is replaced: a second reader moves the wording out of Promote in its own PR.

### X-1 · major · No check can run on this machine
**Answer:** Fixed — Environment line: a session without a Mac reports "needs a Mac" and does not start.

### X-2 · major · The manual check, and AC-1.12 with it, cannot be shown to work
**Answer:** Fixed — §10 opens by stating it also runs S11-40's manual check; a missing Card at step 3 is a stop condition (S11-40 defect).

### X-3 · major · Tests need a `PromoteFlow.Row` that no helper builds
**Answer:** Fixed — D-13: `page(_:)` loads a `PromoteFlow` as `twoHundredLines` does; T-2 uses it.

### X-4 · minor · T-3 needs a dated Observation the helper cannot build
**Answer:** Fixed — D-13: `observation` gains `date:`; dates are `CatalogDateValueInput(kind: "point", …)`, expected text through `DateRowDisplay.line`.

### X-5 · minor · Signature and conformance gaps
**Answer:** Fixed — D-6: `struct Pivot: Equatable`, `init?(header: CatalogConclusionHeader?, orders: NameFormatOrders)`; `Recorded` removed (M-3); D-8 places `recorded` after `header`.

### X-6 · minor · The spoken header has two readings
**Answer:** Fixed — D-8: the Card's `header` as built, refs included; AC-1.11 and T-2 state the Card's own neighbor is not on the page.

### X-7 · minor · STEP-1's check cannot tell the work is right
**Answer:** Fixed — every step's Check names its tests.

### X-8 · minor · STEP-2 and STEP-3 both edit `PromoteEvidenceBoard.swift`
**Answer:** Fixed — STEP-3 says its commit carries the `Pivot` type.

### X-9 · minor · `completed.md` path and the PR number
**Answer:** Fixed — full path in STEP-6 and the DoD; the status and map are written after the draft PR opens.

### X-10 · minor · Xcstrings entry shape and comments unstated
**Answer:** Fixed — §6 Strings has a Comment column and points at the skill's entry format.

### X-11 · nit · Temptations
Logged. The name mismatch is gone with the new Strings table; AC-1.10 names its sibling term; AC-1.4's drawn state is a manual check (STEP-5 says so).
