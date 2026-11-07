# Ship log — S14-36 Boolean value type (engine)

**Plan:** [`S14-36-boolean-value-type-engine.md`](S14-36-boolean-value-type-engine.md) · **Branch:** `s14-36-boolean-value-type-engine` · **PR:** #733
**State:** ready

Note: PR #732 used this branch name and carried only the plan, review log and map edits (`main` @ `2ee4adb9`); its branch was deleted on merge. This PR is the first to carry code.

## Round 0 — implement
Head `5cb05743`. Gate PASS (run by the orchestrator). Report deviations: extra tests `TestReconcileBoolean` and `createBeforeTheFirstLoadFallsBackToTheDraftType`; FakeStore does not carry `valueBoolean` through observation writes (left for S14-37); the §10 manual check was not run.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none. UX matches `macos/App/**` and `Localizable.xcstrings`; SEC+PRF matches migrations, `core/valuecompare`, `core/autoreconcile`, `core/subjectprofile`, `api/ffi/**`, `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | `deleteimpact.eraseObservation` keeps its own Observation SELECT, scan and audit row | Fix now → F-2 | Read `facets.go:412-470` and `write.go:136,257`. Same-pass cleanup: supersedes the plan's Out line "leaf package for the Observation row map" |
| MNT-2 | MNT | major | Value columns threaded by hand through seven structs and seven scans | Fix now → F-1 | `HasBoolean bool` in seven structs, verified by grep. Same-pass cleanup: supersedes the Out line "one shared struct" and D-2's field names |
| SEC-1 | SEC | major | An out-of-domain `value_boolean` fails reads and blocks repair | Fix now → F-1 | Contradicted D-1 and D-3, so To user. User chose read-as-absent plus a CHECK on the cache column |
| CON-1 | CON | major | AC-3.5 not tested through the conclusion detail read | Fix now → F-5 | Reviewer's mutants survived; `conclusion_detail_test.go` feeds a hand-built value |
| CON-2 | CON | major | No test on the evidence-graph read of a boolean | Fix now → F-5 | No `HasBoolean` in `graphfill` tests |
| CON-3 | CON | major | FakeStore drops `valueBoolean` on Observation writes | Fix now → F-6 | Implementer's deviation; D-17 and §4 say copy integer line for line |
| MNT-3 | MNT | major | `add-match-rule` step 3 names a list this PR removed | Fix now → F-9 | Read `SKILL.md:61` |
| MNT-4 | MNT | major | `conclusion-reconciliation.md` per-type table has no boolean row | Fix now → F-9 | Read lines 133-141 |
| SEC-2 | SEC | minor | Single-only held only by Go write paths | Fix now → F-4 | CHECK fits the unreleased `000061.sql` rebuild |
| UX-1 | UX | minor | Failed load leaves create offering `text` alone, unexplained | Fix now → F-7 | D-11 did not cover a failed load, so To user. User chose explain in the dialog and refuse confirm |
| UX-2 | UX | minor | "Yes / no" label is spoken but never seen | Later PR → S14-37 | Goes in the S14-37 map scope entry through F-9; the pill is unchanged code and a board call |
| CON-4 | CON | minor | T-8's `Verify` calls check nothing for the handle | Fix now → F-5 | `Verify` walks loaded nodes only (`verify.go:49`) |
| CON-5 | CON | minor | D-13 erase audit field has no test | Fix now → F-2 | No `value_boolean` in `deleteimpact` tests |
| CON-6 | CON | minor | §10 manual check not run | Rejected | ship-pr step 8 gives the manual check to the user; it stays in the report |
| MNT-5 | MNT | minor | `valueTypeOK` in `Create` cannot fire | Fix now → F-8 | `researcherCreatable` needs a found rule |
| MNT-6 | MNT | minor | `nullInt64`/`NullBool` and `optionalInt64`/`optionalBool` near-copies | Fix now → F-3 | Same bodies. Same-pass cleanup: supersedes the Out line "generic `optional[T]`" |
| MNT-7 | MNT | minor | `acceptsValueType` duplicates a branch; bare `"text"` | Fix now (literal) → F-8; Rejected (deletion) | D-11 names `acceptsValueType`, the view passes its own local draft, and F-7 now gives it a job |
| MNT-8 | MNT | minor | Swift test fixture is an unpinned copy of `propertymeta.Types` | Fix now → F-8 | D-12's reasoning; comment claims Go's table |
| MNT-9 | MNT | minor | S14-37 map entry names a deleted Swift member | Fix now → F-9 | Read map line 568 |

Nits (logged): `UX-3 · plan §10 · picker shows raw lowercase keys` · `UX-4 · L10n.swift:7691 · errorValueType names only term; S14-37 could reword` · `MNT-10 · ReconciledValueDisplay.swift:102 · case .boolean(let value) shadows value` · `MNT-11 · PropertiesModelTests createBeforeTheFirstLoad… · expectation restates the code` · `MNT-12 · contract_test.go:31 · move boolValue to boolean_test.go` · `MNT-13 · L10n.swift:7314 · valueBoolean* keys sit inside valueType* group` · `MNT-14 · swift_vocabulary_test.go · hand list of Surface constants` · `MNT-15 · observations.go:70 · FillListedDisplayText comment untrue for boolean` · `MNT-16 · citations_test.go:301 · audit key list omits value_boolean` · `CON-7 · ReconciledValueDisplayTests.swift:151 · raw Yes/No instead of L10n.string` · `CON-8 · propertymeta_test.go · TestTypes pins only boolean's rule` · `CON-9 · subjectprofile.go:226 · copies survive mutation, unreachable yet` · `CON-10 · properties_test.go · tests not table-driven`

