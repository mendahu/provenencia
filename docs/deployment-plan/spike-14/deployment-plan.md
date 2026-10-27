# Deployment plan — Spike 14

Spike 14 is an odds-and-ends spike (working rule "Even spikes are odds and ends"). It takes small features, UX refinements and fixes in batches; each batch is a slice with its own Check. The first batch pulls the "other keys" item and the transparent-polygon item from [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md), and adds rectangle-only depiction drawing. One note, one PR, unless two edit the same part of one view or the same comparer.

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a run of small, independent improvements, each batch landing on its own.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project. A batch adds its outcomes here when it is cut.

**Batch 1 — Event and Place pictures** (new catalogs; working rule "New seeds reach new catalogs only"):

1. **A place has pictures.** On a Citation of a photo of St Mary's Church, the composer offers Picture for the Place Subject; the saved line reads "Picture of St Mary's Church".
2. **An event has pictures.** The same for a wedding photo filed on the Event.
3. **The pages show them.** The Place and Event pages show a Pictures section with the crop, captioned with its Source; a tile opens its record, as a likeness tile does on the Person page.
4. **A depiction is a rectangle.** Drawing a likeness or a picture shows the rectangle tool alone; a Citation's own region still offers all seven tools. "Use the citation's region" on a circled entry sets the depiction to the circle's bounding rectangle.
5. **Promote says picture.** Comparing two Places with pictures, the evidence sheet reads "Same picture" or "Compare these pictures", not "likeness".
6. **Likenesses are unchanged.** A Person's likeness, gallery and portrait work as before.

## Where the notes went

