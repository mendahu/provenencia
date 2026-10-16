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
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/seeded-vocabulary.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) ("Shape", "Settled"), [`docs/interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (Properties `value_type`, Observation value population), [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §3.2–3.3, working rules "A depiction's value is its Citation", "`likeness` is new catalogs only", and "A gallery is the reconciled likeness values", [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5–§7
**Planned:** 2026-10-16 against `main` @ `536d278d`
**Review:** [`S10-32-depiction-value-type.review.md`](S10-32-depiction-value-type.review.md) — round 1 on the first shape answered; re-review of the reconciled shape pending

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

A photograph can be evidence whose value is the picture. This PR adds the Observation value type `depiction`: its value is the Observation's own Citation (Artifact, page, region), so no image column and no copied bytes. A new catalog seeds `likeness`, the Person Property of that type, with cardinality `multiple`, so several faces of one person stay side by side as a gallery instead of competing like two birth dates. The auto-reconciler gets a `depiction` module like every other value type: a value is the Citation, so two Observations of one Citation fold, a negative Observation of that Citation denies it, a provisional member's crop is cached as `provisional`, and a crop carried only by weak evidence is cached as `weak`. The handle's cached likeness values are the gallery S10-35 draws; it shows `kept` and `weak` crops (answered 2026-10-16). A Citation with an empty transcription and a region already saves; a likeness Observation on it now saves too.

**Why now.** Slice 8 starts here. The crop cache (S10-33), the composer row (S10-34), the gallery (S10-35), and Promote's crops (S10-36) all key on a depiction Observation, and none exists yet. From [depictions and likenesses](../depictions-and-likenesses.md).

**What stays.** Every existing value type writes, reads, reconciles, and matches as today, and every existing catalog's cached values are rebuilt unchanged. The Person page does not show a Likeness field row; S10-35 draws the gallery. Citations still accept an empty transcription, as they do now. The composer does not offer `likeness` yet; S10-34 adds that row. Researchers still cannot create a Property of a registry-only type. Existing catalogs gain no Property; a later one-off CLI updates the dogfood projects (answered 2026-10-16).

## 2. User stories

**US-1.** As a researcher, when I cite a face on a photograph, I record a likeness Observation whose value is that Citation, so the picture is the evidence and nothing is copied.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new catalog. Create a Citation on an Artifact with locator `{"version":1,"selectors":[{"type":"artifact"},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.5}]}]}`, transcription `""`, and one Observation: Person subject, Property `likeness`, no value fields | It saves. The Citation's transcription is NULL. The Observation row's `value_text`, `value_integer`, `value_date_id`, `value_name_id`, `value_subject_id`, and `value_term_id` are all NULL. |
| AC-1.2 | Same save as AC-1.1 | The `files` and `file_derivatives` row counts are the same before and after. The likeness stores no raster. |
| AC-1.3 | A likeness Observation that also sets `ValueText: "seated man, left", HasText: true` | Refused with `observations.ErrInvalid`. Nothing is written. |
| AC-1.4 | A likeness Observation with `ValueTermID` set to the `female` term | Refused with `observations.ErrInvalid`. |
| AC-1.5 | A saved likeness Observation is updated through the same path with no value fields | It saves, and every value column stays NULL. |

**US-2.** As a researcher, a person's likenesses are reconciled into a gallery: every distinct crop stays, and the evidence rules that hold for names and dates hold for faces.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Person *Ann* on Source A (Citation `CA`) and *Ann* on Source B (Citation `CB`) each have one likeness Observation. Promote A to a new Person, then file B onto it | `auto_reconciler_values` holds two `likeness` rows for the handle, `value_citation_id` `CA` and `CB`, both `kept`, support 1, and no other value column set. Both Observations have a `kept` outcome. |
| AC-2.2 | One Citation `CA` on Source A carries two likeness Observations, on Subject *Ann* and Subject *Ann (2)*, both filed onto the handle | One `likeness` row, `value_citation_id` `CA`, support 1 (one Source); both outcomes have that row's rank. |
| AC-2.3 | AC-2.1, where Citation `CB` also carries a negative likeness Observation on a third Subject *Ann (3)* of Source B, filed onto the handle with a `high_confidence` claim. (A negative denies only when its evidence is stronger, and on one Citation only the claim confidence can differ.) | The `CB` row's reason is `denied`; its positive Observation's outcome is `denied` with `denied_by` the negative Observation. The `CA` row stays `kept`. |
| AC-2.4 | AC-2.1, but Ann on B is a provisional member | The `CB` row's reason is `provisional`, and so is B's likeness outcome. The `CA` row stays `kept`. |
| AC-2.5 | AC-2.1, with Source B graded `low_trust` | `CA` `kept`, `CB` `weak`. Both rows are stored; the gallery shows both (working rule, 2026-10-16). |
| AC-2.6 | Each of AC-2.1 to AC-2.5 | Rebuild equals upkeep for values and outcomes. The graph cache's `Verify` passes. |
| AC-2.7 | The AC-2.1 handle also has a `name` Observation *Ann Smith* on A | The name row is reconciled as before (`kept`, support 1). The Person detail read (`conclusiondetails`) lists `name` and no `likeness` field. |
| AC-2.8 | Two subjects on different Sources, each with one likeness and the same `name`, one promoted | Promote alignment's exhibits for the pair contain the `name` line and no `likeness` line. |
| AC-2.9 | A catalog whose stored cache version is 14 is opened | The cache rebuilds once and stores version 15. |

**US-3.** As a researcher, a new catalog has a Likeness Property on Person, and an older catalog is unchanged.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Create a catalog | `properties.Lookup(c, "likeness", "provenencia")` is `value_type = depiction`, `cardinality = multiple`, label *Likeness*. It is bound to `person` at sort order 4, and `LockedBinding("person", "likeness")` is true. |
| AC-3.2 | `ListProperties` on a new catalog | 22 Properties. One is `likeness` with value type `depiction` and cardinality `multiple`. |
| AC-3.3 | A catalog created before this PR is opened | The migration widens the CHECK. No `likeness` row is added. |
| AC-3.4 | Researcher Create Property with value type `depiction` | Refused with `properties.ErrInvalid`, the same as `term`. |
| AC-3.5 | Properties page, Person, on a new catalog | The Likeness row's pill shows the raw key `depiction`, and VoiceOver reads it as *Depiction*. Create Property's value-type chips do not offer it. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `properties.value_type` CHECK accepts `depiction` (migration, DDL only) | STEP-1, AC-3.3 |
| `properties.ValueTypeDepiction`; Upsert accepts it; researcher Create refuses it | STEP-2, AC-3.4 |
| Observation write: a depiction value is no value column; any value field is refused | STEP-4, AC-1.1, AC-1.3, AC-1.4, AC-1.5 |
| Empty transcription with a region saves with a likeness | STEP-4, AC-1.1, AC-1.2 |
| `autoreconcile` `depiction` module keyed by the Citation; `Value.CitationID` | STEP-5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| Loader fills `CitationID`; cache column `auto_reconciler_values.value_citation_id`; `CacheVersion` 15 | STEP-1, STEP-6, AC-2.1, AC-2.6, AC-2.9 |
| Graph cache mirrors the new column, so `Verify` covers it | STEP-6, AC-2.6 |
| Person detail read skips depiction fields | STEP-6, AC-2.7 |
| Promote alignment shows no depiction line (guarded by test) | STEP-6, AC-2.8 |
| Seed `likeness` (depiction, multiple) bound to `person`, locked, at Install | STEP-3, AC-3.1, AC-3.2 |
| Swift `PropertyValueType.depiction` and the *Depiction* value-type label; `label` and `researcherCreatable` read the typed enum | STEP-7, AC-3.5 |
| `properties.registryOnly` names the registry-only types once | STEP-2, AC-3.4 |
| Domain docs (including the reconciliation docs) and the seeded-vocabulary skill list the new type | STEP-8 |

| Out | Home |
| --- | --- |
| Place, event, and signature depiction keys | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| The crop bytes and the `file_derivatives` key | S10-33 |
| Adding `depiction` to the composer's `supportedValueTypes` and drawing the row | S10-34 (S10-D15) |
| Person gallery and Persons list thumb, read from the cached `kept` and `weak` likeness values | S10-35 |
| A Reconciliation Claim on a depiction Property | Not in this spike (working rule "Support weighting and negatives land together": Reconciliation Claims are not in this spike) |
| Reading `graphcache.Value.CitationID` in Promote | S10-36 |
| Preferred-portrait pointer | S10-39 |
| Promote comparing two crops | S10-36 |
| `likeness` in existing catalogs | Later one-off CLI for the dogfood projects (user, 2026-10-16); not in the map |
| A graph-card or composer rendering for a saved likeness | S10-34 |
| A value-type pill colour for `depiction`, and moving `PropertiesTypeChrome.valueTypeForeground` onto the typed enum | Not planned; it takes the existing default |
| Moving `CitationObservationValue`'s value-type switches onto the typed enum | S10-34, which adds the depiction case there |
| A parity subtest that Swift `researcherCreatable` matches Go | Logged in the review log (M-3); not planned |
| Requiring a region or an Artifact file on a depiction | Not planned (D-4) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Schema | `core/database/migrations/000024.sql` | Rebuilt `properties` with `CHECK (value_type IN ('text','integer','date','name','subject','term'))`. |
| Schema | `core/database/migrations/000043.sql` | Added `cardinality TEXT NOT NULL DEFAULT 'single' CHECK (cardinality IN ('single','multiple'))` to `properties` by `ALTER`. No later migration touches the `properties` DDL. The latest file is `000051.sql`. |
| Migrate runner | `core/database/migrate.go` | Runs each step with `db.Exec` outside a transaction, so `PRAGMA foreign_keys=OFF` works, as `000024.sql` and `000030.sql` use it. |
| Properties | `core/database/properties/properties.go` constants `ValueTypeText` through `ValueTypeTerm`, `valueTypeOK`, `Create` | Six value types. `Upsert` and `Create` reject anything else as `ErrInvalid`. `Create` refuses `ValueTypeTerm` for researchers with an inline `if valueType == ValueTypeTerm`. |
| Observation write | `core/database/observations/observations.go` `resolveValue`, `hasOtherScalars`, `insertOne` | One `case` per value type; `default` is `ErrInvalid`. `hasOtherScalars(in, keep)` reports any value field other than `keep`. |
| Observation update | `core/database/observations/update.go` `updateOne` | Calls `resolveValue(tx, prop.ValueType, in, &prev)` and writes every value column from the result. |
| Citation | `core/database/citations/citations.go` `CreateWithObservations`, `normalizeCreateInput` | Transcription is trimmed and stored NULL when empty. Only the locator is validated (`core/locator` `Validate`). |
| Autoreconciler | `core/database/autoreconciler/autoreconciler.go` `sqlLoadCandidates`, `load`, `hasValue`, `insertRow`, `sqlInsert`, `CacheVersion` (14) | Loads every Observation on accepted and provisional members, grouped by (handle, Property); `sqlLoadCandidates` already joins `citations c` but does not select `c.id`. `hasValue` drops a candidate with no value, but the group still exists, and `recomputeBatch` calls `autoreconcile.Reconcile(g.valueType, g.candidates, nil, g.cardinality)`, which returns `ErrUnknownValueType` for a type outside `knownValueType`. A likeness on a member would fail every recompute of that handle. `insertRow` writes one column per value kind. Every value is stored with its reason, including `provisional` and `denied` ones. `CacheVersion` is bumped "whenever resolution changes what the table would hold"; a differing stored version rebuilds on open. |
| Truth recompute | `core/database/autoreconciler/truth.go` `truthRows`, `truthRow` | Runs the same `load` and `Reconcile` for `graphcache` `Verify`; `truthRow` copies each value field into `graphcache.TruthRow`. |
| Reconcile | `core/autoreconcile/autoreconcile.go` `Value`, `knownValueType`, `carries`, `Reconcile`; `core/autoreconcile/modules.go` `moduleFor`, `keyModule`, `termKey` | Closed set of six types. `Value` has no Citation field. `keyModule{keyOf}` is a value type whose values are the same exactly when their keys are equal (text, integer, term, subject). `moduleFor` returns nil for an unknown type. |
| Cache table | `core/database/migrations/000034.sql` (`conclusion_resolved_values`), `000038.sql`, `000039.sql` (`ADD COLUMN`), `000040.sql` (rename to `auto_reconciler_values`) | Value columns `value_text`, `value_integer`, `value_term_id`, `value_entity_id`, `value_date`, `value_name`, with no FK. `catalogmodel` lists the table as `BucketSkip` with FKs only on `entity_id` and `property_id`. |
| Graph cache | `core/database/graphcache/graph.go` `Value`; `core/database/graphcache/load.go` `loadValues`; `core/database/graphcache/verify.go` `TruthRow`, `valueEqual` | `loadValues` reads every cached value of a handle into `Value`; `Verify` compares each against `TruthRow` field by field. A new cache column that is not mirrored is invisible to `Verify`. |
| Cache readers | `core/database/matching/matching.go` `scanEntityValues`; `core/database/promotealign/canon.go` `handleValues`; `core/database/promotealign/stats.go`; `core/database/searchindex/handles.go`; `core/database/conclusionheaders/dependents.go`; `core/database/canonicalgraph/canonicalgraph.go` | Each selects by Property key, by match profile, or by `value_entity_id IS NOT NULL`. A `likeness` row is outside every filter; `handleValues` builds an empty `match.Value` for it, which no profile reads. |
| Reconcile docs | `docs/conclusion-reconciliation.md` §7 table, §12 Implementation status; `docs/conclusion-layer-data-model.md` ("Exactly one value representation must be populated") | One §7 row per value type; this PR adds `depiction`. §1 "Every value type needs an auto-reconciler" stays true. |
| Person page fields | `core/database/conclusiondetails/conclusiondetails.go` `sqlPropertyIDs` and the field loop that skips `p.ValueType == properties.ValueTypeSubject` | Lists a handle's fields from `auto_reconciler_outcomes`, skipping subject-valued Properties. With a depiction module, a likeness field would be listed with values the Swift `CatalogConclusionValue` cannot show. |
| Promote exhibits | `core/database/promotealign/exhibits.go` `pairExhibits`, `exhibitCompatible`, `comparerDiscount` | `match.ComparerFor("depiction")` is nil, and `exhibitCompatible` goes through `match.Value.Auto()`, which carries no Citation, so `autoreconcile.Compatible` is false even after D-7. No pair is chosen and no exhibit is written. No test covers it. |
| Matching | `core/match/compare.go` `ComparerFor`; `core/database/matching/matching.go` | Values are read per match-profile Property; `likeness` is in no profile. |
| FFI | `api/ffi/handlers/observations.go` `observationDraftToInputAllowEmptySubject`, `valueInput` | A draft with no value fields becomes an `observations.Input` with none set. Value types travel as strings. No change needed. |
| Seed | `core/database/subjectvocab/registry.go` `seedProperties`, `seedBindings` | 21 Properties. Person bindings: `name` 0 (locked), `sex_at_birth` 1, `gender` 2, `interests` 3. |
| Swift value types | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType` | Six raw cases. |
| Swift parity test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` "value types" | Compares those raw cases with the six Go constants. |
| Properties page | `macos/App/Features/Properties/PropertiesModel.swift` `SubjectPropertyValueType.label`, `researcherCreatable` | `label(_ valueType: String)` switches on raw strings; `default` returns *Text*. `researcherCreatable` is `[String]` text, integer, date, name, subject, read at `PropertiesModel.swift` (`canSubmit` and `submitCreate`) and `PropertiesView.swift` (the value-type chips). The doc comment above says only `term` is registry-only. |
| Pill | `macos/App/Features/Properties/PropertiesSnapshot.swift` `PropertiesValueTypePill` | Shows `Text(verbatim: valueType)`, the raw key. Its `accessibilityLabel`, and the row's accessibility string in `PropertiesView.swift`, use `SubjectPropertyValueType.label`, so VoiceOver would say *Text* for a likeness. |
| Pill colour | `macos/App/Features/Catalog/SubjectTypeChrome.swift` `PropertiesTypeChrome.valueTypeForeground` | `default` is `PVColor.textSecondary`. |
| Composer | `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes` | text, integer, date, term, name. The property menu filters on it, so `likeness` is not offered. |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Properties.valueTypeTerm` (`properties.valueType.term`) | One key per value type. |
| Tests | `core/database/observations/observations_test.go` `TestObservations` (with `mustSeed`); `core/database/properties/properties_test.go` `TestProperties`; `core/database/autoreconciler/autoreconciler_test.go` `newFixture`, `snapshot`, `row`, `TestMultipleCardinalityKeepsDistinctValues`, `TestEnsureCatalogRebuildsStaleVersion`, the claim-status cases in `TestReconciledEvidence`; `core/autoreconcile/modules_test.go` `TestModules`, `TestModuleFor`; `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityFieldsFollowTheRecords`; `core/database/subjectvocab/subjectvocab_test.go` `TestInstallSeedsEngagementTravelInterestsAndGender`; `api/ffi/handlers/subject_defs_test.go` `TestListProperties` (expects 21); `macos/ProvenenciaTests/PropertiesModelTests.swift` | The suites this PR extends. |

