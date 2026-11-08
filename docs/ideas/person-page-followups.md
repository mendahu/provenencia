# Person page follow-ups

**Status:** idea only — not roadmapped. Low-priority Person page items from the [2026-11-08 audit](../deployment-plan/spike-16/person-page-audit.md), parked by the user the same day. One needs a product call before anyone builds it (marked). Pull one in when the code it names next changes, or as a batch of an odds-and-ends spike.

## Birth and death event counts have no plural (P-18)

`provenanceSheet.recordsFromBirthEvents` and `…DeathEvents` read `%#@records@ from %2$lld birth events`: the records count has a plural substitution, the events count does not (its sibling `recordsFromPlaces` has both). English is right because the count is always two or more (`PVProvenanceSheetContent.swift` 325); languages with few and many forms are not. `scripts/check-localizable-xcstrings.py` (96) skips any entry with a substitution, so the bare count passes. **Shape:** add an `events` substitution, and make the check test each `%n$lld <word>` that is not itself a substitution.

## Back resets a section's place (P-19) — product call

Walking the family or the timeline person to person and pressing Back lands at the top of the previous person's section, with its folds closed and nothing focused: the page is keyed by entity id (`PlaceRegistry.swift` 200), folds are view state, and `WorkspaceLocation` holds no scroll position for a section (only list roots keep `listScrollAnchor`, `WorkspaceNavigation.swift` 233). The fold half was raised for Sources in S13-16 (UX-5) and declined by the user; the scroll and focus half was never asked. **Shape, if wanted:** a per-section anchor (the last opened row or tile) restored and focused on Back, as `listScrollAnchor` does for lists.

## The Conclusion lists hide a refresh failure too (map review K-2)

`ConclusionListPresentation` (`ConclusionListPage.swift` 5–24), behind the Persons, Events and Places lists, is built from `QueryHandle.read` and `isFetching` like the detail page was, so a failed reload behind a list keeps the old rows and says nothing, and its error `PVCallout` (158–159) has nothing to press. Spike 16's S16-04 adds a `QueryHandle` reading with the refresh failure beside the value and moves the Conclusion detail page onto it; the lists adopt the same reading and the kit's `PVRetryCallout`. It changes chrome on three list pages, so it needs a brief.
