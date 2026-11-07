# S14-38 — Unpromote rewrites the ends of the links it takes back

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the search hits, lists and detail pages look the same; only stale text is rewritten)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-38-unpromote-stale-search-documents`
**Version bump:** patch — a researcher-visible bug fix in search results ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md) (the edit sits beside `HandleScope` on the `identity_claims` effects entry, which this PR does not change)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [S14-33 ship log](S14-33-auto-file-bridges.ship.md) FU-1; S14-33 plan Landed note (5) and (8); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); [`core/writes/README.md`](../../../../core/writes/README.md); working rule "(Batch 7) Creating a bridge files it when both ends are already handles"
**Planned:** 2026-11-07 against `main` @ `986d304`
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

When the researcher unpromotes a Subject, every filed link it ends is taken back with it. Today the other end of each link keeps its old search document: unpromote James, and the omnibar still finds his birth as "Birth of James Robins" though the Events list and a rebuild say "Unspecified birth". After this PR, Unpromote rewrites the stored rows and the search document of every end a link was read through, in the same write, and names those ends to the app so an open detail page reloads.

**Why now.** Found in the S14-33 ship loop (FU-1): 234 stale documents over 451 unpromotes in the probe. It predates S14-33. The planning probe reproduced it in both directions (§4).

**What stays.** Promote, filing, and Subject delete behave as they do. A bridge delete (`subjects.Delete` of a bridge Subject) already rewrites its ends (`linkedHandles`, `subjects.go:353`) and stays as it is. No end that no row reads a link through (`location.place`, `place_relationship.to`) is walked from: the 6,000-event city stays as cheap as S14-33 left it.

## 2. User stories

**US-1.** As a researcher, when I unpromote a person or an event that a filed participation joins, the other end's search document and list row stop showing the link, so the omnibar and the lists agree with the pages.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | James Robins (named) and a birth dated 1885-05, both promoted, a filed participation (role subject). Unpromote James | The event's search document equals a `searchindex.RebuildAll`: title "Unspecified birth", not "Birth of James Robins" |
| AC-1.2 | The same world. Unpromote the birth | James's search document equals a rebuild: no `1885-05` in its secondary text |
| AC-1.3 | The same world, with the person and event summaries loaded into the display cache. Unpromote James, then the birth in a second world | `Graph().Verify()` passes after the write (no stored display row differs from a fresh one) |
| AC-1.4 | Unpromote James | The write's effects name the event's handle in `Set.Handles`, so the app's `touchedHandles` reloads an open event page |

**US-2.** As a researcher, when I unpromote a place that a filed location or part_of reads, the event or child place that read it is rewritten.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A birth located at Leeds (both promoted, location filed). Unpromote Leeds | The birth's search document and stored rows equal a rebuild |
| AC-2.2 | Toronto part_of Ontario (both promoted, part_of filed) and a birth at Toronto. Unpromote Ontario | Toronto's and the birth's search documents equal a rebuild |

**US-3.** As the engine, an Unpromote at a crowded place costs what it costs today.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Leeds with 3 and with 12 other promoted births located there. Unpromote one located birth | The write's `SummaryDeps` hold none of the other births (no walk from the place end) |
| AC-3.2 | `TestLargeSourceBudgets` fixture (city with 6,000 events). Unpromote an event located at the city | Under `budgetUnpromoteAtCity` (500 ms); measured 45 ms at planning, today and with the prototype of D-1 |

## 3. Scope

| In | Covered by |
| --- | --- |
| A deleted accepted claim of a bridge Subject names the handles of the bridge's walk ends in `Set.Handles` | STEP-3, AC-1.1 – AC-2.2 |
| One table of the ends no row reads a link through, in `core/hops`, read by both `conclusionsummaries.AssociationEndpoints` and the new effects path | STEP-1, D-2 |
| Rebuild-equals-upkeep scenarios for Unpromote of each end of a participation, of a location's place, and of a part_of's whole | STEP-2, T-1 – T-4 |
| A display-cache test and a walk test for the deleted link | STEP-2, T-5, T-6 |
| A budget for Unpromote at the city | STEP-5, T-7 |
| Doc comments on `AssociationEndpoints`, `RecomputeReported`, the effects README | STEP-4 |

| Out | Home |
| --- | --- |
| Telling the app about the place end of a deleted location or part_of (the end no row reads through) without walking from it, so a cached city page reloads | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md): new row, STEP-6 |
| `subjects.Delete` walking from the place end of a deleted location bridge (FU-4) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), existing row |
| Reprojection cost when the whole of a part_of with a crowded part is unpromoted (the FU-3 shape: the part's rows do move) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), existing FU-3 row |
| Moving `subjects.Delete`'s own snapshot (`linkedHandles` + `Dependents`) onto the new effects path | not planned: the bridge's Observations are deleted before `effects.Resolve` runs, so the path cannot read them there |
| Bridges in the seeded rebuild-equals-upkeep generator | not planned: the named scenarios hold each shape |

## 4. Current state

Verified at `986d304`.

| Area | Where | Today |
| --- | --- | --- |
| Unpromote | `core/database/promote/unpromote.go` `Unpromote` | Releases the Subject's accepted claim (`deleteimpact.ReleaseClaim`), then `identityclaims.UnfileBridges`. Returns the changes; `writes.Run` recomputes `effects.Resolve(changes).Handles` |
| Unfiling | `core/database/identityclaims/bridgefile.go` `UnfileBridges`, `FiledBridgesOf` | Releases each filed bridge's accepted claim. The bridge Subject and its Observations stay |
| Handle delete | `core/database/deleteimpact/handle_gone.go` `releaseHandleIfGone` | When a claim was the last accepted member, deletes the `canonical_entities` row. `auto_reconciler_values.entity_id` is `ON DELETE CASCADE` (`core/database/migrations/000034.sql:13`), so the association's value rows go with it, inside the closure |
| Claim effects | `core/database/effects/registry.go` `"identity_claims"` | `Handles: union(field("entity_id"), onStatus("accepted", from("subject_id", observers)))`. `observers` (`registry.go:302`) is the handles whose members point **at** the Subject. Nothing reaches the handles a bridge Subject points at |
| Recompute | `core/database/autoreconciler/autoreconciler.go` `RecomputeReported` | Calls `conclusionsummaries.AssociationEndpoints` "before" and "after" `recomputeBatch`. Both run after the closure, so for a deleted association both are empty: no end is walked from, reprojected, or added to `deps` |
| Ends | `core/database/conclusionsummaries/dependents.go` `AssociationEndpoints`, `linkEnds` | Reads `auto_reconciler_values` joined to `canonical_entities`. `linkEnds` (private map, `dependents.go:156`) names `location.place` and `place_relationship.to`, built from `hops.PlacesOfEvent` and `hops.ParentsOfPlace` |
| Hops | `core/hops/hops.go` | The product hops; no list of unread ends. No tests in the package |
| Bridge delete | `core/database/subjects/subjects.go` `Delete`, `linkedHandles` | Snapshots the bridge's end handles and their `Dependents` before releasing, and calls `autoreconciler.RecomputeTx` itself. Rewrites the ends correctly (scenario below passes) |
| Writes | `core/writes/run.go` `Run` | `effects.Resolve` → `audit.Record` → `RecomputeReported(set.Handles)` → `set.SummaryDeps`. `api/ffi/handlers/write_effects.go` sends `Handles` and `SummaryDeps`; Swift `CatalogWriteEffects.touchedHandles` is both |
| Display cache | `core/database/graphcache/commit.go` `detachAssociation`, `dropCommittedDisplays` | A loaded association is detached and its ends reloaded; the planning probe (AC-1.3 shape) passed today. Stored displays drop `roots` + `SummaryDeps` |
| Tests | `core/database/autoreconciler/autoreconciler_test.go` `TestRebuildEqualsUpkeep_Scenarios`, `(*fixture).assertUpkeepEqualsRebuild`, `handleDocs` | Compare `auto_reconciler_values` and person/event/place search documents with a rebuild. Scenario "promoting both ends of a participation files it, then the bridge is deleted" passes. No Unpromote scenario |
| Tests | `core/database/conclusionsummaries/filedlinks_test.go` `TestFiledLocationWalksOnlyFromItsEvent`, `(*fixture).promotedBirth`, `countIn` | The filing-side walk rule |
| Tests | `core/database/promote/unpromote_test.go` `TestUnpromote` | Calls `autoreconcilertest.AssertCacheMatchesRebuild` (value rows only, not documents), so it passes today |
| Tests | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | Budgets `budgetLocateAtCity`, `budgetSecondPartOfAtCity` (500 ms); no Unpromote budget |

**Probe at planning** (scenario of T-1 and T-2, deleted after): unpromote the person → upkeep event doc `Birth of James Robins`, rebuild `Unspecified birth`. Unpromote the event → upkeep person doc keeps `1885-05`, rebuild drops it. Bridge delete of the same world → equal.

**Precedent.** `subjects.linkedHandles` (`subjects.go:353`) is the same read (a bridge's Observations → `value_subject_id` → accepted claims → handles), done by hand before the delete. S14-33 F-2 (`AssociationEndpoints` walk ends, `linkEnds`) is the rule for which ends to walk. Copy the rule, not the hand-call: the change belongs in the effects registry, which every write reaches through `writes.Run`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The `identity_claims` entry gains `onDelete(onStatus("accepted", from("subject_id", bridgeWalkEnds)))` in its `Handles` union. `bridgeWalkEnds` *(new)* is a `sqlPath`: the handles (accepted or provisional claims) of the `value_subject_id`s of the Subject's Observations, leaving out an Observation whose Subject's type and Property are an unread end (D-2) | The bridge Subject and its Observations outlive an Unpromote, so `effects.Resolve` can read the ends after the closure. The registry is where every write learns what it touches; Unpromote, a reject or any later path that deletes a filed bridge claim gains it | (a) Snapshot ends in `Unpromote` and call `RecomputeTx` there, as `subjects.Delete` does — a second hand-path off `writes.Run`, already listed as debt ("catalog graph cache leftovers"); (b) fix `AssociationEndpoints` — the rows are gone by CASCADE before it runs; (c) unfiltered path (every end) — prototype: 10,004 `SummaryDeps` and 161 ms for one Unpromote at the city, against 0 and 45 ms |
| D-2 | `core/hops` owns the unread ends: `type UnreadEnd struct{ Origin, Bridge, End string }` *(new)* and `func UnreadEnds() []UnreadEnd` *(new)*, returning `{PlacesOfEvent.Origin(), PlacesOfEvent.Bridge(), PlacesOfEvent.To()}` and the same for `ParentsOfPlace`. `conclusionsummaries.linkEnds` is built from it; `effects` builds D-1's exclusion clause from it at package init | One table for one rule: the walk the filing side skips and the walk the unfiling side skips must agree. `hops` is pure (imports `connectrules`, `vocabkey`) and both callers can import it; `effects` cannot import `conclusionsummaries` (cycle: `conclusionsummaries` imports `effects`) | A second literal list in `effects` — two copies drift |
| D-3 | D-1 runs only on a delete of a claim whose old status is accepted | A create or update of a bridge claim keeps the association's value rows, so `AssociationEndpoints` "after" already names its ends (S14-33). Running on filing too would add the ends as recompute roots and walk from them on every connect | `onStatus` alone — doubles the work on every filing |
| D-4 | `HandleScope` stays `field("entity_id")` | An end's last-activity time does not move because a link to it was taken back by another Subject's Unpromote; the audit scopes are unchanged | Adding the ends to `HandleScope` — changes the Sources/People "Updated" sort for a write the researcher did not make on that end |
| D-5 | Tests: the document checks are scenarios in `TestRebuildEqualsUpkeep_Scenarios` (`assertUpkeepEqualsRebuild` already compares rows and documents with a rebuild); the display and walk checks go in `conclusionsummaries` beside `filedlinks_test.go`; the budget is a subtest of `TestLargeSourceBudgets` | Each shape joins the suite that already guards it (the scenarios' comment: "A new write path or trigger adds its operation to the generator and, where it has a characteristic sequence, a scenario") | A new test package or helper — duplicates `handleDocs` |
| D-6 | `budgetUnpromoteAtCity` *(new)* = 500 ms | Same order as `budgetLocateAtCity`; measured 45 ms on the planning machine; CI is about 3× slower | A tighter figure — flakes on CI |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `hops.UnreadEnd`, `hops.UnreadEnds` *(new)* | type / func | `conclusionsummaries.linkEnds` (private map) | Extended by moving: the map becomes a read of `UnreadEnds`; `effects` cannot reach a private map in `conclusionsummaries` |
| `bridgeWalkEnds` *(new)* in `core/database/effects/registry.go` | path | `observers` (`registry.go:302`, the reverse direction); `subjects.linkedHandles` (same read, private, unfiltered, accepted only) | `observers` walks inbound; a `chain(inbound("observations","subject_id"), field("value_subject_id"), membersOf)` cannot express the type-and-Property exclusion, so it is a `sqlPath` as `sqlHandlesForTerm` is. `linkedHandles` stays (Out) |
| Scenarios in `TestRebuildEqualsUpkeep_Scenarios` | test | the "promoting both ends of a participation files it, then the bridge is deleted" scenario | Extended: same fixture and `connect.CreateCitedBridge` call |
| `TestUnpromotedLinkEndsInDisplayCache`, `TestUnpromotedLocationWalksOnlyFromItsEvent` *(new)* | test | `TestDisplayCacheFollowsAPlaceThreeStepsUp`, `TestFiledLocationWalksOnlyFromItsEvent` | Extended: reuse `promotedBirth`, `countIn`, `placeSubject` |
| `budgetUnpromoteAtCity` *(new)* | const | `budgetLocateAtCity` | One more row of the same table |

**Builds on.** The effects registry (`core/database/effects`), its `sqlPath`/`onDelete`/`onStatus`/`from` path helpers, and S14-33's walk-end rule.
**Deletes.** `conclusionsummaries.linkEnds`' literal map body (it becomes a read of `hops.UnreadEnds`). The `RecomputeReported` comment claim that the "before" read finds a link's ends, for a deleted association.
**Next adds.** A new bridge kind with an end no row reads through is one `UnreadEnds` entry, and both the filing and unfiling walks honour it.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Unpromote's effects now name the remaining end handles in `Handles` | Swift `CatalogWriteEffects.touchedHandles` → `conclusionDetail` refresh by id (`CatalogQueryRegistry.swift` `.touchedHandles`) | The end still exists (its claim was not deleted); its page reloads as after any write naming it |
| A deleted claim whose Subject was itself deleted in the same batch (`subjects.Delete` of a bridge) | `bridgeWalkEnds` | The Observations are gone, the query returns nothing; `subjects.Delete` keeps its own snapshot (the existing scenario stays green) |
| A deleted claim of a non-bridge Subject (a person's own claim) | `bridgeWalkEnds` | Returns the handles of that Subject's subject-valued Observations, if it has any; each is a handle whose value the Subject pointed at, so recomputing it is correct and costs one handle each |

### Scale

At the map's scale (S14-33 fixture: 4,000 persons, 8,000 events, one city with 6,000): one Unpromote adds one query per deleted bridge claim (indexed on `observations.subject_id`). Unpromoting an event at the city adds no root (the place end is excluded): prototype 45 ms, 0 `SummaryDeps`, same as today. Unpromoting a person with N filed participations adds N event handles as roots; each walks `Dependents` one hop (subjects of the event). Unpromoting a part_of's whole adds the part as a root: the walk from the part reaches its descendants' events, as filing that part_of does (FU-3, Out).

### Edges

- **Assumptions.** `effects.Resolve` runs after the closure (`run.go`), so the bridge Subject's Observations are read as they stand after Unpromote: present. `membersOf` statuses (accepted, provisional) match `linkedHandles`' accepted-only read plus provisional members, which also feed values. T-1 – T-4 hold it.
- **Neighbors.** Filing (claim create) — unchanged by D-3; `TestFiledLocationWalksOnlyFromItsEvent` and `TestFiledPartOfWalksFromThePart` stay green. Bridge delete — unchanged; its scenario stays green. Unpromote of a Subject with no filed bridges — no bridge claim deleted, nothing added (`TestUnpromote` "a Subject with no filed bridges loses only its own claim").
- **Removed inputs.** The unpromoted Subject's own handle: deleted when last member (in `Set.Deleted`), else recomputed via `field("entity_id")`. The association handle: deleted, or recomputed when a second record keeps it. The remaining end: recomputed and its document rewritten (this PR). The place end: not touched (Out row 1).
- **Real data.** `UnreadEnds` is two rows; `connectrules.Bridges()` today: participation, relationship, location, place_relationship. Walk ends: participation person and event, relationship both persons, location event, place_relationship from.
- **Untrusted size.** None — the path reads catalog rows the app wrote, bounded by the bridge's Observations.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/hops` | `type UnreadEnd struct{ Origin, Bridge, End string }` *(new)*; `func UnreadEnds() []UnreadEnd` *(new)*: "the bridge ends no stored row reads a link through" — location.place and place_relationship.to, from `PlacesOfEvent` and `ParentsOfPlace` |
| `core/database/conclusionsummaries` | `linkEnds` built from `hops.UnreadEnds()` (bridge → end). `AssociationEndpoints` doc: a deleted association has no rows (CASCADE); its ends come from the deleted bridge claim through `effects` |
| `core/database/effects` | `bridgeWalkEnds` *(new)*, SQL sketch below, exclusion clause built from `hops.UnreadEnds()` at init; `identity_claims.Handles` gains the D-1 arm |
| `core/database/autoreconciler` | `RecomputeReported` comment only: a deleted association's ends arrive in `ids` from the effects registry |

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
-- the two NOT lines are generated from hops.UnreadEnds(), not written literally
```

Errors: none new.

### FFI

None — `WriteEffects.handles` already carries `Set.Handles`.

### macOS

None — `touchedHandles` already follows `handles`.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — One table of unread ends**
Files: `core/hops/hops.go`, `core/database/conclusionsummaries/dependents.go`.
Do: add `UnreadEnd` and `UnreadEnds` (D-2) after the product hops. Build `linkEnds` from `hops.UnreadEnds()`; keep the map's name and its use at `dependents.go:141`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/hops/... ./core/database/conclusionsummaries/...` passes (no behavior change).

