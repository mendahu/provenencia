# Spike 16 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S16-06 — Section switches and removals are spoken

**Landed 2026-11-11** in [#785](https://github.com/mendahu/provenencia/pull/785). Choosing a Person page section with ⌘1–⌘5, a click or Space now says "Timeline shown." (and so on), and the keyboard stays on the selected segment of the switcher when it was in the sections or on the switcher; the segmented control's focus handle lands on the selected segment. Unpromoting one of a person's records that keeps the person says "J. Weber unpromoted."; the leave-or-stay choice after a Remove moved out of the Sources section into the shared Unpromote flow (`UnpromotePageOutcome`), where tests reach it. The header counts "7 records" on every Conclusion page, and the Unpromote confirm ends ", and the person stays" (event, place per kind). Version 0.9.1.

## S16-09 — Conclusion pages refresh from the Sources they cite and show their last value while they refresh

**Landed 2026-11-10** in [#780](https://github.com/mendahu/provenencia/pull/780). Renaming a Source, or fixing a Citation's label or page or an Artifact's label, now reaches the Vitals provenance sheet, the Gallery captions and the viewer's record row on every Person, Event and Place page that cites it. Going Back to a page shows its last value at once with "Refreshing" instead of a spinner: a write marks hidden pages stale rather than dropping them, only an entity a write deleted is dropped, and at most 50 hidden entities keep their pages. The unused `CatalogQueryStalePolicy` is gone. Version 0.8.1.

## S16-04 — The Person page says when a reload failed or the diagram is partial

**Landed 2026-11-10** in [#776](https://github.com/mendahu/provenencia/pull/776). When a Person page's reload fails, the page keeps what it showed and says so with an announced Try again notice where that read failed: under the section switcher for the page's own read, above the rows for a section's read, at the top of the diagram column for the record links read; a first load that fails offers Try again too. A person deleted or merged in another window leaves with "This person no longer exists", also when the page was already shown. `QueryHandle` gains `reading`, `failureRevision` and `isReading`, and `PVRetryCallout(failureRevision:)` speaks again for each new failure and keeps the keyboard on Try again after a press; the Evidence graph and Promote use it. A read past `MaxReadWork` sends no record link, and the diagram says when it is partial (not compared, lines dropped, or more than 1,000 records); in a diagram with lines left out, a record with no line is not spoken as sharing no one. Version 0.7.1.

## S16-35 — Layout tests compute their frames instead of waiting on a window

**Landed 2026-11-10** in [#775](https://github.com/mendahu/provenencia/pull/775). Nothing a researcher sees changes. The Swift tests no longer poll a hosted window for layout, which failed at random when the Mac was busy: `SourcePageOverviewLayout.arrange` and `PVPageColumnsLayout.arrange` are static functions tested with made-up sizes, `PVPageColumnsMetrics.layoutWidths` holds the stacked rule and `groupX` the one centering rule, `PVWidthClassTests` force layout and assert at once, `LeavesForSectionRoot.leave` has a plain test, and `WorkspaceDestinationHostTests`' three whole-workspace tests are deleted. `add-swift-test` teaches the rule. The full suite failed once in five runs before (four tests) and in none of five after.

## S16-34 — Tests run without showing a window or opening the researcher's project

**Landed 2026-11-10** in [#772](https://github.com/mendahu/provenencia/pull/772). Running the Swift tests no longer shows a window or opens the researcher's project: under tests `ProvenenciaMain` runs `TestHostApp` (a `Settings`-only accessory app), test windows are never ordered in, and one `.hostedWindows` trait closes each when its test ends. `PVSegmentedControlTests` (clicks through a window) is deleted, and `PVOutsideClick.isInsideSheet` takes a `parent:` closure so its tests open no real sheets. Five full runs in a row passed, one at a time.

## S16-03 — One open section and one list of section reads

**Landed 2026-11-09** in [#769](https://github.com/mendahu/provenencia/pull/769). Nothing a researcher sees changes. A Person page section declares what it reads once, in `ConclusionPageSection.reads`, each read tagged own, companion or supporting, and `WorkspaceLocation.openPageSection` names the section the page shows, the first one included, so the place registry, the page's "Refreshing" line, the fill and the section hosts cannot disagree. Each section computes its accessibility identifier root once and its views build from it. `ConclusionDetailPage.swift` is split: the gallery model, the rows section and the Place relationship types sit beside their kin. No version bump.

## S16-02 — The Timeline places a missing death and builds what it shows

**Landed 2026-11-09** in [#764](https://github.com/mendahu/provenencia/pull/764). A person with no recorded death and a dated burial or funeral of their own now reads "Death unknown" just before it, in the list and on the lifeline (the wire sends one `place`, after or before a dated row, in place of `beside_event_id`). A long Timeline builds rows and person cards only for what its folds show; folded points stay on the lifeline and share one click target that opens the dated fold through the nearest point's page, and only shown points have a hover name. A date edit that moves the line's years takes a new scale. The Timeline's order moves a row only where the dates cannot say which came first: a burial or funeral of the person's own that the recorded death's dates cannot place before it follows the death, funeral before burial when theirs overlap, and a row the recorded birth's dates overlap follows the birth, while dates that clearly conflict keep date order; the line draws each point at its own date. Version 0.5.1.

## S16-07 — One paged fold that moves focus

**Landed 2026-11-09** in [#761](https://github.com/mendahu/provenencia/pull/761). Every Show more on the Person page's Family, Timeline and Sources reveals one page of at most 50 more and says how many; after a keyboard press the keyboard lands on the first revealed item (Family's tiles now take focus), and a lifeline click on a folded event opens through that event's page only. One fold state (`ConclusionFoldState`) replaces the sections' sets, and one `FocusKeeper` in `macos/App/Platform/` holds the keyboard across a Regular/Wide change for the Gallery and the Timeline. Pointing at a graph whose record line is folded lights up its Show more. Version 0.4.4.

## S16-01 — Cap the Person page reads and an event's places against crafted projects

**Landed 2026-11-09** in [#756](https://github.com/mendahu/provenencia/pull/756). A crafted project can no longer stall the Person page: a person's relatives per role (100, the registry knob `kinship.MaxKinPerRole`), implied siblings through more than ten parents (none), a person's records per Evidence graph (10) and an event's places (20, `eventLocations`) are capped before the work they bound, and `Claims` and `PersonFamily` decide "is this a Person" in `personNode`. The record links read sends at most 75,000 lines, strong first, and says in `not_counted` and `lines_dropped` when it did not count facts or dropped lines; the app's own arc cap is gone. Version 0.4.2.

## S16-10 — One cache registry

**Landed 2026-11-11** in [#789](https://github.com/mendahu/provenencia/pull/789). `graphcache` keeps its derived caches in one registry of typed entries, keyed (one value per node id, filled by the reader) or whole (one value, built on first read), each declaring what a commit does to it, what it is derived from, how long it lives, how Verify checks it and, for keyed entries, its cap. The kinship, `for` and `siblings` answers, the alias index, the name format and the match index move onto it with today's results; the flags, `dropDerived`, the re-warm after a call and the dead `canonicalgraph` package are gone. `docs/catalog-caches.md` lists every Go and Swift cache with its updating signal and its size at 4,000 and 16,000 persons, held to the registry by a test. No change a researcher can see, so no version bump.

## S16-11 — Outcome records live on the canonical node

**Landed 2026-11-12** in [#793](https://github.com/mendahu/provenencia/pull/793). Each canonical node holds its auto-reconciler outcome rows and its chosen likeness, reloaded with the handle and read back by Verify. The Conclusion page's fields and why, the gallery, the lists' lead pictures and the portrait check all compose from the node through one gallery rule (`gallery.Of`); the reconciler and gallery SQL, `gallery.ForEntities` and `sqlPreferred` are gone, and a warm page open reads neither outcome nor depiction tables. Nodes grow by 18.6 MB at 16,000 persons. No change a researcher can see, so no version bump.
