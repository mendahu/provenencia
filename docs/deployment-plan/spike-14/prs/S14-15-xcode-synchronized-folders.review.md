# Review — S14-15 Xcode synchronized folders

**Verdict:** Fits — Ready (after round 2 fixes)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- Xcode 16 becomes the minimum (D-4): format 77 is required for synchronized folders; CI's `macos-15` and the README's "last built with Xcode 26" already meet it.
- The conversion is a throwaway script, not Xcode's Convert to Folder (D-1).
- `macos/App/Resources` stays a classic group so the `Fonts` folder reference keeps its subfolder (D-2).
- No version bump: build-system only, as S14-24.
- The exception-list check lives in `check-guidance-names.py`, not a new script (D-7, from M-1).

## Round 1

### X-1 · blocker · STEP-2 check `MARKETING_VERSION = 0.2.5` prints 2, but the file has 4
**Answer:** Fixed — STEP-2 expects 4 (both targets, Debug and Release); added `isa = PBXGroup;` count 5 with the survivors named.

### X-2 · major · Expected SwiftFileList line counts may include generated files
**Answer:** Fixed — AC-2.1 and STEP-1 filter each list to lines under `macos/App/` or `macos/ProvenenciaTests/` before counting.

### X-3 · major · "Save `sort` of each `find`" is ambiguous
**Answer:** Fixed — STEP-1 gives the exact commands; more than one SwiftFileList match is a Stop condition.

### X-4 · major · DesignSystem README section heading not decided
**Answer:** Fixed — D-8 keeps `## Xcode project registration` and the fonts paragraph's cross-reference; `add-ui-component` step 4 cites that heading.

### X-5 · minor · Probe file contents half-specified
**Answer:** Fixed — AC-1.2 gives the test file, shaped like `PVFontTests.swift`.

### X-6 · minor · STEP-4 Check does not verify AC-1.1
**Answer:** Fixed — STEP-4 builds into `$S/dd-probe` and greps its `Provenencia.SwiftFileList` for `SyncProbe.swift`.

### X-7 · minor · AC-3.1's `rg` misses stale text
**Answer:** Fixed — pattern adds `Resources phase`, `manually-authored`, `Wire new files` (and, from M-2, the two "wiring" phrases). At the base it hits exactly the lines the PR rewrites.

### X-8 · minor · Script mechanics left open
**Answer:** Fixed — STEP-2 says groups are removed by walking `children` from the five roots (77 groups), the five survivors are named, and the count is checked.

### X-9 · minor · CI check cannot run on a draft PR
**Answer:** Fixed — DoD box and Stop condition apply after the PR is marked ready (the workflow skips drafts).

### X-10 · nit · Review line
**Answer:** Logged — the Review line now links this log.

### M-1 · major · The exception lists are a new hand-kept list with no off-Mac check
**Answer:** Fixed — D-7 rewritten: `check-guidance-names.py` `synced_folders_unlisted()` checks unlisted non-Swift files, stale exceptions, and a top-level `macos/App` folder that is not a root; new STEP-6, AC-3.3, T-7; precedent `kit_rows_missing`.

### M-2 · major · Guidance sweep misses wiring and README lines
**Answer:** Fixed — §4 **More guidance** row; D-8 gives replacement text for the `add-swift-test` description, `add-workspace-place` line 187, `add-ui-component` line 130 and `docs/design-system-layers.md` line 43; STEP-7 and STEP-8 list them; AC-3.1 widened. The heading question is answered under X-4.

### M-3 · minor · One synchronized root for all of `macos/App`
**Answer:** Declined — the reviewer marks folder paths in `membershipExceptions` unverified, and the shape also moves the `Fonts` folder reference to a new parent; a "try it, else D-2" step would leave the implementer a choice the plan must make. The trap it removes (a new top-level folder not built) is closed by D-7 (c). Out row added.

### M-4 · nit · New ID scheme and one-line layout
**Answer:** Logged — Xcode rewrites layout on first open (Risks row 3).

## Round 2

Fresh reviewers with the plan and this log. Both checked the round-1 answers against the repo and found them holding (X-1 count 4, 77 removed groups and 5 survivors, the three `.md` files, the guidance line numbers).

### M-5 · blocker · `SYNCED_ROOTS` is a second hand-kept copy of the project's folder list
**Answer:** Fixed — D-7 rewritten: the check derives roots, exceptions and the names under `macos/App` from `project.pbxproj` (no constant); adds (d) a root missing on disk; the docstring and the `check-pr.sh` comment are updated in STEP-6. T-4 counts corrected.

### X-11 · blocker · AC-3.1's `rg` uses `\|`, which matches nothing
**Answer:** Fixed — the sweep is a code block outside the table with plain `|`; STEP-7 first confirms 12 hits at the base (measured: 12).

### X-12 · major · Message text and parsing left open
**Answer:** Fixed — D-7 gives the four exact report lines, recursive `os.walk` skipping names starting with `.`, whole-object regex parsing, and the docstring change; AC-3.3 names three plants with their exact output.

### X-13 · minor · (c) covered directories only
**Answer:** Fixed — (c) covers any file or directory directly under `macos/App`.

### X-14 · minor · Two D-8 edits not exact
**Answer:** Fixed — `add-swift-test` lines 52–58 replaced entirely (heading and bundle-id line stay); the DesignSystem README body is given verbatim; `README.md` keeps its `1. ` marker.

### X-15 · minor · Empty phase format
**Answer:** Fixed — D-5 uses the file's multi-line empty `files` form, as phase `AA00000000000000000000B9`.

### X-16 · minor · STEP-6 status check
**Answer:** Fixed — `git status --porcelain` is empty except the step's two files.

### X-17 · nit · plan-docs warnings
**Answer:** Logged — 0 errors; the 7 warnings are bundle paths.

## Verdict

Fits / Ready after round 2's fixes. Two rounds done; no blocker open; nothing To user.
