# Ship log — S14-17 Vocabulary writes return their effects; dead RPCs go

**Plan:** [`S14-17-vocabulary-effects.md`](S14-17-vocabulary-effects.md) · **Branch:** `s14-17-vocabulary-effects` (stacked on `s14-27-lists-through-descriptors`, PR #693) · **PR:** #694
**State:** ready (CI waits on retarget to main)

## Round 0 — implement
Base `s14-27-lists-through-descriptors` @ `f1b3fe65`. The plan was refreshed against that commit and marked Planned in `f946dd72` on this branch (plan review round 2: M-9, M-10, X-1–X-5 Fixed). The PR's base is S14-27's branch until #693 merges, then `main`.
Head `39bbe2a0`. PR #694 (draft, base `s14-27-lists-through-descriptors`). Gate PASS (orchestrator run, `--base s14-27-lists-through-descriptors` → against `f1b3fe65`; xcode-test 50s). Report deviations:
- T-1's term writes use the seeded `event_type` Property, because a researcher cannot create a term-valued Property.
- The plan's Go core table says "either family", but `OnCommit` follows D-5 (property only); D-5 binds.
- New `FakeStore.assignSubjectTypePropertyErrors`. T-9 binds one failing type, because `bindTypeIDs` is a `Set` and a two-type version was flaky.
- T-9 and T-10 assert that the reload happened, not exact fetch counts.
- The `subjectvocab_test` case is renamed "registry presentation locked connect rules".
- `connect_test` uses one `propertyIDByKey` call plus one whole-list walk.
- `onSettled` (`QueryHandleSettled.swift`) is left with no caller.
- The review log's trailing blank line is trimmed.
- `engine.proto` comment wording.
- Manual check §10 not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`macos/App/**`), SEC+PRF (`api/ffi/**`; `core/database/**`; cache and registry code; `Features/**` models). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `artifacts.HasPrimaryFile` / `sqlExistsByFile` lose their last caller | Fix now → F-1 | verified: `rg HasPrimaryFile core api` → only `artifacts.go:399-409`; its one caller was the deleted `EnsureFileThumbnail` |
| MNT-2 | MNT | major | `View.onSettled` has no caller; doc still names it | Fix now → F-2 | verified: `rg onSettled macos` → only `QueryHandleSettled.swift:8`; `macos-client-patterns.md:83` |
| MNT-3 | MNT | major | 16 of 21 descriptors carry an `invalidation` closure that cannot run | Fix now → F-3 | verified: `rg -c 'invalidateOn: \[\]'` = 16 (7 at base); `invalidations(by:)` (:88-94) calls the closure only behind `invalidateOn.contains`; `make` requires one. Same-pass cleanup: this PR adds a dead closure (`propertyTerms`) and empties nine more. Supersedes D-6's literal `invalidation: { _, _ in .allCached(.propertyTerms) }` (no behavior change) |
| MNT-5 | MNT | minor | Conclusion kinds' Property-label refresh is now seven hand copies | Fix now → F-4 | verified: eight `list(.propertyVocabulary)` entries; `conclusionTriggers` was one shared value. Same-pass cleanup |
| MNT-4 | MNT | minor | Two copies of "binding ids of one type from workspace groups" | Fix now → F-5 | verified: `vocabulary_write_effects_test.go:85-96`, `subject_defs_test.go:199-211`, both added here |
| UX-1 | UX | minor | Source types / Metadata header reads "0 … · 0 seeded" while loading or beside a load error | Fix now → F-6 | verified: `countLine` is non-optional over `types` = `handle.value ?? []` (`SourceTypesModel.swift:82,227`; `MetadataModel.swift:63,141`); before this PR the recount's totals filled that window |
| MNT-6, MNT-8, UX-3 | MNT, UX | minor / nit | Stale comments: `CatalogCounts` "refresh or publish"; `CatalogMutation` "They carry the field's id"; Properties create "leaves the created Property shown" | Fix now → F-7 | verified: `CatalogCounts.swift:10`; `CatalogMutation.swift:17-19`; `PropertiesModel.swift:475`. Stale comments in files this PR changes |
| PRF-1 (+ UX outside-lens) | PRF | minor | A term create reports `property_vocabulary`, so the composer reloads the on-screen Properties workspace (per-Property Observation count) and marks or evicts hidden Conclusion keys; the Properties create flow loads the workspace N+1 times | Follow-up | measured by PRF: the count query takes 2.0 ms at Scale (92,000 Observations), 7.8 ms at 4× Scale; off the main thread; the composer's Observation save already pays it through `property_usage`. Plan Reach / Scale accepted this reach. Narrowing it (a term create reports no list; or a stored `usedBy`) revisits the user's two-list answer → follow-up list |
| MNT-7 | MNT | minor | Field relabel/delete still reaches Source pages through a Swift mutation beside its effects | Follow-up | D-6 / D-13 chose this; pre-existing shape shared with the Source-row mutations; engine-side Source scope for field writes is its own change |
| UX-2 | UX | minor | A failed bind after Property create leaves a dialog whose retry fails with duplicate key | Follow-up | predates this PR (`PropertiesModel.swift:466-502`); this PR makes the created row visible, not worse |

