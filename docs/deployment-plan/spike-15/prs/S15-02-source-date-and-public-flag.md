# S15-02 — Sources carry a date and a not-public flag

**Status:** Draft
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: Go and the wire only; no Swift view reads the new fields until S15-03 and S15-04)
**Board:** —
**Depends on:** —
**Blocks:** S15-03, S15-04, S15-05, S15-16, S15-32
**Branch:** `s15-02-source-date-and-public-flag`
**Version bump:** minor — additive schema (one migration that old apps refuse by `user_version`, as every format step) and two new RPCs ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/catalog-model.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/date-values.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "The Source date is one optional DateValue on `sources`", "Not public is a stored switch and nothing else", "One date order" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`source-provenance-date.md`](../source-provenance-date.md); [privacy note](../2026-10-27-source-privacy-toggle.md) "Likely code locations"; [`source-layer-data-model.md`](../../../source-layer-data-model.md) § 4 and § 4.2; [`structured-date-model.md`](../../../structured-date-model.md)
**Planned:** 2026-11-09 against `main` @ `382e85f`
**Review:** [`S15-02-source-date-and-public-flag.review.md`](S15-02-source-date-and-public-flag.review.md) — round 1 answered; round 2 pending

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

A Source gains two facts of its own: an optional date (the original publication, or when the record was made) stored as a shared DateValue, and a not-public flag that defaults to public. The engine stores, reads, writes, audits and deletes both, and every Source on the wire carries the date, a date sort key from the one date order the Timeline already uses, and the flag. This is the engine half of dogfood bars 1 and 2; the Source page (S15-03), the Sources list (S15-04) and Copy Source (S15-32) build on it.

**Why now.** Every later Source-date PR, and the citation renderer's `{source.date}` (S15-16), reads what this PR stores. The flag is the foundation the product owner asked for on 2026-10-27.

**What stays.** No Swift view changes; the generated Swift proto is the only Swift diff. `sources.Update` and `UpdateSourceRequest` keep their fields and behavior. Copy Source copies neither field (the copy is undated and public) until S15-32. The Sources list keeps its newest-first order from Go; Swift sorts by the new key in S15-04. Metadata date fields stay text.

## 2. User stories

**US-1.** As a researcher, when I give a Source a date, change it or clear it, the catalog keeps exactly that date, so the Source page and list can show it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An undated Source; `SetSourceDate` with point `1880` (year only) | `ListSources` and `GetSourceWorkspace` return that Source with `date` kind `point`, `start_year` 1880; one `date_values` row exists for it |
| AC-1.2 | That Source; `SetSourceDate` with `ABT 12 Mar 1880` | The date reads back as `ABT 12 Mar 1880`; `sources.date_value_id` is the same id as before (edited in place) |
| AC-1.3 | That Source; `SetSourceDate` with no date | The Source reads back undated, `date_sort_key` is absent, and its `date_values` row is gone |
| AC-1.4 | `SetSourceDate` with point, no year, month 13 | Error `datevalues.invalid`; the stored date is unchanged |
| AC-1.5 | A dated Source with no Artifacts or Subjects; `DeleteSource` | The Source and its `date_values` row are both gone |
| AC-1.6 | `SetSourceDate` with the date it already has | No revision is written; the response's effects are empty |

**US-2.** As a researcher, when I mark a Source not public, or public again, the catalog keeps it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source created by `CreateSource`, or one that existed before the migration | `is_public` is `true` |
| AC-2.2 | `SetSourcePublic` with `is_public: false`, then a fresh `GetSourceWorkspace` | `is_public` is `false` |
| AC-2.3 | `SetSourcePublic` with `true` on a Source that is public | No revision is written |

