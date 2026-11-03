# Ship log — S14-22 Promote's wire says what Go decided

**Plan:** [`S14-22-promote-wire-says-what-go-decided.md`](S14-22-promote-wire-says-what-go-decided.md) · **Branch:** `s14-22-promote-wire-says-what-go-decided` · **PR:** #700
**State:** ready

## Round 0 — implement
Head `409bcef9`. Gate PASS (orchestrator run against `fe93195f`). Report deviations: `connectrules.Edge` type named `Link` (collides with existing `Edge` func); `Bridge.IsAcyclic` exported for `graphalign.Connections`; page-link record guarded with `len(values) > 1` to satisfy `TestLayerPackages` no-bare-number rule; `pageEnds` lets the `only` scope override a caller's fixed row for a Subject outside scope (D-8 ambiguous); existing `PromoteModelTests` calling `mapRest()` now wait for the proposal (follows D-13); manual check not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX for `macos/App/**` and `Localizable.xcstrings`; SEC for `api/ffi/**`, `core/graphalign`, `core/segmentcompare`; PRF for `core/database/**` and `Features/Promote` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | Each proposal rescans every canon link and rebuilds the adjacency once per place connection (B × E) | Fix now → F-1 | verified: `graphalign/connections.go:87` calls `canonLinks` (full `canon` walk) per bridge; `ClosesCycle` (`filing.go:79`) rebuilds `adj` per call. Reviewer bench: B=300 E=24000 → 201 ms per decision |
| MNT-1 | MNT | major | Callers re-derive the decided ends and acyclic link `Decide` computed; copies already drifted (`== 2` vs `> 1`) | Fix now → F-2 | verified: `filing.go:142` vs `graphalign/connections.go:93`; `bridgefile.go:207` rebuilds `values[i][0].Record` |
| MNT-2 | MNT | major | `pageEnds` re-implements `mergeFixed`'s anchor rule | Fix now → F-3 | verified: `promotealign/connections.go` `pageEnds` member loop = `promotealign.go` `mergeFixed` member loop; `propose` holds `anchors` and drops it |
| PRF-2 | PRF | minor | At Done, `filedLinks` reads every filed link of the term once per place connection | Fix now → F-4 | minor in a function this PR rewrites; F-1 changes `ClosesCycle`'s input shape anyway |
| MNT-8 (from MNT outside-lens → PRF) | PRF | minor | `fileBridge` now runs `bridgeTerms` lookups before the wait and self-link rungs: one `property_terms` query per term for a waiting bridge at Done | Fix now → F-4 | `bridgefile.go:185` before `Decide` at `:195`; folds into F-4's per-pass read |
| MNT-3 | MNT | minor | Two copies of the Target → `EndTarget` switch in `pageEnds` | Fix now → F-3 | minor in a file this PR adds |
| MNT-7 | MNT | nit→minor | `ConnectionState.waits` doc says "first end when both do"; Go picks the end by rule order | Fix now → F-5 | stale comment in a file this PR changes |
| MNT-4 | MNT | minor | Swift `WaitReason` duplicates the three wire waits 1:1 | To user → Fix now → F-6 | D-11 says "`WaitReason` stays"; user (2026-11-03): fix as recommended. Supersedes that clause of D-11 |
| UX-1 | UX | minor | Subrows and footer show the previous proposal's outcomes with no pending sign, most after "Map the rest" | To user | Lag introduced by this PR but accepted by D-12; a pending sign needs new copy/styling the Brief line ("no chrome change") rules out. User (2026-11-03): Later PR → written into the S14-20 map scope entry's **Ships** (`deployment-plan.md`) |
| SEC-note (→ CON) | CON | — | With `only` set, a caller-fixed row outside the scope reads `not_on_page` | Rejected | documented in Landed note; `PromoteModel.swift:298` sends `flow.rows.filter(\.decided)` and out-of-scope rows are not shown until `mapRest`, which clears the scope, so Done never sends one |

Nits (logged): `CON-1 · plan D-1 and sketch still say Edge (Landed note explains Link)` · `MNT-5 · core/graphalign/connections.go:86 · res, _ := drops Decide's error` · `MNT-6 · core/graphalign/connections.go:119-132 · "h:"/"n:" record prefixes spelled twice`
Security: no findings (bound SQL, FFI ids via parseID, 200k-iteration random probe without panic).

