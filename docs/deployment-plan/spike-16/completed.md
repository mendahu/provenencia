# Spike 16 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S16-02 — The Timeline places a missing death and builds what it shows

**Landed 2026-11-09** in [#764](https://github.com/mendahu/provenencia/pull/764). A person with no recorded death and a dated burial or funeral of their own now reads "Death unknown" just before it, in the list and on the lifeline (the wire sends one `place`, after or before a dated row, in place of `beside_event_id`). A long Timeline builds rows and person cards only for what its folds show; folded points stay on the lifeline and share one click target that opens the dated fold through the nearest point's page, and only shown points have a hover name. A date edit that moves the line's years takes a new scale. Version 0.5.1.

## S16-07 — One paged fold that moves focus

**Landed 2026-11-09** in [#761](https://github.com/mendahu/provenencia/pull/761). Every Show more on the Person page's Family, Timeline and Sources reveals one page of at most 50 more and says how many; after a keyboard press the keyboard lands on the first revealed item (Family's tiles now take focus), and a lifeline click on a folded event opens through that event's page only. One fold state (`ConclusionFoldState`) replaces the sections' sets, and one `FocusKeeper` in `macos/App/Platform/` holds the keyboard across a Regular/Wide change for the Gallery and the Timeline. Pointing at a graph whose record line is folded lights up its Show more. Version 0.4.4.

## S16-01 — Cap the Person page reads and an event's places against crafted projects

**Landed 2026-11-09** in [#756](https://github.com/mendahu/provenencia/pull/756). A crafted project can no longer stall the Person page: a person's relatives per role (100, the registry knob `kinship.MaxKinPerRole`), implied siblings through more than ten parents (none), a person's records per Evidence graph (10) and an event's places (20, `eventLocations`) are capped before the work they bound, and `Claims` and `PersonFamily` decide "is this a Person" in `personNode`. The record links read sends at most 75,000 lines, strong first, and says in `not_counted` and `lines_dropped` when it did not count facts or dropped lines; the app's own arc cap is gone. Version 0.4.2.
