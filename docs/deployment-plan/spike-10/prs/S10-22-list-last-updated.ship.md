# Ship log — S10-22 Last updated on People, Events, and Places

**Plan:** [`S10-22-list-last-updated.md`](S10-22-list-last-updated.md) · **Branch:** `s10-22-list-last-updated` · **PR:** [#402](https://github.com/mendahu/provenencia/pull/402)
**State:** rework — implementing

Base: `main` @ `832ccc22` (planning base `360b405f`; the only change since, #400, touches `PVFormDialog` and none of the plan's Current state files).

## Round 0 — implement
Head `1ab370f0`. Gate PASS (orchestrator run against `832ccc22`). Report: no product deviations. Tests normalize U+202F (narrow no-break space before AM/PM on newer macOS) to a space before comparing. The implementer's first gate hit a flake in `WorkspaceSessionTests.invalidateAllMatching()` (not touched by this PR); it passed on rerun. Manual check not run.

## Round 1 — review
Five lenses, base `832ccc22`, head `1ab370f0`. Verdicts: CON Changes needed · MNT Clean · UX Clean · SEC Clean · PRF Clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | No Landed note; code differs from plan (U+202F before AM/PM, `Equatable` on `LastUpdatedDisplay`, `if let` for yesterday) | Fix now → F-1 | Verified: plan diff changes only the Status line. Precedent S10-21 CON-1 |
| PRF-1 | PRF | minor | §11 says hidden lists refetch on next show; they reload immediately (list specs have no `evictWhenHidden`) | Fix now (record) → F-1; behavior Follow-up | Verified: `CatalogQueryRegistry.swift` list specs set no `evictWhenHidden`. The plan's Decisions are unchanged; the Landed note records the correction. Stale-on-show or row patching is a follow-up |
| CON-3, MNT-1, UX (outside) | CON/MNT/UX | minor | `PVList` header doc now attaches to `PVListDetail`; `PVList` has no doc comment | Fix now → F-2 | Verified: `PVList.swift` lines 3–16, no break before `struct PVListDetail`. STEP-6 asked for this comment on `PVList`; a two-type move |
| CON-2 | CON | major | AC-1.7, AC-1.8, AC-2.4 have only manual-check evidence, not yet run | To user | The §10 manual check is the user's step (ship-pr Finish). Listed in the report; the plan's Stop condition stays if step 5 fails |
| MNT-2 | MNT | minor | `LastUpdatedDisplay.help` ignores the injected calendar/locale (`boardStamp` reparses) | Follow-up | D-6 names `boardStamp(rfc3339)`; fix when S10-41 adopts the type |
| PRF-2 | PRF | minor | Handle-scope path walked twice per write (`audit.Record` + `Resolve`) | Follow-up | Out line in the plan; touches the audit API. Benchmark first |
| UX-1 | UX | nit | Time of day for older changes is hover-only | Follow-up | Board decision (PSL-3); optional accessibility hint later |
| UX-2 | UX | nit | `.fixedSize()` text could overrun the ref in a longer locale | Logged | English only today; revisit with a second language |
| MNT-3 | MNT | nit | Edmonton calendar fixture copied in two test files | Logged | Extract when S10-41 adds a third |
| CON-4 | CON | nit | Tests pin English literals | Logged | Plan prescribed the literals |
| — | SEC | — | No findings | — | — |

### Fix plan 1

**F-1 — Add the Landed note** (from CON-1, PRF-1)
Files: `docs/deployment-plan/spike-10/prs/S10-22-list-last-updated.md`.
Do: add a short `## Landed` section directly under the plan header block (the S10-21 plan's Landed section is the shape). Record only where the code differs from the plan: (1) on current macOS the time separator before AM/PM is U+202F, so the cell reads `Today, 3:15 PM` with a narrow no-break space, and the tests normalize it to a space before comparing; (2) `LastUpdatedDisplay` also conforms to `Equatable`; (3) the yesterday test uses `if let` on `calendar.date(byAdding:)` instead of `!`; (4) §11's risk mitigation was wrong: list specs have no `evictWhenHidden`, so a note save reloads every cached list at once, not on next show. Reload-on-show or patching rows in place is a follow-up. Do not change any Decision, AC, or Scope line.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-22-list-last-updated.md` reports no errors.

**F-2 — Give `PVList` its doc comment back** (from CON-3, MNT-1)
Files: `macos/App/DesignSystem/Components/List/PVList.swift`.
Do: move `PVListDetail` and `PVListDetailColumn` (with their own `///` comments) to sit after the `PVList` type and its `Secondary == EmptyView` extension, before `PVListDensity`, so the header doc block sits directly on `struct PVList`. No code change.
Check: the header doc is immediately above `struct PVList<…>`; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `6ba75fec`. Gate PASS (orchestrator run). F-1 done (`6ba75fec`, Landed note); F-2 done (`d4a88df0`, doc comment back on `PVList`). No deviations.

## Round 2 — review
Lenses with a Fix-now finding: CON, MNT, PRF. Scope `1ab370f0...6ba75fec`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | — | Landed note | Resolved | Note matches the code; check-plan-docs 0 errors |
| CON-3 / MNT-1 | CON/MNT | — | `PVList` doc comment | Resolved | Header doc directly above `struct PVList` |
| PRF-1 | PRF | — | §11 correction | Resolved | Landed note accurate; behavior change stays a follow-up |
| CON-r2 | CON | nit | Landed note said "yesterday test" for app code | Fixed by orchestrator (one-word doc edit in the ship-log commit) | `TimestampFormat.swift:67` |

No new findings. Verdicts: CON Clean · MNT Clean · PRF Clean. No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: PRF-1 (reload hidden lists on show, or patch rows in place), PRF-2 (resolve handle scopes once per write), MNT-2 (inject calendar/locale into the help stamp, with S10-41), UX-1 (full stamp for keyboard/VoiceOver) · to user: CON-2 (run §10 manual check steps 2, 4, 5 for AC-1.7, AC-1.8, AC-2.4)

## Rework — user direction (2026-10-15)

After round 2 the user asked for two refactors on this PR and rated both problems as blockers the review should have caught:
1. One timestamp formatter (timestamp, locale, display option → one line of text, no styling) instead of `LastUpdatedDisplay` beside `TimestampFormat`.
2. Surgical cache updates: patch the moved rows' time from the write instead of reloading all three lists; cache infrastructure, not a one-off.

Plan amended (D-3, D-5, D-6, D-8, D-10, D-11 rewritten; D-12…D-18 new; STEP-9…STEP-15; T-7…T-13). The engine side also keeps last activity in the display cache, patched on commit, so list loads stop scanning audit history, and resolves scopes once per write (retires PRF-2). Lenses tightened: maintainability Q14 (one mechanism per cross-cutting job), Q9 (patch what the write reported), "precedent is not a mitigation"; performance treats any common-path cost that grows with the catalog or its history as a blocker.
