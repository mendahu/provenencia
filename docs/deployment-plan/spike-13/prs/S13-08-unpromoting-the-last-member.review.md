# Review — S13-08 Unpromoting the last member

**Verdict:** Fits with changes — Ready with changes (round 1: Not ready; round 2: every blocker Fixed)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None. The product calls were asked before the plan was written (2026-11-04): association handles go too; Subject delete joins S13-08; no cleanup of existing memberless handles; the Unpromote copy as proposed.

## Writer's own calls

- A handle another handle is merged into stays, and the write succeeds (D-6). Unreachable today: no merge writer.
- Only removing an **accepted** claim can delete a handle (D-1). A stub that never had an accepted member stays when a provisional claim on it goes.
- Subject delete copy: "%@ has no other records, so it is deleted." / "…they are deleted." (D-17), in the form of the approved Unpromote copy.
- The Unpromote confirm waits for the preflight before it opens; on a preflight error it opens with today's message and the error (D-15).
- `GetDeleteImpact(canonical_entity)` becomes valid; the app does not call it (D-9).
- The batch Promote path keeps its inline Subject load so its refusal order is unchanged (D-12).
- Old-rule fixtures that fail because a sole-member handle is now deleted get a second member, not a changed assertion (STEP-2).

## Round 1

### M-1 · blocker · Two derivations of "the handle will be deleted"
The preflight used a SQL predicate with a merged clause while the write also ran Impact.
**Answer:** Fixed — D-4: one `handleGoesWithout` (membership SQL, then `Impact(...).Allowed`), used by the write, the cascade and the preflight. The merged clause is gone from SQL.

### M-2 · blocker · `ReleaseEmptiedHandles` decodes audit changes and every writer must remember it
**Answer:** Fixed — D-2: the rule lives in `releaseClaim`, the one claim removal; `releaseClaimRow` (D-3) is the raw removal the dying handle's facet uses. `Unpromote`, `UnfileBridges` and `subjects.Delete` need no code change. `ReleaseEmptiedHandles` and the change decoder are dropped (this also answers X-6).

### M-3 · blocker · `total == 1` key choice
**Answer:** Fixed — D-17: one plural key `deleteImpact.confirm.deletesHandles(count:handleRefs:)`, per `macos-l10n.mdc` and `deletesCitations`.

### M-4 · blocker · FakeStore derives the last-member rule
**Answer:** Fixed — D-14: scripted `unpromoteImpactBySubject`; T-8 scripts it. The "Handles stay." doc is in the sweep and STEP-5.

### M-5 · blocker · Conclusion location mapping duplicated; catch-all to Persons
**Answer:** Fixed — D-8: `workspacelocation.Conclusion`, used by `core/search` `locationFor` and the new projector; associations get no Person link. T-11.

### M-6 · blocker · Kind as a bare String with a person fallback
**Answer:** Fixed — D-15, D-16: `EvidencePrimaryKind`, a `switch` with no `default`, keys `….person|event|place` like `openHandlePage`.

### M-7 · major · `LastMemberHandles` signature fits neither caller
**Answer:** Fixed — D-10, D-11: one `canonical_entities.last_member` `inboundEdge` (`lastMemberEdge`, in `cascadeEdgeList` since round 2); `LastMemberGroup` runs `collectGroups` over it.

### M-8 · major · D-9 broke the `Named` contract
**Answer:** Fixed — D-10 option (a): the Subject facet is no longer `Named`; both groups are cascade-only edges.

### M-9 · major · "No accepted claim left" deletes stubs that never had a member
**Answer:** Fixed — D-1: only the removal of an accepted claim makes a candidate; AC-1.7 and a T-1 row.

### M-10 · major · Third copy of the primary-Subject checks
**Answer:** Fixed — D-12 `loadPrimary`, used by `Unpromote` and `UnpromoteImpact`.

### M-11 · major · Old-rule tests and comments missing
**Answer:** Fixed — §4 lists both "handle stays" tests and the comments; STEP-2 says how to rewrite them; the DoD sweep adds "Handles stay", "handle stays", "the handles stay", "are left alone".

### M-12 · minor · S13-17 would repeat the reading of the preflight
**Answer:** Fixed — D-14: the store returns `CatalogDeleteImpactListed?`.

### X-1 · blocker · `canonical_entities` has no `PK` in `catalogmodel`
**Answer:** Fixed — D-7: `PK: "id"`; `effects/path.go` `pkOf` already defaults to `"id"`, and a Stop condition covers any effects test change.

