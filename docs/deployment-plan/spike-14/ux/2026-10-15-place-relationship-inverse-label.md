# Place relationship rows use the forward label on both ends

- **Date:** 2026-10-15
- **Type:** Bug (labeling)
- **Where:** Citation composer / place observations list
- **Annoyance:** The observations list for a place shows one row per relationship, e.g. "part of". The same forward label appears on both ends of a directional relationship. With *Calgary is part of Alberta*, the Calgary row reads "part of Alberta" (correct) but the Alberta row reads "part of Calgary" (wrong).
- **Wanted:** The inverse side of a directional relationship uses an inverse label — the Alberta row reads "contains Calgary". Likely applies to the other directional place relationship (`succeeded_by` → "succeeded" / "preceded by") too.

## Code notes (observed, not fixed)

- `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` — `listedLine(from:to:rule:termID:termProperty:vocabulary:viewing:)` builds the per-subject row. The `relationship` case flips through `inverseTerm(of:…)` when `viewing == to.id`, and the `location` case picks "at" vs "site of" by end, but the `placeRelationship` case returns `term?.label` regardless of which end is being viewed.
- Same file — `inverseTerm` resolves through the static `inverseTermKey` map, which only lists `relationship_type` keys (`parent`/`child`, …). There is no entry for `part_of` / `succeeded_by`, so a fix likely needs either place entries there or a direct inverse-copy switch for the place case.
- `core/database/subjectvocab/registry.go` — the `place_relationship_type` seed terms are `Directed: true` but carry no `Inverse`, and `place_relationship_type` is product-only, so an inverse term row would not exist to look up. The inverse copy ("contains", "preceded by") would have to come from L10n rather than vocabulary.
- `macos/App/Platform/L10n.swift` — `CitationComposer.isPartOf` / `wasSucceededBy` and `Conclusions.placePartOf` / `placeSucceededBy` are the existing place copy; the Place conclusion page already distinguishes `part_of` vs `contains` (`CatalogPlaceRelationship` kinds), so there is precedent for the inverse wording.
- `macos/App/Features/CitationComposer/CitationConnections.swift` — `groupSaved` keeps `fromSubjectID` / `toSubjectID` from the bridge endpoints, so direction is available when rendering.
