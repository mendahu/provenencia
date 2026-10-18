# Ship log — S10-62 Refuse symlinked objects in the app

**Plan:** [`S10-62-object-symlinks.md`](S10-62-object-symlinks.md) · **Branch:** `s10-62-object-symlinks` · **PR:** —
**State:** stopped — host disk full (180 MiB free); gate cannot build

Base: `main` @ `9467e847`. Worktree: `/Users/jakerobins/Repos/provenencia-s10-62`.

## Round 0 — implement
Stopped: data volume full; `xcodebuild` fails writing DerivedData. STEP-1 to STEP-3 code and tests written, uncommitted, Swift not compiled. `TestSwiftVocabularyMatchesGo` passes (shown red first). STEP-4 docs, commits, push, draft PR not done. Deviations so far: `symlinkedObjectSetsMissingFile` control asserts `emptyReason == .idle` (non-optional) plus `displayImage != nil`. Implementer agent `a20d10be23c71dede`.
