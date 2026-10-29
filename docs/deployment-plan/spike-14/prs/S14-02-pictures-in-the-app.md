# S14-02 — Pictures in the app

**Status:** Blocked — re-plan (2026-10-29): US-3, D-10, D-14 and their steps leave this PR. The Promote evidence sheet they change is removed by Spike 13's [S13-23](../../spike-13/deployment-plan.md#s13-23--promote-stops-asking-for-pins), so the plan is refreshed with `plan-pr` before `ship-pr` runs it.
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: reuses the Person gallery section, header thumb, portrait controls and list thumbnail with Event/Place headings and wording; Promote and Properties change copy only)
**Board:** —
**Depends on:** S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615))
**Blocks:** —
**Branch:** `s14-02-pictures-in-the-app`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` at 0.2.0 until a release is cut, as S14-01 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (D-10 wire field only — no new Method)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/ffi-handlers.mdc`
**Context:** [event-place-pictures.md § What already works](../event-place-pictures.md#what-already-works), [§ Thumbnail parity](../event-place-pictures.md#thumbnail-parity), [§ What changes](../event-place-pictures.md#what-changes); working rules: "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "A depiction is worded as a picture outside the Person's own copy", "The gallery layer is named gallery, not likeness", "Batch 1 is cut by layer"; S14-01 ship follow-ups UX-1 (Properties VO) and UX-2 (Promote wording)
**Planned:** 2026-10-29 against `main` @ `504f4f9b`
**Review:** [`S14-02-pictures-in-the-app.review.md`](S14-02-pictures-in-the-app.review.md) — Fits with changes / Ready, 2 rounds; every blocker Fixed

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

**What stays.** RPC `SetPreferredLikeness` / column `preferred_likeness_observation_id`; Property key `likeness` and Person section title / portrait strings; gallery SQL/seeds (S14-01); Spike 13's redesigned Gallery section chrome (S13-14); rectangle-only drawing (S14-03). One wire field is added: neighbor kind on `PromoteClaimEvidenceRelationship` (D-10) — no new Method.

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
| AC-3.3 | `EvidenceBoard.build(…)` with row `kind: .place` for a depiction line on the Subject's own cards | `line.value` is `.crops(paths:…, mark: .subjectPlace)`; spoken label uses `spokenPicture` / `spokenPictureCount`; no `isPicture` / `isLikeness` property on `Line` |
| AC-3.4 | Same build with row `kind: .person` | Own-card depiction mark is `.subjectPerson`; sentence still "Same picture" / "Compare these pictures" (value-type wording, every kind) |
| AC-3.5 | Person-row board with a relationship whose `CatalogClaimEvidenceRelationship.kind == "event"` and a depiction line in that relationship's sections | Line and member crop marks are `.subjectEvent`, not `.subjectPerson` (mark from `relationship.kind`, not from `PromoteFlow.row`) |

**US-4.** As a researcher using VoiceOver on the Properties list, I can tell `event_picture` from `place_picture` even though both labels are "Picture".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Property `label: "Picture", key: "event_picture"`, bound only to Event | `rowAccessibilityLabel` contains `"event_picture"` and the Event type's `label` |
| AC-4.2 | Property `label: "Picture", key: "place_picture"`, bound only to Place | Label contains `"place_picture"` and the Place type's `label`; the two full strings differ |

**US-5.** As a contributor, after the Swift view-layer rename, gallery/picture names replace likeness view types; composer depiction helpers say Subject.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | After the PR, `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop|ValueFamily\.likeness|isLikeness|isPicture|depictionPerson' macos` | Exit 1 / no matches |
| AC-5.2 | `CachedThumbnail.pictureCrop` | Equals `96`; symbol `likenessCrop` is gone |
| AC-5.3 | `CitationComposerSubjectGroups` depiction helper | Named `depictionSubject`; no `depictionPerson`; comments say Subject |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionDetailGallery` *(new)* + `ConclusionDetailPicture` *(new)* on `ConclusionDetailBody`; remove `likenesses` | STEP-1, AC-1.*, D-1, D-2 |
| Rename `ConclusionDetailLikenessSection` → `ConclusionDetailGallerySection`; `LikenessTileView` → `GalleryTileView`; `CachedThumbnail.likenessCrop` → `pictureCrop`; a11y id `\(identifierRoot).picture.\(tile.id)` | STEP-1, AC-5.1, AC-5.2, D-3 |
| Shared `ConclusionDetailGallery.tiles(from:)` mapping; Person/Event/Place contents fill `gallery` + Event/Place header thumb from `pictureImageRelPath` | STEP-2, AC-1.*, AC-2.1, AC-2.2, D-4, D-5 |
| Static `ChooseCopy` packs (copy + a11y id leaves); page owns `ConclusionPortraitModel` from `userID` + pack; Event/Place/Person views pass `userID` + pack | STEP-3, AC-2.2, AC-2.3, AC-2.5, D-3, D-6, D-7 |
| `EventsList` / `PlacesList` `thumbnailRelPath` → `pictureImageRelPath` | STEP-4, AC-2.4, D-8 |
| Promote: `ValueFamily.picture`, drop `isLikeness`/`isPicture`, picture sentence/spoken keys; crop mark from row kind on own cards and from wired `relationship.kind` on relationship cards; Promote `worded` maps depiction → Picture noun | STEP-5, AC-3.*, D-9, D-10, D-14 |
| Wire `PromoteClaimEvidenceRelationship.kind` (neighbor subject-type key string); map in handler + `GoStore` + `CatalogClaimEvidenceRelationship.kind` | STEP-5, AC-3.5, D-10 |
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
| Seeds / gallery SQL / gallery package rename | S14-01 (landed) |
| New Promote Method or changing claim-evidence build rules in Go | not planned — D-10 only exposes `rel.Kind` already computed |
| Person Timeline event picture | Spike 13 R-26 |
| Pictures as matching weight for Event/Place | not planned ([event-place-pictures.md](../event-place-pictures.md) Stays later) |
| New kit primitive under `DesignSystem/` | not planned (compose existing `PVSectionHeader`, `CachedThumbnail`, `PVBadge`, `PVButton`, `PVActionMenu`) |
| Changing `ReconciledValueDisplay.string(for: .depiction)` / `L10n.Conclusions.likenessValue` | not planned — Person Details rows keep "Likeness"; Promote sheet uses D-14 `worded` override instead |
| CitationComposer L10n catalog comments that still say "person" on `depictionThumbLabel` / `depictionLineLabel` | not planned — values already take any Subject label; D-11 renames the Swift helper and its call-site comments only |

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
| Promote board | `macos/App/Features/Promote/PromoteEvidenceBoard.swift` | `isLikeness`; `EvidenceBoard.build` has no kind; crops have no mark; relationship cards call `neighbor` for names only |
| Promote cards | `macos/App/Features/Promote/PromoteEvidenceCards.swift` `EvidenceCrops` | `fallbackMark: .subjectPerson`; id prefix `"likeness"` |
| Promote relationship wire | `api/proto/engine.proto` `PromoteClaimEvidenceRelationship`; `api/ffi/handlers/promote_evidence.go` `claimEvidenceProto`; `GenealogyStore.swift` `CatalogClaimEvidenceRelationship`; `GoStore.swift` map | Go `claimevidence.Relationship.Kind` is the neighbor's kind and already drives `evidenceSectionsProto(rel.Kind, …)`; proto/Swift omit `kind` (fields through 19; next free is 20) |
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
| D-1 | Replace `likenesses` with one required `gallery: ConclusionDetailGallery` on `ConclusionDetailBody` (like `thumbnail`; no protocol default). Page draws the gallery section iff `!gallery.tiles.isEmpty`. Bodies with no pictures set `tiles: []` and still fill the other gallery fields they own, or use `ConclusionDetailGallery(tiles: [])` with the kind's chrome only when they also set choose — Person/Event/Place always construct a full gallery value in STEP-2 | Working rule "one slot"; avoids a fake `.empty` hybrid (M-6) | Default `gallery` that embeds Pictures title + `.subjectPerson`; keep `likenesses` array |
| D-2 | `ConclusionDetailGallery` holds: `title`, `openTip`, `fromSourceLabel`, `tileLabel`, `fallbackMark`, `identifierRoot`, `tiles`, optional `choose: ChooseCopy?`. `ConclusionDetailPicture` replaces `ConclusionDetailLikeness` (`isLead` replaces `isPortrait`). Format helpers `fromSource` / `tileAccessibility` on the gallery | Slot carries wording; Equatable | Kind enum inside the section |
| D-3 | Rename view symbols to gallery/picture (`ConclusionDetailGallerySection`, `GalleryTileView`, `pictureCrop`). Section and header are kind-blind: tile id `\(gallery.identifierRoot).picture.\(tile.id).\(choose.badgeSuffix)` / `.\(choose.useAsSuffix)`; header uses `choose.headerIdentifier`, `choose.headerIdentifier + ".menu"`, `choose.headerIdentifier + ".useEarliest"`. `ChooseCopy` carries those id strings plus all choose/toast L10n. Static packs: `ChooseCopy.personPortrait` (`badgeSuffix: "portraitBadge"`, `useAsSuffix: "usePortrait"`, `headerIdentifier: "persons.detail.portrait"`, Person portrait L10n), `ChooseCopy.eventThumbnail` / `.placeThumbnail` (`badgeSuffix: "thumbnailBadge"`, `useAsSuffix: "useThumbnail"`, `headerIdentifier: "events.detail.lead"` / `"places.detail.lead"`, thumbnail L10n). Content types and the page's portrait model both take the same pack | Working rule one slot; S13-14 is another pack; Person portrait ids preserved via the Person pack | Per-kind `if identifierRoot ==` in the section; hard-code Person vs Event leaf names in D-3 without putting them on `ChooseCopy` |
| D-4 | `ConclusionDetailGallery.tiles(from:)` static mapper in `ConclusionDetailPage.swift`; Person, Event and Place call it | One mapping | Three copy-pasted maps; a new helper file |
| D-5 | Event/Place `thumbnail` from `header?.pictureImageRelPath`, `imageLabel` via gallery `fromSource`, `offersEarliest` from first tile `!isLead`, empty labels `eventNoImage` / `placeNoImage`; filled-thumb labels `eventPictureThumb` / `placePictureThumb` | Header lead on wire (S14-01) | Ignore header path |
| D-6 | Keep type name `ConclusionPortraitModel`; init takes `ChooseCopy` (toast fields from the pack). `choose` takes `ConclusionDetailPicture` | One write path; pack is the single copy+id source (D-3) | Rename type; duplicate toast strings outside the pack |
| D-7 | `ConclusionDetailPage` takes `userID: String` and `leadChoose: ChooseCopy?`. When `leadChoose != nil`, the page owns `@State ConclusionPortraitModel` (init with that pack) and `.vocabularyToastOverlay` with id `"\(pageIdentifier).toast"`. Kind views pass the matching static pack (`PersonDetailView` → `.personPortrait`, Event → `.eventThumbnail`, Place → `.placeThumbnail`) and the body's `gallery.choose` is that same static (one source). Host passes `userID` into all three detail views | Page owns the machine once; kind views stay thin; S13-14 passes another pack | Copy PersonDetailView's portrait `@State` + overlay into Event and Place views |
| D-8 | `EventsList` / `PlacesList` `thumbnailRelPath` → `header.pictureImageRelPath` | Same as Persons | Custom load |
| D-9 | Promote: `ValueFamily.likeness` → `.picture`; delete `Line.isLikeness` (do not add `isPicture`); branch on `if case .crops = value` for id prefix `"picture"`, spoken path, and suppressing `thisSource`; new keys `lineSamePicture` / `lineComparePictures` / `spokenPicture*`; prune old likeness Promote keys. Person depiction lines use picture sentences | Working rule value-type wording; one depiction test (M-3) | Keep a bool beside `.crops`; per-kind sentences |
| D-10 | Wire neighbor kind: `PromoteClaimEvidenceRelationship.kind = 20` (string subject-type key: `person` / `event` / `place`, same shape as `PromoteGraphAlignmentRow.kind`). Handler sets `Kind: string(rel.Kind)` in `claimEvidenceProto` (always; `rel.Kind` is required on Go relationships). `CatalogClaimEvidenceRelationship.kind: String` (default `""` only for decoding empty fixtures — production maps non-empty). `GoStore` maps `r.kind`; FakeStore / board fixtures set `kind` on every relationship under test. `EvidenceBoard.build`: own cards use `kind.markKey`; relationship groups use `EvidencePrimaryKind(rawValue: relationship.kind)!.markKey` (force-unwrap: empty or unknown kind is a Stop condition, not a sheet-kind fallback). Pass that mark into `lineRow` and depiction `memberRow` crops. Regenerate with `./scripts/generate-proto.sh`. Skill: `add-ffi-handler` (field only) | Engine already has `rel.Kind`; page-row lookup misses off-page neighbors (`shareHeaderBare`); working rule line's kind | Rebuild kind from `PromoteFlow.row`; `?? kind.markKey` fallback; infer from property key |
| D-11 | Rename `depictionPerson` → `depictionSubject`; comments say Subject; leave catalog keys/values | Scope | Retitle catalog keys |
| D-12 | `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:kinds:)` with default `"%1$@, %2$@, %3$@, %4$@, %5$@, %6$@"`; kinds joined with `", "`; unbound kinds `""` | S14-01 UX-1 | Key only or kinds only |
| D-13 | New Conclusions L10n for Pictures section, picture from/tile, Event/Place thumbs, thumbnail choose pack strings used by `ChooseCopy.eventThumbnail` / `.placeThumbnail`. Person pack uses existing likeness/portrait keys (extract `likenessFrom` / `likenessTile` static resources). Promote keys per D-9. Add `L10n.Promote.pictureValue` "Picture" for D-14 | add-localized-string | Reuse "Use as portrait" on Event |
| D-14 | In `PromoteModel.board(of:)` only, `worded` maps `.depiction` to `L10n.string(L10n.Promote.pictureValue)` ("Picture"); all other values still use `ReconciledValueDisplay.string`. Leave `ReconciledValueDisplay` / `likenessValue` unchanged for Person Details rows | Working rule picture wording on the sheet; Person page copy stays | Change shared `likenessValue` globally; leave Promote members saying "Likeness" |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionDetailGallery` *(new)* | type | `ConclusionDetailThumbnail` | Thumbnail is the header slot; gallery is the section slot — same pattern, different field |
| `ConclusionDetailPicture` *(new)* | type | `ConclusionDetailLikeness` | Rename/replace that type (Deletes) |
| `ConclusionDetailGallery.ChooseCopy` *(new)* + static `.personPortrait` / `.eventThumbnail` / `.placeThumbnail` | type | Person L10n + hard-coded ids in the section | One pack for copy, toast, and a11y leaves; section stays kind-blind |
| `ConclusionDetailGallery.tiles(from:)` *(new)* | func | `PersonDetailContent` init `detail.pictures.map` | Lift to one mapper |
| `GalleryTileView` *(new name)* | view | `LikenessTileView` | Rename; take `fallbackMark` from gallery |
| Event/Place picture + thumbnail L10n keys *(new)* | keys | Person likeness/portrait keys | Parallel wording |
| Promote `lineSamePicture` / `lineComparePictures` / `spokenPicture*` / `pictureValue` *(new)* | keys | `lineSameLikeness` / `lineCompareLikenesses` / `spokenLikeness*` | Replace those keys (Deletes) |
| `L10n.Properties.rowAccessibility(label:key:valueType:origin:bound:kinds:)` *(new)* | key | Inline string concat in `rowAccessibilityLabel` | Format via L10n |
| `CatalogClaimEvidenceRelationship.kind` *(new field)* | wire field | `claimevidence.Relationship.Kind` already in Go; `PromoteGraphAlignmentRow.kind` string pattern | Expose existing value; do not invent a second kind source |

