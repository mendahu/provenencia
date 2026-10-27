# S12-20 — Docs and skills match the code

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: docs only, no Swift or Go file changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s12-20-docs-match-code`
**Version bump:** none — docs only ([`docs/versioning.md`](../../../versioning.md): "Do not bump the product version for docs-only or comment-only changes")
**Skills to load:** [`archive-docs`](../../../../.agents/skills/archive-docs/SKILL.md) ("Shipped idea"), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the checklist this PR extends)
**Rules that bind:** `.cursor/rules/catalog-refs.mdc` (this PR edits it)
**Context:** [app-health audit](../app-health-audit.md) A-6, A-23, A-33; working rules "Each PR updates the docs and skills it makes false" (slices 5–7 line: "S12-20 points `add-catalog-model` step 6 at `core/database/effects/README.md` instead of listing `Effect` fields") and "Fix the High and Medium items, grouped by shared files or behavior"; map review K-13, K-20, K-21
**Planned:** 2026-10-27 against `main` @ `67988a5`
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

The skills, Cursor rules, domain docs and kit README that agents and the product owner read describe the code as it is at the planning base. Today five guidance files teach a pattern the code dropped (A-6), eight domain docs contradict the code (A-23), and the DesignSystem README has no row for six kit components (A-33), so the next plan or PR built from them starts wrong. Two idea files that shipped are archived or trimmed so the ideas folder lists only open work.

**Why now.** Slices 5–7 plan from these docs; S12-21 – S12-38 would each re-derive the same corrections. The dogfood bar's item 13 ("the docs, skills and kit README describe the code as it is") closes here for the audit's findings.

**What stays.** No Go, Swift, proto or script file changes. The docs that later slice 5–7 PRs rewrite (`swift-tests.mdc`'s FakeStore paragraph, the effects README's per-table rules, the pbxproj checklist line) keep their current text.

## 2. User stories

**US-1.** As a contributor following a skill or Cursor rule, when I add a catalog table, a searchable kind or a ref prefix, every name and step it gives exists in the code, so that I do not build on a deleted field or miss a required case.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Read `add-catalog-model` step 6 | It names no `Effect` field; it links `core/database/effects/README.md` for the fields. `rg -n CitationSource .agents/skills/add-catalog-model/SKILL.md` → no hit |
| AC-1.2 | Read `add-searchable-kind` | Its checklist and step 3 name `core/search/fts.go` `fieldValuesForKind` and the silent fallback. `rg -c fieldValuesForKind .agents/skills/add-searchable-kind/SKILL.md` → `2` |
| AC-1.3 | Read `.cursor/rules/catalog-refs.mdc` | Reserved prefixes list `USR`, `SRC`, `ART`, `CIT`, `OBS`, `GPH`, matching `core/ref/ref.go` `reservedPrefixes` |

**US-2.** As the product owner reading a domain doc, when a doc states how matching, deletes, effects, Copy Source or the client is built, the statement is true at `main`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `docs/conclusion-reconciliation.md` §6 | Says search reads every cached value and matching reads the profile (displayed positive, denied negative). `rg -n 'search and matching read them all' docs/conclusion-reconciliation.md` → no hit |
| AC-2.2 | `docs/matching.md` | Its layer table has rows 5 and 6 naming `core/segmentcompare`, `core/graphalign`, `core/shortlist`, `core/kinship`; no line says S11-14, S11-15 or S11-16 is not built. `rg -n 'S11-1[456]' docs/matching.md` → no hit |
| AC-2.3 | `docs/matching-model.md` line 3 | Does not say the layers have not landed. `rg -n 'until then' docs/matching-model.md` → no hit |
| AC-2.4 | `docs/catalog-deletes.md` report sketch | The `gate` enum line lists `last_graph`. `rg -n 'edge_locked \| last_graph \| infra' docs/catalog-deletes.md` → one hit |
| AC-2.5 | `docs/macos-client-patterns.md` folder table | `Features/EvidenceGraph/` reads as one Evidence graph by graph id. `rg -n 'for one source' docs/macos-client-patterns.md` → no hit |
| AC-2.6 | `core/database/effects/README.md` | Says a Source title edit names the Source's graphs and reprojects their search documents. `rg -n 'title' core/database/effects/README.md` → a hit naming `onField` and `DocEvidenceGraph` |
| AC-2.7 | `api/proto/README.md` | `rg -n 'no Source tables' api/proto/README.md` → no hit |
| AC-2.8 | `docs/source-layer-data-model.md` §4.2 | Says the copy gets its own Default graph linking the copied Artifacts. `rg -n 'own Default graph' docs/source-layer-data-model.md` → one hit, inside §4.2 |
| AC-2.9 | `docs/promote-graph-alignment.md` §7 | `rg -n "Source's Evidence graph memo" docs/promote-graph-alignment.md` → no hit |

**US-3.** As a contributor reaching for a kit component, every folder under `macos/App/DesignSystem/Components/` and `macos/App/DesignSystem/Recipes/` has a README row, and `add-ui-component` tells me to add one.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The README row loop in STEP-4's check | Prints nothing (every folder has a row) |
| AC-3.2 | `add-ui-component` checklist | Has a line requiring a README row for a new Components/Recipes folder. `rg -n 'README row' .agents/skills/add-ui-component/SKILL.md` → one hit |

**US-4.** As the product owner reading the ideas folder, a shipped idea is in `docs/ideas/archive/` and a done item is gone from an open idea.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `docs/ideas/` | `settled-list-count-publish.md` is only under `docs/ideas/archive/`, distilled per `archive-docs` "Shipped idea"; every link to it resolves |
| AC-4.2 | `docs/ideas/shared-list-helpers.md` | Has no "Query-wait test helper" section. `rg -n 'Query-wait' docs/ideas/shared-list-helpers.md` → no hit |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-6: `add-catalog-model` step 6 points at `core/database/effects/README.md` (D-1) | STEP-1, AC-1.1 |
| A-6: `add-searchable-kind` names `fieldValuesForKind` (D-2) | STEP-1, AC-1.2 |
| A-6: `catalog-refs.mdc` reserves `GPH` | STEP-1, AC-1.3 |
| A-23: `conclusion-reconciliation.md` §6, `matching.md`, `matching-model.md` intro (D-3, D-4) | STEP-2, AC-2.1 – AC-2.3 |
| A-23: `docs/catalog-deletes.md`, `docs/macos-client-patterns.md`, `core/database/effects/README.md`, `api/proto/README.md`, `docs/source-layer-data-model.md` §4.2, `docs/promote-graph-alignment.md` (D-5) | STEP-3, AC-2.4 – AC-2.9 |
| A-33: six DesignSystem README rows and an `add-ui-component` checklist line (D-6) | STEP-4, AC-3.1, AC-3.2 |
| Archive `docs/ideas/settled-list-count-publish.md`; drop "Query-wait test helper" from `docs/ideas/shared-list-helpers.md`; ideas README and map links follow (D-7) | STEP-5, AC-4.1, AC-4.2 |

| Out | Home |
| --- | --- |
| `swift-tests.mdc` event-titles name and `design-system-layers.mdc` example (A-6 bullets 2 and 5) | Done before this plan, in `22708a81` (verified at the base: `eventTitlesByGraph`, no `PVEvidenceIcon`) |
| Rewriting `swift-tests.mdc`'s FakeStore "does not re-implement" paragraph | S12-35 |
| The `add-searchable-kind` reshape (kinds own their columns and locations) | S12-36 |
| `add-ui-component` checklist line "New .swift registered in the Xcode project (pbxproj)" | S12-31 |
| A `check-pr.sh` check that fails on a kit folder with no README row | not planned (map scope entry "Not this PR") |
| `core/database/effects/README.md` "Artifact and Citation rows name every graph of their Source" (true today; changes with the code) | S12-25 |
| `add-catalog-model` steps other than 6 | not planned (no audit finding) |
| The audit's Low items and nits | not scheduled ([`app-health-audit.md`](../app-health-audit.md#not-scheduled)) |
| The settled idea's open question (a stale key read as settled) | S12-26 (working rule "A hidden graph or Source page goes stale") |
| The other items in `docs/ideas/shared-list-helpers.md` | stay in that idea; "Fake store keys" / "Go reads every Swift mirror" are S12-35 |

## 4. Current state

Verified at `67988a5`.

| Area | Where | Today |
| --- | --- | --- |
| Effects skill step | `.agents/skills/add-catalog-model/SKILL.md` step 6 (line 26) | Lists `Source`, `CitationSource`, `SourceScope`, `Handles`, `HandleScope`, `Search`, `Lists`, `ListsOnField`, `Structure`. `core/database/effects/registry.go` `Effect` has `EvidenceGraph`, not `Source`, since S12-08; the list will drift again with S12-25 / S12-26 |
| Effects owner doc | `core/database/effects/README.md` | Describes every `Effect` field and how to add a list. Does not say the `sources` entry names its graphs on a title edit: `registry.go` `"sources"` has `EvidenceGraph: onField("title", graphsOfSource(self()))` and a `DocEvidenceGraph` search doc on the same path |
| Searchable-kind skill | `.agents/skills/add-searchable-kind/SKILL.md` checklist and "3. Projection + retrieval" | Names `locationFor`, `searchindex.Reproject`, `RebuildAll`, `docColumns` / `scanDoc`; never `fieldValuesForKind` |
| Field mapping | `core/search/fts.go` `fieldValuesForKind` (line 425) | Maps a doc's `title` / `ref` / `secondary` / `body` columns onto the registry field names per kind; `default:` returns only `title`, `ref`, `body`, so a declared `description` field scores zero with no error. Called by `scoreFields` callers at lines 149 and 408 |
| Ref rule | `.cursor/rules/catalog-refs.mdc` line 13 | "Reserved prefixes: `USR`, `SRC`, `ART`, `CIT`, `OBS`." `core/ref/ref.go` `reservedPrefixes` also has `PrefixEvidenceGraph = "GPH"`; `add-catalog-ref` already lists `GPH` |
| Reconciliation doc | `docs/conclusion-reconciliation.md` §6, paragraph "Eliminated means not displayed, never dropped." | "search and matching read them all". Search does (`core/database/searchindex/handles.go` `handleDocument`: other cached names stay searchable); matching reads `core/subjectprofile` `Build` (displayed → positive, `ReasonDenied` → negative, the rest left out), as `docs/matching-model.md` line 28 says |
| Matching map | `docs/matching.md` layer table and "## Not built yet" | Table has layers 1–4 only. "Not built yet" lists Segment compare (S11-14), the shortlist index (S11-15) and the layer 6 walk (S11-16); the spike-11 map shows them Landed #493, #497, #501, and `core/segmentcompare`, `core/shortlist`, `core/kinship` are imported by `core/graphalign` and `core/database/promotealign` |
| Matching model | `docs/matching-model.md` line 3 | "Layers land in S11-02 – S11-16; until then [`matching.md`] describes the running code." All six layers run |
| Delete report | `docs/catalog-deletes.md` line 73 | `gate enum // ok \| inbound \| not_found \| edge_locked \| infra \| origin_locked`. `core/database/deleteimpact/kinds.go` also has `GateLastGraph = "last_graph"`; line 11 of the same doc already names it |
| Client folders | `docs/macos-client-patterns.md` line 26 | `Features/EvidenceGraph/` "Evidence graph for one source". `EvidenceGraphView` opens on a graph id since S12-08 |
| Proto README | `api/proto/README.md` "Rebuild the Mac library", line 36 | Ends "There are no Source tables." The catalog holds `sources`, `artifacts`, `evidence_graphs`, … |
| Copy Source doc | `docs/source-layer-data-model.md` §4.2 | Lists what is and is not copied. Does not say the copy gets its own Default graph (`core/database/sources/sources.go` `Create` → `evidencegraphs.CreateDefault`) and that each copied Artifact links to it (`core/database/sourcecopy/sourcecopy.go` calls `artifacts.Create` with no `GraphIDs`; `artifacts.Create` → `evidencegraphs.LinkArtifact`, empty ids = every graph of the Source) |
| Promote loading | `docs/promote-graph-alignment.md` line 167 | "on the Source's Evidence graph memo". `graphcache` keys Evidence graphs and their memos by graph id (line 165 of the same doc) |
| Kit README | `macos/App/DesignSystem/README.md` "What's built vs. not yet" table (lines 149–188) | No row for `Components/Chip`, `Components/Copy`, `Components/FloatingMenuSelection`, `Components/HoverEffect`, `Components/InlineEdit`, `Recipes/LastUpdated`; every other folder has one |
| UI skill checklist | `.agents/skills/add-ui-component/SKILL.md` lines 63–73 | No line about the README row |
| Shipped idea | `docs/ideas/settled-list-count-publish.md` | Status "idea only". Shipped in `c6dcd51d`, `aedf7aa2`, `228bd821` (`onSettled` in `PropertiesView.swift`, `SourcesModel.swift`). Linked from `docs/ideas/README.md` line 32 and `docs/deployment-plan/spike-12/deployment-plan.md` lines 54, 143, 477, 545 |
| Done idea item | `docs/ideas/shared-list-helpers.md` "## Query-wait test helper" | Done: `func waitForQuery` exists only in `macos/ProvenenciaTests/QueryHandleWaiting.swift`. `docs/ideas/README.md` line 30 describes the file as "one list of graph mutations (the other two items are in Spike 11)", which matches no section in it |
| Checks | `scripts/check-guidance-names.py`, `scripts/check-plan-docs.py` | Both pass at the base |

