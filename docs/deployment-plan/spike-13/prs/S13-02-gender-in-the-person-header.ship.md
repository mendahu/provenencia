# Ship log — S13-02 Gender in the person header

**Plan:** [`S13-02-gender-in-the-person-header.md`](S13-02-gender-in-the-person-header.md) · **Branch:** `s13-02-gender-in-the-person-header` · **PR:** #659
**State:** stopped — CI budget test red on the runner, not this PR

## Round 0 — implement
Base `main` @ `a6523137`.
Head `901b9ea6`. Gate PASS (orchestrator run). Report deviations:
1. T-1 "user Property" builds the user's own "Gender" as a text Property (`properties.Create` refuses a researcher term Property); same D-5 assertion. In the Landed note.
2. Test builders resolve `SoleGraph`, `f.prop`, `f.typeID` before `writes.Call`; reading inside the write transaction deadlocks on the single connection. In the Landed note.
3. Commit 4c34ecba alone does not build `./api/...`; the head does.
4. STEP-4 and STEP-5 share commit e78b5185.
5. §10 manual check not run in the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — `macos/App/**` routes UX; `api/ffi/**` routes SEC; `core/database/**` routes PRF.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Persons read loop is the third copy of the stored-or-fresh display loop | Fix now → F-1 | verified: `compose.go` `composePersons`/`composeEvents`/`composePlaces` share the head filter (`g.Display` + type check) and tail (`fresh` or `g.Display(n.ID).(H)`); this PR rewrote the persons tail. Same-pass cleanup |
| MNT-2 | MNT | major | `citeOn` copies `cite`; the citation write is copied across test files | Fix now → F-2 | verified: `walks_test.go:31,47` differ only in graph/artifact; 67 `citations.CreateWithObservations(` sites in 30 test files, most the same `writes.Call` body. This PR adds a copy, so the same-pass rule takes every package; `observationstest.SaveCitation` already holds one copy (with staging) |
| MNT-3 | MNT | minor | `TermFacts` steals `PersonHeader`'s doc comment | Fix now → F-3 | verified: `engine.proto` 1279–1295 |
| MNT-4 | MNT | minor | graphcache says stored display rows are "finished" | Fix now → F-4 | verified: `graphcache/graph.go:36`, `display.go:27`; contradicts this PR's package doc |
| UX-note | UX | — | Go tests need the fts5 build tag | Rejected | existing tooling; `scripts/check-pr.sh` passes the tag; not this PR |

CON: clean. UX: clean. SEC+PRF: clean.

Nits (logged): `CON-1 · PersonGenderTreatmentTests.swift:5 · helper shapes differ from the plan's named shape` · `CON-2 · ship log untracked (orchestrator commits it at finish)` · `MNT-5 · conclusionheaders_test.go:302 · local gender closure repeats f.termObs` · `MNT-6 · compose.go propLabelOf · name vs sibling prop` · `MNT-7 · conclusion_detail.go:187 · second ConclusionTerm literal beside termProto (different input; may stay)`

### Fix plan 1

**F-1 — One stored-or-fresh display read for every header kind** (from MNT-1)
Files: `core/database/conclusionheaders/compose.go`.
Failing test first: none (refactor; behavior is pinned by the existing header tests and `PROVENENCIA_GRAPH_VERIFY=1`).
Do: add two generic helpers in `compose.go`, e.g. `func needed[H any](g *graphcache.Graph, nodes []*graphcache.Node, store bool) []*graphcache.Node` (skip a node whose `g.Display` row is already an `H` when `store`) and `func displayed[H any](g *graphcache.Graph, nodes []*graphcache.Node, fresh map[string]H, finish func(H) H) []H` (fresh row, else the stored `H`, through `finish`). Move `composePersons` (finish `finishPerson`), `composeEvents` (finish `finishEvent`) and `composePlaces` (identity finish, then `finishPlaces` as now) onto them. Keep each kind's own middle (follow, `g.Nodes`, place graph, lead pictures, `oneX`, `SetDisplay`). No behavior change.
Check: `grep -n 'g\.Display(' core/database/conclusionheaders/*.go | grep -v _test` prints only lines inside the two helpers. `PROVENENCIA_GRAPH_VERIFY=1` conclusionheaders tests pass; scripts/check-pr.sh PASS.

