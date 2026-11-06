# Review — S13-37 Unpromote reports what it removes

**Verdict:** Fits with changes — Not ready (round 1). Every finding is Fixed, and no blocker is open. The plan is `Planned`.
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None. The PR itself was the user's answer of 2026-11-05 (S13-17 review M-2, M-3).

## Writer's own calls

- **D-4:** deleted handles are an effects-registry path (`onDelete(self())` on `canonical_entities`), so every handle delete reports with no per-write code. Subject delete gets it too.
- **D-6:** `getUnpromoteImpact` returns `CatalogUnpromoteImpact`, and the session ignores `deletedHandles` until S13-17.
- **T-5:** the handler test covers a Subject with no bridges. The bridge contents are tested at the promote level (T-2), because no FFI-level bridge fixture exists.

## Round 1 — maintainability

### M-1 · blocker · `WriteEffects` field 14 is taken
**Answer:** Fixed. D-5 uses field 16, and the §4 Wire row lists 13 – 15 (`evidence_graphs` = 14, `reads` = 15).

### M-2 · blocker · `ReportsDelete bool` is a second way to name ids
**Answer:** Fixed. D-4 adds an `onDelete` guard and the `Effect.DeletedHandles` path, and `validate` checks where the path ends. T-4 is rewritten to match.

### M-3 · major · The bridge query runs twice, plus one query per bridge
**Answer:** Fixed. `FiledBridge` carries `EntityID`, and `bridgeHeadsSQL` selects it. The preflight reads once and passes the list on (D-1, D-2). The Scale section is corrected.

### M-4 · major · Two FakeStore dicts keyed by the same Subject id
**Answer:** Fixed. `unpromoteImpactBySubject` is retyped to `CatalogUnpromoteImpact` (D-7), and T-10 and its file are dropped.

### M-5 · minor · `promote.Impact` is vague
**Answer:** Fixed. It is renamed `UnpromoteReport`.

### M-6 · minor · A missing doc and a stranded comment
**Answer:** Fixed. STEP-6 updates `docs/macos-client-patterns.md`, and STEP-1 moves `UnfileBridges`' doc comment back above it.

### M-7 · minor · The client learns of a delete two ways
**Answer:** Fixed. STEP-6 logs it in `docs/ideas/structural-debt.md`, and Next adds names the path for a later deleted-Sources set.

## Round 1 — executability

### X-1 · blocker · Field 14 collides
**Answer:** Fixed, the same fix as M-1.

### X-2 · blocker · The review link is broken
**Answer:** Fixed. This log now exists.

### X-3 · major · The no-claim guard has no home
**Answer:** Fixed. D-2 checks `AcceptedEntityForSubjectTx` after `loadPrimary`, as `Unpromote` does. T-2 covers a Subject with no claim but with filed bridges on its graph.

### X-4 · major · The build breaks between STEP-2 and STEP-4
**Answer:** Fixed. STEP-2 updates the handler's `.Deletes` read and checks `go build ./...`.

### X-5 · major · Fixtures are unspecified
**Answer:** Fixed. T-1 and T-2 move to `promote` on `newBridgeWorld` / `newHousehold`, with R created before Q. A bridge Subject's own id expects an empty result. T-5 covers no-bridge only, and T-6 runs `writes.Run` in `promote`, with a Stop condition if the import cycles.

### X-6 · minor · T-7's fixture
**Answer:** Fixed. It is a new case built with `promoteInDir`.

### X-7 · minor · T-4 and `None` entries
**Answer:** Fixed, by M-2's shape. A path is checked through `paths()`, and `None` entries have no paths.

### X-8 · minor · Swift fan-out
**Answer:** Fixed. T-9 names the one test that changes, which now scripts `CatalogUnpromoteImpact`.

### X-9 · minor · The comment and `ClaimID` copies
**Answer:** Fixed in D-1 (the non-nil copies; the comment move).

### X-10 · minor · The `rg` in Edges
**Answer:** No change. It finds the one emitter, so the Stop condition holds at the base.
