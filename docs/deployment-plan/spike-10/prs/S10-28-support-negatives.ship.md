# Ship log — S10-28 Support, negatives, and which values matched

**Plan:** [`S10-28-support-negatives.md`](S10-28-support-negatives.md) · **Branch:** `s10-28-support-negatives` · **PR:** #411
**State:** ready

## Round 0 — implement
Head `c81aa7ad`. Gate PASS (orchestrator run, against `ff14e3ff`). Report deviations:
- `Values.HasPositive` lives in `core/match/pairs.go`, not `match.go`.
- R2 (a value compatible with a `denied` value) fires for any positive value on the asserting side, including a `denied` one (literal reading of R2).
- T-1 rows split across `TestPairValues`, `TestPairValuesAgainstReducesSupport`, `TestPairValuesOrderIndependent`.
- `BenchmarkPairValues/plain` uses `TermComparer` (the text comparer allocates on its own).
- `BenchmarkProposeColdHub`: base 535–577 ms/op, branch 535–574 ms/op (under the 15% bar).

## Round 1 — review
Head `c81aa7ad`. Verdicts: CON changes needed · MNT changes needed · UX changes needed · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | blocker | Support weighting and R3 apply to the multiple-cardinality `toponym` Feature; a Place's minority toponym scores 0 and drops from Rank | To user → Fix now → F-1 | verified: `match/registry.go:287` place profile is `toponym` only; `subjectvocab/registry.go:175` `CardinalityMultiple`; `autoreconcile/pipeline.go:157` keeps every multiple value. D-4's premise "no default profile Feature has cardinality `multiple`" is false. Reviewer's overlay test fails at head (score 0, want 10). Fix reopens D-4. User (2026-10-16): amend D-4, fix here |
| UX-1 | UX | major | The new conflicts never reach Promote when a row has exhibits; a negative Observation lowers the band with no line | To user → Fix now → F-2 | verified: `api/ffi/handlers/promote_align.go:126` sends exhibits only when present; `promotealign/exhibits.go:354` loads positive Observations only. Plan §6 and Manual check 2–3 say the conflict shows. Screen fix crosses D-10 and Out. User (2026-10-16): fix the screen here (widens past D-10) |
| MNT-1 | MNT | major | Fourth copy of "displayed means `kept`"; `Value.Reason` untyped string beside `match.Reason` | Fix now (typing + shared helper) / Follow-up (rename, older call sites) | verified: `autoreconcile.go:146`, `match/pairs.go:135,142`. Rename to `Verdict` contradicts D-8's field name → follow-up |
| MNT-2 | MNT | major | `docs/matching.md:33` still says `ForSubject` reads only positive Observations | Fix now | verified: line unchanged; `sqlSubjectsValues` dropped the filter |
| MNT-3 | MNT | minor | `Denied` duplicates `Reason == denied` | Follow-up | D-8 names the `Denied` field |
| MNT-4 | MNT | minor | `exhibitCompatible` forwards beside `match.compatible` | Follow-up | plan: `exhibits.go` "No other change" |
| MNT-5 | MNT | nit | Credit loop half-uses `pairSet` | Logged | |
| CON-2 | CON | minor | `EntitiesValues` batch query lacks `r.rank` in ORDER BY (D-8) | Fix now (one line, rides with F-1 in the matching package) | verified `matching.go:265`; no caller, no behavior change |
| CON-3 | CON | nit | Plain path calls `Label` for the reported pair, against D-4's wording | Logged (Landed-note line rides with fix round) | AC-3.1 requires labels |
| UX-2 | UX | minor | Minority exact match lowers score while its exhibit line is pinned agree | Follow-up | accepted by D-10 |
| PRF-1 | PRF | minor | Probe re-weighed per candidate; labels built for every scored candidate | Follow-up | constant factor; Promote bench equals base |
| SEC-note | SEC→CON | — | `ORDER BY o.id` orders per batch only | Rejected | D-8 accepts it: one Subject's rows share a batch |

### User decisions (2026-10-16)
- **CON-1:** amend D-4 and fix in this PR. On a multiple-cardinality Property every voting value leads with weight 1 and R3 does not apply; R1 and R2 still apply.
- **UX-1:** fix the screen in this PR, past D-10. A row-level conflict reaches the Promote sheet as a line even when the row has exhibits.

### Fix plan 1

