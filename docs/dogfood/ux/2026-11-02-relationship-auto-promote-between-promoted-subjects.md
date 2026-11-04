# Relationship bridge should auto-promote when both ends are already promoted

- **Date:** 2026-11-02
- **Type:** Feature request (promotion / bridge filing)
- **Where:** Evidence graph / Connect (relationship mid-card between two person subjects)
- **Annoyance (verbatim):** "If you add a relationship subject in the evidence graph between two subjects that are already promoted, the relationship should auto-promote."
- **Consequence:** After both people are handles, a newly drawn spouse/parent/sibling relationship stays an unpromoted bridge on the graph (no membership row, no canonical `relationship` association) until the researcher discovers Promote all or promotes again — even though both ends already satisfy the product rule for automatic bridge filing. Conclusion-layer adjacency and search stay missing for work the researcher already finished on the endpoints.
- **Wanted:** Creating a cited relationship (or other bridge) between two already-promoted primaries files the bridge in the same action — same outcome as when the second endpoint was promoted after the bridge existed: an accepted claim on the bridge Subject onto the correct association handle.

## Does auto-promotion already exist? (partially)

**Yes for the usual order; no for Jake's order.**

| Order | What happens today |
| --- | --- |
| Bridge first, then promote endpoints | **Auto-files.** When the second primary gets an accepted claim, `promote.Save` → `identityclaims.AppendFiling` → `FileBridgesTx` finds unfiled bridges touching that Subject and runs `fileBridge` when both ends resolve to handles (`core/database/promote/promote.go`, `core/database/identityclaims/bridgefile.go`, `identityclaims.go`). Covered by `core/database/promote/bridges_test.go` (`TestBridgeFiling`, including `relationship` spouse/parent keys). |
| Promote batch Done on the graph | **Auto-files** every bridge on that graph whose ends are both handles via `promote.SaveBatch` → `identityclaims.FileEvidenceGraphBridgesTx` (`core/database/promote/batch.go`, `batch_test.go` `TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted`). |
| Accepted claim on a bridge Subject (non-promote path) | **Auto-files** that bridge when ends are ready (`identityclaims.Create` → `AppendFiling`; `bridges_test.go` "an accepted claim create files the bridge"). |
| **Both endpoints already promoted, then add bridge** | **Does not file.** `connect.CreateCitedBridge` mints the bridge Subject, position, Citation, and Observations and returns — it never calls `AppendFiling` / `FileBridgesTx` (`core/database/connect/connect.go`). No new endpoint claim is created, so nothing triggers filing. The product doc describes the intended rule ("after any claim create…" — `docs/promote-graph-alignment.md` §9.1) but this path never creates a claim on either end. |

Bridge cards show `membership` when filed (`PromoteModel.bridgeFacts` sets `alreadyPromoted: bridge.membership != nil`; `EvidenceGraphModel.promoteAllLocation()` comment notes relationships added later can still be filed from Promote). Bridge mid-cards have no per-card Promote footer like primaries (`EvidenceBridgeCard` vs `EvidenceSubjectCard.promoteActionID`).

## How it works today (observed, not fixed)

### Creating the relationship subject (Evidence graph → core)

- **UI:** Connect mode in `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` — `connectOriginID`, `finishConnect(to:)` resolves a `CatalogConnectRule`, then hands off to the citation composer (`connectComposerLocation` with `connectBridgeTypeKey`). After cite, the composer persists via FFI.
- **FFI:** `api/ffi/handlers/connect.go` `CreateCitedBridge` → `writes.Run` action `create_cited_bridge` → `connect.CreateCitedBridge`.
- **Core write:** Inserts a bridge-type Subject (`relationship`, `participation`, `location`, `place_relationship`, …) on the graph, wires endpoint Observations (`person` / `related_to` / `relationship_type` for person–person relationship per `connectrules`), and attaches a Citation. Does not promote or file.

### What "promote" means for a bridge

- **Primary subjects:** `promote.Save` / Promote flow (`macos/App/Features/Promote/PromoteModel.swift`, `PromoteFlow.swift`) — accepted identity claim onto a new or chosen `canonical_entities` row (Person, Event, Place, …).
- **Bridge subjects:** Not promoted through the card footer; they are **filed** — accepted claim onto a **`relationship` / `participation` / `location` / `place_relationship` association** handle whose key is the two endpoint handles plus disambiguation (e.g. `relationship_type` term; directed vs symmetric rules in `identityclaims.fileBridge` / `findAssociation`). Reconciliation and graph cache follow from that claim (`docs/promote-graph-alignment.md` §9.1 table).

### Filing mechanics (Go)

- **`identityclaims.FileBridgesTx`** — lists bridge Subjects touching a Subject id (or, for batch, on a graph via **`FileEvidenceGraphBridgesTx`**) and calls **`fileBridge`** when each end's Observation resolves to exactly one accepted handle (`resolveEndpoint`).
- **`promote.Unpromote`** — inverse for primaries: drops the primary's claim and **`UnfileBridges`** on bridges that have that Subject as an end (association handles remain; bridge claims go away).

### macOS refresh after filing

- Promote / unpromote / connect writes return **`CatalogWriteEffects`**; the graph snapshot reloads membership on bridge rows through the usual catalog session + graph cache path (same as other evidence writes).

## Open questions (not decided)

1. **Layer:** Should filing run inside **`connect.CreateCitedBridge`** (or `writes.Run` effects after it) — e.g. `FileBridgesTx` on the new bridge id or `AppendFiling` on an endpoint — vs. a Swift-only follow-up RPC? Product text says filing is automatic everywhere (§9.1); the gap is core-side today.
2. **Scope:** Same rule for **participation, location, place_relationship** when both ends are already handles, or relationship-only? The same `fileBridge` path covers all bridge kinds; Jake named relationship explicitly.
3. **Promote second endpoint later:** Already defined — bridge created while one end unpromoted files when the lagging end is promoted. No change implied unless filing-on-create duplicates that work (should be idempotent: `fileBridge` no-ops when already filed).
4. **Unpromote / demote:** Unpromoting one **primary** already **`UnfileBridges`** for bridges on that end; the bridge Subject stays on the graph without an accepted claim. If auto-file on create is added, unpromote behaviour likely stays as-is — open: should unpromoting one person also hide or warn on a relationship that was auto-filed between them and a still-promoted neighbor?
5. **Refusals:** Self-links and place-relationship cycles already **soft-refuse** filing (`fileBridge` returns `filed: false`; tests in `bridges_test.go`). Auto-file on create should not fail the connect write; same refusal semantics as promote-triggered filing.
6. **Promote page / graphalign:** `PromoteFlow` already models connections as `.promotes` when both ends are handles (`PromoteFlow.swift` `connections`). Auto-filing on create may make some rows "already promoted" (`BridgeFact.alreadyPromoted`) immediately — confirm that matches researcher expectations vs. showing a brief toast.

## Fix directions (options, not chosen)

1. **Call `FileBridgesTx(tx, newBridge.ID)`** (or `AppendFiling` on the bridge Subject) at the end of **`CreateCitedBridge`** when observations are complete — smallest hook; effects layer recomputes association caches like promote.
2. **Call `FileBridgesTx` on both endpoint ids** after create — redundant but reuses the same "touching" query.
3. **UI follow-up:** after a successful connect RPC, invoke **`promoteSubject`**-like filing only when both ends report membership — duplicates core rules in Swift; discouraged unless core change is blocked.
