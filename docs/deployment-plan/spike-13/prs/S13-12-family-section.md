# S13-12 — Family section

**Status:** Planned
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
**Review:** [`S13-12-family-section.review.md`](S13-12-family-section.review.md) — round 1 Fits with changes / Not ready, round 2 (after the board) Fits with changes / Not ready; every finding Fixed, none open

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
| The origin siblings (each once, the person among them, FC-3 order) and the person's own summary on the response (D-3) | STEP-1, AC-1.1 – AC-1.3, AC-1.7 |
| The b. and d. line: one builder with a stated or omitted missing half (D-6) | STEP-2, AC-2.1 |
| Two-line titles on `ConclusionListRow.title` (D-5) | STEP-2 |
| The person tile, and the hover lift shared with the artifact card (FS-1, FS-2, FS-3, D-16) | STEP-3, AC-2.1 – AC-2.6 |
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
| Thumb, card | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; `macos/App/DesignSystem/Components/Card/PVCard.swift` `PVCard` (`tone`, `border: .solid / .dashed`, `cornerRadius`, `elevated`, `padding`) | — |
| Width class | `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift`; `\.pvWidthClass` | Measured once by the workspace; pages read it. The primary column caps at 1180 at both widths |
| Card hover | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `ArtifactCardHover`, `GrownHitShape` (private) | The artifact card's grow-and-shadow hover, reduce motion, grown hit shape, `zIndex` |
| Focus ring | `macos/App/DesignSystem/Tokens/PVElevation.swift` `pvFocusRing(_:cornerRadius:)`; precedent `macos/App/Features/Sources/SourcesGraphButton.swift` `SourcesGraphButtonBody` (`@Environment(\.isFocused)`) | — |
| Test fixture | `core/database/graphcache/graphcachetest/graphcachetest.go` `Canon` (`Birth`, `Term`, private `observe`) | No helper sets a date on an entity |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift`, `PlaceRegistryTests.swift`, `WorkspaceSessionTests.swift` (builds `CatalogPersonFamily(units: [])`); `core/database/persondetail/family_test.go` `TestPersonFamily` (its `kinWorld.show` sorts names), `family_query_test.go` `TestPersonFamilyQueryCount`, `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personFamilyFollowsEveryTouchedHandle` | — |

**Precedent.** S13-10 ([#634](https://github.com/mendahu/provenencia/pull/634)) filled the Person page's sections inside `PersonDetailSections`; copy its shape. S13-11 ([#650](https://github.com/mendahu/provenencia/pull/650)) navigated from a section with `navigation.go(to:)` from the environment. `ConclusionDetailGallerySection` lays out a wrapping grid of `CachedThumbnail` tiles. `SourcesGraphButtonStyle` is the button style to copy for the tile.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A relative's tile opens `PersonsList.location(summary, orders:).withPageSection(.family)` with one `navigation.go(to:)` (user, 2026-11-01) | R-20; walking a family stays in Family (working rule "A family tile opens the relative on Family"); one place decides a Person link's history entry and title | Vitals (S13-D1 PPF-5), the user chose Family; building `.personDetail(` by hand |
| D-2 | Go orders each unit's **children** by birth, in the page read only. In `conclusionsummaries`, `eventKey` / `eventKeyLess` are renamed `dateKey` / `dateKeyLess` (same fields, same rule); `birthKeyOf(PersonSummary) dateKey` *(new)* reads `dateSort(h.Birth.Date)`; `SortByBirth[T any](rows []T, summary func(T) PersonSummary)` *(new)* in `order.go` beside `SortByTitle` computes each key once and sorts stable with `dateKeyLess`. `persondetail.PersonFamily` sorts each unit's `Children` with it (user, 2026-11-01) | Order is data, and the date sort key lives in Go; one date-order comparison in the package for events and births | Sorting in Swift (no date key there); a separate engine PR (the user chose this PR); sorting inside `kinship.FamilyUnits` (it holds ids); a `BornBefore` comparator (a second copy of `eventKeyLess`, keys recomputed per comparison) |
| D-3 | The left column's sibling list is Go's (M2-3). `persondetail.PersonFamily` also returns the **origin siblings** *(new)*: every child of the person's `.parents` units and `.siblings` unit, each once, with the person among them, in the board's order (FC-3, one rule for any number of parent sets): (1) the siblings in every parent set; (2) walking the parent sets from last to first, each set's siblings not yet placed; (3) the siblings in no parent set (recorded only); each group in birth order by `SortByBirth`. Each carries `parent_sets`, the indexes into the `.parents` units of the sets it is a child of (empty for recorded only). With no parent sets, (3) alone, the person in birth order. With no parents and no siblings, the person alone. The person is left out only when its summary was not read. Units keep S13-04's members (only their children's order changes, D-2). Proto: `GetPersonFamilyResponse` gains `PersonSummary person = 2;` (the person's own summary) and `repeated OriginSibling origin_siblings = 3;`, `message OriginSibling { FamilyMember member = 1; repeated int32 parent_sets = 2; }` (the person's member has no link). Swift: `CatalogPersonFamily.person: CatalogPersonSummary? = nil`, `originSiblings: [CatalogOriginSibling] = []` *(new)*, `struct CatalogOriginSibling: Sendable, Equatable { var member: CatalogFamilyMember; var parentSets: [Int] }`. Swift finds the focus by `entity.id == family.person?.entity.id`. `persondetail.Families` is unchanged | Working rule "Family groups in Go": who sits where, and in what order, is Go's; the future tree reads the same answer. The person's place among siblings is birth order, which only Go can compute | Merging and ordering in Swift (a rule only Swift would hold); putting the person into each unit's children (every unit test string changes, and Swift would still have to merge) |
| D-4 | Recorded and implied relatives draw the same (user, 2026-11-01). The section reads no field of `CatalogFamilyLink` | The user's call | A dashed or `lineInferred` line, a Badge |
| D-5 | The tile is a snowflake in `macos/App/Features/Conclusions/PersonTile.swift` *(new)*: `PersonTile` (view) and `PersonTileContent` (pure, `Equatable`), with no Family vocabulary (M2-2). `PersonTileContent.init(summary: CatalogPersonSummary, role: String?, isFocus: Bool, locale: Locale, orders: NameFormatOrders)` holds `id` (entity id), `title: ConclusionTitleSource` (`PersonTitleDisplay.titleSource(summary, style: .natural, orders:)`), `lines: [PersonLifeDisplay.LifeLine]` (D-6 `.omitted`), `imageRelPath` (`summary.pictureImageRelPath`), `colors: PVThumbnail.Colors?` (`PersonGenderTreatment.read(summary)?.thumbnailColors`), `role`, `isFocus` and `spokenLabel`: a relative `a11yList(role, rest: a11yList(title.spokenText, rest: life))` (without the role when nil); the focus `a11yList(a11yList(title.spokenText, rest: life), rest: familyThisPersonSpoken)`; `life` is the lines' spoken texts joined with `a11yList`. The view (FC-5, frame 11l) is `.frame(height: 92)`, `.clipped()`: `PVCard(cornerRadius: PVRadius.sm, padding: 10)` holding an `HStack(spacing: 10)` of `CachedThumbnail(projectDir:relPath: imageRelPath, size: 70, cornerRadius: PVRadius.sm, fallbackMark: .subjectPerson, fallbackLabel: L10n.Conclusions.personNoLikeness, colors:)` and a `VStack(alignment: .leading, spacing: 3)` centred beside it: the title through `ConclusionListRow.title`, which gains `lineLimit: Int = 1` and `emphasized: Bool = false` (`.fontWeight(.semibold)`; the tile passes 2 and `isFocus`), then a `ConclusionDatePlaceLine(missing: .omitted)` per line. Fonts are the kit's (the list title font; the line's sizes). The focus is `PVCard(tone: .raised, cornerRadius: PVRadius.sm, elevated: true, padding: 10)` with a neutral `PVBadge(L10n.Conclusions.familyThisPerson)` after the title; the board's inset fill is not drawn (no `PVCard` tone has it). A relative is a `Button` with a private `PersonTileButtonStyle` (the focus ring, `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)` from `@Environment(\.isFocused)`, and no other chrome) and the hover lift `.pvLiftHover(scale: 1.025)` (D-16). `.help(title.text)` on every tile. S13-13 moves the tile to `DesignSystem/Recipes/PersonTile/` *(new)* and passes its event role's label | Working rule "One person tile, built once": the Timeline passes a role term's label, and the tile composes its own spoken label; one title styling | A Family role enum in the tile (S13-13 would widen it or re-write the label); a recipe now (one caller); a copy of `titleText`; a new `PVCard` tone for the inset fill |
| D-6 | `PersonDetailContent.LifeLine` and `lifeLine` move to `PersonLifeDisplay` as `PersonLifeDisplay.LifeLine` and `PersonLifeDisplay.lifeLines(_ summary: CatalogPersonSummary?, locale:missing:) -> [LifeLine]` *(moved)*, with `enum Missing: Equatable { case stated, omitted }`. `.stated` is today's header (both lines always, a missing half "unknown", spoken "Born 3 May 1790, Bremen"). `.omitted` returns only lines with a date or a place, leaves a missing half nil, and speaks mid-sentence in lowercase: `a11yLifeLine(L10n.Conclusions.a11yBornInline, date:place:)` when both halves are known ("born 3 May 1790, Bremen"), else `a11yLifeLineKnown(word, parts:)` with the known half ("died 1851"). `ConclusionDatePlaceLine` gains `var missing: PersonLifeDisplay.Missing = .stated`; with `.omitted` a nil half and its "·" are not drawn, the line is one line, and the place truncates at its tail (frames 11a, 11l). The header passes `.stated`, the tile `.omitted` | One builder and one view for the b. / d. line in the header and both tiles; the board leaves the missing half out on tiles and states it in the header (frame 11h); the tile's words sit mid-sentence | A second builder in the tile; "unknown" on every tile; capitalised words mid-sentence |
| D-7 | `PersonFamilyContent` *(new, pure, `Equatable`)* in `macos/App/Features/Conclusions/PersonFamilyContent.swift` *(new)*: `init(family: CatalogPersonFamily, locale: Locale, orders: NameFormatOrders)`, two columns (FC-1) laid out from what Go gives, with no regrouping. `PersonFamilyRole` *(new)* (`parent`, `sibling`, `partner`, `child`) lives here and maps to its `familyRole*` label. **Left:** `parentSets: [[Tile]]`, the partners of each `.parents` unit in order (role parent); `siblings: [Sibling]` from `family.originSiblings` in order, each a `Tile` (role sibling; the focus without a role) and `sets: [Int]` (its `parentSets`); `siblingsCaption: Bool`, true when there are no parent sets and at least one sibling besides the focus. **Right:** `groups: [Group]`, one per `.partner` unit then the `.oneParent` unit, each `partner: Tile?` (role partner; nil for `.oneParent`) and `children: [Tile]` (role child). The children of `.parents` and `.siblings` units are not read (the origin siblings hold them). `isEmpty`: no parent sets, no sibling besides the focus, and no groups (frame 11m). `Tile` has `id` ("\(column)-\(group)-\(entityId)"), `content: PersonTileContent`, `location: WorkspaceLocation?` (D-1; nil for the focus). Captions (hidden from VoiceOver): first parent set "Parents" (two parents, or a later set exists) or "Parent"; later sets "Other parents" / "Other parent"; "Siblings" when `siblingsCaption`; the right column one caption on its first group: "Partners and children", "Partner and children", "Partners" or "Partner" by the partner count and whether any group has children, else "Children". Group labels (VoiceOver): each parent set its caption; the siblings "This person and siblings", or "This person" (`familyThisPerson`) alone; each partner group "Partner Maria Schulz and children" or "Partner Maria Schulz"; the `.oneParent` group "Children" | Go groups and orders; Swift places (working rule "Family groups in Go") | Merging units in Swift; reading `link` for roles |
| D-8 | A page section's query key is declared once: `ConclusionPageSection.queryKey(project: ProjectKey, entityId: String) -> CatalogQueryKey?` *(new)* in `macos/App/Features/Workspace/WorkspaceLocation.swift` (`.family → .personFamily`, every other section nil). Three readers: (a) `PlaceRegistry` Spec `.personDetail` appends `location.pageSection.flatMap { $0.queryKey(project:entityId:) }`; (b) `ConclusionDetailLoaded` passes `refreshing || sectionRefreshing` to the header band, where `sectionRefreshing` is `true` when the open section's key has a handle that holds a value and `isFetching`; (c) the section reads the same key | One source of truth for "is the open section refreshing"; S13-13 and S13-16 each add one case (working rule "Each section reads its own query"); a first load shows the section's spinner only, not the switcher's | A preference tunnel with a mirrored `@State` (no state preference exists in the app); a per-section branch in `PlaceRegistry` |
| D-9 | `ConclusionDetailLoaded` gains the session and the entity id it needs for (b): `ConclusionDetailPage` already holds `session` and `entityId` and passes them in | (b) needs the key and a handle | Putting the session in the sections context |
| D-10 | `ConclusionDetailPresentation.init` becomes generic: `init<Value>(_ read: QueryRead<Value>, isFetching: Bool, content: (Value) -> Content)`; existing callers compile unchanged. The Family section uses `ConclusionDetailPresentation<PersonFamilyContent>`: `.loading` → `ProgressView()` (regular size) centred, 96pt from the top (frame 11n); `.failed(message)` → `PVCallout(tone: .danger, title: L10n.Conclusions.familyErrorTitle, message:)`, id `persons.detail.family.error` (frame 11o); `.content` with `isEmpty` → `PVEmptyState` (glyph `ConclusionPageSection.family.glyph`), id `persons.detail.family.empty` (frame 11m); else the two columns, id `persons.detail.family` | One mapping from read state to view state for every section | A third hand-written switch |
| D-11 | `PersonFamilySection` reads `@Environment(WorkspaceSession.self)` and `@Environment(WorkspaceNavigation.self)`. `PersonDetailSections` gains `entityId: String`, passed from `PersonDetailView`. The section gets its handle with `session.queryHandle(key)` and starts it in `.task { let _: QueryHandle<CatalogPersonFamily> = session.query(key) }`, as `ConclusionDetailPage` does | The session is already in the environment; the key needs the id | Threading the session through `PersonDetailSections` |
| D-12 | The layout is a pure function, `PersonFamilyLayout` *(new)* in `macos/App/Features/Conclusions/PersonFamilyLayout.swift`: `init(content: PersonFamilyContent, width: CGFloat, widthClass: PVWidthClass)` gives `groups: [Group]` in reading order (left column top to bottom, then right column group by group), each with `label`, `frame: CGRect` and `tiles: [(id, frame: CGRect)]` (frames relative to the group); `captions: [(text, origin: CGPoint)]`; `segments: [CGRect]` (1pt horizontal or vertical lines); and `height`. The section reads `@Environment(\.pvWidthClass)` and its width with `onGeometryChange` on the primary column. The view is a `ZStack(alignment: .topLeading)` of that height: the segments as one `Path` filled with `PVColor.borderStrong`, `.accessibilityHidden(true)`; the captions (hidden); then one container per group, at its frame, `.accessibilityElement(children: .contain)` with its label, holding its tiles at their frames, declared in reading order. **Measures** (FC-2 – FC-5, the board's `layTree3`): rail margin 20; gutter 48 at `.regular`, 96 at `.wide`; column width `TP = floor((width − 20 − gutter) / 2)`; left column at x 20, right at `c2 = 20 + TP + gutter`; tile height 92; 12 between tiles; caption band 26 (a block starts with its caption at its top, tiles 26 below); indent 32 (indented tiles are `TP − 32` wide); 20 after each parent set; 28 between right-column groups. **Left:** each parent set at x 20 under its caption; then the siblings at x 52 (at x 20, from y 26, when there are no parent sets; the lone focus at x 20, y 26). Rails per parent set: x 10 for one set; 6 and 14 for two; for n ≥ 3, `4 + round(12 · k / (n − 1))`. Set k's rail joins each of its parents (a horizontal from the rail to x 20 at the parent's centre) and runs from its first parent's centre down to the centre of the last sibling whose `sets` holds k; it draws a horizontal from the rail to x 52 for each such sibling not in an earlier set (an earlier set's horizontal crosses this rail at the shared sibling's centre). Siblings with empty `sets` (recorded only) beside parent sets hang off a bracket at x 40, from the focus's centre down to the last of them, with a horizontal from x 40 to x 52 for the focus and each of them; set horizontals crossing x 40 are drawn as they fall. With no parent sets, one rail at x 10 joins every sibling with a horizontal to x 20. **Right:** groups from y 26 (caption band above the first), top-aligned with the left (FC-4). A group's join y is its partner's centre, or its first child's centre when it has no partner. A partner sits at `c2`; children at `c2 + 32`, beneath the partner, or from the group's top when there is no partner. A bus at x `c2 − 32` runs from the minimum to the maximum of the focus's centre and every group's join y; a horizontal joins the focus tile's right edge to the bus at the focus's centre. A partner joins the bus with a horizontal at its centre; a rail at `c2 − 12` runs from the partner's centre to its last child's centre, with a horizontal from the rail to `c2 + 32` at each child's centre. A group with no partner and one child: a horizontal from the bus to `c2 + 32` at the child's centre; with more: a horizontal from the bus to the rail at the first child's centre, the rail down to the last child's centre, and a horizontal to each child. Primary-only at both widths; at Wide only the gutter and columns widen (frame 11b, FS-W1, FS-W2) | A pure function places everything and its geometry is tested; the width class has one owner (`PVWidthClass`); R-16 | Deriving Wide from the measured width (the primary column caps at 1180 at both widths); a flat tile list (no VoiceOver groups); anchor preferences; `GraphCanvas` |
| D-13 | The board's answers (agreed 2026-11-01): two columns, origin and formed (FC-1); the focus drawn once, in the left column (FS-3); two parent sets stack, "Parents" then "Other parents" (FS-5, FC-3); the right column top-aligned with the left (FC-4; the board's open question on centring is answered as drawn); places truncate on narrow tiles (the board's open question on four-part places is accepted as drawn); empty copy as the brief (FS-11) | The user agreed the board | The brief's stacked units and wrapping rows |
| D-14 | No cap on tiles: every member Go returns is drawn | S13-04 bounds what Go returns (`Rules.MaxPivotEnd`, `kinship.MaxPairSteps`); 50 tiles lay out (§ 6 Scale) | A "show more" (not in the brief) |
| D-15 | The T-2 fixture gains `func (k *Canon) Date(entity []byte, propKey string, v datevalue.Value)` *(new)* in `graphcachetest.go`: `datevalues.InsertTx(k.db, v)`, then `k.observe(entity, propKey, "value_date_id", id)`. T-2 asserts order with its own ordered display, not `kinWorld.show` | No helper dates a birth today | Inserting rows by hand in the test |
| D-16 | `ArtifactCardHover` and `GrownHitShape` (private in `macos/App/Features/Sources/SourcePageArtifactsView.swift`) move to `macos/App/DesignSystem/Components/HoverEffect/PVLiftHover.swift` *(new)* as `.pvLiftHover(scale: CGFloat, resting: [PVShadowLayer]? = nil)`: the grow to `scale` with `PVElevation.md` on hover, reduce motion honoured, `PVMotion.durationFast`, the grown hit shape, and `zIndex` 1 while hovered; `resting` is the shadow it keeps at rest (the artifact card passes `PVElevation.sm`, the tile nil). The artifact card calls `.pvLiftHover(scale: 1.02, resting: PVElevation.sm)` and looks as before; the tile calls `.pvLiftHover(scale: 1.025)`. A README row in `macos/App/DesignSystem/README.md` | One owner for the lift (M2-1): the flicker fix, reduce motion and z-order apply to both cards; content-agnostic, two callers (`add-ui-component`) | A second grow-and-shadow in the tile's button style; scaling on press on top of `PVHoverEffect`'s press scale |

**Rules this plan bends.** S13-D1 PPF-5 (archived brief: "Opening a Person from anywhere lands on Vitals", naming a family tile) → D-1 opens family tiles on Family → user, 2026-11-01 → the planning commit adds the working rule "A family tile opens the relative on Family" to the map. No other rule is bent.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `conclusionsummaries.SortByBirth`, `birthKeyOf` *(new)*; `dateKey` / `dateKeyLess` *(renamed)* | func | `SortByTitle`, `sortEvents`, `eventKeyLess` | Extends them: one comparison, a generic sorter beside the title one |
| `GetPersonFamilyResponse.person`, `origin_siblings`, `OriginSibling` *(new)*; `persondetail.originSiblings` *(new, unexported)* | proto, func | `FamilyUnit`, `dropUnread` | The person's summary and the merged, ordered sibling list are Go's (D-3); units stay the tree's shape |
| `Canon.Date` *(new)* | test helper | `Canon.Term` | Same shape, for dates |
| `PersonLifeDisplay.lifeLines`, `Missing` *(moved / new)* | func, enum | `PersonDetailContent.lifeLine` | Moved, with one option |
| `PersonTile`, `PersonTileContent`, `PersonTileButtonStyle` *(new)* | snowflake | `ConclusionDetailHeader`, the gallery tile, `PVOmnibarHitRow`, `SourcesGraphButtonBody` | None draws one person as a card; the style is the focus ring only |
| `pvLiftHover` *(new, moved)* | component | `ArtifactCardHover` | Moved and given a scale and resting shadow (D-16) |
| `PersonFamilyRole` *(new)* | enum | — | Family's roles, outside the tile |
| `PersonFamilyContent` *(new)* | pure model | `PersonDetailContent` | A different read |
| `PersonFamilyLayout` *(new)* | pure layout | `PVFlowLayout`, `PVPageColumns` | A fixed two-column diagram with lines; neither places lines or knows units |
| `PersonFamilySection` *(new)* | snowflake | `ConclusionDetailGallerySection` | People and lines, not pictures |
| `ConclusionPageSection.queryKey` *(new)* | func | `ConclusionPageSection.label`, `glyph` | One more per-section property |
| `L10n.Conclusions.family*` *(new)* | keys | `L10n.Conclusions.pageSection*`, `a11yList` | New copy; labels compose with `a11yList` |

**Builds on.** `CatalogQueryKey.personFamily` and its descriptor (S13-04), `PlaceRegistry` Specs, `ConclusionDetailPresentation`, `PersonsList.location`, `PersonGenderTreatment`, `CachedThumbnail`, `PVCard`, `PVWidthClass`, `pvFocusRing`, `pvMicroCaps`, `PVEmptyState`, `PVCallout`, `ConclusionDatePlaceLine`, `ConclusionListRow.title`, `a11yList`, `a11yLifeLine`.
**Deletes.** `PersonDetailContent.LifeLine` and its private `lifeLine` (moved). `ArtifactCardHover` and `GrownHitShape` (moved to `pvLiftHover`). `eventKey` / `eventKeyLess` (renamed to `dateKey` / `dateKeyLess`). `.family` leaves the `pageSectionLater` empty case (the key stays for Timeline and Sources).
**Next adds.** S13-13 adds the Timeline's key to `ConclusionPageSection.queryKey`, places `PersonTile` with its role term's label and moves it to `DesignSystem/Recipes/PersonTile/` *(new)*; S13-16 adds `.sources → .personClaims`. Both use `ConclusionDetailPresentation` with their own content.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A family tile opens a Person on Family | `PersonDetailView` / `ConclusionDetailPage`, keyed by entity id | `viewIdentity: { $0.entityId }` resets the page; a relative merged since the read returns `conclusiondetails.not_found`, shown as the page's Callout; history decodes `pageSection` (S13-10 AC-2.4) |
| Back to a Family entry | same | D-8 (a) warms both keys; `.evict` while hidden, so Back reloads (spinner, then diagram) |
| The focus among the origin siblings | `PersonFamilyContent`, the handler | The focus's member has no link; Swift tells it by id (D-3) |
| `person` absent (an old engine, or a summary not read) | `PersonFamilyContent` | No focus tile is drawn; relatives still draw (T-6) |

### Scale

At the brief's sizes (§ 2: typical 10 – 20 tiles, large about 30, malformed 50): one family read per page view, warm with no graph query (S13-04 `TestPersonFamilyQueryCount`, which stays green with D-2 and D-3: the person's summary comes through the same `conclusionsummaries.ByIDs` call). `SortByBirth` computes one key per child. Swift draws one `PersonTile` per shown person; each `CachedThumbnail` loads through `ThumbnailCache`; `PersonFamilyLayout` is linear in the tiles and lines (T-13 at 50 siblings). The key refreshes on any touched handle only while visible (`.evict` when hidden). T-9 builds `PersonFamilyContent` for 50 members.

### Edges

- **Assumptions.** A relative can sit in more than one unit: at or under `Rules.MaxPivotEnd` (10) parents, a sibling sits under every parent set whose parents it shares (S13-04 Landed note (5)); past the bound, once. `originSiblings` lists each sibling once with all its sets, so Swift draws each sibling once and a parent once per set (a parent in two sets is drawn in each, frame 11c). `Tile.id` includes column and group, so ids stay unique (T-6). Past the bound every parent is its own set: up to n rails in the 20pt margin (D-12), apart down to 1pt at 13 sets; past that they overlap, accepted for malformed data. Existing `TestPersonFamily` expectations hold: units keep their members, and `kinWorld.show` sorts names, so the new child order does not show there.
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
| `core/database/persondetail` | `PersonFamily` returns `(Family, []OriginSibling, Summaries, error)`: adds `entityID` to the summaries' `want`; after `dropUnread`, sorts each unit's `Children` by `SortByBirth` through `summaries.Person`; then `originSiblings(family, entityID, summaries) []OriginSibling` *(new, unexported)*, `type OriginSibling struct { Relative kinship.Relative; ParentSets []int }` *(new)*, per D-3. Doc comment says both |
| `core/database/graphcache/graphcachetest` | `Canon.Date` (D-15) |

### FFI

`api/proto/engine.proto`: `GetPersonFamilyResponse` gains `PersonSummary person = 2;` and `repeated OriginSibling origin_siblings = 3;`, with `message OriginSibling { FamilyMember member = 1; repeated int32 parent_sets = 2; }`; its comment adds "children in each unit in birth order, undated last; origin_siblings is every sibling once with the person among them, in the order the page draws them". `api/ffi/handlers/person_detail.go` fills both with `personSummaryProto` and the family member mapper, from the same summaries and picture paths. Regenerate with `./scripts/generate-proto.sh` (`api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`). Swift: `CatalogPersonFamily.person: CatalogPersonSummary? = nil` and `originSiblings: [CatalogOriginSibling] = []` *(new)*, so `FakeStore` and `CatalogPersonFamily(units: [])` in `WorkspaceSessionTests` compile unchanged; `GoStore.getPersonFamily` maps `person` when `resp.hasPerson` and each origin sibling with `mapFamilyMember`.

### macOS

The board is turn 11 of the *Person Detail* project. Every frame below is the Person page pane with the switcher on Family; the header, switcher and other sections are S13-D1's.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Family section | snowflake | `macos/App/Features/Conclusions/PersonFamilySection.swift` *(new)* | D-10 states (a nil handle draws the `.loading` state); `PersonFamilyContent(family:locale: .autoupdatingCurrent, orders: session.nameFormat)` → `PersonFamilyLayout` → groups, captions, lines (D-12); `projectDir` from `session.projectKey.projectDir`. `#Preview`s build `CatalogPersonFamily` values for frames 11a, 11c, 11d and 11h in the same file |
| Layout | pure | `macos/App/Features/Conclusions/PersonFamilyLayout.swift` *(new)* | D-12 |
| Person tile | snowflake | `macos/App/Features/Conclusions/PersonTile.swift` *(new)* | D-5. A relative: `Button` with `PersonTileButtonStyle` and `.pvLiftHover(scale: 1.025)`, `.accessibilityLabel(spokenLabel)`, id `persons.detail.family.tile.<Tile.id>`. The focus: no button, `.accessibilityElement(children: .ignore)` with its spoken label |
| Hover lift | component | `macos/App/DesignSystem/Components/HoverEffect/PVLiftHover.swift` *(new)* | D-16 |
| Captions | text | `PersonFamilySection.swift` | `.pvMicroCaps()` (micro, semibold, caps tracking, uppercase), `PVColor.textMuted`, `.accessibilityHidden(true)` |
| Groups | — | `PersonFamilySection.swift` | One container per `PersonFamilyLayout.Group` (D-12), `.accessibilityElement(children: .contain)` with D-7's group label, declared in reading order so Tab runs the left column top to bottom, then the right column group by group (frame 11q) |
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
- `familyThisPerson` = "This person" (the Badge, and the siblings group label when the focus stands alone); `familyThisPersonSpoken` = "this person"
- `familyRoleParent` = "Parent", `familyRoleSibling` = "Sibling", `familyRolePartner` = "Partner", `familyRoleChild` = "Child"
- Captions: `familyCaptionParent` = "Parent", `familyCaptionParents` = "Parents", `familyCaptionOtherParent` = "Other parent", `familyCaptionOtherParents` = "Other parents", `familyCaptionSiblings` = "Siblings", `familyCaptionPartner` = "Partner", `familyCaptionPartners` = "Partners", `familyCaptionPartnerAndChildren` = "Partner and children", `familyCaptionPartnersAndChildren` = "Partners and children", `familyCaptionChildren` = "Children" (uppercased in the view, not in the copy)
- Group labels: `familyGroupThisPersonAndSiblings` = "This person and siblings", `familyGroupPartner(_ name: String)` = "Partner %@", `familyGroupPartnerAndChildren(_ name: String)` = "Partner %@ and children"; parent sets and the children group reuse their captions
- `a11yBornInline` = "born", `a11yDiedInline` = "died" (mid-sentence; the header keeps `a11yBorn` / `a11yDied`); `a11yLifeLineKnown(_ word: String, part: String)` = "%1$@ %2$@" for a line with one half known ("died 1851"); a line with both halves uses the existing `a11yLifeLine` ("born 3 May 1790, Bremen")

