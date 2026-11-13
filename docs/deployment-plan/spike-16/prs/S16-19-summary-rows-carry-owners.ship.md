# Ship log — S16-19 Summary composers read the place entries and keep what readers re-derive

**Plan:** [`S16-19-summary-rows-carry-owners.md`](S16-19-summary-rows-carry-owners.md) · **Branch:** `s16-19-summary-rows-carry-owners` · **PR:** #821
**State:** ready

## Round 0 — implement
Head `bd0b20e27`. Gate PASS (orchestrator run; go-test cached). Report deviations: AC-2.3 test cites the toponym with `f.Cite` (not `f.citeToponyms`, which `f.Reported` does not hear); the derived-rows query-count case calls `nameformats.Invalidate` before each count (name format loaded once per connection); Steps 5 and 6 in one commit; `source_read.go` gofmt carried into the STEP-7 commit; `func entityOf` sweep still matches two unrelated functions (`core/subjectprofile/invariants_test.go`, `core/database/effects/path.go`); one full-suite run failed `TestLargeSourceBudgets` warm/cold over 2.5 s under load, passes alone (0.90 s / 0.76 s), no budget changed; manual check not run (engine refactor, no UI).

STEP-1 and STEP-8 figures (host `Jakes-M1-Macbook-Pro.local`):

| Figure | Before | After |
| --- | --- | --- |
| First `part_of` from the city | 99.9, 118.4, 143.6 ms (median 118.4) | 310.2, 97.7, 89.3 ms (median 97.7) |
| Reindex of 18,242 handles | 5.07, 7.32, 6.03 s (median 6.03) | 8.26, 4.82, 5.12 s (median 5.12) |
| `catalog_search_inputs` rows | 118,860 | 118,860 |
| `displays-B` at 4,000 / 16,000 | 22,157,576 / 31,435,112 | 24,620,488 / 35,090,984 (+3.7 MB at 16,000) |
| `displays-ms` at 4,000 / 16,000 | 551 / 621 | 199 / 336 |

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF (`core/database/**`, `api/ffi/handlers/**`). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC (+PRF) | major | Stored person rows keep persons × location-links owners, uncapped | Fix now → F-1 | verified: `compose.go` `lifeFacts` appends one owner and one `&ChainedPlace` per event–place link for every person, kept in `displays` by `ListPersons`; reviewer's probe at 400×400 → 160,000 owners, 40.9 MB retained vs 552 KB at base. Plan §6 Edges recorded "uncapped at base and here", but at base nothing was retained, so this PR makes the cost larger. Spike 16 rule "Crafted fan-out is capped, not engineered" (user, 2026-11-08) sets the fix shape |
| PRF-1 | PRF | minor | List fill follows each shared life event's locations once per person (P×K time, at base too) | Follow-up | not made worse (≈160 ms at 400×400 at base and head); the code shape predates this PR |
| MNT-1 | MNT | minor | Timeline dedupes event ids by hand beside `database.UniqueBlobIDs` | Fix now → F-2 | same-pass cleanup in a loop this PR adds (`timeline.go:62-68`) |
| MNT-2 | MNT | minor | `neighborRoles` is a second type its one caller converts to `EventParticipant` | Fix now → F-3 | code this PR moves; `compose.go:987` converts each one straight away |
| MNT-3 | MNT | minor | `PersonDerivedRows` repeats `fieldsOf`'s nil/merged skip | Fix now → F-4 | `fieldsOf` drops nil/merged nodes (`conclusiondetails.go:197`), so `fields[o.ID]` is empty and the check never changes output |
| CON-1 | CON (+SEC outside lens) | minor | Landed note says no ship log was committed; a stub is committed | Fix now → F-5 | `git show 767223df1 --stat`; the orchestrator keeps the ship log on the branch, with the STEP-1/STEP-8 figures above |
| CON-2 | CON | minor | No Timeline test writes a participation after the event row is stored | Fix now → F-6 | `rg expectEvents timeline_test.go`: no write between reads; the Timeline now reads stored rows |
| CON-5 (outside lens) | PRF | minor | Large fixture files only `subject` roles, so STEP-8 cannot show D-7's input growth | Follow-up | growth is one input row per non-subject participation row (linear in catalog rows); changing `bulkCanon` moves every benchmark baseline, outside this PR's Scope |

Nits (logged): `MNT-4 · conclusiondetails.go:663 · copyPlace clones Parents that place owners never have` · `MNT-5 · timeline.go:131 · inline the one-use canonicalPerson closure` · `CON-3 · placechains_test.go:652 · 0x7a sentinel → wantNoRow bool` · `CON-4 · timeline_test.go:669 · STEP-7 sweep matches the AC-3.4 test itself; note it`

### Fix plan 1

