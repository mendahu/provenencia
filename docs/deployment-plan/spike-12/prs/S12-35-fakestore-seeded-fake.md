# S12-35 — FakeStore is a seeded fake

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 6 — Audit fixes: Medium
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: FakeStore and fixture pins only; no view draws differently)
**Board:** —
**Depends on:** —
**Blocks:** S12-22 already landed under the same working rule; later PRs that add FakeStore methods (S12-25 and every store-method PR) follow this shape
**Branch:** `s12-35-fakestore-seeded-fake`
**Version bump:** none — test fake and Go fixture pins; nothing a researcher sees ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** Audit [A-15](../app-health-audit.md#a-15--fakestore-is-a-second-engine-with-no-parity-check); working rule "FakeStore is a seeded fake, not a second engine"; map review [K-9](../deployment-plan.review.md) (slices 5–7, option a); [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) ("Fake store keys", "Go reads every Swift mirror"); S12-22 D-14 (createArtifact linking deferred here)
**Planned:** 2026-10-28 against `main` @ `301ad014`
**Review:** [`S12-35-fakestore-seeded-fake.review.md`](S12-35-fakestore-seeded-fake.review.md) — Fits with changes / Ready → Fixed, 2 rounds

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

FakeStore today re-implements engine rules — graph list order, Citation graph membership, WriteEffects composition, and graph / Copy Source refusals — so Swift tests can pass on behavior the app does not have, with no pin to Go. This PR turns the fake back into a seeded and scripted stand-in: lists return rows in seed order, writes return the effects a test set on `nextWriteEffects`, refusals throw only through `…Error` hooks, and graph membership lives in the seeded `artifactIDs` arrays. The same PR wires `FieldSlug` into the remaining FakeStore key builders and makes Go read `FieldSlug.fixtures` and `ProjectSlug.folderNameFixtures` (decoding their escapes) the way it already reads `TextMatch.fixtures`.

**Why now.** Working rule "FakeStore is a seeded fake, not a second engine" (map review K-9, option a); audit A-15; first in slice 6 with S12-34 so later store-method PRs follow one shape.

**What stays.** Seeded dictionaries and call-recording hooks; `Error` hooks already on the fake; `unlinkImpactBlocks` and `deleteImpactByID` scripting; `insertSource` still appending a Default graph for create/copy (seed-on-create convenience; tests may also seed graphs directly); `bumpActivity` still moving a Source row's `updated` when a write touches that Source (fixture Updated cells; not effects composition); Go tests that own order, membership, effects and refusals; every production view and `GoStore`. Leftover FakeStore surface outside Ships (search-kind mirroring, `getDeleteImpact` synthesizers, vocabulary `CoreInvokeError`s, local list sorts such as `listSubjectPositions`) is unchanged and not taught as the pattern for new methods.

## 2. User stories

**US-1.** As a contributor, when I write a Swift test against FakeStore, the fake returns the rows I seeded in the order I seeded them and the effects and errors I scripted, and never invents graph order, Citation membership, WriteEffects, or a graph / Copy Source refusal on its own.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `evidenceGraphsByProject[dir] = [Smith, Jones]` (seed order), then `listEvidenceGraphs` | Returns `[Smith, Jones]`; no name sort |
| AC-1.2 | `nextWriteEffects = CatalogWriteEffects(lists: [.evidenceGraphs])`, then `createEvidenceGraph` / `deleteEvidenceGraph` / `createArtifact` | Each returns exactly that value; Source `updated` may still move via `bumpActivity` |
| AC-1.3 | `createEvidenceGraphError` set; call with empty name or empty `artifactIDs` with the error nil | Empty inputs succeed (append); with the error set, the call throws that error |
| AC-1.4 | `createArtifact(…, graphIDs: nil)` or `graphIDs: ["g2"]` with graphs already seeded | `createArtifactGraphIDs` records the argument; no `artifactIDs` array on any graph gains the new Artifact |
| AC-1.5 | `rg -n 'graphListOrder|citationIsInGraph|graphEffects|private func bumpSource|CatalogWriteEffects\(' macos/App/Platform/FakeStore.swift` | No matches (`.empty` and typed fields use no constructor call; tests assign `nextWriteEffects` from test files) |

**US-2.** As a contributor, when FakeStore mints a Property or Property-term key from a label, it calls `FieldSlug.kebab`, the same helper Source types and metadata fields already use.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `createProperty(…, label: "Custom Fact", …)` | Stored `key == "custom-fact"` via `FieldSlug.kebab` |
| AC-2.2 | `createPropertyTerm(…, label: "Grandma's album code", …)` | Stored `key == "grandmas-album-code"` |
| AC-2.3 | `rg -n 'replacingOccurrences\(of: " ", with: "-"\)' macos/App/Platform/FakeStore.swift` | No matches |

**US-3.** As a contributor, when a Swift `FieldSlug` or `ProjectSlug` fixture drifts from Go's `slug.Kebab` or `onboarding.FolderName`, a Go test fails, including fixtures that use Swift string escapes.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `FieldSlug.fixtures` row `("Grandma\u{2019}s album code", "grandmas-album-code")` | `TestSwiftFieldSlugMirrorsKebab` decodes U+2019 and matches `slug.Kebab` |
| AC-3.2 | `ProjectSlug.folderNameFixtures` rows with `\\` and `\"` in the label | `TestSwiftProjectSlugMirrorsFolderName` decodes them and matches `onboarding.FolderName` |
| AC-3.3 | A fixture pair whose decoded label's Go result differs from the fixture key | The pin test fails, naming the label |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-15: remove `graphListOrder`, `citationIsInGraph`, `graphEffects`, and effects composition in `bumpSource` | STEP-2, STEP-3, AC-1.1, AC-1.2, AC-1.5, D-1 – D-3 |
| A-15: remove empty-name / empty-`artifactIDs` refusal in `createEvidenceGraph`; Copy Source gains no blank-title refusal (script via `…Error`) | STEP-3, AC-1.3, D-4 |
| S12-22 D-14 leftover: `createArtifact` records `graphIDs` and does not mutate graph links | STEP-3, AC-1.4, D-5 |
| Fake store keys: `createProperty` and `createPropertyTerm` call `FieldSlug.kebab` | STEP-4, AC-2.*, D-6 |
| Go reads every Swift mirror: decode escapes in `readSwiftFixturePairs`; pin `FieldSlug.fixtures` and `ProjectSlug.folderNameFixtures` | STEP-1, AC-3.*, D-7 |
| Swift tests that leaned on the removed rules script order, effects, errors, and links instead | STEP-5, AC-1.*, T-4 – T-8 |
| `.cursor/rules/swift-tests.mdc` line on scripting effects and refusals; idea-file status closes the two pulled items | STEP-6, D-8 |

| Out | Home |
| --- | --- |
| Pinning `apperr` codes to `L10n.Errors` (audit L-2) | not planned (map: "Not this PR") |
| Rewriting `getDeleteImpact`'s source / artifact / subject synthesizers into pure overrides | not planned (map names graph / Copy Source refusals and the four A-15 rules; `deleteImpactByID` already scripts overrides) |
| Removing `insertSource`'s Default-graph append on create/copy | not planned (seed-on-create convenience; not order, membership-of-existing, effects resolution, or refusal) |
| Removing every other FakeStore `CoreInvokeError` (duplicate key, in_use, origin_locked, propertyterms.locked) | not planned (map: "refusal logic in its graph and Copy Source methods" only) |
| Parity golden files / routing graph writes through shared Go | not planned (K-9 option a refused parity) |
| `effects` querier onto `database.Querier` | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (import cycle; map review K-21) |
| Properties empty-search → `PVEmptyState` | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) |
| New production chrome or FFI | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Graph list order | `macos/App/Platform/FakeStore.swift` `graphListOrder`, `listEvidenceGraphs`, `firstGraphID` | Sorts by `localizedCaseInsensitiveCompare` on name, then ref, then id — not Go's `name COLLATE NOCASE, ref, id` |
| Citation membership | same file `citationIsInGraph`, used by `listCitationsByArtifact` and `citationCountsByEvidenceGraph` | Re-implements "Observation in graph, or no Observation anywhere"; counts also require the Citation's Artifact in the graph's `artifactIDs` |
| Effects | same file `graphEffects`, `bumpSource` | `graphEffects` always returns `citationSources` + `lists: [.evidenceGraphs]`; `bumpSource` invents `evidenceGraphs` / `citationSources` / `sourceScopes` / `activity` from the write |
| Graph refusals | same file `createEvidenceGraph` | Throws `StoreBoom` on empty trimmed name or empty `artifactIDs`; already has `createEvidenceGraphError` |
| Copy Source | same file `copySource` | Accepts blank title and any `artifactIDs`; no `copySourceError` hook |
| Artifact links | same file `createArtifact` | Records `createArtifactGraphIDs` and mutates each same-Source graph's `artifactIDs` (`nil` → every graph, ids → those); S12-22 D-14 said S12-35 removes this |
| Property keys | same file `createProperty`, `createPropertyTerm` | Inline `.lowercased()` / space→`-` / filter; `createSourceType` and `createMetadataField` already call `FieldSlug.kebab` |
| FieldSlug | `macos/App/Platform/FieldSlug.swift` `fixtures`, `kebab` | Fixture table includes `\u{2019}`; Swift unit test only; Go `slug.TestKebab` kept in sync by hand |
| ProjectSlug | `macos/App/Platform/ProjectSlug.swift` `folderNameFixtures`, `folderName` | Rows with `\\` and `\"`; Swift unit test only; Go `onboarding.TestFolderName` kept in sync by hand |
| Fixture reader | `api/ffi/swift_vocabulary_test.go` `readSwiftFixturePairs`, `swiftFixturePairs`, `TestSwiftFixturePairsRefusesEscapes`, `TestSwiftTextMatchMirrorsTextKey` | Refuses any backslash in a pair; `TextMatch` is the only mirror wired |
| Swift tests rule | `.cursor/rules/swift-tests.mdc` | Says FakeStore does not re-implement ordering or graph walks; does not say how to script effects and refusals |
| Store tests | `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` | `createThenListSortsByName`, `writesReportTheGraphList`, `refusalsThrowForUnknownGraphsAndEmptyInput`, `createArtifactLinksEveryGraphWhenNoneChosen`, `createSourceMakesADefaultGraph` lean on the rules above |
| Apply-effects tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` `cardSaveKeepsAGraphMadeDuringEdit`, `cardSaveSkipsAGraphDeletedDuringEdit` | `session.apply` the effects returned by create/delete graph |

**Precedent.** S11-08 ([#466](https://github.com/mendahu/provenencia/pull/466)) wired `TextMatch.fixtures` through `swiftFixturePairs` and left FieldSlug / ProjectSlug for this PR; copy that pin-test shape and extend the reader to decode escapes instead of refusing them. S12-22 D-14's `unlinkImpactBlocks` is the scripting shape for effects: one seeded value the method returns, no engine composition. `seededEventHeaders` / `eventTitlesByGraph` are the "seeded rows, Go owns the rule" precedent already in FakeStore's header comment.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Delete `graphListOrder`. `listEvidenceGraphs` returns `evidenceGraphsByProject[projectDir] ?? []` unchanged. `firstGraphID(ofArtifact:)` iterates `evidenceGraphsByProject` values in any order across projects; within each project's array, the first graph whose `artifactIDs` contains the Artifact wins; no name/ref/id sort | Working rule: seeded order; A-15 names `graphListOrder`; tests are single-project | Keeping a "stable by id" sort — still an ordering rule Ships removes |
| D-2 | Delete `citationIsInGraph`. `listCitationsByArtifact` returns `citationsByID.values.filter { $0.artifactID == artifactID }` with no membership rule and no `.sorted` on ref. `citationCountsByEvidenceGraph` keeps only the seeded `graph.artifactIDs` filter (`artifactIDs.contains(citation.artifactID)`) and drops the `citationIsInGraph` conjunct. A test that needs a graph-scoped Citation set seeds only those Citations (or filters in the test) | Working rule: no graph membership in the fake; A-15; both callers of the helper | A seeded `citationsByGraph` map — second membership mechanism; keeping the ref sort — ordering rule on the method this PR rewrites |
| D-3 | Add `var nextWriteEffects: CatalogWriteEffects = .empty` on FakeStore. Every method that returns `CatalogWriteEffects` returns `nextWriteEffects` after its state mutation (sticky until a test assigns another value), including sites that today invent `CatalogWriteEffects(…)` without calling `bumpSource` / `graphEffects` (`unpromoteSubject`, `setPreferredLikeness`, `createCitedBridge`, `createCitation`, `addObservationsToCitation`, `updateCitation`, `updateObservation`, `deleteObservation`, `deleteCitation`, and any other return). Delete `graphEffects`. Rename `bumpSource` → `bumpActivity(_ sourceID: String?)` *(private)*: updates that Source's `updated` / `nextAuditRevision` when the Source is in a project list, returns `Void`. Every current `bumpSource(sourceID, graphs:)` call site becomes `bumpActivity(sourceID)` with the same `sourceID`; paths that only called `graphEffects` get no `bumpActivity`. No caller encodes graphs into effects | Working rule: scripted effects; A-15 effects resolution; matches `unlinkImpactBlocks`; Updated-cell fixture convenience stays in What stays | A per-method effects queue; keeping inventing `CatalogWriteEffects(…)` returns beside `nextWriteEffects`; "call bumpActivity only when you still want Updated" — implementer judgment |
| D-4 | Remove the empty-name and empty-`artifactIDs` `guard`s from `createEvidenceGraph` (trimmed empty name still stores the trimmed string; empty `artifactIDs` still appends). Keep `createEvidenceGraphError`. Add `var copySourceError: Error?`; `copySource` throws it when set, otherwise accepts blank titles and whatever `artifactIDs` the test passed (no new refusal). Unknown-id `StoreBoom` on missing seeded rows stays (fixture hole, not a domain refusal) | Map: script refusals through an `Error` hook; A-15 Copy Source "never refuse" stays (do not add blank-title refusal) | Adding Go-parity refusals — K-9 option a refused parity |
| D-5 | `createArtifact` appends the Artifact to `artifactsBySource` and to `createArtifactGraphIDs`, calls `bumpActivity(sourceID)`, returns `nextWriteEffects`. It does not modify any graph's `artifactIDs`. Tests that need a linked Artifact seed that graph's `artifactIDs` (or seed after create) | S12-22 D-14; working rule: no graph membership | Keeping `nil` → every graph as "wire default" — that is the engine rule |
| D-6 | `createProperty` and `createPropertyTerm` set `key = FieldSlug.kebab(label)`; when `key.isEmpty`, throw `StoreBoom.boom` (same as `createSourceType`). Delete the two inline slug pipelines | Idea "Fake store keys"; one helper | Leaving term creates on the inline copy |
| D-7 | In `api/ffi/swift_vocabulary_test.go` (same file as `readSwiftFixturePairs`), add unexported `decodeSwiftStringLiteral`. Extend `readSwiftFixturePairs` to decode Swift string-literal escapes in each capture: `\\`, `\"`, `\'`, `\n`, `\r`, `\t`, `\0`, and `\u{n}` (1–8 hex digits → UTF-8 rune). Unknown escape → error. No `\xHH`. Rename `TestSwiftFixturePairsRefusesEscapes` → `TestSwiftFixturePairsDecodesEscapes` *(new cases for decode; keep an unknown-escape error row)*. Add consts `swiftFieldSlug` and `swiftProjectSlug`. Add `TestSwiftFieldSlugMirrorsKebab` *(new)*: each `FieldSlug.fixtures` pair equals `slug.Kebab(label)`. Add `TestSwiftProjectSlugMirrorsFolderName` *(new)*: each `ProjectSlug.folderNameFixtures` pair equals `onboarding.FolderName(label)` (error → treat as `""` for the compare, matching Swift's empty return). Leave `slug.TestKebab` and `onboarding.TestFolderName` tables in place (they may have extra rows) | Idea "Go reads every Swift mirror"; S11-08 left this Out; macos-client.mdc already names the pattern | Rewriting fixtures to avoid escapes — ProjectSlug must express `\` and `"` in labels; deleting the Go tables — they cover invalid/empty rows the Swift fixtures omit; inventing `\xHH` — not a Swift string escape |
| D-8 | Rewrite the FakeStore paragraph in `.cursor/rules/swift-tests.mdc` so it matches this PR's In + What stays: FakeStore seeds rows and scripts `nextWriteEffects`, matching `Error` hooks, `unlinkImpactBlocks`, and `deleteImpactByID`; it does not re-implement graph list order, Citation graph membership, WriteEffects composition, or graph / Copy Source refusals (those live in Go). New FakeStore methods follow that scripting shape and add none of those rules. Do not claim every leftover synthesizer or local sort is gone. Add one bullet to `.agents/skills/add-swift-test/SKILL.md` under FakeStore: new methods seed/script (`nextWriteEffects`, `Error` hooks, impact maps); no order, membership, effects-resolution, or graph/Copy Source refusal logic. Update `docs/ideas/shared-list-helpers.md` status line so "Fake store keys" and "Go reads every Swift mirror" read as closed by S12-35 (remove those names from the "went to" clause; no resurrected `##` sections). STEP-6 rewrites the map S12-35 **Check** to the Ships/AC-1.5 surface, and rewrites the **standing** working-rule paragraph titled FakeStore is a seeded fake (not only its S12-35 clause) to: FakeStore returns seeded rows and scripted effects/errors; S12-35 removes graph list order, Citation membership, WriteEffects composition, and graph/Copy Source refusals; a store method a PR adds follows that scripting shape and adds none of those rules; leftover synthesizers, vocabulary `CoreInvokeError`s, and local sorts outside Ships are not the pattern new methods copy; engine rules stay tested in Go | A-6 / S12-20 Out named this rewrite; map Check and standing working rule must match Ships + What stays; closes the idea pull | Clarifying only the S12-35 clause while the standing "no ordering or refusal logic" sentence stays false; a one-line append that leaves the false rule claim intact |

