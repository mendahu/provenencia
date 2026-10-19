# Ship log — S11-14 Segment compare

**Plan:** [`S11-14-segment-compare.md`](S11-14-segment-compare.md) · **Branch:** `s11-14-segment-compare` · **PR:** #493
**State:** ready

## Round 0 — implement
Head `97ea282e` (local, not pushed; STEP-4/5 uncommitted in the working tree). Gate not run.

**Stopped on:** D-11 (c) / §11 stop 1. Under D-5's numbers, Other links score `0.25 + 0.25 × neighbor evidence`, so two named (c) tests lose their Target:
- `TestProposeRelationshipFromEitherEnd`: unnamed end `f.sexOnly` has own 0; parent/child link scores 0.25 + 0.25 × 2.5 = 0.875 < `AcceptScore` 2 → Skip `weak` (was 2.75 + own at fan-out 1).
- `TestGracieObituaryGolden`: burial event has only `event_type`; `participation·subject·burial` is Other → 1.375 → Skip. A date on both sides passes, but that reshapes a (c) fixture.

Also failing under (a), reshape allowed but not done: `TestProposeInverseKinshipTermsCorrespond`, `TestProposeOverlappingBoundShowsTheDateAndResembles`. Already reshaped under (a): `TestNeighborCreditStopsAtOneHop` (person → birth (One) → place via location).

Done and committed: STEP-1 `1a7e6316`, STEP-2 `3208073c`, STEP-3 `97ea282e`. STEP-4/5 built and vetted, uncommitted (D-13 one commit). Not done: STEP-6, STEP-7, docs.

Plan's risk row (§11) anticipated weaker spouse/parent links but assumed the goldens' Targets survive; they do not. Taken to the user.

**User call (2026-10-19):** keep D-5's numbers, change no knobs; the two (c) tests assert the new behavior and carry a `// dogfood:` mark for the Spike 11 matching dogfood. Written into the plan under §5 after D-13. The other (a) failures are reshaped under D-11 as written. Resumed the round-0 implementer.

**Round 0 resumed:** head `b7522d16`, PR #493 (draft). Report deviations: the user's call implemented as written; fixtures reshaped under D-11 (a), each in the Landed note: `TestNeighborCreditStopsAtOneHop` (person → birth → place), `TestProposeInverseKinshipTermsCorrespond` (son named "John Smith"), `TestProposeOverlappingBoundShowsTheDateAndResembles` (death with bounds a month apart instead of a residence; display string updated). Manual check §10 not run (user's).
Gate (orchestrator) on `b7522d16`: PASS against `0dfd85fb`.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes in the diff). SEC+PRF routed by `core/matchregistry`, `core/segmentcompare` (a matching layer), `api/ffi/**`, `core/database/**`, graph code.

All three verdicts Clean; no blocker or major, so no round 2.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | Stale `Stats` in `graphalign` package doc and `promote-graph-alignment.md:76` | Fix now → F-1 | verified: `rg 'stats Stats\|Canon / Stats'` hits both; D-6 and DoD docs line name them |
| CON-2 | CON | minor | AC-4.4's grep matches the two `segment.go` source comments | Fix now → F-2 | verified: the grep prints `segment.go:27,29` only; D-5 prescribed the wording, AC-4.4 forbids it — reword the comment, AC stands |
| MNT-1 | MNT | minor | Layer 5 call built two ways in `align.go` | Fix now → F-3 | code this PR adds; small; working rule "build core infra right" over a follow-up |
| MNT-2 | MNT | minor | `loadTerms` called twice per `Propose` | Fix now → F-4 | verified: `layer.go:124`, `canon.go:30`; D-7 says one source |
| MNT-3 | MNT | minor | Which end is the event decided twice (layer vs canon) | Fix now → F-5 | verified against `linkOf` doc claim "cannot drift" |
| MNT-4 | MNT | minor | `layerLink`/`canonLink` same shape; leftover `fromNeighbor` | Fix now → F-6 | verified in the diff |
| SEC-note | MNT (routed) | minor | `loadTerms` last-row-wins on same term key across origins | Follow-up | pre-existing (deleted `loadKinship` had it); not made worse |

MNT-1 to MNT-4 ride along although their files are not otherwise touched: each is a few lines in code this PR adds, and a follow-up PR for them costs more than the fix.

