---
name: add-ui-component
description: >-
  Classifies and adds Provenencia macOS UI under design-system / recipe /
  snowflake layers (Frost). Use when adding or changing PV* components,
  DesignSystem chrome, dialogs, confirms, sheets, badges, chips, feature-private
  panels, or when tempted to fork a slightly different control instead of
  composing an existing one. Also use for design-system hardening, component
  layering, recipes vs snowflakes, or promoting a one-off into a shared control.
---

# Add or classify a UI component

Provenencia UI is layered so LLM- and rush-driven forks don’t accumulate. Read
[`docs/design-system-layers.md`](../../../docs/design-system-layers.md)
for the full model. Kit conventions: [`macos/App/DesignSystem/README.md`](../../../macos/App/DesignSystem/README.md).
macOS thin-client rules: [`.cursor/rules/macos-client.mdc`](../../../.cursor/rules/macos-client.mdc).

## Decision (do this first)

```
1. Does an existing component or recipe already cover this with props/slots?
   → Compose it. Stop.
2. Is the control content-agnostic (would work in another app with our tokens)?
   → Component. Place under DesignSystem/Components/<Name>/.
3. Is it Provenencia-specific AND needed in ≥2 real call sites?
   → Recipe. Place under DesignSystem/Recipes/<Name>/.
4. Otherwise
   → Snowflake under Features/<Feature>/ (prefer private) or DesignSystem/Snowflakes/<Name>/.
```

**Stop at the highest layer you need.** A view may call `.pvConfirm` / `PVButton`
directly. Do not invent a recipe or snowflake shell for completeness.

**Layer roots + one folder per control.** No `Components/Core|Forms|…`. No loose
`.swift` directly under `Components/` / `Recipes/` / `Snowflakes/`.

| Layer | Home | Carries | Must not |
| --- | --- | --- | --- |
| Component | `DesignSystem/Components/<Name>/` | Tokens, slots, interaction chrome | Catalog models, `GenealogyStore`, feature meaning |
| Recipe | `DesignSystem/Recipes/<Name>/` | Domain → component props | Reimplement badge/chip/panel chrome; category nesting |
| Snowflake | `Features/<Feature>/` or `DesignSystem/Snowflakes/<Name>/` | One-screen glue | Become the de-facto dialog/menu pattern while one-off |

Folder name: PascalCase **without** `PV` (`Button`, `SectionHeader`). Public
types under `Components/` and `Recipes/` **must** use the `PV` prefix
(`PVButton`, `PVSectionHeader`) so they do not collide with SwiftUI/AppKit and
read as kit API — see [`docs/design-system-layers.md`](../../../docs/design-system-layers.md)
§ PV type prefix. Colocate the main Swift file, private helpers, and
component-specific docs in that folder.

### Primitive vs composite (inside Components/)

Same layer — different size; **each still gets its own `<Name>/` folder**:

- **Primitive:** button, badge, input, field, sheet **panel** (title / subtitle / body / optional footer).
- **Composite:** confirm, form dialog — prescribed use of the panel (roles, focus, presentation) still **domain-agnostic**.

Prefer extending a composite with slots over a new root `PV*` for a slight variant.

## Checklist

```
- [ ] Classified: component | recipe | snowflake
- [ ] Public Components/Recipes type uses `PV*` prefix (not bare `Button` / SwiftUI-colliding names)
- [ ] Path is Components|Recipes|Snowflakes/<Name>/… (or Features/) — no category nesting, no loose layer-root .swift
- [ ] README row: a new `Components/<Name>/` or `Recipes/<Name>/` folder has a row in the DesignSystem README's built table (`scripts/check-guidance-names.py` fails without it)
- [ ] Grep’d DesignSystem + Features for an existing near-match
- [ ] Composed down instead of cloning sideways
- [ ] No catalog types in Components/
- [ ] User copy via add-localized-string; a11y ids at call sites when UI-tested
- [ ] #Preview with static sample data for new DesignSystem types
- [ ] File saved under `macos/App/DesignSystem/` or `macos/App/Features/` (synchronized folders; no `project.pbxproj` edit)
- [ ] Second call site? Snowflake → Recipes/<Name>/ (or Components/<Name>/ if agnostic)
- [ ] Each behavior the README promises for it (Return, Escape, dismissal, focus, VoiceOver announcement) is in its code and held by a test. `PVConfirm` promised that Return could not confirm a delete while its Delete button kept the default-action shortcut
```

## Dialogs / sheets (current kit)

