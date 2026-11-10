# S16-09 — Conclusion pages refresh from the Sources they cite and show their last value while they refresh

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 1b — Current data and honest notices
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a page shown stale draws the header's existing "Refreshing" line and its existing content; no view gains or loses a part)
**Board:** —
**Depends on:** S16-04 (landed [#776](https://github.com/mendahu/provenencia/pull/776))
**Blocks:** S16-25, S16-29; Spike 15's S15-29 (restates what the detail key follows)
**Branch:** `s16-09-source-follow-and-stale-pages`
**Version bump:** patch — Vitals, the Gallery and the viewer show a renamed Source's words and Back shows a page at once; a behavior fix with no new screen, RPC or schema ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-6; [cache design](../cache-architecture.md) "Stale while refreshing"; working rules: "(Batch 1) A cached detail follows the Sources its records cite", "(Batch 2) Swift caches one value per view, follows its inputs and shows it stale while refreshing", "(Batch 2) The point of this batch is simpler code" (binds the stale half), "(Batch 1) A refresh failure is read beside its value" (S16-04's reading); [`macos-client-patterns.md`](../../../macos-client-patterns.md) rule 2 and "Conclusion keys"
**Planned:** 2026-11-10 against `main` @ `cd03582`
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

Two things a researcher sees on a Person, Event or Place page get better. First, renaming a Source, or fixing a Citation's label or page or an Artifact's label, now reaches the Vitals provenance sheet, the Gallery captions and the likeness viewer's record row on a page that cites it, as it already reaches the Sources section (audit P-6). Second, going Back to a page visited earlier shows what it showed at once, with "Refreshing" while it reloads, instead of a spinner: a write no longer drops the cached pages that are off screen, it marks them stale. Only an entity a write deleted is dropped, and at most 50 hidden entities keep their pages.

**Why now.** S16-04 landed the reading that turns a not-found reload behind a kept value into "gone", which is what lets a stale page of a deleted or merged entity leave instead of showing old rows. The Source follow and the stale policy edit the same descriptors, so they ship together (S16-23 merged here, user, 2026-11-08).

**What stays.** What a visible page does on a write (it reloads, as today); the lists' policies; S16-04's notices and gone line; which handles each key follows (only `.source` is added to the detail key). The claims and record links keys do not follow `.source` here (S16-12).

## 2. User stories

**US-1.** As a researcher, when I rename a Source or fix a Citation or Artifact label, I see the new words in the Vitals provenance sheet, the Gallery captions and the viewer's record row of every page that cites it, so the page never contradicts its own Sources section.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A warmed detail whose name field has an outcome with `sourceID` "s-field", a derived birth row whose owner's field has an outcome with `sourceID` "s-owner", and a picture whose record has `sourceID` "s-pic"; a write reports `sourceScopes` ["s-field"], then ["s-owner"], then ["s-pic"] | `revalidatedKeys` names the detail key for each of the three |
| AC-1.2 | The same detail; a write reports `sourceScopes` ["s-other"] | `revalidatedKeys` does not name the detail key |
| AC-1.3 | The detail key is visible (its page is on screen); a write reports `sourceScopes` ["s-pic"] | The key reloads: its handle `isFetching` is true, and after `settleWrites` it holds the new value |
| AC-1.4 | The detail key is hidden; a write reports `sourceScopes` ["s-pic"] | The key keeps its value and `isStale` is true; nothing loads |

**US-2.** As a researcher, when I go Back to a Person, Event or Place page I visited, I see its last value at once with "Refreshing" while it reloads, instead of a spinner.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Person e1's page (Vitals) and its Family section were visited, then e3's page is shown; a write reports `handles` ["e1"] | e1's detail and family keys keep their values with `isStale` true; neither loads |
| AC-2.2 | AC-2.1's state; the researcher goes back to e1 (`apply(location:)` e1) | e1's detail handle has its old value and `isFetching` true (the header's "Refreshing"); after the load it holds the new value, `isStale` false |
| AC-2.3 | Each of the five per-entity kinds (`conclusionDetail`, `personClaims`, `personFamily`, `personEvents`, `personRecordLinks`) | `registry.hiddenPolicy` is `.markStale` |
| AC-2.4 | A Property relabel (`lists: [.propertyVocabulary]`) while a Person page is hidden | That page's detail key keeps its value with `isStale` true |

**US-3.** As a researcher, when a person I visited is deleted or merged elsewhere, going Back does not show their old page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | e1's detail, family and claims keys are warmed and hidden; a write reports `deleted` [handle: ["e1"]] | All three keys are evicted (`queryHandle` nil); e3's keys are untouched |
| AC-3.2 | e1's detail key is visible; a write reports `deleted` [handle: ["e1"]] without `leavingPage` | The key is not evicted; it revalidates, and its reload reads gone (S16-04) |
| AC-3.3 | e1's detail key is hidden and stale (merged: its id is in `handles`, not in `deleted`); the store now answers `conclusiondetails.not_found`; the researcher goes back to e1 | The handle's `reading` is `.loaded(oldValue, refreshFailure: not_found)` and `ConclusionDetailPresentation` built from it is `.gone`, so the page leaves |
| AC-3.4 | Unpromote of a person's last record on their own page (`apply(_:leavingPage:)`) | Their detail key is evicted and the Persons list loads once (today's `applyLeavingPageLoadsTheNextListOnce` and `UnpromoteFlowTests` keep passing) |

**US-4.** As the app, I keep at most 50 hidden entities' pages, so a long session does not grow without bound.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A detail key for "never" set with no place showing it; then e0 (on its Family section) through e50 are shown in that order | 51 entities are hidden after e50 is shown with e50 visible: "never" is evicted (an entity no place showed counts as least recently shown), e0's detail and family keys are kept |
| AC-4.2 | AC-4.1, then e51 is shown | e0's detail and family keys are evicted together; e1 – e50 are kept |
| AC-4.3 | AC-4.2, then e1 is shown again, then e52 | e2 is evicted, not e1: the order is by last shown, not first shown |
| AC-4.4 | Any `apply(location:)` | The keys of the entity on screen are never evicted by the cap |

**US-5.** As a contributor, I find no unused stale policy beside the hidden policy.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `rg CatalogQueryStalePolicy macos` and `rg "stalePolicy" macos` | No hit |
| AC-5.2 | `rg "\.evict\b" macos/App` | No hit: `CatalogHiddenPolicy` has `revalidate` and `markStale` only |

## 3. Scope

| In | Covered by |
| --- | --- |
| `.conclusionDetail` follows `.source` for the Sources its field outcomes, derived-row owners' outcomes and pictures cite (P-6) | STEP-2, AC-1.1 – AC-1.4 |
| The five per-entity Conclusion kinds go from `whenHidden: .evict` to `.markStale` | STEP-4, AC-2.1 – AC-2.4 |
| `apply` evicts the hidden keys of the entities in `effects.deleted` (handles) | STEP-3, AC-3.1, AC-3.2, AC-3.4 |
| A not-found reload over a stale value reads as gone, through S16-04's reading (no new code; a test) | STEP-4, AC-3.3 |
| At most 50 hidden entities keep per-entity keys, least recently shown evicted first | STEP-5, AC-4.1 – AC-4.4 |
| `CatalogHiddenPolicy.evict` goes (no descriptor uses it after STEP-4) | STEP-4, AC-5.2 |
| `CatalogQueryKey.entityId` *(new)*, read by the cap, the deleted eviction and the descriptors' own-handle follows, in place of three `guard case` closures | STEP-1 |
| `CatalogQueryStalePolicy` goes: the file, the descriptor field and `make` parameter, every arm's `stalePolicy:`, `CatalogQueryRegistry.stalePolicy(for:)`, the five test assertions | STEP-6, AC-5.1 |
| Docs: `add-workspace-place` (slot list gains `.source`; `whenHidden` values; per-entity keys), `macos-client-patterns.md` rule 2 and "Conclusion keys" (S9-15 exception goes), the map's Batch 1 rule clause on hidden details | STEP-7 |

| Out | Home |
| --- | --- |
| Claims and record links keys following `.source` | S16-12 |
| Narrower refreshes: one `.followsInputs` refresh, keys as plain values | S16-25 |
| A narrower Source effect for renames only | not planned (the audit scope is what the cache can follow; rule "A cached detail follows the Sources its records cite") |
| Prefetching a hovered person's page or other sections | S16-29 |
| A page in another window learning of a write made in this one | not planned (one `WorkspaceSession` per workspace window; no change here) |
| Back restoring a section's scroll, folds and focus (P-19) | [person-page-followups](../../../ideas/person-page-followups.md) |
| Memoising the cited-Source set per value | not planned (rule "No new persisted projections, and no stored answer without a measurement"; S16-24 sends it on the wire) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Hidden policy | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `CatalogHiddenPolicy` | Three cases: `revalidate`, `markStale`, `evict`. Only the three Conclusion descriptors below use `.evict` |
| Descriptor | same file, `CatalogQueryDescriptor` | Holds `stalePolicy: CatalogQueryStalePolicy` (every arm passes `.sessionFresh`; nothing reads it but `stalePolicy(for:)` and tests), `whenHidden`, `refreshes` |
| Per-person helpers | same file, `CatalogQueryDescriptor.personSection` and `personGraphsSection` | Both set `whenHidden: .evict`. `personGraphsSection` takes an `entityId:` closure that each caller (`personClaims`, `personRecordLinks`) writes as a `guard case` on its key |
| Detail descriptor | same file, `case .conclusionDetail` arm | `whenHidden: .evict`; refreshes `.applyHiddenPolicy(follows: .touchedHandles)` on its own id (a `guard case` closure) and `.followsPropertyLabels`. Follows no Source (P-6) |
| Stale policy | `macos/App/Features/Workspace/Session/CatalogQueryStalePolicy.swift` `CatalogQueryStalePolicy` | `sessionFresh`, `refetchOnRevisit` (never set). `CatalogQueryRegistry.stalePolicy(for:)` reads it |
| Key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `CatalogQueryKey` | Five per-entity cases carry `entityId`; no accessor returns it. Their doc comments say "Evicted, not revalidated, while off screen" |
| Session apply | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)` | Patches, then applies the hidden policy once per key named by the mutation, the refreshes, an awaiting patch or a `.run` fallback. Ignores `effects.deleted` |
| Session apply, leaving | same file, `apply(_:_:leavingPage:)` | Clears `visibleKeys`, so every key is hidden: the leaving page's `.evict` keys are dropped and no list loads; then runs `leave` |
| Session navigation | same file, `apply(location:placeRegistry:)` | Sets `visibleKeys` from the place's keys and warms each. Keeps no record of what was shown before |
| Policy switch | same file, `applyHiddenPolicy(to:)`, private `evict(_:)` | `.evict` arm calls `evict`, which cancels the load and forgets the handle |
| Effects | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects.deleted`, `deletedHandles` | Carry the `canonical_entity` ids a write deleted. The doc comment says "`apply` ignores it" |
| Follows | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryFollows.source`, `CatalogQueryRefresh.applyHiddenPolicy(followsValue:ids:)` | `.source` reads `effects.sourceScopes`; only the Sources list's `lastUpdated` patch uses it. `followsValue` matches any id until the key holds a value |
| Detail value | `macos/App/Platform/GenealogyStore.swift` `CatalogConclusionDetail`, `CatalogConclusionField.outcomes`, `CatalogDerivedRow.owners`, `CatalogOwnerField.field`, `CatalogPicture.record`, `CatalogReconcilerOutcome.sourceID` | Every record the page words carries its `sourceID` |
| Gone | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation.init(_:isFetching:content:)` | A `.loaded` reading whose `refreshFailure` is `conclusiondetails.not_found` is `.gone`; the page leaves (`leavesForSectionRoot`). `ConclusionPageSectionHost` reads sections the same way |
| Page handle | same file, `ConclusionDetailPage.body` | Reads `session.queryHandle(key)` and queries once in `.task`; a handle evicted while the page is on screen leaves it on the loading page |
| FakeStore | `macos/App/Platform/FakeStore.swift` `getConclusionDetail`, `seededConclusionDetails` | Returns a seeded detail, else a filed person's, else throws `conclusiondetails.not_found` |
| Tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`, `touchedHandlesRevalidateOnlyTheirDetailPages`, `aKeyInvalidatedAndRefreshedRevalidatesOnce`, `personClaimsFollowHandlesAndGraphs`, `applyLeavingPageLoadsTheNextListOnce` | The first three expect a hidden key to be evicted; the fourth says "the detail page follows only the handle"; the fifth expects the deleted page's key evicted through `leavingPage` |
| Tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `hiddenPolicies`, `personClaimsRevalidateOnTouchedHandle`, `personFamilyRevalidateOnTouchedHandle`, `personEventsRevalidateOnTouchedHandle`, `personRecordLinksRevalidate`, `sessionUsesStandardRegistryByDefault` | Five `.evict` assertions; five `stalePolicy` assertions |
| Tests | `macos/ProvenenciaTests/PropertiesModelTests.swift` `renamingAPropertyStalesConclusionPages` | Expects the hidden detail key nil |
| Tests | `macos/ProvenenciaTests/UnpromoteFlowTests.swift` (the confirm test, "leaving there evicts her page") | Expects the deleted person's detail key nil after `apply(_:leavingPage:)` |

**Precedent.** The `.evidenceGraph` and `.sourceWorkspace` arms in `CatalogQueryRegistry.swift` are `whenHidden: .markStale` and show their cached frame while they reload; copy their policy. `personGraphsSection`'s `.applyHiddenPolicy(followsValue: .evidenceGraphs, ids: graphIDs)` is the shape of the Source follow: ids read from the cached value. Do not copy `personGraphsSection`'s `entityId:` closure parameter; D-6 replaces it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The `.conclusionDetail` arm adds `.applyHiddenPolicy(followsValue: .source, ids:)` whose closure takes `CatalogConclusionDetail` and returns the `Set` of `sourceID`s of `fields[].outcomes[]`, `derivedRows[].owners[].field.outcomes[]` and `pictures[].record`, written inline in the arm as `personClaims`' `graphIDs` closure is | P-6's fix; the rule names exactly these three; inline keeps it beside the key it serves until S16-25 replaces it | A computed property on `CatalogConclusionDetail`: a new public name that S16-24/S16-25 delete; following every `sourceScopes` id with `applyHiddenPolicyOnAny`: reloads every visible detail on every Source write |
| D-2 | `conclusionDetail`, and through `personSection` and `personGraphsSection` the four Person section kinds, are `whenHidden: .markStale` | Map rule "Swift caches one value per view, follows its inputs and shows it stale while refreshing"; Back shows the cached value with `isFetching` through `ensureQuery`'s `beginLoading(keepingValue: true)`, which the header already words as "Refreshing" | Revalidating hidden pages: reloads every visited page on every write (the reason for S9-15) |
| D-3 | `CatalogHiddenPolicy.evict` is deleted with its `applyHiddenPolicy(to:)` arm. The private `evict(_:)` stays; D-4 and D-5 call it | No descriptor uses `.evict` after D-2; a policy no one sets is the dead-code shape this PR removes for `CatalogQueryStalePolicy` | Keeping `.evict` for a future kind: none is named in the map |
| D-4 | `apply(_:_:)` first evicts, through a private `evictEntities(_:)` *(new)*, every warmed key of this project whose `entityId` is in `effects.deletedHandles` and that is not in `visibleKeys`, then runs as today. A visible key of a deleted entity is left to the normal pass, which revalidates it and reads gone | A visible page holds no handle of its own across an eviction (`ConclusionDetailPage.body` reads `queryHandle` and queries once in `.task`), so evicting it would leave a spinner; `apply(_:_:leavingPage:)` clears `visibleKeys` first, so the leaving page's keys are evicted as today | Evicting visible keys too: the spinner above; following `effects.deleted` through a descriptor `refreshes` entry: a refresh re-applies the hidden policy, which no longer evicts |
| D-5 | `WorkspaceSession` keeps `private var lastShown: [String: Int]` *(new)* and `private var showCount = 0` *(new)*. `apply(location:)`, after setting `visibleKeys`, gives each `entityId` among the place's keys the next `showCount`, then, when the warmed keys of this project hold more than `WorkspaceSession.maxHiddenEntities` *(new, `static let` 50, `// decided: user, 2026-11-08`)* distinct entity ids not on screen, sorts those ids by `(lastShown[id] ?? 0, id)` and calls `evictEntities` on the first (count − 50). `evictEntities` also removes its ids from `lastShown`. The cap runs only in `apply(location:)` | The map rule's cap of 50, least recently shown first. Counting entities, not keys, keeps a person's detail and sections together. An entity no place showed (a test's or preview's direct `query`) sorts first, so a later prefetch (S16-29) cannot escape the cap. Only navigation adds an entity in production | An ordered array moved on each visit: the same result with an O(n) move; a cap on keys: splits a person's sections from their page; trimming in `ensureQuery`: runs inside `apply`'s refresh loop |
| D-6 | `CatalogQueryKey.entityId: String?` *(new)* returns the `entityId` of the five per-entity cases and nil for the rest. `personGraphsSection` loses its `entityId:` parameter and follows `key.entityId`; the `.conclusionDetail` arm's own-handle follow uses it | D-4 and D-5 need the id; three `guard case` closures go. S16-25 turns it into the key's ids | A descriptor flag "per entity": the id still has to come from the key |
| D-7 | No code for "gone": a stale value's reload that throws `conclusiondetails.not_found` is already `.loaded(value, refreshFailure:)`, which `ConclusionDetailPresentation` and `ConclusionPageSectionHost` read as gone (S16-04). A test pins it (T-8) | The map scope says S16-04's reading covers it | A gone flag on the session |
| D-8 | `CatalogQueryStalePolicy` goes: `CatalogQueryStalePolicy.swift` is deleted (the `App` group is file-system synchronized; no `project.pbxproj` edit), `CatalogQueryDescriptor.stalePolicy` and `make(stalePolicy:)` go, each arm drops `stalePolicy: .sessionFresh`, `CatalogQueryRegistry.stalePolicy(for:)` goes. In `CatalogQueryRegistryTests` the four section tests drop their `stalePolicy` line and `sessionUsesStandardRegistryByDefault` asserts `session.registry.hiddenPolicy(.sourcesList(project:)) == .markStale` | Map scope; no reader | Keeping `sessionFresh` as documentation |
| D-9 | `apply(_:_:leavingPage:)` keeps its code; its doc comment says the deleted page's keys are evicted by `apply` because they are hidden and named in `deleted`, and that clearing `visibleKeys` keeps the lists from loading | Behavior is the same; the comment's reason changed | Removing the `visibleKeys = []` line: the next list would load twice (`applyLeavingPageLoadsTheNextListOnce`) |
| D-10 | `CatalogWriteEffects.deleted`'s doc comment says `apply` evicts the hidden keys of the deleted handles | It no longer ignores them | — |
| D-11 | Version 0.7.1 → 0.7.2 (or the next patch on `main`) | A visible behavior fix ([`docs/versioning.md`](../../../versioning.md)) | Minor: no new screen, RPC or schema |
| D-12 | The map's Batch 1 rule "A cached detail follows the Sources its records cite" says a hidden detail is marked stale, not "evicted, as today" | The same PR makes both halves; the Batch 2 rule (merge of S16-23, user, 2026-11-08) sets the policy | — |

