# Review — S11-25 Claim evidence RPC and pin policy

Plan: [`S11-25-claim-evidence-rpc.md`](S11-25-claim-evidence-rpc.md). Lint: `check-plan-docs.py` clean before round 1.

**Verdict:** round 1 — Fits with changes / Not ready; every finding answered below. Round 2 pending.
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- None open. Four product calls were answered before review (2026-10-21): relationships keep a "Through" card with a header checkbox; a collapsed line is one row listing every member; no argument is drafted, in Go or Swift; pins the evidence no longer offers are dropped at Done.

## Writer's own calls

- **Already filed Subjects count as neighbors without being named** (D-6, AC-1.4). The neighbor rule moves into `mergeFixed`, which `Propose` already uses. `Evidence` used to count only the neighbors a request named; now a filed neighbor counts on its filed handle. S11-19's `TestEvidenceOnlyMappedNeighbors` is unchanged, because its absent neighbor is not filed.
- **`promote.Save` writes no pins** (D-5). Only tests set `Input.Pairs`, and those tests move to `SaveBatch`.
- **Pin ids are opaque Go strings** (`l:…`, `b:…`, D-3). Swift stores them and sends them back, and never parses them.
- **The sheet refetches its evidence on every open** (D-13). It does not cache across opens.
- **One new string,** the evidence error title (D-14).

## Round 1 (2026-10-21)

Maintainability: **Fits with changes** (M-1 blocker). Executability: **Not ready** (X-1 and X-2 blockers).

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | The neighbor rule lived in four places: `ResolvePins`, Swift `evidenceNeighbors`, `mergeFixed`, and `repropose`'s loop | **Fixed.** D-6: `evidence` reads its neighbors through `mergeFixed`. D-12: Swift sends the rows Done files, through `fixed(for:)`, which `repropose` also uses. AC-1.4 and T-4's subtest cover this, and the old stop condition is gone |
| M-2 | major | `ResolvePins` reloaded `Evidence`'s inputs for each row, and a production hook existed only to prove it did not | **Fixed.** D-6: `Evidence` is `loadInputs` plus a private `evidence`, so `ResolvePins` loads once per page. D-16 (the hook) is removed. T-5 uses the existing query counter |
| M-3 | major | The pin-id rule was exported three ways, and Swift re-derived the checked and offered sets | **Fixed.** D-3: the handler sends `PinOffers` as `offers` (id, checked). D-11: Swift copies them without logic |
| M-4 | major | `Save` kept a pin path that only tests used | **Fixed.** D-5: `Input.Pairs` is deleted, and those tests pin through `SaveBatch` (STEP-3) |
| M-5 | major | `add-match-rule` does not teach the Swift mirror for a code | **Fixed.** STEP-10 updates the skill. The `Causes()` functions are **declined**: hand-listing follows the precedent of the reasons row (`graphalign.Reason*`), and an enumerator would be one more hand list in Go. It is an Out line |
| M-6 | major | The proto-to-`Fixed` parse would be copied | **Fixed.** D-1: `fixedFromProto`, used by both handlers |
| M-7 | minor | `Pin` named two different things | **Fixed.** D-5: the write unit is `promote.PinSet` |
| M-8 | minor | Pin ids used hex, a second id encoding on the wire | **Fixed.** D-3: ids are built with `database.UUIDString` |
| M-9 | minor | `outcomeTone` and the `"conflict"` string comparison survive | **Fixed.** D-10 and D-14 retype them; the second AC-4.2 `rg` guards it |
| M-10 | minor | The loading and error flags allow impossible states | **Fixed.** D-13: `SheetEvidence` enum |
| M-11 | minor | Value-to-text code is spread over several places | **Logged.** Out line; STEP-10 adds a known-gaps line to `ideas/matching-test-suite.md`; the `ComparisonValueDisplay` rename goes to S11-20 |
| X-1 | blocker | AC-4.2's `rg` used `\|`, which ripgrep reads as a literal pipe | **Fixed.** The pattern uses bare alternation. Checked: today it lists 19 files, and the second command finds 6 string switches |
| X-2 | blocker | `wireCase` captures `""`, but the plan did not say how the empty cause is handled | **Fixed.** D-10: `case "": self = .none`, listed first, and `""` added to the three cause rows' `want` |
| X-3 | major | `valuecompare.Cause` has no enumerator | **Fixed.** D-10 decides to hand-list the codes; the Out line is added |
| X-4 | major | `propose`'s destructure would leave `q` unused | **Fixed.** D-9 names line 54 |
| X-5 | major | Choices left open in the handler's wire mapping | **Fixed.** §6 FFI has a field-by-field mapping table; D-1 builds one `EvidenceText` per request; `sources` is sorted by id string |
| X-6 | major | D-6 and D-12 seemed to use different neighbor sets | **Fixed** by M-1: one set, the rows Done files plus `mergeFixed`'s filed Subjects |
| X-7 | major | No fixture for five members, and the handler package needs its own | **Fixed.** T-3 names `namedMembers`; T-7 builds three members with `personOnNewSource` and `promoteOK`. `filedBirth`'s file is corrected in the conventions |
| X-8 | major | The proto pin `date` had no Swift counterpart | **Fixed.** D-11: `GoStore` formats the display from the date and keeps no date field |
| X-9 | minor | `outcomeTone` not named | **Fixed** with M-9 |
| X-10 | minor | `NewEvidenceText` did not say which origin's value type it uses | **Fixed.** D-4: origin `provenencia` only |
| X-11 | minor | Xcode and protoc checks cannot run on Linux | **Fixed.** The DoD has the "host without Xcode" line, as S11-19 did; protoc is the implementer's existing toolchain (`generate-proto.sh`) |
| X-12 | minor | Temptations had no Out line | **Fixed.** Out lines for `Causes()` and for `ComparisonValueDisplay`. Carrying evidence across a re-proposal is already D-12's `makeRow` rule |
