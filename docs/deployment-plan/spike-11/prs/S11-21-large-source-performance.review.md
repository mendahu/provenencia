# Review — S11-21 Large-Source performance check

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. Budgets and remedies were answered before review (user, 2026-10-24: "Snappy" budgets; bound + memos).

## Writer's own calls
- The benchmarks keep their catalog through `benchShape`, except that city events past the first 20,000 become deaths (D-4). Their numbers are reported only.
- No name-format commit effect, although the scope entry names one: the only writer is `nameformats.Install` at onboarding, so the name format is warmed at open like the alias tables (D-9). The scope entry is rewritten in STEP-9.
- ~~Warm Evidence keeps `SoleGraph`'s 2 SELECTs until S12-09 deletes it.~~ Resolved: S12-09 landed (#535) and removed `SoleGraph` from Promote; AC-3.3 now expects 0 SELECTs and AC-3.4 expects 1.
- Not built, measured inside budget: in-place index update, key-value profiling, a bounded canon walk.
- Map: S11-21 depends on S11-19 as well as S11-16; S11-31 depends on S11-21 (it sizes its cap from this fixture).

## Round 1

### M-1 · blocker · The warm name format belongs to `nameformats`, not `valuerules`
`conclusionheaders` `persons` and `handlers.GetNameFormat` also read `nameformats.Load`; a warm slot read only inside `valuerules` makes two ways to read one fact.
**Answer:** Fixed — D-9 now adds `nameformats.Warm`, `Invalidate` and `LoadFor(g, q)`; all three readers move to `LoadFor`; T-9 tests `nameformats`. Checked: `graphcache` imports neither `nameformats` nor `core/database`, so there is no cycle (stop condition added).

### M-2 · major · Steps and comments still describe nominations as unbounded
`promote-graph-alignment.md` §4 steps 1 and 4, `place`'s doc comment, `Rules.TopK` / `TieCap` comments, the `matching-model.md` glossary row **candidate**.
**Answer:** Fixed — §6 Go core gives the three comments; STEP-9 gives the doc sentences; §4 Docs row lists them.

### M-3 · major · `qualifiedEvent` copies `datedOn`'s body
**Answer:** Fixed — `(*fixture).eventOn` takes `*datevalue.Value` (nil: undated); `datedOn` calls it (STEP-1, Codebase fit).

### M-4 · major · The candidates memo holds a whole index pointer
**Answer:** Fixed — D-8 uses a `matchIndex.gen` generation from `indexGen atomic.Uint64`, like `profilesPattern` checks profiles; `loadShortlist` returns it.

### M-5 · minor · The fixture copies the open-time warm list
**Answer:** Fixed — D-9a `catalogsession.WarmOnOpen`, called by `openResearcher` and the fixture.

### M-6 · minor · `NominatedForTest` repeats `Align`'s steps
**Answer:** Fixed — D-7a splits `Align` into `run` plus `proposal`; the seam calls `run`.

### M-7 · nit · Name-format session test in `alias_cache_test.go`
**Answer:** Fixed — `core/catalogsession/warm_test.go` *(new)*.

### X-1 · blocker · How `bulkCanon` writes a date is not decided
**Answer:** Fixed — D-4: one `auto_reconciler_values` row with `value_date = valuecodec.MarshalDate`, `date_lo` / `date_hi` / `sort_key` NULL (`graphcache` `loadValues` reads only `value_date`; the planning prototype did exactly this); `fixtureDate` maps BEF / AFT / none / full day; death term and toponym writes named.

### X-2 · major · Fixture inputs underspecified
**Answer:** Fixed — D-4a: the cumulative-weight sampler, `letters`, one PCG stream and its draw order; STEP-3: the one-place Source is its own Source, its persons' names, participations and `locate`; Source A's events are not located; the 100-Subject count.

