# Spike 16 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

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