### Fix plan 1

**F-1 — One adjacency per type and term per proposal** (from PRF-1)
Files: `core/connectrules/filing.go`, `core/connectrules/filing_test.go`, `core/graphalign/connections.go`, `core/graphalign/connections_test.go`. Skill: `add-match-rule` (graphalign only).
Failing test first: a growth check in `core/graphalign` (e.g. `TestConnectionsGrowsWithThePageNotPageTimesCanon`): run `Connections` with B and 4B unfiled `part_of` place bridges over a canon at the map's Scale (6,000 `location` edges on one place plus a few hundred `part_of` edges); assert the time (or a work counter, if the package already has one for growth checks; reuse the repo's existing growth-test pattern) grows with B, not with B × E. Must fail at `409bcef9`.
Do: give `connectrules` a link set that holds its adjacency (e.g. `type Links` with `Add(Link)` and `ClosesCycle(from, to)`, built by `NewLinks([]Link)`), so a BFS costs only the component it reaches. `ClosesCycle([]Link, …)` becomes a thin wrapper or goes if no caller is left. In `Connections`, build one `Links` per type|term up front from `canon` (one pass over `canon`, keeping only acyclic terms of registered types) and `Add` each page link that files to it; `Decide`'s `filed` input takes the prebuilt set (change `filed func() ([]Link, error)` to return the set). Update `TestClosesCycle`/`TestDecide` to the new shape.
Check: the growth test passes; `TestConnections`, `TestProposeReportsEachConnection` parity cases pass; scripts/check-pr.sh PASS.

**F-2 — `Decide` returns the ends it resolved** (from MNT-1)
Files: `core/connectrules/filing.go`, `core/graphalign/connections.go`, `core/database/identityclaims/bridgefile.go`, tests beside them.
Failing test first: `TestDecide` case asserting a `DecisionFiles` result carries `Records` in endpoint order and, for an acyclic term, the `Link` it checked (nil for a non-acyclic term).
Do: add `Records []string` and `Link *Link` to `Result`, set on `DecisionFiles`. `Connections` adds `*res.Link` to the F-1 set when non-nil and drops its own `len(values) > 1 && rule.IsAcyclic(term)` gate. `fileBridge` builds its endpoint ids from `res.Records` instead of `values[i][0].Record`. If `Bridge.IsAcyclic` then has no caller outside `connectrules`, unexport it. Update the plan's Landed note: the `len(values) > 1` guard line is gone (and the `IsAcyclic` export line if unexported).
Check: `rg -n "values\[\d\]\[0\]\.Record|IsAcyclic\(" core` prints only `core/connectrules`; scripts/check-pr.sh PASS.