### X-3 · major · `CountSelects` can deadlock; nothing checks it early
**Answer:** Fixed — D-10 closes the `*sql.Conn` after `Raw` and clears the authorizer in `t.Cleanup`; STEP-2 adds `TestCountSelects` (T-10). `sqlite3.SQLITE_OK` is exported by mattn v1.14.50 (`sqlite3.go` line 357).

### X-4 · major · STEP-1 and STEP-6 hide many changes
**Answer:** Fixed — STEP-1 split into shape (STEP-1), counter (STEP-2), fixture (STEP-3); subtest `fixture` asserts 100 rows and ≥ 40 handle rows. Nine steps.

### X-5 · minor · T-7 reuses a local table
**Answer:** Fixed — STEP-4 hoists the table; the move is In.

### X-6 · minor · Nil graph
**Answer:** Fixed — D-9: `NameFormat` / `SetNameFormat` nil-receiver safe; `LoadFor(nil, q)` reads `q`; T-8 and T-9 cover both.

### X-7 · minor · Empty candidates entry; T-5 shape
**Answer:** Fixed — D-8: an empty non-nil slice is an entry. T-5 now has one layer birth, named ids and refs, and each case's dates. T-6 names `f.filedBirth` and `f.cite`.

### X-8 · minor · Cold runs under parallel `go test`
**Answer:** Fixed — D-1: cold and after-commit take the fastest of 2; the 4-core parallel assumption is stated.

### X-9 · minor · STEP-7 paths; scope entry and map disagree
**Answer:** Fixed — STEP-9 uses full paths and rewrites the scope entry's Ships line plus a Not needed line. Map: S11-21 depends on S11-16 and S11-19; S11-31 depends on S11-19 and S11-21. §4 Docs and Benchmarks rows corrected.

### X-10 · nit
**Answer:** Logged.

## Refresh after S12-09 (2026-10-24)

