# Review — S13-05 Person timeline read

**Verdict:** round 1 — Fits with changes / Ready with changes
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet).

## To user

- **M-1: one role or every role per row?** The working rule "Timeline is direct participation, any role" names one role term. Answered (user, 2026-11-02): **all roles**. D-5 carries a list for the person and for each participant; **Rules this plan bends** records it; STEP-9 rewords the rule, the map row and the scope entry.

## Writer's own calls

- One row per event even when the person has several links to it (D-5).
- Participants in participation order, the order `EventSummary.Subjects` uses (D-7); the board may ask for another order in a follow-up.
- "Undated" is what the event order puts last: no date and no start date with a sort key; an end date alone is undated (D-8).
- `Dated` lives on the timeline row, not on `EventSummary` (M-6 declined).
- The canonical rule covers events as well as people (D-3).
- `ConclusionPageSection.timeline.queryKey` is S13-13's (S13-12 plan D-8).

## Round 1 — executability

### X-1 · blocker · AC-1.7 (a participation with no role) cannot be built
**Answer:** Fixed — D-13: `Canon.Link` skips its term when `termProp == ""`; AC-1.7 uses `k.Link("participation", …, "", "")` (folded with M-5).

### X-2 · major · Tie order depends on the ref counter (`tst-10` < `tst-9`)
**Answer:** Fixed — D-13: tie tests read the refs back and derive the expected order from them.

### X-3 · major · Drop and sort have no stated order
**Answer:** Fixed — D-9 states the order: loads, summaries, rows in `EventsByIDs`' order, `dropUnreadEvents`, then `Dated`.

### X-4 · minor · Fixture event types and date properties
**Answer:** Fixed — D-13: `vocabkey.EventType("census")` and the like, `PropertyStartDate` / `PropertyEndDate`; no new constants.

### X-5 · minor · An Event's id at the wire
**Answer:** Fixed — AC-3.2 and T-5 add an Event's id.

### X-6 · minor · STEP-9 does not say what the map row gets; "role" is singular in the map
**Answer:** Fixed — STEP-9 names the Plan link, the status and the "roles" rewording (with M-1).

### X-7 · minor · Three places in `invalidationsComeFromRegistrySpecs`
**Answer:** Fixed — T-9 names the two lists and the trigger loop.

## Round 1 — maintainability

### M-1 · blocker · The roles list bends the working rule, but "Rules this plan bends" says None
**Answer:** To user → answered: all roles (2026-11-02). Rules this plan bends filled; STEP-9 updates the rule.

### M-2 · major · `shown` collides with the local slice in `kinOf`, and "canonical" is the domain word
**Answer:** Fixed — D-3: `canonical(n, typeID)`; the local `shown` keeps its name.

### M-3 · major · Use `EventsByIDs`' list order instead of exporting a second sort entry point
**Answer:** Fixed — D-6, D-8, D-9: `EventsByIDs` + `PersonsByIDs` + `NewSummaries`; rows in `EventsByIDs`' order; `sortEvents` and `compose.go` untouched; `SortByEventDate` dropped; STEP-2 shrinks to `HasEventDate`.

### M-4 · major · One helper for links → (neighbor, roles)
**Answer:** Fixed — D-5: `rolesByNeighbor` and `neighborRoles`, used by both walks; `TestRolesByNeighbor` in T-4.

### M-5 · major · Make `Canon.Link`'s term optional instead of a near-copy
**Answer:** Fixed — D-13: `Link` skips the term when none is given; `Location` is one `Link` call. Replaces the round-1 `Canon.Bridge` answer to X-1.

### M-6 · minor · "Dated" on `EventSummary` instead of the row
**Answer:** Declined — no other `EventSummary` reader groups by it, and the summary is a stored display row sent on every list wire; `HasEventDate` keeps the rule in one place for when one does (D-8).

### M-7 · minor · `typeIDs` only forwards
**Answer:** Fixed — D-4: `personTypeID` deleted; each read calls `g.Vocabulary()` once.

### M-8 · minor · Go `Participant` vs proto and Swift `EventParticipant`
**Answer:** Fixed — D-9: Go type `EventParticipant`.

### M-9 · minor · The `EventSubject` fold is "Logged" with no log entry
**Answer:** Fixed — STEP-9 adds a row to `docs/ideas/structural-debt.md`.

### M-10 · minor · Per-person keys need hand-kept test lists; S13-27 adds the next
**Answer:** Fixed (logged) — STEP-9 adds a second `structural-debt.md` row, touched with S13-27.

## Round 2 — executability

### X-8 · major · STEP-9 says the map already links the plan; it does not
**Answer:** Fixed — the map row links the plan at planning (wire step); STEP-9 says so.

### X-9 · major · AC-2.2's fixture does not say what keeps the batches equal
**Answer:** Fixed — AC-2.2: every event located, every co-participant with a dated, placed birth, in both (a) and (b).

### X-10 · minor · How to build a date with no sort key
**Answer:** Fixed — T-1 names a point `datevalue.Value` with no year.

### X-11 · minor · The scope entry names `ListPersonEvents` as the Go function
**Answer:** Fixed — STEP-9 rewords it to the RPC over `persondetail.PersonEvents`.

### X-12 · minor · "no SQL" vs graph queries counted
**Answer:** Fixed — §1 says "no graph query".
