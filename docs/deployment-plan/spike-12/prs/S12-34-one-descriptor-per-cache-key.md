# S12-34 — One descriptor per cache key

**Status:** Landed in [#614](https://github.com/mendahu/provenencia/pull/614) (2026-10-28)
**Landed:** No difference from the plan.
**Spike / slice:** Spike 12 · Slice 6 — Audit fixes: Medium
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: session and registry structure only; no view draws differently)
**Board:** —
**Depends on:** —
**Blocks:** S12-25, S12-26 (their progress, citation, graph-list and Sources-list updates are entries in this PR's `patch` and `refresh` slots); S12-32 and S12-33 shrink its invalidation sets after it
**Branch:** `s12-34-one-descriptor-per-cache-key`
**Version bump:** none — internal cache-registry reshape with no change a researcher sees ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** Audit [A-30](../app-health-audit.md#a-30--a-cache-key-is-eight-hand-kept-lists); working rule "The cache-key descriptor comes first in slice 6"; map review [K-1](../deployment-plan.review.md); maintainability Q2; [`docs/macos-client-patterns.md`](../../../macos-client-patterns.md) § Workspace session
**Planned:** 2026-10-28 against `main` @ `aae1f6b7`
**Review:** [`S12-34-one-descriptor-per-cache-key.review.md`](S12-34-one-descriptor-per-cache-key.review.md) — Fits with changes / Ready → Fixed, 2 rounds

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

Each catalog cache key today is restated across eight hand-kept lists: the key cases, `Kind`, the `kind` and `project` mappings, the registry `specs` table, `load → Any`, `Kind.invalidation`, and `warmQuery`'s typed switch. A missing `Spec` silently falls back to `.sessionFresh` with no invalidation; a loader/`warmQuery` type drift fails only at runtime. This PR gives every `CatalogQueryKey.Kind` one typed descriptor (value, loader, stale/hidden/invalidateOn, invalidation builder, `patches`, `refresh`) built from an exhaustive `Kind` switch, deletes the `Any` load and the `warmQuery` switch, and pins completeness with a test. Behavior for the researcher is unchanged.

**Why now.** Working rule "The cache-key descriptor comes first in slice 6" (map review K-1): S12-25 and S12-26 must add their partial reads as `patch` / `refresh` entries, not new session functions. Audit A-30.

**What stays.** Every key's loader result, stale and hidden policies, `invalidateOn` sets, last-updated and Source-row patch results, progress Get-one + map merge, `CatalogMutation` / `CatalogMutationKind`, feature call sites of `session.query` / `apply` (including the call-site `QueryHandle<T>` annotations), the effects-driven `invalidations(for effects:)` fan-out body (S13-01 / S12-25 add their own id selectors on new invalidation / refresh entries), and what every screen shows.

## 2. User stories

**US-1.** As a contributor, when I add a `CatalogQueryKey.Kind`, the descriptor switch fails to compile until that kind has an arm, and a test fails if `allCases` and the switch disagree, so a new key cannot ship with a silent `.sessionFresh` fallback or a missing warm path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `CatalogQueryKey.Kind.allCases` (19 kinds at the planning base) | Each kind's `descriptor` exists; `everyKindHasOneDescriptor` passes |
| AC-1.2 | The completeness test builds the set of kinds that have a descriptor arm and compares it to `allCases` after one kind is omitted from that observed set (plant by filtering in the test; do not edit production code) | `everyKindHasOneDescriptor` fails, naming the missing kind |
| AC-1.3 | `stalePolicy(for:)`, `hiddenPolicy(_:)`, and the patches accessor for a known key | Read `key.kind.descriptor`; missing kind is unreachable (`preconditionFailure`); no `?? .sessionFresh` / `?? .revalidate` |

**US-2.** As a contributor, when the session warms or loads a key, the loader and warm path come from that kind's descriptor, so the old `load → Any` API and the `warmQuery` type switch are gone and cannot disagree with each other.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `session.query` / `apply(location:)` for each of the 19 kinds | Same `QueryHandle` value type as today; `CatalogQueryRegistryTests` and `WorkspaceSessionTests` pass |
| AC-2.2 | `rg` for `func load\(key: CatalogQueryKey, store:` and for `private func warmQuery` under `macos/App/Features/Workspace/Session/` | No hits |
| AC-2.3 | `CatalogQueryKey.Kind` | Conforms to `CaseIterable` |

**US-3.** As a contributor, when a write reports activity, an enriched Source, or graph progress to refresh, the session runs each warmed key's descriptor `patches` / `refresh` with ids from that slot's own `follows` — the same merges as today — so S12-25 and S12-26 add entries instead of new private session functions.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A Source-page write with `WriteEffects.activity` on a warmed Sources list | The matching row's `updated` moves in place; no list reload (same as `applySourceWriteEffectsPatchesTheListRowWithoutReloading`) |
| AC-3.2 | `session.apply(.updatedSource(source))` with a warmed list and workspace | List row and workspace shell match `source`; workspace stale flag cleared if patched (same as `applyUpdatedSourcePatchesListAndWorkspace`) |
| AC-3.3 | A graph write that reports `evidenceGraphs` (or a mutation with `evidenceGraphId`) while the progress map is warmed | That graph's progress row is Get-one merged (or removed when `isZero`); other graphs untouched (same as `applyGraphEffectsPatchProgressOnly`) |
| AC-3.4 | The same progress write while the progress map was never warmed | No Get-one call (skip, as today) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-30: one typed descriptor per `CatalogQueryKey.Kind` (value, loader, stalePolicy, whenHidden, invalidateOn, invalidation builder, `patches`, `refresh`) | STEP-1 – STEP-4, AC-1.*, AC-2.*, US-3 |
| `Kind: CaseIterable`; descriptor built by exhaustive `Kind` switch; `everyKindHasOneDescriptor` | STEP-1, STEP-5, AC-1.1, AC-1.2, AC-2.3 |
| Generalize last-updated scope into `CatalogQueryFollows` on each patch and on refresh; per-key `refresh(ids)` as the map names | STEP-3, STEP-4, AC-3.*, D-4, D-5, D-6 |
| Move `lastUpdated` + `patchUpdatedSource` into descriptor `patches`; move `patchEvidenceGraphProgress` into `refresh` on `.evidenceGraphProgress` | STEP-3, STEP-4, AC-3.* |
| Delete `load(key:store:) -> Any`, `warmQuery`, private `Spec` / `specs`, `CatalogQueryRegistryError`, session-private patch/refresh helpers, silent missing-kind fallbacks | STEP-2, STEP-4, AC-2.2, AC-1.3 |
| Update `add-workspace-place` (step 3, step 6 lastUpdated line, cache-strategy table); fix false lines in `macos-client-patterns.md` and spike-13 map prose that still say registry `Spec` for cache keys; Session doc comments | STEP-6 |

| Out | Home |
| --- | --- |
| Behavior change to any key's loader, invalidation set, hidden policy, or patch result | not planned (map: "Not this PR. A behavior change to any key.") |
| Moving `invalidations(for effects:)` fan-out onto descriptor rows (touchedHandles → conclusionDetail, citationSources, lists, structure) | S13-01 (adds a per-handle invalidation entry with its own id selector; working rule there), S12-25 / S12-26 (reads / graph-list). This PR's `CatalogQueryFollows` on patch/refresh is the pattern those entries copy; effects invalidation stays in `invalidations(for:)` until those PRs |
| Citation-read / progress `reads` effect and Swift invalidation narrowing | S12-25 |
| Graph list by Source, Sources-list insert, `QueryHandle.isStale`, hidden `.evidenceGraph` / `.sourceWorkspace` | S12-26 |
| Redesign or deletion of `CatalogMutation` / `CatalogMutationKind` / `invalidationKind` | S12-32, S12-33 (A-29) |
| Typing call-site `session.query` / `queryHandle` annotations so `Value` is compile-checked against the descriptor | not planned (log-only; US-2 covers the loader/warm path only) |
| FakeStore seeding rules | S12-35 |
| New Swift file / `project.pbxproj` entry | not planned (descriptor stays in `CatalogQueryRegistry.swift` / `CatalogLastUpdated.swift` as needed; S12-31 owns synchronized folders) |
| Changing what `CitationComposerContext.applyWrite` does after `apply` | S12-25 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Keys | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `CatalogQueryKey`, nested `Kind` | 19 key cases; `Kind` is `Hashable, Sendable` and **not** `CaseIterable`; `kind` and `project` are exhaustive switches; no payload accessors |
| Specs | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` private `Spec`, `specs: [Spec]` | One row per kind: `stalePolicy`, `invalidateOn`, `whenHidden` (default `.revalidate`), optional `lastUpdated`. Lookups use `specs.first { $0.kind == … }?… ?? fallback` |
| Load | same file `load(key:store:) async throws -> Any` | Exhaustive `switch key`; returns concrete values erased to `Any`; throws nothing typed except through the session cast |
| Type mismatch | same file `CatalogQueryRegistryError.typeMismatch` | Only thrown from `WorkspaceSession.query`'s `as? Value` cast |
| Invalidation | same file `invalidations(for:)`, `invalidations(by:)`, private `Kind.invalidation(project:mutation:)` | Specs filter by `invalidateOn`; id-bearing kinds use mutation payload or `.allCached`; effects fan-out is hand-written (lines 310–338) |
| Typed warm | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `query`, `warmQuery` | `query` casts `Any as? Value` or throws `typeMismatch`; `warmQuery` restates all 19 types as `let _: QueryHandle<…> = query(key)` |
| Apply order | same file `apply(_:_:)` | (1) `patchLastUpdated` (2) `patchUpdatedSource` if `mutation?.patchedSource` (3) invalidations (4) `refreshEvidenceGraphProgress` per graph id (5) optional nav recount via `navCountTriggers` / `onNavCountsMoved` |
| Last-updated | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogLastUpdatedPatch.rows`, `CatalogLastUpdatedScope` | Type-erased closure over `Row`; `.handle` / `.source` scopes; declared on four list specs |
| Source patch | `WorkspaceSession.swift` `patchUpdatedSource` | Private; patches `sourcesList` row and `sourceWorkspace.source`; clears workspace stale via `invalidatedKeys.remove` (private) |
| Progress refresh | `WorkspaceSession.swift` `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` | Skip if map never warmed; else Get-one + merge / remove if `isZero`; errors swallowed |
| Skill | `.agents/skills/add-workspace-place/SKILL.md` step 3, step 6, cache-strategy table | Wires `warmQuery`; documents `Spec.lastUpdated` / `lastUpdated: .rows` |
| Docs | `docs/macos-client-patterns.md` rule 4; `docs/deployment-plan/spike-13/deployment-plan.md` S13-01 | "registry row" / "registry `Spec`" for cache invalidation |
| Tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift` | `registryLoadsEveryKey` covers 5 keys; `lastUpdatedPatchesAreDeclaredOnTheFourLists` reads `.scope`; no completeness test |
| Precedent completeness | `ConnectionSentenceRegistryTests.everyKindHasAnEntry`; `PlaceRegistryTests.registryCoversAllPlaceIDs` | Iterate `allCases`, expect one registry entry |

**Precedent.** `CatalogLastUpdatedPatch.rows` (close over `Value` into a type-erased slot) is the typing pattern. `ConnectionSentenceRegistry.entry` + `everyKindHasAnEntry` is the completeness pattern. An exhaustive `Kind` switch (as `Kind.invalidation` already is) is the home for one arm per kind. Copy those; do not invent a protocol-with-`associatedtype Value` table.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `struct CatalogQueryDescriptor` lives in `CatalogQueryRegistry.swift`. Each kind builds one via `extension CatalogQueryKey.Kind { var descriptor: CatalogQueryDescriptor { switch self { … } } }` (exhaustive). `Kind: CaseIterable`. Registry helpers call `key.kind.descriptor` / iterate `Kind.allCases.map(\.descriptor)`. No `specs` array, no `first { $0.kind == }` | Map: descriptor holds value, loader, spec, invalidation; missing kind is a compile error; matches A-30 | A parallel `[CatalogQueryDescriptor]` array plus precondition — second list beside the switch; soft `??` fallbacks |
| D-2 | Descriptor fields: `kind`, `stalePolicy`, `whenHidden`, `invalidateOn`, `invalidation: (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation` (today's `Kind.invalidation` arms moved into the same switch arm), `patches: [CatalogQueryPatch]`, optional `refresh: CatalogQueryRefresh?`, and typed warm/load storage (D-3). No descriptor-level `follows` — each patch and the refresh carry their own (D-4). Policy lookups `preconditionFailure` if somehow reached for an unknown kind (unreachable with an exhaustive switch) | One arm owns the whole kind; fail-closed; avoids one `follows` field serving both `handleScopes` patches and `touchedHandles` invalidation | Keeping `Kind.invalidation` as a second switch; descriptor-level `follows`; release soft fallbacks |
| D-3 | Typed warm: the descriptor factory closes over `Value` and stores `warm: @MainActor (WorkspaceSession, CatalogQueryKey) -> Void` that calls `ensureQuery` / the public query path with a typed loader. Loader type: `(CatalogQueryKey, any GenealogyStore) async throws -> Value`, written in the same `Kind` switch arm with `guard case` for that arm's payload; `preconditionFailure` on mismatch (unreachable when invoked via `key.kind.descriptor`). `WorkspaceSession.query` looks up `key.kind.descriptor` and calls `warm`. Delete `warmQuery`, `load → Any`, and `CatalogQueryRegistryError` | Same erasure as today's `CatalogLastUpdatedPatch.rows`; one switch arm per kind for loader+policies; no new payload-accessor surface on every key | `protocol` + associatedtype table; seven optional id accessors on `CatalogQueryKey`; keeping `warmQuery` beside descriptors |
| D-4 | Rename `CatalogLastUpdatedScope` into `enum CatalogQueryFollows` with cases `.handle`, `.source`, `.evidenceGraphs` (ids = `effects.evidenceGraphs` ∪ optional `mutation?.evidenceGraphId`). Each `CatalogQueryPatch` and `CatalogQueryRefresh` that needs write ids carries its own `follows`. The session computes `ids` from that slot's `follows` and passes them in (same shape as today's `patchLastUpdated`). Mutation-driven patches (`.updatedSourceRow` / `.updatedSourceShell`) have `follows: nil` and read `mutation`. Effects-driven `invalidations(for:)` stays as today (Out); S13-01 adds a separate invalidation entry with its own selector and does not share a field with list patches | One selector per slot; `handleScopes` vs `touchedHandles` stay distinct; T-2 asserts the `follows` the session uses | Descriptor-level `follows` (conflicts when one kind both patches on `.handle` and later invalidates on `.touchedHandles`); bare gathering inside closures with no declared selector |
| D-5 | **Evolve** `CatalogLastUpdatedPatch` into `CatalogQueryPatch` (rename the type; no thin-wrap). Fields: `follows: CatalogQueryFollows?`, and apply that receives `(session, key, ids, activity, mutation?)` — when `follows` is non-nil the session supplies ids+activity; when nil the patch reads `mutation` only. Factories: `.lastUpdated(Row.self, follows:)` (replaces `.rows(_:scope:)`), `.updatedSourceRow`, `.updatedSourceShell` (edits `.source` then `setQueryValue` — never touch `invalidatedKeys`). `sourcesList.patches = [.lastUpdated(CatalogSource.self, .source), .updatedSourceRow]` — array order is composition order. Session walks warmed keys and runs each descriptor's `patches` in order before invalidations | One patch type; T-2 reads real `follows`; S12-26 insert is one more array element | Thin-wrapping `CatalogLastUpdatedPatch` (second mechanism); one opaque `patch` closure; reaching into `invalidatedKeys` |
| D-6 | `struct CatalogQueryRefresh { var follows: CatalogQueryFollows; var run: @MainActor (WorkspaceSession, CatalogQueryKey, [String]) async -> Void }`. Optional on the descriptor. Only `.evidenceGraphProgress` sets it: `follows: .evidenceGraphs`; `run` is today's Get-one + map merge over `ids` for that key. Session after invalidations walks warmed keys whose descriptor has `refresh`, computes `ids` from `refresh.follows`, and if non-empty schedules `Task { await run(session, key, ids) }`. Unwarmed keys are skipped by the walk (no per-entry nil-handle guard). Keep the existing nav-recount `Task` after that (unchanged `navCountTriggers` guard) | Map `refresh(ids)`; per-key walk matches patches so S12-25 citation reads and S12-26 graph-list refresh are body-only entries; nav recount survives | Per-kind refresh with only `ProjectKey` (cannot see cached artifact keys); session-hardcoded progress helper; dropping nav recount |
| D-7 | Completeness test `everyKindHasOneDescriptor`: `Set(Kind.allCases) == Set(Kind.allCases.map { $0.descriptor.kind })` and each `descriptor.kind ==` the case. AC-1.2 plants by asserting a filtered subset fails equality (test-local), not by editing the switch. Repurpose `lastUpdatedPatchesAreDeclaredOnTheFourLists` to read each lastUpdated patch's `follows` on `descriptor.patches` (persons/events/places → `.handle`; sources → `.source`; types → no lastUpdated patch) | Compile-time switch + runtime pin; mirrors `everyKindHasAnEntry` | Only expanding `registryLoadsEveryKey` |
| D-8 | `CatalogMutation` surfaces unchanged | A-29 → S12-32/33 | Folding mutation redesign into this PR |
| D-9 | Update `add-workspace-place` step 3 (descriptor arm, no `warmQuery`), step 6 (`patches` / `follows` instead of `lastUpdated:`), cache-strategy table (patch/refresh via descriptor slots; `patchCached` / `setQueryValue` remain primitives). STEP-6 also rewrites false `Spec` / `lastUpdated` / `warmQuery` lines in `macos-client-patterns.md`, Session doc comments, and spike-13 map/review prose that still say cache-key `Spec` (point them at the descriptor / `follows`) | Working rule names this skill; sweep must reach what this PR makes false | Leaving skill on `warmQuery` / `Spec.lastUpdated` |
| D-10 | No `VERSION` bump; no new chrome; no `pbxproj` change | Internal reshape; S12-31 owns folder sync | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogQueryDescriptor` *(new)* | type | private `Spec` | No value type, loader, follows, patches, or refresh |
| `CatalogQueryFollows` *(new; rename of `CatalogLastUpdatedScope`)* | type | `CatalogLastUpdatedScope` | Needs `.evidenceGraphs` for refresh ids; same id-from-effects job |
| `CatalogQueryPatch` *(new; rename/evolve of `CatalogLastUpdatedPatch`)* | type | `CatalogLastUpdatedPatch` | Must also express updated-Source row/shell; array-valued; old type goes |
| `CatalogQueryRefresh` *(new)* | type | progress helpers on `WorkspaceSession` | Declared per-key refresh with its own `follows` |
| `Kind.descriptor` *(new)* | computed | `Kind.invalidation` switch | Absorbs that switch and Spec into one arm |
| `everyKindHasOneDescriptor` *(new)* | test | `ConnectionSentenceRegistryTests.everyKindHasAnEntry` | Same shape for cache kinds |
| `Kind: CaseIterable` | conformance | `PlaceID` | Not a cache-key kind |

**Builds on.** Today's `CatalogLastUpdatedPatch` erasure pattern (under the new name); exhaustive `Kind` switches; `ensureQuery` / `patchCached` / `setQueryValue`; PlaceRegistry / ConnectionSentenceRegistry completeness.
**Deletes.** private `Spec` / `specs`; `load(key:store:) -> Any`; `warmQuery`; `CatalogQueryRegistryError`; `query`'s `as? Value` cast; `patchLastUpdated`, `patchUpdatedSource`, `refreshEvidenceGraphProgress`, `patchEvidenceGraphProgress`; `lastUpdatedPatch(_:)` accessor; `CatalogLastUpdatedPatch` and `.rows(_:scope:)` (renamed to `CatalogQueryPatch` / `.lastUpdated`); `CatalogLastUpdatedScope` (renamed to `CatalogQueryFollows`); silent `??` policy fallbacks; skill/`docs` lines that teach `warmQuery`, cache-key `Spec`, or `CatalogLastUpdatedPatch`.
**Next adds.** S12-25: citation-read `CatalogQueryRefresh` entries (per cached key) + any new `CatalogQueryFollows` case the reads effect needs. S12-26: graph-list `refresh` + Sources-list insert `patches` element. S12-32/33: shrink `invalidateOn`. S13-01: effects-invalidation entry with its own id selector (map said "registry Spec" — STEP-6 retargets that prose to the descriptor).

### Reach

None — no new path into existing views; no new reachable state.

### Scale

None — no write, list or loop changes its fan-out. Progress still Get-ones only the graphs `follows: .evidenceGraphs` yields; patches still walk warmed keys only.

### Schema

None — no catalog change.

### Go core

None — Swift session only.

### FFI

None — no wire change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Cache-key descriptor | `CatalogQueryDescriptor` *(new)* | `CatalogQueryRegistry.swift` | Built by `Kind.descriptor` |
| Id selector | `CatalogQueryFollows` *(new)* | `CatalogLastUpdated.swift` | `.handle`, `.source`, `.evidenceGraphs`; on each patch/refresh |
| Patch list | `CatalogQueryPatch` *(new)* | same file (renamed from `CatalogLastUpdatedPatch`) | Array on the descriptor |
| Refresh | `CatalogQueryRefresh` *(new)* | same / registry | Per-key; own `follows` |
| Apply funnel | `WorkspaceSession.apply` | `WorkspaceSession.swift` | patches → invalidations → per-key refresh(ids) → nav recount |

States / accessibility — unchanged (no UI).

Sketch (signatures only):

```swift
enum CatalogQueryFollows: Sendable {
    case handle, source, evidenceGraphs
    func ids(in effects: CatalogWriteEffects, mutation: CatalogMutation?) -> [String]
}

struct CatalogQueryPatch: Sendable {
    var follows: CatalogQueryFollows?  // nil → mutation-driven (updatedSource*)
    let apply: @MainActor (
        WorkspaceSession, CatalogQueryKey, Set<String>, CatalogActivity, CatalogMutation?
    ) -> Void
    static func lastUpdated<Row: CatalogLastUpdatedRow>(_: Row.Type, follows: CatalogQueryFollows) -> Self
    static func updatedSourceRow() -> Self
    static func updatedSourceShell() -> Self  // setQueryValue after editing .source
}

struct CatalogQueryRefresh: Sendable {
    var follows: CatalogQueryFollows
    var run: @MainActor (WorkspaceSession, CatalogQueryKey, [String]) async -> Void
}

struct CatalogQueryDescriptor: Sendable {
    let kind: CatalogQueryKey.Kind
    let stalePolicy: CatalogQueryStalePolicy
    let invalidateOn: Set<CatalogMutationKind>
    var whenHidden: CatalogHiddenPolicy = .revalidate
    let invalidation: @Sendable (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation
    var patches: [CatalogQueryPatch] = []
    var refresh: CatalogQueryRefresh? = nil
    // warm: @MainActor (WorkspaceSession, CatalogQueryKey) -> Void  // closed over Value
}

extension CatalogQueryKey.Kind {
    var descriptor: CatalogQueryDescriptor { switch self { /* one arm each */ } }
}
```

### Strings

None — no user-facing copy. Sweep deleted API names in docs/skills/comments (DoD).

## 7. Steps

**STEP-1 — `Kind: CaseIterable` and `Kind.descriptor`**
Files: `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`. Skill: `add-workspace-place`.
Do: Add `CaseIterable` to `Kind`. Introduce `CatalogQueryDescriptor` with D-2 fields. Replace `specs` / `Spec` with `Kind.descriptor` exhaustive switch arms carrying today's stale/hidden/invalidateOn and a temporary loader that still calls through today's `load` switch **or** inlines the same store calls (prefer inlining when cutting over in STEP-2 in the same PR). Move each `Kind.invalidation` arm into the descriptor's `invalidation` property in the same arm. Keep last-updated behavior reachable via early `patches: [.lastUpdated(…)]` — STEP-3 renames `CatalogLastUpdatedPatch` → `CatalogQueryPatch` in the same PR; do not leave a step that drops a kind's policy.
Check: `Kind.allCases.count == 19`; every arm returns `descriptor.kind == self`; tree builds.

**STEP-2 — Typed load/warm; delete `Any` load, `warmQuery`, typeMismatch**
Files: `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`.
Do: Each arm's factory closes over `Value` and store calls (D-3 `guard case`). `query` / `revalidate` / `apply(location:)` use `descriptor.warm`. Delete `load(key:store:) -> Any`, `warmQuery`, `CatalogQueryRegistryError`, and `as? Value`. Policy helpers read `key.kind.descriptor` with no `??` fallbacks.
Check: T-1b `rg` prints nothing; `registryLoadsEveryKey` and an `applyLocation*` test still pass on macOS.

**STEP-3 — `CatalogQueryFollows`, evolve `CatalogLastUpdatedPatch` → `CatalogQueryPatch`, move patches**
Files: `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`.
Do: D-4 and D-5. Rename `CatalogLastUpdatedScope` → `CatalogQueryFollows` and `CatalogLastUpdatedPatch` → `CatalogQueryPatch` (`.rows` → `.lastUpdated`). Wire four list last-updated patches and sourcesList/sourceWorkspace updated-Source patches. Delete `patchLastUpdated` / `patchUpdatedSource` / `lastUpdatedPatch(_:)`. `apply` walks warmed keys, computes ids from each patch's `follows` when non-nil, and runs `patches` in order before invalidations. `updatedSourceShell` uses `setQueryValue`.
Check: T-2, T-3, T-4 and `applyEffectsPatchesLastUpdatedInPlace` pass; `rg CatalogLastUpdatedPatch` under `macos/` and the skill/docs paths in STEP-6 prints nothing.

**STEP-4 — Per-key `CatalogQueryRefresh`; simplify `apply`**
Files: `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`.
Do: `.evidenceGraphProgress` sets `refresh: CatalogQueryRefresh(follows: .evidenceGraphs, run: …)` (D-6). Delete `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress`. After invalidations, walk warmed keys with a refresh and schedule `run` per D-6; **keep** the nav-recount `Task` unchanged after that.
Check: T-7 passes; unwarmed map skips Get-one by absence from the walk; a `.createdSource` still recounts via the hook.

**STEP-5 — Completeness and declaration tests**
Files: `CatalogQueryRegistryTests.swift`. Skill: `add-swift-test`.
Do: Add T-1 (`everyKindHasOneDescriptor` with test-local planting for AC-1.2). Repurpose T-2 onto `descriptor.patches` / `follows`. Keep T-3–T-8 green.
Check: macOS `xcodebuild test` (or CI macOS workflow) for `CatalogQueryRegistryTests` and `WorkspaceSessionTests` passes; AC-1.2 planting fails then restored.

**STEP-6 — Docs**
Files: `.agents/skills/add-workspace-place/SKILL.md`, `docs/macos-client-patterns.md`, `docs/deployment-plan/spike-13/deployment-plan.md` (and its review log if it says cache-key `Spec`), Session file doc comments (`CatalogQueryKey`, `CatalogQueryRegistry`, `WorkspaceSession`, `CatalogLastUpdated`), map row + `completed.md` + this plan on land.
Do: D-9; rewrite `macos-client-patterns.md` rule 4 and `add-workspace-place` cache-strategy / step 6 lines that name `CatalogLastUpdatedPatch` or `lastUpdated: .rows`; DoD sweep names below.
Check: `python3 scripts/check-guidance-names.py` passes; `rg warmQuery .agents/skills/add-workspace-place/SKILL.md` prints nothing; `rg "Spec\\.lastUpdated|CatalogLastUpdatedPatch|CatalogLastUpdatedScope|lastUpdatedPatch|patchLastUpdated|warmQuery|\\.rows\\(" docs/macos-client-patterns.md .agents/skills/add-workspace-place/SKILL.md macos/App/Features/Workspace/Session/ docs/deployment-plan/spike-13/` prints nothing relevant left.

Commits: one per behavior. Subjects are full imperative sentences ending in a period ("Give each catalog cache kind one typed descriptor.", "Warm and load cache keys through the descriptor.", "Move list and Source patches onto descriptor patch slots.", "Refresh evidence graph progress through the descriptor refresh slot.", "Pin every cache kind to one descriptor in tests.", "Teach add-workspace-place the cache-key descriptor.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CatalogQueryRegistryTests.swift` `everyKindHasOneDescriptor` *(new)* | `allCases` ↔ each `descriptor.kind`; planted missing kind fails | AC-1.1, AC-1.2, AC-2.3, D-7 |
| T-1b | Manual | STEP-2 / DoD sweep | `rg -n "func load\\(key: CatalogQueryKey, store:|private func warmQuery" macos/App/Features/Workspace/Session/` → no hits | AC-2.2 |
| T-2 | Swift | `CatalogQueryRegistryTests.swift` `lastUpdatedPatchesAreDeclaredOnTheFourLists` (repurposed) | persons/events/places → a `CatalogQueryPatch.lastUpdated` whose `follows == .handle`; sources → `.source`; types → no lastUpdated patch | AC-3.1, D-4, D-5 |
| T-3 | Swift | `CatalogQueryRegistryTests.swift` `applyUpdatedSourcePatchesListAndWorkspace` | `.updatedSource` → list + workspace patched | AC-3.2 |
| T-4 | Swift | `CatalogQueryRegistryTests.swift` `applySourceWriteEffectsPatchesTheListRowWithoutReloading` | activity → row `updated` moves, no reload | AC-3.1 |
| T-5 | Swift | `CatalogQueryRegistryTests.swift` `hiddenPolicies` | conclusionDetail evict; lists markStale; workspace revalidate | AC-1.3 |
| T-6 | Swift | `CatalogQueryRegistryTests.swift` `registryLoadsEveryKey` + per-kind load tests | loaders return today's shapes | AC-2.1 |
| T-7 | Swift | `WorkspaceSessionTests.swift` `applyGraphEffectsPatchProgressOnly` / `applyPromotedSubjectPatchesProgress` | warmed map merges one graph; unwarmed skips | AC-3.3, AC-3.4, D-6 |
| T-8 | Swift | `WorkspaceSessionTests.swift` `applyLocation*` (existing) | place warm still typed | AC-2.1 |

Must stay green: everything `scripts/check-pr.sh` runs (on macOS for Xcode tests; Linux agents take CI macOS evidence). Suites most likely to break: `CatalogQueryRegistryTests`, `WorkspaceSessionTests`, feature `apply` tests (`SourcesModelTests`, `SourcePageModelTests`, `EvidenceGraphModelTests`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `rg` each deleted/renamed name across `.agents/`, `.cursor/rules/`, `docs/` (including `docs/deployment-plan/spike-13/`), `macos/App/Features/Workspace/Session/` doc comments, and `L10n.swift` with no stale hit; `scripts/check-guidance-names.py` passes. Names: `warmQuery`, `load(key:store:) -> Any`, private registry `Spec`, `Spec.lastUpdated`, `lastUpdatedPatch`, `patchLastUpdated`, `patchUpdatedSource`, `refreshEvidenceGraphProgress`, `patchEvidenceGraphProgress`, `CatalogQueryRegistryError`, `CatalogLastUpdatedScope`, `CatalogLastUpdatedPatch`, `.rows(` (the old factory)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: `macos-client-patterns.md`, `add-workspace-place`, spike-13 map prose retargeted from cache-key `Spec` to descriptor / `follows`
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Engineer-facing; no chrome change. Slice 6 Check for this PR is the automated suite.

1. Confirm `everyKindHasOneDescriptor` fails under the AC-1.2 planting, then passes with the full set.
2. Run the AC-2.2 `rg` from T-1b; confirm no `load(key:store:)` or `warmQuery` remains under `macos/App/Features/Workspace/Session/`.
3. Optional smoke on a dogfood project: open Sources, open a Source, place a Subject on a graph, return to Sources — list last-updated and graph progress still move without a full progress-map reload of unrelated graphs.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Type-erasure drifts into a second warm list | D-3; STEP-2 `rg` |
| Patch order on sourcesList | D-5 array order: lastUpdated then updatedSourceRow |
| Invalidation order if someone moves effects fan-out early | Out — leave `invalidations(for:)` alone |
| `guard case` in loaders becomes a third hand-kept list | Only inside the same exhaustive `Kind` arm; mismatch is preconditionFailure |

**Stop and report if:**

- A **Current state** row is false at the implementer's `main` (a new `Kind` landed, or `apply` order changed).
- A typed descriptor factory cannot drive `session.query` without reintroducing `warmQuery` or `load → Any` (D-3 cannot hold).
- An existing test in T-2–T-8 fails for a behavior reason that would require changing an Out line.
- S12-25/S12-26 cannot express their id sets as a `CatalogQueryFollows` case or a `refresh(ids)` entry — stop; do not expand into those PRs.

## 12. Open questions

None.

## 13. PR

**Title:** S12-34 — One descriptor per cache key
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-34-one-descriptor-per-cache-key.md`.