**Rules this plan bends.** The Batch 1 rule "A cached detail follows the Sources its records cite" says "a hidden detail is evicted, as today" → the Batch 2 rule "Swift caches one value per view, follows its inputs and shows it stale while refreshing" sets `.markStale` for the same key in the same PR → the user merged S16-23 into S16-09 (2026-11-08), so the Batch 2 policy holds → STEP-7 rewrites the clause (D-12).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogQueryKey.entityId` *(new)* | property | `CatalogQueryKey.project`, `personGraphsSection`'s `entityId:` closure, `WorkspaceLocation.entityId` | Extends the key the way `project` does; replaces the closure. `WorkspaceLocation.entityId` names a place, not a cache key |
| `WorkspaceSession.maxHiddenEntities` *(new)* | constant | `PersonRecordLinksContent.maxArcs`-style `// decided:` constants | A Swift cache bound, not a match knob (`core/matchregistry` is for matching layers) |
| `WorkspaceSession.lastShown`, `showCount` *(new)* | private state | `visibleKeys` | `visibleKeys` holds only the current place; the order needs every shown entity |
| `WorkspaceSession.evictEntities(_:)` *(new)* | private func | private `evict(_:)` | Calls `evict` per key of the named entities; one place for D-4 and D-5 |
| The cited-Source closure in the `.conclusionDetail` arm *(new)* | closure | `personClaims`' `graphIDs` closure | Same shape, different value |
| T-1, T-3, T-7, T-8, T-9 *(new)* | tests | `WorkspaceSessionTests` helpers `makeSession`, `detail(_:)`, `waitForQuery`, `settleWrites` | Reused; `detail(_:)` gains no parameter: tests build the cited detail inline |

