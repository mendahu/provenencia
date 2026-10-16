# S10-32 — Depiction value type and `likeness` on Person

**Status:** Draft
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-32--depiction-value-type)
**Brief:** — (no new chrome: an engine value type and a seeded Property; the Properties page only gains one value-type label in an existing pill, and the composer row is S10-34 from S10-D15)
**Board:** —
**Depends on:** —
**Blocks:** S10-33, S10-36
**Branch:** `s10-32-depiction-value-type`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/seeded-vocabulary.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) ("Shape", "Settled"), [`docs/interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (Properties `value_type`, Observation value population), [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.2–3.3, working rules "A depiction's value is its Citation" and "`likeness` is new catalogs only"
**Planned:** 2026-10-16 against `main` @ `536d278d`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A photograph can be evidence whose value is the picture. This PR adds the Observation value type `depiction`: its value is the Observation's own Citation (Artifact, page, region), so no image column and no copied bytes. A new catalog seeds `likeness`, the Person Property of that type, with cardinality `multiple`, so several faces of one person stay side by side as a gallery instead of competing like two birth dates. A Citation with an empty transcription and a region already saves; a likeness Observation on it now saves too.

**Why now.** Slice 8 starts here. The crop cache (S10-33), the composer row (S10-34), the gallery (S10-35), and Promote's crops (S10-36) all key on a depiction Observation, and none exists yet. From [depictions and likenesses](../depictions-and-likenesses.md).

**What stays.** Every existing value type writes, reads, reconciles, and matches as today. Citations still accept an empty transcription, as they do now. The composer does not offer `likeness` yet; S10-34 adds that row. Researchers still cannot create a Property of a registry-only type. Existing catalogs gain no Property; a later one-off CLI updates the dogfood projects (answered 2026-10-16).

## 2. User stories

**US-1.** As a researcher, when I cite a face on a photograph, I record a likeness Observation whose value is that Citation, so the picture is the evidence and nothing is copied.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new catalog. Create a Citation on an Artifact with locator `{"version":1,"selectors":[{"type":"artifact"},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`, transcription `""`, and one Observation: Person subject, Property `likeness`, no value fields | It saves. The Citation's transcription is NULL. The Observation row's `value_text`, `value_integer`, `value_date_id`, `value_name_id`, `value_subject_id`, and `value_term_id` are all NULL. |
| AC-1.2 | Same save as AC-1.1 | The `files` and `file_derivatives` row counts are the same before and after. The likeness stores no raster. |
| AC-1.3 | A likeness Observation that also sets `ValueText: "seated man, left", HasText: true` | Refused with `observations.ErrInvalid`. Nothing is written. |
| AC-1.4 | A likeness Observation with `ValueTermID` set to the `female` term | Refused with `observations.ErrInvalid`. |
| AC-1.5 | A saved likeness Observation is updated through the same path with no value fields | It saves, and every value column stays NULL. |

**US-2.** As a researcher, two likenesses of one person both stay, so a later gallery shows both.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Person *Ann* on Source A and *Ann* on Source B each have one likeness Observation. Promote A to a new Person, then file B onto it | Both promotes succeed. `observations.ListBySubject` returns one likeness for each Subject. |
| AC-2.2 | Same handle as AC-2.1 | `auto_reconciler_values` and `auto_reconciler_outcomes` hold no row for the `likeness` Property on that handle. Rebuild equals upkeep. |
| AC-2.3 | Same handle also has a `name` Observation *Ann Smith* on A | The name row is reconciled as before (`kept`, support 1). |
| AC-2.4 | Two subjects on different Sources, each with one likeness and the same `name` | Promote alignment's exhibits for the pair contain the `name` line and no `likeness` line. |

**US-3.** As a researcher, a new catalog has a Likeness Property on Person, and an older catalog is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Create a catalog | `properties.Lookup(c, "likeness", "provenencia")` is `value_type = depiction`, `cardinality = multiple`, label *Likeness*. It is bound to `person` at sort order 4, and `LockedBinding("person", "likeness")` is true. |
| AC-3.2 | `ListProperties` on a new catalog | 22 Properties. One is `likeness` with value type `depiction` and cardinality `multiple`. |
| AC-3.3 | A catalog created before this PR is opened | The migration widens the CHECK. No `likeness` row is added. |
| AC-3.4 | Researcher Create Property with value type `depiction` | Refused with `properties.ErrInvalid`, the same as `term`. |
| AC-3.5 | Properties page, Person, on a new catalog | The Likeness row's value-type pill reads *Depiction*. Create Property's value-type menu does not offer it. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `properties.value_type` CHECK accepts `depiction` (migration, DDL only) | STEP-1, AC-3.3 |
| `properties.ValueTypeDepiction`; Upsert accepts it; researcher Create refuses it | STEP-2, AC-3.4 |
| Observation write: a depiction value is no value column; any value field is refused | STEP-4, AC-1.1, AC-1.3, AC-1.4, AC-1.5 |
| Empty transcription with a region saves with a likeness | STEP-4, AC-1.1, AC-1.2 |
| Autoreconciler leaves depiction Properties out, so two likenesses never collapse and recompute does not fail | STEP-5, AC-2.1, AC-2.2, AC-2.3 |
| Promote alignment shows no depiction line (guarded by test) | STEP-5, AC-2.4 |
| Seed `likeness` (depiction, multiple) bound to `person`, locked, at Install | STEP-3, AC-3.1, AC-3.2 |
| Swift `PropertyValueType.depiction` and the *Depiction* value-type label | STEP-6, AC-3.5 |
| Domain docs and the seeded-vocabulary skill list the new type | STEP-7 |

| Out | Home |
| --- | --- |
| Place, event, and signature depiction keys | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| The crop bytes and the `file_derivatives` key | S10-33 |
| Adding `depiction` to the composer's `supportedValueTypes` and drawing the row | S10-34 (S10-D15) |
| Person gallery and Persons list thumb | S10-35 |
| Preferred-portrait pointer | S10-39 |
| Promote comparing two crops | S10-36 |
| `likeness` in existing catalogs | Later one-off CLI for the dogfood projects (user, 2026-10-16); not in the map |
| A graph-card or composer rendering for a saved likeness | S10-34 |
| A value-type pill colour for `depiction` | Not planned; it takes the existing default |
| Requiring a region or an Artifact file on a depiction | Not planned (D-4) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Schema | `core/database/migrations/000024.sql` | Rebuilt `properties` with `CHECK (value_type IN ('text','integer','date','name','subject','term'))`. |
| Schema | `core/database/migrations/000043.sql` | Added `cardinality TEXT NOT NULL DEFAULT 'single' CHECK (cardinality IN ('single','multiple'))` to `properties` by `ALTER`. No later migration touches the `properties` DDL. The latest file is `000051.sql`. |
| Migrate runner | `core/database/migrate.go` | Runs each step with `db.Exec` outside a transaction, so `PRAGMA foreign_keys=OFF` works, as `000024.sql` and `000030.sql` use it. |
| Properties | `core/database/properties/properties.go` constants `ValueTypeText`…`ValueTypeTerm`, `valueTypeOK`, `Create` | Six value types. `Upsert` and `Create` reject anything else as `ErrInvalid`. `Create` refuses `ValueTypeTerm` for researchers. |
| Observation write | `core/database/observations/observations.go` `resolveValue`, `hasOtherScalars`, `insertOne` | One `case` per value type; `default` is `ErrInvalid`. `hasOtherScalars(in, keep)` reports any value field other than `keep`. |
| Observation update | `core/database/observations/update.go` `updateOne` | Calls `resolveValue(tx, prop.ValueType, in, &prev)` and writes every value column from the result. |
| Citation | `core/database/citations/citations.go` `CreateWithObservations`, `normalizeCreateInput` | Transcription is trimmed and stored NULL when empty. Only the locator is validated (`core/locator` `Validate`). |
| Autoreconciler | `core/database/autoreconciler/autoreconciler.go` `sqlLoadCandidates`, `load`, `recomputeBatch` | Loads every Observation on accepted and provisional members, grouped by (handle, Property). `hasValue` drops a candidate with no value, but the group still exists, and `recomputeBatch` calls `autoreconcile.Reconcile(g.valueType, g.candidates, nil, g.cardinality)`, which returns `ErrUnknownValueType` for a type outside `knownValueType`. A likeness on a member would fail every recompute of that handle. |
| Reconcile | `core/autoreconcile/autoreconcile.go` `knownValueType`, `Reconcile` | Closed set of six types. |
| Person page fields | `core/database/conclusiondetails/conclusiondetails.go` `sqlPropertyIDs` | Lists a handle's fields from `auto_reconciler_outcomes`. A Property with no outcome row is not listed. |
| Promote exhibits | `core/database/promotealign/exhibits.go` `pairExhibits`, `exhibitCompatible`, `comparerDiscount` | `match.ComparerFor("depiction")` is nil and `autoreconcile.Compatible` is false for an unknown type, so no pair is chosen and no exhibit is written. No test covers it. |
| Matching | `core/match/compare.go` `ComparerFor`; `core/database/matching/matching.go` | Values are read per match-profile Property; `likeness` is in no profile. |
| FFI | `api/ffi/handlers/observations.go` `observationDraftToInputAllowEmptySubject`, `valueInput` | A draft with no value fields becomes an `observations.Input` with none set. Value types travel as strings. No change needed. |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties`, `seedBindings` | 21 Properties. Person bindings: `name` 0 (locked), `sex_at_birth` 1, `gender` 2, `interests` 3. |
| Swift value types | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType` | Six raw cases. |
| Swift parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "value types" | Compares those raw cases with the six Go constants. |
| Properties page | `macos/App/Features/Properties/PropertiesModel.swift` `SubjectPropertyValueType.label`, `researcherCreatable` | `label` maps six types; `default` returns *Text*. A likeness pill would read *Text*. `researcherCreatable` is text, integer, date, name, subject. |
| Pill colour | `macos/App/Features/Catalog/SubjectTypeChrome.swift` `PropertiesTypeChrome.valueTypeForeground` | `default` is `PVColor.textSecondary`. |
| Composer | `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes` | text, integer, date, term, name. The property menu filters on it, so `likeness` is not offered. |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Properties.valueTypeTerm` (`properties.valueType.term`) | One key per value type. |
| Tests | `core/database/observations/observations_test.go` `TestObservations` (with `mustSeed`); `core/database/properties/properties_test.go` `TestProperties`; `core/database/autoreconciler/autoreconciler_test.go` `newFixture`, `TestMultipleCardinalityKeepsDistinctValues`; `core/database/subjectvocab/subjectvocab_test.go` `TestInstallSeedsEngagementTravelInterestsAndGender`; `api/ffi/handlers/subject_defs_test.go` `TestListProperties` (expects 21); `macos/ProvenenciaTests/PropertiesModelTests.swift` | The suites this PR extends. |

**Precedent.** `62f06f09` ("Add Property terms (S7-01b): value_type=term, seed, and FFI CRUD") added a value type: migration `000024.sql` rebuilt `properties` with the wider CHECK, `properties` gained the constant, and `Create` refused it for researchers. Copy the rebuild and the refusal. Do not copy its new table or FFI CRUD. The seed is shaped like `interests` in `3d701390`: one `seedProperties` row, one `seedBindings` row, a `subjectvocab_test.go` case, and the `ListProperties` count in `f06250ae`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000052.sql` rebuilds `properties` with `value_type IN ('text','integer','date','name','subject','term','depiction')` and keeps `cardinality TEXT NOT NULL DEFAULT 'single' CHECK (cardinality IN ('single','multiple'))`. Copy the `000024.sql` shape: `PRAGMA foreign_keys=OFF`, `properties__new`, `INSERT … SELECT` every column, `DROP`, `RENAME`, `PRAGMA foreign_keys=ON`. DDL only, no `INSERT` of vocabulary. | SQLite cannot `ALTER` a CHECK. The map's "Install path only" covers the seed row, not the type's DDL. | Dropping the CHECK — because the schema then accepts any string. A trigger — because no other value rule uses one. |
| D-2 | A depiction value is no value column. In `resolveValue`, `case properties.ValueTypeDepiction` returns `resolvedValue{}` when `hasOtherScalars(in, "depiction")` is false, otherwise `ErrInvalid`. The Citation the Observation already points at is the value. | Working rule "A depiction's value is its Citation". `hasOtherScalars` with a keep that matches no field already flags every set field. | A `value_citation_id` column — because it would always equal `citation_id`. |
| D-3 | `likeness` is new catalogs only. Install seeds it; the migration does not; open does not heal. | User, 2026-10-16. A later one-off CLI updates the dogfood projects. `add-catalog-migration` "Do not" forbids seeds in SQL. | Seeding it in `000052.sql` — user declined. |
| D-4 | A depiction Observation does not require a region, a page, or a file on the Artifact. | A single portrait is the whole frame. The Citation can move to another Artifact later, so a file check at insert would not hold. S10-33 and S10-35 handle an Artifact they cannot draw. | Requiring a `region` selector — because "the whole frame can be a depiction" ([depictions and likenesses](../depictions-and-likenesses.md)). |
| D-5 | Polarity is not special-cased. A negative likeness saves like any negative Observation. | No rule says otherwise. What a gallery shows is S10-35's call. | Refusing negative depictions — no source asks for it. |
| D-6 | Researcher Create refuses `depiction`, next to `term`, with `ErrInvalid`. Swift `researcherCreatable` stays unchanged. | Place, event, and signature keys wait (map Out). A user depiction Property would have no composer row and no gallery. | Allowing it — the user chose new-catalogs-only and a CLI instead (2026-10-16). |
| D-7 | The autoreconciler loads no depiction Observation: `sqlLoadCandidates` adds `AND p.value_type <> 'depiction'`. `autoreconcile.knownValueType` stays six types. | A gallery is not reconciled ("Gallery, not a single concluded face"). Filtering at the load keeps `Reconcile` a closed set and means no group exists to fail. With no outcome rows the Person page lists no empty Likeness field, so `conclusiondetails` needs no change. | Adding a no-op module to `autoreconcile` — because a type that never yields values has no reason to enter `Reconcile`. Skipping in `recomputeBatch` — because `load` would still build the group. |
| D-8 | Promote alignment is not changed. T-6 pins today's behavior: no likeness exhibit. | `pairExhibits` already drops it (§4). S10-36 adds the crop comparison. | Adding a `depiction` skip in `pairExhibits` — because the code path already skips it. |
| D-9 | Seed row: `{Key: "likeness", Label: "Likeness", Description: "A picture of this person. The value is the Citation: its Artifact, page, and region. More than one is allowed.", ValueType: properties.ValueTypeDepiction, Cardinality: properties.CardinalityMultiple}`, after `interests`. Binding `{TypeKey: "person", PropertyKey: "likeness", SortOrder: 4, Locked: true}`. | Locked because S10-35's gallery and S10-39's pointer key into that slot, like `name` and `toponym`. | Unlocked like `interests` — because unbinding it would orphan the gallery. |
| D-10 | Swift: add `case depiction` to `PropertyValueType`. `SubjectPropertyValueType.label` maps `"depiction"` to `L10n.Properties.valueTypeDepiction` (key `properties.valueType.depiction`, value *Depiction*). `valueTypeForeground` is not changed. | Without the label the pill reads *Text*. A pill colour is a design call with no brief. | A new colour case — no board draws it. |
| D-11 | The `TestSwiftVocabularyMatchesGo` "value types" want list adds `properties.ValueTypeDepiction`. | The Swift enum and Go must agree. | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000052.sql` *(new)* | migration | `core/database/migrations/000024.sql` | Extends that shape. A shipped migration is never edited. |
| `properties.ValueTypeDepiction` *(new)* | const | `properties.ValueTypeTerm` | Extends the same constant block and `valueTypeOK`. |
| `case properties.ValueTypeDepiction` in `resolveValue` *(new)* | switch case | `case properties.ValueTypeSubject` | Extends the one switch every write uses. |
| `AND p.value_type <> 'depiction'` in `sqlLoadCandidates` *(new)* | SQL filter | `WHERE ic.status IN (…)` in the same query | Extends the one loader both upkeep and `Rebuild` use. |
| `likeness` row and binding *(new)* | seed rows | `interests` in `seedProperties` / `seedBindings` | Registry entries. No new Install code. |
| `PropertyValueType.depiction` *(new)* | Swift case | `PropertyValueType.term` | Extends the enum the parity test reads. |
| `L10n.Properties.valueTypeDepiction` *(new)* | L10n key | `L10n.Properties.valueTypeTerm` | One key per value type, same namespace. |
| Tests T-1 through T-9 | subtests / test funcs | `TestObservations`, `TestMultipleCardinalityKeepsDistinctValues`, `TestInstallSeedsEngagementTravelInterestsAndGender` | Each extends the suite and fixture that already covers that layer. No new helper. |

**Builds on.** `add-catalog-migration` rebuild pattern (`000024.sql`); `add-seeded-vocabulary` registry (`subjectvocab` `seedProperties` / `seedBindings`); the registry-only refusal in `properties.Create`; `TestSwiftVocabularyMatchesGo` parity.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-33 selects Observations whose Property is `value_type = depiction` to mint crops. S10-34 adds `PropertyValueType.depiction.rawValue` to `CitationComposerVocabulary.supportedValueTypes` and a `case` in `CitationObservationValue`. S10-35 joins accepted claims to depiction Observations. S10-36 gives `pairExhibits` a depiction case. Place and event keys are one `seedProperties` row and one `seedBindings` row each.

### Schema

Migration `core/database/migrations/000052.sql` *(new)*:

```sql
-- S10-32: value_type = depiction. The value is the Observation's Citation.
-- SQLite cannot ALTER CHECK; rebuild properties. likeness is seeded by
-- subjectvocab.Install at create time, not here.
PRAGMA foreign_keys=OFF;

CREATE TABLE properties__new (
	id          BLOB PRIMARY KEY,
	key         TEXT NOT NULL,
	origin      TEXT NOT NULL,
	label       TEXT NOT NULL,
	description TEXT,
	value_type  TEXT NOT NULL,
	cardinality TEXT NOT NULL DEFAULT 'single'
		CHECK (cardinality IN ('single', 'multiple')),

	UNIQUE (key, origin),
	CHECK (value_type IN (
		'text', 'integer', 'date', 'name', 'subject', 'term', 'depiction'
	))
) STRICT;

INSERT INTO properties__new (id, key, origin, label, description, value_type, cardinality)
SELECT id, key, origin, label, description, value_type, cardinality FROM properties;

DROP TABLE properties;
ALTER TABLE properties__new RENAME TO properties;

PRAGMA foreign_keys=ON;
```

`properties` has no secondary index to restore (the only indexes are its PK and `UNIQUE (key, origin)`). `catalogmodel` is unchanged: same table, same columns, same FKs into it. Delete class and audit: unchanged. The schema hash is derived from the migrations; do not touch it.

### Go core

| Package | Change |
| --- | --- |
| `core/database/properties` | `ValueTypeDepiction = "depiction"` *(new)*. `valueTypeOK` accepts it. `Create` refuses `ValueTypeTerm` or `ValueTypeDepiction` with `ErrInvalid`; update its doc comment to name both. |
| `core/database/observations` | `resolveValue`: `case properties.ValueTypeDepiction` per D-2. No change to `insertOne`, `updateOne`, or the SQL. |
| `core/database/autoreconciler` | `sqlLoadCandidates`: `AND p.value_type <> 'depiction'` after the status filter, with a one-line comment that a gallery is not reconciled (D-7). |
| `core/database/subjectvocab` | `seedProperties` and `seedBindings` rows per D-9. |

Errors: `observations.invalid` → a depiction Observation sets a value field. `properties.invalid` → researcher Create with `depiction`.

### FFI

None — value types travel as strings, and a draft with no value fields already maps to an empty `observations.Input` (§4). No proto change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value-type enum | `PropertyValueType` | `macos/App/Platform/InterpretationValueKinds.swift` | `case depiction` after `term`. |
| Value-type label | `SubjectPropertyValueType.label` | `macos/App/Features/Properties/PropertiesModel.swift` | `case "depiction": return L10n.Properties.valueTypeDepiction`. `researcherCreatable` unchanged. |

States: none new. The Properties page row and pill already exist. Accessibility: the pill's `accessibilityLabel` already reads `SubjectPropertyValueType.label`, so it says *Depiction*.

### Strings

L10n key *(new)*: `L10n.Properties.valueTypeDepiction` = "Depiction" (`properties.valueType.depiction`, comment "Property value type label: depiction (registry-only)"), beside `valueTypeTerm`, with its `Localizable.xcstrings` entry. Removed: none.

## 7. Steps

**STEP-1 — Widen the value_type CHECK**
Files: `core/database/migrations/000052.sql` *(new)*. Skill: `add-catalog-migration`.
Do: write the migration in §6 Schema exactly. No test for the migration itself.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `TestPragmaHonesty`).

**STEP-2 — Add the constant and the Create refusal**
Files: `core/database/properties/properties.go`, `core/database/properties/properties_test.go`.
Do: D-6 and §6 Go core row 1. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/...` passes.

**STEP-3 — Seed `likeness`**
Files: `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/subjectvocab_test.go`, `api/ffi/handlers/subject_defs_test.go`. Skill: `add-seeded-vocabulary`.
Do: D-9. Add T-7. In T-8 change the expected count from 21 to 22 and assert the `likeness` row.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/... ./api/ffi/handlers/...` passes.

**STEP-4 — Write a depiction Observation**
Files: `core/database/observations/observations.go`, `core/database/observations/observations_test.go`.
Do: D-2. Add T-2, T-3, and T-4 as subtests of `TestObservations`. `mustSeed` already runs `subjectvocab.Install`, so read `likeness` with `properties.Lookup(c, "likeness", properties.OriginProvenencia)`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/... ./core/database/citations/...` passes.

**STEP-5 — Keep depictions out of reconciliation**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-7. Add T-5 with the autoreconciler `newFixture` (`f.cite`, `f.promote`, `f.join`, `f.rows`, `f.outcomes`, `f.assertUpkeepEqualsRebuild`). Add T-6 with the promotealign `newFixture` (`f.person`, `f.cite`, `f.prop("likeness")`, `f.promote`, `f.propose`, `rowFor`) for D-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/... ./core/database/promotealign/...` passes.

**STEP-6 — Swift value type and label**
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/PropertiesModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-10 and D-11. Add T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` (ProvenenciaTests) passes.

**STEP-7 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/seeded-vocabulary.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do: in `interpretation-layer-data-model.md`, add `'depiction'` to the `properties` DDL sketch and the product value-type list; state that a depiction populates no value column and its value is `citation_id`, as the one exception to "exactly one value representation"; add `depiction` beside `term` in "Who may create". In `seeded-vocabulary.md` §3.2 add the `likeness  depiction  …; cardinality multiple; locked on person; new catalogs only` row, and in §3.3 `person  likeness  # depiction; multiple; locked`. In the skill, add `depiction` to the value-type line and to the registry-only sentence. Then the Definition of done lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/properties/properties_test.go` `TestProperties` subtest "depiction is registry-only" *(new)* | `Upsert` `{Key:"portrait", Origin:"provenencia", ValueType:"depiction"}` → ok; `Create(…, "Portrait", "depiction", "", "")` → `ErrInvalid` | AC-3.4, D-6 |
| T-2 | Go | `core/database/observations/observations_test.go` `TestObservations` subtest "likeness saves with no value and no transcription" *(new)* | AC-1.1 locator, transcription `""`, likeness Observation with no value → saved; citation transcription NULL; all six value columns NULL; `files` and `file_derivatives` counts unchanged | AC-1.1, AC-1.2, D-4 |
| T-3 | Go | same file, subtest "likeness refuses a value" *(new)*, table-driven | `HasText` "seated man, left" → `ErrInvalid`; `ValueTermID` female → `ErrInvalid`; nothing written | AC-1.3, AC-1.4, D-2 |
| T-4 | Go | same file, subtest "likeness update keeps no value" *(new)* | update the T-2 Observation with no value fields → saved; value columns NULL | AC-1.5 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestDepictionIsNotReconciled` *(new)* | `newFixture`; Ann on `f.source` and on `f.newSource("Album")`, each with a likeness and Ann on the first with name *Ann Smith*; promote, join → no row for `f.props["likeness"]` in `f.rows()` or outcomes; both likenesses in `ListBySubject`; name row `kept`, support 1; `f.assertUpkeepEqualsRebuild` | AC-2.1, AC-2.2, AC-2.3, D-7 |
| T-6 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeShowsNoLikenessExhibit` *(new)* | two Persons on different Sources, same `name`, one likeness each, one promoted → the proposal's exhibits have a `name` line and none for `likeness` | AC-2.4, D-8 |
| T-7 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestInstallSeedsLikeness` *(new)* | `Install` → `likeness` is depiction, multiple, label *Likeness*; person binding sort 4; `LockedBinding("person","likeness")` | AC-3.1, D-9 |
| T-8 | Go | `api/ffi/handlers/subject_defs_test.go` `TestListProperties` | new catalog → 22 Properties; `likeness` depiction multiple | AC-3.2 |
| T-9 | Swift | `macos/ProvenenciaTests/PropertiesModelTests.swift` `depictionLabelIsNotCreatable` *(new)* | `SubjectPropertyValueType.label("depiction")` is `L10n.Properties.valueTypeDepiction`; `researcherCreatable` lacks `"depiction"` | AC-3.5, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSwiftVocabularyMatchesGo` (until D-11), `TestListProperties` (count), `catalogmodel.TestPragmaHonesty` and the schema hash (STEP-1), `TestRebuildEqualsUpkeep_Scenarios` and `TestRebuildEqualsUpkeep_SeededSequences`.

AC-3.3 has no test: migrations are not tested directly (`add-catalog-migration`); the schema hash self-check covers the DDL, and the manual check covers an older catalog.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: only the label in D-10)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-32 status and SLICE 8 line; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: interpretation-layer-data-model.md (value types, value population), seeded-vocabulary.md §3.2–3.3, add-seeded-vocabulary skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Create a new project. Open **Properties**, select **Person**. A **Likeness** row is listed; its value-type pill reads *Depiction*, it allows more than one value, and its binding cannot be removed. (AC-3.1, AC-3.5)
2. Click **Create property**. The value-type menu lists Text, Integer, Date, Name, and Subject, and not Depiction. (AC-3.5)
3. Open a Source with a Person, open the composer, add an Observation on the Person. The Property menu does not list Likeness. (S10-34 adds it.)
4. Open a project created before this build. It opens. **Properties → Person** has no Likeness row. (AC-3.3)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rebuild drops a column or default added since `000043.sql` | §4 lists the full column set; T-1 and every `properties` test read `cardinality` after the migration. |
| Another reader errors on an Observation with no value column | §4 lists every reader that branches on value type; only `autoreconcile.Reconcile` errors, and D-7 keeps depictions out of it. |
| A likeness Observation on the graph card shows an empty value | None can be made in the app before S10-34 (manual check 3). |

**Stop and report if:**

- A migration after `000043.sql` changed the `properties` DDL, so the §6 sketch would drop a column.
- `TestPragmaHonesty` or the schema-hash check fails after STEP-1 and the fix needs a change outside `000052.sql`.
- Any test outside `core/autoreconcile` fails with `ErrUnknownValueType` or `observations.invalid` for a likeness row after STEP-5, which means a reader §4 does not list branches on value type.
- `pairExhibits` emits a likeness line in T-6, which proves D-8 wrong.
- `xcodebuild test` fails on a `switch` over `PropertyValueType` that is not exhaustive, which means a Swift switch §4 does not list.

## 12. Open questions

None.

## 13. PR

**Title:** S10-32 — Depiction value type and likeness on Person
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md`.
