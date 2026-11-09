# S16-07 — One paged fold that moves focus

**Status:** Blocked — on board S16-D2
**Spike / slice:** Spike 16 · Slice 1a — Bounded reads, folds and the Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S16-D2](../design/S16-D2-sources-fold-highlight.md), board not agreed yet (the Sources Show more's highlight, FH-1, FH-2; the page count, FH-3)
**Board:** — (recorded here when S16-D2 is agreed)
**Depends on:** S16-D2
**Blocks:** S16-02, S16-27; Spike 15's S15-30
**Branch:** `s16-07-paged-fold-focus`
**Version bump:** patch: a Mac UI fix (long lists unfold a page at a time; Show more keeps the keyboard) with no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md) (no new key; read it to keep `conclusions.showMore` as it is)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-7 (pages), P-14, P-15; working rules "(Batch 1) One paged fold that moves focus", "(Batch 1) The Timeline builds what it shows" (S16-02's half); [`macos-client-patterns.md`](../../../macos-client-patterns.md) "One sectioned detail page" and § 9 Wide rule 5; Spike 15 rule "The provenance table folds a page at a time" ([`../../spike-15/deployment-plan.md`](../../spike-15/deployment-plan.md)), which moves this PR's fold to the kit later (S15-30)
**Planned:** 2026-11-09 against `main` @ `e22d035`
**Review:** — (not reviewed yet; review-plan runs once the board is agreed and the Scope takes its frames)

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

Every "Show N more" on the Person page reveals one page of at most 50 items, says how many it reveals, and moves the keyboard to the first item it revealed, in Family, Timeline and Sources alike. Today one click reveals the whole list (a crafted person with 2,000 events builds them all at once), the Timeline and Sources move focus with private `Task` hops, and Family drops focus when its Show more disappears (P-14). One fold state type replaces the three sections' `Set`s and Sources' `rowsExpanded`. One focus keeper in `Platform/` holds the keyboard across a Regular/Wide change for Gallery and Timeline. When the record links diagram highlights a graph whose lines are all folded, that Source's Show more takes the highlight (P-15).

**Why now.** Bar 3 and bar 7 of the spike, and S16-02 builds only what these folds show. S15-30 then moves this fold to the kit, so it must be one type before then.

**What stays.** The folds' first limits (Family 30, Timeline 50 events and 5 / 12 participants, Sources 5 lines and 50 Sources; user, 2026-11-03), Sources' "six lines show whole" (`lineFoldMinimumHidden` 2), the `conclusions.showMore` copy, every row's and card's layout, the lifeline drawing every point, the Gallery's tiles. The Timeline still builds every row (S16-02).

## 2. User stories

**US-1.** As a researcher, when I press Show more on a long list, I see the next 50 items and the keyboard is on the first of them, so that a long list neither stalls the page nor loses my place.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Timeline with 2,000 dated events, folded | 50 rows, then "Show 50 more"; after one press 100 rows and "Show 50 more"; after 39 presses all 2,000 rows and no Show more |
| AC-1.2 | Sources: one Source with 120 lines | 5 lines, "Show 50 more"; press → 55, "Show 50 more"; press → 105, "Show 15 more"; press → 120, none (FH-3) |
| AC-1.3 | Sources: one Source with 56 lines | 5 lines, "Show 51 more"; press → 56, none (a fold never leaves one line) |
| AC-1.4 | Family: 100 children with one partner (the cap, `kinship.Rules.MaxKinPerRole`) | 30 children, "Show 50 more"; press → 80, "Show 20 more"; press → 100 |
| AC-1.5 | Timeline: an event with 500 participants at Regular | 5 cards, "Show 50 more"; press → 55 cards |
| AC-1.6 | Any section above: press Show more with the keyboard | Focus is on the first revealed item: the Timeline row `d51`, the card `big-p6`, the Sources graph link of the 6th line, the Family tile of the 31st child |
| AC-1.7 | Family: siblings folded and the person is the 40th sibling (already shown after the fold) | The revealed page skips the person's own tile; focus lands on the first revealed sibling that is not the person |
| AC-1.8 | Timeline: click a lifeline point whose row is on the 3rd hidden page | The dated fold opens through that page (200 rows shown, not 2,000) and the list scrolls to the row |

**US-2.** As a keyboard user, when the window crosses the Regular/Wide line, the keyboard stays where it was, so that a resize never strands me.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Timeline at Wide, focus on the 9th participant card of an event; resize to Regular | That card is folded away (5 shown); focus is on the event's title |
| AC-2.2 | Timeline at Wide, focus on the 3rd card; resize to Regular | Focus stays on the 3rd card |
| AC-2.3 | Gallery, focus on a tile's image; resize across the line | The tile scrolls into view and its image has focus again |
| AC-2.4 | Gallery, focus on Use as portrait of a tile that became the lead meanwhile | Focus goes to that tile's image (today's rule, kept) |
| AC-2.5 | Click elsewhere (focus goes to nil), then resize | Focus is not pulled back |

