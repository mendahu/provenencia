# Review — S11-37 One kind-scoped catalog search

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

None.

## Writer's own calls

- The shared home is `macos/App/Features/Catalog/Search/`, with type names `CatalogSearchModel`, `CatalogSearchHitPresentation` and `CatalogSearchHitRow` (D-1, D-7, D-8). The map left the home open.
- `GenealogyStore.searchCatalog`'s `location` becomes optional at the protocol, and a nil location is sent to Go unset (D-3). Promote could have passed its own location instead, but that would nudge ranking by section.
- `closePanel()` now cancels the in-flight search instead of only clearing `isLoading` (D-6). The researcher cannot see the difference.
- The kinds are fixed when the model is created, not passed with each search (D-2).
- `L10n.Workspace.omnibar*` keys and `PVOmnibarHitRow` keep their names (Out).

## Round 1

### X-1 · blocker · `$omnibarResults.search.query` will probably not compile
**Plan:** D-5 and §11 claimed the binding is reference-writable through `@Bindable` when `search` is a `let`.
**Codebase:** `WorkspaceToolbar.swift` line 35 `@Bindable var omnibarResults`, line 167 `text: $omnibarResults.query`. `Bindable`'s dynamic-member subscript needs a `ReferenceWritableKeyPath` at `.search`.
**Answer:** Fixed. D-5 picks `PVInput(text: Bindable(omnibarResults.search).query, …)`. The §11 risk row and the stop condition are reworded to match.

### X-2 · major · The "Robins store" setup has no home
**Answer:** Fixed. D-9 extracts `robinsSearchStore(projectDir:)` at file scope in `SearchCatalogStoreTests.swift`. The original test calls it and keeps its expectations. T-1, T-2 and T-5 call it. T-5 is now its own `@Test`, `searchWithoutLocationHasNoSectionNudge`.

### X-3 · major · The Xcode checks need macOS
**Answer:** Fixed. §7 opens by saying every step runs on macOS with Xcode. A new stop condition says: with no `xcodebuild`, report which checks ran, and never claim PASS.

### X-4 · minor · Small choices left open
**Answer:** Fixed.
- `activateFieldHitFromSearch` keeps its inline loop (D-9).
- `OmnibarResultsModel.minQueryLength` and `debounceNanoseconds` are deleted (D-5), and `isPresented` uses `search.showsResults`.
- `completed.md` gets its paragraph at the top (STEP-5).
- The exact wording for the `add-searchable-kind` and feature-doc edits is in STEP-5.
- T-3's hit is spelled out.
- The pbxproj procedure and the `Search` group shape are in the §7 preamble.

### X-5 · minor · Status and map link
**Answer:** Fixed. The map row links the plan, and the status is set after review.

### M-1 · major · The rule for when results show stays in the omnibar
**Answer:** Fixed. `CatalogSearchModel.showsResults` *(new)* (D-2), and `isPresented` is `!userDismissed && search.showsResults` (D-5). New test T-4.

### M-2 · major · The three-argument extension would become test-only code
**Answer:** Fixed. D-4 now deletes it. Its ten test callers add `kinds: []`, and the protocol doc says what an empty `kinds` and a nil location mean. The matching Out line is removed.

### M-3 · major · `settle()` would be a second polling helper
**Answer:** Fixed. D-9 adds `waitForSearch(_:)` in `QueryHandleWaiting.swift` as a one-line wrapper over `waitUntil`, and drops the private 25 ms loop.

### M-4 · minor · Moved comments and names still say "omnibar"
**Answer:** Fixed. D-5 rewrites the class comment and D-7 the presentation's two comments. D-4 rewrites the protocol doc. D-3 renames `FakeStore.omnibarSearchCatalog` to `vocabularySearchHits`, and the DoD `rg` checks that the old name is gone.

### M-5 · minor · The lead image is chosen in two kind switches
**Answer:** Logged in D-8. The split moves as it is, and is consolidated when a hit kind is next added.

### M-6 · minor · The Out reason misquoted the idea file
**Answer:** Fixed. The Out reason is now "no shared popup presenter exists yet". The map's S11-36 scope gains a line asking the picker not to add another hand-rolled Escape and outside-click monitor.

### M-7 · nit · STEP-2 would have committed two copies of the search state
**Answer:** Fixed. The steps are reordered: STEP-3 creates `CatalogSearchModel` and moves the omnibar onto it in one commit.
