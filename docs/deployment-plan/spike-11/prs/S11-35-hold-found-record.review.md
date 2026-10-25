# Review — S11-35 Hold a record Promote did not offer

**Verdict:** Fits with changes — Ready with changes (all findings answered; no blocker open)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

Round 2 was not run: the one blocker (M-1) was fixed with the shape the reviewer proposed, and no Reshape was made.

## To user
- None.

## Writer's own calls
- D-3: a record already in the row's menu, picked through `hold`, is selected as offered; it is not added twice and keeps its offered sentence.
- D-7: `FakeStore` gets no code change. Its unscripted propose already holds any filed record, because it has no candidates. The map scope entry's parenthetical and the feature doc bullet that implied a fake change are corrected in STEP-5.
- D-9 / D-10: the Go tests assert neither "New with `no_match`" nor exact bands for Robnis, so S11-47 and registry tuning do not break them. Only the FFI test pins band `none`, for a different first name and surname.
- The engine claims were checked before writing by throwaway tests at `22e3878` (deleted): Robnis fixed on Robins → `handle` / `decided` / weak, contested only on a sex conflict; Polly Jones fixed on Mary Smith → `handle` / `decided` / `none` / contested, 2 evidence sections, files with 2 default pins.

## Round 1

### M-1 · blocker · `hold` takes a pre-built `Alternative`, so S11-36 has to format menu titles outside the flow
**Answer:** Fixed. D-1 is now `hold(subjectID:id:ref:header: CatalogConclusionHeader?)`; the flow builds the item through `conclusionTitle` / `conclusionSubtitle` under its own `nameOrders`, with `headerParts(_:)` *(new)* splitting the enum (D-5). AC-1.1 and AC-1.9 test the titles; STEP-5 rewords the feature doc's first Swift bullet.

### M-2 · major · `hold` encodes `"handle:" + id` only to have `setTarget` decode it, and repeats `setTarget`'s guard
**Answer:** Fixed. D-4 extracts `select(_:at:)` from `setTarget`'s body; `setTarget` and `hold` each guard once and call it. The two `"handle:"` spellings in `PromoteView` stay (Out line: S11-36 owns that menu and its sentinel token).

### M-3 · major · `foundRecordFixture` is a second evidence-fixture type beside `evidenceFixture`
**Answer:** Fixed. D-11: `newFoundRecordFixture(t) evidenceFixture`, in `promote_evidence_test.go`; T-3 calls `f.evidence(t)`.

### M-4 · minor · STEP-5 leaves a stale `FakeStore` bullet in the feature doc
**Answer:** Fixed. STEP-5 rewrites that bullet and the call-shape bullet.

### X-1 · major · `hold` must guard before it touches the menu; the order is not stated
**Answer:** Fixed. D-2 says the guard comes first; T-9 checks no menu item is added on refusal.

### X-2 · minor · T-5 and T-6 leave the test setup to inference
**Answer:** Fixed. T-5 – T-8 name `load` / `merge`, the rows and subjects, and the `handleRef` each row needs.

### X-3 · minor · T-10's fake data is underspecified
**Answer:** Fixed. T-10 gives the full `CatalogSubjectMembership` and says the first proposal opens Ada unset.

### X-4 · minor · T-1 is ambiguous for the sex `""` case and for ordering
**Answer:** Fixed. STEP-1 names the three cases, "no sex Observation on either side", and "cite before promote". The `unknown` comparison for the no-sex case was observed at the planning base.

### X-5 · minor · The plan's own Go-test rule is not honoured for T-2, T-3 and T-4
**Answer:** Fixed. D-12: T-1 is a table; T-2 – T-4 are single-case, one RPC each.

### X-6 · minor · The map row must link the plan, but STEP-5 does not say so
**Answer:** Fixed. Planning set the map row's Plan link and status `Planned` in this commit.

### X-7 · minor · A stale sentence in the feature doc is not in STEP-5
**Answer:** Fixed (same as M-4).

### X-8 · nit · T-4 cannot reuse the `apply` closure
**Answer:** Logged. STEP-2 says T-4 writes its propose and Apply inline.
