# S13-12 — Family section

**Status:** Blocked — on board S13-D3, and on S13-04 landing
**Spike / slice:** Spike 13 · Slice 3 — Family
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D3](../design/S13-D3-family-section.md), board not drawn yet
**Board:** — (not agreed). When agreed: the Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, the S13-D3 turn. The implementer may not be able to open it, so on agreement this plan's **Design → macOS** describes each frame in words (S13-D3 § 7)
**Depends on:** S13-04 (implemented on branch `s13-04-person-family-read` @ `b12a7fe`, pull request [#671](https://github.com/mendahu/provenencia/pull/671), not on `main` at the planning base), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S13-D3 (board)
**Blocks:** S13-13 (it reuses the person tile)
**Branch:** `s13-12-family-section`
**Runs on:** macOS: STEP-3 – STEP-8 and `scripts/check-pr.sh` need Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01 and S13-04 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md) (the Go order step)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 2. Family](../person-detail.md#2-family) (R-2, R-3, R-16 – R-22); brief [S13-D3](../design/S13-D3-family-section.md); the S13-04 plan [`S13-04-person-family-read.md`](S13-04-person-family-read.md) on its branch; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Family groups in Go", "One person summary for every tile", "Canonical only, filtered in Go", "One person tile, built once", "Colors reuse `PersonGenderTreatment`", "A family tile opens the relative on Family", "Sections are places"
**Planned:** 2026-11-01 against `main` @ `75d9494` plus the S13-04 branch @ `b12a7fe` (every S13-04 path below is read from that branch)
**Review:** [`S13-12-family-section.review.md`](S13-12-family-section.review.md)

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

A researcher opens a Person's Family section and sees the person's parents, siblings, each partner and each family's children as person tiles joined by lines, and a click on any relative opens that relative's Family, so they can walk the family from person to person. It fills the empty place S13-10 left (S13-D1 PPF-10) with the data S13-04 reads, and builds the person tile the Timeline (S13-13) reuses.

**Why now.** Slice 3's Check is bar 3 ("The family is there") and bar 8; S13-04 has the read, and this is the last PR of the slice.

**What stays.** The header, the switcher, Vitals, Gallery and the Timeline and Sources empty places; `GetPersonFamily`'s units, their order and their filtering (S13-04); the gender colors; every other link into a Person page, which still lands on Vitals.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Family section, I see their nuclear family laid out by family, so that I understand who they grew up with and who grew up with them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (S13-D3 § 5 sample): parents Heinrich and Anna; siblings Margarethe, Friedrich, Wilhelm; wives Maria (children Karl, Luise, Emma) and Sophie (child Otto); son Paul with no other parent | One parent unit (Heinrich, Anna over Margarethe, Johann, Friedrich, Wilhelm), two partner units (Johann + Maria over Karl, Luise, Emma; Johann + Sophie over Otto), one unit of Paul under Johann alone; lines as the board draws them (FS-4, FS-6, FS-7) |
| AC-1.2 | Johann with a second parent set (Heinrich and Katharina) and one sibling under it | Two parent units, each with its own children row (FS-5, R-19) |
| AC-1.3 | A person whose only family is two recorded siblings | One siblings unit: the focus tile and the two siblings, no parent row (FS-8) |
| AC-1.4 | Siblings Friedrich b. 1820, Margarethe b. 1815, Wilhelm undated, returned by Go in link order Friedrich, Wilhelm, Margarethe | Shown Margarethe, Friedrich, Wilhelm: earliest born first, undated last (D-2) |
| AC-1.5 | A person with no promoted relatives (`units` empty) | The EmptyState of FS-11 |
| AC-1.6 | The family read fails with `conclusiondetails.not_found` | A danger Callout in the section with the mapped message; the header and switcher still show |

**US-2.** As a researcher, each relative's tile tells me who they are at a glance.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Anna: woman, a likeness, b. 3 May 1790 · Bremen, d. 1851 | Tile: the likeness with the woman tint and line, "Anna Weber", "b. 3 May 1790 · Bremen", "d. 1851" (FS-1) |
| AC-2.2 | Wilhelm: no gender, no likeness, no dates or places | The person mark with the unknown treatment's colors, the name alone, no b. or d. line |
| AC-2.3 | A child known only by ref `PER-7KQ2M` | The ref in mono as the title |
| AC-2.4 | VoiceOver on Anna's tile | "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851", a button (FS-13) |
| AC-2.5 | The focus tile (Johann on his own page) | Set apart as the board draws it; not a button, no focus stop (FS-3) |

**US-3.** As a researcher, a click on a relative takes me to their family.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | On Johann's Family, click Anna's tile | Anna's Person page opens on Family; breadcrumbs *Persons › Anna Weber › Family* (FS-10) |
| AC-3.2 | Then Back | Johann's Family again |
| AC-3.3 | Tab to Friedrich's tile, press Return | Friedrich's Family opens (FS-13) |

**US-4.** As a researcher, the section stays current while I work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Johann's Family is showing; a write touches Anna's handle | The family reloads; the diagram stays drawn and the switcher row shows "Refreshing" until it lands (FS-12, D-8) |
| AC-4.2 | Opening a Person straight on Family (from a tile, or Back) | The family read starts with the page, not after the detail read (D-9) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Children in each unit in birth order, in Go (D-2) | STEP-1, AC-1.4 |
| The b. and d. line lifted to one shared builder (D-6) | STEP-2 |
| The person tile (FS-1, FS-2, FS-3) | STEP-3, AC-2.1 – AC-2.5 |
| The section's content model: units, roles, locations, accessibility (FS-4 – FS-8, FS-13) | STEP-4, AC-1.1 – AC-1.3, AC-2.4 |
| The section view: diagram, lines, states (FS-4 – FS-12, FS-14, FS-W1 – FS-W3) | STEP-5, AC-1.5, AC-1.6, AC-4.1 |
| A section's own refreshing in the switcher row (FS-12) | STEP-5, AC-4.1 |
| Wiring into the Person page and warming the key on a Family location (FS-10) | STEP-6, AC-3.1 – AC-3.3, AC-4.2 |
| Copy (FS-11, FS-13) | STEP-7 |

| Out | Home |
| --- | --- |
| Grandparents and beyond; a tree view | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Marking recorded versus implied links | not planned (user, 2026-11-01); the data stays on `CatalogFamilyLink` for a later tree |
| The provenance control on a tile or line | not planned this spike (R-13b) |
| The tile in the Timeline, and its move to `DesignSystem/Recipes/` | S13-13 |
| Ordering partner units (by marriage date) | not planned; units keep S13-04's order |
| Editing a family from the section | not planned this spike |
| Opening a Person on Family from anywhere other than a family tile | not planned; other links keep Vitals (S13-D1 PPF-5) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Family section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.family, .timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. It receives `content` and `context`, not the `session` |
| Section context | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing` (the detail read's), `pageIdentifier`, `lead`, `pageSection` |
| Switcher refresh | same file `ConclusionDetailHeaderBand.switcher(selection:)` | Shows `ProgressView` + `L10n.Conclusions.detailRefreshing` when `refreshing`, which is the `conclusionDetail` handle's `isFetching` (from `ConclusionDetailLoaded`) |
| Family read (Swift) | S13-04 branch: `macos/App/Platform/GenealogyStore.swift` `CatalogPersonFamily`, `CatalogFamilyUnit` (`Kind`: `.parents`, `.partner`, `.oneParent`, `.siblings`), `CatalogFamilyMember` (`person: CatalogPersonSummary`, `link: CatalogFamilyLink?`), `getPersonFamily(projectDir:entityID:)` | Units in Go's order. The person themselves is in no unit. `link` is nil for a co-parent |
| Family key | S13-04 branch: `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personFamily(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, refreshes on any touched handle. No view queries it |
| Place warm | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` Spec `.personDetail` | Warms only `.conclusionDetail` |
| Family read (Go) | S13-04 branch: `core/database/persondetail/family.go` `PersonFamily`, `dropUnread` | Returns units with children in link order (`kinship.FamilyUnits`), summaries alongside |
| Handler | S13-04 branch: `api/ffi/handlers/person_detail.go` | Maps units and members to proto in the order given |
| Date order | `core/database/conclusionsummaries/compose.go` `dateSort`, `eventKeyLess` | The Go owner of date sort keys (`autoreconcile.SortKey`); events sort dated first, then by key, then ref |
| Title order | `core/database/conclusionsummaries/order.go` `SortByTitle` | "The one owner of title order; the app shows rows as given" |
| Life lines | `macos/App/Features/Conclusions/PersonDetailContent.swift` `LifeLine`, `lifeLine(_:spoken:facts:locale:)` (private) | Builds the header's b. / d. line from `CatalogLifeFacts`; only the header uses it |
| Line view | `macos/App/Features/Conclusions/ConclusionDatePlaceLine.swift` `ConclusionDatePlaceLine` | Draws *b. date · place* |
| Gender colors | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `read(_:)`, `thumbnailColors` | Reads a `CatalogPersonSummary` |
| Title | `PersonTitleDisplay.titleSource(_:)`, `ConclusionListRow.title(_:)` | Name / italic label / mono ref |
| Thumb | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail` | `fallbackMark`, `colors`, `imageLabel` |
| Card | `macos/App/DesignSystem/Components/Card/PVCard.swift` `PVCard` | Tones `.card`, `.raised`, `.sunken`; solid or dashed |
| Wrapping | `macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift` `PVFlowLayout` | Wrapping row layout |
| Location | `macos/App/Features/Workspace/WorkspaceLocation.swift` `WorkspaceLocation.personDetail(entityId:ref:title:pageSection:)` | Already takes a page section |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`; S13-04 branch `core/database/persondetail/family_test.go` `TestPersonFamily`, `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personFamilyFollowsEveryTouchedHandle` | Content tests; the Go family read; the key's refresh |

**Precedent.** S13-10 ([#634](https://github.com/mendahu/provenencia/pull/634)) filled the Person page's sections inside `PersonDetailSections` and gave Vitals no heading; copy its shape for the section switch. S13-11 ([#650](https://github.com/mendahu/provenencia/pull/650)) navigated from a Person page section with `navigation.go(to:)` from the environment; copy that. `ConclusionDetailGallerySection` (S14-02) is the nearest section that lays out a wrapping grid of tiles with `CachedThumbnail`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The tile opens the relative with `WorkspaceLocation.personDetail(entityId:ref:title:pageSection: .family)` (user, 2026-11-01). It is one `navigation.go(to:)` | R-20; walking a family stays in Family. A new working rule "A family tile opens the relative on Family" records it | Vitals (S13-D1 PPF-5) — the user chose Family for family tiles |
| D-2 | Go orders each unit's **children** by birth: dated first, by the birth date's sort key, then by ref; partners and units keep S13-04's order (user, 2026-11-01). The order is applied in `persondetail.PersonFamily` after `dropUnread`, from the summaries it already holds, through `conclusionsummaries.BornBefore` *(new)* beside `eventKeyLess` | Order is data (working rule "Engine PRs run ahead of their slice"; `SortByTitle`'s "the app shows rows as given"); the date sort key lives in Go. Only the page read sorts: `persondetail.Families`, the tree's reusable step, stays id-only | Sorting in Swift — Swift has no date sort key and would copy a Go rule; a separate engine PR — the user chose to keep it in S13-12; sorting in `kinship.FamilyUnits` — it holds ids, not dates |
| D-3 | Recorded and implied relatives draw the same (user, 2026-11-01). The section reads no field of `CatalogFamilyLink` | The user's call | A dashed or `lineInferred` line, a Badge |
| D-4 | The tile is a snowflake: `PersonTile` *(new)* and its pure `PersonTileContent` *(new)* in `macos/App/Features/Conclusions/PersonTile.swift`. S13-13 moves it to `DesignSystem/Recipes/` when it adds the second caller | Working rule "One person tile, built once" and `add-ui-component` (recipe at the second call site) | A recipe now — one caller |
| D-5 | `PersonTileContent(summary:locale:orders:role:)` builds title (`PersonTitleDisplay.titleSource`), b. and d. lines (D-6), thumb (`pictureImageRelPath`, `.subjectPerson` mark, `PersonGenderTreatment.read(summary)?.thumbnailColors`) and the spoken label. The view only lays it out | Pure and testable like `PersonDetailContent` | Formatting in the view |
| D-6 | `PersonDetailContent.LifeLine` and `lifeLine(_:spoken:facts:locale:)` move to `PersonLifeDisplay` as `PersonLifeDisplay.LifeLine` and `PersonLifeDisplay.lifeLines(_ summary: CatalogPersonSummary?, locale:)` *(moved)*, returning the b. then d. line. `PersonDetailContent.lifeLines` calls it. The tile draws a line only when its date or place is set | One builder for the same line in the header and the tile | A second copy in the tile |
| D-7 | The section's content is `PersonFamilyContent` *(new, pure)* in `macos/App/Features/Conclusions/PersonFamilyContent.swift` *(new)*, built from `CatalogPersonFamily` and the person's own `CatalogPersonSummary` (from `CatalogConclusionDetail.summary?.person`): one display unit per `CatalogFamilyUnit` in order, each with its partner tiles, its children tiles, whether the focus sits among the children (`.parents`, `.siblings`) or beside the partners (`.partner`, `.oneParent`), and a group label. Roles: partners of `.parents` → Parent; children of `.parents` and `.siblings` → Sibling; partners of `.partner` → Partner; children of `.partner` and `.oneParent` → Child | Go groups; Swift places (working rule "Engine PRs run ahead of their slice"). The model is where the roles and focus placement are tested without a view | Regrouping in Swift; reading `link` to decide roles |
| D-8 | A section reports its own refresh through a preference, `ConclusionPageSectionRefreshingKey` *(new)* in `ConclusionDetailPage.swift` (Bool, reduced with `||`). `ConclusionDetailLoaded` reads it with `onPreferenceChange` into `@State` and passes `refreshing || sectionRefreshing` to `ConclusionDetailHeaderBand`. The Family section sets it from its handle's `isFetching` | FS-12: one refresh indicator, in the switcher row. Timeline (S13-13) and Sources (S13-16) read their own keys and use the same preference | A spinner inside the section — FS-12 says not a second spinner; passing the family handle up through the page — the page would know every section's key |
| D-9 | `PlaceRegistry` Spec `.personDetail` also warms `.personFamily(project:entityId:)` when `location.pageSection == .family` | AC-4.2: a tile click and Back start both reads at once (`add-workspace-place`: the place warms its keys) | Querying only from the view's `.task` — the family read would wait for the detail read |
| D-10 | The section reads its key the way `ConclusionDetailPage` reads its own: `session.queryHandle(key)` for the value and `.task { let _: QueryHandle<CatalogPersonFamily> = session.query(key) }`. States: no handle or first load → `ProgressView().controlSize(.small)`; failure → `PVCallout(tone: .danger, message:)` with the presentation's mapped message (as `ConclusionDetailPresentation` maps it); value with no units → `PVEmptyState`; else the diagram | Same states and mapping as the page (FS-11, FS-12) | A bespoke loader |
| D-11 | `PersonDetailSections` gains `session: WorkspaceSession`, passed from `PersonDetailView`. `ConclusionDetailPageSectionsContext` is unchanged | Only the Person page's sections read their own keys; Event and Place keep the default fill | Putting the session in the shared context |
| D-12 | Lines are drawn inside the section with SwiftUI anchor preferences: each tile publishes its bounds (`anchorPreference`) and an `overlayPreferenceValue` draws `Path`s in one existing color token the board names (default `PVColor.borderStrong`), 1pt. No `Canvas` host, no `GraphCanvas` | R-16: a plain laid-out diagram; anchors keep lines on tiles as rows wrap | `GraphCanvas`; fixed coordinates that break when a row wraps |
| D-13 | Tile size, thumb size, unit arrangement, the focus tile's style and the empty-state copy come from the agreed board. Until then the plan uses the brief's recommendations: units stacked in order (FS-9), primary-only at both widths (FS-W2), `PVCard` `.card` for relatives and `.raised` with an inked border for the focus | S13-D3 decides layout | Guessing past the board |
| D-14 | No cap on tiles: every member Go returns is drawn, in `PVFlowLayout` rows | S13-04 bounds what Go returns (`Rules.MaxPivotEnd`, `kinship.MaxPairSteps`); 50 tiles lay out (§ 6 Scale) | A "show more" — not in the brief |

**Rules this plan bends.** S13-D1 PPF-5 (archived brief: "Opening a Person from anywhere lands on Vitals", naming a family tile) → D-1 opens family tiles on Family → user, 2026-11-01 → the planning commit adds the working rule "A family tile opens the relative on Family" to the map. No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `conclusionsummaries.BornBefore` *(new)* | func | `eventKeyLess`, `dateSort` (same file) | Reuses `dateSort`; a person's birth is a different input from an event row |
| `persondetail.sortChildren` *(new, unexported)* | func | `dropUnread` (same file) | Runs after it on the same units |
| `PersonLifeDisplay.lifeLines` *(moved)* | func | `PersonDetailContent.lifeLine` (private) | Moved, not copied (D-6) |
| `PersonTile`, `PersonTileContent` *(new)* | snowflake + pure model | `ConclusionDetailHeader` (one person, page-size), `ConclusionDetailGallerySection` tiles (pictures, not people), `PVOmnibarHitRow` (a row, not a tile) | None draws one person as a card with likeness, title and two life lines |
| `PersonFamilyContent` *(new)* | pure model | `PersonDetailContent` | A different read (`CatalogPersonFamily`) |
| `PersonFamilySection` *(new)* | snowflake view | `ConclusionDetailGallerySection` | Draws people and lines, not pictures |
| `ConclusionPageSectionRefreshingKey` *(new)* | preference key | `ConclusionDetailHeaderCompactKey` (environment, page → header) | Flows the other way, section → page |
| `L10n.Conclusions.family*` *(new)* | L10n keys | `L10n.Conclusions.pageSection*` | New copy for the section |

**Builds on.** `CatalogQueryKey.personFamily` and its descriptor (S13-04), `PlaceRegistry` Specs, `WorkspaceLocation.personDetail(…pageSection:)`, `PersonGenderTreatment`, `CachedThumbnail`, `PVCard`, `PVFlowLayout`, `PVEmptyState`, `PVCallout`, `ConclusionDatePlaceLine`.
**Deletes.** `PersonDetailContent.LifeLine` and its private `lifeLine` (moved to `PersonLifeDisplay`). `.family` leaves the `pageSectionLater` empty case in `PersonDetailSections` (the key stays for Timeline and Sources).
**Next adds.** S13-13 places `PersonTile` for participants and moves it to `DesignSystem/Recipes/PersonTile/` *(new)*; S13-13 and S13-16 set `ConclusionPageSectionRefreshingKey` from their own handles.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A family tile opens a Person on Family | `PersonDetailView` / `ConclusionDetailPage`, keyed by entity id | The page is keyed by entity id (`viewIdentity: { $0.entityId }`), so the lead model and scroll reset; a relative merged since the read returns `conclusiondetails.not_found`, which the page shows as its Callout; history decodes `pageSection` (S13-10 AC-2.4) |
| Back to a Family entry | same | D-9 warms both keys; the family key is `.evict` while hidden, so Back reloads it (a spinner, then the diagram) |
| A family with units but every unit's partners empty except the focus | `PersonFamilyContent` | `.oneParent` and `.siblings` have no partners by kind; Go drops empty units (`dropUnread`), so no empty unit reaches Swift |

### Scale

At the brief's sizes (§ 2: typical 10 – 20 tiles, large about 30, malformed 50): one family read per page view, warm with no graph query (S13-04's `family_query_test.go`); the sort in D-2 is O(n log n) over each unit's children, no query. Swift draws one `PersonTile` per member; each `CachedThumbnail` loads through `ThumbnailCache`, as the gallery does. The key refreshes on any touched handle only while visible (`.evict` when hidden). T-9 builds `PersonFamilyContent` for 50 members.

### Edges

- **Assumptions.** The person is in no unit (S13-04 `FamilyUnits`); the section inserts the focus by kind (D-7). A relative appears once (S13-04 places a sibling once); the model keys tiles by `entity.id` per unit, so a repeat across units would still render. T-6 holds the focus placement for each kind.
- **Neighbors.** A `.parents` unit with one partner (lone parent) → one parent tile over the children (AC, T-6). A `.partner` unit with no children → the focus and the partner, no children row (T-6). A `.partner` unit whose member has `link == nil` (co-parent) → drawn as any partner (D-3, T-6).
- **Removed inputs.** A relative unpromoted or merged after the read: the next refresh (any touched handle) drops them; the click before that lands on the page's not-found Callout.
- **Real data.** None — no table or seed drives the section.
- **Untrusted size.** Bounded in Go by S13-04 (parents per person at `Rules.MaxPivotEnd`; duplicate steps at `kinship.MaxPairSteps`); children per unit are not bounded there and D-14 draws them all (§ Scale).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionsummaries` | `func BornBefore(a, b PersonSummary) bool` *(new)* in `compose.go` beside `eventKeyLess`: a person with a birth date (`dateSort(a.Birth.Date)` ok) sorts before one without; then by key; then by lower-cased ref |
| `core/database/persondetail` | `func sortChildren(family Family, summaries conclusionsummaries.Summaries) Family` *(new)*: `sort.SliceStable` on each unit's `Children` by `BornBefore` of their summaries. `PersonFamily` returns `sortChildren(dropUnread(family, summaries), summaries)`. Doc comment on `PersonFamily` says children are in birth order |

### FFI

None — the proto comment on `GetPersonFamilyResponse` gains one clause: "children in each unit in birth order, undated last". No regeneration needed beyond the comment (run `./scripts/generate-proto.sh` so the generated comments match).

### macOS

Until the board is agreed, the frames are described by the brief's requirements; on agreement this table and the frame list below are refreshed from the board (S13-D3 § 7).

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Family section | snowflake | `macos/App/Features/Conclusions/PersonFamilySection.swift` *(new)* | D-10 states; units per D-7, D-13; lines per D-12; sets `ConclusionPageSectionRefreshingKey` (D-8). Identifier `persons.detail.family` |
| Unit | snowflake (private in the section) | same file | Partners row (focus beside, for `.partner` / `.oneParent`), children row in `PVFlowLayout` (focus among them for `.parents` / `.siblings`). `.accessibilityElement(children: .contain)` with the unit's group label |
| Person tile | snowflake | `macos/App/Features/Conclusions/PersonTile.swift` *(new)* | `PVCard` > `CachedThumbnail` + `ConclusionListRow.title` (two lines, `.help(fullTitle)`) + `ConclusionDatePlaceLine` per D-6 line. A relative's tile is a `Button` (plain style, kit focus ring) calling `navigation.go(to:)`; the focus tile is not a button and is `.accessibilityHidden(false)` static text. Identifier `persons.detail.family.tile.<entityId>` |
| Lines | snowflake | `PersonFamilySection.swift` | `anchorPreference` + `overlayPreferenceValue`; `.accessibilityHidden(true)` |
| Empty | `PVEmptyState` | — | Glyph `ConclusionPageSection.family.glyph` |
| Error | `PVCallout(tone: .danger)` | — | — |
| Refresh | the switcher row (D-8) | `ConclusionDetailPage.swift` | — |

States: loading spinner; error Callout; empty EmptyState; large (50 tiles) wraps and scrolls inside Page columns' primary column. Accessibility: one button per relative tile with the spoken label `L10n.Conclusions.familyTileAccessibility(role:title:life:)`; units are groups; Tab order follows reading order; lines hidden. Wide: primary-only, centered, as the board's FS-W1 decides.

### Strings

L10n keys *(new)*, under `L10n.Conclusions` (final copy from the board; these are the brief's recommendations):

- `familyEmptyTitle` = "No family yet"
- `familyEmptyMessage` = "Relatives appear here when people linked to this person in an Evidence graph are promoted."
- `familyRoleParent` = "Parent", `familyRoleSibling` = "Sibling", `familyRolePartner` = "Partner", `familyRoleChild` = "Child"
- `familyGroupParents` = "Parents", `familyGroupPartner(_ name:)` = "%@ and children", `familyGroupChildren` = "Children", `familyGroupSiblings` = "Siblings"
- `familyTileAccessibility(role:title:life:)` = "%1$@, %2$@, %3$@"

Removed: none (`pageSectionLater` stays for Timeline and Sources).

## 7. Steps

**STEP-1 — Children in birth order (Go)**
Files: `core/database/conclusionsummaries/compose.go`, `core/database/conclusionsummaries/conclusionsummaries_test.go`, `core/database/persondetail/family.go`, `core/database/persondetail/family_test.go`, `api/proto/engine.proto` (comment). Skill: `add-catalog-query`.
Do: D-2 and § 6 Go core; T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/... ./api/ffi/...` passes; `./scripts/generate-proto.sh` leaves only the comment changed.

**STEP-2 — One life-line builder**
Files: `macos/App/Features/Conclusions/PersonLifeDisplay.swift`, `PersonDetailContent.swift`, `ConclusionDetailPage.swift` / `PersonDetailView.swift` (references to `PersonDetailContent.LifeLine`). Skill: `add-swift-test`.
Do: D-6. No visible change.
Check: `PersonDetailTests` pass unchanged; T-3.

**STEP-3 — The person tile**
Files: `macos/App/Features/Conclusions/PersonTile.swift` *(new)*. Skill: `add-ui-component`, `add-localized-string`.
Do: D-4, D-5; the view per § 6 macOS; a `#Preview` with the FS-2 states.
Check: T-4, T-5 pass.

**STEP-4 — The section's content model**
Files: `macos/App/Features/Conclusions/PersonFamilyContent.swift` *(new)*.
Do: D-7.
Check: T-6, T-7, T-9 pass.

**STEP-5 — The section view and its refresh**
Files: `macos/App/Features/Conclusions/PersonFamilySection.swift` *(new)*, `ConclusionDetailPage.swift` (`ConclusionPageSectionRefreshingKey`, `ConclusionDetailLoaded`).
Do: D-8, D-10, D-12, D-13, D-14.
Check: the app builds; a `#Preview` of the section on the § 5 sample family renders; T-10 passes.

**STEP-6 — Wiring and warm**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`. Skill: `add-workspace-place`, `add-workspace-location`.
Do: D-1, D-9, D-11; `.family` draws `PersonFamilySection`.
Check: T-8, T-11 pass.

**STEP-7 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: § 6 Strings.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/S13-D3-family-section.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-12-family-section.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestBornBefore` *(new)* | Table: dated vs undated → dated first; two dates → earlier first; `ABT 1820` vs `1820` → the sort key's order; two undated → by ref | D-2 |
| T-2 | Go | `core/database/persondetail/family_test.go` `TestPersonFamily` (new case "children in birth order") | Siblings filed Friedrich (1820), Wilhelm (undated), Margarethe (1815) → children Margarethe, Friedrich, Wilhelm; a second read issues zero queries | AC-1.4 |
| T-3 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeLinesFromSummary` *(new)* | A summary with birth date and place, no death → two lines; the b. line's date and place set, the d. line's nil | D-6 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` *(new)* `tileContentFromSummary` | Anna (AC-2.1), Wilhelm (AC-2.2), a ref-only child (AC-2.3) → title source, lines, mark, colors | AC-2.1 – AC-2.3 |
| T-5 | Swift | same file `tileSpokenLabel` | Anna as Parent → "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851" | AC-2.4 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonFamilyContentTests.swift` *(new)* `unitsPlaceTheFocusByKind` | One unit of each kind, a lone parent, a partner with no children, a co-parent (`link` nil) → focus among the children for `.parents` and `.siblings`, beside the partner for `.partner` and `.oneParent`; roles per D-7 | AC-1.1 – AC-1.3, D-3, D-7 |
| T-7 | Swift | same file `tilesOpenTheRelativeOnFamily` | A member → `WorkspaceLocation.personDetail(entityId:ref:title:pageSection: .family)`; the focus has none | AC-3.1, D-1 |
| T-8 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnFamilyWarmsTheFamily` *(new)* | A Person location with `pageSection: .family` → keys `.conclusionDetail` and `.personFamily`; on Vitals → `.conclusionDetail` only | AC-4.2, D-9 |
| T-9 | Swift | `PersonFamilyContentTests` `fiftyMembersKeepOrder` | 50 children in one unit → 50 tiles in the order given | D-14 |
| T-10 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` or a page test `sectionRefreshingReachesTheSwitcher` *(new)* | The preference reduce: false, true → true | D-8 |
| T-11 | Swift | `PersonFamilyContentTests` `emptyFamilyIsEmpty` | `units: []` → the empty state | AC-1.5 |

AC-1.6, AC-2.5, AC-3.2, AC-3.3 and AC-4.1 are covered by the manual check (§ 10). Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PersonDetailTests` (STEP-2), `TestPersonFamily` and `TestGetPersonFamily` (STEP-1), the place registry tests (STEP-6).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonDetailContent.LifeLine` and `PersonDetailContent.lifeLine` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (done in the planning commit: the working rule "A family tile opens the relative on Family")
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Bar 3 and the slice's Check. Use a project with the § 5 sample family promoted (or the product owner's own test people with two marriages and a sibling known only through a shared parent).

1. (AC-1.1, AC-2.1, AC-2.2, AC-2.5) Open Johann, choose Family. See one parent unit, two partner units and Paul's unit; Anna's tile with her likeness, plum tint, name and lines; Wilhelm's tile with the mark and his name alone; Johann's tile set apart and not clickable.
2. (AC-1.4) Margarethe, Friedrich, Wilhelm appear in that order.
3. (AC-1.2) Add a baptism naming Katharina as mother, promote; Johann's Family shows two parent units.
4. (AC-3.1, AC-3.2, AC-3.3) Click Anna: her Family opens, breadcrumbs *Persons › Anna Weber › Family*. Back: Johann's Family. Tab to Friedrich, press Return: his Family opens.
5. (AC-4.1) With Johann's Family showing, rename Anna on her page in another place in history, come back: the tile shows the new name; while it reloads, "Refreshing" shows in the switcher row.
6. (AC-1.5) Open a person with no promoted relatives: the EmptyState.
7. (AC-1.6) Open a Family entry from history for a person merged since: the danger Callout.
8. (Bar 8) Run Promote on a Source as before the spike: the proposals are unchanged (no sibling evidence).
9. Resize across 1600pt on Family: section, scroll and focus kept (FS-W3). Dark mode: frames 1, 3, 9 of the board match.
10. VoiceOver on Anna's tile reads "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851, button".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Lines drawn from anchors jitter or misplace when rows wrap at a resize | D-12 anchors to tile bounds, recomputed on layout; manual check 9 |
| The S13-04 branch lands with a different Swift shape | Re-check § 4 rows against `main` after it lands (Implementer step 2) |
| `PVFlowLayout` cannot report rows for line routing | The lines connect a unit's partners to the unit's children block as a whole, not to each wrapped row, unless the board asks otherwise |

**Stop and report if:**

- S13-04 has not landed on `main`, or `CatalogPersonFamily`, `CatalogFamilyUnit.Kind` or `.personFamily` differ from § 4.
- The agreed board draws a recorded / implied difference, a placeholder for an unknown relative, or a secondary column (they contradict D-3, R-2, FS-W2).
- The board needs data `GetPersonFamily` does not return (for example a marriage date to order partners): that is a follow-up engine PR, not this one.
- `onPreferenceChange` cannot carry the section's refresh to the header band across `PVPageColumns` (D-8): report what `PVPageColumns` does with preferences.
- A Go test outside `persondetail` and `conclusionsummaries` changes because of STEP-1.

## 12. Open questions

1. **The board (S13-D3).** Tile and thumb size, the focus tile's style and whether it repeats per unit, how two parent sets sit and their caption, the line token, the empty copy, and what sits side by side at Wide. Recommended: the brief's recommendations, which D-13 drafts. Closed when the board is agreed.

## 13. PR

**Title:** S13-12 — Family section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-12-family-section.md`.
