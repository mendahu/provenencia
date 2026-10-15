# Ship log — S10-22 Last updated on People, Events, and Places

**Plan:** [`S10-22-list-last-updated.md`](S10-22-list-last-updated.md) · **Branch:** `s10-22-list-last-updated` · **PR:** [#402](https://github.com/mendahu/provenencia/pull/402)
**State:** ready

Base: `main` @ `832ccc22` (planning base `360b405f`; the only change since, #400, touches `PVFormDialog` and none of the plan's Current state files).

## Round 0 — implement
Head `1ab370f0`. Gate PASS (orchestrator run against `832ccc22`). Report: no product deviations. Tests normalize U+202F (narrow no-break space before AM/PM on newer macOS) to a space before comparing. The implementer's first gate hit a flake in `WorkspaceSessionTests.invalidateAllMatching()` (not touched by this PR); it passed on rerun. Manual check not run.

## Round 1 — review
Five lenses, base `832ccc22`, head `1ab370f0`. Verdicts: CON Changes needed · MNT Clean · UX Clean · SEC Clean · PRF Clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | No Landed note; code differs from plan (U+202F before AM/PM, `Equatable` on `LastUpdatedDisplay`, `if let` for yesterday) | Fix now → F-1 | Verified: plan diff changes only the Status line. Precedent S10-21 CON-1 |
| PRF-1 | PRF | minor | §11 says hidden lists refetch on next show; they reload immediately (list specs have no `evictWhenHidden`) | Fix now (record) → F-1; behavior Follow-up | Verified: `CatalogQueryRegistry.swift` list specs set no `evictWhenHidden`. The plan's Decisions are unchanged; the Landed note records the correction. Stale-on-show or row patching is a follow-up |
| CON-3, MNT-1, UX (outside) | CON/MNT/UX | minor | `PVList` header doc now attaches to `PVListDetail`; `PVList` has no doc comment | Fix now → F-2 | Verified: `PVList.swift` lines 3–16, no break before `struct PVListDetail`. STEP-6 asked for this comment on `PVList`; a two-type move |
| CON-2 | CON | major | AC-1.7, AC-1.8, AC-2.4 have only manual-check evidence, not yet run | To user | The §10 manual check is the user's step (ship-pr Finish). Listed in the report; the plan's Stop condition stays if step 5 fails |
| MNT-2 | MNT | minor | `LastUpdatedDisplay.help` ignores the injected calendar/locale (`boardStamp` reparses) | Follow-up | D-6 names `boardStamp(rfc3339)`; fix when S10-41 adopts the type |
| PRF-2 | PRF | minor | Handle-scope path walked twice per write (`audit.Record` + `Resolve`) | Follow-up | Out line in the plan; touches the audit API. Benchmark first |
| UX-1 | UX | nit | Time of day for older changes is hover-only | Follow-up | Board decision (PSL-3); optional accessibility hint later |
| UX-2 | UX | nit | `.fixedSize()` text could overrun the ref in a longer locale | Logged | English only today; revisit with a second language |
| MNT-3 | MNT | nit | Edmonton calendar fixture copied in two test files | Logged | Extract when S10-41 adds a third |
| CON-4 | CON | nit | Tests pin English literals | Logged | Plan prescribed the literals |
| — | SEC | — | No findings | — | — |

### Fix plan 1

**F-1 — Add the Landed note** (from CON-1, PRF-1)
Files: `docs/deployment-plan/spike-10/prs/S10-22-list-last-updated.md`.
Do: add a short `## Landed` section directly under the plan header block (the S10-21 plan's Landed section is the shape). Record only where the code differs from the plan: (1) on current macOS the time separator before AM/PM is U+202F, so the cell reads `Today, 3:15 PM` with a narrow no-break space, and the tests normalize it to a space before comparing; (2) `LastUpdatedDisplay` also conforms to `Equatable`; (3) the yesterday test uses `if let` on `calendar.date(byAdding:)` instead of `!`; (4) §11's risk mitigation was wrong: list specs have no `evictWhenHidden`, so a note save reloads every cached list at once, not on next show. Reload-on-show or patching rows in place is a follow-up. Do not change any Decision, AC, or Scope line.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-22-list-last-updated.md` reports no errors.

**F-2 — Give `PVList` its doc comment back** (from CON-3, MNT-1)
Files: `macos/App/DesignSystem/Components/List/PVList.swift`.
Do: move `PVListDetail` and `PVListDetailColumn` (with their own `///` comments) to sit after the `PVList` type and its `Secondary == EmptyView` extension, before `PVListDensity`, so the header doc block sits directly on `struct PVList`. No code change.
Check: the header doc is immediately above `struct PVList<…>`; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `6ba75fec`. Gate PASS (orchestrator run). F-1 done (`6ba75fec`, Landed note); F-2 done (`d4a88df0`, doc comment back on `PVList`). No deviations.

## Round 2 — review
Lenses with a Fix-now finding: CON, MNT, PRF. Scope `1ab370f0...6ba75fec`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | — | Landed note | Resolved | Note matches the code; check-plan-docs 0 errors |
| CON-3 / MNT-1 | CON/MNT | — | `PVList` doc comment | Resolved | Header doc directly above `struct PVList` |
| PRF-1 | PRF | — | §11 correction | Resolved | Landed note accurate; behavior change stays a follow-up |
| CON-r2 | CON | nit | Landed note said "yesterday test" for app code | Fixed by orchestrator (one-word doc edit in the ship-log commit) | `TimestampFormat.swift:67` |

No new findings. Verdicts: CON Clean · MNT Clean · PRF Clean. No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: PRF-1 (reload hidden lists on show, or patch rows in place), PRF-2 (resolve handle scopes once per write), MNT-2 (inject calendar/locale into the help stamp, with S10-41), UX-1 (full stamp for keyboard/VoiceOver) · to user: CON-2 (run §10 manual check steps 2, 4, 5 for AC-1.7, AC-1.8, AC-2.4)

## Rework — user direction (2026-10-15)

After round 2 the user asked for two refactors on this PR and rated both problems as blockers the review should have caught:
1. One timestamp formatter (timestamp, locale, display option → one line of text, no styling) instead of `LastUpdatedDisplay` beside `TimestampFormat`.
2. Surgical cache updates: patch the moved rows' time from the write instead of reloading all three lists; cache infrastructure, not a one-off.

Plan amended (D-3, D-5, D-6, D-8, D-10, D-11 rewritten; D-12…D-18 new; STEP-9…STEP-15; T-7…T-13). The engine side also keeps last activity in the display cache, patched on commit, so list loads stop scanning audit history, and resolves scopes once per write (retires PRF-2). Lenses tightened: maintainability Q14 (one mechanism per cross-cutting job), Q9 (patch what the write reported), "precedent is not a mitigation"; performance treats any common-path cost that grows with the catalog or its history as a blocker.

## Rework — implement
Plan review: two rounds, no open blocker (review log). Implementer built STEP-9…STEP-15 (`6f5a828a`…`bbe90be7`). Head `bbe90be7`. Gate PASS (orchestrator run against `0a1f566b`). Deviations reported: S10-21's historical plan/review/ship logs left with old symbol names; `audit/scopes.go` kept for `sqlInsertScope`; `promote/bridges_test.go` resolves inline; `audit` test helper `recordResolved` falls back to an empty Set on a Resolve error; whitespace-only gofmt of `sources.go`. Push of `bbe90be7` failed (network/auth); local only.

## Rework — review (all five lenses, fresh, tightened lenses)
Base `0a1f566b`, head `bbe90be7`.
Verdicts: CON Clean · MNT Clean · UX Clean · SEC Clean · PRF Changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Source-page edits mark People, Events, Places stale (full reload on next visit) via `.mutatedSourceWorkspace` in `conclusionTriggers` | To user → **this PR** (plan D-22…D-24, STEP-16…18) | Verified: `conclusionTriggers` includes it; Source-page writes report no effects. User chose this PR, 2026-10-15 |
| PRF-2 | PRF | blocker | Source-page edits mark the Sources list stale (`markStaleOn`) | To user → **this PR** (same) | Verified: `sourcesList.markStaleOn` lists five kinds |
| UX-1 | UX | minor | `.stamp` pads single-digit days (`Oct 04`) unlike the cell and the board (`3 Mar 2025`) | Fix now → F-1 (D-5 amended) | Verified against the board's frames |
| PRF-3, MNT-5 | PRF/MNT | minor | `string(_ date:)` formats short and long dates on every call | Fix now → F-1 (same file) | Lines 65–66 |
| UX-2 | UX | minor | `a11yUpdated` translator comment describes the visible text | Fix now → F-2 | `.spoken` is passed |
| MNT-1 | MNT | minor | Scope → ids is a two-way ternary | Fix now → F-3 | `patchLastUpdated` |
| MNT-2 | MNT | minor | `recordResolved` swallows `Resolve` errors | Fix now → F-4 | `record_helper_test.go` |
| MNT-3 | MNT | minor | `AttachLatestRevision` now attaches activity | Fix now → F-5 (D-20 amended) | One non-test caller |
| MNT-4 | MNT | nit | Two revision+time structs; row types depend on `graphcache` | Fix now → F-5 (D-14, D-15 amended) | Q14: one value for one fact |
| CON-2 | CON | minor | AC-2.7 and AC-2.5b only partly proven | Fix now → F-6 | Tests stop before the reload / sort |
| CON-3 | CON | nit | Map rows describe Sources the old way | Fix now → F-7 | `deployment-plan.md` S10-22 Out, S10-41 In |
| CON-1 | CON | minor | Docs say landed while the PR is open; round-0 Landed note dropped | At finish: orchestrator writes the Landed note (U+202F, type order, inline `Resolve` in `bridges_test.go`, S10-21 docs left as history); map/completed confirmed at merge, as S10-21 did | |
| CON-4 | CON | nit | Sweep hits S10-21's historical docs | Logged → Landed note | History, not code |
| — | SEC | — | No findings | — | — |

### Fix plan R1

**F-1 — One date part in the stamp; format only what each style returns** (UX-1, PRF-3, MNT-5)
Files: `macos/App/Platform/TimestampFormat.swift`, `macos/ProvenenciaTests/TimestampFormatTests.swift`.
Failing test first: add `2026-10-04T15:00:00Z` rows to `stylesFormatOneLine`: `.stamp` → `Oct 4, 2026 · 9:00 AM MDT` (`en_US`), `4 Oct 2026 · 09:00 GMT-6` (`en_GB`).
Do: per amended D-5 — `.stamp` uses the `shortDate` date part; compute the short/long date inside the cases that return them.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Translator comment for the spoken suffix** (UX-2)
Files: `macos/App/Platform/L10n.swift` (`L10n.LastUpdated.a11yUpdated`), `macos/App/Resources/Localizable.xcstrings` (`lastUpdated.a11y.updated`).
Do: comment reads "VoiceOver suffix of a list row; argument is the spoken time (Today at 3:15 PM, Yesterday, October 12, 2026)".
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**F-3 — A scope names its own ids** (MNT-1)
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`, `WorkspaceSession.swift`.
Do: `CatalogLastUpdatedScope.ids(in effects: CatalogWriteEffects) -> [String]` as an exhaustive `switch`; `patchLastUpdated` calls it.
Check: `WorkspaceSessionTests` pass.

**F-4 — The audit test helper fails on a resolve error** (MNT-2)
Files: `core/database/audit/record_helper_test.go`, `core/database/audit/scopes_internal_test.go`.
Do: `recordResolved` calls `t.Fatal` when `effects.Resolve` errors; `TestRecordRejectsUnknownEntity` calls `Record(tx, rev, effects.Set{})` directly.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/` passes.

**F-5 — One activity type; `AttachActivity`** (MNT-3, MNT-4)
Files: `core/database/effects/commit.go`, `core/database/graphcache/activity.go` and tests, `core/database/conclusionheaders/*.go`, `core/database/sources/sources.go` and tests, `api/ffi/handlers/sources.go`, `thumbnails.go`, `conclusion_headers.go`, `core/writes/run.go`, and every `graphcache.Activity` / `AttachLatestRevision` reference (`rg`).
Do: per amended D-14, D-15, D-20 — `effects.Activity`, `Commit{Activity; Effects}`, rename to `AttachActivity`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg "graphcache\.Activity|AttachLatestRevision" core api` finds nothing.

**F-6 — Prove the stale reload and the Sources sort** (CON-2)
Files: `macos/ProvenenciaTests/WorkspaceSessionTests.swift` (`hiddenListsGoStaleVisibleListReloads`), `macos/ProvenenciaTests/SourcesModelTests.swift`.
Do: after the AC-2.7 apply, `query` Events and assert `isFetching`, then the reloaded value; add a `SourcesModelTests` case: s1 rev 5, s2 rev 9, sort `.updated`, apply `CatalogWriteEffects(sources: ["s1"], activity: CatalogActivity(revision: 12, at: T))` → rows `["s1", "s2"]`.
Check: both pass.

**F-7 — Map rows** (CON-3)
Files: `docs/deployment-plan/spike-10/deployment-plan.md`.
Do: S10-22 Out → "The Sources list cell, filters, and search (S10-41)"; S10-41 In drops "from the existing `UpdatedRevision`".
Check: `python3 scripts/check-plan-docs.py` passes.

Then the plan's STEP-16, STEP-17, STEP-18 (D-22…D-24).

## Rework — fix R1 + STEP-16…18
Head `4402d0fe` (pushed). Gate PASS (orchestrator run). F-1…F-7 done; STEP-16…18 done. Deviations reported: `markStaleOn`, `staleMarks`, and the `stale:` parameter deleted (no users after D-24); artifact create applies effects only; a Source-page write now also invalidates that Source's `sourceGraph` and `citationCounts`, all cached `citationsByArtifact`, and `propertiesWorkspace` through `invalidations(for:)` (orchestrator note: these were not invalidated by `.mutatedSourceWorkspace` before — round 2 to rate); `FakeStore.bumpSource` stamps a fixed time.

## Rework — review round 2 (all five lenses: STEP-16/17 is new scope)
Scope `a67120a4...4402d0fe`.
Verdicts: CON Changes needed · MNT Changes needed · UX Changes needed · SEC Clean · PRF Changes needed. Round-1 Fix-now items F-1…F-7 all Resolved; PRF-1 and PRF-2 Resolved.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-4, MNT-6, CON-5 | PRF/MNT/CON | blocker | Source-page writes now reload `sourceGraph`, `citationCounts`, all `citationsByArtifact`, `propertiesWorkspace` (and the engine drops the graph): `effects.sources` fuses activity and evidence | Fix now → R2-F1 (plan D-25) | Verified: `invalidations(for:)` and `graphcache/commit.go` `dropSources(set.Sources)`; the registry gives notes/metadata/credibility a `Source` path |
| UX-3 | UX | blocker | Artifact delete leaves the Sources row's `hasArtifact` and cover stale for the session | Fix now → R2-F2 (plan D-26) | Verified: delete path never calls `refreshCoverFromStore`; `markStaleOn` gone |
| MNT-7 | MNT | major | Three `apply` bodies that already drift | Fix now → R2-F2 (D-26) | `WorkspaceSession.swift` 169–207 |
| CON-6 | CON | minor | T-16 does not assert the dedupe or the AC-2.12 detail reload | Fix now → R2-F3 | Same test file |
| MNT-8 | MNT | minor | Comment says notes report handles; `s.Updated` rebuilt from `res` | Fix now → R2-F3 | Touched files |
| CON-7 | CON | nit | Client-patterns doc overstates which writes return effects | Fix now → R2-F3 | One sentence |
| — | SEC | — | No findings | — | — |

### Fix plan R2 (round-2 blockers; last fix round without the user)

**R2-F1 — Activity is not evidence** (PRF-4, MNT-6, CON-5)
Files: `core/database/effects/registry.go`, `resolve.go`, `scopes.go`, `README.md`, effects tests; `core/database/graphcache/commit.go`, `activity.go`; `core/database/audit` (reads via `effects.Scopes`, no change expected); `api/proto/engine.proto` (+ regenerate), `api/ffi/handlers/write_effects.go`, its test, `source_write_effects_test.go`; `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`; `macos/App/Platform/GoStore.swift`, `FakeStore.swift`; every `cacheSources` / `touchedSources` / `CacheSources` reference (`rg`); `docs/macos-client-patterns.md`, `.agents/skills/add-audit-scope/SKILL.md` where they describe `Source` / `CacheSource`.
Failing test first: T-19 (fails at `4402d0fe`: `propertiesWorkspace` fetches after a note write); T-18's note case (fails: `Sources == [S]`).
Do: plan D-25. For each table, keep `Source` only if `graphcache/source.go`, graph progress, citation counts, or citations-by-artifact SQL reads it; record the list in the commit message.
Check: T-8, `TestSourceScopes`, T-18, T-19 pass; `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg "CacheSource|cacheSources|touchedSources|cache_sources" core api macos` finds only `reserved 3`.

**R2-F2 — One apply body; Artifact delete refreshes its row** (MNT-7, UX-3)
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/App/Features/Sources/SourcePageModel.swift`, tests.
Failing test first: T-20.
Do: plan D-26.
Check: T-20 passes; `WorkspaceSessionTests` pass.

**R2-F3 — Tests and wording** (CON-6, MNT-8, CON-7)
Files: `macos/ProvenenciaTests/WorkspaceSessionTests.swift` (T-16), `CatalogQueryRegistry.swift` (`conclusionTriggers` comment), `api/ffi/handlers/sources.go` and `core/database/sources/sources_test.go` (`res.Activity`), `docs/macos-client-patterns.md` (the "every write under a Source" sentence names the exceptions: the Source row's own writes and promote).
Do: T-16 asserts the union has no duplicate key and that a warmed visible `conclusionDetail(e1)` reloads for AC-2.12.
Check: scripts/check-pr.sh PASS.

## Rework — fix R2
Head `e4d65431` (pushed). Gate PASS (orchestrator run against `0a1f566b`). R2-F1, R2-F2, R2-F3 done. Table classification (commit `e4d65431`, pinned by `TestSourceEvidenceIsNotActivity`): evidence `Source` on `artifacts`, `citations`, `observations`, `subjects`, `date_values`, `name_values`, `identity_claims`, `identity_claim_evidence` (the last two without `SourceScope`, as before); activity `SourceScope` only on `sources`, `source_notes`, `source_metadata`, `source_metadata_layout`, `source_credibility_assessments`, `files`, `citation_notes`, `observation_notes`. Deviations: `citation_notes` and `observation_notes` also lost `Source` (no evidence read loads them); `FakeStore.bumpSource` takes `evidence:`. Review cap reached: a further review round needs the user.

## Rework — review round 3 (user approved, 2026-10-15)
All five lenses, fresh. Scope `ede162cb...e4d65431` (Fix plan R2), base `0a1f566b`.
Verdicts: CON Clean · MNT Changes needed · UX Changes needed · SEC Clean · PRF Changes needed. R2 Fix-now items all Resolved (PRF-4/MNT-6/CON-5, UX-3, MNT-7, MNT-8, CON-6 partly — see CON-9, CON-7).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-5 | PRF | blocker | Subject create/rename reloads `propertiesWorkspace` (an Observation `COUNT(*)` per Property; visible on graph and composer) and citation reads that don't read Subjects — new against base | Fix now → R3-F1 (plan D-27) | Verified: `invalidations(for:)` keys all four reads on `effects.sources`; base `.mutatedSourceGraph` reloaded only `sourceGraph` |
| UX-1 | UX | major | Composer opened from a graph card shows a stale Citation picker after saves: its `citationsByArtifact` key isn't visible (location has no `artifactId`) and is now `.markStale` | Fix now → R3-F2 (plan D-28) | Verified: `listedCitations` reads `queryHandle` only; `PlaceRegistry` lists the key only when `location.artifactId` is set. CON's "composer queries the key" is true only on selection |
| MNT-9 | MNT | major | Artifact delete loads the Source workspace two or three times | Fix now → R3-F3 | Create already does `apply(effects)` then `refreshCoverFromStore()` |
| MNT-10 | MNT | major | `docs/audit-revision-history.md:334` says the source scope comes from `Set.Sources` | Fix now → R3-F3 | `rg "Set\.Sources" docs` |
| MNT-11 | MNT | minor | `date_values` / `name_values` chain pasted twice per entry | Fix now → R3-F1 (same file) | `registry.go` |
| CON-9 | CON | minor | T-16's dedupe assertion cannot fail | Fix now → R3-F3 (union moves into the registry, tested there) | `ensureQuery` returns on in-flight |
| CON-10 | CON | nit | FakeStore attach reports activity only | Fix now → R3-F3 (same helper as MNT-12) | `FakeStore.swift:687` |
| MNT-12 | MNT | nit | `bumpSource(_:evidence:)` defaults silently | Fix now → R3-F3 | required argument |
| MNT-13 | MNT | nit | Comment wrap | Fix now → R3-F3 | |
| CON-11 | CON | nit | D-25/T-18 name positions; code (correctly) has `identity_claims` | Landed note | |
| — | PRF | note | `Resolve` walks a shared path twice for `Source`/`SourceScope` | Logged | Per changed row, does not grow with the catalog |
| — | PRF | note | Metadata value set reloads visible `metadataFieldsList` (`COUNT(*)` per field) | Follow-up idea (pre-existing, same shape as PRF-5) | |
| — | SEC | — | No findings | — | — |

### Fix plan R3 (user-approved round 3 findings; then finish)

**R3-F1 — Each read follows only what feeds it** (PRF-5, MNT-11)
Files: `core/database/effects/registry.go`, `resolve.go`, `README.md`, effects tests (`TestSourceEvidenceIsNotActivity`); `api/proto/engine.proto` (+ regenerate), `api/ffi/handlers/write_effects.go`, tests; `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `CatalogQueryRegistry.swift`; `macos/App/Platform/GoStore.swift`, `FakeStore.swift`; tests.
Failing test first: a `WorkspaceSessionTests` case — graph place visible (`propertiesWorkspace` visible), apply a Subject create's effects (`sources: ["s1"]`, no `citationSources`, no `propertyUsage`) → `propertiesWorkspace` and `citationCounts(s1)` not fetching; an Observation create (`citationSources`, `propertyUsage: true`) → both reload.
Do: plan D-27; `valueSource(col)` helper used by both fields of `date_values` and `name_values`.
Check: Go and Swift tests pass.

**R3-F2 — The composer's location names the Artifact it shows** (UX-1)
Files: `macos/App/Features/Workspace/WorkspaceNavigation.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift` (and its view if it owns navigation), tests.
Failing test first: warm `citationsByArtifact(a1)`; enter the composer place without `artifactId`; select `a1`; apply a citation write's effects → the key reloads (it is visible).
Do: plan D-28.
Check: the test passes; the composer's existing tests pass.

**R3-F3 — Small fixes** (MNT-9, MNT-10, CON-9, CON-10, MNT-12, MNT-13)
Files: `macos/App/Features/Sources/SourcePageModel.swift` (delete: `context.session.apply(effects)` then `await context.refreshCoverFromStore()`, as create does; drop the hand-built cover fields the refresh replaces), T-20 asserts `getSourceWorkspaceCalls` rises by exactly one; `docs/audit-revision-history.md:334`; `CatalogQueryRegistry.swift` gains `invalidations(for effects:, by mutation:, project:)` returning the deduped union, `WorkspaceSession.apply` calls it, and `CatalogQueryRegistryTests` asserts a key named by both appears once (T-16's dedupe line goes); `FakeStore.bumpSource` takes a required `evidence:` and attach passes `true`; the `conclusionTriggers` comment rewrapped.
Check: scripts/check-pr.sh PASS.

## Rework — fix R3
Head `8a449b09` (pushed). R3-F1 and R3-F3 done (failing tests first). R3-F2 **stopped** correctly: `CitationComposerEntry.identityKey` includes the Artifact, so replacing the composer's location remounts it and loses the draft. D-28 rewritten (orchestrator, 2026-10-15): the composer's write funnel reloads the Artifact list it shows.

**R3-F2b — The composer reloads the list it shows** (UX-1)
Files: `macos/App/Features/CitationComposer/CitationComposerContext.swift` (`applyWrite`), a composer or `WorkspaceSessionTests` test.
Failing test first: warm `citationsByArtifact(a1)`; composer place visible without `artifactId`; context `artifactID == "a1"`; `context.applyWrite(CatalogWriteEffects(sources: ["s1"], citationSources: ["s1"]))` → the `a1` list handle is fetching.
Do: plan D-28 (rewritten).
Check: the test passes; the composer tests pass; scripts/check-pr.sh PASS.
R3-F2b done (`5659b414`, failing test first: `composerWriteReloadsTheArtifactListItShows`). Orchestrator checked: `selectedArtifactID` is `context.artifactID`, so the reload follows an Artifact switch. Orchestrator read the R3-F1 diff against D-27; it matches. Head `5659b414`. Gate PASS (orchestrator run against `0a1f566b`). R3 was not reviewed by a further round (the user approved round 3, then finish).

## Result
Gate PASS · CI: see PR · open blockers: none · follow-ups: writes still without `WriteEffects` (Source create/update/cover/delete, promote, Property writes); a visible list reloads in full when a handle's values change; metadata value save reloads visible `metadataFieldsList`; `Resolve` walks shared paths once per field; time of day only on hover (UX-1, round 1) · to user: §10 manual check (AC-1.7, AC-1.8, AC-2.4); R3 fixes had no independent review round.