**F-3 — `pageEnds` reads `mergeFixed`'s anchors; one target switch** (from MNT-2, MNT-3)
Files: `core/database/promotealign/connections.go`, `core/database/promotealign/promotealign.go`, tests beside them.
Failing test first: none new needed if behavior is unchanged; `TestProposeReportsEachConnection` and `TestProposeSendsEachConnection` must stay green (refactor).
Do: `pageEnds` takes the `anchors` slice `mergeFixed` already returned in `propose` (plus the caller's set, to tell caller-fixed from member anchors — e.g. have `mergeFixed` mark or return the Subject ids it added from members) and no longer takes `eg` or scans `eg.Members`. Add one `endTarget(t graphalign.Target, handleID []byte) graphalign.EndTarget` helper used for both the caller's fixed row (normalized so a handle pair is `TargetHandle`) and `Row.Start()`; the default arm for an unset row stays `ConnectionUnset`.
Check: `grep -n "eg.Members" core/database/promotealign/connections.go` prints nothing; `grep -c "switch" ` over `pageEnds` shows one target switch; scripts/check-pr.sh PASS.

**F-4 — Done reads filed links and term rows once per pass** (from PRF-2, MNT-8)
Files: `core/database/identityclaims/bridgefile.go`, its tests.
Failing test first: a query-count test in `identityclaims` (reuse the repo's existing query-counting helper, as `promotealign`'s `TestProposeQueryCount*` do): a Done filing N acyclic place bridges of one term, plus one waiting bridge, runs a number of `place_relationship` / `property_terms` reads that does not grow with N. Must fail at `409bcef9`.
Do: in the `fileHeads` pass, read each type|term's filed links once into the F-1 link set, lazily on first need, and `Add` each link that files so a later bridge in the same Done still sees it (AC-1.4). Cache `property_terms` rows by term id across the pass so `bridgeTerms` does not re-query per bridge.
Check: the query-count test passes; `bridges_test.go` "place relationship files; cycles refuse; split and amalgamation ok" passes; scripts/check-pr.sh PASS.

**F-5 — Reword the `waits` doc** (from MNT-7)
Files: `macos/App/Features/Promote/PromoteFlow.swift:81`.
Do: the doc says the end is the one Go names, not "the connection's first end when both do".
Check: scripts/check-pr.sh PASS.

**F-6 — `ConnectionState.waits` carries the wire outcome** (from MNT-4; user decision 2026-11-03, supersedes D-11's "`WaitReason` stays")
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/App/Features/Promote/PromoteView.swift`, `macos/App/Platform/L10n.swift` (only where it switches on `WaitReason`), `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`, `add-localized-string`.
Failing test first: none new; T-9's `waitingNamesTheEndAndWhy` and `connectionsFollowGosOutcome` assert against the outcome instead of `WaitReason` (`unrecognized` still words as not on this page).
Do: `ConnectionState.waits(endID:why:)` carries `PromoteConnectionOutcome` (the wait cases; `unrecognized` words as `notOnPage`); delete `WaitReason` and the case-by-case translation in `state(of:for:)`. No string or copy changes. Note in the plan's Landed note that D-11's `WaitReason` clause was superseded by the user in review.
Check: `rg -n "WaitReason" macos` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5ca10381`. Gate PASS (orchestrator). F-1 – F-6 done. Deviations: F-4's test lives in `core/database/promote/bridges_test.go` (needs `bridgeWorld`) with an `identityclaims.SetReadCounterForTest` hook; F-1's growth check is time-based (best of 5, 4× bridges < 3× time) since `graphalign` has no work counter; F-1/F-2/F-4 share commit `f2cc9fb9`; plan D-11 row text left as is, Landed note records the user's supersession.

## Round 2 — review
Reviewers: MNT, SEC+PRF (each had a Fix-now major). Not sent: CON (no findings), UX (UX-1 went to a later PR). Scope: `409bcef9...5ca10381`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) | Resolved | `Result.Records`/`Link`; `rg "values\[\d\]\[0\]\.Record\|IsAcyclic\(" core` empty |
| MNT-2 | MNT | major | (round 1) | Resolved | `pageEnds` reads `mergeFixed`'s `promoted`; no `eg.Members` |
| PRF-1 | PRF | major | (round 1) | Resolved | reviewer bench at head: E=24000 B=300 3.21 ms (was 201 ms); cost E + B |
| MNT-3, PRF-2, MNT-8, MNT-7, MNT-4 (F-6) | — | minor | (round 1) | Resolved | orchestrator read: one `endTarget`; `filePass` caches links and `property_terms`, `TestBridgeFilingReadsDoNotGrowWithTheBridges`; `waits` doc reworded; `rg WaitReason macos` empty |
| SEC+PRF note (→ CON) | CON | minor | A bridge with two type terms that files mid-Done is added only under `Term(keys)`, so an already-loaded set for its other term misses it where a fresh `filedLinks` read would include it | Follow-up | low confidence; association and graphalign canon key on `Term(keys)` alike, so the fresh-read SQL is the odd one; T-5 parity holds on the two-term case. Not a blocker; round 2 adds only blockers |

Nits (logged): `MNT-9 · core/database/identityclaims/bridgefile.go:286 · inline type|term key and "build Links on first need" closure repeat graphalign's linkKey; a keyed set in connectrules would hold both` · `MNT-10 · macos/App/Platform/L10n.swift:1420 · .waits(why:) admits .files; name .notOnPage and keep default for impossible cases`

No Fix plan 2 (no round-2 blockers).

## Result
Gate PASS · CI pending · open: none · follow-ups: two-term filed-link set (above), MNT-9, MNT-10 (nits) · later PR: UX-1 → S14-20 · to user: MNT-4 (Fix now), UX-1 (S14-20), both answered 2026-11-03 · CON miss: none
