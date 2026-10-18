# Ship log — S10-64 Check a thumbnail before decoding it

**Plan:** [`S10-64-thumbnail-decode-guard.md`](S10-64-thumbnail-decode-guard.md) · **Branch:** `s10-64-thumbnail-decode-guard` · **PR:** #469
**State:** fixing 2

Base: `main` @ `4f15ba70` (plan written against `e97cf54`). Worktree: `.claude/worktrees/s10-64-thumbnail-decode-guard`.

## Round 0 — implement
Head `1bfa7e9c`. Gate PASS (orchestrator run: 7/7 checks, 1187 Swift tests). Report deviations:
- STEP-1–4 landed as one commit (`f87de0a9`), not two: all edit `ThumbnailCache.swift` and its tests together.
- D-10: `main` already carried the S10-61/62 header text in `swift_vocabulary_test.go`; only added `ThumbnailDecodePolicy.maxDeclaredEdge` to the `macos-client.mdc` list.
- T-1 helper types and test method are `fileprivate` (Swift rejected `private`).
- Not run: the STEP-2 mutation check (remove `accepts` → T-3/T-4 fail); STEP-3's named wider suites on their own (covered by the full Xcode run). Manual check §10 not run.

## Round 1 — review
Five lenses at `1bfa7e9c`. Every verdict Clean; no blocker or major. Orchestrator read `ThumbnailCache.swift:63-134` and `add-swift-test/SKILL.md:17` to confirm each finding below.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 (+UX-1) | PRF, UX | minor | FIFO gate with no cancellation: rows scrolled past decode before the visible rows | To user → Fix now → F-2 | user 2026-10-18: "Fix in this PR" (reopens D-7). verified: `withDecodeSlot` resumes `decodeWaiters.removeFirst()`; the inner `Task` is unstructured, so a view's `.task` cancel never reaches it. Fix contradicts D-7 (no waiter cancel) and the Out line (PRF-2). Unmeasured; two lenses rate it minor |
| SEC-1 | SEC | minor | The slot wait sits between the `objectURL` lstat check and the `CGImageSourceCreateWithURL` open, which widens the TOCTOU gap | To user → Fix now → F-1 | user 2026-10-18: "Fix in this PR" (reopens D-5 ordering). verified: line 65 checks and lines 68-70 open, with `withDecodeSlot` between them. The move contradicts D-5's wording ("after `objectURL` returns non-nil"). The threat model is a static crafted folder, and the gap predates this PR |
| MNT-1 | MNT | minor | `add-swift-test` skill still lists `ArtifactFileFixtures` as PNG/PDF only; does not name `writeImage`/`write` | Fix now → F-3 (rides along) | verified: SKILL.md:17. No fix plan touches the file this round |
| MNT-2 (+CON outside-lens) | MNT | nit | Detached closure body not re-indented inside `withDecodeSlot` | Fix now → F-1 (rides along) | verified: lines 69-80 |
| CON-1 | CON | nit | Random-bytes T-1 row passes whether ImageIO yields a source or not | Logged | the plan specifies this shape; the TIFF row proves the type gate |
| CON-2 | CON | nit | T-4 does not assert the oversized key is stored as missing | Logged | the plan's T-4 shape; T-3 proves the storing half on the same nil path |

**User decisions (2026-10-18).** Both To-user items: "Fix in this PR". This amends **D-5**: `objectURL` now runs inside the slot. It also amends **D-7**: a queued decode whose callers have all gone is dropped. Record both in the plan's Landed note as user-approved amendments; do not rewrite the D-5/D-7 rows.

### Fix plan 1

**F-1 — Check the object path inside the decode slot, right before the open** (from SEC-1; MNT-2 rides along)
Files: `macos/App/Platform/ThumbnailCache.swift`.
Failing test first: none. This is an ordering change with no observable behaviour in a unit test; T-2/T-3 and S10-62's symlink tests must stay green.
Do: In `image(projectDir:relPath:)`'s unstructured task, move `guard let url = ProjectFiles.objectURL(...) else { return nil }` inside the `withDecodeSlot` closure, so that the check runs, then `Task.detached` opens the file, with no suspension point between them except the detached hop. Re-indent the detached closure body one level (MNT-2). Keep the class doc comment accurate.
Check: existing `ThumbnailCacheTests` and `ProjectFiles` symlink tests pass; scripts/check-pr.sh PASS.

