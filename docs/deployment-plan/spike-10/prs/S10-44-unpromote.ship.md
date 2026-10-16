# Ship log — S10-44 Unpromote a filed card

**Plan:** [`S10-44-unpromote.md`](S10-44-unpromote.md) · **Branch:** `s10-44-unpromote` · **PR:** #418
**State:** ready

## Round 0 — implement
Head `5a1e1e9e`. Gate PASS (orchestrator run against `98819f86`). PR #418 (draft).
Report deviations: `unpromoteConfirmTitle(name:)` returns `String` via `L10n.format`; `call_test.go` unknown method moved 100→101; `FakeStore.unpromoteSubject` sets `effects.sources` itself; SwiftUI twin's Unpromote accessibility action is always present and no-ops when unpromoted; T-1 cases are subtests; STEP-9 links repointed to S10-D6's archive path; §10 manual check not run.

## Round 1 — review
Lenses: all five on head `5a1e1e9e` (UX re-dispatched once: the first reviewer handed back before reading the diff). SEC clean, PRF clean, UX clean, CON clean, MNT changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `FakeStore.unpromoteSubject` reports a Source revision bump, `citationSources` and `sourceScopes` the engine never reports, and no `handleScopes` | Fix now → F-1 | verified: `FakeStore.swift:1413` calls `bumpSource(sourceID, evidence: true)`; `effects/registry.go:254-262` gives `identity_claims` only Source, Handles and HandleScope; AC-3.7 says the Source revision does not move |
| MNT-2 | MNT | major | Confirm detail header rebuilds `pvMicroCaps()` by hand | Fix now → F-2 | verified: `EvidenceGraphView.swift:224-228` is the body of `PVTypography.swift:164` |
| MNT-3 | MNT | minor | Painted Unpromote fill and glyph hard-code 22/13 instead of `PVControlSize.sm` tokens | Follow-up | real; `EvidenceAppKitCard.swift` is not in the fix plan |
| MNT-4 | MNT | minor | SwiftUI accessibility chain keeps a second list of card actions that already diverges (always lists Unpromote) | Follow-up | D-14 kept the chain in this PR; live card uses `namedActions(for:)`. Dedupes UX-2 |
| UX-2 | UX | nit | Twin always lists Unpromote VoiceOver action | → MNT-4 | same place, same problem |
| UX-1 | UX | minor | Stale-Subject Unpromote failure reads "This subject can't be promoted." | Follow-up | D-4 reuses Promote's codes; rare path; idea file |
| CON-1 | CON | minor | `TestUnpromote` is subtests, not a table | Follow-up | deviation recorded in the Landed note; `unpromote_test.go` is not in the fix plan |
| CON-2 | CON | minor | T-1(a) does not assert audited pin deletes for the bridge claims | Follow-up | real test gap; `ReleaseFacets` runs today; `unpromote_test.go` not in the fix plan |
| CON-3 | CON | minor | Swift tests assert raw English "Unpromote" | Follow-up | test files not in the fix plan; `unpromoteCopyResolves` pins the copy |
| CON-4 | CON | nit | T-5 omits the bridge-card case | Logged | `EvidenceBridgeCard.actionTargets` untouched |
| CON-5 | CON | nit | S10-20 STEP-6 now links the archive path but still says "do not archive" | Logged | named in the PR body |
| CON-OUT-1 | UX (from CON) | minor | Canvas `toolTip` stays "Unpromote" after the write until the next mouse move | Follow-up | `GraphCanvasDocumentView.syncToolTip` runs only on move/exit; not in the fix plan |
| MNT-5 | MNT | nit | `got[:len(prefix)]` can panic; third revision-count helper | Logged | use `strings.HasPrefix` when the file is next touched |

### Fix plan 1

**F-1 — The fake's Unpromote reports the engine's effects** (from MNT-1)
Files: `macos/App/Platform/FakeStore.swift`, and the Landed note line in `docs/deployment-plan/spike-10/prs/S10-44-unpromote.md` that says the fake "sets `effects.sources` itself".
Failing test first: in `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` (or the FakeStore tests if they exist), call `store.unpromoteSubject` on a promoted Subject and `#expect` the result equals `CatalogWriteEffects(handles: <removed entity ids>, sources: [sourceID], handleScopes: <removed entity ids>)` with empty `citationSources`, `sourceScopes` and nil `activity`, and that the Source's `updated` activity in `sourcesByProject` is unchanged. It fails today on `citationSources`/`sourceScopes`/`activity`.
Do: in `unpromoteSubject` drop the `bumpSource` call and the `effects.sources` override; return `CatalogWriteEffects(handles: removed, sources: [sourceID], handleScopes: removed)` (use the memberwise init in `CatalogMutation.swift:58`). Keep `confirmUnpromoteRefreshesToPromote` passing. Update the Landed note line to say the fake returns the engine's `identity_claims` effects (handles, Source, handle scopes) and does not move the Source revision.
Check: the new test and the existing Unpromote Swift tests pass; scripts/check-pr.sh PASS.

**F-2 — Use `pvMicroCaps()` for the "Also unfiled" header** (from MNT-2)
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (`unpromoteDetail`).
Failing test first: none (style refactor; no behavior change).
Do: replace the `.font(...).tracking(...).textCase(.uppercase)` chain on `L10n.EvidenceGraph.unpromoteConfirmBridgesHeader` with `.pvMicroCaps()`, keeping `.foregroundStyle(PVColor.textMuted)`. Do not touch the other hand copies.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `39131984`. Gate PASS (orchestrator run). F-1 done (`fakeUnpromoteReportsTheEnginesEffects` red then green, `5d6330cc`); F-2 done (`39131984`). Note: the round-0 docs commit `9a4cf4cc` picked up the untracked ship log, so it is tracked on the branch.

## Round 2 — review
Lens: MNT only (the one lens with Fix-now findings). Scope `5a1e1e9e...39131984`.
Head `39131984`. MNT round 2: MNT-1 and MNT-2 resolved; no new findings. Verdict clean. No Fix plan 2.

## Result
Gate PASS · CI see PR #418 · open: none · follow-ups: MNT-3, MNT-4 (+UX-2), UX-1, CON-OUT-1, CON-1, CON-2, CON-3 · to user: —