**Builds on.** `ConclusionDetailThumbnail` slot pattern; `ConclusionPortraitModel` + `setPreferredLikeness`; `ConclusionListKind.thumbnailRelPath`; S14-01 wire; `EvidencePrimaryKind.markKey`; `claimevidence.Relationship.Kind`.
**Deletes.** `ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`; protocol member `likenesses`; `CachedThumbnail.likenessCrop`; `ValueFamily.likeness` case; `Line.isLikeness` (no `isPicture` replacement); `depictionPerson`; Promote likeness sentence/spoken keys; a11y `.likeness.` → `.picture.` (Person header ids live on `ChooseCopy.personPortrait.headerIdentifier`).
**Next adds.** S13-14 is another `ChooseCopy` pack on the same slot. A fourth kind adds a pack + list `thumbnailRelPath`. Relationship kind on the wire is available to any later sheet chrome that needs neighbor marks.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Event/Place page with non-empty `gallery.tiles` | `ConclusionDetailGallerySection` | Section absent when tiles empty; tile navigation uses `tile.location` from records (same as Person); missing image path → `CachedThumbnail` fallback mark from gallery |
| Event/Place `gallery.choose` non-nil | Page-owned portrait model + header menu + tile badge/button | `offersEarliest` false → no menu (PDP-11); write failure → danger toast |
| Events/Places list row with `pictureImageRelPath` | `ConclusionListRow` thumb | Empty path → kind mark |
| Promote depiction on own cards / relationship neighbor | `EvidenceCrops` | Mark from row `kind` or wired `relationship.kind` (D-10); empty path still shows mark; off-page neighbor still has kind |

