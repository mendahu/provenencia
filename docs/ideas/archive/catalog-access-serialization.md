# Catalog access serialization (archived)

**Done.** Go holds one exclusive catalog session (`core/catalogsession`). Skill: [`use-catalog-session`](../../../.cursor/skills/use-catalog-session/SKILL.md).

## Decisions

- Open-per-RPC was wrong: exclusive SQLite + parallel Swift `Task`s produced `catalog.already_open` and silent empties.
- The engine **serializes** catalog work on a held session. The Mac opens on first catalog RPC and closes via `GenealogyStore.closeCatalogSession` when leaving the workspace.
- Feature authors must not invent a second “is the catalog ready?” gate. Badge refresh surfaces failures instead of `try?` swallowing them.
- Related: one `GetWorkspaceNavCounts` RPC for sidebar badges — [`aggregate-workspace-nav-counts.md`](aggregate-workspace-nav-counts.md).
