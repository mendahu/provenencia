# Ship log — S14-39 Properties page reads single-only

**Plan:** [`S14-39-properties-single-only.md`](S14-39-properties-single-only.md) · **Branch:** `s14-39-properties-single-only` · **PR:** #737
**State:** ready

## Round 0 — implement
Head `8d14c663`. Gate PASS (run by the orchestrator on the pushed head). Report: two `fileprivate` helpers the plan did not name, `holdsValueText(_:)` and the `LockSpeech` modifier, sit beside `PropertiesLockedValue`; manual check not run (needs the app with the local Go flip), so the picker-truncation stop condition is untested.

## Round 1 — review
Base `82547284`, head `8d14c663`. Reviewers: CON, MNT, UX (`macos/App/**`, `Localizable.xcstrings`), SEC+PRF (`macos/App/Features/**` models, performance paths; no security path, dispatched as one reviewer). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 (CON-6, MNT-12) | CON | major | No Landed note; PR body says "Deviations: none" | Fix now → F-1 (PR body at finish) | verified: the plan diff changes only Status and Brief; round 0 reported two unnamed helpers; Scope line 76 says `singleOnly:` where D-3 says `holdsLocked:` |
| UX note → CON-7 | CON | major | The dialog is 480pt; the plan's frames 1 and 6a and the board say 560pt | Fix now → F-3 | verified: `PropertiesView.swift:722` passes no `width`, and `pvFormDialog`'s default is 480; DoD "UI matches the board"; SO-6 needs the room |
| CON-2 | CON | major | AC-1.3 is not proved: T-1 never leaves a locked type | Fix now → F-4 | verified: T-1 goes name → integer → submit |
| MNT-1 | MNT | major | `LockSpeech` repeats `accessibilityHidden(_:)` | Fix now → F-5 | verified: one call site; D-4 names no such type |
| MNT-3 (CON-4, MNT-4) | MNT | major | T-4 hand-rolls colour resolution, proves only distinctness, and lists types by hand | Fix now → F-6 | verified: `PVPainter.color(_:in:)` exists; the vocabulary test's `rawCase` regex reads cases, so `CaseIterable` is safe |
| MNT-2 | MNT | major | `holdsRow` asks the lock rule twice, with an empty branch | To user → Fix now → F-2 | the user chose one read (2026-11-07); it supersedes the shape D-6 spells, with no visible change |
| MNT-5 | MNT | minor | The reason's copy splices a translated label | To user → Fix now → F-10 | the user kept the copy and asked to record the key in the idea file's splice list (2026-11-07) |
| CON-3 | CON | minor | No test pins D-2's precedence | Fix now → F-7 | verified: no test has a single-only type with a non-user origin |
| UX-1 | UX | minor | The locked Holds value in create has no spoken field name | Fix now → F-8 | verified: the radios' group label is hidden in that branch; D-5 expects "Holds, One value" |
| UX-2 | UX | minor | "Immutable after create" is now a visible tooltip | Fix now → F-9 | verified: D-4 made it visible; the plain sentence is already `createValueTypeHint` |
| MNT-6 | MNT | minor | The picker's reset is wired in a view closure | Rejected | D-3 spells the closure; plan review X2-6 declined this |
| CON-5, MNT-7 | CON, MNT | nit | `}}` on one line | Fix now → F-11 | this PR's slip; three reviewers noticed it |
| MNT-8 | MNT | nit | Stale `setCardinality` comment | Fix now → F-11 | stale comment in a changed file |

Nits (logged): `UX-3 · PropertiesView.swift PropertiesLockedValue · .help plus a matching label may repeat under VoiceOver; check in manual check 2` · `UX-4 · PropertiesView.swift:870 · from S14-37 the picker segment speaks the raw key "boolean"; D-10 keeps keys` · `MNT-9 · PropertiesView.swift:881 · typed text case` · `MNT-10 · PropertiesView.swift:759-782 · one PVField if the hints unify` · `MNT-11 · PropertiesView.swift:962 · PVSymbol.lock.rawValue`

SEC+PRF: no findings.

### Fix plan 1

Plan Decisions still bind, except F-2, which the user approved over D-6's shape.

**F-1 — Write the Landed note last** (from CON-1, CON-6, MNT-12)
Files: `docs/deployment-plan/spike-14/prs/S14-39-properties-single-only.md`.
Do this item after all the others. Add a `**Landed:**` header line, the shape of `S14-33-auto-file-bridges.md:16`, listing every difference from the plan: `holdsValueText(_:)`, a helper the plan did not name; `holdsRow` reads `holdsLockReason` once (F-2, user 2026-11-07, supersedes D-6's shape); the dialog passes `width: 560` (F-3); `valueTypeImmutable`'s default changed (F-9); `PropertyValueType` gained `CaseIterable` (F-6); Scope's `chooseValueType(_:singleOnly:)` is `holdsLocked:` as D-3 says. Do not edit Decisions, Scope, or ACs.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-39-properties-single-only.md` reports no errors.

**F-2 — One read of the lock rule in `holdsRow`** (from MNT-2)
Files: `macos/App/Features/Properties/PropertiesView.swift`.
Failing test first: none (no behaviour change; T-2 holds it).
Do: `if let reason = model.holdsLockReason(valueType: property.valueType, origin: property.origin) { the inspectorMetaRow locked row } else { the radios }`. `canEditCardinality` stays as `setCardinality`'s guard.
Check: `singleOnlyUserPropertyHoldsIsLocked` passes; the `else if let` is gone.

**F-3 — The create dialog is the board's 560pt** (from the UX reviewer's note, CON-7)
Files: `macos/App/Features/Properties/PropertiesView.swift`.
Do: pass `width: 560` to `.pvFormDialog` in `PropertiesCreateHost`, the kit's `width:` parameter. No kit change.
Check: builds; `rg -n "width: 560" macos/App/Features/Properties/PropertiesView.swift` finds it.