### Scale

At the map's dogfood sizes (single-researcher catalogs): gallery section is `LazyVGrid` over that entity's tiles (same as today for Person). List thumb is one `relPath` per row already loaded on the header. Promote board build stays one pass per evidence load; adding `kind` is O(1). Properties label build is O(bound types) per row, already paid for chips.

### Schema

None — no catalog change.

### Go core

None for gallery/seeds. `claimevidence.Relationship.Kind` already set; no build-logic change.

### FFI

`api/proto/engine.proto` `PromoteClaimEvidenceRelationship.kind = 20` *(new)* — string subject-type key. Handler `claimEvidenceProto` sets it from `rel.Kind`. Regenerate with `./scripts/generate-proto.sh`. No new `Method`. Skill: `add-ffi-handler`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Gallery section | `PVSectionHeader` + `LazyVGrid` + `GalleryTileView` | `ConclusionDetailPage.swift` `ConclusionDetailGallerySection` | Snowflakes stay in Features/Conclusions; no new DesignSystem type |
| Tile crop | `CachedThumbnail` | `GalleryTileView` | `size: pictureCrop`, `fallbackMark` from gallery |
| Choose badge / button | `PVBadge`, `PVButton` | gallery section when `gallery.choose != nil` and page has a portrait model | Ids from `ChooseCopy` suffixes |
| Lead menu | `pvRightClickMenu` / `PVActionMenuItem` | `ConclusionDetailHeader` | Copy + `headerIdentifier` from `ChooseCopy` |
| List thumb | existing list row thumb | `EventsList` / `PlacesList` | |
| Promote crops | `EvidenceCrops` | `PromoteEvidenceCards.swift` | Mark on `Value.crops` |

