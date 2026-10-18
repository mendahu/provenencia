# Review — S10-60 The gallery reads only gallery records

**Verdict:** Fits — Ready with changes (round 2: no new maintainability finding; the one executability blocker, X-15, was a stop-condition command and is fixed)
**Rounds:** 2. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user
- None.

## Writer's own calls
- D-1: the gallery admits a record by its own outcome reason through the existing `Reason.InGallery`; no second method, no SQL literal.
- D-2: `InGallery` stays `kept` / `weak`; `TestReasonInGallery` unchanged.
- D-3: no `CacheVersion` bump and no migration; only the read changes.
- D-4: a pointer stored before the fix at a skipped record stays in the row; readers fall back to the earliest.
- D-5: the pointer is captured from every scanned row before the reason check.
- D-6: the gallery rule is one paragraph in `conclusion-layer-data-model.md` §6, linked from the reconciliation doc.
- D-7: one exported `observationstest.SaveNegativeLikeness` and one `fixture.negativeTwin` builder.
- The file name and branch keep the `gallery-member-records` slug; only the title and wording changed (M-1).
- Version bump none, as every Spike 10 / 11 PR since #402.

## Round 1

### M-1 · major · "Member records" gives "member" a third meaning
The reconciler's members (`ReconciledValue.ObservationIDs`, `row.members`, which include provisional and denied entries in a hidden row) and handle membership (§2.2, S11-19's "member's Observation") already use the word.
**Answer:** Fixed — title, §1, §6 Go core comments, D-1, D-4, D-5, §11, STEP-4 doc paragraph, and the PR title now say **gallery records**; §1 says "member" keeps its handle meaning. The file name and branch slug stay `gallery-member-records`: renaming the paths churns the map link, the review-log link, and the branch for no reader gain. The map row title should change to match.

### M-2 · major · The negative-likeness save is copied three times across three packages
**Answer:** Fixed — D-7 and Codebase fit: `observationstest.SaveNegativeLikeness` beside `SaveLikeness`, both through a private `saveLikeness(…, polarity)`. T-1, T-4, T-5, T-6, T-7 and the existing "stronger negative denying the pointed image falls back" subtest use it; the planned private `fixture.saveNegative` is gone. STEP-1 adds it.

### M-3 · minor · The T-1 catalog is rebuilt in several subtests
**Answer:** Fixed — `fixture.negativeTwin() (h []byte, n, a, b observations.Observation)` beside `twoLikenesses` (Codebase fit, D-7); T-1 and T-5 call it. T-4 adds N inside the existing "refused" setup instead, because that subtest's other cases are built on `twoLikenesses`.

### M-4 · minor · `fixture.grade` repeats `autoreconciler_test.go` `fixture.credibility` under a name that means a claim-confidence grade there
**Answer:** Fixed for the name — the helper is `fixture.credibility(src sources.Source, key string)`, the autoreconciler signature. The shared Source-credibility helper for the five inline copies is Declined for this PR: no ideas file covers test helpers (`docs/ideas/shared-list-helpers.md` is about list helpers), so it is an Out line logged here.

### M-5 · minor · The hidden-row reason Out line has no home
`pipeline.go` hidden rows take `reason: e.reason` from the first entry and count every entry's Source in `support()`, against `conclusion-reconciliation.md` "Provisional members… never influence the answer".
**Answer:** Fixed — STEP-4 adds a "A hidden value's reason and support" Still-open bullet to `docs/ideas/depictions-and-likenesses.md`, and the Out row points at it.

### X-1 · blocker · T-2 also fails before the fix; STEP-2 and Stop condition 3 name only T-1 and T-3
**Answer:** Fixed — STEP-2 says all three fail before the query change; the Stop condition is "Any of T-1 – T-3 already passes".

### X-2 · blocker · The T-1 catalog is reused by T-4 – T-7 but no builder is named
**Answer:** Fixed — `fixture.negativeTwin` (T-1, T-5); T-4's setup spelled out inside "refused"; T-6 and T-7 spell out their saves through `observationstest.SaveNegativeLikeness` with each package's fixture (D-7).

### X-3 · major · AC-1.3 says "the same catalog" but T-7 builds one Subject on one Source
**Answer:** Fixed — AC-1.3 now states the single-Subject, single-Source shape and why it is deliberate; a Stop condition covers N denying A there.

### X-4 · major · The DoD gate cannot be shown on Linux
**Answer:** Declined — `.agents/skills/ship-pr/SKILL.md` "Before you start" requires the user's Mac (`scripts/check-pr.sh` needs `xcodebuild`), and every Spike 11 plan carries the same DoD line. The DoD line now says "on the user's Mac".