**STEP-2 — Failing tests**
Files: `core/database/autoreconciler/autoreconciler_test.go`, `core/database/conclusionsummaries/filedlinks_test.go`.
Do: add T-1 – T-6 (§8). Build the participation world once per scenario with the same `connect.CreateCitedBridge` call as the "promoting both ends of a participation files it, then the bridge is deleted" scenario, with a name on the person and `event_type` birth and a 1885-05 date on the event. Unpromote through `writes.Call` with action `unpromote_subject`, as `core/database/promote/unpromote_test.go` `(*bridgeWorld).unpromote` does.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/ ./core/database/conclusionsummaries/ -run 'Scenarios|Unpromoted'` — T-1, T-2, T-3, T-4 fail with "handle search docs differ from a rebuild"; T-5 and T-6 pass (they pin what must not regress).

**STEP-3 — The deleted bridge claim names its walk ends**
Files: `core/database/effects/registry.go`.
Do: add `bridgeWalkEnds` (D-1, D-2) beside `observers` with a comment; add the D-1 arm to `identity_claims.Handles`.
Check: the STEP-2 command passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/database/promote/ ./core/database/identityclaims/ ./core/database/subjects/` passes.

**STEP-4 — Comments and README**
Files: `core/database/conclusionsummaries/dependents.go` (`AssociationEndpoints` doc), `core/database/autoreconciler/autoreconciler.go` (`RecomputeReported` comment), `core/database/effects/README.md` (one sentence: a deleted bridge claim names the handles of its walk ends; the unread ends are `hops.UnreadEnds`).
Check: `gofmt -l core` prints nothing.

