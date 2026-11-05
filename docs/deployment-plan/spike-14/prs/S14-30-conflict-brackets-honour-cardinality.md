# S14-30 — Conflict brackets honour cardinality

**Status:** Landed in #713 (2026-11-04)
**Landed:** Product version is 0.3.2 (patch on 0.3.1 at implement time). Manual check §10 not run in implementer pass.
**Spike / slice:** Spike 14 · Slice 7c — Evidence graph polish
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the ochre bracket, legend strip, and a11y conflict phrase already exist; this PR only changes when they fire)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-30-conflict-brackets-honour-cardinality`
**Version bump:** patch — researcher-visible bug fix for false conflict marks on multiple-cardinality Properties ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 7) Conflict brackets honour cardinality"; bar outcome 25; dogfood note `2026-10-26-conflict-brackets-ignore-multi-cardinality` (deleted on land); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md#8-cardinality) §8; `core/propertymeta` empty→single
**Planned:** 2026-11-04 against `main` @ `4cf944ca`
**Review:** [`S14-30-conflict-brackets-honour-cardinality.review.md`](S14-30-conflict-brackets-honour-cardinality.review.md) — Fits — Ready, 2 rounds; no open blockers; To user none

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

On an Evidence graph subject or bridge card, two or more Observations of one Property paint the ochre conflict bracket, show the canvas legend, and feed the "one of N values" accessibility phrase only when that Property's cardinality is single. Several values of a multiple Property (`languages_spoken`, `likeness`, a researcher Property set to "several values") list plainly — matching reconciliation's `StateMultiple`, not a conflict the researcher must resolve.

**Why now.** Batch 7 dogfood UX while Spike 13 continues; the marks helper (S8-06) predates `properties.cardinality` and never consulted it.

**What stays.** Bracket paint, gutter, legend chrome, negation marks, bridge sentence-key filtering, reconciliation (`StateMultiple` / `StateMixed`), Property schema and the Properties page editor UI layout, Observation proto fields. `CatalogProperty.cardinality` stays a `String` on the wire mapping; the typed enum owns comparisons.

## 2. User stories

**US-1.** As a researcher citing several values of a multiple Property on one subject card, I see no conflict bracket or legend so I am not told to resolve something that is allowed.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Card observations: two rows with `propertyKey` `languages_spoken`; `cardinalityByKey["languages_spoken"] == .multiple` | `cardHasConflict` is false; `conflictRuns` is empty; `isConflicted` is false for both rows |
| AC-1.2 | Same two `languages_spoken` rows on a placed subject; bridges empty | `graphHasConflict(subjects:bridges:cardinalityByKey:)` is false |
| AC-1.3 | Card observations: two `likeness` depiction rows; `cardinalityByKey["likeness"] == .multiple` | Same as AC-1.1 (bracket would wrap tiles today) |

**US-2.** As a researcher with competing values on a single-cardinality Property, I still see the conflict marks so real disagreements stay visible.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Two `name` rows; map entry `.single` or key absent from the map | `cardHasConflict` is true; `conflictRuns == [0..<2]`; both rows `isConflicted` |
| AC-2.2 | Two `name` (single) and two `languages_spoken` (multiple) on one card, in that order | `conflictRuns == [0..<2]` only; `languages_spoken` rows are not in a run |
| AC-2.3 | Subject with two conflicting `name` rows | `graphHasConflict` is true; legend condition remains true |

**US-3.** As a VoiceOver user (or caller of the row a11y helper), the "one of N values" phrase is omitted for multiple-cardinality stacks.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `conflictCounts(in:cardinalityByKey:)` for two `languages_spoken` with `.multiple` | Dictionary has no `languages_spoken` entry; `accessibilityLabel(…, conflictCount: nil, …)` has no `conflictOneOf` substring |
| AC-3.2 | Same helper with two `name` rows and `.single` (or absent); caller passes `conflictCount: counts["name"]` (2) | Label contains `L10n.EvidenceGraph.conflictOneOf(count: 2)` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `PropertyCardinality` *(new)* beside `PropertyValueType`, pinned in `swift_vocabulary_test.go`; every cardinality compare and assignment in Properties / FakeStore / `CatalogProperty`'s default uses its `rawValue`s | STEP-1, D-7, T-7 |
| `EvidenceCitedPropertyMarks` counts / runs / card / graph conflict APIs take `[String: PropertyCardinality]` and omit `.multiple` keys from conflict counts | STEP-2, D-1, D-2, D-3, AC-1.*, AC-2.*, AC-3.* |
| `PropertiesSnapshot.cardinalityByKey` *(new)* builds the typed lookup from `properties` | STEP-2, D-4 |
| `EvidenceCanvasCard` carries `cardinalityByKey`; `EvidenceGraphView` fills it from `fieldsHandle` for live cards and the legend | STEP-3, D-5, AC-1.2, AC-2.3 |
| `EvidenceCardView.paintRows` passes the card's map into `conflictRuns` (parameter after `projectDir`, before `painter`) | STEP-3, AC-1.1, AC-2.2 |
| Swift marks tests for multiple vs single vs mixed card; existing marks tests stay green with default empty map | STEP-4, T-1 – T-6 |
| Patch version bump (every `MARKETING_VERSION` in `project.pbxproj`) | STEP-5 |
| Docs: plan Landed, map Plan link + Status, `completed.md`; delete the dogfood note once landed | STEP-6 |

| Out | Home |
| --- | --- |
| Denormalised `property_cardinality` on `Observation` proto / `CatalogObservation` | not planned — D-1; would go stale on a Properties cardinality edit (graph rows are not on `property_vocabulary`) |
| Changing reconciliation / `StateMultiple` / `StateMixed` | already correct; map **Not this PR** |
| Changing `CatalogProperty.cardinality`'s type from `String` to `PropertyCardinality` | not planned — field stays `String`; its default and every compare/assignment use `PropertyCardinality.*.rawValue` (D-7); snapshot converts with `rawValue:` (D-4) |
| Per-row VoiceOver elements that call `accessibilityLabel` in production (helper + tests only today; `EvidenceCitedPropertyRow` was removed) | not planned — keep the helper honest for callers; no new AppKit row a11y |
| Redesigning bracket / legend chrome | not planned |
| Typing `ObservationPolarity` into `swift_vocabulary_test.go` | not planned — this PR does not touch `isNegated` |
| Inverse / place-relationship polish, Tab into text areas, bridge auto-file | S14-31 – S14-33 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Marks helper | `macos/App/Features/EvidenceGraph/EvidenceCitedPropertyMarks.swift` `conflictCounts` / `isConflicted` / `cardHasConflict` / `conflictRuns` / `graphHasConflict` / `accessibilityLabel` | Counts every non-empty `propertyKey`; ≥2 → conflict. Cardinality never read. `accessibilityLabel` appends `conflictOneOf` when `conflictCount >= 2`. `isNegated` already uses `ObservationPolarity.negative.rawValue` |
| Bracket paint | `macos/App/Features/EvidenceGraph/EvidenceCardView.swift` `paintRows` (~L438–494) | `conflictRuns(in: observations)` with no cardinality; likeness tiles paint then brackets wrap the run |
| Legend | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` ~L325 | `graphHasConflict(subjects:bridges:)` with no cardinality |
| Card model | `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift` | Holds `placed`, `text`, `state`, `nameOrders`, `projectDir` — no cardinality |
| Properties cache | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `fieldsKey` → `CatalogQueryKey.propertiesWorkspace`; `EvidenceGraphView` `fieldsHandle: QueryHandle<PropertiesSnapshot>` | Already loaded for type presentations; `PropertiesSnapshot.properties` each have `cardinality` (`CatalogProperty`, default `"single"`) |
| Snapshot type | `macos/App/Features/Properties/PropertiesSnapshot.swift` | No key→cardinality helper |
| Wire vocabulary | `macos/App/Platform/InterpretationValueKinds.swift` `PropertyValueType`; `api/ffi/swift_vocabulary_test.go` "value types" row | Value types are typed + pinned; cardinality has no Swift enum — Properties / FakeStore compare raw `"single"` / `"multiple"` strings (`PropertiesModel.setCardinality`, `PropertiesView` radios, `FakeStore.createProperty`) |
| Go constants | `core/propertymeta/propertymeta.go` `CardinalitySingle` / `CardinalityMultiple` / `CardinalityOrSingle` | Empty cardinality reads as single |
| Wire Observation | `api/proto/engine.proto` `Observation` | Has `property_key` / `property_label` / `property_value_type`; no cardinality field |
| Seeds | `core/database/subjectvocab/registry.go` `seedProperties` | Seven `CardinalityMultiple` keys: `toponym`, `interests`, `likeness`, `event_picture`, `place_picture`, `religion`, `languages_spoken` |
| Effects | `core/database/effects/registry.go` `"properties"` Lists → `ListPropertyVocabulary` (`property_vocabulary`) | A cardinality edit refreshes `propertiesWorkspace`; it does **not** reload Evidence graph rows |
| Tests | `macos/ProvenenciaTests/EvidenceCitedPropertyMarksTests.swift` | S8-06 cases: two names conflict, same string conflicts, singleton, negation, a11y, extras, `graphHasConflict`; no cardinality cases |