Searched for: "entityId" on `CatalogQueryKey` (none), an LRU or "recent" in `macos/App/Features/Workspace` (none; `NavigationHistoryStore` holds history, not cache order), a set of `sourceID`s over a detail (none: `rg "flatMap\(\\\\.outcomes\)" macos/App` finds nothing).

**Builds on.** `CatalogQueryRefresh.applyHiddenPolicy(followsValue:ids:)` and the `.source` slot (`add-workspace-place` "Refresh"); the `.markStale` policy of `.evidenceGraph` and `.sourceWorkspace`; S16-04's `QueryReading` and `ConclusionDetailPresentation`.
**Deletes.** `CatalogQueryStalePolicy.swift` and the enum; `CatalogQueryDescriptor.stalePolicy` and `make`'s `stalePolicy:` parameter; 22 `stalePolicy: .sessionFresh` argument lines (20 arms plus `personSection` and `personGraphsSection`); `CatalogQueryRegistry.stalePolicy(for:)`; `CatalogHiddenPolicy.evict` and its arm in `applyHiddenPolicy(to:)`; `personGraphsSection`'s `entityId:` parameter and the two `guard case` closures passed to it; the `.conclusionDetail` arm's `guard case` id closure; five `stalePolicy` test assertions; the "S9-15 exception" paragraph and the `.evict` clause in `macos-client-patterns.md`; `.evict` in `add-workspace-place`.
**Next adds.** S16-25 replaces `entityId` with the plain-value key's ids and the cited-Source closure with `.followsInputs`; S16-29's prefetch warms keys that D-5's cap already counts; S16-12 adds `.source` to the claims and record links keys the same way D-1 does here.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Back to a page whose key is stale (was evicted) | `ConclusionDetailPage`, `ConclusionPageSectionHost` | `ensureQuery` refetches behind the value (`beginLoading(keepingValue: true)`), the header shows "Refreshing" (`isRefreshing`, `refreshing` in `ConclusionDetailLoaded`); a not-found reload reads gone (T-8); a failed reload shows S16-04's notice above the kept value |
| A stale value of a merged entity (id in `handles`, not `deleted`) | same | Reload throws `conclusiondetails.not_found` → `.gone` → the page leaves (T-8). Its keys stay counted until D-5 trims them |
| A visible key of a deleted entity, no `leavingPage` (an Unpromote from the Evidence graph with the person's page in another workspace place is hidden, so this is only reachable from code that applies deleted effects while the page shows) | `apply(_:_:)` | Not evicted (D-4); revalidated; reads gone |

### Scale

At the map's Scale (a person in 50 Evidence graphs with one or two records each, about ten fields; 50 hidden entities):

- **Per write.** `apply` walks the warmed keys once more for D-4: at most 50 entities × 5 keys + about 20 others, about 270 keys. `revalidatedKeys` calls D-1's closure for each warmed detail when the write reports `sourceScopes`: about 51 details × a few hundred outcomes, under 20,000 set inserts. A visible detail that cites the Source reloads once (`GetConclusionDetail`); hidden ones are only flagged. Rule "A cached detail follows the Sources its records cite" accepts that any write in a cited Source reloads a visible detail.
- **Per navigation.** D-5 builds the set of hidden entity ids from the warmed keys (about 270) and sorts only when over 50.
- **Memory.** At most 51 entities' detail and four section values. A crafted person's values are bounded by S16-01's caps; the cap of 50 bounds how many are held.

No budget is set or claimed; this PR claims no speed, so rule "Measure growth" does not bind it.

### Edges

- **Assumptions.** `apply(_:_:)`'s passes iterate `Array(handles.keys)` taken after D-4's eviction, so a deleted entity's hidden key is neither patched nor refreshed (T-7). `followsValue` matches any Source until the detail holds a value, so a first load in flight is restarted by a Source write, as today for graphs (`personClaimsFollowEveryGraphUntilLoaded`).
- **Neighbors.** A write naming a Source no record cites leaves the detail alone (AC-1.2). A write naming the person's handle still revalidates the visible page and stales a hidden one (AC-2.1). A key one entity over the cap is evicted; one at the cap is kept (AC-4.1, AC-4.2).
- **Removed inputs.** A deleted entity's hidden keys are dropped (AC-3.1); a merged entity's keys go stale and read gone on return (AC-3.3); the cap drops the least recently shown entity's every key at once (AC-4.2).
- **Real data.** None — no rule driven by a shipped table.
- **Untrusted size.** The cited-Source set is the detail's records, which S16-01's caps bound per graph; the number of details the closure runs over is bounded by D-5. No other list from data the app did not write.

### Schema

None — no catalog change.

### Go core

None — Swift only; the effects already carry `sourceScopes` and `deleted`.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Hidden policy | `CatalogHiddenPolicy` | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` | `revalidate`, `markStale` (D-3) |
| Detail descriptor | `.conclusionDetail` arm | same | `whenHidden: .markStale`; refreshes: own handle via `key.entityId`, `.source` via D-1, property labels |
| Section helpers | `personSection`, `personGraphsSection` | same | `whenHidden: .markStale`; `personGraphsSection(kind:graphIDs:load:)` |
| Key id | `CatalogQueryKey.entityId` | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` | D-6 |
| Session | `apply(_:_:)`, `apply(location:placeRegistry:)`, `evictEntities(_:)`, `maxHiddenEntities` | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` | D-4, D-5 |

Sketch:

```swift
extension CatalogQueryKey {
    /// The canonical entity a per-entity Conclusion key reads; nil for every other key.
    var entityId: String? { /* the five per-entity cases' entityId; nil otherwise */ }
}

// WorkspaceSession
/// Hidden entities whose per-entity keys stay cached (D-5).
static let maxHiddenEntities = 50 // decided: user, 2026-11-08
private var lastShown: [String: Int] = [:]
private var showCount = 0
private func evictEntities(_ ids: Set<String>)
```

States: a stale page shows its content with "Refreshing" (existing); a gone one leaves (existing); a failed reload shows S16-04's notice (existing). Accessibility: unchanged.

### Strings

None — no copy is added, removed or reworded.

## 7. Steps

**STEP-1 — The key names its entity**
Files: `macos/App/Features/Workspace/Session/CatalogQueryKey.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`. Skill: `add-workspace-place`.
Do: D-6. Add `entityId`. `personGraphsSection` drops its `entityId:` parameter and follows `.touchedHandles` with `key.entityId.map { [$0] } ?? []`; `personClaims` and `personRecordLinks` stop passing it; the `.conclusionDetail` arm's own-handle closure becomes the same expression. No policy change in this step.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/WorkspaceSessionTests` passes unchanged.

**STEP-2 — The detail follows the Sources it cites**
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-swift-test`.
Do: D-1. Write T-1 and T-2.
Check: `-only-testing:ProvenenciaTests/WorkspaceSessionTests` passes.

**STEP-3 — A deleted entity's hidden keys are evicted**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift` (comment only).
Do: D-4, D-9, D-10. Add `evictEntities(_:)` and call it at the top of `apply(_:_:)`; rewrite `apply(_:_:)`'s doc comment to start with the eviction. Write T-7. The keys are still `.evict` in this step, so the eviction changes no test result yet; it lands first so STEP-4 never leaves a deleted page's keys stale. The `UnpromoteFlowTests` comment "leaving there evicts her page" becomes "`apply` evicts her page: she is deleted and, leaving, hidden".
Check: `-only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/UnpromoteFlowTests` passes.

**STEP-4 — Hidden Conclusion keys go stale**
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift` (`applyHiddenPolicy(to:)`), `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` (the five doc comments: "Marked stale, not revalidated, while off screen"), `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `macos/ProvenenciaTests/PropertiesModelTests.swift`.
Do: D-2, D-3, D-7. Update the `.evict` assertions in `hiddenPolicies` and the four section tests to `.markStale` and their doc comments ("marked stale while hidden"); rewrite `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`, `touchedHandlesRevalidateOnlyTheirDetailPages`, `aKeyInvalidatedAndRefreshedRevalidatesOnce` and `renamingAPropertyStalesConclusionPages` as T-4, T-5, T-6 and T-11; write T-3 and T-8.
Check: `-only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/PropertiesModelTests -only-testing:ProvenenciaTests/UnpromoteFlowTests -only-testing:ProvenenciaTests/PersonDetailTests` passes (T-12 unchanged); `rg "\.evict\b" macos/App` finds nothing.

**STEP-5 — At most 50 hidden entities**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`.
Do: D-5. Add `maxHiddenEntities`, `lastShown`, `showCount`; extend `apply(location:)` and `evictEntities`. Write T-9.
Check: `-only-testing:ProvenenciaTests/WorkspaceSessionTests` passes.

**STEP-6 — The unused stale policy goes**
Files: `macos/App/Features/Workspace/Session/CatalogQueryStalePolicy.swift` (deleted), `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`.
Do: D-8. Write T-10's `sessionUsesStandardRegistryByDefault` change.
Check: the app builds; `-only-testing:ProvenenciaTests/CatalogQueryRegistryTests` passes; `rg "CatalogQueryStalePolicy|stalePolicy|sessionFresh|refetchOnRevisit" macos` finds nothing.

**STEP-7 — Docs and version**
Files: `.agents/skills/add-workspace-place/SKILL.md` (the "Refresh" row's slot list gains `.source`; the "Mark stale" row: `whenHidden` is `.revalidate` or `.markStale`, the per-entity Conclusion keys are `.markStale`, `apply` evicts the hidden keys of deleted handles, at most `maxHiddenEntities` hidden entities keep keys), `docs/macos-client-patterns.md` (rule 2's `whenHidden` sentence drops `.evict` and says Conclusion detail and section keys are `.markStale`; "Conclusion keys" replaces the "Exception to rule 2 (S9-15)" bullet with what the per-entity keys do while hidden, the deleted eviction and the cap, and says the detail key follows its own handle, property labels and the Sources its records cite), `docs/deployment-plan/spike-16/deployment-plan.md` (Batch 1 rule clause, D-12; map row), `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj`, `docs/deployment-plan/spike-16/completed.md`, this plan.
Do: per the Definition of done.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a renamed Source's words in Vitals, the Gallery and the viewer.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `WorkspaceSessionTests.swift` `detailFollowsTheSourcesItsRecordsCite` *(new)* | `setQueryValue` a detail for "e1" with a name field outcome `sourceID` "s-field", a `.birthDate` derived row owner whose field outcome cites "s-owner", a picture whose record cites "s-pic" → `revalidatedKeys(for: CatalogWriteEffects(sourceScopes: [id]))` == [detail key] for each; ["s-other"] → empty. Then `apply(location:)` e1, wait; `apply(sourceScopes: ["s-pic"])` → `isFetching`, `settleWrites` → value current. Then `apply(location:)` the Persons list; `apply(sourceScopes: ["s-pic"])` → value kept, `isStale`, not fetching | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1 |
| T-2 | Swift | `WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs` (updated) | Doc comment: "the detail page follows the handle and the Sources its records cite". Existing asserts stay; add: a detail citing "s1" → `revalidatedKeys(sourceScopes: ["s1"])` == [detailKey] (the claims key does not follow Sources until S16-12) | AC-1.1, D-1 |
| T-3 | Swift | `WorkspaceSessionTests.swift` `backToAStalePageShowsItsValueWhileItRefreshes` *(new)* | FakeStore seeds details e1, e3 and family e1; show e1 on Family, then Vitals, wait; show e3; `apply(handles: ["e1"])` → e1 detail and family handles non-nil, `isStale`, not fetching; reseed e1's detail with a field; `apply(location:)` e1 → detail `value` is the old one and `isFetching`; wait → the new value, `isStale` false | AC-2.1, AC-2.2, D-2 |
| T-4 | Swift | `WorkspaceSessionTests.swift` `applyTriggerStalesHiddenDetailPagesAndRevalidatesTheVisibleOne` (renamed from `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`) | As today, but the first key's handle is non-nil, keeps its value and `isStale` | AC-2.1, D-2 |
| T-5 | Swift | `WorkspaceSessionTests.swift` `touchedHandlesRevalidateOnlyTheirDetailPages` (updated) | Last assert: e3's handle is non-nil and `isStale`; doc comment "a hidden one goes stale" | AC-2.1 |
| T-6 | Swift | `WorkspaceSessionTests.swift` `aKeyInvalidatedAndRefreshedRevalidatesOnce` (updated) | The hidden claims key is non-nil, `isStale`, not fetching, after `settleWrites` | AC-2.1, AC-2.3 |
| T-7 | Swift | `WorkspaceSessionTests.swift` `deletedEntitiesLoseTheirHiddenKeys` *(new)* | `setQueryValue` detail, family and claims for e1 and detail for e3; `apply(location:)` the Persons list; `apply(CatalogWriteEffects(handles: ["e1"], deleted: [.handle ["e1"]]))` → e1's three handles nil, e3's untouched and not stale. Then `apply(location:)` e3; `apply(deleted: [.handle ["e3"]])` → e3's handle non-nil and fetching (visible, not evicted) | AC-3.1, AC-3.2, D-4 |
| T-8 | Swift | `WorkspaceSessionTests.swift` `aStaleReloadOfAGoneEntityReadsAsGone` *(new)* | Seed e1; show e1, wait; show the Persons list; `apply(handles: ["e1"])` → stale; remove e1's seed (FakeStore throws `conclusiondetails.not_found`); `apply(location:)` e1, `settleWrites` → `reading` is `.loaded(_, refreshFailure: non-nil)` and `ConclusionDetailPresentation(handle.reading, isFetching: false) { $0.entity.id }` == `.gone` | AC-3.3, D-7 |
| T-9 | Swift | `WorkspaceSessionTests.swift` `hiddenConclusionPagesAreCappedAtFifty` *(new)* | FakeStore seeds details e0 – e52 and family e0; `setQueryValue` a detail for "never"; show e0 on Family, then e1 … e50 → "never" nil, e0's detail and family non-nil; show e51 → e0's detail and family nil, e1 – e50 non-nil; show e1, then e52 → e2 nil, e1 non-nil; the visible entity's keys are non-nil after every step | AC-4.1, AC-4.2, AC-4.3, AC-4.4, D-5 |
| T-10 | Swift | `CatalogQueryRegistryTests.swift` `hiddenPolicies`, `personClaimsRevalidateOnTouchedHandle`, `personFamilyRevalidateOnTouchedHandle`, `personEventsRevalidateOnTouchedHandle`, `personRecordLinksRevalidate`, `sessionUsesStandardRegistryByDefault` (updated) | Each per-entity kind → `.markStale`; no `stalePolicy` assert; the default-registry test asserts `hiddenPolicy(.sourcesList) == .markStale` | AC-2.3, AC-5.1, D-8 |
| T-11 | Swift | `PropertiesModelTests.swift` `renamingAPropertyStalesConclusionPages` (updated) | The hidden detail handle is non-nil and `isStale`; doc comment "a hidden Person page is marked stale" | AC-2.4 |
| T-12 | Swift | `WorkspaceSessionTests.swift` `applyLeavingPageLoadsTheNextListOnce`, `UnpromoteFlowTests.swift` confirm test (unchanged asserts) | The deleted page's key is evicted; the list loads once | AC-3.4, D-9 |
| T-13 | Sweep | STEP-4 and STEP-6 checks | `rg "\.evict\b" macos/App` and `rg CatalogQueryStalePolicy macos`, `rg stalePolicy macos`, `rg sessionFresh macos` and `rg refetchOnRevisit macos` → no hit | AC-5.1, AC-5.2, D-3, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceSessionTests` (every test that warms a per-entity key and then expects it dropped), `CatalogQueryRegistryTests`, `PropertiesModelTests`, `UnpromoteFlowTests`, `PersonDetailTests`, `PersonSourcesSection` previews (they build a session and visit one person).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none added)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`CatalogQueryStalePolicy`, `stalePolicy`, `refetchOnRevisit`, `sessionFresh`, `.evict`, "Exception to rule 2 (S9-15)", "Evicted, not revalidated"); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (the Batch 1 clause, D-12)
- [ ] UI matches the brief inventory; no new kit primitive (no brief: no chrome change)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): 0.7.1 → 0.7.2, or the next patch on `main`
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: no brief)
- [ ] add-workspace-place ("Refresh" slot list, "Mark stale") and macos-client-patterns.md (rule 2, "Conclusion keys") say what the per-entity keys do while hidden and what the detail key follows
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person whose birth is recorded in a Source titled "Parish register A" and who has a picture from that Source. Open the person on Vitals and open the birth row's provenance sheet: it names "Parish register A". Close it.
2. Open the Source's page, rename it "Parish register B". Go Back to the person: the page shows at once with "Refreshing" beside the switcher, not a spinner; the provenance sheet and the Gallery caption name "Parish register B"; open the picture in the viewer: its record row names "Parish register B".
3. On the Source page, edit the Citation of the birth record: change its label and page. Back to the person: the provenance sheet shows the new label and page.
4. Open the person, then the Evidence graph that holds their only record, and Unpromote it. Back from the graph: the person's page does not show; the Persons list does.
5. Open 52 different people one after another from the Persons list, then open the first again from the list: it loads with a spinner (the cap evicted its page). Open the 30th: it shows at once (no write happened, so its page is current and does not reload).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test or feature relies on a hidden Conclusion key being dropped after a write | §4 lists every test found by `rg "evicted\|== nil" macos/ProvenenciaTests` around these keys; STEP-4's check runs the suites that hold them |
| A visible key evicted leaves a spinner | D-4 and D-5 skip `visibleKeys`; T-7's second half and T-9's last clause pin it |
| A Source write reloads the visible page often | Accepted by the Batch 1 rule's stated fan-out; one `GetConclusionDetail` per write |

**Stop and report if:**

- `ConclusionDetailPresentation` no longer reads a `.loaded` reading with a `conclusiondetails.not_found` refresh failure as `.gone` (D-7 rests on it).
- A production caller other than `apply(location:)` warms a per-entity key for an entity no place shows (D-5 counts it as least recently shown; `rg "\.conclusionDetail\(project|\.person(Claims|Family|Events|RecordLinks)\(project" macos/App` outside `CatalogQueryRegistry.swift`, `PlaceRegistry.swift`, `ConclusionPageSection*` and the section views finds one).
- `apply(_:_:leavingPage:)` is called with effects whose `deleted` does not name the leaving entity (the page's keys would then go stale instead of being evicted).
- A descriptor other than the five Conclusion kinds sets `whenHidden: .evict` on `main` when you start (D-3 deletes the case).
- T-9 needs more than 53 seeded details to show the cap (the constant or the counting differs from D-5).

## 12. Open questions

None.

## 13. PR

**Title:** S16-09 — Conclusion pages refresh from the Sources they cite and show their last value while they refresh
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-09-source-follow-and-stale-pages.md`.
