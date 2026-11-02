# Ship log — S13-33 Timeline positions

**Plan:** [`S13-33-timeline-positions.md`](S13-33-timeline-positions.md) · **Branch:** `s13-33-timeline-positions` · **PR:** #687
**State:** ready

Base: `main` @ `3eec6173` (plan written against `51dccd4`).

## Round 0 — implement
Head `68608132`. Gate PASS (orchestrator run: all nine checks). PR #687 (draft).
Report deviations: (1) seed-coverage check is `TestParticipantRolesCoverSeed` in `persondetail`, not `subjectvocab` (import direction); (2) T-8 lead-birth case lives in `TestPersonEventsLifeline`; (3) `TestParticipantOrder` is external on `kinWorld`, ties by ref; (4) structural-debt row is `A-36`; (5) `eventOn` helper in `timeline_test.go`; (6) §10 manual check not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. UX: diff touches `macos/App/Platform/**` (no strings or error codes, dispatched because the path matches). SEC+PRF: `api/ffi/**`, `core/database/**`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | blocker | `TimelineSpan` carries any stored year; S13-13's ticks grow with it | To user | Confirmed: reviewer's probe, `datevalue.Validate` accepts year ±2^31 and `datePlace` passes it to the wire; no year bound in `core/datevalue`. Plan Edges said "Untrusted size. None new". The fix needs a supported-year window (a product call) and changes the write path |
| UX-1 | UX | major | Timeline shows a birth placeholder beside the person's own canonical birth when the header picks an earlier provisional one | To user | Confirmed: `lifePicks` (`compose.go:649`) has no accepted-member check; `recorded` matches only the header pick. Fixing it contradicts the plan's § 3 Out line (logged as A-36) and changes the header a researcher sees |
| UX-2 | UX | minor | A no-year `date` with a `start_date`: the header now shows the start date while the Events list shows the `date` | To user | Confirmed: base `eventDate` returned `date` whenever non-nil; `DateRowDisplay.line` still does. D-2 chose `eventDating` for `eventDate`; `structured-date-model.md` § 3.4 says lists follow that rule, which the Swift list does not |
| MNT-1 | MNT | major | Three hand-built `date → dateKey` copies; `eventDating` drops the place it computed | Fix now → F-1 | verified: `lifeFrom`, `eventKeyOf`, `birthKeyOf` each repeat ref + `datePlace` + key; `eventDating` places and its callers place again (2–3 `DateEnds` calls against D-2's one). Same-pass cleanup in code this PR adds |
| MNT-2 / CON-1 | MNT, CON | minor | Old S13-05 doc comment left above the new `ListPersonEventsResponse` comment | Fix now → F-2 | verified `engine.proto:1565-1574` |
| MNT-3 | MNT | minor | `typeOf` matches the event-type key without origin; `listed` checks origin | Fix now → F-3 | verified: key unique only within an origin (`000030.sql:55`); both in files this PR adds |

Nits (logged): `CON-2 · core/database/persondetail/timeline.go:166 · append into people's spare capacity` · `UX-3 · macos/App/Platform/GenealogyStore.swift:726 · say what both-nil marker means` · `MNT-4 · persondetail/lifeline.go:12 · empty beforeBirth slot unexplained` · `MNT-5 · persondetail/lifeline.go:31 · capitalized named results` · `MNT-6 · persondetail/timeline.go:220 · rename recorded`

User answers (2026-11-02): SEC-1 → bound in Go and on write, a supported-year window in `core/datevalue`, plus a tick-count cap in S13-13 (Fix now → F-4, F-6). UX-1 → fix in this PR, supersedes the § 3 Out line and closes A-36 (Fix now → F-5). UX-2 → keep D-2, correct § 3.4, log aligning `DateRowDisplay` as a Follow-up (Fix now → F-7 for the doc; Follow-up for Swift).

### Fix plan 1

**F-4 — A supported-year window, held on read and on write** (from SEC-1)
Files: `core/datevalue/datevalue.go`, `core/datevalue/validate_test.go`, `core/database/conclusionsummaries/compose.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`. Skill: `add-date-value`.
Failing test first: `TestEventSpan` cases "year above the window", "year below the window", "range from -2^31 to 2^31" → undated (no span, sorts in the undated tail); `validate_test.go` cases for a point and a range end with year 10000 and -10000 → `ErrInvalid`.
Do: in `core/datevalue` add `const MinYear, MaxYear = -9999, 9999` (exported, documented once as the years a stored date may hold) and `func YearInWindow(y int) bool`. `validateComponentRanges` returns `ErrInvalid` for a year outside it (the existing error, no new code or copy). In `datePlace`, an end whose `Year` is set and outside the window has no place: the date is not ok when either closed end is out of the window. Keep `datePlace` the one place this is checked on read.
Check: the tests pass; scripts/check-pr.sh PASS.

**F-1 — One `dateKeyOf`; `eventDating` returns the place it computed** (from MNT-1)
Files: `core/database/conclusionsummaries/compose.go`, `order.go`.
Failing test first: none (refactor); `TestEventSpan`, `TestEventOrderAtKnownEnd`, `TestSortByBirth` stay green.
Do: add `dateKeyOf(v *datevalue.Value, ref string) dateKey` beside `dateKey` (lowercases ref, calls `datePlace`); `lifeFrom`, `eventKeyOf` and `birthKeyOf` call it. Change `eventDating` to return the chosen value with its place (a small unexported struct: value, whether it is the start, key, span, ok) so `eventKeyOf`, `EventSpan` and `eventDate` read it instead of placing again: one `DateEnds` call per chosen date (D-2). Check: `grep -n "datePlace(" core/database/conclusionsummaries/*.go | grep -v _test` prints only `datePlace`'s definition, `dateKeyOf` and `eventDating`.
Check: scripts/check-pr.sh PASS.

**F-5 — The header's birth and death pick canonical events only** (from UX-1, MNT-3's `lifePicks` note)
Files: `core/database/conclusionsummaries/compose.go`, tests in `conclusionsummaries_test.go` or `persondetail/timeline_test.go` (wherever a person can be joined to events, as T-8 did), `docs/ideas/structural-debt.md`, the plan's Landed note.
Failing test first: `TestPersonEventsLifeline/provisional earlier birth and accepted later birth` — a provisional-only birth 1818 and an accepted birth 1820: the summary's `Birth.Event` is the 1820 event and `Timeline.Birth.Event` is the 1820 row (not a placeholder). Same shape for death.
Do: in `lifePicks`, skip an event with no accepted members, by the same predicate `persondetail`'s `canonical` uses (`len(n.AcceptedMembers()) > 0`); if the predicate can live once in a shared place both packages already import, put it there and point `persondetail.canonical` at it, otherwise reuse the expression and name the other copy in a comment. Match the event type through the product-origin rule of F-3. Remove the A-36 row from `structural-debt.md` (or mark it resolved, by the file's convention). In the plan's Landed note, say the § 3 Out line on `lifePicks` was superseded by the user on 2026-11-02 (UX-1).
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — One product-key rule for terms in `persondetail`** (from MNT-3)
Files: `core/database/persondetail/timeline.go`, `participants.go`, tests.
Failing test first: `TestLifeMarkers` or `TestPersonEventsLifeline` case: a user-origin `event_type` keyed `burial` after the death is not treated as after-death (death's `Beside` skips it only when it is the product's term).
Do: add `productKey(t *conclusionsummaries.Term) string` (empty unless `t.Origin == vocabkey.OriginProvenencia`); `typeOf` and `listed` both use it.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — One doc comment on `ListPersonEventsResponse`** (from MNT-2, CON-1)
Files: `api/proto/engine.proto`, regenerated `engine.pb.go`, `engine.pb.swift`.
Do: merge the S13-05 paragraph into the new one, keeping the order sentence and "Canonical events and participants only."; `./scripts/generate-proto.sh`.
Check: one comment above the message; scripts/check-pr.sh PASS.

