# Ship log — S11-19 Claim evidence query

**Plan:** [`S11-19-claim-evidence-query.md`](S11-19-claim-evidence-query.md) · **Branch:** `s11-19-claim-evidence-query` · **PR:** #513
**State:** ready

## Round 0 — implement
Base `main` @ `9ca7472e`. Head `8262d21e`. Gate PASS (orchestrator rerun). Report deviations (also in the plan's Landed note):
- D-10's collapse key also includes outcome and property cause; equal whenever the plan's key parts are equal.
- A corresponding link whose bridges all name non-members keeps its Relationship with no bundles (D-13 read as "skip the bridge, not the link").
- The loader leaves out a filed bridge Subject whose endpoint Observations do not name both ends.
- STEP-5 and STEP-6 are one commit; `subjects.SourcesOf` has its own.
- The implementer committed this ship log early, in commit 2.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (matching layers, `core/database/**`). Skipped: UX (no `macos/App/**`, strings, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Every accepted member of every neighbor entity is loaded, but `Build` reads only the ones a bridge names | Fix now → F-1 | verified: `evidence.go:136` `q.members(hid)` for each neighbor; `build.go` `bundles` uses only `n.Members[ni]` that a `Bridge.Neighbor` matches |
| PRF-2 (+ MNT outside-lens PRF) | PRF | major | `members` calls `layerValues` once per member (clone and sort of the whole Source each time); `bridges` scans a Source's Observations once per bridge | Fix now → F-2 | verified: `evidence.go:220-224`, `profiles.go:94-95`, `evidence.go:299`. The batched SQL read the reviewer offers is ruled out by D-8 ("no second Observation read"); grouping by Source is within D-8 |
| MNT-1 | MNT | major | "A layer bridge from one end" rule is copied from `graphalign`'s `newState` into `Evidence` | Fix now → F-3 | verified: `align.go:123-127` and `evidence.go:82-89`. Same-pass cleanup |
| MNT-2 | MNT | major | `bridges` reads endpoint Observations with no polarity check and keeps the last value when an end names two Subjects | Fix now → F-4 | verified: `evidence.go:299-311`; `bridgeNeighbors` reads `polarity = 'positive'` only, `resolveEndpoint` treats two values as unresolved |
| SEC-1 | SEC | minor | `req.Neighbors` has no length bound; `checkFixed` queries before the duplicate check | Fix now → F-5 | verified: `evidence.go:47-57` |
| MNT-3 | MNT | minor | `routeOf` re-derives which way `Rules.Lead` chose the lead | Fix now → F-6 | verified: `sections.go:36-44`, `subjectcompare.go:167-180`. Minor in code this PR adds |
| PRF-3 | PRF | minor | Each `Evidence` call rebuilds the Source's whole layer | Later PR | written into the S11-21 map scope entry (`deployment-plan.md`, S11-21 **Ships**) on this branch |
| MNT out-of-lens CON | CON | — | `relationships` compares against `canons[0].Key` only | Rejected | `segmentcompare.Corresponds` is `a == b` (`key.go:44`), so every corresponding canon key equals `key` |

Nits (logged): `CON-1 · ship log header stale (updated by orchestrator)` · `CON-2 · completed.md says "In review" while the plan says Landed` · `MNT-4 · sections.go:131 · struct key for collapse map` · `MNT-5 · build.go:44 · boolRank(!isOne) reads reversed` · `MNT-6 · build_test.go:158 · equalIDs is slices.EqualFunc` · `MNT-7 · evidence.go:111-131 · g.Node fetched twice per neighbor` · `MNT-8 · promotealign.go:39,54 · propose repeats loadInputs' checks`

### Fix plan 1

**F-1 — Load only the neighbor members a corresponding bridge names** (from PRF-1)
Files: `core/database/promotealign/evidence.go`.
Failing test first: none practical as a behavior test (output is unchanged). Add `TestEvidenceLoadsOnlyBridgedNeighborMembers` in `evidence_test.go`: a neighbor handle with two accepted members, only one reached by a filed bridge; assert the evidence is unchanged and (through an `export_test.go` hook or a counter on `evidenceReader`) that only the bridged member's Source was read. If a read-count hook is not possible without touching `graphcache`, assert instead that `Neighbor.Members` passed to `Build` holds only the bridged member, via a small exported-for-test seam in `export_test.go`.
Do: in the neighbor loop, compute `canon` (`q.canonLinks`) first; collect every `Bridge.Neighbor` id from those links; build the neighbor's member list from `acceptedMembers(H′)` filtered to that set, then read values only for those ids. `claimevidence.Build` does not change. Keep the row entity H's members whole (sections compare every member).
Check: the new test passes; all `TestEvidence*` pass unchanged; scripts/check-pr.sh PASS.

**F-2 — Read member values and bridge Observations once per Source** (from PRF-2, MNT outside-lens PRF)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/profiles.go` if needed.
Failing test first: none (performance only). Existing `TestEvidence*` and `TestBuild` must pass unchanged.
Do: in `members`, group the member ids by Source id (keeping claim order in the output), call `g.Source` once per Source and `layerValues(sg, <set of that Source's member ids>, …)` once per Source (D-8's set signature). In `bridges`, likewise group bridge Subjects by Source and walk each Source's Observations once, bucketing by `SubjectID`. Do not add a second SQL Observation read (D-8).
Check: `TestEvidence*`, `TestBuild` pass; scripts/check-pr.sh PASS.

**F-3 — One owner for "a layer bridge seen from one end"** (from MNT-1)
Files: `core/graphalign/types.go`, `core/graphalign/align.go`, `core/database/promotealign/evidence.go`.
Failing test first: `TestBridgeFrom` in `core/graphalign` (new): A end → (B, KeyA, true); B end → (A, KeyB, true); neither → false.
Do: add `func (b Bridge) From(id []byte) (other []byte, key vocabkey.BridgeKey, ok bool)`. `newState` builds `layerAdj` through it (one call per end) and the `Evidence` layer-link loop calls it instead of its `switch`.
Check: `rg -n "KeyA|KeyB" core/graphalign core/database/promotealign -g '!*_test.go'` prints only `types.go` (the struct and `From`); graphalign and promotealign tests pass; scripts/check-pr.sh PASS.

**F-4 — Read a bridge's ends from positive, unambiguous endpoint Observations** (from MNT-2)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/evidence_test.go`.
Failing test first: `TestEvidenceBridgeEndpoints` (new), subtests: a bridge whose neighbor end is recorded with a non-positive polarity gives no bundle; a bridge whose member end names two different Subjects gives no bundle; a positive single-valued bridge gives one bundle (control).
Do: in `bridges`, skip endpoint Observations whose `Polarity` is not `positive` (match `bridgeNeighbors`' SQL); an end that names two different Subjects is unresolved and the bridge is left out (as `identityclaims.resolveEndpoint`). State the rule in the doc comment.
Check: the test passes; scripts/check-pr.sh PASS.

**F-5 — Bound the neighbor list** (from SEC-1)
Files: `core/database/promotealign/evidence.go`, `evidence_test.go`.
Failing test first: a `TestEvidenceInvalidRequest` case with more neighbors than the Source has primary Subjects → `promote.ErrInvalid`.
Do: before the neighbor loop, reject `len(req.Neighbors) > len(in.primary)` with `promote.ErrInvalid`. Leave the per-entry validation and duplicate rule as they are.
Check: the case passes; scripts/check-pr.sh PASS.

**F-6 — `subjectcompare` says how a property counts under its lead** (from MNT-3)
Files: `core/subjectcompare/subjectcompare.go`, `subjectcompare_test.go`, `core/claimevidence/sections.go`.
Failing test first: `TestRulesRole` (new): a weighted lead → role lead; a pooled property → role pooled; a period end → role period end; a lifespan Span → not ok.
Do: add a role-returning method (e.g. `func (r Rules) Role(kind, key) (lead vocabkey.PropertyKey, role Role, ok bool)`, `Role` one of lead / pooled / period end) holding `Lead`'s walk; `Lead` calls it. `routeOf` reads the role instead of re-checking `Periods`.
Check: `TestRulesRole`, `TestSections*`, all `subjectcompare` tests pass; scripts/check-pr.sh PASS.

Minors riding along (optional, implementer's call, no tests): MNT-7 (keep the node from the first loop).

## Round 1 — fix
Head `2d808c8a`. Gate PASS (orchestrator rerun). F-1 to F-6 done; MNT-7 rode along. Deviations: F-1 used the planned fallback seam (`bridgedOnly`), since there is no read-count hook without touching `graphcache`. F-4's two-Subjects subtest passed before the fix; it guards the rule and the polarity subtest exercises the fix. F-3: `layer.go:199` still names `KeyA`/`KeyB`, but it constructs the `Bridge` literal and does not read it. F-6: `Role` checks pool before period end, as `Lead` does; no default-registry property is both.

## Round 2 — review
Reviewers: MNT, SEC+PRF (both had Fix-now blocker or major). CON not sent (clean in round 1). Scope `8262d21e...2d808c8a`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | (round 1) | Resolved | `bridgedOnly` trims neighbor members before values are read (`evidence.go:138-149, 211-221`) |
| PRF-2 | PRF | major | (round 1) | Resolved | `memberValues` and `bridges` read each Source once per batch |
| MNT-1 | MNT | major | (round 1) | Resolved | `graphalign.Bridge.From`, used by `newState` and `Evidence` |
| MNT-2 | MNT | major | (round 1) | Resolved | positive-only endpoints; an ambiguous end drops the bridge; `TestEvidenceBridgeEndpoints` |
| SEC-1 | SEC | minor | (round 1) | Resolved | bound at `evidence.go:50-52` (checked by orchestrator and SEC+PRF) |
| MNT-3 | MNT | minor | (round 1) | Resolved | `subjectcompare.Rules.Role`; `routeOf` reads `RolePeriodEnd` |
| MNT out-of-lens CON | CON | — | `newState` appends `(B, KeyA)` twice for a self-bridge `A == B` | Rejected | `connect` rejects `from == to` before writing (`connect.go`, `bytes.Equal(from.ID, to.ID)` → `ErrInvalid`), so no self-bridge reaches the layer |

Nits (logged): `MNT-9 · evidence.go:225-231 · members is a one-caller wrapper beside the inline neighbor path` · `MNT-10 · subjectcompare.go:191 · Role returns RoleLead with ok=false; consider RoleNone`

No round-2 blockers; no Fix plan 2. `origin/main` unchanged since the branch was cut (`9ca7472e`).

## Result
Gate PASS · CI green · open: none · follow-ups: PRF-3 (in S11-21 scope) · to user: — · CON miss: none
