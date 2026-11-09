# S15-01 — The credibility argument is a text area

**Status:** Planned
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change beyond the control: the argument becomes the `PVTextArea` the Source page already draws for its description and notes; working rule "The credibility fix is idiomatic and needs no brief")
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s15-01-credibility-argument-text-area`
**Version bump:** patch — a UI bugfix on the Source page, no format or wire change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (a parameter on an existing kit control; no new primitive), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "The credibility fix is idiomatic and needs no brief" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [note](../2026-10-18-credibility-field-truncates-and-ignores-enter.md); [`research-judgment-model.md`](../../../research-judgment-model.md) § 2 (the assessment row); [`macos/App/DesignSystem/README.md`](../../../../macos/App/DesignSystem/README.md) Field row
**Planned:** 2026-11-09 against `main` @ `eab4453` (the implementer's base check compares against this commit; the plan's own commits on top of it are docs only)
**Review:** [`S15-01-credibility-argument-text-area.review.md`](S15-01-credibility-argument-text-area.review.md) — Fits with changes / Ready with changes, 1 round; no open blocker; one To user item answered with the recommendation (R-1)

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

**What stays.** The grade chips, the Save and Cancel buttons, the saved/unsaved status text, the section header, the save path (`SourceCredibilitySection.save()`), the engine write, and the stored column. The note's "Enter saves" wish is answered by the working rule: Return is a new line, Save is the button.

## 2. User stories

**US-1.** As a researcher, when I write a long argument for a Source's credibility grade, I see all of it while I write and after I save, so that I can read and correct it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source page at the default window size (the size a new window opens at); type the argument in § 10 step 2 (about 500 characters) | The text wraps; every character is visible without horizontal or vertical scrolling; the field is taller than one line |
| AC-1.2 | The argument field is empty and unfocused | The field shows the prompt "Why this grade — optional" and is two lines tall |
| AC-1.3 | An argument longer than 12 lines at the field's width | The field stops growing at 12 lines and its text scrolls vertically inside it |

**US-2.** As a researcher, I can break my argument into paragraphs and save it with the button I already use.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Focus in the argument; press Return | A new line is inserted; nothing is saved; Save becomes enabled |
| AC-2.2 | Argument `"Census taker was a neighbour.\nAges rounded to 5."`; press Save | The assessment saves; reopening the Source shows the same two lines |
| AC-2.3 | Argument `"  Line one\nLine two\n\n"` saved | Stored and shown as `"Line one\nLine two"`: ends trimmed, inner line break kept |
| AC-2.4 | Edit the argument; press Cancel | The field returns to the saved argument, line breaks included |
| AC-2.5 | Focus in the argument; press Tab | Focus leaves the field; no tab character is inserted |
| AC-2.6 | Saved argument `"Line one\nLine two"`; the field holds `"Line one\nLine two\n"` | Save is disabled (not dirty). The field holds `"Line one\n\nLine two"` → Save is enabled (dirty) |
| AC-2.7 | VoiceOver on; focus the argument field | It is announced once, as "Why this grade — optional" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVTextArea` takes an own accessibility label, resolved through `PVFieldLabel.resolve(own:field:)` like `PVSelect` | STEP-1, AC-2.7, D-4 |
| Replace `PVInput` with `PVTextArea` for the argument in `SourcePageCredibilityView` | STEP-2, AC-1.1 – AC-1.3, AC-2.1, AC-2.5 |
| One dirtiness rule: the view asks the model, the view's private copy goes | STEP-2, STEP-3, AC-2.6, D-5 |
| Pin, in tests, that a multi-line argument keeps inner line breaks in the engine and the model | STEP-3, STEP-4, AC-2.2 – AC-2.4, AC-2.6 |
| Patch version bump | STEP-5 |
| Docs: DesignSystem README Field row, map status, `completed.md`, this plan's status | STEP-1, STEP-6 |