**US-3.** As the app, I order Sources by date with a key Go computes, so Swift never orders dates itself.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Sources dated `1879`, `bef. 1880`, `1880`, `1881`, `bef. 29 Mar 2011`, `29 Mar 2011`, `aft. 29 Mar 2011`, and one undated | Ordering by `date_sort_key` ascending, absent last, gives exactly that order (the Timeline's `TestEventOrderAtKnownEnd` order) |
| AC-3.2 | `ABT 1880` and `1880` | Equal keys |
| AC-3.3 | Range `BET 1885 AND 1880` (written backwards) | Same key as `BET 1880 AND 1885` |
| AC-3.4 | A phrase-only date ("in the reign of Victoria") | The date is returned; `date_sort_key` is absent |

**US-4.** As a contributor, when a Source's date or flag changes, its "Updated" activity moves, as for any other work under the Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Set, edit in place, and clear a Source's date; set its flag | Each write moves that Source's scope revision and no other Source's |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration `000063.sql`: `sources.date_value_id`, `sources.is_public`, an index on `date_value_id` | STEP-3, AC-2.1 |
| The date order moves to `core/valuecompare` as `DateKey`; `conclusionsummaries` calls it | STEP-1, AC-3.1, AC-3.2, AC-3.3, D-4 |
| `datevalues.PutTx` and the date audit changes, `deleteimpact.ReleaseReplacedRow`; `observations` uses them | STEP-2, D-6, D-7 |
| `catalogmodel` FK; `date_values` effect reaches the Source | STEP-3, STEP-5, AC-4.1 |
| `sources` reads carry date and flag (one batched date lookup for the list); `SetDate`, `SetPublic`, `DateSortKey`; Delete releases the date | STEP-4, STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3 |
| FFI: `Source` fields, `SetSourceDate`, `SetSourcePublic`; generated Go and Swift | STEP-6, STEP-7, all ACs |
| Docs and skills the PR makes false (STEP-8 list) | STEP-8 |

| Out | Home |
| --- | --- |
| The Source page date entry and the switch | S15-03 |
| The Sources list date column and the date sort | S15-04 |
| The Person page's Sources section date | S15-05 |
| Copy Source carrying the date, with a tick on the Copy form | S15-32 (user, 2026-11-09) |
| Copy Source carrying the not-public flag | not planned (working rule "Not public is a stored switch") |
| Any filter, glyph, hide or cascade on not public | not planned (working rule) |
| `CatalogSource`, `GenealogyStore`, `GoStore`, `FakeStore` fields and methods for the date and flag | S15-03 (flag and date write), S15-04 (sort key) |
| `autoreconcile.SortKey`'s date case moving onto `valuecompare.DateKeyOf` (a second, older date key stored in `auto_reconciler_values.sort_key`) | not planned in this PR (review M-6, To user) |
| The date in the omnibar search document | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Schema | `core/database/migrations/000062.sql` | Last step. `sources` is `id, ref, source_type_id, title, description, cover_mode, primary_artifact_id` (`000006.sql`, `000011.sql`, `000015.sql`); index pattern `000031.sql` `sources_primary_artifact_id_idx` |
| Model | `core/database/catalogmodel/model.go` `FKs` | `observations.value_date_id → date_values` is `BucketOwnedOutbound`; `date_values` table is `BucketOwnedOutbound` with PK `id`. No FK from `sources` to `date_values` |
| Source row | `core/database/sources/sources.go` `Source`, `sqlGet`, `sqlGetByRef`, `sqlList`, `scanSource`, `scanSourceList` | No date, no flag. `List` is one query, newest first, then `Graph().Activity` per row |
| Source writes | `sources.go` `Create`, `Update`, `SetCover`, `Delete` | `SetCover` is the precedent for a one-attribute write returning `(Source, []rowchange.Change, error)`. `Delete` calls `deleteimpact.ReleaseFacets` but takes no `SnapshotOwned`, because `sources` has no owned FK yet |
| Date order | `core/database/conclusionsummaries/compose.go` `dateKey`, `dateKeyOf`, `dateKeyLess`, `placedDate`, `datePlace`, `eventDating`; `order.go` `PlaceBirth`, `sortByDateKey` | The one rule (key day plus a BEF/closed/AFT side) is private here |
| Date ends | `core/valuecompare/datewindow.go` `DateEnds`, `DateEnd` | Pure; `datePlace` calls `DateEnds` |
| Date queries | `core/database/datevalues/datevalues.go` `InsertTx`, `UpdateTx`, `LookupTx`, `LookupManyTx`, private `scanValue` | `LookupManyTx` is the batched read three list readers use (`core/database/graphfill/load.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/autoreconciler/autoreconciler.go`) |
| Date ownership | `core/database/observations/observations.go` `resolveValue` (`ValueTypeDate` case), `dateCreateChange`, `dateUpdateChange`, `dateValueMap`, `intPtrAny`; `update.go` `releaseReplacedValue` | Private to `observations`: insert or update in place with its audit change, and release with an audited delete on replace. `dateCreateChange` re-spells `dateValueMap`'s 21 keys. The Source needs the same job |
| Second date key | `core/autoreconcile/autoreconcile.go` `SortKey` (`ValueTypeDate`), `core/autoreconcile/dates.go` `dateSortOffset` | An older key from `DateWindow` with no side tie-break, stored in `auto_reconciler_values.sort_key`; not changed here (Out) |
| Effects | `core/database/effects/registry.go` `"date_values": ownedValue("date_value", "value_date_id")`; `"sources"` entry | A `date_value` change reaches a Source only through `observations.value_date_id`, so an in-place Source date edit would scope to nothing |
| Owned release | `core/database/deleteimpact/release.go` `SnapshotOwned`, `ReleaseSnapshot`, `ReleaseReplaced` | Derived from `catalogmodel` owned FKs; `artifacts.Delete` is the snapshot precedent |
| Wire | `api/proto/engine.proto` `Source` (fields 1–13), `DateValueInput`, `Method` (last 116), `SetSourceCoverRequest/Response` | No date or flag |
| Handlers | `api/ffi/handlers/sources.go` `sourceProto`, `ListSources`, `GetSourceWorkspace`, `SetSourceCover`; `api/ffi/handlers/thumbnails.go` `enrichSourceProto`; `api/ffi/dispatch.go` `MethodSetSourceCover` | Every Source on the wire goes through `sourceProto` |
| Codec | `core/valuecodec/valuecodec.go` `DateToProto`, `DateFromProto` | Reused as is |
| Copy | `core/database/sourcecopy/sourcecopy.go` `ReadDraft`, `CreateFromDraft` | Builds `sources.CreateInput`; unaffected |
| Tests | `core/database/sources/sources_test.go` `TestSources`, `TestSourceDelete`; `api/ffi/handlers/sources_test.go` `TestListSources`, `TestGetSourceWorkspace`, `TestDeleteSource`, `TestCopySource`; `core/database/audit/scopes_test.go` `TestSourceScopes`; `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestEventOrderAtKnownEnd`, `TestSortByBirth`; `core/valuecompare/datewindow_test.go` `TestDateEnds`; `api/ffi/handlers/write_effects_test.go` `TestEveryAuditedWriteReturnsEffects` | — |

**Precedent.** `000015.sql` plus `sources.SetCover` and the `SetSourceCover` RPC added a per-Source attribute end to end: copy the migration shape, the one-attribute write and its response (`Source` plus `WriteEffects`). `observations.resolveValue` (date case) and `update.go` `releaseReplacedValue` own a DateValue: this PR moves that job into `datevalues.PutTx` and `deleteimpact.ReleaseReplacedRow` (D-7) so both owners call it. `000030.sql` `observations_value_date_id_uidx` is the owned-column index shape. `artifacts.Delete` snapshots owned children before its `DELETE`: copy that in `sources.Delete`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One migration `000063.sql`: `ALTER TABLE sources ADD COLUMN date_value_id BLOB REFERENCES date_values(id);`, `ALTER TABLE sources ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1));`, `CREATE UNIQUE INDEX sources_date_value_id_uidx ON sources(date_value_id) WHERE date_value_id IS NOT NULL;` | Working rules; defaults give every existing Source undated and public; the honesty tests require a covering index on an FK child column; every owned value column is unique (`000030.sql`, `000052.sql`), so two Sources can never share a date row | A plain index — lets a bug share one row between owners; a table rebuild — not needed |
| D-2 | `sources.date_value_id → date_values` is registered `NO ACTION`, `BucketOwnedOutbound` | The Source owns its date as an Observation owns its value; `deleteimpact` then derives the release | `SET NULL` / a facet — leaves orphan `date_values` rows |
| D-3 | Two new RPCs, `SetSourceDate` (117) and `SetSourcePublic` (118), each returning `Source` and `WriteEffects`. A `SetSourceDateRequest` with no `date`, or a `date` whose `kind` is blank after trim, clears, whatever else the `date` holds (the same test `valueInput` in `api/ffi/handlers/observations.go` uses); otherwise the handler passes `&v` from `valuecodec.DateFromProto` | Mirrors `SetSourceCover`; the Source page writes each on its own control; `UpdateSourceRequest` is a whole-row write whose proto3 `bool` default would mark every Source not public when an old caller omits it | New fields on `UpdateSourceRequest` — the default-false and omitted-date problem |
| D-4 | The order rule moves to `core/valuecompare/dateorder.go` *(new)*: `type DateKey struct{ Day, Side int }`; `func DateKeyOf(d datevalue.Value) (key DateKey, first, last DateEnd, ok bool)`, holding `datePlace`'s body without the span (ok false for nil-year, both-ends-open, or a closed end outside the year window); `func (k DateKey) Compare(o DateKey) int` (Day, then Side; −1, 0, 1). In `conclusionsummaries/compose.go`: `placedDate` becomes `{at valuecompare.DateKey; span TimelineSpan; first, last valuecompare.DateEnd}`; `dateKey` becomes `{at valuecompare.DateKey; ok bool; ref string}`; `datePlace` calls `DateKeyOf` and builds the span from `at.Side` as today; the three `dateKey{key: …, side: …}` literals (`eventDating` twice, `dateKeyOf`) become `dateKey{at: p.at, ok: true}`; `dateKeyLess` is: `ok` first, then `a.at.Compare(b.at)`, then `ref`. `order.go`'s text changes only where it names `p.key` / `p.side`; its behavior does not | Working rule "One date order"; `TimelineSpan` and the ref tie-break stay Conclusion concerns. No integer encoding here: `core/valuecompare` is a numbers-guarded layer (`core/matchregistry/guard_test.go` `checkNumbers`) | An `Int64` method on `DateKey` — its multiplier is a literal the guard refuses (review X-1) |
| D-5 | Wire: `Source` gains `DateValueInput date = 14`, `optional int64 date_sort_key = 15`, `bool is_public = 16`. Every Source sent goes through `sourceProto`, which sets all three. `date_sort_key` is absent when the Source is undated or `DateKeyOf` is not ok | One reader path; the key is a wire value so S15-04 sorts without date logic; `is_public` matches the column name | `bool not_public` — a second name for one fact |
| D-6 | `sqlGet`, `sqlGetByRef` and `sqlList` select `s.date_value_id, s.is_public` after today's columns; `scanSource` and `scanSourceList` scan them (`is_public` into an `int64`, `IsPublic = v != 0`). `Get`, `getByRef` and `getTx` then call `datevalues.LookupTx` when `DateValueID` is 16 bytes. `List`, after its scan and before the activity loop, collects every `DateValueID` and calls `datevalues.LookupManyTx(db, ids)` once, then sets each `Date` (with `ID`) from the map | The existing batched read with three list callers: at 3,000 Sources it is 6 queries of 500 ids, never one per row; no new scan type | A `LEFT JOIN` with a new `datevalues.Joined` scanner — a second conversion of the 20 columns beside `scanValue` (review M-2) |
| D-7 | The owned-date job moves out of `observations`. In `datevalues` *(new)*: `RowMap(id []byte, v datevalue.Value) map[string]any` (today's `dateValueMap`), `CreateChange(id []byte, v datevalue.Value) rowchange.Change` = entity `date_value`, `ActionCreate`, `rowchange.FullRow(RowMap(id, v))`, `UpdateChange(prev, next datevalue.Value) *rowchange.Change` (today's `dateUpdateChange`, nil when no field differs), and `PutTx(tx *sql.Tx, prevID []byte, v datevalue.Value) (id []byte, ch *rowchange.Change, err error)`: with a 16-byte `prevID` it `LookupTx`s the previous value, returns `prevID` and nil when `UpdateChange` is nil (no write), else `UpdateTx` and the update change; with no `prevID` it `InsertTx`s and returns the create change. In `deleteimpact` *(new, today's `observations.releaseReplacedValue` exported)*: `ReleaseReplacedRow(tx *database.Tx, column, entity string, prevID, newID []byte, rowMap func() (map[string]any, error)) (*rowchange.Change, error)`. `observations.resolveValue`'s date case calls `PutTx`; `observations` `Update`'s release loop calls `ReleaseReplacedRow` for all three value kinds; `dateCreateChange`, `dateUpdateChange`, `dateValueMap`, `releaseReplacedValue` and the now-uncalled `intPtrAny` are deleted. `datevalues` and `deleteimpact` do not import each other | One home for owning a DateValue: reuse the id on edit, audit the update, release and audit on replace; S15-32 and any later owner call it (review M-1, M-5) | Moving only the change builders — leaves the write copied in `sources` |
| D-8 | `sources.SetDate(tx *database.Tx, userID, sourceID []byte, date *datevalue.Value) (Source, []rowchange.Change, error)` *(new)*, after `RequireUserID` and `getTx` (missing Source or short id → `ErrInvalid`). With a date: `datevalues.PutTx(tx.Tx, prev.DateValueID, *date)` (invalid → `datevalue.ErrInvalid`, nothing written); when the Source had no date, `UPDATE sources SET date_value_id = ?` and a `source` update change `date_value_id` nil→`database.UUIDString(id)`; the `PutTx` change, when not nil, goes first. Without a date and with a stored one: `UPDATE sources SET date_value_id = NULL`, a `source` update change id→nil, then `deleteimpact.ReleaseReplacedRow(tx, "sources.date_value_id", "date_value", prev.DateValueID, nil, rowMap)` with `rowMap` = `datevalues.LookupTx` then `datevalues.RowMap`. Without a date and none stored: no changes. It returns the Source re-read with `getTx` (D-6), so `Date.ID` is set | The Observation precedent through the shared helpers; a stable id across edits | Always insert a new row — churns ids and doubles audit rows |
| D-9 | `effects.ownedValue(entity, col string, owners ...Path)` takes further owner paths, unioned into `SourceScope` beside `valueSource(col)`; the registry line becomes `"date_values": ownedValue("date_value", "value_date_id", inbound("sources", "date_value_id"))`. An in-place edit reaches the Source through it. On clear the `date_value` delete reaches no Source (the Source already holds NULL); the `source` update change in the same write carries the scope | Without it an edit moves no Source's activity (AC-4.1); one registry shape for every owned value (review M-7) | A hand-built Effect for `date_values`; a hand-written `source` change on every edit (`add-audit-scope` forbids hand tagging) |
| D-10 | `sources.SetPublic(tx, userID, sourceID []byte, isPublic bool) (Source, []rowchange.Change, error)` *(new)*: one `source` update with `is_public` old→new as `0`/`1`; unchanged returns no changes | Mirrors `SetCover`; the column is an integer | — |
| D-11 | `sources.Delete` takes `deleteimpact.SnapshotOwned(tx.Tx, catalogmodel.KindSource, id)` after `ReleaseFacets` and before its `DELETE`, then `deleteimpact.ReleaseSnapshot`. The delete change adds `date_value_id` when set and always `is_public` | `artifacts.Delete` precedent; AC-1.5 | — |
| D-12 | `Source` gains `DateValueID []byte`, `Date *datevalue.Value`, `IsPublic bool`. `Create` returns `IsPublic: true`. `Update` neither reads nor writes them | One struct for every read | — |
| D-13 | `sources.DateSortKey(s Source) (int64, bool)` *(new)*: false when `s.Date` is nil or `DateKeyOf` is not ok; else `int64(k.Day)*3 + int64(k.Side) + 1`, which keeps `Compare`'s order in one integer since `Side` is −1, 0 or 1. `sourceProto` calls it | The wire encoding is the Source read's concern, outside the numbers-guarded layer (X-1) | Calling `valuecompare` from the handler |
| D-14 | Copy Source is unchanged: a copy is undated and public | S15-32 adds the date tick (user, 2026-11-09); the flag is never carried (working rule) | — |
| D-15 | No migration test; the default read is tested through `sources.Create` and `List` (AC-2.1) | `add-catalog-migration` step 4 | An upgrade test |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000063.sql` *(new)* | migration | `000015.sql` | A new step is the only way to change the schema |
| `core/valuecompare/dateorder.go` *(new)*: `DateKey`, `DateKeyOf`, `Compare` | file / type / func | `conclusionsummaries` `datePlace`, `dateKeyLess` | Moved, not copied: the working rule names this home |
| `datevalues.PutTx`, `CreateChange`, `UpdateChange`, `RowMap` *(new, moved)* | func | `observations` `resolveValue` date case, `dateCreateChange`, `dateUpdateChange`, `dateValueMap` | Moved so both owners share one job; `CreateChange` is `FullRow(RowMap(…))`, so the 21 keys exist once |
| `deleteimpact.ReleaseReplacedRow` *(new, moved)* | func | `observations` `releaseReplacedValue` | Exported so a Source and an Observation release a replaced value the same way |
| `ownedValue(…, owners ...Path)` *(new parameter)* | registry helper | `effects.ownedValue` | Extended, not bypassed |
| `sources.SetDate`, `sources.SetPublic`, `sources.DateSortKey` *(new)* | func | `sources.SetCover` | Same shape; one attribute each |
| `SetSourceDate`, `SetSourcePublic` *(new)* | FFI | `SetSourceCover` | D-3 |
| `Source.date`, `date_sort_key`, `is_public` *(new)* | proto fields | `Source.cover_mode` | Additive |
| `TestDateKeyOf` *(new)* | test | `TestDateEnds` | Same file family and table style |
| `TestSourceDate`, `TestSourcePublic` *(new)* | test | `TestSources` | New tables beside it in `sources_test.go` |
| `TestSetSourceDate`, `TestSetSourcePublic` *(new)* | test | `TestListSources` | `runRPC` tables in `api/ffi/handlers/sources_test.go` |

Searched for: another date sort key (`sortKey`, `DateWindow` "Sort keys" comment in `datewindow.go`, `dateKey` across `core/`), the batched date read (`LookupManyTx`, three callers), another sort key stored for dates (`autoreconcile.SortKey`, Out), another one-attribute Source write (`SetCover`), another owner of a `date_values` row (Observations only), an existing privacy flag (none, per the privacy note).

**Builds on.** `catalogmodel` owned-outbound registration and `deleteimpact`'s derived release; the effects registry's `union` / `inbound` paths; `valuecodec` for the wire date; `writes.Run` and `writeEffectsProto`.
**Deletes.** `conclusionsummaries` loses the key-and-side body of `datePlace` and compares through `valuecompare.DateKey.Compare`. `observations` loses `dateCreateChange`, `dateUpdateChange`, `dateValueMap`, `releaseReplacedValue`, `intPtrAny` and its inline insert-or-update date code in `resolveValue`.
**Next adds.** S15-04 sorts by `date_sort_key`; S15-03 calls the two RPCs; S15-05 reads `Source.Date` into the Source record; S15-16 reads the date for `{source.date}`; S15-32 adds a `copy_date` choice to `sourcecopy` that calls `SetDate` on the copy.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A `sources` row that owns a `date_values` row | `sources.Delete`, `deleteimpact` honesty tests | D-11 releases it; T-5 |
| A Source deleted while dated | `effects` resolve on the `source` delete change | The change carries `date_value_id` (D-11); the date row's own release is not audited, as for an Observation |
| A `date_value` change whose owner is a Source | `effects` `date_values` entry | D-9; T-8 |
| Copy of a dated, not-public Source | `sourcecopy.CreateFromDraft` → `sources.Create` | Copy is undated and public (D-14); T-6 |

### Scale

At the map's Scale (3,000 Sources): `sources.List` stays one query for the rows, then one `datevalues.LookupManyTx` over the dated rows' ids: 6 queries of 500 ids at the worst case where every Source is dated (`sqlbatch.Size` is 500), and one `DateKeyOf` (constant work) per row in `sourceProto`. No per-row query. Each write touches one Source: one Source scope, no Evidence graph, no list count; `SetSourcePublic` reprojects that one Source's search document through the existing `sources` effect. No budget is set or leaned on.

### Edges

- **Assumptions.** `dateKeyLess` assumes a total order on (ok, Day, Side, ref); `Compare` keeps Day then Side, and `TestEventOrderAtKnownEnd` and `TestSortByBirth` hold it unchanged. `DateSortKey` assumes `Side` is −1, 0 or 1; T-10 checks its integer order equals `Compare` on T-1's rows.
- **Neighbors.** `bef. X` < `X` < `aft. X` on one day; `X` year-only equals `1 Jan X`; `ABT X` equals `X` (as written, not widened: `DateWindow`'s rule); a range open at its start equals `BEF` its end; reversed range equals the forward one. Each is a T-1 row.
- **Removed inputs.** Clearing a date deletes its row (AC-1.3); deleting a Source deletes it (AC-1.5). Nothing else reads a Source's `date_values` row.
- **Real data.** None — no table, seed or configuration drives the rule.
- **Untrusted size.** None — one date per Source; the list is bounded by the catalog's Sources.

### Schema

Migration `core/database/migrations/000063.sql` *(new)*:

```sql
-- S15-02: a Source's own date (original publication, or when the record was made) and a not-public flag.
ALTER TABLE sources ADD COLUMN date_value_id BLOB REFERENCES date_values(id);
ALTER TABLE sources ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1));
CREATE UNIQUE INDEX sources_date_value_id_uidx ON sources(date_value_id) WHERE date_value_id IS NOT NULL;
```

Registered in `catalogmodel.FKs`: `{From: "sources", Column: "date_value_id", To: "date_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound}`. Delete class: owned outbound (D-2). Audit: `source` and `date_value` changes (D-8 – D-10).

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `dateorder.go` *(new)* per D-4 |
| `core/database/conclusionsummaries` | `compose.go` `datePlace`, `placedDate`, `dateKey`, `dateKeyOf`, `dateKeyLess`, `eventDating` use `valuecompare.DateKey` (D-4); `order.go` unchanged in behavior |
| `core/database/datevalues` | `PutTx`, `CreateChange`, `UpdateChange`, `RowMap` (D-7) |
| `core/database/deleteimpact` | `ReleaseReplacedRow` (D-7) |
| `core/database/observations` | Calls `PutTx` and `ReleaseReplacedRow`; private copies deleted (D-7) |
| `core/database/catalogmodel` | One FK row (D-2) |
| `core/database/effects` | `ownedValue` owners parameter and the `date_values` line (D-9) |
| `core/database/sources` | `Source` fields (D-12); reads and the batched date lookup (D-6); `SetDate` (D-8); `SetPublic` (D-10); `DateSortKey` (D-13); `Delete` (D-11) |

Errors: `sources.invalid` → missing Source or bad id; `datevalues.invalid` → a date `datevalue.Validate` refuses. Both are mapped in `L10n.Errors` today.

### FFI

`api/proto/engine.proto`:

```proto
// in enum Method
METHOD_SET_SOURCE_DATE = 117;
METHOD_SET_SOURCE_PUBLIC = 118;

// in message Source
// The Source's own date (original publication, or when the record was made); absent when undated.
DateValueInput date = 14;
// Its place in date order (valuecompare.DateKey); absent when undated or not placeable. Sort ascending, absent last.
optional int64 date_sort_key = 15;
// False when the researcher marked the Source not public. Stored only.
bool is_public = 16;

message SetSourceDateRequest {
  string project_dir = 1;
  string user_id = 2;
  string source_id = 3;
  DateValueInput date = 4; // absent or blank kind clears
}
message SetSourceDateResponse { Source source = 1; WriteEffects effects = 2; }

message SetSourcePublicRequest {
  string project_dir = 1;
  string user_id = 2;
  string source_id = 3;
  bool is_public = 4;
}
message SetSourcePublicResponse { Source source = 1; WriteEffects effects = 2; }
```

Handlers `SetSourceDate` and `SetSourcePublic` in `api/ffi/handlers/sources.go`, shaped as `SetSourceCover`'s non-thumbnail branch: `withProjectCatalog`, `writes.Run` with actions `set_source_date` / `set_source_public`, then `enrichSourceProto`. `sourceProto` fills `Date` (`valuecodec.DateToProto`), `DateSortKey` (`sources.DateSortKey`) and `IsPublic`. `api/ffi/dispatch.go`: `MethodSetSourceDate`, `MethodSetSourcePublic` and two `case`s. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None — no view changes. `macos/App/Platform/Generated/engine.pb.swift` is regenerated and committed; nothing reads the new fields yet.

### Strings

None — no new copy and no new error code.

## 7. Steps

**STEP-1 — Move the date order to `valuecompare`**
Files: `core/valuecompare/dateorder.go` *(new)*, `core/valuecompare/dateorder_test.go` *(new)*, `core/database/conclusionsummaries/compose.go`. Skill: `add-date-value`.
Do: D-4. Write T-1 first.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/ ./core/database/conclusionsummaries/ ./core/matchregistry/` passes, `TestEventOrderAtKnownEnd`, `TestSortByBirth` and `TestLayerPackages` included.

**STEP-2 — One home for owning a DateValue**
Files: `core/database/datevalues/datevalues.go`, `core/database/datevalues/datevalues_test.go`, `core/database/deleteimpact/release.go`, `core/database/deleteimpact/release_test.go`, `core/database/observations/observations.go`, `core/database/observations/update.go`. Skills: `add-date-value`, `add-catalog-delete`.
Do: D-7. Behavior of Observation writes does not change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/datevalues/ ./core/database/deleteimpact/ ./core/database/observations/ ./core/database/audit/` passes, T-9 included; `rg 'dateCreateChange|dateUpdateChange|dateValueMap|releaseReplacedValue|intPtrAny' core` has no hit.

**STEP-3 — Migration and model**
Files: `core/database/migrations/000063.sql` *(new)*, `core/database/catalogmodel/model.go`. Skills: `add-catalog-migration`, `add-catalog-model`.
Do: D-1, D-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/catalogmodel/ ./core/database/deleteimpact/ ./core/database/` passes (`TestPragmaHonesty`, schema hash).

**STEP-4 — Source reads carry the date and flag**
Files: `core/database/sources/sources.go`, `core/database/sources/sources_test.go`. Skill: `add-catalog-query`.
Do: D-6, D-12, D-13 for `Get`, `getByRef`, `GetTx`, `List`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sources/` passes with T-4's default rows and T-10.

**STEP-5 — Source writes, delete and scopes**
Files: `core/database/sources/sources.go`, `core/database/sources/sources_test.go`, `core/database/effects/registry.go`, `core/database/audit/scopes_test.go`. Skills: `add-audit-scope`, `add-catalog-delete`.
Do: D-8, D-9, D-10, D-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes, T-3, T-4, T-5 and T-8 included.

**STEP-6 — Wire and handlers**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-3, D-5; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes, T-6, T-7 and `TestEveryAuditedWriteReturnsEffects` included.

**STEP-7 — Version and gate**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`).
Do: next minor from `main` (0.4.3 → 0.5.0 at the planning base).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-8 — Docs**
Files: `docs/source-layer-data-model.md` (§ 4: the two columns and a sentence each; § 4.2: the copy is undated and public until S15-32; § 10's "not a Source-layer attachment" line now names the Source date), `docs/structured-date-model.md` (§ 4 and § 5 rule 1: a Source owns one optional date; the one date order is `valuecompare.DateKeyOf`), `.agents/skills/add-date-value/SKILL.md` (a Source is an owner; `PutTx`, `ReleaseReplacedRow`; `DateKeyOf` / `Compare` is the one date order, beside `datewindow.go`), `.cursor/rules/date-values.mdc` (the same two lines), `.agents/skills/add-audit-scope/SKILL.md` (an owned value with a second owner passes its path to `ownedValue`), `docs/deployment-plan/spike-15/completed.md`, `docs/deployment-plan/spike-15/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-02-source-date-and-public-flag.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/dateorder_test.go` `TestDateKeyOf` *(new)* | Built with `core/datevalue/datevaluetest` `YMD`, `Before`, `After`, `About`, `Between`, `Phrase`; the open-start range is `datevalue.Value{Kind: datevalue.KindRange, EndYear: datevaluetest.Part(1880)}` and both-open is a range with no years. Rows: `1879`, `bef. 1880`, `1880`, `1881`, `bef. 29 Mar 2011`, `29 Mar 2011`, `aft. 29 Mar 2011` strictly ascending by `Compare`; `ABT 1880` = `1880`; `BET 1885 AND 1880` = `BET 1880 AND 1885`; range open at start = `BEF 1880`; phrase-only, both ends open, year outside `datevalue.MinYear` to `MaxYear` → not ok | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-4 |
| T-2 | Go | `conclusionsummaries_test.go` `TestEventOrderAtKnownEnd`, `TestSortByBirth` (unchanged) | Same order as before the move | D-4 |
| T-3 | Go | `sources_test.go` `TestSourceDate` *(new)* | Set `1880` → one row, `Date` reads back; set `ABT 12 Mar 1880` → same `DateValueID`; set same → no changes; clear → `DateValueID` nil, row gone; month 13 → `datevalue.ErrInvalid`, stored date unchanged; unknown Source → `ErrInvalid` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.6, D-8 |
| T-4 | Go | `sources_test.go` `TestSourcePublic` *(new)* | `Create` → `IsPublic` true and `List` true; set false → `Get` false; set false again → no changes; set true → true | AC-2.1, AC-2.2, AC-2.3, D-10, D-12 |
| T-5 | Go | `sources_test.go` `TestSourceDelete` new case "dated source releases its date" | Dated, empty Source → delete → `date_values` row gone; change carries `date_value_id` | AC-1.5, D-11 |
| T-6 | Go | `api/ffi/handlers/sources_test.go` `TestCopySource` new case "copy is undated and public" | Copy of a dated, not-public Source → copy has no `date`, `is_public` true | D-14 |
| T-7 | Go | `api/ffi/handlers/sources_test.go` `TestSetSourceDate`, `TestSetSourcePublic` *(new)*; `TestListSources` new case "date and sort key" | RPC set/clear/invalid as T-3 over the wire, effects present; `ListSources` over AC-3.1's eight Sources: sorting by `date_sort_key`, absent last, gives AC-3.1's order; phrase-only has `date` and no key | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-2.2, AC-3.1, AC-3.4, D-3, D-5 |
| T-8 | Go | `core/database/audit/scopes_test.go` `TestSourceScopes` new case "source date set, edit, clear and public flag" | Each of four writes `bumpsA` | AC-4.1, D-9 |
| T-9 | Go | `datevalues_test.go` `TestPutTx` *(new)*; `core/database/deleteimpact/release_test.go` `TestReleaseReplacedRow` *(new)* | `PutTx` with no prev → new id and a create change; same value → prev id, nil change, row untouched; changed value → prev id and an update change naming only changed fields; invalid → `datevalue.ErrInvalid`. `ReleaseReplacedRow` with an empty prev or prev = new → nil; with a prev → row gone and a delete change carrying `rowMap`'s fields | D-7 |
| T-10 | Go | `sources_test.go` `TestDateSortKey` *(new)* | T-1's ascending rows through `DateSortKey` give strictly ascending integers; equal `Compare` gives equal integers; nil date and phrase-only → false | AC-3.1, AC-3.2, AC-3.4, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `conclusionsummaries` order tests, `catalogmodel` `TestPragmaHonesty`, `deleteimpact` honesty tests, `effects` `TestEveryNonSkipTableHasAnEntry`, `TestEveryAuditedWriteReturnsEffects`, `observations` audit tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-15/deployment-plan.md: map row status; docs/deployment-plan/spike-15/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: source-layer-data-model.md § 4, § 4.2 and § 10; structured-date-model.md § 4 and § 5; add-date-value, date-values.mdc and add-audit-scope as STEP-8 lists
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No view changes, so the check is on the engine through the app's open path:

1. Open the product owner's dogfood project with the branch's app build. It opens (the migration runs, the schema hash passes) and the Sources list looks as before.
2. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestListSources|TestSetSourceDate|TestSetSourcePublic' -v ./api/ffi/handlers/` and see AC-3.1's order in the test's output.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving the order rule changes a Timeline or Person list order | T-2 stays unchanged and green |
| Moving the owned-date job changes an Observation write's audit rows | STEP-2 runs the `observations` and `audit` suites before any Source code |
| A second owned FK into `date_values` breaks a `deleteimpact` assumption of one owner per table | STEP-3's honesty tests run before any write code |
| `ADD COLUMN … CHECK` refused by the bundled SQLite | Stop condition below |

**Stop and report if:**

- `ALTER TABLE sources ADD COLUMN … CHECK (is_public IN (0, 1))` fails in the migration run.
- A `deleteimpact` or `catalogmodel` test fails because `date_values` has two owned-outbound parents, and the fix is outside `catalogmodel.FKs`.
- `TestEventOrderAtKnownEnd` or `TestSortByBirth` needs a changed expectation after STEP-1.
- `effects` `union(valueSource(...), inbound("sources", "date_value_id"))` fails `TestSourceEvidenceIsNotActivity` or another registry completeness test.
- A Swift file other than the generated `engine.pb.swift` fails to build.
- `Method` 117 or 118 is taken on `main`.
- An Observation audit test changes its expected changes after STEP-2.
- `datevalues` or `deleteimpact` would need to import the other to build D-7.

## 12. Open questions

None.

## 13. PR

**Title:** S15-02 — Sources carry a date and a not-public flag
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-02-source-date-and-public-flag.md`.
