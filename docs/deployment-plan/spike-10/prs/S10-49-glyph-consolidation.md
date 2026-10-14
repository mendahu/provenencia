# S10-49 — Glyph consolidation

**Status:** Landed in [#394](https://github.com/mendahu/provenencia/pull/394) (2026-10-14)
**Spike / slice:** Spike 10 · Slice 5 — Artifacts and multi-select (follow-up to S10-19)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-49--glyph-consolidation)
**Brief:** — (no new chrome: the PR moves existing icons onto one value; the only visible changes are two empty-state icons that now match their sidebar rows, which the user chose on 2026-10-14)
**Board:** —
**Depends on:** S10-19 (landed [#393](https://github.com/mendahu/provenencia/pull/393))
**Blocks:** —
**Branch:** `s10-49-glyph-consolidation`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [glyph consolidation idea](../../../ideas/archive/glyph-consolidation.md); [`S10-19-event-mark-sidebar.ship.md`](S10-19-event-mark-sidebar.ship.md) (MNT-1 / CON-2); [`macos/App/DesignSystem/Recipes/Marks/MARKS.md`](../../../../macos/App/DesignSystem/Recipes/Marks/MARKS.md); working rule "One Event mark"
**Planned:** 2026-10-14 against `main` @ `e150a76a`
**Review:** [`S10-49-glyph-consolidation.review.md`](S10-49-glyph-consolidation.review.md) — Fits with changes, Ready with changes (applied); 1 round

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

S10-19 added `PVGlyph` (a symbol or a mark) for the shell. Its review found older code that still makes the same choice by hand, in three different shapes. It also found the subject-type → mark mapping written out in about ten places. This PR moves those sites onto one value and one mapping, and trims the S10-19 test that repeats the production `switch`.

The researcher sees almost nothing change. Two empty states do change: the Source types and Metadata lists now show the same icon as their sidebar row.

**Why now.** The S10-19 follow-ups are fresh, and the user asked for them in one PR (2026-10-14).

**What stays.** Every mark and symbol a researcher sees stays the same, except the two empty states above. Each thumbnail still draws its mark or symbol at the same size and ink, and the Properties "All" card keeps its grid symbol. `ConclusionListKind.mark` and the detail views' `mark:` stay literal constants.

## 2. User stories

**US-1.** As a researcher, when a Source types or Metadata list is empty, its empty state shows the same icon as its sidebar row, so the page matches the row I clicked.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A project whose Source types list is empty → open Source types | The empty state shows the tag symbol (was library) over its existing title and body |
| AC-1.2 | A project with no Metadata fields → open Metadata | The empty state shows the list symbol (was tag) |

**US-2.** As a contributor, when I add a subject type, its mark comes from one constructor and the Go registry pins the same naming rule, so I do not edit a switch in each feature.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `PVMarkKey(subjectTypeKey:)` with each of `person`, `event`, `place`, `relationship`, `participation`, `location`, `place_relationship`, `source` | `.subjectPerson` … `.subjectSource`, one each |
| AC-2.2 | `PVMarkKey(subjectTypeKey:)` with `""`, `"source_type"`, `"widget"` | `nil` |
| AC-2.3 | Every Go registry subject type | `Presentation.IconSymbol == "subject_" + Key`, and the `subject_*` raw values in Swift's `enum PVMarkKey` equal the sorted set of registry `IconSymbol`s |
| AC-2.4 | After the PR: `rg -n 'stripMarkKey|leadMark|leadSymbol|"subject_\\\(|"subject_(relationship|participation|location|place_relationship)"' macos/App` | Prints nothing (it prints 20 lines at the planning base) |

**US-3.** As a contributor, a thumbnail and a type card take one `PVGlyph`, so a call site cannot pass a mark and a symbol at once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `PVThumbnail.Content` | Has `glyph: PVGlyph?` and no `mark` or `icon` field |
| AC-3.2 | Omnibar lead for kind `event` | `.mark(.subjectEvent)`, label `nil` |
| AC-3.3 | Omnibar lead for kind `source_type` with icon key `type_book` | `.mark(.typeBook)`, label the kind label; with an empty icon key, `.symbol(.library)` |
| AC-3.4 | Omnibar lead for `metadata_field`, and for an unknown kind | `.symbol(.tag)`, `.symbol(.search)` |
| AC-3.5 | Sources list, omnibar, citation composer, Persons list, Properties strip | Look as they did before this PR (manual) |

**US-4.** As a contributor, the S10-19 section-glyph test checks keys, not a copy of the `switch`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `sectionGlyphsMatchTheBoard` | Has no `for` loop and no `switch`; asserts the three Conclude sections' literal marks |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PVMarkKey.init?(subjectTypeKey:)` *(new)*; a Go test pinning `IconSymbol == "subject_" + Key`; a `TestSwiftVocabularyMatchesGo` row matching Swift's `subject_*` cases to the registry | STEP-1, AC-2.1, AC-2.2, AC-2.3 |
| Every string-keyed subject-type → mark site uses that constructor: `PropertiesTypeChrome.stripMarkKey` (deleted), the composer's `"subject_\(…)"` keys and `kindChrome` mark strings, `CitationComposerSubjectGroups` | STEP-2, AC-2.4 |
| `PVThumbnail.Content.glyph` replaces `mark` and `icon`, and every call site moves | STEP-3, AC-3.1, AC-3.5 |
| `OmnibarHitPresentation.leadThumbnail(kind:iconKey:)` *(new)* replaces `leadMark` and `leadSymbol` | STEP-4, AC-3.2…AC-3.4 |
| Properties `typeCard` takes a `PVGlyph`; `PVSymbol.grid` *(new)* names its "All" symbol | STEP-5, AC-3.5 |
| `VocabularyListStrings.emptyGlyph` replaces `emptyIcon`; Source types and Metadata pass their section glyph | STEP-6, AC-1.1, AC-1.2 |
| Trim `sectionGlyphsMatchTheBoard` | STEP-7, AC-4.1 |
| Docs: `MARKS.md`, DesignSystem README, `add-seeded-vocabulary` skill, idea archived with its leftovers, S10-19 plan links | STEP-9, STEP-10 |

| Out | Home |
| --- | --- |
| `ConclusionListKind.mark` and the detail views' `mark:` literals (`.subjectPerson`, `.subjectEvent`, `.subjectPlace`) | not planned; they are constants, not mappings |
| `EvidencePrimaryKind.markKey` / `EvidenceBridgeKind.markKey` switches | not planned; exhaustive switches over closed enums are typed constants, like `ConclusionListKind.mark` (D-3); T-3 checks they agree with the rule |
| `FakeStore` `syntheticPresentation`'s `iconSymbol: typeKey` | not planned; no Swift code reads `iconSymbol` (D-2) |
| Omnibar lead symbols for Source types (`.library`) and Metadata fields (`.tag`), which differ from their sidebar glyphs | leftover in the archived idea note; a visible change the user has not chosen |
| `PropertiesTypeChrome.sortIndex`, `showsBridgeLabel`, `ink`, `CitationComposerAuthoring.typeRank`, `FakeStore.syntheticPresentation` (per-type switches over data the registry carries) | leftover in the archived idea note |
| Reading `CatalogSubjectTypePresentation.iconSymbol` at runtime instead of the naming rule | not planned; the sidebar and lists draw before any catalog read, and the Go test pins the rule (D-2) |
| `CachedThumbnail`'s raster / well layout, `PVThumbnail` sizes and inks | not planned (D-4 keeps them) |
| Properties pane empty state, other empty states | not planned; only the two vocabulary panes drifted |
| Changing the `PVThumbnail` or `PVEmptyState` visuals | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Glyph value | `macos/App/DesignSystem/Recipes/Marks/PVMark.swift` `PVGlyph` (line 381), `PVGlyphView` (line 388), `PVMarkKey.fallback` (line 68) | `PVGlyph { case symbol(PVSymbol); case mark(PVMarkKey) }`. `PVMarkKey(catalogKey:)` falls back to `.typeEvidence`. No subject-type constructor. |
| Go registry | `core/database/subjectvocab/registry.go` `IconSymbol` | Every entry's `IconSymbol` is `"subject_" + Key` (8 entries). `core/database/subjectvocab/subjectvocab_test.go` checks only that it is non-empty, and `subject_person` for person. |
| Drift test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, `swiftEnumBody` | Table rows read a Swift enum body and compare its cases with Go values. No row for marks. `enum PVMarkKey` closes at column 0, so `swiftEnumBody` can read it. |
| Unused field | `macos/App/Platform/GenealogyStore.swift:658` `CatalogSubjectTypePresentation.iconSymbol` | No Swift reader. `FakeStore.swift:2114` sets it to `typeKey`. |
| Properties mapping | `macos/App/Features/Catalog/SubjectTypeChrome.swift` `PropertiesTypeChrome.stripMarkKey(typeKey:)` | An 8-case switch equal to `"subject_" + key`. Callers: `macos/App/Features/Properties/PropertiesView.swift` lines 181, 605, 906. |
| Evidence kinds | `macos/App/Features/EvidenceGraph/EvidenceCanvasInputMode.swift` `EvidencePrimaryKind.markKey`, `EvidenceBridgeKind.markKey` | Switches. The enums' raw values (`macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift`) are the type keys, `placeRelationship = "place_relationship"`. |
| Composer marks | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` lines 170 (`markKey: String`), 192 (`markKey: String?`), 321 (`markKey: nil`), 363, 463 (`markKey: chrome.mark`), `kindChrome` (467) | Mark keys are strings built `"subject_\(…)"` or spelled out per bridge kind, then read with `PVMarkKey(catalogKey:)` in `CitationComposerAuthoringBox.swift:677` and `CitationComposerSubjectGroups.swift:142` |
| Composer groups | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` lines 96, 429 | `PVMark(PVMarkKey(catalogKey: "subject_\(…)"), …)` |
| Thumbnail | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.Content` | `mark: PVMarkKey?` and `icon: PVSymbol?`; mark wins. A mark draws at `size * 0.82` in `textSecondary`, decorative when `label == nil`; a symbol draws at `max(12, size * 0.41)` in `textFaint`. Accessibility label: `label`, else the mark's `accessibilityName`, else loading / empty. |
| Thumbnail callers | `macos/App/Platform/CachedThumbnail.swift` (lines 53, 90, 94), `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` (259), `macos/App/Features/Conclusions/ConclusionListRow.swift` (30), `macos/App/DesignSystem/Components/List/PVList.swift` (308, preview), `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift` (96, 106, previews), `macos/App/Features/Workspace/OmnibarResultsHost.swift` (227–245), `PVThumbnail` `#Preview` | Build `Content(mark:)` / `Content(icon:)`; `CachedThumbnail.wellBody` reads `content.mark` |
| Omnibar lead | `macos/App/Features/Workspace/OmnibarHitPresentation.swift` `leadMark(for:)` (100), `leadSymbol(for:)` (109); `OmnibarResultsHost.leadView` | Host switches on kind: `source` → `CachedThumbnail`; person/event/place → `ConclusionListRow.thumbnail(mark:)` with `leadMark`; `source_type` → mark from `iconKey` with the kind label, or `.library`; default → `leadSymbol` (its `"source"` arm is unreachable). Siblings (`title(for:)`, `secondary(for:)`) take a `CatalogSearchHit`. |
| Properties card | `macos/App/Features/Properties/PropertiesView.swift` `typeCard(… mark: PVMarkKey?)` (190) | `if let mark { PVMark(mark, size: 16).foregroundStyle(ink) }` (named, not decorative) `else { Image(systemName: "square.grid.2x2").font(.system(size: 14, weight: .medium)).foregroundStyle(PVColor.textPrimary).frame(width: 16, height: 16) }`. The "All" card passes `mark: nil`, `ink: PVColor.textPrimary`. |
| Vocabulary empty | `macos/App/Features/CatalogVocabulary/VocabularyListPane.swift` `VocabularyListStrings.emptyIcon` (8), `PVEmptyState(icon:)` (66); `macos/App/Features/SourceTypes/SourceTypesListPane.swift:29` `.library`; `macos/App/Features/Metadata/MetadataListPane.swift:31` `.tag` | Sidebar rows use `.tag` (Source types) and `.list` (Metadata) |
| Section test | `macos/ProvenenciaTests/WorkspaceSidebarSectionsTests.swift` `sectionGlyphsMatchTheBoard` | A loop whose `switch` copies `WorkspaceSection.glyph`, then three literal `#expect`s |
| Omnibar test | `macos/ProvenenciaTests/OmnibarResultsModelTests.swift:165` | `#expect(OmnibarHitPresentation.leadMark(for: "event") == .subjectEvent)` |

**Precedent.** S10-19 ([#393](https://github.com/mendahu/provenencia/pull/393)) moved `WorkspaceSection`, `PVSidebarNavItem`, and `PVEmptyState` onto `PVGlyph`; copy its shape (one field, `PVGlyphView` where size and ink are uniform). `PVMarkKey(catalogKey:)` in `PVMark.swift` is the constructor pattern.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Add `init?(subjectTypeKey: String)` *(new)* on `PVMarkKey`, in `PVMark.swift` beside `init(catalogKey:)`: `PVMarkKey(rawValue: "subject_" + subjectTypeKey)`, returning `nil` unless the result's `family == .subject`. Callers that need a mark for any key write `?? .fallback`. | One mapping. Failable keeps today's "no mark for an unknown type" in Properties. | A non-failable init that falls back: Properties would draw the evidence mark for an unknown type instead of nothing. |
| D-2 | The rule `subject_<key>` is pinned twice in Go tests: `subjectvocab_test.go` asserts `IconSymbol == "subject_" + Key` for every `AllTypes()` entry, and a new `TestSwiftVocabularyMatchesGo` row reads the `case \w+ = "(subject_[a-z_]+)"` raw values from `enum PVMarkKey` in `macos/App/DesignSystem/Recipes/Marks/PVMark.swift` and compares them with the sorted registry `IconSymbol`s. Swift does not read `iconSymbol`, and `FakeStore` is not changed. | The sidebar and lists draw before a catalog read, so Swift needs a static rule; the two tests close the chain Go key → `IconSymbol` → Swift case, so a new subject type without a mark fails CI (review M-1). | Reading `iconSymbol` at runtime: not available to static chrome. Changing `FakeStore`'s unused `iconSymbol`: it would bring back a `"subject_\(…)"` string for a field nobody reads (review M-2). |
| D-3 | `PropertiesTypeChrome.stripMarkKey` is deleted; its three callers use `PVMarkKey(subjectTypeKey:)`. The composer's `markKey` fields become `PVMarkKey` (line 170) and `PVMarkKey?` (line 192). `kindChrome` returns `(label:, hint:)` only; the caller at line 463 sets `markKey: PVMarkKey(subjectTypeKey: rule.bridgeTypeKey) ?? .fallback`. Line 363 sets `PVMarkKey(subjectTypeKey: connection.bridgeTypeKey) ?? .fallback`. Line 321 stays `nil`. The two readers (`CitationComposerAuthoringBox.swift:677`, `CitationComposerSubjectGroups.swift:142`) pass the key straight to `PVMark`. `CitationComposerSubjectGroups` lines 96 and 429 use `PVMarkKey(subjectTypeKey: …) ?? .fallback`. `EvidencePrimaryKind.markKey` and `EvidenceBridgeKind.markKey` keep their exhaustive switches. | Same marks as today: `PVMarkKey(catalogKey:)` fell back to `.fallback` for an unknown string too. The string sites are the duplication; the evidence switches are compiler-checked constants over closed enums, like `ConclusionListKind.mark` (review M-3). | Moving the evidence switches to the rule: adds an unreachable fallback for no gain. Keeping `markKey` as `String`. |
| D-4 | `PVThumbnail.Content` replaces `mark` and `icon` with `glyph: PVGlyph?`; `init(image:glyph:loading:label:)`. Drawing keeps today's per-case size and ink (mark: `size * 0.82`, `textSecondary`, decorative when `label == nil`; symbol: `max(12, size * 0.41)` rounded, `textFaint`), so it switches on the glyph instead of using `PVGlyphView`. The private `State` cases stay. Accessibility keeps `label`, else a mark's `accessibilityName`. | One field; no visual change. | `PVGlyphView`: one size and ink for both cases would change every thumbnail. |
| D-5 | `CachedThumbnail.wellBody` reads `if case .mark(let key)? = content.glyph`. Its builders pass `glyph: .mark(key)`. | Same behavior. | — |
| D-6 | `OmnibarHitPresentation.leadThumbnail(for hit: CatalogSearchHit) -> PVThumbnail.Content` *(new)* replaces `leadMark` and `leadSymbol` (both deleted). `"person"`, `"event"`, `"place"` → `ConclusionListRow.thumbnail(mark: PVMarkKey(subjectTypeKey: hit.kind) ?? .fallback)`; `"source_type"` → `Content(glyph: .mark(PVMarkKey(catalogKey: hit.iconKey)), label: kindLabel(for: hit.kind))` when `hit.iconKey` is not empty, else `Content(glyph: .symbol(.library))`; `"metadata_field"` → `.symbol(.tag)`; any other kind → `.symbol(.search)`. There is no `"source"` case: `OmnibarResultsHost.leadView` keeps `"source"` → `CachedThumbnail` and calls `PVThumbnail(OmnibarHitPresentation.leadThumbnail(for: hit), size: 40)` for every other kind. `ConclusionListRow.thumbnail(mark:)` keeps its signature; its body becomes `PVThumbnail.Content(glyph: .mark(mark), label: nil)`. | One testable function owns the lead, with the siblings' `for hit:` convention (review M-8). The omnibar and the Conclusion lists share one thumbnail definition, so S10-35 has one place to look (review M-4). The old `"source"` symbol was unreachable (review M-5). | Leaving the choice in the view: untestable. |
| D-7 | `typeCard` takes `glyph: PVGlyph` in place of `mark: PVMarkKey?` and switches on it: `.mark(let key)` → `PVMark(key, size: 16).foregroundStyle(ink)` (named, as today); `.symbol(let symbol)` → `PVIcon(symbol, size: 14).frame(width: 16, height: 16).foregroundStyle(ink)`. The "All" card passes `.symbol(.grid)` (its `ink` is already `PVColor.textPrimary`); a type card passes `PVMarkKey(subjectTypeKey: type.key).map(PVGlyph.mark) ?? .symbol(.grid)`. `PVSymbol.grid = "square.grid.2x2"` *(new)*. The grid symbol's medium font weight becomes `PVIcon`'s default rendering; that weight shift on the All card is accepted. | One value per card; size and VoiceOver naming stay as today. | `PVGlyphView`: one size for both cases and always decorative, which would change the card (review X-2). Keeping the raw `Image(systemName:)`. |
| D-8 | `VocabularyListStrings.emptyIcon: PVSymbol` becomes `emptyGlyph: PVGlyph`; `VocabularyListPane` calls `PVEmptyState(glyph:title:message:)`. `SourceTypesListPane` passes `WorkspaceSection.sourceTypes.glyph`, `MetadataListPane` passes `WorkspaceSection.metadata.glyph`. | The user chose to fix the drift (2026-10-14). The section owns the glyph. | Hard-coding `.tag` / `.list`: a second copy. |
| D-9 | `sectionGlyphsMatchTheBoard` keeps only `#expect(WorkspaceSection.persons.glyph == .mark(.subjectPerson))` and the events and places lines. | T-1 (`sidebarRowsCarryTheSectionGlyph`) already pins all seven literals; the loop restated production code (MNT-1 / CON-2). | Deleting the test: the section-level literal check is cheap. |
| D-10 | New Swift tests go in `macos/ProvenenciaTests/PVGlyphTests.swift` *(new)*, wired into the `ProvenenciaTests` group and its Sources phase in `project.pbxproj` with the ID scheme `add-swift-test` gives. The Go assertions go in the existing `subjectvocab_test.go` loop over `AllTypes()` and in `TestSwiftVocabularyMatchesGo`. The omnibar test edits `OmnibarResultsModelTests`. | No existing suite owns marks. | Putting mark tests in `EvidenceCitedPropertyMarksTests`: that suite is about cited properties. |
| D-11 | No new user-facing strings. | Labels are unchanged. | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVMarkKey.init?(subjectTypeKey:)` *(new)* | init | `PVMarkKey(catalogKey:)`, `PropertiesTypeChrome.stripMarkKey` (deleted) | It replaces the switch; `catalogKey:` takes a full key, not a type key. |
| `OmnibarHitPresentation.leadThumbnail(for:)` *(new)* | func | `leadMark`, `leadSymbol` (deleted), `ConclusionListRow.thumbnail(mark:)` | It merges the two and the host's inline `source_type` branch, and calls `ConclusionListRow.thumbnail` for subject kinds. |
| `PVThumbnail.Content.glyph` *(new)* | property | `Content.mark`, `Content.icon` (deleted) | Same slot, one field. |
| `VocabularyListStrings.emptyGlyph` *(new)* | property | `emptyIcon` (deleted) | Same slot. |
| `PVSymbol.grid` *(new)* | case | raw `Image(systemName: "square.grid.2x2")` | Names the symbol in the kit. |
| `macos/ProvenenciaTests/PVGlyphTests.swift` *(new)* | test file | `WorkspaceSidebarSectionsTests` | Marks are not sidebar behavior. |
| `TestSwiftVocabularyMatchesGo` row "subject marks" *(new)* | test row | the existing table rows | It is one more row in the repo's Swift ↔ Go drift test. |

**Builds on.** `PVGlyph` / `PVGlyphView` (S10-19), `PVMarkKey(catalogKey:)`, the Go `subjectvocab` registry.
**Deletes.** `PropertiesTypeChrome.stripMarkKey`, `OmnibarHitPresentation.leadMark` and `leadSymbol`, `PVThumbnail.Content.mark` and `.icon`, `VocabularyListStrings.emptyIcon`, the composer's string mark keys and `kindChrome`'s mark strings, the loop in `sectionGlyphsMatchTheBoard`, and the idea note (archived).
**Next adds.** A new subject type needs a registry entry, a `subject_<key>` `PVMarkKey` case, its asset, and its accessibility name; `TestSwiftVocabularyMatchesGo` fails until the case exists. Properties, the composer, and the omnibar then pick up the mark with no switch to edit; the evidence-kind enums gain a case the compiler asks for.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/subjectvocab` | Test only: in the existing `AllTypes()` loop in `subjectvocab_test.go`, fail when `info.Presentation.IconSymbol != "subject_"+info.Key`. |
| `api/ffi` | Test only: a `TestSwiftVocabularyMatchesGo` row `subject marks`, with a new const for the `PVMark.swift` path, regex `case \w+ = "(subject_[a-z_]+)"`, enum `PVMarkKey`, and `want` the sorted `IconSymbol` of every `subjectvocab.AllTypes()` entry. |

### FFI

None — no RPC change.

### macOS

```swift
extension PVMarkKey {
    /// The `subject_*` mark for a subject-type key (`person` → `subject_person`).
    /// Nil when no such mark exists. Mirrors the Go registry's `IconSymbol` rule.
    init?(subjectTypeKey: String)
}
```

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Subject mark rule | `PVMarkKey(subjectTypeKey:)` | `macos/App/DesignSystem/Recipes/Marks/PVMark.swift` | D-1 |
| Thumbnail | `PVThumbnail` (Extend) | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` | D-4 |
| Omnibar lead | `OmnibarHitPresentation.leadThumbnail` | `macos/App/Features/Workspace/OmnibarHitPresentation.swift` | D-6 |
| Properties card | `typeCard` | `macos/App/Features/Properties/PropertiesView.swift` | D-7 |
| Vocabulary empty | `VocabularyListPane` | `macos/App/Features/CatalogVocabulary/VocabularyListPane.swift` | D-8 |

States: unchanged. Accessibility: unchanged; `PVThumbnail` keeps its label rules.

### Strings

None — no copy changes (D-11).

## 7. Steps

**STEP-1 — One subject-mark rule**
Files: `macos/App/DesignSystem/Recipes/Marks/PVMark.swift`, `core/database/subjectvocab/subjectvocab_test.go`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/PVGlyphTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-1, D-2. Add T-1, T-2, T-7. Wire the new test file into the `ProvenenciaTests` target.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectvocab/... ./api/ffi/...` passes. `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PVGlyphTests CODE_SIGNING_ALLOWED=NO` passes.
Commit: "Map a subject type to its mark in one place, pinned to the registry on both sides."

**STEP-2 — Use the rule at the string-keyed sites**
Files: `macos/App/Features/Catalog/SubjectTypeChrome.swift`, `macos/App/Features/Properties/PropertiesView.swift` (lines 605 and 906 only in this step), `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift`, `macos/ProvenenciaTests/PVGlyphTests.swift` *(new)*, created in STEP-1.
Do: D-3. Line 181's `stripMarkKey` call becomes `PVMarkKey(subjectTypeKey: type.key)` in this step; STEP-5 changes that call again. Add T-3.
Check: `rg -n 'stripMarkKey|"subject_\\\(|"subject_(relationship|participation|location|place_relationship)"' macos/App` prints nothing. The STEP-1 `xcodebuild test` command passes.
Commit: "Read the composer and Properties subject marks through the shared rule."

**STEP-3 — Thumbnails take a glyph**
Files: `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift`, `macos/App/Platform/CachedThumbnail.swift`, `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `macos/App/Features/Conclusions/ConclusionListRow.swift`, `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift`, `macos/App/Features/Workspace/OmnibarResultsHost.swift`. Skill: `add-ui-component`.
Do: D-4, D-5. In `OmnibarResultsHost`, change only the `Content(mark:)` / `Content(icon:)` arguments to `glyph:` here; STEP-4 replaces the switch. `CitationComposerEvidenceColumn`: `glyph: artifactMark(artifact).map(PVGlyph.mark)`.
Check: `rg -n -U 'Content\(\s*(mark|icon):|content\.(mark|icon)' macos` prints nothing. `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO` succeeds.
Commit: "Give thumbnails one glyph instead of a mark and a symbol."

**STEP-4 — Omnibar lead in one function**
Files: `macos/App/Features/Workspace/OmnibarHitPresentation.swift`, `macos/App/Features/Workspace/OmnibarResultsHost.swift`, `macos/ProvenenciaTests/OmnibarResultsModelTests.swift`.
Do: D-6. Replace the line-165 assertion with T-4, building each `CatalogSearchHit` the way the suite's other tests do.
Check: `rg -n "leadMark|leadSymbol" macos` prints nothing. `xcodebuild test … -only-testing:ProvenenciaTests/OmnibarResultsModelTests …` (same flags as STEP-1) passes.
Commit: "Choose the omnibar lead thumbnail in one presentation function."

**STEP-5 — Properties type card takes a glyph**
Files: `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/DesignSystem/Components/Icon/PVIcon.swift`.
Do: D-7. Add `case grid = "square.grid.2x2"` to `PVSymbol` with a one-line doc comment ("Lucide `layout-grid` — the Properties All card.").
Check: `rg -n "square.grid.2x2" macos/App` prints only `PVIcon.swift`. The STEP-3 build command succeeds.
Commit: "Draw the Properties type cards through the shared glyph."

**STEP-6 — Vocabulary empty states match their rows**
Files: `macos/App/Features/CatalogVocabulary/VocabularyListPane.swift`, `macos/App/Features/SourceTypes/SourceTypesListPane.swift`, `macos/App/Features/Metadata/MetadataListPane.swift`.
Do: D-8.
Check: `rg -n "emptyIcon" macos` prints nothing. The STEP-3 build command succeeds.
Commit: "Show the sidebar icon on the empty Source types and Metadata lists."

**STEP-7 — Trim the section-glyph test**
Files: `macos/ProvenenciaTests/WorkspaceSidebarSectionsTests.swift`.
Do: D-9.
Check: `sectionGlyphsMatchTheBoard` contains no `for` and no `switch`; `xcodebuild test … -only-testing:ProvenenciaTests/WorkspaceSidebarSectionsTests …` passes.
Commit: "Check the Conclude section marks by key instead of restating the switch."

**STEP-8 — Full gate**
Files: none.
Do: run the repo gate.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Docs, the skill, and the idea**
Files: `macos/App/DesignSystem/Recipes/Marks/MARKS.md`, `macos/App/DesignSystem/README.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `docs/ideas/archive/glyph-consolidation.md`, `docs/ideas/README.md`, `docs/deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.md`.
Do:
- `MARKS.md`: in the Families table, the `subject_*` Persistence cell becomes "Registry `IconSymbol` = `subject_<key>`; Swift maps with `PVMarkKey(subjectTypeKey:)`; `TestSwiftVocabularyMatchesGo` pins the cases". In **Glyph**, add "`PVThumbnail.Content`, the Properties type card, and the omnibar lead take a `PVGlyph` too."
- DesignSystem `README.md`: on the Thumbnail row, append after its closing parenthesis "; content takes an image or a `PVGlyph` (S10-49)".
- `add-seeded-vocabulary` `SKILL.md`: after the paragraph that ends "deferred until a later PR.", add one paragraph: "A new subject type also adds a `subject_<key>` case to `PVMarkKey`, its `provenencia_subject_<key>` asset, and its `accessibilityName` with an `L10n.DesignSystem` string. `TestSwiftVocabularyMatchesGo` fails until the case exists."
- `git mv docs/ideas/glyph-consolidation.md docs/ideas/archive/glyph-consolidation.md`. Title becomes "# Glyph consolidation (archived)"; the Status line becomes "**Done** in S10-49. The leftovers below are not scheduled."; add one `../` to each relative link. Add a section **Leftovers** listing: the omnibar lead symbols for Source types (`.library`) and Metadata fields (`.tag`), which differ from their sidebar glyphs; `PropertiesTypeChrome.sortIndex`, `showsBridgeLabel`, and `ink`, `CitationComposerAuthoring.typeRank`, and `FakeStore.syntheticPresentation`, which switch over type keys for data `CatalogSubjectTypePresentation` already carries. Remove its line from `docs/ideas/README.md`.
- S10-19 plan: change its four `../../../ideas/glyph-consolidation.md` links to `../../../ideas/archive/glyph-consolidation.md`. Change nothing else in that file.
Check: `rg -n "ideas/glyph-consolidation" docs` prints nothing (every link names `ideas/archive/`). `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.md` reports no errors.
Commit: "Document the shared subject-mark rule and archive the glyph idea with its leftovers."

**STEP-10 — Plan bookkeeping**
After the draft PR is open. Use the date the PR is opened.
Files: `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do:
- Map, `#### S10-49`: insert the landed line copied from `#### S10-19` with this PR's date and number, delete the **Plan:** line, keep the table.
- Map, PR sequence: append " — done, #N" to the `S10-49` line.
- `completed.md`: a `## S10-49 — Glyph consolidation` heading at the top and one **Landed** paragraph in the voice of the S10-19 entry.
- This plan: Status `Landed in [#N](…) (date)` and a **Landed** note listing every difference from this plan.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-49-glyph-consolidation.md` reports no errors.
Commit: "Record S10-49 as landed in pull request N."

Commits: one per behavior, as listed in each step. The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PVGlyphTests.swift` `subjectTypeKeysMapToSubjectMarks` *(new)* | The 8 type keys → the 8 `subject_*` keys; `""`, `"source_type"`, `"widget"` → `nil` | AC-2.1, AC-2.2, D-1 |
| T-2 | Go | `core/database/subjectvocab/subjectvocab_test.go` (existing `AllTypes()` loop) | Every entry: `IconSymbol == "subject_" + Key` | AC-2.3, D-2 |
| T-7 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo/subject_marks` *(new row)* | Swift `PVMarkKey` `subject_*` raw values == sorted registry `IconSymbol`s (8 today) | AC-2.3, D-2 |
| T-3 | Swift | `PVGlyphTests` `evidenceKindsAgreeWithTheSubjectRule` *(new)* | Every `EvidencePrimaryKind.allCases` and `EvidenceBridgeKind.allCases`: `markKey == PVMarkKey(subjectTypeKey: rawValue)` | D-3 |
| T-4 | Swift | `macos/ProvenenciaTests/OmnibarResultsModelTests.swift` (replace the `leadMark` line) | `leadThumbnail(for:)` with a hit of kind `event` → glyph `.mark(.subjectEvent)`, label `nil`; kind `source_type`, icon key `type_book` → `.mark(.typeBook)`, label non-nil; `source_type` with empty icon key → `.symbol(.library)`; `metadata_field` → `.symbol(.tag)`; `widget` → `.symbol(.search)` | AC-3.2, AC-3.3, AC-3.4, D-6 |
| T-5 | Swift | `WorkspaceSidebarSectionsTests` `sectionGlyphsMatchTheBoard` (trimmed) | Three literal Conclude marks | AC-4.1, D-9 |
| T-6 | Shell | STEP-2, STEP-3, STEP-4, STEP-6 `rg` checks | No `stripMarkKey`, built `"subject_\(` strings or bridge mark literals, `Content(mark:`/`(icon:`, `leadMark`/`leadSymbol`, `emptyIcon` | AC-2.4, AC-3.1 |

`PVThumbnail.Content` has no `Equatable`; T-4 compares `content.glyph` (which is `Equatable`) and `content.label != nil`.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSwiftVocabularyMatchesGo`, `OmnibarResultsModelTests`, `ThumbnailCacheTests`, `PropertiesModelTests`, `WorkspaceSidebarSectionsTests`, the `subjectvocab` Go tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected: D-11)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief; PVSymbol.grid names an existing symbol)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none (none)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-49 entry and PR sequence line; completed.md: one short paragraph of what shipped
- [ ] MARKS.md, DesignSystem/README.md, and the add-seeded-vocabulary skill updated; the idea archived with its leftovers and removed from docs/ideas/README.md; S10-19 plan links point at the archive
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The implementer runs these if it can drive the macOS app; otherwise it lists each line in the PR body as "not run — for the reviewer". Use the researcher's dogfood project and a new project.

1. **AC-1.1, AC-1.2.** In the new project, open Source types: the empty state shows the tag symbol. Open Metadata: the list symbol.
2. **AC-3.5.** In the dogfood project: the Sources list thumbnails, a ⌘K search listing a Source, a Person, an Event, and a Source type, the citation composer's artifact thumbnail and bridge rows, the Persons list, and the Properties type strip (All card and each type card) all look as they did on `main` before this PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A thumbnail changes size or ink | D-4 keeps both per case; manual check 2. |
| An unknown bridge key loses its fallback mark in the composer | D-3 keeps `?? .fallback` wherever `PVMarkKey(catalogKey:)` fell back before. |

**Stop and report if:**

- `enum PVMarkKey` does not close at column 0, so `swiftEnumBody` cannot read it.

- A registry entry's `IconSymbol` is not `"subject_" + Key` when T-2 runs.
- `rg -n -U 'PVThumbnail\.Content\(' macos` finds a caller not listed in §4.
- `rg -n "stripMarkKey|leadMark|leadSymbol|emptyIcon" macos` at the start finds a reader not listed in §4.
- A test outside §8 reads `Content.mark`, `Content.icon`, `markKey` as a `String`, or `emptyIcon`.

## 12. Open questions

None.

## 13. PR

**Title:** S10-49 — Glyph consolidation
**Body:** objective in one sentence; what shipped against US-1 to US-4; AC evidence (T-1 to T-7 and the manual check lines, each run or "not run — for the reviewer"); deviations from the plan; `Plan: docs/deployment-plan/spike-10/prs/S10-49-glyph-consolidation.md`.

## Landed

Differences from this plan: AC-2.4's `rg` still prints the four `case subject... = "subject_..."` lines of the `PVMarkKey` enum, which define the raw values. `ConclusionDetailPage.swift` built its thumbnail with `.init(mark:label:)`, a caller section 4 did not list, and moved to `.init(glyph: .mark(mark), label:)`. `PVGlyphTests.swift` uses pbxproj IDs `AA...09E0` and `AA...09E1` because `03C0` and `03C1` were taken. STEP-9's `rg "ideas/glyph-consolidation" docs` still matches this plan and the two review logs, which name the old path as history. The manual check was not run by the implementer; it is for the reviewer. The plan's Context link and STEP-9 Files line now name the idea at `docs/ideas/archive/`, because `check-plan-docs.py` rejects the old path once the file moves.

Fix round 1 (user-requested follow-ups, on this PR): `PVThumbnail`'s private `State` carries the mark or symbol it matched, so it decides its glyph once. Omnibar Source-type hits with no icon key and Metadata field hits read `WorkspaceSection.sourceTypes.glyph` and `.metadata.glyph` (tag and list), which changes the Source-type symbol from library and the Metadata symbol from tag. `PropertiesTypeChrome.sortIndex` is gone: Properties palette order reads `paletteSort` and the bridge label and muted ink read the presentation's `role`. `CitationComposerAuthoring.typeRank` reads `EvidencePrimaryKind`. `FakeStore.syntheticPresentation` stays as the Go registry's test double.
