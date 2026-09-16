---
name: add-workspace-location
description: >-
  Wires a new Provenencia macOS workspace place into first-class navigation
  history (WorkspaceLocation, WorkspaceNavigation.go(to:), Back/Forward,
  persist/restore). Use when adding or changing a sidebar destination,
  Source-style deep page, master–detail selection, workspace section,
  breadcrumb/omnibar target, or any view that should restore via history —
  WorkspaceNavigation, WorkspaceSection, NavigationHistoryStore,
  selectedSection bypass, closeSource without go(to:), or navigation-history.md.
---

# Add a workspace location (navigation history)

Every **committed place** the researcher can return to must be a
[`WorkspaceLocation`](../../../macos/App/Features/Workspace/WorkspaceLocation.swift)
pushed through
[`WorkspaceNavigation.go(to:)`](../../../macos/App/Features/Workspace/WorkspaceNavigation.swift).
`WorkspaceNavigation` is owned by `WorkspaceView` and injected with
`.environment(navigation)`. Do **not** use SwiftUI `NavigationStack` /
`NavigationPath`. Do **not** leave parallel “where we are” state that bypasses
the stack. Sidebar collapse stays on `WorkspaceModel` (chrome only).

Authoritative behavior: [`docs/deployment-plan/archive/spike-3/navigation-history.md`](../../../docs/deployment-plan/archive/spike-3/navigation-history.md).

## When this applies

| Change | History work required? |
| --- | --- |
| New sidebar section / destination host | **Yes** — section + list root |
| List → detail page (Source-style) | **Yes** — deep id on the location |
| Master–detail row select (fields/types pattern) | **Yes** — selection is a place |
| Omnibar / breadcrumb / cross-link → a place | **Yes** — call `go(to:)` |
| Search query, scroll, focus, dirty drafts | **No** — UI ephemera, omit |
| Modal/sheet that is not a restoreable place | **No** — unless product says Back should return into it |

## Checklist

```
- [ ] WorkspaceLocation carries any new deep id (and == ignores ref/title fluff)
- [ ] New sidebar section added to WorkspaceSection if needed
- [ ] Every commit path calls navigation.go(to:) (sidebar, open, close/list-back, row select, links)
- [ ] No raw selectedSection = / openSource / closeSource / select as the sole place change
- [ ] Destination applies currentLocation on appear + onChange (after query handle ready)
- [ ] Missing deep id → navigation.fallbackToSectionRoot() after list/workspace ready
- [ ] Destination uses @Environment(WorkspaceNavigation.self); not a private parallel nav store
- [ ] WorkspaceNavigationTests cover push + restore (+ prune if new deep id)
```

## Steps

### 1. Extend `WorkspaceLocation` (if new deep kind)

Add an optional stable id field (e.g. `artifactId`, `nodeId`). Keep denormalized
`ref` / `title` for jump-menu chrome only.

Update `==` to include the new id (still **ignore** `ref` / `title`).

JSON under Application Support is v1 `Codable` — new optional keys are fine;
do not bump casually. See `NavigationHistoryStore` / `NavigationHistoryDocument`.

### 2. Sidebar section (if new destination)

Add `WorkspaceSection` case (raw value kebab-case), L10n label/icon
([`add-localized-string`](../add-localized-string/SKILL.md)), sidebar item, and
`WorkspaceContent` host.

Sidebar select **always** pushes a **section list root**:

```swift
navigation.go(to: .sectionRoot(.yourSection))
```

Never assign `navigation.selectedSection` directly.

### 3. Record navigation → `go(to:)`

| User action | Call |
| --- | --- |
| Open deep page / activate list row | `go(to: WorkspaceLocation(section:…, yourId:…, ref:…, title:…))` |
| Back to list / clear selection | `go(to: .sectionRoot(.yourSection))` |
| Master–detail select | Binding `set` → `go(to:)` with deep id (model `select` runs from **apply**) |

Feature models may still implement `open` / `select` / `close` for **applying**
UI state. Those must not be the only path that changes place when the user
commits navigation.

### 4. Apply location → UI

In the destination view (pattern: `SourcesListView`, `SourceFieldsView`):

1. Take `@Environment(WorkspaceNavigation.self) private var navigation`.
2. On `onAppear` + `onChange(of: navigation.currentLocation)` (+ `onChange` on query
   handle status/value when list loads async):
   - Guard `location.section == .yourSection`.
   - Deep id present + found → apply selection/open (model `syncSelection(from:)`).
   - Deep id present + missing after handle ready → `navigation.fallbackToSectionRoot()`.
   - Deep id nil → clear to list root (no second `go(to:)`).

Catalog reads come from `WorkspaceSession` query handles (Spike 4) — see
[`add-workspace-place`](../add-workspace-place/SKILL.md). Do **not** add `.task {
load(from:) }` as the primary hydration path.

### 5. Tests

Extend [`WorkspaceNavigationTests`](../../../macos/ProvenenciaTests/WorkspaceNavigationTests.swift)
([`add-swift-test`](../add-swift-test/SKILL.md)):

- Push the new location; Back/Forward restore the deep id.
- Missing-entity prune → section root (if applicable).

Inject a temp navigation file URL via `attachProject(uuid:fileURL:)`.

## Do not

- Bypass history with hierarchical “up” controls that only clear local state
- Push on every search keystroke or focus change
- Key history by project path (use `project.uuid` + `InstallPaths.navigationFile`)
- Put history JSON inside `.provenencia` or UserDefaults
- Fuse navigation back into `WorkspaceModel` (chrome-only: sidebar collapse)
- Disable ⌘[ / ⌘] menu items based on `canGoBack` / `canGoForward` in
  `ProvenenciaApp.commands` — Scene commands often freeze the first `.disabled`
  state; `goBack` / `goForward` already no-op at the ends

## Related

- Persist/API: `NavigationHistoryStore`, `WorkspaceNavigation.attachProject`
- Keyboard: `NavigationCoordinator` + `ProvenenciaApp` (`⌘[` / `⌘]`)
- Catalog session for destination loads: [`use-catalog-session`](../use-catalog-session/SKILL.md)
- Place registry + query cache: [`add-workspace-place`](../add-workspace-place/SKILL.md)
