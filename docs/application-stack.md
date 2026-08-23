# Provenance Genealogy Software — Proposed Application Stack

## Status

This document captures the current proposed application architecture and technology stack. It is intended as a working design direction rather than a final implementation specification. Domain table schemas live in the Source, Interpretation, and Conclusion data-model docs, not here.

**MVP implementation:** ship only what the next dogfood step needs (Source catalog and ingest first). Do not pre-build Interpretation, Conclusion, GEDCOM, or unused FFI RPCs. Architecture already constrains *how* those later pieces attach; it does not require implementing them yet.

The primary goals are:

- local-first operation;
- strong user ownership and control of data;
- offline capability;
- long-term interoperability;
- native desktop experience;
- a clean path from a personal macOS application to a commercial cross-platform product;
- shared domain and persistence logic across platforms;
- optional collaboration and cloud synchronization without making cloud services authoritative.

## 1. Platform Strategy

The initial application would target **macOS as a native application**.

If the application is later commercialized, additional clients could be developed for Windows, potentially Linux, and potentially the web.

The architecture should therefore avoid making the genealogy domain model, storage layer, import/export logic, or synchronization system dependent on Apple technologies.

```text
                         Shared Go Core
                              │
           ┌──────────────────┼──────────────────┐
           │                  │                  │
           ▼                  ▼                  ▼
        macOS              Windows            Linux
     Swift/SwiftUI        C#/WinUI        future native UI

                              │
                              │ shared protocols/data formats
                              ▼
                         Web / Server
```

The native UIs are platform-specific. The genealogy engine underneath them is shared.

## 2. Core Technology Choices

### Go

Go would implement the portable application core.

This is attractive because it runs on macOS, Windows, Linux, and servers; is well suited to domain logic and data processing; has mature SQLite support; can be compiled into native shared libraries; allows substantial core reuse by future synchronization or web services; and naturally supports command-line tooling.

The Go core would own functionality such as:

```text
domain model (Source, Interpretation, Conclusion)
database access
schema migrations
query logic
file ingest and derivatives
typed-graph validation
search
sameness / membership closure
canonical merge
audit / revision history
sync algorithms (future)
import adapters (GEDCOM and others, later)
```

The platform UI should not reimplement this logic.

## 3. macOS Client

The first client would be a native macOS application built primarily with Swift, SwiftUI, and AppKit where necessary.

SwiftUI would provide the main application UI, while AppKit could be used where macOS-native functionality or more mature desktop controls are required.

The macOS layer would primarily own windows, navigation, menus, toolbars, keyboard commands, drag and drop, file dialogs, document integration, accessibility, platform-specific presentation, and view state. It should not directly implement genealogy business rules.

```text
SwiftUI / AppKit
       │
       ▼
Swift application layer
       │
       ▼
FFI boundary
       │
       ▼
Go genealogy core
       │
       ▼
SQLite + files
```

## 4. Future Windows Client

A commercial Windows client could be implemented using C# and WinUI 3. The Windows client would implement its own native presentation layer while calling the same Go core used by macOS.

```text
C# / WinUI
     │
     ▼
native interop
     │
     ▼
Go genealogy core
     │
     ▼
SQLite
```

This avoids rewriting genealogy rules, persistence logic, migrations, search, merge logic, synchronization, validation, and most non-UI application behavior.

The Windows application would still require its own UI implementation and platform integration. This is an intentional tradeoff: native UIs require more development effort than a shared web UI, but provide better platform integration and allow each application to behave naturally on its host operating system.

## 5. Linux

The same core architecture can support Linux. The Go core can be built as a Linux shared library and used by a future Linux UI.

The unresolved decision would primarily be the Linux frontend toolkit, for example GTK, Qt, or another native UI framework. Linux is therefore not an immediate target but should not require redesigning the genealogy engine or data model.

## 6. Foreign Function Interface

The native applications communicate with Go through an **FFI — Foreign Function Interface**.

Go can compile the core as a shared library exposing a C-compatible ABI:

```text
macOS      .dylib
Windows    .dll
Linux      .so
```

The platform applications call exported functions from that library. The FFI should be treated as a deliberate application boundary and should not expose the internal Go object graph directly.