CON: no findings. SEC: no findings.
Nits (logged): `MNT-9 · harness_test.go:125-153 · propertyIDByKey and subjectTypeIDByKey each make their own GetPropertiesWorkspace call` · (MNT-8 and UX-3 ride F-7)

### Fix plan 1
**F-1 — Delete the callee `EnsureFileThumbnail` left behind** (from MNT-1)
Files: `core/database/artifacts/artifacts.go`.
Do: delete `HasPrimaryFile` and the `sqlExistsByFile` constant, as D-11 deletes `PlaceableTypes` and `subjectpositions.Clear`.
Check: `rg -n 'HasPrimaryFile|sqlExistsByFile' core api macos` prints nothing; scripts/check-pr.sh PASS.

**F-2 — Delete `onSettled`** (from MNT-2)
Files: `macos/App/Features/Workspace/QueryHandleSettled.swift` (delete the file; remove it from `macos/Provenencia.xcodeproj/project.pbxproj` if the project lists files), `docs/macos-client-patterns.md:83`.
Do: delete the file. In the `QueryHandle` row, describe `isSettled` ("ready, not refetching and not stale") without "read by `.onSettled`". `QueryHandle.isSettled` stays (tests wait on it).
Check: `rg -n onSettled macos docs/*.md .agents` prints nothing; scripts/check-pr.sh PASS.

