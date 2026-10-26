# Review — S11-49 Kinship links in the graph cache

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- The bound, the linear dedupe and the `subject` role move from S11-39 into this PR (they live in `core/kinship` and exist only on #567's branch); the map's S11-39 entry no longer lists them.
- Results are derived per pivot on first read, not at install (D-4).
- The package is `canonlinks`, not `impliedlinks` (M-5).
- `displays` stays off the new memo slots (Out line): its drop is tied to install and the header deps walk.
- `invariants_test.go:153` switches its role literal to `"witness"`, `lifespan_test.go:68` to `subject`.

## Round 1

### M-1 · blocker · Kinship drops hard-coded to one hop inside `install`/`removeNode`
`loadChunk` installs on every first load, so dropping in `install` throws away fresh results during a walk; the commit path already has `roots` plus owner-registered deps (`SetDisplayDeps`).
**Answer:** Fixed — D-5: memos drop only in `OnCommit` (roots plus the slot's registered `Deps`), on a vocabulary commit for `DropOnVocabulary` slots, and in `clear`; `install`/`removeNode` do not drop. D-6: `canonlinks` owns the one-hop `Deps`. AC-2.7 and T-7 cover a first load keeping results.

### M-2 · major · A third hand-rolled per-node `map[string]any` memo
**Answer:** Fixed — D-5 adds `RegisterNodeMemo` slots. `details` has no caller outside `graphcache` (grep), so it is deleted rather than moved. `displays` stays: Out line with the reason.

### M-3 · major · `NewReader` would build a third props-by-id map
**Answer:** Fixed — D-8: `Vocabulary.PropertiesByID`, `TypesByID`; `promotealign.propertiesByID` and `conclusiondetails`' inline map move onto them (STEP-4).

### M-4 · major · Docs the PR makes false are missing from STEP-8
**Answer:** Fixed — STEP-9 adds `docs/ideas/kinship-view.md`, the add-match-rule "Add a kinship pattern" step for `canonlinks`, and kinship-shapes' "One pure package" paragraph.

### M-5 · minor · `impliedlinks` covers two concepts
**Answer:** Fixed — D-6: `core/database/canonlinks`; map and kinship-shapes renamed.

### M-6 · minor · The moved `keptTerm` stays a copy of the kept-value read
**Answer:** Fixed — D-8: `Node.Kept`; `conclusionheaders.kept` and `promotealign.keptTerm` go.

### M-7 · minor · Emptying kinship in `dropDerived` also drops on alias and name-format warms
**Answer:** Fixed — D-5: `DropOnVocabulary` slots empty in `OnCommit`'s `ListVocabulary` branch, not in `dropDerived`.

### M-8 · minor · Two canon test builders
**Answer:** Fixed — D-12: `graphcache/graphcachetest` with `OpenCatalog` and `Canon`; `TestCommitOfAnAssociationLoadsEndpoints` moves onto it.

### X-1 · blocker · `docs/matching-model.md` precedent hunks do not apply
**Answer:** Fixed — Precedent: only the knob row, with its anchor; STEP-9 adds the glossary rows.

### X-2 · blocker · "promotealign tests do not exist on main" is false; `subjectprofile` literals
**Answer:** Fixed — §4 "Test terms and literals" row names which `spouse` literals stay (relationship term) and the two that change (D-3).

### X-3 · blocker · STEP-5 left unused variables
**Answer:** Fixed — D-7 deletes `selfKind` and `fromKind`; STEP-6 names two `canonKeys` and two `loadTerms` call sites.

### X-4 · blocker · `Reader.For` semantics open
**Answer:** Fixed — D-9: filter to `id` as an end, order, no dedupe and why, how parents and children are found; D-4: neighbors are already unmerged.

### X-5 · major · Verify does not know a result's pivot kind
**Answer:** Fixed — D-10: `pivotResult.kind`.

### X-6 · major · Test builder underspecified
**Answer:** Fixed — D-12 lists every method, the `ends` keys, the id and ref scheme, and the catalog helper.

### X-7 · major · `KinshipLoadedForTest` placement
**Answer:** Fixed — replaced by the slot API (`Memo` reports presence); no test-only accessor.

### X-8 · major · `dropKinship` edges
**Answer:** Fixed — moot: D-5 drops on commit only.

### X-9 · minor · STEP-7's files repeat STEP-6's
**Answer:** Fixed — STEP-8 says both files were created in STEP-7.

### X-10 · minor · Checks
**Answer:** Fixed — STEP-1 says to revert the 1000 bound; STEP-3 runs `scripts/check-pr.sh` (fts5, Xcode) and allows the comment in `swift_vocabulary_test.go`.

### X-11 · minor · Metadata
**Answer:** Fixed — Status and Review set after this log; add-match-rule line 33 named.

Nits (logged): `M-9 · D-10 keeps Rules on each result for a test-only variation` · `X-10 · grep -c exits 1 when it prints 0`.

## Round 2

Executability (Sonnet): Ready with changes, no new blocker.
- X-12 · major · `EdgeOf` cannot reach `Node.Kept` without a property id — **Fixed:** D-7: `EdgeOf` takes `eventTypeProp`; `EventTypeProperty` resolves it once; first rank-1 kept row.
- X-13 · major · `openCatalog` has callers in `vocab_test.go` and `evidence_graph_walk_test.go` — **Fixed:** D-12 and STEP-5 move every caller.
- X-14 · major · Verify needs `SetTruth` registered — **Fixed:** D-12: `graphcachetest` blank-imports `autoreconciler`.
- X-15 · major · Tests cannot reach the slot — **Fixed:** D-5 defines `MemoSlot`; D-6 names `kinshipSlot` and makes the tests internal.
- X-16 · major · A middle person can be found twice — **Fixed:** D-9: parents and children are a set, relationship edges first, then birth links.
- X-17 · minor · `TypesByID` unused — **Fixed:** dropped; the Reader reads `Link.NeighborTypeKey`; S11-39 replaces `typeKey`.
- X-18 · minor · Nil `Check`, edge `Associations`, `Memo` guards — **Fixed:** D-4 and D-5 state them.

Maintainability (Opus): Fits with changes. M-1, M-4, M-5, M-7 hold.
- M-10 · blocker · Slot types copy the display registration — **Fixed:** D-5: `DisplayCheck` renamed `MemoCheck` and shared; `MemoDeps` returns an error so the header deps fit later; `Cached` replaces `LoadedForTest`. Moving `displays` itself stays an Out line.
- M-3 (incomplete) — **Fixed:** D-8: `Vocabulary.TypeOrder` uses `PropertiesByID`.
- M-6 (incomplete) — **Fixed:** D-8: `Rank1` moves beside `Kept`; D-7 uses both. `autoreconcile.ReasonKept` instead of the literal: logged (Out line).
- M-8 (incomplete) — **Fixed:** D-12: `TestCommitReloadsASharedEndpointOnce` moves onto the builder too.

## Verdict
Round 2 is the cap. No blocker open; no To user item. Status Planned.
