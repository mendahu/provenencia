# Spike 5 — Completed steps

Finished Spike 5 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S5-NN`, `S5-DN`). Do not renumber when moving steps here.

## Plan revisions

Changes to the plan itself, as opposed to landed work.

| After | Change |
| --- | --- |
| S5-01 | **The graph became the only surface for Nodes, Citations, and Observations** (design note §1.3). This spike therefore ships no Node UI: the *Source nodes list* step and the *Source-page entry* step were both cut, along with their design briefs, and the closing docs step moved from S5-10 to S5-09 — safe to renumber because nothing past S5-01 had landed. Ten steps became nine and four design briefs became two. A Source may open a stub / "coming soon" destination; the Node data path is proven by Go tests and `FakeStore` parity. Accessibility moved with it: the canvas's own accessibility representation is slice-2 work in Spike 6, not polish, because there is no list view to fall back to. |

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S5-01](#s5-01--pr-node-type-prefix-validation) | PR | Node Type prefix validation + reserved-prefix guard in `core/ref` |
| [S5-D1](#s5-d1--design-interpretation-nav-entry) | Design | Interpretation sidebar destination — label, icon, placement |

---

## Steps

### S5-01 — PR: Node Type prefix validation

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — |
| **Deliverables** | Done. [`core/ref/ref.go`](../../../core/ref/ref.go): `ValidatePrefix(prefix)` normalizing a Node Type prefix and rejecting the reserved catalog prefixes, plus `ErrReservedPrefix` and the `reservedPrefixes` set. New wire code `ref.reserved_prefix` in [`core/apperr/apperr.go`](../../../core/apperr/apperr.go). No new mint or validate path: candidate Nodes are ordinary refs. |
| **Tests** | Done. [`core/ref/ref_test.go`](../../../core/ref/ref_test.go): `TestValidatePrefix` covering all five reserved prefixes, lowercase normalization, `SRN` allowed, and candidate prefixes (`CPR`, `CSR`, `cev`) validating identically to canonical ones. Rows added to `TestValid` asserting `CPR-7KD45` is an ordinary valid ref and that an extra segment is not. |
| **Dogfood** | App unchanged. No schema, no FFI, no Swift — nothing is user-visible. |
| **Out** | `node_types` / `nodes` tables and their Go packages (S5-02…S5-04). Cross-column prefix uniqueness and ref uniqueness retry, both of which belong to the write paths in S5-02 / S5-04. **L10n mapping for `ref.reserved_prefix`**, deferred because the code is unreachable from Swift until researchers can define Node Types in the vocabulary browser (Spike 7); map it then. |

**Landed:** the reserved-prefix guard the Node Type vocabulary needs — and, more usefully, a much smaller step than planned.

This PR was originally built around an infix candidate marker (`PER-C-7KD45`) so one `ref_prefix` could serve both layers. That required `MintCandidate`, a second validator pair, `ValidAny` / `ValidPartial`, and a fix to [`core/search/refpath.go`](../../../core/search/refpath.go), whose partial-ref regex had no room for a second dash. All of it was reverted in favour of giving candidates **their own prefix** (`CPR-7KD45`, seeded in [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1). One ref format survives, `ref.Mint` already produces Node refs, and every existing consumer of `Valid` / `Validate` handles them untouched. The search package ends up with a zero-line diff.

The cost moved into the vocabulary: `node_types` now needs `candidate_ref_prefix` alongside `ref_prefix`, because `canonical_entities` and `nodes` share that table. That column lands with the table in S5-02.

Two rules `ValidatePrefix` deliberately does **not** enforce, both left to the Node Type write path: cross-column prefix uniqueness (the per-column SQL `UNIQUE` catches only half of a single shared namespace), and the leading `C` on candidate prefixes, which is convention and carries no meaning to the code.

Docs updated with the step: [`catalog-refs.md`](../../catalog-refs.md) §2 and §4, [`data-model-source-interpretation-conclusion.md`](../../data-model-source-interpretation-conclusion.md) §2, [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §4.1–4.2, [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §6, [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1, the [`catalog-refs`](../../../.cursor/rules/catalog-refs.mdc) rule, and [`add-catalog-ref`](../../../.cursor/skills/add-catalog-ref/SKILL.md) (whose stale test command was also corrected to include `-tags fts5`).

**Verify:**

```bash
go test ./core/ref/
CGO_ENABLED=1 go test -tags fts5 ./...
```

### S5-D1 — Design: Interpretation nav entry

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | Shipped workspace chrome (S2-01) |
| **Deliverables** | Done. Board for the Interpretation sidebar destination: label, icon, placement among existing sections, selected/idle states, empty badge slot. Brief archived: [`design/archive/S5-D1-interpretation-nav-entry.md`](design/archive/S5-D1-interpretation-nav-entry.md). |
| **Dogfood** | Design only — nothing in the app yet. Implements in S5-07. |
| **Out** | The Sources list behind the destination (S5-D2); the canvas and Node UI (Spike 6). |

**Landed:** the nav chrome for the layer. Small board by design; S5-07 commits to the label and `PVSymbol` from it.