**US-3.** As a researcher, when I point at a node of the record links diagram whose record lines are folded, the Show more that hides them lights up, so that I can tell which Source it is on (P-15).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A Source with 9 lines on 9 graphs, folded; highlight the graph of the 7th line | That Source's Show more is highlighted; no line is (FH-1) |
| AC-3.2 | Same, highlight the graph of the 3rd line | The 3rd line is highlighted; the Show more is not |
| AC-3.3 | A graph with two of the person's records, one shown and one folded | Only the shown line is highlighted |
| AC-3.4 | 60 Sources, highlight a graph on the 55th | The section's Show more under the 50th Source is highlighted (FH-2) |
| AC-3.5 | Highlight goes (pointer leaves the node) | No Show more is highlighted |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ConclusionFold` pages: `pageSize` 50, `shown(…, pages:)`, `nextPageCount`, `firstRevealed(…, pages:)`, `pages(toShow:)` | STEP-1, AC-1.1 – AC-1.5 |
| `ConclusionFoldState<Fold>`: one fold state type in place of `Set<Fold>` (Family, Timeline), `Set<String>` + `rowsExpanded` (Sources) | STEP-1, STEP-3 – STEP-5 |
| `ConclusionShowMoreRow` words the page (FH-3), claims focus for the revealed item through `FocusClaim`, and draws a highlighted state (FH-1, FH-2) | STEP-2, AC-1.6, AC-3.x |
| Family: tiles take focus; content names the first revealed tile; Show more moves focus (P-14) | STEP-3, AC-1.4, AC-1.6, AC-1.7 |
| Timeline: pages for dated, undated and participants; `Task` hop goes; a folded point opens through its page | STEP-4, AC-1.1, AC-1.5, AC-1.8 |
| Sources: pages for lines and rows; `Task` hops go; folded-graph highlight on both Show mores (P-15, FH-1, FH-2) | STEP-5, AC-1.2, AC-1.3, AC-3.x |
| `FocusKeeper` in `Platform/`: Gallery and Timeline keep focus across a width change (Wide rule 5) | STEP-6, STEP-7, US-2 |
| FH-W1: the highlight works the same at Regular and Wide | STEP-5, manual check |
| Docs: `macos-client-patterns.md` fold and keeper lines | STEP-8 |

| Out | Home |
| --- | --- |
| Building only shown Timeline rows and cards; click targets only for shown rows | S16-02 |
| The fold and `PVShowMoreRow` in the kit | S15-30 |
| Back restoring a section's folds, scroll and focus (P-19) | [person page follow-ups idea](../../../ideas/person-page-followups.md) |
| Memoising section content | S16-27 |
| Hovering a Show more highlighting nodes; a count of the highlighted graph's folded lines | not planned (S16-D2 §2.1) |
| The Gallery's own fold (it has none) and the Event and Place pages' lists | not planned |
| The Sources rows' narrow arrangement | [wide layout idea](../../../ideas/wide-layout.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Fold rule | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` `ConclusionFold.shown(_:limit:minimumHidden:expanded:)`, `firstRevealed(_:limit:minimumHidden:)` | Folded or all; no page. `firstRevealed` is `list[limit]` |
| Show more | same file, `ConclusionShowMoreRow(count:expand:)` | Ghost small `PVButton` "Show N more" with the total folded; no focus, no highlight |
| Focus claim | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)` | Writes on the next turn and again a turn later. 10 callers |
| Family state | `PersonFamilySection.swift` `@State expanded: Set<PersonFamilyContent.Fold>`; `PersonFamilyDiagram.expand` | Inserts the fold; no focus anywhere in the diagram (P-14: focus is lost) |
| Family content | `PersonFamilyContent.swift` `init(family:expanded:locale:orders:)`, `foldLimit` 30, `Fold` (`parentSets`, `siblings`, `partners`, `children(Int)`), `*Hidden` counts | Folds each list; appends the person to siblings when past the fold |
| Family layout | `PersonFamilyLayout.swift` `PlacedMore(fold:count:frame:)` | Places one more row per folded list with the `*Hidden` count |
| Timeline state | `PersonTimelineSection.swift` `PersonTimelineSection.expanded: Set<PersonTimelineContent.Fold>`; `PersonTimelineList.unfold` (`Task { @MainActor in focused = revealed }`), `select` (`expand(.dated)` for a folded point), `focusedOwner`, `lastFocused`, `keepFocus()` | Private focus hop; private width-change keeper; a folded point opens every dated row |
| Timeline content | `PersonTimelineContent.swift` `init(timeline:expanded:locale:orders:widthClass:)`, `eventFoldLimit` 50, `participantFoldLimit` 5 / 12, `revealed`, `firstRevealed(by:)`, the hidden-run placement (lines 186–204) | Folds dated rows past 50 in one `.more(.dated, count:)`; builds every row (S16-02's) |
| Sources state | `PersonSourcesSection.swift` `expanded: Set<String>`, `rowsExpanded: Bool`; `PersonSourcesList.unfold` and the rows' Show more (`Task { @MainActor in focus = … }`) | Two `Task` hops; lines' and rows' Show mores never highlight |
| Sources highlight | `PersonSourcesSection.swift` `PersonSourcesRowView.lines` (`highlightedGraphID = hover.id ?? focusedGraphID`), `PersonSourcesLineView` background `PVColor.surfaceSelected`, `PVRadius.sm` | Only shown lines highlight (P-15) |
| Sources content | `PersonSourcesContent.swift` `init(groups:sourceTypes:orders:expanded:rowsExpanded:)`, `Row.hiddenCount`, `firstRevealedLineID`, `hiddenRowCount`, `firstRevealedRowLineID`, `claimGraphs` | Folds lines at 5 (min hidden 2) and rows at 50 |
| Gallery focus | `ConclusionGallerySection.swift` `lastFocused`, `pendingFocus`, `restoreFocus(to:)`, `applyPending()`, tile `onAppear` | Private keeper for width change and viewer close; 400 ms expiry |
| Tests | `macos/ProvenenciaTests/ConclusionFoldTests.swift` `foldRule`; `PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`, `PersonTimelineContentTests.swift`, `PersonTimelineLifelineTests.swift`, `PersonSourcesContentTests.swift`, `PersonRecordLinksContentTests.swift` | All pass `expanded:`; each changes to `folds:` |
| Previews | `PersonTimelineSection.swift` `PersonTimelinePreview.expandedAtStart`; `PersonSourcesSection.swift` `PersonSourcesPreview.expandedAtStart`, `rowsExpanded` | Debug-only; change with the types |

**Precedent.** `6401628` ("Show the Sources graph link's focus, open it on Return…") is the Sources focus shape; S13-17 (#729) added `FocusClaim` and its use in `PersonSourcesList.applyPendingFocus`: copy that call (`FocusClaim.claim($focus, value)`) for every Show more. Do not copy the `Task { @MainActor in focus = … }` hops.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A page is 50 items in every fold, after the fold's unchanged first limit: shown = `min(count, limit + pages × 50)`, and when what would stay hidden is less than `minimumHidden`, everything shows. `ConclusionFold.pageSize = 50` | Working rule; one rule for every section | A page equal to each fold's first limit (5 cards per click is too slow) |
| D-2 | The Show more label counts the next page: `nextPageCount(hidden:minimumHidden:)` = `hidden` when `hidden - 50 < minimumHidden`, else 50. The copy key stays (user, 2026-11-09) | FH-3 | The total folded (user chose the page) |
| D-3 | `ConclusionShowMoreRow` takes `hidden` and `minimumHidden` and words the page itself; callers keep passing the folded total | One place does FH-3; `PersonFamilyLayout`'s `PlacedMore.count` and every `*Hidden` keep meaning "folded" | Each content type computing a page count |
| D-4 | `ConclusionFoldState<Fold: Hashable>` *(new)*, in `ConclusionShowMoreRow.swift` beside `ConclusionFold`: `pages(_:) -> Int`, `mutating revealPage(_:)`, `mutating reveal(_:pages:)` (keeps the larger). Each section holds one in `@State`; content inits take `folds:` in place of `expanded:` | Working rule "one fold state type"; S15-30 moves this file's types to `Platform/` | A dictionary per section; keeping `Set` beside a page count |
| D-5 | Sources' folds are `PersonSourcesContent.Fold` *(new)*: `.rows`, `.lines(sourceID: String)` | One state type for both of its folds | Keeping `rowsExpanded` |
| D-6 | `firstRevealed` names the first item the next Show more adds, at the current pages. The section reads it from the content it drew, before revealing | Same contract as today; the content after the press has moved on | Computing it after the press |
| D-7 | `ConclusionShowMoreRow` becomes generic over the focus value: `ConclusionShowMoreRow<Focus: Hashable>(hidden:minimumHidden:highlighted:focus:revealed:expand:)`; its action runs `expand()`, then `FocusClaim.claim(focus, revealed)` when `revealed` is not nil | Working rule: the row claims focus through `FocusClaim`; the `Task` hops go | A closure per caller that moves focus |
| D-8 | Family gains `@FocusState focused: String?` in `PersonFamilyDiagram`; each relative's tile gets `.focused($focused, equals: placed.id)`. `PersonFamilyContent.firstRevealed(by:) -> String?` *(new)* names a tile id, skipping the person's own tile (it is not a button) | P-14; AC-1.7 | Focusing the group container |
| D-9 | A lifeline click on a folded point opens the dated fold through that row's page: `PersonTimelineContent.pages(toShow:) -> Int?` *(new)* and `folds.reveal(.dated, pages:)` (user, 2026-11-09) | AC-1.8 | Opening every row (user chose the page) |
| D-10 | `FocusKeeper<Stop: Hashable>` *(new)*, `macos/App/Platform/FocusKeeper.swift`, a value type held in `@State`, with the view modifier `focusKeeper(_:focus:fallback:reveal:)` *(new)* that wires `onChange(of: focus)` and `onChange(of: pvWidthClass)` and claims through `FocusClaim`. State: `last` (the last stop, kept one main-actor turn after focus goes nil), `lastFallback`, `pending` (the stop being given back, cleared 400 ms after a claim). On a width change it takes `focus ?? last`; `reveal(stop)` scrolls and returns the stop to claim (the stop itself when it is still shown, else the fallback recorded when it took focus) | Wide rule 5 from one place; Gallery's 400 ms expiry and Timeline's owner fallback become its two parameters | A `DesignSystem` component (it holds no chrome; working rule names `Platform/`); an `@Observable` class (a `FocusState` binding is only valid at call time) |
| D-11 | Gallery's viewer-close restore uses the same keeper (`keeper.restore(stop)`); the tile `onAppear` re-claim reads `keeper.pending`; the Use as portrait → image rule stays in Gallery's `reveal` closure | One keeper; no second pending state in Gallery | Keeping Gallery's private `pendingFocus` |
| D-12 | Timeline's fallback for a card is its row's id, recorded when the card takes focus (`content.row(holdingCard:)` *(new)*); `focusedOwner`, `lastFocused` and `keepFocus()` go | After the width change the card is no longer in the content | Parsing the card id |
| D-13 | Sources highlight: `PersonSourcesContent.Row.foldedGraphIDs: Set<String>` *(new)* = graphs with a folded line and no shown line in that row; `PersonSourcesContent.foldedRowGraphIDs: Set<String>` *(new)* = graphs on Sources past the row fold. A Show more is highlighted when its set contains `hover.id ?? focusedGraphID` | AC-3.1 – AC-3.4; content already maps claims to graphs | Searching the folded lines in the view on every hover |
| D-14 | The highlighted Show more look is the agreed board's FH-1; until then the plan assumes the recommended `PVColor.surfaceSelected` background with `PVRadius.sm` behind the row, as `PersonSourcesLineView` draws | S16-D2 recommendation | A new token or a Badge |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionFoldState` *(new)* | type | `Set<PersonFamilyContent.Fold>`, `Set<PersonTimelineContent.Fold>`, `Set<String>` + `rowsExpanded` | Replaces all four; a `Set` cannot hold a page count |
| `ConclusionFold.pageSize`, `nextPageCount`, `pages(toShow:)` *(new)* | func | `ConclusionFold.shown`, `firstRevealed` | Extends the one fold rule (searched: `prefix(`, `Show more`, `limit:` in `Features/`) |
| `PersonSourcesContent.Fold` *(new)* | type | `PersonFamilyContent.Fold`, `PersonTimelineContent.Fold` | Same shape per section; Sources had none |
| `FocusKeeper`, `focusKeeper(_:focus:fallback:reveal:)` *(new)* | type + modifier | Gallery's `restoreFocus`/`applyPending`, Timeline's `keepFocus` | Both are deleted onto it; `FocusClaim` stays the writer it calls (searched: `widthClass) {`, `lastFocused`, `pendingFocus` in `macos/App`) |
| `PersonFamilyContent.firstRevealed(by:)`, `PersonTimelineContent.row(holdingCard:)`, `pages(toShow:)` *(new)* | func | `PersonTimelineContent.firstRevealed(by:)`, `isShown` in `PersonTimelineList` | Content owns ids; the views ask it |
| `Row.foldedGraphIDs`, `foldedRowGraphIDs` *(new)* | property | `claimGraphs`, `graphLocations` | Built in the same pass over claims |
| `FocusKeeperTests` *(new)* | test file | `ConclusionFoldTests` | New type |

