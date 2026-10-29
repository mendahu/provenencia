# Deployment plan — Spike 14

Spike 14 is an odds-and-ends spike (working rule "Even spikes are odds and ends"). It takes small features, UX refinements and fixes in batches; each batch is a slice with its own Check. The first batch pulls the "other keys" item and the transparent-polygon item from [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md), and adds thumbnail parity and rectangle-only depiction drawing. It also takes four Spike 10 likeness-audit leftovers that edit the same code: the gallery's image source (S14-01) and the composer's drawing feedback, captions and cap announcement (S14-03) (user, 2026-10-27, while planning Spike 12's audit fixes). Batch 1 is cut by layer (engine, app, drawing) because almost all of it reuses Person behavior that already exists (working rule "Batch 1 is cut by layer").

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a run of small, independent improvements, each batch landing on its own.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project. A batch adds its outcomes here when it is cut.

**Batch 2 — Matching calibration:**

10. **Thin name plus kinship is Weak.** Promoting a Source person recorded only as a short given name who shares a lined-up grandparent (or other non-birth/death link) with a matching neighbor reaches at least the Weak band — not "Little evidence" for that shape.

**Batch 1 — Event and Place pictures** (new catalogs; working rule "New seeds reach new catalogs only"):

1. **A place has pictures.** On a Citation of a photo of St Mary's Church, the composer offers Picture for the Place Subject; the saved line reads "Picture of St Mary's Church".
2. **An event has pictures.** The same for a wedding photo filed on the Event.
3. **The pages show them.** The Place and Event pages show a Pictures section with the crop, captioned with its Source; a tile opens its record, as a likeness tile does on the Person page.
4. **A depiction is a rectangle.** Drawing a likeness or a picture shows the rectangle tool alone; a Citation's own region still offers all seven tools. "Use the citation's region" on a circled entry sets the depiction to the circle's bounding rectangle.
5. **An event or place has a thumbnail.** The Event and Place page headers lead with the earliest picture; choosing another tile as the thumbnail (or going back to the earliest) works as a Person's portrait does, and the Events and Places lists lead with the same picture.
6. ~~**Promote says picture.**~~ Dropped 2026-10-29: the evidence sheet it worded is removed by Spike 13's [S13-23](../spike-13/deployment-plan.md#s13-23--promote-stops-asking-for-pins).
7. **Likenesses are unchanged.** A Person's likeness, gallery and portrait work as before.
8. **A tile shows its own picture.** Every gallery tile, and the header or list thumbnail it leads, shows the image stored with its own record; an edited reconciler cache row cannot show another picture.
9. **Drawing guides and reports.** While a region or a depiction is drawn, Set page, the locator chip × buttons and Find are disabled and the draw hint hides; Draw is blocked on a file the viewer cannot cut, and a failed cut shows with Save disabled; a PDF Citation's page can be set from the needs-page callout; the captions name the region buttons, not only Draw; an edit that saves no image does not say "Image to save"; VoiceOver hears the freehand tool's point cap.

## Where the notes went

| Dogfood note or idea | PR | Brief |
| --- | --- | --- |
| Promote Jake + grandparent read "Little evidence" under Weak (user, 2026-10-28); raise `segment.OtherSupport` | S14-08 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Other keys of the same value type" (place and event) | S14-01, S14-02 | — |
| Thumbnail parity for Event and Place (user, 2026-10-27) | S14-01, S14-02 | — |
| Promote wording for a non-Person depiction (map review K-4) | Descoped (2026-10-29): the evidence sheet goes in Spike 13's S13-23 | — |
| Rename the gallery layer away from "likeness" (map review K-5) | S14-01, S14-02 | — |
| Rectangle-only depiction drawing (user, 2026-10-27) | S14-03 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Transparent polygon" | Refused (working rule "A depiction is drawn as a rectangle") | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Gallery image from the record" (likeness audit SEC-4) | S14-01 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) composer feedback (UX-5, UX-6, UX-7), copy (UX-12, UX-13) and cap feedback, moved from Spike 12's audit planning (user, 2026-10-27) | S14-03 | — |

