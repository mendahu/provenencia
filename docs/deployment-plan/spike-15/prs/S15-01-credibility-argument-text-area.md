# S15-01 — The credibility argument is a text area

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change beyond the control: the argument becomes the `PVTextArea` the Source page already draws for its description and notes; working rule "The credibility fix is idiomatic and needs no brief")
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s15-01-credibility-argument-text-area`
**Version bump:** patch — a UI bugfix on the Source page, no format or wire change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (compose the existing kit control; no new primitive), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (one table case in an existing handler test)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "The credibility fix is idiomatic and needs no brief" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [note](../2026-10-18-credibility-field-truncates-and-ignores-enter.md); [`research-judgment-model.md`](../../../research-judgment-model.md) § 2 (the assessment row)
**Planned:** 2026-11-09 against `main` @ `eab4453`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

On the Source page, the credibility argument is a single-line field: a longer argument scrolls sideways and is cut off, and Return does nothing. After this PR the argument is a multi-line `PVTextArea` like the Source description and the note composer: it wraps, grows with its text up to a bound, then scrolls; Return adds a line; the existing Save button saves it, line breaks included.

**Why now.** Dogfood note [2026-10-18](../2026-10-18-credibility-field-truncates-and-ignores-enter.md); first PR of slice 1, no dependencies.

**What stays.** The grade chips, the Save and Cancel buttons and their enabling rule, the saved/unsaved status text, the section header, the model (`SourceCredibilitySection`), the engine write, and the stored column. The note's "Enter saves" wish is answered by the working rule: Return is a new line, Save is the button.

## 2. User stories

**US-1.** As a researcher, when I write a long argument for a Source's credibility grade, I see all of it while I write and after I save, so that I can read and correct it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source page; type a 500-character argument (§ 10 gives the text) into the argument field at the default window width | The text wraps; every character is visible without horizontal scrolling; the field is taller than one line |
| AC-1.2 | The argument field is empty and unfocused | The field shows the prompt "Why this grade — optional" and is two lines tall |
| AC-1.3 | An argument longer than 12 lines at the field's width | The field stops growing at 12 lines and its text scrolls vertically inside it |

**US-2.** As a researcher, I can break my argument into paragraphs and save it with the button I already use.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Focus in the argument; press Return | A new line is inserted; nothing is saved; Save becomes enabled |
| AC-2.2 | Argument `"Census taker was a neighbour.\nAges rounded to 5."`; press Save | The assessment saves; reopening the Source shows the same two lines |
| AC-2.3 | Argument `"  Line one\nLine two\n\n"` saved | Stored and shown as `"Line one\nLine two"`: ends trimmed, inner line break kept |
| AC-2.4 | Edit the argument; press Cancel | The field returns to the saved argument, line breaks included |
| AC-2.5 | Focus in the argument; press Tab | Focus moves to the next control (the Save button when enabled); no tab character is inserted |

## 3. Scope

| In | Covered by |
| --- | --- |
| Replace `PVInput` with `PVTextArea` for the argument in `SourcePageCredibilityView` | STEP-1, AC-1.1 – AC-1.3, AC-2.1, AC-2.5 |
| Keep the field's accessibility identifier and give it a spoken name | STEP-1, D-4 |
| Pin, in tests, that a multi-line argument round-trips through the engine and the model | STEP-2, STEP-3, AC-2.2 – AC-2.4 |
| Patch version bump | STEP-4 |
| Docs: map status, `completed.md`, this plan's status | STEP-5 |

| Out | Home |
| --- | --- |
| A shared default-action shortcut (Return or ⌘Return saves) for every `PVInlineEditActions` editor | not planned — the working rule chose platform text-area keys; reopen as an idea file only if dogfood asks again |
| Return saves the argument (the note's second wish) | not planned — refused by the working rule (user, 2026-11-09) |
| Removing the view's own `isDirty` in favor of `SourceCredibilitySection.isDirty` | not planned — unrelated tidy; the view's copy exists because the draft is local |
| A Source credibility argument length limit | not planned — the column is unbounded `TEXT` today and stays so |
| Any change to the description, notes or metadata editors | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Argument field | `macos/App/Features/Sources/SourcePageCredibilityView.swift` `SourcePageCredibilityView.body` | `PVInput(text: $argumentLocal, size: .sm, prompt: L10n.Sources.credibilityArgumentPlaceholder)` with `.accessibilityIdentifier("sources.page.credibility.argument")`. One line; long text scrolls sideways; Return does nothing (no `.onSubmit`). Save and Cancel are `PVInlineEditActions` below it; `isDirty` compares trimmed `argumentLocal` with `model.credibility.savedArgument` |
| Local draft | same file, `@State argumentLocal`, `.onAppear`, `.onChange(of: model.credibility.argumentDraft)` | Typing stays local; Save copies it into `model.credibility.argumentDraft`. Unchanged by this PR |
| Placement | `macos/App/Features/Sources/SourcePageView.swift` `overviewColumns` | Left column: description, `Spacer`, credibility pinned to the bottom; column `minHeight` tracks the metadata column. A taller argument grows the column; no layout change needed |
| Kit control | `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` `PVTextArea` | `TextEditor`-backed; `lineLimit` sets min/max height in lines, then scrolls; Return inserts a newline; Tab / Shift-Tab move focus, Option-Tab inserts a tab; `prompt` is drawn as an `accessibilityHidden` overlay, so it is not spoken; `.pvFieldNamed()` speaks only an enclosing `PVField` label |
| Sibling uses | `SourcePageDescriptionView.swift` (`lineLimit: 3...12`, `.bodyLarge`, inside `PVInlineEdit`); `SourcePageNotesView.swift` composer (`lineLimit: 2...8`, default `.body`, a `PVButton` below) | The notes composer is the same shape as the credibility block: an always-visible area with a button row beneath |
| Font today | `macos/App/DesignSystem/Components/Input/PVInput.swift` (`size.font` for `.sm`) | Small body type, matching `PVTextArea.Typography.body` (`PVTypeScale.bodySmall`) |
| Model | `macos/App/Features/Sources/SourceCredibilitySection.swift` `save()` | Trims the argument (`whitespacesAndNewlines`) and calls `store.upsertSourceCredibilityAssessment`. Inner newlines pass through |
| Engine | `core/database/sourcecredibility/assessments.go` `Upsert` | `strings.TrimSpace(in.Argument)`; stores `argument` as nullable `TEXT`. Inner newlines are stored as is |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources.credibilityArgumentPlaceholder` = "Why this grade — optional" (key `sources.page.credibilityArgumentPlaceholder` in `macos/App/Resources/Localizable.xcstrings`) | Reused; no new key |
| Tests | `api/ffi/handlers/credibility_test.go` `TestSourceCredibility`; `macos/ProvenenciaTests/SourcePageModelTests.swift` `selectCredibilityDraftDoesNotPersistUntilSave`, `cancelCredibilityRestoresDraftFromSaved` | Single-line arguments only; nothing pins line breaks |
| FakeStore | `macos/App/Platform/FakeStore.swift` `upsertSourceCredibilityAssessment` | Stores the `argument` it is given in `credibilityBySource` |

