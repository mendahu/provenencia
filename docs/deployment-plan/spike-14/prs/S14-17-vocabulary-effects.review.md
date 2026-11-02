# Review — S14-17 Vocabulary writes return their effects; dead RPCs go

**Verdict:** Fits with changes — Ready with changes (round 1, all findings answered; status Blocked — on S14-27, so round 2 runs on the refresh after S14-27 lands)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **M-1** One `vocabulary` list reaching every label key bends "One cache owns each list". **Answered 2026-11-02: split into two lists** (`source_vocabulary`, `property_vocabulary`). → D-5, D-6.
- **M-2** D-5 grew the `invalidations(for list:)` switch twelvefold right before S14-27 deletes it. **Answered 2026-11-02: S14-27 first.** Map reordered: S14-27 depends on S14-16, S14-17 depends on S14-27 and is Blocked on it. → D-6 uses S14-27's slot.
- Earlier answers recorded in the plan (2026-11-02): one PR, not split; vocabulary counts removed, not moved onto the recount; the type–field join stays unaudited and drops `user_id`.

## Writer's own calls

- D-6 keeps `.updatedMetadataField(id:)` / `.deletedMetadataField(id:)` for `sourceWorkspace` only: a list cannot tell a field relabel from a create, and a Source page following `source_vocabulary` would go stale on every type create.
- D-9 also deletes `ListSubjectTypes` and `ClearSubjectPosition` (their last Swift callers are the test-only methods the map deletes), and D-11 deletes their now-dead callees.
- `evidenceGraph` / `citationsByArtifact` do not follow `property_vocabulary` (M-3 declined in part, below).
- D-7: the Properties create flow applies each write's effects as it lands (X-3).
- S14-27's slot is named `CatalogQueryFollows.list(_:)` here and in the S14-27 scope entry; S14-27's plan may choose otherwise, and this plan is refreshed when it lands.

## Round 1

### M-1 · blocker · D-5 bends "One cache owns each list"
**Answer:** To user → split (2026-11-02). Fixed — D-5 replaces `ListVocabulary` with two families; D-6 maps each key to its family; Reach, Scale, T-2, T-5, T-10 updated.

### M-2 · blocker · The switch grows right before S14-27 deletes it
**Answer:** To user → S14-27 first (2026-11-02). Fixed — status Blocked — on S14-27; D-6 uses the slot; map row, order, dependency graph and S14-27's Context updated.

### M-3 · major · `vocabulary` doc false on day one; `evidenceGraph` and `citationsByArtifact` left out
**Answer:** Declined in part, with evidence. The composer place holds `evidenceGraph` on screen (`PlaceRegistry.swift` `.sourceCitationComposer` keys), so following `property_vocabulary` would reload the visible graph on every custom term — the frequent vocabulary write — at a census graph's size, a cost not measured. They stay Out with a structural-debt line, and the doc comment no longer claims every label key (D-6 names the keys each family reaches). Fixed: the claim.

### M-4 · major · Guidance sweep misses lines
**Answer:** Fixed — D-13 adds `add-workspace-place` line 176, the `sourceWorkspace` and `typeSuggestions` descriptor comments, the `workspaceNavCounts` doc, the `CatalogCounts` class doc, "One cache owns each list" and the effects README; STEP-8's `rg` extended.

### M-5 · major · `CatalogCountSummary` only forwards
**Answer:** Fixed — D-8 deletes it; the `countLine`s read `VocabularyPane.swift`'s `seededCount` / `userCount` / `pluginCount`; `badge(for:)` has one `default` arm.

### M-6 · major · `subjectvocab.PresentationFor` is already dead
**Answer:** Fixed — D-11 deletes it with `PlaceableTypes`; STEP-7's `rg` covers it.

### M-7 · major · Inline type-id lookups
**Answer:** Fixed — D-11: both helpers in `harness_test.go` on `mustRPC`; direct reads only where the whole list is walked; STEP-7 checks the files that name `GetPropertiesWorkspaceRequest`.

### M-8 · minor · Effects assertions in three places
**Answer:** Fixed — the reorder effects assertion moves into T-1.

### X-1 · major · T-1 names a fixture that does not exist
**Answer:** Fixed — T-1 starts from `sourceFixture`; the cardinality row lists its own steps (`citationFixture`, `CreateProperty`, `AssignSubjectTypeProperty` to `place`, `CreateCitationWithObservations`, `promoteInDir`, `UpdateProperty`); the Codebase fit claim is gone.

### X-2 · major · Seeded rows refuse edits
**Answer:** Fixed — every T-1 row works on rows it creates; the binding rows use the `person` type.

### X-3 · major · The Properties create flow's apply point
**Answer:** Fixed — D-7: the create's effects apply right after `createProperty`, each assign's after it; a thrown assign leaves the create applied with `formError`; T-9 `createAppliesEachWriteAsItLands`; §4 corrected (one mutation after the loop today).

### X-4 · major · `TestListProperties` covers kept RPCs
**Answer:** Fixed — D-11 moves its two cases into `TestGetPropertiesWorkspace` and keeps the `ListConnectRules`, locked-binding and palette assertions.

### X-5 · major · The SQL in "missing position invalid" has no access path
**Answer:** Fixed — D-11: `s.c.DB()` then `Exec`, outside `writes.Call`, because `writes.Run` rolls back a closure with no changes.

### X-6 · major · `nextWriteEffects = [.vocabulary]` does not compile
**Answer:** Fixed — every test uses `CatalogWriteEffects(lists: [...])`.

### X-7 · major · Swift checks need a Mac
**Answer:** Fixed — a line above the Steps and a Stop condition: finish STEP-1 and stop when `protoc` or `xcodebuild` is missing.

### X-8 · minor · Four handlers pass no `UserID` to `writes.Op`
**Answer:** Fixed as an Out line with a follow-up task.

### X-9 · minor · T-5 order and T-6 fixtures
**Answer:** Fixed — T-5 asserts per-kind membership (no order: the switch is gone); T-6 names its `CatalogSource` fixture.

### X-10 · minor · STEP-3 too big
**Answer:** Fixed — store and callers (STEP-4) split from the deletions (STEP-5), with a check after each.

### X-11 · minor · No base-check command
**Answer:** Fixed — §11 gives `git diff a5e057b origin/main --stat -- api core macos/App`.

### X-12 · minor · Smaller gaps
**Answer:** Fixed — D-8 names the `WorkspaceView` argument; `badge(for:)` `default` arm; the `enrichSourceProto` comment text in D-9; the reserved pair's place and comment in D-9; the structural-debt line's text in STEP-8. §4 names `GoStore.mapOriginCounts`.
