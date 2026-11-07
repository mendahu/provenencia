# Ship log — S13-26 Scoring two Evidence graphs

**Plan:** [`S13-26-scoring-two-evidence-graphs.md`](S13-26-scoring-two-evidence-graphs.md) · **Branch:** `s13-26-scoring-two-evidence-graphs` · **PR:** [#736](https://github.com/mendahu/provenencia/pull/736)
**State:** ready

## Round 0 — implement

Base `d0662711` (`main`). Plan planned against `aaf3b15`. Dispatched `pr-implementer`.
Head `1173f22f`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). PR #736 (draft). Report: deviations — `TestLines` AC-2.2 uses `pt.Given("Amy")` because `pt.Name` reads one word as a surname; `kinship_test.go` `part` / `rel` keep an unused `t` (D-12). STEP-8 (f) Landed edits sit uncommitted in the working tree: the auto-mode classifier denied the implementer's `git commit` ("Credential Leakage"); the diff holds only status lines and landing notes. Taken to the user at finish.

## Round 1 — review

Reviewers: CON, MNT, SEC+PRF (all on the agent's default). Skipped: UX (no `macos/App/**`, strings, `core/apperr` or error codes in the diff). SEC+PRF: matching layers and a new package that loops over Source-written members, links and values.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC+PRF | blocker | `Lines` has no bound on members per entity, links or values; the product reaches 49 s for one graph pair | To user (bound values and over-bound result) | verified: `Lines`, `links` and `properties` nest with no cap; reviewer's probe k=160 → 49 s, k=41 v=50 → 41.5 s. Knobs are the user's (`kinship.MaxPivotEnd` precedent). Corrects Edges "Untrusted size" and D-5 |
| CON-1 | CON | major | A link skipped by shape is lost when the pivot's links do not correspond (father / mother on one birth) | To user | verified: `links` skips on `SameShape` without checking the pivot fact counted; reviewer's row gives 0, and 1 when the births are on two entities. Follows D-7 literally; contradicts §6 "A link skipped by shape loses nothing" |
| CON-2 | CON | major | D-3 first-wins and three D-7 rules have no test; mutants survive | Fix now → F-2 | verified: T-8 uses distinct ids and asserts no exact counts; reviewer ran each mutant |
| MNT-1 | MNT | major | Third copy of the edge builder | Fix now → F-3 | verified; same-pass cleanup. Supersedes the plan's Out line on moving other test key pairs, for `links_test.go` |
| MNT-2 | MNT | major | `t` is dead on `part`, `rel`, `link` | Fix now → F-3 | verified; same-pass. D-12's "keep `t` for their other uses" rests on a false premise |
| MNT-3 | MNT | major | Agree-or-partial is written in three places | Fix now → F-4 | verified: `recordlinks.go`, `graphalign/rows.go:192`, `subjectcompare.go:123`; same-pass, no output change |
| CON-3 | CON | minor | Two records that agree on an unpooled gender (`non_binary`) count nothing | To user | product call: the drop was decided for Promote |
| MNT-4 | MNT | minor | `Rules.Unit` collides with `kinship.Unit` and the unit codes | To user | contradicts D-6's name; see memory on coined terms |
| CON-4 | CON | minor | `evidence-graph-comparison.md:34-35` says one shared word counts, and one fact is thin | Fix now → F-5 | verified; working rule "Each PR updates what it makes false" |
| CON-5 | CON | minor | AC-2.8 passes with the derived rule removed | Fix now → F-2 | verified: `Metas` holds no lifespan keys |
| MNT-5 | MNT | minor | Doc comments overstate or lag `CompareProperty` | Fix now → F-5 | verified |
| MNT-6 | MNT | minor | `partKeys` / `relKeys` only forward | Fix now → F-3 | verified |
| SEC-2 | SEC+PRF | minor | No fuzz target for `Lines` | Fix now → F-2 | verified: `FuzzFamilyUnits` is the neighbor's model |

Nits (logged): `MNT-7 · recordlinks.go:149 · bytes.Clone([]byte(e)) copies twice` · `MNT-8 · recordlinks.go:195 · seen keyed on one side's BridgeKey holds only while Corresponds is equality` · `MNT-9 · recordlinks.go:88 · NewGraph files links links never reads` · `MNT-10 · recordlinks_test.go:32 · entM is a person and a marriage entity` · `MNT-11 · deployment-plan.md:87 · "A matching Property counts" still names propertycompare.Compare` · `CON-6 · matching-model.md:33 · no blank line after the new paragraph` · `CON-7 · add-match-rule/SKILL.md:37 · Record links bullet sits after Rules.Role, not kinship`

### To user — answered 2026-11-07

- **SEC-1:** two knobs, `recordlinks.MaxEntityMembers` = 10 and `recordlinks.MaxUnitValues` = 20. An entity over the member bound keeps its line at 0 facts and counts no links; a property over the value bound counts nothing. → F-1
- **CON-1:** count the link. A pair is skipped only when the pivot's own fact counted. Supersedes D-7's skip rule and corrects D-8's Why. → F-2
- **CON-3:** keep the drop for the first version and pin it. → F-3
- **MNT-4:** rename `Rules.Unit` to `Rules.CountsUnder`; `CompareProperty`'s parameter becomes `key`. Supersedes D-6's name. → F-7

Decisions updated in the table above: SEC-1 Fix now → F-1; CON-1 Fix now → F-2; CON-3 Fix now → F-3; MNT-4 Fix now → F-7. CON-2, CON-5 and SEC-2 are in F-3; MNT-1, MNT-2 and MNT-6 in F-4; MNT-3 in F-5; CON-4, MNT-5 and nits MNT-11, CON-6, CON-7 in F-6.

### Fix plan 1

Order: F-7 first if it makes the rest easier; the implementer may order commits as it likes, one per behavior. Every F-# closes with the Landed note in the plan listing what it changed from the plan, by Decision ID.

**F-1 — Bound what one `Lines` call can compare** (from SEC-1)
Files: `core/recordlinks/recordlinks.go`, `core/matchregistry/recordlinks.go`, `core/recordlinks/recordlinks_test.go`, `docs/matching-model.md`. Skill: `add-match-rule` "Add a knob".
Failing test first: `TestLinesBounds` *(new)*, in a table:
- 10 person members on F in A → the line counts as usual; 11 → one line for F, 0 facts, weak.
- 11 members of any kind on F in B only → the same (symmetric).
- A far entity E with 11 members in one graph → links to E count nothing, the line itself stands.
- A unit with 20 values on one side → compares as usual; 21 → counts nothing, other units still count. Pooled values add up per unit (11 gender + 10 sex at birth = 21 → counts nothing).
Do:
- `Rules` gains `MaxEntityMembers` and `MaxUnitValues` with doc comments; `defaultRecordLinks` sets 10 and 20 with `// decided: user, 2026-11-07 (S13-26 SEC-1): …` comments.
- `Graph` gains a count of members of any kind per non-nil entity. `NewGraph` keeps its signature: the bound is read in `Lines`, through one helper (say `over(e, a, b, cfg)`), true when either graph's count for e exceeds `MaxEntityMembers`.
- `Lines`: over the bound, the line is made at 0 facts with no pair compared (strength by the usual rule).
- `links` takes `cfg` and passes over a far entity that is over the bound.
- `properties`: count each side's values per unit while collecting units (the sum of `len(v[key])` over keys that count under the unit); a unit where either side's count exceeds `MaxUnitValues` is not compared. Do not call `Side` from `recordlinks`.
- In `links`, test `seen[lx.key]` in the outer loop, before the walk over `byY[e]`.
- Add two rows to the inventory after `recordlinks.StrongAt`.
- `BenchmarkLinesAtBounds` *(new)*: one pair at the worst input the bounds allow (10 members on each of two entities, fully linked, 20 values on a name and a multiple-term property). Report its ns/op in the fix report.
Check: `TestLinesBounds` passes; `TestKnobInventoryMatchesRegistry` and `TestDefaultIsFresh` pass; scripts/check-pr.sh PASS.

**F-2 — Skip a link only when its pivot's fact counted** (from CON-1)
Files: `core/recordlinks/recordlinks.go`, `core/recordlinks/recordlinks_test.go`.
Failing test first: `TestLines` rows:
- `pivot on one entity, roles differ`: reviewer's row. A: F `father` of a birth of the page's person on entE. B: F `mother` of the same. → `{entF, 1, weak}`.
- The same with `rel(…, "parent", …)` written down on both sides → 1 fact.
- A chain case: the grandparent link through a middle parent filed on one entity, whose parent links do not correspond → the grandparent link counts.
Do: in `links`, `same(pa, pb)` becomes true only when the pivots are filed on one non-nil entity E **and** x's and y's links to E hold a corresponding pair (`segmentcompare.Corresponds`), E not over the member bound. No recursion: a pivot link that is itself skipped is counted through its own pivot, and shapes end at a recorded participation or parent link. Memoize by E inside one `links` call. Update `links`' doc comment. AC-2.9 to AC-2.19 stay as they are.
Check: new rows pass; every existing row passes unchanged; `TestLinesInvariants` passes; scripts/check-pr.sh PASS.

**F-3 — Pin the rules no test holds** (from CON-2, CON-3, CON-5, SEC-2)
Files: `core/recordlinks/recordlinks_test.go`, `core/recordlinks/fuzz_test.go` *(new)*.
Failing test first: none: these pin behavior that holds at head; F-2's change must leave them true.
Do:
- `TestLines` rows from CON-2: `D-3 first member wins`, `D-7 both pivots unfiled`, `D-7 two links one key one entity`, `D-7 any pair not skipped`, and a zero `recordlinks.Graph{}` row giving nil. Row inputs are in round 1's CON-2 finding, copied here: first wins: A `pageA, person(idA(2), entF, name john), person(idA(2), entS, name john)`, B `pageB, person(idB(2), entF, name john), person(idB(3), entS, name john)` → `{entF, 1, weak}`. Both unfiled: AC-2.14 with both events on nil → `{entF, 1, weak}`. Two links: A F subject of two births on entE, B of one on entE → `{entF, 1, weak}`. Any pair: A `pageA, person(idA(3), entP), person(idA(2), entF), event(evA(1), entM), event(evA(2), nil)`, F and `idA(1)` subjects of `evA(1)`, F and `idA(3)` subjects of `evA(2)`; B AC-2.14's graph → `{entF, 2, strong}`.
- CON-3: a row with `gender: non_binary` on both sides → `{entF, 0, weak}`, named so it reads as decided.
- CON-5: the AC-2.8 row runs with a local copy of metas that gives `lifespan` and `lifespan_span` a date meta.
- `FuzzLines` on the model of `FuzzFamilyUnits`: decode bytes into two graphs with bounded sizes, assert no panic and `Lines(p,a,b)` equals `Lines(p,b,a)`.
Check: each row passes; `go test ./core/recordlinks/ -run FuzzLines` passes on its seeds; scripts/check-pr.sh PASS.

**F-4 — One edge builder for tests** (from MNT-1, MNT-2, MNT-6)
Files: `core/kinship/kinshiptest/kinshiptest.go` *(new)*, `core/matchregistry/guard_test.go`, `core/kinship/kinship_test.go`, `core/kinship/links_test.go`, `core/kinship/siblings_test.go`, `core/kinship/families_test.go`, `core/recordlinks/recordlinks_test.go`, `core/graphalign/align_test.go`.
Failing test first: none: a test-only refactor; the existing suites are the check.
Do:
- `kinshiptest.Part(person, event []byte, role vocabkey.BridgeTerm, eventType vocabkey.EventType, assocs ...[]byte) kinship.Edge` and `kinshiptest.Rel(person, relatedTo []byte, term vocabkey.BridgeTerm, assocs ...[]byte) kinship.Edge`, built on `segmentcomparetest.PartKeys` / `RelKeys`; package doc in the shape of the other `*test` packages. `layerPackages` row `{"core/kinship/kinshiptest", false}` after `core/kinship`.
- `kinship_test.go`: `part`, `rel` and `link` drop `t` and stay thin id adaptors over `kinshiptest`; update every caller (about 69 lines across `kinship_test.go`, `siblings_test.go`, `families_test.go`).
- `links_test.go`: the `part` and `spouse` closures use `kinshiptest`, wrapping in `kinship.Link`; remove what that leaves unused.
- `recordlinks_test.go`: `part` and `rel` move onto `kinshiptest`, or are deleted if a direct call reads as well.
- `align_test.go`: delete `partKeys` and `relKeys`; `partLink` and `relLink` call `segmentcomparetest.PartKeys` / `RelKeys`. `bridge` stays (S13-29).
This supersedes D-12's "keep `t` for their other uses" and, for `links_test.go`, the plan's Out line on other hand-built key pairs. Other Out-line files stay.
Check: `rg -n "kinship\.Edge\{A:" core --glob '*_test.go'` prints no participation or relationship builder that computes keys itself, only one-off edges inside single tests, each listed in the report; `rg -n "\b(partKeys|relKeys)\(" core` prints nothing; `rg -n "(part|rel|link)\(t," core/kinship` prints nothing; scripts/check-pr.sh PASS.

**F-5 — One word for "this outcome supports the match"** (from MNT-3)
Files: `core/propertycompare/propertycompare.go`, `core/propertycompare/propertycompare_test.go`, `core/recordlinks/recordlinks.go`, `core/graphalign/rows.go`, `core/subjectcompare/subjectcompare.go`.
Failing test first: `TestOutcomeMatches` *(new)* — agree and partial true; conflict and unknown false.
Do: `func (o Outcome) Matches() bool` with a short doc comment. The three sites use it. `subjectcompare.Compare`'s switch becomes that check and the conflict branch, with no output change.
Check: `rg -n "OutcomeAgree \|\||OutcomePartial \|\||OutcomeAgree, propertycompare.OutcomePartial|!= propertycompare.OutcomePartial" core --glob '!*_test.go'` prints nothing but `Matches`; `core/subjectcompare` and `core/graphalign` tests pass unchanged; scripts/check-pr.sh PASS.

**F-6 — Docs say what shipped** (from CON-4, MNT-5, nits MNT-11, CON-6, CON-7)
Files: `docs/deployment-plan/spike-13/evidence-graph-comparison.md`, `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/pairs.go`, `.agents/skills/add-match-rule/SKILL.md`, `docs/matching-model.md`, `docs/deployment-plan/spike-13/deployment-plan.md`.
Failing test first: none: comments and docs.
Do:
- `evidence-graph-comparison.md:34-35`: reword both bullets to match the map's working rule (a name the name compare does not call a conflict counts; a line with 0 or 1 fact is thin).
- `CompareProperty`'s doc: "the one place `Side` meets property compare" (the lifespan rule calls `propertycompare.Compare` directly).
- `Side`'s doc and parameter: `lead` becomes `key`, since it now takes non-lead keys.
- `add-match-rule`: a Do-not line against composing `Side` with property compare outside `CompareProperty`; move the Record links bullet after the kinship bullet; "Its one knob" becomes the three knobs.
- `matching-model.md:33`: a blank line after the new paragraph.
- `deployment-plan.md:87`: "A matching Property counts" names `subjectcompare.CompareProperty`. The working rule "A matching link counts": a pair is skipped through a pivot only when the pivot's fact counted.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-26-scoring-two-evidence-graphs.md` reports no errors; scripts/check-pr.sh PASS.

**F-7 — `Rules.Unit` becomes `Rules.CountsUnder`** (from MNT-4)
Files: `core/subjectcompare/subjectcompare.go`, `core/subjectcompare/subjectcompare_test.go`, `core/recordlinks/recordlinks.go`, `.agents/skills/add-match-rule/SKILL.md`, `docs/matching-model.md`, `docs/deployment-plan/spike-13/deployment-plan.md`.
Failing test first: none: a rename; `TestUnit` becomes `TestCountsUnder`.
Do: rename the method; `CompareProperty`'s `unit` parameter becomes `key`; `recordlinks`' `units` and doc words say "keys that count under" or "properties"; `MaxUnitValues` stays as the user answered it. Keep the plan and the review log as written and note the rename in the Landed note. Do not rewrite this ship log.
Check: `rg -n "\.Unit\(|Rules\.Unit|TestUnit" core .agents docs/matching-model.md docs/deployment-plan/spike-13/deployment-plan.md` prints nothing outside `kinship` and valuecompare's unrelated hits; scripts/check-pr.sh PASS.

## Round 1 — fix

Head `0d7bd8bb` (local, not pushed). Gate PASS (orchestrator re-ran on the tree `07fb46f4` commits). F-1 to F-5 and F-7 committed; F-6 done in the working tree, uncommitted. `BenchmarkLinesAtBounds` about 221 ms/op on an M1 against the reviewer's 49 s probe. Deviations: `TestOutcomeMatches` and F-1's test failed on compile or were written after the fields, all in one commit; the "parent links do not correspond" chain row hand-builds graph B, because no derived chain can produce it, with a derived chain row beside it; `held` helper because the numbers guard rejects a literal 2.

The classifier denied the implementer's docs commit a second time ("Credential Leakage"). It stopped and did not retry. Orchestrator scan of the uncommitted diff: 6 files, 28 changed lines, no secret-like strings; remote is SSH with no token. Taken to the user; the orchestrator has not run the denied commit.

User reviewed the diff and approved the commit and push, 2026-11-07. Orchestrator committed the six files by name as `07fb46f4` and pushed.

## Round 2 — review

Reviewers: CON, MNT, SEC+PRF, each had a Fix-now blocker or major. Scope: `1173f22f...07fb46f4`.

CON: CON-1 Resolved, CON-2 Resolved (mutants re-run; one equivalent), no new finding. MNT: MNT-1, MNT-2, MNT-3 Resolved, no new finding. SEC+PRF: **SEC-1 Not resolved.**

Minors fixed in round 1, read by the orchestrator through the reviewers' reports and the fix report: CON-3 Resolved (row pins it), CON-4 Resolved, CON-5 Resolved (mutant fails the row), MNT-4 Resolved, MNT-5 Resolved, MNT-6 Resolved, SEC-2 Resolved (`FuzzLines`, 1.63M execs in CON's scratch run).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 (restated) | SEC+PRF | blocker | Inside both knobs one graph pair still runs for minutes: words per value, shared entities and shapes per link are open in product | To user | reviewer's probes: 10×10 members with 20 `interests` values of 64 words did not finish in 70 s; 64 shared entities 12.7 s; 800 marriages 16 s (quadratic). `BenchmarkLinesAtBounds` is not the worst input: it uses two-word names, one shape, two entities. The fix plan's "worst input the bounds allow" was the orchestrator's wording and is wrong; the Landed note repeats it |

