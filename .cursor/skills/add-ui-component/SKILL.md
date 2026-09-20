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
macOS thin-client rules: [`.cursor/rules/macos-client.mdc`](../../rules/macos-client.mdc).

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

**Stop at the highest layer you need.** A view may call `.pvConfirmSheet` / `PVButton`
directly. Do not invent a recipe or snowflake shell for completeness.

**Layer roots + one folder per control.** No `Components/Core|Forms|…`. No loose
`.swift` directly under `Components/` / `Recipes/` / `Snowflakes/`.

| Layer | Home | Carries | Must not |
| --- | --- | --- | --- |
| Component | `DesignSystem/Components/<Name>/` | Tokens, slots, interaction chrome | Catalog models, `GenealogyStore`, feature meaning |
| Recipe | `DesignSystem/Recipes/<Name>/` | Domain → component props | Reimplement badge/chip/panel chrome; category nesting |
| Snowflake | `Features/<Feature>/` or `DesignSystem/Snowflakes/<Name>/` | One-screen glue | Become the de-facto dialog/menu pattern while one-off |

Folder name: PascalCase **without** `PV` (`Button`, `EvidenceIcon`). Public
types under `Components/` and `Recipes/` **must** use the `PV` prefix
(`PVButton`, `PVEvidenceIcon`) so they do not collide with SwiftUI/AppKit and
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
- [ ] Grep’d DesignSystem + Features for an existing near-match
- [ ] Composed down instead of cloning sideways
- [ ] No catalog types in Components/
- [ ] User copy via add-localized-string; a11y ids at call sites when UI-tested
- [ ] #Preview with static sample data for new DesignSystem types
- [ ] New .swift registered in the Xcode project (pbxproj)
- [ ] Second call site? Snowflake → Recipes/<Name>/ (or Components/<Name>/ if agnostic)
```

## Dialogs / sheets (current kit)

| Need | Use |
| --- | --- |
| Short create/edit form | `.pvDialog` / `PVDialogContent` |
| Destructive / irreversible, plain text | `.pvConfirm` (system alert) — preferred when enough |
| Same, rich detail (key chip, list) | `.pvConfirmSheet(item:)` — snapshot via `item:`, not Bool |
| One-off body | Pass a snowflake into the DS body/detail slot |
| Custom chrome that copies header/footer by hand | **Don’t** — compose panel/composite; only the unique body stays snowflake |

macOS sheets: **no scrim/dim**, no redrawing window corner radius/shadow on panel content.
Dismissal for async work stays on the caller binding (`isRunning` can show). See
`PVConfirm.swift` / DesignSystem README “Confirmations are system chrome.”

## Anti-patterns

```swift
// ❌ BAD — domain meaning inside Components/, or category nesting / loose root file
// DesignSystem/Components/Research/PVSourceTypeBadge.swift
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
.pvDialog(isPresented: $open, copy: …, onConfirm: …) { artifactForm }
```

```swift
// ❌ BAD — preemptive recipe with one caller
enum VocabularyDeleteDialog { static func sheet(…) }  // only SourceFields uses it

// ✅ GOOD — wait for second call site; use .pvConfirmSheet at the view today
```

## Steps when adding a component

1. Confirm layer = component (decision step 2).
2. Create **`DesignSystem/Components/<Name>/`** and add `PV<Name>.swift` (type
   `PV<Name>` — required prefix; plus optional README/helpers in the same folder).
3. Follow `PVButton.swift` shape: header names mirrored `.jsx` / deliberate deviations; tokens only; `#Preview` at bottom.
4. Register new files in `macos/Provenencia.xcodeproj`.
5. Prefer View modifiers for presentation (`.pvDialog`, `.pvConfirmSheet`) when the control is a sheet/alert family.
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

- [`evaluate-ui-component`](../evaluate-ui-component/SKILL.md) — point at an existing type; cousins + compose-down + small PR plan (no drive-by mega-refactor)
- [`add-localized-string`](../add-localized-string/SKILL.md) — all user-facing copy
- [`review-app-health`](../review-app-health/SKILL.md) — dimension 12 audits this layering across the app
- Dialog/token details: DesignSystem README; [`docs/design-system-layers.md`](../../../docs/design-system-layers.md)
