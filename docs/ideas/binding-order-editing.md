# Editing a type's property order

**Status:** idea only — not roadmapped. Found in S11-32 (D-22).

## What happens

The order of a subject type's properties (`subject_type_properties.sort_order`) sets the conclusion page's fields. S11-32 added the engine side: `ReorderSubjectTypeProperties` takes a type and every bound property in the new order, audits the move, and drops the cached vocabulary. No screen calls it. The Properties page is property-centric: a type's bindings are toggles on a property (`PropertiesView.swift`), so there is no ordered list per type to drag.

## Suggested shape

- On the Properties page, an ordered list of a type's bound properties, reorderable by drag, calling `ReorderSubjectTypeProperties` through one new `GenealogyStore` method.
- Existing catalogs keep their seeded order (likeness last on a person) until the researcher changes it here.

## Open questions

- Whether the list lives beside the type's toggles or on its own page.
- Whether a locked binding can be dragged (the engine allows it).
- It is new chrome, so it needs a Claude Design brief before a plan.
