# S14-38 — Unpromote rewrites the ends of the links it takes back

**Status:** Blocked — re-plan after S14-42 and S14-43 (user, 2026-11-07: working rule "(Batch 8) The structural fix comes before the patch"). This plan is not built. Its scenarios (T-1 – T-6, T-10) move into S14-42 and S14-43 as acceptance tests; the mechanism it designs (D-1, D-2, D-9) is replaced by the read index ([design](../derived-read-index.md)).
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the search hits, lists and detail pages look the same; only stale text is rewritten)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-38-unpromote-stale-search-documents`
**Version bump:** patch — a researcher-visible bug fix in search results ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md) (the edit sits beside `HandleScope` on the `identity_claims` effects entry, which this PR does not change), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md) (`subjects.Delete` changes its linked-handle read)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [S14-33 ship log](S14-33-auto-file-bridges.ship.md) FU-1 and FU-4; S14-33 plan Landed note (5) and (8); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); [`core/writes/README.md`](../../../../core/writes/README.md); [`docs/catalog-deletes.md`](../../../catalog-deletes.md); working rule "(Batch 7) Creating a bridge files it when both ends are already handles"
**Planned:** 2026-11-07 against `main` @ `986d304`
**Review:** [`S14-38-unpromote-stale-search-documents.review.md`](S14-38-unpromote-stale-search-documents.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blockers; To user: bridge delete stops walking from the place end (M-1, D-6)

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

When the researcher unpromotes a Subject, every filed link it ends is taken back with it. Today the other end of each link keeps its old search document: unpromote James, and the omnibar still finds his birth as "Birth of James Robins" though the Events list and a rebuild say "Unspecified birth". After this PR, Unpromote rewrites the stored rows and the search document of every end a link was read through, in the same write, and names every end to the app so an open detail page reloads, a place end included, without walking from the place. Bridge delete reads the same ends through the same read, so it stops walking from a location's place end (S14-33 FU-4).

**Why now.** Found in the S14-33 ship loop (FU-1): 234 stale documents over 451 unpromotes in the probe. It predates S14-33. The planning probe reproduced it in both directions (§4).

**What stays.** Promote and filing behave as they do. Bridge delete still rewrites its ends; it no longer recomputes the end no row reads a link through (`location.place`, `place_relationship.to`), which has no row that changes. A bridge delete still does not name that end to the app (logged, §3 Out). The 6,000-event city stays as cheap as S14-33 left it.

## 2. User stories

**US-1.** As a researcher, when I unpromote a person or an event that a filed participation joins, the other end's search document and list row stop showing the link, so the omnibar and the lists agree with the pages.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | James Robins (named) and a birth dated 1885-05, both promoted, a filed participation (role subject). Unpromote James | The event's search document equals a `searchindex.RebuildAll`: title "Unspecified birth", not "Birth of James Robins" |
| AC-1.2 | The same world. Unpromote the birth | James's search document equals a rebuild: no `1885-05` in its secondary text |
| AC-1.3 | The same world, with the person and event summaries loaded into the display cache. Unpromote James (and, in a fresh world, the birth) | `Graph().Verify()` returns nil after the write |
| AC-1.4 | Unpromote James | The write's `Effects.Handles` holds the event's handle, so the app's `touchedHandles` reloads an open event page |

**US-2.** As a researcher, when I unpromote a place that a filed location or place relationship reads, the event or place that read it is rewritten.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A birth located at Leeds (both promoted, location filed). Unpromote Leeds | The birth's search document and stored rows equal a rebuild |
| AC-2.2 | Toronto part_of Ontario (both promoted, part_of filed) and a birth at Toronto. Unpromote Ontario | Toronto's and the birth's search documents equal a rebuild |
| AC-2.3 | York succeeded_by Toronto (both promoted, filed). Unpromote York; in a fresh world, unpromote Toronto | Search documents equal a rebuild and `Graph().Verify()` returns nil (planning probe: both pass with D-1; Predecessors and Successors are filled only on the detail read) |

**US-3.** As the engine, an Unpromote or a bridge delete at a crowded place costs what filing there costs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Leeds with 3 and with 12 other promoted births located there. Unpromote one further located birth | The write's `Effects.SummaryDeps` holds none of the other births, and `Effects.Handles` does not hold Leeds |
| AC-3.2 | Leeds with 3 other located births, their event documents marked stale. Delete one further birth's location bridge (`subjects.Delete`) | The other births' documents keep the mark (no walk from Leeds); the deleted link's event document is rewritten |
| AC-3.4 | Leeds with 3 other located births. Unpromote one further located birth | `Effects.SummaryDeps` holds Leeds (so an open Leeds page reloads) and `Effects.Handles` does not; no sibling birth is in either |
| AC-3.3 | `TestLargeSourceBudgets` fixture (city with 6,000 events). Unpromote an event located at the city | Under `budgetUnpromoteAtCity` (500 ms); 45 ms at planning, today and with the prototype of D-1 |

## 3. Scope

| In | Covered by |
| --- | --- |
| `connectrules.Endpoint.Unread` marks the ends no row reads a link through; `conclusionsummaries.linkEnds` is built from it | STEP-1, D-2 |
| One filtered read of a Subject's link-end handles in `effects` (`LinkEndHandles` and its SQL), used by the registry and by `subjects.Delete`; `subjects.linkedHandles` deleted | STEP-3, STEP-4, D-1, D-7 |
| A deleted accepted claim names the handles of its Subject's read link ends in `Set.Handles` | STEP-3, AC-1.1 – AC-2.3 |
| Bridge delete walks only from read ends | STEP-4, AC-3.2 |
| Every end of a link an Unpromote takes back, read or unread, is named to the app in `summary_handles` without a recompute or a walk (`Effect.Reload`, `Set.Reload`) | STEP-5, AC-3.4, D-9 |
| Tests: rebuild-equals-upkeep scenarios, an effects registry subtest, display-cache and walk tests, a budget | STEP-2, STEP-6, §8 |
| Doc comments and the effects, writes and catalog-deletes docs | STEP-7 |

| Out | Home |
| --- | --- |
| Reprojection cost when the whole of a part_of with a crowded part is unpromoted (the part's rows do move) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the existing row "The first `part_of` filed from a city reprojects its events" |
| Moving `subjects.Delete`'s recompute onto `writes.Run` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the existing row "catalog graph cache leftovers": the bridge's Observations are deleted before `effects.Resolve` runs |
| `(*bridgeWorld).unpromote` in `core/database/promote/unpromote_test.go` moving onto the new `promotetest.Unpromote` | not planned: working test code this PR does not otherwise touch |
| Naming a deleted bridge's unread end (a place) to the app after `subjects.Delete` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md): new row, STEP-8. At `effects.Resolve` the bridge's claim and Observations are gone, so no registry path can tell a filed bridge from an unfiled one; an `observations` delete arm would reload the Persons, Events and Places lists on every unfiled-bridge delete (review M-13). Bridge delete did not name it before either |
| Moving the existing private bridge builders (D-7) onto `connecttest` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md): new row, STEP-8 |
| Bridges in the seeded rebuild-equals-upkeep generator | not planned: the named scenarios hold each shape |

## 4. Current state

Verified at `986d304`.

| Area | Where | Today |
| --- | --- | --- |
| Unpromote | `core/database/promote/unpromote.go` `Unpromote` | Releases the Subject's accepted claim (`deleteimpact.ReleaseClaim`), then `identityclaims.UnfileBridges`. `writes.Run` recomputes `effects.Resolve(changes).Handles` |
| Unfiling | `core/database/identityclaims/bridgefile.go` `UnfileBridges` | Releases each filed bridge's accepted claim. The bridge Subject and its Observations stay |
| Handle delete | `core/database/deleteimpact/handle_gone.go` `releaseHandleIfGone` | Deletes an emptied handle's `canonical_entities` row. `auto_reconciler_values.entity_id` is `ON DELETE CASCADE` (`core/database/migrations/000034.sql:13`), so the association's value rows go inside the closure |
| Claim delete change | `core/database/deleteimpact/facets.go` `releaseClaimRow` | The `identity_claim` delete change carries `subject_id`, `entity_id` and `status` in its old fields, so `onStatus` and `from("subject_id", …)` work on it |
| Claim effects | `core/database/effects/registry.go` `"identity_claims"` (line 441) | `Handles: union(field("entity_id"), onStatus("accepted", from("subject_id", observers)))`. `observers` (line 302) walks inbound. Nothing reaches the handles a bridge Subject points at |
| Path helpers | `core/database/effects/path.go` `sqlPath`, `from`, `onDelete`, `onStatus`, `union`, `queryIDs` | `sqlPath(query)` binds one `?` per input id (`queryIDs(w.q, query, id)`); fed by `from(col, …)` it gets the column's ids, as `sqlHandlesForSource` does at `registry.go:334` |
| Recompute | `core/database/autoreconciler/autoreconciler.go` `RecomputeReported` | Calls `conclusionsummaries.AssociationEndpoints` before and after `recomputeBatch`; both run after the closure, so a deleted association yields no ends |
| Ends | `core/database/conclusionsummaries/dependents.go` `AssociationEndpoints`, `linkEnds` (line 156) | `linkEnds` is a literal map built from `hops.PlacesOfEvent` and `hops.ParentsOfPlace` (`.Bridge()`, `.To()`); its use is at line 141 |
| Bridges | `core/connectrules/connectrules.go` `Endpoint`, `productBridges`, `Bridges` | `Endpoint{PropertyKey, TypeKey}`. `Register` has no caller outside tests. `connectrules` imports only `vocabkey` |
| Bridge delete | `core/database/subjects/subjects.go` `Delete` (lines 205–278), `linkedHandles` (line 353) | `linkedHandles`: the accepted handles of the Subjects this one's Observations point at, unfiltered. `Delete` adds them to `touched` and walks `Dependents` from them, so a location bridge delete walks from the place end (`docs/ideas/structural-debt.md` row "`subjects.Delete` calls `Dependents` directly") |
| Effects set | `core/database/effects/registry.go` `Set`, `Effect`, `(Effect).paths`; `resolve.go` `Resolve` | `Set.SummaryDeps` is filled only by `Run` from the recompute walk; every id in `Set.Handles` is a recompute root and a walk root. No field names a handle to the app alone. The `observations` entry's `Handles` is `observationMembers` (the Observation's own Subject's handles), so a deleted subject-valued Observation names nothing on its value end |
| App | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.conclusionDetail` | Reloads a detail page whose entity id is in `touchedHandles` (`handles` + `summary_handles`); a place page is a `.conclusionDetail` |
| Writes | `core/writes/run.go` `Run` | `effects.Resolve` → `audit.Record` → `RecomputeReported(set.Handles)` → `set.SummaryDeps`; `api/ffi/handlers/write_effects.go` sends both; Swift `CatalogWriteEffects.touchedHandles` is both |
| Display cache | `core/database/graphcache/commit.go` `detachAssociation`, `dropCommittedDisplays` | A loaded association is detached and its ends reloaded; the AC-1.3 shape passed in the planning probe today |
| Place rows | `core/database/conclusionsummaries/places.go` `PlaceSummary` | `PartOf`, `Contains`, `Predecessors`, `Successors` are filled only for detail (`AttachPlaceRelationships`), not stored |
| Tests | `core/database/autoreconciler/autoreconciler_test.go` `TestRebuildEqualsUpkeep_Scenarios`, `assertUpkeepEqualsRebuild`, scenario "promoting both ends of a participation files it, then the bridge is deleted" (line 1355) | Compare value rows and person/event/place documents with a rebuild. No Unpromote scenario. The fixture has `subject`, `cite`, `nameIn`, `termIn`, `dateIn`, `place`, `mustProp`, `promote`; no location or place-relationship helper |
| Tests | `core/database/conclusionsummaries/walks_test.go` `participation`, `location`, `named`, `at`, `promoteSubject`; `placechains_test.go` `placeSubject`, `placeRel`; `filedlinks_test.go` `promotedBirth`, `countIn`, `staleEventDocs`, `eventDocsStale` | Helpers T-5 – T-8 use |
| Tests | `core/database/effects/conclusion_handles_test.go` `TestConclusionHandleSets` (subtest "accepted mint covers claim handle, filings, and observers", line 229) | The registry-level test of `identity_claims.Handles` |
| Tests | `core/database/promote/promotetest/promotetest.go` `Save`, `ClaimTx`, `ClaimWithoutFiling` | Shared promote helpers; imports `core/writes`. No Unpromote helper |
| Tests | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `budgetLocateAtCity`, `fastest`, `over`; `promotealign_test.go` `birth`, `locate`, `promote` | The S14-33 budgets |

