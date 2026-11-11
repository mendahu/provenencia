# Ship log — S15-07 Metadata operations take an owner

**Plan:** [`S15-07-metadata-operations-take-an-owner.md`](S15-07-metadata-operations-take-an-owner.md) · **Branch:** `s15-07-metadata-operations-take-an-owner` · **PR:** #787
**State:** ready

## Round 0 — implement
Head `d9fcfdfd`. Gate PASS (orchestrator run). PR #787 (draft). Report deviations: STEP-1 and STEP-2 in one commit (no standalone STEP-1 commit); store holds `listSuffix`/`layoutListSuffix` fields (same SQL); `scanRow` scans the owner column into `Row.Owner`; `TestListWorkspaces` reads each Source before opening its read tx (one connection); manual check (§10) not run (no Swift or wire change).

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, `Localizable.xcstrings`, `core/apperr`, or new error codes in the diff).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1, MNT-1 | SEC, MNT | blocker | Four new by-id `IN` reads skip `sqlbatch`; >32,766 ids fails | To user → **Fix now → F-1** | Verified: `sqlbatch.go` doc "every by-id read goes through here"; SEC probe `n=40000 err=too many SQL variables`. Contradicted D-10 "No chunking"; user (2026-11-11) chose sqlbatch, including the `autoreconciler` `load` leftover |
| CON-1 | CON | major | Domain doc § 5.3 describes `date_value_id`, dropped by `000028.sql` | Fix now → F-2 | Verified: `000028.sql:1` "drop date_value_id"; `TestSourceMetadata/source_metadata_has_no_date_value_id` passes on head. Plan §4 "Schema" row premise was wrong |
| MNT-2 | MNT | major | 16-byte + unique field-id check copied three times | Fix now → F-3 | Verified: `sourcemetadata.go:335-357` (two in `apply`) and `:534-543` (`Reorder`); same-pass cleanup |
| MNT-3 | MNT | major | New test setup/count helpers copy existing ones | Fix now → F-4 | Same-pass cleanup: `seed`/`count`/`metaCount` vs `testWorld`/`applyWorld`/`countRows` |
| MNT-4 | MNT | minor | `insertValueTx` normalizes, `updateValueTx` does not; `apply` normalizes twice | Fix now → F-5 | Minor in code this PR adds |
| MNT-5 | MNT | minor | Six-column field scan still in `metadatafields.List` and `sourcevocab.listSuggestions` | Fix now → F-6 | Same-pass cleanup: this PR adds `scanField` and moved `listSuggestions`' body |
| CON-3, MNT-8 | CON, MNT | minor | `ListSuggestions` checks `HoldsMetadata` twice; `ListSuggestionsTx(nil, …)` panics | Fix now → F-6 | Minor in a function this PR adds; rides with F-6 |
| CON-2 | CON | minor | AC-1.2 test does not pin suggestion order before extras | Fix now → F-7 | F-4 touches the file |

Nits (logged): `MNT-6 · sourcemetadata.go:772 · typeIn suffix literal (goes away with F-1)` · `MNT-7 · sourcemetadata.go:325-330 · apply repeats nil-tx and RequireUserID checks its callers make` · `MNT-9 · sourcemetadata_test.go:708 · values and readDraft helpers both open a read tx over the same state`

### Fix plan 1

Supersedes D-10's "No chunking" (user, 2026-11-11, on SEC-1/MNT-1). Record that in the plan's Landed note.

