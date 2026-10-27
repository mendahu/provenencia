# Settled-list count publish (archived)

**Status:** **Done** in `c6dcd51d`, `aedf7aa2` and `228bd821`. Found after S12-16.

A list publishes its sidebar count only from a settled handle: `QueryHandle.isSettled` (`status == .ready && !isFetching`) and the `.onSettled(handle)` modifier (`QueryHandleSettled.swift`), which runs the publish on appear and on each `revision` bump while the handle is settled. The Sources, Properties and Metadata lists publish through it, and their publishes right after a write are gone, so a create or delete no longer leaves the Properties badge one off.

## Leftovers

| Leftover | Home |
| --- | --- |
| A key marked stale while hidden is `.ready` and not fetching until its view reloads, so it can publish its cached count on appear (`QueryHandle.isStale`) | Spike 12 S12-26 |
