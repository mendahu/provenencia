# Review — S11-19 Claim evidence query

Plan: [`S11-19-claim-evidence-query.md`](S11-19-claim-evidence-query.md). Lint: `check-plan-docs.py` clean before round 1.

## Round 1 (2026-10-21)

Maintainability: **Fits with changes**. Executability: **Not ready** (X-1, X-2 blockers).

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | `Record` / `Member` add a second value shape beside `subjectcompare.Values` and a second candidate → value conversion beside `ForSubject` | **Fixed.** D-8: `ForSubject` sets `Ref`; `Member.Values` is `subjectcompare.Values`; `Record` dropped; the incoming side is `loadLayer`'s values; `layerValues` takes a set of ids. T-13 added |
| M-2 | major | Association ids as a fourth `loadLayer` return that `propose` discards | **Fixed.** D-14: `graphalign.Bridge.Associations` |
| M-3 | major | `Evidence` copies `propose`'s setup and `mergeFixed`'s checks | **Fixed.** D-3: `loadInputs` and `checkFixed`, moved and shared |
| M-4 | major | Lifespan and likeness lines have no engine slot, and S11-20 is scoped "no engine" | **To user → answered 2026-10-21:** option 1, S11-20 owns their engine part (S10-54, depiction values over the FFI, already landed in #433). Map S11-20 entry and the plan's Out line say so |
| M-5 | minor | `memberSources` SQL in the feature package | **Fixed.** D-16: `subjects.SourcesOf` in `core/database/subjects` |
| M-6 | minor | `event_type` as the type-defining property is stated in three places | **Fixed (logged).** Out line and STEP-8: a known-gaps line in `ideas/matching-test-suite.md` |
| X-1 | blocker | STEP-4 breaks `export_test.go`'s `loadLayer` call | **Fixed** by M-2: `loadLayer` keeps three returns |
| X-2 | blocker | `Evidence` and `mergeFixed` not reconciled; invalid and ignored neighbors undecided | **Fixed.** D-3 names `checkFixed`, no anchors, empty row handle invalid, non-neighbor and New / Skip ignored; AC-2.4, AC-2.7 widened |
| X-3 | major | Stop condition forbade a test import of `matchregistry` | **Fixed.** D-1 and §11: non-test files only |
| X-4 | major | How to build `primary` for member records | **Fixed** by M-1: `layerValues` takes ids |
| X-5 | major | Segment inputs, `own`, `Points` unstated | **Fixed.** D-12 names each |
| X-6 | major | Pooled vs period-end rule ambiguous | **Fixed.** D-6 three cases; AC-1.9 points |
| X-7 | major | STEP-5 too large | **Fixed.** STEP-5 to STEP-7 |
| X-8 | major | Association node members unverified; map keyed by private type | **Fixed.** Current state row (`identityclaims.fileBridge`); map removed by M-2 |
| X-9 | minor | Key on `.Value`, side A, missing meta, invariants test, doc path, commit grouping | **Fixed.** D-5, D-6, D-10, T-12, STEP-8 paths, commit line |

## Round 2 (2026-10-21)

Fresh reviewers with the plan and this log. Both confirmed every round-1 answer. Maintainability: **Fits with changes**. Executability: **Not ready** (X-10, X-11 blockers).

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-7 | blocker | Relationship and bundle policy sits in the loader, not in `claimevidence` | **Fixed.** D-1, D-12, D-13: `claimevidence.Build` with `Neighbor`, `LayerLink`, `CanonLink`, `Bridge` inputs owns correspondence, segment state, the endpoint rule and the defining split; the loader only reads. STEP-6 and T-16 |
| X-10 | blocker | `evidence.Defining.event` cannot match the inventory walk, which reads a map of slices as one leaf | **Fixed.** D-2 and STEP-2 rename the row `evidence.Defining` (precedent `alignment.Shortlist.Words`); T-4 and §11 updated |
| X-11 | blocker | AC-1.10's `occupation` is not in `subjectcomparetest.Metas()` | **Fixed.** AC-1.10 uses `event_name`, in the fixture and unweighted |
| X-12 | major | A period `Side` builds leaves `RefB` empty | **Fixed.** D-5: no line; AC-1.12 |
| X-13 | minor | T-12 contradicts D-11's tie order | **Fixed.** T-12 allows lines tied on the incoming `Ref` to reorder |
| X-14 | minor | `cardinality` conflict never reaches a line | **Fixed.** Removed from D-7 |
| X-15 | minor | Test package, `layerValues` arguments, revision read, nil catalog, bridge dedupe index | **Fixed.** STEP-5, D-8, D-3; the dedupe index is left to the implementer as derivable |
| — | — | `check-pr.sh` needs Xcode | **Fixed.** DoD says how to report on a host without Xcode |

## Verdict

Two rounds done. No open blocker; M-4 answered by the user. Plan **Planned**.
