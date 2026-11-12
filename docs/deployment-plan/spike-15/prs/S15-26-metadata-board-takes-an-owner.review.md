# Review — S15-26 The metadata board takes an owner

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. M-2 (asked in round 1) was answered by measurement in round 2; see the writer's calls.

## Writer's own calls
- The draft owner (a board on a form) moves from S15-26 to S15-10, its first caller; S15-11 reuses it. The protocol (D-4) is shaped so it conforms with `.empty` effects. Map scope entries for S15-26 and S15-10 updated.
- The section header and intro stay on the Source page; the board is rows, suggestions and the Add form (D-7), because S15-D1 SD-9 draws the Artifact board under a Disclosure.
- One Add dialog and one clear confirm per page, not per board (D-7).
- After round 1: the plan now also touches Go and the wire (Clear answers the owner's rows, M-5) and deletes `.mutatedSourceMetadata` (M-3), so it takes a patch bump. The map's S15-26 scope entry says so.
- **M-2 resolved without you (veto if you disagree):** measured at planning, at the map's Scale (60 Artifacts with 3 values each), `GetSourceWorkspace` averages 2.3 ms and a Source-level `SetMetadata` 1.8 ms in Go. So a board write keeps reloading the workspace, as today; S15-09 measures the Swift redraw of 60 boards and patches the cache only if that is slow.
- Set now also answers the owner's rows (M-11), and `SetMetadataResponse.entry` is reserved: its only reader is `GoStore`, shipped in the same bundle.
- `getMetadataBoard` waits for S15-11; renaming `SourcePageMetadata*` chrome waits for S15-09.

## Round 1

### X-1 · blocker · STEP-1's check cannot print nothing
`SourceMetadata\(` matches the kept `.mutatedSourceMetadata(` in `CatalogMutation.swift`, `CatalogQueryRegistry.swift`, `SourcePageContext.swift` and `CatalogQueryRegistryTests.swift`.
**Answer:** Fixed — STEP-1's check is now `\b(set|clear|reorder)SourceMetadata\b|dismissSourceMetadataSuggestion|SourceMetadataError` with `--glob '!*.pb.swift'`, and says `.mutatedSourceMetadata` stays.

### X-2 · major · FakeStore for non-Source owners: choices left open
`bumpActivity`, usage count, workspace read, copy and delete read only `metadataBySource`.
**Answer:** Fixed — D-3: `bumpActivity` only for `.source`; usage count, workspace read, Copy Source and delete stay Source-only; the missing-field fallback is the same for every level.

### X-3 · major · Comments and doc strings that name deleted symbols
**Answer:** Fixed — STEP-1 rewrites the store and fake doc comments for an owner (with the dismiss wording given) and the `replaceEntry` comment; STEP-4 rewrites the text editor's comment to `MetadataBoardModel.drafts`.

### X-4 · major · T-6 as written does not compile
**Answer:** Fixed — T-6 keys by `MetadataOwner(level: .artifact, id: "a1")`, seeds `fields: [authorField(), repositoryField()]` through `makeStore`, and asserts field keys and the value.

### X-5 · major · STEP-4 and DoD checks need a GUI
**Answer:** Fixed — STEP-4 adds a runnable check comparing the sorted identifier suffixes at `7d2f3e7` with the new files; seeing them in the running app stays in the manual check.

### X-6 · minor · Choices the plan leaves implicit
**Answer:** Fixed — D-7: the row views keep their names; D-5 lists `isSavingAdd` and `isClearing`, stores `private let owner`, and says `moveSaved` restores `previous` even when the owner is gone (T-3 covers it); STEP-3 says `SourcePageView` needs no edit.

Note: decision IDs below are the round-1 plan's; the plan was renumbered when answering M-1 – M-10.

### M-1 · blocker · Each board keeps its own Add dialog and clear-confirm state, but the page hosts only one dialog
**Answer:** Fixed — D-9: `MetadataBoardDialogs` is the page's one Add dialog and clear confirm, targeting a board; the board keeps row state only (D-7) and exposes `add` and `clear` for it; `MetadataBoardDeleteItem` carries its board; T-6, T-7.

### M-2 · blocker · Every board write reloads the whole Source workspace; §6 Scale said None
**Answer:** Fixed in part — §6 Scale states the fan-out (one write reloads one `GetSourceWorkspace` with up to 60 Artifact boards), unchanged by this PR; `didWrite` is the one place a host changes it. Whether to patch instead, and where: **To user** (recommend S15-09). Plan status Blocked until answered.

### M-3 · major · `didWrite(_:valuesChanged:)` builds a dead distinction
**Answer:** Fixed — D-5 `didWrite(_ effects:)`; D-8 deletes `.mutatedSourceMetadata`, its kind, registry entries and `applyMetadataWrite`, edits `CatalogQueryRegistryTests`, `add-workspace-place` lines 172/175 and `docs/macos-client-patterns.md`; STEP-5; sweep extended.

### M-4 · major · The draft-host slot assumed a draft builds engine rows
**Answer:** Fixed — **Next adds** states the draft host's contract: typed value trimmed, empty `openHref`, `suggested` from the given rows, dismiss local and recorded in the draft, `sourcemetadata.invalid` at the form's save, no field-error path.

### M-5 · major · The shared board widens a Swift copy of the engine's Clear rule
**Answer:** Fixed in this PR — D-1: `ClearMetadataResponse.metadata` from `ListWorkspace` as dismiss does; `applyCleared` deleted (D-7); STEP-1, T-1; version bump patch. The fake keeps a scripted result (it scripts, per `swift-tests.mdc`).

### M-6 · major · FakeStore keeps two maps for one concept
**Answer:** Fixed — D-4: one stored `metadataByOwner`; `metadataBySource` a computed view over Source owners; no picker helpers.

### M-7 · minor · The board re-inlines `showFailure`
**Answer:** Fixed — D-7's private `report(_:)`.

### M-8 · minor · "Owner" names two things
**Answer:** Fixed — the protocol is `MetadataBoardHost`, the Source page's `SourcePageMetadataHost`; "owner" is only the `MetadataOwner` identity.

### M-9 · minor · The Source host hard-codes `.source`
**Answer:** Fixed — D-6: `SourcePageMetadataHost(context:owner:rows:setRows:)` with a `.source(context:)` factory; S15-09 instances it per Artifact.

### M-10 · nit · T-6 tested the fake
**Answer:** Fixed — the fake-map test and its AC are dropped.

## Round 2

Both reviewers confirmed the round-1 answers hold against the code (M-1, M-3, M-5, M-6, M-7 and X-1 – X-6). The maintainability reviewer asked that M-2 carry a Go measurement, and found one new blocker.

### M-2 (round 2) · the deferred reload must be measured
**Answer:** Fixed — measured with a throwaway handler test on `TestMetadataOwners`' fixture grown to 60 Artifacts × 3 values and 4 Source values, 30 runs: `GetSourceWorkspace` 2.3 ms, `SetMetadata` 1.8 ms. §6 Scale carries the numbers; the Out row sends a patch to S15-09 only if its Swift redraw is slow. Taken off To user.

### M-4 (round 2) · the draft contract left out clear and reorder
**Answer:** Fixed — **Next adds**: a draft `clear` empties a suggested row and removes any other (the one client place this rule lives, since nothing is stored); `reorder` returns its rows in the given order; `set` returns its rows.

### M-11 · blocker · Set still answers one row, so the board keeps a Swift copy of the engine's row placement
**Answer:** Fixed — D-1: `SetMetadataResponse` answers `metadata` from the list the handler already reads, `entry` reserved; D-3/D-5 `set` returns the list; D-7 deletes `replaceEntry`; D-4 the fake's set returns its list; T-1 checks the Set list order (AC-3.3); T-2 scripts the list.

### X-1 (round 2) · major · STEP-2's temporary clear path
**Answer:** Fixed — STEP-2 changes only the list line in `replaceEntry` and `applyCleared` and keeps their other lines; STEP-4 deletes both, and its check says so.

### X-2 (round 2) · minor · T-11 named in STEP-4's diff check
**Answer:** Fixed — dropped.

### X-3 (round 2) · minor · dialog guards and flags
**Answer:** Fixed — D-9: guards, defers and the nil-target return as today; `add` calls `clearFailure` first.

### X-4 (round 2) · minor · computed `metadataBySource` setter
**Answer:** Fixed — D-4 gives the getter, the setter (replace every `.source` key) and no lock.

### X-5 (round 2) · minor · two `fieldComboOptions`
**Answer:** Fixed — D-9: the dialogs' reads `addTarget?.fieldComboOptions ?? []`; the board keeps its own for the tests near lines 882 and 902.

### X-6 (round 2) · minor · the identifier comparison is by eye
**Answer:** Fixed — STEP-6 calls that half of the check manual.

### X-7 (round 2) · minor · Blocked on M-2
**Answer:** Fixed — M-2 answered by measurement; status Planned.
