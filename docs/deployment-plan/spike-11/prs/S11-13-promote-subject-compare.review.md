# Review — S11-13 Migrate Promote scoring onto subject compare

Plan: [`S11-13-promote-subject-compare.md`](S11-13-promote-subject-compare.md)

## Summary

- **User calls before review (2026-10-18):** candidates need entity compare evidence at least the weak band (D-4); the derived lifespan stays off the sheet and the row reason until S11-20 (D-6). Both are in the map: S11-13's and S11-20's scope entries.
- **Round 1:** maintainability "Fits with changes" (1 blocker, 4 majors), executability "Ready with changes" (1 blocker, 4 majors). One blocker fixed, one declined with evidence.
- **Round 2:** maintainability "Fits with changes" (M-10 blocker, fixed); executability "Ready with changes" (X-9 blocker, fixed; X-10 – X-12 fixed). Both accepted the round-1 answers, including the declines of M-5 and X-1.

## Round 1 — maintainability

### M-1 · blocker · Three decoders for one value row, and a second "displayed" rule
**Answer:** Fixed. D-8 adds one `decode` for Source and cache rows; D-10's `reconciled` is `decode` plus support and reason. D-9 takes life dates from values already decoded: the event Subject's `layerValues` candidates, and the event node's `reconciled` rows filtered by `Reason.Displayed()`. The cross-package decoder copies (`conclusiondetails`, `conclusionheaders`, `api/ffi/handlers/source_read.go`, `searchindex`) are left alone; they are not this PR's files.

### M-2 · major · `Comparison` copies `PropertyResult` and drops `Cause`
**Answer:** Fixed. D-5: `Comparison` embeds `subjectcompare.PropertyResult` and adds `ValueType` and `Derived`. AC-1.5, AC-4.1, STEP-4 and the §6 sketch read `Key`, `Points` and `Result`.

### M-3 · major · "Derived" guessed from a missing catalog row
**Answer:** Fixed. D-12 adds `subjectcompare.Rules.Derived` (a `Lifespans` lead or `Span`); D-5 and D-14 use it. D-14 now counts only properties the subject section reads (`Lead` ok, not derived), which also keeps a remark-only person on Skip / `empty` as today. `TestDerived` is in T-1.

### M-4 · major · `datedEvent` near-copies `birthEvent`
**Answer:** Fixed. STEP-6 renames `birthEvent` to `datedEvent(f, src, art, eventType, year, polarity)`; callers pass `"birth"`.

### M-5 · major · Hand-written meta helpers despite the skill
**Answer:** Declined. The skill line ("Tests build values with `propertycomparetest` and metas with `subjectcomparetest.Metas`") sits in its Layer 3 bullet (`.agents/skills/add-match-rule/SKILL.md`), and `core/graphalign` is not in `layerPackages` (`core/matchregistry/guard_test.go`). The walk goldens compare names as text on purpose (`personMetas` uses `ValueTypeText`); moving them to `subjectcomparetest.Metas()` turns every text name into a name value and moves their numbers for no behavior. New tests use `subjectcomparetest.Metas()` (D-16).

### M-6 · minor · `Rules.Lead` claimed for S11-14
**Answer:** Fixed. Next adds says `pairExhibits` is its only caller and S11-19 / S11-25 decide whether it stays.

### M-7 · minor · `graphcache.SourceBridge` is dead
**Answer:** Fixed. STEP-3 deletes `SourceBridge` and `SourceGraph.Bridges` after a grep confirms no use; listed in Deletes.

### M-8 · minor · Fake term id; second SQL read in exhibits
**Answer:** Fixed for the term id (D-11 selects `o.value_term_id`). The second read is an Out line homed in S11-19 / S11-25, which replace `pairExhibits`.

### M-9 · nit · Product origin encoded twice
**Answer:** Logged, not changed.

## Round 1 — executability