Avoid a very granular interface such as `getSourceTitle()`, `getArtifactMimeType()`, or (later) `getPersonBirthDate()`. Prefer application-level operations. The first slice is Source-layer, for example `createSource()`, `addArtifact()`, `ingestFile()`, `getSourceWorkspace()`, `listSources()`. Later slices add Interpretation and Conclusion operations (`createCitation()`, `createObservation()`, `createSamenessClaim()`, `getPersonWorkspace()`) without changing this granularity rule.

This minimizes coupling and reduces FFI chatter.

## 7. FFI Data Format

**Decision:** Protocol Buffers for all structured FFI payloads, from the first slice.

A `.proto` tree is the language-independent specification of the application API. Generated types in Go, Swift, and later C# stay aligned. Field numbers support evolution. The same messages can later ride a sync or HTTP adapter without a second schema. Performance is not the motivation; the contract and codegen are.

```text
                api/proto/*.proto
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
         Go         Swift         C#
```

```text
Swift generated message
    │
protobuf encode
    ▼
byte buffer
    │
FFI (C ABI)
    ▼
Go
    │
protobuf decode
    ▼
domain use-case
```

JSON remains useful for logs, debug dumps, and tests if a message is printed; it is not the FFI codec. The Mac and Go builds both need `protoc` (and the Swift/Go protobuf plugins).

Structured messages carry ids, catalog fields, and small metadata. **File bytes do not go in protobuf** (no `bytes` fields for scans).

**Read path:** Go returns `file_id`, MIME type, `original_filename`, and a path **relative to the project directory** (e.g. `objects/8f/ce/8fce3b…`). Swift already holds the folder it opened and reads the file with AppKit/SwiftUI. SQLite still stores no absolute paths.

**Write path:** ingest only through a use-case (`ingestFile` / equivalent). Swift must not rewrite `objects/` itself.

A later fetch-from-cloud cache can fill the same relative path before display. App Sandbox, if used, is security-scoped access to the **whole project directory**, not per-file bookmarks in the proto.

## 8. Asynchronous Behavior

FFI calls themselves are ordinary synchronous native calls. A long-running Go operation must therefore not block the macOS main UI thread.

The Swift application should generally expose asynchronous APIs such as:

```swift
let source = try await store.source(id: sourceID)
```

Internally, the operation may invoke a synchronous Go function on non-UI execution.

Long-running tasks may warrant true asynchronous operations within the Go core, including file ingest and derivative generation, search index rebuilding, large imports, synchronization, and large exports. The Go runtime can use goroutines internally.

## 9. Persistence

### SQLite

SQLite is the proposed canonical local database.

Advantages include embedded operation, no database server, excellent portability, transactional safety, mature tooling, predictable backups, cross-platform database files, and support for indexes, JSON functionality, FTS, and other useful features.

The Go core should own database access:

```text
Native UI
    │
    ▼
Go core
    │
    ▼
SQLite
```

Swift or C# should not independently issue queries against the same database. This creates a single authoritative persistence implementation across platforms.

## 10. SQLite driver

