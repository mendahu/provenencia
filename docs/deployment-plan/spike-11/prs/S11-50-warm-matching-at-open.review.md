# Review — S11-50 Matching warms while the project opens

**Verdict:** Fits — Ready
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None. (The open budget of 2.5 s, D-5, is the writer's call; the user can veto it.)

## Writer's own calls
- `budgetWarmOnOpen` is 2.5 s, the old cold figure (D-5).
- A handle commit does not set the warm flag, so it does not rebuild the index on every write (D-3); its upkeep stays in `docs/ideas/name-index-on-commit.md`.
- After a drop, `Do` runs the whole `WarmOnOpen` list, aliases and name format included (D-4).

## Round 1

### M-1 · blocker · `Warm` would be the third config build
**Answer:** Fixed — D-1: `loadConfig` shared by `loadInputs`, `Warm` and `export_test.go` `LayerBridges`.

### M-2 · major · Two lists of what to warm
**Answer:** Fixed — D-2: `WarmOnOpen` is the one list and ends with `MarkWarm`; D-4: `Do` runs `WarmOnOpen` when `WarmNeeded`.

### M-3 · minor · The match index is cleared inline in three places
**Answer:** Fixed — D-3: `dropMatchIndex`, called by `dropDerived`, `OnCommit`'s handle branch and `ResetMatchIndexForTest`.

### X-1 · blocker · No accessor for the index generation
**Answer:** Fixed — `MatchIndexGenForTest` in `promotealign/export_test.go` for T-2; `catalogsession` tests assert presence after a drop (AC-2.1).

### X-2 · blocker · T-4's commits unnamed
**Answer:** Fixed — AC-2.1 to AC-2.5 name the `OnCommit` effect sets and the error subtest's `Drop` plus `errors.New`.

### X-3 · major · What a "person name commit" carries
**Answer:** Fixed — AC-2.4 and T-1 use `effects.Set{Handles: …}` literals.

### X-4 · major · T-2's fixture
**Answer:** Fixed — T-2 uses `newWarmFixture` and names the kinds asserted.

### X-5 · major · The "warm at open" subtest shape
**Answer:** Fixed — D-6 gives the `fastest`/`since`/`must` shape.

### X-6 · minor · Stop condition and `ResetMatchIndexForTest`
**Answer:** Fixed — named in the stop condition.

### X-7 · minor · Paths and metadata
**Answer:** Fixed — STEP-5 full paths; Status and Review set after this log.

### X-8 · minor · Kinds present only in seeded catalogs
**Answer:** Fixed — AC-1.1 says it covers onboarding-made projects.

### X-9 · minor · Open runs `Do` several times
**Answer:** Fixed — AC-1.1 and T-3 assert after `onboarding.Complete` plus one more `Do`, including `WarmNeeded()` false.

Nits (logged): `X-10 · grep -c exits 1 when it prints 0`.

## Round 2

Executability (Sonnet): Ready, no new blocker.
- X-11 · major · "the index is empty" is not assertable from `catalogsession` — **Fixed:** AC-1.3 says "present".
- X-12 · minor · `OnCommit` takes `effects.Commit` — **Fixed:** AC-2.1, AC-2.4 and T-1 write `OnCommit(effects.Commit{Effects: …})`.
- X-13 · minor · T-1 needs a seeded index and a real graph — **Fixed:** T-1 runs on `openCatalog(t)`'s graph and calls `SetMatchIndex` first.

Maintainability (Opus): Fits, no new blocker. M-1, M-2 and M-3's answers checked against the code and hold.
- Non-blocking: after a failed `fn`, the next `Do` runs its own `fn` before re-warming, so a lazily built index is built twice — **Logged:** it only follows an error; checking `WarmNeeded` before `fn` too is left to a later change if it shows up.
