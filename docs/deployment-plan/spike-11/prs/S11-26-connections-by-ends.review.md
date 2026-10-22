# Review — S11-26 Connections file by their ends

**Verdict:** Fits — Ready
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- The copy changes no source dictated: the hint now reads "A connection files once both its ends are filed; to dispute one, change it on the Evidence graph" (D-7). The leave guard drops its connections count (D-8). The `.files` caption stays "Filed" until S11-27 rewords it (D-5).
- Keys keep their names when only their value or placeholders change (D-6, D-8): the catalog is English only, with `extractionState: manual`.
- `filedBridges` / `connectionLines()` are renamed to `connections` in this PR, not in S11-27 (D-3, from M-4).
- No design brief: removing a switch is not new chrome, and S11-27 replaces the panel under S11-D1.

## Round 1

### X-1 · blocker · STEP-2 check and AC-2.1 can never pass: generated code keeps the reserved name
Reserved names stay in the generated descriptors (`date_summary` is in `engine.pb.go` and `engine.pb.swift` today), so the check fails after regeneration.
**Answer:** Fixed. The STEP-2 check, AC-2.1 and manual check step 5 now exclude `*.pb.go` and `*.pb.swift`. A second STEP-2 check confirms the generated accessors (`SkipBridgeIds` / `skipBridgeIds`) are gone. The stop condition allows the descriptor bytes to change.

### X-2 · blocker · STEP-4 breaks the build and its own check; it needs STEP-5
`PromoteView` and `L10n.connectionState` read `bridgeOff`, `.off` and `toggleBridge`.
**Answer:** Fixed. The old STEP-4 and STEP-5 are merged into STEP-4 ("Connection states come from their ends; the panel shows them"), which says why. The steps are renumbered, docs are now STEP-5, and the commit grouping follows.

### X-3 · blocker · D-11 and T-3 are not enough to write the Swift test
**Answer:** Fixed. D-11 now states the whole fixture: `mappedRest = true`, handle ids `e1` / `e2`, facts, the bridge, the `"skip"` / `"new"` tokens, the expected state after each step, and `batchRows()` count, order and pins (`a` stays a handle row). T-3 points to it.

### X-4 · major · T-1 setup: is `promote.Save` needed?
**Answer:** Fixed. D-10 spells out the fixture with no `promote.Save` call and adds `f.hasClaim(person.ID)`.

### X-5 · major · STEP-5 and D-4/D-6/D-8: unspecified details
**Answer:** Fixed.
- D-4 places the identifier after `.accessibilityElement(children: .combine)` and keeps `.multilineTextAlignment(.trailing)`.
- D-7 states that the header layout is not changed.
- STEP-5 names where `completed.md` gets its entry and in what format (the S11-32 entry), and the map status word (`In progress`, per the README status table).

### X-6 · minor · Plan wording nits
**Answer:** Fixed. The AC-2.1 pipes are gone: it now points to the §10 command. D-9 names the one `.rows` site and leaves the `!= nil` / `== nil` sites. STEP-3 lists `PromoteModel.swift` for the Done call only.

### M-1 · major · STEP-6 leaves two statements of the switch rule in `promote-graph-alignment.md`
**Answer:** Fixed. STEP-5 now also edits the §5 four-case table row and the §9.1 "When:" bullet, with exact text. Its check is now `rg -i "switch"` over both docs. The Scope, Current state and DoD lines name §5 and §9.1.

### M-2 · minor · `connectionMeta` re-counts what the flow owns
**Answer:** Fixed. D-6 passes `connectionCount` and `connections.count - connectionCount`.

### M-3 · minor · `connectionFiled` stays a standalone static used only by `connectionState`
**Answer:** Fixed. D-5 inlines it into `case .files:`, with the same key and value and the new comment. The STEP-4 check greps for `connectionFiled`.

### M-4 · minor · `filedBridges` / `connectionLines()` are misnamed and one forwards to the other
**Answer:** Fixed. D-3 merges them into `var connections: [Bridge]`, which S11-27's `connections(hostedBy:)` filters. Codebase fit has the row, and Deletes and Next adds are updated.

## Round 2

Maintainability: **Fits**, with every round 1 answer verified. Executability: **Ready with changes**. After the fixes below, no blocker is open.

### X-A · blocker · STEP-2's check cannot pass until STEP-3, and the Swift tree does not build between them
**Answer:** Fixed. The `macos/App` part of the STEP-2 check moved to STEP-3. STEP-2 now says that the Swift target does not compile until STEP-3, so it runs no Xcode check.

### X-B · major · Current state row is false about `PromoteModelTests`
**Answer:** Fixed. The Tests row names one `.rows` read (line 171) and two nil checks (lines 169, 242).

### X-C · major · Several checks cannot run without macOS tools
**Answer:** Fixed. A new Stop condition says: with `protoc`, the plug-ins or Xcode missing, stop after STEP-1 and report, and never hand-edit generated files.

### X-D · minor · STEP-4's `.off` check misses the enum declaration
**Answer:** Fixed. The STEP-4 grep adds `case off`.

### X-E · minor · STEP-5 leaves small choices open
**Answer:** Fixed. `completed.md` is written after the draft PR is open. The plan's Status becomes `In progress` with the link, and `ship-pr` sets Landed. The **Review:** line and the map row's Scope text are left as they are.

### X-F · nit · The accessibility identifier moves to the row
Logged. Nothing references it.

### M-5 · nit · The page predicts Go's filing rule in Swift
Logged for S12-09: when that PR changes `fileHeads`, its plan checks `PromoteFlow.connections` still agrees.

### M-6 · nit · "bridge" and "connection" both name one concept in `PromoteFlow`
Logged. The name matches S11-27's `connections(hostedBy:)`.
