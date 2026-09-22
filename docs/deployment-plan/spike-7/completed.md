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
| [S7-03](#s7-03--pr-citations--observations--locator) | PR | Citations + Observations + locator validate + FFI + graph `isCited` |
| [S7-D6](#s7-d6--design-curated-marks) | Design | Marks pack brief; gates S7-12 |
| [S7-12](#s7-12--pr-curated-marks-consolidation) | PR | `Recipes/Marks/` + asset pack; retire EvidenceIcon/SubjectIcon |
| [S7-D8](#s7-d8--design-pvcallout-actions) | Design | Callout actions slot; gates S7-14 |
| [S7-14](#s7-14--pr-pvcallout-actions-slot) | PR | Optional `@ViewBuilder` actions on `PVCallout` |
| [S7-D3](#s7-d3--design-evidence-graph-updates) | Design | Card chrome + No-Artifact; gates S7-09 / S7-10 |
| [S7-09](#s7-09--pr-add-property--composer-navigation) | PR | Evidence graph card updates + composer stub place |

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
| **Deliverables** | Done. Migration [`000025.sql`](../../../core/database/migrations/000025.sql) adds `name_values` + `name_value_parts` per structured-name-model §§2–3. Package [`core/database/namevalues/`](../../../core/database/namevalues/) with DateValue-shaped `Insert` / `Lookup` (write-once value object; transactional parent + parts). `apperr.CodeNameValuesInvalid`. Product part-type **compiled registry** (`PartTypes` / `KnownPartType`); Insert rejects unknown non-empty types; empty type = untyped. Skill [`add-name-value`](../../../.cursor/skills/add-name-value/SKILL.md) + rule `name-values.mdc`. |
| **Tests** | Done. Go: `namevalues` table-driven Insert/Lookup (form-only, parts, rejects unknown type, schema/`user_version`); registry coverage. |
| **Dogfood** | Schema/Go only — no app UI yet. Consumers: Observations `value_name_id` (**S7-03**); Swift editor (**S7-02b**). |
| **Out** | Swift NameValue editor (**S7-02b**); Observation FK / composer (**S7-03** / **S7-08**); `name_format_profiles` / project defaults; FFI; user-minted part-type catalog. |

**Landed:** shared NameValue persistence so later Observations can reference structured names.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/namevalues/... ./core/database/...
```

### S7-03 — PR: Citations + Observations + locator

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S7-01, S7-01b, S7-02 |
| **Deliverables** | Done. Migration [`000026.sql`](../../../core/database/migrations/000026.sql) adds `citations`, `citation_notes`, `observations`, `observation_notes`. Package [`core/locator/`](../../../core/locator/) validates `page` / `region` / `text_quote` (unknown types preserved). Packages [`core/database/citations/`](../../../core/database/citations/) (`CreateWithObservations`) and [`core/database/observations/`](../../../core/database/observations/) (`AddToCitation`, `ListBySource`). Tx-scoped `datevalues.InsertTx` / `namevalues.InsertTx`. Property / property-term `InUse` checks Observations. FFI: create + append + list-by-source. Swift: GenealogyStore / FakeStore / GoStore; `SourceGraphSnapshot` sets `isCited` + per-subject observations; `CatalogMutation.createdCitation` / `addedObservations`. L10n for `locator.invalid` / `citations.invalid` / `observations.invalid`. |
| **Tests** | Done. Go: locator, citations, observations; FFI `citations_test` via `runRPC`. Swift: SourceGraphSnapshot cited flag; CatalogQueryRegistry invalidation. |
| **Dogfood** | Schema/FFI ready — composer submit UI is **S7-08**; card chrome **S7-09**. Create-with-observations + append + list round-trip via FakeStore / Go tests. |
| **Out** | Composer UI (S7-08); Add-property / cited-row chrome (S7-09); NameValue Swift editor (S7-02b); durable connect (S7-10). |

**Landed:** durable Citation + Observation writes (first submit and append) so the Evidence graph can mark cited subjects.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/locator/... ./core/database/citations/... ./core/database/observations/... ./core/database/... ./api/ffi/...
python3 scripts/check-localizable-xcstrings.py
```

### S7-D6 — Design: Curated marks

| | |
| --- | --- |
| **Kind** | Design (Claude Design handoff / Design System pack) |
| **Depends on** | Shipped EvidenceIcon + SubjectIcon; Subject fields type strip (S7-05); S7-01 registry presentation |
| **Deliverables** | Done. Brief + asset pack for unified **Marks** recipe: `file_*` / `type_*` / `subject_*` (seven kinds incl. **source** folio from Subject fields). Tint = template assets + call-site `.foregroundStyle`. UI building-block inventory (New/Extend Marks; Retire EvidenceIcon/SubjectIcon). Brief archived: [`design/archive/S7-D6-curated-marks.md`](design/archive/S7-D6-curated-marks.md). |
| **Dogfood** | Design only — implements in **S7-12**. |
| **Out** | Card / Add-property chrome (**S7-09**); researcher subject-icon picker; new metaphors beyond seven kinds. |

**Landed:** contract for `Recipes/Marks/` consolidation before graph chrome thickens.

### S7-12 — PR: Curated Marks consolidation

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | **S7-D6**; schedule after S7-03, before S7-09 |
| **Deliverables** | Done. `DesignSystem/Recipes/Marks/` (`PVMark` / `PVMarkKey` / `PVMarkFamily` / `PVMarkSize` + `PVFileTypeGlyph`) with `Assets.xcassets/Marks/` (9 file + 22 type + 7 subject template SVGs). Colocated [`MARKS.md`](../../../macos/App/DesignSystem/Recipes/Marks/MARKS.md). All former `PVEvidenceIcon` / `PVSubjectIcon` call sites rewritten to `PVMark` (graph cards/palette/bridges, Subject fields strip, Sources/types/thumbs/omnibar). Registry `IconSymbol` → curated `subject_*` keys. L10n `designSystem.mark.*` (renamed from `evidenceIcon`). Retired `Recipes/EvidenceIcon/` + `Recipes/SubjectIcon/` + `EvidenceIcons` catalog. DesignSystem README + layers docs updated. |
| **Tests** | Done. Go: `subjectvocab` IconSymbol assert; Swift: SourceTypes / PVFileTypeGlyph tests on `PVMarkKey`. |
| **Dogfood** | Sources type icons still tint; Subject fields strip shows all seven kinds incl. source; Evidence graph cards/palette/bridges show subject marks at zoom; no Canvas leftover. |
| **Out** | Card Add-property / cited-row UX (**S7-09**); Subject fields IA; researcher-editable subject icons; merging with SF Symbols `PVIcon`. |

**Landed:** one asset-backed Marks pack for file, Source-type, and subject glyphs — clean cut, no shims.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/... ./api/ffi/...
python3 scripts/check-localizable-xcstrings.py
rg -n 'PVEvidenceIcon|PVSubjectIcon|EvidenceIcons|evidenceIcon|Recipes/EvidenceIcon|Recipes/SubjectIcon' macos
```

### S7-D8 — Design: PVCallout actions

| | |
| --- | --- |
| **Kind** | Design (Claude Design kit handoff) |
| **Depends on** | Shipped `PVCallout`; S7-D3 message-center need |
| **Deliverables** | Done. Board / kit specimen for optional Callout **actions** under the body (one / two buttons; compact). Content-agnostic slot — call-site `PVButton`s. Defers `onDismiss` / `detail` / `plain`. Brief archived: [`design/archive/S7-D8-pvcallout-actions.md`](design/archive/S7-D8-pvcallout-actions.md). |
| **Dogfood** | Design only — implements in **S7-14**. |
| **Out** | Evidence graph No-Artifact gate wiring (**S7-09**); new banner/message-center component. |

**Landed:** contract for `PVCallout` actions before graph message center.

### S7-14 — PR: PVCallout actions slot

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | **S7-D8**; after S7-12, before S7-09 |
| **Deliverables** | Done. `PVCallout` is generic over an optional `@ViewBuilder` actions slot (row under body, `space-5` gap / top padding). `EmptyView` convenience init keeps text-only call sites unchanged. Previews cover zero / one / two actions + compact. DesignSystem README Callout row updated. `onDismiss` / `detail` / `plain` still deferred. |
| **Tests** | Existing call sites type-check; SwiftUI preview specimens. |
| **Dogfood** | Preview shows callout + action; Subject fields / Sources locked notes still look correct. |
| **Out** | Graph No-Artifact gate (**S7-09**); new banner component. |

**Landed:** kit Callout can host a recovery CTA for the Evidence graph message center.

**Verify:**

```bash
# Existing PVCallout(…) call sites compile; preview actions under body.
rg -n 'struct PVCallout' macos/App/DesignSystem/Components/Callout/PVCallout.swift
```

### S7-D3 — Design: Evidence graph updates

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | Spike 6 cards; prefer S7-12 / S7-14 |
| **Deliverables** | Done. Board for Add property, cited rows, refs, No-Artifact message center, connect handoff (S7-10). Brief archived: [`design/archive/S7-D3-evidence-graph-updates.md`](design/archive/S7-D3-evidence-graph-updates.md). |
| **Dogfood** | Design only — implements in **S7-09** / **S7-10**. |
| **Out** | Composer layout (S7-D4); durable edge summaries (S7-10). |

**Landed:** binding inventory for graph chrome before Add property ships.

### S7-09 — PR: Add property + composer navigation

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S7-03, S7-12, S7-14, **S7-D3** |
| **Deliverables** | Done. Primary cards **264pt** with type·ref line, cited Observation rows, quiet **Add property**, Edit pencil. Edit label/description reuses create `PVFormDialog` + `updateSubject` (**shipped delta** vs board “create-only” sheet). No-Artifact graph-wide `PVCallout` (warning + Add an Artifact → Source page) disables palette / Connect / Add property from `canCite`. Composer `WorkspaceLocation` (`sourceSurface: .citationComposer` + `subjectId`) with stub destination. Bridge cards show ref (honesty body until S7-10). Registry presentation tokens drive card/palette ink/tint/chip/line where exposed. L10n + hit-target action zones for AppKit pointer ownership. |
| **Tests** | Navigation / destination host / pointer action hit tests; EvidenceGraphModel edit + No-Artifact; xcstrings check. |
| **Dogfood** | Place subject → Edit → ref visible → Add property → stub composer → Back; open graph with zero Artifacts → callout + disabled tools. |
| **Out** | Real composer form (S7-08); connect disambiguation / edge summaries (S7-10). |

**Landed:** Add property navigates; cards grow with cited rows; Artifact gate is honest.

**Verify:**

```bash
python3 scripts/check-localizable-xcstrings.py
# xcodebuild test — EvidenceGraphModel / WorkspaceNavigation / GraphCanvasPointerHitTesting
```

### S7-D4 — Design: Citation composer place

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | S7-D3 navigation handoff; S7-03 writers |
| **Deliverables** | Done. Board for navigable viewer\|form place, Artifact pick, citation fields, Observation list (text/term/integer/date), connect-edge Property exclusion, breadcrumbs (`Citation for {scope}`), history keep-composer-entry. Brief archived: [`design/archive/S7-D4-citation-composer.md`](design/archive/S7-D4-citation-composer.md). |
| **Dogfood** | Design only — thin implement in **S7-08**; viewers/locators/NameValue later. |
| **Out** | Image/PDF viewers (S7-06); locator tools (S7-07); NameValue host (S7-D5 / S7-02b); Connect prefill (S7-10). |

**Landed:** Option B composer place contract before thin submit ships.

### S7-08 — PR: Thin citation composer

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S7-09, S7-03, **S7-D4** |
| **Deliverables** | Done. Replaced composer stub with `Features/CitationComposer/` (`CitationComposerView` + `@Observable CitationComposerModel`), then **realigned to Citation Composer board** Frames 1–8 / 11–13: left viewer chrome (source · Artifact N of M · Change; page/zoom/Draw region/Clear locator stubs; locator crumb + errors); right form with summary Observation rows; **Add observation** `PVFormDialog` (text/term/integer/date); Frame 7 Choose artifact Cancel/Continue; Frame 8 no-Artifact gate → Source page + inert form; locator required for Save (page selector; no silent inject). Submit → `createCitationWithObservations` → `CatalogMutation.createdCitation` → Evidence graph; breadcrumb **Citation for {scope}**. |
| **Tests** | `CitationComposerModelTests` (picker Continue, dialog commit, locator + observation gates, edge-property exclude, no-Artifact inert, subject missing); breadcrumb / PlaceRegistry; `check-localizable-xcstrings.py`. |
| **Dogfood** | Add property on Person → composer → (pick Artifact if needed) → Draw region / page locator → Add observation dialog → Save → graph cited row → Back/Forward as normal places. |
| **Out** | Real PDF/image viewers (S7-06); polygon region drawing (S7-07); NameValue (S7-02b); Connect prefilled endpoints (S7-10). |

**Follow-on (same Spike 7 line):** Evidence graph primary-card chrome refresh (cite badge on mark, per-row edit pencil, uncited trash), `GetCitation` / `UpdateCitationWithObservations`, and composer edit mode via `WorkspaceLocation.citationId`.

**Landed:** Board-aligned thin cite path grows cards; viewers and Connect fill in later.

**Verify:**

```bash
python3 scripts/check-localizable-xcstrings.py
# xcodebuild test — CitationComposerModelTests / WorkspaceToolbarBreadcrumbTests / PlaceRegistryTests
```