**Rules this plan bends.** None — STEP-6 rewrites the standing working rule and the map Check to match Ships + What stays (K-9 option a); Out leftovers stay outside the pattern new methods copy.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `FakeStore.nextWriteEffects` | field | `unlinkImpactBlocks`, `promoteEvidenceError`, `createEvidenceGraphError` | Effects need a return value every write can carry; Error hooks only cover throws |
| `FakeStore.copySourceError` | field | `createEvidenceGraphError`, `updateSourceError` | Copy Source had no script hook; A-15 names its refusals |
| `FakeStore.bumpActivity` | func | today's `bumpSource` without the return | Same activity mutation, no effects composition |
| `decodeSwiftStringLiteral` *(new, unexported)* in `api/ffi/swift_vocabulary_test.go` | helper | `readSwiftFixturePairs` in the same file | Escapes must decode; TextMatch fixtures stay plain |
| `TestSwiftFieldSlugMirrorsKebab`, `TestSwiftProjectSlugMirrorsFolderName` | tests | `TestSwiftTextMatchMirrorsTextKey` | Two more rows in the same index |

Searched: `rg "graphListOrder|citationIsInGraph|graphEffects|bumpSource|FieldSlug|swiftFixturePairs|nextWriteEffects|folderNameFixtures|citationCountsByEvidenceGraph" macos/ api/ffi/ docs/ .cursor/rules/ .agents/`.

