# Spike 1 — Scaffold, local project, first-run onboarding

**Done.** Proved the stack: SwiftUI → protobuf FFI → Go core → SQLite + files on disk. Live models: [`application-stack.md`](../../application-stack.md), [`user-identity-model.md`](../../user-identity-model.md).

## Decisions

- **No account, no network** to use the app. OS / iCloud username is not the identity provider.
- **Install identity** lives in Application Support (`identity.json`: `user_id`, `display_name`, `USR-…` ref), not inside the project.
- A project is a **`*.provenencia` folder**: `provenencia.sqlite`, `objects/`, `derivatives/`. Folder basename is a kebab slug of the family/project label.
- Contributor UUID is minted or adopted from the catalog; the project `users` row is the attribution source of truth.
- **cgo SQLite in a Go dylib** was the early packaging risk. Plan A shipped (`libprovenencia.dylib`, `-buildmode=c-shared`). Fallbacks unused.
- File bytes never travel over protobuf. Swift does not query SQLite.

Open/create polish and Source catalog came later (Spike 2).
