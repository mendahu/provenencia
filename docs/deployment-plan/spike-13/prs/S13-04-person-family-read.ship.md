# Ship log — S13-04 Person family read

**Plan:** [`S13-04-person-family-read.md`](S13-04-person-family-read.md) · **Branch:** `s13-04-person-family-read` · **PR:** #671
**State:** ready

## Round 0 — implement
Head `14a8ebcb`. Gate PASS (orchestrator run). Report deviations (all tests, recorded in the plan's Landed note):
1. `graphcachetest.Canon` ids carry UUIDv7 version bits so `parseID` accepts them.
2. D-10 dropped-relative case tested through a split-out `dropUnread` in `family_drop_test.go`.
3. `everyKindHasOneDescriptor` expects 20 kinds.
4. Query-count tests call `nameformats.Warm`.
5. `Canon.Marriage` / `Canon.Birth` return `(event, joins)`.
6. `TestGetPersonFamily` does not assert the name (`Canon` has no name setter); AC-3.1 mentions name.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `api/ffi/**`, `core/kinship`, `core/database/**`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker | Parent sets × siblings unbounded; handler re-serializes summary and shapes per placement | To user → Fix now → F-1 | verified: `parentUnits` adds a sibling to every set inside S (`families.go` `for k, set := range sets`); `familyMemberProto` builds a full summary + shapes per placement. Reviewer probe: 120k input shapes → 108M wire shapes. The fix changes D-4 ("joins every set wholly inside S", "malformed data is shown, not resolved") and disproves the plan's Untrusted-size line |
| CON-1 | CON | major | `Siblings` unit emitted before partner and `OneParent` units | Fix now → F-2 | verified: `parentUnits` appends `unplaced` last, then `FamilyUnits` appends partners and oneParent after it; against D-3 |
| MNT-1 | MNT | major (re-rated from blocker: correct today, breaks D-2's promised one-entry add) | `persondetail.familyRoles` and the `{Parent, Child}` list restate `kinship`'s role table | Fix now → F-3 | verified: `family.go` `familyRoles`, inline role list in `Families`; `nuclearRoles` in `families.go:69` |
| MNT-2 | MNT | major | `dropUnread` restates which side of each `UnitKind` holds members | Fix now → F-4 | verified: `switch u.Kind` with `default` in `family.go` |
| MNT-3 | MNT | major | Handler drops relatives a second time (D-11 says it only maps) | Fix now → F-5 | verified: `familyMemberProto` returns nil on a missing summary; `members` filters nil |
| MNT-4 | MNT | minor | `PersonFamily` checks Person twice | Fix now → F-6 | verified: `isPerson` before `Families`, then `len(families)==0` |
| MNT-5 | MNT | minor | `isPerson` copies `nodesOf`'s check | Fix now → F-7 (same-pass cleanup) | verified: `compose.go:494`, `family.go:24`; grep shows only these two |
| MNT-6 | MNT | minor | `relationshipKeys` / `readsTerm` live in the family file | Fix now → F-8 | minor in a file this PR adds/changes |

Nits (logged): `UX-1 · GoStore.swift:1270-1283 · unknown KinshipPattern dropped quietly; doc that implied may be empty` · `MNT-7 · reader_test.go · "name_" comment → note` · `MNT-8 · canonlinks.go:7 · rewrap package doc` · `MNT-9 · families.go KinOf · compute relationshipKeys once` · `MNT-10 · reader.go LoadRelationships · loadSteps run twice with Siblings` · `MNT-11 · graphcachetest.go Provisional · literal 'provisional' vs identityclaims.StatusProvisional`

**To user — SEC-1.** User, 2026-11-01: "Bound at MaxPivotEnd". This narrows D-4's "malformed data is shown, not resolved" past `MaxPivotEnd` shown parents; record it in the plan's Landed note as a user-approved deviation from D-4 and correct the plan's Scale and Edges "Untrusted size" lines in the Landed note (not by editing D-4 itself).

### Fix plan 1
**F-1 — Bound parent pairing and sibling placement at `MaxPivotEnd`** (from SEC-1)
Files: `core/kinship/families.go`, `core/kinship/families_test.go`, `core/database/persondetail/family.go`, `core/database/persondetail/family_test.go` callers, `api/ffi/handlers` if it calls `FamilyUnits`. Skill: `add-match-rule` (no new knob; reuse `Rules.MaxPivotEnd`).
Failing test first: `TestFamilies/more parents than MaxPivotEnd` — 11 recorded parents of Pat, each pair also sharing a `birth_parents` pivot (or spouse links between every pair), and 20 siblings each with `shared_parent` shapes through all 11 → today every pair becomes a set and each sibling sits in every set. Expect: 11 one-parent sets (no pairing), then one `Parents` unit whose partners are exactly the 11 shared parents holding all 20 siblings; total placed siblings == 20. Plus `TestFamiliesGrowth`: double parents-over-bound and siblings, assert total placements and total shapes across all `Relative.Link`s grow at most linearly (count ≤ c × input shapes, c from the test's input). Also add `FuzzFamilyUnits` (seed from the table rows; assert no panic and placements ≤ siblings × (C(MaxPivotEnd,2)+MaxPivotEnd) + partners + children).
Do: `FamilyUnits(person, kin, kinOf, rules Rules)` (take `Rules` as `Siblings` does; `persondetail.Families` passes `matchregistry.Default().Kinship`). In `parentUnits`, when `len(parents) > rules.MaxPivotEnd`: emit each parent as a lone set with no pairing, and place each sibling only in the unit whose partners are exactly S (its own lone set when |S| = 1, else the made-after unit), never in every set inside S. At or under the bound, D-4 is unchanged: AC-1.2, AC-1.5, the three-parent row and the ten-parent row still pass unchanged. No numeric literal other than 0 and 1 in `families.go` (`guard_test.go`).
Check: new tests pass; every existing `TestFamilies` row unchanged; scripts/check-pr.sh PASS.

**F-2 — Emit the `Siblings` unit last** (from CON-1)
Files: `core/kinship/families.go`, `core/kinship/families_test.go`.
Failing test first: `TestFamilies/siblings unit comes last` — `slices.Concat(johnMaryPat, []kinship.Edge{sibling(1, 11, 1), parent(1, 14, 2), spouse(1, 2, 3)})` → `Parents{6,7;}`, `Partner{2;}`, `OneParent{;14}`, `Siblings{;11}` (today `Siblings` comes second).
Do: `parentUnits` returns the unplaced `Siblings` unit separately, as `partnerUnits` returns `oneParent`; `FamilyUnits` appends it after `OneParent` (D-3).
Check: the row passes; scripts/check-pr.sh PASS.

**F-3 — One role table, in `core/kinship`** (from MNT-1)
Files: `core/kinship/families.go`, `core/kinship/families_test.go`, `core/database/persondetail/family.go`.
Failing test first: none required (no behavior change today); `TestKinOf` and `unitsOf` must read the same exported source.
Do: export the roles in `nuclearRoles` order (`func Roles() []Role` built from `nuclearRoles`) and a method saying whose kin `FamilyUnits` reads (`func (k Kin) Around(person []byte) [][]byte`: the distinct ids of `k[RoleParent]` then `k[RoleChild]`, in order). `Families` collects relatives over `kinship.Roles()` (drop `familyRoles`) and builds `around` from `kin.Around(id)` (drop the inline `{RoleParent, RoleChild}`). Point `unitsOf` and `TestKinOf` in `families_test.go` at the same functions.
Check: `grep -rn "familyRoles\|\[\]kinship.Role{" core api` prints nothing outside `core/kinship`; scripts/check-pr.sh PASS.

**F-4 — The empty-unit rule lives beside `UnitKind`** (from MNT-2)
Files: `core/kinship/families.go`, `core/kinship/families_test.go`, `core/database/persondetail/family.go`.
Failing test first: `TestUnitEmpty` — each `UnitKind` with only partners, only children, and neither → Parents/Partner empty iff no partners; OneParent/Siblings empty iff no children.
Do: `func (u Unit) Empty() bool` with an exhaustive switch over every `UnitKind` (no `default`; an unknown kind panics or returns true with a comment, matching the package's idiom). `dropUnread` calls `u.Empty()`.
Check: test passes; `TestDropUnread` unchanged and passing; scripts/check-pr.sh PASS.

**F-5 — The handler only maps** (from MNT-3)
Files: `api/ffi/handlers/person_detail.go`.
Failing test first: none (unreachable branch).
Do: `familyMemberProto` no longer returns nil: a missing summary is an invariant failure — return an error from the mapping (wrapped, not a new code) — or map with `h, _ :=` per D-11; prefer the error. Remove the nil filter in `members`.
Check: `TestGetPersonFamily` passes; scripts/check-pr.sh PASS.

**F-6 — One Person guard in `PersonFamily`** (from MNT-4)
Files: `core/database/persondetail/family.go`.
Do: keep `conclusiondetails.Handle`, call `Families`, and let `len(families) == 0` be the not-found case; remove the first `personTypeID` / `isPerson` pair.
Check: `TestPersonFamily/AC-1.12` passes; scripts/check-pr.sh PASS.

**F-7 — `graphcache.Node.Is(typeID)`** (from MNT-5, same-pass cleanup)
Files: `core/database/graphcache` (the file holding `AcceptedMembers`), `core/database/conclusionsummaries/compose.go`, `core/database/persondetail/family.go`.
Do: add `func (n *Node) Is(typeID []byte) bool` (non-nil, unmerged, `SubjectTypeID` equals `typeID`) next to `AcceptedMembers`, with a test. `nodesOf` (`compose.go:494`) and `shownPerson` use it; drop `isPerson`.
Check: `grep -rnE "Merged.*SubjectTypeID|SubjectTypeID.*Merged|!n\.Merged && bytes\.Equal" core api | grep -v _test` prints only the new method; scripts/check-pr.sh PASS.

**F-8 — Move the relationship reader next to `ParentStep`** (from MNT-6)
Files: `core/kinship/families.go`, `core/kinship/kinship.go`.
Do: move `relationshipKeys` and `readsTerm` into `kinship.go` beside `ParentStep`, unchanged.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b12a7fe3`. Gate PASS (orchestrator run). F-1 – F-8 done (code in `7085f038`, docs in `b12a7fe3`). Implementer's first gate run hit a `promotealign` `TestLargeSourceBudgets/warm_at_open` timing miss (2.79s vs 2.5s) while its own builds overlapped; passed alone and on rerun, and on the orchestrator's run. Deviations: grouped commits; fuzz seeds hand-encoded; one `add-match-rule` sentence added for the bound.

## Round 2 — review
Reviewers: SEC+PRF (SEC-1), CON (CON-1), MNT (MNT-1 – MNT-3). Scope `14a8ebcb...b12a7fe3`. Skipped: UX (no Fix-now finding).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | (round 1) Siblings unit last | Resolved | `FamilyUnits` appends Siblings after OneParent; `TestFamilies/siblings unit comes last` |
| MNT-1 – MNT-3 | MNT | major | (round 1) role table, `Unit.Empty`, handler maps only | Resolved | `kinship.Roles()`, `Kin.Around`, `Unit.Empty`, `familyMemberProto` returns an error |
| MNT-4 – MNT-6 | MNT | minor | (round 1) | Resolved (orchestrator read `7085f038`) | single guard in `PersonFamily`; `graphcache.Node.Is`; `readsTerm`/`relationshipKeys` in `kinship.go` |
| SEC-1 | SEC | blocker | (round 1) | Not resolved → SEC-1R | sibling half fixed; parent half open |
| SEC-1R | SEC | blocker | Above the bound, made-after `Parents` units repeat each parent and its whole link, so output is subsets × link size | Fix now → F-9 | verified: `families.go` `extra`/`extraAt` branch still runs when `!pairing`; Fix plan 1 asked for it (orchestrator's error, not the implementer's). Reviewer probe: 25k input edges → 1.6M association ids. Within the user's "Bound at MaxPivotEnd" decision |
| PRF-1 | PRF | minor | `Kin.Around` dedupes with `slices.ContainsFunc`, quadratic in parents + children | Fix now → F-10 | minor in code this PR adds |

Nits (logged): `MNT-12 · families_test.go:181 · TestKinOf inline role list → kinship.Roles()` (rides F-10)

### Fix plan 2
**F-9 — Above the bound, each parent appears once** (from SEC-1R)
Files: `core/kinship/families.go`, `core/kinship/families_test.go`, the plan's Landed note.
Failing test first: extend `TestFamiliesGrowth`: P > `MaxPivotEnd` parents, each parent→person link carrying several associations (or several `birth_parents` shapes), and one sibling per distinct 2-subset of parents; make `placements` count associations as well as shapes. Today partner placements and association counts grow as subsets × link size. Expect: partner placements == P, child placements == S, total associations across all `Relative.Link`s ≤ the input's associations. Also `TestFamilies/over the bound a sibling sharing several parents sits with the first` — 11 parents, one sibling sharing parents 3 and 7 → it is a child of parent 3's lone set only; no extra `Parents` unit.
Do: when `!pairing`, make no made-after units: a sibling with one or more shown shared parents joins the lone set of its first shared parent (in parent order) and nothing else; a sibling with none goes to `Siblings`. At or under the bound nothing changes (every existing `TestFamilies` row and `TestFamiliesMoreParentsThanMaxPivotEnd` updated only where it asserted the made-after unit above the bound). Update the `parentUnits` doc comment and the Landed note's description of the over-bound placement to match.
Check: tests pass; scripts/check-pr.sh PASS.

**F-10 — `Kin.Around` dedupes with a map** (from PRF-1, MNT-12)
Files: `core/kinship/families.go`, `core/kinship/families_test.go`.
Do: keep the parent-then-child order, dedupe with a `map[string]bool` as the replaced code did. In `TestKinOf`, use `kinship.Roles()` in place of the inline role list.
Check: `TestKinAround` and `TestKinOf` pass; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `1f708b75`. Gate PASS (orchestrator run; `FuzzFamilyUnits` 10s clean per implementer). F-9, F-10 done. Orchestrator read the F-9 diff: past the bound every set is one parent at its index, and a sibling joins `units[shared[0]]` only, so SEC-1R is **Resolved**; PRF-1 and MNT-12 **Resolved**. Deviations: growth test is the new `TestFamiliesGrowthPairs`; its association bound is ≤ 2 × input (a sibling's implied link carries the person's parent-edge associations too), linear either way; `TestFamiliesMoreParentsThanMaxPivotEnd` rewritten for the new placement. No round 3 (cap).

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: SEC-1 (answered 2026-11-01, "Bound at MaxPivotEnd") · CON miss: none
