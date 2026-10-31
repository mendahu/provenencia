# Provenance on demand

**Status:** idea only — not roadmapped. A performance option for the Person page if its load gets slow. Raised as S13-06 review PRF-1 and deferred by the user on 2026-10-30; looked at again after S13-11 (#650) and parked on 2026-10-31.

## Today

`GetConclusionDetail` is the one read behind a detail page. For the Person page, it returns every Property's values, and every record behind each value with its outcome, Source, Artifact, Citation and page. It does the same for the fields the life rows read off the person's events and event–place links (`conclusiondetails.PersonDerivedRows`).

- **Cost (S13-06 PRF-1).** Four extra batched SQL rounds per Person page load, the same however many events or Places the person has. The rows returned grow with every record of each owner, across all of its Properties (`sqlOutcomes` has no Property filter). S13-11 added two text columns and one locator decode per record.
- **Not measured.** The Spike 13 map's **Scale** gives no records-per-person or records-per-Place figure.
- **Cache.** Swift keeps a detail for the session until a write clears it. The cost falls on the first open and on each open after an edit.
- **Who uses it.** Most visits never open a provenance sheet. The researcher may be on the way to Family, Timeline or Gallery, and the records are read anyway.

## The idea

Load a row's records only when its provenance sheet opens.

- **The page read keeps a summary per field.** The row needs more than its values: its Source count, "1 record disagrees", the counts on its other values, and whether it has any records at all (which decides whether the control shows). Today Swift works these out from the full record list (`ReconciledValueDisplay.sourceCount`, `againstLine`). Go would return them as counts from one grouped query. That query skips the Citation, Artifact and Source joins, the locator decode, and the date and name lookups.
- **The gallery keeps its lead records.** Each picture tile carries its earliest record (`leadOutcomes`), a few rows.
- **A new read loads one row's records**, keyed by owner and Property, so it also covers life rows and event–place links. The sheet calls it when it opens. The Event and Place pages' Why accordion could call it when it expands.

## What it would take

About the size of S13-11:
- the Go counts query and the new RPC;
- the proto change;
- row models built from counts instead of records;
- a loading state and a failed state in the sheet (the S13-D2 board draws neither);
- fetching again when a write lands while the sheet is open, so it stays true across a refresh (S13-11 US-4);
- `FakeStore` changes;
- rewriting the many Swift tests that build records by hand.

## Before pulling it in

Measure the current read on a heavily researched person (a name with 30 records, a birth with 15) at the map's **Scale**. Do it only if that number shows the page load is too slow. Settle the life-row owners first: S13-11 moved birth and death place rows to their event–place links, and the new read should key on what the rows actually read.

Related: [S13-06 ship log](../deployment-plan/spike-13/prs/S13-06-derived-rows-carry-their-records.ship.md) (PRF-1, PRF-2), [S13-11 ship log](../deployment-plan/spike-13/prs/S13-11-provenance-sheet.ship.md).
