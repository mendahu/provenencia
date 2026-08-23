# Spike 1 — Scaffold, local project, first-run onboarding

## Status

Implementation milestone. Not the MVP. The stopping point is: Jake can build and run Provenance on his MacBook, complete first-run onboarding, and see a durable local identity in Application Support plus a matching `users` row in a real project database.

Authoritative models: [`application-stack.md`](../application-stack.md), [`user-identity-model.md`](../user-identity-model.md). This note only sequences the first slice of those decisions.

---

## Goal

A new researcher launches the macOS app with no cloud account. The app asks whether they already have a project, then either mints a contributor or adopts one from an existing file, writes identity on the machine, and opens or creates a named project folder.

```text
Launch
  → have a file already? / create new
  → create: display name + family name → mint User
  → open: list users in the catalog → adopt that UUID or mint a new one
  → write identity file (Application Support)
  → create or open {Family Name}.provenance
  → users row matches identity
  → idle “you’re in” screen
```

That is enough to prove the stack: SwiftUI → protobuf FFI → Go core → SQLite + files on disk.

---

## In scope

- Repository layout for `core/`, `api/`, `cmd/`, `macos/` as sketched in the stack doc.
- Go core as an xcframework/dylib with cgo SQLite (`github.com/mattn/go-sqlite3`).
- Protocol Buffers for FFI payloads.
- Project directory on **local disk**: `*.provenance` folder with `provenance.sqlite` (+ WAL/SHM while open), empty `objects/` and `derivatives/` as needed.
- Schema enough for identity: `users` (`id` BLOB UUIDv7, `display_name` TEXT), SQLite `STRICT`, migrations owned by Go.
- Install-local identity file in Application Support (outside the project). Same UUID reused for later projects.
- First-run UI: collect **your name** (contributor display name) and a **project / family name** (this document). No email, password, or login.
- Create a first project as part of onboarding. Default parent: Documents. Folder basename is the project name plus `.provenance` (for example `Robins Family.provenance`).
- Returning launch: if an active project is stored and still on disk, skip the picker and open it. Identity plus folders in Documents is not enough.
- Developer path: `open` the Xcode project / run the Mac target and talk to a live Go core and SQLite file on this machine.

## Out of scope (later spikes / MVP)

- Source catalog, file ingest, Interpretation, Conclusion, GEDCOM, search, merge, sync, cloud accounts.
- Full audit/revision tables and `create_project` as an audited revision. Spike 1 may insert `users` without a complete audit log; the next spike that mutates research data should add audit before those writes.
- Windows, Linux, web, CLI product commands (a tiny Go test binary for the core is fine).
- App Sandbox polish, notarization, Sparkle, icons-as-document-package beyond “it is a folder.”
- Encryption, iCloud/Dropbox/SMB as working copy (still refused if we detect it; no need to over-build the detector).

---

## Requirements

### Product

1. No network and no account to use the app.
2. First launch with no identity file: ask whether they already have a project or want a new one. Then either collect **your name** and a **project / family name**, or list contributors in the opened file so they can adopt an existing `users.id` or mint a new one. Empty/whitespace names are rejected.
3. Confirming creates:
   - a UUIDv7 contributor id;
   - an identity file in Application Support for this install (your name + UUID only — not the project name);
   - a project directory `{sanitized family name}.provenance` with `provenance.sqlite`;
   - a `users` row with that id and display name.
4. Quit and relaunch: do not ask for your name again; the same UUID is used. Reopen the last project when an active-project pointer exists.
5. Display name is attribution text, not a login. It need not be unique.
6. Project / family name labels **this file** in Finder and the window title. It is not the User identity. Spike 1 does not need a separate title table: the folder basename is the name. Renaming the project later can be a Finder rename or a later use-case.
7. Sanitize the folder name for the filesystem (`/`, `:`, leading dots, etc.). If `{Name}.provenance` already exists in the chosen parent, fail with a clear error (do not silently overwrite).
8. OS account is not the identity provider. Do not key identity off macOS username or iCloud. Do not default the project name to the macOS account name without asking.

### Identity file

Store under the app’s Application Support directory (not inside the project), for example:

```text
~/Library/Application Support/Provenance/identity.json
```

Minimum contents (JSON is fine for v0; this is install-local, not the portable project format):

```json
{
  "user_id": "<uuidv7 canonical string>",
  "display_name": "Jake Robins"
}
```

`user_id` in the file must match `users.id` in the project (16-byte BLOB in SQLite). Display name in the file may be a cache of the last known name; the project `users` row is what later audit UI will show.

Do not put passwords, tokens, or email in this file.

### Project on disk

```text
Some Family.provenance/
├── provenance.sqlite
├── objects/
└── derivatives/
```

Go opens SQLite with WAL, busy timeout, `PRAGMA foreign_keys = ON`, tiny connection pool. Swift does not query SQLite. One writer; a second process opening the same folder is refused (can be a stub error for this spike).

