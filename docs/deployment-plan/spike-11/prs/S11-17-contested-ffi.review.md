# Review — S11-17 Contested flag through FFI

**Verdict:** Fits — Ready with changes (round 2; every round-2 finding fixed)
**Rounds:** 2 (round 2 required: M-1 was a fixed blocker that reshaped D-4, D-6, D-7). **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- None open. Two product calls were answered before review (2026-10-19): the move-reason copy goes to S11-18, not S11-17; cause codes go to S11-25.

## Writer's own calls

- Flat wire fields `contested`, `move_reason`, `move_neighbor_subject_id` instead of a nested message (D-1), matching `via_*`.
- One Swift wire enum `PromoteReasonCode` for row and move reasons, mirroring Go's single `graphalign.Reason` (D-3).
- An interim `reasonText` arm wording the three move cases as `reasonEmpty` until S11-18 (D-4). It is unreachable: no row reason is a move reason.
- A decided row's contested flag refreshes on every proposal, as its band does; anchor rows are never contested (D-7).
- Wire reason fields typed as `PromoteReasonCode` in the store, which changes about a dozen literals in `FakeStore` and the Promote tests (D-6, D-8).
- The S11-18 scope entry now leaves to S11-D1 where a move shows, because a move and `Row.updated` measure different things.
- An empty neighbor id gives `.empty` for all four neighbor reasons (`via` and the three moves), today's `via` rule (round 2, M-8).

## Round 1

### M-1 · blocker · `PromoteFlow.Move` carries the wire enum, so S11-18 gets a second way to word a reason
`PromoteModel.reasonText` is the page's one wording switch over `PromoteFlow.Reason`, and its `.via` branch already names a neighbor. A `Move` holding `PromoteReasonCode` would make S11-18 write a second switch with a `default` and copy the neighbor lookup. Better: move cases on `PromoteFlow.Reason`, `Row.move: Reason?`, drop `Move`.
**Answer:** Fixed. D-4 adds `.neighborPlaced/.neighborMoved/.neighborReleased(neighborID:)` and an interim `reasonText` arm; D-7 makes `Row.move: Reason?`; `PromoteFlow.Move` removed from D-7, Codebase fit and §6 macOS; AC-3.1, AC-3.4, AC-4.3 updated. The interim arm and the unrecognized-move → `.empty` behavior are writer calls (above), not product calls: the arm is unreachable and the copy stays S11-18's.

### M-2 · major · Reason codes are parsed in the feature, where house precedent parses them in `GoStore`
`CatalogReconciledValue.reason: ReconcilerReason` and `GoStore`'s `ReconcilerReason(wire:)` parse in the store; the plan copied the enum but not the parse site.
**Answer:** Fixed. D-6 types `reason` on the wire row and alternative and `moveReason` as `PromoteReasonCode?`, parsed in `GoStore`; D-8 now types `FakeStore` and test literals; the old D-8 carve-out and its Out line are gone; AC-4.4 covers `macos/App` and `macos/ProvenenciaTests`. Steps regrouped (STEP-3 does D-3, D-6, D-4, D-8 together because they do not build apart).

### M-3 · major · "move" and "updated" mean different things, but the plan treats them as one
`Row.updated` is a Swift diff between proposals; `graphalign.Row.Move` is a release within one `Align` run.
**Answer:** Fixed. D-7's doc says so and drops "for the Updated line"; the Out line says S11-18 with S11-D1 decides where a move shows; the S11-18 scope entry in `deployment-plan.md` says the same.

### M-4 · minor · Doc comments list the reason codes again, with no guard
**Answer:** Fixed. D-6's doc comments name the type ("A row reason", "A move reason"), not the codes.

### M-5 · minor · Other Promote wire literals stay outside the vocabulary file, and their Out line has no home
**Answer:** Fixed. The Out line names `"handle"`, `"new"`, `"skip"`, `"conflict"` and `PromoteFlow.Assessment`, and homes them in S12-09, which rewrites the same Promote code.

### X-1 · blocker · STEP-7 tells me to make doc edits that are already in `deployment-plan.md`
**Answer:** Fixed before the finding arrived (the reviewer read an earlier draft). STEP-7 says the scope entries were updated at planning and must not be edited; only the map row status, `comparison-layers.md` and `completed.md` remain.