**F-2 — Drop a queued decode when every caller for its key has gone** (from PRF-1, UX-1)
Files: `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`; check `macos/App/Platform/CachedThumbnail.swift` and `EvidenceCardView.swift` handle a cancelled load returning nil (no change expected). Skill: `add-swift-test`.
Failing test first: `ThumbnailCacheTests/dropsAQueuedDecodeWhoseCallersHaveGone` *(new)*, `.timeLimit(.minutes(1))`. Set `maxConcurrentDecodes = 1` and write three 64×64 PNGs k0, k1, k2. Start three `Task { @MainActor in await cache.image(...) }` for k0, k1, k2 and cancel k1's task right away. Then await all three. Expect: k0 and k2 non-nil, k1 nil; `cache.entryCount == 2`, which proves k1 was **not** stored as `.missing`; a fresh `await cache.image(k1)` afterwards returns an image. Run it red at the current head first. If it cannot be made red deterministically, report that and stop on F-2.
Do:
- Count live callers per key (main-actor state, e.g. `private var inflightCallers: [Key: Int]`). Every caller that awaits the inflight task (the creator and every coalesced joiner) increments the count before awaiting. It decrements once, either when its await returns or when its task is cancelled.
- Make cancellation reach the cache: await through `withTaskCancellationHandler`, whose `onCancel` hops to the main actor to decrement. Make sure a caller decrements once only. It must return promptly when cancelled (nil), not wait for the decode.
- Inside the `withDecodeSlot` body, before `objectURL` (F-1), check the count for the key. If it is 0, return early with a distinct "abandoned" outcome, releasing the slot through the existing `defer`; there is still no separate acquire or release call (D-5's structural release).
- An abandoned or cancelled result must **not** be inserted as `.missing`; only a real refusal or failure is. Remove the `inflight` entry so a later request starts fresh.
- `clear()` keeps cancelling inflight tasks, and those must not insert either.
- Coalescing (`coalescesInflightLoads`, AC-2.2) is unchanged for callers that stay.
Check: the new test is green; `coalescesInflightLoads`, `runsAtMostTheCappedDecodes`, and every existing ThumbnailCache test pass; scripts/check-pr.sh PASS.

**F-3 — Bring the `add-swift-test` fixture line up to date** (from MNT-1)
Files: `.agents/skills/add-swift-test/SKILL.md` (line 17).
Do: Match the new `ArtifactFileFixtures` doc comment: an image (PNG, JPEG, TIFF) via `writeImage(fileType:)`, raw bytes via `write(_:)`, a PDF, and the catalog row.
Check: `scripts/check-plan-docs.py` clean; scripts/check-pr.sh PASS.

**Docs:** In the plan's Landed note, record the two user-approved amendments (D-5 ordering; D-7 queued decodes whose callers have gone are dropped, which pulls part of PRF-2 forward) and the new test. In `docs/application-stack.md`'s Read-path sentence, add that a queued decode is dropped when no caller still wants it, if the sentence describes the queue.

## Round 1 — fix
Head `69a750c9`. Gate PASS (orchestrator run; 1188 Swift tests). F-1, F-2, F-3 and the Docs line done. `dropsAQueuedDecodeWhoseCallersHaveGone` was run red at `230685fe` (k1 non-nil, `entryCount` 3), then green. Deviations:
- F-1 and F-2 landed as one commit (`594738ef`), since both rewrite the same task body.
- Not covered by a test: `clear()` and then a re-request of the same key before the old task finishes. The implementer left it as is.

## Round 2 — review
Scope `230685fe...69a750c9`. Lenses: SEC, PRF, UX, MNT (those with Fix-now findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | — | (round 1) | Resolved | `objectURL` now runs inside the slot, right before the detached open |
| PRF-1 (+UX-1) | PRF, UX | — | (round 1) | Resolved | abandoned waiters cost one slot handoff and no I/O; `dropsAQueuedDecodeWhoseCallersHaveGone` |
| MNT-1, MNT-2 | MNT | — | (round 1) | Resolved | SKILL.md:17; re-indented closure |
| MNT-3 (+PRF outside-lens) | MNT | blocker | Load state is split across `inflight`, `waiters` and `clearGeneration`, so a load cancelled by `clear()` resolves the next load's callers with nil | Fix now → F-4 | verified: `finish` (lines 122-135) gates only `inflight`/`insert` on generation; `waiters.removeValue(forKey:)` always runs. `clear()` has no caller in `macos/App` or `macos/ProvenenciaTests` (grep). This PR added `clearGeneration` for a function nothing calls |
| UX-R2-a | UX | minor (observation) | `CachedThumbnail.load()` writes fallback on a cancelled nil without checking `Task.isCancelled`; stale-write predates PR | Follow-up | below blocker in round 2; Fix plan 2 does not touch `CachedThumbnail.swift` |
| MNT-R2-a | MNT (CON) | nit | Plan Out row "Cancelling a decode when every waiter has gone" now partly done | Fix now → F-4 (rides along; F-4 edits the plan's Landed note anyway) | plan line 79 vs Landed note |

### Fix plan 2

**F-4 — Delete the unused `clear()` so one load's state cannot split** (from MNT-3; MNT-R2-a rides along)
Files: `macos/App/Platform/ThumbnailCache.swift`; `docs/deployment-plan/spike-10-followups/prs/S10-64-thumbnail-decode-guard.md` (Landed note, Out row).
Failing test first: none. The failure is reachable only through `clear()`, which this fix deletes. Confirm with `grep -rn "clear()" macos` that nothing calls `ThumbnailCache.clear()` before deleting it. If something does, stop and report.
Do: Delete `clear()`, `clearGeneration`, and its doc comment. Drop the `generation` parameter from `finish` and the capture in `image(projectDir:relPath:)`. `finish` always sets `inflight[key] = nil` and inserts any outcome other than `.abandoned`; fix its doc comment to match. Change nothing else in the cancellation design. In the plan: add a sentence to the Landed note that `clear()` (no caller) was removed in fix round 2. Amend the Out row "Cancelling a decode when every waiter has gone" so it reads as the part still left: cancelling a decode that is already running.
Check: all `ThumbnailCacheTests` pass, `dropsAQueuedDecodeWhoseCallersHaveGone` and `coalescesInflightLoads` included; scripts/check-pr.sh PASS.
