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
| [S5-02](#s5-02--pr-interpretation-schema-migration) | PR | `subject_types`, `subjects`, `subject_positions` via migration `000021` |
| [S5-03](#s5-03--pr-subject-type-vocabulary-and-seed) | PR | Seven Subject types seeded at catalog create |
| [S5-04](#s5-04--pr-subject-crud-with-audit) | PR | Audited Subject Create / Update / Delete |
| [S5-05](#s5-05--pr-graph-layout-positions) | PR | Unaudited `subject_positions` Set / Get / Clear |
| [S5-06](#s5-06--pr-ffi-for-subjects-types-and-positions) | PR | Eight catalog RPCs for types, subjects, positions |
| [S5-07](#s5-07--pr-nested-sources-nav-and-evidence-graph-place) | PR | Nested Sources rail + Subject stubs + graph place |
| [S5-08](#s5-08--pr-sources-list--evidence-graph) | PR | Sources list split-row → page or graph; artifact gate |
| [S5-D1](#s5-d1--design-interpretation-nav-entry) | Design | Interpretation sidebar destination — **superseded** |
| [S5-D2](#s5-d2--design-sources-list--evidence-graph) | Design | Sources list dual action → Evidence graph |
| [S5-D3](#s5-d3--design-sources-section-nav) | Design | Nested Sources family nav — Subject types / fields stubs |

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

**Landed (design only):** nav chrome for a top-level Interpretation item. **Do not ship.** See plan revisions and [`design/archive/S5-D3-sources-section-nav.md`](design/archive/S5-D3-sources-section-nav.md).

### S5-D3 — Design: Sources section nav

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | Shipped workspace chrome (S2-01), Source types / Source fields nav |
| **Deliverables** | Done. Board for the nested Sources family: Sources as primary work destination; Source types, Source fields, Subject types, and Subject fields as de-emphasized config children; Subject types / Subject fields stub destinations; no Interpretation item. Brief archived: [`design/archive/S5-D3-sources-section-nav.md`](design/archive/S5-D3-sources-section-nav.md). |
| **Dogfood** | Design only — nothing in the app yet. Implements in S5-07 (sidebar hierarchy + stubs). |
| **Out** | Sources list dual action and Evidence graph stub (S5-D2); Subject types / Subject fields editors; the canvas. |

**Landed:** the nav shape for the unified Sources product layer. S5-07 commits to labels, icons, nesting, and stub destinations from it.

### S5-02 — PR: Interpretation schema migration

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-01 (reserved Subject type prefixes) |
| **Deliverables** | Done. [`core/database/migrations/000021.sql`](../../../core/database/migrations/000021.sql): `subject_types` (two prefixes), `subjects` (NO ACTION FKs to `sources` / `subject_types`), and `subject_positions` (PK `subject_id`, CASCADE on subject delete; unaudited layout). Schema presence test in [`core/database/subjects/schema_test.go`](../../../core/database/subjects/schema_test.go). |
| **Tests** | Done. `TestMigrationCreatesSubjectTables`: `user_version >= 21` and `PRAGMA table_info` columns for all three tables after `database.Create`. |
| **Dogfood** | App unchanged. No seed, CRUD, FFI, or Swift — nothing is user-visible. |
| **Out** | Seeded Subject types (S5-03); Subject CRUD + audit (S5-04); positions query package (S5-05); FFI and UI (S5-06…). |

**Landed:** the Interpretation catalog tables the rest of Spike 5 builds on. Format version is 21; schema hash is init-derived from the new migration (no golden bump).

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/...
```

### S5-03 — PR: Subject type vocabulary and seed

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-02 (`subject_types` table) |
| **Deliverables** | Done. [`core/database/subjecttypes/`](../../../core/database/subjecttypes/): `Upsert` / `Lookup` / `GetByID` / `List` / `Install` with `ref.ValidatePrefix` on both prefixes and cross-column uniqueness (`subjecttypes.duplicate_prefix`). Registry of all seven types from [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1. Wired into [`onboarding.createCatalog`](../../../core/onboarding/ready.go) only — never healed on open. New wire codes `subjecttypes.invalid` and `subjecttypes.duplicate_prefix`. |
| **Tests** | Done. Package: Install → 7 rows, idempotent ids, reserved prefix, same-column and cross-column collisions. Onboarding: `TestCreateCatalogSeedsSubjectTypes`, `TestOpenCatalogDoesNotHealSubjectTypes`. |
| **Dogfood** | Creating a project seeds Subject types in the catalog. No UI lists them yet (S5-07 stubs / Spike 7 editors). |
| **Out** | Subject CRUD + audit (S5-04); positions (S5-05); FFI; properties / `subject_type_fields` seed; Swift L10n for the new codes until a vocabulary browser can raise them. |

**Landed:** create-time Subject type vocabulary so S5-04 can mint candidate refs off `candidate_ref_prefix`.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/subjecttypes/... ./core/onboarding/...
```

### S5-04 — PR: Subject CRUD with audit

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-03 (seeded `subject_types` with `candidate_ref_prefix`) |
| **Deliverables** | Done. [`core/database/subjects/subjects.go`](../../../core/database/subjects/subjects.go): audited `Create` / `Update` / `Delete` / `Get` / `GetByRef` / `ListBySource`. Create reads `candidate_ref_prefix` in-tx and mints with the sources-style 8-attempt unique retry. Audit EntityType `subject`; ActionTypes `create_subject` / `update_subject` / `delete_subject`. `subject_type_id` immutable after insert. Wire code `subjects.invalid`. |
| **Tests** | Done. [`subjects_test.go`](../../../core/database/subjects/subjects_test.go): CPR/CEV minting, audit actions + entity type, update/no-op, delete + position CASCADE, ListBySource scoping, invalid ids. |
| **Dogfood** | App unchanged (no FFI yet). Verifiable by Go tests / sqlite inspection of audit rows. |
| **Out** | Positions query package (S5-05); FFI (S5-06); UI; omnibar search. |

**Landed:** Interpretation Subjects can be created, renamed, and deleted under audit with candidate refs (`CPR-…`, `CEV-…`, …).

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/...
```

### S5-05 — PR: Graph layout positions

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-04 (`subjects` rows to place) |
| **Deliverables** | Done. [`core/database/subjectpositions/`](../../../core/database/subjectpositions/): unaudited `Set` / `Get` / `ListBySource` / `Clear`. PK `subject_id`; graph scope via join on `subjects.source_id`. Absence of a row = tray. No `userID`, no `audit.Record`. Wire code `subjectpositions.invalid`. |
| **Tests** | Done. Round-trip (signed cells), overwrite, tray Clear, ListBySource scoping, unknown subject, no audit growth, Close/Open persistence. |
| **Dogfood** | App unchanged (no FFI yet). Verifiable by Go tests / sqlite. |
| **Out** | FFI (S5-06); canvas / tray UI; FakeStore. |

**Landed:** bubble positions persist outside the audit trail so Spike 6 can drag without drowning revision history.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectpositions/...
```

### S5-06 — PR: FFI for subjects, types, and positions

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-03…S5-05 (types seed, subject CRUD, positions package) |
| **Deliverables** | Done. Proto Methods **45–52** (`ListSubjectTypes`, `Create`/`Update`/`Delete`/`ListSubjects`, `Set`/`Clear`/`ListSubjectPositions`) in [`engine.proto`](../../../api/proto/engine.proto); generated Go + Swift. Dispatch + [`api/ffi/handlers/subjects.go`](../../../api/ffi/handlers/subjects.go) via `withProjectCatalog`. `GenealogyStore` / `GoStore` / `FakeStore` / `ThrowingStore` parity. L10n for `subjects.invalid` and `subjectpositions.invalid` (no Subject-type editor codes). |
| **Tests** | Done. [`subjects_test.go`](../../../api/ffi/handlers/subjects_test.go) `runRPC` coverage (create/list/update/delete, position set/list/clear with session close persistence). [`SubjectStoreTests.swift`](../../../macos/ProvenenciaTests/SubjectStoreTests.swift) FakeStore round-trips. |
| **Dogfood** | App unchanged — no UI destination. Verifiable by Go FFI tests and FakeStore. |
| **Out** | Sidebar / Evidence graph place (S5-07); Sources list Interpret (S5-08); canvas / tray UI; Subject type CRUD RPCs; omnibar Subject search; product SemVer bump. |

**Landed:** Spike 6 can call subject types, subjects, and positions over FFI with FakeStore parity for Swift UI work.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/SubjectStoreTests
```

### S5-07 — PR: Nested Sources nav and Evidence graph place

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-D3 (nav chrome); S5-06 (FFI available, unused by UI) |
| **Deliverables** | Done. Nested Sources-family sidebar in [`PVSidebarNav`](../../../macos/App/DesignSystem/Components/Navigation/PVSidebarNav.swift) / [`WorkspaceSidebar`](../../../macos/App/Features/Workspace/WorkspaceSidebar.swift) (always-expanded config children; no group eyebrow). `WorkspaceSection.subjectTypes` / `.subjectFields` with stub destinations. `SourceSurface` page-vs-graph discriminator on [`WorkspaceLocation`](../../../macos/App/Features/Workspace/WorkspaceLocation.swift) (legacy decode → `.page`). Independent Evidence graph place (`PlaceID.sourceGraph`, `CatalogQueryKey.sourceGraph`, coming-soon stub). Sidebar still highlights **Sources** for both page and graph. |
| **Tests** | Done. `PlaceRegistryTests`, `WorkspaceDestinationHostTests`, `WorkspaceNavigationTests` (surface inequality + legacy Codable), breadcrumb graph leaf. |
| **Dogfood** | Nested Subject types / Subject fields stubs visible in the rail. Graph place registered but not opened from the list yet (S5-08). |
| **Out** | Sources list dual action / no-Artifact gate (S5-08); canvas; Subject type/field editors; omnibar Subjects; SemVer bump. |

**Landed:** Sources family nav matches S5-D3; Spike 6 can route to an Evidence graph place without colliding Source page history.

**Verify:**

```bash
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/PlaceRegistryTests \
  -only-testing:ProvenenciaTests/WorkspaceDestinationHostTests \
  -only-testing:ProvenenciaTests/WorkspaceNavigationTests \
  -only-testing:ProvenenciaTests/WorkspaceToolbarBreadcrumbTests
```

### S5-D2 — Design: Sources list → Evidence graph

| | |
| --- | --- |
| **Kind** | Design (Claude Design board) |
| **Depends on** | S5-D3 (done); shipped Sources list (S2-04) |
| **Deliverables** | Done. Adopted **split-row** direction: sibling filing / Evidence graph zones (`minmax(0,1fr) \| 210px`), caption band, no-Artifact sunken zone + tooltip, custom filter/sort popups. Brief archived: [`design/archive/S5-D2-sources-list-graph-entry.md`](design/archive/S5-D2-sources-list-graph-entry.md). |
| **Dogfood** | Design only until S5-08. |
| **Out** | Canvas; Source-page graph entry; Subject editors. |

**Landed:** dual-action Sources list chrome for S5-08 — split zones, not a nested button.

### S5-08 — PR: Sources list → Evidence graph

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S5-D2 (adopted split-row); S5-07 (graph place + `SourceSurface`) |
| **Deliverables** | Done. `Source.has_artifact` on list/workspace enrich (`artifacts.HasAnyForSource`); `CatalogSource.hasArtifact` + FakeStore/GoStore. Sources list split-row ([`SourcesSplitRow`](../../../macos/App/Features/Sources/SourcesSplitRow.swift)) with page vs graph zones and blocked “Needs an artifact” state; caption band; custom filter/sort via [`PVPopupMenuButton`](../../../macos/App/DesignSystem/Components/Core/PVPopupMenu.swift). Graph destination remains the S5-07 stub. |
| **Tests** | Done. Go `ListSources` `has_artifact` false→true after `CreateArtifact`. Swift [`SourcesListNavigationTests`](../../../macos/ProvenenciaTests/SourcesListNavigationTests.swift) page/graph locations + FakeStore gate. |
| **Dogfood** | From Sources list: open filing page (left zone) or Evidence graph stub (right zone when an Artifact exists); no-Artifact rows show inert graph zone. |
| **Out** | Canvas / Spike 6; Source-page Evidence graph button; SemVer bump; Subject type/field editors. **Floating-menu unify** (history jump onto `PVContextMenu` + open policies) deferred to **S5-09**. |

**Landed:** researcher can open an Evidence graph from the Sources list with an honest Artifact gate.

**Verify:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run TestListSources
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourcesListNavigationTests
```