### FFI slice

Coarse use-cases only, for example:

- `GetInstallIdentity` / `CompleteOnboarding`
- `OpenProject` / `GetActiveProject` / `RemoveActiveProject`
- `GetCurrentUser`

No Source RPCs. File bytes never in protobuf.

### Definition of done

Jake can, on his MacBook, without a server:

1. Build the Mac app and the Go dylib in a documented local workflow.
2. Launch, enter your name and a family/project name, finish onboarding.
3. Find the identity file on disk with a UUID and **your** name.
4. Find `{Family Name}.provenance` (e.g. under Documents) and open `provenance.sqlite` to see the matching `users` row.
5. Relaunch and land past onboarding as that same user when an active project is stored.

---

## Technical risk to burn early

The stack already calls this out: **cgo SQLite inside a Go xcframework loaded by Swift**. If that build is untenable, stop and switch to a local Go process or pure-Go SQLite before any onboarding UI. Do not paper over a broken dylib with a fake in-memory store.

---

## PR sequence

PR1–PR9 are done.

Small, reviewable chunks. Each PR should leave `main` buildable. Later PRs may add UI on a stub that the next PR fills in.

```text
PR1 repo scaffold
  └─ PR2 proto + FFI hello
       └─ PR3 cgo SQLite dylib spike (done)
            └─ PR4 migrations + users + create/open project (Go) (done)
                 ├─ PR5 install identity file (Go) (done)
                 │    └─ PR6 onboarding use-case (person + family name → identity + project) (done)
                 │         └─ PR7 FFI GetInstallIdentity + CompleteOnboarding (done)
                 │              └─ PR8 SwiftUI first-run + Application Support path + returning launch (done)
                 └─ PR9 last-project pointer / reopen (done)
```

PR5 can start once PR3 proves the dylib; it should not land identity writes that skip the database. Prefer merging PR4 before PR6 so onboarding has a real catalog.

| PR | Title | Does | Depends on | Notes |
| --- | --- | --- | --- | --- |
| **1** | Repository scaffold | Done. `go.mod`, stub `core/` / `api/proto/` / `cmd/` / `macos/` Xcode app (macOS 14) that launches a blank window. README: how to open and run on a Mac. | — | No genealogy yet. |
| **2** | Proto + FFI hello | Done. `engine.proto`, codegen, Swift `GenealogyStore` / `GoStore` calls `Ping` / `GetVersion` over one C ABI; Go returns protobuf. See `api/proto/README.md`. | 1 | No SQLite yet. |
| **3** | SQLite-in-dylib spike | Done. Temp SQLite via mattn/cgo inside `libprovenance.dylib`; Swift `sqliteProbe()`. Existing c-shared dylib pipeline, not an xcframework. | 2 | **Gate.** If this fails, choose fallback before continuing. |
| **4** | Project format + `users` | Done. Create/open `*.provenance` in Go; `application_id` `'PROV'`; `user_version` 1; empty `users`; exclusive open. No FFI create/open, no CLI. | 3 | Schema: only what Spike 1 needs. Refuse foreign catalogs. |
| **5** | Install identity store | Done. Go `core/identity`: `{dir}/identity.json` (`user_id` UUIDv7 + `display_name`). Caller passes `dir`; no Mac path hardcoded. No FFI, no `users` insert. | 4 | File format pinned in this PR. |
| **6** | Onboarding use-case | Done. Go `core/onboarding.Complete`: mint/load identity, sanitize family name, `Create` `*.provenance`, upsert `users`. No FFI. | 4, 5 | No Source tables. |
| **7** | Onboarding FFI | Done. `GetInstallIdentity` / `CompleteOnboarding` over the existing C ABI; paths in protobuf; failed calls return UTF-8 `err.Error()` in `out`. No Swift UI. | 6 | Go still does not hardcode Mac paths. |
| **8** | Onboarding UI + install path | Done. First-run two fields, home stub with project basename; returning user skips the form if identity exists. Swift resolves Application Support `Provenance/` and Documents. | 7 | Last-project reopen is PR9. |
| **9** | Reopen last project | Done. `active-project.json` next to identity; two-screen create/open then identity; adopt a catalog user on a new Mac or mint. Relaunch goes home if the active folder still exists. | 8 | Path in Application Support, not security-scoped bookmarks. |

Do not combine 3 with 8. Do not put ingest or `sources` into 4 “while we’re in migrations.”

### Suggested PR descriptions (why, not a task dump)

- **1** — Give the repo a runnable Mac shell so later work has a place to land.
- **2** — Lock protobuf-over-C-ABI before domain code grows.
- **3** — Prove the persistence packaging the whole product depends on.
- **4** — Make the inspectable project directory real, with a `users` table.
- **5** — Keep contributor UUID off the project folder and on the install.
- **6** — One core operation for “I am this person; this file is the Robins family.”
- **7** — Expose onboarding over FFI so Swift can pass directories without embedding Mac paths in Go.
- **8** — The only user-visible milestone: first-run names → identity + named `.provenance` folder; Swift supplies the Application Support path.
- **9** — Make the second launch feel like an app, not a demo.