### X-2 · blocker · The plan says S11-16 is unmerged, but the repo says it landed
**Answer:** Fixed. The S11-16 branch writes its Landed status ahead of the merge; `origin/main` @ `3bd0862` lacks it. The header now says the merge into `main` is what counts, gives the check (`git log origin/main` holds #501's merge and `rows.go` sets `Flags.Contested`), and says to branch from `origin/main`, redo §4 and set status `In progress`.

### X-3 · blocker · STEP-1 cannot run in this environment, and no fallback is given
**Answer:** Fixed. New stop condition: if `protoc`, its plugins or `xcodebuild` is missing, stop; the PR is built on macOS, where `scripts/check-pr.sh` runs; no hand-edited generated files.

### X-4 · major · The Swift checks cannot run on Linux and the plan has no fallback
**Answer:** Fixed by the X-3 stop condition: Swift tests are never reported as passing without running them.

### X-5 · major · STEP-7 does not fix the `completed.md` sentence that this PR makes false
**Answer:** Declined. `completed.md`'s S11-16 paragraph records what S11-16 shipped, and "neither reaches the page until S11-17" stays true as history; S11-17's own paragraph says what changed. The S11-14 plan is landed; landed plans change only by their Landed note (`plan-pr` "After the PR lands"), and the map's S11-25 scope entry now owns cause codes, which the plan's Out line records.

### X-6 · major · T-1 (`TestProposeSendsContested`) leaves its fixture mechanics open
**Answer:** Fixed. STEP-2 names the setup to copy, how the incoming id leaves `reqFn`, and the fixed row's `Fixed` entry.

### X-7 · major · T-2 (`TestProposalProtoSendsMove`) is under-specified
**Answer:** Fixed. STEP-2 gives the `graphalign.Proposal` literal and the expected `n.String()`.

### X-8 · major · T-4 to T-6 leave out how AC-3.2 reaches "decided on e1"
**Answer:** Fixed. STEP-6 names `load`, `mapRest()`, `setTarget(subjectID: "a", token: "handle:e1")`, `merge`.

### X-9 · minor · The AC-4.4 `rg` is quoted ambiguously
**Answer:** Fixed. The command sits unescaped in STEP-3's check; AC-4.4 points at it.

### X-10 · minor · Commit grouping is left to me
**Answer:** Fixed. §7 names four commits and their subjects.

### X-11 · minor · D-4's `.via` handling is only described in prose
**Answer:** Logged; AC-4.3 pins the behavior.

## Round 2

Both reviewers confirmed every round-1 answer (X-5's decline accepted). No new blockers.

### M-6 · minor · D-6 leaves a doc comment on the alternative that lists codes
**Answer:** Fixed. D-6 rewrites `CatalogPromoteGraphAlignmentAlternative.reason`'s comment; D-1's `move_reason` comment names `graphalign.Reason` instead of listing codes.

### M-7 · minor · `PromoteFlow.move(_:)` puts a third wire-row-to-`Reason` conversion outside `Reason`
**Answer:** Fixed. D-7 adds `PromoteFlow.Reason.init?(move:)` beside the row and alternative `init`s; Codebase fit updated.

### M-8 · minor · D-4 applies two different empty-neighbor rules in one initializer
**Answer:** Fixed. D-4 states one rule (empty neighbor → `.empty`) for all four neighbor cases; AC-3.4 and AC-4.3 follow it.

### X-12 · major · AC-4.4 says "No match", but STEP-3's `rg` cannot return none
**Answer:** Fixed. STEP-3's check is two searches scoped to the reason literals; run at planning, they print exactly D-8's literals and `PromoteFlow.Reason.init`'s six cases, so after the PR both print nothing. AC-4.4 points at them.

### X-13 · minor · STEP-6 says the `row(…)` helper "stays as is", but D-8 edits its body
**Answer:** Fixed. STEP-6 says its signature stays.

### X-14 · minor · STEP-7 leaves the Landed record underspecified
**Answer:** Fixed. STEP-7 gives the order (PR first, then commit 4), the `completed.md` heading and shape (S11-16's entry), the map status and the plan status.

### X-15 · minor · GoStore placement of the new fields is implicit
**Answer:** Fixed. D-6 says "in struct order: after `viaRole:`, before `person:`".

## Refresh after S11-16 merged

S11-16 merged in #501 (`main` @ `ed45eba`), with S12-03 alongside it. The plan was rebased and §4 rechecked against `main`:

- `core/graphalign`: `RowFlags.Contested`, `Move`, `Row.Move` and the ten `Reason` constants are unchanged; the last S11-16 fix round only removed a dead nil guard in `rows.go`.
- The handlers' `uuidString` is gone; `database.UUIDString` (`core/database/uuid.go`) replaces it with the same result ("" for a non-16-byte id). D-2 now names it.
- `swift_vocabulary_test.go` gained a `CatalogList` row read by a new `listCase` regex (a raw-value enum). D-3's rejected alternative now says why reasons keep `ReconcilerReason`'s shape: a raw-value enum has no `.unrecognized`.
- Proto `PromoteGraphAlignmentRow` still ends at field 22; every Swift symbol and test helper the plan names is unchanged.
- The contested probe re-run at `ed45eba` gives AC-1's results: negative unfixed → skip / weak / contested; negative fixed → handle / decided / contested; positive → not contested.

No Decision changed, so no new review round. Status → Planned.
