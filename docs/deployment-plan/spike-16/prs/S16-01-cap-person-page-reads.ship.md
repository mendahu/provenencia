# Ship log — S16-01 Cap the Person page reads and an event's places against crafted projects

**Plan:** [`S16-01-cap-person-page-reads.md`](S16-01-cap-person-page-reads.md) · **Branch:** `s16-01-cap-person-page-reads` · **PR:** #756
**State:** ready

## Round 0 — implement
Base `main` @ `eab44535`. Dispatched pr-implementer.
Head `12d8e981`. PR #756 (draft). Report deviations: (1) `Families` batch-loads requested nodes (`g.Nodes(want)`) so unknown ids add no query (AC-2.3 case failed at base); (2) `core/kinship/families_test.go` `linksFor` lifts `MaxPivotEnd` for its `Siblings` call so three over-bound `FamilyUnits` cases keep their expectations; (3) `maxRecordLinkLines` comment says "moved from the app's arc cap" rather than naming `maxArcs`; (4) T-3 cannot fail at STEP-2 (STEP-1's cap already bounds it), T-6 shown red by lifting the cap; (5) §10 manual check not run. Implementer saw one flaky `TestLargeSourceBudgets/warm` (414ms vs 375ms, during Xcode build); passed on rerun.
Gate (orchestrator) PASS on `12d8e981`: diff-check, gofmt, go-vet, go-test, localizable, guidance-names, plan-docs, xcode-test, swift-warnings.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none — UX for `macos/App/Features/Conclusions/**` and `macos/App/Platform/**`; SEC for `core/kinship`, `core/recordlinks`, `core/matchregistry`, `api/ffi/**`; PRF for `core/database/**` and `macos/App/Features/**` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | AC-2.8: no test that GoStore carries `notCounted` / `linesDropped` from the wire | Fix now → F-1 | verified: decode is inline in `getPersonRecordLinks` (`GoStore.swift:1272-1288`) behind `provenenciaCall`; T-14 checks only the default |
| MNT-1 | MNT | major | `linksFor` lifts `MaxPivotEnd` for every caller, hiding that D-2 leaves `parentUnits`' `!pairing` placement unreached in production | Fix now → F-2 | verified: `families_test.go:45-50` lifts the bound for all 16 call sites incl. `FuzzFamilyUnits`; only `Siblings` makes `PatternSharedParent` (`siblings.go:55`). Option (b): keep the branch as a defence, scope the lift to the over-bound tests |
| MNT-2 | MNT | major | 4th and 5th copies of the drop / warm / time cold-read loop | Fix now → F-3 (same-pass cleanup) | verified five copies: `persondetail/recordlinks_perf_test.go:83-87`, `promotealign/perf_test.go:182-186` and `:244-248`, new `persondetail/caps_test.go:20-29`, `conclusionsummaries/eventplaces_cap_test.go:149-159` |
| MNT-3 (+UX-1, UX-2) | MNT, UX | minor | Cap values (75,000; 100) written as numbers in comments outside Go | Fix now → F-4 | file this PR changes; supersedes the plan's literal wording in D-9/D-11/D-13 comments (no behavior change) |
| CON-2 | CON | minor | Implied siblinghood one-sided when only one child has > `MaxPivotEnd` parents | Fix now → F-5 (doc line) | verified by reviewer's probe; crafted-only, no cheap symmetric fix (`siblingsOf` cannot see the other child's parents). Document it in `Siblings`' comment and the `matching-model.md` `MaxPivotEnd` row |
| CON-3 | CON | minor | Three cap tests check less than their AC / Test row | Fix now → F-6 | tests this PR adds: T-7 never checks `places[20]` moved in; `RowsForSearch` result discarded; AC-1.1 compares against `KinOf`'s own output |
| CON→PRF note | PRF | — | `canonlinks.siblingsOf` loads every parent's edges before `Siblings` returns nil | Rejected | linear in input, and the plan's Untrusted size line names "`canonlinks`' batched loads of a hub's neighbors" as left linear (plan line 209) |
| UX note→CON | CON | — | §10 manual check not run | Rejected | the user runs §10 after ready (ship-pr step 8); reported to the user |