**Precedent.** `62f06f09` ("Add Property terms (S7-01b): value_type=term, seed, and FFI CRUD") added a value type: migration `000024.sql` rebuilt `properties` with the wider CHECK, `properties` gained the constant, and `Create` refused it for researchers. Copy the rebuild and the refusal. Do not copy its new table or FFI CRUD. The cache column follows `1851994c` (migration `000038.sql`, `ADD COLUMN` on the reconciler cache) and the `CacheVersion` history in `autoreconciler.go`; the module follows `keyModule{keyOf: termKey}` in `modules.go`. The seed is shaped like `interests` in `3d701390`: one `seedProperties` row, one `seedBindings` row, a `subjectvocab_test.go` case, and the `ListProperties` count in `f06250ae`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000052.sql` rebuilds `properties` with `value_type IN ('text','integer','date','name','subject','term','depiction')` and keeps `cardinality TEXT NOT NULL DEFAULT 'single' CHECK (cardinality IN ('single','multiple'))`. Copy the `000024.sql` shape: `PRAGMA foreign_keys=OFF`, `properties__new`, `INSERT … SELECT` every column, `DROP`, `RENAME`, `PRAGMA foreign_keys=ON`. DDL only, no `INSERT` of vocabulary. | SQLite cannot `ALTER` a CHECK. The map's "Install path only" covers the seed row, not the type's DDL. | Dropping the CHECK — because the schema then accepts any string. A trigger — because no other value rule uses one. |
| D-2 | A depiction value is no value column. In `resolveValue`, `case properties.ValueTypeDepiction` returns `resolvedValue{}` when `hasOtherScalars(in, "depiction")` is false, otherwise `ErrInvalid`. The Citation the Observation already points at is the value. | Working rule "A depiction's value is its Citation". `hasOtherScalars` with a keep that matches no field already flags every set field. | A `value_citation_id` column on `observations` — because it would always equal `citation_id`. The cache column in D-12 is different: the cache stores the reconciled value, not the evidence row. |
| D-3 | `likeness` is new catalogs only. Install seeds it; the migration does not; open does not heal. | User, 2026-10-16. A later one-off CLI updates the dogfood projects. `add-catalog-migration` "Do not" forbids seeds in SQL. | Seeding it in `000052.sql` — user declined. |
| D-4 | A depiction Observation does not require a region, a page, or a file on the Artifact. | A single portrait is the whole frame. The Citation can move to another Artifact later, so a file check at insert would not hold. S10-33 and S10-35 handle an Artifact they cannot draw. | Requiring a `region` selector — because "the whole frame can be a depiction" ([depictions and likenesses](../depictions-and-likenesses.md)). |
| D-5 | Polarity is not special-cased. A negative likeness saves like any negative Observation and, through D-7, denies the same Citation on that handle. | "This photo is not Ann" is evidence, and the pipeline already applies it. | Refusing negative depictions — it would discard that evidence. |
| D-6 | Researcher Create refuses `depiction`, like `term`, with `ErrInvalid`. The rule is named once: `func registryOnly(valueType string) bool` *(new)*, unexported, beside `valueTypeOK`, true for `ValueTypeTerm` and `ValueTypeDepiction`; `Create` calls it in place of the inline `term` check. Swift `researcherCreatable` keeps the same five types. | Place, event, and signature keys wait (map Out). A user depiction Property would have no composer row and no gallery. | Allowing it — the user chose new-catalogs-only and a CLI instead (2026-10-16). |
| D-7 | `depiction` is reconciled by a key module: `moduleFor(ValueTypeDepiction)` returns `keyModule{keyOf: citationKey}`, where `citationKey(v)` *(new)* returns `string(v.CitationID), len(v.CitationID) > 0`. `autoreconcile.Value` gains `CitationID []byte` *(new)*, commented "the depiction's Citation; the loader sets it from the Observation". `knownValueType` and `carries` accept depiction (`carries`: `len(v.CitationID) > 0`). `SortKey` returns no key for it, so rows order by rank. Folding, majority, weak, denial, and provisional rules are the shared pipeline's; nothing in `pipeline.go` changes. | Every value type has a module (`docs/conclusion-reconciliation.md` §1, `TestModuleFor`). A gallery is cardinality `multiple`, which keeps every distinct value. Two Observations of one Citation are one picture. A negative or provisional member means what it means for a name. | Skipping depictions in the loader (this plan's first shape) — because negatives, provisional members, and evidence strength would then be re-implemented by S10-35, and the cache invariant would have a hole. Keying by File plus locator — because S10-33's crop cache owns that key, and two Citations of one region on one Source are rare. |
| D-12 | The cache stores the value in `auto_reconciler_values.value_citation_id BLOB` *(new)*, no FK, added by `ALTER TABLE … ADD COLUMN` in `000052.sql`. `sqlLoadCandidates` selects `c.id`; `load` sets `Value.CitationID` only when the group's value type is `depiction`; `hasValue` accepts it; `insertRow` and `sqlInsert` write it. `CacheVersion` becomes 15, with the comment line `15: depiction values reconcile by Citation (S10-32).` | Same shape as `value_entity_id`, which the loader also fills from a join. No FK, like every other cache value column; the cache is `BucketSkip` and rebuilt by upkeep. The bump follows the `CacheVersion` rule. | Storing the Observation id and joining back — because a folded value has several Observations. |
| D-13 | The graph cache mirrors the column: `graphcache.Value.CitationID` and `graphcache.TruthRow.CitationID` *(new)*; `loadValues` selects `r.value_citation_id`; `truthRow` copies `v.CitationID`; `valueEqual` compares it. No consumer reads it in this PR. | `Verify` compares cache to truth field by field; an unmirrored column could drift unseen. | Leaving the graph cache alone — because `Verify` would not cover the new value. |
| D-14 | `conclusiondetails` skips depiction fields the way it skips subject fields: the loop's skip condition also matches `properties.ValueTypeDepiction`. | The gallery is S10-35's own surface; a field row would need a Swift value case no board draws. The skip sits in the display read, not in reconciliation. | A Swift `CatalogConclusionValue.depiction` case — S10-35 draws the gallery instead. |
| D-8 | Promote alignment is not changed. T-6 pins today's behavior: no likeness exhibit. | `pairExhibits` has no comparer for depiction (`match.ComparerFor` is nil), and the exhibit loader reads no Citation into `match.Value`, so no pair is chosen. S10-36 adds the crop comparison. | Adding a `depiction` skip in `pairExhibits` — because the code path already skips it. |
| D-9 | Seed row: `{Key: "likeness", Label: "Likeness", Description: "A picture of this person. The value is the Citation: its Artifact, page, and region. More than one is allowed.", ValueType: properties.ValueTypeDepiction, Cardinality: properties.CardinalityMultiple}`, after `interests`. Binding `{TypeKey: "person", PropertyKey: "likeness", SortOrder: 4, Locked: true}`. | Locked because S10-35's gallery and S10-39's pointer key into that slot, like `name` and `toponym`. | Unlocked like `interests` — because unbinding it would orphan the gallery. |
| D-10 | Swift: add `case depiction` to `PropertyValueType` after `term`. `SubjectPropertyValueType.label(_ valueType: String)` resolves `PropertyValueType(rawValue:)` and switches over the enum with no `default`, so a future case fails to compile until it has a label; `.depiction` returns `L10n.Properties.valueTypeDepiction` (key `properties.valueType.depiction`, value *Depiction*); an unknown raw string still returns `L10n.Properties.valueTypeText`. `researcherCreatable` becomes `[PropertyValueType]` `[.text, .integer, .date, .name, .subject]`; `canSubmit` and `submitCreate` test `researcherCreatable.contains { $0.rawValue == draft.valueType }`, and the chips iterate `researcherCreatable.map(\.rawValue)`. The doc comment says `term` and `depiction` are registry-only. `valueTypeForeground` is not changed. | Without the label VoiceOver says *Text*. `.cursor/rules/macos-client.mdc` says views use the typed `PropertyValueType` cases. A pill colour is a design call with no brief. | Adding one more string `case` — because the string `default` is what mislabels a new type. A new colour case — no board draws it. |
| D-11 | The `TestSwiftVocabularyMatchesGo` "value types" want list adds `properties.ValueTypeDepiction`. | The Swift enum and Go must agree. | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000052.sql` *(new)* | migration | `core/database/migrations/000024.sql` | Extends that shape. A shipped migration is never edited. |
| `properties.ValueTypeDepiction` *(new)* | const | `properties.ValueTypeTerm` | Extends the same constant block and `valueTypeOK`. |
| `properties.registryOnly` *(new)* | func | the inline `valueType == ValueTypeTerm` check in `properties.Create` | Replaces it, so the next registry-only type is one entry. |
| `case properties.ValueTypeDepiction` in `resolveValue` *(new)* | switch case | `case properties.ValueTypeSubject` | Extends the one switch every write uses. |
| `citationKey` and the `moduleFor` depiction case *(new)* | module | `keyModule{keyOf: termKey}` | A key module, like term; no new module type. |
| `autoreconcile.Value.CitationID` *(new)* | field | `Value.EntityID` (loader-filled) | Same pattern: the loader fills it from a join. |
| `auto_reconciler_values.value_citation_id` *(new)* | cache column | `value_entity_id` | One column per value kind, as the table already does. |
| `graphcache.Value.CitationID`, `TruthRow.CitationID` *(new)* | fields | `EntityID` on both | Mirror, so `Verify` covers it. |
| `likeness` row and binding *(new)* | seed rows | `interests` in `seedProperties` / `seedBindings` | Registry entries. No new Install code. |
| `PropertyValueType.depiction` *(new)* | Swift case | `PropertyValueType.term` | Extends the enum the parity test reads. |
| `L10n.Properties.valueTypeDepiction` *(new)* | L10n key | `L10n.Properties.valueTypeTerm` | One key per value type, same namespace. |
| Typed `label` / `researcherCreatable` | Swift refactor in place | `SubjectPropertyValueType` string switch | Same type, read through `PropertyValueType`; no new type. |
| Tests T-1 through T-12 | subtests / test funcs / table cases | `TestObservations`, `TestModules`, `TestMultipleCardinalityKeepsDistinctValues`, `TestInstallSeedsEngagementTravelInterestsAndGender`, `TestForEntityFieldsFollowTheRecords` | Each extends the suite and fixture that already covers that layer. No new helper. |

