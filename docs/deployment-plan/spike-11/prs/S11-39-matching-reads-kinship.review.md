# Review — S11-39 Matching reads kinship shapes

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
