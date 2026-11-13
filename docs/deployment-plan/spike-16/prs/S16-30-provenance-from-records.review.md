# Review — S16-30 Provenance records compose from cached records

**Verdict:** round 1: Fits with changes / Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None. Asked before review (2026-11-13): the handler reads the subject type from the vocabulary; the picture paths stay one SQL read; bar 11 and the scope Check reworded to match.

## Writer's own calls
- "The second pass of `PersonDerivedRows`" read as composing every field of every owner and keeping one: D-7 and D-8 name each owner's Property before composing (`only` replaces `subjectValued`).
- Citation words live on each `EvidenceObservation` (D-1), as the working rule words it ("each Observation's Citation words"), not once per Citation (M-7).
- The detail budget times `ForEntity` only, on `households(f, 80)` (X-2); the life rows are held by query counts.

## Round 1

### M-1 · blocker · The reader builds its own Observation index and toponym memo; the cache layer should own them
**Answer:** Fixed — D-4: `newGraphRecords` builds `byObservation`, `GraphRecords.Observation(id)` *(new)*; D-3's reader is lookups and a copy. The toponym rule moves to `observations.FirstTextOf` (M-4); the reader keeps a per-read memo of value Subjects' toponyms (one map), since a stored toponym map would be a second answer per Subject the records do not need.

### M-2 · blocker · D-5 adds a second value decoder beside `attachValues`
**Answer:** Fixed — D-12: `valueOf` *(new)*, extracted from `attachValues`; both callers use it. `evidencelayer.Decode` declined: it returns the matching layer's value without term labels.

### M-3 · major · D-12's premise is false: the Observation carries `DateID` and `NameID`
**Answer:** Fixed — `valueOf` sets the ids from `DateID` / `NameID`; the old D-12, its Risk row and its Stop condition are gone.

### M-4 · major · A second "Observations → TextRecord → FirstText" builder beside `grapheventtitles`
**Answer:** Fixed — D-2: `observations.FirstTextOf` *(new)*; `grapheventtitles.FromStored` and the detail read call it; `graphfill` gains no `observations` import.

### M-5 · minor · `SubjectLabels`' "the one label rule" comment
**Answer:** Fixed — STEP-2 updates the comment.

### M-6 · minor · Stale `sqlOutcomes` names in the map and the design doc
**Answer:** Fixed — map row and scope entry now name `sqlRecords` (done in this planning commit); STEP-7 rewrites `cache-architecture.md`'s "Provenance composed"; the DoD sweep lists `sqlOutcomes`. `spike-15/provenance-loading.md` describes the history S15-28 starts from; left to S15-28's planner.

### M-7 · minor · Citation words repeat per Observation
**Answer:** Declined — the working rule ("each Observation's Citation words join with S16-30", user 2026-11-13) and S16-13 D-4 name a per-Observation shape; a per-Citation map needs the same in the loaded-graph fill. The repeat is now in §6 Scale's memory line and measured in STEP-1 and STEP-6 against the 35 MB budget.

### M-8 · nit · The budget test lives in `persondetail`
Logged.

### X-1 · major · STEP-4's `rg` check hits test files
**Answer:** Fixed — the check globs out `*_test.go` and checks `export_test.go` on its own.

### X-2 · major · `households` gives `PersonDerivedRows` no owners
**Answer:** Fixed — AC-1.5 and D-13 time `ForEntity` only; the life rows are held by T-5's query counts.

### X-3 · major · STEP-3 leaves `only` undecided for existing test callers
**Answer:** Fixed — STEP-3: `TestFieldsOf`'s cases pass nil; `CountFieldsOfQueries` takes `only` built from `owners` until STEP-4.

### X-4 · major · Where `recordReader` is wired into `fieldsOf`
**Answer:** Fixed — D-3 says how rows are collected, where each result goes, where a Subject's graph comes from, and the drop-then-error order; D-5 takes `GraphID` from the records.

### X-5 · major · No fixture builds the Citation T-1 and T-4 need
**Answer:** Fixed — T-1 and T-4 name the create and update calls and the two fill paths.

### X-6 · minor · A base that misses its budget
**Answer:** Fixed — STEP-1 and §11 stop and report.

### X-7 · minor · Warm SELECTs from paths this PR does not touch
**Answer:** Fixed — §11 Stop condition.

### X-8 · minor · Small choices
**Answer:** Fixed — D-1 states the join and the flag are never NULL (`000026.sql`). The rest needs no change: a cross-graph value Subject is a Stop condition already; an owner with no outcome for its Property has no field, as today.
