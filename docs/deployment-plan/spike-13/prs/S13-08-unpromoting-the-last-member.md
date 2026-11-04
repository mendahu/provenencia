# S13-08 — Unpromoting the last member

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Evidence graph's Unpromote confirm and the Subject delete confirm keep their layout, controls and tone; each gains one message variant, the copy the user approved 2026-11-04)
**Board:** —
**Depends on:** S13-25 (landed [#644](https://github.com/mendahu/provenencia/pull/644))
**Blocks:** S13-17
**Branch:** `s13-08-unpromoting-the-last-member`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, and this PR adds no catalog format step ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-l10n.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-38, R-39; [conclusion §5.2](../../../conclusion-layer-data-model.md#52-when-a-member-leaves), §6, §6.1, §10 invariant 7; [`catalog-deletes.md`](../../../catalog-deletes.md) item 9; working rules: "One Unpromote, deletes through the registry", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-11-04 against `main` @ `bb3effb`
**Review:** [`S13-08-unpromoting-the-last-member.review.md`](S13-08-unpromoting-the-last-member.review.md) — round 1 answered; round 2 pending

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

Today, unpromoting a Person's only record leaves a Person with no records: it vanishes from the Persons list but still counts in the sidebar, and its id lives on in the catalog with nothing behind it. After this PR, removing a handle's last accepted member deletes the handle (canonical entity) in the same revision, through `core/database/deleteimpact`, audited. The rule lives in the one claim removal, `releaseClaim`, so both writes that remove a member get it: Unpromote (on the Evidence graph now, from the Person page's Remove in S13-17) and Subject delete. Both confirms say so first: Unpromote reads a new preflight, `GetUnpromoteImpact`, and Subject delete reads a new cascade group in its existing `GetDeleteImpact` report. The same rule deletes the participation, relationship and location handles whose last bridge member an Unpromote or a bridge Subject's delete removes.

**Why now.** S13-17 (Remove a claim) needs the engine to delete the person and the confirm to know it will; S13-25 has landed, so `deleteimpact.ReleaseClaim` is stable.

**What stays.** Unpromote still deletes the Subject's accepted claim and the accepted claims of its bridges, and never rejects. The cards stay on the Evidence graph, and the Evidence graph stays where it is after the write. A handle that keeps another accepted member is untouched, and so is a handle that never had an accepted member (an Inferred, Asserted or Unlinked stub, conclusion §6.1). Promote, the Persons list and the Person page do not change in this PR (S13-17 owns landing on the Persons list). Evidence graph and Source deletes are unchanged: they are refused while they hold Subjects (`subjects.graph_id`, `subjects.source_id` inbound edges), so they never remove a member.

## 2. User stories

**US-1.** As a researcher, when I unpromote a Subject that is its Person's only record, the Person is deleted with it, so no empty Person is left behind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person `PER-A` has one accepted member, Subject `S1`; Unpromote `S1` | `PER-A`'s `canonical_entities` row is gone; the revision holds a `canonical_entity` delete change for it; the write's effects carry `PER-A` in `handles` and `counts` in `lists` |
| AC-1.2 | Person `PER-B` has members `S1` (graph 1) and `S2` (graph 2); Unpromote `S1` | `PER-B` stays; only `S1`'s claim is deleted (today's behavior) |
| AC-1.3 | `S1` is the only record of `PER-A` and the subject of a birth participation `PAR-1` whose only member is the bridge on `S1`'s graph; Unpromote `S1` | `PER-A` and `PAR-1` are both deleted; the bridge's other end (the birth Event) stays |
| AC-1.4 | The same as AC-1.1 for an Event and for a Place | The Event or Place handle is deleted |
| AC-1.5 | `PER-A`'s only accepted member is `S1`, and `PER-C` has `merged_into_id = PER-A`; Unpromote `S1` | The claim is deleted; `PER-A` stays (Impact refuses on `canonical_entities.merged_into_id`); the write succeeds |
| AC-1.6 | After AC-1.1 | The FTS document for `PER-A` is gone; `conclusionsummaries.ListPersons` omits it; the graph cache holds no node for it; an `audit_transaction_scopes` row names it |
| AC-1.7 | Handle `PER-D` has no accepted member and one provisional claim from `S3`; `S3` is deleted | The provisional claim is released; `PER-D` stays |

**US-2.** As a researcher, before I confirm an Unpromote on the Evidence graph, the confirm tells me when it will delete the Person, Event or Place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `GetUnpromoteImpact` for `S1`, the only member of `PER-A` | `deletes` is a group: via `canonical_entities.last_member`, kind `canonical_entity`, total 1, listed `PER-A` with location section `persons` and `entity_id` set |
| AC-2.2 | `GetUnpromoteImpact` for `S1` when `PER-B` has another member, for an unpromoted Subject, and for AC-1.5's `PER-A` | `deletes` is unset |
| AC-2.3 | `GetUnpromoteImpact` with a malformed id; with an id no Subject has; with a bridge Subject | malformed → `parseID`'s error, as `UnpromoteSubject`; unknown → `promote.ErrInvalid`; bridge → `promote.ErrUnsupportedType` |
| AC-2.4 | The Evidence graph's Unpromote on `James Robins`, the only record of `PER-7KD45` | The confirm's message reads "PER-7KD45 has no other records, so this person is deleted. The cards stay on the graph." (event: "this event", place: "this place") |
| AC-2.5 | The Unpromote on a Subject whose Person keeps another record | Today's message: "This promotion is removed. The cards stay on the graph, and the handle stays." |
| AC-2.6 | `GetUnpromoteImpact` fails (store error) | The confirm opens with today's message and the error in its danger callout; Confirm still runs the write |

**US-3.** As a researcher, when I delete the only Subject of a Person (or a bridge that is a link's only record), the handle is deleted with it, and the delete confirm says so.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `subjects.Delete` on `S1`, the only accepted member of `PER-A` | `PER-A` is deleted in the same revision with a `canonical_entity` delete change |
| AC-3.2 | `GetDeleteImpact(subject, S1)` for AC-3.1 | `cascades` holds `canonical_entities.last_member` listing `PER-A`, and no `identity_claims.subject_id` group |
| AC-3.3 | `GetDeleteImpact(subject, S2)` where `S2` has an accepted claim on `PER-B` (which keeps another member) | `cascades` holds `identity_claims.subject_id` listing `PER-B` only (today's line) |
| AC-3.4 | The Subject delete confirm for AC-3.1 | Its cascade line reads "PER-7KD45 has no other records, so it is deleted." in place of "It’s removed from PER-7KD45. That record stays."; for two refs, "PER-A and REL-1 have no other records, so they are deleted." |
| AC-3.5 | `subjects.Delete` on a bridge Subject whose accepted claim is the only member of `REL-1` | `REL-1` is deleted |

## 3. Scope

| In | Covered by |
| --- | --- |
| `releaseClaim` deletes the handle when the claim it removes was the handle's last accepted member (`releaseHandleIfGone` *(new)*) | STEP-2, AC-1.1, AC-1.3, AC-1.5, AC-1.7, AC-3.1, AC-3.5 |
| `canonical_entities` becomes deletable: `PK` in `catalogmodel`, `deletableKinds`, projector through `workspacelocation.Conclusion` *(new)*, facet release for a dying handle's remaining claims | STEP-1 |
| One "the handle goes without this Subject" check (`handleGoesWithout` *(new)*) used by the write, the Subject delete cascade and the Unpromote preflight | STEP-2, STEP-3, AC-2.1, AC-3.2 |
| Subject delete's Impact splits handles it deletes from handles it leaves | STEP-3, AC-3.2, AC-3.3 |
| `GetUnpromoteImpact` RPC *(new)* and `promote.UnpromoteImpact` *(new)*; one primary-Subject loader for both Unpromote paths | STEP-4, AC-2.1, AC-2.2, AC-2.3 |
| Swift: store method, FakeStore script, async `beginUnpromote`, the Unpromote confirm's last-record message | STEP-5, STEP-6, AC-2.4, AC-2.5, AC-2.6 |
| Swift: `PVDeleteImpactCopy.cascadeLines` words `canonical_entities.last_member` | STEP-6, AC-3.4 |
| Old-behavior tests rewritten; docs: conclusion §5.2, §6, §10 invariant 7; `catalog-deletes.md` items 9 and the facet table; `deleteimpact` README; `add-catalog-delete`; the map | STEP-2, STEP-3, STEP-7 |

| Out | Home |
| --- | --- |
| The Person page's Remove control, the last-claim warning there, landing on the Persons list | S13-17 |
| Cleaning up memberless handles already in a project (user, 2026-11-04: no dogfood project has any) | not planned |
| A researcher-facing Delete for a handle (a Person page Delete button, `DeleteCanonicalEntity` RPC) | not planned |
| Merging handles (the `merged_into_id` writer) | not planned this spike ([conclusion §9](../../../conclusion-layer-data-model.md#9-canonical-merge)) |
| Naming in a confirm the provisional or rejected claims a dying handle releases (no writer creates them; conclusion §5.3) | not planned |
| Rewording "the handle stays" in the not-last Unpromote message | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Unpromote write | `core/database/promote/unpromote.go` `Unpromote` | Loads the Subject (`subjects.GetTx`, no rows → `ErrInvalid`), checks the type inline (`subjecttypes.GetByIDTx` + `vocabkey.IsPrimaryType`), finds the accepted claim, calls `deleteimpact.ReleaseClaim`, then `identityclaims.UnfileBridges`. Doc comment: "deletes no handle" |
| Primary check | `core/database/promote/promote.go` `requirePrimary` | The same type check as `Unpromote`'s inline copy |
| Unpromote handler | `api/ffi/handlers/promote.go` `UnpromoteSubject` | `parseID`, then `writes.Run` action `unpromote_subject` |
| Bridge unfiling | `core/database/identityclaims/bridgefile.go` `UnfileBridges` | Releases each accepted bridge claim through `deleteimpact.ReleaseClaim`. Comment line 94: "The bridges' handles (their associations) are left alone." |
| Claim release | `core/database/deleteimpact/facets.go` `ReleaseClaim`, `releaseClaim` | The one claim removal: scans `id, subject_id, entity_id, subject_type_id, status`; facets (none); row delete; `identity_claim` delete change. Callers: `ReleaseClaim` (from `Unpromote`, `UnfileBridges`) and the `identity_claims.subject_id` facet's `releaseEach` |
| Subject's claim facet | `core/database/deleteimpact/facets.go` `buildFacetReleases` entry `identity_claims.subject_id` | `Named: true`, child `KindCanonicalEntity`, Count/List every handle the Subject has a claim on (any status), release `releaseEach(…, releaseClaim)`. Comment: "the handles stay" |
| Named contract | `core/database/deleteimpact/facets.go` `facetRelease.Named` | "Count / List must select exactly what Release removes" |
| Cascades | `core/database/deleteimpact/facets.go` `cascadesFor`; `impact.go` `collectGroups` | `Named` releases become `Report.Cascades` groups through `inboundEdge` Count/List |
| Deletable kinds | `core/database/deleteimpact/register.go` `deletableKinds`, `tableByKind`; `extraGateList` (the per-kind list pattern) | No `KindCanonicalEntity`; `tableByKind` skips a table with no `PK` |
| Inbound edges | `register.go` `inboundEdgeList` | Has `canonical_entities.merged_into_id` |
| Catalog model | `core/database/catalogmodel/model.go` `Tables` | `{Name: "canonical_entities", Kind: KindCanonicalEntity, Bucket: BucketResource}` with **no `PK`**; `KindCanonicalEntity` comment "no delete path yet". `core/database/effects/path.go` `pkOf` already defaults to `"id"` |
| Projectors | `core/database/deleteimpact/projectors.go` `projectKind` | No canonical entity case; default returns the ref with no `Section` |
| Conclusion locations | `core/search/fts.go` `locationFor` | Builds `{Section: Persons|Events|Places, EntityID, Ref, Title}` inline per kind; `core/workspacelocation/workspacelocation.go` has the shared constructor `Graph` |
| Honesty tests | `core/database/deleteimpact/pragma_test.go` `TestDeletableKindsResolve`, `TestPragmaHonestyProjectors`, `TestFacetReleaseHonesty` | A deletable kind needs a `PK`, a projector with a `Section`, and one release per audited CASCADE into it |
| Subject delete | `core/database/subjects/subjects.go` `Delete` | Impact → Refuse → `ReleaseFacets` → `DELETE` → `effects.Handles(released)` → `RecomputeTx` |
| Effects | `core/database/effects/registry.go` `canonical_entities` | `Handles: self()`, `HandleScope: self()`, `Lists: counts` (`ListsOnField: "merged_into_id"`: create and delete move counts) |
| Graph cache / search | `core/database/graphcache/commit.go` `OnCommit`; `core/database/searchindex/handles.go` `reprojectHandleBatch` | A handle id with no row is removed from the cache; its FTS doc is deleted |
| Impact preflight precedent | `api/ffi/handlers/evidence_graphs.go` `GetEvidenceGraphUnlinkImpact`; `api/ffi/handlers/delete_impact.go` `groupToProto` | |
| Proto | `api/proto/engine.proto` `UnpromoteSubjectRequest` comment "No handle is deleted."; `DeleteImpactGroup`; last method `METHOD_LIST_PERSON_EVENTS = 114` | |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `unpromoteSubject` (doc "No handle is deleted"), `CatalogDeleteImpactListed`; `GoStore.swift` `unpromoteSubject`, `mapDeleteImpactGroup`; `CoreInvoke.swift` `CoreMethod`; `FakeStore.swift` `unpromoteSubject` (doc "Handles stay."), scripted knobs `unlinkImpactBlocks`, `deleteImpactByID`, the Subject `getDeleteImpact` cascade near line 2899; `macos/ProvenenciaTests/OnboardingModelTests.swift` test store | |
| Swift confirm | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `PendingUnpromote`, `beginUnpromote` (sync), `confirmUnpromote`, `performCardAction`; `EvidenceCanvasCard.swift` named action; `EvidenceGraphView.swift` `.pvConfirm(item: $model.pendingUnpromote …)`; `EvidenceGraphSnapshot.swift` `EvidencePrimaryKind`, `EvidenceGraphPlacedSubject.kind`; `L10n.EvidenceGraph.openHandlePage(kind:)` (the per-kind key pattern) | |
| Delete recipe | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactCopy.cascadeLines`, preview fixture near line 471; `L10n.DeleteImpact.deletesCitations(count:refs:)` (plural precedent) | |
| Tests asserting the old rule | `core/database/identityclaims/identityclaims_test.go` "subject delete cascades its claim, handle stays"; `core/database/deleteimpact/conclusion_test.go` "subject delete removes its claim explicitly and audits it; handle stays" and the `assertLeaves` cases | Each asserts a sole-member handle survives |
| Other tests | `core/database/promote/unpromote_test.go` `TestUnpromote`; `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject`; `core/database/subjects/delete_test.go`; `core/database/deleteimpact/conclusion_release_test.go` `TestMemberDeleteEndToEnd` (two-member fixture `newConclusionFixture`); `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges`, `unpromoteWithNoBridgesHasEmptyDetail`, `confirmUnpromoteRefreshesToPromote`; `macos/ProvenenciaTests/PVDeleteImpactTests.swift`; `macos/ProvenenciaTests/SubjectStoreTests.swift` (asserts the FakeStore cascade via) | |

**Precedent.** `releaseCitation` in `core/database/deleteimpact/facets.go`: one removal in `deleteimpact` that every writer reaches, so all of them audit the same rows. This PR puts the handle rule in `releaseClaim` the same way. `GetEvidenceGraphUnlinkImpact` is the precedent for a preflight returning a `DeleteImpactGroup`; `beginDelete` in `EvidenceGraphModel.swift` (async, through `Task` from `performCardAction`) is the Swift precedent; `L10n.EvidenceGraph.openHandlePage(kind:)` for per-kind keys; `deletesCitations(count:refs:)` for a plural key.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The rule:** when a write removes a handle's last accepted claim, the handle is deleted in the same transaction, for every handle type: Person, Event, Place and the association types (user, 2026-11-04). Only the removal of an **accepted** claim makes a handle a candidate | Working rule "deletes the entity wherever Unpromote runs"; "last member" is the last accepted claim (conclusion §2.2, §10 invariant 10); a stub with no accepted member is a designed state (§6.1) | Any released claim makes a candidate — would delete an Asserted stub when its provisional claim goes |
| D-2 | **Where:** inside `releaseClaim`. It keeps its scan, delete and change, and then, when the removed claim's `status` was `accepted`, calls `releaseHandleIfGone(tx, entityID, subjectID)` *(new)* and merges what it released. `ReleaseClaim` and the `identity_claims.subject_id` facet keep calling `releaseClaim`, so `promote.Unpromote`, `UnfileBridges` and `subjects.Delete` gain the rule with no code change | One claim removal already; every writer reaches it (Subject delete through the facet, Unpromote and bridges through `ReleaseClaim`). A future claim-removing writer cannot forget the rule | A `ReleaseEmptiedHandles` call each writer adds (round 1, M-2) — a second decoder of change fields and a call to forget |
| D-3 | **A dying handle's other claims:** the new facet release `identity_claims.entity_id` *(new)*, parent `KindCanonicalEntity`, not `Named`, `Remaining` `SELECT COUNT(*) FROM identity_claims WHERE entity_id = ?`, release `releaseEach("SELECT id FROM identity_claims WHERE entity_id = ? ORDER BY id", releaseClaimRow)`. `releaseClaimRow` *(new, unexported)* is today's body of `releaseClaim` (scan, delete, change) returning also the scanned `entity_id` and `status`; `releaseClaim` becomes `releaseClaimRow` plus D-2's step | The handle's leftover claims are provisional or rejected (D-1), so they never recurse into another handle delete; `UNIQUE (subject_id, entity_id)` (migration `000058.sql`) means none of them is the claim being removed | Letting SQLite `CASCADE` drop them — unaudited (`TestFacetReleaseHonesty`) |
| D-4 | **One check:** `handleGoesWithout(tx *sql.Tx, entityID, subjectID []byte) (bool, error)` *(new)*: true when the row exists, no Subject other than `subjectID` holds an accepted claim on it, and `Impact(tx, KindCanonicalEntity, entityID).Allowed`. The write calls it after the claim row is gone; the preflight and the cascade call it before. The SQL holds only the membership test; the merge block comes from Impact's `canonical_entities.merged_into_id` edge | One rule for the confirm and the write; any later inbound edge into `canonical_entities` (Reconciliation Claims, conclusion §7) reaches both through Impact (`add-catalog-delete` "from the same query the write refuses on") | A SQL predicate with a merged clause beside Impact (round 1, M-1) — two copies of the merge rule |
| D-5 | **`releaseHandleIfGone`:** if `handleGoesWithout` is false, nothing. Else `releaseFacets(tx, KindCanonicalEntity, id)` (D-3), `DELETE FROM canonical_entities WHERE id = ?`, and a `canonical_entity` delete change with `rowchange.DeletedRow` over `id`, `subject_type_id`, `ref`, `argument`, `label`, `merged_into_id`, `preferred_likeness_observation_id` (UUIDs as `database.UUIDJSON`, text as `database.NullString`, as `canonicalentities.InsertTx` writes them) | `add-catalog-delete` one-row release; the change lets the ghost map scope the revision to the handle (`add-audit-scope`) | |
| D-6 | **A refused handle stays** and the write succeeds (AC-1.5). Impact refuses only when another handle is merged into it; the app has no merge writer | The claim's removal is the researcher's request; a stub that merges point at is safe | Failing the write with a new error code — unreachable, and would need copy and a double |
| D-7 | **The table becomes deletable:** `catalogmodel.Tables` row gains `PK: "id"`; `KindCanonicalEntity` joins `deletableKinds`; its comment says the handle is deleted by `releaseClaim` when its last member goes. `core/database/effects/path.go` `pkOf` already returns `"id"` for it, so effects do not change | `TestDeletableKindsResolve` requires a `PK` (round 1, X-1) | |
| D-8 | **Location:** `workspacelocation.Conclusion(typeKey, entityID, ref, title string) (Location, bool)` *(new)*: `vocabkey.Person` → `Persons`, `Event` → `Events`, `Place` → `Places`, with `EntityID`, `Ref`, `Title`; false for any other key. `core/search/fts.go` `locationFor`'s three Conclusion cases call it (with `vocabkey` keys, which equal `KindPerson`/`KindEvent`/`KindPlace`). `projectCanonicalEntity` *(new)* in `projectors.go` reads `ref`, `label`, the type `key` and `origin`; title = label, else ref; when `vocabkey.IsPrimaryType(key, origin)` it uses `Conclusion`; otherwise (association types) `{Section: Sources, Ref: ref, Title: title}`; a missing row returns the ref and `{Section: Persons, Ref: ref}`, as `projectEvidenceGraph` does for its missing row | One mapping for search and delete reports (round 1, M-5); an association gets no Person link; `TestPragmaHonestyProjectors` needs a `Section` | A catch-all to Persons — links a relationship to a Person page |
| D-9 | `GetDeleteImpact(kind: "canonical_entity")` becomes valid as a side effect of D-7. No UI calls it | One registry | Hiding the kind from `ParseKind` — a second list |
| D-10 | **Subject delete's report:** the `identity_claims.subject_id` facet becomes not `Named` (its release is unchanged). Two cascade-only edges *(new)* in `extraCascadeList` *(new)*, a `sync.OnceValue` list of `inboundEdge` beside `extraGateList` in `register.go`, with `Bucket: catalogmodel.BucketFacet`; `cascadesFor(kind)` returns the `Named` facet edges for `kind` followed by the `extraCascadeList` entries whose `Parent == kind`. **(a)** `canonical_entities.last_member`: parent `KindSubject`, child `KindCanonicalEntity`, rows = `lastMemberRows(tx, subjectID)` *(new)*: the handle of the Subject's accepted claim, as `probeRow{ID, Ref}`, kept when `handleGoesWithout(handle, subjectID)` (0 or 1 rows: invariant 3). **(b)** `identity_claims.subject_id`: parent `KindSubject`, child `KindCanonicalEntity`, rows = every distinct handle the Subject has a claim on (any status, today's List SQL) minus (a)'s, ordered by ref. Count is `len(rows)`, List caps at `limit` | `Named` keeps its contract (round 1, M-8); the two groups name every handle the release touches (T-7) | Narrowing a `Named` facet's Count/List — breaks "Count / List must select exactly what Release removes" |
| D-11 | **The preflight:** `deleteimpact.LastMemberGroup(tx *sql.Tx, subjectID []byte) (Group, error)` *(new)* runs `collectGroups` over the one `canonical_entities.last_member` edge and returns its group, or `Group{}`. `promote.UnpromoteImpact(c *database.Catalog, subjectID []byte) (deleteimpact.Group, error)` *(new)* opens a read transaction and rolls it back (as `deleteimpact.Snapshot`), loads through `loadPrimary` *(new)*, and returns `LastMemberGroup` | One edge feeds both reports; associations are not listed (the confirm lists bridges already) | A boolean `deletes_handle` — the working rule asks for an Impact report |
| D-12 | **One loader:** `loadPrimary(q *sql.Tx, subjectID []byte) (subjects.Subject, error)` *(new)* in `promote.go`: `subjects.GetTx`, no rows → `ErrInvalid`, then `requirePrimary`. `Unpromote` uses it and drops its inline type check | The preflight and the write refuse the same inputs (round 1, M-10) | A third copy |
| D-13 | **RPC:** `GetUnpromoteImpact` *(new)*, `METHOD_GET_UNPROMOTE_IMPACT = 115`, request `{project_dir = 1, subject_id = 2}`, response `{DeleteImpactGroup deletes = 1}`, set only when the group's via is non-empty (as `GetEvidenceGraphUnlinkImpact` sets `citations`); handler in `api/ffi/handlers/promote.go` through `withProjectCatalog`; `parseID` errors as `UnpromoteSubject` | A read, as `GetEvidenceGraphUnlinkImpact` | A dry-run flag on `UnpromoteSubjectRequest` |
| D-14 | **Swift store:** `getUnpromoteImpact(projectDir:subjectID:) async throws -> CatalogDeleteImpactListed?` *(new)*: the first listed row of `deletes`, `nil` when unset. FakeStore is scripted, not derived: `unpromoteImpactBySubject: [String: CatalogDeleteImpactListed]` *(new)*, `getUnpromoteImpactError: Error?` *(new)*, `getUnpromoteImpactCalls` *(new)*; it returns the scripted row or `nil` | One reading of the report for both callers (S13-17 reuses it; round 1, M-12); `swift-tests.mdc`: FakeStore does not re-implement membership (round 1, M-4) | FakeStore deriving the rule from `membershipBySubject` |
| D-15 | **Swift model:** `PendingUnpromote` gains `kind: EvidencePrimaryKind` (from `placed.kind`) and `deletedHandleRef: String?` *(new)*. `beginUnpromote(subjectID:)` becomes `async`: it builds the pending value from the snapshot as today, awaits `getUnpromoteImpact`, sets `deletedHandleRef` to the row's `ref`, and assigns `pendingUnpromote` once. On a thrown error it assigns the pending value with `deletedHandleRef = nil` and `error = L10n.Errors.message(for: error)`. `performCardAction`'s case and `EvidenceCanvasCard`'s named action call it through `Task { await … }`, as `beginDelete` | One open, no message change under the reader | Opening first and patching the message after the fetch |
| D-16 | **Unpromote copy** (user, 2026-11-04): `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind: EvidencePrimaryKind, ref: String) -> String` *(new)*, a `switch` with no `default` over keys `evidenceGraph.subject.unpromoteConfirmMessageDeletes.person` / `.event` / `.place`, each filled by `L10n.format`. The view passes it as the confirm's `message` when `deletedHandleRef != nil`, else today's `L10n.string(L10n.EvidenceGraph.unpromoteConfirmMessage)` | Typed kind; a new kind is a compile error (round 1, M-6); same pattern as `openHandlePage(kind:)` | A `String` kind with a person fallback |
| D-17 | **Subject delete copy** (planner's call, in the form of the approved Unpromote copy): `L10n.DeleteImpact.deletesHandles(count: Int, handleRefs: String) -> String` *(new)*, one key `deleteImpact.confirm.deletesHandles` with `plural` variations in `Localizable.xcstrings`: one "%2$@ has no other records, so it is deleted.", other "%2$@ have no other records, so they are deleted." `cascadeLines` gains the case `canonical_entities.last_member` → `deletesHandles(count: group.total, handleRefs: joined)` | `macos-l10n.mdc` "Never `count == 1 ? oneKey : otherKey`" (round 1, M-3); precedent `deletesCitations(count:refs:)` | Two keys picked by count |
| D-18 | **FakeStore's Subject cascade** stays as it is (it emits `identity_claims.subject_id`); the test that needs the new via scripts `deleteImpactByID` | FakeStore scripts reports, it does not derive them | Teaching FakeStore the last-member rule |
| D-19 | **Tests sit where the code is:** Go table tests in `deleteimpact`, `promote`, `subjects`, `identityclaims`, handlers; Swift Testing for the model and the recipe copy | `go-tests.mdc`, `add-swift-test` | |

**Rules this plan bends.** Two domain rules change, each on the user's word, and each is updated in STEP-7:

- [conclusion §5.2](../../../conclusion-layer-data-model.md#52-when-a-member-leaves) "It does not reject, and it deletes no handle", §6 "Dropping the last member does not delete the entity or change its id" and §10 invariant 7 → working rule "One Unpromote, deletes through the registry" (user, 2026-10-26; R-39) and Subject delete (user, 2026-11-04). §6.1's stubs (no accepted member ever) keep their rule (D-1).
- [`catalog-deletes.md`](../../../catalog-deletes.md) item 9 "Handles have no delete path yet" and its facet table line "`entity_id` has no delete path yet" → same answers.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `releaseHandleIfGone`, `releaseClaimRow` *(new)* | func | `releaseClaim`, `releaseCitation` | Extends the one claim removal; `releaseClaimRow` is today's `releaseClaim` body, split so the entity facet can release without recursing |
| `handleGoesWithout` *(new)* | func | `deleteimpact.Impact` | Adds only the membership test in front of Impact; searched `rg "status = 'accepted'" core/database` for an existing "only member" query: none |
| `extraCascadeList`, `lastMemberRows`, `LastMemberGroup` *(new)* | registry / func | `extraGateList`; `collectGroups`; `ListedGroup` | The per-kind list pattern for cascades; `facetReleaseList` cannot hold a cascade with no release |
| facet release `identity_claims.entity_id` *(new)* | registry entry | facet release `identity_claims.subject_id` | Same claim release, other parent |
| `workspacelocation.Conclusion` *(new)*, `projectCanonicalEntity` *(new)* | func | `core/search/fts.go` `locationFor`; `workspacelocation.Graph` | Lifts `locationFor`'s three inline cases into the shared constructor |
| `promote.UnpromoteImpact`, `loadPrimary` *(new)* | func | `evidencegraphs.UnlinkImpact`; `requirePrimary` | Beside `Unpromote`; `loadPrimary` wraps `requirePrimary` |
| `GetUnpromoteImpact` *(new)* | FFI | `GetEvidenceGraphUnlinkImpact`, `GetDeleteImpact` | `GetDeleteImpact` takes the row to delete; here a claim goes and the handle only as a consequence |
| `getUnpromoteImpact`, FakeStore knobs *(new)* | store method | `getEvidenceGraphUnlinkImpact`, `unlinkImpactBlocks` | |
| `PendingUnpromote.kind`, `.deletedHandleRef` *(new)* | fields | `PendingUnpromote.bridges` | |
| keys `evidenceGraph.subject.unpromoteConfirmMessageDeletes.person|event|place`, `deleteImpact.confirm.deletesHandles` *(new)* | key | `openHandlePage.*`, `deletesCitations` | New sentences |

**Builds on.** `deleteimpact` (the one claim removal, Impact, facet releases, projectors, `collectGroups`), the effects entry for `canonical_entities` (moves `handles`, handle scope and `counts` on delete), `graphcache.OnCommit`'s deleted-handle path, `searchindex`'s missing-handle delete, `workspacelocation`, `DeleteImpactFlow` and the `PVDeleteImpact` recipe.
**Deletes.** `Unpromote`'s inline type check (replaced by `loadPrimary`); `locationFor`'s three inline Conclusion locations (replaced by `Conclusion`). Comments made false: "deletes no handle" (`unpromote.go`), "No handle is deleted" (`UnpromoteSubjectRequest`, `GenealogyStore.unpromoteSubject`), "Handles stay." (`FakeStore.unpromoteSubject`), "are left alone" (`bridgefile.go` line 94), "the handles stay" (the `identity_claims.subject_id` facet), "no delete path yet" (`KindCanonicalEntity`, `catalog-deletes.md` item 9 and facet table). Tests asserting the old rule are rewritten (STEP-2, STEP-3). No function loses its last caller.
**Next adds.** S13-17's Remove calls `getUnpromoteImpact` and `unpromoteSubject` with no engine change; a researcher-facing handle Delete reuses `releaseHandleIfGone`'s release and the projector; Reconciliation Claims' FK into `canonical_entities` reaches the preflight and the write through Impact (D-4).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A handle id in a write's effects whose row is gone | `graphcache.OnCommit`, `searchindex.reprojectHandleBatch`, `autoreconciler.RecomputeReported` / `RecomputeTx` | Removes the node; deletes the FTS doc; recompute deletes and reloads no rows (T-4) |
| A Person (or Event / Place) page in history whose handle was deleted from the Evidence graph | `GetConclusionDetail` → `conclusiondetails.ErrNotFound` → "This record no longer exists. It may have been merged into another." | Existing error state; landing on the Persons list is S13-17 |
| Lists and counts that cached the handle | `CatalogQueryRegistry` `.personsList` / `.eventsList` / `.placesList` refresh on any touched handle; sidebar counts on `counts` | Effects carry the id and `counts` (T-2) |
| `GetDeleteImpact(canonical_entity, id)` | `deleteimpact.Impact` | Allowed for an existing handle not merged into; unused by the app (D-9) |
| A search hit's location for a Person, Event or Place | `workspacelocation.Conclusion` | Same fields as today (T-11) |

### Scale

At the map's Scale (a Source of 100 people; the 41-Subject obituary): each accepted claim removal adds `handleGoesWithout` (one indexed count on `identity_claims(entity_id)` plus Impact's one count on `merged_into_id`) and, for a dying handle, one claim scan and one delete. An Unpromote removes one primary claim and its bridges (tens on the obituary's household), so tens of extra queries per write. Subject delete's Impact adds two small edges over the Subject's claims (a handful). `GetUnpromoteImpact` is a few indexed queries. Recompute, cache and search work are today's, keyed on the same ids. No budget is set.

### Edges

- **Assumptions.** A Subject has at most one accepted claim (§10 invariant 3), so `canonical_entities.last_member` has 0 or 1 rows (T-6). `UNIQUE (subject_id, entity_id)` means a dying handle's leftover claims belong to other Subjects and are never accepted (D-1, D-3), so the entity facet's releases never start a second handle delete (T-1 "leftover claims"). `RecomputeTx` and `OnCommit` accept ids with no row (T-4). `releaseClaim`'s callers already expect it to return all the changes it made, and `ReleaseClaim` already schedules the purge for everything in `released` (a handle owns no File).
- **Neighbors.** One other accepted member → stays (AC-1.2); no other accepted member but a provisional claim from another Subject → handle deleted and that claim released with it (T-1); a provisional claim removed from a handle with no accepted member → stays (AC-1.7); merged into → stays (AC-1.5); an association shared with another graph's bridge → stays (T-2).
- **Removed inputs.** The handle's reconciler rows go by SQLite CASCADE (skip bucket) and are recomputed away; its preferred likeness pointer, label and argument are recorded in the delete change (T-1). Its neighbors' cached links drop in `OnCommit` (the deleted handle's neighbors become roots).
- **Real data.** None — no table, seed or configuration drives the rule.
- **Untrusted size.** None — the input is the claims one write removes.

### Schema

None — no format step. `catalogmodel` gains `PK: "id"` on `canonical_entities` (the column is already the primary key); `TestPragmaHonesty` must stay green.

### Go core

| Package | Change |
| --- | --- |
| `core/database/catalogmodel` | `canonical_entities` row `PK: "id"`; `KindCanonicalEntity` comment |
| `core/workspacelocation` | `func Conclusion(typeKey, entityID, ref, title string) (Location, bool)` *(new)* |
| `core/search` | `locationFor`'s Person, Event and Place cases call `workspacelocation.Conclusion` |
| `core/database/deleteimpact` | `releaseClaimRow` *(new)*, `releaseClaim` + `releaseHandleIfGone` *(new)*, `handleGoesWithout` *(new)*; `KindCanonicalEntity` in `deletableKinds`; facet release `identity_claims.entity_id` *(new)*; the `identity_claims.subject_id` facet not `Named`; `extraCascadeList` *(new)* with its two edges, `lastMemberRows` *(new)*; `cascadesFor` appends; `LastMemberGroup` *(new)*; `projectCanonicalEntity` *(new)*; `ReleaseClaim` doc comment |
| `core/database/promote` | `loadPrimary` *(new)*; `Unpromote` uses it, doc comment; `UnpromoteImpact` *(new)* |
| `core/database/identityclaims` | `bridgefile.go` comment only |

Sketch of the membership half of D-4:

```sql
SELECT COUNT(*) FROM identity_claims
WHERE entity_id = ? AND status = 'accepted' AND subject_id != ?
```

Errors: none new.

### FFI

`api/proto/engine.proto`: `METHOD_GET_UNPROMOTE_IMPACT = 115` *(new)*; `GetUnpromoteImpactRequest { string project_dir = 1; string subject_id = 2; }` *(new)*; `GetUnpromoteImpactResponse { DeleteImpactGroup deletes = 1; }` *(new)*, comment "the handle the Unpromote deletes because the Subject is its last member; unset when it stays". In `UnpromoteSubjectRequest`'s comment, "No handle is deleted." becomes "A handle whose last accepted member this removes is deleted (conclusion §5.2)." Handler `GetUnpromoteImpact` in `api/ffi/handlers/promote.go`; `api/ffi/dispatch.go` constant and case. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store read | `getUnpromoteImpact` *(new)* | `GenealogyStore.swift`, `GoStore.swift` (+ `CoreMethod.getUnpromoteImpact` *(new)*), `FakeStore.swift`, `OnboardingModelTests.swift` test store | D-14 |
| Unpromote confirm | existing `.pvConfirm` | `EvidenceGraphView.swift` | D-16 |
| Subject delete cascade | `PVDeleteImpactCopy.cascadeLines` | `PVDeleteImpact.swift` | D-17; the preview fixture gains a `canonical_entities.last_member` group |

States: preflight error → not-last message, error in the callout (D-15). Accessibility: identifiers unchanged (`evidenceGraph.unpromote.*`); the message is plain text.

### Strings

L10n keys *(new)*:

- `evidenceGraph.subject.unpromoteConfirmMessageDeletes.person` = "%@ has no other records, so this person is deleted. The cards stay on the graph."
- `evidenceGraph.subject.unpromoteConfirmMessageDeletes.event` = "%@ has no other records, so this event is deleted. The cards stay on the graph."
- `evidenceGraph.subject.unpromoteConfirmMessageDeletes.place` = "%@ has no other records, so this place is deleted. The cards stay on the graph."
- `deleteImpact.confirm.deletesHandles` (plural on `%1$lld`): one "%2$@ has no other records, so it is deleted.", other "%2$@ have no other records, so they are deleted."

Old-scope copy: `evidenceGraph.subject.unpromoteConfirmMessage` ("This promotion is removed. The cards stay on the graph, and the handle stays.") stays for the not-last case; `deleteImpact.confirm.leavesHandle` ("It’s removed from %@. That record stays.") stays for handles that keep members. Removed: none.

## 7. Steps

**STEP-1 — The handle table becomes deletable**
Files: `core/database/catalogmodel/model.go`, `core/workspacelocation/workspacelocation.go`, `core/workspacelocation/workspacelocation_test.go`, `core/search/fts.go`, `core/database/deleteimpact/register.go`, `projectors.go`, `facets.go`. Skill: `add-catalog-model`, `add-catalog-delete`.
Do: D-7, D-8, D-3 (the `identity_claims.entity_id` facet with `releaseClaimRow`; `releaseClaim` still equals `releaseClaimRow` in this step).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/catalogmodel/ ./core/database/deleteimpact/ ./core/workspacelocation/ ./core/search/` passes, including `TestPragmaHonesty`, `TestDeletableKindsResolve`, `TestPragmaHonestyProjectors`, `TestFacetReleaseHonesty`, T-11.

**STEP-2 — Removing the last member deletes the handle**
Files: `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/handle_release_test.go` *(new)*, `core/database/promote/promote.go`, `unpromote.go`, `unpromote_test.go`, `core/database/identityclaims/bridgefile.go` (comment), `core/database/identityclaims/identityclaims_test.go`, `core/database/subjects/delete_test.go`, `core/database/deleteimpact/conclusion_test.go`. Skill: `add-catalog-delete`, `add-audit-scope`.
Do: D-2, D-4, D-5, D-6, D-12. Rewrite "subject delete cascades its claim, handle stays" to give `per1` a second accepted member (the handle stays) and add a sole-member row (the handle goes). In `conclusion_test.go`, the "handle stays" case keeps its name only if its fixture handle has another member; otherwise rename it with "handle deleted" in place of "handle stays" and assert the `canonical_entity` change in the revision types.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes; T-1, T-2, T-3, T-4.

**STEP-3 — Subject delete's report names the handles it deletes**
Files: `core/database/deleteimpact/register.go`, `facets.go`, `impact.go`, `core/database/deleteimpact/impact_test.go`, `conclusion_test.go` (the `assertLeaves` cases).
Do: D-10, and `LastMemberGroup` of D-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/ ./core/database/subjects/` passes; T-6, T-7.

**STEP-4 — The Unpromote preflight**
Files: `core/database/promote/unpromote.go`, `api/proto/engine.proto`, generated code, `api/ffi/dispatch.go`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/promote_test.go`. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-11 (`UnpromoteImpact`), D-13; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/database/promote/` passes; T-5.

**STEP-5 — Swift store method**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `CoreInvoke.swift`, `FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`. Skill: `add-swift-test`.
Do: D-14; rewrite the `unpromoteSubject` docs in `GenealogyStore.swift` and `FakeStore.swift`.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) builds with no warning.

**STEP-6 — The confirms word the outcome**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `EvidenceCanvasCard.swift`, `EvidenceGraphView.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-15, D-16, D-17. In the three existing unpromote tests: `unpromoteActionOpensConfirmNamingBridges` waits with `waitUntil` for `model.pendingUnpromote != nil` after `performCardAction`; `unpromoteWithNoBridgesHasEmptyDetail` and `confirmUnpromoteRefreshesToPromote` call `await model.beginUnpromote(subjectID:)`.
Check: `xcodebuild test` passes; T-8, T-9, T-10.

**STEP-7 — Docs**
Files: `docs/conclusion-layer-data-model.md` §5.2 (removing a handle's last accepted member, by Unpromote or Subject delete, deletes the handle unless a handle is merged into it), §6 (the sentence "Dropping the last member does not delete the entity or change its id." says the same; stubs with no accepted member are unaffected), §10 invariant 7; `docs/catalog-deletes.md` item 9 and the facet table line "`entity_id` has no delete path yet" (→ "`entity_id`: released with the handle when its last member goes"); `core/database/deleteimpact/README.md` (`releaseClaim` deletes a handle whose last member it removes; `canonical_entities.last_member`); `.agents/skills/add-catalog-delete/SKILL.md` (the same sentence beside `ReleaseClaim`); `docs/deployment-plan/spike-13/deployment-plan.md` (map row; scope entry; working rule "One Unpromote" gains Subject delete, user 2026-11-04); `completed.md`; this plan's status.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-08-unpromoting-the-last-member.md` and `python3 scripts/check-guidance-names.py` pass.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/deleteimpact/handle_release_test.go` `TestReleaseClaimDeletesEmptiedHandle` *(new)* | table over `ReleaseClaim`: sole accepted member → row gone, one `canonical_entity` delete change carrying all seven fields; another accepted member → stays; leftover provisional claim from another Subject → handle and that claim gone, both changes, no second handle delete; provisional claim released from a handle with no accepted member → stays; merged into → stays, no error | AC-1.1, AC-1.2, AC-1.5, AC-1.7, D-1, D-2, D-3, D-5, D-6 |
| T-2 | Go | `core/database/promote/unpromote_test.go` `TestUnpromote` (new cases) | sole member through `writes.Run` → handle gone, effects carry its id in `Handles` and `counts` in `Lists`; two members → stays; birth participation only on this graph → person and participation gone, Event stays; association with a bridge on another graph too → stays; Event and Place sole members → gone | AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| T-3 | Go | `core/database/subjects/delete_test.go` `TestDeleteDeletesEmptiedHandles` *(new)*; `core/database/identityclaims/identityclaims_test.go` (rewritten case) | sole member → handle gone in the same revision; bridge that is a relationship's only member → relationship gone; handle with another member → stays | AC-3.1, AC-3.5, AC-1.2 |
| T-4 | Go | `core/database/deleteimpact/conclusion_release_test.go` `TestLastMemberDeleteEndToEnd` *(new)*, its own one-member fixture | Unpromote of the only member through `writes.Run` → `searchindex` has no doc for the handle (query the FTS table by entity id), `c.Graph()` returns no node for it, `conclusionsummaries.ListPersons` omits it, `audit_transaction_scopes` has a `canonical_entity` row for its id at that revision | AC-1.6 |
| T-5 | Go | `api/ffi/handlers/promote_test.go` `TestGetUnpromoteImpact` *(new)*, `runRPC` | sole member → `deletes` via `canonical_entities.last_member`, kind `canonical_entity`, total 1, listed ref, location section `persons`, `entity_id`; other member → unset; unpromoted → unset; merged into → unset; malformed id → `parseID` error; unknown → `promote.ErrInvalid`; bridge Subject → `promote.ErrUnsupportedType`; bad bytes → unmarshal error | AC-2.1, AC-2.2, AC-2.3 |
| T-6 | Go | `core/database/deleteimpact/handle_release_test.go` `TestLastMemberGroupMatchesRelease` *(new)* | for each T-1 fixture, `LastMemberGroup` before the release lists exactly the handles the release deletes | D-4, D-11 |
| T-7 | Go | `core/database/deleteimpact/impact_test.go` `TestSubjectCascadesSplitHandles` *(new)* | Subject with the accepted claim on sole-member `PER-A` and a provisional claim on `PER-B` (which has members) → `canonical_entities.last_member` [`PER-A`], `identity_claims.subject_id` [`PER-B`]; union = every handle the Subject has a claim on | AC-3.2, AC-3.3, D-10 |
| T-8 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteLastRecordNamesTheDeletedHandle` *(new)* | `householdModel()` with `unpromoteImpactBySubject["s-james"]` scripted to a row with ref `PER-7KD45` → `pendingUnpromote.deletedHandleRef == "PER-7KD45"`, `kind == .person`; nothing scripted → `nil` | AC-2.4, AC-2.5, D-15 |
| T-9 | Swift | `EvidenceGraphModelTests.swift` `unpromotePreflightFailureStaysInTheConfirm` *(new)* | `getUnpromoteImpactError` set → `pendingUnpromote` open with `error` set and `deletedHandleRef == nil`; `confirmUnpromote` records `unpromoteSubject id=s-james` | AC-2.6 |
| T-10 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` `cascadeLinesNameDeletedHandles` *(new)* | group `canonical_entities.last_member` total 1 [`PER-7KD45`] → "PER-7KD45 has no other records, so it is deleted."; total 2 [`PER-A`, `REL-1`] → "PER-A and REL-1 have no other records, so they are deleted."; `unpromoteConfirmMessageDeletes(kind:ref:)` for `.person`, `.event`, `.place` → the three sentences | AC-3.4, AC-2.4, D-16, D-17 |
| T-11 | Go | `core/workspacelocation/workspacelocation_test.go` `TestConclusion` *(new)* | `person`/`event`/`place` → `Persons`/`Events`/`Places` with `EntityID`, `Ref`, `Title`; `relationship` → false | D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break, and rewritten in STEP-2 / STEP-3 rather than patched: the two "handle stays" cases and the `assertLeaves` cases; also `TestFacetReleaseHonesty`, `TestUnpromoteSubject`, the three existing unpromote Swift tests, `SubjectStoreTests` (must stay green unchanged, D-18).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: "deletes no handle", "No handle is deleted", "Handles stay", "handle stays", "the handles stay", "are left alone", "no delete path yet" and "Dropping the last member does not delete" `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/`, `api/proto/`, `macos/App/` and test names, with every remaining hit true; `scripts/check-guidance-names.py` passes
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI: the two confirms keep their layout; only the messages of §6 Strings change
- [ ] Version bump: none (header)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Domain docs updated: conclusion §5.2, §6, §10 invariant 7; catalog-deletes.md item 9 and facet table; deleteimpact README; add-catalog-delete
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a dogfood project. On an Evidence graph, pick a promoted Person card whose Person has no record on any other graph (its Person page's Sources show one Source). Choose Unpromote. The confirm reads "PER-… has no other records, so this person is deleted. The cards stay on the graph." Confirm. The card shows Promote again; the Persons list no longer has the person; the sidebar Persons count drops by one.
2. On a Person with records on two graphs, Unpromote one card. The confirm reads "…and the handle stays." The Person stays in the list.
3. Promote a Subject alone onto a new Person, then Delete that Subject card. The delete confirm says "PER-… has no other records, so it is deleted." Confirm; the Person is gone from the list and count.
4. Go Back to a deleted Person's page in history: it shows "This record no longer exists."

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `RecomputeTx` / `RecomputeReported` on a deleted id writes a stale row | T-4 |
| Moving the rule into `releaseClaim` changes what an existing caller expects back | Its callers already append all returned changes; T-2, T-3 cover each caller |
| Making `beginUnpromote` async lets a second click race the first | `pendingUnpromote` is assigned once after the fetch; a second open replaces it, as `beginDelete` |

**Stop and report if:**

- `releaseClaim` has a caller at the base other than `ReleaseClaim` and the `identity_claims.subject_id` facet.
- Adding `PK: "id"` to `canonical_entities` changes any `core/database/effects` test result.
- `TestFacetReleaseHonesty` or another `deleteimpact` test fails once the `identity_claims.subject_id` facet is not `Named`.
- `autoreconciler.RecomputeTx`, `graphcache.OnCommit` or `searchindex` fails or leaves rows for a handle id with no row (T-4 cannot pass without changing them).
- `effects.Resolve` on a `canonical_entity` delete change does not yield `counts` in `Lists`.
- Importing `vocabkey` into `workspacelocation` creates an import cycle.
- A Swift caller of `beginUnpromote` other than `performCardAction` and `EvidenceCanvasCard` exists at the base.

## 12. Open questions

None.

## 13. PR

**Title:** S13-08 — Unpromoting the last member
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-08-unpromoting-the-last-member.md`.
