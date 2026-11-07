# S14-39 — Properties page reads single-only

**Status:** Blocked — on board S14-D9
**Spike / slice:** Spike 14 · Slice 7e — Boolean value type
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D9](../design/S14-D9-properties-single-only.md), board not drawn yet
**Board:** the *Properties View* project, a [migrated design on claude.ai](https://claude.ai/artifact/XRS4WQDZbaDa7jAzVKYQdc), page *S7-D2 Subject Fields.dc.html*, section "S14-D9 · Single-only value type" (to draw). The implementer may not be able to open it: **Design → macOS** describes each frame in words, refreshed from the agreed board.
**Depends on:** S14-36 (landed [#733](https://github.com/mendahu/provenencia/pull/733)), S14-D9
**Blocks:** S14-37
**Branch:** `s14-39-properties-single-only`
**Version bump:** none — no researcher can make a single-only Property until S14-37 marks boolean creatable, so nothing a researcher sees changes ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** working rule "(Batch 7) Boolean is a first-class value type in two PRs" (Go owns the value-type rules; Swift keeps no copy); working rule "(Batch 5) FakeStore is a seeded fake, not a second engine"; bar outcome 31; S14-36 plan D-11 (`singleOnly` mapped, read here) and its **Landed** note
**Planned:** 2026-11-07 against `main` @ `aaf3b15`
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

A yes/no Property holds one value: the engine refuses one that holds several (S14-36). The Properties page still offers "Several values" for it. This PR makes the page read Go's single-only rule from the Properties snapshot. For a single-only type, the New property dialog shows Holds as a read-only "One value" with its reason, and the inspector shows the locked Holds row with its own spoken reason. The `boolean` key gets a pill colour of its own. Split out of S14-37 so each PR changes one view (user, 2026-11-07).

**Why now.** S14-37 marks boolean creatable; before it does, the page must stop offering what Go refuses.

**What stays.** Every other type's Holds radios and their notice; the seeded Property's locked row ("Set by Provenencia"); the other pills; the table, type strip, bindings and Delete. The create picker offers what Go marks creatable, as today (boolean only from S14-37).

## 2. User stories

**US-1.** As a researcher creating a Property, I am not offered several values for a type that holds one.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Rules with `boolean` creatable and single-only; create open; Value type `boolean` chosen | `model.holdsIsLocked(draft.valueType)` true; the Holds field shows "One value" and the hint "Yes / no Properties hold one value" |
| AC-1.2 | The draft has cardinality `multiple` (Several picked under `text`), then `boolean` is chosen | `draft.cardinality == "single"` |
| AC-1.3 | AC-1.2, then `text` chosen | `holdsIsLocked` false; cardinality stays `single`; the radios show |
| AC-1.4 | Create submitted with `boolean` | The Property is stored with cardinality `single` |

**US-2.** As a researcher reading my own yes/no Property, Holds is locked and says why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A user-origin Property of type `boolean` selected, rules loaded | `canEditCardinality` false; `holdsLockReason` = "Yes / no Properties hold one value"; `setCardinality("multiple")` returns false and writes nothing |
| AC-2.2 | A user-origin `text` Property selected | `canEditCardinality` true, as today |
| AC-2.3 | A seeded (`provenencia`) Property selected | `canEditCardinality` false; `holdsLockReason` = "Set by Provenencia", as today |
| AC-2.4 | A user-origin `boolean` Property, rules not loaded yet (`valueTypeRules` empty) | `canEditCardinality` true (Go refuses a change; the existing error shows) |

**US-3.** As a researcher, `boolean` reads like the other types, in its own colour.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `PropertiesTypeChrome.valueTypeForeground("boolean")` | The board's pigment, not `textSecondary` and not another type's |
| AC-3.2 | The `boolean` pill | Mono key `boolean`; VoiceOver label "Yes / no" (as today) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PropertiesSnapshot.isSingleOnly(_:)` from Go's rules | STEP-1, AC-1.1, AC-2.1 |
| `PropertiesModel`: `holdsIsLocked`, `canEditCardinality` for single-only, `holdsLockReason`, `chooseValueType` | STEP-1, AC-1.1 – AC-1.4, AC-2.1 – AC-2.4 |
| Copy `L10n.Properties.holdsSingleOnly(typeLabel:)` | STEP-2 |
| Create Holds field for a single-only type (SO-1, SO-2) | STEP-3, AC-1.1, AC-1.3 |
| Inspector locked Holds row with its reason (SO-3, SO-4) | STEP-3, AC-2.1 – AC-2.3 |
| `boolean` pill colour (SO-5) | STEP-4, AC-3.1 |
| The six-type picker (SO-6), Wide (SO-W1) | manual check |
| Docs | STEP-6 |

| Out | Home |
| --- | --- |
| Marking boolean creatable | S14-37 |
| The composer's Yes / no editor | S14-37 (brief S14-D8) |
| Value type labels in place of keys | not planned (user, 2026-11-07) |
| A Go change | not planned (S14-36 sends `single_only`; `properties.Update` already refuses multiple) |

## 4. Current state

Verified at `aaf3b15`.

| Area | Where | Today |
| --- | --- | --- |
| Rules on the snapshot | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesSnapshot.valueTypeRules`, `creatableValueTypes`; `macos/App/Platform/GenealogyStore.swift` `CatalogPropertyValueTypeRule.singleOnly` | `singleOnly` mapped, read nowhere |
| Create model | `macos/App/Features/Properties/PropertiesModel.swift` `Draft` (`valueType`, `cardinality` default single), `offeredValueTypes`, `acceptsValueType`, `canSubmitCreate`, `submitCreate` | No single-only rule |
| Inspector model | `PropertiesModel.canEditCardinality` (~202: `selectedProperty?.origin == user`), `setCardinality` (~386, guarded by it), `cardinalityNotice` | A user boolean Property is editable |
| Create form | `macos/App/Features/Properties/PropertiesView.swift` `createForm` (~770): Value type `PVSegmentedControl` (~833, `select: { draft?.valueType = $0 }`), Holds `PVField` with two `PVRadio` (~847–870) | Always two radios |
| Inspector | `PropertiesView.inspectorBody` (~433), `holdsRow` (~518): radios when `canEditCardinality`, else `inspectorMetaRow` with text and `lock.fill` labelled `L10n.Properties.holdsLocked` | Locked branch says "Set by Provenencia" only |
| Pill | `PropertiesSnapshot.swift` `PropertiesValueTypePill`; `macos/App/Features/Catalog/SubjectTypeChrome.swift` `PropertiesTypeChrome.valueTypeForeground` | `boolean` falls to `PVColor.textSecondary` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Properties.holds*` (~7450–7500), `holdsGroup(label:)` is the format precedent; `PropertyValueType.label` (`InterpretationValueKinds.swift`), `SubjectPropertyValueType.label(_:)` | No single-only copy |
| Tests | `macos/ProvenenciaTests/PropertiesModelTests.swift` `goValueTypeRules` (~257), `createDefaultsToOneValueAndCanSubmitSeveral` (~388), `userPropertyCardinalityChangeKeepsTheNoticeUntilSelectionMoves` (~400), helpers `makeModel`, `warm`, `userProperty`, `personType` | Templates |

**Precedent.** S14-36 ([#733](https://github.com/mendahu/provenencia/pull/733)) added `creatableValueTypes` and `acceptsValueType` reading the same rules: copy that shape (a snapshot read, the model's rule, empty rules before the first load leave Go to decide). The inspector's seeded branch of `holdsRow` is the locked row to reuse.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PropertiesSnapshot.isSingleOnly(_ valueType: String) -> Bool` *(new)* = `valueTypeRules.first { $0.valueType == valueType }?.singleOnly ?? false` | Go owns the rule; with no rules loaded Go decides, as `acceptsValueType` does | A Swift list of single-only types — the working rule forbids a copy |
| D-2 | `PropertiesModel.holdsIsLocked(_ valueType: String) -> Bool` *(new)* = `snapshot.isSingleOnly(valueType)`. `canEditCardinality` becomes `selectedProperty?.origin == user && !holdsIsLocked(selectedProperty.valueType)` | One model rule the view and `setCardinality` both follow | Checking in the view only — `setCardinality` would still write |
| D-3 | `PropertiesModel.holdsLockReason: String` *(new)*: for the selected Property, `L10n.Properties.holdsSingleOnly(typeLabel:)` when `holdsIsLocked`, else `L10n.string(L10n.Properties.holdsLocked)` | SO-3: the lock says why | A second lock row |
| D-4 | `PropertiesModel.chooseValueType(_ valueType: String, in draft: inout Draft)` *(new)* sets `draft.valueType` and, when `holdsIsLocked(valueType)`, `draft.cardinality = single`. The picker's `select` calls it on the view's draft | SO-2: a hidden "Several" never reaches Go; one place | Resetting in `submitCreate` — the form would show the wrong state until submit |
| D-5 | Create form: when `model.holdsIsLocked(draft.valueType)`, the Holds `PVField` keeps its label and takes `hint: L10n.Properties.holdsSingleOnly(typeLabel:)` (a `String`, which `PVField`'s `hint: (any PVCopy)?` takes); its content is the inspector locked row's content (the "One value" `Text` and the `lock.fill` image) with no radios. Otherwise the radios and `holdsEventsHint`, as today | SO-1; the user chose read-only "One value" (2026-11-07) | A disabled radio — the user declined it |
| D-6 | Inspector: `holdsRow`'s locked branch labels the lock `Text(verbatim: model.holdsLockReason)` in place of `Text(L10n.Properties.holdsLocked)` | SO-3; the seeded case reads as today | — |
| D-7 | Copy (user, 2026-11-07): `L10n.Properties.holdsSingleOnly(typeLabel: String) -> String` *(new)*, key `properties.inspector.holdsSingleOnly`, "%@ Properties hold one value", comment "Why Holds is locked for a single-only value type; argument is the type's label (Yes / no)". The argument is `L10n.string(SubjectPropertyValueType.label(valueType))` | The type's label, so the sentence reads "Yes / no Properties hold one value" and fits a later single-only type | A boolean-only sentence |
| D-8 | Pill colour: `valueTypeForeground` gains `case "boolean": return <the board's PVPalette token>`. Filled from the agreed board | SO-5 | — |
| D-9 | Keys stay: the pill and picker show `boolean` (user, 2026-11-07) | Every type keeps its mono key | Labels |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

Searched: `singleOnly`, `creatableValueTypes`, `acceptsValueType`, `canEditCardinality`, `holdsLocked`, `holdsRow`, `valueTypeForeground`, `holdsGroup`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PropertiesSnapshot.isSingleOnly` | func | `creatableValueTypes` | A per-type read beside it |
| `PropertiesModel.holdsIsLocked`, `holdsLockReason`, `chooseValueType` | func / var | `acceptsValueType`, `canEditCardinality` | Extends the model's rules |
| `L10n.Properties.holdsSingleOnly(typeLabel:)` | key | `holdsGroup(label:)` | Same namespace and shape |
| `boolean` arm | switch arm | `term` arm | Extends the table |

**Builds on.** S14-36's rules on the snapshot; the inspector's locked Holds row.
**Deletes.** None — nothing becomes dead (the seeded branch is reused).
**Next adds.** S14-37 flips boolean creatable; the picker then offers it and every rule here applies. A later single-only type needs only its `Types` row in Go.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A user-origin single-only Property in the inspector | `holdsRow` locked branch, `setCardinality` | Locked, reason spoken, no write (AC-2.1) |
| A single-only type chosen in create | Holds field, `submitCreate` | Read-only, cardinality single (AC-1.1, AC-1.4) |
| Rules not loaded | `holdsIsLocked` false | Today's radios; Go refuses a multiple boolean with `properties.invalid` through the existing error path (AC-2.4) |

### Scale

None — one lookup in an eight-row list per draw.

### Edges

- **Assumptions.** `PVField`'s `hint` is `(any PVCopy)?`, and `String` is `PVCopy` (`macos/App/DesignSystem/Components/Copy/PVCopy.swift`). `holdsRow` keeps its accessibility group for the radios only.
- **Neighbors.** A seeded single-only Property (none today) reads the single-only reason, since the type is the stronger reason. A user `text` Property keeps its radios (AC-2.2).
- **Removed inputs.** None.
- **Real data.** `propertymeta.Types`: boolean is the only single-only row (`TestTypes` pins it).
- **Untrusted size.** None.

### Schema

None — no catalog change.

### Go core

None — Go already sends and enforces the rule.

### FFI

None.

### macOS

From the S14-D9 inventory.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value type picker | `PVSegmentedControl` (Ship) | `PropertiesView.createForm` | `select` calls `chooseValueType` (D-4) |
| Holds field (create) | `PVField` + `PVRadio` (Extend) | same | D-5 |
| Holds row (inspector) | `holdsRow` locked branch (Extend) | `PropertiesView` | D-6 |
| Value type pill | `PropertiesValueTypePill`, `valueTypeForeground` (Extend) | `PropertiesSnapshot.swift`, `SubjectTypeChrome.swift` | D-8 |

Frames, in words: written from the agreed board (Blocked). Brief frames: 1 create with `boolean` and the read-only Holds; 2 back to `text`, radios; 3 inspector Can read, locked; 4 inspector Languages spoken, radios; 5 table pills; 6 dark of 1 and 3; 7 Wide of 2 and 4.

### Strings

New: `properties.inspector.holdsSingleOnly` (D-7). Removed: none.

## 7. Steps

**STEP-1 — The model reads single-only**
Files: `macos/App/Features/Properties/PropertiesSnapshot.swift`, `PropertiesModel.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-3, D-4.
Check: the app builds.

**STEP-2 — The reason's copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-7.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The create form and the inspector**
Files: `macos/App/Features/Properties/PropertiesView.swift`. Skill: `add-ui-component`.
Do: D-4 (the picker's `select`), D-5, D-6.
Check: the app builds with no Swift warning; manual checks 1–3.

**STEP-4 — The `boolean` pill**
Files: `macos/App/Features/Catalog/SubjectTypeChrome.swift`.
Do: D-8.
Check: manual check 4.

**STEP-5 — Tests**
Files: `macos/ProvenenciaTests/PropertiesModelTests.swift`. Skill: `add-swift-test`.
Do: add `singleOnlyRules` *(new fixture)*: `("text", true, false), ("boolean", true, true)`; add T-1 – T-3.
Check: Xcode tests pass.

**STEP-6 — Docs**
Files: `../deployment-plan.md` (S14-39 row status), `../completed.md`, this plan, `../design/S14-D9-properties-single-only.md` → `../design/archive/`, `../design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-39-properties-single-only.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PropertiesModelTests` `singleOnlyCreateHoldsOneValue` *(new)* | rules `singleOnlyRules`; draft `text` + `multiple`; `chooseValueType("boolean")` → `single`, `holdsIsLocked` true; `chooseValueType("text")` → still `single`, not locked; submit `boolean` → stored `single` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1, D-2, D-4 |
| T-2 | Swift | `PropertiesModelTests` `singleOnlyUserPropertyHoldsIsLocked` *(new)* | user `boolean` Property, rules loaded → `canEditCardinality` false, `holdsLockReason` = `L10n.Properties.holdsSingleOnly(typeLabel: "Yes / no")`, `setCardinality("multiple")` false, store unchanged; user `text` → true; seeded → false, reason `holdsLocked` | AC-2.1, AC-2.2, AC-2.3, D-3, D-7 |
| T-3 | Swift | `PropertiesModelTests` `singleOnlyWaitsForTheRules` *(new)* | user `boolean`, `valueTypeRules` empty → `canEditCardinality` true, `holdsIsLocked` false | AC-2.4 |
| — | manual | §10 | Holds field, inspector, pill, picker, Wide | AC-3.1, AC-3.2, SO-5, SO-6, SO-W1 |

Assert copy with `L10n.string(...)` / the `L10n` format function. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PropertiesModelTests` cardinality tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: no name renamed or deleted; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the S14-D9 board and inventory; no new kit primitive
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S14-39 row status Landed #N; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc: none changes (the single-only rule is already in interpretation-layer-data-model.md)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with a Person type. In the FakeStore preview or a debug build seeded so Go marks boolean creatable (a local, uncommitted flip of `propertymeta.Types`), or with a boolean Property written through a Go test catalog:

1. New property: choose `integer`, pick Several values, then choose `boolean`: Holds reads "One value" with the lock and "Yes / no Properties hold one value". Choose `text`: the radios are back, One value selected. (AC-1.1, AC-1.2, AC-1.3, SO-1, SO-2, SO-6)
2. Create "Can read" as `boolean`; the inspector's Holds is locked; VoiceOver on the lock reads "Yes / no Properties hold one value". (AC-1.4, AC-2.1, SO-3)
3. Select a user `text` Property: the radios work as before. Select a seeded Property: "Set by Provenencia". (AC-2.2, AC-2.3, SO-4)
4. The table and inspector show `boolean` in its pigment, in light and dark. (AC-3.1, AC-3.2, SO-5)
5. Widen past 1600pt of pane: the dialog and inspector look the same. (SO-W1)

Revert the local flip before committing.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Holds locked before the rules load | D-1 reads false with no rules; Go refuses; AC-2.4 |
| A seeded Property's lock text changes | D-3 keeps `holdsLocked` for non-single-only; T-2 |

**Stop and report if:**

- The six-segment picker truncates a segment at the dialog's width on the agreed board or in the app (SO-6).
- The agreed S14-D9 board differs from §6 macOS in anything but wording or the pigment.

## 12. Open questions

None from the code. Blocked until the S14-D9 board is agreed; then D-8's pigment and §6 frames are filled in and `review-plan` runs.

## 13. PR

**Title:** S14-39 — Properties page reads single-only
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-39-properties-single-only.md`.