**STEP-5 — Budget**
Files: `core/database/promotealign/perf_test.go`.
Do: add `budgetUnpromoteAtCity` (D-6) with a comment in the style of `budgetLocateAtCity`, and T-7 as the last subtest.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` logs the new line under budget.

**STEP-6 — Version and docs**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj` (`MARKETING_VERSION`), `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, this plan.
Do: patch bump per `docs/versioning.md`. Add a structural-debt row: "— | S14-38 plan | Unpromote does not name the place end of a deleted location or part_of to the app (no row reads a link through it, so it is not walked from); a cached place page that lists the link can stay stale until it reloads. Report that end without a walk when the effects Set can carry ends apart from `Handles`." Then the DoD lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-38-unpromote-stale-search-documents.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestRebuildEqualsUpkeep_Scenarios` "a filed participation, then the person is unpromoted" *(new)* | James + birth 1885-05, participation filed, unpromote James → `assertUpkeepEqualsRebuild` passes (event title "Unspecified birth") | AC-1.1, D-1 |
| T-2 | Go | same, "a filed participation, then the event is unpromoted" *(new)* | unpromote the birth → passes (person doc has no 1885-05) | AC-1.2, D-1 |
| T-3 | Go | same, "a filed location, then the place is unpromoted" *(new)* | birth located at a named place, both promoted; unpromote the place → passes | AC-2.1, D-2 |
| T-4 | Go | same, "a filed part_of, then the whole is unpromoted" *(new)* | Toronto part_of Ontario, a birth at Toronto; unpromote Ontario → passes | AC-2.2, D-2 |
| T-5 | Go | `core/database/conclusionsummaries/filedlinks_test.go` `TestUnpromotedLinkEndsInDisplayCache` *(new)* | table {unpromote person, unpromote event}: load `PersonsByIDs` and `EventsByIDs`, unpromote via `writes.Run` → `f.c.Graph().Verify()` is nil; the result's `Effects.Handles` holds the remaining end | AC-1.3, AC-1.4 |
| T-6 | Go | same file, `TestUnpromotedLocationWalksOnlyFromItsEvent` *(new)* | table {3, 12} sibling births at Leeds; a further birth located there; unpromote that birth via `writes.Run` → `countIn(res.Effects.SummaryDeps, siblings...) == 0` and Leeds is not in `Effects.Handles` | AC-3.1, D-1, D-3 |
| T-7 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "unpromote an event at the city" *(new)* | 3 births promoted and located at the city; fastest of 3 Unpromotes → under `budgetUnpromoteAtCity` | AC-3.2, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/effects` (`TestHandlePaths`, `registry_completeness_test.go`), `core/database/conclusionsummaries` `TestFiled*`, `core/database/promote` `TestUnpromote*`, the S14-33 budgets in `TestLargeSourceBudgets`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): patch
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none; core/database/effects/README.md updated (STEP-4)
- [ ] docs/ideas/structural-debt.md has the STEP-6 row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a dogfood project. In an Evidence graph, make a person (name "James Robins") and an event (type Birth, date May 1885), connect them with a participation (role Subject). Promote both.
2. Search the omnibar for "Birth": the hit reads "Birth of James Robins".
3. Unpromote James. Search "Birth" again: the hit reads "Unspecified birth", the Events list row agrees, and the open event page (if open) no longer lists James.
4. Promote James again (the participation files), then unpromote the event. Search "James": the person hit shows no 1885 date.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The new arm fires on a path other than Unpromote and widens its recompute | D-3 (delete only); T-6 and the S14-33 budgets |
| A part_of whole Unpromote is slow at scale | Out (FU-3 row); the cost equals filing that part_of |

**Stop and report if:**

- `RecomputeReported` or `AssociationEndpoints` no longer match §4 (S14-33 follow-ups moved them).
- T-1 or T-2 passes before STEP-3 (the bug is fixed elsewhere; report the commit).
- Any S14-33 budget (`budgetLocateAtCity`, `budgetSecondPartOfAtCity`) or T-7 goes over after STEP-3.
- `effects` cannot import `core/hops` without an import cycle (`go list -deps ./core/hops` gains a `core/database` package).
- `TestHandlePaths` or `registry_completeness_test.go` fails on the new `sqlPath` and the fix needs a change to `path.go`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-38 — Unpromote rewrites the ends of the links it takes back
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (T-1 – T-7 output); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-38-unpromote-stale-search-documents.md`.
