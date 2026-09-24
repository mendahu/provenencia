# Spike 3 — Navigation history and omnibar

**Done.** First-class Back/Forward and project search. Behavior contracts (still live):

- [`navigation-history.md`](navigation-history.md)
- [`omnibar-search.md`](omnibar-search.md)

Client patterns: [`macos-client-patterns.md`](../../../macos-client-patterns.md).

## Decisions

- **One coordinator.** Every committed navigation goes through `go(to:)` / Back / Forward. No parallel `selectedSection` / `closeSource()` paths.
- History is **persisted** per `project.uuid` (Application Support JSON) and restored on relaunch. Not SwiftUI `NavigationStack`.
- A history entry is a **restorable place** (section + deep ids). Search query, scroll, focus, and dirty drafts are omitted.
- **Omnibar** is engine-side catalog search (registry + FTS5 + ranking), not per-destination `LIKE`. Hits navigate via `go(to:)`.
- Per-destination list search was removed in favor of the omnibar.

Catalog session serialization is a separate decision: [`catalog-access-serialization.md`](../../../ideas/archive/catalog-access-serialization.md).