| Need | Use |
| --- | --- |
| Short create/edit form | `.pvFormDialog` / `PVFormDialogContent` (on `PVPanel`) |
| Destructive / irreversible confirm | `.pvConfirm(item:)` — snapshot via `item:`, not Bool; detail slot for key chip / callout |
| Resource delete with inbound check | **DeleteImpact recipe** (S8-D9 / **S8-13**): confirm when `Impact.allowed`, notice (not Confirm-as-acknowledge) when blocked. Do not hand-build a second inbound list. |
| One-off body | Pass a snowflake into the DS body/detail slot |
| Custom chrome that copies header/footer by hand | **Don’t** — compose `PVPanel` / FormDialog / Confirm; only the unique body stays snowflake |

**Every sheet presents through `.pvSheet(item:)` / `.pvSheet(isPresented:)`** (`Components/Panel/PVSheet.swift`),
never a raw `.sheet`. It gives the sheet Escape, a click outside, and the
discard question for a dirty form, all from one rule (`PVSheetDismiss`). Do not
add `.pvOutsideClick` or `.keyboardShortcut(.cancelAction)` to a sheet yourself.

macOS sheets: **no scrim/dim**, no redrawing window corner radius/shadow on panel content.
Dismissal for async work stays on the caller binding (`isRunning` can show). See
`PVConfirm.swift` / DesignSystem README “Confirmations are system chrome.”

## Anti-patterns

```swift
// ❌ BAD — domain meaning inside Components/, or category nesting / loose root file
// DesignSystem/Components/Forms/PVSourceTypeBadge.swift
// DesignSystem/Components/PVSourceTypeBadge.swift

// ✅ GOOD — recipe, per-component folder
// DesignSystem/Recipes/SourceTypeBadge/SourceTypeBadge.swift
struct SourceTypeBadge: View {
    let type: CatalogSourceType
    var body: some View { PVBadge(…) }
}
```

```swift
// ❌ BAD — new dialog stack for one screen
struct AddArtifactSheet: View { /* custom scrim, footer, buttons */ }

// ✅ GOOD — component form dialog + snowflake form body
.pvFormDialog(isPresented: $open, copy: …, onConfirm: …) { artifactForm }
```

```swift
// ❌ BAD — preemptive recipe with one caller
enum VocabularyDeleteDialog { static func sheet(…) }  // only Metadata uses it

// ✅ GOOD — wait for second call site; use .pvConfirm at the view today
```

## Steps when adding a component

1. Confirm layer = component (decision step 2).
2. Create **`DesignSystem/Components/<Name>/`** and add `PV<Name>.swift` (type
   `PV<Name>` — required prefix; plus optional helpers in the same folder; a README there is listed in the folder's membership exceptions, which `check-guidance-names.py` enforces).
3. Follow `PVButton.swift` shape: header names mirrored `.jsx` / deliberate deviations; tokens only; `#Preview` at bottom.
4. Save the file in its folder; the synchronized folder adds it to the build. A non-Swift file (a README) in the folder needs a membership exception in `project.pbxproj` (see the DesignSystem README § Xcode project registration).
5. Prefer View modifiers for presentation (`.pvFormDialog`, `.pvConfirm`) when the control is a sheet/alert family.
6. Leave `.accessibilityIdentifier` to call sites unless the control owns fixed chrome ids by documented convention.

## Steps when adding a recipe

1. Confirm ≥2 real call sites (or an imminent second in the same change).
2. Create **`DesignSystem/Recipes/<Name>/`** and add the primary Swift file (+ colocated docs/helpers).
3. Keep it a thin map: domain value → existing component props.
4. Do not duplicate sunken footers, focus rings, or menu hosts — those stay in `Components/`.
5. If only one caller exists, leave a snowflake and note “promote when reused.”

## Steps when adding a snowflake

1. Prefer `Features/<Feature>/` (`private`); use `DesignSystem/Snowflakes/<Name>/` only for named kit-side one-offs.
2. Prefer filling a component slot over owning sheet chrome.
3. If a second feature needs it, promote to `Recipes/<Name>/` (or `Components/<Name>/` if agnostic) before copying.

## Related skills

- [`add-design-brief`](../add-design-brief/SKILL.md) — Claude Design brief + kit inventory before a UI PR
- [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md) — point at an existing type; cousins + compose-down + small PR plan (no drive-by mega-refactor)
- [`add-localized-string`](../add-localized-string/SKILL.md) — all user-facing copy
- [`review-pr`](../review-pr/SKILL.md) — the maintainability lens (Q5) checks this layering on every PR; its area audit checks it across the app
- Dialog/token details: DesignSystem README; [`docs/design-system-layers.md`](../../../docs/design-system-layers.md)