**F-6 — S13-13 caps the tick count** (from SEC-1; Later PR)
Files: `docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` D-13 and its Tests.
Do: in D-13's **Ticks**, add: if even the 50-year step gives more than 200 ticks, the step grows by powers of ten (100, 200, 500, 1000, …) until it gives at most 200; add a test row to S13-13's Tests for a placed range of 9999 years drawing at most 200 ticks. Note in its text that the engine bounds years to `datevalue.MinYear…MaxYear` (S13-33).
Check: `scripts/check-plan-docs.py` 0 errors.

**F-7 — § 3.4 says what the Events list shows** (from UX-2)
Files: `docs/structured-date-model.md` § 3.4.
Do: the rule is the engine's (`conclusionsummaries`: order, span, header and Persons list dates); say that the Events list row (`DateRowDisplay`) still shows the kept `date` whenever one is set, an open follow-up. Add the window from F-4: a year outside `MinYear…MaxYear` has no place.
Check: scripts/check-pr.sh PASS.

Order: F-4, F-1, F-5, F-3, F-2, F-6, F-7. Nits stay logged; fix CON-2 and MNT-5 only if the line is already being edited.

Follow-ups (round 1): UX-2 Swift — align `DateRowDisplay.line` with the engine's field choice (a no-year `date` with a placed `start_date`); suggested home: `docs/ideas/` or the next Events-list PR.