**Precedent.** `15d725fc` ("Document the shared subject-mark rule and archive the glyph idea with its leftovers"): a docs-only change that moved the glyph idea into `docs/ideas/archive/` with a **Done in S10-49** status line and a Leftovers paragraph, dropped its `docs/ideas/README.md` Current line, and added a kit README row. Copy its archive shape ([`docs/ideas/archive/glyph-consolidation.md`](../../../ideas/archive/glyph-consolidation.md)). `ab9dd201` ("Name the effects registry path in the searchable-kind skill's projection step") is the same kind of one-line skill correction in the file STEP-1 edits.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `add-catalog-model` step 6 becomes exactly: "6. A table outside the skip bucket also needs an `effects` entry in the same PR (`core/database/effects/registry.go`). One `Effect` per table. Its fields, what each one moves, and how to add a catalog-wide list are in `core/database/effects/README.md` (a relative link from the skill, `../../../core/database/effects/README.md`); read them there, not from a copy here. An empty field means that job is unaffected; `None: true` is the explicit empty entry. Skip tables stay out of the registry. A path step that names an edge has to be a real `catalogmodel.FKs` row; package init panics when it is not." | Working rule (K-13) names this fix; the README is the owner and S12-25 / S12-26 change the fields again | Renaming `Source` → `EvidenceGraph` in the list — drifts again at S12-25 |
| D-2 | `add-searchable-kind`: one checklist line after the line that starts "FTS projector in core/database/searchindex", reading: - [ ] A `fieldValuesForKind` case (`core/search/fts.go`) mapping the document's columns onto every registry field the kind declares. And one sentence appended to "3. Projection + retrieval" item 2: "A kind whose document fills `secondary` or `body` with a declared field also needs a case in `core/search/fts.go` `fieldValuesForKind`; without one the kind falls back to `title`, `ref` and `body`, and a declared field such as `description` scores zero with no error." | Audit A-6 bullet 3; the failure is silent, so the skill is the only guard until S12-36 | A test that every registry field is mapped — code, and S12-36 reshapes this area |
| D-3 | `docs/conclusion-reconciliation.md` §6: the sentence "The state and list counts read the displayed values; search and matching read them all, so an outvoted spelling still finds its Person." becomes "The state and list counts read the displayed values; search reads them all, so an outvoted spelling still finds its Person. Matching reads the profile instead: displayed values count for a match, denied values against it, and outvoted, weak and provisional values are left out (a link to `matching-model.md`)." | Matches `subjectprofile.Build` and `docs/matching-model.md` line 28 | Linking an anchor — `docs/matching-model.md` has no heading for the profile builder |
| D-4 | `docs/matching.md`: two table rows after row 4. Row 5: layer "5 Segment compare", package `core/segmentcompare` (linked the way rows 1–4 link theirs), says "Whether two primary Subjects joined by a bridge line up with a research segment". Row 6: layer "6 Graph alignment", package `core/graphalign` with `core/shortlist` and `core/kinship` (each linked the same way), says "Which handle, New or Skip each primary Subject of a layer goes to: band, contested mark, reasons". The "## Not built yet" section keeps its heading and only the merge-hints sentences. `docs/matching-model.md` line 3's first two sentences become "The matching model Spike 11 built; all six layers run, and `matching.md` (linked) says where." | Package docs of the three packages; spike-11 map rows Landed | Folding `docs/matching.md` into `docs/matching-model.md` — a restructure, not a correction |
| D-5 | The remaining A-23 lines are each one edit: (a) `docs/catalog-deletes.md` line 73 enum reads `ok \| inbound \| not_found \| edge_locked \| last_graph \| infra \| origin_locked`. (b) `docs/macos-client-patterns.md` line 26 Purpose reads "One Evidence graph, opened by graph id: cards, connect, composer handoff." (c) `core/database/effects/README.md`: after the sentence ending "and Subject, Observation and claim rows name their Subject's graph." insert "A Source row names its graphs only on a write that changes `title` (`onField`), because a graph's search document and canvas hold the Source title; the same write reprojects those graphs' search documents (`DocEvidenceGraph`)." (d) `api/proto/README.md` line 36: delete the sentence "There are no Source tables." (e) `docs/source-layer-data-model.md` §4.2: after the **Cover.** paragraph add "**Graphs.** Like every new Source, the copy gets its own Default graph (`sources.Create`), and each copied Artifact is linked to it (`artifacts.Create` with no graph ids links every graph of the Source). Nothing of the original's graphs carries over." (f) `docs/promote-graph-alignment.md` line 167: "on the Source's Evidence graph memo" → "on its Evidence graph's memo". | Each matches the code cited in §4 | Rewriting the surrounding sections — out of the audit finding |
| D-6 | Six rows go in the `macos/App/DesignSystem/README.md` "What's built vs. not yet" table, file paths relative to `macos/App/DesignSystem/` as the other rows write them, each placed right after the existing row it sits beside: **Chip** after Badge — "`PVChip.swift` under Components/Chip (a meaning-free selectable pill for exclusive options such as credibility and date kind, and `PVChipGroup` to lay a set out; sibling to the display-only `PVBadge`; domain vocabulary stays in the feature that composes it)". **Copy** after Button — "`PVCopy.swift` under Components/Copy (`PVCopy`: the copy a component shows or speaks, a `LocalizedStringResource`, an already formatted `String` or an `AttributedString`, drawn through `pvText`; kit titles and labels take it)". **FloatingMenuSelection** after ContextMenu — "`PVFloatingMenuSelection.swift` under Components/FloatingMenuSelection (clamped ↑/↓ highlight movement with no wrap, used by `PVList`, `PVTable`, `PVSelect`, `PVContextMenu`, the history menu and the omnibar; `PVTypeSelectMatcher` is the type-to-select buffer)". **HoverEffect** after IconButton — "`PVHoverEffect.swift` under Components/HoverEffect (the interaction chrome inside a `ButtonStyle` body: hover tracking, disabled opacity, reduce-motion-aware press scale and one hit shape; callers draw from `showHover`)". **InlineEdit** after TextArea — "`PVInlineEdit.swift` under Components/InlineEdit (display and pencil → editor and Save / Cancel, used by the Source page fields and Properties; `PVInlineEditActions` is the Save / Cancel row for always-editable surfaces; accessibility ids stay at the call site unless `accessibilityIdentifierPrefix` is set)". **LastUpdated** after OmnibarHitRow — "`PVLastUpdated.swift` under Recipes/LastUpdated (the last-updated cell of a list row, shared by the Conclusion lists and the Sources list: history icon, relative time with the full stamp as help, a faint dash when none, and the row label's spoken suffix; `PVListDetailCell` draws it)". `add-ui-component` gains, after the checklist line that starts "Path is Components": "- [ ] README row: a new `Components/<Name>/` or `Recipes/<Name>/` folder has a row in the DesignSystem README's built table". | Each row's text comes from the type's own doc comment and its callers at the base | A separate table for these six — the README has one built table |
| D-7 | `docs/ideas/settled-list-count-publish.md` moves (git mv) to the ideas archive folder and is rewritten in the `glyph-consolidation.md` shape: title "Settled-list count publish (archived)"; status "**Done** in `c6dcd51d`, `aedf7aa2` and `228bd821`."; one paragraph of what shipped (`QueryHandle.isSettled`, the `.onSettled(handle)` modifier, Sources, Properties and Metadata publish through it, the post-write publishes dropped, the Properties off-by-one fixed); a Leftovers line: "A key marked stale while hidden can publish its cached count on appear: Spike 12 S12-26 (`QueryHandle.isStale`)." The problem essays (Properties, Metadata sections) and the Size line are dropped. `docs/ideas/README.md`: its Current line goes and an Archived line is added. The map's four links (lines 54, 143, 477, 545) point at the archived file. In `docs/ideas/shared-list-helpers.md` the "## Query-wait test helper" section is deleted; its `docs/ideas/README.md` line becomes a link to the file followed by " — small shared helpers: Properties' empty search state, one catalog querier, a trimmed empty-string binder, one popup dismissal and one popup list". | `archive-docs` "Shipped idea" steps 1–4; map scope entry | Deleting the idea outright — `archive-docs` keeps the decisions |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| Six README rows *(new)* | doc rows | The existing rows of the same table | Extends that table; no new table |
| `add-ui-component` "README row" checklist line *(new)* | doc line | Same checklist | Extends it |
| `add-searchable-kind` `fieldValuesForKind` checklist line *(new)* | doc line | Same checklist | Extends it |
| `docs/ideas/archive/settled-list-count-publish.md` *(new)*, moved from `docs/ideas/` | doc | `docs/ideas/archive/glyph-consolidation.md` | Same shape, per `archive-docs` |
| `matching.md` rows 5 and 6 *(new)* | doc rows | Rows 1–4 of the same table | Extends it |

