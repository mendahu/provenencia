# Promote: find another record

**Status:** pulled into [Spike 11](deployment-plan.md) on 2026-10-24 (S11-35, S11-36, S11-37, brief S11-D3). Idea first written in S11-05 (D-16), as `docs/ideas/promote-find-another-record.md`.

Today's page: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §3 (the page). Matching: [`comparison-layers.md`](comparison-layers.md) (relations, bands, contested).

## The problem

Promote's target menu lists only the records Promote offers: the proposed record and its alternatives, then New and Skip. Since matching moved onto value compare (S11-05), a record whose first given name or surname differs from the Subject's is `Different`, so it is never a candidate. A row with no candidate opens New (reason `no_match`) with nothing else in the menu.

A researcher who knows the Subject and a differently named record are one person (a nickname the aliases don't hold, a recorded variant, a married name) has no way to file the Subject on that record from Promote. The only route today is New, then a merge somewhere else, which the app does not have yet.

## Shape

- **An entry at the end of the target menu**, "Find another record…", on every row that has a menu (not an anchor row). It opens a search over the catalog's records of the row's kind.
- **The pick is a decision like any other row choice.** The row becomes decided on that record. The page sends it to Go as a fixed pair on each re-proposal, like a record picked from the menu.
- **Held, not moved.** Go never moves a fixed pair. It does score it: `graphalign`'s `anchor` makes the held handle the Subject's first candidate, and `loadCanon` loads every fixed anchor whether or not the shortlist would have found it. So the row shows the record's own band, the Contested marker when a property conflicts, and the evidence sheet, just as for a menu pick (the decided-row branch of `PromoteFlow.makeRow`).
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

S11-35 pins this with tests rather than assuming it, extending the existing fixed-handle tests (`promote_align_test.go`, `promotealign_test.go` `TestProposeWithOnlyFixedSubjectsBuildsNoIndex`) and checking filing in the handler or `promote` tests: a fixed handle no shortlist key reaches is loaded, banded, marked contested when it conflicts, has evidence, and files.

## Where the Swift side changes

- `PromoteFlow.target(for:on:)` resolves only handles in `row.menu`; any other token leaves the target as it was. A picked record needs a way in: a flow call that takes the record (id, ref, sorted and reading titles, subtitle), adds it to `row.menu` and sets the target, with the same follow-up as `setTarget` (decided, duplicate notes, re-propose).
- A decided row keeps its `menu` across proposals (`makeRow`'s `kept = previous`), so the picked record survives a re-proposal. Its band is refreshed from the proposal. Its reason is not: Go reports `decided` for every fixed row and `makeRow` keeps a menu item's reason when the new one is `decided`, so a menu pick keeps the sentence it was offered with. A found record was never offered, so the call writes reason `decided`. Its band can be None for a `Different` pick; S11-D3 decides how that reads.
- `FakeStore`'s propose must hold a fixed handle that is not in its candidates, for the Swift tests and previews.
- The menu entry: `PVSelect` is a selection control with no action item. The entry is a sentinel option the row's binding intercepts before `setTarget`, which would otherwise mark the row decided and re-propose even though the target did not change; `PromoteTargetMenuRow` learns to draw it. If S11-D3 draws something a sentinel cannot do, a kit PR is cut ahead of S11-36.
- The search: the omnibar's results model (`OmnibarResultsModel`: debounce, generation guard, errors, empty state) and its private hit row (`OmnibarHitRowView`) are Workspace internals, and the model sends no `kinds`. S11-37 moves both to a shared home with a `kinds` input and an optional location, so Promote reuses them instead of growing a copy.

## What stays later

- **Searching from the evidence sheet**, or comparing two records side by side before picking.
- **A warning beyond Contested** for a pick the comparison calls `Different` (refused, user 2026-10-24).
- **Records of another kind.** A person row searches persons only.
- **Merging records.** Picking a record files this Subject on it; it does not merge two records.

## Decided (user, 2026-10-24)

1. **Search surface: an inline picker on the row.** The omnibar is not connected to the Promote page, so the search opens from the row's target menu and stays on the page. It reuses the omnibar's search model and hit row (S11-37), not its surface. S11-D3 draws it.
2. **No warning for a differently named pick.** The row's band and Contested marker are enough. A `Different` pick shows no Callout.
3. **The search lists every canonical entity of the row's kind**, sorted by how well it matches the search string (`SearchCatalog`'s own ranking). Records already in the menu or taken by another row are not left out; the duplicate note (`sharesHandle`) does its usual job when two rows share one.
