# Properties page and terms — later

**Status:** idea, not scheduled. Binding order from S11-32 (D-22); term management from 2026-10-09; event-type scoping and Occupation seed from dogfood 2026-10-15 / 2026-10-25. Placement of a terms view (panel vs page) is undecided.

## Editing a type's property order

The order of a subject type's properties (`subject_type_properties.sort_order`) sets the conclusion page's fields. S11-32 added the engine side: `ReorderSubjectTypeProperties` takes a type and every bound property in the new order, audits the move, and drops the cached vocabulary. No screen calls it. The Properties page is property-centric: a type's bindings are toggles on a property (`PropertiesView.swift`), so there is no ordered list per type to drag.

**Shape:** On the Properties page, an ordered list of a type's bound properties, reorderable by drag, calling `ReorderSubjectTypeProperties` through one new `GenealogyStore` method. Existing catalogs keep their seeded order (likeness last on a person) until the researcher changes it here.

**Open questions:** Whether the list lives beside the type's toggles or on its own page; whether a locked binding can be dragged (the engine allows it); needs a Claude Design brief before a plan.

## Lock event properties to specific event types

Today a Property bound to `event` is offered on every event, whatever its `event_type`. When associating a Property with events there should be an option to restrict it so it is only offered for specific event types (e.g. `cause_of_death` only on `death`, `officiant` only on `marriage` / `baptism`).

**Wanted:** Per-event-type scoping of a Property's availability, with the citation composer's Property menu (and anywhere else the type's bindings drive a picker) honoring it.

**Open question on the reading:** "types" means *event_type* terms here (binding to subject types already exists). "lock" means *restrict*, not the Properties page's registry-held "locked" meaning — UI copy must avoid colliding.

**How properties reach event types today:**

- **Storage.** `subject_type_properties`: `(subject_type_id, property_id, sort_order)`. It keys on **subject type**, not on `event_type` term. `event_type` is a locked term-valued Property on `event`, so the kind of an event is one Observation among others.
- **Go.** A term-scoped binding would need a new column or join (`subject_type_property_terms`?) plus catalogmodel / migration work; `ListBindings` / the proto `Binding` would carry the allowed term ids.
- **UI.** Natural slot: a secondary picker under the `event` checkbox that lists `event_type` terms (all / selected).
- **Where they are offered.** `CitationComposerVocabulary.propertyOptions` filters by subject type, connect rules, and locator capabilities — not by the event's current `event_type` observation.
- **Other readers.** Conclusion detail and reconciliation list values by Properties actually observed; a scoped-out Property that already has values would still display. [`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.3 would need the new concept.

## Seed an Occupation event type

Horizon key listed in [`seeded-vocabulary.md`](../seeded-vocabulary.md) §3.4. Occupation is one of the most common facts a census, directory, or civil record gives, and the Properties page copy already steers researchers toward recording it as an event. The create-time seeded event types stop at vital events plus census / residence / migration / travel.

**Wanted:** Ship `occupation` as a product-seeded `event_type` term alongside `residence`. Add it in `subjectvocab` `seedTerms` next to `residence`. Title composition follows the generic subject-count rule. Install runs only at catalog create; existing projects would not pick it up unless a separate backfill is decided. Follow [`.agents/skills/add-seeded-vocabulary`](../../.agents/skills/add-seeded-vocabulary/SKILL.md).

## A place to manage terms

Term-valued Properties each hold a vocabulary of terms. The app has no view of them.

- **Adding** a term happens only mid-transcription: the Citation Composer's "add custom term" dialog (`CitationComposerModel.confirmCustomTerm`). It asks for a label and nothing else.
- **Editing, describing, or deleting** a term isn't possible in the app. Go and the FFI already support it (`propertyterms.Update` / `Delete`, with delete-impact and in-use checks), but no Swift code calls them.
- **Direction and inverses can't be set.** A relationship term that reads one way only is `directed` and can name its `inverse_key`, so Promote's matching reads both directions as one relationship ([`promote-graph-alignment.md`](../promote-graph-alignment.md)). A term the composer creates is always undirected with no inverse.

**Shape:** One place to see and manage every term, grouped by its Property:

- List a Property's terms with origin, label, description, and how many Observations use each.
- Add, rename, describe, and delete the researcher's own terms, with the delete-impact preview. Product and plugin terms are read-only (`propertyterms.ErrLocked`); `place_relationship_type` takes no new terms at all.
- For relationship-like Properties: mark a term as directed, and pick its inverse from the same Property (mutual).

The composer dialog stays as the quick path while transcribing. Terms made there can be refined here later.

**Open questions:**

1. **Placement:** a terms panel inside a term-valued Property on the Properties page, or a first-class Terms page.
2. **Which Properties offer direction and inverse:** only `relationship_type`, or any term Property. Carry `inverse_key` on the wire term and delete the Swift `SeededRelationshipTypeTerm.inverse` map (Citation Composer is the only reader after S13-23).
3. **Merging terms:** two researcher terms that mean the same. S11-07 catches duplicates on create/rename; existing duplicates: rename-only, or a merge that repoints Observations.
4. **Seeded vocabulary changes** stay in `subjectvocab`; this view never edits product terms.
5. **The duplicate-term error could offer the existing term** ("Use “godfather”") instead of only refusing (S11-07 ship log).
6. **Seeded terms are listed by hand in Swift** (`Seeded…Term` enums, `L10n.PropertyTerm`); carrying terms on the wire would retire the hand lists (S11-20 ship log).

**What it would touch:** Go `propertyterms.Create` / `Update` accept `directed` and `inverse_key`; FFI term messages carry both; Swift store methods, a term list and editor, the `propertyTerms` query key, and L10n.

## A failed bind after a Property create

When the bind after a Property create fails, the dialog stays open and Retry calls `createProperty` again, which fails on the duplicate key (`PropertiesModel.swift`). Retry binds the Property already made.
