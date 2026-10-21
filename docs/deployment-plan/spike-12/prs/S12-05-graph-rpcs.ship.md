# Ship log — S12-05 Graph RPCs and graph list

**Plan:** [`S12-05-graph-rpcs.md`](S12-05-graph-rpcs.md) · **Branch:** `s12-05-graph-rpcs` · **PR:** #511
**State:** ready

## Round 0 — implement
Head `74dda084`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh` against `e169b711`). Report deviations:
`call_test.go` unknown-method case moved from 103 to 9999 (103 is now `ListEvidenceGraphs`); handler tests use a local
`rpc`/`mustRPC` helper instead of `runRPC` tables (D-16); `FakeStore.listEvidenceGraphs` also marks the session held;
commit `a534214a` does not compile its tests alone (fixture lands in the next commit); manual check not run.

## Round 1 — review
Reviewers: CON (sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `Localizable.xcstrings`,
new error codes, `api/ffi/**`, `core/database/**`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | Graph list never invalidated on Source delete; spec comment claims effects carry it | To user → user (2026-10-21): add `.deletedSource` → F-1 | verified: `DeleteSourceResponse {}` (`engine.proto:561`), `GoStore.deleteSource` returns Void, `SourcePageModel.swift:199` applies only `.deletedSource`; contradicts D-11's rejected `invalidateOn: [.deletedSource]` |
| PRF-1 | PRF | blocker | Hidden graph list revalidates on every Artifact/Source create | To user → user (2026-10-21): use `.markStale` → F-2 | verified: `WorkspaceSession.swift:221` always revalidates visible keys, so `.markStale` only affects hidden ones and D-11's reason does not hold; `sourcesList` uses `.markStale`; contradicts D-11's rejected `.markStale` |
| CON-1 | CON | major | AC-1.3 `CreateArtifact` effects clause untested in Go | Fix now → F-3 | verified: `TestCreateArtifactGraphIDs` reads only the artifact id |
| MNT-2 | MNT | major | `rpc`/`mustRPC` helpers in a feature test file, not in the skill | Fix now → F-4 | verified: shared helpers live in `harness_test.go`; skill names only `runRPC` |
| UX-1 | UX | minor | `artifact_in_use` says "that artifact" when an Update removes several | To user → user (2026-10-21): reword now → F-6 | copy fixed by plan §6 Strings; Update path refuses for a set (AC-2.4) |
| SEC-out | PRF→CON | minor | `FakeStore.createArtifact` links graphs in every project, not the Artifact's own | Fix now → F-5 | file this PR changes; Go scopes by project |
| MNT-out | MNT→CON | — | `graphFixture` beside `citationFixture` (D-16) | Rejected | wraps `citationFixture`, no copy; Landed note gets a line in F-4 |

Nits (logged): `CON-2 · evidence_graphs_test.go · subtests not table-driven (D-16 deviation)` · `UX-2 · PVDeleteImpact.swift:678 · add #Preview("Gate last graph")` · `UX-3 · L10n.swift:9664 · S12-12 inline validation for invalid graph` · `MNT-3 · deleteimpact/impact.go:121 · share projectRows between ListedGroup and collectGroups` · `MNT-4 · FakeStore.swift · four copies of graph index guard` · `MNT-5 · FakeStore.swift · trims name but not description`

### Fix plan 1
User approved F-1, F-2 and F-6 against D-11 and §6 Strings on 2026-10-21. Record each correction in the plan's Landed note.

**F-1 — Source delete drops the graph list** (from MNT-1)
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`.
Failing test first: a `CatalogQueryRegistryTests` case where `.deletedSource(id:)` invalidates `.key(.evidenceGraphsList(project:))`. It fails today because `invalidateOn: []`.
Do: on the `evidenceGraphsList` Spec, set `invalidateOn: [.deletedSource]`. Replace the comment with: "Graph and link writes report the `evidence_graphs` list in `WriteEffects`; `deleteSource` reports none, so `.deletedSource` drops it."
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Hidden graph list marks stale instead of reloading** (from PRF-1)
Files: same as F-1.
Failing test first: `registry.hiddenPolicy(.evidenceGraphsList(project:)) == .markStale`.
Do: add `whenHidden: .markStale` to the same Spec, with a short comment in the style of `sourcesList` (a visible key still revalidates; `WorkspaceSession.apply(invalidations:)`).
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — Go asserts CreateArtifact reports the graph list** (from CON-1)
Files: `api/ffi/handlers/artifacts_test.go`.
Failing test first: none. This is a guard that should pass on arrival (`TestCreateArtifactGraphIDs/effects_report_graph_list`).
Do: have the fixture's create helper return the `CreateArtifactResponse`, or add a variant that does. For the chosen-graph case and the empty-`graph_ids` case, call `assertGraphEffects(t, resp.GetEffects())`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — Shared RPC test helpers live in the harness** (from MNT-2)
Files: `api/ffi/handlers/evidence_graphs_test.go`, `api/ffi/handlers/harness_test.go`, `.agents/skills/add-ffi-handler/SKILL.md`, plan Landed note.
Do: move `rpc` and `mustRPC` into `harness_test.go` beside `marshalProto`. Under the skill's "Handler tests (`runRPC`)" section, add one line: `runRPC` runs the call under test; arrange steps that chain several RPCs use `mustRPC` / `rpc`. In the Landed note's D-16 deviation, add that `graphFixture` wraps `citationFixture`.
Check: `grep -n "func rpc\|func mustRPC" api/ffi/handlers/*_test.go` prints only `harness_test.go`; scripts/check-pr.sh PASS.

**F-5 — FakeStore links a new Artifact only within its own project** (from the SEC+PRF Outside-my-lens note)
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`.
Failing test first: `EvidenceGraphStoreTests.createArtifactLinksOnlyItsProjectsGraphs`. Seed graphs in two projects that share a Source id, create an Artifact in one, and expect the other project's graph to stay unlinked.
Do: in `createArtifact`, link only graphs under `evidenceGraphsByProject[projectDir]`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-6 — `artifact_in_use` copy works for one or many** (from UX-1)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, plan §6 Strings via the Landed note. Skill: `add-localized-string`.
Do: change the English value of `L10n.Errors.evidenceGraphsArtifactInUse` to "Citations in this evidence graph still use an artifact you removed, so it stays linked." The code stays plain, with no params (D-12).
Check: `evidenceGraphErrorsHaveCopy` passes; `python3 scripts/check-localizable-xcstrings.py` OK; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `1e3331ba`. Gate PASS (orchestrator re-ran). F-1 through F-6 done. F-1, F-2, F-5 and F-6 each went red first; F-3 passed on arrival, as a guard.
Deviation: F-6 also updated the `L10nResolutionTests` expectation. The Landed note records the D-11 Spec correction, the `artifact_in_use` copy change, and that `graphFixture` wraps `citationFixture`.

## Round 2 — review
Reviewers: CON, MNT, SEC+PRF (each had a Fix-now blocker or major). UX not re-sent: UX-1 was a minor.
UX-1 **Resolved**: `L10n.swift:9669` and `Localizable.xcstrings:2524` now read "Citations in this evidence graph still use an artifact you removed, so it stays linked." The code is still plain.
CON-1, MNT-1, MNT-2, PRF-1 and F-5 (FakeStore scoping) are each **Resolved**, as reported by their round-2 reviewers; all three verdicts are Clean. No new blockers, so there is no Fix plan 2.

Nits (logged): `MNT-6 · S12-05-graph-rpcs.md Landed note · D-16 line still names evidence_graphs_test.go as the helper location before saying harness_test.go` · `CON-2 · unchanged`

## Result
Gate PASS at `1e3331ba` · `origin/main` unchanged since the branch was cut (`e169b711`) · CI green (Go tests, macOS tests at `1e3331ba`; the ship-log commit skips by path) · open: none ·
to user: MNT-1, PRF-1 (D-11 corrections) and UX-1 (§6 copy), all answered 2026-10-21 and fixed ·
follow-ups: UX-2 (last-graph gate preview), UX-3 (S12-12 inline validation), MNT-3, MNT-4, MNT-5 (logged nits) ·
CON miss: none (MNT-1 came from D-11's false premise; the code matched the plan, so conformance had nothing to miss)