Labels compose with the existing `a11yList`. Removed: none. Each key is added in the step that first uses it (`add-localized-string`), with its entry in `macos/App/Resources/Localizable.xcstrings`.

## 7. Steps

**STEP-1 — Birth order and the focus (Go, proto)**
Files: `core/database/conclusionsummaries/compose.go`, `order.go`, `conclusionsummaries_test.go`; `core/database/persondetail/family.go`, `family_test.go`; `core/database/graphcache/graphcachetest/graphcachetest.go`; `api/proto/engine.proto`, `api/ffi/handlers/person_detail.go`, `person_detail_test.go`; the generated files; `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`. Skill: `add-catalog-query`, `add-ffi-handler`.
Do: D-2, D-3, D-15; § 6 Go core and FFI.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes (T-1, T-2, T-3; `TestPersonFamily`, `TestGetPersonFamily` and `TestPersonFamilyQueryCount` unchanged but for `PersonFamily`'s new return value at their call sites); `./scripts/generate-proto.sh` regenerates the two generated files; the app builds.

**STEP-2 — Shared line and title**
Files: `macos/App/Features/Conclusions/PersonLifeDisplay.swift`, `PersonDetailContent.swift`, `PersonDetailView.swift` (`PersonDetailLifeLines`), `ConclusionDatePlaceLine.swift`, `ConclusionListRow.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-swift-test`, `add-localized-string`.
Do: D-5's `lineLimit`, D-6. No visible change on the header or lists.
Check: `PersonDetailTests` and `PersonsListTests` pass unchanged; T-4.

**STEP-3 — The hover lift and the person tile**
Files: `macos/App/DesignSystem/Components/HoverEffect/PVLiftHover.swift` *(new)*, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/DesignSystem/README.md`, `macos/App/Features/Conclusions/PersonTile.swift` *(new)*, L10n files. Skill: `add-ui-component`, `add-localized-string`.
Do: D-16 first (the artifact card looks and behaves as before), then D-5; a `#Preview` with frame 11l's states.
Check: the Source page's artifact cards hover as before; T-5 passes; `scripts/check-guidance-names.py` passes with the README row.

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
| T-2 | Go | `core/database/persondetail/family_test.go` `TestPersonFamilyOriginSiblings` *(new)*, table-driven | Each case builds people with `Canon.Birth` and dates them with `Canon.Date(event, vocabkey.PropertyDate, …)`, then reads `PersonFamily`'s origin siblings as (name, sets). (a) One pair; Friedrich 1820, Wilhelm undated, Margarethe 1815, the person 1817 → Margarethe [0], person [0], Friedrich [0], Wilhelm [0]; the units' children in the same birth order. (b) Two sets; the person and Friedrich in both, Christian in the second, Margarethe and Wilhelm in the first → person [0 1], Friedrich [0 1], Christian [1], Margarethe [0], Wilhelm [0]. (c) A pair plus a recorded sibling with no parent → the set's siblings, then the recorded one with []. (d) Recorded siblings only (Ida 1838, the person 1841, Bernhard 1844) → Ida, person, Bernhard, all []. (e) A spouse only → the person alone | AC-1.1 – AC-1.4, AC-1.7, D-2, D-3 |
| T-3 | Go | `api/ffi/handlers/person_detail_test.go` `TestGetPersonFamily` (new case "carries the person and the origin siblings") | The person with one parent pair, one sibling and one partner → `person` is the person's summary; `origin_siblings` lists the sibling and the person, the person's member with no link, each with `parent_sets` [0]; the partner unit unchanged | D-3 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `lifeLinesOmitMissingHalves` *(new)* | Birth date and place, death date only: `.stated` → two lines, the d. place nil and spoken "place unknown"; `.omitted` → two lines, spoken "born 3 May 1790, Bremen" and "died 1851"; no dates or places with `.omitted` → no lines | D-6, AC-2.1 |
| T-5 | Swift | `macos/ProvenenciaTests/PersonTileTests.swift` *(new)* `tileContentFromSummary`, `spokenLabels` | Anna (AC-2.1), Wilhelm (AC-2.2), a ref-only child (AC-2.3) → title source, lines, mark, colors. Spoken: Anna with role "Parent" → "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851"; with role nil → no role; the focus → "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person" | AC-2.1 – AC-2.3, AC-2.4, AC-2.5, D-5 |
| T-6 | Swift | `macos/ProvenenciaTests/PersonFamilyContentTests.swift` *(new)* `columnsFromFamily` | Frame 11a's family (units and origin siblings as Go returns them) → left: one parent set (Heinrich, Anna), siblings Margarethe, Johann (focus), Friedrich, Wilhelm with sets [0]; right: Maria (Karl, Luise, Emma), Sophie (Otto), no partner (Paul). 11c → two parent sets; siblings in Go's order, each once, with their sets. 11d → `siblingsCaption`. 11h (Anna: partner only) → the focus alone on the left, no caption. A co-parent (`link` nil) → a partner. `person` nil → no tile is the focus. The children of `.parents` and `.siblings` units are not shown twice. `Tile.id`s unique | AC-1.1 – AC-1.3, AC-1.7, AC-2.5, D-4, D-7 |
| T-7 | Swift | same file `tilesOpenTheRelativeOnFamily` | A member → `PersonsList.location(summary, orders:).withPageSection(.family)` | AC-3.1, D-1 |
| T-8 | Swift | same file `captionsAndGroupLabels` | Captions for 11a ("Parents", "Partners and children"), 11c ("Parents", "Other parents"), 11e ("Parent"), 11h (none left, "Partner and children"), a partner with no children ("Partner"), children only ("Children"); group labels "This person and siblings", "This person", "Partner Maria Schulz and children", "Partner Sophie Brandt" | D-7 |
| T-9 | Swift | same file `fiftyMembersKeepOrder`; `emptyFamilyIsEmpty` | 50 children in one unit → 50 tiles in order; `units: []` → `isEmpty` | D-14, AC-1.5 |
| T-10 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnFamilyWarmsTheFamily` *(new)* | A Person location with `pageSection: .family` → `.conclusionDetail` and `.personFamily`; on Vitals → `.conclusionDetail` only | AC-4.2, D-8 |
| T-11 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` *(new)* | `ConclusionPageSection.family.queryKey(project:entityId: "e1")` → `.personFamily(project:entityId: "e1")`; every other section → nil | D-8 |
| T-12 | Swift | `macos/ProvenenciaTests/PersonFamilyLayoutTests.swift` *(new)* `regularAndWideMeasures` | 11a's content at width 1038, `.regular` → column 485 (`floor((1038 − 20 − 48) / 2)`), right column at x 553, Margarethe at x 52 width 453, Maria at y 26, the focus's horizontal ending at the bus x 521; at width 1180, `.regular` → gutter 48, column 556, and `.wide` → gutter 96, column 532. The bus spans from Maria's centre (y 72) to Paul's; Paul at x `c2 + 32`. 11c → rails at x 6 and 14; the second rail reaches Christian, with a horizontal only to Christian. 11h → the focus at (20, 26). One set → rail at x 10; three sets → 4, 10, 16. 11d → siblings at x 20, rail at 10. Tiles listed left column then right column | D-12, FS-W1 |
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