`main` moved to `cd54db8` (S12-18 #532, S12-09 #535). Every Current state row was re-checked there. Changes to the plan: `Propose` and `EvidenceRequest` take a graph id (`f.graph(src)` in the fixture); `loadInputs` no longer calls `SoleGraph`, so AC-3.3 is 0 SELECTs and AC-3.4 is 1, D-10 compares with literals, and the S12-09 Out line and Next-adds line are gone. No Decision changed. Not re-reviewed: the change removes work and adds no new choice.

## Refresh 2026-10-24 — D-12 (alias index), after the ship-pr round-0 stop

The user chose to cache the built alias index on the graph (ship log, Round 0). Reviewers saw only the amendment: maintainability (opus), executability (sonnet).

### Writer's own calls
- The slot holds the bare `valuecompare.AliasIndex`, keyed with `valueWords`; `base.Words` no longer reaches the alias index (M-1). Safe because `AliasToken` reads no word rule at `Same`.
- The graph slot is renamed `AliasIndex` / `SetAliasIndex` (D-12a), beyond the comment fix M-3 asked for.

### Round 1

**M-1 · blocker · D-12 adds a second answer to "which word rules key alias tokens"** — `add.go` `valueWords` already keys stored tokens; `joinedKey` ignores `Folds` at `Same`, so the per-call `reflect.DeepEqual` guarded nothing.
**Answer:** Fixed — D-12 stores only the built index, keyed with `valueWords` on the warm and cold paths; no `words` field, no DeepEqual, no fallback. Verified `core/valuecompare/text.go:112` and `add.go:152`. The import-cycle stop condition is dropped (`add.go` already imports `matchregistry`).

**M-2 · major · T-11 (b) cannot fail** — **Answer:** Fixed — case dropped; T-11 is now (a) warm allocations and (b) nil graph equal to warm.

**M-3 · minor · `graphcache` comments say tables** — **Answer:** Fixed — D-12a renames the slot and rewrites its comments, the `dropDerived` / `MatchIndex` / `evidence_graph.go` comments, the `namealiases` package doc and `Warm`'s doc.

**M-4 · nit · pointer slot value** — **Answer:** Fixed — a value (`AliasIndex`).

**X-1 · blocker · "under 10 allocations" unreachable with DeepEqual** — measured: `reflect.DeepEqual` on the registry's word rules is 26 allocations, `Default()` is 46. **Answer:** Fixed — DeepEqual is gone (M-1); AC-3.6 and T-11 take `base` outside the closure.

**X-2 · blocker · STEP-9's check does not discriminate** — **Answer:** Fixed — STEP-9 is verified by T-11 (a), run red first; `evidence warm` times before and after go in the PR body; a stop condition covers a T-11 (a) that passes before the change.

**X-3 · major · STEP-10 doc bullets already on the branch** — **Answer:** Fixed — STEP-10 says check them at `512a8572`, do not edit again.

**X-4 · major · STEP-10 leaves map, status and completed.md wording open** — **Answer:** Fixed — STEP-10 gives the scope entry's Ships / Not this PR / Check as a set, the "Record S11-21 as landed in pull request N." commit, the plan Status and Landed note, the map row `Landed #N`, and the `completed.md` section and its content.

**X-5 · major · fixture "58", unverifiable profile figures, untested error path** — **Answer:** Fixed in part — STEP-3 notes the branch measures 62. Declined for the profile figures: they come from the implementer's `-cpuprofile` run in the ship log, and the allocation count was measured while refreshing (13,341). Declined for an error-path test: the message is unchanged in kind and `Rules` had no such test before.

Nits/minors logged: X-6 (copying `valueWords` maps — moot, no stored words), X-7 (which rune — moot, case dropped), X-8 (Invalidate rebuilds — D-12 says it replaces the index), X-9.

### Round 2
Maintainability (opus): **Fits**, no new blocker. Minor: `NewAliasIndex(tables, valueWords)` would be written twice. **Answer:** Fixed — STEP-9 adds private `buildIndex`, called by `Warm` and `withAliases`.

Executability (sonnet): **Ready with changes**, no new blocker.
- X-10 grep prints `resetAliasTables()` — **Fixed:** the grep is now `"\.AliasTables()\|SetAliasTables"`, expected empty (14 lines before).
- X-11 which `graphcache` comments — **Fixed:** D-12a names every comment in `graph.go` and `evidence_graph.go` that says tables or slice.
- X-12 where `Load` runs in T-11 (a) — **Fixed:** outside the closure.
- X-13 `TestDropLeavesAliasIndex` value — **Fixed:** keeps `[]string{"jas"}`.
- X-14 `Warm`'s copy — **Fixed:** dropped.
- X-15 map scope cell and re-run — **Fixed:** STEP-10 gives the cell text and re-runs `check-pr.sh` after the landed commit.
- X-16 nit — logged.

**Refresh verdict:** Fits — Ready. No open blocker; no To user item.

## Addition 2026-10-24 — card symbol and font caches (US-5, D-13, D-14, STEP-11)

The user asked for the two caches on this PR after a profile of evidence-graph panning (plan §4 "Card painting"). Reviewers: maintainability (opus), executability (sonnet).

### Round 1
**X-1 · blocker · `Painter` is not main-actor, so a `@MainActor` cache will not compile** — **Fixed:** D-13 makes `Painter` `@MainActor`; it is built only in `draw(_:)` and the static `measure`, both on the `NSView`'s main actor.
**M-1 · major · key the symbol cache by `PVSymbol`** — **Fixed:** `Painter.symbol(_ symbol: PVSymbol, …)` like `mark`; callers drop `.rawValue`; T-12 uses `.penLine`, `.trash`, `.imageUp`.
**M-2 · major · two cache mechanisms; D-13 rejects what D-14 adopts** — **Fixed:** D-14 uses a `Hashable` key and a lock-guarded dictionary, the shape of `TimestampFormat`; `NSCache` is gone from both.
**X-2 · major · `===` cannot prove a cache; `!==` pairs may share an instance** — **Fixed:** DEBUG miss counters; T-12 and T-13 assert misses, red first; `!==` cases dropped.
**X-3 · major · `nsMono` family check under-specified** — **Fixed:** `familyName == PVFontFamily.mono` in the hosted test (the host registers the bundled fonts); the stop condition now names T-13.
**X-4 · major · pbxproj wiring and IDs** — **Fixed:** STEP-11 names the four entries per file, the groups and phases, and free IDs `FA`–`FF`.
**M-3 · minor · `PVFont.mono` through `nsMono`** — **Fixed.**
**M-4 · minor · test file name** — **Fixed:** `PVFontTests.swift`.
**X-5 · minor · bare `mono(` callers** — **Fixed:** STEP-11 says every bare call; the build is the check.
**X-6 · minor · cache details** — **Fixed:** each raw weight is its own key; fallbacks are cached; §6 notes the mono fallback change.
**X-7 · minor · checks** — **Declined in part:** AC-5.3 and the visual check stay manual (§10 step 5), with the user driving the app as for the first profile.
**X-8 · nit · temptations** — `Painter.mark` gets an Out line; the README sentence is in D-14.

### Round 2
Executability (sonnet): **Ready with changes**, no new blocker; X-1's answer confirmed (every `Painter` use is main-actor) and the pbxproj IDs confirmed.
- X-9 §6 signature — **Fixed:** `image(_:pointSize:)`.
- X-10 grep misses multi-line calls — **Fixed:** greps for `PVSymbol.x.rawValue`; the build also catches a leftover.
- X-11 global counters flake under parallel tests — **Fixed:** per-key DEBUG `isCached` checks on sizes no card uses, replacing the counters.
- X-12 cached fallbacks stick — **Fixed:** stop condition added.
- X-13 nit — logged.

**Addition verdict:** Fits — Ready. No open blocker; no To user item.

## Addition 2026-10-24 — card layers (D-15, D-16, STEP-12)

A draw log showed every card repainting on every pan frame, drawn into the document's layer (plan §4 "Card layers"). The user asked for the fix on this PR. Reviewers: maintainability (opus), executability (sonnet).

### Round 1
**X-1 · blocker · how the scroll view reaches the item layer** / **M-1 · major · two classes compute the scale** — **Fixed:** D-16 puts every trigger in `GraphCanvasNSScrollView` behind `syncItemContentsScale()`; it reaches the items through `weak var items`, set in `makeNSView` from the document it creates. `GraphCanvasDocumentView` drops out.
**X-2 · blocker · the scale source; AC-5.4 vs T-14** — **Fixed:** the item layer applies the `contentsScale` it is given; AC-5.4 and T-14 agree; opening may draw a card twice (AC-5.6).
**X-3 · blocker · live-magnify observer lifecycle** / **M-3 · minor · "settled" means two things** — **Fixed:** no notifications; `magnify(with:)`'s phase sets `isPinching`; every other source goes through one per-run-loop coalesce.
**X-4 · wheel and restored-camera repaints** — **Fixed:** stated as accepted costs in D-16.
**X-5 / M-2 · empty `viewDidChangeBackingProperties` override** — **Fixed:** deleted; the reason moves to `contentsScale`'s doc comment; the Check greps for it.
**X-6 · checks** — **Fixed in part:** red first is compile-level (stated); AppKit resetting a layer's scale is covered by comparing every layer on each sync, plus a stop condition for a display move.
**X-7 · Current state** — **Fixed:** `syncBackingScale` is noted as likely a no-op today.
**X-8 / M-4 · README and stale plan text** — **Fixed:** the README edits are named; §4 "Card painting" points at "Card layers"; the Out row names `GraphCanvasItemLayer` as its home.
**M-5 · nit · property, not setter** — **Fixed:** `contentsScale` is a property.

### Round 2
Maintainability (opus): **Fits with changes**. **M-6 · blocker · a rebuilt card keeps the backing scale while zoomed past 1×** — **Fixed:** `GraphCanvasItemLayer.insert` sets the scale after `addSubview`; both build sites call it; stop condition and manual check cover a selection at 1.25×. Animated zooms outside `magnify(with:)` named as an accepted cost.
Executability (sonnet): **Ready with changes**, no new blocker.
- X-9 open draw count — **Fixed:** two per card on Retina, one on 1×.
- X-10 no scroll-view test — **Declined:** a hosted `NSScrollView` in a window is flaky for what AC-5.6 checks by hand; stated in D-16.
- X-11 nil window and stuck pinch flag — **Fixed:** sync returns with no window; `viewDidMoveToWindow` clears `isPinching`.
- X-12 tolerance — **Fixed:** `0.000_1`, as the magnification code.
- X-13 sub-steps and README anchors — **Fixed.**
- X-14 nit — logged; covered by the stop condition.

**Addition verdict:** Fits — Ready. Round cap reached; no open blocker; no To user item.

## Addition 2026-10-24 — cached item image (D-17, STEP-13)

Own layers (STEP-12) did not stop the pan repaint; a `wantsUpdateLayer` prototype did (plan §4 "Card layers after STEP-12"). Reviewers: maintainability (opus), executability (sonnet).

### Round 1
**M-1 / X-1 · blocker · removing `draw(_:)` breaks `EvidenceCardLayoutTests.bitmap(of:)`** — **Fixed:** the base keeps `draw(_:) { paint() }`; the layout tests run unchanged through it; one `renderImage(scale:)` serves `updateLayer` and T-16.
**X-2 · blocker · image orientation and pixel mapping** — **Fixed:** D-17 gives the transform and a per-axis scale from the rounded pixel size; T-16 checks a red top-left pixel.
**M-2 · major · contract only in docs** — **Fixed:** `makeView` returns `GraphCanvasItemView`; `insert` drops the duplicate `wantsLayer`; `SizedView` subclasses the base; AC-5.4 reworded.
**X-3 · major · immediate scale unspecified** — **Fixed:** `applyItemContentsScale()` at once on window and backing change, then the async pass catches an AppKit reset.
**X-4 · major · ghost scale** — **Fixed:** stated; the ghost renders at the backing scale, accepted softness above 1× up to 1.25×.
**X-5 · major · init chain** — **Fixed:** `override init(frame:)` and an unavailable `required init?(coder:)` with `fatalError`, as the card has.
**M-3 · minor · cache key on pixel size** — **Fixed.** **M-4 · minor · comments name `needsDisplay` vs `invalidatePaint()`** — **Fixed.** **M-5 · nit · name** — **Fixed:** `GraphCanvasItemView`. **X-6 · move `draw` body verbatim** — **Fixed.** **X-7, X-8** — **Fixed** (key comparison stated; README wording given). **M-6 · nit · third bitmap-context copy** — logged; rejected-alternatives column says why.

### Round 2
Executability (sonnet): **Ready with changes**. **X-1 · blocker · `cardView(for:)` returns `NSView`** — **Fixed:** it returns `EvidenceCardView`; `EvidenceCanvasCard.swift` in Files. **X-2 · blocker · nothing re-renders the ghost on window or backing change** — **Fixed:** the base's `viewDidMoveToWindow` / `viewDidChangeBackingProperties` set `needsDisplay`; subclasses call `super`. X-3 commit split, X-4 pending-flag order and pinch guard, X-5 T-16 pixel read, X-6 Check names T-14 — **Fixed.**
Maintainability (opus): **Fits with changes**. **M-7 · major · the cache across `updateLayer` has no measured reason and adds a second signal** — **Fixed:** dropped; `updateLayer` always renders, `needsDisplay` is the one signal; the prototype's under-200 count is the evidence, AC-5.8 rechecks it. M-8 (= X-1), M-9 README sentence replaced not added, M-10 nits (D-13 wording, `isOpaque` on the base, empty-bounds guard) — **Fixed.**

**Addition verdict:** Fits — Ready. Round cap reached; no open blocker. To user after STEP-13's numbers: the memory trade (PRF-1).
