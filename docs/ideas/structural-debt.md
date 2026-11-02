# Structural debt

**Status:** parked 2026-10-29 when Spike 12 closed. Refactors from the first app-health audit, now at [`deployment-plan/spike-14/app-health-audit.md`](../deployment-plan/spike-14/app-health-audit.md), with no user-visible change. Pull one in when a feature next changes that code.

| Item | Was | What |
| --- | --- | --- |
| A-31 | S12-36 | Search kinds should own their columns and locations: `KindSpec` names each field's column and carries its location builder; `Reproject` and `RebuildAll` read a projector table; a completeness test. Touch it with the next searchable kind ([`add-searchable-kind`](../../.agents/skills/add-searchable-kind/SKILL.md)). |
| A-32 | S12-37 | The Evidence graph canvas keeps its tool and sheet state as loose flags; one tool enum and one sheet enum would replace them. Touch it with the next canvas tool. |
| A-33 | S13-05 | Two event-person shapes: `EventSummary.subjects` (`EventSubject`, from `hops.SubjectsOfEvent`: subject role, not filtered to canonical) and `PersonEvent` participants (`EventParticipant`, from `hops.ParticipantsOfEvent`: canonical, with roles). Fold them when an Event page lists participants. |
| A-34 | S13-05 | The per-person query keys `.personFamily` and `.personEvents` (and S13-27's) repeat one descriptor shape; share one descriptor builder, and derive `everyKindHasOneDescriptor`'s count from `Kind.allCases`. Touch it with S13-27. |
| A-36 | S13-33 | `conclusionsummaries.lifePicks` reads a person's subject-role birth and death events without filtering to canonical, so the header can name an event the Timeline drops (the timeline shows a placeholder, S13-33 AC-2.9). Filter in Go, per the working rule "Canonical only, filtered in Go", when the header next changes. |
| — | S14-11 plan | `CatalogQueryStalePolicy`: no descriptor arm sets `.refetchOnRevisit` and only two tests read `stalePolicy(for:)`, beside the cache's `isStale` and `whenHidden: .markStale`. Remove or use it when the cache next changes. |
| — | S12-15 ship log | Swift's `SourceSurface` raw values and Go's `workspacelocation.Surface*` constants are matched only by a comment; a `swift_vocabulary_test.go` row would catch drift. |