States: empty `tiles` → no section; loading/error → existing page presentation; large gallery → same adaptive grid. Accessibility from gallery + `ChooseCopy`; Properties per D-12.

Sketch — gallery type:

```swift
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var fromSourceLabel: LocalizedStringResource
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
        var badgeSuffix: String       // "portraitBadge" | "thumbnailBadge"
        var useAsSuffix: String       // "usePortrait" | "useThumbnail"
        var headerIdentifier: String  // "persons.detail.portrait" | "events.detail.lead" | …
        static let personPortrait: ChooseCopy = /* D-3 / D-13 */
        static let eventThumbnail: ChooseCopy = /* … */
        static let placeThumbnail: ChooseCopy = /* … */
    }

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

Protocol: required `var gallery: ConclusionDetailGallery { get }` (no default). Person sets `choose: .personPortrait`; Event `.eventThumbnail`; Place `.placeThumbnail`.

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
| `promote.sheet.pictureValue` / `L10n.Promote.pictureValue` | Picture | D-14 Promote `worded` for depiction |
| `properties.table.rowAccessibility` | %1$@, %2$@, %3$@, %4$@, %5$@, %6$@ | label, key, valueType, origin, bound, kinds |

Removed: Promote likeness sentence/spoken keys; `CachedThumbnail.likenessCrop`; `Line.isLikeness`. **Sweep:** `.likeness.` a11y segments; Promote id prefix `likeness`. Person L10n and `ChooseCopy.personPortrait` header ids **stay**.

Event/Place `openTip` passes `L10n.Conclusions.likenessOpenTip` ("Open citation").

## 7. Steps

**STEP-1 — Gallery slot types and view-layer rename**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Platform/CachedThumbnail.swift`, call sites of `likenessCrop` (`PromoteEvidenceCards.swift`). Skill: `add-ui-component` (compose only; no new kit type).
Do: Add `ConclusionDetailGallery` / `ConclusionDetailPicture` / `ChooseCopy` (+ three static packs) / `tiles(from:)` per D-2/D-3/D-4. Required `gallery` on the protocol. Rename section and tile view; section and header read chrome and ids only from `gallery` / `choose` (kind-blind). `GalleryTileView` takes `fallbackMark`. Rename `likenessCrop` → `pictureCrop`. Land STEP-1 and STEP-2 in one commit.
Check: `rg 'ConclusionDetailLikeness|LikenessTileView|likenessCrop' macos` empty; no new `.swift` file for gallery types; App target builds before STEP-3.

