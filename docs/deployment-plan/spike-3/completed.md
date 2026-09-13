# Spike 3 — Completed steps

Finished Spike 3 work kept for history. The live to-do list is [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S3-NN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S3-01](#s3-01--design-app-layout-toolbar) | Design | App Layout toolbar (Back/Forward, breadcrumbs, omnibar field) |
| [S3-02](#s3-02--design-omnibar-results) | Design | Omnibar results dropdown + hit row |
| [S3-03](#s3-03--pr-catalog-projectuuid) | PR | Catalog `project.uuid` (mint, heal, ProjectInfo) |
| [S3-04](#s3-04--pr-workspacelocation--navigation-history) | PR | WorkspaceLocation + persisted navigation history |

---

## Steps

### S3-01 — Design: App Layout (toolbar)

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | — |
| **Deliverables** | Done. Board for the main-column toolbar: Back/Forward (long-press / secondary-click jump menus), flex-grow breadcrumbs, trailing omnibar field (`⌘K`). Layout summary: [`design/README.md`](design/README.md). Boards stay in Claude Design (not checked into git). |
| **Context** | [`navigation-history.md`](navigation-history.md); [`omnibar-search.md`](omnibar-search.md) field placement. |
| **Out** | Results dropdown (S3-02); engine search; history persistence. |
| **Feeds** | S3-05, S3-06 |

---

### S3-02 — Design: Omnibar results

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | S3-01 (omnibar field placement) |
| **Deliverables** | Done. Board for flat ranked rich hit rows (`PVOmnibarHitRow` slot contract), dropdown chrome, empty/loading/keyboard states. Brief archived: [`design/archive/omnibar-results-dropdown-brief.md`](design/archive/omnibar-results-dropdown-brief.md). Summary: [`design/README.md`](design/README.md). |
| **Context** | [`omnibar-search.md`](omnibar-search.md). |
| **Out** | Search engine / FTS; faceted query UI. |
| **Feeds** | S3-10 |

---

### S3-03 — PR: Catalog `project.uuid`

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — |
| **Deliverables** | Done. Migration `000017` adds `project.uuid` (`BLOB`, unique when set). `project.NewID` / Upsert mints UUIDv7 on Create; `project.EnsureUUID` heal-mints on Open / `createCatalog` / `catalogsession` (beside `users.EnsureRefs`). Existing uuid never rewritten on Upsert conflict. Exposed on protobuf / Swift `ProjectInfo.uuid` via Complete, Open, and GetProjectInfo. |
| **Context** | [`navigation-history.md`](navigation-history.md) § Project key; skill [`add-catalog-migration`](../../../.cursor/skills/add-catalog-migration/SKILL.md). Active-project pointer stays path-based. |
| **Out** | History JSON; toolbar; short `PRJ-…` ref. |
| **Dogfood** | New + upgraded projects show a stable UUID on open; rename/move of the folder does not change it. |
| **Feeds** | S3-04 (history file key); later install-local chrome |

---

### S3-04 — PR: `WorkspaceLocation` + navigation history

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-03 (done) |
| **Deliverables** | Done. `WorkspaceLocation` (section + `sourceId` / `fieldId` / `typeId`, denormalized ref/title). `NavigationHistoryStore` owned by `WorkspaceModel`: `go(to:)`, `goBack()`, `goForward()`, `go(toIndex:)`, coalesce identical places, truncate forward, cap 100. Persist `Application Support/Provenencia/navigation/{uuid}.json` (v1). Sidebar, Source open/close/create, and vocabulary row selection all commit via `go(to:)`. Destinations apply `currentLocation` on appear/change; missing entities prune to section root. `⌘[` / `⌘]` via `NavigationCoordinator`. |
| **Context** | [`navigation-history.md`](navigation-history.md). Hand-rolled coordinator — not SwiftUI `NavigationStack` / `NavigationPath`. |
| **Out** | Final toolbar visuals / jump menus (S3-05); omnibar results; search RPC. |
| **Dogfood** | Keyboard Back/Forward restores deep locations across destinations; relaunch returns to the leave-off place for that `project.uuid`. |
| **Feeds** | S3-05 (toolbar chrome); S3-07+ Hit `location` shape |