**Probe at planning** (deleted after): unpromote the person → upkeep event doc `Birth of James Robins`, rebuild `Unspecified birth`. Unpromote the event → upkeep person doc keeps `1885-05`. Bridge delete of the same world → equal. With the D-1 prototype: all equal; succeeded_by and part_of, either end unpromoted → equal and `Verify` nil.

**Precedent.** `subjects.linkedHandles` is the read; S14-33 F-2 (`AssociationEndpoints`, `linkEnds`) is the rule for which ends to walk. This PR puts the rule on the bridge declaration and the read in the effects registry, and points `subjects.Delete` at it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `core/database/effects` owns one read: `sqlLinkEndHandles` *(new, unexported const built at init)* and `func LinkEndHandles(q *sql.Tx, subjectID []byte) ([][]byte, error)` *(new)* — the handles (claims accepted or provisional) of the `value_subject_id`s of the Subject's Observations, leaving out an Observation whose Subject type and Property are an unread end (D-2). The registry path `linkEnds` *(new)* is `sqlPath(sqlLinkEndHandles)` | The bridge Subject and its Observations outlive an Unpromote, so `effects.Resolve` reads the ends after the closure. One SQL text serves Unpromote and bridge delete, so the walk rule cannot diverge between them (review M-1) | (a) snapshot in `Unpromote` and call `RecomputeTx` there — a second hand-path off `writes.Run`; (b) fix `AssociationEndpoints` — the rows are gone by CASCADE; (c) every end, unfiltered — prototype: 10,004 `SummaryDeps` and 161 ms for one Unpromote at the city, against 0 and 45 ms; (d) a new path beside `linkedHandles` — two copies of one read |
| D-2 | `connectrules.Endpoint` gains `Unread bool` *(new)*: "no stored row reads a link through this end". `productBridges` sets it on `location`'s `place` and `place_relationship`'s `to`. `conclusionsummaries.linkEnds` is built from `connectrules.Bridges()` (bridge key → unread end's Property key); the `hops` import stays for the walks | The fact belongs on the endpoint declaration every bridge reader already uses; it applies to the whole `place_relationship` bridge (part_of and succeeded_by), which a part_of hop cannot say (review M-2) | A list in `hops` built from filtered hops — a second declaration of bridges, and term-scoped by implication |
| D-3 | The registry arm is `onDelete(onStatus("accepted", from("subject_id", linkEnds)))`, appended as the third argument of the `identity_claims` `Handles` union. Exact expression: `union(field("entity_id"), onStatus("accepted", from("subject_id", observers)), onDelete(onStatus("accepted", from("subject_id", linkEnds))))` | A create or update keeps the association's rows, so `AssociationEndpoints` "after" already names its ends (S14-33). Running on filing would add the ends as roots on every connect. `from("subject_id", …)` feeds subject ids to the `sqlPath`; the claim id (`fromChange`) would match no row | `onStatus` alone — doubles the work on every filing |
| D-4 | `HandleScope` stays `field("entity_id")` | An end's last activity does not move because another Subject's Unpromote took a link back | Adding the ends — moves the "Updated" sort for a write the researcher did not make on that end |
| D-5 | The unread-end exclusion is written into the SQL text at package init from `connectrules.Bridges()`, as single-quoted literals: one line per unread end, `AND NOT (st.origin = 'provenencia' AND st.key = 'location' AND p.key = 'place')`. A key containing `'` panics at init. `path.go` does not change | `sqlPath` binds only the input id; the keys are product constants from code, not catalog data (review X-5, M-5) | Bound `?` arguments — needs a `sqlPath` variant for one caller |
| D-6 | `subjects.Delete` calls `effects.LinkEndHandles` where it called `linkedHandles`, for both `touched` and the `Dependents` input. `linkedHandles` is deleted. Provisional members now count, as they do in every registry path (`membersOf`) | One read; the place end of a deleted location is no longer walked from (FU-4) or recomputed: its rows do not read the link (AC-2.3 probe, S14-33 F-2) | Keeping `linkedHandles` for Delete — the second copy M-1 names |
| D-7 | Test helpers. `promotetest.Unpromote(t testing.TB, c *database.Catalog, userID, subjectID []byte) writes.Result` *(new)* runs `promote.Unpromote` in `writes.Run` with action `unpromote_subject` and fails the test on error. A new package `core/database/connect/connecttest` *(new)* holds `Participation(t, c, userID, artifactID []byte, person, event subjects.Subject, roleKey string) subjects.Subject`, `Location(t, c, userID, artifactID []byte, event, place subjects.Subject) subjects.Subject` and `PlaceRelationship(t, c, userID, artifactID []byte, from, to subjects.Subject, termKey string) subjects.Subject` *(new)*: each one `connect.CreateCitedBridge` through `writes.Call` (action `create_cited_bridge`, locator `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`), Properties from `properties.Lookup(c, key, vocabkey.OriginProvenencia)` (`person`, `event`, `role`; `event`, `place`; `from`, `to`, `place_relationship_type`), terms from `propertyterms.Lookup(c, prop.ID, key, vocabkey.OriginProvenencia)`, returning the bridge Subject. The autoreconciler scenarios use them; every world creates its bridges first and then promotes the ends, which files them, as the line-1355 scenario does | Five location builders and three place-relationship builders already live in test packages (review M-6: `conclusionsummaries` `location`/`placeRel`, `promotealign` `locate`/`placeRel`, `promote` `(*bridgeWorld).location`/`placeRelationship`, `grapheventtitles` and `graphcache` `location`); the repo's convention for shared test helpers is a `…test` package (`promotetest`, `subjectstest`) | A sixth and fourth private copy in `autoreconciler_test.go` |
| D-8 | `budgetUnpromoteAtCity` *(new)* = 500 ms | Same order as `budgetLocateAtCity`; 45 ms on the planning machine; CI is about 3× slower | A tighter figure — flakes on CI |

| D-9 | The effects registry gains one job: `Effect.Reload Path` and `Set.Reload [][]byte` *(new)*, "handles whose app pages a change moves; not recomputed, not walked; `Run` folds them into `SummaryDeps`". One entry sets it: `identity_claims`, `Reload: onDelete(onStatus("accepted", from("subject_id", valueMembersOfSubject)))`. `valueMembersOfSubject` *(new)* is a named var beside `observers`: `chain(inbound("observations", "subject_id"), field("value_subject_id"), membersOf)`, doc "the handles a Subject's Observations point at; `linkEnds` is this minus the unread ends of `connectrules`". `(Effect).paths` includes `Reload`; `Resolve` appends it after `HandleScopes`. `writes.Run`, after the `len(set.Handles) > 0` block: `set.SummaryDeps = database.UniqueBlobIDs(append(set.SummaryDeps, set.Reload...))`. `SummaryDepsWalked` and `core/database/graphcache` do not change: a write with a deleted accepted claim always has `Handles` (its `field("entity_id")`), so the graph cache drops `SummaryDeps` displays as today, and an unread end's display has no stored row that reads the link. The wire is unchanged: `summary_handles` carries `SummaryDeps` | The user asked to fix the cached place page in this PR (2026-11-07). Every id in `Handles` is a walk root, so a place there walks its events; `SummaryDeps` is the app's "reload these" list. Prototype at planning (with an extra `observations` arm since dropped): Leeds in `SummaryDeps` and not in `Handles` after an Unpromote, no sibling in either, `Graph().Verify()` nil; effects, writes, promote, subjects, autoreconciler, conclusionsummaries, identityclaims, graphcache, catalogsession, `api/ffi/...` green; S14-33 budgets under budget. Named for its job, as `Handles` and `HandleScope` are (review M-10) | (a) The place in `Handles` — walks 6,000 events (161 ms, 10,004 ids); (b) a filtered SQL for unread ends only — a second query where duplicates cost nothing; (c) a new wire field — `summary_handles` already means "reload these"; (d) an `observations` delete arm for bridge delete — reloads all three lists on an unfiled-bridge delete (Out) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `connectrules.Endpoint.Unread` *(new)* | field | `conclusionsummaries.linkEnds` | Replaces it as the source; `linkEnds` becomes a read of it |
| `effects.LinkEndHandles`, `sqlLinkEndHandles`, path `linkEnds` *(new)* | func / const / path | `subjects.linkedHandles` (deleted), `observers` (the inbound direction) | Moves `linkedHandles` to the package both callers import, with the filter |
| `promotetest.Unpromote` *(new)* | helper | `promotetest.Save`; `(*bridgeWorld).unpromote` (private to `promote_test`) | Extends `promotetest` |
| `connecttest.Participation`, `Location`, `PlaceRelationship` *(new)* | helper package | five private location and three place-relationship builders (D-7) | The first shared home; the existing copies move later (Out) |
| `Effect.Reload`, `Set.Reload` *(new)*; `valueMembersOfSubject` *(new)* | registry field; path | `Set.SummaryDeps` (filled by `Run`), `Effect.Handles` (recompute set); `observers` (the inbound chain) | `Handles` makes a walk root; `SummaryDeps` has no registry path. The chain is named once beside its inbound twin, and `linkEnds` is documented as its filtered form (review M-11) |
| Tests in §8 | test | the line-1355 scenario, `TestFiledLocationWalksOnlyFromItsEvent`, `TestConclusionHandleSets` | Extend those suites |
| `budgetUnpromoteAtCity` *(new)* | const | `budgetLocateAtCity` | One more row |

**Builds on.** The effects registry and its path helpers; `connectrules` bridge declarations; S14-33's walk-end rule.
**Deletes.** `subjects.linkedHandles`. The "unread end not named to the app" gap S14-33 F-2 closed for filing only. The literal body of `conclusionsummaries.linkEnds`. The structural-debt row "`subjects.Delete` calls `Dependents` directly, so deleting a location bridge still walks from the place end" (fixed here). The words "linked handles" in the `RecomputeTx` doc comment (`autoreconciler.go:190`).
**Next adds.** A new product bridge in `productBridges` with an end no row reads through sets `Unread` on that endpoint; filing, Unpromote and bridge delete honour it. `linkEnds` and the SQL are built once at init, so a bridge added later through `connectrules.Register` is not covered (Stop condition), and `linkEnds` holds one unread end per bridge.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Unpromote names the remaining read-end handles in `Handles` | Swift `touchedHandles` → `conclusionDetail` refresh by id | The end still exists; its page reloads as after any write naming it |
| A deleted claim whose Subject is deleted in the same batch (`subjects.Delete`) | path `linkEnds` | The Observations are gone; it returns nothing. `Delete` reads `LinkEndHandles` itself before releasing |
| A deleted claim of a non-bridge Subject | path `linkEnds` | Returns the handles of that Subject's subject-valued Observations, if any: each is a handle the Subject pointed at, so a recompute is correct; one handle each |
| A bridge delete no longer recomputes its unread end | `subjects.Delete` | That end's rows and document do not read the link (AC-2.3); the app is not told about it, as before (Out) |
| An Unpromote names every end of each link it takes back in `summary_handles` | Swift `touchedHandles` → `.conclusionDetail`; the Persons, Events and Places lists (`applyHiddenPolicyOnAny(follows: .touchedHandles)`) | Those lists already reload on an Unpromote (its `Handles` is never empty), so the added ids reload nothing new; an open place page now reloads |

### Scale

At the map's Scale (a project of 3,000 Sources, 37,000 Subjects and 92,000 Observations; 6,000 events on one place), measured on the `TestLargeSourceBudgets` fixture (4,000 persons, 8,000 events, one city with 6,000): one Unpromote adds one indexed query (`observations.subject_id`) per deleted claim. At the city it adds no root: prototype 45 ms, 0 `SummaryDeps`, as today. A person with N filed participations adds N event roots, each walked one hop. Unpromoting a part_of's whole adds the part as a root and walks its descendants' events, as filing that part_of does (Out row 2). Bridge delete of a location at the city stops walking the city's 6,000 events.

### Edges

- **Assumptions.** `effects.Resolve` runs after the closure (`run.go`), so a taken-back bridge's Observations are present. Ends are unique per write (`UniqueBlobIDs` in `RecomputeReported`), so a self-link or two links to one end recompute it once. T-1 – T-6 hold it.
- **Neighbors.** Filing: unchanged (D-3); `TestFiledLocationWalksOnlyFromItsEvent`, `TestFiledPartOfWalksFromThePart` stay green. Unpromote of a Subject with no filed bridges: nothing added (`TestUnpromote` "a Subject with no filed bridges loses only its own claim"). succeeded_by: unpromoting `from` leaves `to` untouched and nothing stale (AC-2.3, T-4). A reject or a later path deleting a filed bridge claim gains the same ends (intended).
- **Removed inputs.** The unpromoted Subject's handle: deleted when last member (`Set.Deleted`), else recomputed by `field("entity_id")`. The association: deleted, or recomputed when a second record keeps it. Read ends: recomputed and rewritten (this PR). Unread ends: nothing stored changes; named to the app through `Set.Reload` after an Unpromote (D-9); not after a bridge delete (Out).
- **Real data.** Bridges today: participation (person, event), relationship (person, related_to), location (event, place*), place_relationship (from, to*); * = `Unread`.
- **Untrusted size.** None — catalog rows the app wrote, bounded by one Subject's Observations.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/connectrules` | `Endpoint.Unread bool` *(new)*, doc comment; set on two endpoints in `productBridges` |
| `core/database/conclusionsummaries` | `linkEnds` initialised from `connectrules.Bridges()` (`map[string]string`, bridge key → the `Unread` endpoint's Property key), same name and use. `AssociationEndpoints` doc: a deleted association has no rows (CASCADE); its read ends come from the deleted claim through `effects` |
| `core/database/effects` | `sqlLinkEndHandles` (built in a `func linkEndHandlesSQL() string` called at package var init), `LinkEndHandles`, path `linkEnds`, the D-3 arm |
| `core/database/effects` (D-9) | `Effect.Reload`, `Set.Reload` *(new)* with doc comments in the style of their neighbours; `valueMembersOfSubject`; the `identity_claims` entry; `paths()`; `Resolve`. `Set.SummaryDeps` doc becomes: "SummaryDeps are handles whose stored rows read Handles, from the walk RecomputeTx already did, plus Reload. Run fills it. SummaryDepsWalked says that walk happened, including when it found nobody, so a commit drops those rows and does not walk again." |
| `core/writes` | `Run` folds `set.Reload` into `set.SummaryDeps` (D-9) |
| `core/database/subjects` | `Delete` uses `effects.LinkEndHandles`; `linkedHandles` deleted |
| `core/database/autoreconciler` | `RecomputeReported` comment only |
| `core/database/promote/promotetest` | `Unpromote` *(new)* |

```sql
SELECT DISTINCT ic.entity_id FROM observations o
JOIN subjects s ON s.id = o.subject_id
JOIN subject_types st ON st.id = s.subject_type_id
JOIN properties p ON p.id = o.property_id
JOIN identity_claims ic ON ic.subject_id = o.value_subject_id
  AND ic.status IN ('accepted', 'provisional')
WHERE o.subject_id = ?
  AND NOT (st.origin = 'provenencia' AND st.key = 'location' AND p.key = 'place')
  AND NOT (st.origin = 'provenencia' AND st.key = 'place_relationship' AND p.key = 'to')
-- the NOT lines are generated from connectrules.Bridges() (D-5)
```

Errors: none new.

### FFI

None — `WriteEffects.handles` already carries `Set.Handles`.

### macOS

None — `touchedHandles` already follows `handles` and `summary_handles`. One doc comment changes (`CatalogLastUpdated.swift` `.touchedHandles`, STEP-7); no Swift code.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Mark the unread ends**
Files: `core/connectrules/connectrules.go`, `core/database/conclusionsummaries/dependents.go`.
Do: add `Endpoint.Unread` and set it (D-2). Build `linkEnds` from `connectrules.Bridges()` with a package-level `var linkEnds = unreadEnds()` and a small func; keep its name and its use at line 141. Remove the `hops` lines from the map only.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/connectrules/... ./core/database/conclusionsummaries/...` passes (no behavior change).

**STEP-2 — Helpers and failing tests**
Files: `core/database/promote/promotetest/promotetest.go`, `core/database/connect/connecttest/connecttest.go` *(new)*, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/conclusionsummaries/filedlinks_test.go`, `core/database/effects/conclusion_handles_test.go`.
Do: add `promotetest.Unpromote` and `core/database/connect/connecttest/connecttest.go` *(new)* (D-7), with a package doc comment. Add T-1 – T-8 (§8). In T-1 – T-4 name the person with `nameIn(p, f.props["name"], "given=James|surname=Robins")`, type the event with `termIn(e, f.props["event_type"], f.terms["birth"])` and date it with `dateIn(e, f.props["date"], 1885, ip(5), nil)`, through `f.cite`; name places with `textIn(pl, f.props["toponym"], "Leeds")`. Each T-5 / T-6 row builds a fresh world.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/ ./core/database/conclusionsummaries/ ./core/database/effects/ -run 'Scenarios|Unpromoted|DeletedLocation|TakenBack|ConclusionHandleSets'` — T-1, T-2, T-3, T-4's part_of rows, T-5's `Handles` assertion, T-7's positive case and T-10 fail; T-4's succeeded_by rows, T-6, T-8 and T-7's negative cases pass (they pin what must not regress). T-4 succeeded_by fails today only if its `Verify` or document check finds a stale row; record the result.

**STEP-3 — The deleted claim names its read link ends**
Files: `core/database/effects/registry.go`.
Do: add `linkEndHandlesSQL`, `sqlLinkEndHandles`, `LinkEndHandles` and path `linkEnds` beside `observers`, each with a doc comment; append the D-3 arm.
Check: the STEP-2 command passes except T-8 (bridge delete, STEP-4) and T-10 and T-7's `Reload` assertion (STEP-5); `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/database/promote/ ./core/database/identityclaims/` passes.

**STEP-4 — Bridge delete reads the same ends**
Files: `core/database/subjects/subjects.go`. Skill: `add-catalog-delete`.
Do: replace the `linkedHandles` call with `effects.LinkEndHandles(tx.Tx, id)`; delete `linkedHandles`; update `Delete`'s doc comment ("linked handles" → "the handles of its read link ends").
Check: T-8 passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/ ./core/database/autoreconciler/` passes, including the line-1355 scenario.

**STEP-5 — Name every end to the app**
Files: `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/writes/run.go`.
Do: D-9, with the `Set.SummaryDeps` doc text in §6 Go core. Do not change `SummaryDepsWalked` or `core/database/graphcache`.
Check: T-10 and T-7's `Reload` assertion pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/writes/ ./core/database/conclusionsummaries/ ./core/database/promote/ ./core/database/subjects/ ./core/database/autoreconciler/ ./core/database/graphcache/... ./api/ffi/...` passes.

**STEP-6 — Budget**
Files: `core/database/promotealign/perf_test.go`.
Do: add `budgetUnpromoteAtCity` (D-8) with a comment in the style of `budgetLocateAtCity`, and T-9 as the last subtest: 3 births from `f.birth(lf.onePlace, lf.pageArt)`, each promoted with `f.promote(e)` and then located at `lf.citySubject` with `f.locate(lf.onePlace, lf.pageArt, e, lf.citySubject)` (so the location files), as the "locate an event at the city" subtest does; time `promotetest.Unpromote(t, f.c, userID, e.ID)` with `fastest(3, …)` and `over(…)`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` logs the new line under budget and every S14-33 budget under its own.

**STEP-7 — Comments and docs of the read**
Files: `core/database/conclusionsummaries/dependents.go` (`AssociationEndpoints` doc: after "A save calls it before rewriting those rows and again after." add "A deleted association has no rows by then (CASCADE); the effects registry names its read ends through the deleted claim."), `core/database/autoreconciler/autoreconciler.go` (`RecomputeTx` doc: "linked handles" → "the handles of its read link ends"), `core/database/effects/README.md` (a deleted accepted claim names the handles of its Subject's read link ends, through `LinkEndHandles`, which Subject delete also reads; an end is unread when its `connectrules.Endpoint` says so; and `Effect.Reload`: handles whose app pages a change moves, which `Run` folds into `SummaryDeps` with no recompute or walk; a deleted accepted claim names every end of its Subject's links there; `valueMembersOfSubject` is the outbound twin of `observers`), `core/database/graphcache/commit.go` (`dropCommittedDisplays` doc: "When the write already walked dependents, those ids, and the handles the write named for the app (`Set.Reload`), are dropped too and the walk is not repeated."), `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` (`.touchedHandles` doc: "every handle a write changed, whose summary it changed, or whose page it moved (`Set.Reload`)"), `core/writes/README.md` (Run's steps: after the recompute, `Run` adds `Set.Reload` to the summary handles), `core/writes/README.md` and `docs/catalog-deletes.md` ("linked handles" → "the handles of its read link ends (`effects.LinkEndHandles`)").
Check: `gofmt -l core` prints nothing; `rg -n 'linkedHandles|linked handles' core docs .agents .cursor` finds nothing.

**STEP-8 — Version and docs**
Files: `VERSION` and `core/version.go` (0.4.0 → 0.4.1), `macos/Provenencia.xcodeproj/project.pbxproj` (all four `MARKETING_VERSION` lines), `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, this plan.
Do: the patch bump per `docs/versioning.md` (use the next patch after `main`'s `VERSION` if it moved). In `structural-debt.md` delete the row that starts "`subjects.Delete` calls `Dependents` directly" and add two rows: "— | S14-38 review M-13 | A bridge delete (`subjects.Delete`) does not name the link's unread end (a place) to the app, so an open place page that lists the link stays stale until it reloads. At `effects.Resolve` the bridge's claim and Observations are gone; reporting it needs `subjects.Delete` to hand ids to the write's effects, or a path that knows the bridge was filed." and "— | S14-38 review M-6 | Private bridge builders in test packages (`conclusionsummaries` `location`/`placeRel`, `promotealign` `locate`/`placeRel`, `promote` `(*bridgeWorld).location`/`placeRelationship`, `grapheventtitles` and `graphcache` `location`) move onto `connecttest` when those tests are next touched." Then the DoD lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-38-unpromote-stale-search-documents.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestRebuildEqualsUpkeep_Scenarios` "a filed participation, then the person is unpromoted" *(new)* | James + birth 1885-05, participation filed, unpromote James → `assertUpkeepEqualsRebuild` passes | AC-1.1, D-1 |
| T-2 | Go | same, "a filed participation, then the event is unpromoted" *(new)* | unpromote the birth → passes | AC-1.2 |
| T-3 | Go | same, "a filed location, then the place is unpromoted" *(new)* | birth located at Leeds, both promoted; unpromote Leeds → passes | AC-2.1, D-2 |
| T-4 | Go | same, table of four scenarios "a filed part_of / succeeded_by, then the whole / from / to is unpromoted" *(new)* | Toronto part_of Ontario + a birth at Toronto, unpromote Ontario; York succeeded_by Toronto, unpromote York; and unpromote Toronto → each passes | AC-2.2, AC-2.3 |
| T-5 | Go | `core/database/conclusionsummaries/filedlinks_test.go` `TestUnpromotedLinkEndsInDisplayCache` *(new)* | rows {unpromote person, unpromote event}, fresh world each: `PersonsByIDs` and `EventsByIDs` loaded, `promotetest.Unpromote` → `f.c.Graph().Verify()` nil and `res.Effects.Handles` holds the remaining end; row {York succeeded_by Toronto, `PlacesByIDs` loaded, unpromote York} → `Verify` nil | AC-1.3, AC-1.4, AC-2.3 |
| T-6 | Go | same file, `TestUnpromotedLocationWalksOnlyFromItsEvent` *(new)* | rows {3, 12} sibling births at Leeds (`placeSubject("Leeds", nil, nil, 2, 0)`, its handle from `promoteSubject`); one further `promotedBirth(leeds)`; unpromote that birth → `countIn(res.Effects.SummaryDeps, siblings...) == 0` and `!database.ContainsBlobID(res.Effects.Handles, leedsID)` | AC-3.1, D-1, D-3 |
| T-7 | Go | `core/database/effects/conclusion_handles_test.go` `TestConclusionHandleSets` "deleted accepted bridge claim covers its read ends" *(new)* | in `open()`: `claim(tx.Tx, birth, nil, identityclaims.StatusAccepted, true)`, then `claim(tx.Tx, bob, nil, identityclaims.StatusAccepted, true)` (files the existing bob–birth participation; keep its changes as `filed`). Read the bridge claim id: `SELECT ic.id FROM identity_claims ic JOIN observations o ON o.subject_id = ic.subject_id WHERE o.value_subject_id = ? AND ic.status = 'accepted' AND ic.entity_id IN (assocs)` with bob's id. `deleteimpact.ReleaseClaim(tx, id)` → `effects.Handles(tx.Tx, released)` covers the birth's and bob's handles, and `effects.Resolve(tx.Tx, released)` has `Reload` covering both. Negative (D-3): `effects.Handles(tx.Tx, filed)` does not hold the birth's handle (a claim create names no link end) | D-1, D-3, D-9 |
| T-8 | Go | `core/database/conclusionsummaries/filedlinks_test.go` `TestDeletedLocationWalksOnlyFromItsEvent` *(new)* | 3 sibling births located at Leeds, a further birth located there (`promotedBirth(leeds)`); its bridge Subject id read with `SELECT s.id FROM subjects s JOIN subject_types st ON st.id = s.subject_type_id AND st.key = 'location' JOIN observations o ON o.subject_id = s.id WHERE o.value_subject_id = ?` (the birth's Subject id), so `location` and `bridge` do not change; `staleEventDocs()`; `subjects.Delete` of that bridge through `writes.Call` with action `delete_subject` → the siblings' documents keep `display_subtitle = 'stale'`, the further birth's does not | AC-3.2, D-6 |
| T-10 | Go | `core/database/conclusionsummaries/filedlinks_test.go` `TestTakenBackLinkNamesItsPlace` *(new)* | Leeds made and promoted as in T-6 (`placeSubject("Leeds", nil, nil, 2, 0)`, `promoteSubject`), 3 located sibling births and one further `promotedBirth(leeds)`; `res := promotetest.Unpromote(t, f.c, userID, birth.ID)` → `res.Effects.SummaryDeps` holds Leeds, `res.Effects.Handles` does not, `countIn` of the siblings in either is 0, `f.c.Graph().Verify()` nil | AC-3.4, D-9 |
| T-9 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "unpromote an event at the city" *(new)* | fastest of 3 → under `budgetUnpromoteAtCity` | AC-3.3, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/effects` (`TestHandlePaths`, `registry_completeness_test.go`, `TestConclusionHandleSets`), `core/database/subjects`, `core/database/conclusionsummaries` `TestFiled*`, `core/database/promote` `TestUnpromote*`, the S14-33 budgets.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`linkedHandles`, "linked handles"); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): patch
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-deletes.md (STEP-7)
- [ ] docs/ideas/structural-debt.md: the FU-4 row deleted, the STEP-8 row added
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a dogfood project. In an Evidence graph, make a person (name "James Robins") and an event (type Birth, date May 1885), connect them with a participation (role Subject). Promote both.
2. Search the omnibar for "Birth": the hit reads "Birth of James Robins".
3. Unpromote James. Search "Birth" again: the hit reads "Unspecified birth", the Events list row agrees, and the open event page no longer lists James.
4. Promote James again (the participation files), then unpromote the event. Search "James": the person hit shows no 1885 date.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The new arm fires on a path other than Unpromote and widens its recompute | D-3; T-6, T-7 negative cases, the S14-33 budgets |
| Bridge delete misses an end it used to recompute | Only unread ends drop (D-6); T-8, the line-1355 scenario |

**Stop and report if:**

- `RecomputeReported`, `AssociationEndpoints` or `subjects.Delete` no longer match §4.
- T-1 or T-2 passes before STEP-3 (the bug is fixed elsewhere; report the commit).
- Any S14-33 budget or T-9 goes over after STEP-3 or STEP-4.
- `effects` cannot import `core/connectrules`, or `subjects` cannot call `effects.LinkEndHandles`, without an import cycle.
- `registry_completeness_test.go` or a test that lists `Effect` fields needs a rule for `Reload` that D-9 does not give.
- `TestHandlePaths` or `registry_completeness_test.go` fails on the new `sqlPath` and the fix needs a change to `path.go`.
- A test registers a bridge through `connectrules.Register` after `effects` init and needs its unread end excluded.

## 12. Open questions

None.

## 13. PR

**Title:** S14-38 — Unpromote rewrites the ends of the links it takes back
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (T-1 – T-10 output); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-38-unpromote-stale-search-documents.md`.
