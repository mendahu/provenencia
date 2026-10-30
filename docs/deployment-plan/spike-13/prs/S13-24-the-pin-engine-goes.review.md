# Review — S13-24 The pin engine goes

**Verdict:** Fits with changes — Ready with changes (every blocker Fixed; To user answered 2026-10-30; Planned)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet), fresh each round

## To user
Answered 2026-10-30:
- **M-2, `graphalign.Subject.Provenance`:** goes in Spike 14's S14-20 (its map row and scope entry say so).
- **M-5, provenance read in two places:** done in this PR: D-13 gives `autoreconciler` the one copy (`ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest`), and `conclusiondetails` reads through it; AC-2.4, T-12.
- **M-6, shared proto header message:** goes in Spike 14's S14-16, which rewrites the Promote mapping (its scope entry says so).

## Writer's own calls
- The Vitals cause "low-confidence claim" and `ReconcilerOutcomeDetail`'s claim confidence fields go here, with no brief (user, 2026-10-30).
- The router moves as `subjectcompare.Rules.RouteOf` only, with no compare over every Property (user, 2026-10-30).
- One PR, not three (user, 2026-10-30); listed under **Rules this plan bends**.
- `canonlinks.Reader.Relationships` stays, because S13-04 calls it; the map's S13-24 list is corrected in STEP-8 (D-12).
- `identityclaims` stops reading and writing the two claim columns here, not in S13-25 (D-5); moved in the map in STEP-8.
- A held New or Skip goes out on the wire as given, where Swift drafted it unset; unobservable, because the page keeps a decided row (D-8).
- `Alternative.Score` stays for the goldens (X-6, D-9).
- `TestUnpromoteKeepsOtherMembersPins` is deleted rather than re-fixtured: Unpromote releases pins through `deleteimpact.ReleaseClaim`, whose pin tests stay until S13-25 (X-11).
- `TestProfilesLiveWithTheMatchIndex` reads the match index after a proposal instead of an evidence read (X-10); if a proposal keeps no profiles, the implementer stops (it is a test of the index, which Propose builds).

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
**Answer:** To user; the user chose this PR (2026-10-30). Fixed: D-13.

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

## Round 2
### M-8 · blocker · `Route.Role` reads `RoleLead` for a property with no lead
**Answer:** Fixed. D-11 adds only `subjectcompare.Rules.Meta`, built on `Role` like `Lead`; no `Route` type. T-4 is `TestMeta`; the `add-match-rule` sentence and the map scope entry match.

### M-9 · blocker · STEP-5 and STEP-8 sweeps match code this PR may not touch
**Answer:** Fixed (with X-13, X-14). STEP-5's check is case-sensitive on the names this PR removes; STEP-8's checks exclude `core/database/migrations/**` and S13-25's paths, `core/database/effects/**` included.

### Round-1 answer notes from M (round 2)
- M-2: `Via.Key` is kept as the walk's edge the goldens observe; D-9 says so.
- M-6: the row and alternative oneofs are distinct wrapper types, so one `headerFor` would still branch per caller. The Go helper is withdrawn (Declined with that evidence); the proto message stays To user.
- M-7: `add-match-rule` line 78 is rewritten for move reasons only (STEP-8); `swift_vocabulary_test.go`'s unrecognized-case subtest is in D-9.
- X-1 to X-4: the moved fixture's comment is rewritten (STEP-1). The two `deleteimpact` fixtures each re-pin with a short `PinTx` loop; a shared helper is not worth it for tests S13-25 deletes.

### X-10 · blocker · `filedBirth` lives in the deleted `evidence_test.go`
**Answer:** Fixed. D-10 moves `filedBirth` to `promotealign_test.go`, drops `warmFixture.request` and `fixture.evidence`, says what each memo test keeps, and drops `TestMemoIsNotMutated`'s `Members` term.

### X-11 · blocker · `unpromote_test.go` uses `pins_test.go`'s fixture
**Answer:** Fixed. STEP-2 deletes `TestUnpromoteKeepsOtherMembersPins` (Writer's own calls) and lists `unpromote_test.go`.

### X-12 · blocker · `identityclaims_test.go` reads `cl.ConfidenceGradeID`
**Answer:** Fixed. STEP-4 drops the term and the then-unused `claimconfidencegrades` import.

### X-13 · blocker · STEP-5's `rg` cannot pass
**Answer:** Fixed. Case-sensitive pattern on the removed names only.

### X-14 · blocker · STEP-8's `rg` checks cannot pass
**Answer:** Fixed. Migrations and S13-25's paths are excluded; the `autoreconciler_test.go` line-918 comment is edited in STEP-5; `interpretation-layer-data-model.md` line 693 is in STEP-8.

### Smaller round-2 points
**Answer:** Fixed. §4's proto row lists 17 and 18; STEP-6 says every `row(…)` call drops `assessment:`. The unused `observations` import in STEP-1 is left to the compiler.
