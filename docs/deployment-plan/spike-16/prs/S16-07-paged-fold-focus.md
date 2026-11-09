# S16-07 — One paged fold that moves focus

**Status:** Planned
**Spike / slice:** Spike 16 · Slice 1a — Bounded reads, folds and the Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S16-D2](../design/archive/S16-D2-person-page-folds.md), board agreed 2026-11-09 (one brief for the Person page's Family, Timeline and Sources folds: FH-1 – FH-5, FH-W1, FH-W2)
**Board:** [Claude Design — Person Detail](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ), turn "S16-D2 · Folds" on three pages: *Person Detail - Family* (frames F5a, F5b), *Person Detail - Timeline* (F6a, F6b, F7a, F7b, F8, F9) and *Person Detail - Sources* (F1a, F1b, F1 narrow pane, F2a, F2b, F3a, F3b, F4). The implementer reads it with the Artifact tool's `read` on that link. Each turn's decisions panel restates FH-1 – FH-5 and FH-W2
**Depends on:** S16-D2 (board agreed 2026-11-09)
**Blocks:** S16-02, S16-27; Spike 15's S15-30
**Branch:** `s16-07-paged-fold-focus`
**Version bump:** patch: a Mac UI fix (long lists unfold a page at a time; Show more keeps the keyboard) with no wire or catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md) (no new key; read it to keep `conclusions.showMore` as it is)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-7 (pages), P-14, P-15; working rules "(Batch 1) One paged fold that moves focus", "(Batch 1) The Timeline builds what it shows" (S16-02's half); [`macos-client-patterns.md`](../../../macos-client-patterns.md) "One sectioned detail page" and § 9 Wide rule 5; Spike 15 rule "The provenance table folds a page at a time" ([`../../spike-15/deployment-plan.md`](../../spike-15/deployment-plan.md)), which moves this PR's fold to the kit later (S15-30)
**Planned:** 2026-11-09 against `main` @ `e22d035` (no `macos/` change since, at `55f7f88`)
**Review:** [`S16-07-paged-fold-focus.review.md`](S16-07-paged-fold-focus.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blocker; To user: none

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
| AC-1.9 | Any section: click Show more with the pointer while focus is elsewhere | The page reveals, and focus stays where it was |

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
| `ConclusionShowMoreRow` words the page (FH-3) and claims focus for the revealed item through `FocusClaim` after a key press (FH-4) | STEP-2, AC-1.6, AC-1.9 |
| The Sources Show mores take the record line's highlight (FH-1, FH-2) | STEP-5, AC-3.1 |
| Family: tiles take focus; content names the first revealed tile; Show more moves focus (P-14, FH-4) | STEP-3, AC-1.4, AC-1.6, AC-1.7 |
| Timeline: pages for dated, undated and participants; `Task` hop goes; a folded point opens through its page (FH-5) | STEP-4, AC-1.1, AC-1.5, AC-1.8 |
| Sources: pages for lines and rows; `Task` hops go; folded-graph highlight on both Show mores (P-15, FH-1, FH-2) | STEP-5, AC-1.2, AC-1.3, AC-3.x |
| `FocusKeeper` in `Platform/`: Gallery and Timeline keep focus across a width change (Wide rule 5, FH-W2) | STEP-6, STEP-7, US-2 |
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
| Fold rule | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` `enum ConclusionFold`: `static shown(_:limit:minimumHidden:expanded:)`, `static firstRevealed(_:limit:minimumHidden:)` | Folded or all; no page. Each content passes its `limit` and `minimumHidden` to both, two to four times per fold |
| Show more | same file, `ConclusionShowMoreRow(count:expand:)` | Ghost small `PVButton` "Show N more" with the total folded, `.pvActivatesOnReturn(expand)` (the modifier that owns its own `@FocusState`); no focus move, no highlight. Five call sites: `PersonFamilySection.swift` (1), `PersonTimelineSection.swift` (2: the `.more` item and a row's participants), `PersonSourcesSection.swift` (2: a Source's lines and the section's rows) |
| Return activation | `macos/App/DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` `pvActivatesOnReturn(_:)`, `pvActivatesOnReturn(when:_:)` | The second form is for a caller that owns its focus state and applies `.focused` itself |
| Focus claim | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)` | Writes on the next turn and again a turn later. 14 calls in `macos/App` |
| Scroll into view | `macos/App/DesignSystem/Components/ScrollRowIntoView/PVScrollRowIntoView.swift` `pvScrollsRowIntoView(_:anchor:…whenRequested:)` | Scrolls the enclosing `ScrollView` to the view tagged `.id(row)`; nil `anchor` moves the least. The Timeline already uses it inside the page's scroll view |
| Family state | `PersonFamilySection.swift` `@State expanded: Set<PersonFamilyContent.Fold>`; `PersonFamilyDiagram.expand` | Inserts the fold; no `@FocusState` in the diagram (P-14: focus is lost when the Show more goes) |
| Family content | `PersonFamilyContent.swift` `init(family:expanded:locale:orders:)`, `foldLimit` 30, `Fold` (`parentSets`, `siblings`, `partners`, `children(Int)`), `*Hidden` counts, tile ids `l-p{k}-{id}`, `l-s-{id}`, `r-{group}-{id}` | Folds each list; folds the siblings with the person among them, then appends the person when it falls at or past `limit` |
| Family layout | `PersonFamilyLayout.swift` `PlacedMore(fold:count:frame:)` | Places one more row per folded list with the `*Hidden` count |
| Timeline state | `PersonTimelineSection.swift` `PersonTimelineSection.expanded: Set<PersonTimelineContent.Fold>`; `PersonTimelineList.unfold` (`Task { @MainActor in focused = revealed }`), `select` (`expand(.dated)` for any id not in the items), `focusedOwner`, `lastFocused`, `keepFocus()` (called from `rebuild()` on every content, height and viewport change and on `widthClass`), `isShown(_:)`, the `onChange(of: focused)` branch | Private focus hop; a card that leaves for any reason (a width change or a reload) hands the keyboard to its row's title; a folded point opens every dated row |
| Timeline content | `PersonTimelineContent.swift` `init(timeline:expanded:locale:orders:widthClass:)`, `eventFoldLimit` 50, `participantFoldLimit` 5 / 12, static `row(_:expanded:limit:locale:orders:)`, `revealed: [Fold: String]`, `firstRevealed(by:)`, the dated hidden-run placement (lines 186–204), the undated fold (206–215) | Folds dated rows before the death marker past 50 into one `.more(.dated, count:)`; builds every row (S16-02's) |
| Sources state | `PersonSourcesSection.swift` `expanded: Set<String>`, `rowsExpanded: Bool`; `PersonSourcesList.unfold` and the section's Show more (`Task { @MainActor in focus = … }`) | Two `Task` hops; neither Show more highlights |
| Sources highlight | `PersonSourcesSection.swift` `PersonSourcesRowView.lines` (`highlightedGraphID = hover.id ?? focusedGraphID`); `PersonSourcesLineView` `.background(highlighted ? PVColor.surfaceSelected : Color.clear, in: RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))` | Only shown lines highlight (P-15). No `.id(` in the file |
| Sources content | `PersonSourcesContent.swift` `init(groups:sourceTypes:orders:expanded:rowsExpanded:)`, `Row.hiddenCount`, `Row.firstRevealedLineID`, `hiddenRowCount`, `firstRevealedRowLineID`, `claimGraphs` | Folds lines at 5 (min hidden 2) and rows at 50 |
| Gallery focus | `ConclusionGallerySection.swift` `lastFocused`, `pendingFocus`, `scrollTarget`, `scrollRequest`, the `onChange(of: focused)` block (lines 51–66), `onChange(of: widthClass)`, `restoreFocus(to:)`, `requestScroll()`, `applyPending()` (Use as portrait → image), the viewer-close restore (line 97), the tile `onAppear` re-claim | Private keeper for width change and viewer close; 400 ms expiry |
| Tests | `macos/ProvenenciaTests/ConclusionFoldTests.swift` `foldRule`; `expanded:` calls in `PersonFamilyContentTests.swift` (4), `PersonFamilyLayoutTests.swift` (3), `PersonTimelineContentTests.swift` (7), `PersonTimelineLifelineTests.swift` (7), `PersonSourcesContentTests.swift` (3, through its `content(_:expanded:rowsExpanded:)` helper), `PersonRecordLinksContentTests.swift` (1) | Each changes to `folds:`; `PersonTimelinePlacementTests.swift` has none |
| Previews | `PersonTimelineSection.swift` `PersonTimelinePreview.expandedAtStart`; `PersonSourcesSection.swift` `PersonSourcesPreview.expandedAtStart`, `rowsExpanded` | Debug-only; change with the types |

**Precedent.** `6401628` ("Show the Sources graph link's focus, open it on Return…") is the Sources focus shape; S13-17 (#729) added `FocusClaim` and its use in `PersonSourcesList.applyPendingFocus`: copy that call (`FocusClaim.claim($focus, value)`) for every Show more. `PersonTimelineSection.swift` `card(_:)` already puts `.focused(focused, equals: link.id)` on a `PVPersonCard`: Family's tiles follow it (D-8). Do not copy the `Task { @MainActor in focus = … }` hops.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ConclusionFold` becomes a value: `struct ConclusionFold: Hashable { let limit: Int; let minimumHidden: Int; static let pageSize = 50 }`. A page is 50 items after the fold's unchanged first limit: with `p` pages, `shownCount = min(count, limit + p × 50)`, and when `count − shownCount < minimumHidden`, everything shows. Each content declares its folds once as constants: `PersonFamilyContent.fold` (30, 1), `PersonTimelineContent.eventFold` (50, 1), `PersonTimelineContent.participantFold(_ widthClass:)` (5 or 12, 1), `PersonSourcesContent.lineFold` (5, 2), `PersonSourcesContent.rowFold` (50, 1); the old `foldLimit`, `eventFoldLimit`, `participantFoldLimit`, `lineFoldLimit`, `lineFoldMinimumHidden`, `rowFoldLimit` constants go | Working rule; each fold's numbers in one place (review M-1); S15-30 moves one value to the kit, S15-27 declares one more | Static functions taking `limit` and `minimumHidden` at every call (a Show more passing 1 where its content used 2 says 50 and reveals 51) |
| D-2 | The Show more label counts the next page: `fold.nextPageCount(hidden:)` = `hidden` when `hidden − 50 < minimumHidden`, else 50. The copy key stays (user, 2026-11-09) | FH-3 | The total folded (user chose the page) |
| D-3 | `ConclusionShowMoreRow` takes `hidden` and `fold` (the declared constant) and words the page itself; callers keep passing the folded total | One place does FH-3; `PersonFamilyLayout`'s `PlacedMore.count` and every `*Hidden` keep meaning "folded" | Each content type computing a page count |
| D-4 | `ConclusionFoldState<Fold: Hashable>: Hashable` *(new)*, in `ConclusionShowMoreRow.swift` beside `ConclusionFold`: `pages(_:) -> Int` (0 when never revealed), `mutating revealPage(_:)`, `mutating reveal(_:pages:)` (keeps the larger). Each section holds one in `@State`; content inits take `folds: ConclusionFoldState<Fold> = ConclusionFoldState()` in place of `expanded:` | Working rule "one fold state type"; `Hashable` so S16-27 can key derived values on it (M-7) | A dictionary per section; keeping `Set` beside a page count |
| D-5 | Sources' folds are `PersonSourcesContent.Fold` *(new)*: `.rows`, `.lines(sourceID: String)` | One state type for both of its folds | Keeping `rowsExpanded` |
| D-6 | Every section's content answers `firstRevealed(by: Fold) -> String?`: the focus id of the first item the next Show more adds, at the current pages; nil when the fold hides nothing. Timeline keeps its method; Family gains one (D-8); Sources gains one returning a claim id and drops `Row.firstRevealedLineID` and `firstRevealedRowLineID` (M-4). The view reads it from the content it drew, before revealing | One question, one shape in three sections; the content after the press has moved on | Per-row and per-section fields in Sources |
| D-7 | `ConclusionShowMoreRow<Focus: Hashable>(hidden: Int, fold: ConclusionFold, focus: FocusState<Focus?>.Binding, revealed: Focus?, expand: @escaping () -> Void)`. Inside: one `@FocusState private var hasKeyboard: Bool` and one `press()` that runs `expand()` then, when `hasKeyboard` and `revealed` is not nil, `FocusClaim.claim(focus, revealed)`. Wired as `Button(action: press)`, `.focused($hasKeyboard)`, `.pvActivatesOnReturn(when: hasKeyboard, press)`; the plain `.pvActivatesOnReturn(expand)` goes. Space, Return and a keyboard-focused click claim; a pointer click leaves focus where it is (board FH-4) | Working rule: the row claims focus through `FocusClaim`; one focus state and one action path (M-3, X-1); a borderless Button does not take the keyboard on a click | Two `.focused` states on one Button; moving focus on every press |
| D-8 | Family: `PersonFamilyDiagram` gains `@FocusState private var focused: String?`; each relative's tile gets `.focused($focused, equals: placed.id)` and `.id(placed.id)`; each Show more passes `content.firstRevealed(by: more.fold)` and `$focused`. `PersonFamilyContent.firstRevealed(by:)` *(new)*: `.parentSets` → the first partner tile of the first revealed set; `.partners` → the first revealed group's partner tile, else its first child tile (a group with no other parent); `.children(i)` → the first revealed child of group `i`; `.siblings` → the first revealed sibling. `PlacedMore` is unchanged | P-14; D-6 | Focusing the group container; carrying the id on `PlacedMore` |
| D-9 | Family's siblings fold runs over the siblings other than the person: `fold.shown(others, pages:)`. The person's tile is always shown, at its place in Go's order when the sibling before it is shown (or it is first), else after the last shown sibling. `siblingsHidden` counts other siblings only, so "Show 50 more" adds 50 tiles | The label is exact (X-5); the person is never in a page | Folding with the person counted (a page of 50 adds 49 when the person falls inside it) |
| D-10 | A lifeline click on a point opens the dated fold through that row's page: `PersonTimelineContent.pages(toShow eventID: String) -> Int?` *(new)* returns, for an id in `datedFoldIDs` *(new: a private `let [String]` the init keeps, the ids in order of the event rows the dated fold counts, which are the dated rows before the death marker, a recorded birth's row among them)*, `eventFold.pages(toShow: index, count: datedFoldIDs.count)`; nil for any other id (an undated row, a placeholder, a row after the death marker: none of them is ever folded). `select` reveals `.dated` through that many pages when it is non-nil and greater than the current, then scrolls as today. `ConclusionFold.pages(toShow:count:)` is the fewest `p ≥ 0` with `index < shownCount(p)` (user, 2026-11-09) | AC-1.8; FH-5 | Opening every row (user chose the page) |
| D-11 | `FocusKeeper<Stop: Hashable>: Hashable` *(new)*, `macos/App/Platform/FocusKeeper.swift`, held by the view in `@State`, with the modifier `focusKeeper(_ keeper: Binding<FocusKeeper<Stop>>, focus: FocusState<Stop?>.Binding, isShown: @escaping (Stop) -> Bool, fallback: @escaping (Stop) -> Stop?, reveal: @escaping (Stop) -> Stop)` *(new)* and two statics: `FocusKeeper.reclaim(keeper:focus:reveal:)` *(new)*, the one place a pending stop is claimed: `reveal(pending)` (the view scrolls and maps the stop) then `FocusClaim.claim(focus, revealed)`, with no new expiry; and `FocusKeeper.hand(_ stop:keeper:focus:reveal:)` *(new)*: `give(stop)` (sets `pending`), `reclaim`, then after 400 ms `expire(stop)` clears `pending` if it is still `stop`. The modifier makes every claim. Rules: **(a)** focus moves to a stop → `took(stop, fallback: fallback(stop), revealedPending: pending.map(reveal))` records it and clears `pending` when the stop is neither `pending` nor its revealed form (another stop took the keyboard, as Gallery's `onChange` does today); **(b)** focus goes nil → `lost(isShown:)` reports `.reclaim` when `pending` is set (a lazy view let the stop go: the modifier calls `reclaim`), `.fallback(stop)` when the last stop is no longer shown (a card folded away by a width change or a reload, M-5: the modifier claims it through `FocusClaim`), else `.none`, and the modifier schedules `forgetLast()` on the next main-actor turn unless focus came back (a click elsewhere: nothing is pulled back, AC-2.5); **(c)** `pvWidthClass` changes → `widthStop(current: focus, isShown:)` is `focus ?? last`, replaced by its fallback when not shown, then `hand`ed back. The order of the focus and width `onChange`s does not matter: either path reaches the same stop | Wide rule 5 and "focus outlived its view" in one place, with one claim path that always maps and scrolls (M2-2), as Gallery's `applyPending()` is today | A `DesignSystem` component (it holds no chrome; the working rule names `Platform/`); an `@Observable` class (a `FocusState` binding is only valid at call time); re-claims that skip `reveal` (a Use as portrait stop written to a tile that became the lead) |
| D-12 | Gallery adopts the keeper with `isShown` = the tile is in `gallery.tiles`, `fallback` = nil, `reveal` = set `scrollTarget` and bump `scrollRequest` (kept), then map `.useAs(id)` to `.image(id)` when that tile is the lead or `lead.isSaving` (today's `applyPending` rule, evaluated at each claim). The viewer close calls `FocusKeeper.hand(.image(old.shownID), …)`; the tile `onAppear` calls `FocusKeeper.reclaim(…)` when `keeper.pending?.id == tile.id`. Deleted: `lastFocused`, `pendingFocus`, the `onChange(of: focused)` block, `onChange(of: widthClass)`, `restoreFocus(to:)`, `requestScroll()`, `applyPending()` | One keeper and one claim path; no second pending state (X-3, M2-2) | Keeping Gallery's private keeper |
| D-13 | Timeline adopts the keeper with `isShown` = `content.isShown(_:)` *(moved from the view)*, `fallback` = `content.row(holdingCard:)` *(new, `String?`: the row id of a shown card, nil for a row title or an unknown id)*, `reveal` = identity. Deleted: `focusedOwner`, `lastFocused`, `keepFocus()`, `PersonTimelineList.isShown(_:)`, the `onChange(of: focused)` and `onChange(of: widthClass)` handlers, and `rebuild()`'s `keepFocus()` call | D-11 (b) keeps today's reload fallback | Parsing the card id |
| D-14 | Timeline wiring: `PersonTimelineItemView` and `PersonTimelineEventRow` gain `revealed: (PersonTimelineContent.Fold) -> String?` (passed `content.firstRevealed(by:)`); the `.more(fold, count:)` item's Show more and a row's participants Show more pass `revealed(fold)` and the `focused` binding they already hold. `unfold(fold)` keeps setting `pendingReveal` for `.dated` and `.undated` (the scroll to the top) and drops the `Task`; a card reveal does not scroll. The static `row(…)` helper takes `folds` and reads `pages(.participants(eventId:))`; the undated fold reads `pages(.undated)`; the `revealed` map is built with `firstRevealed(_:pages:)` | X-9, X-11 | A second focus path in the list |
| D-15 | Sources highlight: `PersonSourcesContent.foldedGraphIDs(by: Fold) -> Set<String>` *(new)*: for `.lines(sourceID:)`, the graphs with a line behind that Source's fold and no shown line on it; for `.rows`, the graphs of the Sources behind the row fold (a graph is on one Source). The sets go down as data and only views that already read the hover compute the highlight, so `PersonSourcesList` never reads `hover.id` (its documented contract): `PersonSourcesRowView` gains `foldedGraphIDs: Set<String>` and highlights its Show more inside `lines`, beside the line highlight; the section's Show more becomes `PersonSourcesRowsShowMore` *(new, private in `PersonSourcesSection.swift`)*, which takes `hover`, `focusedGraphID` and the `.rows` set. The rule "the pointer's graph, else the keyboard's" gets one home, `CircularLinkHover.highlighted(focused: ID?) -> ID?` *(new)* = `id ?? focused`, used by `CircularLinkDiagram.body` (line 53), `PersonSourcesRowView.lines` and `PersonSourcesRowsShowMore` | AC-3.1 – AC-3.4; a hover rebuilds nothing upstream (`CircularLinkHover` doc); one definition of the highlighted graph instead of three (M2-1) | The list computing the flag (re-runs its body and the diagram host's `make` on every pointer move) |
| D-16 | The highlight look lives once, as a private modifier in `PersonSourcesSection.swift`: `personSourcesHighlight(_ on: Bool)` *(new)* = today's `PersonSourcesLineView` background (`PVColor.surfaceSelected`, `RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)`). `PersonSourcesLineView` and both Sources Show mores use it; the Show mores get `.frame(maxWidth: .infinity, alignment: .leading)` first, so the background spans the line's full width (board FH-1). `ConclusionShowMoreRow` has no highlight parameter | One look, one owner; the shared row carries no Sources state into the kit (M-2) | A `highlighted:` flag on the shared row that two of three callers ignore |
| D-17 | After a keyboard Show more, Family and Sources scroll the revealed item just into view: each keeps `@State scrollTarget`/`scrollRequest` and applies `pvScrollsRowIntoView(scrollTarget, whenRequested: scrollRequest)` (nil `anchor`): Family on `PersonFamilyDiagram`'s outer frame with `.id(placed.id)` on each tile; Sources on the `rows` `VStack` with `.id(line.id)` on `PersonSourcesLineView` (the file has no `.id` today). The Show more's `expand` closure sets the target to the revealed id. The page's `ScrollView` is the scroll host, as for the Timeline | Board FH-4 ("scrolled just into view"; Timeline: "scrolls it to the top") (X-12) | Relying on focus to scroll (macOS does not scroll a `ScrollView` to a focused view) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `struct ConclusionFold` (was an `enum` of statics) with `pageSize`, `shown(_:pages:)`, `nextPageCount(hidden:)`, `firstRevealed(_:pages:)`, `pages(toShow:count:)` | type | `enum ConclusionFold` | Extends the one fold rule in place (searched: `prefix(`, `Show more`, `limit:` in `Features/`) |
| `ConclusionFoldState` *(new)* | type | `Set<PersonFamilyContent.Fold>`, `Set<PersonTimelineContent.Fold>`, `Set<String>` + `rowsExpanded` | Replaces all four; a `Set` cannot hold a page count |
| `PersonSourcesContent.Fold` *(new)* | type | `PersonFamilyContent.Fold`, `PersonTimelineContent.Fold` | Same shape per section; Sources had none |
| `FocusKeeper`, `focusKeeper(…)`, `FocusKeeper.hand(…)` *(new)* | type + modifier | Gallery's `restoreFocus`/`applyPending`, Timeline's `keepFocus` | Both are deleted onto it; `FocusClaim` stays the writer it calls (searched: `widthClass) {`, `lastFocused`, `pendingFocus` in `macos/App`) |
| `PersonFamilyContent.firstRevealed(by:)`, `PersonSourcesContent.firstRevealed(by:)`, `PersonTimelineContent.row(holdingCard:)`, `isShown(_:)`, `pages(toShow:)` *(new)* | func | `PersonTimelineContent.firstRevealed(by:)`, `PersonTimelineList.isShown(_:)` | Content owns ids; the views ask it |
| `PersonSourcesContent.foldedGraphIDs(by:)` *(new)* | func | `claimGraphs`, `graphLocations` | Built in the same pass over claims |
| `personSourcesHighlight(_:)` *(new)* | modifier (private) | `PersonSourcesLineView`'s background | Moves the one look to one place for three uses |
| `PersonSourcesRowsShowMore` *(new)* | view (private) | the section's inline Show more in `PersonSourcesList.rows` | Only this view reads the hover for the rows fold (D-15) |
| `CircularLinkHover.highlighted(focused:)` *(new)* | func | `hover.id ?? focusedGraphID` in `PersonSourcesRowView.lines`, `hover.id ?? focusedID` in `CircularLinkDiagram.body` | One home for the rule (D-15) |
| `PersonTimelineContent.datedFoldIDs` *(new)* | property (private) | the `counted` local in `init` | `pages(toShow:)` needs it after the init (D-10) |
| `FocusKeeperTests` *(new)* | test file | `ConclusionFoldTests` | New type |

**Builds on.** `FocusClaim` (S13-17), `ConclusionFold` ("the one fold rule"), `pvActivatesOnReturn(when:_:)`, `pvScrollsRowIntoView`, `CircularLinkHover` (the one hovered graph).
**Deletes.** `ConclusionFold`'s static `shown(_:limit:minimumHidden:expanded:)` and `firstRevealed(_:limit:minimumHidden:)`; the fold-number constants D-1 lists; `PersonTimelineList.focusedOwner`, `lastFocused`, `keepFocus()`, `isShown(_:)` and its two `onChange` handlers (D-13); `ConclusionGallerySection.lastFocused`, `pendingFocus`, its `onChange(of: focused)` and `onChange(of: widthClass)` blocks, `restoreFocus(to:)`, `requestScroll()`, `applyPending()` (D-12); `PersonSourcesContent.Row.firstRevealedLineID`, `firstRevealedRowLineID`; the three `Task { @MainActor in focus… = … }` hops; `PersonSourcesSection.rowsExpanded`, `PersonSourcesList.expandRows`; `ConclusionShowMoreRow`'s `.pvActivatesOnReturn(expand)`; the `expanded:` init parameters and preview `expandedAtStart` sets (become `ConclusionFoldState` values).
**Next adds.** S16-02 reads `folds` to build only shown rows; S15-30 moves `ConclusionFold`, `ConclusionFoldState` and the row to `Platform/` and the kit (and owns any rename); S15-27's provenance table declares one `ConclusionFold(limit: 50, minimumHidden: 1)`; S16-27 keys derived values on the `Hashable` fold state; any later view whose focused item can disappear adopts `focusKeeper`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A fold with `pages > 0` and a list that shrinks after a reload | `ConclusionFold.shown` | `min(count, …)`: a page past the end shows the whole list, no crash; T-1 |
| Another person's page | each section's `@State folds` and keeper | The sections are new views per entity id, as their `Set`s are today: folds and keepers start empty |
| A revealed item gone before the claim lands | `FocusClaim.claim` | Writes a value with no focusable view; SwiftUI leaves focus nil, as the Sources `applyPendingFocus` path does today |
| A highlighted graph that is not on the page | `foldedGraphIDs(by:)` | No set contains it; nothing highlighted (T-9) |
| A focused card removed by a reload | `FocusKeeper.lost(isShown:)` | Its row's title takes the keyboard, as today (T-10) |

### Scale

At the spike's crafted Scale (2,000 events, 500 participants on one event, a Source with hundreds of lines): `shown`, `nextPageCount` and `firstRevealed` are arithmetic plus one `prefix`; `pages(toShow:)` is one index lookup per click. `foldedGraphIDs` is filled in the pass over the claims `PersonSourcesContent.init` already makes. The keeper's `isShown` is a set lookup in the content. The Timeline still builds every row until S16-02; this PR adds no work per row. Real data (200 events, 50 graphs) never passes a first limit except Timeline's participants, as today.

### Edges

- **Assumptions.** `firstRevealed` assumes items keep their order between the draw and the press (Go's order; true for one content value). The Timeline's hidden-run placement assumes the hidden rows are a suffix of the counted rows; with pages it is still a suffix (T-4). Family's siblings: the person is never in a page (D-9), so a page always adds `nextPageCount` tiles (T-6).
- **Neighbors.** `minimumHidden` 2 on Sources lines: 6 lines → whole; 7 → 5 + "Show 2 more"; 55 → 5 + "Show 50 more" then 55; 56 → "Show 51 more"; 57 → "Show 50 more" then "Show 2 more" (T-1). Participants at Regular 5 vs Wide 12 with the same pages (T-5). Family siblings with the person within the first 30 Go places: 30 others plus the person show (31 tiles), where today 30 tiles include the person (D-9; the writer's call, in the review log).
- **Removed inputs.** A reload that drops items keeps `pages`; `shown` caps at `count` (T-1). An unpromoted line inside a revealed page: `lineAfterRemoving` reads shown lines only, unchanged. A focused card removed by a reload: its row takes the keyboard (D-11 b, T-10).
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
| Show more row | `ConclusionShowMoreRow` (snowflake, Extend) | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` | Ghost small `PVButton`, as today; label `L10n.Conclusions.showMore(fold.nextPageCount(hidden:))`; claims focus (D-7) |
| Folded-record highlight | `personSourcesHighlight(_:)` on the Sources Show mores (snowflake) | `PersonSourcesSection.swift` | The record line's look, board FH-1 (D-16) |
| Focus keeper | `FocusKeeper` *(new)* | `macos/App/Platform/FocusKeeper.swift` | No chrome |

```swift
struct ConclusionFold: Hashable {
    static let pageSize = 50
    let limit: Int
    let minimumHidden: Int
    func shown<T>(_ list: [T], pages: Int) -> (shown: [T], hidden: Int)
    func nextPageCount(hidden: Int) -> Int
    func firstRevealed<T>(_ list: [T], pages: Int) -> T?
    /// The fewest pages p ≥ 0 that show `list[index]`.
    func pages(toShow index: Int, count: Int) -> Int
}

struct ConclusionFoldState<Fold: Hashable>: Hashable {
    func pages(_ fold: Fold) -> Int
    mutating func revealPage(_ fold: Fold)
    mutating func reveal(_ fold: Fold, pages: Int)
}

struct FocusKeeper<Stop: Hashable>: Hashable {
    private(set) var last: Stop?
    private(set) var lastFallback: Stop?
    private(set) var pending: Stop?
    enum Lost: Equatable { case none, reclaim, fallback(Stop) }
    mutating func took(_ stop: Stop, fallback: Stop?, revealedPending: Stop?)
    mutating func lost(isShown: (Stop) -> Bool) -> Lost
    mutating func forgetLast()
    /// The stop to hand back on a width change, or nil when nothing had the keyboard.
    func widthStop(current: Stop?, isShown: (Stop) -> Bool) -> Stop?
    mutating func give(_ stop: Stop)
    mutating func expire(_ stop: Stop)
    @MainActor static func reclaim(keeper: Binding<FocusKeeper>, focus: FocusState<Stop?>.Binding, reveal: @escaping (Stop) -> Stop)
    @MainActor static func hand(_ stop: Stop, keeper: Binding<FocusKeeper>, focus: FocusState<Stop?>.Binding, reveal: @escaping (Stop) -> Stop)
}
```

States: an empty section has no fold (unchanged). Accessibility: the Show more's label is spoken as drawn ("Show 50 more"); the highlight is not spoken (S16-D2 §2.1); focus moves are the VoiceOver cursor's moves.

### Strings

None — `conclusions.showMore` ("Show %lld more") is unchanged; only its number changes (D-2). No copy names the old "show all" scope.

## 7. Steps

**STEP-1 — The paged fold and its state**
Files: `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift`, `macos/ProvenenciaTests/ConclusionFoldTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, D-4: turn `ConclusionFold` into the struct; keep the two old statics, marked for deletion, until STEP-5 removes their last caller; add `ConclusionFoldState`. Rewrite `foldRule` onto the instance methods (its cases keep their expectations with `pages: 0` and with pages large enough to show all).
Check: T-1, T-2 and `foldRule` pass (`xcodebuild test -scheme Provenencia -only-testing:ProvenenciaTests/ConclusionFoldTests`).

**STEP-2 — The Show more row words the page and claims focus**
Files: `ConclusionShowMoreRow.swift`, the five call sites (§4). Skill: `add-ui-component`.
Do: D-3, D-7, and the parts of D-1 and D-8 the new init needs: add every fold constant D-1 lists beside the old constants (each section step removes its old ones), and add D-8's `@FocusState private var focused: String?` to `PersonFamilyDiagram`. Each call site passes its fold constant (the Timeline row's participants: `PersonTimelineContent.participantFold(widthClass)`, from the row's `pvWidthClass` environment), its section's focus binding and `revealed: nil` for now, so the tree builds; the label reads the page while a click still reveals all until STEP-3 – STEP-5.
Check: the app builds; T-1's label cases pass.

**STEP-3 — Family pages and moves focus**
Files: `PersonFamilySection.swift`, `PersonFamilyContent.swift`, `macos/ProvenenciaTests/PersonFamilyContentTests.swift`, `PersonFamilyLayoutTests.swift`.
Do: D-1, D-4, D-6, D-8, D-9, D-17.
Also: remove `foldLimit`; re-point `PersonFamilyContentTests.foldsEachListAtThirty`'s `foldLimit == 30` at `PersonFamilyContent.fold.limit == 30` (allowed under the Stop conditions).
Check: T-6 passes; existing Family tests pass with `folds:` (build a state with `var f = ConclusionFoldState<PersonFamilyContent.Fold>(); f.revealPage(.siblings)`).

**STEP-4 — Timeline pages and moves focus**
Files: `PersonTimelineSection.swift`, `PersonTimelineContent.swift`, `macos/ProvenenciaTests/PersonTimelineContentTests.swift`, `PersonTimelineLifelineTests.swift`.
Do: D-1, D-4, D-6, D-10, D-14; add `isShown(_:)` and `row(holdingCard:)` to the content (used in STEP-7); remove `eventFoldLimit` and `participantFoldLimit`; re-point `PersonTimelineContentTests`' assertion `participantFoldLimit(.regular) == 5 && participantFoldLimit(.wide) == 12` at `participantFold(_:).limit` (allowed under the Stop conditions).
Check: T-3, T-4, T-5, T-7 pass; `PersonTimelineLifelineTests` pass unchanged in meaning (every point still drawn).

**STEP-5 — Sources pages, moves focus and highlights a folded graph**
Files: `PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `macos/App/Features/Conclusions/CircularLinkDiagram/CircularLinkData.swift` (`highlighted(focused:)`), `CircularLinkDiagram.swift` (line 53 uses it), `macos/ProvenenciaTests/PersonSourcesContentTests.swift` (its helper takes `folds:`), `PersonRecordLinksContentTests.swift` (drops `expanded: []`).
Do: D-1, D-4, D-5, D-6, D-15, D-16, D-17; the view wraps `firstRevealed(by:)`'s claim id as `.graph(claimID:)`, as `PersonSourcesList` does today; remove `lineFoldLimit`, `lineFoldMinimumHidden`, `rowFoldLimit` and `ConclusionFold`'s old statics (no caller left).
Check: T-8, T-9 pass; `rg "expanded: Bool|rowsExpanded|firstRevealedLineID|Task \{ @MainActor in focus" macos/App/Features/Conclusions` finds nothing.

**STEP-6 — The focus keeper**
Files: `macos/App/Platform/FocusKeeper.swift` *(new)*, `macos/ProvenenciaTests/FocusKeeperTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-11.
Check: T-10 passes.

**STEP-7 — Gallery and Timeline keep focus through the keeper**
Files: `ConclusionGallerySection.swift`, `PersonTimelineSection.swift`.
Do: D-12, D-13.
Check: manual checks 4 and 5; `rg "pendingFocus|lastFocused|keepFocus|focusedOwner" macos/App/Features/Conclusions/ConclusionGallerySection.swift macos/App/Features/Conclusions/PersonTimelineSection.swift` finds nothing.

**STEP-8 — Docs**
Files: `docs/macos-client-patterns.md` ("One sectioned detail page": one sentence on the paged fold, `ConclusionFoldState` and Show more's focus claim; in § 9, one sentence after the paragraph that follows the six rules, naming `FocusKeeper` for rule 5 — the rules' text stays word for word, since `add-design-brief` repeats it), `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status), `docs/deployment-plan/spike-16/design/S16-D2-person-page-folds.md` → `docs/deployment-plan/spike-16/design/archive/` *(new)*, `docs/deployment-plan/spike-16/design/README.md` (move the S16-D2 row from Scheduled to Archived, Built by S16-07).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-07-paged-fold-focus.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Unfold long Person page lists fifty at a time and keep the keyboard on the first revealed item.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `ConclusionFoldTests.swift` `pagedFold` *(new)* | `ConclusionFold(limit: 5, minimumHidden: 2)`, 120 items: pages 0/1/2/3 → shown 5/55/105/120, `nextPageCount` 50/50/15/—; 56 items → 5 then `nextPageCount` 51; 7 → 2; 6 → whole; pages 9 on 20 items → all 20 | AC-1.2, AC-1.3, D-1, D-2 |
| T-2 | Swift | `ConclusionFoldTests.swift` `foldStateAndPagesToShow` *(new)* | `revealPage` twice → 2; `reveal(pages: 1)` after 2 → 2; unknown fold → 0; `ConclusionFold(limit: 50, minimumHidden: 1).pages(toShow:count: 2000)`: 49 → 0, 50 → 1, 199 → 3; `ConclusionFold(limit: 5, minimumHidden: 2).pages(toShow: 5, count: 6)` → 0 (six show whole) | D-4, D-10 |
| T-3 | Swift | `PersonTimelineContentTests.swift` `twoThousandEventsUnfoldInPages` *(new)* | 2,000 dated events: each of 39 reveals adds 50 rows and `firstRevealed(by: .dated)` names `d(50k+1)`; after 39 no `.more` | AC-1.1, AC-1.6 |
| T-4 | Swift | `PersonTimelineContentTests.swift` `foldedMarkersStayAfterPagedRun` *(new)* | 120 dated rows with a placeholder death among the hidden: one page → `.more(.dated)` sits after row 100, the marker still after it | D-1, D-6 |
| T-5 | Swift | `PersonTimelineContentTests.swift` `participantsPageAtBothWidths` *(new)* | 500 participants, `pages 1`: Regular 55 cards, Wide 62; `firstRevealed` names card 56 / 63 | AC-1.5 |
| T-6 | Swift | `PersonFamilyContentTests.swift` `childrenAndSiblingsPage` *(new)* | 100 children → 30/80/100 and `firstRevealed(by: .children(0))` = the 31st child's tile id; 3 parent sets with 2 partners each, folded at 1 → `firstRevealed(by: .parentSets)` = set 2's first partner tile; 130 siblings with the person 40th in Go's order → folded: 30 others + the person appended, `siblingsHidden` 99, label 50; one page: 80 others with the person in Go's place, `siblingsHidden` 49; `firstRevealed(by: .siblings)` is never the person | AC-1.4, AC-1.6, AC-1.7, D-8, D-9 |
| T-7 | Swift | `PersonTimelineContentTests.swift` `pointOnFoldedPageOpensThroughIt` *(new)* | 2,000 events, `pages(toShow: "d180")` → 3; with it, `d180` is shown and `d201` is not; an undated row's id and `"death"` → nil | AC-1.8, D-10 |
| T-8 | Swift | `PersonSourcesContentTests.swift` `linesAndRowsPage` *(new)* | 120 lines on one Source; 60 Sources: pages as AC-1.2 and rows 50 → 60; `firstRevealed(by: .lines(sourceID: "src"))` and `firstRevealed(by: .rows)` per page | AC-1.2, AC-1.3, AC-1.6, D-6 |
| T-9 | Swift | `PersonSourcesContentTests.swift` `foldedGraphsTakeTheHighlight` *(new)* | 9 lines on 9 graphs: `foldedGraphIDs(by: .lines(sourceID:))` = graphs of lines 6–9; a graph with lines 3 and 7 is not in it; 60 Sources: `foldedGraphIDs(by: .rows)` holds the 51st–60th Sources' graphs; an unknown graph in neither; `CircularLinkHover.highlighted(focused:)`: id "g1", focused "g2" → "g1"; id nil → "g2" | AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, D-15 |
| T-10 | Swift | `FocusKeeperTests.swift` `keepsTheStopAcrossAWidthChangeAndAReload` *(new)* | `took("b", fallback: "row")`; `widthStop(current: nil, isShown: all)` → "b"; with "b" not shown → "row"; `lost(isShown: none)` → "row" (a reload took the card); `lost(isShown: all)` → nil (a click elsewhere), then `forgetLast()` → `widthStop` nil; `give("x")` sets `pending`, `lost` → `.reclaim`; `took("z", revealedPending: "x")` clears `pending`; `give(".useAs")` then `took(".image", revealedPending: ".image")` keeps it; `expire("y")` keeps it; `expire("x")` clears it | AC-2.1, AC-2.2, AC-2.5, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PersonTimelineContentTests`, `PersonTimelineLifelineTests`, `PersonFamilyContentTests`, `PersonFamilyLayoutTests`, `PersonSourcesContentTests` (their `expanded:` calls and the siblings change).

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
2. Same on an event's participants and on Family's children (30 shown): the keyboard lands on the first revealed card or tile, and Family's tile is scrolled just into view. Click a Show more with the pointer while a title has the keyboard: the list unfolds and the title keeps the keyboard (AC-1.9). Cross 1600 with a Family tile focused: it keeps the keyboard.
3. Sources: point at a node whose graph's line is folded: the Source's Show more lights up (FH-1); at Wide the node and the row are both on screen (FH-W1). With more than 50 Sources, a node on a folded Source lights the section's Show more (FH-2).
4. Timeline at Wide, focus the 9th card of an event, narrow the window below 1600: the keyboard is on the event's title. Widen again with the 3rd card focused: it keeps the keyboard.
5. Gallery: focus a tile, cross the width line both ways: the tile is in view with the keyboard (AC-2.3). Focus a tile's Use as portrait, make it the lead, cross the line: the keyboard is on its image (AC-2.4). Open and close the likeness viewer: the keyboard returns to the tile.
6. Click a lifeline point on a folded row deep in the list: the list opens through that page only and scrolls to it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `FocusClaim`'s two-turn write lands before a revealed row exists | It writes twice by design; the Sources `applyPendingFocus` path already relies on it. T-3 covers the id; manual check 1 the landing |
| The keeper changes the Gallery's lazy-grid timing | D-11 keeps the tile `onAppear` re-claim and the 400 ms expiry; manual check 5 |
| Family tiles gaining `.focused` changes Tab order | Groups are already containers in reading order (`PersonFamilyDiagram` comment); manual check 2 |

**Stop and report if:**

- The board and this plan disagree on a frame: frames F1–F9 map to US-1 – US-3 and D-16, D-17.
- `ConclusionShowMoreRow`'s own `@FocusState` cannot tell a key press from a pointer click (D-7), or a Family tile loses the keyboard when the window crosses 1600 (Family has no keeper in this plan).
- `FocusClaim.claim` does not land on a revealed Timeline row after two turns in manual check 1 (needs a different claim path; do not add a third write).
- `PersonFamilyLayout` needs to place a Show more differently because of pages (D-3 says only its count changes).
- A test outside §8 must change its expected output for reasons other than `expanded:` → `folds:`.

## 12. Open questions

None.

## 13. PR

**Title:** S16-07 — One paged fold that moves focus
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-07-paged-fold-focus.md`.
