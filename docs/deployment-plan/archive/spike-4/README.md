# Spike 4 — Workspace session and navigation performance

## Status

**Done.** Spike archived after dogfood (S4-01…S4-09). Finished steps: [`completed.md`](completed.md). **S4-10+** Go read-model tiering was **descoped** — Mac session cache hits are sufficient on local SQLite; notes remain in [`deployment-plan.md`](deployment-plan.md) for historical reference only.

Replace per-view load-on-appear with a **project-scoped catalog query cache** and **declarative place registry**, so navigation is fast, consistent, and extensible for future destinations.

Authoritative design: [`page-navigation-performance.md`](../../../ideas/archive/page-navigation-performance.md).

Agent skill: [`.cursor/skills/add-workspace-place/SKILL.md`](../../../../.cursor/skills/add-workspace-place/SKILL.md).

## Goal

A researcher can:

1. Navigate Sources list ↔ Source page, vocabulary, and sidebar destinations **without list flash**, **without cold reload** when returning to a section or source already loaded this session.
2. Use Back/Forward, omnibar, and breadcrumbs with **predictable** latency (cache hit vs first load is obvious and correct).
3. Add a new workspace place by registering query keys + place specs — follow `add-workspace-place`, not `.task` / `reconcileNavigation` boilerplate.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, cache strategy, descoped S4-10+ notes |
| [**Completed**](completed.md) | Finished steps (S4-01…S4-09) |
| [Page navigation performance](../../../ideas/archive/page-navigation-performance.md) | Design rationale + landed summary |

## Relationship to Spike 3

Spike 3 shipped **where** you can go (`WorkspaceNavigation`, persisted history, `go(to:)`). Spike 4 shipped **how** places load and render. No change to history stack semantics.

## Out of scope (for now)

- Go RPC read-model tiering (S4-10…S4-12 — descoped)
- Scroll/search restoration in history entries
- Cross-project or multi-window cache
- SwiftUI `NavigationStack` for session history
