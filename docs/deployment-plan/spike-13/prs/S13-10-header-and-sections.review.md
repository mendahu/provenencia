# Review — S13-10 Header and sections (and brief S13-D1)

**Verdict:** Fits with changes — Not ready → every finding Fixed; plan stays `Blocked — on board S13-D1`
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet). Both reviewed the plan and the brief together. Round 2 runs when the plan is refreshed after the board (Q-1 – Q-5), and checks the round-1 fixes against the code with it.

## To user

- **M-2 (part).** The segmented picker (`PVChipGroup(style: .segmented)` + `PVChip(…, expands: true, selectionLift: true)` + a group label) is hand-written in `DateValueEditorForm`, `CitationComposerAuthoringBox` and `PropertiesView`; the Person switcher would be the fourth. Extract a selection-driven segmented initializer on `PVChipGroup` and move all four onto it? **Recommendation:** not in S13-10. It touches three unrelated features and their tests in a Person page PR. Add a small kit PR to the Spike 14 odds-and-ends batch instead. S13-10 already moves the selected trait and focus ring into `PVChip` itself, so the copies do not drift on accessibility meanwhile.

## Writer's own calls

- **D-1, kind-neutral sections.** M-1 accepted: one `ConclusionPageSection` and one `pageSection` field, with `offered(by:)` per `WorkspaceSection`, instead of a Person-only field. The section enum lives in `WorkspaceLocation.swift`; Event and Place adopt sections without a second field.
- **D-2, the first section is stored as nil.** Vitals by argument, by default, and by `withPageSection(.vitals)` are one location.
- **D-6, `PVChip` is extended for every caller.** Ten existing chip call sites gain the selected trait and a keyboard focus ring.
- **D-9, Tab, not arrow keys.** Arrow-key movement between chips is dropped; the brief's PPF-13 now asks for Tab.
- **D-13, "vitals" renamed in the header.** `PersonDetailContent.Vital`/`vitals`, `PersonDetailVitals` and `L10n.Conclusions.a11yVital` become life lines.
- **Glyphs.** Family `gitBranch`, Timeline `calendar`, Gallery `photo`, Sources `library`; the board can swap within existing symbols.
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
