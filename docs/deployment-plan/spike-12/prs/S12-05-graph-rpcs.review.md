# Review — S12-05 Graph RPCs and graph list

**Verdict:** Fits — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None.

## Writer's own calls
- Seven RPCs named `…EvidenceGraph…`, numbered 103 – 109, all in one handler file `evidence_graphs.go` (D-1). Link and unlink stay as their own RPCs beside `Update`, for S12-13's Artifact card.
- The graph list RPC is project-wide with no Source filter; Swift readers filter (D-2). The cache key is `evidenceGraphsList`, revalidated when hidden (D-11).
- Only Create returns the graph; Update, Delete, Link and Unlink return effects only (D-3).
- The unlink preflight is a new read RPC, `GetEvidenceGraphUnlinkImpact`, returning one `DeleteImpactGroup` per blocking Artifact; its via is `evidencegraphs.artifact_in_use` (D-4, D-8, M-5).
- `CitationsUsingArtifactTx` now returns `[]deleteimpact.Listed` (id and ref) instead of ids only, reversing S12-03 D-13's ids-only call (D-5, from M-1).
- `PVDeleteImpactControls.isExtraGate` is deleted and `CatalogDeleteImpactGate` becomes `CaseIterable` (D-9, from M-2).
- All copy in §6 Strings: five error sentences, the last-graph gate title and body, two delete-confirm headings.
- `parseIDs` returns nil for an empty list; the three existing loops move onto it (D-6).

## Follow-ups (not this PR)
- M-8: S12-12 uses `createEvidenceGraph`'s returned graph only for its id and navigation; display reads the list key. Written into the plan's Next adds.
- M-9: a check that every user-kind `apperr` code has a `L10n.Errors.message(code:)` case. Not logged as an idea file by this plan; offered to the user.

## Round 1

### X-1 · blocker · Steps 2 and 5 – 8 cannot be checked off macOS / without protoc
**Answer:** Fixed — Risks row and a Stop condition: no proto tools → stop after STEP-1; generated files are never hand-edited; Xcode checks reported as not run.

### X-2 · major · `graphListInvalidations: 1` is not a field of `ListCase`
**Answer:** Fixed — T-13 says `ListCase` gains the field (1 for the graph case, 0 for the other five) and the test asserts it; STEP-7 names every switch.

### X-3 · major · PR number and status before the PR exists
**Answer:** Fixed — STEP-9 says it is the last commit, after the draft PR is open, and names the map status `Landed #N`, as S12-04 did. The plan is marked Planned before implementation.

### X-4 · major · `FakeStore.createArtifact` signature in STEP-5
**Answer:** Fixed — STEP-5 adds the ignored `graphIDs` parameter; STEP-6 gives it behavior.

### X-5 · minor · Test choices left open
**Answer:** Fixed — T-1 names its builder, T-2 names the fixture helpers, T-5 names its fixture, AC-2.3 checks one revision, and the proto comment says Create returns links in request order.

### X-6 · minor · `CatalogQueryKey` switches not all named
**Answer:** Fixed — STEP-7 lists the case, `Kind`, `kind` and `project` arms, and places the `Spec` after `sourceGraphProgress`.

### X-7 · minor · Copy comments, xcstrings check timing, docs wording
**Answer:** Fixed — comments given; STEP-5 says the localizable check may fail until STEP-8; the `catalog-deletes.md` bullet is written out. pbxproj ids stay with `add-swift-test`.

### X-8 · nit · `parseIDs` nil vs empty
Logged; no caller distinguishes them.

### M-1 · blocker · `ListedGroup` re-reads refs the domain query already selected
**Answer:** Fixed — D-5: `ListedGroup` takes `[]Listed` rows with id and ref; `CitationsUsingArtifactTx` returns them. The per-row ref SQL, the bucket rule, its stop condition and T-1's vocab case are gone.

### M-2 · major · `isExtraGate` has no app caller
**Answer:** Fixed — D-9 deletes it; `CatalogDeleteImpactGate` is `CaseIterable` and T-15 loops `allCases`.

### M-3 · major · New patterns not in the skills
**Answer:** Fixed — STEP-9 adds a line to `add-catalog-delete`, the `deleteimpact` README and `add-ffi-handler`.

### M-4 · minor · `groupsToProto` wrapped in a one-element slice
**Answer:** Fixed — `groupToProto` *(new)*, which `groupsToProto` calls (D-4).

### M-5 · minor · Via key looks like a column
**Answer:** Fixed — renamed to `evidencegraphs.artifact_in_use`, matching the refusal code, with a doc comment (D-8, D-13).

### M-6 · minor · FakeStore links half kept
**Answer:** Fixed — D-14: `createArtifact` always reports `.evidenceGraphs`; `deleteArtifact` and `deleteSource` drop links and graphs; T-12 covers it.

### M-7 · minor · Name the existing test helpers
**Answer:** Fixed — T-2 names `newFixture`, `fixture.newArtifact`, `fixture.subjectAndCitation`, `fixture.createGraph`.

### M-8 · nit · Create's row overlaps the list reload
**Answer:** Logged in Next adds for S12-12.

### M-9 · minor · No check that every code has copy
**Answer:** Declined for this PR (preparatory work, not this PR's files); offered to the user as a follow-up.

## Round 2

Maintainability: **Fits**, no new findings; the round-1 answers hold. Two points it noted without filing them:
- S12-12 will compute the removed set in Swift (stored minus checked), as the map's scope entry asks; Go skips unlinked Artifacts, so preview and write agree.
- `ListedGroup` repeats `collectGroups`' six-line projection loop; both go through `projectListed` and `listedCap`. Left as is.

Executability: **Ready with changes**.

### X-9 · major · `isExtraGate` deleted in STEP-5 but its test edited in STEP-8
**Answer:** Fixed — STEP-5 deletes it and rewrites the test over `allCases`; STEP-8 keeps `lastGraphGateCopy` and the via rows.

### X-10 · major · STEP-5 Files omit `FakeStore.swift` and `L10n.swift`
**Answer:** Fixed — both added, with `PVDeleteImpactTests.swift`.

### X-11 · major · How Source B is made in handler tests
**Answer:** Fixed — D-16: same project, `ListSourceTypes` → `CreateSource` → `CreateArtifact`.

### X-12 · minor · Arrange order for multi-write tests
**Answer:** Fixed — D-16: arrange with direct handler calls in the test body; `runRPC` only for the call under test.

### X-13 · minor · Header status and map link
**Answer:** Fixed — Status Planned, Review line linked, map row links the plan with status Planned.

### X-14 · nit · "Last method is 102"
**Answer:** Fixed — "highest method number".

Final verdict: **Fits — Ready with changes** (all changes applied). No blocker open; nothing to the user.
