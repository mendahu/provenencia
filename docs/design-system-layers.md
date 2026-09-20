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

## Folder layout (flat)

Layer **is** the folder. Do **not** nest UI-category subfolders (`Core`, `Forms`, `Feedback`, `Navigation`, `Data`, `Research`, etc.). Discoverability is by layer and file name, not by “this is a form control.”

### macOS (`macos/App/DesignSystem/`)

```text
DesignSystem/
  Tokens/           # colors, type, space, radii, elevation, motion
  Components/       # flat — design-system PV* only
  Recipes/          # flat — Provenencia-specific, ≥2 call sites
  Snowflakes/       # flat — named one-off kit pieces (rare)
Features/<Feature>/ # screen views/models; private snowflake helpers OK here too
```

| Folder | Layer | Contents |
|---|---|---|
| `DesignSystem/Components/` | Design system | Content-agnostic `PV*` (Button, Badge, Dialog, Input, …) — **flat** |
| `DesignSystem/Recipes/` | Recipes | Product maps onto components (evidence icons, subject icons, omnibar hit row, …) — **flat** |
| `DesignSystem/Snowflakes/` | Snowflakes | Named one-screen controls that still live in the DesignSystem tree — **flat** |
| `Features/<Feature>/` | Snowflakes (typical) | Feature views + `private` helpers; prefer this for screen glue |

Prefer **`Features/<Feature>/`** for snowflakes that are glued to one screen. Use **`DesignSystem/Snowflakes/`** only when a named type should live next to the kit (e.g. shared preview host) but is not a recipe or component.

**Incorrect:** `Components/Forms/PVInput.swift`, `Components/Research/PVEvidenceIcon.swift`.  
**Correct:** `Components/PVInput.swift`, `Recipes/PVEvidenceIcon.swift`.

### Claude Design / web kit (parity)

Same three flat roots — no `components/forms/` style nesting:

```text
tokens/
components/    # flat
recipes/       # flat
snowflakes/    # flat
boards/        # screen mockups that compose the above
```

### Transition

Until the move PR lands, some files may still live under legacy `Components/Core|Forms|Feedback|Navigation|Data|Research/`. **Do not add new files to those category folders.** New work targets the flat layout above; the cleanup PR flattens and reclassifies what remains.

## 1. Components (design system)

Shared, **content-agnostic** controls in `DesignSystem/Components/`. They would work if dropped into another app with the same tokens.

- **What they carry:** Provenencia theming (`PV*` tokens) and generic interaction chrome (slots for title, body, actions).
- **What they do not carry:** Source / Subject / catalog vocabulary, `GenealogyStore` types, or feature-specific meaning.
- **Inside the kit**, still distinguish *primitive* vs *composite* when useful — same folder, different size (not subfolders):
  - **Primitive:** button, badge, input, field, sheet **panel** (title / subtitle / body / optional footer).
  - **Composite:** confirm dialog, form dialog — prescribed use of the panel (button roles, focus defaults, presentation API) that remain domain-agnostic.

Composites are still design-system components in Frost’s sense: maximal reuse, no product entity types. Prefer flexible slots over forking a whole new root control for a slight variant.

**Examples:** `PVButton`, `PVBadge`, `PVInput`, `PVField`, `PVConfirm` / `PVDialog` (target: both on one panel primitive).

## 2. Recipes

Product-specific compositions used **consistently across more than one place**. Valuable and reusable, but not agnostic enough for the portable kit.

- Map Provenencia meaning → component props (icons, colors, grade, type key).
- Thin wrappers preferred — do not reimplement badge / chip / row chrome.
- Live flat under `DesignSystem/Recipes/`.

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
5. **Flat folders only.** Layer = `Components/` | `Recipes/` | `Snowflakes/` (plus feature-local snowflakes under `Features/`). No UI-category nesting.
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
