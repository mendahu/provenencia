# Value row owner check

**Status:** parked 2026-11-09 from S15-02's review (SEC-2, ship log `docs/deployment-plan/spike-15/prs/S15-02-source-date-and-public-flag.ship.md`). Only a crafted catalog reaches it.

## Problem

A `date_values` row has one owner: an Observation (`observations.value_date_id`) or, since S15-02, a Source (`sources.date_value_id`). Each column has its own unique index, so two Sources (or two Observations) cannot share a row. Nothing stops a crafted catalog from pointing a Source and an Observation at the **same** row.

- Setting the Source's date calls `datevalues.PutTx(tx, prevID, v)`, which edits the row in place, so the Observation's date changes too. The audit carries both scopes (the `date_values` effect follows every inbound owner), so the edit is recorded, but it is still made.
- Clearing the date or deleting either owner tries to delete a row still referenced; the `NO ACTION` FK refuses it, so that path is safe.

The app never creates the shared row: every owner write mints its row through `PutTx`.

## Rough shape

One owner check in `datevalues.PutTx` (and `namevalues.PutTx`, which has the same shape) before an in-place update: count references to `prevID` across every owned outbound FK into the value table, as `catalogmodel` registers them (the same list `deleteimpact.ReleaseReplaced` checks). More than one → insert a fresh row for this owner instead of editing in place, and leave the old row to its other owner. Not a Sources-only guard.

Touch it when `PutTx` next changes, or with the next new kind of owner of a date row.

## Check

A test that points a Source and an Observation at one date row with foreign keys off, sets the Source's date, and sees the Observation's date unchanged.
