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
| `core/database/sourcevocab/` | `registry.go` (`seedTypes` / `seedFields` / `seedSuggestions`) | Source types, metadata fields, type→field suggestions (uses `sourcetypes` / `metadatafields`) |
| `core/database/sourcecredibilitygrades/` | `registry.go` (`seedGrades`) | Credibility grades (`low_trust` / `standard` / `high_trust`) |
| `core/database/claimconfidencegrades/` | `registry.go` (`seedGrades`) | Claim confidence grades (`low_confidence` / `moderate` / `high_confidence`) for Identity (later Reconciliation) Claims |
| `core/connectrules/` | `productBridges` / `productRefusals` (`Bridge.Endpoints` + `Disambiguation` + `Pairs`); `All()` / `Register` | Product connect matrix: which Properties on which bridge types are endpoints or disambiguation. Impact, edge-lock, Connect, and Install bridge bindings loop `All()` / `BridgeBindings()`. Future plugins append via `Register`. |
| `core/database/subjectvocab/` | `registry.go` (`seedTypes` / `seedProperties` / `seedBindings` / `seedTerms`) | Interpretation Subject types, Properties, non-bridge `subject_type_properties`, **`property_terms`** (S7-01b), plus compiled presentation. Bridge bindings are derived from `connectrules.BridgeBindings()` (product origin only at create). |
| `core/database/nameformats/` | `registry.go` (`seedProfiles`) | Name format profiles: part order, roles, weights, and which profile is the project default. Install writes the list. Load and Format do not branch on the profile key. |
| `core/database/namealiases/` | `registry.go` (`productAliasTables`, with `nicknames_seed.go` and `third_party/nicknames/`); `Register` appends | Abbreviation and nickname tables. Install writes the registered list. Matching reads the catalog rows; the reconciler cache reads the product seed, so a seed edit bumps `CacheVersion`. Each product set is a head and the names that stand for it (`FormalFirst`); `nicknameDirections` says which name a fold keeps where the list pairs two names both ways (S11-28). An empty catalog makes no alias hits. Opening a project stores that read on the catalog graph until close. `Invalidate` reloads it. Origin does not filter scoring. |

### Source vocabulary (`sourcevocab`)

1. Edit **`core/database/sourcevocab/registry.go`** — declarative lists:
   - `seedTypes` — `key`, `label`, `description`
   - `seedFields` — `key`, `label`, `data_type` (`text`|`url`), optional description. Date-named keys (`record_date`, …) are `text`.
   - `seedSuggestions` — `(type_key, field_key, sort_order)`
2. Prefer a tiny starter set. New projects get whatever is in the registry at
   create time; existing catalogs are not backfilled on open.
3. All registry rows are **`origin = provenencia`**. Researcher/plugin rows use
   Upsert APIs with `user` / `plugin:<id>` — never put those in the registry.
   Origins, primary kinds and the Property keys code reads by name come from
   `core/vocabkey`. A vocabulary package does not declare its own
   `OriginProvenencia` or origin check; it calls `vocabkey.ValidOrigin`.
4. **No seed `INSERT`s** in `migrations/*.sql`. DDL-only migrations for new tables.
5. Adding rows to the registry is enough for normal growth of the create-time
   starter. Do not reintroduce open-time heal.
6. Tests in `sourcevocab_test.go`: empty → trimmed counts; user twin left alone;
   deleted rows stay deleted without re-Install. Onboarding: open does not heal.
7. Run `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcevocab/... ./core/onboarding/...`.

### Subject vocabulary / credibility grades

Same pattern as Source: edit that package’s `registry.go`, keep
`origin = provenencia`, no SQL seeds, tests that create installs and open does
not heal. For `subjectvocab`, `TestInstallMatchesRegistry` already checks every seeded Property, term and binding against the registry, so a new row needs no Install test: add its product values (labels, value type, cardinality, sort order, lock) as a case in `TestSeedProductValues`. Tests in other packages read the registry (`BindingsForType`, `len(seedProperties)`) instead of restating seed lists or counts. `core/claimevidence`'s `seedOrder` is a matching fixture, not a copy of the seed, and stays as it is. Subject type keys/prefixes: `docs/seeded-vocabulary.md` §3.1.
Property value types: `text` \| `integer` \| `date` \| `name` \| `subject` \| **`term`** \| **`depiction`**.
Kind/edge Properties (`event_type`, `role`, `relationship_type`) use `term` + `property_terms` (S7-01b) — not free-text Observation strings.
**`value_type = term` and `depiction` Properties are Install/registry only** (`origin=provenencia` or `plugin:<id>`). Researcher Create Property must refuse `term` and `depiction`.
Researchers may still add `origin=user` **term rows** under those Properties via the composer picker.
Capabilities and presentation tokens stay in the compiled `subjectvocab`
registry (not SQL columns). Non-bridge locked bindings stay in `seedBindings`.
The connect matrix is **`core/connectrules` only** — author a `Bridge`
(Endpoints, Disambiguation property key, Pairs) and Impact / edge-lock /
Connect / Install bridge bindings pick it up. Add new property rows in
`seedProperties` / `seedTerms` when the property is new. Do not hand-list
bridge rows in `seedBindings`. `Register` is the extension seam for future
`plugin:<id>` modules (no policy `case` changes). The `connectrules` leaf
also lets `deleteimpact` classify connection facets without importing
`subjectvocab`. `CatalogConnectRule.productMatrix` in
`macos/App/Platform/GenealogyStore.swift` is the FakeStore / unit-test double —
update it in the same change as product bridges. Live Mac connect reads
`listConnectRules` and does not fall back to that table. Term capabilities
(birthday facets, tree-edge roles, …) are deferred until a later PR.