**F-1 — Multiple-cardinality Features: no support weighting, no R3** (from CON-1; amends D-4 by user decision)
Files: `core/match/match.go` (`Feature`), `core/match/registry.go` (`DefaultProfile`), `core/match/pairs.go` (`pairValues`, `weigh`/leaders), `core/match/pair.go` (`compareProperty` passes `card`), `core/match/match_test.go`, `core/match/pairs_test.go`, `core/database/promotealign/promotealign_test.go`, plan D-4 row.
Failing test first: `TestScoreMultipleToponyms` in `core/match/match_test.go` — the CON-1 test: place profile, candidate toponyms `York` (`SetEvidence(3,0,1,"kept")`) and `Eboracum` (`SetEvidence(1,0,2,"kept")`), probe `Eboracum` → `Score` is 10 and `Rank(p, probe, cands, 10)` returns the Place. Add a `pairValues` table row: multiple cardinality, a negative probe value Compatible with a candidate value is still a contradiction (R1). Add `TestProposeMultipleToponyms` in `promotealign_test.go`: a Place handle with two kept toponyms of unequal support (set by SQL like T-8, then `f.c.Graph().Drop()`), incoming the minority toponym → the toponym comparison is `agree`.
Do:
- Add `Cardinality string` to `match.Feature` (same convention as `PropertyMeta.Cardinality`: `properties.Cardinality*`, empty means single). Set `Cardinality: properties.CardinalityMultiple` on the `toponym` Feature in `DefaultProfile("place")`.
- Add a test (package `match_test`, or wherever it can import both without a cycle) that every `DefaultProfile` Feature's `Cardinality` equals the subject vocabulary registry's cardinality for that Property (`core/database/subjectvocab/registry.go`), so the two cannot drift. If an import cycle blocks this, stop and report it.
- `pairValues` takes the cardinality: `pairValues(cmp Comparer, card string, probe, candidate []Value) PairResult`. On `CardinalityMultiple`, every voting value is a leader with weight 1; R3 is skipped; R1 and R2 run unchanged. `Score` passes `f.Cardinality`; `compareProperty` passes `card`. The plain path is unchanged.
- In the plan, amend the D-4 row: replace "No default profile Feature has cardinality `multiple` today, and…" with the multiple rule above, and append "Amended 2026-10-16 in the ship log (CON-1, user decision)." Do not change any other Decision.
Check: the new tests pass; `TestPairValues*`, `TestScoreSupport`, T-8 still pass; `BenchmarkPairValues/plain` still 0 allocs/op; scripts/check-pr.sh PASS.

