# S11-22 — Pure value types

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 1 — Graspable model and pure foundations
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a Go package move)
**Board:** —
**Depends on:** —
**Blocks:** S11-23, S11-02
**Branch:** `s11-22-pure-value-types`
**Version bump:** none — internal refactor, no behavior change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rules "Pure value types and kinds come first", "Pure means the package builds and tests with no build tags and no `core/database` import", "High tolerance for churn"; [`comparison-layers.md`](../comparison-layers.md#purity-requirements)
**Planned:** 2026-10-17 against `origin/main` @ `af50bd6`
**Review:** [`S11-22-pure-value-types.review.md`](S11-22-pure-value-types.review.md) — Fits with changes / Ready with changes → fixed, 1 round

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

The date value, the name value, and the property value-type and cardinality constants move into three pure packages that import nothing from `core/database`. Every layer Spike 11 builds depends on these types; today, using any of them pulls in the whole database package and its start-up migration, which is why `go test ./core/match` fails without `-tags fts5`.

**Why now.** S11-02 builds the first pure layer (value compare). Without this move it would either import `core/database` or define its own date and name structs, and every loader would convert between two copies.

**What stays.** Behavior is unchanged. The database packages `core/database/datevalues`, `core/database/namevalues` and `core/database/properties` keep their queries and their package names. Swift and the FFI wire format are untouched.

## 2. User stories

**US-1.** As a contributor building a pure matching layer, I can import the date value, the name value and the property constants without importing `core/database`, so the layer builds and tests with no build tags.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `go list -deps ./core/datevalue ./core/namevalue ./core/propertymeta` | No line starts with `github.com/mendahu/provenencia/core/database` |
| AC-1.2 | `go test ./core/datevalue/... ./core/namevalue/... ./core/propertymeta/...` with no build tags | Passes |
| AC-1.3 | A search for the moved symbols under their old packages in non-test and test Go files | No hits (D-6's pattern) |

**US-2.** As a researcher, nothing about dates, names or properties behaves differently.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The whole Go and Xcode suite | `scripts/check-pr.sh` prints `check-pr: PASS` with no test changed except moved files and import paths |
| AC-2.2 | `datevalue.CompactDisplay` on today's display cases | Same strings as before the move (`About 1882-04`, `Between 1882 and 1882-04-03`) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `core/propertymeta` (new): `ValueType*` (six), `CardinalitySingle`, `CardinalityMultiple` | STEP-1, STEP-2, AC-1.1 |
| `core/datevalue` (new): `Value`, `KindPoint`, `KindRange`, `QualifierABT` / `BEF` / `AFT`, `ErrInvalid`, `Validate`, `CompactDisplay` | STEP-3, STEP-4, AC-1.1, AC-2.2 |
| `core/namevalue` (new): `Value`, `Part`, `PartType*`, `ErrInvalid`, `Validate`, `PartTypeInfo`, `PartTypes`, `KnownPartType`; `core/namevalue/namevaluetest` (new) `Western` | STEP-5, STEP-6, AC-1.1 |
| Every caller moved to the new packages; the old symbols deleted | STEP-2, STEP-4, STEP-6, AC-1.3 |
| Skills, `.cursor` rules, domain docs and Swift doc comments name the new paths | STEP-8 |

| Out | Home |
| --- | --- |
| A purity guard test | S11-02 (working rule "One guard, not one per package"); it lists these three packages |
| Primary kinds and locked property keys | S11-23 |
| `properties.Property`, `OriginCounts`, the error vars and `OriginProvenencia` / `OriginUser` | stay in `core/database/properties`. Origin is catalog-wide (nine packages declare it); one pure home for it is S11-23 |
| Moving the string trimming out of `InsertTx` / `UpdateTx` | not planned: it stays in the DB functions |
| Making `core/match`, `core/autoreconcile` or `core/graphalign` pure | S11-05, S11-06, S11-13: they still import other database packages |
| Swift's copy of `sideLessOrEqual` (`DateValueDraft.swift`) | not planned: client-side validation |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Date value | `core/database/datevalues/datevalues.go` `Value` (L83), `KindPoint`, `KindRange`, `QualifierABT/BEF/AFT` (L37–43, in a const block with the unexported `sql*` constants L45–79), `ErrInvalid` (L35), `validate` (L267) and its helpers `side`, `pointQualifierOK`, `hasCivil`, `validateComponentRanges`, `side.empty`, `sideLessOrEqual`, `monthOK`, `dayOK` — all pure | The file imports `core/database` for `*database.Catalog` and `database.UniqueBlobIDs`, and `sqlbatch` |
| Date display | `core/database/datevalues/display.go` `CompactDisplay` and `appendPhrase`, `applyQualifier`, `formatSide`; `display_test.go` | Pure; imports `fmt`, `strings` |
| Date queries | `datevalues.go` `Insert`, `InsertTx`, `UpdateTx`, `Lookup`, `LookupManyTx`, `LookupTx`, `nullIfEmpty`, `nullInt`, `intPtr`, `scanValue` | Stay. `InsertTx` / `UpdateTx` trim, then call `validate` |
| Name value | `core/database/namevalues/namevalues.go` `Value` (L68), `Part` (L60), `PartType*` (L29–35, in a block with `sql*` L37–56), `ErrInvalid` (L25), `validate(form, parts)` (L300) | Pure symbols in a package that imports `core/database` |
| Name part registry | `core/database/namevalues/registry.go` `PartTypeInfo`, `PartTypes`, `KnownPartType`, `partTypeRegistry`, `partTypeByKey`, `init`; `registry_test.go` | Pure; imports `strings` |
| Name test helper | `core/database/namevalues/namevaluestest/namevaluestest.go` `Western` | Pure; 11 test files import it |
| Name queries | `namevalues.go` `Insert`, `InsertTx`, `UpdateTx`, `Lookup`, `LookupTx`, `LookupMany`, `LookupManyTx`, `nullIfEmpty` | Stay |
| Property constants | `core/database/properties/properties.go` const block L37–51: `OriginProvenencia`, `OriginUser` (stay), `ValueType{Text,Integer,Date,Name,Subject,Term}`, `CardinalitySingle`, `CardinalityMultiple` (move); `sql*` constants from L53; used inside `valueTypeOK` L466, `cardinalityOK` L475 | Pure constants in a package that imports `core/database`, `catalogmodel`, `rowchange`, `deleteimpact` |
| Importers | Counted by import path (`rg -l 'provenencia/core/database/<pkg>"' core api`). datevalues: 21 non-test files in 12 packages; 7 also call queries (`core/database/autoreconciler/autoreconciler.go`, `core/database/matching/matching.go`, `core/database/observations/observations.go`, `core/database/observations/update.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/promotealign/exhibits.go`, `core/database/sourcefill/load.go`); 26 test files. namevalues: 25 non-test files in 16 packages, 8 call queries; 26 test files. properties: 18 non-test files; the import survives in 4 that use queries or `properties.Property` (`api/ffi/handlers/subject_defs.go`, `core/database/connect/connect.go`, `core/database/subjectvocab/subjectvocab.go`, `core/database/observations/observations.go`); `api/ffi/handlers/subject_defs.go` and `core/database/connect/connect.go` use no moved constant and do not change; 45 test files | Pure-only users include `core/valuecodec`, `core/eventtitle`, `core/autoreconcile`, `core/match`, `core/graphalign`, `api/ffi/handlers/source_read.go` |
| Skills, rules and docs | `.agents/skills/add-date-value/SKILL.md` L4, L12, L19–24, L71; `.agents/skills/add-name-value/SKILL.md` L4, L12, L19–24, L39, L57–58; `.agents/skills/add-catalog-query/SKILL.md` L30 and L40 ("Domain validation sentinels stay in this package"); `.cursor/rules/date-values.mdc`, `.cursor/rules/name-values.mdc` (`description`, `globs`); `docs/seeded-vocabulary.md` L501; `docs/structured-name-model.md` L66; Swift doc comments `macos/App/Features/Names/NameValueDraft.swift` L4, `macos/App/Features/Dates/DateValueDraft.swift` L4, `macos/App/Features/Names/NamePartType.swift` L3; `core/match/profiles.go` L14 comment | Name only `core/database/datevalues` / `namevalues` / `properties` |
| Tests | `core/database/datevalues/datevalues_test.go` `TestInsertLookup` and peers; `display_test.go`; `core/database/namevalues/namevalues_test.go`; `registry_test.go` | Validation is tested through `Insert` against a catalog |

**Precedent.** `e20a542` "Move the stored-bytes purge after a delete into the files package so every delete can use it." moved code between packages and repointed every caller in one commit. Copy that: one package per commit, callers repointed in the same commit, nothing left behind. There is no earlier pure-types split.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | New packages are `core/datevalue`, `core/namevalue` (singular) and `core/propertymeta`. Package docs say "pure value type; queries live in `core/database/datevalues`" (and the same for names) | Singular names never collide with the database packages, so a caller using both needs no alias; they read well (`datevalue.Value`, `namevalue.PartTypeGiven`); `core/` is where the repo keeps pure domain packages (`core/eventtitle`, `core/valuecodec`) | `core/dates` / `core/names` — because `dates` and `names` are common local variable names and would shadow the package; the same name under a new path — because callers using both would need aliases |
| D-2 | Types move; the database packages keep their names and their queries, and import the pure packages | The queries need `*database.Catalog`; moving the queries instead would rename 15 query call sites for no gain | Moving the queries into a new package — because the types are what the layers need |
| D-3 | `validate` becomes exported `Validate(v Value) error` in each pure package, with its helpers; `InsertTx` / `UpdateTx` call `datevalue.Validate(v)` / `namevalue.Validate(namevalue.Value{Form: v.Form, Parts: parts})` after their trimming | Validation is pure and is part of the type's contract; the layers and the FFI can call it | Leaving `validate` in the DB package — because the pure package would then accept values it can't check |
| D-4 | `ErrInvalid` moves with `Validate`; the database packages return the pure package's `ErrInvalid` (same `apperr` code, `core/apperr/apperr.go` L34–35) | One error value per type; `errors.Is` callers update their import | Two `ErrInvalid` vars with one code — because `errors.Is` would fail across them |
| D-5 | `propertymeta` holds only the eight value-type and cardinality constants, untyped exactly as today. `OriginProvenencia` / `OriginUser`, `properties.Property`, `OriginCounts`, the error vars and `valueTypeOK` / `cardinalityOK` / `originOK` stay in `core/database/properties`; the two checks read `propertymeta` | Mechanical move; typed `ValueType` / `Cardinality` types are S11-02's call when the registry needs typed keys | Introducing typed constants now — because it changes every comparison in 60 files for no behavior |
| D-6 | Done means none of these remain: `datevalues\.(Value|KindPoint|KindRange|QualifierABT|QualifierBEF|QualifierAFT|ErrInvalid|CompactDisplay)\b`, `namevalues\.(Value|Part|PartType\w*|PartTypes|PartTypeInfo|KnownPartType|ErrInvalid)\b`, `properties\.(ValueType\w+|Cardinality\w+)\b` (and the `properties.Cardinality*` / `properties.ValueType*` comment wording), and `namevaluestest` | The churn rule: no aliases, no re-exports, nothing kept for compatibility | Type aliases in the old packages — because they are compatibility shims |
| D-7 | Moved tests: `display_test.go` → `core/datevalue/display_test.go` (new), as external `package datevalue_test`, cases unchanged; `registry_test.go` → `core/namevalue/registry_test.go` (new), internal `package namevalue`; `namevaluestest` → `core/namevalue/namevaluetest` (new). New `TestValidate` files: `core/datevalue/validate_test.go` (new) internal `package datevalue`; `core/namevalue/validate_test.go` (new) external `package namevalue_test`, because it uses `namevaluetest` (an internal test importing it would be an import cycle). The catalog-backed tests in the database packages stay, with `ErrInvalid` requalified | Pure functions are tested where they live, with no catalog | Moving the `Insert` tests — because they test the queries |
| D-8 | One commit per package, in the order propertymeta → datevalue → namevalue, each repointing every caller and passing the full suite | Each commit leaves the tree building; a reviewer can read one move at a time | One big commit — because a failure in one move hides the others |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/propertymeta` *(new)* | package | `core/database/properties` | That package imports `core/database`, `catalogmodel`, `rowchange`, `deleteimpact` (D-5) |
| `core/datevalue` *(new)* | package | `core/database/datevalues` | Imports `core/database` and `sqlbatch` for its queries (D-2) |
| `core/namevalue` *(new)* | package | `core/database/namevalues` | Same as above |
| `core/namevalue/namevaluetest` *(new)* | test helper | `core/database/namevalues/namevaluestest` | Moves; the old one is deleted (D-6) |
| `datevalue.Validate`, `namevalue.Validate` *(new)* | func | `datevalues.validate`, `namevalues.validate` | Exported moves of the same code (D-3) |
| `TestValidate` ×2 *(new)* | test | `datevalues_test.go`, `namevalues_test.go` | Those go through `Insert` and a catalog (D-7) |

**Builds on.** `core/apperr` codes `CodeDateValuesInvalid`, `CodeNameValuesInvalid`; the pure-package convention of `core/eventtitle` and `core/valuecodec`; precedent `e20a542`.
**Deletes.** `core/database/namevalues/namevaluestest`; the moved symbols from `datevalues`, `namevalues`, `properties`; `display.go`, `display_test.go`, `registry.go`, `registry_test.go` from the database packages.
**Next adds.** S11-02's value compare imports `datevalue`, `namevalue` and `propertymeta`, and its purity guard lists all three (map entry updated in STEP-9). S11-09 adds `propertymeta.Meta{ValueType, Cardinality}` as property compare's `PropertyMeta`, so the concept has one home. S11-23 adds primary kinds, locked property keys and one pure home for the Origin constants.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/propertymeta` *(new)* | `propertymeta.go`: the eight value-type and cardinality constants from `properties.go` L37–51, unchanged values |
| `core/database/properties` | The eight constants removed; `valueTypeOK`, `cardinalityOK` and every internal use read `propertymeta`; Origin constants stay |
| `core/datevalue` *(new)* | `datevalue.go`: `Value`, the five kind and qualifier constants, `ErrInvalid`, `func Validate(v Value) error` and its unexported helpers; `display.go`: `CompactDisplay` and helpers |
| `core/database/datevalues` | Keeps queries and the `sql*` constants; `InsertTx` / `UpdateTx` call `datevalue.Validate`; signatures take and return `datevalue.Value` |
| `core/namevalue` *(new)* | `namevalue.go`: `Value`, `Part`, `PartType*`, `ErrInvalid`, `func Validate(v Value) error`; `registry.go`: `PartTypeInfo`, `PartTypes`, `KnownPartType` and their unexported state |
| `core/database/namevalues` | Keeps queries; calls `namevalue.Validate`; signatures use `namevalue.Value` |
| Every importer | Import path and qualifier updated; local variables named `datevalue` or `namevalue` renamed where they would shadow the package |

Errors: unchanged codes; `datevalue.ErrInvalid` and `namevalue.ErrInvalid` are the only values.

### FFI

None — `api/ffi/handlers/source_read.go` changes its import only; `api/ffi/handlers/subject_defs.go` does not change; no proto change.

### macOS

None — no Swift code change; three Swift doc comments name the new package (STEP-8).

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Create `core/propertymeta`**
Files: `core/propertymeta/propertymeta.go` (new), `core/database/properties/properties.go`. Skill: `add-catalog-query`.
Do: move the eight value-type and cardinality constants (D-5); repoint `properties.go`'s own uses.
Check: `go build ./core/propertymeta/ ./core/database/properties/` passes.

**STEP-2 — Repoint property constant users**
Files: every file using `properties.ValueType*` or `properties.Cardinality*` (`rg -l 'properties\.(ValueType|Cardinality)' core api`; 17 files at the planning base, plus their tests), and the comments in `core/match/profiles.go` L14, `core/match/pair.go` and `core/autoreconcile/autoreconcile.go`.
Do: replace the qualifier with `propertymeta`; add the import; drop the `properties` import where nothing else uses it (it survives in the four files §4 names). Comments naming `properties.ValueType*` or `properties.Cardinality*` say `propertymeta`.
Check: `gofmt -l core api` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes. Commit (D-8).

**STEP-3 — Create `core/datevalue`**
Files: `core/datevalue/datevalue.go` (new), `core/datevalue/display.go` (new, moved), `core/datevalue/display_test.go` (new, moved), `core/datevalue/validate_test.go` (new), `core/database/datevalues/datevalues.go`; delete `core/database/datevalues/display.go`, `display_test.go`. Skill: `add-date-value`.
Do: move the pure symbols and `validate` as `Validate` (D-3, D-4); split the const block so the `sql*` constants stay; the database package calls `datevalue.Validate`.
Check: `go test ./core/datevalue/` passes with no build tags; `go list -deps ./core/datevalue | grep core/database` prints nothing.

**STEP-4 — Repoint date value users**
Files: the 21 non-test and 26 test files from `rg -l 'provenencia/core/database/datevalues"' core api` (not `engine.proto`, `engine.pb.go` or `core/apperr/apperr.go`, whose `datevalues.invalid` is an error code string and stays).
Do: pure symbols to `datevalue`; query calls stay on `datevalues`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; the `datevalues` part of D-6's pattern has no hits. Commit (D-8).

**STEP-5 — Create `core/namevalue`**
Files: `core/namevalue/namevalue.go` (new), `core/namevalue/registry.go` (new, moved), `core/namevalue/registry_test.go` (new, moved), `core/namevalue/validate_test.go` (new), `core/namevalue/namevaluetest/namevaluetest.go` (new, moved), `core/database/namevalues/namevalues.go`; delete `core/database/namevalues/registry.go`, `registry_test.go`, `namevaluestest/`. Skill: `add-name-value`.
Do: move the pure symbols, `validate` as `Validate`, and the part-type registry; split the const block.
Check: `go test ./core/namevalue/...` passes with no build tags; `go list -deps ./core/namevalue | grep core/database` prints nothing.

**STEP-6 — Repoint name value users**
Files: the 25 non-test and 26 test files from `rg -l 'provenencia/core/database/namevalues' core api` (the `namevalues.invalid` error code string in `core/apperr/apperr.go` stays).
Do: pure symbols to `namevalue`; `namevaluestest.Western` to `namevaluetest.Western`; query calls stay on `namevalues`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; D-6's full pattern has no hits. Commit (D-8).

**STEP-7 — Verify purity**
Files: none.
Do: run AC-1.1 and AC-1.2.
Check: `go list -deps ./core/datevalue ./core/namevalue ./core/propertymeta | grep 'provenencia/core/database'` prints nothing; `go test ./core/datevalue/... ./core/namevalue/... ./core/propertymeta/...` passes without `-tags`.

**STEP-8 — Skills and domain docs**
Files: `.agents/skills/add-date-value/SKILL.md`, `.agents/skills/add-name-value/SKILL.md`, `.agents/skills/add-catalog-query/SKILL.md`, `.cursor/rules/date-values.mdc`, `.cursor/rules/name-values.mdc`, `docs/seeded-vocabulary.md`, `docs/structured-name-model.md`, `macos/App/Features/Names/NameValueDraft.swift`, `macos/App/Features/Dates/DateValueDraft.swift`, `macos/App/Features/Names/NamePartType.swift` (comments only).
Do: each names the pure package for types, validation, display and the part-type registry, and the database package for queries. Update the package trees, test commands, and the `.mdc` `description` and `globs` (both packages). In `add-catalog-query`, rewrite L40: a value type that matching layers compare lives in a pure `core/<singular>` package that owns its type, `Validate` and `ErrInvalid`, and the table package keeps only the queries; add to its "Do not" list: no new value struct under `core/database/<table>` when a matching layer compares it.
Check: `rg -n 'core/database/(datevalues|namevalues)' .agents docs/seeded-vocabulary.md docs/structured-name-model.md` lists only lines about queries.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row; S11-02 scope entry), `docs/deployment-plan/spike-11/comparison-layers.md` (§Layer 2 `PropertyMeta`), this plan.
Do: per the Definition of done. In the S11-02 scope entry, add "the guard also lists `core/datevalue`, `core/namevalue` and `core/propertymeta`". In `comparison-layers.md` §Layer 2 Contract, say `PropertyMeta` is `propertymeta.Meta`.
Check: `python3 scripts/check-plan-docs.py` reports no new errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/datevalue/display_test.go` *(new)*, moved | `TestCompactDisplay`'s seven cases unchanged (`About 1882-04`, `Between 1882 and 1882-04-03`) | AC-2.2, D-7 |
| T-2 | Go | `core/datevalue/validate_test.go` `TestValidate` *(new)* | Point with month 13 → `ErrInvalid`; range with start after end → `ErrInvalid`; point `ABT 1817` → nil; empty value → `ErrInvalid` | AC-1.2, D-3 |
| T-3 | Go | `core/namevalue/registry_test.go` *(new)*, moved | Today's cases unchanged | AC-1.2, D-7 |
| T-4 | Go | `core/namevalue/validate_test.go` `TestValidate` *(new)*, `package namevalue_test` | Unknown part type → `ErrInvalid`; form and parts empty → `ErrInvalid`; `*namevaluetest.Western("James Robins")` → nil | AC-1.2, D-3, D-7 |
| T-5 | Go | the existing `core/database/datevalues` and `namevalues` tests | Unchanged except types and imports | AC-2.1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/observations`, `core/database/conclusionheaders`, `core/match` and `core/autoreconcile` tests, which build many values.

Before writing T-2 and T-4, read today's `validate` functions and pick the cases from their branches; a case above that `validate` does not reject today is a Stop condition, not a behavior change.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/seeded-vocabulary.md, docs/structured-name-model.md (package paths); skills and .cursor rules per STEP-8
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Run `go list -deps ./core/datevalue ./core/namevalue ./core/propertymeta | grep 'provenencia/core/database'`: no output (AC-1.1).
2. Run `go test ./core/datevalue/... ./core/namevalue/... ./core/propertymeta/...` without build tags: passes (AC-1.2).
3. Run D-6's pattern with `rg -n -e` over `core` and `api`: no hits (AC-1.3).
4. Run `scripts/check-pr.sh`: `check-pr: PASS` (AC-2.1). On Linux it prints `SKIP xcode-test "not macOS"`; that is accepted because only Swift comments change.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A local variable named `datevalue` or `namevalue` shadows the new package | §6 Go core: rename the local in the same file |
| An import cycle appears | The pure packages import only the standard library and `core/apperr`; the database packages import them, never the reverse |
| Validation behavior drifts during the move | `Validate` is moved code; T-2 and T-4 pin today's branches |

**Stop and report if:**

- A pure symbol in §4 turns out to need `database/sql`, `core/database` or `sqlbatch`.
- `go list -deps` of a new package includes anything under `core/database` after its step.
- A caller needs a moved symbol and a database query in a way that creates an import cycle.
- A T-2 or T-4 case does not match what `validate` does at the planning base.
- More files than §4 counts import the old packages by path at your base (`origin/main` moved).

## 12. Open questions

None.

## 13. PR

**Title:** S11-22 — Pure value types
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-22-pure-value-types.md`.