**Back in the ideas folder** — [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md): signature and the Spike 10 audit leftovers not taken here.

## Working rules

These are the calls for this spike. PR plans treat them as binding. A rule names its batch.

- **Even spikes are odds and ends.** From Spike 13, odd-numbered spikes are large new features and even-numbered spikes collect small features, UX refinements and fixes. An odds-and-ends spike adds work in batches: each batch is a slice with its own Check and bar outcomes, cut with [`plan-spike`](../../../.agents/skills/plan-spike/SKILL.md) when it is pulled in.
- **(Batch 1) Batch 1 is cut by layer.** One engine PR, one app PR, one drawing PR. The app PR touches the Event and Place pages and the Events and Places lists (the Promote evidence sheet left its scope on 2026-10-29), more than one view, because each piece reuses a Person behavior with no new chrome and no engine path. If its plan runs past about 10 steps, the Promote wording splits out as its own PR. Decided 2026-10-27, replacing a seven-PR cut.
- **(Batch 1) Event and Place each get their own picture Property, labelled "Picture".** `event_picture` bound to `event` and `place_picture` bound to `place`, value type `depiction`, cardinality `multiple`, locked, each with its own description. The data model is right first and the UI adapts to it: one key per kind keeps each description true to its kind, though a Property can bind to several kinds (`start_date`). Cost accepted: a composer run that saves an Event's Picture and then picks a Place clears the Property, and the researcher picks Picture again (`applyObservationSubject`). `likeness` stays the Person key. No Property is named `depiction`. [Design](event-place-pictures.md#two-properties-one-label). Binds S14-01.
- **(Batch 1) The gallery layer is named gallery, not likeness.** "Gallery" for the layer and "picture" for one item in it, in Go, proto and Swift (for example `core/database/likenesses` → `gallery`, `LikenessTile`, `LikenessDetail`, `CatalogLikeness`, `ConclusionDetailLikeness`, `LikenessTileView`, `CachedThumbnail.likenessCrop`; Promote's `isLikeness` goes with the evidence sheet in Spike 13's S13-23). Proto field numbers are kept, so the wire stays compatible. "Likeness" stays only for the Person key, the Person page's copy and the portrait. The column `canonical_entities.preferred_likeness_observation_id` keeps its name: a column rename is a migration for a name only, and `catalogmodel` and the docs say what it holds. S14-01 renames Go, proto and the Swift wire mapping; S14-02 the Swift view layer. Binds S14-01, S14-02.
- **(Batch 1) The gallery reads one list of gallery keys.** The gallery query filters on a list in `core/vocabkey` (`likeness`, `event_picture`, `place_picture`), not on `PropertyLikeness` alone, and a `subjectvocab` test holds each to a seeded `depiction` Property. The next depiction Property that belongs in a gallery is one entry. Binds S14-01.
- **(Batch 1) New seeds reach new catalogs only.** As in Spike 10: the two Properties come from Install, and no migration adds them to an existing catalog. Binds S14-01.
- **(Batch 1) An Event or Place chooses its thumbnail as a Person chooses its portrait.** Feature parity: the earliest picture leads unless one is chosen; the header, the gallery's chosen badge, set and go-back-to-earliest, and the list row all work as on a Person. One mechanism serves all three kinds: the preferred pointer, `canonicalentities.SetPreferredLikeness` (it checks gallery membership, not kind), the lead rule and `ConclusionPortraitModel`. No per-kind column or RPC. A Person's copy keeps "portrait"; an Event's or Place's says "thumbnail". S14-01 updates `conclusion-layer-data-model.md` (the pointer is the handle's, not the Person's). Binds S14-01, S14-02.
- **(Batch 1) The gallery is one slot on the detail body.** `ConclusionDetailBody` carries one gallery value (title, tile wording, chosen-badge and set wording, fallback mark, tiles) built by one tile mapping, in place of `likenesses` and the Person-only strings, ids and mark in `ConclusionDetailLikenessSection`. Precedent: `ConclusionDetailThumbnail`. Spike 13's Gallery section (S13-14) is built from the same slot, so the old and new galleries can live side by side and retiring the old one does not drop Event and Place galleries. Binds S14-02.
- **(Batch 1) A depiction is worded as a picture outside the Person's own copy.** The Promote evidence sheet words a depiction line by value type ("Same picture", "Compare these pictures") for every kind, and takes its fallback mark from the line's kind. Binds S14-02. **Descoped for Promote (2026-10-29):** Identity Claims no longer carry pins and Spike 13's S13-23 removes the evidence sheet; S14-02 drops its Promote wording and the relationship `kind` wire field. The rule still holds for the Event and Place pages.
- **(Batch 1) A depiction is drawn as a rectangle.** The composer arms the viewer with the rectangle tool alone for a depiction; a Citation's own region keeps all seven tools. "Use the citation's region" takes a non-rectangle region's bounding rectangle. Go keeps accepting any valid depiction region, so a depiction saved as a polygon loads, displays and saves as it is until redrawn. The transparent-polygon crop is refused. Binds S14-03.
- **(Batch 2) Other-link support is 0.75.** `segment.OtherSupport` is 0.75 (`OtherCredit` stays 0.25): a partial given name plus one lined-up non-birth/death link clears Weak on the 2026-10-28 Jake + grandparent dogfood shape, without a kinship-only tier or a name-curve change. Decided user 2026-10-28 (option A). Binds S14-08.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S14-01 | Pictures engine | 1 | Rename the gallery layer (Go, proto, Swift wire mapping); seed `event_picture` and `place_picture`; the gallery reads the gallery keys and each picture's own image; Event and Place headers carry the lead picture | — | — | [S14-01-pictures-engine.md](prs/S14-01-pictures-engine.md) | Landed #615 |
| S14-02 | Pictures in the app | 1 | The gallery slot and the Swift view-layer rename; Event and Place pages show Pictures, a header thumbnail and choosing a thumbnail; Events and Places lists lead with it (Promote wording dropped 2026-10-29) | — | S14-01, S13-23 | [S14-02-pictures-in-the-app.md](prs/S14-02-pictures-in-the-app.md) | Blocked — re-plan |
| S14-03 | Rectangle-only depiction drawing | 1 | Depiction drawing offers only the rectangle tool; the Citation's region as a depiction uses its bounding rectangle; the composer's drawing feedback, captions and cap announcement | — | S12-28 | — | Scoped |
| S14-04 | Promote wording for pictures | — | Folded into S14-02 (working rule "Batch 1 is cut by layer") | — | — | — | Descoped (2026-10-27) → S14-02 |
| S14-05 | Event and Place headers lead with a picture | — | Folded into S14-01 | — | — | — | Descoped (2026-10-27) → S14-01 |
| S14-06 | Choose an Event or Place thumbnail | — | Folded into S14-02 | — | — | — | Descoped (2026-10-27) → S14-02 |
| S14-07 | Events and Places lists lead with a picture | — | Folded into S14-02 | — | — | — | Descoped (2026-10-27) → S14-02 |
| S14-08 | Other-link support calibration | 2 | Raise `segment.OtherSupport` 0.25 → 0.75 so a thin name plus kinship clears Weak | — | — | [S14-08-other-support-calibration.md](prs/S14-08-other-support-calibration.md) | Landed #613 |
| S14-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

