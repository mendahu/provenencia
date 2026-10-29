# Review — S14-24 Give every workspace page its width class

**Verdict:** Fits with changes — Ready with changes, after 2 rounds. Two To user items open; the plan stays Draft until they are answered
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet, the implementer's model)

## To user

- **Round 2 M-5 / M-6: what the pane is for a page with a side pane.** The working rule said "minus any inspector". Round 1 re-measured Properties beside its inspector; round 2 found Metadata and Source types have the same shape, and that a re-measured column also drops the gap and the gutters, so the code would drift from the rule. **Recommendation, written into the plan (D-6) and reworded on this branch in the working rule, § 9, the Paste block and S14-D1:** the class is the whole page's (window minus sidebar); a page with a side pane plans that pane in its own Wide frame. One measurement, and the CSS container query measures the same thing. If you prefer the old rule, D-6 goes back to re-measuring beside each side pane (three call sites) and § 9 says that such a page turns Wide later.
- **M-2: the Citation composer's own breakpoint.** `CitationComposerView.composeSplit` switches its evidence area from two-up to three-up at 1400pt with a private `GeometryReader`. That is a second breakpoint beside the new 1600 class. Moving it onto `pvWidthClass` in this PR would change the composer for panes between 1400 and 1599pt, against US-2 ("nothing I see changes"). **Recommendation:** leave it in this PR, logged as an Out row, and move it in the composer's next brief, where its Wide frame is drawn.

## Writer's own calls

- D-5: the host fills the detail area before measuring. Round 2 found the host otherwise measures its child.
- D-9: the test-mount helper moves the three existing copies in this PR (review M-3), which touches two test files outside the feature.
- Version bump none: nothing a researcher sees changes.

## Round 1

### M-1 · blocker · D-6 bends the working rule on inspectors
The working rule and § 9 measure the pane minus any inspector; D-6 measured Properties whole and deferred the fix.
**Answer:** Fixed — D-6 now re-measures `listCard` in `PropertiesContent`; new In row, AC-1.6, T-3, STEP-3, manual check 4; the Out row is gone. § 9 states what a re-measured column excludes.

### M-2 · major · `CitationComposerView` already has a 1400pt breakpoint
**Answer:** Fixed (docs) — named in Current state and Codebase fit, and an Out row. The code move is **To user** (above).

### M-3 · major · T-2 adds a fourth copy of the borderless-window mount
**Answer:** Fixed — D-9 adds `hostInBorderlessWindow` in `macos/ProvenenciaTests/HostedWindow.swift` and moves `mountHost` and the two `PVSelectPlacementTests` mounts onto it (STEP-4, Codebase fit, Deletes).

### M-4 · minor · S14-D1 says the page measures itself and calls the width class a token
**Answer:** Fixed in the brief directly (domain fact "The pane, not the screen"; inventory row Layer Component), so it is true before S14-24 lands.

### X-1 · blocker · T-2 cannot tell `.regular` from "not measured"
**Answer:** Fixed — T-2 waits for `.wide` at 2000, then resizes the same window to 1200 and waits for `.regular`.

### X-2 · major · The probe's size is not specified
**Answer:** Fixed — STEP-4: the probe is `.frame(maxWidth: .infinity, maxHeight: .infinity)` inside the modifier's content, mounted with the helper at the window's size.

### X-3 · major · AC-1.5 and AC-2.2 have no automated test
**Answer:** Fixed in part — § 8 says AC-1.5 is manual check 1 because a probe cannot sit inside a destination without changing it; T-3 covers the nesting the host and Properties rely on. AC-2.2 stays a manual check: it is a scroll position in a real Source page.

### X-4 · major · Manual checks need a wide display
**Answer:** Fixed — § 10 says steps 3 and 5 are reported as not run without a display wider than about 1864pt, and T-2 and T-3 are the evidence.

### X-5 · minor · Doc paths and placements
**Answer:** Fixed — STEP-5 names `docs/deployment-plan/spike-14/completed.md` and the map, puts the README row after `FlowLayout`, and the D-8 note as one sentence under the Token file map.

### X-6 · minor · pbxproj entries miscounted
**Answer:** Fixed — Current state lists the five app entries and four test entries; STEP-2 says new IDs are checked with `rg` before use.

### X-7 · minor · Small unstated choices
**Answer:** Fixed — STEP-2 gives the doc comments' content, STEP-3 the two comments verbatim, STEP-4 `@MainActor` and `@Test(arguments:)`, D-10 `Sendable` with a plain `static let`, STEP-2 the build command.

### X-8 · nit · Temptations
**Answer:** Fixed — Out row "An animation or transaction on a class flip".

## Round 2

Reviewers: maintainability (Opus), executability (Sonnet). Fresh agents with the plan and this log.

**Round 1 answers.** Executability: X-1, X-2, X-4 – X-8, M-1, M-4 hold; X-3 holds in part (below, X-9). Maintainability: M-2, M-3, M-4 and X-1 – X-8 hold; M-1 does not (below, M-5, M-6).

### X-9 · blocker · The host measures its child, not the pane
`WorkspaceDestinationHost.body` has no frame, and several branches render nothing; a hugging root would read Regular at any width.
**Answer:** Fixed — D-5 adds `.frame(maxWidth: .infinity, maxHeight: .infinity)` before the modifier; Current state row corrected; AC-1.6 and T-3 (a non-filling probe under the frame reads Wide); a Stop condition for a destination that moves.

### X-10 · major · Where the Properties re-measure goes
**Answer:** Moot — D-6 no longer re-measures (To user above).

### X-11 · major · The helper does not fit the `PVSelectPlacementTests` mounts
**Answer:** Fixed — D-9 gives the signature: `windowSize`, `hostSize` (default the window's), `orderFront` (default true); the PVSelect cases pass 900 × 200, 900 × 80 and `orderFront: false` and keep their own layout call; callers hold the window.

### X-12 · minor · Small T-2 / T-3 choices
**Answer:** Fixed — T-2 resizes with `window.setContentSize`; each T-3 case is its own mount; STEP-4 repeats the `rg` rule for the test files' IDs.

### M-5 · blocker · Metadata and Source types also have a side pane
**Answer:** To user (above), with the recommendation applied: no page re-measures; Out row and `docs/ideas/wide-layout.md` list the three side-pane pages for their own Wide briefs.

### M-6 · blocker · The re-measure bent the working rule
**Answer:** To user (above). The plan's **Rules this plan bends** names it; the rule is reworded on this branch pending the answer.

### Noted below blocker
- The S14-24 map scope entry omitted round 1's changes. **Answer:** Fixed.
- § 9's intro said pages sit left-aligned. **Answer:** Fixed — left on the Source page, centered elsewhere.
- The helper's `orderFront` against the PVSelect tests. **Answer:** Fixed with X-11.

**Verdict:** Fits with changes — Ready with changes. No open blocker once the two To user items are answered.
