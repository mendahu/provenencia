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
**Skills to load:** [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-l10n.mdc`
**Context:** [`person-detail.md`](../person-detail.md) R-38, R-39; [conclusion §5.2](../../../conclusion-layer-data-model.md#52-when-a-member-leaves), §6, §10 invariant 7; [`catalog-deletes.md`](../../../catalog-deletes.md) item 9; working rules: "One Unpromote, deletes through the registry", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-11-04 against `main` @ `bb3effb`
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

Today, unpromoting a Person's only record leaves a Person with no records: it vanishes from the Persons list but still counts in the sidebar, and its id lives on in the catalog with nothing behind it. After this PR, a write that leaves a canonical entity (a handle) with no accepted member deletes that handle in the same revision, through `core/database/deleteimpact`, audited. Two writes can do that: Unpromote (on the Evidence graph now, from the Person page's Remove in S13-17) and Subject delete. Both confirms say so first: Unpromote reads a new preflight, `GetUnpromoteImpact`, and Subject delete reads a new cascade group in its existing `GetDeleteImpact` report. The same rule deletes the participation, relationship and location handles that an Unpromote or a bridge Subject's delete leaves with no member.

**Why now.** S13-17 (Remove a claim) needs the engine to delete the person and the confirm to know it will; S13-25 has landed, so `deleteimpact.ReleaseClaim` is stable.

**What stays.** Unpromote still deletes the Subject's accepted claim and the accepted claims of its bridges, and never rejects. The cards stay on the Evidence graph, and the Evidence graph stays where it is after the write. A handle that keeps another accepted member is untouched. Promote, the Persons list and the Person page do not change in this PR (S13-17 owns landing on the Persons list). Evidence graph and Source deletes are unchanged: they are refused while they hold Subjects (`subjects.graph_id`, `subjects.source_id` inbound edges), so they never empty a handle.

## 2. User stories

**US-1.** As a researcher, when I unpromote a Subject that is its Person's only record, the Person is deleted with it, so no empty Person is left behind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Person `PER-A` has one accepted member, Subject `S1`; Unpromote `S1` | `PER-A`'s `canonical_entities` row is gone; the revision holds a `canonical_entity` delete change for it; the write's effects carry `PER-A` in `handles` and `counts` in `lists` |
| AC-1.2 | Person `PER-B` has members `S1` (graph 1) and `S2` (graph 2); Unpromote `S1` | `PER-B` stays; only `S1`'s claim is deleted (today's behavior) |
| AC-1.3 | `S1` is the only record of `PER-A` and the subject of a birth participation `PAR-1` whose only member is the bridge on `S1`'s graph; Unpromote `S1` | `PER-A` and `PAR-1` are both deleted; the bridge's other end (the birth Event) stays |
| AC-1.4 | The same as AC-1.1 for an Event and for a Place | The Event or Place handle is deleted |
| AC-1.5 | `PER-A`'s only accepted member is `S1`, and `PER-C` has `merged_into_id = PER-A`; Unpromote `S1` | The claim is deleted; `PER-A` stays (Impact refuses on `canonical_entities.merged_into_id`); the write succeeds |
| AC-1.6 | After AC-1.1 | The FTS document for `PER-A` is gone; `PersonSummary` reads no longer list it; the graph cache holds no node for it |

**US-2.** As a researcher, before I confirm an Unpromote on the Evidence graph, the confirm tells me when it will delete the Person, Event or Place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `GetUnpromoteImpact` for `S1`, the only member of `PER-A` | `deletes` is a group: via `canonical_entities.last_member`, kind `canonical_entity`, total 1, listed `PER-A` (location: Persons, `entity_id` set) |
| AC-2.2 | `GetUnpromoteImpact` for `S1` when `PER-B` has another member, for an unpromoted Subject, and for AC-1.5's `PER-A` | `deletes` is empty (no via, total 0) |
| AC-2.3 | `GetUnpromoteImpact` with a malformed id; with an id no Subject has | malformed → `invalid` error as other handlers; unknown → `promote.invalid` (`promote.ErrInvalid`) |
| AC-2.4 | The Evidence graph's Unpromote on `James Robins`, the only record of `PER-7KD45` | The confirm's message reads "PER-7KD45 has no other records, so this person is deleted. The cards stay on the graph." (event: "this event", place: "this place") |
| AC-2.5 | The Unpromote on a Subject whose Person keeps another record | Today's message: "This promotion is removed. The cards stay on the graph, and the handle stays." |
| AC-2.6 | `GetUnpromoteImpact` fails (store error) | The confirm opens with today's message and the error in its danger callout; Confirm still runs the write |

**US-3.** As a researcher, when I delete the only Subject of a Person (or a bridge that is a link's only record), the handle is deleted with it, and the delete confirm says so.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `subjects.Delete` on `S1`, the only accepted member of `PER-A` | `PER-A` is deleted in the same revision with a `canonical_entity` delete change |
| AC-3.2 | `GetDeleteImpact(subject, S1)` for AC-3.1 | `cascades` holds `canonical_entities.last_member` listing `PER-A`, and no `identity_claims.subject_id` group |
| AC-3.3 | `GetDeleteImpact(subject, S2)` where `S2` has accepted claims on `PER-B` (which keeps another member) and on nothing else | `cascades` holds `identity_claims.subject_id` listing `PER-B` only (today's line) |
| AC-3.4 | The Subject delete confirm for AC-3.1 | Its cascade line reads "PER-7KD45 has no other records, so it is deleted." in place of "It’s removed from PER-7KD45. That record stays." |
| AC-3.5 | `subjects.Delete` on a bridge Subject whose accepted claim is the only member of `REL-1` | `REL-1` is deleted |

## 3. Scope

| In | Covered by |
| --- | --- |
| `deleteimpact.ReleaseEmptiedHandles` *(new)*: the one handle removal for a write that released claims | STEP-1, STEP-2, AC-1.1, AC-1.5 |
| `KindCanonicalEntity` becomes a deletable kind: projector, facet release for its remaining claims, Impact allowed when no handle is merged into it | STEP-1, AC-1.5 |
| The last-member query (`deleteimpact.LastMemberHandles` *(new)*) shared by the Unpromote preflight and the Subject delete cascade | STEP-1, AC-2.1, AC-3.2 |
| `promote.Unpromote` deletes the handles it empties, primary and association | STEP-2, AC-1.1 – AC-1.4 |
| `subjects.Delete` deletes the handles it empties; its Impact splits handles it deletes from handles it leaves | STEP-3, AC-3.1 – AC-3.3, AC-3.5 |
| `GetUnpromoteImpact` RPC *(new)* and `promote.UnpromoteImpact` *(new)* | STEP-4, AC-2.1 – AC-2.3 |
| Swift: store method, FakeStore, async `beginUnpromote`, the Unpromote confirm's last-record message | STEP-5, STEP-6, AC-2.4 – AC-2.6 |
| Swift: `PVDeleteImpactCopy.cascadeLines` words `canonical_entities.last_member` | STEP-6, AC-3.4 |
| Docs: conclusion §5.2, §6, §10 invariant 7; `catalog-deletes.md` item 9; `deleteimpact` README; the map (row, scope entry, working rule) | STEP-7 |

| Out | Home |
| --- | --- |
| The Person page's Remove control, the last-claim warning there, landing on the Persons list | S13-17 |
| Cleaning up memberless handles already in a project (user, 2026-11-04: no dogfood project has any) | not planned |
| A researcher-facing Delete for a handle (a Person page Delete button, `DeleteCanonicalEntity` RPC) | not planned |
| Merging handles (the `merged_into_id` writer) | not planned this spike ([conclusion §9](../../../conclusion-layer-data-model.md#9-canonical-merge)) |
| Rewording "the handle stays" in the not-last Unpromote message | not planned |
| A per-handle title projector beyond the ref and label (the confirm titles from its own name) | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Unpromote write | `core/database/promote/unpromote.go` `Unpromote` | Loads the Subject, refuses non-primary types, finds the accepted claim (`identityclaims.AcceptedEntityForSubjectTx`), calls `deleteimpact.ReleaseClaim`, then `identityclaims.UnfileBridges`; returns both change lists. Deletes no handle (doc comment says so) |
| Unpromote handler | `api/ffi/handlers/promote.go` `UnpromoteSubject` | `writes.Run` action `unpromote_subject` around `promote.Unpromote`; returns `WriteEffects` |
| Bridge unfiling | `core/database/identityclaims/bridgefile.go` `UnfileBridges` | Releases each accepted bridge claim touching the Subject through `deleteimpact.ReleaseClaim`; association handles left with no member stay |
| Claim release | `core/database/deleteimpact/facets.go` `ReleaseClaim`, `releaseClaim` | The one claim removal: facets (none), row delete, `identity_claim` delete change carrying `entity_id` |
| Subject's claim cascade | `core/database/deleteimpact/facets.go` `buildFacetReleases`, entry `identity_claims.subject_id` | `Named` cascade listing every handle the Subject has a claim on; releases every claim through `releaseEach(..., releaseClaim)`. Comment: "the handles stay" |
| Cascade listing | `core/database/deleteimpact/facets.go` `cascadesFor` | Turns `Named` facet releases into `Report.Cascades` groups |
| Deletable kinds | `core/database/deleteimpact/register.go` `deletableKinds` | No `KindCanonicalEntity` |
| Inbound edges | `core/database/deleteimpact/register.go` `inboundEdgeList` | Already has `canonical_entities.merged_into_id` (parent and child `KindCanonicalEntity`) |
| Projectors | `core/database/deleteimpact/projectors.go` `projectKind` | No canonical entity case: the default returns the ref with no `Section` |
| Honesty tests | `core/database/deleteimpact/pragma_test.go` `TestFacetReleaseHonesty`, `TestPragmaHonestyProjectors` | Every audited CASCADE into a deletable table needs one facet release; every deletable kind needs a projector with a `Section` |
| Catalog model | `core/database/catalogmodel/model.go` | `canonical_entities` is `BucketResource`; inbound: `identity_claims.entity_id` (audited facet CASCADE), `auto_reconciler_values.entity_id` / `auto_reconciler_outcomes.entity_id` (skip CASCADE), `canonical_entities.merged_into_id` (NO ACTION). Comment on `KindCanonicalEntity`: "no delete path yet" |
| Subject delete | `core/database/subjects/subjects.go` `Delete` | Impact → Refuse → snapshot handles → `deleteimpact.ReleaseFacets` → `DELETE` → `autoreconciler.RecomputeTx` on touched handles → subject delete change |
| Effects | `core/database/effects/registry.go` entry `canonical_entities` | `Handles: self()`, `HandleScope: self()`, `Lists: counts` with `ListsOnField: "merged_into_id"` (create and delete move counts) |
| Graph cache | `core/database/graphcache/commit.go` `OnCommit` | A handle id in the set with no row is removed (`removeNode`) and its neighbors become roots; an association with no row is detached |
| Search | `core/database/searchindex/handles.go` `reprojectHandleBatch` | A handle id with no row deletes its FTS document |
| Impact preflight precedent | `api/ffi/handlers/evidence_graphs.go` `GetEvidenceGraphUnlinkImpact`; `api/ffi/handlers/delete_impact.go` `groupToProto` | A read RPC returning `DeleteImpactGroup`s |
| Proto | `api/proto/engine.proto` `UnpromoteSubjectRequest` / `Response` (comment "No handle is deleted"); `DeleteImpactGroup`; last method `METHOD_LIST_PERSON_EVENTS = 114` | |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `unpromoteSubject` (doc "No handle is deleted"), `CatalogDeleteImpactGroup`; `GoStore.swift` `unpromoteSubject`, `mapDeleteImpactGroup` (private static); `FakeStore.swift` `unpromoteSubject`; `macos/ProvenenciaTests/OnboardingModelTests.swift` test store `unpromoteSubject` stub | |
| Swift confirm | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `PendingUnpromote`, `beginUnpromote` (sync), `confirmUnpromote`, `performCardAction` (calls `beginUnpromote` at the `unpromoteActionID` case); `EvidenceCanvasCard.swift` named action calls `model.beginUnpromote`; `EvidenceGraphView.swift` `.pvConfirm(item: $model.pendingUnpromote …)` with `L10n.EvidenceGraph.unpromoteConfirmMessage` | |
| Delete recipe | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactCopy.cascadeLines` (case `identity_claims.subject_id` → `L10n.DeleteImpact.leavesHandle`), preview fixture near line 471 | |
| Tests | `core/database/promote/unpromote_test.go` `TestUnpromote`; `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject`; `core/database/subjects/delete_test.go`; `core/database/deleteimpact/conclusion_release_test.go` `TestMemberDeleteEndToEnd`; `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges`, `unpromoteWithNoBridgesHasEmptyDetail`, `confirmUnpromoteRefreshesToPromote`; `macos/ProvenenciaTests/PVDeleteImpactTests.swift` | |

**Precedent.** `releaseCitation` / `ReleaseCitation` in `core/database/deleteimpact/facets.go`: a one-row removal living in `deleteimpact`, with an exported `*database.Tx` wrapper, called by two domain writers (`citations.Delete` and graph delete) so both audit the same row. Copy its shape for `ReleaseEmptiedHandles`. `GetEvidenceGraphUnlinkImpact` is the precedent for a preflight that returns `DeleteImpactGroup` through `deleteimpact.ListedGroup`; copy its handler shape. `beginDelete` in `EvidenceGraphModel.swift` (async, called through `Task` from `performCardAction`) is the Swift precedent.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The rule:** after a write releases claims, every handle named by a released claim that has no accepted claim left is deleted in the same transaction. It applies to every handle type: Person, Event, Place and the association types (user, 2026-11-04) | Working rule "deletes the entity wherever Unpromote runs"; one rule for primaries and associations, one loop | Delete only the primary — user chose all, 2026-11-04 |
| D-2 | **Writers:** `promote.Unpromote` and `subjects.Delete` (user, 2026-11-04). Evidence graph and Source deletes are not touched | They are the only writes that release claims and can leave a handle memberless; graph and Source deletes refuse while Subjects exist | Unpromote only — user chose to include Subject delete |
| D-3 | **One removal:** `deleteimpact.ReleaseEmptiedHandles(tx *database.Tx, released []rowchange.Change) ([]rowchange.Change, error)` *(new)* reads the `entity_id` of each `identity_claim` delete change in `released`, dedupes, sorts by id, and for each runs `releaseHandle` *(new, unexported)* | `deleteimpact` writes those changes (`releaseClaim`), so it owns their shape; both writers share one function, like `ReleaseCitation` | A `canonicalentities.Delete` — the delete is not a researcher command and needs `deleteimpact` internals; reading entity ids with `effects.Handles` — wider (observers), not what was released |
| D-4 | **`releaseHandle`** for one id: skip if the row is gone; skip if an accepted claim remains; `Impact(KindCanonicalEntity)`; skip (no error) when not allowed; `releaseFacets(KindCanonicalEntity)` (its remaining provisional or rejected claims); `DELETE FROM canonical_entities`; append a `canonical_entity` delete change with `rowchange.DeletedRow` over `id`, `subject_type_id`, `ref`, `argument`, `label`, `merged_into_id`, `preferred_likeness_observation_id` | Follows the `add-catalog-delete` pattern for a one-row release; the claim rule is checked by the caller, not by Impact | An Impact extra gate "has members" — no command deletes a handle with members, so the gate would guard nothing |
| D-5 | **A refused handle stays** and the write succeeds (AC-1.5). Impact refuses only when another handle is merged into it; the app has no merge writer | The claim's removal is the researcher's request; keeping a stub that merges point at is safe | Failing the whole write with a new error code — an unreachable path that would need copy and a test double |
| D-6 | `KindCanonicalEntity` joins `deletableKinds`. A facet release `identity_claims.entity_id` *(new)*, parent `KindCanonicalEntity`, not `Named`, `Remaining` `SELECT COUNT(*) FROM identity_claims WHERE entity_id = ?`, release `releaseEach(... ORDER BY id, releaseClaim)`. `projectCanonicalEntity` *(new)* in `projectors.go`: title = label, else ref; location `Section` from the type key (`person` → `persons`, `event` → `events`, `place` → `places`, others → `persons`) with `EntityID`; a missing row returns the ref and `Section: persons`, as `projectEvidenceGraph` does for its missing row | Required by `TestFacetReleaseHonesty` and `TestPragmaHonestyProjectors` once the kind is deletable; claims stay released through `releaseClaim` | A Section-less projector — fails the honesty test |
| D-7 | `GetDeleteImpact(kind: "canonical_entity")` becomes valid as a side effect of D-6 (`ParseKind`). No UI calls it | One registry; the kind is deletable now | Hiding the kind from `ParseKind` — a second list to keep in step |
| D-8 | **Last member, one query:** `deleteimpact.LastMemberHandles(tx *sql.Tx, subjectID []byte, limit int) ([]Listed, int, error)` *(new)* returns the handles on which `subjectID` holds the accepted claim, no other Subject holds an accepted claim, and no handle is merged into it, ordered by ref, capped at `limit`, with the total. One SQL predicate constant `sqlHandleEmptied` *(new)* is shared with `releaseHandle`'s checks | The preflight and the write must agree; one predicate keeps them so (T-6) | Two queries, one per caller |
| D-9 | **Subject delete's Impact:** the `identity_claims.subject_id` group's `Count` and `List` exclude the handles `LastMemberHandles` returns; a new cascade-only edge `canonical_entities.last_member` *(new)*, child `KindCanonicalEntity`, lists them. It is one entry in a new `extraCascadeList` *(new)* in `register.go`, appended by `cascadesFor` for its kind. The facet release is unchanged (it still releases every claim) | The confirm must not say a handle both stays and goes; the two groups together name every handle the release touches (T-7) | A second `Named` facet release — two releases would claim one FK (`TestFacetReleaseHonesty` forbids duplicates) |
| D-10 | **The preflight:** `promote.UnpromoteImpact(c *database.Catalog, subjectID []byte) (deleteimpact.Group, error)` *(new)* opens a read transaction (rolled back, as `deleteimpact.Snapshot`), validates as `Unpromote` does (bad id or unknown Subject → `ErrInvalid`, non-primary → `ErrUnsupportedType`), and returns `deleteimpact.ListedGroup(tx, "canonical_entities.last_member", KindCanonicalEntity, rows)` from `LastMemberHandles`; an empty `Group{}` when there are none. Associations are not listed: the confirm already lists the bridges | Same via as Subject delete, so S13-17 and the recipe share one key; a primary Subject has at most one accepted claim, so the group has 0 or 1 rows | A boolean `deletes_handle` — the working rule asks for an Impact report, not a one-off flag |
| D-11 | **RPC:** `GetUnpromoteImpact` *(new)*, `METHOD_GET_UNPROMOTE_IMPACT = 115`, request `{project_dir = 1, subject_id = 2}`, response `{DeleteImpactGroup deletes = 1}`; handler in `api/ffi/handlers/promote.go` through `withProjectCatalog`, no `writes.Run` | A read, as `GetEvidenceGraphUnlinkImpact` | Extending `UnpromoteSubjectRequest` with a dry-run flag — mixes a read into a write |
| D-12 | **Swift:** `beginUnpromote(subjectID:)` becomes `async`: it builds the pending confirm from the snapshot as today, then awaits `store.getUnpromoteImpact`, and sets `PendingUnpromote.deletedHandleRef` *(new, `String?`)* to the group's first listed ref and `PendingUnpromote.kind` *(new, `String`, the membership kind)*. It sets `pendingUnpromote` once, after the fetch. A fetch error opens the confirm with `deletedHandleRef = nil` and `error` set to `L10n.Errors.message(for:)` (AC-2.6). `performCardAction` and `EvidenceCanvasCard` call it through `Task { await … }`, as `beginDelete` | Same flow as `beginDelete`; one open, no flicker between messages | Opening first and patching the message after the fetch — the message would change under the reader |
| D-13 | **Unpromote copy** (user, 2026-11-04): `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` *(new)* picks one of three keys by kind; `kind` other than `event` / `place` uses the person key. Title, buttons, tone and the bridges list are unchanged | Three full sentences translate cleanly; a noun slot does not | One key with a kind noun argument |
| D-14 | **Subject delete copy** (planner's call, same form as the user-approved Unpromote copy): `L10n.DeleteImpact.deletesHandle(handleRefs:)` *(new)* = "%@ has no other records, so it is deleted." when `group.total == 1`; `L10n.DeleteImpact.deletesHandles(handleRefs:)` *(new)* = "%@ have no other records, so they are deleted." when `group.total > 1`. The argument is the refs `cascadeLines` already joins | The recipe owns via wording; an unknown via already renders through `alsoAffects` | Leaving the fallback `alsoAffects` line — it does not say the records go |
| D-15 | **Tests sit where the code is:** Go table tests in `deleteimpact`, `promote`, `subjects`, handlers; Swift Testing for the model and the recipe copy | `.cursor/rules/go-tests.mdc`, `add-swift-test` | |

**Rules this plan bends.** Two domain rules change, each on the user's word, and each is updated in STEP-7:

- [conclusion §5.2](../../../conclusion-layer-data-model.md#52-when-a-member-leaves) "It does not reject, and it deletes no handle", §6 "Dropping the last member does not delete the entity or change its id" and §10 invariant 7 → working rule "One Unpromote, deletes through the registry" (user, 2026-10-26; R-39) and Subject delete (user, 2026-11-04).
- [`catalog-deletes.md`](../../../catalog-deletes.md) item 9 "Handles have no delete path yet" and the `KindCanonicalEntity` comment "no delete path yet" → same answers.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `deleteimpact.ReleaseEmptiedHandles` *(new)*, `releaseHandle` *(new)* | func | `deleteimpact.ReleaseClaim` / `ReleaseCitation` | Same shape (exported Tx wrapper + unexported release), different row; searched `rg "DELETE FROM canonical_entities"`: none exists |
| `deleteimpact.LastMemberHandles` *(new)*, `sqlHandleEmptied` *(new)* | func / const | `identityclaims.AcceptedEntityForSubjectTx`; the `identity_claims.subject_id` facet's `List` SQL | Neither answers "is this the only accepted member"; searched `rg "status = 'accepted'" core/database` |
| `extraCascadeList` *(new)* | registry | `extraGateList` in `register.go` | Same per-kind list idea for cascades; `facetReleaseList` cannot hold a cascade with no release |
| `projectCanonicalEntity` *(new)* | func | `projectEvidenceGraph` | One case in `projectKind`, required by the honesty test |
| facet release `identity_claims.entity_id` *(new)* | registry entry | facet release `identity_claims.subject_id` | Same release (`releaseClaim`), other parent |
| `promote.UnpromoteImpact` *(new)* | func | `evidencegraphs.UnlinkImpact` | Lives beside `Unpromote` so validation is shared |
| `GetUnpromoteImpact` *(new)* | FFI | `GetEvidenceGraphUnlinkImpact`, `GetDeleteImpact` | `GetDeleteImpact` takes a row to delete; Unpromote deletes a claim, and the handle only as a consequence |
| `getUnpromoteImpact` *(new)* | store method | `getEvidenceGraphUnlinkImpact` | |
| `PendingUnpromote.deletedHandleRef`, `.kind` *(new)* | fields | `PendingUnpromote.bridges` | Extends the existing pending struct |
| L10n `evidenceGraph.subject.unpromoteConfirmMessageDeletesPerson` / `…Event` / `…Place`, `deleteImpact.confirm.deletesHandle`, `deleteImpact.confirm.deletesHandles` *(new)* | key | `unpromoteConfirmMessage`, `leavesHandle` | New sentences |

**Builds on.** `deleteimpact` (Impact, facet releases, projectors, `ListedGroup`), the effects registry's `canonical_entities` entry (already moves `handles`, handle scope and `counts` on delete), `graphcache.OnCommit`'s deleted-handle path, `searchindex`'s missing-handle delete, `DeleteImpactFlow` and the `PVDeleteImpact` recipe.
**Deletes.** The sentences "deletes no handle" / "No handle is deleted" in `unpromote.go`'s doc comment, the `UnpromoteSubjectRequest` proto comment and `GenealogyStore.unpromoteSubject`'s doc; "the handles stay" in the `identity_claims.subject_id` facet comment; "no delete path yet" on `KindCanonicalEntity` and in `catalog-deletes.md` item 9. No function loses its last caller.
**Next adds.** S13-17's Remove calls `GetUnpromoteImpact` and `UnpromoteSubject` with no engine change; a future researcher-facing handle Delete reuses `releaseHandle` and the projector; a future merge writer meets `merged_into_id` already handled by D-5.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A handle id in a write's effects whose row is gone | `graphcache.OnCommit`, `searchindex.reprojectHandleBatch`, `autoreconciler.RecomputeReported` | Removes the node; deletes the FTS doc; recompute deletes and reloads no rows (T-4 asserts each) |
| A Person page (or Event / Place) open in history whose handle was deleted from the Evidence graph | `GetConclusionDetail` → `conclusiondetails.ErrNotFound` → "This record no longer exists. It may have been merged into another." | Existing error state; landing on the Persons list is S13-17 |
| Lists and sections that cached the deleted handle | `CatalogQueryRegistry` `.personsList` / `.eventsList` / `.placesList` refresh on any touched handle; sidebar counts refresh on `counts` | Effects carry the id and `counts` (T-2) |
| `GetDeleteImpact(canonical_entity, id)` | `deleteimpact.Impact` | Allowed for any existing handle not merged into; unused by the app (D-7) |

### Scale

At the map's Scale (a Source of 100 people; the 41-Subject obituary): an Unpromote releases one primary claim and its bridges (tens on the obituary's household); `ReleaseEmptiedHandles` runs one predicate query and, for an emptied handle, Impact (one count on `merged_into_id`) plus one claim scan and one delete — a constant number of queries per released claim, tens per write. `GetUnpromoteImpact` is one indexed query. Subject delete's Impact gains one query. The recompute and cache work are today's, keyed on the same ids. No budget is set.

### Edges

- **Assumptions.** `releaseClaim` changes carry `entity_id` (T-1 reads it); a Subject has at most one accepted claim (§10 invariant 3), so the preflight has 0 or 1 rows (T-6); `RecomputeTx` and `OnCommit` accept ids with no row (T-4).
- **Neighbors.** A handle with exactly one other accepted member stays (AC-1.2, T-2); a handle with no other accepted member but a provisional or rejected claim is deleted and that claim released with it (T-1 "provisional claim left"); a merged-into handle stays (AC-1.5, T-1); an association shared with another graph's bridge stays (T-2 "shared association").
- **Removed inputs.** The deleted handle's reconciler rows go by SQLite CASCADE (skip bucket); its preferred likeness pointer goes with the row and is recorded in the delete change; its label and argument are recorded in the delete change (T-1 asserts the fields).
- **Real data.** None — no table, seed or configuration drives the rule.
- **Untrusted size.** None — the input is the claims the write released, bounded by one Subject's bridges.

### Schema

None — no catalog change. `canonical_entities` is already registered; this PR adds delete policy only.

### Go core

| Package | Change |
| --- | --- |
| `core/database/deleteimpact` | `func ReleaseEmptiedHandles(tx *database.Tx, released []rowchange.Change) ([]rowchange.Change, error)` *(new)*; `func releaseHandle(tx *sql.Tx, id []byte) (released, error)` *(new)*; `func LastMemberHandles(tx *sql.Tx, subjectID []byte, limit int) ([]Listed, int, error)` *(new)*; `sqlHandleEmptied` *(new)*; `KindCanonicalEntity` in `deletableKinds`; facet release `identity_claims.entity_id` *(new)*; `identity_claims.subject_id` Count/List exclude emptied handles; `extraCascadeList` *(new)* with `canonical_entities.last_member`; `cascadesFor` appends it; `projectCanonicalEntity` *(new)* |
| `core/database/catalogmodel` | Comment on `KindCanonicalEntity` only |
| `core/database/promote` | `Unpromote` appends `deleteimpact.ReleaseEmptiedHandles(tx, append(released, bridges...))`; doc comment; `func UnpromoteImpact(c *database.Catalog, subjectID []byte) (deleteimpact.Group, error)` *(new)* |
| `core/database/subjects` | `Delete` calls `deleteimpact.ReleaseEmptiedHandles(tx, released)` after the `DELETE` and before `RecomputeTx`, and appends its changes before the subject change |

Sketch of the predicate (the implementer writes the final SQL):

```sql
-- sqlHandleEmptied: the handle has no accepted claim and no handle is merged into it
NOT EXISTS (SELECT 1 FROM identity_claims WHERE entity_id = e.id AND status = 'accepted')
AND NOT EXISTS (SELECT 1 FROM canonical_entities m WHERE m.merged_into_id = e.id)
```

`LastMemberHandles` selects handles `e` where `subjectID` holds the accepted claim and no other Subject does, plus the merged clause.

Errors: none new. `UnpromoteImpact` returns `promote.ErrInvalid` / `promote.ErrUnsupportedType` as `Unpromote` does.

### FFI

`api/proto/engine.proto`: `METHOD_GET_UNPROMOTE_IMPACT = 115` *(new)*; `GetUnpromoteImpactRequest { string project_dir = 1; string subject_id = 2; }` *(new)*; `GetUnpromoteImpactResponse { DeleteImpactGroup deletes = 1; }` *(new)*, comment "the handle the Unpromote deletes because the Subject is its last member; empty when it stays". In `UnpromoteSubjectRequest`'s comment, "No handle is deleted." becomes "Every handle the write leaves with no accepted member is deleted (conclusion §5.2)." Handler `GetUnpromoteImpact` in `api/ffi/handlers/promote.go`; `dispatch.go` method constant and case. `deletes` is set only when the group has a via (`groupToProto`). Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store read | `getUnpromoteImpact(projectDir:subjectID:) async throws -> CatalogDeleteImpactGroup?` *(new)* | `GenealogyStore.swift`, `GoStore.swift` (`CoreMethod.getUnpromoteImpact` *(new)* in `CoreInvoke.swift`), `FakeStore.swift`, `OnboardingModelTests.swift` test store | `nil` when `deletes` has no via. FakeStore derives it: a group listing the Subject's membership entity ref when no other Subject in `membershipBySubject` shares that entity id; `getUnpromoteImpactError` *(new)* and `getUnpromoteImpactCalls` *(new)* knobs |
| Unpromote confirm | existing `.pvConfirm` | `EvidenceGraphView.swift` | `message:` is `unpromoteConfirmMessageDeletes(kind:ref:)` when `deletedHandleRef != nil`, else today's message |
| Subject delete cascade | `PVDeleteImpactCopy.cascadeLines` | `PVDeleteImpact.swift` | New case `canonical_entities.last_member`; preview fixture gains the group |

States: preflight error → message as not-last, error in the callout (D-12). Accessibility: the confirm keeps its identifiers (`evidenceGraph.unpromote.*`); the message is plain text read by VoiceOver.

### Strings

L10n keys *(new)*:

- `evidenceGraph.subject.unpromoteConfirmMessageDeletesPerson` = "%@ has no other records, so this person is deleted. The cards stay on the graph."
- `evidenceGraph.subject.unpromoteConfirmMessageDeletesEvent` = "%@ has no other records, so this event is deleted. The cards stay on the graph."
- `evidenceGraph.subject.unpromoteConfirmMessageDeletesPlace` = "%@ has no other records, so this place is deleted. The cards stay on the graph."
- `deleteImpact.confirm.deletesHandle` = "%@ has no other records, so it is deleted."
- `deleteImpact.confirm.deletesHandles` = "%@ have no other records, so they are deleted."

Old-scope copy that names the old behavior: `evidenceGraph.subject.unpromoteConfirmMessage` ("This promotion is removed. The cards stay on the graph, and the handle stays.") stays for the not-last case; `deleteImpact.confirm.leavesHandle` ("That record stays.") stays for handles that keep members. Removed: none.

## 7. Steps

**STEP-1 — `deleteimpact` learns to delete a memberless handle**
Files: `core/database/deleteimpact/register.go`, `facets.go`, `projectors.go`, `core/database/deleteimpact/handle_release_test.go` *(new)*, `core/database/catalogmodel/model.go` (comment). Skill: `add-catalog-delete`, `add-audit-scope`.
Do: D-3, D-4, D-6, D-8, D-9: `deletableKinds`, the `identity_claims.entity_id` facet release, `projectCanonicalEntity`, `sqlHandleEmptied`, `LastMemberHandles`, `releaseHandle`, `ReleaseEmptiedHandles`, `extraCascadeList`, the narrowed `identity_claims.subject_id` Count/List. Update the facet comment ("the handles stay" → handles the Subject empties are listed by `canonical_entities.last_member` and deleted by its writer).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/ ./core/database/catalogmodel/` passes, including `TestFacetReleaseHonesty`, `TestPragmaHonestyProjectors`, T-1, T-6, T-7.

**STEP-2 — Unpromote deletes the handles it empties**
Files: `core/database/promote/unpromote.go`, `core/database/promote/unpromote_test.go`. Skill: `add-catalog-delete`.
Do: append `ReleaseEmptiedHandles(tx, append(released, bridges...))` to the returned changes; rewrite the doc comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/` passes, T-2.

**STEP-3 — Subject delete deletes the handles it empties**
Files: `core/database/subjects/subjects.go`, `core/database/subjects/delete_test.go`, `core/database/deleteimpact/conclusion_release_test.go`.
Do: D-2: call `ReleaseEmptiedHandles(tx, released)` right after `tx.Exec(sqlDelete, id)`. Leave `releasedHandles` as it is: `effects.Handles` already reads the deleted handles' ids from the claim changes, so `RecomputeTx` sees them. Return the handle changes between `released` and the subject change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/ ./core/database/deleteimpact/` passes, T-3, T-4.

**STEP-4 — The Unpromote preflight**
Files: `core/database/promote/unpromote.go` (`UnpromoteImpact`), `api/proto/engine.proto`, generated code, `api/ffi/dispatch.go`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/promote_test.go`. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-10, D-11; update the `UnpromoteSubjectRequest` comment; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/database/promote/` passes, T-5.

**STEP-5 — Swift store method**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `CoreInvoke.swift`, `FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`. Skill: `add-swift-test`.
Do: the store row of §6 macOS; rewrite `unpromoteSubject`'s doc ("deletes every handle the write leaves with no accepted member").
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) builds with no warning.

**STEP-6 — The confirms word the outcome**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `EvidenceCanvasCard.swift`, `EvidenceGraphView.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-12 – D-14. Existing tests that call `beginUnpromote` synchronously gain `await`; `unpromoteActionOpensConfirmNamingBridges` waits for `pendingUnpromote` with `waitUntil`.
Check: `xcodebuild test` passes, T-8 – T-10.

**STEP-7 — Docs**
Files: `docs/conclusion-layer-data-model.md` §5.2 (Unpromote deletes a handle it leaves with no accepted member, as Subject delete does), §6 ("Dropping the last member does not delete the entity or change its id." → the rule, with the merged-into exception), §10 invariant 7; `docs/catalog-deletes.md` item 9; `core/database/deleteimpact/README.md` (`ReleaseEmptiedHandles` is the one handle removal; `canonical_entities.last_member`); `.agents/skills/add-catalog-delete/SKILL.md` (the claim-releasing writers call `ReleaseEmptiedHandles`); `docs/deployment-plan/spike-13/deployment-plan.md` (map row scope: "and Subject delete"; scope entry; working rule "One Unpromote" gains "Subject delete too (user, 2026-11-04)"); `completed.md`; this plan's status.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-08-unpromoting-the-last-member.md` and `scripts/check-guidance-names.py` pass; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/deleteimpact/handle_release_test.go` `TestReleaseEmptiedHandles` *(new)* | table: claim released, handle has no other claim → row gone, one `canonical_entity` delete change with all seven fields; another accepted member → stays, no change; provisional claim left → handle and that claim deleted, two changes; merged-into → stays, no change, no error; non-claim changes in `released` → ignored; same handle twice → one delete | AC-1.1, AC-1.2, AC-1.5, D-3, D-4, D-5 |
| T-2 | Go | `core/database/promote/unpromote_test.go` `TestUnpromote` (new rows) | only member → handle gone and effects (through `writes.Run`) carry the id in `Handles` and `counts` in `Lists`; two members → stays; birth participation only on this graph → person and participation gone, Event stays; shared association (bridge on another graph too) → stays; Event and Place only members → gone | AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| T-3 | Go | `core/database/subjects/delete_test.go` `TestDeleteReleasesEmptiedHandles` *(new)* | delete the only member → handle gone in the same revision; bridge Subject that is a relationship's only member → relationship gone; Subject on a handle with another member → handle stays | AC-3.1, AC-3.5 |
| T-4 | Go | `core/database/deleteimpact/conclusion_release_test.go` `TestMemberDeleteEndToEnd` (new rows) | after an Unpromote of the only member: no FTS row for the handle, `graphcache` has no node, `conclusionsummaries.ListPersons` omits it, an `audit_transaction_scopes` row names the handle | AC-1.6 |
| T-5 | Go | `api/ffi/handlers/promote_test.go` `TestGetUnpromoteImpact` *(new)*, `runRPC` | only member → `deletes` via `canonical_entities.last_member`, kind `canonical_entity`, total 1, listed ref and location section `persons`; other member → empty; unpromoted → empty; merged-into → empty; malformed id → invalid; unknown Subject → `promote.invalid`; bad bytes → unmarshal error | AC-2.1, AC-2.2, AC-2.3 |
| T-6 | Go | `core/database/deleteimpact/handle_release_test.go` `TestLastMemberMatchesRelease` *(new)* | for each T-1 fixture, `LastMemberHandles` of the Subject lists exactly the handles `ReleaseEmptiedHandles` deletes after its claims are released | D-8 |
| T-7 | Go | `core/database/deleteimpact/impact_test.go` `TestSubjectCascadesSplitHandles` *(new)* | Subject with accepted claim on `PER-A` (only member) and provisional claim on `PER-B` (has members) → cascades `canonical_entities.last_member` [`PER-A`] and `identity_claims.subject_id` [`PER-B`]; their union is every handle the Subject has a claim on | AC-3.2, AC-3.3, D-9 |
| T-8 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteLastRecordNamesTheDeletedHandle` *(new)* | FakeStore with `s-james` the only member of its Person → `pendingUnpromote.deletedHandleRef` is the entity ref, `kind == "person"`; household where another Subject shares the entity → `nil` | AC-2.4, AC-2.5, D-12 |
| T-9 | Swift | `EvidenceGraphModelTests.swift` `unpromotePreflightFailureStaysInTheConfirm` *(new)* | `getUnpromoteImpactError` set → `pendingUnpromote` opens with `error` set and `deletedHandleRef == nil`; `confirmUnpromote` still records `unpromoteSubject id=s-james` | AC-2.6 |
| T-10 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` `cascadeLinesNameDeletedHandles` *(new)* | group `canonical_entities.last_member` [`PER-7KD45`] → `L10n.DeleteImpact.deletesHandle(handleRefs: "PER-7KD45")`; two refs → `deletesHandles`; `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` for `person`, `event`, `place`, `relationship` → the person, event, place, person sentences | AC-3.4, AC-2.4, D-13, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestFacetReleaseHonesty`, `TestPragmaHonestyProjectors`, `TestMemberDeleteEndToEnd`, `TestUnpromoteSubject`, the three existing `EvidenceGraphModelTests` unpromote tests (they become async).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: "deletes no handle", "No handle is deleted", "no delete path yet", "Handles have no delete path yet" and "Dropping the last member does not delete" `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/`, `api/proto/` and `macos/App/` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI: the two confirms keep their layout; only the messages of §6 Strings change
- [ ] Version bump: none (header)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Domain docs updated: conclusion §5.2, §6, §10 invariant 7; catalog-deletes.md item 9; deleteimpact README; add-catalog-delete
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
| `RecomputeTx` / `RecomputeReported` on a deleted id writes a stale row | T-4 asserts no reconciler or FTS row survives |
| The narrowed `identity_claims.subject_id` Count/List trips a honesty check that compares `Named` counts with `Remaining` | T-7 and `TestFacetReleaseHonesty`; see Stop conditions |
| Making `beginUnpromote` async lets a second click race the first | `pendingUnpromote` is set once after the fetch; a second open replaces it, as `beginDelete` |

**Stop and report if:**

- `releaseClaim`'s delete change no longer carries `entity_id`.
- `TestFacetReleaseHonesty` or another `deleteimpact` test requires a `Named` facet's Count to equal the rows its Release removes, so D-9's narrowed Count fails it.
- `autoreconciler.RecomputeTx`, `graphcache.OnCommit` or `searchindex` fails or leaves rows for a handle id with no row (T-4 cannot pass without changing them).
- `effects.Resolve` on a `canonical_entity` delete change does not yield `counts` in `Lists`.
- Adding `KindCanonicalEntity` to `deletableKinds` makes `TestPragmaHonesty` or an Impact test for another kind fail.
- A Swift caller of `beginUnpromote` other than `performCardAction` and `EvidenceCanvasCard` exists at the base.

## 12. Open questions

None.

## 13. PR

**Title:** S13-08 — Unpromoting the last member
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-08-unpromoting-the-last-member.md`.
