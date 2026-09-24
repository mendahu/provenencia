# Page navigation performance (archived)

**Done** — Spike 4. Skill: [`add-workspace-place`](../../../.cursor/skills/add-workspace-place/SKILL.md). Patterns: [`macos-client-patterns.md`](../../macos-client-patterns.md) § workspace session. Spike note: [`archive/spike-4`](../../deployment-plan/archive/spike-4/).

## Problem

History (`go(to:)`) was first-class. Loading was not: each destination `.task { load() }` on remount, Source page always refetched, and list vs detail could flash because presentation state lagged `currentLocation`.

## Decisions

- Treat catalog reads as a **session cache**, not per-view appear I/O.
- **Patch** when the write returns enough to update the handle; **invalidate** otherwise; **stale-while-revalidate** on navigation.
- A **place registry** declares which keys a location needs. `session.apply(location:)` warms them on every commit.
- Views **observe handles**. Do not call `session.query()` from `body`.
- One cache owns each list. Derived numbers get their own key.
- **Do not** add a Go read-model tier for this. Local SQLite plus the Mac cache is enough.
- Scroll offset and search query stay out of history (see [`navigation-history.md`](../../deployment-plan/archive/spike-3/navigation-history.md)).
