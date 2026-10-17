# Review — S10-35 Person gallery and Persons list thumb

**Verdict:** Fits — Ready with changes (applied)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. The four product calls were answered before review (2026-10-17): "earliest" is first recorded; one PR, not split; the agreed gallery frames match PDP-7 to PDP-9 and PSL-7 as the briefs word them; S10-35 ships after S10-34 so its manual check can draw faces.

## Writer's own calls
- D-1: the gallery is a separate list beside the fields. S10-52 D-9's skip of depiction fields stays, which goes against S10-52's Next adds ("S10-35 replaces D-9's skip").
- D-8: the lead likeness is stored with the cached Person header and chosen in `onePerson`. S10-39 changes only that choice.
- D-17: `PVList` gets a `Thumbnail` view generic so a row can load an image.
- D-11: each Conclusion kind supplies its whole header thumb (mark, label, image), as S10-23's Next adds said.
- Header thumb label is "Likeness", by role (round-2 maintainability note), so S10-39's chosen portrait reads the same.
- D-7: a tile reuses the Why record (`ReconcilerOutcomeDetail`) of the value's earliest Observation, so it opens the composer the way a Why record does.

## Round 1

### M-1 · blocker · The gallery's PNG check is a second copy of the rule `Readable` owns
The plan exported `depictionvalues.ImageMediaType` and had `likenesses` compare the joined media type with it. The `depictionvalues` package doc says every reader that returns a depiction goes through `Readable`.
**Answer:** Fixed. D-6: `depictionvalues.ReadableImage(imageFileID, mediaType) []byte` *(new)* is `Readable`'s image step. `Readable` and `likenesses` both call it, and nothing is exported for the literal. D-3, §4, Codebase fit, and Go core were updated to match.

### M-2 · major · The tile caption re-implements the Why record's Source-title fallback
**Answer:** Fixed. D-15: `ReconciliationRecord.sourceTitle(for:)` *(new)* sits beside `composerLocation(for:)`. `ReconciliationRecord.init` and the gallery mapping both call it. Added to Codebase fit and to STEP-8's files.

### M-3 · major · Next adds sends S10-39's portrait choice to Swift; the names lock "earliest" into the lead slot
**Answer:** Fixed.
- Next adds: S10-39 changes only the choice in `onePerson`.
- D-8 and the proto doc describe the field by role ("the likeness the Person leads with; S10-35: the earliest").
- The list accessor is renamed to `ConclusionListKind.thumbnailRelPath`.

### M-4 · minor · `CachedThumbnail` gains a second fallback channel; its doc goes stale
**Answer:** Fixed. D-13 narrows the parameter to `fallbackMark: PVMarkKey?` plus `fallbackLabel`, and the type doc's preference line is updated. Codebase fit says why `typeIconKey` stays separate: it validates a catalog key to the `.type` family. T-10 follows the new signature.

### M-5 · minor · Tests for the new `likenesses` package live in another package's test file
**Answer:** Fixed. T-2b adds `core/database/likenesses/likenesses_test.go` `TestForEntities`, covering no ids, a handle with no likeness, earliest-first, and a redraw keeping its place. T-2 keeps the denied, provisional, and weak cases, because only the reconciler fixture builds those worlds; the T-2b row says so.

### M-6 · minor · The tile shell is the second copy of the artifact card's preview button
**Answer:** Fixed as a log: Next adds tells S10-39 to run `evaluate-ui-component` on the shell before writing a third copy. The snowflake stays in this PR, as the brief requires.

### M-7 · nit · Two different `Likeness` types
**Answer:** Fixed. The detail type is now `conclusiondetails.LikenessTile`.

### X-1 · blocker · The plan is not in a state the implementer may start from
**Answer:** Fixed by this review. The plan goes to Planned only after round 2, with this log linked.

### X-2 · major · `PersonDetailContent.thumbnail` has no `header` to read
**Answer:** Fixed. D-11: `PersonDetailContent` stores `likenessImageRelPath`, set in `init` the way `treatment` is. D-15 stores `likenesses` the same way.

