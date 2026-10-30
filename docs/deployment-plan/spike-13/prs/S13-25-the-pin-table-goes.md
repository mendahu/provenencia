# S13-25 — The pin table goes

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 0 — Removing Observation pins
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the delete confirm's "removed from the evidence for …" line can show only for a pinned Observation, and the migration leaves none; every other confirm reads the same)
**Board:** —
**Depends on:** S13-24 (landed [#639](https://github.com/mendahu/provenencia/pull/639))
**Blocks:** S13-08
**Branch:** `s13-25-the-pin-table-goes`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut; the catalog format step is its own number ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md) (removing a vocabulary), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md) (removing one key), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-l10n.mdc`
**Context:** [conclusion §4, §5, §5.3](../../../conclusion-layer-data-model.md#5-membership-evidence), [`catalog-deletes.md`](../../../catalog-deletes.md) item 9, [`seeded-vocabulary.md`](../../../seeded-vocabulary.md) §5.5; working rules: "Pins leave in three PRs, the client first", "Identity Claims carry no confidence and no argument", "Audit history keeps its pin rows", "Each removal PR updates what it makes false"
**Planned:** 2026-10-30 against `main` @ `588e4a82`
**Review:** [`S13-25-the-pin-table-goes.review.md`](S13-25-the-pin-table-goes.review.md) — Fits with changes / Ready with changes, 2 rounds; every blocker Fixed; no To user item

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

The last of the three pin-removal PRs. S13-23 stopped the app sending pins, confidence and argument; S13-24 deleted the engine that wrote them, so nothing writes `identity_claim_evidence`, `identity_claims.confidence_grade_id` or `identity_claims.argument` any more. This PR drops them from the catalog in one format step, with the `claim_confidence_grades` table and its seed, and deletes the code that still reads them: the pin delete facet and its Impact line, `identityclaims` pins, the effects entry, the grades vocabulary package, and the Mac app's pin confirm line. A project made before the PR opens on the new format with its claims and their statuses intact, and an Identity Claim is membership only: subject, entity, status.

**Why now.** Slice 0 runs in order and S13-24 landed; S13-08 waits on this PR because both change `deleteimpact.ReleaseClaim`.

**What stays.** `identity_claims` with `id`, `subject_id`, `entity_id`, `subject_type_id`, `status`, its CHECK, composite FKs, unique pair, and its two indexes; `canonical_entities.argument`; every `audit_changes` and `audit_transaction_scopes` row already written for `identity_claim_evidence` or carrying the two claim keys (working rule "Audit history keeps its pin rows"); `ReleaseClaim` as the one claim removal; `KindIdentityClaim` (`graphcache` names the claims table by it); `source_credibility_grades` and its seed.

## 2. User stories

**US-1.** As a researcher, when I open a project made before this PR, it opens on the new format with every claim kept, so removing pins costs me nothing I filed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A project made at the planning base with one handle, two accepted members, one rejected claim, four pins, and one claim with `confidence_grade_id` = `low_confidence` and `argument` = `same age`; open it with the PR's build | It opens with no error. `PRAGMA table_info(identity_claims)` lists `id, subject_id, entity_id, subject_type_id, status`; `sqlite_schema` has no `identity_claim_evidence`, no `claim_confidence_grades`, no `identity_claims_confidence_grade_id_idx`; `SELECT id, subject_id, entity_id, status FROM identity_claims` returns the same three rows as before |
| AC-1.2 | The same project | `SELECT COUNT(*) FROM audit_changes WHERE entity_type = 'identity_claim_evidence'` returns the same count before and after; `audit_transaction_scopes` row count is unchanged |
| AC-1.3 | A new project created with the PR's build | It has no `claim_confidence_grades` table, and its create installs no claim grade |

**US-2.** As a researcher, when I delete a promoted member's records, the confirm and the audit name only what is really removed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Impact of an Observation of a promoted member | `Allowed`, gate OK, no Groups, no Cascades; the confirm names no handle |
| AC-2.2 | Delete that Observation | The revision `delete_observation` audits `observation` only; the member's claim stays as written |
| AC-2.3 | Delete the now-empty joined member B of a handle A grounds | Impact names the handle it leaves; the revision `delete_subject` audits `identity_claim,subject`; A's claim is the handle's only accepted member; the auto-reconciler cache equals a rebuild |
| AC-2.4 | `ReleaseFacets(KindSubject, B)` on that handle's member B | Changes are one `identity_claim`; `effects.Handles` on them is the one handle |

**US-3.** As a contributor, the code and guidance hold no trace of pins or claim grades.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg -n 'identity_claim_evidence\|claim_confidence_grade\|claimconfidencegrades\|confidence_grade_id\|ClaimConfidence\|PinTx\|PinnedObservations\|pinRelease\|claimEvidenceGraph\|pinsByObservation\|leavesEvidence\|pinnedObservation' core api macos/App macos/ProvenenciaTests -g '!core/database/migrations/**' -g '!**/Generated/**' -g '!api/proto/**'` | Hits only in `core/database/migrations/` (history), none elsewhere |
| AC-3.2 | The same names over `.agents .cursor docs -g '!docs/deployment-plan/**' -g '!docs/**/archive/**'` | Hits only where a doc says they were dropped (`dropped by S13-25`) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration `000058.sql` drops `identity_claim_evidence`, `identity_claims.confidence_grade_id` (with its FK and index) and `.argument`, and `claim_confidence_grades` (D-1) | STEP-1, AC-1.1, AC-1.2 |
| `catalogmodel` loses the two tables, three FKs and `KindClaimConfidenceGrade` (D-2) | STEP-1 |
| `deleteimpact` loses the two pin facets, `pinRelease`, the grade edge, and `releaseClaim`'s two columns (D-3) | STEP-2, AC-2.1 – AC-2.4 |
| `core/database/identityclaims/pins.go` goes (D-6) | STEP-2 |
| `effects` loses the `identity_claim_evidence` entry, `claimEvidenceGraph` and the `claim_confidence_grades` entry (D-4) | STEP-3 |
| `core/database/claimconfidencegrades` goes, with its create-time `Install`, its seed test, its `Install` in three test fixtures, and `apperr.CodeClaimConfidenceInvalid` (D-5) | STEP-4, AC-1.3 |
| Tests that build pins are rewritten or deleted (D-7) | STEP-2, STEP-4 |
| Swift: `PVDeleteImpact`'s pin case, `L10n.DeleteImpact.leavesEvidence` and its key, the `pinnedObservation` preview, `FakeStore.pinsByObservation` and its cascade, and two Swift tests (D-8) | STEP-5 |
| Comments that name pins (D-10) | STEP-2 – STEP-4 |
| Docs and guidance the removal makes false (D-9) | STEP-6, AC-3.2 |

