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
