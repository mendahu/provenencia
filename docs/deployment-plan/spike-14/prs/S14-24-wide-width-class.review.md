# Review — S14-24 Give every workspace page its width class

**Verdict:** Fits with changes — Ready with changes (round 1); round 2 below
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet, the implementer's model)

## To user

- **M-2: the Citation composer's own breakpoint.** `CitationComposerView.composeSplit` switches its evidence area from two-up to three-up at 1400pt with a private `GeometryReader`. That is a second breakpoint beside the new 1600 class. Moving it onto `pvWidthClass` in this PR would change the composer for panes between 1400 and 1599pt, against US-2 ("nothing I see changes"). **Recommendation:** leave it in this PR, logged as an Out row, and move it in the composer's next brief, where its Wide frame is drawn.

## Writer's own calls

- D-6: a page with an inspector measures its content column, which also excludes the gap and the gutters, so Properties turns Wide at a wider window than a page with no inspector. § 9 says so. The alternative, measuring "pane minus inspector" exactly, needs the page to know its gutters.
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

See the end of this file once round 2 reports.
