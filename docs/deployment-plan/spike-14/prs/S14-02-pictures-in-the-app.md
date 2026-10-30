# S14-02 — Pictures in the app

**Status:** Blocked — on board S14-D2. Refreshed 2026-10-30 after S13-23 landed ([#633](https://github.com/mendahu/provenencia/pull/633)): the Promote wording, the relationship `kind` wire field and their steps left this PR, and the PR now waits on a board because it changes how the Event and Place pages and lists look (user, 2026-10-30).
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D2](../design/S14-D2-event-place-pictures.md), board not yet agreed
**Board:** — (when agreed: the Claude Design link and frame numbers go here, and **Design → macOS** describes each frame in words, as S13-10 does for S13-D1)
**Depends on:** S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615)), S13-23 (landed [#633](https://github.com/mendahu/provenencia/pull/633)), S14-D2
**Blocks:** —
**Branch:** `s14-02-pictures-in-the-app`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, as S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [event-place-pictures.md § What already works](../event-place-pictures.md#what-already-works), [§ Thumbnail parity](../event-place-pictures.md#thumbnail-parity), [§ What changes](../event-place-pictures.md#what-changes); working rules: "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "A depiction is worded as a picture outside the Person's own copy", "The gallery layer is named gallery, not likeness", "Batch 1 is cut by layer"; S14-01 ship follow-up UX-1 (Properties VoiceOver); Spike 13's [S13-10 plan](../../spike-13/prs/S13-10-header-and-sections.md) §4 and §11 (shared code)
**Planned:** 2026-10-30 against `main` @ `1253dfaa`
**Review:** [`S14-02-pictures-in-the-app.review.md`](S14-02-pictures-in-the-app.review.md) — round 3 (refresh) below rounds 1–2 of 2026-10-29

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

Event and Place pages show a Pictures section, lead their headers with the picture the engine already sends, and let the researcher choose a thumbnail the way a Person chooses a portrait. Events and Places lists show that lead. The Swift view layer names the gallery gallery/picture; the Person page's likeness and portrait copy stays. The Properties list tells two Properties labelled "Picture" apart for VoiceOver.

**Why now.** S14-01 landed the seeds, the gallery keys, the header leads and the wire rename; bar outcomes 3, 5 and 7 need the app. Spike 13's Gallery section (S13-14) builds on the same slot.

**What stays.** RPC `SetPreferredLikeness` and column `preferred_likeness_observation_id`; Property key `likeness`, the Person section title and portrait strings; gallery SQL and seeds (S14-01); Spike 13's redesigned Gallery section (S13-14); rectangle-only drawing (S14-03). No wire, Go or schema change.

## 2. User stories

**US-1.** As a researcher, when an Event or Place has pictures, I see a Pictures section on its page with crops that open their Citations, so I can browse them as I browse likenesses on a Person. (S14-D2 EPP-1 – EPP-3)

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Place detail whose `CatalogConclusionDetail.pictures` has one tile (Source title "Church photo") | `PlaceDetailContent.gallery.tiles` has one tile; `gallery.title` is `L10n.Conclusions.picturesSection` (key `conclusions.detail.pictures`, "Pictures"); tile caption "Church photo"; tile label "Picture from Church photo. Open citation" |
| AC-1.2 | Event detail with two pictures | Same title "Pictures"; tile ids are the two observation ids in wire order; `EventDetailContent.gallery.tiles` comes from `ConclusionDetailGallery.tiles(from:)` |
| AC-1.3 | Event or Place detail with no pictures | `gallery.tiles` is empty; the page draws no gallery section (EPP-2) |
| AC-1.4 | Person detail with likenesses | `gallery.title` stays `L10n.Conclusions.likenessesSection` ("Likenesses"); tile label stays likeness wording; existing Person tests pass with renamed members |

**US-2.** As a researcher, when an Event or Place has pictures, the page header and the list row lead with that picture, and I can choose another tile as the thumbnail or go back to the earliest. (EPP-4 – EPP-8)

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Event header `pictureImageRelPath = "objects/ab/cd.png"`, its tile A `isLead == true` and first | `EventDetailContent.thumbnail.imageRelPath == "objects/ab/cd.png"`; `imageLabel == "Picture from <A's Source>"`; `offersEarliest == false` |
| AC-2.2 | Same Event; tile B leads (second), A first and not lead | `offersEarliest == true`; `gallery.choose == .eventThumbnail` |
| AC-2.3 | Two tiles share one `imageRelPath` (identical crops; objects are content-addressed), the second leads | Header `imageLabel` names the second tile's Source (lead found by `isLead`, not by path) |
| AC-2.4 | `ConclusionPortraitModel` with `.eventThumbnail`: `choose(B)` then `useEarliest()` on a `FakeStore` | Each calls `setPreferredLikeness` with that entity id (observation B's id, then `""`); toast titles are "Thumbnail updated"; bodies "Using the picture from <B's Source> as the thumbnail." and "Using the earliest picture as the thumbnail." |
| AC-2.5 | `EventsList.thumbnailRelPath(header)` / `PlacesList.thumbnailRelPath(header)` | Equals `header.pictureImageRelPath`; `""` when the header has none |
| AC-2.6 | Person page choose / use earliest with `.personPortrait` | Toast and badge English unchanged ("Portrait updated", "Portrait", "Use as portrait", today's bodies) |
| AC-2.7 | Place page with no picture | `thumbnail.imageRelPath == ""`, `label == L10n.Conclusions.placeNoImage`, `offersEarliest == false` |

**US-3.** As a researcher using VoiceOver on the Properties list, I can tell `event_picture` from `place_picture` even though both labels are "Picture".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Property `label: "Picture", key: "event_picture"`, bound only to the Event type | `PropertiesSnapshot.rowAccessibilityLabel(for:)` contains `"event_picture"` and the Event type's `label` |
| AC-3.2 | Property `label: "Picture", key: "place_picture"`, bound only to the Place type | Contains `"place_picture"` and the Place type's `label`; the two strings differ |

**US-4.** As a contributor, the Swift view layer says gallery/picture, and the composer's depiction helper says Subject.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | After the PR, `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop|ValueFamily|depictionPerson|var likenesses' macos/App` | Exit 1 (no match) |
| AC-4.2 | `CachedThumbnail.pictureCrop` | Equals `96` |
| AC-4.3 | `CitationComposerSubjectGroups` | Has `depictionSubject`; comments on it say Subject |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionDetailGallery` *(new)* + `ConclusionDetailPicture` *(new)* on `ConclusionDetailBody`; remove `likenesses` | STEP-1, AC-1.*, D-1, D-2 |
| Rename `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection`, `LikenessTileView` → `GalleryTileView`, `CachedThumbnail.likenessCrop` → `pictureCrop`; tile a11y ids `\(identifierRoot).picture.\(tile.id)…` | STEP-1, AC-4.1, AC-4.2, D-3 |
| Shared `ConclusionDetailGallery.tiles(from:)` and `leadLabel` helpers; Person, Event and Place contents fill `gallery`; Event and Place header thumbs from `pictureImageRelPath` (EPP-1, EPP-4) | STEP-2, AC-1.*, AC-2.1 – AC-2.3, AC-2.7, D-4, D-5 |
| Static `ChooseCopy` packs; the page owns `ConclusionPortraitModel` from `userID` + pack; Event, Place and Person views pass both; host passes `userID` to all three (EPP-5 – EPP-7) | STEP-3, AC-2.2, AC-2.4, AC-2.6, D-3, D-6, D-7 |
| `EventsList` / `PlacesList` `thumbnailRelPath` → `pictureImageRelPath` (EPP-8) | STEP-4, AC-2.5, D-8 |
| Delete the private `L10n.Promote.ValueFamily` enum S13-23 left with no reader; composer `depictionSubject` rename | STEP-5, AC-4.1, AC-4.3, D-9, D-10 |
| Properties row VoiceOver label via `PropertiesSnapshot.rowAccessibilityLabel(for:)` *(new)* with key and bound type labels | STEP-6, AC-3.*, D-11 |
| L10n: Pictures section, picture tile and header labels, thumbnail choose strings, Properties row label (EPP-9) | STEP-1 – STEP-3, STEP-6, D-12 |
| The S14-D2 board's answers: each EPP / EPP-W requirement as the board draws it, at Regular and Wide | STEP-1 – STEP-4, §6 macOS, §10 |
| Swift tests | STEP-7, T-* |
| Docs: map, `completed.md`, brief archive, `event-place-pictures.md` stale evidence-sheet lines | STEP-8 |

| Out | Home |
| --- | --- |
| Spike 13's redesigned Gallery section and viewer | S13-14, S13-15 |
| Sections or a switcher on the Event or Place page | not planned (S13-D1 R-1) |
| Rectangle-only drawing; composer draw feedback | S14-03 |
| Renaming `preferred_likeness_observation_id` or `SetPreferredLikeness` | not planned (working rule "The gallery layer is named gallery") |
| Changing the Person section title or portrait copy | not planned (working rule: likeness stays on the Person) |
| Promote picture wording, relationship `kind` on the wire | Descoped 2026-10-29 — the evidence sheet is gone (S13-23); S13-24 removes the claim-evidence wire |
| Changing `ReconciledValueDisplay.string(for: .depiction)` / `L10n.Conclusions.likenessValue` | not planned — a depiction row on a Person keeps "Likeness"; Event and Place pages show no depiction row (their rows are fixed) |
| Composer catalog comments on `depictionThumbLabel` / `depictionLineLabel` that say "person" | not planned — their values already take any Subject label; D-10 renames the Swift helper and its comments only |
| Person Timeline event picture | Spike 13 R-26 |
| Pictures as matching weight for Event or Place | not planned ([event-place-pictures.md](../event-place-pictures.md) Stays later) |
| A new kit primitive under `DesignSystem/` | not planned (S14-D2 inventory: compose `PVSectionHeader`, `CachedThumbnail`, `PVBadge`, `PVButton`, the right-click menu, `PVToast`) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Body protocol | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailBody` | `var likenesses: [ConclusionDetailLikeness]` with extension default `[]`; `ConclusionDetailDefaultPageSections` draws `ConclusionDetailLikenessSection` when non-empty, between the rows and the relationship `ForEach` |
| Tile type | same file `ConclusionDetailLikeness` | `id`, `imageRelPath`, `sourceTitle`, `location`, `isPortrait` |
| Gallery section | same file `ConclusionDetailLikenessSection` (private) | Hard-codes `L10n.Conclusions.likenessesSection`, likeness tip and tile label, ids `persons.detail.likeness.\(tile.id).portraitBadge` / `.usePortrait`, `PVBadge(L10n.Conclusions.portraitBadge…)`, `PVButton(L10n.Conclusions.useAsPortrait…)` |
| Tile view | same file `LikenessTileView` (private) | `CachedThumbnail.likenessCrop` (96), fallback mark `.subjectPerson` |
| Thumbnail slot | same file `ConclusionDetailThumbnail` | Kind-neutral (`mark`, `label`, `imageRelPath`, `imageLabel`, `colors`, `offersEarliest`) |
| Header menu | same file `ConclusionDetailHeader` | `showsPortraitMenu` = `portrait != nil && offersEarliest`; Person L10n `portraitMenuTitle` / `portraitMenuHint` / `useEarliestLikeness`; ids `persons.detail.portrait`, `.menu`, `.useEarliest` |
| Page and context | same file `ConclusionDetailPage`, `ConclusionDetailPageSectionsContext` | Page takes `portrait: ConclusionPortraitModel? = nil`; context carries `projectDir`, `refreshing`, `portrait` (S13-09) |
| Person content | `PersonDetailContent.swift` | `likenessImageRelPath` from header; maps `detail.pictures` → `likenesses`; `leadLabel` finds the tile by `imageRelPath` (wrong when two tiles share a content-addressed path) |
| Event / Place content | `EventDetailContent.swift`, `PlaceDetailContent.swift` | `thumbnail` is the mark + `eventNoImage` / `placeNoImage`; ignore `header.pictureImageRelPath` and `detail.pictures`. Place rows Names and Period, then relationship `sections` |
| Portrait model | `ConclusionPortraitModel.swift` | `init(session:userID:entityId:)`; `choose(_: ConclusionDetailLikeness)`; Person toast strings only |
| Detail views | `PersonDetailView.swift` (owns `@State portrait` + `.vocabularyToastOverlay(…, identifier: "persons.detail.toast")`); `EventDetailView.swift`, `PlaceDetailView.swift` (no `userID`) | |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` | Passes `userID` to `PersonDetailView` only (lines 101–113) |
| List thumbs | `PersonsListView.swift` `thumbnailRelPath`; `ConclusionListPage.swift` protocol default `""` (line 72) | Events and Places use the default |
| Leftover from S13-23 | `macos/App/Platform/L10n.swift` line 1349 `private enum ValueFamily { case name, date, likeness, other }` | No reader since the evidence sheet went |
| Composer | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` | `depictionPerson` (line 239) and comments that say person |
| Properties VO | `macos/App/Features/Properties/PropertiesView.swift` `rowAccessibilityLabel` (private, line 416) | `"\(label), \(valueType), \(origin), \(boundText)"`; two Picture rows read the same (S14-01 UX-1). `PropertiesSnapshot.boundTypes(for:)` gives the bound types |
| Crop size | `macos/App/Platform/CachedThumbnail.swift` | `static let likenessCrop: CGFloat = 96` |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `PersonsListTests.swift` (line 21 asserts `EventsList.thumbnailRelPath(event) == ""`), `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PropertiesModelTests.swift` (builds `PropertiesSnapshot`) | The test target has no synchronized groups: a new test file needs `project.pbxproj` entries, so this PR adds none |
| Shared code with S13-10 | Spike 13's S13-10 (Planned) | Makes `ConclusionDetailLikenessSection` internal, calls it from `PersonDetailSections` with `content.likenesses` / `context.portrait`, and adds `L10n.Conclusions.galleryEmptyTitle` / `galleryEmptyMessage`. Either PR may land first (D-13) |

**Precedent.** Person gallery and portrait: `d31e1856` ("Let a researcher choose a portrait from the gallery and go back to the earliest from the header thumb.") and `e402e1fe`. List thumb: `5476b3b0` (`thumbnailRelPath` on `ConclusionListKind`). One shared body field: `ConclusionDetailThumbnail` (S10-23 / S10-35). Wire rename: #615. Copy the Person mapping and portrait wiring; give Event and Place their own Picture / Thumbnail keys.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Replace `likenesses` with one required `gallery: ConclusionDetailGallery` on `ConclusionDetailBody` (no protocol default, like `thumbnail`). The page draws the gallery section iff `!gallery.tiles.isEmpty`. Person, Event and Place each build a full gallery value | Working rule "one slot"; no Person/Pictures hybrid default | A default `gallery` with Pictures title and `.subjectPerson`; keeping `likenesses` |
| D-2 | `ConclusionDetailGallery` holds `title`, `openTip`, `tileLabel` (format resource: "Picture from %@. Open citation" or the Person's likeness tile string), `leadLabel` (format resource: "Picture from %@" or the Person's `likenessLabel` string), `fallbackMark`, `identifierRoot`, `tiles`, `choose: ChooseCopy?`. `ConclusionDetailPicture` replaces `ConclusionDetailLikeness`; `isLead` replaces `isPortrait` | The slot carries the wording; Equatable | A kind enum inside the section |
| D-3 | Rename to `ConclusionDetailGallerySection`, `GalleryTileView`, `pictureCrop`. Section and header are kind-blind: tile ids `\(gallery.identifierRoot).picture.\(tile.id).\(choose.badgeSuffix)` / `.\(choose.useAsSuffix)`; header ids `choose.headerIdentifier`, `+ ".menu"`, `+ ".useEarliest"`. `ChooseCopy` carries those strings and all choose, menu and toast copy. Static packs: `.personPortrait` (`"portraitBadge"`, `"usePortrait"`, `"persons.detail.portrait"`, today's Person L10n), `.eventThumbnail` / `.placeThumbnail` (`"thumbnailBadge"`, `"useThumbnail"`, `"events.detail.lead"` / `"places.detail.lead"`, thumbnail L10n) | One slot; S13-14 is one more pack; Person header ids kept | `if identifierRoot ==` in the section |
| D-4 | `static func tiles(from: CatalogConclusionDetail) -> [ConclusionDetailPicture]` and `func leadLabel(imageRelPath:) -> String?` on `ConclusionDetailGallery`, in `ConclusionDetailPage.swift`. `leadLabel` is nil when the path is empty, else formats `leadLabel` with the Source of `tiles.first(where: \.isLead)`, else nil (the thumbnail's `label` speaks) | One mapping; lead by `isLead` is right when two tiles share a content-addressed path (AC-2.3) | Three copied maps; matching by path |
| D-5 | Event / Place `thumbnail`: `imageRelPath` from `header?.pictureImageRelPath ?? ""`; `imageLabel` from `gallery.leadLabel`; `offersEarliest` = `tiles.first.map { !$0.isLead } ?? false`; `label` = `eventNoImage` / `placeNoImage` when the path is empty, else `eventPictureThumb` / `placePictureThumb`. Person uses the same helpers | The header lead is on the wire (S14-01) | Ignoring the header path |
| D-6 | Keep the type name `ConclusionPortraitModel`; `init(session:userID:entityId:copy: ChooseCopy)`; toast title and bodies from the pack; `choose(_ picture: ConclusionDetailPicture)` | One write path; the pack is the one source of copy | Renaming the type; a second toast source |
| D-7 | `ConclusionDetailPage` replaces `portrait:` with `userID: String` and `leadChoose: ChooseCopy?`. When `leadChoose != nil` the page builds `@State ConclusionPortraitModel` in its init (as `PersonDetailView` does today), puts it in `ConclusionDetailPageSectionsContext.portrait`, hands it to the header, and adds `.vocabularyToastOverlay($model.toast, identifier: "\(pageIdentifier).toast")`. Each kind view passes its static pack, the same value its content sets as `gallery.choose`. `PersonDetailView` drops its own `@State` and overlay; the toast id stays `persons.detail.toast` because `pageIdentifier` is `persons.detail` | The page owns the machine once; kind views stay thin; S13-10's `context.portrait` keeps working | Copying Person's `@State` and overlay into Event and Place views |
| D-8 | `EventsList` / `PlacesList` override `thumbnailRelPath` to `header.pictureImageRelPath` | Same as Persons | A custom image load |
| D-9 | Delete `L10n.Promote.ValueFamily` | S13-23 removed its last reader; its `likeness` case is in this PR's rename sweep | Leaving dead code with a likeness name |
| D-10 | Rename `depictionPerson` → `depictionSubject`; its comments and its two call sites' comments say Subject; catalog keys and values stay | Map scope | Retitling catalog keys |
| D-11 | Move the row label to `PropertiesSnapshot.rowAccessibilityLabel(for property: CatalogProperty) -> String` *(new)*, formatted by `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:kinds:)` *(new)*, default `"%1$@, %2$@, %3$@, %4$@, %5$@, %6$@"`; `kinds` = `boundTypes(for:).map(\.label).joined(separator: ", ")`, `""` when unbound. `PropertiesView` calls it; its private `rowAccessibilityLabel` goes. Tested in `PropertiesModelTests.swift` | Testable without a view or a new test file; the snapshot already owns `boundTypes` | Key only or kinds only; a new test file (needs pbxproj wiring) |
| D-12 | New `L10n.Conclusions` keys for the section, tile and header labels and the thumbnail pack (§ Strings). The Person pack uses today's keys; `likenessLabel` / likeness tile strings become static resources the gallery formats | `add-localized-string` | Reusing "Use as portrait" on an Event |
| D-13 | **Landing order with S13-10.** If S13-10 has not landed: rename as above. If it has: also rename its `PersonDetailSections` call to `ConclusionDetailGallerySection(projectDir:gallery:portrait:)` reading `content.gallery`, branch on `content.gallery.tiles.isEmpty`, and add `emptyTitle` / `emptyMessage: LocalizedStringResource?` to `ConclusionDetailGallery`, set to `galleryEmptyTitle` / `galleryEmptyMessage` by the Person only (Event and Place: nil, EPP-2) | The map: "whichever lands second rebases"; S13-10 §11 names the same seam | Leaving S13-10's empty keys beside the slot |
| D-14 | **The board binds the chrome.** Spacing, tile size, badge and button placement, list-thumb size and the Wide arrangement follow the agreed S14-D2 frames, written into §6 macOS when the board is agreed. Until then §6 lists the brief's requirements | Repo rule: each view change has a board (user, 2026-10-30) | Building from the Person page alone |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionDetailGallery` *(new)* | type | `ConclusionDetailThumbnail` | Thumbnail is the header slot; the gallery is the section slot, same pattern |
| `ConclusionDetailPicture` *(new)* | type | `ConclusionDetailLikeness` | Replaces it (Deletes) |
| `ConclusionDetailGallery.ChooseCopy` *(new)* + `.personPortrait` / `.eventThumbnail` / `.placeThumbnail` | type | Person L10n and hard-coded ids in the section and header | One pack for copy, toast and ids; the section stays kind-blind |
| `ConclusionDetailGallery.tiles(from:)`, `leadLabel(imageRelPath:)` *(new)* | func | `PersonDetailContent` init map and `leadLabel` | Lifted to one place and fixed (D-4) |
| `GalleryTileView` *(new name)* | view | `LikenessTileView` | Rename; takes the fallback mark |
| Event/Place picture and thumbnail keys *(new)* | keys | Person likeness / portrait keys | Parallel wording |
| `PropertiesSnapshot.rowAccessibilityLabel(for:)`, `L10n.Properties.rowAccessibility` *(new)* | func, key | `PropertiesView.rowAccessibilityLabel` (private, string concat) | Moved and localized |

**Builds on.** The `ConclusionDetailThumbnail` slot; `ConclusionPortraitModel` and `setPreferredLikeness`; `ConclusionListKind.thumbnailRelPath`; S14-01 wire; `PropertiesSnapshot.boundTypes`.
**Deletes.** `ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`, `likenesses` on the protocol and its default; `PersonDetailContent.likenessImageRelPath` and its path-matching `leadLabel`; `CachedThumbnail.likenessCrop`; `PersonDetailView`'s `@State portrait` and overlay; the page's `portrait:` parameter; `PropertiesView.rowAccessibilityLabel`; `L10n.Promote.ValueFamily`; `depictionPerson`; ids with a `.likeness.` segment.
**Next adds.** S13-14 is another pack on the slot. A fourth kind adds a pack and a list `thumbnailRelPath`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Event/Place page with tiles | `ConclusionDetailGallerySection` | No section when empty; a tile opens `tile.location` as on the Person; a missing image shows the gallery's mark |
| Event/Place `choose` | Page-owned model, header menu, tile badge and button | The model is built per page init, and the host keys the page by `viewIdentity` (the entity id, `PlaceRegistry`), so it resets on a new entity; no menu when `offersEarliest` is false; a failed write shows the danger toast |
| Events/Places list row with a path | `PVList` thumbnail slot | Empty path → the kind's mark |

### Scale

At the map's sizes (one researcher's catalog): the gallery is a `LazyVGrid` over one entity's tiles, as for the Person; the list thumb is one path per row already on the header; the Properties label is O(bound types) per row, already paid for its chips. No new write.

### Edges

- **Assumptions.** The header path and the tile `leads` flag come from one lead rule (S14-01 `gallery.Lead`). The label code no longer assumes one tile per path (D-4, T-3).
- **Neighbors.** A Person with likenesses keeps every string and id except the tile id segment (`.likeness.` → `.picture.`) (T-1, T-5). An Event with one picture: the lead is the earliest, so no menu (T-2).
- **Removed inputs.** A deleted picture drops out of `detail.pictures`; if it led, the engine's next lead fills the header, and `offersEarliest` follows the first tile (T-2).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — tiles come from the app's own catalog through the existing read.

### Schema

None — no catalog change.

### Go core

None.

### FFI

None — S14-01 put pictures and header paths on the wire.

### macOS

Draft from the S14-D2 brief; when the board is agreed, each frame is described here in words and the rows below take its answers (D-14).

| Building block | Component | Home | Requirement |
| --- | --- | --- | --- |
| Gallery section | `PVSectionHeader` (title, count as meta) + `LazyVGrid` + `GalleryTileView` | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | EPP-1, EPP-2, EPP-3; below the rows, above relationship sections |
| Tile crop | `CachedThumbnail` | `GalleryTileView` | `pictureCrop`, `gallery.fallbackMark` |
| Badge / Use as | `PVBadge` accent + photo glyph, `PVButton` ghost small | gallery section when `gallery.choose != nil` | EPP-5; ids from `ChooseCopy` |
| Earliest menu | `pvRightClickMenu` / `PVActionMenuItem` | `ConclusionDetailHeader` | EPP-6; copy and ids from `ChooseCopy` |
| Toast | `.vocabularyToastOverlay` | `ConclusionDetailPage` | EPP-7 |
| Header thumb | `CachedThumbnail` 120pt | `ConclusionDetailHeader` | EPP-4 |
| List thumb | `PVList` thumbnail slot | `ConclusionListPage` via `thumbnailRelPath` | EPP-8, EPP-W2 |
| Wide | Page columns, primary-only (after S13-10), or today's centered column (before) | `ConclusionDetailPage` | EPP-W1; the grid adds tiles per row |

States: no tiles → no section; loading and failure → today's page states. Accessibility from the gallery and `ChooseCopy` (EPP-9).

Sketch:

```swift
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var tileLabel: LocalizedStringResource   // "… from %@. Open citation"
    var leadLabel: LocalizedStringResource   // "… from %@"
    var fallbackMark: PVMarkKey
    var identifierRoot: String               // the kind's pageIdentifier
    var tiles: [ConclusionDetailPicture]
    var choose: ChooseCopy?

    struct ChooseCopy: Equatable {
        var badge, useAs, useEarliest, menuTitle, menuHint, toastTitle, toastEarliestBody: LocalizedStringResource
        var useAsLabel: LocalizedStringResource       // "Use the … from %@ as …"
        var toastChosenBody: LocalizedStringResource  // "Using the … from %@ as the …."
        var badgeSuffix: String, useAsSuffix: String, headerIdentifier: String
        static let personPortrait: ChooseCopy = /* today's Person keys */
        static let eventThumbnail: ChooseCopy = /* § Strings */
        static let placeThumbnail: ChooseCopy = /* § Strings */
    }

    static func tiles(from detail: CatalogConclusionDetail) -> [ConclusionDetailPicture]
    func leadLabel(imageRelPath: String) -> String?
}

struct ConclusionDetailPicture: Equatable, Identifiable {
    var id: String
    var imageRelPath: String
    var sourceTitle: String
    var location: WorkspaceLocation
    var isLead: Bool = false
}
```

### Strings

| Key / member | Default | Notes |
| --- | --- | --- |
| `conclusions.detail.pictures` / `picturesSection` | Pictures | Event and Place section title |
| `conclusions.picture.tile` / `pictureTile` | Picture from %@. Open citation | Tile label |
| `conclusions.picture.from` / `pictureFrom` | Picture from %@ | Header image label |
| `conclusions.event.pictureThumb` / `eventPictureThumb` | Event picture | Header label while a picture leads |
| `conclusions.place.pictureThumb` / `placePictureThumb` | Place picture | same |
| `conclusions.thumbnail.badge` | Thumbnail | |
| `conclusions.thumbnail.useAs` | Use as thumbnail | |
| `conclusions.thumbnail.useAsLabel` | Use the picture from %@ as thumbnail | |
| `conclusions.thumbnail.useEarliest` | Use earliest picture | |
| `conclusions.event.thumbnailMenuTitle` / `conclusions.place.thumbnailMenuTitle` | Event thumbnail / Place thumbnail | |
| `conclusions.thumbnail.menuHint` | Right-click for thumbnail options | |
| `conclusions.thumbnail.toastTitle` | Thumbnail updated | |
| `conclusions.thumbnail.toastChosenBody` | Using the picture from %@ as the thumbnail. | |
| `conclusions.thumbnail.toastEarliestBody` | Using the earliest picture as the thumbnail. | |
| `properties.table.rowAccessibility` | %1$@, %2$@, %3$@, %4$@, %5$@, %6$@ | label, key, value type, origin, bound count, bound type labels |

Event and Place `openTip` is today's `L10n.Conclusions.likenessOpenTip` ("Open citation"). Copy the board's wording where it differs, and record the change in the Landed note. **Sweep:** `.likeness.` id segments (`rg -n 'detail\.likeness\.' macos` finds nothing after the PR).

## 7. Steps

**STEP-1 — Gallery slot and the view-layer rename**
Files: `ConclusionDetailPage.swift`, `CachedThumbnail.swift`. Skill: `add-ui-component` (compose only).
Do: D-1 – D-4. Add the types, packs and helpers; required `gallery`; rename the section and tile view; section and header read chrome and ids from `gallery` / `choose`; `likenessCrop` → `pictureCrop`. Commit STEP-1 with STEP-2 (the target does not build between them).
Check: `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop' macos` finds nothing; no new `.swift` file.

**STEP-2 — Person, Event and Place contents**
Files: `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: each content builds its gallery with `identifierRoot` = its `pageIdentifier` (`persons.detail`, `events.detail`, `places.detail`) and its pack; thumbnails per D-5.
Check: the app target builds; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The page owns the portrait model**
Files: `ConclusionDetailPage.swift`, `ConclusionPortraitModel.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `PlaceDetailView.swift`, `WorkspaceDestinationHost.swift`. Skill: —.
Do: D-6, D-7, and D-13 if S13-10 has landed.
Check: `rg -n 'ConclusionPortraitModel\(' macos/App` hits only `ConclusionDetailPage.swift`; `rg -n 'EventDetailView\(|PlaceDetailView\(' macos/App/Features/Workspace` shows `userID:`.

**STEP-4 — List thumbnails**
Files: `EventsListView.swift`, `PlacesListView.swift`. Skill: —.
Do: D-8.
Check: T-6.

**STEP-5 — Dead enum and composer helper**
Files: `L10n.swift`, `CitationComposerSubjectGroups.swift`. Skill: —.
Do: D-9, D-10.
Check: `rg -n 'ValueFamily|depictionPerson' macos` finds nothing.

**STEP-6 — Properties VoiceOver**
Files: `PropertiesSnapshot.swift`, `PropertiesView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-11.
Check: T-7.

**STEP-7 — Tests**
Files: `PersonDetailTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PersonsListTests.swift`, `PropertiesModelTests.swift`. Skill: `add-swift-test`.
Do: T-1 – T-8; update Person tests for `gallery`, `isLead`, the pack and `.picture.` ids; change line 21 of `PersonsListTests` to expect the header path.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-8 — Docs**
Files: `../completed.md`, `../deployment-plan.md`, this plan, `../design/S14-D2-event-place-pictures.md` → `../design/archive/`, `../design/README.md`, `../event-place-pictures.md`. Skill: —.
Do: the DoD's docs lines. In `event-place-pictures.md`, rewrite the "Promote." bullet under What already works (the evidence sheet is gone, S13-23) and drop "the evidence sheet only lists them (S14-02)" under Stays later.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` reports no errors.

Commits: one per behavior; the subject is one imperative sentence ending in a period that says what changed for the researcher ("Show Event and Place pictures in a Pictures section."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonDetailTests` (update the likeness and portrait tests) | Person `gallery.tiles` ids, order and `isLead`; title key `conclusions.detail.likenesses`; tile id prefix `persons.detail.picture.` | AC-1.4, D-1, D-3 |
| T-2 | Swift | `EventDetailTests` *(new cases)* | No pictures → empty tiles, mark thumbnail; one picture → path, label "Picture from …", no `offersEarliest`; two with B leading → `offersEarliest`, `choose == .eventThumbnail`, title key `conclusions.detail.pictures` | AC-1.2, AC-1.3, AC-2.1, AC-2.2, D-5 |
| T-3 | Swift | `EventDetailTests` *(new case)* | Two tiles with one `imageRelPath`, second leads → `imageLabel` names the second Source | AC-2.3, D-4 |
| T-4 | Swift | `PlaceDetailTests` *(new cases)* | One picture "Church photo" → tile label and title; no picture → `placeNoImage` | AC-1.1, AC-2.7 |
| T-5 | Swift | `PersonDetailTests` portrait model tests | `.eventThumbnail` on `FakeStore`: choose then use earliest → ids sent and thumbnail toast English; `.personPortrait` English unchanged | AC-2.4, AC-2.6, D-6 |
| T-6 | Swift | `PersonsListTests` (extend) | `EventsList` / `PlacesList.thumbnailRelPath` = header path; `""` when unset | AC-2.5, D-8 |
| T-7 | Swift | `PropertiesModelTests` *(new case)* | Snapshot with Event and Place types and two "Picture" properties → labels contain each key and type label and differ | AC-3.1, AC-3.2, D-11 |
| T-8 | Gate | STEP-5 / DoD `rg` | AC-4.1 command exits 1; `CachedThumbnail.pictureCrop == 96`; `depictionSubject` exists | AC-4.1, AC-4.2, AC-4.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonDetailTests`, `PersonsListTests`. No test queries a `persons.detail.likeness.*` id today (`rg -n 'detail.likeness' macos` hits only L10n keys and the section).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: event-place-pictures.md evidence-sheet lines (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes the slice Check true for this PR. Each page at Regular and at Wide (pane 1600pt or more), light and dark, against the S14-D2 frames.

1. New project. Compose a Place picture on a Citation (Property Picture). Open the Place: the Pictures section shows the crop; a click opens the Citation; the header shows the picture. (EPP-1, EPP-3, EPP-4)
2. Add a second Place picture. **Use as thumbnail** on it: the header and the Places list row show it, the Thumbnail Badge moves, the Toast reads "Thumbnail updated". Right-click the header thumb → **Use earliest picture** restores the first. (EPP-5 – EPP-8)
3. Repeat 1–2 on an Event; open an Event with no picture: no Pictures section, the mark in the header. (EPP-2)
4. Widen the window past 1600pt of pane: the grid shows more tiles per row in the centered column. (EPP-W1, EPP-W2)
5. Open a Person with likenesses: "Likenesses", "Portrait" and "Use as portrait" as before.
6. VoiceOver (or Accessibility Inspector) on Properties: the `event_picture` and `place_picture` rows read differently.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S13-10 lands while this PR is open | D-13 names both orders; S13-10 §11 names the same seam |
| The board changes a requirement (a Pictures empty state, a different list-thumb size) | D-14: the plan is refreshed from the agreed board before it is `Planned`; nothing is built before then |
| `LocalizedStringResource` in an Equatable struct | `ConclusionDetailThumbnail` already holds one |

**Stop and report if:**

- `CatalogConclusionDetail.pictures`, `CatalogPicture.leads` or a header's `pictureImageRelPath` is not as S14-01 left it.
- `ConclusionDetailBody` already has a gallery field from another PR that conflicts with D-1.
- A Step cannot pass without renaming `SetPreferredLikeness` or the preferred column.
- The agreed board asks for a kit primitive the inventory does not mark New.
- `scripts/check-pr.sh` fails on a `main` breakage this PR did not cause.

## 12. Open questions

None. (The plan stays `Blocked — on board S14-D2` until the board is agreed and its answers are copied in.)

## 13. PR

**Title:** S14-02 — Pictures in the app
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md`.