**Builds on.** `core/database/effects/README.md` as the one owner of `Effect` fields; the kit README's built table; `archive-docs` "Shipped idea".
**Deletes.** The `Effect` field list in `add-catalog-model` step 6; "There are no Source tables."; `matching.md`'s not-built list for S11-14 – S11-16; `matching-model.md`'s "until then"; the "Query-wait test helper" idea section; `docs/ideas/settled-list-count-publish.md` (moved).
**Next adds.** S12-25 and S12-26 change `Effect` fields and edit only `core/database/effects/README.md`; S12-36 reshapes `add-searchable-kind` from a skill that names `fieldValuesForKind`; S12-31 edits the `add-ui-component` checklist beside the new README-row line.

### Reach

None — docs only, no new path into code.

### Scale

None — no write, list or loop changes.

### Schema

None — no catalog change.

### Go core

None — `core/database/effects/README.md` is the only file under `core/` this PR edits.

### FFI

None — `api/proto/README.md` loses one sentence; `engine.proto` is unchanged.

### macOS

None — `macos/App/DesignSystem/README.md` only.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Skills and rules (A-6)**
Files: `.agents/skills/add-catalog-model/SKILL.md`, `.agents/skills/add-searchable-kind/SKILL.md`, `.cursor/rules/catalog-refs.mdc`.
Do: D-1; D-2; in `catalog-refs.mdc` line 13 the list becomes `` `USR`, `SRC`, `ART`, `CIT`, `OBS`, `GPH` ``.
Check: `rg -n CitationSource .agents/skills/add-catalog-model/SKILL.md` → no hit; `rg -n 'effects/README.md' .agents/skills/add-catalog-model/SKILL.md` → step 6; `rg -c fieldValuesForKind .agents/skills/add-searchable-kind/SKILL.md` → `2`; `rg -n GPH .cursor/rules/catalog-refs.mdc` → line 13; `python3 scripts/check-guidance-names.py` → OK.

