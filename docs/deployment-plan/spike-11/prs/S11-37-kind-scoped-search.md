# S11-37 — One kind-scoped catalog search

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 8 — Find another record
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the omnibar draws the same panel and rows from the moved code; Promote's picker is S11-36's, gated on S11-D3)
**Board:** —
**Depends on:** —
**Blocks:** S11-36
**Branch:** `s11-37-kind-scoped-search`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md) (STEP-6 edits it)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [Find another record § Where the Swift side changes](../promote-find-another-record.md#where-the-swift-side-changes), [§ Decided](../promote-find-another-record.md#decided-user-2026-10-24); working rule "No new search RPC, and one search model"; [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Planned:** 2026-10-25 against `main` @ `7f8117d`
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

The omnibar's search state (query, debounce, generation guard, loading, error and empty states, keyboard selection) and its hit row move out of `Features/Workspace` into one shared home, `macos/App/Features/Catalog/Search/` *(new)*, as `CatalogSearchModel` and `CatalogSearchHitRow`. The model takes the kinds to search and an optional location, so a caller with no workspace place to give `SearchCatalog` (Promote's picker) and a caller that wants one kind (a person row) use the same model the omnibar uses. The omnibar keeps every behavior: its results model becomes the omnibar's panel state around a `CatalogSearchModel`.

