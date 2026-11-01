# Review — S13-02 Gender in the person header

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus, fresh context), executability (Sonnet, fresh context)

Round 1 raised three blockers (M-1, X-1, X-2), all Fixed. Round 2 is not run: the fixes are local edits to Tests, Steps and Design, and no Reshape was raised.

## To user

- **M-3 — does `EventHeader` move onto `TermFacts` in this PR?** Writer recommended no. **User, 2026-11-01: yes, move it at the same time.** D-11 now moves `EventHeader.EventType` and `EventTypeCount` to `Type TermFacts` in Go; the wire and Swift's `CatalogEventHeader` stay as they are (Out row, Event page redesign).

## Writer's own calls

- D-1: Go carries both terms raw, with counts. Swift keeps the gender-before-sex rule and the colors.
- D-2: one `TermFacts` shape; `EventType` is renamed `Term`.
- D-3: the spoken label uses the Property's catalog label, because a product Property can be relabeled.
- D-5: only the product Properties feed the header. A researcher's own Property keyed `gender` stops coloring the badge.
- D-11 (after the user's answer): only the Go header moves; the event wire fields and `CatalogEventHeader` stay, so the PR adds no client change for events.

## Round 1

### M-1 · blocker · A second Term-to-wire mapper on each side of the FFI
`eventHeaderProto` already builds a `ConclusionTerm` literal, and `GoStore.mapConclusionValue` already maps one; the plan added a second copy of each.
**Answer:** Fixed. New D-10: Go `termProto` and Swift `GoStore.mapConclusionTerm` are the one mapper on each side. The existing literal and `.term` case call them (§6 Codebase fit, FFI, macOS; STEP-3, STEP-4).

### M-2 · major · `checkDisplay`'s person branch is not on the plan, and Verify will fail
`display_register.go` `checkDisplay` compares the stored person row with a fresh one by `reflect.DeepEqual`; finished fresh rows would no longer match.
**Answer:** Fixed. New Current state row; §6 Go core and STEP-2 make the person case compare with `finishPerson(v, row)`, as the event case does with `finishEvent`. §8 names the Verify tests that must stay green.

### M-3 · major · `EventHeader` keeps its own term pair beside the new `TermFacts`
**Answer:** To user; the user chose to move it (2026-11-01). Fixed: D-11, §3 In, §4 Term type row, §6 Go core and FFI, STEP-1, STEP-2, T-3, T-11, Edges "Event type count", DoD sweep. The wire and Swift event triple are an Out row.

### M-4 · major · `scanProperties` adds a third private scan loop
**Answer:** Fixed. D-3 and §6 Go core: `scanVocab` scans `id` plus every further string column. `loadVocab`'s properties query and the old `scanTerms` query both run through it, and `scanTerms` is deleted.

### M-5 · major · The plan makes "finished header is stored" false in two docs
**Answer:** Fixed. STEP-6 rewords the `conclusionheaders` package doc and the working rule "The family read is served from the graph cache" (relatives' tiles are read through `ByIDs`, which fills labels with no SQL when warm). Both are in the DoD.

### M-6 · minor · `CatalogTermFacts.term` can hold values that are not terms
**Answer:** Fixed. §6 macOS gives the property a doc comment: always `.term` or nil, built by `mapConclusionTerm`. Moving `CatalogEventHeader`'s flat triple follows M-3's answer.

### M-7 · nit · `propLabel` is keyed by id but always looked up by key
**Answer:** Fixed. `propLabel` is keyed by Property key, read through `propLabelOf(key)`.

### X-1 · blocker · STEP-1's check cannot pass: T-1 needs labels that STEP-2 adds
**Answer:** Fixed. T-1 asserts keys, counts and term presence only. The label assertions moved to T-2 (`TestPersonHeaderTermLabels`, case "labels"), which STEP-2 writes.

### X-2 · blocker · T-1 "two Sources disagree" needs a helper the plan doesn't name
**Answer:** Fixed. T-1 names `citeOn(art, s, in...)` *(new)* and gives the case's exact calls (`f.book`, `subjectstest.OfType` on the second Source's sole graph, `promotetest.Save` with `EntityID`).

### X-3 · major · The user-Property case in T-1 and the term-creation call in T-2 are under-specified
**Answer:** Fixed. T-1 and T-2 give the exact `properties.Create`, `propertyterms.Create` / `Update` and `properties.Update` calls with their `writes` action names. They also give the `subjectvocab.EnsureBinding` call that `observations.go` `sqlBindingExists` requires.

### X-4 · major · STEP-5 and the DoD sweep check weakly
**Answer:** Fixed. STEP-5 checks that `PersonGenderTreatment.swift` no longer mentions `read(_ fields` or `CatalogConclusionField`. The DoD sweep matches `type EventType struct` and `PersonGenderTreatment.read(detail.fields`.

### X-5 · major · Several wiring details are left to guess
**Answer:** Fixed.
- STEP-4's check is `scripts/check-pr.sh`.
- STEP-3 stops if a proto tool is missing.
- STEP-2 states that `finishTerm` never writes through the pointer it is given.

### X-6 · minor · Rules that bind omit the Swift test rule
**Answer:** Fixed. `.cursor/rules/swift-tests.mdc` added.

### X-7 · minor · T-6 to T-8 give no helper shapes
**Answer:** Fixed. T-6 gives the `facts(…)` and `header(gender:sex:)` helper signatures.

### X-8 · nit · Stale "Edges" line
**Answer:** Fixed. The Edges test references match §8.
