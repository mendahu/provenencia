# Review — S16-01 Cap the Person page reads and an event's places against crafted projects

**Verdict:** round 1: Fits with changes — Ready with changes
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- M-1 (answered 2026-11-09): where the 100-relatives cap lives, since `core/kinship` may hold no number. The user chose the registry knob `kinship.Rules.MaxKinPerRole`. The map rule "Crafted fan-out is capped, not engineered" now says a cap inside a matching layer package is a `Rules` knob.

## Writer's own calls
- No manual check on a crafted project: bar 2's sizes cannot be built by hand, so timing tests prove them (§10).
- Fixture sizes of 4,000 children and 4,000 places, not bar 2's 20,000: today's code fails the 1 s budget at 4,000, and that size builds in seconds (D-14).
- `NotCounted` reports only `MaxReadWork`; lines that `MaxPairWork` or `MaxEntityMembers` leave at 0 facts stay silent. That gap and an uncapped number of Evidence graphs per person become two idea rows.
- Wire names `not_counted` and `lines_dropped`, so the zero value is the ordinary read (D-8).
- Patch version bump.

## Round 1

### M-1 · blocker · `maxKinPerRole = 100` puts a number in `core/kinship`, which the rules and CI forbid
`core/matchregistry/guard_test.go` `layerPackages` marks `core/kinship` numbers true, and `.cursor/rules/match-rules.mdc` says numbers live only in `core/matchregistry`. The map's rule told the planner to use a package constant.
**Answer:** To user, then Fixed. The user chose the registry knob (2026-11-09). D-1 now adds `kinship.Rules.MaxKinPerRole`, set to 100 in `matchregistry`, and `KinOf` takes `rules`. STEP-9 adds the inventory row. `core/kinship/export_test.go` is dropped. **Rules this plan bends** records the map rule's change, and the map rule was amended with this plan.

### M-2 · major · `placesOfEvent` beside `placesOfEvents`, and the map names the other one
**Answer:** Fixed. The new function is renamed `eventLocations` (D-6), and `placesOfEvents` is a `neighborsOf` loop over it. The map's two working rules (lines 99 and 130), the S16-01 row and the S16-01 scope entry now name `eventLocations`. S16-19's scope says "keep S16-01's capped `placesOfEvents`", which still holds, because `placesOfEvents` reads the cap.

### M-3 · major · `Canon.Bulk` adds a mode with a manual "commit every id" duty
**Answer:** Fixed in part. `Bulk` now records every entity it observed, recomputes them once on exit and appends the reported rows to `pending`, so no caller lists ids. `next` is widened to 24 bits (D-15), which removes the 65,536-id stop condition. Declined: deferring every `observe`'s recompute for all `Canon` users. That changes when kept rows appear for the tests of six packages that use `NewCanon` (`rg -l "NewCanon\("`), which this PR does not re-check. `bulkCanon` stays as it is (logged).

### M-4 · minor · `capLines` belongs beside `recordlinks.Pair`
**Answer:** Fixed. It is now `recordlinks.CapLines` (D-9), tested in `core/recordlinks` (T-8). `maxRecordLinkLines` stays in `persondetail`.

### M-5 · nit · `personNode` lives in `claims.go`
**Answer:** Logged; D-4 keeps it there.

### X-1 · blocker · `personNode` in `Families` can add a query per unknown id
**Answer:** Fixed. D-4 keeps `Families`' skip as `g.Cached(id)` plus `n.Is(typeID)`. That is the predicate `personNode` applies, on a node already loaded, so it adds no query. T-15 adds a query-count case with two unknown ids.

### X-2 · major · merged and nil behaviour of `Families`
**Answer:** Fixed by X-1. The skip is unchanged, and D-4 states that `Node.Is` is false for a nil or merged node.

### X-3 · major · `linksBetween` and `cfg`
**Answer:** Fixed. D-10 gives `recordLinks` its full signature, and `linksBetween` sets `cfg.Value` on its copy and keeps the caller's other sections.

### X-4 · major · `capLines` semantics
**Answer:** Fixed. D-9 defines strong as `StrengthStrong`, `most` as at least 1, and says kept pairs keep their order.

### X-5 · major · `Canon.Members` insert shape
**Answer:** Fixed. D-15 names the tables, columns, id families and transaction, and says to `Commit(entity)` after.

### X-6 · major · protoc and Xcode are not on a Linux runner
**Answer:** Fixed. A stop condition says STEP-7 and STEP-8 run on macOS with the repo's toolchain, and generated files are never hand-edited.

### X-7 · major · timing tests on slow machines
**Answer:** Fixed. D-14 builds fixtures outside the timed call and times cold reads after `Drop` and `WarmOnOpen`. A stop condition says to report the figure and never raise the budget.

### X-8 · minor · edges at the cut
**Answer:** Fixed. D-3 covers non-canonical relatives inside the first 100, D-2 counts distinct parents, D-5 counts the person's own record, D-6 counts a place whose node does not load, and D-7 counts a read of no graph.

### X-9 · minor · paperwork
**Answer:** Fixed. T-14 and T-15 are back in order. Codebase fit names every test. T-2 names its edge builder (`TestSiblingsManyDuplicatesTimesChildren`'s shape). Status stays Draft until round 2.

## Round 2

### X-10 · blocker · the test `kinship.Rules` literals have no `MaxKinPerRole`
**Answer:** Fixed — D-1: both literals in `core/kinship/kinship_test.go` gain `MaxKinPerRole: 100`; a value below 1 keeps nothing, as every `Rules` bound; STEP-1 lists the file.

### X-11 · major · the DoD sweep hits the spike's own docs
**Answer:** Fixed — the sweep excludes `docs/deployment-plan/**`.

### X-12 · minor · `linksBetween`'s return shape
**Answer:** Fixed — D-10 gives `([]recordlinks.Pair, bool, error)`.

### X-13 · minor · the risk row contradicted D-4
**Answer:** Fixed — §11 now says `Families` keeps `Cached` + `Is` and `personNode` runs once per `PersonFamily`.
