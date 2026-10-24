# Ship log — S11-21 Large-Source performance check

**Plan:** [`S11-21-large-source-performance.md`](S11-21-large-source-performance.md) · **Branch:** `s11-21-large-source-performance` · **PR:** [#536](https://github.com/mendahu/provenencia/pull/536)
**State:** ready

## Round 0 — implement
Head `512a8572` (local, not pushed; 9 commits on `origin/main` @ `7a1398ce`). Gate not run.

**Stopped on** §11 "a `TestLargeSourceBudgets` subtest exceeds its D-1 budget after the step that should meet it". `evidence warm` (AC-3.1, ≤ 100 ms) measures 90–96 ms alone, 103–125 ms beside other subtests, and 107 ms under the parallel `go test ./core/... ./api/...` run (fail).

- Profile: 0.29 s of 0.32 s under `promotealign.Evidence` is `loadInputs` → `valuerules.Load` → `namealiases.Rules` → `withAliases` → `valuecompare.NewAliasIndex`, about 1.5 ms per call. After STEP-8 a warm Evidence issues 0 graph queries and 0 SELECTs, so the remaining cost is CPU.
- Cause: every `Evidence` call rebuilds the alias index from the warm alias tables. D-11 installs the product alias tables, so each rebuild is large. The plan's prototype figure ("Evidence 20 ms") was measured without them.
- No Step, Decision or In line caches the built index. Options: (a) keep the built `valuecompare.AliasIndex` on the graph beside the alias tables (`namealiases.Warm` / `Rules`), dropped with them; (b) raise the D-1 Evidence budget.

Everything else is done: STEP-1 – STEP-8 meet their ACs and budgets (cold 0.78–1.1 s / 2.5; warm 65–82 ms / 250; after commit 0.41–0.56 s / 1.5; one-place Source cold 1.7–2.2 s / 5; warm 1.0–1.5 s / 3). STEP-9 docs partly done.

Deviations: `benchShape()` is a function (fresh `rand` stream); `largeShape(r, names)` shares the Zipf stream; pair-memo key is `type pairKey struct{ subject, handle string }` (`[2]string` fails `TestLayerPackages`); `place` de-duplicates fresh nominations by handle id before ranking; `warmFixture` gained `artB` for T-6; `evidence warm` landed in the STEP-8 commit, not STEP-3.

**Resolved (user, 2026-10-24):** "Amend the S11-21 plan with plan-pr: add a decision to keep the built valuecompare.AliasIndex on the graph beside the alias tables, then resume ship-pr." The plan gains D-12 / D-12a (the alias slot holds the built index, keyed with `valueWords`; renamed `AliasIndex` / `SetAliasIndex`), AC-3.6, STEP-9 (new) and T-11; Docs is STEP-10. Review log "Refresh 2026-10-24 — D-12": Fits / Ready after 2 rounds. Round 0 resumes at STEP-9.

Resumed: head `fa3631ad`. Gate PASS (orchestrator, test cache cleared: go-test 53 s, xcode-test 19 s). Report: STEP-9 and STEP-10 done, no deviation from the Decisions; T-11 ran red first (13,340 allocations) then green. Times at head (M1): cold 796 ms, warm 66 ms, after commit 408 ms, evidence warm 7.8 ms (90.4 ms before STEP-9), one-place Source cold 1.68 s, warm 974 ms.

## Round 1 — review
Reviewers: CON (sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Two paths into layer 5; `compare`'s "one call to layer 5" comment false | Fix now → F-1 | verified: `align.go` `compareSegments` builds its own `Segment`s for `CompareNeighbor`; `st.compare` (only caller `rows.go:200`) builds the same for `Compare` |
| MNT-2 | MNT | major | `use-catalog-session` skill says open warms only alias tables | Fix now → F-2 | verified: `SKILL.md:62`; `session.go` calls `WarmOnOpen` |
| MNT-3 | MNT | major | Domain docs miss the candidates memo and open-time warms; §12 query count stale | Fix now → F-3 | verified: diff of `promote-graph-alignment.md` touches only §4 and §11; comparison-layers step 6 still says "if built" |
| MNT-4 | MNT | minor | Candidates memo inline in `loadCanon`; `memberValues` writes into a stored memo | Fix now → F-4 | verified: `canon.go:90-130`; `evidence.go:340-347` assigns into the map `EvidenceGraphMemo` returned, against its read-only contract. No shared generic helper: the two memos differ in validity (generation) and value type; extract + copy-on-write makes them one shape |
| PRF-1 | PRF | minor | Pair memo key allocates two strings per lookup | Fix now → F-5 | measured by reviewer: 82 MB flat on the key build over the run. Nested `map[string]map[string]` lookups do not allocate. Storing only kept nominations: Rejected — `place` can meet the same pair again in the same `Align`, and the memo is per-`Align` (D-6) |
| MNT-5 | MNT | minor | Second warm slot is a field-for-field copy; "a new warm is one line" is false | Follow-up | two copies are fine under Q5 and no map row names a third warm; one keyed `Warmed(key)` slot when a third arrives. The plan's Next-adds claim is corrected in the Landed note (F-6) |
| CON-1, CON-2 | CON | nit → minor | Landed note says "none" but the code departs from D-4 / STEP-3 | Fix now → F-6 | raised: a stale line in a file this PR adds; round-0 report listed six deviations |

Nits (logged): `CON-3 / MNT-7 · canon.go:33 · rewrap loadCanon doc` (rides F-4) · `MNT-6 · namealiases.go:220 · inline withAliases` (rides F-7) · `MNT-8 · bench_test.go · Zipf helpers live in bench_test.go, D-4a says perf_test.go` (Landed note) · `MNT-9 · session.go:67 · WarmOnOpen doc says alias tables` (rides F-7) · `MNT-10 · promote-graph-alignment.md:166 · "new alias tables"` (rides F-3) · `MNT-11 · warm tests split across two files`

### Fix plan 1

**F-1 — One call to layer 5 in graphalign** (from MNT-1)
Files: `core/graphalign/align.go`, `core/graphalign/rows.go`. Skill: `add-match-rule`.
Failing test first: none — behavior is unchanged; every `core/graphalign` golden and `TestLargeSourceBudgets` must stay green.
Do: `st.compare` gains a `neighbor func() subjectcompare.Result` parameter and calls `segmentcompare.CompareNeighbor` (it is the only place that builds the two `segmentcompare.Segment` literals). `compareSegments` calls `st.compare(…, func() subjectcompare.Result { return st.pair(ns, nh) })`; `rows.go` passes `func() subjectcompare.Result { return st.evaluate(sa.Kind, sa.Values, sb.Values) }`. Keep the "compare is the one call to layer 5" comment, now true. If `st.compare` no longer needs `st.metas`, drop what goes unused.
Check: `grep -n "segmentcompare\.Compare\(Neighbor\)\?(" core/graphalign/*.go` prints exactly one non-test line, in `compare`; `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/ ./core/database/promotealign/ ./core/matchregistry/` passes with no test edited.

**F-2 — The catalog-session skill names the open-time warms** (from MNT-2)
Files: `.agents/skills/use-catalog-session/SKILL.md`.
Do: rewrite the Lifecycle "Open" bullet: the open runs `catalogsession.WarmOnOpen`, the one list of open-time warms (the built alias index, `namealiases.Warm`; the default name format, `nameformats.Warm`) held on the catalog graph for the life of the project; readers go through `namealiases.Rules` and `nameformats.LoadFor`; a writer calls the owning package's `Invalidate`; a new open-time warm is added to `WarmOnOpen`, not called inline. Also check `.cursor/rules/catalog-session.mdc` and fix it if it says the same.
Check: `grep -rn "namealiases.Warm" .agents .cursor` shows no text that names it as the only open-time warm.

**F-3 — Domain docs describe the new caches** (from MNT-3, MNT-10)
Files: `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/comparison-layers.md`.
Do: §8 Caching: one bullet on the candidates memo (each Subject's ranked candidate ids, on the Source's Evidence graph memo, valid for one match index generation; dropped with the Evidence graph memo, and ignored when the index generation changes), one on the open-time warms (name format and built alias index, held on the graph until the project closes; `SetAliasIndex` / `SetNameFormat` drop derived caches). Line 166's "new alias tables" → "a new alias index (`SetAliasIndex`) or name format (`SetNameFormat`)". §12 Query count: "a warm proposal reads `MAX(revision)` only; a warm Evidence request reads nothing from the catalog (`TestLargeSourceBudgets` `warm queries`)". comparison-layers step 1 cache sentence: name the candidates memo. Step 6: add "Built in S11-21: each Subject's candidates are remembered across proposals; `memo matches a fresh run` and `TestCandidatesMemoFollowsCommits` are the invariant."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` 0 errors; links resolve.

**F-4 — The candidates memo has its own loader; the members memo stops writing into a stored map** (from MNT-4, MNT-7, CON-3)
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/evidence.go`.
Failing test first: none — behavior unchanged; `TestCandidatesMemoFollowsCommits`, `memo matches a fresh run`, `TestEvidenceWarmIssuesNoQuery` stay green.
Do: move the memo block of `loadCanon` into `loadCandidatesMemo` *(new, `canon.go`)*, shaped like `layer.go` `loadLayerMemo` (key, read with the generation check, fill the missing Subjects, store a new map); `loadCanon` calls it and keeps the `add` loop. In `evidence.go` `memberValues`, build a new map holding the stored entries and the new ones (`maps.Copy`, as the candidates memo does) instead of assigning into the map `EvidenceGraphMemo` returned. Rewrap the `loadCanon` doc comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/...` passes.

**F-5 — The pair memo looks up without allocating** (from PRF-1)
Files: `core/graphalign/align.go`.
Do: `state.pairs` becomes `map[string]map[string]subjectcompare.Result` keyed by subject id, then handle id, looked up as `st.pairs[string(s.ID)][string(h.ID)]` (the shape of `st.cands`); delete `pairKey`. Store every computed pair as today.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/ ./core/database/promotealign/ ./core/matchregistry/` passes, `TestLayerPackages` included; `TestLargeSourceBudgets -v` one-place times logged in the report.

**F-6 — The Landed note lists the deviations** (from CON-1, CON-2, MNT-5, MNT-8)
Files: `docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` (the **Landed** line only).
Do: replace "none — the code matches the plan…" with the deviations: `benchShape()` is a function (a fresh `rand` stream per benchmark); `largeShape(r, names)` takes the shared `*zipfNames`; `zipfNames` and `largeShape` live in `bench_test.go`, not `perf_test.go`; the one-place Source is the Source that promotes the city ("Births at one place"), not a second Source; the pair memo is a nested map, not `[2]string` (`TestLayerPackages` allows no numeric literal but 0 and 1); `place` de-duplicates fresh nominations by handle id before ranking; `warmFixture` gained `artB` for T-6; `evidence warm` landed in the STEP-8 commit; round 0 stopped on AC-3.1 and D-12 / D-12a / STEP-9 were added (ship log); "a new open-time warm is one line in `WarmOnOpen`" holds for the call, but a new warm also needs its own graph slot (MNT-5, follow-up); `withAliases` is inlined (F-7); `st.compare` takes the neighbor closure (F-1); `memberValues` is copy-on-write (F-4).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` 0 errors.

**F-7 — Two stale alias comments and a forwarding wrapper** (from MNT-6, MNT-9)
Files: `core/database/namealiases/namealiases.go`, `core/catalogsession/session.go`.
Do: inline `withAliases` into `Rules`' cold branch (`base.Names.Aliases = buildIndex(tables)`) and delete it. `WarmOnOpen`'s doc says it stores the built alias index and the default name format.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/namealiases/ ./core/catalogsession/` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `21d94329`. Gate PASS (orchestrator, test cache cleared). F-1 – F-7 done; F-1 and F-5 share one commit (`e03f3123`, both in `align.go`). One-place Source cold 1.46 s, warm 818 ms after F-5. `.cursor/rules` had no warm text to change (F-2).
Read by the orchestrator: PRF-1 **Resolved** (`state.pairs` is `map[string]map[string]subjectcompare.Result`, `pairKey` gone); CON-1 / CON-2 **Resolved** (Landed note lists the deviations).

## Round 2 — review
Reviewers: MNT (round 1 had majors). Not sent: CON (nits only), SEC+PRF (one minor, resolved above).
MNT round 2: **Clean**. MNT-1, MNT-2, MNT-3 Resolved; F-4 and F-7 confirmed. No new blocker.

Nits (logged): `MNT-r2 · comparison-layers.md step 1 · "new alias tables drops them" — same stale wording MNT-10 fixed in the promote doc`

## Result
Gate PASS (`21d94329`) · CI pending · open: none · follow-ups: MNT-5 (one keyed warm slot on the third open-time warm), MNT-r2 wording · to user: — · CON miss: none

## After ready — vocabulary (user, 2026-10-24)
User: "strike that word from the record… refer to the data using the terms we already have." "Hub" was never an approved term. It is gone from the code and the records: the test data is "a Source whose births are all located at one place with many filed events" (the one-place Source), and its place is the city. Renamed: `catalogShape.Cities` / `CityEvents`, `BenchmarkProposeColdCity` / `WarmCity`, subtests `births at one place cold` / `warm`, `TestPlaceWithManyEventsProposalStaysOnTheLayer`, the `align_test.go` place ids. Docs rewritten in this plan, its review and ship logs, the map, and the S11-15 and S11-16 plans and ship logs. Unrelated meanings left alone: the "Hubert" nickname, a film reel's hub (`L10n.swift`), "generational hubs" in `docs/narrative-layer-data-model.md`, a code "hub package" (S11-02 review). `promote-connection-subrows.md` also called an event "the hub that ties people and places"; reworded.

## STEP-11 — card symbol and font caches (user, 2026-10-24)
User: "I would like to try caching the symbols and fonts… plan that and implement it on this same S11-21 performance PR." Plan addition (US-5, D-13, D-14, STEP-11, T-12, T-13) reviewed: Fits / Ready, 2 rounds (review log, "Addition 2026-10-24"). Implemented at `eed784e7`; gate PASS (orchestrator). Red-first was compile-level only (the tests' `isCached` hooks did not exist). `abc57274` alone does not build (`nsMono` and the pbxproj entries arrive in `71b32326`); the head does.

Reviewers: CON (sonnet), MNT, UX, SEC+PRF — all Clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | minor | No after-profile; vector draw in `drawTinted` remains | Fix now → re-profile with the user (§10 step 5) | reviewer's sample split: caches remove ~900–1,100 of 3,309 `draw` samples; ~890 of glyph rasterizing remain, owned by the §3 Out bitmap row |
| UX-1 | UX | minor | SwiftUI mono fallback is now monospaced; Landed note silent | Fix now → F-8 | verified `PVTypography.swift:151-153` |
| MNT-1 | MNT | nit → minor | Lock comment gives the wrong reason | Fix now → F-8 | stale comment in a file this PR changes; `PVFont` is not actor-isolated |
| MNT-3 | MNT | nit → minor | Landed note lists a non-difference | Fix now → F-8 | D-13 already says nil for a `pv.` glyph |

Nits (logged): `CON-1 · PVFontTests · 13.75 never asked for` · `MNT-2 · …ForTest vs …ForTesting suffix` (rides F-8) · `MNT-4 · README/PVTypography say serif/monospace fallback; display and body fall back to sans` (rides F-8)

### Fix plan 3
**F-8 — Comments and the Landed note say what the code does** (from UX-1, MNT-1, MNT-2, MNT-3, MNT-4)
Files: `macos/App/DesignSystem/Tokens/PVTypography.swift`, `macos/App/DesignSystem/README.md`, `macos/App/Features/EvidenceGraph/EvidenceCardSymbols.swift`, `macos/ProvenenciaTests/EvidenceCardSymbolsTests.swift`, `macos/ProvenenciaTests/PVFontTests.swift`, `docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` (Landed line).
Do: the font-cache lock comment says `PVFont` is not actor-isolated, so any thread may call it. Rename `isCachedForTest` / `nsIsCachedForTest` to `isCachedForTesting` / `nsIsCachedForTesting`, as `ArtifactViewerModel.installImageRasterForTesting` is named. The README `PVFont` paragraph and the `PVTypography.swift` comment near line 66 say display and body fall back to the system font and mono to the system monospaced face. In the Landed line, drop "`Painter.symbol` returns early when the symbol is not an SF Symbol" and add "the SwiftUI `PVFont.mono` fallback is now the system monospaced face (only when IBM Plex Mono is missing)".
Check: scripts/check-pr.sh PASS.

## Round STEP-11 — fix and re-profile
Head `a25697b4`. Gate PASS (orchestrator). F-8 done. UX-1, MNT-1, MNT-3 **Resolved** (read by the orchestrator).
PRF-1 **Resolved** — re-profile with the user driving (same 110-Subject Source, 20 s pan, optimized Debug build): hitch total 12.5 s → 9.4 s, worst 427 → 400 ms, hitches over 100 ms 43 → 49, main thread busy 4.3 → 3.8 s, card draw share 81 % → 72 %. Symbol and font creation are gone; `drawTinted` (vector symbol drawn on every paint) is 48 % and is the next cost. User: "I wouldn't describe it as particularly improved from an anecdotal perspective." The §3 Out row (keep each card's painted bitmap so panning never re-runs `draw(_:)`) is the fix that removes the repaint itself.

## STEP-12 — card layers (user, 2026-10-24)
User: "plan and implement that on this same PR." Plan (D-15, D-16, STEP-12) reviewed Fits / Ready in 2 rounds. Implemented at `3acec99d`; gate PASS (orchestrator). Reviewers: CON (sonnet), MNT, UX, SEC+PRF.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | AC-5.6 measurement not recorded | Fix now → measured | Draw log at `3acec99d` (user driving): 11,887 draws in the pan, ~106 per card; own layers did **not** stop the repaint. Scroll and clip view layers: 6,041 in 16 s, no change. Cached-image prototype (`wantsUpdateLayer`): under 200 refreshes for open and pan; user: "noticeably much, much better". Recorded in plan §4 "Card layers after STEP-12" → D-17, STEP-13 |
| PRF-1 | PRF | major | Layer memory unmeasured; zoomed-out scale is a user call | Fix now (measure, AC-5.8) → To user after STEP-13's numbers | `phys_footprint` 146 MB open / 175 MB after pan with own layers; 252 MB with the cached-image prototype, 1× display |
| UX-1 | UX | minor | Placement ghost lost its scale sync | Fix now → D-17 | the ghost subclasses `GraphCanvasItemView`; its paint is keyed by the layer's scale |
| UX-2 | UX | minor | Cards at 1× for a frame at open on Retina | Fix now → D-17 | scale set at once on window and backing change |
| MNT-1 | MNT | minor | `makeView` doc lacks the redraw contract | Fix now → D-17 | doc and README name `GraphCanvasItemView` and `invalidatePaint()` |
| MNT-2 | MNT | nit → minor | Codebase fit row says `makeView` | Fix now → STEP-13 | plan text |
| CON-2 | CON | minor | A host outside the scroll view must set `contentsScale` | Logged | no such host; README says the scroll view sets it |

Nits (logged): `MNT-3 · 0.000_1 tolerance repeated ~7 times in GraphCanvas` · `UX-3 · isPinching cleared only on .ended/.cancelled/window move`

## STEP-13 — paint through updateLayer (user, 2026-10-24)
User: "noticeably much, much better" (prototype). Plan (D-17, STEP-13) reviewed Fits / Ready, 2 rounds. Implemented at `3089fe81`; gate PASS (orchestrator). Implementer deviation: `displayIgnoringOpacity(_:in:)` override so `cacheDisplay` goes through `draw(_:)`. Reviewers: CON (sonnet) Clean; MNT Clean; SEC+PRF one major; UX one blocker, one major.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | blocker | Shadows shrink by the render scale (half size on Retina) | Fix now → F-9 | reviewer's scratch program: a 4 pt `NSShadow` offset renders 2 pt at scale 2 through the raw `CGContext`; `cacheDisplay` into a point-sized rep keeps 4 pt at 1, 2, 2.5. Quartz shadows sit in the context's base space. Corrects D-17's Render recipe (orchestrator's plan error; told to the user) |
| UX-2 | UX | major | Fractional card heights stretch the image | Fix now → F-9 | reviewer: a 1 pt line in a 30.4 pt view renders as two partial rows at scale 1; `measure` returns fractional heights |
| UX-3 | UX | minor | Tests check `cacheDisplay`, the screen shows `renderImage` | Fix now → F-9 | one renderer for both after F-9 |
| PRF-1 | PRF | major | Memory of one image per card unmeasured on shipped code / Retina | Fix now → measure (AC-5.8) → To user | runs in parallel with F-9 (memory does not depend on it) |
| MNT-1 | MNT | minor | `draw(_:)` open to override | Fix now → F-9 | `final` on `draw(_:)` and `displayIgnoringOpacity(_:in:)` |
| MNT-2 | MNT | minor | Two docs still say an item is an `NSView` | Fix now → F-9 | `GraphCanvasItemLayer.swift:5`, `README.md:51` |

Nits (logged, ride F-9): `MNT-3 · view(for:) returns NSView?` · `MNT-4 · applyItemContentsScale's dead ?? 1` · `MNT-5 · test pixel read via NSBitmapImageRep.colorAt` · `CON-1 · no direct test of the displayIgnoringOpacity override` · `PRF note · override has no super; item views have no subviews (doc line)`

### Fix plan 4
**F-9 — One scale-aware renderer, whole-point items, enforced contract** (from UX-1, UX-2, UX-3, MNT-1, MNT-2, MNT-3, MNT-4, MNT-5)
Files: `macos/App/Features/GraphCanvas/GraphCanvasItemView.swift`, `macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift`, `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift`, `macos/App/Features/GraphCanvas/README.md`, `macos/ProvenenciaTests/GraphCanvasItemViewTests.swift`, `docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` (D-17 Render text and the Landed line).
Failing tests first (in `GraphCanvasItemViewTests`): (1) a subclass painting a 10 pt square at (10, 10) with an `NSShadow` (offset (0, −4) in flipped terms, i.e. 4 pt down, blur 0, black): `renderImage(scale: 1)` and `renderImage(scale: 2)` both show the shadow's top edge 4 pt below the square's (read through `NSBitmapImageRep(cgImage:)`'s `colorAt`, converting pixels to points by the scale); (2) a view whose frame height is set to 30.4 has `bounds.height == 31`, and a 1 pt line painted at y = 20 renders at scale 1 as one row with alpha above 250 and its neighbours below 5.
Do: `renderImage(scale:)` makes an `NSBitmapImageRep` (`bitmapDataPlanes: nil`, `pixelsWide` / `pixelsHigh` = `Int((bounds.width * scale).rounded(.up))` / height likewise, 8 bits, 4 samples, alpha, `.deviceRGB`… use `.calibratedRGB` only if `.deviceRGB` misses the sRGB colors; the rep's `size = bounds.size`), and inside `effectiveAppearance.performAsCurrentDrawingAppearance` calls `cacheDisplay(in: bounds, to: rep)`, then returns `rep.cgImage`; this goes through the existing `displayIgnoringOpacity` → `draw` → `paint`, so screen, tests and printing share one path and AppKit scales shadows. Override `setFrameSize(_:)` in the base to round both sides up to whole points (`ceil`) before calling `super`, so `bounds` is integral and the image maps 1:1 at integer scales. Mark `draw(_:)` and `displayIgnoringOpacity(_:in:)` `final`; the latter's doc says item views have no subviews. Update D-17's Render text in the plan to this recipe and note the change in the Landed line ("D-17's raw `CGContext` render shrank shadows by the scale; renders through `cacheDisplay` into a point-sized rep"). MNT-2: `GraphCanvasItemLayer.swift:5` and `README.md:51` (and its type-table row) say a `GraphCanvasItemView`. MNT-3: `view(for:)` returns `GraphCanvasItemView?`. MNT-4: `guard !isPinching, let window` then `window.backingScaleFactor`. MNT-5: the existing red-pixel test reads through `NSBitmapImageRep(cgImage:).colorAt`.
Check: both new tests fail first and pass after; T-14, T-16 and `EvidenceCardLayoutTests` pass; scripts/check-pr.sh PASS.

## STEP-13 — fix and measure
Head `9e1c40ae`. Gate PASS (orchestrator). F-9 done; both failing tests ran red first. Implementer deviation: `displayIgnoringOpacity(_:in:)` flips the identity context `cacheDisplay` hands it (the paint came out upside down otherwise). UX round 2: **Clean**; UX-1, UX-2, UX-3 Resolved. MNT-1, MNT-2 Resolved (read by the orchestrator in the round-2 diff).
AC-5.8, user driving the diagnostic build (`PVPAINT` log): open 110 paints; 10 s pan **2** paints (was 9,970); 231 MB on the 1× display; 596 MB on the built-in Retina display; 723 MB at 1.25× (peak 1.29 GB while zooming); zoom out and back in: 692 repaints (552 at 2×, 140 at 2.5×). User: "still performing really well".
PRF-1 → **To user**: layer memory on Retina.
Outside-lens notes for the user: card shadow offsets are positive and AppKit casts them upward (pre-existing on `main`); the image is untagged device RGB (P3 colour shift to check).

## CI red on `45d48966` (docs-only commit)
Go tests: `TestLargeSourceBudgets` — `perf_test.go:126: UNIQUE constraint failed: canonical_entities.ref`. Every earlier branch run passed. Cause: `bulkCanon` writes digit-only refs (`PER-%05d`, `PLC-%05d`, `PLC-T%04d`, `EVT-`/`PAR-`/`LOC-` + `%05d`) that are also valid `ref.Mint` tokens (Crockford alphabet, 5 chars), so a ref the fixture mints at random can collide; about 2 % a run by estimate. On `main` only the benchmarks used them; this PR runs the fixture in every CI run. Gate failure → one fix item.

**F-10 — Fixture refs can never equal a minted ref**
Files: `core/database/promotealign/bench_test.go`.
Do: one helper `benchRef(prefix string, n int) string` *(new)* returning `prefix + "-U" + fmt.Sprintf("%05d", n)` (`U` is not in `core/ref` `alphabet`); every ref `bulkCanon` writes goes through it (towns pass their own counters with a distinct prefix inside the number, e.g. `benchRef("PLC", 90000+k)`, so towns and cities stay distinct). Comment: why `U`.
Check: `CGO_ENABLED=1 go test -tags fts5 -count=20 -run TestLargeSourceBudgets/fixture ./core/database/promotealign/` passes; `git grep -nE 'Sprintf\("[A-Z]{3}-%0|"(EVT|PAR|LOC)-"\+' -- core/database/promotealign/bench_test.go` prints nothing; scripts/check-pr.sh PASS.

## STEP-14 — review and measure
Head `cdc09696`. Gate PASS (orchestrator). Reviewers: CON (sonnet) Clean; MNT Clean; UX+PRF Clean. One minor from CON-1 / MNT-1: `benchRef` sits between `letters`' doc comment and `func letters` → F-12.
Zoom log, user driving (built-in Retina, 110 Subjects): pinch out 23,652 `updateLayer` calls, **2** paints (2,756 at `9e1c40ae`); user: "That performed a lot better." Memory 585 MB at 1×, 338 MB zoomed out, 774 MB at 1.25×. Pinch in: 16,173 calls, **555** paints, five rounds of 110 at scales 2.52 – 2.71 — the pinch overshoots past `maxMagnification` and springs back outside `magnify(with:)`, so each spring-back step settled at a new scale → F-11.

### Fix plan 5
**F-11 — The item scale never passes the maximum zoom** (zoom log)
Files: `macos/App/Features/GraphCanvas/GraphCanvasCoordinates.swift`, `macos/ProvenenciaTests/GraphCanvasCoordinatesTests.swift`.
Failing test first: T-15 case `magnification: 1.355` → 2.5 (today 2.71).
Do: `itemContentsScale` is `backingScale * min(max(1, magnification), GraphCanvasCamera.maxMagnification)`; its doc says why (a pinch overshoots and springs back). The plan's D-16 already says so.
Check: T-15 passes; scripts/check-pr.sh PASS.

**F-12 — `letters` keeps its doc comment** (from CON-1, MNT-1; MNT-4 rides)
Files: `core/database/promotealign/bench_test.go`, `macos/App/Features/EvidenceGraph/EvidenceCardView.swift`.
Do: move `// letters is the lowercase base 26 of k…` to sit directly above `func letters`; `loadMissingLikenesses`' doc names `invalidatePaint()`.
Check: `go vet -tags fts5 ./core/database/promotealign/` clean; scripts/check-pr.sh PASS.

## Fix plan 5 — result
Head `c102c773`. Gate PASS (orchestrator). F-11, F-12 done; F-11's test ran red first (2.71 vs 2.5). Zoom check, user driving (Retina): pinch out and back in to 1.25×: **7** paints, all at 2× and none at 2.5×; user: "felt good", "Everything looks great". Memory 336 MB. So the scale never settled to 2.5 after the pinch-in. The earlier zoom log (diag at `3089fe81`) shows 4 `.began` and 16 `.changed` magnify events and no `.ended`; the run at `9e1c40ae` did settle (spring-back repaints at 2.52–2.71). The pinch end reaches `magnify(with:)` unreliably, and a missed end leaves `isPinching` set, which also blocks backing-change updates (moving the window between displays) → F-13.

### Fix plan 6
**F-13 — Settle the item scale on a timer, not on the pinch end**
Files: `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift`, `macos/App/Features/GraphCanvas/README.md` (zoom-cap clause), `docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` (Landed line).
Do: delete `isPinching`, `itemScaleSyncPending` and the `magnify(with:)` override. Add `static let itemScaleSettleDelay: TimeInterval = 0.15` and `private var itemScaleSettle: DispatchWorkItem?`. `scheduleItemContentsScale()` *(replaces `syncItemContentsScale()`)* cancels the pending item and schedules a new one on the main queue after the delay (inside `MainActor.assumeIsolated`), which calls `applyItemContentsScale()`. The `magnification` `didSet` calls it. `viewDidMoveToWindow` and `viewDidChangeBackingProperties` call `applyItemContentsScale()` at once and then `scheduleItemContentsScale()` (to catch an AppKit reset). `applyItemContentsScale()` guards only on `let window`. Doc comments say why a timer: a pinch, a wheel step and the spring-back all change `magnification`, and the end of a pinch is not a reliable event.
Check: `git grep -n "isPinching\|magnify(with" -- macos/App` prints nothing; scripts/check-pr.sh PASS; manual: a pinch in to 1.25× logs one paint per card at 2.5×, and moving the window between displays after a pinch repaints at the new scale.

## Fix plan 6 — result
Head `f19d32e8`. Gate PASS (orchestrator). Review (MNT+PRF, round 2): Clean. Zoom check, user driving (Retina): a pinch in to 1.25× painted 4 cards, all at 2× — the timer never started. Experiment (diagnostic only, not committed): during a live pinch the `magnification` setter is **never called** (0 calls); `NSScrollView.didEndLiveMagnifyNotification` fires once at the end (mag 1.250); with it restarting the settle timer, the pinch in painted **110** cards at 2.5×, and moving the window to the 1× display painted 110 at 1.25×. Memory 287 MB on the 1× display at 1.25×. → F-14.

### Fix plan 7
**F-14 — A live pinch's end starts the settle timer**
Files: `macos/App/Features/GraphCanvas/GraphCanvasScrollView.swift`, `macos/App/Features/GraphCanvas/README.md` (zoom-cap clause), `docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` (D-16, D-17, §6 and STEP-12/13 text that still names `isPinching`, `syncItemContentsScale` or `magnify(with:)`; the Landed line).
Do: `GraphCanvasNSScrollView` observes `NSScrollView.didEndLiveMagnifyNotification` with `object: self`, `queue: .main`, a `[weak self]` closure inside `MainActor.assumeIsolated` that calls `scheduleItemContentsScale()`; the token is `nonisolated(unsafe) private var liveMagnifyEndObserver: Any?`, registered once in `viewDidMoveToWindow` when nil (the class has no `init` override) and removed in `deinit`, as `GraphCanvasDocumentView.resignKeyObserver` is. `scheduleItemContentsScale()`'s doc says a live pinch does not call the `magnification` setter, so its end comes from the notification; wheel steps and the spring-back come through the setter. In the plan, rewrite D-16's settle text and every stale mention to the shipped design: settle timer fed by the `magnification` setter and the live-magnify end; window and backing changes apply at once.
Check: `git grep -n "isPinching\|syncItemContentsScale\|magnify(with" -- macos/App docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` prints only the D-16 history note; scripts/check-pr.sh PASS; manual: a pinch in paints each card once at 2.5×.

## Fix plan 7 — result
Head `27c1778f`. Gate PASS (orchestrator); CI green (Go tests, macOS tests, localizable). F-14's observer is the code the experiment validated (pinch in: 110 paints at 2.5×; display move: 110 at 1.25×), plus its `deinit` removal — read by the orchestrator, **Resolved**.

## Result (evidence graph additions)
Gate PASS (`27c1778f`) · CI green · open: none blocking · to user: answered — card image memory parked as `docs/ideas/evidence-graph-card-memory.md` (user, 2026-10-24) · follow-ups: shadow direction (pre-existing), untagged device RGB on P3 · CON miss: none

## Canvas health — STEP-15 – STEP-17 (user, 2026-10-24)
User: "Do groups A, B and C on this PR, and park group D as a family-tree idea file." Group D parked in `docs/ideas/graph-canvas-second-host.md`. Plan (D-19 – D-23, STEP-15 – STEP-17) reviewed in 2 rounds (review log, "Addition 2026-10-24 — canvas health"). Implemented `f2ac4573` – `ff10f66a` (14 commits); gate PASS (orchestrator). Deviations in the plan's Landed line. Preview now sits at `card.origin`, the created card's landing spot (small visible change, flagged to the user).

Reviewers: CON (sonnet) Clean; MNT Changes needed; UX+PRF Changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | README reuse contract says hit targets are document rects | Fix now → F-15 | `GraphCanvas/README.md:33` vs D-20 |
| MNT-2 | MNT | major | Grid ↔ document conversion written three ways | Fix now → F-15 | `GraphCanvasDocumentView.swift:61-62`, `GraphCanvasItemLayer.swift:240-245` inline beside `GraphCanvasCoordinates` |
| MNT-3 | MNT | major | "One trimmed-label rule" ships twice beside more trim copies | Fix now → F-16 | same-pass cleanup: this PR added the copies; byte-identical extensions in `EvidenceCardText.swift` |
| MNT-4 | MNT | major | Text-height cap written twice in the Painter component | Fix now → F-15 | `PVPainter.swift:121`, `PVTextMetrics.swift:23` already differ at 0 lines |
| MNT-5 | MNT | minor | `PVSymbolImages` doc describes the card | Fix now → F-15 | |
| MNT-6 | MNT | minor | `previewID` convention nothing reads; preview ignores fields | Fix now → F-15 | narrower `GraphCanvasPreview` |
| MNT-7 | MNT | minor | `intrinsicContentSize` unrounded, disagrees with the frame | Fix now → F-15 | one ceiled size |
| UX-1 | UX | major | Connect rubber band looks solid below ~0.5× | Fix now → F-17 | width scales by `1 / m` (STEP-17), dash `[5, 4]` does not; round caps close the gap |
| UX-2 | UX | minor | Preview paints over cards where a click does nothing | Fix now → F-17 | the old ghost sat under the items; D-22's "always on top" reversed: the preview goes under the items (an occupied cell stays visibly occupied) |
| PRF-1 | PRF | minor | Restack re-inserts every view above the first change | Fix now → F-17 | a rebuilt view goes straight to its slot; a raise or lower moves only those views |
| MNT-8 | MNT | minor | Bridge estimate vs measured empty-text rules | Follow-up → idea file | removed with the constants fallback in `graph-canvas-second-host.md` "Hit targets from measured layout" |

Nits (logged; ride F-15 / F-16): `MNT-9 place/ghost → place/preview` · `MNT-10 rotor example "Subjects"` · `MNT-11 pointsAtCards → pointsAtItems` · `MNT-12 / CON-2 double blank line in GraphCanvasItemView` · `MNT-13 documentDelta: → delta:` · `MNT-14 tile uses a non-continuous rounded rect` · `MNT-15 item doc` · `CON-1 design-system-layers.md exception list gains PVPainter.mark`

### Fix plan 8
**F-15 — One conversion, one size, one text cap, a narrow preview, docs that match** (from MNT-1, MNT-2, MNT-4, MNT-5, MNT-6, MNT-7; nits MNT-9 – MNT-15, CON-1, CON-2)
Files: `macos/App/Features/GraphCanvas/GraphCanvasCoordinates.swift`, `GraphCanvasDocumentView.swift`, `GraphCanvasItemLayer.swift`, `GraphCanvasItemView.swift`, `GraphCanvasPointerController.swift`, `README.md`; `macos/App/DesignSystem/Components/Painter/PVPainter.swift`, `PVTextMetrics.swift`, `Components/Icon/PVSymbolImages.swift`; `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `EvidenceGraphModel.swift`; `docs/design-system-layers.md`; tests `GraphCanvasCoordinatesTests.swift`, `GraphCanvasItemViewTests.swift`, `GraphCanvasItemLayerTests.swift`, `PVPainterTests.swift`.
Failing tests first: `GraphCanvasCoordinates.gridPoint(_:origin:)` round-trips with `documentPoint`; a view with `contentSize` 40 × 30.4 has `intrinsicContentSize == frame.size`; `PVTextMetrics.maxHeight(font:lines:)` at 0 and 2 lines.
Do:
- `GraphCanvasCoordinates.gridPoint(_ document: CGPoint, origin: CGPoint)` *(new)*; `GraphCanvasDocumentView.gridPoint(forDocumentPoint:)` and `GraphCanvasItemLayer.place` call `GraphCanvasCoordinates` instead of inline sums.
- `GraphCanvasItemView` derives both the init frame and `intrinsicContentSize` from one `private static func frameSize(for contentSize:)` (ceiled).
- `PVTextMetrics.maxHeight(font:lines:)` (lines < 1 treated as 1) used by `textHeight` and `PVPainter.text`.
- `GraphCanvasItemLayer.preview` takes a `GraphCanvasPreview(origin:model:makeView:)` *(new)*, still through `insert` / `place`; `previewID` and its README sentence go.
- README reuse row: "string ids + grid-space rects" linking "Coordinate seams".
- `PVSymbolImages` doc: for views that paint with AppKit, `PVPainter.symbol` draws the vector.
- Nits: "place/preview"; neutral rotor example; `pointsAtItems`; blank lines; `EvidenceGraphModel` `documentDelta:` → `delta:`; `tile` clips with `continuous`; a better item doc line; `docs/design-system-layers.md` lists `PVPainter.mark` with the recipe-artwork exceptions.
Check: the new tests fail first and pass; scripts/check-pr.sh PASS.

**F-16 — One trimmed-label rule** (from MNT-3)
Files: `macos/App/Platform/GenealogyStore.swift` (beside `CatalogSubject`), `macos/App/Features/EvidenceGraph/EvidenceCardText.swift`, `EvidenceCardView.swift`, `EvidenceSubjectCard.swift`, `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift`, `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`.
Do: `extension CatalogSubject { var trimmedLabel: String? }` (nil when blank). `titleOrTypeLabel` on both placed types becomes `subject.trimmedLabel ?? typeLabel`. Every other copy of the trim (`EvidenceSubjectCard.swift` ~316, `EvidenceGraphSnapshot.swift` ~53, `EvidenceBridgeEdgeSummary.swift` ~98, ~151, and any found by `git grep -n "label.trimmingCharacters(in: .whitespacesAndNewlines)" -- macos/App`) uses `trimmedLabel` with its own fallback. `EvidenceCardView`'s `typeLine` / `typeRef` share one helper.
Check: that grep prints only `trimmedLabel`'s definition (plus any trim of a non-Subject string, listed in the report); scripts/check-pr.sh PASS.

**F-17 — The band stays dashed, the preview sits under the items, restacks move only what changed** (from UX-1, UX-2, PRF-1)
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphEdgeLayer.swift`, `macos/App/Features/GraphCanvas/GraphCanvasItemLayer.swift`, `README.md` (preview line), the plan's D-22 text (\"always on top\" → \"below every item\"), `macos/ProvenenciaTests/GraphCanvasItemLayerTests.swift`.
Failing tests first: the preview is the first item-layer subview (below every item) across pushes; a selection-style rebuild of the middle of three items, and a raise of the first, each move only the affected view (count `addSubview` moves with a container subclass that records `didAddSubview`).
Do: the rubber band's dash is `[5 * pixel, 4 * pixel]`. The preview is kept directly below the lowest item view (`positioned: .below`), so cards cover it and its hit-less frame never sits over a card. `insert` puts a rebuilt view straight into its desired slot (`.below` the next desired view that is already placed, else `.above` the last), and `restack` moves only views whose position differs (raised ones to the top, lowered ones back into place).
Check: the new tests fail first and pass; T-18 – T-20 pass (T-20's "stays on top" assertion becomes "stays below every item"); scripts/check-pr.sh PASS.
