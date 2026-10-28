# Review — S12-34 One descriptor per cache key

**Verdict:** Fits with changes — Ready → Fixed (round 2); no open blockers
**Rounds:** 2. **Reviewers:** maintainability (Claude Opus 4.5), executability (Composer 2.5)

## To user

None.

## Writer's own calls

- Declined moving `invalidations(for effects:)` onto descriptor rows in this PR: map Out / S13-01 and S12-25 own that fan-out; this PR lands `CatalogQueryFollows` on patch/refresh slots as the pattern those PRs copy.
- Chose exhaustive `Kind.descriptor` switch over a `[CatalogQueryDescriptor]` array (M-2).
- Evolved `CatalogLastUpdatedPatch` into array-valued `CatalogQueryPatch` with no thin-wrap (M-3 / M-11).
- Loader payload via `guard case` inside the same Kind arm, not new key accessors (M-4).
- Per-key `CatalogQueryRefresh` with its own `follows`, not per-kind refresh (M-10).
- No descriptor-level `follows` — each patch and refresh carries its own (M-9).

## Round 1

### M-1 · blocker · The descriptor has no slot for "which ids of a write this kind follows"
**Answer:** Fixed — D-4 adds `CatalogQueryFollows` for patch and refresh id selection. Declined moving `invalidations(for effects:)` in this PR: Out cites S13-01 / S12-25 / S12-26 and map "Not this PR. A behavior change to any key."

### M-2 · major · `Kind.invalidation` stays outside the descriptor
**Answer:** Fixed — D-1/D-2: exhaustive `Kind.descriptor` switch; each arm owns `invalidation` builder.

### M-3 · major · Opaque `patch` closure breaks T-2; hand-composed sourcesList patch
**Answer:** Fixed — D-5: `patches: [CatalogQueryPatch]` with factories; sourcesList declares `[.lastUpdated(…), .updatedSourceRow]`.

### M-4 · major · How a Kind-keyed loader reads key payload
**Answer:** Fixed — D-3: loader with `guard case` in the same Kind arm; preconditionFailure on mismatch.

### M-5 · major · Sweep incomplete
**Answer:** Fixed — D-9 / STEP-6 / DoD list the deleted names, spike-13 prose, Session comments.

### M-6 · major · Deletes misses typeMismatch and lastUpdatedPatch
**Answer:** Fixed — Deletes and DoD name `CatalogQueryRegistryError` and `lastUpdatedPatch(_:)`.

### M-7 · minor · sourceWorkspace patch clearing stale via private `invalidatedKeys`
**Answer:** Fixed — D-5: `updatedSourceShell` uses `setQueryValue`.

### M-8 · minor · US-2 overstates compile-time Value typing at call sites
**Answer:** Fixed — US-2 reworded to loader/warm path only; Out logs call-site typing as not planned.

### X-1 · major · STEP-1 lastUpdated vs D-2 patch fields
**Answer:** Fixed — STEP-1 may place early `patches: [.lastUpdated(…)]`; STEP-3 renames types in the same PR.

### X-2 · major · STEP-4 omits nav recount
**Answer:** Fixed — D-6 / STEP-4 keep nav-recount `Task` after refresh.

### X-3 · minor · Fail-closed in release
**Answer:** Fixed — D-2: exhaustive switch; `preconditionFailure` if somehow missing; no soft `??`.

### X-4 · minor · Patch accessor name for T-2
**Answer:** Fixed — T-2 reads `descriptor.patches` / each patch's `follows`; `lastUpdatedPatch(_:)` deleted.

### X-5 · major · Xcode checks not runnable on Linux agent
**Answer:** Declined as a plan change — DoD / STEP-5 note CI macOS evidence for Linux agents; macOS implementers run `xcodebuild` as written.

### X-6 · minor · D-3 storage names
**Answer:** Fixed — sketch names `warm: @MainActor (WorkspaceSession, CatalogQueryKey) -> Void`.

### X-7 · nit · AC-1.2 planting
**Answer:** Fixed — test-local filtered set only; STEP-5 / AC-1.2 say so.

## Round 2

### M-9 · blocker · `follows` declared at two levels; descriptions disagree
**Answer:** Fixed — Dropped descriptor-level `follows`. Each `CatalogQueryPatch` and `CatalogQueryRefresh` carries its own; session computes ids from that slot (D-4, D-5, D-6, sketch).

### M-10 · blocker · Per-kind `refresh` cannot serve S12-25 citation reads
**Answer:** Fixed — D-6: `CatalogQueryRefresh` is per-key (`run(session, key, ids)`); session walks warmed keys after invalidations, same as patches.

### M-11 · blocker · Evolve-or-thin-wrap left open; `CatalogLastUpdatedPatch` not in sweep
**Answer:** Fixed — D-5 decides **evolve** (rename); Deletes / DoD / STEP-6 `rg` name `CatalogLastUpdatedPatch`, `.rows(`, and the skill / `macos-client-patterns.md` lines.

Executability round 2: no new blockers — Ready.