**STEP-2 — Person / Event / Place contents**
Files: `PersonDetailContent.swift`, `EventDetailContent.swift`, `PlaceDetailContent.swift`. Skill: `add-localized-string`.
Do: Each body builds a full `gallery` (D-1): Person likeness chrome + `choose: .personPortrait`; Event/Place Pictures chrome + `.eventThumbnail` / `.placeThumbnail`; `identifierRoot` equals that kind's `pageIdentifier` (`"persons.detail"`, `"events.detail"`, `"places.detail"`); tiles from mapper; Event/Place thumbnails per D-5.
Check: App target builds; gallery title keys match AC-1.*; Person tile ids start with `persons.detail.picture.`.

**STEP-3 — Page owns portrait model**
Files: `ConclusionDetailPage.swift`, `ConclusionPortraitModel.swift`, `PersonDetailView.swift`, `EventDetailView.swift`, `PlaceDetailView.swift`, `WorkspaceDestinationHost.swift`. Skill: —.
Do: D-6/D-7 — page takes `userID` + `leadChoose`; owns model and toast overlay when `leadChoose != nil`. Kind views pass `userID` and the static pack (same as `gallery.choose`); drop per-view portrait `@State`. Host passes `userID` to Event and Place. Model `choose(_ picture:)`.
Check: `rg 'EventDetailView\\(session' macos` shows `userID:`; Person toast English unchanged; `ConclusionPortraitModel(` constructed only inside `ConclusionDetailPage` (or its private loaded child), not in Event/Place views.

