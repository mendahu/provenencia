# Design system layers (macOS)

Authoritative composition model for Provenencia UI. Agents: [`.cursor/rules/design-system-layers.mdc`](../.cursor/rules/design-system-layers.mdc), [`.cursor/skills/add-ui-component`](../.cursor/skills/add-ui-component/SKILL.md), [`.cursor/skills/evaluate-ui-component`](../.cursor/skills/evaluate-ui-component/SKILL.md). Kit conventions (tokens, `PV*` file shape): [`macos/App/DesignSystem/README.md`](../macos/App/DesignSystem/README.md). Client folder layout: [`macos-client-patterns.md`](macos-client-patterns.md).

We follow Brad Frost’s **[components / recipes / snowflakes](https://bradfrost.com/blog/post/design-system-components-recipes-and-snowflakes/)** split — three layers of increasing product coupling, not a stack you must climb on every screen.

```
┌─────────────────────────────────────────────────────────┐
│  3. Snowflakes — one screen / feature-private           │  more bespoke
├─────────────────────────────────────────────────────────┤
│  2. Recipes — product-specific, reused across views     │
├─────────────────────────────────────────────────────────┤
│  1. Components — content-agnostic PV* kit               │  more generic
└─────────────────────────────────────────────────────────┘
```

Every building block belongs in **exactly one** layer. Higher layers may compose lower ones; they must not reinvent them. Lower layers must not import higher ones (no catalog / genealogy types inside `Components/`).

**You stop at the highest layer you need.** Most call sites never touch all three:

- A view can drop in a **component** directly (`.pvConfirmSheet`, `PVButton`) with copy and slots filled at the call site.
- A **recipe** exists only when the same Provenencia meaning is reused and deserves a named mapping.
- A **snowflake** exists only for glue unique to that screen.

Do **not** invent a recipe or snowflake shell “for completeness.” Extra wrappers without a second call site (or real product meaning) are the same drift this model prevents.

When a new control is needed, ask which layer it is **before** writing a new file. Prefer composing downward over forking sideways.

## Folder layout

Three **layer roots** — not UI-category buckets (`Core`, `Forms`, `Feedback`, `Research`, …). Under each root, **one folder per component** so Swift, docs, helpers, and models colocate.

### macOS (`macos/App/DesignSystem/`)

```text
DesignSystem/
  Tokens/                         # shared tokens (not a component)
  Components/                     # design-system layer
    Button/
      PVButton.swift
      # optional: README.md, private helpers, previews-only fixtures
    Input/
      PVInput.swift
    Confirm/
      PVConfirm.swift
    …
  Recipes/                        # product recipes
    EvidenceIcon/
      PVEvidenceIcon.swift
      # optional: EVIDENCE-ICONS.md, asset notes
    SubjectIcon/
      PVSubjectIcon.swift
    …
  Snowflakes/                     # named kit-side one-offs (rare)
    <Name>/
      …
Features/<Feature>/               # screens; private snowflake helpers OK here too
```

| Root | Layer | Child folders |
|---|---|---|
| `DesignSystem/Components/<Name>/` | Design system | One folder per content-agnostic control |
| `DesignSystem/Recipes/<Name>/` | Recipes | One folder per product recipe |
| `DesignSystem/Snowflakes/<Name>/` | Snowflakes | One folder per named kit-side one-off |
| `Features/<Feature>/` | Snowflakes (typical) | Feature views + `private` helpers |

**Folder naming:** PascalCase **without** the `PV` prefix (`Button`, `EvidenceIcon`). The primary Swift type may still be `PVButton` / `PVEvidenceIcon` inside that folder.

**Colocate** in the component folder: the main view, private backing types, small helpers, and component-specific docs. Do **not** dump unrelated controls into the same folder. Shared cross-cutting tokens stay in `Tokens/`.

Prefer **`Features/<Feature>/`** for snowflakes glued to one screen. Use **`DesignSystem/Snowflakes/<Name>/`** only when a named type should live next to the kit but is not a recipe or component.

**Incorrect:** `Components/Forms/PVInput.swift`, `Components/PVInput.swift` (loose file at layer root), `Components/Research/PVEvidenceIcon.swift`.  
**Correct:** `Components/Input/PVInput.swift`, `Recipes/EvidenceIcon/PVEvidenceIcon.swift`.

### Claude Design / web kit (parity)

```text
tokens/
components/
  Button/
  Input/
  …
recipes/
  EvidenceIcon/
  …
snowflakes/
  <Name>/
boards/        # screen mockups that compose the above
```

No `components/forms/` category nesting. Each control is `components/<Name>/`.

### Transition

Until the move PR lands, some files may still live under legacy `Components/Core|Forms|Feedback|Navigation|Data|Research/`. **Do not add new files to those category folders or as loose files at a layer root.** New work uses `Components|<Name>/`, `Recipes/<Name>/`, or `Snowflakes/<Name>/` (or feature-local snowflakes).

## 1. Components (design system)

Shared, **content-agnostic** controls in `DesignSystem/Components/`. They would work if dropped into another app with the same tokens.

- **What they carry:** Provenencia theming (`PV*` tokens) and generic interaction chrome (slots for title, body, actions).
- **What they do not carry:** Source / Subject / catalog vocabulary, `GenealogyStore` types, or feature-specific meaning.
- **Inside the kit**, still distinguish *primitive* vs *composite* when useful — same layer, different size (each still gets its own `<Name>/` folder):
  - **Primitive:** button, badge, input, field, sheet **panel** (title / subtitle / body / optional footer).
  - **Composite:** confirm dialog, form dialog — prescribed use of the panel (button roles, focus defaults, presentation API) that remain domain-agnostic.

Composites are still design-system components in Frost’s sense: maximal reuse, no product entity types. Prefer flexible slots over forking a whole new root control for a slight variant.

**Examples:** `PVButton`, `PVBadge`, `PVInput`, `PVField`, `PVConfirm` / `PVDialog` (target: both on one panel primitive).

## 2. Recipes

Product-specific compositions used **consistently across more than one place**. Valuable and reusable, but not agnostic enough for the portable kit.

- Map Provenencia meaning → component props (icons, colors, grade, type key).
- Thin wrappers preferred — do not reimplement badge / chip / row chrome.
- Live under `DesignSystem/Recipes/<Name>/`.

**Examples:** `PVEvidenceIcon`, `PVSubjectIcon`, evidence-grade / source-type badge families; a shared “delete vocabulary row” confirm shell *if* several screens share the same wiring.

## 3. Snowflakes

Bespoke UI for **one** screen (or private helpers inside one feature). Written once, called from that feature only.

- Prefer `Features/<Feature>/`, `private` / file-scoped; or `DesignSystem/Snowflakes/` when a named kit-side one-off is warranted.
- **OK:** layout glue, one-off section chrome, a sheet body that will never be reused.
- **Not OK:** quietly becoming the de-facto way we do dialogs while living as a one-off — promote into `Components/` or `Recipes/` instead.
- **Second call site promotes** a snowflake → recipe (same product kind, now shared) or into `Components/` if it was content-agnostic all along.

**Example:** `SourceTypeIconPickerSheet`’s icon grid (and any chrome that isn’t absorbed by a shared panel).

## Layering rules

1. **Stop at the layer you need.** Components are valid end products at a call site. Recipes and snowflakes are optional specialization.
2. **Compose down, don’t clone sideways.** A slightly different confirm is still DS confirm + props, not a new root dialog.
3. **Components stay portable.** No `GenealogyStore`, catalog models, or feature `L10n` namespaces inside `Components/` (generic chrome labels owned by the control are fine).
4. **Reuse is what separates recipes from snowflakes** — same product-coupled kind; promote on the second real call site instead of pre-building a recipe.
5. **Layer roots + per-component folders.** Roots are only `Components/` | `Recipes/` | `Snowflakes/` (plus feature-local snowflakes under `Features/`). No UI-category nesting. Each control is `…/<Name>/` with colocated Swift/docs/helpers — not a loose `.swift` at the layer root.
6. **No orphans.** Recipes and named snowflakes get an intentional home — don’t leave domain chrome under `Components/`.

## Dialogs as a worked example

Not every delete flow needs all three layers:

| Layer | Role | Required? |
|---|---|---|
| Component | Sheet **panel** primitive; **confirm** and **form dialog** composites on top (cancel-first / primary confirm, `isRunning`, no scrim — window owns chrome on macOS). | Usually enough |
| Recipe | Only if several screens share the same catalog-shaped confirm wiring. | Optional |
| Snowflake | Icon picker grid, graph create form fields, etc., passed into a component body slot — or omitted when slots + call-site copy are enough. | Optional |

Typical Source-fields delete: **component confirm at the view** (`.pvConfirmSheet` + `PVConfirmCopy` + optional `PVConfirmKeyChip`). That is fine. No recipe required unless the same wiring repeats.

`PVDialog` and `PVConfirm` sheet content still duplicate panel chrome in places; feature sheets that reimplement header/footer by hand should compose the kit instead. Chip away with [`evaluate-ui-component`](../.cursor/skills/evaluate-ui-component/SKILL.md) rather than one mega-refactor.

Sheet presentation details (no scrim, sunken footer as content): DesignSystem README “Confirmations are system chrome” and `PVConfirm.swift` / `PVDialog.swift`.
