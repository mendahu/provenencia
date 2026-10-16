# S10-50 — Promote sheet lines show the values they compared

**Status:** Blocked — on S10-28 ([#411](https://github.com/mendahu/provenencia/pull/411)) merging into `main`
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-50--promote-sheet-lines-show-the-values-they-compared)
**Brief:** — (no new chrome: lines the sheet already draws gain their values and labels)
**Board:** —
**Depends on:** S10-28 ([#411](https://github.com/mendahu/provenencia/pull/411)), whose `ValueRef` and `withRowConflicts` this PR extends
**Blocks:** —
**Branch:** `s10-50-promote-sheet-values`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [S10-28 ship log](S10-28-support-negatives.ship.md) round 2 (UX-3, CON-4), [S10-28 plan](S10-28-support-negatives.md) D-7, D-10 and its Landed note, [`docs/matching.md`](../../../matching.md) "Support weights and negatives"
**Planned:** 2026-10-16 against `s10-28-support-negatives` @ `70521f33`, the head of open PR #411. That branch already records S10-28 as landed (its implementer wrote the docs ahead of the merge). Before you start: `gh pr view 411 --json state` must say `MERGED`; then branch from `main` and recheck every §4 row there
**Review:** [`S10-50-promote-sheet-values.review.md`](S10-50-promote-sheet-values.review.md) — 2 rounds: Fits with changes / Not ready, then Fits / Ready with changes; every finding answered; no open blocker

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

Every line on Promote's evidence sheet shows the two values it compared, in the words the researcher sees elsewhere. S10-28 added lines for row-level conflicts, but a term on those lines reads as its key (`male`, or a slug such as `two-spirit` for a researcher's term) instead of its label (`Male`, `Two-Spirit`). A row that has no exhibit at all sends its lines with no values, so the sheet shows a Property, an outcome and a weight beside blank value cells. Both came out of the S10-28 review (UX-3, CON-4). The fix gives the sheet one way to turn a value into text, `match.Value.Label`, for every line.

**Why now.** S10-28 made row-level conflicts visible. A conflict whose values are blank, or read as internal keys, does not tell the researcher why the match got weaker.

**What stays.** Scores, outcomes, weights, and which lines are pinned are unchanged. `pairExhibits` keeps its member picker. A row with exhibits gets no new line beyond S10-28's conflict rule. Exhibit lines keep the text they show today, except that a name with no `Form` now shows its parts instead of nothing.

## 2. User stories

**US-1.** As a researcher, when a Promote line names a term, I read the term's label, so the line matches the term picker and the Person page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Subject `Mary Smith` with a negative `sex_at_birth` male Observation, against a handle `Mary Smith` that is male on three Sources | The sex conflict line's incoming value is `Male` with the negative flag (the sheet reads "not Male"); the member value is `Male` |
| AC-1.2 | A `match.Value` with `Term` `two-spirit` and `TermLabel` `Two-Spirit` | `Label()` is `Two-Spirit` |
| AC-1.3 | A `match.Value` with `Term` `male` and no `TermLabel` | `Label()` is `male` (key fallback, unchanged) |
| AC-1.4 | `matching.SubjectsValues`, `matching.CandidatesOfType`, and `matching.EntitiesValues` read a Subject or handle whose sex term is `male` | The loaded `Value` has `Term` `male` and `TermLabel` `Male` |
| AC-1.5 | An exhibit line whose name Observation has parts and no `Form` | Its display is the parts joined by spaces, not empty |

**US-2.** As a researcher, when a Promote row has no exhibit lines, its sheet still shows each comparison with the two values it compared.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Person row fixed to a handle whose only member claim is provisional (so `acceptedMembers` finds none and no exhibit is built), incoming `Mary Smith`, female, against the handle `Mary Smith`, female | `Row.Exhibits` has one own-group line per row comparison whose outcome is not `unknown`, with the same Property and outcome: name `Mary Smith` / `Mary Smith`, sex `Female` / `Female`. `gender` (neither side has one) is unknown and has no line |
| AC-2.2 | The same row | Every such line has `Pinned` false and no Observation ids |
| AC-2.3 | A birth Event row (event type birth, `date` 1890) fixed to a birth Event handle with date 1890 whose only member claim is provisional | The `date` line carries `IncomingDate` and `MemberDate`, so the client formats them for the locale |
| AC-2.4 | The FFI response for the AC-2.1 row | Its `comparisons` have `incoming_display` and `member_display` set; the handler maps `Row.Exhibits` only |
| AC-2.5 | A birth Event Subject with event type birth and a negative `date` 1890 Observation, fixed to a birth Event handle dated 1890 (accepted member) | The S10-28 conflict line for `date` carries `IncomingNegative` true, `IncomingDate`, and `MemberDate` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `match.Value.TermLabel` *(new)*; `Label()` prefers it; the tie-break uses a stable key | STEP-1, AC-1.2, AC-1.3 |
| `match.ValueRef.Date` *(new)*, set by `ref()` | STEP-1, AC-2.3, AC-2.5 |
| `graphcache.Vocabulary.TermsByID` *(new)*; its three hand-built copies move onto it | STEP-2 |
| Every loader that builds a term `match.Value` fills `TermLabel` | STEP-3, STEP-4, AC-1.1, AC-1.4 |
| Exhibit lines take their text from `Value.Label()` | STEP-4, AC-1.5 |
| Every row's sheet lines are built in `attachExhibits`: S10-28's conflict rule, plus a line per comparison for a row with no exhibit | STEP-5, AC-2.1, AC-2.2, AC-2.3, AC-2.5 |
| The FFI sends `Row.Exhibits` only; `graphalign.Comparison.Pinned` is deleted | STEP-6, AC-2.4 |

| Out | Home |
| --- | --- |
| Lines for non-conflict comparisons on a row that has exhibits (a Property with a row comparison and no exhibit line) | not planned: the sheet shows Observation pairs where they exist (S10-28 D-10) |
| Rewording "not %@" for a capitalized label ("not Male") | not planned |
| UX-2 (a minority exact match lowers the score while its exhibit line stays pinned agree) | S10-28 follow-up list |
| Other S10-28 follow-ups (MNT-1 rest, MNT-3, MNT-4, PRF-1, CON-5) | S10-28 follow-up list |
| Any Swift change | not planned: the sheet already renders `incomingDisplay`, `memberDisplay`, and the dates for any line (D-7) |

## 4. Current state

Verified on `s10-28-support-negatives` @ `70521f33`. Line numbers are approximate.

| Area | Where | Today |
| --- | --- | --- |
| Value label | `core/match/pairs.go` `Value.Label` (line 97) | A term returns `v.Term`, the key. `match.Value` (`core/match/match.go` line 34) has no label field. A name with no `Form` joins its parts |
| Value ref | `core/match/pairs.go` `ValueRef` (line 19), `Value.ref` (line 118) | `Label`, `ObservationID`, `Negative`. No date |
| Tie-break | `core/match/pairs.go` `order` (line 228) | Orders equal keys by `strings.Compare(vs[i].Label(), vs[j].Label())` (S10-28 D-4) |
| Pair entry point | `core/match/pairs.go` `pairValues(cmp Comparer, card string, probe, candidate []Value) PairResult` | Unchanged by this PR |
| Subject loader | `core/database/matching/matching.go` `sqlSubjectsValues` (line 69), `SubjectsValues` (line 288; scan near line 312, value at line 318) | Selects `t.key, COALESCE(t.neutral, 0)`, not `t.label` |
| Candidate loader | `core/database/matching/matching.go` `sqlEntityValues` (line 77; used by `sqlEntityValuesOfType`, `sqlEntityValuesOfOne`, and the `EntitiesValues` batch at line 257; `CandidatesOfType` at line 243), `scanEntityValues` (line 373; scan near line 392, value at line 396) | Same: `t.key, COALESCE(t.neutral, 0)`, no label |
| Exhibit loader | `core/database/promotealign/exhibits.go` `loadExhibitObservations` (line 375), `exhibitObs.display` (line 26), `pending.termLabel` | Selects `t.key, t.label`. Builds `display` with its own chain: `CompactDisplay`, then `n.Form`, then text, then `termLabel`, then the integer. A name with no `Form` displays empty. `pairExhibits` copies `display` to `IncomingDisplay`/`MemberDisplay` (lines 265, 269) |
| Promote handle values | `core/database/promotealign/canon.go` `handleValues` (line 245), `neutralTerms` (line 287); `loadCanon` line 26 (`neutral, err := neutralTerms(g)`) and call sites at lines 57, 100, 176 | `handleValues` sets `Term`, `TermID`, and `Neutral` from a neutral-by-id map. No label |
| Vocabulary maps | `core/database/graphcache/vocab.go` `Vocabulary` (line 6), `VocabTerm` (line 23, has `Label`); `core/database/conclusiondetails/conclusiondetails.go` `termLabels` (line 331), `termName`; `api/ffi/handlers/source_read.go` `loadTermInfo` (line 44, returns an empty map when `c` or `c.Graph()` is nil), `termInfo` (line 24), `listedFromSource` (line 88), caller at line 66; `api/ffi/handlers/event_titles.go` line 62 (calls `loadTermInfo`, copies it into a `map[string]struct{ Key, Label string }` for `sourceeventtitles.FromStored`) | Three hand-built id → term maps over `Vocabulary().Terms` |
| Row comparisons | `core/graphalign/types.go` `Comparison` (line 129, `Pinned` at line 133, doc comment "suitable for drafting pins later (S9-43)"); `core/graphalign/align.go` `comparisonsFrom` (line 882, sets `Pinned` at line 890) | Carries `Probe` and `Candidate` since S10-28. Only a row with a handle target gets comparisons (`proposal`, line 646); a New or Skip row has none |
| Exhibit | `core/graphalign/types.go` `Exhibit` (line 149) | Has `IncomingDate`, `MemberDate` (`*datevalues.Value`), `IncomingNegative` |
| Conflict lines | `core/database/promotealign/exhibits.go` `withRowConflicts` (line 175), called at line 166 inside `attachExhibits` (line 37); `cloneDate` | Adds a line per unshown conflict; displays from `Probe.Label` / `Candidate.Label`; no dates. Runs only for rows in `plans`. `attachExhibits` returns at line 141 when `want` is empty, which happens only when no row has a handle target |
| Attach call | `core/database/promotealign/promotealign.go` line 85 | `attachExhibits` is called once, from `Propose` |
| FFI | `api/ffi/handlers/promote_align.go` lines 126–161 | `if len(r.Exhibits) > 0 { for … } else { for … r.Comparisons … }`. The `else` branch sends outcome, value type, `Pinned`, weight, and no values. It is the only production reader of `Comparison.Pinned` |
| Swift | `macos/App/Features/Promote/PromoteView.swift` `comparisonLine`, `incomingText`; `ComparisonValueDisplay.swift`; `PromoteFlow.swift` `draftedPins`, `draftedArgument`, `batchRow` | Draws any line's displays and dates. The pin box is disabled for an empty `incomingObservationID`; `batchRow` skips a pin without both ids. `draftedArgument` joins `incomingDisplay` of `agree` lines that have one |
| Seed labels | `core/database/subjectvocab/registry.go` `seedTerms` (line 208) | `sex_at_birth` `female` → `Female`, `male` → `Male`; `event_type` `birth` → `Birth` |
| Tests that read term keys or `Comparison.Pinned` | `core/database/matching/matching_test.go` `TestForSubjectJudgesAgainstTheMajority` (lines 477, 495: `Probe.Label`/`Candidate.Label` `female`/`male`); `core/database/promotealign/promotealign_test.go` `TestProposeJudgesAgainstTheMajority` (line 1088; its negative case asserts `IncomingDisplay` and `MemberDisplay` `male` at line 1140), `TestProposeWeighsACommonNameLess` (line 529, `!c.Pinned`); `core/graphalign/align_test.go` line 256 (`c.Pinned`); `api/ffi/handlers/promote_align_test.go` `TestProposeNegativeConflictLine` (line 105, `GetIncomingDisplay() != "male"`) | These change in this PR (§8) |
| Test fixtures | `promotealign_test.go` `newFixture`, `f.person`, `f.bare`, `f.cite`, `f.prop`, `f.term`, `f.promote`, `f.propose`, `rowFor`, `conflictExhibit`; `TestProposeOneHopDateExhibit` (line 390) builds dated birth Events; `api/ffi/handlers/promote_align_test.go` `subjectFixture`, `personOnNewSource`, `promoteOK`, `withProjectCatalog` | Used by the new tests |

**Precedent.** S10-28's `withRowConflicts` (commit `86726467`) and `TestProposeNegativeConflictLine` (commit `b218653e`): a Go-side line built from the row's `ValueRef`s, and an FFI test that it reaches the response. `TestProposeOneHopDateExhibit` is the precedent for a dated birth Event fixture. `neutralTerms` and `termLabels` show the vocabulary map `TermsByID` replaces.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `match.Value.TermLabel string` *(new)* holds the term's catalog label. `Label()` returns `TermLabel` when `Term` is set and `TermLabel` is not empty, else `Term`. Comparers keep comparing by `Term`. `order` (line 228) breaks ties on a private `sortKey()` *(new)*: `Term` for a term value, `Label()` for any other, so the tie-break does not move when a label is renamed | One label for every reader of a `ValueRef` and every sheet line; matching stays independent of copy | Resolving labels in `promotealign` at line-build time — `Rank` reasons would still name keys. Ordering by label — a rename would change which pair a reason names |
| D-2 | `func (v Vocabulary) TermsByID() map[string]VocabTerm` *(new)* in `core/database/graphcache/vocab.go`, keyed by `string(t.ID)`. `neutralTerms` (canon.go), `termLabels` and `termName` (conclusiondetails.go), and `loadTermInfo` and `termInfo` (source_read.go) are deleted; their callers read `VocabTerm` fields (`Key`, `Label`, `Neutral`) from the new map. In `source_read.go` the caller keeps `loadTermInfo`'s guard inline: when `c.Graph()` is nil the map is empty. `event_titles.go` reads `c.Graph().Vocabulary()` (it already calls `c.Graph()` unguarded) and builds its `termLookup` from `TermsByID()`; `sourceeventtitles.FromStored` keeps its signature | The type that owns the data builds the map once; three copies become one | A `termsByID` helper in `promotealign` — a third copy beside two others |
| D-3 | Every loader that builds a term `match.Value` fills `TermLabel`. `sqlSubjectsValues` and `sqlEntityValues` add `COALESCE(t.label, '')` immediately after `COALESCE(t.neutral, 0)`; `SubjectsValues` and `scanEntityValues` scan it into a `termLabel string` right after the neutral variable and set `TermLabel: termLabel`. `loadExhibitObservations` scans its existing `t.label` column into `p.o.value.TermLabel` and drops `pending.termLabel`. `handleValues(props, terms map[string]graphcache.VocabTerm, n, namePattern)` sets `Neutral` and `TermLabel` from `terms[string(row.TermID)]`; a row with no `TermID`, or an id not in the map, gets the zero values. `loadCanon`'s local `neutral` becomes `terms` | The loaders are where term rows enter (S10-28 D-8). The vocabulary is in memory; no query is added | A `TermLabel` on `graphcache.Value` — one string per cached row, for data the vocabulary holds once |
| D-4 | Exhibit text comes from `Value.Label()`. `exhibitObs.display` is deleted; `pairExhibits` sets `IncomingDisplay: in.value.Label()` and `MemberDisplay: chosen.value.Label()`. The if-chain in `loadExhibitObservations` that built `display` is deleted | One function turns a value into sheet text. Both chains agree on every value type except a name with no `Form`, where `Label()` is right (AC-1.5) | Keeping the exhibit chain — two mechanisms for one sheet |
| D-5 | `match.ValueRef.Date *datevalues.Value` *(new)*. `ref()` sets it to `v.Date`, the pointer, with no copy | The sheet formats dates from the structured value. The pointer keeps `BenchmarkPairValues/plain` at 0 allocs/op | Formatting in Go — the client owns locale formatting (`ComparisonValueDisplay`) |
| D-6 | `withRowConflicts` becomes `withRowLines(exhibits []graphalign.Exhibit, comparisons []graphalign.Comparison) []graphalign.Exhibit` *(renamed)*. When `exhibits` is empty, every comparison whose outcome is not `unknown` becomes a line (an unknown comparison has empty refs, and `pairExhibits` drops unknown lines too); otherwise S10-28's rule (a conflict with no own-group conflict line on its Property). Lines come from `refLine(c graphalign.Comparison) graphalign.Exhibit` *(new)*: Property, outcome, value type, weight, `Pinned` false, no Observation ids, no `GroupSubjectID`, `IncomingDisplay`/`MemberDisplay` from the refs' `Label`, `IncomingDate: cloneDate(c.Probe.Date)`, `MemberDate: cloneDate(c.Candidate.Date)`, `IncomingNegative: c.Probe.Negative`. Inside `attachExhibits`, the `return nil` when `want` is empty becomes a skip of the Observation load and pairing, and the function ends with one loop over every row of `prop.Rows`: `r.Exhibits = withRowLines(r.Exhibits, r.Comparisons)`. `Propose` is unchanged. New and Skip rows have no comparisons, so they get no lines, as today | One function in one file decides every row's sheet lines, including when no row has an exhibit. `cloneDate` matches `pairExhibits`, so each line owns its dates | Filling values in the FFI's `else` branch — two places that build lines. Running the loop in `promotealign.go` — splits the sheet between two files |
| D-7 | No Swift change. A ref line is never pinned: `Pinned` false and the box disabled (empty `incomingObservationID`). Before this PR the `else` branch sent `Pinned` true for an `agree` comparison, which drew a checked, disabled box that `batchRow` never committed. A no-exhibit row's `agree` lines now carry `incomingDisplay`, so `draftedArgument` includes them: the argument lists the incoming Subject's agreeing values, and these are its own | A pin needs an Observation pair; a ref line has none. The argument rule already reads that way; the values were missing | Sending the probe's `ObservationID` as the incoming id — a pin also needs a member Observation, which a cached cluster lacks. Filtering ref lines out of the argument |
| D-8 | The FFI's `else` branch is deleted and the `if len(r.Exhibits) > 0` guard is unwrapped into a plain `for _, ex := range r.Exhibits`. `graphalign.Comparison.Pinned` is deleted with its setter in `comparisonsFrom`; the `Comparison` doc comment says comparisons are the score's inputs and pins come from `Exhibit` | D-6 makes the branch dead, and the branch was the field's only reader. A field that suggests drafting pins from a comparison invites the mistake D-7 rules out | Keeping either as a guard — neither would run |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `match.Value.TermLabel` *(new)* | field | `match.Value.Term`, `Neutral` | Extends the value the loaders build; `Term` stays the comparison key |
| `match.Value.sortKey` *(new)* | func | `Value.Label` | `Label` is display; the tie-break needs a key that does not change on a rename |
| `match.ValueRef.Date` *(new)* | field | `graphalign.Exhibit.IncomingDate` | Extends S10-28's ref; the exhibit fields are the destination |
| `graphcache.Vocabulary.TermsByID` *(new)* | method | `neutralTerms` (promotealign), `termLabels` (conclusiondetails), `loadTermInfo` (handlers) | Replaces all three |
| `withRowLines` *(renamed from `withRowConflicts`)* | func | `withRowConflicts` | Extended, not copied |
| `refLine` *(new)* | func | the literal in `withRowConflicts` | Pulled out so both rules build the same line |

**Builds on.** S10-28's `ValueRef` and row-line rule; the graph vocabulary; `cloneDate`; `ComparisonValueDisplay` on the client.
**Deletes.** The FFI `else` branch and its `if` guard. `graphalign.Comparison.Pinned`, its setter, and the two test assertions on it. `exhibitObs.display`, `pending.termLabel`, and the display if-chain in `loadExhibitObservations`. `neutralTerms`, `termLabels`, `termName`, `loadTermInfo`, `termInfo`. The S10-28 Out line "Unifying `Value.Label` with the exhibit display chain" is done.
**Next adds.** A Promote reason sentence reads `Reason.Probe.Label` and gets term labels for free. A term editor ([`docs/ideas/term-management.md`](../../../ideas/term-management.md)) reads terms through `TermsByID`. Lines for non-conflict comparisons on rows with exhibits, if wanted later, are one more case in `withRowLines`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/match` | `Value.TermLabel` *(new)*, `Label()`, `sortKey()` *(new)*, `order` per D-1; `ValueRef.Date` *(new)*, `ref()` per D-5 |
| `core/database/graphcache` | `Vocabulary.TermsByID` *(new)* |
| `core/database/conclusiondetails` | `termLabels` and `termName` deleted; readers use `TermsByID` and `VocabTerm.Key` / `.Label` |
| `core/database/matching` | `sqlSubjectsValues`, `sqlEntityValues`, `SubjectsValues`, `scanEntityValues` per D-3 |
| `core/database/promotealign` | `canon.go`: `neutralTerms` deleted, `handleValues` per D-3. `exhibits.go`: `loadExhibitObservations` and `pairExhibits` per D-3 and D-4; `withRowLines`, `refLine`, and the loop in `attachExhibits` per D-6 |
| `core/graphalign` | `Comparison.Pinned` deleted, `comparisonsFrom` stops setting it, doc comment per D-8 |

Errors: none new.

### FFI

`api/ffi/handlers/promote_align.go` per D-8. `api/ffi/handlers/source_read.go`: `loadTermInfo` and `termInfo` deleted; `listedFromSource` takes `map[string]graphcache.VocabTerm`. `api/ffi/handlers/event_titles.go` builds `termLookup` from `TermsByID()` (D-2). No proto change: `incoming_display`, `member_display`, `incoming_date`, `member_date`, and `incoming_negative` exist.

### macOS

None — no Swift change (D-7).

### Strings

None — no copy change.

## 7. Steps

**STEP-1 — Term labels, a stable tie-break, and dates on refs.**
Files: `core/match/match.go`, `core/match/pairs.go`, `core/match/pairs_test.go`.
Do: D-1 and D-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match` passes (T-1, T-2, T-3); `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench PairValues -benchmem ./core/match` shows `plain` at 0 allocs/op.

**STEP-2 — One vocabulary map.**
Files: `core/database/graphcache/vocab.go`, `core/database/graphcache/graph_test.go`, `core/database/conclusiondetails/conclusiondetails.go`, `api/ffi/handlers/source_read.go`, `api/ffi/handlers/event_titles.go`.
Do: D-2, except `neutralTerms`, which STEP-4 deletes.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache ./core/database/conclusiondetails ./api/ffi/handlers` passes (T-4).

**STEP-3 — The matching loaders read the label.**
Files: `core/database/matching/matching.go`, `core/database/matching/matching_test.go`.
Do: D-3, matching half. Update `TestForSubjectJudgesAgainstTheMajority`: line 477 expects Probe `Female`, Candidate `Male`; line 495 expects Probe `Male`, Candidate `Male`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/matching` passes (T-5).

**STEP-4 — Promote's values read the label, and exhibits take their text from `Label()`.**
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/exhibits.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-3, promotealign half (including `neutralTerms` → `TermsByID`), and D-4. Update `TestProposeJudgesAgainstTheMajority` line 1140: `IncomingDisplay` and `MemberDisplay` expect `Male`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign` passes (T-6, T-7).

**STEP-5 — Every row's lines are built in `attachExhibits`.**
Files: `core/database/promotealign/exhibits.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign ./core/graphalign` passes (T-8, T-9, T-10).

**STEP-6 — The FFI sends exhibits only; `Comparison.Pinned` goes.**
Files: `api/ffi/handlers/promote_align.go`, `api/ffi/handlers/promote_align_test.go`, `core/graphalign/types.go`, `core/graphalign/align.go`, `core/graphalign/align_test.go`, `core/database/promotealign/promotealign_test.go`. Skill: `add-ffi-handler`.
Do: D-8. Update `TestProposeNegativeConflictLine` line 105: `GetIncomingDisplay() != "male"` becomes `!= "Male"`. Remove the `c.Pinned` clauses from `align_test.go` line 256 and `TestProposeWeighsACommonNameLess` line 529; keep their outcome checks.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers -run Promote` passes (T-11); `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Docs.**
Files: `docs/matching.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do: In `docs/matching.md`, the "Reasons." bullet under "Support weights and negatives" says a `ValueRef` names a term by its label and carries a date, and that `attachExhibits` builds every Promote row's lines (a row with no exhibit gets one per comparison). In `completed.md`, add a `## S10-50 — Promote sheet lines show the values they compared` heading above `## S10-28` with one paragraph. In `deployment-plan.md`: the PR sequence line `S10-50  Promote sheet lines show the values they compared (…) (after S10-28)` (near line 243) ends `— done, #N`; the `#### S10-50` entry's `**Plan:**` line (near line 588) ends `— Landed #N` instead of `— Blocked — on S10-28`. This plan: Status `Landed in #N (date)` and a Landed note.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-50-promote-sheet-values.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/match/pairs_test.go` `TestValueLabel` (extend) | `Value{Term: "two-spirit", TermLabel: "Two-Spirit"}` → `Two-Spirit`; `term("male")` → `male` | AC-1.2, AC-1.3, D-1 |
| T-2 | Go | `core/match/pairs_test.go` `TestOrderIgnoresTermLabel` *(new)* | `order(vs, info, 0, 1)` with `vs` two plain term values `{Term: "a", TermLabel: "Zed"}` and `{Term: "b", TermLabel: "Alpha"}` and `info` two equal `voteInfo` entries (same `displayed`, same `eff`) → negative (`a` first); swapped → positive | D-1 |
| T-3 | Go | `core/match/pairs_test.go` `TestPairValuesRefDate` *(new)* | `pairValues(ComparerFor(properties.ValueTypeDate), "", []Value{date(1890, nil, nil)}, []Value{date(1890, nil, nil)})` (helper `date` in `dates_test.go`) → `Probe.Date` and `Candidate.Date` are the input values' `Date` pointers | D-5 |
| T-4 | Go | `core/database/graphcache` `TestTermsByID` *(new)*, in `graph_test.go` | A vocabulary with two terms → a map of two entries keyed by `string(ID)` with `Key`, `Label`, `Neutral` | D-2 |
| T-5 | Go | `core/database/matching/matching_test.go` `TestValuesCarryTermLabel` *(new)* | A Subject with `termed("sex_at_birth", "male")`, read by `SubjectsValues` → `Term` `male`, `TermLabel` `Male`. A handle with sex male, read by `CandidatesOfType` and by `EntitiesValues` → the same | AC-1.4, D-3 |
| T-6 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeJudgesAgainstTheMajority` (update the negative case) | The sex conflict line: `IncomingDisplay` `Male`, `IncomingNegative` true, `MemberDisplay` `Male` | AC-1.1, D-3 |
| T-7 | Go | `core/database/promotealign/promotealign_test.go` `TestExhibitNameWithoutForm` *(new)* | Promote a person whose name Observation is `namevalues.Value{Parts: [{Value: "Mary"}, {Value: "Smith"}]}` with no `Form`; an incoming person with the same parts and no `Form`, fixed to the handle → the name exhibit's `IncomingDisplay` and `MemberDisplay` are `Mary Smith` | AC-1.5, D-4 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeRowWithoutExhibitsShowsValues` *(new)* | Handle: `f.person` `Mary Smith` plus `f.cite` with `observations.Input{PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", "female").ID}`, then `f.promote`. Through `db, _ := f.c.DB()`: `UPDATE identity_claims SET status = 'provisional' WHERE entity_id = ?`, then `f.c.Graph().Drop()`. Incoming the same on a new Source, fixed to the handle → `row.Exhibits` has one own-group line per `row.Comparisons` entry whose outcome is not unknown (name and sex), same Property and outcome, and no `gender` line; name `Mary Smith` / `Mary Smith`; sex `Female` / `Female`; every line `Pinned` false with empty Observation ids | AC-2.1, AC-2.2, D-6, D-7 |
| T-9 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeEventWithoutExhibitsShowsDates` *(new)* | As `TestProposeOneHopDateExhibit` builds a birth Event with `date` 1901: build one with `date` 1890 and promote it; make its claim provisional and drop the graph as in T-8; an incoming birth Event dated 1890 fixed to it → the `date` line has `IncomingDate` and `MemberDate` set, and the `event_type` line reads `Birth` / `Birth` | AC-2.3, D-5, D-6 |
| T-10 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeNegativeDateConflictLine` *(new)* | A promoted birth Event dated 1890 (accepted claim); an incoming birth Event with event type birth and a `date` 1890 Observation with `Polarity: observations.PolarityNegative`, fixed to it → `row.Exhibits` holds a `date` line with outcome `conflict`, `IncomingNegative` true, `IncomingDate` and `MemberDate` set | AC-2.5, D-6 |
| T-11 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposeRowWithoutExhibitsSendsValues` *(new)* | `subjectFixture`; `personOnNewSource(…, "female", observations.PolarityPositive)` and `h := promoteOK(…)`; inside `withProjectCatalog(dir, …)` run the T-8 `UPDATE` with `h.Entity.GetId()` (parsed to bytes) through `c.DB()` and `c.Graph().Drop()`; a second `personOnNewSource(…, "female", positive)` fixed to the handle → every response comparison has `incoming_display` and `member_display` set (name `Mary Smith`, sex `Female`). Both Sources are titled "Register positive"; titles are not unique | AC-2.4, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. The expected updates are named in STEP-3, STEP-4, and STEP-6. Next most likely to break: `macos/ProvenenciaTests` `PromoteFlowTests` that build lines from a fake (they do not read `Comparison.Pinned` from Go, so they should hold).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/matching.md "Reasons." bullet
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Add a negative Observation "sex at birth: male" to a Subject whose handle is male on most Sources, and open Promote's evidence sheet for that row. The conflict line reads "not Male" against "Male", not "male".
2. On a Person handle, make the only member claim provisional (`UPDATE identity_claims SET status = 'provisional' WHERE entity_id = …` in a scratch copy of the project), relaunch, and open Promote for a Subject fixed to it. Every line shows both values, and no pin box is checked.
3. Do the same with a birth Event handle. The date line uses your locale's date format.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A no-exhibit row's drafted argument now lists its agreeing values | Intended (D-7) |
| A no-exhibit row loses its blank `unknown` lines | Intended (D-6): they had no values, and exhibit rows already omit them |
| A no-exhibit `agree` line's box is no longer drawn checked | Intended (D-7): it was never committed |
| An exhibit line's text changes beyond the name-without-`Form` case | D-4 rests on both chains agreeing for every other value type; T-7 and the existing exhibit tests (`TestProposeOneHopDateExhibit` reads `1901` and `Birth`; `TestProposeOverlappingBoundShowsTheDateAndResembles` reads `Before 2001-03-31`) guard it |

**Stop and report if:**

- On `main` after #411 merges, `withRowConflicts`, `ValueRef`, `pairValues`, or the FFI branch no longer match §4.
- `BenchmarkPairValues/plain` reports more than 0 allocs/op after STEP-1.
- A test not named in STEP-3, STEP-4, or STEP-6 needs a changed expected label or display to pass; name it and its old and new values.
- T-8's or T-9's fixture (a provisional-only claim) still yields an exhibit, or yields no row comparisons; report what `acceptedMembers` and `Row.Comparisons` hold.
- `termName` or `termInfo` is read for anything other than `Key` and `Label`.
- A `PromoteFlowTests` case fails because a line's pin state changed.

## 12. Open questions

None.

## 13. PR

**Title:** S10-50 — Promote sheet lines show the values they compared
**Body:** Every Promote sheet line shows the two values it compared, with term labels; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-50-promote-sheet-values.md`.
