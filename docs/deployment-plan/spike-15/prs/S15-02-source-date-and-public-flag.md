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
| `datevalues` gains the shared join scan and the date audit changes; `observations` uses the changes | STEP-2, D-6, D-7 |
| `catalogmodel` FK; `date_values` effect reaches the Source | STEP-3, STEP-5, AC-4.1 |
| `sources` reads carry date and flag (one join for the list); `SetDate`, `SetPublic`, `DateSortKey`; Delete releases the date | STEP-4, STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3 |
| FFI: `Source` fields, `SetSourceDate`, `SetSourcePublic`; generated Go and Swift | STEP-6, STEP-7, all ACs |
| `source-layer-data-model.md` § 4 and § 4.2 | STEP-8 |

| Out | Home |
| --- | --- |
| The Source page date entry and the switch | S15-03 |
| The Sources list date column and the date sort | S15-04 |
| The Person page's Sources section date | S15-05 |
| Copy Source carrying the date, with a tick on the Copy form | S15-32 (user, 2026-11-09) |
| Copy Source carrying the not-public flag | not planned (working rule "Not public is a stored switch") |
| Any filter, glyph, hide or cascade on not public | not planned (working rule) |
| `CatalogSource`, `GenealogyStore`, `GoStore`, `FakeStore` fields and methods for the date and flag | S15-03 (flag and date write), S15-04 (sort key) |
| `observations.scanListed` and the `deleteimpact` projector moving onto `datevalues.Joined` | not planned in Spike 15 (offered as a separate chore) |
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
| Date queries | `core/database/datevalues/datevalues.go` `InsertTx`, `UpdateTx`, `LookupTx`, `LookupManyTx`, private `scanValue` | No exported scan for a `LEFT JOIN date_values dv` |
| Date audit | `core/database/observations/observations.go` `dateCreateChange`, `dateUpdateChange`, `dateValueMap`; `update.go` `releaseReplacedValue` | Private to `observations`; the Source needs the same three |
| Hand-scanned join | `observations.go` `sqlListSelect`, `scanListed`; `core/database/deleteimpact/projectors.go` (`LEFT JOIN date_values dv`) | Two copies of the 20-column scan |
| Effects | `core/database/effects/registry.go` `"date_values": ownedValue("date_value", "value_date_id")`; `"sources"` entry | A `date_value` change reaches a Source only through `observations.value_date_id`, so an in-place Source date edit would scope to nothing |
| Owned release | `core/database/deleteimpact/release.go` `SnapshotOwned`, `ReleaseSnapshot`, `ReleaseReplaced` | Derived from `catalogmodel` owned FKs; `artifacts.Delete` is the snapshot precedent |
| Wire | `api/proto/engine.proto` `Source` (fields 1–13), `DateValueInput`, `Method` (last 116), `SetSourceCoverRequest/Response` | No date or flag |
| Handlers | `api/ffi/handlers/sources.go` `sourceProto`, `ListSources`, `GetSourceWorkspace`, `SetSourceCover`; `api/ffi/handlers/thumbnails.go` `enrichSourceProto`; `api/ffi/dispatch.go` `MethodSetSourceCover` | Every Source on the wire goes through `sourceProto` |
| Codec | `core/valuecodec/valuecodec.go` `DateToProto`, `DateFromProto` | Reused as is |
| Copy | `core/database/sourcecopy/sourcecopy.go` `ReadDraft`, `CreateFromDraft` | Builds `sources.CreateInput`; unaffected |
| Tests | `core/database/sources/sources_test.go` `TestSources`, `TestSourceDelete`; `api/ffi/handlers/sources_test.go` `TestListSources`, `TestGetSourceWorkspace`, `TestDeleteSource`, `TestCopySource`; `core/database/audit/scopes_test.go` `TestSourceScopes`; `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestEventOrderAtKnownEnd`, `TestSortByBirth`; `core/valuecompare/datewindow_test.go` `TestDateEnds`; `api/ffi/handlers/write_effects_test.go` `TestEveryAuditedWriteReturnsEffects` | — |