| Out | Home |
| --- | --- |
| A shared default-action shortcut (Return or ⌘Return saves) for every `PVInlineEditActions` editor | not planned — the working rule chose platform text-area keys; reopen as an idea file only if dogfood asks again |
| Return saves the argument (the note's second wish) | not planned — refused by the working rule (user, 2026-11-09) |
| `PVTextArea` speaking its `prompt` when it has no own or field label (would name the notes composer and other prompt-only areas) | not planned in this PR — review R-1; changes what VoiceOver says on other pages, so it waits for a PR that checks each of them |
| A Source credibility argument length limit | not planned — the column is unbounded `TEXT` today and stays so |
| Any change to the description, notes or metadata editors | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Argument field | `macos/App/Features/Sources/SourcePageCredibilityView.swift` `SourcePageCredibilityView.body` | `PVInput(text: $argumentLocal, size: .sm, prompt: L10n.Sources.credibilityArgumentPlaceholder)` with `.accessibilityIdentifier("sources.page.credibility.argument")`. One line; long text scrolls sideways; Return does nothing (no `.onSubmit`) |
| View dirtiness | same file, `private var isDirty` | Repeats the model's rule (grade key differs, or trimmed argument differs from trimmed saved) on `argumentLocal`. Drives `saveDisabled`, `showsCancel`, and the status text. The only dirtiness the app reads |
| Local draft | same file, `@State argumentLocal`, `.onAppear`, `.onChange(of: model.credibility.argumentDraft)` | Typing stays local; Save copies it into `model.credibility.argumentDraft`. Unchanged by this PR |
| Model | `macos/App/Features/Sources/SourceCredibilitySection.swift` `isDirty`, `save()` | `isDirty` is the same rule on `argumentDraft`; read only by `SourcePageModelTests`. `save()` trims (`whitespacesAndNewlines`) and calls `store.upsertSourceCredibilityAssessment`; inner newlines pass through |
| Placement | `macos/App/Features/Sources/SourcePageView.swift` `overviewColumns` | Left column: description, `Spacer`, credibility pinned to the bottom; column `minHeight` tracks the metadata column. A taller argument grows the column; no layout change needed |
| Kit control | `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift` `PVTextArea` | `TextEditor`-backed; `lineLimit` sets min/max height in lines, then scrolls; Return inserts a newline; Tab / Shift-Tab move focus, Option-Tab inserts a tab; `prompt` is an `accessibilityHidden` overlay, so it is not spoken; ends with `.pvFieldNamed()`, which speaks only an enclosing `PVField` label. No own-label parameter |
| Naming rule | `macos/App/DesignSystem/Components/Field/PVField.swift` `PVFieldLabel.resolve(own:field:)`, `PVFieldNamed`, `View.pvFieldNamed()` | `resolve` says "own label wins, else the field's". `PVFieldNamed` always passes `own: nil`. Callers of `pvFieldNamed()`: `PVInput.swift`, `PVTextArea.swift` |
| Precedent for an own label | `macos/App/DesignSystem/Components/Select/PVSelect.swift` init parameter `accessibilityLabel: LocalizedStringResource? = nil`, `accessibilityLabelText` | Feeds its own label to `PVFieldLabel.resolve(own:field:)` |
| Sibling uses | `SourcePageDescriptionView.swift` (`lineLimit: 3...12`, `.bodyLarge`, inside `PVInlineEdit`); `SourcePageNotesView.swift` composer (`lineLimit: 2...8`, default `.body`, a `PVButton` below) | The notes composer is the same shape as the credibility block: an always-visible area with a button row beneath |
| Font today | `macos/App/DesignSystem/Components/Input/PVInput.swift` (`size.font` for `.sm`) | Small body type, the size of `PVTextArea.Typography.body` (`PVTypeScale.bodySmall`) |
| Engine | `core/database/sourcecredibility/assessments.go` `Upsert` | `strings.TrimSpace(in.Argument)`; stores `argument` as nullable `TEXT`. Inner newlines are stored as is |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources.credibilityArgumentPlaceholder` = "Why this grade — optional" (key `sources.page.credibilityArgumentPlaceholder` in `macos/App/Resources/Localizable.xcstrings`) | Reused; no new key |
| DesignSystem README | `macos/App/DesignSystem/README.md` Field row | "gives its label to the `PVInput`, `PVTextArea` or `PVSelect` inside it …; a control's own label wins (`PVFieldLabel.resolve`)" — true today only for `PVSelect` |
| Tests | `core/database/sourcecredibility/assessments_test.go` `TestAssessments` (table of `run` funcs; helpers `mustUser`, `mustSource`, `mustGrades`); `macos/ProvenenciaTests/SourcePageModelTests.swift` `selectCredibilityDraftDoesNotPersistUntilSave`, `cancelCredibilityRestoresDraftFromSaved`; `macos/ProvenenciaTests/PVFieldLabelTests.swift` `ownLabelWins` | Single-line arguments only; nothing pins line breaks |
| FakeStore | `macos/App/Platform/FakeStore.swift` `upsertSourceCredibilityAssessment` | Stores the `argument` it is given (no trim), so a Swift test sees the model's trim |
| Version | `VERSION`, `core/version.go` `Version`, `macos/Provenencia.xcodeproj/project.pbxproj` | `0.4.1`; the pbxproj has four `MARKETING_VERSION = 0.4.1;` lines (app and test targets, Debug and Release); the last bump (`9509e881`) changed all four |

**Precedent.** The note composer in `SourcePageNotesView.swift` for the call shape: a `PVTextArea(text:lineLimit:prompt:)` with `.accessibilityIdentifier` on the area and the action row below. `PVSelect`'s `accessibilityLabel` parameter for the kit change. Do not copy `PVInlineEdit` from the description (the credibility block has no resting/editing swap).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PVTextArea(text: $argumentLocal, lineLimit: 2...12, prompt: L10n.Sources.credibilityArgumentPlaceholder, accessibilityLabel: L10n.Sources.credibilityArgumentPlaceholder)` with the default `.body` typography | Working rule: the description's control. `.body` is the size of today's `PVInput(size: .sm)` text, so only the height changes. Lower bound 2 keeps the empty field close to today's height; upper bound 12 is the description's. At the default window size the argument in § 10 fits unscrolled; at the 700-pt minimum window width (`PVSpacing.widthWindowMin`) it may scroll, which AC-1.3 allows | `.bodyLarge` — changes the type size, a visible change the rule did not ask for; `2...8` (the notes composer) — too close to the length of a 500-character argument at half width |
| D-2 | Return inserts a newline (`PVTextArea`'s default); no `.onSubmit`, no key routing, no shortcut | Working rule "platform's text-area behavior (Return is a new line) and the existing Save button" (user, 2026-11-09) | Return-to-save with Shift-Return for a newline — refused by the rule |
| D-3 | No `focus:` handle is passed | `PVTextArea` without a handle sends Tab to the key-view loop, so Tab leaves the field (AC-2.5). Whether it lands on Save depends on the system's Keyboard navigation setting, which this PR does not depend on | Wrapping in `PVTabPair` — that is for an input and an area in one form |
| D-4 | `PVTextArea` gains `accessibilityLabel: LocalizedStringResource? = nil` *(new parameter)*, the last init parameter. `View.pvFieldNamed()` becomes `pvFieldNamed(own: LocalizedStringResource? = nil)` and `PVFieldNamed` gets `var own: LocalizedStringResource?`, passed as `PVFieldLabel.resolve(own: own.map(L10n.string), field: fieldLabel.map(L10n.string))`. `PVTextArea` calls `.pvFieldNamed(own: accessibilityLabel)`; `PVInput`'s call is unchanged (`own` defaults to nil). No prompt fallback (Out, R-1) | Today the `TextField` prompt names the field to VoiceOver; `PVTextArea` hides its prompt, so the swap would leave the field unnamed. The kit already has the rule "a control's own label wins" (`PVFieldLabel.resolve`, `PVSelect`); a call-site `.accessibilityLabel` would be a second naming mechanism beside it (review M-1) | A call-site `.accessibilityLabel` — second mechanism; wrapping in `PVField` — adds a visible label (chrome change); a new key "Credibility argument" — new copy, and the prompt already says what the field is for |
| D-5 | `SourceCredibilitySection` gets `func isDirty(argument: String) -> Bool` *(new)* holding the existing rule; `var isDirty: Bool` becomes `isDirty(argument: argumentDraft)`. The view deletes its `private var isDirty` and reads `model.credibility.isDirty(argument: argumentLocal)` at its three uses | One dirtiness rule, in the model, tested; the local draft stays because typing must not rewrite the page's observation graph (review M-2) | Keeping the view copy — the rule the button uses would stay untested; moving the draft into the model — the reason for `argumentLocal` |
| D-6 | The accessibility identifier stays `sources.page.credibility.argument` | Nothing else changes about the field's identity | — |
| D-7 | Tests: one Go table case beside the trim rule (`TestAssessments`), one Swift model test; the view swap and VoiceOver are checked by hand | The trim lives in `sourcecredibility.Upsert`; the handler passes the string through (review M-3). The kit naming rule is already tested by `PVFieldLabelTests.ownLabelWins`. No UI test target exists in `macos/` | A handler test in `api/ffi/handlers/credibility_test.go` — one layer above the rule |
| D-8 | Patch bump to `0.4.2`, all four `MARKETING_VERSION` lines | A Mac UI bugfix ([`docs/versioning.md`](../../../versioning.md) "What bumps"); four lines as the last bump `9509e881` did | none |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVTextArea(… accessibilityLabel:)` *(new parameter)* | component parameter | `PVSelect` `accessibilityLabel` | Extends `PVTextArea` the way `PVSelect` already works |
| `pvFieldNamed(own:)` *(new parameter on the existing modifier)* | modifier parameter | `View.pvFieldNamed()`, `PVFieldLabel.resolve(own:field:)` | Extends the one naming modifier; no second modifier |
| `SourceCredibilitySection.isDirty(argument:)` *(new)* | func | `SourceCredibilitySection.isDirty`, the view's `private var isDirty` | Same rule, one home; the property delegates to it, the view copy is deleted |
| `TestAssessments` case "multi-line argument keeps inner line breaks" *(new)* | test case | `TestAssessments` "upsert create then update" | Extends that table with its helpers |
| `SourcePageModelTests.saveCredibilityKeepsLineBreaks` *(new)* | test | `cancelCredibilityRestoresDraftFromSaved` | Same file and helpers |

Searched for: another multiline field component (`TextEditor`, `axis: .vertical` in `Features/`), an own-label slot in the kit (`accessibilityLabel:` in `DesignSystem/Components`, found `PVSelect`), other readers of `credibility.isDirty` (tests only), the trim rule's own test (`TestAssessments`).

**Builds on.** `PVTextArea`; `PVFieldLabel.resolve` and `pvFieldNamed`; `PVSelect`'s own-label parameter; the notes composer call shape.
**Deletes.** The `PVInput(… size: .sm …)` call in `SourcePageCredibilityView`; the view's `private var isDirty`. No callee loses its last caller. No string is removed.
**Next adds.** Any other `PVTextArea` outside a `PVField` (the notes composer, the note editor) can name itself with `accessibilityLabel:`; a later prompt fallback (R-1) is one line in `PVTextArea`.

### Reach

None — no new path into existing code; the argument can already hold line breaks through the engine (it trims only the ends), so a stored multi-line argument is not a new state, only one the UI can now write and show. `pvFieldNamed(own:)` defaults to nil, so `PVInput` and every existing `PVTextArea` speak exactly what they speak today.

### Scale

None — no write, list or loop changes; one control on one page.

### Edges

- **Assumptions.** The dirtiness rule compares trimmed strings; a trailing newline is whitespace, so `"a\nb\n"` vs saved `"a\nb"` is not dirty, while an added inner blank line is (AC-2.6, T-2).
- **Neighbors.** `PVFieldLabel.resolve`: own nil + field label → field label (unchanged for every `PVField` user); own set → own wins (`ownLabelWins`).
- **Removed inputs.** None — nothing is deleted or excluded.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — one string the user types; unbounded today and unchanged.

### Schema

None — no catalog change.

### Go core

None — no code change; one test case (T-1).

### FFI

None — the request already carries `argument` as a string.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Own label slot | `PVTextArea` (`accessibilityLabel:`), `pvFieldNamed(own:)` | `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/App/DesignSystem/Components/Field/PVField.swift` | D-4 |
| Credibility argument | `PVTextArea` (D-1) | `macos/App/Features/Sources/SourcePageCredibilityView.swift` | Replaces `PVInput`; same binding `$argumentLocal`; `.accessibilityIdentifier("sources.page.credibility.argument")` kept |
| Save / Cancel, status text | `PVInlineEditActions`, `Text` | same file | Same views; their dirtiness now comes from `model.credibility.isDirty(argument: argumentLocal)` (D-5) |

States: empty → prompt, two lines tall; long → grows to 12 lines then scrolls; saving → unchanged (`PVInlineEditActions.isSaving`). The area is not disabled while saving (today's `PVInput` was not either). Accessibility: identifier kept; label per D-4; Tab leaves the area (D-3).

### Strings

None — no new or removed key. `L10n.Sources.credibilityArgumentPlaceholder` is reused as the prompt and the accessibility label.

## 7. Steps

**STEP-1 — Give `PVTextArea` an own accessibility label**
Files: `macos/App/DesignSystem/Components/Field/PVField.swift`, `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`.
Do: per D-4. In the README Field row, change "a control's own label wins (`PVFieldLabel.resolve`)" so it names the controls that take one: "a `PVTextArea` or `PVSelect` given its own `accessibilityLabel` keeps it (`PVFieldLabel.resolve`)". Leave `PVInput` without the parameter.
Check: `rg -n "pvFieldNamed" macos/App` shows the definition taking `own`, `PVInput` calling `.pvFieldNamed()`, `PVTextArea` calling `.pvFieldNamed(own: accessibilityLabel)`; the project builds (STEP-4's `xcodebuild` builds it).

**STEP-2 — Swap the argument field and use the model's dirtiness**
Files: `macos/App/Features/Sources/SourcePageCredibilityView.swift`, `macos/App/Features/Sources/SourceCredibilitySection.swift`.
Do: replace the `PVInput` call with D-1's `PVTextArea`, keep `.accessibilityIdentifier("sources.page.credibility.argument")`. Add `func isDirty(argument: String) -> Bool` to `SourceCredibilitySection` with the body of today's `isDirty` using `argument` in place of `argumentDraft`; make `var isDirty` return `isDirty(argument: argumentDraft)`. In the view, delete `private var isDirty` and replace its three reads (`saveDisabled: !isDirty`, `showsCancel: isDirty`, `if !isDirty`) with `model.credibility.isDirty(argument: argumentLocal)`. Change nothing else in either file.
Check: `rg -n "PVInput|private var isDirty" macos/App/Features/Sources/SourcePageCredibilityView.swift` prints nothing.

**STEP-3 — Pin multi-line arguments in the engine**
Files: `core/database/sourcecredibility/assessments_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: add T-1 as a case in `TestAssessments`'s `tests` table, using `mustUser`, `mustSource`, `mustGrades` and the `writes.Call` shape of "upsert create then update" (action `create_source_credibility_assessment`, grade `grades[2]`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecredibility/ -run TestAssessments` passes.

**STEP-4 — Pin multi-line arguments and dirtiness in the model**
Files: `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: add T-2 beside `cancelCredibilityRestoresDraftFromSaved`, using the file's `makeStore` / `makeModel` helpers.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/SourcePageModelTests` passes with no new Swift warning.

**STEP-5 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj`. Rule: `.cursor/rules/versioning.mdc`.
Do: `0.4.1` → `0.4.2` in `VERSION`, `core.Version`, and all four `MARKETING_VERSION` lines (D-8). If `main` has moved past `0.4.1`, take the next patch of what `main` has.
Check: `cat VERSION`, `rg -n "Version = " core/version.go`, `rg -n "MARKETING_VERSION" macos/Provenencia.xcodeproj/project.pbxproj` all show the same version, four pbxproj lines.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan.
Do: after the PR is opened (so #N exists): set this plan's Status to `Landed in [#N](https://github.com/mendahu/provenencia/pull/N) (YYYY-MM-DD)` with a **Landed** note listing differences from the plan (or "None."); set the map row's Status to `Landed #N`; add under the heading of `completed.md` a paragraph in this form: "**S15-01 — The credibility argument is a text area** ([#N](…)). The Source page's credibility argument is a multi-line text area that wraps and grows to 12 lines; Return adds a line and Save saves it with its line breaks. `PVTextArea` takes its own accessibility label." The note `2026-10-18-credibility-field-truncates-and-ignores-enter.md` stays; spike close handles it ([`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md)).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-01-credibility-argument-text-area.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. Suggested: STEP-1 "Let a text area name itself to VoiceOver with its own label."; STEP-2 to STEP-4 "Let the credibility argument wrap and hold several lines on the Source page."; then the version bump; then docs.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sourcecredibility/assessments_test.go` `TestAssessments`, case "multi-line argument keeps inner line breaks" *(new)* | `Upsert` with `Argument: "  Line one\nLine two\n\n"` → returned `Assessment.Argument == "Line one\nLine two"`; `GetBySource` → `Argument == "Line one\nLine two"` | AC-2.3, AC-2.2 |
| T-2 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `saveCredibilityKeepsLineBreaks` *(new)* | `warmFromSession`; `selectDraft(key: "high_trust")`; `argumentDraft = "Census taker was a neighbour.\nAges rounded to 5.\n"`; `await save()` → `store.credibilityBySource[sourceID]?.argument == "Census taker was a neighbour.\nAges rounded to 5."`, `hasSavedAssessment`, `!isDirty`; `isDirty(argument: "Census taker was a neighbour.\nAges rounded to 5.\n") == false`; `isDirty(argument: "Census taker was a neighbour.\n\nAges rounded to 5.") == true`; then `argumentDraft = "x"`, `cancel()` → `argumentDraft` equals the two-line value | AC-2.2, AC-2.4, AC-2.6, D-5 |

AC-1.1 – AC-1.3, AC-2.1, AC-2.5 and AC-2.7 are checked by hand (§ 10, D-7). `PVFieldLabelTests.ownLabelWins` already covers the resolve rule D-4 uses.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (credibility tests share fixtures), `PVFieldLabelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (this PR adds none)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] No new kit primitive; `PVTextArea` gains one parameter; DesignSystem README Field row updated
- [ ] Version bump as the header says (VERSION, core/version.go, four MARKETING_VERSION lines together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status `Landed #N`; completed.md: the paragraph in STEP-6
- [ ] Brief moved to design/archive/ and design/README.md updated — n/a, not gated
- [ ] Domain doc updated if a rule in docs/*.md changed — none changes
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes the scope entry's Check true: "A 500-character argument is fully visible and editable; Return adds a line; Save saves it."

1. Open any project, open a Source page in a window at its default size. The argument field is empty, shows "Why this grade — optional", and is two lines tall (AC-1.2).
2. Paste this argument (invented, about 500 characters): `The enumerator lived two doors from the household and knew the family well, which makes the names and relationships credible. Ages, however, look rounded to the nearest five years for every adult in the street, and the birthplace column repeats the parish name for everyone, including a lodger known from other records to be foreign-born. Treat names and household composition as strong; treat ages as approximate and birthplaces as weak. The page is a clean original image, not a later copy or a transcript.` Every word is visible; the text wraps; no scrolling (AC-1.1).
3. Click after "foreign-born." and press Return. A new line appears; Save becomes enabled; nothing is saved yet (AC-2.1).
4. Press Tab. Focus leaves the field; no tab character appears in the text (AC-2.5).
5. Press Save. Navigate away and back to the Source. The argument shows with the line break (AC-2.2).
6. Add a Return at the very end: Save stays disabled. Add a Return between two lines instead: Save is enabled (AC-2.6). Press Cancel.
7. Paste the argument four more times to pass 12 lines. The field stops at 12 lines and scrolls inside (AC-1.3). Press Cancel: the saved argument returns, line break included (AC-2.4).
8. With VoiceOver on, focus the field: it is announced once, as "Why this grade — optional" (AC-2.7).
9. Widen the window to Wide and repeat step 2: the field wraps to the wider column.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A taller argument pushes the left column past the metadata column's height | `overviewColumns` sets `minHeight` only; the column grows. Manual check step 7 looks at it |
| The 12-line bound is too short at the default window size | Manual check step 2; Stop condition below |
| Changing `pvFieldNamed` alters what `PVInput` speaks | `own` defaults to nil, so `resolve(own: nil, field:)` is today's call; `PVFieldLabelTests` stays green |

**Stop and report if:**

- `SourcePageCredibilityView` no longer uses `PVInput` for the argument, no longer holds a local `argumentLocal` draft, or no longer has a `private var isDirty` (§ 4 rows are stale).
- At the default window size, the argument in § 10 step 2 needs more than 12 lines (D-1's bound is wrong): report the line count and the window width.
- Tab inside the area inserts a tab character or does not leave the field (D-3 is wrong).
- T-1 shows the engine drops or rewrites inner line breaks (a Go change would be needed; Out of this plan).
- VoiceOver announces the field twice or with no name (D-4 is wrong).
- Another caller of `pvFieldNamed` appears besides `PVInput` and `PVTextArea` (§ 4 row is stale).

## 12. Open questions

None.

## 13. PR

**Title:** S15-01 — The credibility argument is a text area
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence (T-1, T-2, manual check steps); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-01-credibility-argument-text-area.md`.
