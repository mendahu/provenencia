# S10-66 — Pure depiction value package

**Status:** Draft
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: engine refactor)
**Board:** —
**Depends on:** —
**Blocks:** S11-19 (Spike 11 map: the claim evidence query carries likeness lines through the pure matching stack)
**Branch:** `s10-66-pure-depiction-value`
**Version bump:** none — follow-up PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 and Spike 11 do ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md) (the shape the new skill copies)
**Rules that bind:** `.cursor/rules/go-tests.mdc` — the two new tests are tables; `.cursor/rules/catalog-queries.mdc`
**Context:** Spike 11 working rule "Pure value types and kinds come first" ([`spike-11/deployment-plan.md`](../../spike-11/deployment-plan.md)); [purity](../../spike-11/comparison-layers.md#purity-requirements); [Spike 10 decisions](../../archive/spike-10/README.md) ("A depiction is an owned value row: a region plus an image File"); [conclusion reconciliation, depiction row](../../../conclusion-reconciliation.md); working rules: "Spike 10 decisions hold"
**Planned:** 2026-10-18 against `main` @ `e97cf54`
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

The depiction value type moves into a pure package, `core/depictionvalue`, the way S11-22 moved the date and name value types. Today `core/autoreconcile` carries a likeness as a `*depictionvalues.Value`, so `core/match`, `core/graphalign` and `core/autoreconcile` reach `core/database` through that one struct (audit 2026-10-18, MNT-1). The PR also gives depictions the skill and rule that dates and names already have, so the read and write rules stop living only in doc comments (audit MNT-4).

**Why now.** S11-19 builds the claim evidence query through the pure matching stack, and S11-20 shows likeness lines as two crops. Both carry a depiction value through layers the purity guard keeps free of `core/database`. Without this move they would either import the database package or add a second depiction struct.

**What stays.** Behavior is unchanged. `core/database/depictionvalues` keeps its name, its queries (`InsertTx`, `LookupTx`) and the join helper `FromJoin`. The error code string `depictionvalues.invalid` is unchanged, so Swift's `L10n.Errors` mapping and the FFI wire format are untouched. No schema, FFI or Swift code changes.

## 2. User stories

**US-1.** As a contributor building a matching layer, I import the depiction value from a pure package so that the layer stays free of `core/database`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `go list -deps ./core/depictionvalue \| grep 'provenencia/core/database'` | Prints nothing |
| AC-1.2 | `go list -deps ./core/match ./core/graphalign ./core/autoreconcile \| grep 'provenencia/core/database'` (and the same with `-test`) | Prints nothing (at the base it prints `core/database/depictionvalues`) |
| AC-1.3 | `go test ./core/depictionvalue/ ./core/matchregistry/` with no build tags | Passes; `TestLayerPackages` has a `core/depictionvalue/pure` subtest |
| AC-1.4 | `rg -n 'depictionvalues\.(Value\|Readable\|ReadableImage\|ErrInvalid)\b' core api` | No hits: those symbols exist only in `core/depictionvalue` |

**US-2.** As the engine, I store, read and reconcile likenesses exactly as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `depictionvalue.Validate` on `{RegionJSON: "  ", ImageFileID: <16 bytes>}`, on `{RegionJSON: LikenessRegion, ImageFileID: <15 bytes>}`, and on `{RegionJSON: LikenessRegion, ImageFileID: <16 bytes>}` | `ErrInvalid`, `ErrInvalid`, nil — the same shape `InsertTx` checks at the base |
| AC-2.2 | `depictionvalue.Readable` with an on-page region and `image/png`; an off-page region; `image/jpeg`; an empty image id | Region and image kept; region emptied; image emptied; image emptied — the same results `TestFromJoin` asserts at the base |
| AC-2.3 | The existing tests of `depictionvalues`, `observations`, `sourcefill`, `autoreconciler`, `likenesses`, `conclusiondetails`, `conclusionheaders`, `autoreconcile` and `api/ffi/handlers` | Pass with only import and qualifier changes |
| AC-2.4 | A refused depiction save | Still returns error code `depictionvalues.invalid` (`apperr.CodeDepictionValuesInvalid`) |

**US-3.** As an agent adding a depiction reader or writer, I load one skill and one rule that name the package, the read rule, the write path, release, the gallery rules and the test fixtures.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Open `.agents/skills/add-depiction-value/SKILL.md` *(new)* | It has the sections of §6 Skill, each naming existing symbols |
| AC-3.2 | Edit a file under `core/depictionvalue/` or `core/database/depictionvalues/` in Cursor | `.cursor/rules/depiction-values.mdc` *(new)* applies by its globs |
| AC-3.3 | `rg -n 'depictionvalues\.Readable' docs .agents .cursor` outside `docs/deployment-plan/` | No hits: docs name `depictionvalue.Readable` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/depictionvalue` *(new)*: `Value`, `ErrInvalid`, `Validate`, `Readable`, `ReadableImage` | STEP-1, AC-1.1, AC-2.1, AC-2.2 |
| `core/database/depictionvalues` keeps `InsertTx`, `LookupTx`, `FromJoin` on `depictionvalue.Value`; its copies of the moved symbols are deleted | STEP-2, AC-1.4, AC-2.3 |
| Every importer repointed; comments that name the moved symbols name the new package | STEP-2, AC-1.2, AC-1.4 |
| A `core/depictionvalue` row in `layerPackages` (`core/matchregistry/guard_test.go`) | STEP-3, AC-1.3 |
| `add-catalog-query` rule widened to "a value type a matching layer or the reconciler carries" | STEP-4 |
| `add-depiction-value` skill and `depiction-values.mdc` rule *(new)*; the `plan-pr` skill table names the skill | STEP-4, AC-3.1, AC-3.2 |
| `docs/conclusion-reconciliation.md` names `depictionvalue.Readable` | STEP-5, AC-3.3 |

| Out | Home |
| --- | --- |
| A depiction comparer in `core/valuecompare` (today the type says it has none, `comparison-layers.md` "One entry point per value type") | S11-20 if it compares two likenesses as pictures; not planned otherwise |
| Moving `FromJoin` into the pure package | not planned — D-3 |
| Guard rows for `core/match`, `core/graphalign` or `core/autoreconcile` | not planned — `core/match` is deleted in S11-13; AC-1.2 checks them once here |
| One stage-then-write entry point for depictions (`api/ffi/handlers` `observations.go`, `citations.go`, `connect.go`) | [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) — named in the skill's "Do not" as the current shape, not changed |
| Any change to `locator.ValidateDepictionRegion`, `ingest.StageImage` or the read results | S10-61 (region caps); not this PR |
| Swift `DepictionValueDraft` and the FFI messages | not planned — no wire or app change |

## 4. Current state

Verified at `e97cf54`.

| Area | Where | Today |
| --- | --- | --- |
| Value and validation | `core/database/depictionvalues/depictionvalues.go` `Value`, `ErrInvalid`, `InsertTx` | `Value{ID, RegionJSON, ImageFileID}`. `ErrInvalid = apperr.New(apperr.CodeDepictionValuesInvalid, apperr.KindUser)`. `InsertTx` returns `ErrInvalid` on a nil tx, then trims `RegionJSON` and returns `ErrInvalid` when it is empty or `len(ImageFileID) != 16`: the row's shape check is inline, not a function |
| Read rules | same file `Readable(v, citationLocatorJSON, imageMediaType)`, `ReadableImage(imageFileID, mediaType)` | `Readable` empties `RegionJSON` unless `locator.ValidateDepictionRegion` passes and calls `ReadableImage`, which keeps the id only for `image/png`. Both import only `core/locator` and stdlib |
| Join helper | same file `FromJoin(id, region sql.NullString, imageFileID, citationLocator, imageMediaType sql.NullString) *Value` | Returns nil when `imageFileID` is empty, else `Readable` of copied inputs |
| Imports of the package | `core/database/depictionvalues` imports `database/sql`, `strings`, `github.com/google/uuid`, `core/apperr`, `core/locator` | No `core/database` import in non-test code; its test imports `core/database` and `core/database/files` |
| `core/locator` | `core/locator/validate.go`, `core/locator/depiction.go` | Imports `encoding/json`, `math`, `strings`, `core/apperr` only: pure |
| Importers (non-test) | `api/ffi/handlers/source_read.go` (builds `depictionvalues.Value`, L96); `core/autoreconcile/autoreconcile.go` (`Value.Depiction *depictionvalues.Value`, L89, comment L87–88); `core/database/likenesses/likenesses.go` (`depictionvalues.ReadableImage`, L71); `core/database/observations/depiction.go` (`depictionvalues.Value`, `InsertTx`, `depictionValueMap`); `core/database/observations/observations.go` (`Depiction *depictionvalues.Value` L148, comment L147, `FromJoin` L649); `core/database/observations/update.go` (`LookupTx`, L101); `core/database/sourcefill/load.go` (`FromJoin`, L119); `core/database/autoreconciler/autoreconciler.go` (`FromJoin`, L416) | `rg -l 'provenencia/core/database/depictionvalues"' core api` lists these eight files plus two tests |
| Importers (test) | `core/autoreconcile/modules_test.go` (`&depictionvalues.Value{…}`, comment L20); `core/database/depictionvalues/depictionvalues_test.go` | — |
| Comments naming moved symbols | `core/database/graphcache/source.go` L58 ("depictionvalues.Readable returns it") | — |
| Purity leak | `go list -deps ./core/match ./core/graphalign ./core/autoreconcile \| grep 'provenencia/core/database'` | Prints `github.com/mendahu/provenencia/core/database/depictionvalues` (also with `-test`) |
| Guard | `core/matchregistry/guard_test.go` `layerPackages`, `checkPure` | Rows for `core/valuecompare`, `core/matchregistry`, `core/datevalue`, `core/namevalue`, `core/namevalue/namevaluetest`, `core/propertymeta`, `core/vocabkey`. `checkPure` walks imports, test imports included, and fails on any `core/database` path |
| Error code | `core/apperr/apperr.go` `CodeDepictionValuesInvalid = "depictionvalues.invalid"`; `macos/App/Platform/L10n.swift` maps `"depictionvalues.invalid"` | The string is the wire contract and stays |
| Skill rule | `.agents/skills/add-catalog-query/SKILL.md` L40 ("except for a value type that matching layers compare"), "Do not" ("when a matching layer compares it") | Lets the depiction value stay in the table package because no layer *compares* it, though the reconciler and matching *carry* it |
| Value skills and rules | `.agents/skills/add-date-value/SKILL.md`, `.agents/skills/add-name-value/SKILL.md`, `.cursor/rules/date-values.mdc`, `.cursor/rules/name-values.mdc` | No depiction counterpart; `rg -n 'Readable\|FromJoin' .agents .cursor` finds nothing |
| plan-pr skill table | `.agents/skills/plan-pr/SKILL.md` L133 `\| Dates, names \| add-date-value, add-name-value \|` | No depiction row |
| Domain docs | `docs/conclusion-reconciliation.md` L142 (table row "an image `depictionvalues.Readable` refuses") and L144 ("the loader reads it through `depictionvalues.Readable`") | Name the old package |
| Tests | `core/database/depictionvalues/depictionvalues_test.go` `TestDepictionValues` (insert and lookup, empty region, short image id, missing row), `TestFromJoin` (four rows plus "does not alias inputs") | Cover the shape check and `Readable` through the database package only |

**Precedent.** S11-22 ([#434](https://github.com/mendahu/provenencia/pull/434), plan `docs/deployment-plan/spike-11/prs/S11-22-pure-value-types.md`) did this for dates and names: STEP-3 created `core/datevalue` with the type, `ErrInvalid` and `Validate` (moved from a private `validate`), STEP-4 repointed users with the query calls left on `datevalues`, STEP-8 rewrote the skills, rules and `add-catalog-query`. Copy its package doc split (pure package says "queries live in core/database/…"; table package says "the value type … lives in the pure package"), its `Validate` export, and its commit subjects ("Move the date value type, its validation and its display into a pure package."). Its Landed note: a path that the change deletes is written as plain text in §4, not a code span, because the plan-docs check refuses missing paths — this plan names no deleted file. Do not copy its alias-free deletion blindly: here, too, the old symbols are deleted with no type alias (D-4).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The pure package is `core/depictionvalue`, one file `depictionvalue.go`, holding `Value`, `ErrInvalid`, `Validate`, `Readable`, `ReadableImage` | Singular package beside `core/datevalue` and `core/namevalue`, as `add-catalog-query` names them | `core/depiction` — because the siblings use the `…value` form |
| D-2 | `Readable` and `ReadableImage` move into the pure package | Both import only `core/locator` and stdlib, and `core/locator` is pure (§4). The read rule is a fact about the value, and S11-19's evidence lines and S11-20's crops need it without the catalog | Leaving them in the table package — because a pure layer would then reach `core/database` to apply the read rule |
| D-3 | `FromJoin` stays in `core/database/depictionvalues` and returns `*depictionvalue.Value` | It is the join contract (which tables a reader joins, `sql.NullString` columns) and its callers are all database loaders | Moving it — because its doc names the catalog joins, which a pure package does not own |
| D-4 | The old symbols are deleted; no type alias or wrapper is left in `depictionvalues` | One name per concept; S11-22 did the same | `type Value = depictionvalue.Value` — because an alias keeps two names alive and AC-1.4 could not hold |
| D-5 | `Validate(v Value) error` returns `ErrInvalid` when `strings.TrimSpace(v.RegionJSON) == ""` or `len(v.ImageFileID) != 16`, else nil. It does not trim `v` (it takes a copy) and does not check the region against the Citation | It is exactly today's inline shape check. The Citation check needs the Citation's locator and stays with the caller (`observations.resolveDepiction` → `locator.ValidateDepictionRegion`), as the package doc says today | Folding `ValidateDepictionRegion` into `Validate` — because `InsertTx` has no Citation locator and the write path already checks it before inserting |
| D-6 | `InsertTx` keeps its nil-tx check returning `depictionvalue.ErrInvalid`, trims `RegionJSON`, then calls `depictionvalue.Validate`; `LookupTx` returns `depictionvalue.ErrInvalid` for a nil tx or bad id | Same results as today; the table package owns no error of its own, as `datevalues` uses `datevalue.ErrInvalid` | A table-local `ErrInvalid` — because two sentinels for one code break `errors.Is` |
| D-7 | `ErrInvalid` keeps `apperr.CodeDepictionValuesInvalid` (`depictionvalues.invalid`) | The code string is the FFI and L10n contract | Renaming the code — because Swift maps the string (`L10n.swift`) |
| D-8 | The guard gets one row, `{"core/depictionvalue", false}` | A pure value package, like `core/datevalue`; `numbers` is false because `16` (the id length) is a fact, not a tunable | Rows for `core/match`, `core/graphalign`, `core/autoreconcile` — Out; AC-1.2 checks them once |
| D-9 | `TestFromJoin` stays unchanged in `depictionvalues_test.go`; the pure package gets its own `TestValidate` and `TestReadable` tables | `TestFromJoin` tests the join wrapper (nil on no image, no aliasing) and S10-61's plan relies on its "off-page region" row; the pure tests pin the rules where they now live | Moving `TestFromJoin` rows — because S10-61 cites that row by name |
| D-10 | `add-catalog-query` L40 reads: "…except for a value type that a matching layer or the reconciler carries: that lives in a pure `core/<singular>` package (`core/datevalue`, `core/namevalue`, `core/depictionvalue`) that owns its type, `Validate` and `ErrInvalid`, and the table package keeps only the queries". Its "Do not" line reads "…when a matching layer or the reconciler carries it" | Audit MNT-1: "compares" let this type through although the reconciler and matching carry it | Leaving the rule — because the next value type would slip through the same way |
| D-11 | The new skill is `.agents/skills/add-depiction-value/SKILL.md` with the frontmatter and sections of §6 Skill; the rule is `.cursor/rules/depiction-values.mdc` with the shape of `date-values.mdc` | Dates and names each have a skill and a rule (audit MNT-4); one place for the read, write, release and gallery rules | Adding depiction lines to `add-ffi-handler` and `add-audit-scope` only — because those already carry one line each and the rules stay scattered |
| D-12 | Landing order with the other follow-ups: S10-60 edits `core/database/likenesses/likenesses.go` (the gallery query beside the `ReadableImage` call) and `docs/conclusion-reconciliation.md` L144; S10-61 runs `depictionvalues` tests and edits `core/locator`. Whichever of S10-60, S10-61 and S10-66 lands second merges `main` and keeps both changes: `depictionvalue.ReadableImage` with S10-60's query, and S10-60's rewritten sentence with `depictionvalue.Readable` | The map's rule for shared files | Declaring a dependency — because none needs the other's code |
| D-13 | Two code commits and one docs commit: STEP-1–STEP-3 "Move the depiction value type, its validation and its read rules into a pure package."; STEP-4–STEP-5 "Add the depiction value skill and rule, and point the docs at the pure package."; then the post-PR commit (`implement-pr` Round 0 step 8) | S11-22's subjects; STEP-1 alone leaves a duplicate the next step deletes, so it shares the commit | One commit per step — because STEP-1's copy is not a behavior |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/depictionvalue/depictionvalue.go` *(new)* | file / package | `core/datevalue/datevalue.go`, `core/namevalue/namevalue.go` | Same shape, new type; searched `rg -l 'package .*value$' core` |
| `depictionvalue.Validate` *(new)* | func | the inline check in `depictionvalues.InsertTx` | It is that check, exported (D-5) |
| `core/depictionvalue/depictionvalue_test.go` *(new)* `TestValidate`, `TestReadable` | test | `core/datevalue/validate_test.go` `TestValidate`; `depictionvalues_test.go` `TestFromJoin` | Pure tests next to the rules (D-9); `TestFromJoin` stays for the wrapper |
| `layerPackages` row `core/depictionvalue` *(new)* | test row | the `core/datevalue` row | A row, not a test |
| `.agents/skills/add-depiction-value/SKILL.md` *(new)* | skill | `.agents/skills/add-date-value/SKILL.md`, `.agents/skills/add-name-value/SKILL.md` | No depiction skill exists; searched `ls .agents/skills \| rg -i depict` |
| `.cursor/rules/depiction-values.mdc` *(new)* | rule | `.cursor/rules/date-values.mdc` | Same shape |

**Builds on.** S11-22's split (pure type package, query package), the `layerPackages` guard, the `add-catalog-query` rule, the existing read rule `Readable` and write path `observations.resolveDepiction` → `locator.ValidateDepictionRegion` → `depictionvalues.InsertTx`.
**Deletes.** `depictionvalues.Value`, `depictionvalues.ErrInvalid`, `depictionvalues.Readable`, `depictionvalues.ReadableImage` and the inline shape check in `InsertTx`; the stale comment wording in `core/database/graphcache/source.go`, `core/database/observations/observations.go`, `core/autoreconcile/autoreconcile.go`, `core/autoreconcile/modules_test.go`.
**Next adds.** S11-19 and S11-20 carry `depictionvalue.Value` through pure layers with no new struct. A depiction comparer (S11-20 or later) lives in `core/valuecompare` and imports `core/depictionvalue`. A second depiction key (signature, place photo; `docs/ideas/depictions-and-likenesses.md`) reuses the type and the skill's gallery section names where its gallery rule goes.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/depictionvalue` *(new)* | Package doc: "Package depictionvalue is the pure depiction value type: the region of a Citation's page that shows a face and the image File Go stored from the app's rendering of it, the row's shape check, and the rules a reader applies before showing it. It imports no database package; queries live in core/database/depictionvalues." `var ErrInvalid = apperr.New(apperr.CodeDepictionValuesInvalid, apperr.KindUser)`; `type Value struct { ID, RegionJSON, ImageFileID }` (field comments as today); `func Validate(v Value) error` (D-5); `func Readable(v Value, citationLocatorJSON, imageMediaType string) Value` and `func ReadableImage(imageFileID []byte, mediaType string) []byte`, bodies and doc comments moved verbatim except that `Readable`'s doc says "LookupTx in core/database/depictionvalues does not" |
| `core/database/depictionvalues` | Package doc: "Package depictionvalues stores the depiction_values rows an Observation owns. The value type, its shape check and the read rules live in the pure package core/depictionvalue; this package holds the queries and the join helper. A row is replaced, never rewritten: a new region always comes with a new image. The caller checks the region against the Citation (locator.ValidateDepictionRegion)." `InsertTx(tx *sql.Tx, v depictionvalue.Value) ([]byte, error)` per D-6; `LookupTx(tx *sql.Tx, id []byte) (depictionvalue.Value, error)`; `FromJoin(...) *depictionvalue.Value` calling `depictionvalue.Readable`. Imports: drop `core/locator`, add `core/depictionvalue` |
| `core/autoreconcile` | `Value.Depiction *depictionvalue.Value`; comment "read through depictionvalue.Readable" |
| `core/database/observations` | `depiction.go`: `depictionvalue.Value` for the row and `depictionValueMap`, `depictionvalues.InsertTx` stays; `observations.go`: `Depiction *depictionvalue.Value`, comment names `depictionvalue.Readable`, `depictionvalues.FromJoin` stays; `update.go` unchanged (`LookupTx` stays) |
| `core/database/likenesses` | `depictionvalue.ReadableImage`; drop the `depictionvalues` import |
| `api/ffi/handlers` | `source_read.go`: `&depictionvalue.Value{…}`; comment "Matches depictionvalues.FromJoin's non-nil result" stays |
| `core/database/sourcefill`, `core/database/autoreconciler` | No change beyond what the compiler needs (`FromJoin` stays on `depictionvalues`) |

Errors: unchanged — `depictionvalues.invalid` from `depictionvalue.ErrInvalid`.

### FFI

None — no proto or handler signature change.

### macOS

None — no Swift change. `DepictionValueDraft.swift` and `GenealogyStore.swift` name the table `depiction_values`, which stays.

### Skill

`.agents/skills/add-depiction-value/SKILL.md` *(new)*. Frontmatter `name: add-depiction-value`, `description:` "Uses or extends Provenencia depiction values (pure type in core/depictionvalue, queries in core/database/depictionvalues): a region of a Citation's page plus an image File, the likeness gallery and the preferred portrait. Use when saving, reading, redrawing or reconciling a depiction, value_depiction_id, depiction_values, likeness, Readable, FromJoin, StageDepictions, the gallery or the portrait." Sections, in this order, each naming the symbols in brackets:

1. **Package** — the tree (`core/depictionvalue/depictionvalue.go`, `depictionvalue_test.go`; `core/database/depictionvalues/depictionvalues.go`, `depictionvalues_test.go`), and "Matching layers and the reconciler import `core/depictionvalue`, never `core/database/depictionvalues`"; `TestLayerPackages` keeps it pure.
2. **Model** — owned value row, a region plus an image File, replaced never rewritten; the image is not an Artifact and not a thumbnail; Go re-encodes its own PNG (`ingest.StageImage`, `core/raster`); link `docs/interpretation-layer-data-model.md` and the Spike 10 decisions.
3. **Write path** — the FFI handler stages inside `withProjectCatalog` before `writes.Run` (`ingest.StageDepictions`), `observations.Input.Depiction` (`observations.DepictionInput`), `resolveDepiction` checks `locator.ValidateDepictionRegion` then inserts the File and `depictionvalues.InsertTx`; a redraw releases the old row (`releaseReplacedValue` in `core/database/observations/update.go`); release follows `catalogmodel` owned-outbound FKs (`observations.value_depiction_id`, `depiction_values.image_file_id`); audit: `effects.ownedValue`, see `add-audit-scope`.
4. **Read path** — every reader that returns a depiction to the app goes through `depictionvalue.Readable`, through `depictionvalues.FromJoin` and the three joins its doc names; `LookupTx` does not (audit records what is stored); paths come from `files.RelPaths` / `files.StorageRelPath`, never from a request.
5. **Reconcile, gallery and portrait** — depictions group by image File id (`core/autoreconcile/modules.go`); the gallery is `likenesses.ForEntities`, admitted by `autoreconcile.Reason.InGallery`; `likenesses.Lead` is the one lead rule; the preferred portrait is `canonical_entities.preferred_likeness_observation_id`, written by `canonicalentities.SetPreferredLikeness`.
6. **Tests** — `observationstest.SaveLikeness`, `Crop`, `LikenessRegion`, `ImageOf`, `FileArtifact`, `SaveCitation`; pure checks `go test ./core/depictionvalue/`; query and read tests `CGO_ENABLED=1 go test -tags fts5 ./core/database/depictionvalues/ ./core/database/likenesses/`.
7. **Do not** — store the app's bytes; take a File id or an `objects/…` path from a request; read a stored depiction without `Readable`; add a second depiction struct; add a likeness field to the Person page (the gallery is the display); stage after `writes.Run`.

`.cursor/rules/depiction-values.mdc` *(new)*: frontmatter `description: Depiction values: pure type in core/depictionvalue, queries in core/database/depictionvalues`, `globs: core/depictionvalue/**,core/database/depictionvalues/**,core/database/likenesses/**,core/ingest/depictions.go,core/locator/depiction.go`, `alwaysApply: false`; body "Follow the project skill `.agents/skills/add-depiction-value/SKILL.md`." and three bullets: types, `Validate` and the read rules live in `core/depictionvalue`, storage through `depictionvalues.InsertTx` / `LookupTx` / `FromJoin`; every reader goes through `Readable`; schema changes go through catalog migrations.

`.agents/skills/plan-pr/SKILL.md` table row L133 becomes `| Dates, names, depictions | `add-date-value`, `add-name-value`, `add-depiction-value` |`.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Create `core/depictionvalue`**
Files: `core/depictionvalue/depictionvalue.go` *(new)*, `core/depictionvalue/depictionvalue_test.go` *(new)*. Skill: `add-catalog-query`.
Do: write the package per §6 Go core (D-1, D-2, D-5, D-7): copy `Value`, `ErrInvalid`, `Readable`, `ReadableImage` with their doc comments, and add `Validate`. Write T-1 and T-2.
Check: `go test ./core/depictionvalue/` passes with no build tags; `go list -deps ./core/depictionvalue | grep 'provenencia/core/database'` prints nothing.

**STEP-2 — Repoint the table package and every user; delete the old symbols**
Files: `core/database/depictionvalues/depictionvalues.go`, `core/database/depictionvalues/depictionvalues_test.go`, and the files in §4 "Importers (non-test)", "Importers (test)" and "Comments naming moved symbols".
Do: per §6 Go core and D-3, D-4, D-6. Value types and the read rules use `depictionvalue`; `InsertTx`, `LookupTx` and `FromJoin` stay on `depictionvalues`. Drop an import that nothing uses. In `depictionvalues_test.go`, only the type names change (`depictionvalue.Value`); the test cases stay.
Check: `gofmt -l core api` prints nothing; `rg -n 'depictionvalues\.(Value|Readable|ReadableImage|ErrInvalid)\b' core api` prints nothing (AC-1.4); `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-3 — Guard row and purity check**
Files: `core/matchregistry/guard_test.go`.
Do: add `{"core/depictionvalue", false},` after the `core/namevalue/namevaluetest` row (D-8).
Check: `go test ./core/matchregistry/ -run TestLayerPackages -v` lists `core/depictionvalue/pure` and passes; `go list -deps ./core/match ./core/graphalign ./core/autoreconcile | grep 'provenencia/core/database'` and `go list -deps -test ./core/match ./core/graphalign ./core/autoreconcile | grep 'provenencia/core/database'` print nothing (AC-1.2). Commit: "Move the depiction value type, its validation and its read rules into a pure package." (D-13).

**STEP-4 — Skill, rule and the catalog-query rule**
Files: `.agents/skills/add-depiction-value/SKILL.md` *(new)*, `.cursor/rules/depiction-values.mdc` *(new)*, `.agents/skills/add-catalog-query/SKILL.md`, `.agents/skills/plan-pr/SKILL.md`.
Do: write the skill and rule per §6 Skill (D-11); rewrite `add-catalog-query` L40 and its "Do not" line per D-10; add the `plan-pr` table row.
Check: every symbol the skill names exists (`rg -n 'func (StageDepictions|ForEntities|Lead|SetPreferredLikeness|RelPaths|StorageRelPath|releaseReplacedValue|SaveLikeness|ImageOf|Crop|FileArtifact|SaveCitation)\b' core` lists each once; `rg -n 'LikenessRegion\s*=' core` and `rg -n 'func \(r Reason\) InGallery' core` each print one line); `rg -n 'matching layers compare|a matching layer compares it' .agents` prints nothing.

**STEP-5 — Docs**
Files: `docs/conclusion-reconciliation.md`.
Do: L142 and L144 name `depictionvalue.Readable` (D-12 governs a merge with S10-60's rewrite of L144). The folder has no `completed.md`, and the PR number does not exist yet: this plan's Status and the S10-66 map row are set to `Landed #N` in the post-PR commit, `implement-pr` Round 0 step 8 ("Record S10-66 as landed in pull request N."). This step makes only the doc edit above; it joins STEP-4's commit: "Add the depiction value skill and rule, and point the docs at the pure package." (D-13).
Check: `rg -n 'depictionvalues\.Readable' docs .agents .cursor --glob '!docs/deployment-plan/**'` prints nothing (AC-3.3); `python3 scripts/check-plan-docs.py` reports no errors.

Commits: two, as STEP-3 and STEP-5 say, then the post-PR commit. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/depictionvalue/depictionvalue_test.go` `TestValidate` *(new)*, table | region `"  "` + 16-byte id → `ErrInvalid`; region `""` + 16-byte id → `ErrInvalid`; a three-point region document + 15-byte id → `ErrInvalid`; same region + nil id → `ErrInvalid`; same region + 16-byte id → nil; `errors.Is(err, depictionvalue.ErrInvalid)` and the error's code is `depictionvalues.invalid` | AC-2.1, AC-2.4, D-5, D-7 |
| T-2 | Go | `core/depictionvalue/depictionvalue_test.go` `TestReadable` *(new)*, table | Citation locator `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`, region on page 1, `image/png` → both kept; region on page 9 → region `""`, image kept; `image/jpeg` → region kept, image nil; empty image id → image nil; `ReadableImage(id, "image/png")` → id, `ReadableImage(nil, "image/png")` → nil | AC-2.2, D-2 |
| T-3 | Go | `core/matchregistry/guard_test.go` `TestLayerPackages` | New subtest `core/depictionvalue/pure` passes | AC-1.3, D-8 |
| T-4 | Go | the existing tests named in AC-2.3, including `depictionvalues_test.go` `TestDepictionValues` and `TestFromJoin` unchanged except type names | Pass | AC-2.3, D-3, D-6, D-9 |

T-1 and T-2 write their region document inline as a const in the test file (no `core/database` import is allowed, so `observationstest.LikenessRegion` is out of reach); use the same JSON text as `LikenessRegion` and `TestFromJoin`'s `region`.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/observations`, `core/database/likenesses`, `core/autoreconcile` and `api/ffi/handlers` tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or command output for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks) — on the user's Mac, as ship-pr "Before you start" requires
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] Post-PR commit (implement-pr Round 0 step 8): this plan's Status and the S10-66 row in docs/deployment-plan/spike-10-followups/deployment-plan.md → Landed #N; a Landed note lists every difference from the plan (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/conclusion-reconciliation.md names depictionvalue.Readable; skills and rules per STEP-4
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — behavior is unchanged and no screen changes. The reviewer runs:

1. `go list -deps ./core/match ./core/graphalign ./core/autoreconcile | grep 'provenencia/core/database'` → no output.
2. `go test ./core/depictionvalue/ ./core/matchregistry/` → ok, with no build tags.
3. Opens `.agents/skills/add-depiction-value/SKILL.md` and follows one symbol from each section to its definition.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A file outside §4 imports `depictionvalues` by the time this is implemented (S10-60, S10-61 or Spike 11 landed first) | STEP-2's `rg` check finds it; repoint it the same way and list it in the Landed note |
| `checkPure` walks test imports, so a future test in `core/depictionvalue` that imports a database fixture fails the guard | T-1 and T-2 use inline JSON (§8) |
| S10-60 rewrites `docs/conclusion-reconciliation.md` L144 | D-12: the second to land keeps both |

**Stop and report if:**

- `go list -deps ./core/match ./core/graphalign ./core/autoreconcile | grep 'provenencia/core/database'` prints a package other than `core/database/depictionvalues` at the base (the leak is wider than this plan covers).
- `core/locator` imports anything under `core/database` (`go list -deps ./core/locator | grep 'provenencia/core/database'` prints a line), so `Readable` cannot be pure (D-2).
- A test outside `core/database/depictionvalues` asserts `errors.Is(err, depictionvalues.ErrInvalid)` and cannot switch to `depictionvalue.ErrInvalid` without changing its meaning.
- Any existing test in AC-2.3 needs a change other than an import or a qualifier.

## 12. Open questions

None.

## 13. PR

**Title:** S10-66 — Pure depiction value package
**Body:** objective in one sentence; what shipped against each US; AC evidence (the `go list -deps` outputs before and after); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-66-pure-depiction-value.md`.
