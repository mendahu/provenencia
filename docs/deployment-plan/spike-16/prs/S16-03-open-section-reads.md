# S16-03 — One open section and one list of section reads

**Status:** Landed in [#769](https://github.com/mendahu/provenencia/pull/769) (2026-11-09)
**Spike / slice:** Spike 16 · Slice 1b — Current data and honest notices
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Person page draws the same header, switcher, sections and states before and after)
**Board:** —
**Depends on:** —
**Blocks:** S16-04, S16-06, S16-25, S16-27, S16-29; Spike 15's S15-05 and S15-22 (shared files)
**Branch:** `s16-03-open-section-reads`
**Version bump:** none — a preparatory refactor with no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit](../person-page-audit.md) P-13, P-26 (part); working rule "(Batch 1) One open section, one list of reads"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) "Conclusion keys" (the "One detail key for every kind" and "One sectioned detail page" lines)
**Landed:** No version bump, as planned. One difference from the plan: the §6 sketch lists `pageIdentifier:` in the host's init, but D-7 and D-8 say the host takes `accessibilityIdentifierRoot:`; the decisions shipped. The header band's parameter was renamed to `openPageSection` too, from review MNT-1, superseding STEP-5's "keeps its name".
**Planned:** 2026-11-10 against `main` @ `9897bf19`
**Review:** [`S16-03-open-section-reads.review.md`](S16-03-open-section-reads.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blocker; To user: none

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

A Person page section declares what it reads once, and every part of the app that needs the open section finds it the same way. Today a section's reads are three slot functions (`queryKey`, `companionQueryKey`, `supportingQueryKeys`) that the place registry, the page's refreshing state and the section hosts each re-list, and the open section is resolved three ways (the place reads the raw `pageSection`, which is nil for the first section; the page uses `?? offered[0]`; the fill uses `?? .vitals`). So a first section that read its own key would be missing from the place's keys and a write would evict it while on screen (P-13). This PR adds `WorkspaceLocation.openPageSection` and one tagged read list per `ConclusionPageSection`, and first splits `ConclusionDetailPage.swift` (15 types in 736 lines) by moving types beside their kin (P-26, part).

**Why now.** S16-04 (Try again re-queries the open section's reads), S16-06 (section announcements) and S16-29 (a `prefetch` tag) all build on one list and one open section, and S16-04 and S16-06 edit `ConclusionDetailPage.swift`.

**What stays.** Everything the researcher sees and hears: the sections, their order, the switcher and its ⌘1–⌘5, the "Refreshing" line (it follows the open section's own and companion reads, not the Source types list), every state, every accessibility identifier, navigation history files (`pageSection` stays stored nil for the first section). The keys each section reads and the order the place warms them in.

## 2. User stories

**US-1.** As a contributor, when I give a page section a read, I add one entry to its read list, and the place warms it, the refreshing state follows it and the host reads it, whichever position the section has.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `ConclusionPageSection.reads(project:entityId:)` for each Person section, entity `e1` | Vitals `[]`; Gallery `[]`; Family `[own personFamily(e1)]`; Timeline `[own personEvents(e1)]`; Sources `[own personClaims(e1), companion personRecordLinks(e1), supporting sourceTypesList]`, in that order |
| AC-1.2 | `PlaceRegistry.standard.resolve` for a Person location on each section, the first one stored as nil included | its `queryKeys` are `conclusionDetail(e1)` followed by the keys of `location.openPageSection!.reads(…)` whose role `warmsWithPlace`, in order; Sources gives `[conclusionDetail, personClaims, personRecordLinks, sourceTypesList]` |
| AC-1.3 | `WorkspaceSession.apply(location:)` on a Person page open on Family | `visibleKeys` holds `personFamily(e1)`, so a write revalidates it instead of evicting it |
| AC-1.4 | `Role.showsRefreshing` and `Role.warmsWithPlace` for each role | `showsRefreshing`: own true, companion true, supporting false; `warmsWithPlace`: true for all three |

**US-2.** As a contributor, I find the open section in one place, so the page, the place and the fill cannot disagree.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `.personDetail(entityId: "per-1", …)` with `pageSection` nil, `.family`, `.timeline`, `.gallery`, `.sources` | `openPageSection` is `.vitals`, `.family`, `.timeline`, `.gallery`, `.sources` |
| AC-2.2 | `.eventDetail(…)`, `.placeDetail(…)`, `.sectionRoot(.persons)` | `openPageSection` is nil |
| AC-2.3 | `rg -n -F -e 'offeredPageSections[0]' -e 'pageSection ?? .vitals' macos/App` | no hit |

**US-3.** As a VoiceOver or UI-test user, the Person page's identifiers do not change.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ConclusionPageSection.family.accessibilityIdentifierRoot(pageIdentifier: "persons.detail")`, and `.timeline`, `.sources` | `persons.detail.family`, `persons.detail.timeline`, `persons.detail.sources` |
| AC-3.2 | The Sources diagram column's host, in the app | `persons.detail.sources.column` (manual check 3) |
| AC-3.4 | `rg -n '"persons\.detail\.' macos/App/Features/Conclusions` | One hit: `PersonDetailView.swift`'s `errorIdentifier: "persons.detail.error"` (the `pageIdentifier: "persons.detail"` line has no trailing dot). Every section identifier is built from its root |
| AC-3.3 | The app, a Person page on each section | Same screen and identifiers as before (manual check 1–3) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Move the gallery model to its own file, `ConclusionDetailRowsSection` beside `ReconciledValueRow`, and the Place relationship model and section beside `PlaceRelationshipRow`, out of `ConclusionDetailPage.swift` (P-26 part) | STEP-1 |
| `ConclusionPageSection` moves out of `WorkspaceLocation.swift` into its own file | STEP-2 |
| `ConclusionSectionRead` and `ConclusionPageSection.reads(project:entityId:)` replace `queryKey`, `companionQueryKey`, `supportingQueryKeys` | STEP-2, AC-1.1, AC-1.4 |
| `WorkspaceLocation.openPageSection` | STEP-3, AC-2.1, AC-2.2 |
| `PlaceRegistry`'s `.personDetail` spec reads `openPageSection` and the read list | STEP-4, AC-1.2, AC-1.3 |
| The page's open section and refreshing state read `openPageSection` and the read list; the fill switches on the context's section with no default | STEP-5, AC-2.3 |
| `ConclusionPageSectionHost` takes a read role and the page identifier; its id root, and every identifier inside the Family, Timeline and Sources sections and the record links column, come from the section's root | STEP-6, STEP-7, STEP-8, AC-3.1, AC-3.2, AC-3.4 |
| `add-workspace-location` teaches page sections; `add-workspace-place` and `macos-client-patterns.md` name the read list | STEP-9 |

| Out | Home |
| --- | --- |
| Event and Place page sections (`offered(by:)` stays Person-only) | not planned: Event and Place keep their own arrangement ([`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md)) |
| A per-case spec table for `label`, `glyph` and `fillsWidth` | not planned (map scope entry: the next section is one list entry either way) |
| Try again re-querying the open section's reads | S16-04 |
| Section switch announcements, the leave-or-stay decision | S16-06 |
| A `prefetch` read role | S16-29 |
| Keys as plain values, `.followsInputs` | S16-25 |
| The other `Features/Conclusions` tidy-ups of P-26 | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (P-26 row) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Section slots | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection` (lines 15–99): `queryKey(project:entityId:)`, `supportingQueryKeys(project:entityId:)`, `companionQueryKey(project:entityId:)` (38–70) | Three switches over the five cases; Family, Timeline, Sources own reads; Sources' companion `personRecordLinks` and supporting `sourceTypesList` |
| Stored section | same file `WorkspaceLocation.pageSection`, `normalizedPageSection` (450–461), `withPageSection(_:)` (391) | The first offered section is stored as nil, so `pageSection` alone does not name the open section |
| Place keys | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.personDetail` spec (191–203) | Appends `location.pageSection?.queryKey`, `companionQueryKey`, `supportingQueryKeys`: a nil `pageSection` (the first section) adds nothing, whatever that section reads |
| Visible keys | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(location:placeRegistry:)` (214–230), `effectivePolicy(for:)` (203) | `visibleKeys = Set(place.queryKeys)`; a key outside it follows its hidden policy (`conclusionDetail`-like person keys evict) |
| Page's open section | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailLoaded.offeredPageSections`, `pageSection` (483–490) | `offeredPageSections.isEmpty ? nil : (pageSection ?? offeredPageSections[0])` |
| Page's refreshing | same, `sectionRefreshing` (492–499) | Lists `queryKey` and `companionQueryKey` again |
| Fill | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailSections.body` (33–34) | `context.pageSection ?? .vitals` |
| Section host | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `key` (60–65), `idRoot` (72–74) | Takes `queryKey: CatalogQueryKey?` else `section.queryKey`; id root hard-coded `persons.detail.<section>` (+ `.column`) |
| Host callers | `PersonFamilySection.swift` (20), `PersonTimelineSection.swift` (22), `PersonSourcesSection.swift` (27 own, 206–209 companion) | The Sources column host passes `ConclusionPageSection.sources.companionQueryKey(…)` as `queryKey`. Sources reads `sourceTypesList` with `session.queryHandle` (24) |
| Crumbs | `macos/App/Features/Workspace/WorkspaceLocationLabels.swift` `labels(for:)` (46) | Reads the stored `location.pageSection` on purpose: the first section adds no crumb (`WorkspaceToolbarBreadcrumbTests.pageSectionAddsACrumb`, `locationLabelsNameTheSection`). Stays as it is |
| Section identifiers | `PersonFamilySection.swift` (97, 114), `PersonTimelineSection.swift` (93, 284, 324, 336, 392, 403, 408), `PersonSourcesSection.swift` (187, 264, 377, 403, 429, 466, 497, 506), `PersonRecordLinksColumn.swift` (50) | Each hard-codes `persons.detail.<section>…` |
| Section doc comments | `PersonFamilySection.swift`, `PersonTimelineSection.swift`, `PersonSourcesSection.swift` (line 5 each) | Name `ConclusionPageSection.queryKey`; Sources also says the place warms the Source types list beside it |
| Fill → sections | `PersonDetailView.swift` `PersonDetailSections` (55–64) | Passes `context.pageIdentifier` to the gallery only; Family, Timeline, Sources get `entityId` (and `userID`) |
| Big file | `ConclusionDetailPage.swift`: `ConclusionDetailRelationship` (34), `ConclusionDetailSection` (48), `ConclusionDetailPicture` (92), `ConclusionDetailGallery` with `ViewerCopy`, `LeadCopy` (120–240), `ConclusionDetailRowsSection` (417–462), `private ConclusionDetailRelationshipSection` (626–663) | 736 lines, 15 types |
| Move targets | `ReconciledValueRow.swift` (158 lines), `PlaceRelationshipRow.swift` (85) | Hold the row and the relationship row. `ConclusionDetailGallery` is read by seven files (`PersonDetailContent`, `EventDetailContent`, `PlaceDetailContent`, `ConclusionLeadModel`, `ConclusionPictureViewer`, `ConclusionGallerySection`, `ConclusionDetailPage`) |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | `App` is a file-system synchronized group: a new file under `macos/App/` needs no project edit |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys` (462–482) | Pins the three slot functions |
| Tests | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailOnFamilyWarmsTheFamily`, `personDetailOnTimelineWarmsTheTimeline`, `personDetailOnSourcesWarmsTheRecordLinks` (273–302) | Pin the place keys per section; stay green unchanged |
| Tests | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift` "Page sections" (757–856) | Pin storage, normalization, `offered(by:)` |
| Docs | `.agents/skills/add-workspace-place/SKILL.md` line 90; `docs/macos-client-patterns.md` lines 103–104; `.agents/skills/add-workspace-location/SKILL.md` | Name the three slots; the location skill never mentions page sections |

**Precedent.** `f4688953` ("Share one Show more row, one section host and one person card link between Family and the sections to come") folded per-section copies into one shared Conclusions type; copy its shape: one shared declaration, each caller reduced to the parts that differ. `5c209b6c` added `companionQueryKey` and is the slot this PR folds back in; read it to see every place a slot had to be repeated.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One type `ConclusionSectionRead` *(new)* `{ key: CatalogQueryKey, role: Role }` with `enum Role { case own, companion, supporting }`, declared beside `ConclusionPageSection` in `macos/App/Features/Workspace/ConclusionPageSection.swift` *(new)*, which `ConclusionPageSection` moves to verbatim from `WorkspaceLocation.swift` | The working rule names the three tags; a role enum leaves a slot for S16-29's `prefetch`. Read roles are a cache concern, not navigation history, so they leave the location file | A struct of three optional fields — keeps three slots under one name |
| D-2 | `ConclusionPageSection.reads(project:entityId:) -> [ConclusionSectionRead]` *(new)* is one `switch`; order is own, companion, supporting | Keeps the place's warm order (companion right after own) and today's key order in `PlaceRegistryTests` | Separate lists per role — the three-slot shape again |
| D-3 | `func read(_ role: ConclusionSectionRead.Role, project:entityId:) -> CatalogQueryKey?` *(new)* returns the first read of that role. A section has at most one own and one companion read | The host needs one key per role; supporting reads are read by the section itself with `session.queryHandle` | A host that takes a raw key — today's `queryKey:` override, which let the caller pick any key |
| D-4 | Each consumer's policy is an exhaustive `switch` property on `Role` *(new)*: `showsRefreshing` (own, companion true; supporting false) and `warmsWithPlace` (all three true). The page's refreshing state is `reads(…).filter(\.role.showsRefreshing).contains { session.isRefreshing($0.key) }`; the place keeps `reads(…).filter(\.role.warmsWithPlace)` | Keeps today's behavior (a catalog-wide vocabulary's reload shows no "Refreshing"), and a new role such as S16-29's `prefetch` does not compile until each policy is decided for it (review M-2) | `role != .supporting` — a new role silently shows Refreshing and warms at apply time |
| D-5 | `WorkspaceLocation.openPageSection: ConclusionPageSection?` *(new)* is `pageSection ?? ConclusionPageSection.offered(by: section).first` when `entityId != nil`, else nil | Same guard as `normalizedPageSection`; a list root has no page | Storing the first section explicitly — changes history files |
| D-6 | `ConclusionDetailPageSectionsContext.pageSection` and `ConclusionDetailLoaded`'s private `pageSection` are renamed `openPageSection`. The fill switches on `context.openPageSection` with a `case nil` that draws nothing; there is no `?? .vitals` | One word per concept: `pageSection` is the stored value (nil is the first), `openPageSection` the one shown. The page passes `openPageSection`, never nil for a Person | Keeping a default — a third resolution; keeping the name — the ambiguity that invited `?? .vitals` |
| D-7 | `ConclusionPageSectionHost` drops `queryKey:` and gains `accessibilityIdentifierRoot: String` (the section's root, D-8) and `read: ConclusionSectionRead.Role = .own`. Its `idRoot` is that root plus `.column` for the column placement, as today. Placement stays a separate parameter | The host resolves its key from the list; it takes the root its section already computed from the page identifier (the working rule), so the Sources diagram's host inside `PersonSourcesList` needs no page identifier; placement holds the loading height, which is view data | Deriving placement from the role — the column's height is not a read property |
| D-8 | `extension ConclusionPageSection { func accessibilityIdentifierRoot(pageIdentifier: String) -> String }` *(new)* in `ConclusionPageSectionHost.swift` gives `"\(pageIdentifier).\(rawValue)"`. Each section view computes it once from its `pageIdentifier` and passes it down one route: an init parameter `let accessibilityIdentifierRoot: String` *(new)* on every view that builds an identifier, which builds each as `"\(accessibilityIdentifierRoot).…"` with today's suffix. The views: the host; Family → `PersonFamilyDiagram`; Timeline → `PersonTimelineList` → `PersonTimelineItemView` → `PersonTimelineEventRow`, and `PersonTimelineList` → `PersonTimelineLine`; Sources → `PersonSourcesList` → `PersonSourcesRowView` → `PersonSourcesLineView`, and `PersonSourcesList` → the diagram's host and `PersonRecordLinksColumn`. Previews pass `"preview"` (`PersonRecordLinksColumn`'s preview already does). The name is `CircularLinkDiagram`'s parameter (`CircularLinkDiagram.swift` line 22) | A unit test can pin the root without a view; one source and one route for a section's prefix, so the host and its children cannot disagree (review M-1, M-11) | An environment value set by the host — it covers only the host's content, not the Sources diagram's second host, and the gallery's precedent is a parameter |
| D-9 | `PersonFamilySection`, `PersonTimelineSection` and `PersonSourcesSection` gain `let pageIdentifier: String`, passed from `context.pageIdentifier` in `PersonDetailSections`; they are the only views that take the page identifier. `PersonRecordLinksColumn` takes `accessibilityIdentifierRoot` (D-8) in place of its literal `"persons.detail.sources"` | The gallery is passed `pageIdentifier` the same way | — |
| D-10 | The split moves types only: no renames, no access changes except dropping `private` from `ConclusionDetailRelationshipSection` (it moves to another file and `ConclusionDetailDefaultPageSections` still calls it). The gallery models go to `macos/App/Features/Conclusions/ConclusionDetailGallery.swift` *(new)*, a model file beside its view as `ReconciledValueRowModel.swift` is; the rows section and the Place relationship model and section join their row files, as the audit says | Seven files read the gallery model; it is a model, not part of the gallery view. `App` is a synchronized group, so a new file needs no project edit | The gallery model in `ConclusionGallerySection.swift` (the audit's text) — a model inside a view file (review M-3) |
| D-11 | The Sources section reads its Source types through the list: `ConclusionPageSection.sources.read(.supporting, project: session.projectKey, entityId: entityId).flatMap { session.queryHandle($0) }` in place of the literal `CatalogQueryKey.sourceTypesList(…)` | The key is written once, in `reads`, so the place cannot warm one key while the section reads another (review M-8) | The literal key — a second place to write it |
| D-12 | Version: none | No visible behavior changes; the eviction P-13 describes cannot happen while Vitals, which reads nothing, is first | Patch — there is no release-worthy change |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ConclusionSectionRead` *(new)*, `ConclusionSectionRead.Role` *(new)* | type | `ConclusionPageSection` `queryKey` / `companionQueryKey` / `supportingQueryKeys` | They are replaced by it (searched `rg "QueryKeys\|Role" macos/App/Features/Workspace`: no tagged-read type exists) |
| `ConclusionPageSection.reads(project:entityId:)` *(new)* | func | the three slot functions | Replaces them |
| `ConclusionPageSection.read(_:project:entityId:)` *(new)* | func | `ConclusionPageSectionHost.key` | Host calls it; one lookup instead of the override |
| `ConclusionSectionRead.Role.showsRefreshing`, `.warmsWithPlace` *(new)* | property | `ConclusionDetailLoaded.sectionRefreshing`'s hand list; the place's unfiltered append | The policies move into data, one exhaustive switch each |
| `macos/App/Features/Workspace/ConclusionPageSection.swift` *(new)* | file | `WorkspaceLocation.swift` | `ConclusionPageSection` moves out verbatim; the location file keeps history and identity |
| `macos/App/Features/Conclusions/ConclusionDetailGallery.swift` *(new)* | file | `ConclusionGallerySection.swift`; `ReconciledValueRowModel.swift` beside `ReconciledValueRow.swift` | A model file beside its view, as the row's |
| `WorkspaceLocation.openPageSection` *(new)* | property | `ConclusionDetailLoaded.pageSection`, `PersonDetailSections`' `?? .vitals` | Replaces both and the place's raw read |
| `ConclusionPageSection.accessibilityIdentifierRoot(pageIdentifier:)` *(new)* | func | `ConclusionPageSectionHost.idRoot`, `ConclusionGallerySection`'s `"\(pageIdentifier).picture…"` | Testable root, computed once per section |
| `accessibilityIdentifierRoot` *(new)* on `ConclusionPageSectionHost`, `PersonFamilyDiagram`, `PersonTimelineList`, `PersonTimelineItemView`, `PersonTimelineEventRow`, `PersonTimelineLine`, `PersonSourcesList`, `PersonSourcesRowView`, `PersonSourcesLineView`, `PersonRecordLinksColumn` | property | `CircularLinkDiagram.accessibilityIdentifierRoot` | Same name and job, passed one level further up |
| `pageIdentifier` on three section views *(new)* | property | `ConclusionGallerySection.pageIdentifier` | Same parameter, same source; only the section views take it |

**Builds on.** `PlaceRegistry` specs, `WorkspaceLocation` normalization, `ConclusionPageSectionHost` (from `f4688953`), `ConclusionDetailPageSectionsContext.pageIdentifier`.
**Deletes.** `ConclusionPageSection.queryKey(project:entityId:)`, `supportingQueryKeys(project:entityId:)`, `companionQueryKey(project:entityId:)`; `ConclusionPageSectionHost`'s `queryKey` parameter and its `preconditionFailure` message wording; `ConclusionDetailLoaded.pageSection`'s `offeredPageSections[0]` and the fill's `?? .vitals`; the hard-coded `persons.detail.` in the host, the three sections and the record links column; the literal `sourceTypesList` key in `PersonSourcesSection`; `PersonDetailTests.sectionQueryKeys` (replaced by T-1).
**Next adds.** S16-04: Try again iterates `openPageSection.reads(…)`. S16-06: the announcement reads `openPageSection`. S16-29: one `Role` case `prefetch`, whose `showsRefreshing` and `warmsWithPlace` the compiler makes it decide (both false: it warms after the open section settles), and one entry each in Family and Timeline. S16-25: `reads` keeps its shape when keys become values. A kind that adopts sections adds its cases to `offered(by:)` and `reads`, gives its place spec the same open-section line as `.personDetail`, and supplies its own fill; Event and Place have none today.

### Reach

None — no new path into existing code, no new state reachable. `openPageSection` returns what the page already computed; the place gains the first section's reads, which are empty today.

### Scale

None — no write, list or loop changes. `reads` returns at most three entries; the refreshing state checks at most two keys, as today.

### Edges

- **Assumptions.** The host assumes a section has at most one own and one companion read (D-3); T-1 pins every section's list and each role's two policies. The place assumes `openPageSection` is non-nil for a Person detail location; T-3 pins it for every stored value, a decoded unknown section included (it normalizes to nil, so open is Vitals).
- **Neighbors.** A location with `pageSection == .vitals` passed explicitly normalizes to nil and opens Vitals (existing `pageSectionOnlyWhereOffered`); an Event location never gets a section (T-3).
- **Removed inputs.** None — no input is decided, excluded or deleted.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — the list is code, not data.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — Swift only.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section reads | `ConclusionSectionRead` *(new)*, `ConclusionPageSection.reads`, `.read` | `macos/App/Features/Workspace/ConclusionPageSection.swift` *(new)* | Sketch below |
| Open section | `WorkspaceLocation.openPageSection` *(new)* | `macos/App/Features/Workspace/WorkspaceLocation.swift` | |
| Place | `PlaceRegistry` `.personDetail` spec | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` | `[.conclusionDetail(…)] + (location.openPageSection?.reads(project:entityId:).filter(\.role.warmsWithPlace).map(\.key) ?? [])` |
| Page | `ConclusionDetailLoaded` | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` | private `pageSection` → `openPageSection` = `navigation.currentLocation.openPageSection`; the context field renamed (D-6); `sectionRefreshing` per D-4 |
| Fill | `PersonDetailSections` | `macos/App/Features/Conclusions/PersonDetailView.swift` | D-6, D-9 |
| Host | `ConclusionPageSectionHost` | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` | D-7, D-8 |
| Section identifiers | `PersonFamilySection`, `PersonTimelineSection`, `PersonSourcesSection`, `PersonRecordLinksColumn` | `macos/App/Features/Conclusions/` | D-8, D-9 |

```swift
/// One read a page section makes, tagged by how the page treats it. Each
/// consumer's policy is a switch on `Role`, so a new role must decide each.
struct ConclusionSectionRead: Equatable, Sendable {          // (new)
    enum Role: Equatable, Sendable {
        /// The section's own per-entity read; its host draws it in the section's band.
        case own
        /// A second per-entity read drawn in its own region (a column).
        case companion
        /// A catalog-wide vocabulary the rows label themselves with; its reload shows no Refreshing.
        case supporting

        /// The page's "Refreshing" follows this read.
        var showsRefreshing: Bool {
            switch self {
            case .own, .companion: return true
            case .supporting: return false
            }
        }

        /// The place warms this read when the location is applied, and keeps it visible.
        var warmsWithPlace: Bool {
            switch self {
            case .own, .companion, .supporting: return true
            }
        }
    }
    let key: CatalogQueryKey
    let role: Role
}

extension ConclusionPageSection {
    func reads(project: ProjectKey, entityId: String) -> [ConclusionSectionRead]   // (new)
    func read(_ role: ConclusionSectionRead.Role, project: ProjectKey, entityId: String) -> CatalogQueryKey?   // (new)
}

extension WorkspaceLocation {
    var openPageSection: ConclusionPageSection? { get }   // (new)
}
```

The host's init becomes `init(section:entityId:pageIdentifier:read: = .own, placement: = .band, errorTitle:emptyTitle:emptyMessage:make:isEmpty:body:)`. Its `key` is `section.read(read, project:entityId:)`, with `preconditionFailure("\(section.rawValue) has no \(read) read")` when nil.

States: unchanged. Accessibility: unchanged identifiers (AC-3.1, AC-3.2).

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — Split `ConclusionDetailPage.swift`**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Conclusions/ConclusionDetailGallery.swift` *(new)*, `ReconciledValueRow.swift`, `PlaceRelationshipRow.swift`.
Do: cut `ConclusionDetailPicture` and `ConclusionDetailGallery` (with its nested `ViewerCopy` and `LeadCopy` and their doc comments) to the new `ConclusionDetailGallery.swift` (starting `import SwiftUI`); `ConclusionDetailRowsSection` to the end of `ReconciledValueRow.swift`; `ConclusionDetailRelationship`, `ConclusionDetailSection` and `ConclusionDetailRelationshipSection` (dropping `private`) to the end of `PlaceRelationshipRow.swift`. Move each type's imports' needs with it (all four files `import SwiftUI`). Each type keeps its doc comment and its order. `ConclusionDetailThumbnail` and `ConclusionDetailBadge` stay. No other edit (D-10). This is its own commit.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` passes with no new warning; `wc -l macos/App/Features/Conclusions/ConclusionDetailPage.swift` is under 500; `git status` shows only these four files.

**STEP-2 — One read list per section**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/ConclusionPageSection.swift` *(new)*, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-swift-test`.
Do: first move `ConclusionPageSection` (its doc comment and all members) verbatim from `WorkspaceLocation.swift` to the new file, starting `import Foundation`; build. Then add `ConclusionSectionRead`, `reads(project:entityId:)` and `read(_:project:entityId:)` per the sketch, with doc comments carrying what the three old functions' comments said; keep the old three for this step only, reimplemented as `read(.own, …)`, `read(.companion, …)` and `reads(…).filter { $0.role == .supporting }.map(\.key)`. Replace `PersonDetailTests.sectionQueryKeys` with T-1, keeping its `fillsWidth` assertion.
Check: T-1 passes.

**STEP-3 — The open section**
Files: `WorkspaceLocation.swift`, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`. Skill: `add-workspace-location`.
Do: add `openPageSection` (D-5) beside `pageSection`, with a doc comment: "The section the page shows: the stored one, else the first the kind offers. Nil for a location with no entity or a kind without sections. The place, the page and the fill all read this." Add T-3 in the "Page sections" mark.
Check: T-3 passes.

**STEP-4 — The place reads the list**
Files: `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: the `.personDetail` spec's `queryKeys` becomes the line in the macOS table. Add T-2 and T-4.
Check: T-2, T-4 and the three existing `personDetailOn…` tests pass.

**STEP-5 — The page and the fill read the open section**
Files: `ConclusionDetailPage.swift`, `PersonDetailView.swift`.
Do: in `ConclusionDetailLoaded`, rename the private `pageSection` to `openPageSection` and return `navigation.currentLocation.openPageSection`; `offeredPageSections` stays for the switcher; `sectionRefreshing` per D-4. Rename `ConclusionDetailPageSectionsContext.pageSection` to `openPageSection` (its callers: `ConclusionDetailLoaded`, `PersonDetailSections`). In `PersonDetailSections`, `switch context.openPageSection` with `case nil: EmptyView()` and no `?? .vitals` (D-6). `ConclusionDetailHeaderBand`'s `pageSection` parameter keeps its name (Out: no rename the plan does not name).
Check: the Xcode tests pass; AC-2.3's `rg` finds nothing.

**STEP-6 — The host takes a role and a root; Family builds from its root**
Files: `ConclusionPageSectionHost.swift`, `PersonDetailView.swift`, `PersonFamilySection.swift`, `PersonFamilyPreviewData.swift`, `PersonTimelineSection.swift`, `PersonSourcesSection.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`.
Do: D-7 and D-8's helper in the host file; update the host's doc comment to name `reads`, the roles and the root. D-9: the three sections take `pageIdentifier` from `PersonDetailSections`, compute their root, and pass it to their own host. Family: `PersonFamilyDiagram` takes `accessibilityIdentifierRoot` and lines 97 and 114 build from it; the preview at `PersonFamilyPreviewData.swift:113` passes `"preview"`. Timeline and Sources pass their root to their own host only here; their inner literals are STEP-7 and STEP-8. Rewrite the doc comment at line 5 of each section file to say the section declares its reads in `ConclusionPageSection.reads` (Sources: its own claims, the record links as companion, the Source types as supporting). Add T-5.
Check: T-5 passes; `rg -n '"persons\.detail\.family' macos/App` finds no hit; the Xcode tests pass with no warning.

**STEP-7 — The Timeline builds from its root**
Files: `PersonTimelineSection.swift`.
Do: `PersonTimelineList`, `PersonTimelineItemView`, `PersonTimelineEventRow` and `PersonTimelineLine` take `accessibilityIdentifierRoot` (D-8), passed down from `PersonTimelineSection`; the literals at lines 93, 284, 324, 336, 392, 403 and 408 build from it with the same suffixes; `PersonTimelinePreview` (568) passes `"preview"`.
Check: `rg -n '"persons\.detail\.timeline' macos/App` finds no hit; the Xcode tests pass with no warning.

**STEP-8 — Sources builds from its root; the old slots go**
Files: `PersonSourcesSection.swift`, `PersonRecordLinksColumn.swift`, `ConclusionPageSection.swift`.
Do: `PersonSourcesList`, `PersonSourcesRowView`, `PersonSourcesLineView` and `PersonRecordLinksColumn` take `accessibilityIdentifierRoot` (D-8); the literals at lines 187, 264, 377, 403, 429, 466, 497, 506 and `PersonRecordLinksColumn.swift:50` build from it; `PersonSourcesPreview` (527) passes `"preview"`. In `PersonSourcesList.diagram`, the diagram's host passes `read: .companion` and the root, and the `let project = session.projectKey` local (line 203), whose only use was the old key, is deleted. D-11 for the Source types, keeping the `QueryHandle<[CatalogSourceType]>?` annotation so `queryHandle` infers its type; a missing read still yields `[]` through `typesHandle?.value ?? []`. Then delete `queryKey`, `companionQueryKey` and `supportingQueryKeys` from `ConclusionPageSection`.
Check: `rg -n "queryKey\(project|companionQueryKey|supportingQueryKeys|ConclusionPageSection\.queryKey" macos` finds no hit; AC-3.4's `rg` has its one hit; the Xcode tests pass with no warning.

**STEP-9 — Docs**
Files: `.agents/skills/add-workspace-location/SKILL.md`, `.agents/skills/add-workspace-place/SKILL.md` (line 90), `docs/macos-client-patterns.md` (lines 103–104), `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row), this plan (status).
Do: `add-workspace-location` gains a short "Page sections" section: a page section is a sub-place stored on `WorkspaceLocation.pageSection` (first stored as nil, frozen raw values), changed with `withPageSection(_:)` through `navigation.go(to:)`, and two properties with two jobs: `openPageSection` is the section the page shows (the place, the page and the fill read it), and the stored `pageSection` is history identity and the crumb, where the first section adds none (`WorkspaceLocationLabels`). A kind that adopts sections adds its cases to `offered(by:)` and `reads`, gives its place spec the open-section line, and supplies its own fill. `add-workspace-place` line 90 and `macos-client-patterns.md` 103–104 describe `reads(project:entityId:)`, its three roles and their two policies in place of the three slots, the host's `read:` and `accessibilityIdentifierRoot:` in place of `queryKey:`, the one route a section's identifier root takes, and `ConclusionDetailPageSectionsContext.openPageSection`. Then the DoD lines.
Check: `python3 scripts/check-guidance-names.py` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionReads` *(new, replaces `sectionQueryKeys`)* | Each section's `reads(project:entityId: "e1")` → the lists of AC-1.1; `read(.own/.companion/.supporting, …)` → the first of that role or nil; `Role.showsRefreshing` and `Role.warmsWithPlace` per role; `fillsWidth` as today | AC-1.1, AC-1.4, D-2, D-3, D-4 |
| T-2 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `personDetailWarmsEveryReadOfTheOpenSection` *(new)* | For `pageSection` in `[nil] + allCases`: `resolve(.personDetail(…).withPageSection(it))?.queryKeys` → `[conclusionDetail] + openPageSection!.reads(…).filter(\.role.warmsWithPlace).map(\.key)` | AC-1.2 |
| T-3 | Swift | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift` `openPageSectionIsTheStoredOrFirst` *(new)* | Person on nil / each case → `.vitals` / that case; decoded `"pageSection":"lineage"` → `.vitals`; event, place, persons root → nil | AC-2.1, AC-2.2, D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `applyPersonPageKeepsTheOpenSectionsReadsVisible` *(new)* | Setup as `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`: `makeSession(store: FakeStore())`, no seeding (the test reads `visibleKeys` only and awaits no load). `session.apply(location: .personDetail(…, pageSection: .family))` → `visibleKeys` contains `.personFamily(project:entityId:)`; with `pageSection` nil → `visibleKeys == [conclusionDetail]` | AC-1.3 |
| T-5 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionIdentifierRoots` *(new)* | `.family`/`.timeline`/`.sources` `accessibilityIdentifierRoot(pageIdentifier: "persons.detail")` → `persons.detail.<raw>` | AC-3.1, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PlaceRegistryTests` (key order), `WorkspaceNavigationTests` "Page sections", `PersonDetailTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md`, `L10n.swift` and `macos/App` (doc comments included) with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md "Conclusion keys" (STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

The sweep's old names: `ConclusionPageSection.queryKey`, `companionQueryKey`, `supportingQueryKeys`, `queryKey:` on `ConclusionPageSectionHost`, `ConclusionDetailPageSectionsContext.pageSection`. No version bump (D-12). No brief.

## 10. Manual check

On macOS with Xcode: the Xcode tests and these checks need the app.

1. (AC-3.3) Open a project with a person who has relatives, events and records in at least two Evidence graphs. Open the person: Vitals shows with no spinner in the switcher once loaded. Press ⌘2, ⌘3, ⌘4, ⌘5: each section draws as before, Sources with its diagram column at Wide.
2. On Sources, rename one of the person's Sources in another window: "Refreshing" shows beside the switcher while the claims and record links reload, as before.
3. (AC-3.2, AC-3.3) With Accessibility Inspector, check the Family section's root is `persons.detail.family`, a family tile's is `persons.detail.family.tile.<id>`, a Timeline row's title is `persons.detail.timeline.title.<id>`, a Sources line is `persons.detail.sources.line.<id>`, and the Sources diagram column's is `persons.detail.sources.column`.
6. (AC-3.4) Run AC-3.4's `rg`: its one hit is `PersonDetailView.swift`'s `errorIdentifier`.
4. Quit and reopen the project on Timeline: the page reopens on Timeline.
5. (AC-2.3) Run `rg -n -F -e 'offeredPageSections[0]' -e 'pageSection ?? .vitals' macos/App`: no hit.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The split changes behavior through access or extension order | D-10 moves types verbatim; STEP-1 is its own commit with only four files in its diff |
| Key order changes and an existing test or warm order breaks | D-2 keeps own, companion, supporting; the three existing `PlaceRegistryTests` stay unchanged |
| S16-07 or Spike 15 has edited `PersonSourcesSection.swift` since the base | Implementer rule 2: compare the rows of §4 first |

**Stop and report if:**

- `ConclusionPageSection` gains a section with two own or two companion reads (D-3 no longer holds).
- A type in STEP-1 cannot move without changing its access level beyond D-10, or without a new file.
- `ConclusionPageSectionHost` has a caller outside `PersonFamilySection.swift`, `PersonTimelineSection.swift` and `PersonSourcesSection.swift`.
- Any existing `PlaceRegistryTests` or `WorkspaceNavigationTests` case fails after STEP-4 without a test edit this plan names.
- A section identifier cannot be rebuilt from its root with the same value (AC-3.3), or a view named in D-8 cannot take the parameter without a change the plan does not name.
- T-4's unseeded `FakeStore` loads throw or crash the test instead of failing quietly; report it rather than seeding a store.
- A UI test or script in the repo reads an identifier this PR touches and expects a value other than AC-3.1 / AC-3.2.

## 12. Open questions

None.

## 13. PR

**Title:** S16-03 — One open section and one list of section reads
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-03-open-section-reads.md`.
