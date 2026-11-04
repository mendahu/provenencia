# A cap on the records a provenance table draws

**Status:** idea only — not roadmapped. The user accepted this risk on 2026-11-03 (S13-36 review SEC-1). Pick it up when the Event and Place pages adopt the provenance sheet.

## Today

`PVProvenanceRecordTable` draws every record of a value at once, on the main thread, and nothing caps the count: `sqlOutcomes` has no `LIMIT`, and the assembler maps every outcome. A sheet opens on a click, so the wait comes after the researcher asks for it, never on project open or page load.

| Records on one value | First layout, Debug build, three runs |
| --- | --- |
| 30 | 172, 157, 155 ms |
| 300 | 1595, 1713, 1825 ms |
| 1,200 | 8704, 10693, 12155 ms |

- **Not measured.** A Release build; the old `Grid` table, so there is no baseline; the Why accordion, which reads the same records in its own `Grid`; the assembly, which these timings start after.
- **Assembly.** It is O(n²) for negative records: `deniedSomething` and the denier lookup in `ReconciledValueDisplay` each scan `field.outcomes`.
- **Exposure.** A catalog from another tool can hold any count, so a crafted project can freeze the app for as long as its count makes it. The likelier case is ordinary: a Place's name row reads every Observation on that Place, so it grows with the catalog. The stated ceiling for now is about 300 records on one value.

## The idea

- **Fold past a fixed count** with a "Show all" control, as the Timeline folds at 50 events (S13-13). The table keeps its `VStack`, so the Tab order holds for the rows shown. This is a visible change past the cap, so it needs a board note and one string. It belongs on the shared table, and the Why needs the same cap.
- **Or make the rows lazy** with `LazyVStack`, as `PVList` and `PVTable` do. It needs no chrome, but Tab order through rows that do not exist yet must be checked in the app: S13-14's focus work shows how that fails.
- **Either way,** build one set of denying Observation ids per field in `PVProvenanceSheetContent.records(_:)`, so the assembly is linear.

## Before pulling it in

Take a Release figure at 300 and 1,200 records, and count the records on the largest Place name in a real project.

Related: [S13-36 ship log](../deployment-plan/spike-13/prs/S13-36-provenance-sheet-in-the-kit.ship.md) (SEC-1, F-11), [`provenance-on-demand.md`](provenance-on-demand.md).
