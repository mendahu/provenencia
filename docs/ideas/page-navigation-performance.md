# Page navigation performance

**Status:** idea only — not roadmapped.

Evaluate performance and visual flashing when navigating to deep workspace places (Sources, Source fields, Source types, and similar). Revisit after current navigation-history work lands; approach TBD.

## Known symptoms

- **Sources Back/Forward → detail:** Brief flash of the source **list** before `SourcePageView` appears, even though list and detail are separate views. History updates `navigation.currentLocation` immediately, but `SourcesView` gates on `model.openedSourceID`, which is set in `onChange` inside a `Task` — one run-loop late.
- **Vocabulary (fields/types):** Master–detail can flash empty detail before selection applies on restore (partially addressed via `load(from:)`; may still flash while catalog loads or on in-place navigation).
- **Section remounts:** Switching sidebar destinations remounts destination views and reruns `.task { load }`.

## Low-hanging fruit (Sources)

- Reconcile **synchronously** on `onChange(of: navigation.currentLocation)` for Sources (no `Task` when `apply(from:)` is sync), **or** drive list vs detail from `navigation.currentLocation.sourceId` after initial load while keeping the model in sync.
- Does not fix first-open load latency or `SourcePageView`’s own load; targeted at Back/Forward within Sources.
