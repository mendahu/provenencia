# Spike 4 — Workspace session and place registry

**Done.** Replaced per-view load-on-appear with a project-scoped catalog query cache and a declarative place registry. Skill: [`add-workspace-place`](../../../../.cursor/skills/add-workspace-place/SKILL.md). Rationale: [`page-navigation-performance.md`](../../../ideas/archive/page-navigation-performance.md).

## Decisions

- Spike 3 owns **where** you are (`WorkspaceNavigation`). This spike owns **how** a place loads.
- `WorkspaceSession` + `CatalogQueryKey` + `QueryHandle`: patch when the mutation returns enough; invalidate otherwise; stale-while-revalidate on navigation reads.
- `PlaceRegistry` maps a location to presentation + query keys. Views subscribe to handles; they do not `.task { load() }` on appear.
- One cache owns each list. Do not hang derived numbers on the list payload (later: Sources-list counts in Spike 8).
- **Go RPC read-model tiering was descoped.** Local SQLite + the Mac session cache is enough.
- Scroll/search restoration and cross-window cache stayed out.