### X-3 · major · `loadOutcomes` changes shape beyond the signature given
**Answer:** Fixed. D-7 gives:
- the full signature, `(map[string]Outcome, error)`;
- where a lead row goes in the `!ok` branch;
- the one-row-per-key guarantee, from the primary key;
- that lead `Recorded` values have no date or name, and that `conclusionValueProto` maps that to an empty value.

### X-4 · major · No test helper for saving two likenesses on one Person
**Answer:** Fixed. T-3, T-5, T-6, and T-7 name their helpers: `observationstest`, `f.promote(b, h)` for the second member, and `likenessFixture` / `saveLikeness` / `PromoteSubject` in the handler tests. Each also names the save order that makes A earliest.

### X-5 · major · STEP-6's check cannot confirm the `PVThumbnail` line
**Answer:** Fixed. STEP-6's check says the image-state line is checked by hand. Manual check step 3 sets gender Woman and looks for the `lineMaternal` outline with no fill (AC-3.2).

### X-6 · minor · The new thumbnail fields leave several choices open
**Answer:** Declined. The reviewer found nothing to change; T-8's initializer follows D-11's field order.

### X-7 · minor · Three structs hold `mark` and `thumbnailLabel`
**Answer:** Fixed. D-11 and the stop condition name `ConclusionDetailPage`, `ConclusionDetailLoaded`, and `ConclusionDetailHeader`.

### X-8 · minor · The section's initializer is not stated
**Answer:** Fixed. D-16 gives the struct's properties (`projectDir`, `likenesses`, navigation environment, focus state) and where `ConclusionDetailLoaded` draws it.

### X-9 · minor · Section accessibility label source
**Answer:** Fixed. §6 macOS uses `L10n.string(L10n.Conclusions.likenessesSection)`.

### X-10 · minor · The manual check cannot be run until S10-34 lands
**Answer:** Fixed. §10: if S10-34 has not merged, the implementer reports the manual check as not run, and the PR does not merge until someone runs it on a build that has S10-34.

### X-11 · minor · DoD brief-archive line
**Answer:** Declined. The DoD line already says S10-D12 stays in `design/` for S10-39.

## Round 2

Maintainability: every round-1 answer holds. No new blocker. **Verdict: Fits.** One note below blocker level: the header label "Earliest likeness" would need rewording in S10-39. **Answer:** Fixed. The copy is "Likeness" (§6 Strings, AC-3.2, §10).

Executability: every round-1 answer holds, and no Current state row is false. **Verdict: Ready with changes.**

### X-12 · major · The plan never says the cache is current when a header is composed inside a save
**Answer:** Fixed. D-8 states the order in `core/writes/run.go` `Run`: recompute, then the search reprojection (`store=false`), then commit, then `OnCommit` drops the stored rows. Stored rows are composed only from the committed graph. T-5 also checks the lead after the first save.

### X-13 · major · T-2b has no fixture in the new package
**Answer:** Fixed. T-2b copies `conclusiondetails_test.go` `newFixture`, `source`, `personOn`, and `promote` into a package-local fixture and names the setup calls. T-3's `likenessOn` gets its full signature.

### X-14 · major · Most step checks cannot run on Linux
**Answer:** Fixed. A **Machines** note under the Implementer block says which steps need protoc or a Mac, and that a check that cannot run is reported as not run.

### X-15 · minor · D-7 leaves tile assembly open
**Answer:** Fixed. D-7 says:
- `leads` is built from each likeness's `ObservationIDs[0]`;
- field outcomes still attach as today;
- tiles are built in `ForEntities` order.
Stop condition 2 also covers a lead that has no outcome row.

### X-16 · minor · Empty dash before load when a path is present
**Answer:** Fixed. D-13 says this is accepted and stays as every `CachedThumbnail` behaves today.

### X-17 · minor · The section's outer layout is open
**Answer:** Fixed. D-16 gives the `space9` page gap, the `VStack(spacing: 0)`, and where the accessibility modifier goes, following `ConclusionDetailRelationshipSection`.

### X-18 · nit · "aside" in the brief versus the `meta` slot
**Answer:** Fixed. The §6 macOS row says the brief's "aside" is the kit header's `meta` slot.
