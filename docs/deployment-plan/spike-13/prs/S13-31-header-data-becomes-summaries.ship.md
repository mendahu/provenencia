# Ship log — S13-31 Header data becomes summaries

**Plan:** [`S13-31-header-data-becomes-summaries.md`](S13-31-header-data-becomes-summaries.md) · **Branch:** `s13-31-header-data-becomes-summaries` · **PR:** #660 (stacked on #659, base `s13-02-gender-in-the-person-header`)
**State:** ready

## Round 0 — implement
Base `s13-02-gender-in-the-person-header` @ `016134b9` (stacked on #659, user 2026-11-01; `main` @ `c7d19d1e` is already merged into it).
Head `b438e15f`. Gate PASS (orchestrator run). PR #660 (draft, base `s13-02-gender-in-the-person-header`). Report deviations:
1. Four code commits instead of one per step: Go core (STEP-1–3), wire (STEP-4), Swift (STEP-5–7), docs; the steps depend on each other inside each layer.
2. `headerHits` in the Evidence card files kept: it holds `EvidenceCardHeaderActionHits` frames (layout), so D-2/D-3's type rule keeps it; D-3's list named it by mistake.
3. Added `conclusionsummaries.NewSummaries` and `Want.Empty`.
4. `titleIsTheHeadersChosenTitle` → `titleIsTheSummaryChosenTitle`.
5. Band-describing comments kept (`placechains.go:299`, `conclusiondetails.go:127`, a proto comment, `GenealogyStore.swift:663`).
6. §10 manual check not run in the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — `macos/App/**` routes UX; `api/ffi/**` routes SEC; `core/database/**` routes PRF. Base `016134b9`, head `b438e15f`.
First MNT reviewer stopped before reviewing (its context carried an unrelated Docs-service instruction block from the session's MCP servers, not from the repo: `git grep 'Claude Docs'` finds nothing); a fresh MNT reviewer ran in its place.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-X1 (UX outside lens) | CON | major | `bandLineReadsTheSummary` builds the `date` field with no records, so AC-1.6's "records but no displayed value" edge is never exercised | Fix now → F-1 | verified: `EventDetailTests.swift:123` passes `[]` outcomes |
| UX-1 | UX | minor | In D-13's edge, the band shows a span the Date row never shows or backs | To user → answered (2026-11-01): Date row follows the band → F-2 | `EventDetailContent.dateRow` takes the `date` branch whenever `outcomes` is non-empty; `ReconciledValueRowModel.lead` is nil then |
| MNT-1 | MNT | major | Three living-doc lines still say "header" for the data | Fix now → F-3 | verified: `macos-client-patterns.md:101` ("per-entity header key"), `catalog-deletes.md:46` ("header dependents"), `docs/ideas/promote-ranking.md:16` ("header display rows") |
| MNT-2 | MNT | major | `PersonsFromQuerier`, `EventsFromQuerier`, `PlacesFromQuerier` have no callers, a second compose-without-storing path beside `RowsForSearch` | Fix now → F-4 | verified: `git grep FromQuerier` at head hits only their definitions in `compose.go` |
| MNT-3 | MNT | minor | "A kind with no ids is not read" lives in `ByIDs` paths and in `Want.Empty` + a searchindex check | Fix now → F-5 | file this PR adds; `RowsForSearch` has no per-kind skip |
| MNT-4 | MNT | minor | Four callers repeat `x.hasSummary ? x.summary : nil` before `mapConclusionSummary` | Fix now → F-6 | the generated getter returns an empty message whose `kind` is nil; the mapper already maps nil |
| MNT-5 | MNT | minor | No living doc names the mixed-kind summary pattern | Fix now → F-7 | `git grep ConclusionSummary` in docs/skills prints nothing |
| UX-2 | UX | minor | `secondaryDate`/`secondaryPlace` doc comments describe the old source | Fix now → F-8 | stale comment in a changed file |
| MNT-8 | MNT | nit | `ConclusionDetailPage.swift:65-66` says "header path" for the summary's picture path | Fix now → F-8 | stale comment in a changed file (D-7) |

CON: clean (2 minors noted: `docs/promote-graph-alignment.md` renamed though not listed in D-8, consistent with its intent; an Event detail with no summary shows no band date, not reachable since Event detail opens only Provenencia Events). SEC+PRF: clean.

Nits (logged): `MNT-6 · compose.go:245-254, 289-296 · Summaries receiver and NewSummaries loop vars still named h` · `MNT-7 · conclusion_summaries.go:25,34,136,145,209,218 · h / ph / eh locals left from header naming`

### Fix plan 1

**F-1 — The band-line test exercises AC-1.6's edge** (from CON-X1)
Files: `macos/ProvenenciaTests/EventDetailTests.swift`.
Failing test first: none needed (the case passes today; it must test the right input).
Do: in `bandLineReadsTheSummary`'s second case, give the `date` field at least one record (an outcome) and no displayed value, with the same span summary; keep the expected span. Use the existing `D.field`/`D.value` helpers the other tests use for a recorded but undisplayed value.
Check: the test passes; scripts/check-pr.sh PASS.

**F-2 — The Date row follows the band's rule** (from UX-1; user, 2026-11-01)
Files: `macos/App/Features/Conclusions/EventDetailContent.swift` (`dateRow`), `macos/ProvenenciaTests/EventDetailTests.swift`.
Failing test first: `EventDetailTests.dateRowFallsToTheSpanWhenTheDateShowsNothing` *(new)* — a `date` field with a record and no displayed value, plus kept `start_date` 1849 and `end_date` 1851 fields → `rows[0].lead == L10n.Dates.rowSpan(start: "1849", end: "1851", locale: en)`, `rows[0].records.map(\.id)` contains the start and end ids, and `rows[0].lead == page.secondaryDate`.
Do: in `dateRow`, use the `date` field when it has a displayed value; otherwise, when the start or end field has a displayed value, build today's span row (`.spanning`); otherwise keep today's behavior (the `date` field's row when it has records, else the empty row). `dateFieldWinsOverASpan`, `spanLeadWhenThereIsNoDate`, `missingDateIsAnEmptyDateRow` keep their expectations. Record the rule in the plan's D-13 Landed note.
Check: the new test passes; scripts/check-pr.sh PASS.

**F-3 — Finish the living-doc rename** (from MNT-1)
Files: `docs/macos-client-patterns.md:101` ("per-entity summary key"), `docs/catalog-deletes.md:46` ("summary dependents", each occurrence), `docs/ideas/promote-ranking.md:16` ("summary display rows").
Check: `git grep -niP 'header (dependents|rows?|display|read|key|composer|walk)' -- ':(glob)docs/*.md' docs/ideas ':!docs/ideas/archive' .agents .cursor` prints nothing.

**F-4 — Delete the dead `*FromQuerier` readers** (from MNT-2)
Files: `core/database/conclusionsummaries/compose.go`.
Do: delete `PersonsFromQuerier`, `EventsFromQuerier`, `PlacesFromQuerier`; `RowsForSearch` is the one compose-without-storing read. Add them to the plan's Landed note as deleted.
Check: `git grep -n FromQuerier -- core api` prints nothing; go test PASS.

**F-5 — One "skip an empty kind" rule** (from MNT-3)
Files: `core/database/conclusionsummaries/compose.go` (`RowsForSearch`, `Want.Empty`), `core/database/searchindex/handles.go` (`loadHandleSummaries`).
Do: `RowsForSearch` skips each kind with no ids, as `ByIDs` does through `PersonsByIDs`/`EventsByIDs`/`PlacesByIDs`; delete `Want.Empty` and its check in `loadHandleSummaries`.
Check: `git grep -n 'Empty()' -- core/database/conclusionsummaries core/database/searchindex` prints nothing new; searchindex and conclusionsummaries tests PASS.

**F-6 — One presence rule in the Swift mapper** (from MNT-4)
Files: `macos/App/Platform/GoStore.swift` (`mapConclusionSummary` and its four callers).
Do: make the parameter non-optional; callers pass `x.summary`; the mapper's `kind == nil` case returns nil.
Check: `git grep -n 'hasSummary ?' -- macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-7 — Document the mixed-kind summary pattern** (from MNT-5)
Files: `docs/conclusion-reconciliation.md` (the `conclusionsummaries` code-map row) and the `ConclusionSummary` comment in `api/proto/engine.proto` (regenerate).
Do: one sentence: a message that carries a summary of any kind has one `ConclusionSummary` field, built by `conclusionSummaryProto` and mapped by `mapConclusionSummary`; mixed-kind reads go through `Want.Add` and `ByIDs`.
Check: scripts/check-pr.sh PASS (guidance-names).

**F-8 — Stale comments** (from UX-2, MNT-8)
Files: `macos/App/Features/Conclusions/EventDetailContent.swift:11-14` (say the date and place come from the summary's kept date or span and its places), `macos/App/Features/Conclusions/ConclusionDetailPage.swift:65-66` ("summary's picture path").
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `8e31cfd5` (one commit for F-1–F-8; the implementer left it on a detached HEAD and unpushed, so the orchestrator fast-forwarded the branch to it and pushed). Gate PASS (orchestrator run). All F-# done. Deviations: the F-2 test was written with the fix and not run red first; the F-2 rule is Landed note item (6); `completed.md` gained the Date-row sentence; MNT-6/7 nits untouched.
- F-2 (UX-1, minor): **Resolved** — orchestrator read `EventDetailContent.dateRow`: a displayed `date` → that field; else a displayed start or end → the span row; else a `date` with records → that field; else empty.

## Round 2 — review
Reviewers: CON (had CON-X1, major), MNT (had MNT-1, MNT-2, majors). Not re-sent: UX (minor only; resolved above), SEC+PRF (clean). Scope `b438e15f...8e31cfd5`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-X1 | CON | — | (round 1) band-line test edge | **Resolved** | round-2 CON: `bandLineReadsTheSummary` now gives the `date` field a kept outcome and no displayed value |
| MNT-1, MNT-2 | MNT | — | (round 1) docs; dead `*FromQuerier` | **Resolved** | round-2 MNT and CON: both greps print nothing |
| MNT-9 | MNT | blocker | The Date row re-derives the engine's date-or-span choice in Swift from `detail.fields` instead of reading the summary the init already holds | Fix now → F-9 | verified: `EventDetailContent.init` passes the summary to `placeRow(summary)` and the band but `dateRow(detail.fields, …)` decides with `dateValue`/`spanShows`; the summary's rule is `oneEvent` (`compose.go:881-922`) |

CON round 2: clean. Notes (not findings), riding with F-9 as docs: the F-7 sentence sits in the "Place chains at a date" row of `conclusion-reconciliation.md`; the plan's Landed note says "five places" but lists six.

### Fix plan 2

**F-9 — The Date row reads the summary's choice** (from MNT-9)
Files: `macos/App/Features/Conclusions/EventDetailContent.swift` (`dateRow`, `init`), `macos/ProvenenciaTests/EventDetailTests.swift`; docs ride-along: `docs/conclusion-reconciliation.md`, the S13-31 plan's Landed note.
Failing test first: `EventDetailTests.dateRowFollowsTheSummaryWhenFieldsDisagree` *(new)* — the summary carries only a start 1849 and end 1851, while the `date` field has a displayed value 14 May 1817 → `rows[0].lead == L10n.Dates.rowSpan(start: "1849", end: "1851", locale: en)` and `rows[0].lead == page.secondaryDate`. Run it red against the current `dateRow` first.
Do: pass the summary into `dateRow`, as `placeRow(summary)` gets it. Branch on it: `summary?.date != nil` → the `date` field's row; else `summary?.startDate != nil || summary?.endDate != nil` → today's `.spanning` row from the start and end fields (their records stay the Why); else the `date` field's row when it has records, else the empty row. Delete the `dateValue(date) != nil` / `spanShows` decision. `dateRowFallsToTheSpanWhenTheDateShowsNothing` and the three D-13 tests keep their expected values (their fixtures give the summary the same dates as the fields; add the summary where one is missing). Docs: move the F-7 sentence into the `conclusionsummaries` package row (or its own row) of `conclusion-reconciliation.md`'s code map; fix the Landed note's count to match its list and record that the Date row reads the summary.
Check: the new test fails before and passes after; `git grep -n 'spanShows' -- macos/App` prints nothing; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `63a1f2c3` (on the branch and pushed). Gate PASS (orchestrator run). F-9 done: `dateRowFollowsTheSummaryWhenFieldsDisagree` ran red first (lead "14 May 1817", expected the span), passes after; `spanShows` and both `dateValue` helpers gone. Deviation: the span row's lead reads `summary.startDate`/`endDate` (not the fields), so no part of the choice is re-derived; the Why still lists the start and end fields' records. Orchestrator read the diff: matches F-9. Docs ride-along done (own code-map row; Landed note "six places").

## Finish
`origin/main` @ `c7d19d1e` and `s13-02-gender-in-the-person-header` @ `016134b9` are both ancestors of the head; no merge needed.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: UX-1 (answered: Date row follows the band) · CON miss: CON-X1 (round-1 CON passed `bandLineReadsTheSummary` though its `date` field had no records; UX caught it)
