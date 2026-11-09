# Provenance loading and drawing

**Status:** Spike 15 design input (pulled in from `docs/ideas/provenance.md` on 2026-11-09, user). Two performance follow-ups from Spike 13, now PRs in [`deployment-plan.md`](deployment-plan.md): the cap is **S15-27** (brief S15-D9 for the sheet, S15-D11 for the Why accordion); loading records when the sheet opens is **S15-28** (engine) and **S15-29** (app), growing the on-open call S15-21 adds for citations. The measurements each section asks for are the first step of those PRs' plans.

## Cap the records a provenance table draws

**Spike 15** (S15-21) shows each record's formatted citation and groups the records that share one, so the table draws fewer rows when a household's Citations group, but the rows it draws are longer. Measure the cap with grouped, formatted rows.

The user accepted this risk on 2026-11-03 (S13-36 review SEC-1).

`PVProvenanceRecordTable` draws every record of a value at once, on the main thread, and nothing caps the count: `sqlOutcomes` has no `LIMIT`, and the assembler maps every outcome. A sheet opens on a click, so the wait comes after the researcher asks for it, never on project open or page load.

| Records on one value | First layout, Debug build, three runs |
| --- | --- |
| 30 | 172, 157, 155 ms |
| 300 | 1595, 1713, 1825 ms |
| 1,200 | 8704, 10693, 12155 ms |

- **Not measured.** A Release build; the old `Grid` table, so there is no baseline; the Why accordion, which reads the same records in its own `Grid`; the assembly, which these timings start after.
- **Assembly.** It is O(n²) for negative records: `deniedSomething` and the denier lookup in `ReconciledValueDisplay` each scan `field.outcomes`.
- **Exposure.** A catalog from another tool can hold any count, so a crafted project can freeze the app for as long as its count makes it. The likelier case is ordinary: a Place's name row reads every Observation on that Place, so it grows with the catalog. The stated ceiling for now is about 300 records on one value.

**Shape:** Fold past a fixed count with a "Show all" control, as the Timeline folds at 50 events (S13-13). The table keeps its `VStack`, so the Tab order holds for the rows shown. This is a visible change past the cap, so it needs a board note and one string. It belongs on the shared table, and the Why needs the same cap. **Or** make the rows lazy with `LazyVStack`, as `PVList` and `PVTable` do — no chrome, but Tab order through rows that do not exist yet must be checked (S13-14). Either way, build one set of denying Observation ids per field in `PVProvenanceSheetContent.records(_:)`, so the assembly is linear.

**Before pulling it in:** Take a Release figure at 300 and 1,200 records, and count the records on the largest Place name in a real project.

Related: the S13-36 ship log (in git with #702), SEC-1 and F-11.

## Load records when the sheet opens

Raised as S13-06 review PRF-1 and deferred by the user on 2026-10-30; looked at again after S13-11 (#650) and parked on 2026-10-31.

`GetConclusionDetail` is the one read behind a detail page. For the Person page, it returns every Property's values, and every record behind each value with its outcome, Source, Artifact, Citation and page. It does the same for the fields the life rows read off the person's events and event–place links (`conclusiondetails.PersonDerivedRows`).

- **Cost (S13-06 PRF-1).** Four extra batched SQL rounds per Person page load, the same however many events or Places the person has. The rows returned grow with every record of each owner, across all of its Properties (`sqlOutcomes` has no Property filter). S13-11 added two text columns and one locator decode per record.
- **Not measured.** The Spike 13 map's **Scale** gives no records-per-person or records-per-Place figure.
- **Cache.** Swift keeps a detail for the session until a write clears it. The cost falls on the first open and on each open after an edit.
- **Who uses it.** Most visits never open a provenance sheet. The researcher may be on the way to Family, Timeline or Gallery, and the records are read anyway.

**Since Spike 16 Batch 2** (S16-11, S16-30) the records compose from memory (the node's outcome rows, Subject records, Source records) instead of SQL, so this idea is only about the size of the page read's payload; measure that before pulling it in.

**Spike 15 takes the first step** (2026-11-09), and S15-28 and S15-29 finish it. Formatted citations are fetched when the provenance sheet opens, not with the page: S15-21 asks one batch render call for the sheet's Citations, which reads SQL in a few batched queries and caches nothing ([`deployment-plan.md`](deployment-plan.md) rule "Citations render on demand, from SQL"; S15-16 measures it at 300 Citations). So the sheet already gains a load-on-open call, its loading and failed states (S15-D9), and the "a write lands while the sheet is open" question for citation text. The records themselves still arrive with the page read until S15-28 and S15-29 move them onto the same on-open call; S15-21's call is the one that grows, not a second one.

**Shape:** Load a row's records only when its provenance sheet opens.

- The page read keeps a summary per field (Source count, "1 record disagrees", counts on other values, whether any records exist). Go returns those as counts from one grouped query that skips Citation / Artifact / Source joins, locator decode, and date and name lookups.
- The gallery keeps its lead records (`leadOutcomes`).
- A new read loads one row's records, keyed by owner and Property, so it also covers life rows and event–place links. The Event and Place pages' Why accordion could call it when it expands.

**What it would take** (about the size of S13-11): the Go counts query and new RPC; proto; row models from counts; loading / failed states in the sheet (S13-D2 draws neither); refetch when a write lands while the sheet is open; `FakeStore`; rewriting Swift tests that build records by hand.

**Before pulling it in:** Measure the current read on a heavily researched person at the map's **Scale**. Do it only if that number shows the page load is too slow. Settle the life-row owners first: S13-11 moved birth and death place rows to their event–place links.

Related: the S13-06 ship log (in git with #640), PRF-1 and PRF-2; the S13-11 ship log (in git with #650).
