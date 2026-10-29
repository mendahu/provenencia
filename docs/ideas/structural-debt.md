# Structural debt

**Status:** parked 2026-10-29 when Spike 12 closed. Refactors from the first app-health audit, now at [`deployment-plan/spike-14/app-health-audit.md`](../deployment-plan/spike-14/app-health-audit.md), with no user-visible change. Pull one in when a feature next changes that code.

| Item | Was | What |
| --- | --- | --- |
| A-31 | S12-36 | Search kinds should own their columns and locations: `KindSpec` names each field's column and carries its location builder; `Reproject` and `RebuildAll` read a projector table; a completeness test. Touch it with the next searchable kind ([`add-searchable-kind`](../../.agents/skills/add-searchable-kind/SKILL.md)). |
| A-32 | S12-37 | The Evidence graph canvas keeps its tool and sheet state as loose flags; one tool enum and one sheet enum would replace them. Touch it with the next canvas tool. |
| — | S12-15 ship log | Swift's `SourceSurface` raw values and Go's `workspacelocation.Surface*` constants are matched only by a comment; a `swift_vocabulary_test.go` row would catch drift. |
