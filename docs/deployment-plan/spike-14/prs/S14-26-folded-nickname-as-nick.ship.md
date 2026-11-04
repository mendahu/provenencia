# Ship log — S14-26 A folded nickname shows as a nickname

**Plan:** [`S14-26-folded-nickname-as-nick.md`](S14-26-folded-nickname-as-nick.md) · **Branch:** `s14-26-folded-nickname-as-nick` · **PR:** #709
**State:** ready

## Round 0 — implement
Head `361b21d1`. Implementer gate PASS. Report: search test title updated for reconciled nick (*Thomas Tom Smith*); Manual §10 not run.

## Gate — orchestrator
Head `361b21d1`. First two orchestrator runs FAIL on flaky `PVSegmentedControlTests` (gate-1: Not reproduced; implementer 4× PASS). Orchestrator re-run PASS. Gate PASS.

## Round 1 — review
Reviewers: CON (Composer), MNT (Grok 4.7), SEC+PRF (Grok 4.7). Skipped: UX (no `macos/App/**` product code, strings, or error codes; only MARKETING_VERSION in `project.pbxproj`).
Base `d56de8d3` · Head `361b21d1`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 + PRF-1 | MNT/PRF | blocker | Folded nick parts from different part types follow map iteration | Fix now → F-1 | verified: `assemble` ranged `folded` map; Go map order unstable |
| SEC-1 | SEC | major | Folded nickname keeps the raw part, with no length bound | Fix now → F-2 | verified: `nicknameParts` appends raw; punctuation-padded alias still folds |
| MNT-2 | MNT | major | Nickname-expansion check is written twice | Fix now → F-3 | verified: duplicate `CauseAlias`+`Expands` |
| CON-1 | CON | minor | AC-2.1 omits stale version 22 in rebuild test | Fix now → F-4 | verified: stale slice ended at 21 |
| MNT-3 | MNT | minor | Nick insertion looks up prefix by raw string | Fix now → F-5 | verified: `"prefix"` literal |
| MNT-4 | MNT | minor | Doc comments left on the wrong functions | Fix now → F-6 | verified in review |
| MNT-5 | MNT | minor | Second name-spec formatter beside `nameSpec` | Fix now → F-7 | verified: this PR added `nameSpecFromValue` |
| CON-2 | CON | minor | T-7 Scale gate is not enforced in tests | Rejected | Plan T-7 / STEP-2 are implementer-run bench, not CI `Test` |
| CON-3 | CON | minor | Manual check §10 not performed | Logged | User manual check at finish |

Nits (logged): `CON-4 · core/search/handles_test.go:590 · Step file list omits this golden update` · `MNT-6 · names.go:299 · copied Idx discarded when renumbered`

Outside lens (logged): CON note on `nickClass` starting true for all-Same Contains; UX note on search title `Thomas Tom Smith` unquoted (quotes stay on S14-09 surfaces).

### Fix plan 1
F-1 stable folded-nick order · F-2 `MaxPartBytes` / `PartValueOK` on assemble · F-3 `aliasExpandsNickname` · F-4 stale 22 · F-5 `PartTypePrefix` · F-6 comments · F-7 `joinNameSpecParts`

## Round 1 — fix
Head `8662bd82`. Gate PASS. F-1–F-7 done. Deviation: stability test expects settled `nick=William` before folded Bill/Billy.

Orchestrator verified minors: CON-1 stale includes 22; MNT-3 uses `PartTypePrefix`; MNT-5 `joinNameSpecParts`; F-2 `MaxPartBytes=1024` + assemble skips.

## Round 2 — review
Reviewers: MNT, SEC+PRF (had Fix-now blocker/major). Skipped: CON (only Fix-now minors; orchestrator verified). Scope `361b21d1...8662bd82`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | map iteration | Resolved | walks `order` then `sortedKeys`; `TestAssembleFoldedNickOrderStable` |
| MNT-2 | MNT | major | duplicate predicate | Resolved | `aliasExpandsNickname` |
| SEC-1 | SEC | major | unbounded raw nick | Resolved | `PartValueOK` on settled + folded copy; oversized Bill test |
| PRF-1 | PRF | minor | unstable order reprojects | Resolved | same as MNT-1 |

No new blockers.

## Result
Gate PASS · CI pending · open: none · follow-ups: Validate/Insert part-length bound (write path; assemble-only in this PR) · to user: Manual §10 · CON miss: none
