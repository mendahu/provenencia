# S10-50 — Promote sheet lines show the values they compared

**Status:** Blocked — on S10-28 ([#411](https://github.com/mendahu/provenencia/pull/411)) landing
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-50--promote-sheet-lines-show-the-values-they-compared)
**Brief:** — (no new chrome: lines the sheet already draws gain their values and labels)
**Board:** —
**Depends on:** S10-28 ([#411](https://github.com/mendahu/provenencia/pull/411), open), whose `ValueRef` and `withRowConflicts` this PR extends
**Blocks:** —
**Branch:** `s10-50-promote-sheet-values`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [S10-28 ship log](S10-28-support-negatives.ship.md) round 2 (UX-3, CON-4), [S10-28 plan](S10-28-support-negatives.md) D-7, D-10 and its Landed note, [`docs/matching.md`](../../../matching.md)
**Planned:** 2026-10-16 against `s10-28-support-negatives` @ `70521f33` (S10-28 not merged). Redo §4 against `main` when S10-28 lands
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Every line on Promote's evidence sheet shows the two values it compared, in the words the researcher sees elsewhere. S10-28 added lines for row-level conflicts, but a term on those lines reads as its key (`male`, or a slug such as `two-spirit` for a researcher's term) instead of its label (`Male`, `Two-Spirit`). A row that has no exhibit at all sends its lines with no values, so the sheet shows a Property, an outcome and a weight with blank value cells. Both came out of the S10-28 review (UX-3, CON-4).

**Why now.** S10-28 made row-level conflicts visible. A conflict whose values are blank, or read as internal keys, does not tell the researcher why the match got weaker.

**What stays.** Scores, outcomes, weights, and pins are unchanged. `pairExhibits` and the member picker are untouched. Exhibit lines built from Observation pairs keep their display chain. A row with exhibits gets no new line beyond S10-28's conflict rule.

## 2. User stories

**US-1.** As a researcher, when Promote shows a conflict line that S10-28 added for a term Property, I read the term's label, so the line matches the term picker and the Person page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Subject `Mary Smith` with a negative `sex_at_birth` male Observation, against a handle `Mary Smith` that is male on three Sources | The sex conflict line's incoming value is `Male` with the negative flag (the sheet reads "not Male"); the member value is `Male` |
| AC-1.2 | A `match.Value` with `Term` `two-spirit` and `TermLabel` `Two-Spirit` | `Label()` is `Two-Spirit` |
| AC-1.3 | A `match.Value` with `Term` `male` and no `TermLabel` | `Label()` is `male` (key fallback, unchanged) |
| AC-1.4 | `matching.SubjectsValues` and the candidate loaders read a Subject or handle whose sex term is `male` | The loaded `Value` has `Term` `male` and `TermLabel` `Male` |

**US-2.** As a researcher, when a Promote row has no exhibit lines, its sheet still shows each comparison with the two values it compared.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A row mapped to a handle whose only member claim is provisional (so `acceptedMembers` finds none and no exhibit is built), incoming `Mary Smith`, female, against a handle `Mary Smith`, female | `Row.Exhibits` has one own-group line per row comparison, the same Properties and outcomes as `Row.Comparisons`: name `Mary Smith` / `Mary Smith`, sex `Female` / `Female` |
| AC-2.2 | The same row | Every such line has `Pinned` false and no Observation ids |
| AC-2.3 | The same row with a birth date on both sides | The date line carries `IncomingDate` and `MemberDate`, so the client formats them for the locale |
| AC-2.4 | The FFI response for that row | `comparisons` holds those lines with `incoming_display`, `member_display`, `incoming_date`, `member_date` set; the handler has one branch for exhibits and no `r.Comparisons` fallback |
| AC-2.5 | A row S10-28 already gives a conflict line, with a date conflict | That line carries `IncomingDate` and `MemberDate` too |

## 3. Scope

| In | Covered by |
| --- | --- |
| `match.Value.TermLabel` *(new)*; `Label()` prefers it | STEP-1, AC-1.2, AC-1.3 |
| `match.ValueRef.Date` *(new)*, set by `ref()` | STEP-1, AC-2.3, AC-2.5 |
| The matching loaders and Promote's `handleValues` fill `TermLabel` | STEP-2, STEP-3, AC-1.1, AC-1.4 |
| Every row's sheet lines are built in Go: S10-28's conflict rule, plus a line per comparison for a row with no exhibit | STEP-4, AC-2.1, AC-2.2, AC-2.3, AC-2.5 |
| The FFI sends `Row.Exhibits` only | STEP-5, AC-2.4 |

| Out | Home |
| --- | --- |
| Lines for non-conflict comparisons on a row that has exhibits (a Property with a row comparison and no exhibit line) | not planned: the sheet shows Observation pairs where they exist (S10-28 D-10) |
| An exhibit's name display when the name has no `Form` (`loadExhibitObservations` shows an empty string) | not planned; log it if seen |
| Rewording "not %@" for a capitalized label ("not Male") | not planned |
| UX-2 (a minority exact match lowers the score while its exhibit line stays pinned agree) | S10-28 follow-up list |
| Other S10-28 follow-ups (MNT-1 rest, MNT-3, MNT-4, PRF-1, CON-5) | S10-28 follow-up list |
| Any Swift change | not planned: the sheet already renders `incomingDisplay`, `memberDisplay`, and the dates for any line (D-6) |

## 4. Current state

Verified on `s10-28-support-negatives` @ `70521f33`.

| Area | Where | Today |
| --- | --- | --- |
| Value label | `core/match/pairs.go` `Value.Label` (line 97) | A term returns `v.Term`, the key. `match.Value` (`core/match/match.go` line 34) has no label field |
| Value ref | `core/match/pairs.go` `ValueRef` (line 19), `Value.ref` (line 118) | `Label`, `ObservationID`, `Negative`. No date |
| Tie-break | `core/match/pairs.go` line 228 | Orders equal keys by `Label()` ascending (S10-28 D-4) |
| Subject loader | `core/database/matching/matching.go` `sqlSubjectsValues` (line 69), `SubjectsValues` (line 288, value built at line 318) | Selects `t.key` and `t.neutral`, not `t.label` |
| Candidate loader | `core/database/matching/matching.go` `sqlEntityValues` (line 77; `sqlEntityValuesOfType`, `sqlEntityValuesOfOne`, and the `EntitiesValues` batch (line 257) all use it; `CandidatesOfType` line 243), `scanEntityValues` (line 373, value at line 396) | Same: key and neutral, no label |
| Promote handle values | `core/database/promotealign/canon.go` `handleValues` (line 245), `neutralTerms` (line 287) | `neutralTerms` reads `g.Vocabulary().Terms` into a neutral-by-id map; `graphcache.VocabTerm` (`core/database/graphcache/vocab.go` line 23) already has `Label`. `handleValues` sets `Term` and `TermID`, no label |
| Promote subject values | `core/database/promotealign/layer.go` line 48 | Calls `matching.SubjectsValues` |
| Row comparisons | `core/graphalign/types.go` `Comparison` (line 129); `core/graphalign/align.go` `comparisonsFrom` (line 882) | Carries `Probe` and `Candidate` `match.ValueRef` since S10-28 |
| Exhibit | `core/graphalign/types.go` `Exhibit` (line 149) | Has `IncomingDate`, `MemberDate` (`*datevalues.Value`) and `IncomingNegative` |
| Conflict lines | `core/database/promotealign/exhibits.go` `withRowConflicts` (line 175), called at line 166 inside `attachExhibits` | Adds a line per unshown conflict; displays from `Probe.Label` / `Candidate.Label`; no dates. Runs only for rows in `plans`, and not at all when `attachExhibits` returns early (`want` empty, line 141) |
| Exhibit display | `core/database/promotealign/exhibits.go` `loadExhibitObservations` (line 375) | Joins `t.label` and uses it as the term display |
| Attach call | `core/database/promotealign/promotealign.go` line 85 | `attachExhibits(q, c.Graph(), layer, &prop, cfg, stats)` |
| FFI | `api/ffi/handlers/promote_align.go` lines 126–161 | Sends exhibits when `len(r.Exhibits) > 0`, else `r.Comparisons` with outcome, value type, pinned, weight, and no values |
| Swift | `macos/App/Features/Promote/PromoteView.swift` `comparisonLine`, `incomingText`; `ComparisonValueDisplay.swift`; `PromoteFlow.swift` `draftedPins`, `draftedArgument`, `batchRow` | Draws any line's displays and dates. The pin box is disabled for an empty `incomingObservationID`; `batchRow` skips a pin without both ids. `draftedArgument` joins `incomingDisplay` of `agree` lines that have one |
| Seed labels | `core/database/subjectvocab/registry.go` `seedTerms` (line 208) | `sex_at_birth` `female` → `Female`, `male` → `Male` |
| Tests | `core/match/pairs_test.go` `TestValueLabel` (line 179); `core/database/matching/matching_test.go` `TestSubjectsValuesNegatives` (line 501), `termed`, `negated`; `core/database/promotealign/promotealign_test.go` `TestProposeJudgesAgainstTheMajority` (line 1088, asserts `IncomingDisplay` `male`), `conflictExhibit`, `comparisonOutcome`; `api/ffi/handlers/promote_align_test.go` `TestProposeNegativeConflictLine` (line 85) | Green at the base |

**Precedent.** S10-28's `withRowConflicts` (commit `86726467`) and `TestProposeNegativeConflictLine` (commit `b218653e`): a Go-side line built from the row's `ValueRef`s, and an FFI test that it reaches the response. Copy both. `neutralTerms` is the precedent for reading term data from the graph vocabulary.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `match.Value.TermLabel string` *(new)* holds the term's catalog label. `Label()` returns `TermLabel` when `Term` is set and `TermLabel` is not empty, else `Term`. Comparers keep comparing by `Term`. The tie-break at `pairs.go` line 228 keeps calling `Label()`, so equal-key terms now order by label | One label for every reader of a `ValueRef`: Promote lines, `Rank` reasons, and the reason sentence S10-28 named as a next add. A label is data the loaders already join | Resolving labels in `promotealign` at line-build time — `Rank` reasons would still name keys, and a second lookup would sit beside the loaders |
| D-2 | The loaders fill it. `sqlSubjectsValues` and `sqlEntityValues` select `COALESCE(t.label, '')` after `t.key`; `SubjectsValues` and `scanEntityValues` set `TermLabel`. `handleValues` reads it from the graph vocabulary: `neutralTerms` becomes `termsByID(g) (map[string]graphcache.VocabTerm, error)` *(new name)*, and `handleValues` takes that map and sets `Neutral` and `TermLabel` from the entry for `row.TermID` | The loaders are where term rows enter (S10-28 D-8). The vocabulary is already in memory; no query is added | A `TermLabel` on `graphcache.Value` — one string per cached row of every handle, for data the vocabulary holds once |
| D-3 | `match.ValueRef.Date *datevalues.Value` *(new)*. `ref()` sets it to `v.Date` (the pointer, no copy) | The sheet formats dates from the structured value; a line with only `CompactDisplay` text reads differently from its neighbors. The pointer keeps `BenchmarkPairValues/plain` at 0 allocs/op | Formatting in Go — the client owns locale formatting (`ComparisonValueDisplay`) |
| D-4 | Row lines move out of `attachExhibits`. `withRowConflicts` becomes `withRowLines(exhibits []graphalign.Exhibit, comparisons []graphalign.Comparison) []graphalign.Exhibit` *(renamed)*: when `exhibits` is empty, every comparison becomes a line; otherwise S10-28's rule (a conflict with no own-group conflict line on its Property). A line is built by `refLine(c graphalign.Comparison) graphalign.Exhibit` *(new)*: Property, outcome, value type, weight, `Pinned` false, no Observation ids, `IncomingDisplay`/`MemberDisplay` from the refs' `Label`, `IncomingDate`/`MemberDate` from the refs' `Date`, `IncomingNegative` from `Probe.Negative`. `attachExhibits` no longer calls it. `promotealign.go` runs `r.Exhibits = withRowLines(r.Exhibits, r.Comparisons)` for every row right after `attachExhibits` returns | One place decides what the sheet shows, for every row, including the early-return path. The FFI stops having to know | Filling values in the FFI's fallback branch — two places that build lines, and the FFI would need the refs on its own |
| D-5 | A line built from refs is never pinned: `Pinned` false, and the box stays disabled (empty `incomingObservationID`). Before this PR the fallback sent `Pinned` true for an `agree` comparison, which drew a checked, disabled box that `batchRow` never committed | A pin needs an Observation pair; a ref line has none. A checked box that does nothing is wrong | Sending the probe's `ObservationID` as the incoming id — a pin also needs a member Observation, which a cached cluster does not have |
| D-6 | No Swift change. A no-exhibit row's `agree` lines now carry `incomingDisplay`, so `draftedArgument` includes them in the drafted argument. That is intended: the argument lists the incoming Subject's agreeing values, and these are its own values | The rule already reads that way; the values were missing | Filtering ref lines out of the argument — the argument would skip agreeing values only because the row had no exhibit |
| D-7 | The FFI's `else` branch (`promote_align.go` lines 151–161) is deleted. The handler maps `r.Exhibits` only | D-4 makes it dead | Keeping it as a guard — it would never run |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `match.Value.TermLabel` *(new)* | field | `match.Value.Term`, `Neutral` | Extends the value the loaders already build; `Term` stays the comparison key |
| `match.ValueRef.Date` *(new)* | field | `graphalign.Exhibit.IncomingDate` | Extends S10-28's ref; the exhibit fields are the destination |
| `termsByID` *(renamed from `neutralTerms`)* | func | `neutralTerms`, `loadKinship` (`terms.go`) | Same vocabulary read; returns the whole term so callers read `Neutral` and `Label` |
| `withRowLines` *(renamed from `withRowConflicts`)* | func | `withRowConflicts` | Extended, not copied |
| `refLine` *(new)* | func | the literal in `withRowConflicts` | Pulled out so both rules build the same line |

**Builds on.** S10-28's `ValueRef` and row-line rule; the graph vocabulary (`graphcache.Vocabulary`); `ComparisonValueDisplay` on the client.
**Deletes.** The FFI fallback branch (`promote_align.go` lines 151–161). `neutralTerms` (renamed). The S10-28 `Out` line "Unifying `Value.Label` with the exhibit display chain" is answered for terms: both now read `property_terms.label`.
**Next adds.** A Promote reason sentence reads `Reason.Probe.Label` and gets term labels for free. Lines for non-conflict comparisons on rows with exhibits, if wanted later, are one more case in `withRowLines`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/match` | `Value.TermLabel` *(new)*; `Label()` per D-1; `ValueRef.Date` *(new)*; `ref()` sets it |
| `core/database/matching` | `sqlSubjectsValues`, `sqlEntityValues` select `COALESCE(t.label, '')` after `t.key`; `SubjectsValues` and `scanEntityValues` scan it into `TermLabel` |
| `core/database/promotealign` | `canon.go`: `neutralTerms` → `termsByID`; `handleValues(props, terms, n, namePattern)` sets `Neutral` and `TermLabel`; every caller passes the map. `exhibits.go`: `withRowConflicts` → `withRowLines` + `refLine` (D-4); `attachExhibits` drops the call. `promotealign.go`: the loop after `attachExhibits` |

Errors: none new.

### FFI

`api/ffi/handlers/promote_align.go`: delete the `else` branch (D-7). No proto change: `incoming_display`, `member_display`, `incoming_date`, `member_date`, and `incoming_negative` exist.

### macOS

None — no Swift change (D-6).

### Strings

None — no copy change.

## 7. Steps

**STEP-1 — Term labels and dates on values and refs.**
Files: `core/match/match.go`, `core/match/pairs.go`, `core/match/pairs_test.go`.
Do: D-1 and D-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match` passes (T-1, T-2); `go test -tags fts5 -bench PairValues -benchmem ./core/match` shows `plain` at 0 allocs/op.

**STEP-2 — The matching loaders read the label.**
Files: `core/database/matching/matching.go`, `core/database/matching/matching_test.go`.
Do: D-2, matching half.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/matching` passes (T-3).

**STEP-3 — Promote's handle values read the label.**
Files: `core/database/promotealign/canon.go`.
Do: D-2, `termsByID` half.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign` passes, with `TestProposeJudgesAgainstTheMajority` updated to expect `Male` (T-4).

**STEP-4 — Every row's lines are built in Go.**
Files: `core/database/promotealign/exhibits.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-4, D-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign ./core/graphalign` passes (T-5, T-6).

**STEP-5 — The FFI sends exhibits only.**
Files: `api/ffi/handlers/promote_align.go`, `api/ffi/handlers/promote_align_test.go`. Skill: `add-ffi-handler`.
Do: D-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers -run Promote` passes (T-7); `scripts/check-pr.sh` PASS.

**STEP-6 — Docs.**
Files: `docs/matching.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done. In `docs/matching.md`, the S10-28 section says a ref names a term by its label and carries a date, and that every Promote row's lines are built in `promotealign`.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-50-promote-sheet-values.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/match/pairs_test.go` `TestValueLabel` (extend) | `Value{Term: "two-spirit", TermLabel: "Two-Spirit"}` → `Two-Spirit`; `term("male")` → `male` | AC-1.2, AC-1.3, D-1 |
| T-2 | Go | `core/match/pairs_test.go` `TestPairValuesRefDate` *(new)* | Probe and candidate `date(1890, nil, nil)` (helper in `dates_test.go`), comparer `ComparerFor(properties.ValueTypeDate)` → `PairResult.Probe.Date` and `.Candidate.Date` point at the input values' dates | D-3 |
| T-3 | Go | `core/database/matching/matching_test.go` `TestValuesCarryTermLabel` *(new)* | Subject with `termed("sex_at_birth", "male")`, read by `SubjectsValues` → `Term` `male`, `TermLabel` `Male`. A handle with sex male, read by `CandidatesOfType` and by `EntitiesValues` → the same on both | AC-1.4, D-2 |
| T-4 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeJudgesAgainstTheMajority` (update the negative case) | The sex conflict line: `IncomingDisplay` `Male`, `IncomingNegative` true, `MemberDisplay` `Male` | AC-1.1, D-2 |
| T-5 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeRowWithoutExhibitsShowsValues` *(new)* | Promote `Mary Smith` (female, born 1890); `UPDATE identity_claims SET status = 'provisional'` for that handle; `f.c.Graph().Drop()`; incoming `Mary Smith` (female, born 1890) fixed to the handle → `row.Exhibits` has one own-group line per `row.Comparisons` entry, same Property and outcome; name `Mary Smith`/`Mary Smith`; sex `Female`/`Female`; the birth date line has `IncomingDate` and `MemberDate` set; every line `Pinned` false, empty Observation ids | AC-2.1, AC-2.2, AC-2.3, D-4, D-5 |
| T-6 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeDateConflictLineCarriesDates` *(new)* | A handle with birth 1890 on three Sources and 1850 on one (by SQL, like the stray-name case); incoming born 1850 → the S10-28 conflict line for the birth date has `IncomingDate` and `MemberDate` | AC-2.5, D-3 |
| T-7 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeRowWithoutExhibitsSendsValues` *(new)* | The T-5 fixture through the handler → each response comparison has `incoming_display` and `member_display`; the date line has `incoming_date` and `member_date` | AC-2.4, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (`TestProposeJudgesAgainstTheMajority` expects `male`), `core/match` tests that compare `Reason` labels for terms, and `macos/ProvenenciaTests` `PromoteFlowTests` that build lines from a fake.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/matching.md (term labels on refs, row lines built in promotealign)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Add a negative Observation "sex at birth: male" to a Subject whose handle is male on most Sources, and open Promote's evidence sheet for that row. The conflict line reads "not Male" against "Male", not "male".
2. Make a handle's only member claim provisional (or use a handle whose members have no Observation that pairs with the incoming Subject's), and open the sheet for a Subject fixed to it. Every line shows both values; the date line uses your locale's date format; no pin box is checked.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Equal-key term pairs now break ties by label, so a reported pair can change | Similarity is equal by definition of an equal key; only the named pair moves. T-1 and the existing `TestPairValuesOrderIndependent` cover determinism |
| A no-exhibit row's drafted argument now lists its agreeing values | Intended (D-6) |
| A no-exhibit `agree` line's box is no longer drawn checked | Intended (D-5): it was never committed |

**Stop and report if:**

- `withRowConflicts`, `ValueRef`, or the FFI branch on `main` after S10-28 lands no longer match §4.
- `BenchmarkPairValues/plain` reports more than 0 allocs/op after STEP-1.
- An existing test outside `TestProposeJudgesAgainstTheMajority` needs a changed expected label to pass; name it.
- T-5's fixture (a provisional-only claim) still yields an exhibit, or no row comparisons — report what `acceptedMembers` and `Row.Comparisons` hold.
- A `PromoteFlowTests` case fails because a line's pin state changed.

## 12. Open questions

None.

## 13. PR

**Title:** S10-50 — Promote sheet lines show the values they compared
**Body:** Every Promote sheet line shows the two values it compared, with term labels; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-50-promote-sheet-values.md`.
