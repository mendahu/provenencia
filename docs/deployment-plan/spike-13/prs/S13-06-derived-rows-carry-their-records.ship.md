# Ship log — S13-06 Derived rows carry their records

**Plan:** [`S13-06-derived-rows-carry-their-records.md`](S13-06-derived-rows-carry-their-records.md) · **Branch:** `s13-06-derived-rows-carry-their-records` · **PR:** #640
**State:** ready

Base: `main` @ `588e4a82` (S13-24 has landed since the plan's `705cd98`; the plan's §11 says whichever lands second rebases).

## Round 0 — implement
Head `81fd278f`. Report: no stop. Deviations: STEP-4 – STEP-7 in one Swift commit; three test fix-up commits; small unplanned helper `lifeOwners(g, v, pg, picks)` in `compose.go`; T-4 is one shared "not a Person" assertion; landing docs committed in round 0. §10 manual check not run (needs the app). Push failed: `github.com` does not resolve from this machine's shells (ssh, sandbox off too); origin and #640 sit one commit behind the local head.
Gate (orchestrator) on `81fd278f`: first run FAIL `xcode-test`, 23 issues all in `SourceGraphFormModelTests` and one `SourcePageModelTests` case (`loadState → .loading`: an async race in files this PR does not touch); rerun PASS. Logged as a flake, not a gate failure.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `api/ffi/**`, `core/database/**`, Features models).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | Every Person page open loads every property's outcomes of each owner, and every toponym record of each life Place | (1) filter to the owned property: To user (refines D-5's SQL); (2) defer or cap records: To user (D-7) | verified: `sqlOutcomes` has no property filter; `PersonDerivedRows` keeps one field per owner (`conclusiondetails.go:539-541`). Not measured; the map's **Scale** names no records-per-Place figure |
| UX-1 | UX | major | One owner that is not the shown one (lead event has no Place, another is at York) lists greyed "Kept" names under "No birth place recorded" and never names that owner | To user (contradicts D-12 / LW-3) | verified: `leadLife` takes Places from the lead only (`compose.go:633`); owners come from every pick; one group → `owner: nil` |
| PRF-2 | PRF | major | Derived records rebuilt on every `body` pass of `ConclusionDetailLoaded` | Pending PRF-1 (moot if records load on open of the Why) | verified: `presentation` is computed in `body` and calls `make` |
| MNT-1 | MNT | major | Toponym field → Place title written twice (`ownerLine`, `PlaceDetailContent.init`), already drifted | Fix now (same-pass) | verified by reading both |
| MNT-2 | MNT | major | `PersonLifeOwners` copies `composePersons`' event-Place gather + `buildPlaceGraph` (third copy in `composeEvents`) | Fix now (same-pass) | `grep PlacesOfEvent\|buildPlaceGraph(` → 469/477, 679/687, 740/751 |
| CON-1 | CON | major | AC-1.6's denied-only date has no Go test | Fix now | `rg -i deny` over the new Go tests finds nothing |
| UX-2 | UX | minor | A life Place row whose only Place has no kept name lost its open arrow | To user (contradicts D-10's "instead of `.first`") | verified: `opens = shown.map(...)`, `shown` nil when no `placeLine` |
| UX-3 | UX | minor | VoiceOver cannot tell which owner group is the shown value; ink alone carries it | To user (LW-7 fixes the spoken text) | verified: `counted` only sets `foregroundStyle` |
| SEC-1 | SEC | minor | Date owners are not deduped by event id (Places are) | Fix now | verified: `lifeOwners` appends `Dates` per pick |
| MNT-3 | MNT | minor | Date-row kind compared twice in `PersonDerivedRows` | Fix now | lines 498 and the `Event` attach |
| CON-2 | CON | minor | `TestFieldsOf`, `TestPersonDerivedRows` not table-driven | Fix now | `go-tests.mdc` |
| CON-3 | CON | minor | "S13-06 is Planned" left in the map's Order paragraph | Fix now | `deployment-plan.md:230` |
| UX-4 | UX | minor | §10 manual check not run | Rejected (the user's step) | ship-pr §8: the user does the manual check |
| SEC (outside lens → CON) | CON | — | Person page stale after a Place-only write | Rejected | `conclusionheaders.Dependents` walks Place → events at it → their birth/death people, so the Person is in `HeaderDeps` |

Nits (logged): `MNT-4 · conclusion_detail.go derivedRowsProto · kinds map per call; package var` · `MNT-5 · compose.go eventDate · six returns; small struct` · `MNT-6 · PersonDetailContent.rows · DerivedPlace.shown runs 3× per place row` · `MNT-7 · DerivedRowKind string type never used as a string` · `UX-5 · ReconciliationReasoningView.swift:93 · .gridCellColumns(4) on a non-GridRow child does nothing` · `UX-6 · ReconciliationRecord.swift:475 · single group id ""`

**State:** answered — see User answers below.

**User answers (2026-10-30):** UX-1 → name the owner whenever its group is not the shown one's (overrides D-12 / LW-3 for that case). UX-2 → open arrow falls back to the first Place; ink still follows the shown Place (overrides D-10's open target). UX-3 → Later PR: into S13-11's scope entry. PRF-1 → open: user asked what the graph cache holds. Answer given: `graphcache.Node.Values` holds reconciled values (what rows show); the per-Observation records (`sqlOutcomes` joins) are SQL on every `GetConclusionDetail`, for a handle's own fields too; Swift caches each detail for the session (`sessionFresh`) until a write invalidates it; an invalidated detail that is not on screen is dropped (`whenHidden: .evict`) and reloads on the next visit. Recommendation: filter to the owned property here, plan a `conclusiondetails` `NodeMemo` of composed fields as its own PR.
PRF-1 → user, 2026-10-30: "accept the additional performance cost on the additional SQL calls for right now and then adjudicate it more once we have more of this page built out." **Follow-up** (no property filter here either); PRF-2 rides with it. Facts for the later call: four extra batched SQL rounds per Person detail load, constant in owners; rows grow with every record of each owner (all properties); Swift caches a detail for the session until a write invalidates it.

### Fix plan 1
Order: majors, then minors. Supersedes D-10's open target and D-12 / LW-3's one-group rule (user, 2026-10-30); record both in the plan's Landed note.

**F-1 — One event-Place graph builder** (from MNT-2)
Files: `core/database/conclusionheaders/compose.go`.
Failing test first: none (refactor; existing header and `TestPersonLifeOwners` / `TestPersonDerivedRows` stay green).
Do: extract `eventPlaceGraph(g *graphcache.Graph, v vocab, eventIDs [][]byte) (*placeGraph, error)`: skip nil nodes, `follow(g, id, hops.PlacesOfEvent)` each, `buildPlaceGraph(g, v, placeIDs)`. `composePersons` and `PersonLifeOwners` call it. `composeEvents` calls it too when its Place gather can be separated from its subject hop without changing the order or number of graph queries; if not, say so in the report.
Check: `rg -n "buildPlaceGraph\(" core/database/conclusionheaders/compose.go` prints only the definition and the line inside `eventPlaceGraph` (plus `composeEvents` if it was left, with the reason); scripts/check-pr.sh PASS.

**F-2 — Prove the denied-only date in Go; table the new tests** (from CON-1, CON-2)
Files: `core/database/conclusionheaders/placechains_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`.
Failing test first: `TestPersonLifeOwners/denied-only date` and a `TestPersonDerivedRows` case — a birth event whose only `date` Observation is denied (deny it as the existing denied-name fixture in `conclusiondetails_test.go` does). Expect `birth.Dates` = [{event, date id}], and `PersonDerivedRows` one Birth date owner whose field has one outcome with `Reason == "denied"`.
Do: make `TestFieldsOf` and `TestPersonDerivedRows` table-driven (`tests := []struct{…}`), keeping every existing case.
Check: the new cases pass; scripts/check-pr.sh PASS.

**F-3 — One date owner per event; one "has an Event header" rule** (from SEC-1, MNT-3)
Files: `core/database/conclusionheaders/compose.go` (`lifeOwners`), `placechains_test.go`; `core/database/conclusiondetails/conclusiondetails.go` (`PersonDerivedRows`).
Failing test first: `TestPersonLifeOwners/same event linked twice` — two participation links from the person to one birth event → `birth.Dates` has that event once. If the fixture cannot create two links to one event, report Not reproduced and still dedupe.
Do: dedupe `Dates` by event id the way `Places` uses `listed`. In `PersonDerivedRows`, give the `lists` entries an `events bool` (true for the two date kinds) and read it in both places that compare kinds today.
Check: `rg -n "DerivedRowBirthDate \|\| |== DerivedRowDeathDate" core/database/conclusiondetails` finds nothing; scripts/check-pr.sh PASS.

**F-4 — One toponym-to-Place-title step** (from MNT-1)
Files: `macos/App/Features/Conclusions/PlaceTitleDisplay.swift`, `ReconciledValueRowModel.swift` (`ownerLine` place branch), `PlaceDetailContent.swift` (`init`).
Failing test first: none (refactor); `competingOwnersAreGroupedAndSecondary` and the Place detail tests stay green.
Do: add `PlaceTitleParts(toponym: CatalogConclusionField?, entity: CatalogCanonicalEntity, locale: Locale, orders: NameFormatOrders = NameFormatOrders())` beside `CatalogPlaceHeader.titleParts`, mapping `displayedValues` through `ReconciledValueDisplay.string(for:locale:orders:)` and dropping empty strings. Move both callers onto it.
Check: `rg -n "PlaceTitleParts\(names:" macos/App` prints only the new initializer's body and `CatalogPlaceHeader.titleParts`; scripts/check-pr.sh PASS.

**F-5 — Name an owner that is not shown; keep the Place arrow** (from UX-1, UX-2; user answers)
Files: `ReconciledValueRowModel.swift` (`derived`), `PersonDetailContent.swift` (`rows`), `macos/ProvenenciaTests/PersonDetailTests.swift`, `DerivedPlaceTests.swift`.
Failing test first: (a) `PersonDetailTests` — header `birth.places == []`, `eventCount == 2`, one Birth place owner (York) → one group whose `owner` is set (title "York"), its records `counted == false`. (b) `DerivedPlaceTests` — one Place with no kept name and no parent → the Person life row `opens` that Place, and `shownOwnerID` is nil.
Do: in `derived`, set `owner` on a group when there are two or more groups **or** its owner id is not `shownOwnerID` (one group that is the shown owner keeps no line). In `rows`, `opens = (shown ?? side?.places.first).map(.place)`; `shownOwnerID` stays `shown?.entity.id`. Update `derived`'s doc comment.
Check: both tests pass, existing `lifeRowsCarryTheirOwnersRecords` unchanged; scripts/check-pr.sh PASS.

**F-6 — Docs** (from CON-3, UX-3, user answers)
Files: `docs/deployment-plan/spike-13/deployment-plan.md`, `prs/S13-06-derived-rows-carry-their-records.md`.
Do: `deployment-plan.md:230` — drop "S13-06 is Planned …" from the Order paragraph (it has landed). S13-11's scope entry gains: "VoiceOver names which owner group is behind the shown value (S13-06 UX-3): the group's heading says so, a new L10n key." The plan's Landed note gains: D-10's open target falls back to the first Place when none has a line; D-12 / LW-3: a single group that is not the shown owner gets its group line (user, 2026-10-30).
Check: `rg -n "S13-06 is Planned" docs/deployment-plan/spike-13` finds nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.md` reports no errors.

## Round 1 — fix (blocked)
Head `5f12bc61` (F-6 committed). F-1, F-3 (dedupe only) and part of F-2 are uncommitted edits in `compose.go` and `placechains_test.go`; F-4, F-5 not started. **Blocked:** the implementer's chained `git stash pop` applied an unrelated old stash (`stash@{0}`, s12-03 WIP) and left `core/database/effects/registry.go` and `resolve.go` unmerged with conflict markers. The stash entry is still in the list. This PR's diff does not touch `core/database/effects`. Restoring the two files to HEAD was denied by the permission check; taken to the user.
User approved restoring the two files (2026-10-30); orchestrator ran `git checkout HEAD --` on them. Stash list intact; `go build ./core/...` OK. Round 1 resumes with a new implementer from the uncommitted F-1 / F-2 / F-3 edits.

## Round 1 — fix
Head `7e4a09ca`. Gate (orchestrator): first run FAIL `xcode-test`, again only `SourceGraphFormModelTests` and one `SourcesModelTests` case (same flake family as round 0, files this PR does not touch); rerun PASS. F-1 – F-6 done. Report deviations: `composeEvents` keeps its own `buildPlaceGraph` (its one `g.Nodes(persons+places)` batch would split into two); F-3's test is Not reproduced (the graph cache folds two participation links to one event) — dedupe kept as a guard, SEC-1 stays Fix now (it is code this PR adds, no harm); F-2's denied-only cases pass on old code too (coverage, not red-green); F-4 left `Opens.place`'s `PlaceTitleParts(names:)` (it builds from header strings, not a field); commits group F-1/F-2/F-3 and F-4/F-5 by file.
Minors read by the orchestrator: SEC-1 **Resolved** (`lifeOwners` `dated` map); MNT-3 **Resolved** (`events bool`, no kind comparison left); CON-2 **Resolved** (`tests := []struct` in both); CON-3 **Resolved** (no "S13-06 is Planned"); UX-2 **Resolved** (`opens = (shown ?? side?.places.first)`, `shownOwnerID` still `shown`). UX-3 written into S13-11's Context (`deployment-plan.md:346`).

## Round 2 — review
Reviewers: CON (CON-1), MNT (MNT-1, MNT-2), UX (UX-1). Not sent: SEC+PRF (PRF-1 / PRF-2 are Follow-up by the user; SEC-1 minor, read above). Scope `81fd278f...7e4a09ca`.
CON-1 **Resolved** (`TestPersonLifeOwners/denied-only date`, `TestPersonDerivedRows/denied-only date`). MNT-1 **Resolved** (`PlaceTitleParts(toponym:entity:locale:orders:)`, both callers moved). MNT-2 **Resolved** (`eventPlaceGraph`; `composeEvents` keeps its own gather to save a query round — reviewer agrees). UX-1 **Resolved** (`derived` names any group that is not the shown owner's; `anOwnerThatIsNotTheShownOneIsNamed`). No new blockers from any reviewer. No Fix plan 2.

## Result
Gate PASS (`7e4a09ca`, after one rerun; `SourceGraphFormModelTests` / `SourcesModelTests` flaked on 2 of 4 runs, outside this PR) · `main` unmoved since `588e4a82`, no merge · CI green (`test` 2m36s, `localizable`; push worked on retry after the round-2 review) · open: none · follow-ups: PRF-1 + PRF-2 (user: accept the extra SQL for now, adjudicate once more of the page is built), MNT `placesOfEvents` ID gather shared by `composeEvents`, the Source-model test flake · Later PR: UX-3 → S13-11 · to user: answered (PRF-1, UX-1, UX-2, UX-3) · CON miss: none