**Builds on.** Working rule "FakeStore is a seeded fake"; `swiftFixturePairs` / `TestSwiftTextMatchMirrorsTextKey`; `FieldSlug.kebab` already used by `createSourceType` / `createMetadataField`; S12-22's `unlinkImpactBlocks` scripting.
**Deletes.** `graphListOrder`, `citationIsInGraph`, `graphEffects`, `bumpSource` (replaced by `bumpActivity` + `nextWriteEffects`), inventing `return CatalogWriteEffects(…)` sites (replaced by `return nextWriteEffects`), the empty-name / empty-`artifactIDs` guards in `createEvidenceGraph`, the `createArtifact` link-mutation loop, the `listCitationsByArtifact` ref sort, the two inline slug pipelines in `createProperty` / `createPropertyTerm`, `TestSwiftFixturePairsRefusesEscapes` (replaced by the decode test).
**Next adds.** A PR that adds a FakeStore method scripts `nextWriteEffects` / an `Error` hook and seeds rows; it adds no graph-list order, Citation membership, effects composition, or graph/Copy Source refusal (binds S12-25 and every later store-method PR).

### Reach

None — no new path into existing production code; test fake and Go pins only.

### Scale

None — no production write, list or loop changes.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `api/ffi` | `decodeSwiftStringLiteral` *(new)*; `readSwiftFixturePairs` decodes each capture; `TestSwiftFixturePairsDecodesEscapes` *(renamed)*; `TestSwiftFieldSlugMirrorsKebab` *(new)*; `TestSwiftProjectSlugMirrorsFolderName` *(new)*; consts `swiftFieldSlug`, `swiftProjectSlug` |

