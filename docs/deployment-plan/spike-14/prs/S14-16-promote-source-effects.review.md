# Review — S14-16 Promote and Source writes return their effects

Plan: [`S14-16-promote-source-effects.md`](S14-16-promote-source-effects.md). Planning base `main` @ `25ed4bd`.

## Planning notes

- **Shared proto header message.** The map row's last item had already landed in [#662](https://github.com/mendahu/provenencia/pull/662) (`3a1984b5`, `ConclusionSummary`, `GoStore.mapConclusionSummary`). The plan does nothing for it; the map row and scope entry say so (2026-11-02).
- **`.deletedSource` in `conclusionTriggers`.** A Source delete is refused while the Source has Artifacts or Subjects (`deleteimpact` inbound edges), so it touches no handle; dropping it loses no reload.

## Round 1 (2026-11-02)

Maintainability: **Fits with changes**. Executability: **Not ready**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | Promote header would read a live title beside a frozen ref and mark (`attachSource`) | **Fixed.** D-16: `PromoteModel.source` is the one read; title, ref and mark are computed; `attachSource` becomes `warmSourceKeys`. AC-2.6, T-16 type-change case |
| M-2 | blocker | Second copy of the toolbar's "session title, else carried" rule | **Fixed.** D-16: `WorkspaceLocationLabels.sourceTitle(sourceID:carried:names:)`, used by the toolbar and Promote. T-17 |
| M-3 | major | `promote.Save` becomes a test-only near-copy of `applyRow` | **Fixed.** D-17 `fileSubject`; STEP-4; T-4b; stop line |
| M-4 | major | `seedPromotion` would copy the fake's Apply mint block | **Fixed.** D-15: `seedPromotion` is the fake's one membership writer; Apply calls its locked form |
| M-5 | major | `revalidatedKeys`' `mutation` and `BatchResult.Revision` go dead | **Fixed.** D-13, D-4, Deletes |
| M-6 | major | Docs left false: `add-workspace-place` § 6, effects README `evidence_graphs` list, `List` doc, `add-ffi-handler` | **Fixed.** D-1 (`List` doc), STEP-1 (README), STEP-10, DoD sweep adds "report(s) none" |
| M-7 | major | Bridge mints are `canonical_entities` rows, so a join can report `counts` | **Fixed.** D-1 wording, AC-1.2, Neighbors, T-4 bridge case. One extra recount accepted |
| M-8 | minor | Delay hook shape exists (`promoteProposeDelays: [Duration]`) | **Fixed.** D-10, Codebase fit, T-12 |
| M-9 | minor | Two exclusive knobs on `Effect` | **Declined.** Replacing `ListsOnField` reshapes the `observations` entry and its tests for a gain no named add needs; D-2's doc comments now say the two fields are one choice |
| M-10 | minor | `publish*` badge writes sit outside the newest-wins guard | **Logged.** Out row; not planned |
| M-11 | nit | `invalidation:` closure that never runs | **Logged.** Out row → S14-27 |
| M-12 | nit | Pin test beside the other `WriteEffects` descriptor tests | **Fixed.** D-8 moves it into `api/ffi/handlers/write_effects_test.go` (still under `api/ffi`, as the working rule says) |
| X-1 | blocker | STEP-3's regeneration breaks the Swift build until STEP-8 | **Fixed.** Steps reordered: Swift `promoteSubject` goes in STEP-3, before STEP-4 regenerates; Risk row rewritten |
| X-2 | blocker | Sweep patterns hit `unpromoteSubject`, `PVDeleteImpactPreviewData.promotedSubject`, the placechains helper | **Fixed.** Word-bounded and qualified patterns; expected unrelated hits named |
| X-3 | major | Seed rewrite shapes unspecified | **Fixed.** D-7: `personBeside` signature, id conversions, every deleted case and where core covers it |
| X-4 | major | No rule for a missing `protoc` / Xcode | **Fixed.** Steps preamble and a stop line: never hand-edit generated files |
| X-5 | major | Stale `evidence_graphs` sentence in the effects README; test name | **Fixed.** STEP-1 names the README edit; checks name `TestSwiftVocabularyMatchesGo` |
| X-6 | major | `wrote` binding; all-skip Done behavior | **Fixed.** D-4 names `_, wrote, err :=` and the empty-effects outcome; Neighbors |
| X-7 | minor | Unknown-method number | **Fixed.** 1000 (T-7, STEP-4) |
| X-8 | minor | `completed.md` path; brief box | **Fixed.** Full paths; brief box N/A |
| X-9 | minor | Delay queue empty; `applySource` callers | **Fixed.** Empty means no delay (D-10); `SourceArtifactsSection` dropped from STEP-7, the two callers named in D-12 |

Round 1 had blockers, so round 2 runs.

## Round 2 (2026-11-02)

Maintainability: **Fits with changes**. Every round-1 answer holds except M-10.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-13 | blocker | (M-10 reopened) The Sources badge keeps `publishSources` beside the `counts` recount, and `add-workspace-place` teaches both | **Fixed.** D-18 deletes the Sources publish path (`publishSources`, `SourcesModel.syncCatalogCounts`, the `.onSettled` hook, their tests); STEP-6 and STEP-10 rewrite the skill's "Sidebar badges" line and checklist item. The vocabulary `publish*` writers go on S14-17's scope entry (map change, reported to the user) |

Below the round-2 bar, not raised: `promote.Result` and `promote.Written` are the same pair (D-17's `fileSubject` returns it as a tuple); `refreshGeneration` is another hand-rolled generation counter.

Executability: **Ready with changes**. Every round-1 answer holds.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-10 | blocker (environment) | No `protoc` or Xcode in the reviewer's Linux sandbox | **Declined.** Not a plan defect: the stop line covers it, and `implement-pr` runs on a Mac with both. Logged |
| X-11 | major | T-12 passes without the guard if the fake reads counts after its sleep | **Fixed.** D-10: the fake snapshots inside `withState`, then sleeps, then returns the snapshot; T-12 says it fails without the guard |
| X-12 | major | Which checks stay in the fake's Apply and which the seed makes `preconditionFailure` | **Fixed.** D-15 names `seedPromotionLocked`'s signature, Apply's kept guards, and the seed's three `preconditionFailure`s |
| X-13 | minor | Test syntax, the T-12 Source add, `EvidenceGraphStoreTests.deleteSource`, the source-types read | **Fixed.** T-8, T-13 use `CatalogWriteEffects(…)`; T-12 adds the Source through `createSource`; STEP-7 names the `rg` and `EvidenceGraphStoreTests`; D-16 names the `queryHandle` read |
| X-14 | minor | Pin-test helper shape; a request with no response | **Fixed.** D-8: `missingEffects(in:pending:)`; a missing response is a failure |
| X-15 | minor | "STEP-10 drops it from the map" is already done | **Fixed.** "Already on `main`" says the map already says so |
| — | note | `identityClaimProto` loses its last callers | **Fixed.** D-6 and Deletes name it |

No open blocker after round 2. Verdict: **Planned**.
