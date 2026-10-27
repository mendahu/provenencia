# Review — S12-21 The canvas keeps its graph and shows its failures

**Verdict:** (round 2 pending)
**Rounds:** 1 so far. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- The Source page's `prepare(for:)` is removed, not kept as the working rule's exception (D-4). The host's `.id(sourceId)` builds one model per Source, so its guard could never fire.
- Master–detail places (Source types, Metadata) declare `viewIdentity: nil` instead of today's `typeId` / `fieldId` (D-2). A row click must not remount the list.
- The deleted-graph state outranks a read error, and a read error outranks every existing canvas notice (D-7).
- `evidenceGraphPresence` reads `.unknown`, not gone, while the graph list refetches, and `canCite` follows it. During that window the No-artifact callout no longer flashes either, which changes `canCite` behavior (D-5).
- The composer's Artifact switch still shows the new Artifact after a failed Citation load, unlike its sibling, which returns early (D-8).
- Copy for the two new callouts (D-10): "This evidence graph no longer exists" and "This evidence graph could not be read", with **Go to Source page** and **Read again**.
- T-1 observes which cache keys the canvas model loads, not on-screen text (D-9).

## Round 1

### X-1 · major · No Swift check or test can run in the review environment
The reviewer ran on Linux, where there is no `xcodebuild`.
**Answer:** Logged. This plan is implemented on macOS, like every Swift PR (`scripts/check-pr.sh` `xcode-test`).

### X-2 · major · T-1 leaves the harness choices open
**Answer:** Fixed. D-9 now names:
- `userID`, `sessionDisplayName` and `CatalogCounts(projectDir:store:)`
- the window size, and who holds the window
- a `makeCanvasStore()` that copies `EvidenceGraphModelTests.makeStore`'s vocabulary lines
- the two `CatalogSource` rows
- that `projectUuid` is not set

### X-3 · major · T-5 cannot use `seedCitation` for `cit-2` on `art-1`
**Answer:** Fixed. T-5 builds `cit-2` as a `CatalogCitation` literal. It now makes the model with `citationID: "cit-1"`, an edit entry with no dialog, so the switch does not stop at `pendingLeave`.

### X-4 · major · D-10 gives no translator comments
**Answer:** Fixed. D-10 now carries the five comments, and each `.xcstrings` entry copies the shape of `evidenceGraph.noArtifact.action`.

### X-5 · minor · Temptation to drop `.task(id:)`
**Answer:** Superseded by M-3. D-11 changes all four to `.task`. The Out row added for X-5 was removed.

### X-6 · minor · T-5 and T-6 mechanics
**Answer:** Fixed. Both tests now use `model.session`, and T-6 seeds `occupationObservation`. A local `TestFailure` is kept, because FakeStore's `…Error` hooks need an error value.

### X-7 · nit · Skill step 4 still says "deep id"
**Answer:** Fixed in STEP-7.

### X-8 · nit · `SourcePageModelTests` listed as likely to break
**Answer:** Fixed. Removed from §8.

### M-1 · blocker · "Graph missing from the list" would have two rules in one model
**Answer:** Fixed. D-5 adds `EvidenceGraphPresence` and `WorkspaceSession.evidenceGraphPresence(id:)`, and both `canCite` and `graphIsGone` read them. AC-3.2 and T-4 (e) cover the refetch window.

### M-2 · blocker · `ConclusionDetailPage` keeps its own `.id(entityId)`
**Answer:** Fixed. D-11 deletes it, and §4, Reach and Deletes were corrected. STEP-2 now checks `rg '\.id\(entityId\)'`.

### M-3 · major · `.task(id: <same id>)` stays, and the skill teaches it
**Answer:** Fixed. D-11 changes four tasks to `.task`, STEP-7 rewrites skill line 163, and STEP-2 and STEP-7 check with `rg`.

### M-4 · major · Callout order is a seven-arm `if/else` in the view
**Answer:** Fixed. D-7 adds `EvidenceGraphCanvasNotice` and `EvidenceGraphModel.canvasNotice`, keeping today's order, and T-7 tests the order. AC-3.4 is now unit-tested.

### M-5 · major · Four copies of the composer's failed-load catch
**Answer:** Fixed. D-8 adds the private `showingLoadFailure(_:)` with four callers. STEP-6 checks that one `fields.error = L10n.Errors.message` assignment remains.

### M-6 · major · `sync(from:sourceID:)` keeps a guard D-4 makes always true
**Answer:** Fixed. D-4 drops the parameter and the guard at five call sites.

### M-7 · minor · `presentation(for:)` becomes a forwarder only tests call
**Answer:** Fixed. D-3 deletes it, and the tests' private helper calls `place(for:)`.

### M-8 · minor · `graphReadError` copies `connectRulesError`
**Answer:** Fixed with a private model helper, `errorMessage(_:)` (D-6). The writer did not add a `QueryHandle` accessor, because that would put `L10n` into the session layer.

### M-9 · minor · A third copy of the window-hosting setup in tests
**Answer:** Declined. `macos/Provenencia.xcodeproj/project.pbxproj` has no `PBXFileSystemSynchronizedRootGroup`, so a new helper file needs hand-written pbxproj entries until S12-31. T-1 is the only new caller. An Out row records it.

### M-10 · nit · T-2 repeats the expectations STEP-1 updates
**Answer:** Fixed. STEP-1 deletes the scattered `deepId` assertions, and T-2 is the one table test.

## Round 2

Fresh reviewers checked the round-1 answers and the new plan text.

### X-9 · major · STEP-2 check pattern can never pass
`sourceID: sourceID)$` also matches `SourcePageContext.swift:104`, a line this PR keeps.
**Answer:** Fixed. The alternative is now `sync\(from: [a-zA-Z]+, sourceID:`.

### X-10 · minor · Empty overlay lives in `EvidenceGraphDocumentBody`
**Answer:** Fixed in §4 and D-7.

### X-11 · minor · Choices left open in the new callouts
**Answer:** Fixed. D-7 now copies `noArtifactCallout`'s call shape and says which view carries each id. D-5 spells out `session.evidenceGraphPresence(id: graphID)`.

### X-12 · minor · Read-error flash before the gone notice
**Answer:** Fixed. §11 has a risk row saying the order is deliberate and must not be "fixed".

Nits fixed: T-7 moved after T-6; T-2 names the composer and Promote location builders to reuse.