**Precedent.** `000015.sql` plus `sources.SetCover` and the `SetSourceCover` RPC added a per-Source attribute end to end: copy the migration shape, the one-attribute write and its response (`Source` plus `WriteEffects`). `observations.resolveValue` (date case) and `update.go` `releaseReplacedValue` own a DateValue: copy insert-or-update-in-place and the release on clear. `artifacts.Delete` snapshots owned children before its `DELETE`: copy that in `sources.Delete`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One migration `000063.sql`: `ALTER TABLE sources ADD COLUMN date_value_id BLOB REFERENCES date_values(id);`, `ALTER TABLE sources ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1));`, `CREATE INDEX sources_date_value_id_idx ON sources(date_value_id);` | Working rules; defaults give every existing Source undated and public; the honesty tests require a covering index on an FK child column | A table rebuild — not needed for additive columns |
| D-2 | `sources.date_value_id → date_values` is registered `NO ACTION`, `BucketOwnedOutbound` | The Source owns its date as an Observation owns its value; `deleteimpact` then derives the release | `SET NULL` / a facet — leaves orphan `date_values` rows |
| D-3 | Two new RPCs, `SetSourceDate` (117) and `SetSourcePublic` (118), each returning `Source` and `WriteEffects`. A `SetSourceDateRequest` with no `date`, or a `date` whose `kind` is blank after trim, clears | Mirrors `SetSourceCover`; the Source page writes each on its own control; `UpdateSourceRequest` is a whole-row write whose proto3 `bool` default would mark every Source not public when an old caller omits it | New fields on `UpdateSourceRequest` — the default-false and omitted-date problem |
| D-4 | The order rule moves to `core/valuecompare/dateorder.go` *(new)*: `type DateKey struct{ Day, Side int }`, `func DateKeyOf(d datevalue.Value) (key DateKey, first, last DateEnd, ok bool)`, `func (k DateKey) Compare(o DateKey) int` (Day, then Side), `func (k DateKey) Int64() int64` = `Day*3 + Side + 1`. `DateKeyOf` holds `datePlace`'s body without the span. `conclusionsummaries.datePlace` calls it and builds the span from the side; `placedDate` and `dateKey` hold a `valuecompare.DateKey`; `dateKeyLess` keeps its `ok`-first and ref tie-break and compares days through `Compare` | Working rule "One date order"; `TimelineSpan` and the ref tie-break stay Conclusion concerns; `Side` is in −1, 0 or 1, so `Day*3 + Side + 1` keeps the order in one integer | Two wire fields (day, side) — Swift would re-implement the tie rule |
| D-5 | Wire: `Source` gains `DateValueInput date = 14`, `optional int64 date_sort_key = 15`, `bool is_public = 16`. Every Source sent goes through `sourceProto`, which sets all three. `date_sort_key` is absent when the Source is undated or `DateKeyOf` is not ok | One reader path; the key is a wire value so S15-04 sorts without date logic; `is_public` matches the column name | `bool not_public` — a second name for one fact |
| D-6 | `datevalues` gains `JoinColumns` *(new const, the 20 `dv.` columns in `sqlLookup` order)* and `Joined` *(new: `Dest() []any`, `Value(id []byte) *datevalue.Value`, nil when `kind` is NULL)*. `sources` uses them in `sqlGet`, `sqlGetByRef` and `sqlList` with `LEFT JOIN date_values dv ON dv.id = s.date_value_id` | Working rule: one join for the list (3,000 Sources); `datevalues` owns the column list | `LookupManyTx` after the list — a second round of queries; a third hand copy of the scan |
| D-7 | `observations`' `dateCreateChange`, `dateUpdateChange` and `dateValueMap` move to `datevalues` as `CreateChange`, `UpdateChange`, `RowMap`; `observations` calls them; the private copies are deleted | One audit shape for a `date_value` row, whoever owns it | A copy in `sources` |
| D-8 | `sources.SetDate(tx, userID, sourceID []byte, date *datevalue.Value) (Source, []rowchange.Change, error)` *(new)*. nil→value: `datevalues.InsertTx`, point the Source at it; changes: `date_value` create, `source` update `date_value_id` nil→id. value→value: `datevalues.UpdateTx` in place; change: `datevalues.UpdateChange` only (none when equal). value→nil: point the Source at NULL, `deleteimpact.ReleaseReplaced(tx, "sources.date_value_id", old)`; changes: `source` update id→nil and `date_value` delete with `rowchange.DeletedRow(datevalues.RowMap(old, prev))`. nil→nil: no changes. A missing Source or a short id is `ErrInvalid`; an invalid date is `datevalue.ErrInvalid` | The Observation precedent; a stable id across edits | Always insert a new row — churns ids and doubles audit rows |
| D-9 | The `date_values` effect gains the Source: `SourceScope: union(valueSource("value_date_id"), inbound("sources", "date_value_id"))`; `EvidenceGraph`, `Reads` and `Handles` stay as `ownedValue` gives them | An in-place edit writes only a `date_value` change; without this it moves no Source's activity (AC-4.1) | A hand-written `source` change on every edit — `add-audit-scope` forbids hand tagging |
| D-10 | `sources.SetPublic(tx, userID, sourceID []byte, isPublic bool) (Source, []rowchange.Change, error)` *(new)*: one `source` update with `is_public` old→new as `0`/`1`; unchanged returns no changes | Mirrors `SetCover`; the column is an integer | — |
| D-11 | `sources.Delete` takes `deleteimpact.SnapshotOwned(tx.Tx, catalogmodel.KindSource, id)` after `ReleaseFacets` and before its `DELETE`, then `deleteimpact.ReleaseSnapshot`. The delete change adds `date_value_id` when set and always `is_public` | `artifacts.Delete` precedent; AC-1.5 | — |
| D-12 | `Source` gains `DateValueID []byte`, `Date *datevalue.Value`, `IsPublic bool`. `Create` returns `IsPublic: true`. `Update` neither reads nor writes them | One struct for every read | — |
| D-13 | `sources.DateSortKey(s Source) (int64, bool)` *(new)* calls `valuecompare.DateKeyOf` and `Int64`; `sourceProto` calls it | The rule's caller is the `sources` package, as the working rule names | Calling `valuecompare` from the handler |
| D-14 | Copy Source is unchanged: a copy is undated and public | S15-32 adds the date tick (user, 2026-11-09); the flag is never carried (working rule) | — |
| D-15 | No migration test; the default read is tested through `sources.Create` and `List` (AC-2.1) | `add-catalog-migration` step 4 | An upgrade test |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/migrations/000063.sql` *(new)* | migration | `000015.sql` | A new step is the only way to change the schema |
| `core/valuecompare/dateorder.go` *(new)*: `DateKey`, `DateKeyOf`, `Compare`, `Int64` | file / type / func | `conclusionsummaries` `datePlace`, `dateKeyLess` | Moved, not copied: the working rule names this home |
| `datevalues.JoinColumns`, `datevalues.Joined` *(new)* | const / type | `datevalues.scanValue` (private), `observations.scanListed` | `scanValue` scans a non-null row; a `LEFT JOIN` needs nullable `kind`. `Joined` wraps the same column order |
| `datevalues.CreateChange`, `UpdateChange`, `RowMap` *(new, moved)* | func | `observations` `dateCreateChange`, `dateUpdateChange`, `dateValueMap` | Moved so both owners share them |
| `sources.SetDate`, `sources.SetPublic`, `sources.DateSortKey` *(new)* | func | `sources.SetCover` | Same shape; one attribute each |
| `SetSourceDate`, `SetSourcePublic` *(new)* | FFI | `SetSourceCover` | D-3 |
| `Source.date`, `date_sort_key`, `is_public` *(new)* | proto fields | `Source.cover_mode` | Additive |
| `TestDateKeyOf` *(new)* | test | `TestDateEnds` | Same file family and table style |
| `TestSourceDate`, `TestSourcePublic` *(new)* | test | `TestSources` | New tables beside it in `sources_test.go` |
| `TestSetSourceDate`, `TestSetSourcePublic` *(new)* | test | `TestListSources` | `runRPC` tables in `api/ffi/handlers/sources_test.go` |

Searched for: another date sort key (`sortKey`, `DateWindow` "Sort keys" comment in `datewindow.go`, `dateKey` across `core/`), another `LEFT JOIN date_values` (two hits, above), another one-attribute Source write (`SetCover`), another owner of a `date_values` row (Observations only), an existing privacy flag (none, per the privacy note).

**Builds on.** `catalogmodel` owned-outbound registration and `deleteimpact`'s derived release; the effects registry's `union` / `inbound` paths; `valuecodec` for the wire date; `writes.Run` and `writeEffectsProto`.
**Deletes.** `conclusionsummaries` loses the key-and-side body of `datePlace` and compares through `valuecompare.DateKey.Compare`. `observations` loses `dateCreateChange`, `dateUpdateChange`, `dateValueMap`.
**Next adds.** S15-04 sorts by `date_sort_key`; S15-03 calls the two RPCs; S15-05 reads `Source.Date` into the Source record; S15-16 reads the date for `{source.date}`; S15-32 adds a `copy_date` choice to `sourcecopy` that calls `SetDate` on the copy.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A `sources` row that owns a `date_values` row | `sources.Delete`, `deleteimpact` honesty tests | D-11 releases it; T-5 |
| A Source deleted while dated | `effects` resolve on the `source` delete change | The change carries `date_value_id` (D-11); the date row's own release is not audited, as for an Observation |
| A `date_value` change whose owner is a Source | `effects` `date_values` entry | D-9; T-8 |
| Copy of a dated, not-public Source | `sourcecopy.CreateFromDraft` → `sources.Create` | Copy is undated and public (D-14); T-6 |

### Scale

At the map's Scale (3,000 Sources): `sources.List` stays one query, with one `LEFT JOIN date_values` on its primary key per row, and one `DateKeyOf` (constant work) per row in `sourceProto`. No new per-row query. Each write touches one Source: one Source scope, no Evidence graph, no list count; `SetSourcePublic` reprojects that one Source's search document through the existing `sources` effect. No budget is set or leaned on.

### Edges

- **Assumptions.** `dateKeyLess` assumes a total order on (ok, Day, Side, ref); `Compare` keeps Day then Side, and `TestEventOrderAtKnownEnd` and `TestSortByBirth` hold it unchanged. `Int64` assumes `Side ∈ {−1, 0, 1}`; T-1 checks `Int64` order equals `Compare` on every row of its table.
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
CREATE INDEX sources_date_value_id_idx ON sources(date_value_id);
```

