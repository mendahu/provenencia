# Review — S11-05 Migrate matching onto value compare

Plan: [`S11-05-migrate-matching.md`](S11-05-migrate-matching.md). Lint: `check-plan-docs.py` clean before round 1.

Before the review, a prototype of D-1 – D-11 ran the whole Go suite at `b6d1d48`; its failures are the plan's §8 "Expected changes" table, and D-16 records the one row that changes a proposal's target (Skip → New).

## Round 1

Maintainability: Fits with changes. Executability: Ready with changes.

### M-1 · blocker · Every caller builds the catalog's value rules from the same three calls
**Answer:** Fixed. D-9 adds `core/database/valuerules` `Load(g, q) (Rules, namePattern, error)`; Propose calls it once and passes the key to `loadExhibitObservations`, which drops its second `nameformats.Load`; D-13's skill line names it; T-10 tests it.

### M-2 · major · Two copies of the similarity → outcome rule once the floors go
**Answer:** Fixed. D-5: exported `match.OutcomeOf`, called by `evaluate` and `ApplyScale`.

### M-3 · major · Docs and comments the PR makes false that D-14 missed
**Answer:** Fixed. D-14 adds `matching-model.md` "Not knobs" and `promote-graph-alignment.md` L111; D-6 rewrites the `graphalign/registry.go` header comment.

### M-4 · minor · `rules()` helpers only forward
**Answer:** Fixed. D-11: tests write `matchregistry.Default().Value` inline; the Codebase fit row and the risk row are gone.

### M-5 · nit · `Feature.ValueType` repeats `PropertyMeta.ValueType`
Logged for S11-10, which ports the profile check onto `vocabkey`.

### X-1 · blocker · AC-4.1's `rg` contradicts the test helpers
**Answer:** Fixed. AC-4.1's `rg` excludes `_test.go`.

### X-2 · blocker · `rules()` placement does not fit the test packages
**Answer:** Fixed by M-4: no helpers.

### X-3 · major · STEP-2 intermediate expected values missing
**Answer:** Fixed. The discount removal joins the value-compare step (STEP-2); the comparers and the discount share `DiscountComparer`, so no intermediate commit builds cleanly without one of them. D-15 says so; every table row is STEP-2.

### X-4 · major · STEP-5 check contradicts D-14
**Answer:** Fixed. STEP-4's `rg` names only deleted identifiers, covers `matching-model.md`, and states which lines may remain (inventory source cells, Removed knobs).

### X-5 · major · Floor comments left open
**Answer:** Fixed. D-5 deletes or rewords every comment naming a floor in `core/graphalign`.

### X-6 · major · Open choices in STEP-3
**Answer:** Fixed. `layer.go` dropped from the file list (no comparer there); D-5 states `Equivalent` stays `partial` until S11-13. `Config` is built only by `DefaultConfig` — no change needed.

### X-7 · minor · Codebase-fit search claim false
**Answer:** Fixed. The search line lists the real commands and hits.

### X-8 · minor · Map S11-05 "Ships" names `Loaded.Comparer`
**Answer:** Fixed. The map's scope entry names `valuerules.Load`.

### X-9 · minor · Checks that cannot run here; rule modifiers for T-1
**Answer:** Fixed for T-1 (each case copies `Default().Value` and sets the field). `check-pr.sh` runs on the user's Mac (STEP-5), as in S11-04.

## Round 2

Maintainability: Fits with changes.

### M-6 · blocker · STEP-3's check searched for `namealiases.Rules`, which D-13's skill line does not name
**Answer:** Fixed. STEP-3's check is `rg -n 'valuerules.Load' .agents/skills/add-match-rule/SKILL.md`.

Logged, not raised: `matching.ForEntity` loads the name format before `rank` loads it again through `valuerules.Load` (deleted with the package in S11-13); `core/autoreconcile/names.go` reads registry words without catalog aliases (S11-06).