Nits (logged): `UX-1/UX-2` folded into F-4 · `MNT-4 · core/database/persondetail/family.go:20-23 · Families doc omits the per-role cap and D-3's no-replace` · `MNT-5 · conclusionsummaries/eventplaces_cap_test.go:90,108,201 · literal 20; export maxEventPlaces` · `MNT-6 · persondetail/recordlinks_test.go:270,276,282 · literal 75_000; export maxRecordLinkLines` · `MNT-7 · eventplaces_cap_test.go:182 · contains → slices.ContainsFunc` · `CON-4 · persondetail/claims_test.go:281 · "at the cap" row does not check the prefix` · `CON-5 · core/kinship/families.go:96 · cap check before the person-match test`

### Fix plan 1

**F-1 — Test that GoStore carries the two partial flags** (from CON-1)
Files: `macos/App/Platform/GoStore.swift`, `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`. Skill: `add-swift-test`.
Failing test first: `goStoreCarriesThePartialFlags` — build a `Provenencia_Engine_V1_GetPersonRecordLinksResponse` with `notCounted = true`, `linesDropped = true`; `GoStore.mapPersonRecordLinks(resp)` carries both true; `GoStore.mapPersonRecordLinks(.init())` carries both false. (Fails to compile today.)
Do: move the body after `provenenciaCall` in `getPersonRecordLinks` into `static func mapPersonRecordLinks(_ resp: Provenencia_Engine_V1_GetPersonRecordLinksResponse) -> CatalogPersonRecordLinks`, following `mapObservation`'s shape (`GoStore.swift:2331`); `getPersonRecordLinks` returns `Self.mapPersonRecordLinks(resp)`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — Scope the `MaxPivotEnd` lift to the over-bound `FamilyUnits` tests** (from MNT-1)
Files: `core/kinship/families_test.go`, `core/kinship/families.go` (comment only).
Failing test first: none (test-helper change); proof is that `TestFamiliesMoreParentsThanMaxPivotEnd`, `TestFamiliesGrowth*` and the "over the bound a sibling sharing several parents…" case of `TestFamilies` still pass and every other caller, `FuzzFamilyUnits` included, builds with the registry rules.
Do: `linksFor` builds with the `rules` it is given (no lift; its comment stays true). Give the three over-bound cases their own builder (e.g. `linksPastSiblingBound`, or a `rules` parameter they pass with `MaxPivotEnd` raised) whose comment says why: `Siblings` implies none past `MaxPivotEnd` in production, and these cases exercise `parentUnits`' defence for kin that shows more parents than `Siblings` counted. In `parentUnits`' doc comment (`families.go:171-174`), say that with the registry rules `Siblings` implies no sibling past the bound, so the "placed once, with the first parent it shares" path is a defence for a sibling link that carries shared-parent shapes from elsewhere; keep the branch.
Check: `grep -n "MaxPivotEnd = " core/kinship/*_test.go` prints only the new over-bound builder (and any pre-existing `TestSiblingsParentBound` setup); `go test ./core/kinship/...` PASS; scripts/check-pr.sh PASS.

**F-3 — One cold-read timing helper for all five copies** (from MNT-2; same-pass cleanup)
Files: `core/catalogsession/catalogsessiontest/catalogsessiontest.go`, `core/database/persondetail/caps_test.go`, `core/database/persondetail/recordlinks_perf_test.go`, `core/database/conclusionsummaries/eventplaces_cap_test.go`, `core/database/promotealign/perf_test.go`.
Failing test first: none (refactor); the five call sites keep their budgets and pass.
Do: add `func Cold(t testing.TB, c *database.Catalog, runs int, read func()) time.Duration` to `catalogsessiontest` (`t.Helper()`; `databasetest.Fastest(runs, …)` around `c.Graph().Drop()`, `catalogsession.WarmOnOpen(c)` failing with `t.Fatal`, `databasetest.Since(read)`). If that import makes a cycle, put it in `databasetest` beside `Fastest` instead and say so in the report. Move all five copies onto it; delete `(*kinWorld).cold` and the inline closure.
Check: `grep -rn -B2 -A2 "WarmOnOpen" core --include='*_test.go' | grep -n "Graph().Drop()"` prints no line inside a `Fastest(` closure — only `catalogsessiontest.go`'s helper body has Drop → WarmOnOpen → Since; `go test ./core/database/persondetail/ ./core/database/conclusionsummaries/ ./core/database/promotealign/` PASS; scripts/check-pr.sh PASS.

