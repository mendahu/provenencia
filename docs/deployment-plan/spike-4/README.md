# Spike 4 — Workspace session and navigation performance

**Status:** **done** (S4-01…S4-09). Completed steps: [`completed.md`](completed.md). Optional Go RPC tiering: [`deployment-plan.md`](deployment-plan.md) (S4-10+).

Replace per-view load-on-appear with a **project-scoped catalog query cache** and **declarative place registry**, so navigation is fast, consistent, and extensible for future destinations.

Authoritative design: [`docs/ideas/page-navigation-performance.md`](../../ideas/page-navigation-performance.md).

Agent skill: [`.cursor/skills/add-workspace-place/SKILL.md`](../../../.cursor/skills/add-workspace-place/SKILL.md).

## Goal (dogfood bar)

A researcher can:

1. Navigate Sources list ↔ Source page, vocabulary, and sidebar destinations **without list flash**, **without cold reload** when returning to a section or source already loaded this session.
2. Use Back/Forward, omnibar, and breadcrumbs with **predictable** latency (cache hit vs first load is obvious and correct).
3. Add a new workspace place by registering query keys + place specs — follow `add-workspace-place`, not `.task` / `reconcileNavigation` boilerplate.

## Relationship to Spike 3

Spike 3 shipped **where** you can go (`WorkspaceNavigation`, persisted history, `go(to:)`). Spike 4 shipped **how** places load and render. No change to history stack semantics.

## Explicit non-goals

- Scroll/search restoration in history entries
- Cross-project or multi-window cache
- SwiftUI `NavigationStack` for session history
- Generic `entityKind + id` on persisted `WorkspaceLocation` JSON (registry abstracts in Swift; JSON keeps per-kind optional fields)
