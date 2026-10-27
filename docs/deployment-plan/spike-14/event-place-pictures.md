# Event and Place pictures

Design for Spike 14. PR order and status are in the [map](deployment-plan.md).

## Problem

Spike 10 made a face on a photo a `depiction` value and seeded one depiction Property, `likeness`, bound to `person` ([Spike 10 decisions](../archive/spike-10/README.md)). A photo of a church, a street, or a wedding party has nowhere to go: Event and Place have no depiction Property, and the gallery reads `likeness` only.

Separately, a depiction is drawn with the Artifact viewer's full region strip (rectangle, four L-shapes, circle, freeform). The stored image is the bounding box of the region either way ([depictions idea](../../ideas/depictions-and-likenesses.md), "Transparent polygon"), so every shape but the rectangle is a way to draw the same picture with more steps.

## Shape

### Two Properties, one label

| Key | Label | Value type | Cardinality | Bound to | Sort | Locked |
| --- | --- | --- | --- | --- | --- | --- |
| `likeness` | Likeness | `depiction` | multiple | `person` | 1 | yes (unchanged) |
| `event_picture` | Picture | `depiction` | multiple | `event` | 5 (after `end_date`) | yes |
| `place_picture` | Picture | `depiction` | multiple | `place` | 3 (after `end_date`) | yes |

Keys are unique per origin (`UNIQUE (key, origin)`), so the two pictures need their own keys; labels are not unique, so both read "Picture". Each kind owning its key follows `event_name` beside `name` (`core/database/subjectvocab/registry.go`); these are the first two seeded Properties to share a label. Each has its own description ("A picture of this event: …", "A picture of this place: …").

Why not one shared key: a Property's label and description live on the Property, not on the binding (`seedBinding` holds kind, key, sort order, lock). One key across kinds would force one label on all of them, and "Likeness" is the right word only for a face.

Why not a Property named `depiction`: it would share its name with the value type. `name` and `date` already do this, and it does no harm in code (keys are `vocabkey` constants, value types `propertymeta` constants), but "picture" is the plainer word and keeps "depiction" meaning the value type alone.

### What already works

- **Reconcile.** `autoreconcile` groups depictions by value type, not key (`core/autoreconcile/modules.go`), so a Place's pictures reconcile like a Person's likenesses.
- **Composer.** The composer offers any Property bound to the Subject's kind and treats a depiction by value type (`CitationComposerVocabulary.isDepiction`). Its depiction copy is built from the Property label and the Subject label (`depictionThumbLabel`: "%1$@ of %2$@"), so it already reads "Picture of St Mary's Church".
- **Detail wire.** `conclusiondetails.ForEntity` calls `likenesses.ForEntities` for every kind, and `attachLikenesses` sends the tiles for every kind (`api/ffi/handlers/conclusion_detail.go`). Only the Person header gets a lead path.
- **Detail chrome.** `ConclusionDetailBody.likenesses` defaults to `[]`; `ConclusionDetailPage` draws `ConclusionDetailLikenessSection` for any body that returns tiles. The section's wording, accessibility ids and fallback mark still say Person (S14-02 makes them the body's).
- **Promote.** `promotealign.Evidence` lists every bound Property per kind, so the evidence sheet will show Event and Place pictures. Its wording calls every depiction a likeness (`L10n.Promote.lineSentence`, `ValueFamily.likeness`); S14-04 words it by value type.

### What changes

- **Seed.** Two Properties and two bindings in `subjectvocab/registry.go`; `vocabkey` constants for the keys.
- **Gallery keys.** `likenesses.sqlGallery` filters `p.key = 'likeness'`. It reads one list of gallery keys instead, held in `core/vocabkey` beside `PropertyLikeness`, so the next depiction Property that belongs in a gallery is one entry.
- **One gallery slot on the body.** `ConclusionDetailBody.likenesses` becomes one `gallery` value (title, tile wording, fallback mark, tiles), built by one tile mapping all three bodies call, as `ConclusionDetailThumbnail` is one slot. "Likenesses" on a Person, "Pictures" on an Event or Place; a body with no portrait marks no tile as the portrait.
- **Rectangle-only drawing.** When the composer arms the viewer for a depiction, the region strip shows the rectangle tool alone. A Citation's own region keeps all seven tools.
- **The Citation's region as a depiction.** "Use the citation's region" on a Citation whose region is not a rectangle sets the depiction to that region's bounding rectangle, built from `ArtifactRegionGeometry.boundingRect(of:)`, the rule the crop already uses, so the image is the same as before.

### Data

No migration. New seeds reach new catalogs only (Spike 10 rule "New seeds are for new catalogs"). Existing depictions saved with a polygon keep their region; they load, display and save as they are, and become rectangles only when redrawn. Go keeps accepting any valid region (`locator.ValidateDepictionRegion`); the rectangle rule is the app's drawing tool, not a stored constraint.

## Domain docs the PRs update

- [`seeded-vocabulary.md`](../../seeded-vocabulary.md): the Property table and the binding block gain `event_picture` and `place_picture` (S14-01).
- [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §depiction: "`likeness` on `person` is the seeded depiction Property" names all three (S14-01).
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md): the gallery paragraph says a Person's, Event's, or Place's gallery; the preferred pointer stays Person-only (S14-01).
- [`conclusion-reconciliation.md`](../../conclusion-reconciliation.md): "a person's images form a gallery" covers Event and Place pictures (S14-01).
- [`add-depiction-value`](../../../.agents/skills/add-depiction-value/SKILL.md): its description, the gallery keys, the `SaveDepiction` test helper, and the Do-not line about the Person page (S14-01); depictions are drawn as rectangles (S14-03).

## Stays later

- **A preferred picture and a header thumb for an Event or Place.** The header keeps its mark (`eventNoImage`); the gallery is the only display. In Go the pointer is generic (`SetPreferredLikeness` checks gallery membership, not kind) so a later PR can reuse `likenesses.Lead`; the app is what keeps it to the Person.
- **Pictures on the Events and Places lists.**
- **The event picture in the Person Timeline.** Spike 13 reserves the slot (R-26); filling it is Spike 13's call once S14-01 lands.
- **Pictures as matching evidence for an Event or Place.** `subjectcompare` gives pictures no weight; the evidence sheet only lists them (S14-04).
- **Signature.** Still open in the [depictions idea](../../ideas/depictions-and-likenesses.md).