**F-1 — Every new by-id `IN` read goes through `sqlbatch`** (from SEC-1, MNT-1, MNT-6)
Files: `core/database/sourcemetadata/sourcemetadata.go`, `core/database/metadatafields/metadatafields.go`, `core/database/autoreconciler/autoreconciler.go`, `.agents/skills/add-catalog-query/SKILL.md`, tests beside each.
Failing test first: `sourcemetadata` `TestListByOwnersTx/order holds across batches` — more than `sqlbatch.Size` owners (e.g. 1,201 Sources at the source level, ids in random insert order) return rows ordered by owner id then field id; and `metadatafields` `TestListByIDsTx/reads past one batch` — more than `sqlbatch.Size` ids (most unknown) return the known fields. (The bind-limit failure itself is pinned by `sqlbatch.TestQueryReadsPastTheBindLimit`; do not add a 32,767-id catalog test.)
Do: `listByOwners`, `listLayoutByOwners`, the `typeIn` read in `listWorkspaces` and `ListByIDsTx` call `sqlbatch.Query(q, head, tail, nil, ids, scan)` with the store's prefix and suffix (give `typeIn` its own suffix field, so no `")"` literal stays at a call site). Before batching, sort each level's unique ids with `bytes.Compare` (SQLite orders BLOBs by memcmp), so the per-batch `ORDER BY` results concatenate in global order. Put the per-level walk (for each level in `vocabkey.MetadataLevels()`: `stores[level]`, its ids, sorted) in one unexported helper so the three `sourcemetadata` loops share one body. Move `autoreconciler`'s `load` onto `sqlbatch.Query` too, keeping its `(entity id, property id)` group order (sort ids first if its ORDER BY depends on them across batches; check what its callers already pass — `recompute` batches through `sqlbatch.Each`, so it may need only the call change). Add one line to `add-catalog-query` SKILL.md: a by-id `IN` read goes through `core/database/sqlbatch`, never a hand-built `SQLInPlaceholders` list.
Check: `rg -n "SQLInPlaceholders" core --glob '*.go' --glob '!*_test.go'` prints only the definition in `core/database` and any call inside `sqlbatch` (list what remains in the report and why each is not a by-id read); the tests pass; scripts/check-pr.sh PASS.

