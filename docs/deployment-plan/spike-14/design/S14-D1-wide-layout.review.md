# Review — S14-D1 Wide layout in the design system

**Verdict:** Fits with changes — Ready, after 2 rounds
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- The pane definition for a page with a side pane is the S14-24 To user item (its review log); this brief already says a page with a side pane plans it in its own Wide frame.

## Writer's own calls

- One gutter rule: `--gutter-page` at both widths, and centering past the maximums widens them. `--gutter-page-wide` stays unused.
- Regular centers the column (as the Person, Event and Place pages and lists do today); the Source page left-aligns today and will move when it adopts Page columns in its own brief.
- WL-5's sizing: the primary grows first, then the secondary, then the group centers. 1600 gives a 1164 primary for the 16pt between 1600 and 1616.
- Page columns owns the scroll; the page owns its background. Vertical padding `--space-8` top and bottom, matching the two Conclusion scaffolds, so moving them changes nothing.

## Round 1

### M-1 · blocker · Page columns ships with no caller
**Answer:** Fixed — § 3 ships `ConclusionDetailPage.page` and `ConclusionListPage.page` onto it, primary only, with no visible change; the map's S14-25 row and scope entry match.

### M-2 · major · A Wide-only gutter that the rest of the chrome cannot follow
**Answer:** Fixed — WL-4, WL-5 and WL-7 use `--gutter-page` at both widths; § 9 rule 6 says gutters are never narrower than at Regular.

### M-3 · major · No slot for the Source page's pinned header
**Answer:** Fixed — WL-3 adds a header slot that scrolls or is pinned; WL-8 aligns it with the group; WL-W4 and frame 7 draw it. The Source page adopts it in its own brief.

### M-4 · major · The rules in five copies, already different
**Answer:** Fixed — § 9 and the `add-design-brief` Paste block now carry the same six rules word for word, each saying to change both; WL-2 copies the brief's own block into the Claude Design kit README; the Swift kit README links to § 9 (§ 3). Copies: § 9, the Paste block, the Claude Design kit.

### M-5 · major · The 480 maximum has no token
**Answer:** Fixed — WL-1 adds `--width-secondary-max: 480px`; inventory row; S14-25 transcribes it.

### M-6 · minor · Per-page measuring and the Token layer
**Answer:** Fixed — domain fact reworded; Width class row is Component.

### M-7 · minor · The Conclusion list page is not a Page columns page at Wide
**Answer:** Fixed — the inventory says the lists stay primary-only until their own list-beside-detail brief.

### X-1 · blocker · 56 gutters contradict the 1600 arithmetic
**Answer:** Fixed with M-2 — one gutter rule; WL-5 states what shrinks between 1600 and 1616.

### X-2 · blocker · Secondary growth undefined; frame 3 already capped
**Answer:** Fixed — WL-5 gives the formula and the panes where each column fills (1616, 1756); frame 4 (1920) grows, frame 5 (2560) is capped.

### X-3 · blocker · No header slot; scroll and states unowned
**Answer:** Fixed — WL-3 (slots, scroll, background, states, padding), WL-8.

### X-4 · blocker · "Replaces three scaffolds" against "adoption stays out"
**Answer:** Fixed with M-1 — the inventory names the two first callers in S14-25 and the Source page as next.

### X-5 · major · Regular alignment undecided
**Answer:** Fixed — WL-4 centers; frame 3 shows a 1464 pane.

### X-6 · major · A custom property cannot drive `@container`
**Answer:** Fixed — WL-10: a literal 1600px on the page element's container query, with a comment naming the token.

### X-7 · major · No component interface
**Answer:** Fixed — WL-3 (three slots, scroll, padding), WL-4 (stacked gap), WL-9 (focus order). Swift names are S14-25's plan.

### X-8 · major · Thin Wide requirements for the board
**Answer:** Fixed — WL-W1 – WL-W4, each with its frame; WL-W3 pairs Regular and Wide.

### X-9 · minor · Wording and gaps
**Answer:** Fixed — window widths 1863 and 1864 in frame 8; the scroll rule is a note; handoff writes `completed.md`; WL-2 says the Claude Design kit README and § 3 the Swift one; the domain fact says the Source page is one column at page level.

## Round 2

Reviewers: maintainability (Opus), executability (Sonnet). Fresh agents with the brief and this log.

**Round 1 answers.** All hold except M-1's "changes nothing on screen", which depended on slot spacing (M-8). The WL-5 arithmetic, the frames' window and pane widths, and the six rules (word for word in § 9, the Paste block and this brief) were checked.

### M-8 · blocker · "Changes nothing" depends on slot spacing; the header slot has no caller
**Answer:** Fixed — WL-3: a slot holds the page's view as it is and Page columns adds no spacing inside it; § 3: the two Conclusion scaffolds keep their own `VStack` and header in the primary slot; the header slot and pinned mode are drawn on the board (WL-W4) but built with their first caller, the Source page, listed in `docs/ideas/wide-layout.md`. Map scope entry matches.

### X-13 · minor · CSS and Swift panes differ for a page with an inspector
**Answer:** Fixed with S14-24's To user recommendation — both measure the whole pane; the domain fact says so.

### X-14 · nit · The secondary slot's states
**Answer:** Logged for S14-25's plan: what the secondary slot shows while its content loads or fails.

**Verdict:** Fits with changes — Ready. No open blocker.
