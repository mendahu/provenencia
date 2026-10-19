# Ship log — S11-14 Segment compare

**Plan:** [`S11-14-segment-compare.md`](S11-14-segment-compare.md) · **Branch:** `s11-14-segment-compare` · **PR:** —
**State:** implementing (round 0 resumed after the user's D-11 call)

## Round 0 — implement
Head `97ea282e` (local, not pushed; STEP-4/5 uncommitted in the working tree). Gate not run.

**Stopped on:** D-11 (c) / §11 stop 1. Under D-5's numbers, Other links score `0.25 + 0.25 × neighbor evidence`, so two named (c) tests lose their Target:
- `TestProposeRelationshipFromEitherEnd`: unnamed end `f.sexOnly` has own 0; parent/child link scores 0.25 + 0.25 × 2.5 = 0.875 < `AcceptScore` 2 → Skip `weak` (was 2.75 + own at fan-out 1).
- `TestGracieObituaryGolden`: burial event has only `event_type`; `participation·subject·burial` is Other → 1.375 → Skip. A date on both sides passes, but that reshapes a (c) fixture.

Also failing under (a), reshape allowed but not done: `TestProposeInverseKinshipTermsCorrespond`, `TestProposeOverlappingBoundShowsTheDateAndResembles`. Already reshaped under (a): `TestNeighborCreditStopsAtOneHop` (person → birth (One) → place via location).

Done and committed: STEP-1 `1a7e6316`, STEP-2 `3208073c`, STEP-3 `97ea282e`. STEP-4/5 built and vetted, uncommitted (D-13 one commit). Not done: STEP-6, STEP-7, docs.

Plan's risk row (§11) anticipated weaker spouse/parent links but assumed the goldens' Targets survive; they do not. Taken to the user.

**User call (2026-10-19):** keep D-5's numbers, change no knobs; the two (c) tests assert the new behavior and carry a `// dogfood:` mark for the Spike 11 matching dogfood. Written into the plan under §5 after D-13. The other (a) failures are reshaped under D-11 as written. Resumed the round-0 implementer.