**STEP-2 — Matching docs (A-23)**
Files: `docs/conclusion-reconciliation.md`, `docs/matching.md`, `docs/matching-model.md`.
Do: D-3; D-4.
Check: AC-2.1 – AC-2.3 commands; `rg -n 'core/segmentcompare\|core/shortlist\|core/kinship' docs/matching.md` → the row-6 and row-5 lines.

**STEP-3 — Catalog and client docs (A-23)**
Files: `docs/catalog-deletes.md`, `docs/macos-client-patterns.md`, `core/database/effects/README.md`, `api/proto/README.md`, `docs/source-layer-data-model.md`, `docs/promote-graph-alignment.md`.
Do: D-5 (a)–(f).
Check: AC-2.4 – AC-2.9 commands.

**STEP-4 — Kit README rows (A-33)**
Files: `macos/App/DesignSystem/README.md`, `.agents/skills/add-ui-component/SKILL.md`. Skill: `add-ui-component`.
Do: D-6.
Check: from `macos/App/DesignSystem`, `for d in Components/* Recipes/*; do grep -q "\`$d/" README.md || echo "MISSING ROW: $d"; done` prints nothing; `rg -n 'README row' .agents/skills/add-ui-component/SKILL.md` → one hit.

**STEP-5 — Ideas (archive and trim)**
Files: `docs/ideas/settled-list-count-publish.md` → `docs/ideas/archive/settled-list-count-publish.md` *(new)* (`git mv`), `docs/ideas/shared-list-helpers.md`, `docs/ideas/README.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (four links). Skill: `archive-docs` "Shipped idea".
Do: D-7.
Check: `test ! -e docs/ideas/settled-list-count-publish.md && test -f docs/ideas/archive/settled-list-count-publish.md`; `rg -n 'ideas/settled-list-count-publish' docs .agents .cursor` → no hit; `rg -n 'Query-wait' docs/ideas/shared-list-helpers.md` → no hit.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-20-docs-match-code.md` → 0 errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Script | `scripts/check-guidance-names.py` | Every code name a skill or rule cites exists → `check-guidance-names: OK` | AC-1.1 – AC-1.3, AC-3.2 |
| T-2 | Script | `scripts/check-plan-docs.py` on this plan | → 0 errors | D-7 (map links), STEP-6 |
| T-3 | Shell | STEP-4's README row loop | → no output | AC-3.1 |
| T-4 | Shell | The `rg` / `test` command in each AC row | Each → the result its AC states | AC-1.1, AC-1.2, AC-1.3, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-2.8, AC-2.9, AC-3.2, AC-4.1, AC-4.2 |

