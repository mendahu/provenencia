# Review — S14-14 Copy and VoiceOver on graph work

**Verdict:** pending round 2
**Rounds:** 1 of 2 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None open. Answered before review (2026-11-01): one PR rather than a split (the plan bends the scope entry's split line; STEP-13 rewords it); all three ship-log leftovers in; no brief; history rows read graph name then surface; the composer's load error sits under the Citation picker.

## Writer's own calls
- `PVDiscardGate` is a kit type shared by sheets and in-page edits (M-9), not a sheet-only state.
- `PVCopy` gains a `pvString` requirement so a callout can announce its title (X-2).
- A second identical announcement while a view stays mounted is silent (D-6, X-13).
- `identity_claim` gets the noun "identity claim"; `narrative` and `reconciliation_claim` nouns are dropped as not Go kinds (M-4).
- The composer error line sits under the whole stepper row in the panel layout so the bottom-aligned steppers do not move (X-6).
- The history row's ref column keeps the location's ref, so a composer row shows its Citation ref twice (D-14).
- `.notOpened` copy: "macOS couldn't open “%@” in its app."

## Round 1

### M-1 · blocker · An Out line keeps a hand-made announcement for a `PVErrorText`
**Answer:** Fixed — D-8 opts `citationComposer.depiction.savePageError` in and drops `savePage()`'s error post; the Out row is gone; AC-7.8, manual check 5, STEP-2 check.

### M-2 · blocker · D-19 hand-rolls a focus hand-back beside `FocusClaim`
**Answer:** Fixed — D-18 uses `FocusClaim.claim` for the close hand-back and moves the combo's `.onAppear` onto it; § 4 precedent is `FocusClaim`; `PromoteAlignmentRow` logged for S14-20 (§ 3 Out).

### M-3 · blocker · `PVFieldLabel.resolve` beside `PVSelect.accessibilityLabelText`
**Answer:** Fixed — D-10: `resolve(own: String?, field: String?)`; `PVSelect.accessibilityLabelText` calls it with its explicit or spoken label as `own`; `PVInput` / `PVTextArea` call it with `own: nil`.

### M-4 · blocker · Kind nouns from a Swift-only list with no parity test
**Answer:** Fixed — D-19 makes `kindNouns` a declared table over Go's 17 kinds plus `item`, drops `narrative` / `reconciliation_claim`, adds `identity_claim`; D-20 / T-18 add the parity subtest; header and § 6 Go core name the one Go test change.

### M-5 · major · Three places build the name lookup; canvas reaches into toolbar statics
**Answer:** Fixed — D-13: `WorkspaceLocationLabels.swift` holds `LocationNames(session:)`, `labels`, `graphLabel`, `surfaceLabel`; toolbar, history host and canvas model use them.

### M-6 · major · `PVAnnouncement` second type; hard-coded join
**Answer:** Fixed — D-6: rules on `AccessibilityAnnouncement`; join through `designSystem.announcement.join`; `pvAnnounce` in `DesignSystem/Components/Announce/`.

### M-7 · major · Two `ObjectRefusal` mappings; three test seams
**Answer:** Fixed — D-17: `L10n.ObjectOpen.message(for:filename:missing:)`; one `launch` seam on `openObject`; T-16 tests the mapping.

### M-8 · major · Card-switch confirm missing from `acceptsArtifactPageDrop`
**Answer:** Fixed — D-12 adds the term; AC-2.9 asserts it.

### M-9 · major · Second hold-then-discard state
**Answer:** Fixed — D-3 `PVDiscardGate` serves the sheet (D-4) and the card (D-12); Notes and Metadata logged (§ 3 Out).

### M-10 · major · `WorkspaceLeaveGuard` missing from Codebase fit
**Answer:** Fixed — § 4 row, Codebase fit row, Builds on, and the README line in STEP-5 say how the two layers split.

### M-11 · minor · `activate` doc comment and the notice's two close paths
**Answer:** Fixed — D-5: the notice's Done calls `pvSheetClose()`, `onDismiss` goes; doc comment rewritten; the old signature is in the DoD sweep.

### M-12 · nit · Other hand posts on change
**Answer:** Logged — § 3 Out row.

### X-1 · blocker · `PVFieldLabel.resolve` signatures disagree
**Answer:** Fixed — D-10 (String in, String out); how `PVInput` / `PVTextArea` apply it is stated.

### X-2 · blocker · `PVCallout` has no title String
**Answer:** Fixed — D-7 adds `PVCopy.pvString`; the callout keeps its `(any PVCopy)?`; AC-3.9.

### X-3 · blocker · T-17 does not make the load throw
**Answer:** Fixed — T-17 invalidates `.citationsByArtifact` first, as the existing test does. The named case is reached through `loadedCitationRef` of the active Citation (D-16), since a listed Citation is served from the cache and never throws.

### X-4 · major · AC-7.3 has no valid test
**Answer:** Fixed — AC-7.3 uses `selectCitation("cit-missing")`; `loadFailureUnlistedIsUnnamed` is its own test; the existing test now expects the named copy (AC-7.1).

### X-5 · major · `showingLoadFailure` callers' ids
**Answer:** Fixed — D-16 lists each caller's id.

### X-6 · major · Error line placement
**Answer:** Fixed — D-16: `citationLoadErrorLine` with `if let`, `.micro`; side form under the stepper, panel form under the stepper row; § 6 macOS notes Wide is unchanged.

### X-7 · major · Unnamed second helper
**Answer:** Fixed — D-13 names `surfaceLabel(for:)` with `promoteAll`; AC-5.4 covers both Promote forms; the entity branch is stated.

### X-8 · major · `canvasAccessibilityLabel` type
**Answer:** Fixed — D-15 returns `String`; the view wraps it in `Text(verbatim:)`; AC-5.7 adds the armed case.

### X-9 · major · `.pvConfirm` signature
**Answer:** Fixed — D-12 gives the binding, tone, prefix, and wires `cancelEditSwitch` through the binding's setter.

### X-10 · major · Sheet close API gaps
**Answer:** Fixed — D-3 states each answer's effect on the pending step; D-4 states the closure types, the preview default, rebuild per body pass, Escape while asking (AC-2.8), and a new item.

### X-11 · major · Plural without a number unproven
**Answer:** Fixed — STEP-9 probes one key first; D-19 states the prefix the plural test skips and the substitution shape for `deleteImpact.via.unknown`.

### X-12 · major · `openObject` isolation
**Answer:** Fixed — D-17: `openObject` and `launch` are `@MainActor`; the existing tests `await`; Stop condition covers strict-concurrency warnings.

### X-13 · major · `pvAnnounce` edges
**Answer:** Fixed — D-6 accepts the repeated-text case and states the toast and ordering behavior.

### X-14 · minor · Pencil focus
**Answer:** Fixed — D-18 places the `.onChange` on the `metaRow` `HStack`; `.focused` goes on the `PVIconButton` (a `Button`); `FocusClaim` covers the remount.

### X-15 · minor · Script path
**Answer:** Fixed — D-21: `main(catalog:)`, the message prints the given path, every check runs on it; T-14 requires the output to name `test.planted`.

### X-16 · minor · Map link
**Answer:** Fixed — the map row links the plan with this review's commit; STEP-13 gives the exact text for the Form fields part.

### X-17 · minor · Smaller choices
**Answer:** Fixed — `HistoryJumpMenuHost` is under `WorkspaceView`'s `.environment(session)` (§ 4); every opted-in callout is a `PVCallout` (§ 4); the two opted-in `PVErrorText`s are named (D-8); README wording given for every row (STEP-2); D-2 names the initializer; the duplicate ref is accepted (D-14); the header says Spike 12 and Spike 11 leftovers.
