# S14-02 — Pictures in the app

**Status:** Planned
**Refreshed:** 2026-10-30 after S13-23 landed ([#633](https://github.com/mendahu/provenencia/pull/633)): the Promote wording, the relationship `kind` wire field and their steps left this PR. Again on 2026-10-30 after the four boards were agreed and S13-10 landed ([#634](https://github.com/mendahu/provenencia/pull/634)).
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D2](../design/S14-D2-event-page-pictures.md) (Event page), [S14-D3](../design/S14-D3-place-page-pictures.md) (Place page), [S14-D4](../design/S14-D4-events-list-thumbnail.md) (Events list), [S14-D5](../design/S14-D5-places-list-thumbnail.md) (Places list); boards agreed 2026-10-30
**Board:** Four Claude Design projects, agreed 2026-10-30: S14-D2 [*Event Detail*](https://claude.ai/design/p/b3fd0cd4-787b-474d-b114-ea52e26537b2?via=share) (turn 3, frames 3a – 3g); S14-D3 [*Place Detail*](https://claude.ai/design/p/f7e0a599-285f-4099-ab92-532c68fd939d?via=share) (turn 2, frames 2a – 2g); S14-D4 [*Events List*](https://claude.ai/design/p/2799d7a0-387e-4570-b138-fc387562ac91?via=share) (frames 1 – 2); S14-D5 [*Places List*](https://claude.ai/design/p/e8910bb1-f564-4d7a-b829-69de0e60af5a?via=share) (frames 1 – 2). Earlier turns on those projects (S9-D6 and S9-D7 on the pages; S9-D3, S9-D4, S10-D10 and S10-D11 on the lists) are today's views, not this PR. §6 **macOS** describes each frame in words; the implementer does not need the boards.
**Depends on:** S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615)), S13-23 (landed [#633](https://github.com/mendahu/provenencia/pull/633)), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S14-D2, S14-D3, S14-D4, S14-D5 (agreed)
**Blocks:** S13-14 (Spike 13's Gallery section reads the `gallery` slot)
**Branch:** `s14-02-pictures-in-the-app`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, as S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [event-place-pictures.md § What already works](../event-place-pictures.md#what-already-works), [§ Thumbnail parity](../event-place-pictures.md#thumbnail-parity), [§ What changes](../event-place-pictures.md#what-changes); working rules: "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "A depiction is worded as a picture outside the Person's own copy", "The gallery layer is named gallery, not likeness", "Batch 1 is cut by layer"; S14-01 ship follow-up UX-1 (Properties VoiceOver); Spike 13's [S13-10 plan](../../spike-13/prs/S13-10-header-and-sections.md) D-3, D-4, D-5, D-10, D-16
**Planned:** 2026-10-30 against `main` @ `705cd98c` (S13-10 landed)
**Review:** [`S14-02-pictures-in-the-app.review.md`](S14-02-pictures-in-the-app.review.md) — refresh rounds 3 – 4

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

**US-1.** As a researcher, when an Event or Place has pictures, I see a Pictures section on its page with crops that open their Citations, so I can browse them as I browse likenesses on a Person. (S14-D2 EVP-1 – EVP-3, S14-D3 PLP-1 – PLP-3)

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Place detail whose `CatalogConclusionDetail.pictures` has one tile (Source title "Church photo") | `PlaceDetailContent.gallery.tiles` has one tile; `gallery.title` is `L10n.Conclusions.picturesSection` (key `conclusions.detail.pictures`, "Pictures"); tile `sourceTitle == "Church photo"`; tile `accessibilityLabel == "Picture from Church photo. Open citation"` |
| AC-1.2 | Event detail with two pictures | Same title; tile ids are the two observation ids in wire order |
| AC-1.3 | Event or Place detail with no pictures | `gallery.tiles` is empty; the default fill draws no gallery section (EVP-2, PLP-2) |
| AC-1.4 | Person detail with likenesses | `gallery.title` is `L10n.Conclusions.likenessesSection`; tile `accessibilityLabel == L10n.Conclusions.likenessTileLabel(source)`; `gallery.choose == .personPortrait` |

**US-2.** As a researcher, when an Event or Place has pictures, the page header and the list row lead with that picture, and I can choose another tile as the thumbnail or go back to the earliest. (EVP-4 – EVP-7, PLP-4 – PLP-7, EVL-1, PLL-1)

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Event detail whose first tile A (`imageRelPath = "objects/ab/cd.png"`) is `isLead` | `thumbnail.imageRelPath == "objects/ab/cd.png"`; `thumbnail.mark == .subjectEvent`; `thumbnail.label == L10n.Conclusions.eventPictureThumb`; `thumbnail.imageLabel == "Picture from <A's Source>"`; `gallery.offersEarliest == false` |
| AC-2.2 | Same Event; tile B (second) leads, A not | `gallery.offersEarliest == true`; `gallery.choose == .eventThumbnail`; B's `useAsLabel` is "Use the picture from <B's Source> as thumbnail" and A's likewise |
| AC-2.3 | Two tiles share one `imageRelPath` (identical crops; objects are content-addressed), the second leads | `thumbnail.imageLabel` names the second tile's Source (D-5) |
| AC-2.4 | Person detail whose second likeness B leads | `thumbnail.imageRelPath` is B's path; `label == personLikenessThumb`; `imageLabel == L10n.Conclusions.likenessLabel(<B's Source>)`; `colors` from the gender treatment, as today |
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
| AC-4.1 | After the PR, `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop|ConclusionPortraitModel|ValueFamily|depictionPerson|var likenesses|thumbnail\.offersEarliest|offersEarliest: (true|false)|likenessImageRelPath' macos/App` | Exit 1 (no match) |
| AC-4.2 | `CachedThumbnail.pictureCrop` | Equals `96` |
| AC-4.3 | `CitationComposerSubjectGroups` | Has `depictionSubject`; its comments say Subject |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionDetailGallery` *(new)* + `ConclusionDetailPicture` *(new)* on `ConclusionDetailBody`; remove `likenesses` | STEP-1, AC-1.*, D-1, D-2 |
| Rename `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection`, `LikenessTileView` → `GalleryTileView`, `CachedThumbnail.likenessCrop` → `pictureCrop`; ids per D-3 | STEP-1, AC-4.1, AC-4.2, D-3 |
| `ConclusionDetailThumbnail.leading(…)` *(new)*; `offersEarliest` moves from the thumbnail to the gallery; Person, Event and Place contents fill `gallery` and their thumbnails (EVP-1, EVP-4, PLP-1, PLP-4) | STEP-1, AC-1.*, AC-2.1 – AC-2.4, AC-2.8, D-4, D-5 |
| `ConclusionPortraitModel` → `ConclusionLeadModel`, built by the page from `userID`; copy read at call time from `gallery.choose`; Event and Place views pass `userID`; host passes it (EVP-5 – EVP-7, PLP-5 – PLP-7) | STEP-1, AC-2.2, AC-2.5, AC-2.6, D-6, D-7 |
| `EventsList` / `PlacesList` `thumbnailRelPath`; delete the protocol default (EVL-1, EVL-2, PLL-1, PLL-2) | STEP-2, AC-2.7, D-8 |
| Delete `L10n.Promote.ValueFamily`; composer `depictionSubject` | STEP-3, AC-4.1, AC-4.3, D-9, D-10 |
| Properties row VoiceOver label (S14-01 UX-1) | STEP-4, AC-3.*, D-11 |
| L10n: Event/Place picture and thumbnail functions and resources (§ Strings), Properties row label (EVP-8, PLP-8) | STEP-1, STEP-4, D-12 |
| The four boards' answers (S14-D2 – S14-D5), at Regular and Wide | STEP-1, STEP-2, §6 macOS, §10 |
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
| A new kit primitive under `DesignSystem/` | not planned (the S14-D2 – S14-D5 inventories) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Body protocol | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailBody` | `var likenesses: [ConclusionDetailLikeness]`, extension default `[]` |
| Default fill | same file `ConclusionDetailDefaultPageSections` | `ConclusionDetailRowsSection` titled `detailsSection` ("Details"), then `ConclusionDetailLikenessSection` when `likenesses` is non-empty, between rows and relationship sections |
| Tile type | same file `ConclusionDetailLikeness` | `id`, `imageRelPath`, `sourceTitle`, `location`, `isPortrait` |
| Gallery section | same file `ConclusionDetailLikenessSection` (internal since S13-10) | Init `(projectDir:likenesses:portrait:)`. Hard-codes `likenessesSection`, `likenessOpenTip`, `likenessTileLabel`, ids `persons.detail.likeness.\(tile.id).portraitBadge` / `.usePortrait`, `portraitBadge`, `useAsPortrait`, `useAsPortraitLabel` |
| Tile view | same file `LikenessTileView` (private) | `CachedThumbnail.likenessCrop` (96), fallback mark `.subjectPerson` |
| Thumbnail slot | same file `ConclusionDetailThumbnail` | `mark`, `label`, `imageRelPath`, `imageLabel`, `colors`, `offersEarliest` (read only by the header's `showsPortraitMenu`) |
| Header | same file `ConclusionDetailHeaderBand` (private; holds `pageIdentifier`, passes `portrait` and `compact: isScrolled`) → `ConclusionDetailHeader` (private) | `portrait: ConclusionPortraitModel?`; thumb 120pt, 40pt when compact; menu iff `portrait != nil && thumbnail.offersEarliest`; `portraitMenuTitle`, `portraitMenuHint`, `useEarliestLikeness`; ids `persons.detail.portrait`, `.menu`, `.useEarliest` |
| Page | same file `ConclusionDetailPage` | Synthesized memberwise init `(session, entityId, pageIdentifier, errorIdentifier, make, summary, portrait = nil, pageSections)` and a convenience init `(…, summary, portrait = nil)` in an extension for the default fill. `ConclusionDetailLoaded` (private; holds `portrait`) draws `PVPageColumns`; the loading page is `PVPageColumns(primary:)` |
| Context | same file `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing`, `portrait: ConclusionPortraitModel?`, `pageSection: ConclusionPageSection? = nil` |
| Person content | `PersonDetailContent.swift` | `likenessImageRelPath`, `likenesses`, `thumbnail` (labels `personNoLikeness` / `personLikenessThumb`, `colors`), `leadLabel` matched by path with fallback `personLikenessThumb` |
| Event / Place content | `EventDetailContent.swift`, `PlaceDetailContent.swift` | `thumbnail` = mark + `eventNoImage` / `placeNoImage`; ignore `header.pictureImageRelPath` and `detail.pictures`. Place rows Names and Period, then relationship `sections` |
| Portrait model | `ConclusionPortraitModel.swift` (pbxproj lines 101, 549, 1332, 2416) | `init(session:userID:entityId:)`; `choose(_: ConclusionDetailLikeness)`, `useEarliest()`; Person toast strings |
| Person view | `PersonDetailView.swift` | Owns `@State portrait` and `.vocabularyToastOverlay($portrait.toast, identifier: "persons.detail.toast")`; passes `portrait` through the full init with `pageSections: { PersonDetailSections(content: $0, context: $1) }`; private `PersonDetailSections` switches on `context.pageSection ?? .vitals`; `.gallery` case calls `ConclusionDetailLikenessSection(projectDir:likenesses:portrait:)` or `PVEmptyState` with `galleryEmptyTitle` / `galleryEmptyMessage` |
| Event / Place views | `EventDetailView.swift`, `PlaceDetailView.swift` | Convenience init; no `userID` |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` | Holds `userID` and passes it to several destinations, including `PersonDetailView` (line 101); not to `EventDetailView` (107) or `PlaceDetailView` (113). Pages are keyed by `viewIdentity` = entity id (`PlaceRegistry`) |
| List thumbs | `ConclusionListPage.swift` | `ConclusionListKind.thumbnailRelPath` with extension default `""` (line 72); only `PersonsList` overrides it; `PVList` thumbnail at its default 44pt |
| Leftover from S13-23 | `macos/App/Platform/L10n.swift` line 1356 | `private enum ValueFamily`, no reader |
| Composer | `CitationComposerSubjectGroups.swift` line 239 | `depictionPerson`, comments say person |
| Properties VO | `PropertiesView.swift` `rowAccessibilityLabel` (private, line 416) | `"\(label), \(valueType), \(origin), \(boundText)"` with the raw origin string; `PropertiesSnapshot.boundTypes(for:)`; `PropertiesOriginCell` maps `CatalogOrigin.user` → `originUserShort`, `.provenencia` → `originSeededShort`; `PropertiesModel.deleteAccessibilityLabel` is the feature's other spoken label |
| Stale text | `CachedThumbnail.swift` line 15 ("Promote evidence sheet"); `ConclusionListRow.swift` line 22 ("the likeness a Person leads with"); `docs/macos-client-patterns.md` line 102 ("likeness gallery"); `event-place-pictures.md` lines 32 (`likenesses` default), 33 and 64 (evidence sheet) | |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` (uses `ConclusionDetailLikeness`, `likenesses`, `isPortrait`, the portrait model), `PersonsListTests.swift` (line 21: `EventsList.thumbnailRelPath(event) == ""`), `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PropertiesModelTests.swift` (`makeModel` helper) | The test target has no synchronized groups; this PR adds no test file |

**Precedent.** Person gallery and portrait: `d31e1856` ("Let a researcher choose a portrait from the gallery and go back to the earliest from the header thumb.") and `e402e1fe`. List thumb: `5476b3b0`. One shared body field: `ConclusionDetailThumbnail` (S10-23 / S10-35). Formatted Strings held by content: `PersonDetailContent.Vital.accessibilityLabel`, `members`. Wire rename: #615.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Replace `likenesses` with one required `gallery: ConclusionDetailGallery` on `ConclusionDetailBody` (no default). The default fill draws the gallery section iff `!gallery.tiles.isEmpty` | Working rule "one slot" | A default gallery; keeping `likenesses` |
| D-2 | `ConclusionDetailGallery` holds `title: LocalizedStringResource`, `openTip: LocalizedStringResource`, `fallbackMark: PVMarkKey` (`.subjectPerson` / `.subjectEvent` / `.subjectPlace`), `tiles: [ConclusionDetailPicture]`, `choose: LeadCopy` (non-optional), and computes `lead` (`tiles.first(where: \.isLead)`) and `offersEarliest` (`tiles.first.map { !$0.isLead } ?? false`). `ConclusionDetailPicture` holds `id`, `imageRelPath`, `sourceTitle`, `location`, `isLead`, and ready Strings `accessibilityLabel`, `useAsLabel`, `chosenToastBody`. `static func tiles(from: CatalogConclusionDetail, tileLabel: (String) -> String, useAsLabel: (String) -> String, chosenBody: (String) -> String)` builds them, passing each picture's Source title to the three functions | Copy with arguments is formatted by L10n functions, once, in the content (`macos-l10n.mdc`; review M-1); one source per fact; the slot carries the fallback mark, as working rule "The gallery is one slot" lists | Format templates stored in the slot and filled by views |
| D-3 | Rename to `ConclusionDetailGallerySection`, `GalleryTileView`, `pictureCrop`. Ids are built from the page's `pageIdentifier` with fixed leaves, for every kind: tile `\(page).picture.\(tile.id)`, badge `…\(tile.id).leadBadge`, button `…\(tile.id).useAsLead`; header thumb `\(page).lead`, menu `\(page).lead.menu`, item `\(page).lead.useEarliest`. Person ids change from `persons.detail.portrait*` / `.likeness.*.portraitBadge` / `.usePortrait` (no test or UI test reads them: `rg -n 'detail\.portrait|\.usePortrait|detail\.likeness\.' macos` hits only `ConclusionDetailPage.swift`). The tile fallback mark is `gallery.fallbackMark` | One rule, no per-kind id fields (review M-3) | Id strings on the pack |
| D-4 | `LeadCopy` *(new, nested in `ConclusionDetailGallery`)* holds only static copy: `badge`, `useAs`, `useEarliest`, `menuTitle`, `menuHint` (`LocalizedStringResource`), `toastTitle`, `toastEarliestBody` (`String`, as today's Person toast members). Statics `.personPortrait` (today's Person members, table in § Strings), `.eventThumbnail`, `.placeThumbnail` (they differ only in `menuTitle`) | The section and header stay kind-blind; S13-14 is another pack | A kind enum in the section |
| D-5 | `static func leading(gallery: ConclusionDetailGallery, noImage: LocalizedStringResource, withImage: LocalizedStringResource, leadLabel: (String) -> String, colors: PVThumbnail.Colors? = nil) -> ConclusionDetailThumbnail` *(new)*: `mark` = `gallery.fallbackMark`; `imageRelPath` = `gallery.lead?.imageRelPath ?? ""`; `label` = `gallery.lead == nil ? noImage : withImage`; `imageLabel` = `gallery.lead.map { leadLabel($0.sourceTitle) }`. The page reads no header path: `GetConclusionDetail` sets the header path and every picture path from one lead rule and one `files.RelPaths` call (`api/ffi/handlers/conclusion_detail.go` `attachHeader`, `attachPictures`), so `gallery.lead` is the one source. Person: `personNoLikeness`, `personLikenessThumb`, `likenessLabel`, `treatment?.thumbnailColors`. Event: `eventNoImage`, `eventPictureThumb`, `pictureFrom`. Place: `placeNoImage`, `placePictureThumb`, `pictureFrom`. Each content stores `gallery` and then `thumbnail` as `var`s set in `init`, in that order (the Person sets `treatment` first). Delete `ConclusionDetailThumbnail.offersEarliest`; the header reads `content.gallery.offersEarliest` | One builder, one source for the lead (review M-4, M2-3); lead by `isLead` is right when two tiles share a path | Three copies; reading the header path beside the gallery; matching by path |
| D-6 | Rename `ConclusionPortraitModel` → `ConclusionLeadModel` and its file → `ConclusionLeadModel.swift` (edit the four pbxproj lines in §4 by hand: path, name and both comments). `init(session:userID:entityId:)` unchanged; `choose(_ picture: ConclusionDetailPicture, copy: LeadCopy)` toasts `copy.toastTitle` / `picture.chosenToastBody`; `useEarliest(copy: LeadCopy)` toasts `copy.toastTitle` / `copy.toastEarliestBody`. The `isSaving` guard and the danger toast on failure stay | Kind-neutral name beside `isLead` (review M-9); copy has one source, read at call time (review M-2) | Copy passed at init |
| D-7 | **The page owns the model.** `ConclusionDetailPage` replaces `portrait` with `userID: String` and gains `@State private var lead: ConclusionLeadModel`. Both inits become hand-written: full `init(session:entityId:pageIdentifier:errorIdentifier:userID:make:summary:pageSections:)` and convenience `init(session:entityId:pageIdentifier:errorIdentifier:userID:make:summary:)`; each sets `_lead = State(initialValue: ConclusionLeadModel(session: session, userID: userID, entityId: entityId))`. The model is always built: every kind can choose. The outermost view adds `.vocabularyToastOverlay($lead.toast, identifier: "\(pageIdentifier).toast")`. `ConclusionDetailPageSectionsContext.portrait` becomes `lead: ConclusionLeadModel` and gains `pageIdentifier: String`; `ConclusionDetailLoaded`, `ConclusionDetailHeaderBand` and `ConclusionDetailHeader` take `lead` (and the header `pageIdentifier`) in place of `portrait`; the header shows the menu iff `content.gallery.offersEarliest`, at both thumb sizes as today. `PersonDetailView` drops its `@State` and overlay and passes `userID` (its toast id stays `persons.detail.toast`); `EventDetailView` / `PlaceDetailView` take `userID` and pass it; the host passes `userID` to both | One machine, one owner, no Optional nobody sets to nil (review M-2); the page is keyed by entity id, so the model resets per entity | A `leadChoose` page parameter beside `gallery.choose`; copying the Person's `@State` into each view |
| D-8 | `EventsList` / `PlacesList` implement `thumbnailRelPath` as `header.pictureImageRelPath`; delete the extension default in `ConclusionListPage.swift` so a fourth kind must answer | No dead default (review M-7) | Keeping the default |
| D-9 | Delete `L10n.Promote.ValueFamily` | S13-23 removed its last reader; its `likeness` case is in this PR's rename sweep | Leaving it |
| D-10 | Rename `depictionPerson` → `depictionSubject`; its comments and its two call sites' comments say Subject | Map scope | Retitling catalog keys |
| D-11 | `PropertiesModel.rowAccessibilityLabel(for property: CatalogProperty) -> String` *(new)* beside `deleteAccessibilityLabel`. It reads `snapshot.boundTypes(for:)`, speaks the origin through `PropertiesOriginCell.spokenOrigin(_:) -> String` *(new static; `user` → `originUserShort`, `provenencia` → `originSeededShort`, a `plugin:` origin → `L10n.Origin.pluginPill(pluginID: CatalogOrigin.pluginID(from:))`, the words `OriginPill` shows; any other origin → the raw string, which `OriginPill` does not show and today's label speaks)* and formats with `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:)` *(new, "%1$@, %2$@, %3$@, %4$@, %5$@")*, where `bound` is today's `rowBoundAnnouncement(count:)`; when bound types exist it appends `", "` + their labels joined with `", "`. `PropertiesView` calls it; its private function goes | Testable through `makeModel` in `PropertiesModelTests`; origin spoken as the cell shows it (review M-10); no stray comma (X-10) | A raw origin string in a new key; a six-slot format with an empty last slot |
| D-12 | New L10n per § Strings, each a `static let` resource or a `static func … -> String` through `L10n.format`, as `add-localized-string` step 4 says. Person functions are reused as they are | `macos-l10n.mdc` | Reusing "Use as portrait" on an Event |
| D-13 | **On S13-10's page.** `PersonDetailSections`' `.gallery` case becomes `ConclusionDetailGallerySection(projectDir: context.projectDir, pageIdentifier: context.pageIdentifier, gallery: content.gallery, lead: context.lead)` when `!content.gallery.tiles.isEmpty`, else its `PVEmptyState` reading `galleryEmptyTitle` / `galleryEmptyMessage` as S13-10 wrote it. No empty-state field on the slot | One design on the landed page (review M-5, M-8) | Adding empty-state fields to the slot |
| D-14 | **The boards bind the chrome**, as §6 macOS describes them, with three answers. (a) The Pictures heading's meta is the bare count ("3"), as the Event board and the Person's Likenesses; the Place board's "3 pictures" is not built (user, 2026-10-30). (b) The no-picture header label stays `eventNoImage` / `placeNoImage` ("No image recorded"); the boards' mock text "No picture record" is not copy. (c) The lists keep today's Wide layout, the content column up to 1180pt, centered (`ConclusionListPage.page`); the briefs' "fills the width as today" was wrong about today, the Events board draws today's layout and the Places board's note says nothing moves. (d) The tile caption stays one line, truncated, with the full title in `.help`, for every kind; the boards wrap it to two lines, which would change the Person's tiles too. (e) The Toast stays top-trailing (`VocabularyToastOverlay`, `.topTrailing`); the boards' bottom Toast is not built. (f) The "· today" after the Place chain is S9-D7 mock text, not built | Repo rule: each view change has a board (user, 2026-10-30) | Building from the Person page alone; a worded count |

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
| Event/Place page with tiles | `ConclusionDetailGallerySection` | No section when empty; a tile opens `tile.location` as on the Person; a missing image shows `gallery.fallbackMark` |
| Event/Place page with a lead model | Header menu, tile badge and button | The page is keyed by `viewIdentity` (entity id), so the model is rebuilt per entity; no menu when `offersEarliest` is false; a tap while saving is ignored (the `isSaving` guard, as on the Person); a failed write shows the danger toast; a later success replaces it |
| Events/Places list row with a path | `PVList` thumbnail slot | Empty path → the kind's mark |

### Scale

At the map's sizes (one researcher's catalog): the gallery is a `LazyVGrid` over one entity's tiles, as for the Person; three Strings per tile are formatted once per content build; the list thumb is one path per row already on the header; the Properties label is O(bound types) per row. No new write.

### Edges

- **Assumptions.** The page reads the lead from `gallery.lead` alone; the engine sets the header path from the same lead rule in the same call (D-5), and the lists keep reading the header path (D-8). D-5 no longer assumes one tile per path (T-3).
- **Neighbors.** A Person keeps every string; its ids change per D-3 (T-1, T-5). An Event with one picture leads with it, so no menu (T-2). An empty Source title shows the Observation ref (`ReconciliationRecord.sourceTitle(for:)`), as today: "Picture from OBS-2".
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

The agreed frames, in words (D-14). Every frame is the pane only, Regular 1102 × 900 unless marked Wide 2296 × 1440.

**Event page (S14-D2).** 3a: "Wedding of John Smith and Mary Brown, 1902", header thumb is the lead crop (Wedding photograph, 1902), the date · place line, ref and member count at the right; "Details" with Date (Merged, 2 Sources, Why) and Place (1 Source, Why); then "Pictures" with meta "3" over a rule, three 96pt tiles left-aligned in the grid, each captioned with its Source title on one line, truncated, with the full title as the tooltip, as the Person's tile today (D-14d); under the lead tile the accent "Thumbnail" Badge, under the others the ghost "Use as thumbnail" button. 3c: after choosing "Family album, 1920", the header thumb shows it, the Badge sits under it, the first tile now offers the button, and the success Toast "Thumbnail updated" / "Using the picture from Family album, 1920 as the thumbnail." shows top-trailing, where the page's toast overlay draws it today (D-14e). 3d: right-click on the chosen header thumb opens the menu titled "Event thumbnail" with one row "Use earliest picture". 3e: one picture, Badge, no button. 3f: no picture, the Event mark in the header, no Pictures section. 3g: 3a in dark mode. 3b (Wide): the same page centered at the 1180pt column; eight tiles in one row instead of seven; no secondary column.

**Place page (S14-D3).** 2a: "St. Mary's Church", chain "Halifax, Nova Scotia, Canada", ref and members; "Details" with Names and Period; "Pictures" (meta "3", D-14a) with "Church photo, 1890" leading and badged; then "Part of", "Contains" and "Succession" as today. 2b: after choosing "Postcard, 1951", header, Badge and Toast as 3c. 2c: menu "Place thumbnail" / "Use earliest picture", hint "Right-click for thumbnail options". 2d: "Halifax" with one picture, Badge, no button, its Contains section below. 2e: no picture, Place mark, no section. 2f: 2a dark. 2g (Wide): centered 1180pt column, eight tiles a row, relationship sections share the column edge.

**Events list (S14-D4)** and **Places list (S14-D5).** Frame 1: six rows; two show the lead picture in the 44pt thumb with the 3pt corner the Persons list uses, the rest the kind's mark, and one whose image fails to load shows the mark. Titles, secondary lines, Updated column, refs, row height and spoken labels are unchanged; dark mode is the board's theme toggle. Frame 2 (Wide): today's list, centered in the content column (D-14c).

| Building block | Component | Home | Requirement |
| --- | --- | --- | --- |
| Gallery section | `PVSectionHeader` (title, bare count as meta, D-14a) + `LazyVGrid` (adaptive 120pt columns, as today) + `GalleryTileView` | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | EVP-1 – EVP-3, PLP-1 – PLP-3; below the rows, above a Place's relationship sections |
| Tile crop | `CachedThumbnail` | `GalleryTileView` | `pictureCrop`; fallback `gallery.fallbackMark` |
| Badge / Use as | `PVBadge` accent + photo glyph, `PVButton` ghost small | gallery section | EVP-5, PLP-5; copy from `gallery.choose`, labels from the tile |
| Earliest menu | `pvRightClickMenu` / `PVActionMenuItem` | `ConclusionDetailHeader` | EVP-6, PLP-6 |
| Toast | `.vocabularyToastOverlay` (draws `PVToast`) | `ConclusionDetailPage` | EVP-7, PLP-7 |
| Header thumb | `CachedThumbnail` 120pt | `ConclusionDetailHeader` | EVP-4, PLP-4 |
| List thumb | `PVList` thumbnail slot, default 44pt | `ConclusionListPage` | EVL-1 – EVL-3, PLL-1 – PLL-3; EVL-W1, PLL-W1: no layout change (D-14c) |
| Wide | `PVPageColumns`, primary-only (S13-10) | `ConclusionDetailLoaded` | EVP-W1, PLP-W1: no code change; the grid fills the primary column (up to 1180pt, about one tile more per row than the 1038pt Regular column) |

Sketch:

```swift
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var fallbackMark: PVMarkKey
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
| `toastThumbnailUpdatedTitle` (`String`, as the Person's) | `conclusions.thumbnail.toastTitle` | Thumbnail updated |
| `toastThumbnailChosenBody(_:)` | `conclusions.thumbnail.toastChosenBody` | Using the picture from %@ as the thumbnail. |
| `toastThumbnailEarliestBody` (`String`, as the Person's) | `conclusions.thumbnail.toastEarliestBody` | Using the earliest picture as the thumbnail. |
| `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:)` | `properties.table.rowAccessibility` | %1$@, %2$@, %3$@, %4$@, %5$@ |

`L10n.Sources.useAsThumbnail` ("Set cover", key `sources.page.useAsThumbnail`) is a different member in a different namespace; no clash.

Person (reused as they are): title `likenessesSection`, tip `likenessOpenTip`, tile `likenessTileLabel(_:)`, header `personNoLikeness` / `personLikenessThumb` / `likenessLabel(_:)`, `.personPortrait` = `portraitBadge`, `useAsPortrait`, `useEarliestLikeness`, `portraitMenuTitle`, `portraitMenuHint`, `toastPortraitUpdatedTitle`, `toastPortraitEarliestBody`; tile functions `useAsPortraitLabel(_:)`, `toastPortraitChosenBody(_:)`.

Copy the board's wording where it differs, and record it in the Landed note. **Sweep:** `rg -n 'detail\.likeness\.|detail\.portrait' macos` finds nothing after the PR.

## 7. Steps

**STEP-1 — Gallery slot, lead model and contents (one buildable unit, one commit)**
Files: `ConclusionDetailPage.swift`, `CachedThumbnail.swift`, `ConclusionPortraitModel.swift` → `ConclusionLeadModel.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`, `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`, `PersonDetailView.swift` (and its `PersonDetailSections`), `EventDetailView.swift`, `PlaceDetailView.swift`, `WorkspaceDestinationHost.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-ui-component` (compose only), `add-localized-string`.
Do: D-1 – D-7, D-12, D-13, in this order: (1) the L10n members and catalog entries; (2) the slot types, `leading(…)` and `tiles(from:…)`; (3) rename the model file and type and edit the four pbxproj lines, then build once to confirm the project file; (4) the page, context, header band and header; (5) the three contents; (6) the views and the host. The app target does not build between (4) and (6).
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
Files: `docs/macos-client-patterns.md`, `docs/deployment-plan/spike-14/event-place-pictures.md`, `macos/App/Platform/CachedThumbnail.swift` (comment), `macos/App/Features/Conclusions/ConclusionListRow.swift` (comment), `ConclusionLeadModel.swift` (doc comment), this plan, `../completed.md`, `../deployment-plan.md`, the four briefs `../design/S14-D2-event-page-pictures.md`, `../design/S14-D3-place-page-pictures.md`, `../design/S14-D4-events-list-thumbnail.md`, `../design/S14-D5-places-list-thumbnail.md` → `../design/archive/`, `../design/README.md`. Skill: —.
Do: D-13 docs: in `macos-client-patterns.md` line 102, say "gallery slot" and add one sentence: a Conclusion kind passes `userID`, and the page owns the lead model and its toast. In `event-place-pictures.md`, line 32 says `gallery` (one slot, three kinds); rewrite the line-33 "Promote." bullet (the evidence sheet is gone, S13-23); drop "the evidence sheet only lists them (S14-02)" at line 64. In `event-place-pictures.md` line 41, the slot's contents read "title, tile wording, fallback mark, tiles, choose copy" and drop "a body with no portrait marks no tile as the portrait"; line 47 names `ConclusionLeadModel`. In `deployment-plan.md` working rule "An Event or Place chooses its thumbnail" (line 87) and the S14-02 Ships entry (line 228), `ConclusionPortraitModel` → `ConclusionLeadModel`; in `../../spike-13/deployment-plan.md` S13-14's Context (line 422), `ConclusionPortraitModel` → `ConclusionLeadModel` and add "the `gallery` slot". The two Swift comments drop "Promote evidence sheet" and say "the picture a kind leads with"; the model's doc comment says it chooses the lead of any kind. Then the DoD's docs lines.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` reports no errors; `rg -n 'evidence sheet' docs/deployment-plan/spike-14/event-place-pictures.md macos/App/Platform/CachedThumbnail.swift` finds nothing.

Commits: one per behavior; the subject is one imperative sentence ending in a period that says what changed for the researcher ("Show Event and Place pictures in a Pictures section."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonDetailTests` (update the likeness tests) | AC-2.4 case; Person `gallery.tiles` ids, order, `isLead`; title `likenessesSection`; tile `accessibilityLabel` from `likenessTileLabel`; `choose == .personPortrait`; `thumbnail.label` and `imageLabel` as today | AC-1.4, AC-2.4, D-1, D-2, D-5 |
| T-2 | Swift | `EventDetailTests` *(new cases)* | No pictures → empty tiles, mark thumbnail; one picture → path, `eventPictureThumb`, "Picture from …", `offersEarliest == false`; two with B leading → `offersEarliest`, `choose == .eventThumbnail`, `useAsLabel` English, title `picturesSection` | AC-1.2, AC-1.3, AC-2.1, AC-2.2, D-5 |
| T-3 | Swift | `EventDetailTests` *(new cases)* | Two tiles with one path, second leads → `imageLabel` names the second Source | AC-2.3, D-5 |
| T-4 | Swift | `PlaceDetailTests` *(new cases)* | One picture "Church photo" → tile label and title; no picture → `placeNoImage`, nil `imageLabel` | AC-1.1, AC-2.8 |
| T-5 | Swift | `PersonDetailTests` model tests (rename to the lead model) | `FakeStore`: `choose(B, copy: .eventThumbnail)`, `useEarliest(copy: .eventThumbnail)` → recorded calls and thumbnail toast English; same with `.personPortrait` → today's English | AC-2.5, AC-2.6, D-6 |
| T-6 | Swift | `PersonsListTests` (extend) | `EventsList` / `PlacesList.thumbnailRelPath` = header path; `""` when unset | AC-2.7, D-8 |
| T-7 | Swift | `PropertiesModelTests` *(new cases)* | `makeModel` with Event and Place types and two "Picture" properties, warmed → the AC-3.1 string, a different AC-3.2 string; an unbound user property → no trailing ", "; a `plugin:findagrave` origin → the label contains "Supplied by the findagrave plugin" | AC-3.1, AC-3.2, AC-3.3, D-11 |
| T-8 | Gate | STEP-1, STEP-3 checks, DoD `rg` | AC-4.1 command exits 1; `rg -n 'pictureCrop: CGFloat = 96' macos/App` finds one line; `rg -n 'func depictionSubject' macos/App` finds one line | AC-4.1, AC-4.2, AC-4.3 |

Event and Place picture tests build details with a private helper in each test struct, copied from `PersonDetailTests`' `likeness(_:path:title:)`; no shared test file. Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests`, `PersonsListTests`, `PropertiesModelTests`.

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
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated): S14-D2, S14-D3, S14-D4, S14-D5
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md line 102 and event-place-pictures.md lines 32, 33, 64 (STEP-6)
- [ ] Sweep, wider: `rg -n 'ConclusionPortraitModel|ConclusionDetailLikeness|LikenessTileView' docs/deployment-plan/spike-13 docs/deployment-plan/spike-14 docs/*.md` hits only this plan, its review log, `completed.md` and `archive/`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes the slice Check true for this PR. Each page at Regular and at Wide (pane 1600pt or more), light and dark, against the S14-D2 – S14-D5 frames.

1. New project. Compose a Place picture on a Citation (Property Picture). Open the Place: the Pictures section shows the crop below Names and Period and above any relationship section; a click opens the Citation; the header shows the picture. (PLP-1, PLP-3, PLP-4)
2. Add a second Place picture. **Use as thumbnail** on it: the header and the Places list row show it, the Thumbnail Badge moves, the Toast reads "Thumbnail updated". Right-click the header thumb → **Use earliest picture** restores the first. (PLP-5 – PLP-7, PLL-1)
3. Repeat 1–2 on an Event; open an Event with no picture: no Pictures section, the mark in the header. (EVP-2, PLP-2)
4. Widen the window past 1600pt of pane: each page stays one centered column and the grid fills it; each list stays as today. (EVP-W1, PLP-W1, EVL-W1, PLL-W1)
5. Open a Person with likenesses: Gallery shows "Likenesses", "Portrait" and "Use as portrait"; choosing shows "Portrait updated".
6. VoiceOver (or Accessibility Inspector) on Properties: the `event_picture` and `place_picture` rows read differently.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A board detail §6 does not describe | §6 macOS is the spec; the boards are linked for reference only |
| The pbxproj rename | Four lines named in §4; the build check in STEP-1 catches a missed one |

**Stop and report if:**

- `CatalogConclusionDetail.pictures`, `CatalogPicture.leads` or a header's `pictureImageRelPath` is not as S14-01 left it.
- `ConclusionDetailBody` already has a gallery field from another PR.
- A Step cannot pass without renaming `SetPreferredLikeness` or the preferred column.
- The agreed board asks for a kit primitive the inventory does not mark New.
- `scripts/check-pr.sh` fails on a `main` breakage this PR did not cause.

## 12. Open questions

None.

## 13. PR

**Title:** S14-02 — Pictures in the app
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md`.
