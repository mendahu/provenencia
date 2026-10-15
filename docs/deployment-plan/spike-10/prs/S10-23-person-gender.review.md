# Review — S10-23 Sex / gender on the Person page (with brief S10-D12)

**Verdict:** Fits with changes — Not ready (round 1, on the draft blocked on its board). Every finding is answered; round 2 runs when the S10-D12 board is agreed and the colour tokens are in D-5.
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet). Both reviewed the plan and the brief together.

## To user
- None open. The colour tokens and the board link are the plan's Open questions, answered by the agreed board.

## Writer's own calls
- **Badge, not Chip** (D-8). The user picked "Chip + thumb tint". `PVChip` is a selectable control in this kit; the display-only pill is `PVBadge`, and it draws its text uppercase.
- **Portrait follows the Source cover pattern** (brief PDP-10, PDP-11, from M-5). There is no "Clear portrait" button on the tile. Going back to the earliest likeness is on the header thumb's menu, like the Source page's cover revert.
- **`likeness` is not a Details row** (brief PDP-7, from M-6). The gallery is the only place it is drawn. This ships in S10-35.
- **Spoken label is "Gender, Woman"** (D-9, from M-4). It uses the row's `a11yList` shape rather than new "Gender: %@" copy.
- **Image-state colours deferred to S10-35** (D-6, from X-7). This PR exercises only the mark state.

## Round 1

### M-1 · blocker · [plan] Tint on `PVThumbnail.Content` is wiped by `CachedThumbnail` in S10-35
**Answer:** Fixed. D-6 makes `colors:` a `PVThumbnail` init parameter beside `size` and `cornerRadius`, and leaves `Content` unchanged. The `CachedThumbnail` pass-through goes in S10-35 (§3 Out, Next adds), since nothing here renders a cached thumb.

### M-2 · major · [plan + brief] No real slot for the S10-35 and S10-39 header thumb or the gallery
**Answer:** Fixed. D-7 adds one member, `ConclusionDetailBody.thumbnail: ConclusionDetailThumbnail`, and S10-35 adds the crop to it. The brief's Shared Conclusion page row no longer promises `sections`; it says S10-35 adds a slot below Details.

### M-3 · major · [plan] Term keys hidden from the wire-vocabulary file and its drift test
**Answer:** Fixed. D-10 adds `SeededGenderTerm` and `SeededSexAtBirthTerm` to `ConclusionVocabulary.swift`, plus T-7, a subtest that looks each key up with `propertyterms.Lookup`.

### M-4 · major · [plan + brief] A second way to speak a field
**Answer:** Fixed. D-9 speaks `a11yList(field.label, rest: text)`. The two a11y keys are dropped, and brief PDP-5 is reworded.

### M-5 · major · [brief] Portrait re-solves the Source cover interaction
**Answer:** Fixed. Brief PDP-10, PDP-11, and frame 10 now follow the cover pattern: an accent Portrait Badge with the photo icon, the button hidden on the chosen tile, a toast, and "Use earliest likeness" on the header thumb menu. The inventory rows were added.

### M-6 · major · [brief] Tile names only `PVThumbnail`; `likeness` would also show as a Details row
**Answer:** Fixed. The inventory now has `CachedThumbnail`, the artifact preview button shell, and `ReconciliationRecord` as the tile data. PDP-7 says tiles are the `likeness` field's records and that `likeness` is never a Details row.

### M-7 · minor · [plan] "Tint" means three things
**Answer:** Fixed. The names are now `PersonGenderTreatment.Family`, `PVThumbnail.Colors { tint, line }`, and `Family.tint/line/ink`, after `EvidenceSubjectKindStyle`, which is listed in §4.

### M-8 · minor · [plan] Hand-built tinted tiles in Promote could move onto `PVThumbnail`
**Answer:** Declined for this PR and logged in §3 Out as not planned. The mark sizes differ, so it is not a drop-in change. Raise it with `evaluate-ui-component` if it bites.

### X-1 · blocker · Checks cannot run on Linux
**Answer:** Fixed. D-11 and a Stop condition say the Swift checks and the manual check run on a Mac; elsewhere the implementer runs the Go checks, stops, and reports which checks did not run (S10-41 precedent). The DoD box was updated.

### X-2 · major · Badge look not decided
**Answer:** Fixed. D-8 is `tone: .neutral, subtle: false`, no icon; the implementer stops if the board differs. Brief PDP-3 says the same.

### X-3 · major · `Tint` shape not declared
**Answer:** Fixed. D-5 now gives full signatures.

### X-4 · major · Non-term or empty displayed value
**Answer:** Fixed. D-2 says either is neutral, with no fallback for gender. T-4 has rows for both.

### X-5 · major · Neutral-term stop condition has no check
**Answer:** Fixed. The Stop condition now names `core/autoreconcile/modules_test.go` `TestModules`, and §4 cites it.

### X-6 · major · Manual check has no recipe, and the case and VoiceOver are inconsistent
**Answer:** Fixed. §10 gives the composer and Promote recipe, notes that `PVBadge` draws uppercase, and adds VoiceOver to step 1. The ACs say "drawn uppercase".

### X-7 · major · Image-state tint untested and unreachable
**Answer:** Fixed. D-6 applies colours to the mark and symbol states only; the image state moves to S10-35 (§3 Out).

### X-8 · major · STEP-3 `rg` check does not test edits
**Answer:** Fixed. STEP-3 now uses `git diff --stat -- macos/App`.

### X-9 · major · Brief and plan disagree on archiving the brief
**Answer:** Fixed. Brief §7 now says S10-39 archives it.

### X-10 · major · Board line omits ink; Claude Design not told to name tokens
**Answer:** Fixed. Brief §2 and the new PDP-12 require labelled tint, line, and ink tokens, and §7 step 1 records ink.

### X-11 · major · Brief inventory missing rows for the later frames
**Answer:** Fixed with M-6. PDP-7 fixes the tile at 96pt, the count as the SectionHeader aside, and the caption. Frame 10 shows the Portrait Badge without a gender Badge conflict, since they are on different elements.

### X-12 · minor · Lint warning on the New row
**Answer:** Declined. `check-plan-docs.py` warns on every New row. The row states it is a single-call-site feature snowflake, not a kit primitive, which `add-design-brief` allows. S10-D9 carried the same warning.

### X-13 · minor · New files not registered in pbxproj
**Answer:** Fixed. STEP-1 now registers both files, and §4 has an Xcode project row.

### X-14 · minor · Path missing a directory
**Answer:** Fixed. §4 Badge row now has the full path.

### X-15 · minor · Smaller choices
**Answer:** Fixed:
- The STEP-1 check runs `go test ./api/ffi/...`.
- The init order is fixed in D-6.
- `PVBadge` already makes itself one accessibility element (D-8).
- `indeterminate` is in AC-2.4 and T-4.
- T-6 checks the Event and Place defaults.
- The no-op Stop line was removed.
- STEP-6 names the README row and the map wording.
- Brief frame 8b draws Unknown at birth.
