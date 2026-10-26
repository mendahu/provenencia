# Review — S11-39 Matching reads kinship shapes

**Verdict (re-plan):** Fits with changes — Ready with changes. No blocker open; status Planned.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None. The re-plan follows the slice 9 re-cut and S11-49/S11-50's landed code; every product call was already made (2026-10-26).

## Writer's own calls
- Close #567 and re-plan against `main` (user agreed), porting named commits by hand.
- `loadCanon` calls `Reader.For` per walked person after the walk and adds only links whose two ends are walked handles (D-7).
- No key interning; `linkSet.add`'s share is measured instead (D-8, M-5).
- M-7 declined (below).

## Re-plan round 1

### M-1 · blocker · D-8 added a second type-id → kind lookup
**Answer:** Fixed — D-8: `Vocabulary.TypeKeysByID`; `(*profiler).kind` is the one lookup for `loadCanon` and `checkFixed`; `typeKey` deleted; STEP-5 names the files.

### M-2 · major · A canon link read from both ends stores its shape twice (also X-5)
**Answer:** Fixed — D-1 `Recording.Merge` keeps shapes distinct; D-5 `addImplied` is the one entry for implied links; T-4 row.

### M-3 · major · `Shapes` + `OnlyImplied` copied as a field pair onto four carriers
**Answer:** Fixed — D-1 `kinship.Recording`, embedded in `Bridge`, `CanonEdge`, `adjLink`, `sideLink`; D-2 `CountedOnce(a, b Recording, paired)`.

### M-4 · major · Skill and context docs left describing the deleted merge (also X-10)
**Answer:** Fixed — STEP-8 rewrites add-match-rule lines 32–33, kinship-shapes § Counting once and the map's working rule; `matching-model.md` paragraph is added after line 28 with the "implied link" and "shape" glossary rows.

### M-5 · minor · Key interning unmeasured
**Answer:** Fixed — dropped; STEP-7 reports `linkSet.add`'s share.

### M-6 · minor · `linkSet` in a file named for terms
**Answer:** Fixed — `promotealign/links.go`.

### M-7 · minor · `Shape.Pivot` a third name for `Event`/`Through`
**Answer:** Declined — embedding `Shape` in `Implied` changes S11-49's landed `Implied` fields and their readers (`canonlinks/reader.go`, its tests, `kinship_test.go`) for no behavior here; `Implied.Shape()` is the one view.

### X-1 · blocker · Two tests named in the wrong package
**Answer:** Fixed — §4 Tests row; STEP-4 runs `./core/graphalign/` too.

### X-2 · blocker · CI cannot report budget figures
**Answer:** Fixed — AC-4.1, D-12, DoD: local `-v` figures in the PR body; CI pass/fail.

### X-3 · blocker · `typeKey` has callers outside `canon.go`
**Answer:** Fixed — with M-1.

### X-4 · major · `For` call order undefined (`walked` is a map)
**Answer:** Fixed — D-7: `walkedPeople` slice in wave order; "walked handles" means keys of `handles`.

### X-6 · major · Tree counts wrong
**Answer:** Fixed — AC-4.1: 4,095 persons, 4,094 parent links, 2,047 marriages.

### X-7 · major · STEP-4 had no failing test
**Answer:** Fixed — STEP-4 ports the `kin…` helpers and the AC-1.2, AC-1.4, AC-1.6 subtests first.

### X-8 · major · `LinkSetForTest` signature
**Answer:** Fixed — D-5 gives it with `rules`; per-link association set kept.

### X-9 · major · Ported `spouse` roles not all listed
**Answer:** Fixed — Precedent: every marriage role `spouse` → `subject`, `lookupTree` and the tree ids named.

### X-11 · minor · Smaller choices
**Answer:** Fixed — STEP-2 runs the purity guard; STEP-3/4/5 skip `TestLargeSourceBudgets` until STEP-7; D-3 quotes the closure and builds it once per state; the branch is cut from `main`. Canon-side associations: `CanonEdge` carries none, and the set clones them as for any link (logged, no change).

## Re-plan round 2