To user (answered 2026-11-06): SEC-1 read as absent, with a CHECK on the cache column; UX-1 explain in the dialog and refuse confirm.

### Fix plan 1

Supersedes, by the same-pass rule: the plan's Out lines "one shared struct…and a leaf package for the Observation row map" and "generic `optional[T]`", and D-2's field names. By the user's answers: D-1 (no CHECK on the cache column) and D-3 (scans read `sql.NullBool`), and D-11 for a failed load. Record each in the plan's Landed note.

Dependency fact: `core/database` imports `graphcache`, so nothing `graphcache` imports may import `core/database`. `core/propertymeta` imports nothing in the repo.

**F-1 — One Scalar value and one column scan, with an out-of-domain boolean read as absent** (from MNT-2, SEC-1)
Files: `core/propertymeta/scalar.go` (new), `core/database/scalarcols/` (new leaf: imports `database/sql` and `propertymeta` only), migration `000061.sql`, `core/autoreconcile`, `core/database/{conclusiondetails,graphcache,graphfill,observations,autoreconciler,deleteimpact,promotealign}`, `core/subjectprofile`, `api/ffi/handlers`, and tests. Skills: `add-catalog-migration` (schema hash).
Failing test first: `core/database/graphcache` or `autoreconciler` `TestOutOfDomainBooleanReadsAbsent`: with `PRAGMA ignore_check_constraints = ON` on one connection, set `auto_reconciler_values.value_boolean = 2` on a promoted handle. `Graph` node load returns no error and the value reads `HasBoolean false`; `RecomputeTx` succeeds and the row holds 0 or 1 again. Second row: set `observations.value_boolean = 2` and `cache_version = 0`; `EnsureCatalog`, the citation Observation list and `conclusiondetails.ForEntity` all return no error.
Do:
- `propertymeta.Scalar{Text string; HasText bool; Integer int64; HasInteger bool; Boolean bool; HasBoolean bool}`, plus constructors `TextScalar(string)`, `IntegerScalar(int64)`, `BooleanScalar(bool)` for literals.
- Embed `propertymeta.Scalar` in all seven structs: `autoreconcile.Value`, `conclusiondetails.Value`, `observations.Observation`, `observations.Input`, `graphcache.Value`, `graphcache.TruthRow`, `graphcache.EvidenceObservation`. `Observation` and `Input` drop `ValueText`/`ValueInteger`/`ValueBoolean` for `Text`/`Integer`/`Boolean`. Leave `valuecompare.Value` alone (no `Has` flags).
- `scalarcols.Columns{Text sql.NullString; Integer, Boolean sql.NullInt64}` with `Scalar() propertymeta.Scalar`: `HasBoolean` true only for 0 or 1. Use it at every site that scans the three columns. If two sites disagree on `HasText` (for example `Valid` against non-empty), keep each site's behavior and report it; do not unify silently.
- `promotealign.decode` takes a `propertymeta.Scalar` in place of its six parameters.
- `000061.sql`: `auto_reconciler_values.value_boolean INTEGER CHECK (value_boolean IN (0, 1))`.
- Update the catalog docs that describe the cache column.
Check: `rg -n 'HasBoolean\s+bool' core api -g '!*_test.go'` prints only `core/propertymeta/scalar.go`. `rg -n 'sql\.NullBool' core api` prints nothing. `grep -rnE 'HasInteger\s+bool' core api --include='*.go'` prints only `scalar.go`. Test passes; scripts/check-pr.sh PASS.