**Why now.** S11-36's inline picker searches records of one kind from the Promote page. Without this PR it rebuilds the omnibar's debounce and generation guard inside Promote and copies the private `OmnibarHitRowView` (map review K-1, 2026-10-24; working rule "No new search RPC, and one search model"). S11-37 runs in parallel with S11-35 (landed in [#543](https://github.com/mendahu/provenencia/pull/543)) ahead of S11-36.

**What stays.** The omnibar's look, copy, keyboard handling, outside-click dismissal, accessibility identifiers and ranking; `SearchCatalog` and the Go search engine (no Go change); the `PVOmnibarHitRow` recipe; every `L10n.Workspace.omnibar*` key.

## 2. User stories

**US-1.** As a researcher, when I type in the omnibar, I get the same results, states and keyboard behavior as before, so that the refactor costs me nothing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Every case in `OmnibarResultsModelTests` at the planning base, retargeted per D-6 | Each passes with the same inputs and the same expectations |
| AC-1.2 | Omnibar: type `Ilminster` on the Sources list of a project with a Source titled `Ilminster parish` | The panel shows that Source as the first hit, with its type label and thumbnail; ↓ / ↑ move the selection; Return opens the Source and clears the field; Esc closes the panel and keeps the query |

**US-2.** As a contributor (S11-36), I drive one search model with a set of kinds and no location, so that a person row's picker searches only person records without a copy of the omnibar's model.

In this table, **the Robins store** is `FakeStore` set up as `SearchCatalogStoreTests.kindPriorityIsATenthAndSectionNudgeStillMatters` sets it up: a Source `SRC-1` titled `Robins register` and one person handle (James Robins, two member Subjects) promoted from `sub-1` and `sub-2`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The Robins store; `CatalogSearchModel(kinds: ["person"])`, query `Robins`, `scheduleSearch(projectDir:location: nil, store:)` | After the search settles, `hits.map(\.kind) == ["person"]` and `hits.first?.ref` is James's handle ref; `hasSearched` true, `searchError` nil |
| AC-2.2 | The Robins store; `CatalogSearchModel()` (no kinds), query `Robins`, location nil | `hits.map(\.kind) == ["person", "source"]`: no section nudge, so the kind share orders them, as on `.sectionRoot(.persons)` today |
| AC-2.3 | The Robins store; `store.searchCatalog(projectDir:query: "Robins", location: nil, kinds: [])` | `["person", "source"]`, the same order as AC-2.2 |
| AC-2.4 | `GoStore.searchCatalog` with `location: nil` | The request's `location` field is unset (`hasLocation == false`); Go reads an unset location as no section (`locationFromProto` returns the zero location, `contextMultiplier` returns 1). Pinned on the Go side already by `TestSearchCatalogHandleKinds`, which sends no location |

## 3. Scope

| In | Covered by |
| --- | --- |
| `GenealogyStore.searchCatalog`'s `location` becomes `WorkspaceLocation?`; `GoStore` leaves the proto field unset on nil; `FakeStore` gives no section boost on nil; `ThrowingStore` (test) follows the signature | STEP-1, AC-2.3, AC-2.4 |
| `CatalogSearchModel` *(new)*, `Features/Catalog/Search/CatalogSearchModel.swift`: the search state moved out of `OmnibarResultsModel`, with `kinds` (init) and an optional location (`scheduleSearch`) | STEP-2, AC-2.1, AC-2.2 |
| `OmnibarHitPresentation` moves to `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` *(new)* and is renamed `CatalogSearchHitPresentation` | STEP-3, AC-1.1 |
| `OmnibarHitRowView` moves out of `OmnibarResultsHost.swift` to `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` *(new)* as internal `CatalogSearchHitRow` | STEP-3, AC-1.2 |
| `OmnibarResultsModel` keeps only the omnibar's panel state and composes a `CatalogSearchModel`; `OmnibarResultsHost`, `OmnibarResultsPanel`, `WorkspaceToolbar`, `WorkspaceContent` read search state through it | STEP-4, AC-1.1, AC-1.2 |
| Tests: `OmnibarResultsModelTests` retargeted; search-only and presentation cases move to `CatalogSearchTests.swift` *(new)*; new kinds and nil-location cases | STEP-5, T-1 – T-6 |
| `add-searchable-kind` names the shared search as the Swift consumer; the feature doc's search bullet and `docs/ideas/archive/glyph-consolidation.md` are left as history (D-10); the map row, `completed.md` | STEP-6 |

| Out | Home |
| --- | --- |
| Promote's picker, the menu sentinel, `PromoteFlow.hold` callers | S11-36 |
| Any change to ranking, hit kinds, `SearchCatalog`, `core/search` | not planned (map scope "Not this PR") |
| Renaming `PVOmnibarHitRow` or its folder | not planned: the recipe is kit API named for its board (Spike 3 Omnibar Hit Row); S11-36 composes it through `CatalogSearchHitRow` |
| Renaming `L10n.Workspace.omnibar*` keys the presentation reads | not planned: no copy changes, and key renames churn `Localizable.xcstrings` for no reader-visible gain |
| Moving `OmnibarResultsHost`'s Escape / outside-click monitor onto a shared popup presenter | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) ("move each onto it when its host is next touched" — this PR does not touch the monitor) |
| A picker view (field + list) shared by omnibar and Promote | not planned: S11-36 draws its own surface per S11-D3; the omnibar's panel stays an overlay |
| Removing the three-argument `searchCatalog(projectDir:query:location:)` extension | not planned (D-4) |

## 4. Current state

Verified at `7f8117d`.

| Area | Where | Today |
| --- | --- | --- |
| Omnibar model | `macos/App/Features/Workspace/OmnibarResultsModel.swift` `OmnibarResultsModel` (line 7) | One `@Observable` class mixing search state (`query`, `hits`, `isLoading`, `hasSearched`, `searchError`, `selectedIndex`, `searchGeneration`, `debounceTask`, `minQueryLength`, `debounceNanoseconds`, `trimmedQuery`, `showsLoadingState` / `showsErrorState` / `showsEmptyState`, `dismiss()`, `scheduleSearch(projectDir:location:store:)`, `moveSelection(delta:)`, `selectedHit()`, `runSearch`) with omnibar panel state (`panelWidth`, `userDismissed`, `fieldFrame`, `isPresented`, `closePanel()`, `clearAfterNavigate()`, `activate(_:navigation:)`). `runSearch` (line 122) calls the three-argument `store.searchCatalog`, so it never sends `kinds`; `location` is required. `closePanel()` (line 47) sets `userDismissed` and `isLoading = false` without cancelling the in-flight search |
| Omnibar host | `macos/App/Features/Workspace/OmnibarResultsHost.swift` `OmnibarResultsHost` (line 9), `OmnibarResultsPanel` (line 92, private), `OmnibarHitRowView` (line 193, private), `OmnibarFieldAnchorKey` (line 232) | The host places the panel under `fieldFrame` and runs the Esc / ↑ / ↓ / Return / outside-click monitor; the panel switches on `showsErrorState` / `showsEmptyState` / `showsLoadingState` and draws each hit as a `PVMenuRowButtonStyle` button around `OmnibarHitRowView`, with ids `workspace.toolbar.omnibar.hit.<id>`. `OmnibarHitRowView` reads `WorkspaceSession.nameFormat` from the environment and draws `CachedThumbnail` for a `source` hit, `PVThumbnail` otherwise, into `PVOmnibarHitRow` |
| Hit copy | `macos/App/Features/Workspace/OmnibarHitPresentation.swift` `enum OmnibarHitPresentation` | Kind label, match context, ref accent, VoiceOver label, title (person / event / place header or `hit.title`), secondary line, lead thumbnail. Kind-agnostic; reads `L10n.Workspace.omnibar*`. Extracted from the model in `941c87b2` |
| Toolbar | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `@Bindable var omnibarResults` (line 35); `.onPreferenceChange(OmnibarFieldAnchorKey.self)` (line 62); `.onChange(of: omnibarResults.query)` → `scheduleSearch(projectDir:location: navigation.currentLocation, store:)` (lines 73 – 79); `PVInput(text: $omnibarResults.query, …)` (line 167); preview `OmnibarResultsModel()` (line 592) | Binds the field to the model's query and reschedules on each change |
| Content | `macos/App/Features/Workspace/WorkspaceContent.swift` `@State omnibarResults` (line 16); `OmnibarResultsHost(results:projectDir:onActivate:)` overlay (lines 39 – 47) | Owns the model; activation calls `omnibarResults.activate(hit, navigation:)` |
| Store API | `macos/App/Platform/GenealogyStore.swift` protocol `searchCatalog(projectDir:query:location:kinds:)` (line 1344, `location: WorkspaceLocation`); extension `searchCatalog(projectDir:query:location:)` (line 1631) passing `kinds: []` | The only app caller of either is `OmnibarResultsModel.runSearch`. The three-argument form is also called by every case in `SearchCatalogStoreTests` |
| Go store | `macos/App/Platform/GoStore.swift` `searchCatalog` (line 809) | Always sets `req.location = Self.mapWorkspaceLocationToProto(location)` |
| Fake store | `macos/App/Platform/FakeStore.swift` `searchCatalog` (line 1213), `omnibarSearchCatalog(projectDir:query:location:)` (private) | Handle hits ignore location. The Source / type / field fake boosts by `location.section` (`== .sources`, `== .sourceTypes`, `== .metadata` → 1.1, about lines 1377, 1412, 1443) |
| Throwing store | `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` (line 342), its `searchCatalog` (line 567) | The third `GenealogyStore` conformer; must follow the protocol signature |
| Go side | `api/ffi/handlers/search_catalog.go` `SearchCatalog`, `locationFromProto` (line 43); `core/search/registry.go` `contextMultiplier` (line 193) | A nil proto location becomes the zero `search.WorkspaceLocation`; an empty section gives multiplier 1. `api/ffi/handlers/search_catalog_test.go` `TestSearchCatalogHandleKinds` (line 212) already searches with no location and `kinds: ["person"]` |
| Shared feature glue | `macos/App/Features/Catalog/` (pbxproj group `Catalog`, `AA0000000000000000000220`, line 869), subfolder `macos/App/Features/Catalog/Connect/` (group `AA0000000000000000000660`) | Cross-feature, catalog-domain Swift used by several features (`CatalogCounts`, `PropertiesTypeChrome`, the Connect folder used by the composer and Evidence graph) |
| Session in env | `macos/App/Features/Workspace/WorkspaceView.swift` `.environment(session)` (line 72) | Every workspace destination, Promote included, has `WorkspaceSession` in the environment, so the row can keep reading `nameFormat` from it |
| Tests | `macos/ProvenenciaTests/OmnibarResultsModelTests.swift` (14 cases: `shortQueryDoesNotPresent`, `debouncedSearchReturnsHits`, `exactRefHitIsTop`, `clearAfterNavigateResetsQuery`, `accessibilityLabelComposesVisibleRowMetadata`, `accessibilityLabelOmitsEmptyOptionalParts`, `matchContextLocalizesBodyFieldsAndHidesTitleRef`, `activateCommitsNavigationAndClearsResults`, `activateFieldHitFromSearch`, `closePanelSetsUserDismissedWithoutClearingQuery`, `loadingPresentsBeforeHitsArrive`, `emptyStateRequiresCompletedSearchWithoutError`, `searchFailureSurfacesMappedErrorNotEmpty`, `moveSelectionClampsAndSelectedHit`), private `waitForSearch(_:)` (line 10), private `leadHit(kind:iconKey:)`; `macos/ProvenenciaTests/SearchCatalogStoreTests.swift` `kindPriorityIsATenthAndSectionNudgeStillMatters` (line 144) | Pin the model and presentation; the store test pins the section nudge and `kinds: ["person"]` at the store level |

**Precedent.** `d9278991` ("…a catalog connect folder") made `Features/Catalog/Connect/` the shared home for Connect glue that several features read: copy its shape — a subfolder of `Features/Catalog`, a pbxproj group under `Catalog`, no `PV` prefix (snowflake-layer naming, [`design-system-layers.md`](../../../design-system-layers.md) § PV type prefix). `941c87b2` split `OmnibarHitPresentation` out of the model: copy its "move the type whole, change no body" discipline. Do not copy `2fec75f7`'s move into `DesignSystem/` (that was a content-agnostic painter; this code reads `GenealogyStore` and `WorkspaceSession`, which recipes and components must not).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The shared home is `macos/App/Features/Catalog/Search/` *(new)* with a pbxproj group `Search` under `Catalog`. Types: `CatalogSearchModel`, `CatalogSearchHitPresentation`, `CatalogSearchHitRow`, no `PV` prefix | `add-ui-component`: recipes and components must not touch `GenealogyStore` or feature state, and the model calls the store and the row reads `WorkspaceSession`. `Features/Catalog` already holds cross-feature catalog glue (precedent `d9278991`). `CatalogSearch*` matches the store's `CatalogSearchHit` | a recipe folder under `DesignSystem/Recipes/` — because a recipe may not read the store or session; keeping the files in `Features/Workspace` — because Promote would import Workspace internals; `Platform/` — because Platform holds store and infrastructure, not views |
| D-2 | `CatalogSearchModel` is `@MainActor @Observable final class` with `init(kinds: [String] = [])` and `let kinds: [String]`. It owns, moved unchanged from `OmnibarResultsModel`: `minQueryLength`, `debounceNanoseconds`, `query`, `hits`, `isLoading`, `hasSearched`, `searchError`, `selectedIndex`, `trimmedQuery`, `showsLoadingState`, `showsErrorState`, `showsEmptyState`, `dismiss()` (minus `userDismissed`), `moveSelection(delta:)`, `selectedHit()`, the generation guard, the debounce task and `runSearch`. `scheduleSearch(projectDir:location: WorkspaceLocation?, store:)`. `runSearch` calls the four-argument `store.searchCatalog(…, location: location, kinds: kinds)`. Stored properties stay `var` and settable, as today | Kinds are fixed per picker (a person row searches persons for its life), so they belong at init; location changes with navigation, so it is a call argument. Settable state keeps today's tests' setups | Kinds as a `scheduleSearch` argument — because a caller could change kinds between debounce and result and the generation guard would not know; a `SearchScope` struct — because two fields do not need a type |
| D-3 | `GenealogyStore.searchCatalog`'s `location` becomes `WorkspaceLocation?`. `GoStore`: `if let location { req.location = Self.mapWorkspaceLocationToProto(location) }`. `FakeStore.omnibarSearchCatalog` takes `WorkspaceLocation?` and its three boosts read `location?.section == …`. `ThrowingStore` changes its parameter type only | Promote has no workspace place to rank by; an unset proto location is what Go already reads as "no section" (AC-2.4). Passing Promote's own location would nudge by whatever section the Promote surface lives under, which is a ranking side effect the map rules out | A neutral `WorkspaceLocation` value — because `WorkspaceSection` has no empty case and any real section nudges ranking; making only the model's parameter optional and substituting a location — same reason |
| D-4 | The three-argument extension `searchCatalog(projectDir:query:location:)` stays, its `location` stays non-optional, and its doc comment changes from "Omnibar search: the default kind set." to "The default kind set." | `SearchCatalogStoreTests` calls it in every case; after this PR no app code calls it, but it is the store-level spelling of "default kinds", not omnibar code | Deleting it — because it churns ten store tests for no behavior |
| D-5 | `OmnibarResultsModel` keeps only `panelWidth`, `userDismissed`, `fieldFrame`, `isPresented`, `closePanel()`, `clearAfterNavigate()`, `activate(_:navigation:)`, plus `let search = CatalogSearchModel()` and two wrappers: `scheduleSearch(projectDir:location: WorkspaceLocation, store:)` sets `userDismissed = false` then calls `search.scheduleSearch`; `dismiss()` calls `search.dismiss()` then sets `userDismissed = false`. `isPresented` reads `search.trimmedQuery`, `search.hasSearched`, `search.isLoading`. `clearAfterNavigate()` sets `search.query = ""` then calls `dismiss()`. No forwarding properties for `hits`, `query` and the rest: callers write `results.search.hits` | Composition keeps one owner per piece of state. Today `scheduleSearch` also resets `userDismissed` on the short-query path and the normal path; the wrapper does it once before either, which is the same observable result | Subclassing — because a subclass of an `@Observable` class does not observe its own new stored properties without a second macro expansion, a known trap; forwarding computed properties — because each is a second spelling of the same state |
| D-6 | `closePanel()` sets `userDismissed = true` and calls `search.cancel()` *(new)*: cancels the debounce task, bumps the generation, sets `isLoading = false`, and leaves `query`, `hits`, `hasSearched`, `searchError` as they are | Today `closePanel` writes `isLoading = false` while the search keeps running; with composition that write would reach into the shared model. After Esc the panel stays hidden until `scheduleSearch` or `dismiss` runs, and both replace or clear `hits`, so dropping the in-flight result is not observable. `closePanelSetsUserDismissedWithoutClearingQuery` still passes (`!isLoading`, query kept) | Letting the omnibar set `search.isLoading = false` — because it leaves a search running whose result lands with `isLoading` already false, and makes a second writer of the shared state |
| D-7 | `OmnibarHitPresentation` is moved and renamed `CatalogSearchHitPresentation` with no body change. It keeps reading `L10n.Workspace.omnibar*` | One type name per shared concept; the body is already kind-agnostic. Key names are not reader-visible | Keeping the `Omnibar` name in the shared folder — because S11-36 would call omnibar code from Promote; renaming the L10n keys — Out |
| D-8 | `OmnibarHitRowView` moves whole to `CatalogSearchHitRow.swift`, internal (not `private`), renamed `CatalogSearchHitRow`, same stored properties (`hit`, `projectDir`, `selected`), same `@Environment(WorkspaceSession.self)` read. The `Button`, `PVMenuRowButtonStyle`, accessibility label and `workspace.toolbar.omnibar.hit.<id>` identifier stay in `OmnibarResultsPanel` | The row is the shared piece; the button chrome and ids belong to each call site (`add-ui-component`: a11y ids at call sites) | Moving the panel's button loop too — because S11-36's picker list is S11-D3's to draw; passing `nameFormat` as a parameter — because every workspace destination has the session in the environment (Current state "Session in env") and the call sites stay one line |
| D-9 | Tests: a case stays in `OmnibarResultsModelTests` when it reads an omnibar-only member (`isPresented`, `userDismissed`, `activate`, `closePanel`, `clearAfterNavigate`); it moves to `CatalogSearchTests.swift` *(new)* otherwise. Staying: `shortQueryDoesNotPresent`, `debouncedSearchReturnsHits`, `clearAfterNavigateResetsQuery`, `activateCommitsNavigationAndClearsResults`, `activateFieldHitFromSearch`, `closePanelSetsUserDismissedWithoutClearingQuery`, `loadingPresentsBeforeHitsArrive`, `searchFailureSurfacesMappedErrorNotEmpty`. Moving to `@Suite struct CatalogSearchModelTests`: `exactRefHitIsTop`, `emptyStateRequiresCompletedSearchWithoutError`, `moveSelectionClampsAndSelectedHit`. Moving to `@Suite struct CatalogSearchHitPresentationTests`: `accessibilityLabelComposesVisibleRowMetadata`, `accessibilityLabelOmitsEmptyOptionalParts`, `matchContextLocalizesBodyFieldsAndHidesTitleRef` (with `leadHit`). Each case keeps its name, inputs and expectations; only member paths (`model.search.hits`) and type names change. `waitForSearch` becomes one internal test helper, `extension CatalogSearchModel { func settle() async }`, in `CatalogSearchTests.swift`, used by both files | One rule a reviewer can check; one copy of the wait loop | Moving every case — because the panel cases test omnibar state; leaving presentation cases in the omnibar file — because they test a shared type |
| D-10 | Docs: `add-searchable-kind` § "4. FFI + Mac store" gains one bullet naming `CatalogSearchModel` (`macos/App/Features/Catalog/Search/` *(new)*) as the Swift consumer of `searchCatalog`, with kinds at init and an optional location, used by the omnibar and S11-36's picker; § Related's "Omnibar UI" line adds the shared folder. `promote-find-another-record.md`'s "The search:" bullet is rewritten in the past tense with the new names. `docs/ideas/archive/glyph-consolidation.md` is history and is not edited | The map scope names the skill; the feature doc is read by S11-36's planner | Editing archived docs — because archives record what was true |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `macos/App/Features/Catalog/Search/` *(new)* | folder / pbxproj group | `Features/Catalog/Connect/` | Same shape, sibling folder (D-1) |
| `CatalogSearchModel` *(new)* | type | `OmnibarResultsModel` | It **is** that model's search half, moved; the omnibar half stays (D-5). Searched `Debounce`, `searchGeneration`, `Task.sleep(nanoseconds:` in `macos/App`: no other debounced search model exists |
| `CatalogSearchModel.cancel()` *(new)* | func | `dismiss()` | `dismiss()` also clears hits and `hasSearched`, which `closePanel` must not (D-6) |
| `CatalogSearchHitPresentation` *(renamed)* | type | `OmnibarHitPresentation` | Same type, moved and renamed (D-7) |
| `CatalogSearchHitRow` *(renamed)* | view | `OmnibarHitRowView`, `PVOmnibarHitRow` | Same view, moved and made internal; it maps a hit onto the existing recipe (D-8) |
| `CatalogSearchTests.swift` *(new)*, `CatalogSearchModel.settle()` *(new, test)* | test file / helper | `OmnibarResultsModelTests.waitForSearch` | That private helper, made shared (D-9) |

**Builds on.** `PVOmnibarHitRow` (recipe), `PVFloatingMenuSelection.moveIndex`, `L10n.Errors.message(for:)`, `GenealogyStore.searchCatalog`'s existing `kinds`, Go's nil-location handling.
**Deletes.** `macos/App/Features/Workspace/OmnibarHitPresentation.swift` (moved), the private `OmnibarHitRowView` in `OmnibarResultsHost.swift` (moved), the search half of `OmnibarResultsModel` (moved), the private `waitForSearch` in `OmnibarResultsModelTests.swift`.
**Next adds.** S11-36 holds one `CatalogSearchModel(kinds: [row kind])` per open picker, calls `scheduleSearch(projectDir:location: nil, store:)` on query change, draws `CatalogSearchHitRow` for each hit, and hands the picked hit's `id`, `ref`, `header` to `PromoteFlow.hold`. A later "pick a record" surface (merge, link to an existing record) does the same.

### Schema

None — no catalog change.

### Go core

None — `SearchCatalog` already takes `kinds` and treats an unset location as no section (Current state "Go side").

### FFI

None — no proto change; `SearchCatalogRequest.location` is already a message field that may be unset.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Search state | `CatalogSearchModel` *(new)* | `macos/App/Features/Catalog/Search/CatalogSearchModel.swift` | D-2, D-6 |
| Hit copy | `CatalogSearchHitPresentation` | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` *(new)* | D-7 |
| Hit row | `CatalogSearchHitRow` → `PVOmnibarHitRow` | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` *(new)* | D-8 |
| Omnibar panel state | `OmnibarResultsModel` | `macos/App/Features/Workspace/OmnibarResultsModel.swift` | D-5 |
| Omnibar panel | `OmnibarResultsHost`, `OmnibarResultsPanel` | `macos/App/Features/Workspace/OmnibarResultsHost.swift` | Read `results.search.*`; draw `CatalogSearchHitRow` |

Sketch:

```swift
@MainActor @Observable
final class CatalogSearchModel {
    static let minQueryLength = 2
    static let debounceNanoseconds: UInt64 = 180_000_000
    let kinds: [String]
    var query = ""
    var hits: [CatalogSearchHit] = []
    // isLoading, hasSearched, searchError, selectedIndex as today
    init(kinds: [String] = [])
    func scheduleSearch(projectDir: String, location: WorkspaceLocation?, store: any GenealogyStore)
    func cancel()
    func dismiss()
    // trimmedQuery, shows*State, moveSelection(delta:), selectedHit() as today
}
```

States: loading, error, empty and results draw exactly as today. Accessibility: labels and identifiers unchanged.

### Strings

None — no copy changes; the moved presentation reads the same `L10n.Workspace.omnibar*` keys.

## 7. Steps

**STEP-1 — Optional location in the store**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (`ThrowingStore`), `macos/ProvenenciaTests/SearchCatalogStoreTests.swift`. Skill: `add-swift-test`.
Do: D-3 and D-4. Add T-5 to `SearchCatalogStoreTests`. `OmnibarResultsModel.runSearch` still compiles unchanged (a non-optional argument converts).
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/SearchCatalogStoreTests -only-testing:ProvenenciaTests/OnboardingModelTests` passes.

**STEP-2 — The shared model**
Files: `macos/App/Features/Catalog/Search/CatalogSearchModel.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-ui-component` (classification: snowflake-layer shared glue, D-1).
Do: create the `Search` group under `Catalog` and the file per D-2 and D-6, moving bodies from `OmnibarResultsModel` verbatim (only `userDismissed` lines drop, and `runSearch` sends `kinds`). Leave `OmnibarResultsModel` untouched in this step; the two coexist for one commit.
Check: the app target builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'`).

**STEP-3 — Move the presentation and the hit row**
Files: `macos/App/Features/Workspace/OmnibarHitPresentation.swift` → `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` (`git mv`), `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` *(new)*, `macos/App/Features/Workspace/OmnibarResultsHost.swift`, `macos/ProvenenciaTests/OmnibarResultsModelTests.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: D-7, D-8. Move the pbxproj file reference for the presentation from the `Workspace` group to `Search` and update its path; add the row file. Rename `OmnibarHitPresentation` → `CatalogSearchHitPresentation` at every use (`rg -n OmnibarHitPresentation macos` must print nothing after).
Check: `rg -n "OmnibarHitPresentation|OmnibarHitRowView" macos` prints nothing; the app and tests build.

**STEP-4 — The omnibar on the shared model**
Files: `macos/App/Features/Workspace/OmnibarResultsModel.swift`, `macos/App/Features/Workspace/OmnibarResultsHost.swift`, `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/App/Features/Workspace/WorkspaceContent.swift`.
Do: D-5, D-6. In the host, panel and toolbar, search state reads go through `results.search` / `omnibarResults.search` (`$omnibarResults.search.query` for the field, `.onChange(of: omnibarResults.search.query)`). `WorkspaceContent` needs no change unless the compiler asks; `activate` stays on `OmnibarResultsModel`.
Check: `rg -n "searchGeneration|debounceTask" macos/App/Features/Workspace` prints nothing; the app builds.

**STEP-5 — Tests**
Files: `macos/ProvenenciaTests/OmnibarResultsModelTests.swift`, `macos/ProvenenciaTests/CatalogSearchTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-9; add T-1 – T-4.
Check: `xcodebuild test … -only-testing:ProvenenciaTests/OmnibarResultsModelTests -only-testing:ProvenenciaTests/CatalogSearchModelTests -only-testing:ProvenenciaTests/CatalogSearchHitPresentationTests -only-testing:ProvenenciaTests/SearchCatalogStoreTests` passes; the case count across the two omnibar-derived files is 14 + the new cases.

**STEP-6 — Docs**
Files: `.agents/skills/add-searchable-kind/SKILL.md`, `docs/deployment-plan/spike-11/promote-find-another-record.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status).
Do: D-10 and the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-37-kind-scoped-search.md` reports no errors; links resolve; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/CatalogSearchTests.swift` `CatalogSearchModelTests.kindsRestrictHits` *(new)* | The Robins store; `CatalogSearchModel(kinds: ["person"])`, query `Robins`, location nil → `hits.map(\.kind) == ["person"]`, first ref is James's handle ref, `searchError == nil` | AC-2.1, D-2 |
| T-2 | Swift | same, `CatalogSearchModelTests.noLocationGivesNoSectionNudge` *(new)* | The Robins store; `CatalogSearchModel()`, query `Robins`, location nil → `["person", "source"]` | AC-2.2, D-3 |
| T-3 | Swift | same, `CatalogSearchModelTests.cancelKeepsResultsAndStopsLoading` *(new)* | Model with `query = "Ilminster"`, `hasSearched = true`, one hit, `isLoading = true`; `cancel()` → `isLoading == false`, query and hits kept, `hasSearched` true | D-6 |
| T-4 | Swift | `macos/ProvenenciaTests/OmnibarResultsModelTests.swift` `closePanelSetsUserDismissedWithoutClearingQuery` (retargeted) | As today, through `model.search` → `userDismissed`, not presented, query kept, `!search.isLoading` | AC-1.1, D-6 |
| T-5 | Swift | `macos/ProvenenciaTests/SearchCatalogStoreTests.swift` `kindPriorityIsATenthAndSectionNudgeStillMatters` (extended) | Same store; `searchCatalog(…, query: "Robins", location: nil, kinds: [])` → `["person", "source"]` | AC-2.3, D-3 |
| T-6 | Swift | every other case listed in D-9, retargeted | Same inputs and expectations as at the planning base | AC-1.1 |

AC-2.4 is pinned by the existing Go test `TestSearchCatalogHandleKinds` (no location sent) and by T-5 for the fake; `GoStore` needs the dylib and is not unit-tested (`add-swift-test`). AC-1.2 is the manual check.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `OmnibarResultsModelTests`, `SearchCatalogStoreTests`, `OnboardingModelTests` (the `ThrowingStore` conformer), and any SwiftUI preview that builds `OmnibarResultsModel` (`WorkspaceToolbar` line 592).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (this PR adds none)
- [ ] UI matches today's omnibar; no new kit primitive
- [ ] Version bump: none
- [ ] rg -n "OmnibarHitPresentation|OmnibarHitRowView" macos prints nothing
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] add-searchable-kind and promote-find-another-record.md updated per D-10
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with at least one Source titled `Ilminster…` and one promoted person. On the Sources list, press ⌘K and type `Ilminster`. See the panel with the Source first, its type badge and thumbnail.
2. Press ↓ and ↑: the selection moves and clamps. Press Esc: the panel closes and the field keeps `Ilminster`. Type one more letter: the panel opens again with results.
3. Press Return on the Source hit: the Source page opens and the field clears.
4. Type the person's surname: the person hit shows its sorted name and life line. Type `zzqx`: the "no matches" state shows.
5. With VoiceOver on, a hit reads title, kind, subtitle, ref and match context as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `$omnibarResults.search.query` does not bind through `@Bindable` | `search` is a `let` reference to an `@Observable` class, so the key path is reference-writable; if the compiler refuses, stop (below) rather than adding a forwarding `query` |
| The `closePanel` change (D-6) drops a result someone relied on | T-3, T-4 and manual step 2 |
| pbxproj edits break the build | Copy the four objects of an existing file as `add-swift-test` describes; build after each step |

**Stop and report if:**

- `OmnibarResultsModel`, `OmnibarResultsHost.swift` or `GenealogyStore.searchCatalog` differs from §4 at your base.
- Any retargeted case in D-9 needs a changed input or expectation to pass.
- `$omnibarResults.search.query` or `.onChange(of: omnibarResults.search.query)` does not compile or does not fire on typing (D-5 would need forwarding).
- A fourth `GenealogyStore` conformer appears that §4 does not list.
- The omnibar's ranking changes for a query on any section (AC-2.2 would then be wrong about today's order).

## 12. Open questions

None.

## 13. PR

**Title:** S11-37 — One kind-scoped catalog search
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-37-kind-scoped-search.md`.
