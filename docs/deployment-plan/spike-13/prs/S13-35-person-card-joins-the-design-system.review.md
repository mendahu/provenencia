# Review — S13-35 The person card joins the design system

**Verdict:** Fits with changes — Not ready (round 1) → ready once S13-34 lands (status Blocked — on S13-34)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

All answered 2026-11-02:

- **Cut the `.row` size** (M-1). The shell keeps regular and compact; S13-13's event row stays a Timeline snowflake; a future event card with a surface uses the shell's sizes and becomes a recipe on its second caller. **Applied:** D-1, D-4; the Out table; S13-13 updated.
- **Term names everywhere** and **the Events list uses the kit line** (S13-34 M-2, M-1). Applied in S13-34; this plan reads the names it lands.

## Writer's own calls

- D-2's second form (`when:`) for a caller that owns an enum focus binding; the composer row keeps its own Space handler.
- D-3 moves the gallery tile and the artifact preview onto the style in this PR; `CitationLineButton` and `SourcesGraphButtonBody` are logged (different shape: a ring outside the style, and hover chrome).
- D-5: a compact focus draws raised with no badge (no caller places one).

## Round 1

### M-1 · blocker · `.row` is a different component behind a size
**Answer:** To user → answered: Fixed. `.row` cut; non-optional measures; S13-13's event row is a snowflake.

### M-2 · blocker · The layering check fails today; the permanent check was dropped
**Answer:** Fixed. S13-34 D-12 owns one permanent kit check in `scripts/check-guidance-names.py` with `WorkspaceLocation` allowed; this plan's D-8 drops its own `rg` and cites it; § 4's Recipes row is corrected.

### M-3 · major · `pvActivatesOnReturn` misses the composer's copy
**Answer:** Fixed. D-2 adds `pvActivatesOnReturn(when:_:)`; the composer row moves onto it (AC-2.7).

### M-4 · major · `PVFocusRingButtonStyle` leaves near-copies with a fixed radius
**Answer:** Fixed. D-3 `cornerRadius:`; the gallery tile and artifact preview move onto it (AC-2.6); the two other shapes logged.

### M-5 · major · Guidance and test names teach the old types
**Answer:** Fixed. D-10: tests `git mv`'d to `PVPersonCardTests` and `PVGenderTreatmentTests`; the working rule and R-3 renamed in the planning commit; the layering sentence and the `.mdc` / add-ui-component lines are S13-34's (one place); the sweep covers `docs/**/*.md`.

### M-6 · minor · Two keys with one copy
**Answer:** Fixed. D-9: one `L10n.PersonCard.noLikeness`, read by the card and the header; `personNoLikeness` deleted.

### M-7 · minor · Size measures the shell does not apply; a mirror enum
**Answer:** Fixed. `PVPersonCard` takes `PVTileCardSize`; D-4 says which measures the slots apply.

### M-8 · minor · Family's tooltip changes
**Answer:** Fixed. D-5: regular keeps the title tooltip; compact uses `helpText` (AC-2.5).

### M-9 · minor · Compact edges
**Answer:** Fixed. § 6 Edges and T-2 cover ranges, phrases, truncation and no dates.

### M-10 · nit · A file for a one-member extension
**Answer:** Fixed. D-7 builds the badge in `PersonDetailContent.badge`.

### X-1 · blocker · The layer check regex is malformed
**Answer:** Fixed with M-2 (the check is S13-34's).

### X-2 · blocker · "No recipe reads `Features/`" is false
**Answer:** Fixed. § 4's Recipes row names `PVDeleteImpact`'s `WorkspaceLocation`, allowed by S13-34's check.

### X-3 · blocker · No field for the compact span
**Answer:** Fixed. D-5 stores `years` and `helpText`; a compact focus is decided.

### X-4 · major · Layout frames not carried into D-1
**Answer:** Fixed. D-1 gives the inner frames by `size.thumbnail`, the `VStack` frame and the vertical alignment; D-4 adds `linesLineLimit`.

### X-5 · major · Tooltip vs "Family unchanged"
**Answer:** Fixed with M-8.

### X-6 · major · `noLikeness` duplicate key
**Answer:** Fixed with M-6.

### X-7 · major · The sweep cannot pass
**Answer:** Fixed. Tests renamed; `GenealogyStore.swift:468` in STEP-3's files.

### X-8 · major · Step file lists and wording
**Answer:** Fixed. `PersonLifeDisplay.swift` is no longer marked new; `PersonFamilyPreviewData.swift` dropped; STEP-1 says the fold row takes the Return modifier only.

### X-9 · major · `LifeLine.text`, `helpText`, `years` edges
**Answer:** Fixed. D-6 and T-2 give the death-only, phrase and no-lines outputs; "\n" is a separator, not copy.

### X-10 · minor · Compact numbers
**Answer:** No change: they are board S13-D4 frame 12f's, recorded in S13-13.

### X-11 · minor · S13-34 names
**Answer:** Fixed: § 4 updated to S13-34's final names (`PVTitleText` `style:`, `PVDatePlaceLine` `dateSize:` / `extraCount:`).

### X-12 · minor · Temptations
**Answer:** Fixed. The recipe preview builds its summary inline; identifiers stay at call sites; README row texts given in § 6.