Registered in `catalogmodel.FKs`: `{From: "sources", Column: "date_value_id", To: "date_values", OnDelete: "NO ACTION", Bucket: BucketOwnedOutbound}`. Delete class: owned outbound (D-2). Audit: `source` and `date_value` changes (D-8 – D-10).

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `dateorder.go` *(new)* per D-4 |
| `core/database/conclusionsummaries` | `compose.go` `datePlace`, `placedDate`, `dateKey`, `dateKeyOf`, `dateKeyLess`, `eventDating` use `valuecompare.DateKey` (D-4); `order.go` unchanged in behavior |
| `core/database/datevalues` | `JoinColumns`, `Joined` (D-6); `CreateChange`, `UpdateChange`, `RowMap` (D-7) |
| `core/database/observations` | Calls the moved helpers; private copies deleted |
| `core/database/catalogmodel` | One FK row (D-2) |
| `core/database/effects` | `date_values` entry (D-9) |
| `core/database/sources` | `Source` fields (D-12); reads join (D-6); `SetDate` (D-8); `SetPublic` (D-10); `DateSortKey` (D-13); `Delete` (D-11) |

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

**STEP-2 — Shared date join and date audit changes**
Files: `core/database/datevalues/datevalues.go`, `core/database/datevalues/datevalues_test.go`, `core/database/observations/observations.go`, `core/database/observations/update.go`. Skill: `add-date-value`.
Do: D-6, D-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/datevalues/ ./core/database/observations/` passes; `rg 'dateCreateChange|dateUpdateChange|dateValueMap' core` has no hit.

**STEP-3 — Migration and model**
Files: `core/database/migrations/000063.sql` *(new)*, `core/database/catalogmodel/model.go`. Skills: `add-catalog-migration`, `add-catalog-model`.
Do: D-1, D-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/catalogmodel/ ./core/database/deleteimpact/ ./core/database/` passes (`TestPragmaHonesty`, schema hash).

