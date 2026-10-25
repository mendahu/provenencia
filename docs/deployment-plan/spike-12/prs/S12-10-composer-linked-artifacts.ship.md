# Ship log — S12-10 Composer offers linked Artifacts

**Plan:** [`S12-10-composer-linked-artifacts.md`](S12-10-composer-linked-artifacts.md) · **Branch:** `s12-10-composer-linked-artifacts` · **PR:** #540
**State:** ready

## Round 0 — implement
Head `e9f3f8a6`. Gate PASS (orchestrator run against `933cc3c1`). Report: no deviations from Decisions. Notes: FakeStore create-graph returns no `citationSources` (D-10 lists only link, unlink, update); manual check not run; two new tests fixed before first push (Go list compared as a set; Swift fixture subject position).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — UX for `macos/App/**` and `Localizable.xcstrings`; SEC for `api/ffi/**`; PRF for `core/database/**` and `macos/App/Features/**` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | "Citation used by graph" rule spelled four times across `citations` and `evidencegraphs` | Fix now → F-1 | verified: `citations.go` `sqlUsedByGraph`, the `observation_count` subquery and the names joins, plus `evidencegraphs.sqlCitationsUsing`, linked by a comment. Same-pass cleanup: this PR adds three of the copies. Supersedes D-1's "twin + comment" and D-3's "its SQL stay as they are"; D-3's intent (Unlink ignores Observation-less Citations) is kept. No researcher-visible change, so not To user |
| CON-1, MNT-3 | CON, MNT | major | `docs/catalog-deletes.md:104` and `.cursor/rules/catalog-deletes.mdc:18` still say `.deletedSubject(sourceId:)` | Fix now → F-2 | verified: `git grep "deletedSubject(sourceId"` prints those two lines only |
| PRF-1 | PRF | major | Composer `artifacts` re-scans the project graph list on every read, A×G inside the picker's row builder | Fix now → F-3 | verified: computed property at `CitationComposerModel.swift:140`; six reads in `CitationComposerEvidenceColumn.swift` (one per menu row, line 161) and about ten in the model. Fix is read-once per body / per method; the session-level id index goes to S12-13 (Later PR) |
| MNT-2 | MNT | major | Graph linked-Artifacts session read lives in `SourcesListNavigation`; the graph-list handle lookup is written five times | Fix now → F-4 | verified: `SourcesListNavigation.swift:33,48,69`, `SourcesModel.swift:184`, and the composer. Same-pass cleanup (third copy added by this PR). Moves D-7's session read to `Workspace/Session/`; supersedes D-7's file placement only, its rule and signature shape stay |
| MNT-4 | MNT | major | Four more hand-written `create_citation_with_observations` wrappers in `citations_test.go` | Fix now → F-5 | verified: 15 copies in the file; same-pass cleanup |
| PRF-2, MNT-5 | PRF, MNT | minor | Person-names query re-checks `sqlInGraph` its `so` join already guarantees | Fix now → F-6 | verified: both branches join `subjects so ... so.graph_id = ?`; D-2 asks only for that join |
| MNT-6 | MNT | minor | FakeStore graph create reports no `citationSources`; the engine's Create links through `Link` | Fix now → F-7 | verified: `evidencegraphs.go:188` calls `Link`; D-10's reason is "as the engine now does" |
| CON-2 | CON | minor | Plan has no Landed note | Fix now → F-8 | DoD asks for it; F-1 and F-4 now make deviations to list |

Nits (logged): `CON-3 · deployment-plan.md:100 · historical .deletedSubject mention, leave` · `UX-1 · L10n.swift:4960 · noArtifactsCallout says only "add"; once S12-13 ships checkboxes, offer linking too` (Later PR: written into the S12-13 map entry) · `MNT-7 · effects/registry.go:58-59,90-92 · Go comments describe Swift's cache keying` · `MNT-8 · CitationComposerModel.swift:401 · evidenceGraphsKey duplicates the key linkedArtifacts builds` (folded into F-4)

Later PR (S12-13 map scope entry, `deployment-plan.md` § S12-13): the canvas's `canCite` / `fileRows` read the graph through the session accessor from F-4; when they read per row, give the session one id-indexed view of the graph list (PRF-1 option 2), and point `SourcePageModel.evidenceGraphLocation`'s `graphsBySource` regroup at it. UX-1 copy note.

### Fix plan 1

