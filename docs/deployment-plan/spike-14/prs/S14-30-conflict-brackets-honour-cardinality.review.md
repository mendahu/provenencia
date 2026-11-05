# Review — S14-30 Conflict brackets honour cardinality

**Verdict:** Fits — Ready
**Rounds:** 2. **Reviewers:** maintainability (Grok 4.7), executability (Composer)

## To user
None.

## Writer's own calls
- D-1: PropertiesSnapshot over Observation wire cardinality (stale after Properties edit).
- D-7: introduce `PropertyCardinality` and point every Properties / FakeStore / `CatalogProperty` default cardinality literal at it (maintainability M-1, M-3).
- T-2 pins `likeness` (depiction / tile path) instead of a second text multiple.
- Plan **Branch** stays `s14-30-…` for ship-pr / implement-pr; this cloud docs branch is `jake/…-dc42` only for landing the plan.

## Round 1

### M-1 · blocker · Cardinality is parsed as a raw string beside the typed wire vocabulary
**Answer:** Fixed — D-2 / D-7 / STEP-1 / Codebase fit: `PropertyCardinality` beside `PropertyValueType`; `"cardinality"` vocabulary-test row; marks map is `[String: PropertyCardinality]`; PropertiesModel / PropertiesView / FakeStore use `.rawValue`. `ObservationPolarity` vocab row left Out.

### M-2 · major · The shipped multiple Properties are not the cases the plan checks
**Answer:** Fixed — Edges Real data lists all seven `seedProperties` multiples; AC-1.3 / T-2 use `likeness` (tile path); neighbors note depiction brackets.

### X-1 · major · Branch name vs cloud-agent branch template
**Answer:** Declined — Plan **Branch** is the implementer / ship-pr name (`s14-*-slug`), matching every other Spike 14 plan (e.g. S14-28). The cloud agent's `jake/…-dc42` branch is only for this docs PR. No change to the header.

### X-2 · minor · STEP-5 map row omits plan link
**Answer:** Fixed — STEP-6 sets the map Plan cell to the plan file link when landing.

### X-3 · minor · `paintRows` parameter order unspecified
**Answer:** Fixed — D-5 / STEP-3: after `projectDir`, before `painter`.

### X-4 · minor · D-4 empty catalog cardinality implicit
**Answer:** Fixed — D-2 / D-4: `PropertyCardinality(rawValue:) ?? .single` when building the map; empty normalizes once.

### X-5 · minor · STEP-4 `MARKETING_VERSION` scope
**Answer:** Fixed — STEP-5 bumps every `MARKETING_VERSION` in `project.pbxproj` to the same value.

### X-6 · minor · Checks not runnable in this dry-run environment
**Answer:** Logged — STEP-4 already names the Linux `rg` gate and macOS CI for `xcodebuild`; no plan shape change.

## Round 2

### Maintainability
### M-3 · blocker · Cardinality literals the new enum does not own
**Answer:** Fixed — D-7 / STEP-1 / Deletes / Out: every cardinality compare and assignment in `PropertiesModel.swift`, `PropertiesView.swift`, and `FakeStore.swift` (locked Holds, draft/`openCreate` defaults included) uses `PropertyCardinality.*.rawValue`; `CatalogProperty.cardinality` default uses the same; STEP-1 `rg` must print no `"single"` / `"multiple"` matches in those three files.

No further maintainability blockers. Verdict: Fits.

### Executability
Round-1 answers hold at `4cf944ca`. X-7 logged (toolchain / no Xcode here). Verdict: Ready.
