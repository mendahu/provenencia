# Review — S14-17 Vocabulary writes return their effects; dead RPCs go

**Verdict:** Fits with changes — Ready with changes (round 2 on the refresh against S14-27, 2026-11-03; every finding Fixed; status Planned)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet), fresh each round

## To user

- **M-1** One `vocabulary` list reaching every label key bends "One cache owns each list". **Answered 2026-11-02: split into two lists** (`source_vocabulary`, `property_vocabulary`). → D-5, D-6.
- **M-2** D-5 grew the `invalidations(for list:)` switch twelvefold right before S14-27 deletes it. **Answered 2026-11-02: S14-27 first.** Map reordered: S14-27 depends on S14-16, S14-17 depends on S14-27 and is Blocked on it. → D-6 uses S14-27's slot.
- Earlier answers recorded in the plan (2026-11-02): one PR, not split; vocabulary counts removed, not moved onto the recount; the type–field join stays unaudited and drops `user_id`.

## Writer's own calls

- D-6 keeps `.updatedMetadataField(id:)` / `.deletedMetadataField(id:)` for `sourceWorkspace` only: a list cannot tell a field relabel from a create, and a Source page following `source_vocabulary` would go stale on every type create.
- D-9 also deletes `ListSubjectTypes` and `ClearSubjectPosition` (their last Swift callers are the test-only methods the map deletes), and D-11 deletes their now-dead callees.
- `evidenceGraph` / `citationsByArtifact` do not follow `property_vocabulary` (M-3 declined in part, below).
- D-7: the Properties create flow applies each write's effects as it lands (X-3).
- S14-27 landed its slot as `CatalogQueryFollows.list(CatalogList)` and `.structure` (#693, stacked base), as D-6 names it.
- Refresh (2026-11-03): D-11 extends to the test sites the first pass missed (`TestSubjectTypePropertiesAndRegistry`, `TestReorderSubjectTypeProperties`'s `list` helper, `TestSubjectPositions`' clear half), in the shape D-11 already used.
- M-10's fix (D-14) is a keyed read, not a third list: it is not a vocabulary family, so the user's "two lists" answer stands.

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

## Round 2 — refresh against S14-27 (2026-11-03)

Plan refreshed against `s14-27-lists-through-descriptors` @ `f1b3fe65` (S14-16 merged into `main` @ `42eb321c`). Before review, a §4 sweep corrected: the `ListVocabulary` test files (three were missing), `citations_test.go` line 75, the dead-RPC test sites (`TestSubjectTypePropertiesAndRegistry`, `TestReorderSubjectTypeProperties`, `TestSubjectPositions`, dedicated tests), the S14-27 Registry and Follow-slot rows, the Docs and Swift tests line numbers, and the sidebar's `catalogCounts` path. Fresh reviewers then read the refreshed plan.

### M-9 · major · The Go display cache would still drop on Source-vocabulary writes it holds nothing from
`graphcache.OnCommit` (`commit.go:25–31`) drops labels, the vocabulary, every graph's derived memos and the match index, and sets `warmNeeded`; `graphcache.Vocabulary` holds only Properties, terms, types and bindings, and no non-test graphcache, canonlinks or conclusionsummaries file reads Source types or fields. D-5 as written made every Source type or field write rebuild that at the map's Scale.
**Answer:** Fixed — D-5: `OnCommit` drops only on `ListPropertyVocabulary`; Edges/Assumptions reworded; T-3 asserts both sides (`warm_test.go`, `graph_test.go`). Verified: `rg -l 'source_types|metadatafields' core/database/graphcache core/database/canonlinks core/database/conclusionsummaries` lists only test files.

### M-10 · major · `propertyTerms` would reload every warmed key on any Property-family write
Today `.createdPropertyTerm(propertyId:)` names one key; `CatalogPropertyTerm` holds no Property label, binding or type. The repo's keyed-read mechanism (`effects.Read`, `KeyTable`, `ReadPath` with `field(…)`; `ReadEvidenceGraphs`) keeps the precision with Go declaring it.
**Answer:** Fixed — new D-14 (`ReadPropertyTerms`, `KeyTable` `properties`, `Reads` on `property_terms`; Swift `CatalogRead.propertyTerms`); D-6 drops `propertyTerms` from the `property_vocabulary` family and gives it a `.reads(.propertyTerms)` refresh; AC-1.4, T-1, T-5, T-10, Scale, Codebase fit, Scope, STEP-1/2/3 and D-13's README line updated.

### X-1 · major · STEP-1's check cannot pass: the Swift parity test fails until STEP-2
**Answer:** Fixed — STEP-1's check runs `./core/... ./api/ffi/handlers/...` and says the `api/ffi` parity rows ("catalog-wide lists", "catalog reads") go green in STEP-2; D-5 names `write_effects_test.go` 77–78 (`TestWriteEffectsProtoLists` wants `"property_vocabulary"`).

### X-2 · minor · The family for table-less test commits was left open
**Answer:** Fixed — D-5: the table-less commits use `ListPropertyVocabulary`; `Canon.CommitVocabulary` keeps its name with a Property-side doc.

### X-3 · minor · §4 names two helpers that read no deleted RPC
**Answer:** Fixed — the row names `subjectFixture` (`ListSubjectTypes`) and says `mustRPC` is the reader the new helpers use; `sourceFixture` (reads `ListSourceTypes`, which stays) is dropped.

### X-4 · minor · The `typeSuggestions` / `propertyTerms` `invalidation` closures left a choice
**Answer:** Fixed — D-6: `typeSuggestions`' closure stays as it is (its `default` arm stays for exhaustiveness); `propertyTerms`' closure is the constant D-6 already gave.

### X-5 · nit · Status still read Blocked
**Answer:** Fixed — Status `Planned`.