**F-2 — Restore the domain doc: `source_metadata` has no `date_value_id`** (from CON-1)
Files: `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-15/prs/S15-07-metadata-operations-take-an-owner.md`.
Do: in § 5.3 keep `UNIQUE (source_id, field_id)`, remove the `date_value_id` DDL line, and restore "There is no `date_value_id` on this table" with "(`000028.sql` dropped it)". Correct the plan's §4 "Schema" row and STEP-4 text, and add a Landed-note line saying the premise was wrong.
Check: `rg -n date_value_id docs/source-layer-data-model.md` shows only the "There is no" sentence in § 5.3 (other sections' unrelated tables may still name it); scripts/check-pr.sh PASS.

**F-3 — One field-id check** (from MNT-2)
Files: `core/database/sourcemetadata/sourcemetadata.go`.
Do: one unexported `checkFieldIDs(ids [][]byte) (map[string]struct{}, error)` — each id 16 bytes and listed once, else `ErrInvalid`. `Reorder` and both loops in `apply` call it; `apply` keeps the values' returned set as `inValues`.
Check: `rg -n "len\(.*FieldID\) != 16|len\(fieldID\) != 16" core/database/sourcemetadata/*.go` prints only `checkFieldIDs`' line; existing `TestApply` and `Reorder` cases pass unchanged; scripts/check-pr.sh PASS.

**F-4 — One test setup and one row count in `sourcemetadata_test.go`** (from MNT-3)
Files: `core/database/sourcemetadata/sourcemetadata_test.go`.
Do: keep `testWorld`, `newSource`, `applyWorld`, `countRows`. Move `TestSourceMetadata` (`mustUser`, `mustSeededSource`, `metaCount`) and `TestCreateAll` (`seed`, `count`) onto them, and fold `TestOwnerLevels`' `counts` onto `countRows`. Delete the replaced helpers. Assertions unchanged.
Check: `rg -n "func (mustUser|mustSeededSource|metaCount)\b|seed :?= func|count :?= func|counts :?= func" core/database/sourcemetadata/sourcemetadata_test.go` prints nothing; tests pass; scripts/check-pr.sh PASS.

**F-5 — Callers normalize; both value builders take normalized text** (from MNT-4)
Files: `core/database/sourcemetadata/sourcemetadata.go`.
Do: `insertValueTx` takes normalized text like `updateValueTx`. `apply` passes `normalized[i]`; `Set` normalizes once before choosing insert or update.
Check: the normalize function is called once per value on each path (`rg -n` for it in the file and say where in the report); tests pass; scripts/check-pr.sh PASS.

**F-6 — One field scan; one level check in `ListSuggestions`; nil tx refused** (from MNT-5, CON-3, MNT-8)
Files: `core/database/metadatafields/metadatafields.go`, `core/database/sourcevocab/sourcevocab.go`, tests beside each.
Failing test first: `sourcevocab` `TestAppendSuggestion/ListSuggestionsTx with a nil tx is invalid` → `ErrInvalid`, no panic.
Do: `scanField(row, extra ...any)` scans the six columns then any trailing destinations; `List` uses it for `UsedBy`. Export it as `metadatafields.ScanField` (one doc line on the column order it expects) and have `sourcevocab.listSuggestions` use it with its trailing column(s). Drop `ListSuggestions`' outer `HoldsMetadata` check. `ListSuggestionsTx` returns `ErrInvalid` on a nil `tx`.
Check: `rg -n "Scan\(&f\.ID, &f\.Key|&f\.Origin, &f\.Label" core --glob '*.go'` prints only `ScanField`'s line; tests pass; scripts/check-pr.sh PASS.

**F-7 — AC-1.2 test pins the suggestion order** (from CON-2)
Files: `core/database/sourcemetadata/sourcemetadata_test.go`.
Do: in "list workspace reads several extras", assert the workspace keys equal the keys of `sourcevocab.ListSuggestions(c, typeID, vocabkey.LevelSource)` in order, followed by `a`, `b`, `c`.
Check: the case passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `34cc1505`. Gate PASS (orchestrator run). F-1 – F-7 done. F-1's new tests (`TestListByOwnersTx/order holds across batches`, `TestListByIDsTx/reads past one batch`) passed on the old code too, since ~1,200 ids is under the bind limit; kept as regression guards for the batch order. SEC-1 stays Fixed, not dropped: the bind-limit failure was verified by the SEC probe and is pinned by `sqlbatch.TestQueryReadsPastTheBindLimit`. `autoreconciler.load` also gained `UniqueBlobIDs`. Added test helper `birthOfAlice`. Plan Landed note records D-10 superseded and the wrong Schema premise.

## Round 2 — review
Reviewers: CON (CON-1 major), MNT (MNT-1 blocker, MNT-2/3 major), SEC+PRF (SEC-1). Scope `d9fcfdfd...34cc1505`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | CON, MNT, SEC+PRF | — | No new findings above nit | — | All three verdicts Clean |

Round-1 items: SEC-1/MNT-1 **Resolved** (SEC+PRF and MNT: every by-id read on `sqlbatch.Query` via `eachLevel`, ids sorted with `bytes.Compare`; order pinned by `TestListByOwnersTx/order holds across batches`). CON-1 **Resolved** (§ 5.3 restored; plan §4 and Landed note corrected). MNT-2 **Resolved** (`checkFieldIDs` sole list check). MNT-3 **Resolved** (one setup and `countRows`). Minors read by the orchestrator in the fix commits and confirmed by the round-2 reviewers: MNT-4 **Resolved** (`normalizeValue` once per path, `sourcemetadata.go:108`, `:374`), MNT-5/CON-3/MNT-8 **Resolved** (`metadatafields.ScanField` used by `List` and `sourcevocab.listSuggestions`; nil tx → `ErrInvalid`, tested), CON-2 **Resolved** (full key order asserted).

Remaining `SQLInPlaceholders` sites (traced by SEC+PRF and MNT round 2): `autoreconciler.go:317` deletes (callers batch via `sqlbatch.Each`), `searchindex/inputs.go`, `searchindex/handles.go`, `deleteimpact/handle_gone.go` (pre-existing, untouched), `gallery.go`, `identityclaims/bridgefile.go` (bounded key lists).

Nits (logged): `MNT-10 · sourcemetadata_test.go:542-548 · "source delete cascades layout" still counts with its own SELECT COUNT(*)` · `MNT-11 · sourcemetadata_test.go:920 · testWorld doc says it makes a Source; it makes none` · `MNT-12 · sourcemetadata_test.go:24,753 · tests set their own userID and drop testWorld's`

## Result
Gate PASS (`34cc1505`; `origin/main` unmoved since `6c8736d5`) · CI green · open: none · follow-ups: none · to user: SEC-1/MNT-1 (answered 2026-11-11: use `sqlbatch`, D-10 "No chunking" superseded)
