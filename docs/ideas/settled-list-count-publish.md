# Settled-list count publish

**Status:** idea only, not roadmapped. Found after S12-16 (mendahu/provenencia#492), whose maintainability review flagged Properties and Metadata as possibly sharing the Sources count bug. Properties has a live bug with a different cause. Metadata does not.

## What S12-16 fixed for Sources

`SourcesListView` published the Sources count on `onAppear` and on `onChange(of: status) == .ready`. A stale-while-revalidate refetch keeps `status == .ready` (`QueryHandle.beginLoading(stale: true)` only sets `isFetching`), so:

- `onAppear` during a refetch publishes the old cached count;
- the reload goes `.ready → .ready`, so `onChange(of: status)` never fires and the right count is never published.

S12-16 added `SourcesModel.syncCatalogCountsIfLoaded(_:)` (publish only when `status == .ready && !isFetching`), called on `onAppear` and `onChange(of: revision)`. It also removed the post-write publish. The rule is in `add-workspace-place`: "Publish from a list only on load."

## Properties: live bug

`PropertiesModel.submitCreate` and `confirmPendingImpact` (delete) do not patch the cached snapshot. They call `session.apply(.createdProperty / .deletedProperty / .mutatedSubjectTypeProperties)` and then, synchronously, `syncCatalogCounts()`. Properties is on screen, so `apply` revalidates the `propertiesWorkspace` key (`isFetching = true`, old value kept). The publish therefore sends the **pre-write** count. The reload then lands `.ready → .ready`, `onChange(of: handle.status)` does not fire, and `onChange(of: handle.value)` only reconciles selection.

Result: after creating or deleting a property, the Properties sidebar badge is off by one until the researcher leaves Properties and comes back (`onAppear` then publishes the fresh value). `CatalogCounts.refreshAll()` does not count properties, so nothing else corrects it. `PropertiesModelTests` has no count assertion, which is why no test caught it.

## Metadata: no live bug

- Only `MetadataModel` creates or deletes fields. It patches `metadataFieldsList` (`patchFieldsList`) before `syncCatalogCounts()`, so the post-write publish is correct.
- `metadataFieldsList` revalidates when hidden (the default policy), and no write outside Metadata changes its row count. `mutatedSourceMetadata` and `updatedMetadataField` reload `usedBy` and labels, not the count.
- `SourceTypesModel.publishCounts()` publishes Metadata from the fields handle it reads, which the same rules keep correct.

The shape is still the one S12-16 replaced: any future writer outside the list, such as an inline "new field" or "new property" from a Source page or the citation composer, would bring the Sources bug back.

## Shared rule (shape)

One rule for every list that publishes a count: **publish only from a settled handle, on appear and on each `revision` bump.**

- `QueryHandle.isSettled`: `status == .ready && !isFetching`.
- A view modifier, for example `.onSettled(handle) { model.syncCatalogCounts() }`, that runs the action on `onAppear` and on `onChange(of: handle.revision)` when `handle.isSettled`. Sources, Properties and Metadata use it. `SourcesModel.syncCatalogCountsIfLoaded(_:)` goes away.
- Drop the post-write `syncCatalogCounts()` calls in `PropertiesModel` (create, delete) and `MetadataModel` (create, delete). A patch bumps `revision` but leaves a revalidation running, and the reload's bump publishes the right count. This fixes Properties without adding a patch, and it is the skill rule S12-16 wrote down.
- Tests: a `PropertiesModelTests` case (create, then `waitForQuery`, then drive the settled publish, then `counts.properties?.total == 2`) that fails before the change, and the Metadata and Sources equivalents through the shared helper. The modifier's `onChange` trigger has no unit test, as in S12-16.

Open question: a key marked stale while hidden (`CatalogHiddenPolicy.markStale`: Sources, Properties) is `.ready` and not fetching until the view's query starts the reload. So `onAppear` can briefly publish the cached count first (S12-16's UX-note; it corrects itself on the reload's `revision` bump). Moving the stale flag onto the handle (`QueryHandle.isStale`, set by `markStale`, cleared by `beginLoading`) would let `isSettled` exclude it and close that window. It is cheap, but it touches `WorkspaceSession.invalidatedKeys`, which `ensureQuery` also reads.

Size: one small macOS PR (QueryHandle, one modifier, three views, two models, tests, one skill line).