A new subject type also adds a `subject_<key>` case to `PVMarkKey`, its `provenencia_subject_<key>` asset, and its `accessibilityName` with an `L10n.DesignSystem` string. `TestSwiftVocabularyMatchesGo` fails until the case exists.

**A seeded term the Mac app names has a Swift step too.** For a term of `gender`, `sex_at_birth`, `role`, `event_type`, `relationship_type` or `place_relationship_type`, add its `Seeded…Term` constant in `macos/App/Platform/ConclusionVocabulary.swift` and its entry in `L10n.PropertyTerm.wording` (`macos/App/Platform/L10n.swift`): the term's label, and for a role, event type or relationship type its share phrase and similar phrase, each whole, and for a relationship type that an implied link can carry its `recordedDirect` phrase ("recorded as spouses", which the evidence sheet's caption uses), with the catalog keys (`add-localized-string`). `TestSwiftVocabularyMatchesGo` fails, naming the term, when the seed has a term with no constant or a constant has no wording. A new term-bearing Property whose terms the app names needs the same two steps and an entry in that test's list. One the app only shows by catalog label (`religion`, `languages_spoken`) needs neither.

Uniqueness is **`UNIQUE (key, origin)`**. Lookup is always `(key, origin)`, never bare key. Domain FKs store vocabulary **`id`**, not key.

### Install semantics (pinned)

- Each domain’s `Install` upserts its registry into a new catalog.
- Call **only** from `onboarding.createCatalog` (after `database.Create`). Today that is:
  `sourcevocab.Install` → `sourcecredibilitygrades.Install` → `claimconfidencegrades.Install` → `subjectvocab.Install` → `nameformats.Install` → `namealiases.Install`
  (then `catalogsession.ReconcileOnOpen`).
- **Do not** call `Install` from `OpenCatalog`, `catalogsession.Do`, or on every open.
- **Do not** heal deleted `provenencia` rows or restored suggestion joins on open.
- Calling `Install` twice would refresh labels via Upsert — create path calls each once.
- Unused-vocab delete goes through `deleteimpact` after that page’s PR (**S8-15** types, **S8-16** fields, **S8-17** properties) — not a private `sqlInUse`. **Origin:** plugin never; seeded types/fields may erase when unused; seeded properties/terms stay locked. See [`add-catalog-delete`](../add-catalog-delete/SKILL.md).

## Where create vs open run

`database.Create` / `database.Open` stay **migrate-only** (tests, low-level).
Researcher-facing paths:

```
core/onboarding/ready.go
  createCatalog → database.Create + sourcevocab.Install
                  + sourcecredibilitygrades.Install
                  + claimconfidencegrades.Install
                  + subjectvocab.Install + nameformats.Install
                  + namealiases.Install
                  + catalogsession.ReconcileOnOpen
  OpenCatalog   → database.Open + catalogsession.ReconcileOnOpen   # one-shot / tests

core/catalogsession
  ReconcileOnOpen(c) → users.EnsureRefs, evidencegraphs.EnsureRefs,
                       project.EnsureUUID, autoreconciler.EnsureCatalog,
                       searchindex.EnsureCatalog
  Do(projectDir, fn) → open once like OpenCatalog (ReconcileOnOpen), hold + serialize
```

Complete uses **`createCatalog`** (create-then-close). Open / ProjectInfo /
ListContributors and catalog FFI handlers use **`catalogsession.Do`** (or
`withProjectCatalog` in handlers), not raw `database.Create`/`Open` and not
open-per-call `OpenCatalog`. See `.agents/skills/use-catalog-session/SKILL.md`.

Do **not** scatter `*.Install` at each use-case.

When adding another seed domain’s create-time install, call it from
**`createCatalog` only** (not open / not `Do`), unless that domain explicitly needs
open-time policy of its own. Mirror an existing domain (`subjectvocab` or
`sourcecredibilitygrades`) rather than inventing a generic multi-domain framework.

## Future vocabulary domains

Mirror Source / Subject types — do **not** build a generic multi-domain seed framework:

1. Migration for definition tables (`origin`, `UNIQUE (key, origin)`). Follow `.agents/skills/add-catalog-migration/SKILL.md`.
2. Nested query package(s) under `core/database/<domain>/`. Follow `.agents/skills/add-catalog-query/SKILL.md`.
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
- Write `"provenencia"`, `"plugin:"`, a primary kind or a `vocabkey` Property key as a Go string literal outside the seed (SQL text keeps its literals)