**Builds on.** The auto-reconciler's one pipeline and `keyModule` (`docs/conclusion-reconciliation.md` §5, §7); `add-catalog-migration` rebuild pattern (`000024.sql`) and cache `ADD COLUMN` (`000038.sql`); `add-seeded-vocabulary` registry (`subjectvocab` `seedProperties` / `seedBindings`); the registry-only refusal in `properties.Create`; `TestSwiftVocabularyMatchesGo` parity.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-33 selects Observations whose Property is `value_type = depiction` to mint crops. S10-34 adds `PropertyValueType.depiction.rawValue` to `CitationComposerVocabulary.supportedValueTypes` and a `case` in `CitationObservationValue`. S10-35 reads the handle's `likeness` rows with reason `kept` or `weak` from `auto_reconciler_values` and maps `value_citation_id` to a crop; it adds its own gallery read beside D-14's skip. S10-36 gives `pairExhibits` a depiction case. Place and event keys are one `seedProperties` row and one `seedBindings` row each.

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

-- The reconciler cache stores a depiction value as its Citation.
ALTER TABLE auto_reconciler_values ADD COLUMN value_citation_id BLOB;
```

`properties` has no secondary index to restore (the only indexes are its PK and `UNIQUE (key, origin)`). `catalogmodel` is unchanged: same tables, and the new cache column has no FK (`add-catalog-model` registers FKs, not columns). Delete class and audit: unchanged. The schema hash is derived from the migrations; do not touch it.

### Go core

| Package | Change |
| --- | --- |
| `core/database/properties` | `ValueTypeDepiction = "depiction"` *(new)*. `valueTypeOK` accepts it. `registryOnly` *(new)* per D-6; `Create` refuses when it is true, with `ErrInvalid`; its doc comment names both types. |
| `core/database/observations` | `resolveValue`: `case properties.ValueTypeDepiction` per D-2. No change to `insertOne`, `updateOne`, or the SQL. |
| `core/autoreconcile` | `Value.CitationID`; `knownValueType`, `carries`, `moduleFor` per D-7; `citationKey` *(new)* in `modules.go` beside `termKey`. |
| `core/database/autoreconciler` | `sqlLoadCandidates` selects `c.id` after `o.value_subject_id`; `load`, `hasValue`, `insertRow`, `sqlInsert`, `CacheVersion` per D-12; `truthRow` per D-13. |
| `core/database/graphcache` | `Value.CitationID`, `TruthRow.CitationID`, `loadValues`, `valueEqual` per D-13. |
| `core/database/conclusiondetails` | Field loop skip per D-14. |
| `core/database/subjectvocab` | `seedProperties` and `seedBindings` rows per D-9. |

Errors: `observations.invalid` → a depiction Observation sets a value field. `properties.invalid` → researcher Create with `depiction`.

### FFI

None — value types travel as strings, and a draft with no value fields already maps to an empty `observations.Input` (§4). No proto change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Value-type enum | `PropertyValueType` | `macos/App/Platform/InterpretationValueKinds.swift` | `case depiction` after `term`. |
| Value-type label and creatable list | `SubjectPropertyValueType` | `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/Properties/PropertiesView.swift` | Per D-10. |

States: none new. The Properties page row and pill already exist; the pill shows the raw key `depiction`. Accessibility: the pill's `accessibilityLabel` reads `SubjectPropertyValueType.label`, so it says *Depiction*.

### Strings

L10n key *(new)*: `L10n.Properties.valueTypeDepiction` = "Depiction" (`properties.valueType.depiction`, comment "Property value type label: depiction (registry-only)"), beside `valueTypeTerm`, with its `Localizable.xcstrings` entry: copy the `properties.valueType.term` block (`extractionState` `manual`, `en` `stringUnit` `translated`) and insert it after `properties.valueType.date`. Removed: none.

## 7. Steps

**STEP-1 — Widen the value_type CHECK and add the cache column**
Files: `core/database/migrations/000052.sql` *(new)*. Skill: `add-catalog-migration`.
Do: write the migration in §6 Schema exactly. No test for the migration itself.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `TestPragmaHonesty`).

**STEP-2 — Add the constant and the Create refusal**
Files: `core/database/properties/properties.go`, `core/database/properties/properties_test.go`.
Do: D-6 and §6 Go core row 1. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/...` passes.