Below the round-2 bar, from CON: the pivot's "E within the member bound" clause and the per-entity memo have no test (two probe rows pass at head, mutants survive); `deployment-plan.md:90` "The cutoff, 2, is the one knob" is stale.

### To user — answered 2026-11-07

- **SEC-1 (restated):** the user: this would never happen unless someone was trying to break the app, so stop it with a high cap a researcher would never hit, and do not design further. One cap on the whole pair, counted before comparing. → F-8
- **Third review:** approved for SEC+PRF only, on the fix commits. The two gaps below the bar ride along. → F-9

### Fix plan 2

**F-8 — One high cap on the work of one graph pair** (from SEC-1 restated)
Files: `core/recordlinks/recordlinks.go`, `core/valuecompare` (a word-count helper, if none exists), `core/matchregistry/recordlinks.go`, `core/recordlinks/recordlinks_test.go`, `docs/matching-model.md`, the plan's Landed note, `completed.md`. Skill: `add-match-rule` "Add a knob".
Failing test first: `TestLinesWorkCap` *(new)*: a graph pair over the cap on each axis alone (words: long values; shared people: many entities; shapes: many marriages between the same people) gives one line per shared person, each 0 facts and weak, in both argument orders; a pair just under the cap counts as usual.
Do:
- `Rules` gains `MaxPairWork`; `defaultRecordLinks` sets it with `// decided: user, 2026-11-07 (S13-26 SEC-1): a high cap a researcher never reaches, so a crafted catalog cannot hang a read`; inventory row after `recordlinks.MaxUnitValues`.
- At the top of `Lines`, before any compare, add up the work for the whole call. For each shared person entity within `MaxEntityMembers`, each member pair: for each property that would be compared, words on x × words on y (a term or other one-word value is 1 word; the values' words add up per side, so values and words share one number); for each far entity both members link to, the shapes of x's links × the shapes of y's links, a link with no shape counting 1. The count stops as soon as it passes the cap. It costs no pair compare: linear in the input per member pair. Symmetric: it must not depend on argument or input order.
- Over `MaxPairWork`: every shared person still gets a line, 0 facts, strength by the usual rule, with nothing compared. This is the existing over-bound rule, one level up.
- Keep `MaxEntityMembers` and `MaxUnitValues` as they are. Do not add a shape index, per-axis word caps, or a shared-people cap: the one cap covers them.
- Value: set from measurement. High: at least 20 times the work of the 41-Subject obituary with realistic values (each person a three-word name, sex, one or two more properties, a handful of links), and the worst input just under the cap costs about half a second or less on this Mac. Report both numbers, and the value, in the fix report.
- `BenchmarkLinesAtBounds` becomes `BenchmarkLinesAtCap`: the slowest input just under the cap, built on each worst axis in turn as sub-benchmarks: names at `Names.MaxExactPairing` words; `interests` at `Text.MaxWords` words; many shared people; many shapes between the same people. Report each ns/op.
- Correct the Landed note: drop "`BenchmarkLinesAtBounds` is the worst input the bounds allow"; add the cap and what it does.
Check: `TestLinesWorkCap` passes; `TestKnobInventoryMatchesRegistry`, `TestDefaultIsFresh`, `TestLinesInvariants`, `FuzzLines` seeds pass; every `TestLines` row unchanged; the reviewer's round-2 inputs (20 `interests` of 64 words at 10×10 members; 64 shared entities; 800 marriages) each finish under the cap's budget in a test or sub-benchmark; scripts/check-pr.sh PASS.

**F-9 — Pin the pivot's bound and the memo; fix the map line** (from CON round 2, below the bar)
Files: `core/recordlinks/recordlinks_test.go`, `docs/deployment-plan/spike-13/deployment-plan.md`.
Failing test first: none: these pin behavior at head; each row must fail on its mutant.
Do:
- `TestLinesBounds` row: AC-2.14 with 11 events on `entM` in A → `{entF, 1, weak}` (fails if `over(…)` leaves `same`).
- `TestLines` row: marriage on `entM` plus father / mother on `entE` → `{entF, 2, strong}` (fails if the `counted` memo is keyed on one fixed key).
- `deployment-plan.md:90`: "The cutoff, 2, is the one knob, in the match registry's record-links section." says the cutoff is the one tuning knob there, and that the section's other knobs only stop malformed data.
Check: both rows pass; scripts/check-pr.sh PASS.

## Round 2 — fix

Head `b7975705`, pushed. F-8 and F-9 done (`aade61a5`, `b7975705`). `recordlinks.MaxPairWork` = 25,000, counted before any compare in `pairWork`; obituary-sized work 1,029, so the cap is about 24× it. `BenchmarkLinesAtCap`: names 596 ms, shared people 621 ms, interests 3.5 ms, shapes 0.08 ms. Slowest axis about 0.6 s, a little over the plan's "about half a second": the 20× floor (20,580) sets it. Round-2 inputs trip the cap in milliseconds. Deviations: no red run, since the names did not compile first; F-9's rows ride in `aade61a5`; `export_test.go` exposes `PairWork`; `valuecompare.WordCount` *(new)*.

## Round 3 — review

User-approved, SEC+PRF only. Scope: `07fb46f4...b7975705`. Gate PASS on `b7975705` (orchestrator re-ran). The count is bounded, symmetric and order-free (4,000 seeded rounds); link and shape work is fully counted. **SEC-1 Not resolved:** two inputs get past the count.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-3 | SEC+PRF | blocker | The count does not see word length; a property over `MaxUnitValues` costs 0 but is tokenized per pair | To user | reviewer's probe: 10×10 members, 15 one-word 64 KB names, work 22,500 → 78.8 s, linear in length; `held` path work 0 → 10.1 s at 131 KB per word |
| SEC-4 | SEC+PRF | blocker | A value with both name and text is counted as the name, compared as the text | To user | probe: work 22,500 → 29.3 s; no CHECK in `observations`; `graphfill/load.go:120` reads both columns |

Below the bar, from SEC+PRF: one 159-word `interests` on one person in two records trips the cap and drops every line in the pair to 0 facts, though text over `Text.MaxWords` exits cheaply. That contradicts "a cap a researcher never hits". The cap is per pair: 0.6 s × 1,225 pairs is about 12 minutes for a read where every graph is crafted; arithmetic, for S13-27's budget.

### To user — answered 2026-11-07

- **SEC-3, SEC-4:** one size limit, then ship. `recordlinks.MaxValueBytes` = 256; no fourth review. The reviewer's probe inputs become tests. → F-10

### Fix plan 3

**F-10 — One size limit on a value, and count everything a value holds** (from SEC-3, SEC-4)
Files: `core/recordlinks/recordlinks.go`, `core/valuecompare/words.go`, `core/matchregistry/recordlinks.go`, `core/recordlinks/recordlinks_test.go`, `docs/matching-model.md`, `completed.md`, the plan's Landed note, `.agents/skills/add-match-rule/SKILL.md`. Skill: `add-match-rule` "Add a knob".
Failing test first: `TestLinesValueSize` *(new)*, reds run first since the names exist:
- 10×10 members on one entity, 15 one-word 64 KB names each: `{entF, 0, weak}`, under a second (the probe's 78.8 s input).
- 21 two-word `interests` of long words per member (the over-`MaxUnitValues` path): under a second.
- 10×10 members, 15 `interests` of 64 words plus `Name` "x" on each value: the cap trips (the probe's 29.3 s input). A row with both fields set.
- One 159-word `interests` on one person in both records, other people shared: that property counts nothing, every other line counts as usual, the cap does not trip.
- A value of exactly 256 bytes compares as usual; 257 makes the property count nothing.
Do:
- `Rules` gains `MaxValueBytes`; `defaultRecordLinks` sets 256 with `// decided: user, 2026-11-07 (S13-26 SEC-3): a high limit on one value, so a crafted value cannot hang a read; a longer note loses that one fact`; inventory row after `recordlinks.MaxPairWork`.
- A helper gives a value's size: the sum of `len` of every string it holds (`Text`, the name's form and parts, a term), with no parsing and no `switch` precedence, so its cost is the fields, not the words. Check it before `WordCount` tokenizes anything.
- A property where either side holds a value over `MaxValueBytes` is not compared and charges 0 in `pairWork`, the same over-bound rule as `MaxUnitValues`. Both `properties` and `pairWork` read one shared check, computed per member once rather than per member pair.
- `WordCount` adds up name words and text words instead of taking the name first (SEC-4).
- `properties` stops tokenizing per pair: `held` counts `len` of values only.
- `BenchmarkLinesAtCap` gains a sub-benchmark at the longest words `MaxValueBytes` allows.
- The Landed note and `completed.md` name the limit. `add-match-rule` says five knobs.
Check: `TestLinesValueSize` passes after failing red; `TestLinesWorkCap`, `TestLinesInvariants`, `FuzzLines` seeds, every `TestLines` row and `TestKnobInventoryMatchesRegistry` pass; each sub-benchmark's ns/op reported, the slowest at about a second or less; scripts/check-pr.sh PASS.

## Round 3 — fix

Head `8b44b41a`, pushed. Gate PASS (orchestrator re-ran). F-10 done; `TestLinesValueSize` red first (the 64 KB row took 70 s). SEC-3 and SEC-4 Resolved by the reviewer's own inputs as test rows; no fourth review, by the user's answer. `BenchmarkLinesAtCap`: names 651 ms, shared people 610 ms, longest values 174 ms, interests 3.7 ms, shapes 0.07 ms. Deviation: `TestLinesWorkCap`'s text axis rebuilt with a `compact` helper, since long-word values now trip the size limit first.

Later PR: the whole-read bound goes to S13-27's scope entry in the map (Check line), where the read's budget subtest lives.

## Result

Gate PASS · CI: see PR · open: none · follow-ups: move the value limit into `propertycompare` so Promote gets it (user's idea, 2026-11-07; home to be asked) · nits logged: MNT-7, MNT-8, MNT-9, MNT-10 · to user: all answered; the user approved a third review round for SEC+PRF