Errors: none new on the product path.

### FFI

None — no proto or handler change. The pin tests live in `api/ffi` beside the other Swift-mirror tests.

### macOS

None — no view, session, or kit change. FakeStore-only under `#if DEBUG`.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| — | — | `macos/App/Platform/FakeStore.swift` | `nextWriteEffects`, `copySourceError`, `bumpActivity`; deletions per Deletes |

States: — (no UI). Accessibility: —.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Decode fixture escapes; pin FieldSlug and ProjectSlug**
Files: `api/ffi/swift_vocabulary_test.go`. Skill: — (Go tests: `.cursor/rules/go-tests.mdc`).
Do: D-7. Implement `decodeSwiftStringLiteral` in this file. Switch `readSwiftFixturePairs` to decode. Replace `TestSwiftFixturePairsRefusesEscapes` with `TestSwiftFixturePairsDecodesEscapes` (rows: plain pair; `\u{2019}`; `\\` and `\"`; unknown escape → error). Add the two pin tests and path consts.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run 'TestSwiftFixturePairs|TestSwiftFieldSlug|TestSwiftProjectSlug|TestSwiftTextMatch'`.

**STEP-2 — Seeded graph order; drop Citation membership**
Files: `macos/App/Platform/FakeStore.swift`. Skill: —.
Do: D-1, D-2. Delete `graphListOrder` and `citationIsInGraph`. `listEvidenceGraphs` returns the seeded array. `firstGraphID` per D-1. `listCitationsByArtifact` filters by `artifactID` only and drops its ref `.sorted`. `citationCountsByEvidenceGraph` keeps only the `artifactIDs.contains` filter. Reword the `evidenceGraphsByProject` field comment: create/copy seed a Default graph for tests; tests may also seed graphs directly (do not say "as the engine does").
Check: `rg -n 'graphListOrder|citationIsInGraph|localizedCaseInsensitiveCompare' macos/App/Platform/FakeStore.swift` prints nothing under `listCitationsByArtifact` / `listEvidenceGraphs` / `firstGraphID` (other methods' local sorts outside Ships may remain).

**STEP-3 — Scripted effects; drop graph/copy refusals and createArtifact linking**
Files: `macos/App/Platform/FakeStore.swift`. Skill: —.
Do: D-3, D-4, D-5. Add `nextWriteEffects` and `copySourceError`. Every `bumpSource(sourceID, graphs:)` call site becomes `bumpActivity(sourceID)`. Every method that returns `CatalogWriteEffects` returns `nextWriteEffects` (or `(value, nextWriteEffects)`), including inventing `CatalogWriteEffects(…)` sites named in D-3. Delete `graphEffects` and `bumpSource`. Remove empty-name / empty-`artifactIDs` guards. Wire `copySourceError`. Stop `createArtifact`'s link loop.
Check: `rg -n 'graphEffects|private func bumpSource|artifactIDs\.append\(art\.id\)' macos/App/Platform/FakeStore.swift` prints nothing; `rg -n 'CatalogWriteEffects\(' macos/App/Platform/FakeStore.swift` prints nothing (covers bare returns and tuple/`(…, CatalogWriteEffects(…))` inventing sites); `rg -n 'nextWriteEffects|bumpActivity|copySourceError' macos/App/Platform/FakeStore.swift` hits the new symbols.

**STEP-4 — FakeStore keys through FieldSlug**
Files: `macos/App/Platform/FakeStore.swift`. Skill: —.
Do: D-6.
Check: `rg -n 'replacingOccurrences\(of: " ", with: "-"\)' macos/App/Platform/FakeStore.swift` prints nothing; `rg -n 'FieldSlug\.kebab' macos/App/Platform/FakeStore.swift` hits `createSourceType`, `createMetadataField`, `createProperty`, `createPropertyTerm`.

**STEP-5 — Rewrite Swift tests that leaned on the rules**
Files: `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift` (`cardSaveKeepsAGraphMadeDuringEdit`, `cardSaveSkipsAGraphDeletedDuringEdit`). Skill: `add-swift-test`.
Do: Apply T-4 – T-10. For every other ProvenenciaTests failure that asserts FakeStore-invented order, effects, links, or empty-input refusal, script `nextWriteEffects` / an `Error` hook / seeded `artifactIDs` the same way — do not restore a deleted rule. Do not edit production views.
Check: On a Mac, `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` (CI: `.github/workflows/macos-test.yml`). Go pins from STEP-1 pass; AC-1.5 / AC-2.3 `rg` checks pass.

**STEP-6 — Docs**
Files: `.cursor/rules/swift-tests.mdc`, `.agents/skills/add-swift-test/SKILL.md`, `docs/ideas/shared-list-helpers.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row Status, S12-35 **Check**, working rule "FakeStore is a seeded fake" S12-35 sentence), this plan (status).
Do: D-8; Definition of done docs lines. Map Check becomes: `rg` finds none of `graphListOrder`, `citationIsInGraph`, `graphEffects`, `bumpSource`, or `CatalogWriteEffects(` in `FakeStore.swift`; Swift tests pass with scripted values; fixture pins fail on drift. Standing working-rule paragraph rewritten per D-8 (Ships removals + new-method scripting shape; leftovers not the pattern).
Check: links resolve; `python3 scripts/check-guidance-names.py` passes; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-35-fakestore-seeded-fake.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher (or contributor), ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftFixturePairsDecodesEscapes` *(renamed)* | `\u{2019}` → U+2019; `\\` → `\`; `\"` → `"`; unknown escape → error; plain pairs unchanged | D-7, AC-3.1 |
| T-2 | Go | same file `TestSwiftFieldSlugMirrorsKebab` *(new)* | Each `FieldSlug.fixtures` pair → `slug.Kebab(label) == key` | AC-3.1, AC-3.3, D-7 |
| T-3 | Go | same file `TestSwiftProjectSlugMirrorsFolderName` *(new)* | Each `folderNameFixtures` pair → `FolderName(label)` matches folder (`""` when Go errors) | AC-3.2, AC-3.3, D-7 |
| T-4 | Swift | `EvidenceGraphStoreTests` `listReturnsSeedOrder` *(renamed from `createThenListSortsByName`)* | Seed `[Smith, Jones]` via two creates (append order) or direct seed → list names `["Smith", "Jones"]` | AC-1.1, D-1 |
| T-5 | Swift | `EvidenceGraphStoreTests` `writesReturnScriptedEffects` *(renamed from `writesReportTheGraphList`)* | Set `nextWriteEffects = CatalogWriteEffects(lists: [.evidenceGraphs])` → create/update/delete each return that value | AC-1.2, D-3 |
| T-6 | Swift | `EvidenceGraphStoreTests` `createEvidenceGraphErrorIsScripted` *(replaces `refusalsThrowForUnknownGraphsAndEmptyInput`)* | `createEvidenceGraphError = Boom()` → create throws; with error nil and empty `artifactIDs`, create succeeds; `deleteEvidenceGraph(graphID: "nope")` still throws `StoreBoom` (unknown id) | AC-1.3, D-4 |
| T-7 | Swift | `EvidenceGraphStoreTests` `createArtifactRecordsGraphIDsWithoutLinking` *(rewrites `createArtifactLinksEveryGraphWhenNoneChosen`, `createArtifactRecordsItsGraphIDs`, and `createArtifactLinksOnlyItsProjectsGraphs`)* | Seed graphs (this project and another) with fixed `artifactIDs`; create with `nil` / `["g2"]` / `[]` → `createArtifactGraphIDs` matches; every graph's `artifactIDs` unchanged | AC-1.4, D-5 |
| T-8 | Swift | `SourcePageModelTests` `cardSaveKeepsAGraphMadeDuringEdit`, `cardSaveSkipsAGraphDeletedDuringEdit` | Before create/delete graph, set `nextWriteEffects = CatalogWriteEffects(lists: [.evidenceGraphs])` so `session.apply` reloads the list | AC-1.2, D-3 |
| T-9 | Swift | `EvidenceGraphStoreTests` `createSourceMakesADefaultGraph` *(trimmed)* | createSource still seeds a Default graph; drop assertions that FakeStore composed `evidenceGraphs` effects; createArtifact no longer auto-links — seed or assert `createArtifactGraphIDs` only; copySource still seeds a Default with the copied Artifact ids on that new graph via `insertSource` | D-5, Out (insertSource kept) |
| T-10 | Swift | `EvidenceGraphStoreTests` `createPropertyAndTermKeysUseFieldSlug` *(new)* | Seed one property id `p1`; `createProperty(projectDir:userID:label: "Custom Fact", valueType: "text", description: "", cardinality: "")` → key `custom-fact`; `createPropertyTerm(projectDir:userID:propertyID: "p1", label: "Grandma's album code", description: "")` → key `grandmas-album-code` | AC-2.1, AC-2.2, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Suites most likely to break: `EvidenceGraphStoreTests`, `SourcePageModelTests`, `PropertiesModelTests` (key minting), any test that applied FakeStore-invented effects without scripting.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`graphListOrder`, `citationIsInGraph`, `graphEffects`, `bumpSource` as the effects-returning helper); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: swift-tests.mdc FakeStore paragraph, add-swift-test FakeStore bullet, idea status line, map Check and working-rule S12-35 sentence
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. No researcher-facing UI change. Slice Check (AC-1.5): `rg -n 'graphListOrder|citationIsInGraph|func graphEffects|private func bumpSource|CatalogWriteEffects\(' macos/App/Platform/FakeStore.swift` prints nothing.
2. AC-2.3: `rg -n 'replacingOccurrences\(of: " ", with: "-"\)' macos/App/Platform/FakeStore.swift` prints nothing.
3. Swift tests pass with scripted `nextWriteEffects` / `Error` hook values (T-4 – T-10). Go pin tests fail when a Swift fixture drifts (T-2, T-3).
4. Optional spot-check on a Mac: ProvenenciaTests for `EvidenceGraphStoreTests` and `SourcePageModelTests` after STEP-5.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Many call sites of `bumpSource` miss the `nextWriteEffects` cutover | STEP-3 `rg` for `bumpSource` / `graphEffects` must be empty before STEP-5 |
| A UI test applied invented effects and now sees a stale cache | T-8 pattern: set `nextWriteEffects` before the write; STEP-5 covers every failure the same way |
| Escape decoder mishandles a fixture and the pin silently skips a row | Decoder errors on unknown escapes; T-1 covers `\u{…}`, `\\`, `\"` |
| Implementer expands into deleting all FakeStore `CoreInvokeError`s | Out line; Stop below |

**Stop and report if:**

- `FakeStore.swift` no longer contains `graphListOrder`, `citationIsInGraph`, `graphEffects`, or `bumpSource` at the planning base (already removed by another PR) — re-verify Current state and shrink Steps.
- A ProvenenciaTests failure cannot pass without restoring order, membership, effects composition, or a graph/Copy Source refusal (Out).
- `readSwiftFixturePairs` decode of an existing `TextMatch.fixtures` row changes its bytes (TextMatch must stay green).
- `insertSource`'s Default-graph append is required to be removed to satisfy a reviewer of this PR — that is Out; report and do not delete it here.

## 12. Open questions

None.

## 13. PR

**Title:** S12-35 — FakeStore is a seeded fake
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-35-fakestore-seeded-fake.md`.
