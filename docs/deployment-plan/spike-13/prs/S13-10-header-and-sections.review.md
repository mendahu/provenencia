# Review — S13-10 Header and sections (and brief S13-D1)

**Verdict:** Round 1 Fits with changes — Not ready; round 2 Fits with changes — Ready with changes → every finding Fixed; the plan is Planned
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet), fresh each round. Round 1 reviewed the plan and the brief; round 2 reviewed the Wide rework and the agreed board's answers.

## To user

- **M-2 / M2-4 — answered 2026-10-29.** Extract the segmented picker? The writer recommended a separate kit PR. **User:** "Let's fix it on this plan. I think we have a segmented picker in the design kit, so we should definitely be consolidating this and moving all the callers over to it." The kit's segmented control is `PVChipGroup variant="segmented"` with `PVChip expands selectionLift` (S14-D1 design system, `PVChip.jsx`). Plan D-20 builds `PVSegmentedControl` and moves the three callers; map scope S13-10 "One segmented control".

## Rework after the Wide layout landed (2026-10-29; reviewed in round 2)

`main` gained the Wide width class (S14-24, #628), the S14-D1 Page columns board, the binding "Wide layout" section of `add-design-brief`, and the working rule "Redesigns adopt Page columns", which folds S14-25 into S13-10. The branch was rebased; the map conflict kept `main`'s new S13-10 scope and dependencies with this branch's brief, plan and status. Changes since round 1:

- **Brief S13-D1:** the "Two widths — draw both" Paste block; Page columns in the domain facts, the gate and the inventory (New, built in S13-10); PPF-W1 – PPF-W5; a Wide layout table; Wide frames 14 – 19; non-goals for Wide-only content and new Page columns modes.
- **Plan:** US-3 (Wide) and AC-4.4 (sizing); D-14 `PVPageColumns` from WL-3 – WL-9 with one pure `columnWidths` and a `Layout`; D-15 `widthSecondaryMax` and the breakpoint check; D-16 `ConclusionDetailPage.page` on Page columns; D-12 moves the scroll reset into `PVPageColumns`; D-5 now hands the page `ConclusionDetailSectioning` data and the page draws the switcher in the header slot, so pinned or not needs no second view slot; D-11 adds the kit file, its tests and their pbxproj entries (S14-24 precedent); Q-1 now uses Page columns' pinned mode; Q-6 asks the board for Vitals' and Gallery's Wide arrangement; T-12 – T-14; ten steps.

Round 2 runs on this rework together with the board's answers.

## Writer's own calls

- **D-1, kind-neutral sections.** M-1 accepted: one `ConclusionPageSection` and one `pageSection` field, with `offered(by:)` per `WorkspaceSection`, instead of a Person-only field. The section enum lives in `WorkspaceLocation.swift`; Event and Place adopt sections without a second field.
- **D-2, the first section is stored as nil.** Vitals by argument, by default, and by `withPageSection(.vitals)` are one location.
- **D-6, `PVChip` is extended for every caller.** Ten existing chip call sites gain the selected trait and a keyboard focus ring.
- **D-9, Tab, not arrow keys.** Arrow-key movement between chips is dropped; the brief's PPF-13 now asks for Tab.
- **D-13, "vitals" renamed in the header.** `PersonDetailContent.Vital`/`vitals`, `PersonDetailVitals` and `L10n.Conclusions.a11yVital` become life lines.
- **Glyphs.** Family `gitBranch`, Timeline `calendar`, Gallery `photo`, Sources `library`; the board can swap within existing symbols.
- **D-5 (rework), the page draws the switcher from data.** Rather than a header-accessory view slot, a kind passes `ConclusionDetailSectioning` (sections, selected, select).
- **Q-6 recommendation.** Vitals and Gallery primary-only and centered at Wide in S13-10; the secondary column waits for S13-D2 and S13-D6.
- **Version.** None, following S13-09, S14-01 and S14-02 (no bump until a release is cut), not "minor" as first drafted.
- **Section keys.** `conclusions.section.*` and `conclusions.gallery.empty.*`, named so S14-02 can move the gallery keys onto its slot.

## Round 1 — maintainability

### M-1 · blocker · The section field and its crumb and jump-row branches only work for Person; the named next kind has to copy them
**Plan:** D-1, D-2, D-8, Next adds. **Codebase:** `WorkspaceLocation` (`==`, `init`, `init(from:)`, `CodingKeys`), `WorkspaceToolbar.breadcrumbItems`, `HistoryJumpRow.leafLabel`; R-1 and working rule "One sectioned Conclusion page" name Event and Place as later adopters. **Better shape:** one `pageSection: ConclusionPageSection?` with a per-`WorkspaceSection` list of offered cases.
**Answer:** Fixed — D-1 (`ConclusionPageSection`, `offered(by:)`), D-2 (`pageSection`, `withPageSection`, one normalizer), D-8 reads `pageSection` kind-blind, AC-3.3, T-4, T-11, Next adds.

### M-2 · major · `PVChip` lacks the selected trait and focus ring; the segmented composition is hand-written three times
**Plan:** D-6, D-9, Stop condition 3; brief switcher row Ship. **Codebase:** `PVChip` (no trait, no `pvFocusRing`); `PVRadio`, `PVIconButton`, `PVSidebarNav`, `PVTable` own their trait; copies in `DateValueEditorForm`, `CitationComposerAuthoringBox`, `PropertiesView`.
**Answer:** Fixed (trait and focus ring) — D-6 extends `PVChip`, STEP-4, Reach row, brief inventory row now **Extend**. The segmented extract → **To user** above.

### M-3 · major · Scroll-to-top belongs to the page that owns the `ScrollView`; Q-1's framing adds a second slot
**Plan:** D-12, Q-1. **Codebase:** `ConclusionDetailPage.page(identifier:body:)` owns the only `ScrollView`; `ListScrollAnchorModifier`; `SourcesListView` pinned headers.
**Answer:** Fixed — D-12 now: `scrollResetKey` on `ConclusionDetailPage`, a `ScrollViewReader` in `page`, nil for Event and Place; STEP-5; AC-2.9. Q-1 reframed: pinning stays inside one `ScrollView` and one slot (the `SourcesListView` precedent); a split goes to the user.

### M-4 · major · S14-02 renames and re-slots the gallery code this plan reuses
**Plan:** D-4, D-5, D-10, §4. **Codebase:** S14-02 plan D-1 – D-3; Spike 14 working rules "The gallery layer is named gallery".
**Answer:** Fixed — §4 Spike 14 row, header Depends on note, §3 Out row (the rename is S14-02's), D-10 names the gallery keys so S14-02 moves them unchanged, §11 risk row (S14-02 first → refresh D-4, D-5). The Spike 14 map's S14-02 scope gets one line naming S13-10.

### M-5 · major · Breadcrumbs and jump rows each derive the location label
**Plan:** D-8, T-7. **Codebase:** `breadcrumbItems` leaf `title ?? ref ?? "…"`; `HistoryJumpRow.leafLabel` `title ?? ref`.
**Answer:** Fixed — D-8 adds `WorkspaceToolbar.locationLabels(for:)`, used by both; T-7 is now a unit test.

### M-6 · major · "Vitals" names two things in the same file
**Plan:** Codebase fit note; Out row. **Codebase:** `PersonDetailVitals`, `PersonDetailContent.Vital`/`vitals`, `L10n.Conclusions.a11yVital`, `PersonDetailTests` lines 268, 269, 318.
**Answer:** Fixed — D-13, STEP-2, DoD sweep names the old names; Out row removed.

### M-7 · minor · Next adds does not name where later sections' queries load
**Answer:** Fixed — Next adds: each section's key is warmed from the `.personDetail` spec's `queryKeys`, which receives the location; the section reads a `QueryHandle`.

### M-8 · minor · D-1's glyph has no values; PPF-10 invites new marks
**Answer:** Fixed — D-1 names one existing `PVSymbol` per case; brief PPF-10, inventory row and a non-goal say no new mark.

### M-9 · minor · Brief PPF-12 leaves the narrow-window fallback unbound
**Answer:** Fixed — brief inventory row "Narrow-window fallback" (`PVSelect` or shorter labels) and a non-goal (no overflow menu, no scrolling strip).

## Round 1 — executability

### X-1 · blocker · `WorkspaceLocation.init` gets no `pageSection` parameter; normalization is unplaced for `init(from:)` and the copy helper
**Answer:** Fixed — D-2: memberwise parameter `pageSection: … = nil`; one private `normalizedPageSection` that `init`, `init(from:)` and `withPageSection` all call; T-2, T-4 cover each path.

### X-2 · major · D-12 has no mechanism for the unpinned case; PPF-7 is not checked
**Answer:** Fixed — see M-3; AC-2.9 and §10 step 5.

### X-3 · major · The per-section glyph is not chosen
**Answer:** Fixed — see M-8.

### X-4 · major · Arrow-key focus movement between `PVChip` buttons is asserted without proof
**Answer:** Fixed — D-9 drops arrow keys; Tab moves focus; brief PPF-13 changed to match; Out row.

### X-5 · major · The breadcrumb and jump-row edits are underdetermined
**Answer:** Fixed — D-8 names both crumb ids and actions, the ref fallback, and that a Vitals entry has no suffix; brief PPF-6 and frame 8 now say the Vitals entry reads the name alone.

### X-6 · major · T-7 is conditional and `HistoryJumpRow` is private
**Answer:** Fixed — see M-5.

### X-7 · minor · T-10 can contradict Q-4; the title must be optional
**Answer:** Fixed — `ConclusionDetailRowsSection.title` is optional (D-4); T-10 no longer checks `detailsSection`.

### X-8 · minor · The Gallery heading inside the tab is unspecified
**Answer:** Fixed — §1 What stays and AC-1.2: the "Likenesses" heading and count stay inside Gallery until S13-14.

### X-9 · minor · Call-shape details
**Answer:** Fixed — D-5 uses `message: L10n.string(…)`; D-9 takes the index from `enumerated()`; D-10 gives every key name and requires a `comment:`; Steps give full paths; D-3 keeps the toast overlay and `portrait`.

### X-10 · minor · Version bump target not stated
**Answer:** Fixed — the bump is none (S13-09, S14-01, S14-02 precedent: no bump until a release is cut), so no `project.pbxproj` edit.

### X-11 · nit · Ordering and Edge notes
**Answer:** Fixed — STEP-6 says it runs after the board; D-1 says raw values are frozen.

## Board answers (2026-10-29)

The S13-D1 board was agreed; its archive was checked against the brief (frames 4a – 4y rendered and compared). Every frame is the pane only, with no sidebar. It answered Q-1 – Q-6 and drew five changes the plan absorbed: 36pt chips with glyph + label (D-6), a compact header on scroll (D-17), the refresh indicator in the switcher row (D-18), a 28pt reserved slot on Vitals rows (D-19), and per-section empty-state lines (D-10). Frames F2 (pattern B), F11 (narrow, set aside at the user's request) and F15 (Wide growing, moot with no secondary column) were not drawn.

## Round 2 — maintainability

### M2-1 · blocker · `PVChip.Size { regular, large }` is a second control-size vocabulary beside `PVControlSize`
**Answer:** Fixed — D-6: `size: PVControlSize = .md`, height from `size.height`, glyph size from a private `PVControlSize.chipGlyphSize`.

### M2-2 · major · `ConclusionDetailSectioning` and D-3's glue repeat what the page can derive
**Answer:** Fixed — D-5: the page reads navigation, draws the switcher and pins when `offered(by:)` is non-empty, and passes `section` in `ConclusionDetailPageSectionsContext`; `ConclusionDetailSectioning` is gone; D-3: `PersonDetailView` no longer reads navigation; T-10 checks it.

### M2-3 · major · `reservesProvenanceSlot: Bool` is a one-caller flag S13-11 must pile onto
**Answer:** Fixed — D-19: `ReconciledValueRow.Provenance { why, sheet }`; S13-11 changes only `.sheet`.

### M2-4 · major · Round-1 M-2 not answered: the switcher is a fourth segmented copy
**Answer:** Fixed with the user's answer (To user above) — D-20 `PVSegmentedControl`, STEP-3 moves the three callers, AC-4.1, AC-4.2, AC-5.5, T-16, T-17.

### M2-5 · major · D-12's reset is a second `ScrollViewReader` beside `PVScrollRowIntoView`
**Answer:** Fixed — D-12 extends `pvScrollsRowIntoView` with `restartsAt:whenChanged:`; `PVPageColumns` applies it.

### M2-6 · major · `PVChip`'s focus ring via a private `@FocusState`
**Answer:** Fixed — D-6: a private `PVChipStyle: ButtonStyle` reads `\.isFocused`, as `PVButtonBody` does; no `.focused` binding in the kit.

### M2-7 · major · The pinned bar's kit fill contradicts WL-3
**Answer:** Fixed — D-14: no fill; only the rule spans the pane; the page keeps its background.

### M2-8 · major · `add-design-brief` does not offer Page columns
**Answer:** Fixed — STEP-9 adds "Page columns" and "Segmented control" rows to the Paste block's Reach for (kit) and replaces the "raised once a second view needs it" sentence; DoD line.

## Round 2 — executability

### X2-1 · blocker · The compact header's one-line summary has no mechanism
**Answer:** Fixed — D-17, D-21: `\.conclusionDetailHeaderCompact` environment value; `PersonDetailLifeLines` switches its stack.

### X2-2 · blocker · The scroll reset lands at 24pt, which reads as scrolled
**Answer:** Fixed — D-12: the anchor is the zero-height sentinel before the `space8` padding, at offset 0; AC-2.9 now says the full header returns.

### X2-3 · major · Sentinel, offset sign and short-content oscillation underspecified
**Answer:** Fixed — D-17 names the coordinate space, `offset = −minY`, the sentinel's place, no animation, and `compactMinOverflow` (120pt) so compacting cannot undo itself; AC-1.10; T-15 covers overflow 119 / 120.

### X2-4 · major · `PVPageColumns` API and layout guesses
**Answer:** Fixed — D-14 gives the three inits, `hasSecondary` from `Secondary.self`, the band layout sharing `columnWidths`, the pinned and unpinned stacks, no fill; D-16 gives the band's padding from the board; T-14 names `hostInBorderlessWindow` and `waitUntil`.

### X2-5 · major · D-16 page plumbing
**Answer:** Fixed — D-16: `page` deleted; one `PVPageColumns` in `ConclusionDetailLoaded` with the header closure and a `switch` in the primary; no new page property; `refreshing` comes from the presentation; band spacing stated.

### X2-6 · major · `locationLabels` would change non-Person jump rows
**Answer:** Fixed — D-8: `leaf` is `title ?? ref` (today's jump-row rule), `String?`; the non-entity crumb branch is untouched; T-7 covers a Source location and a metadata root.

### X2-7 · major · STEP-2's check cannot pass
**Answer:** Fixed — STEP-2 searches `macos/App/Features`, `macos/App/Platform` and the tests (not the kit's "Vital records" preview) and adds `vital`; D-13 renames the closure variable and the test local.

### X2-8 · major · `PVChip` large/glyph details and Tab focus
**Answer:** Fixed — D-6: both inits, padding unchanged, the style carries focus; AC-1.6 and §10 say Full Keyboard Access is on for Tab.

### X2-9 · minor · D-14 needs a clamp
**Answer:** Fixed — `max(0, …)` in both primary formulas.

### X2-10 · minor · Board numbers missing
**Answer:** Fixed — D-16 and §6 give the band's paddings from the board (`space9`, `space7`, `space5`) and the rule.

### X2-11 · minor · Compaction should be gated
**Answer:** Fixed — only a pinned page compacts; unpinned `isScrolled` is always false (D-14, D-17).

### X2-12 · minor · Stale text
**Answer:** Fixed — Status Planned, map row Planned, Q-references removed, Review line updated.
