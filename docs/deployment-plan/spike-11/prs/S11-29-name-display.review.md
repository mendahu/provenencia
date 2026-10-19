# Review — S11-29 Quote nicknames and add former surnames

**Verdict:** Fits with changes — Not ready → fixed in round 1; round 2 Ready
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. Product calls were asked before planning (2026-10-19): curly quotes, parentheses + italic, new catalogs only.

## Writer's own calls
- Former surname sits before the surname (*Mary (Smith) Jones*), role `family` for matching.
- Italics only on the Persons list row and Person page title; every other surface shows the parentheses as plain text.
- No design brief: existing titles gain punctuation and one italic run.
- A Swift golden that shows a `nick` part is updated to the quoted string.

## Round 1

### X-1 · blocker · Run splitting and the L10n shape unspecified
**Answer:** Fixed — D-7 (`NamePartDecoration.runs`, private-use slot, two-piece rule with bare-word fallback), D-8 (`NameDisplayRun`, separators are plain runs), Strings (static resource constants), T-7.

### X-2 · blocker · How the title knows a run is emphasized
**Answer:** Fixed — D-9 / D-10: one `attributed` call; `.name(_ styled:)` keeps `styled` only when a run has `.emphasized`; form fallback is plain.

### X-3 · major · AC-1.4 callers
**Answer:** Fixed — §4 lists every `NameValueDisplay.string` caller; D-13 / Out: the header-less omnibar hit keeps its indexed title; AC-1.4 by T-4 (shared function) and manual step 2.

### X-4 · major · Enum payload change sites
**Answer:** Fixed — D-9 splits the `text` arm; macOS table names the three matches.

### X-5 · major · Checks need macOS
**Answer:** Fixed — §7 preamble and a stop condition: Swift checks and the manual check go to a macOS runner.

### X-6 · major · No fallback for an upright italic
**Answer:** Fixed — D-11 binding fallback (italic `.font` on the run).

### X-7 · major · Test mechanics
**Answer:** Fixed — `western` fixture; T-2 `TestWesternSeedOrder`; T-6 `TestFormerSurnamePairsAsFamily` in `core/matchregistry` (value compare tests can't import the registry); `partCodes` unchanged (§4 Precedent); T-5 counts emphasized runs only.

### X-8 · minor · Doc anchors
**Answer:** Fixed — §4.2a, §4.5, fenced §4.1 line, international-names wording in STEP-5.

### X-9 · minor · Row vs sort key
**Answer:** Fixed — §4.5 rewrite (M-1) and a risk row.

### X-10 · nit — logged.

### M-1 · major · structured-name-model §4.5 becomes false
**Answer:** Fixed — STEP-5 rewrites requirement 2 and adds the Decoration row; comments on `Format` and `runs`.

### M-2 · major · Name stored and joined twice
**Answer:** Fixed — D-8, D-9, D-10; T-5 asserts `text == String(styled.characters)`.

### M-3 · minor · Two switches per decorated type
**Answer:** Fixed — `NamePartDecoration` and one `decoration` property (D-6); split lives on it (D-7).

### M-4 · minor · Magic count; Sort and parity
**Answer:** Fixed — T-1 asserts the key list; `Sort` and Go↔Swift parity logged in international-names (Out).

### M-5 · minor · `NameFormSplit` asymmetry
**Answer:** Fixed — Out line, logged in international-names.

### M-6 · nit · Name `TestSeedMatchesRegistryWestern`
**Answer:** Fixed — §4, D-5, T-2.

### M-7 · nit · Key segment
**Answer:** Fixed — `nameValue.display.former_surname`.

## Round 2

Reviewer (opus, both lenses): no new blockers; every round-1 answer accepted. Verdict: Ready. Notes, all fixed:
- R2-1 · minor · `conclusionheaders/order.go` `personSortKey` comment says the key is the row string — added to STEP-1.
- R2-2 · minor · D-7 split call — `components(separatedBy:)`, count then drop empties; T-7 case for `(%@`.
- R2-3 · nit · `attributed` fallback narrower than `string` — "when the runs are empty".
- R2-4 · nit · list-row italic font named (`PVFont.body(size: PVTypeScale.body)`); map row status updated.