### X-2 · blocker · D-1 and D-8 disagreed on non-accepted claims
**Answer:** Fixed — D-1 (accepted claims only) with D-4's single check makes the write and the preflight one rule; the leftover provisional/rejected claims of a dying handle are released (D-3) and not named in a confirm (Out line: no writer creates them).

### X-3 · major · Sweep hits `catalog-deletes.md` line 50, not in STEP-7
**Answer:** Fixed — STEP-7 names the facet table line with its new wording.

### X-4 · major · T-4 names the two-member fixture
**Answer:** Fixed — T-4 is a new `TestLastMemberDeleteEndToEnd` with its own one-member fixture and named assertions.

### X-5 · major · `extraCascadeList` shape unspecified
**Answer:** Fixed — D-10 names the type (`inboundEdge`), `Bucket`, `sync.OnceValue`, and how `cascadesFor` filters and orders.

### X-6 · major · No decoder for `entity_id` in a change
**Answer:** Fixed — obsolete with M-2: no change is decoded.

### X-7 · minor · FakeStore's Subject cascade and doc
**Answer:** Fixed — D-18: the cascade stays scripted; the doc is in STEP-5 and the sweep.

### X-8 · minor · Swift choices not made
**Answer:** Fixed — D-16 return type `String`; D-13 / AC-2.3 name `parseID`'s error; STEP-6 spells out each of the three tests; T-8 scripts its own data instead of relying on fixture membership.

### X-9 · minor · Associations not named in the Unpromote confirm
**Answer:** Declined — the confirm already lists the bridges under "Also unpromoted" (`EvidenceGraphView.unpromoteDetail`); an association handle has no page a researcher visits, so naming its ref adds nothing. Recorded in D-11.

## Round 2

Both reviewers checked the round-1 answers against the code and found them sound, except as below.

### M-13 · blocker · `extraCascadeList` as `sync.OnceValue` is an initialization cycle
Its edge calls `Impact` → `cascadesFor` → the list.
**Answer:** Fixed — D-10: `cascadeEdgeList`, a plain slice built in the `init()` that builds `facetReleaseList` (the precedent for the same cycle), plus `lastMemberEdge`; Codebase fit row cites `facetReleaseList`.

### M-14 · blocker · A new invented via crosses to Swift as a raw string with a fallback and no parity test
**Answer:** Fixed — D-10 exports `ViaLeavesHandle`, `ViaDeletesCitations`, `ViaLastMember`; D-17 adds `DeleteImpactCascadeVia` in `ConclusionVocabulary.swift`, switched on by `cascadeLines`, with a `TestSwiftVocabularyMatchesGo` row.

### M-15 · major · `loadPrimary` leaves copies in `Save` and `batch.go`
**Answer:** Fixed in part — `Save` uses it (D-12). Declined for `batch.go`: it checks the graph id between the load and `requirePrimary` (`batch.go` lines 113–125), so moving it would turn a foreign-graph non-primary row from `ErrInvalid` into `ErrUnsupportedType`; D-12 records why.

### M-16 · major · The un-Named facet keeps dead Count/List and its SQL is copied
**Answer:** Fixed — D-10 and Deletes: the facet keeps `Parent`, `Via`, `Remaining`, `Release`; its List SQL moves into edge (b).

### M-17 · minor · An accepted-members query exists
**Answer:** Fixed — Codebase fit names `identityclaims.AcceptedMembersTx` and the import cycle that rules it out.

### X-10 · blocker · A fourth Swift test calls `beginUnpromote` synchronously
**Answer:** Fixed — STEP-6 names all four tests and says every direct call gets `await`, the `performCardAction` one waits with `waitUntil`, and why the expectations hold.

### X-11 · blocker · Old-rule Go tests not listed
**Answer:** Fixed — STEP-2 lists the candidates and gives one rule (second accepted member in the fixture, assertion unchanged; any other cause stops); STEP-3 rewrites `TestGetDeleteImpactPromotedSubject`.

### X-12 · major · Wrong path for `EvidenceGraphSnapshot.swift`
**Answer:** Fixed — §4.

### X-13 · major · `completed.md` without a path
**Answer:** Fixed — STEP-7 and the DoD give `docs/deployment-plan/spike-13/completed.md`.

### X-14 · major · Small open choices
**Answer:** Fixed — T-4 names `catalog_search_docs.entity_id`; `LastMemberGroup` reads `lastMemberEdge`; D-10 says refused reports still carry cascades; the GoStore row says `.listed.first` when `hasDeletes`; a Stop condition covers the Subject type list's new locations.