**STEP-3 — Seed `likeness`**
Files: `core/database/subjectvocab/registry.go`, `core/database/subjectvocab/subjectvocab_test.go`, `api/ffi/handlers/subject_defs_test.go`. Skill: `add-seeded-vocabulary`.
Do: D-9. Add T-7. In T-8 change `len(resp.Properties) != 21` to 22, and add a `foundLikeness` flag beside `foundInterests`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/... ./api/ffi/handlers/...` passes.

**STEP-4 — Write a depiction Observation**
Files: `core/database/observations/observations.go`, `core/database/observations/observations_test.go`.
Do: D-2. Add T-2, T-3, and T-4 as subtests of `TestObservations`. `mustSeed` already runs `subjectvocab.Install`, so read `likeness` with `properties.Lookup(c, "likeness", properties.OriginProvenencia)`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/... ./core/database/citations/...` passes.

**STEP-5 — A depiction module in the auto-reconciler**
Files: `core/autoreconcile/autoreconcile.go`, `core/autoreconcile/modules.go`, `core/autoreconcile/modules_test.go`.
Do: D-7. Add T-10. `TestModules` runs single cardinality, so its depiction cases check folding and empties; T-5 checks `multiple`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/...` passes.

**STEP-6 — Cache depiction values and keep them off the Person field list**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/truth.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/graphcache/graph.go`, `core/database/graphcache/load.go`, `core/database/graphcache/verify.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-12, D-13, D-14. In the autoreconciler test, add `"likeness"` to `newFixture`'s `for _, k := range []string{"name", "sex_at_birth", "event_type", "date", "toponym"}` loop, and add `ValueCitationID []byte` to `row` and `value_citation_id` to `snapshot`'s SELECT and Scan, so `assertUpkeepEqualsRebuild` compares it. Add T-5, T-12, and T-6 (D-8). T-11 needs no edit.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes, including every `Verify` caller.

**STEP-7 — Swift value type and label**
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/PropertiesModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-10 and D-11. Change `PropertiesModelTests.swift`'s `#expect(!(SubjectPropertyValueType.researcherCreatable.contains("term")))` to `.contains(.term)`. Add T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes (it reads the Swift enum); `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test` (ProvenenciaTests) passes. `xcodebuild` runs on macOS only: on another host, report the Xcode check and T-9 as not run, and the PR description says so.

