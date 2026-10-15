# S10-21 — Handle last-activity scope

**Status:** Landed in [#398](https://github.com/mendahu/provenencia/pull/398) (2026-10-14)
**Spike / slice:** Spike 10 · Slice 6 — Lists and the Person page
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-21--handle-last-activity-scope)
**Brief:** — (no new chrome: an engine scope and list fields; the row cell is S10-22 from S10-D9, S10-D10, and S10-D11)
**Board:** —
**Depends on:** —
**Blocks:** S10-22
**Branch:** `s10-21-handle-last-activity`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`docs/audit-revision-history.md` §4.1](../../../audit-revision-history.md#41-scopes), [`core/database/effects/README.md`](../../../../core/database/effects/README.md), working rule "**Last updated** is an audit scope, not a timestamp column", gotchas "Audit scope fan-out" and "Last updated is not create time" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-14 against `main` @ `a8d7971a`
**Review:** [`S10-21-handle-last-activity.review.md`](S10-21-handle-last-activity.review.md) — Fits with changes, Ready with changes (applied); 2 rounds

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

Every Person, Event, and Place handle gets a "last activity" revision. It moves when the researcher works on that handle: a claim filed, changed, or released, an evidence pin, the handle's own row, or evidence on a member Subject (its Observations, their date and name values, and their notes). The People, Events, and Places list reads return that revision and its time, the way the Sources list already returns `UpdatedRevision`. S10-22 draws the row cell from these fields.

**Why now.** Slice 6 shows "last updated" on every list. Sources already has a `source` scope. Handles have none: conclusion-layer writes store no scope at all today, so the lists have nothing to show.

**What stays.** The `source` scope and every revision it stores are unchanged. Conclusion-layer writes still move no Source (`TestSourceScopes` "promote does not move the source"). The `Handles` recompute set in `effects` resolves the same ids for every table. List order, header composition, and the stored display rows are unchanged. No Swift code changes beyond the generated protobuf file.

## 2. User stories

**US-1.** As a researcher, when I promote, file, or release a Subject onto a Person, Event, or Place, that handle's last activity moves to that revision, so the list can say when I last worked on it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Promote Subject *Bob* on Source A to a new Person | The new handle's latest `canonical_entity` revision is the promote revision. Source A's `UpdatedRevision` does not move. |
| AC-1.2 | Delete Subject *Bob* (an accepted member) | Bob's handle moves to the delete revision. Source A moves too. |
| AC-1.3 | A handle has no stored `canonical_entity` scope. Through `writes` this cannot happen: creating a handle records a `canonical_entity` change. It happens for history the backfill could not map, and the test builds it with a raw `INSERT` (T-3, T-4) | It is still in the list. Its `Updated` is the zero `audit.Latest`: revision `0`, time `""`. |

**US-2.** As a researcher, when I edit evidence on a member Subject, that handle's last activity moves, and the Source's still moves too.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Add an Observation (toponym "Boston") on member Subject *Boston* | Boston's handle and Source A are both at the latest revision. |
| AC-2.2 | Edit that Observation's text to "Boston Town" | Both move again. |
| AC-2.3 | Save that Observation with a note, then with the note edited | Both move each time. |
| AC-2.4 | Edit a member Event's date Observation in place (1842 → 1843) | The Event's handle and Source A move. |
| AC-2.5 | Delete that Observation (it has a note) | Both move. The stored scopes equal `effects.HandleScopes` and `effects.Sources` for that revision. |
| AC-2.6 | Add an Observation on Subject *Salem*, which has no claim | Source A moves. No handle moves. |

**US-3.** As a researcher, a change that only recomputes a handle without being work on it does not move its last activity.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A change list that edits a property's `cardinality` | `effects.HandleScopes` is empty. `effects.Handles` is not empty. |
| AC-3.2 | A change list that flips a citation's `transcription_uncertain`, or changes a Source credibility grade | `effects.HandleScopes` is empty. |
| AC-3.3 | *Ada* (member of handle P1) has an Observation of a person-to-person property (`seenWith`, created by `TestConclusionHandleSets`; P1 is `adaPromoted.Entity.ID`) → Subject *Bob*. Bob is claimed onto a new handle P2 | `effects.HandleScopes` for that claim is `{P2}` and does not contain P1. |

**US-4.** As the app, the People, Events, and Places list reads carry each row's latest revision and its time.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `ListPersonHeaders`, `ListEventHeaders`, `ListPlaceHeaders` after promotes | Every promoted row has `updated_revision` > 0 and `updated_at` equal to that revision's `audit_transactions.created_at` (RFC3339 UTC). |
| AC-4.2 | Edit a member Observation of Person *Bob* and list again | Bob's row has the new revision. Every other row keeps its value. |
| AC-4.3 | Open a copy of a catalog at the previous format | The migration backfills `canonical_entity` scopes from history. Open passes the schema check. Every promoted handle has at least one scope row (manual check step 2). |

## 3. Scope

| In | Covered by |
| --- | --- |
| `effects.Effect.HandleScope` *(new)* path, `Effect.paths` *(new)*, and `effects.HandleScopes` *(new)* resolver | STEP-1, AC-3.1, AC-3.2, AC-3.3 |
| `audit.ScopeCanonicalEntity` *(new)*; `audit.Record` stores every scope type from one table, `scopeResolvers` *(new)* | STEP-2, AC-1.1, AC-1.2, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6 |
| `audit.LatestByScope` *(new)*: latest revision and its time per scope id | STEP-2, AC-1.3 |
| Parity: stored scopes of each type equal their resolver for every revision a scope test writes | STEP-3, AC-2.5 |
| One shared scope-test seed, `newScopeSeed` *(new)*, used by `TestSourceScopes` and `TestHandleScopes` *(new)* | STEP-3, STEP-4 |
| `conclusionheaders.ListPersons` / `ListEvents` / `ListPlaces` fill `Updated audit.Latest` *(new)* | STEP-5, AC-4.1, AC-4.2 |
| Proto `updated_revision` / `updated_at` on `PersonHeader`, `EventHeader`, `PlaceHeader`; handlers map them | STEP-6, AC-4.1 |
| Migration `000051.sql` *(new)* backfills `canonical_entity` scopes from history | STEP-7, AC-4.3 |
| Docs: audit doc §4.1, `add-audit-scope` skill, effects README, map working rule and S10-41 entry | STEP-8 |

| Out | Home |
| --- | --- |
| The row cell on People, Events, and Places; Swift model fields; `GoStore` / `FakeStore` mapping | S10-22 |
| Last-updated time on the Sources list, and moving `sources.sqlList` / `sqlLatestRevision` to `audit.LatestByScope` | S10-41 (STEP-8 writes this into its map entry) |
| Sorting People, Events, or Places by last activity | not planned (S10-22 says it adds no sort) |
| An `updated_at` column on `canonical_entities` | not planned (map row Out) |
| Citation edits, Subject label edits, Source credibility, vocabulary edits, and other handles whose relationships point here moving a handle | not planned (decided 2026-10-14, see D-2) |
| Filling `Updated` on `PersonsByIDs`, `EventsByIDs`, `PlacesByIDs`, search hits, Promote alignment, or Conclusion detail | not planned (D-6) |
| A `HandleScope` field on `effects.Set`, or any change to `effects.Resolve` / `writes.Run` / `WriteEffects`. A note-only Observation save reports no `touchedHandles`, so the app's cached People, Events, and Places lists do not reload on it | S10-22 (STEP-8 writes this into its map entry) |
| Editing the shipped `000036.sql` comment that says only `source` is written | not planned (shipped migrations are never edited) |
| A test of the migration itself | not planned ([`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md) step 4) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Scope type | `core/database/audit/scopes.go` `ScopeSource`, `Scope`, `sqlInsertScope` | One scope type, `source`. The map's working rule names a `noScope` resolver in this file; that is stale. Resolvers moved to `effects` and there is no `noScope`. |
| Record | `core/database/audit/audit.go` `Record` (end of the function) | Inserts the transaction and changes, then stores one `source` scope per id from `effects.Sources(tx, rev.Changes)`. Nothing for handles. |
| Effect registry | `core/database/effects/registry.go` `Effect`, `registry`, `membersOf`, `observers`, `Effect.validate` | `Source` is the audit scope path. `Handles` is the recompute set: it also fans out on vocabulary (`properties.cardinality`, `property_terms.neutral`), `citations.transcription_uncertain`, credibility grades, and `observers` on an accepted claim. `canonical_entities`, `identity_claims`, `identity_claim_evidence` have no `Source`. `observation_notes` has no `Handles`. `validate` lists the four paths (`Source`, `CacheSource`, `Handles`, `Structure`) twice (the `None` guard, which also tests `Search` and `Vocabulary`, and the check loop). `observations`, `date_values`, and `name_values` spell their member paths inline. |
| Resolvers | `core/database/effects/resolve.go` `Sources`, `Handles`, `resolve`, `Resolve` | `Sources` and `Handles` are `resolve` with a field picker. `Resolve` builds `Set` for `writes.Run`. |
| Path DSL | `core/database/effects/path.go` `self`, `up`, `field`, `from`, `inbound`, `across`, `chain` | A delete reads parent ids from the change's old fields and the ghost map. `across` reads live `identity_claims` only. |
| Write path | `core/writes/run.go` `Run` | `WithoutNone`, then `audit.Record`, then `effects.Resolve`. |
| Claim changes | `core/database/identityclaims/identityclaims.go` `InsertTx`; `core/database/deleteimpact/facets.go` claim release and `pinRelease` | Claim create and delete carry `entity_id`. A pin change uses the claim id as `EntityID` and carries `identity_claim_id`. A Subject delete releases its claims as a facet (`facets.go`, `Via: "identity_claims.subject_id"`). |
| Observation changes | `core/database/observations/write.go` `replaceNotes`, `update.go` | Delete carries the full row (`subject_id`, `citation_id`). Notes are written through `observations.Update` with `Input.Notes`; `replaceNotes` emits `observation_note` create and delete changes carrying `observation_id`. Value rows are released as `date_value` / `name_value` deletes. |
| Source revision read | `core/database/sources/sources.go` `sqlList`, `sqlLatestRevision`, `AttachLatestRevision` | Correlated `MAX(t.revision)` per row, `scope_type = 'source'`. Returns the revision only, no time. |
| List reads | `core/database/conclusionheaders/compose.go` `ListPersons`, `ListEvents`, `ListPlaces`, `persons`, `events`, `places` | Compose from `c.Graph()`. Finished rows are stored in the graph's display cache (`g.Display`). No revision. |
| Header types | `core/database/conclusionheaders/conclusionheaders.go` `PersonHeader`; `events.go` `EventHeader`; `places.go` `PlaceHeader` | No revision fields. Each has `Entity canonicalentities.Entity`. |
| Proto | `api/proto/engine.proto` `PersonHeader` (fields 1–5), `EventHeader` (1–14), `PlaceHeader` (1–11); `Source` field 12 `updated_revision`; `ProjectInfo` field 4 `updated_at` | No revision on headers. |
| Handlers | `api/ffi/handlers/conclusion_headers.go` `ListPersonHeaders`, `ListEventHeaders`, `ListPlaceHeaders`, `personHeaderProto`, `eventHeaderProto`, `placeHeaderProto` | The `*HeaderProto` helpers are also used by `promote_align.go`, `search_catalog.go`, and `conclusion_detail.go`. |
| Migrations | `core/database/migrations/000036.sql` (scope table + `source` backfill), `000032.sql` (`canonical_entities`: `id`, `subject_type_id` NOT NULL, `ref` UNIQUE NOT NULL, `argument`, `label`, `merged_into_id`), latest `000050.sql` | Next step is `000051.sql`. Primary key `(scope_type, scope_id, audit_transaction_id)` covers the list read. |
| Scope tests | `core/database/audit/scopes_test.go` `TestSourceScopes` | `seed`, `mustSeed`, `revisions`, `latest`, `bumpsA`, `mustCitation`, `textObs` are closures inside the test, capturing `userID` and `validLocator`, so another test cannot use them. The seed has Sources A and B, an artifact, Subjects person (Bob), event (Birth), place (Boston), place2 (Salem), and the toponym and date properties. It promotes nothing. |
| Parity check | `core/database/audit/scopes_effects_test.go` `seenEffectsRevisions`, `checkNewEffects`, `loadSourceScopes`, `loadRevisionChanges`, `sameIDSet` | Compares stored `source` rows with `effects.Sources` only. |
| Effects tests | `core/database/effects/handle_paths_test.go` `TestHandlePaths`, `mustHandles`, `observationSubject`, `idSet`; `conclusion_handles_test.go` `TestConclusionHandleSets`; `registry_completeness_test.go` `hasWork` | `TestHandlePaths` has members alice (accepted), drew (provisional on alice's handle), cara (rejected), bob, eve, boston, a citation `created`, and subtests "members keep accepted and provisional", "transcription_uncertain only", "date value reaches its observation". `TestConclusionHandleSets` has `seenWith`, Ada, Bob and subtests "accepted mint covers claim handle, filings, and observers", "grade change matches source handles", "cardinality matches property handles". `hasWork` lists the effect fields by hand. |
| Header tests | `core/database/conclusionheaders/conclusionheaders_test.go` `newFixture`, `fixture.person`, `fixture.event`, `fixture.place`, `fixture.book`, `fixture.list`, `fixture.events`, `fixture.places`, `TestListPersons`, `TestListPersonsQueryCountIsConstant`; `walks_test.go` `fixture.bare`, `fixture.cite`, `fixture.promoteSubject` | `person`, `event`, and `place` create, cite, and promote and return the handle id. `QueryCount` tests count graph queries only (`SetQueryCounterForTest`). |
| Handler tests | `api/ffi/handlers/conclusion_headers_test.go` `TestListPersonHeaders`, `TestListEventHeaders`, `TestListPlaceHeaders` | Promote rows and assert the response. |

**Precedent.** `000036.sql` and the `Source` effect path are the same shape: a scope type, stored inside `Record` from one effect path, read with `MAX(revision)`, backfilled with `INSERT OR IGNORE … SELECT`. Copy the effect-path-per-table shape and the backfill's temp-table style. Do not copy the hand-written per-entity resolvers 000036's header mentions; those were replaced by `effects`. `Effect.CacheSource` is the precedent for a second path beside an existing one ("drops a loaded Source graph; audit still uses Source"). `membersOf` and `observers` are the precedent for naming a shared path.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The handle scope is a new `Effect` field, `HandleScope Path`, resolved by `effects.HandleScopes(tx, changes)`. It is separate from `Handles`. | `Handles` is the recompute set and fans out on vocabulary, credibility, uncertainty, and observers. Those are not work on the handle (D-2). The audit doc says "a new scope type is one effect path plus a backfill migration". | Reusing `Handles`, because a property cardinality edit would move every Person. A hand resolver in `core/database/audit/scopes.go`, because `add-audit-scope` forbids it. |
| D-2 | Paths (product call, 2026-10-14): `canonical_entities` → `self()`; `identity_claims` → `field("entity_id")`; `identity_claim_evidence` → `up("identity_claim_id", "entity_id")`; `observations` → `observationMembers`; `observation_notes` → `chain(up("observation_id", "subject_id"), membersOf)`; `date_values` → `valueMembers("value_date_id")`; `name_values` → `valueMembers("value_name_id")`. No other table gets a `HandleScope`. Two named paths *(new)* sit beside `membersOf`: `var observationMembers = from("subject_id", membersOf)` and `func valueMembers(col string) Path { return chain(inbound("observations", col), field("subject_id"), membersOf) }`. The `Handles` entries of `observations`, `date_values`, and `name_values` use the same names, so they resolve exactly what they do today. | The working rule names claim writes and member Observation writes. The user chose to include the Observation's parts (values, notes), pins, and the handle row. Citation, Subject, credibility, vocabulary, and observer fan-out are out. Naming the shared paths keeps the scope set from drifting from the recompute set where they are meant to match. | Only claims and Observations, because a note or a date edit is evidence work on the member. Every recompute trigger, because it moves rows the researcher never touched. Repeating the chains inline, because the value chain would appear four times. |
| D-3 | Membership is `membersOf` as it is: accepted and provisional claims, read live. A claim change reaches its handle through its own `entity_id`, whatever its status. | Same membership the reconciler uses. A rejected claim's own write still moves that handle. A member Subject delete releases its claims in the same revision, so the claim deletes scope the handle even after the live claims are gone. | Accepted only, because a provisional member already shapes the reconciled values. |
| D-4 | `audit.ScopeCanonicalEntity = "canonical_entity"` *(new)*. `scopes.go` gets `var scopeResolvers = []struct { Type string; Resolve func(*sql.Tx, []rowchange.Change) ([][]byte, error) }{{ScopeSource, effects.Sources}, {ScopeCanonicalEntity, effects.HandleScopes}}` *(new)*. `Record` loops over it and stores each id with `sqlInsertScope`. `core/database/audit/export_test.go` *(new)* exports it as `ScopeResolvers` for the parity check. `effects.Set`, `effects.Resolve`, and `writes.Run` are unchanged. | One change can yield a Source and several handles. Scopes stay derived inside `Record`. One table means a third scope type is one row, read by both `Record` and the parity check. | An extra-scopes field on `audit.Revision`, because the skill forbids call-site tagging. A second hand-written block in `Record`, because the parity test would need its own copy. |
| D-5 | The read is in `scopes.go`: `type Querier interface { Query(query string, args ...any) (*sql.Rows, error) }`, `type Latest struct { Revision int64; At string }`, and `func LatestByScope(q Querier, scopeType string) (map[string]Latest, error)` *(all new)*. One query: `SELECT sc.scope_id, MAX(t.revision), t.created_at FROM audit_transaction_scopes sc JOIN audit_transactions t ON t.id = sc.audit_transaction_id WHERE sc.scope_type = ? GROUP BY sc.scope_id`. The map key is `string(scope_id)`. Nil `q` or empty `scopeType` → `ErrInvalid`. Query, scan, and `rows.Err` errors are returned as they are. | One indexed query per list, whatever its length. SQLite returns the bare `t.created_at` from the row that holds the `MAX`. The reader sits beside the writer constant. `Querier` matches `conclusionheaders.Querier` and `graphcache.Querier`. | A correlated subquery per row like `sources.sqlList`, because the header lists are composed from the graph, not one SQL row. A per-id `IN (…)` list, because the lists read every handle anyway. A new `latest.go`, because `scopes.go` is 16 lines and owns the scope type. |
| D-6 | Each header gets one field, `Updated audit.Latest` *(new)*, commented "Filled by List only; zero means none", as `sources.Source.HasArtifact` is. `compose.go` gets `func latestHandles(c *database.Catalog) (map[string]audit.Latest, error)` *(new)*, which calls `c.DB()` and `audit.LatestByScope(db, audit.ScopeCanonicalEntity)`. `ListPersons`, `ListEvents`, and `ListPlaces` call `latestHandles` once after `persons` / `events` / `places` return, and set `out[i].Updated = latest[string(out[i].Entity.ID)]`. A `c.DB()` or read error is returned from the list. `PersonsByIDs`, `EventsByIDs`, `PlacesByIDs`, `*FromQuerier`, and `RowsForSearch` leave `Updated` zero. | The display cache (`g.Display`) is dropped by the recompute set, not by the handle scope. A note edit moves the scope without dropping the cached row, so a revision stored in the cache would go stale. Filling the returned copies after the cache read keeps the cache free of revisions. One typed field keeps "revision and its time" one concept from audit to the list. | Storing the revision in the composed header, because it goes stale. Two scalar fields per header, because that is six fields for one concept. Filling it in the FFI handler, because Go callers of `List*` would differ from the FFI. |
| D-7 | Proto: `PersonHeader` `int64 updated_revision = 6; string updated_at = 7;`, `EventHeader` `= 15` / `= 16`, `PlaceHeader` `= 12` / `= 13`. `updated_at` is the RFC3339 UTC string from `audit_transactions.created_at`, empty when `updated_revision` is 0. The three `*HeaderProto` helpers copy `h.Updated.Revision` and `h.Updated.At`. Other callers of those helpers send zero. | The engine sends a time string, as `ProjectInfo.updated_at` does. S10-22 formats it. Both fields keep a later sort possible without another read. | A `Timestamp` message, because no other engine field uses one. A shared `LastActivity` message, because two scalars match `Source.updated_revision`. |
| D-8 | Swift is untouched except `macos/App/Platform/Generated/engine.pb.swift`, regenerated by `./scripts/generate-proto.sh`. | The map row puts the row UI in S10-22; this PR ends at the engine contract. | Adding Swift header fields now, because `FakeStore` would need revision bumps that S10-22's tests own. |
| D-9 | Backfill `000051.sql` is the SQL in §6 Schema, verbatim. A claim id resolves to a handle through the change's own `entity_id` diff, else the live claim or any recorded `entity_id` of that claim. An Observation change resolves to Subjects through its own `subject_id` diff, else the live row or any recorded `subject_id`. A note resolves through its observation; a value through the live observation that holds it. Subjects map to handles through today's accepted and provisional claims. History that resolves to nothing stays unscoped. | Same rule as `000036.sql`: change fields first, then live rows. Membership at the time of an old write is not recorded, so today's membership is the best approximation, and it matches what the list will show. | A Go backfill at open, because migrations are SQL only (`core/database/migrate.go` embeds `migrations/*.sql`). No backfill, because every existing handle would read as never updated. |
| D-10 | Tests extend the existing fixtures. `TestSourceScopes`' seed and helpers move to file level in `scopes_test.go` as `scopeSeed` and `newScopeSeed(t)`; `TestHandleScopes` *(new)* lives in the same file. Effects cases are new subtests of `TestHandlePaths` and `TestConclusionHandleSets`, checked with `mustHandleScopes` *(new)* beside `mustHandles`. | One fixture per world. The next audited table or scope type adds rows, not a third seed. | A sibling test file with a copied seed, because two ~100-line seeds drift. A new effects fixture, because the existing ones already build every case. |
| D-11 | `effects` gets `func (e Effect) paths() []Path` *(new)* returning `Source`, `CacheSource`, `Handles`, `HandleScope`, `Structure`. `validate` uses it for the `None` guard and the check loop. `hasWork` in `registry_completeness_test.go` uses it too. Both still test `len(e.Search) > 0 || e.Vocabulary` by hand, since those are not paths. | Three hand lists of effect fields become one. A table with only a `HandleScope` counts as work. | Adding `HandleScope` to each hand list, because the next field is half-wired again. |
| D-12 | No new error codes, no L10n strings. | Nothing user-facing changes. | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Effect.HandleScope` *(new)* | field | `Effect.Handles`, `Effect.Source`, `Effect.CacheSource` | `Handles` means recompute (D-1). `CacheSource` is the precedent for a second path with a different job. |
| `Effect.paths` *(new)* | method | `Effect.validate`'s two inline lists; `hasWork` | Replaces three hand lists (D-11). |
| `observationMembers`, `valueMembers` *(new)* | var, func | `membersOf`, `observers` | Same pattern; names the paths `Handles` already spells inline (D-2). |
| `effects.HandleScopes` *(new)* | func | `effects.Sources`, `effects.Handles` | The same one-line `resolve` wrapper with a different picker. |
| `audit.ScopeCanonicalEntity`, `scopeResolvers` *(new)* | const, var | `audit.ScopeSource`; the hard-coded block in `Record` | The block becomes a loop over one table (D-4). |
| `audit.Querier`, `audit.Latest`, `audit.LatestByScope` *(new)* in `scopes.go` | type, type, func | `sources.sqlList` correlated subquery; `sources.sqlLatestRevision`; `conclusionheaders.Querier` | The header lists need a map for many handles, plus the time. `sources` keeps its query in this PR; S10-41 moves it (STEP-8 records that in the map). |
| `core/database/audit/export_test.go` *(new)* | test file | none in `audit` | Exports `scopeResolvers` to the `audit_test` package. |
| `Updated audit.Latest` on `PersonHeader`, `EventHeader`, `PlaceHeader`; `latestHandles` *(new)* | field, func | `sources.Source.UpdatedRevision` | Same meaning, on the header types the lists already return (D-6). |
| Proto `updated_revision`, `updated_at` on three headers *(new)* | proto fields | `Source.updated_revision` (12), `ProjectInfo.updated_at` (4) | Same names and types. |
| `core/database/migrations/000051.sql` *(new)* | migration | `000036.sql` | A shipped migration is never edited. |
| `scopeSeed`, `newScopeSeed` and its methods *(moved from closures in `TestSourceScopes`)* | test helpers | `TestSourceScopes`' `seed`, `mustSeed`, `revisions`, `latest`, `bumpsA`, `mustCitation`, `textObs` | Moved, not copied (D-10). |
| `loadScopes(tx, id, scopeType)` *(renamed from `loadSourceScopes`)* | test helper | `loadSourceScopes` | Takes the scope type so the parity check covers each type. |
| `TestHandleScopes` *(new)* in `scopes_test.go` | test | `TestSourceScopes` | Same seed and parity check; asserts the handle scope. |
| `mustHandleScopes` *(new)* in `handle_paths_test.go` | test helper | `mustHandles` | Same shape with `effects.HandleScopes`. |

**Builds on.** The `effects` registry and path DSL (`add-audit-scope`), `audit.Record`'s derived scopes, `000036.sql`'s backfill shape, the `checkNewEffects` parity check, the existing effects and header fixtures, and the header list handlers.
**Deletes.** The hard-coded `effects.Sources` → `ScopeSource` block in `Record` and the source-only comparison in `checkNewEffects` (both become loops). The two inline path lists in `Effect.validate` and the hand list in `hasWork`. The `seed` / `mustSeed` closures inside `TestSourceScopes` (moved). The stale "`canonical_entity` is `noScope` in `core/database/audit/scopes.go`" sentence in the map's working rule. The audit doc and skill sentences that say `source` is the only scope type.
**Next adds.** S10-22 reads `updated_at` off the three headers. S10-41 shows the Sources time with `audit.LatestByScope(db, audit.ScopeSource)` and drops the `sources` scope reads it makes dead. A later sort by last activity uses `updated_revision`. A new scope type is one `scopeResolvers` row and one effect field; a new conclusion-layer table is one `HandleScope` entry.

### Schema

Migration `core/database/migrations/000051.sql` *(new)*. Data only: no DDL on catalog tables, so the schema hash is unchanged. Write it as below.

```sql
-- canonical_entity audit scopes: when work on a Person, Event, or Place last
-- happened (S10-21). audit.Record stores them from effects.HandleScopes.
-- This backfill mirrors those paths for older history, best effort: a
-- change's own fields first, then live rows, then other recorded changes of
-- the same row. Subjects map to handles through today's accepted and
-- provisional claims; membership at the time of the write is not recorded.

-- claim id -> handle id, from the live claim and every recorded entity_id.
CREATE TEMP TABLE claim_handle AS
SELECT id AS claim_id, entity_id FROM identity_claims
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.entity_id.old'), '-', ''))
	FROM audit_changes WHERE entity_type = 'identity_claim'
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.entity_id.new'), '-', ''))
	FROM audit_changes WHERE entity_type = 'identity_claim';

-- observation id -> subject id, from the live row and every recorded subject_id.
CREATE TEMP TABLE observation_subject AS
SELECT id AS observation_id, subject_id FROM observations
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.subject_id.old'), '-', ''))
	FROM audit_changes WHERE entity_type = 'observation'
UNION
SELECT entity_id, unhex(replace(json_extract(changes_json, '$.subject_id.new'), '-', ''))
	FROM audit_changes WHERE entity_type = 'observation';

-- revision -> subject touched by evidence on it.
CREATE TEMP TABLE change_subject AS
SELECT audit_transaction_id AS tx_id,
	unhex(replace(json_extract(changes_json, '$.subject_id.old'), '-', '')) AS subject_id
	FROM audit_changes
	WHERE entity_type = 'observation' AND json_type(changes_json, '$.subject_id') IS NOT NULL
UNION
SELECT audit_transaction_id, unhex(replace(json_extract(changes_json, '$.subject_id.new'), '-', ''))
	FROM audit_changes
	WHERE entity_type = 'observation' AND json_type(changes_json, '$.subject_id') IS NOT NULL
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_subject os ON os.observation_id = c.entity_id
	WHERE c.entity_type = 'observation' AND json_type(c.changes_json, '$.subject_id') IS NULL
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_subject os ON os.observation_id =
		unhex(replace(json_extract(c.changes_json, '$.observation_id.old'), '-', ''))
	WHERE c.entity_type = 'observation_note'
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_subject os ON os.observation_id =
		unhex(replace(json_extract(c.changes_json, '$.observation_id.new'), '-', ''))
	WHERE c.entity_type = 'observation_note'
UNION
SELECT c.audit_transaction_id, os.subject_id FROM audit_changes c
	JOIN observation_notes n ON n.id = c.entity_id
	JOIN observation_subject os ON os.observation_id = n.observation_id
	WHERE c.entity_type = 'observation_note'
UNION
SELECT c.audit_transaction_id, o.subject_id FROM audit_changes c
	JOIN observations o ON o.value_date_id = c.entity_id
	WHERE c.entity_type = 'date_value'
UNION
SELECT c.audit_transaction_id, o.subject_id FROM audit_changes c
	JOIN observations o ON o.value_name_id = c.entity_id
	WHERE c.entity_type = 'name_value';

INSERT OR IGNORE INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id)
SELECT tx_id, 'canonical_entity', entity_id FROM (
	SELECT audit_transaction_id AS tx_id, entity_id FROM audit_changes
		WHERE entity_type = 'canonical_entity'
	UNION
	SELECT audit_transaction_id, unhex(replace(json_extract(changes_json, '$.entity_id.old'), '-', ''))
		FROM audit_changes
		WHERE entity_type = 'identity_claim' AND json_type(changes_json, '$.entity_id') IS NOT NULL
	UNION
	SELECT audit_transaction_id, unhex(replace(json_extract(changes_json, '$.entity_id.new'), '-', ''))
		FROM audit_changes
		WHERE entity_type = 'identity_claim' AND json_type(changes_json, '$.entity_id') IS NOT NULL
	UNION
	SELECT c.audit_transaction_id, ch.entity_id FROM audit_changes c
		JOIN claim_handle ch ON ch.claim_id = c.entity_id
		WHERE (c.entity_type = 'identity_claim' AND json_type(c.changes_json, '$.entity_id') IS NULL)
			OR c.entity_type = 'identity_claim_evidence'
	UNION
	SELECT cs.tx_id, ic.entity_id FROM change_subject cs
		JOIN identity_claims ic ON ic.subject_id = cs.subject_id
		WHERE ic.status IN ('accepted', 'provisional')
)
WHERE entity_id IS NOT NULL AND length(entity_id) = 16;

DROP TABLE change_subject;
DROP TABLE observation_subject;
DROP TABLE claim_handle;
```

No `catalogmodel` change (no table or FK). No delete class change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `Effect.HandleScope Path` *(new)* with comment "handles whose canonical_entity audit scope this change moves; Handles is the recompute set". `observationMembers`, `valueMembers` *(new)* next to `membersOf`; the `observations`, `date_values`, `name_values` `Handles` entries use them. `HandleScope` entries per D-2. `Effect.paths` *(new)*; `validate` uses it in the `None` guard and the check loop. `func HandleScopes(tx *sql.Tx, changes []rowchange.Change) ([][]byte, error)` *(new)* = `resolve(tx, changes, func(e Effect) Path { return e.HandleScope })`, commented "HandleScopes resolves the handles whose canonical_entity audit scope a batch moves." `Set` and `Resolve` are unchanged. |
| `core/database/audit` | `scopes.go`: `ScopeCanonicalEntity` *(new)*, commented "work on a Person, Event, or Place handle: its row, its claims and pins, and evidence on a member Subject"; `scopeResolvers` (D-4); `Querier`, `Latest`, `LatestByScope` (D-5); `Scope`'s comment becomes "Record fills these from scopeResolvers." `audit.go`: `Record` replaces the `effects.Sources` block with a loop over `scopeResolvers`; its doc comment says "plus one scope row per id each scopeResolvers entry resolves (source and canonical_entity)". |
| `core/database/conclusionheaders` | `Updated audit.Latest` *(new)* on `PersonHeader`, `EventHeader`, `PlaceHeader`. `latestHandles` *(new)* in `compose.go`. `ListPersons`, `ListEvents`, `ListPlaces` fill `Updated` (D-6). The private `persons` / `events` / `places` functions are unchanged. |

Errors: none new.

### FFI

`api/proto/engine.proto`: fields per D-7, with one comment on each header: "updated_revision is the latest canonical_entity audit revision (0 = none); updated_at is its RFC3339 UTC time, empty when 0. Filled on the List*Headers responses only." `api/ffi/handlers/conclusion_headers.go`: `personHeaderProto`, `eventHeaderProto`, `placeHeaderProto` set `UpdatedRevision: h.Updated.Revision, UpdatedAt: h.Updated.At`. Regenerate with `./scripts/generate-proto.sh`. No new method.

### macOS

None — no Swift source change. `macos/App/Platform/Generated/engine.pb.swift` changes only through `./scripts/generate-proto.sh` (D-8).

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Handle scope path in `effects`**
Files: `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/database/effects/registry_completeness_test.go`, `core/database/effects/handle_paths_test.go`, `core/database/effects/conclusion_handles_test.go`. Skill: `add-audit-scope`.
Do: add `HandleScope`, `paths`, the named paths, the D-2 entries, and `HandleScopes`; switch `validate` and `hasWork` to `paths()`. Add `mustHandleScopes` beside `mustHandles`. Write T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes (init does not panic; existing subtests and T-1 green).

**STEP-2 — Record stores every scope type; the latest read**
Files: `core/database/audit/scopes.go`, `core/database/audit/audit.go`. Skill: `add-audit-scope`.
Do: add `ScopeCanonicalEntity`, `scopeResolvers`, `Querier`, `Latest`, `LatestByScope`; loop over `scopeResolvers` in `Record` (D-4, D-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/...` passes, including `TestSourceScopes` "promote does not move the source" and `TestResolversCoverEveryEntityType`.

**STEP-3 — Shared seed and parity for each scope type**
Files: `core/database/audit/scopes_test.go`, `core/database/audit/scopes_effects_test.go`, `core/database/audit/export_test.go` *(new)*.
Do: move the captured `userID` and `validLocator` to file level as `var scopeUserID` (the same 16 bytes) and `const scopeLocator` (the same JSON); every case uses them. Move `seed` to file level as `type scopeSeed struct` (same fields); the `tests` table's `run` becomes `func(t *testing.T, s scopeSeed)`, and the name `seed` goes away. Move `mustSeed` to `func newScopeSeed(t *testing.T) scopeSeed` (it still registers `checkNewEffects` in `t.Cleanup` and calls it at the end). Move `revisions`, `latest`, `bumpsA`, `mustCitation`, `textObs` to methods: `(s scopeSeed) sourceRevisions(t) (a, b int64)`, `(s scopeSeed) latest(t) int64`, `(s scopeSeed) bumpsA(t, act func())`, `(s scopeSeed) citation(t, in, obs...) citations.CreateResult`, `(s scopeSeed) textObs(subject, text) observations.Input`. Add `(s scopeSeed) promote(t, subject subjects.Subject) []byte` (runs `promote.Save` through `writes.Call` with action `promote_subject`, returns `Result.Entity.ID`) and `(s scopeSeed) handleRevision(t, id []byte) int64` (reads `audit.LatestByScope(db, audit.ScopeCanonicalEntity)`). `TestSourceScopes` keeps every case, now calling these. In `export_test.go`, `var ScopeResolvers = scopeResolvers`. Rename `loadSourceScopes` to `loadScopes(tx, id, scopeType)`. `checkNewEffects` loops over `audit.ScopeResolvers` for each unseen revision and fails with the revision number and the scope type. Update the `seenEffectsRevisions` comment to "compared to every scope resolver".
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/... -run TestSourceScopes` passes with the same cases as before.

**STEP-4 — End-to-end handle scope tests**
Files: `core/database/audit/scopes_test.go`.
Do: write T-2 and T-3 with `newScopeSeed`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/...` passes.

**STEP-5 — List reads fill `Updated`**
Files: `core/database/conclusionheaders/conclusionheaders.go`, `events.go`, `places.go`, `compose.go`, `conclusionheaders_test.go`.
Do: add the field with its comment and `latestHandles`; fill `Updated` in the three `List*` functions (D-6). Write T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/...` passes, including `TestListPersonsQueryCountIsConstant` and `TestListEventsQueryCountIsConstant` (they count graph queries; `LatestByScope` goes through `c.DB()` and does not change them).

**STEP-6 — Proto and handlers**
Files: `api/proto/engine.proto`, `api/ffi/handlers/conclusion_headers.go`, `api/ffi/handlers/conclusion_headers_test.go`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: add the fields (D-7); run `./scripts/generate-proto.sh`; map them in the three helpers; extend T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes. `git diff --stat` shows no Swift change except `macos/App/Platform/Generated/engine.pb.swift`, and the generated diffs touch only the new fields and their accessors.

**STEP-7 — Backfill migration**
Files: `core/database/migrations/000051.sql` *(new)*. Skill: `add-catalog-migration`.
Do: write the §6 Schema SQL as given.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (parse rules, `verifySchema` on Create and Open, `TestPragmaHonesty`). No migration test is added. The backfill's result is checked in §10 step 2.

**STEP-8 — Docs**
Files: `docs/audit-revision-history.md` §4.1, `.agents/skills/add-audit-scope/SKILL.md`, `core/database/effects/README.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do:
- §4.1: there are two scope types, from `scopeResolvers`: `source`, and `canonical_entity` from `effects.HandleScopes` (claims, pins, the handle row, member Observations and their values and notes). Keep "conclusion-layer changes carry no Source scope"; add that the handle scope is additional.
- Skill: replace "Today the only scope type is `source`" with both types. "New scope type" step 1 becomes "a const and a `scopeResolvers` row", step 2 "an `Effect` path field, added to `Effect.paths`"; name `HandleScope` as the worked example. The Read section names `audit.LatestByScope` as the reader for a list and keeps the single-row SQL for one aggregate.
- Effects README: name `HandleScope` and `HandleScopes` beside `Sources`.
- Map: in the "**Last updated** is an audit scope" working rule, replace the `noScope` sentence with the D-2 list. In the S10-21 entry's **Testable**, replace "A handle with no claims does not appear." with "A handle with no scope rows is listed with revision 0." In the S10-22 entry's **In**, add: "A note-only Observation save reports no handles (`observation_notes` has no `Handles`), so `CatalogQueryRegistry` does not reload the lists. This PR reports `effects.HandleScopes` ids on `WriteEffects` and invalidates the three lists on them, so the cell moves." In the slice 6 PR sequence, mark the S10-21 line "— done, #N". In the S10-41 entry's **In**, add: "The time comes from `audit.LatestByScope(db, audit.ScopeSource)`, and this PR removes the `sources` scope reads it makes dead."
- Skill "Do not": after "Add a hand-written resolver map in `scopes.go`", add "(`scopeResolvers` lists scope types, one row each; it is not an entity map)".
- Then the Definition of done lines.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-21-handle-last-activity.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Record when each Person, Event, and Place was last worked on."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/conclusion_handles_test.go` `TestConclusionHandleSets` (extend) and `handle_paths_test.go` `TestHandlePaths` (extend) | In "cardinality matches property handles": `mustHandleScopes` on the same changes → empty. In "grade change matches source handles": → empty. In "accepted mint covers claim handle, filings, and observers": `mustHandleScopes(t, tx.Tx, changes...)` = `{entity.ID}` (Bob's new handle) and does not contain `adaPromoted.Entity.ID`. In "transcription_uncertain only": the uncertain change → empty. New subtest "observation note reaches member handle": `rowchange.Change{EntityType: "observation_note", EntityID: <new UUIDv7>, Action: rowchange.ActionCreate, Fields: map[string]rowchange.FieldDiff{"observation_id": {New: dashed(created.Observations[0].ID)}}}` (Boston's date Observation) → `{bostonRes.Entity.ID}`. New subtest "subject move scopes both handles": `observationSubject(bob.ID)` with its `subject_id` diff set to `{Old: dashed(alice.ID), New: dashed(bob.ID)}` → `{aliceRes.Entity.ID, bobRes.Entity.ID}`. New subtest "rejected member is not scoped": `observationSubject(cara.ID)` → empty | AC-3.1, AC-3.2, AC-3.3, D-1, D-2, D-3 |
| T-2 | Go | `core/database/audit/scopes_test.go` `TestHandleScopes` *(new)* | Table, each case on `newScopeSeed`. `observations.Update` inputs copy the full `observations.Input` from `TestSourceScopes`' update cases; dates use `s.dateProp`: promote `person` → `handleRevision` = latest, Source A unchanged; promote `place`, then `citation` with toponym "Boston" on it → handle and Source A at latest; `observations.Update` to "Boston Town" → both; `observations.Update` with `Notes: {"faded"}`, then `{"faded ink"}` → both each time; promote `event`, `citation` with a 1842 date, update to 1843 → both; `observations.Delete` of the noted Observation → both; `citation` with toponym on `place2` (no claim) → Source A moves, `place` handle unchanged; promote `person`, then `subjects.Delete` it → its handle and Source A at latest | AC-1.1, AC-1.2, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-3, D-4 |
| T-3 | Go | `core/database/audit/scopes_test.go` `TestLatestByScope` *(new)* | `newScopeSeed`; promote `person` and update one of its Observations; raw `INSERT INTO canonical_entities (id, subject_type_id, ref)` with a new UUIDv7, the person type id, and a ref from `ref.Mint` with the person type's prefix (`subject_types.ref_prefix`) → map has the promoted handle at its latest revision and that revision's `created_at`; raw handle absent; `LatestByScope(db, "")` → `ErrInvalid` | AC-1.3, D-5 |
| T-4 | Go | `core/database/conclusionheaders/conclusionheaders_test.go` `TestListHeadersCarryUpdated` *(new)* | `newFixture`; first, before any list or promote, the raw-inserted Person handle (as T-3), so the graph is not loaded yet. Then `f.person("Bob Smith")`, `f.person("Ann Lee")`, `f.event` with the inputs `TestListEvents` passes for `fire`, and `f.place(art, "Boston")` with `art` from `f.book("Gazetteer")` → each `List*` row has `Updated.Revision` > 0 and `Updated.At` = that revision's `created_at`. `f.list()` once (warms the display cache), then `observations.Update` on Bob's Observation with `Notes: {"seen"}`, then `f.list()` → Bob's row has the new revision, Ann's is unchanged. The raw-inserted handle → listed with zero `Updated`. `PersonsByIDs` → zero `Updated` | AC-1.3, AC-4.1, AC-4.2, D-6 |
| T-5 | Go | `api/ffi/handlers/conclusion_headers_test.go` `TestListPersonHeaders`, `TestListEventHeaders`, `TestListPlaceHeaders` (extend) | In each test's case that already promotes rows and lists them, every response row has `updated_revision` > 0 and a non-empty `updated_at` | AC-4.1, D-7 |
| T-6 | Go | `core/database/audit/scopes_effects_test.go` `checkNewEffects` (extend) | Every revision `TestSourceScopes` and `TestHandleScopes` write: stored rows of each `ScopeResolvers` type = that resolver's ids | AC-2.5, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSourceScopes` (seed move and `Record` change), `TestResolversCoverEveryEntityType`, `TestCascadeIntoNonEmptyEffectIsAudited` (`hasWork`), `TestHandlePaths`, `TestConclusionHandleSets`, `catalogmodel.TestPragmaHonesty`, and the `verifySchema` path in `core/database` tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: the slice 6 sequence line, the working-rule sentence, the S10-21 Testable line, the S10-22 and S10-41 lines; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/audit-revision-history.md §4.1, add-audit-scope skill, effects README
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Copy a dogfood project folder that has promoted Persons, Events, and Places. Build and run the app from this branch, and open the copy. It opens with no `catalog.schema_mismatch` (the migration ran). This is AC-4.3.
2. Quit the app. On the copy only, run `sqlite3 <copy>/provenencia.sqlite` with:
   `SELECT COUNT(*) FROM canonical_entities e WHERE e.merged_into_id IS NULL AND NOT EXISTS (SELECT 1 FROM audit_transaction_scopes sc WHERE sc.scope_type = 'canonical_entity' AND sc.scope_id = e.id);` → `0` (every handle's create change was mapped).
   `SELECT c.entity_type, COUNT(DISTINCT c.audit_transaction_id) FROM audit_changes c JOIN audit_transaction_scopes sc ON sc.audit_transaction_id = c.audit_transaction_id AND sc.scope_type = 'canonical_entity' GROUP BY c.entity_type;` → rows for `canonical_entity`, `identity_claim`, and `observation`, each with a count > 0.
   Pick one Person you edited recently; its `MAX(revision)` under `canonical_entity` is at or after that edit's revision.
3. Reopen the copy. Open People, Events, and Places: the rows, their order, and their counts are the same as on `main`. Open Sources: the "Updated" sort is the same as on `main`.
4. Open a promoted Person, go to a Source where that Person is a member, edit one of its Observations, and save. Return to People: the list loads with no error. (The value is not drawn until S10-22. T-4 and T-5 are the evidence for the field.)

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A delete reaches a handle only through live claims that are already gone | Claim deletes carry `entity_id` (D-3). T-2 "delete member Subject" and T-6 parity catch a miss. |
| Stale revision from the display cache | `Updated` is filled on copies after the cache read and never stored (D-6). T-4 adds a note, which moves the scope without dropping the cached row, and lists again. |
| Backfill maps old work through today's membership | Accepted in D-9 and written in the migration comment. |
| Moving the seed changes `TestSourceScopes` | STEP-3's check runs it alone with the same cases before any new test is added. |

**Stop and report if:**

- `audit.Record` no longer calls `effects.Sources`, or `writes.Run` no longer calls `audit.Record` before `effects.Resolve`.
- A member Subject delete in T-2 does not emit `identity_claim` delete changes carrying `entity_id`, so the handle cannot be reached without a new path or a hand resolver.
- `checkNewEffects` parity fails for a revision whose stored handle scope is right but whose re-resolve differs because a live row changed later. Report the revision and the write; do not loosen the check.
- Switching the `Handles` entries to `observationMembers` / `valueMembers` changes any existing `TestHandlePaths`, `TestConclusionHandleSets`, or `TestEvidenceHandleSets` result.
- `verifySchema` fails after adding `000051.sql` (the migration must not change the schema), or the migration fails on a catalog because `unhex` or `json_type` is missing.
- Manual check step 2's first query is not `0`.
- `./scripts/generate-proto.sh` fails because a generator is missing, or its diff rewrites `engine.pb.go` / `engine.pb.swift` beyond the new fields (a different generator version).

## 12. Open questions

None.

## 13. PR

**Title:** S10-21 — Handle last-activity scope
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3, and US-4; AC evidence (test names); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-21-handle-last-activity.md`.

## Landed

Differences from this plan: T-3 (`TestLatestByScope`) promotes a place Subject and edits its toponym, not a person's, because a toponym is not valid on a person. T-4 (`TestListHeadersCarryUpdated`) builds the Event from a single "Fire" `event_name` Observation, not the inputs `TestListEvents` passes for `fire`. The implementer set Status and the map row to Landed ahead of the merge. The manual check was not run by the implementer; it is for the reviewer. Review round 1 added: the unused `audit.Scope` type was deleted; `.cursor/rules/catalog-audit.mdc` and the audit doc's section 4.1 sentence now describe the effect-path resolvers and the handle scope; `TestLatestByScope` also checks a nil querier returns `ErrInvalid`; `TestHandlePaths` has an "evidence pin reaches its handle" subtest; `TestSourceScopes` uses the seed's `updateObservation`, `deleteObservation`, and `promote` helpers.
