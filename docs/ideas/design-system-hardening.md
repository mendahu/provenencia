# Design system hardening

**Status:** idea only — not roadmapped. Spike 7 pauses here until this cleanup has a home (or we deliberately resume Spike 7 first).

## Problem

Rapid LLM-assisted UI work has left the macOS design system drifting:

- Near-duplicate components that differ only slightly from an existing `PV*` piece
- Whole new implementations when a call site did not match an existing API one-to-one
- Domain or screen chrome living where a reusable primitive should
- Composability suffering as every new view invents its own pattern

This will get worse as we add more workspace surfaces. Before more Spike 7 UI, we need shared standards for **how components are layered**, so maintainability holds long-term.

This note starts with the composition model. Later passes can inventory drift, refactor call sites, and tighten `macos/App/DesignSystem/README.md` / folder layout against these categories.

## Component layers

We follow Brad Frost’s **[components / recipes / snowflakes](https://bradfrost.com/blog/post/design-system-components-recipes-and-snowflakes/)** split — three layers of increasing product coupling, not a stack you must climb on every screen.

(An earlier draft used four layers by splitting “domain” and “view-local.” Those are the same *kind* of thing — product-aware UI — differing only by **reuse**. Frost already names that: recipes vs snowflakes. Collapsing to three keeps the decision tree honest.)

```
┌─────────────────────────────────────────────────────────┐
│  3. Snowflakes — one screen / feature-private           │  more bespoke
├─────────────────────────────────────────────────────────┤
│  2. Recipes — product-specific, reused across views     │
├─────────────────────────────────────────────────────────┤
│  1. Design system — content-agnostic PV* kit            │  more generic
└─────────────────────────────────────────────────────────┘
```

Every building block belongs in **exactly one** layer. Higher layers may compose lower ones; they must not reinvent them. Lower layers must not import higher ones (no catalog / genealogy types inside the design system kit).

**You stop at the highest layer you need.** Most call sites never touch all three:

- A view can drop in a **design-system** control directly (`.pvConfirmSheet`, `PVButton`) with copy and slots filled at the call site.
- A **recipe** exists only when the same Provenencia meaning is reused and deserves a named mapping.
- A **snowflake** exists only for glue unique to that screen.

Do **not** invent a recipe or snowflake shell “for completeness.” Extra wrappers without a second call site (or real product meaning) are the same drift we’re trying to stop.

When a new control is needed, ask which layer it is **before** writing a new file. Prefer composing downward over forking sideways.

### 1. Design system components

Shared, **content-agnostic** controls in `macos/App/DesignSystem/`. They would work if dropped into another app with the same tokens.

- **What they carry:** Provenencia theming (`PV*` tokens) and generic interaction chrome (slots for title, body, actions).
- **What they do not carry:** Source / Subject / catalog vocabulary, `GenealogyStore` types, or feature-specific meaning.
- **Inside the kit**, still distinguish *primitive* vs *composite* when useful — same layer, different size:
  - **Primitive:** button, badge, input, field, sheet **panel** (title / subtitle / body / optional footer).
  - **Composite:** confirm dialog, form dialog — prescribed recipes *of* the panel (button roles, focus defaults, presentation API) that remain domain-agnostic.

Composites are still design-system components in Frost’s sense: maximal reuse, no product entity types. Prefer flexible slots over forking a whole new root control for a slight variant.

**Examples:** `PVButton`, `PVBadge`, `PVInput`, `PVField`, `PVConfirm` / `PVDialog` (target: both on one panel primitive).

### 2. Recipes

Product-specific compositions used **consistently across more than one place**. Valuable and reusable, but not agnostic enough for the portable kit.

- Map Provenencia meaning → design-system props (icons, colors, grade, type key).
- Thin wrappers preferred — do not reimplement badge / chip / row chrome.
- Live in a shared product home when reused app-wide (e.g. `Components/Research/` or a clear domain module), not copy-pasted into each feature.

**Examples:** source-type / evidence-grade badge families; a shared “delete vocabulary row” confirm shell *if* several screens share the same wiring (otherwise the DS confirm + call-site copy is enough).

### 3. Snowflakes

Bespoke UI for **one** screen (or private helpers inside one feature). Written once, called from that feature only.

- Live under `Features/<Feature>/`, prefer `private` / file-scoped.
- **OK:** layout glue, one-off section chrome, a sheet body that will never be reused.
- **Not OK:** quietly becoming the de-facto way we do dialogs while living in one feature file — promote into the design system (or a recipe) instead.
- **Second call site promotes** a snowflake → recipe (same product kind, now shared) or further down into the DS if it was actually content-agnostic all along.

**Example:** `SourceTypeIconPickerSheet`’s icon grid (and any chrome that isn’t absorbed by a shared panel).

## Layering rules (starter)

1. **Stop at the layer you need.** Design-system controls are valid end products at a call site. Recipes and snowflakes are optional specialization.
2. **Compose down, don’t clone sideways.** A slightly different confirm is still DS confirm + props, not a new root dialog.
3. **Design system stays portable.** No `GenealogyStore`, catalog models, or feature `L10n` namespaces inside layer 1 (generic chrome labels owned by the control are fine).
4. **Reuse is what separates recipes from snowflakes** — same product-coupled kind; promote on the second real call site instead of pre-building a recipe.
5. **Design-system folder vs Features.** Layer 1 lives under `DesignSystem/`. Snowflakes stay under `Features/`. Recipes get an intentional shared home — don’t leave orphans.

## Dialogs as a worked example

Not every delete flow needs all three layers:

| Layer | Role | Required? |
|---|---|---|
| Design system | Sheet **panel** primitive; **confirm** and **form dialog** composites on top (cancel-first / primary confirm, `isRunning`, no scrim — window owns chrome on macOS). | Usually enough |
| Recipe | Only if several screens share the same catalog-shaped confirm wiring. | Optional |
| Snowflake | Icon picker grid, graph create form fields, etc., passed into a DS body slot — or omitted when slots + call-site copy are enough. | Optional |

Typical Source-fields delete today: **DS confirm at the view** (`.pvConfirmSheet` + `PVConfirmCopy` + optional `PVConfirmKeyChip`). That is fine. No recipe required unless the same wiring repeats.

Today `PVDialog` and `PVConfirm` sheet content duplicate panel chrome; the icon picker reimplements header/footer by hand. Hardening should collapse that toward the table above rather than adding another modal type.

## Open questions (later)

- Exact folder / naming for recipes vs `PV*` design-system types
- Whether `Components/Research/` is exclusively recipes
- Inventory pass: classify every existing `PV*` and feature-private sheet
- Fold agreed bits into `macos/App/DesignSystem/README.md` once the cleanup ships

## Agent guidance (started)

- Rule: [`.cursor/rules/design-system-layers.mdc`](../../.cursor/rules/design-system-layers.mdc) (applies under `macos/App/**/*.swift`)
- Skill: [`.cursor/skills/add-ui-component/SKILL.md`](../../.cursor/skills/add-ui-component/SKILL.md) — classify before adding
- Skill: [`.cursor/skills/evaluate-ui-component/SKILL.md`](../../.cursor/skills/evaluate-ui-component/SKILL.md) — point at one type; cousins + compose-down + incremental PR plan
- App-health dimension 12 uses the same vocabulary
## Related docs

- Brad Frost: [Design system components, recipes, and snowflakes](https://bradfrost.com/blog/post/design-system-components-recipes-and-snowflakes/)
- [`macos/App/DesignSystem/README.md`](../../macos/App/DesignSystem/README.md)
- [`macos-client-patterns.md`](../macos-client-patterns.md)
- Spike 7: [`deployment-plan/spike-7/`](../deployment-plan/spike-7/) (paused relative to this cleanup — not a commitment either way until scheduled)