**Precedent.** The note composer in `SourcePageNotesView.swift`: a `PVTextArea(text:lineLimit:prompt:)` with `.accessibilityIdentifier` on the area and the action button in the row below. Copy its call shape; do not copy `PVInlineEdit` from the description (the credibility block has no resting/editing swap).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVTextArea(text: $argumentLocal, lineLimit: 2...12, prompt: L10n.Sources.credibilityArgumentPlaceholder)` with the default `.body` typography | Working rule: the description's control. `.body` is the same size as today's `PVInput(size: .sm)` text, so only the height changes. Lower bound 2 keeps the empty field close to today's height; upper bound 12 is the description's, so a 500-character argument fits unscrolled in the half-width column | `.bodyLarge` — changes the type size, a visible change the rule did not ask for; `2...8` (the notes composer) — 500 characters at half width reaches about 8 lines, too close to the bound for AC-1.1 |
| D-2 | Return inserts a newline (`PVTextArea`'s default); no `.onSubmit`, no key routing, no shortcut | Working rule "platform's text-area behavior (Return is a new line) and the existing Save button" (user, 2026-11-09) | Return-to-save with Shift-Return for a newline — refused by the rule |
| D-3 | No `focus:` handle is passed | `PVTextArea` without a handle moves Tab to the next control in the key-view loop (the Save button when enabled), which AC-2.5 asks for. There is no `PVInput` in this block to pair with `PVTabPair` | Wrapping in `PVTabPair` — that is for an input and an area in one form |
| D-4 | The area gets `.accessibilityLabel(Text(L10n.Sources.credibilityArgumentPlaceholder))` after `.accessibilityIdentifier("sources.page.credibility.argument")` | Today the `TextField` prompt names the field to VoiceOver; `PVTextArea` hides its prompt from accessibility, so without a label the field would lose its name. Reusing the existing string adds no copy | A new key "Credibility argument" — new copy for one label, and the prompt already says what the field is for; wrapping in `PVField` — adds a visible label the page does not have (chrome change) |
| D-5 | The accessibility identifier stays `sources.page.credibility.argument` | Nothing else changes about the field's identity | — |
| D-6 | Tests are one Go table case and one Swift model test; the view swap is checked by hand | The view has no logic of its own to test; the risk is that line breaks are lost on the way to the catalog, which the two tests pin. Swift view snapshots are not a repo practice | A UI test target — none exists in `macos/` |
| D-7 | Patch bump to `0.4.2` | A Mac UI bugfix ([`docs/versioning.md`](../../../versioning.md) "What bumps") | none — "docs-only" does not apply |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Table case "multi-line argument keeps inner line breaks" in `TestSourceCredibility` *(new)* | test case | `api/ffi/handlers/credibility_test.go` "upsert by grade id then workspace" | Extends that table; copies its `reqFn` fixture shape |
| `SourcePageModelTests.saveCredibilityKeepsLineBreaks` *(new)* | test | `SourcePageModelTests.cancelCredibilityRestoresDraftFromSaved` | Same file, same `makeStore` / `makeModel` helpers |

Searched for: another multiline field component (`TextEditor`, `axis: .vertical` in `Features/`), an existing credibility test with line breaks, an accessibility-label helper for `PVTextArea` (`pvFieldNamed`, `PVFieldLabel`). Only `PVTextArea` fits; `pvFieldNamed` needs a `PVField`, which D-4 rejects.

**Builds on.** `PVTextArea` (kit, `docs/design-system-layers.md` design-system layer); the notes composer call shape.
**Deletes.** The `PVInput(… size: .sm …)` call in `SourcePageCredibilityView`. No callee loses its last caller (`PVInput` has many). No string is removed.
**Next adds.** None named.

### Reach

None — no new path into existing code; the argument can already hold line breaks through the engine (it trims only the ends), so a stored multi-line argument is not a new state, only one the UI can now write and show.

### Scale

None — no write, list or loop changes; one control on one page.

### Edges

None — no rule, comparison or algorithm changes. The trim at both ends (Swift `save()`, Go `Upsert`) is existing behavior; T-1 and T-2 pin that it keeps inner line breaks.

### Schema

None — no catalog change.

### Go core

None — no code change; one test case (T-1).

### FFI

None — the request already carries `argument` as a string.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Credibility argument | `PVTextArea` (D-1) | `macos/App/Features/Sources/SourcePageCredibilityView.swift` | Replaces `PVInput`; same binding `$argumentLocal`; modifiers: `.accessibilityIdentifier("sources.page.credibility.argument")`, `.accessibilityLabel(Text(L10n.Sources.credibilityArgumentPlaceholder))` |
| Save / Cancel, status text | `PVInlineEditActions`, `Text` | same file | Unchanged |

States: empty → prompt, two lines tall; long → grows to 12 lines then scrolls; saving → unchanged (`PVInlineEditActions.isSaving`). The area is not disabled while saving (today's `PVInput` was not either). Accessibility: identifier kept; label per D-4; Tab leaves the area (D-3).

### Strings

None — no new or removed key. `L10n.Sources.credibilityArgumentPlaceholder` is reused as the prompt and the accessibility label.

## 7. Steps

**STEP-1 — Swap the argument field to a text area**
Files: `macos/App/Features/Sources/SourcePageCredibilityView.swift`. Skill: `add-ui-component`.
Do: replace the `PVInput(text: $argumentLocal, size: .sm, prompt: …)` call with `PVTextArea(text: $argumentLocal, lineLimit: 2...12, prompt: L10n.Sources.credibilityArgumentPlaceholder)` (D-1), keep `.accessibilityIdentifier("sources.page.credibility.argument")`, add `.accessibilityLabel(Text(L10n.Sources.credibilityArgumentPlaceholder))` (D-4). Change nothing else in the file.
Check: the app builds with no new Swift warning (`xcodebuild test` in STEP-3 builds it); `rg -n "PVInput" macos/App/Features/Sources/SourcePageCredibilityView.swift` prints nothing.

**STEP-2 — Pin multi-line arguments in the engine**
Files: `api/ffi/handlers/credibility_test.go`. Skill: `add-ffi-handler`, `.cursor/rules/go-tests.mdc`.
Do: add T-1 as a second case in the `runRPC(t, UpsertSourceCredibilityAssessment, …)` table, built like "upsert by grade id then workspace".
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run TestSourceCredibility` passes.

