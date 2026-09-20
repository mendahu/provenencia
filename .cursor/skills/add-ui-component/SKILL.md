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
[`docs/ideas/design-system-hardening.md`](../../../docs/ideas/design-system-hardening.md)
for the full rationale. Kit conventions: [`macos/App/DesignSystem/README.md`](../../../macos/App/DesignSystem/README.md).
macOS thin-client rules: [`.cursor/rules/macos-client.mdc`](../../rules/macos-client.mdc).

## Decision (do this first)

```
1. Does an existing PV* (or recipe) already cover this with props/slots?
   → Compose it. Stop.
2. Is the control content-agnostic (would work in another app with our tokens)?
   → Design system (primitive or composite). Place under DesignSystem/.
3. Is it Provenencia-specific AND needed in ≥2 real call sites?
   → Recipe. Shared product home; thin wrapper over PV*.
4. Otherwise
   → Snowflake under Features/<Feature>/ (prefer private). Do not add a PV*.
```

**Stop at the highest layer you need.** A view may call `.pvConfirmSheet` / `PVButton`
directly. Do not invent a recipe or snowflake shell for completeness.

| Layer | Home | Carries | Must not |
| --- | --- | --- | --- |
| Design system | `macos/App/DesignSystem/` | Tokens, slots, interaction chrome | Catalog models, `GenealogyStore`, feature meaning, hard-coded “Sources” |
| Recipe | Shared (e.g. `Components/Research/`) | Domain → `PV*` mapping | Reimplement badge/chip/panel chrome |
| Snowflake | `Features/<Feature>/` | One-screen glue | Become the de-facto dialog/menu pattern while private |

### Primitive vs composite (inside the design system)

Same layer — different size:

- **Primitive:** button, badge, input, field, sheet **panel** (title / subtitle / body / optional footer).
- **Composite:** confirm, form dialog — prescribed use of the panel (roles, focus, presentation) still **domain-agnostic**.

Prefer extending a composite with slots over a new root `PV*` for a slight variant.

## Checklist

```
- [ ] Classified: design system | recipe | snowflake (written down or in PR notes)
- [ ] Grep’d DesignSystem + Features for an existing near-match
- [ ] Composed down instead of cloning sideways
- [ ] Layer home matches table above (no catalog types in PV*)
- [ ] User copy via add-localized-string; a11y ids at call sites when UI-tested
- [ ] #Preview with static sample data for new DesignSystem types
- [ ] New .swift under DesignSystem registered in the Xcode project (pbxproj)
- [ ] Second call site? Snowflake → recipe (or into DS if it was agnostic)
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
// ❌ BAD — domain meaning inside DesignSystem
struct PVSourceTypeBadge: View {
    let type: CatalogSourceType  // catalog type in the kit
}

// ✅ GOOD — recipe (if reused) or call-site mapping onto PVBadge
struct SourceTypeBadge: View {  // Features or Research/
    let type: CatalogSourceType
    var body: some View { PVBadge(…) }
}
```

```swift
// ❌ BAD — new dialog stack for one screen
struct AddArtifactSheet: View { /* custom scrim, footer, buttons */ }

// ✅ GOOD — DS form dialog + snowflake form body
.pvDialog(isPresented: $open, copy: …, onConfirm: …) { artifactForm }
```

```swift
// ❌ BAD — preemptive recipe with one caller
enum VocabularyDeleteDialog { static func sheet(…) }  // only SourceFields uses it

// ✅ GOOD — wait for second call site; use .pvConfirmSheet at the view today
```

## Steps when adding a design-system type

1. Confirm layer = design system (step 2 in Decision).
2. Pick `Components/<Category>/` mirroring the web kit (`core`, `forms`, `feedback`, …).
3. Follow `PVButton.swift` shape: header names mirrored `.jsx` / deliberate deviations; tokens only; `#Preview` at bottom.
4. Add the file to `macos/Provenencia.xcodeproj` (not a synchronized root).
5. Prefer View modifiers for presentation (`.pvDialog`, `.pvConfirmSheet`) when the control is a sheet/alert family.
6. Leave `.accessibilityIdentifier` to call sites unless the control owns fixed chrome ids by documented convention.

## Steps when adding a recipe

1. Confirm ≥2 real call sites (or an imminent second in the same change).
2. Keep it a thin map: domain value → existing `PV*` props.
3. Do not duplicate sunken footers, focus rings, or menu hosts — those stay in DesignSystem.
4. If only one caller exists, leave a snowflake and note “promote when reused.”

## Steps when adding a snowflake

1. Colocate under `Features/<Feature>/`; mark `private` when possible.
2. Prefer filling a DS slot over owning sheet chrome.
3. If a second feature needs it, promote before copying.

## Related skills

- [`add-localized-string`](../add-localized-string/SKILL.md) — all user-facing copy
- [`review-app-health`](../review-app-health/SKILL.md) — dimension 12 audits this layering
- Dialog/token details: DesignSystem README; idea note above
