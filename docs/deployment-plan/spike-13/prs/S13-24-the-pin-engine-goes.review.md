# Review — S13-24 The pin engine goes

**Verdict:** Fits with changes — Not ready (round 1); round 2 pending
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- **M-2, remaining audit 2 B-37 item.** `graphalign.Subject.Provenance` is always 1 (`core/database/promotealign/layer.go`), and no PR holds B-37. This plan closes B-37's wire half (both `score` fields, `via_bridge_type`, `via_role`). Recommendation: add `Subject.Provenance` to S14-20's scope, which already changes `promotealign`; or drop it as not planned.
- **M-5, provenance read in two places.** `conclusiondetails.sqlOutcomes` rebuilds `autoreconcile.Provenance` with its own copy of `autoreconciler.sqlLoadCandidates`' credibility and certainty columns. The duplication predates this PR, which only narrows both. Recommendation: log an idea file for one owner of the provenance columns before Source quality is added, rather than merging cross-package SQL in a removal PR.
- **M-6, shared proto header message.** One proto message for the four `oneof header` copies changes Swift's mappers, which this PR's working rule keeps out. Recommendation: not planned until a PR already edits those mappers (S14-16 rewrites the Promote mapping).

## Writer's own calls
- The Vitals cause "low-confidence claim" and `ReconcilerOutcomeDetail`'s claim confidence fields go here, with no brief (user, 2026-10-30).
- The router moves as `subjectcompare.Rules.RouteOf` only, with no compare over every Property (user, 2026-10-30).
- One PR, not three (user, 2026-10-30); listed under **Rules this plan bends**.
- `canonlinks.Reader.Relationships` stays, because S13-04 calls it; the map's S13-24 list is corrected in STEP-8 (D-12).
- `identityclaims` stops reading and writing the two claim columns here, not in S13-25 (D-5); moved in the map in STEP-8.
- A held New or Skip goes out on the wire as given, where Swift drafted it unset; unobservable, because the page keeps a decided row (D-8).
- `Alternative.Score` stays for the goldens (X-6, D-9).

## Round 1
### M-1 · major · `Row.Start` returns a bare tuple; S13-30 adds a status
**Answer:** Fixed. D-8 now has `graphalign.RowStart{Target, HandleID, Ref}` and `Row.Start() RowStart`; Next adds names S13-30's `Status` field.

### M-2 · major · `via_bridge_type`/`via_role` Out line points at no owner
**Answer:** Fixed in part. D-2 reserves 17 and 18; D-9 drops the handler's two lines. `Subject.Provenance` → To user.

### M-3 · major · `Route` repeats `Rules.Role`; "pooled" means two things
**Answer:** Fixed. D-11's `Route` is `{Lead, Role, Meta}`; T-4 and the `add-match-rule` sentence match.

### M-4 · major · Prose comments made false are not in the plan
**Answer:** Fixed. D-4, D-7 and D-12 name each comment; §4 has a Prose row; STEP-8 adds a case-insensitive prose `rg` over `core api` that leaves out the packages S13-25 owns.

### M-5 · major · Provenance columns read in two SQL copies
**Answer:** To user (see above). It is existing duplication that this PR narrows; it adds no copy.

### M-6 · minor · Shared header: Go switches duplicated, proto message "not planned" with no reason
**Answer:** Fixed for Go: one `headerFor(kind, id)` in `promote_align.go` (D-8, Codebase fit). The proto message → To user, with the reason in Out.

### M-7 · minor · `PromoteReasonCode` holds only move codes
**Answer:** Fixed. D-9 renames it `PromoteMoveReason`.

### X-1 · blocker · `deleteimpact` fixtures build pins through `PinSets`
**Answer:** Fixed. STEP-2 lists `conclusion_release_test.go` and `pins_e2e_test.go`; their fixtures pin through `identityclaims.PinTx` in the same `writes.Call`; STEP-2's check runs `./core/database/deleteimpact/...` and names the three pin tests.

### X-2 · blocker · STEP-4 breaks `autoreconciler_test.go`'s `grades` loop
**Answer:** Fixed. The loop's grade lines go in STEP-4, whose check builds `./core/database/autoreconciler/...` and `./core/database/deleteimpact/...`.

### X-3 · blocker · Other `promote` tests that use deleted fields
**Answer:** Fixed. `rg 'PinSets|\.Pins\b'` over the tests finds only `TestSaveBatchPinsANeighborRecord` in `batch_test.go`; STEP-2 now says so.

### X-4 · major · `dispatch_test.go` row and "files a found record"
**Answer:** Fixed. STEP-1 deletes the dispatch row, rewrites "files a found record" without `DefaultPins` and pins, and moves `newFoundRecordFixture` (type renamed `foundRecordFixture`) into `promote_align_test.go`, which `TestProposeHoldsAFoundRecord` uses. §4's handler-tests row lists every fixture.

### X-5 · major · `rg` checks hit `sqlBindingOrder` and the onboarding test
**Answer:** Fixed. AC-4.1 and STEP-3 use `\.BindingOrder\(` and `ListClaimConfidenceGrades`, and exclude `api/proto/**`; STEP-8 uses `Vocabulary\.BindingOrder|BindingOrder\(\)`.

### X-6 · major · Golden assertions on deleted fields
**Answer:** Fixed. `Alternative.Score` stays; STEP-6 says only the term on a deleted field goes, and `Assessment != AssessNone` becomes `Score >= cfg.Alignment.Bands.Weak`.

### X-7 · major · T-8 has no weak-row fixture
**Answer:** Fixed. T-8 adds `TestProposalProtoStartsAWeakRow`, built like `TestProposalProtoSendsMove`, with a literal `graphalign.Proposal` and a promoted person for the header.

### X-8 · minor · `Start` inputs
**Answer:** Fixed. D-8 defines a held row (`Target` is `TargetNew` or `TargetSkip`) and the header lookup (`addHeaderID` takes the started id). An empty menu with a `.handle` target cannot occur: `menu(for:)` lists `proposal.handleID` first whenever the target is `.handle` and the id is set.

### X-9 · minor · Smaller gaps
**Answer:** Declined with evidence. `sqlOutcomes` selects nothing from `ic` except through `cg` (`conclusiondetails.go` lines 130–152); `inventory_test.go` needs no edit; `vocabkey.PropertyLifespanSpan` has no lead in `matchregistry/subject.go`, and `end_date`/`start_date` are the event `date` period there.
