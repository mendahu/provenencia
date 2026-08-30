---
name: add-catalog-query
description: >-
  Adds Provenencia catalog table access as a nested domain package under core/database.
  Use when adding or changing SQLite queries, Upsert/Lookup/List for a catalog table,
  core/database/users or a new core/database/<table>/ folder, or when the user asks
  for database query functions without putting another file in core/database/.
---

# Add a catalog query package

Table SQL and functions live in **`core/database/<domain>/`**, not in the `database` package root. Root stays the exclusive catalog client (`Catalog`, `Create`/`Open`/`Close`, format errors, migrations).

New **schema** is a separate step: follow `.cursor/skills/add-catalog-migration/SKILL.md` first if the table does not exist yet.

## Layout

Copy `core/database/users/`.

```
core/database/<domain>/
  <domain>.go       # SQL consts, functions, domain sentinels
  <domain>_test.go  # table-driven tests
```

Package name is the folder name (`users`, later `sources`, …). Call sites use `users.Upsert(c, …)`, not `Catalog` methods.

## Functions

- Take `*database.Catalog`. Get the handle with `c.DB()`; return that error (`database.ErrClosed`).
- Keep SQL as unexported consts next to the functions. Do not add `core/database/queries/`.
- Domain validation sentinels stay in this package (`users.ErrInvalid`). Do **not** add them to `core/database/errors.go` (that file is lock/format/folder/closed only).
- `database` must **not** import table packages (import cycle). Use-cases (`onboarding`, later FFI) import both.

## Tests

Table-driven, in the same folder. `database.Create` a temp `*.provenencia`, `defer Close()`, cover happy path, validation, and closed catalog.

Run `CGO_ENABLED=1 go test ./core/database/...` (and the use-case package you called from).

## Do not

- Add `<table>.go` / `<table>_test.go` under `core/database/` itself
- Attach table methods to `Catalog`
- Put identity, FFI, or ingest in the query package
- Reach into `Catalog`’s unexported `db` field from another package
