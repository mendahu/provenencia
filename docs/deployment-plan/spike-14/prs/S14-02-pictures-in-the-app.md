# S14-02 — Pictures in the app

**Status:** Blocked — on board S14-D2 and on S13-10. Refreshed 2026-10-30 after S13-23 landed ([#633](https://github.com/mendahu/provenencia/pull/633)): the Promote wording, the relationship `kind` wire field and their steps left this PR. The PR waits on a board because it changes how the Event and Place pages and lists look (user, 2026-10-30), and on S13-10 because it is written against the page S13-10 leaves (Page columns, `PersonDetailSections`).
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D2](../design/S14-D2-event-place-pictures.md), board not yet agreed
**Board:** — (when agreed: the Claude Design link and frame numbers go here, and **Design → macOS** describes each frame in words, as S13-10 does for S13-D1)
**Depends on:** S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615)), S13-23 (landed [#633](https://github.com/mendahu/provenencia/pull/633)), S13-10 (Spike 13, Planned), S14-D2
**Blocks:** S13-14 (Spike 13's Gallery section reads the `gallery` slot)
**Branch:** `s14-02-pictures-in-the-app`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, as S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [event-place-pictures.md § What already works](../event-place-pictures.md#what-already-works), [§ Thumbnail parity](../event-place-pictures.md#thumbnail-parity), [§ What changes](../event-place-pictures.md#what-changes); working rules: "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "A depiction is worded as a picture outside the Person's own copy", "The gallery layer is named gallery, not likeness", "Batch 1 is cut by layer"; S14-01 ship follow-up UX-1 (Properties VoiceOver); Spike 13's [S13-10 plan](../../spike-13/prs/S13-10-header-and-sections.md) D-3, D-4, D-5, D-10, D-16
**Planned:** 2026-10-30 against `main` @ `1253dfaa` (S13-10 not yet landed: §4 rows marked *S13-10* describe the code its plan leaves; redo them on its Landed note)
**Review:** [`S14-02-pictures-in-the-app.review.md`](S14-02-pictures-in-the-app.review.md) — refresh round 3: Fits with changes / Not ready (board) → all findings answered

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Event and Place pages show a Pictures section, lead their headers with the picture the engine already sends, and let the researcher choose a thumbnail the way a Person chooses a portrait. Events and Places lists show that lead. The Swift view layer names the gallery gallery/picture and the kind-neutral choose machine "lead"; the Person page's likeness and portrait copy stays. The Properties list tells two Properties labelled "Picture" apart for VoiceOver.

**Why now.** S14-01 landed the seeds, the gallery keys, the header leads and the wire rename; bar outcomes 3, 5 and 7 need the app. Spike 13's Gallery section (S13-14) builds on the same slot.

**What stays.** RPC `SetPreferredLikeness` and column `preferred_likeness_observation_id`; Property key `likeness`, the Person section title and portrait strings and their L10n functions; gallery SQL and seeds (S14-01); Spike 13's redesigned Gallery section (S13-14); rectangle-only drawing (S14-03). No wire, Go or schema change.

## 2. User stories

**US-1.** As a researcher, when an Event or Place has pictures, I see a Pictures section on its page with crops that open their Citations, so I can browse them as I browse likenesses on a Person. (S14-D2 EPP-1 – EPP-3)

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Place detail whose `CatalogConclusionDetail.pictures` has one tile (Source title "Church photo") | `PlaceDetailContent.gallery.tiles` has one tile; `gallery.title` is `L10n.Conclusions.picturesSection` (key `conclusions.detail.pictures`, "Pictures"); tile `sourceTitle == "Church photo"`; tile `accessibilityLabel == "Picture from Church photo. Open citation"` |
| AC-1.2 | Event detail with two pictures | Same title; tile ids are the two observation ids in wire order |
| AC-1.3 | Event or Place detail with no pictures | `gallery.tiles` is empty; the default fill draws no gallery section (EPP-2) |
| AC-1.4 | Person detail with likenesses | `gallery.title` is `L10n.Conclusions.likenessesSection`; tile `accessibilityLabel == L10n.Conclusions.likenessTileLabel(source)`; `gallery.choose == .personPortrait` |

**US-2.** As a researcher, when an Event or Place has pictures, the page header and the list row lead with that picture, and I can choose another tile as the thumbnail or go back to the earliest. (EPP-4 – EPP-8)

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Event header `pictureImageRelPath = "objects/ab/cd.png"`; its tile A is first and `isLead` | `thumbnail.imageRelPath == "objects/ab/cd.png"`; `thumbnail.label == L10n.Conclusions.eventPictureThumb`; `thumbnail.imageLabel == "Picture from <A's Source>"`; `gallery.offersEarliest == false` |
| AC-2.2 | Same Event; tile B (second) leads, A not | `gallery.offersEarliest == true`; `gallery.choose == .eventThumbnail`; B's `useAsLabel` is "Use the picture from <B's Source> as thumbnail" and A's likewise |
| AC-2.3 | Two tiles share one `imageRelPath` (identical crops; objects are content-addressed), the second leads | `thumbnail.imageLabel` names the second tile's Source (D-5) |
| AC-2.4 | Header path set, no tile leads (header and gallery read at different times) | `thumbnail.imageLabel == L10n.string(eventPictureThumb)` (D-5 fallback, as the Person today) |
| AC-2.5 | `ConclusionLeadModel` on a `FakeStore`: `choose(B, copy: .eventThumbnail)` then `useEarliest(copy: .eventThumbnail)` | `store` records `setPreferredLikeness` with the entity id and B's observation id, then `""`; toast titles "Thumbnail updated"; bodies "Using the picture from <B's Source> as the thumbnail." and "Using the earliest picture as the thumbnail." |
| AC-2.6 | Same calls with `.personPortrait` and a Person tile | Toast English unchanged ("Portrait updated", today's bodies) |
| AC-2.7 | `EventsList.thumbnailRelPath(header)` / `PlacesList.thumbnailRelPath(header)` | Equals `header.pictureImageRelPath`; `""` when the header has none |
| AC-2.8 | Place page with no picture | `thumbnail.imageRelPath == ""`, `label == L10n.Conclusions.placeNoImage`, `imageLabel == nil` |

**US-3.** As a researcher using VoiceOver on the Properties list, I can tell `event_picture` from `place_picture` even though both labels are "Picture".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Property `label: "Picture", key: "event_picture"`, origin `provenencia`, bound only to the Event type | `PropertiesModel.rowAccessibilityLabel(for:)` == "Picture, event_picture, <value type>, seeded, bound to 1 type, Event" (value type and bound text from today's L10n) |
| AC-3.2 | Same with `place_picture` bound to Place | Contains "place_picture" and "Place"; differs from AC-3.1 |
| AC-3.3 | A user Property bound to no type | Ends with the bound text; no trailing ", " (D-11) |

**US-4.** As a contributor, the Swift view layer says gallery, picture and lead, and the composer's depiction helper says Subject.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | After the PR, `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop|ConclusionPortraitModel|ValueFamily|depictionPerson|var likenesses|offersEarliest: ' macos/App` | Exit 1 (no match) |
| AC-4.2 | `CachedThumbnail.pictureCrop` | Equals `96` |
| AC-4.3 | `CitationComposerSubjectGroups` | Has `depictionSubject`; its comments say Subject |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionDetailGallery` *(new)* + `ConclusionDetailPicture` *(new)* on `ConclusionDetailBody`; remove `likenesses` | STEP-1, AC-1.*, D-1, D-2 |
| Rename `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection`, `LikenessTileView` → `GalleryTileView`, `CachedThumbnail.likenessCrop` → `pictureCrop`; ids per D-3 | STEP-1, AC-4.1, AC-4.2, D-3 |
| `ConclusionDetailThumbnail.leading(…)` *(new)*; `offersEarliest` moves from the thumbnail to the gallery; Person, Event and Place contents fill `gallery` and their thumbnails (EPP-1, EPP-4) | STEP-1, AC-1.*, AC-2.1 – AC-2.4, AC-2.8, D-4, D-5 |
| `ConclusionPortraitModel` → `ConclusionLeadModel`, built by the page from `userID`; copy read at call time from `gallery.choose`; Event and Place views pass `userID`; host passes it (EPP-5 – EPP-7) | STEP-1, AC-2.2, AC-2.5, AC-2.6, D-6, D-7 |
| `EventsList` / `PlacesList` `thumbnailRelPath`; delete the protocol default (EPP-8) | STEP-2, AC-2.7, D-8 |
| Delete `L10n.Promote.ValueFamily`; composer `depictionSubject` | STEP-3, AC-4.1, AC-4.3, D-9, D-10 |
| Properties row VoiceOver label (S14-01 UX-1) | STEP-4, AC-3.*, D-11 |
| L10n: Event/Place picture and thumbnail functions and resources (§ Strings), Properties row label (EPP-9) | STEP-1, STEP-4, D-12 |
| The S14-D2 board's answers, at Regular and Wide | STEP-1, STEP-2, §6 macOS, §10 |
| Swift tests | STEP-5, T-* |
| Docs this PR makes false (D-13), map, `completed.md`, brief archive | STEP-6 |

| Out | Home |
| --- | --- |
| Spike 13's redesigned Gallery section and viewer | S13-14, S13-15 |
| Sections or a switcher on the Event or Place page | not planned (S13-D1 R-1) |
| Rectangle-only drawing; composer draw feedback | S14-03 |
| Renaming `preferred_likeness_observation_id` or `SetPreferredLikeness` | not planned (working rule "The gallery layer is named gallery") |
| Changing the Person section title, portrait copy or their L10n functions | not planned (working rule: likeness stays on the Person) |
| Promote picture wording, relationship `kind` on the wire | Descoped 2026-10-29 — the evidence sheet is gone (S13-23); S13-24 removes the claim-evidence wire |
| Changing `ReconciledValueDisplay.string(for: .depiction)` / `L10n.Conclusions.likenessValue` | not planned — Event and Place pages show no depiction row (their rows are fixed) |
| Composer catalog comments on `depictionThumbLabel` / `depictionLineLabel` | not planned — their values already take any Subject label |
| Moving `vocabularyToastOverlay` into the kit under a neutral name | logged (review M-11); not planned |
| Person Timeline event picture | Spike 13 R-26 |
| Pictures as matching weight for Event or Place | not planned ([event-place-pictures.md](../event-place-pictures.md) Stays later) |
| A new kit primitive under `DesignSystem/` | not planned (S14-D2 inventory) |

## 4. Current state

Rows marked *S13-10* are the code its plan leaves (S13-10 plan §6 and D-3 – D-5, D-16). Re-read them on S13-10's Landed note before STEP-1.

| Area | Where | Today |
| --- | --- | --- |
| Body protocol | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailBody` | `var likenesses: [ConclusionDetailLikeness]`, extension default `[]` |
| Default fill | same file `ConclusionDetailDefaultPageSections` (*S13-10*: `ConclusionDetailRowsSection`, the likeness `if`, the relationship `ForEach`) | Draws `ConclusionDetailLikenessSection` when `likenesses` is non-empty, between rows and relationship sections |
| Tile type | same file `ConclusionDetailLikeness` | `id`, `imageRelPath`, `sourceTitle`, `location`, `isPortrait` |
| Gallery section | same file `ConclusionDetailLikenessSection` (private; *S13-10*: internal) | Init `(projectDir:likenesses:portrait:)`. Hard-codes `likenessesSection`, `likenessOpenTip`, `likenessTileLabel`, ids `persons.detail.likeness.\(tile.id).portraitBadge` / `.usePortrait`, `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel` |
| Tile view | same file `LikenessTileView` (private) | `CachedThumbnail.likenessCrop` (96), fallback mark `.subjectPerson` |
| Thumbnail slot | same file `ConclusionDetailThumbnail` | `mark`, `label`, `imageRelPath`, `imageLabel`, `colors`, `offersEarliest` (read only by the header's `showsPortraitMenu`) |
| Header | same file `ConclusionDetailHeader` | `portrait: ConclusionPortraitModel?`; menu iff `portrait != nil && thumbnail.offersEarliest`; `portraitMenuTitle`, `portraitMenuHint`, `useEarliestLikeness`; ids `persons.detail.portrait`, `.menu`, `.useEarliest` |
| Page | same file `ConclusionDetailPage` | Synthesized memberwise init `(session, entityId, pageIdentifier, errorIdentifier, make, summary, portrait = nil, pageSections)` and a convenience init for the default fill. *S13-10* keeps that order (D-16), draws `PVPageColumns` in `ConclusionDetailLoaded`, and deletes `page(identifier:body:)` |
| Context | same file `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing`, `portrait: ConclusionPortraitModel?` (*S13-10* adds `section`) |
| Person content | `PersonDetailContent.swift` | `likenessImageRelPath`, `likenesses`, `thumbnail` (labels `personNoLikeness` / `personLikenessThumb`, `colors`), `leadLabel` matched by path with fallback `personLikenessThumb` |
| Event / Place content | `EventDetailContent.swift`, `PlaceDetailContent.swift` | `thumbnail` = mark + `eventNoImage` / `placeNoImage`; ignore `header.pictureImageRelPath` and `detail.pictures`. Place rows Names and Period, then relationship `sections` |
| Portrait model | `ConclusionPortraitModel.swift` (pbxproj lines 99, 543, 1324, 2393) | `init(session:userID:entityId:)`; `choose(_: ConclusionDetailLikeness)`, `useEarliest()`; Person toast strings |
| Person view | `PersonDetailView.swift` | Owns `@State portrait` and `.vocabularyToastOverlay($portrait.toast, identifier: "persons.detail.toast")`. *S13-10*: passes `portrait` through the full init with `pageSections: PersonDetailSections(…)`; `.gallery` case calls `ConclusionDetailLikenessSection(projectDir:likenesses:portrait:)` or `PVEmptyState` with `galleryEmptyTitle` / `galleryEmptyMessage` |
| Event / Place views | `EventDetailView.swift`, `PlaceDetailView.swift` | Convenience init; no `userID` |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` | Holds `userID` and passes it to several destinations, including `PersonDetailView` (line 101); not to `EventDetailView` (107) or `PlaceDetailView` (113). Pages are keyed by `viewIdentity` = entity id (`PlaceRegistry`) |
| List thumbs | `ConclusionListPage.swift` | `ConclusionListKind.thumbnailRelPath` with extension default `""` (line 72); only `PersonsList` overrides it; `PVList` thumbnail at its default 44pt |
| Leftover from S13-23 | `macos/App/Platform/L10n.swift` line 1349 | `private enum ValueFamily`, no reader |
| Composer | `CitationComposerSubjectGroups.swift` line 239 | `depictionPerson`, comments say person |
| Properties VO | `PropertiesView.swift` `rowAccessibilityLabel` (private, line 416) | `"\(label), \(valueType), \(origin), \(boundText)"` with the raw origin string; `PropertiesSnapshot.boundTypes(for:)`; `PropertiesOriginCell` maps `CatalogOrigin.user` → `originUserShort`, `.provenencia` → `originSeededShort`; `PropertiesModel.deleteAccessibilityLabel` is the feature's other spoken label |
| Stale text | `CachedThumbnail.swift` line 15 ("Promote evidence sheet"); `ConclusionListRow.swift` line 22 ("the likeness a Person leads with"); `docs/macos-client-patterns.md` line 102 ("likeness gallery"); `event-place-pictures.md` lines 32 (`likenesses` default), 33 and 64 (evidence sheet) | |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` (uses `ConclusionDetailLikeness`, `likenesses`, `isPortrait`, the portrait model), `PersonsListTests.swift` (line 21: `EventsList.thumbnailRelPath(event) == ""`), `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PropertiesModelTests.swift` (`makeModel` helper) | The test target has no synchronized groups; this PR adds no test file |

**Precedent.** Person gallery and portrait: `d31e1856` ("Let a researcher choose a portrait from the gallery and go back to the earliest from the header thumb.") and `e402e1fe`. List thumb: `5476b3b0`. One shared body field: `ConclusionDetailThumbnail` (S10-23 / S10-35). Formatted Strings held by content: `PersonDetailContent.Vital.accessibilityLabel`, `members`. Wire rename: #615.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Replace `likenesses` with one required `gallery: ConclusionDetailGallery` on `ConclusionDetailBody` (no default). The default fill draws the gallery section iff `!gallery.tiles.isEmpty` | Working rule "one slot" | A default gallery; keeping `likenesses` |
| D-2 | `ConclusionDetailGallery` holds `title: LocalizedStringResource`, `openTip: LocalizedStringResource`, `tiles: [ConclusionDetailPicture]`, `choose: LeadCopy` (non-optional), and computes `lead` (`tiles.first(where: \.isLead)`) and `offersEarliest` (`tiles.first.map { !$0.isLead } ?? false`). `ConclusionDetailPicture` holds `id`, `imageRelPath`, `sourceTitle`, `location`, `isLead`, and ready Strings `accessibilityLabel`, `useAsLabel`, `chosenToastBody`. `static func tiles(from: CatalogConclusionDetail, tileLabel: (String) -> String, useAsLabel: (String) -> String, chosenBody: (String) -> String)` builds them, passing each picture's Source title to the three functions | Copy with arguments is formatted by L10n functions, once, in the content (`macos-l10n.mdc`; review M-1); one source per fact | Format templates stored in the slot and filled by views |
| D-3 | Rename to `ConclusionDetailGallerySection`, `GalleryTileView`, `pictureCrop`. Ids are built from the page's `pageIdentifier` with fixed leaves, for every kind: tile `\(page).picture.\(tile.id)`, badge `…\(tile.id).leadBadge`, button `…\(tile.id).useAsLead`; header thumb `\(page).lead`, menu `\(page).lead.menu`, item `\(page).lead.useEarliest`. Person ids change from `persons.detail.portrait*` / `.likeness.*.portraitBadge` / `.usePortrait` (no test or UI test reads them: `rg -n 'detail.portrait|portraitBadge"|usePortrait|detail\.likeness\.' macos` hits only `ConclusionDetailPage.swift`). The tile fallback mark is `content.thumbnail.mark` | One rule, no per-kind id fields (review M-3) | Id strings on the pack |
| D-4 | `LeadCopy` *(new, nested in `ConclusionDetailGallery`)* holds only static copy: `badge`, `useAs`, `useEarliest`, `menuTitle`, `menuHint` (`LocalizedStringResource`), `toastTitle`, `toastEarliestBody` (`String`, as today's Person toast members). Statics `.personPortrait` (today's Person members, table in § Strings), `.eventThumbnail`, `.placeThumbnail` (they differ only in `menuTitle`) | The section and header stay kind-blind; S13-14 is another pack | A kind enum in the section |
| D-5 | `static func leading(mark: PVMarkKey, noImage: LocalizedStringResource, withImage: LocalizedStringResource, imageRelPath: String, gallery: ConclusionDetailGallery, leadLabel: (String) -> String, colors: PVThumbnail.Colors? = nil) -> ConclusionDetailThumbnail` *(new)*: `label` = `imageRelPath.isEmpty ? noImage : withImage`; `imageLabel` = nil when the path is empty, else `gallery.lead.map { leadLabel($0.sourceTitle) } ?? L10n.string(withImage)`. Person: `.subjectPerson`, `personNoLikeness`, `personLikenessThumb`, `likenessLabel`, `treatment?.thumbnailColors`. Event: `.subjectEvent`, `eventNoImage`, `eventPictureThumb`, `pictureFrom`. Place: `.subjectPlace`, `placeNoImage`, `placePictureThumb`, `pictureFrom`. Delete `ConclusionDetailThumbnail.offersEarliest`; the header reads `content.gallery.offersEarliest` | One builder for three kinds (review M-4); lead by `isLead` is right when two tiles share a path; the fallback keeps today's Person label | Three copies of the logic; matching by path |
| D-6 | Rename `ConclusionPortraitModel` → `ConclusionLeadModel` and its file → `ConclusionLeadModel.swift` (edit the four pbxproj lines in §4 by hand: path, name and both comments). `init(session:userID:entityId:)` unchanged; `choose(_ picture: ConclusionDetailPicture, copy: LeadCopy)` toasts `copy.toastTitle` / `picture.chosenToastBody`; `useEarliest(copy: LeadCopy)` toasts `copy.toastTitle` / `copy.toastEarliestBody`. The `isSaving` guard and the danger toast on failure stay | Kind-neutral name beside `isLead` (review M-9); copy has one source, read at call time (review M-2) | Copy passed at init |
| D-7 | **The page owns the model.** `ConclusionDetailPage` replaces `portrait` with `userID: String` and gains `@State private var lead: ConclusionLeadModel`. Both inits become hand-written: full `init(session:entityId:pageIdentifier:errorIdentifier:userID:make:summary:pageSections:)` and convenience `init(session:entityId:pageIdentifier:errorIdentifier:userID:make:summary:)`; each sets `_lead = State(initialValue: ConclusionLeadModel(session: session, userID: userID, entityId: entityId))`. The model is always built: every kind can choose. The outermost view adds `.vocabularyToastOverlay($lead.toast, identifier: "\(pageIdentifier).toast")`. `ConclusionDetailPageSectionsContext.portrait` becomes `lead: ConclusionLeadModel` and gains `pageIdentifier: String`; `ConclusionDetailLoaded` and `ConclusionDetailHeader` take `lead` and `pageIdentifier` in place of `portrait`. `PersonDetailView` drops its `@State` and overlay and passes `userID` (its toast id stays `persons.detail.toast`); `EventDetailView` / `PlaceDetailView` take `userID` and pass it; the host passes `userID` to both | One machine, one owner, no Optional nobody sets to nil (review M-2); the page is keyed by entity id, so the model resets per entity | A `leadChoose` page parameter beside `gallery.choose`; copying the Person's `@State` into each view |
| D-8 | `EventsList` / `PlacesList` implement `thumbnailRelPath` as `header.pictureImageRelPath`; delete the extension default in `ConclusionListPage.swift` so a fourth kind must answer | No dead default (review M-7) | Keeping the default |
| D-9 | Delete `L10n.Promote.ValueFamily` | S13-23 removed its last reader; its `likeness` case is in this PR's rename sweep | Leaving it |
| D-10 | Rename `depictionPerson` → `depictionSubject`; its comments and its two call sites' comments say Subject | Map scope | Retitling catalog keys |
| D-11 | `PropertiesModel.rowAccessibilityLabel(for property: CatalogProperty) -> String` *(new)* beside `deleteAccessibilityLabel`. It reads `snapshot.boundTypes(for:)`, speaks the origin through `PropertiesOriginCell.spokenOrigin(_:) -> String` *(new static; `user` → `originUserShort`, `provenencia` → `originSeededShort`, else the raw string, as the cell's `OriginPill` branch shows it)* and formats with `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:)` *(new, "%1$@, %2$@, %3$@, %4$@, %5$@")*, where `bound` is today's `rowBoundAnnouncement(count:)`; when bound types exist it appends `", "` + their labels joined with `", "`. `PropertiesView` calls it; its private function goes | Testable through `makeModel` in `PropertiesModelTests`; origin spoken as the cell shows it (review M-10); no stray comma (X-10) | A raw origin string in a new key; a six-slot format with an empty last slot |
| D-12 | New L10n per § Strings, each a `static let` resource or a `static func … -> String` through `L10n.format`, as `add-localized-string` step 4 says. Person functions are reused as they are | `macos-l10n.mdc` | Reusing "Use as portrait" on an Event |
| D-13 | **Against S13-10's page.** STEP-1 starts from the page S13-10 leaves. `PersonDetailSections`' `.gallery` case becomes `ConclusionDetailGallerySection(projectDir: context.projectDir, pageIdentifier: context.pageIdentifier, gallery: content.gallery, fallbackMark: content.thumbnail.mark, lead: context.lead)` when `!content.gallery.tiles.isEmpty`, else its `PVEmptyState` reading `galleryEmptyTitle` / `galleryEmptyMessage` as S13-10 wrote it. No empty-state field on the slot | S14-D2 needs Page columns, which S13-10 builds; one design, not two (review M-5, M-8) | Branching on landing order |
| D-14 | **The board binds the chrome.** Spacing, badge and button placement and the Wide arrangement follow the agreed S14-D2 frames, written into §6 macOS when the board is agreed | Repo rule: each view change has a board (user, 2026-10-30) | Building from the Person page alone |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionDetailGallery` *(new)* | type | `ConclusionDetailThumbnail` | Same slot pattern for the section |
| `ConclusionDetailPicture` *(new)* | type | `ConclusionDetailLikeness` | Replaces it (Deletes) |
| `ConclusionDetailGallery.LeadCopy` *(new)* + 3 statics | type | Person L10n hard-coded in the section and header | The section stays kind-blind |
| `ConclusionDetailGallery.tiles(from:tileLabel:useAsLabel:chosenBody:)` *(new)* | func | `PersonDetailContent` init map | Lifted to one place |
| `ConclusionDetailThumbnail.leading(…)` *(new)* | func | `PersonDetailContent.thumbnail` / `leadLabel` | Lifted and fixed (D-5) |
| `ConclusionLeadModel` *(renamed)* | type | `ConclusionPortraitModel` | Rename (D-6) |
| `GalleryTileView` *(renamed)* | view | `LikenessTileView` | Rename |
| Event/Place picture and thumbnail L10n *(new)* | keys | Person likeness / portrait L10n | Parallel wording |
| `PropertiesModel.rowAccessibilityLabel(for:)`, `PropertiesOriginCell.spokenOrigin(_:)`, `L10n.Properties.rowAccessibility` *(new)* | func, key | `PropertiesView.rowAccessibilityLabel`, the cell's branches | Moved, localized, shared |

**Builds on.** The `ConclusionDetailThumbnail` slot; `setPreferredLikeness`; `ConclusionListKind.thumbnailRelPath`; S13-10's `PersonDetailSections` and context; `PropertiesSnapshot.boundTypes`.
**Deletes.** `ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`, `likenesses` and its default; `ConclusionDetailThumbnail.offersEarliest`; `PersonDetailContent.likenessImageRelPath` and `leadLabel`; `CachedThumbnail.likenessCrop`; `ConclusionPortraitModel` (renamed); `PersonDetailView`'s `@State` and overlay; the page's and context's `portrait`; the `thumbnailRelPath` default; `PropertiesView.rowAccessibilityLabel`; `L10n.Promote.ValueFamily`; `depictionPerson`; ids `persons.detail.portrait*` and `.likeness.` segments.
**Next adds.** S13-14 reads the same slot and model. A fourth kind adds a `LeadCopy` static, a `leading(…)` call and a list `thumbnailRelPath` (the compiler asks).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Event/Place page with tiles | `ConclusionDetailGallerySection` | No section when empty; a tile opens `tile.location` as on the Person; a missing image shows `content.thumbnail.mark` |
| Event/Place page with a lead model | Header menu, tile badge and button | The page is keyed by `viewIdentity` (entity id), so the model is rebuilt per entity; no menu when `offersEarliest` is false; a tap while saving is ignored (the `isSaving` guard, as on the Person); a failed write shows the danger toast; a later success replaces it |
| Events/Places list row with a path | `PVList` thumbnail slot | Empty path → the kind's mark |

### Scale

At the map's sizes (one researcher's catalog): the gallery is a `LazyVGrid` over one entity's tiles, as for the Person; three Strings per tile are formatted once per content build; the list thumb is one path per row already on the header; the Properties label is O(bound types) per row. No new write.

### Edges

- **Assumptions.** The header path and `leads` come from one lead rule (S14-01 `gallery.Lead`); D-5 no longer assumes one tile per path (T-3) and keeps a label when the two reads disagree (T-3).
- **Neighbors.** A Person keeps every string; its ids change per D-3 (T-1, T-5). An Event with one picture leads with it, so no menu (T-2). An empty Source title gives "Picture from " as the Person's likeness label does today; unchanged.
- **Removed inputs.** A deleted picture drops out of `detail.pictures`; if it led, the engine's next lead fills the header and `offersEarliest` follows the first tile (T-2).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — tiles come from the app's own catalog through the existing read.

### Schema

None — no catalog change.

### Go core

None.

### FFI

None — S14-01 put pictures and header paths on the wire.

### macOS

Draft from the S14-D2 brief; when the board is agreed, each frame is described here in words and the rows take its answers (D-14).

| Building block | Component | Home | Requirement |
| --- | --- | --- | --- |
| Gallery section | `PVSectionHeader` (title, count as meta) + `LazyVGrid` (adaptive 120pt columns, as today) + `GalleryTileView` | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | EPP-1 – EPP-3; below the rows, above relationship sections |
| Tile crop | `CachedThumbnail` | `GalleryTileView` | `pictureCrop`; fallback `content.thumbnail.mark` |
| Badge / Use as | `PVBadge` accent + photo glyph, `PVButton` ghost small | gallery section | EPP-5; copy from `gallery.choose`, labels from the tile |
| Earliest menu | `pvRightClickMenu` / `PVActionMenuItem` | `ConclusionDetailHeader` | EPP-6 |
| Toast | `.vocabularyToastOverlay` (draws `PVToast`) | `ConclusionDetailPage` | EPP-7 |
| Header thumb | `CachedThumbnail` 120pt | `ConclusionDetailHeader` | EPP-4 |
| List thumb | `PVList` thumbnail slot, default 44pt | `ConclusionListPage` | EPP-8, EPP-W2 |
| Wide | `PVPageColumns`, primary-only (S13-10) | `ConclusionDetailLoaded` | EPP-W1: no code change; the grid fills the primary column (up to 1180pt, about one tile more per row than the 1038pt Regular column) |

Sketch:

```swift
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var tiles: [ConclusionDetailPicture]
    var choose: LeadCopy
    var lead: ConclusionDetailPicture? { tiles.first(where: \.isLead) }
    var offersEarliest: Bool { tiles.first.map { !$0.isLead } ?? false }

    struct LeadCopy: Equatable {
        var badge, useAs, useEarliest, menuTitle, menuHint: LocalizedStringResource
        var toastTitle, toastEarliestBody: String
        static let personPortrait = LeadCopy(/* § Strings */)
        static let eventThumbnail = LeadCopy(/* … menuTitle: eventThumbnailMenuTitle */)
        static let placeThumbnail = LeadCopy(/* … menuTitle: placeThumbnailMenuTitle */)
    }

    static func tiles(
        from detail: CatalogConclusionDetail,
        tileLabel: (String) -> String,
        useAsLabel: (String) -> String,
        chosenBody: (String) -> String
    ) -> [ConclusionDetailPicture]
}

struct ConclusionDetailPicture: Equatable, Identifiable {
    var id: String
    var imageRelPath: String
    var sourceTitle: String
    var location: WorkspaceLocation
    var isLead: Bool
    var accessibilityLabel: String
    var useAsLabel: String
    var chosenToastBody: String
}
```

### Strings

New (Event and Place; `L10n.Conclusions`):

| Member | Key | Default |
| --- | --- | --- |
| `picturesSection` | `conclusions.detail.pictures` | Pictures |
| `pictureOpenTip` | `conclusions.picture.open` | Open citation |
| `pictureTileLabel(_:)` | `conclusions.picture.tile` | Picture from %@. Open citation |
| `pictureFrom(_:)` | `conclusions.picture.from` | Picture from %@ |
| `eventPictureThumb` | `conclusions.event.pictureThumb` | Event picture |
| `placePictureThumb` | `conclusions.place.pictureThumb` | Place picture |
| `thumbnailBadge` | `conclusions.thumbnail.badge` | Thumbnail |
| `useAsThumbnail` | `conclusions.thumbnail.useAs` | Use as thumbnail |
| `useAsThumbnailLabel(_:)` | `conclusions.thumbnail.useAsLabel` | Use the picture from %@ as thumbnail |
| `useEarliestPicture` | `conclusions.thumbnail.useEarliest` | Use earliest picture |
| `eventThumbnailMenuTitle` / `placeThumbnailMenuTitle` | `conclusions.event.thumbnailMenuTitle` / `conclusions.place.thumbnailMenuTitle` | Event thumbnail / Place thumbnail |
| `thumbnailMenuHint` | `conclusions.thumbnail.menuHint` | Right-click for thumbnail options |
| `toastThumbnailUpdatedTitle` | `conclusions.thumbnail.toastTitle` | Thumbnail updated |
| `toastThumbnailChosenBody(_:)` | `conclusions.thumbnail.toastChosenBody` | Using the picture from %@ as the thumbnail. |
| `toastThumbnailEarliestBody` | `conclusions.thumbnail.toastEarliestBody` | Using the earliest picture as the thumbnail. |
| `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:)` | `properties.table.rowAccessibility` | %1$@, %2$@, %3$@, %4$@, %5$@ |

`L10n.Sources.useAsThumbnail` ("Set cover", key `sources.page.useAsThumbnail`) is a different member in a different namespace; no clash.

Person (reused as they are): title `likenessesSection`, tip `likenessOpenTip`, tile `likenessTileLabel(_:)`, header `personNoLikeness` / `personLikenessThumb` / `likenessLabel(_:)`, `.personPortrait` = `portraitBadge`, `useAsPortrait`, `useEarliestLikeness`, `portraitMenuTitle`, `portraitMenuHint`, `toastPortraitUpdatedTitle`, `toastPortraitEarliestBody`; tile functions `useAsPortraitLabel(_:)`, `toastPortraitChosenBody(_:)`.

Copy the board's wording where it differs, and record it in the Landed note. **Sweep:** `rg -n 'detail\.likeness\.|detail\.portrait' macos` finds nothing after the PR.

## 7. Steps

**STEP-1 — Gallery slot, lead model and contents (one buildable unit, one commit)**
Files: `ConclusionDetailPage.swift`, `CachedThumbnail.swift`, `ConclusionPortraitModel.swift` → `ConclusionLeadModel.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`, `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`, `PersonDetailView.swift` (and its `PersonDetailSections`), `EventDetailView.swift`, `PlaceDetailView.swift`, `WorkspaceDestinationHost.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-ui-component` (compose only), `add-localized-string`.
Do: D-1 – D-7, D-12, D-13. The app target does not build until all of these are done.
Check: the app target builds (`xcodebuild -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' build`); `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop|ConclusionPortraitModel' macos/App` finds nothing; `rg -n 'ConclusionLeadModel\(' macos/App` hits only `ConclusionDetailPage.swift`; `python3 scripts/check-localizable-xcstrings.py` passes. Tests compile again in STEP-5.

**STEP-2 — List thumbnails**
Files: `EventsListView.swift`, `PlacesListView.swift`, `ConclusionListPage.swift`. Skill: —.
Do: D-8.
Check: the app target builds.

**STEP-3 — Dead enum and composer helper**
Files: `L10n.swift`, `CitationComposerSubjectGroups.swift`. Skill: —.
Do: D-9, D-10.
Check: `rg -n 'ValueFamily|depictionPerson' macos` finds nothing.

**STEP-4 — Properties VoiceOver**
Files: `PropertiesModel.swift`, `PropertiesSnapshot.swift` (`PropertiesOriginCell`), `PropertiesView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-11.
Check: the app target builds.

**STEP-5 — Tests**
Files: `PersonDetailTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PersonsListTests.swift`, `PropertiesModelTests.swift`. Skill: `add-swift-test`.
Do: T-1 – T-8. Update Person tests for `gallery`, `isLead`, `ConclusionLeadModel` and its `copy:` argument; change `PersonsListTests` line 21 to expect the header path.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-6 — Docs**
Files: `docs/macos-client-patterns.md`, `docs/deployment-plan/spike-14/event-place-pictures.md`, `macos/App/Platform/CachedThumbnail.swift` (comment), `macos/App/Features/Conclusions/ConclusionListRow.swift` (comment), `ConclusionLeadModel.swift` (doc comment), this plan, `../completed.md`, `../deployment-plan.md`, `../design/S14-D2-event-place-pictures.md` → `../design/archive/`, `../design/README.md`. Skill: —.
Do: D-13 docs: in `macos-client-patterns.md` line 102, say "gallery slot" and add one sentence: a Conclusion kind passes `userID`, and the page owns the lead model and its toast. In `event-place-pictures.md`, line 32 says `gallery` (one slot, three kinds); rewrite the line-33 "Promote." bullet (the evidence sheet is gone, S13-23); drop "the evidence sheet only lists them (S14-02)" at line 64. The two Swift comments drop "Promote evidence sheet" and say "the picture a kind leads with"; the model's doc comment says it chooses the lead of any kind. Then the DoD's docs lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` reports no errors; `rg -n 'evidence sheet' docs/deployment-plan/spike-14/event-place-pictures.md macos/App/Platform/CachedThumbnail.swift` finds nothing.

Commits: one per behavior; the subject is one imperative sentence ending in a period that says what changed for the researcher ("Show Event and Place pictures in a Pictures section."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonDetailTests` (update the likeness tests) | Person `gallery.tiles` ids, order, `isLead`; title `likenessesSection`; tile `accessibilityLabel` from `likenessTileLabel`; `choose == .personPortrait`; `thumbnail.label` and `imageLabel` as today | AC-1.4, D-1, D-2, D-5 |
| T-2 | Swift | `EventDetailTests` *(new cases)* | No pictures → empty tiles, mark thumbnail; one picture → path, `eventPictureThumb`, "Picture from …", `offersEarliest == false`; two with B leading → `offersEarliest`, `choose == .eventThumbnail`, `useAsLabel` English, title `picturesSection` | AC-1.2, AC-1.3, AC-2.1, AC-2.2, D-5 |
| T-3 | Swift | `EventDetailTests` *(new cases)* | Two tiles with one path, second leads → `imageLabel` names the second Source; header path with no lead tile → `imageLabel == L10n.string(eventPictureThumb)` | AC-2.3, AC-2.4, D-5 |
| T-4 | Swift | `PlaceDetailTests` *(new cases)* | One picture "Church photo" → tile label and title; no picture → `placeNoImage`, nil `imageLabel` | AC-1.1, AC-2.8 |
| T-5 | Swift | `PersonDetailTests` model tests (rename to the lead model) | `FakeStore`: `choose(B, copy: .eventThumbnail)`, `useEarliest(copy: .eventThumbnail)` → recorded calls and thumbnail toast English; same with `.personPortrait` → today's English | AC-2.5, AC-2.6, D-6 |
| T-6 | Swift | `PersonsListTests` (extend) | `EventsList` / `PlacesList.thumbnailRelPath` = header path; `""` when unset | AC-2.7, D-8 |
| T-7 | Swift | `PropertiesModelTests` *(new cases)* | `makeModel` with Event and Place types and two "Picture" properties, warmed → the AC-3.1 string, a different AC-3.2 string; an unbound user property → no trailing ", " | AC-3.1, AC-3.2, AC-3.3, D-11 |
| T-8 | Gate | STEP-1, STEP-3 checks, DoD `rg` | AC-4.1 command exits 1; `CachedThumbnail.pictureCrop == 96`; `depictionSubject` exists | AC-4.1, AC-4.2, AC-4.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests`, `PersonsListTests`, `PropertiesModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md line 102 and event-place-pictures.md lines 32, 33, 64 (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes the slice Check true for this PR. Each page at Regular and at Wide (pane 1600pt or more), light and dark, against the S14-D2 frames.

1. New project. Compose a Place picture on a Citation (Property Picture). Open the Place: the Pictures section shows the crop below Names and Period and above any relationship section; a click opens the Citation; the header shows the picture. (EPP-1, EPP-3, EPP-4)
2. Add a second Place picture. **Use as thumbnail** on it: the header and the Places list row show it, the Thumbnail Badge moves, the Toast reads "Thumbnail updated". Right-click the header thumb → **Use earliest picture** restores the first. (EPP-5 – EPP-8)
3. Repeat 1–2 on an Event; open an Event with no picture: no Pictures section, the mark in the header. (EPP-2)
4. Widen the window past 1600pt of pane: the page stays one centered column; the grid fills it. (EPP-W1, EPP-W2)
5. Open a Person with likenesses: Gallery shows "Likenesses", "Portrait" and "Use as portrait"; choosing shows "Portrait updated".
6. VoiceOver (or Accessibility Inspector) on Properties: the `event_picture` and `place_picture` rows read differently.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S13-10 lands differently from its plan | §4 rows marked *S13-10* are redone on its Landed note before STEP-1 |
| The board changes a requirement (a Pictures empty state, a list-thumb size) | D-14: the plan is refreshed from the agreed board before it is `Planned` |
| The pbxproj rename | Four lines named in §4; the build check in STEP-1 catches a missed one |

**Stop and report if:**

- S13-10 has not landed, or its `PersonDetailSections`, full init or context differ from §4's *S13-10* rows.
- `CatalogConclusionDetail.pictures`, `CatalogPicture.leads` or a header's `pictureImageRelPath` is not as S14-01 left it.
- `ConclusionDetailBody` already has a gallery field from another PR.
- A Step cannot pass without renaming `SetPreferredLikeness` or the preferred column.
- The agreed board asks for a kit primitive the inventory does not mark New.
- `scripts/check-pr.sh` fails on a `main` breakage this PR did not cause.

## 12. Open questions

None. (The plan stays `Blocked` until the board is agreed and S13-10 has landed; then §4 and §6 macOS are refreshed and the status becomes `Planned`.)

## 13. PR

**Title:** S14-02 — Pictures in the app
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md`.
