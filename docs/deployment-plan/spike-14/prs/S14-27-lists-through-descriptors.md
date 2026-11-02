# S14-27 — Every write reaches its keys through their descriptors

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes (batch 5; S13-01 ship log)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: no view draws anything different, and no reload a researcher sees changes)
**Board:** —
**Depends on:** S14-16 (PR [#686](https://github.com/mendahu/provenencia/pull/686), open at planning; its branch is the planning base)
**Blocks:** S14-17
**Branch:** `s14-27-lists-through-descriptors`
**Version bump:** none — no researcher-visible change, as S14-24 took none ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [`macos-client-patterns.md`](../../../macos-client-patterns.md) cache contract rule 2; [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) "Refresh"; S13-01 ship log (a cache-wide look after S13-01 moved the other effect slots onto descriptor `refreshes`); the S14-17 plan, which uses this slot ([`S14-17-vocabulary-effects.md`](S14-17-vocabulary-effects.md) D-6; map reorder, user 2026-11-02)
**Planned:** 2026-11-02 against `s14-16-promote-source-effects` @ `a5e057b` (S14-16's head = `main` @ `04a2c99` + S14-16)
**Review:** — (not reviewed yet)

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

A write's effects reach the keys they reload by one path. Each cached kind already declares in its own descriptor the ids it follows (`touchedHandles`, `evidenceGraphs`, `reads`); this PR lets it declare the catalog-wide `lists` and `structure` it follows the same way, and deletes the hand-written switch in `CatalogQueryRegistry` that names them centrally. The registry keeps only what a mutation name invalidates. No reload a researcher sees changes.

**Why now.** S14-17 (vocabulary writes) splits `vocabulary` into two lists that thirteen keys follow. Planned against the switch, it grew the switch twelvefold right before this PR deleted it; the user moved this PR ahead of S14-17 (2026-11-02, S14-17 review M-2). After this PR a key that follows a list is one `refreshes` line on its descriptor.

**What stays.** Which keys every write reloads, and when. The mutation path (`invalidateOn`, `invalidation`, `invalidations(by:project:)`) and every `CatalogMutation` case. The `counts` list and its recount (`WorkspaceSession.onNavCountsMoved`). `CatalogQueryRefresh`'s two actions. The Go side and the wire.

## 2. User stories

**US-1.** As a contributor, when a key's payload depends on a catalog-wide list or on structure, I declare it on that key's descriptor, beside the ids it follows, and nowhere else.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `CatalogQueryKey.Kind.propertiesWorkspace.descriptor.refreshes.map(\.follows)` | `[.list(.vocabulary), .list(.propertyUsage), .structure]` |
| AC-1.2 | `rg -n 'invalidations\(for' macos/App` | No hit |
| AC-1.3 | Every kind's descriptor | No `refreshes` entry follows `.list(.counts)`; no `.run` (merged) entry follows `.list(_)` or `.structure` |

**US-2.** As a researcher, every write reloads what it reloaded before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A warmed, visible `propertiesWorkspace`; `apply(CatalogWriteEffects(lists: [.vocabulary]))` | It revalidates once |
| AC-2.2 | The same with `[.propertyUsage]`, with `structure: true` alone, with `[.vocabulary, .propertyUsage]` and `structure: true` together | It revalidates once each time |
| AC-2.3 | A warmed, hidden `propertiesWorkspace` (`whenHidden: .markStale`); `apply(CatalogWriteEffects(lists: [.vocabulary]))` | It is marked stale and not fetched |
| AC-2.4 | `apply(CatalogWriteEffects(lists: [.counts]))` with `onNavCountsMoved` set and a warmed `propertiesWorkspace` | The recount runs once; `propertiesWorkspace` does not revalidate |
| AC-2.5 | `apply(CatalogWriteEffects(evidenceGraphs: ["g1"], lists: [.propertyUsage]), .deletedSubject)` with warmed `propertiesWorkspace` and `evidenceGraph(g1)` visible | Each revalidates once (the effects and the mutation both name `propertiesWorkspace`) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CatalogQueryFollows.list(CatalogList)` and `.structure` (D-1) | STEP-1, AC-1.1, AC-1.3 |
| `propertiesWorkspace` follows `vocabulary`, `property_usage` and structure on its descriptor (D-2) | STEP-1, AC-1.1, AC-2.1–AC-2.3 |
| The lists / structure half of the registry and its combined entry point go; `apply` takes the mutation's invalidations alone (D-3) | STEP-2, AC-1.2, AC-2.4, AC-2.5 |
| Tests move to the session path (D-4) | STEP-1, STEP-2 |
| Guidance (D-5) | STEP-3 |

| Out | Home |
| --- | --- |
| Splitting `vocabulary` into two lists and the keys that follow them | S14-17 |
| Removing `invalidateOn` or any `CatalogMutation` case | S14-17 shrinks them; not planned beyond |
| A sentinel-free id model for catalog-wide slots (a separate `Bool` follow API) | not planned (D-1 Rejected) |
| `.deletedSubject` on `propertiesWorkspace`, now that a Subject delete reports `property_usage` | not planned in this PR: removing it would change nothing a researcher sees but is a mutation-path change; S14-17 lists every remaining `invalidateOn` |

## 4. Current state

Verified on `s14-16-promote-source-effects` @ `a5e057b`.

| Area | Where | Today |
| --- | --- | --- |
| Follow slots | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryFollows` (`handle`, `source`, `reads(CatalogRead)`, `touchedHandles`, `evidenceGraphs`), `ids(in:)` | Each slot returns the ids of one effects field; an empty list means the write did not move it |
| Refresh entries | same file, `CatalogQueryRefresh` (`.revalidate`, `.run`), `applyHiddenPolicy(follows:ids:)`, `applyHiddenPolicy(followsValue:ids:)`, `applyHiddenPolicyOnAny(follows:)`, `merged(follows:as:read:merge:afterMerge:)` | `applyHiddenPolicyOnAny` revalidates a warmed key whenever its slot reports any id |
| Lists switch | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for effects:project:)`, `static invalidations(for list:project:)`, `invalidations(for:by:project:)` | `.vocabulary` and `.propertyUsage` → `.key(.propertiesWorkspace)`; `.counts` → none; `structure` → `.key(.propertiesWorkspace)`; the third unions the first with `invalidations(by:project:)`, deduped |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)` (line 268), `revalidatedKeys(for:)`, `cachedKeys(for:)`, `applyHiddenPolicy(to:)` | `apply` takes `cachedKeys(for: registry.invalidations(for: effects, by: mutation, project:))`, adds `revalidatedKeys(for: effects)`, the patches' awaiting keys and the `.run` fallbacks, and applies each key's hidden policy once; it runs `onNavCountsMoved` on `counts`. Its doc names "the catalog-wide lists and structure the effects reported" as registry work |
| Properties descriptor | `CatalogQueryRegistry.swift` `case .propertiesWorkspace` | `invalidateOn: [.createdProperty, .updatedProperty, .deletedProperty, .mutatedSubjectTypeProperties, .deletedSubject]`, `whenHidden: .markStale`, no `refreshes`; its comment says Observation writes "report their Source in `WriteEffects`, which invalidates it" |
| Docs in code | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogList.counts` doc, `CatalogWriteEffects.lists` doc ("see `CatalogQueryRegistry.invalidations(for:project:)`") | Point at the switch |
| Guidance | `.agents/skills/add-workspace-place/SKILL.md` line 64 ("Refresh" row: slots `.reads(…)`, `.touchedHandles`, `.evidenceGraphs`; "never a block in `invalidations(for:project:)`, which keeps only `lists` and `structure`"); `docs/macos-client-patterns.md` rule 2 | The skill names the switch as the lists' home |
| Tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`: `listCases` / `listInvalidations` (around 1236–1255), `countsListNamesNoKey` (1285), and `invalidations(for: effects, project:)` assertions at 661, 899–910, 912–918 (the `by: .updatedProperty` union), 929–931, 1020, 1024, 1297, 1333, 1380; `macos/ProvenenciaTests/WorkspaceSessionTests.swift` 942 (`isEmpty`), 1012 (union with `.updatedProperty`), 1472 (`apply(CatalogWriteEffects(lists: [.vocabulary]))`) | Assert the switch's output |

**Precedent.** S13-01 moved the Evidence graph, list and detail keys off registry blocks onto descriptor `refreshes` (`.touchedHandles`, `.evidenceGraphs`), and S14-11 added `.reads(…)` the same way; `refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)]` on `personsList` is the shape to copy. Do not copy a `.merged` entry: a list names no rows to re-read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `CatalogQueryFollows` gains `case list(CatalogList)` and `case structure`. `ids(in:)` returns `[list.rawValue]` when `effects.lists.contains(list)`, else `[]`; `.structure` returns `["structure"]` when `effects.structure`, else `[]`. Their doc says: a catalog-wide slot names no row, so it reports one token when the write moved it and is used only with `applyHiddenPolicyOnAny(follows:)`. `CatalogQueryRefresh.merged(follows:…)` gets `precondition(!follows.isCatalogWide, …)`, with `var isCatalogWide: Bool` *(new)* on `CatalogQueryFollows` (`true` for `.list`, `.structure`) | One slot type, one session pass (`revalidatedKeys` already treats any reported id as a hit for `applyHiddenPolicyOnAny`), no change to `WorkspaceSession`'s loops. The precondition keeps a list from reaching a `.run` that would read its token as a row id | A `refreshesOnLists: Set<CatalogList>` property beside `refreshes` — a second mechanism for one job; `structure` as a `CatalogList` case — Go's `Set.Structure` is a flag, not a list, and the wire carries it as `WriteEffects.structure` |
| D-2 | `case .propertiesWorkspace` gets `refreshes: [.applyHiddenPolicyOnAny(follows: .list(.vocabulary)), .applyHiddenPolicyOnAny(follows: .list(.propertyUsage)), .applyHiddenPolicyOnAny(follows: .structure)]`. Its comment says it follows a vocabulary label change, an Observation's Property usage and graph structure through these entries, and the mutations in `invalidateOn` | Exactly what the switch named, nothing more | — |
| D-3 | Deleted: `CatalogQueryRegistry.invalidations(for effects:project:)`, `static invalidations(for list:project:)`, `invalidations(for:by:project:)`. `WorkspaceSession.apply` computes `let invalidated = mutation.map { cachedKeys(for: registry.invalidations(by: $0, project: projectKey)) } ?? []`; the rest of `apply` is unchanged. `apply`'s doc names the registry's part as "what the mutation names" and the descriptor `refreshes` as following every effects slot, catalog-wide lists and structure included. `CatalogList.counts` doc: no descriptor follows it; the session recounts. `CatalogWriteEffects.lists` doc: "each key that holds one follows it through its descriptor `refreshes`" | One path from effects to keys (`revalidatedKeys`) beside the mutation names | Keeping `invalidations(for:by:project:)` as a mutation-only shim — a second name for `invalidations(by:project:)` |
| D-4 | Tests. `listCases` / `listInvalidations` become `listsAndStructureRevalidatePropertiesOnce` *(new)*, parameterized over the same five cases, on a session with a warmed visible `propertiesWorkspace`: `apply` → one revalidation, counted by `FakeStore.getPropertiesWorkspaceCalls` *(new; incremented in `getPropertiesWorkspace`, as `getSourceWorkspaceCalls` is in `getSourceWorkspace`)*. `countsListNamesNoKey` becomes `noDescriptorFollowsCounts` *(renamed)* over `CatalogQueryKey.Kind.allCases`. The `invalidations(for: effects, project:).isEmpty` assertions (661, 929–931, 1020, 1024, 1297, 1333, 1380; `WorkspaceSessionTests` 942) are deleted, since the API is gone and each test's `revalidatedKeys` or `refreshes.map(\.follows)` assertion beside it already says where the slot goes. The 899–910 assertion becomes `session.revalidatedKeys(for: thoseEffects)` containing the warmed `propertiesWorkspace` key. The 912–918 union becomes `invalidations(by: .updatedProperty, project:)` not containing `.key(.evidenceGraph(…))`. `WorkspaceSessionTests` 1012 becomes `invalidations(by: .updatedProperty, project:)` containing `.allCached(.personClaims)` | Each assertion keeps what it guarded, through the path that now holds it | — |
| D-5 | Guidance: `add-workspace-place` "Refresh" row lists the slots `.reads(…)`, `.touchedHandles`, `.evidenceGraphs`, `.list(…)`, `.structure`, says a catalog-wide slot is used only with `applyHiddenPolicyOnAny(follows:)`, and replaces "never a block in `invalidations(for:project:)`, which keeps only `lists` and `structure`" with "never a block in the registry, which keeps only what a mutation names". `macos-client-patterns.md` rule 2 gains one clause: the catalog-wide `lists` and `structure` reach their keys through descriptor `refreshes` too | Sweep rule | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogQueryFollows.list(_:)`, `.structure` *(new)* | enum cases | `CatalogQueryFollows.reads(_:)` | Extends the enum; same shape |
| `CatalogQueryFollows.isCatalogWide` *(new)* | computed property | none | The one fact `merged` needs |
| `listsAndStructureRevalidatePropertiesOnce`, `noDescriptorFollowsCounts` | tests | `listInvalidations`, `countsListNamesNoKey` | Replace them |
| `FakeStore.getPropertiesWorkspaceCalls` *(new)* | test hook | `FakeStore.getSourceWorkspaceCalls` | Same counter for the read T-3 counts |

Searched: `rg 'CatalogQueryFollows' macos`, `rg 'invalidations\(for' macos`, `rg 'applyHiddenPolicyOnAny' macos/App`.

**Builds on.** S13-01 / S14-11's descriptor `refreshes`; `WorkspaceSession.revalidatedKeys`; `applyHiddenPolicyOnAny(follows:)`.
**Deletes.** `invalidations(for effects:project:)`, `static invalidations(for list:project:)`, `invalidations(for:by:project:)`; `listInvalidations`, `countsListNamesNoKey` (replaced); the `invalidations(for:)` test assertions of D-4; the switch's doc lines.
**Next adds.** S14-17: a key that holds a vocabulary label adds one `applyHiddenPolicyOnAny(follows: .list(…))` entry, and the two families are two `CatalogList` cases with no registry change. A new catalog-wide list is a `CatalogList` case and an entry on each descriptor that follows it.

### Reach

None — no new path into existing code; the same keys revalidate on the same writes.

### Scale

None — no write, list or loop changes what it loads. `revalidatedKeys` already walks every warmed key's `refreshes` per `apply`; three more entries on one kind add three `ids(in:)` checks per `apply` for each warmed `propertiesWorkspace` key (at most one per project).

### Edges

- **Assumptions.** `revalidatedKeys` treats a `nil` `dependsOn` (from `applyHiddenPolicyOnAny`) as "meets any reported id", so a one-token slot hits exactly when the write moved the list (T-1, T-2). `apply` dedupes `invalidated + refreshed`, so a key named by both a mutation and a list revalidates once (AC-2.5, T-3). `.run` fallbacks never see a catalog-wide slot (D-1 precondition, T-4).
- **Neighbors.** `counts` with a warmed `propertiesWorkspace`: recount only (AC-2.4). Effects with no list and no structure: no change from today (the existing session tests).
- **Removed inputs.** None — nothing is deleted or excluded by this change.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — no wire change.

### macOS

None — no chrome change. Session and registry changes are D-1 to D-3.

### Strings

None — no copy.

## 7. Steps

`xcodebuild` needs the macOS toolchain (see Stop conditions).

**STEP-1 — Keys follow lists and structure on their descriptor**
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Platform/FakeStore.swift` (`getPropertiesWorkspaceCalls`), `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`.
Do: T-1 and T-4 first. D-1, D-2. The switch still exists in this step, and `apply` dedupes, so each key still revalidates once.
Check: T-1, T-2, T-4 pass under `xcodebuild test` (as `scripts/check-pr.sh` runs it).

**STEP-2 — The registry keeps only the mutation path**
Files: `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: D-3, D-4.
Check: `rg -n 'invalidations\(for' macos` prints nothing; T-3, T-5 and every existing `WorkspaceSessionTests` / `CatalogQueryRegistryTests` case pass; `xcodebuild test` passes with no Swift warning.

**STEP-3 — Guidance and docs**
Files: `.agents/skills/add-workspace-place/SKILL.md`, `docs/macos-client-patterns.md`; `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (S14-27 row status), this plan (status, Landed note).
Do: D-5; the Definition of done's docs lines.
Check: `rg -n 'invalidations\(for' .agents docs/*.md .cursor` prints nothing; `python3 scripts/check-guidance-names.py` passes; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CatalogQueryRegistryTests.swift` `catalogWideSlotsReportOneToken` *(new)* | `CatalogQueryFollows.list(.vocabulary).ids(in: CatalogWriteEffects(lists: [.vocabulary]))` → `["vocabulary"]`; with `lists: [.propertyUsage]` → `[]`; `.structure` with `structure: true` → `["structure"]`, with `false` → `[]`; `isCatalogWide` true for both, false for `.touchedHandles` | D-1 |
| T-2 | Swift | `CatalogQueryRegistryTests.swift` `propertiesWorkspaceFollowsListsAndStructure` *(new)* | `Kind.propertiesWorkspace.descriptor.refreshes.map(\.follows)` → `[.list(.vocabulary), .list(.propertyUsage), .structure]`; on a session with the key warmed, `revalidatedKeys(for: CatalogWriteEffects(lists: [.vocabulary]))` → `[propertiesWorkspace]`, and `(structure: true)` → the same | AC-1.1, D-2 |
| T-3 | Swift | `CatalogQueryRegistryTests.swift` `listsAndStructureRevalidatePropertiesOnce` *(new, replaces `listInvalidations`)*, five cases | Warmed, visible `propertiesWorkspace`; `apply` with each case (`[.vocabulary]`; `[.propertyUsage]`; `structure: true`; `[.vocabulary, .propertyUsage]`; `[.vocabulary]` + `structure: true`) → `getPropertiesWorkspaceCalls` up by one. Hidden case: `[.vocabulary]` → `isStale`, no load. With `.deletedSubject` and `evidenceGraphs: ["g1"]` → `propertiesWorkspace` and `evidenceGraph(g1)` load once each | AC-2.1, AC-2.2, AC-2.3, AC-2.5 |
| T-4 | Swift | `CatalogQueryRegistryTests.swift` `noDescriptorFollowsCounts` *(renamed from `countsListNamesNoKey`)* | For every `CatalogQueryKey.Kind.allCases` descriptor: no refresh follows `.list(.counts)`; no `.run` refresh follows a catalog-wide slot | AC-1.3, D-1 |
| T-6 | Swift | STEP-2's `rg -n 'invalidations\(for' macos` check (a source check, not a test file) | No hit after D-3 | AC-1.2, D-3 |
| T-5 | Swift | `WorkspaceSessionTests.swift` near 1465 (the `onNavCountsMoved` test, changed) | Add a warmed `propertiesWorkspace`; `apply(CatalogWriteEffects(lists: [.counts]))` → `calls == 1`, no `getPropertiesWorkspace` load | AC-2.4 |

Must stay green: everything `scripts/check-pr.sh` runs; most likely to break: `CatalogQueryRegistryTests`, `WorkspaceSessionTests`, the composer and Properties model tests that apply `property_usage` effects (`CitationComposerModelTests`, `PropertiesModelTests`).

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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md rule 2 (D-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project, go to Properties, then open a Source's citation composer and save an Observation: the composer's Property list shows the Observation's count change, as today (`property_usage`).
2. On Properties rename a user Property: the page shows the new label, as today (`vocabulary`).
3. Add a Source: the sidebar's Sources badge goes up by one, as today (`counts`), and the Properties page does not reload (Instruments or the debug log `PROVENENCIA_DEBUG_SESSION_APPLY=1` shows no `getPropertiesWorkspace`).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A key loses a reload the switch gave it | T-3 covers every case the switch handled; the switch named only `propertiesWorkspace` |
| A list token reaches a `.run` entry as a row id | D-1 precondition; T-4 |

To compare the base: `git fetch origin main && git diff a5e057b origin/main --stat -- macos/App/Features/Workspace/Session macos/ProvenenciaTests`, then read each §4 path that shows.

**Stop and report if:**

- `invalidations(for list:project:)` names any key other than `propertiesWorkspace` at the base (S14-16 or a later PR added one): D-2 does not cover it.
- Any caller outside `WorkspaceSession.apply` and the tests of §4 uses `invalidations(for:project:)` or `invalidations(for:by:project:)`.
- `revalidatedKeys` no longer treats a `nil` `dependsOn` as "any reported id" at the base.
- `xcodebuild` is missing: report and stop; every step needs it.

## 12. Open questions

None.

## 13. PR

**Title:** S14-27 — Every write reaches its keys through their descriptors
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-27-lists-through-descriptors.md`.