Must stay green: everything `scripts/check-pr.sh` runs. A docs-only branch changes no Go or Swift file, so `guidance-names` and `plan-docs` are the checks it can break.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: the rg / test command and its output for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`ideas/settled-list-count-publish.md` outside `archive/`, `Effect.Source`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: the A-23 docs in STEP-2 and STEP-3 are this PR
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open `.agents/skills/add-catalog-model/SKILL.md` step 6 and follow its link: it opens `core/database/effects/README.md`, which names `EvidenceGraph`.
2. Open `docs/matching.md`: the table has six rows; "Not built yet" holds only merge hints.
3. Open `macos/App/DesignSystem/README.md`: Chip, Copy, FloatingMenuSelection, HoverEffect, InlineEdit and LastUpdated each have a row next to their neighbor named in D-6.
4. Open `docs/ideas/README.md`: settled-list count publish is under Archived, and its link opens the archived file.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A line moved between the base and the implementer's checkout | Each Step's check is an `rg` on text, not a line number |
| `check-guidance-names.py` flags a name the new text cites | Every name in D-1 – D-7 was checked to exist at the base |

**Stop and report if:**

- A §4 row is no longer true at your base (the doc was already fixed, or the code moved): report the row; do not re-word a fixed line.
- `core/database/effects/README.md` no longer describes the `Effect` fields D-1 sends readers to.
- `check-guidance-names.py` fails on a name D-1 – D-7 cites.
- `scripts/check-pr.sh` cannot run `xcode-test` on your machine (not macOS): report its full output; do not edit the script.
- Any step would need a change to a `.go`, `.swift`, `.proto` or script file.

## 12. Open questions

None.

## 13. PR

**Title:** S12-20 — Docs and skills match the code
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (the `rg` outputs); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-20-docs-match-code.md`.
