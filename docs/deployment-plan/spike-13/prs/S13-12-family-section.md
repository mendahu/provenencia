# S13-12 — Family section

**Status:** Blocked — on board S13-D3
**Spike / slice:** Spike 13 · Slice 3 — Family
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D3](../design/S13-D3-family-section.md), board agreed 2026-11-01
**Board:** agreed 2026-11-01 — Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Family.dc.html*, turn 11 "S13-D3 · Family in two columns — origin and own family", frames 11a – 11q, decisions FC-1 – FC-5. The implementer may not be able to open it: **Design → macOS** describes every frame in words and gives every measure, and D-12 is the layout
**Depends on:** S13-04 (landed [#671](https://github.com/mendahu/provenencia/pull/671)), S13-10 (landed [#634](https://github.com/mendahu/provenencia/pull/634)), S13-D3 (board)
**Blocks:** S13-13 (it reuses the person tile and the section-key mapping)
**Branch:** `s13-12-family-section`
**Runs on:** macOS: STEP-1 needs `protoc` and its plugins (`./scripts/generate-proto.sh`), and STEP-2 – STEP-7 and `scripts/check-pr.sh` need Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-01 and S13-04 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (STEP-1)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [person-detail.md § 2. Family](../person-detail.md#2-family) (R-2, R-3, R-16 – R-22); brief [S13-D3](../design/S13-D3-family-section.md); the S13-04 plan [`S13-04-person-family-read.md`](S13-04-person-family-read.md) and its **Landed** note; working rules: "Engine PRs run ahead of their slice" (data, not layout), "Each section reads its own query", "Family groups in Go", "One person summary for every tile", "Canonical only, filtered in Go", "One person tile, built once", "Colors reuse `PersonGenderTreatment`", "A family tile opens the relative on Family", "Sections are places"
**Planned:** 2026-11-01 against `main` @ `75d9494` with the S13-04 branch; refreshed 2026-11-01 against `main` @ `be51621`, after S13-04 landed
**Review:** [`S13-12-family-section.review.md`](S13-12-family-section.review.md) — round 1 Fits with changes / Not ready; every finding Fixed; refreshed after S13-04 landed; status stays Blocked on the board

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

**What stays.** The header, the switcher, Vitals, Gallery and the Timeline and Sources empty places; `GetPersonFamily`'s units, their order and their filtering (S13-04), apart from D-2 and D-3; `persondetail.Families`, the tree's reusable step; the gender colors; every other link into a Person page, which still lands on Vitals.

## 2. User stories

**US-1.** As a researcher, when I open a Person's Family section, I see their nuclear family laid out by family, so that I understand who they grew up with and who grew up with them.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (b. 1817; S13-D3 § 5 sample): parents Heinrich and Anna; siblings Margarethe (1815), Friedrich (1820), Wilhelm (undated); wives Maria (children Karl, Luise, Emma) and Sophie (child Otto); son Paul with no other parent | Frame 11a. Left column under "PARENTS": Heinrich, Anna stacked; indented beneath on a rail: Margarethe, **Johann** (focus, "This person"), Friedrich, Wilhelm. Right column under "PARTNERS AND CHILDREN": Maria with Karl, Luise, Emma indented beneath; Sophie with Otto; Paul with no partner tile. One line from Johann crosses the gutter to Maria, Sophie and Paul (FS-4, FS-6, FS-7, FC-1, FC-4) |
| AC-1.2 | Johann with a second parent set (Heinrich and Katharina); Friedrich in both sets; Christian (1823) in the second only | Frame 11c. "PARENTS" (Heinrich, Anna) then "OTHER PARENTS" (Heinrich, Katharina) stacked, each set on its own rail; beneath, each sibling once: Johann, Friedrich (both sets), Christian (second only), Margarethe, Wilhelm (first only) (FS-5, FC-3, R-19) |
| AC-1.3 | Clara (1841), whose only family is recorded siblings Ida (1838) and Bernhard (1844) | Frame 11d. Under "SIBLINGS", not indented: Ida, Clara (focus), Bernhard on one rail; right column empty (FS-8) |
| AC-1.4 | Go returns the siblings in link order Friedrich, Wilhelm, Margarethe | Shown Margarethe, Johann, Friedrich, Wilhelm: earliest born first, the focus by its own birth, undated last (D-2, D-3) |
| AC-1.5 | A person with no promoted relatives (`units` empty) | The EmptyState of FS-11 |
| AC-1.6 | The family read fails with `conclusiondetails.not_found` | Frame 11o. A danger Callout in the section titled "Family could not be read" with the mapped message; the header and switcher still show |
| AC-1.7 | Anna Weber, with no parents or siblings on record and Heinrich as partner | Frame 11h. Her focus tile alone on the left; on the right "PARTNER AND CHILDREN": Heinrich with Margarethe, Johann, Friedrich, Wilhelm beneath (FC-1) |

**US-2.** As a researcher, each relative's tile tells me who they are at a glance.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Anna: woman, a likeness, b. 3 May 1790 · Bremen, d. 1851 (no death place) | Frame 11l. A 92pt tile: the 70pt likeness outlined in the woman line, "Anna Weber", "b. 3 May 1790 · Bremen", "d. 1851" (no "place unknown") (FS-1, FC-5, D-6) |
| AC-2.2 | Wilhelm: no gender, no likeness, no dates or places | The person mark with the unknown treatment's colors, the name alone, no b. or d. line |
| AC-2.3 | A child known only by ref `PER-7KQ2M` | The ref in mono as the title |
| AC-2.4 | VoiceOver on Anna's tile | "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851", a button (FS-13, D-7) |
| AC-2.6 | Hover over Friedrich's tile | It grows 2.5% with the kit's medium shadow; keyboard focus draws the kit focus ring (frame 11l) |
| AC-2.5 | The focus tile (Johann on his own page) | A raised, elevated card, the name semibold with a neutral "This person" Badge beside it; not a button, no focus stop; VoiceOver reads "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person" as static text (FS-3) |

**US-3.** As a researcher, a click on a relative takes me to their family.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | On Johann's Family, click Anna's tile | Anna's Person page opens on Family; breadcrumbs *Persons › Weber, Anna › Family* (the sorted title `PersonsList.location` gives, D-1) |
| AC-3.2 | Then Back | Johann's Family again |
| AC-3.3 | Tab to Friedrich's tile, press Return | Frame 11q: Tab runs the left column top to bottom, then the right column partner by partner, skipping the focus; Return opens Friedrich's Family (FS-13) |

**US-4.** As a researcher, the section stays current while I work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Johann's Family is showing; a write touches Anna's handle | The family reloads; the diagram stays drawn and the switcher row shows "Refreshing" until it lands (FS-12, D-8) |
| AC-4.2 | Opening a Person straight on Family (from a tile, or Back) | The family read starts with the page, not after the detail read (D-8) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Children in each unit in birth order, in Go (D-2) | STEP-1, AC-1.4 |
| The person among the children of their parent and sibling units, and the person's own summary on the response (D-3) | STEP-1, AC-1.1, AC-1.3 |
| The b. and d. line: one builder with a stated or omitted missing half (D-6) | STEP-2, AC-2.1 |
| Two-line titles on `ConclusionListRow.title` (D-5) | STEP-2 |
| The person tile and its button style (FS-1, FS-2, FS-3) | STEP-3, AC-2.1 – AC-2.5 |
| The section's content model: units, roles, locations, accessibility (FS-4 – FS-8, FS-13) | STEP-4, AC-1.1 – AC-1.3, AC-2.4, AC-3.1 |
| One read-state mapping for sections (D-10) | STEP-5 |
| A page section's own query key: warm, read, refresh (D-8) | STEP-5, STEP-6, AC-4.1, AC-4.2 |
| The two-column layout and its lines, as a pure function (FC-1 – FC-4, D-12) | STEP-4, AC-1.1 – AC-1.3, AC-1.7 |
| The section view: tiles, lines, states (FS-4 – FS-12, FS-14, FS-W1 – FS-W3) | STEP-6, AC-1.5, AC-1.6 |
| The pattern recorded in `docs/macos-client-patterns.md` (D-8) | STEP-7 |

| Out | Home |
| --- | --- |
| Grandparents and beyond; a tree view | [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md) |
| Marking recorded versus implied links | not planned (user, 2026-11-01); the data stays on `CatalogFamilyLink` for a later tree |
| The provenance control on a tile or line | not planned this spike (R-13b) |
| The tile in the Timeline, and its move to `DesignSystem/Recipes/` | S13-13 |
| Ordering partner units (by marriage date) | not planned; units keep S13-04's order |
| Editing a family from the section | not planned this spike |
| Opening a Person on Family from anywhere other than a family tile | not planned; other links keep Vitals (S13-D1 PPF-5) |
| `EvidenceGraphModel`'s own hand-built `.personDetail(` location | not this PR; logged in the review log |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Family section | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections` | `.family, .timeline, .sources` draw one `PVEmptyState` with `L10n.Conclusions.pageSectionLater`. It receives `content` and `context`; `PersonDetailView.entityId` is not passed |
| Section context | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPageSectionsContext` | `projectDir`, `refreshing` (the detail read's), `pageIdentifier`, `lead`, `pageSection`; no entity id |
| Switcher refresh | same file `ConclusionDetailHeaderBand.switcher(selection:)` | Shows `ProgressView` + `L10n.Conclusions.detailRefreshing` when `refreshing`, which `ConclusionDetailLoaded` takes from the `conclusionDetail` presentation |
| Read states | same file `ConclusionDetailPresentation<Content>` | `.loading` / `.failed(String)` / `.content(Content, refreshing:)`; `init(_ read: QueryRead<CatalogConclusionDetail>, isFetching:content:)`, failure through `L10n.Errors.message(for:)` |
| Session in views | `macos/App/Features/Workspace/WorkspaceView.swift` | Injects `.environment(session)`; `SourcePageView` and others read `@Environment(WorkspaceSession.self)` |
| Family read (Swift) | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonFamily` (`units`), `CatalogFamilyUnit` (`Kind`: `.parents`, `.partner`, `.oneParent`, `.siblings`), `CatalogFamilyMember` (`person: CatalogPersonSummary`, `link: CatalogFamilyLink?`); `GoStore.getPersonFamily`, `mapFamilyUnit` | Units in Go's order. The person is in no unit; the response has no summary of the person |
| Family key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.personFamily(project:entityId:)`; `CatalogQueryRegistry.swift` its descriptor | `.sessionFresh`, `.evict` when hidden, refreshes on any touched handle. No view queries it |
| Place warm | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` Spec `.personDetail` | Warms only `.conclusionDetail` |
| Family read (Go) | `core/database/persondetail/family.go` `PersonFamily`, `dropUnread` | Units with children in link order (`kinship.FamilyUnits`), summaries of the relatives only |
| Handler | `api/ffi/handlers/person_detail.go` | Maps units and members (`personSummaryProto(h, paths)`) in the order given |
| Proto | `api/proto/engine.proto` `GetPersonFamilyResponse` | `repeated FamilyUnit units = 1;` |
| Date order | `core/database/conclusionsummaries/compose.go` `eventKey`, `eventKeyLess`, `eventKeyOf`, `sortEvents`, `dateSort` | Events: keys computed once per row, dated first, then key, then lower-cased ref |
| Title order | `core/database/conclusionsummaries/order.go` `SortByTitle` | "The one owner of title order; the app shows rows as given" |
| Life lines | `macos/App/Features/Conclusions/PersonDetailContent.swift` `LifeLine`, `lifeLine(_:spoken:facts:locale:)` (private); used by `PersonDetailView.swift` `PersonDetailLifeLines` | Header lines; a missing half is spoken as unknown (`a11yLifeLine`) |
| Line view | `macos/App/Features/Conclusions/ConclusionDatePlaceLine.swift` `ConclusionDatePlaceLine` | Draws *b. date · place*; a nil half draws `dateUnknown` / `placeUnknown` |
| Title | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionListRow.title(_:extraCount:)` | `.lineLimit(1)` inside; name / italic label / mono ref |
| Person link | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList.location(_:orders:)`; `WorkspaceLocation.withPageSection(_:)` | The summary → Person page location, sorted title |
| Gender colors | `macos/App/Features/Conclusions/PersonGenderTreatment.swift` `read(_:)`, `thumbnailColors` | Reads a `CatalogPersonSummary` |
| Thumb, card, wrap | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; `macos/App/DesignSystem/Components/Card/PVCard.swift` `PVCard` (`tone`, `border: .solid / .dashed`, `elevated`); `macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift` `PVFlowLayout` (leading-aligned) | — |
| Button chrome | `macos/App/DesignSystem/Components/HoverEffect/PVHoverEffect.swift`; precedent `macos/App/Features/Sources/SourcesGraphButton.swift` `SourcesGraphButtonStyle`, `SourcesGraphButtonBody` | Hover, press and focus ring for a feature button |
| Test fixture | `core/database/graphcache/graphcachetest/graphcachetest.go` `Canon` (`Birth`, `Term`, private `observe`) | No helper sets a date on an entity |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `PlaceRegistryTests.swift`; `core/database/persondetail/family_test.go` `TestPersonFamily` (its `kinWorld.show` sorts names), `family_query_test.go` `TestPersonFamilyQueryCount`, `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personFamilyFollowsEveryTouchedHandle` | — |

**Precedent.** S13-10 ([#634](https://github.com/mendahu/provenencia/pull/634)) filled the Person page's sections inside `PersonDetailSections`; copy its shape. S13-11 ([#650](https://github.com/mendahu/provenencia/pull/650)) navigated from a section with `navigation.go(to:)` from the environment. `ConclusionDetailGallerySection` lays out a wrapping grid of `CachedThumbnail` tiles. `SourcesGraphButtonStyle` is the button style to copy for the tile.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A relative's tile opens `PersonsList.location(summary, orders:).withPageSection(.family)` with one `navigation.go(to:)` (user, 2026-11-01) | R-20; walking a family stays in Family (working rule "A family tile opens the relative on Family"); one place decides a Person link's history entry and title | Vitals (S13-D1 PPF-5), the user chose Family; building `.personDetail(` by hand |
| D-2 | Go orders each unit's **children** by birth, in the page read only. In `conclusionsummaries`, `eventKey` / `eventKeyLess` are renamed `dateKey` / `dateKeyLess` (same fields, same rule); `birthKeyOf(PersonSummary) dateKey` *(new)* reads `dateSort(h.Birth.Date)`; `SortByBirth[T any](rows []T, summary func(T) PersonSummary)` *(new)* in `order.go` beside `SortByTitle` computes each key once and sorts stable with `dateKeyLess`. `persondetail.PersonFamily` sorts each unit's `Children` with it (user, 2026-11-01) | Order is data, and the date sort key lives in Go; one date-order comparison in the package for events and births | Sorting in Swift (no date key there); a separate engine PR (the user chose this PR); sorting inside `kinship.FamilyUnits` (it holds ids); a `BornBefore` comparator (a second copy of `eventKeyLess`, keys recomputed per comparison) |
| D-3 | In the page read, `PersonFamily` puts the person among the **children** of each `.parents` and `.siblings` unit as a member with no link, before D-2's sort, so the person falls in birth order with the siblings. It adds the person to the summaries it reads, and `GetPersonFamilyResponse` gains `PersonSummary person = 2;`, the person's own summary (`CatalogPersonFamily.person: CatalogPersonSummary?` *(new)*). Swift finds the focus by `entity.id == family.person?.entity.id`. `persondetail.Families` is unchanged | The focus's place among siblings is birth order, which only Go can compute; the section needs the person's summary for the focus tile | Swift inserting the focus (no date key); a per-unit focus index on the wire (a second ordering field) |
| D-4 | Recorded and implied relatives draw the same (user, 2026-11-01). The section reads no field of `CatalogFamilyLink` | The user's call | A dashed or `lineInferred` line, a Badge |
| D-5 | The tile is a snowflake in `macos/App/Features/Conclusions/PersonTile.swift` *(new)*: `PersonTile` (view), `PersonTileContent` (pure), `PersonTileRole` (`enum { parent, sibling, partner, child }` with `label: LocalizedStringResource`), and a private `PersonTileButtonStyle` composed like `SourcesGraphButtonStyle`: `PVHoverEffect` gives the hover; a hovered or pressed tile takes `.scaleEffect(1.025)` and `.pvShadow(PVElevation.md)`, animated with `PVMotion.easeStandard` (frame 11l); keyboard focus draws `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)`. The tile (FC-5, frame 11l) is 92pt tall at the width D-12 gives: `PVCard(cornerRadius: PVRadius.sm, padding: 10)`, inside an `HStack(spacing: 10)` of the 70pt `CachedThumbnail` (`size: 70`, `cornerRadius: PVRadius.sm`, `colors` from the gender treatment) and a `VStack(alignment: .leading, spacing: 3)` centred beside it: the title through `ConclusionListRow.title`, which gains `lineLimit: Int = 1` (the tile passes 2, semibold for the focus), then the D-6 lines. The focus tile is `PVCard(tone: .raised, cornerRadius: PVRadius.sm, elevated: true, padding: 10)` with a neutral `PVBadge(L10n.Conclusions.familyThisPerson)` after the title; the board's inset fill is not drawn (no `PVCard` tone has it). `.help(title.text)` on every tile. S13-13 moves the tile to `DesignSystem/Recipes/PersonTile/` *(new)* with the second caller | Working rule "One person tile, built once"; `add-ui-component` (recipe at the second call site); one title styling; one hover and focus chrome; the board's measures | A recipe now (one caller); a copy of `titleText`; hand-written `onHover` / `@FocusState`; a new `PVCard` tone for the inset fill (a card style for one caller, which the brief forbids) |
| D-6 | `PersonDetailContent.LifeLine` and `lifeLine` move to `PersonLifeDisplay` as `PersonLifeDisplay.LifeLine` and `PersonLifeDisplay.lifeLines(_ summary: CatalogPersonSummary?, locale:missing:) -> [LifeLine]` *(moved)*, with `enum Missing { case stated, omitted }`. `.stated` is today's header (both lines always, a missing half "unknown"). `.omitted` returns only lines with a date or a place and leaves a missing half nil, and its spoken text drops the missing half ("died 1851"). `ConclusionDatePlaceLine` gains `var missing: PersonLifeDisplay.Missing = .stated`; with `.omitted` a nil half and its "·" are not drawn, the line keeps to one line, and the place truncates at its tail (frames 11a, 11l). The header passes `.stated`, the tile `.omitted` | One builder and one view for the b. / d. line in the header and both tiles; the board leaves the missing half out on tiles and states it in the header (frame 11h) | A second builder in the tile; "unknown" on every tile |
| D-7 | `PersonFamilyContent` *(new, pure)* in `macos/App/Features/Conclusions/PersonFamilyContent.swift` *(new)*: `init(family: CatalogPersonFamily, locale:orders:)`, built in two columns (FC-1). **Left (`origin`)**: `parentSets: [[Tile]]`, one per `.parents` unit in order, each its partners (role parent); `siblings: [Sibling]`, each `Sibling` a `Tile` (role sibling, or the focus) and `sets: [Int]`, the indexes of the parent sets it is a child of; `siblingsCaptioned: Bool` (true when there are no parent sets). Sibling order (FC-3, one rule for any number of sets): the children found in every parent set, in the first set's order; then, walking the parent sets from last to first, each set's children not yet placed, in that set's order; then the `.siblings` unit's children not yet placed, in its order. With no parent sets, the `.siblings` unit's children in order. Go gives each unit's children in birth order with the person among them (D-2, D-3), so every group is in birth order. With no parent sets and no siblings, the focus alone (frame 11h). **Right (`formed`)**: `groups: [Group]`, one per `.partner` unit then the `.oneParent` unit, each `partner: Tile?` (role partner; nil for `.oneParent`) and `children: [Tile]` (role child). `isEmpty`: no parent sets, fewer than two left tiles, and no groups (frame 11m). `Tile` has `id` ("\(column)-\(group)-\(entityId)"), `content: PersonTileContent`, `isFocus: Bool`, `location: WorkspaceLocation?` (nil for the focus). Captions (visual, `aria-hidden` on the board): first parent set "Parents" (two parents, or a second set exists) or "Parent"; later sets "Other parents" / "Other parent"; siblings "Siblings" only when `siblingsCaptioned`; the right column one caption on its first group: "Partners and children", "Partner and children", "Partners" or "Partner" by the partner count and whether any group has children, else "Children". Group labels (VoiceOver): each parent set its caption; the siblings "This person and siblings" (or "This person" alone); each partner group "Partner Maria Schulz and children" (or "Partner Maria Schulz"); the `.oneParent` group "Children". Spoken label of a relative's tile: `a11yList(role.label, rest: a11yList(title.spokenText, rest: life))`, `life` the `.omitted` lines' spoken texts joined with `a11yList`; the focus: `a11yList(title.spokenText, rest: life)` then `familyThisPersonSpoken` ("this person") | Go groups; Swift places, as the board lays out (FC-1, FC-3). Every member named, so tests can assert | Repeating the focus per unit (the board draws it once); regrouping people Go did not group; reading `link` for roles |
| D-8 | A page section's query key is declared once: `ConclusionPageSection.queryKey(project: ProjectKey, entityId: String) -> CatalogQueryKey?` *(new)* in `macos/App/Features/Workspace/WorkspaceLocation.swift` (`.family → .personFamily`, every other section nil). Three readers: (a) `PlaceRegistry` Spec `.personDetail` appends `location.pageSection.flatMap { $0.queryKey(project:entityId:) }`; (b) `ConclusionDetailLoaded` passes `refreshing || sectionRefreshing` to the header band, where `sectionRefreshing` is `true` when the open section's key has a handle that holds a value and `isFetching`; (c) the section reads the same key | One source of truth for "is the open section refreshing"; S13-13 and S13-16 each add one case (working rule "Each section reads its own query"); a first load shows the section's spinner only, not the switcher's | A preference tunnel with a mirrored `@State` (no state preference exists in the app); a per-section branch in `PlaceRegistry` |
| D-9 | `ConclusionDetailLoaded` gains the session and the entity id it needs for (b): `ConclusionDetailPage` already holds `session` and `entityId` and passes them in | (b) needs the key and a handle | Putting the session in the sections context |
| D-10 | `ConclusionDetailPresentation.init` becomes generic: `init<Value>(_ read: QueryRead<Value>, isFetching: Bool, content: (Value) -> Content)`; existing callers compile unchanged. The Family section uses `ConclusionDetailPresentation<PersonFamilyContent>`: `.loading` → `ProgressView()` (regular size) centred, 96pt from the top (frame 11n); `.failed(message)` → `PVCallout(tone: .danger, title: L10n.Conclusions.familyErrorTitle, message:)`, id `persons.detail.family.error` (frame 11o); `.content` with `isEmpty` → `PVEmptyState` (glyph `ConclusionPageSection.family.glyph`), id `persons.detail.family.empty` (frame 11m); else the two columns, id `persons.detail.family` | One mapping from read state to view state for every section | A third hand-written switch |
| D-11 | `PersonFamilySection` reads `@Environment(WorkspaceSession.self)` and `@Environment(WorkspaceNavigation.self)`. `PersonDetailSections` gains `entityId: String`, passed from `PersonDetailView`. The section gets its handle with `session.queryHandle(key)` and starts it in `.task { let _: QueryHandle<CatalogPersonFamily> = session.query(key) }`, as `ConclusionDetailPage` does | The session is already in the environment; the key needs the id | Threading the session through `PersonDetailSections` |
| D-12 | The layout is a pure function, `PersonFamilyLayout` *(new)* in `macos/App/Features/Conclusions/PersonFamilyLayout.swift`: `init(content: PersonFamilyContent, width: CGFloat, wide: Bool)` gives `tiles: [(id, frame: CGRect)]` in reading order, `captions: [(text, origin: CGPoint)]`, `segments: [CGRect]` (1pt horizontal or vertical lines) and `height`. The view is a `ZStack(alignment: .topLeading)` of that height, placing each tile with `.frame(width:height:)` and `.offset`, captions likewise, and the segments as one `Path` filled with `PVColor.borderStrong`, `.accessibilityHidden(true)`. Measures (FC-2 – FC-5, the board's `layTree3`): `width` is the primary column's width (read with `onGeometryChange`); rail margin `L0` 20; gutter 48 at Regular, 96 at Wide; column width `TP = floor((width − 20 − gutter) / 2)`; left column at x 20, right at `c2 = 20 + TP + gutter`; tile height 92, gap 12 between tiles, caption band 26, indent 32 for children (indented tiles are `TP − 32` wide). **Left:** each parent set: caption, then its parents stacked at x 20, then 20pt before the next block; the siblings stacked at x 52 (x 20, with the "Siblings" caption, when there are no parent sets). Each parent set has a rail: x 10 for one set; x 6 and 14 for two; for n ≥ 3, `4 + round(12 · k / (n − 1))`. A set's rail joins each of its parents' vertical centres to the margin (a horizontal from the rail to x 20), runs down from its first parent's centre to its last sibling's centre, and joins each of its siblings (a horizontal from the rail to x 52); for the second and later sets, only the siblings not in an earlier set. Siblings in no shown set (recorded only) hang off a bracket at x 40 from the focus's centre down to the last of them. With no parent sets, one rail at x 10 joins every sibling. **Right:** groups top-aligned with the left (FC-4), from y 26, 28pt between groups; each partner at `c2`, its children beneath at `c2 + 32`. A bus at `c2 − 32` runs from the focus's centre to the lowest group line; a horizontal runs from the focus tile's right edge to the bus. A partner joins the bus with a horizontal at its centre; a rail at `c2 − 12` runs from the partner's centre to its last child's centre, with a horizontal to each child. A group with no partner: one child is joined straight from the bus; more join through the rail. The diagram is primary-only at both widths; at Wide only the gutter and columns widen (frame 11b, FS-W1, FS-W2) | The board's layout is fixed rows of fixed-height tiles, so a pure function places everything and tests the geometry; R-16: a plain laid-out diagram | Anchor preferences and wrapping rows (the board does not wrap); `GraphCanvas`; a `Canvas` host |
| D-13 | The board's answers (agreed 2026-11-01): two columns, origin and formed (FC-1); the focus drawn once, in the left column (FS-3); two parent sets stack, "Parents" then "Other parents" (FS-5, FC-3); the right column top-aligned with the left (FC-4; the board's open question on centring is answered as drawn); places truncate on narrow tiles (the board's open question on four-part places is accepted as drawn); empty copy as the brief (FS-11) | The user agreed the board | The brief's stacked units and wrapping rows |
| D-14 | No cap on tiles: every member Go returns is drawn | S13-04 bounds what Go returns (`Rules.MaxPivotEnd`, `kinship.MaxPairSteps`); 50 tiles lay out (§ 6 Scale) | A "show more" (not in the brief) |
| D-15 | The T-2 fixture gains `func (k *Canon) Date(entity []byte, propKey string, v datevalue.Value)` *(new)* in `graphcachetest.go`: `datevalues.InsertTx(k.db, v)`, then `k.observe(entity, propKey, "value_date_id", id)`. T-2 asserts order with its own ordered display, not `kinWorld.show` | No helper dates a birth today | Inserting rows by hand in the test |

**Rules this plan bends.** S13-D1 PPF-5 (archived brief: "Opening a Person from anywhere lands on Vitals", naming a family tile) → D-1 opens family tiles on Family → user, 2026-11-01 → the planning commit adds the working rule "A family tile opens the relative on Family" to the map. No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `conclusionsummaries.SortByBirth`, `birthKeyOf` *(new)*; `dateKey` / `dateKeyLess` *(renamed)* | func | `SortByTitle`, `sortEvents`, `eventKeyLess` | Extends them: one comparison, a generic sorter beside the title one |
| `GetPersonFamilyResponse.person` *(new)* | proto field | `FamilyMember.person` | The person is not a member of the partner units; the section needs the summary in every case |
| `Canon.Date` *(new)* | test helper | `Canon.Term` | Same shape, for dates |
| `PersonLifeDisplay.lifeLines`, `Missing` *(moved / new)* | func, enum | `PersonDetailContent.lifeLine` | Moved, with one option |
| `PersonTile`, `PersonTileContent`, `PersonTileRole`, `PersonTileButtonStyle` *(new)* | snowflake | `ConclusionDetailHeader`, the gallery tile, `PVOmnibarHitRow`, `SourcesGraphButtonStyle` | None draws one person as a card; the style copies the Sources one |
| `PersonFamilyContent` *(new)* | pure model | `PersonDetailContent` | A different read |
| `PersonFamilyLayout` *(new)* | pure layout | `PVFlowLayout`, `PVPageColumns` | A fixed two-column diagram with lines; neither places lines or knows units |
| `PersonFamilySection` *(new)* | snowflake | `ConclusionDetailGallerySection` | People and lines, not pictures |
| `ConclusionPageSection.queryKey` *(new)* | func | `ConclusionPageSection.label`, `glyph` | One more per-section property |
| `L10n.Conclusions.family*` *(new)* | keys | `L10n.Conclusions.pageSection*`, `a11yList` | New copy; labels compose with `a11yList` |

**Builds on.** `CatalogQueryKey.personFamily` and its descriptor (S13-04), `PlaceRegistry` Specs, `ConclusionDetailPresentation`, `PersonsList.location`, `PersonGenderTreatment`, `CachedThumbnail`, `PVCard`, `PVFlowLayout`, `PVHoverEffect`, `PVEmptyState`, `PVCallout`, `ConclusionDatePlaceLine`, `ConclusionListRow.title`.
**Deletes.** `PersonDetailContent.LifeLine` and its private `lifeLine` (moved). `eventKey` / `eventKeyLess` (renamed to `dateKey` / `dateKeyLess`). `.family` leaves the `pageSectionLater` empty case (the key stays for Timeline and Sources).
**Next adds.** S13-13 adds the Timeline's key to `ConclusionPageSection.queryKey`, places `PersonTile` and moves it to `DesignSystem/Recipes/PersonTile/` *(new)*; S13-16 adds `.sources → .personClaims`. Both use `ConclusionDetailPresentation` with their own content.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A family tile opens a Person on Family | `PersonDetailView` / `ConclusionDetailPage`, keyed by entity id | `viewIdentity: { $0.entityId }` resets the page; a relative merged since the read returns `conclusiondetails.not_found`, shown as the page's Callout; history decodes `pageSection` (S13-10 AC-2.4) |
| Back to a Family entry | same | D-8 (a) warms both keys; `.evict` while hidden, so Back reloads (spinner, then diagram) |
| The focus inside a unit's children | `PersonFamilyContent`, the handler | The focus has `link` unset like a co-parent but sits in children, where no co-parent sits; Swift tells it by id (D-3) and draws it once, merging the units (D-7) |
| `person` absent (an old engine, or a summary not read) | `PersonFamilyContent` | No focus tile is drawn; relatives still draw (T-6) |

### Scale

At the brief's sizes (§ 2: typical 10 – 20 tiles, large about 30, malformed 50): one family read per page view, warm with no graph query (S13-04 `TestPersonFamilyQueryCount`, which stays green with D-2 and D-3: the person's summary comes through the same `conclusionsummaries.ByIDs` call). `SortByBirth` computes one key per child. Swift draws one `PersonTile` per shown person; each `CachedThumbnail` loads through `ThumbnailCache`; `PersonFamilyLayout` is linear in the tiles and lines (T-13 at 50 siblings). The key refreshes on any touched handle only while visible (`.evict` when hidden). T-9 builds `PersonFamilyContent` for 50 members.

### Edges

- **Assumptions.** A relative can sit in more than one unit: at or under `Rules.MaxPivotEnd` (10) parents, a sibling sits under every parent set whose parents it shares (S13-04 Landed note (5)); past the bound, once. Go puts the person in every parent and sibling unit (D-3). Swift draws each sibling and the focus once (D-7's merge) and a parent once per set (a parent in two sets is drawn in each, frame 11c). `Tile.id` includes column and group, so ids stay unique (T-6). Past the bound every parent is its own set: up to n rails in the 20pt margin (D-12), which crowd but stay apart down to 1pt at 13 sets; past that rails overlap, accepted for malformed data. `dateKeyLess` is a strict weak order (keys, then ref); refs are unique per entity.
- **Neighbors.** A `.parents` unit with one partner → one parent over the children; a `.partner` unit with no children → the focus and the partner, no children row; a co-parent (`link == nil`) → drawn as any partner; `.oneParent` → the focus alone above the children (T-6).
- **Removed inputs.** A relative unpromoted or merged after the read: the next refresh drops them; a click before that lands on the page's not-found Callout.
- **Real data.** None — no table or seed drives the section.
- **Untrusted size.** Bounded in Go by S13-04 (parents at `Rules.MaxPivotEnd`; duplicate steps at `kinship.MaxPairSteps`); children per unit are not bounded and D-14 draws them all.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/conclusionsummaries` | `compose.go`: `eventKey` → `dateKey`, `eventKeyLess` → `dateKeyLess` (callers in the same file); `func birthKeyOf(h PersonSummary) dateKey` *(new)*. `order.go`: `func SortByBirth[T any](rows []T, summary func(T) PersonSummary)` *(new)*, doc: "orders people by birth date: dated first, earliest first, then by ref; the one owner of birth order" |
| `core/database/persondetail` | `PersonFamily`: adds `entityID` to the summaries' `want`; after `dropUnread`, `withFocus(family, summaries)` *(new, unexported)* appends the person (`kinship.Relative{ID: entityID}`) to the children of each `.parents` and `.siblings` unit when its summary was read; then each unit's `Children` sorted by `SortByBirth` through `summaries.Person`. Doc comment says both |
| `core/database/graphcache/graphcachetest` | `Canon.Date` (D-15) |

### FFI

`api/proto/engine.proto`: `GetPersonFamilyResponse` gains `PersonSummary person = 2;` and its comment: "children in each unit in birth order, undated last; the person among the children of each parents and siblings unit, with no link". `api/ffi/handlers/person_detail.go` fills `person` with `personSummaryProto` from the same summaries and picture paths. Regenerate with `./scripts/generate-proto.sh` (`api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`). Swift: `CatalogPersonFamily.person: CatalogPersonSummary?` *(new)*; `GoStore.getPersonFamily` maps it when `resp.hasPerson`.

### macOS

The board is turn 11 of the *Person Detail* project. Every frame below is the Person page pane with the switcher on Family; the header, switcher and other sections are S13-D1's.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Family section | snowflake | `macos/App/Features/Conclusions/PersonFamilySection.swift` *(new)* | D-10 states; `PersonFamilyContent` → `PersonFamilyLayout` → tiles, captions, lines (D-12) |
| Layout | pure | `macos/App/Features/Conclusions/PersonFamilyLayout.swift` *(new)* | D-12 |
| Person tile | snowflake | `macos/App/Features/Conclusions/PersonTile.swift` *(new)* | D-5. A relative: `Button` with `PersonTileButtonStyle`, id `persons.detail.family.tile.<Tile.id>`. The focus: no button, `.accessibilityElement(children: .ignore)` with its spoken label |
| Captions | text | `PersonFamilySection.swift` | `.pvMicroCaps()` (micro, semibold, caps tracking, uppercase), `PVColor.textMuted`, `.accessibilityHidden(true)` |
| Groups | — | `PersonFamilySection.swift` | Each parent set, the siblings and each right-column group is `.accessibilityElement(children: .contain)` with D-7's group label, its tiles declared in reading order so Tab runs the left column top to bottom, then the right column group by group (frame 11q) |
| Empty / error / loading | `PVEmptyState`, `PVCallout(tone: .danger, title:)`, `ProgressView` | — | D-10 |
| Refresh | the switcher row | `ConclusionDetailPage.swift` | D-8 (b), frame 11p |

**Frames in words.**

- **11a / 11b Full family, Regular and Wide.** AC-1.1. Two columns, both from the top. Left: "PARENTS", Heinrich and Anna full width; beneath, indented 32pt, Margarethe, Johann (focus), Friedrich, Wilhelm; a rail in the margin joins both parents and every child. Right: "PARTNERS AND CHILDREN"; Maria full width with Karl, Luise, Emma indented beneath on a rail; 28pt gap; Sophie with Otto; then Paul with no partner tile, joined straight from the bus. A horizontal from Johann's tile crosses the gutter to the bus, which links Maria, Sophie and Paul. 11b is the same at Wide: wider columns, a 96pt gutter.
- **11c Conflicting parents.** AC-1.2. "PARENTS" (Heinrich, Anna) then "OTHER PARENTS" (Heinrich, Katharina) stacked; the siblings once beneath, each set's rail reaching only its own children.
- **11d Recorded siblings only.** AC-1.3. "SIBLINGS", not indented, one rail; empty right column.
- **11e One parent.** "PARENT"; Dorothea alone on the rail over Heinrich, Martha (focus), Sophie.
- **11f / 11g Large family.** Two parent sets, ten siblings, three partners, eleven children with one undated: both columns grow down and the page scrolls; at Wide the same, wider.
- **11h After clicking Anna.** AC-1.7, AC-3.1. Toolbar crumbs *Persons › Anna Weber › Family* with the Back menu open on "Johann Weber · Family". Anna's focus tile alone on the left; Heinrich and their four children on the right.
- **11i Long name.** Two lines, then cut.
- **11j / 11k Dark.** 11a and 11c in dark.
- **11l Tile sheet.** AC-2.1 – AC-2.6: likeness (woman), person mark (man), non-binary with birth only, unknown with name only, long name, working label in italic, ref only in mono, death year with no place, hover (grown, shadow), keyboard focus (ring), the focus tile with and without a likeness.
- **11m No family.** The EmptyState: "No family yet", "Relatives appear here when people linked to this person in an Evidence graph are promoted."
- **11n First load.** A spinner centred where the columns go.
- **11o Read failed.** The danger Callout, titled "Family could not be read".
- **11p Refreshing.** Columns drawn; "Refreshing" at the end of the switcher row.
- **11q Keyboard.** Focus ring on Friedrich; the Tab order note.

Accessibility: one button per relative with D-7's label; groups as above; lines and captions hidden. Wide: primary-only, centred; only the gutter and the columns widen (FS-W1, FS-W2).

### Strings

L10n keys *(new)* under `L10n.Conclusions`, copy from the board (turn 11):

- `familyEmptyTitle` = "No family yet"; `familyEmptyMessage` = "Relatives appear here when people linked to this person in an Evidence graph are promoted."
- `familyErrorTitle` = "Family could not be read"
- `familyThisPerson` = "This person" (the Badge); `familyThisPersonSpoken` = "this person"
- `familyRoleParent` = "Parent", `familyRoleSibling` = "Sibling", `familyRolePartner` = "Partner", `familyRoleChild` = "Child"
- Captions: `familyCaptionParent` = "Parent", `familyCaptionParents` = "Parents", `familyCaptionOtherParent` = "Other parent", `familyCaptionOtherParents` = "Other parents", `familyCaptionSiblings` = "Siblings", `familyCaptionPartner` = "Partner", `familyCaptionPartners` = "Partners", `familyCaptionPartnerAndChildren` = "Partner and children", `familyCaptionPartnersAndChildren` = "Partners and children", `familyCaptionChildren` = "Children" (uppercased in the view, not in the copy)
- Group labels: `familyGroupThisPersonAndSiblings` = "This person and siblings", `familyGroupThisPerson` = "This person", `familyGroupPartner(_ name: String)` = "Partner %@", `familyGroupPartnerAndChildren(_ name: String)` = "Partner %@ and children"; parent sets and the children group reuse their captions
- `a11yLifeLineKnown(_ word: String, parts: String)` = "%1$@ %2$@" for D-6 `.omitted` ("died 1851"; "born 3 May 1790, Bremen" with `parts` from `a11yList`)

Labels compose with the existing `a11yList`. Removed: none. Each key is added in the step that first uses it (`add-localized-string`), with its entry in `macos/App/Resources/Localizable.xcstrings`.

## 7. Steps

**STEP-1 — Birth order and the focus (Go, proto)**
Files: `core/database/conclusionsummaries/compose.go`, `order.go`, `conclusionsummaries_test.go`; `core/database/persondetail/family.go`, `family_test.go`; `core/database/graphcache/graphcachetest/graphcachetest.go`; `api/proto/engine.proto`, `api/ffi/handlers/person_detail.go`, `person_detail_test.go`; the generated files; `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`. Skill: `add-catalog-query`, `add-ffi-handler`.
Do: D-2, D-3, D-15; § 6 Go core and FFI.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes (T-1, T-2, T-3; `TestPersonFamilyQueryCount` unchanged); `./scripts/generate-proto.sh` regenerates the two generated files; the app builds.

**STEP-2 — Shared line and title**
Files: `macos/App/Features/Conclusions/PersonLifeDisplay.swift`, `PersonDetailContent.swift`, `PersonDetailView.swift` (`PersonDetailLifeLines`), `ConclusionDatePlaceLine.swift`, `ConclusionListRow.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-5's `lineLimit`, D-6. No visible change on the header or lists.
Check: `PersonDetailTests` and `PersonsListTests` pass unchanged; T-4.

**STEP-3 — The person tile**
Files: `macos/App/Features/Conclusions/PersonTile.swift` *(new)*, L10n files. Skill: `add-ui-component`, `add-localized-string`.
Do: D-5; a `#Preview` with the FS-2 states.
Check: T-5 passes.

**STEP-4 — The section's content and layout**
Files: `macos/App/Features/Conclusions/PersonFamilyContent.swift` *(new)*, `macos/App/Features/Conclusions/PersonFamilyLayout.swift` *(new)*, L10n files.
Do: D-7, using D-1's location; D-12.
Check: T-6, T-7, T-8, T-9, T-12, T-13 pass.

**STEP-5 — One read mapping and the section key**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift` (`ConclusionDetailPresentation`, `ConclusionDetailLoaded`), `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`. Skill: `add-workspace-place`, `add-workspace-location`.
Do: D-8, D-9, D-10's generic init.
Check: existing detail page and list tests pass; T-10, T-11 pass.

**STEP-6 — The section view and wiring**
Files: `macos/App/Features/Conclusions/PersonFamilySection.swift` *(new)*, `PersonDetailView.swift`, L10n files.
Do: D-10 states, D-11, D-12's view, D-13, D-14; `.family` in `PersonDetailSections` draws `PersonFamilySection(entityId:)`.
Check: the app builds; `#Preview`s of the section on frames 11a, 11c, 11d and 11h's data render as the frames read in § 6; `scripts/check-pr.sh` passes.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `design/S13-D3-family-section.md` → `design/archive/`, `design/README.md`, `docs/macos-client-patterns.md` ("One sectioned detail page": a section that reads its own key declares it in `ConclusionPageSection.queryKey`, which the place warms and the switcher's refresh follows).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-12-family-section.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestSortByBirth` *(new)* | Table of `PersonSummary` rows: dated vs undated → dated first; 1815 vs 1820 → 1815 first; two undated → by ref; equal dates → by ref; input order kept for equal keys and refs differing only in case | D-2 |
| T-2 | Go | `core/database/persondetail/family_test.go` `TestPersonFamilyOrdersChildrenByBirth` *(new)* | `Canon.Birth` for each sibling and the person under one parent pair, dates by `Canon.Date(event, vocabkey.PropertyDate, …)`: Friedrich 1820, Wilhelm none, Margarethe 1815, the person 1817 → the parents unit's children by name: Margarethe, person, Friedrich, Wilhelm | AC-1.4, D-3 |
| T-3 | Go | `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily` (new case "carries the person") | The person with one parent and one partner → `person` is the person's summary; the person is a child of the parents unit with no link and absent from the partner unit's children | D-3 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeLinesOmitMissingHalves` *(new)* | Birth date and place, death date only: `.stated` → two lines, the d. place nil and spoken "place unknown"; `.omitted` → two lines, spoken "born 3 May 1790, Bremen" and "died 1851"; no dates or places with `.omitted` → no lines | D-6, AC-2.1 |
| T-5 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` *(new)* `tileContentFromSummary` | Anna (AC-2.1), Wilhelm (AC-2.2), a ref-only child (AC-2.3) → title source, lines, mark, colors | AC-2.1 – AC-2.3 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonFamilyContentTests.swift` *(new)* `columnsFromUnits` | Frame 11a's units → left: one parent set (Heinrich, Anna), siblings Margarethe, Johann (focus), Friedrich, Wilhelm; right: Maria (Karl, Luise, Emma), Sophie (Otto), no partner (Paul). Frame 11c's units (Friedrich and the person in both sets, Christian in the second) → siblings Johann, Friedrich, Christian, Margarethe, Wilhelm, each once, with their `sets`. Recorded siblings beside a parent set → after the set's siblings. 11d (siblings unit only) → `siblingsCaptioned`. 11h (Anna: partner only) → the focus alone on the left. A co-parent (`link` nil) → a partner. `person` nil → no tile is the focus. `Tile.id`s unique | AC-1.1 – AC-1.3, AC-1.7, AC-2.5, D-4, D-7 |
| T-7 | Swift | same file `tilesOpenTheRelativeOnFamily` | A member → `PersonsList.location(summary, orders:).withPageSection(.family)` | AC-3.1, D-1 |
| T-8 | Swift | same file `labelsAndCaptions` | Anna as parent → "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851"; the focus → "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person"; partner group → "Partner Maria Schulz and children"; captions for 11a ("Parents", "Partners and children"), 11c ("Other parents"), 11e ("Parent"), 11h ("Partner and children") | AC-2.4, AC-2.5, D-7 |
| T-9 | Swift | same file `fiftyMembersKeepOrder`; `emptyFamilyIsEmpty` | 50 children in one unit → 50 tiles in order; `units: []` → `isEmpty` | D-14, AC-1.5 |
| T-10 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnFamilyWarmsTheFamily` *(new)* | A Person location with `pageSection: .family` → `.conclusionDetail` and `.personFamily`; on Vitals → `.conclusionDetail` only | AC-4.2, D-8 |
| T-11 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` *(new)* | `ConclusionPageSection.family.queryKey(project:entityId: "e1")` → `.personFamily(project:entityId: "e1")`; every other section → nil | D-8 |
| T-12 | Swift | `macos/ProvenenciaTests/PersonFamilyLayoutTests.swift` *(new)* `regularAndWideMeasures` | 11a's content at width 1038, Regular → column 485 (`floor((1038 − 20 − 48) / 2)`), right column at x 553, Margarethe at x 52 width 453, Maria at y 26, the focus's horizontal ending at the bus x 521; at width 1180, Wide → gutter 96, column 532. 11c → rails at x 6 and 14, Christian joined only from 14. One set → rail at x 10; three sets → 4, 10, 16. 11d → siblings at x 20, rail at 10. Tiles listed left column then right column | D-12, FS-W1 |
| T-13 | Swift | same file `fiftySiblingsStack` | One set of two parents over 50 children, the focus among them → 52 tiles, height `26 + 2·104 − 12 + 20 + 50·104 − 12`, no tile overlap | D-14, Scale |

AC-1.6, AC-2.6, AC-3.2, AC-3.3 and AC-4.1 are covered by the manual check (§ 10). Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `PersonDetailTests` (STEP-2), `TestPersonFamily`, `TestGetPersonFamily` and `TestPersonFamilyQueryCount` (STEP-1), `PlaceRegistryTests` (STEP-5).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `PersonDetailContent.LifeLine`, `PersonDetailContent.lifeLine`, `eventKeyLess` and `eventKey` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (done in the planning commit)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] docs/macos-client-patterns.md: the section-key sentence (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Bar 3 and the slice's Check. Use a project with the § 5 sample family promoted (or the product owner's own test people with two marriages and a sibling known only through a shared parent).

1. (AC-1.1, AC-2.1, AC-2.2, AC-2.5) Open Johann, choose Family. It matches frame 11a as § 6 reads it: parents and siblings on the left, partners and children on the right, the line from Johann across the gutter; Anna's tile with her likeness, "b. 3 May 1790 · Bremen" and "d. 1851"; Wilhelm's with the mark and his name alone; Johann's raised with "This person" and not clickable.
2. (AC-1.4) Under the parents: Margarethe, Johann, Friedrich, Wilhelm.
3. (AC-1.2) Add a baptism naming Heinrich and Katharina as Friedrich's and Johann's parents, promote; Johann's Family matches frame 11c.
3b. (AC-1.7) Open Anna's Family: frame 11h.
3c. (AC-2.6) Hover a tile: it grows with a shadow; Tab: the focus ring.
4. (AC-3.1, AC-3.2, AC-3.3) Click Anna: her Family opens, breadcrumbs *Persons › Weber, Anna › Family*. Back: Johann's Family. Tab to Friedrich, press Return: his Family opens.
5. (AC-4.1) With Johann's Family showing, edit Anna (for example through her Evidence graph), return through Back: the tile shows the change; while it reloads, "Refreshing" shows in the switcher row.
6. (AC-1.5) Open a person with no promoted relatives: the EmptyState.
7. (AC-1.6) Open a Family entry from history for a person merged since: the danger Callout.
8. (Bar 8) Run Promote on a Source as before the spike: the proposals are unchanged.
9. Resize across 1600pt on Family: section, scroll and focus kept (FS-W3); at Wide the gutter and columns widen (frame 11b). Dark mode: frames 11j and 11k.
10. VoiceOver on Anna's tile reads "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851, button".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A tile's text overflows 92pt in Swift's fonts (a two-line name and two lines) | The tile clips to its frame and centres its text (D-5); the board's fonts fit, and the long-name frame 11i is in the manual check |
| `main` moves under the S13-04 files before this PR starts | Implementer step 2 re-checks § 4 against `main` @ `be51621` |
| Renaming `eventKey` touches another package | It is unexported; the rename stays in `compose.go` |

**Stop and report if:**

- `CatalogPersonFamily`, `CatalogFamilyUnit.Kind`, `.personFamily` or `persondetail.PersonFamily` differ from § 4.
- `PVPageColumns` does not give the primary column's width to `onGeometryChange` inside the section (D-12).
- The board needs data `GetPersonFamily` does not return (for example a marriage date to order partners): a follow-up engine PR, not this one.
- `TestPersonFamilyQueryCount` needs more queries after STEP-1.
- A Go test outside `conclusionsummaries`, `persondetail` and `api/ffi/handlers` changes because of STEP-1.

## 12. Open questions

None.

## 13. PR

**Title:** S13-12 — Family section
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-12-family-section.md`.