**Decision:** Official SQLite C amalgamation via cgo, using [`github.com/mattn/go-sqlite3`](https://github.com/mattn/go-sqlite3).

```text
Go
 │
cgo (mattn/go-sqlite3)
 │
SQLite C amalgamation (bundled in the driver, not the OS libsqlite)
```

Open the database through `database/sql` with that driver. Connection hygiene for SQLite:

- **WAL** and a **busy timeout** in the DSN (and `PRAGMA foreign_keys = ON` on each connection).
- **Tiny pool:** `SetMaxOpenConns` small (typically 1 for writers, or a single shared connection). `database/sql`’s default pool fights SQLite.

Enable compile features the product needs with driver build tags (at least `fts5` and JSON when those land). Do not link against macOS’s system SQLite.

**Escape hatch:** if the cgo + Swift dylib build becomes untenable, a pure-Go engine (`modernc.org/sqlite` or similar) can open the same `.sqlite` files. That is a packaging change, not a schema change.

## 11. Bundled SQLite

The application controls its SQLite implementation through the amalgamation above, rather than whichever SQLite the OS ships.

That keeps FTS5, JSON, R-tree, custom functions, and `STRICT` behavior consistent across macOS, Windows, and Linux. The SQLite **file** remains portable across platforms.

## 12. Project directory (document format)

**Decision:** A project is an **ordinary directory**. That directory is the portable, documented interchange format. It is not a proprietary container, zip-as-working-copy, or a single SQLite file that embeds scans.

```text
Robins Family.provenance/         # folder; Mac may show a package icon (optional)
├── provenance.sqlite             # catalog (open schema; use any SQLite viewer)
├── provenance.sqlite-wal         # present while the app is open (WAL)
├── provenance.sqlite-shm
├── objects/                      # ingested Files (see Source layer)
└── derivatives/                  # thumbnails, previews; disposable
```

The public wrapper suffix is **`.provenance`**, not `.genealogy`. It is a branding and Finder hint. Any app can create a folder with that suffix; Provenance must not treat the name as proof of ownership.

On macOS, register the folder as a document package with an exported reverse-DNS UTI (for example `app.provenance.project`) whose imported filename extension is `provenance`. Windows and Linux see a directory. Copying a project to another machine is **copy this whole directory** (after quit or including WAL/SHM). The same Go core opens it; there are no Mac-only paths in the database.

**Open by content:** the suffix and UTI are hints. Create/open must verify `provenance.sqlite`, a documented SQLite `application_id` (32-bit, assigned at implementation), and a schema/`user_version` this engine understands. A `.provenance` folder with a foreign catalog is refused, not migrated. Do not add a parallel `metadata.json` merely to claim the folder.

**Interchange:** Document the **directory layout + schema**. The `.sqlite` file is an inspectable catalog, not a complete project by itself (evidence bytes live under `objects/`). Users may browse tables and files with other tools. Provenance must not rely on obfuscation.

**Where it may live (v0 / MVP):** the **app runs on the local machine** and the **project directory lives on a local disk** (internal SSD, or USB that is not a sync client). Network-open and cloud-open are out of scope.

**Backups (not MVP):** later, copy or sync a **closed** project directory (or an object-store export) to a home server, internet backup, or S3-compatible storage. Variants can include scheduled copy-on-quit or periodic snapshots. Those are optional features. They must not turn iCloud/Dropbox/SMB into the live working copy.

**Where a live database must not live:**

- **iCloud Drive, Dropbox, Google Drive, etc.** — they sync files, they do not provide SQLite’s locking. Silent corruption.
- **SMB/NFS/AFP network shares (typical NAS)** — SQLite needs working POSIX/Windows locking and, in WAL mode, the main DB and WAL on the same reliable lock domain. Consumer NAS locking is famously incomplete. “Home server folder, Mac opens it over the LAN” is the same class of risk as iCloud, even if the NAS is not on the internet.

**Working-copy encryption (MVP):** none. The catalog and `objects/` are **readable** (inspectable SQLite, real JPEG/PDF bytes). Secrecy for a stolen disk is **OS volume encryption** (FileVault, BitLocker, LUKS), not Provenance.

**Later (not MVP):** optional **locked** projects (user key; encrypt catalog **and** objects). Convert readable ↔ locked is a full rewrite. **Backup/export** payloads may be encrypted separately without changing the live readable directory.

**Object names:** Files on disk are named by **SHA-256 of their bytes** (see [`source-layer-data-model.md`](source-layer-data-model.md)), sharded as `objects/{hh}/{hh}/{checksum_hex}`. `original_filename` stays in SQLite. Hash names are identity, not encryption: a JPEG is still a JPEG in Preview.

**Concurrency (MVP):** one live **writer** per project directory. A second Provenance process opening the same folder is **refused** (or the existing window is focused). Two engines must not edit one `provenance.sqlite`. In-process: one Go core, WAL, busy timeout, tiny pool (§10); Swift does not open the SQLite file. Split views of one tree are a **later UI** on that single engine. Quit before writing the catalog with another tool; the CLI uses the same exclusive-open rule.

Platform-specific absolute paths must not be stored in the genealogy database.

## 13. File ingest

Documents, photographs, scans, PDFs, and other evidence should normally be **ingested into the genealogy project**, rather than referencing their original filesystem locations.

This prevents research from depending on fragile absolute paths. A copied genealogy project should retain its associated evidence. A Source with no Artifact, and an Artifact with no File, remain valid (physical-only evidence, or a catalog stub).

## 14. Artifact and File model

The Source-layer model is authoritative ([`source-layer-data-model.md`](source-layer-data-model.md)). The stack must not invent a parallel media catalog.

```text
Source
  └── Artifact  →  File (optional)
                     └── file_derivatives (thumbnails, …)
```

- **Source** is the evidentiary catalog object (`SRC-…`).
- **Artifact** is one representation of that Source (`ART-…`): a scan, PDF, photo, or a placeholder with no File yet.
- **File** is an immutable, content-addressed byte stream (`checksum_sha256`). Storage path is derived from the checksum, not stored as a competing truth. MIME type, size, and `original_filename` live on the File.
- **Citations** (Interpretation) reference an Artifact, not a filesystem path, cloud URL, or File id as the citation target. Displaying the scan uses the Artifact's current File.

Replacing a scan ingests a new File and updates the Artifact pointer. Old Files stay for audit reconstruction.

## 15. Content-addressed Files

File blobs are content-addressed as in the Source layer:

```text
SHA-256(file contents)
        │
        ▼
8fce3b...
        │
        ▼
objects/8f/ce/8fce3b...
```

Benefits include automatic duplicate detection, integrity checking, efficient synchronization, immutable blob identity, independence from OS filename rules, and the same paths on every platform. The object is still a normal JPEG/PDF: hash names are identity, not encryption. `original_filename` remains File metadata.

## 16. Local-First Identity

The desktop application should not require a cloud account merely to operate.

Three identities should be kept conceptually separate:

```text
OS user
   │
   ▼
Local User (contributor UUID)
   │
   ▼
Optional cloud account
```

The operating system controls access to local files. A local **User** ([`user-identity-model.md`](user-identity-model.md)) is a UUIDv7 stored in **application support** and copied into each project’s `users` table for audit. The project directory does not know which machine it is on. On open, the app compares install UUID to Users already in the file: match; or “continue as this person”; or “that’s not me” (new User, new writes only). Cloud accounts later attach to `users.id`; they are not the OS profile.

This ensures attribution remains meaningful even if cloud services later disappear.

## 17. Cloud Accounts

A cloud account becomes necessary only when the user enables online services such as synchronization, collaboration, sharing, remote backup, or family invitations.

The account should be associated with an existing local User rather than replacing it.

```text
User
   │
   └── optional CloudAccount
```

Cloud authentication therefore answers, “Who is making this network request?” The User answers, “Who authored this research?” These should remain separate concepts.

## 18. Authorization

For collaborative projects, the server may maintain workspace membership separately from research authorship.

```text
WorkspaceMembership
  workspace_id
  account_id
  user_id
  role
```

Possible roles might include owner, editor, and viewer. Authentication, authorization, and research authorship therefore remain distinct concerns.

## 19. Local-First Collaboration

Cloud services should not become the canonical owner of the genealogy database. Each native client should retain a complete local database.

```text
Mac
  local SQLite
       │
       │
       ▼
    Sync service
       ▲
       │
Windows
  local SQLite
```

The cloud service coordinates replication rather than becoming the only authoritative database. The application should continue to function offline.

## 20. Sync Server

A future synchronization service could also be written in Go. This potentially allows shared packages for domain definitions, mutation models, validation, sync rules, conflict handling, and serialization.

The server might use PostgreSQL even though desktop clients use SQLite.

```text
Desktop Go Core
      │
      │ sync protocol
      ▼
Go Sync Server
      │
      ▼
PostgreSQL / object storage
```

Persistence details should not dictate the domain model.

## 21. Cloud Media

When collaboration is enabled, local File blobs can be synchronized into object storage such as S3, R2, B2, or an equivalent service.

The genealogy database should still not store provider-specific URLs. It should retain stable File ids and checksums. The synchronization layer maps those identifiers to cloud object keys.

## 22. Lazy File Synchronization

A shared tree could eventually contain many gigabytes of scans and photographs. A new collaborator should not necessarily be required to download every original File immediately.

A useful model would synchronize Artifact/File catalog rows, filenames, hashes, MIME types, and derivatives first, while fetching full-resolution blobs on demand.

```text
User opens an Artifact
      │
      ▼
Is File local?
   │       │
  yes      no
   │       │
   ▼       ▼
display   fetch from cloud
              │
              ▼
           cache locally
```

This allows large collaborative repositories while retaining local access to actively used content.

## 23. Web Application Path

The shared Go architecture does not prevent a future web client.

A conventional web implementation could use:

```text
Browser
   │
HTTP / WebSocket
   │
Go server
   │
shared genealogy packages
   │
database
```

The Go domain logic should not know whether it was invoked through FFI, HTTP, CLI, or sync. Instead, platform-specific adapters invoke the same application services.

```text
                     Go Core
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
       FFI             HTTP             CLI
        │               │
     Native UI       Web client
```

## 24. Local-First Web

A conventional server-backed web client would not provide the same degree of local-first ownership as the desktop application.

A future local-first web version could instead use a browser-local database and File cache with a sync client connecting to the Go server. Potential browser storage technologies include SQLite compiled to WebAssembly and persistent browser filesystem/storage APIs.

This is considered feasible but is not necessary for the initial application.

## 25. Command-Line Tooling

A Go core creates a natural opportunity for an independent CLI.

Possible commands could eventually include:

```text
genealogy validate
genealogy import
genealogy export
genealogy search
genealogy migrate
genealogy stats
genealogy repair
```

The CLI would be valuable for automated tests, database recovery, migrations, diagnostics, developer tooling, power users, and batch operations. This also provides a way to exercise the core independently of any UI.

## 26. Proposed Repository Shape

A possible repository structure is:

```text
/
├── core/
│   ├── domain/
│   ├── application/
│   ├── database/
│   ├── migrations/
│   ├── import/          # later adapters (GEDCOM, …)
│   ├── search/
│   ├── merge/
│   ├── sync/
│   ├── files/
│   └── validation/
│
├── api/
│   ├── ffi/
│   └── proto/
│
├── cmd/
│   ├── genealogy-cli/
│   └── genealogy-server/
│
├── macos/
│   ├── App/
│   ├── Views/
│   ├── ViewModels/
│   └── Platform/
│
├── windows/
│   └── future/
│
├── web/
│   └── future/
│
└── docs/
```

The exact package boundaries can evolve, but the dependency direction should remain clear:

```text
platform UI
     │
     ▼
application API
     │
     ▼
domain/core
     │
     ▼
persistence
```

Platform code should not leak downward into the core.

## 27. Architectural Principles

### The application is not the data

The user's research should survive replacing the UI, switching operating systems, discontinuation of the commercial service, and loss of cloud access.

### Cloud services are optional capabilities

An account may provide sync, sharing, backup, and collaboration, but should not be required to open or edit local research.

### Genealogy semantics remain platform-independent

Core concepts such as sources, artifacts, files, interpretations, claims, conclusions, citations, users, relationships, events, places, and change history must not depend on SwiftUI, WinUI, HTTP, or any particular storage provider.

### Platform UIs are replaceable clients

SwiftUI is the first presentation layer, not the application architecture.

### Local data is canonical

The desktop client should be capable of operating normally without a network connection.

### Storage references are logical, not physical

Domain objects reference stable IDs. They should not directly reference absolute filesystem paths, S3 URLs, or platform-specific identifiers.

### Auditability is first-class

Authorship and change history are part of the research record ([`audit-revision-history.md`](audit-revision-history.md)), not `created_by` columns on domain tables.

## 28. Current Preferred Initial Stack

For the first macOS implementation:

```text
UI
  Swift
  SwiftUI
  AppKit where appropriate

Core
  Go

Interop
  C ABI / FFI
  Protocol Buffers (generated Go / Swift types)
  Go core as xcframework/dylib (cgo + mattn); spike before Source UI

Persistence
  SQLite via github.com/mattn/go-sqlite3 (cgo, bundled amalgamation)
  WAL + busy timeout; small connection pool

Media
  project-local content-addressed Files
  derivatives beside objects, not a second catalog

Document format
  project = *.provenance directory (provenance.sqlite + objects/ + derivatives/)
  interchange = that directory, not sqlite-alone
  live DB on local disk only (not iCloud / Dropbox / SMB NAS)

Accounts
  none required for local use

Identity
  local User (UUID + display name)

Cloud
  none required initially
```

A later commercial architecture could add:

```text
Windows
  C#
  WinUI 3

Linux
  native frontend TBD

Server
  Go

Server database
  likely PostgreSQL

Blob storage
  S3-compatible object storage

Sync
  shared Go protocol/domain packages

Web
  frontend framework TBD
  Go server
  potentially local-first browser storage
```

## 29. Decisions Still Open

Settled:

- **FFI codec:** Protocol Buffers (see §7).
- **SQLite:** cgo + official amalgamation, `github.com/mattn/go-sqlite3` (see §10). WAL, busy timeout, tiny `MaxOpenConns`.
- **Project format:** inspectable directory named `*.provenance`; catalog is `provenance.sqlite`; Files on disk as SHA-256 object names (`objects/{hh}/{hh}/{hex}`); copy whole folder across Mac/Windows. Suffix is a hint; ownership is `application_id` + schema (see §12).
- **Runtime:** app and live project on **local disk** only. Backups to NAS/internet/object storage are post-MVP (copy or sync a closed project).
- **Encryption:** live project is plaintext; rely on OS disk encryption. Optional locked projects and encrypted backups are post-MVP (see §12).
- **Locking:** one writer per project; second process refused. Multi-pane UI later, same engine (see §12).
- **Identity:** install-local User UUID vs project `users`; adopt or mint on first open. Cloud maps onto that UUID later ([`user-identity-model.md`](user-identity-model.md)). OS profile is not the IdP.
- **File display:** protobuf has relative paths + metadata only; Swift reads bytes from the project directory (§7).
- **Build order:** implement only the use-cases needed for the current step. First FFI RPCs are Source-shaped (`OpenProject`, `CreateSource`, `IngestFile`, …). Person/GEDCOM verbs wait until that step exists. Adding a proto method later is expected; freezing a full ABI up front is not.

**Deferred (post-MVP):** sync protocol, mutation/change-log format, server-side persistence, conflict-resolution semantics, Linux frontend, web frontend. They do not block the Source-layer prototype.

## 30. Summary

The proposed architecture treats the genealogy system as a **portable Go application engine with native platform clients**.

The initial application would be:

```text
SwiftUI
   │
   ▼
Go Core
   │
   ├── genealogy domain logic
   ├── SQLite
   ├── files
   ├── validation
   ├── audit
   └── future: import adapters, sync
```

This allows development to begin as a polished native macOS application without committing the underlying genealogy system to Apple.

If commercialization later warrants Windows support, a native Windows client can be built against the same Go core. Linux and web clients can follow without changing the fundamental storage or domain model.

The architecture prioritizes local-first operation, ownership, privacy, interoperability, and long-term durability while preserving a practical route to a multi-platform commercial product.

Authoritative domain schemas: [`source-layer-data-model.md`](source-layer-data-model.md), [`interpretation-layer-data-model.md`](interpretation-layer-data-model.md), [`conclusion-layer-data-model.md`](conclusion-layer-data-model.md). This document owns runtime and packaging, not table shapes.

---

# 31. Architectural risks to address

These are known tensions. Several are already constrained by §12 and §29 (no live iCloud/NAS, directory is the format, GEDCOM is an adapter, exclusive writer, MVP-only RPCs). Remaining items to pin for implementation are called out below.

1. **File bytes and FFI** — **Decided:** relative path + metadata in protobuf; Swift reads the object on disk; ingest-only writes (§7).

2. **Synced and network filesystems** — Live projects are local-disk only (§12). Backups (NAS, internet, object storage, scheduled copy-on-quit) are post-MVP. Product still needs to refuse iCloud/Dropbox/SMB as the working copy.

3. **Import adapters vs core** — GEDCOM (and similar) maps poorly onto Source / Interpretation / Conclusion. Keep import as a later adapter so the engine is not designed around a GEDCOM person table.

4. **Typed-graph validation vs convention** — Core validation should reject malformed rows (value column vs `value_type`, FK shape), not one-grain Places or empty Persons. Warnings and badges are UI.

5. **cgo SQLite + Swift dylib** — **MVP plan: A.** Spike mattn/go-sqlite3 inside a Go xcframework loaded by Swift (open DB, trivial write, round-trip) **before** Source UI. If that build is untenable, fall back to a local Go process (B) or pure-Go SQLite in the dylib (C). On-disk project format does not change.

6. **Project wrapper vs source of truth** — The directory is the format. No parallel `metadata.json` catalog. Optional Mac package icon must not hide or encrypt contents.

7. **Sync of the research graph** — Replicating Observations, Sameness components, and identity anchors is much harder than syncing a person table. Shared Go structs on a PostgreSQL server do not by themselves define conflict semantics. Stay TBD until a real multi-device need exists.

8. **First FFI surface** — **Not an architecture fork.** Implementation order only: add RPCs when a screen needs them. Coarse use-cases stay the rule (§7); the *list* of methods grows with the prototype.