## Round 1 — fix
Head `04b191b6`. Gate PASS (orchestrator run). F-4, F-1, F-3, F-5, F-2, F-6, F-7 done. Deviations: F-3 before F-5 (F-5 uses F-3's rule); F-1's `eventDated` struct carries the key, `birthKeyOf` uses `dateKeyOf`; F-5 repeats the accepted-members predicate with a comment naming `persondetail.canonical`; AC-2.9's last assertion inverted per UX-1; `Canon.UserTerm` added to `graphcachetest`; A-36 row removed.
MNT-2 / CON-1 (F-2): **Resolved** — one `ListPersonEventsResponse is` comment above the message (`f6929f9f`).

## Round 2 — review
Reviewers: SEC+PRF (SEC-1), MNT (MNT-1), UX (UX-1). Not sent: CON (no Fix-now blocker or major). Scope: `68608132...04b191b6`.

SEC-1: **Resolved** (SEC+PRF round 2: write and read bound, 500,000-case seeded probe, F-6 cap in S13-13). MNT-1: **Resolved**. UX-1: **Resolved**.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-7 | MNT | blocker | Canonical-event rule and product-key rule each in two packages after F-5, though `graphcache` / `conclusionsummaries` are shared | Fix now → F-8 | verified: `persondetail/family.go:24` `canonical` and `compose.go:673` inline `len(ev.AcceptedMembers()) == 0`; `participants.go:21` `productKey` and `compose.go:677` inline origin check. `graphcache.Node` has `Is` and `AcceptedMembers` (`graph.go:146,161`), imported by both packages. Fix plan 1's condition for the comment fallback was not met |
| SEC-3 | SEC (outside lens → MNT) | minor | S13-13 D-13 says a placed range is at most 19998 years; it is 19999 (-9999.0 to 10000.0) | Fix now → F-9 | verified by the `To` rule (end of year 9999 is 10000.0) |

### Fix plan 2

**F-8 — One canonical rule and one product-key rule** (from MNT-7)
Files: `core/database/graphcache/graph.go`, `core/database/persondetail/family.go`, `timeline.go`, `participants.go`, `core/database/conclusionsummaries/compose.go` (and wherever `Term` is declared).
Failing test first: none (refactor); `TestPersonEventsLifeline` (all cases, including "provisional earlier birth and accepted later birth" and "a researcher's burial term is not the product's"), `TestParticipantOrder`, `TestPersonEvents` stay green.
Do: add `func (n *Node) Canonical(typeID []byte) bool { return n.Is(typeID) && len(n.AcceptedMembers()) > 0 }` to `graphcache`, documented as the one rule for a canonical entity; delete `persondetail.canonical` and call the method at its call sites. `lifePicks` calls `ev.Canonical(<event type id vocab already holds>)`. Add `func (t *Term) ProductKey() string` (nil-safe; empty unless `Origin == vocabkey.OriginProvenencia`) beside `Term` in `conclusionsummaries`; delete `persondetail.productKey` and call the method in `typeOf` and `listed`; `lifePicks` gets its type key through the same rule (build the type `Term` through the `termOrigin` lookup `finishTerm` already uses, or a small unexported helper that `Term.ProductKey` also calls). Remove the cross-reference comment.
Check: `rg -n "len\([a-z]+\.AcceptedMembers\(\)\) [>=]" core --type go | grep -v _test` prints only `graphcache/graph.go`'s `Canonical`; `rg -n "OriginProvenencia" core/database/persondetail core/database/conclusionsummaries --type go | grep -v _test` prints `ProductKey` (and its helper if any) plus any hits that existed at base `3eec6173` (list them in the report); scripts/check-pr.sh PASS.

**F-9 — S13-13's range bound** (from SEC-3)
Files: `docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md:411`.
Do: "at most 19998 years" → "at most 19999 years (-9999.0 to 10000.0)".
Check: `scripts/check-plan-docs.py` 0 errors.

## Round 2 — fix
Head `10cefdef`. Gate PASS (orchestrator run). F-8 done (`93c03fa3`): `graphcache.(*Node).Canonical` and `conclusionsummaries.(*Term).ProductKey` are the one rule each; `persondetail.canonical` and `productKey` deleted; `lifePicks` calls both (now also skips a merged event node, as the other callers do). F-9 done (`10cefdef`).
MNT-7: **Resolved** (orchestrator check: `rg "len\([a-z]+\.AcceptedMembers\(\)\) [>=]"` prints only `graphcache/graph.go:170`; no `func canonical` / `func productKey` left). SEC-3: **Resolved**. No round 3 (cap).

Merged `origin/main` @ `8918bc27` (S14-17 / S14-27 plans, docs only) → head `0ec3bba0`. Gate PASS against `8918bc27`.

## Follow-ups
- UX-2 Swift: align `DateRowDisplay.line` with the engine's event-date field choice (a no-year `date` with a placed `start_date`); suggested home `docs/ideas/` or the next Events-list PR. Noted in `structured-date-model.md` § 3.4.

## Result
Gate PASS · CI green · open: none · follow-ups: UX-2 (Swift) · to user: SEC-1, UX-1, UX-2 (answered 2026-11-02) · CON miss: none

CI: first run, Go `test` failed `promotealign` `TestLargeSourceBudgets/warm` (385.8ms, budget 350ms) at `fd7a44de`. Not this PR: local warm timing ×3 is base `8918bc27` 106–114ms, head 101–109ms. Re-ran the failed job: green.