**STEP-8 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/conclusion-reconciliation.md`, `docs/conclusion-layer-data-model.md`, `docs/seeded-vocabulary.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `interpretation-layer-data-model.md`: add `'depiction'` after `'term'` in the `properties` DDL sketch. The product value-type sentence lists `depiction` last. Below the `term` paragraph add: "`value_type = 'depiction'` means the Observation's value is its own Citation (`citation_id`): the Artifact, page, and region. No value column is set. It is the one exception to the value population rule below. `likeness` on `person` is the seeded depiction Property." In "Who may create", name `term` and `depiction` as registry-only. After "Value population (application invariant): exactly one value representation must be populated" add "(none for `depiction`, whose value is `citation_id`)".
- `conclusion-reconciliation.md`: add a last row to the §7 table: Value type **depiction**; Same value "the same Citation"; Fold "none"; Merge "the Citation"; No evidence "no Citation". Below the table add: "A depiction's value is its Observation's Citation; the loader sets it, and the cache stores it as `value_citation_id`. `likeness` is `multiple`, so a person's crops form a gallery; the gallery shows `kept` and `weak` values." In the §12 table add a row: Piece "Depictions by Citation, cached as `value_citation_id`"; Where "`core/autoreconcile`, `core/database/autoreconciler`".
- `conclusion-layer-data-model.md`: after the sentence that begins "Exactly one value representation must be populated" add "No Reconciliation Claim is written for a `depiction` Property in Spike 10; the auto-reconciler cache stores a depiction as `value_citation_id`."
- `seeded-vocabulary.md` §3.2: after the `interests` line add a line: key `likeness`, value_type `depiction`, note "the Citation is the picture; cardinality multiple; locked on person; new catalogs only (S10-32)", aligned to the block's columns. §3.3: after `person  interests` add `person          likeness            # depiction; multiple; locked`.
- `.agents/skills/add-seeded-vocabulary/SKILL.md`: the value-type line adds **`depiction`** after **`term`**, and the next sentence reads "**`value_type = term` and `depiction` Properties are Install/registry only**".
- `deployment-plan.md`: the S10-32 **Plan:** line reads "— Landed #N"; the SLICE 8 line `S10-32  depiction value type + likeness on person` gains "— done, #N".
- `completed.md`: a `## S10-32 — Depiction value type` section at the top, in the S10-50 shape: "**Landed** (date) in (PR link)." then two or three sentences of what shipped.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/properties/properties_test.go` `TestProperties` subtest "depiction is registry-only" *(new)*, table-driven over `term` and `depiction` | `Upsert` `{Key:"portrait", Origin:"provenencia", ValueType:"depiction"}` → ok; `Create(tx, userID, "Portrait", vt, "", "")` → `ErrInvalid` for both | AC-3.4, D-6 |
| T-2 | Go | `core/database/observations/observations_test.go` `TestObservations` subtest "likeness saves with no value and no transcription" *(new)* | `mustSeed`'s `seed.artifact` and `seed.person`; `citations.CreateWithObservations` through `writes.Run` with the AC-1.1 locator, transcription `""`, one likeness Input with no value → saved; raw `SELECT` on `c.DB()`: citation `transcription IS NULL`, the six value columns NULL; `SELECT COUNT(*) FROM files` and `FROM file_derivatives` equal before and after | AC-1.1, AC-1.2, D-4 |
| T-3 | Go | same file, subtest "likeness refuses a value" *(new)*, table-driven | `HasText` "seated man, left" → `ErrInvalid`; `ValueTermID` `seed.femaleTerm` → `ErrInvalid`; `SELECT COUNT(*)` of `citations` and `observations` equal before and after | AC-1.3, AC-1.4, D-2 |
| T-4 | Go | same file, subtest "likeness update keeps no value" *(new)* | update the T-2 Observation with no value fields → saved; value columns NULL | AC-1.5 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestDepictionsReconcileByCitation` *(new)*, subtests per AC-2.1 to AC-2.5 | `newFixture`; `album := f.newSource("Album")`; `a := f.subject("person")`, `b := f.subjectOn(album, "person")`; a likeness input is `observations.Input{SubjectID: s.ID, PropertyID: f.props["likeness"].ID}`, cited with `f.cite`, which puts every input of one call on one Citation and returns the saved Observations (read `CitationID` from them); `h := f.promote(a)`, `f.join(b, h)`. AC-2.2: `f.cite(likeness(a), likeness(a2))` with `a2 := f.subject("person")` joined. AC-2.3: `f.cite(likeness(b), negative(likeness(c)))` with `c := f.subjectOn(album, "person")`, filed through `promote.Save` with `ConfidenceGradeID: f.grade("high_confidence")` as `TestReconciledEvidence` does for `low_confidence`. AC-2.4: file B with `identityclaims.Create(… Status: identityclaims.StatusProvisional)` as the claim-status cases in `TestReconciledEvidence` do. AC-2.5: `f.credibility(album, "low_trust")`. Each subtest checks `rowsFor(f.rows(), h, f.props["likeness"].ID)` (reasons, `Support`, `ValueCitationID`) and `f.outcomes()` as its AC says, then `f.assertUpkeepEqualsRebuild` and `must(t, f.c.Graph().Verify())`. A local `likeness(s)` helper builds the input | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-5, D-7, D-12, D-13 |
| T-6 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeShowsNoLikenessExhibit` *(new)* | `aPerson := f.person(f.source, f.artifact, "Gracie")` plus `f.cite(f.artifact, aPerson, observations.Input{SubjectID: aPerson.ID, PropertyID: f.prop("likeness").ID})`; promote it; the same on `f.newSource("Album")`; `f.propose(srcB.ID, nil)` → `rowFor(got, bPerson.ID).Exhibits` has a `name` line and none with `Property.Key == "likeness"`. A comment says S10-36 replaces this assertion with the crop comparison | AC-2.8, D-8 |
| T-7 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestInstallSeedsLikeness` *(new)* | `Install` → `likeness` is depiction, multiple, label *Likeness*; person binding sort 4; `LockedBinding("person","likeness")` | AC-3.1, D-9 |
| T-8 | Go | `api/ffi/handlers/subject_defs_test.go` `TestListProperties` | new catalog → 22 Properties; `likeness` depiction multiple | AC-3.2 |
| T-10 | Go | `core/autoreconcile/modules_test.go` `TestModuleFor` and `TestModules` | `TestModuleFor`'s list gains `properties.ValueTypeDepiction` → not nil. `TestModules` cases (add a `depiction(id, citation byte)` helper beside `term`): "one Citation is one value" (two candidates, same `CitationID`) → one value of both, `StateMerged`; "two Citations" → two values, `StateMixed`; "no Citation is no evidence" → `StateEmpty` | AC-2.1, AC-2.2, D-7 |
| T-11 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestEnsureCatalogRebuildsStaleVersion` | existing test; it reads `CacheVersion`, so 15 passes unedited → the stored version after open equals `CacheVersion` | AC-2.9, D-12 |
| T-12 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityFieldsFollowTheRecords` subtest "no likeness field" *(new)* | a promoted Person with a `name` and a likeness → fields list `name` and no `likeness` | AC-2.7, D-14 |
| T-9 | Swift | `macos/ProvenenciaTests/PropertiesModelTests.swift` `depictionLabelIsNotCreatable` *(new)* | `L10n.string(SubjectPropertyValueType.label("depiction")) == "Depiction"`; `L10n.string(SubjectPropertyValueType.label("unknown")) == "Text"`; `!SubjectPropertyValueType.researcherCreatable.contains(.depiction)` | AC-3.5, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSwiftVocabularyMatchesGo` (until D-11), every `Verify` caller (until D-13), `TestModuleFor`, `TestListProperties` (count), `catalogmodel.TestPragmaHonesty` and the schema hash (STEP-1), `TestRebuildEqualsUpkeep_Scenarios` and `TestRebuildEqualsUpkeep_SeededSequences`.

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
- [ ] Domain doc updated: interpretation-layer-data-model.md (value types, value population), conclusion-reconciliation.md §7, §12, conclusion-layer-data-model.md, seeded-vocabulary.md §3.2–3.3, add-seeded-vocabulary skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Create a new project. Open **Properties**, select **Person**. A **Likeness** row is listed; its value-type pill shows `depiction` and VoiceOver reads it as *Depiction*, it allows more than one value, and its binding cannot be removed. (AC-3.1, AC-3.5)
2. Click **Create property**. The value-type chips are text, integer, date, name, and subject, and not depiction. (AC-3.5)
3. Open a Source with a Person, open the composer, add an Observation on the Person. The Property menu does not list Likeness. (S10-34 adds it.)
4. Open a project created before this build. It opens. **Properties → Person** has no Likeness row. (AC-3.3)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rebuild drops a column or default added since `000043.sql` | §4 lists the full column set; T-1 and every `properties` test read `cardinality` after the migration. |
| Another reader errors on an Observation with no value column | §4 lists every reader that branches on value type; only `autoreconcile.Reconcile` errors, and D-7 gives it a module. |
| A cache reader picks up `likeness` rows | §4 "Cache readers": each filters by key, profile, or `value_entity_id`. |
| A likeness Observation on the graph card shows an empty value | None can be made in the app before S10-34 (manual check 3). |

**Stop and report if:**

- A migration after `000043.sql` changed the `properties` DDL, so the §6 sketch would drop a column.
- `TestPragmaHonesty` or the schema-hash check fails after STEP-1 and the fix needs a change outside `000052.sql`.
- Any test fails with `ErrUnknownValueType` or `observations.invalid` for a likeness row after STEP-6, which means a reader §4 does not list branches on value type.
- A `match.Rank`, search, or display test changes result after STEP-6 because a `likeness` cache row was read, which proves §4 "Cache readers" wrong.
- AC-2.3's negative does not deny, which means the pipeline's denial needs more than an equal key and proves D-7 incomplete.
- `pairExhibits` emits a likeness line in T-6, which proves D-8 wrong.
- `xcodebuild test` fails on a `switch` over `PropertyValueType` that is not exhaustive, which means a Swift switch §4 does not list.

## 12. Open questions

None.

## 13. PR

**Title:** S10-32 — Depiction value type and likeness on Person
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-32-depiction-value-type.md`.
