---
name: evaluate-ui-component
description: >-
  Evaluates one Provenencia macOS UI type against design-system / recipe /
  snowflake layering, finds near-duplicate implementations, checks lower-layer
  PV* composition opportunities, and returns a small refactor plan. Use when
  the user points at a component (PV*, sheet, dialog, badge, chip, panel, row,
  menu) and asks to evaluate, audit, harden, refactor for composability, drill
  down, compose upward, find duplicates, or chip away at design-system debt—
  not for whole-app health (review-app-health) or for adding a brand-new control
  from scratch (add-ui-component).
---

# Evaluate a UI component (incremental hardening)

Point-at-one-component review. Goal: **classify → find cousins → compose down →
plan small PRs**. Do **not** implement unless the user asks. Do **not** expand
into a whole-app cleanup (use [`review-app-health`](../review-app-health/SKILL.md)
for that). Prefer chipping away: one focused plan the user can approve slice by
slice.

Layers: [`docs/ideas/design-system-hardening.md`](../../../docs/ideas/design-system-hardening.md).
Adding new chrome: [`add-ui-component`](../add-ui-component/SKILL.md).
Kit notes: [`macos/App/DesignSystem/README.md`](../../../macos/App/DesignSystem/README.md).

## When invoked

1. Identify the **subject** (type name + file path). If the user named a feature
   screen, pick the concrete view/helper they mean or ask once.
2. Run the workflow below (read code; Grep cousins; inventory lower `PV*`).
3. Deliver the **report template** in chat. Stop at the plan unless asked to
   implement a numbered item.

## Progress checklist

```
Evaluate UI component:
- [ ] 1. Subject + call sites
- [ ] 2. Classify Frost layer (actual vs ideal)
- [ ] 3. Near-duplicates / cousins
- [ ] 4. Lower-layer compose opportunities
- [ ] 5. Wrong-layer / anti-pattern flags
- [ ] 6. Refactor plan (small PRs, ordered)
```

## Workflow

### 1. Subject + call sites

- Read the type and its immediate private helpers.
- Grep production call sites under `macos/App/Features/` and `Platform/`
  (previews/tests alone don’t count as product use).
- Note presentation: `.sheet` / `.alert` / overlay / inline.

### 2. Classify (actual vs ideal)

| Question | Points to |
| --- | --- |
| Content-agnostic (tokens + slots only)? | Design system |
| Provenencia meaning, ≥2 real call sites? | Recipe |
| One feature / one screen? | Snowflake |
| Slight variant of an existing `PV*`? | Should compose that `PV*` — not a new root |

Inside DesignSystem: **primitive** vs **composite** (same layer, different size).

State clearly: **today’s layer** vs **ideal layer** (may already be correct).

### 3. Near-duplicates / cousins

Search for siblings that share chrome but diverge in payload or open gesture:

- Same visual band: sunken footer, `PVDivider`, title `h3` + subtitle, card/shadow,
  dismiss monitor, hover row, chip/badge shell.
- Name patterns: `*Sheet`, `*Dialog`, `*Panel`, `*Menu`, `*Chip`, `*Badge`, `*Row`.
- Grep distinctive tokens (`surfaceSunken`, `.pvDialog`, `.pvConfirm`, custom
  scrim, `xmark` dismiss in a sheet header).

List cousins with path + one-line how they differ. Flag **copy-paste drift**.

### 4. Lower-layer compose opportunities (“compose upward” from primitives)

Walk **down** the stack the subject should sit on:

1. Tokens / existing primitives (`PVButton`, `PVBadge`, `PVField`, `PVDivider`, …)
2. Composites (`.pvDialog`, `.pvConfirm`, `.pvConfirmSheet`, `PVContextMenu`, …)
3. Existing recipes in `Components/Research/` (or similar)

Ask: can this subject’s chrome be **deleted** in favor of a lower control +
slots/props? Prefer extending a lower type’s slot over keeping a parallel stack.

macOS dialog reminder: no scrim/dim; don’t redraw window radius/shadow on panel
content; prefer DS confirm/form dialog over bespoke sheet chrome.

### 5. Wrong-layer / anti-pattern flags

- Catalog / `GenealogyStore` / hard-coded feature copy inside `DesignSystem/`
- `PV*` with a single call site (premature promotion)
- Recipe with one caller (should still be a snowflake)
- Snowflake that owns shared dialog/menu chrome (should fill a DS slot)
- Orphan `PV*` (zero production call sites)

### 6. Plan (small PRs only)

Produce an **ordered** list of the smallest valuable changes. Each item:

- **Problem** — evidence (`path`, lines)
- **Move** — compose into X / extract recipe Y / demote to snowflake / delete orphan
- **Scope** — files touched; what stays out of scope
- **Risk** — visual/a11y/regression notes
- **Verify** — preview, existing tests, or manual path

Do **not** propose one mega-PR that rewires every cousin. Default: fix the
**subject** first; optional follow-ups for cousins the user can schedule later.

Ask which numbered items to implement — do not start coding unprompted.

## Report template

```markdown
## Subject
- Type / file:
- Call sites (production):
- Today’s layer: design system (primitive|composite) | recipe | snowflake
- Ideal layer:

## Cousins
| Path | Similarity | Difference |
| --- | --- | --- |
| … | … | … |

## Compose-down opportunities
- Could use: `PV…` / `.pv…` because …
- Cannot yet: … (missing slot / width / footer API) → optional prerequisite PR

## Flags
- …

## Plan (smallest PRs first)
1. …
2. …

Which items should I implement?
```

## Anti-scope

- Whole-app sampling → `review-app-health`
- Brand-new control with no subject to evaluate → `add-ui-component`
- Rewriting DesignSystem README / tokens unless the plan explicitly needs it
- Implementing “while we’re here” cousins without user approval

## Example triggers

- “Evaluate `SourceTypeIconPickerSheet` against our layering rules”
- “Look at `PVDialog` / `PVConfirm` for compose opportunities”
- “Audit this badge — can we drill into a primitive and make a recipe?”