| Out | Home |
| --- | --- |
| A Reconciliation Claim exhibit ([conclusion §13](../../../conclusion-layer-data-model.md#13-open-schema-questions) item 5) | not planned |
| `canonical_entities.argument` | stays (working rule "Identity Claims carry no confidence and no argument") |
| Where the reasoning that records belong together lives | not decided (working rule) |
| Reconciliation Claim confidence vocabulary | open ([`research-judgment-model.md`](../../../research-judgment-model.md) §4) |
| Unpromote of the last member deleting the person, an unpromote Impact preflight | S13-08 |
| Rewriting or deleting old audit rows for pins or the two claim keys; a retired-entity-type reader | not planned (working rule "Audit history keeps its pin rows") |
| One seeded test catalog helper replacing the hand-listed vocabulary installs in test fixtures (review M-3) | not planned (review log) |
| A typed delete-impact via with a `api/ffi/swift_vocabulary_test.go` parity row for `PVDeleteImpactCopy.cascadeLines` (review M-4) | S13-08's plan, which adds an unpromote Impact preflight |
| Provisional claims at Promote | [`docs/ideas/provisional-identity-claims.md`](../../../ideas/provisional-identity-claims.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Schema | `core/database/migrations/000032.sql` | Creates `claim_confidence_grades`, `identity_claims` with `confidence_grade_id BLOB REFERENCES claim_confidence_grades(id)` and `argument TEXT`, `identity_claim_evidence`, and indexes `identity_claims_one_accepted_per_subject`, `identity_claims_entity_id_idx`, `identity_claims_confidence_grade_id_idx`, `identity_claim_evidence_observation_id_idx` |
| Schema | `core/database/migrations/000033.sql` | Rebuilds `identity_claim_evidence` so `observation_id` has no CASCADE. The latest step is `000057.sql` |
| Catalog model | `core/database/catalogmodel/model.go` `KindClaimConfidenceGrade`, tables (lines 93, 118), FKs (lines 196 – 198) | Registers `claim_confidence_grades` (vocab), `identity_claim_evidence` (facet), FK `identity_claims.confidence_grade_id` (resource, NO ACTION), and the pin table's two audited facet FKs. The `KindIdentityClaim` comment says it "parents facet releases (its pins) only" |
| Delete facets | `core/database/deleteimpact/facets.go` `buildFacetReleases` | Two entries: Parent `KindObservation` via `identity_claim_evidence.observation_id` (Named, Child canonical entity: the Impact cascade line), and Parent `KindIdentityClaim` via `identity_claim_evidence.identity_claim_id`; both release through `pinRelease`. The Subject entry's comment says "Claims' own pins go first" |
| Claim release | `core/database/deleteimpact/facets.go` `releaseClaim`, `ReleaseClaim` | Selects `confidence_grade_id, argument` with the row, calls `releaseFacets(tx, KindIdentityClaim, id)` (today only the pin facet), deletes the row, and audits a delete change whose `DeletedRow` carries `confidence_grade_id` and `argument` |
| Grade edge | `core/database/deleteimpact/register.go` `joinEdge(catalogmodel.KindClaimConfidenceGrade, "identity_claims.confidence_grade_id", …)` | Inbound probe from a grade to the handles whose claims use it |
| Pins | `core/database/identityclaims/pins.go` `PinTx`, `PinnedObservations` | Writes and reads pins. No non-test caller since S13-24 |
| Effects | `core/database/effects/registry.go` `claimEvidenceGraph`, `"identity_claim_evidence"`, `"claim_confidence_grades"` | The pin table's handle, handle-scope and Evidence-graph paths, and the grade table's `None` entry. The registry init refuses an entry for a table `catalogmodel` does not declare |
| Effects comment | `core/database/effects/scopes.go` `ScopeCanonicalEntity` | Says the scope covers "its claims and pins" |
| Grades vocabulary | `core/database/claimconfidencegrades/` (`grades.go`, `registry.go`, `grades_test.go`) | `Upsert`, `Lookup`, `GetByID`, `List`, `Install`, `ErrInvalid` (`apperr.CodeClaimConfidenceInvalid`); `Install` seeds `low_confidence`, `moderate`, `high_confidence` |
| Create path | `core/onboarding/ready.go` `createCatalog`, `OpenCatalog` | `createCatalog` calls `claimconfidencegrades.Install`; both doc comments name claim confidence grades |
| Error code | `core/apperr/apperr.go` `CodeClaimConfidenceInvalid` | Only `claimconfidencegrades.ErrInvalid` uses it; the Mac app maps no `claimconfidence.*` code |
| Comments | `core/database/promote/unpromote.go` `Unpromote`, `core/database/identityclaims/bridgefile.go` `UnfileBridges`, `core/database/observations/write.go` (line 95), `core/database/deleteimpact/README.md` | Each says pins are released first or stay |
| Swift confirm | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactCopy.cascadeLines`, `PVDeleteImpactPreviewData.pinnedObservation` | A case for via `identity_claim_evidence.observation_id` returns `L10n.DeleteImpact.leavesEvidence`; the preview fixture builds that cascade. Any other via falls to `alsoAffects` |
| Swift copy | `macos/App/Platform/L10n.swift` `L10n.DeleteImpact.leavesEvidence(handleRefs:)`; `macos/App/Resources/Localizable.xcstrings` `deleteImpact.confirm.leavesEvidence` | "It’s also removed from the evidence for %@. Those claims stay, with less evidence." |
| Fake store | `macos/App/Platform/FakeStore.swift` `pinsByObservation`, `deleteObservation`, `observationDeleteImpact` | Holds pins per Observation, clears them on delete, and builds the pin cascade |
| Go tests | `core/database/deleteimpact/pins_e2e_test.go` `TestPinnedObservationDeleteEndToEnd`, `TestPinnedMemberDeleteEndToEnd`, `newPinnedHandle`, `cacheMatchesRebuild`, `joinPinned` | Pin end-to-end tests; `joinPinned` files B onto A's handle with `promote.SaveBatch` and pins with `identityclaims.PinTx` |
| Go tests | `core/database/deleteimpact/conclusion_release_test.go` `newConclusionFixture`, `TestSubjectDeleteReleasesOwnClaimPins` | Fixture: Subjects A and B (Places, one toponym Observation each), A grounds a handle, B joins through `joinPinned` |
| Go tests | `core/database/deleteimpact/facets_test.go` `TestReleaseFacetsReportsHandles` | Releases `KindObservation` for `obsA` and expects two `identity_claim_evidence` changes and one handle |
| Go tests | `core/database/deleteimpact/conclusion_test.go` `TestImpactConclusion` | Subtests "pinned observation is allowed and names the handle its claims weaken" (inserts pins by SQL) and "observation delete removes its pins explicitly and audits them" |
| Go tests | `core/database/promote/batch_test.go` `TestSaveBatchFilesClaimsAndBridgesOnly`; `core/database/promote/bridges_test.go` `TestBridgeSubjectDeleteReleasesPins` | The first counts `identity_claim_evidence` rows (S13-24's no-pin check); the second pins a bridge edge with `PinTx` and checks the bridge delete releases it |
| Go tests | `core/database/effects/handle_paths_test.go` subtest "evidence pin reaches its handle"; `core/database/effects/reads_test.go` `TestReadsDeclare`; `core/database/effects/registry_completeness_test.go` `TestGradeTablesHaveNoHandles`, the Evidence-graph table list (line 112), `TestEvidenceGraphPathsEndAtGraphs` | Each lists `identity_claim_evidence` or `claim_confidence_grades` |
| Go tests | `core/onboarding/ready_test.go` `TestCreateCatalogSeedsClaimConfidenceGrades`; `core/database/autoreconciler/autoreconciler_test.go` `newFixture`, `fixture.grade`, subtest "a claim's grade does not weaken its values"; `core/database/conclusiondetails/conclusiondetails_test.go` `newFixture`; `core/database/gallery/gallery_test.go` fixture (line 65) | Seed test, the S13-24 grade subtest (writes `confidence_grade_id` by SQL), and three fixtures that call `claimconfidencegrades.Install` |
| Swift tests | `macos/ProvenenciaTests/PromoteModelTests.swift` `aPinnedObservationDeleteNamesThePerson`; `macos/ProvenenciaTests/PVDeleteImpactTests.swift` `pinnedObservationConfirmNamesTheEvidenceItLeaves` | Pin confirm tests, left for this PR by S13-23 |

**Precedent.** Table rebuilds with `-- migrate: foreign_keys off`: `000030.sql` (`321bf906`, `artifacts__new` … `DROP TABLE artifacts; ALTER TABLE artifacts__new RENAME TO artifacts;`, then its indexes recreated) and `000052.sql` (`f3dfb629`). Copy the `__new` / copy / drop / rename shape and the index recreation; do not copy their data changes. Removal shape: S13-24 ([#639](https://github.com/mendahu/provenencia/pull/639)), which deleted code and tests whole and swept the names with `rg` in its DoD.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One step, `core/database/migrations/000058.sql` *(new)*, with `-- migrate: foreign_keys off`. In order: `DROP TABLE identity_claim_evidence;` then rebuild `identity_claims` as `identity_claims__new` with `id, subject_id, entity_id, subject_type_id, status`, the same CHECK, the two composite FKs with `ON DELETE CASCADE`, and `UNIQUE (subject_id, entity_id)`; copy those five columns; `DROP TABLE identity_claims; ALTER TABLE identity_claims__new RENAME TO identity_claims;`; recreate `identity_claims_one_accepted_per_subject` and `identity_claims_entity_id_idx` exactly as `000032.sql` writes them; then `DROP TABLE claim_confidence_grades;`. The header comment names S13-25 and says audit rows are kept. No `INSERT`, `UPDATE` or `DELETE` touches an audit table | SQLite's `DROP COLUMN` refuses a column that has an FK or an index; the working rule asks for one format step; dropping the pin table first leaves no FK into `identity_claims`, so the rebuild breaks none (`migrate` runs `foreign_key_check` before commit) | `ALTER TABLE … DROP COLUMN` (fails on `confidence_grade_id`); two steps (the rule says one); deleting audit rows (working rule) |
| D-2 | `catalogmodel`: delete the `claim_confidence_grades` and `identity_claim_evidence` tables, the three FKs at lines 196 – 198, and `KindClaimConfidenceGrade`. `KindIdentityClaim` stays; its comment reads "KindIdentityClaim names the identity_claims table; claims have no ref and no delete of their own (deleteimpact.ReleaseClaim removes one)." | `TestPragmaHonesty` requires the model to match the live schema; `graphcache` still calls `mustTable(KindIdentityClaim)` | Deleting `KindIdentityClaim` — `core/database/graphcache/graph.go` reads it |
| D-3 | `deleteimpact`: delete both pin `facetRelease` entries, `pinRelease`, and the grade `joinEdge` in `register.go`. `releaseClaim` selects `id, subject_id, entity_id, subject_type_id, status`, and its `DeletedRow` has those five keys. It keeps the `releaseFacets(tx, catalogmodel.KindIdentityClaim, id)` call, so a facet later registered under claims is released; its doc comment reads "releaseClaim removes one identity claim, audited: its registered facets (none today), then the row, then the claim's own delete change carrying the fields the effects registry resolves a ghost row from." The rest of the comment stays. The Subject entry's comment drops "Claims' own pins go first." `nullStringJSON` stays (`releaseObservation` uses it) | The facet registry is how a release finds child rows; removing the call would let a future claim facet register and never run | Removing the `releaseFacets` call — a silent gap for the next facet; keeping the two keys in `DeletedRow` as nulls — the columns are gone |
| D-4 | `effects`: delete the `"identity_claim_evidence"` entry, `claimEvidenceGraph`, and the `"claim_confidence_grades"` entry. `ScopeCanonicalEntity`'s comment reads "its row, its claims, and evidence on a member Subject". `TestGradeTablesHaveNoHandles` loops over `source_credibility_grades` only and its comment drops the claim grade sentence; the "evidence pin reaches its handle" subtest goes; `TestReadsDeclare`, the line-112 row and `TestEvidenceGraphPathsEndAtGraphs`'s `want` list (line 186) drop `identity_claim_evidence` | The init (`registry.go`, "entry for unknown table") refuses an entry for a table `catalogmodel` no longer declares (working rule); nothing replays stored audit rows by entity type | A retired-type entry kept for old audit rows (working rule: not built) |
| D-5 | Delete `core/database/claimconfidencegrades/` whole and `apperr.CodeClaimConfidenceInvalid`. `createCatalog` drops the `Install` call; `OpenCatalog`'s comment reads "Does not re-install or heal Source vocabulary, credibility grades, Interpretation subject vocabulary, name formats, or alias tables." Delete `TestCreateCatalogSeedsClaimConfidenceGrades` and the `claimconfidencegrades.Install` lines and imports in the autoreconciler, conclusiondetails and gallery fixtures. Delete the autoreconciler subtest "a claim's grade does not weaken its values" and `fixture.grade` | No caller is left once the table goes; S13-24's subtest needs the dropped column to set a grade, and the reconciler reads no grade since S13-24 | Keeping the package for Reconciliation Claims — their confidence is open (Out); keeping the subtest without a grade — it would test nothing |
| D-6 | Delete `core/database/identityclaims/pins.go` | No non-test caller since S13-24; its table is gone | — |
| D-7 | Tests. Delete `pins_e2e_test.go`; the cache check becomes one helper, `autoreconcilertest.AssertCacheMatchesRebuild(t testing.TB, c *database.Catalog)` *(new)* in `core/database/autoreconciler/autoreconcilertest/autoreconcilertest.go` *(new)*: the `SELECT *` dump of `auto_reconciler_values` and `auto_reconciler_outcomes` ordered by 1, 2, 3, then `autoreconciler.Rebuild` on a rolled-back transaction, then compare (the body of `bridgeWorld.cacheMatchesRebuild` with `cacheDump` from `core/database/promote/bridges_test.go`, taking `c.DB()` inside the helper and failing with its message "upkeep differs from a rebuild"). T-2 calls it. In `core/database/promote`, the four `w.cacheMatchesRebuild()` calls (`bridges_test.go`, `unpromote_test.go`) and the one `assertCacheMatchesRebuild(t, db)` call (`batch_test.go`) call it with the fixture's catalog, and `bridgeWorld.cacheMatchesRebuild`, `cacheDump`, `assertCacheMatchesRebuild` and `cacheLines` are deleted. In `core/database/autoreconciler/autoreconciler_test.go`, `assertUpkeepEqualsRebuild` replaces its values block and its outcomes block with one `autoreconcilertest.AssertCacheMatchesRebuild(f.t, f.c)` call and keeps its search-documents block and its `at` label on that block's message; `rebuiltSnapshot` and `outcomesSnapshot` stay (other tests read them). `newConclusionFixture` joins B with `promotetest.Save(t, c, userID, promote.Input{SubjectID: b.ID, EntityID: grounding.Entity.ID})`, gains `entityRef string` set to `grounding.Entity.Ref`, keeps `testCatalog` and its Source, Artifact and Citation setup, and its comments drop pins. `TestSubjectDeleteReleasesOwnClaimPins` becomes `TestMemberDeleteEndToEnd` *(new)* (T-2). `TestReleaseFacetsReportsHandles` releases `KindSubject` for `f.b` (T-3). In `TestImpactConclusion` the two pin subtests become T-1, and `otherClaimID` (used only to pin across claims) and its assignment in "allowed promoted subject still names the handle it leaves" go. In `TestSaveBatchFilesClaimsAndBridgesOnly` the `identity_claim_evidence` count block goes. `TestBridgeSubjectDeleteReleasesPins` becomes `TestBridgeSubjectDeleteKeepsTheEndClaims` *(new)* (T-4) Imports follow the compiler: an import a deleted helper or subtest leaves unused is removed. `conclusionFixture` drops `cb` and `obsA` once no test reads them (T-2 and T-3 do not). `countRows` in `conclusion_test.go` is deleted: its three callers are pin checks this PR removes; `lastRevision`, `assertLeaves` and `mustImpact` stay. T-1's two subtests replace the old two in place, before "subject delete removes its claim explicitly…". T-4 drops the `person` claim variable, the `edge` loop, the `PinTx` revision, and both audit-type assertions; it keeps the bridge delete and the `james` check | Each keeps the non-pin behavior it covered (member delete, cache upkeep, handle reporting, bridge delete); the pin assertions have no table to read | Deleting all four tests — member and bridge deletes would lose their end-to-end cover |
| D-8 | Swift. `cascadeLines` loses the `identity_claim_evidence.observation_id` case; `L10n.DeleteImpact.leavesEvidence` and key `deleteImpact.confirm.leavesEvidence` go; `PVDeleteImpactPreviewData.pinnedObservation` goes; `FakeStore.pinsByObservation`, its line in `deleteObservation`, and the cascade in `observationDeleteImpact` go (after the not-found and edge-locked gates it returns `CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])`); `aPinnedObservationDeleteNamesThePerson` and `pinnedObservationConfirmNamesTheEvidenceItLeaves` go | Go reports that via no more; an unknown via still reaches `alsoAffects`, which `PVDeleteImpactTests`' existing `alsoAffects` test covers, so no test is added | Keeping the case for old engines — the app and engine ship together |
| D-9 | Docs. The map's S13-25 list (working rule "Each removal PR updates what it makes false"), plus three files the map missed, added to its list in the planning commit: [conclusion §7](../../../conclusion-layer-data-model.md#7-reconciliation_claims) (`reconciliation_claims.confidence_grade_id` has no vocabulary table: "open; `claim_confidence_grades` dropped by S13-25"), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md) (registry row, install order, create diagram), and [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) (Status: "superseded: Identity Claims carry no pins since S13-25; review of how a person's records connect is slice 7") Edits per doc, everything else stays: (a) `conclusion-layer-data-model.md`: §4's `claim_confidence_grades` DDL and its "Seed keys" line are deleted; the `identity_claims` DDL shows the five columns; the confidence paragraph (line 202) and the `argument` paragraph are replaced by one sentence ("An Identity Claim carries no confidence grade and no argument; S13-25 dropped both columns, the pin table and `claim_confidence_grades`."); §3 item 2 reads "a status, with no Observation exhibit"; line 252 reads "the columns were dropped by S13-25"; §5's "planned work" sentence is past tense; §7's `reconciliation_claims.confidence_grade_id` DDL comment and line 407 say its vocabulary is open (`claim_confidence_grades` was dropped by S13-25); §10 invariant 19 reads "Identity Claims carry no confidence; Reconciliation Claim confidence is open."; §12's claim argument examples drop the Identity Claim argument. (b) `seeded-vocabulary.md`: the line-44 row goes; §5.5's body (lines 656 – 667) becomes the one sentence "Dropped by S13-25: Identity Claims carry no confidence; a Reconciliation Claim vocabulary is open." (c) `research-judgment-model.md`: unchanged except line 166's comment, which reads "claim_confidence_grades: dropped by S13-25; kept here as the shape an open Reconciliation Claim vocabulary would start from" (its §4 is the open question, §3 Out). (d) `catalog-deletes.md`: lines 17, 30 ("notes and owned values"), 46 and 83. Hits the sweeps keep, each a decision note or Reconciliation Claim text: `promote-graph-alignment.md` lines 7, 51, 176, 210; `conclusion-reconciliation.md` 145; `research-judgment-model.md` 200; `data-model-source-interpretation-conclusion.md` 144; `conclusion-layer-data-model.md` 139, §5's rationale (212 – 218), 352 – 358, 370, 450 – 452, 490, 536, 697, 708 | Each states pins or claim grades as current | Leaving the idea file — it reads as a plan built on pins |
| D-10 | Comments: `core/database/promote/batch_test.go` line 305 reads "no other member's claim changes."; `Unpromote` drops "each with its pins" and "and their pins"; `UnfileBridges` drops "pins first, "; `core/database/observations/write.go` line 95 reads "Notes are released first; the caller records the returned changes."; `core/database/deleteimpact/README.md` reads "`ReleaseClaim` is the one claim removal (audited)." | They describe code this PR deletes | — |
| D-11 | No `autoreconciler.CacheVersion` bump and no proto change | Since S13-24 no reader selects the two columns (`rg confidence_grade_id core --glob '!*_test.go' --glob '!core/database/migrations/**'` hits only `core/database/catalogmodel/model.go`, `core/database/deleteimpact/facets.go` and `core/database/deleteimpact/register.go`, all edited here), and the wire is already reserved | Bumping — no cached value changes |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000058.sql` *(new)* | file | `000030.sql` rebuild | Format steps are append-only; it follows that shape |
| `autoreconcilertest.AssertCacheMatchesRebuild` *(new)* in `core/database/autoreconciler/autoreconcilertest/autoreconcilertest.go` *(new)* | test helper package | `pinnedHandle.cacheMatchesRebuild`, `bridgeWorld.cacheMatchesRebuild` + `cacheDump`, `assertCacheMatchesRebuild` + `cacheLines` (three copies of one check), and the first two blocks of `fixture.assertUpkeepEqualsRebuild` (a fourth) | It replaces the three; a `*test` package beside its package is the repo's shape (`promotetest`, `graphcachetest`) and every caller is an external `_test` package, so no import cycle |
| `TestMemberDeleteEndToEnd` *(new)* | test | `TestPinnedMemberDeleteEndToEnd` | It is that test without pins, on `newConclusionFixture` |
| `TestBridgeSubjectDeleteKeepsTheEndClaims` *(new)* | test | `TestBridgeSubjectDeleteReleasesPins` | Renamed: the old name describes a pin |

**Builds on.** `add-catalog-migration` rebuild rules; the `catalogmodel` / `deleteimpact` / `effects` registries, whose honesty and completeness tests prove the removal is whole.
**Deletes.** `countRows` (deleteimpact tests); `bridgeWorld.cacheMatchesRebuild`, `cacheDump`, `assertCacheMatchesRebuild`, `cacheLines` (promote tests); `identity_claim_evidence`, `claim_confidence_grades`, two `identity_claims` columns and one index; `core/database/claimconfidencegrades/` (3 files); `core/database/identityclaims/pins.go`; `apperr.CodeClaimConfidenceInvalid`; `catalogmodel.KindClaimConfidenceGrade`; two facet releases, `pinRelease`, one probe edge; `effects` `claimEvidenceGraph` and two entries; `pins_e2e_test.go`, `TestCreateCatalogSeedsClaimConfidenceGrades`, `fixture.grade`, one autoreconciler subtest, one effects subtest; Swift `leavesEvidence` (L10n and catalog key), `pinnedObservation`, `pinsByObservation`, two Swift tests.
**Next adds.** S13-08 changes `ReleaseClaim` with no pin step in the way.

### Reach

None — no new path into existing code. The one new state, a project with no pin table, is the only state after the migration, and every reader of the table goes in the same PR.

### Scale

At the map's Scale (a project whose every accepted claim has pins): the migration does one `DROP TABLE` of the pin table and one copy of `identity_claims` (one row per claim, a few thousand at most), inside the format step's transaction at open. No write, list or loop is added; Observation delete and Subject delete do less work (no pin probe or release).

### Edges

- **Assumptions.** The rebuild assumes no table, trigger or view references `identity_claims` other than `identity_claim_evidence` (checked: `rg 'REFERENCES identity_claims|TRIGGER|VIEW' core/database/migrations` finds only the pin table and a `sources` trigger). `migrate`'s `foreign_key_check` holds the rest; the schema hash check holds the result.
- **Neighbors.** A rejected or provisional claim is copied like an accepted one (AC-1.1 includes a rejected claim). A claim with NULL grade and argument is copied the same as one with both.
- **Removed inputs.** A claim's grade and argument are lost on open; the user's call (working rule, 2026-10-29). Old audit rows keep them as history.
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** None — no list is read from data the app did not write.

### Schema

Migration `core/database/migrations/000058.sql` *(new)*:

```sql
-- S13-25: Identity Claims are membership only. Drops the Observation pins,
-- the claim confidence grade and argument columns, and the grades table.
-- Audit rows already written for pins and for the two columns are kept.
-- migrate: foreign_keys off

DROP TABLE identity_claim_evidence;

CREATE TABLE identity_claims__new (
	id              BLOB PRIMARY KEY,
	subject_id      BLOB NOT NULL,
	entity_id       BLOB NOT NULL,
	subject_type_id BLOB NOT NULL,
	status          TEXT NOT NULL,

	CHECK (status IN ('provisional', 'accepted', 'rejected')),
	FOREIGN KEY (subject_id, subject_type_id)
		REFERENCES subjects (id, subject_type_id)
		ON DELETE CASCADE,
	FOREIGN KEY (entity_id, subject_type_id)
		REFERENCES canonical_entities (id, subject_type_id)
		ON DELETE CASCADE,
	UNIQUE (subject_id, entity_id)
) STRICT;
INSERT INTO identity_claims__new (id, subject_id, entity_id, subject_type_id, status)
SELECT id, subject_id, entity_id, subject_type_id, status FROM identity_claims;
DROP TABLE identity_claims;
ALTER TABLE identity_claims__new RENAME TO identity_claims;

CREATE UNIQUE INDEX identity_claims_one_accepted_per_subject
	ON identity_claims (subject_id)
	WHERE status = 'accepted';
CREATE INDEX identity_claims_entity_id_idx ON identity_claims(entity_id, status);

DROP TABLE claim_confidence_grades;
```

`catalogmodel`: per D-2. Delete class: unchanged for `identity_claims` (facet of Subject and handle). Audit: no new entity type.

### Go core

| Package | Change |
| --- | --- |
| `core/database/catalogmodel` | D-2 |
| `core/database/deleteimpact` | D-3, D-10 (README) |
| `core/database/identityclaims` | D-6, D-10 (`core/database/identityclaims/bridgefile.go`) |
| `core/database/effects` | D-4 |
| `core/database/claimconfidencegrades` | deleted (D-5) |
| `core/onboarding` | D-5 |
| `core/apperr` | D-5 |
| `core/database/promote`, `core/database/observations` | D-10 comments |

Errors: `claimconfidence.invalid` is gone; no code returns it.

### FFI

None — the claim fields and the grade RPC were reserved by S13-24 (D-11).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Delete confirm cascade line | `PVDeleteImpactCopy.cascadeLines` | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` | One `case` removed (D-8) |
| Fake store | `FakeStore` | `macos/App/Platform/FakeStore.swift` | D-8 |

States: unchanged. Accessibility: unchanged.

### Strings

Removed: `L10n.DeleteImpact.leavesEvidence(handleRefs:)`, key `deleteImpact.confirm.leavesEvidence`. No key added.

## 7. Steps

**STEP-1 — The format step and the model**
Files: `core/database/migrations/000058.sql` *(new)*, `core/database/catalogmodel/model.go`, `core/database/deleteimpact/register.go`, `core/database/effects/registry.go`. Skill: `add-catalog-migration`, `add-catalog-model`, `add-audit-scope`.
Do: per D-1 and D-2. In the same step, delete the grade `joinEdge` in `register.go` (it names `KindClaimConfidenceGrade`, D-3) and D-4's three `registry.go` deletions (`"identity_claim_evidence"`, `claimEvidenceGraph`, `"claim_confidence_grades"`): `core/database/listen.go` imports `effects`, whose `init` panics on an entry for a table `catalogmodel` no longer declares, so every `core/database` test dies until they go. Other packages' tests fail until STEP-4 (the pin facets, the effects tests and the grades package still name the dropped tables).
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/...` succeeds; `CGO_ENABLED=1 go test -tags fts5 ./core/database/ -run 'TestCreateOpen|TestMigrate|TestParseMigrations'` passes (`TestCreateOpen` runs `verifySchema` on the new step).

**STEP-2 — Delete facets, claim release, pins**
Files: `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/README.md`, `core/database/identityclaims/pins.go` (delete), `core/database/identityclaims/bridgefile.go`, `core/database/promote/unpromote.go`, `core/database/observations/write.go`, `core/database/deleteimpact/pins_e2e_test.go` (delete), `core/database/deleteimpact/conclusion_release_test.go`, `core/database/deleteimpact/facets_test.go`, `core/database/deleteimpact/conclusion_test.go`, `core/database/promote/batch_test.go`, `core/database/promote/bridges_test.go`, `core/database/promote/unpromote_test.go`, `core/database/autoreconciler/autoreconcilertest/autoreconcilertest.go` *(new)*. Skill: `add-catalog-delete`.
Do: D-3, D-6, D-10, and D-7: T-1, T-2, T-3, T-4, and the pin count block in `TestSaveBatchFilesClaimsAndBridgesOnly`.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./core/database/deleteimpact/... ./core/database/identityclaims/... ./core/database/promote/...` passes (tests run in STEP-4).

**STEP-3 — Effects tests and comment**
Files: `core/database/effects/scopes.go`, `core/database/effects/handle_paths_test.go`, `core/database/effects/reads_test.go`, `core/database/effects/registry_completeness_test.go`. Skill: `add-audit-scope`.
Do: the rest of D-4 (the registry entries went in STEP-1).
Check: `CGO_ENABLED=1 go vet -tags fts5 ./core/database/effects/...` passes.

**STEP-4 — The grades vocabulary**
Files: `core/database/claimconfidencegrades/` (delete), `core/apperr/apperr.go`, `core/onboarding/ready.go`, `core/onboarding/ready_test.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/gallery/gallery_test.go`. Skill: `add-seeded-vocabulary`.
Do: D-5, and D-7's `assertUpkeepEqualsRebuild` change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes, including `TestPragmaHonesty`, `TestFacetReleaseHonesty`, `TestResourceAndOwnedHonesty`, `TestPragmaHonestyProjectors`, the effects completeness tests, T-1 – T-4.

**STEP-5 — Swift**
Files: `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-8.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; Xcode tests pass (`scripts/check-pr.sh` runs them) with no Swift warning.

**STEP-6 — Docs**
Files: `docs/conclusion-layer-data-model.md` (§4 confidence and `argument` paragraphs and the `identity_claims` DDL, §5 "planned work" sentence in the past tense, §5.3 "status and argument" → "status", §7 per D-9, §10 invariants 19 and 20, §12's claim argument examples), `docs/seeded-vocabulary.md` (line 44 row, §5.5), `docs/research-judgment-model.md` (line 166, per D-9 (c)), `docs/deployment-plan/spike-13/person-detail.md` R-40, `docs/catalog-deletes.md` (item 9, the facet table row, the pinned-Observation bullet), `docs/audit-revision-history.md` (lines 335 and 337), `docs/catalog-deletes.md` line 46 (drop "including a pin's claim after that claim row is gone"), `docs/conclusion-layer-data-model.md` §3 item 2 ("a status, with no Observation exhibit"), `.agents/skills/add-catalog-delete/SKILL.md` (line 77 comment drops "pins"), `.agents/skills/add-audit-scope/SKILL.md` (lines 28, 54 and 109: "its claims", "claim, member evidence", no pins), `.agents/skills/add-seeded-vocabulary/SKILL.md` (lines 28, 88, 104), `.cursor/rules/catalog-audit.mdc` ("claims, pins, promote" → "claims, promote"), `docs/ideas/identity-claim-review.md` (Status), `completed.md`, `deployment-plan.md` (map row), this plan (status).
Do: per D-9 and the Definition of done.
Check: AC-3.1 and AC-3.2's `rg` give the stated hits; `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/deleteimpact/conclusion_test.go` `TestImpactConclusion`, subtests "a member's observation names no handle" and "observation delete audits only the observation" *(new)* | `york`'s promoted Observation `obsID` → Impact `Allowed`, no Groups, no Cascades; delete it → last revision `delete_observation` with types `observation`; `identityclaims.Get(c, claim.ID)` succeeds | AC-2.1, AC-2.2 |
| T-2 | Go | `core/database/deleteimpact/conclusion_release_test.go` `TestMemberDeleteEndToEnd` *(new)* | `newConclusionFixture`: delete `obsB` (Impact allowed, no Cascades); then Impact of B → `assertLeaves(t, report, f.entityRef)`; delete B → types `identity_claim,subject`; `AcceptedMembers(entityID)` is `[ca]`; `autoreconcilertest.AssertCacheMatchesRebuild(t, c)` | AC-2.3, D-7 |
| T-3 | Go | `core/database/deleteimpact/facets_test.go` `TestReleaseFacetsReportsHandles` | `ReleaseFacets(tx, KindSubject, f.b.ID)` → changes `identity_claim`; `effects.Handles` → `[f.entityID]` | AC-2.4 |
| T-4 | Go | `core/database/promote/bridges_test.go` `TestBridgeSubjectDeleteKeepsTheEndClaims` *(new)* | James and Birth promoted, participation bridge filed; delete the bridge Subject → no error; `identityclaims.AcceptedEntityForSubject(c, james.ID)` succeeds | D-7 |
| T-5 | Go | `core/database/catalogmodel` `TestPragmaHonesty`; `core/database/deleteimpact/pragma_test.go` `TestFacetReleaseHonesty` (existing) | Live schema after `000058.sql` → model and releases match | AC-1.1, D-1, D-2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/effects` completeness tests (table lists), `core/database/deleteimpact` honesty tests, `TestUnpromote` (`ReleaseClaim`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Prose sweep: `rg -n -w -i 'pins?|pinned|exhibit' .agents .cursor docs/*.md` hits only the kept lines D-9 lists, "dropped by S13-25" sentences, and unrelated pins (cover pins, map pins, "pinned" headers, "the test pins", `macos-client-patterns.md`, `narrative-layer-data-model.md`, `application-stack.md`, `interpretation-layer-data-model.md` 575)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: conclusion-layer-data-model.md §4, §5, §5.3, §7, §10, §12; seeded-vocabulary.md §5.5; catalog-deletes.md; audit-revision-history.md (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The user runs this on a Mac after the PR is opened; it needs the app built at two commits. An implementer on a host without Xcode marks §10 and STEP-5's Xcode check "not run (no macOS host)" in the PR description and writes no upgrade harness (`add-catalog-migration` forbids migration tests). AC-1.1 – AC-1.3 are evidenced by §10; T-5 and `TestCreateOpen` give the shape half on any host.

1. At the planning base (`588e4a82`), build the app, make a project, and add one Source with three Place Subjects. Promote A, join B onto A's handle, promote C and reject its claim by SQL (`UPDATE identity_claims SET status = 'rejected' WHERE subject_id = <C>`). Insert four pins and set a grade and argument by SQL (`INSERT INTO identity_claim_evidence …`; `UPDATE identity_claims SET confidence_grade_id = (SELECT id FROM claim_confidence_grades WHERE key = 'low_confidence'), argument = 'same age' WHERE subject_id = <B>`). Record `SELECT id, subject_id, entity_id, status FROM identity_claims` and `SELECT COUNT(*) FROM audit_changes WHERE entity_type = 'identity_claim_evidence'`, and `SELECT COUNT(*) FROM audit_transaction_scopes`.
2. Open that project with the PR's build. Run the queries of AC-1.1 and AC-1.2: the claim rows and the two counts match step 1; the two tables, the index and the two columns are gone (AC-1.1, AC-1.2).
3. In that project, delete one of B's Observations: the confirm names no person or place handle (AC-2.1).
4. Create a new project with the PR's build: `SELECT name FROM sqlite_schema WHERE name = 'claim_confidence_grades'` returns nothing (AC-1.3).
5. Run AC-3.1's and AC-3.2's `rg` (AC-3.1, AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The rebuild loses claims or an index | D-1 copies every row by column name; the schema hash and `TestPragmaHonesty` catch a missing index; manual check 2 compares the rows |
| A deleted test was the only cover for a delete path that stays | D-7 keeps member delete, cache upkeep, handle reporting and bridge delete in T-2 – T-4 |
| A project mid-upgrade is left half migrated | The step runs in one transaction with its `user_version` write (`add-catalog-migration`) |

**Stop and report if:**

- `migrate` refuses `000058.sql` with a foreign key violation (D-1's premise that nothing else references `identity_claims`).
- `verifySchema` reports `catalog.schema_mismatch` after a correct `000058.sql`.
- `rg -n 'confidence_grade_id' core --glob '!*_test.go' --glob '!core/database/migrations/**'` at the planning base hits a file other than `core/database/catalogmodel/model.go`, `core/database/deleteimpact/facets.go` and `core/database/deleteimpact/register.go` (D-11).
- `rg -n 'claimconfidencegrades|PinTx|PinnedObservations' core api` finds a non-test importer not named in §4.
- `TestFacetReleaseHonesty` requires a release for a `KindIdentityClaim` facet after the pin entries go.
- The effects registry init panics on a table this plan does not name.

## 12. Open questions

None.

## 13. PR

**Title:** S13-25 — The pin table goes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-25-the-pin-table-goes.md`.