**STEP-4 — Source reads carry the date and flag**
Files: `core/database/sources/sources.go`, `core/database/sources/sources_test.go`. Skill: `add-catalog-query`.
Do: D-6, D-12, D-13 for `Get`, `getByRef`, `GetTx`, `List`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sources/` passes with T-4's default rows.

**STEP-5 — Source writes, delete and scopes**
Files: `core/database/sources/sources.go`, `core/database/sources/sources_test.go`, `core/database/effects/registry.go`, `core/database/audit/scopes_test.go`. Skills: `add-audit-scope`, `add-catalog-delete`.
Do: D-8, D-9, D-10, D-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes, T-3 – T-5 and T-8 included.

**STEP-6 — Wire and handlers**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/dispatch.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-3, D-5; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes, T-6, T-7 and `TestEveryAuditedWriteReturnsEffects` included.

**STEP-7 — Version and gate**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`).
Do: next minor from `main` (0.4.3 → 0.5.0 at the planning base).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-8 — Docs**
Files: `docs/source-layer-data-model.md` (§ 4: the two columns and a sentence each; § 4.2: the copy is undated and public until S15-32), `completed.md`, `deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-15/prs/S15-02-source-date-and-public-flag.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/dateorder_test.go` `TestDateKeyOf` *(new)* | Rows: `1879`, `bef. 1880`, `1880`, `1881`, `bef. 29 Mar 2011`, `29 Mar 2011`, `aft. 29 Mar 2011` strictly ascending by `Compare` and by `Int64`; `ABT 1880` = `1880`; `BET 1885 AND 1880` = `BET 1880 AND 1885`; range open at start = `BEF` its end; phrase-only, both ends open, year outside `MinYear to MaxYear` → not ok | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-4 |
| T-2 | Go | `conclusionsummaries_test.go` `TestEventOrderAtKnownEnd`, `TestSortByBirth` (unchanged) | Same order as before the move | D-4 |
| T-3 | Go | `sources_test.go` `TestSourceDate` *(new)* | Set `1880` → one row, `Date` reads back; set `ABT 12 Mar 1880` → same `DateValueID`; set same → no changes; clear → `DateValueID` nil, row gone; month 13 → `datevalue.ErrInvalid`, stored date unchanged; unknown Source → `ErrInvalid` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.6, D-8 |
| T-4 | Go | `sources_test.go` `TestSourcePublic` *(new)* | `Create` → `IsPublic` true and `List` true; set false → `Get` false; set false again → no changes; set true → true | AC-2.1, AC-2.2, AC-2.3, D-10, D-12 |
| T-5 | Go | `sources_test.go` `TestSourceDelete` new case "dated source releases its date" | Dated, empty Source → delete → `date_values` row gone; change carries `date_value_id` | AC-1.5, D-11 |
| T-6 | Go | `api/ffi/handlers/sources_test.go` `TestCopySource` new case "copy is undated and public" | Copy of a dated, not-public Source → copy has no `date`, `is_public` true | D-14 |
| T-7 | Go | `api/ffi/handlers/sources_test.go` `TestSetSourceDate`, `TestSetSourcePublic` *(new)*; `TestListSources` new case "date and sort key" | RPC set/clear/invalid as T-3 over the wire, effects present; `ListSources` over AC-3.1's eight Sources: sorting by `date_sort_key`, absent last, gives AC-3.1's order; phrase-only has `date` and no key | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-2.2, AC-3.1, AC-3.4, D-3, D-5 |
| T-8 | Go | `core/database/audit/scopes_test.go` `TestSourceScopes` new case "source date set, edit, clear and public flag" | Each of four writes `bumpsA` | AC-4.1, D-9 |
| T-9 | Go | `datevalues_test.go` `TestJoined` *(new)* | `LEFT JOIN` with no row → `Value` nil; with a range row → the value equals `LookupTx`'s | D-6 |

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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: source-layer-data-model.md § 4 and § 4.2
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
| A second owned FK into `date_values` breaks a `deleteimpact` assumption of one owner per table | STEP-3's honesty tests run before any write code |
| `ADD COLUMN … CHECK` refused by the bundled SQLite | Stop condition below |

**Stop and report if:**

- `ALTER TABLE sources ADD COLUMN … CHECK (is_public IN (0, 1))` fails in the migration run.
- A `deleteimpact` or `catalogmodel` test fails because `date_values` has two owned-outbound parents, and the fix is outside `catalogmodel.FKs`.
- `TestEventOrderAtKnownEnd` or `TestSortByBirth` needs a changed expectation after STEP-1.
- `effects` `union(valueSource(...), inbound("sources", "date_value_id"))` fails `TestSourceEvidenceIsNotActivity` or another registry completeness test.
- A Swift file other than the generated `engine.pb.swift` fails to build.
- `Method` 117 or 118 is taken on `main`.

## 12. Open questions

None.

## 13. PR

**Title:** S15-02 — Sources carry a date and a not-public flag
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-02-source-date-and-public-flag.md`.
