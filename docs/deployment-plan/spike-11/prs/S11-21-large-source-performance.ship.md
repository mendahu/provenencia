# Ship log — S11-21 Large-Source performance check

**Plan:** [`S11-21-large-source-performance.md`](S11-21-large-source-performance.md) · **Branch:** `s11-21-large-source-performance` · **PR:** —
**State:** implementing (round 0 resumed after the D-12 plan refresh)

## Round 0 — implement
Head `512a8572` (local, not pushed; 9 commits on `origin/main` @ `7a1398ce`). Gate not run.

**Stopped on** §11 "a `TestLargeSourceBudgets` subtest exceeds its D-1 budget after the step that should meet it". `evidence warm` (AC-3.1, ≤ 100 ms) measures 90–96 ms alone, 103–125 ms beside other subtests, and 107 ms under the parallel `go test ./core/... ./api/...` run (fail).

- Profile: 0.29 s of 0.32 s under `promotealign.Evidence` is `loadInputs` → `valuerules.Load` → `namealiases.Rules` → `withAliases` → `valuecompare.NewAliasIndex`, about 1.5 ms per call. After STEP-8 a warm Evidence issues 0 graph queries and 0 SELECTs, so the remaining cost is CPU.
- Cause: every `Evidence` call rebuilds the alias index from the warm alias tables. D-11 installs the product alias tables, so each rebuild is large. The plan's prototype figure ("Evidence 20 ms") was measured without them.
- No Step, Decision or In line caches the built index. Options: (a) keep the built `valuecompare.AliasIndex` on the graph beside the alias tables (`namealiases.Warm` / `Rules`), dropped with them; (b) raise the D-1 Evidence budget.

Everything else is done: STEP-1 – STEP-8 meet their ACs and budgets (cold 0.78–1.1 s / 2.5; warm 65–82 ms / 250; after commit 0.41–0.56 s / 1.5; hub cold 1.7–2.2 s / 5; hub warm 1.0–1.5 s / 3). STEP-9 docs partly done.

Deviations: `benchShape()` is a function (fresh `rand` stream); `largeShape(r, names)` shares the Zipf stream; pair-memo key is `type pairKey struct{ subject, handle string }` (`[2]string` fails `TestLayerPackages`); `place` de-duplicates fresh nominations by handle id before ranking; `warmFixture` gained `artB` for T-6; `evidence warm` landed in the STEP-8 commit, not STEP-3.

**Resolved (user, 2026-10-24):** "Amend the S11-21 plan with plan-pr: add a decision to keep the built valuecompare.AliasIndex on the graph beside the alias tables, then resume ship-pr." The plan gains D-12 / D-12a (the alias slot holds the built index, keyed with `valueWords`; renamed `AliasIndex` / `SetAliasIndex`), AC-3.6, STEP-9 (new) and T-11; Docs is STEP-10. Review log "Refresh 2026-10-24 — D-12": Fits / Ready after 2 rounds. Round 0 resumes at STEP-9.