No briefs. S14-02 reuses the Person page's gallery section, header thumb, portrait controls and list thumbnail with new headings and wording; S14-03 shows fewer radios on an existing strip.

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Event and Place pictures (batch 1)
  S14-01  Pictures engine
  S14-02  Pictures in the app
  S14-03  Rectangle-only depiction drawing
  Check: bar outcomes 1–9 on a new project.

SLICE 2 — Matching calibration (batch 2)
  S14-08  Other-link support calibration
  Check: bar outcome 10 on the Jake + grandparent dogfood Source (or the calibration test's shape).

S14-99 — Dogfood close / docs
```

S14-01 and S14-03 can run in parallel; S14-02 follows S14-01. S14-03 waits on Spike 12's S12-28, which changes in place the one predicate that says whether the viewer can cut a file; S14-03 gates Draw on it (Spike 12 map review K-25). S14-08 is independent of batch 1.

### Dependencies at a glance

```text
S14-01 ──▶ S14-02
S13-23 ──▶ S14-02   (Spike 13: deletes the Promote files the old plan renamed)
S12-28 ──▶ S14-03
S14-08  (independent)
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Event and Place pictures

#### S14-01 — Pictures engine

**Ships.** The gallery-layer rename in Go and proto, with the Swift wire mapping (`GoStore`, `CatalogLikeness`) renamed to compile. `event_picture` and `place_picture` in `subjectvocab/registry.go` with `vocabkey` constants; the gallery reads the gallery keys, so `conclusiondetails.ForEntity` returns tiles for an Event or Place. `EventHeader` and `PlaceHeader` carry the lead image from the lead rule (Go, proto, `GoStore`), as `PersonHeader` does, and a chosen thumbnail refreshes the cached header. The gallery reads each picture's image File from its own `depiction_values` row, not the copy in `auto_reconciler_values.value_image_file_id` (likeness audit SEC-4). `observationstest.SaveDepiction(…, key, …)`, with `SaveLikeness` a wrapper. Updates the docs in [design § Domain docs](event-place-pictures.md#domain-docs-the-prs-update).
**Not this PR.** Anything the app shows (S14-02); drawing (S14-03); renaming the preferred-pointer column.
**Check.** `TestSeedProductValues` cases for both keys and the gallery-key test; a gallery test where a Place's picture is in its gallery and a Person's likeness is unchanged; an Event's and a Place's header carry the earliest picture, then the chosen one after `SetPreferredLikeness`; a handler test for the detail's header path; a gallery test where an edited `auto_reconciler_values.value_image_file_id` does not change a tile's image; the app builds and its tests pass with the renamed mapping.
**Context.** Working rules "Event and Place each get their own picture Property", "The gallery layer is named gallery", "The gallery reads one list of gallery keys", "New seeds reach new catalogs only", "An Event or Place chooses its thumbnail"; [design § What changes](event-place-pictures.md#what-changes); `compose.go` `onePerson`, `conclusion_detail.go` `attachHeader`; [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-depiction-value`](../../../.agents/skills/add-depiction-value/SKILL.md), [`add-ffi-handler`](../../../.agents/skills/add-ffi-handler/SKILL.md).

#### S14-02 — Pictures in the app

**Ships.** The gallery slot, with the Swift view-layer rename (`ConclusionDetailLikeness`, `ConclusionDetailLikenessSection`, `LikenessTileView`, `CachedThumbnail.likenessCrop`). The Event and Place pages show Pictures, lead their header thumbnail with the picture, and take a `ConclusionPortraitModel` (set as thumbnail, chosen badge, use the earliest). `EventsListView` and `PlacesListView` rows lead with it as `PersonsListView.thumbnailRelPath` does. The Promote sheet's wording left this PR on 2026-10-29: the sheet is removed by Spike 13's S13-23. The composer's depiction translator comments and `depictionPerson` say Subject. Strings through `add-localized-string`. Properties list VoiceOver distinguishes two Properties that share the label Picture (include the machine key and/or bound kind labels in `rowAccessibilityLabel` — S14-01 ship UX-1).
**Not this PR.** Spike 13's redesigned Gallery section (S13-14).
**Check.** On a new project: a Place picture shows under Pictures and opens its record; choosing a second picture leads the header and the Places list with it and badges it; use the earliest goes back; the same on an Event; a Person's page and list are unchanged; VoiceOver on the Properties list can tell `event_picture` from `place_picture`. Swift tests beside `PersonDetailTests` and `PersonsListTests`.
**Context.** Working rules "The gallery is one slot on the detail body", "An Event or Place chooses its thumbnail", "A depiction is worded as a picture", "Batch 1 is cut by layer"; [design § What already works](event-place-pictures.md#what-already-works); `ConclusionDetailPage.swift`, `ConclusionDetailThumbnail`, `PersonDetailContent`, `PersonDetailView`, `ConclusionListRow.swift`, `PropertiesView.swift` `rowAccessibilityLabel`.

#### S14-03 — Rectangle-only depiction drawing

**Ships.** The viewer's region strip takes the tools to offer as `[ArtifactRegionTool]` composer state beside `viewerArmedRegionTool` (the viewer stays unaware of depictions); a depiction draw offers `.rectangle` alone, a Citation region all of `ArtifactRegionTool`. `ArtifactRegionDraft.boundingRectangle`, beside `.whole` and built from `ArtifactRegionGeometry.boundingRect(of:)` and `rectanglePoints`, is what "Use the citation's region" sets on the depiction. While a region or a depiction is drawn, Set page, the locator chip × buttons and Find are disabled and the draw hint hides; Draw and the region buttons are blocked on a file the viewer cannot cut, and a failed cut shows with Save disabled; the needs-page callout sets a PDF Citation's page (UX-5, UX-6, UX-7). The empty, needs-page and region-hidden captions name the region buttons, not only Draw, and an edit that saves no image drops "Image to save" (UX-12, UX-13). The freehand tool's point cap (`ArtifactRegionOverlay`) is announced to VoiceOver, not only beeped. Draw's block and the failed-cut state read one "can cut" predicate, `ArtifactRegionCrop.snapshot`, as S12-28 leaves it.
**Not this PR.** Removing shapes from Citation regions; Go-side rectangle enforcement; rewriting saved polygon depictions.
**Check.** Draw on a likeness shows one radio; a Citation region shows seven; Swift tests for the strip's tool list and the bounding-rectangle conversion; an old polygon depiction still opens and saves; bar outcome 9, with Swift tests for the disabled controls, the blocked Draw and the failed-cut state.
**Context.** Working rule "A depiction is drawn as a rectangle"; `ArtifactViewer.swift` `regionTools`, `CitationComposerModel+Depiction.swift` `startDepictionDrawing`, `depictionCitationRegion`, `useDepictionRegion`; the audit items in [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) (2026-10-18 likeness audit). If the plan runs past about ten steps, the composer feedback splits out as its own PR.

### Slice 2 — Matching calibration

#### S14-08 — Other-link support calibration

**Ships.** `segment.OtherSupport` 0.25 → 0.75 in `core/matchregistry/segment.go`; inventory and `promote-graph-alignment.md` quote the new number; `segmentcompare` AC-2.5 / AC-2.8 expects; a graphalign calibration test for Jake given-only + Carol grandparent clearing Weak. Patch VERSION.
**Not this PR.** Name-curve relaxation; Weak-bar change; `OtherCredit`; kinship-only tier; multi-Source corroboration.
**Check.** `TestPartialNameWithGrandparentClearsWeak`; `TestCompare` AC-2.5 / AC-2.8; dogfood Promote row for Jake reads Weak (or stronger), not Little evidence.
**Context.** Working rule "Other-link support is 0.75"; [`matching-model.md`](../../matching-model.md) `segment.OtherSupport`; [`add-match-rule`](../../../.agents/skills/add-match-rule/SKILL.md).
