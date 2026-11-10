# Review — S15-03 The Source page shows the date and the Private switch (and brief S15-D1)

**Verdict:** Fits with changes — Ready with changes. Round 2 found 3 blockers and 10 majors in round-1 and board-validation text; all Fixed. Plan Planned 2026-11-10.
**Rounds:** 2, plus a board validation between them. **Reviewers:** maintainability (fresh context, session model), executability (fresh context, Sonnet)

## To user
- None beyond the plan's open question (how the agreed board reaches the implementer).

## Writer's own calls
- Clear date empties the form inside the dialog and Save confirms it, with no separate Confirm (D-9). This avoids a sheet over a sheet, and Cancel keeps the date.
- Every Source page section's write failure goes through `SourcePageContext.fail` (D-5), not only the date and switch, so the gone path is one door.
- The Person page moves onto the extended `leavesForSectionRoot(when:notice:)` (D-7).
- The shared `error.datevalues.invalid` copy changes for the composer too (D-15).
- The page uses a feature-local layout inside `PVPageColumns`' full-width mode instead of the kit's secondary (D-17); the header band keeps the content width at Wide (brief SD-W1).

## Round 1 — maintainability

- **M-1 · blocker · raw error-code strings against `macos-client.mdc`.** Fixed — D-4 adds `EngineErrorCode.sourceNotFound` and the vocabulary-test row; D-15 removes the date-code branch; FakeStore gets `sourceNotFound` (D-14).
- **M-2 · blocker · sections change parent at 1600 and lose drafts.** Fixed — D-17 `SourcePageOverviewLayout` over fixed subviews; T-9; the PreferenceKey loop is deleted.
- **M-3 · blocker · two "missing Source" answers in Go.** Fixed — D-3 maps once in `getOne`; `ReadDraft` and `ListWorkspace` pass it through; T-1 covers `GetTx` and `ReadDraft`; the stop condition greps `Get` and `GetTx`.
- **M-4 · blocker · wrong date formatter and strings.** Fixed — `DateRowDisplay.line(date:)`; ACs and the brief use its output ("abt 12 Mar 1880", "bet 1834 and 1840").
- **M-5 · major · the date dialog skips the page's dialog pattern.** Fixed — D-8 `SourceDateFormModel` + `.sourceDateDialog`.
- **M-6 · major · identity resets on any Source change.** Fixed — D-12 `resetIfChanged`; AC-2.4, T-6.
- **M-7 · major · gone handling should be one mechanism.** Fixed — D-5 `fail`, D-6 `gone`, D-7 `leavesForSectionRoot(when:notice:)`; `sourceGoneNotice` dropped for `L10n.Errors.sourcesNotFound`.
- **M-8 · major · header keeps its own card and rule; loading shell outside.** Fixed — D-16; brief SD-W1, SD-W2 and inventory.
- **M-9 · major · page-only date error copy.** Fixed — D-15 changes the shared copy.
- **M-10 · major · docs made false; Out line in an unrelated idea file.** Fixed — STEP-8 adds `macos-client-patterns.md` § 9 and `ideas/wide-layout.md`; Out lines go to `ideas/structural-debt.md`.
- **M-11 · minor · key prefix shared with the date form.** Fixed — D-18 `sourceDate*`; moving the form's keys logged as Out.

## Round 1 — executability

