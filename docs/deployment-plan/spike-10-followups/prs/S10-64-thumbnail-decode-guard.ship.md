# Ship log — S10-64 Check a thumbnail before decoding it

**Plan:** [`S10-64-thumbnail-decode-guard.md`](S10-64-thumbnail-decode-guard.md) · **Branch:** `s10-64-thumbnail-decode-guard` · **PR:** #469
**State:** fixing 1

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
