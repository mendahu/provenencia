# Review — S14-14 Copy and VoiceOver on graph work

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2, plus round 3 scoped to the D-19 revision (2026-11-01). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None open. Answered in round 2 (2026-11-01): M-13 — move the composer's held step onto `PVDiscardGate` in this PR (D-22). Answered before review (2026-11-01): one PR rather than a split (the plan bends the scope entry's split line; STEP-13 rewords it); all three ship-log leftovers in; no brief; history rows read graph name then surface; the composer's load error sits under the Citation picker.

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

## Round 2

### M-13 · blocker · `PVDiscardGate` beside the composer's `stepContinuation`
**Answer:** To user → answered (2026-11-01): move the composer in this PR. D-3 narrows the gate to the held step only, since each owner's question stays in its own presentation (the composer's in `presented`); D-22 replaces `stepContinuation`; AC-2.13, T-19, manual check 8; Codebase fit row names it.

### M-14 · blocker · Card switch stored twice; `beginEdit` still silent
**Answer:** Fixed — D-12: `beginEdit` is the asking entry and `openEdit` is private; the target lives only in the held step; `PendingEditSwitch` is an id-only question value, as `PendingLeave`.

### M-15 · blocker · Third copy of the Citation ref lookup
**Answer:** Fixed — D-16 adds `citationRef(for:)`; `activeCitationRef` and `announceIdentityChange()` call it.

### M-16 · blocker · Hand-written on-change post in a file STEP-2 edits
**Answer:** Fixed — D-8 moves the cut-failed caption onto `pvAnnounce`; § 3 Out keeps only `NameValueEditorForm`, which posts after an action.

### X-18 · major · `StoreBoom` is private
**Answer:** Fixed — AC-7.1, AC-7.3 and T-17 assert by prefix.

### X-19 · major · When `citationLoadError` clears
**Answer:** Fixed — D-16 clears it in `openBlankCitation()` and at the start of `startIdentitySwitch`; AC-7.9.

### X-20 · major · Where `PVInput` applies the label; nested fields
**Answer:** Fixed — D-10: on the outermost view of `PVInput` and `PVTextArea`; `PVField` always sets the value, nil without a label; several controls in one field share it (accepted).

### X-21 · major · T-15 depends on the host's PDF app
**Answer:** Fixed — D-17 moves the app lookup into `ProjectFiles.launch`, so the seam covers it; T-15 also checks a refusal passes through.

### X-22 · minor · How the Go test gets the kinds
**Answer:** Fixed — D-20 / T-18 collect `Table.Kind` over `catalogmodel.Tables`.

### X-23 · minor · Save-page line speaks for any `fields.error`
**Answer:** Fixed — D-8 states it and accepts it.

### X-24 · minor · Two `rows` callers
**Answer:** Fixed — D-14 names both callers and the new `HistoryJumpRow` initializer.

### X-25 · minor · Toasts on onboarding and bridges now speak
**Answer:** Fixed — D-8 and manual check 5 accept it.

Smaller notes (README sentence wraps, no Announce row to extend, about 43 `PVField` sites, a 21st non-plural count): Fixed in STEP-1, STEP-2, § 4 and § 6 Edges.

## Round 3 — D-19 revision (2026-11-01)

Why: ship round 0 stopped at STEP-9's probe. A top-level `variations.plural` whose forms carry no number is a hard catalog compile error ("Plural variation requires referencing the number in the string"), for all 18 kind keys. User chose (2026-11-01): probe a substitution, then revise. A scratch catalog compiled with `xcstringstool compile` (exit 0; format emitted as `%1$#@noun@`) and resolved "citation" for 1, "citations" for 0, 2, 5; `deleteImpact.via.unknown`'s shape compiled in the same run. D-19 now uses a `%#@noun@` substitution. Reviewers saw only the plan diff: maintainability (Opus), executability (Sonnet). The two-round cap is spent, so no round 4; the two blockers were missing detail, both fixed.

### X-1 · blocker · The catalog entry's JSON is not written out
**Answer:** Fixed — § 6 Strings carries the full `deleteImpact.kind.citation` entry; the forms carry no `%arg`, unlike `rowSummary`.

### X-2 · blocker · Current state is not stated: the working tree holds the rejected shape
**Answer:** Fixed — STEP-9 opens with a resume starting point: `bc8f22ff`, commit STEP-7/8 first, convert the 18 entries in place, `citation` first.

### X-3 · major · How `noun(_:count:)` passes the count is only implied
**Answer:** Fixed — D-19: signature and one-argument `L10n.format` call kept; count reaches `noun` through `argNum` 1.

### X-4 · major · T-13's skip: why it still works
**Answer:** Fixed with M-2 — the prefix skip is replaced; the compiled format is `%1$#@noun@` (probe output), which `PluralShape.position(ofVariable:in:)` matches.

### X-5 · minor · The script accepts the shape, unstated
**Answer:** Fixed — § 6 Edges Assumptions says the script needs no change, and why.

### X-6 · minor · Stop condition has no direction
**Answer:** Fixed — "Stop and report; do not switch to separate one / other keys."

### X-7 · minor · `promotedBody` reference may be copied for nouns
**Answer:** Fixed — § 4 A-20 row says that shape must print its number, so not for a bare noun. The verb/noun pairing note is logged as a nit.

### M-1 · major · The skill does not teach the bare-noun shape or the compile trap
**Answer:** Fixed — STEP-13's `add-localized-string` edit gains the substitution rule and the compile-error trap.

### M-2 · major · The plural test exempts a key prefix
**Answer:** Fixed — D-19 / T-13: `PluralShape.printsNumber` *(new)* gates the digit checks; the `deleteImpact.kind.` prefix test goes.

### M-3 · minor · `defaultValue` for the kind resources unstated
**Answer:** Fixed — D-19: `defaultValue: "%#@noun@"`, matching the catalog value per `add-localized-string` step 5.

### Writer's own calls (round 3)
- `defaultValue` mirrors the catalog (`%#@noun@`) rather than holding the English singular.
- On a failed in-app probe the implementer stops rather than falling back to separate keys.
