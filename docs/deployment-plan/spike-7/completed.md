# Spike 7 — Completed steps

Finished Spike 7 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S7-NN`, `S7-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S7-01](#s7-01--pr-properties--bindings--subject-registry) | PR | `properties` + `subject_type_fields` + `subjectvocab` registry / Install / FFI |
| [S7-01b](#s7-01b--pr-property-terms) | PR | `property_terms` + `value_type=term` + kind/edge seed + FFI term CRUD |

## Steps

### S7-01 — PR: Properties + bindings + subject registry

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | Spike 5 `subject_types`; Spike 7 plan |
| **Deliverables** | Done. Migration [`000023.sql`](../../../core/database/migrations/000023.sql) adds `properties` (five `value_type`s at land — **`term` follows in S7-01b**) and `subject_type_fields` (`sort_order`). Package [`core/database/properties/`](../../../core/database/properties/) with audited CRUD. Package [`core/database/subjectvocab/`](../../../core/database/subjectvocab/) owns create-time Install (types + Properties + bindings), capabilities, presentation tokens, locked bindings, and connect matrix. Onboarding `createCatalog` calls `subjectvocab.Install` (replaces `subjecttypes.Install`). FFI: Property CRUD, subject-type field assign/remove (locked refuse), placeable / presentation / connect lookup. Swift: GenealogyStore + FakeStore + GoStore stubs. Docs: interpretation-layer DDL + `add-seeded-vocabulary` skill. Kind/edge Properties (`event_type`, `role`, `relationship_type`) are **omitted** from Install until **S7-01b** introduces them as `term`. |
| **Tests** | Done. Go: `properties`, `subjectvocab`, `subjecttypes`, onboarding; FFI `subject_defs_test` via `runRPC`. |
| **Dogfood** | Create a new project → seeded Properties (§3.2) and bindings (§3.3) present; placeable list is person/event/place; removing a locked participation edge binding fails. |
| **Out** | Property terms (**S7-01b**); Subject fields UI (S7-05); Observations (S7-03); Evidence graph Swift migration off hard-coded kinds (S7-09); NameValue tables (S7-02); Subject type researcher CRUD. |

**Landed:** central Interpretation subject registry as the SoT for seed + app behavior; structural catalog rows only in SQLite.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/... ./core/database/subjectvocab/... ./core/onboarding/... ./api/ffi/...
python3 scripts/check-localizable-xcstrings.py
```

### S7-01b — PR: Property terms

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S7-01 |
| **Deliverables** | Done. Migration [`000024.sql`](../../../core/database/migrations/000024.sql) rebuilds `properties` CHECK to include `term` and adds `property_terms`. Package [`core/database/propertyterms/`](../../../core/database/propertyterms/) with audited CRUD (product/plugin terms locked; `ErrInUse` stub until S7-03). `properties.Create` refuses `value_type = term` for researchers; Upsert/seed allows it. `subjectvocab` Install seeds kind/edge Properties (`event_type`, `role`, `relationship_type`), bindings, and `seedTerms` (§3.4–3.6). Term capabilities (birthday / tree-edge) deferred. FFI: List/Create/Update/Delete PropertyTerm. Swift: GenealogyStore + FakeStore + GoStore stubs + L10n for `propertyterms.*`. Docs: seeded-vocabulary §3.6 starter set; interpretation binding matrix. |
| **Tests** | Done. Go: `properties`, `propertyterms`, `subjectvocab`, onboarding; FFI `property_terms_test` via `runRPC`. |
| **Dogfood** | Create a new project → 15 seeded Properties including term rows (`event_type`, `role`, `relationship_type`, `sex_at_birth`); ListPropertyTerms on `event_type` includes `birth`; Create Property with `term` fails; user term create/update/delete works; product term update fails. |
| **Out** | Observation `value_term_id` (S7-03); composer term picker UI (S7-D4 / S7-08); Subject fields UI (S7-05); Event types / Roles admin destinations. |

**Landed:** kind/edge vocabulary as Property terms; registry-only `term` Properties.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/... ./core/database/propertyterms/... ./core/database/subjectvocab/... ./core/onboarding/... ./api/ffi/...
python3 scripts/check-localizable-xcstrings.py
```