### X-5 · major · Precedent points at the session scratchpad
**Answer:** Fixed — the scratchpad reference is gone. Precedent names `autoreconciler_test.go` `fixture.credibility` and `TestDepictionsReconcileByImage` "a provisional member" (the `create_identity_claim` call), and T-1 – T-3 spell out the save, grade, promote, and claim order.

### X-6 · major · Status values, commit grouping, and PR-number timing are unspecified
**Answer:** Fixed — STEP-4 makes only the three doc edits; the map row and plan Status go to `Landed #N` in the post-PR commit "Record S10-60 as landed in pull request N." (`implement-pr` Round 0 step 8). The DoD line says the same. The Landed note follows the Spike 11 plans (`S11-06` `## Landed`).

### X-7 · major · `go-tests.mdc` (table-driven) conflicts with the narrative subtests
**Answer:** Fixed — the Rules line says the new cases follow each file's existing shape: narrative subtests, plus one row in the "refused" table.

### X-8 · minor · Helper signatures and comment text left open
**Answer:** Fixed — §6 Go core gives `SaveNegativeLikeness`'s signature and doc comment, `saveLikeness`'s signature (D-7), the package doc comment, the `ObservationIDs` comment, and the `InGallery` comment in ASCII quotes.

### X-9 · minor · STEP-1 leaves a false comment, and a comment-only commit has no researcher subject
**Answer:** Fixed — the `InGallery` comment moves into STEP-2; STEP-1 is the fixture step, and STEP-1 and STEP-2 are one commit.

### X-10 · minor · Stop conditions 1 and 2 have no check that can detect them
**Answer:** Fixed — both are now `rg` commands, and the circular "member's stored reason" wording is gone.

### X-11 · minor · T-5 order and read path
**Answer:** Fixed — STEP-3 and T-5: the SQL runs after every promote, then `f.gallery(h)`.

### X-12 · minor · STEP-4 doc wording and link check
**Answer:** Fixed — STEP-4 gives the full resulting sentence and the link target; the check greps the §6 heading the anchor names.

### X-13 · nit · Status is Draft
**Answer:** Declined — `plan-pr` step 8: a plan is `Planned` only after review-plan; the writer sets it once this log is answered.

### X-14 · nit · The map links a plan that did not exist yet
**Answer:** Declined — all four plans now exist on the branch.

## Round 2

Fresh reviewers got the plan and this log. Maintainability: Fits, every round-1 answer holds, no new blocker. Executability: Not ready on X-15 only.

### Maintainability leftover · note · §1's one-line definition of "member" is too narrow
"Member" is used for provisional members too (`autoreconcile.Candidate.Provisional`, conclusion-reconciliation.md "Provisional members").
**Answer:** Fixed — §1 now defines a member as a Subject linked to the handle by an Identity Claim, accepted or provisional.

### X-15 · blocker · Stop condition 2's command cannot detect what it names
`CREATE TABLE auto_reconciler_outcomes` (000040.sql:21) and `reason` (:25) are on different lines, so the single-line pipe prints nothing on a true base.
**Answer:** Fixed — §11 uses `rg -n -A8 'CREATE TABLE auto_reconciler_outcomes' core/database/migrations/*.sql | rg reason` and quotes the line it prints at the base (verified: `000040.sql-25-`).

### X-16 · major · Stop condition 1's `rg` hits unrelated name tests
**Answer:** Fixed — §11 says the condition fires only on a hit in `core/database/likenesses` or `TestDepictionsReconcileByImage`, and names the three expected name-property hits (lines 527, 579, 1018).

### X-17 · major · T-2 and T-3 both need a provisional claim, and no helper is named
**Answer:** Fixed — `fixture.provisional(s subjects.Subject, onto []byte)` is a Codebase fit row, added in STEP-1, used by T-2 and T-3; it adds the `identityclaims` import.

### X-18 · minor · `credibility` return shape and the `ObservationIDs` comment placement
**Answer:** Fixed — `credibility` calls `must(f.t, err)` and returns nothing; the `ObservationIDs` comment stays trailing and keeps "[0] is the first recorded".

### X-19 · minor · STEP-4's Files line contradicts its Do line
**Answer:** Fixed — STEP-4 Files lists the three docs only.

### X-20 · minor · Commit grouping for STEP-3 and STEP-4
**Answer:** Fixed — STEP-1 – STEP-3 are one commit ("Leave a negative or provisional record of a picture out of the Person's gallery."), STEP-4 is "Document which records make a Person's likeness gallery.", then the post-PR commit.

### X-21 · minor · Manual check cannot be followed as written
**Answer:** Fixed — step 1 puts the same photo file on Album and Register; step 3 names the Promote flow (new Person from Register, then onto that Person).

### Nit · D-7 omits T-4
**Answer:** Fixed — D-7 lists T-4.

### Nit · Status Draft and "1 round"
**Answer:** Fixed — Status `Planned`, Review line "2 rounds".