**F-1 — One spelling of "an Observation on this Citation in graph G"** (from MNT-1)
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/citations/citations.go`. Skill: `add-catalog-query`.
Failing test first: none (refactor); the existing `TestCitations` list/count/names cases, `TestCitationCountsByEvidenceGraph`, and the evidencegraphs Unlink-refusal tests must stay green before and after.
Do: In `evidencegraphs`, export one constant, `SQLObservationsInGraph` = `observations o JOIN subjects s ON s.id = o.subject_id WHERE o.citation_id = c.id AND s.graph_id = ?` (doc: the Observations of Citation `c` on Subjects of one graph; binds one graph id; the Citation table is aliased `c`). Rebuild `sqlCitationsUsing` as `SELECT c.id, c.ref FROM citations c WHERE c.artifact_id = ? AND EXISTS (SELECT 1 FROM ` + it + `) ORDER BY c.ref COLLATE NOCASE` (binds unchanged: artifact, graph; it still does not take the "no Observation" half, per D-3's reason). In `citations`, build `sqlUsedByGraph` as `EXISTS (SELECT 1 FROM ` + `evidencegraphs.SQLObservationsInGraph` + `)` and the `observation_count` subquery as `(SELECT COUNT(*) FROM ` + it + `)`. Delete the "change both together" comment; say instead that the rule lives in `evidencegraphs.SQLObservationsInGraph`. The names query's per-Observation graph join (F-6) filters Observations, not Citations; leave it a join and say so in one comment line.
Check: `grep -rn "s.graph_id = ?" core/database/citations core/database/evidencegraphs` prints only `SQLObservationsInGraph`'s line and the names query's branch filters from F-6; Go tests pass; scripts/check-pr.sh PASS.

**F-2 — Drop the deleted `.deletedSubject(sourceId:)` from the rule and the doc** (from CON-1, MNT-3)
Files: `docs/catalog-deletes.md`, `.cursor/rules/catalog-deletes.mdc`.
Do: Copy the wording at `.agents/skills/add-catalog-delete/SKILL.md:111`.
Check: `git grep -n "deletedSubject(sourceId"` prints nothing outside `docs/deployment-plan/`.

**F-3 — Read the composer's linked Artifacts once per body and once per method** (from PRF-1)
Files: `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `CitationComposerModel.swift`.
Do: In the evidence column, read `model.artifacts` once into a local at the top of the view(s) that use it and pass that array to `isDisabled`, the PVSelect row builder (line 161), `artifactOptions` and `artifactPosition` (make those take the array). In the model, any method or computed property that reads `artifacts` more than once (`prepare` around 461–475, `hasNoArtifacts`/selected Artifact around 289–291, 975, 1116, 1151) binds it to a `let` once. Keep `artifacts` computed so it stays observed; do not add a stored cache.
Check: `grep -c "model.artifacts" CitationComposerEvidenceColumn.swift` is at most one per view body; Swift tests pass; scripts/check-pr.sh PASS.