Nits (logged): `MNT-5 · core/graphalign/align_test.go:824 · rename TestEdgeSupportCountsTheSeedingNeighborOnce` · `MNT-6 · promotealign/layer.go · neighborTypeProperty has one entry; use vocabkey.PropertyEventType` · `MNT-7 / CON-3 · comparison-layers.md · two blank lines after "Built in S11-14"`

### Fix plan 1
**F-1 — Drop the stale Stats references** (from CON-1)
Files: `core/graphalign/types.go`, `docs/promote-graph-alignment.md`.
Do: package doc reads "S9-42 loads Layer / Canon and calls Align" (D-6); the signature at `promote-graph-alignment.md:76` drops `stats Stats,` to match `align.go`.
Check: `rg -n 'stats Stats|Canon / Stats' core docs` prints nothing.

**F-2 — Meet AC-4.4's grep** (from CON-2)
Files: `core/matchregistry/segment.go`.
Do: reword the two `Gates` source comments without the deleted identifier, e.g. "carried from graphalign's old part_of scale, from end" / "…, to end; decided: the whole's end had no conflict factor". Update the matching inventory source text in `docs/matching-model.md` only if `TestKnobInventoryMatchesRegistry` compares it.
Check: `rg -n 'FanOut|loadStats|LoadStatsForTest|computeStats|EdgeSignature|SignatureScale|EdgePoints|EdgeFact' core api` prints nothing.

**F-3 — One segment builder in the walk** (from MNT-1)
Files: `core/graphalign/align.go`.
Do: one helper builds a `segmentcompare.Segment` from (key, kind, values) and one call path passes `st.metas, st.cfg.Segment, st.cfg.Subject, st.cfg.Value`; `segments` and `newPairSegments` both use it. No behavior change.
Check: `go test ./core/graphalign/...` unchanged.

**F-4 — Load terms once per Propose** (from MNT-2)
Files: `core/database/promotealign/promotealign.go`, `layer.go`, `canon.go`, `export_test.go`.
Do: `Propose` calls `loadTerms(g)` once and passes it to `loadLayer` and `loadCanon`; `LayerBridges` loads it the same way.
Check: `rg -n loadTerms core/database/promotealign` shows one non-test call site; promotealign tests pass.

**F-5 — One rule for the event end** (from MNT-3)
Files: `core/database/promotealign/terms.go`, `layer.go`, `canon.go`.
Do: a helper beside `linkOf` (e.g. `eventEndpoint(b connectrules.Bridge) (first, ok bool)`) decides which endpoint carries the event type; both sides use it; a bridge with no event end gets no event type (do not fall through to `e.To`). Keep the `linkOf` doc claim true.
Check: promotealign tests pass, including `TestLayerBridgesCoverEveryConnectRule`.

**F-6 — One link type in the walk** (from MNT-4)
Files: `core/graphalign/align.go`.
Do: merge `layerLink` and `canonLink` into one unexported type; in `scoreFixed` compare `link.back` directly and drop `fromNeighbor`.
Check: `go test ./core/graphalign/...`; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5d73271e`. Gate (orchestrator) PASS against `0dfd85fb`. F-1 to F-6 done; read by the orchestrator, all **Resolved** (no Fix-now blocker or major, so no round 2):
- CON-1 Resolved: `types.go` package doc and `promote-graph-alignment.md:76` match `Align(layer, canon, fixed, cfg)`.
- CON-2 Resolved: AC-4.4's grep prints nothing; the `Gates` comments read "graphalign's old part_of scale".
- MNT-1 Resolved: one `st.compare` helper; `compareSegments` and `newPairSegments` call it.
- MNT-2 Resolved: `Propose` loads terms once and passes them to `loadLayer` and `loadCanon`.
- MNT-3 Resolved: `eventEndpoint` beside `linkOf`; both sides use it; no fall-through to `e.To` for a bridge with no event end.
- MNT-4 Resolved: one `adjLink` type; `fromNeighbor` gone.

## Result
Gate PASS · CI pending · open: none · follow-ups: SEC-note (`loadTerms` same-key collision across origins, pre-existing), the D-11 dogfood marks (parent/child and burial weights) · to user: D-11 (c) stop — answered 2026-10-19 · CON miss: none