**F-3 — A descriptor that names no mutation carries no `invalidation`** (from MNT-3; supersedes D-6's literal `propertyTerms` closure)
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` (and the `CatalogQueryDescriptor` struct wherever it is declared); `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` only if it builds descriptors by hand. Skill: `add-workspace-place`.
Do: in `CatalogQueryDescriptor.make`, give `invalidateOn` the default `[]` and make `invalidation` an optional closure defaulting to `nil`. Add `precondition(invalidateOn.isEmpty == (invalidation == nil), …)` in `make`, so a kind names mutations exactly when it says how they invalidate. `invalidations(by:)` unwraps the closure. Remove `invalidateOn: []` and the dead `invalidation:` closure from every descriptor that has `invalidateOn: []` (16), including `propertyTerms` and `conclusionDetail` (drop its "Any Conclusion edit can touch any handle…" comment with the closure). `typeSuggestions` keeps its closure; its `default:` arm stays because `CatalogMutation` is not exhaustively matched by the two cases. No behavior change.
Check: `rg -c 'invalidateOn: \[\]' macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` prints `0`; T-5, T-6, T-11 and every `CatalogQueryRegistryTests` / `WorkspaceSessionTests` test pass; scripts/check-pr.sh PASS.

**F-4 — One shared refresh for "rows carry a Property label"** (from MNT-5)
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` (or `CatalogLastUpdated.swift` beside `applyHiddenPolicyOnAny`).
Do: add one named value, `static let followsPropertyLabels: CatalogQueryRefresh = .applyHiddenPolicyOnAny(follows: .list(.propertyVocabulary))`, on `CatalogQueryRefresh`, with a one-line doc ("a key whose rows carry a Property, term or Subject type label"). Use it on the seven Conclusion descriptors and `propertiesWorkspace`, in place of the eight inline entries.
Check: `rg -n 'list\(\.propertyVocabulary\)' macos/App` prints only the definition; T-5 passes; scripts/check-pr.sh PASS.

**F-5 — One test helper for a type's binding ids** (from MNT-4)
Files: `api/ffi/handlers/harness_test.go`, `vocabulary_write_effects_test.go`, `subject_defs_test.go`.
Do: add `bindingIDsForType(t *testing.T, dir, typeID string) []string` beside `propertyIDByKey` / `subjectTypeIDByKey`, reading `GetPropertiesWorkspace` through `mustRPC` and returning the group's Property ids in order. Use it at `vocabulary_write_effects_test.go:85-96` and `subject_defs_test.go:199-211`.
Check: `rg -n 'range .*\.Groups' api/ffi/handlers/*_test.go` prints only the helper (and any test that walks every group for another reason); `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/` passes.

**F-6 — No vocabulary count line until the list has loaded** (from UX-1)
Files: `macos/App/Features/SourceTypes/SourceTypesModel.swift`, `macos/App/Features/Metadata/MetadataModel.swift`, their views if the call site needs it; tests `SourceTypesModelTests.swift`, `MetadataModelTests.swift`. Skill: `add-swift-test`.
Failing test first: `countLineIsAbsentUntilTheListLoads` *(new, each file)* — with no value on the list handle (before load, and after a load that throws with no rows), `countLine == nil`.
Do: `countLine` returns `String?`: `nil` when the list handle has no value; otherwise today's string. `VocabularyHeader.countLine` is already `String?` and hides `nil`.
Check: the new tests pass; T-12 still passes; scripts/check-pr.sh PASS.

**F-7 — Stale comments** (from MNT-6, MNT-8, UX-3)
Files: `macos/App/Features/Catalog/CatalogCounts.swift:10`, `macos/App/Features/Workspace/Session/CatalogMutation.swift:17-19`, `macos/App/Features/Properties/PropertiesModel.swift:475`.
Do: "Absent until a refresh succeeds for that section"; drop "They carry the field's id" (keep the payload); the create-flow comment says a failed bind keeps the dialog open with `formError` and the created Property in the list behind it.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b8043abc`. Gate PASS (orchestrator run, against `f1b3fe65`; xcode-test 51s).
- **F-1** done (`7745df28`): `rg HasPrimaryFile|sqlExistsByFile` is empty.
- **F-2** done (`7745df28`): `rg onSettled` over macos, `docs/*.md` and `.agents` is empty; the spike map's batch-4 working rule (`deployment-plan.md:95`) keeps its historical mention.
- **F-3** done (`9a5025d9`): optional `invalidation` with a precondition; zero `invalidateOn: []`. Also updates the `add-workspace-place` line 102 that said `invalidation` was always required. The Landed note records that this supersedes D-6's literal `propertyTerms` closure.
- **F-4** done (`cc068e26`, `9a5025d9`): `CatalogQueryRefresh.followsPropertyLabels`.
- **F-5** done (`66a81d7b`): `bindingIDsForType`. `subject_defs_test.go:149,163` still walk every group, for other reasons.
- **F-6** done (`e5d84a3e`): `countLine: String?` is nil until the handle has a value; new `countLineIsAbsentUntilTheListLoads` in each model's tests. The "after a load that throws" case is not added, because `FakeStore` has no list-load error hook; with no value, the guard covers both.
- **F-7** done (`455dbad2`).

Minors read by the orchestrator in `39bbe2a0..b8043abc`:
- UX-1: **Resolved**.
- MNT-6, MNT-8, UX-3: **Resolved**.
- MNT-4: **Resolved**.
- MNT-5: **Resolved**.

## Round 2 — review
Reviewers: MNT only (MNT-1, MNT-2, MNT-3 were Fix-now majors). CON, UX and SEC+PRF had no Fix-now blocker or major.
MNT-1, MNT-2 and MNT-3 are **Resolved**. No new findings. Verdict: Clean.

## Result
Gate PASS (against `f1b3fe65`). `main` (`42eb321c`) and S14-27's branch (`f1b3fe65`) are unchanged since this branch was cut.

CI has not run: the workflows trigger only on PRs into `main`. It runs when #693 merges and #694's base moves to `main`.

- **Open:** none.
- **Follow-ups:** PRF-1 (with the UX outside-lens note), MNT-7, UX-2.
- **To user:** none.
- **CON miss:** none.
