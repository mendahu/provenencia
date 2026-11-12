# Review — S16-11 Outcome records live on the canonical node

**Verdict:** Fits with changes — Ready with changes (round 1; no round 2: every blocker was Fixed without a reshape)
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- **The memory line (plan §12).** The map's "under 6 MB more" assumed far fewer rows than the large fixture holds (102,458 values at 16,000 persons, counted at the base), so outcomes on the node are est. 15–18 MB there, on top of caches already at ~118 MB against the rule's 40 MB. Recommendation: record before and after, stop at 20 MB of growth, and restate "Memory is budgeted" in the map as a ceiling that follows the catalog's size. **Answered (user, 2026-11-11):** agreed; the map's rule now reads 250 MB at 16,000 persons, and a PR may warm a cache at open when it measures the open cost and the task-to-task gain.

## Writer's own calls
- No registry entry: outcomes and the pointer are node fields reloaded by `Set.Handles` (D-13).
- Verify reads outcomes back from the table instead of recomputing them; values' truth check already catches a skipped recompute (D-4).
- `SetPreferredLikeness` reads its previous pointer from the node; `sqlPreferred` goes (D-9).
- No new speed budget except "clear the portrait at the city", which measures the write's one-shot node load at the 6,000-event place.
- `LEFT JOIN observations` for the Subject, so the node holds every table row and the fixture's synthetic outcomes load (D-2).

## Round 1

### M-1 · blocker · The node's outcome row reuses the reconciler's in-memory type
`autoreconcile.Outcome.Value` is an index into a reconcile result, not a stored rank, and has no Subject; every reader would convert ±1.
**Answer:** Fixed — D-1 adds `graphcache.Outcome`, a stored-row type beside `Value` and `Member`, 1-based `ValueRank`, string `Reason`, `SubjectID`, vote ints; no `autoreconcile` import in `graphcache`. D-5, D-6, T-1 follow.

### M-2 · major · D-2 drops the map's Subject on an unmeasured cost; S16-30 and S15-28 need it
**Answer:** Fixed — D-2 keeps the Subject (`LEFT JOIN observations`), as the map and design say; STEP-9 no longer edits them for it; the Out row went.

### M-3 · major · `ReadableImage` loses its outside caller; the skill's "every reader goes through Readable" reads as broken
**Answer:** Fixed in part — STEP-9 adds the skill line saying the cached image was read through `Readable` when written. Declined for the export: `Readable` itself calls `ReadableImage` (`core/depictionvalue/depictionvalue.go:41`), so it is live; unexporting it is a tidy outside §6.

### M-4 · minor · Third `leadPictures` caller (Places)
**Answer:** Fixed — §4, D-8 and Go core name `composePlaces` (~line 1237), which passes `need`. Props are resolved once per call, which already gives one vocabulary lookup per batch.

### M-5 · minor · Key the interim record read by member Subjects
**Answer:** Fixed — D-6 reads `sqlRecords` by the wanted rows' Subject ids with `INDEXED BY observations_subject_id_idx` and keeps the wanted Observations; §6 Scale updated.

### M-6 · minor · The write's one-shot graph at the 6,000-event city is unmeasured
**Answer:** Fixed — §6 Scale names the city case; STEP-1 adds "clear the portrait at the city" with a 500 ms budget, measured before and after.

### M-7 · nit · `gallery.Of(n, gallery.PropertyIDs(v))` repeated
**Answer:** Logged.

### X-1 · major · Three `leadPictures` callers, not two
**Answer:** Fixed — as M-4.

### X-2 · major · How a gallery lead's outcome is built from the node
**Answer:** Fixed — D-6: a lead's row is found by scanning that node's gallery-Property rows; a lead with no row or no record is skipped; fields copy across with no conversion (D-1).

### X-3 · major · The 6 MB bound disagrees with the fixture
**Answer:** To user, answered 2026-11-11 (20 MB stop; map rule 250 MB) — counted: 90,445 values at 4,000 and 102,458 at 16,000. Plan §12 carries the question; AC-3.1, STEP-8 and the Stop condition point at it.

### X-4 · minor · T-9 and `catalogsession.Do`
**Answer:** Fixed — counters installed in one `Do`, the second open made outside it, copying `person_detail_test.go` ~133.

### X-5 · minor · T-2 / T-3 need a real Observation id
**Answer:** Fixed — `SELECT id FROM observations WHERE subject_id = ? LIMIT 1` for the Canon entity.

### X-6 · minor · Figure names
**Answer:** Fixed — AC-3.2 lists the logged labels; STEP-1, STEP-8 and the DoD refer to them.

### X-7 · minor · How to report `nodes-ms`
**Answer:** Fixed — STEP-1 wraps only the nodes call; `measure` unchanged.

### X-8 · minor · `loadProperties`' `q`
**Answer:** Fixed — D-6 drops it.

### X-9 · minor · Map row and status
**Answer:** Fixed — the map row links the plan with status Draft until §12 is answered.

### X-10 · minor · `check-pr.sh` runs Xcode tests
**Answer:** Declined — the DoD's check-pr line is the repo's gate; the implementer runs it where Xcode is (the ship loop's macOS step), and this PR changes no Swift.
