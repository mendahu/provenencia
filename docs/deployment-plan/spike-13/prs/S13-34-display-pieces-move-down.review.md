# Review — S13-34 Display pieces move down a layer

**Verdict:** Fits with changes — Not ready (round 1) → Planned after the answers below
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

All answered 2026-11-02:

- **Term names everywhere** (M-2, M-3). Extend the one term rule beyond Vitals and the gender badge to the provenance sheet's "Read as" lines, Evidence graph cards (and so Promote's Subject cards) and bridge sentences, and delete the bridge's private copy. **Applied:** D-2, D-3; AC-1.1, AC-1.5; T-2, T-3; manual checks 1 and 2; the Objective names the visible change.
- **The Events list uses the kit line** (M-1). `PVDatePlaceLine` gains a date size and an extra count; the Events list's line becomes a call to it, its place taking the Person header's small secondary style. **Applied:** D-5; US-3; T-4; manual check 3.
- **The card shell's `.row` size is cut** (S13-35 review M-1). Affects S13-35 and S13-13 only; recorded here because S13-13's event row now reads `PVDatePlaceLine(…, dateSize:)` from this PR.

## Writer's own calls

- D-9's `.year` edge outputs: a month or day with no year gives "" (not the full wording); an equal-year range collapses to the year.
- D-12's exclusions: names also declared under `DesignSystem/` or `Platform/`, and SwiftUI's `Text`, are not counted as feature types; comments are stripped.
- M-5 Declined with code evidence (below); the editor flags are logged.

## Round 1

### M-1 · blocker · `PVDatePlaceLine` lands beside a feature copy that S13-13 then extends
**Answer:** To user → answered: Fixed. D-5 adds `dateSize` and `extraCount`; `EventSecondaryLine` becomes a call; S13-13 reads `PVDatePlaceLine(…, dateSize: PVTypeScale.body)`.

### M-2 · blocker · "One rule names a term value everywhere" is false
**Answer:** To user → answered: Fixed. D-3 routes `ObservationValueDisplay`'s term branch through `PropertyTermDisplay` and deletes `EvidenceBridgeEdgeSummary.termLabel`; Promote follows through `EvidenceCardView`.

### M-3 · blocker · `readAs` / `readAsIsPhrase` would keep a second rule
**Answer:** Fixed. D-2 threads `propertyKey` through both; `ReconciliationRecord` passes `field.propertyKey`; § 4 lists the sites and tests.

### M-4 · blocker · Moved paths unswept; a Go parity test reads a moved file
**Answer:** Fixed. D-8 changes `api/ffi/swift_vocabulary_test.go:51`; the header says one Go test constant changes; STEP-5 fixes the `add-date-value` link; the DoD sweeps every moved path with `rg --hidden`.

### M-5 · major · `DateValueDraft` moves into `Platform/`
**Answer:** Declined with evidence. `DateValueDisplay`'s wording runs on the draft (`DateValueDisplay.swift:6`; `string(for input:)` builds a draft at line 58 – 59), so `DateValueDisplay` cannot move without it or a rewrite; the editor flags are set from the input in `DateValueDraft(from:)` (line 330 – 332) and toggled inside the draft (line 118), so splitting them is an editor refactor. The draft is Foundation and `L10n` only. The flags are logged in `structural-debt.md` (STEP-8) with this reasoning.

### M-6 · major · The "+N" badge is three hand-built copies
**Answer:** Fixed. D-6 `PVBadge.moreCount`; `boundOverflow` logged.

### M-7 · major · The detail page title re-implements `PVTitleText`'s switch
**Answer:** Fixed. D-4 adds `Style.display(compact:)`; the detail page draws `PVTitleText`; `italicizingEmphasis` becomes private. S13-34 stays at 8 steps.

### M-8 · major · Callers re-derive the unknown halves
**Answer:** Fixed. D-5 `spokenDate(_:)` / `spokenPlace(_:)`; `PersonLifeDisplay` and `EventDetailView` use them.

### M-9 · major · The layer check is a one-off; S13-35 does not own a permanent one
**Answer:** Fixed. D-12 adds `kit_names_feature_types()` to `scripts/check-guidance-names.py`, code only, with `WorkspaceLocation` allowed and logged; S13-35 drops its own check.

### M-10 · major · Wrong keys in the sweep; the layout doc
**Answer:** Fixed. The DoD names `conclusions.list.more`, `conclusions.detail.dateUnknown`, `conclusions.detail.placeUnknown`; D-13 updates `macos-client-patterns.md:14` and writes the one layering sentence here (S13-35 drops its own).

### M-11 · minor · Two qualifier switches
**Answer:** Fixed (constants) in D-9; unifying the two formatters is logged.

### M-12 · minor · Key naming convention
**Answer:** Fixed. D-7 uses `designSystem.badge.moreCount` and `designSystem.datePlaceLine.*`.

### M-13 · minor · A dummy `""` key
**Answer:** Fixed. D-2 passes the unwrapped `toponym.propertyKey`.

### M-14 · nit · A Platform formatter names a kit nested type
**Answer:** Logged in D-5 and D-10: `Missing` is a policy value, not view state.

### X-1 · blocker · `otherValue` has four callers, not "both calls"
**Answer:** Fixed. D-2 gives `otherValue(_:propertyKey:locale:orders:)` and each call site's key.

### X-2 · blocker · Test fixtures have no key
**Answer:** Fixed. D-2 names the key per fixture line; line 155 changes from "male" to "Male".

### X-3 · major · `moreCount` read by kit code between steps
**Answer:** Fixed. D-7 moves `badgeMoreCount` in STEP-2, its first kit reader.

### X-4 · major · Wrong catalog keys in the DoD
**Answer:** Fixed (with M-10).

### X-5 · major · The layer check's output and timing
**Answer:** Fixed. D-12 harvests at run time and excludes names declared outside `Features/` and SwiftUI's `Text`, code only.

### X-6 · major · `.year` edges
**Answer:** Fixed. D-9 and AC-2.4 give the month-only, year-less, equal-year and phrase-only outputs.

### X-7 · minor · `italicizingEmphasis` callers and test files
**Answer:** Fixed. § 4 names `PersonTitleDisplayTests.swift:122`; STEP-2 lists it.

### X-8 · minor · Platform names `PVTitleSource`
**Answer:** Fixed. D-10 says so.

### X-9 · minor · Smaller choices
**Answer:** Fixed. D-1 maps an empty name to nil and the gender guard keeps line 70 nil; `NameValueDisplay.string(form:)` moves; structural-debt rows take the next free A-numbers; the S13-13 and S13-35 plans already record the hand-over.
