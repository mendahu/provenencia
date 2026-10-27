# S11-43 — Lower the candidate gate

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 10 — Name candidates
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the target menu already lists a row's alternatives; this PR only lets one more record into that list)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s11-43-candidate-gate`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [name candidates § Case 1](../name-candidates.md#case-1--amy-against-amy-lee-anne-long), [§ What stays later](../name-candidates.md#what-stays-later), [`matching-model.md` § Knob inventory](../../../matching-model.md#knob-inventory), [bands](../comparison-layers.md#bands); working rules: "The candidate gate drops to 0.4", "Gate only; the curve is held", "Registry numbers: carry over, then estimate", "Bands come from evidence"
**Planned:** 2026-10-27 against `main` @ `bac7368`
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

A Source person recorded only as "Amy" is offered the record "Amy Lee Anne Long" in Promote's target menu. Today the record is shortlisted (both share the word `amy`) but the candidate gate drops it: name similarity is 2 / 4.5 = 0.44, under `alignment.MinSimilarity` 0.5. This PR lowers the gate to 0.4, so a single given name reaches a record with up to three more given names and a surname.

**Why now.** Dogfood, 2026-10-24 ([case 1](../name-candidates.md#case-1--amy-against-amy-lee-anne-long)). The user chose the gate over the name curve the same day (working rule "Gate only; the curve is held").

**What stays.** The name curve, every band and every score: bands come from evidence, and evidence does not read the gate. A row the walk does not place still has no band, so Swift still opens it unset (`PromoteFlow.draftTarget`, `assessment == .none`). What Go sends as the target of an unplaced row (today Skip `weak` when a candidate fell short, New `no_match` when there was none) is S11-47's change, not this one.

## 2. User stories

**US-1.** As a researcher promoting a Source that names a woman only as "Amy", when the catalog has "Amy Lee Anne Long", I see that record in the row's target menu so that I can file the Source on her without searching.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Subject name parts: given `Amy`, no surname. Handle `PER-1` name parts: given `Amy`, given `Lee`, given `Anne`, surname `Long`. `graphalign.Candidates` with the shipped registry | `PER-1` is a candidate; subject compare similarity is 2 / 4.5 (0.444…) |
| AC-1.2 | Same Subject. Handle `PER-2`: given `Beth`, given `Amy`, surname `Long` (similarity 0.375) | `PER-2` is not a candidate |
| AC-1.3 | AC-1.1's Subject and handle through `graphalign.Align` (via the test helper `alignWith`), nothing fixed | The row is not placed (`Target` is not `TargetHandle`), its `Assessment` is `AssessNone`, and its one alternative is `PER-1` with `Assessment` `AssessNone` |
| AC-1.4 | The whole Go suite at the new gate | Every existing test and golden passes unchanged: the obituary golden and sparse cases keep their targets |
| AC-1.5 | A real project with a person "Amy Lee Anne Long"; a Source whose person is "Amy" (given part only) | The Promote row for "Amy" opens unset and its target menu lists "Amy Lee Anne Long" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `alignment.MinSimilarity` 0.5 → 0.4 in `core/matchregistry/alignment.go`, with a source comment that drops "at least half alike" and names the Amy case | STEP-2, AC-1.1, AC-1.3 |
| The knob inventory row in `docs/matching-model.md` carries 0.4 and the same reason | STEP-2 |
| A `graphalign` test that a one-given-name Subject and a record with three more names is a candidate and shows as an alternative on an unplaced row | STEP-1, T-1, T-2 |
| A `graphalign` test that a record whose first given name differs is still not a candidate | STEP-1, T-1 |
| The docs that state the gate as 0.5 or "half alike" say 0.4 (D-5) | STEP-3 |

| Out | Home |
| --- | --- |
| Relaxing the name curve (one-side-only words at less than full weight) | Held until dogfood after this PR ([what stays later](../name-candidates.md#what-stays-later)); not planned in Spike 11 |
| An unplaced row's target (Skip `weak`, New `no_match`) and the possible-duplicate check | S11-47 |
| "Des" against "Desiree" | S11-44 (nickname sets), S11-46 and S11-45 (prefix rule) |
| Surname-only records tying at similarity 1 | `docs/ideas/coverage-aware-similarity.md` |
| A Source that recorded "Amy" as a surname (conflicts with "Long", evidence −3) | Not planned: a data-entry fix, [case 1](../name-candidates.md#case-1--amy-against-amy-lee-anne-long) last paragraph |
| A `promotealign` (catalog-level) test for the Amy case | Not planned: D-3 |

## 4. Current state

Verified at `bac7368`.

| Area | Where | Today |
| --- | --- | --- |
| The gate's value | `core/matchregistry/alignment.go` `defaultAlignment` | `MinSimilarity: 0.5, // estimate: at least half alike; a name with a sex conflict is 0.71, a birth against a death 0.09`. The only place the number lives |
| The gate's field | `core/graphalign/rules.go` `Rules.MinSimilarity` | Doc comment "the least similarity for a shortlisted handle to be a candidate"; no number. Unchanged by this PR |
| The gate | `core/graphalign/candidates.go` `Candidates` | Drops a handle when `r.Similarity < cfg.Alignment.MinSimilarity`, then `keepTop`. Doc comment names no number. Unchanged |
| Name similarity | `core/valuecompare/name.go` step 9 | Paired words' earned weight over every word on both sides; weights surname 1.5, first given 1, other given 0.5. Measured at the base with the shipped registry, Subject given-only `Amy`: `Amy Long` 0.571, `Amy Lee Long` 0.500, `Amy Lee Anne Long` 0.444, `Amy Lee Anne Marie Long` 0.400000…02, `Beth Amy Long` 0.375, `Beth Long` 0 (conflict) |
| Row at the base | `core/graphalign/rows.go` | At 0.5, AC-1.3's pair gives Target `new`, Reason `no_match`, no alternatives. At 0.4 (probe, 2026-10-27) it gives Target `skip`, Reason `weak`, Assessment `none`, one alternative `PER-1` at score 1.111, Assessment `none` |
| Swift draft | `macos/App/Features/Promote/PromoteFlow.swift` `draftTarget` | `assessment == .none` → `.unset`; the menu is built from the proposal's alternatives (`menu(for:kind:orders:)`). Unchanged |
| Inventory row | `docs/matching-model.md` line 195 | `` | `alignment.MinSimilarity` | 0.5 | estimate: at least half alike; … | Least similarity for a shortlisted handle to be a candidate | `` |
| Inventory guard | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | Checks that each registry field has a row by name; does **not** compare values. The row's 0.5 must be edited by hand |
| Bound guard | `core/matchregistry/registry_test.go` (line 93) | Fails if `MinSimilarity` is outside [0, 1]. 0.4 passes |
| Docs stating 0.5 / "half alike" | `docs/matching.md` line 16; `docs/promote-graph-alignment.md` line 103; `docs/ideas/coverage-aware-similarity.md` line 7; `docs/deployment-plan/spike-11/name-candidates.md` line 10 | "at least half alike (`graphalign.Candidates`)"; "at least half alike (similarity 0.5)"; "`alignment.MinSimilarity` (0.5)"; "`alignment.MinSimilarity` (0.5)" |
| Test fixtures | `core/graphalign/align_test.go` `nameVals(given, surname)`, `surnameVals(surname)`, `nameMetas()`, `nameOf`, `alignWith`, `rowBySubject`, `id` | `nameVals` builds exactly one given part and one surname. No package-level builder for a name of several given parts or of a given part alone; `TestPartialNamesAreWeak` has local `given` / `sur` / `full` closures |
| Fixtures to avoid | `core/propertycompare/propertycomparetest` `Name` → `namevaluetest.Western` | A one-word form is a **surname**: `Name("Amy")` against `Name("Amy Lee Anne Long")` is a conflict (similarity 0). Do not use it for the Amy Subject |
| Candidate tests | `core/graphalign/candidates_test.go` `TestCandidates` | Subtests for the floor (`James Robins` vs `Mary Smith`, similarity 0), kind, empty probe, sex conflict, ties, cap, birth vs death, top K. All still pass at 0.4 |
| Whole suite at 0.4 | `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` | Run at the base with only the number changed (2026-10-27): every package passes. The map review saw the same on 2026-10-24 |

**Precedent.** `bdbfabfa` ("widen its window by ten years") changed a subject-section number: the value and its source comment in `core/matchregistry/subject.go`, the inventory row in `docs/matching-model.md`, and the tests that pin the behavior, in one commit. Copy that shape. Do not copy its test churn: no existing test changes here (AC-1.4).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The value is 0.4**, written `MinSimilarity: 0.4,` with the comment `// decided 2026-10-24 (S11-43): one given name reaches a record with three more names ("Amy" / "Amy Lee Anne Long" 0.44); a name with a sex conflict is 0.71, a birth against a death 0.09`. gofmt realigns the block's comments; accept its output | The user's call (working rule "The candidate gate drops to 0.4"); "decided" is the source word the registry uses for user calls (`// decided: comparison-layers layer 6 procedure`). The two old reference points stay because they still explain where 0.4 sits | 0.44 or 0.444 — tuned to one case, and the rule says no tuning beyond the goldens; keeping "estimate" — it is now a decision |
| D-2 | **The inventory row** becomes `` | `alignment.MinSimilarity` | 0.4 | decided 2026-10-24 (S11-43): one given name reaches a record with three more names ("Amy" / "Amy Lee Anne Long" 0.44); a name with a sex conflict is 0.71, a birth against a death 0.09 | Least similarity for a shortlisted handle to be a candidate | `` | One wording in code and doc; the guard does not compare values, so the row is edited by hand | — |
| D-3 | **Tests live in `core/graphalign` only**: one new `TestCandidates` subtest (T-1) and one new top-level test in `align_test.go` (T-2). No catalog-level test | The scope names a `graphalign` test; `graphalign` is pure and is where the gate is applied. The catalog path (shortlist → `Candidates` → `Align`) is already covered for names by `promotealign` tests that pass unchanged; the shortlist key `name\|w:amy` is shared and needs no new proof | A `promotealign` test — a catalog fixture for a one-line number change, and the gate is not in that package |
| D-4 | **T-2 does not assert the row's `Target` or `Reason`** beyond "not `TargetHandle`". It asserts `Assessment == AssessNone` and the one alternative (`HandleID`, `Assessment == AssessNone`) | S11-47 changes an unplaced row's target and reason and runs in parallel ("any time"). Asserting `TargetSkip` / `ReasonWeak` here would make S11-47 rewrite a test that is about the gate | Asserting `TargetSkip` + `ReasonWeak` — couples this test to S11-47's subject |
| D-5 | **Docs sweep, numbers only.** `docs/matching.md:16` "at least half alike" → "at least `alignment.MinSimilarity` alike (0.4)". `docs/promote-graph-alignment.md:103` "at least half alike (similarity 0.5)" → "with similarity at least `alignment.MinSimilarity` (0.4)". `docs/ideas/coverage-aware-similarity.md:7` "(0.5)" → "(0.4)". `docs/deployment-plan/spike-11/name-candidates.md:10` "(0.5)" → "(0.4 since S11-43; 0.5 before)". Nothing else in those sentences changes. The case 1 table in `name-candidates.md` ("Candidate at 0.5") stays: it is the diagnosis. `completed.md`'s S11-16 paragraph stays: it is history | Working rule: each PR updates the docs it makes false. These four and the inventory row are every live statement of the number in `docs/` outside the spike archive, the PR plans, `completed.md` and the map's own scope entry | Rewriting the coverage idea's paragraph about Skip `weak` — S11-47 owns that |
| D-6 | **A new fixture helper `givenVals(surname string, given ...string) subjectcompare.Values`** *(new)* in `core/graphalign/align_test.go`, directly after `surnameVals`. Doc: `// givenVals is a name of the given parts in order, then the surname when it is not empty.` It builds one `vocabkey.PropertyName` value through `nameOf(&namevalue.Value{Form: form, Parts: parts})`, where `form` is the given words then the surname joined by single spaces, given parts typed `namevalue.PartTypeGiven` at `Idx` 0 up to len(given) − 1 and the surname typed `namevalue.PartTypeSurname` last | A Subject with a given part and no surname cannot be built with `nameVals` or `propertycomparetest.Name` (§4). One package-level helper serves T-1 and T-2, and S11-45 / S11-46 (Des against Desiree, given only) next | Changing `nameVals`' signature — about 100 call sites in `align_test.go` and `invariants_test.go`; copying `TestPartialNamesAreWeak`'s closures into two more tests |
| D-7 | **Order: tests first, then the number.** STEP-1 adds T-1 and T-2 and shows they fail at 0.5; STEP-2 changes the number and they pass | Proves the tests test the gate | — |

## 6. Design

### Codebase fit

Searched: `rg -n "func .*Vals\(" core/graphalign/*_test.go`, `rg -n "PartTypeGiven" core/graphalign core/valuecompare --glob '*_test.go'`, `rg -n "MinSimilarity" core docs`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `givenVals` *(new)* | test helper, `core/graphalign/align_test.go` | `nameVals(given, surname)`; `surnameVals`; `TestPartialNamesAreWeak`'s local `full` | `nameVals` always writes one given and one surname; changing it touches every caller. `full` is a closure local to one test (D-6) |
| `TestCandidates/AC-1.1 one given name reaches a record with three more names` *(new)* | subtest, `candidates_test.go` | the other `TestCandidates` subtests | Extends the existing table of subtests; no new top-level test |
| `TestOneGivenNameIsOfferedAFullerRecord` *(new)* | test, `align_test.go` | `TestIsolatedSubjectScoresOnTheWalkScale` | That table asserts `Target` / `Assessment` / `Reason` and no alternatives; adding a row would assert the unplaced target S11-47 changes (D-4) |

**Builds on.** The registry's alignment section and its inventory (`add-match-rule` "Add a knob" steps 2–3, for a changed value). The precedent `bdbfabfa`.
**Deletes.** None — nothing becomes dead. The number changes in place.
**Next adds.** S11-46 and S11-45 build Des / Desiree Subjects with `givenVals("", "Des")`. S11-47 changes unplaced rows' targets without touching T-2 (D-4). If dogfood after this PR shows a fuller record still missing, the curve change is a `valuecompare` name rule, not another gate move.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/matchregistry` | `defaultAlignment`: `MinSimilarity` 0.5 → 0.4 with D-1's comment |
| `core/graphalign` | Tests only: `givenVals` (D-6), T-1, T-2 |

### FFI

None — the wire is unchanged; the alternatives list already carries any candidate.

### macOS

None — `draftTarget` already opens an unbanded row unset and lists its alternatives in the menu.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Tests that pin the gate (red)**
Files: `core/graphalign/align_test.go`, `core/graphalign/candidates_test.go`. Skill: `add-match-rule`, rule `go-tests.mdc`.
Do: add `givenVals` (D-6). Add T-1 as a subtest of `TestCandidates` and T-2 as a new top-level test, exactly as §8 says.
Check: `go test ./core/graphalign/ -run 'TestCandidates|TestOneGivenNameIsOfferedAFullerRecord'` fails, and only on T-1's "is a candidate" assertion and T-2's alternative assertion (the gate is still 0.5). Every other subtest passes. If T-1's similarity assertion (2 / 4.5) fails, hit the stop condition.

**STEP-2 — Lower the gate (green)**
Files: `core/matchregistry/alignment.go`, `docs/matching-model.md`. Skill: `add-match-rule`.
Do: D-1 in `defaultAlignment`; `gofmt -w core/matchregistry/alignment.go`. D-2 in the knob inventory row.
Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/subjectcompare/... ./core/subjectprofile/... ./core/segmentcompare/... ./core/kinship/... ./core/shortlist/... ./core/graphalign/... ./core/claimevidence/... ./core/matchregistry/...` passes (the skill's Check), T-1 and T-2 included. `rg -n "MinSimilarity: 0.4" core/matchregistry/alignment.go` prints one line; `rg -n "alignment.MinSimilarity\` \| 0.4" docs/matching-model.md` prints one line.

**STEP-3 — Docs that state the gate**
Files: `docs/matching.md`, `docs/promote-graph-alignment.md`, `docs/ideas/coverage-aware-similarity.md`, `docs/deployment-plan/spike-11/name-candidates.md`.
Do: the four edits of D-5, word for word.
Check: `rg -n 'half alike|MinSimilarity` \(0\.5|similarity 0\.5\)' docs/matching.md docs/matching-model.md docs/promote-graph-alignment.md docs/ideas/coverage-aware-similarity.md docs/deployment-plan/spike-11/name-candidates.md` prints nothing.

**STEP-4 — Whole suite**
Files: none.
Do: run `scripts/check-pr.sh`.
Check: it prints `check-pr: PASS`. No existing test was edited (`git diff --stat main -- '*_test.go'` lists only `align_test.go` and `candidates_test.go`, and `git diff main -- core/graphalign/*_test.go` shows additions only).

**STEP-5 — Docs**
Files: `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-43-candidate-gate.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Offer a fuller record to a Subject with one given name by lowering the candidate gate to 0.4."). STEP-1 to STEP-3 are that one commit; STEP-5 is a docs commit ("Record S11-43 as landed."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/graphalign/candidates_test.go` `TestCandidates` / `AC-1.1 one given name reaches a record with three more names` *(new subtest)* | Probe `givenVals("", "Amy")`. Handles `PER-1` = `givenVals("Long", "Amy", "Lee", "Anne")`, `PER-2` = `givenVals("Long", "Beth", "Amy")`, built as `graphalign.Handle{ID: id("h-" + ref), Ref: ref, Kind: "person", Values: …}`. Metas `nameMetas()`, `cfg` the subtest's shipped `cfg`. (a) `subjectcompare.Compare(vocabkey.Person, probe, PER-1's values, nameMetas(), cfg.Subject, cfg.Value).Similarity` is 2 / 4.5 within 1e-9 and ≥ `cfg.Alignment.MinSimilarity`. (b) `refs(graphalign.Candidates("person", probe, handles, nameMetas(), cfg))` equals `["PER-1"]` | AC-1.1, AC-1.2, D-1 |
| T-2 | Go | `core/graphalign/align_test.go` `TestOneGivenNameIsOfferedAFullerRecord` *(new)* | Layer: one Subject `{ID: id("s-amy"), Ref: "PER-S", Kind: "person", Values: givenVals("", "Amy")}`, `Metas: nameMetas()`. Canon: one handle `{ID: id("h-amy"), Ref: "PER-1", Kind: "person", Values: givenVals("Long", "Amy", "Lee", "Anne")}`. `row := rowBySubject(alignWith(layer, canon, nil, matchregistry.Default().AlignConfig()), id("s-amy"))` → `row.Target != graphalign.TargetHandle`, `row.Assessment == graphalign.AssessNone`, `len(row.Alternatives) == 1`, `bytes.Equal(row.Alternatives[0].HandleID, id("h-amy"))`, `row.Alternatives[0].Assessment == graphalign.AssessNone`. A failure prints the row with `%+v` | AC-1.3, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs (AC-1.4). Most likely to break if the analysis is wrong: `core/graphalign` (`align_test.go` rows asserting `TargetNew` / `ReasonNoMatch` with no alternatives, such as `TestIsolatedSubjectScoresOnTheWalkScale` and `TestLifespanRulesOutAPerson`), `core/database/promotealign` (goldens, `TestLargeSourceBudgets`), `core/matchregistry` (`TestKnobInventoryMatchesRegistry`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: AC-1.1, AC-1.2 → T-1; AC-1.3 → T-2; AC-1.4 → check-pr; AC-1.5 → manual check
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored; no existing test was edited
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — n/a, no copy
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no chrome
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status Landed #N and Plan link; completed.md: one short paragraph of what shipped, newest first
- [ ] Brief moved to design/archive/ and design/README.md updated — n/a, no brief
- [ ] Domain docs updated: docs/matching-model.md inventory row (D-2), docs/matching.md, docs/promote-graph-alignment.md, docs/ideas/coverage-aware-similarity.md, name-candidates.md (D-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person whose name is entered as given `Amy`, given `Lee`, given `Anne`, surname `Long`.
2. Add a Source whose Citation records a person with the name given `Amy` only (no surname part). Open Promote for that Source.
3. See: the "Amy" row opens unset, and its target menu lists "Amy Lee Anne Long" (with no band). Choosing it files the Source person on her.
4. On the same page, a row for a Source person given `Beth` surname `Long` does not list "Amy Lee Anne Long".

Steps 1–3 are AC-1.5. This makes slice 10's Check line "'Amy' is offered 'Amy Lee Anne Long'" true.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| More candidates per Subject slow Promote on a large Source | `TopK` (5) and `TieCap` (20) still bound each Subject's candidates; `TestLargeSourceBudgets` runs in check-pr and passed at 0.4 on 2026-10-27 |
| A look-alike sharing only a surname now becomes a candidate | No: a differing first given name is a conflict (similarity 0, `Beth Long`), and a record that only shares a later given name stays under 0.4 (`Beth Amy Long` 0.375, T-1) |
| A row that opened with an empty menu now lists an unbanded alternative | Intended (working rule "The candidate gate drops to 0.4"); the row still opens unset |

**Stop and report if:**

- `subjectcompare.Compare` for T-1's pair does not return similarity 2 / 4.5 at the base (the name weights or step 9 changed since planning).
- Any test or golden other than T-1 and T-2 fails at 0.4 — name the test and its output; do not edit it.
- `TestLargeSourceBudgets` goes over budget at 0.4.
- `core/matchregistry/alignment.go` no longer holds `MinSimilarity` in `defaultAlignment`, or `docs/matching-model.md` has no `alignment.MinSimilarity` row.
- In the manual check, the "Amy" row opens on a target other than unset (S11-47 may have landed first: report what it opens on).

## 12. Open questions

None.

## 13. PR

**Title:** S11-43 — Lower the candidate gate
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-43-candidate-gate.md`.