**STEP-3 — Pin multi-line arguments in the model**
Files: `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: add T-2 beside `cancelCredibilityRestoresDraftFromSaved`, using the file's `makeStore` / `makeModel` helpers.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/SourcePageModelTests` passes.

**STEP-4 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`, Debug and Release). Rule: `.cursor/rules/versioning.mdc`.
Do: `0.4.1` → `0.4.2` in all three (D-7). If `main` has moved past `0.4.1`, take the next patch of what `main` has.
Check: `cat VERSION`, `rg -n "Version = " core/version.go`, `rg -n "MARKETING_VERSION" macos/Provenencia.xcodeproj/project.pbxproj` all show the same version.

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done. The note `2026-10-18-credibility-field-truncates-and-ignores-enter.md` stays; spike close handles it ([`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md)).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-01-credibility-argument-text-area.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. Suggested: STEP-1 to STEP-3 as "Let the credibility argument wrap and hold several lines on the Source page.", then the version bump, then docs.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/handlers/credibility_test.go` `TestSourceCredibility`, case "multi-line argument keeps inner line breaks" *(new)* | Upsert with `Argument: "  Line one\nLine two\n\n"` → response `Assessment.Argument == "Line one\nLine two"`; `GetSourceWorkspace` → `Credibility.Argument == "Line one\nLine two"` | AC-2.3, AC-2.2 |
| T-2 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `saveCredibilityKeepsLineBreaks` *(new)* | `warmFromSession`; `selectDraft(key: "high_trust")`; `argumentDraft = "Census taker was a neighbour.\nAges rounded to 5.\n"`; `await save()` → `store.credibilityBySource[sourceID]?.argument == "Census taker was a neighbour.\nAges rounded to 5."`, `hasSavedAssessment`, `!isDirty`, `argumentDraft` equals the stored value; then `argumentDraft = "x"`, `cancel()` → `argumentDraft` equals the two-line value | AC-2.2, AC-2.4 |

AC-1.1 – AC-1.3, AC-2.1 and AC-2.5 are checked by hand (§ 10, D-6).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (credibility tests share fixtures).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (this PR adds none)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI uses only the existing kit `PVTextArea`; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated — n/a, not gated
- [ ] Domain doc updated if a rule in docs/*.md changed — none changes
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes the slice 1 credibility line and the scope entry's Check true: "A 500-character argument is fully visible and editable; Return adds a line; Save saves it."

1. Open any project, open a Source page at the default window width. The argument field is empty, shows "Why this grade — optional", and is two lines tall (AC-1.2).
2. Paste this 500-character argument (invented): `The enumerator lived two doors from the household and knew the family well, which makes the names and relationships credible. Ages, however, look rounded to the nearest five years for every adult in the street, and the birthplace column repeats the parish name for everyone, including a lodger known from other records to be foreign-born. Treat names and household composition as strong; treat ages as approximate and birthplaces as weak. The page is a clean original image, not a later copy or a transcript.` Every word is visible; the text wraps; no horizontal scrolling (AC-1.1).
3. Click after "foreign-born." and press Return. A new line appears; Save becomes enabled; nothing is saved yet (AC-2.1).
4. Press Tab. Focus moves to Save; no tab character appears in the text (AC-2.5).
5. Press Save. Navigate away and back to the Source. The argument shows with the line break (AC-2.2).
6. Paste the argument four more times to pass 12 lines. The field stops at 12 lines and scrolls inside (AC-1.3). Press Cancel: the saved argument returns, line break included (AC-2.4).
7. With VoiceOver on, focus the field: it is announced as "Why this grade — optional" (D-4).
8. Repeat step 2 at a Wide window width: the field wraps to the wider column.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A taller argument pushes the left column past the metadata column's height | `overviewColumns` already sets `minHeight` only; the column grows. Manual check step 6 looks at it |
| The 12-line bound is too short at the narrowest window for 500 characters | Manual check step 2 at the default width; Stop condition below |
| `PVTextArea`'s Tab handling does not reach Save in this layout | Manual check step 4; Stop condition below |

**Stop and report if:**

- `SourcePageCredibilityView` no longer uses `PVInput` for the argument, or no longer holds a local `argumentLocal` draft (§ 4 row is stale).
- At the app's minimum window width, the 500-character argument in § 10 needs more than 12 lines (D-1's bound is wrong): report the line count and width.
- Tab inside the area inserts a tab character or does not leave the field (D-3 is wrong).
- T-1 shows the engine drops or rewrites inner line breaks (Go change would be needed; Out of this plan).
- The Swift build prints a warning for `.accessibilityLabel` on `PVTextArea`, or VoiceOver announces the field twice (D-4 is wrong).

## 12. Open questions

None.

## 13. PR

**Title:** S15-01 — The credibility argument is a text area
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1, T-2, manual check steps); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-01-credibility-argument-text-area.md`.