**Builds on.** `FocusClaim` (S13-17), `ConclusionFold` ("the one fold rule"), `CircularLinkHover` (the one hovered graph).
**Deletes.** `ConclusionFold.shown(…, expanded:)` and `firstRevealed(_:limit:minimumHidden:)` (replaced by the `pages:` forms); `PersonTimelineList.focusedOwner`, `lastFocused`, `keepFocus()`, `isShown(_:)` (moved to content or the keeper); `ConclusionGallerySection.lastFocused`, `pendingFocus`, `restoreFocus(to:)`, `requestScroll()`; the three `Task { @MainActor in focus… = … }` hops; `PersonSourcesSection.rowsExpanded`, `PersonSourcesList.expandRows`; the `expanded:` init parameters and preview `expandedAtStart` sets (become `ConclusionFoldState` values).
**Next adds.** S16-02 reads `folds` to build only shown rows; S15-30 moves `ConclusionFold`, `ConclusionFoldState` and the row to `Platform/` and the kit; S15-27's provenance table folds with them; any later width-sensitive view adopts `focusKeeper`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A fold with `pages > 0` and a list that shrinks after a reload | `ConclusionFold.shown` | `min(count, …)`: a page past the end shows the whole list, no crash; T-1 |
| Another person's page | each section's `@State folds` | The sections are new views per entity id, as their `Set`s are today: folds start empty |
| A revealed item gone before the claim lands | `FocusClaim.claim` | Writes a value with no focusable view; SwiftUI leaves focus nil, as the Sources `applyPendingFocus` path does today |
| A highlighted graph that is not on the page | `foldedGraphIDs` | No set contains it; nothing highlighted (T-9) |

