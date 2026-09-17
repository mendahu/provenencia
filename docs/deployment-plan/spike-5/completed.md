# Spike 5 — Completed steps

Finished Spike 5 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S5-NN`, `S5-DN`). Do not renumber when moving steps here.

## Plan revisions

Changes to the plan itself, as opposed to landed work.

| After | Change |
| --- | --- |
| S5-01 | **The graph became the only surface for Subjects, Citations, and Observations** (design note §1.3). subject list and Source-page entry cut; closing step renumbered S5-10 → S5-09. Data path proven by Go tests / FakeStore; canvas a11y moved to Spike 6 slice 2. |
| S5-01 | **Unified Sources product layer** (design note §1.4): no Interpretation sidebar section; dual action on Sources list → graph; **Subject types / Subject fields** naming; S5-D1 superseded by S5-D3; S5-D2 rewritten. |
| S5-01 | **Nested config nav** + product rename to **Evidence graph** (§1.4, §12): four vocabulary destinations are children of Sources; UI chrome uses Evidence graph rather than Interpretation graph. |
| S5-01 | **Schema rename Node → Subject** before S5-02: `subjects`, `subject_types`, `subject_positions`, `subject_type_fields`. Product and engine share one word (design note decision 19). |

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S5-01](#s5-01--pr-node-type-prefix-validation) | PR | Subject type prefix validation + reserved-prefix guard in `core/ref` |
| [S5-D1](#s5-d1--design-interpretation-nav-entry) | Design | Interpretation sidebar destination — label, icon, placement |

---

## Steps

### S5-01 — PR: Subject type prefix validation

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — |
| **Deliverables** | Done. [`core/ref/ref.go`](../../../core/ref/ref.go): `ValidatePrefix(prefix)` normalizing a Subject type prefix and rejecting the reserved catalog prefixes, plus `ErrReservedPrefix` and the `reservedPrefixes` set. New wire code `ref.reserved_prefix` in [`core/apperr/apperr.go`](../../../core/apperr/apperr.go). No new mint or validate path: candidate subjects are ordinary refs. |
| **Tests** | Done. [`core/ref/ref_test.go`](../../../core/ref/ref_test.go): `TestValidatePrefix` covering all five reserved prefixes, lowercase normalization, `SRN` allowed, and candidate prefixes (`CPR`, `CSR`, `cev`) validating identically to canonical ones. Rows added to `TestValid` asserting `CPR-7KD45` is an ordinary valid ref and that an extra segment is not. |
| **Dogfood** | App unchanged. No schema, no FFI, no Swift — nothing is user-visible. |
| **Out** | `subject_types` / `subjects` tables and their Go packages (S5-02…S5-04). Cross-column prefix uniqueness and ref uniqueness retry, both of which belong to the write paths in S5-02 / S5-04. **L10n mapping for `ref.reserved_prefix`**, deferred because the code is unreachable from Swift until researchers can define Subject types in the vocabulary browser (Spike 7); map it then. |

**Landed:** the reserved-prefix guard the Subject type vocabulary needs — and, more usefully, a much smaller step than planned.

This PR was originally built around an infix candidate marker (`PER-C-7KD45`) so one `ref_prefix` could serve both layers. That required `MintCandidate`, a second validator pair, `ValidAny` / `ValidPartial`, and a fix to [`core/search/refpath.go`](../../../core/search/refpath.go), whose partial-ref regex had no room for a second dash. All of it was reverted in favour of giving candidates **their own prefix** (`CPR-7KD45`, seeded in [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1). One ref format survives, `ref.Mint` already produces Subject refs, and every existing consumer of `Valid` / `Validate` handles them untouched. The search package ends up with a zero-line diff.

The cost moved into the vocabulary: `subject_types` now needs `candidate_ref_prefix` alongside `ref_prefix`, because `canonical_entities` and `subjects` share that table. That column lands with the table in S5-02.

Two rules `ValidatePrefix` deliberately does **not** enforce, both left to the Subject type write path: cross-column prefix uniqueness (the per-column SQL `UNIQUE` catches only half of a single shared namespace), and the leading `C` on candidate prefixes, which is convention and carries no meaning to the code.

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
| **Dogfood** | Design only — nothing in the app. **Superseded before implementation:** product dropped the Interpretation sidebar item in favour of a unified Sources family (S5-D3). Do not implement this board. |
| **Out** | — |

**Landed (design only):** nav chrome for a top-level Interpretation item. **Do not ship.** See plan revisions and [`design/S5-D3-sources-section-nav.md`](design/S5-D3-sources-section-nav.md).