- **X-1 · blocker · an empty secondary still stacks.** Fixed — no secondary (D-16, D-17).
- **X-2 · blocker · strings created after the step that needs them.** Fixed — all copy in STEP-2.
- **X-3 · major · raw code; read error extraction; `initial`.** Fixed — D-4, D-6 (`if case .failed(let error)`), D-7 keeps `initial: true`.
- **X-4 · major · header and shell loose ends.** Fixed — D-16 (band chrome, loading shell, Callouts in primary, `scrollResetKey: sourceID`, header always passed).
- **X-5 · major · a date or switch write wipes other edits.** Fixed — D-12.
- **X-6 · major · Confirm over the FormDialog.** Fixed — D-9 removes the Confirm.
- **X-7 · minor · tests the plan cannot back.** Fixed — FakeStore hooks and `sourceNotFound` (D-14); T-7 covers cover and update; T-8 covers the read; "called once" ACs read the store; `pendingPrivate` named (D-11); `SourceDateClear` removed.
- **X-8 · minor · Go edges.** Fixed — `Delete`'s `Codes.NotFound`, `ListWorkspace` passthrough (D-3); `save` returns `fail`'s nil (D-5).
- **X-9 · minor · version step.** Fixed — STEP-8, 0.5.0 → 0.6.0.
- **X-10 · minor · Metadata narrower at Wide.** Fixed — Edges and manual check 6; risk row.
- **X-11 · minor · precedent kind mismatch.** Fixed — Codebase fit names `identity.ErrNotFound`.
- **X-12 · minor · meta row wrapping and Private's place.** Fixed — macOS row and brief SD-1.
- **X-13 · nit · `refreshCoverFromStore`, `metadataColumnHeight`.** Fixed — covered by D-5's 16 sites and D-17's Deletes.
- **X-14 · nit · incomplete Current state rows.** Fixed — Go and Page columns rows list the callers.
- **X-B1 · major · the preview's place at Regular.** Fixed — under Metadata in Metadata's column at both widths.
- **X-B2 · major · header row behavior.** Fixed — SD-1 (wrap, Private's place), SD-W1 (page error Callout).
- **X-B3 · minor · Reach-for table lacks Switch, Disclosure, ClipboardButton.** Fixed — named in the enhancement note (the Paste block stays verbatim).
- **X-B4 · minor · New row with no second call site.** Fixed — the preview row is Extend.
- **X-B5 · nit.** Logged.

## Board validation (2026-11-10)

The S15-D1 board ([Claude Design — Source Detail](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK)) was read against SD-1 – SD-14, SD-W1 and SD-W2. It draws every frame the brief asks for (Regular and Wide by a pane toggle on each frame) and records a decision per requirement. Where it differs from the plan:

- **Header band at Wide spans the group** (SD-W1), which the plan's Stop condition named. Fixed — D-16 adds a `fillMax` option to `PVPageColumns` full-width mode; T-10.
- **"Add publication date"** in place of "Add date". Adopted — Strings.
- **A "Reads as" strip** under the date form in the dialog. Adopted as a dialog-level snowflake; `DateValueEditorForm` is unchanged — macOS row, Codebase fit.
- **A titled Callout for a failed Private flip** ("Couldn’t change Private"). Adopted — D-11, `pageErrorTitle`.
- **A per-dialog discard body.** Not adopted: the kit's standard FormDialog discard copy stays — Strings.



## Round 2 — maintainability

- **M2-1 · blocker · `pageErrorTitle` beside `pageError` races.** Fixed — D-19: one `PageError` value, written only through `showPageError` / `clearPageError`; `apply` calls `clearPageError()`; T-6 asserts the title is cleared by the next error.
- **M2-2 · major · `fillMax` hand-sums `split`.** Fixed — D-16 renames it `fillSplitSideMax` and takes primary and band from `PVPageColumnsMetrics.split(…).group`; T-10.
- **M2-3 · major · guidance made false by the kit option.** Fixed — STEP-8 and the DoD: `macos-client-patterns.md` § 9 rule 2 with `add-design-brief`'s Paste block, `pvKeepsPageBand()`, `DesignSystem/README.md`, the doc comments.
- **M2-4 · major · the page error spans the group at Wide.** Fixed — D-17 subview E in the main column at both widths; D-16 keeps only loading and load error in `primary`; T-9.
- **M2-5 · major · "Reads as" as raw styling.** Fixed — `PVCard(tone: .sunken, cornerRadius: PVRadius.sm)` + `pvMicroCaps()`; hidden while the draft is blank, invalid or clearing.
- **M2-6 · major · brief rows and ACs state pre-board copy.** Fixed — brief SD-2, SD-3, SD-6, SD-8, SD-W1, § 4.1, frames 2–3, inventory (Page columns Extend); plan ACs and manual checks say "Add publication date" and the titled Callout.

## Round 2 — executability

- **X2-1 · blocker · the read-path notice is lost on mount.** Fixed — D-6 `goneNotice(read:)` is pure and read in the same body pass as `isFailed` (D-7); T-8.
- **X2-2 · blocker · STEP-4 needs STEP-5's `fail`.** Fixed — the gone door and `PageError` are STEP-4; the header and date form models are STEP-5.
- **X2-3 · major · Callouts at Wide.** Fixed — as M2-4.
- **X2-4 · major · date form edges.** Fixed — D-8 defines `clears` reset on typing, `isDirty`, `canSave`, and that `pvFormDialog(item:)` clears the binding; T-4 asserts on the store, not on `dateForm`.
- **X2-5 · major · "Add date" vs "Add publication date".** Fixed — as M2-6.
- **X2-6 · major · layout interface.** Fixed — D-17: `init(widthClass:)`, internal, wrapper reads the environment, subviews and gaps named; T-9 hosting pattern named.
- **X2-7 · major · `fail` returns optional.** Fixed — D-5's rule: on nil, return without writing state.
- **X2-8 · minor · stale "board is not agreed".** Fixed — removed.
- **X2-9 · minor · `fillMax` sites.** Fixed — D-16 lists the overload, the band layout, the three inits, the doc comments, the existing test.
- **X2-10 · minor · title clearing, AC-2.3, T-6, `resetIfChanged`.** Fixed — D-19, AC-2.3, T-6 calls `sync`, D-12 states `from == nil` and what `reset()` leaves.
- **X2-11 · minor · `sourcecopy_test.go` "unknown source".** Fixed — named in §8's likely-to-break line.
- **X2-12 · minor · meta row wording, date text style, delete after deletion.** Fixed — macOS rows follow the board; delete-impact on a deleted Source is an Out line to `ideas/structural-debt.md`.
- **Brief (executability).** Fixed — as M2-6.