**F-1 — Cap the owners a life row keeps, and share one place per folded place** (from SEC-1)
Files: `core/database/conclusionsummaries/compose.go`, `core/database/conclusionsummaries/placechains_test.go` (beside `TestCraftedEventPlaces`), `docs/deployment-plan/spike-16/deployment-plan.md` (rule "Crafted fan-out is capped, not engineered", its caps list), the plan's Landed note.
Failing test first: `TestCraftedLifeOwners` (new, table-driven). One birth event with P person subjects and K `Location` links to one place, with K at least 20 × the cap (bar 2). `ListPersons` then asserts that each person's `len(Owners.Dates)+len(Owners.Places)` is at most `maxLifeOwners`, that the kept owners are the first ones in today's order (date order, then link order), and that the call stays under a time bound in the style of the existing crafted tests. Red at head on the owner count.
Do: add `const maxLifeOwners = 64` beside `maxEventPlaces`, with a doc comment and `// decided: user, 2026-11-08 (Spike 16 rule "Crafted fan-out is capped, not engineered"): a real life row has one to three events with one to three places each`. In `lifeFacts`, count date owners and place owners together, and stop appending both once the count reaches the cap. Check before each append, so nothing past the cap is allocated; past the cap, drop silently (the rule). AC-2.4's 21 filings → 21 owners must still pass. Build one `*ChainedPlace{Entity, Names}` per folded place `p` (outside the links loop) and point every owner of that place at it; `PersonDerivedRows` already copies through `copyPlace`. Do not change `facts.Places`, the seeds, or what `eventLocations` reads. Add the cap to the rule's caps list in `deployment-plan.md` as one bullet ("a life row's owners, 64, applied in `lifeFacts` (S16-19)"), and record in the Landed note that it supersedes the plan §6 Edges line "Owners per life row … uncapped" (SEC-1).
Check: `TestCraftedLifeOwners`, `TestCraftedEventPlaces`, `TestPersonRowOwners` and `TestPersonDerivedRows` pass; report the largest owner count per life row on the large fixture (it must be far below 64); scripts/check-pr.sh PASS.

**F-2 — Dedupe the Timeline's event ids with `database.UniqueBlobIDs`** (from MNT-1)
Files: `core/database/persondetail/timeline.go`.
Failing test first: none (refactor; `TestPersonEvents*` and `TestDropUnreadEvents` hold the behavior).
Do: collect `l.Neighbor` for each link and set `eventIDs := database.UniqueBlobIDs(ids)`; delete `seenEvent`.
Check: `rg -n "seenEvent" core/database/persondetail` prints nothing; the persondetail tests pass.

**F-3 — `rolesByNeighbor` returns `[]EventParticipant`** (from MNT-2)
Files: `core/database/conclusionsummaries/events.go`, `compose.go`, `placechains_internal_test.go`.
Failing test first: none (refactor; `TestRolesByNeighbor*` and `TestEventParticipants` hold the behavior).
Do: return `[]EventParticipant`, copying ids as the caller copies them today, with the role-seen bookkeeping as a local map. Delete `neighborRoles` and the conversion loop in `oneEvent`, and update the internal tests' expectations to `EventParticipant`.
Check: `rg -n "neighborRoles" core` prints nothing; tests pass.

**F-4 — Drop `PersonDerivedRows`' second nil/merged skip** (from MNT-3, with nit MNT-4)
Files: `core/database/conclusiondetails/conclusiondetails.go`.
Failing test first: none (the merged, unknown and non-person cases of `TestPersonDerivedRows` hold it).
Do: delete `if node == nil || node.Merged { continue }`. `fieldsOf` already drops those nodes, so the field loop never runs for them. Keep `nodeOf` for the `FromNode` lookup. In `copyPlace`, drop the `Parents` clone if no owner can carry parents; otherwise keep it and say why in its comment.
Check: tests pass.

**F-5 — Make the Landed note match the committed ship log** (from CON-1)
Files: the plan's Landed note.
Do: replace the bullet "No ship log was committed …" with one saying the ship log is committed on the branch and carries the STEP-1 and STEP-8 figures. Do not edit the ship log itself; the orchestrator owns it.
Check: plan-docs PASS in scripts/check-pr.sh.

**F-6 — The Timeline reads again after a participation changes on a stored event row** (from CON-2)
Files: `core/database/persondetail/timeline_test.go`.
Failing test first: add cases to `TestPersonEventsReadsParticipantsFromRows`, or to `TestPersonEvents`. Read the Timeline, which stores the event row. Join a new person to the event with a non-subject role, read again, and expect the new participant in `sortParticipants`' order. Add a second case that changes an existing witness's role. Both are expected green today, as guards. If either is red, that is a real defect: report it before fixing it.
Check: passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `34e35071c`. Gate PASS. F-1 done (`TestCraftedLifeOwners` red first at 1,281 owners, then green; largest owner count per life row on the 4,000-source fixture is 1). F-2–F-6 done; F-6's two guards green as expected. Three fix commits lack the Co-Authored-By trailer (not amended).

## Round 2 — review
Reviewers: SEC+PRF (SEC-1 major). CON and MNT not sent back: their findings were all minors.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | Uncapped life-row owners | Resolved | reviewer: `maxLifeOwners` checked before each append in `lifeFacts`; one shared `ChainedPlace` per folded place; `TestCraftedLifeOwners` 1,280 filings → 64 |
| MNT-1, MNT-2, MNT-3 (+MNT-4) | MNT | minor | F-2, F-3, F-4 | Resolved | orchestrator read `bd0b20e27..34e35071c`: `UniqueBlobIDs` in `timeline.go`; `neighborRoles` gone, `rolesByNeighbor` returns `[]EventParticipant`; the nil/merged skip and the `Parents` clone gone |
| CON-1, CON-2 | CON | minor | F-5, F-6 | Resolved | Landed note matches the committed ship log; `TestPersonEventsFollowsParticipationWrites` added |

No new blockers.

## Finish
Merged `origin/main` @ `64a91aeb7` (S15-34) → `61c6664ce`. Gate PASS.

## Result
Gate PASS · CI pending · open: none · follow-ups: PRF-1, CON-5 · to user: —