| Dogfood note or idea | PR | Brief |
| --- | --- | --- |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Other keys of the same value type" (place and event) | S14-01, S14-02 | — |
| Rectangle-only depiction drawing (user, 2026-10-27) | S14-03 | — |
| Promote wording for a non-Person depiction (map review K-4) | S14-04 | — |
| [`depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) "Transparent polygon" | Refused (working rule "A depiction is drawn as a rectangle") | — |

**Back in the ideas folder** — [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md): signature, a preferred picture and header thumb for an Event or Place, pictures on the Events and Places lists, and the Spike 10 audit leftovers.

## Working rules

These are the calls for this spike. PR plans treat them as binding. A rule names its batch.

- **Even spikes are odds and ends.** From Spike 13, odd-numbered spikes are large new features and even-numbered spikes collect small features, UX refinements and fixes. An odds-and-ends spike adds work in batches: each batch is a slice with its own Check and bar outcomes, cut with [`plan-spike`](../../../.agents/skills/plan-spike/SKILL.md) when it is pulled in.
- **(Batch 1) Event and Place each get their own picture Property, labelled "Picture".** `event_picture` bound to `event` and `place_picture` bound to `place`, value type `depiction`, cardinality `multiple`, locked, each with its own description. Keys are unique per origin and a label lives on the Property, not the binding, so one shared key would force one label on every kind. `likeness` stays the Person key. No Property is named `depiction`. [Design](event-place-pictures.md#two-properties-one-label). Binds S14-01.
- **(Batch 1) The gallery reads one list of gallery keys.** `likenesses.ForEntities` filters on a list in `core/vocabkey` (`likeness`, `event_picture`, `place_picture`), not on `PropertyLikeness` alone. The next depiction Property that belongs in a gallery is one entry. Binds S14-01.
- **(Batch 1) New seeds reach new catalogs only.** As in Spike 10: the two Properties come from Install, and no migration adds them to an existing catalog. Binds S14-01.
- **(Batch 1) No preferred picture for an Event or Place.** The gallery is the only display. The header keeps its mark and `ConclusionDetailPage`'s portrait stays Person-only. In Go, `preferred_likeness_observation_id` and `SetPreferredLikeness` stay generic (gallery membership, not kind), so the app is what limits the portrait; S14-01 says so in `conclusion-layer-data-model.md`. A body with no portrait marks no tile as the portrait. Binds S14-01, S14-02.
- **(Batch 1) The gallery is one slot on the detail body.** `ConclusionDetailBody` carries one gallery value (title, tile wording, fallback mark, tiles) built by one tile mapping, in place of `likenesses` and the Person-only strings, ids and mark in `ConclusionDetailLikenessSection`. Precedent: `ConclusionDetailThumbnail`. Binds S14-02.
- **(Batch 1) A depiction is drawn as a rectangle.** The composer arms the viewer with the rectangle tool alone for a depiction; a Citation's own region keeps all seven tools. "Use the citation's region" takes a non-rectangle region's bounding rectangle. Go keeps accepting any valid depiction region, so a depiction saved as a polygon loads, displays and saves as it is until redrawn. The transparent-polygon crop is refused. Binds S14-03.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S14-01 | Picture Properties for Event and Place | 1 | Seed `event_picture` and `place_picture` with bindings; the gallery reads the gallery keys; domain docs and `add-depiction-value` | — | — | — | Scoped |
| S14-02 | Event and Place pictures sections | 1 | The Event and Place pages show their gallery under a "Pictures" heading | — | S14-01 | — | Scoped |
| S14-03 | Rectangle-only depiction drawing | 1 | Depiction drawing offers only the rectangle tool; the Citation's region as a depiction uses its bounding rectangle | — | — | — | Scoped |
| S14-04 | Promote wording for pictures | 1 | The evidence sheet words a depiction line by value type ("picture"), with the line's kind for its fallback mark | — | S14-01 | — | Scoped |
| S14-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

No briefs. S14-02 reuses the gallery section the Person page draws today, with a new heading; S14-03 shows fewer radios on an existing strip.

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Event and Place pictures (batch 1)
  S14-01  Picture Properties for Event and Place
  S14-02  Event and Place pictures sections
  S14-03  Rectangle-only depiction drawing
  S14-04  Promote wording for pictures
  Check: bar outcomes 1–6 on a new project.

S14-99 — Dogfood close / docs
```

S14-01 and S14-03 can run in parallel; S14-02 and S14-04 follow S14-01 and can run in parallel.

### Dependencies at a glance

```text
S14-01 ──▶ S14-02
S14-01 ──▶ S14-04
S14-03
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Event and Place pictures

#### S14-01 — Picture Properties for Event and Place

**Ships.** `event_picture` and `place_picture` in `subjectvocab/registry.go` (Properties and bindings) with `vocabkey` constants; `likenesses.ForEntities` reads the gallery keys, so `conclusiondetails.ForEntity` returns tiles for an Event or Place. A `subjectvocab` test holds every gallery key to a seeded `depiction` Property. `observationstest.SaveDepiction(…, key, …)` takes the key; `SaveLikeness` wraps it with `vocabkey.PropertyLikeness`. Updates the docs in [design § Domain docs](event-place-pictures.md#domain-docs-the-prs-update).
**Not this PR.** The pages (S14-02); drawing (S14-03); a preferred picture or header thumb (ideas).
**Check.** `TestSeedProductValues` cases for both keys; a `likenesses` test where a Place's picture is in its gallery and a Person's likeness is unchanged; a Place picture saved through the real write path (an `observationstest` helper beside `SaveLikeness`) reaches the Place's `conclusiondetails` tiles.
**Context.** Working rules "Event and Place each get their own picture Property", "The gallery reads one list of gallery keys", "New seeds reach new catalogs only"; [design § What changes](event-place-pictures.md#what-changes); [`add-seeded-vocabulary`](../../../.agents/skills/add-seeded-vocabulary/SKILL.md), [`add-depiction-value`](../../../.agents/skills/add-depiction-value/SKILL.md).

#### S14-02 — Event and Place pictures sections

**Ships.** The gallery slot (working rule "The gallery is one slot on the detail body"): one tile mapping for all three bodies; the section's heading, tile label, help, accessibility ids and fallback mark come from the slot ("Likenesses" on a Person, "Pictures" on an Event or Place; `add-localized-string`). The composer's depiction translator comments and `depictionPerson` name say Subject, not person.
**Not this PR.** A portrait or header thumb (working rule "No preferred picture"); Spike 13's redesigned Gallery section (S13-14), which is the Person page's.
**Check.** On a new project, a saved Place picture shows on the Place page under Pictures and opens its record; an Event likewise; Swift tests on both bodies.
**Context.** Working rules "No preferred picture for an Event or Place", "The gallery is one slot on the detail body"; [design § What already works](event-place-pictures.md#what-already-works); `ConclusionDetailPage.swift` `ConclusionDetailLikenessSection`, `ConclusionDetailThumbnail`; `L10n.swift` `depictionThumbLabel`.

#### S14-03 — Rectangle-only depiction drawing

**Ships.** The viewer's region strip takes the tools to offer as `[ArtifactRegionTool]` composer state beside `viewerArmedRegionTool` (the viewer stays unaware of depictions); a depiction draw offers `.rectangle` alone, a Citation region all of `ArtifactRegionTool`. `ArtifactRegionDraft.boundingRectangle`, beside `.whole` and built from `ArtifactRegionGeometry.boundingRect(of:)` and `rectanglePoints`, is what "Use the citation's region" sets on the depiction.
**Not this PR.** Removing shapes from Citation regions; Go-side rectangle enforcement; rewriting saved polygon depictions.
**Check.** Draw on a likeness shows one radio; a Citation region shows seven; Swift tests for the strip's tool list and the bounding-rectangle conversion; an old polygon depiction still opens and saves.
**Context.** Working rule "A depiction is drawn as a rectangle"; `ArtifactViewer.swift` `regionTools`, `CitationComposerModel+Depiction.swift` `startDepictionDrawing`, `depictionCitationRegion`, `useDepictionRegion`.

#### S14-04 — Promote wording for pictures

**Ships.** `L10n.Promote.lineSentence` words a `.depiction` value as a picture ("Same picture", "Compare these pictures"), with `ValueFamily.likeness` renamed `.depiction`; the spoken likeness lines and `EvidenceCrops`' fallback mark take the line's kind.
**Not this PR.** Pictures as matching evidence for an Event or Place ([design § Stays later](event-place-pictures.md#stays-later)).
**Check.** Two Places with the same picture read "Same picture" on the evidence sheet; a Person's likeness line reads "picture" too (one wording per value type); Swift tests in `PromoteEvidenceDisplayTests`.
**Context.** [design § What already works](event-place-pictures.md#what-already-works) (Promote); `L10n.swift` `lineSentence`; `PromoteEvidenceCards.swift` `EvidenceCrops`.