**F-2 — One Observation row for write and erase** (from MNT-1, CON-5)
Files: `core/database/observationrow/` (new; may import `core/database` and `scalarcols`), `core/database/observations/write.go`, `core/database/deleteimpact/facets.go`, and a `deleteimpact` test. Skill: `add-catalog-delete`.
Failing test first: `core/database/deleteimpact` erase-audit case: a connection-facet release erases a boolean Observation with `false`; the audit row's `value_boolean` is `false`, not nil. It passes today, so prove it by nil-ing the value once and seeing it fail.
Do: the package owns the column list (`Columns`), the `Row` struct, `Scan`, and `AuditFields()`. `observations` (`sqlGetRow`, `scanRawObservation`, `observationRowMap`) and `deleteimpact.eraseObservation` move onto it; delete `eraseObservation`'s own SELECT list, scan and map, and `facets.go`'s `integerJSON`/`booleanJSON`. `observations` imports `deleteimpact`, so `observationrow` imports neither. If `nullStringJSON(text)` and `database.NullString(o.ValueText)` differ on any value an existing test holds, stop and report: an audit row's values must not change.
Check: `rg -n '"value_integer":' core -g '!*_test.go'` prints one line, in `observationrow`. scripts/check-pr.sh PASS.

**F-3 — One nullable binder and one optional helper** (from MNT-6)
Files: `core/database/null.go`, `core/database/observations`, `core/database/autoreconciler/autoreconciler.go`, `api/ffi/handlers/observations.go`.
Do: move `nullInt64` to `database.NullInt64` beside `NullBool`; bind integers through it in `insertRow` and the Observation writes. Replace `optionalInt64` and `optionalBool` with one generic `optional[T any](has bool, v T) *T`.
Check: `grep -rnE '^func (nullInt64|optionalInt64|optionalBool)\(' core api` prints nothing. scripts/check-pr.sh PASS.

**F-4 — The table holds single-only for boolean** (from SEC-2)
Files: `core/database/migrations/000061.sql`, `core/database/properties` test.
Failing test first: `TestBooleanCardinalityStaysSingle` row: raw `UPDATE properties SET cardinality = 'multiple'` on a boolean Property fails.
Do: add `CHECK (value_type <> 'boolean' OR cardinality = 'single')` to `properties__new`.
Check: scripts/check-pr.sh PASS.

