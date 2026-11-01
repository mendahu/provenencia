# Review — S13-31 Header data becomes summaries

**Verdict:** Fits with changes — Not ready → Ready with changes after round 1 fixes (pending D-4)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- **D-4 (M-6): fold the four `oneof header` blocks into one `ConclusionSummary` message?** Recommendation: yes. The user asked (2026-11-01) whether it still makes sense if the person, event and place detail pages diverge. The fold shares only the wrapper that says which kind a summary is; each kind's message stays its own, and page content lives in `ConclusionDetail`'s fields and sections. Open until answered.
- M-1 answered (2026-11-01): the band's line becomes `secondary` (D-10).
- M-2 answered (2026-11-01): `HeaderPlace` → `ChainedPlace` (D-1).

## Writer's own calls
- `PersonTitleDisplay` for `PersonHeaderDisplay` (M-3), matching its Event and Place siblings.
- `summaryChainDepth` for `headerChainDepth`.
- The folded field takes the lowest of each old oneof's numbers and reserves the other two.
- Planning base is #659's head, not `main`, since S13-02 has not merged.

## Round 1

### M-1 · blocker · "Summary" already names the band's line under the title
**Answer:** To user → answered: band line → `secondary`. D-10 *(new)*, STEP-5 *(new)*, §1.

### M-2 · major · `SummaryPlace` and `PlaceSummary` swap two words
**Answer:** To user → answered: `ChainedPlace`. D-1, D-3 (`headerPlace` → `chainedPlace`).

### M-3 · major · `PersonHeaderDisplay` only builds titles
**Answer:** Fixed — D-1 and D-6 name it `PersonTitleDisplay`; D-9 the test file.

### M-4 · major · `HeaderDependents` forwards to `Dependents`
**Answer:** Fixed — D-5 deletes it and lists its callers; STEP-2; §6 Deletes; D-9 `TestDependentsIncludeChildPlaces`.

### M-5 · major · The old-name sweep cannot see D-3 names
**Answer:** Fixed — §2 adds the word sweep over changed files with D-2's keep list (AC-1.2, T-5); D-3 names `addHeaderID`, `handleHeaders`, `headerChainDepth` and the other compounds; D-2 lists every kept band name.

### M-6 · major · The person/event/place oneof is written four times in proto, Go and Swift
**Answer:** To user — D-4 written for the fold, Open question 1. Contradicts the first "rename only" framing, so the user decides.

### M-7 · minor · `searchindex.handleHeaders` rebuilds the package's by-id result
**Answer:** Declined for this PR — it is a refactor of `RowsForSearch`'s return type, not a rename; Out row with the PR body as its follow-up home.

### M-8 · minor · `TitleSource` typealias with one outside use
**Answer:** Fixed — D-11.

### M-9 · nit · `conclusionsummaries.Summaries` stutters; line 41 comment
Logged. The comment is reworded (D-4); `Summaries` kept, as D-1's user word.

### X-1 · blocker · STEP-1 and STEP-2 leave `api/` unbuildable
**Answer:** Fixed — STEP-1 edits the `api/ffi/handlers` imports and type names; STEP-1 and STEP-2 check `./core/... ./api/...`.

### X-2 · blocker · STEP-3 / STEP-4 Swift boundary is a judgment call
**Answer:** Fixed — STEP-3 touches only lines that name a generated type or field (with the grep that finds them) and keeps `CatalogConclusionHeader`; STEP-4 renames Swift types; STEP-6 properties and locals. The dylib question is moot: the Xcode build phase runs `scripts/build-macos-core.sh` (`project.pbxproj:2204`), noted in §4.

### X-3 · major · Map lines named do not match the map
**Answer:** Fixed — D-8 lists the map lines (rules 91, 96, 97, 98, 111; S13-04 and S13-05 rows and entries) and what stays; S13-12 and S13-13 dropped; the map sweep (AC-1.5, T-6).

### X-4 · major · Lowercase names unchecked
**Answer:** Fixed with M-5.

### X-5 · major · STEP-4 and STEP-5 too large
**Answer:** Fixed — Swift is now STEP-3 (wire lines), STEP-4 (types), STEP-5 (band line), STEP-6 (names and files, one folder at a time, building between folders allowed).

### X-6 · minor · pbxproj has no `name =` line
**Answer:** Fixed — D-6: replace the name in `path =` and the `/* … */` comments; §4 Xcode row.

### X-7 · minor · `completed.md` path, PR number timing, N/A boxes
**Answer:** Fixed — STEP-7 and DoD name `docs/deployment-plan/spike-13/completed.md`, do the Landed edits after `gh pr create`, and mark N/A boxes.

### X-8 · minor · Skill and rule lines, proto comments
**Answer:** Fixed — D-8 names `add-audit-scope/SKILL.md:80` and `swift-tests.mdc:23`; D-7 names the proto comment lines.

Nits (logged): STEP-3's number check is a script, not a reading (X, "Checks I could not run").