**STEP-4 — List thumbnails**
Files: `EventsListView.swift`, `PlacesListView.swift`. Skill: —.
Do: Override `thumbnailRelPath` to `header.pictureImageRelPath` on both lists.
Check: unit expectations in STEP-8; `PersonsList.thumbnailRelPath` unchanged.

**STEP-5 — Promote picture wording, relationship kind wire, and crop marks**
Files: `api/proto/engine.proto`, generated protobuf under `api/` and `macos/App/Platform/Generated/`, `api/ffi/handlers/promote_evidence.go`, `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift` (relationship fixtures), `L10n.swift`, `Localizable.xcstrings`, `PromoteEvidenceBoard.swift`, `PromoteEvidenceCards.swift`, `PromoteEvidenceDisplay.swift`, `PromoteModel.swift`. Skills: `add-localized-string`, `add-ffi-handler`.
Do: D-10 wire field + map; D-9 / D-14 wording; own cards use row `kind.markKey`; relationship groups use `EvidencePrimaryKind(rawValue: relationship.kind)!.markKey` per D-10. Apply mark to line and member depiction crops.
Check: `./scripts/generate-proto.sh` then `rg 'lineSameLikeness|lineCompareLikenesses|spokenLikeness|isLikeness|isPicture|ValueFamily\.likeness' macos` empty; a handler or board test asserts relationship `kind` round-trips; `python3 scripts/check-localizable-xcstrings.py` passes.

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
Do: Implement T-1 through T-9 (including T-6b, T-7b, T-7c). Update Person tests for `gallery` / `isLead` / pack. Every `EvidenceBoard.build` gains `kind:`. Update `.crops` assertions for marks; drop `isLikeness` expects. Run the AC-5.1 `rg` as T-9.
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
| T-6b | Swift | `PromoteEvidenceDisplayTests` spoken likeness cases (rename with D-9) | `spokenPicture` / `spokenPictureCount` English starts with "Picture from this source"; rename former likeness spoken tests to picture helpers/English | D-9 |
| T-7 | Swift | `PromoteEvidenceBoardTests` | Own-card depiction with `kind: .place` → `.crops(..., mark: .subjectPlace)`; `kind: .person` → `.subjectPerson`; no `isLikeness`/`isPicture` on `Line` | AC-3.3, AC-3.4, D-9, D-10 |
| T-7b | Swift | `PromoteEvidenceBoardTests` *(new case)* | Person-row `kind: .person` + relationship with `kind: "event"` and a depiction line (neighbor callback may return nil) → crop mark `.subjectEvent` on line and member | AC-3.5, D-10 |
| T-7d | Go | `api/ffi/handlers` promote evidence test *(extend)* | Fixture with an Event-kind relationship → response relationship `kind == "event"` | D-10 |
| T-7c | Swift | `PromoteEvidenceBoardTests` or display test | Expanded depiction member spoken value uses "Picture", not "Likeness" (D-14 `worded`) | D-14 |
| T-8 | Swift | `PropertiesRowAccessibilityTests` *(new)* | Two Picture properties different keys → different labels containing each key and bound type label | AC-4.1, AC-4.2, D-12 |

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
- Field number 20 on `PromoteClaimEvidenceRelationship` is already taken on `main` (rebase and pick the next free number).
- `claimevidence.Relationship.Kind` is empty for a shipped relationship (handler cannot set a truthful kind).
- `scripts/check-pr.sh` fails on an unrelated main breakage; report, do not expand scope.

## 12. Open questions

None.

## 13. PR

**Title:** S14-02 — Pictures in the app
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-02-pictures-in-the-app.md`.
