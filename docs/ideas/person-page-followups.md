# Person page follow-ups

**Status:** idea only — not roadmapped. Low-priority Person page items from the [2026-11-08 audit](../deployment-plan/spike-16/person-page-audit.md), parked by the user the same day. Two need a product call before anyone builds them (marked). Pull one in when the code it names next changes, or as a batch of an odds-and-ends spike.

## Hovering a node whose record line is folded highlights nothing (P-15)

The record links diagram and the Sources rows highlight each other (Spike 13 working rule "The view is a circular link diagram"), but only lines that are drawn take the highlight (`PersonSourcesSection.swift` 356–365). A Source with more than five of the person's graphs folds its sixth line on (`PersonSourcesContent.swift` 13–17), so hovering or focusing that graph's node on the circle shows nothing in the list. A parish register filed one graph per entry reaches it. **Shape:** highlight that row's "Show N more" when the graph's line is folded (the content already maps folded claims to graphs in `claimGraphs`), or unfold on focus.

## Wording on the Person page (P-17) — product call

- The Unpromote confirm says "…, and the handle stays" (`unpromote.confirmMessage`, `L10n.swift` about 8387). It is the only user-facing string that says "handle"; on the Person page what stays is the person. Per kind, like `confirmMessageDeletes`, would fix it; the copy was agreed as RC-2 in S13-17.
- The header counts "7 members" (`conclusions.memberCount`, the only "member" in the UI) while Sources and the confirms call the same things records.
- A deleted or merged person shows "This record no longer exists" (`conclusionDetailsNotFound`), where "record" now names the person on a page where it names a claim. A Person-page line ("This person no longer exists…") would separate them.

## Birth and death event counts have no plural (P-18)

`provenanceSheet.recordsFromBirthEvents` and `…DeathEvents` read `%#@records@ from %2$lld birth events`: the records count has a plural substitution, the events count does not (its sibling `recordsFromPlaces` has both). English is right because the count is always two or more (`PVProvenanceSheetContent.swift` 325); languages with few and many forms are not. `scripts/check-localizable-xcstrings.py` (96) skips any entry with a substitution, so the bare count passes. **Shape:** add an `events` substitution, and make the check test each `%n$lld <word>` that is not itself a substitution.

## Back resets a section's place (P-19) — product call

Walking the family or the timeline person to person and pressing Back lands at the top of the previous person's section, with its folds closed and nothing focused: the page is keyed by entity id (`PlaceRegistry.swift` 200), folds are view state, and `WorkspaceLocation` holds no scroll position for a section (only list roots keep `listScrollAnchor`, `WorkspaceNavigation.swift` 233). The fold half was raised for Sources in S13-16 (UX-5) and declined by the user; the scroll and focus half was never asked. **Shape, if wanted:** a per-section anchor (the last opened row or tile) restored and focused on Back, as `listScrollAnchor` does for lists.

## The Timeline keeps an old scale after a date edit (P-20)

`PersonTimelineList.rebuild` (`PersonTimelineSection.swift` about 139) keeps the chosen scale while the item ids and the viewport height stay the same. A date edit from another window keeps the ids, so the old points-per-year is applied to the new range: a short life drawn dense, then an event moved decades away, gives a line many times its usual height. **Shape:** key the kept scale on the line's year range too.

## The Conclusion lists hide a refresh failure too (map review K-2)

`ConclusionListPresentation` (`ConclusionListPage.swift` 5–24), behind the Persons, Events and Places lists, is built from `QueryHandle.read` and `isFetching` like the detail page was, so a failed reload behind a list keeps the old rows and says nothing, and its error `PVCallout` (158–159) has nothing to press. Spike 16's S16-04 adds a `QueryHandle` reading with the refresh failure beside the value and moves the Conclusion detail page onto it; the lists adopt the same reading and the kit's `PVRetryCallout`. It changes chrome on three list pages, so it needs a brief.