### Scale

At the spike's crafted Scale (2,000 events, 500 participants on one event, a Source with hundreds of lines): `shown`, `nextPageCount`, `firstRevealed`, `pages(toShow:)` are O(1) arithmetic plus one `prefix`; `pages(toShow:)` is O(n) once per click (an index lookup). `foldedGraphIDs` is one pass over the claims already walked in `PersonSourcesContent.init`. The Timeline still builds every row until S16-02; this PR adds no work per row. Real data (200 events, 50 graphs) never passes a first limit except Timeline's participants, as today.

### Edges

- **Assumptions.** `firstRevealed` assumes items keep their order between the draw and the press (Go's order; true for one content value). The Timeline's hidden-run placement assumes the hidden rows are a suffix of `counted`; with pages it is still a suffix (T-4). Family's siblings fold assumes the person is appended when past the shown prefix; with pages that prefix grows, so the person is appended only while still hidden, and is never revealed twice (T-6).
- **Neighbors.** `minimumHidden` 2 on Sources lines: 6 lines → whole; 7 → 5 + "Show 2 more"; 55 → 5 + "Show 50 more" then 55; 56 → "Show 51 more"; 57 → "Show 50 more" then "Show 2 more" (T-1). Participants at Regular 5 vs Wide 12 with the same `pages` (T-5).
- **Removed inputs.** A reload that drops items keeps `pages`; `shown` caps at `count` (T-1). An unpromoted line inside a revealed page: `lineAfterRemoving` reads shown lines only, unchanged.
- **Real data.** None — no table, seed or configuration drives the fold.
- **Untrusted size.** The folds bound what is drawn per click to 50; the lists come from Go reads S16-01 capped (100 relatives per role, 10 records per graph). Timeline events are not capped in Go; S16-02 bounds what is built.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Show more row | `ConclusionShowMoreRow` (snowflake, Extend) | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` | Ghost small `PVButton`, as today; `highlighted` draws D-14's background; label `L10n.Conclusions.showMore(nextPageCount)` |
| Record line highlight | `PersonSourcesLineView` (Ship) | `PersonSourcesSection.swift` | Unchanged; its tokens are FH-1's look |
| Focus keeper | `FocusKeeper` *(new)* | `macos/App/Platform/FocusKeeper.swift` | No chrome |

```swift
struct ConclusionFoldState<Fold: Hashable>: Equatable {
    func pages(_ fold: Fold) -> Int
    mutating func revealPage(_ fold: Fold)
    mutating func reveal(_ fold: Fold, pages: Int)
}

enum ConclusionFold {
    static let pageSize = 50
    static func shown<T>(_ list: [T], limit: Int, minimumHidden: Int, pages: Int) -> (shown: [T], hidden: Int)
    static func nextPageCount(hidden: Int, minimumHidden: Int) -> Int
    static func firstRevealed<T>(_ list: [T], limit: Int, minimumHidden: Int, pages: Int) -> T?
    /// The fewest pages that show `list[index]`; 0 when the first limit shows it.
    static func pages(toShow index: Int, count: Int, limit: Int, minimumHidden: Int) -> Int
}

struct FocusKeeper<Stop: Hashable>: Equatable {
    private(set) var last: Stop?
    private(set) var pending: Stop?
    mutating func focusChanged(to stop: Stop?, fallback: Stop?)
    /// The stop to give back on a width change, or nil when nothing had the keyboard.
    mutating func widthChanged(current: Stop?) -> (stop: Stop, fallback: Stop?)?
    mutating func restore(_ stop: Stop)
    mutating func expire(_ stop: Stop)
}
```

States: an empty section has no fold (unchanged). Accessibility: the Show more's label is spoken as drawn ("Show 50 more"); the highlight is not spoken (S16-D2 §2.1); focus moves are the VoiceOver cursor's moves.

### Strings

None — `conclusions.showMore` ("Show %lld more") is unchanged; only its number changes (D-2). No copy names the old "show all" scope.

## 7. Steps

**STEP-1 — The paged fold and its state**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, `macos/ProvenenciaTests/ConclusionFoldTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-4: add `pageSize`, the `pages:` forms, `nextPageCount`, `pages(toShow:count:limit:minimumHidden:)` and `ConclusionFoldState`; keep the `expanded:` forms until STEP-5 removes their last caller.
Check: T-1, T-2 pass (`xcodebuild test -scheme Provenencia -only-testing:ProvenenciaTests/ConclusionFoldTests`).

**STEP-2 — The Show more row words the page, claims focus and highlights**
Files: `ConclusionShowMoreRow.swift`. Skill: `add-ui-component`.
Do: D-3, D-7, D-14. Update the four call sites to the new init with today's behavior (`revealed: nil`, `highlighted: false`) so the tree builds.
Check: the app builds; the Family, Timeline and Sources previews show "Show 50 more" on a list longer than its first limit plus 50.

**STEP-3 — Family pages and moves focus**
Files: `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `PersonFamilyLayout.swift` (only if `PlacedMore` needs the fold's revealed id), `macos/ProvenenciaTests/PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`.
Do: D-4, D-8: `folds: ConclusionFoldState<Fold>`, `firstRevealed(by:)`, `.focused` on relative tiles, Show more passes `revealed` and `$focused`.
Check: T-6 passes; existing Family tests pass with `folds:`.

**STEP-4 — Timeline pages and moves focus**
Files: `PersonTimelineSection.swift`, `PersonTimelineContent.swift`, `macos/ProvenenciaTests/PersonTimelineContentTests.swift`, `PersonTimelineLifelineTests.swift`.
Do: D-4, D-6, D-7, D-9: `folds:` in the init; the dated hidden-run placement reads `pages(.dated)`; `unfold` keeps `pendingReveal` and drops the `Task`; `select` uses `pages(toShow:)`; `row(holdingCard:)`.
Check: T-3, T-4, T-5, T-7 pass; `PersonTimelineLifelineTests` pass unchanged in meaning (every point still drawn).

**STEP-5 — Sources pages, moves focus and highlights a folded graph**
Files: `PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`, `PersonRecordLinksContentTests.swift`.
Do: D-5, D-7, D-13: one `@State folds`; `expandRows` goes; both Show mores pass `revealed`, `$focus` and `highlighted`. Remove `ConclusionFold`'s `expanded:` forms (no caller left).
Check: T-8, T-9 pass; `rg "expanded: Bool|rowsExpanded|Task \{ @MainActor in focus" macos/App/Features/Conclusions` finds nothing.

**STEP-6 — The focus keeper**
Files: `macos/App/Platform/FocusKeeper.swift` *(new)*, `macos/ProvenenciaTests/FocusKeeperTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-10.
Check: T-10 passes.

**STEP-7 — Gallery and Timeline keep focus through the keeper**
Files: `ConclusionGallerySection.swift`, `PersonTimelineSection.swift`.
Do: D-11, D-12; delete what **Deletes** names for these two files.
Check: manual check 4 and 5; `rg "pendingFocus|lastFocused|keepFocus" macos/App/Features/Conclusions/ConclusionGallerySection.swift macos/App/Features/Conclusions/PersonTimelineSection.swift` finds nothing.

**STEP-8 — Docs**
Files: `docs/macos-client-patterns.md` ("One sectioned detail page": one sentence on the paged fold, `ConclusionFoldState` and Show more's focus claim; under § 9's six rules, a sentence after the list naming `FocusKeeper` for rule 5 — the rules' text stays word for word, since `add-design-brief` repeats it), `completed.md`, `deployment-plan.md` (map row), this plan (status), the brief S16-D2 → the spike's design archive, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-07-paged-fold-focus.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Unfold long Person page lists fifty at a time and keep the keyboard on the first revealed item.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `ConclusionFoldTests.swift` `pagedFold` *(new)* | 120 items, limit 5, min 2: pages 0/1/2/3 → shown 5/55/105/120, `nextPageCount` 50/50/15/—; 56 → 5 then "51"; 7 → "2"; 6 → whole; pages 9 on 20 items → all 20 | AC-1.2, AC-1.3, D-1, D-2 |
| T-2 | Swift | `ConclusionFoldTests.swift` `foldStateAndPagesToShow` *(new)* | `revealPage` twice → 2; `reveal(pages: 1)` after 2 → 2; `pages(toShow: 49, limit 50)` → 0, 50 → 1, 199 → 3 | D-4, D-9 |
| T-3 | Swift | `PersonTimelineContentTests.swift` `twoThousandEventsUnfoldInPages` *(new)* | 2,000 dated events: each of 39 reveals adds 50 rows and `firstRevealed(by: .dated)` names `d(50k+1)`; after 39 no `.more` | AC-1.1, AC-1.6 |
| T-4 | Swift | `PersonTimelineContentTests.swift` `foldedMarkersStayAfterPagedRun` *(new)* | 120 dated rows with a placeholder death among the hidden: one page → `.more(.dated)` sits after row 100, the marker still after it | D-1, D-6 |
| T-5 | Swift | `PersonTimelineContentTests.swift` `participantsPageAtBothWidths` *(new)* | 500 participants, `pages 1`: Regular 55 cards, Wide 62; `firstRevealed` names card 56 / 63 | AC-1.5 |
| T-6 | Swift | `PersonFamilyContentTests.swift` `childrenAndSiblingsPage` *(new)* | 100 children → 30/80/100 and `firstRevealed(.children(0))` = 31st tile id; 60 siblings with the person 40th → page 1 shows 60 with the person once, `firstRevealed(.siblings)` skips the person | AC-1.4, AC-1.6, AC-1.7 |
| T-7 | Swift | `PersonTimelineContentTests.swift` `pointOnFoldedPageOpensThroughIt` *(new)* | 2,000 events, row `d180` → `pages(toShow:)` 3; with it, `d180` is shown and `d201` is not | AC-1.8 |
| T-8 | Swift | `PersonSourcesContentTests.swift` `linesAndRowsPage` *(new)* | 120 lines on one Source; 60 Sources: pages as AC-1.2 and rows 50 → 60; `firstRevealedLineID` / `firstRevealedRowLineID` per page | AC-1.2, AC-1.3, AC-1.6 |
| T-9 | Swift | `PersonSourcesContentTests.swift` `foldedGraphsTakeTheHighlight` *(new)* | 9 lines on 9 graphs: `foldedGraphIDs` = graphs of lines 6–9; a graph with lines 3 and 7 is not in it; 60 Sources: `foldedRowGraphIDs` holds the 51st–60th Sources' graphs; an unknown graph in neither | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, D-13 |
| T-10 | Swift | `FocusKeeperTests.swift` `keepsTheStopAcrossAWidthChange` *(new)* | focus "b" with fallback "row" → width change returns ("b", "row"); focus nil then `expire` → width change returns nil; `restore("x")` sets `pending`; `expire("x")` clears it; `expire("y")` leaves it | AC-2.1, AC-2.2, AC-2.5, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonTimelineContentTests`, `PersonTimelineLifelineTests`, `PersonFamilyLayoutTests`, `PersonSourcesContentTests` (their `expanded:` calls).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none added: D-2)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`rowsExpanded`, `expandedAtStart`, `keepFocus`, `restoreFocus(to:)`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the S16-D2 board and its inventory; no new kit primitive
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): 0.4.2 → 0.4.3, or the next patch on main
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief S16-D2 moved to the spike's design archive and design/README.md updated
- [ ] Domain doc updated: docs/macos-client-patterns.md (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use the large fixture or a project with a person who has more than 100 dated events, a Source with more than 10 of their records (one graph each), and an event with more than 60 people.

1. Timeline, Regular: Tab to "Show 50 more" under the 50th row, press Return. 50 rows appear, the keyboard is on the 51st row's title, the list scrolls to it. Bar 3.
2. Same on an event's participants and on Family's children (30 shown): the keyboard lands on the first revealed card or tile. Bar 7.
3. Sources: point at a node whose graph's line is folded: the Source's Show more lights up (FH-1); at Wide the node and the row are both on screen (FH-W1). With more than 50 Sources, a node on a folded Source lights the section's Show more (FH-2).
4. Timeline at Wide, focus the 9th card of an event, narrow the window below 1600: the keyboard is on the event's title. Widen again with the 3rd card focused: it keeps the keyboard.
5. Gallery: focus a tile, cross the width line both ways: the tile is in view with the keyboard (AC-2.3). Focus a tile's Use as portrait, make it the lead, cross the line: the keyboard is on its image (AC-2.4). Open and close the likeness viewer: the keyboard returns to the tile.
6. Click a lifeline point on a folded row deep in the list: the list opens through that page only and scrolls to it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `FocusClaim`'s two-turn write lands before a revealed row exists | It writes twice by design; the Sources `applyPendingFocus` path already relies on it. T-3 covers the id; manual check 1 the landing |
| The keeper changes the Gallery's lazy-grid timing | D-11 keeps the tile `onAppear` re-claim and the 400 ms expiry; manual check 5 |
| Family tiles gaining focus changes Tab order | Groups are already containers in reading order (`PersonFamilyDiagram` comment); manual check 2 |

**Stop and report if:**

- The agreed S16-D2 board draws FH-1 with anything other than D-14's tokens, or adds a frame this plan has no AC for.
- `FocusClaim.claim` does not land on a revealed Timeline row after two turns in manual check 1 (needs a different claim path; do not add a third write).
- `PVPersonCard` does not take `.focused` for a relative's tile (`PVTileCard` with an `action`).
- `PersonFamilyLayout` needs to place a Show more differently because of pages (D-3 says only its count changes).
- A test outside §8 must change its expected output for reasons other than `expanded:` → `folds:`.

## 12. Open questions

- **Board.** S16-D2 is not agreed. How does the agreed board reach the implementer (exported frames in the repo, screenshots, or a link they can open)? Recommendation: export frames 1a/1b, 3a/3b and 4 into `design/` beside the brief.

## 13. PR

**Title:** S16-07 — One paged fold that moves focus
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-07-paged-fold-focus.md`.