**F-5 — Tests on the read paths S14-37 uses** (from CON-1, CON-2, CON-4)
Files: `core/database/conclusiondetails/conclusiondetails_test.go`, `api/ffi/handlers/citations_test.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Do: add CON-1's `TestForEntityBoolean` (agree true, agree false, disagree; adjust to F-1's literals). Extend `TestCreateCitationBooleanObservation` to read through the graph-backed citation list RPC (the caller of `listedProtosFromGraph`) and assert `ValueBoolean != nil` and false. In `TestReconciledBoolean`, load the handle into the graph before each `Verify` and assert `HasBoolean` and `Boolean`.
Check: each new assertion fails when the field copy it covers is removed by hand once (revert after); scripts/check-pr.sh PASS.

**F-6 — FakeStore keeps a boolean** (from CON-3)
Files: `macos/App/Platform/FakeStore.swift`, a `macos/ProvenenciaTests` file. Skill: `add-swift-test`.
Failing test first: `fakeStoreKeepsABooleanDraft`: a draft with `valueBoolean = false` comes back `false` from `addObservationsToCitation`.
Do: pass `valueBoolean` at both builders (~3120, ~3194), directly after the integer arguments. `CitationObservationRows.swift:238` stays with S14-37.
Check: test passes; scripts/check-pr.sh PASS.

**F-7 — A failed load explains itself in the create dialog** (from UX-1)
Files: `macos/App/Features/Properties/PropertiesModel.swift`, `PropertiesView.swift`, `macos/ProvenenciaTests/PropertiesModelTests.swift`. Skill: `add-swift-test`.
Failing test first: `createRefusesAfterAFailedLoad`: FakeStore fails `getPropertiesWorkspace` on a cold page; `model.loadError != nil`; `acceptsValueType("text")` is false; `canSubmitCreate` is false and `submitCreate` does not write.
Do: `acceptsValueType` returns false when `snapshot.valueTypeRules.isEmpty && loadError != nil`. In `createForm`, below the `formError` callout, show `PVCallout(tone: .danger, message: L10n.Errors.message(for: loadError))` under the same condition, identifier `properties.create.loadError`. The still-loading path stays as D-11 says and `createBeforeTheFirstLoadFallsBackToTheDraftType` still passes. No new strings.
Check: test passes; scripts/check-pr.sh PASS.

**F-8 — Riding minors** (from MNT-5, MNT-7, MNT-8)
Files: `core/database/properties/properties.go`, `PropertiesModel.swift`, `PropertiesView.swift`, `PropertiesModelTests.swift`.
Do: drop `!valueTypeOK(valueType)` from `Create`. Use `PropertyValueType.text.rawValue` as the one default in place of the bare `"text"` literals this PR added. Replace the Swift fixture with a small neutral one (one creatable, one not, order not alphabetical) and remove the comment claiming it is Go's table.
Check: scripts/check-pr.sh PASS.

**F-9 — Docs** (from MNT-3, MNT-4, MNT-9, UX-2)
Files: `.agents/skills/add-match-rule/SKILL.md`, `docs/conclusion-reconciliation.md`, `docs/deployment-plan/spike-14/deployment-plan.md`, `docs/ideas/structural-debt.md`, the plan's Landed note.
Do:
- Skill step 3: a new type adds a `propertymeta.Types` row; a type with no comparer adds a `noComparer` entry in `TestCompareUnsupported`.
- Reconciliation doc: a `boolean` row (equal; none; the value; —), the D-9 rule that a negative `true` denies `true` and does not create `false`, and a status-table mention.
- S14-37 map entry: point Context at `propertymeta.Types` and `PropertiesSnapshot.creatableValueTypes`. Add to Ships: decide in S14-D8 whether the type pill and picker show the raw key `boolean` or `PropertyValueType.label` ("Yes / no"), because today the label reaches VoiceOver only.
- Structural debt: remove or shrink the S14-36 row to what F-1 and F-2 did not close (term, entity, date, name, image columns).
- Landed note: list every superseded line above.
Check: `rg -n 'PropertiesModel.researcherCreatable|ComparerTypes\(\). list' docs .agents` prints only the plan and review log; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `23d95eb9`. Gate PASS (run by the orchestrator). F-1 to F-9 done. Implementer notes: `HasText` kept per site (`scalarcols` from `Valid`; `scanRawObservation` overrides to non-empty); `nullStringJSON` and `database.NullString` differ only on a stored `""`, which no write path stores and no test holds; FakeStore gained a `getPropertiesWorkspaceError` hook for F-7. §10 not run.

## Round 2 — review
Reviewers: CON, MNT, SEC+PRF (each had a Fix-now blocker or major). Skipped: UX (only a Fix-now minor). UX-1 **Resolved**: orchestrator read `5cb05743...23d95eb9` of `PropertiesModel.swift` and `PropertiesView.swift`; `acceptsValueType` refuses when the rules are empty and the load failed, `createLoadError` drives the dialog callout, and the still-loading path is unchanged.

Round-1 findings: MNT-1 to MNT-4, SEC-1, CON-1 to CON-3 **Resolved** per the reviewers, with mutation evidence for the CON tests. SEC-2, CON-4, CON-5 Resolved per reviewer notes and mutants.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-17 | MNT | blocker | `observationrow.Row` is a second declaration of `observations.Observation`, joined by two hand-written copies | Fix now → F-10 | Reviewer ran an overlay with plain type conversions; a one-sided field compiles today and would record nil in the audit row, which is the drift MNT-1 existed to end |

Noted by reviewers below the round-2 bar, riding on F-11 because Fix plan 2 exists and each is code or docs this PR added: `Columns.Dest()` has no caller (MNT, SEC, CON); F-8's fixture no longer proves AC-4.2's order; `TestTypes` pins only boolean's flags (nit CON-8), and CON's mutant turned date's flag off with all of `./core/... ./api/...` passing; the Landed note misses the fixture change and the FakeStore hook; `completed.md` sentence order. Logged, not fixed: SEC-1's Observation half asserts no error only; `fakeObservation`'s `valueBoolean` has no test of its own; stored `""` and a stored `2` audit as null on crafted rows only.

### Fix plan 2

**F-10 — One Observation row type** (from MNT-17)
Files: `core/database/observations/observations.go`, `core/database/observations/write.go`, `core/database/observationrow/observationrow.go`.
Do: replace the `Observation` struct with `type Observation = observationrow.Row`, moving its field comments onto `Row`. `scanRawObservation` drops its field-by-field copy and keeps the `HasText` override. `observationRowMap` goes; callers use `o.AuditFields()`. If `Observation` turns out to need a method of its own, use conversions (`Observation(r)`, `observationrow.Row(o)`) instead, so a one-sided field is a compile error.
Check: `rg -n 'observationrow\.Row\{|Observation\{\s*$' core/database/observations -g '!*_test.go'` shows no field-by-field copy between the two; scripts/check-pr.sh PASS.

**F-11 — Riding leftovers**
Files: `core/database/scalarcols/scalarcols.go`, `core/propertymeta/propertymeta_test.go`, `macos/ProvenenciaTests/PropertiesModelTests.swift`, the plan's Landed note, `docs/deployment-plan/spike-14/completed.md`.
Do: delete `Columns.Dest()`. Make `TestTypes` check every row's `ResearcherCreatable` and `SingleOnly`, table-driven. Give the neutral Swift fixture two creatable types in non-alphabetical order plus one not creatable, and assert the order, so `createOffersGoCreatableTypes` proves AC-4.2 again. Add the fixture change and FakeStore's `getPropertiesWorkspaceError` to the Landed note. In `completed.md`, say "a boolean Property" where "one" now misreads, and say the only visible change is the failed-load notice in the create dialog.
Check: `rg -n 'Dest\(\)' core/database/scalarcols` prints nothing; turning off any one flag in `propertymeta.Types` by hand fails `TestTypes` (revert after); scripts/check-pr.sh PASS.

## Round 2 — fix
Head `7ccfb100`. Gate PASS (run by the orchestrator). F-10 and F-11 done. Fix plan 2 existed for MNT-17; no third review round, per the cap.

## Result
Gate PASS · main unchanged since the branch was cut · open: none · follow-ups: none from review (the structural-debt row was shrunk in F-9) · to user: SEC-1 and UX-1, both answered.
