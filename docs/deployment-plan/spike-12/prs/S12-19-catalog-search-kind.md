# S12-19 — One typed spelling of the omnibar's hit kinds

**Status:** In progress
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: no view changes what it draws)
**Board:** —
**Depends on:** —
**Blocks:** S12-15 (it adds the `evidence_graph` case this PR's pin then demands)
**Branch:** `claude/elegant-archimedes-y44jl5` (the session's assigned branch; the plan and the code ship together)
**Version bump:** none — an internal refactor with no change a researcher sees ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc` (Go wire vocabulary is spelled once in Swift and pinned by `api/ffi/swift_vocabulary_test.go`), `.cursor/rules/go-tests.mdc`
**Context:** working rule "Omnibar" (graphs are a searchable kind, S12-15); [omnibar search (Spike 3 archive)](../../archive/spike-3/omnibar-search.md), `core/search/registry.go` `Registry`
**Planned:** 2026-10-27 against `main` @ `b11d1cb`
**Review:** [`S12-19-catalog-search-kind.review.md`](S12-19-catalog-search-kind.review.md) — Fits with changes / Not ready → all fixed, 1 round

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

Search hit kinds cross the wire as strings (`SearchHit.kind`), and Swift repeats them as literals in the omnibar row, its presentation, and `FakeStore`. A renamed or added kind in `core/search` fails silently in the app: a hit takes the Source label or the generic search glyph. This PR spells the kinds once in Swift, as `CatalogSearchKind`, and pins that enum to the Go registry the way the rest of the wire vocabulary is pinned.

**Why now.** S12-15 adds `evidence_graph` as a kind. With the pin in place, that PR fails a Go test and a Swift compile until the app labels and draws the new kind.

**What stays.** The wire (`SearchHit.kind` stays a string), `CatalogSearchHit.kind` (stays `String`, so an unknown kind still decodes), `GenealogyStore.searchCatalog(kinds: [String])`, and what every row shows.

## 2. User stories

**US-1.** As a contributor, when I add or rename a kind in `core/search` `Registry`, a test names the Swift enum that disagrees, so the omnibar never shows a hit with the wrong label.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `CatalogSearchKind` cases equal the `Kind` of every `search.Registry` entry | `TestSwiftVocabularyMatchesGo/search_kinds` passes |
| AC-1.2 | A kind is in `Registry` and not in `CatalogSearchKind` (the state S12-15 starts in) | `TestSwiftVocabularyMatchesGo/search_kinds` fails, naming `CatalogSearchKind` |

**US-2.** As a researcher, the omnibar shows each hit exactly as before, and a hit of a kind this build does not know still shows.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A hit with `kind` `person`, `event`, `source_type`, `metadata_field` | The same label and lead thumbnail as today |
| AC-2.2 | A hit with `kind` `widget` (unknown) | `searchKind` is nil; label is the Source label and the lead is the search glyph, as today |
| AC-2.3 | FakeStore search for "Robins" with no kinds, and with `kinds: ["person"]` | The same hits in the same order as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CatalogSearchKind` *(new)* with one case per `Registry` kind, and `CatalogSearchHit.searchKind` | STEP-1, AC-2.2 |
| Go pin of the enum's raw values to `Registry` | STEP-2, AC-1.1, AC-1.2 |
| `CatalogSearchHitPresentation.kindLabel(for:)` and `leadThumbnail(for:)` switch on the enum | STEP-3, AC-2.1, AC-2.2 |
| `CatalogSearchHitRow.leadView` switches on the enum | STEP-3, AC-2.1 |
| `FakeStore` search (`searchCatalog`, `searchPriority`, `handleSearchHits`, `handleSearchHeader`, `vocabularySearchHits`) uses the enum | STEP-4, AC-2.3 |
| `CatalogSearchKind.init(EvidencePrimaryKind)` links the handle kinds to the primary kinds | STEP-1, STEP-4 |
| `.cursor/rules/macos-client.mdc` names the test as the index of pinned enums; the maintainability lens Q2 row and `add-searchable-kind` follow | STEP-5 |

| Out | Home |
| --- | --- |
| `evidence_graph` case | S12-15 (it adds the `Registry` entry; the pin then makes it add the case) |
| Typing `searchCatalog(kinds:)` and `CatalogSearchModel(kinds:)` as `[CatalogSearchKind]` | not planned (Promote passes an `EvidencePrimaryKind` raw value; the request side has no literal) |
| `FakeStore` delete-impact kinds (`getDeleteImpact`) and subject-type keys | not planned (different wire vocabularies) |
| Pinning `KindSpec.Priority` or `Handle` to Swift | not planned (FakeStore mirrors them as fake data) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Go kinds | `core/search/search.go` `KindSource`, `KindSourceType`, `KindMetadataField`; `core/search/registry.go` `Registry` + `init` `handle(KindPerson…)` | Six kinds: source, source_type, metadata_field, person, event, place |
| Swift hit | `macos/App/Platform/GenealogyStore.swift` `CatalogSearchHit.kind: String` | Raw wire string; doc comment lists three kinds |
| Label | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `kindLabel(for: String)` | `switch` on six literals, `default` = Source label |
| Lead | same file `leadThumbnail(for:)` | `EvidencePrimaryKind(rawValue:)` for handles, then `case "source_type"`, `case "metadata_field"`, `default` |
| Row | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` `leadView` | `case "source"` draws the Source cover |
| Fake | `macos/App/Platform/FakeStore.swift` `searchCatalog`, `searchPriority: [String: Double]`, `handleSearchHits(kinds: Set<String>)`, `handleSearchHeader(kind: String…)`, `vocabularySearchHits` | Literals for every kind; handles filtered through `primaryTypeKeys`; `CatalogSubjectMembership.kind` is a subject-type key (`String`) |
| Handles | `core/search/search.go` `KindPerson = vocabkey.Person`; Swift `EvidencePrimaryKind` (`EvidenceGraphSnapshot.swift`), `EvidencePrimaryKind.markKey` (`EvidenceCanvasInputMode.swift`) | Go defines handle kinds as the primary kinds (`TestHandleKindsArePrimaryKinds`); Swift has the list and its marks |
| Pin | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` | No search-kind row |
| Tests | `macos/ProvenenciaTests/CatalogSearchTests.swift` `CatalogSearchHitPresentationTests.matchContextLocalizesBodyFieldsAndHidesTitleRef` | Calls `kindLabel(for: "person")` and reads `leadThumbnail(…).glyph`; `CatalogSearchModelTests.kindsRestrictHits` and `SearchCatalogStoreTests.kindPriorityIsATenthAndSectionNudgeStillMatters` read FakeStore ordering |

**Precedent.** `CatalogList` in `macos/App/Features/Workspace/Session/CatalogMutation.swift`: a `String` raw-value enum with explicit raw values, pinned by the `catalog-wide lists` row with `listCase`. Copy that row's shape.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `enum CatalogSearchKind: String, CaseIterable, Sendable` in `macos/App/Platform/CatalogSearchKind.swift` *(new)*, cases in `Registry` order, each with an explicit raw value (`case sourceType = "source_type"`) | Platform owns wire vocabulary; explicit raw values let one regex (`listCase`) read every case | A case in `ConclusionVocabulary.swift` — search is not Conclusion vocabulary |
| D-2 | `CatalogSearchHit.kind` stays `String`; `extension CatalogSearchHit { var searchKind: CatalogSearchKind? }`, in `CatalogSearchKind.swift` after the enum, reads it | Tolerant decoding without an `unrecognized` case; tests and sort tie-breaks keep the raw string | `kind: CatalogSearchKind?` — loses the raw string and changes every hit literal in tests |
| D-3 | `kindLabel(for: CatalogSearchKind?)` and `leadThumbnail(for:) -> PVThumbnail.Content?` are exhaustive over the enum plus `nil`, no `default`. `leadThumbnail` returns nil for `.source` (the row draws the cover) and the search glyph for `nil`; `CatalogSearchHitRow.leadView` is `if let` on it, with no switch of its own. Handle cases read `EvidencePrimaryKind.<case>.markKey` | One switch decides every lead, and a new case fails to compile until it has a label and a lead | `default:` — S12-15 would ship with the Source label; a switch in the row too — two switches to keep in step |
| D-4 | Pin with the existing `listCase` regex against `sorted(Registry kinds…)` | Reuses the regex and the table row; registry is the source of truth | A new regex or an exported Go `Kinds()` — no second caller |
| D-5 | Handle kinds come from `EvidencePrimaryKind`, as Go's come from `vocabkey`: `CatalogSearchKind.init(_: EvidencePrimaryKind)`. FakeStore's `handleSearchKinds = Set(EvidencePrimaryKind.allCases.map { CatalogSearchKind($0) })`; `handleSearchHits(kinds: Set<CatalogSearchKind>)` decodes `EvidencePrimaryKind(rawValue: membership.kind)` (skip when nil) and switches on it exhaustively for the section and for `handleSearchHeader(kind: EvidencePrimaryKind…)` | One list of handle kinds in Swift; handle switches stay exhaustive with no dead branches | A third hand-written `[.person, .event, .place]` set; switching subject-type keys as search kinds — forces a `default` over non-handle kinds |
| D-6 | `wanted` is every case when `kinds` is empty, else the known kinds of `kinds`; an all-unknown request is an empty set and returns no hits | Matches Go `effectiveKinds` and FakeStore today | Treating an empty decoded set as "every kind" — would return everything |
| D-7 | FakeStore's `searchPriority(_: CatalogSearchKind?) -> Double` is an exhaustive switch (person 4, event 3, place 2, source 1, others and nil 0) | A new kind must state its fake priority | A dictionary keyed by the enum — silently 0 |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogSearchKind` *(new)* | type / file | `EvidencePrimaryKind` (person, event, place) | Covers three of six kinds and means subject type, not hit kind |
| `CatalogSearchHit.searchKind` *(new)* | computed property | `EvidencePrimaryKind(rawValue: hit.kind)` in `leadThumbnail` | That call becomes the switch's handle cases |
| `CatalogSearchKind.init(_:)` *(new)* | func | `PVMarkKey(subjectTypeKey:)` string round trip | Typed; the round trip and its unreachable `.fallback` go |
| `search kinds` row, `searchKinds()` *(new)* | Go test case / helper | `catalog-wide lists` row, `listStrings()` | Same shape, one more row |

**Builds on.** `swift_vocabulary_test.go` table, `listCase`, `swiftEnumBody`.
**Deletes.** Every search-kind literal in the four Swift files; `default` branches in `kindLabel`, the lead, and `handleSearchHeader`; the row's own kind switch; the `PVMarkKey(subjectTypeKey:)` round trip in the lead.
**Next adds.** S12-15 adds `case evidenceGraph = "evidence_graph"`. The pin asks for the case; the compiler then asks for its label, lead and FakeStore priority. Fake hits for it are on the `add-searchable-kind` checklist (no compiler check).

### Schema

None — no catalog change.

### Go core

None — `core/search` is read by the test only.

### FFI

None — the wire is unchanged. `api/ffi/swift_vocabulary_test.go` gains a row:

```go
{name: "search kinds", path: swiftSearchKinds, enum: "CatalogSearchKind", re: listCase, want: searchKinds()},
```

with `swiftSearchKind = "../../macos/App/Platform/CatalogSearchKind.swift"` and `searchKinds()` (above the `sortedRelations` comment) returning the sorted `Kind` of every `search.Registry` entry.

### macOS

```swift
enum CatalogSearchKind: String, CaseIterable, Sendable {
    case source = "source"
    case sourceType = "source_type"
    case metadataField = "metadata_field"
    case person = "person"
    case event = "event"
    case place = "place"

    init(_ kind: EvidencePrimaryKind)          // .person → .person, …
}

extension CatalogSearchHit {
    var searchKind: CatalogSearchKind? { CatalogSearchKind(rawValue: kind) }
}
```

The file joins the Provenencia target in `macos/Provenencia.xcodeproj/project.pbxproj` beside `ConclusionVocabulary.swift` (build file, file reference, Platform group, Sources phase).

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — The enum**
Files: `macos/App/Platform/CatalogSearchKind.swift` *(new)*, `project.pbxproj`, `GenealogyStore.swift` (doc comment on `CatalogSearchHit`).
Do: D-1, D-2.
Check: `rg -c CatalogSearchKind macos/Provenencia.xcodeproj/project.pbxproj` prints 4.

**STEP-2 — Pin it to Go**
Files: `api/ffi/swift_vocabulary_test.go`.
Do: D-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi -run TestSwiftVocabularyMatchesGo` passes; removing a case from the Swift file fails `search_kinds`.

**STEP-3 — Row and presentation**
Files: `CatalogSearchHitPresentation.swift`, `CatalogSearchHitRow.swift`, `macos/ProvenenciaTests/CatalogSearchTests.swift`. Skill: `add-swift-test`.
Do: D-3. Callers pass `hit.searchKind`. Update the existing test's calls to `.person` and `?.glyph`; its `widget` line moves into T-2, and a `source` hit asserts a nil lead.
Check: no kind literal in `kindLabel`, `leadThumbnail`, or `CatalogSearchHitRow` (the match-field codes in `showsMatchContext` and `matchContextText`, such as `"place"`, are a different vocabulary and stay).

**STEP-4 — FakeStore**
Files: `macos/App/Platform/FakeStore.swift`.
Do: D-5, D-6, D-7; hits built with `CatalogSearchKind.<case>.rawValue`.
Check: no search-kind literal in the search functions; `SearchCatalogStoreTests` unchanged and green.

**STEP-5 — Docs**
Files: `.cursor/rules/macos-client.mdc`, `.agents/skills/review-pr/lenses/maintainability.md`, `.agents/skills/add-searchable-kind/SKILL.md`, `completed.md`, `deployment-plan.md`, this plan.
Do: the rule's vocabulary line names `swift_vocabulary_test.go` as the index of pinned enums and lists `CatalogSearchKind`; the lens Q2 row says the same; the skill gets a `## Checklist` line after "FakeStore mirrors kinds" and a bullet in `### 4. FFI + Mac store`; DoD lines.
Check: links resolve; `python3 scripts/check-plan-docs.py` passes.

Commits: one per behavior, full imperative sentence ending in a period.

Swift checks (STEP-3, STEP-4, T-2–T-4, the Xcode half of `check-pr.sh`) run on macOS only. Off macOS, the evidence is the `macOS tests` workflow on the non-draft PR.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo/search_kinds` *(new)* | Swift cases vs `Registry` kinds → equal | AC-1.1, AC-1.2, D-4 |
| T-2 | Swift | `CatalogSearchTests.swift` `CatalogSearchHitPresentationTests.unknownKindKeepsTheFallbacks` *(new)* | hit kind `widget` → `searchKind` nil, Source label, search glyph | AC-2.2, D-2 |
| T-3 | Swift | `CatalogSearchTests.swift` `CatalogSearchHitPresentationTests.matchContextLocalizesBodyFieldsAndHidesTitleRef` | `.person` label; `event`, `source_type`, `metadata_field` leads → unchanged; `source` → nil lead | AC-2.1, D-3 |
| T-4 | Swift | `SearchCatalogStoreTests.swift` `kindPriorityIsATenthAndSectionNudgeStillMatters`; `CatalogSearchTests.swift` `CatalogSearchModelTests.kindsRestrictHits` | Robins ordering, `kinds: ["person"]` → unchanged | AC-2.3, D-5, D-6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `SearchCatalogStoreTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] No search-kind literal left in kindLabel, leadThumbnail, CatalogSearchHitRow, or FakeStore's search functions
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] .cursor/rules/macos-client.mdc and add-searchable-kind name CatalogSearchKind
- [ ] PR description: what shipped, AC evidence, deviations
```

## 10. Manual check

1. Open a project with a Source, a Person and a Source type. Type a query that hits each in the omnibar. Each row shows the same label and lead as on `main`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S12-15 lands first and adds `evidence_graph` | The pin fails on this branch; add the case and its label and lead from S12-15's plan, and record it in the Landed note |

**Stop and report if:**

- `search.Registry` holds a kind not listed in §4 and S12-15 has not landed.
- A test in §8 needs a changed expectation to pass (the PR changes no behavior).

## 12. Open questions

None.

## 13. PR

**Title:** S12-19 — One typed spelling of the omnibar's hit kinds
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; `Plan: docs/deployment-plan/spike-12/prs/S12-19-catalog-search-kind.md`.