**F-4 — Name the owning constant instead of the cap's value in comments** (from MNT-3, UX-1, UX-2)
Files: `macos/App/Platform/GenealogyStore.swift:821`, `macos/App/Features/Conclusions/PersonRecordLinksContent.swift:32`, `macos/App/Features/Conclusions/PersonFamilyContent.swift:23`, `api/proto/engine.proto:1639` (regenerate `engine.pb.go` / `engine.pb.swift` if the comment is carried), `docs/matching-model.md:33` (prose only; the inventory row keeps its value).
Do: replace "75,000" with Go's `persondetail.maxRecordLinkLines` and "100 relatives per role" with `kinship.Rules.MaxKinPerRole`. This supersedes the literal numbers in the plan's prescribed comment wording (D-9, D-11, D-13); note it in the Landed note.
Check: `rg -n "75,000|75_000|100 relatives" macos api/proto docs/matching-model.md` prints only the `matching-model.md` inventory row value (if any); scripts/check-pr.sh PASS.

**F-5 — Say that the parent bound makes implied siblinghood one-sided** (from CON-2)
Files: `core/kinship/siblings.go` (comment above `Siblings`), `docs/matching-model.md` (`MaxPivotEnd` row or the kinship prose beside it).
Do: one sentence each: a person past `MaxPivotEnd` parents implies no sibling, but a child of one of those parents still implies that person as a sibling through the shared parent, since its read does not see the person's other parents. Crafted data only.
Check: scripts/check-pr.sh PASS (plan-docs, guidance-names).

**F-6 — Make the three cap tests check what their rows state** (from CON-3)
Files: `core/database/conclusionsummaries/eventplaces_cap_test.go`, `core/database/persondetail/caps_test.go`.
Failing test first: n/a (tightened assertions on tests this PR adds; they should pass at head — if one fails, report it as a defect, do not loosen it).
Do: T-7 (`TestEventPlacesCapInputs`): also assert `w.places[20]` is in `after.Places`. `TestCraftedEventPlaces`: keep the `RowsForSearch` result and assert the event's search row names at most 20 places (`maxEventPlaces`). `TestCraftedFamily`: compare the shown children against the first 100 kids in build order (`kids[:100]`), not against `kinship.KinOf`'s output.
Check: `go test ./core/database/conclusionsummaries/ ./core/database/persondetail/ -run 'EventPlaces|Crafted'` PASS; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `bb79af76`. Gate (orchestrator) PASS. F-1, F-2, F-3, F-4, F-5 done. F-6 partly: T-7 now checks `places[20]` moved in; `TestCraftedEventPlaces` asserts the search row names at most 20 places. The `TestCraftedFamily` part was not reproduced: the fix plan's premise was wrong, because link order is not `kids` build order (the shown kids were indexes 13, 142–150, 1428…). That sub-point of CON-3 is **Rejected**: AC-1.1's order inside `KinOf` is proven by `TestKinOfCap`, and `TestCraftedFamily` proves the count and the timing. Implementer flakes, each passing on rerun: `TestLargeSourceBudgets/warm_at_open` and `/cold` while Xcode built, and two timing-based `WorkspaceDestinationHostTests` cases.
Minors read by the orchestrator: MNT-3/UX-1/UX-2 (F-4) **Resolved**; CON-2 (F-5) **Resolved**; CON-3 (F-6) **Resolved** apart from the rejected sub-point.

## Round 2 — review
Reviewers: CON (CON-1), MNT (MNT-1, MNT-2). Scope `12d8e981...bb79af76`. Not re-sent: UX and SEC+PRF (no Fix-now blocker or major).
CON-1 **Resolved** (`goStoreCarriesThePartialFlags`); MNT-1 **Resolved**; MNT-2 **Resolved**. No new findings.

Nits (logged): `MNT-8 · core/kinship/families_test.go:98 · unitsWith has no doc comment`

## Finish
Merged `origin/main` @ `a6f02c7a` (docs-only S15-01 plan). Head `af14be76`. Gate PASS.

## Result
Gate PASS · CI <pending> · open: none · follow-ups: none · to user: —
