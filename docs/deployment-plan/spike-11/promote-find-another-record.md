# Promote: find another record

**Status:** pulled into [Spike 11](deployment-plan.md) on 2026-10-24 (S11-35, S11-36, S11-37, brief S11-D3). Idea first written in S11-05 (D-16), as `docs/ideas/promote-find-another-record.md`.

Today's page: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 (the page). Matching: [`comparison-layers.md`](comparison-layers.md) (relations, bands, contested).

## The problem

Promote's target menu lists only the records Promote offers: the proposed record and its alternatives, then New and Skip. Since matching moved onto value compare (S11-05), a record whose first given name or surname differs from the Subject's is `Different`, so it is never a candidate. A row with no candidate opens New (reason `no_match`) with nothing else in the menu.

A researcher who knows the Subject and a differently named record are one person (a nickname the aliases don't hold, a recorded variant, a married name) has no way to file the Subject on that record from Promote. The only route today is New, then a merge somewhere else, which the app does not have yet.

## Shape

- **An entry at the end of the target menu**, "Find another record…", on every row that has a menu (not an anchor row). It opens a search over the catalog's records of the row's kind.
- **The pick is a decision like any other row choice.** The row becomes decided on that record. The page sends it to Go as a fixed pair on each re-proposal, like a record picked from the menu.
- **Held, not moved.** Go never moves a fixed pair. It does score it: `graphalign`'s `anchor` makes the held handle the Subject's first candidate, and `loadCanon` loads every fixed anchor whether or not the shortlist would have found it. So the row shows the record's own band, the Contested marker when a property conflicts, and the evidence sheet, just as for a menu pick (the decided-row branch of `PromoteFlow.makeRow`). A record the researcher chose that the evidence rates below Weak reads "Little evidence" instead of "No match" (`SelectionReadout.chosen`); a suggested row at that band still reads "No match".
- **The record joins the row's menu.** Once picked, the record is an entry in that row's menu, so the menu shows it as the selection and the researcher can go back to an offered record, New or Skip. It stays in the menu while the page is open.
- **Search reuses `SearchCatalog`.** The request already takes `kinds`, and person, event and place hits already carry their list header (`SearchHit.header`, S9-35). No new search RPC.

## What the engine already does

| Need | Where it is today |
| --- | --- |
| Hold a pair Promote did not offer | `PromoteGraphAlignmentFixed` on `ProposePromoteGraphAlignmentRequest`; `graphalign.Align` places fixed pairs first |
| Score it | `state.anchor` scores each fixed handle; `rows.go` reports its band, comparisons and Contested, with reason `decided`. `ConflictWithFixed` cannot fire: a fixed Subject gets no shortlist, so there is no other candidate to beat it |
| Load a handle the shortlist missed | `promotealign.loadCanon` adds fixed anchors before candidates |
| Evidence for it | `GetPromoteClaimEvidence` takes any `handle_id`; `checkFixed` checks it |
| File it | Apply sends the row's handle to `promote.SaveBatch` (`api/ffi/handlers/promote_batch.go`) like any decided handle |
| Find records of one kind | `SearchCatalog` with `kinds`, hits with headers; merged handles are left out (`searchindex/handles.go`) |

`checkFixed` checks only that the handle is unmerged and of the Subject's kind, never that it was a candidate (map review, 2026-10-24).

S11-35 pinned this with tests and changed no Go code: `TestProposeHoldsAnUnshortlistedHandle` (`promotealign`) shows a fixed off-shortlist handle is loaded, scored and marked contested on a conflict, and over the FFI `TestProposeHoldsAFoundRecord` (propose), `TestGetPromoteClaimEvidenceForAFoundRecord` (evidence) and `TestApplyPromoteGraphAlignment` "files a found record" (filing) show it is explained and filed.

## Where the Swift side changes

- `PromoteFlow.target(for:on:)` resolves only handles in `row.menu`; any other token leaves the target as it was. A picked record needs a way in: a flow call, `hold(subjectID:id:ref:header:)`, that takes the record's id, ref and list header and titles it as the menu titles every record, adds it to `row.menu` and sets the target, with the same follow-up as `setTarget` (decided, duplicate notes, re-propose).
- A decided row keeps its `menu` across proposals (`makeRow`'s `kept = previous`), so the picked record survives a re-proposal. Its band is refreshed from the proposal. Its reason is not: Go reports `decided` for every fixed row and `makeRow` keeps a menu item's reason when the new one is `decided`, so a menu pick keeps the sentence it was offered with. A found record was never offered, so the call writes reason `decided`. Its band can be None for a `Different` pick; S11-D3 decides how that reads.
- `FakeStore`'s unscripted propose already holds a fixed handle of any filed record (it has no candidates); `PromoteModelTests` pins it.
- The menu entry (S11-36): `PVSelect` is a selection control with no action item, so the entry is a sentinel option, `PromoteFlow.Target.findToken`. `PromoteModel.setTarget` catches it and calls `openRecordSearch` before the flow sees it, and `PromoteFlow.setTarget` refuses any token that names no target, so opening or cancelling the search never marks the row decided. `PromoteTargetMenuRow` draws the entry; `Row.menuItem(forToken:)` is the one decoder of a record token.
- The search: S11-37 moved the omnibar's search state to `CatalogSearchModel` and its hit row to `CatalogSearchHitRow` (`macos/App/Features/Catalog/Search/`), with kinds at init and an optional location, so Promote reuses them instead of growing a copy.
- The picker (S11-36): `PromoteModel.recordSearch` holds the one open search (the row's Subject and a `CatalogSearchModel` over the row's kind). `PromoteRecordSearch` swaps it into the row's target cell as a `PVComboBox` and a Cancel button; a pick goes through `hold`, and Done or leaving closes the search.

## What stays later

- **Searching from the evidence sheet**, or comparing two records side by side before picking.
- **A warning beyond Contested** for a pick the comparison calls `Different` (refused, user 2026-10-24).
- **Records of another kind.** A person row searches persons only.
- **Merging records.** Picking a record files this Subject on it; it does not merge two records.

## Decided (user, 2026-10-24)

1. **Search surface: an inline picker on the row.** The omnibar is not connected to the Promote page, so the search opens from the row's target menu and stays on the page. It reuses the omnibar's search model and hit row (S11-37), not its surface. S11-D3 draws it.
2. **No warning for a differently named pick.** The row's band and Contested marker are enough. A `Different` pick shows no Callout.
3. **The search lists every canonical entity of the row's kind**, sorted by how well it matches the search string (`SearchCatalog`'s own ranking). Records already in the menu or taken by another row are not left out; the duplicate note (`sharesHandle`) does its usual job when two rows share one.