**F-4 — One session accessor for the graph list; the linked-Artifacts read moves beside it** (from MNT-2, MNT-8)
Files: new `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` (add to the App target), `Features/Sources/SourcesListNavigation.swift`, `Features/Sources/SourcesModel.swift`, `Features/CitationComposer/CitationComposerModel.swift`, `macos/ProvenenciaTests/*` that call `SourcesListNavigation.linkedArtifacts`. Skill: `add-workspace-place`.
Do: `extension WorkspaceSession` with `var evidenceGraphsKey: CatalogQueryKey` (`.evidenceGraphsList(project: projectKey)`), `var evidenceGraphsHandle: QueryHandle<[CatalogEvidenceGraph]>?` (the one `queryHandle` lookup), and `func linkedArtifacts(graphID:artifacts:) -> [CatalogArtifact]?` (D-7's body as is: `nil` while no value, `[]` with no such graph, else `CatalogEvidenceGraph.linkedArtifacts(_:)`). Point `SourcesListNavigation.graphsBySource(in:)`, `graphListLoaded(in:)` (the line-69 lookup) and `SourcesModel.graphsHandle` at `evidenceGraphsHandle`; delete `SourcesListNavigation.linkedArtifacts`; the composer's `artifacts` calls `session.linkedArtifacts(...)` and its `evidenceGraphsKey` returns `session.evidenceGraphsKey`. Leave `CatalogEvidenceGraph.linkedArtifacts(_:)` in `GenealogyStore.swift`. `session.query(...)` warm calls (`SourcesModel.swift:140`, `SourcePageModel.swift:123`) use `session.evidenceGraphsKey` too.
Check: `grep -rn "evidenceGraphsList(project: session.projectKey)" macos/App` prints nothing; `grep -rn "evidenceGraphsList(project: projectKey)" macos/App` prints only the new accessor; `grep -rn "SourcesListNavigation.linkedArtifacts" macos` prints nothing. Supersedes D-7's placement in `SourcesListNavigation.swift`: note it in the Landed note.

**F-5 — One test helper for creating a Citation** (from MNT-4)
Files: `core/database/citations/citations_test.go` (and `delete_test.go` if it holds the same wrapper).
Do: Add a file-local `mustCreate(t *testing.T, c *database.Catalog, userID []byte, artifactID []byte, obs ...<the observation input type>) CreateResult` (match the parameter types the wrapper uses; take `LocatorJSON` from `validLocator`, and add a sibling or option only if a case needs a different locator). Move every copy in the package onto it.
Check: `grep -rn 'Action: "create_citation_with_observations"' core/database/citations/*_test.go` prints only the helper's line; Go tests pass.

**F-6 — Drop the redundant graph re-check from the person-names query** (from PRF-2, MNT-5)
Files: `core/database/citations/citations.go`.
Do: Remove `AND ` + `sqlInGraph` from both branches. In branch 1 drop the `so` join and put `AND s.graph_id = ?` on `s` (same join). Branch 2 keeps `so`. Binds per branch become graph, artifact (adjust `attachPersonNames` and the bind comment).
Check: "list counts and names the graph's Observations" passes; Go tests pass.

**F-7 — FakeStore graph create reports its Source for the citation reads** (from MNT-6)
Files: `macos/App/Platform/FakeStore.swift`, `core/database/evidencegraphs/evidencegraphs_test.go`.
Failing test first: extend `TestLinkWritesMoveCitationReads` with a Create case (and a Delete case) asserting what `CitationSources` the engine reports.
Do: Make the fake match what the test shows: pass `sourceID` on create (and delete, if the engine reports it); fix the doc comment at `FakeStore.swift:836-838`.
Check: Go test passes; Swift tests pass.

**F-8 — Landed note**
Files: the plan.
Do: Under Status, a Landed note listing: D-1/D-3 SQL now built from `evidencegraphs.SQLObservationsInGraph` (F-1, same-pass cleanup); D-7's session read moved to `WorkspaceSession+EvidenceGraphs.swift` (F-4); D-10's create (and delete if F-7 found it) also reports `citationSources`.
Check: scripts/check-pr.sh PASS (plan-docs).

## Round 1 — fix
Head `7efaa61c`. Gate PASS (orchestrator run). F-1 to F-8 done. Deviations: F-5 helper is `createCitation(c, userID, in, obs) (CreateResult, error)` since several cases assert on the error and need different `CreateInput` fields; F-7's new create/delete cases passed from the start (engine reports the Source on create, nothing on delete), so the fake's create was the only change.
CON-1 / CON-2 (CON had no Fix-now blocker or major of its own; CON-1's major is MNT-3's): read by the orchestrator. **Resolved**: `git grep "deletedSubject(sourceId"` outside `docs/deployment-plan/` prints nothing; the plan's Landed note lists the D-1/D-3, D-7 and D-10 differences. PRF-2/MNT-5, MNT-6 ride with their reviewers' round 2.

## Round 2 — review
Reviewers: MNT (MNT-1 blocker; MNT-2, MNT-3, MNT-4 majors), SEC+PRF (PRF-1 major). Not sent: CON (minors only), UX (no findings). Scope `e9f3f8a6...7efaa61c`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | (round 1) | Resolved | `evidencegraphs.SQLObservationsInGraph` is the one spelling; `sqlCitationsUsing`, `sqlUsedByGraph`, `observation_count` built from it |
| MNT-2 | MNT | major | (round 1) | Resolved | `WorkspaceSession+EvidenceGraphs.swift` is the only `evidenceGraphsList(` outside the registries |
| MNT-3 | MNT | major | (round 1) | Resolved | both docs say `.deletedSubject` |
| MNT-4 | MNT | major | (round 1) | Resolved | `createCitation` helper; the action string appears only in it and one audit assertion |
| PRF-1 | PRF | major | (round 1) | Resolved | one `model.artifacts` read in the evidence column; model methods bind once |
| PRF-2, MNT-5 | PRF, MNT | minor | (round 1) | Resolved | SEC+PRF round 2: binds graph, artifact per branch; same results |
| MNT-6 | MNT | minor | (round 1) | Resolved | orchestrator read `1b2ef52e`: fake create passes `sourceID`; Go test covers create and delete |

No new blockers from either reviewer.

## Result
Gate PASS · CI pending · open: none · follow-ups: none (S12-13 notes written into its map entry: graph-list id index, `noArtifactsCallout` link wording) · to user: — · CON miss: none
