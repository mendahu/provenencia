# S11-44 — Des as a nickname of Desiree and Desirae

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the command is a terminal command, user 2026-10-27; matching and the Person page already show what it adds)
**Board:** —
**Depends on:** S11-28 (landed [#589](https://github.com/mendahu/provenencia/pull/589)), S11-48 (landed [#563](https://github.com/mendahu/provenencia/pull/563))
**Blocks:** —
**Branch:** `s11-44-des-nickname`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [Case 2](../name-candidates.md#case-2--desiree-against-des) and decision 3 in [`name-candidates.md`](../name-candidates.md#decided-user-2026-10-24); [`docs/seeded-vocabulary.md`](../../../seeded-vocabulary.md) §4.5; [S11-28 plan](S11-28-alias-fold.md) (Landed note); [S11-48 plan](S11-48-religion-languages.md) D-7; working rules: "Des is a nickname of Desiree and of Desirae", "High tolerance for churn"
**Planned:** 2026-10-27 against `main` @ `5312337`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A Source person named "Des" is offered the record "Desiree" (or "Desirae") in Promote's target menu, and the other way round, in new projects and in the researcher's existing project. Two nickname sets, `{desiree, des}` and `{desirae, des}`, join the product nickname table. A terminal command copies product vocabulary an existing catalog lacks into it: the new nickname sets, and S11-48's Religion and Languages spoken Properties with their terms and person bindings.

**Why now.** Dogfood on 2026-10-24: "Desiree" was not offered "Des" ([case 2](../name-candidates.md#case-2--desiree-against-des)). Alias rows are catalog data written once at create, so a seed edit alone reaches only new catalogs (decision 3, user 2026-10-24). S11-48 left its two Properties for this command (user 2026-10-25). The user asked on 2026-10-27 for the command to be one they paste into a terminal, not an app menu item.

**What stays.** Nothing runs on open; `add-seeded-vocabulary`'s "Do not heal" rule stands. Install and `onboarding.createCatalog` are unchanged. The Mac app, the FFI and L10n are unchanged. A seeded row the catalog already has is never rewritten, so a label the researcher edited keeps the edit, and a binding the researcher removed stays removed.

## 2. User stories

**US-1.** As a researcher, when a Source names "Des" and my tree has "Desiree", Promote offers Desiree as a target, so that a nickname record finds its formal-name person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A new catalog; `namealiases.Load` + `valuerules.Load`; compare given "des" with given "desiree", and "des" with "desirae" | Both `Equivalent` with similarity `matchregistry` `AliasCredit` (the same score `TestAddProductNicknamesGrowsTheShortList` checks for an alias pair today) |
| AC-1.2 | Same catalog; compare "desiree" with "desirae" | Not an alias hit: the two sets share only `des` and stay separate (the table's rule, `TestSharedNicknameDoesNotMergeCanonicalNames`) |
| AC-1.3 | The reconciler's rules (`autoreconciler.ValueRules`): given "Des" and "Desiree" (surname Robins) in one entity; then "Des" and "Desirae"; then "Desiree" and "Desirae" | One value, given "Desiree"; one value, given "Desirae"; both given names kept (S11-28's fold keeps the head of the set that holds the other) |

**US-2.** As a researcher with a project made before this PR, I run one command in a terminal with the project closed, so that the project gains the product vocabulary added since it was made.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A catalog with the full seed minus the two Des sets, `religion`, `languages_spoken`, their terms and their person bindings; run the command | Adds 2 nickname sets (4 tokens), 2 Properties, 2 terms, 2 bindings; prints `Added 2 nickname sets (4 names), 2 properties, 2 terms and 2 bindings.`; exit 0 |
| AC-2.2 | Same catalog after AC-2.1; run the command again | Adds nothing; prints `Already up to date.`; exit 0 |
| AC-2.3 | After AC-2.1 | The catalog matches the subject vocabulary registry as `TestInstallMatchesRegistry` checks it; person's last two bindings are `religion` then `languages_spoken`, after every binding the catalog already had |
| AC-2.4 | A catalog where the researcher relabelled seeded Property `interests` to "Hobbies" and removed person's `interests` binding; run the command | `interests` keeps label "Hobbies"; person has no `interests` binding |
| AC-2.5 | The project is open in Provenencia; run the command | Prints `The project is open in Provenencia. Close it, then run this again.` to stderr; exit 1; the catalog is unchanged |
| AC-2.6 | No argument, or a path with no catalog | Prints usage (`usage: add-product-vocabulary <project folder>`) or the open error to stderr; exit 2 for usage, 1 for the error |

## 3. Scope

| In | Covered by |
| --- | --- |
| `{desiree, des}` and `{desirae, des}` in `nicknameSets`; `CacheVersion` 19 with its changelog line; `productAliasPin` re-pinned; set count 1229 | STEP-1, AC-1.1 – AC-1.3 |
| `subjectvocab.AddMissing` *(new)*: Properties and terms missing by `(key, origin)`, and a binding only for a Property it adds, in one transaction | STEP-3, AC-2.1, AC-2.3, AC-2.4 |
| `properties.UpsertTx`, `propertyterms.UpsertTx`, `propertyterms.LookupTx` *(new)* so `AddMissing` runs in one transaction | STEP-2 |
| `onboarding.AddProductVocabulary` *(new)*: open, `AddMissing`, `AddProductNicknames`, close | STEP-4, AC-2.1, AC-2.2, AC-2.5 |
| `api/cmd/add-product-vocabulary` *(new)*: the terminal command | STEP-5, AC-2.1, AC-2.2, AC-2.5, AC-2.6 |
| Docs: `seeded-vocabulary.md` §4.5 and the subject vocabulary section, `add-seeded-vocabulary` skill, map scope entry and row | STEP-6, STEP-7 |

| Out | Home |
| --- | --- |
| An app menu command, FFI handler, toast and L10n copy | not planned (user 2026-10-27: a terminal command) |
| Removing seed pairs a catalog still has (the 31 pairs [#499](https://github.com/mendahu/provenencia/pull/499) dropped; audit L-4) | not planned; the command only adds (`docs/deployment-plan/spike-12/app-health-audit.md` L-4) |
| Missing seeded Subject types, Source vocabulary, credibility and confidence grades, name formats, abbreviation sets | not planned: no seed has added one since any dogfood catalog was made. The next PR that seeds one adds it to the command |
| Prefix names ("Des" for "Desmond" with no set) | S11-46, S11-45 |
| An editor for alias tables | not planned |
| Updating a seeded row's label or description in an existing catalog | not planned (a researcher's edit wins) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Nickname seed | `core/database/namealiases/nicknames_seed.go` `nicknameSets` | 1227 sets, head first, sorted by head. No `des`, `desiree` or `desirae` token anywhere (`rg '"des"\|"desiree"\|"desirae"' core/database/namealiases` → no hit). `{"dennison", "dennis", "denny"}` is followed by `{"derek", …}` |
| Set count guard | `core/database/namealiases/registry_test.go` `TestNicknameDirectionsCoverTwoWayPairs` | Asserts `len(nicknameSets) == 1227`, 108 directions, 143 two-way pairs |
| Cache version | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` | 18, changelog line "18: a nickname or abbreviation … (S11-28)" |
| Seed pin | `core/database/autoreconciler/autoreconciler_test.go` `productAliasPin`, `TestProductAliasTablesPinCacheVersion` | `{CacheVersion: 18, Hash: "dd2d088826d1…"}` (truncated here); the failure message prints the new hash |
| Copy into an existing catalog | `core/database/namealiases/add.go` `AddProductNicknames`, `NicknameAdd` | Copies product nickname sets the catalog lacks, in one transaction; a set that could join two stored sets is added beside them. No caller outside tests (`TestAddProductNicknamesFillsAnEmptyCatalog`, `TestAddProductNicknamesGrowsTheShortList`) |
| Subject vocabulary seed | `core/database/subjectvocab/subjectvocab.go` `Install`; `registry.go` `seedProperties`, `seedTerms`, `seedBindings`; `allBindings` | `Install` upserts every type, Property, term and binding with `*database.Catalog` calls (no transaction). Upsert overwrites label and description, so Install cannot be re-run on an existing catalog. `religion` / `languages_spoken` bind to person at `SortOrder` 5 / 6 |
| Binding append | `subjectvocab.go` `sqlAppendBinding` (used by `AppendBinding`) | Inserts at `MAX(sort_order) + 1` of the type, `ON CONFLICT DO NOTHING` |
| Property and term upsert | `core/database/properties/properties.go` `Upsert`, `LookupTx`; `core/database/propertyterms/propertyterms.go` `Upsert`, `Lookup` | `Upsert` takes `*database.Catalog` and runs `db.Exec`; `properties.LookupTx(*sql.Tx, key, origin)` exists; `propertyterms` has no `LookupTx` |
| One connection | `core/database/catalog.go` `openDB` | `db.SetMaxOpenConns(1)`: a `*Catalog` call made while a `*sql.Tx` is open blocks for good |
| Open | `core/database/catalog.go` `Open`, `takeExclusiveLock`, `mapLockErr`; `core/database/errors.go` `ErrAlreadyOpen` | Migrates, then takes an exclusive lock; a project held by the app returns `ErrAlreadyOpen` |
| Seeded rows are delete-locked | `core/database/deleteimpact/register.go` (`KindProperty`, `KindPropertyTerm`: `SeededLocked`) | A seeded Property or term cannot be deleted, so one missing by `(key, origin)` was never in the catalog. A binding can be removed (`DeleteBinding` refuses only registry-locked ones) |
| Create path | `core/onboarding/ready.go` `createCatalog`, `OpenCatalog` | Install runs at create only; `OpenCatalog` heals nothing (`TestOpenCatalogDoesNotHeal…` in `ready_test.go`) |
| Binaries | `api/libprovenencia/main.go` | The only `package main`; `scripts/check-pr.sh` vets and tests `./api/... ./core/...` |
| Registry check | `core/database/subjectvocab/subjectvocab_test.go` `TestInstallMatchesRegistry` | Creates, Installs, then checks each seeded Property, term, type and binding against the registry inline |

**Precedent.** S11-28 ([#589](https://github.com/mendahu/provenencia/pull/589)) for the seed edit: a `nicknameSets` change plus a `CacheVersion` line and a re-pinned `productAliasPin`. Copy its changelog style. `AddProductNicknames` (`add.go`) for "copy what is missing, never rewrite, one transaction"; `AddMissing` follows its shape and its count struct. S11-48 D-7 promised this command a "catalog matches the registry" check: reuse `TestInstallMatchesRegistry`'s body (D-8).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The two sets are `{"desirae", "des"}` and `{"desiree", "des"}`, inserted after `{"derrick", …}` and before `{"deuteronomy", …}` (sorted by head) | The map's call (user 2026-10-24); head first so S11-28's fold keeps the formal name | One set `{desiree, desirae, des}` — joins two formal names |
| D-2 | `CacheVersion` 19, line `19: Des is a nickname of Desiree and of Desirae (S11-44).`; `productAliasPin` `{CacheVersion: 19, Hash: <the hash the failing test prints>}`; `TestNicknameDirectionsCoverTwoWayPairs` expects 1229 sets (directions and two-way counts unchanged: neither pair is two-way) | S11-28's recipe; the pin fails until both move | — |
| D-3 | The command is a Go binary at `api/cmd/add-product-vocabulary/main.go` *(new)*, run from a checkout: `CGO_ENABLED=1 go run -tags fts5 ./api/cmd/add-product-vocabulary "<project folder>"`. `main` only parses one argument, calls `onboarding.AddProductVocabulary`, prints and sets the exit code | User 2026-10-27: a command pasted into a terminal. A committed binary is tested, vetted by `check-pr.sh` (`./api/...`), and reused when the next seed grows. `api/` already holds the repo's one binary | An app menu item / FFI handler (user declined); a `sqlite3` script (duplicates the seed, untested, breaks `UNIQUE` ids); `core/cmd` (`core` holds no binaries) |
| D-4 | The command runs only on a closed project: `database.Open` returns `ErrAlreadyOpen` while the app holds it, and the command prints AC-2.5's line. It does not call `namealiases.Invalidate` (no graph is warm in its process) and does not run `catalogsession.ReconcileOnOpen` (the app does on the next open, and the `CacheVersion` bump rebuilds the cache there) | The exclusive lock already serializes with the app; nothing in the command's process is cached | Running through `catalogsession` from the app (no app entry point, D-3); `onboarding.OpenCatalog` (runs the reconciler rebuild in the command for no gain) |
| D-5 | `onboarding.AddProductVocabulary(projectDir string) (VocabularyAdd, error)` *(new)* in `core/onboarding/addvocab.go`: `database.Open`, then `subjectvocab.AddMissing`, then `namealiases.AddProductNicknames`, then `Close`. `VocabularyAdd{Subject subjectvocab.Added; Nicknames namealiases.NicknameAdd}` with `Empty() bool` (every count zero) | `onboarding` owns which seeds a catalog gets (`createCatalog`); the command and its test share one function | Wiring the three calls in `main` (untested); a new package (one function) |
| D-6 | `subjectvocab.AddMissing(c *database.Catalog) (Added, error)` *(new)*, `Added{Properties, Terms, Bindings int}` *(new)*, in `core/database/subjectvocab/add.go` *(new)*, one `*sql.Tx`: (1) each `seedProperties` row with no `(key, provenencia)` row is inserted (`properties.UpsertTx`) and remembered as added; (2) each `seedTerms` row whose Property exists (looked up by key, provenencia) and has no `(property, key, provenencia)` term is inserted (`propertyterms.UpsertTx`); (3) each `allBindings()` row whose `PropertyKey` was added in (1) and whose type exists (`subjecttypes.LookupTx`) is appended with `sqlAppendBinding`, in `allBindings` order. A binding whose type is missing is skipped (not counted). Rows present are never written | Missing by `(key, origin)` is the map's rule (user 2026-10-25); seeded Properties and terms are delete-locked, so "missing" means "new since create". Binding only an added Property keeps a researcher's removal (map). Appending keeps the researcher's order and puts new Properties last | Re-running `Install` (rewrites labels and re-sorts bindings); registry `SortOrder` (collides with a reordered list); restoring every missing binding (undoes a researcher's removal) |
| D-7 | `properties.UpsertTx(tx *sql.Tx, p Property) ([]byte, error)` and `propertyterms.UpsertTx(tx *sql.Tx, t Term) ([]byte, error)` *(new)*; `propertyterms.LookupTx(tx *sql.Tx, propertyID []byte, key, origin string) (Term, error)` *(new)*. Each package's `Upsert` and `Lookup` become a call to one unexported helper that takes a querier (`interface{ Exec; QueryRow }` local to the package), so the trimming, validation and SQL exist once | `SetMaxOpenConns(1)`: a `*Catalog` call inside the transaction would block. One transaction makes a failed run leave nothing, so a rerun still binds what it adds | Calling the `*Catalog` functions with no transaction (a failure between a Property and its binding leaves a Property the rerun never binds); copying the upsert SQL into `subjectvocab` (two copies) |
| D-8 | `subjectvocab_test.go` gains `assertMatchesRegistry(t, c)` *(new)*, the body of `TestInstallMatchesRegistry` after its Install; `TestInstallMatchesRegistry` calls it, and `TestAddMissing` *(new)* calls it after `AddMissing` | S11-48 D-7 (review M-1): one "catalog matches the registry" check for both | A second hand-written check |
| D-9 | No audit row. `AddMissing` and `AddProductNicknames` write seed rows the way `Install` does | Seed rows are product data with no user behind them; `Install` and `AddProductNicknames` are not audited | `writes.Run` with a system user (no such user; "last activity" would show a seed copy) |
| D-10 | Output, exact: on any addition `Added %d nickname sets (%d names), %d properties, %d terms and %d bindings.` (every count printed, zeros included) to stdout; when `Empty()` `Already up to date.`; `ErrAlreadyOpen` (`errors.Is`) → stderr `The project is open in Provenencia. Close it, then run this again.`, exit 1; any other error → stderr `add-product-vocabulary: <err>`, exit 1; argument count ≠ 1 → stderr `usage: add-product-vocabulary <project folder>`, exit 2. English only | A developer command, not app copy: no L10n (the app's strings rule covers `macos/App` only) | Localized output (no Swift side) |
| D-11 | The command migrates the catalog it opens (`database.Open`), so it is run from the same commit the installed app was built from. §10 and the `seeded-vocabulary.md` paragraph say so | A newer checkout can migrate the catalog past the installed app's schema (`catalog.schema_mismatch`) | Opening without migrating (no such entry point) |

**Rules this plan bends.** None. `add-seeded-vocabulary` says "Do not call `Install` from `OpenCatalog`" and "Do not heal deleted `provenencia` rows or restored suggestion joins on open": the command calls neither on open and is run by the researcher (decision 3 in `name-candidates.md`, user 2026-10-24). STEP-6 adds the command to the skill.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `subjectvocab.AddMissing`, `Added` *(new)* | func, type | `subjectvocab.Install`; `namealiases.AddProductNicknames`, `NicknameAdd` | `Install` overwrites labels and binding order (D-6). Same shape as `AddProductNicknames` for another seed |
| `core/database/subjectvocab/add.go` *(new)* | file | `core/database/namealiases/add.go` | Mirrors it |
| `properties.UpsertTx`, `propertyterms.UpsertTx`, `propertyterms.LookupTx` *(new)* | func | `properties.Upsert`, `properties.LookupTx`, `propertyterms.Upsert`, `propertyterms.Lookup` | Same body through one helper (D-7); `properties.LookupTx` is the existing precedent |
| `onboarding.AddProductVocabulary`, `VocabularyAdd` *(new)*, `core/onboarding/addvocab.go` *(new)* | func, type, file | `onboarding.createCatalog` | Create installs into a new folder; this opens an existing one and adds only what is missing |
| `api/cmd/add-product-vocabulary/main.go` *(new)* | binary | `api/libprovenencia/main.go` | That is the FFI library; this is a terminal command (D-3) |
| `assertMatchesRegistry` *(new)* | test helper | `TestInstallMatchesRegistry` body | Extracted, not copied (D-8) |
| `TestAddMissing`, `TestAddProductVocabulary` *(new)* | tests | `TestAddProductNicknamesGrowsTheShortList` | New behavior |

**Builds on.** `add-seeded-vocabulary` (registries are the one list), `AddProductNicknames` (S9 era, no caller until now), S11-28's seed pin.
**Deletes.** None — nothing becomes dead. `AddProductNicknames` gains its first caller.
**Next adds.** A later seeded Property or term reaches old catalogs by its registry row alone; a later nickname set by its seed line plus the S11-28 pin. A seeded Subject type, Source vocabulary row or grade is one more `AddMissing`-style function called from `AddProductVocabulary`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A catalog whose `name_alias_members` grew after create | `namealiases.Load`, `NewAliasIndex` at the next open | `Load` reads whatever rows exist; `TestAddProductNicknamesGrowsTheShortList` already loads a grown catalog |
| A Property bound to person after other Properties the researcher ordered | Properties page, Citation composer, Person page (`ListBindings` by `sort_order`) | `sqlAppendBinding` is how the Properties page's own Add binds; the order is the same state |
| A seeded type missing from the catalog | `AddMissing` step (3) | Skips the binding (D-6) |

### Scale

At dogfood size (one catalog, 1229 product sets, about 30 seeded Properties): `AddProductNicknames` reads every stored set once and is linear in sets × product sets it compares by subset (as today, untested at scale beyond `TestLoadHugeSetIndex`); `AddMissing` does one lookup per registry row. The command runs once by hand; no cache in its process.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/namealiases` | `nicknames_seed.go`: D-1's two lines. `registry_test.go`: 1229 (D-2) |
| `core/database/autoreconciler` | `CacheVersion = 19` and D-2's line; `autoreconciler_test.go` `productAliasPin` re-pinned |
| `core/database/properties` | `func UpsertTx(tx *sql.Tx, p Property) ([]byte, error)` *(new)*; `Upsert` and `UpsertTx` share `upsert(q querier, lookup func(key, origin string) (Property, error), p Property)` or an equivalent single helper (D-7) |
| `core/database/propertyterms` | `func UpsertTx(tx *sql.Tx, t Term) ([]byte, error)` *(new)*, `func LookupTx(tx *sql.Tx, propertyID []byte, key, origin string) (Term, error)` *(new)*, sharing helpers with `Upsert` / `Lookup` |
| `core/database/subjectvocab` | `add.go` *(new)*: `type Added struct{ Properties, Terms, Bindings int }`, `func AddMissing(c *database.Catalog) (Added, error)` (D-6) |
| `core/onboarding` | `addvocab.go` *(new)*: `type VocabularyAdd struct{ Subject subjectvocab.Added; Nicknames namealiases.NicknameAdd }`, `func (a VocabularyAdd) Empty() bool`, `func AddProductVocabulary(projectDir string) (VocabularyAdd, error)` (D-5) |
| `api/cmd/add-product-vocabulary` | `main.go` *(new)* (D-3, D-10) |

Errors: `database.ErrAlreadyOpen` (`catalog.already_open`) → D-10's message. No new error code.

### FFI

None — the command is not reached through the app (D-3).

### macOS

None — no Swift change.

### Strings

None — no L10n key; the command's English output is D-10.

## 7. Steps

**STEP-1 — Des sets in the seed**
Files: `core/database/namealiases/nicknames_seed.go`, `core/database/namealiases/registry_test.go`, `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/namealiases/namealiases_test.go`. Skill: `add-match-rule`, `add-name-value`.
Do: D-1, D-2. Run T-7 once, copy the hash it prints into `productAliasPin`. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/namealiases/... ./core/database/autoreconciler/... ./core/autoreconcile/...` passes.

**STEP-2 — Transaction variants of the vocabulary upserts**
Files: `core/database/properties/properties.go`, `core/database/propertyterms/propertyterms.go`. Skill: `add-catalog-query`.
Do: D-7. `Upsert` / `Lookup` keep their signatures and behavior.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/properties/... ./core/database/propertyterms/... ./core/database/subjectvocab/...` passes unchanged.

**STEP-3 — `subjectvocab.AddMissing`**
Files: `core/database/subjectvocab/add.go` *(new)*, `core/database/subjectvocab/subjectvocab_test.go`. Skill: `add-seeded-vocabulary`.
Do: D-6, D-8, D-9. Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/...` passes.

**STEP-4 — `onboarding.AddProductVocabulary`**
Files: `core/onboarding/addvocab.go` *(new)*, `core/onboarding/addvocab_test.go` *(new)*.
Do: D-4, D-5. Add T-5, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/onboarding/...` passes.

**STEP-5 — The terminal command**
Files: `api/cmd/add-product-vocabulary/main.go` *(new)*.
Do: D-3, D-10. `main` holds no logic beyond argument count, the call, `errors.Is(err, database.ErrAlreadyOpen)`, printing and `os.Exit`.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/...` passes; `CGO_ENABLED=1 go run -tags fts5 ./api/cmd/add-product-vocabulary` prints the usage line and exits 2; T-5 covers the run against a stripped catalog; the real-project run is §10.

**STEP-6 — Docs and skill**
Files: `docs/seeded-vocabulary.md` (§4.5 paragraph: replace the last three sentences, from "`namealiases.AddProductNicknames` copies the product nickname list" through "Call `namealiases.Invalidate` if that catalog stays open.", with "A catalog made before the list grew gets the new sets from `add-product-vocabulary` (S11-44), which calls `namealiases.AddProductNicknames`; run it on a closed project, from the commit the installed app was built from."; and in the subject vocabulary section, one sentence: "`subjectvocab.AddMissing`, through the same command, adds seeded Properties and terms a catalog lacks and binds only the Properties it adds, after the type's existing bindings."), `.agents/skills/add-seeded-vocabulary/SKILL.md` (under **Install semantics (pinned)**, a bullet: "**Existing catalogs** get new `subjectvocab` Properties and terms and new product nickname sets only from `CGO_ENABLED=1 go run -tags fts5 ./api/cmd/add-product-vocabulary <project folder>`, run by the researcher on a closed project (`onboarding.AddProductVocabulary`). It adds what is missing by `(key, origin)` and never rewrites a row. A new seed domain that should reach old catalogs adds a function to it."), `core/database/namealiases/namealiases.go` package comment (the sentence "AddProductNicknames copies the current product nickname sets into a catalog that already exists." ends "exists; the add-product-vocabulary command calls it.").
Do: as listed.
Check: `python3 scripts/check-guidance-names.py` passes.

**STEP-7 — Map and plan docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row status), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-44-des-nickname.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/namealiases/namealiases_test.go` `TestDesIsANicknameOfDesireeAndDesirae` *(new)* | New catalog, `Install`, `Load`, rules with the loaded tables: given "des" vs "desiree" → `Equivalent`, similarity `AliasCredit`; "des" vs "desirae" → same; "desiree" vs "desirae" → not `Equivalent` | AC-1.1, AC-1.2 |
| T-2 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestValueRulesFoldNicknamesOnly`, three rows | `{"Des", "Desiree", "Desiree"}`, `{"Des", "Desirae", "Desirae"}`, `{"Desiree", "Desirae", ""}` | AC-1.3 |
| T-3 | Go | `core/database/subjectvocab/subjectvocab_test.go` `TestAddMissing` *(new)*, table-driven | (a) Install, then delete `religion`, `languages_spoken`, their terms and bindings by SQL → `Added{2, 2, 2}`, `assertMatchesRegistry` passes, person's last two bindings are `religion`, `languages_spoken`; (b) run again → `Added{}`; (c) Install, relabel `interests` to "Hobbies" by SQL and delete its person binding → `Added{}`, label "Hobbies", no binding; (d) Install, delete the `catholic` term only → `Added{0, 1, 0}` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-6 |
| T-4 | Go | `subjectvocab_test.go` `TestInstallMatchesRegistry` | Now calls `assertMatchesRegistry`; still passes | D-8 |
| T-5 | Go | `core/onboarding/addvocab_test.go` `TestAddProductVocabulary` *(new)* | `createCatalog` in a temp dir, close; strip as T-3 (a) plus the two Des sets' `name_alias_members` rows; `AddProductVocabulary(dir)` → `Subject{2,2,2}`, `Nicknames{Sets: 2, Tokens: 4}`, not `Empty()`; again → `Empty()` | AC-2.1, AC-2.2 |
| T-6 | Go | `addvocab_test.go` `TestAddProductVocabularyRefusesAnOpenProject` *(new)* | Hold the catalog open with `database.Open(dir)`; `AddProductVocabulary(dir)` → `errors.Is(err, database.ErrAlreadyOpen)` | AC-2.5, D-4 |
| T-7 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestProductAliasTablesPinCacheVersion` | Re-pinned at 19 → passes | D-2 |
| T-8 | Go | `core/database/namealiases/registry_test.go` `TestNicknameDirectionsCoverTwoWayPairs` | 1229 sets, 108 directions, 143 two-way | D-2 |

AC-2.6 is the manual check (step 2). Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/autoreconcile` and `core/search` goldens that name "Des" (none at the planning base: `rg -i '\bdes\b' core --type go -l` lists no test fixture), and `TestOpenCatalogDoesNotHeal…` (unchanged; the command is not open).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no app copy (D-10)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it) — "Call `namealiases.Invalidate` if that catalog stays open" gone from docs
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status `Landed #N`; completed.md: a section at the top, `## S11-44 — Des as a nickname of Desiree and Desirae`, `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and two sentences: Des is a nickname of Desiree and Desirae in new projects, and `add-product-vocabulary` adds new nickname sets and seeded Properties to a project made earlier
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/seeded-vocabulary.md §4.5 and the subject vocabulary sentence (STEP-6); add-seeded-vocabulary skill bullet
- [ ] PR description: what shipped, AC evidence, deviations, manual check, and the command line to paste
```

## 10. Manual check

1. Build the app from this branch. Quit Provenencia (or close the project).
2. In a terminal at the checkout: `CGO_ENABLED=1 go run -tags fts5 ./api/cmd/add-product-vocabulary` → usage line, exit 2.
3. Copy the dogfood project folder (made before S11-48) as a backup. Run `CGO_ENABLED=1 go run -tags fts5 ./api/cmd/add-product-vocabulary "<path to the project folder>"` → `Added 2 nickname sets (4 names), 2 properties, 2 terms and 2 bindings.` (counts may be higher if the project predates other seed rows). Run it again → `Already up to date.`
4. Open the project in Provenencia, then run the command again → the "project is open" line, exit 1.
5. In the app: a person's Properties include Religion and Languages spoken; Promote on a Source naming "Des" offers a filed "Desiree" (and the reverse).
6. Create a new project: same Promote result with no command.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A checkout newer than the installed app migrates the catalog past the app's schema | D-11; §10 builds the app from the same branch; the doc line says so |
| `AddProductNicknames` merges `{desiree, des}` into a stored set that holds `des` | No stored set holds `des` (seed has none; a researcher alias editor does not exist); its "sole smaller set" rule needs a stored subset |
| The fold of "Des" when one entity has both Desiree and Desirae | S11-28's `Expands` keeps the head of the set that holds the other; with two heads the reconciler sees two candidates — out of this PR's AC; logged if T-2's neighbours show it |

**Stop and report if:**

- `nicknameSets` already holds `des`, `desiree` or `desirae` at your base.
- `TestNicknameDirectionsCoverTwoWayPairs` reports a two-way count other than 143 after D-1 (a set you added pairs both ways).
- `properties.Upsert` or `propertyterms.Upsert` cannot share one helper with the `Tx` variant without changing an existing caller's signature.
- `database.Open` on a project the app holds does not return `ErrAlreadyOpen` (T-6 fails for another reason).
- `allBindings()` gives a binding for `religion` or `languages_spoken` to a type other than person.

## 12. Open questions

None.

## 13. PR

**Title:** S11-44 — Des as a nickname of Desiree and Desirae
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; the command to paste (`CGO_ENABLED=1 go run -tags fts5 ./api/cmd/add-product-vocabulary "<project folder>"`); `Plan: docs/deployment-plan/spike-11/prs/S11-44-des-nickname.md`.
