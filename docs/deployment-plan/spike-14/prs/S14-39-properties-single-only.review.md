# Review — S14-39 Properties page reads single-only

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet). Both read the plan and brief S14-D9.

## To user

- None open. The product calls were asked before writing (2026-11-07): read-only "One value" in create, the inspector's locked row with its own reason, raw key `boolean`.

## Writer's own calls

- **D-2.** A single-only type's reason wins over a seeded origin's "Set by Provenencia". No seeded Property is single-only today.
- **D-8 / SO-5.** `depiction` gets a pigment too. The typed switch needs one, and it has been grey by accident. That makes the PR a **patch** bump, not none.
- **D-9.** The tests use `integer` as the single-only type in a neutral fixture, not `boolean`.

## Round 1

### Maintainability

**M-1 · blocker · Two lookups into the rules (`isSingleOnly` here, a dictionary in S14-37).** **Answer:** Fixed. D-1 is `PropertiesSnapshot.valueTypeRule(_:)`. S14-37 D-6 now reads it through the composer's kept `fields` snapshot, and its dictionary is gone (same commit).

**M-2 · blocker · `case "boolean":` in a string switch, against `.cursor/rules/macos-client.mdc`.** **Answer:** Fixed.
- D-8 switches exhaustively on `PropertyValueType`.
- The `"term"` literal is typed.
- `depiction` gets a pigment (brief SO-5).

**M-3 · major · The lock reason is built in two places.** **Answer:** Fixed. D-2 makes `holdsLockReason(valueType:origin:) -> String?` the one rule, and `canEditCardinality` derives from it.

**M-4 · major · Three hand copies of the lock.** **Answer:** Fixed. D-4 adds `PropertiesLockedValue` for the Value type row, the inspector Holds row and the create Holds field. The image stays the same; the agreed board adds the reason as the lock's tooltip on all three, which is the one visible change.

**M-5 · major · An `inout` model method beside dead `toggleDraftBind`.** **Answer:** Fixed.
- D-3 puts `chooseValueType` and `toggleBind` on `Draft`.
- `toggleDraftBind` is deleted, and the DoD sweep names it.

**M-6 · minor · A second rules fixture.** **Answer:** Fixed. D-9 extends `goValueTypeRules` with `integer` single-only.

**M-7 · minor · The brief ignores the read-only `PVInput` look.** **Answer:** Declined, with evidence. The user chose the inspector row's look (2026-11-07). The brief's inventory row now says so, and names the read-only Input as rejected.

**M-8 · minor · SO-5 leaves out `depiction`.** **Answer:** Fixed. SO-5 and frame 5 include it.

### Executability

**X-1 · major · STEP-1 uses copy from STEP-2.** **Answer:** Fixed. Copy is STEP-1, and the commits are mapped.

**X-2 · major · The `inout` call shape over an Optional `@State`.** **Answer:** Fixed by M-5. The call is `draft?.chooseValueType(…)`.

**X-3 · major · The create lock content and its label are unspecified.** **Answer:** Fixed. D-4 and D-5 give the shared view, the spoken reason, and the identifier `properties.create.holds.locked`.

**X-4 · minor · Who builds the type label.** **Answer:** Fixed by M-3: `holdsLockReason` builds it.

**X-5 · minor · Edges.** **Answer:** Fixed.
- A nil reason means editable.
- An unknown type is not locked.
- The pre-load race is answered in Reach: the picker offers only `text` before the rules load.
- `submitCreate` does not re-force cardinality (D-3), and T-1 does not rely on it.

**X-6 · minor · Test helpers.** **Answer:** Fixed. D-9 gives `userProperty` a `valueType` parameter, and T-2 uses `nameProperty()` for the seeded case.

**X-7 · minor · FakeStore does not enforce the rule.** **Answer:** Fixed. §8 says T-1 – T-3 are model-level, and that Go's refusal is S14-36's test. FakeStore enforcement is Out.

**X-8 · major · No route for the manual check.** **Answer:** Fixed. §10 uses a local, uncommitted flip of `propertymeta.Types`, rebuilt by Xcode's Build Go core phase on ⌘R, then reverted.

**X-9 · minor · The SO-6 stop cannot be checked.** **Answer:** Fixed. The stop is tied to manual check 1, which has six segments with the flip, and says not to shorten keys.

**X-10 · minor · No build command.** **Answer:** Fixed. STEP-2 and STEP-3 give the `xcodebuild` lines from the README.

**X-11 · nit.** Logged.

**X-12 · nit · Temptations.** **Answer:** Fixed. Out lines cover the label fallback for an unknown key and the stale `CatalogProperty.valueType` comment.

## Round 2 (after the board was agreed)

### Maintainability

**M2-1 · blocker · Two readers of the single-only rule; the create field shows a fixed "One value".** **Answer:** Fixed.
- D-3's picker resets cardinality from `holdsLockReason(...) != nil`, the same function that draws the lock. `Draft.chooseValueType` takes `holdsLocked:`.
- D-5 draws the draft's own cardinality.

**M2-2 · blocker · Eight `pvDynamic` pairs inline in feature code.** **Answer:** Fixed. D-8 adds eight `PVColor.valueType…Ink` tokens in `Tokens/PVColor.swift`, beside the subject-kind inks; `valueTypeForeground` returns them. STEP-4 names the file.

Below the bar, also taken:
- The single `PropertiesView.toggleBind` forwarder is deleted (D-3).
- The two `L10n` comments now say tooltip (§6 Strings).

### Executability

**X2-1 · minor · Picker segments have no pigment.** **Answer:** Fixed. Out line: the kit control has no colour slot, and the board's picker is plain.

**X2-2 · minor · AC-3.1 and AC-3.2 are manual only.** **Answer:** Fixed. T-4 resolves the eight inks under aqua and darkAqua and checks the fallback.

**X2-3 · minor · T-1 does not submit the type it locks.** **Answer:** Fixed. T-1 chooses `name`, then `integer`, then submits that draft.

**X2-4 · minor · The reason is spoken twice in create.** **Answer:** Fixed. `PropertiesLockedValue(speaksReason: false)` in create, where the hint speaks it. The view is `.accessibilityElement(children: .combine)` so the identifier lands on one element.

**X2-5 · minor · Optional and argument shapes.** **Answer:** Fixed. D-5 and D-6 spell out both call sites with `if let`.

**X2-6 · minor · The select glue is untested.** **Answer:** Declined. After M2-1 the glue is one call to the tested `holdsLockReason`, and manual check 1 covers it.

**X2-7, X2-8 · nit.** Logged. The header and status are updated.

**Verdict after round 2:** no open blocker; Planned.