**F-4 — T-1 leaves a locked type** (from CON-2)
Files: `macos/ProvenenciaTests/PropertiesModelTests.swift`.
Test (it passes today, so this one does not run red; do not report it Not reproduced): in `singleOnlyCreateHoldsOneValue`, after the `integer` expects and before the submit, on a copy: `var back = draft`, `back.chooseValueType("name", holdsLocked: model.holdsLockReason(valueType: "name", origin: CatalogOrigin.user) != nil)`, expect `back.valueType == "name"` and `back.cardinality == "single"`.
Check: the test passes.

**F-5 — Delete `LockSpeech`** (from MNT-1)
Files: `macos/App/Features/Properties/PropertiesView.swift`.
Failing test first: none (accessibility modifiers; behaviour unchanged).
Do: on the lock image, `.accessibilityLabel(Text(verbatim: reason)).accessibilityHidden(!speaksReason)`; delete `LockSpeech`.
Check: `rg -n "LockSpeech" macos` finds nothing.

**F-6 — T-4 pins D-8's table through the kit's resolver** (from MNT-3, CON-4, MNT-4)
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `macos/ProvenenciaTests/PropertiesModelTests.swift`.
Do: add `CaseIterable` to `PropertyValueType`, as `NamePartType` has. Rewrite `valueTypeInksAreDistinct` as `throws`: a table of `(PropertyValueType, light: Color, dark: Color)` from D-8 (`text` paper700 / paper300 … `term` ochre900 / ochre300 …, the `PVPalette` values); `#expect` the table's types equal `Set(PropertyValueType.allCases)`; under each of `.aqua` and `.darkAqua`, from `try #require(NSAppearance(named:))`, compare `PVPainter.color(PropertiesTypeChrome.valueTypeForeground(type.rawValue), in: appearance)` with `PVPainter.color(<table shade>, in: appearance)` in sRGB; then expect eight distinct colours per appearance. Keep the `"nope"` fallback. No `performAsCurrentDrawingAppearance` and no force unwrap in the test. Keep the test's name.
Check: T-4 passes; `go test ./api/ffi/ -run Vocabulary` passes (with the repo's tags).

**F-7 — Pin D-2's precedence** (from CON-3)
Files: `macos/ProvenenciaTests/PropertiesModelTests.swift`.
Test (passes today, same note as F-4): in `singleOnlyUserPropertyHoldsIsLocked`, `#expect(model.holdsLockReason(valueType: "integer", origin: CatalogOrigin.provenencia) == L10n.Properties.holdsSingleOnly(typeLabel: L10n.string(L10n.Properties.valueTypeInteger)))`, using the repo's seeded origin constant.
Check: the test passes.

**F-8 — The locked Holds value names its field** (from UX-1)
Files: `macos/App/Features/Properties/PropertiesView.swift`.
Do: in `holdsField`'s locked branch only, beside the identifier: `.accessibilityLabel(Text(L10n.Properties.holds))` and `.accessibilityValue` the One / Several text for the draft's cardinality. Not `.pvFieldNamed()`. The inspector rows do not change.
Check: builds.

**F-9 — The Value type lock's tooltip in plain words** (from UX-2)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: `properties.inspector.valueTypeImmutable`'s default becomes "Cannot be changed once the property exists", the words of `createValueTypeHint`.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**F-10 — Record the splice** (from MNT-5, user 2026-11-07)
Files: `docs/ideas/international-names.md`.
Do: under "Whole-phrase copy, app-wide", add `properties.inspector.holdsSingleOnly` (a value type label inside "… Properties hold one value") to the sentence listing copy that still splices a word.
Check: `rg -n "holdsSingleOnly" docs/ideas/international-names.md` finds it.

**F-11 — Two small slips** (from CON-5, MNT-7, MNT-8)
Files: `macos/App/Features/Properties/PropertiesView.swift`, `PropertiesModel.swift`.
Do: split `    }}` onto two lines. `setCardinality`'s doc comment: "Seeded, plugin and single-only Properties stay fixed."
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

Update the `completed.md` paragraph if a fix changed what shipped.

## Round 1 — fix
Head `cef1ffb3`. Gate PASS (run by the orchestrator on the pushed head). F-1 to F-11 done; F-4 and F-7 passed on first run, as the fix plan said. `holdsValueLabel(_:)` was split out of `holdsValueText(_:)` for F-8. UX-1 **Resolved**: the locked create value has the Holds label and the One / Several value. UX-2 **Resolved**: the default and the catalog entry read "Cannot be changed once the property exists". Both read in the fix commits by the orchestrator.

## Round 2 — review
Scope `8d14c663...cef1ffb3`. Reviewers: CON, MNT (each had a Fix-now major). Not sent: UX (minors only, resolved above); SEC+PRF (no findings).
Both reviewers: no new blocker; every round-1 major resolved. No Fix plan 2.

## Result
Gate PASS on `cef1ffb3` (the only later commit is this log) · main unchanged since the branch was cut · open: none · follow-ups: none · to user: MNT-2 and MNT-5, both answered 2026-11-07. Manual check (plan §10) not run: the SO-6 truncation stop condition, AC-3.3 and "UI matches the board" wait on it.