**F-2 — One test helper for filing a Citation with Observations** (from MNT-2)
Files: `core/database/observations/observationstest/observationstest.go`, and every `*_test.go` under `core` and `api` that files a Citation through `writes.Call(… "create_citation_with_observations" …, citations.CreateWithObservations(tx, userID, citations.CreateInput{GraphID, ArtifactID, LocatorJSON}, inputs))` with nothing else in the transaction.
Failing test first: none (test-fixture refactor).
Do: add `observationstest.Cite(c *database.Catalog, userID, graphID, artifactID []byte, locator string, inputs ...observations.Input) (citations.CreateResult, []rowchange.Change, error)` holding the one `writes.Call` body; make `SaveCitation` stage then call it. In `conclusionheaders`, keep one fixture body: `citeOn` calls `observationstest.Cite`, `f.cite` is a one-line call to `citeOn` with the fixture's graph and artifact (keep `f.graph` if `SoleGraph` must be resolved outside the write, per round-0 deviation 2), `citeToponyms` and the inline copies at `conclusionheaders_test.go:106,194,554`, `placechains_test.go:503` call it. In every other package listed by the grep below, move each site that matches the body shape above onto `observationstest.Cite`. Leave a site where: (a) the test's subject is `citations.CreateWithObservations` or the write itself (`observations_test.go` tests of the write, error-path assertions on that call), (b) the transaction does more than the one call, or (c) importing `observationstest` would create an import cycle (in-package tests of `artifacts`, `properties`, `observations`, `citations`, `ingest`, `writes`, or anything `observationstest` imports). Name every left site and its reason in the report.
Check: `grep -rn 'citations\.CreateWithObservations(' core api --include='*_test.go'` prints only the left sites the report names, plus the one line in `observationstest.go`. scripts/check-pr.sh PASS.

**F-3 — `PersonHeader` keeps its doc comment** (from MNT-3)
Files: `api/proto/engine.proto`, regenerated `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`.
Do: move `message TermFacts` and its comment above the `PersonHeader` comment; regenerate with the repo's proto generation command.
Check: generated `TermFacts` doc starts "TermFacts is one term-valued Property"; `PersonHeader` doc starts "PersonHeader is one Person as a row". scripts/check-pr.sh PASS.

**F-4 — graphcache says a stored row is composed, not finished** (from MNT-4)
Files: `core/database/graphcache/graph.go:36`, `core/database/graphcache/display.go:27` (and `display.go:3` if it says the same).
Do: say a stored header row is composed but not finished; its owner fills read-time labels on read.
Check: `grep -n finished core/database/graphcache/*.go` shows no claim that the stored row is finished. scripts/check-pr.sh PASS.

## Round 1 — fix
Head `748c9d2a`. Gate PASS (orchestrator run). Commits: F-1 `772f97e4`, F-3 `281b2829`, F-4 `4c10b197`, F-2 `748c9d2a`.
- F-1 done: `needed[H]`, `displayed[H]`; `g.Display(` only inside the helpers.
- F-2 done: `observationstest.Cite`; `SaveCitation` stages then calls it; conclusionheaders fixture on one body; 25 more sites moved. Left sites, with reasons: import cycle (`artifacts_test.go:551,1216`); input carries more than GraphID/ArtifactID/LocatorJSON (`autoreconciler_test.go:935`, `conclusiondetails_test.go:230,450`, `placechains_test.go:504`, `observations_test.go:475`, `properties_delete_test.go:357`, `deleteimpact/facets_test.go:142`, `impact_test.go:242,681`, `conclusion_release_test.go:84`, `conclusion_test.go:63`); whole `CreateInput` passed through (`audit/scopes_test.go:193`); runs inside the caller's transaction (`effects/reads_resolve_test.go:130,155`, `evidence_graphs_test.go:215,276`). Deviation: six `"create_citation"`-action fixture sites now record `"create_citation_with_observations"`; two `writes.Run` sites moved too.
- F-3 (MNT-3, minor): **Resolved** — orchestrator read the proto diff; each message has its own comment, regenerated.
- F-4 (MNT-4, minor): **Resolved** — orchestrator read the graphcache diff.

## Round 2 — review
Reviewers: MNT (had majors MNT-1, MNT-2). Not re-sent: CON, UX, SEC+PRF (no Fix-now blocker or major).
Both round-1 majors **Resolved**: MNT-1 (one `needed`/`displayed` pair in `compose.go`; F-1 grep holds), MNT-2 (`observationstest.Cite` is the one body; ten left sites read, each matches its reason). No new blockers. No Fix plan 2.

Nits (logged): `MNT-8 · observationstest.go · Cite takes only GraphID/ArtifactID/LocatorJSON, so 13 left test sites that set Transcription, Notes, Label or TranscriptionUncertain still hand-write the write; taking a citations.CreateInput would move most of them`

## Finish
Merged `origin/main` @ `c7d19d1e` (S14-12 plan docs only) → head `1f728311`. Gate PASS.

## CI
Go tests fails `TestLargeSourceBudgets/after_a_handle_commit`: 1.615s, then 2.014s on a rerun of the failed job (budget 1.5s). Not this PR: locally the case runs 0.59–0.63s on the head and 0.61–0.70s on `main` @ `c7d19d1e`; the PR does not touch promotealign's read path (only a test fixture moved onto `observationstest.Cite`); S14-11's run 41316389989 failed the same case at 1.89s before it merged, then passed. Taken to the user.

## Result
Gate PASS · CI red: `TestLargeSourceBudgets/after_a_handle_commit` (promotealign) over its 1.5s budget on the runner, 1.615s then 2.014s on rerun · open: none · follow-ups: MNT-8 (widen `observationstest.Cite` to take a `citations.CreateInput` and move the 13 left sites) · to user: — · CON miss: none