**F-2 — A row-level conflict shows on the Promote sheet when the row has exhibits** (from UX-1; user decision, past D-10)
Files: `core/graphalign/types.go` (`Comparison`, `Exhibit`), `core/graphalign/align.go` (`comparisonsFrom`), `core/database/promotealign/exhibits.go` (`attachExhibits`), `api/proto/engine.proto` (`PromoteGraphAlignmentComparison`), generated Go/Swift proto code (regenerate with the repo's usual generator), `api/ffi/handlers/promote_align.go`, `macos/App/Features/Promote/PromoteView.swift` (`comparisonLine`), `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, tests. Skills: `add-ffi-handler` (handler test), `add-localized-string`, `add-swift-test` if a Swift test fits.
Failing test first: in `promotealign_test.go`, extend `TestProposeJudgesAgainstTheMajority` (or add `TestProposeConflictReachesExhibits`) so both cases assert on `row.Exhibits`, not only `row.Comparisons`: (a) negative `sex_at_birth` male case — the row has a name exhibit, and `row.Exhibits` contains a `sex_at_birth` line with outcome `conflict`, `Pinned` false, empty `IncomingObservationID`, `IncomingNegative` true, `IncomingDisplay` the probe label, `MemberDisplay` the candidate label; (b) majority case — `row.Exhibits` contains a `conflict` line for the Property whose row comparison is a conflict, alongside any agreeing exhibit line. Add an FFI handler test that such a line reaches the response with `incoming_negative`.
Do:
- Add `Probe, Candidate match.ValueRef` to `graphalign.Comparison`; `comparisonsFrom` copies them from the `PropertyComparison`.
- Add `IncomingNegative bool` to `graphalign.Exhibit`.
- In `attachExhibits`, after the row's exhibits are built: for each `r.Comparisons` entry with outcome `conflict` where no exhibit line in the own group (`GroupSubjectID` nil) on the same Property has outcome `conflict`, append an own-group `Exhibit` with that Property, outcome, value type, weight, `Pinned` false, no Observation ids (so the line cannot be pinned), `IncomingDisplay`/`MemberDisplay` from `Probe.Label`/`Candidate.Label`, and `IncomingNegative` from `Probe.Negative`. Keep it a small named helper in `exhibits.go`. Do not change `pairExhibits` or the member picker.
- Proto: add `bool incoming_negative = 17;` to `PromoteGraphAlignmentComparison`, with a one-line comment. The FFI exhibit branch sets it.
- Swift: in `comparisonLine`, when `incomingNegative` is true, show the incoming value through a new typed L10n string meaning "not <value>" (e.g. `L10n.Promote.negatedValue(_:)`, English "not %@"), in the cell and in the pin's accessibility label. Add it to `Localizable.xcstrings`. No layout change. The pin checkbox is already disabled for an empty `incomingObservationID`.
- In the plan: change the §6 macOS line and Manual check steps 2–3 to say the sheet shows a conflict line (and the callout) for these cases, and change the D-10 Risks row ("A Promote row reads conflict while its exhibit lines show an agreeing pair") to say the conflict line now appears beside the agreeing pair. Append "Amended 2026-10-16 in the ship log (UX-1, user decision)." to D-10.
Check: the new tests pass; `xcode-test` passes; scripts/check-pr.sh PASS.

**F-3 — One "displayed" rule, a typed verdict** (from MNT-1)
Files: `core/autoreconcile/autoreconcile.go`, `core/match/match.go`, `core/match/pairs.go`, tests that build `Value.Reason` by hand.
Failing test first: none (refactor); existing `core/match` and `core/autoreconcile` tests must stay green.
Do: add `func (r Reason) Displayed() bool { return r == ReasonKept }` in `autoreconcile`; `ReconciledValue.Displayed` calls it. Type `match.Value.Reason` as `autoreconcile.Reason`; `SetEvidence` keeps its `reason string` parameter (D-8) and converts once. `displayed()` and `effective()` call `v.Reason.Displayed()`. Comment the field as the auto-reconciler's verdict, not a `match.Reason`. Keep the field name `Reason` and the `Denied` field (D-8). Do not touch `conclusiondetails`, `conclusionheaders`, or `canon.go:237` (follow-up).
Check: `go test ./core/match ./core/autoreconcile ./core/database/...` passes; scripts/check-pr.sh PASS.

**F-4 — Docs** (from MNT-2, CON-3, and F-1/F-2)
Files: `docs/matching.md`, the plan's Landed note.
Do: `docs/matching.md:33` → `ForSubject` reads an Interpretation Subject's Observations, negative ones marked; only a positive value earns credit. In the support/negatives section, add the multiple-cardinality rule (F-1) and that a row-level conflict now shows as a sheet line (F-2). Landed note: one line that the plain path calls `Label` once for the reported pair (a name without `Form` or a date allocates); one line each for F-1 and F-2 as user-approved amendments.
Check: scripts/check-pr.sh PASS (plan-docs).

**F-5 — Batch query orders by rank** (from CON-2)
Files: `core/database/matching/matching.go:265`.
Do: `) ORDER BY r.entity_id, r.rank`.
Check: `go test ./core/database/matching` passes.

## Round 1 — fix
Head `ab0950e1`. Gate PASS (orchestrator run). F-1 to F-5 done. Deviations: F-2's majority (stray-name) case already had a `conflict` exhibit from `pairExhibits`, so only the negative case adds a row-level line; D-4 row text also names the new `pairValues(cmp, card, probe, candidate)` signature; the profile/vocabulary drift test lives in `subjectvocab` (internal; its registry is unexported); F-2 plan edits cover §6 macOS, Manual check 2–3, D-10 and its Risks row.

## Round 2 — review
Scope `c81aa7ad...ab0950e1`. Lenses: CON, MNT, UX (the lenses with a Fix-now finding). Verdicts: all clean. CON-1, CON-2, MNT-1, MNT-2, UX-1 resolved. No new blocker.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-3 | UX | minor | The row-level conflict line shows a term's key (`male`), not its label (`Male`); `match.Value.Label()` returns `v.Term` | Follow-up | round 2 adds blockers only; carry the term label on `ValueRef` or resolve it like `exhibits.go:449` |
| CON-4 | CON→UX | minor | Rows with no exhibits still send conflict comparisons with no values or `incoming_negative` (`promote_align.go:152`) | Follow-up | predates the PR |
| CON-5 | CON→MNT | minor | `TestPairValuesMultiple` "no shared value" case does not prove R3 is skipped (similarity 0 either way) | Follow-up | test strength only |
| MNT-6 | MNT | nit | Duplicate doc comment above `autoreconcile.Reason.Displayed`; `ReconciledValue.Displayed` lost its comment | Logged | |

## Result
Gate PASS (`ab0950e1`) · CI pending · open: none · follow-ups: MNT-1 (rename `Value.Reason`, move `conclusiondetails`/`conclusionheaders`/`canon.go:237` onto `Reason.Displayed`), MNT-3, MNT-4, UX-2, UX-3, CON-4, CON-5, PRF-1 · to user: CON-1, UX-1 (answered 2026-10-16)
