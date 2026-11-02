# Review — S13-33 Timeline positions

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet). No round 2: every blocker was Fixed with the reviewer's own better shape, and its premise checked in the code (below).

## To user

None. The one product question (a "bef."-only date) was answered before review (2026-11-02: on the line, one meaning of "dated").

## Writer's own calls

- D-2: `dateKey.key` becomes an `int` day number from `datePlace`, and `conclusionsummaries` stops calling `autoreconcile.SortKey`. Checked first: `dateKey` is compared only by `dateKeyLess` inside the package; `keptDateAt`'s key is read only by `eventDate` → `lifeFrom`; nothing compares these keys with the reconciler cache's `sort_key`; `SortKey`'s date key is `DateWindow(d).Lo`, which is `EndsWindow(first, last).Lo` = the first end's `Lo` when closed, so a closed-start date keeps its exact order.
- D-2: `eventDating` picks `Date` when it has a place, else `StartDate`, for both the event order and the header's life date. Today the header reads a `Date` with no year over a dated `StartDate`; after, it reads the `StartDate`. Accepted as one rule (§ 11 Risks).
- D-6: a header birth that is not canonical becomes a placeholder on the timeline (AC-2.9), rather than the next canonical birth: the header and the timeline name the same event or the timeline names none.
- D-10: a participant with no role ranks last; unlisted roles order by `SortByTitle` over their labels.

## Round 1

### M-1 · blocker · The markers re-derive the header's pick instead of reading it
**Answer:** Fixed. D-6: `PersonEvents` adds `entityID` to its existing `PersonsByIDs` batch and takes `PersonSummary.Birth.Event` / `.Death.Event` when that event is a timeline row; `lifeMarkers` computes only `Beside`. The first-row pick, its equality test and its Stop condition are gone. The `lifePicks` canonical gap is logged to `docs/ideas/structural-debt.md` in STEP-6 and listed in § 3 Out. Query count unchanged (one id in an existing batch).

### M-2 · blocker · "Dated", the order key and the span come from separate paths
**Answer:** Fixed. D-2: `datePlace` from one `DateEnds` call gives the key and the span; `dateSort`, `eventKeyOf`, `birthKeyOf` and `lifeFrom` key through it; `eventDate` and `eventKeyOf` share `eventDating`; `EventSpan` returns the same span; `autoreconcile.SortKey` is untouched and no longer imported by `compose.go`. Premise verified in the code (Writer's own calls).

### M-3 · blocker · `Span` added beside `Dated`
**Answer:** Fixed. D-4: `PersonEvent.Dated` and `HasEventDate` are deleted (`TestHasEventDate` folds into `TestEventSpan`); proto field 4 is reserved; Swift's `dated` is computed from `span`, so S13-13 reads the same name. The map's working rule "Timeline is direct participation, any role" no longer names `HasEventDate` (planning commit). AC-1.7 now holds by type.

### M-4 · major · `participantRoleOrder` has no completeness check
**Answer:** Fixed. T-2 adds the reverse check: every seeded `role` term is in `participantRoleOrder`.

### M-5 · major · The unlisted-role tie-break has no collation
**Answer:** Fixed. D-10: unlisted roles order by `SortByTitle` over their labels (keys as refs), the package's one title order. T-9 has a case-insensitive case.

### M-6 · minor · Participant order in `lifeline.go`
**Answer:** Fixed. `participants.go` and `participants_test.go`; `lifeline.go` keeps the table and `lifeMarkers`.

### M-7 · minor · One concept, three names
**Answer:** Fixed. One Go type, `conclusionsummaries.TimelineSpan`, returned by `EventSpan` and held by `PersonEvent.Span`; proto `TimelineSpan`; Swift `CatalogTimelineSpan`.

### M-8 · minor · Birth and death compared as string literals
**Answer:** Fixed. D-7: `lifePicks` uses `vocabkey.EventBirth` / `EventDeath`.

### M-9 · minor · The new "dated" rule has no home in the domain docs
**Answer:** Fixed. STEP-6 and the DoD add the display date rule to `docs/structured-date-model.md`.

### M-10 · nit · AC-3.2 guards an older engine that cannot ship
**Answer:** Fixed. AC-3.2 now states the defaults of an empty `CatalogPersonTimeline`.

### X-1 · major · Participant order leaves three cases open
**Answer:** Fixed. D-10: no role ranks last; several roles take the best rank; ties keep `PersonsByIDs`' order (ref included), so "then by ref" is dropped.

### X-2 · major · Fixtures cannot build the cases
**Answer:** Fixed. STEP-3 adds `kinWorld.eventOn(name, typ, *datevalue.Value)` with `datevaluetest` values; AC-2.2's marriage has a date (1840); T-7 reads the refs back and expects the lower ref first.

### X-3 · major · The span formula only by examples
**Answer:** Fixed. D-2 states the three cases (both closed, open last, open first).

### X-4 · minor · Checks that cannot run here
**Answer:** No change needed: the header's **Runs on** says so; § 6 FFI now notes STEP-1 – STEP-3 do not need the regenerated proto.

### X-5 · minor · The handler with an empty marker
**Answer:** Fixed. D-8: the handler always sends both markers; T-5 asserts the empty `event_id`.

### X-6 · minor · S13-13's plan still names the old value type
**Answer:** Fixed. STEP-6 and the DoD sweep say S13-13's STEP-0 refreshes it; this PR does not edit S13-13's plan.

### X-7 · nit · Bookkeeping
**Answer:** Fixed. D-10 names the four role constants; T-2's text is de-duplicated; STEP-3's check names T-3, T-4 and T-9.
