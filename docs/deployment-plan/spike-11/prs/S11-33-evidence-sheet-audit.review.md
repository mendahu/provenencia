# Review — S11-33 Evidence sheet audit follow-up

**Verdict:** Fits with changes — Ready with changes (round 2 blockers fixed by the writer; no open blocker)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None open. Reported with the plan: D-2 (confirms and notices now close on an outside click) and D-1's AC-1.6 (Cancel still discards at once while Escape and outside click ask) carry the user's "standardize closing" request; reported to the user with the plan.

## Writer's own calls
- D-1: one kit presenter `.pvSheet` with a hidden cancel-action button and the outside-click watcher; Cancel buttons stay immediate.
- D-4: `segmentcompare.ReadFrom` and the six-row reading table.
- D-15: the board's grid numbers stay named constants; only values equal to a token use it.
- D-16: `LikenessTile` folds back into the gallery.
- Out: menu/popup dismissal and app-wide word-splicing go to idea files.

## Round 1

### X-1 · blocker · D-14 changes the layout
`memberIndent` from tokens gives 32, not 44; `box` 16, not 18.
**Answer:** Fixed. D-15 keeps the board's grid numbers (`box` 18, `memberIndent` 44, the 116/200/170 columns) as named constants with a source comment; only `cropGap` and `relationshipIndent` take tokens (equal values). Nothing moves.

### X-2 · blocker · D-4 cannot be built from what Go has
**Answer:** Fixed. D-4 now defines `segmentcompare.ReadFrom(k, first, second, end, terms) (term, namesEnd)` beside `Key`, with the rule and six worked examples; `promotealign/evidence.go` fills the new fields after `claimevidence.Build`; labels and origins come from `graphcache.VocabTerm` through a new `termWording` lookup. STEP-2 names the files.

### X-3 · major · audit items not restated
**Answer:** Fixed. STEP-8 names each merge (`EvidenceSectionLines`, `EvidencePinBox`, `isDepiction`, `PromoteUpdatedCaption`) and its sites; STEP-11 lists every comment and preview with its new text.

### X-4 · major · T-2 and T-6 not writable
**Answer:** Fixed. T-2 tests the pure `PVSheet.dismissRoute(isRunning:isDirty:)` and the presenters' arguments; T-6 uses `EvidenceBoard.build(evidence:row:labels:worded:sourceTitles:)`.

### X-5 · major · D-9 storage and draft
**Answer:** Fixed. D-9: built in `PromoteModel.loadEvidence`, stored as `Row.board` replacing `Row.evidence`; D-10 drops the draft (M-4).

### X-6 · major · loading label; keys unnamed
**Answer:** Fixed. Loading moved out of T-11 to the manual check (AC-6.4); count keys `spokenRelationshipCount`, `spokenLikenessCount`, the twelve subtitle and heading keys, `titleHandle` and `cardHeaderLead` are named in §6 Strings.

### X-7 · major · T-10 untestable
**Answer:** Fixed. T-10 changes the FakeStore graph's bridges after load and checks `flow.bridges` still equals the loaded ones after a repropose.

### X-8 · minor · open choices
**Answer:** Fixed. Confirm and notice pass `isDirty: false` (D-1); the discard question's Escape returns to the form (D-1); D-3 notes no test reads the ids; D-13 has manual check 8; D-16 names the private subview; D-5 lists `SegmentCause` sites; STEP-12 gives the `completed.md` path.

### X-9 · minor · Current state inaccuracies
**Answer:** Fixed. The Go evidence row no longer calls `Key.End` the row's endpoint unconditionally; paths are full; enum lines corrected to :302 – :376.

### M-1 · blocker · flat strings instead of `ConclusionTerm`; event-type key dropped
**Answer:** Fixed. D-5: `neighbor_term` and `event_type` are `ConclusionTerm`; labels through `PropertyTermDisplay.name`, product keys through `productKey` (D-4 Swift side); Builds on lists both.

