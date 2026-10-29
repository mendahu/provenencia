# S14-02 — Pictures in the app

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: reuses the Person gallery section, header thumb, portrait controls and list thumbnail with Event/Place headings and wording; Promote and Properties change copy only)
**Board:** —
**Depends on:** S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615))
**Blocks:** —
**Branch:** `s14-02-pictures-in-the-app`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, as S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [event-place-pictures.md § What already works](../event-place-pictures.md#what-already-works), [§ Thumbnail parity](../event-place-pictures.md#thumbnail-parity), [§ What changes](../event-place-pictures.md#what-changes); working rules: "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "A depiction is worded as a picture outside the Person's own copy", "The gallery layer is named gallery, not likeness", "Batch 1 is cut by layer"; S14-01 ship follow-ups UX-1 (Properties VO) and UX-2 (Promote wording)
**Planned:** 2026-10-29 against `main` @ `504f4f9b`
**Review:** — (not reviewed yet)

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

Event and Place pages show a Pictures section, lead their headers with the picture the engine already sends, and let the researcher choose a thumbnail the way a Person chooses a portrait. Events and Places lists show that lead. Promote words depiction lines as pictures. The Swift view layer names the gallery gallery/picture; Person page likeness and portrait copy stay.

**Why now.** S14-01 landed the seeds, gallery keys, header leads and wire rename. Bar outcomes 3, 5, 6 and 7 need the app. Spike 13's Gallery section (S13-14) builds from the same slot.

**What stays.** RPC `SetPreferredLikeness` / column `preferred_likeness_observation_id`; Property key `likeness` and Person section title / portrait strings; Go/proto (no engine change); Spike 13's redesigned Gallery section chrome (S13-14); rectangle-only drawing (S14-03).

## 2. User stories

**US-1.** As a researcher, when an Event or Place has pictures, I see a Pictures section on its page with crops that open their Citations, so I can browse them as I browse likenesses on a Person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Place detail whose `CatalogConclusionDetail.pictures` has one tile (Source title "Church photo") | Body `gallery.tiles` has one tile; section title is `L10n.Conclusions.picturesSection` ("Pictures"); tile caption is "Church photo"; open tip and tile a11y use Picture wording |
| AC-1.2 | Event detail with two pictures | Same section title "Pictures"; tile ids follow observation ids; `EventDetailContent(detail:).gallery.tiles` maps `detail.pictures` via the shared mapper |
| AC-1.3 | Person detail with likenesses | Section title stays `L10n.Conclusions.likenessesSection` ("Likenesses"); tile a11y stays likeness wording; Person tests that asserted likeness copy still pass |

**US-2.** As a researcher, when an Event or Place has pictures, the page header and the Events/Places list row lead with that picture, and I can choose another tile as the thumbnail or go back to the earliest.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Event header `pictureImageRelPath = "objects/ab/cd.png"` and matching gallery tile with `leads == true` | `EventDetailContent.thumbnail.imageRelPath` is that path; `offersEarliest` is false |
| AC-2.2 | Same Event; second tile preferred (`leads` on B, first tile not lead) | Lead tile B wears the Thumbnail badge; non-lead tile offers "Use as thumbnail"; header `offersEarliest` is true; menu title/hint use thumbnail wording |
| AC-2.3 | `ConclusionPortraitModel.choose` / `useEarliest` on an Event entity with Event toast pack | Calls `setPreferredLikeness` with that entity id; toast title/body use thumbnail copy (not "Portrait updated") |
| AC-2.4 | `EventsList.thumbnailRelPath(header)` / `PlacesList.thumbnailRelPath(header)` | Equals `header.pictureImageRelPath`; empty path when the header has none |
| AC-2.5 | Person page choose / use-earliest | Toast and badge copy unchanged ("Portrait", "Use as portrait", existing toast bodies) |

**US-3.** As a researcher, when I compare two Places (or Events, or Persons) with depiction values on the Promote evidence sheet, depiction lines read as pictures and empty crops use that kind's mark.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `L10n.Promote.lineSentence(relation: .same, cause: .exact, propertyCause: .none, value: depictionValue)` | Resource key is `promote.sheet.line.samePicture` with default "Same picture" |
| AC-3.2 | `lineSentence(relation: .unknown, cause: .unjudged, propertyCause: .none, value: depictionValue)` | Resource key is `promote.sheet.line.comparePictures` with default "Compare these pictures" |
| AC-3.3 | `EvidenceBoard.build(evidence:kind:neighbor:labels:worded:orders:)` with `kind: .place` for a depiction line | `line.isPicture == true`; `line.value` is `.crops(paths:…, mark: .subjectPlace)`; spoken label uses `spokenPicture` / `spokenPictureCount` |
| AC-3.4 | Same build with `kind: .person` | `line.value` mark is `.subjectPerson`; sentence still "Same picture" / "Compare these pictures" (value-type wording, every kind) |

**US-4.** As a researcher using VoiceOver on the Properties list, I can tell `event_picture` from `place_picture` even though both labels are "Picture".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Property `label: "Picture", key: "event_picture"`, bound only to Event | `rowAccessibilityLabel` contains `"event_picture"` and the Event type's `label` |
| AC-4.2 | Property `label: "Picture", key: "place_picture"`, bound only to Place | Label contains `"place_picture"` and the Place type's `label`; the two full strings differ |

**US-5.** As a contributor, after the Swift view-layer rename, gallery/picture names replace likeness view types; composer depiction helpers say Subject.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | After the PR, `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop|ValueFamily\.likeness|isLikeness|depictionPerson' macos` | Exit 1 / no matches |
| AC-5.2 | `CachedThumbnail.pictureCrop` | Equals `96`; symbol `likenessCrop` is gone |
| AC-5.3 | `CitationComposerSubjectGroups` depiction helper | Named `depictionSubject`; no `depictionPerson`; comments say Subject |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionDetailGallery` *(new)* + `ConclusionDetailPicture` *(new)* on `ConclusionDetailBody`; remove `likenesses` | STEP-1, AC-1.*, D-1, D-2 |
| Rename `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection`; `LikenessTileView` → `GalleryTileView`; `CachedThumbnail.likenessCrop` → `pictureCrop`; a11y id `\(identifierRoot).picture.\(tile.id)` | STEP-1, AC-5.1, AC-5.2, D-3 |
| Shared `ConclusionDetailGallery.tiles(from:)` mapping; Person/Event/Place contents fill `gallery` + Event/Place header thumb from `pictureImageRelPath` | STEP-2, AC-1.*, AC-2.1, AC-2.2, D-4, D-5 |
| Event/Place `ConclusionPortraitModel` + toast/lead copy packs; wire `EventDetailView` / `PlaceDetailView` / `WorkspaceDestinationHost` with `userID` | STEP-3, AC-2.2, AC-2.3, AC-2.5, D-6, D-7 |
| `EventsList` / `PlacesList` `thumbnailRelPath` → `pictureImageRelPath` | STEP-4, AC-2.4, D-8 |
| Promote: `ValueFamily.picture`, `isPicture`, picture sentence/spoken keys, `EvidenceBoard.build` takes `kind`, `cropFallbackMark` on crops | STEP-5, AC-3.*, D-9, D-10 |
| Composer `depictionSubject` rename + Subject comments | STEP-6, AC-5.3, D-11 |
| Properties `rowAccessibilityLabel` includes `property.key` and bound type labels | STEP-7, AC-4.*, D-12 |
| L10n: Pictures section, Event/Place picture/thumbnail strings, Promote picture keys; prune renamed Promote likeness keys | STEP-1–5, D-13 |
| Swift tests: extend PersonDetailTests; Event/Place gallery+thumb tests; list thumb tests; Promote sentence/board; Properties a11y | STEP-8, T-* |
| Map status, `completed.md`, plan Landed note when merged | STEP-9 |

| Out | Home |
| --- | --- |
| Spike 13 redesigned Gallery section chrome / viewer | S13-14, S13-15 |
| Rectangle-only depiction drawing; composer draw feedback | S14-03 |
| Renaming column `preferred_likeness_observation_id` or RPC `SetPreferredLikeness` | not planned (working rule) |
| Changing Person section title away from "Likenesses" / portrait copy | not planned (working rule: likeness stays on Person) |
| Go / proto / seeds / gallery SQL | S14-01 (landed) |
| Person Timeline event picture | Spike 13 R-26 |
| Pictures as matching weight for Event/Place | not planned ([event-place-pictures.md](../event-place-pictures.md) Stays later) |
| New kit primitive under `DesignSystem/` | not planned (compose existing `PVSectionHeader`, `CachedThumbnail`, `PVBadge`, `PVButton`, `PVActionMenu`) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Body protocol | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailBody` | `var likenesses: [ConclusionDetailLikeness]`; default `[]`; page draws `ConclusionDetailLikenessSection` when non-empty |
| Tile type | same file `ConclusionDetailLikeness` | `id`, `imageRelPath`, `sourceTitle`, `location`, `isPortrait` |
| Gallery section | same file `ConclusionDetailLikenessSection` | Hard-codes `L10n.Conclusions.likenessesSection`, likeness open tip / tile label, `persons.detail.likeness.*` ids, portrait badge / "Use as portrait" |
| Tile view | same file `LikenessTileView` | `CachedThumbnail.likenessCrop` (96), fallback mark always `.subjectPerson` |
| Thumbnail slot | same file `ConclusionDetailThumbnail` | Kind-neutral already (`mark`, `label`, `imageRelPath`, `imageLabel`, `offersEarliest`); header menu hard-codes Person portrait L10n and `persons.detail.portrait*` ids |
| Person content | `macos/App/Features/Conclusions/PersonDetailContent.swift` | Maps `detail.pictures` → `likenesses`; header from `pictureImageRelPath`; portrait wording |
| Event content | `macos/App/Features/Conclusions/EventDetailContent.swift` | `thumbnail` is mark + `eventNoImage` only; no `likenesses` override; ignores `header.pictureImageRelPath` and `detail.pictures` |
| Place content | `macos/App/Features/Conclusions/PlaceDetailContent.swift` | Same as Event with `placeNoImage` |
| Portrait model | `macos/App/Features/Conclusions/ConclusionPortraitModel.swift` | `choose` / `useEarliest` call `setPreferredLikeness`; toast strings are Person portrait L10n only; parameter type `ConclusionDetailLikeness` |
| Person view | `macos/App/Features/Conclusions/PersonDetailView.swift` | Builds `ConclusionPortraitModel` with `userID`; passes `portrait:` |
| Event / Place views | `EventDetailView.swift`, `PlaceDetailView.swift` | No `userID`, no portrait model |
| Destination host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` | Passes `userID` into `PersonDetailView` only (lines ~98–110) |
| List thumbs | `PersonsListView.swift` `thumbnailRelPath`; `EventsListView.swift` / `PlacesListView.swift` | Persons returns `pictureImageRelPath`; Events/Places use protocol default `""` (`ConclusionListPage.swift` extension) |
| Promote board | `macos/App/Features/Promote/PromoteEvidenceBoard.swift` | `isLikeness`; `EvidenceBoard.build` has no kind; crops have no mark |
| Promote cards | `macos/App/Features/Promote/PromoteEvidenceCards.swift` `EvidenceCrops` | `fallbackMark: .subjectPerson`; id prefix `"likeness"` |
| Promote L10n | `macos/App/Platform/L10n.swift` `L10n.Promote.lineSentence` / `ValueFamily` | `.depiction` → `.likeness`; `lineSameLikeness` "Same likeness"; `lineCompareLikenesses` "Compare these likenesses"; `spokenLikeness*` |
| Composer | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` | `depictionPerson`; comments say person |
| Properties VO | `macos/App/Features/Properties/PropertiesView.swift` `rowAccessibilityLabel` | `"\(label), \(valueType), \(origin), \(boundText)"` — no key, no kind labels; two Picture rows collide for VoiceOver (S14-01 UX-1) |
| Crop size | `macos/App/Platform/CachedThumbnail.swift` | `static let likenessCrop: CGFloat = 96` |
| Tests | `PersonDetailTests.swift`, `PersonsListTests.swift`, `PromoteEvidenceDisplayTests.swift`, `PromoteEvidenceBoardTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift` | Person gallery/portrait covered; Events list thumb asserts `""`; Promote expects likeness keys; Event/Place detail have no picture cases |
| Wire (done) | `GenealogyStore` `CatalogPicture`, `pictures`, `pictureImageRelPath` on three headers | S14-01; app views still say likeness |

**Precedent.** Person gallery + portrait: `d31e1856` ("Let a researcher choose a portrait from the gallery and go back to the earliest from the header thumb.") and `e402e1fe` (Portrait Badge / earliest menu). List thumb requirement: `5476b3b0` (`thumbnailRelPath` on `ConclusionListKind`). Thumbnail slot precedent for one shared body field: `ConclusionDetailThumbnail` (S10-23 / S10-35). Gallery wire rename: #615 / S14-01. Copy the Person content mapping and portrait wiring; do not copy Person-only L10n into Event/Place — supply parallel Picture/Thumbnail keys. Do not add a second gallery section beside the slot.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Replace `likenesses` with one `gallery: ConclusionDetailGallery` on `ConclusionDetailBody` (default empty tiles). Page draws the gallery section iff `!gallery.tiles.isEmpty` | Working rule "one slot"; Spike 13 builds on it | Keep `likenesses` array + parallel Event array — two mechanisms |
| D-2 | `ConclusionDetailGallery` holds: `title`, `openTip`, `fromSourceLabel` and `tileLabel` (`LocalizedStringResource` with `%@`), `fallbackMark`, `identifierRoot` (Person `"persons.detail"`, Event `"events.detail"`, Place `"places.detail"`), `tiles: [ConclusionDetailPicture]`, optional `choose: ChooseCopy?`. `ConclusionDetailPicture` is the renamed tile (`isLead` replaces `isPortrait`). Format helpers on the gallery: `func fromSource(_ title: String) -> String` and `func tileAccessibility(_ title: String) -> String` call `L10n.format` with those resources | Slot carries wording (working rule); Equatable without closures; S13-14 can pass its own resources later | Enum of kinds inside the section that switches L10n — couples the section to Person/Event/Place; closures for formatters — not Equatable |
| D-3 | Rename view symbols: section → `ConclusionDetailGallerySection`; tile view → `GalleryTileView`; `CachedThumbnail.pictureCrop`; a11y ids `\(identifierRoot).picture.\(tile.id)` and choose controls under the same root; header menu ids `\(identifierRoot).lead.*` when `choose != nil` | Working rule gallery/picture naming; Person root stays `persons.detail` so existing UI-test prefixes keep the page; `picture` replaces `likeness` in the leaf | Keep `persons.detail.likeness` for Event pages — wrong kind; rename Person section title — working rule keeps Likenesses |
| D-4 | One shared mapper `ConclusionDetailGallery.tiles(from: CatalogConclusionDetail) -> [ConclusionDetailPicture]` as a static function on `ConclusionDetailGallery` in `ConclusionDetailPage.swift` (same file as the types). Person, Event and Place call it | One mapping; bodies only differ on copy and header path | Three copy-pasted `detail.pictures.map` blocks; a new helper file |
| D-5 | Event/Place `thumbnail` reads `header?.pictureImageRelPath`, sets `imageLabel` via gallery `fromSourceLabel`, `offersEarliest` from first tile `!isLead`, empty label stays `eventNoImage` / `placeNoImage`; add filled-thumb labels `eventPictureThumb` / `placePictureThumb` ("Picture") parallel to `personLikenessThumb` | Header lead already on wire (S14-01); parity with Person | Derive lead only from tiles and ignore header path — header is the list/detail contract |
| D-6 | Keep type name `ConclusionPortraitModel`; inject `ChooseCopy` toast fields at init (`toastTitle`, `toastChosenBody` resource, `toastEarliestBody`). `choose` takes `ConclusionDetailPicture`. Event/Place views construct the model with thumbnail `ChooseCopy`; Person keeps portrait `ChooseCopy` from existing L10n | Working rule names this model; one write path | Rename type to `ConclusionLeadModel` this PR — churn without product need; hard-code portrait toasts for Event — wrong copy |
| D-7 | `EventDetailView` / `PlaceDetailView` take `userID` like Person; `WorkspaceDestinationHost` passes `userID`. Pass `portrait:` into `ConclusionDetailPage` whenever the body has `gallery.choose != nil` | Same wiring as Person | Skip choose chrome on Event/Place and only show the image — fails bar outcome 5 |
| D-8 | `EventsList.thumbnailRelPath` / `PlacesList.thumbnailRelPath` return `header.pictureImageRelPath` | Same as Persons; protocol already requires the override | Custom per-row image load — engine already leads |
| D-9 | Promote: rename `ValueFamily.likeness` → `.picture`; `isLikeness` → `isPicture`; id prefix `"picture"`; change sentence defaults to "Same picture" / "Compare these pictures" under new keys `lineSamePicture` / `lineComparePictures`; rename `spokenLikeness*` → `spokenPicture*`; prune old likeness Promote keys from L10n + xcstrings. Person depiction lines use the same picture sentences (working rule: by value type for every kind) | Working rule "worded as a picture outside the Person's own copy" | Keep "likeness" sentences for Person rows and "picture" for Event/Place — second mechanism; only change Event/Place — violates value-type rule |
| D-10 | `EvidenceBoard.build(evidence:kind:neighbor:labels:worded:orders:)` adds `kind: EvidencePrimaryKind` as the second parameter. `EvidenceBoard.Value.crops` becomes `crops(paths: [String], mark: PVMarkKey)`; picture lines pass `kind.markKey`. `EvidenceCrops` reads that mark. `PromoteModel.board(of:)` passes `self.kind`. Member crops for depiction pins use the same mark | Working rule fallback mark from the line's kind | Infer mark from property key — fragile; leave `.subjectPerson` — wrong for Place; separate `cropFallbackMark` on `Line` beside `Value` — duplicate |
| D-11 | Rename `depictionPerson` → `depictionSubject`; update comments to Subject; leave `depictionThumbLabel` / `depictionLineLabel` catalog keys and `"%1$@ of %2$@"` values (already kind-agnostic) | Scope: translator comments + helper name | Retitle catalog keys to say Subject — optional churn; Out unless comments-only is insufficient for AC-5.3 |
| D-12 | `rowAccessibilityLabel` calls `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:kinds:)` with one catalog key whose English default is `"%1$@, %2$@, %3$@, %4$@, %5$@, %6$@"` (label, key, valueType, origin, bound announcement, bound type labels joined with `", "`). When unbound, `kinds` is `""`. Spoken string for `event_picture` ≠ spoken string for `place_picture` | S14-01 UX-1; key is stable; bound labels help VoiceOver | Key only without kinds — weaker; kinds only without key — two customs with same bindings could collide later |
| D-13 | New Conclusions L10n: `picturesSection` "Pictures"; `eventPictureThumb` / `placePictureThumb`; `pictureFrom` / `pictureTile` resources; thumbnail choose pack (`thumbnailBadge` "Thumbnail", `useAsThumbnail`, `useAsThumbnailLabel`, `useEarliestPicture`, `eventThumbnailMenuTitle` "Event thumbnail", `placeThumbnailMenuTitle` "Place thumbnail", `thumbnailMenuHint`, toast title/bodies). Person gallery passes existing `likenessesSection`, `likenessOpenTip`, and extracted `likenessFrom` / `likenessTile` `LocalizedStringResource` static lets (same catalog keys as today's format helpers). Promote keys per D-9 | add-localized-string; Person copy unchanged | Reuse Person "Use as portrait" on Event — wrong word |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionDetailGallery` *(new)* | type | `ConclusionDetailThumbnail` | Thumbnail is the header slot; gallery is the section slot — same pattern, different field |
| `ConclusionDetailPicture` *(new)* | type | `ConclusionDetailLikeness` | Rename/replace that type (Deletes) |
| `ConclusionDetailGallery.ChooseCopy` *(new)* | type | Person L10n portrait constants inlined in the section | Move copy onto the slot so Event/Place and S13-14 pass their own |
| `ConclusionDetailGallery.tiles(from:)` *(new)* | func | `PersonDetailContent` init `detail.pictures.map` | Lift to one mapper |
| `GalleryTileView` *(new name)* | view | `LikenessTileView` | Rename; take `fallbackMark` from gallery |
| Event/Place picture + thumbnail L10n keys *(new)* | keys | Person likeness/portrait keys | Parallel wording; do not reuse portrait strings |
| Promote `lineSamePicture` / `lineComparePictures` / `spokenPicture*` *(new)* | keys | `lineSameLikeness` / `lineCompareLikenesses` / `spokenLikeness*` | Replace those keys (Deletes) |
| `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:kinds:)` *(new)* | key | Inline string concat in `rowAccessibilityLabel` | Format via L10n |

**Builds on.** `ConclusionDetailThumbnail` slot pattern; `ConclusionPortraitModel` + `setPreferredLikeness`; `ConclusionListKind.thumbnailRelPath`; S14-01 wire (`pictures`, `pictureImageRelPath`); `EvidencePrimaryKind.markKey`.
**Deletes.** Types/aliases `ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`; protocol member `likenesses`; `CachedThumbnail.likenessCrop`; `ValueFamily.likeness`; `Line.isLikeness`; `depictionPerson`; Promote L10n keys `lineSameLikeness`, `lineCompareLikenesses`, `spokenLikeness`, `spokenLikenessCount` (and matching xcstrings); any `persons.detail.likeness.*` id strings rewritten to `.picture.*`.
**Next adds.** S13-14 sets a second gallery presentation on the same `gallery` slot (or replaces Person's) without forking Event/Place. A fourth conclusion kind with pictures fills `gallery` + list `thumbnailRelPath` + optional `ChooseCopy`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Event/Place page with non-empty `gallery.tiles` | `ConclusionDetailGallerySection` | Section absent when tiles empty; tile navigation uses `tile.location` from records (same as Person); missing image path → `CachedThumbnail` fallback mark from gallery |
| Event/Place `choose` non-nil + portrait model | Header lead menu + tile badge/button | `offersEarliest` false → no menu (existing PDP-11); `setPreferredLikeness` failure → danger toast (existing model) |
| Events/Places list row with `pictureImageRelPath` | `ConclusionListRow` thumb | Empty path → kind mark (existing list behavior) |
| Promote depiction line for event/place kind | `EvidenceCrops` | Mark from `kind`; empty path still shows mark |

### Scale

At the map's dogfood sizes (single-researcher catalogs): gallery section is `LazyVGrid` over that entity's tiles (same as today for Person). List thumb is one `relPath` per row already loaded on the header. Promote board build stays one pass per evidence load; adding `kind` is O(1). Properties label build is O(bound types) per row, already paid for chips.

### Schema

None — no catalog change.

### Go core

None — engine unchanged (S14-01).

### FFI

None — `SetPreferredLikeness` already used.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Gallery section | `PVSectionHeader` + `LazyVGrid` + `GalleryTileView` | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | Snowflakes stay in Features/Conclusions; no new DesignSystem type |
| Tile crop | `CachedThumbnail` | `GalleryTileView` | `size: pictureCrop`, `fallbackMark` from gallery |
| Choose badge / button | `PVBadge`, `PVButton` | gallery section when `choose != nil` && `portrait != nil` | |
| Lead menu | `pvRightClickMenu` / `PVActionMenuItem` | `ConclusionDetailHeader` | Copy from `gallery.choose`; ids from `identifierRoot` |
| List thumb | existing list row thumb | `EventsList` / `PlacesList` | |
| Promote crops | `EvidenceCrops` | `PromoteEvidenceCards.swift` | Mark from line |

States: empty gallery → no section (unchanged rule); loading/error → existing page presentation; large gallery → same adaptive grid as Person. Accessibility: section label = gallery title; tile labels from `tileLabel`; choose controls from `ChooseCopy`; Properties row per D-12.

Sketch — gallery type:

```swift
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var fromSourceLabel: LocalizedStringResource  // "%@" Source title
    var tileLabel: LocalizedStringResource
    var fallbackMark: PVMarkKey
    var identifierRoot: String
    var tiles: [ConclusionDetailPicture]
    var choose: ChooseCopy? = nil

    struct ChooseCopy: Equatable {
        var badge: LocalizedStringResource
        var useAs: LocalizedStringResource
        var useAsLabel: LocalizedStringResource
        var useEarliest: LocalizedStringResource
        var menuTitle: LocalizedStringResource
        var menuHint: LocalizedStringResource
        var toastTitle: String
        var toastChosenBody: LocalizedStringResource
        var toastEarliestBody: String
    }

    /// Default for bodies with no gallery; `tiles` empty hides the section.
    static let empty = ConclusionDetailGallery(
        title: L10n.Conclusions.picturesSection,
        openTip: L10n.Conclusions.likenessOpenTip,
        fromSourceLabel: L10n.Conclusions.pictureFrom,
        tileLabel: L10n.Conclusions.pictureTile,
        fallbackMark: .subjectPerson,
        identifierRoot: "",
        tiles: []
    )

    static func tiles(from detail: CatalogConclusionDetail) -> [ConclusionDetailPicture] {
        detail.pictures.map {
            ConclusionDetailPicture(
                id: $0.record.observationID,
                imageRelPath: $0.imageRelPath,
                sourceTitle: ReconciliationRecord.sourceTitle(for: $0.record),
                location: ReconciliationRecord.composerLocation(for: $0.record),
                isLead: $0.leads
            )
        }
    }
}

struct ConclusionDetailPicture: Equatable, Identifiable {
    var id: String
    var imageRelPath: String
    var sourceTitle: String
    var location: WorkspaceLocation
    var isLead: Bool = false
}
```

Protocol:

```swift
protocol ConclusionDetailBody: Equatable {
    var title: ConclusionTitleSource { get }
    var ref: String { get }
    var showsRef: Bool { get }
    var members: String { get }
    var rows: [ReconciledValueRowModel] { get }
    var sections: [ConclusionDetailSection] { get }
    var thumbnail: ConclusionDetailThumbnail { get }
    var badge: ConclusionDetailBadge? { get }
    var gallery: ConclusionDetailGallery { get }
}
extension ConclusionDetailBody {
    var gallery: ConclusionDetailGallery { .empty }
}
```

### Strings

New / replaced (exact English defaults):

| Key / member | Default | Notes |
| --- | --- | --- |
| `conclusions.detail.pictures` / `picturesSection` | Pictures | Event/Place section title |
| `conclusions.event.picture` / `eventPictureThumb` | Picture | Filled Event header a11y noun |
| `conclusions.place.picture` / `placePictureThumb` | Picture | Filled Place header a11y noun |
| `conclusions.picture.from` / `L10n.Conclusions.pictureFrom` | Picture from %@ | Header imageLabel resource (gallery formats it) |
| `conclusions.picture.tile` / `L10n.Conclusions.pictureTile` | Picture from %@. Open citation | Tile a11y resource |
| `conclusions.thumbnail.badge` | Thumbnail | |
| `conclusions.thumbnail.useAs` | Use as thumbnail | |
| `conclusions.thumbnail.useAsLabel` | Use the picture from %@ as thumbnail | |
| `conclusions.thumbnail.useEarliest` | Use earliest picture | |
| `conclusions.thumbnail.menuTitle` | Event thumbnail / Place thumbnail — **two keys** `eventThumbnailMenuTitle` "Event thumbnail", `placeThumbnailMenuTitle` "Place thumbnail" | |
| `conclusions.thumbnail.menuHint` | Right-click for thumbnail options | shared |
| `conclusions.thumbnail.toastTitle` | Thumbnail updated | |
| `conclusions.thumbnail.toastChosenBody` | Using the picture from %@ as the thumbnail. | |
| `conclusions.thumbnail.toastEarliestBody` | Using the earliest picture as the thumbnail. | |
| `promote.sheet.line.samePicture` | Same picture | replaces sameLikeness |
| `promote.sheet.line.comparePictures` | Compare these pictures | replaces compareLikenesses |
| `promote.sheet.spokenPicture` / `spokenPictureCount` | Picture from this source and %1$@. %2$@. (count variant keeps today's plural shape) | replaces spokenLikeness* |
| `properties.table.rowAccessibility` | %1$@, %2$@, %3$@, %4$@, %5$@, %6$@ | label, key, valueType, origin, bound, kinds |

Removed: Promote likeness sentence/spoken keys above; `CachedThumbnail.likenessCrop` name. **Old-scope copy to sweep:** section ids `persons.detail.likeness`, Promote id prefix `likeness`, comments saying Event/Place likeness. Person keys `conclusions.detail.likenesses` and portrait keys **stay**.

Event/Place `openTip` passes `L10n.Conclusions.likenessOpenTip` ("Open citation") — one shared tip key; no second identical key.

## 7. Steps

**STEP-1 — Gallery slot types and view-layer rename**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Platform/CachedThumbnail.swift`, call sites of `likenessCrop` (`PromoteEvidenceCards.swift`). Skill: `add-ui-component` (compose only; no new kit type).
Do: Add `ConclusionDetailGallery` / `ConclusionDetailPicture` / `tiles(from:)` per D-2/D-4. Replace protocol `likenesses` with `gallery`. Rename section and tile view; section reads all chrome from `gallery` + optional `portrait`; `GalleryTileView` takes `fallbackMark`. Rename `likenessCrop` → `pictureCrop`. Header menu uses `gallery.choose` copy and `gallery.identifierRoot` when choose is non-nil; if `portrait == nil` or `choose == nil`, no choose chrome (Person always passes both).
Check: `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop' macos` empty; no new `.swift` file added for the gallery types.

**STEP-2 — Person / Event / Place contents**
Files: `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`. Skill: `add-localized-string` (add Pictures / picture thumb keys used here).
Do: Person builds `gallery` with likeness title/labels, `.subjectPerson`, `identifierRoot: "persons.detail"`, portrait `ChooseCopy`, tiles from mapper; keep thumbnail behavior, switch `likenesses` → `gallery.tiles` / `isLead`. Event/Place: map tiles; set Pictures title and picture labels; `choose` thumbnail pack; fill `thumbnail.imageRelPath` from header `pictureImageRelPath` (D-5).
Check: compile; `PersonDetailTests` updated expectations still expressible (STEP-8 runs them).

**STEP-3 — Portrait model + Event/Place views**
Files: `ConclusionPortraitModel.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `PlaceDetailView.swift`, `WorkspaceDestinationHost.swift`. Skill: —.
Do: Portrait model takes full `ConclusionDetailGallery.ChooseCopy` at init (toast fields read from it; badge/button copy stay on the gallery for the section). `choose(_ picture: ConclusionDetailPicture)`. Person view passes portrait `ChooseCopy`. Event/Place take `userID`, own model + thumbnail `ChooseCopy`, `.vocabularyToastOverlay`, pass `portrait:`. Host passes `userID`.
Check: `rg 'EventDetailView\\(session' macos` shows `userID:`; Person toast tests still match portrait English.

**STEP-4 — List thumbnails**
Files: `EventsListView.swift`, `PlacesListView.swift`. Skill: —.
Do: Override `thumbnailRelPath` to `header.pictureImageRelPath` on both lists.
Check: unit expectations in STEP-8; `PersonsList.thumbnailRelPath` unchanged.

**STEP-5 — Promote picture wording and crop mark**
Files: `L10n.swift`, `Localizable.xcstrings`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift`, `PromoteModel.swift`. Skill: `add-localized-string`.
Do: D-9 and D-10. Update `lineSentence` family case `.picture`. Pass `kind` from `PromoteModel.board(of:)`.
Check: `rg 'lineSameLikeness|lineCompareLikenesses|spokenLikeness|isLikeness|ValueFamily' macos` empty for old names; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-6 — Composer Subject helper**
Files: `CitationComposerSubjectGroups.swift`. Skill: —.
Do: Rename `depictionPerson` → `depictionSubject`; comments say Subject (D-11).
Check: `rg depictionPerson macos` empty.

**STEP-7 — Properties VoiceOver**
Files: `PropertiesView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-12. Bound kinds = `bound.map(\.label).joined(separator: ", ")`.
Check: a test in STEP-8 can build two labels that differ on key.

**STEP-8 — Tests**
Files: `PersonDetailTests.swift`, `EventDetailTests.swift`, `PlaceDetailTests.swift`, `PersonsListTests.swift` (extend T-5 list-thumb cases here beside the existing `EventsList.thumbnailRelPath` assert), `PromoteEvidenceDisplayTests.swift`, `PromoteEvidenceBoardTests.swift`, `PromoteFlowTests.swift` (its `EvidenceBoard.build` call), `PropertiesRowAccessibilityTests.swift` *(new)* wired into the ProvenenciaTests target in `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: Implement T-1 through T-9. Update Person tests for `gallery` / `isLead` / `pictureCrop` renames. Every `EvidenceBoard.build` call site gains `kind:`. Update `.crops` equality assertions to include the associated mark. Run the AC-5.1 `rg` as the T-9 check.
Check: `scripts/check-pr.sh` PASS.

**STEP-9 — Docs**
Files: `completed.md`, `deployment-plan.md` (map row), this plan (status), `event-place-pictures.md` only if a sentence still says the app does not show Event/Place pictures (update "What already works" Detail chrome / Promote lines to past tense for this PR's landing — on land). Skill: —.
Do: per Definition of done (status Landed only when merging).
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md` clean after status Planned.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show Event and Place pictures in the detail gallery."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonDetailTests` update `likenessesBecomeTilesEarliestFirst` / portrait tests | Person `gallery.tiles` ids/order/isLead; `gallery.title` key is `conclusions.detail.likenesses`; Event/Place contents with pictures → non-empty `gallery.tiles` and title key `conclusions.detail.pictures` | AC-1.1, AC-1.2, AC-1.3, AC-2.5, D-1 |
| T-9 | Swift / gate | DoD sweep (STEP-8 check + check-pr) | `rg` old likeness view names empty under `macos/`; `CachedThumbnail.pictureCrop == 96`; `depictionSubject` exists and `depictionPerson` does not | AC-5.1, AC-5.2, AC-5.3 |
| T-2 | Swift | `EventDetailTests` *(new cases)* | Header path + pictures → thumbnail path, `offersEarliest`, gallery title key `conclusions.detail.pictures`, `choose != nil` | AC-1.2, AC-2.1, AC-2.2, D-5 |
| T-3 | Swift | `PlaceDetailTests` *(new cases)* | Same as T-2 for Place mark/labels | AC-1.1, AC-2.1, D-5 |
| T-4 | Swift | `PersonDetailTests` portrait model tests | Event toast pack on model → thumbnail toast English; Person pack unchanged | AC-2.3, AC-2.5, D-6 |
| T-5 | Swift | `PersonsListTests` (extend) | `EventsList.thumbnailRelPath` / `PlacesList.thumbnailRelPath` equal `pictureImageRelPath`; empty when unset | AC-2.4, D-8 |
| T-6 | Swift | `PromoteEvidenceDisplayTests` `lineSentencesFollowTheTable` | depiction → `lineSamePicture` / `lineComparePictures` keys | AC-3.1, AC-3.2, D-9 |
| T-7 | Swift | `PromoteEvidenceBoardTests` | `build(evidence:kind:.place,…)` depiction line → `isPicture` and `.crops(..., mark: .subjectPlace)`; `kind: .person` → mark `.subjectPerson` | AC-3.3, AC-3.4, D-10 |
| T-8 | Swift | `PropertiesRowAccessibilityTests` *(new)* or Properties test file | Two Picture properties different keys → different `rowAccessibilityLabel` strings containing each key and bound type label | AC-4.1, AC-4.2, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `PersonDetailTests`, `PromoteEvidenceDisplayTests`, `PromoteEvidenceBoardTests`, `PersonsListTests`.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: event-place-pictures.md "What already works" Detail chrome / Promote lines when this lands
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Makes slice Check bar outcomes 3, 5, 6, 7 true for this PR (1–2 and 8 need S14-01; 4 and 9 need S14-03).

1. New project. Compose a Place picture on a Citation (Property Picture). Open the Place: **Pictures** section shows the crop; click opens the Citation. Header shows the picture.
2. Add a second Place picture; **Use as thumbnail** on the second: header and Places list show it; badge on that tile; right-click header → **Use earliest picture** restores the first.
3. Repeat 1–2 on an Event.
4. Promote two Places that share the same picture: evidence line reads **Same picture** (or **Compare these pictures** when unjudged). Crops use the Place mark when empty.
5. Open a Person with likenesses: section still **Likenesses**; portrait controls unchanged.
6. VoiceOver (or Accessibility Inspector) on Properties: `event_picture` and `place_picture` rows announce differently (key and Event vs Place labels).
7. Confirm AC-5.1–AC-5.3 via the DoD `rg` sweep (T-9) before merge.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Plan size (many views) | Working rule allows it; steps capped at 9; Promote stays in unless STEP count would exceed ~10 — already fitted |
| S13-14 lands first on `ConclusionDetailLikenessSection` | Map says S14-02 moves it onto the slot; Stop if the section API this plan renames was replaced without a gallery slot |
| `ChooseCopy` + Equatable LocalizedStringResource | Same pattern as existing body fields holding `LocalizedStringResource`; if Equatable fails on resource, store `String` via `L10n.string` for toast-only fields (already `String` for some) |
| FakeStore Event/Place details omit `pictures` | Tests build `CatalogConclusionDetail` locally (as PersonDetailTests); no FakeStore seed required unless a view test needs it |

**Stop and report if:**

- `CatalogConclusionDetail.pictures` or header `pictureImageRelPath` no longer exist as §4 / S14-01 left them.
- `ConclusionDetailBody` already gained a different gallery slot from S13-14 that conflicts with D-1 (two gallery fields).
- Implementing Promote picture keys forces a second `ValueFamily` path per kind (would violate D-9).
- A Step cannot pass without renaming `SetPreferredLikeness` or the preferred column (Out).
- `scripts/check-pr.sh` fails on an unrelated main breakage; report, do not expand scope.

## 12. Open questions

None.

## 13. PR

**Title:** S14-02 — Pictures in the app
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md`.