Executability (Sonnet): Ready with changes; every round-1 answer holds.
- X-12 · blocker · `checkFixed` has no profiler, so it cannot use `profiler.kind` — **Fixed:** D-8: the profiler keeps `typeKeys`; `checkFixed` reads `g.Vocabulary().TypeKeysByID()`; `loadCanon`'s `kinds` stays `map[string]string`; no caller signature changes.
- X-13 · minor · STEP-8 named `deployment-plan.md` twice — **Fixed:** one entry.
- X-14 · minor · T-4's implied rows — **Fixed:** built from recorded edges so `Derive` produces them.

Maintainability (Opus): Fits with changes; M-1 – M-6 answers hold, M-7 decline sound.
- M-8 · blocker · D-3's `paired` closure copies `(*state).placedOf` — **Fixed:** D-3 passes `st.placedOf`.
- M-1 gap (`checkFixed` has no profiler) — **Fixed** with X-12 (reads `TypeKeysByID` directly; no signature change).
- M-2 note (pivots cloned on a hit, not a miss) — **Fixed:** D-5 clones pivots on a miss too.
- M-4 note (the map's S11-39 scope entry still named the old merge) — **Fixed** in the same commit.

Round 2 is the cap. No blocker open; no To user item.

## First plan (2026-10-26, shipped as #567, closed in the slice 9 re-cut)

**Verdict:** Fits with changes — Ready with changes. No blocker open.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- None open. Answered before round 1 (user, 2026-10-26): count a fact once when both sides imply it through the same pattern and the pivot is paired (chosen over accepting both credits and over keeping the stronger one); Source-side implied lines reach the evidence sheet now.

## Writer's own calls

- `Shape` and `CountedOnce` live in `core/kinship`; `graphalign` imports it.
- A Source with both a spouse relationship and a marriage, against a tree with only the marriage, counts once (D-4, AC-2.5).
- Two recorded links keep today's double credit.
- The duplicate check (`newPairSegments`) sees implied Source links but does not apply the count-once rule.
- No memo for the canon's implied links (31 ms at 8,190 parent links).

## Planning prototype

The numbers in §2 and §11 come from a throwaway prototype at `7efa02f` (not in the repo): loader derivation, the placement-aware rule in `corresponding`, scratch goldens across every shape pair both ways, pure `graphalign` fixtures with the invariants, the evidence query, and `TestLargeSourceBudgets` with and without derivation. Every existing `graphalign`, `promotealign`, `claimevidence` and `api/ffi` test passed with it.

## Round 1

### M-1 · blocker · The count-once rule is a private graphalign method, but S11-40 must apply the same rule in the evidence query
**Answer:** Fixed — D-2 makes it `kinship.CountedOnce` with a `paired` function; D-3 calls it from `corresponding`; T-2a tests it purely; §6 Next adds names S11-40's call.

### M-2 · blocker · Derive-and-merge written twice now, three times after S11-40, and the copies disagree
**Answer:** Fixed — D-5 adds one helper, `withImplied` over `sideLink` in `terms.go`, which registers a miss; D-6 and D-7 call it; T-4 covers two marriages of one couple.

### M-3 · minor · `Implied bool` reuses the name of `kinship.Implied`
**Answer:** Fixed — `OnlyImplied` / `onlyImplied` throughout (D-3).

### M-4 · minor · Doc comments in edited files become false
**Answer:** Fixed — §6 Go core names the `Layer`, `CanonEdge` and `adjLink` comments; STEP-8 says loaders copy by field name and never convert.

### M-5 · nit · `marriageShapes` picks its mode from a string
**Answer:** Fixed — STEP-2 adds `implied` and `canonOf` helpers composing with `bridge` and `edge`; T-2 names its shapes from them.

### M-6 · nit · `Shape.Pivot` is a third name for `Event` / `Through`
**Answer:** Fixed — D-1 puts the mapping in `Shape`'s doc comment.

### X-1 · blocker · STEP-8's "no caller" sentences do not exist
**Answer:** Fixed — §6 Deletes is "None"; STEP-8's check greps for the new terms and for the removed "the fields of `graphalign.Bridge`" claim.

### X-2 · minor · Depends on disagrees with the map
**Answer:** Fixed — header Depends on is S11-38 only; S11-41 moves to Context; §9 says the map's Depends on column stays.

### X-3 · major · Mixed recorded / implied pairs under D-3 and D-4
**Answer:** Fixed — D-4 decides it (skip when either is only implied), AC-2.5 and T-2's (both, marriage) and (marriage, both) rows pin it; verified in the prototype (2.875 undated, equal to event-only).

### X-4 · major · New implied canon edges not registered
**Answer:** Fixed — `withImplied` registers each miss (D-5), T-4.

### X-5 · minor · Copy and aliasing in the loaders
**Answer:** Fixed — D-5 clones implied ends and associations; T-4 checks non-aliasing.

### X-6 · major · `marriageShapes` cannot express T-2
**Answer:** Fixed — T-2 lists ids, values, metas, the dated value, the shapes, the fixed rows and each equality.

### X-7 · major · No way to build the baseline rows
**Answer:** Fixed — every Then names a sibling row built in the same subtest (D-11); the no-implied-link equality is pure (T-2's `event` shape); prototype numbers are stop-condition references only.

### X-8 · minor · `Via.Key` written as a string
**Answer:** Fixed — AC-1.5 gives the struct; the prototype printed `relationship·grandchild·related_to` for Cal's row.

### X-9 · minor · Implied links reach the duplicate check
**Answer:** Fixed — the Out row says they reach `newPairSegments` and credit there; accepted.

### X-10 · major · STEP-3's check does not exercise what it changes
**Answer:** Fixed — STEP-3 and STEP-4 run the whole `graphalign` and `promotealign` packages; §4 Tests says no existing test has a kinship shape.

### X-11 · minor · STEP-7's run on `main` is unspecified
**Answer:** Fixed — STEP-7 uses `git worktree add ../s11-39-base 7efa02f`.

### X-12 · minor · Checks that cannot run here
**Answer:** Fixed — §8 says the Xcode part of `check-pr.sh` runs on macOS only, and T-9 is an existing test that stays green.

### X-13 · minor · Smaller gaps
**Answer:** Fixed — STEP-9 gives the exact paragraph, glossary rows and §12 bullet; §9 names the map row change; `TestLayerBridgesCoverEveryConnectRule` joins "Most likely to break".

## Round 2

Maintainability (Opus): M-1, M-3, M-4, M-5, M-6 accepted; M-2 accepted in part (led to M-7); X-4, X-5, X-8, X-9 accepted. Verdict: Fits with changes.

### M-7 · blocker · `withImplied` is a second same-ends, same-key merge beside the one `loadLayerBridges` keeps, against the working rule
**Plan:** D-5 `withImplied` indexes and merges on its own; the inline merge in `loadLayerBridges` and `edgeSeen` in `loadCanon` stay; D-6 / D-7 copy out and back.
**Codebase:** `core/database/promotealign/layer.go` `loadLayerBridges` (`seen`, inline merge); `core/database/promotealign/canon.go` `loadCanon` (`edgeSeen`); working rule "One fact, one link per side" ("through the loaders' existing merge (`edgeSeenKey`), not a second one").
**Better shape:** one `linkSet` with `add` in `terms.go`, used by both loaders' walks and by `withImplied`; one conversion per loader.
**Answer:** Fixed — D-5 is `sideLink` / `linkSet` / `add` / `withImplied`; D-6 and D-7 replace `seen` and `edgeSeen` with `set.add` and convert once; §6 Deletes lists both inline merges; T-4 is `TestLinkSet` through `LinkSetForTest`; STEP-8's skill text names `linkSet.add` as the one merge.

Executability (Sonnet): X-1 to X-13 accepted; every Current state row true at `7efa02f`. Verdict: Ready with changes.

### X-14 · major · AC-2.2 cites the wrong number for the witness control
**Answer:** Fixed — AC-2.2 says "above AC-1.1's witness control (2.5)", with 2.875 as AC-2.2's own measurement.

### X-15 · minor · T-3 reuses T-2's fixture but no builder is named
**Answer:** Fixed — STEP-2 names `marriageFixture(layerShape, canonShape string, dated bool)` and gives T-3's row literally.
