# Spike 7 — Completed steps

Finished Spike 7 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S7-NN`, `S7-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S7-01](#s7-01--pr-properties--bindings--subject-registry) | PR | `properties` + `subject_type_fields` + `subjectvocab` registry / Install / FFI |
| [S7-01b](#s7-01b--pr-property-terms) | PR | `property_terms` + `value_type=term` + kind/edge seed + FFI term CRUD |
| [S7-D2](#s7-d2--design-subject-fields) | Design | Type strip over property table; gates S7-05 |
| [S7-05](#s7-05--pr-subject-fields-ui) | PR | Subject fields destination: strip + table + inspector |
| [S7-02](#s7-02--pr-namevalue-schema--go) | PR | `name_values` / `name_value_parts` + `namevalues` Insert/Lookup |

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

### S7-D2 — Design: Subject fields

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | S7-01 / S7-01b schema; Subject fields stub nav (S5-D3) |
| **Deliverables** | Done. Board **S7-D2 Subject Fields**: type strip (All + seven fixed types) over a dense property table with search/filters and a right inspector (description, bindings, delete). Create Property offers five researcher value_types only (**not** `term`). Locked registry bindings as lock boxes (not refused checkboxes). Seeded `term` Properties visible as product vocabulary. Briefs archived: [`design/archive/S7-D2-subject-fields.md`](design/archive/S7-D2-subject-fields.md), [`design/archive/S7-D2-subject-fields-addendum-property-terms.md`](design/archive/S7-D2-subject-fields-addendum-property-terms.md). |
| **Dogfood** | Design only — implements in **S7-05**. |
| **Out** | Composer / term picker (S7-D4); NameValue (S7-D5); Event types / Roles admin; Source-fields chrome. |

**Landed:** IA for Subject fields config; feeds **S7-05**.

### S7-05 — PR: Subject fields UI

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S7-01, S7-01b, **S7-D2** |
| **Deliverables** | Done. Replaces Subject fields stub with S7-D2 IA: type strip (All + seven types) over searchable property table + right inspector (description, bindings, delete). Create Property sheet offers five researcher value_types only (**not** `term`). Locked registry bindings as lock boxes with reason callout. Seeded `term` Properties visible with inspector note. Workspace place: `subjectFieldsWorkspace` CatalogQueryKey, PlaceRegistry warm, CatalogCounts badge, DestinationHost → `SubjectFieldsView`. L10n.SubjectFields + xcstrings. |
| **Tests** | Done. `SubjectFieldsModelTests` (FakeStore): type filter, create without `term`, locked binding callout, assign/remove unlocked, refuse seeded/in-use delete. `PlaceRegistryTests` expect workspace query key. |
| **Dogfood** | Open Subject fields → strip filters table → create a `name` Property → bind to Person → locked event date binding cannot be removed → seeded `event_type` shows as `term` without create-`term`. |
| **Out** | Composer / term picker (S7-D4 / S7-08); NameValue (S7-D5 / S7-02b); Observation editors; Subject types CRUD; Source-fields layout reuse; Event types / Roles admin; Property-term CRUD UI. |

**Landed:** Subject fields config destination matching S7-D2 board.

**Verify:**

```bash
python3 scripts/check-localizable-xcstrings.py
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/SubjectFieldsModelTests -only-testing:ProvenenciaTests/PlaceRegistryTests CODE_SIGNING_ALLOWED=NO
```

### S7-02 — PR: NameValue schema + Go

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — (parallel after S7-01; not gated on S7-D5) |
| **Deliverables** | Done. Migration [`000025.sql`](../../../core/database/migrations/000025.sql) adds `name_values` + `name_value_parts` per structured-name-model §§2–3. Package [`core/database/namevalues/`](../../../core/database/namevalues/) with DateValue-shaped `Insert` / `Lookup` (write-once value object; transactional parent + parts). `apperr.CodeNameValuesInvalid`. Starter part-type consts (open vocabulary, not enforced). Skill [`add-name-value`](../../../.cursor/skills/add-name-value/SKILL.md) + rule `name-values.mdc`. |
| **Tests** | Done. Go: `namevalues` table-driven Insert/Lookup (form-only, parts, rejects, schema/`user_version`). |
| **Dogfood** | Schema/Go only — no app UI yet. Consumers: Observations `value_name_id` (**S7-03**); Swift editor (**S7-02b**). |
| **Out** | Swift NameValue editor (**S7-02b**); Observation FK / composer (**S7-03** / **S7-08**); `name_format_profiles` / project defaults; FFI. |

**Landed:** shared NameValue persistence so later Observations can reference structured names.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/namevalues/... ./core/database/...
```