**Precedent.** [#S8-06 / `4c2de66a`](https://github.com/mendahu/provenencia/commit/4c2de66a) introduced `EvidenceCitedPropertyMarks` and the three call shapes (runs → paint, `graphHasConflict` → legend, `accessibilityLabel` → row phrase). Copy that helper-centric shape: keep the rule in one place, thread one lookup into every entry point. For the vocabulary enum, copy `PropertyValueType` in `InterpretationValueKinds.swift` and its `swift_vocabulary_test.go` "value types" row. Do not copy a proto denorm for metadata the session already caches.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Cardinality comes from `PropertiesSnapshot.cardinalityByKey`, not a new `Observation.property_cardinality` wire field | Map's first option; `fieldsKey` is already `propertiesWorkspace`; a Properties cardinality edit emits `property_vocabulary` and refreshes that snapshot, while Evidence graph Observation rows are **not** reloaded — a denorm on the Observation would stay stale until some other graph write | Wire field beside `property_value_type` — because it lies after a cardinality flip and needs Go + proto + regenerators for a Swift-only display rule |
| D-2 | Only `PropertyCardinality.multiple` suppresses conflict; a missing map key is conflict-eligible; empty / unknown catalog strings become `.single` when building the map (`PropertyCardinality(rawValue:) ?? .single`) | Matches `propertymeta.CardinalityOrSingle`; unknown keys stay conservative (show conflict); one typed spelling per `.cursor/rules/macos-client.mdc` | Raw `"multiple"` string compares in the marks helper — because value types already use a pinned enum and Properties already branches on the same literals; treat missing as multiple — because that would hide real conflicts when the snapshot has not loaded |
| D-3 | `conflictCounts(in:cardinalityByKey:)` omits keys whose map value is `.multiple`; `isConflicted` / `cardHasConflict` / `conflictRuns` keep reading counts as today; default `cardinalityByKey: [:]` preserves every existing call and test (all keys eligible) | One filter at the source; callers and a11y stay `counts[key]`; empty default = current behaviour | Parallel `isConflicted(…, cardinality:)` checks beside raw counts — because two rules drift; changing `isConflicted`'s meaning without filtering counts — because a11y callers that pass `counts[key]` would still say "one of 2" |
| D-4 | `PropertiesSnapshot.cardinalityByKey: [String: PropertyCardinality]` *(new)* maps each `CatalogProperty.key` through `PropertyCardinality(rawValue: cardinality) ?? .single`, with `uniquingKeysWith` preferring `.multiple` when the same key appears twice | Observation conflict grouping is by `propertyKey`; UNIQUE is `(key, origin)` so a collision is rare; preferring multiple avoids a false conflict; empty catalog strings normalize here, not at every read | Key by `propertyID` — because `conflictCounts` groups by `propertyKey` today; leave `CatalogProperty.cardinality` as `String` and convert only at this boundary — keep wire mapping unchanged (Out) |
| D-5 | `EvidenceCanvasCard.cardinalityByKey` defaults to `[:]`. `EvidenceGraphView.cards` and the legend pass `fieldsHandle.value?.cardinalityByKey ?? [:]`. Ghost cards keep the default. `paintRows` takes `cardinalityByKey` after `projectDir` and before `painter`, from `card.cardinalityByKey` | Paint and legend share one source; ghost has no cited multi-rows that matter; nil snapshot → temporary single treatment until fields load | Reach into `session` from `EvidenceCardView` — because the card view is a value-driven AppKit paint with no session |
| D-6 | No new L10n keys; no chrome redesign | Behaviour-only; existing legend / bracket / `conflictOneOf` copy stay | New "several values" legend — out of scope |
| D-7 | Add `enum PropertyCardinality: String, Sendable { case single, case multiple }` in `InterpretationValueKinds.swift` beside `PropertyValueType`. Add a `swift_vocabulary_test.go` row `"cardinality"` wanting `propertymeta.CardinalitySingle` and `CardinalityMultiple`. In `PropertiesModel.swift`, `PropertiesView.swift`, and `FakeStore.swift`, every cardinality compare and assignment uses `PropertyCardinality.single.rawValue` / `.multiple.rawValue` (including `setCardinality`, hold/create radios, locked Holds branch, `Draft.cardinality` / `openCreate` defaults, and FakeStore's empty→single / allow-list). Set `CatalogProperty.cardinality`'s default to `PropertyCardinality.single.rawValue` (field type stays `String`) | macos-client.mdc: one typed Swift enum per Go wire vocabulary; views and models use the cases; a STEP-1 `rg` that still finds quoted `"single"` / `"multiple"` in those three files fails the step | Leave raw string literals in Properties / FakeStore beside the new enum — because the locked Holds line and draft default would keep a second spelling; change `CatalogProperty.cardinality` to the enum type — Out (wire mapping stays string) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PropertyCardinality` *(new)* | enum | `PropertyValueType` in `InterpretationValueKinds.swift` | Value-type enum does not carry cardinality; same file and same vocabulary-test pattern |
| `PropertiesSnapshot.cardinalityByKey` *(new)* | computed property | `PropertiesSnapshot.boundTypes(for:)` / `binding(propertyID:typeID:)` | Those answer type bindings, not a key→cardinality map for marks |
| `cardinalityByKey` parameter on marks entry points *(new args, same functions)* | API shape | Existing `conflictCounts(in:)` | Extend the helper; do not add a second marks type |
| `EvidenceCanvasCard.cardinalityByKey` *(new stored property)* | field | `nameOrders`, `projectDir` on the same card | Same pattern: paint inputs that are not in `placed` |

**Builds on.** S8-06 `EvidenceCitedPropertyMarks`; graph already holds `PropertiesSnapshot` via `fieldsKey`; `PropertyValueType` + `swift_vocabulary_test.go` pattern; `propertymeta` cardinality constants; seeded multiple Properties in `subjectvocab`.
**Deletes.** Quoted `"single"` / `"multiple"` cardinality literals in `PropertiesModel.swift`, `PropertiesView.swift`, and `FakeStore.swift` (replaced by `PropertyCardinality.*.rawValue`). The dogfood note file is removed in STEP-6 after land (map already links the PR).
**Next adds.** A later Property-meta display rule reads `.multiple` / `.single`; S14-31–33 do not touch this helper. Boolean (S14-36) extends `PropertyValueType`, not this enum.

### Reach

None — no new navigation path. The reachable state change is "multi-row multiple Property with no bracket / no legend", which existing empty-run paint and the legend `if` already handle.

### Scale

At the map's Scale (3,000 Sources / 37,000 Subjects): `cardinalityByKey` is O(number of Properties in the catalog), built once per view read from an already-cached snapshot (tens to low hundreds of Properties, not per Observation). `conflictCounts` stays one pass over the card's observations. No new catalog write or invalidation.

### Edges

- **Assumptions.** Conflict grouping stays by `propertyKey` (unchanged). Map values are `PropertyCardinality`; missing key → conflict-eligible; `.multiple` → not. Test: missing → conflict; `.multiple` → no conflict; `.single` → conflict; catalog `""` → `.single` via D-4.
- **Neighbors.** Two identical strings on a multiple Property → still no conflict (AC-1.1). Two identical strings on single → still conflict (existing `sameStringStillConflicts`). Negative + positive on multiple → no conflict bracket; negation prefix still paints (`isNegated` unchanged). Bridge extras: `graphHasConflict` still runs `extraObservations` first, then cardinality-aware `cardHasConflict`. Two `likeness` / `event_picture` / `place_picture` rows → no bracket around the tiles (AC-1.3).
- **Removed inputs.** When cardinality flips single→multiple on the Properties page, `property_vocabulary` refreshes `propertiesWorkspace`; live graph cards drop the bracket without reloading Observations (D-1). Flip multiple→single: bracket returns for ≥2 rows.
- **Real data.** All seven seeded multiples from `core/database/subjectvocab/registry.go` `seedProperties` are conflict-suppressed when ≥2 rows share the key: `toponym`, `interests`, `likeness`, `event_picture`, `place_picture`, `religion`, `languages_spoken`. Tests pin `languages_spoken` (term) and `likeness` (depiction / tile path).
- **Untrusted size.** None — no new list from untrusted bytes; Property list is catalog-sized and already cached.

### Schema

None — no catalog change.

### Go core

None — Swift display rule only. Reconciliation already honouring cardinality is Out. The vocabulary test reads existing `propertymeta.Cardinality*` constants.

### FFI

None — no proto or handler change (D-1). Add only the `"cardinality"` row in `api/ffi/swift_vocabulary_test.go` (D-7).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Cardinality vocabulary | `PropertyCardinality` *(new)* | `macos/App/Platform/InterpretationValueKinds.swift` | Beside `PropertyValueType`; D-7 |
| Conflict rule | `EvidenceCitedPropertyMarks` | `macos/App/Features/EvidenceGraph/EvidenceCitedPropertyMarks.swift` | Filter `.multiple` in `conflictCounts`; thread map through `cardHasConflict`, `conflictRuns`, `graphHasConflict` |
| Lookup | `PropertiesSnapshot.cardinalityByKey` *(new)* | `macos/App/Features/Properties/PropertiesSnapshot.swift` | D-4 |
| Card paint input | `EvidenceCanvasCard.cardinalityByKey` | `EvidenceCanvasCard.swift` | Default `[:]`; type `[String: PropertyCardinality]` |
| Call sites | `EvidenceGraphView` legend + `cards`; `EvidenceCardView.paintRows` | same feature folder | D-5; `paintRows` param after `projectDir`, before `painter` |
| Existing compares | `PropertiesModel.setCardinality`, `PropertiesView` radios, `FakeStore.createProperty` | Properties + Platform | Switch literals to `PropertyCardinality.*.rawValue` |

States: empty snapshot → all keys conflict-eligible until load. Accessibility: production card label unchanged; `accessibilityLabel(propertyLabel:…)` stays for tests / future callers and uses filtered counts when they pass `counts[key]`.

### Strings

None — no new keys; no action-scope copy change. Existing `L10n.EvidenceGraph.legendConflict` / `conflictOneOf` keep their meaning (true single-cardinality conflicts).

## 7. Steps

**STEP-1 — `PropertyCardinality` vocabulary**
Files: `macos/App/Platform/InterpretationValueKinds.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Properties/PropertiesModel.swift`, `macos/App/Features/Properties/PropertiesView.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Platform/GenealogyStore.swift` (`CatalogProperty.cardinality` default only). Skill: —.
Do: Add `enum PropertyCardinality: String, Sendable { case single; case multiple }` beside `PropertyValueType`. Add the `"cardinality"` row to `TestSwiftVocabularyMatchesGo` with `path: swiftValueKinds`, `enum: "PropertyCardinality"`, `re: rawCase`, `want: sorted(propertymeta.CardinalitySingle, propertymeta.CardinalityMultiple)`. In the three Properties/FakeStore files, replace every cardinality compare and assignment (including `setCardinality`, hold/create radios, locked `holdsRow`, `Draft.cardinality`, `openCreate`, FakeStore empty→single and allow-list) with `PropertyCardinality.single.rawValue` / `.multiple.rawValue`. Set `CatalogProperty.cardinality`'s default to `PropertyCardinality.single.rawValue`; leave the property typed as `String`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `rg -n '"single"|"multiple"' macos/App/Features/Properties/PropertiesModel.swift macos/App/Features/Properties/PropertiesView.swift macos/App/Platform/FakeStore.swift` prints no matches.

**STEP-2 — Cardinality-aware marks + snapshot lookup**
Files: `macos/App/Features/EvidenceGraph/EvidenceCitedPropertyMarks.swift`, `macos/App/Features/Properties/PropertiesSnapshot.swift`. Skill: —.
Do: Add `PropertiesSnapshot.cardinalityByKey` per D-4. Change `conflictCounts(in:cardinalityByKey: [String: PropertyCardinality] = [:])` to skip keys whose map value is `.multiple` (D-2, D-3). Thread the same defaulted parameter through `cardHasConflict`, `conflictRuns`, and `graphHasConflict` (subjects path and bridges/`extraObservations` path). Leave `isConflicted`, `isNegated`, `extraObservations`, `sentencePropertyKeys`, and `accessibilityLabel` signatures unchanged. File comment: missing map key and non-`.multiple` values are conflict-eligible.
Check: `rg -n "func conflictCounts|func cardHasConflict|func conflictRuns|func graphHasConflict|cardinalityByKey" macos/App/Features/EvidenceGraph/EvidenceCitedPropertyMarks.swift macos/App/Features/Properties/PropertiesSnapshot.swift` shows the typed parameter and property.

**STEP-3 — Wire card + legend + paint**
Files: `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift`, `EvidenceGraphView.swift`, `EvidenceCardView.swift`. Skill: —.
Do: Add `var cardinalityByKey: [String: PropertyCardinality] = [:]` on `EvidenceCanvasCard`. In `EvidenceGraphView.cards`, pass `fieldsHandle.value?.cardinalityByKey ?? [:]` into every `EvidenceCanvasCard(…)`. Change the legend condition to `graphHasConflict(subjects:bridges:cardinalityByKey: fieldsHandle.value?.cardinalityByKey ?? [:])`. Add `cardinalityByKey: [String: PropertyCardinality]` to `paintRows` after `projectDir` and before `painter`; both subject and bridge call sites pass `card.cardinalityByKey` into `conflictRuns`. Ghost initializer unchanged (default empty map).
Check: `rg -n "conflictRuns\\(in:|graphHasConflict\\(|cardinalityByKey" macos/App/Features/EvidenceGraph/` shows production call sites passing the map (or relying on the default only on tests).

**STEP-4 — Tests**
Files: `macos/ProvenenciaTests/EvidenceCitedPropertyMarksTests.swift`. Skill: `add-swift-test`.
Do: Keep existing tests (empty map → prior behaviour). Add T-1 – T-6 below. Reuse the private `observation` helper as it is; pass explicit `[String: PropertyCardinality]` dictionaries in new tests (`.multiple` / `.single`).
Check: On macOS CI / local: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceCitedPropertyMarksTests`. On Linux agents: `rg -n "languages_spoken|likeness|cardinalityByKey|PropertyCardinality|twoLanguagesSpoken|mixedCard" macos/ProvenenciaTests/EvidenceCitedPropertyMarksTests.swift` shows the new cases; full Swift execution is macOS CI. Vocabulary row covered by STEP-1's `go test`.

**STEP-5 — Patch version**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: —; rule `versioning.mdc`.
Do: Bump patch by one from the planning-base product version. Set every `MARKETING_VERSION` line in `project.pbxproj` to that same value (four lines today: keep them equal).
Check: The three files agree; `rg -n "MARKETING_VERSION" macos/Provenencia.xcodeproj/project.pbxproj` shows one version only.

**STEP-6 — Docs**
Files: this plan (status + Landed note), `docs/deployment-plan/spike-14/deployment-plan.md` (map **Plan** link + Status), `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/ux/2026-10-26-conflict-brackets-ignore-multi-cardinality.md` (delete; map Where-the-notes row already names S14-30).
Do: per Definition of done. Map Plan cell: link text `S14-30-conflict-brackets-honour-cardinality.md`, href `prs/S14-30-conflict-brackets-honour-cardinality.md` (same shape as landed neighbours in `deployment-plan.md`).
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-30-conflict-brackets-honour-cardinality.md` still clean after status → Landed.

Commits: one per behavior. Subjects are full imperative sentences ending in a period (for example "Ignore multiple-cardinality keys when marking Evidence graph conflicts.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `EvidenceCitedPropertyMarksTests` `twoLanguagesSpokenAreNotConflicted` *(new)* | Two `languages_spoken` rows; map `[.multiple]` → `!cardHasConflict`; `conflictRuns` empty; `!isConflicted` either row; `conflictCounts` has no `languages_spoken` | AC-1.1, D-2, D-3 |
| T-2 | Swift | `…` `twoLikenessesAreNotConflicted` *(new)* | Two `likeness` rows; map `.multiple` → same as T-1 | AC-1.3, Edges Real data |
| T-3 | Swift | `…` `graphHasConflictIgnoresMultipleCardinality` *(new)* | Placed subject with two `languages_spoken`; map `.multiple` → `!graphHasConflict`; same subject with two `name` and map `.single`/absent → `graphHasConflict` | AC-1.2, AC-2.3 |
| T-4 | Swift | `…` `mixedMultipleAndSingleOnlyMarksSingle` *(new)* | Rows `[name A, name B, languages_spoken x, languages_spoken y]`; map name `.single` (or absent), languages `.multiple` → `conflictRuns == [0..<2]` | AC-2.2 |
| T-5 | Swift | `…` `missingCardinalityStillConflicts` *(new)* | Two `name` rows; `cardinalityByKey: [:]` → conflict as today | AC-2.1, D-2 |
| T-6 | Swift | `…` `accessibilityOmitsConflictForMultipleCounts` *(new)* | Build counts with `.multiple` map for `languages_spoken`; pass `conflictCount: counts["languages_spoken"]` (nil) → label has no `conflictOneOf`; control with name counts 2 → phrase present | AC-3.1, AC-3.2 |
| T-7 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` / `"cardinality"` *(new row)* | Swift `PropertyCardinality` cases == `propertymeta.CardinalitySingle` + `CardinalityMultiple` | D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceCitedPropertyMarksTests` existing cases if the default map stops treating empty as single; `PropertiesModelTests` if cardinality string compares drift from the enum raw values; `EvidenceCardLayoutTests` only if `EvidenceCanvasCard` Equatable gains a required argument without a default.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row Plan link + status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none owed (conclusion-reconciliation §8 already states the rule; graph marks now match it)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
- [ ] Dogfood note `ux/2026-10-26-conflict-brackets-ignore-multi-cardinality.md` deleted after land (map row already cites S14-30)
```

## 10. Manual check

1. Open a project; on a Person subject card in an Evidence graph, cite two `Languages spoken` values. See: no ochre bracket on those rows; canvas legend absent if that is the only multi-row key.
2. On the same or another card, cite two competing `Name` (or another single Property) values. See: ochre bracket on that run; legend "Conflict — competing values" visible.
3. On Properties, flip a researcher Property from "several values" to "one value" while the graph shows two rows of it — bracket appears after the workspace refresh without leaving the graph.

Makes slice 7c Check bar outcome 25 true for this PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Fields snapshot still nil → brief false conflicts on multiples | D-2 / D-5; `prepare` already awaits `fieldsKey`; same window as other field-dependent chrome |
| Duplicate property keys in snapshot | D-4 prefers `.multiple` |
| Call site misses the new argument | Default `[:]`; STEP-3 check `rg`; T-5 pins default behaviour |
| Properties tests expect bare `"multiple"` string in asserts | Keep store/CatalogProperty as `String` with the same raw values; update asserts only where STEP-1 changes the call site |

**Stop and report if:**

- `EvidenceCitedPropertyMarks.conflictCounts` / `conflictRuns` / `graphHasConflict` no longer match §4 (renamed or split) without a Landed note path.
- `EvidenceGraphModel.fieldsKey` is no longer `propertiesWorkspace`, or Property updates no longer emit `property_vocabulary` (D-1 evidence fails).
- Passing `cardinalityByKey` into `EvidenceCanvasCard` forces a wide unrelated Equatable / preview break outside §6.
- A production caller of `accessibilityLabel(… conflictCount:)` appears that still passes raw ≥2 counts for multiple keys — extend STEP-3 to that caller rather than inventing a second rule.
- `TestSwiftVocabularyMatchesGo` has no place for a new `rawCase` row beside `"value types"` — stop; do not invent a second vocabulary test file.

## 12. Open questions

None.

## 13. PR

**Title:** S14-30 — Conflict brackets honour cardinality
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; `Plan: docs/deployment-plan/spike-14/prs/S14-30-conflict-brackets-honour-cardinality.md`.
