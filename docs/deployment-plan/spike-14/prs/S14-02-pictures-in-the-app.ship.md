# Ship log — S14-02 Pictures in the app

**Plan:** [`S14-02-pictures-in-the-app.md`](S14-02-pictures-in-the-app.md) · **Branch:** `s14-02-pictures-in-the-app` · **PR:** #641
**State:** fixing 1

## Round 0 — implement

Base `c54f34bf` (`main`). Dispatched `pr-implementer`.
Head `1a6da81c`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). PR #641 (draft). Report: deviations — two earlier gate runs failed (a test that never set `leads`, fixed; a `SourcesModelTests.listReadsEachGraphsProgress` timeout in untouched code, passed on rerun); AC-3.1 compares against today's `rowBoundAnnouncement(count:)` ("bound to 1 types"); `GalleryTileView` takes `fallbackMark` as an init argument; extra test `theLeadCopyIsReadAtCallTime`; archived brief links gained one `../`; L10n, xcstrings and pbxproj grouped in the STEP-1 commit; older plans and map entries that describe the rename were left as they are.

## Round 1 — review

Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`. SEC+PRF: `macos/App/Features/**` models (lead model, list rows, gallery tiles) iterate over catalog-loaded depictions.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Choosing a thumbnail or going back to the earliest is silent for VoiceOver | Fix now → F-1 | verified: `ConclusionLeadModel.write` sets `toast` only; `PVToast` posts no announcement; `AccessibilityAnnouncement.post` has three callers, none here |
| MNT-1 | MNT | major | Event and Place contents build the same gallery block twice; two `LeadCopy` statics differ only in `menuTitle` | Fix now → F-2 | verified: `EventDetailContent.swift:35-52` and the Place block differ in four values. Same-pass cleanup. D-4's named statics stay (a Decision); one is built from the other |
| MNT-2 | MNT | major | `spokenOrigin` copies `OriginPill`'s plugin wording | Fix now → F-3 (plugin branch) | verified: `L10n.Origin.pluginPill(pluginID: CatalogOrigin.pluginID(from:))` at `PropertiesSnapshot.swift` and `OriginMarkers.swift:35`. The in-file `user` / `provenencia` branches stay: each already reads one key, and `OriginPill` shows different words for `provenencia` ("Seeded by Provenencia" vs the cell's "seeded"), so there is no single word table to share |
| CON-1, MNT-3 | CON, MNT | minor | New method splits `canEditCardinality`'s doc comment; not beside `deleteAccessibilityLabel` (D-11) | Fix now → F-4 | verified: `PropertiesModel.swift:170-188` |
| UX-2 | UX | minor | Keyboard focus is lost after "Use as thumbnail" | Fix now → F-5 | verified: the button is swapped for `PVBadge`; `focusedID` is only bound to the open button |
| MNT-4 | MNT | minor | Third copy of the test `CatalogPicture` builder | Fix now → F-6 | verified: `rg -n "CatalogPicture\(" macos/ProvenenciaTests` gives three sites. Same-pass cleanup; supersedes §8's "no shared test file" |
| CON-2 | CON | minor | The wider DoD sweep still hits landed Spike 13 and S14-01 plans, and the Landed note does not say so | Fix now → F-7 | verified: those hits describe history |
| PRF-1 | PRF | minor | A thumbnail choice invalidates the whole Events, Places and Persons lists | Follow-up | existing mechanism in `CatalogQueryRegistry.invalidations` that every handle write shares; this PR adds a caller but does not change the mechanism or the per-write cost. Home: an idea file for header patches on list keys |

Nits (logged): `UX-3 · L10n.swift thumbnailMenuHint · "Right-click for thumbnail options" tells VoiceOver users to right-click; copies portraitMenuHint` · `UX-4 · PropertiesModel.swift:171 · row label speaks the raw key (plan's choice)` · `MNT-5 · docs/macos-client-patterns.md:102 · clause repeats itself` · `MNT-6 · CitationComposerSubjectGroups.swift:248 · re-wrap doc line`

### Fix plan 1

**F-1 — VoiceOver hears the thumbnail and portrait result** (from UX-1)
Files: `macos/App/Features/Conclusions/ConclusionLeadModel.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift` (or the lead-model test file in use).
Do: in `ConclusionLeadModel.write`, after each `toast =`, call `AccessibilityAnnouncement.post` (`macos/App/Platform/AccessibilityAnnouncement.swift`) with the toast's title, plus `", "` and body when the body is not empty. Use `.high` priority for the danger toast and the default for success. Reuse the existing helper; add no new announcement mechanism and do not change `VocabularyToastOverlay` or `PVToast`.
Check: build passes; scripts/check-pr.sh PASS.

**F-2 — One builder for the Event and Place picture slot** (from MNT-1)
Files: `ConclusionDetailPage.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`, the Event and Place tests if they need it.
Do: add `static func pictures(from detail: CatalogConclusionDetail, fallbackMark: PVMarkKey, choose: LeadCopy) -> ConclusionDetailGallery` on `ConclusionDetailGallery`, beside `tiles(from:…)`. It fills `picturesSection`, `pictureOpenTip` and the three tile functions (`pictureTileLabel`, `useAsThumbnailLabel`, `toastThumbnailChosenBody`). The Event and Place inits each call it once, then `.leading(…)` with their own `noImage` / `withImage` and `pictureFrom`. The Person content keeps its own call. Keep the statics `.eventThumbnail` and `.placeThumbnail` by name (D-4), and build one from the other with only `menuTitle` changed, so the shared six fields are written once.
Check: `EventDetailTests`, `PlaceDetailTests` pass unchanged in what they assert; `rg -n "pictureTileLabel|useAsThumbnailLabel|toastThumbnailChosenBody" macos/App/Features` prints only the new builder; scripts/check-pr.sh PASS.

**F-3 — One function for the plugin origin's words** (from MNT-2)
Files: `macos/App/Features/Catalog/OriginMarkers.swift`, `macos/App/Features/Properties/PropertiesSnapshot.swift`.
Do: add `static func pluginLabel(for origin: String) -> String` on `OriginPill`, returning `L10n.Origin.pluginPill(pluginID: CatalogOrigin.pluginID(from: origin))`. `OriginPill.body` and `PropertiesOriginCell.spokenOrigin` both call it. Visible text does not change.
Check: `rg -n "pluginPill\(pluginID" macos/App` prints only `pluginLabel` and the L10n definition; `rowAccessibilitySpeaksAPluginOriginAsThePillDoes` passes; scripts/check-pr.sh PASS.

**F-4 — Put `canEditCardinality`'s doc back and move the row label beside `deleteAccessibilityLabel`** (from CON-1, MNT-3)
Files: `macos/App/Features/Properties/PropertiesModel.swift`.
Do: move `rowAccessibilityLabel(for:)` with its own doc comment to sit directly after `deleteAccessibilityLabel`, as D-11 says. Leave the line "Seeded and plugin Properties show Holds read-only…" on `canEditCardinality`.
Check: scripts/check-pr.sh PASS.

**F-5 — Keep keyboard focus on the tile after "Use as thumbnail"** (from UX-2)
Files: `ConclusionLeadModel.swift`, `ConclusionDetailPage.swift` (the gallery section).
Do: make `choose` return whether the save worked (`@discardableResult … -> Bool`). In the tile's "Use as" button action, when it returns true, set the section's existing `focusedID = tile.id`, so focus moves to that tile's open-Citation button.
Check: scripts/check-pr.sh PASS. Manual: with Full Keyboard Access, focus stays on the tile after choosing.

**F-6 — One test builder for `CatalogPicture`** (from MNT-4)
Files: `macos/ProvenenciaTests/ReconciledValueDisplayTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PersonDetailTests.swift`.
Do: add `static func picture(_ id: String, path: String, title: String, leads: Bool = false) -> CatalogPicture` beside `outcome` in `ReconciledValueDisplayTests` (the existing shared fixture; no new file). Point all three tests at it and delete their own copies. This supersedes the plan's §8 note "no shared test file".
Check: `rg -n "CatalogPicture\(" macos/ProvenenciaTests` prints only the new builder; scripts/check-pr.sh PASS.

**F-7 — Name the historical sweep hits, and land the supersessions, in the Landed note** (from CON-2)
Files: `docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md`.
Do: append to the **Landed** line: the wider sweep still finds the old names in the landed S13-09, S13-10 and S14-01 plans, the Spike 13 map, and `deployment-plan.review.md`, which describe history and were left alone; F-2's shared builder and F-6's shared test builder supersede the plan's per-content blocks and its §8 "no shared test file" note.
Check: `python3 scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.
