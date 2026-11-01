---
name: use-catalog-session
description: >-
  Uses Provenencia's held exclusive catalog session (core/catalogsession) so
  researcher FFI and onboarding paths serialize ops instead of open-per-RPC.
  Use when adding or changing catalog FFI handlers, withProjectCatalog,
  catalogsession.Do/Close, OpenCatalog vs session, catalog.already_open,
  CloseCatalogSession, GenealogyStore.closeCatalogSession, WorkspaceView leave,
  SignOut catalog close, or overlapping store/catalog calls.
---

# Use the catalog session

Researcher catalog access is a **held exclusive session** plus a **serial queue**, not open → work → close per RPC.

Authoritative product note: [`docs/ideas/archive/catalog-access-serialization.md`](../../../docs/ideas/archive/catalog-access-serialization.md). Stack: [`docs/application-stack.md`](../../../docs/application-stack.md) §10 / §12.

## Doorways

| Path | Use |
| --- | --- |
| `catalogsession.Do(projectDir, fn)` | Researcher work that needs `*database.Catalog` (onboarding list/info/open, non-handler core) |
| `withProjectCatalog(projectDir, fn)` | **FFI handlers only** — thin wrapper around `Do` in [`api/ffi/handlers/catalog.go`](../../../api/ffi/handlers/catalog.go) |
| `catalogsession.Close(projectDir)` / `CloseAll()` | Release the lock (tests, SignOut, project switch) |
| `METHOD_CLOSE_CATALOG_SESSION` | Explicit FFI close for one `project_dir` |
| `onboarding.OpenCatalog` | One-shot open (tests / low-level). Prefer `Do` for product paths |
| `database.Open` / `Create` | Migrate-only / package tests. **Never** from FFI handlers |

`Do` get-or-opens once (same open policy as `OpenCatalog`: migrate + `catalogsession.ReconcileOnOpen`, the one list of open-time repairs: `users.EnsureRefs`, `evidencegraphs.EnsureRefs`, `project.EnsureUUID`, the derived caches), holds the connection, and runs `fn` under a per-session mutex. Opening a **different** `projectDir` closes other held sessions first (one live project).

## FFI catalog handler pattern

```go
func ListThings(in []byte) ([]byte, error) {
	var req engine.ListThingsRequest
	if err := proto.Unmarshal(in, &req); err != nil {
		return nil, unmarshalErr("list_things", err)
	}
	// Parse/validate IDs before the session when possible.
	var out *engine.ListThingsResponse
	err := withProjectCatalog(req.GetProjectDir(), func(c *database.Catalog) error {
		rows, err := things.List(c)
		if err != nil {
			return err
		}
		out = &engine.ListThingsResponse{…}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(out)
}
```

- Do **not** `c.Close()` inside the callback — the session owns the handle.
- Do **not** call `database.Open`, `onboarding.OpenCatalog`, or a leftover `openProjectCatalog`.
- New catalog RPCs: still follow `.agents/skills/add-ffi-handler/SKILL.md`, then this pattern.

## Lifecycle

- **Open (get-or-open):** first `Do` / `withProjectCatalog` / Mac catalog store call for that dir (typically `CatalogCounts.refreshAll` after workspace mount). No `OpenCatalogSession` FFI. The open runs `catalogsession.WarmOnOpen`, the one list of open-time warms: the built alias index (`namealiases.Warm`), the default name format (`nameformats.Warm`), then matching's shortlist kinds and shortlist index (`promotealign.Warm`), held on the catalog graph for the life of the project. A matching warm that fails never fails the open: the shortlist index stays unbuilt and the next proposal builds it and reports the error. `Do` runs the alias index and name format warm again (not the shortlist index, which waits for the next proposal) after an `fn` that left `Graph().WarmNeeded()` true (a `Drop`, a vocabulary or structure commit, a new alias index or name format); a handle commit does not set it, a failed `fn` skips it, and a warm error never replaces `fn`'s result. Readers go through `namealiases.Rules` and `nameformats.LoadFor`; a writer calls the owning package's `Invalidate`. A new open-time warm is added to `WarmOnOpen`, not called inline. Closing the session drops the graph.
- **A panicking callback:** the panic goes on to `ffi.Call`, which returns `internal.unknown`. `Do` closes that session and drops it, so the next call reopens the catalog on fresh caches; handlers never `recover`.
- **Mac leave:** `WorkspaceView.onDisappear` → `GenealogyStore.closeCatalogSession(projectDir:)`. Destination content mounts immediately; overlapping loads are fine (Go queues).
- **Also close:** Go `SignOut` / `RemoveActiveProject` / `OpenProject` (switch) call `Close`/`CloseAll`.
- **Create (`Complete` / `createCatalog`):** create-then-close; do not register a closed handle in the session map. Later RPCs open via `Do`.

## File work before the session

Copying, hashing, decoding and encoding never run inside `Do`. A handler stages first, outside it, and enters the session only to look up, insert and install:

- **Imports and depictions:** `withStaged(projectDir, stage, fn)` (and `withStagedDepictions`) in [`api/ffi/handlers/catalog.go`](../../../api/ffi/handlers/catalog.go). `stage` makes the request's `ingest.Stages` (`ingest.StageFile`, `StageDepictions`) from the normalized project folder; `fn` runs in the session with `stages.Settle(c)` deferred first. Temps live in `objects/` (`.ingest-`, `.stage-`).
- **Thumbnails:** `thumbnailPass` (`api/ffi/handlers/thumbnail_pass.go`): `need` inside the first session (`derivatives.Plan`), `run` after it (`Job.Stage` outside, `derivatives.Commit` in batches of 16). `SetSourceCover` shares the last batch's session.
- **Install and purge only inside `Do`.** `Staged.Settle(c)` is the one in-session cleanup; `Staged.Discard()` runs outside. A purge and an import of the same bytes never interleave.
- **Tests:** `catalogsessiontest.Hold(t, projectDir)` holds the session so a test can see the staged temp appear before the release. Build fixtures through `catalogsession.Do`, never on a `database.Open` catalog (`ErrAlreadyOpen`).

## `catalog.already_open`

Means something **bypassed** the session (second `database.Open` while held). Overlapping FFI calls must **queue** on `Do`, not fail. Treat `already_open` in product paths as a bug to fix, not normal UX.

## Tests

- After FFI/`Do` use: `t.Cleanup(func() { _ = catalogsession.CloseAll() })` (`runRPC` already does this).
- Assert against the catalog with `catalogsession.Do`, not `database.Open`, while a session may be held.
- External test package (`catalogsession_test`) if the test imports `onboarding` (avoids import cycle).
- Mac: `FakeStore` tracks `heldCatalogProjectDir` / `lastClosedCatalogProjectDir` (`CatalogSessionStoreTests`). New catalog store entry points on FakeStore must call `markCatalogSessionHeld`.
- Badge refresh (`CatalogCounts.refreshAll`) must not swallow errors silently — set `lastRefreshError` / surface via workspace toast.
- Gate: `CGO_ENABLED=1 go test ./core/catalogsession/ ./core/onboarding/ ./api/ffi/...` and Mac ProvenenciaTests for store/workspace changes.

## Do not

- Open-per-RPC in handlers or onboarding list/open/info
- Nest a second exclusive `database.Open` for the same dir while `Do` holds it
- Put session registry logic in `database` or Swift (Go owns the contract)
- Reintroduce `isCatalogReady`-style gates for catalog locking
- Call `ingest.File`, `derivatives.Ensure` or `derivatives.EnsureThumbnail` from a handler (they do file work inside the session; tests and fixtures only), or copy, hash or decode inside `withProjectCatalog`