### X-1 · blocker · Plan and map edits are uncommitted
**Answer:** Declined. **Planned** names the code base the Current state was read at (`97eea0b3`), as every Spike 11 plan does (S11-12 "against `main` @ `b799b67`"). The plan, this log and the map edits land on `main` as one docs commit before `ship-pr` runs; that is the plan-pr wiring step, not an implementer step.

### X-2 · major · AC-5.2's escaped pipes
**Answer:** Fixed. The greps moved to a fenced `bash` block in STEP-5's Check; AC-5.2 and AC-5.3 point there.

### X-3 · major · §1 overpromised the alternatives gate
**Answer:** Fixed. §1 and the Risks row say only seeding and the fallback are gated; a walk-nominated handle is still recorded and listed at any score (S11-16 decides). AC-1.3 asserts D is in `Alternatives` with Assessment none.

### X-4 · major · Seed loop pre-check has no replacement
**Answer:** Fixed. D-14: no pre-check; the loop calls `Candidates` for every unfixed Subject of a weighted kind, with the reason it is safe.

### X-5 · major · "Update expected values per D-16" is open-ended
**Answer:** Fixed. D-16 lists the four expected changes (comparison lists, scores from `cfg.Subject.Weights`, seeding / fallback candidates under 2, exhibit weights); any other change is a stop condition.

### X-6 · minor · Open choices in D-9, D-10, D-11, D-13, D-3, D-5
**Answer:** Fixed. D-9: `person` = `e.From`, `event` = `e.To`, dates in any order; role `"subject"` in STEP-6. D-10: kind from `typeKey`, `loadCanon` keeps nodes while walking and profiles all of them after. D-11: scan targets and the `subject` skip. STEP-4: the label closure is built per row. D-3: `ConflictPenalty` and `accept` comments, `scoredCand.result`, `evaluate`'s signature. D-5: the empty-list trigger.

### X-7 · minor · Doc anchors and stale mentions
**Answer:** Fixed: lines 166–174 with the `Compatible` paragraph kept; DOC-9 keeps the trailing sentence; `promote-find-another-record.md` lines 7 and 11 corrected; STEP-7's grep covers the edited idea files. Declined for `matching-model.md` (glossary and Removed knobs record the old names on purpose) and `docs/ideas/matching-test-suite.md` (its knob-inventory item records where knobs lived when the idea was written; S11-01 shipped it).

### X-8 · nit · Smaller points
**Answer:** Source graph row says "among its fields"; `birthEvent` is renamed rather than copied (M-4). Logged otherwise.

## Round 2 — maintainability

### M-10 · blocker · Two life-event builders, one on a second "displayed" rule
**Answer:** Fixed. D-9 builds event profiles first (`ForSubject` on Source events, the profiler's `Build` on neighbor events) and reads a person's life from them with one `lifeEvent(event, role)`; `participation` drops `eventType`. D-10 builds neighbor events before persons. `keptTerm` stays for `canonSignature` only.

## Round 2 — executability

### X-9 · blocker · AC-5.3's grep hits `core/matchregistry/value.go` comments
**Answer:** Fixed. STEP-1 rewrites the two comments (exact text given).

### X-10 · major · The `SourceBridge` grep matches live code
**Answer:** Fixed. STEP-3 deletes the declarations and lets `go build ./...` prove them unused, then `rg -n 'SourceBridge'` prints nothing.

### X-11 · minor · How `Points` fits `Compare`'s switch
**Answer:** Fixed. STEP-1: the switch keeps `counted` and `score`; `pr.Points = Points(res.Outcome, sim, w)` after it.

### X-12 · minor · Small gaps
**Answer:** Fixed: D-14 names the function-to-method change and its call site; D-8 names the properties-by-id map `loadLayer` builds; §2 conventions map bare names to `vocabkey` constants and say `withLife` sets `lifespan` and `lifespan_span` on a copy. STEP-3's check stays a compile check by design (behavior is checked in STEP-4 and STEP-6); D-17's commit shape is stated.

## Verdict

**Fits / Ready, 2 rounds (findings fixed).** No blocker open, nothing To user. Declined with evidence: M-5, X-1, part of X-7.