---

## Retrospective — improvement backlog

From a code review of the Spike 1 codebase (Go core, FFI layer, Swift client). Overall: separation of concerns is good (thin dispatch router, domain packages, store protocol + fake in Swift), Go test coverage is strong and table-driven. The real gaps are zero Swift tests, a few correctness bugs around edge-case input and side-effect ordering, and repo hygiene around the committed dylib. Each item below is one PR.

### High value

- **H1 — Fix the SQLite DSN path bug and harden folder-name sanitization.** Done. `dsn()` builds a `file:` URL so `?`/`#` in paths stay in the filename; `FolderName` also replaces Windows-reserved basename characters (`*?"<>|`).
- **H2 — Fix side-effect ordering and silent identity replacement in `onboarding.Open`.** Done. Identity is written only after a successful catalog open; adopt of a different catalog UUID replaces `identity.json` (tested).
- **H3 — Add a Swift unit test target for `OnboardingModel`.** Done. Unit target + `OnboardingModel` / `InstallPaths` coverage via `FakeStore`.
- **H4 — Validate the active project as a catalog, not just a directory.** Done. `isProjectDirectory` / picker require `provenance.sqlite` to exist (no DB open).
- **H5 — Stop growing the repo by ~14 MB per dylib rebuild.** Done. Dylib gitignored; Xcode + CI build via `scripts/build-macos-core.sh`.

### Medium value

- **M1 — Derive dispatch method constants from the generated proto enum.** Done. `dispatch.go` uses `int32(engine.Method_METHOD_…)`.
- **M2 — Deduplicate the onboarding use-cases.** Done. Shared `loadOrMint` (persist timing) and close/`rememberActive` helpers; H2 identity-after-open unchanged.
- **M3 — Make `ListContributors` read-only.** Done. `database.OpenReadOnly` (`mode=ro`, no migrate, no exclusive lock).
- **M4 — Add a macOS CI job that builds the app.** Done. Non-draft PRs to `main` run `xcodebuild test` (unsigned) on `macos-15` via `.github/workflows/macos-test.yml`.
- **M5 — Collapse sign-out into one `SignOut` RPC.** Done. One FFI call removes active-project.json then identity.json; stale last-project still uses `RemoveActiveProject` only.
- **M6 — Map sentinel errors to user-facing copy.** The UI shows raw `err.Error()` text including full paths. Map the known sentinels (`ErrAlreadyExists`, `ErrAlreadyOpen`, `ErrNotAProject`, `ErrUnsupportedVersion`) to friendly strings in Swift; an error-code field over FFI can wait.

### Low value

- **L1 — Handler/test file organization.** `OpenProject` / `ListProjectUsers` live in `handlers/onboarding.go` but their tests live in `session_test.go`; align them. Consider renaming `core/session` (reads like auth sessions) to something like `core/installstate` before more code depends on it.
- **L2 — Dead code sweep.** Obsolete blank import of `core/database` in `api/libprovenance/main.go`; unnecessary `len(in) > 0` guard in `handlers/version.go`; unused `Mode` conformances (`CaseIterable`, `Identifiable`); `FakeStore` force unwrap; stray untracked `macos/Provenance/` folder on disk.
- **L3 — Extract a shared JSON file-store helper** for `core/identity` and `core/session` (Load/Save/Remove/ErrNotFound are near-identical). Optional.
- **L4 — C ABI polish.** Make `in` const in the bridging header and Go export; document the status codes (0/1/2) in the header rather than only in `api/proto/README.md`.
- **L5 — Split `OnboardingView` (~290 lines) into per-screen subview files** once the flow stabilizes.
- **L6 — Ship-prep signing.** The app relies on `disable-library-validation` for the unsigned dylib. Before distribution, sign the dylib with the same team identity and drop that entitlement.
- **L7 — Version bump.** Spike 1 is done and user-visible; set `VERSION` / `core.Version` / `MARKETING_VERSION` to `0.1.0` and tag `v0.1.0` per [`versioning.md`](../versioning.md).

Suggested order: H1 and H2 first (small, real bugs), then H3 + H4 together (the test target pays for itself immediately), then H5 / M4 as infrastructure, with the rest as filler PRs between feature work.

---

## After Spike 1

Next likely spike: open/create project without onboarding, “continue as / that’s not me,” then Source catalog. Audit should arrive with the first researched mutation, not as a decorative table in Spike 1 unless it is cheaper to add an empty `audit_transactions` migration in PR4 to avoid a later rewrite. Prefer empty audit tables in PR4 if the audit doc’s first migration is already small; do not implement revision UI.
