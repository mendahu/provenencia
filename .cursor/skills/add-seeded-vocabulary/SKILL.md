---
name: add-seeded-vocabulary
description: >-
  Extends Provenencia product-seeded catalog vocabulary (origin=provenencia)
  via declarative registries and create-time Install. Use when adding or changing
  source types, source metadata fields, type→field suggestions, sourcevocab
  registry/Install, subject types, source credibility grades, seeding later
  Interpretation domains, or when the user mentions seeded vocabulary, dogfood
  seeds, or create-time seed.
---

# Add seeded vocabulary

Shipped taxonomy is **data** installed once at catalog create, not SQL enums
and not migration `INSERT`s. Authoritative key catalog (horizon intent):
[`docs/seeded-vocabulary.md`](../../../docs/seeded-vocabulary.md) §1.1.
Source schema: [`docs/source-layer-data-model.md`](../../../docs/source-layer-data-model.md).

## Create-time seed domains (today)

Each domain owns a declarative `registry.go` + `Install(c) error`. Wire every
new Install into **`onboarding.createCatalog` only**.

| Package | Registry | What it seeds |
| --- | --- | --- |
| `core/database/sourcevocab/` | `registry.go` (`seedTypes` / `seedFields` / `seedSuggestions`) | Source types, metadata fields, type→field suggestions (uses `sourcetypes` / `sourcefields`) |
| `core/database/sourcecredibilitygrades/` | `registry.go` (`seedGrades`) | Credibility grades (`low_trust` / `standard` / `high_trust`) |
| `core/database/subjectvocab/` | `registry.go` (`seedTypes` / `seedProperties` / `seedBindings` / `seedTerms` / `seedConnect`) | Interpretation Subject types, Properties, `subject_type_fields`, **`property_terms`** (S7-01b), plus compiled presentation / locked bindings / connect (uses `subjecttypes` / `properties`) |

### Source vocabulary (`sourcevocab`)

1. Edit **`core/database/sourcevocab/registry.go`** — declarative lists:
   - `seedTypes` — `key`, `label`, `description`
   - `seedFields` — `key`, `label`, `data_type` (`text`|`date`), optional description
   - `seedSuggestions` — `(type_key, field_key, sort_order)`
2. Prefer a tiny starter set. New projects get whatever is in the registry at
   create time; existing catalogs are not backfilled on open.
3. All registry rows are **`origin = provenencia`**. Researcher/plugin rows use
   Upsert APIs with `user` / `plugin:<id>` — never put those in the registry.
4. **No seed `INSERT`s** in `migrations/*.sql`. DDL-only migrations for new tables.
5. Adding rows to the registry is enough for normal growth of the create-time
   starter. Do not reintroduce open-time heal.
6. Tests in `sourcevocab_test.go`: empty → trimmed counts; user twin left alone;
   deleted rows stay deleted without re-Install. Onboarding: open does not heal.
7. Run `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcevocab/... ./core/onboarding/...`.

### Subject vocabulary / credibility grades

Same pattern as Source: edit that package’s `registry.go`, keep
`origin = provenencia`, no SQL seeds, tests that create installs and open does
not heal. Subject type keys/prefixes: `docs/seeded-vocabulary.md` §3.1.
Property value types: `text` \| `integer` \| `date` \| `name` \| `subject` \| **`term`**.
Kind/edge Properties (`event_type`, `role`, `relationship_type`) use `term` + `property_terms` (S7-01b) — not free-text Observation strings.
**`value_type = term` Properties are Install/registry only** (`origin=provenencia` or `plugin:<id>`). Researcher Create Property must refuse `term`.
Researchers may still add `origin=user` **term rows** under those Properties via the composer picker.
Capabilities, presentation tokens, locked bindings, and the connect matrix stay
in the compiled `subjectvocab` registry (not SQL columns). Term capabilities
(birthday facets, tree-edge roles, …) are deferred until a later PR.

Uniqueness is **`UNIQUE (key, origin)`**. Lookup is always `(key, origin)`, never bare key. Domain FKs store vocabulary **`id`**, not key.

### Install semantics (pinned)

- Each domain’s `Install` upserts its registry into a new catalog.
- Call **only** from `onboarding.createCatalog` (after `database.Create` +
  `users.EnsureRefs` / `project.EnsureUUID`). Today that is:
  `sourcevocab.Install` → `sourcecredibilitygrades.Install` → `subjectvocab.Install`
  (then `searchindex.EnsureCatalog`).
- **Do not** call `Install` from `OpenCatalog`, `catalogsession.Do`, or on every open.
- **Do not** heal deleted `provenencia` rows or restored suggestion joins on open.
- Calling `Install` twice would refresh labels via Upsert — create path calls each once.
- Types/fields of any origin may be deleted when unused (`ErrInUse` while referenced).

## Where create vs open run

`database.Create` / `database.Open` stay **migrate-only** (tests, low-level).
Researcher-facing paths:

```
core/onboarding/ready.go
  createCatalog → database.Create + users.EnsureRefs + project.EnsureUUID
                  + sourcevocab.Install + sourcecredibilitygrades.Install
                  + subjectvocab.Install + searchindex.EnsureCatalog
  OpenCatalog   → database.Open + users.EnsureRefs   # one-shot / tests

core/catalogsession
  Do(projectDir, fn) → open once like OpenCatalog, hold + serialize
```

Complete uses **`createCatalog`** (create-then-close). Open / ProjectInfo /
ListContributors and catalog FFI handlers use **`catalogsession.Do`** (or
`withProjectCatalog` in handlers), not raw `database.Create`/`Open` and not
open-per-call `OpenCatalog`. See `.cursor/skills/use-catalog-session/SKILL.md`.

Do **not** scatter `*.Install` at each use-case.

When adding another seed domain’s create-time install, call it from
**`createCatalog` only** (not open / not `Do`), unless that domain explicitly needs
open-time policy of its own. Mirror an existing domain (`subjectvocab` or
`sourcecredibilitygrades`) rather than inventing a generic multi-domain framework.

## Future vocabulary domains

Mirror Source / Subject types — do **not** build a generic multi-domain seed framework:

1. Migration for definition tables (`origin`, `UNIQUE (key, origin)`). Follow `.cursor/skills/add-catalog-migration/SKILL.md`.
2. Nested query package(s) under `core/database/<domain>/`. Follow `.cursor/skills/add-catalog-query/SKILL.md`.
3. Package owning the seed: `registry` lists + `Install(c) error` (or domain-specific name).
4. Wire install into `createCatalog` in `ready.go` when create-time-only is correct for that domain.
5. Document intended keys in `docs/seeded-vocabulary.md`; implement only the dogfood slice needed now.
   Prefer mirroring `subjectvocab` (or `sourcevocab`) rather than inventing a generic multi-domain framework.

Join/suggestion tables have **no `origin`** column.

## Do not

- Seed via SQL migrations (unless a later explicit backfill policy says otherwise)
- Call domain `Install` from FFI handlers, Swift, `OpenCatalog`, or `catalogsession.Do`
- Put seed install inside `database.Open`/`Create` (import cycle; couples migrate to product policy)
- Treat `provenencia` as a `builtin` boolean — use `origin`
- Expand the full horizon catalog in one PR “just because” it is listed in the docs
- Reintroduce open-time Ensure/heal for seeded vocabulary
