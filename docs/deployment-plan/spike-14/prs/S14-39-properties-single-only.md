# S14-39 — Properties page reads single-only

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 7e — Boolean value type
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D9](../design/S14-D9-properties-single-only.md), board agreed 2026-11-07
**Board:** S14-D9 — the *Properties View* project, a [migrated design on claude.ai](https://claude.ai/artifact/XRS4WQDZbaDa7jAzVKYQdc), page *S7-D2 Subject Fields.dc.html*, section "S14-D9 · Single-only value type", frames 1, 2, 3, 4, 5, 6a, 6b and the "Value type pills · SO-5" table. Changes from the brief: every pill gets a light and a dark shade, `text` moves to paper-700 and `term` to ochre-900 to pass 4.5:1 (D-8); Holds in create keeps the radios' height (D-5); the lock also shows its reason as a tooltip (D-4); dark mode is the board's theme switch, not separate frames. The implementer may not be able to open it: **Design → macOS** describes each frame in words.
**Depends on:** S14-36 (landed [#733](https://github.com/mendahu/provenencia/pull/733)), S14-D9
**Blocks:** S14-37
**Branch:** `s14-39-properties-single-only`
**Version bump:** patch — before S14-37 a researcher sees only the pills' new pigments (`depiction` coloured, `text` and `term` darker, dark-mode shades), a polish of an existing screen ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc` (Go wire vocabulary is a typed enum; views and models use the typed cases)
**Context:** working rule "(Batch 7) Boolean is a first-class value type in two PRs" (Go owns the value-type rules; Swift keeps no copy); working rule "(Batch 5) FakeStore is a seeded fake, not a second engine"; bar outcome 31; S14-36 plan D-11 (`singleOnly` mapped, read here) and its **Landed** note
**Planned:** 2026-11-07 against `main` @ `aaf3b15`
**Review:** [`S14-39-properties-single-only.review.md`](S14-39-properties-single-only.review.md) — Fits with changes — Ready with changes, round 1 answered; round 2 runs once the board is agreed

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

A yes/no Property holds one value: the engine refuses one that holds several (S14-36). The Properties page still offers "Several values" for it. This PR makes the page read Go's single-only rule from the Properties snapshot. For a single-only type, the New property dialog shows Holds as a read-only "One value" with its reason, and the inspector shows the locked Holds row with its own spoken reason. The value-type pill gets a pigment for `boolean` and `depiction`, chosen through the typed `PropertyValueType` enum so the next type cannot fall to grey. Split out of S14-37 so each PR changes one view (user, 2026-11-07).

**Why now.** S14-37 marks boolean creatable; before it does, the page must stop offering what Go refuses.

**What stays.** Every other type's Holds radios and their notice; the seeded Property's locked row ("Set by Provenencia"); the other pills; the table, type strip, bindings and Delete. The create picker offers what Go marks creatable, as today (boolean only from S14-37).

## 2. User stories

Tests use a neutral rules fixture in which `integer` is creatable and single-only (D-9), so they prove nothing is tied to `boolean`.

**US-1.** As a researcher creating a Property, I am not offered several values for a type that holds one.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Rules with `integer` single-only; `model.holdsLockReason(valueType: "integer", origin: user)` | "Integer Properties hold one value" |
| AC-1.2 | A draft with cardinality `multiple`, then `draft.chooseValueType("integer", singleOnly: true)` | `draft.valueType == "integer"`, `draft.cardinality == "single"` |
| AC-1.3 | AC-1.2, then `draft.chooseValueType("name", singleOnly: false)` | cardinality stays `single`; `holdsLockReason(valueType: "name", origin: user)` is nil (the radios show) |
| AC-1.4 | Create submitted after AC-1.2 | The stored Property has cardinality `single` (model-level: the draft carries it; Go's refusal of `multiple` is S14-36's `TestValueTypeRules`) |
| AC-1.5 | No rules loaded; `holdsLockReason(valueType: "integer", origin: user)` | nil (Go decides) |

**US-2.** As a researcher reading my own single-only Property, Holds is locked and says why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A user-origin `integer` Property selected, rules loaded | `canEditCardinality` false; `holdsLockReason` for it is "Integer Properties hold one value"; `setCardinality("multiple")` returns false and the store is unchanged |
| AC-2.2 | A user-origin `name` Property selected | `canEditCardinality` true, as today |
| AC-2.3 | A seeded (`provenencia`) Property of a type that is not single-only | `canEditCardinality` false; reason "Set by Provenencia", as today |
| AC-2.4 | A user-origin `integer` Property, no rules loaded | `canEditCardinality` true (FakeStore does not enforce; Go refuses with `properties.invalid` through the existing error path) |

**US-3.** As a researcher, every value type has its own pill colour.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `PropertiesTypeChrome.valueTypeForeground` for each `PropertyValueType` case | D-8's table: eight pigments, no two the same, each a light 700 (900 for `term`) and dark 300 shade |
| AC-3.2 | `valueTypeForeground("nope")` | `PVColor.textSecondary` |
| AC-3.3 | The `boolean` pill | Mono key `boolean`; VoiceOver "Yes / no" (as today) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PropertiesSnapshot.valueTypeRule(_:)`, the one lookup into Go's rules | STEP-2, AC-1.1, AC-1.5 |
| `PropertiesModel.holdsLockReason(valueType:origin:)`; `canEditCardinality` from it | STEP-2, AC-1.1, AC-1.3, AC-1.5, AC-2.1 – AC-2.4 |
| `Draft.chooseValueType(_:singleOnly:)`, `Draft.toggleBind(typeID:)`; dead `toggleDraftBind` deleted | STEP-2, AC-1.2, AC-1.3 |
| Copy `L10n.Properties.holdsSingleOnly(typeLabel:)` | STEP-1 |
| One locked-value view for the three lock sites; create Holds field (SO-1, SO-2); inspector row (SO-3, SO-4) | STEP-3 |
| Typed pill colours with `boolean` and `depiction` (SO-5); the `"term"` literal typed | STEP-4, AC-3.1, AC-3.2 |
| Six-type picker (SO-6), Wide (SO-W1) | manual check |
| Docs | STEP-6 |

| Out | Home |
| --- | --- |
| Marking boolean creatable | S14-37 |
| The composer's Yes / no editor | S14-37 (brief S14-D8) |
| Value type labels in place of keys | not planned (user, 2026-11-07) |
| A Go change | not planned (S14-36 sends `single_only`; `properties.Update` refuses multiple) |
| FakeStore refusing a single-only `multiple` | not planned (a seeded fake, working rule "FakeStore is a seeded fake") |
| `SubjectPropertyValueType.label` falling back to "Text" for an unknown key | not planned (Go's CHECK allows only known keys) |
| The stale value-type list in the `CatalogProperty.valueType` comment (`GenealogyStore.swift`) | not planned |

## 4. Current state

Verified at `aaf3b15`.

| Area | Where | Today |
| --- | --- | --- |
| Rules on the snapshot | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesSnapshot.valueTypeRules`, `creatableValueTypes`; `macos/App/Platform/GenealogyStore.swift` `CatalogPropertyValueTypeRule.singleOnly` (~948), mapped in `GoStore.swift` (~1944) | `singleOnly` read nowhere; no per-type lookup |
| Create model | `macos/App/Features/Properties/PropertiesModel.swift` `Draft` (`valueType`, `cardinality` default single, `bindTypeIDs`), `offeredValueTypes`, `acceptsValueType`, `canSubmitCreate`, `submitCreate`; `toggleDraftBind(typeID:)` (~330, no caller) | No single-only rule |
| Inspector model | `PropertiesModel.canEditCardinality` (~202: `selectedProperty?.origin == user`), `setCardinality` (~386, guarded by it), `cardinalityNotice` | A user single-only Property is editable |
| Create form | `macos/App/Features/Properties/PropertiesView.swift` `PropertiesCreateHost` (`@State private var draft: PropertiesModel.Draft?` ~726), `createForm` (~770): Value type `PVSegmentedControl` (~833, `select: { draft?.valueType = $0 }`), Holds `PVField` + two `PVRadio` (~847–870); `toggleBind(typeID:)` (~925) | Always two radios |
| Inspector | `PropertiesView.inspectorBody` (~433): Value type row with a `lock.fill` image (~447); `holdsRow` (~518): radios when `canEditCardinality`, else `inspectorMetaRow` with text and a `lock.fill` image (~562) labelled `L10n.Properties.holdsLocked`; `property.valueType == "term"` literal (~475) | Two hand copies of the lock; the lock says "Set by Provenencia" only |
| Pill | `PropertiesSnapshot.swift` `PropertiesValueTypePill`; `macos/App/Features/Catalog/SubjectTypeChrome.swift` `PropertiesTypeChrome.valueTypeForeground(_ valueType: String)` | String switch; `depiction` and `boolean` fall to `PVColor.textSecondary` |
| Value types | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType` (eight cases, `label`), pinned to Go by `api/ffi/swift_vocabulary_test.go`; `PropertiesModel.swift` `SubjectPropertyValueType.label(_:)` | — |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Properties.holds*` (~7450–7500); `holdsGroup(label:)` is the format precedent | No single-only copy |
| Field | `macos/App/DesignSystem/Components/Field/PVField.swift` `hint: (any PVCopy)?`; `String` is `PVCopy` (`macos/App/DesignSystem/Components/Copy/PVCopy.swift`) | Takes a formatted hint |
| Tests | `macos/ProvenenciaTests/PropertiesModelTests.swift` `goValueTypeRules` (~257: name creatable, integer creatable, date not), `createDefaultsToOneValueAndCanSubmitSeveral` (~388), `userPropertyCardinalityChangeKeepsTheNoticeUntilSelectionMoves` (~400); helpers `makeModel`, `warm`, `personType`, `nameProperty` (seeded `name`), `userProperty(id:usedBy:)` (hard-codes `valueType: "text"`) | Templates |

**Precedent.** S14-36 ([#733](https://github.com/mendahu/provenencia/pull/733)) added `creatableValueTypes` and `acceptsValueType` over the same rules, with empty rules leaving Go to decide: copy that shape. The inspector's seeded branch of `holdsRow` is the locked row.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `PropertiesSnapshot.valueTypeRule(_ valueType: String) -> CatalogPropertyValueTypeRule?` *(new)* = `valueTypeRules.first { $0.valueType == valueType }`. It is the one lookup into Go's rules; S14-37's composer reads it too (its D-6 is changed to match, review M-1) | One lookup, no stored copy; a later flag needs no new lookup | `isSingleOnly(_:)` — one predicate per flag; a dictionary on the composer — a second copy |
| D-2 | `PropertiesModel.holdsLockReason(valueType: String, origin: String) -> String?` *(new)*: `snapshot.valueTypeRule(valueType)?.singleOnly == true` → `L10n.Properties.holdsSingleOnly(typeLabel: L10n.string(SubjectPropertyValueType.label(valueType)))`; else `origin != CatalogOrigin.user` → `L10n.string(L10n.Properties.holdsLocked)`; else nil (editable). `canEditCardinality` becomes `selectedProperty.map { holdsLockReason(valueType: $0.valueType, origin: $0.origin) == nil } ?? false`. A single-only type wins over a seeded origin | One place decides "locked?" and "why" for create and the inspector (review M-3); `setCardinality` is already guarded by `canEditCardinality` | Three model APIs and a view-built hint |
| D-3 | `PropertiesModel.Draft` gains `mutating func chooseValueType(_ valueType: String, singleOnly: Bool)` *(new)*: sets `valueType`; when `singleOnly`, sets `cardinality = single`. And `mutating func toggleBind(typeID: String)` *(new)*, the body of today's toggle. The picker's `select` is `{ vt in draft?.chooseValueType(vt, singleOnly: model.snapshot.valueTypeRule(vt)?.singleOnly == true) }`; `PropertiesView.toggleBind(typeID:)` becomes `draft?.toggleBind(typeID: typeID)`. `PropertiesModel.toggleDraftBind(typeID:)` is deleted | Draft edits live on the draft; fits the Optional `@State` (`draft?.…`); removes a dead copy (review M-5, X-2) | An `inout` model method — an unwrap at the call site |
| D-4 | `fileprivate struct PropertiesLockedValue<Content: View>: View` *(new, PropertiesView.swift)* `{ let reason: String; @ViewBuilder let content: Content }`: an `HStack(spacing: 6)` of `content` and the existing `lock.fill` image (12pt semibold, `PVColor.textFaint`, `.accessibilityLabel(Text(verbatim: reason))`, `.help(reason)`). Callers: the Value type row (`content` = its pill, `reason` = `L10n.string(L10n.Properties.valueTypeImmutable)`), `holdsRow`'s locked branch and the create Holds field (`content` = the One / Several `Text` in `PVFont.body(size: PVTypeScale.bodySmall)`, `PVColor.textPrimary`, as today) | Three sites, one look (review M-4, X-3); the board shows the reason as the lock's tooltip too | `PVIcon(.lock)` — a visible change the board did not draw |
| D-5 | Create Holds field: when `model.holdsLockReason(valueType: draft.valueType, origin: CatalogOrigin.user)` is a reason, the `PVField(label: L10n.Properties.holds, hint: reason)` holds a `ZStack(alignment: .topLeading)` of the two radios, `.hidden()` and `.accessibilityHidden(true)` (they reserve the height, so Bind to subject types does not move: board frame 1), and `PropertiesLockedValue(reason: reason) { Text(L10n.Properties.holdsOne) }` with `.accessibilityIdentifier("properties.create.holds.locked")`; VoiceOver reads "Holds, One value, Yes / no Properties hold one value". Otherwise the radios and `holdsEventsHint`, as today | SO-1, SO-2; the user chose read-only "One value" drawn like the inspector row (2026-11-07); the board keeps the field's height | A disabled radio; a read-only `PVInput` (review M-7); a fixed `minHeight` — a number that drifts from the radios |
| D-6 | Inspector: `holdsRow` shows the radios when `model.canEditCardinality`, else `PropertiesLockedValue(reason: model.holdsLockReason(…) ?? "")` with the One / Several text, as today | SO-3, SO-4 | — |
| D-7 | Copy (user, 2026-11-07): `L10n.Properties.holdsSingleOnly(typeLabel: String) -> String` *(new)*, key `properties.inspector.holdsSingleOnly`, default "%@ Properties hold one value", comment "Why Holds is locked for a single-only value type; argument is the type's label, such as Yes / no", built like `holdsGroup(label:)` | Reads "Yes / no Properties hold one value"; fits a later single-only type | A boolean-only sentence |
| D-8 | `valueTypeForeground(_ valueType: String)` switches on `PropertyValueType(rawValue: valueType)`, exhaustively over the eight cases, each `Color.pvDynamic(light:dark:)` from the agreed board: `text` paper700 / paper300; `integer` copper700 / copper300; `date` lapis700 / lapis300; `name` iron700 / iron300; `subject` plum700 / plum300; `term` ochre900 / ochre300; `boolean` verdigris700 / verdigris300; `depiction` madder700 / madder300 (all `PVPalette`). `nil` → `PVColor.textSecondary`. `inspectorBody`'s `property.valueType == "term"` becomes `== PropertyValueType.term.rawValue` | `.cursor/rules/macos-client.mdc`: views use the typed cases; the next type fails to compile until it has a pigment (review M-2). The board measured each on the sunken pill: every pair passes 4.5:1 (ochre700 was 4.15:1, so `term` moves to ochre900; `text` moves from paper600 to paper700) | Adding `case "boolean":` to the string switch; one shade for both themes, as today |
| D-9 | Test fixture: `goValueTypeRules` gains single-only on `integer`: `("name", true, false), ("integer", true, true), ("date", false, false)`. `userProperty` gains `valueType: String = "text"` | The rule is general; one fixture for S14-37 to extend (review M-6, X-6) | A second, boolean fixture |
| D-10 | Keys stay: the pill and picker show `boolean` (user, 2026-11-07) | Every type keeps its mono key | Labels |

**Rules this plan bends.** None. D-8 follows `.cursor/rules/macos-client.mdc`.

## 6. Design

### Codebase fit

Searched: `singleOnly`, `creatableValueTypes`, `acceptsValueType`, `canEditCardinality`, `holdsLocked`, `holdsRow`, `lock.fill`, `valueTypeForeground`, `holdsGroup`, `toggleDraftBind`, `toggleBind`, `PropertyValueType(rawValue:`, `PVInput(… isReadOnly`, `PVSymbol.lock`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PropertiesSnapshot.valueTypeRule(_:)` | func | `creatableValueTypes` | A per-type lookup beside it; S14-37 reuses it |
| `PropertiesModel.holdsLockReason(valueType:origin:)` | func | `canEditCardinality` | Now derived from it |
| `Draft.chooseValueType`, `Draft.toggleBind` | mutating funcs | `PropertiesView.toggleBind`, `PropertiesModel.toggleDraftBind` | Replaces both copies |
| `PropertiesLockedValue` | fileprivate view | two inline `lock.fill` rows | Replaces them |
| `L10n.Properties.holdsSingleOnly(typeLabel:)` | key | `holdsGroup(label:)` | Same namespace and shape |
| Typed `valueTypeForeground` | switch | the string switch | Replaces it |

**Builds on.** S14-36's rules on the snapshot; the inspector's locked Holds row; `PropertyValueType`.
**Deletes.** `PropertiesModel.toggleDraftBind(typeID:)`; the body of `PropertiesView.toggleBind(typeID:)`; the two inline lock rows; the string switch in `valueTypeForeground`; the `"term"` literal.
**Next adds.** S14-37 reads `valueTypeRule(_:)?.negatable` in the composer and flips boolean creatable; the picker then offers it and every rule here applies. A later single-only type needs only its `Types` row in Go; a later value type needs a pigment to compile.

### Reach

| New path or state | Destination | Checked |
| --- | --- | --- |
| A user-origin single-only Property in the inspector | `holdsRow` locked branch, `setCardinality` | Locked, reason spoken, no write (AC-2.1) |
| A single-only type chosen in create | Holds field, `submitCreate` | Read-only, cardinality single (AC-1.2, AC-1.4) |
| Rules not loaded | `holdsLockReason` nil | Today's radios; Go refuses (AC-1.5, AC-2.4). A draft cannot hold a single-only type before the rules load: the picker offers only the draft's own type, `text`, until then |
| A seeded Property of a single-only type (none today) | `holdsLockReason` | The single-only reason wins (D-2) |
| The `depiction` pill (three seeded Properties) | `valueTypeForeground` | Its new pigment (AC-3.1) |

### Scale

None — one lookup in an eight-row list per draw.

### Edges

- **Assumptions.** `PVField`'s `hint` is `(any PVCopy)?` and `String` is `PVCopy`. `holdsRow`'s accessibility group wraps the radios only. `PropertyValueType` lists every Go type (`swift_vocabulary_test.go` holds it).
- **Neighbors.** A user `name` Property keeps its radios (AC-2.2); a seeded non-single-only one keeps "Set by Provenencia" (AC-2.3); choosing a non-single-only type keeps the draft's cardinality (AC-1.3).
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
| Value type picker | `PVSegmentedControl` (Ship) | `PropertiesView.createForm` | `select` per D-3 |
| Holds field (create) | `PVField` + `PVRadio` (Extend) | same | D-5 |
| Holds row (inspector) | `holdsRow` locked branch (Extend) | `PropertiesView` | D-6 |
| Locked value | `PropertiesLockedValue` (snowflake, new) | `PropertiesView.swift` | D-4 |
| Value type pill | `PropertiesValueTypePill`, `valueTypeForeground` (Extend) | `PropertiesSnapshot.swift`, `SubjectTypeChrome.swift` | D-8 |

Frames, in words (from the agreed board; the Properties pane with the Person type selected, and the table holding Name, Birth, Occupation, Likeness, Languages spoken, Can read, Age at event, Event type, Event place, Picture):

1. **Create, a single-only type.** New property dialog (560pt) over the page: Label "Can read", Key `can-read`, Value type `boolean` selected among `text`, `integer`, `date`, `name`, `subject`, `boolean` (segments share the row, none truncated). Holds: "One value" and the lock, no radios, in the space the radios take; hint "Yes / no Properties hold one value". Bind to subject types unchanged below it.
2. **Create, back to text.** Frame 1 with `text` chosen: the two radios, One value selected, hint "Facts that change over time, like occupation or residence, usually belong in events". Nothing else moves.
3. **Inspector, Can read.** Value type row: the `boolean` pill in verdigris and its lock. Holds: the locked row ("One value" and the lock), no radios, no notice. The lock's tooltip and VoiceOver: "Yes / no Properties hold one value".
4. **Inspector, Languages spoken.** A user `text` Property: both Holds radios, Several values selected, as today.
5. **Table pills.** All properties: every key in its pigment; Likeness and Picture's `depiction` in madder. Likeness selected: its Holds locked with "Set by Provenencia".
6a, 6b. **Wide** of 2 and 4: the table takes the width, the inspector stays 340pt, the dialog stays 560pt and centred.

Dark: the board's theme switch shows every frame with the 300 shades.

### Strings

New: `properties.inspector.holdsSingleOnly` (D-7). Removed: none.

## 7. Steps

**STEP-1 — The reason's copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-7.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — The model reads single-only**
Files: `macos/App/Features/Properties/PropertiesSnapshot.swift`, `PropertiesModel.swift`, `PropertiesView.swift` (the two draft call sites only); `macos/ProvenenciaTests/PropertiesModelTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-3, D-9. Add T-1 – T-3.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes (macOS only; `scripts/check-pr.sh` runs it).

**STEP-3 — The create form and the inspector**
Files: `macos/App/Features/Properties/PropertiesView.swift`. Skill: `add-ui-component`.
Do: D-4, D-5, D-6.
Check: `xcodebuild -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' build` succeeds with no Swift warning; manual checks 1–3.

**STEP-4 — Pill pigments, typed**
Files: `macos/App/Features/Catalog/SubjectTypeChrome.swift`, `PropertiesView.swift` (the `"term"` literal). Skill: `add-ui-component`.
Do: D-8.
Check: the build succeeds; manual check 4.

**STEP-5 — Tests pass whole**
Files: none new.
Do: run the suite.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-6 — Docs**
Files: `../deployment-plan.md` (S14-39 row status), `../completed.md`, this plan, `../design/S14-D9-properties-single-only.md` → `../design/archive/`, `../design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-39-properties-single-only.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits. STEP-1 and STEP-2 are one commit (the model and its copy); STEP-3, STEP-4 and STEP-6 one each.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PropertiesModelTests` `singleOnlyCreateHoldsOneValue` *(new)* | `store.valueTypeRules = goValueTypeRules`, warm; `holdsLockReason(valueType: "integer", origin: user)` = `holdsSingleOnly(typeLabel: L10n.string(L10n.Properties.valueTypeInteger))`; draft `multiple` → `chooseValueType("integer", singleOnly: true)` → `single`; → `chooseValueType("name", singleOnly: false)` → still `single`, reason nil; submit `integer` → stored `single` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1, D-2, D-3, D-7 |
| T-2 | Swift | `PropertiesModelTests` `singleOnlyUserPropertyHoldsIsLocked` *(new)* | rules seeded; `userProperty(valueType: "integer")` selected → `canEditCardinality` false, `setCardinality("multiple")` false, store unchanged; `userProperty(valueType: "name")` → true; `nameProperty()` (seeded `name`) → false, reason `holdsLocked` | AC-2.1, AC-2.2, AC-2.3, D-2 |
| T-3 | Swift | `PropertiesModelTests` `singleOnlyWaitsForTheRules` *(new)* | no rules (FakeStore default); `userProperty(valueType: "integer")` → `canEditCardinality` true; `holdsLockReason(valueType: "integer", origin: user)` nil | AC-1.5, AC-2.4 |
| — | manual | §10 | Holds field, inspector, pills, picker, Wide | AC-3.1, AC-3.2, AC-3.3, SO-5, SO-6, SO-W1, D-4, D-8 |

T-1 – T-3 are model-level: FakeStore does not enforce the rule (Out); Go's refusal is S14-36's `TestValueTypeRules`. Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PropertiesModelTests` (the fixture change), cardinality tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `rg -n "toggleDraftBind|valueType == \"term\"" macos .agents docs/*.md` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the S14-D9 board and inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — patch
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S14-39 row status Landed #N; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc: none changes (the single-only rule is already in interpretation-layer-data-model.md)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with a local, **uncommitted** edit to `core/propertymeta/propertymeta.go`: the `ValueTypeBoolean` row gains `ResearcherCreatable: true`. Run the app (⌘R): the **Build Go core** phase rebuilds `libprovenencia.dylib` from it (README). Open a project with a Person type.

1. New property: the picker offers `text`, `integer`, `date`, `name`, `subject`, `boolean`, no segment truncated. Choose `integer`, pick Several values, then choose `boolean`: Holds reads "One value" with the lock and the hint "Yes / no Properties hold one value". Choose `text`: the radios are back, One value selected. (AC-1.2, AC-1.3, SO-1, SO-2, SO-6)
2. Create "Can read" as `boolean`; the inspector's Holds is locked; VoiceOver on the lock reads "Yes / no Properties hold one value". (AC-1.4, AC-2.1, SO-3)
3. Select a user `text` Property: the radios work as before. Select a seeded Property: "Set by Provenencia". (AC-2.2, AC-2.3, SO-4)
4. The table and inspector show every key in D-8's pigment, `boolean` verdigris and `depiction` (Likeness) madder, in light and in dark (the 300 shades). (AC-3.1, AC-3.3, SO-5)
5. Widen past 1600pt of pane: the dialog and inspector look the same. (SO-W1)

Then `git checkout core/propertymeta/propertymeta.go` before committing.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Holds locked before the rules load | D-1 reads nil with no rules; Go refuses; AC-1.5, AC-2.4 |
| A seeded Property's lock text changes | D-2 keeps `holdsLocked` for non-single-only; T-2 |
| The local Go flip is committed | §10 ends with the checkout; the DoD's Out check |

**Stop and report if:**

- Manual check 1 shows a truncated segment in the six-type picker at the dialog's width (SO-6). Do not shorten the keys (D-10).
- `PropertyValueType` has a case not in Go's `propertymeta.Types`, or misses one, at your base (D-8 would be wrong).

## 12. Open questions

None. The board is agreed; `review-plan` round 2 runs before Planned.

## 13. PR

**Title:** S14-39 — Properties page reads single-only
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-39-properties-single-only.md`.