### M-2 · blocker · neighbor reading gets a second home; empty for half the terms
**Answer:** Fixed. `ReadFrom` sits in `segmentcompare` beside `Key` and never returns an empty term: `namesEnd` says whether it names the neighbor. Place and role links keep their wording; `term-management.md:28` updated in STEP-12.

### M-3 · major · `EvidenceBoard` beside `Row.evidence`; `EvidenceContext` survives
**Answer:** Fixed. D-9: the board replaces `Row.evidence` and carries the offered and checked pin ids; `EvidenceContext` is in Deletes; D-17 drops "context".

### M-4 · major · debounced draft
**Answer:** Fixed. D-10 keeps the direct binding.

### M-5 · major · Swift test needs a hand list
**Answer:** Fixed. D-7 does both directions in the Go test; T-8 is reduced to a few phrase spot checks.

### M-6 · major · new `SeededTermWording` file beside `L10n.PropertyTerm`
**Answer:** Fixed. D-6 extends `L10n.PropertyTerm.wording` in `L10n.swift`.

### M-7 · major · two Escape handlers
**Answer:** Fixed. D-1 has one route: the presenter's hidden cancel-action button; kit Cancel buttons drop their own shortcut; the evidence sheet presents through `.pvSheet`.

### M-8 · minor · Updated caption home
**Answer:** Fixed. `PromoteUpdatedCaption` goes in `PromoteReadout.swift` (renamed from `PromoteReadoutBadges.swift`).

### M-9 · minor · `PVThumbnailSize.likeness`
**Answer:** Fixed. `CachedThumbnail.likenessCrop`.

### M-10 · minor · `flushLeading` on the wrong control
**Answer:** Fixed. On `PVButton`; `PVDisclosureButton` forwards.

### M-11 · minor · one ellipsis wrapper
**Answer:** Fixed. D-14 `promote.sheet.cardHeaderLead`.

### M-12 · nit · `reloadEvidence` via `setSheetSubject`
**Answer:** Fixed. §6 macOS.

### Outside lens (UX) · Cancel on a dirty form
**Answer:** Fixed. AC-1.6: Cancel still closes at once; Escape and outside click ask.

## Round 2

### X-10 · blocker · D-4's reading of `Key.End` contradicts the code
**Answer:** Fixed, after checking `segmentcompare_test.go:57 – :67`. The examples were right in the key's reading, but the plan did not say which reading it meant. D-4 now says `k.End` is the row's endpoint in the key's reading (not always the stored end), drops `ReadFrom`'s `end` argument (the neighbor's end is the other endpoint), and gives eight worked examples in the test file's `rel(...)` form with their keys.

### X-11 · major · how `evidence.go` finds the neighbor's end
**Answer:** Fixed. D-4 "Who fills it": `first`/`second` from `connectrules.Bridges()` by `rel.Key.Type`; no `Link` carried through.

### X-12 · minor · T-3 file
**Answer:** Fixed. T-3 is in `core/segmentcompare/segmentcompare_test.go`.

### M-13 · blocker · `termWording` beside `EvidenceValues`
**Answer:** Fixed. D-4 extends `EvidenceValues` with `Term(property, key)`; no second lookup; `promotealign.Evidence` keeps its signature.

### M-14 · blocker · deleting the inverse map breaks the composer
**Answer:** Fixed with option (a). The sheet stops reading the map; the map and its parity subtest stay for `CitationComposerAuthoring.inverseTerm` until term-management item 2. STEP-3 check and STEP-12 line updated. (The audit summary to the user said B would delete the map; the user is told it stays for the composer.)

### M-15 · blocker · discard question's Escape has no route
**Answer:** Fixed. `.pvSheet` owns `confirmingDiscard` and renders the question; `PVSheetDismiss.action` gains `isConfirmingDiscard` and `.keepEditing`; T-1 gains the row.

### Outside lens · name format change; STEP-1 check
**Answer:** Fixed. D-9 reloads the open sheet's evidence when `flow.nameOrders` changes; STEP-1's check matches `sheet(item|isPresented` with or without a dot.
