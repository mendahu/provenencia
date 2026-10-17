# Review — S11-01 One-page matching model

**Verdict:** Fits with changes — Not ready, then fixed. **Rounds:** 1. **Reviewers:** maintainability (fresh agent), executability (fresh agent).

## To user

None.

## Writer's own calls

- New `docs/matching-model.md` instead of rewriting `docs/matching.md` (D-1). Later PRs fold or delete the old docs (D-5).
- Estimated starting values for knobs that don't exist yet (§6 inventory, rows marked "estimate").
- The drift guard compares keys only; the Go registry is the only source of a current value (D-6, after M-4).

## Round 1

### M-1 / X-1 · blocker · STEP-5 trips its own stop condition
Particles, letter folds, the Western part roles, date periods and several identifiers fit neither table.
**Answer:** Fixed. Inventory rows added for `value.Words.Particles`, `value.Words.Folds`, `value.Names.Western.PartRoles`, `subject.Pairs.event.date` and `subject.Pairs.person.sex`. A "Not knobs" list covers identifiers and catalog data. STEP-5's check and the stop condition now name all three lists.

### M-2 / X-4 · major · Derived sources and double-listed symbols
**Answer:** Fixed. `subject.event.date.Agree` and `subject.place.toponym.Agree` are estimates from today's cold-start log-odds (≈2.1, ≈3.2, rounded). `ConflictPenalty` and `PropertyAgreement` are in the removed table as "value carried into …". `value.Dates.NearMinimum` now cites `ValueTypeScale["date"].Floor`.

### M-3 / X-3 · major · Cardinality and shortlist keys in prose
**Answer:** Fixed. They are now table rows (`segment.Cardinality.birth` / `.death`, `alignment.Shortlist.person` / `.event` / `.place`). The apparent conflict in the design doc was two meanings of "cardinality". `comparison-layers.md` now says property cardinality stays in the catalog, and bridge cardinality is a registry rule.

### M-4 · major · Every value in two places
**Answer:** Fixed in D-6. The guard compares keys only. Start is the starting value. The Go registry is the only source of a current value.

### M-5 / X-2 · minor / major · Worked example arithmetic
**Answer:** Fixed. The birth's own evidence is 2.5 (type 0.5 + date 2), the segment is 2.75, and the total is 5.25. STEP-2 checks the sums.

### M-6 · minor · Links into superseded sections
**Answer:** Fixed. D-5 makes S11-16 re-point `docs/conclusion-layer-data-model.md` L256 and `docs/conclusion-reconciliation.md` L141.

### X-5 · minor · Values in neither table
**Answer:** Fixed. The sex / gender floor, the `part_of` to-end `MinNode` and the start / end profile features each have a row.

### X-6 · minor · Banner wording
**Answer:** Fixed. STEP-6 gives the exact blockquote lines and where they go.

### X-7 · minor · STEP-7 map edit vague
**Answer:** Fixed. STEP-7 gives the status, the `completed.md` paragraph and the plan status. The map row now links this plan.

### X-8 · minor · "speaking pair" missing from the glossary
**Answer:** Fixed. It is added to D-3.

### X-9 · nit · Not reviewed yet
**Answer:** Done. This log; status Planned.
