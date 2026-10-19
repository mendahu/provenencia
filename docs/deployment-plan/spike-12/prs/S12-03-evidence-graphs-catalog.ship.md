# Ship log — S12-03 Evidence graphs in the catalog

**Plan:** [`S12-03-evidence-graphs-catalog.md`](S12-03-evidence-graphs-catalog.md) · **Branch:** `s12-03-evidence-graphs-catalog` · **PR:** #498
**State:** ready

Worktree: `/Users/jakerobins/Repos/provenencia-s12-03`, branched from `origin/main` @ `84a9b318`.

## Round 0 — implement
Head `686c8e29`. Gate PASS (orchestrator run). Report: no deviations from Decisions; test details the plan left open (evidencegraphs tests strip the create-time links and link explicitly; T-15/T-16 use seeded `birth_certificate`; T-19 builds its second Source in the transaction). AC-1.1 migration backfill and §10 manual check not run by the implementer.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`core/apperr`, new error codes), SEC+PRF (security: `migrations/000055.sql`, `core/catalogsession`; performance: `core/database/**`). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 (+ SEC outside-lens note to CON) | PRF | major | `evidencegraphs.EnsureRefs` backfills one ref per Source with no collision retry and one autocommit per row; first open after upgrade can fail | To user → user 2026-10-19: override D-15, fix here → F-1 | verified: `evidencegraphs.go:361-` loop has no `IsUniqueConflict` branch and no transaction; 32^5 token space gives ~1.5% collision chance at 1,000 Sources, ~31% at 5,000. Contradicts D-15 "no retry (as `users`)" |
| MNT-3 | MNT | major | `add-catalog-ref` skill step 3 does not name `ReconcileOnOpen` as the home for open-time backfills | Fix now → F-2 | verified: `.agents/skills/add-catalog-ref/SKILL.md:74,88` name only `users.EnsureRefs`; `docs/catalog-refs.md:112` already says `ReconcileOnOpen` |
| MNT-1 | MNT | major | `EvidenceGraphs` is the third catalog-wide bool on `effects.Set`/`Effect` | Follow-up | plan Out line and working rule "The graph list has its own effect" chose the bool; already logged in `docs/ideas/shared-list-helpers.md#catalog-wide-list-flags`. Best landed before S12-05 puts it on the wire |
| MNT-2 | MNT | major | Seventh copy of the ref-mint-and-insert loop; more private `uuidJSON`/`nullStr` | Follow-up | D-8 and plan Out line chose the copy; already logged in `docs/ideas/shared-list-helpers.md#one-ref-mint-loop` |
| UX-1 | UX | minor | `subjects.graph_id` inbound edge has no `viaFormats` delete-confirm heading | Follow-up | §6 Strings: None; no macOS in this PR. Put the line in the S12-04/S12-05 plan Strings |
| UX-2 | UX | minor | Three new `evidencegraphs.*` codes have no `L10n.Errors` mapping | Follow-up | D-17 defers; unreachable at this head. S12-05 DoD should check each code |
| SEC-1 | SEC | minor | `PRAGMA foreign_key_check` result rows are discarded by `migrate` | Follow-up | mechanism is `migrate.go` (unchanged here); same gap in 000030/000052. `docs/ideas/` |
| MNT-4 | MNT | minor | `seedGraphRefless`/`graphRef` fixture copied into two test packages | Follow-up | T-15/T-16 are named plan tests; files not in the fix plan |
| MNT-5 | MNT | minor | `effects.Set` doc comment lists an incomplete set of catalog-wide flags | Follow-up | `registry.go:28`; file not in the fix plan (rides with MNT-1's follow-up) |

Nits (logged): `CON-1 · effects/registry_completeness_test.go:3-6 · regroup stdlib imports` · `MNT-6 · evidencegraphs.go:316-322 · LinkArtifact's 16-byte check repeats checkLink` · `MNT-7 · catalogsession/session.go:102-105 · comment says "Create"; name onboarding.createCatalog` · `MNT-8 · defaultGraphID copied into four test files; move onto S12-04's List` · `UX-3 · evidencegraphs.go:18 · DefaultName "Default" is English catalog data; note for S12-11/S12-14`

CON: clean. UX: clean (verdict). SEC+PRF, MNT: changes needed.

**Paused:** PRF-1 to user (contradicts D-15). User 2026-10-19: "Override D-15 for PRF-1: run the graph-ref backfill in one transaction with the maxRefRetries / IsUniqueConflict retry, plus a test that forces a collision."

### Fix plan 1

**F-1 — Backfill graph refs in one transaction, retrying a ref collision** (from PRF-1; user override of D-15's "no retry")
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`, plan D-15 Landed note / deviation line in `S12-03-evidence-graphs-catalog.md`.
Failing test first: `TestEnsureRefs/retries_a_ref_collision` — seed two Sources (graphs A and B); set B's `ref` NULL and keep A's; make the package's mint seam return A's existing ref on its first call and a fresh `ref.Mint` after; after `EnsureRefs`, B has a valid `GPH-` ref ≠ A's and A is unchanged. Red today: the UPDATE returns a UNIQUE error.
Do: add an unexported seam `var mintRef = func() (string, error) { return ref.Mint(ref.PrefixEvidenceGraph) }` and use it in both `insert` and `EnsureRefs` (tests restore it with `t.Cleanup`). Rewrite `EnsureRefs`: read the NULL-ref ids as today, then `db.Begin()`; for each id loop `attempt < maxRefRetries` — `mintRef()`, `tx.Exec(sqlSetRef, r, id)`, break on nil, continue on `database.IsUniqueConflict(err)`, roll back and return on any other error; after the loop, a still-failing row rolls back and returns `ErrInvalid` (as `insert` does); `Commit()` at the end. Same retry shape as `insert` (lines ~80–95); do not change `users.EnsureRefs`. Update D-15's doc line in the plan's Landed note as a deviation ("backfill runs in one transaction with the `insert` retry, per user override of D-15 in review round 1"), and any doc sentence this PR added that says the graph backfill has no retry (`grep -rn "no retry" docs .agents core/database/evidencegraphs`).
Check: `TestEnsureRefs` (all cases) passes; scripts/check-pr.sh PASS.

**F-2 — Name `ReconcileOnOpen` in the `add-catalog-ref` skill** (from MNT-3)
Files: `.agents/skills/add-catalog-ref/SKILL.md`.
Failing test first: none (docs).
Do: change checklist step 3 to: "**Backfill** — if opening old catalogs, `EnsureRefs(c)` in the table package (see `users.EnsureRefs`, `evidencegraphs.EnsureRefs`), mint in one transaction with the `maxRefRetries` / `IsUniqueConflict` retry when the table grows with the catalog, and call it once from `catalogsession.ReconcileOnOpen`; never from an opener directly." Add `evidencegraphs.EnsureRefs` and `catalogsession.ReconcileOnOpen` to the Reference implementation list beside `users.EnsureRefs`.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `e9cd7580`. Gate PASS (orchestrator run; `TestEnsureRefs` re-run uncached, 3/3 pass). F-1 done (`612cfd39`: `mintRef` seam, one-transaction backfill with retry; extra subtest `exhausted_retries_roll_back`; plan Landed note records the D-15 deviation). F-2 done (`e9cd7580`).

## Round 2 — review
Reviewers: SEC+PRF (PRF-1), MNT (MNT-3). Scope `686c8e29...e9cd7580`. Not sent: CON, UX (no Fix-now blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | (round 1) graph-ref backfill retry and one transaction | Resolved | `evidencegraphs.go:365-389`: one `db.Begin`, `maxRefRetries` loop on `IsUniqueConflict`, `ErrInvalid` + rollback on exhaustion; `TestEnsureRefs` 3/3 |
| MNT-3 | MNT | major | (round 1) skill names `ReconcileOnOpen` | Resolved | `add-catalog-ref/SKILL.md:74,88-89` match `session.go:49-55` and `docs/catalog-refs.md:112` |

New blockers: none. MNT note (not raised): the mint-and-retry loop now appears twice in `evidencegraphs.go` (`insert`, `EnsureRefs`); the "One ref mint loop" follow-up (MNT-2) should cover the backfill path too.

## Round 2 — user re-decision (2026-10-19)
User: "For MNT1 and MNT2, let's just implement those on this PR … MNT4 and MNT5, same as 1 and 2. … UX1 and UX2, you should write those into the scope for those PRs." `ship-pr` now carries this as the **Same-pass cleanups** rule and the **Later PR** decision (`7029cef8`).

| ID | New decision | Where |
| --- | --- | --- |
| MNT-1 | Fix now → F-3 (supersedes plan Out line "One set of catalog-wide list flags…" and D-5's bool shape) | |
| MNT-2 (+ round-2 MNT note: backfill copy) | Fix now → F-4 (supersedes Out line "One ref-mint-and-insert helper and one `uuidJSON`" and D-8's copy) | |
| MNT-4 | Fix now → F-5 | |
| MNT-5 | Fix now → folded into F-3 (the comment is rewritten with the field) | |
| UX-1, UX-2 | Later PR | S12-05 map entry, Ships and Check (`7543f40e`) |
| MNT-8 (nit) | Later PR | S12-04 map entry (`defaultGraphID` copies move onto `List`) |
| UX-3 (nit) | Later PR | S12-11 map entry (the `Default` label) |
| CON-1, MNT-6, MNT-7 (nits) | Fix now, ride along (files in this fix plan) | F-3, F-4, F-5 |
| SEC-1 | Follow-up (unchanged: `migrate.go`, no copy this PR touches) | |

### Fix plan 2

**F-3 — One declared set of catalog-wide list keys, on `effects`, the wire and Swift** (from MNT-1, MNT-5, CON-1)
Files: `core/database/effects/{registry.go,resolve.go,README.md}` and its tests (`registry_completeness_test.go`, `evidence_graphs_test.go`, any test reading `Set.Vocabulary` / `Set.PropertyUsage` / `Set.EvidenceGraphs`), `core/database/graphcache/commit.go`, `api/proto/engine.proto` + regenerated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (`scripts/generate-proto.sh`), `api/ffi/handlers/write_effects.go` (+ its test), `macos/App/Features/Workspace/Session/{CatalogMutation.swift,CatalogQueryRegistry.swift}`, `macos/App/Platform/{GoStore.swift,FakeStore.swift}`, the Swift tests that build `CatalogWriteEffects(propertyUsage:/vocabulary:)`. Skills: `add-ffi-handler`, `add-swift-test`.
Failing test first: `effects.TestResolveLists` *(new)* — a batch with a `subject_types` change and an `evidence_graphs` change resolves `Set.Lists == [ListVocabulary, ListEvidenceGraphs]`; an `observations` update that does not write `property_id` adds no list, one that does adds `ListPropertyUsage`. Red: `Set.Lists` does not exist (compile failure counts).
Do:
- In `effects`: `type List string` with `ListVocabulary = "vocabulary"`, `ListPropertyUsage = "property_usage"`, `ListEvidenceGraphs = "evidence_graphs"`, and one declared `var Lists = []List{…}` in that order. On `Effect`, replace `Vocabulary`, `PropertyUsage`, `EvidenceGraphs bool` with `Lists []List` and `ListsOnField string` ("when set, an update moves `Lists` only if it writes this field; creates and deletes always do"); `observations` gets `Lists: []List{ListPropertyUsage}, ListsOnField: "property_id"` (today's `usesProperty` rule, now general: rename it `movesLists`). On `Set`, replace the three bools with `Lists []List` (deduped, in declared order) and `func (s Set) Has(l List) bool`. `Resolve` merges lists in one place. `validate`: a `None` entry has no `Lists` (this also closes the drift where `PropertyUsage` was not checked), every listed key is declared, `ListsOnField` needs `Lists`. Rewrite the `Set` doc comment: `Lists` and `Structure` are the catalog-wide fields (MNT-5). `Structure` stays its own field (it is a graph-cache drop, not a list).
- `graphcache.OnCommit`: `set.Has(effects.ListVocabulary)`.
- Wire: in `WriteEffects`, remove `bool vocabulary = 4` and `bool property_usage = 12`, add `reserved 4, 12;` and `repeated string lists = 13;` ("catalog-wide lists this write changed: `effects.List` values"). `write_effects.go` copies `set.Lists` as strings. `structure` stays.
- Swift: `enum CatalogList: String, Sendable { case vocabulary; case propertyUsage = "property_usage"; case evidenceGraphs = "evidence_graphs" }`; `CatalogWriteEffects` replaces `propertyUsage` / `vocabulary` with `var lists: Set<CatalogList> = []`; `GoStore.mapWriteEffects` decodes with `compactMap(CatalogList.init(rawValue:))` (unknown keys ignored). `CatalogQueryRegistry` gets one mapping `static func invalidations(for list: CatalogList, project: ProjectKey) -> [CatalogQueryInvalidation]` (`vocabulary`, `propertyUsage` → `propertiesWorkspace`; `evidenceGraphs` → `[]` with a comment that S12-05 adds the graph list key), and `invalidations(for:project:)` loops `effects.lists` through it, keeping today's dedupe and `structure` → `propertiesWorkspace`. `FakeStore` writes `lists: [.propertyUsage]`.
- If `api/ffi/swift_vocabulary_test.go`'s pair reader can read a Swift enum's raw values, add a pair that checks `CatalogList` against `effects.Lists`; otherwise say so in the report.
- Ride along: regroup the imports in `registry_completeness_test.go` (CON-1).
- Docs: `effects/README.md`; working rule "The graph list has its own effect" in `docs/deployment-plan/spike-12/deployment-plan.md` (S12-03 sets list key `evidence_graphs`; S12-05 adds one `CatalogQueryRegistry` list-key line); remove the "Catalog-wide list flags" section from `docs/ideas/shared-list-helpers.md`; the plan's Landed note records that review superseded the Out line and D-5's bool. `grep -rn "Set.Vocabulary\|Set.PropertyUsage\|Set.EvidenceGraphs\|propertyUsage\b" docs .agents` and fix what is now false.
Check: `TestResolveLists` passes; the existing effects, graphcache, write-effects handler and Swift `CatalogQueryRegistry` tests pass unchanged in meaning; scripts/check-pr.sh PASS.

**F-4 — One ref-mint-and-insert helper and one home for the UUID and empty-string binders** (from MNT-2, the round-2 MNT note, MNT-6)
Files: new `core/database/refinsert.go` and `core/database/uuid.go` (+ tests); `core/database/nullblob.go`; every package below. Skill: `add-catalog-ref`.
Failing test first: `database.TestInsertWithRef` *(new)* — an insert func that returns a UNIQUE conflict twice then nil returns the third minted ref; one that always conflicts returns `database.ErrRefsExhausted` after `MaxRefRetries` calls; a non-conflict error returns at once, unchanged. Red: the helper does not exist.
Do:
- `core/database`: `const MaxRefRetries = 8`; `var ErrRefsExhausted = errors.New("database: ref retries exhausted")`; `func Minter(prefix string) func() (string, error)` (wraps `ref.Mint(prefix)`); `func InsertWithRef(mint func() (string, error), insert func(ref string) error) (string, error)` — mint, insert, retry on `IsUniqueConflict`, return a mint error or a non-conflict error at once. Move every loop onto it: `sources`, `artifacts`, `citations`, `observations`, `subjects`, `canonicalentities`, and `evidencegraphs` (`insert` and the `EnsureRefs` backfill; its `mintRef` seam becomes `var mintRef = database.Minter(ref.PrefixEvidenceGraph)`). Each caller keeps exactly today's error mapping: `errors.Is(err, database.ErrRefsExhausted)` → that package's current exhaustion error, anything else → its current mapping. Delete each package's `maxRefRetries`.
- `core/database/uuid.go`: `func UUIDString(id []byte) string` ("" unless a 16-byte UUID) and `func UUIDJSON(id []byte) any` (nil unless a 16-byte UUID). Move every private `uuidString` and `uuidJSON` in `core/database/**` (non-test) onto them: `artifacts`, `canonicalentities`, `citations`, `deleteimpact` (`facets.go`, `kinds.go`), `evidencegraphs`, `identityclaims`, `metadatafields`, `observations`, `properties`, `propertyterms`, `searchindex`, `sourcecredibility`, `sourcemetadata`, `sources`, `sourcetypes`, `subjectpositions`, `subjects`. If a copy behaves differently for a non-16-byte id and a test depends on it, keep the test's meaning and say which in the report.
- `NullString(s string) any` (nil when empty) beside `NullBlob`; rename `nullblob.go` to `null.go`. Move every private `nullStr` / `nullIfEmpty` with that body onto it: `artifacts`, `audit`, `canonicalentities`, `citations`, `datevalues`, `evidencegraphs`, `files`, `identityclaims`, `namevalues`, `observations`, `sourcecredibility`, `sourcemetadata`, `sources`, `subjects`. Leave `deleteimpact.nullStringJSON` (different input).
- Ride along: in `evidencegraphs.LinkArtifact`, drop the 16-byte check that `checkLink` repeats and keep only the repeated-id check (MNT-6).
- Docs: `add-catalog-ref` skill (the mint loop is `database.InsertWithRef`); remove "One ref mint loop" from `docs/ideas/shared-list-helpers.md`; the plan's Landed note records that review superseded the Out line and D-8's copy.
Check: `TestInsertWithRef` passes; every existing package test passes (`TestEnsureRefs` 3/3 included); `grep -rn "maxRefRetries\|^func uuidJSON\|^func uuidString\|^func nullStr\|^func nullIfEmpty" core --include='*.go' | grep -v _test` prints nothing; scripts/check-pr.sh PASS.

**F-5 — One graph-ref fixture, one test of the repair list** (from MNT-4, MNT-7)
Files: new `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` (the `observationstest` pattern), `core/catalogsession/session_test.go`, `core/onboarding/ready_test.go`, `core/catalogsession/session.go` (comment only).
Failing test first: none (test-only move); the moved tests must still go red if `ReconcileOnOpen` drops `evidencegraphs.EnsureRefs` — check that once by commenting the line out locally, and say so in the report.
Do: move `seedGraphRefless` and `graphRef` into `evidencegraphstest` as exported helpers (`SeedRefless(t, c)`, `Ref(t, c, id)`); both test files call them and their private copies go. Keep T-15 and T-16 (they check two openers' wiring). Ride along: the `session.go` comment names `onboarding.createCatalog` instead of "Create" (MNT-7).
Check: `TestDo`, `TestOpenCatalogMintsGraphRefs` pass; scripts/check-pr.sh PASS.

## Round 2 — fix (Fix plan 2)
Head `dee13d99`. Gate PASS (orchestrator run; uncached `go test ./core/... ./api/...` all ok). F-3 done (`de0b4411`; added a `CatalogList` ↔ `effects.Lists` pair in `swift_vocabulary_test.go` and `TestWriteEffectsProtoLists`), F-4 done (`51edb1d1`), F-5 done (`dee13d99`; mutation check: both tests fail with the `evidencegraphs.EnsureRefs` line removed). Deviations: `TestResolveLists` uses `source_type` for the vocabulary key (`subject_types` has no `Entity`); `datevalues.nullIfEmpty` and `audit.nullIfEmpty` trim whitespace, so they stay; old `artifacts.uuidJSON` returned `""` for a malformed non-empty id, shared `UUIDJSON` returns nil (no test depended on it); `engine.pb.swift` regenerated with `protoc-gen-swift` 1.38.1, matching the resolved SwiftProtobuf 1.38.1 runtime, which drops the old `_StorageClass` shape (~300 lines of churn; accepted); `sources/notes.go` gofmt'd.

## Round 3 — review (user-requested scope)
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`, `core/database/**`). Scope `2d4208e6...dee13d99`. Beyond the two-round cap because the user added this work to the PR; blockers only.

UX: clean. SEC+PRF: clean (PRF-1 re-checked, still resolved). CON: clean, 1 minor. MNT: changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-9 | MNT | blocker | 15 renamed private copies of `NullString` / `NullBlob` (`nullJSON`, `emptyAsNil`, `nullBlob`, `nullBytes`) | Fix now → F-6 | verified: `grep -rnE "^func \w+\((s\|b) (string\|\[\]byte)\) any \{" core` lists them; every package is in this PR's diff. F-4 searched by name, not by body |
| MNT-10 | MNT | blocker | `api/ffi/handlers.uuidString` copies `database.UUIDString` | Fix now → F-7 | verified: `handlers/catalog.go:41`; package already imports `core/database` and is in this PR's diff (`write_effects.go`) |
| MNT-11 | MNT | major | `add-catalog-model` skill lists the removed `Vocabulary` / `PropertyUsage` / `EvidenceGraphs` fields | Fix now → F-8 | verified: `.agents/skills/add-catalog-model/SKILL.md:26` |
| MNT-12 | MNT | major | `sources.isUniqueConflict` is dead after F-4 | Fix now → F-6 | verified: `sources.go:516`, no caller |
| CON-R2-1 | CON | minor | No Swift test for each `CatalogList` key, `structure`, and dedupe in the rewritten invalidation loop | Fix now → F-9 | code this PR rewrote (`CatalogQueryRegistry.swift`); old tests never covered vocabulary/structure |
| MNT-13 | MNT | nit→minor | `datevalues` and `audit` trimming `nullIfEmpty` are copies of each other | Follow-up | neither package is in this PR's diff (same-pass rule: copies untouched → follow-up); log in `docs/ideas/shared-list-helpers.md` |

Nits (ride along, files in this fix plan): `CON-R2-2 · registry_completeness_test.go:160 · "tables with EvidenceGraphs" message; TestEvidenceGraphsFlag / TestPropertyUsage names → list wording` · `MNT-14 · session_test.go:10, ready_test.go:8 · fold evidencegraphstest import into the module group` · `MNT-15 · plan Landed note (3) says maxRefRetries` · `MNT-16 · add-catalog-ref step 3: say users.EnsureRefs is the small-table, no-retry exception`

### Fix plan 3

**F-6 — Every empty-value binder goes through `database.NullString` / `NullBlob`** (from MNT-9, MNT-12)
Files: `core/database/{artifacts,sources,sourcemetadata,propertyterms,properties,metadatafields,sourcetypes,sourcecredibility,subjects,citations,observations,identityclaims}`.
Failing test first: none (mechanical; existing package tests cover the bindings).
Do: replace every call of `nullJSON` and `emptyAsNil` with `database.NullString`, and of `nullBlob` and `nullBytes` with `database.NullBlob`; delete the private funcs. Delete `sources.isUniqueConflict` (no caller). Keep `datevalues.nullIfEmpty` and `audit.nullIfEmpty` (they trim). Search by body, not name: `grep -rnE "^func \w+\((s|b) (string|\[\]byte)\) any \{" core api --include='*.go' | grep -v _test` must print only `database/null.go` (two lines), `datevalues.go` and `audit.go`.
Check: that grep; `go test ./core/...` passes; scripts/check-pr.sh PASS.

**F-7 — FFI handlers use `database.UUIDString`** (from MNT-10)
Files: `api/ffi/handlers/*.go` (non-test and test where they call it).
Failing test first: none (mechanical rename).
Do: replace every `uuidString(` in `api/ffi/handlers` with `database.UUIDString(`, delete `handlers.uuidString` (`catalog.go:41`). Then `grep -rnE "^func \w+\(\w+ \[\]byte\) (string|any) \{" core api --include='*.go' | grep -v _test` must print only `core/database/uuid.go` (two lines) and `null.go`'s `NullBlob`; report any other hit and move it too if its body is the same.
Check: `go test ./api/...` passes; scripts/check-pr.sh PASS.

**F-8 — Skills and plan say what the code now does** (from MNT-11, MNT-15, MNT-16)
Files: `.agents/skills/add-catalog-model/SKILL.md`, `.agents/skills/add-catalog-ref/SKILL.md`, `docs/deployment-plan/spike-12/prs/S12-03-evidence-graphs-catalog.md` (Landed note), `docs/ideas/shared-list-helpers.md`.
Do: `add-catalog-model` step 6 lists `Effect`'s fields as they are now (`Lists`, `ListsOnField` in place of the three bools) and points to `effects/README.md` for adding a catalog-wide list. `add-catalog-ref` step 3 says `users.EnsureRefs` is the small-table exception with no retry; tables that grow with the catalog backfill through `database.InsertWithRef`. Landed note (3): `database.MaxRefRetries` through `InsertWithRef`, not `maxRefRetries`; note (2) lists the renamed binders F-6 removed. Add a short "Trimmed empty-string binder" entry to the ideas file for MNT-13. `grep -rn "Vocabulary bool\|PropertyUsage\b\|EvidenceGraphs bool\|maxRefRetries\|nullJSON\|emptyAsNil" .agents docs` and fix what is false (leave historical ship/review logs).
Check: scripts/check-pr.sh PASS.

**F-9 — One Swift table test over the list keys** (from CON-R2-1, CON-R2-2, MNT-14)
Files: `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`; ride-along: `core/database/effects/registry_completeness_test.go` and the effects tests named in CON-R2-2; `core/catalogsession/session_test.go`, `core/onboarding/ready_test.go`. Skill: `add-swift-test`.
Failing test first: `CatalogQueryRegistryTests` `listInvalidations` *(new, parameterized)* — `[.vocabulary]`, `[.propertyUsage]`, `[.vocabulary, .propertyUsage]`, `structure: true` alone, and `[.vocabulary]` + `structure: true` each give exactly one `.propertiesWorkspace` invalidation; `[.evidenceGraphs]` alone gives none. Show it red once by breaking the dedupe locally, then restore.
Do: add the test. Ride along: reword the "tables with EvidenceGraphs" failure message and rename `TestEvidenceGraphsFlag` / `TestPropertyUsage` to list wording (`TestEvidenceGraphsList`, `TestPropertyUsageList`); fold the `evidencegraphstest` import into the module import group in both test files.
Check: the new test passes; scripts/check-pr.sh PASS.

## Round 3 — fix (Fix plan 3)
Head `283874b4`. Gate PASS (orchestrator run; uncached `go test ./core/... ./api/...` ok; body grep prints only `null.go` ×2, `datevalues.nullIfEmpty`, `audit.nullIfEmpty`). F-6 `9a36980e`, F-7 `b48b9976`, F-8 `5e8c5197`, F-9 `283874b4` (red once with the dedupe removed). CON-R2-1: **Resolved** (orchestrator read: `listInvalidations`, 6 cases). Skill lesson `9dd6fda3`: same-pass copies are found by body, not name.

## Round 4 — review (MNT, scope `3e63ac29...283874b4`)
MNT-9, MNT-10, MNT-11, MNT-12: Resolved.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-17 | MNT | blocker | 13 inline `var desc any; if x == "" { desc = nil }` copies of `NullString` | Fix now → F-10 | verified: `grep -rn "desc = nil" core` → 13 (`metadatafields`, `properties`, `propertyterms`, `sourcetypes` ×3 each, in this PR's diff; `subjecttypes` ×1, same copy) |

### Fix plan 4
**F-10 — Inline description binders use `database.NullString`** (from MNT-17)
Files: `core/database/{metadatafields,properties,propertyterms,sourcetypes,subjecttypes}`.
Do: at each `var desc any` / `if … == "" { desc = nil } else { desc = … }` block, delete the block and pass `database.NullString(<x>)` in the `Exec`. Leave `propertyterms`' trimmed `InverseKey`, `audit.go:55`, `observations/write.go:257` (different rules).
Check: `grep -rn "desc = nil" core --include='*.go'` prints nothing; `go test ./core/...` passes; scripts/check-pr.sh PASS. Orchestrator verifies; no further review round.

## Round 4 — fix (Fix plan 4)
Head `df4e2cd7`. Gate PASS (orchestrator run; uncached `go test ./core/... ./api/...` ok). F-10 done; `grep -rn "desc = nil" core` prints nothing. MNT-17: **Resolved** (orchestrator read of the diff: 13 blocks → `database.NullString`, 5 files, −91/+13).

## Result
Gate PASS · CI pending · open: none · fixed on this PR: PRF-1, MNT-1, MNT-2, MNT-3, MNT-4, MNT-5, MNT-9, MNT-10, MNT-11, MNT-12, MNT-17, CON-R2-1 · later PR: UX-1, UX-2 (S12-05), MNT-8 (S12-04), UX-3 (S12-11) · follow-ups: SEC-1, MNT-13 · to user: PRF-1 (override D-15), MNT-1/2/4/5 (fix here) · CON miss: none · rounds: 4 review (rounds 3–4 user-requested)
